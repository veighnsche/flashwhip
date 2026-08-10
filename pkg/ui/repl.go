package ui

import (
	"context"
	stdErrors "errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chzyer/readline"
	adkmodel "google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	fagent "flashwhip/pkg/agent"
	"flashwhip/pkg/agent/middleware"
	"flashwhip/pkg/config"
	"flashwhip/pkg/db"
	"flashwhip/pkg/errors"
	ollama "flashwhip/pkg/provider/ollama"
)

// RunInteractiveREPL launches an interactive multi-turn REPL prompt loop.
func RunInteractiveREPL(ctx context.Context, cfg *config.Config, targetSessionID string, maxTurns int) error {
	fmt.Print(RenderBanner(cfg.ModelName, cfg.BaseURL, cfg.ProjectRoot))
	fmt.Println()

	appAgent, activeModel, err := fagent.BuildAgentWithModel(ctx, cfg)
	if err != nil {
		return errors.Wrap(errors.ErrCodeAgentBuildFailed, "failed to build agent", err)
	}

	sessionID := targetSessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("chat-%d", time.Now().Unix())
	}

	sessionSvc := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:           "flashwhip",
		Agent:             appAgent,
		SessionService:    sessionSvc,
		AutoCreateSession: true,
	})
	if err != nil {
		return errors.Wrap(errors.ErrCodeRunnerExecutionFailed, "failed to initialize runner", err)
	}

	homeDir, _ := os.UserHomeDir()
	historyFile := filepath.Join(homeDir, ".flashwhip_history")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          InfoValue.Render("flashwhip> "),
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return errors.Wrap(errors.ErrCodeHistorySaveFailed, "failed to initialize readline", err)
	}
	defer rl.Close()

	cmdRegistry := DefaultRegistry()

	// Graceful SIGINT (Ctrl+C) handling. readline drives the terminal in raw mode
	// while a line is being typed, so SIGINT only reaches us while the terminal is
	// in cooked mode — i.e. during streamed response generation or tool execution.
	// In that window we cancel the active turn so the loop returns cleanly to the
	// prompt instead of garbling the terminal or leaving orphans behind.
	var activeCancel context.CancelFunc
	var cancelMu sync.Mutex
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer signal.Stop(sigChan)
	go func() {
		for range sigChan {
			fmt.Printf("\n\n%s\n", ToolResultBadge.Render("⏹ Interrupted. Cleaning up..."))
			cancelMu.Lock()
			c := activeCancel
			cancelMu.Unlock()
			if c != nil {
				c()
			}
		}
	}()

	for {
		line, err := rl.Readline()
		if err != nil {
			fmt.Println("Goodbye!")
			break
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		cmdCtx := &CommandContext{
			Ctx:         ctx,
			Config:      cfg,
			AppAgent:    &appAgent,
			ActiveModel: &activeModel,
			Runner:      &r,
			SessionSvc:  sessionSvc,
			SessionID:   &sessionID,
			Readline:    rl,
		}

		res := cmdRegistry.Dispatch(input, cmdCtx)
		if res.ShouldExit {
			break
		}
		if res.Handled {
			continue
		}

		userMsg := &genai.Content{
			Role: "user",
			Parts: []*genai.Part{
				{Text: input},
			},
		}

		fmt.Printf("\n%s\n", AssistantBadge.Render("[Assistant]"))
		var usageTracker *ollama.Usage
		if activeModel != nil {
			usageTracker = activeModel.Usage()
		}
		tracker := NewStreamTrackerWithConfig(sessionID, cfg, usageTracker)

		// Per-turn cancelable context so a SIGINT mid-turn can stop the stream loop
		// and return control to the prompt without corrupting session state.
		streamCtx, streamCancel := context.WithCancel(ctx)
		cancelMu.Lock()
		activeCancel = streamCancel
		cancelMu.Unlock()

		runErr := ExecuteStreamLoop(streamCtx, r, sessionID, userMsg, tracker, maxTurns)

		cancelMu.Lock()
		activeCancel = nil
		cancelMu.Unlock()
		streamCancel()

		if runErr != nil {
			// A cancelled context is the normal outcome of a user interrupt; we've
			// already printed the "Interrupted" notice, so skip the red error line.
			if stdErrors.Is(runErr, context.Canceled) {
				fmt.Println()
			} else {
				fmt.Printf("\n\033[31m[Error]: %v\033[0m\n", runErr)
			}
		}

		// Auto-prune: adaptively compact in-memory session history when history is long
		// or context saturation exceeds 75% threshold.
		pruneSessionInMemory(ctx, sessionSvc, sessionID, tracker)

		fmt.Println()
	}

	return nil
}

// pruneSessionInMemory retrieves the current session events from DB/sessionSvc,
// adaptively prunes their thought & tool-response payloads via middleware.PruneContentsAdaptive,
// and rebuilds both SQLite and sessionSvc so that future agent turns work with an optimized context.
// It is a best-effort operation: failures are silently ignored so they never
// interrupt the user's session.
func pruneSessionInMemory(ctx context.Context, sessionSvc session.Service, sessionID string, tracker *StreamTracker) {
	const autoPruneEvery = 10 // prune every N turns

	database, dbErr := db.DefaultDB()
	var contents []*genai.Content

	if dbErr == nil && database != nil {
		contents, _ = database.GetSessionGenAIContents(sessionID)
	}

	if len(contents) == 0 && sessionSvc != nil {
		resp, err := sessionSvc.Get(ctx, &session.GetRequest{
			AppName:   "flashwhip",
			UserID:    "user",
			SessionID: sessionID,
		})
		if err == nil && resp.Session != nil {
			for ev := range resp.Session.Events().All() {
				if ev != nil && ev.Content != nil {
					contents = append(contents, ev.Content)
				}
			}
		}
	}

	if len(contents) == 0 {
		return
	}

	pct := 0.0
	if tracker != nil {
		pct, _, _ = tracker.ContextSaturationPct()
	}

	// Trigger auto-compaction if context usage reaches 75%+ or history reaches turn limit
	if pct < 75.0 && len(contents) < autoPruneEvery*2 {
		return
	}

	pruned := middleware.PruneContentsAdaptive(contents, pct)

	// 1. Sync pruned history back to SQLite database so token metrics drop immediately
	if database != nil {
		_ = database.ReplaceSessionGenAIContents(sessionID, pruned)
	}

	// 2. Rebuild the in-memory runner session with pruned history
	rebuildSessionEvents(ctx, sessionSvc, sessionID, pruned)

	if pct >= 75.0 {
		fmt.Printf("\n%s Context saturation (%.1f%%) reached threshold — Auto-compacted history.\n", ToolResultBadge.Render("⚡ [Auto-Compacted Context]:"), pct)
	}
}

// rebuildSessionEvents replaces a session's in-memory event history with the
// given GenAI contents (used after pruning or loading). Best-effort: failures
// are silently ignored.
func rebuildSessionEvents(ctx context.Context, sessSvc session.Service, sessionID string, contents []*genai.Content) {
	if sessSvc == nil {
		return
	}
	sessResp, crErr := sessSvc.Create(ctx, &session.CreateRequest{
		AppName:   "flashwhip",
		UserID:    "user",
		SessionID: sessionID,
	})
	if crErr != nil || sessResp == nil {
		return
	}
	for _, c := range contents {
		if c == nil {
			continue
		}
		_ = sessSvc.AppendEvent(ctx, sessResp.Session, &session.Event{
			Author: c.Role,
			LLMResponse: adkmodel.LLMResponse{
				Content: c,
			},
		})
	}
}
