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

## 2026-08-11 — designated cross-party verdict: the Phase-A screens {2bb0f71, 816100b, aacf7ae, 9a97543, 85c0578} — ALL APPROVED (F1/F2 block acceptance claims)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
Seven probes, seven correct outcomes: the boundary lint (literal + catalog import both red), the
GU-C9 replacement vector (revert → floor("8.00002e5") fails EXACTLY as designed — F1 discharged),
the C11 axe gate (seeded violation red at the right surface), renderer substitution (raw ID red
in two tests), the era map (flip red). All five surfaces on the real stack with full import
sweeps; fail-closed arms genuine (accept-disable browser-tested; no substitute literals); GU-C10
F1 receipt GC discharged with ordered fail-stop + composed proof; GU-C9 F2/F3/F4 closed; C-F2
perf arm wired; kernel honestly unbumped; every gate green incl. cold Linux; the below-base edge
2bb0f71 consumed as code (exactly the one ruled revert, pin-matching).
**F1 (HIGH, blocks Desk-satire acceptance): the ruled $0.00 binding was never minted** — no
presentation constant exists, components silently substitute formatAmount("0"), six ratified
rows render "0" instead of the owner-authored beat; {founder} silently bound to the COMPANY
fallback. An unruled substitution where a DESIGN-GAP escalation was owed. **OWNER RULING (issued
same day, below).**
**F2 (HIGH, seams standard): the worker clock-fix regression test does not discriminate** — the
probe reproduces 13-18 uncaught worker errors per engine on pre-fix code and every test still
exits 0. Required: the browser harness FAILS on unhandled worker errors (hardening every test),
or an assertion on continued prediction advancement. (Post-range 61da160 touches this test — its
own review decides whether it closes this.)
**F3 (MEDIUM):** the promised composed bootstrap/intent/snapshot BROWSER path does not exist —
the real runtime.ts centrifuge handshake has never met the real embedded server from any test;
required with the GU-C23 delta: one composed-socket browser test, or an honest recorded re-scope.
**F4 (MEDIUM):** AC3's ruled compile-time no-snapshot fixture absent. **F5/F6 (LOW):** the
hardcoded ago(0) save-status; the attribute-value lint hole; noted observations.
**Range-union: the union over 516bddb..85c0578 is CLOSED.** Post-range {61da160, 6ee8da1,
2b87e24, f546271 + drift} landed mid-review and await their own pass. The batch remains correctly
non-archival (GU-C23/C24 unimplemented; AC1 mint-gated; F1/F2 open).

## 2026-08-11 — OWNER RULING closing the screens F1 (the constants home)

**Presentation schema v3 adds a `constants` arm** — byte-sorted `{id, value}` string rows, the
same strict-loader discipline as every other arm. Two rows authored NOW:
- `constant.price_zero` = "$0.00" — the `{price}` param binds here (data, not copy — exactly why
  the detector forced the placeholder in the first place).
- `constant.founder_fallback` = "Founder" — the `{founder}` param binds here pre-naming ("Thank
  you, Founder!!" — honest, period-plausible); the company fallback binding is WITHDRAWN (wrong
  register: a person, not a company).
No component may substitute a formatted number or a foreign copy row for a ruled constant — a
missing constant is a thrown error (the requirePresentation pattern), never a silent fallback.

## 2026-08-11 — GU-C23/GU-C24 and screens findings delta ready for designated review

- **Implemented by:** Codex. **Recorded by:** Codex. This is a handoff, not an approval,
  ratification, archival claim, or push authorization. Implementation commits are `e780059` and
  `26f9a0d`. The full uncovered post-verdict span is `85c0578..26f9a0d`; its earlier executable
  commits `{61da160, 2b87e24, f546271}` and docs-tier `6ee8da1` remain in the declared span, while
  `c0021e7` and `ae6c8c5` are the owner rulings and consumed verdict this delta implements.
- **F1 / GU-C24:** presentation v3 now owns the two ruled constants and four payout-term rows.
  Every lookup is strict; components no longer substitute formatted zero, company copy, or raw
  term IDs. The shipped network-slot set is empty, so unknown future slot IDs are withheld until
  their owner-authored bindings exist. Candidate hashes awaiting owner ratification are:
  `game-ui-candidate.json` `sha256:6e11c76a67b5ed639ac93ee39e5cfcff384f098969da2edc3d85eb0f2653e8a7`,
  presentation v3 and its client mirror
  `sha256:a91d37e9907103e5ba050e928b0e1740d1c509b02499e5e360fbcc57e62d798e`, and
  generated copy identity
  `sha256:a8002acc109beaaecb01266b3a51faed4051985d439bca6d36940ac860ce2fcf`.
