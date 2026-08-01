# RFC: Gameserver Composition — the two missing owner contracts

- **Status:** implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-02
- **Design refs:** `docs/commons.md` (Health weights, sampling), `docs/transport.md`, transport RFC T6 (composition + deny-closed resolver ruling), L5c (epoch seed sync before readiness)
- **Depends on:** everything implemented — this RFC exists to close the last two contracts blocking `cmd/gameserver`
- **Planning:** `planning/gameserver-composition/`

## Summary

Two contracts Codex correctly refused to invent, plus the composition manifest that consumes
them. After this, `cmd/gameserver` is pure assembly of reviewed parts.

## Specification

### GC1 — Commons participation-weight resolver (the pre-first-sample rule)

The live accrual path needs `commons_weight_ppm` for every compact member (replay records it via
RA; the resolver is the live source). Contract:

- **Authority:** the commons projection's member record, keyed by founder. The resolver returns
  `weight_ppm` = the member's CURRENT solidarity-scaled participation weight exactly as
  `compact_sampled` events already compute it (one formula, already generated into the formula
  artifacts — the resolver reads, never recomputes).
- **Before a member's first sample** (signed this run, zero `compact_sampled` rows): weight =
  **`default_tithe_ppm`-normalized entry weight, i.e. the member's declared tithe over the
  catalog default, clamped to [minimum, maximum] band, at solidarity zero** — the same value
  their FIRST sample will compute. Deterministic, catalog-derived, no special case downstream:
  the pre-first-sample member is simply a member whose solidarity is zero, which is already what
  signing sets. (No NULL, no 0-weight limbo — a signer participates from the signing instant.)
- **Non-members:** resolver returns absent (the existing `CompactMember == (weight != nil)`
  biconditional stands).
- Fail-closed: resolver error aborts the intent (typed `internal_invariant`), never a silent
  absent.

### GC2 — World-snapshot state/revision schema

The `world` channel's `snapshot` payload (`scope: "world"`) gets its closed schema:

```json
{"v": 1,
 "world_rev": int,            // monotonic counter owned by the world aggregator, NOT a save revision
 "planet": {"depletion_ppm": int, "health_ppm": int},
 "commons": {"server_health_ppm": int, "active_founders": int, "compact_members": int},
 "population": {"online": int, "founders_total": int},
 "milestones": {"active_id": string|null, "progress_ppm": int},
 "epoch": {"epoch_id": int, "name": string}}
```

- **Producer:** one world aggregator in the gameserver, sampling the projections it composes
  (commons server-health, presence counts, milestone progress) at the `world` publish cadence
  (4 Hz policy). `world_rev` increments per published snapshot; the envelope `rev` = `world_rev`
  (satisfying the wire contract's rev binding without touching save revisions — the world scope
  has no save stream and never will).
- **Recovery:** latest-only (history size 1, already policy). All fields exact integers/ppm —
  no Decimals on the world wire; display formatting is the client's.
- Closed set; additions by RFC. Loader-validated Phase-0 literals for cadence and any caps.
- Every field maps to an existing computed source — this schema ADDS no new computation, it
  names what the aggregator reads. Milestone/planet fields read zero-values until their owning
  systems ship (declared, not stubbed: the aggregator publishes `progress_ppm: 0` with
  `active_id: null`, honest about absence).

### GC3 — The composition manifest (assembly of ruled parts)

`cmd/gameserver` composes, in dependency order, all previously ruled: config+catalog load →
epoch seed sync before readiness (L5c) → save store + production service with **real**
`guild.Service.PendingSettlements` (closing the F4 seam), prestige runtime, faction catalogs,
GC1 resolver → projections (commons, routes, guild, leaderboard queue worker + relay drivers,
sweep scheduler, session GC — every parked job gets its driver here) → account API + transport
node with Account/Commons/**Guild** memberships (guild resolver replaces deny-closed; `match:*`
stays deny-closed) → world aggregator (GC2) → readiness. Drain per T6. Every seam already has
its contract; this RFC's AC list is the composition proof.

## Acceptance criteria

1. GC1: a signer's first accrual after `sign_compact` resolves the entry weight (golden vector
   = their first `compact_sampled` value); non-member absent; resolver-error aborts typed;
   replay of a first-accrual intent reproduces byte-identically (the weight is in
   replay_inputs — RA already guarantees this; the AC proves live=entry-formula).
2. GC2: world snapshots validate against the closed schema in both runtimes; `world_rev`
   strictly monotonic across a soak; zero-value honesty asserted (null milestone, 0 progress).
3. GC3: one binary boots against real Postgres to readiness; kills cleanly through drain;
   the full integration matrix runs against the COMPOSED binary (not per-package fixtures) for:
   account create→play→exit→verify→board round-trip, guild lifecycle + clearing driver tick,
   receipt+event relay over a real socket (AC6 finally), session GC tick.
4. The F4 seam closed: `PendingSettlements` composed real, with the argument-order/empty-vs-
   batch integration proof the reviewer demanded.

## Changelog

- 2026-08-02: created — GC1/GC2 answer the two contracts Codex refused to invent (correctly);
  GC3 assembles.
- 2026-08-02: accepted for implementation by owner assignment; GC1-GC3 are normative.
- 2026-08-02: GC1-GC3 implemented; awaiting the designated independent diff review before archival.
