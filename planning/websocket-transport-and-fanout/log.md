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