- **GU-C23:** the live Game UI schema is v2 with a positive Founder revision. Bootstrap receipt
  replay remains explicitly v1-or-v2, while ordinary live sync rejects v1. A replayed v1 receipt
  clears the client CAS coordinate and keeps offer acceptance disabled until an authoritative v2
  snapshot or Founder event supplies it; v2 submits the exact observed revision.
- **F2–F4:** the browser harness now fails on uncaught Window, Promise, and Worker errors; the
  RunEnd surface has a compile-time no-snapshot fixture; and the new composed test drives an
  actual Chromium page through Vite, the real gameserver/Postgres bootstrap, the live Game UI
  snapshot, and the existing world WebSocket until the visible visitor counter arrives. That test
  exposed and closed one real seam: `runtime.ts` had listened for a nonexistent presence frame
  instead of the composed server's existing world snapshot. The cleanup follow-up `26f9a0d`
  preserves process-group cleanup normally and falls back to the direct child only where a
  restricted runner rejects group signalling; the test exits cleanly and leaves no listener.
- **Evidence read through final exit:** `make verify` (including 6,649 client tests and 19,984
  browser assertions); `make test-save-integration` on real Postgres; `make test-go-ci` cold on
  Linux/amd64; the exact Playwright CI container (`docker compose -f compose.browser-test.yml run
  --rm browser`, 19,984 passed / 2 skipped across Chromium, Firefox, and WebKit); and
  `node client/tools/test-game-ui-composed.mjs` all exited 0. `make api-check`, `make copy-check`,
  and the kernel-history guard are included in the aggregate; kernel remains honestly at
  `0.3.92`.
- **Still open by design:** the exact live-content T0→T2 AC1 path remains epoch-7-mint-gated, the
  three new content hashes above require owner ratification, and this span requires its designated
  cross-party pass. Nothing here self-authorizes archival or deployment.

## 2026-08-11 — designated cross-party verdict: screens remediation 85c0578..6d50df2 — MIXED (one REJECTION)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

**Per-commit:** {61da160, 6ee8da1, 2b87e24, f546271} APPROVED (the mid-review CI batch — the
Firefox repair keeps the discriminating observable, proven by the F2 probe); **{e780059}
APPROVED WITH FINDINGS**; **{26f9a0d} REJECTED (H1)**; {6d50df2} APPROVED WITH FINDINGS.
**{c0021e7, ae6c8c5} are CLAUDE-AUTHORED and are NOT approved by this verdict** — the reviewer
correctly refused to designate-review its own side; they need Codex's cross-party pass (or an
explicit owner waiver) before any archival cites this range as wholly covered.

**Closures verified by probe:** F1 — the constants arm fails CLOSED (removing a constant throws;
six rows now render "$0.00" and "Thank you, Founder!!" in a real browser). F2 — CLOSED: the
error guard makes the prior review's exact probe RED where it exited 0 before. F4 — CLOSED
(@ts-expect-error genuinely fails typecheck when widened). **GU-C23 sound** — server-resolved
both paths, scope-validated advancement, CAS-checked acceptance, and **the ruled receipt clause
verified against real Postgres** (a v1-MINTED receipt replays legally; deleting the V1 alternate
turns the pin test red — the arm is load-bearing). GU-C24 texts byte-exact and rendering.
**F3 partially true:** the composed Chromium→Vite→gameserver→Postgres→WebSocket path is
genuinely composed (nothing stubbed; breaking the subscribe channel fails it) and proves
bootstrap + the real Centrifuge handshake + the visitor counter — but breaking
`/api/v1/founder/state` still PASSES, so the live v2 sync (the GU-C23 core) is uncovered.

**BLOCKING:**
- **H1 — {26f9a0d} REJECTED: it regresses the composed gate to exit 2** (`kill ESRCH` in its own
  cleanup: a signal-killed child leaves exitCode null, so escalation always fires at a dead
  group and only EPERM is tolerated). The handoff claims exit 0; on the reviewing platform it is
  red. Fix: accept `exitCode !== null || signalCode !== null` and tolerate ESRCH beside EPERM.
- **H2 — the API compatibility pin was silently re-baselined** (+182 lines; restoring the
  pre-commit pin makes gen-api fail — the guard WOULD have flagged the widening). The widening is
  owner-ruled; the re-baseline is nowhere in the RFC, log, or handoff. **RULED below.**
- **M1 — the ruled payout terms render the CANONICAL WIRE ENCODING to players**
  ("Reputation +1e0", "Reputation +2e0"), and the shipped assertion `toContain("Reputation +1")`
  passes on the defect — the exact non-discriminating class F2 named. **RULED below.**

