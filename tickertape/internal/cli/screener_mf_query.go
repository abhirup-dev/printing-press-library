package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newScreenerMfQueryCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "mf-query",
		Short:       "Deferred POST-based mutual-fund screener query (not executed).",
		Annotations: map[string]string{"pp:endpoint": "screener.mf_query", "pp:method": "POST", "pp:path": "/mf-screener/query", "mcp:read-only": "false"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "screener mf-query")
			}
			return fmt.Errorf("screener mf-query is deferred: POST execution is not validated for the live-only read contract")
		},
	}
}
