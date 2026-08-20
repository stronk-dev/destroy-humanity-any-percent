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
| Mechanically backed or cold-witness green, proof/range replay pending | 16 | Code/tests appear to exist, but the remaining rows have not completed discrimination and range review. |
| Proven or historically proven with a named qualification | 14 | Reconciled lifecycle rows plus CI AC2/AC4, Transport AC1, and Account AC1/AC4's current integration/security proofs; qualifications remain explicit. |
| Unmet or partial | 32 | API/Combat/consumer gaps plus Account/Transport/Game UI and literal-witness gaps. |
| Contradicted or failed at current HEAD | 6 | Exact current contradictions plus failed Leaderboards AC5, Minigame Platform AC2, and CI AC5. |
| Withdrawn / refuted | 4 | Dispatch Integrity is not implementable work. |

The 16 mechanically backed/cold-witness rows are the remaining Wave-3 proof population. The ledger names their
specific witness/range route rather than treating them as a green block.

## Confirmed criterion-level downgrades

### CI Baseline

- **AC1 false:** `make verify` does not pass in current hosted CI because the harness job is
  cancelled. The complete local aggregate was also not allowed to reach a verdict in the initial
  audit; the focused current local harness check subsequently passed.
- **AC3 false twice:** the blocking workflow does not finish under five minutes, and it does not
  finish under its later 30-minute harness ceiling on hosted public runners. R-001 carries the
  exact run evidence.
- **AC2 proven:** current hosted Node and three-engine browser jobs pass; a restored shared-vector
  corruption makes the actual client gate fail exactly at the corrupted row.
- **AC4 proven:** the verifier explicitly consumes all 20 schema files, passes current production,
  and a restored malformed production catalog makes the command exit nonzero.
- **AC5 false:** permissions and cancellation match, but harness/numeric cache Go build outputs
  despite dependency-only claims; schedule/manual runs all blocking jobs, and numeric failure can
  fail the supposedly non-blocking nightly workflow (RP-057).
- RFC/plan topology is stale (RP-056), review ranges are fragmented (RP-058), and R-001's next
  harness instrumentation has no active accepted implementation owner (RP-059).

### Minigame Platform

- **AC2 false:** ranked-power/breadth loaders work, but quality decay is called only when a new
  result immediately replaces it; start reads stale stored grades, and no automated-output
  consumer uses the destination. The active/idle bridge cannot occur (RP-044).
- **AC1 partial:** solo Pitch is proven through the composed API; production hardcodes solo and no
  async-snapshot tenant lifecycle exists (RP-045).
- **AC3 partial:** cap/forfeit and reduction arithmetic are proven; an actual bot match is absent
  because no bot tenant/content exists (RP-045).
- **AC4 cold-green pending final range reconciliation; AC5 proven integration.**
- **AC6 false:** no combat duel tenant exists or registers. Only The Pitch is present in production
  catalog and gameserver composition. The draft duel child and combat docs confirm the absence.
  RP-016 and RP-046 block authorial reconciliation and archival.

### Account & Session Bootstrap

The criterion and lifecycle pass is recorded in `account-session-lifecycle-audit.md`.

- AC1 and AC4 are proven: strict real-stack create→session→intent plus a correctly signed
  extra-claim JWT rejection execute cold.
- AC2 remains partial because database family revocation and socket authentication are not one
  revoked-live-session witness.
- AC3/AC5/AC6/AC7 are partial against “unlimited,” actual import→board exclusion, all save-stream
  survival, and all unauthenticated endpoints respectively (RP-049).
- Outside the backend criteria, the production UI never consumes its refresh token, so the
  15-minute access expiry has no renewal/reconnect path (RP-048). Existing RP-004–RP-007 retain the
  absent export/delete/recovery/local-fallback user workflows.
- The active review record cites obsolete hashes and does not union current successor work
  (RP-050); anonymous account storage retention remains unowned (RP-051).

### WebSocket Transport & Fan-out

The criterion and lifecycle pass is recorded in `websocket-transport-lifecycle-audit.md`.

- AC1 is proven: 5,000 actual in-memory WebSockets each expose a terminal trace across ten world
  ticks, with monotonic subsequences and wrong-kind/public-receipt negative cases in the oracle.
- AC2 is partial. The server's real Centrifuge recovery returns missed private receipts and the
  newest world snapshot, but the production Game UI never stores epoch/offset, requests recovery,
  or reconnects (RP-052).
- AC3 is partial because the live stall is under one second and permits one in-flight frame plus
  the newest rather than the literal 10 seconds/exactly one (RP-054).
- AC4/AC5 are partial: typed overflow and bounded server drain are real, while the browser discards
  close codes, ignores `resume_after_ms`, and has no reconnect/full-sync path (RP-052).
- AC6 is partial. Denied Match and positive member cases exist, but no non-member Guild fixture can
  falsify an authorize-every-Guild resolver (RP-054).
- The per-scope revision cursor is unit-tested but unused by production, so duplicate/gap/
  historical-compensation behavior is an orphaned primitive rather than a shipped safety property
  (RP-053). Plan/body/review history also require current-range reconciliation (RP-055).

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

### Prestige & Exits

The criterion and range pass is recorded in `prestige-lifecycle-audit.md`.

- AC1 and AC7 have current cold witnesses but no archival-qualifying full-range designated review.
- AC2–AC6 are partial against their literal text: the current suite lacks an offer-age property,
  a full reseed/non-empty repeated-ledger witness, an integrated New-Founder lifecycle, an
  eligible-state/mid-event-chain matrix, and the required checked-in run-2 golden.
- AC8 is proven under the owner-approved first-hour policy: 97/97 current runs complete with no
  failures and the successor implementation range is designated-reviewed. Its 45-minute lower
  edge is a policy precondition, so this is availability evidence rather than human-choice timing.
- RP-034–RP-036 capture stale body/docs, the absent Advisor control, and the deferred Quarter
  bridge; RP-038 captures the missing tracked post-rewrite full-range review authority.

## Next execution batches

1. Account AC1–AC7 against the real Postgres lane, including fixture discrimination and transport
   subscribe revocation.
2. WebSocket AC1–AC6 cold, including actual 5k population, dropped subscriber, overflow, drain,
   and authz negatives.
3. Leaderboards range-union, compaction, and current-bundle acceptance replay.
4. API, Game UI, and Minigame consumer traces; keep incomplete rows downgraded.
5. Risk-ranked archived-RFC replay, starting with release, account data, epoch identity, replay,
   and harness instruments rather than sampling only cheap unit suites.
