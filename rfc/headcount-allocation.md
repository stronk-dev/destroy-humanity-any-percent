# RFC: Headcount Allocation

- **Status:** accepted — owner batch acceptance 2026-09-25; implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/01-tiers.md` Tier 2 (Produce / New grammar / New decision) and Tier 3
  (Automates); `design/07-roadmap.md` Phase 1; `design/10-playstyles.md §4` (`No Hires`);
  `design/02-economy-balancing.md §2.2`, `§2c`, `§7`, `§11b`; `design/06-tech.md §idle-math`;
  `design/research/idle-landscape.md §7` (paradigm map) and its NGU / Kittens entries
- **Depends on:** Production Engine & Intent API (archived), Purchasable Content Foundation
  (archived), Save Layer & Migrations (archived), Production Accrual Math (archived), Run Genesis &
  Replay (archived), Leaderboards & Balance Epochs (implementing; kernel identity and replay
  verifier), Balance Harness Foundation + Relevance Harness (archived), Meters (archived; only for
  OD-5)
- **Parent / amends:** Adds an economy-engine grammar that sits beside Purchasable Content. It
  amends no archived RFC. The canonical docs it would update are listed in AC-12.
- **Sibling:** `rfc/tier2-content.md` (being drafted in parallel). **This RFC owns the engine
  only.** Tier2-content owns the catalog rows: the actual T2 pool, its roles, their rates, prices
  and caps, the T1→T2 gate, copy, and T2 pacing. See §S0.
- **Supersedes / superseded by:** —
- **Planning:** `planning/headcount-allocation/` (once accepted)
- **Evidence coordinate:** repository HEAD `c2d9bbc`. At that commit: `kernel/VERSION` is
  `0.3.101`; `server/save/state.go` sets `LatestCompanyVersion = 18`; `server/save/runlog.go` sets
  `ReplayInputsVersion = 8`; `balance/catalogs/phase0.json` is economy `schema_version: 4`; the
  epoch-8 bundle includes `opportunities`, so new runs start on Company v18. This draft is a static
  trace of design, docs, schema and engine source. No product code was run for it.

## Summary

`design/01` makes Tier 2's new grammar **allocation**: the player has a finite number of hires and
splits them across roles. The economy catalog has no way to express this today. The only generator
roles are `provision`, `synergy_feed`, `manual_output` and `stock_rate`. There is no pool of seats,
no role assignment, and no intent that changes an assignment.

This RFC adds one **headcount pool** grammar to the engine:

- **Data.** Economy catalog v5 gains `headcount_pools`. Each pool has a gate window, a hire price
  on an existing cost curve, a visible seat hardcap, and a closed list of roles. Each role has
  exactly one effect from a closed set: `produce`, `convert` (the Middle Manager 1:1 shape), or
  `pool_feed`.
- **State.** Company save v19 stores two exact-integer maps: seats hired per pool and heads
  assigned per role.
- **Intents.** Two logged intents: `hire_headcount` and `reallocate_headcount`.
- **Arithmetic.** An integer effective-heads kernel feeds the existing rate sum. Headcount is
  therefore constant between intents, and production stays closed-form and lazy.
- **Everything else.** Replay, Go/TS parity, harness and UI contracts, all with fail-closed gates.

This RFC ships no gameplay rows. An empty `headcount_pools` array must behave exactly like
catalog v4.

## Motivation

`design/07` Phase 1 lists "headcount allocation" as a v0.1 deliverable. The v0.1 release manifest
(`rfc/v0.1-garage-release-manifest.md` row G05) records it as **Absent**, with "no allocation
primitive". It proposes this split under RFC-0000 rule 2 in case `tier2-content` does not specify
the grammar. Keeping the engine separate from the content does three things:

1. **Tier2-content stays catalog-only.** It can be fixture-first and tuned by the harness, as the
   First Content Epoch foundations were.
2. **The engine is independently testable.** Kernel-version, replay-input and save-version changes
   form one reviewable unit with its own cross-runtime corpus.
3. **Law 2 is proved once.** The headcount contribution is shown to be constant between intents
   here, instead of being argued again for each content row.

**Out of scope:** the concrete T2 roles, rates, prices, caps and copy (`tier2-content`); morale
modulation (OD-5); crunch events and Soul debits (Events Layer 1); payroll upkeep (OD-9); layoffs
(an Events-L1 beat, OD-10); Tier 3 HR automation and Middle-Manager self-replication; the `No Hires`
challenge runtime; the mounted UI panel (OD-11).

## Design basis (quoted) and what design does not say

These quotes are the full design authority for this RFC. Everything in the Specification comes
from them, reuses an archived mechanic, or appears as an Owner Decision.

| Ref | Quoted text | Used for |
|---|---|---|
| `design/01` T2 Produce | "billable output from **headcount**." | Heads produce resources (`produce` effect) |
| `design/01` T2 New grammar | "allocation — a finite hiring budget across roles (Engineers / Sales / Support / Middle Managers who convert Engineers into Slide Decks at 1:1). Kittens-Game/NGU worker-allocation paradigm. Employees have morale; crunch events introduce the **Soul** ledger" | Pool + roles; `convert` effect; seat cap; morale → OD-5; crunch → excluded |
| `design/01` T2 New decision | "allocation ratios and re-allocation timing" | `reallocate_headcount` as a first-class intent; timing → OD-2 |
| `design/01` T3 Automates | "headcount management (HR department; Middle Managers now self-replicate, which is the joke)." | Excluded here; window `to_gate` is the seam (§S5.3) |
| `design/07` Phase 1 | "Tiers 0–2 complete: cost-curve economy, headcount allocation, first Exit loop, Reputation tree v1." | Phase placement |
| `design/10 §4` | "`No Hires` \| headcount system deleted; automation only \| delete a system" | Empty-pool behavior must equal v4 (AC-2) |
| `research/idle-landscape.md §7` | "IT company \| **hiring/allocation** — finite headcount across roles \| Kittens Game, NGU" | Paradigm |
| `research/idle-landscape.md` NGU | "you allocate a finite *energy/magic/resource-3 budget* across dozens of parallel stat trainers, and the allocation itself is the game." | Fixed budget, free re-split (OD-2 default) |
| `design/02 §2.2` | "Hard rule: **hardcaps, never softcaps.** Any cap is a visible number with a tooltip explaining it." | Seat hardcap with `reason_key`; `min()` bottleneck, no diminishing returns |
| `design/06 §idle-math` | "Never tick players server-side … integrate production analytically over Δt (constant: `rate×Δt` …). Bucketed simulation … only for threshold-crossing mechanics that resist closed form." | Heads are constant between intents; no bucketing is added |
| `design/02 §2c.1` | "Our kernel ships **mandatory flow-through sinks** (upkeep, consumables, training fees …)" | Payroll question → OD-9 |
| `design/02 §2c.2` | "Every cap mints a currency … Any cap we ship is designed KNOWING what parallel currency it mints — named in the catalog" | Seat cap → OD-12 |
| `design/02 §7` | "Trust: five constituencies × two independent bars — {Users, **Employees**, Regulators, Press, Investors} × {Standing, Grievance}" … "Meters modulate; facts gate" | Morale candidate → OD-5 |
| `design/02 §11b.4` | "every generator class declares ≥1 non-production role in the catalog, loader-enforced" | Scope question for roles (Deviations D-4) |
| `research/endgame-grammar.md §5.1` | "`stage_rate = allocated × per_unit_rate × multipliers`" and "Throughput = `min(stage₁, stage₂, stage₃)`" | **Precedent only** (T7 research, not T2 authority): a fixed allocation with a `min()` bottleneck stays closed-form |

**What design does not specify.** Each gap below becomes an Owner Decision:

- how seats are obtained, what the "budget" is denominated in, and how it grows (OD-1);
- whether reallocation costs anything or has a delay (OD-2);
- whether the Middle Manager conversion works on rates or on stock, and what a Slide Deck does
  (OD-3);
- whether production multipliers apply to headcount output (OD-4);
- what morale is and how it acts (OD-5);
- what Sales and Support actually do — a content gap that `tier2-content` must close using the
  effect vocabulary fixed by OD-6;
- whether wages or payroll exist (OD-9);
- whether the player can fire staff (OD-10).

## Specification

The Specification is written against the **recommended default** of every Owner Decision. If a
decision is ruled otherwise, the named sections change in the same edit, under the
body-reconciliation rule.

### S0 — Division of responsibility with `tier2-content`

| Concern | This RFC | `tier2-content` |
|---|---|---|
| `headcount_pools` schema, loaders, validation | ✅ | consumes |
| Effect vocabulary (`produce` / `convert` / `pool_feed`) | ✅ | picks per role |
| Effective-heads kernel and rate composition | ✅ | — |
| Save v19, intents, events, replay-inputs v9, kernel bump | ✅ | — |
| Fixture catalog + cross-runtime corpus (mechanical IDs only) | ✅ | — |
| The T2 pool row: roles, rates, prices, cap, window gates | — | ✅ |
| Slide Deck resource row and any consumer of it | — | ✅ (OD-3) |
| Copy keys, flavor, and the seat-cap currency name (OD-12) | — | ✅ |
| T2 pacing targets and the content epoch mint | — | ✅ |

The fixture catalog in this RFC uses mechanical IDs such as `headcount.fixture`,
`role.fixture_producer` and `role.fixture_converter`. It must not reuse or pre-empt any
`tier2-content` ID.

### S1 — Economy catalog schema v5

Catalog v5 is v4 plus one required top-level array, `headcount_pools`. The array may be empty. v1–v4
stay loadable under their pinned semantics, and a v4 catalog can never acquire headcount semantics.

```json
"headcount_pools": [
  {
    "id": "headcount.fixture",
    "window": { "from_gate": "gate.fixture_open", "to_gate": null },
    "hire_price": {
      "resource_id": "company.cash",
      "base": "1e3",
      "curve": { "kind": "geometric", "ratio": "1.13e0" }
    },
    "seat_hardcap": { "count": 500, "reason_key": "headcount.fixture.seat_cap" },
    "roles": [
      { "id": "role.fixture_producer",
        "effect": { "kind": "produce", "resource_id": "company.cash", "base_rate_per_head": "5e0" } },
      { "id": "role.fixture_converter",
        "effect": { "kind": "convert", "from_role_id": "role.fixture_producer",
                    "heads_per_converter": 1, "resource_id": "company.fixture_decks",
                    "base_rate_per_head": "1e0" } },
      { "id": "role.fixture_feeder",
        "effect": { "kind": "pool_feed", "pool_id": "pool.fixture" } }
    ]
  }
]
```

**Field rules.** Both loaders (`economy.LoadCatalog` in Go and `parseCatalog` in TypeScript) and
the JSON Schema enforce all of these. Any violation rejects the whole catalog.

1. **`id`.** Every pool ID and role ID is a lowercase mechanical identifier. It must be unique
   across all pools and roles. It must also be distinct from every resource, generator class,
   upgrade, manual action, synergy pool and multiplier source ID in the catalog.
2. **`window`.** Same shape and meaning as the upgrade window: the pool is open once `from_gate`
   has been crossed and before `to_gate` is crossed. Both gates may be `null`. Gate IDs are checked
   against the pinned Routes artifact in both runtimes, the same way doctrine references are.
3. **`hire_price`.** Uses the existing `price` definition, so the curve kinds are exactly
   `constant`, `linear` and `geometric`. The price resource must be Company-scoped. The owned count
   `n` is the pool's current seat count.
4. **`seat_hardcap`.** `count` is a positive safe integer and `reason_key` is a mechanical ID.
   Neither can be null: this grammar has no uncapped pool.
5. **`roles`.** At least one role. The order in the file is not significant; every computation
   iterates roles in raw-byte ascending ID order.
6. **`effect`.** Exactly one per role, from a closed union:
   - `produce {resource_id, base_rate_per_head}`: the resource must be Company-scoped, and the rate
     is a positive canonical Decimal per second.
   - `convert {from_role_id, heads_per_converter, resource_id, base_rate_per_head}`:
     - `from_role_id` must name a `produce` role in the **same pool**. That means no chains, no
       self-reference and no cross-pool conversion.
     - At most one `convert` role may name a given source role.
     - `heads_per_converter` is a positive safe integer. The design value is `1` ("at 1:1"), but it
       is data, not a constant.
     - The resource and rate follow the same rules as `produce`.
   - `pool_feed {pool_id}`: binds to a declared synergy pool. That pool must list this role as a
     source (S1.1). This mirrors the existing `synergy_feed` binding.
7. **Roles with no effect, or an unknown effect `kind`, are rejected.** Roles describe executed
   mechanics; they are not labels.

**S1.1 — Synergy-pool source kind `role`.** v5 widens `synergy_source.kind` from
`generator | upgrade` to `generator | upgrade | role`. A `role` source contributes
`assigned_heads × per_count_ppm` to the pool sum. The existing linear and log pool formulas,
quantize-once rule, and "pools cannot feed pools" rule are unchanged. A `role` source without a
matching `pool_feed` binding is rejected, and so is the reverse.

**S1.2 — Multiplier targets.** In v5, both `multiplier_sources[].target` and
`upgrades[].effects[].target` may name a role ID, in addition to the existing generator IDs,
manual-action IDs and `all`. A role target that does not resolve rejects the catalog.

### S2 — Effective-heads kernel (pure exact integers)

Inputs: the pinned pool definition, `seats[pool]`, and `assigned[role]` for every role in the pool.
Everything below is a safe integer. No Decimal is used before S3.

```text
for each convert role c, with source s = c.from_role_id and k = c.heads_per_converter:
    capacity_c = assigned[c] × k                       (conceptual; never computed if it could overflow)
    diverted_c = assigned[s]           if assigned[c] > floor(assigned[s] / k)
               = assigned[c] × k       otherwise       (this product is ≤ assigned[s], so it is safe)
    idle_converters_c = assigned[c] − ceil(diverted_c / k)
