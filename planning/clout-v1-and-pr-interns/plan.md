# Clout v1 + PR Interns implementation plan

RFC: `rfc/clout-v1-and-pr-interns.md` (accepted 2026-09-25; every owner decision at its
recommended default: OD-1 = A, OD-2 = `achievement_attainment_run`, OD-3 = one-time cash upgrades
through `buy_upgrade`, OD-4 = social mint deferred, OD-6 = slot after `milestones`, OD-9 = Clout line
withheld). Fixture-first: no production epoch is minted by this plan; the CV8 numbers are proposals
ratified later by owner SHA (OD-5).

Under A, CV6 (the Clout ledger) and CV7 (the social seam) are **not** implemented: the RFC's CV0
table marks the ledger "no" for A, and the acceptance rejects B/C. The Gaia-law test (AC4) still
lands, with the allowed writer set empty.

## Batches

- [x] P1 (`ace4e67e`, formulas `80f1bd47`) — CV1 economy schema v5 (Go `server/economy`, TS `client/src/economy-kernel.ts`):
  `axis_stack` block, the axis upgrade-effect arm, the upgrade-only `axis_at_least` requirement,
  the `axis_stack` multiplier slot (after `milestones`), provider/declaration pairing, and
  cross-artifact validation against `achievements`. Shared loader corpus. AC1.
- [x] P2 (`f256b235`, formulas `63aa0b62`) — CV3 formula: axis contributions from Company state, the clamp, `saturated`; shared
  Go-authored vectors reproduced by TS. AC2. Formulas artifact regenerated in its own commit
  (AC10).
- [x] P3 (`f256b235`) — implementation presence, not full acceptance: CV2/CV4/CV5 Company save v19 (next free) attainment set + derived score, second hook
  pass (Go + TS), `achievement_reattained.v1` event + migration, new-run activation, Exit discard,
  migration corpus. ACs 3, 5, 6, 7, 8.
  RP-307: AC6's seeded timing/partition property and retroactive-factor failing case
  are absent from the original population; formula/helper parity is not that proof.
- [x] P4 (`4c089f00`) — AC4 Gaia-law structural test (empty writer set under A; seeded writer fails).
- [x] P5 (`2f2c5a04`) — CV9 snapshot producer (optional v4 field) + Desk PR row progress: implementation presence, not full AC11/default composed journey acceptance. RP-303 locally corrects native progress associations; its Codex span independently needs Claude.
- [ ] P6 — PARTIAL (see log P6; blocked on harness attainment evaluation) — CV10 harness: scenario bundle rejection without achievements, relevance mask, dead-row
  fixture, observation + `axis_input_within_cap` invariant. AC9.
- [ ] P7 — Canon docs (AC12). Partial: `docs/axis-stack.md` and pointers landed; RFC index/manifest G09 rows and the final range are owed at completion.
- [ ] Designated Codex review, archival (Codex).

