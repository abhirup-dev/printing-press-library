# Tickertape absorb manifest

## Scope and evidence

This manifest absorbs the approved Tickertape research surface from the live EgoBrowser discovery, the sanitized reverse-engineered community endpoint inventory, and the current internal endpoint spec. It is intentionally limited to online, read-only research. Every domain-data command fetches live data and emits provenance plus `observed_at`; it does not cache, sync, persist, or write stock/fund/report/scorecard/forecast/MMI/screener responses.

Authenticated and entitlement-aware routes are represented as read-only probes. The stock page in EgoBrowser TaskSpace 9 visibly showed analyst ratings, forecast, and current holdings and generated authenticated traffic for user, credit, portfolio, ratings, forecast, AI summary, and deal routes. The printed CLI must preserve 401/403/locked responses as access metadata rather than treating synthetic replay failures as proof that the account lacks access. Credential/session persistence, if later wired, belongs only in a secure credential store or permissioned auth config and never in source, manifests, logs, proofs, generated output, or domain-data storage.

No portfolio, watchlist, saved-screen, broker, trading, mutation, social, export, payment, loan, gold, or account workflow is absorbed.

## Absorbed public and entitlement-aware features

| # | Feature / route family | Best source | Our implementation | Added value |
|---:|---|---|---|---|
| 1 | Symbol and asset suggestions | Live `/search/suggest`; community wrapper | `(generated endpoint) search suggest` | Live search with structured JSON and provenance. |
| 2 | Market Mood Index now | Live `/mmi/now` | `(generated endpoint) market mood` | Preserves MMI components and timestamp without stale local state. |
| 3 | India/US market status | Live `gms-api /market/{market}/status` | `(generated endpoint) market status` | One typed command for both market calendars/status envelopes. |
| 4 | Indian security quote snapshots | Live `quotes-api /quotes` | `(generated endpoint) market quotes` | Explicit IDs and source host in output. |
| 5 | US latest quotes | Live `gms-api /quotes/US/latest` | `(generated endpoint) market us-latest-quotes` | Keyed ticker snapshots with live observation time. |
| 6 | Company identity and labels | Live `/stocks/info/{sid}` | `(generated endpoint) company info` | Keeps Tickertape identifiers, labels, ratios, and URL identity. |
| 7 | Company summary | Live `/stocks/summary/{sid}` | `(generated endpoint) company summary` | Preserves financialSummary, peers, forecast envelope, events, news, and presentations. |
| 8 | Company inter-day and intra-day charts | Live stock chart routes | `(generated endpoint) company chart-inter`, `(generated endpoint) company chart-intra` | Online chart snapshots without a resident WebSocket/browser sidecar. |
| 9 | Income, balance-sheet, cash-flow statements | Live `/stocks/financials/...` | `(generated endpoint) company financials` | Statement/period/view remain explicit source-specific parameters. |
| 10 | Company investment checklist | Live `/stocks/investmentChecklists/{sid}` | `(generated endpoint) company checklist` | Preserves checklist records instead of flattening them into advice. |
| 11 | Company news | Live `/stocks/news/{sid}` | `(generated endpoint) company news` | Distinct from event/presentation timeline and provenance-labelled. |
| 12 | Company key ratios | Live `/stocks/keyratioList/{sid}` | `(generated endpoint) company ratios` | Keeps the upstream ratio list and labels intact. |
| 13 | Company smallcase references | Live `/stocks/smallcases/{sid}` | `(generated endpoint) company smallcases` | Makes source-specific references visible without brokerage actions. |
| 14 | Six-part scorecard | Live `analyze-api /stocks/scorecard/{sid}` | `(generated endpoint) company scorecard` | Preserves `locked`, score, rank, peers, callout, comment, stack, and elements. |
| 15 | Entry-point and red-flag analysis | Scorecard elements | `(behavior in tickertape-pp-cli company scorecard) preserve source-specific element types and flags` | Does not merge opinionated signals into a universal investment score. |
| 16 | Analyst ratings | Live stock-page traffic and `/stocks/ratings/{sid}` | `(generated endpoint) company ratings` | Entitlement-aware: 401/403/locked fields are surfaced, not hidden. |
| 17 | Price/revenue/EPS forecasts | Live stock-page traffic and `/stocks/estimates/forecast/{sid}` | `(generated endpoint) company forecast` | Same explicit access treatment; no invented values when unavailable. |
| 18 | AI summary | Live stock-page traffic and `/stocks/aiSummary/{sid}` | `(generated endpoint) company ai-summary` | Read-only route with response/access envelope; no generated investment advice. |
| 19 | Company deal signals | Live stock-page traffic and `/stocks/aggregateddeals/{sid}` | `(generated endpoint) company aggregated-deals` | Preserves access state and source identity. |
| 20 | Mutual-fund universe | Live `/mutualfunds/list` | `(generated endpoint) mutualfund list` | Searchable source universe without a local mirror. |
| 21 | Mutual-fund identity/NAV/AMC | Live `/mutualfunds/{mfId}/info` | `(generated endpoint) mutualfund info` | Keeps IDs, NAV, option, AMC, sector, and labels. |
| 22 | Mutual-fund summary and peers | Live `/mutualfunds/{mfId}/summary` | `(generated endpoint) mutualfund summary` | Preserves scheme info, ratios, CAGR, tax, peers, and AMC details. |
| 23 | Mutual-fund holdings and allocation | Live `/mutualfunds/{mfId}/holdings` | `(generated endpoint) mutualfund holdings` | Shows current/allocation/sector fields with red-flag metadata. |
| 24 | Mutual-fund NAV and SIP charts | Live chart routes | `(generated endpoint) mutualfund chart-inter`, `(generated endpoint) mutualfund chart-sip` | Live chart retrieval with duration left to source contract. |
| 25 | Mutual-fund managers/checklists/widget | Live managers/checklists/widget routes | `(generated endpoint) mutualfund managers`, `checklist`, `widget` | Preserves source-specific fund diligence fields. |
| 26 | Equity screener filter metadata | Live `/screener/filters` | `(generated endpoint) screener equity-filters` | Exposes premium/locked filter metadata rather than pretending all filters work. |
| 27 | Equity prebuilt screens and universes | Live `/screener/prebuilt`, `/universes` | `(generated endpoint) screener equity-prebuilt`, `equity-universes` | Browse prebuilt research ideas without saved-screen state. |
| 28 | MF screener filter metadata | Live `/mf-screener/filters` | `(generated endpoint) screener mf-filters` | Keeps MF-specific filter taxonomy. |
| 29 | MF prebuilt screens and universes | Live `/mf-screener/prebuilt`, `/universes` | `(generated endpoint) screener mf-prebuilt`, `mf-universes` | Supports discovery while avoiding account/saved-screen workflows. |
| 30 | Read-only screener query envelopes | Community endpoint inventory; request contract not fully validated | `(stub) requires safe request-contract fixture and premium-field validation` | Explicitly deferred rather than guessing POST bodies or silently dropping premium errors. |
| 31 | Public deal ideas | Live `analyze-api /stocks/deals/ideas` | `(generated endpoint) deals ideas` | Keeps idea records and provenance separate from personal holdings. |
| 32 | Deal insight query | Live `analyze-api /stocks/deals/insight` | `(generated endpoint) deals insight` | Requires explicit source query fields; 400s remain actionable errors. |
| 33 | US stock/ETF identity | Live GMS security/ETF info routes | `(generated endpoint) us security-info`, `us etf-info` | Cross-asset research without broker/account operations. |
| 34 | US stock/ETF overview | Live GMS overview routes | `(generated endpoint) us security-overview`, `us etf-overview` | Preserves metrics, labels, peers, holdings, asset type, and ETF type. |
| 35 | US charts and financials | Live GMS charts/financials routes | `(generated endpoint) us chart`, `us financials` | Statement and duration selectors remain source-specific. |
| 36 | US filter metadata | Live GMS `/US/filters` | `(generated endpoint) us filters` | Makes the global research universe discoverable. |
| 37 | Homepage event/publication stream | Live `analyze-api /v2/homepage/events` | `(generated endpoint) homepage events` | Adds count/offset/sids/type controls while leaving unresolved enums explicit. |
| 38 | Provenance/access/freshness envelope | Discovery report plus all live routes | `(behavior in tickertape-pp-cli) every response records source, route, observed_at, entity, and access metadata` | Agents can distinguish public, premium-locked, authenticated, unavailable, and stale-free results. |

