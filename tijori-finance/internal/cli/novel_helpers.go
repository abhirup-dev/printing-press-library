package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"tijori-finance-pp-cli/internal/client"
)

// novelGET is the common read-only path for hand-authored workflows. It never
// selects local data and leaves the response body untouched so callers can
// preserve the provider's shape and provenance.
func novelGET(ctx context.Context, flags *rootFlags, path string, params map[string]string, headers map[string]string, resource string, guardJSON bool) (json.RawMessage, DataProvenance, error) {
	c, err := flags.newClient()
	if err != nil {
		return nil, DataProvenance{}, err
	}
	data, err := c.GetWithHeaders(ctx, path, params, headers)
	if err != nil {
		return nil, DataProvenance{}, err
	}
	if guardJSON {
		if err := assertLiveJSONBody(data); err != nil {
			return nil, DataProvenance{}, err
		}
	}
	return data, DataProvenance{Source: "live", ResourceType: resource}, nil
}

func printNovelResult(cmd *cobra.Command, flags *rootFlags, data json.RawMessage, prov DataProvenance) error {
	wrapped, err := wrapWithProvenance(data, prov)
	if err != nil {
		return err
	}
	return printOutputWithFlagsMeta(cmd.OutOrStdout(), wrapped, flags, map[string]any{"source": "live"})
}

func novelRequiredArg(cmd *cobra.Command, flags *rootFlags, args []string, usage string) (string, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return args[0], nil
	}
	if flags.asJSON {
		_ = printJSONFiltered(cmd.OutOrStdout(), map[string]any{"error": "missing required argument", "usage": cmd.CommandPath() + " " + usage}, flags)
	}
	return "", usageErr(fmt.Errorf("missing required argument\nUsage: %s %s", cmd.CommandPath(), usage))
}

func novelPath(slug, suffix string) string {
	return "/company/" + strings.Trim(slug, "/") + "/" + strings.TrimPrefix(suffix, "/")
}

type novelProbe struct {
	Path       string `json:"path"`
	Method     string `json:"method"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"http_status,omitempty"`
	Content    string `json:"content,omitempty"`
	Error      string `json:"error,omitempty"`
}

// probeNovelGET intentionally reports metadata only; it never returns a
// response body or credential-bearing header.
func probeNovelGET(ctx context.Context, flags *rootFlags, path string, params map[string]string, resource string) novelProbe {
	p := novelProbe{Path: path, Method: http.MethodGet}
	data, _, err := novelGET(ctx, flags, path, params, nil, resource, false)
	if err == nil {
		if isDryRunResponse(flags.dryRun, data) {
			p.Status = "dry-run"
			return p
		}
		p.Status = "ok"
		if json.Valid(data) {
			p.Content = "json"
		} else if strings.HasPrefix(strings.TrimSpace(string(data)), "<") {
			p.Content = "html"
		} else {
			p.Content = "text"
		}
		return p
	}
	p.Status = "error"
	p.Error = sanitizeError(err)
	var apiErr *client.APIError
	if As(err, &apiErr) {
		p.HTTPStatus = apiErr.StatusCode
		switch apiErr.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			p.Status = "blocked"
		case http.StatusNotFound:
			p.Status = "empty"
		}
		if apiErr.StatusCode == http.StatusUnauthorized {
			p.Status = "expired"
		}
	}
	return p
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	// Never echo headers, cookies, or auth values if a transport includes them.
	for _, key := range []string{"authorization:", "cookie:", "x-csrftoken:", "csrf-token:"} {
		if i := strings.Index(strings.ToLower(s), key); i >= 0 {
			s = s[:i] + key + " [redacted]"
		}
	}
	return s
}

func jsonObject(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func novelHTMLLinks(ctx context.Context, flags *rootFlags, path, resource string) (json.RawMessage, DataProvenance, error) {
	c, err := flags.newClient()
	if err != nil {
		return nil, DataProvenance{}, err
	}
	return novelHTMLLinksWithClient(ctx, c, path, resource)
}

func novelHTMLLinksWithClient(ctx context.Context, c *client.Client, path, resource string) (json.RawMessage, DataProvenance, error) {
	raw, err := c.GetWithHeaders(ctx, path, nil, map[string]string{"X-Printing-Press-HTML-Response": "true"})
	if err != nil {
		return nil, DataProvenance{}, err
	}
	prov := DataProvenance{Source: "live", ResourceType: resource}
	if isDryRunResponse(c.IsDryRun(), raw) || json.Valid(raw) {
		return raw, prov, nil
	}
	extracted, extractErr := extractHTMLResponse(raw, htmlExtractionOptions{
		Context:        ctx,
		Mode:           "links",
		BaseURL:        htmlExtractionRequestURL(c.BaseURL, path, nil),
		LinkPrefixes:   []string{"/company", "/results"},
		ScriptSelector: "script#__NEXT_DATA__",
	})
	if extractErr != nil {
		return nil, DataProvenance{}, extractErr
	}
	return extracted, prov, nil
}

func novelCompanyPage(ctx context.Context, flags *rootFlags, slug, suffix, resource string) (json.RawMessage, DataProvenance, error) {
	c, err := flags.newClient()
	if err != nil {
		return nil, DataProvenance{}, err
	}
	return novelCompanyPageWithClient(ctx, c, slug, suffix, resource)
}

func novelCompanyPageWithClient(ctx context.Context, c *client.Client, slug, suffix, resource string) (json.RawMessage, DataProvenance, error) {
	path := "/company/" + strings.Trim(slug, "/") + "/" + strings.TrimPrefix(suffix, "/")
	raw, err := c.GetWithHeaders(ctx, path, nil, map[string]string{"X-Printing-Press-HTML-Response": "true"})
	if err != nil {
		return nil, DataProvenance{}, err
	}
	prov := DataProvenance{Source: "live", ResourceType: resource}
	if isDryRunResponse(c.IsDryRun(), raw) || json.Valid(raw) {
		return raw, prov, nil
	}
	extracted, extractErr := extractHTMLResponse(raw, htmlExtractionOptions{Context: ctx, Mode: "page", BaseURL: htmlExtractionRequestURL(c.BaseURL, path, nil), ScriptSelector: "script#__NEXT_DATA__"})
	if extractErr != nil {
		return nil, DataProvenance{}, extractErr
	}
	return extracted, prov, nil
}
