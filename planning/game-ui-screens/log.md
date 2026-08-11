# Game UI Screens implementation log

## 2026-08-10 — Codex acceptance review

The UI Foundation substrate is implemented at `3483ab1`, but the screen draft cannot yet compile
against the repository's actual data boundary. The shipped shell snapshot lacks most Desk/run
fields; required event payloads are not generated typed unions; no literal surface/fact or Copy
manifest exists; first-session bootstrap and local timer/PB persistence are directional; the era
claim contradicts the two-era first hour; and the performance gate has no executable budget.

GU-C1–GU-C8 in the RFC propose exact ownership shapes without choosing product copy or balance
content. Status remains draft; no screen code was improvised.

## 2026-08-10 — Codex contract/composition batch, ready for designated review

Implemented the unblocked mechanical half of GU-C1–GU-C8:

- one generated `game_ui_snapshot.v1` projection over the latest Company revision, its pinned
  replay bundle, the Founder sibling, and composed contribution providers;
- the existing authenticated `GET /api/v1/founder/state` sync route now returns that closed DTO
  instead of exposing raw save bytes; the gameserver composition attaches the real projector;
- exact mechanical-ID-sorted server rows and client fail-closed validation, plus an adapter into
  the existing `AuthoritativeSnapshot` consumed by `ShellRuntime`;
- the five-row surface registry, lifecycle-event cursor precedence, destructive-settings deferral,
  server-sampled monotonic RTA, corrupt-local-PB fallback, closed tier→era mapping, and GU-C8's
  executable performance literals;
- typed decoding for the actually shipped gate/run-end/offer/system event bytes. Declined and
  expired offers normalize to the client `exit_offer_resolved` arm; accepted resolution remains
  blocked by GU-C11 rather than fabricating a missing offer ID.

The projection deliberately rejects schema-v4 or Company-v18 state because the public production
surface cannot assemble content/active-play contributions. GU-C9 records the required kernel-owned
read model. GU-C10 records the idempotent-bootstrap gap exposed by the shipped two-operation account
flow. GU-C12 records the already-requested owner-copy dependency. No Svelte screen has been filled
with placeholder prose.

Focused evidence, all through ordinary root Make targets:

- `make test-go GO_PACKAGES='./gameui ./account ./gameserver' GO_TEST_FLAGS='-count=1'`
- `make test-client` — 6,640 passed, 4 skipped
- `make typecheck` — 0 TypeScript/Svelte diagnostics
- `make test-save-integration` — the full real-Postgres integration suite passed with cache bypass;
  its first run exposed two composition-only seams that focused tests could not: epoch-6 saves
  legitimately omit the pre-schema-v4 provisioned-count map, and `LoadLatest` leaves
  `Revision.OwnerID` empty even though contribution providers require the Founder owner. Both now
  have regression coverage in the composed snapshot path.
- `make api-pin` — generated private-v1 schema/types/compatibility pin refreshed for the existing
  Founder-state route

This batch is not an archival claim. It is ready for the cross-party designated review after its
commit range is recorded; GU-C9–GU-C12 remain explicit blockers to the full screen acceptance gate.

The implementation commit is `1e8f628`. Post-commit `make verify` completed successfully through
the generated API pin, composed harness, TypeScript/Svelte typecheck, production build, 6,640 client
tests, kernel-history guard, copy/schema checks, and 19,932 browser assertions. The designated
review range for this batch is `ddb106f..1e8f628`; it contains only this Game-UI implementation
commit.

## 2026-08-10 — designated cross-party verdict: contract/composition layer {1e8f628} — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
All eight GU rulings verified implemented-or-honestly-deferred: the DTO on the EXISTING route
(which this commit converts from a hand-mounted raw-save-bytes chi route to the registry-mounted
schema-validated operation — closing a save-state escape, the MA-C14 discipline); five-kind exact
event unions with the honest pre-GU-C11 offer handling; RTA/PB per ruling; the five-surface
registry with tested preemption; GU-C8 literals count-asserted with a red probe; BOTH composition
seams genuine with discriminating tests (pre-v18 zero-provisioned; the OwnerID fill that fails
composed-end-to-end when reverted); schema-v4 rejection fail-closed; kernel honest; all gates
green incl. cold -count=1 and the full Postgres suite.
**Routed to the GU-C9/screens batch (non-blocking): C-F1 (MEDIUM) — a store-level test driving
GameUISnapshot itself over a v4 catalog/v18 stream (the predicate alone is tested today); C-F2
(LOW) — wire the CI-observable performance arm when the first screen ships. Plus the copy
ruling's renderer-substitution requirement (mechanical IDs never render) as a GU acceptance test.**
Range D {00cb2a6} records: APPROVED (honest scoping, clean provenance). Combined union:
a8bbd5d..00cb2a6 complete, no uncovered edge commits.

