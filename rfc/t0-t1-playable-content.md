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

## T01-C11 implementation note (2026-08-10)

The Go loader and JSON Schema now enforce the ruled version split and rejection matrix. The
review candidate uses all three exact intervals and keeps one optimization milestone, reducer,
delta family, and simulation path. Its policy/scenario hashes are recorded in
`planning/t0-t1-content/candidate-proposal.md`; they are review coordinates pending designated
review and owner ratification, not production pins.

## Owner ratification — the eight-document candidate core (Marco, 2026-08-10)

Ratified after the FINDINGS-FIRST content review, the F1–F10 remediation, and the designated
re-review's RATIFY-READY verdict (all SHAs recomputed; planning/t0-t1-content/log.md):

- categories `ff63b341ff8a7439e48cbfa7cb91dcf51089809fcbb0e6e54201965e5911b9a5`
- curriculum `17e5e0c7e8b8f7217c6063b41067af0bed41a34cc26a22e9d4ddfc00513e98d9`
- economy v4 `3b18304a2a56e06619d027f3512f671cf88ddc1da4daacd77045d0b762679ac1`
- event-copy `6413fa05f76c56797ec49e82de28ecf81f52cfa502d5b687f8d764d335a94210`
- harness-scenario `e74e271be3b844bfde411887af16de06890a9a281596d45b8ad9deb7b1a502a5`
- opportunities `63e51084863bd00da7d5a0b358f54741b0b0682d8ef25b2fc7cb3da2c77f27cb`
- presentation `70953a6dfa53794f9e1e03627f0b2ddb06abb4870550dccc608a9ca0daeba0d7`
- routes `a84cce06ae67a68817174b99cfe7191e3c2f9bf47c1c20b4ebab1704baf99cfa`

Any edit to a ratified document records a replacement hash here. **Pin #9 (relevance) follows the
T01-C11 multi-segment implementation** through the same review-then-ratify flow. Still ahead
before any epoch-7 mint: the copy-text authoring round (FCE-C7 pattern, orphan-first), the
loader/owner implementations for the four new artifact families, the T01-C6/AC0 bootstrap
catchup, the composed harness golden (post-opportunities-mint per EH-C8), and the separate
owner mint sign-off.

## Owner ratification — pins #9a/#9b (Marco, 2026-08-10): THE CANDIDATE SET IS COMPLETE

- relevance policy `f8878cbf6705581eb5ffd88ea51e3719ebf2641c661bbe3d87ff7667002d30bf`
- relevance scenario `2f1afb928e1f2d84d2e9748fbbd565bbbdebce24d7c411a0df7feb0b47692629`

Ratified after the designated RATIFY-READY verdict (planning/t0-t1-content/log.md, 2026-08-10).
The nine-document epoch-7 candidate core is COMPLETE. Remaining before any epoch-7 mint: the
screen-copy assembly ratification (the v1 ruling is issued; SHAs pending), the four new artifact
families' loaders/owners, the T01-C6/AC0 bootstrap catchup, the mandatory relevance + composed
harness runs over the complete set, and the separate owner mint sign-off.

## Mint-runway implementation blockers (Codex, 2026-08-11 — T01-C12–T01-C15)

The first normal `make t0-t1-relevance` execution and a production-owner audit found four
contracts that must be reconciled before a promotion manifest can honestly describe a mint.

### T01-C12 — The ratified relevance work budget fails before dispatch

The schema-v2 candidate declares `relevance_budget_max_transitions: 2000000`. Under the shipped
fail-before-dispatch proof, its 19 items, one reference seed, width 8, and 128 decisions require a
static ceiling of **4,529,004,752** transitions. The mandatory report therefore cannot start; the
runner correctly rejects before executing work above the declared budget. Raising the literal to
4.53 billion would make a routine mint gate operationally unusable, while silently weakening the
preflight or the ruled beam changes reviewed harness semantics.

**Proposed contract:** keep the two-million hard budget and rule a bounded beam expansion that has
an exact static ceiling below it (including the order in which children are selected before an
expensive rollout), then re-run the 5% greedy-gap proof and re-ratify the scenario hash. Alternative:
explicitly ratify the 4,529,004,752 budget and accept its cost. Implementation must not choose.

### T01-C13 — Earlier core pins conflict with the later screens ratification

The nine-document list still pins `presentation-v1` (`70953a6d…`) and `event-copy-v1`
(`6413fa05…`). The later owner-ratified screens package explicitly supersedes presentation with
`presentation-v3` (`c387402b…`) and event copy with `event-copy-v2` (`71d88ebb…`). One epoch cannot
claim both versions as the single artifact authority, and the promotion manifest cannot infer
which ratification wins.

**Proposed contract:** the explicit supersession wins: epoch 7 promotes presentation-v3 and
event-copy-v2, and this RFC's owner-pin table records those replacement hashes. Presentation-v1
and event-copy-v1 remain historical candidate inputs only.

### T01-C14 — Four candidate artifacts still lack production owners

The candidate proposal honestly records `curriculum`, `event-copy`, `presentation`, and the
first-hour harness scenario as exact proposed grammars whose production loaders/owners were not
implemented. Current source confirms that state. In particular, no curriculum loader owns the
lazy terminal trigger, `curriculum_failure.v1` is absent from the event registry, and no runner
understands the candidate milestone/invariant extensions. Minting these bytes would make them
identity baggage rather than executable content.

**Proposed contract:** land strict owners before promotion: presentation/event-copy exact-set
loaders bound to Copy and the emitted event registry; the curriculum loader plus its existing-Exit
coordinator integration and additive event grammar; and the first-hour scenario loader/runner with
the nine named invariants. Each joins `CatalogBundle`/replay identity only where its mechanic is
actually consumed. Exact curriculum receipt/event payload bytes that are not present in the
candidate document require a narrow owner ruling before code.

### T01-C15 — AC0 names a boundary but not an executable session-open operation

The repository has anonymous account bootstrap, refresh, and normal intent routes, but no
authenticated session-open command or receipt/idempotency store that could own T01-C6's frozen
`{opened_at_ms, offline_span}` transition. The accepted paragraph does not enumerate the public
operation, server-authored command kind, receipt/event bytes, or its race coordinate with Exit.

**Proposed contract:** rule that exact API/transition grammar (or bind catchup to an existing
named operation) before implementation. It must remain one replay-logged Company mutation using
the run-pinned offline policy and a claim-token/idempotency pattern already shipped; authentication
alone must never mutate the save.
