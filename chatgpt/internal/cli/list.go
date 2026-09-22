// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source auto

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"chatgpt-pp-cli/internal/cliutil"
	"chatgpt-pp-cli/internal/gptconv"
	"chatgpt-pp-cli/internal/store"
)

type listRow struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Created        string `json:"created"`
	Updated        string `json:"updated"`
	UserTurns      int    `json:"user_turn_count,omitempty"`
	AssistantTurns int    `json:"assistant_turn_count,omitempty"`
	TotalMessages  int    `json:"total_message_count,omitempty"`
	SpanSeconds    int64  `json:"span_seconds,omitempty"`
	ContentBytes   int    `json:"content_bytes,omitempty"`
	Model          string `json:"model,omitempty"`
	Archived       bool   `json:"archived"`
}

type listView struct {
	Items        []listRow `json:"items"`
	ScannedItems int       `json:"scanned_items"`
	MaxScanPages int       `json:"max_scan_pages"`
	Hydrated     int       `json:"hydrated,omitempty"`
	Note         string    `json:"note,omitempty"`
}

// parseDurationExt extends time.ParseDuration with day (d) and week (w) units
// because Go's native parser rejects "30d".
func parseDurationExt(s string) (time.Duration, error) {
	if n := len(s); n > 1 {
		switch s[n-1] {
		case 'd':
			if d, err := time.ParseDuration(s[:n-1] + "h"); err == nil {
				return d * 24, nil
			}
		case 'w':
			if d, err := time.ParseDuration(s[:n-1] + "h"); err == nil {
				return d * 24 * 7, nil
			}
		}
	}
	return time.ParseDuration(s)
}

// parseWindow accepts a duration (30d, 2w, 12h, 15m) or an ISO date
// (2026-09-01) and returns the cutoff time it implies.
func parseWindow(s string, now time.Time) (time.Time, bool, error) {
	if s == "" {
		return time.Time{}, false, nil
	}
	if d, err := parseDurationExt(s); err == nil && d > 0 {
		return now.Add(-d), true, nil
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true, nil
		}
	}
	return time.Time{}, false, fmt.Errorf("--since/--before/--updated-within must be a duration (30d, 2w, 12h, 15m) or ISO date (2026-09-01)")
}

