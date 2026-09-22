# Moneycontrol discovery report

**Target:** Moneycontrol  
**Canonical URL:** <https://www.moneycontrol.com/>  
**Discovery date:** 2026-09-21  
**Mode:** discovery only; no CLI generation, build, publish, PR, or library mutation

## Executive recommendation

Moneycontrol should be the **market-intelligence and news context layer** in a combined personal investing CLI—not the brokerage, portfolio ledger, or execution layer.

Recommended product shape:

```text
Kite/Zerodha holdings + positions
        │ symbol mapping
        ▼
Moneycontrol: market news, company-tag news, article bodies,
indices, price/volume context, corporate events and commentary
        │
        ▼
portfolio-news / market-wrap / event-risk summaries
```

The strongest initial value is a read-only, public surface:

1. latest Indian market and stock news;
2. exact article extraction;
3. company/tag news for a supplied watchlist or holding set;
4. SENSEX/NIFTY and per-stock snapshot context;
5. a combined market wrap.

**Architecture constraint:** the eventual CLI must be stateless, online, on-demand and source-limited for Moneycontrol data. Every command queries Moneycontrol live and returns the result. No response cache, article database, sync history, offline index, stored data cursor, background daemon or resident browser sidecar is permitted. If an authenticated feature is later approved, credentials/cookies/tokens/session metadata may be securely persisted or refreshed only through an established secure store; they must never appear in source, logs, reports or command output. A portfolio CLI may pass current holdings or symbols into Moneycontrol commands, but Moneycontrol must not become a persisted portfolio ledger.

The site exposes much more—financials, technicals, forecasts, screeners, corporate actions, F&O, commodities, mutual funds, alerts and portfolio tools—but the deeper JSON routes are more parameter-sensitive and/or subscription- or session-dependent. They should be ranked at the checkpoint, not silently included in an MVP.

## Evidence status

| Label | Evidence |
|---|---|
| **Observed now** | Live read-only exploration in exactly one EgoBrowser TaskSpace, `Moneycontrol` (space 6), page `p1`. Public news, stock, index, scanner, corporate-action, filings, mutual-fund, commodities, gold/silver, IPO and F&O pages were visited. Representative requests and response shapes were sampled from that page context. |
| **Recovered from CASS** | Prior session `6ccea8ce-9a44-4f79-a0f9-10685fe6c9ff`; prior Printing Press run and CLI scope; prior Akamai/pricefeed findings; prior 101/101 dogfood claim; prior compound commands; prior PR/patch evidence. |
| **Recovered and rechecked externally** | GitHub PR `https://github.com/mvanhorn/printing-press-library/pull/1701` was queried read-only. Current state is **CLOSED**, not merged. Its fork branch is `abhirup-dev/printing-press-library`, `feat/moneycontrol`. |
| **Inferred** | Competitor positioning and the recommended combined-CLI role. These are product recommendations, not Moneycontrol guarantees. |
| **Inaccessible/auth-gated** | Personal portfolio, watchlist, alerts and profile payloads were not opened or replayed. The live browser showed an authenticated greeting and `Logout`, but no personal data or auth values were copied. Some PRO and dynamic research APIs returned only `success:0` when called without the exact page-generated parameters. |

No cookies, token values, session headers, account settings or financial actions were touched. The initial home navigation briefly landed on an ad interstitial; all subsequent exploration used direct public pages in the same TaskSpace. EgoBrowser showed an update notice; it was not run.

## User-visible feature inventory

### High-priority financial/news surfaces

Observed in the live site navigation and page content:

- **Markets/news navigation:** Markets, Business, Markets, Stocks, Economy, Companies, IPO, Earnings, Commodities, Mutual Funds, Indian Indices, US Markets, F&O, Open Interest Trends, Technical Trends, All Stats and Unlisted Shares.
- **Latest news:** `/news/latest-news/` rendered a `Latest News` page with article titles, timestamps and canonical article links.
- **Market news:** `/news/business/markets/` rendered a `Markets` page with market setup, stocks-to-watch and market commentary.
- **Stock news:** `/news/business/stocks/` rendered `Stock Market Today` with stocks-to-watch, recommendations, closing-bell coverage and company developments.
- **Stock/topic tags:** `/news/tags/reliance-industries.html` rendered a Reliance Industries page with company description and tagged article history.
- **Exact article pages:** article URLs follow a category/slug/numeric-ID pattern. Current article content was visible in `#contentdata.content_wrapper`; the page also exposed canonical URL and JSON-LD metadata.
- **Indian indices:** homepage and `/markets/indian-indices/` showed SENSEX, NIFTY 50, NIFTY BANK and NIFTY Midcap 100 with price, change and percentage change. Current observed homepage values were SENSEX `74858.99`, NIFTY 50 `23414.30`, NIFTY BANK `56470.65`, and Midcap 100 `62013.10`; these are evidence of shape only, not durable data.
- **Company/stock pages:** a Reliance page exposed price/volume, 52-week range, technical rating, forecast labels, earnings forecast, financials, ROE, shareholding (promoter/FII/DII), corporate action, peers, company information, news, and F&O links.
- **Technical scanners:** `/stock-scanner/technical/` exposed technical scanner families such as intraday/supertrend, moving-average, price scans, volume/delivery and candlestick patterns.
- **Fundamental scanners:** `/stock-scanner/fundamental/` exposed fundamental scanner navigation; stock pages linked to ratio scans, improving ROE/ROCE, profit-and-loss scans and MC-curated scans.
- **Market action:** homepage showed a Market Action panel with index breadth (advancers/decliners) and stock-action affordances. The underlying stockaction widgets were not CLI-reliable in current probing (see constraints).
- **Trending stocks:** a `Trending Stocks Today` widget is present and returns HTML in a same-site browser fetch; direct navigation can produce an empty body, so it needs replay validation.
- **Corporate actions:** `/markets/corporate-action/` is a public page titled `CORPORATE ACTION`.
- **Corporate filings:** `/markets/corporate-filings/` is a public page titled `Corporate Filings`.
- **Earnings/results news:** `/news/business/earnings/` is a public earnings-news page with trending result topics.
- **IPO news:** `/news/business/ipo/` is a public IPO-news page with subscription/status topics.
- **F&O:** public futures-gainer pages exist under `/stocks/fno/marketstats/...`; the site exposes futures/options expiry-data routes from stock pages.

### Finance-adjacent surfaces to rank, not automatically include

- **Commodities:** `/news/business/commodities/` and dedicated gold/silver pages.
- **Gold/silver rates:** `/news/gold-rates-today/` and `/news/silver-rates-today/` rendered current city/commodity prices.
- **Mutual funds:** `/news/business/mutual-funds/` is public news; the navigation also exposes MF prices, investment and watchlist tools.
- **Currencies/forex:** present in the Markets navigation and Watchlist taxonomy; no dedicated replayable route was promoted during this pass.
- **Portfolio/watchlist:** navigation exposes My Portfolio, My Watchlist, My Alerts, Price Alerts, MF investment and MF prices.
- **Forecasts/recommendations:** stock pages show forecast/analyst/recommendation labels, including PRO teasers and Buy/Hold/Sell sections.
- **Big Shark Portfolios, seasonality, economic calendar and open-interest trends:** exposed by the markets navigation.

### Explicitly deprioritized content

General India/politics, technology, entertainment, lifestyle, travel and generic news are present on Moneycontrol but are not part of the recommended financial CLI scope unless needed to understand a shared article taxonomy.

## Network/API inventory

All methods below are **GET** unless marked otherwise. Query values are intentionally redacted. `SC_ID` is Moneycontrol's public internal stock identifier (for example `RI` for Reliance); it is not a credential.

