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

## 2026-08-03 — accepted-ruling reconciliation check

- **Review by:** Codex
- **Recorded by:** Codex
- **Reviewed surface:** owner-ruling delta in the uncommitted RFC after C1-C11.
- **Verdict:** C1-C11's architectural rulings are accepted; C12-C14 are narrow executable-text
  blockers. Implementation did not start because the stated role law still cannot be validated.

### C12 — Remove ruling contradictions from the normative text

The RFC still has `Status: draft`; AC4 still requires the loader to reject a roleless upgrade,
contradicting C4's optional upgrade roles; and both Open Questions repeat choices C4 already
settled (capacity is deferred, manual targets are included). Set the accepted lifecycle state,
change AC4 to generator classes only, remove the resolved questions, and make the P1 effect fields
match the ruled Decimal-only union (`factor`, never `factor|ppm`).

### C13 — Role bindings and executed-mechanic proofs

The four ruled role names are still bare labels. `provision` and `synergy_feed` can be proven from
the existing edge/pool rows, but `manual_output` and `stock_rate` have no catalog binding or
formula. The existing faction hook accrues one unit per interval regardless of generator state,
so merely labeling a generator `stock_rate` would remain a self-confirming role declaration.

**Proposal:** generator `roles` are typed objects, not strings:

- `{kind:"provision", generator_id}` must exactly match that class's provision edge;
- `{kind:"synergy_feed", pool_id}` must exactly match a generator source in that pool;
- `{kind:"manual_output", action_id, per_purchased_ppm}` emits manual factor
  `1 + purchased_count*per_purchased_ppm/1e6`, quantized once, through the `upgrades` slot;
- `{kind:"stock_rate", per_purchased_ppm}` scales faction stock progress by the same factor,
  using exact integer-ppm multiplication with a persisted remainder. Purchased counts only feed
  both roles. Duplicate `(kind,target)` bindings reject. The loader proves every role object has
  its corresponding executable declaration; the harness records activation only when the bound
  mechanic produces a non-neutral result.

If a different stock/manual formula is intended, it must replace this proposal explicitly; field
names alone are not the missing contract.

### C14 — Provision hardcap and bucket wire shape

C5 requires a declared per-class hardcap/reason and C6 requires per-edge remainders, but neither
has an exact catalog/save shape. **Proposal:** every provision target generator declares
`provisioned_hardcap:{count:int64,reason_key:string}`; save v14 adds
`generators_provisioned:{id:int64}` and `provision_remainders_ppm:{target_id:int64}` with complete
catalog key sets, including zeros. At each absolute tick, all edges read the pre-boundary total
counts, compute `numerator = source_total*rate_ppm + prior_remainder` in overflow-safe integer
arithmetic, stage `floor(numerator/1e6)`, retain `numerator mod 1e6`, then commit all staged target
increments simultaneously with accrual-only saturation. New units affect the following bucket.
The cap reason is exported in the authoritative snapshot/formula artifact, and the client uses it
when a provisioned counter is frozen.

C12-C14 must be reconciled in the RFC before the catalog/save implementation batch begins.

## 2026-08-03 — catalog and save-v14 foundation landing

- Implemented economy schema v4 loaders in Go and TypeScript for typed generator roles,
  provision topology/caps, staggered ladders, upgrades, and synergy pools. Both runtimes consume
  `testdata/economy-foundation-v4.json`; bundle loading cross-validates upgrade gate and resource
  references against the pinned Routes catalog.
- Added save v14's closed wire state (`upgrades_owned`, complete provisioned-count and per-edge
  remainder maps, and the stock-rate remainder), with strict current-version presence checks,
  catalog cap/identity validation, and v13 zero-state migration in both runtimes.
- Regenerated the shared ApplyLogged fixture at save v14. This is an encoding-only replay delta;
  authoritative behavior and the live schema-v3 Phase-0 economy artifact remain unchanged until
  the RFC's governed schema-v4 mint.
- Proof run: `make test-go GO_PACKAGES='./save ./economy ./routes ./production ./replaycatalog'`,
  `make typecheck`, `make test-client`, and `make replay-fixture-check` are green (6,496 client
  tests passed, 3 skipped).

## 2026-08-03 — authoritative mechanics and cross-runtime landing

- Added state-derived contribution assembly inside both Go replay boundaries: owned-upgrade
  bundles, cumulative purchased-only ladders, manual-output roles, and purchased-only linear/log
  synergy pools combine once with frozen external contributions and pass the existing source
  validator/order.
