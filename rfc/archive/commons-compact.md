# RFC: The Commons Compact

- **Status:** implemented
- **Author:** Marco (drafted by Claude; boundary split per Codex's 2026-07-28 review)
- **Design refs:** `design/05 §5` (the commons + the front door, as designed 2026-07-28), `design/02 §7` (Trust constituencies; derivation rule), `design/10 §1` (Open Source = heavier participation)
- **Depends on:** Save Layer (implemented — compact state is stream data), Production Engine (implemented — consumes one named slot)
- **Parent / boundary split from:** `archive/production-engine-and-intents.md`
- **Planning:** `planning/archive/commons-compact/`

## Summary

The Mutual Aid Compact: membership, the Health/Capacity computation, the Enclosure index, cohort assignment, and the **single named production slot** through which all of it reaches the economy. Codex's boundary, adopted: **production consumes the computed modifier through a generic slot and knows nothing else** — the commons package computes, the production package multiplies.

## Specification

### D1 — Membership

- `sign_compact` / `leave_compact` are intents (Production RFC contract: idempotent, evented, revision-tied). This RFC supplies the authoritative membership contract. The incorporation line-item, Open Source auto-sign behavior, and player-facing panel are owned by [Commons Onboarding & Governance](../commons-onboarding-and-governance.md), because no incorporation/faction or client-shell contract exists yet.
- Membership is company-scoped state (resets at Exit, like the contract it lives in); the founder ledger records signature history as dated facts.
- Leaving is always allowed, takes effect at the next accrual boundary, and zeroes `sᵢ` (Solidarity rebuilds from scratch on re-signing — exit from a commons must be real and always available, and re-entry must carry a real price).

### D2 — Health, Capacity, and the Enclosure index

Normative here:

- **Enclosure index `dᵢ ∈ [0,1]`** — derived **entirely from the member's own production stack**: the fraction of their multiplier stack (weight-normalized) coming from dark-pattern stages, externality-ledger-generating sources, and route-of-harm slots, evaluated at each accrual boundary. **No reports, no votes, no declared route flags.** The exact slot-weight table is balance data; the formula is published in the generated artifact now and rendered in-game by the successor.
- **Health** = weighted mean of member compliance `(1 − dᵢ)`. This foundation computes cohort and server scales. Until the guild model lands, every member uses the specified guildless substitution, so `H = 0.8·H_cohort + 0.2·H_server`; the successor wires the full `0.5·H_guild + 0.3·H_cohort + 0.2·H_server` blend.
- **Capacity** = absolute sum of tithes; drives caps and content gates only, never the buff rate.
- **The buff**: `M = 1 + 5·[0.6·f(H) + 0.4·sᵢ]`, `f(H) = ((H−0.35)/0.65)^1.5` clamped ≥ 0. `sᵢ` (personal Solidarity, `[0,1]`) is the 30-day rolling mean of personal compliance `cᵢ = clamp(tithe/target,0,1)·(1−dᵢ)`. Parameters are balance data and published in the generated artifact.

### D3 — The slot boundary

The commons package exposes exactly one value to production: `commons_modifier(member) → Decimal`, populated into the fixed named slot `commons` (Production RFC D2). Non-members: slot absent (not 1.0 — absent; the future panel therefore has no line to render). **The production package must not import the commons package or vice versa** — both depend on a shared slot-contract package only (compile-enforced).

### D4 — Cohorts

- **Server-assigned, non-elective, persistent, target size ~150** (`05 §5`). Assignment on first sign; rebalancing only on population collapse (merge, never split below floor 40) — cohort identity is the shadow of the future and must not churn.
- This server foundation owns stable cohort assignment, the cohort-scale Health term, collapse merge,
  and queryable current standing. The panel (named neighbors and co-ops), monthly tithe-dial vote,
  guild participation, and mercy presentation are owned by [Commons Onboarding & Governance](../commons-onboarding-and-governance.md).
- Alt-resistance in this foundation is non-elective, Founder-persistent assignment within the server's activity bracket. Account-age weighting awaits the account/session identity contract and is owned by the onboarding/governance successor.

### D5 — Ambient surfaces (the front door's server half)

Dispatch events for commons state transitions (Health band crossings, cascade onset, recovery) are Layer-3 server events (`09 §4`) available to the future transport, including non-member delivery. The one NPC recruiting event fires per founder per career at mid-T3 if never-signed. NPC fallback holds Health near neutral below population floor and is labeled in query state and the generated formula artifact; the successor renders that label.

## Acceptance criteria

1. `dᵢ` derivation: golden fixtures over representative production stacks (canon-heavy, ethical, mixed) produce the specified indices; **no code path reads route flags or player declarations.**
2. The buff formula matches spec across the H × sᵢ grid, including the clamp and the 40%-of-bonus Solidarity floor (at maximum loyalty, total collapse leaves ×3 of the maximum ×6 and never reduces the base game).
3. Slot boundary: commons and production packages share no imports beyond the slot contract (build-enforced); a non-member has no `commons` slot.
4. Sign/leave/re-sign: evented, idempotent, and `sᵢ` zeroed on leave. Incorporation presentation is a successor acceptance gate.
5. Cohort assignment is deterministic given server state, non-elective, and stable across runs; merge preserves standing.
6. Population invariance: Health-driven buff rates are statistically indistinguishable at simulated 200 vs 20,000 CCU (harness scenario).
7. All formulas and `dᵢ`'s exact slot-weight table are generated from executable authorities into the canonical formula artifact; the onboarding/client successor owns their in-game rendering.

## Open questions

None for this foundation. The Phase-0 source-weight table (explicitly empty), Solidarity window,
tithe band, Health controls, population controls, and convex exponent are shipped balance data.
Future value changes use the harness/balance-change process; the guild term and visible surfaces are
owned by the linked successor.

## Deviations from design

- The design specifies onboarding, the cohort panel, guild participation, and monthly direction voting
  as part of one player-facing Commons. This implementation deliberately splits those surfaces into
  [Commons Onboarding & Governance](../commons-onboarding-and-governance.md). Implementing them here
  would require inventing the still-undrafted account/incorporation and guild models and would violate
  RFC-0000's DESIGN-GAP rule. The shipped server foundation retains their required membership,
  cohort, Health, Capacity, event, and query boundaries.
- The design's account-age-weighted assignment cannot run before account/session bootstrap. The shipped
  resolver boundary accepts a server/activity bracket and makes assignment non-elective and permanent;
  the successor adds age to bracket derivation without changing cohort ownership.
- An earlier proposal had 24-hour leave and seven-day rejoin cooldowns. The accepted membership contract
  intentionally ships immediate exit/re-entry with the full Solidarity reset as its price; adding a
  cooldown later would be a behavior change requiring a follow-up RFC.

## Executable contracts

### C1 — Numeric and catalog representation

All bounded ratios use integer millionths (`ppm`, inclusive range `0..1_000_000`) in
authoritative state and event payloads. Time is exact integer milliseconds. Production values
and the exported multiplier remain RFC-0001 canonical Decimal strings. A strict
`commons.schema_version = 1` balance catalog owns every provisional value: source/slot
Enclosure weights, tithe band/default, Health blend, collapse/exponent, smoothing rates,
Solidarity window, cohort target/floor, NPC weight/compliance, and `B_max`. Unknown fields,
duplicate source IDs, non-canonical Decimals, ratios outside the closed range, or a source not
declared by the economy catalog reject the combined catalog.

### C2 — Enclosure and modifier arithmetic

For active non-commons multiplier contributions, let `F_all` be their canonical product and
`F_clean` the same product with every catalog-declared forsworn source removed. Each factor is
raised to its catalog slot weight before multiplication. The Enclosure index is
`d = clamp(1 - F_clean/F_all, 0, 1)`; no active factors means `d = 0`. This is evaluated only
from source IDs and factors already accepted by the multiplier contract. Route state and
player declarations are not inputs. Health, Solidarity, and the modifier are quantized once at
the public boundary. `x^1.5` means `x * sqrt(x)` through the Decimal implementations in both
runtimes; native JS arithmetic is forbidden on this path.

### C3 — Membership state and intent wire contracts

Company save v6 adds `compact_member`, `compact_tithe_ppm`, `compact_solidarity_ppm`, and
`compact_solidarity_samples` (canonical hourly buckets covering at most the configured 30-day
window). `sign_compact` has exact fields `intent_id,kind,expected_revision,tithe_ppm`;
`leave_compact` has only the common three fields. Signing an existing membership rejects
`already_member`; leaving a non-member rejects `not_member`; an out-of-band tithe rejects
`invalid`. Both accrue first at the old membership state, mutate at that boundary, and emit one
revision-tied event. Leaving clears every Solidarity bucket and sets Solidarity to zero.

### C4 — Projection and cohort assignment

`compact_signed` is projected idempotently by event ID. The first founder signature performs
one insert-if-absent assignment into the oldest non-closed cohort in the same server and
activity bracket with fewer than 150 founders; ties use `(created_at, cohort_id)` raw-byte
order. If none exists, it creates the next server sequence. The founder assignment is never
changed by later runs. Collapse merge is an explicit operator: cohorts below 40 may merge into
the oldest compatible cohort; standing is combined by weighted numerator/denominator sums,
never averaged from already-rounded means. Concurrent first-sign races are serialized by a
server advisory transaction lock and tested against Postgres.

### C5 — Health snapshots and AI fallback

The projection stores each member's latest `(weight_ppm, compliance_ppm)` sample and their
cumulative canonical-Decimal Capacity (the absolute sum of every idempotently projected tithe), then
aggregates numerator/denominator pairs at cohort and server scope. Empty real scopes add
the declared NPC term; for fewer than 40 real members its weight is
`max(0,40-member_count)*0.5` at compliance `0.75`, and is labeled in the query result. A
guildless member substitutes cohort Health for the 0.5 guild term. Smoothed Health advances
only over elapsed whole milliseconds with the configured asymmetric rates; no per-player or
world tick loop is introduced.

### C6 — Neutral production boundary

The `server/commons` and client Commons modules expose a pure contribution builder whose output
is either no contribution for a non-member or exactly one `multiplier.Contribution` with slot
`commons`, source `commons.member`, target `all`, and the computed factor. Production imports no
Commons package and Commons imports no production package. A build check enforces both edges;
the composition layer is responsible for supplying the already-computed snapshot.

### C7 — Events and ambient surfaces

Event schema v1 kinds are `compact_signed`, `compact_left`, `compact_sampled`,
`compact_health_band_changed`, `compact_cascade_started`, `compact_recovered`, and
`compact_recruitment_offered`. Membership events carry founder/company/run identity, tithe,
and prior/new membership. Sample events carry only canonical mechanical inputs and the
resulting ppm/Decimal outputs. The projector is at-least-once safe. Recruitment has a unique
founder key and is emitted at most once; transport/client presentation remains owned by its RFC.

### C8 — Population-invariance gate

The balance harness runs identical compliance distributions at 200 and 20,000 members using
SplitMix64 and the shipped cohort/NPC rules. The report contains integer/canonical-string data
only. The mean Health-driven modifier must differ by no more than 100 ppm and the 95% interval
must overlap; violating that bound fails `harness-check`. Capacity is reported but excluded
from the rate comparison by construction.

## Changelog

- 2026-07-28: created (draft) from the commons front-door design + Codex's boundary split.
- 2026-07-29: accepted for implementation by owner direction; C1-C8 bind the previously open
  wire, numeric, persistence, concurrency, projection, and population-invariance contracts.
- 2026-07-29: split client onboarding, faction auto-sign, guild participation, cohort-panel, and
  monthly governance surfaces into `commons-onboarding-and-governance.md`; the current RFC remains
  the complete server foundation and neutral production boundary.
- 2026-07-29: implemented and archived after full Go/Postgres/Node/browser/schema/formula/harness
  verification and a complete normative-claim review.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
