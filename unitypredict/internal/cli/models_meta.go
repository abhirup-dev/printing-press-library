// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Absorb manifest row 16.
// pp:data-source live

package cli

// models meta set — safe metadata editing for the fields the console edits:
// name, short description, category, pricing note, active/public flags, alias.
// POST /api/models is a FULL-RECORD upsert: a partial body clobbers the record
// (including aiEngineConfig), so this command owns the GET -> mutate -> POST
// round-trip and verifies the write with a cache-bypassed re-read.

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newUptModelsMetaCmd(flags *rootFlags) *cobra.Command {
	var envFlag, name, description, category, pricingNote, alias string
	var active, public, iKnow bool
	cmd := &cobra.Command{
		Use:   "set <modelId>",
		Short: "Edit model metadata (name, description, category, pricing, flags)",
		Long: "Safe metadata edit via the full-record upsert discipline: GET the model, apply only the\n" +
			"flags you set, POST the whole record back, then verify with a fresh read. Mutating\n" +
			"individual fields directly against POST /api/models clobbers the record (including the\n" +
			"input/outcome contract). Long description is a hosted file — use `models describe set`.\n" +
			"Categories are an enum: Audio to Audio, Image to 3D, Image to Image, LLMs and Chatbots,\n" +
			"Speech to Text, Text to Audio, Text to Image, Text to Speech, Text to Video, Training,\n" +
			"Text Interpretation, Video to Video, Vision, Video/Image Interpretation, Other.",
		Example: "  unitypredict-pp-cli models meta set <modelId> --description \"Short SEO summary.\"\n" +
			"  unitypredict-pp-cli models meta set <modelId> --category \"Text Interpretation\" --active --public",
		Annotations: map[string]string{
			"pp:data-source": "live",
			"pp:happy-args":  "modelId=c6985b70-15e0-4ba9-8cb1-128200595a8f;--description=recon test note;--dry-run",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "models meta set")
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("modelId is required\nUsage: %s <modelId>", cmd.CommandPath()))
			}
			activeSet := cmd.Flags().Changed("active")
			publicSet := cmd.Flags().Changed("public")
			if name == "" && description == "" && category == "" && pricingNote == "" && alias == "" &&
				!activeSet && !publicSet {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("nothing to set: pass at least one of --name, --description, --category, --pricing-note, --alias, --active, --public"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			modelID := args[0]
			env := resolveUptEnv(envFlag)
			if err := guardUptProdMutation(env, "edit model metadata", iKnow); err != nil {
				return err
			}
			c, err := uptEnvClient(flags, env)
			if err != nil {
				return err
			}
			recordJSON, err := c.GetNoCache(ctx, "/api/models/"+modelID, nil)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			var record map[string]any
			if err := json.Unmarshal(recordJSON, &record); err != nil {
				return fmt.Errorf("parsing model record: %w", err)
			}
			applied := map[string]any{}
			if name != "" {
				record["modelName"] = name
				applied["modelName"] = name
			}
			if description != "" {
				record["modelDescription"] = description
				applied["modelDescription"] = description
			}
			if category != "" {
				if !uptValidCategory(category) {
					return usageErr(fmt.Errorf("--category %q is not one of the platform's enum values (see --help)", category))
				}
				record["modelCategory"] = category
				applied["modelCategory"] = category
			}
			if alias != "" {
				record["modelAlias"] = alias
				applied["modelAlias"] = alias
			}
			if pricingNote != "" {
				pi, _ := record["pricingInfo"].(map[string]any)
				if pi == nil {
					pi = map[string]any{}
				}
				pi["servicePricingInfo"] = pricingNote
				record["pricingInfo"] = pi
				applied["pricingInfo.servicePricingInfo"] = pricingNote
			}
			if activeSet {
				record["active"] = active
				applied["active"] = active
			}
			if publicSet {
				record["isPublic"] = public
				applied["isPublic"] = public
			}
			payload, err := json.Marshal(record)
			if err != nil {
				return err
			}
			if _, _, err := c.Post(ctx, "/api/models", json.RawMessage(payload)); err != nil {
				return fmt.Errorf("upserting model record: %w", err)
			}
			// Verify with a cache-bypassed re-read (read-after-write can serve
			// stale data in a warm session).
			verify, err := c.GetNoCache(ctx, "/api/models/"+modelID, nil)
			verified := false
			if err == nil {
				var v map[string]any
				if json.Unmarshal(verify, &v) == nil {
					verified = true
					for k, want := range applied {
						if k == "pricingInfo.servicePricingInfo" {
							if got, _, _ := uptDig(v, "pricingInfo", "servicePricingInfo"); got != want {
								verified = false
							}
						} else if v[k] != want {
							verified = false
						}
					}
				}
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: post-write verification read failed: %v\n", err)
			}
			view := struct {
				ModelID  string         `json:"model_id"`
				Env      string         `json:"env"`
				Applied  map[string]any `json:"applied"`
				Verified bool           `json:"verified"`
			}{modelID, env, applied, verified}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "model %s updated (%d field(s)); verified: %v\n", modelID, len(applied), verified)
			for k, v := range applied {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s = %v\n", k, v)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFlag, "env", "", "Tenant: dev or prod (default $UNITYPREDICT_ENV or dev)")
	cmd.Flags().StringVar(&name, "name", "", "Model name (modelName)")
	cmd.Flags().StringVar(&description, "description", "", "Short description (modelDescription); the long description is a hosted file, see models describe set")
	cmd.Flags().StringVar(&category, "category", "", "Model category (enum — see help)")
	cmd.Flags().StringVar(&pricingNote, "pricing-note", "", "Pricing text (pricingInfo.servicePricingInfo)")
	cmd.Flags().StringVar(&alias, "alias", "", "Model alias (modelAlias)")
	cmd.Flags().BoolVar(&active, "active", false, "Set the model active (owner predicts work either way)")
	cmd.Flags().BoolVar(&public, "public", false, "Set the model public (end-user visibility)")
	cmd.Flags().BoolVar(&iKnow, "i-know", false, "Allow the mutation against prod (off by default: prod is read-only)")
	return cmd
}

func uptValidCategory(c string) bool {
	switch c {
	case "Audio to Audio", "Image to 3D", "Image to Image", "LLMs and Chatbots", "Speech to Text",
		"Text to Audio", "Text to Image", "Text to Speech", "Text to Video", "Training",
		"Text Interpretation", "Video to Video", "Vision", "Video/Image Interpretation", "Other":
		return true
	}
	return false
}

func uptDig(m map[string]any, path ...string) (any, bool, error) {
	cur := any(m)
	for _, k := range path {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf("path %v not found", path)
		}
		cur, ok = mm[k]
		if !ok {
			return nil, false, fmt.Errorf("path %v not found", path)
		}
	}
	return cur, true, nil
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		if modelsCmd, _, err := root.Find([]string{"models"}); err == nil {
			metaParent := &cobra.Command{
				Use:   "meta",
				Short: "Model metadata (safe full-record edit)",
				Example: "  unitypredict-pp-cli models meta set <modelId> --description \"...\"",
			}
			metaParent.AddCommand(newUptModelsMetaCmd(flags))
			addNovelCommandIfAbsent(modelsCmd, metaParent)
		}
	})
}
