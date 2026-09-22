// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
package cli

import (
	"github.com/spf13/cobra"
)

func newNovelActivityCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "activity",
		Short:       "Work with activity",
		Example:     "  kite-zerodha-pp-cli activity buys --agent --select trades",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelActivityBuysCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelActivitySellsCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelActivityFillsCmd(flags))
	return cmd
}
