// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseKiteInstrumentCSV(t *testing.T) {
	input := "\ufeffexchange,tradingsymbol,name,last_price\nNSE,ABC,Acme,123.45\nNSE,XYZ,\"Quoted, Name\",67.8\n"
	data, err := parseKiteInstrumentCSV([]byte(input))
	if err != nil {
		t.Fatalf("parseKiteInstrumentCSV: %v", err)
	}
	var rows []map[string]string
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatalf("decoded rows: %v", err)
	}
	if len(rows) != 2 || rows[1]["name"] != "Quoted, Name" || rows[0]["last_price"] != "123.45" {
		t.Fatalf("unexpected parsed rows: %#v", rows)
	}
}

func TestParseKiteInstrumentCSVRejectsSchemaErrors(t *testing.T) {
	cases := []string{
		"exchange,tradingsymbol\nNSE\n",
		"exchange,exchange\nNSE,NSE\n",
		"exchange,tradingsymbol\nNSE,ABC\nNSE,ABC,extra\n",
		"\n",
	}
	for _, input := range cases {
		_, err := parseKiteInstrumentCSV([]byte(input))
		if err == nil || !strings.Contains(err.Error(), "Kite instruments CSV schema error") {
			t.Errorf("parseKiteInstrumentCSV(%q) error = %v", input, err)
		}
	}
}
