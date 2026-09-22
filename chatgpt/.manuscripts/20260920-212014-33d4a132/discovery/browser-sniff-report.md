# ChatGPT Browser-Sniff Discovery Report

## 1. User Goal Flow
- Goal: "Inventory conversations chronologically with rich metadata, then extract one complete conversation" (research brief workflow #1/#2). Secondary flows: search (Ctrl+K), long-thread TOC/outline navigation, send/continuation on a NEW dummy conversation (supervisor-approved).
- Steps completed:
  1. Load chatgpt.com home (authenticated) → captured app's own API calls (conversations list, models, gizmos, composer).
  2. Minted bearer in-page from /api/auth/session → validated auth matrix (bearer-only 200 / cookie-only 200-masked-empty / no-auth 401).
  3. Listed conversations (offset/limit/order params; total is a limit+1 hint, page-walk required).
  4. Opened long conversation /c/6aaa52c7… → app uses plural /conversations/{id}?num_turns=10; legacy singular 404s.
  5. Controlled pagination reads: num_turns=2 window truncation (has_next_page can lie); /messages?after= forward cursor proven.
  6. Search probes: legacy /search 404; /conversations/search?query= works; Ctrl+K UI traced → POST /global/search (cursor-paginated, federated).
  7. TOC trace: TableOfContentsSidebar React component; items = user messages; labels = verbatim user-prompt excerpts; click = scroll-to-anchor, URL unchanged.
  8. Models enumeration (19 slugs, default gpt-5-6).
  9. Write path (dummy conversation only): UI sends traced via CDP → sentinel prepare/finalize + f/conversation/prepare (conduit token) + POST /f/conversation (SSE, typed events, message_stream_complete terminal). Direct legacy POST /backend-api/conversation → 403 (dead path; turnstile-gated).
- Steps skipped: none.
- Coverage: 9/9 steps completed.

## 2. Pages & Interactions
- https://chatgpt.com/ (home) — observed app API surface; verified session (no login button).
- /c/6aaa52c7-… (user's long thread) — READ ONLY: detail fetches, pagination probes, TOC rail click ("Prompt 2" → scrolled to user-message anchor bbb2155c, URL unchanged), hover popover inspection, scroll behaviors.
- Ctrl+K global search dialog — typed "haptics", "camera comparison long term"; captured POST /global/search.
- /c/6ab00407-… (NEW dummy conversation created for write tests) — 6 benign UI sends ("reply OK/DONE/TRACED/PROTOCOL-OK/CAPTURED") traced via CDP; NO existing conversation was mutated, renamed, archived, or deleted. Dummy left in place per instruction.
- Right-edge TOC rail interaction (5 rows) on the long thread; DOM/React fiber inspection (no page mutation).

## 3. Browser-Sniff Configuration
- Backend: ego-browser (operator-mandated), TaskSpace id=1, single page p1, page.fetch + page.evaluate + CDP Network domain + keyboard/UI driving. No browser-use/agent-browser (declined at preflight; ego-browser supersedes).
- Pacing: ~1s between API probes; no 429s encountered.
- Proxy pattern: none (REST/JSON + SSE; no proxy-envelope, no GraphQL BFF).
- CDP Network.enable used within single script invocations for POST header/body + SSE capture.

## 4. Endpoints Discovered
| Method | Path | Status | Content-Type | Auth |
|---|---|---|---|---|
| GET | /api/auth/session | 200 | application/json | cookies (session); mints bearer |
| GET | /backend-api/conversations | 200 | application/json | bearer required |
| GET | /backend-api/conversations/{id} | 200 | application/json | bearer required |
| GET | /backend-api/conversations/{id}/messages | 200 | application/json | bearer required (after= cursor) |
| GET | /backend-api/conversations/search | 200 | application/json | bearer required |
| POST | /backend-api/global/search | 200 | application/json | bearer required |
| GET | /backend-api/models | 200 | application/json | bearer required |
| POST | /backend-api/sentinel/chat-requirements | 200 | application/json | bearer (PoW + turnstile context) |
| POST | /backend-api/f/conversation/prepare | 200 | application/json | bearer + sentinel context |
| POST | /backend-api/f/conversation | 200 | text/event-stream | bearer + conduit + sentinel + turnstile |
| GET | /backend-api/conversations/{id}/stream_status | 404* | application/json | (app polls during active streams; 404 when idle) |
| POST | /backend-api/conversation (legacy singular) | 403 | application/json | dead for API-only callers |

Auth classification: all /backend-api endpoints are auth-required (cookie-only returns masked empty data: 200 with total:0 or 404 conversation_inaccessible).

## 5. Traffic Analysis (traffic-analysis.json)
- Protocols: rest_json (high confidence); SSE streaming on f/conversation.
- Auth signals: Authorization: Bearer <JWT from /api/auth/session>; OAI-Device-Id, OAI-Session-Id, OAI-Client-Version headers on sends; x-conduit-token; OpenAI-Sentinel-* headers (requirements/proof/turnstile) — names only, values never captured.
- Protection signals: Cloudflare (challenge-platform scripts on load); sentinel PoW (sha3-512 hashcash, ~1/40 difficulty observed); Turnstile required=true for sends on this account; legacy POST path returns 403 "Unusual activity".
- Generation hints: requires_browser_auth=true (cookie import + bearer mint), requires_protected_client for sends only; reads replay via bearer (bearer-only worked from browser context — CLI runtime transport settled at Phase 1.9 probe).
- Warnings: error_status_cluster (the two deliberate 403/404 probes documenting dead/legacy paths); reserved_resource_name. Both are expected evidence, not blockers.

## 6. Coverage Analysis
- Exercised: session/auth, list, detail, messages pagination, search (legacy + conversations/search + global/search), models, TOC derivation, full send protocol (prepare→conduit→SSE), stream_status probe.
- Likely missed: archived/starred filtering variants (params known), projects/gizmos (out of scope), images/documents search tabs (source_statuses show them as sources), voice mode. All out of the two-pillar scope.

## 7. Response Samples
- See research/chatgpt-browser-sniff-spec-samples/ (12 endpoints; bodies embedded in capture with structural placeholders for user content and REDACTED for all credential values).
- Key shapes: list item (27 fields incl. ISO timestamps); detail (unix-float timestamps, messages[], page_info, current_node); message ({id, author{role}, create_time, content{content_type,parts}, status, weight, recipient, metadata{turn_id, model_slug, is_visually_hidden_from_constance…}}); global search item ({id:"conversation:<cid>:message:<mid>", title, snippet, match_kind, payload{conversation_id, message_id}}); SSE typed events (delta_encoding v1, resume_conversation_token, input_message, message_marker, server_ste_metadata, message_stream_complete, [DONE]).

## 8. Rate Limiting Events
- None encountered. ~30 authenticated requests across ~25 minutes, paced ≥1s. effective rate < 0.05 req/s.

## 9. Authentication Context
- Authenticated session used throughout (operator-provided Ego Browser session).
- Transfer method: in-page bearer mint from /api/auth/session (page context holds cookies + token; token values never left the page).
- Auth-only endpoints: all /backend-api/* (entire surface).
- Scheme: Authorization: Bearer <JWT, TTL 10 days>; session expires 2026-12-19 (~3 months) → cookies outlive token; re-mint while session lives.
- Cookie-only replay behavior: masked empty (list: total 0; detail: 404 conversation_inaccessible) — CLI must detect both masked shapes and re-mint.
- Session state excluded from manuscripts: yes (SESSION_DIR outside DISCOVERY_DIR; no cookie/token values in any artifact).

## 10. Bundle Extraction
- Not run. All endpoints were discovered empirically through the authenticated session with response bodies; bundle grepping unnecessary. Note: page scripts are CSP-restricted; only challenge-platform script visible via script[src] enumeration.
