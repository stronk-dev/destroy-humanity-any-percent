# Combat Shared Data & Arithmetic — append-only log

## 2026-07-29 — acceptance and executable-surface audit

- Accepted for implementation by the ordered batch manifest and moved only the shared parent to
  implementing; child engines remain draft behind this dependency.
- C3 and C4 are executable as written. Implementation starts there: exact staged floor arithmetic,
  saturating int32 stores, the six-cycle chart, and one SplitMix64/rejection-sampling authority with
  FNV-1a-labeled independent substreams in Go and TypeScript.
- `DESIGN-GAP:` C2 calls move effects a closed union but enumerates no effect kinds or exact fields;
  lane spells likewise have no executable effect schema. A strict exact-key loader cannot be written
  until those members exist. Proposed owner amendment: enumerate each effect object and identify
  which unit/building fields apply to spells instead of using one ambiguous row shape.
- `DESIGN-GAP:` C5 assigns Trust/Obedience and Soul modulation to integer piecewise tables but gives
  no breakpoints or output values beyond an internally contradictory prose slope (“50%→30%
  disobedience across Trust 1.00→0.80”) and gives no Soul function at all. AC5 cannot be made golden
  without inventing mechanics. Proposed owner amendment: literal ordered `(input, output_ppm)` points,
  interpolation/clamp rules, composition order, and the valid Soul integer range.

## 2026-07-29 — exact arithmetic and deterministic substreams

- Moved SplitMix64 and rejection sampling from the balance harness into `server/determinism`; the
  harness now aliases that authority, so production combat and balance simulation cannot drift by
  copying the generator. Added the exact battle-seed expansion and FNV-1a-labeled substream helper.
- Implemented the six-cycle Temperament chart and staged damage in Go and TypeScript. Both paths use
  base×ATK/64 -> exact 13/10 or 10/13 chart ratio -> 3/2 critical, flooring at each declared stage,
  then minimum-one and saturating int32 storage. TypeScript intermediates are BigInt.
- Added shared schema-1 vectors covering neutral/advantage/disadvantage, critical order, minimum
  damage, zero power, saturation, battle seed, and crit/obedience/bot-policy substream draws. Property
  tests prove every Temperament has exactly two wins and losses and a newly added consumer leaves an
  existing substream byte-identical.
- Added a fail-closed client combat division guard. It self-tests against a seeded direct `/`
  expression before scanning every combat module; the only division authority is the external
  integer helper.
- Focused Go tests, strict TypeScript/Svelte checks, the guard, and the full 6,437-test Node suite are
  green. Canonical `docs/combat.md` describes only this shipped kernel and preserves the catalog/table
  gaps explicitly.
