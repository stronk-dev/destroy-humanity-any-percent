# First Content Epoch implementation log

## 2026-08-07 — achievement copy orphan stage

The FCE-C7/FCE-C9 owner-authored literals are assembled as one byte-sorted `copy.v1` source and
generated through the existing Copy Pipeline. The thirteen rows intentionally remain orphans until
the achievements candidate is staged and the epoch-6 mint atomically installs its reference rows.
No active balance artifact or epoch seed changed.

Candidate hashes awaiting owner ratification:

- `copy/catalog/achievements-candidate.json` —
  `0dd211486b3e988c0fffa5311ed95c216f5bc08b4f9fd6ef7068409c3a091cf3`
- `client/src/copy/generated/catalog.json` —
  `2f299e2a3babb8f04e02c8406a127f8d3fbab321689038fbb22614cc6dba166b`
- `copy/generated/orphans.v1.json` —
  `dd7da21767197d5cdc627a68e197d0f56e2aa847e26dd8b6be18a64f20472be2`

`make copy-check` exits 0 with 65 generated keys and 53 intentional orphan warnings.

## 2026-08-08 — literal candidate composition and cross-runtime gate

The owner-ratified meters, achievements, and pets documents were copied byte-identically out of
the ignored drafting bank into tracked candidate paths. Their SHA-256 values remain exactly the
FCE-C3 pins. Permits Routes, Minigame API, Minigames/Pitch, and every unchanged base artifact now
compose from literal files rather than decoded test-helper mutations.

The sixteen-artifact candidate loads through `replaycatalog.Load` in Go and
`loadReplayCatalogBundle` in TypeScript under the same computed label:
`sha256:1a4463bcf67440ce1ba01e6c6eb850c0614329cac63064ef07725d042c7cf21a`.
The machine-readable source/production/hash table is
`planning/first-content-epoch/promotion-manifest.candidate.v1.json`. Its status remains
`candidate_awaiting_owner_ratification`; no production artifact, epoch row, or active seed changed.

Composition exposed three previously hidden byte decisions, filed as FCE-C10–C12 in the RFC:

- Categories must add `gate.t3_to_t4` to `full_gate_set`; unchanged bytes fail both loaders.
- Economy and Fiscal tests had been appending reviewed cross-artifact rows only in memory.
- The Soul Recovery fixture's `owner_kind: fixture` debit source is intentionally rejected for an
  epoch-seeded catalog; the ruled recovery-only production artifact therefore has an empty
  `debit_sources` array.

One implementation defect was found and fixed during the parity pass: TypeScript already enforced
meter/economy resource separation, while the Go composition loader omitted its existing validator.
Go now calls `ValidateResourceSeparation` over the pinned Economy resources, a discriminating
collision test rejects the previously accepted bundle, and kernel semantics advance honestly from
0.3.85 to 0.3.86.

Focused evidence from the repository root:

- `make test-go GO_PACKAGES='./replaycatalog' GO_TEST_FLAGS='-run FirstContentCandidate -count=1 -v'`
  — three tests pass, including stale-category and meter/resource-collision rejection.
- `make test-client` — 6,607 passed, 3 skipped; the shared candidate loads and the stale category
  set rejects in TypeScript.
- `make verify` — exits 0 after the complete balance harness, Go vet/all-package tests, generated
  artifact drift checks, 0-error TypeScript/Svelte diagnostics, 6,607 client tests, kernel-history
  guard at 0.3.86, copy/schema/content-manifest gates, and 19,830 browser assertions.

The same `make verify` command was repeated after implementation commit `d0dccb4`; it exits 0 and
the committed-history guard explicitly reports `kernel version parity and history guard ok:
0.3.86`.

Pending before the manifest can become owner-approved: C10–C12 rulings/ratification, the candidate
designated review (which replaces every `pending` verdict field), and the composed harness report.

<a id="2026-08-08-b0277a1"></a>

## 2026-08-08 — designated cross-party verdict: candidate staging {3530b08, d0dccb4, a0ca14e} — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

