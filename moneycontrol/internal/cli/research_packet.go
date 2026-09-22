// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// pp:data-source live

type researchPacketView struct {
	SCID             string             `json:"sc_id"`
	TagSlug          string             `json:"tag_slug"`
	Quote            any                `json:"quote,omitempty"`
	PriceVolume      any                `json:"price_volume,omitempty"`
	News             []newsRow          `json:"news"`
	Filings          []htmlLink         `json:"filings"`
	Results          []htmlLink         `json:"results"`
	CorporateActions []htmlLink         `json:"corporate_actions"`
	Availability     []liveAvailability `json:"availability"`
	Source           string             `json:"source"`
	Note             string             `json:"note,omitempty"`
}

func newNovelResearchPacketCmd(flags *rootFlags) *cobra.Command {
	var flagScId string

	cmd := &cobra.Command{
		Use:         "packet",
		Short:       "Assemble live quote, price-volume, company news, filings, results, and corporate-action context for one company.",
		Example:     "  moneycontrol-pp-cli research packet --sc-id RI --agent --select quote,price_volume,news,filings,results,corporate_actions",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:typed-exit-codes": "0,2,3,5,7"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !cmd.Flags().Changed("sc-id") && !flags.dryRun {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "research packet")
			}
			if strings.TrimSpace(flagScId) == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--sc-id is required; pass a Moneycontrol SC ID such as RI"))
			}
			id := strings.ToUpper(strings.TrimSpace(flagScId))
			tag := tagSlugForSCID(id)
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			view := researchPacketView{SCID: id, TagSlug: tag, News: make([]newsRow, 0), Filings: make([]htmlLink, 0), Results: make([]htmlLink, 0), CorporateActions: make([]htmlLink, 0), Availability: make([]liveAvailability, 0), Source: "live"}
			var firstErr error

			quote, quoteErr := fetchJSONDocument(ctx, c, "https://priceapi.moneycontrol.com", "/pricefeed/nse/equitycash/"+id, nil)
			view.Availability = append(view.Availability, quote.Availability)
			if quoteErr != nil {
				firstErr = quoteErr
			} else {
				var decoded any
				if json.Unmarshal(quote.Data, &decoded) == nil {
					view.Quote = decoded
				}
			}

			priceVolume, priceErr := fetchJSONDocument(ctx, c, "https://api.moneycontrol.com", "/mcapi/v1/stock/price-volume", map[string]string{"scId": id})
			view.Availability = append(view.Availability, priceVolume.Availability)
			if priceErr != nil && firstErr == nil {
				firstErr = priceErr
			} else if priceErr == nil {
				var decoded any
				if json.Unmarshal(priceVolume.Data, &decoded) == nil {
					view.PriceVolume = decoded
				}
			}

			c.BaseURL = "https://www.moneycontrol.com"
			newsDoc, newsErr := fetchHTMLDocument(ctx, c, "/news/tags/"+tag+".html", 80)
			view.Availability = append(view.Availability, newsDoc.Availability)
			if newsErr != nil && firstErr == nil {
				firstErr = newsErr
			}
			if newsDoc.Availability.Status == "blocked" && firstErr == nil {
				firstErr = availabilityError(newsDoc.Availability)
			}
			view.News = parseNewsRows(newsDoc.Raw, "https://www.moneycontrol.com", id, tag, 20)

			pages := []struct {
				field string
				path  string
			}{
				{"filings", "/markets/filings/"},
				{"results", "/markets/earnings/"},
				{"corporate-actions", "/markets/corporate-action/"},
			}
			for _, page := range pages {
				c.BaseURL = "https://www.moneycontrol.com"
				doc, pageErr := fetchHTMLDocument(ctx, c, page.path, 100)
				view.Availability = append(view.Availability, doc.Availability)
				if pageErr != nil && firstErr == nil {
					firstErr = pageErr
				}
				if doc.Availability.Status == "blocked" && firstErr == nil {
					firstErr = availabilityError(doc.Availability)
				}
				matched := filterLinksForIDs(doc.Links, []string{id})[id]
				switch page.field {
				case "filings":
					view.Filings = matched
				case "results":
					view.Results = matched
				case "corporate-actions":
					view.CorporateActions = matched
				}
			}
			if firstErr != nil {
				view.Note = "Some live packet sources were unavailable; inspect availability.status/error before treating omitted fields as absent."
			}
			if view.Quote == nil && view.PriceVolume == nil && len(view.News) == 0 && len(view.Filings) == 0 && len(view.Results) == 0 && len(view.CorporateActions) == 0 && firstErr == nil {
				view.Note = "The live sources returned no parseable research rows."
			}
			sortAvailability(view.Availability)
			if err := emitComposed(cmd.OutOrStdout(), flags, view); err != nil {
				return err
			}
			if firstErr != nil && allUnavailable(view.Availability) {
				return classifyAPIErrorOnly(fmt.Errorf("research packet returned no usable live data: %w", firstErr))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagScId, "sc-id", "", "Moneycontrol SC ID to research, for example RI")
	return cmd
}
