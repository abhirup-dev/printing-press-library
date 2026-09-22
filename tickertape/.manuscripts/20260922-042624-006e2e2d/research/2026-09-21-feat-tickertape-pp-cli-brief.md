# Tickertape discovery report

**Discovery only.** No CLI was generated, built, published, or added to `~/printing-press/library`. No broker was connected and no transaction, login, OTP, CAPTCHA, account, watchlist, basket, or portfolio mutation was initiated.

## Executive conclusion

Tickertape should play the **decision-support and portfolio-intelligence** role in a combined personal investing CLI—not the execution or source-of-truth ledger role.

Its strongest contribution is a compact, opinionated layer over Indian stocks, mutual funds, ETFs, indices, and US securities:

- company and fund scorecards (`Performance`, `Valuation`, `Growth`, `Profitability`, `Entry point`, `Red flags`);
- analyst-consensus and price/revenue/EPS forecast surfaces, where entitlement permits;
- peer comparison, key ratios, financial statements, shareholding/deal signals, and investment checklists;
- report/presentation attachments and company-event records;
- mutual-fund holdings, sector/asset allocation, manager/checklist/peer views;
- Market Mood Index (MMI), market movers, sector pages, collections, smallcase references, and social/content surfaces;
- portfolio diversification, red-flag, score-tier, and forecast views, but only behind authenticated/personal-session boundaries.

Relative to Tijori, Tickertape is more **scored, forecast-oriented, and consumer-decision friendly**. Relative to Kite/Zerodha, it is more **research-oriented** but is not the authoritative execution/holdings system. Relative to Moneycontrol, it is more **structured and analytical**, while Moneycontrol is the broader news/market-information surface.

The recommended initial product is a **public research CLI with explicit entitlement markers**, followed later by a carefully isolated authenticated read-only module. Do not make the first version a broker/portfolio automation tool.

## Evidence classification and method

- **Observed now:** live browser and read-only API observations from one dedicated EgoBrowser TaskSpace named `Tickertape discovery`, `spaceId=4`, using Page `p1`. The TaskSpace was closed after exploration.
- **Fetched now:** current public website pages and official/current Tickertape content fetched through web research.
- **Recovered from CASS:** prior session evidence that there was no built Tickertape CLI and no credible endpoint catalogue at that time; one prior note explicitly contrasted Tickertape with a Tijori-only local setup.
- **Third-party/inferred:** `The-Great-One/tickertape-api-client` GitHub repository and its reverse-engineered endpoint map. It is useful evidence, not an official contract. Its endpoint map says it was generated from production bundles and may drift.
- **Inaccessible/auth-gated:** endpoints that returned `401` without authenticated replay, and private/premium surfaces observed only as sanitized browser request metadata.

Required `agent-reach` routing was attempted (`agent-reach doctor`, Python module check, and `agent-reach check-update`), but this environment has neither the executable nor the module. No installation was performed because the discovery rules prohibit installing tools. Public web research therefore used the available search/fetch tools plus EgoBrowser. No secret, cookie, token, authorization header, or private payload was written to this report.

## Role in a combined investing/portfolio CLI

### Recommended role

```text
Kite / Zerodha  ── execution + broker-authoritative holdings
Tijori          ── sourced fundamentals + operational/company timeline research
Moneycontrol    ── broad news, market coverage, IPO/corporate-action context
Tickertape      ── scores, forecasts, red flags, MMI, structured cross-asset triage
```

Tickertape is most useful when a user already has an instrument or portfolio candidate and wants a quick answer to:

- “How does this asset score relative to peers?”
- “What are the obvious red flags or entry-point signals?”
- “What does analyst consensus/forecast data say, and what is premium-gated?”
- “What are the company’s public results/events/presentations and latest financial shape?”
- “How does this mutual fund/ETF/US security compare with alternatives?”
- “What is the current market mood or deal/insider activity?”

It should not claim to replace broker data for executed positions, order history, fills, taxes, or trades.

## User-visible feature inventory

### Navigation and asset coverage — observed now

The home navigation exposed these meaningful roots/surfaces:

- Indian `stocks`, `indices`, `etfs`, `mutualfunds`, `market-sectors`, `market-movers`, `market-mood-index`;
- `screener` with equity and mutual-fund screeners and prebuilt screens;
- `us-stocks` and US stocks/ETF pages;
- `news-events`, `social`, `knowledge-base`, `blog`, `glossary`;
- portfolio, gold, smallcase, loan-against-stocks, and loan-against-mutual-funds surfaces;
- pricing, disclosures, community guidelines, and support.

### Stock page — observed now

Live `/stocks/reliance-industries-RELI` exposed the following public-facing sections/headings:

- company identity, exchange, quote and key metrics;
- `Scorecard`;
- `Analyst Ratings & Forecast`;
- `Price Upside`, `Earnings Growth`, `Rev. Growth`;
- `Company Profile`, `Peers`, `Business Sentiment Analysis`, `Forecast`;
- a scorecard composed of `Performance`, `Valuation`, `Growth`, `Profitability`, `Entry point`, and `Red flags`.

The page also rendered authenticated `Current holdings` and Buy/Sell affordances in the ambient browser session. Those personal values and identifiers were deliberately not recorded or inspected further.

The live scorecard endpoint returned six records. Sanitized fields showed:

- score dimensions carry `name`, `tag`, `type`, `description`, `colour`, `score`, `rank`, `peers`, `locked`, `callout`, `comment`, `stack`, and `elements`;
- score objects carry `percentage`, `max`, `value`, and `key`;
- `Entry point` has element records with `title`, `type`, `description`, `flag`, `display`, `score`, and `source`;
- `Red flags` has the same element shape and multiple flag records;
- on the observed stock, the first four score dimensions were marked `locked=true`, while Entry point and Red flags were not. This is the clearest live evidence that the endpoint can be public while detailed score content remains premium-locked.

