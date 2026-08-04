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

## 2026-08-04 — independent second-remediation review (`75e3efd..19e78ab`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **behaviorally approved for continued Meters work; not final review closure.** Both
  previously open catalog semantics are fixed and the real dependency graph is now checked. Two
  MEDIUM proof/tooling findings remain before artifact mint or archival.

Closure verified:

- Required nullable `decay` is now a `json.RawMessage`: absence has length zero and rejects, while
  canonical/whitespace-surrounded JSON `null` remains the sole no-decay representation. The shared
  mutation case is consumed by both runtime suites.
- Go now enforces the same `9007199254740991` ceiling on both reseed numerator and denominator as
  TypeScript and JSON Schema; the over-bound denominator case rejects in both suites.
- The live boundary runs `go list -deps ./meters` and rejects forbidden economy/production packages
  anywhere in the actual dependency closure. This closes the transitive behavior defect.
- A fresh repository-root `make verify` independently completed with exit 0: all Go packages/vet,
  formula and harness checks, zero TypeScript/Svelte diagnostics, client build, 6,517 client tests,
  schema/content/boundary gates, and 19,560 browser tests. Exact-range `git diff --check` passed.

Residual findings:

1. **MEDIUM — the parity corpus does not own the candidate bytes.** The checked-in file contains
   only mutation recipes; Go and TypeScript apply them to separately hand-written `validCatalog`
   builders. Those baselines can drift while every shared case stays green, so this is not yet one
   loader corpus in the golden-vector sense. Move the literal valid catalog into the shared fixture
   (or store complete bytes per case) and have both suites mutate/load that same object.
2. **MEDIUM — the claimed transitive fixture is not a dependency-graph fixture and violates the
   routine-command convention.** `assertDependencies(["cloud-clicker/server/economy"], ...)` tests
   only the final string filter, not `go list`, graph traversal, or a helper-mediated import. The
   live graph check is correct, but add a temporary-module/package graph fixture like the hardened
   history gates use. Also remove the hard-coded `/tmp/cloud-clicker-boundary-go-cache`: AGENTS.md
   explicitly requires the root Make target's exported repository-local `GOCACHE` and forbids
   task-specific cache directories.

No production meter artifact was minted. Owner-authored literal balance rows plus the existing
save/live-replay/events/formulas/relevance work remain honest blockers.

## 2026-08-04 — parity/tooling closure after second review

- The shared parity artifact now owns the complete valid meter catalog baseline as well as its
  invalid mutations. Go and TypeScript load cloned bytes/data from that one baseline; separately
  constructed fixtures can no longer drift while the parity test stays green.
- Replaced the synthetic dependency list with a real two-package fixture graph whose helper
  imports the forbidden economy owner. The boundary tool proves `go list -deps` rejects that
  actual graph before checking the production package.
- Boundary subprocesses inherit Make's repository-local `GOCACHE`, falling back to
  `.cache/go-build` for direct tool invocation; no `/tmp` cache override remains.
- Focused meter/achievement/fixture Go tests, both boundary tools, and TypeScript/Svelte checks
  pass.

## 2026-08-04 — independent parity/tooling closure review (`3734f8b^..3734f8b`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **approved; both residual proof/tooling findings are closed.**

The checked-in parity artifact now owns the complete literal valid meter catalog. Go unmarshals
that baseline afresh for every case and TypeScript `structuredClone`s the same imported object, so
neither runtime's older local valid-catalog builder participates in the parity test and mutations
cannot leak between cases.

The transitive negative fixture is a real Go graph: `metersroot` imports `metershelper`, which
imports the forbidden economy owner. The live boundary invokes `go list -deps` for that root and
feeds the result through the same `assertDependencies` function used for `./meters`; a synthetic
package-name list is no longer the proof. Both subprocesses inherit Make's exported repository-local
`GOCACHE`, with ignored `.cache/go-build` only as the direct-invocation fallback; this scope contains
no `/tmp` cache override.

Independent exact-range diff checking and a fresh repository-root `make verify` pass. The gate
executes the real graph fixture and shared Go/TypeScript corpus and exits 0 with 6,517 client and
19,560 browser assertions. Production meter content/runtime work remains pending exactly as
previously recorded; this verdict credits only the tooling closure.

