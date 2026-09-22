// pp:data-source live
package cli

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"tijori-finance-pp-cli/internal/client"
)

func newNovelResearchResultPackCmd(flags *rootFlags) *cobra.Command {
	var flagPeriod string
	cmd := &cobra.Command{
		Use:         "result-pack <slug>",
		Short:       "Connect a result or event to report, concall, and source-document metadata.",
		Example:     "  tijori-finance-pp-cli research result-pack tata-steel-limited --period Jun-26 --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "slug=tata-steel-limited;--period=Jun-26"},
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := novelRequiredArg(cmd, flags, args, "<slug>")
			if err != nil {
				return err
			}
			if strings.TrimSpace(flagPeriod) == "" {
				flagPeriod = "latest"
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "research result-pack")
			}
			type sourceSpec struct {
				name     string
				path     string
				resource string
			}
			specs := []sourceSpec{
				{name: "reports", path: "/financials/", resource: "reports"},
				{name: "concalls", path: "/results/concall-monitor/", resource: "concalls"},
				{name: "company_documents", path: "/company/" + slug + "/", resource: "company-documents"},
			}

			// Fetch the snapshot and independent source pages concurrently. Tijori can
			// stall one route until the request timeout; sequential probes multiplied
			// that delay by four and made the compound command effectively unusable.
			// Resolve Keychain-backed config exactly once before starting
			// goroutines, then give each request an independent HTTP client.
			clients, err := flags.newClients(len(specs) + 1)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			companyClient := clients[0]
			sourceClients := clients[1:]

			// Apply one command-wide deadline. The generated client may retry a
			// failed request, so a per-attempt HTTP timeout alone does not bound the
			// compound workflow.
			ctx, cancel := context.WithTimeout(cmd.Context(), flags.timeout)
			defer cancel()

			var (
				company    json.RawMessage
				prov       DataProvenance
				companyErr error
				wg         sync.WaitGroup
			)
			sourceResults := make([]map[string]any, len(specs))
			wg.Add(len(specs) + 1)
			go func() {
				defer wg.Done()
				company, prov, companyErr = novelCompanyPageWithClient(ctx, companyClient, slug, "", "company")
			}()
			for i := range specs {
				i := i
				go func() {
					defer wg.Done()
					spec := specs[i]
					sourceResults[i] = resultPackSource(ctx, sourceClients[i], spec.path, spec.resource)
				}()
			}
			wg.Wait()

			companySnapshot := any(company)
			if companyErr != nil {
				companySnapshot = map[string]any{
					"endpoint":      "/company/" + slug + "/",
					"method":        "GET",
					"source":        "live",
					"resource_type": "company",
					"observed_at":   time.Now().UTC().Format(time.RFC3339),
					"status":        "error",
					"availability":  resultPackAvailability(companyErr),
					"stability":     "unavailable",
					"error":         sanitizeError(companyErr),
				}
				prov = DataProvenance{Source: "live", ResourceType: "company"}
			}

			sources := make(map[string]any, len(specs))
			for i, spec := range specs {
				sources[spec.name] = sourceResults[i]
			}
			out := map[string]any{
				"company":          slug,
				"period":           flagPeriod,
				"company_snapshot": companySnapshot,
				"sources":          sources,
				"document_fetch":   "metadata-only; fetch a document only through an explicit safe GET command",
			}
			return printNovelResult(cmd, flags, jsonObject(out), prov)
		},
	}
	cmd.Flags().StringVar(&flagPeriod, "period", "", "Result or event period label")
	return cmd
}

func resultPackSource(ctx context.Context, c *client.Client, path, resource string) map[string]any {
	observedAt := time.Now().UTC().Format(time.RFC3339)
	result := map[string]any{
		"endpoint":    path,
		"method":      "GET",
		"source":      "live",
		"observed_at": observedAt,
	}
	data, _, err := novelHTMLLinksWithClient(ctx, c, path, resource)
	if err != nil {
		result["status"] = "error"
		result["availability"] = resultPackAvailability(err)
		result["stability"] = "unavailable"
		result["error"] = sanitizeError(err)
		return result
	}
	result["status"] = "ok"
	result["availability"] = "available"
	result["stability"] = "observed"
	result["data"] = json.RawMessage(data)
	return result
}

func resultPackAvailability(err error) string {
	message := strings.ToLower(sanitizeError(err))
	switch {
	case strings.Contains(message, "not authenticated"), strings.Contains(message, "unauthorized"), strings.Contains(message, "forbidden"):
		return "blocked"
	case strings.Contains(message, "not found"), strings.Contains(message, "empty response"):
		return "empty"
	default:
		return "upstream_error"
	}
}
