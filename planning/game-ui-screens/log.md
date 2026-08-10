# Game UI Screens implementation log

## 2026-08-10 — Codex acceptance review

The UI Foundation substrate is implemented at `3483ab1`, but the screen draft cannot yet compile
against the repository's actual data boundary. The shipped shell snapshot lacks most Desk/run
fields; required event payloads are not generated typed unions; no literal surface/fact or Copy
manifest exists; first-session bootstrap and local timer/PB persistence are directional; the era
claim contradicts the two-era first hour; and the performance gate has no executable budget.

GU-C1–GU-C8 in the RFC propose exact ownership shapes without choosing product copy or balance
content. Status remains draft; no screen code was improvised.

## 2026-08-10 — Codex contract/composition batch, ready for designated review

Implemented the unblocked mechanical half of GU-C1–GU-C8:

- one generated `game_ui_snapshot.v1` projection over the latest Company revision, its pinned
  replay bundle, the Founder sibling, and composed contribution providers;
- the existing authenticated `GET /api/v1/founder/state` sync route now returns that closed DTO
  instead of exposing raw save bytes; the gameserver composition attaches the real projector;
- exact mechanical-ID-sorted server rows and client fail-closed validation, plus an adapter into
  the existing `AuthoritativeSnapshot` consumed by `ShellRuntime`;
- the five-row surface registry, lifecycle-event cursor precedence, destructive-settings deferral,
  server-sampled monotonic RTA, corrupt-local-PB fallback, closed tier→era mapping, and GU-C8's
  executable performance literals;
- typed decoding for the actually shipped gate/run-end/offer/system event bytes. Declined and
  expired offers normalize to the client `exit_offer_resolved` arm; accepted resolution remains
  blocked by GU-C11 rather than fabricating a missing offer ID.

The projection deliberately rejects schema-v4 or Company-v18 state because the public production
surface cannot assemble content/active-play contributions. GU-C9 records the required kernel-owned
read model. GU-C10 records the idempotent-bootstrap gap exposed by the shipped two-operation account
flow. GU-C12 records the already-requested owner-copy dependency. No Svelte screen has been filled
with placeholder prose.

Focused evidence, all through ordinary root Make targets:

- `make test-go GO_PACKAGES='./gameui ./account ./gameserver' GO_TEST_FLAGS='-count=1'`
- `make test-client` — 6,640 passed, 4 skipped
- `make typecheck` — 0 TypeScript/Svelte diagnostics
- `make test-save-integration` — the full real-Postgres integration suite passed with cache bypass;
  its first run exposed two composition-only seams that focused tests could not: epoch-6 saves
  legitimately omit the pre-schema-v4 provisioned-count map, and `LoadLatest` leaves
  `Revision.OwnerID` empty even though contribution providers require the Founder owner. Both now
  have regression coverage in the composed snapshot path.
- `make api-pin` — generated private-v1 schema/types/compatibility pin refreshed for the existing
  Founder-state route

This batch is not an archival claim. It is ready for the cross-party designated review after its
commit range is recorded; GU-C9–GU-C12 remain explicit blockers to the full screen acceptance gate.

The implementation commit is `1e8f628`. Post-commit `make verify` completed successfully through
the generated API pin, composed harness, TypeScript/Svelte typecheck, production build, 6,640 client
tests, kernel-history guard, copy/schema checks, and 19,932 browser assertions. The designated
review range for this batch is `ddb106f..1e8f628`; it contains only this Game-UI implementation
commit.

## 2026-08-10 — designated cross-party verdict: contract/composition layer {1e8f628} — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
All eight GU rulings verified implemented-or-honestly-deferred: the DTO on the EXISTING route
(which this commit converts from a hand-mounted raw-save-bytes chi route to the registry-mounted
schema-validated operation — closing a save-state escape, the MA-C14 discipline); five-kind exact
event unions with the honest pre-GU-C11 offer handling; RTA/PB per ruling; the five-surface
registry with tested preemption; GU-C8 literals count-asserted with a red probe; BOTH composition
seams genuine with discriminating tests (pre-v18 zero-provisioned; the OwnerID fill that fails
composed-end-to-end when reverted); schema-v4 rejection fail-closed; kernel honest; all gates
green incl. cold -count=1 and the full Postgres suite.
**Routed to the GU-C9/screens batch (non-blocking): C-F1 (MEDIUM) — a store-level test driving
GameUISnapshot itself over a v4 catalog/v18 stream (the predicate alone is tested today); C-F2
(LOW) — wire the CI-observable performance arm when the first screen ships. Plus the copy
ruling's renderer-substitution requirement (mechanical IDs never render) as a GU acceptance test.**
Range D {00cb2a6} records: APPROVED (honest scoping, clean provenance). Combined union:
a8bbd5d..00cb2a6 complete, no uncovered edge commits.

