# WebSocket Transport & Fan-out lifecycle audit

Coordinate: product tree `190a4fa`; audit checkpoint after `2351cdb`; 2026-08-20.

This pass re-derived the active Transport RFC from its complete specification, plan/log, transport
policy and wire implementation, Centrifuge node, durable player outbox, gameserver composition and
drain lifecycle, production Game UI runtime, canonical docs, current tests, successor planning
records, history-rewrite map, and tracked review provenance. It did not edit product code,
owner-authored RFC/design text, canonical product docs, or implementation-plan checkboxes.

## Bottom line

The server transport is one of the repository's stronger foundations. A real embedded Centrifuge
node enforces authentication, closed subscription authorization, public/private channel rules,
history, bounded queues, typed disconnects, latest-only world gauges, durable transaction-owned
player messages, deterministic poison handling, 5,000 live in-memory WebSockets, a real production
gameserver composition, and a bounded server-side drain order.

The shipped browser consumer does not implement the recovery contract that makes those mechanics a
user outcome:

- it manually sends Centrifuge protocol JSON but never stores a channel epoch/offset or subscribes
  with `recover: true`;
- it opens one socket with no reconnect loop, retry policy, or resubscription path;
- it ignores WebSocket close codes, including queue overflow, auth expiry, replacement, and drain;
- it decodes `resume_after_ms` but does not schedule a reconnect after that delay;
- `PlayerRevisionCursor` exists only as an isolated class and unit test; the production runtime
  delivers events without gap detection, same-revision event-ID dedupe, or historical-compensation
  cursor semantics; and
- the UI's full-state snapshot button can respond to an explicit `resync_required` publication,
  but queue overflow and ordinary/drain close never produce that path in the browser.

Consequently AC1 is proven, while AC2–AC6 remain partial against their literal or integrated
requirements. In particular, server recovery tests do not prove a recovering client. This RFC is
not archival-ready until the accepted browser recovery behavior is wired and exercised, two narrow
literal witness gaps are closed, the stale plan/body/docs are reconciled, and the exact current
implementation span receives the required cross-party closeout verdict.

## Current cold evidence

All valid commands ran from repository root:

- `make test-go GO_PACKAGES='./transport ./gameserver' GO_TEST_FLAGS='-count=1'` — green; Transport
  4.104 s and gameserver 0.486 s;
- `make test-client` — 39 files passed, two skipped; 6,655 tests passed, 15 skipped;
- the immediately preceding Account pass ran real-Postgres `./account ./leaderboard ./gameserver`
  integration and the composed Chromium bootstrap/snapshot/Centrifuge handshake green at the same
  product coordinate.

An attempted root `pnpm --filter ...` command did not start any tests because this repository has no
root package manifest. It is an invalid invocation, not evidence; the root Make gate above is the
recorded client result.

