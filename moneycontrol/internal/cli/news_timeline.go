// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// pp:data-source live

type newsTimelineView struct {
	SCID         string             `json:"sc_id"`
	TagSlug      string             `json:"tag_slug"`
	Articles     []newsRow          `json:"articles"`
	Availability []liveAvailability `json:"availability"`
	Source       string             `json:"source"`
	Note         string             `json:"note,omitempty"`
}

func newNovelNewsTimelineCmd(flags *rootFlags) *cobra.Command {
	var flagScId string
	var flagLimit string

	cmd := &cobra.Command{
		Use:         "timeline",
		Short:       "Normalize a live company-tag page into URL-preserving chronological story rows.",
		Example:     "  moneycontrol-pp-cli news timeline --sc-id RI --limit 20 --agent --select articles.title,articles.timestamp,articles.url",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:typed-exit-codes": "0,2,3,5,7"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !cmd.Flags().Changed("sc-id") && !cmd.Flags().Changed("limit") && !flags.dryRun {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "news timeline")
			}
			if strings.TrimSpace(flagScId) == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--sc-id is required; pass a Moneycontrol SC ID such as RI"))
			}
			limit, err := parsePositiveInt(flagLimit, "--limit", 20)
			if err != nil {
				_ = cmd.Usage()
				return usageErr(err)
			}
			id := strings.ToUpper(strings.TrimSpace(flagScId))
			tag := tagSlugForSCID(id)
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			c.BaseURL = "https://www.moneycontrol.com"
			doc, fetchErr := fetchHTMLDocument(ctx, c, "/news/tags/"+tag+".html", limit*3)
			view := newsTimelineView{SCID: id, TagSlug: tag, Articles: make([]newsRow, 0), Availability: []liveAvailability{doc.Availability}, Source: "live"}
			if len(doc.Raw) > 0 {
				view.Articles = parseNewsRows(doc.Raw, "https://www.moneycontrol.com", id, tag, limit)
			}
			// Parsed timestamps are sorted newest-first; undated rows retain their
			// live DOM order after dated rows so no date is fabricated.
			sort.SliceStable(view.Articles, func(i, j int) bool {
				a, b := parseNewsDate(view.Articles[i].Timestamp), parseNewsDate(view.Articles[j].Timestamp)
				if a.IsZero() || b.IsZero() {
					return !a.IsZero() && b.IsZero()
				}
				return a.After(b)
			})
			if len(view.Articles) > limit {
				view.Articles = view.Articles[:limit]
			}
			if fetchErr != nil {
				view.Note = "The company tag response could not be parsed; inspect availability.error."
			}
			if doc.Availability.Status == "blocked" {
				view.Note = "Moneycontrol blocked the company tag route; retry later or inspect a less fragile public route."
			}
			if len(view.Articles) == 0 && fetchErr == nil && doc.Availability.Status == "ok" {
				view.Note = "The live tag page contained no parseable article cards."
			}
			sortAvailability(view.Availability)
			if err := emitComposed(cmd.OutOrStdout(), flags, view); err != nil {
				return err
			}
			if fetchErr != nil || doc.Availability.Status == "blocked" {
				return classifyAPIErrorOnly(fmt.Errorf("news timeline unavailable: %s", firstNonEmpty(doc.Availability.Error, "live tag extraction failed")))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagScId, "sc-id", "", "Moneycontrol SC ID to timeline, for example RI")
	cmd.Flags().StringVar(&flagLimit, "limit", "20", "Maximum live article rows to return")
	return cmd
}

func parseNewsDate(raw string) time.Time {
	for _, layout := range []string{"2 Jan 2006", "2 January 2006", "Jan 2, 2006", "January 2, 2006", "2006-01-02"} {
		if parsed, err := time.Parse(layout, strings.TrimSpace(raw)); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
