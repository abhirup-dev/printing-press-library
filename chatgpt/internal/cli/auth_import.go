// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"
	"github.com/spf13/cobra"

	"chatgpt-pp-cli/internal/cliutil"
	"chatgpt-pp-cli/internal/gptapi"
)

func newAuthImportCmd(flags *rootFlags) *cobra.Command {
	var flagCookies, flagCodex string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a browser session (cookies) or Codex login and mint a bearer token.",
		Long:  `Imports your chatgpt.com session from a Netscape-format cookie export (any browser extension can produce one) or reuses your codex login, then mints the short-lived bearer token via /api/auth/session and saves it to the CLI config. Cookies outlive the 10-day token TTL, so the CLI can keep re-minting until the session itself expires. Secret values are stored with 0600 permissions and never appear in output.`,
		Example: stringsTrimNl(`
  chatgpt-pp-cli auth import --cookies cookies.txt
  chatgpt-pp-cli auth import --codex
`),
		Annotations: map[string]string{"mcp:read-only": "false"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if cliutil.IsAnyHarness() {
				return writeHarnessRefusal(cmd.OutOrStdout(), flags, "import session credentials")
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "auth import")
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			if flagCookies != "" {
				n, err := gptapi.ImportCookieJar(flagCookies)
				if err != nil {
					return usageErr(err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "imported %d chatgpt.com cookies\n", n)
			} else if flagCodex == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("pass --cookies <file> or --codex"))
			}
			gc, err := newGptClient(flags)
			if err != nil {
				return err
			}
			if flagCodex != "" {
				// Codex path: no cookie mint; verify the token works.
				state := gc.State()
				if !state.CodexAuth {
					return fmt.Errorf("no ~/.codex/auth.json token found — run 'codex login' first")
				}
				fmt.Fprintln(cmd.OutOrStdout(), "codex auth detected; CHATGPT_TOKEN fallback will use it")
				return nil
			}
			expiry, err := gc.MintNow(ctx)
			if err != nil {
				return fmt.Errorf("minting bearer: %w", err)
			}
			st := gc.State()
			return printJSONFiltered(cmd.OutOrStdout(), map[string]any{
				"imported":       true,
				"bearer_expiry":  expiry,
				"has_cookie_jar": st.HasCookieJar,
				"source":         st.Source,
			}, flags)
		},
	}
	cmd.Flags().StringVar(&flagCookies, "cookies", "", "Netscape-format cookie export file for chatgpt.com")
	cmd.Flags().StringVar(&flagCodex, "codex", "", "reuse ~/.codex/auth.json (flag value unused; pass any non-empty string or '1')")
	return cmd
}

func stringsTrimNl(s string) string {
	if len(s) > 0 && s[0] == '\n' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return s
}