## 2026-08-10 — GU-C12 mechanical assembly review

- **Review by:** Codex (implementer-side acceptance review). **Recorded by:** Codex.
- Mechanical extraction proves 130 affected rows = 129 new keys plus the adopted replacement of
  existing `resource.company_cash.cap.phase0`; no duplicate authored key exists.
- The shipped exit-type union was independently enumerated from both Go and TypeScript as
  `acquihire`, `acquisition`, `collapse`, `ipo`, `scripted_first`.
- Fixed one unambiguous Copy validator defect discovered by the extraction: valid underscore
  placeholders such as `{run_seq}` were mistaken for Markdown emphasis even though Copy C3's
  parameter grammar explicitly permits underscores. The validator now removes only lexically
  valid placeholders before its markup probe; an explicit regression passes `make copy-check`.
- GU-C13–GU-C16 record the four remaining owner-contract gaps: README slot grammar, mandatory
  provenance collisions, exact param types, and presentation-v2/exit-title bytes. No candidate
  SHA is claimed until those are ruled.

## 2026-08-10 — GU-C9 implementation handoff — ready for designated review

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementer handoff, not an
  independent verdict or archival authorization.
- **Implementation commit:** `bc2370a` (the complete GU-C9 implementation range). The production
  kernel now exposes one pure `ProjectRates` read model over the pinned bundle, replay-owned
  Company state, resolved external contributions, and attended coordinate. It assembles the exact
  existing upgrade/ladder/synergy/active-play contribution paths, including the product clamp,
  and returns byte-sorted generator/resource rows without mutating a cursor or producing state,
  events, receipts, or replay inputs. Game UI consumes this seam and no longer reconstructs rates.
- **C-F1 closed:** a real-Postgres test persists a schema-v4, Company-v18 stream and drives
  `GameUISnapshot` itself. Its expected rows require purchased + provisioned counts, the ladder and
  synergy paths, and a live active-play factor; reverting the GU-C9 seam restores the former hard
  rejection and fails the test.
- **R-F1 closed in the same honest kernel bump:** `toFloat64` now crosses a non-inlined float64
  materialization boundary before the snap comparison, preventing arm64 FMA fusion from changing
  the decision. The mandatory shared `floor-fma-snap` vector is consumed by Go and TypeScript;
  `make test-go-ci` proves it on cold Linux/amd64. Kernel `0.3.89 -> 0.3.90` is intentional.
- **R-F3 closed:** `docs/ci.md` now names `make test-go-ci` and its production-architecture scope.

Normal root-target evidence on the exact implementation tree:

- `make test-go GO_PACKAGES='./decimal ./production ./gameui' GO_TEST_FLAGS='-count=1'` — green;
- `make test-save-integration` — every real-Postgres integration package green;
- `make test-go-ci` — every Go package green, cold, on Linux/amd64 with Postgres;
- `make vectors-check` — 6,296 deterministic vectors, 23 mandatory edges, no drift;
- `make verify` — full aggregate green before commit: 6,641 client tests, 19,935 browser
  assertions, zero TypeScript/Svelte diagnostics, and all guard/harness/schema/generated checks;
- post-commit `make verify` repeated every gate through schema; its first browser-server startup
  was denied by the execution sandbox (`listen EPERM`, before tests). The ordinary root
  `make test-browser` target was immediately rerun and passed all 19,935 assertions. No product
  test failed and the committed tree is clean.

GU-C9 is ready for the required cross-party designated review. GU-C10, GU-C11, and GU-C12's
GU-C13–GU-C16 owner blockers remain open; nothing is self-approved or archived.

