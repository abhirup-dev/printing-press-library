package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newScreenerEquityQueryCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "equity-query",
		Short:       "Deferred POST-based equity screener query (not executed).",
		Annotations: map[string]string{"pp:endpoint": "screener.equity_query", "pp:method": "POST", "pp:path": "/screener/query", "mcp:read-only": "false"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "screener equity-query")
			}
			return fmt.Errorf("screener equity-query is deferred: POST execution is not validated for the live-only read contract")
		},
	}
}
