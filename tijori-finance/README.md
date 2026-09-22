# Tijori Finance CLI

**A source-linked terminal research notebook for Indian equities.**

Tijori Research CLI combines fundamentals, reports, events, screeners, market context, and permitted portfolio reads without pretending those sources are interchangeable. Every invocation stays live, preserves units and provenance, and keeps execution outside the tool.

## Install

The recommended path installs both the `tijori-finance-pp-cli` binary and the `pp-tijori-finance` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install tijori-finance
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install tijori-finance --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install tijori-finance --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install tijori-finance --agent claude-code
npx -y @mvanhorn/printing-press-library install tijori-finance --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.6 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/other/tijori-finance/cmd/tijori-finance-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/tijori-finance-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install tijori-finance --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-tijori-finance --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-tijori-finance --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install tijori-finance --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Runtime contract

This is a standalone CLI, not an MCP server. Every domain read is a fresh GET against Tijori; the binary does not provide sync, history, SQLite, domain-data cache, or mutation commands. Outputs retain live-source provenance and never include credentials or cookies.

## Authentication

Public research works without a session. Portfolio, watchlist, alerts, and self-timeline reads use the approved secure browser-session bridge: pipe one auth-header line or session-material JSON to `auth import --stdin`; on macOS this uses Keychain with a 0600 file fallback. Run `auth status --json` and `doctor --json` for sanitized live capability checks, then `auth logout --json` to clear local secure material. Never place Tijori credentials, cookies, CSRF values, or session material in shell history or output.

## Quick Start

```bash
# Check local configuration without contacting a mutation endpoint.
tijori-finance-pp-cli doctor --dry-run

# Resolve a company to its stable slug and ID.
tijori-finance-pp-cli company search "Tata Steel" --json

# Inspect periods and units from the source table.
tijori-finance-pp-cli company financials tata-steel-limited --json

# Attach report and document provenance to the research record.
tijori-finance-pp-cli reports list --json

```

## Unique Features

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

## Usage

