# Active-RFC acceptance audit

Coordinate: product tree `190a4fa`; 2026-08-20. State: **first classification pass complete;
criterion-by-criterion execution and range review in progress.**

`active-acceptance-ledger.tsv` contains all 111 true acceptance criteria across the 21 product or
process RFCs in the active directory. The extraction stops at any nested heading. That boundary is
load-bearing: a naive `##`-only parser misclassified Combat Duel's four D4 hardening rulings as a
second AC1–AC4 set and reported 115 rows.

## Classification rules

- **Mechanically backed** means a plausible implementation and named test/gate exist. It is not a
  proof verdict until the fixture/oracle is inspected, the gate runs cold at this coordinate, and
  the implementation/review ranges cover the behavior.
- **Unmet per live plan** means the owning plan itself leaves the criterion's required producer or
  consumer unfinished.
- **Contradicted** means current executable or repository evidence falsifies the criterion.
- **Draft / not a completion claim** preserves future acceptance criteria without pretending their
  unimplemented RFC is defective merely for being draft.
- **Withdrawn** retains the refuted Dispatch Integrity proposal as history, not active work.

## Distribution after completed cold/lifecycle batches

| State family | Rows | Meaning now |
|---|---:|---|
| Draft / not a completion claim | 39 | Eight draft RFCs have no implementation claim. |
| Mechanically backed or cold-witness green, proof/range replay pending | 43 | Code/tests appear to exist, but the remaining rows have not completed discrimination and range review. |
| Proven or historically proven with a named qualification | 7 | Permits/FCE rows whose witnesses and designated implementation ranges were reconciled; qualifications remain explicit. |
| Unmet or partial | 15 | API/Combat/consumer gaps plus Account/WebSocket/Game UI and two newly confirmed Permits/FCE witness/provenance gaps. |
| Contradicted at current HEAD | 3 | CI AC1/AC3 and Minigame Platform AC6 are false now. |
| Withdrawn / refuted | 4 | Dispatch Integrity is not implementable work. |

The 43 mechanically backed/cold-witness rows are the remaining Wave-3 proof population. The ledger names their
specific witness/range route rather than treating them as a green block.

## Confirmed criterion-level downgrades

### CI Baseline

- **AC1 false:** `make verify` does not pass in current hosted CI because the harness job is
  cancelled. The complete local aggregate was also not allowed to reach a verdict in the initial
  audit; the focused current local harness check subsequently passed.
- **AC3 false twice:** the blocking workflow does not finish under five minutes, and it does not
  finish under its later 30-minute harness ceiling on hosted public runners. R-001 carries the
  exact run evidence.
- AC2 and AC4 have mechanical gates, but their seeded-failure discrimination still requires a
  current replay. AC5's permission, trigger, concurrency, and declared cache shape mechanically
  match D2/D3; behavioral and history review remains.

### Minigame Platform

- **AC6 false:** no combat duel tenant exists or registers. Only The Pitch is present in production
  catalog and gameserver composition. The later C9 ruling agrees with runtime reality while the
  header, MP1, MP5, C3/C8/C12/C14, and AC6 retain the opposite requirement. RP-016 blocks authorial
  reconciliation and archival.

### Game UI

- **AC1 unmet:** the browser proof reaches the Desk/bootstrap boundary, not the full ruled first
  hour through run end and run 2. The AC body also contradicts the later GU-C25–C28 ruling over the
  clock seam. Author reconciliation precedes R-004.
- AC2–AC5 have code/tests to inspect, but are not yet promoted. Component tests, screenshots, axe,
  or a performance fixture do not collectively prove the complete human workflow.
- The first cold browser pass now leaves AC2, AC3, and AC5 mechanically green pending mutation and
  review-range audit. AC4 is downgraded: the cap primitive has a real browser assertion, but the
  Game UI fixture sets drain and resync simultaneously and checks only axe/mechanical-ID absence.
  It does not assert either story beat's content or recovery transition (RP-026).

### API Foundation

- AC3 and AC5 remain unfinished with the public DTO/readers/router/privacy integration in the live
  plan. AC4 is partial: generated types exist, but deletion of every superseded hand-written client
  layer has not been traced.

### Minigame API & Surface

- AC5 remains explicitly open. The backend lifecycle, schema, limiter, and privacy rows have
  mechanical evidence; no production `minigame_session` surface mounts The Pitch as required.

### Combat Shared Data

- AC1, AC4, AC5, and AC6 remain open with the full catalog, roster properties, Obedience/Soul
  tables, and combat division lint. Arithmetic and PRNG rows exist but await cold dual-runtime
  execution and discrimination review.

## Lifecycle rows that need transactional reconciliation

The Permits/First Content lifecycle pass is now recorded in
`permits-first-content-lifecycle-audit.md`.

- Permits AC1/AC2 are proven integration. AC3 is partial: atomic two-resource rejection/debit is
  discriminating, but the promised Go/TypeScript replay row for that exact crossing is absent.
  AC4's implemented review/activation chronology follows PT-C1, while the normative criterion
  still says the ruled-impossible opposite. Canonical Economy/Routes docs remain pre-mint.
- First Content AC1/AC3 are proven. AC2 was proven at the mint but its exact epoch-6 test later
  drifted to deploy-current epoch 8. AC4 is partial because the changelog does not resolve every
  consumed verdict to an exact reviewed range. AC5 is proven under the accepted range-head
  interpretation, but FCE5.5/AC5 still retain the contradicted mint-commit wording.

The historical implementation ranges and designated verdicts are real; they do not erase these
closeout defects. No plan box or archive state was changed by audit inference. Minigame Platform
and epoch-7 planning still require the same successor/range reconciliation.

## Next execution batches

1. Account AC1–AC7 against the real Postgres lane, including fixture discrimination and transport
   subscribe revocation.
2. WebSocket AC1–AC6 cold, including actual 5k population, dropped subscriber, overflow, drain,
   and authz negatives.
3. Leaderboards/Prestige/FCE/Permits range-union and current-bundle acceptance replay.
4. API, Game UI, and Minigame consumer traces; keep incomplete rows downgraded.
5. Risk-ranked archived-RFC replay, starting with release, account data, epoch identity, replay,
   and harness instruments rather than sampling only cheap unit suites.
