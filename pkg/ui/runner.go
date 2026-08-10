package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/genai"

	"flashwhip/pkg/db"
)

// ErrMaxTurnsReached is returned when the agent exceeds the configured turn limit.
var ErrMaxTurnsReached = errors.New("max turns reached")

// ExecuteStreamLoop runs the streaming event loop for a given prompt session and updates the UI stream tracker.
// Text tokens are buffered and rendered through glamour markdown at the end of each turn for clean output.
// maxTurns limits the number of completed agent turns (tool-call round-trips); 0 means unlimited.
func ExecuteStreamLoop(ctx context.Context, r *runner.Runner, sessionID string, userMsg *genai.Content, tracker *StreamTracker, maxTurns int) error {
	var assistantParts []*genai.Part
	var textBuf strings.Builder
	completedTurns := 0
	pendingCalls := make(map[string]map[string]any)
	var lastPrintedToolLine string

	for ev, err := range r.Run(ctx, "user", sessionID, userMsg, agent.RunConfig{StreamingMode: agent.StreamingModeSSE}) {
		if err != nil {
			return err
		}
		if ev == nil {
			continue
		}

		// Count completed agent turns for the MaxTurns guard.
		if ev.TurnComplete {
			completedTurns++
			if maxTurns > 0 && completedTurns >= maxTurns {
				fmt.Printf("\n%s Max turns (%d) reached. Use --max-turns to increase the limit.\n",
					ToolCallBadge.Render("⚠ [Flashwhip]:"), maxTurns)
				break
			}
		}

		if ev.Content != nil {
			for _, part := range ev.Content.Parts {
				if part.Text != "" && ev.Partial {
					if part.Thought {
						// Thinking tokens: print immediately, no buffering.
						tracker.TransitionToThinking()
						fmt.Print(ThinkingBadge.Render(part.Text))
						_ = os.Stdout.Sync()
					} else {
						// Regular text: buffer silently; will be glamour-rendered at end of turn.
						tracker.TransitionToOutputting()
						textBuf.WriteString(part.Text)
						assistantParts = append(assistantParts, part)
					}
				}

				if part.FunctionCall != nil {
					tracker.TransitionToIdle()
					if part.FunctionCall.Args != nil {
						pendingCalls[part.FunctionCall.Name] = part.FunctionCall.Args
					}
					assistantParts = append(assistantParts, part)
				}

				if part.FunctionResponse != nil {
					tracker.TransitionToIdle()
					args := pendingCalls[part.FunctionResponse.Name]
					cwd, _ := os.Getwd()
					line := RenderCombinedToolExecution(part.FunctionResponse.Name, args, part.FunctionResponse.Response, cwd)
					if line != "" && line != lastPrintedToolLine {
						fmt.Printf("%s\n", line)
						_ = os.Stdout.Sync()
						lastPrintedToolLine = line
					}
					delete(pendingCalls, part.FunctionResponse.Name)
				}
			}
		}
	}

	tracker.TransitionToIdle()

	// Render buffered text through glamour markdown now that the full response is assembled.
	if textBuf.Len() > 0 {
		fmt.Print(RenderMarkdown(textBuf.String()))
		_ = os.Stdout.Sync()
	}

	// Persist to embedded SQLite database (structured GenAI Content trees).
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
