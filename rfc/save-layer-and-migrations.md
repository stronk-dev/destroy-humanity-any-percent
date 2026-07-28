# RFC: Save Layer & Migrations

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-28
- **Design refs:** `design/06-tech.md §1.7`, `design/02-economy-balancing.md §1` (the four scopes), `design/07-roadmap.md` Phase 0
- **Research:** `design/research/tech-stack.md §1.7`, Profectus + Antimatter Dimensions source reading, `design/research/adaptive-balancing.md` (Balance Epoch), `design/research/morality-systems.md` (the `New Founder` requirement)
- **Depends on:** RFC-0001 (implemented), RFC-0002 Economy Kernel (implemented — its ledger has no persistence path by design)
- **Planning:** `planning/save-layer-and-migrations/` (on implementing)

## Summary

RFC-0002 built an authoritative in-memory ledger and **deliberately omitted persistence**: it has no public balance setter, and states that loading persisted balances "must enter through a separately validated constructor or restore path." This RFC is that path, plus the versioning and migration chain that has to exist from the first commit rather than be retrofitted.

It also closes a coupling that is currently **true by accident**: RFC-0001 defined a canonical wire grammar, and saves will contain those strings. **Nothing says the wire grammar is also the persistence grammar.** If this RFC invents its own envelope we get two grammars for the same values and a migration problem on day one.

## Motivation

Save corruption is the one class of bug an idle game cannot recover from socially. A player who loses a month of progress does not come back, and "we restored from backup" is not available when the state *is* the game.

Three specific things force this RFC now:

1. **The Economy Kernel is complete but unpersistable.** Nothing else can be built on top until state survives a restart.
2. **Migration chains must start at commit #1.** Antimatter Dimensions and Profectus both demonstrate that a save format acquires an unbroken chain of migrations from its first shipped version; there is no later moment when adding one is cheaper.
3. **`design/research/morality-systems.md` escalated a requirement that lands here:** because we are server-authoritative, **we took away the player's save folder.** A single-player idle game lets you delete a file and start over. We must provide that affordance explicitly or we have silently removed it.

**Out of scope:** leaderboard storage and the Balance Epoch's *enforcement* semantics (leaderboard RFC — D6 reserves the field only); offline progression maths (production engine RFC); anti-cheat forensics beyond the audit log's existence; client-side caching.

## Specification

### D1 — The canonical grammar is the persistence grammar

**NORMATIVE: every `Decimal` in a save is an RFC-0001 canonical wire string.** No JSON numbers, no alternate encoding, no per-field precision. Parse → quantize → re-serialize is exact and idempotent, so a save round-trip is a no-op on unchanged values.

This is not a new decision; it is the *writing down* of one that was implicit. **A save is a wire payload at rest.** Integer counts follow contract §1: exact, capped at `2^53 − 1`, and stringified where JSON-number coercion could reach them.

### D2 — Storage shape

```sql
saves (
  player_id      uuid    not null,
  scope          text    not null,   -- 'company' | 'founder' | 'world' | 'guild'
  version        int     not null,   -- save-format version, not a save counter
  revision       bigint  not null,   -- monotonic per (player_id, scope)
  state          jsonb   not null,
  constants_hash text    not null,   -- D6
  updated_at     timestamptz not null default now(),
  primary key (player_id, scope, revision)
);
```

**NORMATIVE:**
- **One row per scope, not one blob per player.** `design/02 §1` defines four scopes with different lifetimes — Company resets on Exit, Founder persists, World is server-owned, Guild is shared. A single blob would make a prestige a rewrite of data it must not touch.
- **Keep the last 5 revisions per (player, scope);** prune older on write. This is the entire disaster-recovery story and it is worth its storage.
- A thin **append-only `events`** table records purchases, prestiges, and match results — **not clicks.** Forensics and rollback, not analytics.

### D3 — The NaN guard is a persistence invariant, not a check

**NORMATIVE: a save containing a non-finite value is rejected at the persistence boundary and the transaction fails without mutating stored state.** RFC-0001 §5 already makes non-finite values invalid gameplay state; this RFC makes the database the place that cannot be lied to.

This is Profectus's rule and it is load-bearing: **once a NaN is persisted, every subsequent load re-poisons the session,** and the player's save is unrecoverable. Rejecting the write costs one failed request; accepting it costs the account.

The rejection path must **log the offending field path and the intent that produced it.** A NaN reaching this boundary is an invariant violation upstream, and a silent 500 discards the only evidence.

