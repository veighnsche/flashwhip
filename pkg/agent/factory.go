package agent

import (
	"context"
	"fmt"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"

	"flashwhip/pkg/config"
	"flashwhip/pkg/provider/ollama"
	"flashwhip/pkg/tools"
)

// BuildAgentWithModel creates and initializes an ADK 2.0 LLM agent and returns both the agent and underlying Ollama model.
func BuildAgentWithModel(ctx context.Context, cfg *config.Config) (agent.Agent, *ollama.Model, error) {
	model, err := ollama.NewModel(cfg.ModelName, cfg.BaseURL, cfg.APIKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Ollama model provider (%s): %w", cfg.BaseURL, err)
	}

	defaultTools, err := tools.DefaultTools()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load default tools: %w", err)
	}

	// Inject live project context (cwd, directory layout, coding rules) into the
	// system instruction so the agent is immediately oriented to the project.
	systemInstruction := cfg.SystemInstruction + config.BuildProjectContext(cfg.ProjectRoot)

	a, err := llmagent.New(llmagent.Config{
		Name:        "flashwhip_agent",
		Model:       model,
		Description: "Flashwhip AI Terminal Assistant powered by ADK 2.0 & Ollama.",
		Instruction: systemInstruction,
		Tools:       defaultTools,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create llmagent: %w", err)
	}

	return a, model, nil
}

// BuildAgent creates and initializes an ADK 2.0 LLM agent.
func BuildAgent(ctx context.Context, cfg *config.Config) (agent.Agent, error) {
	a, _, err := BuildAgentWithModel(ctx, cfg)
	return a, err
}