producing_heads[s] = assigned[s] − diverted_c         (or assigned[s] when s has no converter)
producing_heads[r] = assigned[r]                      for pool_feed roles (they only feed the pool)
unassigned[pool]   = seats[pool] − Σ_{r in pool} assigned[r]
```

- This is the hard `min()` bottleneck: `diverted_c = min(assigned[c]·k, assigned[s])`, evaluated
  without forming a product that could overflow. Extra converters sit **idle** and are shown to the
  player, not hidden (S10). There is no softcap and no diminishing return.
- `diverted_c`, `producing_heads`, `idle_converters_c` and `unassigned` are **derived and never
  persisted**. The save holds only `seats` and `assigned` (S4).
- The kernel is a pure function over the pool definition and those two maps. It reads no clock and
  no Decimal state. Go and TypeScript implement it identically, and TypeScript uses integer-safe
  `Math.floor` / `Math.ceil` only on values ≤ 2^53−1.

### S3 — Rate composition (closed-form, constant between intents)

The published production formula (`docs/generated/production-formulas.json`) gains a second term:

```text
rate(resource) =
    Σ_generators ( (purchased + provisioned) × base_rate × Π contributions(target ∈ {generator, all}) )
  + Σ_produce roles r with r.resource = resource ( producing_heads[r] × r.base_rate_per_head × Π contributions(target ∈ {r, all}) )
  + Σ_convert roles c with c.resource = resource ( diverted_c × c.base_rate_per_head × Π contributions(target ∈ {c, all}) )