### Company reports, events, and publication timeline — observed now

A live public `GET /stocks/summary/:sid` response had this top-level data shape:

- `financialSummary` → `fiscalYearToData` with rows keyed by `year`, `revenue`, and `profit`;
- `shareHoldings` → records with `title`, `message`, `description`, `mood`;
- `news` → array;
- `events` → records with `title`, `desc`, `date`, `type`, `attachement` (spelling observed), `subType`, `value`;
- `aboutAndPeers` → records with `name`, `ticker`, `sid`, `sector`, `slug`, `ratios`, `description`;
- `forecast` → `totalReco`, `percBuyReco`;
- `brands` → `brandId`, `name`, `description`;
- `keyRatios` → `backL`, `value`;
- `investorPresentations` → `annId`, `broadcastTime`, `description`, `sid`, `subject`, `attachment`.

This is important for the combined CLI: Tickertape does have a practical company-report/timeline surface, but it is not a standalone filings database. The summary endpoint is the likely normalized source for company events, presentations, and result/publication records. The separate homepage event route observed on `/news-events` was:

```text
GET https://analyze.api.tickertape.in/v2/homepage/events?count&offset&sids&type
```

The real page supplied values and received `200`; a blank-parameter replay returned `400`. This strongly suggests pagination (`count`, `offset`) and optional scoping (`sids`, `type`), but the valid enum values for `type` were not guessed.

### Mutual funds — observed/fetched now

The public mutual-fund surface includes:

- collections, Top AUM, Top Performers, Bottom Performers;
- NAV, historical returns/chart/SIP chart;
- scorecard with Performance, Risk, Cost, Composition, and Red flags;
- expense ratio, benchmark, exit load, lock-in, SIP/lumpsum information;
- peer comparisons;
- actual portfolio allocation, sector distribution, holdings, fund managers, AMC profile, and checklist.

For a public sample fund, all of these direct GETs returned `200` JSON without credentials:

- `/mutualfunds/:mfId/info` → identity, AMC, NAV, option, sector/subsector;
- `/mutualfunds/:mfId/summary` → `meta`, `peers`, `schemeInfo`, `keyRatios`, `cagrSeries`, `taxMeta`, `amcDetails`, `labels`;
- `/mutualfunds/:mfId/holdings` → `assetAllocationHistory`, `currentAllocation`, `sectorDistribution`, `sectorWeightage`, `targetAssetAllocation`, `totalRedFlagsCount`;
- `/mutualfunds/:mfId/charts/inter?duration=...`;
- `/mutualfunds/:mfId/charts/sip`;
- `/mutualfunds/:mfId/fundmanagers`;
- `/mutualfunds/:mfId/investmentChecklists`;
- `/mutualfunds/:mfId/widget` → `info`, `points`, `ratios`.

### Equity and mutual-fund screeners — observed/fetched now

The equity screener page says it supports **200+ filters in 13 categories**. The current third-party filter inventory contains 247 filter rows, 69 marked premium/locked. Categories include profitability, growth, valuation, ownership, F&O, price/volume, financial ratios, Tickertape Special, analyst ratings, technical indicators, balance sheet/cash flow, ETFs, and income statement.

Premium-heavy groups include:

- forward growth and forward valuation;
- F&O/OI and rollover signals;
- Tickertape Special ranks;
- all analyst-rating fields;
- many technical indicators.

The MF screener page says it has **60+ filters** and prebuilt screens. Live public GETs returned 14 equity/MF prebuilt records respectively for the observed routes.

Observed/fetched screen concepts include Wealth Compounders, Analyst-Backed Bets, Penny Picks, Near 52W Lows, Momentum Monsters, Nearing Breakout, Hidden Gems, Dividend Gems, Cash Rich Smallcaps, and MF screens such as Top Tax Savers, Long Term Compounders, Bolder Bets, Value Picks, debt-focused, hybrid, and multi-asset screens.

The endpoint inventory documents `POST /screener/query` and `POST /mf-screener/query`, but no POST was invoked during discovery. These are semantically read operations but remain a separate implementation/review decision because they are POSTs and premium fields can return access errors.

### MMI and market surfaces — observed now

The live MMI page exposed:

- current Market Mood Index;
- change versus NIFTY;
- fear/greed state;
- FII activity;
- volatility/skew;
- momentum;
- additional indicator components.

The direct public endpoint is:

```text
GET https://api.tickertape.in/mmi/now
```

It returned JSON with `success`, `data`, and `error`; `data` contained `date`, `fii`, `skew`, `momentum`, `goldOnNifty`, `gold`, `nifty`, `extrema`, `fma`, `sma`, `trin`, and `indicator`.

The market-movers page is visibly about insider trades and bulk/block deals, with “Top Ideas” and investor/deal pages. Its live public routes included:

```text
GET https://analyze.api.tickertape.in/stocks/deals/ideas
GET https://analyze.api.tickertape.in/stocks/deals/insight?type&duration&sortBy
```

`/stocks/deals/ideas` returned a public array; the insight route returned `400` when the required query values were blank. No mutation was attempted.

### ETF, index, and US surfaces — observed now

Observed public pages:

- index page: ETFs tracking the index, key metrics, about/index description, all constituents;
- ETF page: scorecard, key metrics, AMC profile;
- US page: US stocks/ETFs from India, themes, global markets, and marketing for account opening/fractional shares/brokerage;
- digital gold and loan pages are separate financial-product funnels and should not be part of the first research CLI.

