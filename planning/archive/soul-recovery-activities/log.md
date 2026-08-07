# Soul Recovery Activities implementation log

## 2026-08-07 — Codex acceptance review: blocked (SR-C1–SR-C8)

Review by: Codex. Recorded by: Codex.

The three zero-output activities fit the Soul covenant, but the draft assumes a public coordinator,
presentation catalog, and UI surface that do not exist. It also omits literal rows and the complete
multi-artifact first-Soul mint. Eight blockers with proposed resolutions are filed in the RFC; no
production Soul artifact, copy, or client toy was introduced.

## 2026-08-07 — Codex implementation-readiness review: blocked (SR-C9–SR-C14)

Review by: Codex. Recorded by: Codex.

The owner rulings choose the correct architecture, but do not enumerate the authenticated HTTP
wire, limiter, copy IDs, or production epoch. The body also retains the retracted “zero server
mechanics” claim and UI-shipping language while UI Foundation is blocked. Six narrow blockers with
proposed contracts are filed; no public API, artifact mint, or client surface was improvised.

## 2026-08-07 — Fixture content, coordinator API, and scheduler implemented

Implemented the exact eight-field activity grammar in Go and TypeScript, the three literal fixture
rows, twelve copy-pipeline entries, authenticated four-route coordinator API with server-owned
Founder/Company/session identity, the session-keyed 6-token catalog-rate limiter, and the
framework-neutral visible-only scheduler. Focused Go, TypeScript, copy, replay-corpus, and kernel
guards pass. No epoch manifest or production artifact set changed; UI toy AC4/AC5 remain carried.

Implementation handoff only. Ready for the mandatory review gates after the complete RFC range is
assembled; this entry does not authorize archival.

## 2026-08-07 — Production-shaped lifecycle proof and typed API failures

Replaced the recovery API's error-string classification with typed sentinel matching. Extended the
real-Postgres coordinator fixture to include the three exact production rows and drove every row
through start, ceiling-bounded heartbeat accumulation, resolve, exact Soul credit (including
saturation), immutable Founder replay, and terminal idempotency without a test-only attendance
shortcut. The shared replay corpus was regenerated because the fixture bundle bytes intentionally
changed.

Focused Go tests, client typecheck, and the production package's full Postgres integration tests
pass. The implementation remains fixture-first: the First Content Epoch mint and UI toy AC4/AC5
stay open by contract. Ready for designated review; this entry does not authorize archival.

The subsequent full `make verify` completed Go, formulas, harness, client typecheck/build, and
6,587 client assertions, then correctly failed the history guard: commit `d4c2312` includes a
wire-identical typed-sentinel refactor in kernel-watched production files without a kernel bump.
That refactor must be dropped under the unpushed/unreviewed behavior-identical history carve-out;
it must not receive a false kernel semantic version. No green full-gate claim is made here.

## 2026-08-07 — Approved rewrite complete; ready for designated review

The owner-approved prepublication rewrite replaced `d4c2312` with `ab9d15e`. The replacement drops
only the typed-sentinel edits from the three kernel-watched production files; the exact activity
lifecycle tests, replay corpus, plan state, and evidence remain. The implementation range is
`4973c8e^..ab9d15e`.

A fresh root `make verify` completed successfully after the rewrite: Go vet/tests, formulas,
population invariance, balance harness, TypeScript/Svelte typecheck, production build, 6,587 client
unit assertions, kernel history/adversarial guards, copy/schema/content-manifest checks, and 19,770
browser assertions all pass. No production epoch was minted and AC4/AC5 remain carried to the UI
successor. Ready for cross-party designated review; Codex does not authorize archival.

## 2026-08-07 — designated cross-party verdict: Soul Recovery (4973c8e^..ab9d15e) — NOT APPROVED (narrow)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

