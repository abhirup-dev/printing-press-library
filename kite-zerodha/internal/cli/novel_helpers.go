package cli

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"kite-zerodha-pp-cli/internal/client"
)

// fetchNovelLive is the only data path used by the hand-written analytics
// commands. It deliberately uses the live strategy: current broker state must
// not fall back to a local snapshot or be written through to one.
func bindNovelCommand(cmd *cobra.Command, flags *rootFlags) (*cobra.Command, func()) {
	ctx, cancel := boundCtx(cmd.Context(), flags)
	cmd.SetContext(ctx)
	return cmd, cancel
}

func fetchNovelLive(cmd *cobra.Command, flags *rootFlags, c *client.Client, resource, path string, list bool) (json.RawMessage, error) {
	data, _, err := resolveReadWithStrategyAndResponsePath(cmd.Context(), c, flags, "live", resource, list, path, nil, nil, "data", cmd.ErrOrStderr())
	if err != nil {
		return nil, classifyAPIError(cmd.OutOrStdout(), err, flags)
	}
	return data, nil
}

func printNovelLive(cmd *cobra.Command, flags *rootFlags, value any, documented ...map[string]bool) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return printOutputWithFlagsMeta(cmd.OutOrStdout(), raw, flags, map[string]any{"source": "live"}, documented...)
}

func rawRecords(data json.RawMessage) ([]map[string]any, error) {
	if len(data) == 0 || string(data) == "null" {
		return []map[string]any{}, nil
	}
	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err == nil && rows != nil {
		return rows, nil
	}
	var one map[string]any
	if err := json.Unmarshal(data, &one); err != nil {
		return nil, fmt.Errorf("decode live response: %w", err)
	}
	if one == nil {
		return []map[string]any{}, nil
	}
	return []map[string]any{one}, nil
}

func rawObject(data json.RawMessage) (map[string]any, error) {
	if len(data) == 0 || string(data) == "null" {
		return map[string]any{}, nil
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, fmt.Errorf("decode live response: %w", err)
	}
	if object == nil {
		return map[string]any{}, nil
	}
	return object, nil
}

func cloneRow(row map[string]any) map[string]any {
	out := make(map[string]any, len(row)+4)
	for k, v := range row {
		out[k] = v
	}
	return out
}

func stringValue(row map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := row[key]; ok && value != nil {
			if text, ok := value.(string); ok {
				return text
			}
			return fmt.Sprint(value)
		}
	}
	return ""
}

func numberValue(row map[string]any, keys ...string) float64 {
	for _, key := range keys {
		value, ok := row[key]
		if !ok || value == nil {
			continue
		}
		switch n := value.(type) {
		case float64:
			return n
		case float32:
			return float64(n)
		case int:
			return float64(n)
		case int64:
			return float64(n)
		case json.Number:
			if parsed, err := n.Float64(); err == nil {
				return parsed
			}
		case string:
			var parsed float64
			if _, err := fmt.Sscan(n, &parsed); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func hasValue(row map[string]any, key string) bool {
	value, ok := row[key]
	return ok && value != nil
}

func finiteOrZero(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

func pnlValue(row map[string]any) float64 {
	return finiteOrZero(numberValue(row, "pnl", "m2m", "unrealised", "unrealized"))
}

func rowQuantity(row map[string]any) float64 {
	return numberValue(row, "quantity", "net_quantity")
}

func rowMarketValue(row map[string]any) float64 {
	if hasValue(row, "value") {
		return numberValue(row, "value")
	}
	return rowQuantity(row) * numberValue(row, "last_price", "price", "close_price")
}

func rowCost(row map[string]any) float64 {
	return rowQuantity(row) * numberValue(row, "average_price", "price")
}

func turnover(row map[string]any) float64 {
	return math.Abs(numberValue(row, "quantity") * numberValue(row, "average_price", "price"))
}

func asOf() string { return time.Now().UTC().Format(time.RFC3339) }

func sortRowsByString(rows []map[string]any, key string) {
	sort.Slice(rows, func(i, j int) bool { return stringValue(rows[i], key) < stringValue(rows[j], key) })
}

func rawArrayField(object map[string]any, key string) []map[string]any {
	value, ok := object[key]
	if !ok {
		return []map[string]any{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return []map[string]any{}
	}
	rows, err := rawRecords(data)
	if err != nil {
		return []map[string]any{}
	}
	return rows
}

var documentedNovelFields = map[string]bool{
	"as_of": true, "source": true, "scope": true, "status": true,
	"summary": true, "holdings": true, "positions": true, "margins": true,
	"trades": true, "orders": true, "groups": true, "metrics": true,
	"limitations": true, "capabilities": true,
}

func normalizeGroup(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "symbol"
	}
	return value
}
