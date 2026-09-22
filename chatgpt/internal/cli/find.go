// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type findItem struct {
	ID         string          `json:"id"`
	SourceType string          `json:"source_type"`
	Title      string          `json:"title"`
	Snippet    string          `json:"snippet"`
	UpdateTime json.Number     `json:"update_time"`
	MatchKind  string          `json:"match_kind"`
	Payload    json.RawMessage `json:"payload"`
}

type sourceStatus struct {
	SourceType string      `json:"source_type"`
	Status     string      `json:"status"`
	HasMore    bool        `json:"has_more"`
	DurationMS json.Number `json:"duration_ms"`
	ErrorCode  *string     `json:"error_code"`
}

type findView struct {
	Query          string         `json:"query"`
	Items          []findItem     `json:"items"`
	Cursor         string         `json:"cursor,omitempty"`
	HasMore        bool           `json:"has_more"`
	PartialResults bool           `json:"partial_results"`
	SourceStatuses []sourceStatus `json:"source_statuses"`
	PagesWalked    int            `json:"pages_walked"`
	MaxScanPages   int            `json:"max_scan_pages"`
	Note           string         `json:"note,omitempty"`
}

func newNovelFindCmd(flags *rootFlags) *cobra.Command {
	var flagLimit, flagMaxScanPages int
	var flagAll bool

	cmd := &cobra.Command{
		Use:   "find [query]",
		Short: "Server-side message search across all conversations (the Ctrl+K surface).",
		Long:  `Searches message content server-side via the same federated endpoint the web UI's Ctrl+K uses — message-level hits with snippets, conversation and message ids. This is the default search. Use framework 'search' for offline FTS over previously synced bodies when no session is available.`,
		Example: strings.Trim(`
  chatgpt-pp-cli find "haptics comparison" --agent --select items.title,items.snippet
  chatgpt-pp-cli find "camera" --limit 20 --json
`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "query=haptics"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "find")
			}
			if len(args) < 1 || strings.TrimSpace(args[0]) == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("query text is required"))
			}
			if cliutilDogfood() && flagMaxScanPages > 1 {
				flagMaxScanPages = 1
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			gc, err := newGptClient(flags)
			if err != nil {
				return err
			}
			view := findView{Query: args[0], Items: make([]findItem, 0), SourceStatuses: make([]sourceStatus, 0), MaxScanPages: flagMaxScanPages}
			cursor := ""
			for page := 1; page <= flagMaxScanPages; page++ {
				body := map[string]any{"query": args[0], "limit": flagLimit}
				if cursor != "" {
					body["cursor"] = cursor
				}
				raw, status, err := gc.Post(ctx, "/backend-api/global/search", body, nil)
				if err != nil || status != 200 {
					return fmt.Errorf("global search: HTTP %d: %v", status, err)
				}
				var resp struct {
					Items          []findItem     `json:"items"`
					Cursor         string         `json:"cursor"`
					PartialResults bool           `json:"partial_results"`
					SourceStatuses []sourceStatus `json:"source_statuses"`
				}
				if err := json.Unmarshal(raw, &resp); err != nil {
					return fmt.Errorf("parsing search response: %w", err)
				}
				view.PagesWalked = page
				view.Items = append(view.Items, resp.Items...)
				view.PartialResults = view.PartialResults || resp.PartialResults
				if page == 1 {
					view.SourceStatuses = resp.SourceStatuses
				}
				if !flagAll || resp.Cursor == "" || len(resp.Items) == 0 {
					view.Cursor = resp.Cursor
					view.HasMore = resp.Cursor != ""
					break
				}
				cursor = resp.Cursor
				if len(view.Items) >= flagLimit*flagMaxScanPages {
					view.Cursor = cursor
					view.HasMore = true
					break
				}
			}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			out := cmd.OutOrStdout()
			if len(view.Items) == 0 {
				fmt.Fprintln(out, "No matches found.")
				if view.PartialResults {
					fmt.Fprintln(out, "note: server reported partial results — some sources failed or timed out")
				}
				return nil
			}
			for _, it := range view.Items {
				fmt.Fprintf(out, "%s  %s\n    %s\n", it.Title, it.MatchKind, truncateStr(it.Snippet, 90))
			}
			if view.HasMore {
				fmt.Fprintln(out, "\nmore results available (pass --all or a cursor)")
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&flagLimit, "limit", 10, "results per page")
	cmd.Flags().IntVar(&flagMaxScanPages, "max-scan-pages", 3, "maximum cursor pages to walk")
	cmd.Flags().BoolVar(&flagAll, "all", false, "walk all cursor pages up to --max-scan-pages")
	return cmd
}

func cliutilDogfood() bool { return cliutilDogfoodImpl() }

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		addNovelCommandIfAbsent(root, newNovelFindCmd(flags))
	})
}