## 2026-08-10 — GU-C10 acceptance review — owner blockers GU-C17–GU-C22

- **Review by:** Codex (implementer-side acceptance review). **Recorded by:** Codex.
- GU-C10's direction is accepted, but the executable wire, possession boundary, secret-at-rest
  policy, atomic coordinator transaction, receipt expiry, and supposedly mismatched request are
  not determined. The sharp contradiction is exact durable replay of recovery/refresh secrets
  against the shipped hash-only credential law.
- GU-C17–GU-C22 are filed in the RFC with recommended contracts. No coordinator, migration, API
  row, or secret-storage behavior was invented. GU-C10 remains blocked pending owner rulings;
  GU-C11 stays mechanically independent and may proceed after GU-C9's designated handoff.

## 2026-08-10 — GU-C11 implementation handoff — ready for designated review

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementer handoff, not an
  independent verdict or archival authorization.
- **Implementation commit:** `607e5a2`. Accepted offers now emit exact
  `exit_offer_resolved` schema v1 bytes `{offer_id,resolution:"accepted"}` immediately before
  `run_ended` in the same Company revision. Declined and expired rows are unchanged.
- Go and TypeScript produce the same event and order in the shared terminal replay fixture. The
  Game-UI decoder extends its normalized resolution union without manufacturing a run sequence
  for the accepted arm.
- Migration `00072_exit_offer_resolved.sql` extends the closed production event-kind constraint;
  the live Prestige integration test writes the event to PostgreSQL, verifies its exact two-key
  payload, and proves its order before `run_ended`.
- Kernel `0.3.90 -> 0.3.91` is an honest transition/event/replay semantic bump. The canonical
  Prestige documentation records the additive event contract.
- The repository now has a normal `make validate-migrations` target for cold real-Postgres
  validation of migration and save integration tests; no ad-hoc database command is required.

Normal root-target evidence on the exact implementation tree:

- `make test-go GO_PACKAGES='./save ./production' GO_TEST_FLAGS='-count=1'` — green;
- `make replay-fixture-check` — green, including the discriminating accepted-event assertion;
- `make test-client` — 6,641 passed, 4 skipped;
- `make validate-migrations` — green on real PostgreSQL;
- `make test-save-integration` — every integration package green;
- `make test-go-ci` — every Go package green, cold, on Linux/amd64 with PostgreSQL;
- `make verify` — full aggregate green: 6,641 client tests, 19,935 browser assertions, zero
  TypeScript/Svelte diagnostics, and all guard/harness/schema/generated checks;
- post-commit `make verify-kernel-version` and `make replay-fixture-check` — green. A combined
  shell invocation's trailing Docker segment was sandbox-denied before execution; the ordinary
  standalone `make validate-migrations` target was rerun immediately and passed.

GU-C11 is ready for the required cross-party designated review. GU-C10 and GU-C12 remain blocked
only on their recorded owner rulings; nothing is self-approved or archived.

## 2026-08-10 — designated cross-party verdict: GU-C11 {607e5a2, 349757a} — BOTH APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
The ruled contract implemented exactly: emission guarded on the accept intent, appended
immediately before run_ended in the same batch (same revision structurally); ordering
discriminated TWICE in both runtimes (order-swap probes fail the corpus byte-compares and the
in-test assertion); the two pre-existing accepted-offer corpus cases correctly updated in place
with every other case byte-identical; declined/expired paths byte-untouched; both-runtime
lockstep validation with negative tests; migration 00072 append-only, mechanically verified as
00068's Up + exactly the new kind, Down byte-restores; `make validate-migrations` sound and
justified; kernel 0.3.90→0.3.91 honest; all five gates green incl. cold Linux.
**Carried obligation:** when the GU-C3 generator lands, it must own the hand-added union arm.
**Range-union: this verdict covers exactly {607e5a2, 349757a}. NOT covered and pending:**
{bc2370a, 46676e9} (GU-C9 — its own review, launched same day with golden-vector scrutiny) and
{f2cecba, 3e16587} (copy-assembly validation + blocker filings — consumed by the ratification
and ruling flows). The union over 516bddb..HEAD completes when those land.

