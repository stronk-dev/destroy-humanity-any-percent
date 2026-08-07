# RFC: Events Engine, Layer 1 (personal narrative events)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/09 §a` (the narrative-event model: trigger DSL, MTTH, options, chains, hidden nodes, pacing locks — the required-primitives list is normative), `design/09 §d` (hot-reloadable data files, one evaluator for all three layers)
- **Depends on:** Production Engine + Run Genesis (implemented — events evaluate INSIDE `ApplyLogged` under the replay-input rule), T0–T1 Content (draft — first event pack ships there)
- **Planning:** `planning/events-engine-layer1/` (once implementing)

## Summary

The biggest unbuilt system: the data-driven personal-event evaluator. Design/09 already specifies
the format; this RFC makes it executable under our determinism laws — which is the one place we
deliberately DIVERGE from design/09's stated shape: **no wall-clock MTTH.** Events evaluate at accrual boundaries
inside the replay boundary, with draws from the save-seeded stream, so every event is replayable
by construction.

## Specification

### E1 — The catalog (`balance/events/*.json`, strict loader, epoch artifact)

Closed event schema per design/09 §a's primitives, exactly: `{id, category, trigger, mtth,
immediate[], options[], cooldown_ms, fire_only_once, hidden}` — with:
- **Trigger DSL:** closed boolean union over COMMITTED state facts (resource comparisons, gate
  flags, meter bands, ledger-fact presence, tier, compact membership, faction). Same closed-
  predicate discipline as routes and categories; grows by RFC. No projection reads, no clocks —
  triggers see only what `ApplyLogged` sees.
- **MTTH, determinized:** `mtth_ms` with a multiplicative modifier list, but
  evaluated as a **per-accrual-boundary hazard draw from the save-seeded SplitMix64 substream**
  (the exact derivation in E4): P(fire) = 1 − exp(−elapsed_ms/mtth_effective) computed in
  fixed-point ppm (integer, both runtimes — the exp via the E4 approximation contract, golden-
  vectored). Offline elapsed uses the same attended/offline split as everything else: events
  fire on ATTENDED time only (an event queue greeting your return is Layer-2/3 territory).
- **Options:** per-option triggers, `show_as_unavailable`, `dangerous`/`fallback` flags, effects
  from a closed effect union (resource deltas via the ledger, meter shifts, flag sets,
  `trigger_event {id, delay_ms_range}` chains — delays measured in attended-ms, drawn from the
  substream). Option choice is an INTENT (`choose_event_option {event_instance, option}` — C1
  envelope, evented, replayable); an unanswered event with a `fallback` option auto-resolves at
  its declared timeout boundary.
- **Hidden logic-node events** (invisible dispatchers) and **weighted pools with the 0=nothing
  bucket**: both are just events (hidden fires immediately, no UI).
- **Category pacing locks:** `{category, min_gap_ms}` table — the cheapest anti-fatigue device;
  enforced in the evaluator, recorded in state (`last_event_at_by_category`).

### E2 — Evaluation site and replay

One hook in the existing accrual chain (position: after Commons — the C8 order grows to
Prestige → faction → Guild → Commons → **Events**, a kernel-version bump). Pending event
instances and category timestamps are save-state fields (next version, corpus fixtures).
Everything the evaluator consumes is state + catalog + the seeded stream — replay_inputs carries
NOTHING new (the draw is state-derived like offers; C4's union is untouched). Fired events emit
`event_presented {event_id, instance}`; choices emit `event_resolved` — both feed the UI and the
run log like every event kind.

### E3 — Phase-0 scope

The evaluator + schema + ONE shipped pack: the T0–T1 events the content RFC writes (8–12 events:
first-customer beats, the garage-plaque event, an early regulator knock as the pressure-meter
teaser). Layer 2 (meters) and Layer 3 (server events) are successor RFCs on the same evaluator —
E1's schema is designed so a meter is a hidden recurring event and a Layer-3 situation is a
server-authored catalog overlay; nothing here forecloses them.

### E4 — Exact determinism contracts (pre-acceptance hardening, 2026-08-07)

Applying the established patterns so the acceptance round is short:
- **The hazard draw is a pure function of immutable inputs** (the Fiscal-F10 pattern; NO mutable
  PRNG cursor): `draw_ppm = Substream(run_seed ⊕ FNV1a64(event_id) ⊕
  uint64(boundary_evaluated_through_ms), "events.hazard.v1").Bound(1_000_000)` — one draw per
  (event, accrual boundary), reproducible from the save alone. Chain delays and pool weights use
  the same construction with labels `"events.chain.v1"` / `"events.pool.v1"`. Instance IDs are
  UUIDv7-compatible from `"events.instance.v1"` + the attended coordinate (the Active-Play pattern).
- **The exp approximation is a versioned pure contract, not an assumed helper:** if a shipped
  log-domain kernel helper exists, bind it by name; otherwise the hazard uses a 64-entry
  piecewise-linear table over `x = elapsed_ms·1e6/mtth_effective ∈ [0, 16e6]` ppm, the table
  shipped as artifact bytes and golden-vectored byte-identically in both runtimes. Either way the
  approximation's identity is pinned (no float `exp` in either runtime).
- **Bounded evaluation:** at most `max_events_per_boundary` fire per accrual boundary (catalog,
  visible hardcap + reason key); chain `delay_ms_range` minimum is loader-enforced > 0 (no
  same-boundary cascades); the pending queue has a visible hardcap.

### E5 — Save/version/artifact identity

Pending instances (`{instance_id, event_id, presented_at_attended_ms, timeout_at_attended_ms}`,
raw-byte sorted) and `last_event_at_by_category` are Company save fields at **Company v19** (the
chain: v17 doctrines → v18 active-play → v19 events), activating new-run-forward under ONE pinned
`events` artifact (biconditional with floor ≥ 19; requires the v17/v18 chain). `balance/events/*.json`
compiles to that single artifact; `CatalogBundle.Events` is the sole live/replay resolver (the C37
discipline). Timeout/fallback resolution is LAZY: the first accrual boundary at/after
`timeout_at_attended_ms` auto-resolves the declared `fallback` option (attended-ms; absence pauses).

### E6 — Exact wire

`choose_event_option {intent_id, kind, expected_revision, event_instance_id, option_id}` (exact
keys). Rejections reuse the closed taxonomy: `unknown_id/event_instance`, `unknown_id/option`,
`not_eligible/option_unavailable`, `not_eligible/event_resolved`. Events registered:
`event_presented.v1 {instance_id, event_id, presented_at_attended_ms}` and `event_resolved.v1
{instance_id, event_id, option_id, resolution: chosen|fallback|timeout}` — exact payloads, Go + DB
registry, byte-compared in the sequential corpus.

## Acceptance criteria

1. Golden hazard vectors (mtth × elapsed → fire-probability ppm) byte-identical Go/TS; the
   substream isolation regression (adding the `"events"` consumer changes no other stream).
2. A full run with events in the sequential corpus: fired, chosen, chained, fallback-resolved,
   pacing-locked — replaying byte-identically in both kernels (the whole point of E2).
3. Trigger-DSL closed-union conformance suite (unknown predicate rejected at load, per arm).
4. Fatigue property: with the Phase-0 pack and pacing locks, the chaos persona sees ≤ N events
   per attended hour (N in the scenario file — the conservative-frequency lesson as a gate).
5. `fire_only_once` and cooldowns survive Exit per their declared scope (run-scoped by default,
   founder-scoped by flag — the scripted-first pattern generalized).

## Open questions

- Effect-union breadth for Phase 0 (recommend: ledger deltas, meter shifts, flags, chains ONLY —
  no generator grants until an event needs one).
- Event UI presentation (modal vs feed card) — Game-UI RFC's next screen set; the engine is
  UI-agnostic.

## Changelog

- 2026-08-03: created (draft) — design/09 §a made executable under the determinism laws;
  MTTH determinized to attended-time hazard draws.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