- Implemented `buy_upgrade`, typed ownership/window/predicate/unaffordable rejections, strict
  ledger purchase, and the closed `upgrade_purchased` event across the Go parser, event registry,
  and append-only database CHECK migration.
- Implemented absolute `run_started_at` provisioning buckets with overflow-safe integer ppm,
  simultaneous staged commits, carried remainders, next-bucket production, and accrual-only cap
  saturation. Production reads purchased+provisioned; costs, ladders, pools, manual roles, and
  stock-rate roles read purchased only.
- Ported the same mechanics to TypeScript and added four Go-authored schema-v4 replay cases for
  upgrade purchase, manual role output, fixed-grid provisioning, and combined ladder/pool/upgrade
  execution. Existing fixtures now construct the complete v14 state used by restored production
  saves and new prestige runs.
- Self-review correction: `buy_upgrade` initially inherited the Route Registry projection
  dependency from route hints. It now requires only the immutable Routes catalog; only the
  founder-scope hint path requires the projector.
- Proof run: `make test-go GO_PACKAGES='./production ./economy ./faction ./prestige ./save'`,
  `make typecheck`, `make test-client` (6,500 passed, 3 skipped), and
  `make replay-fixture-check` are green.

## 2026-08-03 — simulation-only ablation boundary

- Added the closed `AblationMask` and one simulation entrypoint used by the balance harness. The
  authoritative `Transition`/`ApplyLogged` paths still have no mask-shaped parameter, catalog
  field, save state, or replay input.
- Generator masks null base production, ladders, manual-role factors, pool feeds, and outgoing
  provisioning while preserving units and costs. Upgrade masks null their contribution bundles;
  removed actions reject `unknown_id` before accrual.
- Added a Go-AST repository guard with seeded positive/decoy self-tests: non-test calls to the
  simulation entrypoint are permitted only from `server/harness`. Masked generator, provision
  edge, and removed-action fixtures prove the ruled semantics and rollback boundary.
- Added exact linear/log synergy-factor golden assertions. Proof run:
  `make test-go GO_PACKAGES='./production ./harness'` is green.
- KV-1 correction is staged as autosquash fixups because the two preceding unreviewed Foundation
  commits omitted their required in-commit kernel bumps. No implementation review cites those
  hashes and nothing is pushed; the history gate remains intentionally red until the pre-review
  autosquash repair makes the three semantic landings `0.3.9`, `0.3.10`, and `0.3.11`.

## 2026-08-03 — closure surface and canonical artifacts

- Added schema-v4 JSON Schema coverage, the client `buy_upgrade` dispatcher arm, exact
  provisioned-cap/reason snapshots in both replay kernels, and generated formula schema v6. The
  formula artifact now fingerprints the executable purchased+provisioned rate, contribution,
  provisioning, multiplier-order, and Commons authorities and exports the ruled provisioning,
  manual, stock, ladder, and linear/log pool formulas.
- Audited every non-test evaluation site after introducing the simulation policy. Exit-offer
  decline was the one bypass: it now uses the policy-aware evaluator, with a regression proving a
  masked producer accrues nothing on that intent. Authoritative replay remains mask-free.
- Added the acceptance criterion's two-pool fixture. It loader-validates a second logarithmic pool
  and proves raw-byte pool order plus exact factors (`1.01e0`, `1.77815125038e0`). The shared
  Go-authored replay fixture continues to prove schema-v4 state, receipt, and event byte parity in
  TypeScript, now including the complete provisioned hardcap/reason map.
- Added canonical `docs/purchasable-content.md` and reconciled the economy/production docs without
  claiming a live content epoch: the engine supports schema v4 while the active Phase-0 balance
  artifact remains schema v3. Therefore this Foundation changes executable capability and kernel
  version but does **not** mint balance data; the T0–T1 content RFC owns that later mint.
- Proofs in this landing: strict schema-v4 fixture validation, save-v14 and cross-runtime replay
  corpus, client dispatcher exact-key tests, two-pool ordering, cap snapshot parity, and the
  policy-aware decline regression. `make typecheck`, `make test-client` (6,500 passed, 3 skipped),
  `make verify-schema`, and `make replay-fixture-check` are green. Formula generation is
  deterministic; `formulas-check` is expected to compare clean after this generated artifact is
  committed.
- Kernel `0.3.12` covers this final receipt/snapshot semantic change. Before independent review,
  the two earlier fixup commits must be autosquashed so every semantic commit carries its own
  version bump and the history-walking KV-1 gate can pass without exception.
