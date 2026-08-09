package ui

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/genai"
)

// ExecuteStreamLoop runs the streaming event loop for a given prompt session and updates the UI stream tracker.
func ExecuteStreamLoop(ctx context.Context, r *runner.Runner, sessionID string, userMsg *genai.Content, tracker *StreamTracker) error {
	for ev, err := range r.Run(ctx, "user", sessionID, userMsg, agent.RunConfig{StreamingMode: agent.StreamingModeSSE}) {
		if err != nil {
			return err
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
	return nil
}