Kernel protocol: every commit touching a `kernel/affecting-paths.json` prefix bumps
`kernel/VERSION` (+ Go/TS constants) in the same commit. Save/snapshot/migration numbers are
assigned in landing order (the RFC's "v19" means next-free Company version).

## RP-303 — accessible PR progress, bounded consumer correction

Predeclared at b2b2d5a5, 2026-10-06. First reproduce the unnamed native progress
bar through a browser role/name query, without changing the production panel.
Test both unowned interns together, then a snapshot replacement with one owned:
exact name-to-row association, numeric value/maximum, description and remaining
bar must hold; no gameplay intent may be emitted. Negative controls remove the
label association and point it at the wrong row; both must fail after the fix.
Use existing copy keys only. No server/schema/arithmetic/kernel/epoch/CI changes.
Run cold Chromium/WebKit cases via root Make, type/client/build/boundary/copy
checks; report any Firefox launch failure separately, never as a passed case.
No checkboxes or full P5/AC11 promotion. New Codex implementation needs Claude.

Executed: unchanged-production failure and both compiling label controls
discriminate; restored Chromium/WebKit six cases plus the independent performance
lane pass. Types/build/client/boundary/copy/topology pass; Firefox zero executed
after session timeout, not waived. Full evidence and review range are in log.md.

## RP-304 — mandatory roles on axis upgrades

Predeclared at b30808e5, 2026-10-06. Review Claude ace4e67e^..ace4e67e's
Go/TS axis-role branches against CV1 item5. Existing branch loops check role
identity/duplicates but not cardinality. New separate shared cases must reproduce
empty roles on either/both PR rows; unchanged fixture, one declared role, and
legacy static-empty-role cases are controls. Original corpus stays unchanged.
If confirmed, reject empty roles only for axis effects, preserving static catalog
semantics. Watched loader changes require honest kernel161→162 in the same commit.
Independent compiling guard omissions in both runtimes must fail, then restore.
Run cold relevant packages/vet, full client/types/build and applicable root checks.
No schema/balance/copy/mint/role-vocabulary/pool retune or whole-P1 approval.
Role names alone do not prove executed pool binding; that question stays separate.
Every new Codex implementation/record edge requires Claude independently.

Executed: six-case paired corpus reproduces three illegal admissions per
runtime; both narrow guards fix them and independent omissions fail again.
Kernel162 accompanies the real acceptance-set change. Final cold relevant Go/
vet,8248units/types/build/formulas/boundaries/topology,80 native Chromium/WebKit
cases and real-Postgres production/gameui packages pass. Full historical kernel
guard is still red at50a3a514 (RP-131), not bypassed. Full new span afterb30808e5
including records requires Claude. Truthful pool binding remains D-022/RP-305;
no full P1/AC1, hosted CI or mint acceptance follows.

## RP-307 — actual timing and partition audit

Predeclared at07bb3d9d, 2026-10-07. Test-only first: actual ApplyLogged
generator-purchase transition on an admitted v19 fixture, with an owned PR Intern
and derivation-valid x=6. Check the interval preceding re-attainment uses x=6;
the purchase raises x to8 and only the subsequent interval uses x=8. Assert
actual ledger/debit, ordered achievement events and unchanged independent pins.
Use seeded nonzero intervals and online/offline accrual arms. Separately compare
full encoded Company states after one-shot versus seeded millisecond partitions
of the same axis-enabled interval, using actual Evaluate and pinned contributions.
Do not claim extra evaluations are extra player intents or full composed/SQL proof.

The oracle is exact, not a tolerance. Include whole-second and tiny-cut controls;
any numeric divergence is a finding, never justification to narrow AC6. Before
runtime faults, the unchanged source must execute; a retroactive-factor fault is
allowed only as a temporary compiling observation, with exact-byte restoration.
Do not change runtime/save/kernel/balance/copy/CI or owner-authored specifications
in this range. If proof requires a new persistence/numeric contract, route it to
the author/owner rather than inventing one. All new Codex test/record edges require
Claude; no checkbox flip, full P3/AC6, mint or archival approval follows.

## R-012 first wave — conserved-state feasibility, not a runtime repair

Predeclared atb6a3c48d,2026-10-07. Test-only Go/Rat and TS/BigInt models;
no production arbitrary-precision dependency or accepted save representation.
Exact domain: exponent-32..32, terminating decimal scale<=48, intervals<=60000ms.
Projection uses the existing Decimal quantizer, not a newly invented rational
half-even rule: the shipped midpoint goldens are the authority.

Population is542 primary cases per runtime:

- 512 ordinary frozen-rate cases:2 admitted axis states/rates(1.15115,2.4048),
  4 initial balances(0,1e0,1e4,1e8),2 efficiencies(1,.9),32 intervals/cuts.
  LCG uint32 seed120307, multiplier1664525/increment1013904223. Fixed policies:
  end2000/cut1000, end3114/cut1553, end60000/cut1; remaining29 draw end
  2+next%59999,cut1+next%(end-1). Record actual Go Evaluate full-state and cash
  one-shot/split results; TS compares the same scalar frozen-rate primitive, not
  a claim of served TS Company transition parity.
- 8 cap cases:2 rates*2 efficiencies*2 initial balances(at1e4 and one rounding
  unit below). Only test-local catalog cap bytes change, rehashed and validated.
- 8 rate/debit boundaries:2 efficiencies*2 first-phase cuts(1,1553 of3114ms)*
  2 second-phase cuts(1,2000 of4000ms), start1e4, debit11.3, then new rate2.4048.
  Experimental action settlement resets residue after the ordinary quantized
  debit; this is a measured branch, not the adopted persistence policy.
- 4 cap→debit→accrue boundaries:2 rates*2 efficiencies; cap overflow must not
  become deferred income after spending11.3. Two1000ms accrual phases.
- 10 projection controls:9 bounded existing canonical numeric goldens (zero,
  +/-1,20-digit integer,four signed midpoints,exponent carry), plus the same
  carry coefficient at exponent-4. Never restamp existing expected outputs.

Eight invalid-input/restore controls cover out-of-domain exponents, nonterminating
rationals, excessive decimal scale, missing residue, negative reconstructed
balance, noncanonical wire, and unsupported elapsed interval. Whole-second
controls at1e4 must match. Discarded-residue/restart, retroactive repricing and
retained-cap-overflow models must demonstrably disagree on eligible rows.

Report every actual baseline/reference/prototype difference, both wire and
residue, counts, source SHA identities and explicit excluded domains in a new
shared research artifact. Generation must refuse invalid/incomplete evidence;
committed artifact must replay exactly in Go and TS. No production/save/kernel/
balance/old-golden/CI change; no full-domain, AC6/CI green or release acceptance.
Entire new Codex test/artifact/record span needs Claude independently.

First wave executed2026-10-07 after61f01349: all542 primary cases/eight
refusals per runtime complete. Current45 partition differences; bounded conserved
models match their references, dropped carry45/retroactive8/cap4 controls fire.
Full client8799pass/340skip/types clean. Current production still fails the
existing27/128 AC6 cases; no box or acceptance promotion. Results, exact commands,
false-promotion probe and unmeasured representation/restore/settlement domains
are in partition-research.md. Further research must be predeclared; no new
persistence contract or full-domain implementation follows from this prototype.

## R-012 second wave — frozen-anchor and reconstruction comparison

Predeclared atb54fc7ef,2026-10-07, before new instrument or measurement.
Research only, using unchanged Go AccrueConstant / TS accrueConstant and
canonical numeric boundaries; no production/save/kernel/CI/old-artifact edits.
First-wave artifacts/instruments stay frozen, and AC6's red regression remains.

The experimental anchor stores canonical initial balance, source rates,
efficiency/cap, elapsed integer milliseconds and its recomputed visible balance.
Mere evaluations preserve the initial anchor; real experimental debit or rate
change settles current visible state and starts a new anchor. Each split
serializes/restores before proceeding. This is an empirical settlement arm,
not an adopted save field or authority to change player mechanics.

Population:615 primary cases per runtime, plus16 restore refusals each:

- All520 first-wave frozen rows, unchanged. Compare final anchor one/split
  snapshots to existing current one-shot cash, not an idealized rounding law.
  Rebase-on-every-evaluation negative arm must disagree on at least the45 known
  cases; record every disagreement, including any additional ones.
- 64 numeric-domain cases:8 exponents(-MAX,-1000000,-33,0,33,1000000,MAX-14,
  MAX-1),2 source sets([1.15115eE],[1.15115eE,2eE]),2 efficiencies(1,.9),2
  intervals(2000/cut1000; MAX_EXACT_INTEGER/cut floor(MAX_EXACT_INTEGER/2)).
  Initial1eE/cap1eMAX. Expected four upper-domain large-interval refusals;
  neither cap nor split may hide an invalid intermediate. Exactly60 accepted
  plus4 rejected arms per runtime, no omitted cells. This samples domain edges,
  not all exponents nor actual engine/server-clock execution at enormous times.
- All16 existing production-accrual goldens unchanged, initial0/cap1eMAX.
  Split at floor(elapsed/2), including zero/single-ms controls as applicable.
  Exactly11 accepted/five refused; valid delta must match each literal golden.
- All12 first-wave rate/debit and cap/debit rows. Record agreement/difference
  versus that earlier rational settlement branch, rather than insisting two
  distinct rounding paths are universally equivalent. Rate-change negative
  prices the prior interval using the future rate; all8 eligible controls fire.
  Cap/debit resets must not expose retained pre-cap overflow.
- Three near-cap cases:initial9.99999999999e3,cap1e4,efficiency1; rates4e-9,
  5e-9,6e-9; end2000/cut1000. Report one/split and post-debit11.3 followed by
  1000ms. Do not infer an unruled visible-cap absorption policy from the result.

Restore is a test-only canonical JSON codec. Exact allowed fields/order/version,
canonical nonnegative numeric strings, source count<=3, exact nonnegative elapsed,
recomputed visible wire and full byte re-encoding must agree. Its512-byte limit
is a conservative schema bound:seven at-most31-byte canonical strings plus fixed
field syntax/three source elements and the16-digit elapsed integer, not a
performance ceiling chosen from a truncated run. Report actual maximum size;
reject unsupported source counts rather than truncate them.

Sixteen explicit restore negatives:missing version,version2,missing initial,
missing rates,negative initial,noncanonical initial,negative efficiency,wrong
valid visible wire,wrong valid rate,unknown field,trailing JSON,duplicate wire,
noncanonical field order,unsafe elapsed,input length513,four source rates.
Use a valid end3114/rate1.15115/cash1e4 base; each must fail independently.

Three additional Go-only diagnostics exercise the unchanged first-wave carry
codec:inconsistent wire0/residue1,valid snapshot followed by another JSON value,
and below-cap exact quantity whose visible projection is the cap. Record actual
admission/refusal rather than retroactively hardening or rewriting that model.
These are prototype limitations, not discovered production save bugs.

Write a new complete source-pinned observation only after every cell and
declared control completes; old corpus bytes cannot change. Declare missing,
unexpected refusal or disagreement explicitly, never skip or add a tolerance.
Result may narrow a follow-up contract proposal, not authorize implementation,
whole Clout/CI acceptance, SQL/migration/offline/provision/default-player proof,
owner-body reconciliation, archival/mint/push/deploy or release. New full span
afterb54fc7ef needs Claude independently; earlier review obligations remain.

Second wave executed2026-10-07 aftercadd7111:615 primary cases per runtime,
606accepted/nine refused,16 restore refusals each. Rebase45 plus one near-cap
disagreement; domain60/4 and unchanged primitive goldens11/5; zero prior-boundary
cash differences. Largest of1215 valid serialized snapshots223bytes. Go-only
carry diagnostics show two admissions/one refusal; both new reconstruction
faults fire semantic negatives, source/artifacts restored exactly. Client9434
pass/340skip/types/vet clean; production retains27/128 red acceptance cases.
No checkbox or production contract promotion. anchor-research.md records scope,
negative evidence and next live-producer/SQL/offline/provision/migration seams.

## R-012 third wave — actual rate producers and jsonb fidelity

Predeclared at9050fe4d,2026-10-07, before measurement. Test-only, no production,
numeric/save/kernel/old artifact/balance/CI/body changes. A new generation-only
root Make lane may pass the SQL writer switch into the existing declared Docker
test service; it must have no automatic CI/verify dependencies or budget changes.

CPU population:64 actual admitted Company producer profiles,8 generator counts
(1,2,9,99,12345,123456789,1234567890123,MaxExactInteger) ×2 efficiencies(1,.9)
×4 elapsed milliseconds(2000,3114,2045,60000). Use the admitted x6/owned-PR1
diagnostic state, set purchased-total consistently to count, initial cash0 and a
test-local cash cap1e100(rehash/admit). Freeze all original fixture/golden bytes.

For each profile:actual assembly/Rates; actual Evaluate cash; primitive accrual
of original raw rates; the same operation after each rate is serialized to its
twelve-digit canonical string; original existing Company Encode/Restore plus
re-derived rates. The restored existing context must preserve each raw rate and
actual delta exactly. Report all raw/rounded/context/engine outcomes and raw
mantissa round-trip diagnostic strings/IEEE bits, not newly permitted save wire.
TS independently consumes diagnostics through pinned Decimal and compares raw
primitive arithmetic and canonical-rate accrual. This is scalar diagnostic
parity, not served TS producer or naturally reachable gameplay evidence.

Every64 cell must be reported; admission errors are explicit instrument failures,
not skipped counts or permission to weaken validation. Do not demand a positive
rounding-difference count merely to validate the hypothesis:zero is a valid
research result. The actual-engine/primitive/context bindings still must agree.
Go/TS source pin and complete census require exact reproduction; generation
refuses failed bindings. Save the result in a separate new corpus.

SQL population:all1215 valid serialized snapshot instances from the frozen
second-wave artifact plus all16 existing malformed/framing payloads. Run ONLY
against the declared Postgres16 service; missing DB under generation is invalid,
ordinary host skip is explicitly NOT EXECUTED. Record actual server version,
runtime platform and selected source identities. Never print credentials.

Write the1215 objects into a new transaction-local temporary jsonb table and
read every row back in declared ordinal order. No migration, live save/table
write, truncate or cleanup authority. Rollback owns the temporary table. Compare
the existing strict byte-framed decoder with a TEST-ONLY logical reader:
strict known-field/type/EOF decode, canonical re-encoding, then the existing
recomputed-state validation. Logical reader must preserve the complete original
snapshot exactly, not only visible cash. Strict-vs-SQL results are observations,
not a new adopted wire policy. Record normalized JSON and all statuses.

Each of16 negative payloads traverses an actual SQL jsonb cast under a savepoint.
Record typed SQLSTATE22P02 for syntactically invalid JSON; unrelated errors are
invalid measurement, never successful refusals. Record every post-SQL decoder
admission/refusal, distinguishing logically unchanged ordering/duplicate-equal/
whitespace normalization from corrupt fields. This comparison cannot authorize
accepting unvalidated transport bytes or assume SQL preserves duplicate history.

Both reports are separate source-pinned artifacts; no host-only CPU writer can
create SQL evidence. All populations/controls and provenance must complete before
their writer can run. Next decisions/contracts follow results; no retrospective
expected-output change, old-schema field, tolerance, acceptance/status/checkbox,
whole-CI, SQL Service/Store/replay, offline/provision/1.0, archive/mint/push/deploy
claim. Entire new range after9050fe4d needs Claude independently; prior holds
and pending reviews remain. Full nine-tier/platform1.0 goal unchanged.

Executed third wave:all64 actual producer profiles complete; raw bits change64,
early rounding changes five deltas, existing restored contexts preserve all raw
rates/deltas and actual engine cash. TS independently agrees. Actual SQL16.15/
linux-arm64 preserves all1215 snapshots logically; old strict bytes refuse all.
Sixteen SQL negatives complete:12semantic/domain refusals, one22P02, three
representation-only normalizations. Both writers reject instrument faults before
overwriting; exact source restoration. Host writer without DB fails. Recorded
environment is explicitly separate from replay equality; semantic/source/object
checks remain exact. Local AMD64 lane fails before Go (exec format), not passed.
Restored9499client pass/340skip/types/vet/topology clean;128production partitions
still27fail/101pass. No acceptance flip. Results and next questions:
`rate-and-sql-research.md`; full commands/provenance in log.
