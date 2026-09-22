// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// pp:data-source live

type holdingContextView struct {
	SCID         string             `json:"sc_id"`
	TagSlug      string             `json:"tag_slug"`
	Quote        any                `json:"quote,omitempty"`
	News         []newsRow          `json:"news"`
	Events       []htmlLink         `json:"events"`
	Availability []liveAvailability `json:"availability"`
}

type contextHoldingsView struct {
	Holdings     []holdingContextView `json:"holdings"`
	Availability []liveAvailability   `json:"availability"`
	Source       string               `json:"source"`
	Note         string               `json:"note,omitempty"`
}

func newNovelContextHoldingsCmd(flags *rootFlags) *cobra.Command {
	var flagScIds string

	cmd := &cobra.Command{
		Use:         "holdings",
		Short:       "Join supplied broker holding symbols to live Moneycontrol quotes, tagged news, and event context.",
		Example:     "  moneycontrol-pp-cli context holdings --sc-ids RI,INFY --agent --select holdings.sc_id,holdings.quote,holdings.news",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:typed-exit-codes": "0,2,3,5,7"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !cmd.Flags().Changed("sc-ids") && !flags.dryRun {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "context holdings")
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

			view := contextHoldingsView{Holdings: make([]holdingContextView, 0, len(ids)), Availability: make([]liveAvailability, 0), Source: "live"}
			var firstErr error
			for _, id := range ids {
				tag := tagSlugForSCID(id)
				holding := holdingContextView{SCID: id, TagSlug: tag, News: make([]newsRow, 0), Events: make([]htmlLink, 0), Availability: make([]liveAvailability, 0)}
				quote, quoteErr := fetchJSONDocument(ctx, c, "https://priceapi.moneycontrol.com", "/pricefeed/nse/equitycash/"+id, nil)
				holding.Availability = append(holding.Availability, quote.Availability)
				view.Availability = append(view.Availability, quote.Availability)
				if quoteErr != nil {
					if firstErr == nil {
						firstErr = quoteErr
					}
				} else if quote.Availability.Status != "empty" {
					var decoded any
					if err := jsonUnmarshalRaw(quote.Data, &decoded); err == nil {
						holding.Quote = decoded
					}
				}

				previousBase := c.BaseURL
				c.BaseURL = "https://www.moneycontrol.com"
				newsDoc, newsErr := fetchHTMLDocument(ctx, c, "/news/tags/"+tag+".html", 50)
				c.BaseURL = previousBase
				holding.Availability = append(holding.Availability, newsDoc.Availability)
				view.Availability = append(view.Availability, newsDoc.Availability)
				if newsErr != nil && firstErr == nil {
					firstErr = newsErr
				}
				if newsDoc.Availability.Status == "blocked" && firstErr == nil {
					firstErr = availabilityError(newsDoc.Availability)
				}
				if len(newsDoc.Raw) > 0 {
					holding.News = parseNewsRows(newsDoc.Raw, "https://www.moneycontrol.com", id, tag, 10)
				}
				view.Holdings = append(view.Holdings, holding)
			}

			// One broad event page supplies optional context for all requested IDs;
			// its availability remains explicit instead of silently becoming [] .
			previousBase := c.BaseURL
			c.BaseURL = "https://www.moneycontrol.com"
			eventDoc, eventErr := fetchHTMLDocument(ctx, c, "/markets/corporate-action/", 100)
			c.BaseURL = previousBase
			view.Availability = append(view.Availability, eventDoc.Availability)
			if eventErr != nil && firstErr == nil {
				firstErr = eventErr
			}
			if eventDoc.Availability.Status == "blocked" && firstErr == nil {
				firstErr = availabilityError(eventDoc.Availability)
			}
			matched := filterLinksForIDs(eventDoc.Links, ids)
			for i := range view.Holdings {
				view.Holdings[i].Events = matched[view.Holdings[i].SCID]
				if view.Holdings[i].Events == nil {
					view.Holdings[i].Events = make([]htmlLink, 0)
				}
			}
			sortAvailability(view.Availability)
			if firstErr != nil {
				view.Note = "One or more live Moneycontrol sources were unavailable; inspect availability.status/error before relying on empty fields."
			}
			if err := emitComposed(cmd.OutOrStdout(), flags, view); err != nil {
				return err
			}
			if firstErr != nil && allUnavailable(view.Availability) {
				return classifyAPIErrorOnly(fmt.Errorf("context holdings returned no usable live data: %w", firstErr))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagScIds, "sc-ids", "", "Comma-separated Moneycontrol SC IDs to review, for example RI,INFY")
	return cmd
}

func allUnavailable(items []liveAvailability) bool {
	if len(items) == 0 {
		return true
	}
	for _, item := range items {
		if item.Status == "ok" {
			return false
		}
	}
	return true
}

// jsonUnmarshalRaw keeps generated command files free of repetitive RawMessage
// plumbing while preserving the exact live JSON shape in the composed result.
func jsonUnmarshalRaw(raw []byte, dst any) error {
	return json.Unmarshal(raw, dst)
}
