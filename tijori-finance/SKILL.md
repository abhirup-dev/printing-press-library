---
name: pp-tijori-finance
description: "A source-linked terminal research notebook for Indian equities. Trigger phrases: `research Tata Steel fundamentals`, `check my Tijori portfolio events`, `run a Tijori screener`, `find the source report for this result`, `use tijori-finance`, `run Tijori Finance CLI`."
author: "dev-abhirup-sc"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - tijori-finance-pp-cli
    install:
      - kind: go
        bins: [tijori-finance-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/other/tijori-finance/cmd/tijori-finance-pp-cli
---

# Tijori Finance — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `tijori-finance-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install tijori-finance --cli-only
   ```
2. Verify: `tijori-finance-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.6 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/finance/tijori-finance/cmd/tijori-finance-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

Tijori Research CLI combines fundamentals, reports, events, screeners, market context, and permitted portfolio reads without pretending those sources are interchangeable. Every invocation stays live, preserves units and provenance, and keeps execution outside the tool.

## When to Use This CLI

Use this CLI for dated Indian-equity research: statements, ownership, operating metrics, reports, filings, screeners, market context, and portfolio-aware evidence. Prefer it when the answer needs source URLs, exact units, periods, and endpoint diagnostics rather than a quote alone.

## Anti-triggers

Do not use this CLI for:
- Do not use this CLI to place, modify, cancel, or recommend trades.
- Do not use it as the broker or authoritative execution/account system.
- Do not use it to maintain a local research database or background alert daemon.
- Do not treat an AI summary as a substitute for the linked source document.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Evidence-linked research
- **`research peers`** — Compare live peer financial, ownership, and operating evidence in one unit-aware matrix.

  _Use it when a screener hit needs mechanically comparable evidence across peers._

  ```bash
  tijori-finance-pp-cli research peers tata-steel-limited --metrics revenue,opm,roe,debt,holding --json
  ```
- **`research review-queue`** — Join current holdings or watchlist names with upcoming results, events, reports, and alerts.

  _Use it for a focused review ritual instead of opening every company page manually._

  ```bash
  tijori-finance-pp-cli research review-queue --portfolio --json
  ```
- **`research changes`** — Emit deterministic row-level changes between two server-returned financial periods.

  _Use it when the question is what changed, not just what the latest table says._

  ```bash
  tijori-finance-pp-cli research changes tata-steel-limited --from Mar-25 --to Mar-26 --json
  ```
- **`research ownership-check`** — Align same-period ownership movements with financial and cash-flow changes.

  _Use it to inspect ownership changes alongside fundamentals without inferring investment advice._

  ```bash
  tijori-finance-pp-cli research ownership-check tata-steel-limited --period Jun-26 --json
  ```
- **`research result-pack`** — Connect a result or event to report, concall, and validated source-document metadata.

  _Use it when a number needs its filing or investor document attached._

  ```bash
  tijori-finance-pp-cli research result-pack tata-steel-limited --period Jun-26 --json
  ```
- **`feed inspect`** — Resolve timeline items to normalized event, report, and document metadata while preserving cursors.

  _Use it when an event feed item needs the underlying source and continuation state._

  ```bash
  tijori-finance-pp-cli feed inspect --company tata-steel-limited --json
  ```

### Agent-native plumbing
- **`screener run`** — Return the exact native query, matched fields, result page, and request provenance (use --explain).

  _Use it when a screen must be inspected or handed to another agent exactly as submitted._

  ```bash
  tijori-finance-pp-cli screener run --query '( ROE > 15 ) and ( Market Capitalization > 500 )' --explain --json
  ```

### Cross-surface context
- **`research context`** — Bundle dated company, sector, macro, and raw-material observations without a synthetic verdict.

  _Use it when company performance needs its operating environment beside it._

  ```bash
  tijori-finance-pp-cli research context deepak-nitrite-limited --sector --json
  ```
- **`portfolio quality`** — Join live portfolio weights with public company quality evidence and explicit unavailable states.

  _Use it for a current exposure review that does not silently drop inaccessible names._

  ```bash
  tijori-finance-pp-cli portfolio quality --json
  ```
- **`research input-context`** — Place current raw-material observations beside live operational and revenue-mix data.

  _Use it when input prices are part of the research question but a buy/sell conclusion is not._

  ```bash
  tijori-finance-pp-cli research input-context deepak-nitrite-limited --json
  ```

### Reachability mitigation
- **`auth capabilities`** — Report public, authenticated, empty, blocked, expired, or changed-shape status per safe GET endpoint.

  _Use it before a portfolio-aware workflow when session reachability is uncertain._

  ```bash
  tijori-finance-pp-cli auth capabilities --json
  ```

## Command Reference

**alerts** — Authenticated read-only price alerts

- `tijori-finance-pp-cli alerts` — List active or historical alerts

**company** — Company research and fundamentals

- `tijori-finance-pp-cli company financials` — Fetch annual and quarterly financial statement tables
- `tijori-finance-pp-cli company fund-flow` — Fetch fund-flow sources and uses
- `tijori-finance-pp-cli company knowledge-base` — List company report and investor document links
- `tijori-finance-pp-cli company market-share` — Fetch company market-share cards
- `tijori-finance-pp-cli company operational-metrics` — Fetch one operational metric time series
- `tijori-finance-pp-cli company overview` — Fetch a company overview page
- `tijori-finance-pp-cli company revenue-mix` — Fetch revenue mix page data
- `tijori-finance-pp-cli company search` — Search companies by name
- `tijori-finance-pp-cli company shareholding` — Fetch historical shareholding tables

**feed** — Company events and timeline

- `tijori-finance-pp-cli feed company` — Fetch company-specific event data
- `tijori-finance-pp-cli feed timeline` — Fetch a cursor page of self timeline events

**ideas** — Market discovery and ideas

- `tijori-finance-pp-cli ideas` — Fetch ideas dashboard metadata and links

**macro** — Indian macro indicators

- `tijori-finance-pp-cli macro` — Fetch industry, demand, and GDP tables

**market** — Market and group context

- `tijori-finance-pp-cli market conglomerate-constituents` — Fetch conglomerate constituents
- `tijori-finance-pp-cli market monitor` — Fetch market monitor tables
- `tijori-finance-pp-cli market sector-constituents` — Fetch niche sector constituents

**portfolio** — Authenticated read-only portfolio views

- `tijori-finance-pp-cli portfolio company-summary` — Fetch one portfolio's company holdings
- `tijori-finance-pp-cli portfolio family-company-summary` — Fetch family holdings by company
- `tijori-finance-pp-cli portfolio family-gainers-losers` — Fetch family portfolio gainers and losers
- `tijori-finance-pp-cli portfolio family-summary` — Fetch family portfolio summary
- `tijori-finance-pp-cli portfolio gainers-losers` — Fetch one portfolio's gainers and losers
- `tijori-finance-pp-cli portfolio summary` — Fetch one portfolio summary

**raw_materials** — Commodity and raw-material context

- `tijori-finance-pp-cli raw-materials` — Fetch chemical, spread, and metal price tables

**reports** — Reports, results, concalls, and upcoming events

- `tijori-finance-pp-cli reports concalls` — Fetch the concall monitor
- `tijori-finance-pp-cli reports list` — List financial report cards
- `tijori-finance-pp-cli reports quarterly-results` — Fetch quarterly results tables
- `tijori-finance-pp-cli reports upcoming` — Fetch upcoming self-scoped result/AGM/dividend events

**screener** — Native Tijori screener DSL and field catalog

- `tijori-finance-pp-cli screener advanced-search` — Run a native Tijori screener query
- `tijori-finance-pp-cli screener fields` — Search the screener field catalog
- `tijori-finance-pp-cli screener popular` — List popular screener cards
- `tijori-finance-pp-cli screener popular-results` — Run a saved popular screener query

**watchlist** — Authenticated read-only watchlist views

- `tijori-finance-pp-cli watchlist <watchlist_id> <page>` — Fetch one watchlist data page


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
tijori-finance-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query. `--json` (and other machine formats) keep that exit-2 contract and write `{"matches":[]}` on stdout so agents can inspect the envelope without treating a miss as success.

## Recipes

### From screen to evidence

```bash
tijori-finance-pp-cli screener run --query '( ROE > 15 ) and ( Market Capitalization > 500 )' --explain --json
```

Keep the native screen and its provenance before drilling into selected companies.

### Read a results package

```bash
tijori-finance-pp-cli research result-pack tata-steel-limited --period Jun-26 --json
```

Connect a result to its report and conference-call sources.

### Review portfolio evidence

```bash
tijori-finance-pp-cli research review-queue --portfolio --json
```

Join current holdings with the live event and report queue.

### Inspect only selected evidence

```bash
tijori-finance-pp-cli company financials tata-steel-limited --agent
```

Keep an agent response focused on periods, units, and provenance.

## Auth Setup

Public research works without a session. Portfolio, watchlist, alerts, and self-timeline reads use the approved secure browser-session bridge: pipe one auth-header line or session-material JSON to `auth import --stdin`; on macOS this uses Keychain with a 0600 file fallback. Run `auth status --json` and `doctor --json` for sanitized live capability checks, then `auth logout --json` to clear local secure material. Never place Tijori credentials, cookies, CSRF values, or session material in shell history or output.

Run `tijori-finance-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color`.

Global format flags share one contract on promoted and novel read paths:

- `--json` — one JSON document on stdout
- `--compact` — keep identity/status/timestamp fields; does not change the document vs stream shape
- `--csv` / `--plain` — tabular rows (collection envelopes unwrap to the row array)
- `--quiet` — one identity value per row, no envelope

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  tijori-finance-pp-cli alerts --agent
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Live-only** — every domain read is fetched on demand; no sync, history, SQLite, snapshot, or domain-data cache is used
- **Non-interactive** — never prompts, every input is a flag
- **Explicit confirmation** — `--agent` does not imply `--yes`; pass `--yes` separately only after the target, arguments, and side effects are clear

### Response envelope

Read commands wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live"},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to verify the live source. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal and no machine-format flag is set.

## Paths and state

Only configuration and credential references may be persisted. Domain responses are never written to disk. Use `--home <dir>` or `TIJORI_FINANCE_HOME` to relocate secure configuration; there are no data, state, cache, sync, or history stores.

## Output delivery

Results are written to stdout and diagnostics to stderr. File and webhook delivery are disabled so the CLI remains read-only and does not persist or POST domain data.

## Named Profiles

Profiles are optional configuration only; they never contain credentials or domain responses.


A profile is a saved set of flag values, reused across invocations. Use it when a scheduled or recurring agent reuses the same saved flags while providing different input each run.

```
tijori-finance-pp-cli profile save briefing --json
tijori-finance-pp-cli --profile briefing alerts
tijori-finance-pp-cli profile list --json
tijori-finance-pp-cli profile show briefing
tijori-finance-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `tijori-finance-pp-cli --help` output
2. **Starts with `install`** → install the CLI; otherwise → execute the requested read-only CLI command with `--agent`
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)
