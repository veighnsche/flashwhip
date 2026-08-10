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

	"flashwhip/pkg/agent/middleware"
	"flashwhip/pkg/config"
)

// RunInteractiveREPL launches an interactive multi-turn REPL prompt loop.
func RunInteractiveREPL(ctx context.Context, appAgent agent.Agent, cfg *config.Config, targetSessionID string, maxTurns int) error {
	fmt.Print(RenderBanner(cfg.ModelName, cfg.BaseURL, cfg.ProjectRoot))
	fmt.Println()

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
			Ctx:        ctx,
			Config:     cfg,
			AppAgent:   &appAgent,
			Runner:     &r,
			SessionSvc: sessionSvc,
			SessionID:  &sessionID,
			Readline:   rl,
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
		tracker := NewStreamTrackerWithConfig(sessionID, cfg)

		if err := ExecuteStreamLoop(ctx, r, sessionID, userMsg, tracker, maxTurns); err != nil {
			fmt.Printf("\n\033[31m[Error]: %v\033[0m\n", err)
		}

		// Auto-prune: every 10 turns compact the in-memory session history to
		// prevent context window exhaustion on long sessions.
		pruneSessionInMemory(ctx, sessionSvc, sessionID)

		fmt.Println()
	}

	return nil
}

// pruneSessionInMemory retrieves the current session events from sessionSvc,
// prunes their tool-response payloads via middleware.PruneContents, and
// rebuilds the session so that the next agent turn works with a compacted history.
// It is a best-effort operation: failures are silently ignored so they never
// interrupt the user's session.
func pruneSessionInMemory(ctx context.Context, sessionSvc session.Service, sessionID string) {
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

	if len(contents) < autoPruneEvery*2 {
		return // not enough history to warrant pruning yet
	}

	pruned := middleware.PruneContents(contents, autoPruneEvery)

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
}
