# RFC: Production Engine & Intent API

- **Status:** implemented
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-28
- **Design refs:** `design/02 §2` (production stack, cost curves), `design/02 §10` (the daily clock), `design/06 §idle-math` (closed-form, lazy, server-authoritative), `AGENTS.md` law 7 (offline default)
- **Research:** `design/research/tech-stack.md §1` (swarmsim closed forms, intent validation), `design/research/pacing-science.md` (progress checker), `design/research/cookie-clicker.md` (multiplier stack order)
- **Depends on:** RFC-0002 (ledger), Save Layer & Migrations (implemented), Geometric Afford Fast Path (implemented)
- **Split follow-up:** `production-accrual-math.md` (implemented closed-form numeric primitive)
- **Split follow-up:** `generator-production-state.md` (implemented catalog output, owned counts, save cursor)
- **Boundary follow-up:** `gate-predicates-and-routes.md` (gate alternatives and routes; production is read-only input)
- **Boundary follow-up:** `commons-compact.md` (commons computes one generic multiplier-slot contribution)
- **Planning:** `../../planning/archive/production-engine-and-intents/`

## Summary

The layer that makes the ledger a *game*: generator ownership, closed-form lazy production evaluated over `Δt` (never per-player server ticks), the intent API (clients send intents, never results), and offline progress. **This RFC also adopts the balance constants currently stranded in `AGENTS.md`** — the last items in the deferred-decisions register with no spec home.

## Motivation

RFC-0002 deliberately excluded "production sources, multiplier stacks, time integration, and offline progress." Nothing can accrue, and no client can buy anything, until this exists. Out of scope: WebSocket transport & fan-out (its own RFC), minigame matches, prestige (needs `design/02 §3` Exit maths — a follow-up), the pacing *values* themselves (balance data, harness-gated).

## Specification

### D1 — Closed-form lazy evaluation

- Per player: use the implemented save cursor `evaluated_through`; on any read/intent, integrate production analytically over `Δt` (swarmsim model) at full intermediate precision, commit through the ledger **once** (RFC-0002 K3 quantization). **Never tick players server-side.**
- This RFC evaluates continuous production and visible resource hardcaps. Gate/unlock threshold
  crossings remain owned by `gate-predicates-and-routes.md`; the production engine exposes evaluated
  committed state to that package and does not invent a second threshold model here.
- **Server clock only.** Client-supplied timestamps are never trusted (kills clock-rollback).

### D2 — The production stack (shape is code, values are data)

`rate(resource) = Σ_generators [count × base_rate × Π(multipliers)]` with the multiplier stack in a **documented, fixed order** (the Cookie Clicker lesson: order is observable; publish it). Multiplier sources register into named slots (upgrades, milestones at 25/50/100 owned, faction rules, commons buff, Trust modulation per `02 §7`). All declared in the balance catalog; **no formula strings in data** (RFC-0002 K2 rule extends here). C2 defines the closed union, canonical order, runtime contribution boundary, and catalog declaration.

