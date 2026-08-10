package ui

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/genai"

	"flashwhip/pkg/errors"
)

// RunSinglePrompt executes a single prompt against the agent with multi-stage thinking & tool loop tracking.
func RunSinglePrompt(ctx context.Context, appAgent agent.Agent, promptText string, maxTurns int) error {
	r, err := runner.NewInMemory("flashwhip", appAgent)
	if err != nil {
		return errors.Wrap(errors.ErrCodeAgentRunnerInitFailed, "failed to initialize runner", err)
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
	if err := ExecuteStreamLoop(ctx, r, sessionID, userMsg, tracker, maxTurns); err != nil {
		return errors.Wrap(errors.ErrCodeUISinglePromptFailed, "agent run error", err)
	}

	fmt.Println()
	return nil
}

