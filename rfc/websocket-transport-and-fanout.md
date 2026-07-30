# RFC: WebSocket Transport & Fan-out

- **Status:** implementing
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

## Executable contracts (answering the 2026-07-29 review)

### T1 — Identity (answered by a new owner RFC)

`rfc/account-and-session-bootstrap.md` (drafted 2026-07-29) owns accounts, sessions, JWT claims (exactly `{sub, fid, exp, iat, jti}`), refresh/revocation, and founder binding. This RFC consumes its access token at connect; membership lookups stay server-side at subscribe (D1 unchanged). **Origin policy:** allowlist from deployment config, checked at upgrade; same-origin in dev.

### T2 — Inbound intents (resolved: HTTP, not WS-RPC)

Intents travel `POST /api/v1/intents` (Account RFC D3): body = one Production-C1 envelope verbatim — **no adapter, no second envelope**; response = the receipt JSON. Client timeout 10 s; retry = resubmit the same `intent_id` (idempotency is already the engine's contract); an HTTP failure after commit is healed by the receipt arriving on `player:{fid}` — the channel is the authority, the HTTP response is a convenience copy.

### T3 — Outbound wire schemas (closed, versioned)

The envelope (D3) gains `v: 1`. Closed `kind` payloads, exact-key validated in the TS decoder:
- `receipt`: the production receipt JSON **unmodified** (its schema is C1's; this RFC adds nothing and strips nothing). The shell's narrower internal types are produced by the shell's own mapping (already implemented and reviewed); `internal_invariant` maps to the shell's existing wire result of that name.
- `snapshot`: `{scope: "company"|"world"|"guild"|"cohort", rev, state}` — `state` is the save-layer canonical state for `company`, the published aggregate schemas (Commons/world dials, already generated artifacts) for the rest.
- `event`: `{event_id, kind, rev, payload}` — the event-envelope registry as-is.
- `presence`: `{joined: [id], left: [id], count}` · `system`: `{code, resume_after_ms?}` with closed code set `{server_restarting, history_expired, resync_required}`.
Unknown `kind` ignored (forward-compat); unknown field inside a known kind = decode error (strictness where we have a contract).

### T4 — Recovery authority

Client persists `(channel, centrifuge stream position/epoch)` from the SDK. `player:*` history: **size 512 messages / TTL 10 min** (config, Phase-0 literals). Recovery inside the window replays in order; **outside it centrifuge reports unrecoverable → client receives `system:resync_required` semantics and performs the authoritative full sync: `GET /api/v1/founder/state`** (Account RFC surface; returns latest committed revision + state, same bytes as a `snapshot`), then resubscribes from live. A revision gap detected by the shell (rev N+2 after N) triggers the same path. Full sync is the single recovery of last resort everywhere; nothing else invents catch-up.

### T5 — Backpressure constants (Phase-0 literals, config-validated)

`world` publish 4 Hz; `feed` history 50; `player:*` queue bound 256 messages / 1 MiB; message size cap 64 KiB; subscriptions per connection ≤ 16; connections per account ≤ 3 (oldest closed); drain timeout 15 s. Typed close codes: `4000 queue_overflow`, `4001 auth_expired`, `4002 replaced`, `4003 server_drain`. All loader-validated (positive, caps ≥ minima) like every catalog.

### T6 — Runnable lifecycle (the composed server, named here)

`cmd/gameserver` composes: chi router (Account RFC surface + `/api/v1/intents`) + centrifuge node + production engine + projections, one binary. `/healthz` = process up; `/readyz` = Postgres reachable ∧ catalogs loaded ∧ not draining. Drain seam: `Drain(ctx)` — set not-ready → broadcast `server_restarting` → stop accepting intents (503 typed rejection) → wait in-flight transactions (bounded by drain timeout) → close connections with 4003. AC5 exercises `Drain` directly in-process; deployment infra later just calls the same seam.

## Changelog

- 2026-07-28: created (draft).
- 2026-07-29: removed a disproved review finding: Production C1 and the live parser both require
  `window_ms`; D3 maps the shell's mechanical `windowMs` field normally.
- 2026-07-29: Codex acceptance review recorded six executable-contract gaps; remains draft.
- 2026-07-29: all six answered (T1–T6); identity delegated to the new Account & Session Bootstrap RFC; inbound path ruled HTTP.
- 2026-07-29: accepted by the ordered Codex batch manifest; implementation started after the account, Prestige, and epoch-pinning foundations landed.