func newNovelListCmd(flags *rootFlags) *cobra.Command {
	var flagSince, flagBefore, flagUpdatedWithin, flagOrder string
	var flagHydrate, flagIncludeArchived bool
	var flagLimit, flagMaxScanPages, flagMaxHydrate, flagHydrateBudget int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Chronological conversation inventory with turn counts, span, and freshness.",
		Long:  `Lists conversations in chronological order with normalized timestamps and — with --hydrate — turn counts and time span computed from the full conversation bodies (the list API returns no counts). Server ordering is update-time only; --order created sorts client-side. Use 'conversations list' for the raw endpoint mirror.`,
		Example: strings.Trim(`
  chatgpt-pp-cli list --since 30d --order created --hydrate --agent
  chatgpt-pp-cli list --updated-within 48h --hydrate --json
  chatgpt-pp-cli list --before 2026-09-01 --limit 20
`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto", "pp:novel-scaffold": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "list")
			}
			now := time.Now().UTC()
			sinceCut, hasSince, err := parseWindow(flagSince, now)
			if err != nil {
				_ = cmd.Usage()
				return usageErr(err)
			}
			beforeCut, hasBefore, err := parseWindow(flagBefore, now)
			if err != nil {
				_ = cmd.Usage()
				return usageErr(err)
			}
			withinCut, hasWithin, err := parseWindow(flagUpdatedWithin, now)
			if err != nil {
				_ = cmd.Usage()
				return usageErr(err)
			}
			if flagOrder != "" && flagOrder != "created" && flagOrder != "updated" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--order must be created or updated (server supports updated only; created sorts client-side)"))
			}
			if cliutil.IsDogfoodEnv() && flagMaxScanPages > 2 {
				flagMaxScanPages = 2
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()

			view, err := runList(ctx, flags, listOptions{
				sinceCut: sinceCut, hasSince: hasSince,
				beforeCut: beforeCut, hasBefore: hasBefore,
				withinCut: withinCut, hasWithin: hasWithin,
				order:           flagOrder,
				hydrate:         flagHydrate,
				includeArchived: flagIncludeArchived,
				limit:           flagLimit,
				maxScanPages:    flagMaxScanPages,
			})
			if err != nil {
				return err
			}
			if optHydrate(flagHydrate) && len(view.Items) > 0 {
				if cliutil.IsAnyHarness() && flagMaxHydrate > 2 {
					flagMaxHydrate = 2
				}
				gc, err := newGptClient(flags)
				if err != nil {
					return err
				}
				hctx, hcancel := context.WithTimeout(ctx, time.Duration(flagHydrateBudget)*time.Second)
				hydrated, hErr := hydrateBounded(hctx, flags, gc, view.Items, flagMaxHydrate, cmd.ErrOrStderr())
				hcancel()
				view.Hydrated = hydrated
				if hydrated < len(view.Items) {
					view.Note = fmt.Sprintf("hydrated %d of %d listed within the %ds budget (re-run to continue; raise --hydrate-budget or --max-hydrate for more)", hydrated, len(view.Items), flagHydrateBudget)
				}
				if hErr != nil && hydrated == 0 {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: hydration incomplete: %v\n", hErr)
				}
			}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			out := cmd.OutOrStdout()
			if len(view.Items) == 0 {
				if view.Note != "" {
					fmt.Fprintln(out, view.Note)
					return nil
				}
				fmt.Fprintln(out, "No conversations matched the given filters.")
				return nil
			}
			rows := make([]map[string]any, 0, len(view.Items))
			for _, it := range view.Items {
				id := it.ID
				if len(id) > 8 {
					id = id[:8] + "…"
				}
				rows = append(rows, map[string]any{
					"updated": it.Updated, "turns": it.UserTurns, "msgs": it.TotalMessages,
					"id": id, "title": truncateStr(it.Title, 40),
				})
			}
			return printAutoTable(out, rows)
		},
	}
	cmd.Flags().StringVar(&flagSince, "since", "", "only conversations created after this cutoff (30d, 12h, or ISO date)")
	cmd.Flags().StringVar(&flagBefore, "before", "", "only conversations created before this cutoff (30d ago, or ISO date)")
	cmd.Flags().StringVar(&flagUpdatedWithin, "updated-within", "", "only conversations updated within this window (48h, 7d)")
	cmd.Flags().StringVar(&flagOrder, "order", "updated", "created or updated (created sorts client-side)")
	cmd.Flags().BoolVar(&flagHydrate, "hydrate", false, "fetch details for turn counts and span (incremental; cached in the local store)")
	cmd.Flags().BoolVar(&flagIncludeArchived, "include-archived", false, "include archived conversations")
	cmd.Flags().IntVar(&flagLimit, "limit", 50, "maximum conversations to return")
	cmd.Flags().IntVar(&flagMaxScanPages, "max-scan-pages", 10, "maximum list pages to scan (page size 100)")
	cmd.Flags().IntVar(&flagMaxHydrate, "max-hydrate", 8, "maximum conversations to hydrate per run (incremental across runs)")
	cmd.Flags().IntVar(&flagHydrateBudget, "hydrate-budget", 6, "seconds to spend hydrating per run; remaining conversations stay cached-metadata-only until next run")
	return cmd
}

type listOptions struct {
	sinceCut, beforeCut, withinCut time.Time
	hasSince, hasBefore, hasWithin bool
	order                          string
	hydrate, includeArchived       bool
	limit, maxScanPages            int
}

// apiListItem mirrors GET /backend-api/conversations items.
type apiListItem struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
	IsArchived bool   `json:"is_archived"`
}

// fetcher is the subset of gptapi.Client list needs (keeps tests injectable).
type fetcher interface {
	Get(ctx context.Context, path string, params url.Values) (json.RawMessage, error)
}

func runList(ctx context.Context, flags *rootFlags, opt listOptions) (*listView, error) {
	var f fetcher
	var err error
	f, err = newGptClient(flags)
	if err != nil {
		return nil, err
	}
	return listWithFetcher(ctx, f, opt)
}