No blocking findings. Verified: the candidate-landing law HELD (active artifacts/seed
byte-untouched; only copy-orphan additions; constants_hash unchanged, copy_hash moved per the
FCE-C8 stage); ALL 16 manifest hashes recomputed and matching; byte-fidelity traced for every
candidate — meters/achievements/pets = the FCE-C3 ratified pins; economy v3 = the ratified
Permits pin + exactly the two fiscal multiplier declarations (verbatim from reviewed helpers);
routes = the ratified pin; fiscal = reviewed baseline + the TP-C15 ruled pitch unlock (cost 3);
soul = reviewed recovery fixture with debit_sources → [] per SR-C13 (single hunk); doctrines/
minigames/pitch untouched since their reviewed commits; minigame_api = the MA-C15 reviewed shape;
copy document = ALL THIRTEEN ruled texts verbatim incl. the FCE-C9 possession literal, structural
defaults exact, orphan discipline held (zero achievement rows in the active seed). Categories:
exactly one sorted line (gate.t3_to_t4 in full_gate_set), probe-proven load-bearing in BOTH
runtimes, and FCE-C10 is an honest new-ruling request, not a slip. The Go resource-separation fix
matches pre-existing TS behavior exactly, probe-discriminated, nothing rode along; kernel
0.3.85→0.3.86 honest (3530b08 correctly bump-free). Composed candidate constants hash
sha256:1a4463bc…f21a confirmed by execution in BOTH runtimes independently. make verify + the
Postgres suite green at a0ca14e.

Observations (non-blocking): OBS-1 — FCE1's body pre-states the Categories change; the C10 ruling
must confirm it (or the body re-edits). OBS-2 — manifest lacks a provenance field for the
owner-authored lane (carried by the RFC table; consider adding at ratification). OBS-3/OBS-4 —
the economy fiscal-declaration rows and the wrapper→document extractions are recorded as
understood context for the C11 ruling.

**Range-union:** consumed {3530b08, d0dccb4, a0ca14e}; with all prior verdicts, EVERY
implementation commit from e935c22 through a0ca14e is verdict-covered. Remaining non-implementation
debt: the record-review pass ({b14f82e} + the MA handoff-discipline notes).

**Verdict: APPROVED as the candidate designated review. NOT mint authorization — FCE-C10–C12
owner rulings, the FCE5.3 composed harness report, and the FCE5.6 owner sign-off remain.**

## 2026-08-09 — FCE5.3 ratified-bundle composed pacing report

`make first-content-harness` completed successfully over the ratified 16-artifact manifest at
candidate constants hash
`sha256:1a4463bcf67440ce1ba01e6c6eb850c0614329cac63064ef07725d042c7cf21a`.
The command validated every source SHA-256, rejected pending verdicts, loaded the complete bundle
through `replaycatalog.Load`, and then executed the production pacing scenario against the
candidate Economy, Routes, and Commons bytes. It ran 300 deterministic policy/seed/mode cases.

The canonical report is
`planning/first-content-epoch/composed-harness-report.v1.json`, SHA-256
`f2b41978a33309b29d6904391ffc299dc193976543207e2348b94daf6a09006c`.

Results:

- **Invariant failures: 0.** All 300 runs completed and the candidate report contains no
  deterministic envelope, conservation, resource-bound, or other harness invariant failure.
- **Pacing warnings (10–25%): 0.**
- **Owner-visible pacing findings (>25%): 7.** All seven are in `chaos.phase0`; Casual is unchanged
  on all eight statistics. Chaos first-manual p50 is also unchanged.
- Chaos first-manual p95: 1,500,000 → 2,400,000 ms (**+60.0%**).
- Chaos first-generator-purchase p50/p95: 1,800,000 → 2,400,000 ms (**+33.3%**) and
  4,200,000 → 6,000,000 ms (**+42.9%**).
- Chaos generator-count-1 p50/p95: the same 1,800,000 → 2,400,000 ms (**+33.3%**) and
  4,200,000 → 6,000,000 ms (**+42.9%**) movement.
- Chaos T0-progress-1 p50/p95: 2,400,000 → 3,300,000 ms (**+37.5%**) and
  4,800,000 → 7,200,000 ms (**+50.0%**).

