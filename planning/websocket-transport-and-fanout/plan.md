# WebSocket Transport & Fan-out — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/websocket-transport-and-fanout.md`
- **Started:** 2026-07-29

1. [x] Add strict Phase-0 transport policy and closed Go/TypeScript wire envelopes.
2. [x] Embed Centrifuge with the implemented channel authorization, history, and backpressure policy.
3. [x] Map committed production receipts/events into one transaction-owned private-player outbox;
   each scope is revision-ordered and snapshot recovery uses the authoritative full-state endpoint.
4. [ ] Compose `cmd/gameserver`, readiness, HTTP intents, and bounded drain lifecycle. (lifecycle
   complete; binary blocked on the Commons owner resolver recorded in the log). This item owns
   wiring the real `guild.Service.PendingSettlements`; account integration uses an explicitly
   empty constructor fixture and is not production composition.
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
are restored. The exact `bfd9b65..afb5bf0` range is ready for designated cross-party review.
