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
