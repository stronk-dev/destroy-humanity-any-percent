# T0–T1 candidate proposal

Status: **eight-document core owner-ratified; replacement phase-scoped relevance pins are review
candidates; not mint-authorized**

RFC: `rfc/t0-t1-playable-content.md`

Candidate directory: `balance/testdata/t0-t1/`

## Candidate documents

| Document | SHA-256 | Validation state |
|---|---|---|
| `categories-v1.json` | `ff63b341ff8a7439e48cbfa7cb91dcf51089809fcbb0e6e54201965e5911b9a5` | Real Go leaderboard loader + candidate full-gate-set test |
| `curriculum-v1.json` | `17e5e0c7e8b8f7217c6063b41067af0bed41a34cc26a22e9d4ddfc00513e98d9` | Exact proposed grammar + named Copy binding; transition owner not implemented |
| `economy-v4.json` | `fb75e5cf32f545d9470cc8512a8c63f45ed9edd96c68ba65cfeabe0ce2c7f37d` | Real Go economy loader + candidate binding test; owner-signed C24–C27 tuple applied |
| `event-copy-v1.json` | `6413fa05f76c56797ec49e82de28ecf81f52cfa502d5b687f8d764d335a94210` | Exact proposed grammar; production loader not implemented |
| `harness-scenario-v1.json` | `e74e271be3b844bfde411887af16de06890a9a281596d45b8ad9deb7b1a502a5` | Exact proposed first-hour grammar; harness extensions not implemented |
| `opportunities-v1.json` | `63e51084863bd00da7d5a0b358f54741b0b0682d8ef25b2fc7cb3da2c77f27cb` | Real Go active-play loader against candidate economy |
| `presentation-v1.json` | `70953a6dfa53794f9e1e03627f0b2ddb06abb4870550dccc608a9ca0daeba0d7` | Exact proposed grammar; production loader not implemented |
| `relevance-policy-t0-v2.json` | `18a531274f60a7b53517c8c591204e4d02c8050691b7ffea960a1c166ed8c235` | Real Go loader against candidate economy/routes; exact pre-T1 item set |
| `relevance-policy-t1-v2.json` | `f513360cc421e9b5a4ca624c977fdde055104b52052952e28a4aa5d5443554ef` | Real Go loader against candidate economy/routes; cumulative exact pre-T2 item set |
| `routes-v1.json` | `a84cce06ae67a68817174b99cfe7191e3c2f9bf47c1c20b4ebab1704baf99cfa` | Real Go Routes loader + candidate literal test |

The non-relevance hashes above retain their recorded owner ratification in
`rfc/t0-t1-playable-content.md`; production artifact paths remain untouched. T01-C18 replaces the
old single relevance coordinate with two review candidates: T0 scenario
`008c08df62da0792b84b5a1c5367f52cdfbfaa5ac46603fd920dcdaa94035a18` and T1 scenario
`0d6049b2736eee7560d82f9459972f8da62e7b8b1933fca885631b0e09195419`, paired with the two policy
hashes above. All four are pending designated review and owner re-ratification; the historical
pin #9 remains an immutable record, not authority for these replacement bytes.

## Findings-first remediation

- **F1 — closed:** `categories-v1.json` is the live epoch-6 categories artifact plus the one
  raw-byte-sorted `gate.t0_to_t1` insertion in `full_gate_set`. A real loader test composes it
  against the candidate Routes gate set.
- **F2 — superseded by the owner-signed C24–C27 tuple:** all paired `cost.amount` and
  `resource_at_least.value` bytes agree. The final signed coordinates are Continuous Feed `1.2e4`,
  Hold Music `2e4`, Business Cards `2e4`, Reply-All `4e1`, CRT Degauss `2.48832e8`, Handbook
  `7.5e7`, Refurbished Sticker `2e8`, and Institutional Memory `1e8`; Institutional Memory targets
  `generator.garage_rack`. These remain provisional candidate bytes until designated review and
  owner re-ratification.
- **F3 — closed:** the body now declares the actual `1.07`–`1.13` T0 band. The `1.13` Beige Tower
  ratio deliberately preserves its live epoch-6 byte; this is declared preservation rather than
  silently narrowing the rule around it.
- **F4 — declared change:** schema v4 requires an executable role for Legal Department. The
  candidate deliberately makes it a `synergy_feed` source in `pool.institutional_knowledge` at
  provisional `4000 ppm`, so a permit faucet also feeds the global upgrades multiplier. This is
  a new owner-ratified balance choice, not epoch-6 byte preservation.
- **F5 — contract made honest:** the six shipped invariant names remain
  `state_encodes`, `numeric_domain`, `resource_bounds`, `ledger_reconciles`,
  `revision_monotone`, and `must_reach`. Three candidate harness extensions are explicit:
  `artifact_identity` asserts the pinned bundle hash never changes during a run;
  `replay_parity` byte-compares Go/TypeScript state, receipt, and event outputs; and
  `role_activation` requires every declared generator-role binding to execute a non-neutral
  result. Unknown names reject until those runner arms land.
- **F6 — closed:** every milestone names its clock. Run 1 and run 2 gate milestones explicitly
  name their run sequence, and `relations` requires run 2's attended-time gate crossing to be
  strictly earlier than run 1 for the same seed.
