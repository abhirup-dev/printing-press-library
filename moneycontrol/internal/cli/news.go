// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command scaffold. Implement the RunE body before shipping.
// generate --force preserves implemented bodies; untouched TODO scaffolds may refresh.
// pp:data-source auto
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"github.com/spf13/cobra"
)

func newNovelNewsCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "news",
		Short:       "Live news intelligence",
		Example:     "  moneycontrol-pp-cli news timeline --sc-id RI --limit 20 --agent --select articles.title,articles.timestamp,articles.url",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelNewsTimelineCmd(flags))
	return cmd
}