Public US API shapes returned `200`:

- US latest quotes: `data` object keyed by ticker;
- US security info and ETF info: `data.assets`;
- US stock overview: identity, exchange, asset type, sector, industry, description, flags, `updatedAt`, metrics, labels, holdings, peers;
- US ETF overview: same plus `etfType` and `topHoldings`;
- US historical chart: `marketStatus`, `points`, `h`, `l`, `r`;
- US financials: `financials`;
- US filters: `filters`.

### Pricing and premium gates — observed now

The public pricing page showed current displayed plans of ₹399 monthly, ₹899 quarterly, and ₹2,999 annual. Treat these as time-sensitive observations, not hard-coded product constants.

The page advertises the following Pro-gated or enhanced areas:

- smart/multi-asset portfolio analysis, multiple demat accounts, portfolio-vs-market comparisons;
- diversification score, red-flagged assets, forecast and performance tiers;
- advanced/custom screens and custom filters/universes;
- analyst ratings, price forecast, revenue/EPS forecast;
- score ranges and performance/valuation/profitability scorecards;
- business sentiment and stock-deal insights;
- data export, financial reports/presentation files;
- MF screener results, sector distribution, shareholding history;
- historical MMI data and index constituents;
- ad-free experience and Pro alerts.

## Network/API inventory

The following table focuses on representative routes that were either observed live or strongly corroborated by the current reverse-engineered wrapper. Query values and dynamic IDs are intentionally redacted.

