# Tier 2 Content — implementation log

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review; never
self-approved.

Scope for this session follows `plan.md` C1–C6. Batch order: C1+C2+C3 (candidates, loaders,
reachability), then C4 (UI), then C5/C6 (harness). Kernel protocol: any commit touching a
`kernel/affecting-paths.json` path bumps `kernel/VERSION` (and the Go/TS constants) in the same
commit. Candidate files under `balance/testdata/` are not governed epoch artifacts, so they are not
`BALANCE-CHANGE:` commits. No production artifact, epoch seed or golden changes in this session.

**DESIGN-GAP T2-DG-1 (seat source held):** §B1's generator rows carry a `headcount_seats` role whose
meaning depends on the unruled OD-1. The candidate rows omit that role and keep the rest of the
row literal. When OD-1 is ruled, the role is appended to `generator.open_plan_floor` and
`generator.hot_desk_program` exactly as §B1 writes it, together with §H.

## 2026-09-25 — C1–C3: candidates, loaders, reachability (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review.

- `client/tools/generate-t2-candidates.mjs` derives `balance/testdata/t2/{economy,routes,categories}-candidate-v1.json`
  from the epoch-8 bytes by **insertion only**: §A1 gate row, §A2 gate-set entry, §B1 rows (minus
  `headcount_seats`, T2-DG-1) and the garage-rack `provisioned_hardcap`, §B2 upgrades, §B3 pool and
  its `multiplier_sources` row. `candidates.sha256` pins the file bytes;
  `candidate-bundle-hash.txt` pins the constants identity of the epoch-8 bundle with the three
  candidates swapped in (`sha256:cc2d122b…84a5`). The economy stays schema v4: v5 is headcount-only
  (§H1), held on OD-1.
- **AC0 (Go, `server/production/tier2_candidate_test.go`):** on the candidate bundle, a run-2, Tier-1,
  v18 Company with a foundation-complete Founder carry crosses `gate.t1_to_t2` through `ApplyLogged`,
  gets `Tier == 2`, and `incorporate open_source` applies. Failing case: the same commands on the
  epoch-8 bundle reject with exact receipts `unknown_id/gate.t1_to_t2` and `not_eligible/tier`.
- **Insertion-only witness:** every epoch-8 economy/routes line survives in order (a line may gain
  only the trailing comma a following insertion needs); categories differ only by the gate-set
  insertion. The candidate categories load against candidate route gates; the epoch-8 categories
  are rejected against them (AC1 subset).
- **TS parity (`client/test/tier2-candidates.test.ts`):** `loadReplayCatalogBundle` accepts the same
  artifact set under the Go-pinned hash; with the epoch-8 categories under their own correct hash the
  loader rejects, so the rejection is the gate-set check and not the label.
- **Severing probes:** S1, removing the gate row from the routes candidate, fails both Go tests. S2,
  dropping the gate from the categories candidate, fails the Go insertion witness and both TS tests.
  S3, editing a live economy line, fails the insertion witness. A wrong pinned bundle hash fails the
  reachability test. All were restored, and `--check` is clean.
- **Observation:** the replay bundle does not load categories (identity only), so category validity
  is witnessed through `leaderboard.LoadCategoryCatalog` in Go and through `loadReplayCatalogBundle`
  in TS.
- **Gate:** `make t2-candidates-check` (new) runs the drift check plus both runtimes. Test files
  under `server/production/` are kernel-guard-exempt (`_test.go`), so no version bump was needed.
- **Noted, not mine:** `gofmt -l server` lists four pre-existing files
  (`copykeys/generated.go`, `deploymentrehearsal/host_test.go`, `minigame/session.go`,
  `replaycatalog/catalog_test.go`).

## 2026-09-25 — C4a: `era_2010` theme and tier-2 era mapping (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review.

- `ui/themes/era_2010.json`: a flat 2010 startup theme (white/cool-grey surfaces, flat blue accent,
  radius 6px, 120/200/320 ms eased motion under `budget: respect`). It stays inside the closed UI
  Foundation token domains: the font domain allows only the three shipped stacks, so the theme
  uses Verdana rather than extending the domain. The values are **candidate design data** for owner
  ratification (§M3).
