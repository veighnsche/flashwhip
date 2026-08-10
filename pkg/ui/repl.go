package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"google.golang.org/adk/v2/agent"
	adkmodel "google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	fagent "flashwhip/pkg/agent"
	"flashwhip/pkg/agent/middleware"
	"flashwhip/pkg/config"
	ollama "flashwhip/pkg/provider/ollama"
)

// RunInteractiveREPL launches an interactive multi-turn REPL prompt loop.
func RunInteractiveREPL(ctx context.Context, appAgent agent.Agent, cfg *config.Config, targetSessionID string, maxTurns int) error {
	fmt.Print(RenderBanner(cfg.ModelName, cfg.BaseURL, cfg.ProjectRoot))
	fmt.Println()

	var activeModel *ollama.Model
	if newAgent, modelPtr, aErr := fagent.BuildAgentWithModel(ctx, cfg); aErr == nil {
		appAgent = newAgent
		activeModel = modelPtr
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
		return fmt.Errorf("failed to initialize runner: %w", err)
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
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	cmdRegistry := DefaultRegistry()

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

		if err := ExecuteStreamLoop(ctx, r, sessionID, userMsg, tracker, maxTurns); err != nil {
			fmt.Printf("\n\033[31m[Error]: %v\033[0m\n", err)
		}

		// Auto-prune: adaptively compact in-memory session history when history is long
		// or context saturation exceeds 75% threshold.
		pruneSessionInMemory(ctx, sessionSvc, sessionID, tracker)

		fmt.Println()
	}

	return nil
}

// pruneSessionInMemory retrieves the current session events from sessionSvc,
// adaptively prunes their thought & tool-response payloads via middleware.PruneContentsAdaptive,
// and rebuilds the session so that future agent turns work with an optimized context.
// It is a best-effort operation: failures are silently ignored so they never
// interrupt the user's session.
func pruneSessionInMemory(ctx context.Context, sessionSvc session.Service, sessionID string, tracker *StreamTracker) {
	const autoPruneEvery = 10 // prune every N turns

	resp, err := sessionSvc.Get(ctx, &session.GetRequest{
		AppName:   "flashwhip",
		UserID:    "user",
		SessionID: sessionID,
	})
	if err != nil {
		return
	}

	// Collect contents from all events.
	var contents []*genai.Content
	for ev := range resp.Session.Events().All() {
		if ev != nil && ev.Content != nil {
			contents = append(contents, ev.Content)
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

	// Rebuild the session with pruned history.
	_, crErr := sessionSvc.Create(ctx, &session.CreateRequest{
		AppName:   "flashwhip",
		UserID:    "user",
		SessionID: sessionID,
	})
	if crErr != nil {
		return
	}
	for _, c := range pruned {
		if c == nil {
			continue
		}
		ev := &session.Event{
			Author: c.Role,
			LLMResponse: adkmodel.LLMResponse{
				Content: c,
			},
		}
		_ = sessionSvc.AppendEvent(ctx, resp.Session, ev)
	}

	if pct >= 75.0 {
		fmt.Printf("\n%s Context saturation (%.1f%%) reached threshold — Auto-compacted history.\n", ToolResultBadge.Render("⚡ [Auto-Compacted Context]:"), pct)
	}
}