The transport package's actual-socket witnesses exercise the JSON protocol over coder/websocket,
not only an in-memory queue model. The 5k soak is also inside the normal cold package run and
inspects each subscriber's trace across ten world ticks. These are meaningful server results, but
their client roles are test protocol drivers rather than the production browser runtime.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Proven integration** | `TestFiveThousandConnectionWorldFanoutSoak` opens 5,000 authenticated WebSockets on one embedded node, subscribes each to `world`, publishes ten 10-Hz revisions, requires every subscriber's strictly increasing subsequence to terminate at revision 10, rejects a receipt-shaped public publish every tick, and fails any non-world/non-snapshot push. The protocol-aware sniffer has explicit malformed/private-push negative cases. | Preserve the cold soak and include `928126b` plus its transport dependencies in the final reviewed range. Do not turn the current short correctness soak into an unstated production resource-capacity claim. |
| AC2 | **Partial** | `TestActualWebsocketRecoveryReplaysPrivateReceiptsAndLatestWorld` records real epoch/offset positions, disconnects, publishes missed private revisions 2/3 and world revisions 2–20, then explicitly reconnects with recovery and receives ordered private history plus world revision 20 only. The production browser never records those positions, asks for recovery, or reconnects. | Wire one production recovery controller that persists positions, reconnects/resubscribes, consumes recovered publications through the same event path, and falls back to full state when recovery is unavailable. Prove it across a real disconnect with missed receipts and a newest-world control. |
| AC3 | **Cold witness green; designated review pending** | `TestActualTenSecondStalledWorldConsumerConvergesToLatest` holds an actual socket unread for 10.50 s, accepts only bounded live catch-up or disconnect, then reconnects from a known position and requires the one cached publication to contain final revision 93. Refusing every post-baseline world revision fails after the full stall. The production browser separately proves abnormal-close recovery and composed missed-receipt convergence. | Preserve the literal duration, bounded branch, recovery landing and mutation; include `afb5bf0` in the exact Q-003 review. |
| AC4 | **Partial / consumer absent** | An actual slow private socket closes with code 4000 after the independent application message bound fills, and the generic UI can request a full HTTP snapshot after an explicit `resync_required` system envelope. The browser close listener discards the close event's code and emits only `transport_closed`; it neither reconnects nor invokes full sync. The two witnesses therefore cannot compose into the criterion. | Exercise actual overflow through the production runtime, discriminate code 4000, fetch the authoritative committed snapshot, reset both scope cursors, reconnect/resubscribe, and prove the landing revision. |
| AC5 | **Partial / consumer absent** | Gameserver tests enforce broadcast → admission close/in-flight completion → outbox flush → socket close/shutdown, including deadline and broadcast-failure branches. Actual recovery is proven separately. The browser merely sets `draining=true` on the courtesy envelope, ignores `resume_after_ms`, and treats close as permanent offline. No fixture carries an in-flight receipt across drain into browser reconnect/recovery. | Compose the real gameserver and production recovery controller: delay reconnect by the advertised bound, reconnect with saved positions/new credentials if needed, and assert zero receipt loss and no duplicate state mutation. |
| AC6 | **Partial** | Authorization is structurally server-side. Unit coverage proves own/other player, allowed Guild, and denied Match; production composition proves allowed real Cohort/Guild and denied Match. No test attempts a non-member Guild subscription, so the literal `guild:*` and `match:*` negative pair would pass if the Guild resolver accidentally authorized every ID. | Add a real second-account/non-member Guild rejection beside the member control and retain the denied/participant Match pair when the Match owner lands. |

No Transport plan checkbox was changed. Plan item 4 is factually stale rather than unfinished;
item 7 remains open.

## Production recovery trace

The missing consumer can be shown without inference:

1. RFC T4 requires the client to persist each channel's Centrifuge epoch/offset, request recovery,
   full-sync on unrecoverable history or a revision gap, and then resubscribe live.
2. `createBrowserGameUIRuntime.subscribe` creates one WebSocket and sends a connect command with the
   access token captured at subscription time.
3. After the connect reply it sends two subscribe commands containing only `channel`. It neither
   parses subscribe positions nor retains publication offsets.
4. On a socket push it decodes only `push.pub.data`; it discards the publication offset and channel
   recovery metadata.
5. On `close`, regardless of close code or reason, it emits `transport_closed`. No timer, socket
   factory call, recovery request, full sync, or credential rotation follows.
6. `GameUIApp.svelte` maps that message to `offline=true`, clears its local subscription marker,
   and calls the already-closing unsubscribe. No subsequent state transition calls `subscribe`
   again.
7. `server_restarting` sets a draining flag, but its decoded `resume_after_ms` has no consumer.

This also explains why the composed browser handshake is green: it proves initial bootstrap,
snapshot, connect, subscribe, and one presence message, then ends before any recovery boundary.

## Orphaned reconciliation authority

