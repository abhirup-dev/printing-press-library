// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// pp:data-source live

type portfolioTriageView struct {
	Items        []newsRow          `json:"items"`
	SCIDs        []string           `json:"sc_ids"`
	Availability []liveAvailability `json:"availability"`
	Source       string             `json:"source"`
	Note         string             `json:"note,omitempty"`
}

func newNovelPortfolioTriageCmd(flags *rootFlags) *cobra.Command {
	var flagScIds string

	cmd := &cobra.Command{
		Use:         "triage",
		Short:       "Emit live headlines, timestamps, URLs, and event labels for supplied external holdings.",
		Example:     "  moneycontrol-pp-cli portfolio triage --sc-ids RI,INFY --agent --select items.sc_id,items.title,items.timestamp,items.url",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:typed-exit-codes": "0,2,3,5,7"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !cmd.Flags().Changed("sc-ids") && !flags.dryRun {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "portfolio triage")
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
			view := portfolioTriageView{Items: make([]newsRow, 0), SCIDs: ids, Availability: make([]liveAvailability, 0), Source: "live"}
			var firstErr error
			seen := map[string]bool{}
			for _, id := range ids {
				tag := tagSlugForSCID(id)
				c.BaseURL = "https://www.moneycontrol.com"
				doc, fetchErr := fetchHTMLDocument(ctx, c, "/news/tags/"+tag+".html", 80)
				view.Availability = append(view.Availability, doc.Availability)
				if fetchErr != nil && firstErr == nil {
					firstErr = fetchErr
				}
				if doc.Availability.Status == "blocked" && firstErr == nil {
					firstErr = availabilityError(doc.Availability)
				}
				for _, row := range parseNewsRows(doc.Raw, "https://www.moneycontrol.com", id, tag, 20) {
					if !seen[row.URL] {
						view.Items = append(view.Items, row)
						seen[row.URL] = true
					}
				}
			}
			if len(view.Items) == 0 && firstErr == nil {
				view.Note = "No current tag-feed headlines matched the supplied holdings."
			}
			if firstErr != nil {
				view.Note = "One or more holding tag feeds were unavailable; inspect availability.status/error before treating an empty triage as no news."
			}
			sortAvailability(view.Availability)
			if err := emitComposed(cmd.OutOrStdout(), flags, view); err != nil {
				return err
			}
			if firstErr != nil && allUnavailable(view.Availability) {
				return classifyAPIErrorOnly(fmt.Errorf("portfolio triage returned no usable live data: %w", firstErr))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagScIds, "sc-ids", "", "Comma-separated external holding SC IDs, for example RI,INFY")
	return cmd
}
