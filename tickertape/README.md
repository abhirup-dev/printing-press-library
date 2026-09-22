# Tickertape CLI

**Live, provenance-aware Tickertape research without stale local market data**

Use Tickertape's public Indian and US research surfaces from the terminal: company fundamentals, scorecards, mutual funds, screen metadata, MMI, deal ideas, and US assets. Every domain-data call is live and carries source and access context; locked or entitlement-aware fields are surfaced rather than silently omitted.

Learn more at [Tickertape](https://api.tickertape.in).

Created by [@dev-abhirup-sc](https://github.com/dev-abhirup-sc) (dev-abhirup-sc).

## Install

The recommended path installs both the `tickertape-pp-cli` binary and the `pp-tickertape` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install tickertape
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install tickertape --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install tickertape --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install tickertape --agent claude-code
npx -y @mvanhorn/printing-press-library install tickertape --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.6 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/other/tickertape/cmd/tickertape-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/tickertape-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install tickertape --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-tickertape --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-tickertape --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install tickertape --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/tickertape-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/other/tickertape/cmd/tickertape-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "tickertape": {
      "command": "tickertape-pp-mcp"
    }
  }
}
```

</details>

## Authentication

Public research commands require no credential. The current EgoBrowser profile demonstrated authenticated page traffic for user, credit, portfolio, ratings, forecast, AI summary, and deal routes; when a supported secure credential/session mechanism is wired for optional entitlement-aware calls, credentials may live only in the secure auth store or permissioned auth config. Never copy credentials, cookies, headers, or personal portfolio values into output or domain-data storage.

## Quick Start

```bash
# Verify the live-read configuration without making a domain-data request.
tickertape-pp-cli doctor --dry-run

# Start with current market mood and component indicators.
tickertape-pp-cli market mood --agent

# Inspect score dimensions while preserving locked fields.
tickertape-pp-cli company scorecard RELI --agent

# Read a mutual-fund summary once an ID is known.
tickertape-pp-cli mutualfund summary INF179K01BB08 --agent

# Use the same CLI for a US asset overview.
tickertape-pp-cli us security-overview VOO --agent

