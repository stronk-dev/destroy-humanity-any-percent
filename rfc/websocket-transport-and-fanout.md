# RFC: WebSocket Transport & Fan-out

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/06 §backend/fan-out`, `design/05 §2` (presence & feed), `design/00` law 2 (AI fallback — NPC traffic rides the same channels)
- **Research:** `design/research/tech-stack.md §1` (centrifuge-embedded recommendation, aggregate-then-broadcast, backpressure), `design/research/cicd-deploy.md §5` (the drain handshake)
- **Depends on:** Production Engine and Client Shell (implemented — receipts/snapshots and the abstract consuming stream are the boundaries)
- **Planning:** `planning/websocket-transport-and-fanout/` (once implementing)

## Summary

The realtime layer: `coder/websocket` transport with **`centrifugal/centrifuge` embedded as a library** (channels, presence, history/recovery, JWT — not the standalone server), the channel taxonomy, the message envelope, and the fan-out discipline that keeps an idle MMO's chatter from ever scaling per-click. Single node to ~10k CCU; the Redis broker is a named follow-up, not a day-one component.

## Specification

### D1 — Channel taxonomy

| Channel | Scope | Content | Recovery |
|---|---|---|---|
| `player:{founder_id}` | private | intent receipts, authoritative snapshots, personal events | **history-backed** (centrifuge recovery) — a reconnect replays missed receipts |
| `world` | public | coalesced global counters, milestone bars, Planet/Health dials, epoch beats | **latest-only** — recovery is the next snapshot; history is meaningless for a gauge |
| `feed` | public | the live feed (`05 §2`), dispatches | shallow history (last N=50) |
| `guild:{id}` / `cohort:{id}` | membership-gated | guild/cohort events, tithe/standing updates | shallow history |
| `match:{id}` | participants | minigame/lane frames | match-local; a dropped player rejoins from the match snapshot |

JWT-authorized subscriptions; membership checks server-side at subscribe (never claims in the token beyond identity).

### D2 — Fan-out discipline (normative, the tech-stack rules made contract)

- **Aggregate-then-broadcast:** global counters coalesce into **one `world` snapshot at 4–10 Hz** (rate is config). Nothing publishes per-click, per-purchase, or per-player to a public channel — *ever*. The feed is the only exception and it publishes **curated events**, already rate-shaped at source.
- **Per-connection buffered queue with drop-stale:** for gauge-type messages (`world`), a slow consumer's queue keeps only the newest snapshot per channel; for receipt-type messages (`player:*`), the queue is lossless up to a bound, beyond which the connection is closed and the client re-syncs through recovery — **receipts are never silently dropped.**
- Presence: join/leave events + a periodic aggregate count; full roster only on subscribe (and only where the surface needs it — guild/cohort).

### D3 — Message envelope

```json
{"ch":"player:…","kind":"receipt|snapshot|event|presence|system",
 "rev":42,"constants_hash":"sha256:…","ts":"server","payload":{…}}
```
Payload decimals are canonical strings (a wire payload is a wire payload); `rev` ties `player:*` messages to save revisions so the client-shell reconciliation (its D2) keys on the same ordering the server committed. Unknown `kind` is ignored by clients (forward compatibility), never errored.

### D4 — The drain handshake (owned here; the deploy RFC consumes it)

`{"kind":"system","payload":{"code":"server_restarting","resume_after_ms":N}}` broadcast on all channels at drain start (CI RFC D5's sequence). Clients render it diegetically ("scheduled maintenance" in-fiction), suppress reconnect attempts for `resume_after_ms`, then reconnect with recovery. **The client reconnect path is the same one as any network drop** — the drain adds courtesy, not a second code path.

### D5 — Limits

Per-connection subscribe caps (config), per-channel publish authz (only server actors publish; clients publish nothing — intents travel the request path, not pub/sub), message size caps, and connection-count metrics feeding the CCU telemetry the impact modifier reads.

## Acceptance criteria

1. A soak test at simulated 5k connections holds one node: `world` at 10 Hz, zero per-click messages observed on any public channel (asserted by a wire sniffer in the test).
2. Kill a subscriber mid-burst: on reconnect, `player:*` recovery replays every missed receipt in order; `world` shows only the latest snapshot.
3. Drop-stale property: a consumer stalled for 10 s receives exactly one `world` snapshot on resume.
4. A receipt-queue overflow closes the connection with a typed close code; the client's re-sync lands on the committed revision.
5. Drain: broadcast → in-flight commits flush → connections close within the bounded timeout → a reconnecting client resumes with zero lost receipts (the CI RFC's deferred deploy AC, executable here).
6. Non-participants cannot subscribe to `guild:*`/`match:*` (authz test).

## Open questions

- Redis broker threshold (named follow-up; single-node until telemetry says otherwise).
- Feed curation rules live with `design/05 §2` content work, not here.

## DESIGN-GAPs blocking acceptance (Codex review, 2026-07-29)

1. **Identity and session authority:** account/session bootstrap does not exist. Define connection
   authentication, token issue/refresh/revocation, Origin policy, Founder identity binding, and
   membership lookup before JWT-authorized private channels can be implemented.
2. **Inbound intent protocol:** choose and specify the request path named in D2 (HTTP or WebSocket
   RPC), its exact versioned envelope, `intent_id`/revision mapping, timeout and retry behavior, and
   the adapter into the production engine's closed intent union.
3. **Exact outbound payloads:** D3's `payload:{…}` is not a wire contract. Define closed, versioned
   schemas for snapshots, applied/rejected receipts, events, presence, and system messages, including
   the mapping from production's receipt JSON into the Client Shell's deliberately narrower types and
   the `internal_invariant` wire result already assigned here.
4. **Recovery authority:** define the Centrifuge stream position/epoch persisted by the client, the
   bounded history size and expiry, the gap/expired-history response, and the authoritative full-sync
   operation. A revision alone cannot recover messages from an evicted in-memory history.
5. **Backpressure constants:** queue bounds, message/subscription limits, typed close codes, publish
   cadence, and drain timeout must be catalog/config fields with literal Phase-0 values and loader
   validation; “up to a bound” is not executable.
6. **Runnable lifecycle:** no account HTTP surface or composed game server exists yet. Name the server
   bootstrap, health/readiness behavior, in-flight transaction drain boundary, and test seam that AC5
   exercises rather than assuming deployment infrastructure owns them.

## Changelog

- 2026-07-28: created (draft).
- 2026-07-29: removed a disproved review finding: Production C1 and the live parser both require
  `window_ms`; D3 maps the shell's mechanical `windowMs` field normally.
- 2026-07-29: Codex acceptance review recorded six executable-contract gaps; remains draft.
