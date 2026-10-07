# Permits & T3 gate implementation log

## 2026-08-07 — candidate assembly by Codex

The candidate economy/routes documents were assembled from the active Phase-0 bytes using the
owner-ruled insertion order. The RFC's human-form amounts were normalized to RFC-0001 canonical
Decimal bytes (`24` → `2.4e1`, `12` → `1.2e1`) after the real loader rejected the noncanonical
spellings; values and mechanics are unchanged.

Evidence shipped in the same range:

- shared Go and TypeScript candidate loading;
- route-resource and doctrine-route composition;
- neutral and non-neutral all-target multiplier behavior;
- 24-hour offline accrual and near-cap saturation;
- generated copy outputs and orphan record for the inactive cap-reason binding.

`make copy-check` and `make test` passed. This is a self-review record only; it does not satisfy
the designated independent review or authorize activation, archival, or the epoch-6 mint.

## 2026-08-07 — designated cross-party verdict: candidate implementation {7d9cb37} — NOT APPROVED (narrow)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Reviewed range:** `7d9cb37^..7d9cb37` (the RFC's full implementation span to date; b14f82e is
  RFC-text-tier, not consumed).

**The candidate bytes are excellent — all verified clean:** all four manifest SHAs recomputed and
matching; P1/P2/P3 rows exact with the ruled insertion order (pure additive hunks, surrounding
bytes byte-identical to active); `2.4e1`/`1.2e1` confirmed pure RFC-0001 canonical normalization;
the PT-C4 owner copy literal byte-for-byte on the bound key; the PT-C1 no-active-artifact law HELD
(subject correctly not BALANCE-CHANGE:); orphan-stage copy landing consistent with FCE-C8 stage 1;
kernel honestly unchanged (guards green).

**BLOCKING:**
- **F1 (HIGH — repo gate red from this commit forward):** `verify-schema` (a `make verify`
  component) semantically validates every `balance/routes-testdata/valid/` fixture against the
  ACTIVE company-resource universe — the routes candidate's `company.permits` requirement fails
  (`/gates/1/requirement/1 references unknown company resource`). Green at 7d9cb37^, red at
  7d9cb37; every tracked commit until the mint fails the repository's own gate. The log claimed
  only copy-check + `make test` (which excludes verify-schema) — undisclosed by gap, not intent.
  Fix: pair candidate routes with their candidate economy in the sweep, or move the candidate out
  of the semantically-swept `valid/` directory (content unchanged → manifest HASHES survive; the
  manifest PATH cells get a reconciling edit).
- **F2 (MEDIUM, probe-proven):** the "near-cap saturation" case is non-discriminating — it lands
  exactly at the cap without needing the clamp (clamp neutered → case still passes; only the
  offline case catches it). One-line fix: start at 2.39995e1 or run 2 s so the unclamped result
  overshoots.
