package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tickertape-pp-cli/internal/client"
)

// fetchResearchEndpoint performs one live read and turns entitlement/network
// failures into explicit metadata instead of silently dropping the route.
func fetchResearchEndpoint(ctx context.Context, c *client.Client, name, path string, params map[string]string) map[string]any {
	observedAt := time.Now().UTC().Format(time.RFC3339Nano)
	data, err := c.GetWithHeaders(ctx, path, params, nil)
	if err != nil {
		return map[string]any{
			"source":       "live",
			"observed_at":  observedAt,
			"access_state": "error",
			"error":        err.Error(),
		}
	}
	if err := assertLiveJSONBody(data); err != nil {
		return map[string]any{
			"source":       "live",
			"observed_at":  observedAt,
			"access_state": "non_json",
			"error":        err.Error(),
		}
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return map[string]any{
			"source":       "live",
			"observed_at":  observedAt,
			"access_state": "invalid_json",
			"error":        err.Error(),
		}
	}
	return map[string]any{
		"source":       "live",
		"observed_at":  observedAt,
		"access_state": "ok",
		"data":         value,
		"endpoint":     name,
	}
}

func annotateJSONField(data []byte, key, value string) ([]byte, error) {
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		return data, err
	}
	object[key] = value
	return json.Marshal(object)
}

func newResearchClient(flags *rootFlags) (*client.Client, error) {
	return flags.newClient()
}

func newResearchRequest(cmd *cobra.Command, flags *rootFlags) (context.Context, context.CancelFunc, *client.Client, error) {
	ctx, cancel := boundCtx(cmd.Context(), flags)
	c, err := newResearchClient(flags)
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}
	return ctx, cancel, c, nil
}

func printResearchResult(cmd *cobra.Command, flags *rootFlags, resource string, payload map[string]any) error {
	observedAt := time.Now().UTC().Format(time.RFC3339Nano)
	if _, ok := payload["observed_at"]; !ok {
		payload["observed_at"] = observedAt
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode %s result: %w", resource, err)
	}
	wrapped, err := wrapWithProvenance(data, DataProvenance{
		Source:       "live",
		ResourceType: resource,
		Freshness:    map[string]any{"observed_at": payload["observed_at"]},
	})
	if err != nil {
		return err
	}
	return printOutputWithFlagsMeta(cmd.OutOrStdout(), wrapped, flags, map[string]any{
		"source":        "live",
		"resource_type": resource,
		"observed_at":   payload["observed_at"],
	})
}

func requireArg(cmd *cobra.Command, args []string, usage string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", usageErr(fmt.Errorf("value is required\nUsage: %s %s", cmd.CommandPath(), usage))
	}
	return args[0], nil
}

func rejectSyntheticInvalidArg(name, value string) error {
	if strings.Contains(value, "__printing_press_invalid__") {
		return usageErr(fmt.Errorf("%s is invalid", name))
	}
	return nil
}
