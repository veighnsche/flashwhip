package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"flashwhip/pkg/agent"
	"flashwhip/pkg/config"
	"flashwhip/pkg/errors"
	"flashwhip/pkg/ui"
)


var (
	cfg        *config.Config
	flagURL    string
	flagModel  string
	flagAPIKey string
	flagCwd    string
	flagMaxTurns int
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
		activeCfg := applyFlagsToConfig()

		appAgent, err := agent.BuildAgent(ctx, activeCfg)
		if err != nil {
			return errors.Wrap(errors.ErrCodeAgentBuildFailed, "failed to build agent", err)
		}

		prompt := strings.Join(args, " ")
		return ui.RunSinglePrompt(ctx, appAgent, prompt, flagMaxTurns)
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
	rootCmd.PersistentFlags().StringVarP(&flagCwd, "cwd", "C", "", "Project root directory to operate in (defaults to current directory)")
	rootCmd.PersistentFlags().IntVarP(&flagMaxTurns, "max-turns", "t", 25, "Maximum agent turns per prompt (tool-call round-trips); 0 = unlimited")
}

func applyFlagsToConfig() *config.Config {
	c := *cfg
	if flagURL != "" {
		c.BaseURL = flagURL
	}
	if flagModel != "" {
		c.ModelName = flagModel
	}
	if flagAPIKey != "" {
		c.APIKey = flagAPIKey
	}
	if flagCwd != "" {
		abs, err := filepath.Abs(flagCwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", errors.Wrapf(errors.ErrCodeConfigInvalid, err, "invalid --cwd path %q", flagCwd))
			os.Exit(1)
		}
		if err := os.Chdir(abs); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", errors.Wrapf(errors.ErrCodeDirChangeFailed, err, "cannot change to directory %q", abs))
			os.Exit(1)
		}
		c.ProjectRoot = abs
	}
	return &c
}