| Host | Method + route pattern | Response | Access | IDs/pagination | Confidence |
|---|---|---|---|---|---|
| `www.tickertape.in` | `GET /`, `/stocks/...`, `/mutualfunds/...`, `/screener/...`, `/market-mood-index`, `/news-events` | HTML/SSR document | Public shell; may hydrate ambient account state | URL slugs contain `sid`/`mfId`-like identifiers | High, live browser |
| `api.tickertape.in` | `GET /mmi/now` | JSON `success,data,error`; MMI component fields | Public | none observed | High, live 200 |
| `quotes-api.tickertape.in` | `GET /quotes?sids=<csv>` | JSON `data` array | Public | comma-separated `sids` | High, live 200 |
| `gms-api.tickertape.in` | `GET /market/{IN\|US}/status` | JSON with market/window/holiday fields | Public | market enum | High, live 200 |
| `gms-api.tickertape.in` | `GET /quotes/US/latest?tickers=<csv>` | JSON `data` keyed by ticker | Public | comma-separated tickers | High, live 200 |
| `gms-api.tickertape.in` | `GET /US/securities/info?ticker=<csv>`; `/US/etfs/info?ticker=<csv>` | JSON `data.assets` | Public | ticker list | High, live 200 |
| `gms-api.tickertape.in` | `GET /US/{securities\|etfs}/{ticker}/overview` | JSON overview object | Public | ticker | High, live 200 |
| `gms-api.tickertape.in` | `GET /US/securities/{ticker}/charts/{intra\|inter}?duration=...` | JSON chart object | Public | ticker + duration | High, live 200 |
| `gms-api.tickertape.in` | `GET /US/securities/{ticker}/financials/{income\|balancesheet\|cashflow}?view=...` | JSON `financials` | Public | ticker + statement + view | High, live 200 |
| `api.tickertape.in` | `GET /search/suggest` (live `q` worked; wrapper documents `text`) | JSON `success,data`; data keys `examples,history,label` | Public; history may be session-sensitive | query key drift; no pagination observed | High for shape, medium for parameter contract |
| `api.tickertape.in` | `GET /stocks/info/{sid}` | JSON info/ratios/identity/labels | Public | `sid` | High, live 200 |
| `api.tickertape.in` | `GET /stocks/summary/{sid}` | JSON with financials, shareholdings, news, events, peers, forecast, presentations | Public summary; premium subsections may be locked | `sid` | High, live 200 |
| `api.tickertape.in` | `GET /stocks/charts/inter/{sid}?duration=...`; `/charts/intra/{sid}` | JSON chart data | Public | `sid`, duration | High, live 200 |
| `api.tickertape.in` | `GET /stocks/financials/{statement}/{sid}/{period}/{view}` | JSON array; observed annual income returned six rows | Public | statement/period/view + `sid` | High, live 200 |
| `api.tickertape.in` | `GET /stocks/investmentChecklists/{sid}` | JSON array; observed five records | Public | `sid` | High, live 200 |
| `api.tickertape.in` | `GET /stocks/news/{sid}` | JSON array | Public | `sid` | High, live 200; sample had empty array |
| `api.tickertape.in` | `GET /stocks/keyratioList/{sid}` | JSON array; observed six records | Public | `sid` | High, live 200 |
| `api.tickertape.in` | `GET /stocks/smallcases/{sid}?count=<n>&offset=<n>` | JSON array; observed three records | Public | explicit count/offset | High, live 200 |
| `api.tickertape.in` | `GET /stocks/commentaries/{sid}?keys[]=...` | JSON object; direct malformed/sparse query returned only `sid` | Public route, exact key enum unresolved | `sid`, repeated keys | Medium |
| `analyze.api.tickertape.in` | `GET /stocks/scorecard/{sid}` | JSON array of six score records | Public envelope; four score dimensions were `locked=true` | `sid` | High, live 200 |
| `analyze.api.tickertape.in` | `GET /stocks/aiSummary/{sid}` | JSON error shape on direct replay | Auth/entitlement; direct fetch `401` | `sid` | High for gate, low for schema |
| `api.tickertape.in` | `GET /stocks/estimates/forecast/{sid}`; `/stocks/ratings/{sid}` | JSON error envelope on direct replay | Auth/entitlement; direct fetch `401`; ambient page requested them successfully | `sid` | High for gate, medium for route |
| `api.tickertape.in` | `GET /stocks/aggregateddeals/{sid}` | JSON error envelope on direct replay | Auth/entitlement; direct fetch `401` | `sid` | High for gate |
| `analyze.api.tickertape.in` | `GET /stocks/deals/ideas` | JSON array; observed three records | Public | none observed | High, live 200 |
| `analyze.api.tickertape.in` | `GET /stocks/deals/insight?type&duration&sortBy` | JSON error on blank query; 400 | Public route with required query values not inferred | type/duration/sortBy | High for route/gate |
| `analyze.api.tickertape.in` | `GET /v2/homepage/events?count&offset&sids&type` | JSON route used by news/events page | Public page route; exact enums unresolved | count/offset/sids/type | High for live request; replay parameters unresolved |
| `api.tickertape.in` | `GET /mutualfunds/list` | JSON `data.universe` | Public | MF IDs, ISINs, slugs | High, live 200 |
| `api.tickertape.in` | `GET /mutualfunds/{mfId}/info` | JSON identity/AMC/NAV/labels | Public | `mfId` | High, live 200 |
| `api.tickertape.in` | `GET /mutualfunds/{mfId}/summary` | JSON peers, scheme info, ratios, CAGR, tax, AMC | Public | `mfId` | High, live 200 |
| `api.tickertape.in` | `GET /mutualfunds/{mfId}/holdings` | JSON allocation/sector/red-flag fields | Public | `mfId` | High, live 200 |
| `api.tickertape.in` | `GET /mutualfunds/{mfId}/charts/inter?duration=...`, `/charts/sip` | JSON chart arrays | Public | duration | High, live 200 |
| `api.tickertape.in` | `GET /mutualfunds/{mfId}/fundmanagers`, `/investmentChecklists`, `/widget` | JSON array/object | Public | `mfId` | High, live 200 |
| `api.tickertape.in` | `GET /screener/filters`, `/prebuilt`, `/universes` | JSON object/arrays; observed prebuilt array length 33 | Public | universe/filter metadata | High, live 200 |
| `api.tickertape.in` | `GET /mf-screener/filters`, `/prebuilt`, `/universes` | JSON object/arrays; observed prebuilt array length 14 | Public | universe/filter metadata | High, live 200 |
| `api.tickertape.in` | `POST /screener/query`, `POST /mf-screener/query` | JSON result payload (not invoked) | Public for some fields; premium fields may 403 | request body includes filters/project/sort/paging | Medium, wrapper/docs only |
| `api.tickertape.in` | `GET /screener/customFilters`, `/customUniverses`, `/screens`, `/exportLimit` | JSON | Auth/entitlement; ambient browser requested some | IDs for custom filters/universes/screens | Medium/High for live ambient route |
| `ecosystem.api.tickertape.in` | `GET /mf-screener/...`, `/screener/US/security/...` | JSON | Ambient auth/session for custom surfaces | custom IDs | Medium, live metadata |
| `ecosystem.api.tickertape.in` | `GET /portfolio/v4/holdings/status`, `/portfolio/v2/holdings/{id}`, `/portfolio/v2/metrics/indices`, `/portfolio/v2/insights/forecast/{id}` | JSON | Authenticated ambient session | portfolio ID, private IDs | High for sanitized route metadata; no values retained |
| `ecosystem.api.tickertape.in` | `GET /watchlists?assetClass&market` | JSON | Authenticated ambient session | assetClass/market, watchlist IDs | High for route metadata; no values retained |
| `api.tickertape.in` | `GET /user/basket`, `/user/holdings`, `/user/mfholdings`, `/user/subscription` | JSON | Authenticated; not queried for payload values | user/holding IDs | Medium/High from wrapper and ambient requests |
| `auth.api.tickertape.in` | login/OTP/refresh/user/profile route family | JSON/session | Authenticated flow; no calls made | session/user IDs | Medium, reverse-engineered docs only |
| `channels.api.tickertape.in` | `GET /v3/feeds?types&page&size` | JSON feed | Page requested it with ambient session; blank replay returned 400 | page/size/types | High for route/gate |
| `quotes-api.tickertape.in` | Socket.IO/WebSocket routes for stocks, screener, portfolio, MFs, indices, ETFs, MMI, US stocks | streaming messages | Public or session-dependent by stream | sids/rooms | Medium, wrapper docs; not opened during discovery |

### Host architecture from third-party evidence

The reverse-engineered wrapper groups production hosts as:

- `api.tickertape.in`: primary REST, stocks, screeners, MFs, user/portfolio;
- `quotes-api.tickertape.in`: quotes and real-time transport;
- `gms-api.tickertape.in`: US/global market service, forex, market status;
- `analyze.api.tickertape.in`: scorecard, analysis, deals, homepage events;
- `ecosystem.api.tickertape.in`: portfolio and newer MF/US/custom surfaces;
- `channels.api.tickertape.in`: feeds/notifications;
- `platform-ecosystem.api.tickertape.in`: product/pricing/launch and membership surfaces;
- `auth.api.tickertape.in`: session/auth/broker connection;
- `assets.tickertape.in`: static assets and feature flags.

The repository claims an endpoint map of 227 routes; its checked inventory sections contain 219 listed entries before WebSocket/host aggregate sections. This discrepancy and the undocumented nature of the system are additional reasons to generate an adapter from observed/validated routes rather than treating the wrapper as a stable spec.

