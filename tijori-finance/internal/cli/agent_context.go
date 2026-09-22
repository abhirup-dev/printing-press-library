// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// agentContextSchemaVersion is the stable machine-readable discovery contract.
const agentContextSchemaVersion = "1"

type agentContext struct {
	SchemaVersion string                `json:"schema_version"`
	CLI           agentContextCLI       `json:"cli"`
	Auth          agentContextAuth      `json:"auth"`
	Commands      []agentContextCommand `json:"commands"`
}

type agentContextCLI struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type agentContextAuth struct {
	Mode    string                   `json:"mode"`
	EnvVars []agentContextAuthEnvVar `json:"env_vars"`
}

type agentContextAuthEnvVar struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Required  bool   `json:"required"`
	Sensitive bool   `json:"sensitive"`
}

type agentContextCommand struct {
	Name        string                `json:"name"`
	Use         string                `json:"use,omitempty"`
	Short       string                `json:"short,omitempty"`
	Annotations map[string]string     `json:"annotations,omitempty"`
	Flags       []agentContextFlag    `json:"flags,omitempty"`
	Runnable    bool                  `json:"runnable,omitempty"`
	Subcommands []agentContextCommand `json:"subcommands,omitempty"`
}

type agentContextFlag struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Usage   string `json:"usage,omitempty"`
	Default string `json:"default,omitempty"`
}

func newAgentContextCmd(rootCmd *cobra.Command) *cobra.Command {
	var pretty bool
	cmd := &cobra.Command{
		Use:    "agent-context",
		Short:  "Emit structured JSON describing this CLI for agents",
		Hidden: true,
		Annotations: map[string]string{
			"mcp:read-only": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			enc := json.NewEncoder(cmd.OutOrStdout())
			if pretty {
				enc.SetIndent("", "  ")
			}
			return enc.Encode(agentContext{
				SchemaVersion: agentContextSchemaVersion,
				CLI: agentContextCLI{
					Name:        rootCmd.Name(),
					Description: rootCmd.Short,
					Version:     rootCmd.Version,
				},
				Auth: agentContextAuth{
					Mode: "session_cookie",
					EnvVars: []agentContextAuthEnvVar{
						{Name: "TIJORI_FINANCE_SESSIONID", Kind: "session_cookie", Sensitive: true},
						{Name: "TIJORI_FINANCE_CSRFTOKEN", Kind: "csrf_token", Sensitive: true},
					},
				},
				Commands: collectAgentCommands(rootCmd),
			})
		},
	}
	cmd.Flags().BoolVar(&pretty, "pretty", false, "indent JSON output for human reading")
	return cmd
}

func collectAgentCommands(rootCmd *cobra.Command) []agentContextCommand {
	children := rootCmd.Commands()
	sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })
	out := make([]agentContextCommand, 0, len(children))
	for _, sub := range children {
		if sub.Name() == "agent-context" {
			continue
		}
		entry := agentContextCommand{
			Name: sub.Name(), Use: sub.Use, Short: sub.Short, Runnable: sub.Runnable(),
		}
		if len(sub.Annotations) > 0 {
			entry.Annotations = make(map[string]string, len(sub.Annotations))
			for key, value := range sub.Annotations {
				entry.Annotations[key] = value
			}
		}
		sub.Flags().VisitAll(func(flag *pflag.Flag) {
			entry.Flags = append(entry.Flags, agentContextFlag{Name: flag.Name, Type: flag.Value.Type(), Usage: flag.Usage, Default: flag.DefValue})
		})
		sort.Slice(entry.Flags, func(i, j int) bool { return entry.Flags[i].Name < entry.Flags[j].Name })
		if len(sub.Commands()) > 0 {
			entry.Subcommands = collectAgentCommands(sub)
		}
		out = append(out, entry)
	}
	return out
}
