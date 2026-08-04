# Achievements Foundation implementation log

## 2026-08-03 — accepted-contract reconciliation

The owner accepted C1–C10, narrowing the RFC from Clout + achievements to Achievements only. The
file still contained the rejected Clout mint/stack in its summary, normative sections, acceptance
criteria, open questions, README status, and last changelog entry. Reconciliation makes one rule:

- achievement IDs and exact achievement score are permanent and non-spendable;
- ordinary Company transitions latch run IDs; the existing multi-stream Exit settles Founder
  lifetime state idempotently and resets the next run;
- Clout has no writer or production factor in this foundation; future social activity owns its
  sole mint and PR-Intern content owns any run-local multiplier;
- Phase-A predicates use only implemented state/event boundaries; Meters predicates append later;
- production definitions are content and copy-key data, not invented foundation mechanics.

`make verify` passed at `61b6392` immediately before this implementation round.

## 2026-08-03 — catalog, proof, and evaluation foundation

- Added strict schema-v1 Go/TypeScript loaders. Achievement IDs are byte-sorted; score grants are
  positive exact integers; every copy, generator, event, resource, and counter binding resolves
  through an explicit composition registry.
- Added the bounded condition union and the three proof arms. Possession requires an ownership
  predicate plus justification copy, provenance refuses current-possession predicates, and burn
  requires a registered event/resource plus positive canonical Decimal minimum.
- Added pure evaluation against one observation snapshot, lifetime/run latch exclusion, and score
  derivation from the earned-ID set as the single authority.
- Added JSON Schema fixtures and a fail-closed package/source gate proving this foundation imports
  no economy spending, production/save owner, or lifetime-Clout symbol.
- Focused Go, client, TypeScript type-check, schema, and boundary gates pass. No production
  achievement or copy was invented; fixtures are explicitly pre-mint.

## 2026-08-04 — independent foundation review (`61b6392..d5d7685`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved; remediation required before the next Achievements landing.** The
  Clout-free scope and pure evaluation behavior hold, but the range starts and ends with a red
  mandatory gate and the proof/catalog boundary is not yet cross-runtime or structurally sound.

Findings, ordered:

1. **HIGH — the range re-claims a mandatory gate that is red.** `make typecheck` fails at the
   Meters corpus import of `node:fs`; that defect entered in the range's parent `61b6392`, but
   `d5d7685` nevertheless records that the Achievements TypeScript type-check passed. A batch may
   not inherit a red required gate and report it green. Close the shared Node-type defect and
   re-run the real root gate before this range is approved.
2. **HIGH — Go and TypeScript disagree on accepted achievement catalog bytes.** Go decodes every
   condition/proof arm into one struct and tests zero values, while TypeScript enforces exact keys.
   Thus Go accepts, for example, `fact_present` with an extra `counter:""`, `provenance` with an
   extra `event_kind:""`, or `burn` with `event_kinds:null`; TypeScript and schema reject all three.
   Go also trusts registry membership without mechanically validating registry-backed copy,
   generator, counter, event, resource, and possession-justification IDs, while TypeScript validates
   those identifiers. Add exact arm presence checks and a shared cross-runtime valid+invalid catalog
   corpus; schema validation outside the runtime loader is not a substitute for parity.
3. **HIGH — provenance is registered but not proven.** C6/CA1 require a provenance condition to be
   derived solely from its declared immutable event kinds. The registry exposes independent sets
   of counters/facts and event kinds, with no source mapping, and `parseProof` merely rejects an
   ownership predicate. A `fact_present`/counter condition can name one source while its proof lists
   any unrelated registered event kind and both loaders accept it. Provide typed condition-source
   metadata (or another executable derivation proof) before production definitions or runtime
   earning can land; otherwise AC2 cannot be satisfied from the run log alone.
