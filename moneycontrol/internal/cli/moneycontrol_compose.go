// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	xhtml "golang.org/x/net/html"
	"moneycontrol-pp-cli/internal/client"
)

// liveAvailability is deliberately attached to every composed response. A
// blocked or empty Moneycontrol widget is not equivalent to a successful empty
// result, especially on Akamai-sensitive routes.
type liveAvailability struct {
	Source string `json:"source"`
	URL    string `json:"url"`
	Status string `json:"status"` // ok, empty, blocked, error
	Count  int    `json:"count,omitempty"`
	Error  string `json:"error,omitempty"`
}

type liveDocument struct {
	Data         json.RawMessage  `json:"-"`
	Links        []htmlLink       `json:"-"`
	Availability liveAvailability `json:"-"`
	Raw          []byte           `json:"-"`
}

var moneycontrolDateRE = regexp.MustCompile(`(?i)\b(\d{1,2}\s+(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+\d{4}|(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+\d{1,2},?\s+\d{4}|\d{4}-\d{2}-\d{2})\b`)

var scIDTags = map[string]string{
	"RI":         "reliance-industries",
	"INFY":       "infosys",
	"INF":        "infosys",
	"HDF01":      "hdfc-bank",
	"HDFCBANK":   "hdfc-bank",
	"ITC":        "itc",
	"TCS":        "tata-consultancy-services",
	"SBIN":       "state-bank-of-india",
	"TATAMOTORS": "tata-motors",
	"MARUTI":     "maruti-suzuki-india",
	"ICICIBANK":  "icici-bank",
}

func parseSCIDs(raw string) ([]string, error) {
	seen := map[string]bool{}
	var ids []string
	for _, part := range strings.Split(raw, ",") {
		id := strings.ToUpper(strings.TrimSpace(part))
		if id == "" {
			continue
		}
		if !regexp.MustCompile(`^[A-Z0-9._-]+$`).MatchString(id) {
			return nil, fmt.Errorf("--sc-ids contains invalid stock ID %q; use comma-separated Moneycontrol SC IDs such as RI,INFY", id)
		}
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("--sc-ids is required; pass comma-separated Moneycontrol SC IDs such as RI,INFY")
	}
	return ids, nil
}

func parsePositiveInt(raw, flag string, defaultValue int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer; got %q", flag, raw)
	}
	return n, nil
}

func tagSlugForSCID(id string) string {
	if slug := scIDTags[strings.ToUpper(strings.TrimSpace(id))]; slug != "" {
		return slug
	}
	return strings.ToLower(strings.Trim(strings.ReplaceAll(strings.TrimSpace(id), "_", "-"), "-"))
}

func encodedIndexKey(key string) string {
	// The priceapi route treats the semicolon as part of the index key. Keep it
	// percent-encoded exactly as observed in live browser traffic.
	return strings.ReplaceAll(url.PathEscape(key), ";", "%3B")
}

func withClientBaseURL(c *client.Client, base string, fn func() (json.RawMessage, error)) (json.RawMessage, error) {
	previous := c.BaseURL
	c.BaseURL = strings.TrimRight(base, "/")
	defer func() { c.BaseURL = previous }()
	return fn()
}

func fetchJSONDocument(ctx context.Context, c *client.Client, base, path string, params map[string]string) (liveDocument, error) {
	value, err := withClientBaseURL(c, base, func() (json.RawMessage, error) {
		return c.Get(ctx, path, params)
	})
	av := liveAvailability{Source: base, URL: base + path, Status: "error"}
	if err != nil {
		av.Error = err.Error()
		return liveDocument{Availability: av}, err
	}
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		av.Status = "empty"
		return liveDocument{Data: value, Availability: av}, nil
	}
	var decoded any
	if err := json.Unmarshal(trimmed, &decoded); err != nil {
		av.Status = "error"
		av.Error = fmt.Sprintf("invalid JSON from %s: %v", av.URL, err)
		return liveDocument{Data: value, Availability: av}, fmt.Errorf("%s", av.Error)
	}
	if jsonValueEmpty(decoded) {
		av.Status = "empty"
	} else {
		av.Status = "ok"
		av.Count = jsonValueCount(decoded)
	}
	return liveDocument{Data: value, Availability: av}, nil
}

