# Minigame API & Surface implementation log

## 2026-08-07 — composition slice

- `gameserver.Compose` now constructs the Postgres minigame repository, the closed tenant registry
  containing Pitch `1.0.0`, and the platform service using the immutable runtime catalog set as its
  `TenantContentResolver`.
- `Composition.Minigames` exposes the real service. It owns no background work, so no drain job is
  added.
- Focused `go test ./gameserver ./minigame ./pitch` is green. The existing real-Postgres composed
  test asserts the platform is present; the endpoint-level socket proof remains open until the
  implementation blockers in the RFC are ruled.
- This is implementation evidence, not an independent verdict and not archival authority.

## 2026-08-07 — designated cross-party verdict: MA-C2 composition slice {e935c22} — APPROVED (scoped)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Reviewed range:** `e935c22^..e935c22` (two files, both on-scope; pinned worktree; HEAD had moved).

Verified: repository/registry/service wiring correct; the tenant registry is CLOSED at exactly
Pitch 1.0.0 (post-construction immutable); resolver identity is PINNED per constants-hash bundle
with per-session coordinates (no process-current reads; mismatch rejection package-test-proven);
platform exposed on `Composition`; no background work (grep-verified — the "prove none" arm is
satisfied by the recorded claim + code fact); kernel correctly untouched (`server/gameserver/`
unwatched; guards green at 0.3.78); focused Go suites + full Postgres integration suite
independently green at the pinned commit. Probe A: removing the `Minigames` field fails the
composed test — the presence assertion discriminates.

**MEDIUM finding, tracked as a BINDING acceptance condition for the coordinator slice (not
blocking this range):** Probe B — fully severing `runtimeCatalogs.ResolveTenantContent` (body →
unconditional false) leaves EVERY test in the repo green. The composed test proves presence only;
the resolver delegation, repository transactions, and lifecycle are honestly scoped as open
(MA-C10–C14, plan boxes unchecked) — but the composed lifecycle test that closes MA-C2's
"drives the endpoints" clause MUST fail on a Probe-B-class resolver break, or the pass-through
stays permanently unproven at the composed level. Cited here so the coordinator-slice review
checks it by name.

Minor: commit subject carries no design/RFC section reference (working-convention nit); the
plan-box flip lives in 77aa982 (impl+record pattern, compliant once the record commit is consumed
by the normal record-review flow — 77aa982 and 2d36244 are NOT consumed here).

**Verdict: APPROVED for {e935c22}, as the honestly-scoped composition-only slice.** Range-union:
extends the Minigame Platform designated chain (through pre-filter 06bf0f3 = post-filter c2959a1)
to the previously unbuilt AC1 composition wiring; intervening minigame/pitch code commits
(0eb3772, 2a55e12) belong to The Pitch's separately consumed set. The coordinator adapter,
wire contracts, idempotency/sequencing, and the composed lifecycle proof remain open behind the
now-ruled MA-C10–C14.

## 2026-08-07 — MA-C15 pinned activation candidate and Founder v21 codec

- Added the byte-exact `minigame_api` schema-v1 candidate at
  `balance/testdata/minigame-api-candidate-v1.json`; SHA-256
  `b16b5e0eb6f9426c8b1b94255e2d8e04f53f78b391fdbbb348ad7438d7bab31c`.
- Both replay runtimes now accept `minigame_api` only above the complete Pitch chain, validate its
  closed operation/tenant rows, and derive Founder floor 21 exclusively from its pinned presence.
- Founder v21 adds exactly `minigame_session_seq`; Go and TypeScript codecs reject missing,
  premature, or Company-scoped state, and Exit activation initializes the field to zero.
- Kernel semantics advance from 0.3.78 to 0.3.79. The new Go authority is registered in the
  fail-closed kernel path list; focused Go suites, strict TypeScript checks, and kernel-history
  guards are green.
- This is implementation evidence and a First Content Epoch candidate, not a mint, independent
  verdict, archival decision, or completion claim for MA-C11's still-unbuilt sequencing command.

## 2026-08-07 — MA-C5/MA-C6/MA-C13 receipt persistence primitives

- Migration 00071 makes the accepted one-active-minigame-per-Founder rule a partial-unique database
  invariant and adds immutable Founder-scoped create receipts plus session-scoped command receipts.
- Repository reads distinguish missing, same-key/same-hash replay, and same-key/different-hash
  conflict. The current-session query returns only the sole `active|claimed` row.
- Nonterminal command, snapshot/revision, and canonical API response can now commit in one
  transaction. Terminal callers receive a transaction-scoped insert helper whose `INSERT SELECT`
  is guarded by the exact claim token, so an expired worker cannot publish stale bytes.
- Real-Postgres tests prove cardinality, retry/hash conflict, immutability, atomic nonterminal
  commit, and stale-token rejection. The full save integration suite is green. Kernel semantics
  advance from 0.3.79 to 0.3.80.
- The coordinator and typed handlers still need to consume these primitives; no completion or
  archival claim is made here.

## 2026-08-07 — MA-C11 replay-owned session-start coordinator