4. **MEDIUM — the source boundary is not the fail-closed build-graph gate the log/docs claim.**
   `verify-achievements-boundaries.mjs` scans only immediate non-test Go files and direct path/symbol
   text. A nested/transitive package bypasses the import rule, and the check does not implement C8's
   future sole-Clout-owner/source-registry proof. Current code is clean, but the enforcement claim
   is stronger than the gate. Use a recursive import graph/AST fixture and keep the sole-mint
   registry obligation explicitly pending until its owner exists.

What held:

- Current production source imports only the Decimal boundary and contains no Clout writer,
  lifetime-Clout factor, ledger debit, persistence owner, or production contribution. Achievement
  score remains a distinct derived exact integer and unknown/false earned entries fail closed.
- On valid fixtures, Go and TypeScript agree on byte-ordered definitions, bounded predicate-tree
  evaluation, run/lifetime latch exclusion, possession/provenance basic shape checks, canonical
  positive burn amounts, and derived score. The schema and current direct-source boundary fixtures
  pass, as do the focused Go/client/build/vet gates listed in the Meters review above.
- No achievement artifact was added to the epoch seed. Literal production achievements/copy,
  save v16, live/replay evaluation and ordered events, Exit settlement, relevance observation,
  epoch identity, and archival remain explicitly pending and were not over-credited by this review.

## 2026-08-04 — remediation after independent review

- Added exact-key checks for every Go condition/proof union arm and mechanical validation for all
  composition-registry keys, closing the Go-vs-TypeScript/schema acceptance divergence. Seeded
  null/empty wrong-arm fields now reject.
- Added typed provenance-source metadata. Every fact/counter/exit-count condition maps to sorted
  immutable event kinds; a provenance proof must include every event required to derive every
  predicate leaf. An unrelated registered event can no longer certify an achievement.
- Made the source boundary recursive. The future sole-Clout-owner registry remains honestly
  downstream rather than overclaimed by this foundation.
- Corrected the inherited false green claim: the Node-only Meters test import is gone and the
  complete root `make verify` exits 0 (6,515 client and 19,554 browser assertions).

## 2026-08-04 — independent remediation follow-up (`bf979b4..87f542d`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved yet; all prior HIGHs are closed, but the package-boundary MEDIUM and
  parity-test debt remain.** The catalog/proof implementation itself is clear to continue into its
  next state/runtime batch once the shared Meters gate is repaired.

Finding closure:

1. **Prior HIGH 1 closed.** The inherited TypeScript defect is removed, and the complete fresh
   root `make verify` evidence above was reproduced independently with exit 0.
2. **Prior HIGH 2 closed in behavior.** Go now performs exact-key checks before decoding every
   condition/proof arm, and validates every registry-backed identifier; source comparison against
   TypeScript confirms the same arm keys and mechanical identifier grammar. Wrong-arm null/empty
   and invalid-registry tests cover the Go fix. The requested shared invalid loader corpus did not
   land, and TypeScript has no matching wrong-arm/identifier cases, so retain that as **MEDIUM test
   debt** before artifact mint rather than claiming cross-runtime rejection is structurally locked.
3. **Prior HIGH 3 closed.** Both runtimes now derive typed source keys for every provenance-eligible
   fact/counter/exit-count leaf and require the proof's sorted registered event set to contain every
   event declared by that source. Unmapped and unrelated-event proofs fail; Go and TypeScript each
   carry the discriminating unrelated-event test.
4. **Prior MEDIUM 4 remains partially open.** Recursive directory walking closes a nested-file
   direct-import bypass, but the text scan still cannot see a forbidden dependency through an
   imported helper outside `server/achievements`, and no traversal fixture proves recursion. It
   also correctly does not pretend to enforce the future sole-Clout-writer registry. Add an import-
   graph/AST dependency fixture or narrow the fail-closed build-graph language before approval.

Current code remains Clout-free and production-neutral. The sole-Clout owner/source registry,
owner-authored achievement/copy content, save v16, live/replay evaluation, Exit settlement,
relevance integration, epoch mint, and archive are still explicitly pending rather than silently
credited.

