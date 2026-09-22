# Moneycontrol CLI

**Live Indian market news and research context for every holding review**

Moneycontrol-pp-cli turns Moneycontrol's public financial pages and price feeds into source-linked, read-only commands for Indian market research. It adds live holding context joins, event-risk views, catalyst filtering, and research packets without storing articles, portfolio data, or response history.

Learn more at [Moneycontrol](https://www.moneycontrol.com).

Created by [@dev-abhirup-sc](https://github.com/dev-abhirup-sc) (dev-abhirup-sc).

## Install

The recommended path installs both the `moneycontrol-pp-cli` binary and the `pp-moneycontrol` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install moneycontrol
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install moneycontrol --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install moneycontrol --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install moneycontrol --agent claude-code
npx -y @mvanhorn/printing-press-library install moneycontrol --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.6 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/finance/moneycontrol/cmd/moneycontrol-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/moneycontrol-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install moneycontrol --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-moneycontrol --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-moneycontrol --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install moneycontrol --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

The bundle reuses your local browser session — set it up first if you haven't:

```bash
moneycontrol-pp-cli auth login --chrome
```

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/moneycontrol-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/finance/moneycontrol/cmd/moneycontrol-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "moneycontrol": {
      "command": "moneycontrol-pp-mcp"
    }
  }
}
```

</details>

## Authentication

Public news, quote, index, research, and market commands run without authentication. Optional portfolio, watchlist, alert, and Super Pro reads are endpoint-specific and use an established secure credential/session store or the generated Chrome bootstrap where supported; credentials and session values never appear in source, logs, reports, or output. No write or transaction command is provided.

## Quick Start

```bash
# Check the CLI and live-source configuration without requiring credentials.
moneycontrol-pp-cli doctor --dry-run

# Read a live Sensex snapshot using the encoded priceapi key.
moneycontrol-pp-cli indices --key 'in;SEN'

# Fetch current financial headlines with source URLs.
moneycontrol-pp-cli articles latest

# Join current quotes and company context for supplied broker symbols.
moneycontrol-pp-cli context holdings --sc-ids RI,INFY --agent

```

## Unique Features

These capabilities aren't available in any other tool for this API.

### Live portfolio context
- **`context holdings`** — Join supplied broker holding symbols to live Moneycontrol quotes, tagged news, and event context.

  _Use this when a broker already supplies holdings and the agent needs Moneycontrol narrative context without copy-paste._

  ```bash
  moneycontrol-pp-cli context holdings --sc-ids RI,INFY --agent --select holdings.sc_id,holdings.quote,holdings.news
  ```
- **`portfolio triage`** — Emit live headlines, timestamps, URLs, and event labels for supplied external holdings.

  _Use this for a quick source-linked review of external holdings before deeper research._

  ```bash
  moneycontrol-pp-cli portfolio triage --sc-ids RI,INFY --agent --select items.sc_id,items.title,items.timestamp,items.url
  ```

### Live event context
- **`events upcoming`** — Show live filings, results, and corporate actions grouped for supplied companies.

  _Use this before a review when the agent must identify current company events from one source-limited call._

  ```bash
  moneycontrol-pp-cli events upcoming --sc-ids RI,INFY --agent
  ```
- **`catalysts`** — Filter live dated IPO, earnings, filing, and corporate-action pages by supplied symbols and time window.

  _Use this when an agent needs current catalysts around a watchlist but must not rely on cached articles._

  ```bash
  moneycontrol-pp-cli catalysts --sc-ids RI,INFY --days 30 --agent
  ```

### Live market synthesis
- **`market breadth`** — Combine live index snapshots with available breadth/change-table rows and report partial availability.

  _Use this for a compact market-opening view when index movement and breadth context are needed together._

  ```bash
  moneycontrol-pp-cli market breadth --agent
  ```

### Live news intelligence
- **`news timeline`** — Normalize a live company-tag page into URL-preserving chronological story rows.

  _Use this when the question is what the current Moneycontrol tag page says about one company, not general market news._

  ```bash
  moneycontrol-pp-cli news timeline --sc-id RI --limit 20 --agent --select articles.title,articles.timestamp,articles.url
  ```

### Live research synthesis
- **`research packet`** — Assemble live quote, price-volume, company news, filings, results, and corporate-action context for one company.

  _Use this when an agent needs a broad but bounded company snapshot rather than many uncoordinated endpoint calls._

  ```bash
  moneycontrol-pp-cli research packet --sc-id RI --agent --select quote,price_volume,news,events
  ```

## Usage

Run `moneycontrol-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data: `credentials.toml`, `data.db`, cookies, browser-session proof files, and other auth sidecars |
| `state` | Runtime state such as persisted queries, jobs, and `teach.log` |
| `cache` | Regenerable HTTP/cache files |

