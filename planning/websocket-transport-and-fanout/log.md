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
