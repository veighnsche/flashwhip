package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"flashwhip/pkg/db"
	"flashwhip/pkg/errors"
	"flashwhip/pkg/ui"
)

var sessionsCmd = &cobra.Command{
	Use:     "sessions",
	Aliases: []string{"list", "history"},
	Short:   "List stored conversation sessions",
	Long:    `Displays a stylized table of past multi-turn conversation sessions saved in embedded SQLite (~/.flashwhip/flashwhip.db).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.DefaultDB()
		if err != nil {
			return errors.Wrap(errors.ErrCodeDBOpenFailed, "failed to open database", err)
		}

		sessions, err := database.ListSessions()
		if err != nil {
			return errors.Wrap(errors.ErrCodeDBQueryFailed, "failed to list sessions", err)
		}

		fmt.Println(ui.RenderSessionList(sessions))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sessionsCmd)
}