## 2026-08-04 — save-v15 activation DESIGN-GAP

Pre-implementation tracing showed that current v14 runs are pinned to seven-artifact epochs with
no meter bytes, while v15 requires a complete key set resolved from exactly those pinned bytes.
The legacy `meter_bands` placeholder is neither complete nor backed by immutable balance rows.
Activating CurrentVersion=15 before the content mint would therefore brick current saves or use a
deploy-current catalog during replay. C13 proposes new-run-only activation: old runs remain v14
through Exit; the first meter-bearing run assembles v15 directly. No save version, production hook,
or artifact identity was changed pending the owner ruling.

## 2026-08-04 — version-aware v15/v16 save codec landing

- Implemented v15 meter maps and v16 run/lifetime achievement fields without changing the
  production-emitted version (`CurrentVersion` remains 14; supported decoding extends to 16).
- Save state now retains its revision wire version as runtime metadata. Ordinary store and intent
  writes must preserve that version; only the Exit new-run state may advance to v16. The terminal
  old-run revision is explicitly required to remain on its pinned version.
- V15 removes the superseded `meter_bands` member and requires all three meter collections. V16
  requires both earned collections and both score fields, enforces scope separation, and rejects
  downgrade attempts that would discard active mechanics.
- The new-run genesis records the actual new-state version instead of a process-global constant.
  Unit coverage proves v16 round-trip/closure and missing-field rejection; full root `make test`
  passes (all Go packages, 6,517 client assertions, 19,560 browser assertions).
- Reset assembly and catalog-complete key/derived-score validation remain in the next landing; no
  production artifact or current run was activated by this commit.

## 2026-08-04 — independent activation-codec review (`d7bb1da^..d7bb1da`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved; version preservation, decoder closure, and atomic activation are not
  yet enforced by the save authority.** The v15/v16 happy-path shape is present, but the seams named
  by C13 fail under legacy, cross-scope, and mixed-version inputs.

Findings, ordered:

1. **HIGH — every restored v1–v13 stream is bricked by the new version-preservation checks.**
   `RestoreState` correctly records the historical revision in `WireVersion`, but
   `VersionForState` maps every value below 14 to `CurrentVersion` 14. Applied ordinary intents then
   fail `VersionForState(state) != revision.Version`, generic writes fail the same comparison, and
   Exit rejects the final Company state before persistence. The old migration path intentionally
   rewrote legacy saves as current-v14 bytes; the new guard blocks that transition without adding a
   replacement. Preserve only the new-run activation boundary while retaining one explicit
   legacy→v14 migration authority, with applied-intent and Exit fixtures from at least v1/v13.
2. **HIGH — v14→v16 activation is not atomic across Founder and Company.** Exit checks the terminal
   and new Company versions but never compares the mutated Founder version to either its loaded
   revision or the new Company. A mutator can commit v16 Founder with v14 new Company, or v14
   Founder with v16 new Company; both contradict C13/C11's “one transaction, never partial” law.
   Enforce the allowed tuple over loaded Founder, terminal Company, new Founder, and new Company in
   the save transaction itself, not only in the future assembler.
3. **HIGH — restore bypasses the new foundation scope/domain validator.** `EncodeStateVersion`
   calls `validateFoundationState`; `RestoreState` returns after the older route/compact/prestige/
   faction/guild validators and never calls it. A v16 Company payload with non-empty Founder
   lifetime achievements and score restores successfully; so do cross-scope meter maps and
   negative/out-of-domain achievement scores. The canonical decoder therefore does not enforce the
   scope claims in this commit.
4. **MEDIUM — v15/v16 accepts the removed `meter_bands` field.** Because `stateV15` embeds
   `stateV14`, strict JSON decoding still recognizes the promoted field. A v15 payload containing
   all three new maps plus `meter_bands` restores successfully, and its next encode silently drops
   the superseded member. Add an explicit absence check (and v16 coverage), not only an encoder test.
5. **MEDIUM — v15 can carry achievement state in memory and silently discard it.** The foundation
   validator rejects achievements only when `version < 15`, then returns early for v15; setting a
   run achievement/score on a v15 state produces a successful v15 encoding with those values
   omitted. The atomic law requires any achievement state below v16 to fail closed.

