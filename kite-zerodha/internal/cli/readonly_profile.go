package cli

import "github.com/spf13/cobra"

// newReadOnlyProfileCmd preserves inspection of an existing run profile while
// intentionally omitting profile save/delete mutations from this stateless
// broker CLI.
func newReadOnlyProfileCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "profile",
		Short:       "Inspect named run profiles (read-only)",
		Annotations: map[string]string{"pp:parent-group": "true", "mcp:read-only": "true"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	cmd.AddCommand(newProfileUseCmd(flags))
	cmd.AddCommand(newProfileListCmd(flags))
	cmd.AddCommand(newProfileShowCmd(flags))
	return cmd
}
