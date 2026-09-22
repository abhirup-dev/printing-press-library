# chatgpt-pp-cli Shipcheck Proof — Phase 4

## Final result: SHIP (verdict PASS, 7/7 legs)

```
LEG               RESULT  EXIT
verify            PASS    0
validate-narrative PASS    0
dogfood           PASS    0
workflow-verify   PASS    0
apify-audit       PASS    0
verify-skill      PASS    0
scorecard         PASS    0
Total: 94/100 - Grade A
```

Live sample probe (with real session token via env, never logged): 5/5 features passed
- list inventory (cold isolated HOME, recency early-stop): 4.2s
- transcript verified-complete extraction: PASS
- outline: PASS
- transcript section slicing: PASS
- find (server global search): PASS

## Blockers found and fixed (loops)
1. scorecard leg "manifest evidence unavailable" → CLI dir lacked .printing-press.json (generator omission on browser-sniffed path); created schema-v2 manifest with correctly-shaped novel_features/novel_features_built objects.
2. Live probe: truncated conversation IDs in examples → 400 Invalid conversation. Fixed to full UUIDs everywhere (research.json + command examples).
3. Live probe timeout on `list --since 30d --order created --hydrate`: three causes fixed —
   a. time.ParseDuration rejects "30d": custom parser with d/w units.
   b. Page-walk enumerated the entire 100+ item account before filtering: added recency early-stop (order=updated desc ⇒ stop when a page predates the cutoff). Cold isolated run 54.7s → 4.2s.
   c. Hydration unbounded per run: now time-budgeted (--hydrate-budget 2s default) + --max-hydrate cap + incremental cache; honest note reports hydrated N of M.
4. Mint-death loop: without a cookie jar every request re-attempted a doomed /api/auth/session mint (2× latency). Mint failure now cached per process; env/config token path is direct.
5. Message.weight is a JSON float (0.0) — retyped json.Number (live hydration parse failure).
6. which-index novel commands must be flag-free (generator re-synced them from research.json commands that included flags; normalized at source, dogfood re-sync now stable).
7. `-o` shorthand added to transcript --out (narrative recipe used it).
8. Dual-search command renamed search→find at absorb time (framework search = offline FTS); highlights/which/README all synced.

## Verify pass rate
Before first fix loop: 6/7 legs (validate-narrative FAIL on -o). After fixes: 7/7 PASS.
Scorecard: first measured 91 (2 dims unverified) → final 94/100 Grade A with all dimensions verified (spec semantic validation + live probes).

## Behavioral spot-checks (live)
- `transcript --verify` on the 103-message long thread: complete=true, tail_matches_current_node=true.
- `outline` on same: 5 sections, anchors = user message ids, excerpts verbatim.
- `find "haptics"`: federated items with snippets + message ids, source_statuses ok.

## Recommendation: ship
All ship-threshold conditions met; no known functional bugs in shipping scope. Send surface (chat ask/status) is explicitly experimental + browser-gated by design (no Turnstile solver; clear failure path) per supervisor constraint — documented as such in help/README.