```

## Unique Features

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

## Usage

Run `tickertape-pp-cli --help` for the full command reference and flag list.

## Persistence and authentication

Tickertape domain data is live-only. The CLI does not create or read a local
SQLite store, response cache, sync snapshot, history database, or cursor store.
Every research command fetches the provider route at invocation time and emits
`source: live` plus observation metadata. Credentials and session material are
kept in Keychain or process environment only and are never included in output
or domain-data storage.

Use `tickertape-pp-cli doctor` to verify configuration and API reachability.
Use `tickertape-pp-cli auth status --json` to verify authenticated capability.
## Commands

### company

Indian company identity, fundamentals, reports, scorecards, and public signals.

- **`tickertape-pp-cli company aggregated-deals`** - Entitlement-aware company deal signals.
- **`tickertape-pp-cli company ai-summary`** - Entitlement-aware AI summary; preserves explicit 401 or locked access metadata.
- **`tickertape-pp-cli company chart-inter`** - Inter-day historical stock chart.
- **`tickertape-pp-cli company chart-intra`** - Intra-day stock chart.
- **`tickertape-pp-cli company checklist`** - Public company investment checklist records.
- **`tickertape-pp-cli company financials`** - Company income statement, balance sheet, or cash flow rows.
- **`tickertape-pp-cli company forecast`** - Entitlement-aware analyst forecast envelope.
- **`tickertape-pp-cli company info`** - Company identity, labels, ratios, and quote metadata.
- **`tickertape-pp-cli company news`** - Company news records.
- **`tickertape-pp-cli company ratings`** - Entitlement-aware analyst ratings envelope.
- **`tickertape-pp-cli company ratios`** - Key ratio records for a company.
- **`tickertape-pp-cli company scorecard`** - Six-dimension scorecard with locked, rank, callout, and element metadata.
- **`tickertape-pp-cli company smallcases`** - Public smallcase references associated with a company.
- **`tickertape-pp-cli company summary`** - Financial summary, events, news, peers, forecast envelope, and presentations.

### user / credit / portfolio

Read-only authenticated account and entitlement checks.

- **`tickertape-pp-cli user status`** - Live authenticated identity status.
- **`tickertape-pp-cli credit summary`** - Live credit and entitlement summary.
- **`tickertape-pp-cli portfolio status`** - Live holdings synchronization status.

### deals

Public market-mover and deal-idea surfaces.

- **`tickertape-pp-cli deals ideas`** - Public deal ideas and market-mover highlights.
- **`tickertape-pp-cli deals insight`** - Deal insights by type, duration, and sort; required values are source-specific.

### homepage

Public event and publication stream.

- **`tickertape-pp-cli homepage`** - Paginated homepage events with optional asset and type filters.

### market

Live market-wide status, mood, and quote surfaces.

- **`tickertape-pp-cli market mood`** - Current Market Mood Index and component indicators.
- **`tickertape-pp-cli market quotes`** - Snapshot quotes for comma-separated Tickertape security IDs.
- **`tickertape-pp-cli market status`** - Market open, closed, holiday, and trading-window status.
- **`tickertape-pp-cli market us-latest-quotes`** - Latest US quotes keyed by ticker.

### mutualfund

Mutual-fund discovery, summary, holdings, charts, managers, and checklists.

- **`tickertape-pp-cli mutualfund chart-inter`** - Mutual-fund historical NAV or return chart.
- **`tickertape-pp-cli mutualfund chart-sip`** - Mutual-fund SIP chart.
- **`tickertape-pp-cli mutualfund checklist`** - Mutual-fund investment checklist records.
- **`tickertape-pp-cli mutualfund holdings`** - Allocation history, current allocation, sectors, and red-flag counts.
- **`tickertape-pp-cli mutualfund info`** - Mutual-fund identity, AMC, NAV, option, and sector labels.
- **`tickertape-pp-cli mutualfund list`** - Mutual-fund universe and identifiers.
- **`tickertape-pp-cli mutualfund managers`** - Fund-manager records.
- **`tickertape-pp-cli mutualfund summary`** - Mutual-fund peers, scheme information, ratios, CAGR, tax, and AMC details.
- **`tickertape-pp-cli mutualfund widget`** - Compact mutual-fund widget and score data.

### screener

Equity and mutual-fund filter catalogs, prebuilt screens, universes, and optional queries.

- **`tickertape-pp-cli screener equity-filters`** - Equity screener filter metadata with premium and locked markers.
- **`tickertape-pp-cli screener equity-prebuilt`** - Equity prebuilt screen catalog.
- **`tickertape-pp-cli screener equity-query`** - Read-only equity screener query; premium fields may return access errors.
- **`tickertape-pp-cli screener equity-universes`** - Equity screener universe metadata.
- **`tickertape-pp-cli screener mf-filters`** - Mutual-fund screener filter metadata.
- **`tickertape-pp-cli screener mf-prebuilt`** - Mutual-fund prebuilt screen catalog.
- **`tickertape-pp-cli screener mf-query`** - Read-only mutual-fund screener query; premium fields may return access errors.
- **`tickertape-pp-cli screener mf-universes`** - Mutual-fund screener universe metadata.

### search_resource

Search and symbol discovery.

- **`tickertape-pp-cli search-resource <q>`** - Suggest Indian assets and return search metadata.

### us

US security and ETF research through Tickertape's global market service.

- **`tickertape-pp-cli us chart`** - US security intra-day or inter-day chart.
- **`tickertape-pp-cli us etf-info`** - US ETF identity and asset metadata.
- **`tickertape-pp-cli us etf-overview`** - US ETF overview with type and top holdings.
- **`tickertape-pp-cli us filters`** - US security and ETF filter metadata.
- **`tickertape-pp-cli us financials`** - US stock income, balance-sheet, or cash-flow financials.
- **`tickertape-pp-cli us security-info`** - US security identity and asset metadata.
- **`tickertape-pp-cli us security-overview`** - US stock overview with metrics, labels, holdings, and peers.


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
tickertape-pp-cli mutualfund list

# JSON for scripting and agents
tickertape-pp-cli mutualfund list --json
# Filter to specific fields by name
tickertape-pp-cli mutualfund list --json --select <field>[,<field>...]

# Dry run — show the request without sending
tickertape-pp-cli mutualfund list --dry-run

# Agent mode — JSON + compact + no prompts in one flag
tickertape-pp-cli mutualfund list --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select <field>[,<field>...]` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Live-only** - commands never fall back to stale local domain data
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `5` API error, `6` partial failure, `7` rate limited, `10` config error.

## Health Check

```bash
tickertape-pp-cli doctor
```

Verifies configuration and connectivity to the API.

## Configuration

Run `tickertape-pp-cli doctor` to verify the secure configuration and live API reachability. Domain-data, cache, state, and local-store relocation flags are intentionally unsupported.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

## Troubleshooting
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **A response is 401, 403, or contains locked fields** — Treat it as access metadata; configure only the approved secure auth mechanism and never copy cookies or tokens into command arguments or files.
- **A screener query returns 400** — Use screen filters, prebuilt, or universes first; query bodies and premium-field behavior are intentionally not guessed.
- **A chart or event query rejects a duration/type value** — Use the values exposed by the live metadata/page contract and preserve the upstream error instead of retrying guessed enums.
- **A route is slow or rate-limited** — Retry conservatively with backoff and rerun the live command; the CLI does not serve cached domain data.
