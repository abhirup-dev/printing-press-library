# UnityPredict (UPT) CLI Brief

## API Identity
- **Domain:** Private AI/ML model-hosting and inference platform. Users author *engines*
  (containerised Python inference code), deploy them, wrap them in *models* (the
  invocable, metadata-bearing public face), then invoke and observe them.
- **Users:** Engine developers and platform operators on the UnityPredict org. This CLI
  targets a single operator who owns engines across two environments.
- **Data profile:** 409 models on dev; models, engines, repositories, per-model request
  histories, and multi-granularity logs. Two isolated environments
  (`api.dev.unitypredict.net`, `api.prod.unitypredict.com`) with separate Auth0 tenants
  and separate durable API keys.
- **Not public:** no OpenAPI spec, no npm/PyPI wrapper, no GitHub issues to mine. The
  surface in this run came from a live authenticated browser sniff
  (`~/Codes/UPT/.upt-sniff/`), replayed against dev to capture real response shapes.

## Reachability Risk
- **None.** Verified live this session on both environments. `GET /api/engines/{bare-uuid}`
  returns `200` with a valid durable key on dev and prod; a garbage key returns `401` on
  both. No WAF, no bot challenge, no Cloudflare/Vercel mitigation — plain `curl` reaches
  every route directly.
- Probe-safe endpoint: `GET /api/engines/{engine_id}` (read-only, owner-scoped, and the
  only honest auth discriminator — see Auth Traps).

## Top Workflows
1. **Find things.** Search/list models and engines; reverse-lookup which models use a
   given engine (`GET /api/engines/{id}/models`).
2. **Create and shape a model.** Create it, upsert the full metadata payload, upload the
   long description as a Markdown/HTML file (presign → S3 `PUT` → upsert), and configure
   `aiEngineConfig` inputs/outcomes.
3. **Invoke and poll.** Predict with text or file inputs, poll to completion, download
   File outcomes.
4. **Observe.** Per-model request history, then pull logs at every granularity —
   devLog, `DebugLogs.txt`, `getLastDeployLogs`, and `GET /api/predict/status/{id}/logs`.
5. **Capacity.** Engine concurrency from `GET /api/engines/{id}` → `stateInfo`
   (`numRunningInstances`, `runningNodes[]`, `historicalNodeCountStats`).

## Table Stakes (vs the incumbent)
The incumbent is the first-party `unitypredict` SDK CLI (from `unitypredict-sdks`
`Python/EngineSDK`). Its entire flag surface, captured this session:

```
--configure --profile --list_profiles --engine --create --remove --run --push
--deploy --forceDeploy --delete --pull --engineId --parentRepoId --platform
--baseImage --batchMode --maxBatchSize --disableBatch --getLastDeployLogs
--deployStatus --uploadTimeout --deployTimeout --timeout --yes
```

It is **engine-lifecycle-only**. It has no command for: model search or inventory,
model creation or metadata, description upload, input/outcome configuration,
invoking a model, request history, or any log beyond `--getLastDeployLogs`.

## Why install this CLI instead of the incumbent
It covers the entire console/model surface the incumbent cannot reach, and it is the
only way to script the observability half (requests, multi-granularity logs,
concurrency) without clicking through the web console. The two are complements, not
rivals — this CLI deliberately does not re-implement engine build/push/deploy.

## Data Layer
- **Primary entities:** `model` (modelId, aiEngineConfig.inputs/outcomes, pricing,
  visibility), `engine` (engineId, platform, baseImage, stateInfo), `repository`
  (ParentRepositoryId), `request` (requestId, status, devLogUrl).
- **Sync cursor:** `modifiedDate` on model/engine records; `pageNumber` on the
  requests route.
- **FTS/search:** model name + description. Note the console's search param is
  `?searchBy=`, not `?search=`.

## Auth Traps (must survive into the generated client)
1. **`APIKEY@` prefix is mandatory.** Header is `Authorization: Bearer APIKEY@<key>`.
   Plain `Bearer <key>` → `401`.
2. **The prod key contains `{ [ & * )`.** The previous generated CLI ran its
   unresolved-placeholder guard *after* substituting into `Bearer {token}`, so the guard
   fired on the credential's own bytes and sent the request with **no** Authorization
   header — surfacing as a misleading `401`. The guard must run on the format string
   **before** substitution.
3. **List endpoints return `200` for a garbage bearer**
   (`/api/models/usermodels`, `/api/engines/userengines`, `/api/repository/search`).
   They must never be used to validate auth. `GET /api/engines/{bare-uuid}` is the only
   honest discriminator: valid `200`, bad key `401`, unknown id `404`.
4. **Strip the `AppEngineDefinition-` prefix.** The `id` field on `userengines` carries
   it; passing it through yields a misleading `404`. Use the bare `engineId`.
5. **Presigned URLs must be fetched WITHOUT the Authorization header.** Confirmed
   destructively this run: `GET /api/models/{id}/filekey/MODELDESCRIPTION` redirects to
   S3, and when the auth header rides along S3 returns `400` *and echoes the credential
   back in the error body*. The client must drop `Authorization` on cross-host redirect.
   (This corrects the recon's guess that the route was JWT-only — it is not.)
6. **`engineId=` as a query param on list routes is silently ignored.** Use the
   dedicated reverse-lookup route instead.

## Product Name and Thesis
`unitypredict-pp-cli` — *the console half of UnityPredict, scriptable.* Search, shape,
invoke, and observe models from the terminal; leave engine build/deploy to the
first-party SDK CLI.

## Safety Posture
Prod is read-only by default. Creates, edits, and uploads target dev unless explicitly
overridden.