func fetchHTMLDocument(ctx context.Context, c *client.Client, path string, limit int) (liveDocument, error) {
	base := strings.TrimRight(c.BaseURL, "/")
	value, err := c.GetWithHeaders(ctx, path, nil, map[string]string{client.HTMLResponseHeader: "true"})
	av := liveAvailability{Source: base, URL: base + path, Status: "error"}
	if err != nil {
		av.Error = err.Error()
		return liveDocument{Availability: av}, err
	}
	raw := bytes.TrimSpace(value)
	if len(raw) == 0 {
		av.Status = "empty"
		return liveDocument{Raw: raw, Availability: av}, nil
	}
	if blockedMoneycontrolHTML(raw) {
		av.Status = "blocked"
		av.Error = "Moneycontrol returned a browser/WAF block page; retry later or use a less fragile route"
		return liveDocument{Raw: raw, Availability: av}, nil
	}
	linksRaw, parseErr := extractHTMLResponse(raw, htmlExtractionOptions{
		Context: ctx, Mode: "links", BaseURL: base + "/", LinkPrefixes: []string{}, Limit: limit,
	})
	if parseErr != nil {
		av.Status = "error"
		av.Error = parseErr.Error()
		return liveDocument{Raw: raw, Availability: av}, parseErr
	}
	var links []htmlLink
	if err := json.Unmarshal(linksRaw, &links); err != nil {
		av.Status = "error"
		av.Error = fmt.Sprintf("parsing links from %s: %v", av.URL, err)
		return liveDocument{Raw: raw, Availability: av}, fmt.Errorf("%s", av.Error)
	}
	if len(links) == 0 {
		av.Status = "empty"
	} else {
		av.Status = "ok"
		av.Count = len(links)
	}
	return liveDocument{Links: links, Availability: av, Raw: raw}, nil
}

func blockedMoneycontrolHTML(raw []byte) bool {
	text := strings.ToLower(string(raw))
	markers := []string{
		"access denied", "request unsuccessful", "reference #", "akamai ghost", "you don't have permission to access",
		"temporarily unavailable", "checking your browser", "just a moment", "verify you are human",
	}
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func jsonValueEmpty(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case []any:
		return len(x) == 0
	case map[string]any:
		if len(x) == 0 {
			return true
		}
		for _, key := range []string{"data", "result", "results", "items", "rows"} {
			if nested, ok := x[key]; ok {
				return jsonValueEmpty(nested)
			}
		}
	}
	return false
}

func jsonValueCount(v any) int {
	switch x := v.(type) {
	case []any:
		return len(x)
	case map[string]any:
		for _, key := range []string{"data", "result", "results", "items", "rows"} {
			if nested, ok := x[key]; ok {
				if n := jsonValueCount(nested); n > 0 {
					return n
				}
			}
		}
		return 1
	default:
		return 1
	}
}

func availabilityError(av liveAvailability) error {
	if av.Status == "blocked" || av.Status == "error" {
		return fmt.Errorf("%s unavailable (%s): %s", av.URL, av.Status, av.Error)
	}
	return nil
}

func makeAvailability(items ...liveAvailability) []liveAvailability {
	out := append([]liveAvailability(nil), items...)
	return out
}

func sortAvailability(items []liveAvailability) {
	sort.SliceStable(items, func(i, j int) bool { return items[i].URL < items[j].URL })
}

func emitComposed(cmdOut interface{ Write([]byte) (int, error) }, flags *rootFlags, value any) error {
	return printJSONFiltered(cmdOut, value, flags)
}

