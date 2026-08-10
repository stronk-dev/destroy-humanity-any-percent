# T0–T1 candidate proposal

Status: **draft for owner review; not ratified; not mint-authorized**  
RFC: `rfc/t0-t1-playable-content.md`  
Candidate directory: `balance/testdata/t0-t1/`

## Candidate documents

| Document | SHA-256 | Validation state |
|---|---|---|
| `economy-v4.json` | `022a41470853b627fbb8f29202672b85c69ee762bc1a5c0b77f3b65fdc31ecda` | Real Go economy loader + candidate binding test |
| `routes-v1.json` | `a84cce06ae67a68817174b99cfe7191e3c2f9bf47c1c20b4ebab1704baf99cfa` | Real Go Routes loader + candidate literal test |
| `presentation-v1.json` | `70953a6dfa53794f9e1e03627f0b2ddb06abb4870550dccc608a9ca0daeba0d7` | Exact proposed grammar; production loader not implemented |
| `event-copy-v1.json` | `6413fa05f76c56797ec49e82de28ecf81f52cfa502d5b687f8d764d335a94210` | Exact proposed grammar; production loader not implemented |
| `curriculum-v1.json` | `28b5edee4a938171f1c4d866e1f4967a95188fcb151e6a0e60bd53dcabeaacfb` | Exact proposed grammar; transition owner not implemented |
| `harness-scenario-v1.json` | `00c9496cd3e0f95051f50db9243cc757f92eb76194e9f7fe23b461d6ed7c366f` | Exact proposed first-hour grammar; harness extension not implemented |

These hashes are review coordinates, not owner ratifications. Production artifact paths remain
untouched. The mandatory Relevance document is intentionally absent pending T01-C10: the shipped
schema cannot truthfully express a window beginning at run genesis.

## Economy literals and provenance

The candidate keeps every epoch-6 economy policy byte that this RFC does not own: resources,
manual action mechanics, multiplier-source registrations, progress coordinates, manual policy,
offline policy, the Legal Department, and the fiscal multiplier rows.

The authored T0–T1 progression uses the design's geometric cost band and a regular candidate
ladder so review can reason about the curve before harness tuning:

- Generator cost ratios are `1.10`–`1.13`, within design/02's declared `1.07`–`1.15` band.
- Consecutive base costs multiply by `12`; consecutive base production rates multiply by `6.5`.
  Those are provisional pacing literals, not research-derived facts.
- Generator ladders use three `2x` rungs, staggered among `20|25|30`, `50|55|60`, and `100`
  purchased. These are provisional balance rows shaped by design/02 §11b.
- Ten upgrades cost one decade above the associated progression point and contribute a `2x`
  Decimal factor through the shipped `upgrades` slot. The values are provisional.
- Every upgrade declares only `synergy_feed`, because every upgrade is an actual source in one
  shipped pool. Generator rows collectively execute all four ruled role kinds: `provision`,
  `synergy_feed`, `manual_output`, and `stock_rate`.
- `generator.beige_tower_v2` provisions `generator.beige_tower` at `100000 ppm`; pool source
  weights and both pool curves are provisional. The existing 60-second provision grid is kept.
- `gate.t0_to_t1` requires `1e5 company.cash`, with no route rows. This is a provisional pacing
  literal. Every later epoch-6 gate and route byte remains unchanged.

No value above is presented as an empirical fact. Harness results may change these bytes only in
a later reviewed/owner-ratified candidate revision.

## Proposed content grammars

### Presentation v1

The document binds every mechanical generator and upgrade ID, `manual.click`, and the one Horse
Armor shelf stub to explicit Copy keys. `cosmetic.horse_armor_free` is byte-explicitly
non-purchasable and non-stateful. The production loader must require exact key sets, raw-byte
sorted unique IDs, set equality with the pinned economy catalog, Copy registry resolution, and
the `purchasable=false` + `stateful=false` biconditional for v1 cosmetic stubs.

The candidate intentionally supplies bindings, not prose. Copy authoring begins after mechanical
review so rejected IDs do not create orphan Copy rows.

### Event-copy v1

The closed set is the seven event kinds currently needed by the first-hour flow:
`exit_offer_declined`, `exit_offer_expired`, `exit_offer_spawned`, `gate_crossed`,
`generator_purchased`, `run_ended`, and `upgrade_purchased`. Each row declares the only payload
parameters copy may reference; unknown kinds reject. The production loader must prove those
parameters against the registered event schema rather than trusting this document.

### Curriculum v1

The one trigger fires only for Founder exit-count zero, Company run 1, at least 900,000 attended
milliseconds, and `gate.t0_to_t1` crossed. The first subsequent player Company command evaluates
accrual, replaces the requested action with a terminal `scripted_first` Exit, records the
one-shot marker in Founder exit history, grants the normal first-Exit payout, and uses standard
next-run assembly. The declared event order is exact candidate wire grammar. This requires the
ruled logged transition implementation before any mint.

### First-hour harness scenario v1

The proposed scenario declares three deterministic policies, seven exact milestones, seven
envelopes, the complete invariant set, and a two-million-transition ceiling. Times are the RFC's
provisional pacing targets, not claimed results. The harness must reject unknown predicate arms
and prove the composed human-path fixture consumes these same milestone IDs.

## T01-C10 — owner ruling required

Relevance-policy schema v1 requires a non-null concrete `from_gate`. T0 purchasables exist from
genesis, before the first gate, so no schema-v1 row can describe them accurately. Using
`gate.t0_to_t1` would make the mandatory report omit the T0 interval.

Proposed resolution: schema v2 permits `from_gate:null`, meaning run genesis, in policy windows
and scenario segments. Null sorts before all Routes gates; `to_gate` remains exclusive. Existing
schema-v1 bytes retain their meaning. Required proof: strict Go/schema mutation tests, one
`{null,"gate.t0_to_t1"}` candidate window, and a report showing the T0 item was evaluated in that
segment.

## Review request

Review the six pinned draft documents for mechanical shape and provisional numbers, rule
T01-C10, and author the presentation/event/curriculum Copy text only after the ID set is accepted.
After that ruling Codex can implement the missing strict loaders and logged transitions, produce
the mandatory Relevance artifact, run the composed pacing/relevance gates, and return the final
candidate hashes for owner ratification. No epoch-7 mint is authorized by this proposal.
