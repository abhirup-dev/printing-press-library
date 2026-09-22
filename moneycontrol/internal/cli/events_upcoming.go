// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// pp:data-source live

type upcomingEventView struct {
	SCID     string `json:"sc_id"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Date     string `json:"date,omitempty"`
	URL      string `json:"url"`
}

type eventsUpcomingView struct {
	Events       []upcomingEventView `json:"events"`
	Availability []liveAvailability  `json:"availability"`
	Source       string              `json:"source"`
	Note         string              `json:"note,omitempty"`
}

func newNovelEventsUpcomingCmd(flags *rootFlags) *cobra.Command {
	var flagScIds string

	cmd := &cobra.Command{
		Use:         "upcoming",
		Short:       "Show live filings, results, and corporate actions grouped for supplied companies.",
		Example:     "  moneycontrol-pp-cli events upcoming --sc-ids RI,INFY --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:typed-exit-codes": "0,2,3,5,7"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !cmd.Flags().Changed("sc-ids") && !flags.dryRun {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "events upcoming")
			}
			ids, err := parseSCIDs(flagScIds)
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
			view := eventsUpcomingView{Events: make([]upcomingEventView, 0), Availability: make([]liveAvailability, 0), Source: "live"}
			var firstErr error
			pages := []struct {
				category string
				path     string
			}{
				{"earnings", "/markets/earnings/"},
				{"filing", "/markets/filings/"},
				{"corporate-action", "/markets/corporate-action/"},
			}
			seen := map[string]bool{}
			for _, page := range pages {
				c.BaseURL = "https://www.moneycontrol.com"
				doc, fetchErr := fetchHTMLDocument(ctx, c, page.path, 100)
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
						if seen[link.URL] {
							continue
						}
						seen[link.URL] = true
						view.Events = append(view.Events, upcomingEventView{SCID: id, Category: page.category, Title: firstNonEmpty(link.Name, link.Text), Date: headlineDate(link.Name + " " + link.Text), URL: link.URL})
					}
				}
			}
			if len(view.Events) == 0 && firstErr == nil {
				view.Note = "No matching event links were present in the current live pages; this is not historical completeness."
			}
			if firstErr != nil {
				view.Note = "One or more live event pages were unavailable; inspect availability.status/error before treating an empty category as no event."
			}
			sortAvailability(view.Availability)
			if err := emitComposed(cmd.OutOrStdout(), flags, view); err != nil {
				return err
			}
			if firstErr != nil && allUnavailable(view.Availability) {
				return classifyAPIErrorOnly(fmt.Errorf("events upcoming returned no usable live data: %w", firstErr))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagScIds, "sc-ids", "", "Comma-separated Moneycontrol SC IDs to group, for example RI,INFY")
	return cmd
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
