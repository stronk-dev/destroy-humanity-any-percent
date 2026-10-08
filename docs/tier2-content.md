# Tier 2 content (IT Company) — candidate state

`rfc/tier2-content.md` is accepted and implementing, **fixture-first**. Nothing described here is
minted yet. The live epoch (8) has no Tier 2. Epoch 9 "IT Company" is owner-gated (§M), and the
headcount mechanic (§H) is held until the seat-source decision (OD-1) is ruled.

## Candidate artifacts

`client/tools/generate-t2-candidates.mjs` derives the candidates from the epoch-8 bytes by
insertion only. `--check` fails on drift, and `make t2-candidates-check` runs it together with the Go
and TypeScript witnesses.

| Candidate (`balance/testdata/t2/`) | Adds to epoch 8 |
|---|---|
| `routes-candidate-v1.json` | `gate.t1_to_t2` at `1e7` cash (provisional, measured in §P2) |
| `categories-candidate-v1.json` | `gate.t1_to_t2` in `full_gate_set` |
| `economy-candidate-v1.json` (schema v4) | `generator.open_plan_floor`, `generator.managed_services_contract` (provisions `generator.garage_rack`), `generator.hot_desk_program`; `upgrade.ping_pong_table`, `upgrade.move_fast_break_things`, `upgrade.nap_pod`; `pool.org_chart` and its multiplier source; a behavior-neutral `provisioned_hardcap` on `generator.garage_rack` |
| `presentation-candidate-v3.json` | copy bindings for every candidate generator, upgrade and previewable gate |

`candidates.sha256` pins the candidate file bytes. `candidate-bundle-hash.txt` pins the constants
identity of the epoch-8 bundle with the economy, routes and categories candidates swapped in. Go
(`server/production/tier2_candidate_test.go`) and TypeScript (`client/test/tier2-candidates.test.ts`)
load that bundle under the same identity.

Without the candidate gate, crossing `gate.t2_to_t3` takes a Tier-1 company straight to Tier 3.
On the candidate bundle, crossing `gate.t1_to_t2` gives `Tier == 2`, and `incorporate` then
applies. The same commands on the epoch-8 bundle reject with `unknown_id/gate.t1_to_t2` and
`not_eligible/tier`.

## Game UI

- Tier 2 renders `era_2010` (`ui/themes/era_2010.json`, candidate design data). Tier 3 and above
  still throw.
- The Desk previews `gate.t1_to_t2` at Tier 1 when the pinned routes declare it.
- The optional `transitions.incorporate` control appears exactly when incorporate can apply.
- At `era_2010` the Desk shows a presentation-only FarmVille energy bar. It limits nothing and
  emits nothing, and its curtain is always visible.

See [Game UI](game-ui.md). Candidate copy for owner adoption is in
`copy/catalog/tier2-candidate.json`.

## Pacing measurement (§P2 subset)

`server/harness/tier2_pacing.go` reuses the ratified T0–T1 Chaos and Casual policy literals, with
two changes:

- the elective Exit rule is taken only while Founder Exit history has one entry
  (`t01_c32_readiness_once`);
- `gate.t1_to_t2` joins the crossing set.

It records Founder-attended time at the first crossing in any run. An unreached seed is a visible
`must_reach` failure; it is never dropped. The allocation arms and `reference.greedy` v2 are held
on OD-1.

`TestTier2PacingCalibration` (`CLOUD_CLICKER_T2_PACING=1`) sweeps the gate literal against the
[2 h, 3 h] envelope and pins its report in `balance/testdata/t2/pacing-calibration-v1.json`.
The owner ratifies the literal from that report (§M3).

The independent non-update reproduction covers seven literals and 672 seed runs. None of
those sampled literals satisfies both personas' envelope; at the provisional `1e7`, p50 is
5,054,000 ms for Chaos and 2,605,000 ms for Casual. This finite sweep does not prove that every
literal in `[1e7,1e9)` fails. It establishes neither an adopted balance nor full §P2 acceptance:
Headcount/allocation and the new reference policy remain held.

## Not yet implemented

- §H headcount, the P4 seats row and P5 (OD-1).
- An authoritative §P3 relevance report. The combined `scenario.t1_t2_relevance`
  (`balance/testdata/t2/relevance-scenario-t1-t2-v1.json`, policy `relevance-candidate-v1.json`)
  currently fails its gate. Its non-authoritative diagnostic
  (`relevance-t1-t2-diagnostic-v1.json`) independently reproduces byte-for-byte. All six
  Tier-2 rows are instrument-affected: exact generator removal or removal of an upgrade's
  effect target prevents those findings from being a clean verdict that the content is dead
  or dominated. Reference purchases and deltas are observations of that instrument, not
  sufficient authority to retune. The separate §P3 T0 unchanged-report requirement also fails
  against a fresh epoch-8 control because the ratified Chaos policy's catalog-wide candidate
  population changes (RP-433). Its population/contract must be reconciled explicitly.
  The existing branch diagnostic selects `ping_pong_table` and `move_fast_break_things` on
  controlled legal prefixes, with effect-masked losses of 52,447 ms and 213,996 ms respectively.
  `nap_pod` remains unselected, so that diagnostic also fails. These branch-specific observations
  do not certify the three instrument-affected generators or make the main scenario pass.
- §M mint: goldens, the formulas regeneration, `changelog/epoch-9.md`, and the composed Postgres
  proof.
