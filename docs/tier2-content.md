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

## Not yet implemented

- §H headcount, the P4 seats row and P5 (OD-1).
- An authoritative §P3 relevance report. The combined `scenario.t1_t2_relevance`
  (`balance/testdata/t2/relevance-scenario-t1-t2-v1.json`, policy `relevance-candidate-v1.json`)
  currently fails its gate. Its non-authoritative diagnostic
  (`relevance-t1-t2-diagnostic-v1.json`) shows `open_plan_floor` and `managed_services_contract`
  with zero individual contribution to 1e9, and the three Tier-2 upgrades never bought. Retuning
  is an owner decision.
- §M mint: goldens, the formulas regeneration, `changelog/epoch-9.md`, and the composed Postgres
  proof.