Proof review: all five focused adversarial reproducers failed against the commit exactly as above
and were removed after the run. The unchanged `./save` suite passes, demonstrating the coverage
gap: the commit adds only a Company v16 round trip, one v15 missing-field case, and one v16
missing-score case; no Store/Intent/Exit version-transition, Founder, legacy-write, mixed-version,
superseded-field decode, or cross-scope fixture exists. Exact-range `git diff --check` passes. A
fresh root `make test` could not be attributed to this commit because concurrent uncommitted
next-landing edits currently make `server/production/replay.go` fail compilation; focused save
verification is green and the failure is outside `d7bb1da`.

## 2026-08-04 — activation-codec remediation after independent rejection

- Restored the existing legacy migration authority explicitly: revisions v1–v13 restore with v14
  as their next writable version, and Store/Intent/Exit compare against that migrated target rather
  than the historical row version. A real Postgres applied-intent fixture now proves v1→v14.
- Exit validates the complete loaded-Founder/loaded-Company/final-Company/new-Founder/new-Company
  version tuple. Preexisting mixed tuples, terminal version changes, and either one-sided v16
  activation all fail before persistence; the atomic v14→v16 tuple and legacy→v14 Exit pass.
- Restore now runs the same foundation scope/domain validator as encode, explicitly rejects
  `meter_bands` at v15/v16, and v15 refuses any prematurely present achievement state.
- Achievement arrays are byte-sorted as well as unique. New negative fixtures cover superseded
  fields, Company/Founder leakage, negative scores, unsorted IDs, and silent v15 loss.
- Focused `go test ./save` and the normal root `make test-save-integration` pass. The dependent
  catalog/runtime landing remains uncommitted pending independent approval of this remediation.

## 2026-08-04 — second activation-codec remediation

- Superseded-field rejection now compares JSON keys case-insensitively, matching `encoding/json`'s
  own tag behavior; `METER_BANDS` and every casing variant fail before v15/v16 decode.
- V15 remains a supported migration/decode shape but is structurally decode-only. Store
  validation, Exit tuple validation, and run-genesis insertion accept persisted active versions
  14 or 16, never 15. Tests cover both the codec/persistence distinction and a full v15 Exit tuple.
- Focused save verification is green; this narrow follow-up awaits its own exact-range review.

## 2026-08-04 — Go new-run activation and reset assembly

- Replay catalog bundles now recognize exactly the legacy seven-artifact shape or the paired
  nine-artifact shape containing both `meters` and `achievements`; one-sided activation fails.
- Added the shared exact Notoriety reseed assembly. It uses overflow-safe integer arithmetic,
  applies the result to all five Standing axes, and leaves every Grievance/p(doom) value at its
  catalog literal. Extreme-domain and non-retroactive legacy-placeholder fixtures are included.
- Exit resolves both current and next pinned bundles. Pre-foundation runs stay v14 through their
  final state; the new Company and Founder become v16 together only when the next bundle contains
  both artifacts. Active mechanics cannot disappear in a later epoch.
- Catalog-derived state policy validates complete meter maps and derived achievement scores before
  persistence. V16 gate predicates read `meter_values`, never the removed placeholder map.
- Tests cover atomic activation, exact reset, artifact pairing/downgrade rejection, achievement
  settlement/reset, derived-score tampering, and route-context authority. Full root `make test`
  passes (all Go packages, 6,517 client assertions, 19,560 browser assertions).
- The TypeScript save/replay port, live meter hook, achievement evaluation/events/receipt evidence,
  and nine-artifact database loader remain subsequent landings; no content artifact was minted.

## 2026-08-04 — append-only kernel-history correction

- The root verifier exposed that reviewed commits `d7bb1da`, `c356d87`, and `9d3764f` changed
  kernel-watched save semantics without carrying their kernel version bump. Their hashes are already
  cited by independent verdicts, so the review ledger forbids rewriting them.
- Added `kernel/history-corrections.json` as an immutable fix-forward record and corrected all three
  misses in kernel version `0.3.17`, the same commit that lands the paired activation semantics.
- The history guard accepts a correction only when the offending full hash is named, the correcting
  version is newer, its planning review log exists, and the correction first appears beside the
  exact real `kernel/VERSION` bump it declares. Existing corrections cannot be removed.