**Server side verified APPROVED-QUALITY with adversarial probes:** SR-C10 coordinator commands
(identity from claims only, client founder_id rejected 400, strict decoding, closed error pairs);
SB25 rotation (probe: removing rotation fails the suite at soul_recovery_integration_test.go:194);
SB24 starvation closure (probe: zeroing accumulation fails at :239; all three literal activities
driven through real multi-transaction heartbeats and replay-verified); SR-C6 limiter literals
(burst 6, refill ceiling÷6, 429 rate_limited/recovery_progress, no mutation on limited flood);
SR-C12 rows + all 12 copy keys byte-exact; zero-reward law (company cash "0" post-resolve, no
facts leakage); SB26 watchdog coordinator-preflight-only with forced resolution; replay
determinism; fixture-first containment (constants_hash untouched); kernel 0.3.75→0.3.76 lockstep
with ab9d15e correctly bump-free. Both gate suites independently green at ab9d15e.

**BLOCKING (route to Codex; both inside the SR-C14 accepted scope, NOT the AC4/AC5 carried debt):**
- **F1 (MEDIUM): the scheduler never treats a missed ceiling as a pause.** SR-C6 rules "missed
  ceiling = pause, reconnect via start before resuming"; `client/src/soul/recovery-scheduler.ts`
  `#beat()` checks only backwards time — after any forward gap (sleep-while-visible, long hidden
  interval) it dispatches with the OLD token and resumes directly. The first scheduler test
  ENSHRINES the deviation (10,000 ms hidden gap vs the 5,000 ms fixture ceiling resumes with the
  original token and asserts success).
- **F2 (LOW-MED): two ruled test-matrix cases missing** — duplicate retry (the `#requestInFlight`
  guard is untested) and the watchdog terminal case.

**OWNER-SIDE RULING (recorded now — closes the DESIGN-GAP the reviewer flagged inside F1):** the
ruled scheduler interface carries no ceiling input, and it does not need one. **The canonical
missed-ceiling inference is `elapsed > 3 × beat_interval_ms`** — exact, because SR-C6 defines
cadence = ceiling/3, so 3×interval IS the ceiling in the scheduler's own units. The fix is
confined to the scheduler: on any elapsed gap > 3×beat_interval_ms (visible or hidden), enter the
paused state, emit `on_pause`, and require an upstream `reconnect()` (the SB25 start-as-reconnect,
rotating the token) before beating resumes. No interface change, no server change. The enshrining
test must be corrected to assert pause+reconnect for over-ceiling gaps (an under-ceiling hidden
gap may still resume directly).

**Non-blocking, recorded:**
- **F3:** string-contains error classification in `writeSoulRecoveryResult` is the accepted
  residue of the approved d4c2312 rewrite; reinstate typed sentinels in the next commit that
  honestly bumps the kernel (standing rider, not a gate).
- **F4 (intended readings, now canonical):** (a) transient store errors mapping to 404
  `unknown_id` follows the pre-existing route pattern — acceptable; (b) `session_expired` never
  surfaces as an API error — expiry is delivered as the watchdog-cancel receipt at preflight per
  SB26, with `session_expired:true` reserved for ordinary intents. Both are the intended readings.

**Consumed:** exactly {4973c8e, ab9d15e} + docs-tier {f04c2f3, d1cd39c}. **Range-union:** relative
to the Soul Foundation closing endpoint 3ff2082, the path-filtered implementation span is exactly
{4973c8e, ab9d15e}; all intervening commits are docs-tier. No uncovered edge commits.

**Verdict: NOT APPROVED pending F1 + F2 (scheduler pause + two test cases); re-review is narrow
(the scheduler delta only). Archival blocked until then; everything else in the range needs no
rework.**

## 2026-08-07 — Codex narrow scheduler remediation; ready for designated re-review

Implemented by: Codex. Recorded by: Codex.

- **F1 closed:** the scheduler derives the missed ceiling exactly as
  `3 * beat_interval_ms`. A forward or backward clock gap beyond that boundary enters an explicit
  reconnect-required state, emits `on_pause("network")`, and cannot dispatch again until
  `reconnect()` rotates the token. An under-ceiling hidden interval resumes immediately without
  replaying queued beats.
- **F2 closed:** the fake-clock matrix now covers an in-flight duplicate attempt and watchdog
  terminal delivery in addition to over-ceiling pause/reconnect, under-ceiling resume, and network
  token rotation.
