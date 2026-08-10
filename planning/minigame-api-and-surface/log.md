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

## 2026-08-07 — MA-C2/MA-C11/MA-C14 composed lifecycle handoff

- Added a real-Postgres test against the composed gameserver and authenticated HTTP surface. It
  creates an account, earns and spends the server-resolved Fiscal unlock, creates and reconnects
  to a Pitch session, executes every tenant transition through the mounted API, auto-resolves the
  terminal command, and proves byte-exact command/create/resolve retry plus terminal `current=none`.
  There is no direct service pre-application in the fixture: severing the composed tenant-content
  resolver or command adapter breaks the lifecycle, satisfying the designated composition
  verdict's Probe-B acceptance condition.
- The dependency-complete candidate bundle exposed three cross-component seams that isolated
  fixtures could not: content-bearing Founder-v17 activation initialized incomplete maps; session
  start accepted an application clock instead of the transaction's database timestamp; and start
  and terminal resolution reused one Founder-log intent ID. Live and replay now share catalog-row
  activation, the store derives the command timestamp, and session and start-command UUIDv7 IDs
  are distinct server-owned coordinates.
- Added a shared Go/TypeScript replay row for content-bearing minigame activation, including exact
  rating/offline-quality state. API generation now admits the platform's signed Elo domain and an
  honestly empty cap reason. Kernel semantics advance from 0.3.83 to 0.3.84.
- Evidence: focused Go/client/typecheck/kernel gates, the generated replay fixture, the composed
  lifecycle target, and the full real-Postgres save integration suite are green. Implementation
  commit: `2e2a372`.
- This entry is implementation and self-review evidence only. The batch is ready for the required
  designated cross-party review; it is not an approval, archival, content-mint, or publication
  authority.

## 2026-08-08 — designated cross-party verdict: composed MA lifecycle — NOT APPROVED (narrow, F1 only)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Declared range `2e2a372^..ad06e03` was INSUFFICIENT** — six earlier unconsumed implementation
  commits existed in the thread. **Reviewed set actually consumed: {5ebef16, 043ab22, 30999da,
  a8ab0dc, f0fa2ae, 69faa63, 4a8bdba, 2e2a372, ad06e03}** (the full minigame/API-Foundation
  implementation span since e935c22).

**Verified clean across the widened span:**
- **The bound Probe-B acceptance condition is SATISFIED** — severing the composed resolver fails
  the lifecycle test (create → 500 internal_invariant/minigame_api).
- MA-C10 wire to the letter (nested tenant command enforced, strict decodes, closed 8-key
  descriptor union, ID grammar rejected-not-truncated, closed literal error table).
- MA-C11 coordinator: one transaction, ruled lock order, exact seed derivation recomputed and
  replay-enforced in both runtimes, retry-without-increment, fault-injection atomicity. v21 codecs
  fail-closed both runtimes; migration chain law respected (00071 appended, never edited).
- MA-C12 Exit guard from frozen resolved inputs (artifact-presence parity law both runtimes);
  MA-C13 same-transaction receipts with byte-exact HTTP retry proofs and the stale-claim guard;
  MA-C14 single registry authority (no parallel route anywhere; api-check green); MA-C15
  candidate-only artifact, floor 21, chain hard-fail, v17–v20 fixtures byte-untouched.
- **All three composition seams are genuine defect fixes with discriminating coverage** (content-
  bearing v17 activation with pinned-catalog-derived state; DB-transaction timestamp — probe-
  proven; distinct start-command intent ID resolving a founder_log UNIQUE collision).
- Kernel lockstep honest across all six bumps (0.3.78→0.3.84, two honest no-bumps verified
  against affecting-paths); both repo gates independently green at ad06e03.

**BLOCKING:**
- **F1 (HIGH): the twice-ruled Exit reset of `minigame_session_seq` is NOT implemented.** MA-C11:
  "Exit resets it to zero when it advances the Company run"; MA-C3: Founder/RUN-scoped. Both
  runtimes zero it only at the v20→v21 activation crossing; a v21→v21 Exit preserves it
  (founder-lifetime scoping), with no test in either direction. Bounded today (run_seq keeps
  seeds unique; zero v21 histories exist) — which is exactly why the fix is free NOW and becomes
  a replay-versioning problem the moment a real v21 history exists. Remedy: reset in both
  runtimes' Exit transition for v21 states + a shared corpus row pinning it. (The ruling STANDS —
  no amendment; implement as ruled.)