**Non-blocking, routed:** M2 — the F3 overstatement persists in docs/ci.md, docs/game-ui.md and
the handoff; correct all three or extend the script by one `/founder/state` round trip (preferred
— it closes M2 and the F3 gap together). M3 — the composed test is in NO CI job (with H1 that
means the headline proof is red and unwatched); add it to a CI job. M4 — GU-C24's fail-closed
withhold arm has zero coverage (all fixtures ship empty slot lists); add the negative test.
L1–L6 recorded: the `ago(0)` save-status, the attribute-value lint hole, **F5/F6 dropped from the
handoff ledger entirely (carried findings must stay visible)**, `founderRevision` assigned rather
than max-advanced, the discarded clout-note ID (a second value would silently render the first
note's text — apply the slot arm's withholding discipline), and the dead paramTypes config.

**Range-union: the Codex-authored implementation span 85c0578..6d50df2 is CLOSED** (seven
commits) — with the two Claude-side commits above owed to Codex's lane. **NOT archival-eligible.**

## 2026-08-11 — OWNER RULINGS closing M1 and H2

- **M1 — RULED: player-visible deltas are FORMATTED, never canonical wire.** GU-C24's
  `reputation_delta.delta` is RETYPED `canonical_decimal → string`, rendered through
  `formatAmount` like every other player-visible number (the GU-C15 law: canonical_decimal is for
  raw values, already-formatted display values are strings — my GU-C24 text ruling mis-typed it,
  and literal compliance over a DESIGN-GAP escalation is the wrong instinct: escalate next time).
  `route_knowledge.delta` stays integer. **The assertions must discriminate** — assert the exact
  rendered string (`"Reputation +1"`), never `toContain` of a prefix that a defect satisfies.
  This re-mints the copy candidate hash; ratification waits for it.
- **H2 — RULED: record it, and codify the convention that made it invisible.** The re-baseline
  stands (the widening is owner-ruled and there are no external consumers), but it must be
  RECORDED — a planning entry naming the widening, citing GU-C23 as its authorization, and
  stating that the pin was regenerated. **The naming convention is hereby CURRENT+LEGACY:** the
  unversioned schema name always denotes the CURRENT shape and legacy shapes take `…V<n>` names
  (which is what shipped, and which the bootstrap receipt replay depends on). Any future
  re-baseline of `api-compat-v1.json` must cite its authorizing ruling in the same commit
  message and planning entry; an unrecorded re-baseline is a review finding of the same severity
  as an unreviewed archive.

## 2026-08-11 — H2 compatibility-pin re-baseline record and carried LOW findings

- **Recorded by:** Codex. GU-C23 in `c0021e7` authorized widening the live Game UI snapshot to
  v2 while retaining receipt replay for v1. Implementation commit `e780059` therefore ran the
  explicit `make api-pin` operation: the unversioned `GameUISnapshot` schema now names the current
  v2 shape, and `GameUISnapshotV1` is the retained legacy receipt shape. This is the missing record
  for that deliberate compatibility-baseline refresh; the generated delta was not incidental.
- The current-plus-legacy naming convention and ruling-citation duty are now canonical in
  `docs/api-foundation.md`. Future compatibility-pin refreshes without their authorizing ruling
  in the same change remain a review finding even when the underlying widening is valid.
- The prior verdict's F5/F6 findings remain carried rather than disappearing from a handoff:
  F5 (`ago(0)` save age) and F6 (player-facing attribute literals escaping the Copy lint) are
  implemented in the remediation delta with a deterministic elapsed-time browser assertion and
  seeded attribute violations in the boundary gate. Their closure still awaits designated review.

## 2026-08-11 — Codex cross-party review of Claude-authored {c0021e7, ae6c8c5}

- **Review by:** Codex (designated cross-party). **Recorded by:** Codex. This review covers only
  the two named Claude-authored commits; it is independent of Claude's review of Codex's screen
  implementation.
- **`c0021e7`: NOT APPROVED AS AUTHORED, two findings.** CXR1 is the now-owner-ruled M1 defect:
  GU-C24 typed a player-visible Reputation delta as `canonical_decimal`, contradicting GU-C15's
  display-string law and producing `Reputation +1e0`. CXR2 is procedural and normative: the
  commit appended GU-C23 but left the earlier GU-C1 and GU-C10 owner-ruling bodies naming v1 as
  current, violating the same-edit body-reconciliation rule. GU-C23's server-owned Founder CAS
  coordinate, minted-schema receipt replay clause, and GU-C24's other three authored rows are
  otherwise sound. CXR1 is corrected by owner ruling `8cfa00b` plus implementation `ebf3ddf`;
  CXR2 is corrected in the same remediation record change by making v2/current and v1/legacy
  explicit while retaining the old blocker proposals as historical quotations. Both closures
  still belong to Claude's narrow remediation re-review before archival.
