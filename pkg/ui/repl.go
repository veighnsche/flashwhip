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

// RunInteractiveREPL starts an interactive terminal session with readline persistent history and styled UI.
func RunInteractiveREPL(ctx context.Context, appAgent agent.Agent, cfg *config.Config) error {
	r, err := runner.NewInMemory("flashwhip", appAgent)
	if err != nil {
		return fmt.Errorf("failed to initialize interactive runner: %w", err)
	}

	sessionID := fmt.Sprintf("chat-%d", time.Now().Unix())

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
		inReasoningMode := false
		var finalContentAccumulator strings.Builder

		for ev, err := range r.Run(ctx, "user", sessionID, userMsg, agent.RunConfig{}) {
			if err != nil {
				fmt.Printf("\n\033[31m[Error]: %v\033[0m\n", err)
				break
			}
			if ev == nil {
				continue
			}

			if ev.CustomMetadata != nil {
				if reasoningText, ok := ev.CustomMetadata["reasoning"].(string); ok && reasoningText != "" {
					if !inReasoningMode {
						fmt.Print(ThinkingBadge.Render("🧠 [Thinking]: "))
						inReasoningMode = true
					}
					fmt.Print(ThinkingBadge.Render(reasoningText))
				}
			}

			if ev.Content != nil {
				if inReasoningMode {
					fmt.Print("\n\n")
					inReasoningMode = false
				}

				for _, part := range ev.Content.Parts {
					if part.Text != "" {
						if ev.Partial {
							fmt.Print(part.Text)
						}
						finalContentAccumulator.WriteString(part.Text)
					}
					if part.FunctionCall != nil {
						fmt.Printf("\n%s %s(%v)\n", ToolCallBadge.Render("⚡ [Tool Call]:"), part.FunctionCall.Name, part.FunctionCall.Args)
					}
				}
			}
		}

		if inReasoningMode {
			fmt.Print("\n")
		}

		fullText := finalContentAccumulator.String()
		if fullText != "" {
			rendered := RenderMarkdown(fullText)
			fmt.Print("\n--- [Rendered Output] ---\n")
			fmt.Print(rendered)
		}

		fmt.Println()
	}

	return nil
}
