# Kite by Zerodha CLI

**Live broker truth for read-only portfolio and execution analytics, with historical Console coverage kept honest.**

Kite by Zerodha exposes current holdings, positions, margins, orders, trades, and order-to-fill linkage through official Kite Connect. The CLI adds agent-native summaries while refusing to turn transient broker responses or private Console UI traffic into a fake historical ledger.

## Install

The recommended path installs both the `kite-zerodha-pp-cli` binary and the `pp-kite-zerodha` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install kite-zerodha
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install kite-zerodha --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install kite-zerodha --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install kite-zerodha --agent claude-code
npx -y @mvanhorn/printing-press-library install kite-zerodha --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.6 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/other/kite-zerodha/cmd/kite-zerodha-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/kite-zerodha-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install kite-zerodha --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-kite-zerodha --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-kite-zerodha --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install kite-zerodha --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/kite-zerodha-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `KITE_CONNECT_TOKEN` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/other/kite-zerodha/cmd/kite-zerodha-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "kite-zerodha": {
      "command": "kite-zerodha-pp-mcp",
      "env": {
        "KITE_CONNECT_TOKEN": "<your-key>"
      }
    }
  }
}
```

</details>

## Authentication

Set KITE_CONNECT_TOKEN through an approved secure credential mechanism. Its value is the official Kite Connect composite api_key:access_token credential; never put it in source, logs, reports, or normal output. Zerodha Console login is browser-session-only evidence in this build and is not replayed.

## Quick Start

```bash
# Confirm CLI wiring without contacting the broker.
kite-zerodha-pp-cli doctor --dry-run

# Read the live portfolio view once secure auth is configured.
kite-zerodha-pp-cli portfolio summary --agent

# Trace partial fills for one order.
kite-zerodha-pp-cli activity fills --order-id example-order-id --agent

# Check the official-versus-deferred source boundary.
kite-zerodha-pp-cli source status --agent --select sources

```

## Unique Features

These capabilities aren't available in any other tool for this API.

### Live broker analytics
- **`portfolio summary`** — Combine live holdings, positions, and margin state into one current portfolio view.

  _Use when an agent needs the current broker-grounded portfolio in one response._

  ```bash
  kite-zerodha-pp-cli portfolio summary --agent
  ```
- **`activity buys`** — Summarize current-day buy executions with quantities, prices, and turnover.

  _Use for a current-day buy or sell summary without mistaking orders for fills._

  ```bash
  kite-zerodha-pp-cli activity buys --agent --select trades
  ```
- **`analytics overview`** — Compute current P&L, concentration, turnover, and win/loss summaries from live rows.

  _Use for a current-state portfolio analysis when a local history store is prohibited._

  ```bash
  kite-zerodha-pp-cli analytics overview --by symbol --agent
  ```

### Execution reasoning
- **`activity fills`** — Show each order with the executions and partial fills it spawned.

  _Use when explaining how a specific order actually executed._

  ```bash
  kite-zerodha-pp-cli activity fills --order-id example-order-id --agent
  ```

### Transport boundaries
- **`source status`** — Report which broker and deferred Console capabilities are safe to use.

  _Use before asking for historical tradebook, True P&L, charges, or corporate-action-aware history._

  ```bash
  kite-zerodha-pp-cli source status --agent --select sources
  ```

## Recipes

### Current portfolio

```bash
kite-zerodha-pp-cli portfolio summary --agent
```

Join the live holdings, positions, and margin snapshot.

### Buy execution summary

```bash
kite-zerodha-pp-cli activity buys --agent --select trades
```

Keep the agent payload focused on current buy fills.

### Order partial fills

```bash
kite-zerodha-pp-cli activity fills --order-id example-order-id --agent
```

Inspect the order-to-trade relationship without reconstructing it.

### Historical boundary

```bash
kite-zerodha-pp-cli source status --agent --select sources
```

Make the deferred Console history boundary explicit before analysis.

## Usage

Run `kite-zerodha-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

Runtime broker data is intentionally not stored locally. The CLI reads configuration and credentials from the approved secure mechanism, but does not create a response cache, sync database, snapshot store, browser sidecar, or financial ledger. `KITE_ZERODHA_HOME` and `--home` remain available for config relocation and compatibility with existing installations; they do not enable offline broker data.

For containers and agent sandboxes:

```bash
export KITE_ZERODHA_HOME=/srv/kite-zerodha
kite-zerodha-pp-cli doctor
```

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block if configuration relocation is needed.

## Commands

### auctions

Holdings-related auctions

- **`kite-zerodha-pp-cli auctions`** - List holdings auctions

### gtt

Read-only GTT triggers

- **`kite-zerodha-pp-cli gtt get`** - Get one GTT trigger
- **`kite-zerodha-pp-cli gtt list`** - List active and recent GTT triggers

### holdings

Current long-term holdings

- **`kite-zerodha-pp-cli holdings`** - List current long-term holdings

### instruments

Daily Kite instrument master

- **`kite-zerodha-pp-cli instruments exchange`** - Download one exchange instrument master as CSV
- **`kite-zerodha-pp-cli instruments list`** - Download the daily instrument master as CSV

### margins

Current account margin and funds state

- **`kite-zerodha-pp-cli margins get`** - Get margin state for one segment
- **`kite-zerodha-pp-cli margins list`** - Get equity and commodity margin state

### orders

Current-day orders and order history

- **`kite-zerodha-pp-cli orders get`** - Get the status history for one order
- **`kite-zerodha-pp-cli orders list`** - List current-day orders
- **`kite-zerodha-pp-cli orders trades`** - List executions spawned by one order, including partial fills

### positions

Current net and day positions

- **`kite-zerodha-pp-cli positions`** - Get current net and day positions

### trades

Current-day executions

- **`kite-zerodha-pp-cli trades`** - List current-day trades and executions

### user_profile

Authenticated Kite profile

- **`kite-zerodha-pp-cli user-profile`** - Get the authenticated Kite profile


### Stateless runtime boundary

Normal production commands do not sync, search, cache, journal, or build a local financial ledger. Each broker read issues a fresh official Kite Connect request. `--data-source local` is rejected, and response-cache bypass is forced regardless of flags. Configuration may be read from the approved credential mechanism, but account data is not persisted.

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
kite-zerodha-pp-cli auctions

# JSON for scripting and agents
kite-zerodha-pp-cli auctions --json
# Filter to specific fields
kite-zerodha-pp-cli auctions --json --select auction_number,exchange,tradingsymbol

# Dry run — show the request without sending
kite-zerodha-pp-cli auctions --dry-run

# Agent mode — JSON + compact + no prompts in one flag
kite-zerodha-pp-cli auctions --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select <field>[,<field>...]` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Read-only by default** - this CLI does not create, update, delete, publish, send, or mutate remote resources
- **Stateless** - every broker read is fresh; local snapshots, sync, search, and response caching are disabled
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
kite-zerodha-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Run `kite-zerodha-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is `~/.config/kite-zerodha/config.toml`; `--home`, `KITE_ZERODHA_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `KITE_CONNECT_TOKEN` | per_call | Yes | Set to your API credential. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `kite-zerodha-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `kite-zerodha-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $KITE_CONNECT_TOKEN`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **Kite returns an invalid or expired token error** — Refresh the approved Kite Connect access token and update KITE_CONNECT_TOKEN without printing it.
- **A current-day order or trade is absent** — Remember that Kite order and trade lists are transient; use a separately validated Console export when historical support is added.
- **Historical True P&L or statement data is requested** — Run source status; Console history is deferred until its authenticated, replayable contract is validated.
