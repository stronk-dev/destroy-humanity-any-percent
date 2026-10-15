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

## 2026-08-11 — designated cross-party verdict: {02d00d7, b6a3d2c} — BOTH APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

**The defect anatomy, and a correction to the handoff's framing.** The MA-C11 coordinator seam
minted a distinct server-authored Founder intent ID for `start_minigame_session` (emission side,
`founder_replay.go:67` decorating the lazy Fiscal sweep with `wire.Command.IntentID`) — but the
VALIDATION side passed `request.SessionID`, while `minigame_start.go:49` explicitly enforces that
those two differ. The halves of one seam were guaranteed to disagree whenever a sweep event
existed at all. **It was NOT genuinely intermittent:** `FiscalPeriodOpenedWallMS` only advances
when a sweep or harvest COMMITS, and with the shipped `auto_ms = 300` any founder idle 300 ms
emits the event and 500s — every rollback leaves the condition intact and widening. **Deployed,
this would have been a hard, reproducible outage of the entire minigame entry point**, not a
flake; the "intermittent" reading is an artifact of the test fixture resetting the clock
immediately before the call. Realized exposure: zero (unpushed).
**Blast radius: fail-closed, verified.** The error returns under `defer tx.Rollback()` before
every durable write (revisions, events, founder log, retention, commit); no row, receipt, or
replay history is created, and the idempotency key is not consumed — a replay cannot diverge
because the history never exists.
**Probes:** reverting the fix fails 5/5 deterministically with the exact error; the PRE-FIX test
against PRE-FIX code passes 20/20 (the old coverage was genuinely blind); fixture offsets of
-290 ms and -200 ms bracket the 300 ms boundary empirically (10/10 fail vs 0/30 fail). The new
regression test is discriminating, not decorative. Stress at -count=20: stable. No sibling
defect (every other `validateIntentDecision` call site checked); no client parity divergence (TS
already used the command intent ID). Nothing rode along; the new Make selectors leave every gate
default byte-identical (`make -n` expansion unchanged; `verify` untouched). Kernel 0.3.93 is a
VOLUNTARY honest over-signal (the touched paths are unwatched — see F1). All gates green.
**This reviewer's own lane missed it:** the 2026-08-08 MA verdict blessed the distinct-intent-ID
seam after probing the emission half and never traced its other consumer forty lines below.
Recorded as the lesson: when a review approves a seam, it must enumerate BOTH sides' consumers.

**02d00d7 APPROVED** — CI-only, no product code, sound and scoped, with an honest self-correction
of its own earlier "exact reproduction" claim.

**Findings (none blocking):**
- **F1 (MEDIUM — action before MA archival): `server/save/minigame_start.go` and
  `minigame_resolution.go` are still kernel-UNWATCHED** (`server/save/` is enumerated
  file-by-file). The 2026-08-08 F3 flagged this as theoretical; it is not any more — this file
  just shipped a semantic defect, and the next silent change to either coordinator carries no
  guard. Add both paths to `kernel/affecting-paths.json`.
- **F2 (LOW-MED): the `-run Integration` selector hole is systemic**, closed only for this file.
  Four DB-gated tests still fall outside `make test-save-integration` (active-play buff claim,
  care-action founder replay, the Exit atomicity/fault/replay test, minigame PlayWithReceipt) —
  all verified green against Postgres, so it's a selector gap not a latent red, but rename them
  into the gate.
- **F3 (LOW): `.pnpm-store/v11/index.db` and its symlink remain TRACKED** despite the gitignore
  entry (gitignore does not untrack); the index is still being written today. `git rm --cached`.