func listWithFetcher(ctx context.Context, f fetcher, opt listOptions) (*listView, error) {
	view := &listView{Items: make([]listRow, 0), MaxScanPages: opt.maxScanPages}
	scanned := 0
	var items []apiListItem
	offset := 0
	const pageSize = 100
	for page := 1; page <= opt.maxScanPages; page++ {
		q := url.Values{}
		q.Set("offset", strconv.Itoa(offset))
		q.Set("limit", strconv.Itoa(pageSize))
		q.Set("order", "updated")
		if !opt.includeArchived {
			q.Set("is_archived", "false")
		}
		raw, err := f.Get(ctx, "/backend-api/conversations", q)
		if err != nil {
			return nil, fmt.Errorf("listing conversations page %d: %w", page, err)
		}
		var pageResp struct {
			Items []apiListItem `json:"items"`
			Total int           `json:"total"`
		}
		if err := json.Unmarshal(raw, &pageResp); err != nil {
			return nil, fmt.Errorf("parsing conversations page %d: %w", page, err)
		}
		scanned += len(pageResp.Items)
		items = append(items, pageResp.Items...)
		// total is a hint: min(realTotal, offset+limit+1). Stop on short pages.
		if len(pageResp.Items) < pageSize {
			break
		}
		// Early termination for recency windows: server order is updated-desc,
		// so once a whole page predates the oldest cutoff, later pages cannot
		// match either filter.
		if (opt.hasSince || opt.hasWithin) && len(pageResp.Items) > 0 {
			oldest := time.Time{}
			for _, it := range pageResp.Items {
				if u := gptconv.ParseISOTimestamp(it.UpdateTime, nil); !u.IsZero() && (oldest.IsZero() || u.Before(oldest)) {
					oldest = u
				}
			}
			cutoff := opt.sinceCut
			if !opt.hasSince || (opt.hasWithin && opt.withinCut.After(cutoff)) {
				cutoff = opt.withinCut
			}
			if !oldest.IsZero() && oldest.Before(cutoff) {
				break
			}
		}
		offset += pageSize
	}
	view.ScannedItems = scanned

	for _, it := range items {
		created := gptconv.ParseISOTimestamp(it.CreateTime, nil)
		updated := gptconv.ParseISOTimestamp(it.UpdateTime, nil)
		if opt.hasSince && (created.IsZero() || created.Before(opt.sinceCut)) {
			continue
		}
		if opt.hasBefore && !created.IsZero() && created.After(opt.beforeCut) {
			continue
		}
		if opt.hasWithin && (updated.IsZero() || updated.Before(opt.withinCut)) {
			continue
		}
		view.Items = append(view.Items, listRow{
			ID:       it.ID,
			Title:    it.Title,
			Created:  isoOrEmpty(created),
			Updated:  isoOrEmpty(updated),
			Archived: it.IsArchived,
		})
	}
	if opt.order == "created" {
		sort.SliceStable(view.Items, func(i, j int) bool { return view.Items[i].Created < view.Items[j].Created })
	}
	if len(view.Items) > opt.limit {
		view.Items = view.Items[:opt.limit]
	}
	if len(view.Items) == 0 && scanned > 0 {
		view.Note = fmt.Sprintf("scanned %d conversations across up to %d pages without a match; raise --max-scan-pages to widen the search", scanned, opt.maxScanPages)
	}
	return view, nil
}

// hydrateBounded hydrates up to max conversations within ctx's deadline,
// returning how many were hydrated. Cached rows count instantly.
func hydrateBounded(ctx context.Context, flags *rootFlags, f fetcher, rows []listRow, max int, w io.Writer) (int, error) {
	dbPath := defaultDBPath("chatgpt-pp-cli")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return 0, err
	}
	db, err := store.OpenWithContext(ctx, dbPath)
	if err != nil {
		return 0, fmt.Errorf("opening store: %w", err)
	}
	defer db.Close()

	hydrated := 0
	for i := range rows {
		if hydrated >= max {
			break
		}
		if ctx.Err() != nil {
			return hydrated, ctx.Err()
		}
		row := &rows[i]
		if cached, err := db.Get("conversation_hydration", row.ID); err == nil && len(cached) > 0 {
			var st gptconv.Stats
			if json.Unmarshal(cached, &st) == nil && st.TotalMessages > 0 {
				applyStats(row, st)
				hydrated++
				continue
			}
		}
		raw, err := f.Get(ctx, "/backend-api/conversations/"+row.ID, nil)
		if err != nil {
			if ctx.Err() != nil {
				return hydrated, ctx.Err()
			}
			return hydrated, fmt.Errorf("hydrating %s: %w", shortID(row.ID), err)
		}
		var conv gptconv.Conversation
		if err := json.Unmarshal(raw, &conv); err != nil {
			return hydrated, fmt.Errorf("hydrating %s: parse: %w", shortID(row.ID), err)
		}
		st := gptconv.ComputeStats(&conv)
		_ = db.Upsert("conversation_hydration", row.ID, mustJSON(st))
		_ = db.Upsert("conversations_detail", row.ID, raw)
		applyStats(row, st)
		if row.Model == "" {
			row.Model = conv.DefaultModelSlug
		}
		hydrated++
	}
	return hydrated, nil
}

func applyStats(row *listRow, st gptconv.Stats) {
	row.UserTurns = st.UserTurnCount
	row.AssistantTurns = st.AssistantTurnCount
	row.TotalMessages = st.TotalMessages
	row.SpanSeconds = st.SpanSeconds
	row.ContentBytes = st.ContentBytes
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func isoOrEmpty(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func optHydrate(b bool) bool { return b }

func cliutilDogfoodImpl() bool { return cliutil.IsDogfoodEnv() }