- The semantic scheduler change is inside the existing guarded `client/src/soul/` kernel path.
  The full gate correctly rejected the review's tentative no-bump expectation, so the implementation
  records the honest kernel transition `0.3.77 -> 0.3.78`; no guard exception was added.

Root `make typecheck` reports zero TypeScript/Svelte diagnostics; `make test-client` passes 6,598
tests (3 skipped); and a complete root `make verify` passes Go vet/tests, the balance harness,
client typecheck/build, all 6,598 client tests, kernel/history guards at `0.3.78`, copy/schema gates,
and 19,803 browser assertions. No production epoch was minted. Ready for the narrow designated
cross-party re-review; this entry does not authorize Soul Recovery archival.

## 2026-08-07 — designated cross-party re-review: scheduler remediation 3cfc0e6^..3cfc0e6 — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

Narrow scope confirmed (one commit: scheduler + tests + kernel parity bumps + planning; no server
behavior files). Both blockers closed with discriminating tests:
- **F1 CLOSED:** `#missedCeilingMS = 3 × beat_interval_ms` (the ruled inference, with an integer
  overflow guard); the forward-gap check exists on BOTH paths (`#beat()` for sleep-while-visible,
  `#setVisible(true)` for hidden intervals) routing to `#pauseForReconnect()` (single
  `on_pause("network")`, latch, beat refusal until `reconnect()` rotates the token and emits
  `on_token_rotated` + `on_resume`). Under-ceiling hidden gaps resume directly, test-pinned. The
  previously-enshrining test is corrected (over-ceiling gap asserts pause + rotated-token resume).
  Probe: neutering the ceiling comparison fails exactly the corrected test.
- **F2 CLOSED:** in-flight dedup test (never-resolving transport + hide/show cycle, exactly 1
  dispatch) and watchdog-terminal idempotence test both present; probes on each guard fail exactly
  the matching test.
- **Kernel honest:** 0.3.77→0.3.78 in all three parity points; guard + adversarial fixtures
  independently green. **Gates:** all `make verify` components pass at 3cfc0e6 (6,598 client,
  19,803 browser); Postgres suite waived (client-only delta).
- **F3 rider — assessed NOT TRIGGERED, remains standing, wording re-anchored:** 3cfc0e6's bump is
  client-path-only; folding a server error-taxonomy refactor into a narrow client remediation
  would have broken remediation-only scope. The rider is re-anchored as: **typed sentinels return
  in the next commit that touches `server/account/` behavior or bumps the kernel for a
  server-path change** — it can no longer be skipped by client-side bumps.
- **Non-blocking observations recorded:** O1 — the `#requiresReconnect` latch's solely-owned
  effect (no retry after transport error until reconnect) has no pinning test; suggested future
  test recorded (coverage note, behavior doubly implemented). O2/O3 — harmless `on_resume`
  emission nuance; rapid-toggle cadence contained by the server-side SR-C6 limiter authority.

**Verdict: APPROVED. Combined consumed set {4973c8e, ab9d15e, 3cfc0e6} + docs-tier
{f04c2f3, d1cd39c} unions to the complete Soul Recovery implementation span relative to the Soul
Foundation closing endpoint 3ff2082 (intervening commits are docs-tier or The Pitch's own APPROVED
set). No uncovered edge commits. Soul Recovery is ARCHIVAL-ELIGIBLE**, subject to the carried
AC4/AC5 UI-successor debt (unchanged). The archival move is Codex's to execute citing this entry.

## 2026-08-07 — archival rotation

Implemented by: Codex. Archived under: designated cross-party verdict `5754901`.

The consumed set is exactly `{4973c8e, ab9d15e, 3cfc0e6}` plus docs-tier
`{f04c2f3, d1cd39c}`. Canonical shipped behavior is distilled into `docs/soul-recovery.md`.
AC4/AC5 remain explicit UI-successor debt, and the production artifact mint remains solely owned
by the First Content Epoch RFC; this rotation mints no content and deploys nothing.
