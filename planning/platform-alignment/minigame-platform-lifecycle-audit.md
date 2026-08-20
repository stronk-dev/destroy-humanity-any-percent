# Minigame Platform Foundation lifecycle audit

Coordinate: product tree `190a4fa`; audit checkpoint after `baeaaa8`; 2026-08-20.

This pass re-derived the active Minigame Platform RFC from its complete specification and rulings,
implementation plan and append-only log, current Go/TypeScript kernels, migrations, production
catalog and composition, canonical docs, The Pitch and Minigame API successors, integration tests,
and tracked review verdicts. It did not edit product code, owner-authored RFC text, balance data,
canonical product docs, or implementation-plan checkboxes.

## Bottom line

The repository contains a substantial, cold-green minigame backend: durable claim-token sessions,
immutable command history, pure/versioned tenant dispatch, exact policy loaders, pinned catalog
identity, atomic server-certified resolution, faucet accounting, Founder rating/quality state,
cross-runtime replay, a production Pitch tenant, and authenticated HTTP composition. That is not
the same as satisfying this RFC's six acceptance criteria.

- AC2 is false at HEAD. Offline-quality decay has no observable transition or output consumer.
  Start reads the stored grade without decay; result resolution computes elapsed decay and then
  immediately replaces the grade and zeroes the remainder. A player who stops playing retains the
  old stored grade indefinitely.
- AC1 is partial. Solo Pitch completes through the composed gameserver and public API, but
  production start hardcodes `solo`; `async_snapshot` exists only in the platform enum, fixture
  descriptor, and validation tests.
- AC3 is partial. Cap/forfeit behavior and reduced-rate arithmetic are proven, but no bot tenant,
  bot session, or bot match exists to prove the literal reduced-rate match path.
- AC6 is false. Production registers only The Pitch; the duel RFC is draft and the current combat
  foundation explicitly has no battle engine.
- AC4 and AC5 have strong current integration witnesses, but the platform remains active because
  the failed/partial criteria, stale normative body, stale plan/docs, and cross-RFC closeout have
  not been reconciled.

This is a healthy mechanical foundation with two real production consumers (Pitch and the
authenticated API), plus material acceptance and lifecycle debt. It is not archival-eligible.

## Current cold evidence

All commands ran at repository root; Go claims used `-count=1`:

- `make test-go GO_PACKAGES='./minigame ./pitch ./replaycatalog ./production'
  GO_TEST_FLAGS='-count=1'` — green;
- `make test-client` — 39 files passed, two skipped; 6,655 tests passed, 15 skipped;
- `make typecheck` — TypeScript and Svelte checks green;
- `make test-save-integration SAVE_TEST_PACKAGES='./minigame ./production ./gameserver'` — green
  on real PostgreSQL.

The production Pitch integration proves unlock→start→play→certify→resolve→payout→durable retry,
with replay and Soul gating. The gameserver integration proves the authenticated HTTP lifecycle,
including reconnect and limiter behavior. The resolution integration proves the Founder→Company→
session transaction, failure rollback, faucet window, logs, events, and idempotency. These are real
solo/platform witnesses; none creates an async or bot session.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Partial** | Production `StartMinigameSession` rejects definitions whose first mode is not solo and always sends `ModeSolo`. Pitch's production row declares only `modes:["solo"]`. The fixture registry declares `async_snapshot`, but no composed start/play/resolve/payout witness uses it. Solo lifecycle and frozen inputs are strongly proven through Pitch/API successors. | Add an accepted real async-snapshot tenant/content row and composed lifecycle with an altered-live-state discriminator, or reconcile the RFC to defer that mode explicitly. |
| AC2 | **Failed at HEAD** | Ranked-power rejection and breadth acceptance are tested. Result resolution selects a grade. However, `decayOfflineQuality` is called only inside `ApplyFounderResolution`; its decayed grade/remainder are immediately overwritten by the new result. Start passes `MinigameOfflineQuality.GradePPM` directly into scaling. No production/code consumer applies `automation_destination` to automated output. Therefore lapse-time decay and the promised active/idle bridge do not exist. | Accepted remediation must define the observable decay/application boundary, prove no-play convergence toward neutral with partition invariance, prove new results replace the decayed sample, and bind the grade to its named automated output. Include a mutation that preserves the stale grade. |
| AC3 | **Partial** | Exact conversion, carried ppm, per-send cap, daily quota, forfeiture reason, rollback, and production Pitch payout are proven. `FallbackBot` and `rate_reduction_ppm` load, and `ConvertPayout` unit-tests reduction-before-conversion. No production row or tenant uses the bot arm, so no match proves that an actual bot-certified resolution selects that reduction. Epoch-guard discipline exists, but the literal bot-match chain is absent. | Land the first accepted bot tenant/content through its owning combat RFC and prove bot identity→session genesis→certified result→reduced payout→replay; retain an unreduced control. |
| AC4 | **Cold witness green; final range reconciliation missing** | The real-Postgres resolve fixture proves claim-token ownership, atomic Company/Founder/session writes, server-authored `resolve_minigame_session` logs/events, replay, rollback at each injected boundary, and idempotent terminal retry. The public endpoint requests resolution of an owned server session; its empty request cannot construct the internal certified payload/result. | Preserve the grep/schema negative and transaction faults; include the foundation plus API successor ranges in one exact archival review union. |
| AC5 | **Proven integration** | The loader rejects missing/malformed fallback rows. Production Pitch declares `{"kind":"solo"}` and completes through service and authenticated gameserver/Postgres tests with no peer or multiplayer dependency. | Preserve these discriminators in the eventual range review; no new mechanic is needed for the solo arm. |
| AC6 | **Contradicted at HEAD** | `gameserver/composition.go` registers only `pitch.NewTenant()`. `docs/combat.md` states there is no content catalog or battle engine; the shared combat plan leaves the catalog and full fixture open; the Duel RFC remains draft. | Ruling author reconciles the header, dependency, MP1/MP5, C3/C8/C9/C12/C14, and AC6. Duel registration belongs to the accepted Duel/Integration child after its engine exists. |

