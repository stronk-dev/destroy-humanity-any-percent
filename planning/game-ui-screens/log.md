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
