// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
package cli

import (
	"github.com/spf13/cobra"
)

func newNovelPortfolioCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "portfolio",
		Short:       "Live broker analytics",
		Example:     "  kite-zerodha-pp-cli portfolio summary --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelPortfolioSummaryCmd(flags))
	return cmd
}