No plan checkbox was changed. The later Pitch, First Content, and API work makes two open plan rows
stale, but the checkbox law requires proof and the record flip inside an accepted designated-review
range, not an audit inference.

## Offline-quality defect trace

The defect is a complete producer→consumer break, not merely a missing edge-case assertion:

1. MP2/C10 require a stored grade that decays on Founder-attended time toward a nonzero neutral
   floor, so optional minigames do not become mandatory labor.
2. `server/minigame/resolution.go` calculates elapsed fixed-grid decay only when a new certified
   result resolves.
3. The same function immediately assigns the new score-derived grade, current watermark, and a
   zero remainder. The calculated decayed grade is never externally observable.
4. `server/production/minigame_start.go` copies stored grades directly into scaling without
   advancing them to the current attended watermark.
5. Repository-wide use tracing finds no other decay call and no automated-output consumer of
   `automation_destination`; that field is validated and serialized only.
6. The shared six-row Go/TypeScript fixture explicitly expects “decay then replace.” It proves
   runtime parity, but every expected final value is result-derived, so it cannot prove lapse-time
   decay. The TypeScript comment acknowledges the prior decay is evaluated only to reject invalid
   state before replacement.

The current suite is green because its oracle encodes the non-observable behavior. A valid
acceptance witness must create a charged grade, advance attended time without another minigame
result, observe the grade at its actual automated-output consumer, and fail if the stored value is
used unchanged.

## Producer→consumer reality

| Layer | Reality at HEAD |
|---|---|
| Session authority | Proven: PostgreSQL rows, frozen genesis, command log, claim lease, immutable terminal receipt. |
| Tenant boundary | Proven: pure exact-version registry, schemas/errors/modes, replay from genesis, divergence refusal. |
| Policy/data | Production schema-v3 Pitch row exists with solo fallback, payout, quality, rating, unlock, and Soul gate. No async or bot production row exists. |
| Economy/Founder composition | Proven for solo Pitch: atomic payout, cap/forfeit, rating/quality event and replay state. Quality charging is persisted; time decay/output application is absent. |
| HTTP consumer | Proven: authenticated create/current/command/resolve endpoints exercise Pitch end-to-end. |
| UI consumer | Absent under the active Minigame API & Surface plan: no mounted `minigame_session` component or player-playable Pitch surface. |
| Combat consumer | Absent: shared data arithmetic is incomplete and neither battle engine exists. |

The platform backend is therefore stronger than its stale plan suggests, while the designed player
capability is weaker than “implemented platform” language in successor records suggests.

## Normative, canonical, plan, and successor drift

1. The RFC header says Combat Shared Kernel is implemented and the duel engine is tenant #1.
   MP1/C8 say combat already proves both Phase-A modes; MP5/C3/C9/C12/C14 and AC6 repeat the duel
   adapter claim. Runtime and combat authority falsify all of those current-tense statements.
2. MP5's advertised registry shape (`clock`, top-level `scaling_inputs`, `unlock_condition_ref`)
   is not the implemented immutable definition schema (`modes`, nested `scaling`, typed
   `unlock_condition`, rating/offline rows). C12's “body reconciled” status is therefore false for
   more than the duel sentence.
3. The plan still leaves gameserver solo/async composition and production artifact mint unchecked.
   Successors later composed solo Pitch/API and minted Pitch in epoch 6, but no async path landed.
4. `docs/minigame-platform.md` first says the conformance tenant is test-only, no production policy
   row exists, and decay/policy literals are disabled. The same page later describes the composed
   production API, immutable complete artifact, and live resolution/replay transition. Current
   canonical documentation contradicts itself and the production Pitch row.
5. The archived Pitch RFC declares Minigame Platform “implemented + archived,” while the active
   index and platform RFC correctly remain `implementing`. Pitch itself is genuinely reviewed and
   archived; its dependency status claim is stale.
6. Platform review entries provide designated cross-party coverage for substantial foundation
   slices through the resolve/type remediation. Pitch and API successors have their own designated
   verdicts. The active platform log ends before those successors and contains no final criterion
   trace or exact union explaining which cross-RFC ranges satisfy AC1–AC6. Separate valid reviews
   are not automatically one platform archival verdict.

## Smallest honest closeout order

1. Ruling author reconciles the false combat dependency/current-tense claims and the stale MP5
   schema. Decide explicitly whether async remains a foundation criterion or moves to its named
   first content successor.
2. Under the accepted platform RFC, repair AC2's observable decay/application contract with
   discriminating Go/TypeScript/replay/production witnesses. Do not invent balance values.
3. If async remains in scope, add only the missing accepted async content and lifecycle witness;
   otherwise correct AC1 and route it to the already named Pitch async successor.
4. Keep the bot-match requirement blocked on the Combat Duel/Bots dependency; schema arithmetic is
   not permission to fabricate a bot tenant.
5. Reconcile canonical docs and plan state against the reviewed Pitch/FCE/API successors, without
   claiming the still-absent UI surface.
6. Build an exact post-rewrite review-range union for every foundation and successor commit the RFC
   consumes; obtain the mandatory tracked cross-party verdict over the remediation/closeout span.
7. Archive only after all in-scope criteria, docs, plan, log, RFC status, and successor references
   agree transactionally.