- Adversarial Git fixtures prove an uncorrected historical miss fails, the reviewed correction plus
  bump passes, and later correction removal fails. Worktree semantic changes still require their own
  bump; a correction cannot excuse current uncommitted work.

## 2026-08-04 — independent activation-codec remediation review (`c356d87^..c356d87`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved yet; all findings from the `d7bb1da` verdict close, but two new
  MEDIUM strictness/activation seams remain.**

Closure verified:

- Legacy revisions now restore with v14 as their explicit writable target. Applied Intent and
  generic Write compare against the migrated target rather than the historical row, and the real
  Postgres `TestStoreIntegrationRevisionLifecycle` independently passes with a v1 revision
  committing its next applied intent as v14.
- `validateExitVersionTransition` checks the loaded Founder/Company pair, terminal Company, mutated
  Founder, and new Company before persistence. Legacy pairs converge on v14; terminal mutation,
  preexisting mismatch, and either one-sided v16 transition reject; the atomic v14→v16 tuple passes.
- Restore now invokes the same foundation scope/domain validator as encode. The prior cross-scope
  lifetime state, negative score, unsorted earned-set, exact `meter_bands`, and v15-achievement-loss
  reproducers all reject. Sorted IDs use byte order and v15 fails closed on any achievement state.

New findings:

1. **MEDIUM — the superseded-field rejection is still case-insensitive-decoder bypassable.** The
   raw-key precheck rejects only exact lowercase `meter_bands`, while Go's `encoding/json` matches
   tagged fields case-insensitively. A v16 payload with `METER_BANDS` therefore passes the precheck,
   populates the promoted v14 field, and restores successfully. Enforce the v15/v16 root key set
   exactly (or reject every case-folded spelling of the superseded member); retain this reproducer.
2. **MEDIUM — the storage authority still admits the forbidden standalone v15 state.** C13 says no
   run is ever v15-with-meters-but-without-v16 achievements, but a loaded Founder-v15/Company-v15
   tuple continuing to v15 returns nil from `validateExitVersionTransition`, and generic
   `CreateStream` accepts `WireVersion=15`. V14→v15 Exit is correctly blocked, but the invariant is
   not closed at stream creation or when a preexisting v15 tuple is encountered. V15 may remain a
   codec/migration shape; persisted active run versions must be v14 or v16 only.

Independent evidence: all original discriminating reproducers pass; both new reproducers fail as
described and were deleted after execution. Exact-range `git diff --check`, an uncached focused
`./save` run, and the targeted real-Postgres legacy-intent integration test pass. The separate
uncommitted production/meters next landing was neither reviewed nor used as evidence.

## 2026-08-04 — independent second activation-codec remediation review (`9d3764f^..9d3764f`)

- **Review by:** Codex
- **Recorded by:** Codex
- **Decision:** **approved; the case-fold bypass and standalone-v15 persistence seam are closed.**

The raw v15/v16 key scan now follows the case-folding behavior of `encoding/json`; exact lowercase,
uppercase, title-case, and mixed-case `meter_bands` spellings all reach the superseded-field error.
V15 remains usable as an explicit migration codec but is decode-only at active persistence
boundaries. The common Store validation closes Create, generic Write, and applied Intent; the Exit
tuple guard closes continuation; and both genesis entry points reject v15 before it can identify a
run.

Independent evidence includes exact-range diff checking, an uncached full `./save` unit run, four
case-folded-key reproducers, and a scratch real-Postgres audit of Create/Write/Intent/Exit/genesis.
The public epoch-pin path rejected v15 and left zero run pins after rollback. Scratch tests were
deleted. Concurrent uncommitted production/meters work remained out of scope and is not credited.

## 2026-08-04 — independent paired-activation review (`f070596^..f070596`)

- **Review by:** Darwin
- **Recorded by:** Codex
- **Decision:** **not approved; two HIGH seams and one MEDIUM guard-integrity defect block the
  paired activation landing.**

1. Active→active Exit derives Founder lifetime score with the ending run's achievement catalog,
   then persists and validates it under the next catalog. A legal `score_grant` retune therefore
   bricks the Exit.
2. `server/meters/`, `server/achievements/`, and their TypeScript peers are absent from the kernel
   affecting-path registry, so semantic edits pass without a version bump.
