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

## First-pass distribution

| State family | Rows | Meaning now |
|---|---:|---|
| Draft / not a completion claim | 39 | Eight draft RFCs have no implementation claim. |
| Mechanically backed or cold-witness green, proof/range replay pending | 52 | Code/tests appear to exist; 13 rows have now run cold, but none is promoted to proven without discrimination and range review. |
| Unmet or partial | 13 | API/Combat/consumer gaps plus newly separated Account/WebSocket/Game UI criterion fragments. |
| Contradicted at current HEAD | 3 | CI AC1/AC3 and Minigame Platform AC6 are false now. |
| Withdrawn / refuted | 4 | Dispatch Integrity is not implementable work. |

The 52 mechanically backed/cold-witness rows are the Wave-3 proof population. The ledger names their
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

Permits and First Content Epoch look implemented in later epoch 6–8 code and tests while their live
plans retain mint/review/activation boxes as open. Minigame Platform and epoch-7 planning show the
same successor-landed pattern. Those rows cannot be closed by flipping boxes from code inspection:
the exact exercising tests and designated review ranges must be shown to cover the implementation
span, then the plan/RFC/docs/archive move must be one reviewed closeout.

## Next execution batches

1. Account AC1–AC7 against the real Postgres lane, including fixture discrimination and transport
   subscribe revocation.
2. WebSocket AC1–AC6 cold, including actual 5k population, dropped subscriber, overflow, drain,
   and authz negatives.
3. Leaderboards/Prestige/FCE/Permits range-union and current-bundle acceptance replay.
4. API, Game UI, and Minigame consumer traces; keep incomplete rows downgraded.
5. Risk-ranked archived-RFC replay, starting with release, account data, epoch identity, replay,
   and harness instruments rather than sampling only cheap unit suites.