## Public vs authenticated capability matrix

| Capability | Public now | Premium/entitlement | Authenticated session | Recommended discovery status |
|---|---:|---:|---:|---|
| Search/suggestions | Yes | No | No | MVP |
| Quotes, market status, MMI now | Yes | No | No | MVP |
| Stock identity/summary/ratios/charts | Yes | Some subsections | No for public subset | MVP |
| Financial statements and investment checklists | Yes | May vary by fields | No for observed public subset | MVP |
| Company events and presentation metadata | Yes via summary/events route | Attachments/details may vary | No for observed public subset | MVP with provenance flags |
| Scorecard envelope | Yes | Performance/valuation/growth/profitability locked in observed response | Not necessarily; entitlement matters | MVP as `locked`-aware output |
| Entry point and red flags | Public endpoint fields observed | May have depth limits | No for observed subset | MVP |
| Analyst ratings/price forecasts/AI summary | Not replayable without auth in this session | Likely Pro | Browser session/entitlement | Later, explicit gate |
| Equity/MF filter catalogs and prebuilt screens | Yes | Premium filter rows exist | No for catalogs | MVP |
| Execute screener query | Likely some public fields | Premium fields may 403 | May be required | Later after request contract review |
| MF info/summary/holdings/manager/checklist | Yes for public fund pages | Some deeper history may be Pro | No for observed public subset | MVP |
| ETF/index/US research | Yes for observed public data | Some scorecard/forecast details may gate | No for observed public subset | MVP/later based on breadth |
| Market movers/deal ideas | Public ideas route | Insight/analysis may gate | Some deal detail auth-gated | MVP public subset |
| News/events feed | Public page route | Scope/detail may vary | Social feed session-sensitive | MVP company timeline; social later |
| Watchlists | No anonymous replay validated | N/A | Yes | Later read-only, explicit consent |
| Portfolio holdings/metrics/diversification/forecast | No anonymous replay | Pro features | Yes | Later read-only; never first-run |
| Saved screens/custom universes/export | No anonymous replay validated | Often Pro | Yes | Later |
| Social feed | Page route exists; blank direct replay 400 | Unknown | Ambient session observed | Later |
| Broker connect, orders, trades, cancel | No | N/A | Yes | Explicitly out of scope |
| Digital gold, loans, payments, account opening | Marketing pages public | Product-specific | Yes/write flow | Explicitly out of scope |

## Candidate command tree

### MVP: public, read-only research

```text
tt search <text>

# Company / stock
 tt company get <sid-or-symbol>
 tt company financials <sid> [--statement income|balance-sheet|cash-flow] [--period annual|quarterly]
 tt company reports <sid>                 # presentations + attachment metadata
 tt company timeline <sid> [--limit N]    # events/news/publication records
 tt company checklist <sid>
 tt company ratios <sid>
 tt company peers <sid>
 tt company scorecard <sid>               # always show locked/premium flags
 tt company red-flags <sid>
 tt company entry-point <sid>
 tt company deals <sid>                   # only public deal/ideas subset

# Funds / ETFs / indices / US
 tt mf get <mf-id-or-query>
 tt mf holdings <mf-id>
 tt mf managers <mf-id>
 tt mf checklist <mf-id>
 tt mf peers <mf-id>
 tt mf chart <mf-id> [--duration ...]
 tt etf get <sid>
 tt index get <sid>
 tt us get <ticker>
 tt us financials <ticker> [--statement ...]

# Market and screen metadata
 tt market mood
 tt market movers [--type ...]
 tt market sectors
 tt screen filters [--asset-class equity|mf]
 tt screen presets [--asset-class equity|mf]
 tt market status [IN|US]
```

MVP implementation should cache public GETs with source timestamps and preserve an `access` object rather than silently dropping locked fields. It should not hard-code current prices or plan prices.

### Later: after a scope checkpoint

```text
tt screen run <query>                   # POST, public fields first; premium errors explicit
tt screen saved/list/get <id>            # authenticated
tt screen export <query>                 # authenticated/Pro, export-risk review
tt news timeline <scope>                 # all / company / portfolio / followed

tt portfolio summary                    # authenticated, read-only
tt portfolio analyze                    # diversification, red flags, tiers, forecast
tt portfolio holdings                   # authenticated, sensitive output
 tt watchlist list/get                   # authenticated, sensitive output
 tt analyst ratings <sid>                # Pro/auth replay
 tt analyst forecasts <sid>              # Pro/auth replay
 tt social feed                          # session-scoped, privacy review
 tt realtime quotes                      # Socket.IO transport review
```

### Explicitly out of scope

- OTP, email/password, social-login, CAPTCHA/2FA automation, cookie/token extraction, or session refresh implementation in the discovery CLI;
- broker connect/disconnect, order placement, order preview, trade cancellation, remittances, holdings import, and any buy/sell action;
- watchlist/basket/screen mutations, social posting/commenting/poll voting, or notification changes;
- digital-gold purchase/redemption, loans against stocks/MFs, subscription payment, offers, or account opening;
- using Tickertape as the authoritative tax, transaction, fill, or executed-holdings ledger;
- copying private portfolio values, watchlist contents, user IDs, cookies, or auth headers into local artifacts.

## Tijori ↔ Tickertape crosswalk

