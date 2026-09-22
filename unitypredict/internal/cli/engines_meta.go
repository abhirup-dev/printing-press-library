// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Absorb manifest row 20.
// pp:data-source live

package cli

// engines meta set — safe engine metadata editing (name, description).
// POST /api/engines is a full-record upsert: a partial body clobbers platform,
// memory, batch, and chaining config. GET -> mutate -> POST -> verify.
// Metadata-only: changing an engine record does not trigger a redeploy, but
// platform/resource flags feed the next deploy, so those stay on
// `engines create-engines` where the full record is authored deliberately.

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newUptEnginesMetaCmd(flags *rootFlags) *cobra.Command {
	var envFlag, name, description string
	var iKnow bool
	cmd := &cobra.Command{
		Use:   "set <engineId>",
		Short: "Edit engine metadata (name, description)",
		Long: "Safe engine metadata edit via GET -> mutate -> POST full record. POST /api/engines with a\n" +
			"partial body clobbers platform/memory/batch/chaining config, so never edit engine fields\n" +
			"by hand-posting. Metadata edits do not trigger a redeploy. Engine resource and platform\n" +
			"configuration is authored with `engines create-engines` (full record), not here.",
		Example: "  unitypredict-pp-cli engines meta set b5132be0-a373-4f74-bedf-fabb17dc0342 --description \"GPU smoke test\"\n" +
			"  unitypredict-pp-cli engines meta set b5132be0-a373-4f74-bedf-fabb17dc0342 --name \"GPU Hello World v2\"",
		Annotations: map[string]string{
			"pp:data-source": "live",
			"pp:happy-args":  "engineId=b5132be0-a373-4f74-bedf-fabb17dc0342;--description=recon test note;--dry-run",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "engines meta set")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("engineId is required\nUsage: %s <engineId>", cmd.CommandPath()))
			}
			if name == "" && description == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("nothing to set: pass --name and/or --description"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			engineID := stripEnginePrefix(args[0])
			env := resolveUptEnv(envFlag)
			if err := guardUptProdMutation(env, "edit engine metadata", iKnow); err != nil {
				return err
			}
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			recordJSON, err := c.GetNoCache(ctx, "/api/engines/"+engineID, map[string]string{"loadRD": "true", "modeledit": "true"})
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			var record map[string]any
			if err := json.Unmarshal(recordJSON, &record); err != nil {
				return fmt.Errorf("parsing engine record: %w", err)
			}
			applied := map[string]any{}
			if name != "" {
				record["engineName"] = name
				applied["engineName"] = name
			}
			if description != "" {
				record["engineDescription"] = description
				applied["engineDescription"] = description
			}
			payload, err := json.Marshal(record)
			if err != nil {
				return err
			}
			if _, _, err := c.Post(ctx, "/api/engines", json.RawMessage(payload)); err != nil {
				return fmt.Errorf("upserting engine record: %w", err)
			}
			verify, err := c.GetNoCache(ctx, "/api/engines/"+engineID, nil)
			verified := false
			if err == nil {
				var v map[string]any
				if json.Unmarshal(verify, &v) == nil {
					verified = true
					for k, want := range applied {
						if v[k] != want {
							verified = false
						}
					}
				}
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: post-write verification read failed: %v\n", err)
			}
			view := struct {
				EngineID string         `json:"engine_id"`
				Env      string         `json:"env"`
				Applied  map[string]any `json:"applied"`
				Verified bool           `json:"verified"`
			}{engineID, env, applied, verified}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "engine %s updated (%d field(s)); verified: %v\n", engineID, len(applied), verified)
			for k, v := range applied {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s = %v\n", k, v)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	cmd.Flags().StringVar(&name, "name", "", "Engine name (engineName)")
	cmd.Flags().StringVar(&description, "description", "", "Engine description (engineDescription)")
	cmd.Flags().BoolVar(&iKnow, "i-know", false, "Allow the mutation against prod (off by default: prod is read-only)")
	return cmd
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if enginesCmd, _, err := root.Find([]string{"engines"}); err == nil {
			metaParent := &cobra.Command{
				Use:   "meta",
				Short: "Engine metadata (safe full-record edit)",
				Example: "  unitypredict-pp-cli engines meta set <engineId> --description \"...\"",
			}
			metaParent.AddCommand(newUptEnginesMetaCmd(flags))
			addNovelCommandIfAbsent(enginesCmd, metaParent)
		}
	})
}