Each kind resolves independently. The ladder is:

1. Per-kind env var: `MONEYCONTROL_CONFIG_DIR`, `MONEYCONTROL_DATA_DIR`, `MONEYCONTROL_STATE_DIR`, or `MONEYCONTROL_CACHE_DIR`
2. `--home <dir>` for this invocation
3. `MONEYCONTROL_HOME` for a flat relocated root
4. XDG env vars: `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`
5. Platform defaults matching existing installs

For containers and agent sandboxes, prefer a single relocated root:

```bash
export MONEYCONTROL_HOME=/srv/moneycontrol
moneycontrol-pp-cli doctor
```

Under `MONEYCONTROL_HOME=/srv/moneycontrol`, the four dirs resolve to `/srv/moneycontrol/config`, `/srv/moneycontrol/data`, `/srv/moneycontrol/state`, and `/srv/moneycontrol/cache`.

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block:

```json
{
  "mcpServers": {
    "moneycontrol": {
      "command": "moneycontrol-pp-mcp",
      "env": {
        "MONEYCONTROL_HOME": "/srv/moneycontrol"
      }
    }
  }
}
```

Precedence matters in fleets: an ambient per-kind variable such as `MONEYCONTROL_DATA_DIR` overrides an explicit `--home` for that kind. Use `MONEYCONTROL_HOME` or the per-kind variables for durable fleet relocation; treat `--home` as the weaker per-invocation lever.

Relocation is one-way. Unsetting `MONEYCONTROL_HOME` does not move files back to platform defaults, and `doctor` cannot find credentials left under a former root. Move the files manually before unsetting relocation variables.

Existing installs keep working because the platform-default rung matches the legacy layout. On the first auth write, stored secrets leave `config.toml` and are consolidated into `credentials.toml` under the data directory. Run `moneycontrol-pp-cli doctor --fail-on warn` to check path and credential-location warnings in automation.

## Commands

### account

Optional authenticated Moneycontrol account and Pro reads. Read-only only.

- **`moneycontrol-pp-cli account alerts`** - Read-only current alerts page. No alert mutation is exposed.
- **`moneycontrol-pp-cli account entitlement-status`** - Authenticated account/Pro entitlement status used to select legitimate reads.
- **`moneycontrol-pp-cli account portfolio`** - Read-only current portfolio page. No portfolio mutation is exposed.
- **`moneycontrol-pp-cli account pro-stocklist`** - Super Pro stock-list widget; best-effort authenticated read.
- **`moneycontrol-pp-cli account user-details`** - Authenticated user/account detail shape used for endpoint gating; sensitive fields are redacted by the CLI.
- **`moneycontrol-pp-cli account watchlist`** - Read-only current watchlist page. No watchlist mutation is exposed.

### articles

News articles across categories and per-stock tags.

- **`moneycontrol-pp-cli articles by-category`** - News headlines within a financial category (e.g. business/markets, business/stocks, earnings, ipo, mutual-funds, commodities).
- **`moneycontrol-pp-cli articles by-tag`** - News headlines tagged to a stock or financial topic (e.g. reliance-industries, infosys, nifty).
- **`moneycontrol-pp-cli articles get`** - Full financial article page (title, description, and raw body HTML for extraction).
- **`moneycontrol-pp-cli articles latest`** - Latest financial news headlines across all categories.

### history

Live NSE stock OHLCV history from the priceapi chart feed.

- **`moneycontrol-pp-cli history`** - Historical or intraday OHLCV bars for an NSE symbol.

### indices

Indian index quotes (SENSEX, NIFTY 50, NIFTY BANK, NIFTY IT).

- **`moneycontrol-pp-cli indices`** - Live quote for an Indian index.

### markets

Financial market pages and institutional activity.

- **`moneycontrol-pp-cli markets commodities`** - Commodity, gold and silver market pages.
- **`moneycontrol-pp-cli markets corporate-actions`** - Corporate actions page; best-effort route.
- **`moneycontrol-pp-cli markets earnings`** - Earnings/results pages; best-effort route.
- **`moneycontrol-pp-cli markets fii-dii`** - FII and DII cash/F&O activity page; best-effort page extraction.
- **`moneycontrol-pp-cli markets fno`** - Futures, options and open-interest market statistics.
- **`moneycontrol-pp-cli markets forex`** - Currency and forex pages; best-effort route.
- **`moneycontrol-pp-cli markets indices-table`** - Indian indices page and change-table context.
- **`moneycontrol-pp-cli markets ipo`** - IPO news and market pages.
- **`moneycontrol-pp-cli markets mutual-funds`** - Mutual-fund news and quote pages.

### price_volume

Live stock price-volume period data from the Moneycontrol API.

- **`moneycontrol-pp-cli price-volume`** - Current and historical price-volume period data for a stock.