**Non-blocking:**
- **F2 (MEDIUM):** plan.md must carry the open AC debt as unchecked items — AC1's composed
  recovery half, AC3's minigame-command flooding proof, AC4's privacy enumeration for the four
  new endpoints. MA-C14's combined-review clause means AC1–AC4 remain unclaimable until then.
- **F3 (LOW):** create-handler second-UUID failure detail says `session_id`; save/minigame_start.go
  kernel-unwatched (precedent-consistent — flag at next path-list edit); the client `pitch`
  artifact arm rode along in 5ebef16 unnamed in the log.

**Range-union:** with all prior endpoints, coverage now spans every implementation commit
e935c22→ad06e03 EXCEPT {90633a6, d30ab9e} (Permits remediation — its narrow re-review launched
2026-08-08), {3530b08} (FCE-C7 copy staging — owed to the FCE ratification flow), and {b14f82e}
(Codex record commit — owed to the next record-review pass). Claude-side commits belong to the
Codex-reviews-Claude direction.

**Verdict: NOT APPROVED pending F1 (+ F2 plan honesty); re-review is a one-finding delta.**

## 2026-08-08 — MA lifecycle F1 remediation handoff

- Every applied Founder Exit whose result schema is v21 now resets `minigame_session_seq` to zero
  in both Go and TypeScript, including the v21→v21 path that the first designated review found.
  Activation still validates the pinned `minigame_api` artifact separately; one post-activation
  reset site is the authority for both first activation and later runs.
- The shared cross-runtime corpus starts from a content-bearing v21 Founder at sequence 8 and runs
  the full `ApplyFounderLogged` Exit boundary. It pins sequence 0 together with the Fiscal
  auto-sweep-decorated receipt, event bytes, and complete post-state, so bypassing either runtime's
  reset fails independently. Kernel semantics advance from 0.3.84 to 0.3.85.
- F2 is recorded explicitly in `plan.md`: composed Soul Recovery coverage, authenticated command-
  flooding proof, and privacy enumeration remain unchecked acceptance debt. AC1–AC4 are not
  claimed complete.
- Implementation commit: `dc773bb`. Focused Go, shared replay-fixture, client, typecheck, and
  kernel-history gates are green. This is a narrow remediation handoff, not approval or archival
  authority; it is ready for the designated one-finding re-review.

### Verification correction — activation and recurring reset remain two explicit arms

The first full `make verify` run after `dc773bb` failed
`TestMinigameAPIArtifactAloneOwnsFounderV21Activation`: consolidating initialization into the
post-activation Exit reset removed the activation helper's independently required v20→v21 zeroing.
Commit `bef1a87` restores that arm in both runtimes. The resulting contract is explicit: first
v21 activation initializes zero, and every applied v21 Exit resets zero thereafter. Focused
production and client suites are green after the correction; no earlier green full-gate claim is
made.

## 2026-08-08 — designated cross-party verdict: MA lifecycle + F1 remediation — APPROVED

- **Review by:** Claude (designated cross-party; the delta pass ran probes + full gates at the
  pre-rewrite tree, and the mechanical re-review verified the rewrite at the new head).
  **Recorded by:** Claude.

**Rewrite remap addendum (append-only preserved, d4c2312 pattern):** the F1 remediation
originally landed as `dc773bb` + `bef1a87`; under the owner-ruled protocol remedy (CLAUDE.md
history-rewrite clause: an unpushed behavior-identical change to kernel-watched files that could
neither honestly bump nor pass the guard) the pair was SQUASHED into **`8c3ae4b`**, whose single
honest 0.3.85 bump covers the combined delta. The two prior handoff entries' hashes map:
dc773bb → 8c3ae4b; bef1a87 → (squashed into 8c3ae4b). Record commits remapped f59094a → 8025c24,
ce05afd → d9872fa; later commits preserved (Permits verdict → 88e2054, ratification → 8600533).