| Normalized need | Tijori strength | Tickertape strength | False-equivalence warning |
|---|---|---|---|
| Company identity/quote | Basic research identity | Strong cross-asset identity/quote pages | Quote freshness/source contract differs; Kite remains execution-grade |
| Financial statements | 10-year statements and deeper paid comparisons | Public statement routes and compact annual/quarterly summaries | Tickertape's summary is not a replacement for Tijori's sourced historical research |
| Source documents | Direct “Source” links to base data | Investor-presentation attachment metadata and event records | Tijori is stronger for provenance; Tickertape may expose an attachment but not a full source lineage |
| Results/publication timeline | Personalized Timeline with exchange filings, news, business changes, tweets | Company `events`, `news`, `investorPresentations`, plus paginated homepage events | Tickertape gives a useful normalized event shape; Tijori is the richer tracking feed |
| Company tracking | Watchlist, timeline, alerts | Watchlist/portfolio surfaces, but auth-gated | Do not merge as anonymous public capability |
| Operational metrics | 1,000+ current / 6,000+ historic operational metrics in paid Tijori plans | Some business sentiment, ratios, financials, brands, peers | Tickertape does not expose the same operational-metric/source-depth model |
| Market share / revenue mix | Core Tijori research differentiator | Not established in this discovery | Keep as Tijori-specific extension |
| Screeners | Natural-language and advanced financial screeners | 200+ stock filters, 60+ MF filters, prebuilt screens, premium analyst/technical fields | Filter names and query bodies are source-specific; normalize only common concepts |
| Scorecard | Not the same compact six-dimension scorecard | Performance, valuation, growth, profitability, entry point, red flags | Do not represent both as the same score; use `source_specific.insights` |
| Forecasts/analyst consensus | Reverse DCF and research comparisons | Analyst ratings, price/revenue/EPS forecast surfaces | Tickertape forecast routes were auth/entitlement-gated in direct replay |
| Portfolio analysis | Exposure/risk tracking | Diversification score, redflags, performance tiers, portfolio forecast | Both require different account/linkage models; never use either as broker truth |
| MFs/ETFs/US assets | Primarily company/fund research | Strong MF/ETF/index/US coverage in one navigation/data graph | Tickertape's multi-asset breadth is a meaningful unique contribution |
| Market mood/deal signals | Market Monitor/raw materials/macro | MMI, insider/bulk-deal ideas, market movers | These are complementary, not interchangeable |

### Suggested normalized commands and output schema

```text
research search <query> --source all
research company <symbol> --source tickertape|tijori|all
research report <symbol> --kind financials|presentation|result|event
research timeline <symbol> --scope company|all|portfolio
research screen <asset-class> --source tickertape|tijori
research scorecard <symbol> --source tickertape
research forecast <symbol> --source tickertape
research portfolio --source kite|tickertape|tijori
```

Common envelope:

```json
{
  "source": "tickertape",
  "observed_at": "...",
  "entity": {
    "kind": "stock|mf|etf|index|us_security",
    "sid": "...",
    "ticker": "...",
    "mf_id": "...",
    "exchange": "...",
    "slug": "...",
    "canonical_url": "..."
  },
  "quote": {"price": "...", "change": "...", "as_of": "..."},
  "fundamentals": {"financials": [], "ratios": [], "shareholding": []},
  "reports": [{"kind": "presentation|result|event", "date": "...", "title": "...", "attachment": "..."}],
  "timeline": [{"date": "...", "type": "...", "title": "...", "description": "..."}],
  "insights": {
    "scorecard": [],
    "entry_point": [],
    "red_flags": [],
    "analyst_consensus": null,
    "forecast": null,
    "sentiment": null
  },
  "access": {
    "public": true,
    "premium": false,
    "authenticated": false,
    "locked_fields": []
  },
  "provenance": {"host": "...", "route": "...", "confidence": "high|medium|low"}
}
```

Source-specific extensions should remain explicit:

- `tijori.operations`, `tijori.market_share`, `tijori.revenue_mix`, `tijori.source_links`, `tijori.timeline`;
- `tickertape.scorecard`, `tickertape.entry_point`, `tickertape.red_flags`, `tickertape.analyst_consensus`, `tickertape.forecast`, `tickertape.mmi`, `tickertape.deals`, `tickertape.mf_allocation`;
- `kite.orders`, `kite.executed_holdings`, `kite.tax`, `kite.fills`;
- `moneycontrol.news`, `moneycontrol.ipo`, `moneycontrol.corporate_actions`, `moneycontrol.alerts`.

## Overlap and unique contribution

| Source | Overlap with Tickertape | Unique contribution to the combined CLI |
|---|---|---|
| Kite/Zerodha | Quotes, watchlists, holdings, screener, portfolio analytics | Broker execution, orders/fills, authoritative broker holdings, tax/transaction data; do not duplicate with Tickertape |
| Tijori | Company research, financials, screeners, portfolio/watchlist/tracking | Operational metrics, market share, revenue mix, source links, sector/macro/raw-material research, richer timeline |
| Moneycontrol | Quotes, news, MFs, portfolio, alerts, IPO/corporate actions | Broad newsroom and market-event coverage; useful fallback/corroboration for event/news discovery |
| Tickertape | All of the above at varying depth | Opinionated scorecards, premium forecast/analyst fields, red flags, entry-point framing, MMI, cross-asset MF/ETF/US graph, deal ideas, smallcase references |

The CLI should make the overlap visible in output rather than pretending the sources agree. A good `research company --source all` result should show source-specific sections and provenance, not a single merged number without source labels.

## Auth, anti-bot, and risk assessment

### What was observed