Per FCE5.3, these pacing regressions are findings for Marco's mint decision, not automatic vetoes.
No active artifact, epoch row, current-epoch pointer, or production constants hash changed. The
next step remains FCE5.6: explicit owner sign-off against this report. This entry does **not**
authorize the mint.

Verification from the repository root:

- `make test-go GO_PACKAGES='./harness ./cmd/balance-harness ./replaycatalog' GO_TEST_FLAGS='-count=1'`
  — all three complete package suites pass.
- `make first-content-harness` — exits 0; repeated generation is byte-identical at the report SHA
  above.
- `make verify` — exits 0 after the full active-epoch harness, Go vet/all-package tests, generated
  formula/API checks, 0-error TypeScript/Svelte diagnostics, production client build, 6,607 client
  tests, kernel-history guard at 0.3.86, copy/schema/content-manifest gates, and 19,830 browser
  assertions.

## 2026-08-09 — designated cross-party verdict: FCE5.3 harness report {821dfb9} — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

**Report integrity fully verified:** the complete 300-run candidate harness re-run from the
committed tree reproduces the report BYTE-IDENTICALLY; the composed hash 1a4463bc… recomputed in
an independent implementation over the 16 manifest sources — exact match; a candidate/active
mixture is structurally impossible (loader cross-validation rejects it); a sensitivity probe
(one constant changed) moved 12/16 values including Casual — proving "Casual unchanged" is a real
result, not insensitivity. All honesty checks pass: raw zero-deltas (nothing rounded away),
warnings/invariants genuinely empty, run count uncapped at 300 (200 chaos + 100 casual).

**The 7 Chaos findings are a SAMPLER ARTIFACT, proven mechanically:** the candidate catalog is
purely additive; the chaos policy samples uniformly over manual+generators, so adding the
(1e8, unaffordable-early) legal_dept dilutes per-tick manual probability 1/2 to 1/3 and wastes a
third of ticks on unaffordable buy attempts — the geometric math reproduces every committed delta
to the digit (e.g. p95 first-manual tick 5 to 8 = +60%). Not a bug, and not a human-visible
regression: no real player buy-attempts an unaffordable generator every third action.

**Finding 1 (MODERATE, interpretive — for the FCE5.6 sign-off):** the pacing simulation exercises
only economy/routes/commons; the other 13 artifacts are hash/load/cross-validated but NEVER
SIMULATED. The report therefore provides no pacing evidence for the new content's actual dynamics
(buff windows, fiscal harvests, pitch payouts, permits accrual — incl. fiscal.hoard's x2.0
all-target multiplier reaching legal_dept accrual, bounded by the visible 24 cap). FCE5.3's "runs
the FULL bundle" is satisfied in the load/validate sense only. Minor: the commit ships the
candidate-mode harness tooling itself (reviewed in this pass — strict validation, active goldens
green); a non-resolving verdict anchor slug to fix at the mint; cosmetic NUL-separator strings in
the findings output.

make verify + focused suites green at 821dfb9; kernel honest at 0.3.86; the candidate-landing
law held (zero balance/seed/deployment bytes changed).

**Verdict: APPROVED {821dfb9} (range fb4f1c5..821dfb9 exactly). FCE5.3 is SATISFIED. The FCE5.6
owner sign-off is now the only remaining mint gate.**

## 2026-08-09 — epoch-6 mint implementation handoff — READY FOR DESIGNATED REVIEW

The authorized mint landed in `3ff34bf` as one `BALANCE-CHANGE:` commit: all ratified candidate
documents were copied to their production paths, epoch 6 (`First Content`) was appended without
removing any earlier accepted hash, `current_epoch_id` and the deployment constants identity moved
to `sha256:1a4463bcf67440ce1ba01e6c6eb850c0614329cac63064ef07725d042c7cf21a`, the epoch-6 changelog
was added, and the achievement reference-pointer rows were activated with copy regeneration.

The first full Postgres integration run after the mint found two real latent defects that the
fixture-first suites could not expose:

1. Company replay-inputs v5 carried only the Founder-v16 foundation projection. Epoch 6 activates
   Founder v21, so ordinary commands and Exit could reconstruct neither the v17-v21 mutable
   Founder fields nor the pinned Founder version. The live transition therefore failed closed and
   an honest epoch-6 run could not Exit.
