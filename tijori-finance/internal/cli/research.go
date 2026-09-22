// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command scaffold. Implement the RunE body before shipping.
// generate --force preserves implemented bodies; untouched TODO scaffolds may refresh.
// pp:data-source auto
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"github.com/spf13/cobra"
)

func newNovelResearchCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "research",
		Short:       "Work with research",
		Example:     "  tijori-finance-pp-cli research changes tata-steel-limited --from Mar-25 --to Mar-26 --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelResearchChangesCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelResearchContextCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelResearchInputContextCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelResearchOwnershipCheckCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelResearchPeersCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelResearchResultPackCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelResearchReviewQueueCmd(flags))
	return cmd
}
