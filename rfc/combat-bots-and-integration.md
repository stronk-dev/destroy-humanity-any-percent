# RFC: Combat — Bots, Verification & Match Integration

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-29 (split from Combat Data Model per Codex review blocker #1)
- **Design refs:** `design/05 §4` (PvP table, punch-down multiplier), `design/04 §2` (real ratings, no rubber-band — owner decision 2026-07-28)
- **Depends on:** both engine RFCs, Combat Shared Data (parent), Account & Session Bootstrap (identity), Balance Harness (report format)
- **Planning:** `planning/combat-bots-integration/` (once implementing)

## Summary

Everything around the pure engines: bot policies, server-side replay verification, match records, ratings, matchmaking payouts, and the statistical gates. This is the only combat RFC that touches identity, storage, or the network.

## Specification

### B1 — Bots (deterministic, manifest-published)

- **Duel:** expectimax depth 2 over the small action space; personality-flavored policies = published per-temperament weight tables (catalog data). Decision tie-break: lowest action index. All draws on the `"bot_policy"` substream (parent C4) — a bot's whole game is reproducible from the seed.
- **Lane:** the greedy policy family with **published behaviour-flag manifests**: `{level, flags: {defends_towers, cycles_cheapest, saves_for_combo, uses_spells_reactively, punishes_capacity_leads}}` — each level a catalog row enabling a superset of flags. Tie-break: earliest legal placement, lowest card index, lowest position.
- Bot identity in a replay: `{bot: true, manifest_level, policy_version}` in the battle log header. **Bots never cheat** — same legality path, no hidden state beyond the same snapshot a human attacker would get.

### B2 — Verification and match records

A match result is derived by **server re-simulation of the submitted inputs** (measured ~13,000× real time; cost is negligible) — never client-reported. `match_records(match_id, engine, attacker_fid, defender_fid | bot_ref, seed, inputs bytea, log_hash, winner, verified_at)`; the projection into ratings is idempotent by match_id. A re-simulation mismatch is a terminal `verification_failed` rejection (the client's log is discarded; the server's replay is truth).

### B3 — Ratings and payouts

- **Real Elo (K=32, floor 400, provisional K=64 first 10 matches), no rubber-band** (owner decision). Separate rating per engine. Bot matches rate at **half K** and never below the floor (farmable-bot protection without fake ratings).
- **Payout multiplier: 200% punching up, 100% at-tier, 5% four tiers down** (linear interpolation between, published table) — tiers = 200-Elo bands.
- Matchmaking itself (queues, expanding bands, backfill timing) is the PvP-service RFC; this RFC's boundary is: a matcher hands `(attacker, defender_snapshot | bot_manifest)` in, gets a verified match record out.

### B4 — Statistical gates (blocker #7's reproducibility contract)

All gates run from **committed generator configs**: `harness/combat-gates.json` = `{seed, roster: parent Phase-0 fixture, team_generator: "all-3v3-combinations", deck_generator: "uniform-8-of-N", matches_per_pair: 20, policy: manifest max-level}`. Reports are byte-identical golden artifacts (the balance-harness discipline; BALANCE-CHANGE: protocol applies to combat catalogs).

1. **Anti-Nash-1:** best loadout field win rate ≤ 65% and Nash support ≥ 3 over the sampled space.
2. **Bot monotonicity:** level n beats n−1 in ≥ 60% of `matches_per_pair × pairs` (exact counts in the artifact, no confidence hand-waving — the population is fixed and fully enumerated).
3. **Lane decisiveness** (lane RFC AC2) runs from the same config.

## Acceptance criteria

1. Bot determinism: same `(seed, snapshot, manifest)` → identical choice/placement lists, both runtimes (golden fixtures per manifest level).
2. Verification: a tampered input list fails `verification_failed`; an honest one round-trips to an identical log hash.
3. Rating math golden vectors (provisional K, floor, half-K bot matches, payout interpolation).
4. Gate artifacts regenerate byte-identically from `combat-gates.json`; a seeded catalog nerf changes them and trips the BALANCE-CHANGE: guard.
5. Punch-down payout table matches `design/05 §4` at the published anchor points.

## Open questions

- PvP matchmaking service (queues, bot-backfill timing, queue-depth display) — separate RFC, this one defines its boundary.
- Season/rotating-weakness mechanics — follow-up per `design/04 §2`.

## Changelog

- 2026-07-29: created from the four-way split; answers parent-review blockers #6 (bots/snapshots) and #7 (statistical gates); B3 encodes the 2026-07-28 real-ratings decision.
- 2026-08-06: `punishes_elixir_leads` renamed `punishes_capacity_leads` (lockstep with the lane-engine capacity rename); no spec change.
