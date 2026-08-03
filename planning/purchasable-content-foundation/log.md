# Purchasable Content Foundation — running log

## 2026-08-03 — pre-acceptance implementation review

- **Review by:** Codex
- **Recorded by:** Codex
- **Reviewed surface:** draft `rfc/purchasable-content-foundation.md` against the economy,
  save-v13, production/`ApplyLogged`, TypeScript replay, multiplier, route, harness, epoch, and
  formula-artifact implementations at `9819b22`.
- **Verdict:** bounced; direction accepted, eleven executable contracts still block acceptance.

### C1 — Artifact and schema ownership

Specify whether P1-P4 extend the existing economy artifact or add an eighth constants artifact.
**Proposal:** economy schema v4 owns `upgrades` and `synergy_pools`, plus generator `tier`,
`category`, `window`, `roles`, `provisions`, and `ladder` fields. This keeps one economy identity;
all seven-artifact `CatalogBundle` and epoch sites remain structurally unchanged. A bundle
validator receives route gate IDs and cross-validates every window/predicate reference. Exact
field optionality, uniqueness, raw-byte ordering, and empty-array rules must be normative.

### C2 — Window and predicate semantics

“The routes/categories union” names two incompatible predicate languages. Tier-0 content also
cannot satisfy a required `from_gate` before any gate exists. **Proposal:** reuse only the route
`Condition` union as an implicit all-of predicate; leaderboard terminal predicates are not legal.
`window` is `{from_gate: string|null, to_gate: string|null}` and is available exactly when
`from_gate == null || crossed[from_gate]` and `to_gate == null || !crossed[to_gate]`. Validate
gate/resource/context references at bundle load. Evaluate eligibility after ordinary accrual and
before spending. Specify whether unmet window and unmet predicate use existing `not_eligible` or
new terminal categories; the closed rejection registry, Go/TS decoders, and docs must agree.

### C3 — Upgrade purchase lifecycle

Define the exact envelope, evaluation order, receipt, event, and duplicate semantics. **Proposal:**
`buy_upgrade` has exact keys `{intent_id,kind,expected_revision,upgrade_id}`, applies one unit only,
and rejects unknown, already-owned, ineligible, or unaffordable deterministically. Accrue first;
evaluate window/predicate; spend one fixed canonical-Decimal cost through the ledger; insert the
ID into a sorted unique owned set; emit `upgrade_purchased` v1; return `applied_count:1`. The
effect source declarations live in the economy catalog and are activated by ownership rather than
duplicated in `multiplier_sources`. State the new terminal categories explicitly.

### C4 — Effect and role unions

`factor|ppm`, “capacity grants,” and manual hooks are not executable schemas. The existing
contribution stack accepts positive Decimal factors and affects generator rates only; dynamic
capacity and manual output have no such path. The design law requires roles on generator classes,
not upgrades, so extending it to every upgrade is an undeclared design deviation.

Before acceptance, declare the literal Phase-0 effect variants with exact fields/formulas and the
literal role variants with their owning state transition and activation record. **Proposal:**
support generator-target and manual-action-target Decimal multiplier contributions in Phase 0;
refactor the same ordered slot product for both. Defer dynamic capacity until its ledger/hardcap
RFC. Enforce the §11b role law on generators; either declare the upgrade role law as an explicit
deviation or remove it. A role declaration cannot count as an activation unless its owner executed
a measurable effect. The T0-T1 minigame-input role cannot ship before a minigame owner exists.

### C5 — Provision-chain time semantics

The faction-stock accumulator is unsuitable as written for producing generator counts. If newly
provisioned units begin producing only at each evaluation boundary, splitting one elapsed interval
into many intents yields more currency than evaluating it once; offline and active play diverge.
Define a partition-invariant model: source counts (purchased only or purchased+provisioned), time
unit, rate dimension, multi-edge order, and whether descendants produced inside the interval also
produce during that interval. Specify topology validation (tier field, adjacent downward edges,
acyclic graph, one/many providers), closed-form/bucket algorithm, offline policy, and Go/TS golden
vectors proving `advance(a+b) == advance(advance(a),b)`. Also decide whether provisioned counts are
safe integers with a visible hardcap or canonical Decimals; an unannounced 9e15 chain ceiling is
not acceptable in the big-number core.