```

- Contributions keep the existing slot order (upgrades, milestones, faction, doctrine, commons,
  trust, event buffs, prestige). Within a slot they apply in raw-byte source-ID order, through the
  **same private contribution function** that generators use. Contributions that target a
  generator never apply to a role (OD-4).
- Headcount integers convert to Decimal through the same exact-count conversion that generator
  counts use.
- Each role term is appended to that resource's rate list inside `ratesWithProvisionedAndPolicy`,
  and so enters the existing `accrueBoostedConstant` → `SumDeterministic` → one-quantization path.
  The rest of the pipeline is untouched: fixed provision buckets, offline efficiency and 24 h cap,
  Compute-burst bonus segments, ledger saturation at resource hardcaps.
- **Constancy invariant (Law 2).** Role terms depend only on pinned catalog data, the `seats` and
  `assigned` maps, and the contribution set that Evaluate already receives as a fixed input. `seats`
  and `assigned` change **only** inside an applied `hire_headcount` or `reallocate_headcount`
  transition. So between two consecutive commands, every role term is a constant, `rate × Δt`
  integrates it exactly, and no per-player tick, timer or bucket is added. AC-6 must prove this.
- A converted Engineer's output **is redirected, not duplicated**. Diverted heads leave the source
  role's term and appear only in the converter's term. Conversion never creates a negative ledger
  entry, so `ApplyAccrual`'s non-negative rule still holds (OD-3).
- The pure rate projection used for authoritative UI rates gains **role rows**, sorted by role ID.
  Each row has `{role_id, resource_id, heads, rate}`, where `heads` is `producing_heads` or
  `diverted_c`. The projection still advances nothing.
- `make formulas-check` must pick up the new rate authority. Regenerating the artifact is its own
  reviewed commit, as the docs already require.

### S4 — Company save v19

Company v19 is v18 plus two required complete-key maps:

```json
"headcount_seats":    { "headcount.fixture": 12 },
"headcount_assigned": { "role.fixture_converter": 3, "role.fixture_feeder": 0, "role.fixture_producer": 8 }
```

- **Complete keys.** Every pool in the pinned economy artifact appears in `headcount_seats`, and
  every role of every pool appears in `headcount_assigned`. No extra keys are allowed. Values are
  JSON integers from 0 to 2^53−1. Role IDs are unique across the whole catalog (S1 rule 1), so a
  flat map is unambiguous.
- **Invariant, checked on encode, restore and every transition:**
  `0 ≤ Σ_{r∈pool} assigned[r] ≤ seats[pool] ≤ seat_hardcap.count`. A violation rejects the whole
  state and is never clamped.
- **Company scope only.** Founder, Guild and World codecs reject both fields.
- **Activation.** v19 is only written for a run whose pinned economy artifact is schema v5. It
  follows the Company scalar chain, so a v5 epoch must also carry the v16–v18 prerequisite
  artifacts. The version-floor registry records "economy v5 ⇒ Company ≥ 19". Exit derives the next
  run's version from the next pinned bundle, as it does today. A run pinned to a v4 or earlier
  economy artifact finishes on its current version and never sees headcount. This is the same
  pattern as the Meters, doctrine and active-play overlays. There is no in-place migration of live
  runs.
- **Migration corpus.**
  - A v18 → v19 codec step exists for completeness. It initializes every pool and role to 0, and
    the checked-in migration corpus gains a v18→v19 case.
  - Restoring a v18 state against a pinned v5 artifact that has non-empty pools fails closed. Such
    a pairing can only come from mislabeled evidence.
  - The migration-corpus baseline manifest ratchets its case count in the same reviewed commit.
- **New run and Exit.** A new run's genesis has all zero seats and assignments. Headcount is
  Company-run state (`design/02 §1` layer 1: it resets on Exit) and is never carried to Founder.
  Terminal Exit state records the values as they were.
- The next free SQL migration expands the closed event-kind constraint (S6). Applied migrations
  are append-only.

### S5 — Intents

Both intents are replay-owned Company commands. They go through `save.Store.ApplyIntent` →
`ApplyLogged` with the standard envelope: a lowercase UUIDv7 `intent_id`, a positive safe-integer
`expected_revision`, and a SHA-256 request hash over deterministic JSON that excludes `intent_id`.
Clients never send prices, rates, seat totals or resulting state.

**Order inside every applied transition.** First, evaluate production through `effective_now`
using the **pre-command** headcount. Second, validate. Third, mutate. The new allocation therefore
takes effect exactly at the command instant, in the same way a `buy_generator` count change does.
A reallocation in the middle of a provision bucket splits that bucket's accrual at the command
instant, exactly as a mid-bucket purchase does. It never applies retroactively and never waits for
the next boundary.

#### S5.1 — `hire_headcount`

```json
{ "kind": "hire_headcount", "pool_id": "headcount.fixture", "count": 3 }
```

`count` is a positive safe integer or the literal `"max"`. `"max"` means the verified
max-affordable count from the existing kernel for `hire_price` at `n = seats`, capped at
`seat_hardcap.count − seats`. New seats start **unassigned**: they produce nothing until a
reallocation assigns them. Terminal rejections use only the existing categories:

| Condition | Category / reason |
|---|---|
| Unknown pool | `unknown_id` |
| Window not open (`from_gate` not crossed, or `to_gate` crossed) | `not_eligible` / `window` |
| `seats + count > seat_hardcap.count`, or `"max"` with zero headroom | `cap_exceeded` / `seat_hardcap` |
| Price not affordable, or `"max"` resolving to 0 while headroom exists | `unaffordable` |
| Malformed count | `invalid` |

When applied, the transition debits the exact bulk cost through `Ledger.Apply` (it never clamps),
sets `seats += count`, and emits `headcount_hired`.

#### S5.2 — `reallocate_headcount`

```json
{ "kind": "reallocate_headcount", "pool_id": "headcount.fixture",
  "assignment": { "role.fixture_converter": 3, "role.fixture_feeder": 0, "role.fixture_producer": 8 } }