- **F4 (LOW): the repaired browser job keeps `timeout-minutes: 5`** while gaining a three-engine
  Playwright install — raise to 10 (the composed job's budget, for one engine).

**Range-union: {02d00d7, b6a3d2c} consumed. Every implementation commit from 8cfa00b through
b6a3d2c is now verdict-covered with no gaps**, including the previously carved-out 02d00d7.

## 2026-08-11 — post-verdict guard and gate follow-ups (F1–F4)

- **Implemented by:** Codex. **Recorded by:** Codex. Ready for the ordinary designated
  cross-party review; this entry is not an approval.
- F1: the kernel registry now watches both save-layer minigame coordinators in strict lexical
  order. The registry-only addition changes no runtime semantics and therefore carries no false
  version bump; the history/parity guard remains green at 0.3.93.
- F2: the active-play claim, care-action Founder replay, Exit atomicity/fault/replay, and
  minigame receipt tests now include `Integration` in their test names. The unchanged ordinary
  `make test-save-integration` selector executed all four alongside the existing Postgres suite.
- F3: the pnpm v11 SQLite index and project symlink are removed from Git tracking. The ignored
  local cache files remain present for pnpm; future test runs can no longer dirty the worktree.
- F4: the three-engine browser job now has the same ten-minute budget as the composed browser
  job, covering dependency/browser installation without weakening or skipping its test gate.

## 2026-08-21 — Q-002 platform-alignment backend witness closeout started

- Q-001 received the mandatory designated cross-party approval at `34d04a5`; the serialized
  execution queue therefore unblocks Q-002.
- Scope is test/proof and planning records for MA AC2–AC4 only: exact deterministic error-pair and
  response-byte oracles, a stateful Soul Recovery flood nonmutation/refill witness, and privacy
  enumeration over all eight Minigame plus Recovery operations.
- Predeclared negative controls are: an invalid category/detail cross-product; one appended byte on
  an otherwise exact error response; the seventh Recovery heartbeat with byte-equal authoritative
  state before/after followed by a refill-admitted next transition; and a `founder_id` leak in each
  Recovery request-schema family. A negative control that remains green invalidates the witness.
- Required clean gates are cold root Go tests, sequential Account and Gameserver Postgres
  integration targets, API generated-contract check, and strict client typecheck. Mutations must be
  restored before those gates run.
- Forbidden in this range: new mechanics, changed response semantics, surface contracts or
  components, generated-client changes, player copy, RFC-body reconciliation, status promotion,
  archival, push, or deployment. This entry authorizes no work beyond the accepted Q-002 manifest.

## 2026-08-21 — Q-002 implementation and Codex first-filter

- Implementation commit: `53fd4cb`. No emitted HTTP mapping, mechanic, generated OpenAPI/client
  artifact, surface contract, or player copy changed. API Foundation gained a validation-only
  `ExactJSON` response narrowing; Minigame operations bind their existing status/category/detail
  outputs to exact encoder bytes while the additive generated `APIError` DTO remains unchanged.
- AC2 now compares all 13 deterministic handler cases by exact status and full body bytes, then
  validates those bytes against the owning operation/status literal set. The registry independently
  rejects a schema-valid `idempotency_conflict/minigame_revision` cross-product and one appended
  byte. Exact literals are sorted, schema-valid, deep-cloned, and covered by mutation-safe registry
  tests.
- AC3's real-Postgres Account route fixture now owns explicit Recovery session ID, progress token,
  progress amount, and handler-call count. Six heartbeats transition state; the seventh exact 429
  leaves the serialized authoritative state byte-identical; advancing the injected clock by the
  ruled one-second refill admits exactly one seventh transition.
- AC4 now enumerates all eight Minigame and Soul Recovery operations, their complete path-parameter
  authority, valid requests, and valid success responses. Every request rejects type-valid Founder,
  Company-stream, and server-clock fields; every response rejects a type-valid hidden Founder ID.

Demonstrated failing mutations, all restored before `53fd4cb`:

- changed `conflict/minigame_revision` to the schema-valid but illegal
  `idempotency_conflict/minigame_revision`: the exact handler table failed on the revision row;
- appended `!` after the otherwise correct revision response: the same row failed on full bytes;
- bypassed `recoveryProgress.allow`: the real-Postgres Account witness received HTTP 200 on the
  seventh heartbeat instead of the exact 429;
- admitted a type-valid optional `founder_id` separately in Recovery Start, Progress, and shared
  Finish request schemas: each family failed at its owning operation;
- admitted a type-valid optional `founder_id` separately in Recovery Start, Progress, and Terminal
  response schemas: each family failed at its owning operation.

The first Start-request mutation initially stayed green because the injected value was the string
`attacker`, which the temporary UUID-v7 field rejected on format. That false-green fixture was not
accepted as evidence: production was restored, every private-field value was made type-valid
(including numeric server time), and all six Recovery request/response family mutations above were
rerun to the recorded failures.

Cold restored-tree evidence:

- `make test-go-core CORE_TEST_COUNT=1` — every non-harness Go package green cold.
- `make test-go GO_PACKAGES='./publicapi ./account ./gameserver' GO_TEST_FLAGS=-count=1` and
  `make vet` — focused packages and vet green.
- `make test-save-integration SAVE_TEST_PACKAGES=./account SAVE_TEST_FLAGS='-run Integration'
  SAVE_TEST_COUNT=1` — full Account Postgres population green.
- `make test-save-integration SAVE_TEST_PACKAGES=./gameserver SAVE_TEST_FLAGS='-run Integration'
  SAVE_TEST_COUNT=1` — final full Gameserver Postgres population green; the complete population
  then passed three consecutive cold repetitions (`SAVE_TEST_COUNT=3`, 67.051 s).
- `make api-check` produced no generated diff; `make typecheck` reported zero TypeScript or Svelte
  diagnostics; `git diff --check` was clean.

One earlier restored-tree Gameserver population returned a transient command-2
`500 internal_invariant/minigame_api`. With no code change, the isolated composed lifecycle passed,
the complete population passed, and three further complete repetitions passed. This is disclosed
as non-reproduced flake evidence, not rewritten as an uninterrupted green history and not treated
as proof of a product correction in this range.

Codex first-filter verdict: **APPROVED** for Q-002's bounded backend witness scope.

- **Review by:** Codex.
- **Recorded by:** Codex.
- **Reviewed range:** `eafa2d9..53fd4cb` (predeclared plan boundary plus implementation/tests/docs).

The diff was read in full after the mutations were restored. No temporary edit remains, the
generated API artifacts are byte-unchanged, and the Q-002 assertions bind the actual registry,
mounted Account route, limiter, injected clock, and real Postgres population. This is a first
filter only. Claude must designated-review the exact range after `34d04a5`, including this record
commit, before Q-002 closes or Q-003 starts. It authorizes no surface claim, status promotion,
archival, push, or deployment.

## 2026-08-21 — Claude designated cross-party review of Q-002 `34d04a5..dfaeafe` — CHANGES REQUIRED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Range:** `34d04a5..dfaeafe` = `{eafa2d9, 53fd4cb, dfaeafe}`.

### The witnesses themselves are good — verified by execution

`./account` and `./publicapi` green cold `-count=1`; the Account Postgres integration population
green on an isolated port. My own severing probe: swapping ONE registered error literal
(`access_token` → `token_access`) in a scratch worktree fails the suite immediately — the
byte-equality oracles genuinely discriminate. The corrected type-valid mutation fixtures and the
disclosed transient Gameserver 500 are honest records. None of the findings below is against the
witness content.

### Finding B (BLOCKING): the range exceeds its accepted scope

The accepted `ready-batch-manifest.tsv` row for Q-002 declares `must_not_change:` "…**response
semantics** surface contract…" and a closeout of "one bounded **backend-test** range". Codex's own
predeclaration (`eafa2d9`) states "Scope is test/proof and planning records … only" and forbids
"changed response semantics". The range nonetheless changes production source:
`server/publicapi/registry.go` (+47: a new `ExactJSON` narrowing capability enforced inside
`ValidateResponse`), `server/account/minigame_api_schema.go` (+97: literal population), two
schema-file one-liners, and the canonical `docs/api-foundation.md`.

`ValidateResponse` executes in the runtime hot path (`server/account/api.go:208,238`). The
narrowing is therefore a live acceptance-set change: a future handler regression that emits a
schema-valid but non-literal error body now fails closed at runtime where it previously passed.
That is a response-semantics change — fail-closed and desirable, but categorically the thing the
manifest row forbade, landed inside the same commit as the tests it enables.

**Why this cannot be waved through even though the design is right:** putting the exact bytes in
the registry rather than in test-side tables is the CORRECT home (a test-side byte authority would
be the C34 anti-pattern). But this program's discipline — the one Codex itself has exercised, to
its credit, by retracting its own unruled literals — is that scope boundaries hold even when the
overrun is an improvement. A reviewer who approves a disclosed overrun because it is good work
converts every future predeclaration into a suggestion.

### Required remedy (cheap; the work is not wasted)

1. Split the production tightening (`registry.go`, `minigame_api_schema.go`, the two schema
   one-liners, `docs/api-foundation.md`) into its own predeclared range under the **API
   Foundation** lane where `registry.go` belongs, with its own first filter. The commits are
   unpushed; the standard rewrite carve-out covers a packaging split if Codex prefers it over
   revert-and-reland — either is lawful, state which.
2. Q-002 resubmits as the genuinely test-only remainder plus records. Given my executed evidence
   above, the re-review of both halves will be fast.
3. The manifest row itself needs no amendment — it was right; the batch drifted.

**Verdict: CHANGES REQUIRED. Q-003 remains blocked** (it starts only after the Q-002 designated
verdict records APPROVED). No archival or surface claim.

## 2026-08-21 — Q-002 scope correction after `ba8ca65`

- Claude's changes-required verdict is accepted. The `ExactJSON` registry narrowing is desirable,
  but it is a live production acceptance-set change and therefore did not belong in Q-002's
  backend-test range.
- Remedy chosen: **revert-and-reland**, not history rewriting. The designated verdict now cites
  `{eafa2d9, 53fd4cb, dfaeafe}`; repository history is therefore append-only under the review
  provenance rule. This corrective change removes the API Foundation production/docs portion and
  its foundation tests while retaining Q-002's Account/Minigame witness-only remainder.
- Against the approved Q-001 boundary `34d04a5`, the resulting Q-002 tree changes only
  `server/account/*_test.go` plus its planning records. The operation-labelled handler table still
  requires exact status and complete response bytes, so the wrong-pair and appended-byte production
  mutations remain discriminating without creating a second live registry authority in tests.
- The removed registry narrowing will be predeclared next under `planning/api-foundation/`, then
  re-landed as a separately reviewable production range. Q-002's resubmission will cite that API
  Foundation dependency rather than absorbing it. Q-003 remains blocked throughout.

## 2026-08-21 — Q-002 corrected resubmission and Codex first-filter

- Corrected Q-002 boundary: `34d04a5..b9ebab7`. Its net diff is exactly five files: this plan/log
  plus `server/account/account_integration_test.go`, `minigame_api_schema_test.go`, and
  `minigame_api_test.go`. No production source, schema, canonical behavior doc, generated artifact,
  surface contract, or player copy remains in the corrected Q-002 tree.
- The rejected mixed history was not rewritten. `b9ebab7` reverses its out-of-scope API Foundation
  portion, so the exact base-to-boundary diff is the reviewable test-only remainder; `ba8ca65` and
  the correction record preserve the complete provenance of why the split occurred.
- The required live exact-response authority now exists in the separately predeclared API
  Foundation range `b9ebab7..a854e46` (`a9fdb23`, `0331444`, `a854e46`). Q-002 consumes that
  authority but does not claim or own its production change.
- The original Q-002 negative evidence remains applicable to the unchanged test-only remainder:
  wrong error pair, appended byte, limiter bypass, type-valid Start/Progress/Finish request leaks,
  and type-valid Start/Progress/Terminal response leaks all failed at their named oracle. The API
  range separately demonstrates registry enforcement, literal construction closure, and deep-copy
  ownership.
- Final restored-tree gates after the split are recorded in the API Foundation log: cold focused
  and complete non-harness Go, vet, API drift, typecheck, and sequential Account/Gameserver
  Postgres populations are green.

Codex corrected first-filter verdict: **APPROVED** for Q-002's test-only remainder.

- **Review by:** Codex.
- **Recorded by:** Codex.
- **Reviewed boundary:** net diff `34d04a5..b9ebab7`; this resubmission record is a record-only
  singleton that the designated reviewer must include explicitly.

Claude must issue two explicit verdicts: the corrected Q-002 test-only boundary plus this record,
and the API Foundation range `b9ebab7..a854e46`. Q-003 remains blocked until Q-002 itself records
**APPROVED**. No archival, surface claim, publication, push, or deployment is authorized.

## 2026-08-21 — Claude designated cross-party review of corrected Q-002 (net `34d04a5..b9ebab7` + `{77f35fd}`) — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Boundary verified mechanically:** the net diff `34d04a5..b9ebab7` touches exactly three
  Account test files plus Minigame planning records, and the production files from the rejected
  mixed commit show **zero residual diff** against the pre-Q-002 baseline — the reversal is exact.
  `{77f35fd}` is record-only.
- **Witness value carried over from the `ba8ca65` review, re-confirmed at this head:** witness
  populations green cold; Account Postgres integration green; the byte-equality oracles now pin
  their exact literals test-side against the (separately-approved) registry authority, and the
  earlier severing evidence stands. Finding B is fully discharged: the manifest boundary held, the
  improvement landed anyway — in its right lane.
- **No findings. APPROVED. Q-002 closes; Q-003 is unblocked** and may begin serially per the
  accepted queue. No archival or surface claim is authorized.

## 2026-09-25 — MA AC5 (Pitch half): minigame_session surface implemented (Claude)

**Implemented by:** Claude, on the owner's 2026-09-24 direction that Claude implements. **Review:**
pending Codex designated cross-party review. This is not self-approved.

Scope, following the ruled MA-C9 contract:
- a generated-DTO `MinigameSessionPort`;
- the closed five-state surface plus a host launcher, as GS7/OD-17's alternative, so the ruled
  five states are unchanged;
- an exact rejection table for the server's `minigameErrorJSON` pairs;
- a tenant-surface registry keyed by the pinned `minigame_api` tenant arm, failing closed in both
  directions;
- a presentation-only `PitchTable` child;
- the bundled, hash-verified Pitch content binding;
- Game UI wiring: surface row `minigame_session`, always unlocked. The server's
  `fiscal_unlock_required` answer drives the launcher notice, so no client unlock predicate is
  invented.

For exact props, transitions and the error table, the component enumeration follows the draft
`rfc/garage-player-surfaces.md` GS7 proposal. That section awaits the MA ruling author's adoption,
and any divergence found there is a follow-up. The copy in
`copy/catalog/minigame-surface-candidate.json` is implementer-drafted plain wording. It is owner
adoption material, not ruled copy (AGENTS evidence rule 6).

Evidence (cold):
- **Unit tests:** `vitest run test/minigame-surface.test.ts`, 9/9. Four severings each turned the
  suite red (1 failed | 8 passed):
  - S1 drop the `unknown_offer` rejection row;
  - S2 skip the content-hash check;
  - S3 skip the registry's reverse check;
  - S4 accept an error body with extra keys.
- **Browser tests:** `test/minigame-surface-browser.test.ts`, 12/12 across chromium, firefox and
  webkit. Four severings each turned chromium red (1 failed | 3 passed):
  - B1 retry mints a new `command_id`;
  - B2 no tenant check;
  - B3 fresh create key per attempt;
  - B4 no `play_size` cap.
- **Other gates:** `make typecheck build-client test-client verify-client-boundary
  verify-ci-topology verify-combat-boundary verify-meters-boundary verify-achievements-boundary
  copy-check test-browser` all passed (20094 browser tests), and `make test-game-ui-composed`
  passed.
- **Kernel guard:** `make verify-kernel-version` is RED because of `8add475`'s pending correction
  review (see `planning/minigame-platform-foundation/log.md`). This batch doesn't touch it.

Still open for AC5: the Soul-Recovery surface (MA3's second contract), and a composed real-server
Pitch-through-UI witness.

## 2026-09-25 — MA AC5: composed Pitch-through-UI witness (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review; not
self-approved.

`client/tools/test-game-ui-composed.mjs` now adds a Pitch phase after the existing run-3
continuation. Chromium runs through the real Vite proxy, the composed gameserver and Postgres:
- **Locked path:** opens the `The Pitch` nav tab and presses Start. The server returns a real 409
  `not_eligible/fiscal_unlock_required`, and the typed launcher notice appears.
- **Honest unlock:** `harvest_fiscal_period` then `spend_fiscal_credit {unlock minigame.pitch}`, as
  UUIDv7 player intents over `/api/v1/intents` against the composed epoch's Fiscal catalog
  (100/200/300 ms windows, cost 3). No DB write or fixture clock is used.
- **Play:** Start succeeds (200). Cards are selected by focus+Space on native checkboxes, and
  "Pitch these cards" and "Close the shop" are pressed with Enter, until a terminal receipt appears.
- **Assertions:** the resolved response carries an `applied` `company.cash` receipt; the Game UI
  fetches a snapshot at or beyond `receipt.company_revision`; `current` reads `none`; Back returns
  to the Desk.

Cold result: `make test-game-ui-composed`: "6 UI commands, terminal receipt credited 1e0 at
company revision 11, snapshot refreshed to 11: PASS" followed by the existing lifecycle PASS.

Failing cases, demonstrated then restored:
- **P1:** expecting detail `human_content_locked` fails with "locked Pitch create was not the typed
  fiscal rejection".
- **P2:** aborting every `/api/v1/founder/state` request during play fails with "no post-terminal
  Game UI refresh reached company revision 11; UI snapshots=[10,10] server=11".

Finding: none blocking. The shop phase was exercised only through "Close the shop". Buying a hack
through the composed UI isn't witnessed, because the deterministic seed decides whether the
start-currency covers an offer; the browser suite covers buy_hack.

## 2026-09-25 — MA AC5 (recovery half): soul_recovery surface implemented (Claude)

**Implemented by:** Claude. **Review:** pending Codex designated cross-party review; not
self-approved.

Scope (SR-C3 / SR-C6 / SR-C14):
- `SoulRecoveryPort` over the generated operations;
- the pinned Soul catalog and prestige ceiling bundled with the client, with cadence
  `recovery_beat_ceiling_ms / 3`;
- the picker, with activity copy and disclosures;
- a single `RecoveryScheduler` wired to document visibility;
- hidden and network pauses, with network outranking hidden;
- reconnect-start with token rotation;
- a gone session ends the surface with no reconnect-start;
- finish/stop, and a watchdog message;
- a decorative, seeded, `aria-hidden` toy;
- a `soul_recovery` Game UI tab.

The shared `createOperationCall` was extracted from the minigame port so both ports use the same
exact-error rules. The boundary scanner now covers `src/minigame` and `src/soul` Svelte files.
Copy is added to the candidate catalog, which awaits owner adoption.

**Finding while testing:** a WebKit timing run exposed a real ordering bug, now fixed. A hidden
pause after a network pause replaced the Reconnect prompt, even though the scheduler still required
a reconnect. The fix plus a dedicated test are in this batch. Severing the fix makes that test fail
(chromium 1 failed | 4 passed).

**Evidence (cold):**
- Unit: `soul-recovery-surface.test.ts` passes 5/5. Severing the cadence divisor (/3 → /2) makes it
  fail.
- Browser: 15/15 across three browsers. Severing the gone guard makes the chromium run fail
  (1 failed | 4 passed).
- Boundary: seeding a literal into `RecoveryToy.svelte` is rejected by `verify-client-boundary`.
- Full lanes, all green:
  - `make typecheck build-client test-client verify-client-boundary copy-check test-browser`
    (20124 browser tests);
  - `make test-game-ui-composed`, both the Pitch phase and the v3 lifecycle.
- `verify-kernel-version` remains RED only for `8add475`'s pending correction review.

**AC5 status:** both surface contracts are implemented and pass the UI Foundation gates. The Pitch
child seam is proven in the unit/browser suites and through the composed server (`06edddba`).
MA's remaining boxes are designated review and archival, both owned by Codex.

## 2026-09-25 — F9: Wind Down preview agrees with the MA-C12 rejection (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review; not
self-approved.

**Defect.** Under a `minigame_api` bundle, `transitions.wind_down.eligible` was `tier >= 1`. Exit
nevertheless rejects `not_eligible/minigame_session_active` while a session is `active|claimed`
(`production/prestige.go` freezes `ActiveMinigame` into replay, and `replay.go` rejects on it). So
the Desk offered a Wind Down that silently failed (garage-player-surfaces F9).

**Fix.** No new mechanic.
- `gameui.WithMinigameActivity` binds the same read-only `minigame.Repository.ActiveMinigame`.
- The projector consults it only when the run bundle pins `minigame_api`, the same condition
  production uses, and fails loud without it.
- `previewPhaseATransitions` sets `WindDown = tier >= 1 && !minigameActive`.
- The bootstrap initial projection passes `false`: a Founder created in that transaction cannot
  own a session.
- Composition wires the repository.
- `server/gameui` isn't in `kernel/affecting-paths.json`, so there's no version bump.

**Other Exit previews and AR-F3.**
- The snapshot projects no `accept_exit_offer` eligibility. The Offer Sheet button is gated
  client-side only, so its rejection still surfaces through the receipt path, a separate
  known-open item (garage-player-surfaces F-receipt).
- The previewed `cross_gate` is refuted as an Exit case. The only previewed gate is the uncrossed
  `gate.t0_to_t1`. `scriptedExitDue` under the live curriculum requires
  `GatesCrossed[first_failure.gate_id = gate.t0_to_t1]`, so the previewed crossing is always the
  ordinary transition. Its preview already runs the production transition.
- AR-F3's broader claim is CONFIRMED by code read but not fixed here, since it is design
  territory routed to Game UI/Curriculum. Once the scripted first failure is due, *every* Company
  intent routes to `handleScriptedCrossGateExit`, which rejects with `minigame_session_active`
  while a session is active. The Company is therefore frozen until the session ends.
- Separately, an active Soul-recovery session rejects all non-fiscal intents with
  `exclusive_activity`. The wind_down preview doesn't reflect that either, which is the same class
  of gap and a follow-up for the recovery-exclusivity owner.

**Evidence (cold).**
- **Go:** `make test-go GO_PACKAGES='./gameui ./gameserver ./production' GO_TEST_FLAGS='-count=1'`
  passes. `TestPhaseATransitionPreviewWithholdsWindDownWhileMinigameSessionActive` and
  `TestProjectorMinigameActivityMirrorsExitGate` are new.
- **Postgres integration:** passes for `./gameserver ./account ./gameui ./deploymentrelease`.
- **Composed lane:** `make test-game-ui-composed` passes (Pitch and v3).
- **Severing:**
  - S1: dropping `&& !minigameActive` fails the preview test.
  - S2: making a missing resolver silent fails the gate test.
  - S3: omitting `WithMinigameActivity` in composition fails the composed lane (exit 2: "composed
    live Game UI v3 transition snapshot round trip failed").

## 2026-09-25 — Kernel-guard finding on the surface commits (Claude, self-disclosed)

`kernel/affecting-paths.json` guards `client/src/minigame/` and `client/src/soul/`. `50a3a51` and
`b92a05de` added presentation-only files there without bumping `kernel/VERSION`. They change no
transition, receipt, event, snapshot or state encoding, but `make verify-kernel-version` fails on
them. It went unnoticed because the guard was already red on `8add475`'s pending correction.

- **Forward fix, this commit:** every surface file moves to the unguarded `client/src/game-ui/minigame/`
  and `client/src/game-ui/soul/`, with imports and docs updated. No further commit touches a guarded
  path for UI work.
- **Verification, cold:** typecheck, `test-client`, `verify-client-boundary` (6 component files),
  `build-client`, `test-browser` (20124) and `test-game-ui-composed` (Pitch phase and v3 lifecycle)
  all pass.
- **Still open, owner decision:** the historical commits still trip the guard. The CLAUDE.md
  rewrite class covers exactly this case (an unpushed commit forcing a false version signal through
  a behavior-identical change to a kernel-watched file), and no verdict cites the hashes. Claude
  attempted the replay; the session's permission classifier denied it as a destructive git
  operation, and Claude did not pursue it.
  - **Option (a):** Marco authorizes the rewrite of `0cf9f7a..HEAD`.
  - **Option (b):** append-only history corrections for `50a3a51` and `b92a05de`. This requires a
    `kernel/VERSION` bump with no semantic change behind it, plus Codex's independent review
    sections, as for `8add475`.

## 2026-10-04 — Codex F9 review predeclaration: composed active-session preview

- **Question:** does the actual authenticated Game UI snapshot withhold Wind Down only
  while a real Pitch session is active, then restore eligibility after its terminal
  receipt? The existing composed Pitch flow starts from Tier 0 after two Exits, where
  Wind Down is ineligible independently of session state.
- **Population:** the declared composed browser lane against real Postgres/gameserver,
  with a Tier-1 Company reached through the visible cross-gate control. Read the exact
  `/api/v1/founder/state` snapshot before Pitch creation, during its live session,
  and after the terminal receipt. The Pitch itself remains UI-driven.
- **Criterion and negative control:** eligibility is true → false → true across
  those three states. Temporarily dropping the `!minigameActive` predicate must fail
  the active-state assertion. A setup that cannot reach Tier 1 or start Pitch is a
  failed witness, never a green skip. This is a test-only correction under MA-C12/F9;
  it cannot prove every Exit/recovery exclusivity path or the wider MA AC1–AC5 set.

## 2026-10-04 — Codex designated F9 review; composed witness added

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `2def3612^..2def3612`
  only. **Decision:** **APPROVED** for F9's minigame-session Wind Down preview and
  production composition binding, with the supplemental composed witness below;
  not approval of MA AC1–AC5, Soul exclusivity, offer-button eligibility, or the
  full v0.1 handoff range.
- Source comparison found `Projector.minigameActive` and production Exit use the same
  read-only repository predicate and the same `MinigameAPI != nil` activation. Cold
  `make test-go GO_PACKAGES='./gameui ./gameserver ./production' GO_TEST_FLAGS='-count=1'`
  passed. The existing composed browser test played Pitch from Tier 0, where Wind
  Down was already ineligible; that run did not witness the F9 state transition
  (RP-144). Codex added a Tier-1 cross-gate setup and before/during/after real
  `/api/v1/founder/state` checks in `5ad457ce`. With real Postgres and the composed
  gameserver, they observed eligible → ineligible → eligible while UI-driven Pitch
  reached a terminal receipt. `node --check` passed.
- Temporarily removing `&& !minigameActive` made the composed run fail at
  `active Pitch session offered Wind Down` with `eligible:true`. The first severed
  attempt timed out before Pitch creation, so it was not counted as discrimination;
  the repeated attempt reached and failed the exact assertion. Omitting
  `WithMinigameActivity` in composition failed the composed v4 snapshot path.
  Both mutations were restored, and the full composed target reran green.
- The first unmutated composed attempt could not start cached amd64 Postgres on this
  arm64 host. Pulling the same official `postgres:16-alpine` tag for arm64 allowed
  the declared target to run; no product/Compose file was changed. The test-only
  Codex range `f773cf07^..5ad457ce` requires Claude's designated cross-party
  review. This verdict does not turn that self-authored test into its own approval.

## 2026-10-04 — Codex MA3 visible-only review probe, predeclared

- **Question:** if the recovery start response arrives after the document is already
  hidden, does the default document-visibility adapter suppress every heartbeat
  until the tab becomes visible? MA3/SR-C6 require visible-only beats.
- **Population:** mount the real Soul recovery surface in the browser with its
  default visibility adapter, override `document.visibilityState` to `hidden`
  before start resolves, use a fake coordinator port and a short test cadence.
  Then restore `visible` and dispatch `visibilitychange`.
- **Criterion:** zero progress calls and a visible hidden-pause state throughout
  the hidden interval; a progress call may begin only after the visible event.
  A test that starts visible and merely sends a later hidden event does not
  exercise this case. The negative control is the current adapter with no
  initial-state callback: the hidden-interval assertion must fail.
- **Limit:** this is a browser component/adapter boundary, not a real-server
  Soul terminal or all MA acceptance criteria.

## 2026-10-04 — Codex targeted MA3 Soul-surface review and RP-145 correction

- **Review by:** Codex. **Recorded by:** Codex. **Inspected Claude changes:** the Soul
  surface in `b92a05de` and its path-only relocation in `79ff1aa1`; this is a
  targeted review, not a range-union verdict over Minigame API/Surface.
  **Decision:** **CHANGES REQUIRED** on MA3 visible-only heartbeat behavior.
- The default document adapter listened only for future visibility events.
  The predeclared real-browser component test starts hidden before the start
  response resolves. On the uncorrected code it failed in both Chromium and
  WebKit because the UI never entered the hidden pause; the scheduler's
  initial `visible=true` also permits a first beat. This is RP-145, not an
  inferred defect from static inspection.
- **Codex correction:** emit the document's current visibility synchronously
  when the scheduler subscribes. No scheduler contract, transport, copy or
  server mechanics changed. The same test now shows hidden pause with zero
  progress calls, then a progress call only after a visible event. The whole
  Soul surface browser file passes 6/6 in each of Chromium and WebKit.
  A post-fix severing run removed only that initial callback: the tightened
  assertion failed with two progress calls bearing the old token while the
  document was hidden. Restoring the callback passed the exact case in both
  browsers. This proves the heartbeat defect, not merely the missing notice.
- **Cold gates after correction:** `make typecheck` (0 diagnostics),
  `make test-client` (6,905 passed, 77 skipped), `make copy-check`, and full
  `make test-browser BROWSER_TEST_FLAGS='--browser=chromium'` and
  `...='--browser=webkit'` (6,981 passed, 1 skipped per browser, plus the
  performance lane) all pass. `make verify-client-boundary` also passes on the
  corrected component.
  The composed Postgres/Pitch/v4 lane passed before the adapter edit and does
  not exercise Soul recovery.
- **Unverified gate:** the default three-browser target and a Firefox-only
  rerun both failed before test import: Vitest timed out connecting to the
  Firefox browser session. A direct local Playwright `firefox.launch` reported
  `SIGABRT` from the installed headless Firefox process. This is a local
  browser-launch failure, not a green Firefox result or evidence of a Soul
  assertion failure. The full browser gate stays open.
- **Cross-party gate:** this Codex-authored correction and its record require
  Claude's exact-range designated review. No archive/status promotion follows
  from Codex's own test pass; the larger MA range also remains unreviewed.

## 2026-10-04 — Codex MA3 unmount-during-start probe, predeclared

- **Question:** after the player leaves the Soul surface while a start or
  reconnect response is pending, can the late response start a heartbeat
  scheduler on an unmounted component? MA3's surface lifecycle owns exactly
  one visible-only scheduler while mounted; leaving must stop it.
- **Population:** real browser component, deferred fake `port.start` response,
  short beat interval. Click Begin, unmount before the promise resolves, then
  resolve the response and wait beyond two beat intervals. Repeat for a
  reconnect response after an already active session is paused.
- **Criterion and negative control:** no `port.progress` call after unmount in
  either arm. Current code's only `onDestroy` stop runs before the deferred
  response; the test must fail on that code if a scheduler is created later.
  A test that unmounts only after start settles is not discriminating.
- **Limit:** this checks client lifecycle and heartbeat leakage, not server
  cancellation or eventual session recovery on remount.

## 2026-10-04 — Codex targeted MA3 lifecycle finding and RP-146 correction

- **Review by:** Codex. **Recorded by:** Codex. **Inspected Claude changes:**
  Soul surface in `b92a05de` and path-only move `79ff1aa1`. **Decision:**
  **CHANGES REQUIRED** on the late-response lifecycle boundary; not a
  range-union approval of the MA RFC.
- Both predeclared browser negatives fired on the original code. A deferred
  start response settled after unmount and sent three progress beats. A
  deferred reconnect response settled after unmount and sent one additional
  beat. `onDestroy` had stopped the former scheduler, but both async paths
  rebound a new one afterward. This is RP-146.
- **Codex correction:** a mount-lifetime flag is set before stopping the
  scheduler. Begin and reconnect discard late responses before changing view
  or binding a new scheduler. The two new cases plus the existing six Soul
  browser cases pass in Chromium and WebKit. No server session, cancellation,
  token or reward mechanics changed.
- **Cold gates after correction:** `make typecheck` (0 diagnostics),
  `make test-client` (6,905 passed, 79 browser-only skipped),
  `make verify-client-boundary`, `make copy-check`, `make build-client`, full Chromium and WebKit `make test-browser`
  selectors (6,983 passed, 1 skipped each; performance lane passed) all pass.
  The Firefox-inclusive browser gate remains unverified because of the known
  local Firefox launch failure from RP-145, not a product assertion.
- **Cross-party gate:** Claude must designate-review this Codex correction
  over its exact commit range. Neither this self-authored correction nor its
  tests authorize MA archival or close the wider MA AC1–AC5 review.

## 2026-10-04 — Codex MA AC5 Fiscal-to-Pitch composed witness, predeclared

- **Question:** now that the Fiscal surface exists, can a player unlock The
  Pitch and start it using only rendered controls, rather than the composed
  test's direct API setup? The existing witness proves Pitch keyboard play
  and a real server receipt, but injects both prerequisite Fiscal intents.
- **Population:** the existing composed browser/Vite/gameserver/Postgres lane
  after the next Company begins. Attempt locked Pitch as before, navigate to
  Earnings Calls, click the harvest control after the pinned clock opens,
  then click the Pitch-specific unlock in its list item and return to Pitch.
- **Criterion:** both DOM actions emit applied `/api/v1/intents` requests with
  exact kinds/target; the subsequent UI Start creates a 200 session and the
  existing keyboard-driven Pitch run reaches an applied terminal receipt and
  refreshed Company snapshot. No direct API mutation is allowed in setup.
- **Negative controls:** sever the Fiscal harvest click and separately sever
  the Pitch unlock click. Each must fail at a named action or at locked
  creation; a direct API helper may not rescue the run. This test establishes
  the composed Fiscal→Pitch journey, not all MA AC1–AC5 or 1.0.

## 2026-10-04 — Codex MA AC5 composed handoff result (RP-147)

The direct Fiscal API setup in the Pitch composed path was replaced with
Earnings Calls DOM harvest and Pitch-specific unlock controls. That exposed
the accepted Garage Surfaces GS0.2 applied-refresh violation documented in
`planning/garage-player-surfaces/log.md`. With the host correction, the
Postgres/Vite/browser journey applies both DOM intents, starts Pitch, sends
five keyboard-driven commands, reaches a terminal applied receipt, and
observes a refreshed Company snapshot. Disconnecting either prerequisite
control fails the composed path, and no direct API mutation remains in setup.
This is a Codex-authored test correction awaiting Claude's exact-range
cross-party review; MA AC1–AC5 and the original Claude implementation range
are not thereby approved.

## 2026-10-04 — Codex MA3 Soul keyboard-path probe, predeclared

- **Question:** can a keyboard-only player start a Soul recovery activity and
  stop it early through native controls, with the active and terminal states
  understandable and focus placed on the terminal heading? MA3 inherits the
  UI Foundation keyboard/axe gate; the current Soul tests click controls.
- **Population:** real Chromium/WebKit browser component with a fake
  coordinator port and pinned activity content. Focus the first Begin button,
  activate it with Enter, then focus Stop early and activate it with Enter.
  No `.click()` or synthetic keydown may perform the action.
- **Criterion:** exactly one start and one cancel reach the port, terminal
  cancellation renders, focus moves to the heading, and axe reports no
  serious/critical violations in active and terminal states.
- **Negative control:** disconnect the Begin callback while retaining the
  visible button. The test must fail because no start reaches the port;
  restoring it passes. This is a component keyboard witness, not an assistive
  technology study or real-server recovery proof.

## 2026-10-04 — Codex MA3 Soul keyboard witness result (RP-148)

- The new browser case uses `userEvent.keyboard("{Enter}")` after focusing
  native Begin and Stop early buttons; no `.click()` or synthetic keydown
  performs either action. It observes one `start`, one `cancel`, a rendered
  terminal receipt, focus on `#recovery-heading`, and no serious/critical axe
  violations in active or terminal state. The complete Soul browser file
  passes 9/9 in Chromium and WebKit, with the performance lane passing.
- Temporarily replacing the Begin callback with a no-op made the focused case
  fail at `port.starts` (`[]` instead of `["defrag"]`). The callback was
  restored and both browsers reran green. Cold real-Postgres
  `TestComposedMinigameAPILifecycleUsesPinnedTenantResolverIntegration` and
  `make api-check` also pass at this HEAD; they prove the composed API and
  generated files, not this keyboard path.
- The first static `vitest/browser` import made the ordinary Node client suite
  fail before `skipIf(!browser)` could apply. Moving the import inside the
  browser-only case restored `make test-client` (6,905 passed, 81 skipped),
  and the corrected Soul file reran 9/9 in Chromium and WebKit. The full
  Chromium/WebKit browser populations then passed (6,985 tests each, one
  skipped, plus the performance lane). The failed intermediate run is not
  counted as a green gate.
- **Review by:** Codex. **Recorded by:** Codex. This is a Codex-authored
  MA3/AC5 test-only correction under the accepted RFC, pending Claude's
  exact-range cross-party designated review. It does not establish Firefox,
  screen-reader task success, full MA range-union approval or release readiness.

## 2026-10-08 — Pitch detached-command lifecycle correction (RP-415)

- **Outcome/authority:** MA3/MA-C9 gives the surface lifecycle ownership. Leaving
  during an in-flight command now prevents late rejection handling from reading
  the session again. Deferred content completion and terminal focus also respect
  unmount. No command cancellation, game mechanics, receipt or retry policy change.
- **Regression:** a mounted active table submits one command, leaves through its
  real Leave button, then receives a held revision-conflict/unknown-session reply.
  Both cases failed on the old component in Chromium and WebKit: `current()` ran
  twice instead of once. A held terminal success already passed and remains a
  control: no former-host terminal callback. Initial test execution used the wrong
  Leave label and failed before the behavior; corrected before the actual red run.
- **Coverage:** final complete affected browser file passes14 executions in
  Chromium/WebKit; `minigame-surface.test.ts` passes9; `make typecheck build-client`
  completes with zero errors/warnings and a successful production build. Initial
  sandbox listen denial ran no tests; permitted local-server runs executed normally.
  This controlled port test proves component lifetime, not HTTP/SQL faults or a new
  full-game journey. No Firefox, performance, hosted or full-CI rerun/claim.
- **Handoff:** exact new review range starts after `4b8861d9` through this commit.
  Review by: Codex (implementer first filter). Recorded by: Codex. Designated
  cross-party review remains pending alongside earlier ranges; no acceptance or
  archival promotion. Next: the existing wider API/Surface closeout, not another
  lifecycle framework. The full-product status remains in `planning/CURRENT-STATE.md`.

## 2026-10-08 — terminal failure releases its own claim (RP-416)

- **Outcome/authority:** MA-C13 / Platform C38's token-owned, atomic terminal
  coordinator now releases its claim on a returned failure, just as nonterminal
  play releases rejected work. `ReleaseResolutionClaim` accepts only the opaque
  certification and cannot release a replaced token or reopen a committed result.
  Request cancellation does not cancel cleanup; other cleanup errors join the
  original error. There is no automatic retry or change to the crash lease,
  public schema, successful payout formula or replay arithmetic. Kernel175→176
  lands with the real failure/recovery change in all three mirrors.
- **Regression:** replaced the prior API-fault cleanup-by-deletion with immediate
  same-ID/body retry. Old-source real Postgres run82333 fails because the first
  rolled-back session remains claimed at revision1; subsequent cases cannot start
  because that session blocks exclusive activity, not three independent defect
  reproductions. Final company-write/final-retention/cancellation cases all reach
  their named fault, preserve both full save states/revisions, faucet and empty
  command/receipt history, then immediately credit exactly50 once. Repeated retry
  returns identical bytes and preserves both heads, faucet and single command/
  receipt plus two events. No lease ageing/deletion/restart rescues these cases.
- **Safety controls:** a separate real-DB test deliberately replaces an aged
  worker claim; old-token cleanup refuses without affecting the new token. Current
  cleanup restores the committed snapshot; repeated cleanup and cleanup after a
  real terminal commit refuse unchanged. A real coordinator post-commit projector
  failure preserves the resolved row and exact SQL-backed command response.
- **Executed coverage:** final container80544 passes the new claim test, complete
  resolution/rollback/zero-credit population and composed real-HTTP Pitch/Soul
  lifecycle; container41275 executes all10 minigame DB integration functions with
  no skips. Declared `golang:1.26` image is Linux/amd64 (inspected), with real
  Postgres16. Native minigame/kernel tests and affected vet pass; broad native
  production retains only the original27 RP-307 Clout partition failures. TS
  replay/resolution/surface tests103, types (zero diagnostics), production build
  and byte-unchanged `api-check` pass. No Firefox, browser-fault journey, hosted or
  whole-CI green claim; the historical kernel guard remains independently open.
- **Review/handoff:** Review by: Codex (implementer first filter). Recorded by:
  Codex. Exact new range is `b0eab61f` exclusive through this commit, including
  the regression, runtime, version and these records. Claude designated review
  remains pending; no feature acceptance or archival. Typer/Arcade public-wire
  choice has been requested, not adopted. Clout's undefined projection target,
  numeric/episode decisions and the full nine-tier release scope remain open.

## 2026-10-08 — Bind the public Pitch lifecycle to its stored outcome

Outcome: under MA1/AC1/MA-C13 and Platform MP1/MP4, extend the existing composed HTTP
test, not a new driver. After the first real command, retry returns identical bytes and
current-session reconnect returns the entire applied revision2 snapshot. HTTP still owns
all mutations; independent repository/SQL reads observe frozen start coordinates, the
certified terminal/hash, exact cash delta, unchanged unrated Pitch rating, selected quality,
both save revisions and terminal coordinates, command/receipt counts, required resolution
logs/events, one faucet spend and verified Founder history. Recheck after terminal command,
resolve and create retries. No production, schema, catalog, kernel, balance or CI byte changed.

Regression/control: temporarily incrementing the actual test database's quota after the
successful HTTP retries makes54476 fail `HTTP resolve/retry faucet quota=2`; response-byte
assertions had passed. Removed the injected update before final tests. Initial native compile
caught my use of Ledger.Balance's bool as an error; fixed before DB execution. First DB52712
failed an assumed total event count2 (actual3); the test now checks exactly one of each required
minigame event on its correct stream rather than incorrectly forbidding ancillary events.
Founder replay continues to check its complete event history. Focused79986 then passes.

Coverage: final `make test-save-integration SAVE_TEST_PACKAGES='./gameserver' SAVE_TEST_FLAGS='-v'`
69693 passes the complete package in53.137s, including the changed HTTP test1.62s, all DB/socket
populations and existing wrong-receipt/authorization controls, no skips. The inspected cached
golang:1.26 image is Linux/amd64; Postgres16 is real and declared by Compose. Final native
`make test-go vet GO_PACKAGES='./gameserver' GO_TEST_FLAGS='-count=1'`18897 passes; native DB
tests skip without TEST_DATABASE_URL, not substituted for the container evidence. Earlier full
package38144 passes before the final mid-run reconnect supplement. Gofmt/diff checks pass.

Boundary: accepted public v1 is solo with closed tenant wire schemas. Composition registers
real tenants, not the platform's arbitrary fixture. This supplement does not invent an async/
fixture HTTP endpoint, prove live scaling-source mutation, full Platform AC1, browser/AT tasks,
OS crash recovery or release readiness. No browser/full-CI rerun needed for test-only Go work;
the last inspected hosted kernel-history/Clout failures remain open independently.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new review range:
`8eb2eb70` exclusive through this test/docs/plan/log commit; designated cross-party review
pending. No acceptance flips, archival, push or cleanup. Next: remaining accepted integration
and consolidated review; Typer/Arcade wire choices and other unaccepted successors remain held.

## 2026-10-08 — Native Recovery through the production client and real service

Outcome: RP-446, under MA2/MA3/AC5. Extend the existing composed journey with native keyboard
Begin → real scheduler/progress → leave/stopped beats → same-session reconnect with rotated
capability → Stop early. Independent SQL binds terminal receipt, both save revisions, one
start/terminal event and Founder log, unchanged Soul and Company balances. Before Recovery,
native manual actions at1Hz fund two ordinary generator purchases; positive persisted idle
accrual net of purchase cost makes the production-suppression check nonvacuous. Default CI
gets this bounded lifecycle; an explicit root target waits unchanged minted Repot300000ms
and requires native Finish, exact stored/event/rendered Soul40→45, stopped beats and reload.
Only the full population appends disclosed Founder Soul40/revision setup; no clock/progress
grant, catalog/rate/limiter, production/schema/kernel or CI-workflow/timeout change.

Regression and instrument corrections: earlier blocked-Begin control31123 prevents the native
event reaching the mounted handler, emits zero start requests and fails after30s; no production
mutation. Removed/restored before final runs. It predates the producing-Company supplement,
not falsely represented as a severing of every final assertion. Initial91687/8833 fail my SQL
observer's wrong Founder-log column and psql command-tag parsing. Earlier full31975 passes
the actual five-minute flow, but final review catches its zero-producer population weakness.
Producer attempts12710/66611 find the actual post-gate cash1e0 cannot afford the generator;
use actual manual earnings, not another cash grant.74804 then hits real HTTP429 because my
duplicate observer reads amplify setup traffic; remove the duplicate and use native1Hz input
below the unchanged300/min refill, with no request retry.94296 independently stops in the
existing Pitch prerequisite when an Exit offer preempts Fiscal harvest. The driver now uses
the same bounded visible Decline handling already present at Pitch start/unlock, then fresh
harvest consent; it still requires the applied harvest receipt. These were driver/population
failures, not silently classified as game fixes or discarded green-search attempts.
Both final main journeys report zero offer declines: the added harvest-preemption branch was
not separately seeded in this batch and is not claimed as uniquely reproduced repair evidence.

Coverage: final `make test-soul-recovery-composed`59876 exits0. Fourteen native manual actions,
two purchases and positive stored accrual pass; both cancel and full resolution preserve the
producing Company's balances. The main browser journey takes350231ms and observes180 actual
Recovery progress replies, all200; exact built HTML/JS/CSS/Worker and started Worker pass.
Final whole `make test-game-ui-composed`89786 is RED, not rerun to seek green: the actual bounded
Recovery, main/DB populations, pets/Reputation, Clout, payment negatives and Garden pass, then
Tier2's native second upgrade carries11 against observed12. Read-only retained SQL confirms
the preceding first upgrade committed12 from11. RP-447 routes this stale-revision handoff to
Garage GS0.2 and requires a deterministic read-order regression; no unrelated production
repair or weakening of the exact assertion is mixed into this test-only range. Syntax/diff checks and
`make verify-ci-topology` pass, including all13 rejecting controls. Earlier whole86299/29352
passes precede the producer supplement and do not substitute for the final source run.

Limits/handoff: this is a played real five-minute activity after explicit depletion, not natural
Soul debit/history replay, 15/45-minute activities, real-browser hidden/network/watchdog tasks,
renewal, full MA/accessibility acceptance or a clean-host/release proof. The last inspected hosted
kernel-history/Clout failures remain independent. Review by: Codex (implementer first filter).
Recorded by: Codex. Exact new review range is `cbd2069f` exclusive through this containing
test/Make/docs/record commit; designated cross-party review pending. No acceptance flips,
archive, push or cleanup. Existing public-wire and other unaccepted successor contracts stay held.

## 2026-10-08 — Ignore Recovery terminal replies after the surface leaves (RP-452)

Outcome: conform MA3/AC5's owned surface lifecycle to unmount, matching the existing Begin/
Reconnect guards and detached Minigame surface behavior. Finish and Stop early now check
destruction after the awaited terminal response, before changing the view or notifying the
former host. The server command is not cancelled/retried; timing, payout, schemas, copy,
clock, kernel and CI remain unchanged. Canonical platform docs describe this boundary.

Regression: native Begin and Finish/Stop with a held port reply, then either remain mounted
or unmount and focus another control. Old-source61951 fails all four detached executions
in Chromium/WebKit while four mounted controls pass. Earlier58746/86342 also exposed two
mistakes in my expected terminal text (case and missing existing prefix); corrected against
the unchanged catalog before retaining61951 as the clean baseline. No production copy edit.
The new tests require one command, no detached terminal callback/focus move/heartbeat, and
the normal mounted receipt/text/heading focus. They use a mounted component and port double,
not a delayed real HTTP/SQL response or proof of a server transaction defect.

Coverage: final affected browser file67846 passes26/26 (Chromium/WebKit);75 related scheduler,
surface and generated-port units pass. Root typecheck/build/boundaries92249 passes with zero
Svelte/type diagnostics. Full unchanged `make test-game-ui-composed`24540 exits0: actual
Postgres/HTTP/socket populations, production-built main/Pitch/bounded Recovery, pets/Reputation,
Clout, no-payment negatives, Garden and Tier2. Its Recovery cancel is an ordinary real-service
path, not the held-response regression or a new five-minute resolve run. Tier2 emits two Vite
proxy ECONNRESET diagnostics but completes its existing exact SQL/reload assertions; no claim
of zero transport errors. Diff check passes; all sessions terminal.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new source/test/
docs/ledger/log range begins after `0e465815` through this commit; designated cross-party
review pending. No feature acceptance, archival, push, Firefox/AT/whole-CI or release claim.
Continue accepted integration; the separately requested Typer/Arcade public-wire choice is
not adopted by asking it, and all other owner/author-held contracts remain held.

## 2026-10-08 — Resume public Pitch through a reconstructed application

Outcome: strengthen the existing MA AC1/MA-C2–C4 HTTP lifecycle, not introduce another
driver or public tenant. After command revision2, drain and stop the original server;
Compose fresh catalogs, account middleware, repository, platform and handlers over the
same Postgres. Require fresh authority objects, original create/command bytes and the
complete current snapshot, then finish through the existing exact SQL payout/history/
replay and duplicate-receipt controls. Recovery/refusal/authentication checks also run
through the replacement server. Test/docs/records only; no production behavior changes.

Regression: Linux/Postgres57029/73685/39033 reach the new restart/resume/payout checks,
then fail the existing bootstrap authentication control. A sanitized temporary diagnostic
finds the returned JWT expired and its access-token row absent. Migration73 deliberately
has no account FK; this test's cleanup omitted bootstrap receipts, leaving an earlier
run's encrypted response at its fixed idempotency key. RP-459 adds that test-owned table
to setup/cleanup, without changing authentication, receipt semantics or the assertion.
Exploratory token decoding is removed; the retained failure message reports only error/
presence/matching booleans, never credentials. An initial native vet context-leak finding
is fixed with separately owned original/resumed cancellation functions.

Coverage: focused declared Linux-amd64/Postgres10981 passes (2.341s). That invocation's
flags included count2 but the root target appended count1; it is one execution, not two.
The actual repeated run24410 uses `CI_TEST_COUNT=2`: both full lifecycle executions pass
on the same database (5.752s total), including30 refusal controls each. Native affected
`make test-go GO_PACKAGES='./gameserver' GO_TEST_FLAGS='-count=1'`78941 and
`make vet GO_PACKAGES='./gameserver'` pass; native DB cases skip, not integration evidence.
Full declared `make test-go-ci CI_TEST_PACKAGES='./gameserver' CI_TEST_FLAGS='-v'`
45737 passes (156.690s), with the DB populations actually executed and no skips; the
strengthened lifecycle passes after the first-hour/account/Garden populations as well.
Gofmt and `git diff --check` pass. All command sessions are terminal.

Limits/handoff: graceful in-process reconstruction under identical configuration, not
OS-process death, a clean release host, live scaling-source mutation, generic async HTTP,
Typer/Arcade wire adoption, browser/AT tasks or whole Platform/MA acceptance. No browser
rerun is required for this Go test-only change; no whole/hosted CI green claim follows.
Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new review
range begins after `e0de6299` through this containing commit; designated cross-party
review pending. No acceptance flips, archival, push or unrelated cleanup.

## 2026-10-08 — Browser survives process death after a committed Pitch command

Under MA AC1/MA-C5/C13, extend the existing main journey rather than introduce a
driver or retry policy. The first native command reaches the real handler and
commits a nonterminal result; Playwright withholds its reply, confirms SIGKILL
exit of the owned gameserver child, then reports the lost connection. Require
the actual paused notice/Retry control before restarting the same binary/config
over unchanged Postgres. The browser automatically sends the identical command
ID/body and gets byte-identical committed response bytes. Independent SQL checks
require the complete saved snapshot, revision2, one command and one receipt.
New socket/authoritative Founder read preserves the run; neither account bootstrap
nor document navigation may substitute for recovery. Native play then reaches
the applied payout and exact terminal command/receipt/history counts. The existing
response deadline also bounds a missing native dispatch; fetch failure diagnostics
do not print Playwright's potentially credential-bearing upstream request log.

Executed: whole `make test-game-ui-composed`72014 exits0, including real persisted
populations, the new crash case, bounded Soul Recovery, Cosmetic/Reputation/Clout,
N5 failure controls, Garden and Tier2. Tier2 still emits two Vite WebSocket reset
diagnostics before passing its actual SQL/reload assertions. Subsequently tightened
the saved-snapshot assertion and bounded fault wait; final affected
`node client/tools/test-game-ui-composed.mjs`54224 exits0 on those exact final bytes.
It executes all required DB populations, exact built-client/Worker proof, crash
recovery, six native Pitch commands to payout, and bounded Soul Recovery. Its
request trace retains one deliberately failed command delivery followed by the
successful same-ID retry. Unaffected sibling journeys were not rerun after those
last main-journey guards. Node syntax and diff checks pass; both sessions terminal.

Test/tool/docs only: no production, schema, kernel, balance, CI workflow or timeout
policy changes. This covers process death after commit, not mid-transaction death,
Postgres/host failure, release-image/clean-host recovery, browser credential renewal,
generic async HTTP or whole-feature acceptance. Review by: Codex (implementer first
filter). Recorded by: Codex. Exact new range starts after `a0052ddf` through this
containing commit and requires designated cross-party review. No acceptance flips,
archival or push. Next: remaining accepted API/player integrations and consolidated
review; held public-wire, clock, content and owner decisions remain separate.

## 2026-10-08 — Native Recovery reconnect after a committed heartbeat reply is lost

Under MA2/MA3/AC1 and SR-C6, extend the existing composed Recovery journey, not the protocol,
clock or retry policy. A native scheduled heartbeat reaches the real handler and credits an
interval in Postgres; Playwright drops only its reply. Require the visible network pause,
unchanged displayed progress, no beats/automatic start across the real catalog ceiling, and
exact unchanged stored session. Native Reconnect preserves session/progress and rotates the
capability. Its first heartbeat must credit none of the missed interval; the next fresh interval
advances rendered progress with exactly two resumed dispatches. Existing terminal SQL/event/
Soul and positive-production suppression controls remain. Capabilities/upstream request details
are withheld from fault diagnostics. No new driver, fixture epoch, CI lane or production change.

Executed `node client/tools/test-game-ui-composed.mjs`73472: exit0. All eight persisted refresh
cases and all required selected Account/GameUI/Production/Leaderboard/Gameserver DB populations
execute and pass. The production-built browser main journey passes Fiscal/refusals, acquisition,
active play, both early endings/continuation, Pitch crash/lost-reply recovery/payout, WebSocket
recovery and Recovery's new committed-heartbeat loss plus ordinary cancellation. Exact built
HTML/JS/CSS/Worker proof passes. The sanitized trace records one deliberately failed Recovery
heartbeat delivery, then one native reconnect and two normal resumed beats. Node syntax and final
diff checks pass. The process is terminal; no source edits occurred while it ran.

Limits: deliberate single-response loss after commit, not a whole-network outage, hidden-tab/
watchdog observation, fresh five-minute completion, assistive technology, Linux release host,
whole composed aggregate or hosted CI. Unaffected sibling journeys are not rerun for this tool-only
change. Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new range is
`a1197552` exclusive through this containing test/docs/plan/log commit, pending designated review.
No acceptance flips, archive, mint, push or release promotion. Continue accepted integration and
consolidated review; the typed-dictionary choice is separately requested and remains unruled.

## 2026-10-08 — Recovery watchdog heartbeat becomes an honest terminal transition

MA-C1/MA3 and AC1/AC5 require the existing watchdog receipt, not a new expiry error or
restart policy. RP-464 was reproduced through actual HTTP/Postgres69582 (the registered
progress descriptor rejects the server's terminal) and native Chromium/WebKit72027
(both mounted surfaces crash on non-finite progress and never notify the host; detached
controls pass). Describe the existing watchdog-only cancel arm beside the unchanged
five-key progress arm; regenerate the typed consumer. Consume that receipt in the surface,
stop/unsubscribe the existing scheduler, show the existing terminal message and notify/focus
only while mounted. An intermediate race regression17157 also caught duplicate host
notifications when heartbeat and Stop replies coincide; the terminal transition now commits
once in either order. Coordinator, scheduler, receipt bytes, durations and rewards are unchanged.

RP-465 is the bounded API C2/C18 prerequisite: the old checker rejected response object-to-
union widening despite preserving the old arm. Cold49182 failed three positive embeddings
while the request/shared and breaking controls refused correctly. The checker now requires
a compatible first arm and added closed-object arms provably disjoint from it by required-key
exclusion. Ambiguous/unknown cases and request/shared widening still refuse. Ordinary
generation does not replace the compatibility pin. A structural comparison confirms the
entire API artifact is unchanged except the progress union, whose first arm equals the old
progress object; the separate terminal descriptor and baseline pin remain unchanged.

Executed on final source:
- Cold native publicapi/account tests17651 and affected Go vet pass.
- Full declared Linux-amd64/Postgres publicapi/account/gameserver16776 passes, including
  actual expiry via heartbeat, unchanged Soul and byte-identical subsequent cancel replay;
  the original refresh, security, first-hour and admission populations also execute.
- Native Chromium/WebKit29467 passes all34 executions, including generated-port watchdog,
  detached reply and both Stop/heartbeat orders. Full client84643 passes10321 with981 explicit
  exclusions; typecheck53080 reports zero errors/warnings. Four affected client files pass45.
- Ordinary API generation74975 and staged-output api-check pass without updating the pin.
- Final affected `node client/tools/test-game-ui-composed.mjs`9130 exits0. Required persisted
  populations execute, final client/Worker assets build and load byte-exactly, both early
  endings/continuation, native Pitch process-death/lost-reply retry/payout, WebSocket recovery,
  and native Recovery production suppression/reply-loss/reconnect/cancel all pass. Its trace
  retains the deliberately lost Pitch and Recovery replies, not concealed retries.
- Gofmt and diff checks pass. All test sessions are terminal. Initial bad Go/browser selectors
  and sandbox listener refusal executed no relevant population and are not counted as passes.

Limits: watchdog API expiry uses the injected server clock; native terminal regressions use
controlled replies through the generated port. This is not a natural 24-hour mounted-browser
expiry, full five-minute completion rerun, Firefox/AT, clean Linux release host, whole composed
sibling aggregate or hosted/full-CI verdict. Unaffected sibling journeys were not rerun. No
kernel/balance/catalog/copy/migration/workflow change, false version bump or acceptance flip.
Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new range starts
after `d5b40996` through this containing commit; designated cross-party review remains pending.
No archival, push, release promotion or inferred owner choice. Continue accepted integration
and consolidate corrective review; typed dictionaries, attendance and rights decisions stay held.

## 2026-10-08 — Recovery survives real post-commit gameserver process death

Under MA2/MA3/AC1 and SR-C6, extend the existing real-service journey after its network-only
reply-loss case. An actual scheduled heartbeat commits through HTTP/Postgres, then the driver
SIGKILLs its owned gameserver and withholds that reply. Require visible network pause, no
queued beats across the real catalog ceiling, unchanged complete SQL session/balances/Soul,
different restarted OS process and a recovered WebSocket. Native Reconnect must preserve the
session/progress, rotate its capability, credit no missed interval, then render fresh attended
progress. No new bootstrap or document navigation is permitted. Existing Stop still proves one
stored terminal receipt/event/log and unchanged balances. No runtime, schema, clock, retry,
balance, content or CI-workflow change; reuse the existing process lifecycle and player driver.

Executed `node client/tools/test-game-ui-composed.mjs`55038 and final restored12650: exit0.
All eight persisted refresh cases and required Account/GameUI/Production/Leaderboard/Gameserver
Postgres populations execute. Built-client main journey passes both early endings/continuation,
Fiscal/refusals, active play, Pitch process death/retry/payout, and both Recovery reply-loss cases
plus cancellation. Exact production HTML/JS/CSS/Worker proof passes. Connection refusals during
the deliberate outage are expected diagnostics, not concealed retries.

Fault-control limits are explicit: attempt49874 to reset the disposable session's progress after
restart is rejected by Postgres's nonregression trigger, before the new equality assertion.
Attempt24143 with a transaction-local trigger bypass never reaches Recovery: RP-378 recurs at
the earlier Fiscal invalid arm (Enter targets a non-harvest button, zero request/route/response).
Neither attempt establishes the new equality oracle's fault discrimination. Both temporary SQL
lines are removed before final12650; no trigger, migration or production byte was changed.
The Garage log/ledger retain that recurrence; a subsequent pass does not repair its cause.

Limits: gameserver death after a committed heartbeat, not interrupted transaction/DB/host death,
clean-host R-006, full five-minute completion rerun, background-tab/AT/Firefox, sibling composed
journeys, whole CI or complete MA acceptance. Syntax/diff checks pass; processes are terminal.
Review by: Codex (actual-diff implementer first filter). Recorded by: Codex. Exact new range is
`af610f72` exclusive through this containing commit, pending designated cross-party review.
No acceptance flip, archive, push or release promotion. Next: finish RP-378's bounded readiness/
focus diagnosis, continue accepted construction and consolidate corrective review.

## 2026-10-08 — Complete the Recovery restart data-loss control

After RP-378's driver-readiness repair, normal main32419 exits0 through the real server restart,
native reconnect and terminal proof. For the previously interrupted data-loss control only,
temporary SQL in the disposable test database resets the active session's committed progress
after the actual restart. A transaction-local trigger bypass is necessary because ordinary
regression is correctly forbidden by Postgres. No migration, production helper or persistent
trigger setting changes. Probe23899 reaches the target and exits1 specifically at `server restart
changed the committed Recovery session (capability withheld)`, not a SQL/earlier readiness error.

Remove the temporary line after the process is terminal. Restored helper SHA256
`e8b420c542a691d7299c7825f20b505b0e712e856bbee3bf828945bad1128b12` and main driver
`3a3a49c4e0a8de45c197831fb28c66ce8a81c3034c9f1d18384679024e900829` match normal32419
exactly; there is no residual helper diff against8f5520db. This closes the earlier incomplete
control, not a discovered production data-loss bug, whole reliability claim or MA acceptance.
The Garage log owns the driver repair and exact new pending designated range after8f5520db
through this containing commit. Review by: Codex (executed control); Recorded by: Codex.
Prior incomplete attempts remain recorded, not replaced by this result. No push or archival.

## 2026-10-08 — Native Pitch activation and Tab reachability (RP-478)

Under accepted MA3/UI Foundation keyboard operability, replace programmatic clicks
in the existing full component flow with native Enter/Space for Start, selected cards,
play, buy, end-shop and Return. Keep exact command IDs/revisions/card selection,
play-size refusal, receipt/terminal focus and axe assertions. Existing late-response
cases now leave with native Space and still require no repeat command/refetch/host callback.

Activation-only59640 passes both native engines. Adding actual launcher/terminal Tab
and reverse card traversal exposes WebKit75898: heading→Start goes to body instead
of the button; Chromium and13 other cases pass. Pitch now explicitly declares zero
Tab indices, matching Typer's existing pattern, without positive ordering, key handlers,
copy, wire, retry or pending-disable changes. Final full affected file56075 passes14
in Chromium/WebKit, including Return's exact host callback and no extra gameplay.
Apple documents preference-dependent Tab behavior; this correction is verified for
the actual local WebKit configuration, not a universal browser or assistive-tool claim:
https://help.apple.com/safari/mac/8.0/en.lproj/cpsh003.html

Related units60143 pass74 in the two actual selected files. The earlier guessed
minigame-session-port filename matches no file and contributes no proof; that run
executes only the nine minigame-surface units. Final types/build/boundaries77496 pass,
zero Svelte diagnostics. Initial sandbox listener denial runs no browser tests;
permitted native runs above do. No Firefox/Linux/performance/whole-CI or real-server
rerun: the change is native control focusability, exercised at its mounted boundary.
This does not implement the held GS7 proposal's extra focus/presentation rules or
prove uninterrupted keyboard navigation through the default host/server journey.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new range
begins after7142dd50 through this source/tests/docs/plan/ledger/log commit; designated
review pending. All handles terminal; no acceptance flip, archival, content mint or push.

## 2026-10-09 — Complete designated review of the original Pitch surface

**Review by:** Codex. **Recorded by:** Codex. **Exact reviewed range:**
`50a3a514^..50a3a514` (`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444`, Claude implementation).
**Decision: CHANGES REQUIRED.** Inspected all 23 changed paths: host/runtime/catalog wiring,
six minigame source files, all three affected tests, candidate/generated copy, hash/manifest
bindings, canonical docs and the original log. This completes that original batch's review;
it is not approval of intervening changes or of the whole MA implementation span.

- **RP-478 — keyboard evidence was overstated.** The original browser test labels native
  keyboard selection but invokes `box.click()` and programmatic button clicks. That cannot
  prove keyboard operation. The later retained native WebKit failure at heading→Start and
  corrected Enter/Space/Tab population now supply meaningful component evidence. The
  correction is `7142dd50..a4c09f31`, pending the other party's review.
- **RP-415 — component lifetime was incomplete.** Original `reject`/`refetch` can issue a
  fresh read after Leave/unmount; `bind` also lacks a post-content-await lifetime check.
  The earlier recorded mounted deferred-response regressions reproduce the unwanted second
  read. Their correction and guards remain separate pending-review work, not approved
  retroactively by reviewing the original commit.
- **RP-131 — the original guard explanation was false.** The batch itself adds six files
  under the then-watched minigame prefix without a version signal. Fresh execution still
  rejects this exact commit, rather than merely the older correction it blamed. Full diff
  inspection finds presentation, transport and decoding additions, not simulation/replay
  arithmetic; that does not authorize a false semantic bump, history rewrite or guard bypass.
  Kernel History Guard Integrity remains **draft**, awaiting its owner decision.

The original log honestly discloses that its detailed component enumeration follows the
unadopted GS7 proposal and its new copy is candidate wording. Neither becomes accepted
authority through this review. The two-way tenant/child binding, pinned-content refusal,
same-command retry and generated-DTO transport seam are real implementation, not evidence
of all MA acceptance or full accessibility. No new mechanic or owner text was authored here.

Executed on current source `a4c09f31`, separately from the historical verdict:
`make test-browser-focused BROWSER_TEST_FLAGS='test/minigame-surface-browser.test.ts --project=chromium --project=webkit'`
passes 14 mounted executions (session16628). The root Vitest selector for
`test/minigame-surface.test.ts test/generated-operation-ports.test.ts test/game-ui.test.ts`
passes 173. `make copy-check verify-client-boundary verify-kernel-version` passes generation
fixtures, all 669 current keys/history, deployment manifest and boundaries, then fails on
`50a3a514` (session46868). The independent boundary/kernel67642 run confirms the same red
gate; this duplicate scan adds no coverage.
Read-only checks against the actual original Git objects also verify all 38 new entries exactly
match their generated catalog, TS/Go keys and orphan records, with the original artifact hash
matching the deployment pin. `git diff --check 50a3a514^ 50a3a514` passes.

No current-source edits preceded those runs; all handles are terminal. Current tests are not
a rerun of the original revision, a fresh full server journey, Firefox/Linux/AT, whole CI or
release proof. The existing built-client real-service journey remains separately recorded.
Next: consolidate remaining MA original/corrective review coverage and continue accepted
construction; no archive, feature-completion checkbox, content mint or push follows.

## 2026-10-09 — Original Recovery review and terminal/error race repair

**Review by:** Codex. **Recorded by:** Codex. **Exact original reviewed range:**
`b92a05de^..b92a05de` (`b92a05de3fbd4c8c6b663c008baa2ddb2a910c20`, Claude implementation).
**Decision: CHANGES REQUIRED.** Inspected all22 paths: component/toy/port, shared operation
extraction, host/runtime/catalog, boundaries, both new tests and host test, candidate/generated
copy, deployment pin, canonical docs and original log. This completes the original batch's
disposition, not approval of its corrected descendants or the full MA implementation span.

The original visible-only subscription misses an already-hidden document (RP-145), delayed
start/reconnect can create a scheduler after unmount (RP-146), and delayed Finish/Stop can
notify a departed host (RP-452). The original browser tests use programmatic clicks, not the
claimed keyboard boundary (RP-148). Their recorded failing cases and later corrections remain
separate review ranges. Original candidate copy is not adopted through this review. The
original kernel-red explanation is also false: this batch adds watched `src/soul` presentation
files without a version change, already acknowledged in the later path-relocation record.
That does not authorize a fake arithmetic bump, rewrite or guard exception (RP-131).

Read-only original Git-object checks execute successfully: all18 new candidate entries equal
their generated entries; TS/Go/orphan keys agree; SHA-256 matches the original deployment pin;
removing exactly those18 keys reconstructs the original Go file. Historical diff whitespace
check passes. This is not execution of the original browser revision. The current affected
checks below verify corrected source, not historical acceptance.

**Current correction, RP-480 (MA3/SR-C3):** a pending Finish failure must not overwrite a known
heartbeat terminal or leave contradictory notices beside it. Terminal entry now clears the old
notice; `fail` ignores terminal/detached surfaces. Six controlled generated-client cases cover
transport failure, conflict and not-ready, each before/after the heartbeat receipt. Without a
known terminal, errors must still display; afterward require one exact host receipt, one Back
control, heading focus, no progress/alert/stale notice and an unsubscribed scheduler.

Old-source session69554 reproduces eight failures and four passing reverse-order controls in
Chromium/WebKit. Earlier instrument78361 used the wrong expected watchdog wording; its
reverse-order failures are not regression evidence. Correcting that test expectation precedes
the valid red run. No production copy was changed.

Final commands, all exit0:
- `make test-browser-focused BROWSER_TEST_FLAGS='test/soul-recovery-surface-browser.test.ts test/minigame-surface-browser.test.ts --project=chromium --project=webkit'`
  passes60 native checks (25495).
- Root Vitest selects Soul surface/scheduler, minigame surface and generated operation ports:
 84 units pass (59433).
- `make typecheck build-client verify-client-boundary` passes with zero Svelte diagnostics
  (82928).

These are mounted native-input/generated-port checks with controlled HTTP replies, not a new
Go/Postgres/real-service run, Firefox/AT, whole CI or long-session observation. No server,
retry, timing, payout, catalog or release semantics changed; broader performance/DB reruns are
not required for this local error-ordering correction. All handles are terminal.

**Current-range first filter:** Review by: Codex (implementer). Recorded by: Codex.
Exact new range begins after `5c9d0293` through this source/test/docs/plan/ledger/log and status
checkpoint. Designated cross-party review remains pending; existing review debt does not block
other safe accepted construction. No acceptance checkbox, archival, content mint or push.

## 2026-10-12 — Recovery controls are reachable with native Tab (RP-491)

MA3/MA-C9 require keyboard access. The old Begin/Stop test directly focuses buttons,
so it misses WebKit's omitted implicit Tab stops. Strengthened that test and added
Reconnect/Finish and ended/error Return paths: heading entry, actual forward/reverse
Tab and Enter/Space, exact coordinator callbacks/terminal receipt and host return.
No action button is focused/clicked by these four tests. The heading entry between
states is explicit: this is state-specific traversal, not an uninterrupted whole
player/assistive-technology journey.

Old-source65700 fails all four WebKit paths at Begin; Chromium's four pass. Initial
19661 incorrectly expects Pitch's Back wording and is not valid regression evidence;
corrected the test to Recovery's existing "Back", without editing copy. First sandbox
attempt cannot bind its localhost listener and runs zero tests. The correction adds
only tabindex=0 to existing controls, retaining disabled/order/callback behavior.

Final affected `make test-browser-focused`43808 passes all66 Recovery/Pitch executions
in Chromium/WebKit. Selected Soul surface/scheduler, minigame surface and generated
ports pass84 units. `make typecheck build-client verify-client-boundary`75990 passes
with zero Svelte diagnostics. All handles terminal; actual diff/whitespace checked.
No timing, API, payout, copy, kernel or CI change; no performance/DB/server/Firefox/
AT/whole-CI rerun or full MA/accessibility acceptance is claimed for this DOM fix.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range
starts after `bc5d5f5d` through this containing commit; designated review remains
pending, as do earlier ranges. No acceptance flip, archival, push or deployment.

## 2026-10-12 — Real Recovery play traverses with Tab, including natural Finish

MA3/MA-C9 keyboard gates apply to the real player seam too. Its old helper directly
focused action buttons before Enter, so that journey could bypass unreachable controls.
78149 retains the old-driver failure: after a test-side `tabindex=-1` on Begin,
`assert.rejects` fails with "Missing expected rejection" and the browser issues a start
request. This is an instrument finding, not a new broken-current-product allegation.

Change only the existing helper to traverse from current focus with Tab before Enter.
A bound derived from its actual document control census fails explicitly when unreachable;
no direct action-button focus is left. Keep the excluded-Begin negative, require no Recovery
traffic or stored session on refusal, and restore its original attribute in finally before
normal play. Both normal bounded composition and the explicit full selector consume this
same helper. No product, copy, API, catalog, kernel, duration, cadence or CI changes.

68321 `make test-soul-recovery-composed` completes exit0 against the built production client,
real Go server/Postgres/WebSocket: the negative passes; Tab/Enter drives leave/reconnect,
cancel and both committed-heartbeat reply-loss/process-death arms; then unchanged300000ms
Repot naturally becomes eligible and native Finish grants exactly5 Soul. Stored progress,
terminal receipt/event/log, rendered Soul40→45, unchanged Company balances and reload all
pass. Company production is positively established before Recovery. The explicit Soul40
setup remains disclosed; no natural depletion/history replay, 15/45-minute, Firefox/WebKit,
assistive-technology, full MA, clean-host or hosted-CI proof follows. The shared parent
integration population also executes, including actual public SDK/TS replay consumption.

Two sandbox Docker-socket refusals ran no population; the narrowly permitted existing target
was used for both actual runs. No unrelated containers/volumes were cleaned. Node syntax,
actual diff and whitespace checks pass; all run handles terminal. Docs/plan and RP-491 now
distinguish this real-service supplement from the original controlled-browser repair.
Inspection confirmed catalog owner exports and backup snapshot-origin changes still have
their existing authority holds; neither was implemented or silently adopted.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new test/docs/record
range starts after `15f069da` through this containing commit; Claude designated review remains
pending, consolidatable with earlier Recovery corrections. No acceptance checkbox, archival,
draft adoption, push or release. Continue remaining accepted construction/reviews; shared
clock, long-session renewal, authored contracts and the full nine-tier floor remain open.

## 2026-10-13 — Public Exit refusals cannot strand an active Pitch session

Under accepted MA-C12, extend the existing HTTP/Postgres lifecycle, not a new framework.
Both literal and live purchased-generator scaling populations submit all three ordinary
Exit arms: active, active after application reconstruction, then resolved. All 18 public
requests must return exact intent/outcome/category/detail coordinates. Active attempts
return minigame_session_active; after resolution the ordinary tier/IPO/offer refusals return.
Independent reads require unchanged full Company/Founder save bytes and revisions and
unchanged session state. Existing certified payout, history, retry and Founder replay
checks still pass afterward. Refused commands retain their normal receipt/audit writes.

Initial 52803 fails both terminal Wind Down controls with HTTP 500. Diagnostic 38106 traces
live Founder Exit parity; temporary error-only diagnostic 42081 isolates Fiscal sweep.
The existing flood probe had advanced the injected application clock 15 seconds, then later
writes could precede Exit's real database clock. Remove both diagnostics and let real time
catch up immediately after the existing refill assertion, before later Founder writes;
restore the application's current clock. No bucket reset, saved timestamp rewrite, relaxed
receipt, timeout/retry increase or production backwards-clock repair. Production diagnostic
diff is verified empty. This test setup finding does not settle D-024's shared-clock contract.

Final selected real Linux-amd64/Postgres lifecycle 96317 passes both populations, no skip,
35.55 seconds. `make test-go-ci CI_TEST_PACKAGES='./gameserver'` (88443) passes the complete
affected package in 83.943 seconds. Native `make test-go vet GO_PACKAGES='./gameserver'
GO_TEST_FLAGS='-count=1'` (4270) passes; native DB skips are not SQL proof. Related gameui
unit/vet 39927 and declared Linux/Postgres package 86880 pass; focused preview 57737 passes
its two named tests. An earlier unquoted parenthesized Make selector fails in the shell
before tests and is not evidence. Formatting and actual diff/whitespace checks pass.

Original F9 range `2def3612^..2def3612` already has the 2026-10-04 designated approval above;
reinspection/current checks do not create a duplicate verdict or approve Codex's supplements.
New scope is test/docs/plan/log only. No production, policy, balance, copy, schema, kernel,
CI, acceptance-bound or feature-status change. The added realtime catch-up is bounded by
the existing 15-second flood-clock advance in each population; no suite/performance budget
was changed. No successful eligible Exit, claimed-state race, fresh mounted browser/AT,
whole CI, full MA or release proof is claimed.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range begins
after `c7071178` through this containing commit and needs Claude designated review,
consolidatable with earlier related MA supplements. All checks listed above are terminal;
no acceptance checkbox, archival, draft adoption, push or deployment. Continue accepted
construction; contract decisions and the full nine-tier 1.0 objective remain separate.

## 2026-10-13 — Processing claims retain the public Exit guard

MA-C12's `active|claimed` contract now has the previously missing controlled claimed-state
HTTP proof in the existing two Pitch populations. A real service worker validates the
nonterminal command and pauses at its receipt builder after committing its Postgres claim.
Independent reads require a genuine token/claim timestamp and unchanged frozen session data.
All three public Exit arms refuse with minigame_session_active; complete Company/Founder
bytes/revisions and the held claim remain unchanged. Deliberately refusing the diagnostic
receipt releases the claim, restoring the entire original session except UpdatedAt; the
ordinary public command then runs and existing certification/payout/history/retry/replay
checks remain. No SQL status injection, fake resolver, new public hook or production change.

Executed evidence, all terminal:
- Initial selected Linux/Postgres run46490 passes both populations (35.52s), before tightening
  the post-release comparison from snapshot/revision to the complete session.
- Compiling active-only activity-predicate fault3614 retains idle-session checks but fails
  the first claimed Wind Down: actual tier refusal versus required minigame_session_active
  (16.78s). Exact production query restored; its diff is empty. This is test discrimination,
  not a defect in the restored implementation. No fault is committed.
- Final `make test-go-ci CI_TEST_PACKAGES='./gameserver ./minigame'`66994 passes cold against
  declared real Postgres/Linux-amd64: gameserver80.711s, minigame2.358s. Both lifecycle
  populations remain in that package, including the stronger complete-session comparison.
- `make test-go vet GO_PACKAGES='./gameserver ./minigame' GO_TEST_FLAGS='-count=1'`18888 passes;
  native DB skips are not SQL proof. Formatting and diff checks pass. Existing warned orphan
  services remain untouched; no cleanup, timeout/retry/CI or acceptance-bound change.

The five-second diagnostic context only bounds the held worker and releases it on every test
exit; it is not a new game timeout or measured latency claim. Controlled service/HTTP
interleaving is not a public command-network stall, a create/Exit transaction race, successful
eligible Exit, browser/AT, full hosted CI or full MA/1.0 acceptance. No checkbox or archival.
Review by: Codex (implementer first filter). Recorded by: Codex. The test/docs/plan/log range
after `67d3eece` through this containing commit needs Claude, consolidatable with earlier MA
supplements; those earlier ranges remain separate. No owner decision inferred or push.

## 2026-10-13 — Exact Pitch tenant admission agrees across runtimes (RP-515)

Outcome: conform Go's tenant parser to TP-C12/TP-C14's exact command/snapshot contract,
consumed by accepted MA1. Case-insensitive struct matching admitted wrong-case/colliding
keys; missing/null counters and offer price could become zero; hash admission checked only
prefix/length. TypeScript already refuses these inputs. Go now reuses its existing exact-key
helper, checks required non-null snapshot/offer values and lowercase hexadecimal hashes.
Canonical escaped keys, explicit zero/empty arrays and all three command arms remain legal.
Valid snapshot bytes, content, scoring, payouts and rejection taxonomy are unchanged.

Regression: new `TestTenantRejectsMalformedJSONFields` fails on the old engine (7d5657,
Make exit2, package0.227s), while actual engine-generated playing/shop controls succeed.
The malformed population contains87 snapshot/offer/command/hash cases in each runtime.
Go checks both admission and actual Apply; TS checks actual Apply. Old JSON observation
output was truncated, so no old failure-count denominator is inferred from it. No additional
source severing is needed for this ordinary reproduced regression. Existing API raw-command
admission (RP-444) already protects public requests; this is a tenant/replay inconsistency,
not a demonstrated new public HTTP exploit. Some malformed snapshots were also refused later
by catalog/identity validation; permissive schema admission is not proof they earned payouts.

Final executed checks, all terminal:
- Focused verbose cold Go8585 passes all87 negative subtests and canonical controls (0.317s).
- Native `make test-go vet GO_PACKAGES='./pitch ./minigame ./gameserver ./kernel'
  GO_TEST_FLAGS='-count=1'`78169 passes; native DB skips are not SQL proof.
- Declared `make test-go-ci CI_TEST_PACKAGES='./pitch ./minigame ./gameserver ./kernel'`
  11831 passes on Linux-amd64/real Postgres: Pitch0.499s, Minigame2.362s,
  Gameserver85.498s, Kernel0.033s. Separate verbose HTTP lifecycle89336 explicitly
  passes literal and purchased-generator scaling,35.65s, no skip; existing resume,
  claim/Exit refusal, certification, once-only payout, retry and replay checks remain.
- Actual client Pitch/admission/content-gate43804 passes98 tests across three files,
  including the same87 negatives and the existing shared content/scoring vectors.
- `make typecheck api-check`99649 passes: zero Svelte errors/warnings and byte-unchanged
  generated API artifacts. Formatting and actual diff/whitespace checks pass.

The genuine admission change bumps all three kernel identities184→185. No historic correction,
guard exemption, migration, balance/corpus bytes, player copy, UI, CI policy, timeout or accepted
bound changes. These checks do not establish browser/AT, clean-host, natural pacing, complete
hosted CI or whole MA/1.0 acceptance. RP-131 and all other independent holds remain.
Canonical Pitch docs and this plan/ledger now describe the correction; archived Pitch records
stay untouched. Review by: Codex (implementer first filter). Recorded by: Codex. New range starts
after `6849005f` through this containing commit and needs Claude designated review, consolidatable
with related MA corrections without erasing prior range debt. No acceptance checkbox, archival,
owner adoption, publication or deployment. Continue accepted construction while true contract
choices and the full nine-tier 1.0 objective remain open.

## 2026-10-14 — RP-565 shared raw artifact admission

Resume clean at `3a182e1c`. Previous turn progressed the integration proposal to review-ready,
not adopted; no owner answer inferred. Under existing MA-C15, execute the recorded schema1
conformance finding separately from unaccepted schema2. Thirty shared transformations bind
exactly one occurrence in the real candidate, with canonical and escaped-key/value controls.
Go invokes LoadCatalog; TS loads freshly hashed complete Pitch/replay bundles, not just a
standalone JSON parser or already-decoded object.

Old Go cold run fails8 uppercase root/operation/tenant keys, with22 cases already correct.
Old TS fails17 duplicate/escaped-duplicate and fractional/exponent/rounded-to-one token cases;
13 controls already agree. Its initial failure output dumps whole resolved bundles and is
truncated, so map unexpected success to a fixed diagnostic string and rerun the same old
implementation80643: terminal exit1, all17 failures/13 passes reproduced with complete output.
Do not count the initial truncated diagnostic as a full raw recording.

Correct Go raw shape checks derive mandatory spelling from the original Catalog/Operation/
Tenant DTO tags; duplicate scanning, typed decoding, ordering/version/tenant validation remain.
This is admission, not reflected API schema generation. TS invokes parseIntegerCatalogJSON
before the existing policy parser only for minigame_api. No general replay JSON policy, data
bytes, field/version/operation/tenant identities, save floors or existing good histories change.
Kernel0.3.188→0.3.189 is a genuine watched loader change, with all three version sources updated.

Final Go/vet13555 exits0: minigameapi0.167s, replaycatalog0.444s, kernel0.169s, Account0.476s,
Gameserver0.371s. Node's complete affected replay suite81258 passes134. Native37800 explicitly
selects Chromium/WebKit and passes268, including the30-case vectors and full historical suite;
Firefox is not selected/retried and its preceding zero-execution launch failure remains open.
Type/build/full client/boundary65790 exits0: zero diagnostics,11227 pass/1040 explicit skips.
Whole-history guard38213 terminates2 at the existing50a3a514 RP-131 violation after its checkout
and adversarial fixture checks pass. It stops before evaluating this worktree's changed-prefix
bump: do not claim the whole guard or hosted CI green, bypass it or rewrite that history.

Update canonical platform docs, owning plan, RP-565 and the draft's now-stale prerequisite
sentence in the same batch. Format/diff/source reviewed. No SQL/persisted activation/public
player journey/Linux/artifact-release proof or new metadata2 behavior; Docker/review/author
holds and the full nine-tier1.0 remain open. Review by: Codex (implementer first filter).
Recorded by: Codex. Exact full correction/tests/vectors/version/docs/records range after
`3a182e1c` through this containing commit needs Claude designated review, not approval inferred
from earlier ranges. No acceptance checkbox, archival, draft adoption, mint, push or deployment.

## 2026-10-14 — Preserve newer focus on terminal replies in both hosts (RP-600)

Resume clean `a909db44`; preceding turn made scoped rate-proof progress. Continue accepted
MA3/SR-C3 mounted behavior and the integrated GS0.6 focus floor, not the held performance/API
amendments. Recovery's unconditional post-terminal heading.focus steals a native newer host
control while Finish/Stop is pending, and overwrites focus assigned by its terminal callback.
Four declarations ×Chromium/WebKit fail the old component exactly at that focus assertion,
after the command, outside-control Tab and terminal rendering actually execute.

Inspect the sibling Pitch host before closeout: it has the same unconditional focus. Initial
candidate uses nonexistent .pitch-table and fails four mount assertions; exclude that invalid
instrument run, fix only its selector to actual .pitch, retain assertions/deadlines and rerun
with production unchanged. The valid two declarations ×two engines then fail exactly at newer
focus after the pending shop-close reply/callback. No guessed selector failure is product proof.

Both hosts now snapshot focus before committing the terminal and restore only surface-owned
or unclaimed focus, checking again after tick so a newer surviving target wins. Normal terminal
heading/Return behavior, receipt rendering, one host callback, duplicate/race/unmount guards and
scheduler stop remain unchanged. No new listener/framework, user copy, DTO, server, retry,
heartbeat, policy, schema, balance, version or RFC change. These presentation paths are not
kernel-watched; no path move or cosmetic version bump. Canonical platform/Recovery docs and
owning plan/backlog reconcile the correction in the same batch.

Final root `make test-browser-focused` selects the complete four affected minigame/Recovery
browser/helper files in Chromium/WebKit:106 pass, no deselections/skips in that population.
Exact selector: `BROWSER_TEST_FLAGS='test/minigame-surface-browser.test.ts test/minigame-surface.test.ts test/soul-recovery-surface-browser.test.ts test/soul-recovery-surface.test.ts --project chromium --project webkit'`.
All twelve new regression executions pass alongside ordinary keyboard/terminal/race/unmount
controls. Node minigame-surface/Soul surface/scheduler19 pass; root typecheck zero errors/warnings,
build-client and verify-client-boundary pass. Native deferred port replies prove real mounted
DOM focus/state/callback ordering, not server SQL, authentic payout or whole-player workflows.
No Firefox/Linux/actual-AT/performance/full CI rerun or broader acceptance claim. Existing
Docker/history/Clout/rights/content/release and full-nine-tier obligations remain unchanged.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete two-component/tests/
docs/plan/ledger/log range after `a909db44` through this containing commit needs designated Claude
review, consolidatable with MA3 corrective ranges; no self-acceptance/archival or owner adoption.
Continue accepted player conformance work while unrelated named holds remain. Nothing pushed,
published or deployed; no full Minigame/API/1.0 completion.

## 2026-10-14 — Bind reconnect presentation to the returned activity (RP-601)

Resume clean `40f6df43`; preceding goal turn corrected terminal focus in both hosts. Ground
MA2/MA3 and archived SB25: Start returns the Founder’s existing session with a rotated token.
Actual repository Start selects/returns that active row regardless of the requested activity,
and production serializes its activity/duration. A replacement peer-created session therefore
need not match the old requested activity. This source trace is not a newly executed SQL race.

Begin already resolves response.activity_id through recoveryActivity; reconnect instead reuses
current.activity. Three real mounted/generated-port cases ×Chromium/WebKit fail before correction:
Repot/server-room replacements still display Defrag, and unknown returned activity stays active.
The same JSON/operation decoder emits the exact existing start/progress requests and rejections.
Fix only the owner lookup in reconnect, preserving requested-ID semantics, response-owned
session/token/duration/progress, same-ID rotation and different-ID scheduler rebinding. An
unknown row now takes Begin’s existing generic error and stops beats, not a fabricated old row.
No new global DTO validator, retry/cancel, server, authored copy, cadence or wire/RFC/version change.

Regression success checks replacement title/toy/disclosure, max/value, exact new-session/token
progress body, exactly one old-session beat, two start requests still for Defrag, and no terminal
callback. Unknown control checks error/no toy or progress/no new-session beat/unsubscribed
visibility. This is controlled response/DOM evidence, not naturally reproduced cross-device
server timing, SQL, authentic credit or a full player journey.

Final root `make test-browser-focused BROWSER_TEST_FLAGS='test/soul-recovery-surface-browser.test.ts test/soul-recovery-surface.test.ts test/generated-operation-ports.test.ts --project chromium --project webkit'`
passes308 across the complete selected files, no skipped cases. The six new regressions and
existing same-session, hidden/network/reconnect, terminal/race/focus/unmount controls execute.
Node surface/scheduler/generated-port126 pass; typecheck zero errors/warnings, build-client
and verify-client-boundary pass. Diff checked. No unrelated Go/performance/full CI or
Firefox/Linux/AT/server rerun for this one presentation binding; those larger proofs stay open.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete component/tests/docs/
plan/ledger/log range after `40f6df43` through this containing commit needs designated Claude
review, consolidatable with MA3 corrections. No independent approval, archival, mint, publish,
push, deployment or whole Minigame/API/nine-tier1.0 completion claimed.

## 2026-10-14 — Retire old Recovery heartbeat effects on replacement (RP-603)

Outcome under MA3/SR-C6: at baseline `a12cc529`, a held old-session heartbeat
can survive a visibility/clock missed-ceiling reconnect to a replacement session.
Terminal processing runs inside the surface transport wrapper before the stopped
scheduler's guard; a late terminal ends the new session. A late 404 also writes
shared lastProgressError, so the new scheduler's next missed-ceiling pause wrongly
declares it gone. No production backend race or corrupt DTO is needed in this
controlled generated-port/browser reproduction; server concurrency is not proved.

Regression: six cases (watchdog terminal, resolve terminal, gone error, each for
same-session rotation versus replacement) run in Chromium/WebKit. The old
component fails all six replacement executions and passes all six same-session
controls. Real mounted input/visibility callbacks plus the injected monotonic clock
create the pause; HTTP replies are controlled JSON, not a real Postgres service.
Replacement retains its returned Repot identity/progress, reconnects and dispatches
only its own session/token; the retired token dispatches exactly once. Same-session
terminal/gone still ends the session, and ordinary terminal/error races remain.

Correction: a local opaque owner token follows each bound scheduler. Replacement
and unmount retire that owner before asynchronous response/error effects; merely
stopping on a same-session error does not. The first implementation checked the
current running scheduler pointer instead: the new 12 executions passed, but the full
surface caught two failures in the existing conflict-first Finish/watchdog race.
That over-broad guard was rejected, not the assertions; final ownership preserves
the accepted terminal precedence. No request cancellation/retry, new network
traffic, cadence/reward/timing policy, DTO, player copy or framework scheduler edit.

Finished checks: root `make test-browser-focused` selects
`test/soul-recovery-surface-browser.test.ts test/soul-recovery-surface.test.ts
test/soul-recovery-scheduler.test.ts test/generated-operation-ports.test.ts
--project chromium --project webkit`: 330 pass, no selected skips, 11.00s.
Node's surface/scheduler/generated-port files: 126 pass; `make typecheck` reports
zero errors/warnings; `make build-client` and `make verify-client-boundary` pass.
Diff checked. No affected server/generation/kernel path changes; no false version
bump. Docker remains unavailable from the preceding bounded socket check, so no
SQL/composed/Linux/Firefox/AT/full CI or natural cross-device race claim.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete component/
test/canonical docs/plan/ledger/log span after `a12cc529` through this containing
commit requires designated Claude review, consolidatable with MA3 corrections.
No self-approval, archival, content mint, push, deploy or release promotion. The
full nine-tier 1.0 objective and its remaining accepted construction/owner/platform
requirements remain; this fixes one real lifecycle defect, not the whole feature.

## 2026-10-14 — RP-614 Linux API-consumer wave fails; focused branch passes

At committed `8e4d0877`, execute root `make test-go-ci
CI_TEST_PACKAGES='./publicapi ./publicread ./account ./gameserver'
CI_TEST_FLAGS='-v'` (32379, count1, terminal Make2). Declared Compose forces
the Go process to Linux/amd64 with its existing QEMU AVX2 workaround and real
Postgres16. The project was empty before this test-owned tmpfs service started.

The actual wave fails `TestComposedMinigameAPILifecycleUsesPinnedTenantResolverIntegration`
(44.23s), specifically `literal_scaling` (26.95s); gameserver ends failed at
115.233s. Tool output truncates21,687 tokens in the middle, losing the assertion
message. Preserve that limitation; neither cause nor complete green results for
the other packages are inferred from a truncated transcript. This is not the
full server job or a new diagnosis of the known Clout/history failures.

Focused diagnostic90874: root `make test-go-ci CI_TEST_PACKAGES='./gameserver'
CI_TEST_FLAGS='-run TestComposedMinigameAPILifecycleUsesPinnedTenantResolverIntegration/literal_scaling -v'`,
unchanged source/assertions, same declared environment/count1. Terminal exit0,
18.108s, no skips; whole output retained within the tool budget. This narrower
population is diagnosis, not an unchanged full-wave retry or repaired feature.
RP-614 remains open and unattributed, distinct from RP-459's already corrected
bootstrap-receipt isolation. Do not infer clock/auth/replay or API causation.

Next: retain adequate output on any recurrence and localize its actual assertion
before changing source. No candidate production/test fix, tolerance, deadline,
authentication exception, content mint or acceptance promotion. Verification by:
Codex; Recorded by: Codex. Editorial observation only; existing exact source
review ranges remain open. Remove only the test-owned tmpfs Postgres service
after terminal runs, preserving caches/networks/unrelated services.

Same-context diagnostic supplement51197, explicitly for missing-assertion
capture rather than a fix: repeat the original four-package command once with
adequate per-poll output budget at the same product source. Terminal exit0:
publicapi0.145s, publicread0.238s, account9.798s, gameserver81.922s. Complete
Both test-output chunks retained without middle truncation.
`TestComposedCredentialCleanupRuntimeAndDrainIntegration` explicitly
skips its separate real-minute lane; TypeScript public verifier invocation is
not requested by this Go-only wave and is owned by the root composed journey.
Other selected SQL tests actually run. This is neither the full server job nor
native hosted hardware; the declared emulation workaround stays disclosed.

The original assertion was lost because Codex chose a6,500-token observation
budget, not because Go failed to emit it. That evidence loss is the agent's
error. The later full pass does not recover it or prove a fix, exclusive
context/seed/clock cause or reliability. RP-614 stays open; no more identical
green-seeking repetitions. Final removal uses only the declared `postgres`
selector after terminal checks, discarding test tmpfs; inventory is verified
empty. No durable volume/cache/network/unrelated service or user data removed.

## 2026-10-14 — RP-621 solo membership in a sorted mode list

Accepted MA1/MA-C10 keeps public create server-selected `solo`; Platform MP1 and its catalog
allow both Phase-A modes. The catalog requires byte sorting, so a dual-mode row lists
`async_snapshot` before `solo`. The coordinator incorrectly required modes[0]=solo. Check
membership instead; no request-mode field, async HTTP tenant binding, engine extension, content
adoption or new fallback policy. Kernel209→210 in all three mirrors records the real admission
change, not an attempt to repair historical RP-131.

Extend the existing real-Postgres atomic-start test with solo-only, dual-mode and async-only
diagnostic catalog populations. Rehash/store each bundle normally. The actual Pitch tenant
still executes only solo; the diagnostic extra mode is not a shipped async Pitch engine.
Old source31598 fails only dual-mode with `invalid production intent`; the ordinary solo and
async-only controls pass. Corrected32365 executes all three successfully, package0.925s.
Both admitted lists retain the original five rollback faults, once-only sequence/seed, exact
create receipt/retry, exclusivity, active-session Exit guard and verified Founder history.
Async-only refuses with complete unchanged Founder/Company loads and zero session/create rows.

Finished verification by: Codex. Recorded by: Codex.

- Declared Linux-amd64/Postgres `make test-go-ci CI_TEST_PACKAGES='./production ./gameserver'
  CI_TEST_FLAGS='-run Minigame -v'`91119 passes2.323s/36.111s, including the new three start
  populations, solo/async resolution/replay/fault/retry and both real public HTTP lifecycle arms.
  All selected DB prerequisites execute. This is selected integration, not the full server job,
  native hosted hardware, a browser journey or closure of intermittent RP-614.
- Cold native Minigame/Kernel/Gameserver tests and vet67743 pass; Production vet also passes.
  Native DB skips remain explicit, not SQL proof. Client replay/catalog/resolution/Decimal
  checks18435 pass6522 across four files. Typecheck/formulas/API77930 pass, zero Svelte
  diagnostics and byte-identical generated outputs. Formatting/whitespace checks pass.

Single-commit source/test/version/docs/plan/ledger/log range after `a83d590d` needs designated
Claude review; no self-approval or acceptance/archival checkbox change. Prior review ranges,
author holds, C18/wire/privacy/clock/content and full nine-tier 1.0 requirements stay open.
The preceding actual kernel-history check remains red at50a3a514; no full-CI claim, workaround,
history rewrite, budget change, push or deployment. After terminal checks, remove only this
turn's fresh declared `postgres` service; final inventory is empty. Disposable tmpfs rows are
discarded, with no durable volume/cache/network/unrelated service or user data removed.

## 2026-10-14 — RP-622 replay rejects a solo start under async-only policy

Accepted MA1/MA-C11's recorded start is server-selected solo. Both Founder replay arms now
require solo membership, matching the live coordinator; invalid policy returns the existing
replay error and its enclosing transition restores the entire state. Kernel210→211 records
the genuine Go/TS replay-admission change. No request schema, engine/mode extension, authored
content, rating/payout formula or historical artifact is changed.

Use the committed start fixture and one new four-case shared population: healthy solo/dual
lists, async-only, and async-only after a due automatic Fiscal sweep. Rehash altered diagnostic
artifacts normally; Go checks the typed Pitch context matches every original stored component.
Healthy controls compare literal original receipt/event/post-state goldens. Refusals compare
the complete pre-state, and the due arm first proves an isolated sweep would credit3. Complete
Founder-history readers test the same non-clock-variant history and exact head/receipt.
Old Go63088 and corrected-test old TS execution both fail the two invalid cases: sequence7→8,
due credit0→3, and each full-history verifier returns verified for async-only. Actual ordinary
regressions establish discrimination; no additional mutation ceremony or golden rewrite.

Test-author errors are not product findings:21479 used a full-First-Content loader for a Pitch
fixture without Doctrines;65840 omitted TypeScript's mandatory source:null history coordinate,
making healthy histories fail. Correct those prerequisites before claiming the old baseline.
Typecheck68725 then catches the nullable sweep-control return; explicitly refuse a missing
sweep and rerun the finished client population. No assertion/bound is loosened.

Finished verification by: Codex. Recorded by: Codex.

- Cold selected Production/Minigame replay checks and vet52861 pass. Kernel's name selection
  runs no tests; the separate full Kernel/Minigame run39489 passes their actual populations.
  Native DB skips are not persistence evidence.
- Final four-file Node32270 passes6526; final full replay Chromium/WebKit79980 passes284 with
  no selected skips. Typecheck25350 has zero errors/warnings. Normal replay-fixture-check
  passes without regeneration; its command lacks count1, so explicit cold root counterpart
  6825 reruns TestApplyLoggedCrossRuntimeFixture successfully. Formula/API outputs remain identical.
- Declared Linux-amd64/Postgres62622 passes selected Production2.391s/Gameserver35.296s,
  including new replay cases, actual solo/dual starts and async-only refusal, solo/async atomic
  resolution/replay/faults and both public HTTP lifecycle arms. No skipped selected prerequisite.

Designated Claude review remains pending for the single commit containing this entry, parent
`ca186f94`, all twelve source/test/fixture/version/docs/record paths; earlier review debt stays.
This is controlled rehashed-fixture verification, not a naturally persisted/public/deployed
forgery, async Pitch capability, actual AT, full platform, whole CI or release proof. RP-131's
historical red guard, RP-614 and all owner/author/content/clock/full-nine-tier holds remain.
No mint, archival, push or deployment. After terminal runs, remove only this turn's fresh
declared postgres test service; final inventory is empty. Disposable tmpfs rows are discarded,
with no durable volume/cache/network/unrelated service or user data removed.

## 2026-10-14 — Founder start replay honors Fiscal and Soul admission (RP-624)

Outcome: MA1 requires the pinned Fiscal unlock and Soul gate; MA-C11 owns the logged Founder
sequence. Go/TS replay now checks those Founder-owned inputs before incrementing the sequence,
using the existing catalog policies. Live admission, error precedence, public wire, seed, payout,
receipt and event semantics are unchanged. Kernel0.3.211→0.3.212 is a genuine replay acceptance
change, not a workaround for RP-131. Company tier/scaling frozen-input gaps are not invented away.

Regression: expand the existing shared mode population from four to nine cases. Keep the original
controls, remove the held Fiscal unlock or set Soul to the pinned locked band's last value, and
repeat each refusal with a proved due Fiscal sweep. The first unlocked Soul value is admitted.
Both pre-state and counterfeit head carry the same eligibility change, so full-history refusals
cannot pass merely because of mismatched heads. Original receipt/event goldens remain literal;
unchanged controls pin the original post-state golden. No original fixture/artifact is rewritten.
Old Go29150 and Node95319 fail all four newly locked cases: sequence7→8, due credit0→3, and both
non-clock history readers return verified. This is a controlled fixture finding, not a reproduced
public exploit or naturally persisted forgery. Ordinary regressions establish discrimination.

Finished verification by: Codex. Recorded by: Codex.

- `make test-go GO_PACKAGES='./production ./minigame' GO_TEST_FLAGS='-run Minigame -v'` passes
  cold; native SQL skips are explicit, not integration evidence. Full Kernel/Minigame/Decimal
  packages and affected vet pass. Four-file Node run passes6531; typecheck has zero errors/warnings.
- Full replay Chromium/WebKit passes294 (147 each), no selected skips. Initial sandbox loopback
  denial runs no tests; the permitted same command executes. Production client build passes.
- `make typecheck replay-fixture-check formulas-check api-check` passes; the replay fixture runs
  count1 and all generated contract/formula bytes remain unchanged.
- Declared Linux-amd64/Postgres52703 passes Production2.696s and Gameserver38.520s, including
  solo/async resolution/faults and both literal/purchased-scaling public HTTP lifecycles. The
  subsequent test-only extension reuses the live start population for six cases: missing unlock,
  locked and first-unlocked Soul alongside original modes. Final selected Production73170 passes
  2.938s, with all SQL cases executing. Refusals preserve complete saved Founder/Company state and
  zero sessions/create receipts; the admitted boundary retains fault/retry/history/Exit checks.
  Vet/build24484 passes after that extension. This is declared amd64 emulation on an ARM64 Docker
  daemon, not native hosted hardware, whole server CI, natural player pacing or release-artifact proof.

Review by: Codex (implementer first filter); Recorded by: Codex. Designated Claude review remains
pending for the single commit containing this entry, parent `ac88de36`, all source/test/fixture/
version/docs/record changes. Prior review debt, RP-131/RP-307/current CI and release holds remain;
no feature acceptance, archival, mint, push or deployment. After terminal tests, removed only the
fresh declared Postgres test container and its disposable tmpfs rows. Final service inventory is
empty; no durable volume, cache, network, unrelated service or user data was removed.

## 2026-10-14 — Start replay validates live identity grammar (RP-625)

Outcome: MA-C11's replay-owned start follows the live Store/Platform C2 identity contract:
lowercase UUIDv7 session ID, distinct logged intent ID, and Company stream UUID version1–8
with the RFC variant. Reuse Go's existing identity patterns; constrain only the TS start arm,
not unrelated UUID consumers. Kernel0.3.212→0.3.213 records changed replay admission.
Live creation, public wire, seed, receipt/event semantics and original goldens are unchanged.

Regression: shared fourteen-case coordinate population uses the actual committed Pitch bundle,
pre/head state, seed, command and literal receipts. Only the named session or frozen Company
coordinate changes; counterfeit receipts carry that same session so head/receipt mismatch cannot
fake a refusal. Old Go7357e3 admits all eleven invalid cases, changes state and verifies each
history. Old TS58098 admits session=intent and Company UUID versions0/9 and verifies all three.
Original and version1/version8 Company controls remain accepted. These are syntax controls,
not proof of Company ownership/existence or a persisted/public/deployed exploit.

Initial Go test setup was invalid: byte-valued resolved input and a missing typed API/catalog
artifact made healthy controls refuse. Correct both and prove the old behavior before relying
on the result. Typecheck separately caught unknown input/receipt types; corrected the test types
without weakening outcomes. All source probes restored before final verification.

Verification by: Codex. Recorded by: Codex. Finished root checks:
- Cold selected Production64633 passes0.335s; native SQL cases explicitly skip. Full
  Kernel/Minigame/Decimal packages pass in82876; that earlier command's Production failure is
  the disclosed fixture setup, not a green whole command. Final affected vet passes.
- Four-file Node53470 passes6545. Full replay Chromium/WebKit73679 passes322, no selected skips.
  Typecheck/replay-fixture-check/formulas-check/api-check/build17678 passes, zero type warnings,
  cold original fixture comparison and unchanged generated contract/formula bytes.
- Declared Linux-amd64/Postgres53304 passes Production3.261s and Gameserver35.824s: six real
  start cases, all shared admission/coordinate cases, solo/async resolution/faults/retry/replay,
  and both public HTTP lifecycle arms. No selected SQL prerequisite skips. This is amd64
  emulation on ARM64 Docker, not native hosted hardware, whole CI or release-artifact proof.

Review by: Codex (implementer first filter); Recorded by: Codex. Complete single-commit range
after `4228d038`, including tests/version/docs/records, needs designated Claude review; related
start corrections may be consolidated with exact coverage. Do not restart completed original
Pitch/Recovery reviews. Company input/ownership contracts, earlier review debt, numeric/CI,
content and full-nine-tier release obligations remain. No acceptance, archival, push or deployment.
After terminal checks, removed only this turn's test Postgres container and disposable tmpfs
rows; final service inventory is empty. No durable volumes, caches or unrelated data removed.

## 2026-10-15 — RP-122 terminal producer→account deletion observation

Extend the existing real Pitch/Soul lifecycle only after all original producer,
certification/payout, receipt/retry and bootstrap checks. The predeclared method
and result live in platform-alignment/data-rights-minigame-soul-delete{,-plan}.md.
Both literal/purchased-generator fixtures now join their actual API-produced
terminal rows to authenticated Account DELETE. No directly seeded session/receipt.

Final declared Postgres36269 passes the whole parent36.10s and the classifier:
each population retains15 SQL rows across six selected families identically,
including both terminal Soul progress tokens (2→2). Delete204/account-row removal,
archived/unlinked Founder and post-delete access/refresh401 hold, with valid access
before deletion. Missing-ID queries observe zero; neutral classification separately
distinguishes all five outcomes and does not require retaining these rows forever.
Initial99229 also passes before the explicit token-count supplement. Final native
Gameserver33424/vet pass. No payload/token/digest is printed. Stable ordered SQL
JSON equality is not original HTTP-byte identity or complete nested classification.

This replaces only the earlier seeded-only terminal-producer blind spot, not
active/claimed deletion, Guild/Transport, other tenants/modes, full history/export,
browser/participant rights or delete/restore. Existing injected-clock/catalog setup
is not natural growth/pacing. No product behavior, retention decision, schema,
migration, kernel/content, client or CI changes. Inventory, RP-122 and D-009's
evidence route are reconciled without adopting desired retention or erasure.

Review by: Codex (test/research implementer first filter). Recorded by: Codex.
Whole tests/research/ledger/decision-evidence/log range after `6a6a32b9` through
this containing commit needs designated Claude review; consolidate with related
real-service evidence. No feature acceptance, archival, release, push or deployment.
After terminal checks remove only this turn's inspected disposable tmpfs Postgres
service/rows; preserve durable data/cache/network. Full nine-tier goal remains.

## 2026-10-15 — RP-122 active Pitch producer→deletion supplement

Reuse the existing HTTP Account/Fiscal/Pitch preparation in both scaling modes,
play one real nonterminal command, then authenticate DELETE and observe the same
captured session/Founder identities. Keep both original terminal lifecycles intact.
Method/results and limits live in the existing rights plan/dossier; no direct
session/receipt/Founder-unlock seed or new product/privacy policy.

Both active cases delete204/account absent/Founder archived-unlinked. Four produced
families (session, immutable command, create receipt, command receipt) remain1→1
SQL-content-identical. Faucet/Soul remain0→0 absent. Current active read succeeds
before deletion; access/refresh and retry of the actual earlier command become401.
The observer permits future erasure/modification, not a retention requirement.
Fix absent→absent classification: exact old branch22160 falsely says removed and
fails the new unit case; restore correction. Initial actual Postgres3186 passes.

Final cold declared Postgres71788 runs both original terminal populations36.12s,
both active populations1.32s and classifier controls, package37.526s, no selected
skip. Command selector:
`make test-save-integration SAVE_TEST_PACKAGES='./gameserver' SAVE_TEST_FLAGS='-run "Test(ComposedMinigameAPILifecycleUsesPinnedTenantResolverIntegration|ComposedActiveMinigameAccountDeletionIntegration|RightsRowObservation)" -v'`.
Native Gameserver/vet84116 pass cold. Diff checked; all handles terminal. Remove
only inspected owned tmpfs Postgres; preserve caches/network/durable data.

No product, schema, migration, kernel, content, client or CI changes. Controlled
catalog/clock evidence is not natural pacing, claimed/concurrent deletion, active
Soul, other tenants, browser/export, retention/expiry or restore proof. RP-122
stays open for D-008/D-009/D-015 and accepted implementation. Review by: Codex
(test/research implementer first filter). Recorded by: Codex. Whole eight-path
test/research/ledger/decision-evidence/log range after `144865df` through this
containing commit needs designated Claude review with the earlier rights ranges.
No acceptance, archival, release, push or deployment. Do not repeat these now-
executed active/terminal populations as missing research.

## 2026-10-15 — RP-600/RP-601/RP-603 independent corrective approval

Review by: fresh independent Codex sub-agent `/root/recovery_ui_corrective_review`.
Recorded by: Codex parent, transcribing the returned verdict under Marco's
2026-10-15 corrective-code delegation. **APPROVED**, no actionable findings, for
all changed source/test/docs/records in these exact base-exclusive ranges only:

- RP-600: `a909db44ea0de47166824b9accdcb96af729bf91..40f6df43cd0a0f69e15b01b08fd94306e9ff020d`.
- RP-601: `40f6df43cd0a0f69e15b01b08fd94306e9ff020d..089849ad47717746e47e4f6533b1c4ff87bd13e2`.
- RP-603: `a12cc52956263c8991decec78e779aba6984d0c1..d1fb8d574ad46e69494f67843630598a65908193`.

The reviewer independently inspects complete diffs, focus restoration, returned
activity binding and scheduler ownership. Root current Chromium/WebKit six-file
selection passes366, Node four-file selection135, types zero errors/warnings,
build-client/client-boundary and exact-range whitespace checks pass. Checks run
at288ff35c; affected source/tests match the reviewed corrections. An initial
localhost EPERM runs no browser tests; narrow escalation supplies actual evidence.

A temporary test-only Vite loader replays real pre-fix components without editing
runtime source: RP-600 has12 expected focus failures, RP-601 six replacement/unknown
activity failures, RP-603 six replacement failures with six same-session controls
passing. Browser selectors are `preserves newer focus`, `binds returned activity`
and `retired Recovery heartbeat`; complete current file selector is the six-file
union named in the three original entries plus generated-operation ports.

This closes only those corrective reviews. Intervening commits, prior feature
ranges, new normative/research work and other replay corrections are not covered.
No SQL/backend-race/composed/Firefox/Linux/AT/performance/full-CI, feature acceptance,
archival or release claim. The reviewer makes no repository edit or commit; parent
reconciles current plans/ledger in this record-only change. No push/deployment.

## 2026-10-15 — RP-621/RP-622/RP-624/RP-625 independent corrective approval

Review by: fresh independent Codex sub-agent `/root/minigame_start_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. **APPROVED**, no
actionable findings, every source/test/fixture/version/docs/record path in these
four base-exclusive ranges, not unrelated intervening commits:

- RP-621: `a83d590d25bdef54c9d7cb428d8798a79ea76589..ca186f94416f3959a77720cae4ae45fc6df3d790`.
- RP-622: `ca186f94416f3959a77720cae4ae45fc6df3d790..1f1a54818a07f4da5382a1568617ce6d7cbcda8c`.
- RP-624: `ac88de36f6db36d78eae4e0c739350c6f13069a4..4228d038d318e1d604be9310d3c47fe9edab79b0`.
- RP-625: `4228d038d318e1d604be9310d3c47fe9edab79b0..5210d56164492989f69ef1a58ed26973f95231cb`.

Under MA1/MA-C11 and Platform MP1/C2, membership admits sorted dual-mode policies,
replay enforces Founder-owned gates/live identity grammar, refusals restore complete
Founder state. Historical old-code failures inspected against preimages/regressions,
not re-mutated. Source/tests/shared start fixtures match5210d561; later replay.ts
changes are unrelated Reputation behavior. Checks execute at clean015ae5e4:

- Real Postgres: `make test-save-integration SAVE_TEST_PACKAGES='./production ./save ./gameserver' SAVE_TEST_FLAGS='-run Minigame -v'` passes Production3.394s/Save0.085s/Gameserver37.566s.
  Six live-start populations, nine replay admission cases, fourteen identity cases,
  rollback faults, exact retries, sequencing, Exit guard, persisted histories and
  both HTTP lifecycle arms execute; no selected SQL prerequisite skips.
- Node replay/catalog/resolution/Decimal6545 pass; complete replay Chromium/WebKit322
  pass. Affected vet/types, cold replay-fixture, formulas/API and range diff checks pass.
- Full native Minigame/MinigameAPI/Kernel/Decimal pass. Full Production remains RED
  on all27 known TestAxisAccrualPartitionProperty divergences, not concealed by the
  separate scoped Minigame/SQL pass. No broader server green is claimed.

Invalid browser selector and sandbox-listen attempts execute no tests; corrected
narrow loopback escalation passes. Linux-amd64 SQL containers on ARM are not hosted
hardware. Reviewer makes no edits/mutations/commits; browser/DB slots released.
This discharges only these corrective reviews, not prior feature spans, normative
adoption, Company-owned replay choices, full CI, archival or release approval.

## 2026-10-15 — Exact Pitch tenant admission independently approved (RP-515)

Review by: /root/pitch_tenant_admission_corrective_review (fresh independent Codex,
Marco-delegated corrective code). Recorded by: Codex (root). APPROVED complete
6849005f146a3dd9eca28ab169b48b7488004e85-exclusive through
9cbf33f04961958109231381e4416e2a7cb139e6-inclusive, one commit/all10 paths.
At137f7a8d, engine and both admission test files match endpoint bytes; shared exact-key/
duplicate helpers unchanged. Accepted TP-C12/TP-C14/MA1 conformity; correct original
kernel184→185. Later catalog tightening/draft contracts are outside this verdict.
Safe exact old-Go overlay reproduces malformed accepts (0.197s, exit2), no invented
historical denominator. Cold native Pitch/Minigame/Kernel/vet pass; native DB skips
not credited. Actual three-file TS98, types/API generation unchanged and race1.504s
pass. Declared Linux-amd64 packages execute actual Minigame SQL populations; both real
HTTP/Postgres lifecycle branches pass36.51s (literal18.78/purchased17.73), no skips.
Formatting/whole-range whitespace pass. No content/scoring/payout/schema change or new
public exploit finding. Emulation is not hosted hardware, full feature/rights/CI acceptance.
All handles terminal. Ignored overlay files removed; only twice-inspected owned tmpfs
DB76aa21a6… stopped/removed, synthetic data unrecoverable; same two builders/cache/volumes
preserved. No tracked edits/commits by reviewer. Corrective review closes; no normative
adoption, archival, feature or release approval and no historical guard waiver.

## 2026-10-16 — Schema1 raw artifact admission independently approved (RP-565)

Review by: /root/minigame_catalog_admission_corrective_review (fresh independent Codex under
Marco's corrective-code delegation). Recorded by: Codex root, transcribing its verdict.
Decision: APPROVED, no findings, complete13-path single-commit range
3a182e1c9d16ef6970ec0a4503b8a9815039c9e1-exclusive..bd0c9b9787cda419cf7ec6183d36760107040166-inclusive.
All source/test/shared-fixture/three-version/docs/plan/log/backlog and draft-prerequisite
paths inspected. Existing MA-C15 schema1 conformance, genuine original188→189; unaccepted
draft edit is factual reconciliation, not schema2 adoption/new role/route authority.

Exact historical Go ignored overlay independently reproduces8 alias admissions with22
controls passing, package0.172s/Make2. Reviewer final text does not identify a reliable old-Go
session number; none is invented here. Old TS Vite shim84685 reproduces17 admissions/13
controls (161 deselected), actual freshly hashed complete bundles. Both shims removed.
Current cold five Go owners45106 pass, affected vet passes; Node replay86490 passes191.
Initially named nonexistent decimal.test.ts contributes nothing; discovered decimal-vectors
33843 separately passes6304. Native Chromium/WebKit78596 passes382 after initial sandbox
listener denial runs0. Types57767 zero diagnostics, build/boundary45812 and formatting/range
whitespace pass. No Firefox/AT or whole browser/CI reliability claim.

Declared Linux-amd64/Postgres69326 executes selected raw-admission/load/golden/version and
actual HTTP/Pitch/Soul lifecycle without selected prerequisite skips: MinigameAPI0.050s,
Replaycatalog0.970s, Kernel0.034s, Decimal0.453s, lifecycle36.26s(literal18.88/purchased17.37).
Persisted history/retry/reconnect/refusals execute. Verbose golden subcase output truncates,
not a complete raw recording; terminal package success/later lifecycle completion observed.
Both deletion observations retain6selected families+2Soul tokens, not rights completion.

Execution0c8ff6c1: Go admission/scanner/raw fixture endpoint-identical; replay has later
unrelated changes but exact admission call/population retained. Current kernel is not
retrospectively189. No full Production/27-partition repair/history waiver/full CI/hosted
hardware/performance/feature/archival/release approval. All handles terminal; only owned
twice-inspected tmpfs8681caf8… stopped/removed, disposable rows unrecoverable. Durable data,
caches/networks/two existing builders preserved; no tracked edits/commits/publication.