### C6 — Provision persistence and overflow

The required carried remainder is per provisioning edge, but no wire shape or invariant is named.
Specify save fields, canonical ordering, zero-entry handling, migration defaults, caps, addition
overflow, purchased+provisioned rate input, and saturation behavior/reason key. Preserve the
existing `generators` wire field as purchased counts or explicitly version its rename; do not let
Go and TS infer different compatibility rules. Exit/reset/import/genesis/snapshot paths must list
the new fields.

### C7 — Synergy formulas

`linear|log` is not a formula. Define which counts feed pools (proposal: purchased only), upgrade
ownership weight, accumulator numeric domain, base/scale, logarithm base and zero behavior,
quantization point, factor formula, target, source ID, and overflow behavior. Example contracts
such as `linear = 1 + sum_ppm/1e6` and `log = 1 + log10(1 + sum_ppm/1e6)` would be implementable,
but choosing them is balance design. State whether pools may feed one another (proposal: no), and
validate that every pool maps to exactly one declared multiplier source. Golden vectors must cover
raw-byte source ordering and boundary carry, not only ordinary values.

### C8 — Ladder semantics

Specify whether reached rungs multiply cumulatively, replace the previous rung, or add ppm.
**Proposal:** thresholds are strictly increasing positive purchased counts; every reached rung
emits one positive Decimal factor into `milestones`, with mechanical source ID derived
unambiguously from generator ID and threshold; reached factors multiply in normal raw-byte source
order. Provisioned counts never cross a rung. Define `multiplier_ppm` conversion/rounding and
reject neutral/overflowing rows if they are not intentional.

### C9 — Contribution assembly authority

Today external contributions are resolved before `ApplyLogged` and frozen in `replay_inputs`.
Owned upgrades, ladders, and pools are state-derived and must not be resolved twice. **Proposal:**
inside `ApplyLogged`, derive content-owned contributions from state+economy catalog, combine them
with the recorded external contributions, validate the combined unique source set once, then run
accrual. The same function/order is ported to TS. Formula artifacts must be generated from the
same catalog declarations and formulas, including manual targets if C4 accepts them.

### C10 — Ablation boundary

P5 currently contradicts itself: a mask stored on `CatalogBundle` cannot be both honored by
`ApplyLogged` and ignored by replay. It also lets any caller relabel an authoritative catalog
bundle as ablated. **Proposal:** keep masks out of `CatalogBundle` and `replay_inputs`. Refactor to
one private transition implementation receiving a closed simulation policy; exported
`ApplyLogged` always supplies nil, while a separate simulation entrypoint supplies a mask. A
source/import guard permits that entrypoint only from the harness package and tests. Exact mask
namespaces and effects must be declared: generator mask, for example, must say separately whether
it nulls base production, provisioning, ladders, roles, and pool feeds. Removed actions are
filtered by the harness policy; if nevertheless submitted, define the typed rejection. Real-run
replay has no mask-shaped input, making leakage structurally impossible rather than asserted at
composition time.

### C11 — Surface closure and acceptance gates

The RFC must enumerate every closed registry changed by `buy_upgrade` and save v14: Go/TS intent
parsers, client dispatcher, transport vectors, run-log/replay corpus, event-kind Go validation and
DB CHECK migration, snapshot schemas, save migration corpus, import normalization, genesis, Exit
reset, kernel version, JSON Schema, formula generation, and schema/KV guards. Add adversarial
tests for partition invariance, cap/rounding edges, eligibility ordering, duplicate ownership,
masked-vs-real separation, and full Go/TS state/receipt/event byte parity. State the epoch protocol
for the schema-v4 content change and whether it is a mint or hotfix.

Implementation remains blocked until C1-C11 are answered in the RFC. No mechanics were inferred
or coded during this review.