2. `replayverify.Repository.artifacts` duplicated the old seven-artifact cardinality instead of
   relying on `replaycatalog.Load`'s closed artifact-set authority. Every honest 16-artifact run
   was therefore classified `constants_mismatch` before replay.

`84cf570` closes both findings. Replay-inputs v6 adds a closed `founder_extensions` record for the
v17-v21 minigame, pet, Fiscal, Soul, and session-sequence state; it is required exactly for pinned
Founder floors at or above 17 and is validated against the pinned artifact bytes in Go and
TypeScript. Same-epoch Exit preserves every field while resetting `minigame_session_seq` under the
v21 law. Historical v2-v5 parsing remains in place, and the shared corpus deliberately keeps a v5
row alongside a new byte-compared epoch-6/v21 Exit row. Kernel `0.3.87` is an honest semantic bump.
The verifier's stale cardinality check is deleted; the single closed artifact authority now owns
both completeness and dependency validation.

The active epoch necessarily moved the canonical harness seed away from epoch 5. In accordance
with the repository's baseline guard, `c41b388` is a separate baseline-only `BALANCE-CHANGE:`
commit. Its pacing values reproduce the owner-read FCE5.3 candidate report exactly: the seven
disclosed Chaos movements, zero Casual movement, and no invariant failure.

Discriminating evidence in the implementation range:

- the shared Go/TypeScript `first_content_exit` fixture exercises a legal same-epoch Founder-v21
  Exit, preserves ratings/quality, pets, Fiscal and Soul state, resets the session sequence, and
  compares receipt/state bytes;
- the composed gameserver/Postgres test creates an epoch-6 Founder and Company, pins genesis and
  all frozen contributions, applies a command, Exits, loads all 16 database artifacts through
  `replaycatalog`, and waits for replay verification to project the terminal run to the board;
- epoch-5 fixtures remain exact and pre-epoch-6 histories continue to restore and replay.

Verification from the repository root:

- `make verify` — exits 0: Go vet/all-package tests, active harness, generated formula/API drift,
  TypeScript/Svelte diagnostics (0 errors, 0 warnings), production client build, 6,608 client
  tests, kernel-history guard at 0.3.87, copy/schema/content-manifest gates, and 19,833 browser
  assertions.
- `make test-save-integration` — exits 0 against the real Docker/Postgres stack, including
  gameserver, production, replay verification, leaderboard, save, and migration paths.
- `git diff --check` — clean before each implementation commit.

**Code-and-baseline range for the designated mint review:** `3ff34bf^..c41b388`; the handoff range
also includes the later record commit containing this entry. This entry is an implementation
record, not a review verdict. The mint is implemented and self-checks are green;
Meters, Achievements, and Pet Care remain unarchived until the designated cross-party review
consumes the full range and approves it. Nothing was pushed.

## 2026-08-10 — designated cross-party verdict: THE EPOCH-6 MINT {3ff34bf, 84cf570, c41b388, 08c995e} — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

**The core claim holds without deviation:** all 16 production artifacts byte-identical to their
owner-ratified pins (every SHA recomputed; source==production verified independently of the
in-repo test; the base four byte-untouched vs epoch 5). Registry exact (9 rows appended, epoch-6
entry literal, current_epoch_id 5→6, ALL epoch 1-5 hashes preserved verbatim); changelog cites
resolving APPROVED verdicts (the anchor slug fixed); deployment identity moved to 1a4463bc… with
the ratified copy attestation confirmed by recomputation. FCE-C8 stage 3 atomic; copy-check green
at every commit in the range.

**84cf570 (replay-inputs v5→v6) verified a GENUINE mint-time necessity:** the Postgres suite is
RED at the mint commit alone (reproduced — the 16-artifact bundle breaks active-seed tests), and
the remediation is fail-closed in both runtimes (founder_extensions required iff floor ≥ 17,
v<6 decodable and valid only for floors ≤ 16). **The whole pre-existing corpus is byte-identical
modulo the version field** (33+3 cases diffed old-vs-new; main cases stay pinned to epoch-5).
Baseline c41b388 correct per the baseline-guard two-commit protocol, deltas equal the owner-read
report to the digit, harness-check reproduces. Kernel map clean (one honest bump at 84cf570);
subject discipline clean; AC1-AC5 green at range head; no archival jumped the gate.