## 2026-08-10 — GU-C12 copy assembly handoff — ready for designated review and owner ratification

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementation handoff, not an
  independent verdict, owner ratification, or archival authorization.
- **Implementation commit:** `8a8e485`. A deterministic compiler now extracts the 128 ruled
  screen rows, applies GU-C13–GU-C16's exact rewrites/types/provenance, and emits the orphan
  catalog plus presentation/event-copy revisions. `make game-ui-copy-candidate-check` verifies
  those bytes and is a dependency of the ordinary `make copy-check` gate.
- The current tree already contained the separately reviewed permits cap row. It remains
  byte-identical in `permits-candidate.json`; the ruled cash row is retuned in its existing owner
  catalog. Therefore this post-mint assembly adds 133 unique rows: 126 of the original 128 after
  excluding both existing cap keys, plus accepted-offer copy, one gate title, and five exit-type
  titles. This corrects the pre-permits mechanical count without changing either ruling.
- Copy's strict row union has one new arm only: `text_kind: "longform"`. It is printable ASCII,
  at most 80 columns and 64 lines, has zero params, and retains literal README glyphs without an
  HTML/Markdown escape. Build-time and runtime loaders share the rule and reject unknown kinds,
  params, non-ASCII, over-width, and over-line fixtures. Canonical docs are updated.
- The real Horse Armor and shareware-registration claims resolve through two new append-only
  verified registry rows backed by tracked publishable extracts. Mechanical price/offline values
  are placeholders; README numerals are rendered as prose, leaving only those two deliberately
  sourced historical/statistical rows.

Candidate coordinates for owner ratification (SHA-256 over exact file bytes):

- `copy/catalog/game-ui-candidate.json` —
  `ec6d8294837919a03e32f10f0ed81053d43f45306fdc27c8cca6be834de1ea23`
- `copy/catalog/phase0.json` —
  `24e90af2e48db4dbb5e154aea0c3c5ebe0975c48cf9955e15a5a225bfe9df698`
- `balance/testdata/t0-t1/presentation-v2.json` —
  `42cbe31acd040dc6b4a78c8d2b81d7e27738728cec97791414a9537be3d9f015`
- `balance/testdata/t0-t1/event-copy-v2.json` —
  `71d88ebbdff37582246e0014377e2db852e154801e622bf2a04ad12efe6beb0c`
- generated copy artifact identity —
  `sha256:bfc2c7c0051fcbc6c0e76edf796a3dbef01f205eee975605be74f4ae5e8ce86c`

Normal repository evidence on the exact implementation tree:

- `make game-ui-copy-candidate-check` and `make copy-check` — green, including append-only
  provenance history, orphan report, generated types/Go keys, and deployment identity;
- `make typecheck` — zero TypeScript/Svelte diagnostics;
- `make test-client` — 6,642 passed, 4 skipped;
- `make test-browser` — 19,938 passed across Chromium, Firefox, and WebKit. The first full-gate
  run correctly exposed two stale pre-retune cash-copy expectations; the assertions were updated
  to the owner-authored text and the independent browser target passed before the full rerun;
- `make verify` — complete aggregate green: all Go packages/vet, generated API/formulas,
  balance harness, schema/boundary/history/copy gates, client build/tests, and all browser tests;
- post-commit `make verify-kernel-version` and `make copy-check` — green at kernel `0.3.91`.

GU-C12 is ready for the required cross-party designated review and the candidate hashes are ready
for Marco's ratification. No component consumes these orphan rows yet, GU-C10 remains separate,
and nothing is self-approved or archived.