- Added the server-only `start_minigame_session` Founder command and an atomic
  Founder→Company→session coordinator. The transaction locks the active Founder and both save
  streams in the declared order, validates their shared pinned run, advances only Founder v21's
  `minigame_session_seq`, inserts the frozen tenant genesis/session and create receipt, then writes
  the Founder revision/log before committing.
- Session seeds are derived from the persisted sequence and the save-seeded substream contract in
  both runtimes. The shared replay corpus proves the exact Founder transition, receipt, and seed;
  no deploy-current content is read inside replay.
- Real-Postgres coverage injects faults after Founder genesis, revision, events, log, and retention
  and proves that no session, receipt, or sequence increment survives. It also proves same-key
  replay without a second sequence increment, conflicting active-session rejection, Company-state
  immutability, and verification of the resulting Founder history.
- Kernel semantics advance from 0.3.80 to 0.3.81. The standard Go, client, typecheck, replay-fixture,
  kernel-history, and full Postgres integration gates are green.
- This is implementation and self-review evidence only. The public create handler is intentionally
  still unmounted, MA-C12 and terminal command composition remain open, and this entry is not the
  designated cross-party verdict or archival authority.

## 2026-08-07 — MA-C12/MA-C13 command and Exit coordinator slice

- Added the typed, unmounted command coordinator. It checks the durable `(session_id,command_id)`
  receipt before tenant execution; an equal request hash returns the stored bytes and a different
  hash fails deterministically. Nonterminal tenant state, revision, command row, and exact API
  response commit together under the claim token.
- Terminal commands auto-resolve in the same request. The composite terminal API response and
  command receipt now commit inside the existing Founder→Company→session transaction, while the
  byte-compared Company replay log and durable session retain the exact underlying resolution
  receipt. A stale claim cannot publish either representation.
- Added the MA-C12 read-only active-session resolver to composition. Exit freezes its result in
  replay inputs only when the pinned `minigame_api` artifact is active; both runtimes reject
  `not_eligible/minigame_session_active` before evaluation and restore byte-identical state.
- Real-Postgres tests prove terminal auto-resolution/retry/hash conflict, rollback of the command
  receipt under an injected Company-revision fault, and active-session Exit with unchanged Founder
  and Company revisions. The shared Go/TypeScript corpus pins the frozen Exit rejection.
- Kernel semantics advance from 0.3.81 to 0.3.82. This is self-review evidence only: API mounting,
  generated schemas, current-session/resolve handlers, and the Probe-B composed socket lifecycle
  remain open, and no archival claim is made.

## 2026-08-07 — MA-C10 typed-handler and composed-adapter slice

- Added the four exact typed handler shapes without mounting a handwritten route: create accepts
  only `{idempotency_key}`, command accepts only
  `{command_id,expected_revision,command}`, current accepts no body, and resolve accepts `{}`/empty.
  Founder, Company, clock, seed, engine, mode, scaling, and constants identity remain server-owned.
- Added one composed adapter binding the authenticated account to the already-reviewed start/play
  coordinators and the platform's current/resolved reads. Current returns the closed `none|active`
  union without claim state; resolve returns only the reconstructed stored terminal API receipt and
  rejects active sessions.
- Enumerated the deterministic HTTP mapping in one handler table. Tenant rejections, unlock/Soul,
  exclusivity, idempotency, missing session/founder, claim/revision conflict, invalid input, and
  unknown tenant each have literal status/category/detail behavior; every unlisted/store error is
  `500 internal_invariant`.
- Closed Soul Recovery's carried typed-sentinel rider in this first subsequent server/account
  behavior commit. The API now uses `errors.Is` for recovery-token and not-ready decisions; no
  error-string classification remains on that path.
- Kernel semantics advance from 0.3.82 to 0.3.83. These handlers remain deliberately unmounted:
  API Foundation registration/generation, the combined real-socket lifecycle (including the
  Probe-B resolver-break discriminator), privacy conformance, designated review, and archival all
  remain open.

## 2026-08-07 — MA-C14 authenticated registry and generation slice

- Registered the four minigame operations and the four archived Soul Recovery operations in one
  immutable private-v1 authority. The chi router now mounts those handlers only through the
  registry's declared access-token middleware; there is no parallel handwritten route table.
- Added exact request/response descriptors, path-parameter formats, runtime fixture validation,
  canonical OpenAPI 3.1 and TypeScript generation, and the committed additive-only v1
  compatibility pin. The ordinary `make api-check` target regenerates without mutating the pin;
  an intentional baseline update is the separate `make api-pin` operation.
- Extended the authenticated real-Postgres account test through the actually mounted minigame
  create route and validates the returned bytes against the same registry. This proves routing and
  auth/schema conformance, but it deliberately uses the typed adapter stub: the Probe-B composed
  create → command → terminal → retry/current lifecycle remains open and its plan box is unchanged.
- Focused Go tests, API drift generation, TypeScript typecheck, and the full save integration suite
  are the required evidence for this slice. This is implementation/self-review evidence only,
  ready for the later combined designated review; it is not approval or archival authority.
