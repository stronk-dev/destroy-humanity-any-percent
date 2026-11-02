# Headcount allocation

Only the source-independent arithmetic library is implemented. It is **not wired
into gameplay**, and no headcount catalog, save, intent, resource rate or UI is
activated. Marco selected paid hiring with producing/redirecting roles on
2026-11-02; the two specifications' model translation is independently approved.
Remaining technical identity/acceptance and content/policy contracts still
require reconciliation before integration.

## Arithmetic API

Go `headcount.Project(hardcap, seats, roles, assigned)` and TypeScript
`headcount/allocation.project(...)` implement Headcount Allocation S2. Inputs are
the explicit safe-integer budget and complete assignment, plus the resolved
arithmetic subset of each role: ID, kind, source ID and heads per converter.
These role inputs are **not** a catalog or public-wire codec. Catalog resource,
rate, price, gate, multiplier and synergy bindings are not implemented here.

Both functions reject an incomplete/unknown/negative/over-budget assignment,
invalid count or ratio, invalid/duplicate role, unknown kind, missing source,
conversion chain, or two converters for one source. Inputs are never modified.
Go returns `ErrInvalidAllocation`; TypeScript throws `RangeError`.
Role IDs must be primitive strings before lexical ordering or assignment lookup. TypeScript
does not coerce booleans, null, arrays or boxed strings into mechanical names, and ordering
does not execute their coercion hooks. Valid string names such as `false`, `null` and
`constructor` are not reserved or rewritten. This is an arithmetic-input check, not a new
catalog/public-wire codec or a change to allocation formulas.

Outputs are fresh maps `producing_heads` (produce and pool-feed roles),
`diverted` and `idle_converters` (converter roles), plus `unassigned` seats.
Converters redirect source heads; they do not duplicate them. A partly occupied
converter is not idle. Feed roles retain their own assigned heads independently.

The converter capacity is evaluated without multiplying when that product could
exceed the source count. Go ceiling division also avoids an overflowing addition.
Everything is constant for these inputs; no clock, Decimal, RNG or mutation is read.
This alone does not prove production integration, offline accrual or Law 2 for
the eventual complete system.

## Verification

`testdata/headcount-allocation.json` contains twelve literal effective-heads
vectors, not the full future catalog/rate/transition/replay acceptance corpus.
Go and TypeScript consume the same file. Independent test-only arbitrary-precision
models cover the small domain and maximum-safe-integer boundaries. Rejections,
input immutability, role ordering and fresh result ownership are also checked.

`make verify-client-boundary` also executes a Node-only intermediate-product
witness against the actual TypeScript source. A TypeScript AST transform observes
each native `*` expression while running the twelve shared vectors; an independent
BigInt check refuses products outside the safe-integer domain. This checks an
algorithmic restriction, not merely the clamped output. Its source-derived naive
multiply/clamp control still returns the correct admitted-edge output but is
rejected by the observer. Parse errors, imports, unobserved multiplication,
compound multiplication and exponentiation refuse rather than silently bypass
this narrow instrument. No runtime hooks, new dependencies or browser exclusions
are added. This bounded execution witness is not a proof over all inputs or
arbitrary future arithmetic/helper implementations; existing exact-model tests
remain required. It does not reconcile or satisfy the RFC's impossible TS
overflow-output criterion on its own.

Run from the repository root:

```sh
make test-go GO_PACKAGES='./headcount' GO_TEST_FLAGS='-count=1'
make vet GO_PACKAGES='./headcount'
client/node_modules/.bin/vitest run --root client test/headcount-allocation.test.ts
make test-browser-focused BROWSER_TEST_FLAGS='test/headcount-allocation.test.ts --project=chromium --project=webkit'
make test-go-ci CI_TEST_PACKAGES='./headcount' CI_TEST_FLAGS='-v'
make verify-client-boundary
```

Actual results, unavailable engines and pending designated review live in the
[owning log](../planning/headcount-allocation/log.md). RP-413 records the RFC's
inadmissible overflow example and unproved mandated mutant; no full AC3/AC5 or
player-workflow acceptance is claimed. Register both module prefixes in kernel
identity and bump the kernel version when authoritative integration lands.
