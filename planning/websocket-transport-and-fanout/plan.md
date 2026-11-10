# WebSocket Transport & Fan-out — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/websocket-transport-and-fanout.md`
- **Started:** 2026-07-29

**2026-11-11 RP-785:** the live restart notice can omit its zero stream offset.
The corrected browser waits before reconnecting and preserves its saved
position. The exact `5a71dbe0..12efc365` source correction is independently
approved. The refreshed `2f366a90` bundle independently passed the real
Caddy/Postgres restart check with its packaged site. [Transport evidence](log.md)

2026-11-05 RP-769 corrective batch: PlayerRelay's Flush serializer now honors
waiting callers' cancellation/deadlines before claiming more work. Real locked-
outbox/server regressions fail exact old source at both shorter2s and configured15s
bounds; final corrected race witnesses pass with the lock retained through worker/
backend cancellation. Root native affected race/vet/build pass. Broader SQL run
retains RP-770's first-hour verification failure, also reproduced against the
original relay; no complete SQL-green or cause claim. Final focused relay tests
pass after correcting an unsupported instantaneous-backend-cleanup assertion.
Complete seven-path `3aeb811c..baddca78` receives fresh independent corrective
APPROVAL after exact-old/current native and real-PG race execution plus affected
outbox/recovery/lifecycle controls; the owning log records provenance. No drain limit, receipt/order/lease, kernel226,
epoch8, gameplay or acceptance-policy change. RP-770 diagnosis remains separate.

2026-11-04 RP-768 is independently APPROVED under Account D2 / Transport T1/T5.
The actual periodic recheck previously held Centrifuge's presence mutex/timer
past token expiry. Native and real Linux-amd64/Postgres regressions fail old
7505c5bf and pass the correction. The verified expiry now bounds that read as
well as node cancellation;15-minute TTL/cadence,4001 expiry and4003 drain remain.
Both witnesses forbid early expiry; real SQL ceases waiting with its lock held,
revocation is unchanged and a fresh session connects afterward. Full affected
native and Linux-amd64/Postgres race suites and vet pass, with the existing
opt-in cleanup skip explicit. Complete seven-path7505c5bf..c32f6902 receives fresh
independent corrective APPROVAL after exact-old/current expiry, shutdown/drain and
actual family-revocation execution. Original feature/release gates remain separate.

2026-11-05 RP-767 is independently APPROVED with kernel0.3.226. Independent
technical review /root/membership_version_authority_review establishes that
accepted KV-1's normal same-commit bump applies to this behavior-changing
correction. The prior root-authored neutrality/exception hold was overbroad;
no draft adoption, registry exception/removal or new owner ruling is needed.
Guild/Cohort permission lookups now receive client and node cancellation through
the production resolver chain, including the actual SQL read. On the current
base all4 old races fail; ordinary tracked-test correction passes all4 with
locks still held, under the original2s caller and configured15s bounds.

All8 affected native/Linux-amd64 real-PG race packages, client replay/transport,
strict types, native gameserver build and vet pass. Root/reviewer initially overlap
shared-DB suites and retain their failures; unchanged-source serial reviewer and
full root runs pass. Complete25-path215a98b3..71d8e15c receives independent
corrective APPROVAL; exact verdict and provenance are in the owning log. Full feature,
historical kernel-guard debt and hosted/clean-host/release gates stay separate.
The ordinary regression is
`make test-go-ci CI_TEST_PACKAGES=./gameserver CI_TEST_FLAGS='-race -run TestComposedDrainCancelsBlockedMembershipLookupIntegration -v'`.
The prior six-pathc7a9e385..d446a17f candidate approval stays historical. Its
preparer is now pinned to215a98b3 and deliberately refuses changed source;
use only the matching source checkout for historical overlay reproduction.

2026-11-04 RP-766: pending initial authentication now inherits node cancellation;
late results are refused and late library registrations close before queue setup.
Exact-old unit and actual locked-Postgres/socket controls fail; corrected cases,
the prior periodic witness, revocation and full affected native/Linux SQL race
populations pass. The deliberate library-handoff model is separate from actual
Account proof. Complete seven-path082de7a7..8833d82d is independently APPROVED
after fresh exact-old/current unit/model and actual SQL/socket checks; no
token/timer policy, whole-feature or release acceptance follows.

2026-11-04 RP-765: periodic reauthentication's uncancelable query blocks client
closure during shutdown. The actual callback regression and real locked-Postgres/
socket witness fail against old source and pass with node-owned cancellation.
Full affected Transport/Gameserver/Account native race and Linux-amd64/Postgres
race populations plus vet pass; the log records the existing opt-in skip and
exact evidence. Complete seven-path1992ce15..034ea3cd receives independent
corrective APPROVAL after fresh old/current SQL/socket and revocation checks. This does
not establish RP-450's original hosted cause or full Transport acceptance.

2026-10-15 current-run cohort correction (RP-651): subscribe-time permission now requires the
projected membership to match the current non-archived Company head's run and member flag.
The old real HTTP/Postgres/socket witness admits a new non-member subscription after Exit;
the expanded existing Exit witness also checks legitimate pre-Exit access, stable assignment,
re-sign and leave. The complete corrective range `54f352f6..aa2f97df` is independently
APPROVED under Marco's corrective-review delegation; the owning log records the verdict.
No governance, Health/retention or established-subscription eviction
policy changes; complete Transport acceptance/archival remain open.