**Mechanical re-review verified at the new head:** HEAD tree == pre-rewrite tree (c61472e…,
byte-identical); d9872fa tree == old ce05afd tree; 8c3ae4b's non-planning diff == the combined
old diff EXACTLY (empty diff-of-diffs); kernel guard + adversarial fixtures green at 0.3.85. The
delta pass had already verified this identical tree end-to-end (F1 probes discriminating in BOTH
runtimes on the shared corpus row `minigame_exit_reset_founder_case`; full gate stack green
except the now-resolved guard) — those results transfer by tree identity. The kept client
activation-arm line is shadowed by the unconditional reset; retained deliberately per the
recorded two-arm contract ("first v21 activation initializes zero; every applied v21 Exit resets
zero thereafter") — accepted as the annotation.

**Verdict: APPROVED. Consumed set: {5ebef16, 043ab22, 30999da, a8ab0dc, f0fa2ae, 69faa63,
4a8bdba, 2e2a372, ad06e03, 8c3ae4b} + records {8025c24, d9872fa}.** Range-union: with the prior
endpoints (composition {e935c22}, Pitch, Soul Recovery, Permits {7d9cb37, 90633a6, d30ab9e}),
every implementation commit from e935c22 through 8c3ae4b is verdict-covered EXCEPT {3530b08}
(FCE-C7 copy staging — owed to the FCE ratification flow) and the record-review debt
({b14f82e} + the DF3-noted handoff-discipline items, now including the pre-rewrite records'
claims). **AC1–AC4 remain UNCLAIMABLE** per the carried F2 plan boxes (composed recovery half,
command-flooding proof, privacy enumeration) and MA-C9's surface sequencing — this verdict
covers the implementation to date; it is NOT archival authorization.

## 2026-08-10 — MA composed acceptance-debt implementation handoff

- Extended the composed real-HTTP lifecycle through Soul Recovery without test-only state
  mutation: start, reconnect with token rotation, stale-token rejection, attended heartbeat
  progression, resolve with byte-identical retry, and watchdog cancellation all traverse the
  generated authenticated routes and production coordinator.
- Proved authenticated minigame-command flooding reaches the shared account limiter and leaves the
  active session byte-identical; advancing the injected clock through the ordinary refill policy
  restores service and the same Pitch run completes normally.
- Closed MA4 at both authorities. The generated registry test enumerates all four minigame
  operations as private/access-token-only, pins their sole allowed path parameters, and rejects
  Founder, Company, or server-clock fields. The composed Postgres test proves a second account sees
  no current session and receives byte-identical `unknown_id/minigame_session` responses for a
  foreign versus missing command/resolve target.
- The composed proof exposed a real public-wire divergence: durable recovery receipts used the
  internal replay kinds `resolve_soul_recovery|cancel_soul_recovery`, while the generated API
  contract requires `resolve|cancel`. The coordinator now keeps the internal kind in replay inputs
  and emits the closed public action in the durable receipt. This is an honest semantic correction;
  kernel identity advances from 0.3.87 to 0.3.88.
- The root `test-save-integration` target now accepts `SAVE_TEST_PACKAGES` and `SAVE_TEST_FLAGS`
  overrides while retaining the ordinary Docker-backed full-suite defaults. This keeps focused
  integration work behind the repository's Make interface.
- Normal repository verification is green after the change: `make test-save-integration` and the
  complete `make verify` gate (including 6,623 client tests, 19,881 browser tests, typecheck with
  zero diagnostics, schema/copy/generated checks, and the kernel-history guard at 0.3.88).
- This entry is implementation and self-check evidence only. The batch is not approved or
  archival authority; it is ready for the required designated cross-party review after the full
  gate completes. Surface components remain the sole unchecked MA implementation item.

## 2026-08-10 — designated cross-party verdict: MA AC proofs + recovery wire fix {fa03e02} — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
AC1's composed-socket recovery half, AC3's flooding proof (429 before handler invocation,
post-refill snapshot byte-identical), and AC4's privacy enumeration (byte-equal foreign/missing
rejections; identity fields rejected) are REAL — AC1–AC4 now claimable; AC5/surfaces the sole
remaining item; MA NOT archival-eligible until MA-C9 lands. The recovery wire fix is a genuine
defect (internal ApplyLogged kind leaked into the public receipt's action field vs the generated
contract's closed {"cancel","resolve"}); replay identity untouched; probe-proven discriminating;
kernel 0.3.88 honest. **B-F1 (REQUIRED before MA archival): the MA-C1 owner-routing record — filed
same day in planning/archive/soul-recovery-activities/log.md (see entry there); CLOSED.**
