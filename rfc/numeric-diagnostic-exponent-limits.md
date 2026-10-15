# RFC: Numeric Diagnostic Exponent Limits

- **Status:** draft — proposal, NOT implementation authority
- **Author:** Codex
- **Created:** 2026-10-15
- **Design refs:** `design/06-tech.md` §3, server-authoritative numeric state
- **Parent / amends:** [Numeric Core](archive/0001-numeric-core.md), diagnostic classification in [Boundary Hardening](archive/numeric-core-boundary-hardening.md)
- **Depends on:** existing state/wire guards and RP-663 conformance corrections
- **Planning:** existing `planning/clout-v1-and-pr-interns/` RP-663 evidence; implementation plan after acceptance

## Summary

Propose fail-closed Go diagnostics when a lower-domain power/exp result cannot fit
the Go exponent representation. Return NaN, never valid zero. Explicitly permit a
diagnostic-classification difference from the pinned client only for that exhausted
representation, while requiring both runtimes to reject state/wire and preserving
all existing gameplay values, quantization and representable-result contracts.
This exception is not adopted by writing this draft or by reviewing earlier fixes.

## Motivation and evidence

RP-663's int64-representable corrections preserve the pinned client's lower-domain
diagnostics. They do not solve these actually executed counterexamples:

- `1e-8999999999999999` raised to2048: Go returns valid zero; pinned Node retains
  finite raw(1,-18431999999999997952), invalid state/wire.
- `exp(-1e20)`: Go returns valid zero; pinned Node retains finite
  raw(1,-43429448190325178368), invalid state/wire.

The exponents cannot fit the specified Go int64 field. Substituting a fabricated
finite exponent could later recover to a false valid magnitude; silently extending
the representation changes the foundational contract and every Decimal's cost.
The existing promise of exact diagnostic classification does not authorize that
choice. Source, executed boundary evidence, limitations and exact corrective-review
ranges are in the [owning log](../planning/clout-v1-and-pr-interns/log.md).

## Specification proposal

1. Keep gameplay exponents exactly `[-8999999999999999,8999999999999999]` and the
   existing strict constructors, state guards, canonical grammar and Quantize12.
   No new restriction on otherwise admitted finite inputs or power arguments.
2. Preserve existing evaluation/normalization order while the required exponents
   fit int64, including in-domain results rescued by coefficient normalization or
   later arithmetic. Merely leaving the gameplay range is not exhaustion. Once
   an exponent truly exceeds int64, the proposed NaN is deliberately irreversible,
   even if a longer client-side arithmetic chain could eventually return in-domain;
   this consequence is part of the explicit exception, not full-chain equivalence.
3. When the selected lower-result path's normalized exponent cannot fit int64,
   return Go NaN. This includes negative-infinite exponent overflow reached from
   finite inputs. Check representation bounds before conversion; never depend on
   architecture-specific float-to-int conversion or integer wrap. Do not clamp a
   coefficient/exponent, emit a sentinel amount, or turn exhaustion into zero.
4. For that representation-exhausted result only, the shared contract expects Go
   NaN and the pinned client's actual diagnostic classification separately. Both
   must be non-state and unserializable as canonical gameplay wire values. A
   finite client diagnostic is not permission to persist it or label Go zero
   equivalent. Existing exact zero/domain classifications and finite tolerances
   remain unchanged everywhere outside this explicit exception.
5. Preserve genuine zero results and existing division/zero-power compatibility
   cases. NaN propagation and feature-level domain validation remain required.
   A transition reaching an invalid result refuses atomically with no partial
   ledger mutation. Do not add a silent zero fallback in a consumer.
6. Keep the client library pinned. Do not add layered numbers, a second math
   engine, extra Decimal representation fields or a wire/save schema. Landing is
   a real kernel behavior change and requires the next free kernel version.

## Deviations from existing contracts

This deliberately amends exact cross-runtime diagnostic classification beyond
Go's exponent representation; invalid-state/wire parity remains exact. It does
not change any admitted gameplay exponent, rounding rule, formula or release
criterion. The archived parent bodies remain frozen; canonical docs must describe
the explicit exception only after acceptance and implementation.

## Acceptance criteria

1. Before changing runtime code, permanent targeted regressions reproduce both
   valid-zero counterexamples above. Shared fixtures name each runtime's expected
   class explicitly; no blanket weakening of finite/zero assertions or6296 goldens.
2. Exercise the int64 lower edge, neighboring representable floating exponents,
   integer/fractional and negative-base paths, exp, negative-infinite exponent
   overflow from finite inputs, and coefficient-normalization rescue. Assert the
   exception does not swallow a representable result or genuine zero.
3. Both state/wire guards refuse exhausted results. Actual Ledger.Apply and
   ApplyAccrual refuse a mixed transaction atomically; complete prior snapshots
   and receipts remain unchanged. These are native boundaries, not claimed SQL.
4. Cold Go/Node shared numeric suites, Chromium/Firefox/WebKit, vet, strict types,
   deterministic golden regeneration and the documented fuzz lane pass. Execute
   Linux-amd64 overflow/conversion regressions as well as native checks. Run
   affected consumer populations and preserve/disclose existing red lanes; do
   not regenerate numerical research outputs or relax bounds to hide differences.
5. Update canonical docs, RP-663 and owning records in the behavior batch. Obtain
   designated normative review/adoption first, then independent complete-range
   implementation review. The corrective-review delegation does not adopt this
   exception. AC6/partition, RP-664, hosted CI, full1.0 and release gates remain.

## Open decision before acceptance

Adopt the narrow fail-closed NaN exception, or require an expanded diagnostic
representation with its compatibility/cost proof? The former is recommended:
neither result is gameplay state, and representation exhaustion must not create
a valid payment. This recommendation is not an owner ruling. Valid-zero fallback
is the observed defect, not an acceptable third option or release waiver.

## Changelog

- 2026-10-15: drafted from executed RP-663 representation-limit counterexamples;
  no product code, numeric policy, release scope or adoption changed by this file.
