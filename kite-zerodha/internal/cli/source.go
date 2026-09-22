// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
package cli

import (
	"github.com/spf13/cobra"
)

func newNovelSourceCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "source",
		Short:       "Transport boundaries",
		Example:     "  kite-zerodha-pp-cli source status --agent --select sources",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelSourceStatusCmd(flags))
	return cmd
}