`client/src/transport.ts` defines a sensible `PlayerRevisionCursor`: separate Company/Founder
revisions, same-revision event-ID dedupe, forward-gap detection, and historical delivery without
cursor mutation. Repository-wide import search finds it only in `client/test/transport.test.ts`.
The production Game UI runtime imports the envelope/world decoders but not the cursor.

This is not dead-code tidiness. Durable player delivery is at least once, so a crash after publish
and before outbox acknowledgement may redeliver an event. The server may also dead-letter an event
and signal full sync, and historical compensation intentionally arrives behind the current scope
revision. Without the cursor in the actual consumer:

- duplicate forward events are presented twice;
- a revision jump does not trigger authoritative resync;
- snapshot revisions never initialize/reset the two scope cursors; and
- the reviewed historical-compensation rule exists as a library primitive, not live behavior.

Canonical `docs/transport.md` and RFC T3 describe the cursor as client authority without disclosing
that it is unbound. The repository therefore currently mistakes a tested class for a shipped
consumer.

## Normative, canonical, plan, and review drift

1. D4 and T4 normatively require reconnect/recovery and `resume_after_ms` suppression. The browser
   implementation contradicts them; canonical Transport and Game UI docs describe the server and
   initial subscription but do not expose that recovery is absent.
2. T6 still says production composes Account/Commons plus fail-closed Guild/Match resolvers. The
   later gameserver now composes real Postgres Guild and Cohort membership while Match remains
   deny-closed. The canonical transport doc is current; the active RFC body is not.
3. Plan item 4 says `cmd/gameserver` is blocked on Commons and its note calls Guild settlement
   composition unfinished. Both shipped and were designated-reviewed in later gameserver work.
   The box cannot simply flip because its current text and acceptance ownership must first be
   reconciled with those successor ranges.
4. The Transport log's own final entry records `Review by: pending independent diff review` for
   F2/F3. A later Run Genesis remediation log contains the actual designated Claude verdict. Its
   old range `6141a0f..8578fb1` is recoverable only through the checked-in rewrite map
   (`ae2b3cd..f6d5c1d`). Transport's own record never links that fact, and the latest protocol-aware
   soak (`928126b`) plus production browser consumer (`9a97543` and successors) fall outside that
   2026-08-02 verdict.
5. Historical Transport reviews cite pre-rewrite hashes such as `e2fce6c..2096916`,
   `e2eeadf..e7c1e60`, `c87be53`, `fc97436`, `82ba351`, and `728cd96`; none resolves in current
   history and no same-record mapping/union translates them into a complete current span.

Later successor approvals are valuable evidence for their own bounded changes. They do not form a
declared Transport archival range union, and no designated verdict has reviewed the presently
missing browser recovery as complete because it is not complete.

## Smallest honest closeout order

1. Accept the client-side transport completion under this RFC's existing D4/T4 authority or an
   explicitly accepted Game UI/Account successor: position storage, per-scope cursor binding,
   reconnect/resubscribe, history-unavailable full sync, typed-close handling, drain delay, and
   access-token rotation integration.
2. Add integrated browser/server AC2, AC4, and AC5 witnesses with seeded missing-position,
   duplicate, gap, expired-history, queue-overflow, drain, and credential-expiry failures.
3. AC3's owner-reconciled ten-second witness is green at `afb5bf0`; retain it in the exact Q-003
   review with the production browser's abnormal-close and composed missed-receipt witnesses.
4. Add AC6's absent non-member Guild negative; keep Match's eventual participant positive owned by
   the accepted Match implementation rather than inventing membership here.
5. Ruling author reconciles T6 and the stale plan against the shipped gameserver/Guild successors;
   canonical docs disclose the production client's actual recovery behavior in the same change.
6. Translate all consumed historical verdicts through the checked-in rewrite map, include later
   composition/soak/client ranges, and obtain one tracked cross-party verdict whose cited ranges
   union the complete claimed implementation span.
7. Only then flip the remaining plan items and perform transactional RFC/planning archival.
