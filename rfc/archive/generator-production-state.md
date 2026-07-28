# RFC: Generator Production State

- **Status:** implemented
- **Author:** Codex
- **Created:** 2026-07-28
- **Design refs:** `design/02-economy-balancing.md §1, §2.1–2.2`,
  `design/06-tech.md §stack, §idle-math`
- **Depends on:** Economy Kernel (implemented), Save Layer & Migrations (implemented),
  Production Accrual Math (implemented)
- **Parent:** Production Engine & Intent API draft
- **Planning:** `planning/archive/generator-production-state/`

## Summary

Add the data and persistence contract that turns a priced generator class into a constant-rate
production source. Generator definitions remain balance data; owned counts and the authoritative
evaluation cursor become versioned save state. No purchase intent, clock evaluation, multiplier,
or gameplay policy is implemented here.

## Motivation

The production draft cannot be implemented safely while generator output and ownership do not
exist. Save version 1 persists only balances, and economy catalog version 1 knows how a generator
is priced but not what it produces. Those are schema gaps, not gameplay choices.

This bounded follow-up closes both gaps while keeping undefined intents and multipliers out of
code. It also preserves old catalog/save readability instead of reinterpreting historical bytes.

## Specification

### S1 — Economy catalog version 2

Current balance catalogs use `schema_version: 2`. Every generator class adds:

```json
"production": {
  "resource_id": "company.users",
  "base_rate": "1e0"
}
```

- `resource_id` references a declared resource.
- `base_rate` is a positive canonical Decimal per second.
- The price resource and output resource have the same scope. Generator ownership therefore
  belongs to one save stream and one ledger transaction boundary. Cross-scope transfers such as
  guild tithe require an explicit coordinator RFC; they are not hidden inside base production.
- Catalog version 1 remains loadable for historical save/catalog hashes. Its generator production
  definition is absent and cannot be used as a production source.
- Catalog accessors expose immutable value copies and a deterministic list of generator classes
  belonging to a scope.

### S2 — Save payload version 2

The payload at rest becomes:

```json
{
  "balances": { "company.cash": "0" },
  "generators": { "generator.example": 0 },
  "evaluated_through": "2026-07-28T08:00:00Z"
}
```

- `generators` contains exactly every catalog-v2 generator class in the stream's scope, including
  zero counts. Counts are JSON integers in `0..2^53-1`; Decimal is never used for discrete
  ownership.
- `evaluated_through` is a canonical UTC RFC3339Nano server timestamp. It means production has
  been integrated through that instant; it is not a client timestamp and does not itself grant
  accrual.
- Save encode/restore operates on one state object containing ledger, generator counts, and cursor.
  The repository validates the complete object before every atomic revision write.
- Missing, extra, cross-scope, negative, non-integer, or unsafe generator counts reject the whole
  payload. A missing/non-canonical cursor does likewise.

### S3 — Version-1 migration

Restoring save version 1 initializes every in-scope production-capable generator in its referenced
catalog to zero and uses the revision's server-authored `created_at` as `evaluated_through`.
Version 1 could not represent generator ownership, so this is lossless. The migration baseline is
an explicit restore input; unit migrations never read the wall clock. Moving a save to a different
balance-catalog hash remains a separate Balance Epoch migration concern.

The first subsequent write emits version 2. Future versions migrate one step at a time and version
numbers are never skipped.

## Deviations from design

- Cross-scope production is not implicit. The design's guild tithe will require a later
  transaction/coordinator contract because existing ledgers are intentionally single-scope.
- Multipliers, milestone thresholds, and launch rates remain outside this schema slice. Only the
  constant base-rate shape already required by design is added.

## Acceptance criteria

1. JSON Schema and strict Go loading accept catalog v2 production definitions and reject missing,
   unknown, dangling, non-positive, non-canonical, and cross-scope fields.
2. Historical catalog v1 remains loadable and is explicitly non-production-capable.
3. Save v2 round-trips balances, exact generator counts, and a canonical UTC cursor.
4. Save v1 migrates deterministically from a supplied revision timestamp; malformed v2 state and
   future versions fail closed.
5. Repository create/write/load validates and returns the complete state while retaining revision,
   concurrency, archive, and five-revision behavior.
6. Canonical economy and save docs are updated; full repository verification remains green.

## Open questions

None. Purchase semantics, multiplier slots, cursor advancement, idempotency, events, and offline
policy remain explicit gaps in the parent draft.

## Changelog

- 2026-07-28: created and accepted as the next schema dependency under owner direction to continue
  foundational production work without inventing unresolved intent mechanics.
- 2026-07-28: implementation started.
- 2026-07-28: catalog v2, save v2, migration corpus, repository integration, and canonical docs
  implemented, verified, and archived.
