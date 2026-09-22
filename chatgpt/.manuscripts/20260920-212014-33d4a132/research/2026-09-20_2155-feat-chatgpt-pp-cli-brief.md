# ChatGPT Conversations CLI Brief

## API Identity
- Domain: chatgpt.com (web app internal API; no official public spec for consumer conversation data)
- Users: power users who live in long ChatGPT threads and want local, offline, machine-readable access to their conversation history
- Data profile: conversation metadata (id, title, created_at, updated_at), full message trees (mapping of nodes with parent/children edges), per-message content parts (text, code, tool calls, reasoning), long-thread section/TOC structure (to be traced), user account/session state

## Reachability Risk
- High, but mitigated by design. Evidence: community wrappers (revChatGPT lineage, chatgpt-exporter, conversation extractors) report 401 when session cookies/bearer missing, and Cloudflare 403 challenges against headless clients. chatgpt.com is an authenticated, bot-protected web app.
- Mitigations baked into the build: (1) auth runs inside the user's authenticated browser context (Ego Browser) during discovery; (2) runtime CLI re-uses persisted cookies + re-mints the short-lived Bearer from /api/auth/session (cookies must outlive token TTL — prior CLI bug expired whole session on token TTL); (3) backend calls carry Bearer AND cookies (cookie-only historically returned masked 404); (4) all empirical facts re-validated this run, not copied from old reverse-engineering notes.
- Historical facts to validate (not assume): slug `chatgpt`, binary `chatgpt-pp-cli`, offset/limit pagination on list, GET /backend-api/conversation/{id} returning complete mapping tree in one response (infinite scroll was DOM virtualization), transcript via current_node ancestry with reasoning/tool/code nodes needing deliberate rendering, useful surface = list/get/transcript/sync/search, prior generic sync stored list metadata only (improve: sync conversation detail bodies for local full-text search).

## Top Workflows
1. Inventory sweep: list all chats chronologically with rich metadata (id, title, created, updated, turn count, length/span), filtered by time window — "what did I work on in August?"
2. Full extraction: pull one conversation by ID completely — no truncation — as machine-readable JSON and a readable transcript (chronological turns, timestamps, branch/current_node handled correctly)
3. Local search: after sync, find chats/turns by keyword across full conversation bodies (not just titles), ranked with timestamps so results are actionable
4. Long-thread navigation: outline a very long thread by its section/topic labels (ChatGPT's TOC navigator) and slice the transcript to a section or turn range (scoped enhancement — to be traced empirically)
5. Session resilience: long-lived cookie jar with automatic bearer re-mint so the CLI keeps working across days

## Table Stakes (from exporters/wrappers ecosystem)
- Export conversation to markdown/JSON (chatgpt-exporter, various userscripts)
- Pagination walk of conversation list
- Title/updated metadata in listings
- Auth via session cookies (browser-derived)

## Data Layer
- Primary entities: conversation (metadata + detail body), message node (id, role, created, content parts, recipient), section/outline entry (label, anchor node, ordering, turn range — pending trace), session (cookies + token mint timestamp)
- Sync cursor: conversation updated_at watermark + already-synced detail bodies; detail hydration is incremental (only conversations whose updated_at advanced)
- FTS/search: SQLite FTS5 over message bodies + titles + section labels, with timestamps and conversation ids as queryable columns

## User Vision
- Two acceptance pillars (operator-stated, priority over generic surface):
  1. Conversation inventory/discovery: list + search by timestamp with rich per-chat metadata (id, title, created, updated, turn/message count, length measure: time span and/or content size). If list responses lack counts, hydrate details efficiently and cache locally. Chronological ordering + practical time filters.
  2. Full-chat extraction: fetch one conversation by ID, emit complete chat without truncation; chronological turn order + timestamps preserved; machine-readable JSON and readable transcript/export; branches/current_node handled correctly; completeness verified against a long real conversation.
- Enhancement: long-thread TOC/outline — trace ChatGPT's topic navigator (truncated topical labels per section, current highlighted); determine API vs mapping-metadata vs client-side derivation; preserve label, stable node anchor, ordering, timestamps/turn range, conversation id in local index; design outline command + transcript slicing by section/anchor/range. Evidence first, then model; don't force names prematurely.
- Sync/search exist to serve pillar 1; de-prioritize unrelated endpoints and broad feature expansion.

## Product Thesis
- Name: chatgpt (slug), chatgpt-pp-cli (binary) — per historical convention, validated this run
- Why it should exist: ChatGPT holds users' accumulated knowledge work hostage in a web UI with weak local export; this CLI turns the conversation history into a queryable local archive — chronological inventory, complete extraction, offline full-text search — with session auth that survives token expiry

## Build Priorities
1. Auth/session: cookie persistence + bearer re-mint from /api/auth/session; bearer+cookies on all backend-api calls; masked-404 detection (cookies missing)
2. list (offset/limit walk, metadata, time filters, ordering) + detail hydration cache (turn counts, span) for pillar 1
3. get (full JSON) + transcript (current_node ancestry, roles, timestamps, code blocks; reasoning/tool nodes filtered deliberately) for pillar 2
4. sync (metadata + full bodies, incremental) + search (FTS over bodies/titles) serving pillar 1
5. outline/slice for long threads — pending empirical trace of the TOC navigator
6. Live dogfood against the real account: completeness check on a long conversation is a gate