- The era joins `UI_ERAS`, `CopyEra`, the copy pipeline's era set, the theme schema enum and the
  schema verifier (exactly three themes, in file order). `eraForSnapshot` and `RunEndSurface` map
  tier 2 to `era_2010`; tier ≥ 3 still throws. No `era_2010` copy variants are added: E4 copy is
  owner-authored, and variantless keys resolve to their base text.
- **Evidence, cold:**
  - `game-ui.test.ts` covers tier 1 → `era_2000`, tier 2 → `era_2010`, and tier 3 throwing.
  - A new three-browser test renders the tier-2 Desk and Run End under the WCAG 2.2 AA axe gate and
    checks the installed accent token.
  - svelte-check, the 6760 unit tests, `verify-schema`, `copy-check` and `verify-client-boundary`
    all pass.
- **Severing:** removing the tier-2 mapping makes the browser test fail with the AC8 `RangeError`
  and fails `game-ui.test.ts`. Setting `color.text` to `#dddddd` fails axe on `color-contrast`.
  Both were restored.
- `client/src/copy/index.ts`, `client/src/ui/themes.ts` and `client/src/game-ui/*` are not
  kernel-guarded, so no version bump was needed.

## 2026-09-25 — C4b: Tier-2 Gate and Incorporate controls, energy-bar stub (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review.

- **Server (`server/gameui`, unguarded).** The Gate preview now projects the next adjacent
  standard gate: `gate.t0_to_t1` at Tier 0 and `gate.t1_to_t2` at Tier 1, but only when the pinned
  routes declare it, so the epoch-8 output is unchanged. `transitions.incorporate` (the sorted
  faction ids plus their existing `incorporation_copy_key`s) is emitted only when `incorporate` can
  apply (Tier ≥ 2, no faction) and is `omitempty` otherwise. In the API schema it is an additive,
  optional `GameUITransitions.incorporate`: `make api-generate` changed `api.json`/`types.ts`, and
  `api-compat-v1.json` did not change (no re-pin).
- **Client.** The contract accepts the optional control only at Tier ≥ 2 with sorted rows,
  `copy_key == incorporate.<id>` and exact keys. The Desk renders one button per faction and
  submits `incorporate {faction_id}`. The era_2010 FarmVille energy bar is presentation only: its
  curtain is in the tooltip and in small print, and Refill emits no intent. Candidate copy for
  owner adoption is in `copy/catalog/tier2-candidate.json` (6 keys).
- **Preview limitation, recorded (AR-F3 family):** at Tier 1, a run-1 Company whose scripted first
  failure is due routes every command, `cross_gate` included, into that Exit. The preview models
  only the ordinary transition, and the command's receipt/event stays authoritative.
- **Evidence (cold):**
  - `gameui` Tier-2 tests on the candidate catalogs: the gate is ineligible at 9.99e6 and eligible
    at 1e7, agreeing with `production.TransitionWithRoutes`, and is absent once crossed.
    Incorporate is offered at tier 2 and 3 with no faction, and not at tier 1 or with a faction.
    The wire is omitted when nil.
  - `gameui`/`account` unit tests pass; Postgres integration for `gameui`, `gameserver` and
    `production` passes.
  - Client: `game-ui.test.ts` (17) passes, covering the six contract accept/reject rows. A
    three-browser test checks the incorporate submit, energy-bar inertness (zero intents on
    Refill) and axe.
  - `typecheck`, `test-client` (6761), `test-browser` (20439), `verify-client-boundary`,
    `copy-check` and `test-game-ui-composed` (v4 lifecycle plus the Pitch phase) all pass.
- **Severing, all restored:**
  - G1, dropping the tier-1 gate map entry, fails the gate test.
  - G2, ignoring the existing faction, fails the incorporate test.
  - G3, removing `omitempty`, fails the wire test.
  - U1, binding a fixed faction to every button, fails the browser test.
  - U2, making Refill submit an intent, fails the browser test.