- **F3 (MEDIUM):** the PT-C3 "shared parity fixture" is not shared — the four behavior cases are
  Go-only; the TS client has NO hardcap clamping in its prediction paths and permits will be the
  first reachable cap. Cross-runtime saturation parity must be proven (or the changelog's "shared
  … hardcap proofs" claim retracted and the TS gap filed as its own finding).
- **F4 (MEDIUM):** the PT-C5 activation fixtures (cross-epoch Exit init, fresh-genesis, old-run
  replay) are ABSENT and untracked in plan.md; and this is the repo's FIRST two-requirement gate —
  multi-requirement crossing debit/rejection/replay is untested anywhere. Even if deferred to the
  mint range by design, plan.md must carry the unchecked items; AC1/AC3 were owner-amended to
  include them.

**Verdict: NOT APPROVED pending F1 (urgent — the repo gate) + F2–F4; re-review will be narrow.
Owner ratification of the manifest hashes is HELD until the fix lands (hashes need not change if
files move rather than change content).**

## 2026-08-07 — narrow remediation by Codex — ready for designated re-review

- F1: the unchanged routes candidate moved outside the active-resource semantic sweep in
  `90633a6`; its SHA stayed unchanged and `verify-schema` is green again.
- F2/F3: `testdata/permits-hardcap-v1.json` now drives both the Go engine and the TypeScript
  prediction machine from `2.39995e1` at `1e-3/s` for one second. The unclamped value is
  `2.40005e1`, so the expected `2.4e1` is a discriminating hardcap assertion in both runtimes.
- F4: a candidate-bundle test proves rejection leaves both balances untouched and success debits
  cash and permits atomically. The genesis, cross-epoch activation, old-run replay, and shared
  replay-row proofs cannot honestly run before the owner-gated epoch-6 artifact exists; each is
  now an explicit unchecked mint item in `plan.md`.

This is a remediation handoff, not an approval or archival claim. It is ready for the designated
cross-party re-review after the normal repository gates pass.

## 2026-08-08 — designated cross-party re-review: remediation {90633a6, d30ab9e} — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.

All four findings CLOSED with probes:
- **F1:** pure R100 move (similarity 100%, SHA recomputed unchanged at old-commit/new-commit/HEAD)
  out of every semantic sweep; verify-schema green at 90633a6, full `make verify` green (exit 0,
  synchronous) at d30ab9e; zero stale path references; RFC manifest path cell reconciled with the
  hash cell untouched.
- **F2:** the new shared vector overshoots (2.39995e1 + 1e-3/s × 1s = 2.40005e1 > cap 2.4e1,
  representable at canonical 12 digits both runtimes); probe: neutering the Go clamp fails the
  test at exactly the near-cap assertion.
- **F3:** genuinely shared fixture — TS drives PredictionMachine through 20 real ticks and fails
  on a neutered clamp (probe-proven). CORRECTION recorded: the TS clamp predates 7d9cb37
  (client-shell foundation) — the original "TS has NO clamping" was overstated; the true gap was
  the missing cross-runtime proof, which d30ab9e supplies. No client/src change → no kernel bump
  owed (verified honest).
- **F4:** the first-ever two-requirement gate crossing is now tested atomically (rejection with
  cash satisfied + permits short leaves BOTH balances untouched; success debits both to zero and
  crosses Tier 3→4); the four activation proofs are explicit unchecked plan items tied to the
  epoch-6 landing (the verdict's tracked-deferral branch — sound for law-inactive candidates).

**N1 (ADVISORY):** the manifest's "composed generated copy" hash 8462d6d5… is a point-in-time
composition attestation at 7d9cb37 — interleaved MA copy landings moved the live composed catalog
(now 2f299e2a…, permits row present and byte-exact). Ratify the economy/routes/copy-source hashes
as byte-pins; treat the generated hash as attestation-only, refreshed at mint time.

**Verdict: APPROVED — combined Permits set {7d9cb37, 90633a6, d30ab9e}; range union covers the
full Permits implementation span. The candidate manifest is RATIFICATION-READY:** economy
31af760c…990 · routes 6c7c4350…df2 (at balance/testdata/permits-t3-gate-candidate-v1.json) · copy
e87b0224…512 as byte-pins; generated 8462d6d5…8a2 as the 7d9cb37 attestation per N1. After owner
ratification, the bytes await only the epoch-6 mint (PT-C1: activated atomically by the First
Content Epoch).

## 2026-10-07 — RP-028: shared two-resource gate replay supplement

- Scope: test-only supplement after `a42e680e`; no production, balance, copy, schema or CI bytes
  change. `permits_replay_test.go` executes the live gate handler, restores its input and executes
  `ApplyLogged`; the small `testdata/replay/permits-gate-v1.json` corpus is also consumed by
  `client/test/permits-replay.test.ts`. Complete receipt/event/state bytes must agree. Independent
  checks cover cash short, permits short, both short (cash-first reason), exact payment and surplus
  preservation; refused state remains byte-identical with no events.
- Fixture scope: ratified economy/routes candidates + matching first-content category bytes over
  the historical base catalog, Company v14. Not the complete epoch-6/8 bundle, PostgreSQL commit,
  duplicate-intent delivery or mounted player journey. Activation/history acceptance remains open.
- Checks: cold `make test-go vet GO_PACKAGES='./production ./routes ./economy'
  GO_TEST_FLAGS='-count=1 -v -run Permi'` passed the four production tests and vet; its selector
  selected no routes/economy tests. Those packages passed in the preceding unfiltered run, but
  **that overall run failed** on the existing 27 `TestAxisAccrualPartitionProperty` divergences
  (RP-307); no assertion was weakened. Three affected client files passed (95 tests); `make
  typecheck` returned zero errors/warnings. No aggregate or hosted CI green claimed.
- Initial invalid runs are excluded: Make swallowed the regex end anchor (zero tests), and a
  custom Go test flag before the package caused setup failure. Valid explicit generation is
  `make test-go GO_PACKAGES='./production -args -update-replay-fixture'
  GO_TEST_FLAGS='-count=1 -v -run TestPermitsGateReplayCorpus'`. The first TS run rejected the
  mismatched old category gate list; using the existing matching category artifact fixed the
  instrument, without bypassing its loader or changing gate semantics. Normal tests only compare.
- Review by: Codex (implementer diff check); Recorded by: Codex. Designated cross-party review
  pending for this test/record batch; no independent approval or archival claim.
- Next: designated review of this supplement; author reconciliation RP-027, canonical-doc repair
  RP-029 and the remaining activation/history evidence before Permits archival.
