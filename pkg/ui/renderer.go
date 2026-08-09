package ui

import (
	"context"
	"fmt"
	"strings"
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

// RunSinglePrompt executes a single prompt against the agent and streams the response to stdout.
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

	inReasoningMode := false
	var finalContentAccumulator strings.Builder

	for ev, err := range r.Run(ctx, "user", sessionID, userMsg, agent.RunConfig{}) {
		if err != nil {
			return fmt.Errorf("agent run error: %w", err)
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

	return nil
}
