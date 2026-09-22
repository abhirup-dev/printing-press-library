# ChatGPT Conversations CLI — Absorb Manifest

## Absorbed (match or beat everything that exists)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | Export conversations as Markdown | chatgpt-exporter, ai-chat-md-export, revivalstack/ai-chat-exporter, dkasak userscript | (behavior in chatgpt-pp-cli transcript --format md) markdown export via -o file | Live-API or local-store source, --since filters, branch-aware, TOC in output |
| 2 | Export/backup raw JSON | FdezRomero/chatgpt-exporter, ezwep/chatgpt-exporter | chatgpt-pp-cli get <id> --json | Complete message array incl. hidden/tool nodes, pagination-safe, no truncation |
| 3 | List conversations w/ pagination | shoyu-ramen/codex-chats-mcp (list_conversations) | (generated endpoint) conversations list | Time filters, ordering, page-walk (total is a hint), --agent output |
| 4 | Get single conversation | codex-chats-mcp (get_conversation) | (generated endpoint) conversations get | Handles messages[] + page_info model (current API), masked-404 detection |
| 5 | Search conversations (server) | codex-chats-mcp (search_conversations); UI Ctrl+K | chatgpt-pp-cli search <query> | Uses POST /global/search (federated, snippets, message-level ids) as DEFAULT per user directive |
| 6 | Offline search over exports | bdelanghe/mcp-conversations-sqlite, mychatarchive, Chronicle-MCP | (behavior in chatgpt-pp-cli search --offline) | FTS5 over synced full bodies + titles + snippets; clearly labeled secondary |
| 7 | Integrity checks on extraction | 2015pulsar/chatgpt-conversation-extractor | (behavior in chatgpt-pp-cli transcript --verify) | Completeness vs current_node, role/status audit, count reconciliation |
| 8 | Reasoning/tool/code node handling | 2015pulsar (recaps, tool results) | (behavior in chatgpt-pp-cli transcript --show reasoning\|tools\|code) | thoughts/reasoning_recap/model_editable_context/content types deliberate, default human-readable only |
| 9 | TOC/outline of long threads | revivalstack (TOC in export) | chatgpt-pp-cli outline <id> | Live derivation from user-message anchors w/ timestamps, turn ranges, machine JSON |
| 10 | Models list | pi-gpt, codex-chats-mcp | (generated endpoint) models list | Slugs, titles, reasoning_type, thinking efforts, default badge |
| 11 | Token-based auth from browser session | FdezRomero (--token flag), pi-gpt | chatgpt-pp-cli auth login/import | Cookie jar outlives 10-day token TTL; auto re-mint via /api/auth/session; masked-empty detection |
| 12 | Send message / follow-up | pi-gpt (f/conversation + sentinel) | chatgpt-pp-cli chat send <msg> --conversation-id <id> --model <slug> | EXPLICITLY EXPERIMENTAL/browser-gated per supervisor constraint; dry-run preview default posture; sentinel PoW + conduit implemented in Go |
| 13 | Delete/archive/rename conversations | codex-chats-mcp | NOT ABSORBED — deliberately out of scope | Mutating surface excluded by operator safety boundary (existing-conversation protection) |
| 14 | Sync to local store | bdelanghe (sqlite), mychatarchive | (generated) sync --resources conversations | Incremental by updated_at watermark; syncs FULL bodies (prior CLI stored list metadata only) |

## Transcendence (only possible with our approach)
Inline-run 3-pass brainstorm (operator prohibited subagent delegation; structure honored: customer model → candidates → adversarial cut).

| # | Feature | Command | Score | Buildability | How It Works | Evidence | Long Description |
|---|---------|---------|-------|--------------|--------------|----------|------------------|
| 1 | Rich chronological inventory with hydrated metadata | list --since 30d --order created --hydrate | 10/10 | hand-code | conversations list page-walk + detail hydration cache (turn counts, span, content size) in SQLite; list responses lack counts (verified empirically) | Operator pillar 1; empirical: list items have 27 metadata fields but no turn counts; codex-chats-mcp list returns raw items only | none |
| 2 | Long-thread outline by user anchors | outline <id> | 10/10 | hand-code | Derives sections from user messages (TableOfContentsSidebar parity, verified via React fiber + DOM): label excerpt, anchor msg id, create_time, msg range | Operator scope; UI TOC proven client-side derived; no community tool has outline CLI | Use this command to map a long thread's structure. Do NOT use it for full content; use 'transcript' instead. |
| 3 | Section-sliced transcript extraction | transcript <id> --section 3 (and --from-anchor/--to-anchor) | 9/10 | hand-code | Outline anchors index into messages[]; slice = range between user-message anchors; timestamps preserved | Operator scope; user screenshot workflow; no exporter supports section slicing | Use this command to extract part of a long thread. Do NOT use it for the whole conversation; use 'transcript <id>' without slicing flags. |
| 4 | Complete extraction w/ completeness verification | transcript <id> --verify | 9/10 | hand-code | messages?after= cursor loop until last-id == current_node; verify counts, end_turn chain, status fields | Operator pillar 2; empirically has_next_page can lie on truncated windows; 2015pulsar has integrity checks (absorbed) | none |
| 5 | Dual-mode search (server-fresh + offline) | find <query> (server) + search <query> (offline, framework FTS) | 8/10 | hand-code | `find` = POST /global/search (federated snippets, message ids, default); framework `search` = FTS5 over synced bodies (offline secondary) | User directive: UI Ctrl+K is canonical; bdelanghe/mychatarchive prove offline-search demand | Use 'find' for fresh server-side search. Use framework 'search' when working without a session; results come from the last sync. |

## Supervisor-approved scope (2026-09-20, refined option B)
1. Inventory: normalized timestamps, human/assistant/raw counts, span + time filters, bounded incremental hydration, cache. 2. Verified-complete extraction (plural+page_info + singular compat), JSON + transcript. 3. Outline from user anchors + section/anchor/range slicing. 4. Dual search (server default, offline secondary). 5. Auth import + re-mint (cookies outlive TTL; optional codex auth source). 6. Models list. 7. Send/follow-up/status/wait: experimental, browser-gated, legitimate tokens only, no turnstile solver.
Explicitly skipped: bulk/attachment/project backup, destructive management, deep-research/agent parity, daily journal, model analytics, unrelated endpoints.

## Stubs
None. The send surface (row 12) is NOT a stub: sentinel PoW, conduit prepare, f/conversation POST, and SSE typed-event parsing are implemented in Go; the turnstile token is consumed when legitimately available (browser-assisted auth refresh) and the command fails with an explicit, actionable error otherwise. Tiered honest in README/help: experimental + browser-gated.

## Killed candidates (from Pass 3 cut + supervisor scope cut 2026-09-20)
| Feature | Kill reason | Closest surviving sibling |
|---------|-------------|--------------------------|
| Chat activity journal (daily) | Supervisor scope cut: explicitly skipped | list --hydrate |
| Model usage analytics | Supervisor scope cut: explicitly skipped | list --hydrate |
| Feature | Kill reason | Closest surviving sibling |
|---------|-------------|--------------------------|
| topics timeline (all threads' prompts over time) | Monthly-at-best ritual; overlaps daily | daily |
| find-links / citations extractor | Not weekly; citations already in transcript --json | transcript |
| data-quality audit standalone | Overlaps --verify; not a user-facing verb | transcript --verify |
| conversation diff (between syncs) | Speculative; nobody asked | list --hydrate |
