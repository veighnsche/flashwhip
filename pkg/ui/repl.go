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
	"google.golang.org/adk/v2/runner"
	"google.golang.org/genai"

	"flashwhip/pkg/config"
	"flashwhip/pkg/db"
	"flashwhip/pkg/provider/ollama"
	"flashwhip/pkg/tools"
)

// RunInteractiveREPL launches an interactive multi-turn REPL prompt loop.
func RunInteractiveREPL(ctx context.Context, appAgent agent.Agent, cfg *config.Config, targetSessionID string) error {
	fmt.Print(RenderBanner(cfg.ModelName, cfg.BaseURL))
	fmt.Println()

	sessionID := targetSessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("chat-%d", time.Now().Unix())
	}

	r, err := runner.NewInMemory("flashwhip", appAgent)
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

	fmt.Println(RenderBanner(cfg.ModelName, cfg.BaseURL))

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
				fmt.Printf("  %-20s %s\n", InfoValue.Render("/help"), "Show this help menu")
				fmt.Printf("  %-20s %s\n", InfoValue.Render("/sessions, /list"), "List stored conversation sessions")
				fmt.Printf("  %-20s %s\n", InfoValue.Render("/load <session_id>"), "Load and resume a past conversation session")
				fmt.Printf("  %-20s %s\n", InfoValue.Render("/model, /models"), "Query live available models from endpoint")
				fmt.Printf("  %-20s %s\n", InfoValue.Render("/tools"), "List all active coding harness tools")
				fmt.Printf("  %-20s %s\n", InfoValue.Render("/clear"), "Clear terminal screen")
				fmt.Printf("  %-20s %s\n", InfoValue.Render("exit, quit"), "Exit interactive session")
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
				session, msgs, sErr := database.GetSession(targetID)
				if sErr != nil || session == nil {
					fmt.Printf("\033[31m[Error]: Session %q not found in database\033[0m\n", targetID)
					continue
				}
				sessionID = targetID
				fmt.Printf("\n%s Attached to session %s (%d turns, title: %q)\n\n", ToolResultBadge.Render("✔ [Session Loaded]:"), InfoValue.Render(sessionID), session.TurnCount, session.Title)
				for _, m := range msgs {
					if m.Role == "user" {
						fmt.Printf("%s %s\n", AssistantBadge.Render("[User]:"), m.Content)
					} else {
						fmt.Printf("%s %s\n\n", AssistantBadge.Render("[Assistant]:"), m.Content)
					}
				}
				continue
			case "/model", "/models":
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

		if err := ExecuteStreamLoop(ctx, r, sessionID, userMsg, tracker); err != nil {
			fmt.Printf("\n\033[31m[Error]: %v\033[0m\n", err)
		}

		fmt.Println()
	}

	return nil
}
