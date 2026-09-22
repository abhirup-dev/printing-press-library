---
name: pp-kite-zerodha
description: "Live broker truth for read-only portfolio and execution analytics, with historical Console coverage kept honest. Trigger phrases: `check my Kite portfolio`, `show current Zerodha holdings`, `explain my order fills`, `analyze current broker P&L`, `use Kite by Zerodha`, `run kite-zerodha`."
author: "dev-abhirup-sc"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - kite-zerodha-pp-cli
    install:
      - kind: go
        bins: [kite-zerodha-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/other/kite-zerodha/cmd/kite-zerodha-pp-cli
---

# Kite by Zerodha — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `kite-zerodha-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install kite-zerodha --cli-only
   ```
2. Verify: `kite-zerodha-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.6 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/other/kite-zerodha/cmd/kite-zerodha-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

Kite by Zerodha exposes current holdings, positions, margins, orders, trades, and order-to-fill linkage through official Kite Connect. The CLI adds agent-native summaries while refusing to turn transient broker responses or private Console UI traffic into a fake historical ledger.

## When to Use This CLI

Use this CLI for live read-only broker/account state, current-day order and execution reasoning, and bounded analytics over the rows returned now. Use it when order-versus-trade semantics, partial fills, broker P&L, or concentration matter. Do not use it as a historical ledger until the deferred Console boundary is implemented.

## Anti-triggers

Do not use this CLI for:
- Do not use this CLI to place, modify, cancel, simulate, or convert financial positions.
- Do not use private Kite UI/BFF routes or browser cookies as normal runtime transport.
- Do not use it for charts, generic ticker research, or historical Console claims.

## Unique Capabilities

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

## Command Reference

**auctions** — Holdings-related auctions

- `kite-zerodha-pp-cli auctions` — List holdings auctions

**gtt** — Read-only GTT triggers

- `kite-zerodha-pp-cli gtt get` — Get one GTT trigger
- `kite-zerodha-pp-cli gtt list` — List active and recent GTT triggers

**holdings** — Current long-term holdings

- `kite-zerodha-pp-cli holdings` — List current long-term holdings

**instruments** — Daily Kite instrument master

- `kite-zerodha-pp-cli instruments exchange` — Download one exchange instrument master as CSV
- `kite-zerodha-pp-cli instruments list` — Download the daily instrument master as CSV

**margins** — Current account margin and funds state

- `kite-zerodha-pp-cli margins get` — Get margin state for one segment
- `kite-zerodha-pp-cli margins list` — Get equity and commodity margin state

**orders** — Current-day orders and order history

- `kite-zerodha-pp-cli orders get` — Get the status history for one order
- `kite-zerodha-pp-cli orders list` — List current-day orders
- `kite-zerodha-pp-cli orders trades` — List executions spawned by one order, including partial fills

**positions** — Current net and day positions

- `kite-zerodha-pp-cli positions` — Get current net and day positions

**trades** — Current-day executions

- `kite-zerodha-pp-cli trades` — List current-day trades and executions

**user_profile** — Authenticated Kite profile

- `kite-zerodha-pp-cli user-profile` — Get the authenticated Kite profile


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
kite-zerodha-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query. `--json` (and other machine formats) keep that exit-2 contract and write `{"matches":[]}` on stdout so agents can inspect the envelope without treating a miss as success.

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

## Auth Setup

Set KITE_CONNECT_TOKEN through an approved secure credential mechanism. Its value is the official Kite Connect composite api_key:access_token credential; never put it in source, logs, reports, or normal output. Zerodha Console login is browser-session-only evidence in this build and is not replayed.

Run `kite-zerodha-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color`.

Global format flags share one contract on promoted and novel read paths:

- `--json` — one JSON document on stdout (sync progress events go to stderr)
- `--compact` — keep identity/status/timestamp fields; does not change the document vs stream shape
- `--csv` / `--plain` — tabular rows (collection envelopes unwrap to the row array)
- `--quiet` — one identity value per row, no envelope

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  kite-zerodha-pp-cli auctions --agent --select auction_number,exchange,tradingsymbol
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Stateless** — every broker read is fresh; local snapshots, sync, search, and response caching are disabled
- **Non-interactive** — never prompts, every input is a flag
- **Read-only** — do not use this CLI for create, update, delete, publish, comment, upvote, invite, order, send, or other mutating requests

### Response envelope

Broker read commands wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live"},
  "results": <data>
}
```

Parse `.results` for data. Broker results are always live; there is no local fallback or sync provenance.

## Paths and state

Normal production commands do not create a response cache, sync database, snapshot store, browser sidecar, invocation journal, or financial ledger. Configuration relocation via `--home` / `KITE_ZERODHA_HOME` remains available for compatibility, but account data is never read from a local store. Set `KITE_CONNECT_TOKEN` through an approved secure mechanism.

## Stateless runtime

Normal production commands are read-only and stateless. They issue fresh official Kite Connect requests, never use local snapshots or sync/search state, and do not expose learning, feedback, or profile-write commands. Historical Console routes remain capability diagnostics only until replayable authentication and schemas are validated.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename). Binary-response commands write decoded payload bytes (not the base64 JSON envelope) and print a small JSON receipt on stdout; `--json`/`--csv` do not refuse when this sink is set. |
| `webhook:<url>` | POST the output body to the URL (`application/json`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

Existing profiles can be inspected with `profile list`, `profile show`, and `profile use`. Production does not expose profile save/delete mutations.

```
kite-zerodha-pp-cli profile list --json
kite-zerodha-pp-cli profile show briefing
kite-zerodha-pp-cli profile use briefing
```

Existing profiles are inspection-only in production. `agent-context` may list available profiles, but broker data never comes from a profile or local store.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `kite-zerodha-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/other/kite-zerodha/cmd/kite-zerodha-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add kite-zerodha-pp-mcp -- kite-zerodha-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which kite-zerodha-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   kite-zerodha-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `kite-zerodha-pp-cli <command> --help`.
