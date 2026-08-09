package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"flashwhip/pkg/agent"
	"flashwhip/pkg/ui"
)

var flagSession string

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat REPL session",
	Long:  `Launches an interactive multi-turn terminal conversation with persistent prompt history (~/.flashwhip_history) and embedded SQLite persistence (~/.flashwhip/flashwhip.db).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		activeCfg := applyFlagsToConfig()

		appAgent, err := agent.BuildAgent(ctx, activeCfg)
		if err != nil {
			return fmt.Errorf("failed to build agent: %w", err)
		}

		return ui.RunInteractiveREPL(ctx, appAgent, activeCfg, flagSession, flagMaxTurns)
	},
}

func init() {
	chatCmd.Flags().StringVarP(&flagSession, "session", "s", "", "Session ID to resume or attach to")
	rootCmd.AddCommand(chatCmd)
}