## Explicitly not absorbed

- Portfolio holdings, diversification, red flags, forecast views, watchlists, saved screens, custom universes, user baskets, and personalized feeds: sensitive/session-owned and outside approved first build.
- Login, OTP, CAPTCHA, broker connect, order placement, trade cancellation, payments, exports, social posting, notifications, loans, gold, subscriptions, and account opening: mutation or financially consequential workflows.
- Real-time Socket.IO/WebSocket transport: REST snapshots are sufficient and avoid a resident process.
- Screener POST execution: deferred until a safe fixture validates body schema, pagination, and premium error behavior.
- Universal cross-source score: Tickertape scorecards remain source-specific opinions.

## Transcendence features

| # | Feature | Command | Buildability | Why only this CLI can do it | Long Description |
|---:|---|---|---|---|---|
| 1 | Entitlement-aware scorecards | `company scorecard` | spec-emits | Makes locked, premium, and public score elements first-class terminal data instead of silently hiding them. | none |
| 2 | Company publication timeline | `company timeline` | hand-code | Combines summary events, news, and investor-presentations into a dated source-labelled view without storing a local history. | Use this for a live company timeline; do not confuse it with a persisted personal alert history. |
| 3 | Live multi-asset lookup | `lookup` | hand-code | Gives agents one entry point across Indian stocks, MFs, ETFs, indices, and US securities while retaining source-specific fields. | Use `lookup` for a quick identity/quote triage across asset classes. |
| 4 | Access and provenance envelope | `inspect access` | hand-code | Makes source host, route, observed time, lock state, and entitlement failures explicit in every response. | Use this to explain whether a field is public, locked, authenticated, or unavailable. |
| 5 | MMI component brief | `market brief` | hand-code | Turns MMI's heterogeneous component payload into a concise, timestamped research brief without persisting the snapshot. | Use this for current market mood context, not a trading signal or historical series. |
| 6 | Source-specific research brief | `company brief` | hand-code | Combines live company identity, scorecard, summary, and public reports with explicit access state in one agent-sized response. | Use this for a first-pass company triage; use individual commands for full raw payloads. |
| 7 | Live screener catalog navigator | `screen catalog` | hand-code | Organizes equity and MF filters/prebuilt screens while preserving premium markers and avoiding saved-screen state. | Use this to discover available live filters and presets, not to persist or execute a personal screen. |

## Gate decision

Approved implementation scope is the public read-only surface plus entitlement-aware route envelopes. Hand-code features above are committed after generation; no feature may be silently downgraded to a stub except the explicitly marked screener-query contract. Domain data remains live-only. Secure auth persistence is allowed only through the generator's approved credential mechanism if implemented, never through domain-data storage or build artifacts.
