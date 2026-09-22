// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.

package cli

import (
	"testing"
	"time"
)

func TestParseSCIDs(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
		fail bool
	}{
		{"deduplicates and uppercases", "ri, INFY,ri", []string{"RI", "INFY"}, false},
		{"rejects punctuation", "RI, bad/value", nil, true},
		{"requires value", " , ", nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSCIDs(tc.in)
			if tc.fail {
				if err == nil {
					t.Fatalf("parseSCIDs(%q) expected error", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSCIDs(%q): %v", tc.in, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got %#v, want %#v", got, tc.want)
				}
			}
		})
	}
}

func TestMoneycontrolHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"index key remains encoded", encodedIndexKey("in;SEN"), "in%3BSEN"},
		{"reliance tag", tagSlugForSCID("RI"), "reliance-industries"},
		{"fallback tag", tagSlugForSCID("ABC_123"), "abc-123"},
		{"earnings classification", classifyEvent("Q4 results announced"), "earnings"},
		{"filing classification", classifyEvent("Exchange filing update"), "filing"},
		{"corporate action classification", classifyEvent("Dividend record date"), "corporate-action"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestParseNewsRowsAndDateWindow(t *testing.T) {
	raw := []byte(`<html><body><article><a href="/news/business/stocks/reliance-results-14000001.html"><span>Reliance Q4 results</span><time>12 Jan 2026</time></a></article></body></html>`)
	rows := parseNewsRows(raw, "https://www.moneycontrol.com", "RI", "reliance-industries", 10)
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0].URL == "" || rows[0].SCID != "RI" || rows[0].Event != "earnings" {
		t.Fatalf("unexpected row: %#v", rows[0])
	}
	if !dateWithin("12 Jan 2026", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("date should be inside window")
	}
	if dateWithin("12 Jan 2025", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("date should be outside window")
	}
}