**F1 (HIGH, registered follow-up obligation — not blocking):** `applyFounderReplayOutput`'s v17+
extension-persistence block (server/production/prestige.go:419) has NO discriminating test —
probe-proven: full unit + Postgres suites green with it disabled. Correct by inspection, but a
silent live-state/replay divergence seam. REQUIRED: a case where a founder with non-default
extension state Exits and the PERSISTED founder is asserted (e.g. seq 9→0, fiscal credit
preserved) — lands with the archival batch or the next production-package change.
**F2:** AC5's "green at the mint commit" is structurally unmeetable under the baseline-guard
protocol (Postgres red at 3ff34bf alone, harness-check red until c41b388) — green at range head;
RFC reconciliation note added same day (Claude-side). F3/F4 minor citation-anchor cosmetics;
F5 the mint's mechanical path reconciliation inside sign-off text (correct direction, noted);
F6 epoch-5-pinned integration tests leave active-epoch e2e coverage on the composed test.

**Range-union:** every implementation commit through 08c995e is designated-verdict-covered.
**The Meters, Achievements, and Pet Care archival moves citing this verdict are AUTHORIZED.**
THE PUSH remains separate and Marco-only.

## 2026-08-10 — mint-review F1 discrimination closure

The required persisted-Founder seam test is added to the composed epoch-6 Postgres lifecycle. It
starts Founder v21 with `minigame_session_seq=9`, `fiscal_credit=2`, and `soul=80`, performs a real
Company command and Exit, then reloads the Founder revision and requires sequence reset to 0 while
Fiscal credit and Soul remain 2 and 80. A mutation probe disabling the v17+ copy block at
`server/production/prestige.go` makes that lifecycle fail at Exit with `internal_invariant`; the
unaltered code passes the full Postgres suite. This closes the verdict's named F1 obligation without
changing runtime semantics or the kernel version.

## 2026-08-10 — designated cross-party verdict: archival batch {7b48a9d, 5c20ee3} — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
Meters/Achievements/Pet Care moves verified intact (renames 97-99%, every hunk read); all
citations resolve (the mint verdict + each foundation's full chain — pet care's nine ranges each
located); docs claims spot-checked to the digit against minted bytes; AC3 + acquisition carried
consistently in seven places. **The mint verdict's F1 obligation is DISCHARGED, probe-proven:**
disabling the prestige.go:419 extension-persistence block now fails the composed Exit test at
exactly the predicted seam. The guard-tooling edge commit 7b48a9d (archive-path citation
fallback) is minimal, fail-closed, adversarial-fixture-covered. A-F1 (LOW): Pet Care's body
reconciliation beyond status/path is the rulings-reconcile convention, correctly labeled.

## 2026-08-10 — record note (from the UI Foundation review's F2): the mint-range gate claim was cache-masked

Cold-cache `go test -count=1` fails ./economy, ./epochseed, ./fiscal from the mint commit 3ff34bf
forward: server/fiscal/catalog_test.go appends fiscal.generator.beige_tower/fiscal.hoard, which
the minted economy artifact now carries → duplicate-id rejection. Warm caches masked this through
the mint review and subsequent verdicts. THE MINT BYTES ARE UNAFFECTED (the defect is a stale
test fixture colliding with live content); the fix is a test-file change (non-colliding IDs) +
this record. Gate claims in the affected verdicts remain honest as-run; the masking is a
test-caching hazard now on record — reviewers should prefer -count=1 for gate claims.

## 2026-08-10 — cold-cache test-fixture remediation

The cache-masked failure is closed without changing production bytes or runtime behavior. The
Fiscal fixture now consumes the multiplier declarations already present in the minted economy
catalog instead of appending duplicates; the economy contract expects the complete four-source
production set; and the epoch-seed contract expects all 16 epoch-6 artifacts. The focused cold
run `make test-go GO_PACKAGES='./economy ./epochseed ./fiscal' GO_TEST_FLAGS='-count=1'` passes.
The full uncached Go gate is required before handoff.
