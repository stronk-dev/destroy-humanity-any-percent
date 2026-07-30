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

## 2026-07-30 — amplitude-lock boundary correction

- The first complete `make verify` caught a transitive `production -> faction -> commons` import.
  Gameplay behavior and the atomic Open Source mutation were correct, but the faction loader's
  cross-catalog validation had made the package graph violate the existing Commons amplitude-lock
  gate.
- Replaced the concrete Commons catalog parameter with faction's value-only `CompactTitheBand`.
  Composition still derives those three bounds from the immutable Commons catalog, while the
  production dependency graph can no longer reach the Commons implementation. The boundary gate
  and focused faction/production suites pass after the correction.
- Final executable verification is green: `make verify` passed Go vet/tests, formula drift,
  deterministic harness/history checks, all import boundaries, strict TypeScript and Svelte
  checks, the production build, 6,452 Node tests, every schema gate, and 19,365 browser tests.
  `make test-save-integration` then passed every integration package against the Compose-managed
  PostgreSQL 16 service. The only remaining step is the mandatory independent complete-diff review.
- The final acceptance mapping also pins Phase 0's inert-consumer contract directly: faction
  accrual is tested with a non-zero `consumed_stock_units` value and must preserve it across both
  attended and offline evaluations. The Guild successor may replace that assertion only when its
  evented clearing mutation lands.

## 2026-07-30 — independent complete-diff review (9b5f4b3..91d089b)

Two-lane review: mint/catalog lane by the reviewer directly; intent/accrual lane adversarial with
probe confirmation. **Verdict: APPROVED with two MEDIUM findings, both now ruled below. Guild may
start immediately; the two rulings join the top of its implementation queue.**

Reviewer-verified (mint lane): the epoch-2 mint is protocol-compliant (BALANCE-CHANGE subject,
appended epoch + changelog same-commit, separate baseline BALANCE-CHANGE); the artifact-growth
guard relaxation is growth-only + mint-only with prefix DeepEqual and uniqueness over the running
set (removal/rename/reorder still fail); `make epoch-hash` at HEAD reproduces the epoch-2 accepted
hash exactly; the catalog loader is FC to the letter (strict decode, closed sets, single-cycle,
open_source-only compact with band validation via the value-only CompactTitheBand boundary — and
f506daa's "amplitude boundary" is the real `verify-commons-boundary` build-graph gate, run green).

Agent-verified with evidence (intent lane): C1 conformance incl. byte-identical replay;
incorporate/Exit serialization via stream FOR UPDATE + revision CAS; catalog resolved at the RUN's
hash everywhere; one-transaction OS binding with leave→`faction_bound`; FB accrual exact (offline
skipped with P6-mirrored boundary, carry-while-saturated, once-per-crossing saturation event);
all 8 evaluation sites paired with the hook; save v10 closed field list with pairing/scope-leak
negative tests; migration 00023 faithful incl. Down; tooling change weakens nothing.

Findings and rulings:

1. **MEDIUM — unspecified path: an existing compact signatory who incorporates Open Source is
   rejected `already_member`** (undeclared category for this intent; the only path forward would
   reset Solidarity — perverse for the most-committed player). **Ruling F2a: incorporation
   CONTINUES membership** — the tithe is raised to `max(current, faction tithe)` (never lowered),
   Solidarity and membership history are preserved, and the transition emits `incorporated` +
   `compact_tithe_raised` (new kind, registry grows by this ruling) in the same commit.
   `already_member` is removed from this path's rejection set.
2. **MEDIUM — attended-time ceiling is duplicated, un-cross-checked, and not hash-pinned**
   (faction and prestige runtimes each take their own `catchupCeilingMS`; a mismatch splits the
   attended/offline definitions, and replay under run-genesis would be process-config-dependent).
   **Ruling FB-1/P6c: `catchup_ceiling_ms` becomes a field of the PRESTIGE catalog artifact**
   (P6 owns the attended-time definition; the artifact is hash-pinned) — both runtimes read it
   from the run's resolved prestige policy; the two ServiceOptions collapse to one; `NewService`
   fails if any secondary source disagrees. Client-shell keeps its copy for display pacing only.
3. LOW batch (queued, non-blocking): `run_ended` should carry `faction` (F3's variable currently
   has no producer — acceptable pre-verifier, but add the field now so the verifier never scans);
   wire snapshot exposes `stock_progress_ms` beyond FA's declared three (RFC amended to admit it —
   display-useful, harmless); unknown-faction rejection needs a transition-level test and the
   check-order (unknown_id before tier gate) pinned; the inert-consumption assertion should
   enumerate stock-field writers structurally; `StatePolicyValidator` is opt-in by type assertion
   — the composed gameserver must assert it at construction (fail-closed), noted for composition;
   AC4's next-run re-incorporation and an incorporated-company-through-real-exit test are missing;
   accrual-hook chain order must be pinned before run-genesis lands (byte-stable event order).
