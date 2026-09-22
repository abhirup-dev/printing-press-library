# UnityPredict CLI — Absorb Manifest

## Tools that touch this API (exhaustive — private API, small ecosystem)

| Tool | What it is | Where |
|---|---|---|
| `unitypredict` SDK CLI | First-party, engine lifecycle only | `unitypredict-sdks/Python/EngineSDK`, also PyPI `unitypredict-engines` |
| UPT ApiSDK (Python) | First-party library for invoking models | `unitypredict-sdks/Python/ApiSDK` |
| `upt-workflows` skill scripts | `query_upt_model.py`, `upt_models.py`, `upt.sh` | `~/.config/skillshare/skills/upt-workflows/scripts` |
| Console web UI | The full surface; the only place most of this is reachable | `console.{dev,prod}` |
| Go / JS / .NET clients | First-party, thin API clients | `unitypredict-sdks/UnityPredict.{Go,JS,Net}` |

No community SDKs, no npm package, no third-party GitHub client. Verified by web search:
the only public artifact is the first-party `unitypredict-engines` PyPI package.

## Absorbed (match or beat everything that exists)

| # | Feature | Absorbed from | Notes |
|---|---|---|---|
| 1 | List/search models (`--search`, paging, type filter) | `upt_models.py list` | Console param is `?searchBy=` |
| 2 | Inspect a model record | `upt_models.py inspect` | Includes `aiEngineConfig` |
| 3 | List engines | console | `userengines` |
| 4 | Get engine record | console / API | Bare-uuid rule applies |
| 5 | Invoke a model (text input) | `query_upt_model.py`, ApiSDK | |
| 6 | Invoke a model (file input) | ApiSDK | `initialize` → presign → `PUT` → predict |
| 7 | Poll request to completion | `query_upt_model.py` | |
| 8 | Download File outcomes | ApiSDK | Presigned; **no** auth header |
| 9 | Fetch devLog for a request | `query_upt_model.py --devlog` | Presigned; **no** auth header |
| 10 | Per-model request history | console | `models/requests/{id}` |
| 11 | Last deploy logs / `DebugLogs.txt` | `unitypredict --getLastDeployLogs` | |
| 12 | Deploy status | `unitypredict --deployStatus` | |
| 13 | Profile/env selection (dev/prod) | all | `~/.unitypredict/credentials` |
| 14 | Repository search | console | |
| 15 | Account / identity record | console | `auth?loadRD=true` |
| 16 | Model metadata edit (name, descriptions, category, pricing, flags) | console **only** | Not in any CLI today |
| 17 | Model create | console **only** | Not in any CLI today |
| 18 | Long-description file upload | console **only** | presign → S3 `PUT` → upsert |
| 19 | Input/outcome configuration | console **only** | `aiEngineConfig`; the trickiest payload |
| 20 | Engine config (platform, base image, memory, batch, chaining) | `unitypredict` partial | |
| 21 | Thumbnail fetch | console | `fast.` host |

Rows 16-19 are the core of the thesis: the console is currently the *only* way to do them.

## Transcendence (only possible with our approach)

Scored /10 on user value. `Buildability`: `spec-emits` = generator emits it from the
spec; `hand-code` = hand-written Go after generate (~50-150 LoC each + `root.go` wiring).

| # | Feature | What you get | Score | Buildability |
|---|---|---|---|---|
| T1 | `logs get <requestId> --all` | One command fans out to devLog + `Log.txt` + `DebugLogs.txt` and prints them in order. Today this is 3 routes, 2 of which are presigned and 302 to S3. | 9 | hand-code |
| T2 | `model describe set <id> --file README.md` | Collapses presign → S3 `PUT` → upsert into one command. Today: console-only, 3 manual steps. | 9 | hand-code |
| T3 | `model io set <id> --from io.yaml` | Declarative inputs/outcomes. The single most error-prone payload on the platform. | 8 | hand-code |
| T4 | `engine models <id>` | Reverse lookup: which models use this engine. The `engineId=` query param on list routes is silently ignored, so this is otherwise unanswerable without N calls. | 8 | spec-emits |
| T5 | `auth check` | Validates using `GET /api/engines/{bare-uuid}`, not a list route. Directly prevents the 200-on-garbage-bearer trap that has already burned this account once. | 8 | hand-code (small) |
| T6 | `engine concurrency <id>` | Renders `stateInfo.historicalNodeCountStats` + `numRunningInstances` as a table. Answers "how much was running in parallel" without the Stats page. | 7 | hand-code |
| T7 | `predict --wait --download ./out` | Invoke, poll, and pull File outcomes in one call, dropping `Authorization` on the S3 redirect. | 7 | hand-code |
| T8 | Local store + `search` / `sql` | Offline query across all 409 dev models; the old pp-cli's best feature. | 7 | spec-emits |
| T9 | `diff <entity> --dev --prod` | Compare a model or engine across environments. Two tenants, two keys — nothing today can do this in one step. | 6 | hand-code |
| T10 | Prod read-only guard | Mutating verbs refuse against prod unless `--i-know`. | 6 | hand-code (small) |

## Hand-code commitment
- **10 novel features scored ≥5/10.**
- **8 require hand-written Go** after generate: T1, T2, T3, T5, T6, T7, T9, T10.
- **2 are emitted from the spec**: T4, T8.

## Known gaps carried into generation
- The sniffed spec has **25 endpoints**; `endpoints.md` documents **37 routes**. The
  12-route delta (including `GET /api/engines/{id}/models`) must be merged from
  `endpoints.md` at generate time, not dropped.
- `POST /api/auth` (account upsert / API-key minting) is deliberately **excluded** —
  it is a full-record upsert and a partial body clobbers the account record.
- No stubs are planned. Every row above ships working or is not listed.
