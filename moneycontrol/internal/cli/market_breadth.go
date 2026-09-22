// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// pp:data-source live

type indexBreadthView struct {
	Name         string           `json:"name"`
	Key          string           `json:"key"`
	Quote        any              `json:"quote,omitempty"`
	Availability liveAvailability `json:"availability"`
}

type marketBreadthView struct {
	Indices      []indexBreadthView  `json:"indices"`
	TableRows    []map[string]string `json:"table_rows"`
	Availability []liveAvailability  `json:"availability"`
	Source       string              `json:"source"`
	Note         string              `json:"note,omitempty"`
}

func newNovelMarketBreadthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "breadth",
		Short:       "Combine live index snapshots with available breadth/change-table rows and report partial availability.",
		Example:     "  moneycontrol-pp-cli market breadth --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:typed-exit-codes": "0,3,5,7"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 && !flags.dryRun {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "market breadth")
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			view := marketBreadthView{Indices: make([]indexBreadthView, 0, 4), TableRows: make([]map[string]string, 0), Availability: make([]liveAvailability, 0), Source: "live"}
			var firstErr error
			indices := []struct{ name, key string }{
				{"SENSEX", "in;SEN"},
				{"NIFTY 50", "in;NSX"},
				{"NIFTY BANK", "in;nbx"},
				{"NIFTY IT", "in;ccx"},
			}
			for _, index := range indices {
				doc, fetchErr := fetchJSONDocument(ctx, c, "https://priceapi.moneycontrol.com", "/pricefeed/notapplicable/inidicesindia/"+encodedIndexKey(index.key), nil)
				view.Availability = append(view.Availability, doc.Availability)
				row := indexBreadthView{Name: index.name, Key: index.key, Availability: doc.Availability}
				if fetchErr != nil {
					if firstErr == nil {
						firstErr = fetchErr
					}
				} else {
					var decoded any
					if err := json.Unmarshal(doc.Data, &decoded); err == nil {
						row.Quote = decoded
					}
				}
				view.Indices = append(view.Indices, row)
			}
			c.BaseURL = "https://www.moneycontrol.com"
			tableDoc, tableErr := fetchHTMLDocument(ctx, c, "/markets/indian-indices/", 100)
			view.Availability = append(view.Availability, tableDoc.Availability)
			if tableErr != nil && firstErr == nil {
				firstErr = tableErr
			}
			if tableDoc.Availability.Status == "blocked" && firstErr == nil {
				firstErr = availabilityError(tableDoc.Availability)
			}
			if len(tableDoc.Raw) > 0 {
				view.TableRows = parseHTMLTableRows(tableDoc.Raw, 50)
			}
			if len(view.TableRows) == 0 && tableDoc.Availability.Status == "ok" {
				view.Note = "The index page was reachable but exposed no parseable breadth table in this response."
			}
			if firstErr != nil {
				view.Note = "Some index or breadth sources were unavailable; inspect availability before interpreting partial rows."
			}
			sortAvailability(view.Availability)
			if err := emitComposed(cmd.OutOrStdout(), flags, view); err != nil {
				return err
			}
			if firstErr != nil && allUnavailable(view.Availability) {
				return classifyAPIErrorOnly(fmt.Errorf("market breadth returned no usable live data: %w", firstErr))
			}
			return nil
		},
	}
	return cmd
}