## 2026-08-10 — GU-C12 mechanical assembly review

- **Review by:** Codex (implementer-side acceptance review). **Recorded by:** Codex.
- Mechanical extraction proves 130 affected rows = 129 new keys plus the adopted replacement of
  existing `resource.company_cash.cap.phase0`; no duplicate authored key exists.
- The shipped exit-type union was independently enumerated from both Go and TypeScript as
  `acquihire`, `acquisition`, `collapse`, `ipo`, `scripted_first`.
- Fixed one unambiguous Copy validator defect discovered by the extraction: valid underscore
  placeholders such as `{run_seq}` were mistaken for Markdown emphasis even though Copy C3's
  parameter grammar explicitly permits underscores. The validator now removes only lexically
  valid placeholders before its markup probe; an explicit regression passes `make copy-check`.
- GU-C13–GU-C16 record the four remaining owner-contract gaps: README slot grammar, mandatory
  provenance collisions, exact param types, and presentation-v2/exit-title bytes. No candidate
  SHA is claimed until those are ruled.

## 2026-08-10 — GU-C9 implementation handoff — ready for designated review

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementer handoff, not an
  independent verdict or archival authorization.
- **Implementation commit:** `bc2370a` (the complete GU-C9 implementation range). The production
  kernel now exposes one pure `ProjectRates` read model over the pinned bundle, replay-owned
  Company state, resolved external contributions, and attended coordinate. It assembles the exact
  existing upgrade/ladder/synergy/active-play contribution paths, including the product clamp,
  and returns byte-sorted generator/resource rows without mutating a cursor or producing state,
  events, receipts, or replay inputs. Game UI consumes this seam and no longer reconstructs rates.
- **C-F1 closed:** a real-Postgres test persists a schema-v4, Company-v18 stream and drives
  `GameUISnapshot` itself. Its expected rows require purchased + provisioned counts, the ladder and
  synergy paths, and a live active-play factor; reverting the GU-C9 seam restores the former hard
  rejection and fails the test.
- **R-F1 closed in the same honest kernel bump:** `toFloat64` now crosses a non-inlined float64
  materialization boundary before the snap comparison, preventing arm64 FMA fusion from changing
  the decision. The mandatory shared `floor-fma-snap` vector is consumed by Go and TypeScript;
  `make test-go-ci` proves it on cold Linux/amd64. Kernel `0.3.89 -> 0.3.90` is intentional.
- **R-F3 closed:** `docs/ci.md` now names `make test-go-ci` and its production-architecture scope.

Normal root-target evidence on the exact implementation tree:

- `make test-go GO_PACKAGES='./decimal ./production ./gameui' GO_TEST_FLAGS='-count=1'` — green;
- `make test-save-integration` — every real-Postgres integration package green;
- `make test-go-ci` — every Go package green, cold, on Linux/amd64 with Postgres;
- `make vectors-check` — 6,296 deterministic vectors, 23 mandatory edges, no drift;
- `make verify` — full aggregate green before commit: 6,641 client tests, 19,935 browser
  assertions, zero TypeScript/Svelte diagnostics, and all guard/harness/schema/generated checks;
- post-commit `make verify` repeated every gate through schema; its first browser-server startup
  was denied by the execution sandbox (`listen EPERM`, before tests). The ordinary root
  `make test-browser` target was immediately rerun and passed all 19,935 assertions. No product
  test failed and the committed tree is clean.

GU-C9 is ready for the required cross-party designated review. GU-C10, GU-C11, and GU-C12's
GU-C13–GU-C16 owner blockers remain open; nothing is self-approved or archived.

