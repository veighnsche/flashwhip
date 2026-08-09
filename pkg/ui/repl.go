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
	"flashwhip/pkg/db"
	"flashwhip/pkg/provider/ollama"
	"flashwhip/pkg/tools"
)

// RunInteractiveREPL launches an interactive multi-turn REPL prompt loop.
func RunInteractiveREPL(ctx context.Context, appAgent agent.Agent, cfg *config.Config, targetSessionID string, maxTurns int) error {
	fmt.Print(RenderBanner(cfg.ModelName, cfg.BaseURL))
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

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if strings.HasPrefix(input, "/") {
			parts := strings.Fields(input)
			cmdName := strings.ToLower(parts[0])

			switch cmdName {
			case "/help":
				fmt.Println()
				fmt.Println(ThinkingBadge.Render("⚡ FLASHWHIP REPL COMMANDS:"))
				fmt.Printf("  %-22s %s\n", InfoValue.Render("/help"), "Show this help menu")
				fmt.Printf("  %-22s %s\n", InfoValue.Render("/sessions, /list"), "List stored conversation sessions")
				fmt.Printf("  %-22s %s\n", InfoValue.Render("/load <session_id>"), "Load and resume a past conversation session")
				fmt.Printf("  %-22s %s\n", InfoValue.Render("/model [model_name]"), "View or switch active model")
				fmt.Printf("  %-22s %s\n", InfoValue.Render("/tools"), "List all active coding harness tools")
				fmt.Printf("  %-22s %s\n", InfoValue.Render("/compact"), "Force compact conversation context history")
				fmt.Printf("  %-22s %s\n", InfoValue.Render("/clear"), "Clear terminal screen")
				fmt.Printf("  %-22s %s\n", InfoValue.Render("exit, quit"), "Exit interactive session")
				fmt.Println()
				continue
			case "/sessions", "/list", "/history":
				database, err := db.DefaultDB()
				if err == nil {
					sessions, _ := database.ListSessions()
					fmt.Println()
					fmt.Println(RenderSessionList(sessions))
					fmt.Println()
				}
				continue
			case "/load", "/reload":
				if len(parts) < 2 {
					fmt.Println(ThinkingBadge.Render("Usage: /load <session_id>"))
					continue
				}
				targetID := parts[1]
				database, err := db.DefaultDB()
				if err != nil {
					fmt.Printf("\033[31m[Error]: Failed to open database: %v\033[0m\n", err)
					continue
				}
				sessionInfo, msgs, sErr := database.GetSession(targetID)
				if sErr != nil || sessionInfo == nil {
					fmt.Printf("\033[31m[Error]: Session %q not found in database\033[0m\n", targetID)
					continue
				}

				// Hydrate ADK sessionSvc with past GenAI contents from SQLite
				contents, cErr := database.GetSessionGenAIContents(targetID)
				if cErr == nil && len(contents) > 0 {
					sessResp, crErr := sessionSvc.Create(ctx, &session.CreateRequest{
						AppName:   "flashwhip",
						UserID:    "user",
						SessionID: targetID,
					})
					if crErr == nil && sessResp != nil {
						for _, gc := range contents {
							evt := &session.Event{
								Author: gc.Role,
								LLMResponse: adkmodel.LLMResponse{
									Content: gc,
								},
							}
							_ = sessionSvc.AppendEvent(ctx, sessResp.Session, evt)
						}
					}
				}

				sessionID = targetID
				fmt.Printf("\n%s Attached to session %s (%d turns, title: %q)\n\n", ToolResultBadge.Render("✔ [Session Loaded]:"), InfoValue.Render(sessionID), sessionInfo.TurnCount, sessionInfo.Title)
				for _, m := range msgs {
					if m.Role == "user" {
						fmt.Printf("%s %s\n", AssistantBadge.Render("[User]:"), m.Content)
					} else {
						fmt.Printf("%s %s\n\n", AssistantBadge.Render("[Assistant]:"), m.Content)
					}
				}
				continue
			case "/model", "/models":
				if len(parts) >= 2 {
					targetModel := parts[1]
					cfg.ModelName = targetModel
					newAgent, aErr := fagent.BuildAgent(ctx, cfg)
					if aErr != nil {
						fmt.Printf("\033[31m[Error]: Failed to switch model to %q: %v\033[0m\n\n", targetModel, aErr)
						continue
					}
					appAgent = newAgent
					newRunner, _ := runner.New(runner.Config{
						AppName:           "flashwhip",
						Agent:             appAgent,
						SessionService:    sessionSvc,
						AutoCreateSession: true,
					})
					r = newRunner
					fmt.Printf("\n%s Switched active model to %s\n\n", ToolResultBadge.Render("✔ [Model Switched]:"), InfoValue.Render(cfg.ModelName))
					continue
				}

				fmt.Println()
				fmt.Printf("%s Querying endpoint %s...\n", AssistantBadge.Render("[Flashwhip]"), InfoValue.Render(cfg.BaseURL))
				modelsList, mErr := ollama.FetchAvailableModels(cfg.BaseURL, cfg.APIKey)
				if mErr != nil {
					fmt.Printf("\033[31m[Error]: Failed to fetch models: %v\033[0m\n\n", mErr)
				} else {
					fmt.Println(ThinkingBadge.Render("Available Endpoint Models:"))
					for _, mName := range modelsList {
						if mName == cfg.ModelName {
							fmt.Printf("  • %s %s\n", InfoValue.Render(mName), ToolResultBadge.Render("(active)"))
						} else {
							fmt.Printf("  • %s\n", mName)
						}
					}
					fmt.Println("\n  Tip: Switch active model using '/model <model_name>'")
					fmt.Println()
				}
				continue
			case "/tools":
				fmt.Println()
				toolInfos, tErr := tools.GetToolDescriptions()
				if tErr != nil {
					fmt.Printf("\033[31m[Error]: Failed to inspect tools: %v\033[0m\n\n", tErr)
				} else {
					fmt.Println(ThinkingBadge.Render("Active Code Harness Tools:"))
					for _, ti := range toolInfos {
						fmt.Printf("  • %-20s %s\n", ToolCallBadge.Render(ti.Name), ti.Description)
					}
					fmt.Println()
				}
				continue
			case "/compact":
				database, err := db.DefaultDB()
				if err != nil {
					fmt.Printf("\033[31m[Error]: Failed to open database: %v\033[0m\n", err)
					continue
				}
				contents, cErr := database.GetSessionGenAIContents(sessionID)
				if cErr != nil || len(contents) == 0 {
					fmt.Println(ThinkingBadge.Render("No active session history to compact."))
					continue
				}

				pruned := middleware.PruneContents(contents, 1)
				totalOrigChars := 0
				for _, c := range contents {
					for _, p := range c.Parts {
						totalOrigChars += len(p.Text)
					}
				}
				totalPrunedChars := 0
				for _, c := range pruned {
					for _, p := range c.Parts {
						totalPrunedChars += len(p.Text)
					}
				}

				fmt.Println()
				fmt.Printf("%s Context Compacting Completed:\n", ToolResultBadge.Render("✔ [Context Compacted]:"))
				fmt.Printf("  • Original history parts: %d (%d chars)\n", len(contents), totalOrigChars)
				fmt.Printf("  • Compacted history parts: %d (%d chars)\n\n", len(pruned), totalPrunedChars)
				continue
			case "/clear":
				fmt.Print("\033[H\033[2J")
				fmt.Println(RenderBanner(cfg.ModelName, cfg.BaseURL))
				continue
			}
		}

		userMsg := &genai.Content{
			Role: "user",
			Parts: []*genai.Part{
				{Text: input},
			},
		}

		fmt.Printf("\n%s\n", AssistantBadge.Render("[Assistant]"))
		tracker := NewStreamTracker()

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
