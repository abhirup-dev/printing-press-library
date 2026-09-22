# UnityPredict CLI

**The console-only parts of UnityPredict — model authoring, long descriptions, IO config, cross-env diffing — as CLI commands.**

UnityPredict's engine lifecycle has a Python SDK, but creating models, editing metadata, uploading long descriptions, and configuring inputs/outcomes are console-only. This CLI wraps the full sniffed API surface, including the presigned-S3 flows that break when auth rides along, and adds a dev/prod-aware safety guard.

Learn more at [Unitypredict](https://api.dev.unitypredict.net).

Created by [@dev-abhirup-sc](https://github.com/dev-abhirup-sc) (abhirup).

## Install

The recommended path installs both the `unitypredict-pp-cli` binary and the `pp-unitypredict` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install unitypredict
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install unitypredict --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install unitypredict --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install unitypredict --agent claude-code
npx -y @mvanhorn/printing-press-library install unitypredict --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.6 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/ai/unitypredict/cmd/unitypredict-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/unitypredict-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install unitypredict --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-unitypredict --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-unitypredict --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install unitypredict --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/unitypredict-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `DEV_UNITYPREDICT_TOKEN` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/ai/unitypredict/cmd/unitypredict-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "unitypredict": {
      "command": "unitypredict-pp-mcp",
      "env": {
        "DEV_UNITYPREDICT_TOKEN": "<your-key>"
      }
    }
  }
}
```

</details>

## Authentication

Durable API keys live in ~/.unitypredict/credentials keyed dev/prod; the header needs the literal APIKEY@ prefix (plain keys 401 on strict routes). --env dev|prod selects the tenant; prod mutations refuse without --i-know.

## Quick Start

```bash
# Prove the key works against the one endpoint that cannot lie
unitypredict-pp-cli auth check

# See your models and grab a MODEL id
unitypredict-pp-cli models list-usermodels

# Reverse lookup: which models use this engine
unitypredict-pp-cli engines models <engineId>

```

## Unique Features

These capabilities aren't available in any other tool for this API.

### Diagnostics
- **`logs get`** — One command fetches the per-request devLog, the runtime Log.txt, and the engine build log, in order.

  _When a predict fails you want every log granularity at once, not three manual URL dances that break if you send the auth header to S3._

  ```bash
  unitypredict-pp-cli logs get b5a43cbf-dac1-45a9-a27b-6bbb0e4098f8 --all
  ```

### Model authoring
- **`models describe set`** — Publish a model's long description from a Markdown file in one step, including the S3 upload dance.

  _The long description is a hosted file, not a text field — writing prose into the upsert body is silently ignored, so automation needs the exact 3-step chain._

  ```bash
  unitypredict-pp-cli models describe set c6985b70-15e0-4ba9-8cb1-128200595a8f --file README.md
  ```
- **`models io set`** — Define a model's input and outcome variables from a YAML file instead of the console's per-field editors.

  _The GET-mutate-POST full-record discipline and the hidden-variable display map are only automatable if one command owns the whole round-trip._

  ```bash
  unitypredict-pp-cli models io set c6985b70-15e0-4ba9-8cb1-128200595a8f --from io.yaml
  ```

### Engine lifecycle
- **`engines models`** — List every model built on a given engine.

  _Before redeploying or deleting an engine you must know its blast radius; the console buries this._

  ```bash
  unitypredict-pp-cli engines models b5132be0-a373-4f74-bedf-fabb17dc0342
  ```
- **`engines concurrency`** — Show how many engine nodes ran in parallel over time, as a table.

  _Answers capacity questions (did we actually run 4 in parallel?) without the console._

  ```bash
  unitypredict-pp-cli engines concurrency b5132be0-a373-4f74-bedf-fabb17dc0342
  ```

### Auth and safety
- **`auth check`** — Validate credentials against the one endpoint that cannot lie about auth.

  _Prevents shipping on a wrong-env or malformed key that appeared to work because a list route answered 200._

  ```bash
  unitypredict-pp-cli auth check --env prod
  ```

### Invocation
- **`predict run`** — Invoke a model with text or file inputs, poll to completion, and pull File outcomes into a directory.

  _Cold-started GPU models take minutes; a single wait-and-fetch command is the difference between a script and a one-liner._

  ```bash
  unitypredict-pp-cli predict run ba375bf0-f794-4cbc-9b1a-2ec1a6967d67 --input TargetLanguage=Hindi --wait
  ```

### Promotion
- **`diff model`** — Compare a model or engine record across dev and prod, field by field.

  _Before promoting, you want to know exactly which fields drifted between dev and prod._

  ```bash
  unitypredict-pp-cli diff model c6985b70-15e0-4ba9-8cb1-128200595a8f
  ```

## Recipes

### Full request autopsy

```bash
unitypredict-pp-cli logs get <requestId> --all
```

Fans out devLog, runtime Log.txt, and engine build log for a failed predict.

### Author a model end to end

```bash
unitypredict-pp-cli models io set <modelId> --from io.yaml
```

Declarative inputs/outcomes with the full-record upsert handled for you.

### Promotion check

```bash
unitypredict-pp-cli diff model <modelId>
```

Field-by-field drift between dev and prod tenants.

## Usage

Run `unitypredict-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data: `credentials.toml`, `data.db`, cookies, browser-session proof files, and other auth sidecars |
| `state` | Runtime state such as persisted queries, jobs, and `teach.log` |
| `cache` | Regenerable HTTP/cache files |

