# WebSocket Transport & Fan-out — append-only log

## 2026-07-29 — start

- Accepted through `planning/codex-batch-2026-07-29.md` after T1–T6 closed identity, inbound intent,
  wire, recovery, literal-limit, and runnable-lifecycle gaps.
- Verified the current official Go module releases before pinning: embedded Centrifuge v0.38.0 and
  coder/websocket v1.8.15. Centrifuge remains pre-v1, so its minor version is treated as an API
  boundary and pinned exactly.
- Implementation begins with strict data/config and recovery behavior before the HTTP/WebSocket
  composition, keeping protocol semantics testable without sockets.

## 2026-07-29 — policy, wire, recovery core

- Added the exact T5 Phase-0 literals as strict declarative data, validated in Go and JSON Schema.
  Origins must be distinct HTTP(S) authorities without path/query/fragment; every bound is positive
  and constrained to the accepted operational range.
- Added the v1 Go encoder and strict TypeScript decoder. Unknown top-level kinds are ignored before
  payload interpretation; known envelopes and transport-owned payloads reject unknown fields.
  Production receipts remain the exact C1 object rather than acquiring a transport copy schema.
- Implemented drop-stale world queues and bounded lossless private queues. Overflow is an explicit
  error mapped forward to close code 4000; no receipt-dropping method exists.
- Implemented channel-offset history with player count/TTL recovery and latest-only world behavior,
  plus closed server-side player/guild/cohort/match authorization. Tests demonstrate a truncated
  player history is unrecoverable, 300 queued world snapshots resume as exactly the newest one, and
  a non-participant cannot subscribe.
- Strict TS/Svelte checks, 6,426 client tests, schema verification, and the focused Go transport
  suite are green.

## 2026-07-29 — embedded node and actual socket path

- Embedded Centrifuge v0.38.0 behind its actual WebSocket handler and promoted both Centrifuge and
  coder/websocket from indirect pins to direct dependencies. The handler enforces the declarative
  origin allowlist and 64 KiB inbound cap.
- Connect authentication delegates to the Account repository's access-token authority. Centrifuge
  receives the account ID for connection limiting and only the Founder ID as presence-visible
  connection info. Subscription callbacks apply the closed channel authorization rules and client
  publish callbacks always reject.
- The node preserves the accepted typed lifecycle: access expiry/revocation closes with 4001, a
  fourth connection replaces the oldest with 4002, and drain publishes the courtesy system message
  before closing with 4003. Periodic token revalidation detects a revoked token before its JWT TTL.
- Public world publishing is now structurally coalesced: callers only replace a pending greatest
  revision and the node-owned ticker emits at the configured 4 Hz. Private, feed/social, world, and
  match publications attach their respective Centrifuge history contracts.
- Added actual coder/websocket tests covering allowed/denied origins, token authentication, private
  channel authorization, rejection of client publishing, and twenty world updates collapsing into
  one latest publication. A self-review also found and fixed a byte-cap bypass when replacing an
  already queued world snapshot.
- Remaining acceptance work is explicit: receipt/event mapping from committed intents, composed
  gameserver plus in-flight drain gate, mapping Centrifuge's internal slow-consumer close to typed
  code 4000, actual recovery/drain integration, and the 5k-connection soak.

## 2026-07-29 — crash-safe receipt relay

- Rejected the tempting post-HTTP callback design: a crash after commit but before callback would
  permanently lose the receipt even though the RFC names the player channel as authority.
- Added migration 00015 and an intent-transaction outbox. Applied and rejected Company intents,
  including the two-stream Exit transaction, insert the exact normalized Production receipt with
  Founder identity, authoritative Company revision, constants hash, and database timestamp in the
  same transaction as the intent record and state transition. Idempotent replay does not duplicate
  the unique `(company_stream_id,intent_id)` row.
- Added ordered lease claims using `FOR UPDATE SKIP LOCKED`, explicit release on publish failure,
  claim-token-checked acknowledgement, and a transport relay that maps the stored row to the v1
  player receipt envelope without modifying the payload. Delivery is at least once; duplicate
  delivery after publish-before-ack is intentionally absorbed by existing intent/revision
  idempotency rather than weakened into at-most-once loss.
- Real Postgres tests prove ordinary and Exit intent atomicity, replay deduplication, claim expiry
  ownership, acknowledgement, and rollback at the existing injected Exit fault boundaries. Relay
  unit tests prove exact payload mapping and immediate failure release.

## 2026-07-29 — composed in-process drain lifecycle

- Added the composed server boundary around the account HTTP router, embedded WebSocket handler,
  Postgres readiness, and receipt relay. Health means process-up; readiness additionally requires a
  successful database ping, a healthy relay, and non-draining state.
- Intent admission is an explicit gate around only `POST /api/v1/intents`. Beginning drain closes
  admission and returns typed HTTP 503 responses for new work while retaining an exact count of
  already-admitted transactions. No `sync.WaitGroup.Add` races with `Wait`; the gate's zero channel
  is changed under the same mutex as admission.
- Drain order is now executable: mark not-ready/close admission -> broadcast courtesy -> await
  admitted intents -> flush the transactional receipt outbox to empty -> stop relay -> close sockets
  with 4003 -> shut down Centrifuge, all under the catalog's 15-second bound. Unit coverage blocks an
  intent mid-handler and proves sockets remain open until it commits and its relay flush completes.
- `cmd/gameserver` remains a DESIGN-GAP rather than a fake binary: production composition needs a
  concrete Founder-to-server/activity-bracket resolver and participation-weight resolver for the
  already-shipped Commons projector. The active Commons Onboarding RFC explicitly assigns those
  owners to the still-undrafted faction/incorporation and guild contracts. The generic lifecycle is
  complete, but hard-coding a deployment-wide cohort owner here would improvise that blocked model.
- The pinned Centrifuge API closes its internal byte queue with library code 3008 and exposes no
  per-node override for that disconnect. T5 requires application code 4000. Mutating Centrifuge's
  exported package-global `DisconnectSlow` would make multiple nodes/tests race and is rejected.
  This exact adapter/library mismatch remains an implementation blocker to resolve by an accepted
  dependency patch/fork decision; the separate bounded queue continues to prove the required loss
  semantics without falsely claiming the actual socket emits 4000.

## 2026-07-29 — complete-diff self-review correction

- Found a bounded-drain defect in the lifecycle diff: if an already-admitted intent outlived the
  15-second budget, the timeout branch returned before closing sockets. The ordinary ordering test
  passed because its intent completed, so the missing exceptional cleanup needed a dedicated stalled
  transaction case.
- Every failure/timeout branch now cancels the relay, closes sockets with the typed drain code, and
  invokes Centrifuge shutdown before returning the joined cause. The node also initiates Centrifuge
  shutdown when its caller context is already expired instead of returning before the library sees
  shutdown. A blocked-intent regression proves `broadcast -> close -> shutdown` still occurs at the
  deadline and the caller receives `context.DeadlineExceeded`.

## 2026-07-30 — actual recovery and sandbox-safe network tests

- Added an actual Centrifuge JSON-protocol recovery test over coder/websocket. It records the live
  stream epoch/offset, disconnects after revision 1, publishes private revisions 2 and 3 while the
  client is absent, and proves both return in order with their consecutive offsets on recovery.
- The same test disconnects a world subscriber, submits revisions 2–20 inside one coalescing
  interval, and proves recovery returns exactly one cached publication containing revision 20.
  This exercises the embedded node's real history/recovery options rather than the separate
  in-memory history model, closing Transport AC2 and AC3 at the socket boundary.
- Replaced OS-bound `httptest.NewServer` use with a shared in-memory `net.Pipe` HTTP server. The
  actual HTTP upgrade, WebSocket framing, origin checks, and account API client/server exchange all
  remain exercised, but routine Go tests no longer need permission to bind localhost. Transport and
  Account focused suites now run inside the repository sandbox with the local Go build cache.

## 2026-07-30 — closed outbound wire parity

- Complete-diff review found that the Go encoder recognized kinds but did not bind them to channel
  families or validate transport-owned payload shapes. A generic caller could therefore publish a
  receipt to `world`, violating the no-public-per-click law even though intended call sites did not.
- Go publication and TypeScript decoding now share the same channel-kind matrix: receipts are
  player-private; world carries snapshot/presence/system; feed carries curated
  event/presence/system; social/match channels reject receipts. Channel IDs cannot be empty or
  contain a second separator.
- Both runtimes exact-key validate owned payloads, require snapshot/event revisions to equal the
  envelope revision, bind snapshot scope to the channel family, reject array/scalar state payloads,
  and enforce the closed system-code duration shape. Production C1 still owns receipt internals;
  transport asserts only that its pass-through bytes encode an object.
- Added `testdata/transport/wire-vectors.json` with valid and invalid boundary envelopes and made
  both suites consume it. Focused Go tests, eight TypeScript transport tests, strict TypeScript, and
  Svelte diagnostics are green.

## 2026-07-30 — 5k fan-out and typed slow-consumer close

- Added the literal 5,000-connection acceptance soak against one embedded Centrifuge node over the
  in-memory WebSocket transport. Each connection authenticates as a distinct account, subscribes to
  `world`, and sniffs ten 10-Hz publications. The test inspects 50,000 envelopes for exact
  world/snapshot/revision order and attempts a click-shaped public receipt on every tick; the closed
  publisher rejects every attempt. The complete Transport suite, including the soak, runs in about
  three seconds locally.
- Resolved the library close-code gap without a fork or a race. Centrifuge v0.38 exposes its
  slow-writer disconnect as package policy; `NewNode` sets it through one process-wide `sync.Once`
  before any node can run. Every node therefore uses application code 4000 and no per-node mutation
  occurs after writers exist.
- An actual stalled private WebSocket test fills the configured 64-KiB queue, waits for the embedded
  node to remove the connection, drains the one already-in-flight frame, and observes close code
  `4000 queue_overflow`. Together with the full-state account endpoint and private-history recovery,
  this closes the lossless-overflow/re-sync boundary in AC4.
- **DESIGN-GAP (blocks event relay):** Exit transactions commit both Company-scope events
  (`run_ended`, `run_started`) and Founder-scope events (`founder_advanced`). T3 gives every event a
  `rev` and D3 says player-channel revisions drive shell ordering, but the event payload carries no
  scope and Company/Founder revision sequences are independent. Relaying both onto one
  `player:{founder}` stream would make ordinary revision-gap detection falsely treat valid
  cross-scope events as missing Company state. The owner must either add event scope/stream identity
  to T3 or declare which scope is relayed; no ordering rule is improvised here.
- Verification after the round is green: all Go packages and vet, generated formulas, pacing and
  epoch-history guards, package-boundary checks, strict TypeScript/Svelte, production client build,
  6,441 Node tests (3 skipped), schema validation, and 19,332 browser cases. The aggregate runner's
  output boundary stopped after the long harness command, so the remaining repository targets were
  invoked explicitly and recorded rather than inferred.

