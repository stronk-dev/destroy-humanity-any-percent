# RFC: Faction & Incorporation Model

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-30
- **Design refs:** `design/10 §1` (four Phase-0 factions, rule-variance law, Tier-2 incorporation moment), `design/05 §3–5` (faction interdependence cycle, Commons front door), `design/02` (production stack slots)
- **Depends on:** Production Engine (implemented), Save Layer (implemented), Commons Compact (implemented — Open Source binding), Prestige & Exits (implementing — per-run reset)
- **Unblocks:** Commons Onboarding & Governance blocker #1 (faction owner), transport `guild:*` sibling work, doctrine-intent RFC
- **Planning:** `planning/faction-incorporation/` (once implementing)

## Summary

The owner contract four RFCs have been waiting on: what a faction IS as data, when and how a run
incorporates, what persists, and how Open Source binds to the Commons. Phase 0 ships faction
*identity and its structural hooks* (compact binding, interdependence resource, assisted-adjacent
variables, catalog slots); the full per-faction rule packages (design/10's five rule-change moves)
land as balance/content on these hooks, tier by tier — this RFC makes them expressible, not final.

## Specification

### F1 — The catalog object (closed, Phase 0)

`balance/factions/phase0.json`, strict loader, joins ConstantsHashArtifacts via the epoch seed
(L2a — adding the artifact is a mint):

```json
{"factions": [{
  "id": "bootstrapper" | "vc_funded" | "open_source" | "enterprise",
  "produces": "revenue" | "hype" | "libraries" | "compliance",
  "consumes": "compliance" | "revenue" | "hype" | "libraries",
  "compact": {"auto_sign": bool, "tithe_ppm": int} | null,
  "modifier_slots": [{"slot": string, "ppm": int}],
  "incorporation_copy_key": string}]}
```

- The `produces`/`consumes` cycle is loader-validated as **exactly one Hamiltonian cycle over the
  faction set** (the Last Meadow rule: nothing consumes what it produces; every resource has
  exactly one producer and one consumer faction).
- Phase-0 values: Open Source `compact.auto_sign = true` with `tithe_ppm` above the base compact
  rate (catalog data); the other three `compact: null` (they sign or decline the ordinary line
  item — Commons Onboarding renders it). `modifier_slots` may be empty in Phase 0 — the slot
  namespace joins the production stack's existing closed slot registry, growing by RFC.

### F2 — Incorporation is an intent

- New intent `incorporate {faction_id}` (C1 envelope, idempotent, evented `incorporated`).
  Valid exactly once per run, **at Tier ≥ 2** (`design/10 §1`'s incorporation moment), rejected
  `not_eligible` before Tier 2 and `already_incorporated` after. A run may reach its end
  unincorporated (early exits) — `faction_id: null` is a legal terminal state.
- Persisted on Company state (next save version): `faction_id string|null`,
  `incorporated_at ms|null`. Resets with the run (D6 assembly: null). **Founder state is
  untouched** — faction is run identity, per-run switching is content (Realm Grinder model);
  meta-progression already lives on the Founder.
- **Open Source binding:** the `incorporate {open_source}` intent evaluation *also* applies the
  compact signature in the same transaction (the existing `sign_compact` mutation, tithe from
  `compact.tithe_ppm`) — one intent, one commit, both facts. Leaving the compact while
  incorporated Open Source is rejected `faction_bound` (the always-open door out is Wind Down —
  ending the run — never a mid-run identity flip).
- The interdependence resources (`revenue`/`hype`/`libraries`/`compliance` as tradeable stock)
  are **declared here, activated by the Guild RFC** (exchange is a guild mechanism); until then
  the faction's `produces` output accrues to a ledger resource nothing consumes — visible,
  inert, honest.

### F3 — Board variable

`faction` joins the Leaderboards variable tuple as a fourth structural variable
(`commons`, `advisor`, `glitched`, `faction: string|null`) — recorded at verification from the
run's `incorporated` event (or null). Categories may pin it (an "Open Source%" board is a query,
not new machinery). This grows L7's closed schema by RFC, as that contract requires.

### F4 — What Phase 0 does NOT ship

Per-faction rule packages (the five rule-change moves), doctrine combos, later factions
(Crypto/Regulated Utility/Consultancy), and the public guildless exchange. Each is balance/content
on F1's hooks or a successor RFC. Nothing may improvise a faction mechanic that bypasses the
catalog object.

## DESIGN-GAPs blocking acceptance

- F2 names `revenue`, `hype`, `libraries`, and `compliance`, but the Economy catalog has no
  corresponding scoped resource objects. Their mechanical IDs, Company/Guild scope, initial
  values, and visible hardcaps must be declared before the save ledger can store them.
- AC5 requires the selected faction's stock to accrue, but no formula defines the accrual input,
  rate, timing, rounding, or cap behavior. “Per `produces`” is a relationship, not an executable
  transition rule.
- F1 does not enumerate the four literal Phase-0 catalog objects, including Open Source's exact
  tithe and any registered modifier slots. “Above the base rate” cannot be loader-tested or
  golden-vector-tested. Supply those values or explicitly defer the binding behavior.

These gaps require an owner amendment. Implementing them from inference would invent economy and
balance mechanics, contrary to RFC-0000 and `AGENTS.md`.

## Acceptance criteria

1. Catalog round-trip: strict load, Hamiltonian-cycle validation rejects a seeded self-consuming
   and a seeded two-cycle fixture; artifact joins the epoch seed (guard proves mint required).
2. `incorporate` intent: full C1 conformance (idempotent replay byte-identical, typed rejections
   for pre-Tier-2, double-incorporation, unknown faction).
3. Open Source: one transaction yields `incorporated` + compact membership at the declared tithe;
   compact-leave while bound rejects `faction_bound`; the other three factions leave the ordinary
   compact line item untouched (Commons Onboarding fixtures consume this).
4. Run reset: exit → new run has `faction_id null`; re-incorporation valid at Tier 2 again.
5. Interdependence stock accrues per `produces` and is inert (no consumer path exists yet —
   asserted, not assumed).
6. Save-version migration defaults (`null`/zero) with corpus fixtures both scopes.

## Open questions

- Tier-2 gate exactness (threshold event vs gate crossing) — binds to the existing gate
  predicates; the intent checks `Tier >= 2`, nothing subtler.
- Faction modifier slot contents — balance data, harness-gated, not blocking.

## Changelog

- 2026-07-30: created (draft) — the faction owner contract Commons Onboarding blocker #1 named.