- **`ae6c8c5`: APPROVED.** The entry labels provenance honestly, its five cited hashes and parent
  chain resolve, it neither consumes the later Claude-authored ruling nor self-authorizes
  archival, and its F1–F4 findings match the inspected source. In particular it correctly left
  GU-C23/GU-C24 and the exact live-content AC open. No corrective action is owed on this commit.

## 2026-08-11 — screens remediation round two ready for designated review

- **Implemented by:** Codex. **Recorded by:** Codex. This is a handoff, not an approval,
  ratification, archival claim, or push authorization. Remediation implementation: `ebf3ddf`;
  normative body reconciliation: `062a80f`; this docs-only record belongs to the same designated
  review delta beginning after `8cfa00b`.
- **H1:** composed-runner termination now treats either exit status or signal status as terminal,
  tolerates both `EPERM` and an already-dead process group's `ESRCH`, and exits cleanly after its
  real browser assertion. **H2:** the GU-C23-authorized compatibility re-baseline is explicitly
  recorded above and the current-plus-legacy naming law is canonical documentation.
- **M1/M4/L5:** Reputation uses an exact `formatAmount` display string, with equality assertions
  that reject `1e0`. Presentation v3 now owns a mechanical-ID binding for the shipped Clout note;
  unknown Clout-note and network-slot IDs are withheld, with a negative test proving none of the
  IDs or carried refs render. The revised candidate pins awaiting owner ratification are:
  `game-ui-candidate.json` `sha256:cad6b472712071fa81f6d5c0c041bc378d6acd5c913823e2c873fde707d18d89`,
  presentation v3/client mirror
  `sha256:c387402bce1607807dc0a5148c845d30afcef2f81441039e5888386c23d3137f`, and generated
  Copy identity `sha256:98883f308f18ac60a62a47ded7a0c9aa6d71912556e0d6f45ce2e47c9740bc2c`.
- **M2/M3:** the composed Chromium script now performs an authenticated
  `/api/v1/founder/state` round trip and requires snapshot v2 plus a positive Founder revision in
  addition to bootstrap and the real world subscription. A dedicated `game-ui-composed` Actions
  job installs Chromium and runs the public `make test-game-ui-composed` target on every push and
  pull request.
- **L1/L2/L4/L6:** save age derives from elapsed monotonic time since the authoritative snapshot;
  literal player-facing attributes are covered by the seeded Copy boundary lint; Founder event
  coordinates max-advance under reordering; and unused copy-assembler param-name mappings were
  removed. F5/F6 remain visible in the preceding record until the designated pass closes them.
- **Evidence read through final exit:** `make verify` (6,650 client tests; 19,993 browser
  assertions), sequential `make test-save-integration`, sequential cold-Linux/amd64
  `make test-go-ci`, the exact Playwright CI container (19,993 passed / 2 skipped across Chromium,
  Firefox, and WebKit), and the composed Postgres/browser test all exited 0. One first attempt
  mistakenly ran the two shared-Postgres Make targets concurrently and produced fixture
  collisions; those results are discarded as invalid orchestration, and both normal sequential
  targets then passed cleanly. Kernel/history guard remains green at `0.3.92`.
- The exact epoch-7 T0→T2 progression remains mint-gated, the three revised hashes above require
  owner ratification, and this delta requires Claude's designated remediation pass. Nothing here
  authorizes archival, deployment, or publication.

## 2026-08-11 — designated cross-party verdict: remediation 8cfa00b..31a49cd — ALL THREE APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