## 2026-08-04 — second remediation after independent follow-up

- Added `achievements-catalog-parity-v1.json`, consumed by both Go and TypeScript. The shared
  baseline and invalid mutations lock exact object/union shapes, unrelated-provenance rejection,
  and scope restrictions in both loaders.
- Backstopped the recursive source scan with `go list -deps` and a seeded forbidden-transitive-
  dependency fixture, closing the remaining build-graph bypass without claiming the future
  sole-Clout-writer registry.
- Verification: focused Go tests, both boundary tools, and TypeScript/Svelte checks pass. A fresh
  repository-root `make verify` exits 0 with 6,517 client assertions and 19,560 browser assertions.

## 2026-08-04 — independent second-remediation review (`75e3efd..19e78ab`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **behaviorally approved for continued Achievements work; not final review closure.**
  The shared cases and real dependency graph close the runtime behavior findings, while two MEDIUM
  test/tooling gaps remain before content mint or archival.

Closure verified:

- Both suites consume the new case registry and agree on the valid baseline plus unknown-field,
  wrong-condition-arm, unrelated-provenance, and invalid-scope rejections. The source-level exact-
  arm, mechanical-identifier, and provenance-source semantics reviewed in the prior round remain
  aligned.
- The live boundary now checks `go list -deps ./achievements` and rejects economy, production, or
  save anywhere in the actual dependency closure. The future sole-Clout-writer/source registry is
  still explicitly downstream rather than falsely claimed here.
- The same independently reproduced full-root `make verify` and exact-range diff checks above pass.

Residual findings:

1. **MEDIUM — shared mutation recipes still sit on separate Go/TypeScript baseline builders, and
   identifier parity remains unasserted in TypeScript.** A malformed registry key is tested only in
   Go; the shared corpus cannot express registry mutations. Put the literal baseline and registry
   mutations/bytes under the shared fixture so a baseline or identifier-grammar drift cannot stay
   green in one runtime.
2. **MEDIUM — the “transitive fixture” is a synthetic dependency-name array, not a Go package
   graph.** It does not exercise `go list`, helper-mediated imports, or command failure. Add a real
   temporary-module/package graph fixture, and stop overriding the Makefile's approved repository-
   local cache with `/tmp/cloud-clicker-boundary-go-cache`, which conflicts with the explicit
   AGENTS.md routine-command rule.

Current production code remains Clout-free and neutral. Owner-authored achievement/copy content,
the future sole-Clout owner, save v16, runtime/replay/Exit/relevance integration, epoch mint, and
archive remain honestly pending.

## 2026-08-04 — parity/tooling closure after second review

- The shared parity artifact now owns the complete valid achievement catalog and registry
  baselines as well as catalog/registry mutations. Both runtimes consume them, including the
  previously Go-only malformed registry-identifier rejection.
- Replaced the synthetic dependency list with a real two-package fixture graph whose helper
  imports the forbidden save owner. The boundary tool proves `go list -deps` rejects that actual
  graph before checking the production package.
- Boundary subprocesses inherit Make's repository-local `GOCACHE`, falling back to
  `.cache/go-build` for direct tool invocation; no `/tmp` cache override remains.
- Focused meter/achievement/fixture Go tests, both boundary tools, and TypeScript/Svelte checks
  pass.

## 2026-08-04 — independent parity/tooling closure review (`3734f8b^..3734f8b`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **approved; both residual proof/tooling findings are closed.**

The checked-in parity artifact owns the complete literal valid achievement catalog and registry.
Go unmarshals both afresh per case and TypeScript clones both from the same imported artifact before
mutation. The registry-targeted `INVALID EVENT` case is therefore exercised by both loaders and
proves their mechanical-identifier rejection from one shared vector; the former Go-only proof and
separate hand-built baselines are no longer in the parity path.

