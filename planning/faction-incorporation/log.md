# Faction & Incorporation — implementation log

## 2026-07-30 — accepted and started

- Owner answered the acceptance bounce with FA/FB/FC: stocks are run-scoped int64 Company fields,
  accrual is one unit per attended minute with carried milliseconds and accrual-only saturation,
  and all four Phase-0 faction objects are literal.
- Acceptance re-read confirmed these contracts close the previously recorded resource/rate/data
  gaps. RFC moved to implementing; no inferred balance mechanics are required.
- Implementation order follows the dependency surface: catalog/identity → save shape → intents and
  Commons binding → accrual → boards/reset/docs.

## 2026-07-30 — catalog and epoch-guard foundation

- Added the strict Go catalog and matching JSON Schema/client semantic validator. Both enforce the
  literal four-faction set, one producer/consumer per stock, the single Hamiltonian cycle, empty
  Phase-0 modifier hooks, mechanical copy keys, and the Open Source tithe against the Commons band.
- The accepted RFC grows the constants-artifact manifest, exposing an intentional limitation in
  the existing epoch guard: it rejected every artifact-set change, including successor-RFC growth.
  The guard now permits append-only artifact additions only in a `BALANCE-CHANGE` mint; hotfixes,
  removals, and rewrites still fail closed. Repository-history tests cover both acceptance and
  rejection paths, and the baseline guard recognizes faction inputs.
- Focused Go suites are green. `make verify-schema` reached the existing `pnpm --dir client`
  mirror stall after the validator had already passed earlier in this session; it was interrupted
  rather than misreported as green and will be rerun in the final verification batch.

## 2026-07-30 — implementation batch complete, review gate open

- Minted Balance Epoch 2 with the new faction artifact and a dedicated changelog. The separately
  guarded baseline regeneration changed the constants identity and the then-current deterministic
  final-state hashes; pacing observations stayed unchanged. `make harness-check` accepted both
  guarded commits before the save-version work began.
- Save v10 persists nullable run faction/time and the three exact stock integers. The migration
  corpus grew with explicit v9 Company and Founder cases, scope/pairing validation rejects poisoned
  combinations, and catalog-aware persistence enforces stock cap/remainder bounds without making
  `save` import the feature package.
- Added exact-schema `incorporate`, strict `incorporated` and `faction_stock_saturated` events,
  Open Source's atomic Compact signature at 130,000 ppm, the bound-leave rejection, and derived
  stock-resource receipt fields. Identical persisted replays are asserted against Postgres.
- Stock accrual is a neutral post-accrual hook over the existing canonical elapsed result. It skips
  spans above the catch-up ceiling, carries the integer remainder, saturates only accrual, and emits
  the cap-crossing event once. No consume mutation exists in Phase 0.
- Leaderboard variables and the database exact-key check now include nullable faction; migration
  backfills existing immutable rows under a deliberately disabled/re-enabled user trigger. New-run
  assembly has a regression proving faction and stock state reset.
- The first Compose pass found three fixture assumptions (Epoch 1, eight run-log commands, and a
  v6 fixture derived from a v10 encoding). All were test-only and corrected; the faction SQL
  migration and production transaction themselves passed. The focused reconciler rerun and the
  complete production integration package are green.
- Replaced hanging pnpm script indirection in routine Make targets with the repository's installed
  client binaries. `make test-client` now completes normally (6,452 tests passed), `make typecheck`
  reports zero errors/warnings, and `make verify-schema` validates the faction catalog with the
  rest of the balance surface.
- Canonical docs now describe save v10, Epoch-2 identity, the eleventh production intent, strict
  faction events, board variables, run reset, and Phase-0's intentionally inert stock.
- Save v10 subsequently changed only the harness's encoded terminal-state identity. Regeneration
  reproduced the repository's save-v8 precedent: the two `final_state_hash` values move while the
  pacing baseline and every observation stay byte-identical. That state-envelope-only artifact is
  committed separately from implementation and documentation.
- Implementation remains `implementing` until the mandatory independent diff review is recorded;
  self-review and green verification do not satisfy that archive gate.
