// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command scaffold. Implement the RunE body before shipping.
// generate --force preserves implemented bodies; untouched TODO scaffolds may refresh.
// pp:data-source auto
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"github.com/spf13/cobra"
)

func newNovelEventsCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "events",
		Short:       "Live event context",
		Example:     "  moneycontrol-pp-cli events upcoming --sc-ids RI,INFY --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelEventsUpcomingCmd(flags))
	return cmd
}