- Direct `curl` reachability passed with `200` for the website HTML, MMI JSON, and quote JSON.
- EgoBrowser public GETs passed with `200` for the core public routes.
- No CAPTCHA, Cloudflare page, login wall, or anti-bot challenge was encountered during public exploration.
- The ambient browser profile was already authenticated. Public pages automatically requested authenticated-looking routes such as portfolio/holdings status, watchlists, custom universes, profile, membership, and product-launch routes. This was treated as ambient session state, not permission to exercise the account.
- Direct replay with `credentials: omit` returned `401` for analyst forecast, ratings, AI summary, and aggregated-deals routes. Some public scorecard data returned `200` but marked four score dimensions locked.
- Replaying authenticated/premium routes with `credentials: include` still returned `401` in the simple fetch path, indicating that a real CLI may need page-generated auth/CSRF/browser headers or a supported session handoff. No effort was made to extract or reconstruct them.

### Risk classes

| Class | Examples | Treatment |
|---|---|---|
| Low | Public GET quotes, MMI, stock/MF info, public financials, public checklists, filters, prebuilt screens | Good MVP candidates; cache and rate-limit |
| Medium | Authenticated portfolio/watchlists, saved screens, holdings, personalized feed | Sensitive read-only; explicit user consent and redaction required |
| High | OTP/login, broker connect, trades, cancel, basket/watchlist mutation, export/payment/account actions | Do not include in discovery CLI; separate approval and safety design required |
| Product/financial | Digital gold, loan-against-stocks/MFs, brokerage/account opening | Marketing pages may be read-only, but workflows are financially consequential; out of scope |

### Contract stability

There is no official OpenAPI contract or published rate-limit specification found. The strongest public wrapper explicitly calls these undocumented web-app endpoints and recommends low request rates, caching, retries/backoff, and fallbacks. Use a conservative client: one request at a time by default, bounded retries, cache TTL, response-shape validation, and clear “stale/locked/unavailable” states.

The public real-time quote layer is Socket.IO/WebSocket-based according to the reverse-engineered documentation. A CLI should start with REST snapshots; do not make a resident browser or WebSocket sidecar a prerequisite for MVP.

## Reuse opportunities

1. **Third-party endpoint catalogue:** `https://github.com/The-Great-One/tickertape-api-client` contains a typed public client, host grouping, `docs/endpoints.md`, and `docs/screener-filters.md`. Reuse as research evidence and cross-check, not as a stable API spec.
2. **Endpoint naming/data model:** its public methods map cleanly to the proposed `company`, `mf`, `etf`, `index`, `us`, `screen`, and `market` command groups.
3. **Schema quirks to preserve:** `events.attachement`, summary arrays, `fiscalYearToData`, `investorPresentations.attachment`, scorecard `locked`, and public/premium route splits.
4. **Tijori cross-source model:** align the future adapter behind normalized `entity`, `financials`, `reports`, `timeline`, `screen`, `portfolio`, and `insights` envelopes. Keep Tijori operational/source-link extensions and Tickertape scorecard/forecast/MMI extensions separate.
5. **Prior local evidence:** CASS search recovered no existing Tickertape CLI and no prior credible endpoint catalogue. `printing-press/library` also has no Tickertape directory. Do not mutate the library during discovery.
6. **No direct code reuse yet:** no package was installed, no third-party credentials were copied, and no auth-capture helper was run.

## Recommended checkpoint questions

1. **Public-first boundary:** Should MVP be strictly public read-only, with premium/authenticated commands hidden behind a separate opt-in module? Recommended: yes.
2. **Source precedence:** For common company fields, should Tijori remain the primary fundamentals/source-provenance source and Tickertape be the secondary scored/forecast source, with Kite authoritative for executed holdings? Recommended: yes.
3. **Opinionated signals:** Should scorecards, red flags, entry-point labels, and analyst consensus be emitted as clearly labeled source-specific opinions rather than merged into a universal score? Recommended: yes.
4. **MVP breadth:** Include MF/ETF/index/US research in the first release, or start company-first and add them after the common schema is validated? Tickertape's multi-asset breadth is valuable, but company reports/timelines are the highest cross-source leverage.
5. **Timeline semantics:** Should `company timeline` use Tickertape events/presentations/news only, or merge with Tijori filings/timeline and Moneycontrol news with source badges and de-duplication?
6. **Forecast gate:** Is it acceptable to expose forecast/analyst commands only when the user supplies a legitimate authenticated Pro session, with `401`/locked fields surfaced rather than silently omitted?
7. **Portfolio boundary:** Should Tickertape portfolio/watchlist analysis remain later and read-only, with no broker linking/import/connect flows? Recommended: yes.

## Uncertainties and follow-up work

- Current page data can be personalized by an already-authenticated ambient browser session. Public HTML and public GETs should be revalidated from a clean anonymous context before implementation.
- Forecast, ratings, AI summary, aggregated deals, and private portfolio routes need a consented auth design; their exact response schema was not captured because direct replay returned `401` and no credentials were extracted.
- `screener/query` and `mf-screener/query` request bodies, enum values, pagination envelope, and premium error behavior need a separate safe browser capture or fixture. No POST was sent during this discovery.
- `homepage/events` `type` values and `sids` scoping were left unresolved rather than guessed.
- `search/suggest` showed parameter drift: the live probe accepted `q`, while the third-party client documents `text`. The adapter should discover and validate the current parameter name at implementation time.
- Social feed route parameters and auth requirements remain unresolved; blank replay returned `400`.
- Moneycontrol's portfolio page was HTTP 403 to the fetcher, so its feature comparison is based on search evidence and public descriptions, not live browser validation.
- Pricing, filter counts, and content claims are time-sensitive. Store observation timestamps and do not bake them into command semantics.

## Evidence index

### Live browser pages visited in the dedicated TaskSpace