- **F7 — closed:** the curriculum row now binds its three narrative keys in a closed `copy`
  object (`title_key`, `body_key`, `next_run_key`). The presentation/Copy loader must resolve
  them before mint.
- **F8 — documented:** manual refill ppm, stock-rate ppm, provision ppm, synergy weights/curves,
  the two-million transition budget, and persona seeds are provisional deterministic balance or
  computation-budget literals, not empirical facts. `9007199254740991` is the shared exact-
  integer interoperability ceiling, not a gameplay or research claim.
- **F9 — closed but still gate-owned:** Beige Tower v2's truncated rate is corrected to
  `4.90222789063e5`. Its provisioning role is deliberately showcased rather than trap-exempt;
  it must therefore pass the same mandatory relevance gate as every other purchasable.
- **F10 — carried explicitly:** T01-C6/AC0 still owns the session-bootstrap offline-catchup
  implementation and >24h replay fixture. Copy assembly must follow FCE-C8's orphan-first order:
  Copy rows land before any candidate artifact references them, then the owner-gated mint moves
  the complete referenced set atomically.

## Economy literals and provenance

The candidate keeps every epoch-6 economy policy byte this RFC does not own: resources, manual
action mechanics, progress coordinates, manual/offline policy, the Legal Department's core
permit production, and fiscal multiplier rows. The three active-play multiplier declarations
are added because the EH-C8 opportunities artifact makes those source IDs part of the composed
candidate.

The authored T0–T1 progression is deliberately regular so the harness can expose rather than
hide pacing errors:

- Generator ratios are `1.10`–`1.13`, within design/02's `1.07`–`1.15` band; consecutive base
  costs multiply by `12`, and production rates multiply by `6.5`.
- Generator ladders use provisional `2x` rungs staggered among `20|25|30`, `50|55|60`, and `100`.
- Each upgrade contributes a `2x` Decimal factor through the shipped `upgrades` slot. The revised
  prices above keep the affected rows reachable inside their declared windows.
- Generator rows collectively execute the four shipped role kinds: `provision`, `synergy_feed`,
  `manual_output`, and `stock_rate`. Capacity and minigame-input remain excluded.
- Beige Tower v2 provisions Beige Tower at provisional `100000 ppm` on the existing 60-second
  absolute grid. Pool weights and both pool curves are provisional.
- `gate.t0_to_t1` requires provisional `1e5 company.cash`, with no route rows. Every later
  epoch-6 gate and route byte is unchanged.

None of those literals is represented as empirical fact. Changes after review require a new
candidate hash and owner ratification.

## Proposed content grammars

### Presentation and event copy v1

Presentation binds every mechanical generator and upgrade, `manual.click`, and the Horse Armor
shelf stub to explicit Copy keys. The stub is byte-explicitly non-purchasable and non-stateful.
The future loader must enforce exact sets against economy, resolve every Copy key, and reject any
stateful v1 cosmetic. Event copy remains the closed seven-kind set in the candidate; unknown
kinds reject and parameters must match the registered production event schemas.

### Curriculum v1

The trigger applies only to Founder exit-count zero, Company run 1, at least 900,000 attended
milliseconds, and `gate.t0_to_t1` crossed. The first later player Company command evaluates
accrual, replaces the requested action with terminal `scripted_first`, records the one-shot
Founder-history marker, grants the standard first-Exit payout, and uses normal next-run assembly.
Its exact event order and three Copy bindings are candidate wire grammar. The ruled logged
transition remains implementation debt.

### First-hour harness scenario v1

The scenario declares three deterministic policies, seven clocked milestones, seven envelopes,
one same-seed run-2-faster relation, all six shipped invariants plus the three named extensions,
and a two-million-transition ceiling. Times are provisional targets, not claimed results. The
composed human-path fixture must consume these same milestone IDs.

### Opportunities v1

This is byte-identical to the reviewed Active-Play foundation fixture: the gamma6 schedule,
building/click/lucky/production effects, and combo cap. It is intentionally a candidate artifact,
not a production mint. The economy document adds exactly the three multiplier declarations that
its multiplicative effects resolve; Lucky is a payout, not a multiplier source.

## T01-C11 — implemented candidate boundary

Schema v2 now permits multiple ordered, non-overlapping segments for the same sole milestone;
schema v1 remains exactly one segment. The candidate uses `[run genesis,T0→T1)`,
`[T0→T1,T2→T3)`, and `[T2→T3,T3→T4)`. The Go loader rejects the ruled semantic matrix, the JSON
Schema rejects v1-multiple and duplicate rows, and a schema-v2 fixture proves report/delta bytes
remain single-milestone. Candidate epsilon, horizon, seed, and budget literals are provisional
review coordinates, not empirical claims or production authorization.

## Review request

Re-review the eight present candidate documents and the F1–F10 closures. Rule T01-C11 before the
ninth Relevance document is authored. After that, Codex can implement the missing strict loaders,
logged transitions, and harness extensions; run the composed pacing/relevance gates; and return
the complete candidate hashes for owner ratification. No epoch mint is authorized here.
