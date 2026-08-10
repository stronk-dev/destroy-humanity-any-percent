# UI Foundation implementation log

## 2026-08-03 — post-ruling acceptance pass

C1–C8 were owner-ruled, and Copy Pipeline is now implemented. The follow-up found three literal
contracts still absent despite the draft describing them as closed/pinned: no theme token key or
era-value matrix, no notation package version/options/sign behavior/golden strings, and no named
a11y engine/version. C9–C11 record the narrow owner decisions needed; implementation does not
invent those presentation/dependency choices.

UF3 was reconciled directly to the implemented Copy Pipeline: callers pass an era to `t`, which
uses `era_variants[era]` and falls back to base text. The stale `key@era` grammar was removed rather
than creating a second copy system.

## 2026-08-04 — independent specification follow-up (`87f542d..24203ee`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **implementation remains correctly blocked, but the active RFC is not yet singular.**
  C9–C11 are real, explicitly open owner decisions and no UI implementation was improvised. The
  claimed C1–C8 reconciliation is incomplete, however.

**HIGH — stale normative C1–C8 prose still contradicts the ruled contract.** In the active
specification, UF3 repeats the withdrawn `key@era_2006` grammar immediately after saying no second
grammar exists; UF4 still declares `era_skins`/`unlock_condition_ref` after C5 removes `era_skins`
and defines the closed shell-fact condition; UF5 still allows Storybook after C7 rules it out; and
AC4 still requires an internal `$derived` count after C5 replaces that with observable subscription
counts. Reconcile those decision sites before ruling C9–C11 so resolving the three literal blockers
does not expose a second contradictory implementation path.

What held: the status, plan, and log correctly block all implementation on the missing theme token
matrix/era values, notation package version/options/golden strings, and accessibility engine/version.
No dependency or presentation choice was invented in this range.

## 2026-08-04 — ruled-text reconciliation

Removed the four stale normative paths identified by independent review: the withdrawn
`key@era_2006` grammar, `era_skins`/untyped unlock references, the Storybook option, and the
unobservable `$derived` acceptance assertion. UF3–UF5 and AC4 now state only the C1–C8 rulings.
C9–C11 remain the sole implementation blockers; no presentation literal or dependency choice was
invented while reconciling existing decisions.

## 2026-08-04 — independent ruled-text reconciliation review (`402ba20..85cbea6`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **approved as a specification-only reconciliation; implementation remains correctly
  blocked on C9–C11.**

UF3 now has only the Copy Pipeline's typed `t(key, params, era)` plus `era_variants` fallback; the
withdrawn `key@era_2006` grammar is absent from the active decision site. UF4 owns the ruled closed
surface/unlock row, byte ordering, one observable subscription handle, and token-only styling with
no `era_skins` implementation path. UF5 requires the repository-owned Vite fixture in the existing
browser suite and expressly forbids Storybook. AC4 asserts observable subscribe/unsubscribe and
inactive-callback behavior rather than compiler-internal `$derived` counts.

C9's literal token matrix/era values, C10's notation package pin/options/golden strings, and C11's
accessibility engine/version remain explicit unchecked plan items and the RFC status still blocks
implementation. No presentation constant, dependency, or playable screen was invented in this
range. Exact-range diff checking and the fresh root `make verify` both pass.

## 2026-08-07 — owner rulings C9-C11 (the last implementation blockers)
C9: the closed token key-set (color/type/space/border/chrome/motion, 41 keys) + literal era_1995 and
era_2000 matrices (values provisional/tunable; the KEY SET is the contract; the 1995 motion budget is
literally zero). C10: @antimatter-dimensions/notations Standard, 3 sig digits, plain ints <1000, no
scientific threshold in Phase A, negatives valid; exact version recorded in changelog+lockfile at the
implementing commit. C11: axe-core (same version-recording duty), WCAG 2.2 AA, zero serious|critical;
keyboard/focus/name/motion assertions stay separate. Status -> accepted; implementing. UI Foundation
is now the playability critical path together with the Minigame API & Surface amendment.

## 2026-08-10 — implementation handoff preparation

The C9–C11 literals now have executable owners: two strict 41-token theme artifacts and loader,
Standard notation 1.6.0 with canonical Decimal goldens and the shared 100-ms Amount scheduler, and
axe-core 4.12.1 in the existing three-browser fixture route. The surface registry rejects unknown
shell facts and proves unsubscribe-before-mount plus zero inactive callbacks. `App.svelte` is now a
content-free primitive fixture; it invents no screen or authoritative game value.

Focused typecheck, client tests, schema/boundary gates, and the full browser matrix are green. This
entry is a self-check record only: independent designated review and archival remain open.

## 2026-08-10 — designated cross-party verdict: UI Foundation implementation {3483ab1} — NOT APPROVED (narrow)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

