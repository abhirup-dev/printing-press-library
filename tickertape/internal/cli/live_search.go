package cli

import "github.com/spf13/cobra"

// newLiveSearchCmd exposes the public Tickertape suggestion endpoint at the
// short, user-facing `search` path. It intentionally delegates to the
// generated live endpoint command so the response envelope and provenance
// handling stay consistent.
func newLiveSearchCmd(flags *rootFlags) *cobra.Command {
	cmd := newSearchResourcePromotedCmd(flags)
	cmd.Use = "search <q>"
	cmd.Aliases = []string{"find"}
	cmd.Short = "Search live Tickertape asset suggestions."
	return cmd
}