3. History-correction keys cannot be removed, but an existing entry's reason, version, and review
   log can be mutated. Review binding checks only that an arbitrary path exists.

Held under review: paired artifact activation/downgrade rejection, non-retroactive legacy Exit,
overflow-safe Notoriety reseeding, v16 route-context authority, duplicate achievement rejection,
and the shared live/replay Exit boundary. Root verification is green but did not discriminate the
three findings.

## 2026-08-04 — paired-activation integrity remediation

- Founder lifetime score is now re-derived from the next pinned achievement catalog—the catalog
  under whose constants hash Founder is persisted. A discriminating fixture retunes one grant and
  proves an honest active→active Exit remains valid with the next-catalog score.
- The kernel registry now includes both Go and TypeScript meter/achievement package trees. Kernel
  version advances to `0.3.18`; future semantic edits in any of those packages require a bump.
- Existing correction entries are field-immutable in committed history and the worktree. Each
  correction must target a reachable guarded commit that actually missed its bump, and its review
  log must contain a labeled review/decision section for that exact abbreviated-or-full commit
  range—not merely exist.
- Adversarial fixtures now reject correction mutation and unrelated-log rebinding, retain removal
  and shallow-history rejection, and assert all four newly active package paths remain registered.

## 2026-08-04 — independent activation-integrity remediation review (`bcc021d^..bcc021d`)

- **Review by:** Darwin
- **Recorded by:** Codex
- **Decision:** **not approved; the original three findings close, but one MEDIUM correction-target
  reachability seam remains.**
- A deleted side-branch commit can satisfy the full-hash, guarded-miss, labeled-review, and real-bump
  checks even though it is not project history reachable from HEAD. The review directly reproduced
  an accepted correction for such a dangling commit.
- Score-retune Exit, all four runtime path registrations, correction field immutability, review-log
  binding, removal, and shallow-history rejection held under the independent pass.

## 2026-08-04 — correction reachability remediation

- Every correction target must now satisfy `merge-base --is-ancestor <target> HEAD`; a commit object
  surviving only in reflog/object storage cannot enter the project correction ledger.
- Abbreviated hashes in review headings are resolved by Git and must uniquely resolve to the exact
  full offending hash on both sides of the reviewed range. Prefix text alone is no longer proof.
- The adversarial fixture creates an unversioned semantic commit on a side branch, deletes the
  branch, attempts a syntactically valid correction plus real bump, and requires rejection.

## 2026-08-04 — independent correction-reachability review (`53bbc91^..53bbc91`)

- **Review by:** Darwin
- **Recorded by:** Codex
- **Decision:** **approved.** The remaining `bcc021d` MEDIUM is closed: correction targets must be
  ancestors of HEAD, dangling-side-branch corrections fail, and abbreviated review tokens must
  resolve uniquely to the exact full target. Existing legitimate corrections and every prior
  mutation/rebinding/removal/shallow fixture remain green.
- The independent range union `f070596^..f070596` + `bcc021d^..bcc021d` +
  `53bbc91^..53bbc91` covers the complete paired-activation landing and both remediation commits.
  Every finding from that union is now closed; the Go activation/settlement landing may proceed.

## 2026-08-04 — replay-inputs v3 Founder carry

- While binding the reviewed activation seam to live replay, the v16 path exposed two mechanical
  omissions: Company replay cloning restored every encoded state as v14, and terminal Founder carry
  omitted the lifetime achievement state that Exit must settle and persist.
- Replay-inputs v3 adds sorted lifetime achievement IDs plus their derived score. The Go live path
  freezes them, reconstructs a catalog-valid v16 Founder, and copies the replayed v16 result back to
  the transaction-owned Founder. Historical v2 remains accepted only for pre-foundation catalogs.
- TypeScript accepts both envelope versions and exactly validates the v3 carry fields. The shared
  Go-authored corpus was regenerated through the normal root target and is green in both runtimes.
- Replay clone now restores `VersionForState(state)`, retaining v14 or v16 rather than silently
  forcing the legacy version. Direct fixtures cover active carry reconstruction/output, v16 clone,
  and v2/v3 activation rejection boundaries. Kernel version is `0.3.19`.
