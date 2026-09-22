---
name: pp-tickertape
description: "Live, provenance-aware Tickertape research without stale local market data. Trigger phrases: `research RELI on Tickertape`, `show the Tickertape scorecard`, `check the Market Mood Index`, `look up a mutual fund on Tickertape`, `use Tickertape`, `run tickertape`."
author: "dev-abhirup-sc"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - tickertape-pp-cli
    install:
      - kind: go
        bins: [tickertape-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/other/tickertape/cmd/tickertape-pp-cli
---

# Tickertape — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `tickertape-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install tickertape --cli-only
   ```
2. Verify: `tickertape-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.6 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/other/tickertape/cmd/tickertape-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

Use Tickertape's public Indian and US research surfaces from the terminal: company fundamentals, scorecards, mutual funds, screen metadata, MMI, deal ideas, and US assets. Every domain-data call is live and carries source and access context; locked or entitlement-aware fields are surfaced rather than silently omitted.

## When to Use This CLI

Use Tickertape for live, read-only investment research across Indian stocks, mutual funds, ETFs, indices, and US securities. Reach for it when scorecards, red flags, entry-point signals, MMI, deal ideas, or provenance-aware access state matter. Use a broker tool for orders and authoritative holdings, and a fundamentals source for deeper source-linked operating history.

## Anti-triggers

Do not use this CLI for:
- Do not use this CLI for orders, trades, broker linking, fills, taxes, or authoritative executed holdings.
- Do not use it for watchlists, saved screens, portfolio mutations, social actions, exports, payments, loans, gold, or account opening.
- Do not treat Tickertape scorecards or MMI as a universal investment recommendation or as a cross-source consensus.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Transparent research
- **`company scorecard`** — Show Tickertape score dimensions, locked fields, entry-point signals, and red flags without silently dropping premium data.

  _Choose this when an agent needs transparent Tickertape opinion data and must distinguish locked from absent fields._

  ```bash
  tickertape-pp-cli company scorecard RELI --agent
  ```
- **`company timeline`** — Combine live company events, news, results, and investor-presentation metadata into a dated source-labelled stream.

  _Choose this for a current company-history pass before drilling into individual reports._

  ```bash
  tickertape-pp-cli company timeline RELI --limit 20 --agent
  ```
- **`inspect access`** — Explain source host, route, observation time, entity identity, lock state, and entitlement errors for any live result.

  _Choose this when an agent must explain whether a result is public, locked, authenticated, or unavailable._

  ```bash
  tickertape-pp-cli inspect access company scorecard RELI --agent
  ```

### Cross-asset triage
- **`lookup`** — Triages Indian stocks, mutual funds, ETFs, indices, and US securities through one live command while retaining source-specific identity fields.

  _Choose this when the asset class is unknown or an agent needs consistent first-pass research._

  ```bash
  tickertape-pp-cli lookup VOO --market US --agent
  ```

### Market context
- **`market brief`** — Turn the live Market Mood Index payload into a concise timestamped brief of fear/greed, FII, momentum, volatility, and component signals.

  _Choose this for current market context, not as a trading signal or historical series._

  ```bash
  tickertape-pp-cli market brief --agent
  ```

### Agent-native synthesis
- **`company brief`** — Compose live company identity, summary, scorecard, reports, and access state into one bounded triage response.

  _Choose this for the first pass on a company when raw endpoint-by-endpoint calls would waste agent context._

  ```bash
  tickertape-pp-cli company brief RELI --agent
  ```
- **`screen catalog`** — Organize equity and mutual-fund filters, categories, universes, and prebuilt screens while preserving premium markers.

  _Choose this to discover what can be screened before deciding whether a live query is safe and supported._

  ```bash
  tickertape-pp-cli screen catalog --asset-class equity --agent
  ```

## Command Reference

**company** — Indian company identity, fundamentals, reports, scorecards, and public signals.

- `tickertape-pp-cli company aggregated-deals` — Entitlement-aware company deal signals.
- `tickertape-pp-cli company ai-summary` — Entitlement-aware AI summary; preserves explicit 401 or locked access metadata.
- `tickertape-pp-cli company chart-inter` — Inter-day historical stock chart.
- `tickertape-pp-cli company chart-intra` — Intra-day stock chart.
- `tickertape-pp-cli company checklist` — Public company investment checklist records.
- `tickertape-pp-cli company financials` — Company income statement, balance sheet, or cash flow rows.
- `tickertape-pp-cli company forecast` — Entitlement-aware analyst forecast envelope.
- `tickertape-pp-cli company info` — Company identity, labels, ratios, and quote metadata.
- `tickertape-pp-cli company news` — Company news records.
- `tickertape-pp-cli company ratings` — Entitlement-aware analyst ratings envelope.
- `tickertape-pp-cli company ratios` — Key ratio records for a company.
- `tickertape-pp-cli company scorecard` — Six-dimension scorecard with locked, rank, callout, and element metadata.
- `tickertape-pp-cli company smallcases` — Public smallcase references associated with a company.
- `tickertape-pp-cli company summary` — Financial summary, events, news, peers, forecast envelope, and presentations.

**deals** — Public market-mover and deal-idea surfaces.

- `tickertape-pp-cli deals ideas` — Public deal ideas and market-mover highlights.
- `tickertape-pp-cli deals insight` — Deal insights by type, duration, and sort; required values are source-specific.

**homepage** — Public event and publication stream.

- `tickertape-pp-cli homepage` — Paginated homepage events with optional asset and type filters.

**market** — Live market-wide status, mood, and quote surfaces.

- `tickertape-pp-cli market mood` — Current Market Mood Index and component indicators.
- `tickertape-pp-cli market quotes` — Snapshot quotes for comma-separated Tickertape security IDs.
- `tickertape-pp-cli market status` — Market open, closed, holiday, and trading-window status.
- `tickertape-pp-cli market us-latest-quotes` — Latest US quotes keyed by ticker.

**mutualfund** — Mutual-fund discovery, summary, holdings, charts, managers, and checklists.

- `tickertape-pp-cli mutualfund chart-inter` — Mutual-fund historical NAV or return chart.
- `tickertape-pp-cli mutualfund chart-sip` — Mutual-fund SIP chart.
- `tickertape-pp-cli mutualfund checklist` — Mutual-fund investment checklist records.
- `tickertape-pp-cli mutualfund holdings` — Allocation history, current allocation, sectors, and red-flag counts.
- `tickertape-pp-cli mutualfund info` — Mutual-fund identity, AMC, NAV, option, and sector labels.
- `tickertape-pp-cli mutualfund list` — Mutual-fund universe and identifiers.
- `tickertape-pp-cli mutualfund managers` — Fund-manager records.
- `tickertape-pp-cli mutualfund summary` — Mutual-fund peers, scheme information, ratios, CAGR, tax, and AMC details.
- `tickertape-pp-cli mutualfund widget` — Compact mutual-fund widget and score data.

**screener** — Equity and mutual-fund filter catalogs, prebuilt screens, universes, and optional queries.

- `tickertape-pp-cli screener equity-filters` — Equity screener filter metadata with premium and locked markers.
- `tickertape-pp-cli screener equity-prebuilt` — Equity prebuilt screen catalog.
- `tickertape-pp-cli screener equity-query` — Read-only equity screener query; premium fields may return access errors.
- `tickertape-pp-cli screener equity-universes` — Equity screener universe metadata.
- `tickertape-pp-cli screener mf-filters` — Mutual-fund screener filter metadata.
- `tickertape-pp-cli screener mf-prebuilt` — Mutual-fund prebuilt screen catalog.
- `tickertape-pp-cli screener mf-query` — Read-only mutual-fund screener query; premium fields may return access errors.
- `tickertape-pp-cli screener mf-universes` — Mutual-fund screener universe metadata.

**search_resource** — Search and symbol discovery.

- `tickertape-pp-cli search-resource <q>` — Suggest Indian assets and return search metadata.

**us** — US security and ETF research through Tickertape's global market service.

- `tickertape-pp-cli us chart` — US security intra-day or inter-day chart.
- `tickertape-pp-cli us etf-info` — US ETF identity and asset metadata.
- `tickertape-pp-cli us etf-overview` — US ETF overview with type and top holdings.
- `tickertape-pp-cli us filters` — US security and ETF filter metadata.
- `tickertape-pp-cli us financials` — US stock income, balance-sheet, or cash-flow financials.
- `tickertape-pp-cli us security-info` — US security identity and asset metadata.
- `tickertape-pp-cli us security-overview` — US stock overview with metrics, labels, holdings, and peers.


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
tickertape-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query. `--json` (and other machine formats) keep that exit-2 contract and write `{"matches":[]}` on stdout so agents can inspect the envelope without treating a miss as success.

## Recipes

### Explain a stock scorecard

```bash
tickertape-pp-cli company scorecard RELI --agent --select data
```

Keep the scorecard envelope and locked/access fields while narrowing agent output.

### Build a live company brief

```bash
tickertape-pp-cli company brief RELI --agent
```

Use the composed triage command for identity, summary, scorecard, and timeline context.

### Inspect the MMI

```bash
tickertape-pp-cli market brief --agent
```

Summarize current mood and component indicators without storing a snapshot.

### Find available screen metadata

```bash
tickertape-pp-cli screen catalog --asset-class equity --agent
```

Discover live filters and presets before attempting a query.

### Compare a US security

```bash
tickertape-pp-cli us security-overview VOO --agent --select data
```

Use the same provenance-aware output for US research.

## Auth Setup

Public research commands require no credential. The current EgoBrowser profile demonstrated authenticated page traffic for user, credit, portfolio, ratings, forecast, AI summary, and deal routes; when a supported secure credential/session mechanism is wired for optional entitlement-aware calls, credentials may live only in the secure auth store or permissioned auth config. Never copy credentials, cookies, headers, or personal portfolio values into output or domain-data storage.

Run `tickertape-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color`.

Global format flags share one contract on promoted and novel live-read paths:

- `--json` — one JSON document on stdout
- `--compact` — keep identity/status/timestamp fields; does not change the document vs stream shape
- `--csv` / `--plain` — tabular rows (collection envelopes unwrap to the row array)
- `--quiet` — one identity value per row, no envelope

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  tickertape-pp-cli mutualfund list --agent
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Live-only** — no sync, local store, response cache, or stale fallback is available
- **Non-interactive** — never prompts, every input is a flag

### Response envelope

Live commands wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live", "freshness": {"observed_at": "..."}},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to confirm the live source. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal AND no machine-format flag (`--json`, `--csv`, `--compact`, `--quiet`, `--plain`, `--select`) is set — piped/agent consumers and explicit-format runs get pure JSON on stdout.

## Persistence and authentication

Domain data is fetched live only. The CLI does not create or read SQLite,
response caches, sync snapshots, history, cursor stores, feedback logs, local
profiles, or output-delivery files. Credentials and session metadata may be provided only through the native
Keychain, process environment overrides, or permissioned auth configuration,
and are never included in results or domain-data storage.

`doctor` reports live configuration and API reachability. `agent-context`
intentionally does not expose data, state, or cache paths.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 5 | API error (upstream issue) |
| 6 | Partial failure |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `tickertape-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/other/tickertape/cmd/tickertape-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add tickertape-pp-mcp -- tickertape-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which tickertape-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   tickertape-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `tickertape-pp-cli <command> --help`.
