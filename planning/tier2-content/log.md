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

## 2026-09-25 — C4c: presentation candidate for the Tier 2 bindings (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review.

`balance/testdata/t2/presentation-candidate-v3.json` is generated by the same insertion-only tool.
It carries presentation v3 plus the three T2 generator bindings, the three T2 upgrade bindings and
`gate.t1_to_t2.title`. It also sets `generator.garage_rack`'s `cap_reason_key` to the new
provisioned-hardcap reason key. Without this candidate, the Desk's `requirePresentation` would
throw at the first Tier-2 render or split after mint. The 14 new keys are candidate text in
`copy/catalog/tier2-candidate.json`, for owner adoption (E4 owner copy round). Voice: deadpan, and
the target is the office *system*, never employees.

- **Test:** `tier2-candidates.test.ts` requires the candidate's generator and upgrade ids to equal
  the candidate economy's ids, and its gate bindings to cover every gate the preview can project
  (`gate.t0_to_t1`, `gate.t1_to_t2`). Every title, description and cap key must be declared copy,
  each binding's `cap_reason_key` must equal the economy's `provisioned_hardcap.reason_key`, and
  rows must be byte-sorted.
- **Severing:** P1 (a missing generator row), P2 (a null garage cap reason) and P3 (an undeclared
  title key) each fail the test. All were restored, and `--check` is clean.
- **Mint note (M1):** at mint, `client/src/game-ui/presentation.generated.json` and the
  game-UI-copy candidate lane consume this file. That wiring belongs to the owner-gated M lane.

## 2026-09-25 — C5: Tier 2 pacing measurement and gate-literal calibration (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review.

- **Runner (`server/harness/first_hour_runner.go`, unguarded).** `runWithCareer` now delegates to
  `runWithModes`, which takes an optional Tier 2 runtime. Without it the behavior is unchanged.
  - **Evidence:** `make first-hour-harness` with the ratified epoch-8 tuple (200/2e0/50/1e4/10)
    regenerates `first-hour-epoch8-report.v1.json` semantically identically, except for the
    Reputation lane's already-landed `reputation_exits` field (a JSON compare that strips that
    field is equal). The full default `./harness` package passes cold (42.6 s).
- **`server/harness/tier2_pacing.go` (§P1/§P2 subset).** It reuses the ratified Chaos and Casual
  T0–T1 literals.
  - The exit rule is `t01_c32_readiness_once`: the elective Exit is taken only while the Founder
    Exit history holds exactly one entry.
  - `gate.t1_to_t2` joins the crossing set.
  - `milestone.it_company_gate` records Founder-attended time at the first crossing in any run.
  - An unreached seed is a visible `must_reach` failure; any other runner failure is an invalid
    measurement.
  - Offers are never answered, since the simulated transitions spawn none.
  - `reference.greedy` v2 and the allocation arms are refused while OD-1 is held.
- **Witness.** `TestTier2PacingReachesTheITCompanyGate`: Chaos seed 0 crosses in run 3 with
  exactly one elective Exit. A bundle without the gate is refused, the held policy is refused, and
  a 1 h horizon yields exactly `must_reach:milestone.it_company_gate`.
  **Severing:** without the exit-once clause, seed 0 Exits 11 times at the Garage gate and never
  reaches Tier 2, which confirms the RFC's §P1 claim.

### Calibration (`TestTier2PacingCalibration`, report `balance/testdata/t2/pacing-calibration-v1.json`)

Chaos ran 64 seeds with a 6 h horizon and Casual 32 seeds with a 12 h wall horizon. Each p50 is the
Founder-attended time at `gate.t1_to_t2`; unreached seeds sort as +inf. The envelope is §P2/OD-8,
[2 h, 3 h] for **both** p50s:

| gate literal | Chaos reached | Chaos p50 | Casual reached | Casual p50 |
|---|---|---|---|---|
| 3e6 | 64/64 | 0.62 h | 32/32 | 0.50 h |
| 1e7 (RFC provisional) | 64/64 | 1.40 h | 32/32 | 0.72 h |
| 3e7 | 64/64 | 1.74 h | 32/32 | 1.28 h |
| 1e8 | **0/64** | unreached | 32/32 | 1.51 h |
| 2e8 | 0/64 | unreached | 32/32 | 1.56 h |
| 3e8 | 0/64 | unreached | 32/32 | 1.59 h |
| 5e8 | 0/64 | unreached | 32/32 | 1.78 h |