The transitive negative fixture is a real Go graph: `achievementsroot` imports
`achievementshelper`, which imports the forbidden save owner. The live boundary invokes
`go list -deps` for that root and passes its actual dependency closure to the same filter used for
`./achievements`. Both subprocesses inherit Make's repository-local `GOCACHE`, with ignored
`.cache/go-build` only as a direct-invocation fallback and no `/tmp` override in this scope.

Independent exact-range diff checking and a fresh repository-root `make verify` pass. The gate
executes the real graph fixture, the shared catalog/registry corpus, and both runtime suites and
exits 0 with 6,517 client and 19,560 browser assertions. The future sole-Clout owner, content mint,
save/live/replay/Exit/relevance integration, and archival remain honestly pending; none is credited
by this tooling verdict.

## 2026-08-04 — save-v16 activation DESIGN-GAP

Save/runtime tracing confirmed that Achievements cannot activate independently of Meters: v16
follows meter v15, and current runs are pinned to epochs with neither artifact. Evaluating or
deriving score from a deploy-current achievement catalog would make replay timing-dependent.
C11 proposes atomic new-run activation of both artifacts/state versions, with no retroactive
earning for pre-foundation runs. No save version, hook, Exit settlement, or artifact identity was
changed pending the Meter C13/Achievements C11 owner ruling.

## 2026-08-04 — version-aware v15/v16 save codec landing

- Added v16 Company run and Founder lifetime earned sets/scores to the canonical save codec while
  leaving production emission at v14 until a new run is pinned to both artifacts.
- Persisted scope is structural: Company states cannot carry lifetime ownership; Founder states
  cannot carry run ownership; non-Founder scopes cannot carry either. Scores remain exact-safe,
  non-negative integers, and every collection/score field is presence-required at v16.
- Ordinary writes cannot change a loaded revision's wire version. Exit records the old terminal
  state at its old version and the new genesis at the version chosen by new-run assembly.
- Unit closure plus the full root `make test` are green. Catalog-derived score validation, atomic
  activation/settlement, and runtime evaluation remain explicitly pending the next landings.

## 2026-08-04 — independent activation-codec review (`d7bb1da^..d7bb1da`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved; the codec does not yet enforce C11's backward-compatible, scoped,
  atomic activation boundary.**

Findings, ordered:

1. **HIGH — restored v1–v13 players cannot make another applied intent or Exit.** Restore retains
   the historical `WireVersion`, but `VersionForState` converts every value below 14 to 14. The new
   ordinary-intent/generic-write/final-Exit comparisons then reject the state against its stored
   revision version. Historical saves previously migrated to current bytes on write; C11's
   pre-foundation rule cannot turn that supported migration chain into a permanent account brick.
2. **HIGH — the save transaction permits partial Founder/Company activation.** Only Company
   versions are checked. Founder can independently advance to v16 while the new Company remains
   v14, or remain v14 while the new Company advances to v16, and both revisions are inserted in the
   same successful Exit. C11 requires the Founder lifetime state and Company run state to activate
   as one ruled tuple; the storage authority must reject every other combination.
3. **HIGH — `RestoreState` never invokes `validateFoundationState`.** Consequently a v16 Company
   restores with non-empty Founder lifetime ownership/score, a Founder can restore Company run
   ownership or meter state, and negative/out-of-exact-domain scores load. Encode-time validation
   does not protect immutable historical/imported bytes; decoder scope and domain checks are
   mandatory.
4. **MEDIUM — the sorted-unique achievement wire grammar is only unique.**
   `uniqueMechanicalKeys` rejects malformed and duplicate IDs but never compares adjacent byte
   order, so `['achievement.z','achievement.a']` restores and is silently reordered on encode.
   CA2 requires sorted-unique sets and canonical state bytes.
5. **MEDIUM — v15 silently drops prematurely activated achievements.** The validator's inactive
   check uses `version < 15`; its v15 branch returns before checking run/lifetime achievement sets
   or scores. Encoding a v15 state containing an earned ID and score succeeds while omitting both,
   rather than exposing the illegal pre-v16 activation.
