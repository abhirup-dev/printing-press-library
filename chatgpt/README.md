# ChatGPT CLI

**Your ChatGPT history as a queryable local archive — verified-complete extraction, outlines, and search no other tool has.**

Turn conversations into a chronological index with real turn counts and spans, extract any thread completely with a completeness proof, navigate long threads by outline, and search message content server-side or offline. Sessions survive token expiry via cookie re-mint and optional Codex auth import.

Learn more at [ChatGPT](https://chatgpt.com).

Created by [@abhirup-dev](https://github.com/abhirup-dev) (dev-abhirup-sc).

## Install

The recommended path installs both the `chatgpt-pp-cli` binary and the `pp-chatgpt` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install chatgpt
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install chatgpt --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install chatgpt --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install chatgpt --agent claude-code
npx -y @mvanhorn/printing-press-library install chatgpt --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.6 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/ai/chatgpt/cmd/chatgpt-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/chatgpt-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install chatgpt --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-chatgpt --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-chatgpt --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install chatgpt --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/chatgpt-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `CHATGPT_TOKEN` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/ai/chatgpt/cmd/chatgpt-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "chatgpt": {
      "command": "chatgpt-pp-mcp",
      "env": {
        "CHATGPT_TOKEN": "<your-key>"
      }
    }
  }
}
```

</details>

## Authentication

No API key. Import your chatgpt.com browser session once (cookies) and the CLI mints fresh bearer tokens itself — cookies outlive the 10-day token TTL, so the session keeps working. Alternatively point it at your existing codex login (~/.codex/auth.json). Send commands are experimental and browser-gated: they consume only legitimately issued authorization and fail with a clear error otherwise.

## Quick Start

```bash
# Health check without touching the network
chatgpt-pp-cli doctor --dry-run

# Chronological inventory with turn counts and span
chatgpt-pp-cli list --since 30d --hydrate

# See the structure of a long thread first
chatgpt-pp-cli outline 6aaa52c7-daec-83ee-9c8a-e1635dbf397e

# Extract one section, timestamps preserved
chatgpt-pp-cli transcript 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --section 2

# Server-side message search with compact output
chatgpt-pp-cli find "haptics" --agent --select items.title,items.snippet

```

## Unique Features

These capabilities aren't available in any other tool for this API.

### Inventory that answers questions
- **`list`** — See every chat in chronological order with turn counts, time span, and freshness — not just titles.

  _Pick this when you need a real index of conversations by time and size instead of scrolling the sidebar._

  ```bash
  chatgpt-pp-cli list --since 30d --order created --agent
  ```
- **`transcript`** — Extract an entire conversation with a completeness proof against the live tree state.

  _Use this when an export must contain every message — audit, migration, or feeding another agent._

  ```bash
  chatgpt-pp-cli transcript 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --verify --json --select conversation_id,messages
  ```
- **`find`** — Search message content server-side by default, or offline over your synced archive.

  _Pick this to find the exact conversation and message containing a phrase, fresh or from cache._

  ```bash
  chatgpt-pp-cli find "haptics comparison" --select items.title,items.snippet --agent
  ```

### Long-thread navigation
- **`outline`** — Map a long thread into titled sections at each of your prompts, with timestamps and anchors.

  _Use this before reading a 100-message thread — get the structure first, then pull only the section you need._

  ```bash
  chatgpt-pp-cli outline 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --json
  ```
- **`transcript`** — Extract just part of a long conversation by section number or message anchors.

  _Use this to pull one topic out of a sprawling thread without extracting everything._

  ```bash
  chatgpt-pp-cli transcript 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --section 3
  ```

## Recipes

### Month-in-review index

```bash
chatgpt-pp-cli list --since 30d --order created --hydrate --agent
```

Every conversation from the last month with turn counts and time spans in one table.

### Mine one topic from a long thread

```bash
chatgpt-pp-cli transcript 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --section 4 --json --select messages
```

Outline section 4 only — the slice between user anchors 4 and 5.

### Full verified export

```bash
chatgpt-pp-cli transcript 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --verify --format md -o thread.md
```

Complete extraction with a completeness check written as markdown.

### Find the exact message

```bash
chatgpt-pp-cli find "backport deadline" --agent --select items.title,items.snippet,items.payload.message_id
```

Server-side search narrowed to the fields that matter.

### Offline archive search

```bash
chatgpt-pp-cli search "camera comparison" --limit 10
```

Offline FTS over bodies synced by list --hydrate / transcript runs; no session needed.

## Usage

Run `chatgpt-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data: `credentials.toml`, `data.db`, cookies, browser-session proof files, and other auth sidecars |
| `state` | Runtime state such as persisted queries, jobs, and `teach.log` |
| `cache` | Regenerable HTTP/cache files |

Each kind resolves independently. The ladder is:

1. Per-kind env var: `CHATGPT_CONFIG_DIR`, `CHATGPT_DATA_DIR`, `CHATGPT_STATE_DIR`, or `CHATGPT_CACHE_DIR`
2. `--home <dir>` for this invocation
3. `CHATGPT_HOME` for a flat relocated root
4. XDG env vars: `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`
5. Platform defaults matching existing installs