Run `tijori-finance-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

Only configuration and credential references may be persisted. Domain responses are not written to disk. `--home` and `TIJORI_FINANCE_HOME` may relocate the secure config directory used by the generated runtime, but there is no data, state, cache, sync, or history store.

## Commands

### alerts

Authenticated read-only price alerts

- **`tijori-finance-pp-cli alerts`** - List active or historical alerts

### company

Company research and fundamentals

- **`tijori-finance-pp-cli company financials`** - Fetch annual and quarterly financial statement tables
- **`tijori-finance-pp-cli company fund-flow`** - Fetch fund-flow sources and uses
- **`tijori-finance-pp-cli company knowledge-base`** - List company report and investor document links
- **`tijori-finance-pp-cli company market-share`** - Fetch company market-share cards
- **`tijori-finance-pp-cli company operational-metrics`** - Fetch one operational metric time series
- **`tijori-finance-pp-cli company overview`** - Fetch a company overview page
- **`tijori-finance-pp-cli company revenue-mix`** - Fetch revenue mix page data
- **`tijori-finance-pp-cli company search`** - Search companies by name
- **`tijori-finance-pp-cli company shareholding`** - Fetch historical shareholding tables

### feed

Company events and timeline

- **`tijori-finance-pp-cli feed company`** - Fetch company-specific event data
- **`tijori-finance-pp-cli feed timeline`** - Fetch a cursor page of self timeline events

### ideas

Market discovery and ideas

- **`tijori-finance-pp-cli ideas`** - Fetch ideas dashboard metadata and links

### macro

Indian macro indicators

- **`tijori-finance-pp-cli macro`** - Fetch industry, demand, and GDP tables

### market

Market and group context

- **`tijori-finance-pp-cli market conglomerate-constituents`** - Fetch conglomerate constituents
- **`tijori-finance-pp-cli market monitor`** - Fetch market monitor tables
- **`tijori-finance-pp-cli market sector-constituents`** - Fetch niche sector constituents

### portfolio

Authenticated read-only portfolio views

- **`tijori-finance-pp-cli portfolio company-summary`** - Fetch one portfolio's company holdings
- **`tijori-finance-pp-cli portfolio family-company-summary`** - Fetch family holdings by company
- **`tijori-finance-pp-cli portfolio family-gainers-losers`** - Fetch family portfolio gainers and losers
- **`tijori-finance-pp-cli portfolio family-summary`** - Fetch family portfolio summary
- **`tijori-finance-pp-cli portfolio gainers-losers`** - Fetch one portfolio's gainers and losers
- **`tijori-finance-pp-cli portfolio summary`** - Fetch one portfolio summary

### raw_materials

Commodity and raw-material context

- **`tijori-finance-pp-cli raw-materials`** - Fetch chemical, spread, and metal price tables

### reports

Reports, results, concalls, and upcoming events

- **`tijori-finance-pp-cli reports concalls`** - Fetch the concall monitor
- **`tijori-finance-pp-cli reports list`** - List financial report cards
- **`tijori-finance-pp-cli reports quarterly-results`** - Fetch quarterly results tables
- **`tijori-finance-pp-cli reports upcoming`** - Fetch upcoming self-scoped result/AGM/dividend events

### screener

Native Tijori screener DSL and field catalog

- **`tijori-finance-pp-cli screener advanced-search`** - Run a native Tijori screener query
- **`tijori-finance-pp-cli screener fields`** - Search the screener field catalog
- **`tijori-finance-pp-cli screener popular`** - List popular screener cards
- **`tijori-finance-pp-cli screener popular-results`** - Run a saved popular screener query

### watchlist

Authenticated read-only watchlist views

- **`tijori-finance-pp-cli watchlist <watchlist_id> <page>`** - Fetch one watchlist data page


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
tijori-finance-pp-cli alerts

# JSON for scripting and agents
tijori-finance-pp-cli alerts --json
# Filter to specific fields by name
tijori-finance-pp-cli alerts --json --select <field>[,<field>...]

# Dry run — show the request without sending
tijori-finance-pp-cli alerts --dry-run

# Agent mode — JSON + compact + no prompts in one flag
tijori-finance-pp-cli alerts --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select <field>[,<field>...]` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Explicit confirmation** - `--agent` does not imply `--yes`; pass `--yes` separately only after the target, arguments, and side effects are clear
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Live-only** - every domain read is fetched on demand; no local SQLite or domain-data cache is used
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
tijori-finance-pp-cli doctor
```

Verifies configuration and connectivity to the API.

## Configuration

Run `tijori-finance-pp-cli doctor` to inspect connectivity and the resolved secure config path. The platform-default config path is `~/.config/tijori-finance-pp-cli/config.toml`; `--home` and `TIJORI_FINANCE_HOME` can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence. The approved browser-session bridge should pipe one auth-header line or session-material JSON directly into the secure store; it must not print or log the value:

```bash
# The browser bridge owns the left side and writes only to the pipe.
<approved-browser-session-pipe> | ./tijori-finance-pp-cli auth import --stdin --json
./tijori-finance-pp-cli auth status --json
./tijori-finance-pp-cli doctor --json
./tijori-finance-pp-cli auth logout --json
```

On macOS the store uses Keychain; when Keychain is unavailable it falls back to a 0600 file. `auth status` reports only sanitized store metadata and live GET capability states. A process-scoped `TIJORI_FINANCE_AUTH_HEADER` override remains available for CI and must be unset after use.

## Troubleshooting
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **Public company data works but portfolio reads report authentication required** — Run `tijori-finance-pp-cli auth capabilities --json` and refresh the approved session with the explicit auth command.
- **A document link is blocked or returns an unexpected host** — Use the metadata/source URL output; do not force-fetch an unvalidated external document.
- **A timeline page is empty** — Check the returned cursor, selector scope, and endpoint status; empty is not the same as authentication failure.
