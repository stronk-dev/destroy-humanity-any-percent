# Active-RFC acceptance audit

Coordinate: product tree `190a4fa`; 2026-08-20 through 2026-08-21. State: **all 111 active-RFC
criteria have bounded evidence verdicts; five remain deliberately open on exact review/provenance
closeout. The separate 46-row archived-RFC inventory and 20-row/ten-domain deep replay are
complete.**

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
| Mechanically backed or cold-witness green, proof/range replay pending | 5 | Remaining rows are held for exact archival review/provenance, not unexecuted lifecycle audits. |
| Proven or historically proven with a named qualification | 20 | Reconciled lifecycle rows plus Combat AC2–AC4, Minigame API AC1, Game UI AC2/AC3, CI AC2/AC4, Transport AC1, and Account AC1/AC4; qualifications remain explicit. |
| Unmet or partial | 33 | API/Combat/consumer gaps plus Account/Transport/Game UI/Minigame API and literal-witness gaps. |
| Contradicted or failed at current HEAD | 10 | Exact current contradictions plus Combat AC6, Minigame API AC4, Game UI AC4, API AC4, failed Leaderboards AC5, Minigame Platform AC2, and CI AC5. |
| Withdrawn / refuted | 4 | Dispatch Integrity is not implementable work. |

The five mechanically backed/cold-witness rows are the remaining exact-review population. Every
active RFC lifecycle has now been re-walked; these rows retain their named range route rather than
being promoted by arithmetic.

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
- AC2/AC4/AC5 are implemented pending designated review: the production controller persists
  positions, recovers missed receipts, full-syncs gaps/overflow, interprets typed closes, and honors
  drain delay; the composed browser witness crosses a real disconnect and committed intent.
- AC3's exact-one-frame premise fired under a literal stall and was owner-reconciled on 2026-08-21
  to bounded backlog plus recovery after any disconnect and newest-authoritative-state convergence.
  Its ten-second ruled witness is the remaining implementation proof (RP-054).
- AC6's real second-account/non-member Guild denial now fails under an authorize-every-Guild
  mutation. Match's participant positive remains owned by its eventual implementation.
- The per-scope cursor is bound to production live/recovered publications and authoritative snapshot
  resets, pending the same exact Q-003 designated review (RP-053). Plan/body/review-history
  reconciliation remains separate (RP-055).

### Game UI

- **AC1 unmet:** the browser proof reaches the Desk/bootstrap boundary, not the full ruled first
  hour through run end and run 2. The AC body also contradicts the later GU-C25–C28 ruling over the
  clock seam. Author reconciliation precedes R-004.
- AC2 is proven: a restored direct transport import fails the component boundary by filename.
- AC3 is proven: widening Run End props to accept a snapshot fails the compile-time negative, and
  the browser renders the decoded terminal object through the isolated component.
- **AC4's oracle fails:** suppressing cap, drain, and resync output together leaves the full 20,007
  browser population green. The behavior exists, but the literal three tests do not (RP-026).
- AC5 is partial: deterministic update-count/long-task mechanics run, while the declared 4× CPU
  throttle and dropped-frame allowance are unused and the manual device check is unrecorded
  (RP-066).

### API Foundation

- AC1 is partial: registered descriptor drift fails the generator gate, but 11 of 21 live v1 routes
  bypass the registry and a restored Founder-response mutation leaves `api-check` green.
- AC2 is partial: optional growth passes and removal rejects, but the failure names only the root
  schema rather than the removed nested field; unregistered live routes have no pins.
- AC3 and AC5 remain unfinished: there are zero production public paths, and policy/cursor runtime,
  readers, evidence, formula bytes, privacy enumeration, and composition are absent.
- AC4 is contradicted under C9. There is no generated HTTP client; production uses an injected raw
  fetcher, and a direct runtime `/api/v1` fetch mutation passes the boundary lint. The normative AC
  body also retains the pre-ruling “diff shows deletion” language.

### Minigame API & Surface

- AC1 is proven as a composed backend integration. The cold Postgres/HTTP lifecycle covers Pitch
  and Recovery, and severing the production tenant-content resolver fails at the exact create seam.
- AC2 is partial: generation/registration and drift checks are real, while the error schema allows
  invalid category/detail pairs and an extra-byte rejection mutation survives its substring oracle.
- AC3 is partial: command-flood nonmutation is byte-compared; Recovery flooding proves only a 429
  against a stateless adapter, not unchanged authoritative progress/session state.
- AC4's enumeration oracle fails: it covers four minigame operations, while a Recovery finish
  identity-field mutation leaves all Account unit tests green.
- AC5 remains unmet and body-blocked. No generated HTTP client, exact C9 contract, surface
  components, Pitch child mount, Recovery scheduler wiring, or browser/a11y workflow exists.

### Combat Shared Data

- AC2 and AC3 are proven in both runtimes: independent ATK-stage and substream-label mutations fail
  the shared vector corpus. AC4 is also proven; removing one winning edge fails Go and TypeScript.
- AC1 and AC5 are specification-blocked. The RFC never enumerates its closed effect/spell unions or
  literal Trust/Obedience/Soul tables, and neither loaders nor fixtures exist (RP-072).
- AC6's client half is proven by internal controls and an external nested `.mts` mutation, but the
  literal all-combat-path claim is contradicted: server Combat uses native division and is outside
  the gate (RP-074).

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