For containers and agent sandboxes, prefer a single relocated root:

```bash
export CHATGPT_HOME=/srv/chatgpt
chatgpt-pp-cli doctor
```

Under `CHATGPT_HOME=/srv/chatgpt`, the four dirs resolve to `/srv/chatgpt/config`, `/srv/chatgpt/data`, `/srv/chatgpt/state`, and `/srv/chatgpt/cache`.

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block:

```json
{
  "mcpServers": {
    "chatgpt": {
      "command": "chatgpt-pp-mcp",
      "env": {
        "CHATGPT_HOME": "/srv/chatgpt"
      }
    }
  }
}
```

Precedence matters in fleets: an ambient per-kind variable such as `CHATGPT_DATA_DIR` overrides an explicit `--home` for that kind. Use `CHATGPT_HOME` or the per-kind variables for durable fleet relocation; treat `--home` as the weaker per-invocation lever.

Relocation is one-way. Unsetting `CHATGPT_HOME` does not move files back to platform defaults, and `doctor` cannot find credentials left under a former root. Move the files manually before unsetting relocation variables.

Existing installs keep working because the platform-default rung matches the legacy layout. On the first auth write, stored secrets leave `config.toml` and are consolidated into `credentials.toml` under the data directory. Run `chatgpt-pp-cli doctor --fail-on warn` to check path and credential-location warnings in automation.

## Commands

### chat

Experimental send surface — browser-gated, no Turnstile solver

- **`chatgpt-pp-cli chat prepare`** - Conduit prepare — mints the short-lived (60s) conduit token required by the send call
- **`chatgpt-pp-cli chat requirements`** - Sentinel gate — returns requirements token, PoW seed/difficulty, turnstile requirement (token values are credential material)
- **`chatgpt-pp-cli chat send`** - Send a message — SSE response with typed events; message_stream_complete is the semantic terminal; resume_conversation_token enables reconnection. EXPERIMENTAL: turnstile-gated; fails clearly without legitimately issued authorization.

### conversations

Conversation inventory and full extraction

- **`chatgpt-pp-cli conversations get`** - Full conversation as the webapp consumes it — messages array plus Relay page_info; has_next_page can be false on truncated windows, verify against current_node
- **`chatgpt-pp-cli conversations get-tree`** - Legacy compat read — singular endpoint returns the parent/children mapping tree (visible roles only; no hidden system/thoughts nodes)
- **`chatgpt-pp-cli conversations list`** - Chronological conversation list — offset/limit pagination; total is a min(realTotal, offset+limit+1) hint, walk pages until a short page
- **`chatgpt-pp-cli conversations list-messages`** - Message window with forward cursor — after=<message_id> returns everything after that id (cursor/limit/end_cursor params are ignored here)
- **`chatgpt-pp-cli conversations search`** - Legacy search — query param is literally 'query' (not q); returns message-level hits with snippets
- **`chatgpt-pp-cli conversations stream-status`** - Streaming state for an active send (404 when idle) — used by status/wait

### global_search

Global search (the Ctrl+K surface)

- **`chatgpt-pp-cli global-search`** - Federated message-content search across conversations and projects — strict body schema (extra fields 422); cursor pagination

### models

Available ChatGPT models

- **`chatgpt-pp-cli models`** - Model picker inventory — slugs, reasoning types, thinking efforts, default badge

### session

Authenticated session state — the bearer mint source

- **`chatgpt-pp-cli session`** - Current session, account, and freshly minted access token (values are credential material — never logged)


### Self-learning loop

This CLI caches per-question discovery so repeat queries skip the walk and structurally similar queries get answered via entity substitution. The loop also self-captures: every invocation is journaled locally, and failed-flag corrections plus fresh teaches surface as candidates on the next `recall` for confirm/reject judgment. Agents call `recall` before discovery and fire `teach &` after answering. See the `## Automatic learning` section in `SKILL.md` for the full protocol.

- **`chatgpt-pp-cli recall <query>`** - Look up cached resources for a query before running discovery
- **`chatgpt-pp-cli teach`** - Record a query -> resource mapping (silent on success, safe to background with `&`)
- **`chatgpt-pp-cli learnings list`** - Inspect taught rows
- **`chatgpt-pp-cli learnings forget <query>`** - Undo a teach
- **`chatgpt-pp-cli learnings candidates`** - List auto-captured candidates awaiting confirm/reject
- **`chatgpt-pp-cli learnings stats`** - Local loop metrics: recall hit rate, teach-to-reuse, playbook resolution, candidate counts
- **`chatgpt-pp-cli teach-pattern`** - Install a query/resource template up front
- **`chatgpt-pp-cli teach-lookup`** - Add an entity mapping (e.g. country code, team alias) for pattern substitution

Pass `--no-learn` or set `CHATGPT_NO_LEARN=true` to disable the loop for deterministic flows.

