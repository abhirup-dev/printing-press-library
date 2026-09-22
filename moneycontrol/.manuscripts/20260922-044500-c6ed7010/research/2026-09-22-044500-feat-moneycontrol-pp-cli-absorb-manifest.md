# Moneycontrol absorb manifest

## Source tools and evidence

| Tool/source | Evidence | Features absorbed |
|---|---|---|
| Moneycontrol public site and current EgoBrowser capture | Discovery report plus live TaskSpace 7 validation | Public financial news, article pages, stock/index pages, event/research surfaces, Pro/account GET shapes |
| Historical PR #1701 / `abhirup-dev:feat/moneycontrol` | Closed/unmerged fork inspected read-only | Prior news, quotes, compound command concepts, priceapi-specific transport workaround |
| `pramodhapple504-arch/moneycontrol-mcp` | Public GitHub source/README | Symbol search, quote, fundamentals, index, FII/DII, technical pivots, history, news; public endpoint families and actionable errors |
| PyPI `moneycontrol-api` | Web research | Latest/business/news wrapper concepts |
| GitHub `D4T4R/MCFinEx` and related scrapers | Web research | Earnings/financial-table awareness and client-rendered fragility |

## Absorbed (match or beat existing tools)

| # | Feature | Best Source | Our Implementation | Added Value |
|---:|---|---|---|---|
| 1 | Latest financial news | Moneycontrol site, PR #1701, PyPI wrapper | `(generated endpoint) news latest` | Live HTML extraction; financial-only filtering and explicit empty/blocked metadata |
| 2 | Market news | Moneycontrol site | `(generated endpoint) news market` | Live market-news category with URLs/timestamps |
| 3 | Stock/company news | Moneycontrol site, MCP | `(generated endpoint) news stocks` | Live stock-news category and structured links |
| 4 | Economy/company/results news | Moneycontrol site | `(generated endpoint) news economy|companies|earnings` | Broad financial categories; non-financial categories excluded |
| 5 | IPO news | Moneycontrol site, MCP | `(generated endpoint) news ipo` | Live IPO headlines and article links |
| 6 | Mutual-fund news | Moneycontrol site, MCP | `(generated endpoint) news mutual-funds` | Live MF headlines without local storage |
| 7 | Commodity news | Moneycontrol site | `(generated endpoint) news commodities` | Live commodity-market headlines |
| 8 | Exact article extraction | Moneycontrol site, PR #1701 | `(generated endpoint) article get` | Extracts title, timestamps, canonical URL and `#contentdata.content_wrapper`; preserves raw source link |
| 9 | Company-tag news | Moneycontrol site, PR #1701 | `(generated endpoint) news for` | Live tag pages for supplied slugs/SC IDs |
| 10 | Symbol/stock lookup | MCP, prior source | `(generated endpoint) symbols search` | Resolves SC ID, ticker and tag slug; actionable no-match errors |
| 11 | Index snapshots | Moneycontrol site, MCP, PR #1701 | `(generated endpoint) quote index` | Dedicated `priceapi` host and encoded `%3B` keys |
| 12 | Stock snapshots | Moneycontrol site, MCP, PR #1701 | `(generated endpoint) quote stock` | Quote, change, OHLC, volume, 52-week range and market state |
| 13 | Price-volume | Current API observation | `(generated endpoint) quote price-volume` | Current JSON window with response-shape validation |
| 14 | Stock historical OHLCV | MCP | `(generated endpoint) history stock` | Live techCharts route with interval/countback controls |
| 15 | Fundamentals and returns | MCP, stock page | `(generated endpoint) research fundamentals` | Valuation, ratios, sector and returns where current response is usable |
| 16 | Technical pivots | MCP, stock page | `(generated endpoint) research technicals` | Daily/weekly/monthly pivots and support/resistance |
| 17 | FII/DII activity | MCP, site | `(generated endpoint) markets fii-dii` | Cash/F&O institutional activity with source metadata |
| 18 | Company financials | Site stock page, discovery | `(generated endpoint) research financials` | Financial history/overview as explicit best-effort |
| 19 | Forecasts and recommendations | Site stock page, discovery | `(generated endpoint) research forecasts` | Estimates, analyst ratings and recommendation data as explicit best-effort |
| 20 | Fundamental screens | Site stock scanner | `(generated endpoint) screens fundamental` | Live scanner pages; no silent success on empty widgets |
| 21 | Technical screens | Site stock scanner | `(generated endpoint) screens technical` | Live scanner pages; stable/fragile status exposed |
| 22 | Corporate actions | Site corporate-action pages | `(generated endpoint) events corporate-actions` | Live actions for supplied company or market scope |
| 23 | Filings and results | Site filings/earnings pages | `(generated endpoint) events filings|results` | Live result/filing context with article/page URLs |
| 24 | IPO surfaces | Site IPO pages | `(generated endpoint) markets ipo` | IPO news/listing/research pages as best-effort |
| 25 | F&O and open interest | Site F&O pages | `(generated endpoint) markets fno` | F&O expiries, market stats and OI pages as best-effort |
| 26 | Commodities, gold and silver | Site commodity pages | `(generated endpoint) markets commodities` | Live commodity pages; contract-specific gaps are reported |
| 27 | Mutual funds | Site MF pages | `(generated endpoint) markets mutual-funds` | Live MF news/quote pages; no fund database |
| 28 | Currency/forex | Site currency pages | `(generated endpoint) markets forex` | Live currency pages; availability metadata for fragile routes |
| 29 | Compound market wrap | PR #1701, brief | `moneycontrol market-wrap` | Live index + quotes + financial headlines in one source-limited response |
| 30 | Stock watch | PR #1701, brief | `moneycontrol stock-watch` | Live supplied-stock quote/news/event context |
| 31 | News digest | PR #1701, brief | `moneycontrol news-digest` | Live category/tag article digest with no local history |
| 32 | Live `news-for` filter | PR #1701, brief | `moneycontrol news-for` | Filters current live responses only; no local `since` cursor |
| 33 | Portfolio/watchlist/alert reads | TaskSpace 7, user scope | `moneycontrol portfolio|watchlist|alerts list` | Optional endpoint-specific auth; read-only, secure-store-backed, no data persistence |
| 34 | Super Pro/account insights | TaskSpace 7 page traffic | `moneycontrol account entitlement|pro insights` | Optional auth; page-generated account/Pro GETs only, never infer from header text |
| 35 | Availability/stability metadata | Current discovery and user scope | `(behavior in moneycontrol-pp-cli <every command>) status metadata` | Every command reports source, HTTP status, shape/empty/blocked state and actionable error |
| 36 | Dedicated priceapi client | PR #1701 patch and live capture | `(behavior in moneycontrol-pp-cli transport)` | Keeps `priceapi.moneycontrol.com` separate from Akamai-sensitive `www` client |
| 37 | Encoded index keys | Live capture | `(behavior in moneycontrol-pp-cli quote index)` | Preserves `%3B` URL encoding and documents key mapping |
| 38 | Read-only output contracts | MCP and Printing Press conventions | `(behavior in moneycontrol-pp-cli <commands>) --json/--agent/--select` | Agent-shaped output without cache or local database |
| 39 | Auth bootstrap/refresh | Current user-approved scope | `moneycontrol auth login --chrome` | Uses established secure credential/session store; values never enter source/logs/output |
| 40 | Financial-only boundary | User scope | `(behavior in moneycontrol-pp-cli news and markets commands) source filter` | Rejects politics, entertainment, lifestyle and non-financial sections explicitly |

