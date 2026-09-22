## Customer model

**Indian retail swing investor.**  
**Today (without this CLI):** Checks Moneycontrol market and stock pages, then opens several quote, news, and corporate-event tabs before deciding whether a holding needs attention.  
**Weekly ritual:** Reviews holdings and watchlist context after the weekly close.  
**Frustration:** Price movement, relevant headlines, and upcoming company events are scattered across separate live pages.

**Brokerage-linked long-term investor.**  
**Today (without this CLI):** Gets current holdings from a broker tool and manually searches Moneycontrol for company-tagged news and filings.  
**Weekly ritual:** Checks whether any holding has a material announcement or result-related development.  
**Frustration:** The broker portfolio and Moneycontrol’s narrative context cannot be joined without repetitive copy-paste.

**Indian market-opening researcher.**  
**Today (without this CLI):** Reads index snapshots, market news, stock stories, and event pages in sequence each morning.  
**Weekly ritual:** Builds a personal pre-market or Monday market brief from current Moneycontrol pages.  
**Frustration:** It is tedious to assemble a consistent, source-limited view without stale or locally stored data.

## Candidates (pre-cut)

| Feature | Command | Description | Persona served | Source | Long Description | Inline verdict |
|---|---|---|---|---|---|---|
| Holding context join | `moneycontrol context holdings --sc-ids RI,INFY` | Fetches current stock snapshots, company-tag news, and relevant corporate-event/filing pages for supplied symbols in one live request flow. | Brokerage-linked long-term investor | (a), (c) | none | Keep pending cut; stateless live join. |
| Event-risk board | `moneycontrol events upcoming --sc-ids RI,INFY` | Presents live corporate actions, filings, and earnings items grouped by supplied company. | Long-term investor | (a), (b), (c) | none | Keep pending cut; mechanical extraction only. |
| Market breadth brief | `moneycontrol market breadth` | Combines live Indian-index snapshots with the stable index change-table fragment and prints advancing/declining context. | Market-opening researcher | (a), (b) | none | Keep pending cut; no summary model. |
| Catalyst matrix | `moneycontrol catalysts --sc-ids RI,INFY --days 30` | Extracts dated IPO, earnings, filing, and corporate-action entries for supplied symbols from live Moneycontrol pages. | Swing investor | (a), (b) | none | Keep pending cut; date filtering is local formatting of live responses. |
| Company news timeline | `moneycontrol news timeline --sc-id RI --limit 20` | Joins a company tag feed with article metadata and prints a chronological, URL-preserving timeline. | Swing investor | (b), (c) | none | Keep pending cut; distinct from absorbed single-feed news. |
| Index-to-stock movers | `moneycontrol market movers --index sensex` | Fetches a live index change table and emits constituent rows with price/change fields where supplied by the table. | Market-opening researcher | (b), (c) | none | Keep pending cut; must use a real change-table response. |
| Portfolio news triage | `moneycontrol portfolio triage --sc-ids RI,INFY` | Emits only supplied-company headlines, timestamps, URLs, and event labels for quick manual review. | Brokerage-linked long-term investor | (a), (c) | none | Keep pending cut; mechanical, pipe-friendly output. |
| Research packet | `moneycontrol research packet --sc-id RI` | Produces one live, structured packet containing quote, price-volume context, tagged news, filings, results, and corporate actions. | Swing investor | (b), (c) | none | Keep pending cut; endpoint-backed composition. |
| Personalized alert creation | `moneycontrol alerts create ...` | Creates Moneycontrol alerts for supplied symbols. | Swing investor | (a) | none | Kill candidate; write operation explicitly out of scope. |
| News sentiment ranking | `moneycontrol news sentiment --sc-ids ...` | Ranks company stories by positive or negative sentiment. | Swing investor | (a) | none | Kill candidate; requires NLP. |
| Offline event history | `moneycontrol events since --local-cursor ...` | Finds new events since a stored local cursor. | Long-term investor | (c), (d) | none | Kill candidate; forbidden stored cursor/history and prior PR’s rejected local-state design. |
| Authenticated portfolio dashboard | `moneycontrol portfolio dashboard` | Reads and persists a full personal Moneycontrol portfolio dashboard. | Long-term investor | (a), (e) | none | Kill candidate; authenticated personal surface is deferred and persistence is forbidden. |

## Survivors and kills

### Survivors

| # | Feature | Command | Score | Buildability | How It Works | Evidence | Long Description |
|---|---|---|---:|---|---|---|---|
| 1 | Holding context join | `moneycontrol context holdings --sc-ids RI,INFY` | 9/10 | hand-code | Calls live pricefeed stock quotes plus company-tag news and public event pages, then joins results only for the supplied symbols without persistence. | Brief’s combined-CLI thesis explicitly recommends holdings-to-tagged-news joins; pricefeed and tag pages were observed live. | none |
| 2 | Event-risk board | `moneycontrol events upcoming --sc-ids RI,INFY` | 8/10 | hand-code | Calls live corporate-action, filings, and earnings pages and mechanically filters/group dates by supplied company. | Public corporate-action, filings, and earnings pages were observed; investor event context is the brief’s recommended later value. | none |
| 3 | Market breadth brief | `moneycontrol market breadth` | 7/10 | hand-code | Calls the live encoded index pricefeed and `changeTableData` endpoint, emitting index changes and available breadth rows. | Both encoded index pricefeed and change-table routes were observed; market action widgets were unstable, so this uses the stable table route. | none |
| 4 | Catalyst matrix | `moneycontrol catalysts --sc-ids RI,INFY --days 30` | 8/10 | hand-code | Fetches live earnings, IPO, filings, and corporate-action HTML, extracts dated entries, and filters them mechanically by symbol and date. | Brief inventories all four public event/news surfaces and identifies event context around external holdings as Moneycontrol’s differentiator. | none |
| 5 | Company news timeline | `moneycontrol news timeline --sc-id RI --limit 20` | 6/10 | hand-code | Calls the live company-tag page, extracts article titles, timestamps, canonical URLs, and visible metadata, and orders the current response. | Reliance company-tag HTML and exact article metadata were observed; the timeline is a service-specific tag workflow rather than a generic latest feed. | none |
| 6 | Portfolio news triage | `moneycontrol portfolio triage --sc-ids RI,INFY` | 8/10 | hand-code | Calls live tag feeds for supplied symbols and emits mechanical headline, timestamp, URL, and event-category rows without sentiment or storage. | Brief recommends portfolio-news as a combined-CLI output while explicitly requiring external holdings and stateless live requests. | none |
| 7 | Research packet | `moneycontrol research packet --sc-id RI` | 7/10 | hand-code | Calls live quote, price-volume, company-tag news, filings, results, and corporate-action endpoints and emits one source-linked structured packet. | Pricefeed and price-volume JSON were observed; company pages expose the additional event/news sections needed for the packet. | none |

### Killed candidates

| Feature | Kill reason | Closest-surviving-sibling |
|---|---|---|
| Index-to-stock movers | The observed index table does not guarantee constituent price data in one stable response; it is too close to a fragile endpoint wrapper. | `moneycontrol market breadth` |
| Personalized alert creation | Financial-account mutation is explicitly out of scope. | `moneycontrol portfolio triage` |
| News sentiment ranking | Requires NLP/sentiment analysis, failing the mechanical-feature rule. | `moneycontrol portfolio triage` |
| Offline event history | Requires forbidden local cursor, history, cache, or database state. | `moneycontrol catalysts` |
| Authenticated portfolio dashboard | Personal payloads are unvalidated, authentication is deferred, and persistence is prohibited. | `moneycontrol context holdings --sc-ids ...` |