| Host | Method | Sanitized route pattern | Response / identifiers / pagination | Auth and reachability | Confidence |
|---|---:|---|---|---|---|
| `www.moneycontrol.com` | GET | `/news/latest-news/` | SSR/HTML listing; article links contain numeric IDs; page navigation is HTML-based rather than an observed cursor | Public in browser. Raw HTTP has historical Akamai sensitivity | High; observed now |
| `www.moneycontrol.com` | GET | `/news/business/markets/` | SSR/HTML market-news listing; category path is the selector | Public | High; observed now |
| `www.moneycontrol.com` | GET | `/news/business/stocks/` | SSR/HTML stock-news listing | Public | High; observed now |
| `www.moneycontrol.com` | GET | `/news/tags/{slug}.html` | SSR/HTML stock/topic feed; `{slug}` is the tag identity | Public; large HTML response | High; observed now |
| `www.moneycontrol.com` | GET | `/news/{category}/{slug}-{article_id}.html` | SSR/HTML article; body in `#contentdata.content_wrapper`; canonical and JSON-LD metadata available; `{article_id}` is numeric | Public | High; observed now |
| `www.moneycontrol.com` | GET | `/techmvc/mc_widgets/trending_stocks?limit=<n>&classic=<flag>` | HTML widget; current same-site browser fetch returned 200 and ~2.6 KB; direct navigation returned empty in one pass | Public but replay/referrer-sensitive | High; observed now |
| `www.moneycontrol.com` | GET | `/newsapi/mc_news.php?query=<term>&start=<offset>&limit=<n>&sortby=<field>&sortorder=<dir>&classic=<flag>` | JSON; current safe probe returned 200 with a tiny `length` object and no rows for the chosen query. `start`/`limit` imply offset pagination | Public, but query contract needs further validation | Medium; observed network + safe probe |
| `www.moneycontrol.com` | POST | `/mccode/common/autosuggesion.php?query=<term>&type=<n>&section=<name>` | HTML suggestions; current homepage JS constructs this as POST. Current browser probes returned 404, so treat as legacy/broken until revalidated | Public route construction observed; currently inaccessible | Medium; observed JS, current 404 |
| `priceapi.moneycontrol.com` | GET | `/pricefeed/nse/equitycash/{SC_ID}` | JSON envelope `{code,message,data}`; stock data includes price, high/low, period changes, CAGR, volume and sector-like fields; no pagination | Public; current browser navigation returned 200 | High; observed now |
| `priceapi.moneycontrol.com` | GET | `/pricefeed/bse/equitycash/{SC_ID}` | Same family for BSE quote | Public; observed as a live page resource | Medium; observed network |
| `priceapi.moneycontrol.com` | GET | `/pricefeed/notapplicable/inidicesindia/{index_key}` | JSON envelope `{code,message,data}`; index data includes high/low, period changes, exchange and percentage change. Index keys require URL-encoded semicolon, e.g. `in%3BSEN`, not a raw `;` | Public; current `in%3BSEN` and `in%3BNSX` returned 200. Raw semicolon is a known failure mode | High; observed now + CASS |
| `priceapi.moneycontrol.com` | GET | `/techCharts/indianMarket/stock/config` | Chart configuration JSON/resource | Public page resource; exact replay shape not captured | Medium; observed network |
| `priceapi.moneycontrol.com` | GET | `/techCharts/indianMarket/stock/symbols?symbol=<symbol>` | Chart-symbol lookup JSON/resource | Public page resource; symbol value is page-generated | Medium; observed network |
| `priceapi.moneycontrol.com` | GET | `/techCharts/indianMarket/stock/history?symbol=<symbol>&resolution=<r>&from=<ts>&to=<ts>&countback=<n>&currencyCode=<ccy>` | Chart time-series route; time window and resolution are explicit; likely JSON | Public page resource; not independently replayed | Medium; observed network |
| `priceapi.moneycontrol.com` | GET | `/techCharts/indianMarket/stock/marks?symbol=<symbol>&from=<ts>&to=<ts>&resolution=<r>` | Chart markers/corporate-event route; likely JSON | Public page resource; not independently replayed | Medium; observed network |
| `api.moneycontrol.com` | GET | `/mcapi/v1/stock/price-volume?scId=<SC_ID>&ex=<exchange>&appVersion=<v>` | JSON `{success,data.stock_price_volume_data.{price,volume}}`; fixed keys include 1 Week/1 Month/3 Months/YTD/1 Year/3 Years and Today/Yesterday/1 Week Avg/1 Month Avg volume objects; no observed pagination | Public; current `scId=RI` probe returned 200 JSON | High; observed now |
| `api.moneycontrol.com` | GET | `/mcapi/technicals/v2/details?scId=<SC_ID>&dur=<duration>&deviceType=<device>` | JSON route observed from the stock page; guessed direct calls returned `{success:0,data:...}` because exact dynamic parameters were missing | Browser page resource; not yet CLI-reliable | Medium |
| `api.moneycontrol.com` | GET | `/mcapi/v1/stock/financial-historical/overview?scId=<SC_ID>&ex=<exchange>` | JSON route for financial history/overview; guessed direct call returned `success:0` | Browser page resource; parameter contract incomplete | Medium |
| `api.moneycontrol.com` | GET | `/mcapi/v1/stock/estimates/{check-forecast,price-forecast,consensus,analyst-rating,earning-forecast,valuation,hits-misses}` | JSON estimate/forecast/recommendation family. Query keys include `deviceType`, `scId`, `ex`, and sometimes `financialType`, `frequency`, `type`, `appVersion` | Likely public/PRO-mixed; guessed direct calls returned `success:0`; do not assume free access | Medium |
| `api.moneycontrol.com` | GET | `/marketinsights/v1/markets/get-bse-reports?type=<type>&scId=<SC_ID>&limit=<n>` | JSON market/company report family; explicit `limit` | Browser resource; exact type values not captured | Medium |
| `api.moneycontrol.com` | GET | `/mcapi/extdata/v2/mc-insights?scId=<SC_ID>&type=<type>&deviceType=<device>&appVersion=<v>` | JSON/insight data family | Browser resource; likely entitlement-sensitive | Medium |
| `www.moneycontrol.com` | GET | `/mc/widget/mcinsightspro/insightPro?classic=<flag>&sc_did=<id>&sc_id=<SC_ID>` | PRO insight widget; HTML or embedded data | Subscription/entitlement risk | Medium |
| `www.moneycontrol.com` | GET | `/mc/widget/stockvitals?classic=<flag>&sc_did=<id>&sc_id=<SC_ID>&slug=<slug>` | HTML; current probe returned 200 and ~423 bytes with Altman Z, DuPont/ROE and Graham-number snippets | Public teaser/widget; replayable HTML observed | High; observed now |
| `www.moneycontrol.com` | GET | `/mc/widget/stockscans?classic=<flag>&sc_did=<id>&sc_id=<SC_ID>&isin_id=<ISIN>` | HTML; current probe returned 200 and ~2.4 KB with technical/fundamental scan labels | Public widget; replayable HTML observed | High; observed now |
| `www.moneycontrol.com` | GET | `/stocks/company_info/get_vwap_chart_data.php?classic=<flag>&sc_did=<id>` | Chart data route | Public page resource; exact response not separately captured | Medium |
| `www.moneycontrol.com` | GET | `/mc/widget/stockdetails/getChartInfo?classic=<flag>&scId=<SC_ID>&type=<type>` | Chart metadata/data route | Public page resource; exact response not separately captured | Medium |
| `api.moneycontrol.com` | GET | `/swiftapi/v1/returns/stock?scId=<SC_ID>&type=<type>&exchange=<exchange>&responseType=<format>` | JSON returns family; guessed direct call returned `success:0` | Browser resource; parameter contract incomplete | Medium |
| `api.moneycontrol.com` | GET | `/mcapi/v1/fno/futures/getExpDts?id=<id>` and `/mcapi/v1/fno/options/getExpDts?id=<id>` | Expiry-date JSON routes; `{id}` is page-generated contract/stock identity | Public page resource; exact replay not validated | Medium |
| `www.moneycontrol.com` | GET | `/markets/indian-indices/changeTableData?deviceType=<device>&exName=<exchange>&indicesID=<id>&selTab=<tab>&subTabOT=<x>&subTabOPL=<x>&selPage=<page>&classic=<flag>` | HTML table fragment; current safe probe returned 200 and ~1.3 KB. `selPage` gives page-style pagination | Public | High; observed now |
| `www.moneycontrol.com` | GET | `/mc/widget/stockaction/{topGainers,topLosers,fiftyTwoWeekHigh,fiftyTwoWeekLow,onlyBuyers,onlySellers,priceShockers,volumeShockers,mostActive}` | Intended market-action HTML widgets; current `topGainers` probe returned 200 with zero-byte body | Browser/UI-only or anti-bot sensitive; exclude from MVP unless a stable replay is found | High; observed now + CASS |
| `www.moneycontrol.com` | GET | `/verloop-apis/overall/user-details/?classic=<flag>&token=<redacted>` | Auth/user-detail response; token value must never be captured | Logged-in browser only; do not replay in discovery | High; observed network, values redacted |
| `www.moneycontrol.com` | GET | `/monitoring/mc_user_entitlements.php?classic=<flag>&user_token=<redacted>` | Entitlement/feature gating | Logged-in browser only; do not replay | High; observed network, values redacted |
| `accounts.moneycontrol.com` | GET | `/user/getdetails` | Account details | Authenticated; do not inspect | High; observed network |

