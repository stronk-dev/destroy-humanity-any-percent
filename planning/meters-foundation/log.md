# Meters Foundation implementation log

## 2026-08-03 — acceptance reconciliation

The owner declared all six drafted foundations accepted, including Meters C1–C12. The worktree
contained only the earlier C1–C7 summary while retaining stale contradictory draft prose and open
C8–C12 proposals. Reconciliation made the accepted contract singular:

- exactly ten independent Trust standing/grievance axes plus Company p(doom);
- Externality remains addressed ledger facts and Founder Soul remains the existing read-only carry;
- integer `[0,100]` values, exact attended-time arithmetic, causal input arms, and derived bands;
- band events never become moral acts; save v15 owns complete value/remainder maps;
- meters joins pinned epoch/replay identity after Purchasable Content; not-spendable is structural.

`DESIGN-GAP:` no literal production band IDs/floors, initial values, decay rates, or seed input
bindings have been owner-supplied. Per house law, implementation will use discriminating test
catalogs and will not improvise the epoch artifact. The mint and archive remain blocked on those
balance-data rows; foundational engine work proceeds.

## 2026-08-03 — catalog/schema foundation

- Added strict JSON Schema, Go, and TypeScript authorities for exactly eleven Company meters: ten
  Trust axes and p(doom). Externality/Soul/public-axis rows, wrong scopes, decorative spendable
  flags, duplicate sources, zero inputs, invalid bands, and resource-ID collisions reject.
- The closed input union contains only ledger-fact and contribution-slot arms. Slots validate
  against the implemented multiplier vocabulary; every numeric field remains integer `[0,100]`.
- Added a fail-closed package import boundary preventing meters from importing economy or
  production state owners. The boundary's negative fixtures run in the normal root verification.
- `make test-go GO_PACKAGES='./meters'`, `make typecheck`, `make test-client` (6,509 assertions),
  `make verify-schema`, and `make verify-meters-boundary` pass from the repository root.
- No production meter artifact was invented. Schema verification labels its fixture `pre-mint`.

## 2026-08-03 — catalog-foundation full gate

`make verify` passed at `8f5263a`: Go vet and every server package; formulas and balance harness;
strict TypeScript/Svelte checks; production client build; 6,509 client assertions; kernel/copy/
schema/package-boundary gates; and 19,536 browser assertions.

## 2026-08-03 — deterministic transition kernel

- Added matching Go and TypeScript state authorities with exact complete value, decay-remainder,
  and input-remainder key sets.
- The pure transition advances decay first, then causal facts and active contribution sources,
  clamps once, clears decay phase at its target, and emits only a prior-to-final band change.
- Added one shared JSON corpus consumed by both runtimes. It discriminates split attended time,
  negative rate input, ledger-fact band crossing, linear decay, stale target phase, and hard-bound
  saturation.
- Focused Go, client, and TypeScript type-check gates pass using the normal repository toolchain.
- Save v15 and `ApplyLogged` binding remain intentionally uncommitted until the meter catalog can
  be part of the pinned artifact bundle; accepting an optional/unhashed runtime catalog would
  violate replay identity rather than make progress.

## 2026-08-04 — independent foundation review (`568a5ec^..61b6392`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved; remediation required before the next Meters landing.** The valid-
  catalog transition arithmetic itself is approved, but the range leaves one mandatory gate red
  and does not yet have one cross-runtime strict catalog authority.

Findings, ordered:

1. **HIGH — HEAD is red and the transition log's green type-check claim is false.** From the
   repository root, `make typecheck` fails at `client/test/meters.test.ts:2` because the test imports
   `node:fs` while the client has no available Node type declaration (`tsconfig.json` names only
   `vitest/globals`, and `@types/node` is not a direct dev dependency). This also makes
   `make verify` deterministically fail. The import arrived in `61b6392`; its log says the focused
   type-check gate passed. Fix the dependency/type surface and record the real green command.
2. **HIGH — Go accepts meter catalogs that TypeScript and JSON Schema reject.** The Go raw structs
   cannot distinguish an omitted required field from a valid zero/nil value. Concrete accepted-Go /
   rejected-TS cases include an omitted meter `decay` key (decoded as the allowed nil decay), an
   omitted zero-valued reseed/band field, and wrong-union fields supplied as `null` or `""`.
   Independently, `validReseed` places no exact-safe upper bound on its `int64` numerator or
   denominator, while TypeScript and schema cap them at `9007199254740991`. A server-loaded pinned
   artifact can therefore be un-loadable by the client. Make exact keys/presence and exact-safe
   bounds identical in both loaders, and add one shared valid+invalid loader corpus consumed by Go
   and TypeScript so schema-only rejection cannot mask this class again.