The local store's schema version stamp is one-way: once this version of `chatgpt-pp-cli` opens the database, older binaries refuse it with a version error — upgrade the binary rather than downgrading.

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
chatgpt-pp-cli conversations list

# JSON for scripting and agents
chatgpt-pp-cli conversations list --json
# Filter to specific fields
chatgpt-pp-cli conversations list --json --select items,total,limit

# Dry run — show the request without sending
chatgpt-pp-cli conversations list --dry-run

# Agent mode — JSON + compact + no prompts in one flag
chatgpt-pp-cli conversations list --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select <field>[,<field>...]` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Explicit retries** - add `--idempotent` to create retries when a no-op success is acceptable
- **Explicit confirmation** - `--agent` does not imply `--yes`; pass `--yes` separately only after the target, arguments, and side effects are clear
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `6` partial failure, `7` rate limited, `10` config error.

## Freshness

This CLI owns bounded freshness for registered store-backed read command paths. In `--data-source auto` mode, covered commands check the local SQLite store before serving results; stale or missing resources trigger a bounded refresh, and refresh failures fall back to the existing local data with a warning. `--data-source local` never refreshes, and `--data-source live` reads the API without mutating the local store.

Set `CHATGPT_NO_AUTO_REFRESH=1` to disable the pre-read freshness hook while preserving the selected data source.

Covered command paths:
- `chatgpt-pp-cli conversations`
- `chatgpt-pp-cli conversations get`
- `chatgpt-pp-cli conversations list`
- `chatgpt-pp-cli conversations search`
- `chatgpt-pp-cli models`
- `chatgpt-pp-cli models get`
- `chatgpt-pp-cli models list`
- `chatgpt-pp-cli models search`

JSON outputs that use the generated provenance envelope include freshness metadata at `meta.freshness`. This metadata describes the freshness decision for the covered command path; it does not claim full historical backfill or API-specific enrichment.

## Health Check

```bash
chatgpt-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Run `chatgpt-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is ``; `--home`, `CHATGPT_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `CHATGPT_TOKEN` | per_call | Yes | Set to your API credential. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `chatgpt-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `chatgpt-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $CHATGPT_TOKEN`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **list returns empty or 404 conversation_inaccessible** — Bearer is missing/expired: run chatgpt-pp-cli auth import again; cookie-only access returns masked empty responses.
- **403 Unusual activity on chat send** — Turnstile-gated: sends are experimental and browser-gated; refresh authorization via chatgpt-pp-cli auth import and retry, or use the web UI.
- **transcript --verify reports missing tail** — Re-run with --full-refresh to cursor-walk the conversation from scratch; the page_info flag can under-report.

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

TLS certificates are verified by default. For a trusted development or self-signed endpoint only, pass `--insecure` for one invocation, set `CHATGPT_SKIP_TLS_VERIFY=true` for the current environment, or set `skip_tls_verify = true` in the config file for a persistent override.

## Discovery Signals

This CLI was generated with browser-captured traffic analysis.
- Target observed: https://chatgpt.com/
- Capture coverage: 12 API entries from 12 total network entries
- Reachability: browser_clearance_http (90% confidence)
- Protocols: sse (95% confidence), rest_json (75% confidence)
- Auth signals: bearer_token — headers: Authorization; cookie — headers: Cookie
- Protection signals: captcha (85% confidence)
- Generation hints: requires_browser_auth, requires_protected_client
- Candidate command ideas: create_chat_requirements — Derived from observed POST /backend-api/sentinel/chat-requirements traffic.; create_conversation — Derived from observed POST /backend-api/conversation traffic.; create_prepare — Derived from observed POST /backend-api/f/conversation/prepare traffic.; create_search — Derived from observed POST /backend-api/global/search traffic.; list_6aaa52c7_daec_83ee_9c8a_e1635dbf397e — Derived from observed GET /backend-api/conversations/6aaa52c7-daec-83ee-9c8a-e1635dbf397e traffic.; list_conversations — Derived from observed GET /backend-api/conversations traffic.; list_messages — Derived from observed GET /backend-api/conversations/6aaa52c7-daec-83ee-9c8a-e1635dbf397e/messages traffic.; list_models — Derived from observed GET /backend-api/models traffic.

Warnings from discovery:
- error_status_cluster: Endpoint cluster only observed error HTTP statuses.
- error_status_cluster: Endpoint cluster only observed error HTTP statuses.
- reserved_resource_name: resource name "auth" may conflict with a reserved Printing Press command or template; consider renaming it to a domain-specific command name

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**chatgpt-exporter**](https://github.com/FdezRomero/chatgpt-exporter) — TypeScript (17 stars)
- [**pi-gpt**](https://github.com/anthropics/pi-gpt-npm) — TypeScript
- [**codex-chats-mcp**](https://github.com/shoyu-ramen/codex-chats-mcp) — Python
- [**chatgpt-conversation-extractor**](https://github.com/2015pulsar/chatgpt-conversation-extractor) — Python

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