### research

Company research pages and stock scanners.

- **`moneycontrol-pp-cli research equity-research`** - Equity research, forecasts and recommendations page.
- **`moneycontrol-pp-cli research filings`** - Company filings page; best-effort route.
- **`moneycontrol-pp-cli research fundamental-scanner`** - Fundamental stock scanner; empty/blocked widgets are reported.
- **`moneycontrol-pp-cli research technical-scanner`** - Technical stock scanner; empty/blocked widgets are reported.

### stocks

Per-stock quote (full pricefeed: price, 52H/L, 1d-10y %chg, CAGR, volume, sector)

- **`moneycontrol-pp-cli stocks`** - Full NSE equity quote for a stock by its moneycontrol sc_id.

### trending

Trending stocks widget (most-searched tickers right now).

- **`moneycontrol-pp-cli trending`** - Currently trending stocks on moneycontrol (best-effort; empty widgets are errors).


### Self-learning loop

This CLI caches per-question discovery so repeat queries skip the walk and structurally similar queries get answered via entity substitution. The loop also self-captures: every invocation is journaled locally, and failed-flag corrections plus fresh teaches surface as candidates on the next `recall` for confirm/reject judgment. Agents call `recall` before discovery and fire `teach &` after answering. See the `## Automatic learning` section in `SKILL.md` for the full protocol.

- **`moneycontrol-pp-cli recall <query>`** - Look up cached resources for a query before running discovery
- **`moneycontrol-pp-cli teach`** - Record a query -> resource mapping (silent on success, safe to background with `&`)
- **`moneycontrol-pp-cli learnings list`** - Inspect taught rows
- **`moneycontrol-pp-cli learnings forget <query>`** - Undo a teach
- **`moneycontrol-pp-cli learnings candidates`** - List auto-captured candidates awaiting confirm/reject
- **`moneycontrol-pp-cli learnings stats`** - Local loop metrics: recall hit rate, teach-to-reuse, playbook resolution, candidate counts
- **`moneycontrol-pp-cli teach-pattern`** - Install a query/resource template up front
- **`moneycontrol-pp-cli teach-lookup`** - Add an entity mapping (e.g. country code, team alias) for pattern substitution

Pass `--no-learn` or set `MONEYCONTROL_NO_LEARN=true` to disable the loop for deterministic flows.

The local store's schema version stamp is one-way: once this version of `moneycontrol-pp-cli` opens the database, older binaries refuse it with a version error — upgrade the binary rather than downgrading.

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
moneycontrol-pp-cli articles get --slug india/tribunals-reforms-bill-2026-

# JSON for scripting and agents
moneycontrol-pp-cli articles get --slug india/tribunals-reforms-bill-2026- --json
# Filter to specific fields by name
moneycontrol-pp-cli articles get --slug india/tribunals-reforms-bill-2026- --json --select <field>[,<field>...]

# Dry run — show the request without sending
moneycontrol-pp-cli articles get --slug india/tribunals-reforms-bill-2026- --dry-run

# Agent mode — JSON + compact + no prompts in one flag
moneycontrol-pp-cli articles get --slug india/tribunals-reforms-bill-2026- --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select <field>[,<field>...]` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Read-only by default** - this CLI does not create, update, delete, publish, send, or mutate remote resources
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
moneycontrol-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Run `moneycontrol-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is ``; `--home`, `MONEYCONTROL_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `moneycontrol-pp-cli doctor` to check credentials
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **HTTP 503/403 or an empty widget response** — Retry the command once; inspect its availability metadata. Use the stable priceapi or HTML command instead of assuming empty means success.
- **Index key is rejected or returns the wrong route** — Use an encoded key such as `in\;SEN` or `in\;NSX`; the CLI preserves the required semicolon encoding.
- **Authenticated account or Pro command reports expired session** — Run the endpoint-specific auth bootstrap again through the established secure store/Chrome auth path; public commands remain usable without auth.
- **A deeper financial or forecast page returns an unsupported shape** — Read the command's stability/access metadata and treat the result as best-effort; do not substitute a silent empty success.

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

TLS certificates are verified by default. For a trusted development or self-signed endpoint only, pass `--insecure` for one invocation, set `MONEYCONTROL_SKIP_TLS_VERIFY=true` for the current environment, or set `skip_tls_verify = true` in the config file for a persistent override.

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**moneycontrol-mcp**](https://github.com/pramodhapple504-arch/moneycontrol-mcp) — Python
- [**moneycontrol-pr-1701**](https://github.com/abhirup-dev/printing-press-library) — Go
- [**MCFinEx**](https://github.com/D4T4R/MCFinEx) — Python
- [**fii-dii-activity-api**](https://github.com/chirag127/fii-dii-activity-api) — Python

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