RP-651's history follow-up now executes nonempty-history recovery with membership removed and
restored in all three gated namespaces, plus direct-history refusal. This is a test-only proof
of existing behavior with controlled resolvers, not another product fix or live DB/tenant proof.
The history supplement is included in that approval, with independent full package/vet,
actual Postgres and repeated race checks. Do not repeat these completed checks for reassurance.

2026-10-15 composition reconciliation: the old Commons-resolver/binary blocker is resolved.
`gameserver.New` wires the real Commons projector and Guild settlement service; the current
built-client aggregate at `6201bef8` executes the persisted socket/clearing/readiness/drain
test twice successfully. Item4 records implementation presence, not complete Transport review
or archival. The owning log retains the exact witness and remaining review obligations.

Current known-payload conformance correction (2026-10-13, RP-561): Go validates exact
DTO-owned fields/nullability and safe wire integers; the client requires actual string
enums instead of coercing arrays. The shared174-case corpus, connected publisher/history
refusal, healthy delivery and affected checks pass. The owning log records both old-source
failures and independent corrective APPROVAL of `6a5a742a^..6a5a742a`. Opaque owner payloads,
wire version and recovery policy are unchanged; no whole-feature or release promotion.

Current bounded correction (2026-10-13, RP-560): policy loading validates exact original
JSON members against the existing DTO tags before struct decoding. Case aliases and raw/
escaped duplicate members now refuse; canonical escaping and the declared limits/origins
are unchanged. All48 old-source failures and corrected cold/race/consumer checks are in
the owning log. Independent corrective review APPROVES `b8eb4c58^..b8eb4c58`; no recovery/authentication
policy, balance/kernel, whole Transport acceptance or archival is changed.

1. [x] Add strict Phase-0 transport policy and closed Go/TypeScript wire envelopes.
2. [x] Embed Centrifuge with the implemented channel authorization, history, and backpressure policy.
3. [x] Map committed production receipts/events into one transaction-owned private-player outbox;
   each scope is revision-ordered and snapshot recovery uses the authoritative full-state endpoint.
4. [x] Compose `cmd/gameserver`, readiness, HTTP intents, and bounded drain lifecycle. The real
   composer wires `production.WithCommonsWeightResolver(commonsProjector)` and
   `production.WithGuildSettlements(guildService)`, not the empty account constructor fixture.
   `TestComposedGameserverPostgresSocketClearingAndGCIntegration` passes on actual Postgres
   at `6201bef8`, including three real clearing settlements consumed by an HTTP intent,
   persisted boundary3, member authorization, readiness and bounded drain. Full review remains.
5. [x] Add recovery, overflow, authz, drain, and 5k-connection soak coverage.
6. [x] Update canonical docs and run full verification.
7. [ ] Record independent review before archival.

## Q-003 — production recovery completion (predeclared 2026-08-21)

Authority is the accepted D4/T4 recovery contract and the READY Q-003 manifest row. This batch
binds the existing `PlayerRevisionCursor` and Centrifuge stream positions into the production Game
UI transport consumer. One controller owns initial subscribe, reconnect, resubscribe, recovered
publication delivery, authoritative `/api/v1/founder/state` full sync, typed closes, and the D4
drain delay.

Required discriminating cases:

- initial subscription with no saved position, followed by persisted per-channel epoch/offset;
- same-revision event-ID duplicate suppression, forward revision-gap full sync, and historical
  compensation delivery without cursor movement;
- successful history recovery and expired/unavailable history falling back to full sync before a
  clean live resubscribe;
- queue-overflow (4000), server-drain (4003 plus `resume_after_ms`), and credential-expiry (4001)
  closes entering the same recovery path, with drain suppressing reconnect for its advertised
  delay;
- a literal ten-second connected world stall proving bounded backlog and convergence to the newest
  state, either on the live connection or through reconnect/recovery; a fully blocked socket may
  close abnormally because its close frame is itself undeliverable; and
- a real non-member Guild subscription denied beside a member control.

Cold gates are the root client unit, typecheck, boundary, browser, and composed disconnect/recovery
paths plus cold server Transport, Account, and Gameserver populations. Every claimed oracle must
also be mutation-probed or carry an explicit in-range negative fixture.

Out of scope: Account access-token rotation semantics; a second/raw API client; Game UI snapshot-v3
mechanics; player-facing copy; RFC archival; and unrelated Transport/server behavior. A 4001 close
may use the current persisted access token and fail honestly, but this batch does not invent token
refresh authority.

**Current Q-003 state:** D4/T4 consumer recovery and the AC6 non-member Guild negative landed at
`c63e7e6` with cold and mutation evidence. The owner reconciled AC3 on 2026-08-21 to the standard
player outcome: a ten-second stall has bounded backlog and converges to newest authoritative state,
with disconnect/recovery permitted. `afb5bf0` adds the literal 10.50-second actual-socket witness;
refusing every post-baseline world revision makes it fail after the full stall. All Q-003 cold gates
are restored. The implementation/witness span ends at `afb5bf0`; designated review begins after
`bfd9b65` and must include the final record-only handoff/correction tip named in the relay.

Q-003 received designated cross-party approval at `249719c` for exact range
`bfd9b65..65b2506`. The bounded batch is closed; item 7 and RP-055's separate body/history/archive
union remain open.