## Transcendence (only possible with our approach)

| # | Feature | Command | Buildability | Why Only We Can Do This | Long Description |
|---:|---|---|---|---|---|
| 1 | Holding context join | `context holdings` | hand-code | Joins live broker-supplied SC IDs to current quotes, tag news and event pages without persisting portfolio data. | none |
| 2 | Event-risk board | `events upcoming` | hand-code | Presents live filings, results and corporate actions together for a supplied holding set. | none |
| 3 | Market breadth brief | `market breadth` | hand-code | Combines live encoded index prices with the stable change-table fragment while preserving partial availability. | none |
| 4 | Catalyst matrix | `catalysts` | hand-code | Filters live dated IPO, earnings, filing and corporate-action pages mechanically by symbol and time window. | none |
| 5 | Company news timeline | `news timeline` | hand-code | Normalizes current company-tag HTML into URL-preserving chronological rows. | none |
| 6 | Portfolio news triage | `portfolio triage` | hand-code | Produces a mechanical, pipe-friendly live review set for external holdings without sentiment or storage. | none |
| 7 | Research packet | `research packet` | hand-code | Assembles quote, price-volume, news and event context into one source-linked live packet. | none |

## Approved scope and explicit exclusions

- All Moneycontrol data is live, online, on-demand and source-limited.
- No response cache, article DB, SQLite, sync/history, offline index, stored data cursor, daemon or browser sidecar.
- Authenticated reads are optional per command; approved credentials/session metadata may use an established secure store, never source/logs/reports/output.
- No portfolio/watchlist/alert mutation, order, trade, payment or other transaction.
- No politics, entertainment, lifestyle or non-financial news.
- Fragile/Akamai-sensitive routes are best-effort and must expose availability/stability/access metadata; empty, blocked or shape-changed responses are errors, not silent success.
- Prior `since` is not a local-history command. A live time filter is permitted only when computed from the current response.
