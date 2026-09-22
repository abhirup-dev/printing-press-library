---
name: pp-chatgpt
description: "Your ChatGPT history as a queryable local archive — verified-complete extraction, outlines, and search no other tool has. Trigger phrases: `list my chatgpt conversations`, `export a chatgpt thread`, `chatgpt conversation outline`, `search my chatgpt history`, `use chatgpt`, `run chatgpt`."
author: "dev-abhirup-sc"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - chatgpt-pp-cli
    install:
      - kind: go
        bins: [chatgpt-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/ai/chatgpt/cmd/chatgpt-pp-cli
---

# ChatGPT — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `chatgpt-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install chatgpt --cli-only
   ```
2. Verify: `chatgpt-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.6 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/ai/chatgpt/cmd/chatgpt-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

Turn conversations into a chronological index with real turn counts and spans, extract any thread completely with a completeness proof, navigate long threads by outline, and search message content server-side or offline. Sessions survive token expiry via cookie re-mint and optional Codex auth import.

## When to Use This CLI

Use this CLI when you need chronological inventory of conversations with real metadata, verified-complete single-thread extraction (JSON or readable transcript), long-thread outlines and section slicing, or message-content search (server-fresh or offline). Ideal for archiving, auditing, feeding conversation context to other agents, and navigating very long threads. Use 'use chatgpt' or 'run chatgpt' for these workflows.

## Anti-triggers

Do not use this CLI for:
- Do not use this CLI to delete, rename, or archive conversations — destructive management is deliberately excluded
- Do not use it for bulk attachment or project backups — dedicated exporters cover that
- Do not use it for deep-research or agent-mode chat parity — use the web UI or pi-gpt
- Do not expect headless send reliability — sends are experimental and browser-gated

## Unique Capabilities

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

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

## Discovery Signals

This CLI was generated with browser-observed traffic context.
- Capture coverage: 12 API entries from 12 total network entries
- Protocols: sse (95% confidence), rest_json (75% confidence)
- Auth signals: bearer_token — headers: Authorization; cookie — headers: Cookie
- Generation hints: requires_browser_auth, requires_protected_client
- Candidate command ideas: create_chat_requirements — Derived from observed POST /backend-api/sentinel/chat-requirements traffic.; create_conversation — Derived from observed POST /backend-api/conversation traffic.; create_prepare — Derived from observed POST /backend-api/f/conversation/prepare traffic.; create_search — Derived from observed POST /backend-api/global/search traffic.; list_6aaa52c7_daec_83ee_9c8a_e1635dbf397e — Derived from observed GET /backend-api/conversations/6aaa52c7-daec-83ee-9c8a-e1635dbf397e traffic.; list_conversations — Derived from observed GET /backend-api/conversations traffic.; list_messages — Derived from observed GET /backend-api/conversations/6aaa52c7-daec-83ee-9c8a-e1635dbf397e/messages traffic.; list_models — Derived from observed GET /backend-api/models traffic.
- Caveats: error_status_cluster: Endpoint cluster only observed error HTTP statuses.; error_status_cluster: Endpoint cluster only observed error HTTP statuses.; reserved_resource_name: resource name "auth" may conflict with a reserved Printing Press command or template; consider renaming it to a domain-specific command name

## Command Reference

**chat** — Experimental send surface — browser-gated, no Turnstile solver

- `chatgpt-pp-cli chat prepare` — Conduit prepare — mints the short-lived (60s) conduit token required by the send call
- `chatgpt-pp-cli chat requirements` — Sentinel gate — returns requirements token, PoW seed/difficulty
- `chatgpt-pp-cli chat ask` — EXPERIMENTAL: send a message (new chat or --conversation-id follow-up) with full protocol; message_stream_complete is the semantic terminal; fails clearly when Turnstile-gated
- `chatgpt-pp-cli chat status` — EXPERIMENTAL: report a conversation's streaming state
- `chatgpt-pp-cli chat send` — raw endpoint mirror of the send call (typed body flags)

**conversations** — Conversation inventory and full extraction

- `chatgpt-pp-cli conversations get` — Full conversation as the webapp consumes it — messages array plus Relay page_info
- `chatgpt-pp-cli conversations get-tree` — Legacy compat read — singular endpoint returns the parent/children mapping tree (visible roles only
- `chatgpt-pp-cli conversations list` — Chronological conversation list — offset/limit pagination; total is a min(realTotal, offset+limit+1) hint
- `chatgpt-pp-cli conversations list-messages` — Message window with forward cursor — after=<message_id> returns everything after that id (cursor/limit/end_cursor
- `chatgpt-pp-cli conversations search` — Legacy search — query param is literally 'query' (not q); returns message-level hits with snippets
- `chatgpt-pp-cli conversations stream-status` — Streaming state for an active send (404 when idle) — used by status/wait

**global_search** — Global search (the Ctrl+K surface)

- `chatgpt-pp-cli global-search` — Federated message-content search across conversations and projects — strict body schema (extra fields 422)

**models** — Available ChatGPT models

- `chatgpt-pp-cli models` — Model picker inventory — slugs, reasoning types, thinking efforts, default badge

**session** — Authenticated session state — the bearer mint source

- `chatgpt-pp-cli session` — Current session, account, and freshly minted access token (values are credential material — never logged)


## Freshness Contract

This printed CLI owns bounded freshness only for registered store-backed read command paths. In `--data-source auto` mode, those paths check `sync_state` and may run a bounded refresh before reading local data. `--data-source local` never refreshes. `--data-source live` reads the API and does not mutate the local store. Set `CHATGPT_NO_AUTO_REFRESH=1` to skip the freshness hook without changing source selection.

Covered paths:

- `chatgpt-pp-cli conversations`
- `chatgpt-pp-cli conversations get`
- `chatgpt-pp-cli conversations list`
- `chatgpt-pp-cli conversations search`
- `chatgpt-pp-cli models`
- `chatgpt-pp-cli models get`
- `chatgpt-pp-cli models list`
- `chatgpt-pp-cli models search`

When JSON output uses the generated provenance envelope, freshness metadata appears at `meta.freshness`. Treat it as current-cache freshness for the covered command path, not a guarantee of complete historical backfill or API-specific enrichment.

### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
chatgpt-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query. `--json` (and other machine formats) keep that exit-2 contract and write `{"matches":[]}` on stdout so agents can inspect the envelope without treating a miss as success.

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

## Auth Setup

No API key. Import your chatgpt.com browser session once (cookies) and the CLI mints fresh bearer tokens itself — cookies outlive the 10-day token TTL, so the session keeps working. Alternatively point it at your existing codex login (~/.codex/auth.json). Send commands are experimental and browser-gated: they consume only legitimately issued authorization and fail with a clear error otherwise.

Run `chatgpt-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color`.

Global format flags share one contract on promoted, novel, sync, and `--deliver` paths:

- `--json` — one JSON document on stdout (sync progress events go to stderr)
- `--compact` — keep identity/status/timestamp fields; does not change the document vs stream shape
- `--csv` / `--plain` — tabular rows (collection envelopes unwrap to the row array)
- `--quiet` — one identity value per row, no envelope

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  chatgpt-pp-cli conversations list --agent --select items,total,limit
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Explicit confirmation** — `--agent` does not imply `--yes`; pass `--yes` separately only after the target, arguments, and side effects are clear
- **Explicit retries** — use `--idempotent` only when an already-existing create should count as success

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal AND no machine-format flag (`--json`, `--csv`, `--compact`, `--quiet`, `--plain`, `--select`) is set — piped/agent consumers and explicit-format runs get pure JSON on stdout.

## Paths and state

Agents should treat the CLI's path resolver as part of the runtime contract:

- Use `--home <dir>` for one invocation, or set `CHATGPT_HOME=<dir>` to relocate all four path kinds under one root.
- Use per-kind env vars only when a specific kind must diverge: `CHATGPT_CONFIG_DIR`, `CHATGPT_DATA_DIR`, `CHATGPT_STATE_DIR`, `CHATGPT_CACHE_DIR`.
- Resolution order is per-kind env var, `--home`, `CHATGPT_HOME`, XDG (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`), then platform defaults.
- `config` contains settings like `config.toml` and profiles. `data` contains `credentials.toml`, `data.db`, cookies, and auth sidecars. `state` contains persisted queries, jobs, and `teach.log`. `cache` contains regenerable HTTP/cache files.
- Stored secrets live in `credentials.toml` under the data dir. Existing legacy `config.toml` secrets are read for compatibility and leave `config.toml` on the first auth write.
- Run `chatgpt-pp-cli doctor --fail-on warn` to surface path and credential-location warnings. `agent-context` exposes a schema v4 `paths` block for agents that need the resolved dirs.
- For MCP, pass relocation through the MCP host config. The MCP binary does not inherit CLI flags:

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

Fleet precedence: an inherited per-kind env var overrides an explicit `--home` for that kind. Use `CHATGPT_HOME` or per-kind vars as durable fleet levers, and use `--home` only for a single invocation. Relocation is not reversible by unsetting env vars; move files manually before clearing `CHATGPT_HOME`, or `doctor` will not find credentials left under the former root.

## Automatic learning

This CLI ships a self-capturing learning loop. The CLI does its own bookkeeping: every invocation is journaled locally, a failed flag followed by a corrected retry auto-derives a `flag_alias` candidate, and a `teach` on a query family without a playbook auto-synthesizes a `playbook_candidate` from the session's journal. Your job is judgment only: `recall` first, act on surfaced candidates, `teach` the final answer, `playbook amend` when you observe a correction. You never record failures by hand.

### Step 1: `recall` before any discovery

Before list/search/drill commands on a new user question, pass the question as an argv or MCP tool argument to `recall --agent`. Do not interpolate user-controlled text into a shell command line.

Quoted `recall "<question>"` breaks on an apostrophe, which is ordinary English. A quoted heredoc breaks when a body line equals the delimiter, and that delimiter is published in these docs. Write the question with a non-shell file-writing tool, then read it back as data:

```bash
# Write the question verbatim with your file-writing tool (no shell involved).
# Command substitution on a file only ever yields data — the shell never
# parses the file's bytes as syntax.
QUERY=$(cat /path/to/question.txt)
chatgpt-pp-cli recall "$QUERY" --agent
```

Prefer MCP: pass the question as the tool's query argument. `"$QUERY"` after a file read is argv-safe; putting the question itself in the command text is not.

The response envelope:

```json
{
  "query": "...",
  "normalized": "<normalized form>",
  "query_entities": ["..."],
  "found": true | false,
  "match_score": 0.0,
  "results": [
    { "resource_id": "...", "resource_type": "...", "venue": "...",
      "confidence": 2, "entity_match": "exact|partial|unknown",
      "source": "taught|preseed|pattern", "warnings": ["..."] }
  ],
  "mismatches": [ /* only when --debug-mismatches */ ],
  "warnings": [ /* top-level */ ],
  "candidates": [
    { "id": 12, "class": "flag_alias | playbook_candidate",
      "summary": "...", "sightings": 3, "last_seen": "...",
      "rationale": "...",
      "next_action": ["<trial command>", "chatgpt-pp-cli learnings confirm 12"] }
  ],
  "playbook": {
    "query_family": "...",
    "playbook": {
      "steps": [ { "cmd": "<command with {slot} substitution>", "purpose": "..." } ],
      "entity_slots": ["$ENTITY"],
      "expected_tool_calls": 3
    },
    "slots_resolved": { "$ENTITY": { "token": "<live token>", "canonical": "<canonical>" } },
    "notes": "<workarounds + gotchas for this query family>"
  },
  "notes": "<duplicate surface for non-playbook callers>"
}
```

Empty-store short-circuit: if the store has no learnings, playbooks, or candidates yet (recall finds nothing and `learnings list` and `learnings candidates` are both empty), skip recall for the rest of this session instead of taxing every query; resume recall-first once something has been taught.

### Step 2: decision tree

Read `candidates`, `playbook`, `notes`, `results[0]`, and warnings in that order:

```
if Candidates present (warnings include "candidates_present"):
    -> candidates are try-then-confirm, never facts. Follow each candidate's
       two-step next_action verbatim: run the trial command first, then run
       `learnings confirm <id>` only after the trial verified the behavior.
       Reject a wrong candidate with `learnings reject <id>`.
    -> NEVER re-teach something recall surfaced as a candidate; confirm or
       reject that candidate instead of teaching a duplicate.
    -> candidates ride alongside playbooks and resource hits, not instead of
       them; continue with the branches below after acting on them.

if Playbook present:
    -> READ Playbook.notes verbatim FIRST (workarounds + gotchas the CLI surface doesn't expose)
    -> replay Playbook.steps in order, substituting Playbook.slots_resolved entries
       for the entity slot tokens. If a step's slot is unresolved, fall back to
       discovery for that step only.
    -> the Playbook's expected_tool_calls is a budget; if you find yourself running
       materially more, record the divergence via `chatgpt-pp-cli playbook amend`
       at end-of-session.

elif Notes present (no Playbook):
    -> read Notes verbatim before any discovery step; they carry known gotchas
       for this query family even when no structured choreography exists yet.

elif Found AND Results[0].EntityMatch == "exact" AND Results[0].Confidence >= 2:
    -> skip discovery; fetch live data for Results[*].ResourceID in parallel

elif Found AND Results[0].EntityMatch == "partial":
    -> candidate hint, NOT a hit; read the resource title to validate before trusting

elif (any row in Mismatches[] when --debug-mismatches was passed):
    -> treat as cold start; the stored learning is for a different entity
       (different canonical resolved from query_entities)

else:  // Found == false, no playbook, no notes
    -> cold start; run discovery normally; teach the answer afterward (Step 4).
       If the family has no playbook yet, that teach auto-synthesizes a
       playbook candidate from this session's journal - you do not need to
       record one by hand.
```

Playbook and Notes are orthogonal to the per-resource path. A recall response can carry both a Playbook AND a `Results[]` hit - use both: the Playbook tells you which choreography to run; the resource hits short-circuit specific steps. Default to skipping `mismatches`; pass `--debug-mismatches` only when investigating cold-start surprises.

Candidate judgment details: `learnings confirm <id>` prints the candidate's full payload before materializing it - check that the printed payload matches the behavior you verified. `learnings reject <id>` tombstones the derivation signature so the same candidate does not resurface. The envelope carries only the few candidates worth acting on now; `chatgpt-pp-cli learnings candidates` lists the full open set.

Graceful degradation: if `learnings confirm` is an unknown command, you are driving an older binary - ignore the candidates guidance and follow the rest of the protocol.

### Step 3: always read `warnings`

- `low_confidence`: row exists at `confidence<2`. Treat as a hint, not a skip-discovery hit.
- `resource_not_in_store`: the local store doesn't have the resource the learning points at. The match validator couldn't classify entities — direct-fetch and re-evaluate.
- `cross_alias_match` (per-result): the row was taught under a different alias and matched the live query's canonical via `entity_lookups` (e.g., a "USA" teach satisfying a "United States" recall). Trust the resource_id.
- `similar_shape_different_entity:<canonical>` (top-level): a structurally matching row exists but its canonical entity differs from the live query's. Treated as cold start; the warning carries the conflicting canonical as a hint, but the row is NOT promoted into Results.
- `ambiguous_alias` (top-level): a single query entity resolved to multiple canonicals (e.g., "Cards" → Arizona Cardinals + St. Louis Cardinals). Surface the ambiguity from context before committing to a resource.
- `candidates_present` (top-level): the envelope carries a `candidates` section. Handle it via the candidates branch in Step 2 before anything else.
- `lookup_refresh_available` (top-level): an entity in the query has no lookup row yet, but synced data could provide one. Run `chatgpt-pp-cli sync --resources conversations,models` to refresh entity lookups.
- Top-level `no_learnings_for_query_family`: the table had no rows above the Jaccard floor. Pure cold start.

### Step 4: `teach &` after finalizing your response - always

Teaching is unconditional. After resolving a query the store could not answer, background-teach the final resource mapping - no call-count threshold, no judging whether it was "worth" learning. The teach is the anchor of the loop: it triggers playbook synthesis for a family without a playbook, and same-referent phrasings fold into one family so near-duplicate teaches do not fragment the store. Fire it after assembling your user-facing response but BEFORE emitting it, with a shell `&` so the call returns immediately. Pass the query the same way as recall — argv/MCP, or file-then-`$QUERY`. Do not splice the question into the command text:

```bash
QUERY=$(cat /path/to/question.txt)
chatgpt-pp-cli teach --query "$QUERY" --resource-type <type> --resource <id1> --resource <id2>
# (append shell `&` to background it)
```

Silent on success. Errors only land in `teach.log` under the resolved state dir. Teach the **most specific** resource - if the user asked a broad question and you walked through parent records to find the specific answer, teach the leaf id, not the parent. The CLI uses seeded `entity_lookups` for cross-alias resolution at recall time, so a teach under one alias (e.g., "Niners") satisfies future queries under another alias (e.g., "49ers", "San Francisco") automatically.

PII rule: teach the structural question with identifiers stripped - never include names, emails, phone numbers, account ids, or other personal identifiers in taught queries or notes. The CLI scans teach queries for obvious email/phone shapes and warns, but does not block; strip before teaching rather than relying on the warning.

### Step 5: playbooks - optional flags, automatic synthesis

You do not need to decide whether a session "deserves" a playbook: a teach on a family without one auto-synthesizes a `playbook_candidate` from the session's journal, and the next session judges it via confirm/reject. Attach explicit playbook flags only when you already hold choreography worth recording verbatim - workarounds the CLI didn't surface (silently-dropped flags, undocumented params, pagination tricks, payload gotchas). Prefer the **integrated one-call form** - record the resource learning and the playbook in the same `teach` invocation:

```bash
# Common case: record both the resource learning AND the playbook in one call.
QUERY=$(cat /path/to/question.txt)
chatgpt-pp-cli teach \
  --query "$QUERY" \
  --resource <id> \
  --playbook-file ~/playbooks/<shape>.json \
  --playbook-notes-file ~/playbooks/<shape>-notes.md
# (append shell `&` to background it)

# Alternate: playbook-only (no resource to record alongside).
QUERY=$(cat /path/to/question.txt)
chatgpt-pp-cli teach-playbook \
  --query "$QUERY" \
  --playbook-file ~/playbooks/<shape>.json \
  --notes-file ~/playbooks/<shape>-notes.md
```

Playbook files are JSON with `steps`, `entity_slots`, `expected_tool_calls`. Notes files are markdown carrying the gotchas verbatim. File-free callers (MCP-only agents) pass the same content inline: `--playbook-json` and `--playbook-notes` on the integrated `teach` form, `--playbook-json` and `--notes` on `teach-playbook`. On the integrated `teach` form, the playbook flags are optional - omit them entirely for a resource-only teach. On the standalone `teach-playbook` form, at least one of the playbook and notes flags must be set; both empty is rejected. Playbooks are keyed on the structural query family (entities stripped) so a recipe taught from one entity-shaped query applies to every other query of the same shape, with `slots_resolved` binding the live query's canonical at recall time.

When you DO find a playbook on a future recall, treat it as ground truth: replay the steps with `slots_resolved` substitutions, skip the discovery that the choreography already documents, and read `notes` before any step.

### Step 6: `playbook amend &` when your debug response identifies a correction

If your debug-protocol response identifies a concrete correction the notes or playbook should know — a workaround, an undocumented endpoint shape, a stale field name, observed schema drift, an empty-payload fallback — fire `playbook amend` BEFORE emitting your user-facing response. Same fire-and-forget posture as `teach`. Pass the query and note as argv/MCP arguments, or write each with a non-shell file tool and read them back (`QUERY=$(cat ...)`, `NOTE=$(cat ...)`). Do not interpolate either string into the command text:

```bash
QUERY=$(cat /path/to/question.txt)
NOTE=$(cat /path/to/note.txt)
chatgpt-pp-cli playbook amend \
  --query "$QUERY" \
  --add-note "$NOTE"
# (append shell `&` to background it)
```

What counts as worth amending: a behavior you OBSERVED this session that future-you would benefit from knowing. Examples worth amending:

- A workaround for a CLI surface that silently drops or misorders a flag.
- An undocumented endpoint shape (response wrapped in `{meta, results}`, payload nested two levels deeper than the docs claim).
- Observed schema drift (a field renamed, an index that shifted between seasons, a category label that the API now returns lower-cased).

What does NOT belong in notes:

- The year-specific or entity-specific answer to the user's question. That's the response, not a learning.
- Per-team / per-athlete / per-row data the playbook already retrieves at runtime.
- Statements that paraphrase what the existing notes already say.

The amend command appends to the family's existing notes with a timestamped marker (`[amend YYYY-MM-DDTHH:MMZ]: <text>`). Multiple amends accumulate; the audit trail is visible. If no playbook exists yet for the family, amend creates a notes-only one (so cold-start corrections still land).

#### PII discipline for amend notes

`playbook amend` notes are designed to potentially flow upstream as shared knowledge in future versions of the Printing Press. Keep them clean of user-identifying content so the upstream-contribution path stays open without retroactive scrubbing:

- **Do NOT embed** paths to user filesystems, personal API keys or tokens, user email addresses, user GitHub handles, or specific query histories tied to a single user.
- **Acceptable**: endpoint shapes, undocumented field names, API gotchas, observed schema drift, workarounds for CLI surfaces, generalizable pagination or retry tactics.

If a correction is only meaningful with user-specific context, it belongs in a personal note, not in the playbook amend.

### Measuring the loop

`chatgpt-pp-cli learnings stats` reports recall hit rate, teach-to-reuse, playbook resolution rate, and candidate confirm/reject counts from the local `learn_events` table. Rates are null until they have a denominator; everything stays on this machine. Use it to check whether the loop is earning its keep for this CLI.

### Disabling learning

- `--no-learn` on a single command short-circuits both `recall` and the `teach` write path. Use for deterministic agent flows or tests that must not be affected by accumulated learnings.
- `CHATGPT_NO_LEARN=true` in the environment globally disables the pipeline.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
chatgpt-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
chatgpt-pp-cli feedback --stdin < notes.txt
chatgpt-pp-cli feedback list --json --limit 10
```

Entries are stored locally as `feedback.jsonl` under the resolved data dir. They are never POSTed unless `CHATGPT_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `CHATGPT_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename). Binary-response commands write decoded payload bytes (not the base64 JSON envelope) and print a small JSON receipt on stdout; `--json`/`--csv` do not refuse when this sink is set. |
| `webhook:<url>` | POST the output body to the URL (`application/json`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled or recurring agent reuses the same saved flags while providing different input each run.

```
chatgpt-pp-cli profile save briefing --json
chatgpt-pp-cli --profile briefing conversations list
chatgpt-pp-cli profile list --json
chatgpt-pp-cli profile show briefing
chatgpt-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 6 | Partial failure |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `chatgpt-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/ai/chatgpt/cmd/chatgpt-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add chatgpt-pp-mcp -- chatgpt-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which chatgpt-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   chatgpt-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `chatgpt-pp-cli <command> --help`.
