# WebSocket Transport & Fan-out — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/websocket-transport-and-fanout.md`
- **Started:** 2026-07-29

2026-10-15 current-run cohort correction (RP-651): subscribe-time permission now requires the
projected membership to match the current non-archived Company head's run and member flag.
The old real HTTP/Postgres/socket witness admits a new non-member subscription after Exit;
the expanded existing Exit witness also checks legitimate pre-Exit access, stable assignment,
re-sign and leave. The owning log records final execution and the range after `54f352f6`
needing Claude review. No governance, Health/retention or established-subscription eviction
policy changes; complete Transport acceptance/archival remain open.

2026-10-15 composition reconciliation: the old Commons-resolver/binary blocker is resolved.
`gameserver.New` wires the real Commons projector and Guild settlement service; the current
built-client aggregate at `6201bef8` executes the persisted socket/clearing/readiness/drain
test twice successfully. Item4 records implementation presence, not complete Transport review
or archival. The owning log retains the exact witness and remaining review obligations.

Current known-payload conformance correction (2026-10-13, RP-561): Go validates exact
DTO-owned fields/nullability and safe wire integers; the client requires actual string
enums instead of coercing arrays. The shared174-case corpus, connected publisher/history
refusal, healthy delivery and affected checks pass. The owning log records both old-source
failures and the full range after `b8eb4c58`, pending Claude review. Opaque owner payloads,
wire version and recovery policy are unchanged; no whole-feature or release promotion.

Current bounded correction (2026-10-13, RP-560): policy loading validates exact original
JSON members against the existing DTO tags before struct decoding. Case aliases and raw/
escaped duplicate members now refuse; canonical escaping and the declared limits/origins
are unchanged. All48 old-source failures and corrected cold/race/consumer checks are in
the owning log. The range after `95d390fc` needs Claude review; no recovery/authentication
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
