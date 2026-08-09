package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"flashwhip/pkg/agent"
	"flashwhip/pkg/ui"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat REPL session",
	Long:  `Launches an interactive multi-turn terminal conversation with persistent prompt history (~/.flashwhip_history).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		applyFlagsToConfig()

		appAgent, err := agent.BuildAgent(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to build agent: %w", err)
		}

		return ui.RunInteractiveREPL(ctx, appAgent, cfg)
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