6. **MEDIUM — strict v16 decoding still accepts superseded `meter_bands`.** The promoted v14 field
   remains known through the embedded wire structs, so v16 accepts old and new meter authorities
   together and canonical re-encoding silently drops the old one.

Proof review: focused reproducers for legacy version mapping, cross-scope lifetime ownership,
unsorted IDs, v15 achievement loss, and the superseded meter field all failed against the commit
and were then deleted. The clean `./save` suite passes but has no Founder v16, scope-leak,
sorted-order, ordinary-write/Exit-transition, mixed-version, or legacy-write fixture; its new tests
cover only Company happy-path round trip and two missing required fields. Exact-range diff checking
passes. Root `make test` is presently blocked by unrelated concurrent next-landing edits in
`server/production/replay.go`; that compile failure is not attributed to `d7bb1da`, while focused
save verification is green.

## 2026-08-04 — activation-codec remediation after independent rejection

- Legacy revisions retain the supported v1–v13→v14 write migration instead of being bricked by
  the new activation guard; the behavior is proven through an applied intent against real
  Postgres.
- The save transaction now owns the atomic Founder/Company activation tuple and rejects both
  one-sided v16 combinations, mixed loaded tuples, and terminal-version mutation.
- Decode invokes foundation scope/domain validation. It rejects cross-scope earned state,
  negative/unsafe scores, unsorted earned arrays, superseded `meter_bands`, and achievements in a
  v15 state; encode and restore now fail closed in the same directions.
- Focused save tests plus `make test-save-integration` pass. No runtime/catalog activation work is
  credited until this remediation receives an independent full-range approval.

## 2026-08-04 — second activation-codec remediation

- The v15 migration codec cannot become a standalone persisted run: Store, Exit, and genesis
  authorities allow active versions 14 or 16 only. V16 remains the sole atomic activation target.
- Superseded `meter_bands` rejection follows `encoding/json`'s case-insensitive matching, closing
  the uppercase alias that the independent review demonstrated.
- Focused save verification passes; dependent runtime work remains uncredited pending review.

## 2026-08-04 — Go new-run activation and Exit settlement boundary

- The paired-artifact bundle is the sole activation authority. A pre-foundation run earns nothing
  retroactively; its Exit creates an empty v16 Company run set and empty v16 Founder lifetime set.
- On later active Exits, Company run IDs union into Founder lifetime ownership under the existing
  transaction, lifetime score is re-derived from the current run-pinned catalog, and the next run
  starts empty/zero. Artifact removal, one-sided bundles, overlap with an already-owned lifetime
  ID, and score tampering fail closed.
- Gameserver state policy now checks derived achievement score/ID closure for every v16 write.
  Full root `make test` passes. Event emission, receipt settlement evidence, live/TS evaluation,
  database artifact loading, and the production content mint remain explicitly pending.
- The landing carries kernel version `0.3.17` and the append-only correction record for the three
  already-reviewed save codec commits whose semantic version bumps were missed; the meters planning
  log owns the detailed correction protocol and adversarial fixture evidence.

## 2026-08-04 — independent activation-codec remediation review (`c356d87^..c356d87`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved yet; the six prior defects are remediated, with two new MEDIUM
  structural gaps blocking final activation-codec approval.**

Closure verified:

- The legacy write target is explicitly v14, so v1–v13 no longer collide with ordinary-write or
  Exit preservation. A targeted real-Postgres run proves a restored v1 stream commits an applied
  intent as v14.
- Exit now validates the complete version tuple. Both one-sided v16 activations, a changed terminal
  version, and an already-mixed loaded tuple reject, while legacy→v14 and atomic v14→v16 pass.
- Decoder scope/domain validation is live: Company/Founder ownership leakage, meter leakage,
  negative/unsafe scores, unsorted or duplicate IDs, exact superseded `meter_bands`, and achievement
  state below v16 fail. Encode and Restore agree on the original discriminating cases.

