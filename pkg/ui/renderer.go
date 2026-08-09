package ui

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/glamour"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/genai"
)

// RenderMarkdown converts a markdown string to stylized terminal ANSI output using glamour.
func RenderMarkdown(rawText string) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return rawText
	}
	out, err := renderer.Render(rawText)
	if err != nil {
		return rawText
	}
	return out
}

// RunSinglePrompt executes a single prompt against the agent with multi-stage thinking & tool loop tracking.
func RunSinglePrompt(ctx context.Context, appAgent agent.Agent, promptText string) error {
	r, err := runner.NewInMemory("flashwhip", appAgent)
	if err != nil {
		return fmt.Errorf("failed to initialize runner: %w", err)
	}

	userMsg := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: promptText},
		},
	}

	sessionID := fmt.Sprintf("single-%d", time.Now().Unix())
	fmt.Printf("%s Processing prompt...\n\n", AssistantBadge.Render("[Flashwhip]"))

	tracker := NewStreamTracker()

	for ev, err := range r.Run(ctx, "user", sessionID, userMsg, agent.RunConfig{StreamingMode: agent.StreamingModeSSE}) {
		if err != nil {
			return fmt.Errorf("agent run error: %w", err)
		}
		if ev == nil {
			continue
		}

		if ev.Content != nil {
			for _, part := range ev.Content.Parts {
				if ev.Partial {
					if part.Text != "" {
						if part.Thought {
							tracker.TransitionToThinking()
							fmt.Print(ThinkingBadge.Render(part.Text))
						} else {
							tracker.TransitionToOutputting()
							fmt.Print(part.Text)
						}
						_ = os.Stdout.Sync()
					}

					if part.FunctionCall != nil {
						tracker.TransitionToIdle()
						fmt.Printf("\n%s %s(%v)\n", ToolCallBadge.Render("⚡ [Tool Executing]:"), part.FunctionCall.Name, part.FunctionCall.Args)
						_ = os.Stdout.Sync()
					}
				}

				if part.FunctionResponse != nil {
					tracker.TransitionToIdle()
					fmt.Printf("\n%s %s\n", ToolResultBadge.Render("✔ [Tool Result]:"), part.FunctionResponse.Name)
					_ = os.Stdout.Sync()
				}
			}
		}
	}

	tracker.TransitionToIdle()
	fmt.Println()
	return nil
}