**Verified excellent across C1-C11:** byte-exact 41-key token matrix both eras (era_1995 motion
all-zero as ruled); tokens as schema-validated data with a strict 41-property runtime loader;
DOM-stable era switch; single formatter boundary with notations 1.6.0 pinned same-commit (and the
nested break_infinity 1.3.0 correctly aliased out — law 3 intact); zero component literals (both
seeded-violation probes red); the C11 axe gate LIVE in make verify across Chromium+Firefox+WebKit
and probe-red on a seeded contrast violation; closed surface registry with exact-key/duplicate/
unknown-fact rejection and proven lifecycle ordering; strict TS; honest no-bump kernel handling.

**BLOCKING (F1, HIGH): the C10 ruled literal is not what ships.** Ruled: "3 significant mantissa
digits (1.23 Qa)". Shipped: fixed 2 DECIMAL PLACES (probe: 1.2345e4 → "12.34 K" = 4 sig digits;
1.2345e5 → "123.45 K" = 5). The goldens dodge the exposing range and docs/ui-foundation.md
restates the unshipped behavior as canonical fact. **OWNER-SIDE RULING (recorded now): the C10
literal STANDS — fix the formatter to true 3-significant-digit mantissas** (width-stable under
Standard notation: 1.23 / 12.3 / 123), add the mantissa≥10 goldens AND the missing C3
sub-visible-collapse golden (F3), and the docs become true as written. No re-ruling.

**F2 (CRITICAL, OUT-OF-RANGE — routed to the First Content mint thread): the Go suite is RED
under cold caches since the MINT COMMIT 3ff34bf** — the mint added fiscal.generator.beige_tower/
fiscal.hoard to the live economy artifact while server/fiscal/catalog_test.go:55-60 appends the
same IDs → duplicate-id rejection in ./economy, ./epochseed, ./fiscal with -count=1. Warm Go
caches masked this through the mint review, the B1/B4/B5 review, and every green claim since.
Fix: the test must use non-colliding IDs (test-file fix); the mint thread gets a record note
(the mint verdict's gate claim was cache-masked — the mint BYTES are unaffected).

**F4-F6 (non-blocking, fix with F1 or note):** keyboard activation never dispatched in the C7
test + visible focus unasserted; lint gaps (gradient literals, fetch/WebSocket, non-recursive
readdir); capped Amount aria-label replaces the numeric value in the accessible name.

**Range-union:** consumes exactly {3483ab1}. NOTE closing the reviewer's item 7: 7b48a9d IS
covered — the same-day B1 verdict (planning/first-content-epoch/log.md) explicitly consumed
{7b48a9d, 5c20ee3}; no thread gap exists.

**Verdict: NOT APPROVED pending F1 (+F2 repo-wide, owned by the mint thread; +F3/F4 bundled).
Re-review is a narrow delta.**

## 2026-08-10 — F1/F3-F6 remediation handoff

Commit `8873301` implements the designated verdict's narrow UI delta:

- Amount precision is selected from the three-significant-digit rounded engineering group, pinning
  `1.23 K`, `12.3 K`, and `123 K`; the golden set now includes the two wider mantissas, rounding
  carry, and a sub-visible `4.9e-1 -> 0` case.
- the browser fixture performs a real Enter-key activation, observes the era transition, and
  requires a non-none focus outline;
- the boundary walker is recursive and rejects seeded gradient literals plus raw `fetch` and
  `WebSocket` use;
- a capped Amount's accessible name retains the rendered numeric value before its resolved reason.

Normal repository gates pass: `make verify-client` (0 TypeScript/Svelte diagnostics; 6,626 unit
assertions) and `make test-browser` (19,890 assertions across Chromium, Firefox, and WebKit).
The separate mint-thread remediation also passes the complete uncached Go suite through
`make test-go GO_TEST_FLAGS='-count=1'`. This is an implementation record, not an approval;
`8873301` is ready for the narrow designated cross-party re-review. UI Foundation remains
unarchived pending that verdict.

## 2026-08-10 — designated cross-party re-review: F1/F3–F6 remediation {8873301, 5a7dc01} — NOT APPROVED (one residual)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
F1 core + F3–F6 all CLOSED with probes (ruled probe values render correctly; reverting the
formatter fails exactly the two new goldens; real keyboard activation + focus assertion; all
three lint gaps seeded-red; capped accessible name keeps the number). Gates green (19,890
browser assertions); kernel honest at 0.3.88.
**F1-R (blocking under the standing C10 ruling): the group-carry window still ships 4–5 sig
digits** — `groupDigits` derives from the 3-sig-quantized exponent but `standard.format` receives
the UNROUNDED magnitude, so in ≈[999.5, 999.995) of every engineering group the render is e.g.
"999.60 K" where the ruled literal demands "1.00 M"; every growing counter crosses the window in
every group, and docs/ui-foundation.md:26 is false inside it — the same truth-as-written test the
original F1 applied. Fix: format the ALREADY-QUANTIZED value so group selection and formatted
value agree; add carry-boundary goldens; collapse the "−0" sign nit while in the file.
**Re-review is a one-file delta.**
