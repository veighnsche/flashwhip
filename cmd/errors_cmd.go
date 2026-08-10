package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"flashwhip/pkg/errors"
	"flashwhip/pkg/ui"
)

var errorsCmd = &cobra.Command{
	Use:     "errors [code]",
	Aliases: []string{"error", "errcode"},
	Short:   "List or inspect stable error codes and diagnostics",
	Long:    `Displays a catalog of all stable error codes mapped across Flashwhip, or looks up detailed diagnostics and remedies for a specific numeric code (e.g., 'flashwhip errors 1001' or 'flashwhip errors FW-5002').`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			specs := errors.All()
			var sb strings.Builder
			sb.WriteString(ui.Bold.Render("Flashwhip Mapped Stable Error Codes Catalog\n\n"))
			sb.WriteString(fmt.Sprintf("%-10s %-28s %-15s %s\n", "CODE", "NAME", "CATEGORY", "DESCRIPTION"))
			sb.WriteString(strings.Repeat("─", 80) + "\n")
			for _, spec := range specs {
				sb.WriteString(fmt.Sprintf("%-10s %-28s %-15s %s\n",
					fmt.Sprintf("FW-%04d", spec.Code),
					spec.Name,
					spec.Category,
					spec.Description,
				))
			}
			fmt.Print(sb.String())
			return nil
		}

		rawCode := strings.TrimPrefix(strings.TrimSpace(args[0]), "FW-")
		rawCode = strings.TrimPrefix(rawCode, "fw-")
		num, err := strconv.Atoi(rawCode)
		if err != nil {
			return errors.Wrapf(errors.ErrCodeConfigInvalid, err, "invalid error code %q", args[0])
		}

		spec, ok := errors.Lookup(errors.Code(num))
		if !ok {
			return errors.Errorf(errors.ErrCodeConfigInvalid, "error code FW-%04d is not mapped in the registry", num)
		}

		var sb strings.Builder
		sb.WriteString(ui.Bold.Render(fmt.Sprintf("Error Specification: FW-%04d (%s)\n", spec.Code, spec.Name)))
		sb.WriteString(strings.Repeat("═", 55) + "\n")
		sb.WriteString(fmt.Sprintf("Category:    %s\n", spec.Category))
		sb.WriteString(fmt.Sprintf("Description: %s\n", spec.Description))
		sb.WriteString(fmt.Sprintf("Remedy:      %s\n", spec.Remedy))
		fmt.Print(sb.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(errorsCmd)
}
