package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"flashwhip/pkg/agent"
	"flashwhip/pkg/config"
	"flashwhip/pkg/ui"
)

var (
	cfg        *config.Config
	flagURL    string
	flagModel  string
	flagAPIKey string
)

var rootCmd = &cobra.Command{
	Use:   "flashwhip [command|prompt]",
	Short: "Flashwhip - AI Terminal Assistant powered by Google ADK 2.0 & Ollama",
	Long: `Flashwhip is a high-performance terminal assistant built on Google's ADK 2.0 framework.
It connects to Ollama endpoints (or OpenAI-compatible hosts) to execute single-shot prompts or interactive chat sessions.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		ctx := context.Background()
		applyFlagsToConfig()

		appAgent, err := agent.BuildAgent(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to build agent: %w", err)
		}

		prompt := strings.Join(args, " ")
		return ui.RunSinglePrompt(ctx, appAgent, prompt)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cfg = config.LoadConfig()

	rootCmd.PersistentFlags().StringVarP(&flagURL, "url", "u", cfg.BaseURL, "Ollama / OpenAI endpoint URL")
	rootCmd.PersistentFlags().StringVarP(&flagModel, "model", "m", cfg.ModelName, "Model name identifier")
	rootCmd.PersistentFlags().StringVarP(&flagAPIKey, "api-key", "k", cfg.APIKey, "Optional API key")
}

func applyFlagsToConfig() {
	if flagURL != "" {
		cfg.BaseURL = flagURL
	}
	if flagModel != "" {
		cfg.ModelName = flagModel
	}
	if flagAPIKey != "" {
		cfg.APIKey = flagAPIKey
	}
}
