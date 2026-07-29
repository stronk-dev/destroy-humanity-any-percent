# RFC: Millisecond Cursor Canonicalization

- **Status:** implemented
- **Author:** Marco (drafted by Codex from the round-2 review)
- **Created:** 2026-07-28
- **Design refs:** `design/06-tech.md §idle-math`; `design/00-vision.md` law 2
- **Research:** `planning/production-review-round2/log.md` R2
- **Amends:** `archive/save-layer-and-migrations.md` and
  `archive/production-engine-and-intents.md`
- **Planning:** `planning/archive/millisecond-cursor-canonicalization/`

## Summary

Make integer milliseconds the persistence domain for every authoritative production cursor. The
current save format accepts nanosecond phases while production and manual refill each advance by
whole milliseconds relative to their own phase. A legal phase mismatch can therefore make the
manual cursor overtake the production cursor and permanently prevent manual intents from saving.

## Motivation

The reproduced state starts with `evaluated_through = …00.1009` and
`manual_token_refilled_at = …00.1001`, which satisfies the current ordering invariant. At
`…00.1015`, production sees zero complete milliseconds while token refill sees one; the resulting
manual cursor is `…00.1011`, now above production's `…00.1009`. `save.EncodeState` correctly
rejects it. Buys that do not advance the manual cursor can still work, but every manual action
recreates the invalid state and can never repair it.

## Specification

### D1 — Exact-millisecond time domain

`evaluated_through` and `manual_token_refilled_at` are canonical UTC instants with nanoseconds
divisible by `1,000,000`. Sub-millisecond server time is outside authoritative saved state.
`CanonicalServerTime(t)` means `t.UTC().Truncate(time.Millisecond)` and is the only constructor for
persisted cursor values.

For the current save version, `EncodeState` rejects a cursor outside this domain rather than
silently changing caller state. It retains the invariant:

```text
manual_token_refilled_at <= evaluated_through
```

### D2 — Save version 4 and migration

Increment `save.CurrentVersion` to 4. The JSON field set is unchanged; the semantic invariant is
new.

- Restoring v1–v3 floors both cursor instants independently to their UTC whole millisecond before
  validating their order. Existing exact integer state is unchanged.
- Restoring v4 requires both encoded instants already be canonical whole milliseconds; a lying v4
  save with a sub-millisecond cursor is rejected.
- Encoding always writes v4. The migration corpus gains phase-matched, phase-mismatched, boundary,
  and lying-version fixtures.

No elapsed whole millisecond is invented by migration. At most each historical cursor loses its
fractional remainder, which the old evaluators could never accrue or spend.

### D3 — Cursor advancement

Both production evaluation and manual-token refill use the same rule:

```text
effective_now = CanonicalServerTime(now)
elapsed_ms = effective_now - cursor
if elapsed_ms <= 0: no advancement
otherwise: apply exact integer-ms work and set cursor = effective_now
```

They do not add a floored duration back to a phase-bearing baseline. Clock rollback still grants
nothing. Repeated calls inside one millisecond are idempotent with respect to time; a later call
retains elapsed time because the cursor remains at the last whole-millisecond boundary.

### D4 — Boundary normalization

Every new company state—stream creation, Exit reset, test/harness fixture, import, and future
account bootstrap—must initialize both cursors from the same `CanonicalServerTime(now)`. Receipt
snapshots and persisted JSON therefore expose millisecond-precision RFC3339 timestamps only.

The Balance Harness virtual clock already uses integer milliseconds and needs no alternate path.

## Deviations from design

- The archived Production RFC said adding `floor(now - cursor)` preserves a sub-millisecond
  remainder. This RFC narrows persisted time to the integer-ms domain instead. The remainder is
  still naturally represented by `now - CanonicalServerTime(now)` until the next call, but is not
  persisted as a second independent phase.

## Acceptance criteria

1. The demonstrated `100.9 ms` / `100.1 ms` cursor pair migrates to aligned whole-ms state; the
   manual intent at `101.5 ms` succeeds and the result encodes.
2. v4 encode/restore round-trips exact-ms cursors and rejects either cursor with a non-zero
   sub-millisecond component; v1–v3 migration fixtures floor both fields deterministically.
3. Property tests over cursor phases and call intervals prove the manual cursor never exceeds the
   production cursor after any legal intent sequence.
4. Same-millisecond calls, a `0.999 ms` interval, exact `1 ms`, clock rollback, and the
   `86,400,000 ms` offline boundary preserve existing accrual/token results.
5. Go, receipt JSON, and save migration fixtures all serialize the same millisecond instant; the
   accepted Balance Harness golden run remains architecture-independent.
6. `docs/save-layer.md` and `docs/production-engine.md` document version 4 and the shared cursor
   domain.

## Open questions

None. This RFC changes representation precision, not any rate or balance constant.

## Changelog

- 2026-07-28: drafted from demonstrated round-2 finding R2.
- 2026-07-29: implementation started; planning record opened.
- 2026-07-29: save v4, shared cursor advancement, migration/property coverage, canonical docs,
  and Postgres verification completed; RFC archived.