Each kind resolves independently. The ladder is:

1. Per-kind env var: `UNITYPREDICT_CONFIG_DIR`, `UNITYPREDICT_DATA_DIR`, `UNITYPREDICT_STATE_DIR`, or `UNITYPREDICT_CACHE_DIR`
2. `--home <dir>` for this invocation
3. `UNITYPREDICT_HOME` for a flat relocated root
4. XDG env vars: `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`
5. Platform defaults matching existing installs

For containers and agent sandboxes, prefer a single relocated root:

```bash
export UNITYPREDICT_HOME=/srv/unitypredict
unitypredict-pp-cli doctor
```

Under `UNITYPREDICT_HOME=/srv/unitypredict`, the four dirs resolve to `/srv/unitypredict/config`, `/srv/unitypredict/data`, `/srv/unitypredict/state`, and `/srv/unitypredict/cache`.

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block:

```json
{
  "mcpServers": {
    "unitypredict": {
      "command": "unitypredict-pp-mcp",
      "env": {
        "UNITYPREDICT_HOME": "/srv/unitypredict"
      }
    }
  }
}
```

Precedence matters in fleets: an ambient per-kind variable such as `UNITYPREDICT_DATA_DIR` overrides an explicit `--home` for that kind. Use `UNITYPREDICT_HOME` or the per-kind variables for durable fleet relocation; treat `--home` as the weaker per-invocation lever.

Relocation is one-way. Unsetting `UNITYPREDICT_HOME` does not move files back to platform defaults, and `doctor` cannot find credentials left under a former root. Move the files manually before unsetting relocation variables.

Existing installs keep working because the platform-default rung matches the legacy layout. On the first auth write, stored secrets leave `config.toml` and are consolidated into `credentials.toml` under the data directory. Run `unitypredict-pp-cli doctor --fail-on warn` to check path and credential-location warnings in automation.

## Commands

### auth_resource

Operations on auth

- **`unitypredict-pp-cli auth-resource`** - GET /api/auth

### clients

Operations on Console

- **`unitypredict-pp-cli clients list-console`** - GET /api/clients/Console
- **`unitypredict-pp-cli clients list-router`** - GET /api/clients/router/

### engines

Operations on userengines

- **`unitypredict-pp-cli engines create-engines`** - POST /api/engines
- **`unitypredict-pp-cli engines get-source-file`** - GET /api/engines/download/{download_id}/SourceFile
- **`unitypredict-pp-cli engines get-buildlogs`** - GET /api/engines/buildlogs/{buildlog_id}
- **`unitypredict-pp-cli engines get-engines`** - GET /api/engines/{engine_id}
- **`unitypredict-pp-cli engines list-supportedengines`** - GET /api/engines/supportedengines
- **`unitypredict-pp-cli engines list-userengines`** - GET /api/engines/userengines

### fast

Operations on THUMBNAIL.png

- **`unitypredict-pp-cli fast <model_id>`** - GET /api/fast/models/{model_id}/files/THUMBNAIL.png

### models

Operations on search

- **`unitypredict-pp-cli models create-models`** - POST /api/models
- **`unitypredict-pp-cli models get-modeldescription`** - GET /api/models/{model_id}/filekey/MODELDESCRIPTION
- **`unitypredict-pp-cli models get-modeldescription-2`** - GET /api/models/upload/{upload_id}/MODELDESCRIPTION
- **`unitypredict-pp-cli models get-thumbnail`** - GET /api/models/{model_id}/filekey/THUMBNAIL
- **`unitypredict-pp-cli models get-models`** - GET /api/models/{model_id}
- **`unitypredict-pp-cli models list-search`** - GET /api/models/search
- **`unitypredict-pp-cli models list-usermodels`** - GET /api/models/usermodels

### payments

Operations on accounts

- **`unitypredict-pp-cli payments`** - GET /api/payments/accounts

### predict

Operations on predict

- **`unitypredict-pp-cli predict create-predict`** - POST /api/predict/{predict_id}
- **`unitypredict-pp-cli predict get-logs`** - GET /api/predict/status/{statu_id}/logs
- **`unitypredict-pp-cli predict get-status`** - GET /api/predict/status/{statu_id}

