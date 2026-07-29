# Codex batch — 2026-07-29 handoff

All 38 review blockers from the five drafts are now answered, ruled, or explicitly routed. This is
the ordered queue. Implement top-down; each item lists its spec, what changed since your review, and
its acceptance gate. Per-change review convention applies: I review every diff before archival.

## Implement now (dependency order)

1. **Account & Session Bootstrap** — `rfc/account-and-session-bootstrap.md` (NEW).
   The root your reviews named four times. Anonymous-first accounts (recovery-code credential, zero
   PII), JWT access (5 exact claims) + rotating opaque refresh, `New Founder` lifecycle on the
   existing Save-Layer owner seam (`server/save/store.go` OwnerFounder), closed chi surface incl.
   `POST /api/v1/intents` (rules Transport #2: HTTP for intents, WS for streams), anonymous local
   play + import path with `imported` board exclusion. Nothing upstream of this in the queue works
   without it. ACs 1–7 in the RFC.

2. **Prestige & Exits** — `rfc/prestige-and-exits.md`, contracts **P1–P8** appended.
   Your 8 blockers each have an executable answer: P1 closed persisted-state field lists (both
   scopes, save-version bump + corpus fixtures); P2 integer-binary-search cube root (no floating
   cbrt, golden vectors listed); P3 offer state machine at deterministic evaluation sites with the
   save-seeded SplitMix64 stream + `decline_exit_offer` intent; P4 `ApplyExitTransaction` (lock
   order, same-transaction idempotency, run_seq+1 — no new stream); P5 scripted-first trigger
   (attended_ms ≥ 900_000, empty exit_history) per the elective-Exit ruling; P6 server-derived
   timer facts; P7 defers the run log to Leaderboards L1 but must emit `terminal_seq` on
   `run_ended`. Implementable against Phase-0 catalog with fixture content.

3. **Leaderboards & Balance Epochs** — `rfc/leaderboards-and-epochs.md`, contracts **L1–L8** appended.
   L1 `run_log` table (same-transaction insert in the production engine — this touches item 2's
   commit path, so implement 2 and 3 as one arc or sequence 3 right behind it); L2 replay identity
   (`kernel/VERSION` + `catalog_artifacts`, fifth verdict `engine_mismatch`); L3 timers bound to
   Prestige P6; L4 validator = the existing TS kernel (no WASM); L5 epoch tables + atomic mint +
   `FOR SHARE` run pinning; L6 board projection (competition ranking, keyset cursors, world-first
   partial unique index); L7 **ruling: `commons` and `advisor` are two structural variables** —
   Solo = both false; L8 CI-hook extension of the hardened guard (mint-or-accepted-set, cap-lowering
   = always a mint).

4. **WebSocket Transport & Fan-out** — `rfc/websocket-transport-and-fanout.md`, contracts **T1–T6**
   appended. Identity delegated to item 1; T3 closed wire schemas (receipt passes through C1
   unmodified); T4 recovery authority (512 msg / 10 min history, `GET /api/v1/founder/state` as the
   single full-sync of last resort); T5 Phase-0 literal backpressure constants + typed close codes;
   T6 `cmd/gameserver` composition with the `Drain(ctx)` seam AC5 exercises in-process.

5. **Combat (split executed, 4 RFCs)** — implement in order: `combat-data-model.md` (now the shared
   data/arithmetic parent: catalog objects, exact rational arithmetic ×13/10 & ×10/13, SplitMix64
   substreams, fixture-only Trust/Soul seam) → `combat-duel-engine.md` → `combat-lane-engine.md` →
   `combat-bots-and-integration.md`. Pure engines, no I/O; only the bots RFC touches identity or
   storage. Lower priority than 1–4 (nothing else depends on combat); pick it up when the queue
   above is in review, or bounce individual contracts with lettered blockers as usual.

## Blocked — do not implement, do not improvise

- **Commons Onboarding & Governance** — ruled blocked-on-owner in the RFC: needs the
  **faction/incorporation model** and **guild model** RFCs, which I have not drafted yet. They are
  my next drafting targets after this batch is in flight.

## Rulings this batch (so you don't re-derive them)

- First-Exit pacing envelope measures the **first elective Exit**; scripted ~15-min failure is a
  fixed curriculum segment outside it (`design/02 §3.1` amended; Prestige AC8/P5).
- Intent transport is **HTTP** (`POST /api/v1/intents`); WS is streams-only (Account D3 / Transport T2).
- `Assisted` is **two variables** (`commons`, `advisor`), never one bit (Leaderboards L7).
- Cap-lowering is definitionally a balance change: mint required, clamp-on-migration policy in the
  changelog (Leaderboards L8).
- Imported (anonymous-upgrade) founders are permanently excluded from ranked boards (Account D4).

## Standing

- Repo is still local-only; every CI guard is advisory until it's pushed — Marco's action.
- Remaining undrafted Phase-0 contracts: faction/incorporation · guild model · T0–T1 content ·
  doctrine intents (rfc/README.md footer updated).
