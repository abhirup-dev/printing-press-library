// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// normalizeKiteInstrumentResponse converts Kite's public CSV instrument master
// into the CLI's normal JSON row shape. Dry-run responses remain untouched so
// they continue to show the planned request without pretending to parse data.
func normalizeKiteInstrumentResponse(data json.RawMessage, dryRun bool, prov DataProvenance) (json.RawMessage, DataProvenance, error) {
	prov.ResourceType = "instruments"
	if isDryRunResponse(dryRun, data) {
		return data, prov, nil
	}
	rows, err := parseKiteInstrumentCSV(data)
	if err != nil {
		return nil, DataProvenance{}, err
	}
	observedAt := time.Now().UTC().Format(time.RFC3339)
	prov.Freshness = map[string]any{"observed_at": observedAt}
	return rows, prov, nil
}

func parseKiteInstrumentCSV(data []byte) (json.RawMessage, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, errors.New("Kite instruments CSV schema error: empty response")
	}
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("Kite instruments CSV schema error: read header: %w", err)
	}
	if len(header) == 0 {
		return nil, errors.New("Kite instruments CSV schema error: missing header")
	}
	names := make([]string, len(header))
	seen := make(map[string]struct{}, len(header))
	for i, value := range header {
		name := strings.TrimSpace(value)
		if name == "" {
			return nil, fmt.Errorf("Kite instruments CSV schema error: blank header at column %d", i+1)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("Kite instruments CSV schema error: duplicate header %q", name)
		}
		seen[name] = struct{}{}
		names[i] = name
	}

	rows := make([]map[string]string, 0, 128)
	line := 1
	for {
		record, readErr := reader.Read()
		if errors.Is(readErr, csv.ErrFieldCount) {
			return nil, fmt.Errorf("Kite instruments CSV schema error: row %d has %d fields; expected %d", line+1, len(record), len(names))
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("Kite instruments CSV schema error: row %d: %w", line+1, readErr)
		}
		line++
		if len(record) != len(names) {
			return nil, fmt.Errorf("Kite instruments CSV schema error: row %d has %d fields; expected %d", line, len(record), len(names))
		}
		row := make(map[string]string, len(names))
		for i, value := range record {
			row[names[i]] = value
		}
		rows = append(rows, row)
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		return nil, fmt.Errorf("encode Kite instruments rows: %w", err)
	}
	return encoded, nil
}
