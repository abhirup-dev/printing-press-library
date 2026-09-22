# chatgpt-pp-cli Build Log — Phase 3

Manifest transcendence rows: 5 planned, 5 built. Phase 3 will not pass until all 5 ship.

## Priority 0 — foundation
- `internal/gptapi/` (new): session-aware client — Netscape cookie jar (0600), bearer mint via /api/auth/session with JWT exp tracking, masked-response detection (200+total:0 / 404 conversation_inaccessible) with one re-mint retry, codex auth.json fallback, adaptive rate limiting, sentinel requirements + sha3-512 PoW solver (golang.org/x/crypto/sha3), conduit prepare, SSE typed-event stream reader (message_stream_complete = only terminal; [DONE]/close advisory).
- `internal/gptconv/` (new, pure logic + tests): message/conversation model, mixed-format timestamp normalization (ISO strings in list, unix floats in detail), user-anchor outline derivation (TOC parity), section/anchor slicing, hydration stats (user_turn_count / assistant counts / tool / system / hidden / total / span / bytes), completeness verification, markdown transcript rendering with deliberate reasoning/tool/hidden filtering. 8 table-driven tests, all pass.

## Priority 1 — absorbed (generator-emitted + wiring)
- conversations list/get/get-tree/list-messages/search/stream-status, global-search, models, chat requirements/prepare/send (typed endpoint commands, generated).
- `auth import --cookies <file> | --codex` (new): imports session, mints bearer, reports expiry (no secrets in output; refuses under harness).
- sync: framework sync covers conversations list metadata; full bodies accumulate into `conversations_detail` (FTS-indexed) + `conversation_hydration` (stats cache) via `list --hydrate` and every detail fetch.

## Priority 2 — transcendence (all built)
1. `list` — chronological inventory: --since/--before/--updated-within (duration or ISO), --order created|updated (client-side sort), --hydrate (incremental detail hydration + counts + span + bytes, cached), --limit + --max-scan-pages (scan/filter caps, note on cap-hit), normalized timestamps, --agent/--json via generated helpers. BUILT.
2. `transcript` — complete extraction: plural fetch + forward-cursor completion loop until tail == current_node; --verify (tail match, per-message status audit, counts); --section N / --from-anchor/--to-anchor slicing; --format md (readable transcript, -o file); --show reasoning,tools,hidden,code; --tree legacy mapping read. BUILT.
3. `outline` — TOC derivation from user anchors: index, anchor message id, verbatim excerpt, create_time (ISO + raw), message range, assistant reply count. BUILT.
4. `find` — server global search (default): strict body {query,limit,cursor}, cursor pages (--all, --max-scan-pages), snippets + message ids, source_statuses + partial_results surfaced. Framework `search` remains the offline FTS secondary. BUILT.
5. Dual-mode search — see 4 + framework search over synced bodies. BUILT.

## Priority 3 — polish
- which-index entries normalized to flag-free paths; root highlights updated (find, not search).
- Experimental send surface: `chat ask` (full protocol; explicit Turnstile failure message — no solver, never bypasses), `chat status` (state via current_node status), harness-refusing, --dry-run preview.
- `chatgpt--pp-cli` typo in transcript example fixed.
- Flag descriptions enriched from research (e.g. --since takes duration or ISO date).

## Completion gate
- Per-row Cobra resolution: all manifest paths resolve (list, transcript, outline, find, models, conversations list/get, chat send, auth import behavior rows verified).
- Deterministic backstop: dogfood novel_features_check found==planned, no missing, not skipped — PASS.
- go build ./... ok; go vet ok; go test ./... all packages ok.

## Intentionally deferred / out of scope (supervisor cut)
- daily journal, model-usage analytics (explicitly skipped per approval), bulk/attachment/project backup, destructive management (delete/archive/rename), deep-research/agent-mode chat, turnstile solving.

## Generator limitations found
- traffic-analysis generation_hints must match probe-reachability verdict (requires_page_context from capture auth-type label blocked generation until corrected to requires_browser_auth).
- which-index novel commands must be flag-free paths (test contract).
- Novel feature named `search` collides with the framework search command — renamed to `find` at absorb time; scaffolds reflect it.