### Representative sanitized requests

```text
GET https://priceapi.moneycontrol.com/pricefeed/nse/equitycash/RI
GET https://priceapi.moneycontrol.com/pricefeed/notapplicable/inidicesindia/in%3BSEN
GET https://api.moneycontrol.com/mcapi/v1/stock/price-volume?scId=RI
GET https://www.moneycontrol.com/news/tags/reliance-industries.html
GET https://www.moneycontrol.com/news/<category>/<slug>-<article_id>.html
GET https://www.moneycontrol.com/markets/indian-indices/changeTableData?deviceType=<device>&exName=<exchange>&indicesID=<id>&selPage=<page>
POST https://www.moneycontrol.com/mccode/common/autosuggesion.php?query=<term>&type=<n>&section=<name>
```

## Public vs authenticated capability matrix

| Capability | Public/read-only status | Recommendation |
|---|---|---|
| Latest, market and stock news | Public HTML; works in browser now | MVP |
| Article body and metadata | Public HTML; body selector confirmed | MVP |
| Stock-tag news | Public HTML; Reliance tag confirmed | MVP for holdings/watchlists supplied by another source |
| SENSEX/NIFTY/index quotes | Public priceapi JSON; 200 now | MVP |
| NSE/BSE stock quote | Public priceapi JSON; 200 now | MVP |
| Price/volume summary | Public API route; 200 JSON now | MVP/later depending freshness contract |
| Trending stocks | Public browser widget; same-site fetch works, direct navigation can be empty | Later or best-effort |
| Stock vitals/scans | Public HTML widgets observed | Later; add only with replay tests |
| Corporate actions/filings/results/IPO pages | Public HTML pages observed | Later, high-value |
| Technical/fundamental scanners | Public pages and scanner links observed; deeper APIs not fully replayed | Later, validate exact contracts first |
| Forecasts, analyst ratings and PRO insights | Page labels/public teasers; API calls are parameter/entitlement-sensitive | Later or explicitly PRO-only |
| Gold/silver/commodity news and rates | Public pages observed | Later, separate finance-adjacent module |
| Mutual-fund news and public fund pages | Public navigation/news observed | Later; do not assume portfolio/MF account data is public |
| F&O expiry and market-stat pages | Public pages/routes observed | Later; useful for an advanced market module |
| My Portfolio / My Watchlist / My Alerts / Price Alerts | Navigation exists; current session showed authenticated greeting and logout. Personal data was not opened | Defer behind explicit auth/privacy decision |
| Account details, followed-list and entitlement routes | Authenticated browser resources observed | Out of MVP; if later approved, use live requests with an established secure store and never emit or log credentials |
| Create/update/delete alerts, portfolio mutations or trades | Not explored; would be write-risk | Explicitly out of scope for this source |