**The slot boundary (structural, per Codex's review):** multiplier providers emit mechanical contributions into **fixed named slots**. The commons compact populates its slot through the Commons Compact RFC's computed modifier — production consumes the number and knows nothing else. **Route predicates are structurally prohibited from contributing to any production slot** — the Gate Predicates RFC owns them, and its effects touch gates only. Enforce as a compile-time package boundary (the amplitude-lock pattern from `research/adaptive-balancing.md`), not review discipline.

### D3 — The intent API (contract per Codex's 2026-07-28 review, adopted)

- **Two intents only in this RFC: `buy_generator` and `perform_manual_batch`.** `buy_upgrade`, `toggle`, and feature-specific collection are deferred until their state models exist — an intent without a data model is a name, not a contract.
- The server validates affordability from **its own** evaluated state, executes through the ledger, and returns the mutation receipt + new canonical snapshot.
- **Idempotency is per save stream: `(key, request_hash)`.** Replaying a key returns the original success or the original terminal rejection; **reusing a key with a different request hash is a typed conflict.** Retention: **30 days** (provisional — comfortably beyond any reconnect scenario, bounded for storage; owner may tune).
- `perform_manual_batch` is silently rate-clamped at the exact 25/s C1 token-bucket contract.
- Invariant checks flag impossible jumps to the audit log (forensics, not auto-bans). **The numeric fallback-reporting contract (RFC-0001 §7, routed here by the adversarial review) lands in this handler's audit sink.**
- **Events are immutable and atomically tied to the resulting save revision.** Corrections are compensating events on later revisions — **history is never deleted.** This RFC emits purchases and invariant reports, never clicks; later RFCs add their own event kinds with their state models.

### D4 — Offline progress (adopting the stranded constants)

- **Offline accrual defaults ON at 90% of online rate, capped at 24 h per absence** — moved here from `AGENTS.md` law 7; `AGENTS.md` now cites this RFC.
- Published in-game (it is already the answer to speedrun "attended time" — `05 §6`).
- Beyond the cap, time banks as **Compute Credits** — **exact integer milliseconds, never `Decimal` currency** (Codex's review; time is a count, per RFC-0001 contract §1). Required balance fields, all four: `bank_ratio` (banked ms per excess offline ms), `bank_cap_ms`, `burst_speed` (rate multiplier while spending), `burst_max_duration_ms`. **Provisional launch values, harness-gated: ratio 0.5, cap 72 h, burst ×2, max burst 4 h per activation.**
- Offline evaluation is the same closed form as D1 — there is no separate offline code path to drift.

### D5 — Progress coordinate

Ship `subProgressValue(state) → 0..1` per stage (the AD progress-checker pattern) — the harness's y-axis and telemetry's core dimension. **Coordinates are typed, tier-local definitions, never arbitrary formulas** (Codex's review): the catalog declares one of a closed kind-union per tier — `resource_log` (log-progress toward a resource threshold), `count_fraction` (owned/required exact counts), `composite` (fixed weighted sum of the former two). **Every in-scope tier (T0–T3) requires an explicit monotonic coordinate before implementation completes.** T4+ coordinates land with their tier content.

## Deviations from design

- `AGENTS.md` law 7's constants move into spec + balance data (the register's last stranded item). No other deviation.

## Acceptance criteria

1. A player absent `Δt` accrues exactly the closed-form amount (golden vectors incl. cap and
   credit-banking boundaries). Online and offline evaluation use the same primitive and agree
   when given the same efficiency; the default offline policy supplies `9e-1`, so it intentionally
   yields 90% of the online delta.
2. An unaffordable intent is rejected without mutation; an idempotent replay returns the original receipt.
3. Clock-rollback attempts produce no extra accrual (server-clock property test).
4. Multiplier stack order matches the published documentation (generated from source, per the CI formula-drift gate).
5. `subProgressValue` is monotonic under pure accrual within a stage.
6. Seeded local intent property test: 24 simulated hours × 200 seeds, zero NaN/negative/soft-lock,
   ledger balances. The 200-bot × 30-virtual-day test is owned by the future Balance Harness RFC.

## Executable contracts (resolving the acceptance blockers, 2026-07-28)

Each blocker recorded by the 9effdde review is answered below with an executable contract. The
original blocker list is preserved in the changelog; a blocker is closed only by a contract an
implementer can build and test against without inventing anything.

### C1 — Intent wire contracts

Both intents share an envelope: `intent_id` (lowercase UUIDv7 — **the intent_id IS the idempotency
key**), `kind`, `expected_revision` (the positive safe-integer company-stream revision the client
evaluated against). Decimals are canonical strings; counts and durations are exact safe integers.

```json
{"intent_id":"…","kind":"buy_generator","expected_revision":41,
 "generator_id":"generator.example","count":{"mode":"exact","value":3}}
{"intent_id":"…","kind":"perform_manual_batch","expected_revision":41,
 "action_id":"manual.click","count":7,"window_ms":280}
```

For `buy_generator`, `count.mode` ∈ {`exact`, `max`}; exact values are positive safe integers and
max means verified max-affordable (RFC-0001 §7). Manual `count` and `window_ms` are positive safe
integers. `window_ms` is retained for audit/UX but never grants tokens; authority comes only from
elapsed server time. Receipt schemas:

```json
{"intent_id":"…","outcome":"applied","applied_count":7,
 "receipt":{"changes":[{"resource_id":"company.cash","before":"0","delta":"7e0","after":"7e0"}]},
 "new_revision":42,"evaluated_at":"2026-07-28T08:00:00Z",
 "snapshot":{"balances":{…},"generators":{…},"evaluated_through":"…",
             "compute_credit_ms":0,"manual_token_milli":43000,"manual_token_refilled_at":"…"}}
{"intent_id":"…","outcome":"rejected","current_revision":41,
 "rejection":{"category":"unaffordable","detail":"generator.example"}}
```

**Typed rejection categories (closed):** `unaffordable` · `unknown_id` · `invalid` ·
`cap_exceeded` · `revision_conflict` · `idempotency_conflict` ·
`internal_invariant`. **Authoritative evaluation order, one transaction:** accrue Δt (D1) →
evaluate price/validity → apply through the ledger → persist save revision + events + idempotency
record atomically. Idempotency lookup and revision comparison happen before accrual. A rejected
intent commits no gameplay state; a deterministic terminal rejection persists only its
idempotency record. Accrual is computed on a working state and commits with the action only when
the intent applies.

**Manual-action clamp, exact:** a per-stream integer milli-token bucket persisted in company state:
`manual_token_milli` (initial/cap `50_000`) and `manual_token_refilled_at` (server-authored UTC
instant). Refill is exactly `25` milli-tokens per elapsed millisecond (25 actions/s); one applied
action costs `1_000`. On intent: `tokens = min(50_000, tokens + elapsed_ms × 25)`;
`applied = min(count, floor(tokens/1_000))`; `tokens -= applied × 1_000`. Integer overflow is
prevented by saturating at the cap before addition. Excess requested actions are silently discarded;
`applied_count` makes the clamp honest without making it an error. A zero-applied batch may still
commit ordinary time accrual but emits no click event. **Manual-action catalog object:**
`{"id":"manual.click","output":{"resource_id":"company.cash","amount_per_action":"1e0"}}`.

### C2 — The production slot contract

**Closed multiplier-slot union, canonical order (multiplication order is the published order):**
`upgrades` → `milestones` → `faction` → `doctrine` → `commons` → `trust` →
`event_buffs` → `prestige`. Within-slot combination: **product** of contributions.
`commons` and `trust` are **single-provider slots** — a second contribution is a catalog
validation error. The generator's `base_rate` is applied before this multiplier sequence and is not
a slot. Slots multiply left-to-right; within a slot, contributions multiply by source id in raw-byte
ascending order. The formula panel renders that exact order. Factors must be
positive canonical state Decimals; duplicate source ids and unknown targets/slots reject the catalog.

Catalog schema addition (`multiplier_sources`):
```json
{"id":"upgrade.faster-fans","slot":"upgrades","target":"generator.example",
 "provider":"upgrade.faster-fans"}
```
`target` is a generator id or `"all"`; `provider` identifies the later state-owning package that
may emit the runtime factor. Declaring a source does not activate it. The shared boundary package exports
exactly
`type Contribution { Slot; SourceID; Target; Factor decimal.Decimal }` and the slot-order
constant; production, commons, and future providers import **only** this package (build-enforced,
amplitude-lock pattern). Production validates every runtime contribution against its catalog
declaration before using it. The shipped Phase-0 catalog declares no active multiplier sources.

### C3 — Idempotency persistence

- **Key grammar:** `intent_id` UUIDv7, client-generated, one per logical intent.
- **Canonical request hash:** sha256 over the canonical JSON serialization (sorted keys, canonical
  decimal strings, no whitespace) of the intent **minus `intent_id`**.
- **Terminal (recorded, replayed):** `applied`, `unaffordable`, `unknown_id`, `invalid`,
  `cap_exceeded` — deterministic against the recorded revision. **Non-terminal (never recorded):**
  `revision_conflict`, `internal_invariant` — the client retries with the same key.
- **Schema:** `intent_records(stream_id, intent_id, request_hash, outcome, receipt jsonb,
  created_at, PRIMARY KEY(stream_id,intent_id))`. Applied outcomes are written **in the same
  transaction** as the ledger commit, save revision, and events. Terminal rejections are written in
  a transaction that locks and confirms the referenced stream revision but creates no save revision
  or event. Same key + different hash → `idempotency_conflict`. Expiry is exposed as a store pruning
  operation for the deployment scheduler to call with a 30-day cutoff; this RFC does not invent a
  process-global cron loop.

### C4 — The event envelope

```
events(event_id uuid pk, stream_id, revision, schema_version int, kind text,
       intent_id uuid null, constants_hash text, occurred_at timestamptz, payload jsonb)
```
- `(stream_id, revision)` references the save revision written in the same transaction — an event
  cannot exist without its revision, nor vice versa when the transition is evented.
- **Kind registry v1 (closed, append-only):** `generator_purchased` · `invariant_reported` ·
  `compensation`. Prestige, threshold, manual-click, and upgrade kinds are
  **added by their own RFCs when their state models exist** — the registry grows by RFC, never by
  implementation convenience. Unknown kinds are rejected at write.
- Per-kind payload schemas at `schema_version=1` (exact keys, canonical strings):
  - `generator_purchased`: `{generator_id, count, cost_resource_id, cost}`.
  - `invariant_reported`: `{invariant_kind, detail}` where kind is
    `afford_fallback | residual_clamp | residual_abort`.
  - `compensation`: `{compensates_event_id, reason_key}`.
  Unknown/missing keys reject before persistence.
- **Compensation contract:** kind `compensation`, payload `{compensates_event_id, reason_key}` —
  a new event on a later revision. **No history mutation exists in the API.**

### C5 — Compute Credit accrual & persistence (spend split out)

- **Save field (company scope, resets at Exit):** `compute_credit_ms` — exact integer ms.
  Save-version bump + migration with default `0`.
- Evaluation receives a server-owned mode, `online` or `offline`; it is an internal call parameter,
  never an intent field. Online mode accrues the full non-negative interval at efficiency `1e0` and
  banks nothing. Offline mode applies the partition below at `9e-1`. A future connection/actor RFC
  owns presence classification; this engine owns the policy once handed that trusted mode.
- **Partition at the 24 h boundary, exact:** `capped = min(elapsed_ms, 86_400_000)` accrues at the
  offline rate (D4); `excess = elapsed_ms − capped`; `banked = floor(excess × bank_ratio)` with
  `bank_ratio = 0.5`; `compute_credit_ms = min(compute_credit_ms + banked, bank_cap_ms)` with
  `bank_cap_ms = 259_200_000` (72 h). All integer arithmetic; no Decimal anywhere in this path.
- **Spending is the named future Compute Credit Spend RFC** (burst ×2, 4 h max activation —
  parameters reserved in the catalog now, per the constants_hash lesson). The follow-up is listed
  in the RFC index and will be drafted before implementation.

### C6 — Progress coordinate catalog data (T0–T3, concrete)

Coordinate values are company-scoped, zeroed at Exit, and must be monotone under pure accrual.
`resource_log(x; target) = log10(1 + x) / log10(1 + target)` clamped to [0,1], evaluated on
committed state. **These exact objects become checked-in catalog fixtures at implementation:**

```json
{"tier":0,"kind":"resource_log","resource":"company.cash","target":"1e3"}
{"tier":1,"kind":"composite","terms":[
  {"weight":"5e-1","kind":"count_fraction","count":"generators.total_owned","required":25},
  {"weight":"5e-1","kind":"resource_log","resource":"company.cash","target":"1e6"}]}
{"tier":2,"kind":"resource_log","resource":"company.cash","target":"1e9"}
{"tier":3,"kind":"resource_log","resource":"company.cash","target":"1e12"}
```
Targets are provisional, harness-gated. T4+ land with their tier content (unchanged).

### C7 — Numeric invariant reporting

`type InvariantSink interface { Report(InvariantReport) }` with
`InvariantReport{Kind: afford_fallback | residual_clamp | residual_abort; IntentID; Detail}`.
The production handler owns a transaction-local collecting sink; collected reports become
`invariant_reported` events (C4) in the same commit. Metrics increment only after that commit
succeeds. `residual_abort` has no gameplay commit to attach an event to, so it writes the structured
server audit sink and metric only; fabricating a save revision for a failed mutation would violate
C1/C3. RFC-0001 §7's dormant normative text becomes live here, and nowhere else.

### C8 — Chaos ownership split

Acceptance criterion 6 (200-bot × 30-day chaos) **moves to the Balance Harness RFC** by name.
Retained locally: a property test driving the two intents with a seeded random policy over
24 simulated hours × 200 seeds — zero NaN/negative/soft-lock, ledger balances, runs in tier-1 CI
budget. The harness inherits the big version; production ships the small one.

## Named follow-ups

- Compute Credit spending and UX, if split from the accrual/persistence scope above (needs
  `design/02 §9` detail).
- Prestige/Exit maths (needs `design/02 §3`'s cube-root formula plus the run-end sequence from
  `design/11`).
- `buy_upgrade`, feature collection, and toggles return only with their authoritative state models.

## Changelog

- 2026-07-28: created (draft).
- 2026-07-28: reviewed by Codex; recorded blocking schemas instead of improvising them and split
  the settled cross-runtime constant-rate accrual primitive into
  `archive/production-accrual-math.md`.
- 2026-07-28: corrected the draft acceptance criterion that had required default 90%-efficient
  offline accrual to equal 100%-efficient online accrual.
- 2026-07-28: split the settled generator output, ownership, and save-cursor contract into
  `archive/generator-production-state.md`; undefined intent and policy mechanics remain gaps here.
- 2026-07-28: generator production state follow-up implemented and archived; remaining open
  questions still block the parent intent engine.
- 2026-07-28: the eight 9effdde acceptance blockers each answered with an executable contract (C1–C8); chaos AC moved to the Balance Harness RFC by name; Compute Credit spend split to a named follow-up.
- 2026-07-28: adversarial review routed the currently-latent numeric fallback-reporting contract
  here, where the first authoritative purchase handler can provide its audit sink.
- 2026-07-28: Codex acceptance review rejected premature acceptance: reconciled adopted decisions
  with the open-question section, corrected the implemented save-cursor name, and recorded the
  remaining executable contracts as explicit DESIGN-GAPs.
- 2026-07-28: Codex re-reviewed C1–C8 at 8e8938b, resolved the remaining contradictory boundaries
  (click events, chaos ownership, exact token persistence, event payloads, and trusted online/offline
  classification), and accepted the RFC for implementation.
- 2026-07-28: implementation clarification: fixed within-slot contribution order to source-id
  raw-byte ascending so catalog order and map iteration cannot alter rounded results.
- 2026-07-28: implementation clarification: applied receipts include the canonical authoritative
  snapshot promised by D3; abort-only numeric invariants use audit+metrics because no gameplay
  revision exists to carry an event.
- 2026-07-28: implemented and archived after Go/TypeScript/schema/formula-drift and Chromium,
  Firefox, WebKit gates passed; real Postgres integration passed separately.