## 2026-08-11 — designated cross-party verdict: GU-C9 {bc2370a, 46676e9} — BOTH APPROVED (follow-ups routed)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
**The seam is genuinely kernel-owned:** ProjectRates assembles through the IDENTICAL private
functions the live/replay paths use (activePlayContributions with the same saturation clamp,
assembleContributions, ratesWithProvisionedAndPolicy, SumDeterministic) — no second math; purity
proven (clone-before-RecordOfflineSpan verified at all three mutation sites; DeepEqual pins);
active-play saturation exercised through the seam; the v4/v18 gate lifted exactly behind it;
**C-F1 DISCHARGED with a real probe** (neutering the attended wiring fails the store-level
Postgres test — the ×7 buff factor drops out of the expected rows).
**The R-F1 decimal fix is genuine and empirically proven both directions on arm64** (the exact
non-ULP-multiple fused diff reproduced; the reviewer's own probes got FMA-fused twice) — and the
discriminating class was FOUND: Floor(8.00002e5) = 800002 unfixed-arm64 vs 800001 everywhere
else — a whole-integer cross-runtime kernel split, now killed. Vector-file scrutiny clean (one
added vector, zero expected-byte changes, byte-identical regeneration). Kernel 0.3.89→0.3.90
honest and required.
**Routed follow-ups (none blocking; F1+F2 should land by the GU-C9 archival gate):**
- **F1 (MEDIUM): the shipped floor-fma-snap vector does NOT discriminate** — reverting the fix
  stays green on both architectures (the vector errs upward; floor is invariant there). Add a
  below-integer-window vector (e.g. floor(8.00002e5) → 800001) and correct the
  docs/numeric-core.md "pins the same decision" sentence.
- **F2 (MEDIUM):** no projection-vs-transition equivalence test — add the property: resource
  rate × Δt == transition accrual delta for the same state/inputs.
- **F3 (LOW):** the fix uses //go:noinline rather than the spec-blessed explicit float64()
  conversion — switch or record the deliberate choice. **F4 (LOW, pre-existing):** the frozen
  external-contribution wiring has no rate-asserting test (nil-provider everywhere).
**RANGE-UNION CAUTION (recorded for the ratification flow): f2cecba is NOT pure ruling text** —
it carries an executable copy-pipeline validator change (placeholder-stripping before the
Markdown scan; inspected sound, fixture-exercised). The GU-C12 ratification verdict MUST cite it
as code or the union over 516bddb..607e5a2^ has an unreviewed executable commit. 3e16587 is pure
ruling text, correctly tiered.

## 2026-08-10 — GU-C10 bootstrap coordinator handoff — ready for designated review

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementation handoff, not an
  independent verdict or archival authorization.
- **Implementation commit:** `9c23e80`. **Designated-review range:** `b1c82b1..HEAD` (the record
  commit completing this entry is docs-tier; the sole implementation commit is `9c23e80`).
- The generated authority now owns `create_bootstrap`, `POST /api/v1/bootstrap`, the exact
  request/201 response, and the closed bootstrap error details. Runtime mounting uses the same
  registry entry; the old account/session routes remain compatibility surfaces.
- One transaction now creates the account, active Founder, Company+Founder streams/genesis/run
  pin/frozen contributions, session family and token pair, transaction-local initial
  `game_ui_snapshot.v1`, and encrypted retry receipt. The response is registry-validated before
  insert/commit. The composed test asserts the transaction-local snapshot is byte-identical to
  the first committed GET.
- Bootstrap keys require exactly 32 random bytes encoded as 64 lowercase hexadecimal characters.
  Only SHA-256 of that string is stored. Equal-key requests serialize through a transaction
  advisory lock; concurrent first calls produce one account and byte-identical 201 receipts.
- Receipt plaintext is AES-256-GCM protected with digest+key-ID associated data. API config owns
  current/previous key IDs and clones the material on construction. Tests prove old-key replay
  through rotation and fail closed on ciphertext/digest tampering; database probes confirm the
  recovery code and both tokens are absent from stored ciphertext.
- Refresh expiry performs the sole live-to-tombstone transition: ciphertext, nonce, and key ID
  are destroyed while request digest/account coordinate remain. Bounded `SKIP LOCKED` GC performs
  the same transition; a trigger forbids deletion or reversal, and account deletion tombstones
  live receipts before removing credentials.
- Migration `00073_bootstrap_receipts.sql` is append-only. The production binary now requires a
  separate deployment-resolved `CLOUD_CLICKER_BOOTSTRAP_KEY_ID` + base64
  `CLOUD_CLICKER_BOOTSTRAP_KEY`; canonical gameserver docs wait for the Game UI archival move.
- Fault injection covers all eleven coordinator boundaries. Every injected failure leaves zero
  account/founder/stream/family/token/receipt residue; registry-response rejection is separately
  proven to roll back. Exact retry, key rotation, concurrent first calls, expiry, GC, permanent
  tombstones, invalid public keys, and composed HTTP success are all exercised on real Postgres.

Normal repository evidence on the exact implementation tree:

- focused Go packages and bootstrap schema/crypto tests — green;
- `make validate-migrations` — green;
- focused real-Postgres coordinator fault matrix and composed gameserver lifecycle — green;
- `make test-save-integration` — complete single-run, cold `-count=1` Postgres suite green across
  all packages. An earlier accidental overlapping pair of full runs was explicitly discarded
  after the two processes interfered by truncating their shared test database; neither partial
  output was treated as a verdict;
- `make api-check` and `make typecheck` — green;
- `make test-go-ci` — complete cold Linux/amd64 suite green, explicit exit 0;
- `make verify` — complete aggregate green, including kernel 0.3.92 history guard, 6,642 client
  tests, generated artifacts, harness/schema/boundary/copy gates, and 19,938 browser assertions.

GU-C10 is implemented and ready for the required cross-party designated review. GU-C12 remains a
separate ratification/review thread, the screen components do not consume bootstrap yet, and
nothing is self-approved or archived.

## 2026-08-11 — designated cross-party verdict: GU-C10 {9c23e80, 81114cf} — BOTH APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
Contract-exact against all six rulings: the C17 wire closed-set proven over real HTTP (201 +
no-store both paths; no idempotency_conflict reachable); C18 digest-before-BeginTx statically
provable with no plaintext key column anywhere; C19 AES-256-GCM with associated data binding
digest‖key_id (the digest IS the possession proof; wrong-digest decryption fails by
construction), two-key rotation replay proven, coordinator-level raw-byte replay equality; C20
one transaction with ALL ELEVEN fault boundaries enumerated + 7-table residue checks — the
reviewer's own sabotage probe was caught at exactly the injected boundary; C21 tombstone
semantics ENFORCED STRONGER THAN RULED (the migration CHECK makes ciphertext-absent-without-
tombstone unrepresentable; the trigger forbids DELETE and reversal — both SQL-tested); C22's
dead arm verified absent. Migration 00073 append-only; kernel 0.3.92 honest (the range's watched
touch is a real fail-closed tightening); the GET projector refactor byte-equal by construction
and empirically. All five gates green incl. cold Linux.
**Routed (none blocking): F1 (MEDIUM, due by the Game-UI archival gate) — the receipt GC has NO
production caller** (sessions get a periodic job; receipts don't; never-retried ciphertext
outlives expiry indefinitely) — wire a NewPeriodicJob beside the session GC. F2 (LOW) the
DeleteAccount tombstone branch untested under a live receipt; F3 (LOW) the composed replay check
compares re-marshaled envelopes (raw-byte equality separately proven at coordinator level).
**Forward note:** replay revalidates decrypted receipts against the CURRENT registry — a future
GameUISnapshot shape change turns old live receipts into 500s (fail-closed, acceptable, remember
at the next schema move).
**Range-union: consumes exactly {9c23e80, 81114cf}. {8a8e485, b1c82b1} are the GU-C12 thread —
pending its designated verdict (which MUST cite f2cecba as executable code) + owner ratification.**

## 2026-08-11 — designated cross-party verdict: GU-C12 assembly {f2cecba, 8a8e485, b1c82b1} — APPROVED (two owner dispositions gate ratification)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
**f2cecba consumed AS CODE** (independent re-verification + ten seeded probes — the standing
caution closed). **Text fidelity: 131 rows byte-checked against the ruling — zero unexplained
drift outside the findings**; all 24 param types exact; the longform kind contract-exact with
negative fixtures in both pipeline and runtime; provenance clean (two claims, tracked extracts,
HTTPS, zero suppressions — no suppression mechanism even exists); presentation v2 byte-proven
additive over the ratified v1; orphan-first held; all SHAs recomputed twice; kernel honestly
untouched; all gates green at the pinned worktree.
**F1 (HIGH, owner disposition):** three README lines beyond the authored rewrite — two
detector-FORCED but implementer-worded ("mail us zero dollars"; "nineteen ninety-five"), one
UNFORCED ("slide 1 of 1" → "the only slide", probed clean both ways, and inconsistent with the
retained Vision Slide footnote). **F2 (MEDIUM, owner disposition):** the permits-cap ruling rested
on a stale premise (the PT-C4 county line was already the landed OWNER text); Codex kept the
county line — defensible, but which owner text wins is the owner's call. **F3 (LOW):** seven
implementer-chosen tones to adopt. F4/F5 observations recorded.
**Range union 516bddb..81114cf CLOSED** — all 13 commits covered across the four verdicts, no
uncovered edges. Ratification package (pending dispositions): copy candidate ec6d8294… ·
phase0 24e90af2… · presentation-v2 42cbe31a… · event-copy-v2 71d88ebb… · generated bfc2c7c0….

## 2026-08-11 — GU-C12 owner-disposition reassembly — ready for mechanical ratification

- **Implemented by:** Codex. **Recorded by:** Codex. This is a narrow disposition handoff, not
  owner ratification or archival authorization.
- **Implementation commit:** `2bb0f71`. The assembler no longer performs the unforced
  `slide 1 of 1` → `the only slide` replacement. Reassembly restores the owner-authored README
  line exactly to `(see: OUR VISION, slide 1 of 1)` while retaining both adopted detector-forced
  prose rewrites. No other ruled row changed.
- The three reviewed documents remain byte-identical at their approved coordinates:
  `phase0.json` `24e90af2…698`, presentation-v2 `42cbe31a…015`, and event-copy-v2
  `71d88ebb…b0c`.
- **Refreshed owner-ratification pins:**
  - `copy/catalog/game-ui-candidate.json` —
    `7793d334efb8e9c6ba6169466ec51a1b8318145f470b643a849937943725b4f2`
  - generated copy identity —
    `sha256:0f9801e40718d440ce25966d965da2cfff44b99216638dae5afa014b01a25d12`
- Normal drift-gated evidence: `make game-ui-copy-candidate-check`, `make copy-check`, and the
  complete post-commit `make verify` all green; the aggregate included 6,642 client tests and
  19,938 browser assertions. Kernel remains 0.3.92; the change is copy/generator-only.

The two refreshed hashes are ready for Marco's mechanical ratification. Screen components remain
unstarted until that ratification; nothing is self-ratified, archived, or pushed.

## 2026-08-11 — screens implementation DESIGN-GAP: Offer acceptance lacks the Founder CAS coordinate

- **Recorded by:** Codex. This is an implementation-time blocker, not a ruling or completion
  claim. The ratified screen copy and the five surfaces are otherwise being implemented normally.
- `accept_exit_offer` requires `expected_founder_revision`, but `game_ui_snapshot.v1`, bootstrap,
  and `exit_offer_spawned` expose only the Company revision. A resumed client therefore cannot
  construct a sound acceptance intent. Using `1` (or copying the Company revision) would be an
  unruled client-side authority guess and is rejected.
- The current screen fails closed: acceptance remains disabled until a Founder-scoped event has
  supplied an observed revision. Decline remains usable because it is Company-only.
- **GU-C23 proposal:** widen the generated snapshot to `game_ui_snapshot.v2` with one exact
  positive `founder_revision` field, resolved transactionally by the server for bootstrap and
  ordinary sync. The client resamples it on every snapshot and advances it only from decoded
  Founder-scoped event coordinates. Offer acceptance submits that exact coordinate. This keeps
  the existing intent and transport channel, creates no second authority, and makes reconnect
  behavior sound. Owner ruling is required before the accept button can be claimed complete.
- Separate acceptance debt: AC1's exact T0→T2 script depends on the owner-ratified epoch-7
  content candidate, which is not minted in the live epoch-6 bundle. This batch will prove the
  real composed bootstrap/intent/snapshot browser path without fabricating candidate mechanics;
  the full scripted progression remains visibly mint-gated.

## 2026-08-11 — screens implementation DESIGN-GAP: payout terms have no presentation rows

- **Recorded by:** Codex. This is an implementation-time copy blocker, not a ruling.
- U1 requires the Offer Sheet to render the complete `payout_preview` terms and the run-end delta
  list. The decoded wire correctly supplies reputation delta, route knowledge, network-slot
  unlocks, and the Clout-reach note, but the ratified GU-C12 package supplies only the Offer
  heading/actions/countdown and one generic run-end delta heading. It has no legal labels or row
  frames for any of those four term kinds, and network slots have no presentation binding.
  Rendering JSON, mechanical IDs, or implementer-authored labels would violate the zero-literal
  and renderer-substitution laws.
- **GU-C24 proposal:** owner-author four byte-sorted term-row copy bindings shared by Offer Sheet
  and run-end (`reputation_delta`, `route_knowledge`, `network_slot_unlock`, `clout_reach_note`),
  with a presentation binding for every shipped slot/carried-ref ID. The existing decoded terms
  object remains the sole data source; no wire or gameplay change is needed. Until ruled and
  ratified, the screen shows the exit type and truthful expiry but cannot claim the complete
  terms/delta-list arms.

## 2026-08-11 — screens self-review: Window/Worker monotonic origins could reject honest sync

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementer finding and fix,
  pending the batch's designated review.
- Composing `game_ui_snapshot.v1` into the archived `ShellRuntime` exposed a dormant browser race:
  `PredictionWorkerClient` stamped initialization and authoritative refreshes with Window
  `performance.now()`, while the prediction loop compared those samples to Worker
  `performance.now()`. Different time origins (and ordinary message delay after a Worker pulse)
  made an honest authoritative refresh throw `non-monotonic authoritative clock` in real browsers.
- The Worker now owns every monotonic sample it compares: initialize and authoritative messages
  carry snapshot bytes only, and the Worker samples its clock when processing each command. A
  cross-Chromium/Firefox/WebKit browser test composes the real Worker, publishes a second
  authoritative UI snapshot, and proves prediction continues. Reverting either ownership change
  reproduces the uncaught rollback error.
- These are presentation-only Client Shell paths outside the KV-1 watched set; no gameplay kernel
  version bump is owed. Canonical Client Shell documentation records the clock ownership.

## 2026-08-11 — Phase-A screens implementation slice ready for designated review

- **Implemented by:** Codex. **Recorded by:** Codex. This is a handoff, not an approval or an
  archival claim. Code range: `c9fc04f..9a97543`; this planning record is docs-only and belongs
  to the same designated-review range.
- The production Svelte entry now mounts all five Phase-A surfaces over the generated
  `game_ui_snapshot.v1`, the archived prediction Worker, decoded lifecycle events, the bootstrap
  coordinator, existing intents, and the ratified Copy/Presentation bytes. Era follows the
  authoritative tier without replacing persistent focus; presence stays hidden until a real
  world count arrives; run-end accepts only its decoded terminal event.
- The renderer-substitution lint and browser proof, C11 axe pass over every surface/system beat,
  local RTA/PB/split lifecycle, cross-browser Worker composition, deterministic 20 Hz/10 Hz
  performance arm, raw-socket fail-closed decoding, and composed bootstrap-receipt GC are all in
  the range. GU-C9's four routed proof items are closed by the discriminating Floor vector,
  non-neutral provider test, and projection-versus-transition equivalence property.
- **Normal test evidence, read through final exit:** `make test-client` (6,646 passed, 12
  skipped); `make test-browser` (19,972 passed, 2 skipped across Chromium/Firefox/WebKit);
  `make test-go GO_TEST_FLAGS='-count=1'`; `make test-save-integration` on real Postgres;
  `make test-go-ci` on cold Linux/amd64; standalone `make verify-kernel-version` (kernel
  `0.3.92`); and the complete `make verify` aggregate all exited 0. Copy assembly reverified the
  ratified presentation SHA `42cbe31a…f015` and its byte-identical client mirror.
- **Scope remains deliberately incomplete:** GU-C23 and GU-C24 above require owner rulings. Offer
  acceptance therefore fails closed and the complete payout/delta rows do not render. AC1's exact
  T0→T2 browser script also remains epoch-7-mint-gated; this range does not substitute fixtures for
  live content or claim the aggregate plan boxes closed. The implemented slice is ready for the
  cross-party designated pass; it is not archival-eligible on this record.
