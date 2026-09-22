---
name: pp-unitypredict
description: "UnityPredict CLI: invoke models (predict run --wait), edit model metadata/long descriptions/inputs-outputs, reverse-lookup engines, fetch devLogs and build logs, check auth honestly (auth check), diff dev vs prod. Use when the user mentions UnityPredict models, engines, predicts, devlogs, or dev/prod promotion."
author: "abhirup"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - unitypredict-pp-cli
    install:
      - kind: go
        bins: [unitypredict-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/ai/unitypredict/cmd/unitypredict-pp-cli
---

# Unitypredict — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `unitypredict-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install unitypredict --cli-only
   ```
2. Verify: `unitypredict-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.6 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/ai/unitypredict/cmd/unitypredict-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

UnityPredict's engine lifecycle has a Python SDK, but creating models, editing metadata, uploading long descriptions, and configuring inputs/outcomes are console-only. This CLI wraps the full sniffed API surface, including the presigned-S3 flows that break when auth rides along, and adds a dev/prod-aware safety guard.

## Unique Capabilities

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

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

## Discovery Signals

This CLI was generated with browser-observed traffic context.
- Capture coverage: 43 API entries from 43 total network entries
- Protocols: rest_json (75% confidence)
- Auth signals: bearer_token — headers: Authorization
- Candidate command ideas: create_engines — Derived from observed POST /api/engines traffic.; create_models — Derived from observed POST /api/models traffic.; create_predict — Derived from observed POST /api/predict/{predict_id} traffic.; get_MODELDESCRIPTION — Derived from observed GET /api/models/upload/{upload_id}/MODELDESCRIPTION traffic.; get_SourceFile — Derived from observed GET /api/engines/download/{download_id}/SourceFile traffic.; get_THUMBNAIL — Derived from observed GET /api/models/{model_id}/filekey/THUMBNAIL traffic.; get_THUMBNAIL.png — Derived from observed GET /api/fast/models/{model_id}/files/THUMBNAIL.png traffic.; get_buildlogs — Derived from observed GET /api/engines/buildlogs/{buildlog_id} traffic.
- Caveats: empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.; empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.; empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.; empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.; empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.; empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.; empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.; empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.; empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.; empty_payload: API-looking request returned an empty or null payload; schema confidence is weak.

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

## Command Reference

**auth_resource** — Operations on auth

- `unitypredict-pp-cli auth-resource` — GET /api/auth

**clients** — Operations on Console

- `unitypredict-pp-cli clients list-console` — GET /api/clients/Console
- `unitypredict-pp-cli clients list-router` — GET /api/clients/router/

**engines** — Operations on userengines

- `unitypredict-pp-cli engines create-engines` — POST /api/engines
- `unitypredict-pp-cli engines get-source-file` — GET /api/engines/download/{download_id}/SourceFile
- `unitypredict-pp-cli engines get-buildlogs` — GET /api/engines/buildlogs/{buildlog_id}
- `unitypredict-pp-cli engines get-engines` — GET /api/engines/{engine_id}
- `unitypredict-pp-cli engines list-supportedengines` — GET /api/engines/supportedengines
- `unitypredict-pp-cli engines list-userengines` — GET /api/engines/userengines

**fast** — Operations on THUMBNAIL.png

- `unitypredict-pp-cli fast <model_id>` — GET /api/fast/models/{model_id}/files/THUMBNAIL.png

**models** — Operations on search

- `unitypredict-pp-cli models create-models` — POST /api/models
- `unitypredict-pp-cli models get-modeldescription` — GET /api/models/{model_id}/filekey/MODELDESCRIPTION
- `unitypredict-pp-cli models get-modeldescription-2` — GET /api/models/upload/{upload_id}/MODELDESCRIPTION
- `unitypredict-pp-cli models get-thumbnail` — GET /api/models/{model_id}/filekey/THUMBNAIL
- `unitypredict-pp-cli models get-models` — GET /api/models/{model_id}
- `unitypredict-pp-cli models list-search` — GET /api/models/search
- `unitypredict-pp-cli models list-usermodels` — GET /api/models/usermodels

**payments** — Operations on accounts

- `unitypredict-pp-cli payments` — GET /api/payments/accounts

**predict** — Operations on predict

- `unitypredict-pp-cli predict create-predict` — POST /api/predict/{predict_id}
- `unitypredict-pp-cli predict get-logs` — GET /api/predict/status/{statu_id}/logs
- `unitypredict-pp-cli predict get-status` — GET /api/predict/status/{statu_id}

**repository** — Operations on search

- `unitypredict-pp-cli repository` — GET /api/repository/search

**social** — Operations on review

- `unitypredict-pp-cli social <review_id>` — GET /api/social/model/review/{review_id}

**tools** — Operations on analytics

- `unitypredict-pp-cli tools list-analytics` — GET /api/tools/analytics
- `unitypredict-pp-cli tools list-dev.abhirupdas-8c0af724114435` — GET /api/tools/public/dev.abhirupdas_8C0AF724114435


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
unitypredict-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query. `--json` (and other machine formats) keep that exit-2 contract and write `{"matches":[]}` on stdout so agents can inspect the envelope without treating a miss as success.

## Auth Setup

Durable API keys live in ~/.unitypredict/credentials keyed dev/prod; the header needs the literal APIKEY@ prefix (plain keys 401 on strict routes). --env dev|prod selects the tenant; prod mutations refuse without --i-know.

Run `unitypredict-pp-cli doctor` to verify setup.

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
  unitypredict-pp-cli auth-resource --agent
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

- Use `--home <dir>` for one invocation, or set `UNITYPREDICT_HOME=<dir>` to relocate all four path kinds under one root.
- Use per-kind env vars only when a specific kind must diverge: `UNITYPREDICT_CONFIG_DIR`, `UNITYPREDICT_DATA_DIR`, `UNITYPREDICT_STATE_DIR`, `UNITYPREDICT_CACHE_DIR`.
- Resolution order is per-kind env var, `--home`, `UNITYPREDICT_HOME`, XDG (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`), then platform defaults.
- `config` contains settings like `config.toml` and profiles. `data` contains `credentials.toml`, `data.db`, cookies, and auth sidecars. `state` contains persisted queries, jobs, and `teach.log`. `cache` contains regenerable HTTP/cache files.
- Stored secrets live in `credentials.toml` under the data dir. Existing legacy `config.toml` secrets are read for compatibility and leave `config.toml` on the first auth write.
- Run `unitypredict-pp-cli doctor --fail-on warn` to surface path and credential-location warnings. `agent-context` exposes a schema v4 `paths` block for agents that need the resolved dirs.
- For MCP, pass relocation through the MCP host config. The MCP binary does not inherit CLI flags:

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

Fleet precedence: an inherited per-kind env var overrides an explicit `--home` for that kind. Use `UNITYPREDICT_HOME` or per-kind vars as durable fleet levers, and use `--home` only for a single invocation. Relocation is not reversible by unsetting env vars; move files manually before clearing `UNITYPREDICT_HOME`, or `doctor` will not find credentials left under the former root.

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
unitypredict-pp-cli recall "$QUERY" --agent
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
      "next_action": ["<trial command>", "unitypredict-pp-cli learnings confirm 12"] }
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
       materially more, record the divergence via `unitypredict-pp-cli playbook amend`
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

Candidate judgment details: `learnings confirm <id>` prints the candidate's full payload before materializing it - check that the printed payload matches the behavior you verified. `learnings reject <id>` tombstones the derivation signature so the same candidate does not resurface. The envelope carries only the few candidates worth acting on now; `unitypredict-pp-cli learnings candidates` lists the full open set.

Graceful degradation: if `learnings confirm` is an unknown command, you are driving an older binary - ignore the candidates guidance and follow the rest of the protocol.

### Step 3: always read `warnings`

- `low_confidence`: row exists at `confidence<2`. Treat as a hint, not a skip-discovery hit.
- `resource_not_in_store`: the local store doesn't have the resource the learning points at. The match validator couldn't classify entities — direct-fetch and re-evaluate.
- `cross_alias_match` (per-result): the row was taught under a different alias and matched the live query's canonical via `entity_lookups` (e.g., a "USA" teach satisfying a "United States" recall). Trust the resource_id.
- `similar_shape_different_entity:<canonical>` (top-level): a structurally matching row exists but its canonical entity differs from the live query's. Treated as cold start; the warning carries the conflicting canonical as a hint, but the row is NOT promoted into Results.
- `ambiguous_alias` (top-level): a single query entity resolved to multiple canonicals (e.g., "Cards" → Arizona Cardinals + St. Louis Cardinals). Surface the ambiguity from context before committing to a resource.
- `candidates_present` (top-level): the envelope carries a `candidates` section. Handle it via the candidates branch in Step 2 before anything else.
- `lookup_refresh_available` (top-level): an entity in the query has no lookup row yet, but synced data could provide one. Run `unitypredict-pp-cli sync` to refresh entity lookups.
- Top-level `no_learnings_for_query_family`: the table had no rows above the Jaccard floor. Pure cold start.

### Step 4: `teach &` after finalizing your response - always

Teaching is unconditional. After resolving a query the store could not answer, background-teach the final resource mapping - no call-count threshold, no judging whether it was "worth" learning. The teach is the anchor of the loop: it triggers playbook synthesis for a family without a playbook, and same-referent phrasings fold into one family so near-duplicate teaches do not fragment the store. Fire it after assembling your user-facing response but BEFORE emitting it, with a shell `&` so the call returns immediately. Pass the query the same way as recall — argv/MCP, or file-then-`$QUERY`. Do not splice the question into the command text:

```bash
QUERY=$(cat /path/to/question.txt)
unitypredict-pp-cli teach --query "$QUERY" --resource-type <type> --resource <id1> --resource <id2>
# (append shell `&` to background it)
```

Silent on success. Errors only land in `teach.log` under the resolved state dir. Teach the **most specific** resource - if the user asked a broad question and you walked through parent records to find the specific answer, teach the leaf id, not the parent. The CLI uses seeded `entity_lookups` for cross-alias resolution at recall time, so a teach under one alias (e.g., "Niners") satisfies future queries under another alias (e.g., "49ers", "San Francisco") automatically.

PII rule: teach the structural question with identifiers stripped - never include names, emails, phone numbers, account ids, or other personal identifiers in taught queries or notes. The CLI scans teach queries for obvious email/phone shapes and warns, but does not block; strip before teaching rather than relying on the warning.

### Step 5: playbooks - optional flags, automatic synthesis

You do not need to decide whether a session "deserves" a playbook: a teach on a family without one auto-synthesizes a `playbook_candidate` from the session's journal, and the next session judges it via confirm/reject. Attach explicit playbook flags only when you already hold choreography worth recording verbatim - workarounds the CLI didn't surface (silently-dropped flags, undocumented params, pagination tricks, payload gotchas). Prefer the **integrated one-call form** - record the resource learning and the playbook in the same `teach` invocation:

```bash
# Common case: record both the resource learning AND the playbook in one call.
QUERY=$(cat /path/to/question.txt)
unitypredict-pp-cli teach \
  --query "$QUERY" \
  --resource <id> \
  --playbook-file ~/playbooks/<shape>.json \
  --playbook-notes-file ~/playbooks/<shape>-notes.md
# (append shell `&` to background it)

# Alternate: playbook-only (no resource to record alongside).
QUERY=$(cat /path/to/question.txt)
unitypredict-pp-cli teach-playbook \
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
unitypredict-pp-cli playbook amend \
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

`unitypredict-pp-cli learnings stats` reports recall hit rate, teach-to-reuse, playbook resolution rate, and candidate confirm/reject counts from the local `learn_events` table. Rates are null until they have a denominator; everything stays on this machine. Use it to check whether the loop is earning its keep for this CLI.

### Disabling learning

- `--no-learn` on a single command short-circuits both `recall` and the `teach` write path. Use for deterministic agent flows or tests that must not be affected by accumulated learnings.
- `UNITYPREDICT_NO_LEARN=true` in the environment globally disables the pipeline.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
unitypredict-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
unitypredict-pp-cli feedback --stdin < notes.txt
unitypredict-pp-cli feedback list --json --limit 10
```

Entries are stored locally as `feedback.jsonl` under the resolved data dir. They are never POSTed unless `UNITYPREDICT_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `UNITYPREDICT_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

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
unitypredict-pp-cli profile save briefing --json
unitypredict-pp-cli --profile briefing auth-resource
unitypredict-pp-cli profile list --json
unitypredict-pp-cli profile show briefing
unitypredict-pp-cli profile delete briefing --yes
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

1. **Empty, `help`, or `--help`** → show `unitypredict-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/ai/unitypredict/cmd/unitypredict-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add unitypredict-pp-mcp -- unitypredict-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which unitypredict-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   unitypredict-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `unitypredict-pp-cli <command> --help`.

## When NOT to use (anti-triggers)

- Deploying engine images (use the first-party unitypredict SDK CLI)
- Account/API-key management (POST /api/auth is excluded — it clobbers the account record on partial bodies)
- Wrong-env key symptoms (empty prod lists): run `auth check --env prod` before concluding data is missing