New findings:

1. **MEDIUM — v15/v16 decoding is not exact-key for the superseded field.** The remediation scans a
   raw map only for lowercase `meter_bands`, but `encoding/json` subsequently matches
   `METER_BANDS` to the tagged promoted member by case folding. The case-variant payload restores
   and would be silently normalized away on encode. Close the exact-key boundary rather than
   relying on `DisallowUnknownFields` for tagged-name spelling.
2. **MEDIUM — the atomic law still permits an active v15-only run.** The tuple validator accepts
   loaded Founder-v15/Company-v15 continuing as v15, and `CreateStream` can persist a v15 state.
   C11 permits v15 as the embedded codec step but states that activation jumps atomically from v14
   to v16 and no v15-only run exists. Reject v15 at active stream creation/continuation while
   preserving offline codec coverage.

All original scratch reproducers now pass; the case-variant field and standalone-v15 tuple
reproducers fail as described and were removed. Exact-range checking, an uncached focused `./save`
suite, and the isolated real-Postgres legacy-intent fixture pass. Concurrent uncommitted
production/meters work remains outside this review and is not credited.

## 2026-08-04 — independent second activation-codec remediation review (`9d3764f^..9d3764f`)

- **Review by:** Codex
- **Recorded by:** Codex
- **Decision:** **approved; both remaining MEDIUM findings are closed and the activation codec is
  clear for its dependent landing.**

Exact-range review verified that `strings.EqualFold` rejects `meter_bands`, `METER_BANDS`,
`Meter_Bands`, and mixed-case spellings before v15/v16 decoding. The v15 codec remains available
for migration coverage, while the shared Store encoder admits only active v14/v16 state. A scratch
real-Postgres reproducer exercised every persistence surface: Create, generic Write, applied Intent,
the complete Exit tuple, both genesis transaction entry points, and the public epoch-pin/genesis
path all reject v15. The public pinning rejection also rolls back the preliminary `run_epochs` row,
so the failure leaves no partial run identity.

Evidence: exact-range `git diff --check`, an uncached full `./save` unit run, four case-folded-key
reproducers, and the all-surface Postgres reproducer pass. Scratch tests were removed after the run.
The separate uncommitted production/meters work was excluded from both source review and evidence.

## 2026-08-04 — paired-activation review and remediation

- **Review by:** Darwin
- **Recorded by:** Codex
- **Decision on `f070596^..f070596`:** **not approved.** A score-grant retune proved that deriving
  Founder lifetime score with the ending run's catalog contradicts persistence under the next hash.
- The remediation derives the complete Founder lifetime score with the next pinned achievement
  catalog and adds a direct active→active retune fixture. The meter planning log owns the two kernel
  guard findings and their `0.3.18` fix-forward evidence.

## 2026-08-04 — lifetime achievement replay carry

- Replay-inputs v3 freezes `achievements_earned_lifetime` and `achievement_score_lifetime` in the
  terminal Founder carry. Active runs reject historical v2 evidence rather than inventing lifetime
  state; pre-foundation histories remain v2-replayable.
- Founder reconstruction validates the frozen set/score against the run-pinned catalog, and the
  live output copier now persists the replayed v16 fields. The shared Go/TypeScript fixture and
  explicit version-boundary tests cover the contract; no achievement evaluation/content is claimed.

## 2026-08-04 — replay-carry parity remediation

- Darwin's independent `0ecb479^..0ecb479` review rejected the first carry landing because the
  client verifier had no active nine-artifact state model, active ordinary v2 remained accepted,
  and the shared fixture proved neither active carry nor historical v2 compatibility.
- The remediation implements the complete paired artifact/state boundary in both runtimes and
  adds Go-authored active ordinary and Exit fixtures. A non-empty lifetime set/score is validated
  against the pinned catalog, retained through ordinary replay, and re-derived under a retuned
  next catalog at Exit. Literal production content and earning evaluation remain unclaimed.