- `https://www.tickertape.in/`
- `https://www.tickertape.in/screener/home/equity`
- `https://www.tickertape.in/screener/home/mutual-fund`
- `https://www.tickertape.in/market-mood-index`
- `https://www.tickertape.in/pricing?section=features`
- `https://www.tickertape.in/portfolio/equity` — sanitized route metadata only; no personal values retained
- `https://www.tickertape.in/stocks/reliance-industries-RELI` — sanitized public scorecard/routes only; holdings values omitted
- `https://www.tickertape.in/market-sectors`
- `https://www.tickertape.in/market-movers`
- `https://www.tickertape.in/news-events`
- `https://www.tickertape.in/social`
- `https://www.tickertape.in/mutualfunds`
- `https://www.tickertape.in/indices/nifty-50-index-.NSEI`
- `https://www.tickertape.in/etfs/nippon-india-nifty-50-bees-etf-NBES`
- `https://www.tickertape.in/us-stocks`
- `https://www.tickertape.in/digital-gold`
- `https://www.tickertape.in/loan-against-stocks`
- `https://www.tickertape.in/loan-against-mutual-funds`

### Representative sanitized requests

```text
GET https://www.tickertape.in/                                  -> 200 text/html
GET https://api.tickertape.in/mmi/now                          -> 200 application/json
GET https://quotes-api.tickertape.in/quotes?sids=<csv>         -> 200 application/json
GET https://api.tickertape.in/stocks/summary/<sid>             -> 200 application/json
GET https://api.tickertape.in/stocks/financials/<stmt>/<sid>/<period>/<view> -> 200 application/json
GET https://analyze.api.tickertape.in/stocks/scorecard/<sid>   -> 200 application/json; locked fields present
GET https://api.tickertape.in/stocks/estimates/forecast/<sid>  -> 401 application/json without entitlement replay
GET https://api.tickertape.in/stocks/ratings/<sid>             -> 401 application/json without entitlement replay
GET https://api.tickertape.in/mutualfunds/<mfId>/summary       -> 200 application/json
GET https://gms-api.tickertape.in/US/securities/<ticker>/overview -> 200 application/json
GET https://analyze.api.tickertape.in/v2/homepage/events?count&offset&sids&type -> page request 200; blank replay 400
GET https://channels.api.tickertape.in/v3/feeds?types&page&size -> page request 200; blank replay 400
```

### Public research sources

- Tickertape pricing/features: `https://www.tickertape.in/pricing?section=features`
- Equity screener: `https://www.tickertape.in/screener/home/equity`
- Mutual-fund screener: `https://www.tickertape.in/screener/home/mutual-fund`
- MMI: `https://www.tickertape.in/market-mood-index`
- Stock page: `https://www.tickertape.in/stocks/reliance-industries-RELI`
- Mutual-fund page: `https://www.tickertape.in/mutualfunds/mahindra-manulife-focused-fund-M_MAHD`
- Portfolio-analysis product article: `https://www.tickertape.in/blog/diversification-score-redflags-and-portfolio-forecast-our-new-updates-make-portfolio-analysis-quick-and-easy/`
- Tijori features: `https://www.tijorifinance.com/features/`
- Zerodha screener documentation: `https://support.zerodha.com/category/trading-and-markets/general-kite/others-kite/articles/create-and-save-custom-screeners`
- Zerodha portfolio analytics: `https://support.zerodha.com/category/console/portfolio/console-holdings/articles/console-analytics`
- Moneycontrol portfolio/search evidence: `https://www.moneycontrol.com/portfolio-management/portfolio-investment-signup.php`, `https://www.moneycontrol.com/stocksmarketsindia/`, `https://www.moneycontrol.com/ipo/`, `https://www.moneycontrol.com/my-alerts`

### Third-party endpoint evidence

- `https://github.com/The-Great-One/tickertape-api-client`
- `docs/endpoints.md`
- `docs/screener-filters.md`
- `src/tickertape_api/client.py`
- `src/tickertape_api/portfolio_client.py`

### CASS/local evidence

- CASS query terms: `Tickertape CLI`, `Tickertape endpoints`, `Tickertape discovery`.
- Recovered result: prior session explicitly stated no Tickertape CLI existed locally and contrasted it with a Tijori-only setup.
- Local `printing-press/library` has no `tickertape` directory.

## Bottom line

Proceed only after the checkpoint confirms a public-first, research-only scope. The best first Tickertape adapter is not “all 227 endpoints”; it is a stable, provenance-aware slice around:

1. company summary + financials + reports/presentations + events/timeline;
2. scorecard/entry-point/red-flags with locked-state fidelity;
3. MF/ETF/index/US public research;
4. MMI and public market/deal ideas;
5. screener filter catalogs and prebuilt screens.

Treat forecast/ratings/portfolio/watchlists as separate entitlement-aware extensions. Keep execution, broker, financial-product, and account mutation surfaces out of scope.

## User-approved build scope

- Broad public read-only and stateless for all domain data: every command fetches live data; no domain cache, SQLite, local history, sync, cursors, or persistence.
- Optional entitlement-aware analyst ratings, forecasts, AI insights, and deal routes may use a legitimate external ambient session when available, but must return explicit access/locked metadata otherwise and must never require premium auth globally.
- Auth credentials/session metadata may be persisted only in an appropriate secure credential store or permissioned auth config; never in source, manuscripts, logs, proofs, output, or domain-data storage.
- Exclude portfolio/watchlist/saved-screen reads, broker linking/trading, mutations, social actions, exports, payments, loans, gold, and account workflows.

## Reachability Gate
- Decision: PASS
- Evidence: `GET https://api.tickertape.in/mmi/now` returned HTTP 200 with a JSON response.