### repository

Operations on search

- **`unitypredict-pp-cli repository`** - GET /api/repository/search

### social

Operations on review

- **`unitypredict-pp-cli social <review_id>`** - GET /api/social/model/review/{review_id}

### tools

Operations on analytics

- **`unitypredict-pp-cli tools list-analytics`** - GET /api/tools/analytics
- **`unitypredict-pp-cli tools list-dev.abhirupdas-8c0af724114435`** - GET /api/tools/public/dev.abhirupdas_8C0AF724114435


### Self-learning loop

This CLI caches per-question discovery so repeat queries skip the walk and structurally similar queries get answered via entity substitution. The loop also self-captures: every invocation is journaled locally, and failed-flag corrections plus fresh teaches surface as candidates on the next `recall` for confirm/reject judgment. Agents call `recall` before discovery and fire `teach &` after answering. See the `## Automatic learning` section in `SKILL.md` for the full protocol.

- **`unitypredict-pp-cli recall <query>`** - Look up cached resources for a query before running discovery
- **`unitypredict-pp-cli teach`** - Record a query -> resource mapping (silent on success, safe to background with `&`)
- **`unitypredict-pp-cli learnings list`** - Inspect taught rows
- **`unitypredict-pp-cli learnings forget <query>`** - Undo a teach
- **`unitypredict-pp-cli learnings candidates`** - List auto-captured candidates awaiting confirm/reject
- **`unitypredict-pp-cli learnings stats`** - Local loop metrics: recall hit rate, teach-to-reuse, playbook resolution, candidate counts
- **`unitypredict-pp-cli teach-pattern`** - Install a query/resource template up front
- **`unitypredict-pp-cli teach-lookup`** - Add an entity mapping (e.g. country code, team alias) for pattern substitution

Pass `--no-learn` or set `UNITYPREDICT_NO_LEARN=true` to disable the loop for deterministic flows.

The local store's schema version stamp is one-way: once this version of `unitypredict-pp-cli` opens the database, older binaries refuse it with a version error — upgrade the binary rather than downgrading.

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
unitypredict-pp-cli auth-resource

# JSON for scripting and agents
unitypredict-pp-cli auth-resource --json
# Filter to specific fields by name
unitypredict-pp-cli auth-resource --json --select <field>[,<field>...]

# Dry run — show the request without sending
unitypredict-pp-cli auth-resource --dry-run

# Agent mode — JSON + compact + no prompts in one flag
unitypredict-pp-cli auth-resource --agent
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

## Health Check

```bash
unitypredict-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Run `unitypredict-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is `~/.config/unitypredict-pp-cli/config.toml`; `--home`, `UNITYPREDICT_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `DEV_UNITYPREDICT_TOKEN` | per_call | Yes | Set to your API credential. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `unitypredict-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `unitypredict-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $DEV_UNITYPREDICT_TOKEN`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **401 "You cannot use this model without logging in!"** — Auth header missing the literal APIKEY@ prefix — the CLI adds it; check the key in ~/.unitypredict/credentials
- **404 empty body on predict** — You used an ENGINE id; predict needs the MODEL id (console /models/<id>/view)
- **400 from an S3 log/upload URL** — Presigned URLs must be fetched with no Authorization header — the CLI drops it on cross-host redirects
- **Empty result list on prod** — Wrong-env key returns empty results, not an error; run auth check --env prod

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

TLS certificates are verified by default. For a trusted development or self-signed endpoint only, pass `--insecure` for one invocation, set `UNITYPREDICT_SKIP_TLS_VERIFY=true` for the current environment, or set `skip_tls_verify = true` in the config file for a persistent override.

## Discovery Signals

This CLI was generated with browser-captured traffic analysis.
- Target observed: https://api.dev.unitypredict.net/api/clients/Console
- Capture coverage: 43 API entries from 43 total network entries
- Reachability: standard_http (65% confidence)
- Protocols: rest_json (75% confidence)
- Auth signals: bearer_token — headers: Authorization
- Candidate command ideas: create_engines — Derived from observed POST /api/engines traffic.; create_models — Derived from observed POST /api/models traffic.; create_predict — Derived from observed POST /api/predict/{predict_id} traffic.; get_MODELDESCRIPTION — Derived from observed GET /api/models/upload/{upload_id}/MODELDESCRIPTION traffic.; get_SourceFile — Derived from observed GET /api/engines/download/{download_id}/SourceFile traffic.; get_THUMBNAIL — Derived from observed GET /api/models/{model_id}/filekey/THUMBNAIL traffic.; get_THUMBNAIL.png — Derived from observed GET /api/fast/models/{model_id}/files/THUMBNAIL.png traffic.; get_buildlogs — Derived from observed GET /api/engines/buildlogs/{buildlog_id} traffic.

Warnings from discovery:
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.
- empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