**Result: no literal in [1e7, 1e9) meets the OD-8 envelope for both personas.**

- **Casual.** Casual is below 2 h at every literal up to 5e8. It earns most of its progress
  offline (90% rate), and offline time is not Founder-attended time, so its attended clock is
  short. The gate cannot exceed `1e9`: that is the ratified `gate.t2_to_t3` literal, and the §B0
  price band caps Tier 2 below it.
- **Chaos.** Chaos crosses by 1.74 h at 3e7 and not at all within 6 h from 1e8. Its uniform spending
  over every affordable command never banks 1e8 of cash. A follow-up probe at 5e7/7e7 locates the
  cliff (appended below).
- **Owner decisions this blocks (OD-7/OD-8), for the ruling author.** The RFC's defaults cannot
  both hold. Options:
  - (a) Read the envelope on a different clock or statistic, for example Casual on wall time, or a
    single persona.
  - (b) Retune content so Casual slows and Chaos can bank, for example Tier-1 price or yield changes,
    which are epoch-8 content edits outside this RFC.
  - (c) Accept a literal outside the envelope with a recorded deviation.

  The implementer does not pick; the gate literal stays at the provisional `1e7` in the candidate.

### Not done in this session
- **C6, §P3 relevance.** The epoch-8 registered scenario is the T1 scenario that §P3 retires. The
  combined `scenario.t1_t2_relevance` needs the six T2 rows authored into a candidate relevance
  policy, plus solver budgets from measurement (T01-C17). It is not started, and it is the next
  item for this lane.
- The Headcount panel, §H, P4 seats and P5 are held on OD-1.

### 2026-09-25 — C5 addendum: cliff probe and reproducibility (Claude)

- **Cliff probe (`T2_GATE_AMOUNTS="5e7 7e7"`, not committed to the report):**
  - at 5e7, Chaos reaches the gate in 62/64 seeds (p50 1.90 h) and Casual in 32/32 (p50 1.30 h);
  - at 7e7, Chaos reaches it in only **6/64** seeds (p50 unreached), while Casual reaches 32/32
    (p50 1.31 h).

  The Chaos cliff therefore lies between 5e7 and 7e7, and Casual stays below 2 h throughout. The
  probe run failed the report-drift check, as designed, since its sweep differs from the committed
  one; that is the drift gate's demonstrated failing case.
- **Reproducibility:** a cold, non-update `TestTier2PacingCalibration` run over the committed
  default sweep re-derives `pacing-calibration-v1.json` byte-identically (`ok`, 1185.8 s).

## 2026-09-25 — C6: combined T1–T2 relevance on the candidate bundle (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review. Scope is
harness only; no kernel-guarded path is touched (Cosmetic Shop owns those).

- **Inputs, generated insertion-only by the candidate tool:**
  - `balance/testdata/t2/relevance-candidate-v1.json` is the epoch-8 policy plus the six Tier-2
    rows, each windowed `gate.t1_to_t2 → gate.t2_to_t3` at `epsilon_ms` 1000 with no trap
    exemption.
  - `relevance-scenario-t1-t2-v1.json` is §P3's `scenario.t1_t2_relevance`: a 1e9 cash target, the
    three segments `[null→t0_to_t1]`, `[t0_to_t1→t1_to_t2]` and `[t1_to_t2→t2_to_t3]`, and the
    epoch-8 T1 scenario's personas and solver parameters.
- **Test:** `TestTier2RelevanceCandidateAddsExactlyTheSixTierTwoRows` loads the suite and requires
  the policy to equal epoch 8 plus exactly those six rows. Severing it by dropping `upgrade.nap_pod`
  fails the test.
- **Budgets, T01-C17 from measurement.** The first run used a provisional ceiling of 1024 runs and
  2e7 transitions and measured 167 runs and 4,590,515 transitions. The committed budgets are
  **400 runs and 1e7 transitions** (≥ 2× measured). Severing: `max_runs` 166 fails loud at
  preflight ("relevance run budget exceeds scenario limit").
