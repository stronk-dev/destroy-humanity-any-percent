# RFC: Production Contract Assertions & Integrity

- **Status:** implemented
- **Author:** Codex
- **Created:** 2026-07-29
- **Design refs:** `design/00-vision.md` laws 2 and 9; `design/06-tech.md` §idle-math
- **Research:** `planning/production-review-round2/log.md` R4–R8
- **Depends on:** [Production Engine & Intent API](archive/production-engine-and-intents.md),
  [CI Baseline](scaffolding-and-ci.md)
- **Parent / amends:** [Production Engine & Intent API](archive/production-engine-and-intents.md)
  C2, C4, and C7; [CI Baseline](scaffolding-and-ci.md) D4
- **Supersedes / superseded by:** —
- **Planning:** `planning/archive/production-contract-integrity/`

## Summary

Turn production review findings R4–R8 into repository-owned contract tests and fail-closed
integrity gates before the Balance Harness records its first baseline. This follow-up does not add
gameplay, balance values, event kinds, or transport behavior. It promotes behavior already verified
in reviewer scratch tests, repairs observability and receipt-error boundaries, and makes the
published production formula artifact demonstrably dependent on the executable rate path.

## Motivation

The round-2 review found no incorrectly implemented C1–C8 contract after R1–R3 were remediated.
It did find important behavior that exists without durable assertions: multiplier validation,
invariant reporting, formula provenance, and a reliable fresh-database integration target. A
balance baseline created before those gates exist could preserve an implementation while leaving
its surrounding correctness contract easy to regress.

The scope is deliberately limited to assertions and integrity. WebSocket mapping of internal
errors, production migration coordination between deployed processes, new event kinds, and balance
changes remain with their named owning RFCs.

## Specification

### D1 — Multiplier contracts become shared regression tests

The neutral `server/multiplier` package receives direct tests for the closed slot union, the exact
and duplicate-free canonical slot order, and rejection of unknown slots. Its ordering helper must
sort contribution source ids by raw UTF-8 byte order and be the helper used by production rate
evaluation; permutations of the same valid contribution multiset must produce the same canonical
rate.

Shared Go/TypeScript catalog fixtures cover all of the following:

- duplicate source ids reject;
- a second `commons` or `trust` provider rejects, including providers with different source ids;
- slots other than `commons` and `trust` may declare multiple distinct providers;
- unknown slots and targets, and malformed provider identifiers, reject; and
- valid declarations are accepted identically in both runtimes.

Go-only runtime tests additionally cover undeclared, mismatched, duplicate, and non-positive
contributions plus rate permutation invariance. At least one ordering test uses valid ASCII ids
such as `a.a` and `a_a`, whose locale order differs from raw byte order. Any later TypeScript rate
mirror must use the same fixture and must not use `localeCompare` for this contract.

### D2 — Exported invariant boundary and outcome-aware persistence

Production exports the extension point specified by the parent RFC:

```go
type InvariantKind string

const (
    InvariantAffordFallback InvariantKind = "afford_fallback"
    InvariantResidualClamp  InvariantKind = "residual_clamp"
    InvariantResidualAbort  InvariantKind = "residual_abort"
)

type InvariantReport struct {
    Kind     InvariantKind
    IntentID string
    Detail   string
}

type InvariantSink interface {
    Report(InvariantReport)
}
```

The intent handler owns a transaction-local collecting implementation. The report's durable path
depends on whether a gameplay revision exists:

- **Applied mutation:** `afford_fallback` and `residual_clamp` become validated
  `invariant_reported` events on the new revision. Audit output and metrics occur only after the
  save, event, and intent record commit together.
- **Recorded terminal rejection:** no gameplay revision is fabricated. Any collected report,
  including an affordability fallback that ends unaffordable, reaches structured audit output and
  metrics only after the rejection receipt commits. It creates no event.
- **Internal abort:** `residual_abort` reaches structured audit output and metrics only after the
  gameplay transaction rolls back. It creates no save revision, intent record, or event.
- **Replay or idempotency/revision conflict:** no invariant event, audit record, or metric is
  duplicated.

This explicitly amends parent C7's statement that only `residual_abort` is eventless. Event
availability follows the stronger C4 invariant: an event can only belong to a real committed
revision. Observability may not create fake gameplay history.

Unit tests use a fake exported sink. Postgres integration tests prove the applied event's atomicity,
the terminal-rejection audit-only path, rollback behavior for aborts, and replay
non-duplication.

### D3 — Published formula artifact is source-bound, not source-inferred

