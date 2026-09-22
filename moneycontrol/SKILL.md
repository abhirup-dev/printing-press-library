---
name: pp-moneycontrol
description: "Live Indian market news and research context for every holding review"
author: "dev-abhirup-sc"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - moneycontrol-pp-cli
    install:
      - kind: go
        bins: [moneycontrol-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/finance/moneycontrol/cmd/moneycontrol-pp-cli
---

# Moneycontrol — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `moneycontrol-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install moneycontrol --cli-only
   ```
2. Verify: `moneycontrol-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.6 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/finance/moneycontrol/cmd/moneycontrol-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

Moneycontrol-pp-cli turns Moneycontrol's public financial pages and price feeds into source-linked, read-only commands for Indian market research. It adds live holding context joins, event-risk views, catalyst filtering, and research packets without storing articles, portfolio data, or response history.

## When to Use This CLI

Use this CLI for Indian market news, company-tagged stories, index and stock snapshots, financial research context, event risk, and source-linked holding reviews. It is especially useful when a broker or another tool supplies current holdings and the agent needs Moneycontrol's narrative around them. Keep it read-only and live; use the broker for execution and portfolio mutation.

## Anti-triggers

Do not use this CLI for:
- Do not use this CLI to place orders, trade, pay, or mutate a portfolio/watchlist/alert.
- Do not use it as a local article archive, sync database, offline search index, or historical cursor store.
- Do not use it for politics, entertainment, lifestyle, or non-financial news.
- Do not treat fragile or Super Pro surfaces as guaranteed when the command reports blocked, empty, or shape-changed access.
- Do not use it instead of a broker or canonical portfolio ledger.

## Unique Capabilities

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

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

## Command Reference

**account** — Optional authenticated Moneycontrol account and Pro reads. Read-only only.

- `moneycontrol-pp-cli account alerts` — Read-only current alerts page. No alert mutation is exposed.
- `moneycontrol-pp-cli account entitlement-status` — Authenticated account/Pro entitlement status used to select legitimate reads.
- `moneycontrol-pp-cli account portfolio` — Read-only current portfolio page. No portfolio mutation is exposed.
- `moneycontrol-pp-cli account pro-stocklist` — Super Pro stock-list widget; best-effort authenticated read.
- `moneycontrol-pp-cli account user-details` — Authenticated user/account detail shape used for endpoint gating; sensitive fields are redacted by the CLI.
- `moneycontrol-pp-cli account watchlist` — Read-only current watchlist page. No watchlist mutation is exposed.

**articles** — News articles across categories and per-stock tags.

- `moneycontrol-pp-cli articles by-category` — News headlines within a financial category (e.g.
- `moneycontrol-pp-cli articles by-tag` — News headlines tagged to a stock or financial topic (e.g. reliance-industries, infosys, nifty).
- `moneycontrol-pp-cli articles get` — Full financial article page (title, description, and raw body HTML for extraction).
- `moneycontrol-pp-cli articles latest` — Latest financial news headlines across all categories.

**history** — Live NSE stock OHLCV history from the priceapi chart feed.

- `moneycontrol-pp-cli history` — Historical or intraday OHLCV bars for an NSE symbol.

**indices** — Indian index quotes (SENSEX, NIFTY 50, NIFTY BANK, NIFTY IT).

- `moneycontrol-pp-cli indices` — Live quote for an Indian index.

**markets** — Financial market pages and institutional activity.

- `moneycontrol-pp-cli markets commodities` — Commodity, gold and silver market pages.
- `moneycontrol-pp-cli markets corporate-actions` — Corporate actions page; best-effort route.
- `moneycontrol-pp-cli markets earnings` — Earnings/results pages; best-effort route.
- `moneycontrol-pp-cli markets fii-dii` — FII and DII cash/F&O activity page; best-effort page extraction.
- `moneycontrol-pp-cli markets fno` — Futures, options and open-interest market statistics.
- `moneycontrol-pp-cli markets forex` — Currency and forex pages; best-effort route.
- `moneycontrol-pp-cli markets indices-table` — Indian indices page and change-table context.
- `moneycontrol-pp-cli markets ipo` — IPO news and market pages.
- `moneycontrol-pp-cli markets mutual-funds` — Mutual-fund news and quote pages.

**price_volume** — Live stock price-volume period data from the Moneycontrol API.

- `moneycontrol-pp-cli price-volume` — Current and historical price-volume period data for a stock.

**research** — Company research pages and stock scanners.

- `moneycontrol-pp-cli research equity-research` — Equity research, forecasts and recommendations page.
- `moneycontrol-pp-cli research filings` — Company filings page; best-effort route.
- `moneycontrol-pp-cli research fundamental-scanner` — Fundamental stock scanner; empty/blocked widgets are reported.
- `moneycontrol-pp-cli research technical-scanner` — Technical stock scanner; empty/blocked widgets are reported.

**stocks** — Per-stock quote (full pricefeed: price, 52H/L, 1d-10y %chg, CAGR, volume, sector)

- `moneycontrol-pp-cli stocks` — Full NSE equity quote for a stock by its moneycontrol sc_id.

**trending** — Trending stocks widget (most-searched tickers right now).

- `moneycontrol-pp-cli trending` — Currently trending stocks on moneycontrol (best-effort; empty widgets are errors).


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
moneycontrol-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query. `--json` (and other machine formats) keep that exit-2 contract and write `{"matches":[]}` on stdout so agents can inspect the envelope without treating a miss as success.

## Auth Setup

Public news, quote, index, research, and market commands run without authentication. Optional portfolio, watchlist, alert, and Super Pro reads are endpoint-specific and use an established secure credential/session store or the generated Chrome bootstrap where supported; credentials and session values never appear in source, logs, reports, or output. No write or transaction command is provided.

Run `moneycontrol-pp-cli doctor` to verify setup.

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
  moneycontrol-pp-cli articles get --slug india/tribunals-reforms-bill-2026- --agent
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Read-only** — do not use this CLI for create, update, delete, publish, comment, upvote, invite, order, send, or other mutating requests

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

- Use `--home <dir>` for one invocation, or set `MONEYCONTROL_HOME=<dir>` to relocate all four path kinds under one root.
- Use per-kind env vars only when a specific kind must diverge: `MONEYCONTROL_CONFIG_DIR`, `MONEYCONTROL_DATA_DIR`, `MONEYCONTROL_STATE_DIR`, `MONEYCONTROL_CACHE_DIR`.
- Resolution order is per-kind env var, `--home`, `MONEYCONTROL_HOME`, XDG (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`), then platform defaults.
- `config` contains settings like `config.toml` and profiles. `data` contains `credentials.toml`, `data.db`, cookies, and auth sidecars. `state` contains persisted queries, jobs, and `teach.log`. `cache` contains regenerable HTTP/cache files.
- Stored secrets live in `credentials.toml` under the data dir. Existing legacy `config.toml` secrets are read for compatibility and leave `config.toml` on the first auth write.
- Run `moneycontrol-pp-cli doctor --fail-on warn` to surface path and credential-location warnings. `agent-context` exposes a schema v4 `paths` block for agents that need the resolved dirs.
- For MCP, pass relocation through the MCP host config. The MCP binary does not inherit CLI flags:

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

Fleet precedence: an inherited per-kind env var overrides an explicit `--home` for that kind. Use `MONEYCONTROL_HOME` or per-kind vars as durable fleet levers, and use `--home` only for a single invocation. Relocation is not reversible by unsetting env vars; move files manually before clearing `MONEYCONTROL_HOME`, or `doctor` will not find credentials left under the former root.

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
moneycontrol-pp-cli recall "$QUERY" --agent
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
      "next_action": ["<trial command>", "moneycontrol-pp-cli learnings confirm 12"] }
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
       materially more, record the divergence via `moneycontrol-pp-cli playbook amend`
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

Candidate judgment details: `learnings confirm <id>` prints the candidate's full payload before materializing it - check that the printed payload matches the behavior you verified. `learnings reject <id>` tombstones the derivation signature so the same candidate does not resurface. The envelope carries only the few candidates worth acting on now; `moneycontrol-pp-cli learnings candidates` lists the full open set.

Graceful degradation: if `learnings confirm` is an unknown command, you are driving an older binary - ignore the candidates guidance and follow the rest of the protocol.

### Step 3: always read `warnings`

- `low_confidence`: row exists at `confidence<2`. Treat as a hint, not a skip-discovery hit.
- `resource_not_in_store`: the local store doesn't have the resource the learning points at. The match validator couldn't classify entities — direct-fetch and re-evaluate.
- `cross_alias_match` (per-result): the row was taught under a different alias and matched the live query's canonical via `entity_lookups` (e.g., a "USA" teach satisfying a "United States" recall). Trust the resource_id.
- `similar_shape_different_entity:<canonical>` (top-level): a structurally matching row exists but its canonical entity differs from the live query's. Treated as cold start; the warning carries the conflicting canonical as a hint, but the row is NOT promoted into Results.
- `ambiguous_alias` (top-level): a single query entity resolved to multiple canonicals (e.g., "Cards" → Arizona Cardinals + St. Louis Cardinals). Surface the ambiguity from context before committing to a resource.
- `candidates_present` (top-level): the envelope carries a `candidates` section. Handle it via the candidates branch in Step 2 before anything else.
- Top-level `no_learnings_for_query_family`: the table had no rows above the Jaccard floor. Pure cold start.

### Step 4: `teach &` after finalizing your response - always

Teaching is unconditional. After resolving a query the store could not answer, background-teach the final resource mapping - no call-count threshold, no judging whether it was "worth" learning. The teach is the anchor of the loop: it triggers playbook synthesis for a family without a playbook, and same-referent phrasings fold into one family so near-duplicate teaches do not fragment the store. Fire it after assembling your user-facing response but BEFORE emitting it, with a shell `&` so the call returns immediately. Pass the query the same way as recall — argv/MCP, or file-then-`$QUERY`. Do not splice the question into the command text:

```bash
QUERY=$(cat /path/to/question.txt)
moneycontrol-pp-cli teach --query "$QUERY" --resource-type <type> --resource <id1> --resource <id2>
# (append shell `&` to background it)
```

Silent on success. Errors only land in `teach.log` under the resolved state dir. Teach the **most specific** resource - if the user asked a broad question and you walked through parent records to find the specific answer, teach the leaf id, not the parent. The CLI uses seeded `entity_lookups` for cross-alias resolution at recall time, so a teach under one alias (e.g., "Niners") satisfies future queries under another alias (e.g., "49ers", "San Francisco") automatically.

PII rule: teach the structural question with identifiers stripped - never include names, emails, phone numbers, account ids, or other personal identifiers in taught queries or notes. The CLI scans teach queries for obvious email/phone shapes and warns, but does not block; strip before teaching rather than relying on the warning.

### Step 5: playbooks - optional flags, automatic synthesis

You do not need to decide whether a session "deserves" a playbook: a teach on a family without one auto-synthesizes a `playbook_candidate` from the session's journal, and the next session judges it via confirm/reject. Attach explicit playbook flags only when you already hold choreography worth recording verbatim - workarounds the CLI didn't surface (silently-dropped flags, undocumented params, pagination tricks, payload gotchas). Prefer the **integrated one-call form** - record the resource learning and the playbook in the same `teach` invocation:

```bash
# Common case: record both the resource learning AND the playbook in one call.
QUERY=$(cat /path/to/question.txt)
moneycontrol-pp-cli teach \
  --query "$QUERY" \
  --resource <id> \
  --playbook-file ~/playbooks/<shape>.json \
  --playbook-notes-file ~/playbooks/<shape>-notes.md
# (append shell `&` to background it)

# Alternate: playbook-only (no resource to record alongside).
QUERY=$(cat /path/to/question.txt)
moneycontrol-pp-cli teach-playbook \
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
moneycontrol-pp-cli playbook amend \
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

`moneycontrol-pp-cli learnings stats` reports recall hit rate, teach-to-reuse, playbook resolution rate, and candidate confirm/reject counts from the local `learn_events` table. Rates are null until they have a denominator; everything stays on this machine. Use it to check whether the loop is earning its keep for this CLI.

### Disabling learning

- `--no-learn` on a single command short-circuits both `recall` and the `teach` write path. Use for deterministic agent flows or tests that must not be affected by accumulated learnings.
- `MONEYCONTROL_NO_LEARN=true` in the environment globally disables the pipeline.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
moneycontrol-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
moneycontrol-pp-cli feedback --stdin < notes.txt
moneycontrol-pp-cli feedback list --json --limit 10
```

Entries are stored locally as `feedback.jsonl` under the resolved data dir. They are never POSTed unless `MONEYCONTROL_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `MONEYCONTROL_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

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
moneycontrol-pp-cli profile save briefing --json
moneycontrol-pp-cli --profile briefing articles get --slug india/tribunals-reforms-bill-2026-
moneycontrol-pp-cli profile list --json
moneycontrol-pp-cli profile show briefing
moneycontrol-pp-cli profile delete briefing --yes
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
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `moneycontrol-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/finance/moneycontrol/cmd/moneycontrol-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add moneycontrol-pp-mcp -- moneycontrol-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which moneycontrol-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   moneycontrol-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `moneycontrol-pp-cli <command> --help`.