```

`assignment` is the **complete absolute target vector** for the pool, not a delta. That makes
retries and UI sliders trivially correct, and a reordered or partial vector can never be applied
by accident.

| Condition | Category / reason |
|---|---|
| Unknown pool | `unknown_id` |
| Window not open | `not_eligible` / `window` |
| Key set ≠ the pool's role set, or a value that is negative, non-integer or unsafe | `invalid` |
| `Σ assignment > seats[pool]` | `cap_exceeded` / `seats` |
| `assignment` equals the current assignment | `invalid` / `no_change` |

When applied, the transition replaces the pool's assignment and emits `headcount_reallocated`.
Under OD-2's default it has no cost, no cooldown and no ramp.

#### S5.3 — Window closure

When a pool's `to_gate` is crossed, both intents return `not_eligible` / `window`. The existing
assignment **keeps producing** under S3. Closure freezes the allocation; it does not delete it.
This is the seam for Tier 3's "HR department" automation, which a later RFC will specify.
`tier2-content` should set `to_gate: null` until that RFC exists.

#### S5.4 — Idempotency, receipts, run log

These follow the production-engine doc without change:

- Terminal rejections are recorded receipts and consume a run-log sequence.
- Revision and idempotency conflicts are not recorded.
- An applied receipt carries the net ledger changes, resulting revision, cursor, and the complete
  authoritative snapshot, which now includes both v19 maps.

### S6 — Events

Two strict v1 payloads are added to the closed event registry. Each is committed on the same
revision as the state change it records.

```json
{ "kind": "headcount_hired", "schema_version": 1,
  "pool_id": "headcount.fixture", "count": 3, "seats_after": 12,
  "price": { "resource_id": "company.cash", "amount": "<canonical exact bulk cost>" } }

