# RFC: T0–T1 Playable Content (the first hour)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/01 §Tier 0–1` (Sole Proprietor 1995 / Garage 2000s, era beats), `design/02 §2, §11` (curves, pacing targets: scripted failure ~15 min, first elective Exit [45,90] min), `design/03` (T0–1 arcade toys free, staggered unlocks), `design/08` (voice rules, era presentation), UX docs (first-session narrative)
- **Depends on:** Purchasable Content Foundation and Relevance Harness (both archived); epoch 6
  `First Content` is the byte-preserved baseline for this owner-gated next-epoch candidate.
- **Closes:** the oldest line in `rfc/README.md`'s remaining-contracts list
- **Planning:** `planning/t0-t1-content/` (once implementing)

## Summary

Epoch 6 has a live permits-gated ladder and the original Phase-0 economy rows. This RFC turns
tiers 0–1 into a playable,
funny, correctly-paced first hour: the full T0–T1 catalog (generators, upgrades, gates, manual
actions), the first-session script, and the copy — all as data on existing schemas, gated by the
existing harness. One owner-gated content epoch mints the complete candidate only after its
mandatory pacing and relevance reports pass. (Correction 2026-08-03: the original draft claimed
"no new engine code" and "16 milestones" — both false; the Foundation RFC owns the engine work,
and HEAD had 4 milestones × 4 persona-runs = 16 observations. The milestone set grows with this
RFC's content.)

## Specification

### T0 — BLOCKING PRECONDITION (F1 from the Foundation review, 2026-08-03)

The mint activates `provision_tick_ms`, and online-mode `Evaluate` hard-fails past the offline-cap
horizon (`engine.go:153`) while the live service evaluates ONLY `ModeOnline`. A founder idle
> `accrual_cap_ms` (24h) bricks their stream permanently the moment provisioning is live. **This
RFC may not mint until online evaluation drives offline catchup at session boundaries OR clamps
the online horizon** — the fix lands here (or in session-bootstrap) as AC0.

**Arm RULED (owner-side, 2026-08-08): session-boundary offline catchup, owned by Account &
Session Bootstrap.** When a session opens against a stream whose cursor lies past the online
horizon, bootstrap FIRST runs the standard offline catchup (the canonical 90%/24h-cap policy —
`docs/production-engine.md`) up to now, THEN online evaluation proceeds from the caught-up
cursor. The clamp arm is REJECTED: silently clamping the online horizon caps accrual without the
offline policy's visible, published discount — an invisible cap, which the design forbids. AC0
lands in session-bootstrap with a fixture proving a > 24h-idle founder resumes un-bricked with
exactly the offline-policy accrual.

### T1 — Catalog content (the mint)

- **Tier 0 (Sole Proprietor, 1995):** manual action `manual.click` re-skinned ("Reply to a
  Customer"); generators: `beige_tower` (exists — becomes real), `dot_matrix_queue`,
  `answering_machine`, `nephew_intern` (4 generator classes, cost curves in the 1.07–1.13 band
  per design/02, staggered milestone ladders per §11b). Upgrades: 8–12, each with ≥1 non-production-rate
  role through the shipped vocabulary (`provision`, `synergy_feed`, `manual_output`,
  `stock_rate`). Capacity and minigame-input roles remain successor mechanics.
- **Tier 1 (Garage, 2000s):** `gate.t0_to_t1` gains real requirements; generators:
  `garage_rack`, `crt_wall`, `first_hire`, `beige_tower_v2` (chain-provisioning per the §11b
  purchased/generated split); the cosmetic shop appears with `Horse Armor (Free)` as its first
  item (design/01's beat, cosmetic-only, satire copy).
- Era-authentic **event copy** for the existing event kinds (threshold crossings, gate
  crossings, offers) — Layer-1 authored events are OUT of scope (the events-engine RFC owns the
  evaluator; this RFC writes only copy for events the engine already emits).
- All values are provisional balance data; the epoch mint requires a REAL relevance report over
  the complete candidate bundle.

### T2 — The first session script

Deterministic sequence on existing machinery: boot → one manual click pays → first generator
affordable ≤ 30 s → first upgrade ≤ 2 min → `gate.t0_to_t1` crossable at ~8–10 min (chaos
persona) → **the scripted first failure fires ~15 min** (the candidate defines its lazy logged
transition and copy binding; implementation remains acceptance debt) → this RFC writes its
narrative copy — the wind-down screen, the "run 2 opens with" beat) → run 2 reaches Tier 1
faster (D6 assembly working as designed) → first elective Exit lands in [45,90] min (the
existing harness gate now measures REAL content instead of placeholder values — AC1).

### T3 — Copy discipline

Every player-facing string through the flavor bible voice rules; every real-world statistic
carries verified provenance in the claim registry; anything not yet verified is flagged, not
shipped. The copy lands as catalog/copy-system data (the content pipeline design), reviewable
like any diff.

**Carried debt (PT-C4, 2026-08-07):** generator presentation copy has NO binding surface in the
economy grammar (rows carry no copy field, and the Copy Pipeline forbids deriving display text
from mechanical IDs) — this RFC owns defining how generators become player-facing (either the
generator-copy grammar extension PT-C4 declined to add, or a presentation-layer binding). The
first two consumers waiting on it: `generator.beige_tower` (shipped, presentation-less) and
`generator.legal_dept` (Permits RFC; its title/description explicitly carried here).

## Acceptance criteria

0. **F1 precondition:** a founder idle beyond the offline cap under a provisioning catalog
   evaluates successfully (no online-horizon brick) — the mint-blocking regression.
1. The pacing harness passes on REAL content: the grown T0–T1 milestone set's distributions within design/02 §11
   targets; the [45,90] elective-Exit gate green with the T0–T1 catalog (not fixture values).
2. Mandatory Relevance report: zero dead purchasables in the T0–T1 window; every
   generator class shows a non-production role activation.
3. Every upgrade/generator/gate loads through strict loaders; the mint follows the epoch
   protocol; golden reports regenerate.
4. First-session fixture: a scripted persona completes the T2 sequence against the composed
   gameserver (the composition integration harness gains one content-driven run).
5. Copy audit: no 🔴-flagged names, no unverified statistics (grep-able provenance tags in the
   copy source).

## Ruled candidate scope

- Exact mechanical IDs and provisional literals travel through the draft-review-owner-ratify
  lane before mint.
- `Horse Armor (Free)` is a non-purchasable, non-stateful shelf stub in v1. A stateful cosmetic
  requires the successor cosmetics foundation.

## Changelog

- 2026-08-03: created (draft) — the oldest remaining contract, now the critical path to a game.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.

## Codex acceptance-review blockers (2026-08-10 — T01-C1–T01-C9)

The Foundation and Relevance dependencies now exist, but this draft remains directional content
rather than byte-complete catalog input. The mint cannot be implemented without the following
owner contracts.

### T01-C1 — Dependency/status claims are stale

Purchasable Content Foundation and Relevance Harness are archived and epoch 6 is live. This RFC
still calls both drafts, treats relevance as optional, calls the economy rows placeholders, and
speaks of an unnamed next epoch.

**Proposed contract:** rebase the body on epoch 6 and make the relevance report/content gate
mandatory. Name this change as an owner-gated epoch-7 candidate (unless another epoch lands first)
whose complete artifact list and accepted-hash protocol are filled at mint review time.

### T01-C2 — No literal economy catalog exists

T1 provides ranges and counts, not exact generator/upgrade/manual/gate rows. Costs, ratios,
effects, requirements, role bindings, ladders, copy keys, chain edges, synergy pools, provisioning
caps/reasons, and insertion order are all absent.

**Proposed contract:** append the complete strict schema-v4 candidate documents byte-for-byte:
every Phase-A generator and upgrade row, milestone ladder, synergy pool, manual action, gate/route
change, and policy literal. Provisional balance numbers are still literals and must pass both
loaders before status becomes accepted.

### T01-C3 — The promised role set names mechanics that do not exist

The shipped Phase-A role vocabulary activates only executed `provision`, `synergy_feed`,
`manual_output`, and `stock_rate` bindings. T1 requires capacity and minigame-token roles, which
the Foundation deliberately deferred until their owners existed. Permits being a resource does
not create a capacity effect, and Minigame Platform does not create a generator payout hook.

**Proposed contract:** use only the four shipped executable roles in epoch 7, with at least one
non-neutral activation fixture per declared role. Capacity/minigame_input rows require their own
mechanic RFC and are excluded from this mint; alternatively this RFC must specify those new effect
arms, replay inputs, save state, and cross-runtime hooks.

### T01-C4 — Generator/manual/cosmetic presentation has no binding grammar

The carried PT-C4 debt is still open. Generator rows have no copy key, manual action display copy
has no declared reference site, and Horse Armor has neither a cosmetic catalog nor a ruled stub
wire. Orphan Copy entries cannot satisfy the UI law.

**Proposed contract:** add one presentation artifact keyed by mechanical generator, upgrade,
manual-action, and cosmetic-stub IDs, with exact Copy-key families and strict loader/reference
validation. The Horse Armor row is explicitly non-purchasable/non-stateful in v1; if it mutates
state, it belongs to a cosmetics foundation instead.

### T01-C5 — The scripted failure is not an executable event contract

“Fires ~15 min” does not name a condition, authoritative transition, event kind/payload, whether
the run ends, or how it avoids firing twice. Existing first-Exit curriculum checks do not create a
time-triggered failure by themselves.

**Proposed contract:** enumerate the exact lazy-evaluated trigger over attended/progress state, its
one-shot persisted marker, transition outcome, event ordering, terminal/non-terminal behavior,
and `run_ended`/next-run assembly bytes. No background timer: the first eligible command evaluates
the trigger through `ApplyLogged`.

### T01-C6 — Session-boundary offline catchup lacks transaction and replay semantics

The arm is directionally ruled, but “session opens” is not tied to a composed handler, lock order,
catalog pin, intent/log row, or idempotency rule. Running a state mutation during authentication
without a replay record would create a second transition path.

**Proposed contract:** name the exact bootstrap operation and route catchup through the existing
Company logged boundary with a server-authored command and frozen `{opened_at_ms, offline_span}`
inputs. It locks the Company stream after authentication, uses the run-pinned offline policy,
returns the committed revision/snapshot, and is idempotent for one session-open token. A >24h
fixture proves exact discounted accrual, replay parity, and an Exit-race ordering.

### T01-C7 — The first-hour script needs exact milestone and harness rows

Approximate prose times do not identify which catalog facts prove “first upgrade,” “scripted
failure,” “run 2 faster,” or “elective Exit,” nor the persona seeds/reducer/envelopes. Relevance
windows likewise cannot infer the new purchasable set.

**Proposed contract:** append the literal harness scenario, milestone predicates, persona runs,
reducer, time envelopes, relevance windows, and transition budgets. The composed human-path
fixture references those same IDs, so pacing and UI cannot silently define different scripts.

### T01-C8 — T0/T1 gate and era activation are not pinned

The draft says `gate.t0_to_t1 gains real requirements` without literal requirement rows or stating
how a live epoch-6 run crosses into the new catalog. Theme era is not a constants artifact and must
follow authoritative tier rather than the deployment's current content.

**Proposed contract:** include the exact gate requirement/route row and new-run-bound activation
tests: epoch-6 runs finish under their pins, an Exit/fresh genesis under the new epoch receives the
full catalog, and tier 0/1 facts select `era_1995`/`era_2000` through Game UI's ruled mapping.

### T01-C9 — “Event copy for existing kinds” is not a closed set

No event-kind-to-copy-key rows or parameter grammar are listed, so completeness and provenance
cannot be checked.

**Proposed contract:** enumerate every v1 event-copy binding with exact allowed parameters and
fallback behavior. Only kinds already emitted by production may appear; Layer-1 authored events
remain excluded and unknown kinds fail load rather than deriving prose from an ID.

## Owner rulings on T01-C1–T01-C9 (2026-08-10)

- **T01-C1 — accepted.** The body rebases on epoch 6; the relevance report + content gate are
  MANDATORY; this mint is the owner-gated EPOCH-7 candidate (or the next free epoch number at
  mint review time), full artifact list + accepted-hash protocol filled then.
- **T01-C2 — accepted, via the established draft-ratify lane:** Codex drafts the complete strict
  schema-v4 candidate documents (every Phase-A generator/upgrade/manual/gate/route row, ladders,
  synergy pools, policy literals) from T1's ranges + design/02's formula shapes, loader-validated,
  provenance-annotated per value (the mint-content-rows pattern); Claude reviews; Marco ratifies
  by SHA. Provisional numbers are still literals; both loaders green before acceptance.
- **T01-C3 — accepted, shipped-roles arm.** Epoch 7 uses ONLY the four executable roles
  (`provision`, `synergy_feed`, `manual_output`, `stock_rate`), one non-neutral activation
  fixture per declared role. Capacity and minigame_input roles are EXCLUDED — each needs its own
  mechanic RFC (consistent with the Permits capacity deferral).
- **T01-C4 — accepted.** One presentation artifact keyed by mechanical IDs with exact copy-key
  families and strict loader/reference validation — this CLOSES the carried PT-C4/GU-C6 debt.
  Horse Armor is explicitly non-purchasable and non-stateful in v1; any stateful cosmetic belongs
  to a cosmetics foundation.
- **T01-C5 — accepted as proposed.** The scripted first failure is a lazy-evaluated trigger over
  attended/progress state with a one-shot persisted marker, evaluated on the first eligible
  command through `ApplyLogged`; exact transition outcome, event ordering, and next-run assembly
  bytes enumerated in the candidate round. No background timer.
- **T01-C6 — accepted as proposed.** Session-boundary catchup is a server-authored Company logged
  command with frozen `{opened_at_ms, offline_span}` inputs, run-pinned offline policy, locked
  after authentication, idempotent per session-open token; the >24h fixture proves discounted
  accrual + replay parity + the Exit-race ordering. (The executable form of the ruled T0 arm.)
- **T01-C7 — accepted.** Literal harness scenario + milestone predicates + persona seeds/reducer/
  envelopes/relevance windows/transition budgets, authored in the candidate round; the composed
  human-path fixture references the SAME IDs (one script, not two).
- **T01-C8 — accepted.** The literal `gate.t0_to_t1` requirement/route row lands with the T01-C2
  candidates; new-run-bound activation tests exactly as proposed; era follows the authoritative
  tier fact through GU-C7's ruled mapping (never deployment content).
- **T01-C9 — accepted as proposed.** The closed v1 event-copy binding set: only kinds production
  already emits; exact parameters; unknown kinds fail load; Layer-1 authored events excluded.

## Changelog (rulings round)

- 2026-08-10: T01-C1–C9 ALL RULED — epoch-7 candidate framing; draft-ratify lane for the full
  catalog; shipped-roles-only; the presentation artifact closing PT-C4; lazy scripted failure;
  logged bootstrap catchup; single-script harness rows; literal gate + era activation; closed
  event-copy set. The content candidate round is Codex-draftable NOW.

## Candidate-round blocker (2026-08-10 — T01-C10)

### T01-C10 — Relevance cannot encode a pre-first-gate availability window

The mandatory Relevance artifact requires every purchasable to declare
`availability_window.from_gate` as a concrete Routes gate. T0 generators and upgrades are
available from run genesis, before `gate.t0_to_t1`; using that gate as their `from_gate` would
exclude the exact T0 segment the report must measure. The economy window grammar already models
this correctly with `from_gate: null`, but the Relevance loader rejects null and has no
`run_start` sentinel. No literal candidate can be both loadable and semantically true.

**Proposed contract:** relevance-policy schema v2 changes `availability_window.from_gate` to
`null | gate_id`, where null means run genesis and sorts before every declared gate. Scenario
segments receive the same nullable `from_gate`. Validation compares boundaries over the ordered
domain `[run_start, gates...]`; `to_gate` remains exclusive and must follow the start boundary.
Reports preserve null byte-for-byte. Add Go/schema mutation tests, a T0 policy fixture whose
window is `{from_gate:null,to_gate:"gate.t0_to_t1"}`, and a report proof that a T0 purchasable is
evaluated in that segment. Existing schema-v1 artifacts and reports retain their current bytes
and meaning.

## Owner ruling on T01-C10 (2026-08-10)

- **T01-C10 — accepted as proposed.** Relevance-policy schema v2: `availability_window.from_gate`
  becomes `null | gate_id`, null = run genesis, sorting before every declared gate; scenario
  segments get the same nullable form; validation over the ordered domain `[run_start, gates…]`;
  `to_gate` stays exclusive; reports preserve null byte-for-byte; schema-v1 artifacts/reports
  retain their bytes and meaning. The Go/schema mutation tests, the
  `{from_gate: null, to_gate: "gate.t0_to_t1"}` T0 policy fixture, and the report proof land with
  the change. (The grammar the economy window already has, extended to the one loader that
  lacked it — correctly refused rather than mis-encoded.)

## T01-C10 implementation note (2026-08-10)

The schema-v2 compatibility boundary is implemented in Go, TypeScript, and the three JSON
Schemas. Schema v1 continues to reject a null `from_gate`; schema v2 treats it as run genesis,
preserves it in report bytes, and remains byte-compatible with every v1 policy, scenario, and
golden report. This implements only the ruled boundary grammar; the candidate policy remains
subject to the candidate-round review and owner ratification.

## Candidate-remediation blocker (2026-08-10 — T01-C11)

### T01-C11 — One optimization segment cannot cover the candidate's disjoint windows

T01-C10 makes run-genesis windows truthful, but the scenario loader still requires exactly one
segment for its sole optimization milestone. The candidate contains three honest, disjoint
availability intervals: T0 rows from run genesis to `gate.t0_to_t1`; T1 rows from
`gate.t0_to_t1` to `gate.t2_to_t3`; and the retained Legal Department row from
`gate.t2_to_t3` to `gate.t3_to_t4`. One segment cannot bind all three without falsely extending
at least two windows. The candidate Relevance artifact therefore remains unauthorable even
though schema v2 itself is implemented.

**Proposed contract:** schema v2 permits multiple raw-byte-ordered, non-overlapping segments for
the same sole `optimization_milestone`, one per availability interval. Every segment must name
that milestone, have a valid strictly increasing boundary pair, and be ordered without overlap;
duplicate intervals reject. Every policy item must match at least one segment whose start lies
inside the item's availability window. The solver still evaluates one target milestone and does
not change its delta, reducer, budget, or simulation mathematics. Schema v1 retains its exact
single-segment rule and bytes. The T0–T1 candidate uses exactly:

1. `{from_gate:null,to_gate:"gate.t0_to_t1"}`;
2. `{from_gate:"gate.t0_to_t1",to_gate:"gate.t2_to_t3"}`;
3. `{from_gate:"gate.t2_to_t3",to_gate:"gate.t3_to_t4"}`.

Required proof: Go and JSON-Schema rejection fixtures for v1-multiple, unordered, overlapping,
duplicate, unknown-boundary, and unbound-item cases; one schema-v2 suite proving all three
candidate intervals bind while report/delta bytes remain single-milestone.

## Owner ruling on T01-C11 (2026-08-10)

- **T01-C11 — accepted as proposed.** Relevance schema v2 permits multiple raw-byte-ordered,
  NON-OVERLAPPING segments for the same sole `optimization_milestone` (one per honest
  availability interval; strictly increasing boundary pairs; duplicate intervals reject; every
  policy item must match ≥1 segment whose start lies inside its window). The solver's target
  milestone, delta, reducer, budget, and simulation mathematics are UNCHANGED; schema v1 keeps
  its exact single-segment rule and bytes. The candidate uses exactly the three named intervals.
  (Correctly bounced rather than falsely extending two windows — the truthful-encoding
  discipline holding again.)
