// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0.
// Live-only data resolution. Domain responses are never cached or persisted.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"tickertape-pp-cli/internal/client"
)

type liveAllRejectReason string

const (
	liveAllRejectNone    liveAllRejectReason = ""
	liveAllRejectHTML    liveAllRejectReason = "html"
	liveAllRejectNonJSON liveAllRejectReason = "non-json"
)

func liveAllUnsupportedError(reason liveAllRejectReason, _ bool) error {
	switch reason {
	case liveAllRejectHTML:
		return fmt.Errorf("--all is not supported for live HTML responses; omit --all to extract the current page")
	case liveAllRejectNonJSON:
		return fmt.Errorf("--all is not supported for live binary/text responses; omit --all to fetch the current page")
	default:
		return nil
	}
}

// attachFreshness preserves command-level freshness metadata without adding a
// local data source. All successful domain responses have source=live.
func attachFreshness(prov DataProvenance, flags *rootFlags) DataProvenance {
	if flags != nil {
		prov.Freshness = flags.freshnessMeta
	}
	return prov
}

func resolveRead(ctx context.Context, c *client.Client, flags *rootFlags, resourceType string, _ bool, path string, params map[string]string, headers map[string]string, _ io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolveReadWithResponsePath(ctx, c, flags, resourceType, false, path, params, headers, "", nil)
}

func resolveReadWithResponsePath(ctx context.Context, c *client.Client, flags *rootFlags, resourceType string, _ bool, path string, params map[string]string, headers map[string]string, responsePath string, _ io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolveReadWithStrategyAndResponsePath(ctx, c, flags, "live", resourceType, false, path, params, headers, responsePath, nil)
}

func resolveReadWithStrategy(ctx context.Context, c *client.Client, flags *rootFlags, _ string, resourceType string, isList bool, path string, params map[string]string, headers map[string]string, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolveReadWithStrategyAndResponsePath(ctx, c, flags, "live", resourceType, isList, path, params, headers, "", hintWriter)
}

func resolveReadWithStrategyAndResponsePath(ctx context.Context, c *client.Client, flags *rootFlags, _ string, resourceType string, _ bool, path string, params map[string]string, headers map[string]string, responsePath string, _ io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolveReadWithStrategyResponsePathAndJSONGuard(ctx, c, flags, "live", resourceType, false, path, params, headers, responsePath, true, nil)
}

func resolveReadWithStrategyResponsePathAndJSONGuard(ctx context.Context, c *client.Client, flags *rootFlags, _ string, resourceType string, _ bool, path string, params map[string]string, headers map[string]string, responsePath string, guardLiveJSON bool, _ io.Writer) (json.RawMessage, DataProvenance, error) {
	data, err := c.GetWithHeaders(ctx, path, params, headers)
	if err != nil {
		return nil, DataProvenance{}, err
	}
	prov := DataProvenance{Source: "live", ResourceType: resourceType}
	if isDryRunResponse(c.IsDryRun(), data) {
		prov.Source = "dry-run"
		return data, attachFreshness(prov, flags), nil
	}
	if guardLiveJSON {
		if err := assertLiveJSONBody(data); err != nil {
			return nil, DataProvenance{}, err
		}
	}
	return applyResponsePath(data, responsePath), attachFreshness(prov, flags), nil
}

func resolvePaginatedRead(ctx context.Context, c *client.Client, flags *rootFlags, resourceType string, path string, params map[string]string, headers map[string]string, fetchAll bool, cursorParam, paginationType, limitParam string, defaultPageSize int, nextCursorPath, hasMoreField string, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolvePaginatedReadWithStrategy(ctx, c, flags, "live", resourceType, path, params, headers, fetchAll, cursorParam, paginationType, limitParam, defaultPageSize, nextCursorPath, hasMoreField, "", hintWriter)
}

func resolvePaginatedReadWithStrategy(ctx context.Context, c *client.Client, flags *rootFlags, _ string, resourceType string, path string, params map[string]string, headers map[string]string, fetchAll bool, cursorParam, paginationType, limitParam string, defaultPageSize int, nextCursorPath, hasMoreField, responsePath string, hintWriter io.Writer) (json.RawMessage, DataProvenance, error) {
	return resolvePaginatedReadWithStrategyAndJSONGuard(ctx, c, flags, "live", resourceType, path, params, headers, fetchAll, cursorParam, paginationType, limitParam, defaultPageSize, nextCursorPath, hasMoreField, responsePath, true, liveAllRejectNone, hintWriter)
}

func resolvePaginatedReadWithStrategyAndJSONGuard(ctx context.Context, c *client.Client, flags *rootFlags, _ string, resourceType string, path string, params map[string]string, headers map[string]string, fetchAll bool, cursorParam, paginationType, limitParam string, defaultPageSize int, nextCursorPath, hasMoreField, responsePath string, guardLiveJSON bool, liveAllReject liveAllRejectReason, _ io.Writer) (json.RawMessage, DataProvenance, error) {
	if flags != nil && flags.dryRun {
		fetchAll = false
	}
	if err := liveAllUnsupportedError(liveAllReject, false); err != nil {
		return nil, DataProvenance{}, err
	}
	data, err := paginatedGetWithResponsePath(ctx, c, path, params, headers, fetchAll, cursorParam, paginationType, limitParam, defaultPageSize, nextCursorPath, hasMoreField, responsePath)
	if err != nil {
		return nil, DataProvenance{}, err
	}
	prov := DataProvenance{Source: "live", ResourceType: resourceType}
	if isDryRunResponse(c.IsDryRun(), data) {
		prov.Source = "dry-run"
		return data, attachFreshness(prov, flags), nil
	}
	if guardLiveJSON {
		if err := assertLiveJSONBody(data); err != nil {
			return nil, DataProvenance{}, err
		}
	}
	return data, attachFreshness(prov, flags), nil
}