## Candidate command tree

### MVP

```text
moneycontrol news latest [--limit N]
moneycontrol news market [--limit N]
moneycontrol news stocks [--limit N]
moneycontrol news for --sc-ids RI,INF,ITC
moneycontrol article get --url <article-url>
moneycontrol quote stock --sc-id RI
moneycontrol quote index --key sensex|nifty|bank|it
moneycontrol company snapshot --sc-id RI
moneycontrol market wrap [--json]
moneycontrol trending list                 # best-effort, only if replay remains stable
```

Design intent: keep the source read-only, return structured JSON, preserve article URLs/timestamps, and make `market wrap` combine indices + reliable quote context + market/stock headlines. `company snapshot` should use the pricefeed and price-volume routes first; avoid silently promising the deeper forecast APIs.

### Later

```text
moneycontrol financials overview --sc-id RI
moneycontrol technicals details|chart --sc-id RI
moneycontrol forecasts consensus|price|earnings|ratings --sc-id RI
moneycontrol screens technical|fundamental [--scan <name>]
moneycontrol corporate-actions list [--sc-id RI]
moneycontrol filings list [--sc-id RI]
moneycontrol fno expiries|market-stats
moneycontrol commodities gold|silver|rates
moneycontrol mutual-funds news|quote
moneycontrol portfolio snapshot       # only after explicit auth decision
moneycontrol watchlist snapshot       # only after explicit auth decision
moneycontrol alerts list              # read-only, only after explicit auth decision
```