## 2026-07-30 — independent review: transport round (e2fce6c..2096916)

Full-diff review of 8caf05c (wire contracts), 44e74bc (in-memory recovery tests), 7030ab1 (soak +
overflow close), e2fce6c (sandbox-local Go cache), 2096916 (scope gap).

**Verdict: approved.** The channel-kind matrix is enforced symmetrically (Go encoder and TS decoder
share `testdata/transport/wire-vectors.json`, ≥10 vectors asserted by both suites); snapshot/event
revisions bind to the envelope revision and scope binds to channel family; the soak test's sniffers
assert an exact world snapshot revision sequence at 5k connections, so any leaked per-click message
fails structurally, and the test also proves the encoder rejects a receipt aimed at `world`. The
`centrifuge.DisconnectSlow` package-global mutation is correctly Once-guarded and documented; it is a
process-wide policy, acceptable while one Node per process is the deployment shape — if that ever
changes, revisit. Recovery tests exercise real protocol frames over the permission-free in-memory
listener; private replay-in-order and world latest-only both asserted against offsets, not sleeps
alone (the one 300 ms sleep is belt-and-braces before a latest-only assertion — fine).

**Ruling on the cross-scope event DESIGN-GAP (T3 amended in the RFC):** the event payload gains a
required `scope: "company" | "founder"` field; `rev` is the revision within that scope's stream, and
the envelope `rev` equals the payload rev (unchanged rule, now per-scope). The shell keeps one
reconciliation cursor per scope on `player:{fid}`; gap detection is per-scope, so interleaved
Company/Founder events never falsely trip it. Both Exit-transaction event families relay; nothing is
dropped. Wire vectors must gain valid/invalid scope cases.

**Ruling on the ownership resolvers blocking `cmd/gameserver`:** compose the binary now with the
real resolvers that exist — Account (sessions/founders) and Commons cohort membership (server-
assigned, already projected) — and **fail-closed deny-all resolvers for `guild:*` and `match:*`**
(subscribe rejected with the existing authz path). Guild/faction resolvers arrive with their owner
RFCs; the binary's composition does not wait for them. AC1/AC5 run against this composition.

## 2026-07-30 — independent review: transport CORE (e2eeadf..e7c1e60) — the review I owed before approving the follow-ups

Adversarial two-lens review of the core commits, findings verified against source at HEAD (later
commits reconciled). **Verdict: auth/authz/outbox-transactionality/drain-boundedness approved with
evidence; the queue-discipline layer is NOT implemented as documented — fix queue below.** This
entry also corrects my 2026-07-30 approval of 8caf05c, which repeated a doc claim I failed to check.

Verified correct (evidence in the review record): cross-founder `player:*` subscribe denial with
socket-level proof; unconditional client-publish rejection; outbox insert inside the intent-commit
transaction for applied/rejected/Exit receipts with idempotent-replay returning before any insert;
single-drainer claim tokens with `FOR UPDATE SKIP LOCKED`; drain bounded by one 15 s context with
every post-broadcast branch closing sockets (regression-tested against a stalled intent); connect
auth via the account Authenticator with fail-closed Origin allowlist; 4002 oldest-replacement.

Findings (fix queue, ordered):

1. **HIGH — D2's queue disciplines are orphaned types, not wired behavior.** `ConnectionQueue`
   (drop-stale world / lossless-bounded player, `server/transport/buffer.go`) and `History`
   (`history.go`) are referenced only by their own tests; the live path is centrifuge's single
   per-connection byte FIFO (`node.go` ClientQueueMaxSize). Consequences at HEAD: a stalled-but-
   connected consumer accumulates ~40 stale `world` snapshots over 10 s instead of one (AC3 is
   reconnect-only via history size 1, not live-queue drop-stale), and the 256-message player bound
   is unenforced (bytes only). **Correction to my 8caf05c approval:** docs/transport.md's
   "Connection queues implement the two distinct loss rules" describes the orphaned types, not the
   wiring; I approved that text without checking call sites. Fix: wire per-channel-class queue
   discipline into the publish path (centrifuge write hook or replace-on-enqueue for gauge
   channels), or amend D2/docs to the byte-FIFO model deliberately — no silent doc drift.
2. **MEDIUM — outbox ordering breaks across partial failures** (`relay.go:52-58`, `outbox.go`): a
   failed `Publish` releases only the failing claim and a failed `MarkReceiptPublished` releases
   nothing, so the remainder of a claimed batch stays invisible for the 30 s lease while newer rows
   publish first — rev 6 before rev 5 on one player channel, tripping shell gap detection into a
   full resync. Fix: release the whole unclaimed remainder on any failure, and treat per-founder
   ordering as the invariant (skip founders with a pending unacked item in the batch).
3. **MEDIUM — no poison-row path**: a row that deterministically fails Encode/Publish (e.g. receipt
   > 64 KiB; `insertReceiptOutbox` has no size guard) is re-claimed head-of-line every 25 ms
   forever; `runRelay` pins readiness false and everything behind it stalls. Fix: size guard at
   insert + bounded retries then dead-letter with InvariantSink severity.
4. **MEDIUM — `Drain`'s BroadcastDrain-error branch wedges the server half-drained**
   (`gameserver/server.go:96-98`): readiness already false and admission closed, but sockets stay
   open and no Shutdown runs. Fix: fall through to CloseForDrain/Shutdown on that branch too.
5. LOW — drain broadcast enters `player:*` history with rev 0 (recovery replays it; violates the
   rev-ties-to-revisions rule; use skip-history publish for system messages). LOW — 503 gate opens
   before the courtesy broadcast (T6 order). LOW — exported `Node.Drain` implements the wrong
   sequence and is uncalled; unexport or fix. OBSERVATIONS — rejected intents share a `rev` (shell
   must not drop equal-rev receipts as stale — noting for the shell's contract); `OnAlive`
   re-auths against Postgres per tick with no timeout (load + hang risk at scale); only
   company-scope receipts relay (disclosed; T3 scope ruling now governs the event path).

Reconciled: the core-commit gap "close code 4000 never used" was fixed at HEAD by 7030ab1
(`centrifuge.DisconnectSlow`, observed in test). The 2026-07-29 "policy, wire, recovery core" log
entry overstated ("drop-stale queues... implemented"); later entries partially corrected it, and
this entry closes the record.

## 2026-07-30 — HIGH remediation: live queue disciplines

- Replaced the orphaned in-memory `ConnectionQueue` with the application discipline actually
  attached to every live Centrifuge client. The library byte queue remains authoritative for the
  1 MiB bound; an independent reservation counter now enforces the 256-message player bound.
  Reservations are released at the transport-write hook, on publish failure, and when a player
  subscription disappears, so reconnect/resubscribe cannot inherit stale counts.
- Each world flush reserves its revision on every subscribed connection. At the actual writer hook,
  snapshot frames older than that connection's newest reserved revision are skipped; system and
  presence frames are never treated as stale snapshots. Failed publication rolls reservations back.
- Found and closed a Centrifuge seam while writing the live test: queue items retain their channel
  only when `Metrics.GetChannelNamespaceLabel` is configured. The node now supplies a bounded,
  low-cardinality classifier, making `TransportWriteEvent.Channel` a tested correctness dependency
  without exposing founder/guild/match IDs in metrics.
- Deleted the unused parallel `History` implementation. Production and tests now have one history
  authority: Centrifuge's configured stream history/recovery, already exercised over real in-memory
  WebSockets.
- Regression evidence: the stalled-world socket test repeatedly delivers at most the one already
  in-flight frame followed by revision 8, never revisions 2–7; the stalled-player test repeatedly
  reaches close code 4000 after filling 64 queued messages while total bytes remain far below the
  1 MiB limit. The focused transport suite and `go vet ./...` are green.

## 2026-07-30 — MEDIUM/LOW remediation: ordered relay, poison rows, complete drain

- Outbox claim now returns only the oldest unpublished/non-dead row per Founder and sorts SQL's
  unordered `UPDATE ... RETURNING` result by identity. Parallel workers therefore cannot claim a
  later Founder receipt while an earlier row is locked, leased, or retrying.
- Publish and acknowledgement failures share one path: increment the failed head's attempt count,
  release every untouched row from the batch immediately, and return the joined error. After five
  deterministic failures the row receives `dead_lettered_at` and the required invariant sink gets
  one `receipt_dead_letter` report with Founder/intent/outbox identity. Pending readiness excludes
  dead letters, so one poison payload cannot hold the service unavailable forever.
- Added the application and Postgres 60-KiB receipt limit beneath the 64-KiB envelope cap. The real
  Postgres test demonstrates A1/A2/B1 ordering, attempt persistence across leases, fifth-attempt
  dead-lettering, B1 progress afterward, and both size guards.
- Drain now broadcasts before closing admission, but a broadcast error still begins draining,
  stops the relay, closes sockets, and invokes shutdown under the same 15-second context. The
  duplicate exported `Node.Drain` sequence was removed. Rev-0 courtesy messages bypass history;
  a real WebSocket reconnect proves the system frame is delivered live and never replayed.
- Focused save/transport/gameserver suites and `go vet ./...` are green; full Postgres and
  repository verification follow after the remediation commit.

## 2026-07-30 — independent review: transport remediation (c87be53, fc97436)

**Verdict: HIGH 6 and MEDIUMs 10–11 are genuinely fixed on the live path — approved.** Evidence
highlights: `OnTransportWrite` intercepts inside centrifuge's per-connection writer (not a parallel
type; the orphaned `History` type was deleted); a stalled-but-connected consumer test passes under
`-race` asserting ≤ in-flight+newest with drop-stale scoped ONLY to world snapshots (receipts can
never be dropped by the hook); the 256-message player bound closes 4000 at the real socket; per-
founder ordering is enforced IN the claim SQL (`NOT EXISTS` earlier pending row — reordering after
Mark failure is structurally impossible single-process); dead-letter attempt state is a DB column
surviving restart, fifth failure dead-letters atomically with exactly one invariant report and the
founder's stream resumes; BroadcastDrain-error branch closes and shuts down; courtesy frames are
proven live-only (zero publications on recovery).

Findings (fix queue):