func headlineDate(raw string) string {
	match := moneycontrolDateRE.FindStringSubmatch(raw)
	if len(match) == 0 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

type newsRow struct {
	SCID      string `json:"sc_id"`
	Tag       string `json:"tag_slug"`
	Title     string `json:"title"`
	Timestamp string `json:"timestamp,omitempty"`
	URL       string `json:"url"`
	Event     string `json:"event,omitempty"`
}

func parseNewsRows(raw []byte, base, scID, tag string, limit int) []newsRow {
	if limit <= 0 {
		limit = 20
	}
	doc, err := xhtml.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil
	}
	rows := make([]newsRow, 0, limit)
	seen := map[string]bool{}
	walkHTML(doc, func(n *xhtml.Node) {
		if len(rows) >= limit || n.Type != xhtml.ElementNode || !strings.EqualFold(n.Data, "a") {
			return
		}
		href := strings.TrimSpace(attrValue(n, "href"))
		if href == "" || !strings.Contains(href, "/news/") {
			return
		}
		full := normalizeHTMLURL(href, base+"/")
		if full == "" || seen[full] {
			return
		}
		title := cleanHTMLText(nodeTextSuppressing(n))
		if len([]rune(title)) < 8 {
			return
		}
		parentText := title
		if n.Parent != nil {
			parentText = cleanHTMLText(nodeTextSuppressing(n.Parent))
		}
		row := newsRow{SCID: scID, Tag: tag, Title: title, Timestamp: headlineDate(parentText), URL: full, Event: classifyEvent(title)}
		rows = append(rows, row)
		seen[full] = true
	})
	return rows
}

func classifyEvent(text string) string {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "result") || strings.Contains(lower, "earnings"):
		return "earnings"
	case strings.Contains(lower, "ipo") || strings.Contains(lower, "listing"):
		return "ipo"
	case strings.Contains(lower, "dividend") || strings.Contains(lower, "split") || strings.Contains(lower, "buyback") || strings.Contains(lower, "corporate action"):
		return "corporate-action"
	case strings.Contains(lower, "filing") || strings.Contains(lower, "exchange"):
		return "filing"
	default:
		return "news"
	}
}

func filterLinksForIDs(links []htmlLink, ids []string) map[string][]htmlLink {
	out := make(map[string][]htmlLink, len(ids))
	for _, id := range ids {
		tokens := strings.FieldsFunc(strings.ToLower(id+" "+tagSlugForSCID(id)), func(r rune) bool { return r == '-' || r == '_' || r == ' ' })
		for _, link := range links {
			text := strings.ToLower(link.Name + " " + link.Text + " " + link.URL)
			matched := false
			for _, token := range tokens {
				if len(token) >= 3 && strings.Contains(text, token) {
					matched = true
					break
				}
			}
			if matched {
				out[id] = append(out[id], link)
			}
		}
	}
	return out
}

func parseHTMLTableRows(raw []byte, limit int) []map[string]string {
	if limit <= 0 {
		limit = 50
	}
	doc, err := xhtml.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil
	}
	rows := make([]map[string]string, 0, limit)
	walkHTML(doc, func(n *xhtml.Node) {
		if len(rows) >= limit || n.Type != xhtml.ElementNode || !strings.EqualFold(n.Data, "tr") {
			return
		}
		cells := make([]string, 0, 8)
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type != xhtml.ElementNode || (child.Data != "td" && child.Data != "th") {
				continue
			}
			text := cleanHTMLText(nodeTextSuppressing(child))
			if text != "" {
				cells = append(cells, text)
			}
		}
		if len(cells) == 0 {
			return
		}
		row := map[string]string{"cells": strings.Join(cells, " | ")}
		for i, cell := range cells {
			row[fmt.Sprintf("col_%d", i+1)] = cell
		}
		rows = append(rows, row)
	})
	return rows
}

func dateWithin(raw string, since, until time.Time) bool {
	dateText := headlineDate(raw)
	if dateText == "" {
		return true // keep undated rows and expose the missing date to callers
	}
	for _, layout := range []string{"2 Jan 2006", "2 January 2006", "Jan 2, 2006", "January 2, 2006", "2006-01-02"} {
		if parsed, err := time.Parse(layout, dateText); err == nil {
			return !parsed.Before(since) && parsed.Before(until)
		}
	}
	return true
}
