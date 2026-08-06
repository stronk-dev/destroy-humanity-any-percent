# RFC: Deployment Foundation

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/06 §deploy` (single Go binary, Caddy, docker-compose, Postgres 16), `design/07` (roadmap — shipping cadence = in-fiction events)
- **Research:** `cicd-deploy.md` (the banked deploy research — hosted GitHub Actions not Komodo, the drain handshake, `§5` the drain sequence the transport RFC already consumes)
- **Depends on:** Gameserver Composition (implemented — the composed binary with the `Drain(ctx)` seam is what deploys); Transport drain handshake (implemented — the `server_restarting` broadcast); CI Baseline (implementing)
- **Owner ruling honored:** breadth-first — the deploy MECHANICS + the push that makes every guard structural; Marco-gated for the actual push and any secret/host provisioning.
- **Planning:** `planning/deployment-foundation/` (once implementing)

## Summary

The last foundation, and the one only Marco can complete. Everything is built to deploy — the
single binary, the drain seam, the migrate-at-startup, the CI guards — but nothing deploys, and
the repo is local-only, so every guard we built (epoch history, KV-1, balance harness, the review
ledger) is ADVISORY. This RFC specifies the deploy pipeline and the push that makes it all
structural. **The push is the single highest-leverage action remaining in the project.**

## Specification

### DP1 — The deploy artifact

The single `cmd/gameserver` binary + its embedded catalogs (go:embed prod, disk dev) + the client
build (static, served by Caddy or the binary) — one container, per the tech-stack decision. Caddy
for auto-TLS + the COOP/COEP headers (if the Stockfish NNUE minigame path needs them). Postgres 16
as the only stateful dependency. docker-compose for the full stack; the compose file is the
deployment contract.

### DP2 — The release sequence (the drain handshake, consumed)

Deploy = the transport drain handshake the RFC already ships: broadcast `server_restarting` →
drain in-flight (the bounded `Drain(ctx)` seam) → migrate-at-startup (single-process, the invariant
the migration RFCs assumed) → epoch seed sync before readiness (L5c) → ready. Rolling deploy is a
named follow-up (single-node to ~10k CCU per the tech-stack finding — the Redis broker and
multi-node are later); Phase-0 deploy is stop-drain-start, which the drain handshake makes a
diegetic "scheduled maintenance" (design/00's no-real-FOMO: downtime is a story beat, streaks are
no-ops across it — already law).

### DP3 — CI/CD (the banked decision)

Hosted GitHub Actions (NOT Komodo — the research verdict), the full `make verify` gate the repo
already has, plus: the balance-harness/epoch guards run on `fetch-depth: 0` (the F1-adjacent CI
lesson — already fixed), the KV-1 kernel guard, the API drift gate, the copy legal/provenance
lints, the relevance gate (when it lands). **The push makes these enforce** — a guard that runs
only on a laptop enforces nothing.

### DP4 — The push (Marco-gated, the structural moment)

The repo is local-only; four epochs of balance history, every immutability guard, and the entire
review ledger are advisory until pushed. **This is the one action in the whole project that
requires the owner** (external operation — secrets, host, the git remote). Recommended BEFORE the
first content mint (T0–T1) so content epochs are born structural. The push is not a code change;
it's the moment the machine's guarantees become real. Everything else in this RFC is preparation
for it.

### DP5 — Secrets & config

JWT signing keys (the two-key rotation the account RFC specced), the cursor MAC secret (API C8),
Postgres credentials, the trusted-proxy config, the deployment allowlist Origin — all deployment
data, injected by the host, never in the repo (the compliance data-minimisation posture). A
config-validation startup check fails closed on any missing secret (readiness gate).

## Acceptance criteria

1. `docker-compose up` boots the full stack (Postgres + migrated gameserver + Caddy-served
   client) to readiness on a clean host; the compose file is the reproducible contract.
2. Release sequence: a deploy drains cleanly (the `server_restarting` → bounded drain → migrate →
   sync → ready sequence, integration-proven against the composed binary — AC5 of transport,
   finally in its deployment home).
3. CI/CD: the full guard suite runs in hosted Actions on push; a seeded guard violation fails the
   pipeline (the guards enforce, not advise).
4. Secrets: startup fails closed on any missing secret; no secret in the repo (grep-proven).
5. **The push** (Marco): the repo is public/remote; the guards are structural; recorded as the
   phase transition it is.

## Open questions

- Host choice + the Redis-broker threshold (single-node until telemetry says otherwise — the
  tech-stack finding; a scaling follow-up).
- Rolling/multi-node deploy — explicitly a later RFC; Phase-0 is stop-drain-start.

## Changelog

- 2026-08-03: created (draft) — the deploy pipeline + the push that makes every guard structural;
  the one foundation only the owner can complete.