### D4 — Versioning and the migration chain

**NORMATIVE:**
- `version` starts at **1** and increments on any change to state *shape or semantics*. Adding an optional field with a safe default is a version bump; it is cheap, and version numbers are free.
- Migrations are an **ordered, append-only chain of pure functions** `migrate_N_to_N+1(state) -> state`. **No migration is ever edited after release** — a defective migration is fixed by appending a corrective one, because players' saves have already passed through the original.
- **Loading a save runs every migration from its stored version to current, in order.** Loading is the only path; there is no "current-version fast path" that could diverge.
- **Forward-incompatibility is explicit:** a save with `version` greater than the binary's is **refused with a clear message**, never best-effort parsed. This is the rollback case, and guessing corrupts.

**Test gate:** a corpus of real saves at every historical version, migrated to current on every CI run. The corpus only grows.

### D5 — Restore is a validated constructor

RFC-0002 K3 forbids a public balance setter. **NORMATIVE:** restore builds a ledger by (1) loading the catalog, (2) validating every persisted balance against that catalog's finite-state, minimum and hardcap invariants, and (3) constructing the ledger with those values — **rejecting the whole save if any single balance fails.**

A save that references a resource the catalog no longer defines is a **migration failure, not a silently dropped field.** Silent drops are how a player's progress disappears with no error anywhere.

### D6 — `constants_hash` (the Balance Epoch seam)

**NORMATIVE: every save row records the content hash of the balance catalog under which it was written.** This RFC stores and propagates it; it defines no policy.

`design/research/adaptive-balancing.md` proposes the **Balance Epoch**: content-addressed balance files with the hash stamped on runs, so leaderboards freeze per epoch and **a silent nerf becomes structurally impossible.** That policy belongs to the leaderboard RFC. But the field must exist *now*, because it cannot be backfilled — a save written today without it can never be attributed to an epoch later.

### D7 — `New Founder` — the affordance we removed

**NORMATIVE: an unlimited, free, always-available `New Founder` action that archives the current Founder-scope save and starts a fresh one.**

`morality-systems.md` escalated this and the reasoning is sound: a single-player idle game lets you delete the save file. **Being server-authoritative took that away**, and nothing gave it back. It matters more here than in most games because our morality is founder-scoped and the docs currently say it *persists across runs* — see the `DESIGN-GAP:` at `design/10 §5`. Whichever way that resolves, **a player must be able to start genuinely clean.**

Archive, do not delete: the old founder is retained (so "the game remembers" stays available as narrative) but is no longer the active save. **No cost, no cooldown, no confirmation friction beyond a single are-you-sure.** Charging for it, in any currency, would be a dark pattern we would have to satirise.

## Deviations from design

- `design/06 §1.7` describes `saves(player_id, version, state jsonb)` as a single row. This RFC **splits by scope** (D2) because `design/02 §1`'s four lifetimes cannot share a row without prestige rewriting Founder and World data.
- Nothing else. D1 and D6 write down couplings that were already implicit.

## Acceptance criteria

1. Round-trip: ledger → save → restore → ledger produces byte-identical canonical strings for every resource.
2. A save containing a non-finite value is rejected; stored state is unchanged; the field path is logged.
3. A save at every historical `version` migrates to current in CI; the corpus is checked in and only grows.
4. A save with a future `version` is refused with a clear message and no partial parse.
5. A save referencing an undefined resource fails as a migration error, not a silent drop.
6. Prestige rewrites only the Company row; Founder, World and Guild rows are untouched (asserted by revision numbers, not by inspection).
7. `constants_hash` is present and non-empty on every written row.
8. `New Founder` archives and resets Founder scope, is free and unlimited, and leaves the archived save readable.
9. Concurrent writes to the same `(player_id, scope)` cannot interleave into a torn state.

## Open questions

- **Save cadence** — autosave interval vs. write-on-commit. Depends on the production engine's tick model; **defaulting to write-on-commit** since the Economy Kernel already commits atomically, but confirm when that RFC lands.
- **Guild-scope concurrency.** Guild saves are written by multiple players. Optimistic revision checks are probably enough at our scale, but this is the one row where "probably" should be measured.
- **Deferred to the leaderboard RFC:** Balance Epoch policy — what a board does when the hash changes.
- **Not blocking, but adjacent:** the `design/10 §5` morality-persistence `DESIGN-GAP:`. D7 stands either way; only the *contents* of the Founder scope change.

## Changelog

- 2026-07-28: created (draft).