- **Reproduction at the committed budgets:**
  - Every item result, oracle and failure is identical to the ceiling run.
  - `executed_transitions` differs: 4,715,708 against 4,590,515. That is still under the budget,
    at 2.12×. The cause is not isolated; it is recorded rather than smoothed over.
  - The diagnostic is committed as non-authoritative evidence
    (`relevance-t1-t2-diagnostic-v1.json`); the `balance-harness` relevance mode refuses to write
    an authoritative report while failures exist.

### Result: the gate FAILS. Tier-2 content is weak or dominated at 1e9.

Reference trajectory to 1e9 cash. "Bought" is the reference purchase count; "individual" is the
change in time-to-1e9 when that row is removed.

| Row | Bought | Individual | Verdict |
|---|---|---|---|
| `generator.open_plan_floor` | 30 | **0 ms** | **dominated**: removing it does not slow the target (instrument-affected, target excluded) |
| `generator.managed_services_contract` | 2 | **0 ms** | **dominated** (instrument-affected) |
| `generator.hot_desk_program` | 1 | +1,800,000 ms | relevant, bought once |
| `upgrade.ping_pong_table` | 0 | unreached | **never bought** (effect target excluded) |
| `upgrade.move_fast_break_things` | 0 | unreached | **never bought** |
| `upgrade.nap_pod` | 0 | unreached | **never bought** |

**Knock-on effects on Tier 1, compared with the epoch-8 golden:**
- `generator.first_hire` falls from 10 to 2 purchases. `upgrade.employee_handbook_v0` falls from
  1 to 0 and fails its floor as instrument-affected, because its target `first_hire` is excluded.
  Tier-2 content therefore crowds out part of Tier 1.
- `upgrade.crt_degauss_button` now passes (it failed in the epoch-8 golden).
- `generator.beige_tower_v2` and `upgrade.refurbished_sticker` still fail. Both failures predate
  this work: they are in the epoch-8 golden and covered by its branch report.

**Instrument exclusions:** `answering_machine`, `first_hire`, `hot_desk_program`,
`managed_services_contract` and `open_plan_floor`. The greedy oracle is null. The deviation oracle
**passed** (8/8 probes reached, none starved).

**Reading.** At the provisional prices (§B1/§B2) and a 1e9 target, the Tier-2 generators mostly
substitute for Tier-1 production instead of adding a new axis. The upgrades are never bought
before 1e9 is banked: their costs of 3e7–3.6e8 compete with generator ladders that pay back
faster. The RFC expects headcount (§H) to be Tier 2's growth engine, and headcount is held on
OD-1. This measurement is therefore on content without its defining mechanic, which is itself a
finding for the owner.

**Decisions for the owner or ruling author (the implementer does not retune):**
- retune the §B1/§B2 prices and yields, a balance change requiring owner SHA ratification;
- let headcount (OD-1) land first and re-measure;
- record trap exemptions with justification keys, a policy change.

No production artifact changed, and the gate literal stays `1e7`.

## 2026-10-08 — Designated review of original C1–C4, bounded approval