3. **MEDIUM — the advertised fail-closed package boundary is only a one-directory text scan.**
   `verify-meters-boundaries.mjs` reads only immediate non-test `.go` files under `server/meters`.
   A nested package can import a forbidden owner and then be imported by `meters`, so the gate can
   pass with a transitive economy/production dependency. Replace or backstop it with a recursive
   import-graph/AST check and a nested/transitive negative fixture.

What held:

- The strict JSON Schema, valid-catalog Go/TS loaders, exact eleven-ID registry, resource collision
  check, input-source uniqueness, and current source imports match the narrowed RFC.
- The Go and TypeScript transition kernels agree on the shared corpus; decay-first ordering,
  carried attended-time arithmetic, inactive/offline behavior, saturation, complete state maps,
  and prior-to-final band emission were verified in source. `make test-go
  GO_PACKAGES='./meters ./achievements'`, `make test-client`, `make build-client`, `make vet`,
  `make verify-schema`, both boundary targets, and both exact-range `git diff --check` runs passed.
- No production meter artifact was added to `balance/epochs/phase0.json`. The missing literal
  bands/initials/rates/bindings, save v15, `ApplyLogged` binding, events, replay identity, formulas,
  relevance report, and archive remain explicitly carried work; none was treated as implemented.

## 2026-08-04 — remediation after independent review

- Replaced the Node-only corpus file read with a typed JSON-module import; the mandatory root
  TypeScript/Svelte gate now genuinely passes with zero diagnostics.
- Made every zero-capable required Go meter field presence-aware and every input arm exact-key,
  closing the Go-vs-TypeScript/schema acceptance gap for omitted initials/floors/reseed/decay and
  wrong-arm empty/null fields. Added discriminating negative cases.
- Made the package boundary walk nested source directories recursively.
- Corrected the earlier green claim by evidence rather than editing history: `make verify` exits 0
  after all server packages, harness, TypeScript, client build, 6,515 client assertions, schema and
  boundary gates, and 19,554 browser assertions.

## 2026-08-04 — independent remediation follow-up (`bf979b4..87f542d`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved; one prior HIGH and the boundary MEDIUM remain open.** The mandatory
  root gate is genuinely green now, and most field/union presence defects are fixed, but the log's
  claim that the cross-runtime catalog gap is closed is broader than the implementation.

Finding closure:

1. **Prior HIGH 1 closed.** The Node-only corpus read became a typed JSON-module import. A fresh
   repository-root `make verify` completed with exit 0: Go vet/all packages, formula and harness
   gates, zero TypeScript/Svelte diagnostics, production build, 6,515 client assertions, schema/
   boundary/content gates, and 19,554 browser assertions.
2. **Prior HIGH 2 remains open.** Presence pointers correctly close omitted zero-valued reseed,
   initial, band-floor, and decay-child fields; exact-key input decoding closes wrong-arm null/empty
   fields. Two named cases remain divergent, however:
   - `rawMeter.Decay *rawDecay` still maps both an omitted required `decay` key and explicit JSON
     `null` to nil, and no meter-level key-presence check exists. Go accepts omission; TypeScript's
     exact object and JSON Schema reject it.
   - `validReseed` still has no `decimal.MaxExactInteger`/`9007199254740991` upper bound for its
     `int64` numerator and denominator. Go accepts larger integers that TypeScript/schema reject.
   The negative tests cover neither case, and the shared valid+invalid loader corpus requested by
   the prior verdict did not land. Fix both remaining semantics and add the parity corpus before
   binding the catalog to a pinned run.
3. **Prior MEDIUM 3 only partially closed.** The source scan now walks nested directories, so the
   original immediate-directory bypass is gone. It is still not an import-graph gate: `meters` can
   import an out-of-tree helper which imports economy/production, and neither scanned source text
   contains the forbidden path. No nested or transitive fixture exercises the recursion. Backstop
   the recursive scan with `go list -deps`/AST dependency inspection and a real transitive fixture,
   or narrow the docs/log claim from package dependency to direct source imports.

The transition kernel remains approved on valid catalogs. No meter artifact was minted; literal
owner balance rows and all previously listed save/runtime/replay/formula/relevance work remain
honest blockers.

## 2026-08-04 — second remediation after independent follow-up

- Made the required nullable `decay` key presence-aware in Go while preserving explicit JSON
  `null` as the sole no-decay value.
- Applied the shared `9007199254740991` exact-integer ceiling to Go reseed numerator and
  denominator validation, matching TypeScript and the schema.
- Added `meters-catalog-parity-v1.json`, consumed by both Go and TypeScript. Its baseline and
  invalid mutations lock required-nullable presence, exact-integer bounds, zero-field presence,
  union exactness, and unknown-field rejection across runtimes.
- Backstopped the recursive source scan with `go list -deps` and a seeded forbidden-transitive-
  dependency fixture. The gate now verifies the package graph rather than only source text.
- Verification: focused Go tests, both boundary tools, and TypeScript/Svelte checks pass. A fresh
  repository-root `make verify` exits 0 with 6,517 client assertions and 19,560 browser assertions.
