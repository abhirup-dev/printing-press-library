// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command scaffold. Implement the RunE body before shipping.
// generate --force preserves implemented bodies; untouched TODO scaffolds may refresh.
// pp:data-source auto
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"github.com/spf13/cobra"
)

func newNovelAuthCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "auth",
		Short:       "Reachability mitigation",
		Example:     "  tijori-finance-pp-cli auth capabilities --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelAuthCapabilitiesCmd(flags))
	for _, child := range newNovelAuthSessionCommands(flags) {
		addNovelCommandIfAbsent(cmd, child)
	}
	return cmd
}
