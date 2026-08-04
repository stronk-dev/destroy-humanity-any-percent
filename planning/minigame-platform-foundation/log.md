# Minigame Platform Foundation implementation log

## 2026-08-04 — acceptance reconciliation and start

- Owner rulings C1–C18 were reconciled into MP1, MP4, MP5, and the acceptance criteria before
  implementation. One stale AC4 phrase found during the implementation read was corrected from
  client-intent language to the ruled server-authored `resolve_minigame_session` transition.
- Implementation starts at the dependency floor: the Postgres session/claim boundary, followed by
  the tenant registry. No production payout, fallback, or offline-quality number will be invented.

## 2026-08-04 — session and tenant foundation

- Appended migration 00049. Session rows accept only `solo|async_snapshot`, freeze run/catalog/
  engine/scaling/genesis identity, enforce active→claimed→active|resolved revision transitions,
  and make resolved rows update-immutable. `live_pvp` has no accepted storage value.
- The repository locks Founder before session for play, uses database UUID claims with the shipped
  five-minute recovery lease, rejects stale tokens, and exposes a transaction-owned resolved write
  for the later Company→session payout path. A rollback test proves that result/status cannot leak
  across the transaction boundary.
- The pure tenant registry freezes each descriptor and accepts only canonical object envelopes,
  complete exact-integer scaling maps, declared modes/errors, and the closed result shape. It never
  exposes an economy-delta output. A deterministic fixture tenant proves create/play/resolve and
  fail-closed schema/error/output behavior.
- `make test-go GO_PACKAGES='./minigame'`, `make vet GO_PACKAGES='./minigame'`, and the normal root
  `make test-save-integration` pass. The Postgres suite includes concurrent claim, expired-lease
  replacement, token loss, immutable genesis, resolve rollback, and terminal immutability.
- Canonical docs describe only these shipped boundaries. Production catalog rows, faucet math,
  payout composition, combat adapter, and the epoch mint remain explicitly open.