### Explicitly out of scope

- local SQLite/database state, response caches, article DBs, sync jobs, sync history, offline indexes or stored data cursors;
- background daemons, scheduled refreshers or resident browser sidecars;
- order placement, trading, broker execution or any financial transaction;
- create/update/delete portfolio entries, watchlists or alerts;
- ad hoc capture or export of browser cookies, session tokens or user-detail payloads; approved authentication may use an established secure store, but never source/logs/reports/output;
- resident-browser transport as the normal CLI runtime;
- PRO/paywalled material unless the user explicitly approves an authenticated design;
- live tick-by-tick streaming (the confirmed quote route is a snapshot, not a websocket);
- generic politics, entertainment, lifestyle and non-financial news;
- the empty/unreliable stockaction widgets unless a stable replay surface is found.

## Overlap and unique contribution

| Source | Strong overlap | Moneycontrol's unique contribution in the combined CLI |
|---|---|---|
| **Kite/Zerodha** | Quotes, instruments, historical candles, holdings/positions and portfolio context. Kite Connect documents authenticated holdings, positions, quotes, historical candles and live streaming. | No execution role. Adds mainstream Indian market news, exact article bodies, company/tag news, market commentary, index narrative and event context around broker holdings. |
| **Tijori** | Fundamentals, statements, shareholding and advanced stock screening. Tijori is stronger for structured research and compound financial filters, with some deeper data behind Premium. | Fresh news volume, event-driven stories, stock-tag feeds, index/market narrative, analyst/recommendation prose and public corporate-news context. |
| **Tickertape** | Stocks, ETFs, mutual funds, screeners, portfolios, watchlists, saved screens and consumer research workflows. | Stronger newsroom/news-digest contribution: article extraction, market/stock headlines, company-tag news and a market wrap that can be joined to an external holdings list. |
| **Moneycontrol itself** | The website is the source of truth for its own pages and account tools. | A CLI can make the public data composable: `holdings -> symbols -> tagged news`, stateless live JSON joins and a one-command market wrap. |

**Combined-CLI thesis:** Kite/Zerodha supplies what the user owns and can trade; Tijori/Tickertape supply structured research/screening; Moneycontrol supplies the narrative and event stream explaining what moved and why. Moneycontrol should not compete with the broker's execution or be treated as a canonical portfolio ledger.

## Auth, anti-bot and risk classification

### Reachability

- No official public Moneycontrol developer portal or API-key flow was found in web research. The prior run and current site both indicate reverse-engineered web routes.
- `priceapi.moneycontrol.com` is the cleanest current surface: stock and encoded index requests returned HTTP 200 JSON in the browser.
- `www.moneycontrol.com` is protected by Akamai behavior. Prior CASS evidence recorded raw HTTP 403/empty-body behavior and a Printing Press probe that sometimes cleared it. Current browser context can load SSR pages, but individual widgets differ.
- Current browser fetch observations: article/tag pages returned large HTML; `trending_stocks` returned HTML; `stockaction/topGainers` returned HTTP 200 with an empty body; direct navigation to the trending widget also produced an empty body in one pass.
- Prior run evidence reports transient priceapi 503s after repeated probing and a generated-command fix: use a priceapi-specific client rather than the default `www` client for indices/stocks. Treat this as a major implementation constraint, not a solved guarantee.
- The encoded semicolon in index keys (`%3B`) is mandatory in the working route shape; a raw semicolon was parsed incorrectly by the edge/path handling.
- No CAPTCHA or consent challenge was required after direct navigation in this pass. The first home route included an ad interstitial.

