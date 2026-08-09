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
