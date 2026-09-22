// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel command (printing-press preserved file). Manifest row T9.
// pp:data-source live

package cli

// diff model|engine <id> — compare a record across the two UnityPredict
// tenants (dev and prod are separate Auth0 tenants with separate keys; nothing
// today can query both in one step). Volatile fields (timestamps, per-env
// state like stateInfo, counters) are excluded unless --raw.

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

// uptVolatileFields are per-env or churn-by-the-second; excluded from diff
// unless --raw. modelLongDescription embeds the env host so it is normalized
// instead of excluded (see uptNormalizeForDiff).
var uptVolatileFields = map[string]bool{
	"createdDate": true, "modifiedDate": true, "permissions": true,
	"stateInfo": true, "inferenceTimeStats": true, "lastBuildResults": true,
	"buildNumber": true, "deployedBuildNumber": true, "files": true,
	"totalInferences": true, "totalViews": true, "engineStatus": true,
	"performanceStats": true, "performanceScore": true, "totalCost": true,
}

// uptNormalizeForDiff neutralizes per-env URL hosts inside field values.
func uptNormalizeForDiff(in string) string {
	out := strings.ReplaceAll(in, "api.dev.unitypredict.net", "api.{ENV}.unitypredict")
	out = strings.ReplaceAll(out, "api.prod.unitypredict.com", "api.{ENV}.unitypredict")
	out = strings.ReplaceAll(out, "unitypredict-dev-", "unitypredict-{ENV}-")
	out = strings.ReplaceAll(out, "unitypredict-prod-", "unitypredict-{ENV}-")
	return out
}

// uptFlattenJSON collapses a decoded JSON record into dot-path -> leaf string.
func uptFlattenJSON(prefix string, v any, out map[string]string) {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			uptFlattenJSON(uptJoinPath(prefix, k), t[k], out)
		}
	case []any:
		for i, item := range t {
			uptFlattenJSON(fmt.Sprintf("%s[%d]", prefix, i), item, out)
		}
	case nil:
		out[prefix] = "<null>"
	default:
		out[prefix] = uptNormalizeForDiff(fmt.Sprint(t))
	}
}

func uptJoinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func newUptDiffCmd(flags *rootFlags) *cobra.Command {
	var envA, envB string
	var raw bool

	fetchRecord := func(cmd *cobra.Command, entity, id, env string) (map[string]any, bool, error) {
		c, err := uptEnvClient(flags, env)
		if err != nil {
			return nil, false, err
		}
		ctx, cancel := boundCtx(cmd.Context(), flags)
		defer cancel()
		path := "/api/" + entity + "s/" + stripEnginePrefix(id)
		data, err := c.GetNoCache(ctx, path, nil)
		if err != nil {
			if strings.Contains(err.Error(), "HTTP 404") {
				return nil, false, nil
			}
			return nil, false, err
		}
		var rec map[string]any
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil, false, fmt.Errorf("parsing %s record from %s: %w", entity, env, err)
		}
		return rec, true, nil
	}

	makeRun := func(entity, idField string) func(cmd *cobra.Command, args []string) error {
		return func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "diff "+entity)
			}
			if len(args) < 1 || args[0] == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("%s is required\nUsage: %s <%s>", idField, cmd.CommandPath(), idField))
			}
			envA, envB = resolveUptEnv(envA), resolveUptEnv(envB)
			if envA == envB {
				return usageErr(fmt.Errorf("--env-a and --env-b must differ (both %q)", envA))
			}
			id := args[0]
			// Shape-check the id: platform ids are UUIDs (engines may carry an
			// AppEngineDefinition- prefix). A malformed id is invalid input, not
			// "absent in both tenants" — an empty diff would mislead.
			if !uptIsUUID(stripEnginePrefix(id)) {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("%s %q is not a valid id (expected a UUID, engines may carry an AppEngineDefinition- prefix)", idField, id))
			}

			type fetchOut struct {
				rec    map[string]any
				present bool
				err    error
			}
			outCh := make(chan fetchOut, 2)
			var wg sync.WaitGroup
			for _, env := range []string{envA, envB} {
				wg.Add(1)
				go func(env string) {
					defer wg.Done()
					rec, present, err := fetchRecord(cmd, entity, id, env)
					outCh <- fetchOut{rec, present, err}
				}(env)
			}
			wg.Wait()
			close(outCh)
			// Preserve flag order envA then envB.
			outs := make([]fetchOut, 0, 2)
			for o := range outCh {
				outs = append(outs, o)
			}
			var a, b fetchOut
			if len(outs) == 2 {
				a, b = outs[0], outs[1]
			} else if len(outs) == 1 {
				a = outs[0]
			}
			if a.err != nil {
				return fmt.Errorf("fetching %s from %s: %w", entity, envA, a.err)
			}
			if b.err != nil {
				return fmt.Errorf("fetching %s from %s: %w", entity, envB, b.err)
			}

			flatA, flatB := map[string]string{}, map[string]string{}
			if a.present {
				uptFlattenJSON("", a.rec, flatA)
			}
			if b.present {
				uptFlattenJSON("", b.rec, flatB)
			}
			paths := map[string]bool{}
			for p := range flatA {
				paths[p] = true
			}
			for p := range flatB {
				paths[p] = true
			}
			sorted := make([]string, 0, len(paths))
			for p := range paths {
				if !raw && uptVolatileFields[topSegment(p)] {
					continue
				}
				sorted = append(sorted, p)
			}
			sort.Strings(sorted)

			type fieldDiff struct {
				Path    string `json:"path"`
				A       string `json:"a,omitempty"`
				B       string `json:"b,omitempty"`
				Absent  string `json:"absent_in,omitempty"`
				Differs bool   `json:"differs"`
			}
			fields := make([]fieldDiff, 0, len(sorted))
			diffCount := 0
			for _, p := range sorted {
				va, inA := flatA[p]
				vb, inB := flatB[p]
				switch {
				case inA && !inB:
					fields = append(fields, fieldDiff{Path: p, A: va, Absent: envB, Differs: true})
					diffCount++
				case !inA && inB:
					fields = append(fields, fieldDiff{Path: p, B: vb, Absent: envA, Differs: true})
					diffCount++
				case va != vb:
					fields = append(fields, fieldDiff{Path: p, A: va, B: vb, Differs: true})
					diffCount++
				default:
					fields = append(fields, fieldDiff{Path: p, A: va, B: vb, Differs: false})
				}
			}
			view := struct {
				Entity     string      `json:"entity"`
				ID         string      `json:"id"`
				A          string      `json:"a"`
				B          string      `json:"b"`
				PresentA   bool        `json:"present_a"`
				PresentB   bool        `json:"present_b"`
				Raw        bool        `json:"raw"`
				FieldCount int         `json:"field_count"`
				DiffCount  int         `json:"diff_count"`
				Fields     []fieldDiff `json:"fields"`
			}{entity, id, envA, envB, a.present, b.present, raw, len(fields), diffCount, fields}

			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s: dev-present=%v prod-present=%v — %d differing field(s) of %d compared\n\n",
				entity, id, a.present, b.present, diffCount, len(fields))
			rows := make([][]string, 0, diffCount)
			for _, f := range fields {
				if !f.Differs {
					continue
				}
				rows = append(rows, []string{f.Path, truncateForDisplay(f.A, 60), truncateForDisplay(f.B, 60), f.Absent})
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No differences in compared fields (use --raw to include volatile fields).")
				return nil
			}
			return flags.printTable(cmd, []string{"field", envA, envB, "absentIn"}, rows)
		}
	}

	modelCmd := &cobra.Command{
		Use:       "model <modelId>",
		Short:     "Diff a model record between dev and prod",
		Example:   "  unitypredict-pp-cli diff model c6985b70-15e0-4ba9-8cb1-128200595a8f",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
			"pp:happy-args":  "modelId=c6985b70-15e0-4ba9-8cb1-128200595a8f",
		},
		RunE: makeRun("model", "modelId"),
	}
	engineCmd := &cobra.Command{
		Use:       "engine <engineId>",
		Short:     "Diff an engine record between dev and prod",
		Example:   "  unitypredict-pp-cli diff engine b5132be0-a373-4f74-bedf-fabb17dc0342",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
			"pp:happy-args":  "engineId=b5132be0-a373-4f74-bedf-fabb17dc0342",
		},
		RunE: makeRun("engine", "engineId"),
	}
	parent := &cobra.Command{
		Use:   "diff",
		Short: "Compare a model or engine across dev and prod tenants",
		Example: "  unitypredict-pp-cli diff model <modelId>\n  unitypredict-pp-cli diff engine <engineId> --raw",
	}
	for _, sub := range []*cobra.Command{modelCmd, engineCmd} {
		sub.Flags().StringVar(&envA, "env-a", "dev", "First environment to compare")
		sub.Flags().StringVar(&envB, "env-b", "prod", "Second environment to compare")
		sub.Flags().BoolVar(&raw, "raw", false, "Include volatile fields (timestamps, stateInfo, counters)")
		parent.AddCommand(sub)
	}
	return parent
}

func topSegment(path string) string {
	if i := strings.IndexByte(path, '.'); i >= 0 {
		return path[:i]
	}
	if i := strings.IndexByte(path, '['); i >= 0 {
		return path[:i]
	}
	return path
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		addNovelCommandIfAbsent(root, newUptDiffCmd(flags))
	})
}