- A follow-up closes the activation edge itself: legacy→active Exit also requires replay-inputs
  v3, so changing only a valid activation envelope to v2 fails in both runtimes.

## 2026-08-04 — independent active replay parity verdict

- **Review by:** Darwin
- **Recorded by:** Codex
- **Range union:** `3fa1150^..3fa1150` + `712a3b1^..712a3b1`
- **Decision:** **approved.** Active non-empty carry, exact ID/score derivation, v16 preservation,
  active Exit settlement under a retuned next catalog, paired artifact identity, and the complete
  v2/v3 boundary were independently re-run and held. No remaining findings.

## 2026-08-04 — replay-inputs v4 freezes career observations

Career predicates must inspect immutable Founder carry on every applied transition, not only on
gate or Exit commands. Replay-inputs v4 adds the catalog-coherent lifetime achievement set/score,
age, notoriety, facts, and exit count to every active-run resolved input. Current writes require
v4; historical active v3 rows retain their pinned behavior and pre-foundation v2 rows remain
loadable. This is an input-ownership closure, not a new achievement mechanic. Kernel version is
`0.3.22`; normal root Go/client suites and the regenerated shared replay corpus pass.

## 2026-08-04 — deterministic earning hook and proof enforcement

- Active replay-inputs v4 now evaluate achievements once after the complete applied transition,
  after Meters, from one run/career snapshot. Definitions stage simultaneously in byte order;
  lifetime latches suppress repeats; terminal observations include the settling run's attended
  age, facts, and Exit count before atomic Founder settlement.
- Burn proofs are executable rather than labels: the declared event must exist in the same batch
  and the declared resource must have a debit at least equal to the canonical minimum. Dedicated
  tests discriminate missing-event and missing-debit cases.
- `achievement_earned.v1` is exact-key validated in persistence and admitted by migration `00048`.
  The shared Go-authored ordinary/terminal corpus proves earning, event order, lifetime settlement,
  and next-catalog score retuning in TypeScript. Active v16 receipts now include the run earned set
  and derived score. Kernel version advances to `0.3.23`.
- Verification at the landing boundary used only repository-root targets: `make test` and
  `make test-save-integration` pass, alongside the focused Go/client suites. The implementation
  remains unarchived until an independent review covers the exact committed range.

## 2026-08-04 — independent hook review rejected the first landing

- **Review by:** Darwin
- **Recorded by:** Codex
- **Range:** `9a400f1^..25b7d4d` (a superset of the two implementation commits)
- **Decision:** **not approved.** Burn proof used whole-transition net balance movement, so normal
  idle accrual could hide an honest sink. TypeScript also unioned a run-owned ID already present in
  Founder lifetime while Go correctly failed closed. Both are demonstrated cross-boundary defects,
  and the integration claims did not cover either causal seam.
- **Remediation:** transition output now carries action-only debit evidence measured from the
  post-accrual boundary; replay recomputes it and never persists it. Both kernels consume that
  evidence for burn proof. TypeScript mirrors Go's lifetime-overlap rejection. The reviewer cases
  become focused and shared-corpus regressions before re-review.

## 2026-08-04 — action-debit and lifetime-latch remediation implemented

- Applied transitions derive a non-wire action-debit trace from the post-accrual ledger boundary;
  burn proof consumes that trace rather than whole-transition net movement. A shared Go-authored
  case accrues `1e9` cash and burns `1e9` in the same gate action, then earns identically in TS.
- The discriminating case exposed and fixed a second TS defect: the burn loader stored Decimal.js
  display text (`1000000000`) rather than canonical wire text (`1e9`).
- TypeScript now rejects run/lifetime ownership overlap before Exit union, matching Go and CA4.
  Kernel version is `0.3.24`; root unit/browser suites and real-Postgres integration pass.