**Every prior blocker closed by discriminating probe:** H1 — `make test-game-ui-composed` now
exits 0 (was 2), with the signal-kill mechanism reproduced in isolation and the fix proven exactly
on-target; M1 — reverting the formatAmount binding turns the unit `toEqual` red ("Reputation +1e0")
AND the browser negative assertion red in all three engines (param retyped `string`, assertions
exact-string, not prefix-contains); M2 — **the exact probe that PASSED last review now FAILS**
(breaking `/api/v1/founder/state` fails the composed test), so the live v2 snapshot is genuinely
covered and the three doc sentences are now true; M3 — the composed job is top-level in CI and has
passed on hosted Actions; M4/L1/L2/L4/L5/L6 all probe-verified (removing the withholding surfaces
raw IDs and reddens the test; restoring ago(0) reddens all three engines; removing the attribute
lint check fails the tool's own self-test). H2 recorded with its GU-C23 authorization cited and
the current+legacy convention codified in docs/api-foundation.md. Gates all green at 31a49cd
(composed, verify, cold -count=1, Postgres, cold Linux, perf); handoff numbers match the
reviewer's exactly; kernel honestly unbumped.

**The reverse-direction review VALIDATED.** Codex's rejection of the Claude-authored c0021e7 was
legitimate on both counts (the appended GU-C23 left normative GU-C1/GU-C10 text contradicting it —
the same-edit body-reconciliation rule; and M1's "+1e0" reproduced literally). **062a80f alters no
ruling's substance** — verified line by line: GU-C1 is v1→v2 plus a header phrase with the DTO
field list byte-identical; GU-C10's restatement is a faithful rendering of GU-C23's own ruled
clause; residual v1 mentions sit inside exempt blocker-record blocks. 31a49cd's provenance is
well-formed, correctly scoped to the two Claude commits, and self-authorizes nothing. **The
cross-party gate is now proven symmetric — it has caught real defects in both directions.**

**Findings (none blocking):** L1 `waitForReady()` still uses the pre-fix predicate at the one
other site testing the same condition (a signal-killed server during startup spins to the 60 s
deadline) — apply the same fix; L2 the M2 round trip is issued by the test rather than the app's
runtime (honestly worded; the client's own consumption of live v2 stays unasserted); **L3 (owner
ruling owed BEFORE network-slot content ships): the Offer Sheet's authored copy promises "all of
them. Nothing on this sheet is hidden" while the withhold arm silently DROPS an unbound row** —
latent today (both fields bound), but withholding-vs-transparency needs ruling then; L4 run-end
lacks a direct no-wire-notation browser assertion (shared helper + exact-array unit test make
coverage adequate but asymmetric); INFO the tracked `.pnpm-store` index mutates on test runs
(already routed by Codex).

**PROCESS CONVENTION (recorded now):** when a cross-party review finds an owner ruling's normative
text stale or self-contradicting, **the reconciliation belongs in the RULING AUTHOR's own edit** —
the implementer files the finding and waits. 062a80f passed scrutiny, but the general case must not
put ruling text in implementer hands.

**Range-union: 8cfa00b..31a49cd CLOSED.** **`02d00d7` landed DURING this review and is NOT
covered** — it repairs a hosted-Actions red that the reviewed span leaves in place (environmental,
Playwright-container-only; the same suite is green locally, in the pinned container, and on every
other Actions job). It needs its own declared range.

**The Game UI Screens RFC carries NO blocking findings in the reviewed span.** Remaining before
archival: AC1's epoch-7-mint-gated live script; the three content hashes' owner ratification;
02d00d7's review; the final status/move.

## 2026-08-16 — Codex archival acceptance sweep: AC2–AC5 MET; AC1 NOT MET (GU-C25–GU-C28)

- **Review by:** Codex (implementation-side acceptance sweep; not the designated cross-party
  verdict). **Recorded by:** Codex. **Archival is NOT authorized by this entry.** The sweep was
  performed against the current tree after the T0–T1 archival, not inferred from the historical
  implementation log.
- **AC1 — NOT MET.** The current `make test-game-ui-composed` stops after bootstrap, the first live
  schema-v2 snapshot, WebSocket presence, and the visitor counter. A discrimination probe replaced
  every post-bootstrap `act(...)` call with an unconditional throw; the composed test still exited
  0 and printed its unchanged bootstrap/snapshot/socket PASS line. The production UI also has no
  gate-crossing control, no `wind_down` control, and no control that leaves Run End for the already
  active next run. `game_ui_snapshot.v2` exposes no authoritative transition-action projection,
  and the ratified Copy package has no CTA rows for those three operations. The archived first-hour
  script necessarily contains `cross_gate` and `wind_down`; therefore the missing behavior is a
  contract/product gap, not merely a missing assertion.
- **AC2 — MET on the current tree.** `make verify-client` exited 0: TypeScript/Svelte reported zero
  diagnostics, 6,655 unit tests passed, the shell/Game-UI import boundary passed, generated API and
  Copy mirrors matched their governed hashes, and the kernel/history guards passed at `0.3.100`.
- **AC3/AC4 — MET on the current tree.** `make test-browser` ran Chromium, Firefox, and WebKit:
  20,007 assertions passed (3 skipped). This includes the compile-time no-snapshot Run End fixture,
  byte-only terminal rendering, cap explanation, drain and resync story beats, renderer
  substitution, and the C11 axe gate on every Phase-A surface/system state.
- **AC5 — MET on the current tree.** The dedicated performance arm in that same Make flow passed
  at the ruled 20 Hz input / 10 Hz formatting contract (1 test passed, 10 deliberately skipped by
  the performance-only filter).
- **Aggregate disclosure:** `make verify` was run, not assumed. Server packages, generators, drift
  checks, and the first harness invariant completed green, but the aggregate was interrupted after
  23 minutes of silent `balance-harness -mode=check` work. It is not recorded as green. That command
  sequentially recomputes every registered relevance suite without a progress marker; the Game-UI
  closeout should gain an explicit root Make lane rather than requiring an opaque unrelated balance
  recomputation for every UI iteration.

### GU-C25 — AC1 still names a stale/ambiguous script rather than the archived authority

The RFC says “full T2 script,” while the now-archived content authority defines a precise
two-run first-hour script with seven milestones. A browser driver must not invent a different play
policy or silently replace UI actions with direct API calls.

**Proposed contract:** bind Game-UI AC1 to `casual.t0_t1` v1, seed `0`, from the owner-ratified
`scenario.t0_t1_first_hour` and policy artifact. The browser proof consumes the exact command list
produced by the shared `RunExperimentScript` authority. Wait edges advance only the fixture clock;
every non-wait command originates from an enabled, visible DOM control and reaches the server only
through `runtime.ts`. The proof asserts all seven archived milestones, both run-end surfaces, the
run-2 Garage crossing, and the later elective Exit. Direct browser-test `fetch` calls to
`/api/v1/intents` are forbidden. Reconcile “T2” to this exact archived authority rather than
leaving two legal interpretations.

### GU-C26 — the wire has no authoritative gate/Exit action projection

The Desk can buy only generators/upgrades and perform the manual action. Inferring transition
eligibility from progress, balances, or tier would violate U2's no-client-authority rule; always
showing an enabled button would turn ordinary gameplay rejections into the current generic offline
state.

**Proposed contract:** `game_ui_snapshot.v3` adds exactly one `transitions` object:
`{cross_gate:null|{eligible:boolean,gate_id:string,route_id:null},wind_down:{eligible:boolean}}`.
The server computes both advisory booleans through one pure production-kernel preview over the
pinned bundle and replay-owned state; it does not duplicate requirement math, mutate state, emit an
event, or create replay input. Phase A permits only the standard `gate.t0_to_t1` row and
`route_id:null`; any route choice or later gate remains successor work and fails closed. The two
buttons submit the existing `cross_gate` and `wind_down` intents with current Company revision;
`wind_down` also carries the observed Founder revision. Receipts remain authoritative. Stored v1
and v2 bootstrap receipts remain legal and render no transition controls; v3 is the current mint.

### GU-C27 — terminal navigation and its Copy bytes do not exist

`RunEndSurface` correctly accepts only the decoded `run_ended` payload, but nothing lets the player
enter the already-created next run. Adding a snapshot prop to that component would regress AC3.

**Proposed contract:** the parent screen owns a copy-backed continuation control adjacent to the
byte-only Run End component. Activation performs only `runtime.snapshot()`; it must receive
`run_seq == ended.run_seq + 1`, then clear terminal/offer UI state and select Desk. A mismatch or
request failure stays on Run End and renders the existing offline state; no new intent or local
run authority is created. Add three owner-authored Copy rows through the governed pipeline:
`desk.cross_gate`, `desk.wind_down`, and `screen.run_end.continue`. Proposed launch text, pending
owner adoption: “Move Into the Garage”, “Wind Down Company”, and “Start the Next Company”.

### GU-C28 — the real-browser topology cannot run governed time without a fixture seam

The archived script spans attended minutes and offline gaps; sleeping wall-clock time in CI is not
an acceptable implementation, while a production clock-control endpoint would be a security bug.

**Proposed contract:** add a fixture-only composed gameserver executable/harness that uses the real
`Compose` graph and real Postgres with its injected mutable clock. A localhost-only fixture control
channel may expose the next shared-authority command and advance time, but cannot submit an intent
or mutate a save. Chromium must locate and activate the matching visible UI control, then assert
the authoritative revision and visible milestone before advancing. Add root targets
`test-game-ui-first-hour` and `verify-game-ui`; the latter composes `verify-client`,
`test-browser`, `test-game-ui-composed`, and the new first-hour lane. Hosted CI runs that explicit
lane. Mutation requirements: disconnect any one transition button or suppress either terminal
surface and the first-hour test must fail; disabling all `act(...)` calls must fail before the
first manual milestone.

**Handoff boundary:** owner rulings are required on GU-C25–GU-C28 before implementation. AC2–AC5
are evidence-backed; AC1 and both remaining plan boxes stay open. No status change, archival move,
deployment, push, or completion claim is authorized.

## 2026-08-16 — OWNER RULINGS on GU-C25–GU-C28

Ruled by Marco (recorded by Claude from explicit selections and follow-up direction).

1. **GU-C26 ACCEPTED as proposed.** `game_ui_snapshot.v3` gains exactly one `transitions` object
   with server-computed advisory booleans for `cross_gate` and `wind_down`, derived from a single
   pure production-kernel preview over the pinned bundle and replay-owned state — no duplicated
   requirement math, no mutation, no event, no replay input. Buttons submit the existing intents;
   receipts stay authoritative. Stored v1/v2 receipts remain legal and render no controls. The
   always-enabled alternative is REJECTED: it would route ordinary gameplay rejections through the
   same generic error surface as real faults.
2. **GU-C27 ACCEPTED, and the three Copy rows are ADOPTED VERBATIM** as owner-authored content:
   `desk.cross_gate` = **"Move Into the Garage"**, `desk.wind_down` = **"Wind Down Company"**,
   `screen.run_end.continue` = **"Start the Next Company"**. Implementers may not edit this text.
   The continuation control is owned by the parent screen (never a prop on the byte-only Run End
   component, which would regress AC3), performs only `runtime.snapshot()`, requires
   `run_seq == ended.run_seq + 1`, and on mismatch or failure stays on Run End rendering the
   existing offline state.
3. **GU-C28 REJECTED as proposed — no fixture clock seam, no fixture-only binary, no localhost
   control channel.** Owner's direction, verbatim in substance: *do a normal test; do not prove
   the game in CI.* The first hour is already proven twice — 97 headless runs and a composed
   real-Postgres test asserting all seven milestones at a pinned seed. Replaying two virtual hours
   through Chromium to prove it a third time is verification cost compounding, which the
   2026-08-13 ruling exists to stop. **Ruled instead:** the browser proves the UI's own job — that
   each of the three controls renders only when the server says eligible, submits the correct
   intent through `runtime.ts`, and handles its receipt; and that Run End reaches the next run.
   Preconditions may be established server-side as ordinary test setup rather than by playing
   them out. No new `test-game-ui-first-hour` lane and no fixture gameserver executable.
   `verify-game-ui` composing the EXISTING lanes is still welcome, and the disclosed
   23-minute opaque `balance-harness -mode=check` dependency should stop gating UI iteration.
4. **The discrimination requirement SURVIVES the narrowing and is mandatory.** AC1 failed because
   a composed test exited 0 with every post-bootstrap action replaced by a throw. Therefore:
   disconnecting any one transition control, or suppressing either terminal surface, MUST fail the
   test; and disabling the driver's actions MUST fail before the first asserted milestone. A cheap
   test that cannot fail is not what was ruled here — a cheap test that *can* is.
5. **GU-C25 ACCEPTED, narrowed by (3) — consequence stated explicitly for owner correction.** AC1
   binds to the archived authority in the sense that the browser may not invent its own play
   policy and may not bypass the UI with direct `fetch` to `/api/v1/intents`: every non-wait
   command originates from an enabled, visible DOM control and reaches the server through
   `runtime.ts`. It does **not** require the browser to replay all seven milestones across two
   virtual hours — that claim is carried by the archived headless and composed proofs. If the
   owner intended the full seven-milestone browser replay, this clause is the one to correct.

**Codex is unblocked on GU-C25–C28 as ruled.** AC2–AC5 stand as evidence-backed; AC1 and both
remaining plan boxes stay open until the narrowed proof lands and passes its designated review.

## 2026-08-16 — Codex pre-implementation body-reconciliation blocker

- **Recorded by:** Codex. This is a process blocker, not a new product question.
- The GU-C25/C28 ruling narrows the browser's obligation to the three transition/continuation
  controls and explicitly rejects replaying the full seven-milestone/two-hour script in Chromium.
  The normative RFC Acceptance Criterion 1 still says: “the full T2 script (T0–T1 RFC) completed
  through the real UI against the composed gameserver.” These are mutually incompatible legal
  acceptance bars.
- Per the binding same-edit body-reconciliation rule, the ruling author must reconcile that AC1
  sentence. The implementer may not choose one arm silently or edit the owner's normative text.
  Implementation resumes immediately after the owner-authored body fix; GU-C26/C27 mechanics and
  adopted Copy need no further ruling.

## 2026-08-22 — GU-C7/GU-C25–GU-C28 body reconciliation

- Owner direction resumed the highest-leverage preview blocker. Reconciled the live status and
  T0–T1 dependency, U1's milestone name, U2's authoritative two-era mapping, AC1's browser
  obligation, the plan, and canonical docs to the recorded rulings.
- AC1 now requires the UI's actual job: server-eligible Gate/Wind Down controls, both terminal
  states, and parent-owned next-run continuation through `runtime.ts`, with server-side setup
  allowed and direct-intent/fixture-clock/two-hour browser replay explicitly forbidden.
- Queue row 4 is READY for the accepted snapshot-v3/control/oracle implementation. No product code,
  schema, generated artifact, balance, owner-authored copy, status promotion, review, archive, push,
  or publication changed.

## 2026-08-22 — GU-C25–GU-C28 implementation predeclaration

- Added `transition-witness-manifest.md` before product changes. It fixes the production-preview,
  v3/legacy wire, Desk-control, parent-owned continuation, terminal, and three existing browser-story
  populations and names the required severing controls.
- Scope permits the accepted projection/schema/Copy/client/test/doc/Make work and forbids balance,
  formula, CI-workflow, and archived-script changes. The cross-party review and archive gates remain.

## 2026-08-22 — GU-C25–GU-C28 implementation and witness completion

- **Implementation:** `fa5fc42` adds `game_ui_snapshot.v3.transitions`, retaining exact v1/v2
  bootstrap receipt schemas. The Game UI projection invokes the existing production transition on
  a decoded state copy; `server/production` remains byte-identical to the pre-batch tree. Phase A
  exposes only `gate.t0_to_t1` with null route, uses the existing Tier-1 Wind Down rule, and fails
  later transitions closed.
- **Pin record:** GU-C26 explicitly authorizes the v3 widening. Per the current-plus-legacy rule,
  `make api-pin` re-baselined the unversioned current schema and preserved named V1/V2 shapes in the
  same implementation commit. This was deliberate, generated, documented, and not a silent pin edit.
- **Client:** governed adopted Copy bytes drive Gate, Wind Down, and parent-owned continuation.
  Stored v1/v2 receipts render no transition controls. Continuation calls only `runtime.snapshot()`
  and advances only to `ended.run_seq+1`. The implementation also corrected two blockers exposed by
  the real witness: UI intents had minted server-rejected UUIDv4 IDs, and eager post-intent snapshot
  fetching advanced the cursor before `run_ended`, suppressing the terminal surface. UUIDv7 is now
  pinned byte-exactly and ordered event→receipt refresh owns terminal navigation.
- **Cold positive evidence:** focused `./production ./gameui ./account ./gameserver -count=1`, API
  regeneration, Copy/manifest checks, TypeScript/Svelte, 6,660 client tests, 20,037 functional
  browser assertions plus the isolated performance lane, the complete `Integration ./... -count=1`
  real-Postgres population, and the real Chromium/Vite/gameserver/Postgres/WebSocket composed
  witness all passed. The composed witness performs Gate, first Wind
  Down, scripted Run End, run-2 continuation, second Gate/Wind Down, standard Run End, and run-3
  continuation using visible enabled controls through `runtime.ts`.
- **Restored severing evidence:** bypassing the projected gate comparison failed the below-threshold
  preview; raising the projected Wind Down floor failed Tier 1; changing UUID version 7 to 4 failed the
  exact runtime test; disconnecting Gate, Wind Down, or continuation failed their browser request/
  snapshot oracle; suppressing scripted or standard Run End failed its distinct title oracle; and
  removing cap, drain, or resync rendering independently failed its exact-copy browser case. Every
  mutation was restored and the worktree returned byte-clean before this record.
- The first remaining plan box flips in this same designated-review range because its exercising
  tests landed in `fa5fc42`. AC5's 4×/dropped-frame manual release observation remains honestly open
  as separate platform queue row 4a and is not claimed by this implementation.
- **Review handoff:** implementation batch `05acc65^..fa5fc42` plus this record commit is ready for
  Claude's mandatory cross-party exact-range review. The implementer does not approve or archive.

## 2026-08-22 — kernel-version guard correction before review

- The new aggregate `make verify-game-ui` lane was run before designated handoff. Its first run
  stopped at `verify-kernel-version`: the initial implementation had placed the read-only preview
  and a behavior-identical Wind Down predicate extraction under kernel-watched `server/production`
  without a version bump. No browser/composed result from that interrupted aggregate is claimed.
- A `kernel/VERSION` bump was rejected as a false transition-semantics signal, and the guard was not
  weakened. The observer moved to `server/gameui`, where it decodes a Company clone and invokes the
  existing exported production transition. The original production package is byte-identical to
  `05acc65`.
- The affected commits were unpushed and referenced by no review verdict. Under AGENTS.md's narrow
  false-version-signal rewrite carve-out, the fixup was autosquashed before review: original
  `f2a197b` became `fa5fc42`, and record commit `ff04863` became this rewritten record. All live
  planning references were reconciled to the new implementation hash.
