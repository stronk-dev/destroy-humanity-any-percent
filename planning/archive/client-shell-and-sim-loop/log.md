# Client Shell & Sim Loop — append-only log

## 2026-07-29 — acceptance and start

- Owner explicitly directed implementation of the RFC queue. Reviewed the RFC against the binding stack, browser-rendering research, and adopted `design/11 §1` return sequence.
- The draft's offline-return blocker was stale: design already resolves the sequence. Added C1-C6 for the abstract authoritative stream, worker messages/clock, strict policy data, reconciliation/display state, lifecycle behavior, and action/amplitude boundaries.
- Pinned current registry releases Svelte 5.56.8, Vite 8.1.5, and `@sveltejs/vite-plugin-svelte` 7.2.0. No push will be performed.

## 2026-07-29 — foundation and acceptance audit

- Landed the strict policy catalog, Svelte/Vite shell, abstract snapshot/receipt boundary, dedicated
  prediction Worker, reconciliation/display primitives, lifecycle hooks, intent dispatcher, and
  Node/browser/boundary gates in `5bcc00f`.
- Full-diff self-review found two correctness gaps before archive. First, prediction incremented
  each 50-ms tick independently, making rounding depend on pulse partitioning. Prediction now
  anchors to the last authoritative amount and calls the shipped `accrueConstant` closed form over
  total elapsed milliseconds; a 50+50 versus 100-ms regression fixture is exact. Second, the
  ten-minute return path depended on the Worker stalling, but browsers may keep a throttled Worker
  alive. Lifecycle now measures hidden duration directly, opens the return recap, and immediately
  requests authority on visibility return.
- Closed the remaining executable-contract surfaces: exact snapshot/receipt adaptation, closed
  six-intent dispatch union with runtime exact-field validation, aggregate-only reconciliation and
  Worker telemetry, server-coordinate progress rendering, generic discrete-state rendering,
  reduced-motion steps, contract/main/run-end routes, RTA start semantics, and the single
  fast-forward/gains/optional-modal return state.
- Focused verification: typecheck and Svelte diagnostics clean; 6,412 Node tests pass; production
  build emits a separate prediction Worker chunk; 19,245 tests pass across Chromium, Firefox, and
  WebKit; schema and source-boundary checks pass.

## 2026-07-29 — canonical docs and full gate

- Added `docs/client-shell.md`, updated the canonical docs index and root status, and made the
  production entry validate the shipped Phase-0 economy catalog before mounting. The docs state
  the transport gap explicitly: the shell waits on a pending `SnapshotStream`; it does not pretend
  that an end-to-end playable client exists.
- `TEST_DATABASE_URL=… make verify` passes: Go vet and all Go packages; formula drift; Commons
  population invariance; deterministic harness baseline; TypeScript and Svelte checks; production
  build with a separate Worker; 6,412 Node tests; schemas and source boundaries; and 19,245 tests
  across Chromium, Firefox, and WebKit.
- Attempted an extra in-app visual smoke pass after the automated browser gate; this session
  exposed no controllable browser. No alternate UI surface was substituted. Independent per-change
  review remains the only process gate before archive.

## 2026-07-29 (claude — independent per-change review of 86fbd44..10c1407: APPROVED, one finding)

Full read: reconciliation, prediction, display, intents, worker protocol, boundary lint, both
test suites run. Faithful to the RFC's executable contracts:

- **Prediction** re-accrues total elapsed from the committed base each step through the *shared*
  `accrueConstant` (no incremental drift, no reimplemented math), saturates at caps client-side
  mirroring `ApplyAccrual`, honors the 100-step ceiling, resets and emits `offline_required`
  beyond the catch-up ceiling with **no local catch-up** — C2 exactly. Stale revisions ignored;
  non-monotonic clocks throw.
- **Reconciliation** computes PPM divergence through Decimal ops with the `max(|auth|,1)`
  denominator, rejected receipts rebase with the typed code as explanation, discrete facts are
  primitives (`boolean|number|string`) so `===` is value equality — the reference-equality trap
  I chased does not exist.
- **Display** implements C4's `activity_ppm` contract (a producing counter cannot render as
  unchanged even when notation digits cannot move — asserted in test), cap `reason_key` exposed
  at cap with **missing reasons rejecting adaptation**, reduced-motion as 500-ms discrete samples
  of identical values.
- **The dispatcher is the closed six** with exact-key validation (`requireExactKeys` — extra
  fields throw), no predicted state accepted, boundary lint enforced. The lint is pattern-based
  rather than import-graph, acceptable **only because the TS kernel exposes no mutation API at
  all** (RFC-0002 K4) — the lint is belt over an absent surface; noted, not a finding.
- Deviations section is honest (structured-clone now; transferable buffers correctly rejected
  as an optimization because canonical strings can't survive binary floats).

**FINDING (medium, blocks nothing today, must resolve at transport time):**
`perform_manual_batch` carries a client-side **`windowMs`** field that exists in **neither** the
shell RFC's own C-contracts **nor** the server's C1 manual-batch envelope (`{action_id, count}`).
The server's exact-schema validation would reject it on the wire; today there is no wire, so no
runtime bug — but the transport RFC's adapter must either strip it, or the production C1 contract
gains it by amendment. Pick one deliberately; do not let the adapter's mapping decide by accident.
**Routed to the Transport RFC as an open item.**

Minor note: the intent-ID regex accepts any UUID version (variant checked, version not); server
enforces v7 — client leniency is harmless but the pattern could pin `7` for free.

Clear to archive.

## 2026-07-29 (Codex — review-finding correction and archive)

- Re-verified the review's sole finding before acting on it. The finding is false: Production C1
  specifies `perform_manual_batch {action_id,count,window_ms}` and says `window_ms` is retained for
  audit/UX but grants no authority (`rfc/archive/production-engine-and-intents.md` C1). Canonical
  `docs/production-engine.md` says the same. The live exact-schema parser requires `window_ms` at
  `server/production/intents.go`, and both unit and real-Postgres integration fixtures send it.
  The shell's `windowMs` is therefore the correct internal spelling for Transport D3 to map to
  `window_ms`; no production amendment or stripping decision exists.
- Removed the false Transport open question rather than letting a disproved diagnosis survive as
  future work. The independent review's approval remains valid. All acceptance gates and canonical
  docs were already green; Client Shell & Sim Loop is archived.
