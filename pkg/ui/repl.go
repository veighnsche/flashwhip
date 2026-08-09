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

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		userMsg := &genai.Content{
			Role: "user",
			Parts: []*genai.Part{
				{Text: line},
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
