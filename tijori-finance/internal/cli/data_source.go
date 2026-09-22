package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"tijori-finance-pp-cli/internal/client"
)

// Data access is deliberately live-only. The generated templates retain the
// strategy parameters for source compatibility, but local stores, sync, and
// write-through caching are not part of this product.
type liveAllRejectReason string

const (
	liveAllRejectNone    liveAllRejectReason = ""
	liveAllRejectHTML    liveAllRejectReason = "html"
	liveAllRejectNonJSON liveAllRejectReason = "non-json"
)

func liveAllUnsupportedError(reason liveAllRejectReason, _ bool) error {
	switch reason {
	case liveAllRejectHTML:
		return fmt.Errorf("--all is not supported for live HTML responses; omit --all to fetch the current page")
	case liveAllRejectNonJSON:
		return fmt.Errorf("--all is not supported for live binary/text responses; omit --all to fetch the current page")
	default:
		return nil
	}
}

func unsupportedDataSourceError(strategy, requested string) error {
	return fmt.Errorf("data source %q is disabled for this live-only CLI (strategy %q)", requested, strategy)
}

func validateDataSourceStrategy(flags *rootFlags, strategy string) error {
	if strategy == "local" || (flags != nil && (flags.dataSource == "local" || flags.dataSource == "auto")) {
		return unsupportedDataSourceError(strategy, flags.dataSource)
	}
	if strategy != "" && strategy != "auto" && strategy != "live" {
		return fmt.Errorf("unsupported data-source strategy %q", strategy)
	}
	return nil
}

func attachFreshness(prov DataProvenance, flags *rootFlags) DataProvenance {
	if flags != nil {
		prov.Freshness = flags.freshnessMeta
	}
	return prov
}

func liveRead(ctx context.Context, c *client.Client, flags *rootFlags, path string, params, headers map[string]string, responsePath string, guardJSON bool) (json.RawMessage, DataProvenance, error) {
	data, err := c.GetWithHeaders(ctx, path, params, headers)
	if err != nil {
		return nil, DataProvenance{}, err
	}
	if isDryRunResponse(c.IsDryRun(), data) {
		return data, attachFreshness(DataProvenance{Source: "dry-run"}, flags), nil
	}
	if guardJSON {
		if err := assertLiveJSONBody(data); err != nil {
			return nil, DataProvenance{}, err
		}
	}
	return applyResponsePath(data, responsePath), attachFreshness(DataProvenance{Source: "live"}, flags), nil
}

func resolveRead(ctx context.Context, c *client.Client, flags *rootFlags, resourceType string, isList bool, path string, params map[string]string, headers map[string]string, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolveReadWithStrategyResponsePathAndJSONGuard(ctx, c, flags, "live", resourceType, isList, path, params, headers, "", true, hintWriter)
}
func resolveReadWithResponsePath(ctx context.Context, c *client.Client, flags *rootFlags, resourceType string, isList bool, path string, params map[string]string, headers map[string]string, responsePath string, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolveReadWithStrategyResponsePathAndJSONGuard(ctx, c, flags, "live", resourceType, isList, path, params, headers, responsePath, true, hintWriter)
}
func resolveReadWithStrategy(ctx context.Context, c *client.Client, flags *rootFlags, strategy, resourceType string, isList bool, path string, params map[string]string, headers map[string]string, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolveReadWithStrategyResponsePathAndJSONGuard(ctx, c, flags, strategy, resourceType, isList, path, params, headers, "", true, hintWriter)
}
func resolveReadWithStrategyAndResponsePath(ctx context.Context, c *client.Client, flags *rootFlags, strategy, resourceType string, isList bool, path string, params map[string]string, headers map[string]string, responsePath string, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolveReadWithStrategyResponsePathAndJSONGuard(ctx, c, flags, strategy, resourceType, isList, path, params, headers, responsePath, true, hintWriter)
}
func resolveReadWithStrategyResponsePathAndJSONGuard(ctx context.Context, c *client.Client, flags *rootFlags, strategy, resourceType string, _ bool, path string, params map[string]string, headers map[string]string, responsePath string, guardLiveJSON bool, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	if err := validateDataSourceStrategy(flags, strategy); err != nil {
		return nil, DataProvenance{}, err
	}
	return liveRead(ctx, c, flags, path, params, headers, responsePath, guardLiveJSON)
}

func resolvePaginatedRead(ctx context.Context, c *client.Client, flags *rootFlags, resourceType, path string, params map[string]string, headers map[string]string, fetchAll bool, cursorParam, paginationType, limitParam string, defaultPageSize int, nextCursorPath, hasMoreField string, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolvePaginatedReadWithStrategyWithJSONGuard(ctx, c, flags, "live", resourceType, path, params, headers, fetchAll, cursorParam, paginationType, limitParam, defaultPageSize, nextCursorPath, hasMoreField, "", true, liveAllRejectNone, hintWriter)
}
func resolvePaginatedReadWithStrategy(ctx context.Context, c *client.Client, flags *rootFlags, strategy, resourceType, path string, params map[string]string, headers map[string]string, fetchAll bool, cursorParam, paginationType, limitParam string, defaultPageSize int, nextCursorPath, hasMoreField, responsePath string, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolvePaginatedReadWithStrategyWithJSONGuard(ctx, c, flags, strategy, resourceType, path, params, headers, fetchAll, cursorParam, paginationType, limitParam, defaultPageSize, nextCursorPath, hasMoreField, responsePath, true, liveAllRejectNone, hintWriter)
}
func resolvePaginatedReadWithStrategyAndJSONGuard(ctx context.Context, c *client.Client, flags *rootFlags, strategy, resourceType, path string, params map[string]string, headers map[string]string, fetchAll bool, cursorParam, paginationType, limitParam string, defaultPageSize int, nextCursorPath, hasMoreField, responsePath string, guardLiveJSON bool, liveAllReject liveAllRejectReason, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolvePaginatedReadWithStrategyWithJSONGuard(ctx, c, flags, strategy, resourceType, path, params, headers, fetchAll, cursorParam, paginationType, limitParam, defaultPageSize, nextCursorPath, hasMoreField, responsePath, guardLiveJSON, liveAllReject, hintWriter)
}
func resolvePaginatedReadWithStrategyWithJSONGuard(ctx context.Context, c *client.Client, flags *rootFlags, strategy, resourceType, path string, params map[string]string, headers map[string]string, fetchAll bool, cursorParam, paginationType, limitParam string, defaultPageSize int, nextCursorPath, hasMoreField, responsePath string, guardLiveJSON bool, liveAllReject liveAllRejectReason, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	if err := validateDataSourceStrategy(flags, strategy); err != nil {
		return nil, DataProvenance{}, err
	}
	if fetchAll {
		if err := liveAllUnsupportedError(liveAllReject, false); err != nil {
			return nil, DataProvenance{}, err
		}
	}
	data, err := paginatedGetWithResponsePath(ctx, c, path, params, headers, fetchAll, cursorParam, paginationType, limitParam, defaultPageSize, nextCursorPath, hasMoreField, responsePath)
	if err != nil {
		return nil, DataProvenance{}, err
	}
	if isDryRunResponse(c.IsDryRun(), data) {
		return data, attachFreshness(DataProvenance{Source: "dry-run"}, flags), nil
	}
	if guardLiveJSON {
		if err := assertLiveJSONBody(data); err != nil {
			return nil, DataProvenance{}, err
		}
	}
	return data, attachFreshness(DataProvenance{Source: "live", ResourceType: resourceType}, flags), nil
}
