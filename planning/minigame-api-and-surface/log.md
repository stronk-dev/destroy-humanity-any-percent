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