### Read/write classification

| Action | Risk |
|---|---|
| GET public news, quote, index, chart, screen and rate pages | Read-only / low risk |
| POST autosuggest search | Read-only intent, but currently 404/unusable |
| GET logged-in user details, followed-list, entitlements | Read-only transport, **high privacy risk**; not part of the public MVP. If later approved, use live requests with an established secure store and never emit or log credentials |
| Portfolio/watchlist/alert mutation | Write/financial-account risk; not explored |
| Orders, trades, payments | High financial risk; explicitly excluded |

## Reuse opportunities from old CLI and sessions

### Prior CASS session

Session `6ccea8ce-9a44-4f79-a0f9-10685fe6c9ff` produced the prior Moneycontrol run. It reported:

- news latest/category/tag/article extraction;
- SENSEX/NIFTY and stock pricefeed quotes;
- trending stocks;
- compound commands `market-wrap`, `stock-watch`, `news digest`, `news-for` and `since`; the prior `since` implementation depended on locally synced articles and is rejected, although a future live endpoint/time-filter variant could be considered if it uses only the current response;
- removal of `screen scan` after `/mc/widget/stockaction/*` returned empty bodies;
- a 101/101 dogfood claim and live verification claim.

Treat all of that as **recovered evidence**, not current truth. This report revalidated the core news, pricefeed and price-volume paths and re-observed the widget instability.

### PR #1701 / fork

Read-only GitHub inspection found:

- PR: <https://github.com/mvanhorn/printing-press-library/pull/1701>
- Title: `feat(moneycontrol): add moneycontrol`
- Current state: **CLOSED**, `mergedAt: null`; it is not currently in the upstream main branch.
- Head: `abhirup-dev/printing-press-library`, branch `feat/moneycontrol`.
- Fork contains `library/other/moneycontrol/`, `spec.yaml`, README, SKILL, generated commands, tests and 100 files.
- The prior implementation already has the five compound commands and a local SQLite/search model. Reuse only the stateless live-request and parsing portions; reject its SQLite, response-cache, sync, history, stored-data-cursor, local-search and locally-synced-article concepts. Auth, if ever approved, must use an established secure store rather than ad hoc session capture.
- `.printing-press-patches/priceapi-client-override.md` records the important Akamai workaround: generated indices/stocks commands must use a `priceapi.moneycontrol.com`-specific client instead of the default `www` client.

This is a strong starting artifact if the user chooses reuse, but it needs a current reachability pass and a scope decision before any reprint or implementation. No local `~/printing-press/library/moneycontrol` directory was found in this session; the fork branch is the durable source currently available.

## Recommended checkpoint questions

1. **MVP boundary:** Should v1 be public news + article extraction + pricefeed/index quotes + `market-wrap`/`stock-watch`, or should technical/fundamental screens and corporate actions be first-class in v1 despite their more fragile replay contracts?
2. **Authenticated Moneycontrol data:** Should portfolio/watchlist/alerts remain out of scope? If later approved, use live requests with credentials from an established secure store; do not capture an ad hoc browser session or persist Moneycontrol data. I recommend defer.
3. **Runtime contract:** The CLI is fixed to stateless, online, on-demand HTTP/HTML replay with clear per-command reachability errors. Which public surfaces are worth accepting as best-effort under Akamai-compatible request shaping, and which should be excluded until their replay contract is stable? A resident browser or background refresh is not an option.
4. **Research priorities:** Rank these after the news core: (a) corporate actions/results, (b) technical/fundamental screeners, (c) forecasts/analyst ratings, (d) F&O/open interest, (e) commodities/currencies, (f) mutual funds.
5. **Reuse decision:** Should implementation start from the forked PR #1701 artifact and its priceapi patch, or should it be treated as historical evidence and reprinted from a fresh current spec?