{ "kind": "headcount_reallocated", "schema_version": 1,
  "pool_id": "headcount.fixture", "seats": 12,
  "before": { "role.fixture_converter": 0, "role.fixture_feeder": 0, "role.fixture_producer": 12 },
  "after":  { "role.fixture_converter": 3, "role.fixture_feeder": 0, "role.fixture_producer": 8 } }
```

Maps serialize in raw-byte key order, and amounts are canonical Decimal strings. The canonical
corpus fixes the exact price string.

- **Rejected commands emit no event.**
- **Hires do not increment `generators_purchased_total`.** Low% semantics are unchanged; see
  Deviations D-5.

### S7 — Replay, verification and identity

1. **Replay-inputs v9** is v8 plus the resolved-union arms for `hire_headcount` and
   `reallocate_headcount`. Neither arm adds a resolved field beyond the common envelope. Price,
   window and cap all come from the pinned bundle and the restored state. Both runtimes' v8
   validators **reject** these intent kinds, so a v8 verifier can never be asked to replay a
   headcount command. v8 and earlier rows stay readable under their pinned semantics.
2. **`kernel/VERSION`** is bumped in the same commit that changes any registered path in
   `kernel/affecting-paths.json`: transition, receipt, event, snapshot or state encoding. This RFC
   touches all five. The existing drift policy applies unchanged. Runs whose pin predates the bump
   and that continue on the new build are recorded in `run_version_drift`, stay playable, and are
   excluded from boards as `engine_mismatch`. This is the existing policy, **not** a new one, and
   the owner should expect it when scheduling the deploy.
3. **Verifier.** The shared Go/TypeScript `ApplyLogged` replay covers both intents. Final-state
   comparison covers both v19 maps. The Go-authored mixed-run replay corpus gains a
   headcount-bearing run: hire, reallocate, a mid-bucket reallocation, an offline return, a
   rejected over-seat reallocation, and Exit. Every verdict mutation that applies to it (for
   example, a tampered `assigned` value) must produce `state_divergence` in both runtimes.
4. **Harness entrypoint.** The simulation-only entrypoint gains a **role effect mask**. A masked
   role's rate term and pool feed are nulled. Its seats, assignment and hire cost are kept, which
   mirrors generator effect masks. The existing Go source guard keeps masks out of the
   authoritative and replay entrypoints. There is no mask field in the catalog, save or replay
   inputs.

### S8 — Go/TypeScript parity and golden vectors

Law 3 applies: Go `Decimal` and TypeScript `break_infinity.js` 2.2.0, with canonical strings on the
wire. One shared, Go-authored corpus, `testdata/headcount-allocation.json`, is consumed
byte-for-byte by Go, Node, Chromium, Firefox and WebKit. It contains:

- **Catalog cases.** One accept case and one reject case for every rule in S1, S1.1 and S1.2.
  Reject cases include: v5 missing `headcount_pools`; an unknown effect kind; a convert chain; a
  self-converting role; two converters for one source; a cross-pool source; a Founder-scope
  resource; a zero or null seat cap; an ID collision with a generator; a dangling `pool_feed`; a
  `role` pool source without its binding; an unresolved role multiplier target.
- **Effective-heads cases (S2).** `M<E`, `M=E`, `M>E` (idle converters), `k=1` and `k=3`, zero
  source heads, all heads unassigned, and an **overflow edge**: `assigned[c] = 3.1e15`, `k = 3`,
  `assigned[s] = 9.0e15`, where the naive `assigned[c] × k` exceeds 2^53.
- **Rate cases (S3).** `produce` and `convert` terms with `all`-targeted, role-targeted and
  generator-targeted contributions (generator-targeted must be inert), huge-exponent rates, and a
  role `pool_feed` crossing the linear and log pool curves.
- **Transition cases (S5).** Every row of both rejection tables; exact and `"max"` hires; a max
  hire capped by seats and not by cash; a reallocation that splits a provision bucket; a 25-hour
  offline return with non-zero allocation; a Compute-burst window overlapping role output;
  saturation of a role's output resource at its hardcap.
- Canonical receipts, event bytes and final state for each case.

### S9 — Balance harness

1. The scenario policy vocabulary gains both intents. The existing `casual.phase0` and
   `chaos.phase0` policies and their baselines stay **byte-identical**: they run on a v4 bundle.
2. A new fixture scenario, `headcount-fixture`, runs a new `chaos.headcount` v1 policy over the S8
   fixture catalog. The policy makes seeded hire and reallocate choices, including invalid vectors,
   to exercise every rejection path.
3. A new invariant, `headcount_bounds`, joins the registry. It checks the S4 inequality, complete
   keys, and integer domain after every applied transition, and it must be `must_hold` in every
   scenario whose bundle is v5.
4. Relevance (OD-8). `hire_headcount` counts as a purchasable for the ANY/ALL relevance gate.
   Roles are measured through the S7.4 role effect mask, per role and per pool group, and reported,
   but they are **report-only**. They become gating when an allocation-aware reference persona
   exists (that is follow-up work, because the greedy reference compares purchases, not
   allocations).
5. T2 pacing targets and a T2-reaching scenario belong to `tier2-content`.

### S10 — UI surface contract

This section is normative for whichever RFC mounts the panel (OD-11). It needs these authoritative
data paths:

- **Snapshot.** Per pool: `pool_id`, window state (`not_open | open | closed`), `seats`,
  `seat_hardcap {count, reason_key}`, and `unassigned`. Per role: `role_id`, `effect.kind`,
  `assigned`, and the S3 projection row (`heads`, `resource_id`, `rate`). Converters also get
  `diverted` and `idle_converters`. All derived values are server-computed. The client displays
  them and does not re-derive them for authority.
- **Hire price.** The client predicts it with the existing TypeScript cost query against the
  pinned catalog, including `"max"`. The receipt stays authoritative.

Affordances the surface must provide:

1. **Hardcap visible (law 5).** Show `seats / seat_hardcap` with its `reason_key` tooltip. Show
   the resource hardcap of every role output the same way existing resources show theirs.
2. **Bottleneck visible.** Show idle converters and diverted source heads as numbers. Never hide
   the `min()`.
3. **Absolute-vector submit.** Edits are collected locally and sent as one `reallocate_headcount`
   carrying the full vector. Per-keystroke intents are forbidden, and the client never applies a
   predicted allocation as state.
4. **Unassigned seats are loud.** Seats that were hired but not assigned produce nothing, so the
   UI must make that visible (the Kittens idle-kitten lesson).
5. **No player-facing literals.** All text comes from copy keys that `tier2-content` authors
   (voice rules in `design/08 §1`). Components must pass `make verify-client-boundary`.
6. **Accessibility.** The panel meets the Accessibility of Player Workflows acceptance floor and
   the WCAG 2.2 AA axe gate that is already on Game UI. Every stepper and slider is keyboard- and
   screen-reader-operable.

The `game_ui_snapshot` schema bump that carries these fields belongs to the mounting RFC.

### S11 — Law compliance summary

| Law | How this RFC complies |
|---|---|
| 1 No real money | No new currency or purchase path touches money. Hires cost in-game resources. |
| 2 Server-authoritative, closed-form | Intents only. Role terms are constant between intents (S3, AC-6). No tick loop, timer or bucket is added. |
| 3 Decimal parity | One shared corpus across 5 runtimes (S8, AC-5). |
| 4 Declarative data | Rates, prices, caps, `heads_per_converter` and windows are all catalog data. No constants in code. |
| 5 Hardcaps | The seat cap is required and visible. The `min()` bottleneck is visible. Resource hardcaps are unchanged. No softcap. |
| 6 Bot fallback | Not multiplayer. Harness bots use the same validation (S9). |
| 7 Offline default | Role output accrues offline at the standard efficiency and cap. |
| 8 Versioned save | Company v19 overlay plus a migration-corpus case (S4). |
| 9 Published formulas | `production-formulas.json` gains the role term (S3). |
| 10 Curtain pull | No parodied dark pattern is introduced. |

## Deviations from design

- **D-1 — Morale is deferred** under OD-5's default. `design/01` says "Employees have morale".
  This RFC adds no morale state and no morale effect, because design gives no formula for either.
- **D-2 — Crunch events, the Soul debit and layoffs are excluded.** They are Events-Layer-1 beats
  (`design/01` T2; `design/02 §8`). This RFC leaves a clean seam for them but implements none.
- **D-3 — The production formula gains a non-generator term.** `design/02 §2.2` writes
  `Output = Σ_generators[...] × stack`. Headcount output is a second sum over roles, with the same
  stack applied to `all`-targeted contributions. The published formula artifact records this.
- **D-4 — The §11b role-differentiation law does not apply to headcount roles.** §11b is about
  generator classes. A headcount role is not a generator class and is not bought per unit. The
  equivalent guarantee here is S1 rule 7: every role has exactly one executed effect, and the
  loader rejects effectless roles. Hiring, which is purchasable, stays under the relevance gate
  (S9.4).
- **D-5 — Low% is unchanged.** Hires do not count as generator purchases. Whether a "fewest hires"
  variant should exist is a leaderboard question, not an engine one.

## Owner decisions

Each decision gives the options and the recommended default. The Specification is written against
the defaults.

1. **OD-1 — Where seats come from ("a finite hiring budget").**
   - (a) **Recommended:** a `hire_headcount` intent priced on an existing cost curve, with a
     required visible seat hardcap. The budget is finite because of the cap and because the price
     escalates.
   - (b) Seats granted per purchased unit of `category.people` generator classes, via a new
     generator role `headcount_seats`.
   - (c) A fixed per-pool budget that only upgrades raise, which would need a new upgrade effect
     kind.

   (a) reuses the cost-curve kernel wholesale and keeps hiring a separate verb. (b) couples two
   grammars and changes existing T0–T1 generator semantics.
2. **OD-2 — The cost and timing of reallocation ("re-allocation timing").**
   - (a) **Recommended:** free and instant, effective at the command instant (NGU/Kittens).
   - (b) An onboarding ramp: reassigned heads produce 0 until the next provision-grid boundary or
     for N ms. The ramp is closed-form, but its constant is invented.
   - (c) A per-reassignment resource fee.

   Under (a), timing still matters against buffs, bursts, departure before offline time, and gate
   crossings. If the harness shows that timing is decisionless, a follow-up can adopt (b) as data.
3. **OD-3 — Middle Manager semantics ("convert Engineers into Slide Decks at 1:1").**
   - (a) **Recommended:** rate diversion. Each converter redirects up to `k` source heads' output
     into the deck resource, with `diverted = min(M·k, E)`. Idle excess converters are visible.
   - (b) Stock conversion over time: Engineers permanently become decks or managers. This is a
     threshold mechanic that needs bucketing, and it overlaps T3's "self-replicate".
   - (c) Resource-to-resource consumption, where managers eat produced output. This needs negative
     accrual, which `ApplyAccrual` forbids.

   Also recommended: what Slide Decks are *for* is a `tier2-content` question. Upgrade costs and
   `resource_at_least` conditions already let any resource be spent or gated on without new engine
   work.
4. **OD-4 — Which multipliers apply to headcount output.**
   - (a) **Recommended:** `all`-targeted plus role-targeted contributions, in the standard slot
     order. Generator-targeted contributions are inert.
   - (b) None, which leaves headcount outside prestige, commons and buffs.
   - (c) Each role inherits a host generator's multipliers.
5. **OD-5 — Morale ("Employees have morale").**
   - (a) **Recommended:** defer. No morale state. Treat morale as the existing Employees
     Standing/Grievance meters, and let a follow-up RFC add modulation once design gives a formula
     ("meters modulate; facts gate").
   - (b) A new per-pool morale meter, which means amending the closed eleven-meter set.
   - (c) Morale as an event-weight input only, in Events L1.
6. **OD-6 — The effect vocabulary.**
   - (a) **Recommended:** `produce` / `convert` / `pool_feed`.
   - (b) `produce` / `convert` only, which forces Sales and Support to produce their own resources.

   Either way, the concrete jobs of Sales and Support are a `tier2-content` DESIGN-GAP. This RFC
   only fixes what content is able to express.
7. **OD-7 — An event on reallocation.**
   - (a) **Recommended:** emit `headcount_reallocated`. It keeps the audit and replay evidence
     legible, and future Enterprise "fiddling breaks SLAs" (`design/10 §1`) facts can key on it.
   - (b) No event, like manual batches.
8. **OD-8 — Relevance gating.**
   - (a) **Recommended:** hiring is gated as a purchasable; roles are report-only until an
     allocation-aware reference persona exists.
   - (b) Roles are fully gated now, which needs that persona inside this RFC's scope.
9. **OD-9 — Payroll upkeep (`design/02 §2c.1` "mandatory flow-through sinks").**
   - (a) **Recommended:** no payroll here. Hiring is a one-time sink. Record the faucet/sink budget
     in the T2 epoch changelog.
   - (b) Per-head upkeep. That is negative flow with a floor crossing (bankruptcy), which is a
     threshold mechanic that needs its own RFC under `design/06`'s bucketing clause.
10. **OD-10 — Firing.**
    - (a) **Recommended:** no fire intent. Seats only grow within a run. Unassigning to the bench
      is the only reduction. Layoffs arrive later as an Events-L1 beat with its own RFC.
    - (b) A `release_headcount` intent now.
11. **OD-11 — Where the UI panel is mounted.**
    - (a) **Recommended:** this RFC ships the data contract (snapshot fields and projection rows)
      and fixes S10 as normative. The mounted Desk panel and the `game_ui_snapshot` bump belong to
      `garage-player-surfaces` (manifest L3).
    - (b) This RFC also mounts the panel.
12. **OD-12 — What currency the seat cap mints (`design/02 §2c.2`).**
    - (a) **Recommended:** the catalog carries only `reason_key`, the same as every existing cap.
      The minted parallel currency is named in `tier2-content`'s copy and epoch changelog. Plausibly
      that currency is unassigned-seat hoarding, or seat-cap raises as a future reward.
    - (b) Add a required `mints` descriptor field to every cap type, which is a cross-catalog
      schema change.

## Acceptance criteria

Every criterion names the failing case that must be demonstrated. Gates run with `-count=1`.

1. **AC-1 Catalog v5 loads strictly in both runtimes and the schema gate.** Every S8 catalog case
   is accepted or rejected identically in Go, Node and the three browsers, and by
   `balance/economy.schema.json`, including the Decimal semantic check.
   *Failing case:* delete the "at most one converter per source" check from one loader. The
   two-converter reject vector must fail that runtime.
2. **AC-2 An empty pool list is behavior-identical to v4 (the `No Hires` baseline).** Each existing
   v4 fixture catalog is re-emitted as v5 with `headcount_pools: []`. Its receipts, events, final
   states, rate projections and formula outputs stay byte-identical across the existing economy,
   production, content and replay corpora.
   *Failing case:* an implementation that emits empty headcount maps into v18 state, or adds a
   zero role row to the projection, breaks byte-identity.
3. **AC-3 The effective-heads kernel is exact and overflow-safe.** Every S2 vector passes in all
   five runtimes.
   *Failing case:* a TypeScript kernel that computes `Math.min(assigned[c] * k, assigned[s])`
   fails the 3.1e15 × 3 overflow vector, and one that ignores `k` fails the `k = 3` vector. Both
   are demonstrated in the planning log.
4. **AC-4 The intents are complete and fail closed.** Every row of the S5.1 and S5.2 tables
   produces its exact category and reason with zero state mutation, and applied cases emit exactly
   one S6 event.
   *Failing case:* a partial `assignment` vector must return `invalid`, never a merge. An
   implementation that merges partial vectors fails the missing-key vector.
5. **AC-5 The cross-runtime golden vectors pass.** `testdata/headcount-allocation.json` passes in
   Go, Node, Chromium, Firefox and WebKit through the same entry points the production corpora use.
   *Failing case:* moving any one expected canonical string by one 12-digit ulp fails every
   runtime. This is shown by a checked-in negative-control test that mutates a loaded copy. It is
   never a lenient comparator.
6. **AC-6 Closed-form equals stepwise, and Law 2 holds.**
   - *(a) Piecewise constancy.* For every S8 transition schedule, and for every provision-grid
     boundary plus 64 seeded mid-bucket instants between consecutive commands, the S3 role
     projection rows on an Evaluate-to-t clone must equal the rows at the start of the interval,
     byte for byte.
   - *(b) Lazy equals stepwise.* Each schedule is executed twice. The lazy run evaluates only at
     command instants. The stepwise run inserts no-op evaluations every S ms, for
     S ∈ {1,000; 60,000 (grid-aligned); 7,919 (mid-bucket prime)}.
     - On the **exact corpus**, where rates and durations are chosen so every segment product is
       exact at 12 significant digits, both runs must produce identical ledger, provision and
       headcount state.
     - On a **seeded general corpus** of at least 200 schedules, headcount state must be identical.
       Balances must agree with each other, and with a test-only exact rational integral of
       Σ heads·rate·dt, within an error bound. That bound is derived analytically, **before the
       test is written**, from the numeric core's documented quantization points, and the
       derivation is recorded in the planning log. It may never be widened to make a run pass.
   - *Failing cases (all three must be demonstrated):* (i) applying the new allocation from the
     next grid boundary instead of the command instant; (ii) applying it retroactively from the
     previous evaluation; (iii) a test-only time-dependent role rate. Each must fail (a) or (b).
7. **AC-7 Save v19 activates correctly and restores strictly.** v19 round-trips. Missing, extra,
   negative or unsafe keys are rejected, and so is any breach of the S4 inequality. Founder, Guild
   and World scopes reject the fields. The v18→v19 corpus case passes, and the baseline manifest
   count ratchets. An Exit from a v4-pinned run into a v5 epoch yields next-run Company v19 genesis
   with zero maps.
   *Failing case:* a state with `Σ assigned = seats + 1` must be rejected on restore, not clamped.
8. **AC-8 Replay, verification and kernel identity are fail-closed.**
   - The headcount mixed run returns `verified` in Go and TypeScript, and each applicable one-field
     mutation returns its verdict in both runtimes.
   - A v8 envelope carrying `hire_headcount` is rejected by both runtimes.
   - `kernel/VERSION` is bumped. The affecting-paths history gate passes, and it demonstrably fails
     on a scratch commit that touches the transition without the bump.
9. **AC-9 The harness is covered.** `headcount-fixture` runs with the `headcount_bounds` invariant
   and must reach every rejection category at least once. The existing Phase-0 baselines stay
   byte-identical under the baseline change guard. The role effect mask activates only through the
   simulation entrypoint, and the source guard rejects any authoritative caller.
   *Failing case:* an injected state that breaks the seat inequality trips `headcount_bounds` and
   fails the run.
10. **AC-10 The formula artifact matches the code.** `make formulas-check` covers the role-rate
    authority, and a regenerated `production-formulas.json` lands in its own reviewed commit.
    *Failing case:* editing the role term without regenerating the artifact fails the check.
11. **AC-11 The UI data contract is present.** The authoritative snapshot and rate projection
    expose every S10 field, and a contract test asserts each one.
    *Failing case:* dropping `idle_converters` from the snapshot fails the contract test. The
    mounted-panel checks (axe, keyboard operation, no literals) belong to the OD-11 owner.
12. **AC-12 Docs and the repository are updated.**
    - `docs/economy-kernel.md`, `docs/production-engine.md`, `docs/purchasable-content.md` (for
      the pool source kind), `docs/save-layer.md`, `docs/leaderboards-and-epochs.md` (replay-inputs
      v9) and `docs/balance-harness.md` describe the shipped behavior.
    - A new `docs/headcount-allocation.md` is added and indexed in `docs/README.md`.
    - `make verify` and the Postgres suite (`docs/save-layer.md`) are green with `-count=1`.
    - `docs/economy-kernel.md` still says the active artifact is "version 3" while
      `balance/catalogs/phase0.json` is v4. That stale line is corrected in the same docs change.

Archival additionally requires the cross-party designated review verdict required by `CLAUDE.md`,
covering the full implementation range.

## Open questions

Every open question is an Owner Decision above. Items deferred to named future work:

- the Tier 3 HR automation and self-replicating Middle Managers — a future T3 RFC, using the S5.3
  seam;
- morale modulation (OD-5), crunch, Soul debit and layoffs — Events Layer 1 and its follow-ups;
- an allocation-aware relevance reference persona (OD-8) — a Relevance Harness follow-up;
- the `No Hires` challenge runtime — the future challenge-runs RFC. AC-2 already guarantees the
  engine behavior that challenge will rely on.

## Changelog

- 2026-09-25: created as a draft (Claude, for Marco). Split from `tier2-content` as the manifest's
  G05 proposed. Engine only.

## Owner acceptance (2026-09-25)

Marco accepted this RFC in the 2026-09-25 batch (the answer to a direct batch question in-session,
recorded by Claude). Every owner decision this RFC lists is **ruled at its stated recommended
default**. Where the text offers alternatives, the recommended option is the normative one and the
alternatives are rejected. Provisional numbers stay provisional and are ratified by harness
measurement and SHA, as the RFC already requires. Owner-authored copy stays owner-authored:
implementation ships copy keys, with any placeholder or candidate text clearly marked.
**Exception:** OD-1 (source of the seats) is NOT ruled: it conflicts with Tier 2 Content OD-1 and is held for an owner explanation; implement nothing that depends on the seat source until it is ruled.
