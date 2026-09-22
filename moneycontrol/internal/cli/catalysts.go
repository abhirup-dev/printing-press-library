// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// pp:data-source live

type catalystView struct {
	SCID     string `json:"sc_id"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Date     string `json:"date,omitempty"`
	URL      string `json:"url"`
}

type catalystsResultView struct {
	Catalysts    []catalystView     `json:"catalysts"`
	Days         int                `json:"days"`
	Availability []liveAvailability `json:"availability"`
	Source       string             `json:"source"`
	Note         string             `json:"note,omitempty"`
}

func newNovelCatalystsCmd(flags *rootFlags) *cobra.Command {
	var flagScIds string
	var flagDays string

	cmd := &cobra.Command{
		Use:         "catalysts",
		Short:       "Filter live dated IPO, earnings, filing, and corporate-action pages by supplied symbols and time window.",
		Example:     "  moneycontrol-pp-cli catalysts --sc-ids RI,INFY --days 30 --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:typed-exit-codes": "0,2,3,5,7"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !cmd.Flags().Changed("sc-ids") && !cmd.Flags().Changed("days") && !flags.dryRun {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "catalysts")
			}
			ids, err := parseSCIDs(flagScIds)
			if err != nil {
				_ = cmd.Usage()
				return usageErr(err)
			}
			days, err := parsePositiveInt(flagDays, "--days", 30)
			if err != nil {
				_ = cmd.Usage()
				return usageErr(err)
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			view := catalystsResultView{Catalysts: make([]catalystView, 0), Days: days, Availability: make([]liveAvailability, 0), Source: "live"}
			var firstErr error
			pages := []struct{ category, path string }{
				{"earnings", "/markets/earnings/"},
				{"ipo", "/ipo/"},
				{"filing", "/markets/filings/"},
				{"corporate-action", "/markets/corporate-action/"},
			}
			seen := map[string]bool{}
			since, until := time.Now().UTC().Add(-24*time.Hour), time.Now().UTC().Add(time.Duration(days)*24*time.Hour)
			for _, page := range pages {
				c.BaseURL = "https://www.moneycontrol.com"
				doc, fetchErr := fetchHTMLDocument(ctx, c, page.path, 120)
				view.Availability = append(view.Availability, doc.Availability)
				if fetchErr != nil && firstErr == nil {
					firstErr = fetchErr
				}
				if doc.Availability.Status == "blocked" && firstErr == nil {
					firstErr = availabilityError(doc.Availability)
				}
				matched := filterLinksForIDs(doc.Links, ids)
				for _, id := range ids {
					for _, link := range matched[id] {
						if seen[link.URL] || !dateWithin(link.Name+" "+link.Text, since, until) {
							continue
						}
						seen[link.URL] = true
						view.Catalysts = append(view.Catalysts, catalystView{SCID: id, Category: page.category, Title: firstNonEmpty(link.Name, link.Text), Date: headlineDate(link.Name + " " + link.Text), URL: link.URL})
					}
				}
			}
			if len(view.Catalysts) == 0 && firstErr == nil {
				view.Note = "No matching dated catalyst links were present in the current live pages; undated or older items are not inferred."
			}
			if firstErr != nil {
				view.Note = "Some catalyst pages were unavailable; inspect availability.status/error before treating an empty category as no catalyst."
			}
			sortAvailability(view.Availability)
			if err := emitComposed(cmd.OutOrStdout(), flags, view); err != nil {
				return err
			}
			if firstErr != nil && allUnavailable(view.Availability) {
				return classifyAPIErrorOnly(fmt.Errorf("catalysts returned no usable live data: %w", firstErr))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagScIds, "sc-ids", "", "Comma-separated Moneycontrol SC IDs to match, for example RI,INFY")
	cmd.Flags().StringVar(&flagDays, "days", "30", "Forward catalyst window in days; live pages are filtered mechanically")
	return cmd
}