## 2026-08-10 — GU-C10 acceptance review — owner blockers GU-C17–GU-C22

- **Review by:** Codex (implementer-side acceptance review). **Recorded by:** Codex.
- GU-C10's direction is accepted, but the executable wire, possession boundary, secret-at-rest
  policy, atomic coordinator transaction, receipt expiry, and supposedly mismatched request are
  not determined. The sharp contradiction is exact durable replay of recovery/refresh secrets
  against the shipped hash-only credential law.
- GU-C17–GU-C22 are filed in the RFC with recommended contracts. No coordinator, migration, API
  row, or secret-storage behavior was invented. GU-C10 remains blocked pending owner rulings;
  GU-C11 stays mechanically independent and may proceed after GU-C9's designated handoff.

## 2026-08-10 — GU-C11 implementation handoff — ready for designated review

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementer handoff, not an
  independent verdict or archival authorization.
- **Implementation commit:** `607e5a2`. Accepted offers now emit exact
  `exit_offer_resolved` schema v1 bytes `{offer_id,resolution:"accepted"}` immediately before
  `run_ended` in the same Company revision. Declined and expired rows are unchanged.
- Go and TypeScript produce the same event and order in the shared terminal replay fixture. The
  Game-UI decoder extends its normalized resolution union without manufacturing a run sequence
  for the accepted arm.
- Migration `00072_exit_offer_resolved.sql` extends the closed production event-kind constraint;
  the live Prestige integration test writes the event to PostgreSQL, verifies its exact two-key
  payload, and proves its order before `run_ended`.
- Kernel `0.3.90 -> 0.3.91` is an honest transition/event/replay semantic bump. The canonical
  Prestige documentation records the additive event contract.
- The repository now has a normal `make validate-migrations` target for cold real-Postgres
  validation of migration and save integration tests; no ad-hoc database command is required.

Normal root-target evidence on the exact implementation tree:

- `make test-go GO_PACKAGES='./save ./production' GO_TEST_FLAGS='-count=1'` — green;
- `make replay-fixture-check` — green, including the discriminating accepted-event assertion;
- `make test-client` — 6,641 passed, 4 skipped;
- `make validate-migrations` — green on real PostgreSQL;
- `make test-save-integration` — every integration package green;
- `make test-go-ci` — every Go package green, cold, on Linux/amd64 with PostgreSQL;
- `make verify` — full aggregate green: 6,641 client tests, 19,935 browser assertions, zero
  TypeScript/Svelte diagnostics, and all guard/harness/schema/generated checks;
- post-commit `make verify-kernel-version` and `make replay-fixture-check` — green. A combined
  shell invocation's trailing Docker segment was sandbox-denied before execution; the ordinary
  standalone `make validate-migrations` target was rerun immediately and passed.

GU-C11 is ready for the required cross-party designated review. GU-C10 and GU-C12 remain blocked
only on their recorded owner rulings; nothing is self-approved or archived.

## 2026-08-10 — designated cross-party verdict: GU-C11 {607e5a2, 349757a} — BOTH APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
The ruled contract implemented exactly: emission guarded on the accept intent, appended
immediately before run_ended in the same batch (same revision structurally); ordering
discriminated TWICE in both runtimes (order-swap probes fail the corpus byte-compares and the
in-test assertion); the two pre-existing accepted-offer corpus cases correctly updated in place
with every other case byte-identical; declined/expired paths byte-untouched; both-runtime
lockstep validation with negative tests; migration 00072 append-only, mechanically verified as
00068's Up + exactly the new kind, Down byte-restores; `make validate-migrations` sound and
justified; kernel 0.3.90→0.3.91 honest; all five gates green incl. cold Linux.
**Carried obligation:** when the GU-C3 generator lands, it must own the hand-added union arm.
**Range-union: this verdict covers exactly {607e5a2, 349757a}. NOT covered and pending:**
{bc2370a, 46676e9} (GU-C9 — its own review, launched same day with golden-vector scrutiny) and
{f2cecba, 3e16587} (copy-assembly validation + blocker filings — consumed by the ratification
and ruling flows). The union over 516bddb..HEAD completes when those land.