**Review by:** Codex (designated cross-party reviewer of Claude's implementation).
**Recorded by:** Codex. **Verdict:** APPROVED for the bounded changes below, not whole-feature
acceptance. Exact reviewed ranges:

- `6bc4c43b^..6bc4c43b` — candidate derivation, identity, loader/refusal and ApplyLogged
  reachability (C1–C3), associated tests, Make target and records.
- `3623b260^..3623b260` — era tokens, closed era/schema/copy vocabulary, tier mapping,
  mounted theme checks and associated docs/records (C4a).
- `a513d3d6^..a513d3d6` — adjacent-gate projection, optional incorporation DTO/decoder/control,
  inert energy chrome, generated/candidate copy and associated tests/docs/records (C4b).
- `2215fc2a^..2215fc2a` — complete candidate presentation bindings, derivation and copy
  references with their tests/generated outputs/records (C4c).

These four commit diffs cover the original C1–C4 changes. Intervening Pet records, later
harness C5/C6 commits, subsequent Codex UI corrections and their own independent-review debt
are not approved by this verdict. Execution is on clean current source `f3a859be`, not a claim
to have checked out and rerun each historical revision. The original candidate artifacts,
reachability test, theme and Tier-2 preview test remain byte-identical; the generator's later
relevance extension is outside the reviewed C1–C4 scope.

Executed evidence:

- `make t2-candidates-check`5126 PASS: deterministic current derivation, three cold Go
  reachability/insertion tests and three TS candidate tests. The epoch-8 control actually
  rejects the new gate/incorporation; mismatched categories reject under a correct bundle hash.
- Cold complete `./gameui`89237 PASS. Inspection binds the preview to production's discarded
  clone, the cash boundary to actual transition outcomes, and incorporation to the pinned
  faction catalog without changing the production kernel or below-Tier-2 wire fields.
- Declared Linux Playwright service89815, cold frozen install, complete `tier2-candidates`,
  `game-ui`, `game-ui-screens-browser` and `ui/themes` files:615 PASS across all three engines;
  six explicitly separate performance-only skips. Initial focused9412:42 PASS/579 deselected,
  superseded for coverage by the complete file population. These exercise native transition
  input/focus consumers as well as the original theme, incorporation and inert-refill checks.
- `make typecheck api-check copy-check verify-schema vet GO_PACKAGES='./gameui'`78647 PASS:
  zero type/Svelte errors or warnings, byte-unchanged API generation, content/copy consistency,
  all three theme schemas and Go vet. Copy-check retains621 orphan warnings; this does not
  adopt candidate copy or establish a zero-orphan release. No hosted-CI claim.

Important acceptance boundary: the mounted Tier-2 fixtures retain the older generator/gate
content and change the tier/control fields. The current generated presentation is still the
older content presentation; C4c provides candidate bindings, not their release activation.
These tests therefore prove bounded era/control behavior, not actual Tier-2 catalog rendering,
browser→API→Postgres crossing/incorporation, Headcount or AC8/AC10. The Go reachability case
is in-memory ApplyLogged, not persistence. Full integration remains required after contract/
content preparation, not waived by this approval. C5/C6 review, RP-413/RP-414, pacing/relevance,
copy/SHA adoption, mint and the full nine-tier goal remain open. No product edit, acceptance
checkbox, lifecycle/archive, deployment or push is performed in this review.

## 2026-10-08 — C5/C6 review checkpoint before whole-product status handoff

At source `1bc8ce93`, cold `make test-go GO_PACKAGES='./harness'
GO_TEST_FLAGS='-count=1 -run TestTier2 -v'`83704 exits0: the actual pacing reach/refusal/
unreached test and the six-row candidate test pass. `TestTier2PacingCalibration` explicitly
SKIPS because its exhaustive opt-in is unset. No calibration or combined relevance experiment
was rerun, so this is not independent reproduction of either recorded measurement and gives
no C5/C6 designated approval. The candidate test's six added-row checks are not an assertion
that all inherited policy fields are unchanged. Historical finite gate samples likewise must
not be read as independently established proof about every literal in a continuous interval.

Review by: Codex (partial cross-party inspection, no acceptance verdict). Recorded by: Codex.
C1–C4's preceding bounded approval stands; C5/C6 review and full Tier-2 acceptance remain open.
Next in this lane: complete the measurement/policy review and resolve the recorded pacing,
relevance and Headcount boundaries before mint/integration acceptance. No product changes,
report regeneration, checkbox promotion, archive or push in this checkpoint.

## 2026-10-08 — RP-432: make the C6 policy-preservation assertion real

Outcome under §P3: the candidate must preserve inherited policy rows while adding only the
six specified Tier-2 rows. The former test skipped inherited rows entirely. Extracting its
unchanged checks and changing an inherited epsilon produces the expected red regression:
focused37771 exits2, `accepted a changed inherited epsilon`; the unmodified candidate passes.

The test-only correction uses the existing complete policy type, compares inherited fields,
schema and groups, rejects duplicate IDs and retains the exact six new IDs/window/threshold/
no-exemption restrictions. Eleven altered-policy cases exercise those checks. Focused58788
passes; cold complete default `make test-go vet GO_PACKAGES='./harness'
GO_TEST_FLAGS='-count=1'`48178 passes (85.781s plus vet). Exhaustive opt-in measurements are
not part of that default population. No production code, candidate balance or recorded report
changed. `git diff --check` passes.

Review by: Codex (implementer first filter); Recorded by: Codex. The test correction after
`0512405b` through this commit needs designated cross-party review; it cannot approve its own
C6 range or the measurement. Next: non-update exhaustive C5 pacing reproduction, then complete
the C5/C6 review with the recorded instrument and Headcount limitations intact. No acceptance,
archival, content mint or push.

## 2026-10-08 — C6 reproduction and RP-433 T0 population conflict

Review by: Codex (cross-party measurement/contract inspection, not whole C6 acceptance).
Recorded by: Codex. Source `e6a64aff`; the following runs change no production/balance inputs.

- Combined C6: `make t0-t1-relevance
  RELEVANCE_SCENARIO=balance/testdata/t2/relevance-scenario-t1-t2-v1.json
  RELEVANCE_OUTPUT=.cache/tier2-review-e6a64aff.json`53242 exits2 as required by the red gate.
  All167 runs/4,715,708 transitions complete. The diagnostic is byte-identical to the committed
  one (`cmp` exit0; both SHA256 `5896a7005b9541bb741edf56cacd42d4fa61b10d4b8666636b216543b612af1e`).
  All six T2 rows, including the positive hot-desk row, are instrument-affected. The old
  “weak or dominated” headline is not a clean content verdict; neither zero deltas nor missing
  purchases after those exclusions authorize retuning. The historical record is retained;
  current docs now state this limitation explicitly.
- P3 T0 comparison: run the existing `balance/testdata/t0-t1/relevance-scenario-v2.json` control;
  for the candidate, copy it exactly and replace only `catalog` and `routes_catalog` with the
  two T2 candidate paths. Retain the original T0 policy and every other value. The original
  economy/routes files are byte-identical to epoch8's active files (`git diff --no-index` exit0).
  Both use `make t0-t1-relevance` with those scenario paths and separate ignored output paths.
  Initial setup wrongly supplied the complete T1–T2 policy; loader refusal happens before
  measurement and is excluded. The corrected scenario differs only in those two paths,
  checked with Node deep equality. Controls70478 and candidate92020 both finish107 runs and
  correctly exit2 with non-authoritative diagnostics, not an authoritative release golden.
- Removing only `scenario_hash`, `constants_hash` and `relevance_policy_hash` leaves thirteen
  changed report leaves: work180,936→175,182, dot-matrix purchases46→41, nephew5→1, nephew
  delta0→5,400,000ms/support/floor changes, two role counts, and six→five findings. Thus even
  the fresh-control prerequisite for §P3 invariance fails; this does not claim reproduction
  of a historical T0 golden or a player-visible gameplay regression.

Local artifact SHA256s (raw files remain ignored; do not require them for ordinary CI):

- Scenario `.cache/tier2-review-t0-scenario-e6a64aff.json`:
  `c069c0df2295c042f8ad39ad88a70e756ba59ba50c3488861d9f1acdce4d2ccb`.
- Control `.cache/tier2-review-t0-control-e6a64aff.diagnostic.json`:
  `5ee5ae2cad87de352e7d2e8906d915bd46815c57ce9872ade7362fdb90856dcb`.
- Candidate `.cache/tier2-review-t0-candidate-e6a64aff.diagnostic.json`:
  `49827306c4f3cf553534de0e01b8b50b60d44d7a3c59c9801659dfa86bc0bf9e`.

**DESIGN-GAP / RP-433:** archived Balance Harness Foundation requires Chaos v1's uniform
syntactically-valid provider with one template per catalog generator and forbids changing its
behavior under v1. `Suite.intentBytes` still implements it: nine versus twelve generators
changes the draw universe. §P3's unaffordability argument is insufficient to imply report
identity. The changed scenario/constants hashes also enter deviation-probe selection, so the
work-count difference alone is not a content signal. Proposed explicit amendment: a new
versioned controlled comparison population, preserving v1 and executing actual candidate
production rules; reconcile §P3's population/comparison contract before implementation.
This is a proposal, not authority to filter v1, loosen assertions, regenerate goldens or mint.

The original seven-literal C5 reproduction and an existing candidate-owned branch diagnostic
remain live on their original handles; no result for either is claimed here. No software suite
rerun for this evidence/docs-only checkpoint, acceptance checkbox, lifecycle/archive or push.