1. **MEDIUM — the 5k soak test now contradicts the discipline it predates**: it asserts every
   subscriber sees the exact world revision sequence, but drop-stale legitimately skips stale
   snapshots for lagging writers — reproduced failing twice under `-race` ("unexpected envelope
   rev 3"). Plain-mode passes are timing luck. Fix: assert monotonic revisions ending at the final
   rev with no non-world/non-snapshot frames, not exact sequence — that is what the discipline
   promises.
2. **LOW — readiness can flip true mid-drain** (relay tick between `ready=false` and
   `beginDrain()` re-stores true): store a drain flag checked by runRelay before re-asserting
   readiness, or beginDrain before broadcast (the T6 order says not-ready → broadcast anyway).
3. LOW — player message-count discipline decrements on channel-tagged command replies and drain
   broadcasts that never reserved (bound client-inflatable; losslessness/byte-backstop unaffected)
   — decrement only frames Publish reserved (tag reservations, or count only `pub` frames).
   LOW — drop-stale is fail-open for protobuf-protocol clients (metadata parse fails → pass);
   either restrict the endpoint to JSON protocol explicitly or parse both. LOW — the two size
   guards measure different serializations (raw bytes vs jsonb::text expansion): guard the
   canonical jsonb text length in Go too, or the CHECK can abort an intent commit the Go guard
   passed. LOW — dead-letter attempt counting is failure-kind-blind with no backoff (a 125 ms DB
   flap can dead-letter a valid receipt): count only deterministic failures (encode/size), back
   off transients.
4. OBSERVATIONS — cancelled-context release leaves rows leased (lease is the real backstop —
   acceptable, log wording corrected by this entry); multi-instance relay would reopen the
   reordering hole (zombie lease) — the relay is single-process by contract until a successor RFC;
   per-founder head-only claiming caps delivery at ~40 receipts/s/founder (fine at Phase-0 rates;
   noted for the composition's flush cadence); relay unit fixtures use batch shapes the real store
   can't produce and no end-to-end store→relay→node test exists yet — add one with an A3 row
   behind a dead-lettered A2; relay defines its own InvariantSink instead of reusing production's.

## 2026-07-30 — round-2 MEDIUM remediation: drop-stale-compatible soak

- Replaced the 5k sniffer's exact `1..10` expectation with the actual world-gauge contract: each
  connection must receive a strictly increasing revision subsequence whose terminal value is 10.
  Intermediate snapshots may be coalesced; duplicates, regression, overshoot, missing final state,
  wrong channel/kind, and every click-shaped publication still fail structurally.
- The publisher's deliberate private-receipt-on-world self-attack remains before every world tick,
  so relaxing intermediate revision cardinality does not weaken the public-traffic boundary.
- Canonical docs and AC1 now state the same oracle the live writer implements. The focused soak is
  run both normally and under `-race`; timing luck is no longer part of its pass condition.

## 2026-07-30 — round-2 transport boundary cleanup

- Readiness now has an irreversible drain flag serialized with successful relay readiness writes.
  A relay tick that began before Drain cannot raise `/readyz` after Drain marks the process
  unavailable, while the required courtesy-broadcast-before-admission-close order remains intact.
- Player queue reservations are keyed by authoritative revision and released only by actual
  publication frames carrying a matching reservation. Equal-revision rejected receipts retain
  counts; command replies and rev-0 courtesy frames cannot consume another receipt's reservation.
- The live transport-write guard decodes both JSON and protobuf Centrifuge Reply framing before
  inspecting the embedded game envelope. Malformed publication metadata now drops and disconnects
  instead of disabling drop-stale or leaking a private reservation.
- The Go outbox insertion uses PostgreSQL's exact `jsonb::text` byte count in an insert-select, the
  same representation as the database CHECK. A real-Postgres fixture is compact under 60 KiB but
  expands beyond it as jsonb text and is rejected before any row mutation.

## 2026-07-30 — relay poison/transient separation

- Closed the remaining failure-kind-blind dead-letter finding. `ErrInvalidPolicy` is the relay's
  deterministic envelope-policy failure and consumes the bounded five-attempt poison budget.
  Publisher availability failures and acknowledgement-store failures are transient: they retain
  the failed Founder head's claim for a fixed one-second backoff, persist `last_error`, and do not
  increment `attempt_count` or emit a dead-letter invariant.
- The rest of a failed batch is released immediately. The store's existing oldest-pending-row rule
  keeps later receipts for the deferred Founder blocked while allowing other Founders to progress.
- Added unit coverage for transient publish and acknowledgement paths and deterministic poison
  exhaustion. The real-Postgres integration test proves a transient deferral advances
  `claimed_until` while leaving `attempt_count = 0`, then proves the same row can be reclaimed after
  lease expiry alongside the established five-attempt dead-letter ordering fixture.
- Verification: focused `./transport ./save`, full Compose Postgres integration, and transport
  `-race` all pass through repository-root commands.

## 2026-07-30 — independent review: 82ba351 (poison vs transient relay failures)

**Approved.** `DeferReceiptClaim` extends the claim (bounded 100 ms–5 min, token-checked, excludes
published/dead rows) without touching `attempt_count`; the deterministic class is exactly
`errors.Is(err, ErrInvalidPolicy)` — the envelope/encode/size poison family — while publisher
infrastructure and Mark failures defer on a 1 s backoff with later founders' claims still released.
This matches the round-2 ruling: only deterministic failures spend the five-attempt dead-letter
budget. The per-founder head stays claimed through a transient outage, which is the ordering
invariant doing its job. One note: the backoff is flat 1 s (no exponential); acceptable at Phase-0
scale, revisit with the multi-instance relay RFC. e42248b (root-level command convention in
AGENTS.md) is a docs-only workflow rule — no code verdict needed.

## 2026-07-30 — independent review: 728cd96, 346ec8d, 929b078 (writer boundary + soak + Compose)

**Verdict: approved — all four 728cd96 claims verified with evidence, and the fixes are structural,
not cosmetic.** Readiness is irreversible via a never-cleared `draining` flag under one mutex (the
mid-drain flip is impossible by construction); queue releases are keyed to per-revision reservation
counts with a zero-count guard, so command replies, rev-0 courtesy frames, recovery replays, and
equal-rev rejected receipts can neither double-decrement nor leak (attack test present; `-race`
clean including the disconnect-during-flush path); protobuf frames get a real `protocol.Reply`
decode from the same library centrifuge encodes with — the fail-open is now fail-closed; the jsonb
size guard delegates to Postgres inside the same INSERT (`octet_length($::jsonb::text)`), matching
the CHECK by construction with a clean typed error instead of a transaction abort, proven by a
genuine 5,500-key attack fixture against live Postgres. The soak's new oracle (strictly-increasing
world snapshots, terminal rev, wrong-channel/kind fails) matches the drop-stale contract and passes
under `-race` in 3.85 s — fixed for the right reason, leak detection intact. Compose is infra-only,
byte-identical test command, no new silent-skip path.

Findings (minor):
1. LOW — malformed publication metadata disconnects with code 4000 ("queue overflow") — a false
   diagnosis for ops and clients; give validation failures their own close code.
2. LOW (latent) — `ReservePlayer` reports rev<1 as overflow → a future rev-0 publish to `player:*`
   via the exported Publish would 4000 every subscriber; unreachable today (DB CHECK rev>0 on the
   only caller). Separate validation from overflow in the reservation API.
3. INFO — soak sniffer stops at the terminal rev (post-terminal leaks unobserved; primary guard is
   upstream); readiness race test wouldn't catch mutex deletion (guarantee verified by inspection);
   integration suites still silently skip without TEST_DATABASE_URL outside Compose/CI
   (pre-existing; the dedicated target and CI cannot skip).

## 2026-07-30 — round-3 boundary and T3 remediation

- Malformed JSON/protobuf publication metadata now disconnects with typed code 4004
  `invalid_frame`; code 4000 remains exclusively queue capacity. `ReservePlayer` returns a separate
  validation error for non-positive revisions, and the publisher rejects that input without
  disconnecting unrelated subscribers.
- The accepted T3 scope ruling had not reached either wire validator or its shared fixtures.
  Event payloads now require the exact field `scope` with closed values `company|founder` in Go and
  TypeScript; missing and unknown scopes are negative vectors. This repairs the boundary needed by
  the still-unimplemented durable event relay.

## 2026-08-01 — unified durable player event/receipt relay

- Replaced the receipt-only queue with one `transport_player_outbox`. A trigger on the authoritative
  `events` table makes every Founder/Company event writer participate without a second hand-maintained
  call path; receipt rows use the same transaction and queue. The outbox identity is now the sole
  per-Founder delivery lane across scopes; T3's per-scope revision cursors remain authoritative when
  independent Founder and Company transactions interleave.
- Event messages freeze `scope` and that scope's revision in their exact payload. The Exit integration
  proof asserts the committed order founder event → final Company events → next-run Company event →
  receipt; ordinary intent coverage proves a Founder cannot claim its receipt ahead of its event.
- The live relay now maps both closed kinds, preserves the existing lease/transient/poison semantics,
  and emits the generalized `player_message_dead_letter` invariant. Migration 00040 preserves every
  existing receipt row and its claim/dead-letter state.
- Focused unit suites and the real-Postgres save/production integration surface pass through the
  repository-root commands. Plan item 3 flips in this proof-carrying change; `cmd/gameserver` and the
  world snapshot driver remain the forward composition work.

### Self-review follow-up — stream-scoped receipt identity

The post-commit seam audit caught a HIGH before review approval: migration 00040 initially declared
`UNIQUE(message_kind, source_id)`, but receipt source IDs are client-supplied intent IDs whose
authority is only per stream. Two Founders may legitimately submit the same UUID. Append-only
migration 00041 narrows uniqueness to `(message_kind, stream_id, source_id)`; event IDs remain safe
under the same key. A live-Postgres regression inserts the same intent ID for two independent
Founder streams and requires both durable receipt rows. Migration 00040 itself also declares the
stream-scoped key so a fresh install can copy pre-existing same-ID receipts before 00041 runs;
00041 is conditional/idempotent so databases that already applied the original 00040 are repaired.

### Independent review correction — ordering and rollback safety

The review correctly rejected the phrase "sole per-Founder delivery order": sequence identities are
allocated inside transactions, so independent Founder and Company commits may become visible in a
different cross-scope order. T3 never made that order authoritative—the client owns one gap cursor
per scope—so the contract and canonical docs now say exactly that: one durable delivery lane, with
scope revisions as reconciliation authority. Exit rows written by one transaction retain their
asserted local order.

Migration 00041's Down path no longer attempts to resurrect the invalid global source-ID uniqueness
constraint. It restores the stream-scoped key declared by the current 00040 source, so a database
containing two legitimate streams with the same client intent ID can roll back without data loss or
migration failure.

## 2026-08-01 — designated reviewer: transport findings from the 0.3.x span review

1. **F2 HIGH — c782b57's per-scope revision-ordering claim is not enforced for projector-written
   events**: route/commons compensation events insert at HISTORICAL revisions in later
   transactions with no stream lock; outbox delivers in outbox_id order; a founder at rev 9 can
   receive a rev-5 compensation event, and the T3/T4 gap-cursor contract defines forward gaps
   only. Fix: either the relay reorders per (founder, scope) before publish, or T4 gains a
   declared backward-compensation rule the shell handles (surface, don't improvise — owner
   contract either way).
2. **F3 HIGH — the event-size CHECK (00040) can abort authoritative game transactions**: no
   application-side pre-measure exists on the event path (receipts have one). An oversized event
   payload hard-fails every affected intent/exit/projection commit. Fix: pre-measure at event
   write with the typed rejection, mirroring the receipt guard.
3. F8 MEDIUM batch: 00040 Down destroys undelivered event rows (restore both kinds); no
   event backfill on Up (declare); receipt-outbox drop in the same migration is
   single-process-migrate-at-startup-only (document as a deployment invariant); the claimed
   client per-scope gap cursor does not exist in client/src (implement or retract from the
   commit narrative); in-place edits of applied migrations across the span noted (goose doesn't
   checksum — convention: never edit an applied migration file).
4. F12 LOW: receipt scope hardcoded 'company'; dead-lettered events give the client only an
   eventual forward gap (acceptable, now declared).

## 2026-08-01 — F2/F3 remediation implemented

**Recorded by: root Codex. Review by: pending independent diff review.**

F2 is closed by wire v2's declared historical-compensation rule. Migration 00042 labels existing
and future queued events; Go/TypeScript exact validators bind `compensation` to `historical`; and
the concrete client cursor delivers those events without mutating either scope revision. Forward
events remain gap-checked and same-revision event IDs remain independently deliverable.

F3 is closed at the ownership boundary rather than by letting transport reject authoritative
history. The new constraint keeps receipts capped but permits event rows of any size to enter the
outbox. The relay's existing deterministic failure budget dead-letters an unencodable event and
the subsequent forward gap invokes full sync. Real Postgres proves an oversized compensation
commits and oversized receipts remain rejected. Applied migration 00040 was not edited.

Verification: focused Go transport/save tests, TypeScript typecheck, 6,494 client tests, shared
schema validation, and the complete Docker/Postgres integration target are green.

## 2026-08-21 — Q-003 production recovery predeclaration

- Q-001 received the designated cross-party approval at `34d04a5`; corrected Q-002 and its
  separately authorized API Foundation tightening received designated approval at `bfd9b65`.
  Q-003 is therefore the only serial witness batch authorized to begin.
- Re-derived the lifecycle audit against the current runtime. The server already owns positioned
  Centrifuge history, typed closes, live queue discipline, and bounded drain. The production Game
  UI consumer still discards subscription positions and publication offsets, never requests
  recovery, never reconnects, ignores close codes, leaves `resume_after_ms` unscheduled, and does
  not bind the existing per-scope revision cursor.
- Predeclared one recovery authority and the missing-position, duplicate, gap, expired-history,
  overflow, drain, credential-expiry, literal stalled-world, and non-member Guild negatives in the
  plan before touching runtime behavior. Account token rotation, a second API client, snapshot-v3,
  player copy, and Transport archival remain forbidden.

## 2026-08-21 — Q-003 consumer implementation, evidence, and AC3 fired control

**Recorded by:** Codex. **Review by:** Codex first-filter only over `bfd9b65..c63e7e6`; designated
cross-party review is not yet requested because AC3 is owner-blocked.

- `c63e7e6` binds the existing `PlayerRevisionCursor` to the production Game UI runtime and persists
  per-Founder player/world Centrifuge epoch/offset positions. One controller owns initial subscribe,
  ordinary reconnect, recovered-publication delivery, full sync, resubscribe, typed closes, and D4's
  advertised drain delay. It uses the existing authenticated snapshot operation and changes neither
  Account rotation nor snapshot schema.
- Unit/browser populations cover no saved position, valid position persistence, unknown-kind
  position advance, same-revision duplicate suppression, historical delivery without cursor
  movement, forward gap, unrecoverable history, queue overflow 4000, invalid frame 4004, drain 4003,
  auth expiry 4001, and replacement 4002. Chromium/Firefox/WebKit ran 20,019 tests green; client unit
  ran 6,659 tests green with 15 skips; typecheck, production build, and shell boundary passed.
- The composed browser witness now closes the production socket, commits a real Postgres intent in
  the disconnect window, requires the exact saved player epoch/offset in `recover: true`, consumes
  the missed receipt, advances the persisted offset, and lands `/api/v1/founder/state` on the exact
  committed revision. Removing the recover command fails; skipping recovered publications fails.
- Cold Docker `./transport ./account ./gameserver -count=1` passed (4.131 s, 1.232 s, 28.524 s).
  An earlier same-population run reached the Account/Gameserver revocation witness but its two-second
  drain cleanup timed out; the isolated witness then passed in 17.634 s and the restored full cold
  population passed. This timing failure is disclosed, not counted as a green run.
- AC6 is closed by a real second-account Guild negative beside the member control. Temporarily
  forcing the production Guild adapter to authorize every Guild makes that witness fail immediately.
- AC3 fired instead of passing: extending the actual connected slow-reader probe to 100 world
  publications over 11.29 seconds ended in socket EOF. Centrifuge admits frames to its byte queue
  before `OnTransportWrite` can discard stale frames; the short in-flight-plus-newest test therefore
  cannot prove exact-one delivery for a genuinely blocked writer. The experiment was restored, RP-054
  remains open, and canonical docs disclose the result. Browser 4000 full-sync recovery is a valid
  user recovery path but is not relabeled as AC3.
- No plan checkbox, RFC status, archive, player copy, API schema, Account token behavior, snapshot-v3
  mechanic, deployment, or push changed. The next action requires the RFC/ruling author to choose a
  pre-queue authority or reconcile AC3; only then can Q-003 complete and enter designated review.
- The old platform-audit `final-contradiction-validator.mjs` now exits 1 at its intentional
  planning-only range guard because Q-001–Q-003 have since added product/authority paths. It is not
  a Q-003 gate and was not misreported green; its original audited-coordinate evidence is unchanged.

## 2026-08-21 — AC3 dependency-seam exhaustion

**Recorded by:** Codex. **Review by:** pending cross-party review with the eventual Q-003 closeout.

- Read the pinned Centrifuge v0.38.0 writer and experimental channel-batching implementations rather
  than inferring their behavior from API names. `FlushLatestPublication` replaces publications only
  inside one fixed `MaxDelay`/`MaxSize` batch, before that batch enters the per-client byte queue.
- That configuration is channel-wide, not selected for a slow subscriber. A ten-second delay would
  make every healthy `world` subscriber violate D2/T5's 4 Hz live contract; a short delay continues
  flushing batches into the stalled client's bounded queue and therefore reproduces the measured
  overflow instead of enforcing AC3.
- The connection writer removes as many as 16 admitted items before invoking `OnTransportWrite` and
  then calls the socket transport. The dependency exposes neither pending-item replacement nor a
  post-write/client-consumption acknowledgement. The current hook can suppress stale queued frames
  after an in-flight write returns, but cannot retract that write or know when the consumer resumes.
- No supported pinned-dependency seam implements literal per-client ten-second connected-stall
  exact-one delivery without weakening healthy-client cadence. A custom writer/queue, dependency
  change, or protocol acknowledgement would be new architecture and is not authorized by Q-003.
  The owner decision recorded under RP-054 therefore remains genuine, not an unsearched code path.

## 2026-08-21 — AC3 owner reconciliation adopted

**Author/ruling by:** Marco. **Recorded by:** Codex.

- The owner rejected the packet-count criterion as over-engineering for an idle game and adopted the
  player outcome instead: after at least a ten-second stall, the client must converge to newest
  authoritative world state without unbounded backlog or lost committed progress. A bounded typed
  4000 disconnect followed by the normal reconnect/full-sync path is explicitly valid.
- Reconciled D2 and AC3 in the same edit. This does not authorize a dependency fork, custom writer,
  weaker queue bound, Account token rotation, or a second recovery path. Q-003 is unblocked only for
  the literal ruled-outcome witness and its failing mutation before the designated review.

## 2026-08-21 — AC3 close-frame correction from the first ruled witness

**Author/ruling by:** Marco's adopted player-outcome rule. **Recorded by:** Codex.

- The first literal ten-second run ended in plain socket EOF rather than an observable 4000. A
  fully blocked socket cannot be required to deliver the close frame that reports its own overflow;
  ordinary browser WebSockets provide no such guarantee. The initial wording had accidentally
  reintroduced a packet-level requirement the owner had just rejected.
- Reconciled AC3 again to accept typed or abnormal disconnect while requiring the same bounded
  reconnect/recovery and newest-authoritative-state outcome. The failed run remains evidence and
  the witness must now prove recovery after that EOF instead of treating it as an error.

## 2026-08-21 — Q-003 AC3 witness complete; minimal review handoff

**Recorded by:** Codex. **Review by:** Codex first-filter over `bfd9b65..afb5bf0`; mandatory
cross-party verdict pending.

- `afb5bf0` replaces the sub-second probe with an actual socket held unread for 10.50 seconds. It
  accepts either a bounded live path ending at revision 93 or disconnect, then always reconnects
  from a known epoch/offset and requires latest-only recovery to revision 93.
- First run failed on EOF, exposing and removing the accidental close-frame requirement. With the
  owner outcome restored, the exact witness passes in 10.50 seconds. Changing production latest
  selection to accept the baseline but refuse every later world revision fails after 11.00 seconds
  with no convergence; production bytes were restored and the witness passed again.
- Restored cold gates: Transport 13.544 s; client 6,659 pass/15 skip; browser 20,019 pass/3 skip;
  performance 1 pass/10 skip; typecheck, boundary and production build green; composed browser /
  Postgres missed-receipt recovery PASS; sequential Postgres Transport 13.752 s, Account 1.488 s,
  Gameserver 24.916 s.
- No production byte changed in `afb5bf0`; the test exercises existing bounded queue, recovery
  history and production browser behavior. Account rotation, snapshot v3, player copy, archive,
  deploy and push remain untouched. Q-003 now needs only Claude's exact-range verdict; no broader
  work is delegated to Claude.

## 2026-08-21 — Q-003 handoff-range edge correction

- `08acc3e` recorded the evidence after `afb5bf0`, so the earlier implementation endpoint cannot be
  used as the designated review endpoint: it would omit its own record. No history was rewritten.
- The review begins after approved Q-002 tip `bfd9b65` and ends at this correction commit. The relay
  names that literal hash; Claude verifies only that complete range and returns one verdict.

## 2026-08-21 — Claude designated cross-party review of Q-003 `bfd9b65..65b2506` — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Range:** `bfd9b65..65b2506`, eight commits from predeclaration through the handoff-edge
  correction (whose reasoning — the endpoint must include its own record — is right).
- **Scope against the manifest:** Q-003 is the production-authority batch, so `runtime.ts`,
  `GameUIApp.svelte`, transport, docs, and RFC-body changes are in-lane. Checked the
  `must_not_change` list specifically: no Account token-rotation change, no second API client, no
  snapshot-v3 mechanics change (the +3 Svelte lines are recovery-lifecycle wiring), no player
  copy, no Transport archival.
- **AC3 replacement provenance:** properly a fired blocker first ("owner-blocked", review not
  requested), then an owner ruling ("Author/ruling by: Marco. Recorded by: Codex") rejecting the
  packet-count criterion for the player outcome, then a body reconciliation implementing that
  ruling — including Codex catching its own first witness draft reintroducing the rejected
  packet-level requirement. I cannot verify an out-of-band owner ruling from inside this session;
  it is recorded explicitly, is consistent with the owner's standing positions, and the owner can
  contradict this entry if it misstates their ruling.
- **Executed:** the stalled witness cold — `--- PASS` in **10.46 s**, a real socket held unread
  with ~3 MB of world frames published against it, not a simulated clock.
- **Adversarial probing, reported in full including my misses:**
  1. Severing the stale-world write filter: witness PASSES — legitimately, via the disconnect →
     recover arm of the ruled disjunction. Robustness, not vacuity.
  2. Severing `ReserveWorld` monotonicity: PASSES — same legitimate arm.
  3. **Severing world recovery history (`node.go` world publish `WithHistory(1, …)`): FAILS
     immediately** — "recovery did not converge". That is the load-bearing convergence property
     the owner's AC3 names, and the witness discriminates on it. This independently corroborates
     the recorded severing evidence (refusing later revisions → no convergence after 11.00 s).
  4. Process disclosure: one of my probe mutations briefly landed in the main tree via cwd drift
     (a type error, caught at build, reverted immediately; tree verified clean). Mine, recorded.
- **Observation (non-blocking):** the live arm's boundedness assertion (`len(revisions) <= 2`) is
  reachable only when the stall does not overflow the connection first; with heavy padding the
  witness routes through the disconnect arm, so write-path boundedness continues to rest on the
  queue's unit population rather than this witness. Consistent with the ruled disjunction; noted
  so nobody later reads this witness as covering the backlog bound.

**Verdict: APPROVED.** Q-001/Q-002/Q-003 are all closed. No archival, promotion, or push is
authorized by this verdict.

## 2026-10-06 — browser lifetime boundary, diagnosis only (RP-048 / RP-234)

Separate from the closed Q-003 range and pending Garden implementation. Its
predeclaration lives in Account's current plan/log. T1 consumes Account tokens;
existing T4 owns reconnect/history/full-sync. Live `Node.OnRefresh` rejects,
while `OnAlive` checks the original access token against Account. The browser
has neither HTTP refresh nor rotation/reconnect binding. Observe the built client
under one exact disposable-test access-row expiry and an explicit manual HTTP
rotation control; do not change production callbacks, timing, origin, close
codes, stream positions or revocation. A new draft, not Q-003, must specify any
automatic browser renewal policy. No designated approval/archival is claimed.

## 2026-10-13 — Strict Transport policy member admission (RP-560)

At `95d390fc`, the previous goal turn constructed Combat's common lane row. Fresh read-only
Docker ping71399 exits28 after5003ms/zero bytes, so the pending declared backup/restore proof
does not run; no restart authority, cleanup or host-DB substitute is inferred. Continue accepted
Transport D5/T5 conformance. The raw policy reader now uses its DTO tags as the sole field-name
authority, refusing case aliases and raw/escaped duplicates before Go can merge them. Every
required member is present once; typed numeric/origin validation is unchanged. No limits,
origins, policy/artifact bytes, wire schema, queue/recovery behavior or credentials are changed.

Old-loader65430 fails48 cases across all12 fields, returning complete policies with nil errors.
Its twelve escaped-canonical-name controls still execute and pass. Corrected focused42292
passes all cases plus the existing unsafe-origin/unknown-field checks,0.285s; race26561
passes1.348s. Root cold `make test-go GO_PACKAGES='./transport ./account ./gameserver'
GO_TEST_FLAGS='-count=1'`92031 passes13.233/0.490/0.370s; affected vet98701 exits0.
The actual HTTP/WebSocket origin/authentication/member-subscribe/no-client-publish check59828
explicitly passes0.172s. Its `testhttp` uses net.Pipe and controlled identity/membership resolvers:
real protocol framing, not TCP, real Account/Postgres authorization or a deployed player journey.
The larger package population retains recovery/queue/drain checks; no new browser claim.

Native integration selector71359 shows the Account/Gameserver DB-dependent cases skipping
because TEST_DATABASE_URL is absent, and the one-minute cleanup lane not requested. Package
exit0 and a parent with only skipped children are not SQL evidence. No Linux/hosted CI, R-006,
full feature acceptance or release-floor promotion. Go formatting/source/diff checked; all
named handles terminal. Docs/plan and RP-560 updated together, without redundant board rewrites.

Review by: Codex (repair implementer first filter). Recorded by: Codex. Complete source/tests/
docs/records range after `95d390fc` through this containing commit requires designated Claude
review, consolidatable with related conformance fixes while preserving their own ranges.
Earlier Q-003 approval is unchanged and does not approve this correction. Continue accepted
work and actual declared service evidence when available; no push, publication or deployment.

## 2026-10-13 — Known-payload producer/consumer conformance (RP-561)

Baseline `b8eb4c58`; preceding goal turn corrected policy member admission. Typer's accepted
TT-PA4 still conflicts with API C2's pinned v1 request/response contract, as its existing log
records; do not bypass that pin. Continue accepted Transport D3/T3 construction/conformance.
Go now checks Transport-owned DTO fields, including nested world aggregates, before struct
decoding: required keys, exact names and nullability. Existing nullable active_id stays valid.
Envelope rev, presence count and restart delay use the existing maxWireInteger domain. Client
scope/cursor-effect/code require strings, not String(array). No new field, envelope version,
origin, queue/retry/auth policy, authoritative state change or raw-wire duplicate grammar.
Receipt/save/event-owned opaque objects remain opaque; invalid output fails before publishing.

New shared literal corpus174 (12 valid/162 invalid) covers exact/missing/null/aliased fields,
world nesting, actual enum types and safe-integer boundaries. The original Go JSON result was
tool-truncated, so its captured subset is not the complete baseline census. Re-execute the
original `b8eb4c58` wire.go through a Go compiler overlay at
`/tmp/cloud-clicker-wire-baseline.vbwpzx/overlay.json`, never mutating production:14950 exits2,
102 vector failures, all12 positive shapes retained, and the running publisher admits missing
presence count with nil error. Its complete compact output has4273 tokens, not a truncated
census. Old Node40881 independently fails4 enum-type cases while180 checks pass. Old failed
assertions do not prove subsequent source-preservation or healthy-delivery assertions executed.

Corrected focused58714 passes all174 vectors plus connected publisher proof: refused frame
consumes no history offset; a valid presence reaches the real HTTP/WebSocket pipe peer with
offset1 and exact payload. Controlled identity/membership, not real Account/SQL authorization.
Root cold Transport/Account/Gameserver52568 passes13.385/0.489/0.396s; vet exits0. Focused
race57003 passes1.472s. Node focused184 passes; native Chromium/WebKit78231 executes476 wire/
runtime checks,2.60s. Types/Svelte46958 has zero diagnostics. Build/client-boundary18612 pass
(217 modules); full test-client18164 passes11173, with1040 existing skips explicitly excluded.
All handles terminal; formatting/diff checked. No current SQL/composed, Linux/Firefox, hosted
CI, clean-host or full Transport/1.0 claim. Known history/conservation failures remain open.

Review by: Codex (repair implementer first filter). Recorded by: Codex. Complete source/fixture/
tests/docs/records range after `b8eb4c58` through this containing commit needs designated Claude
review, consolidatable with RP-560 without absorbing older review debt. Canonical docs, owning
plan and RP-561 updated together; no feature checkbox, body/ruling change, archival, mint,
push, publication or deployment. Continue accepted work; public Typer/API and other named
contract choices remain explicit, and the full nine-tier1.0 objective remains active.

## 2026-10-15 — Remove the stale server-composition blocker

The plan's unchecked item4 still said the binary was blocked by the Commons owner
resolver. The preserved2026-07-30 ruling already authorized composition, and current
`gameserver.New` constructs the Commons projector and real Guild service, wiring
both into Production. No resolver, subscriber policy or runtime was changed here.

Existing executed witness: current built-player aggregate22749 at `6201bef8`
passes `TestComposedGameserverPostgresSocketClearingAndGCIntegration` in both
main/Tier2 populations. Source inspection confirms actual member/Commons authorization,
three clearing-produced settlements, an HTTP intent consuming them and persisted
watermark3; readiness becomes unavailable after a bounded drain. This is not the
empty account-service settlement fixture. The Garage log owns the full execution,
architecture/setup limits and cleanup. No duplicate test or checkbox-only software run.

Verification by: Codex. Recorded by: Codex. Item4 now records implemented composition
with its named passing witness. Complete feature acceptance, RP-560/561 and earlier
corrective designated reviews remain open; no archival or approval is inferred.
Two-path editorial plan/append-only-log change; factual references/diff checked.

## 2026-10-15 — RP-651 bind cohort permission to the current run

Under Commons D1/C3 and Transport D1/T1, separate persistent cohort assignment from
current Compact permission. The resolver now checks one SQL snapshot: projected
cohort/run membership plus its owned, non-archived Company's latest run/member flag.
No Health/sample/count changes, retention choice, governance implementation or new
eviction protocol for already established subscriptions. No kernel/replay, balance,
catalog, schema, migration, workflow, copy or version bytes change.

Old-source real-Postgres session2731 fails0.824s: actual HTTP sign authorizes the
member socket, Exit saves run2/non-member, then a fresh socket's cohort subscription
is wrongly accepted. Extend the existing Exit/board witness, not a new fake resolver:
retain pre-Exit access, reject the new non-member, preserve assignment, re-sign for
successful access, leave and reject another subscription. Existing live verification,
immutable HTTP downloads, corruption refusal, historical rollover and carry checks stay.

Fixture corrections are explicit: initial34123 wrongly used the replaced Founder's
credential and failed the positive control; issue the session after replacement and
anchor its clock to today for real socket expiry. Two Loaded.Number accesses were
compile errors, not tests. Additional socket/new-run work put the old live-log check
beyond possible background compaction; move that check before the extra work, retaining
archive checks. Current-clock execution then exposes Fiscal14 rather than its expected
cap: seed the capped-carry condition explicitly from the pinned policy instead of
depending on months of application/DB clock skew. Preserve the original cap assertion;
no game policy is changed and no replay defect is inferred from these fixture attempts.

Final focused44297 PASS1.390s. Full declared `make test-save-integration
SAVE_TEST_PACKAGES='./commonsprojection ./gameserver' SAVE_TEST_FLAGS='-run Integration -v'`
1754 PASS: Commons0.343s and Gameserver60.615s, all ten ordinary Gameserver parents;
the separate opt-in real-minute cleanup test alone skips. Linux-amd64 Go image
`sha256:3343c3656871925d896336b9194aa2e3191e0a96868a9d2952a3696e86af94d0`
runs under this arm64 Docker host, not native hosted hardware. Final cold
`make test-go vet GO_PACKAGES='./commonsprojection ./transport ./gameserver'`75587
passes0.153/13.308/0.407s plus vet; native DB skips are not integration credit.

Complete root `make test-game-ui-composed`82461 exits0: both driver populations
require8/8 refresh cases and ten persisted parents, actual generated-SDK/TS current/
historical replay and corrupt receipt/event/gap/constants refusals; main transitions,
both endings/continuation/recovery, Cosmetic/Reputation/Clout, both payment negatives,
controlled Garden harvest and Tier2 candidate all pass with built assets/Worker and
real Postgres. This is native browser/Go plus container Postgres, not hosted CI,
natural growth/earnings, Headcount, actual AT or a complete release artifact.

All handles terminal; inspect/remove only the two owned disposable tmpfs Postgres
services, preserving cache/network/durable data. Diff/formatting checked. Update
canonical docs, owning plan and RP-651. Review by: Codex (implementer first filter).
Recorded by: Codex. Complete seven-path implementation/test/docs/plan/ledger/log
range after `54f352f6` through this containing commit needs designated Claude review,
consolidatable with Transport/Gameserver repairs. No self-approval, archival,
push/deployment or full nine-tier1.0/whole-CI promotion; existing holds remain.

## 2026-10-15 — Retained history does not bypass current membership

Follow RP-651 under existing D1/T1/T4, without another product change. The pinned
Centrifuge client refuses direct history RPCs when no handler is installed; normal
recovery calls the existing subscribe authorization. Execute that boundary rather
than treating source inspection or empty history as proof.

`TestActualWebsocketRecoveryRechecksCurrentMembership` uses actual embedded-node
WebSocket protocol over testhttp, with controlled identity/membership resolvers.
For Guild/cohort/match, require a real retained envelope and exact recovery while
authorized; remove membership, require103 with no subscribe result and refusal of
direct history; restore membership on the same connection and recover the same
bytes/position. A dead socket, missing history, invalid position or deny-all cannot
satisfy the population. Match authorization here is a fixture, not a live combat tenant.

Initial58281 fails the cohort observer, not authorization: a valid reply and presence
join share one WebSocket message. The test's bounded command-ID reader now decodes
the complete JSON message and ignores unrelated pushes; malformed JSON, duplicate
command replies or a missing reply fail within the existing one-second guard.
Do not change runtime framing. Focused29536 PASS0.322s; after the reader's complete-
message check, final cold `make test-go vet GO_PACKAGES=./transport`95394 PASS13.320s
plus vet. Final focused `-race -count=20`20497 PASS1.878s (all three namespaces per
repetition). No production defect/bypass is found and no source mutation is claimed.

Test/helper/docs/plan/log only; no runtime, policy, API, kernel, data or workflow
change. The preceding completed Postgres/built-player proof remains separately scoped;
do not rerun it for this controlled wire test or infer new DB/browser/hosted/release
coverage. All handles terminal; no Docker services or durable resources touched. Diff and
formatting checked. Review by: Codex (test implementer first filter). Recorded by: Codex.
Entire four-path range after `c8ed9421` through this containing commit needs designated
Claude review, consolidatable with the preceding Transport repair. No acceptance,
archival, publication/deployment or full1.0 claim; continue other accepted work/review.

## 2026-10-15 — RP-651 designated corrective-code review: APPROVED

Review by: fresh independent Codex sub-agent `/root/cohort_authorization_corrective_review`.
Recorded by: Codex parent, transcribing its verdict under Marco's corrective-review delegation.
Reviewed complete eight-path range:
`54f352f62759a4ac7aa49fbf48e9f18e1c242b60..aa2f97dfc441dde924156849cbab549b75a5dd23`
(base exclusive, tip inclusive), commits `c8ed9421` and `aa2f97df`.
Verdict: **APPROVED**, no actionable findings.

Independent cold Commons/Transport/Gameserver checks and vet PASS0.292/13.432/0.391s.
Actual declared Postgres integration PASS Commons0.631s/Gameserver58.687s,
including the expanded sign/Exit/re-sign/leave socket witness. Only the opt-in
one-minute cleanup test skips. The controlled nonempty-history recovery population
passes `-race -count=20`, all three namespaces,1.954s. Exact old projector source
through a temporary Go overlay fails0.91s with "previous-run signature authorized
new non-member subscription"; the same corrected-source DB test passes1.45s.
An initial malformed Make selector executes zero tests and receives no credit.
Gofmt/range whitespace are clean; reviewed production/test bytes match the tip.

The reviewer edits no tracked files, removes its temporary overlay and its owned
disposable Postgres container, and verifies absence; shared caches/network and
unrelated work remain. Native DB skips receive no SQL credit; the actual run above
provides it. Recovery resolvers, including match authorization, are controlled.
No independent built-browser or hosted rerun, broader Transport acceptance,
specification adoption, archival or release approval. Current plan and RP-651 now
point here; preceding review debt remains separate.

## 2026-10-15 — RP-560/RP-561 independent corrective approvals

Review by: fresh independent Codex sub-agent `/root/account_transport_shape_corrective_review`,
under Marco's corrective-code delegation. Recorded by: Codex parent transcribing its verdict.
**APPROVED**, no findings, separately for each complete code/fixture/test/docs/record range:

- RP-560, six paths: `95d390fc96328d1c46e4d1871d15638d8c092d80..b8eb4c58756f9b1da92e6e3f47e60b4606739278`.
- RP-561, nine paths: `b8eb4c58756f9b1da92e6e3f47e60b4606739278..6a5a742ad88d954843a9372b368b3d23c9a56ed3`.

At descendant `f60ec0d78fda8fe4ff62bf005ef89cb6dc5c3b2c`, all reviewed production/test/corpus
files remain byte-identical. Independent cold Account/Transport/Gameserver and affected vet
pass. Focused Transport race passes1.439s, executing174 payload vectors,48 policy refusals,
12 escaped-name controls and publisher/history refusal plus healthy offset1 delivery.
Explicit actual WebSocket origin/auth/subscription/no-client-publish and existing policy
controls pass0.156s. Node transport-payload-conformance/transport files pass184; the extra
nonexistent transport-runtime.test.ts selector adds no coverage. Types/Svelte have zero
diagnostics; build passes217 modules. Exact-range diff/source checks pass, tree unchanged.
An initial unquoted race alternation became shell pipelines and did not run the intended
tests; only the corrected quoted selector counts.

Historical48 policy,102 Go payload and four client failures were inspected in prior records,
not independently rerun. Current connected publisher/HTTP peer has controlled resolvers,
not real SQL authorization or a browser journey. Native Integration selector skips SQL
children for absent TEST_DATABASE_URL; no SQL credit. No browser/Docker/Linux/hosted CI,
API generation, whole feature, specification adoption, archival or release approval.
Existing arithmetic/history red gates are not rerun or concealed. These two corrections
no longer wait on Claude; other Transport review debt and policy contracts remain separate.


## 2026-11-04 — RP-765 cancel periodic authentication before node shutdown

Accepted Transport T1/T5/T6/AC5 authorizes this corrective conformance repair.
Centrifuge holds presenceMu while OnAlive calls Account authentication, and client
closure waits for that mutex. Background makes a blocked query outlive shutdown.
Node now owns/cancels that query context before engine shutdown. A canceled node
leaves closure to shutdown instead of mislabelling it auth-expired; ordinary
failures retain4001. No token/TTL/cadence, API, schema, kernel, balance or CI change.
Transport is outside the actual kernel-affecting registry; kernel0.3.225 stays.

Actual registered-callback controlled-authenticator test72276 fails old source
with shutdown deadline14.45s; corrected8208 passes15.10s. It uses unchanged actual
presence cadence and HTTP/WebSocket over net.Pipe, not direct callback invocation.
New real composer/Account/Postgres/native-socket test58943 passes18.27s: keep
access_tokens exclusively locked, observe the actual periodic SELECT lock wait,
drain under the existing2s witness bound, require wire4003/query cancellation
while the lock remains held and no token revocation. Exact-old production overlay
58036 fails drain deadline22.40s. Overlay changes only node.go, not the test or
whole historical tree. Source1992ce15 old-node SHA256
74d947ff192c6c145e05c8539863fde8bd255ee26ce31e0c65b3c88edcb989bd;
recipe/bytes retained under ignored.cache/transport-auth-shutdown-1992ce15/.
An exploratory immediate-close hypothesis survived200 executions and was removed,
not called a bug/fix. An initial Make-anchor selector executed zero tests; an
unused test import first failed compilation. Neither supplies executed evidence.

Finished native `make test-go GO_PACKAGES='./transport ./gameserver ./account'
GO_TEST_FLAGS=-race`78139→0 (39.262/1.911/4.470s); DB cases can skip there.
`make vet GO_PACKAGES='./transport ./gameserver ./account'`→0; gofmt/whitespace pass.
Full actual `make test-go-ci CI_TEST_PACKAGES='./transport ./gameserver ./account'
CI_TEST_FLAGS='-race -v'`80563→0, Linux/amd64 image593dc3550e0514c6ec0f7861c9bf7bf415aeabc9f4852f83f7469bfe06db1335,
real Postgres16 image6a2e324ebe59af6031c4c1cd96c1d5072c2ebc36563d2196365e43b8297c062d.
All three packages pass41.480/168.896/23.109s,647 PASS rows (including subtests),
zero FAIL rows. Only existing opt-in one-minute cleanup-runtime case skips.
New callback18.12s, actual SQL/drain24.11s and existing family-revocation20.05s
all pass, not skip. Initial Docker-socket sandbox denial runs no tests; permitted
identical command supplies80563. Retained linux-race.log SHA256
 e06e9532d7d7968dd1fbb492fdafc8b6a0458ff1650d3a6d3974e79a0a04ea2e.
Source/tests remain fixed through these runs; all handles terminal. Owned test
Postgres76ba2235783f remains for independent review, not a live test execution.

Review by: Codex root (implementer first filter). Recorded by: Codex root.
Complete seven-path range after1992ce15 through this batch awaits independent
corrective review. Canonical docs/plan/ledger reconcile in this commit; no feature
checkbox, archival or release promotion. RP-450 original cause, whole CI/latency,
owner adoptions and full nine-tier1.0 remain open. The retained893bb1e7 release
artifact predates this runtime correction; rebuild at the next release boundary.


## 2026-11-04 — RP-765 independent corrective approval and cleanup

Review by: independent Codex `/root/socket_auth_shutdown_corrective_review`.
Recorded by: Codex root, transcribing that agent's actual verdict. APPROVED, no
findings, complete seven-path1992ce150428f5d18e70cc809965315c39e0e85d..
034ea3cdf6efa5482474b4fb06197f823579b1f6. Full diff, Account SQL, actual
Centrifuge presence/close mutex behavior and evidence boundaries inspected.
Independent Linux-amd64/Postgres/race86752 passes actual callback18.57s and
blocked-SQL/socket23.32s; actual family-revocation80981 passes25.40s, not skip.
Affected vet48dfbd passes. Exact-old overlay94348 exits expected2: BOTH unchanged
regressions fail their intended shutdown/drain deadline17.86/18.75s. Reviewer
independently confirms old bytes match1992ce15 and the recorded old-source hash.

Root's full three-package race647 PASS/zero FAIL/one opt-in skip evidence is
inspected/reused with explicit attribution, not called reviewer-executed. No
reviewer edits/commits; reviewed HEAD and worktree remain fixed through checks.
All reviewer/root execution handles are terminal. Root verifies no test workers
remain and exact owned PostgreSQL76ba2235783f belongs to this repository's
compose.save-test.yml, removes that disposable container, and verifies only the
two pre-existing BuildKit containers remain. Shared Go cache/builders untouched.

Plan/ledger now close RP-765's corrective scope; this record-only commit needs no
software rerun. Original RP-450 cause, full Transport acceptance, current release
artifact rebuilding, whole hosted CI/latency and all broader1.0 gates remain open.
No publication, archival, acceptance promotion or goal pause.


## 2026-11-04 — RP-766 cancel pending connects and close late registration

Accepted Transport T1/T5/T6/AC5 owns this corrective lifetime boundary. Unlike
RP-765's periodic callback, initial authentication uses only the request context
and is not yet in Centrifuge's authenticated-client hub. Actual cold43276 fails:
its query survives completed shutdown, then both cancellation-aware and delayed-
valid-result arms receive successful connect replies. New code links caller and
node cancellation, rejects results after shutdown, and closes late registrations
in actual OnConnect before allocating queues/handlers. No token/TTL/cadence,
public schema, numeric/kernel/epoch, gameplay or workflow change.

Corrected pending-connect arms35726 and handoff26095 pass with race detection.
The handoff deliberately replaces OnConnecting to model a valid result crossing
library registration after shutdown: it requires a real connect reply followed
by4003, zero hub connections/queues. This is not Account authentication proof or
an assertion that no already-produced frame can follow shutdown. Exact-old-node
Linux overlay46113 fails both pending arms plus handoff; the unchanged prior
periodic test is explicitly deselected only in this bounded negative command.
Old node bytes equal082de7a7, SHA256
08f8a123f6c1657be544ee28a2b0e93ffa5f8a28d526fc8e92bcafd32e4bedc0,
retained with overlay recipe under.cache/transport-connect-shutdown-082de7a7/.

Real Linux-amd64/Postgres/socket test81467 passes1.04s. The existing locked-query
fixture now also holds access_tokens before the initial connect request, observes
the actual Account SELECT wait, then requires cancellation/4003 while still locked
and unchanged revocation state. Its original periodic population remains intact;
closed-socket observation now explicitly shares the existing2s drain bound.
Exact-old overlay83739 fails `socket survived completed drain`3.02s. These controls
replace only node.go, not assertions or the whole historical source tree.

Finished native full Transport/Gameserver/Account race87409 passes34.561/1.889/
4.565s; DB cases can skip there. Affected vet passes; gofmt/whitespace pass.
Actual complete `make test-go-ci CI_TEST_PACKAGES='./transport ./gameserver ./account'
CI_TEST_FLAGS='-race -v'`63809 terminates0: Linux/amd64/Postgres, three packages
41.662/172.470/21.902s,652 PASS rows including subtests, zero FAIL rows. Only
existing opt-in one-minute cleanup-runtime case skips. New pending/model cases,
actual initial SQL1.29s, original periodic SQL21.72s/callback18.41s and actual
family-revocation16.31s pass, not skip. Retained linux-race.log SHA256
 df9b471c0cb1446c9c0fb86aee0093abb11eaead8a6b7fb69d354b183f75fa7f.
Source/test bytes remain fixed through each run; all handles terminal. Root-created
Postgres e640901690a1 remains for independent review, not as a live test handle.

Review by: Codex root (implementer first filter). Recorded by: Codex root.
Complete seven-path range after082de7a7 through this batch awaits independent
corrective review. Docs/plan/ledger reconcile here. RP-765 stays closed in its
periodic scope; RP-450 original cause, whole Transport/CI/latency, artifact rebuild,
owner adoptions and full nine-tier1.0 remain open. No archival/publication or goal pause.


## 2026-11-04 — RP-766 independent corrective approval and cleanup

Review by: independent Codex `/root/socket_connect_shutdown_corrective_review`.
Recorded by: Codex root, transcribing that agent's actual verdict. APPROVED, no
findings, complete seven-path082de7a7a4f38ca635fb9b486829d3fb664a0177..
8833d82de5c8a48b190733783e60bcd5af63d81f. Full diff, accepted authority,
Centrifuge admission/hub/OnConnect ordering and original caller/SQL contexts inspected.

Fresh independent Linux-amd64/Postgres/race77245 passes initial two arms0.13s,
real-library handoff0.03s, held initial Account SQL/drain4003 witness1.81s and
actual family-revocation23.44s, not skipped. Exact-old overlay77946 reproduces
both initial-auth failures, handoff deadline1.02s and actual SQL socket surviving
completed drain3.62s; expected Make2. Reviewer independently matches old bytes
and hash to082de7a7. Fresh affected vet/whitespace pass. First unquoted regex-pipe
selector ran no tests; corrected77245 is the executed current evidence.

Root's complete three-package race652 PASS/zero FAIL/one existing opt-in skip
log/digest is independently inspected and reused with attribution, not called a
reviewer whole-suite run. Handoff remains an explicit model permitting a produced
connect reply before closure; actual Account proof is the separate SQL witness.
Reviewed HEAD/worktree stay fixed; no reviewer edits/commits. All handles terminal.
Root confirms no test workers remain, verifies exact root-owned PosteSQL container
e640901690a1 belongs to this repository's compose.save-test.yml, removes it, and
verifies only the two pre-existing BuildKit containers remain. Shared caches and
builders untouched. Plan/ledger close RP-766; no software rerun for this editorial
record. RP-450 cause, full-feature/CI/latency, release-artifact rebuilding, held
adoptions and all broader1.0 requirements remain; no archival/publication or goal pause.


## 2026-11-04 — RP-767 complete membership-cancellation candidate, UNAPPLIED

Accepted Transport D1/T1/T5/T6/AC5 and Guild S7 own these permission/lifecycle
boundaries. Actual Guild/Cohort QueryRow calls have no cancellation parameter;
Centrifuge waits on the pending subscribe during closure. No product source is
changed. New manual Go-overlay probe and source-pinned preparation tool retain
a COMPLETE nine-file candidate: caller/node context→Memberships/Authorized→
composer→owned Guild/Cohort QueryRowContext. Existing predicates/query authority
stay; all three gated namespaces take context, with no legacy/goroutine-only
fallback. Four existing test callers adopt the internal signature; context
refusal cases extend the existing authorization test. No schema/migration,
provider, gameplay, rates, timers, kernel/epoch or workflow bytes change.

Original v1 probe45849 fails both2s caller populations (Guild3.05/Cohort2.92s).
Candidate88271 passes both real PostgreSQL/socket/race cases1.77/1.39s, with
exclusive owner-table lock retained until4003 and SQL-wait disappearance.
These are real nonmember103/public-world controls and actual permission SELECT
waits; they do not prove natural membership progression or complete AC5.
A2s caller timeout ALONE does not establish failure of the configured15s floor.
Root catches this before closeout and expands the SAME probe to four populations,
keeping the2s cases plus actual Node.DrainTimeout15s cases, without raising either.

Final v2 baseline40449 exits expected Make2: Guild caller/configured3.65/16.34s,
Cohort caller/configured3.24/16.21s. Configured cases explicitly observe successful
Drain and wire4003, then owner SQL still blocked through the original drain bound
while exclusive locks/database stay live. Candidate96527 exits0: all four cases
1.78/1.29/1.20/1.20s. Both candidate Cohort cases briefly observe surviving backend
wait after Drain returns, then cancellation completes within the bound. This is
bounded cancellation evidence, NOT an instantaneous physical-quiescence proof.

Prep pins every one of nine base source/test files to exact
c7a9e385436fb730ac360527b5346083e45c6188, checks anchors and current watched scope,
and writes ignored compiler copies only. Exact guarded production path is
server/guild/intents.go. Candidate remains UNAPPLIED pending lawful delivery and
exact eligibility under ADOPTED authority; draft Kernel History Guard Integrity
N1 preserves admission/refusal/callback behavior and is not automatically met by
unchanged math/SQL predicates. No false version, guard removal, automatic scope
exception or new owner answer is inferred. Do not repeat pending adoption questions.

Native five-package candidate race7501 passes42.725/1.925/1.271/1.246/4.775s
(Transport/Gameserver/Guild/Commons projection/Account); DB cases can skip there.
Actual candidate-overlay vet54763 passes all five. Full Linux-amd64/Postgres/race
35546 exits0:45.553/205.883/3.902/1.513/22.154s,771 PASS rows including subtests,
zero FAIL, sole existing opt-in one-minute timer skip. This broad run compiles V1,
not the later four-case V2; final focused96527 supplies V2, with the EXACT SAME
nine generated runtime/test files and patch. No broader V2 run is claimed.
Retained root.cache/transport-membership-shutdown-c7a9e385/:
- candidate-wfl5d9ks: V1 broad artifacts, bindings and v1-replay-docker.json.
- probe-v1.go.txt SHA256c71afe55355628cf3e407e61df7eafecd801e03869c05251e8d0b6c81356e027.
- candidate-lncc7irl: final V2 native/Docker overlays; bindings SHA256
  185d9b73dc7764ac69e554db5c772fdf08038ee5098c73549710c47563be2ca0.
- Identical candidate.patch SHA256dd2290a48c310334d7f754745bc8fe5273eb88f587229d0d33bfaeead1ecabf2.
- Broad linux-race.log SHA256e785941a38a4c715960c9dd9d79ef0c8a65500619d7485d9f0aa8ae0f0a861f1.
Final probe SHA25626773a4da8dd94ca2f2bf8c661041a08773bc935f34c350ee0e7a1cb7b598fef.
Prep controls: intact isolated copy produces byte-identical patch/exit0;
seeded stale source and missing Guild registry entry each refuse/exit1. Source
and scope drift cannot silently choose another patch. Main nine files remain
byte-identical to base through all runs. All handles terminal; owned PG7179d9a54f87
remains for independent review. No new compiler cache or default CI exclusion.

Review by: Codex root (implementer first filter). Recorded by: Codex root.
Complete six-path research/tool/record range afterc7a9e385 through this batch
awaits independent review. Canonical docs disclose the LIVE limitation; ledger
stays OPEN. No shipped fix, eligibility, feature/hosted-CI/native-release,
archival/publication or1.0-completion/goal-pause claim follows.


## 2026-11-04 — RP-767 candidate/probe approval; live defect and delivery hold remain

Citation correction to the preceding entry: the archived Guild resolver authority
is **G5**, not the nonexistent S7. This corrects the record, not the accepted body.

Review by: independent Codex `/root/membership_shutdown_candidate_review`.
Recorded by: Codex root, transcribing that agent's actual verdict. APPROVED for
complete six-path research/tool/docs rangec7a9e385436fb730ac360527b5346083e45c6188..
d446a17f18d207a274ab607f512c1e1d386d17ea. Complete generated nine-file candidate,
prep/probe and accepted authority inspected; no actionable technical finding.
Fresh independent final V2 Linux-amd64/Postgres/race89801 passes all four:
Guild caller/configured1.65/1.46s, Cohort1.20/1.15s. Baseline84150 exits expected
Make2, all four fail3.83/16.31/3.36/16.31s; configured cases observe successful
Drain/4003 then surviving SQL through bound. Three candidate arms briefly retain
backend wait after Drain before cancellation completes: bounded, not instantaneous.

Independent intact prep produces identical patch/exit0; stale source and missing
Guild guard each refuse1. Independent five-package overlay vet/whitespace pass.
All nine live files equal pinned base; V1/V2 runtime/test patch identical. Final
probe/bindings hashes agree. Root's broad V1 five-package771 PASS/zero FAIL/one
opt-in timer skip log is inspected/reused with attribution, not reviewer-executed
or relabelled a complete V2 run. No new natural progression/whole AC5 proof.

Reviewed HEAD/worktree stay fixed; reviewer makes no live edits or commits. All
root/reviewer handles terminal. Root verifies no test workers remain and owned
Postgres7179d9a54f87 belongs to this repository's declared test configuration,
removes it, and confirms only the two pre-existing BuildKit containers remain.
Shared caches/builders untouched. Plan/ledger reconcile this bounded approval;
no software rerun for editorial records. RP-767 stays OPEN in active product.
This verdict does NOT authorize landing, qualify the guarded Guild file, adopt
draft N1/guard authority or complete feature/CI/release/archival gates. No false
version, publication, owner-answer inference or goal pause. Continue authorized
work; lawful delivery/eligibility remains the candidate's named next action.

## 2026-11-04 — RP-768: token expiry cancels a blocked periodic recheck

Previous goal turn made progress: exact-artifact evidence independently approved
and closed at7505c5bf. This routine correction consumes accepted Account D2 and
Transport T1/T5, not an unadopted proposal. While the node remains active, a blocked
OnAlive Account read holds Centrifuge's presence mutex/timer past Credentials.Exp.
The previous node-shutdown fixes cannot cancel it merely because its token expires.

Old7505c5bf production fails the actual registered-callback/net.Pipe socket test
37063 in31.05s and real Linux-amd64/Postgres/socket46240 in32.31s: no4001 before
the observation deadline. Actual Account SQL is first observed blocked on the
exclusive access_tokens lock. The real token retains its normal900-second TTL;
controlled issuance/application-clock advance leaves its final30seconds for the
unmodified real socket/presence scheduler. This is not a full natural15-minute
session or browser-renewal proof. No background node drain rescues either oracle.
Baseline logs .cache/socket-expiry-{baseline,sql-baseline}.log SHA256 respectively
 a87f11fe4a02653b4630155ca5ffb9998329eaef1bff0d4e7ca0db5d0471d7ca /
 90050b608893ad393c77305b99dea5927e5aa09b04b1dcb804f27cbdc91ecb1f.

OnConnecting retains verified expiry beside the private token. OnAlive derives
its query context from that same expiry and existing node cancellation, never a
new timeout/TTL. No early expiry, family revocation, node stop, wire/Info/claim,
policy/cadence or kernel change. Focused7699 native29.83s and89473 actual SQL30.66s
pass:4001 and finished authentication/ceased backend wait while the original lock
stays held, no revocation and fresh session/connection afterward. Additional
no-early-expiry assertions strengthen the final tests without changing old failures.

Finished source frozen: native three-package race76374 passes Transport62.773s /
Account4.740s, then Gameserver aborts at sandbox loopback binding. Its separate
identical-package permitted run82337 passes1.817s; native DB tests skip, not SQL
credit. Complete Linux-amd64/Postgres three-package race56144 passes75.382s /
23.281s /206.212s,654 PASS rows/zero FAIL/one existing opt-in minute cleanup skip.
Existing actual family revocation and node-drain regressions execute. Full SQL log
.cache/socket-expiry-sql-full.log SHA256
3b832e58331b9c9746a13b70c29f8880f41015d4b79de4e47235829495bb4585.
Affected vet7700, formatting and diff checks pass. No unchanged broad rerun,
new mechanism/CI population exemption, budget increase or receipt/golden changes.

All root verification handles terminal; shared declared test Postgres remains for
independent review. No product/auth test bytes changed during live runs. The old
source9768d681 artifact remains valid historical packaging evidence, predating
this correction; rebuild before claiming it contains RP-768. RP-767 remains
OPEN/unapplied and separately held. This seven-path correction after7505c5bf
through the containing commit needs fresh independent corrective review. Original
Account/Transport acceptance, current hosted CI, clean-host/release and full
nine-tier1.0 remain open/active; no publication, archival, owner adoption or pause.
Review by: Codex root (implementer first filter). Recorded by: Codex root.

## 2026-11-04 — RP-768 independently approved

APPROVED, no findings requiring changes, complete seven-path range
7505c5bfc9738c96633bb0c6c8b9b01030bc92e9..
c32f69029f281481b486b0226a2c0d762bca31cb.
Review by: fresh independent Codex /root/socket_expiry_corrective_review.
Recorded by: Codex root, transcribing its completed verdict.

Fresh independent race/uncached native old31.16s FAIL→corrected29.01s PASS;
Linux-amd64 actual Postgres/socket old33.11s FAIL→corrected30.38s PASS. Both old
failures lack4001; overlays use independently extracted7505c5bf bytes, never
main mutations. Normal scheduler/no-early-expiry, lock retained through actual
SQL cancellation/4001, unchanged revocation and fresh connect are verified.
All3 native shutdown cases, both actual blocked-authentication/connect drains,
actual family revocation, affected3-package vet/format/diff pass independently.
Reviewer inspects and explicitly reuses root's full native/permitted-rerun and
654PASS/0FAIL/one opt-in skip SQL evidence with matching digest, not a second
whole-suite execution. Initial reviewer Make selector matched zero tests and
initial Docker sandbox denied execution; corrected named/permitted runs supply
all claimed evidence. Logs .cache/socket-expiry-independent-{native-old,sql-old,
native-current,sql-current,native-controls,sql-drain-controls,sql-revocation}.log.
All reviewer handles terminal and reviewed HEAD/tree stayed fixed/clean.

This closes only RP-768 and its corrective review. Root reconciles status only;
no software rerun for these editorial records. RP-767 stays OPEN/unapplied;
9768d681's retained artifact predates this correction, so no artifact coverage
is inferred. Full Account/Transport, hosted CI, clean-host/release/archival and
nine-tier1.0 remain open/active. No push, adoption, deployment or goal pause.

Cleanup verified: root-owned tmpfs Postgres7fd12c59177a removed only after all
review handles terminate. Container inventory contains only the two pre-existing
BuildKit services; no test process or game service remains.

## 2026-11-05 — RP-767 versioned correction; previous delivery hold corrected

Previous goal turn makes progress: RP-768 independently closed at215a98b3.
Review by: independent Codex /root/membership_version_authority_review (technical
interpretation ONLY). Recorded by: Codex root. Accepted KV-1 structurally requires
a same-commit version bump for changed guarded production paths; AGENTS' false-
signal prohibition concerns behavior-identical work. This correction changes
cancellation/refusal/lifecycle. Identical economics/SQL predicates do not make it
neutral. RFC-0000 Rule3 and Transport T5/T6 / Guild G5 already authorize this
conformance fix. No extra owner ruling, draft N1 adoption or guard exception is
needed. Root's previous hold was an overbroad interpretation, now corrected in
current bodies; historical logs/verdicts stay intact. Reviewer freshly executes
existing kernel-guard fixtures, including60 literal-path refusals; eligibility
is not implementation approval.

Rebase pinned preparation onto215a98b3: only node.go differs among the9 old base
files; its RP-768 credential-return/OnAlive blocks are retained byte-exact. Old
preparer correctly refuses that drift. Apply the complete9-file cancellation
patch, put the reviewed4-case probe into the ordinary integration population
(header only changes), and increment all3 version sources225→226 in this same
coherent batch. No guarded-prefix/correction/history rewrite, applied migration,
epoch8/content/copy/economics or public wire/identity schema changes. Actual
membership policy/SQL predicates stay unchanged; cancelled work now refuses.

Fresh baseline11332 at215 fails all4 actual Linux-amd64/Postgres/race arms:
Guild caller/configured3.73/16.44s, Cohort3.56/16.61s. Configured cases first observe
successful Drain/4003 then SQL surviving the original bound. Tracked correction
94516 passes1.55/1.28/1.30/1.16s with original locks held. Backend cancellation can
briefly finish after Drain returns: bounded, never instantaneous SQL quiescence.
Logs under .cache/transport-membership-shutdown-215a98b3 retain exact populations.

Finished native8-package race2431 passes; native DB cases skip. Linux-amd64 real-
PG full8-package race25947 runs ALONE and passes: Kernel1.063, Transport74.087,
Account23.444, Gameserver199.156, Guild3.906, Commonsprojection1.530,
Replayverify3.106, Replaycatalog9.383s;873PASS/zeroFAIL/one old opt-in minute cleanup
skip. SQL-full-serial.log SHA256
673ee3fd7e828d11d503bb47ec88a86271db86ad5f99d1b3f367359ebf96eef3.
Client375 transport/payload/replay tests3217, strict types40836 zero errors/warnings,
affected vet35519 and native gameserver build2617 pass. Real history guard47931
remains RED at historical50a3a514; source-version parity passes before that failure.
No whole-CI or full-history pass. Root current13 source/test/version hashes stay
fixed through failed and successful executions; final checksum assertion passes.

Invalid verification is retained, not overwritten: root full2862 and reviewer59770
were concurrent against the SAME default fixture DB. Their TRUNCATE/account/epoch
operations interfere: root Minigame409/epoch-history failures; reviewer healthy
family control4001. Both runs remain FAILED (sql-full.log and independent-current.log).
Root scheduled non-independent work in parallel; no product regression is inferred
from that setup. With all other SQL runners stopped and no source/assertion/guard
changes, reviewer33958 passes its selected cases, then root25947 passes full8.
Existing docs/ci now states the shared-DB single-runner requirement; a different
host port alone is not isolation. No new orchestration/lock/RFC/test exemption.

Preliminary independent review finds no code issue: all9 candidate bindings match,
ordinary regression matches prior body, and node.go outside OnSubscribe matches
215 exactly. Fresh serial race33958 covers actual4-case cancellation, healthy
Guild/Commons/recovery, expiry, both authentication shutdowns, family revocation,
Exit/re-sign/leave and Guild lifecycle:34PASS/zeroFAIL/zeroSKIP. It reuses root's
source-bound old proof after inspecting bindings, not a fresh old execution.
Final complete25-path implementation/test/version/docs/record range after215a98b3
still needs its committed-range verdict. Retained9768d681 artifact predates these
changes. Full feature/hosted/rights/native/R-006/release and nine-tier1.0 stay open;
no publication, archival, new owner adoption or goal pause.
Review by: Codex root (implementation/output first filter). Recorded by: Codex root.
