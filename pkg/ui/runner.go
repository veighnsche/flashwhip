package ui

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/genai"

	"flashwhip/pkg/db"
)

// ExecuteStreamLoop runs the streaming event loop for a given prompt session and updates the UI stream tracker.
func ExecuteStreamLoop(ctx context.Context, r *runner.Runner, sessionID string, userMsg *genai.Content, tracker *StreamTracker) error {
	var assistantParts []*genai.Part

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
							assistantParts = append(assistantParts, part)
						}
						_ = os.Stdout.Sync()
					}

					if part.FunctionCall != nil {
						tracker.TransitionToIdle()
						fmt.Printf("\n%s %s(%v)\n", ToolCallBadge.Render("⚡ [Tool Executing]:"), part.FunctionCall.Name, part.FunctionCall.Args)
						_ = os.Stdout.Sync()
						assistantParts = append(assistantParts, part)
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

	// Persist to embedded SQLite database (structured GenAI Content trees)
	database, dErr := db.DefaultDB()
	if dErr == nil && database != nil {
		if userMsg != nil {
			_ = database.SaveContent(sessionID, userMsg)
		}
		if len(assistantParts) > 0 {
			assistantContent := &genai.Content{
				Role:  "model",
				Parts: assistantParts,
			}
			_ = database.SaveContent(sessionID, assistantContent)
		}
	}

	return nil
}