## Evidence log

### Pages visited in the dedicated EgoBrowser TaskSpace

- `https://www.moneycontrol.com/` (initial ad interstitial, then homepage)
- `/news/latest-news/`
- `/news/business/markets/`
- `/news/business/stocks/`
- `/news/tags/reliance-industries.html`
- representative article page ending in `-14034871.html`
- `/markets/` (resolved to the stock-market terminal path)
- `/markets/indian-indices/`
- `/india/stockpricequote/refineries/relianceindustries/RI`
- `/stock-scanner/technical/`
- `/stock-scanner/fundamental/`
- `/markets/corporate-action/`
- `/markets/corporate-filings/`
- `/news/business/earnings/`
- `/news/business/mutual-funds/`
- `/news/business/commodities/`
- `/news/gold-rates-today/`
- `/news/silver-rates-today/`
- `/news/business/ipo/`
- a public F&O futures-gainers page under `/stocks/fno/marketstats/`

### Current response observations

- `priceapi` NSE stock quote: HTTP 200, `application/json`, ~3.4 KB, `{code,message,data}`.
- `priceapi` encoded SENSEX index: HTTP 200, `application/json`, ~1.8 KB, `{code,message,data}`.
- `priceapi` encoded NIFTY index: HTTP 200, `application/json`, ~1.8 KB, `{code,message,data}`.
- `api` price-volume: HTTP 200, JSON, ~1.1 KB, `{success,data}` with fixed period keys.
- Reliance tag page: HTTP 200 HTML, ~636 KB.
- Representative article: HTTP 200 HTML, ~682 KB; `#contentdata.content_wrapper` existed and contained article text.
- `stockvitals`: HTTP 200 HTML, ~423 bytes.
- `stockscans`: HTTP 200 HTML, ~2.4 KB.
- `stockaction/topGainers`: HTTP 200 HTML content type, **0 bytes**.
- autosuggest POST/GET probes: current 404; page JS still constructs the legacy POST route.
- `changeTableData`: HTTP 200 HTML table fragment, ~1.3 KB.
- `newsapi/mc_news.php` safe probe: HTTP 200 JSON, tiny length object and no rows for the chosen query.
- Direct guessed technical/financial/forecast API calls returned JSON-shaped `{success:0,data:...}`; exact page-generated parameter/header contract is unresolved.

### External comparison sources

- Kite Connect portfolio/market-quotes/historical docs: <https://kite.trade/docs/connect/v3/portfolio/>, <https://kite.trade/docs/connect/v3/market-quotes/>, <https://kite.trade/docs/connect/v3/historical/>
- Tijori screener: <https://www.tijorifinance.com/filter/>
- Tickertape mutual-fund screener and portfolio: <https://www.tickertape.in/screener/home/mutual-fund>, <https://www.tickertape.in/portfolio/equity>
- Moneycontrol research/help pages: <https://www.moneycontrol.com/equity-research/>, <https://www.moneycontrol.com/stock-scanner/technical/>, <https://www.moneycontrol.com/stock-scanner/fundamental/>, <https://www.moneycontrol.com/markets/corporate-action/>, <https://www.moneycontrol.com/help/>

Agent Reach was required by the environment instructions and was attempted, but neither `agent-reach` nor the documented `conda` environment was available on PATH; no installation was performed. The live website evidence here came from EgoBrowser, and GitHub/competitor checks were read-only.

## Bottom line

Proceed only after the checkpoint. The safe, high-value Moneycontrol role is **live news and event context around an external portfolio**, with public price/index joins. The eventual CLI must stay stateless and online for data: no response cache, article DB, sync/history, offline index, stored data cursor, background daemon or resident browser. Approved authentication may use an established secure store, but never ad hoc cookie/token capture or emitted credentials. Reuse PR #1701 only for stateless request/parsing ideas, revalidate its priceapi workaround, and reject its local-state/`since` concepts. Do not assume the old screen/widget or authenticated portfolio surfaces are shippable.