The formula generator emits schema version 2. It retains the explicit human-readable
`production_rate`, `multiplier_slot_order`, and `within_slot_order` fields and adds a
`source_fingerprint` containing lowercase hex SHA-256.

The fingerprint is computed from canonical Go AST for these executable authorities:

1. the production rate-evaluation function;
2. the canonical multiplier slot-order declaration; and
3. the raw-byte source-order helper actually called by rate evaluation.

The generator locates those declarations by their exact package/type/symbol identity, strips
comments and source positions through AST reformatting, and hashes the resulting ordered canonical
representation. Missing, duplicate, renamed, or unparsable authorities fail generation. Formatting
and comment-only edits do not change the fingerprint; an executable AST change to any authority
does.

The raw-byte ordering rule moves behind the neutral multiplier helper used by production, so the
artifact and runtime cannot name different ordering sites. The artifact does **not** claim that an
AST tool inferred algebra from arbitrary Go. It proves that the reviewed published formula is
bound to a specific executable implementation. `make formulas-check` requires a separate,
reviewable artifact-regeneration commit after an intentional formula-path change.

Canonical documentation is narrowed from "generated from source" to this precise
source-bound/review-gated guarantee.

### D4 — Fresh-database integration tests serialize migration ownership

`make test-save-integration` runs Go integration packages with package parallelism disabled
(`go test -p 1 ./... -run Integration -count=1`). A fresh Postgres database must pass on its first
run and ten consecutive repetitions without two packages racing the same migration.

This is a test-runner ownership fix. Coordination of migrations between multiple deployed server
processes remains a deployment RFC concern and is not implied by this target.

### D5 — Remaining integrity boundaries are explicit

1. The save migration fixture's top-level metadata is named `corpus_version`, distinct from the
   save versions carried by its cases. A checked-in corpus baseline records minimum case count and
   required case names. `make verify-server` rejects a missing required case or a count decrease;
   changing the baseline requires a separate reviewable commit.
2. Canonical docs state that an event retains its immutable origin revision number and hash but has
   no foreign-key dependency on a retained save snapshot. Snapshot pruning cannot delete history;
   it also cannot promise the historical snapshot remains queryable.
3. Receipt construction is fallible. `wireChanges` returns a typed error when a supposedly
   canonical before/after value cannot parse, and the intent path propagates that failure as
   `internal_invariant`; it must never silently omit a malformed change.
4. A terminal request recorded under a valid `intent_id` remains bound to its canonical request
   hash. Correcting the payload while reusing that id returns `idempotency_conflict`; the client
   must use a new id for a different logical request. An integration regression test and canonical
   documentation make this intended sticky behavior explicit.
5. `internal_invariant` remains an internal Go rejection plus structured audit/metrics signal. Its
   wire representation is deferred to the WebSocket Transport RFC, which owns the rejection
   mapping. This RFC adds no placeholder transport schema.

## Deviations from design

None. The C7 clarification changes an implemented RFC sentence, not game design: it reconciles
numeric observability with the already-binding rule that persisted events belong to real save
revisions.

## Acceptance criteria

1. Go and TypeScript catalog suites assert the shared D1 declaration validation. Go multiplier and
   production suites assert runtime validation, raw-byte ordering, and rate permutation invariance,
   including an ASCII punctuation case that differs under locale ordering.
2. Unit and Postgres tests assert every D2 outcome path, transaction boundary, and replay
   non-duplication rule through the exported `InvariantSink` contract.
3. A production-rate or ordering-authority AST edit changes the formula artifact and fails
   `make formulas-check`; comment-only and formatting-only edits do not. The schema-v2 artifact
   includes its validated fingerprint.
4. A new Postgres database passes `make test-save-integration` on the first run and across ten
   consecutive runs under the serialized target.
5. Verification fails when a required save-migration case is removed, when the corpus shrinks,
   or when receipt change values cannot parse. Idempotency tests cover corrected-payload reuse.
6. Canonical production/save documentation describes the shipped invariant, formula, event
   retention, receipt-error, and retry contracts without claiming deferred transport behavior.
7. `make verify`, `make test-save-integration`, and the formula and corpus drift gates are green
   before the Balance Harness's first baseline is generated.

## Open questions

None. Owner acceptance is required before implementation and before creating the planning
directory.

## Changelog

- 2026-07-29: drafted from the accepted R4–R8 remediation sequence after R1–R3 completed and
  received per-change approval.
- 2026-07-29: accepted by owner direction and implementation started.
- 2026-07-29: D1–D5 implemented; canonical docs updated; full Go, Postgres, TypeScript, schema,
  formula, and three-browser verification green; RFC archived.
