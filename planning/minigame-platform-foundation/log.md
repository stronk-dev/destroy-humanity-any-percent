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

## 2026-08-04 — independent rejection of first session/tenant landing (`151cddc^..151cddc`)

- **Review by:** Darwin
- **Recorded by:** Codex
- **Decision:** not approved; two HIGHs and four MEDIUMs require remediation.
- HIGH: a rejected/divergent tenant command had no no-op claim release, stranding the row for the
  five-minute recovery lease. HIGH: exported `ResolveTx` accepted any canonical object and could
  permanently store a result that never passed the tenant's result schema.
- MEDIUM band: schema-reference strings had no invoked validators; the resolve function could not
  prove Company lock/identity; database constraints did not bind Founder→stream or hash→run and
  allowed claim-arm state mutation; the JSON canonicalizer preserved `1|1.0|1e0` aliases.
- What held: concurrent/stale claim CAS, rollback atomicity, frozen named genesis fields, terminal
  update immutability, defensive clones, closed errors/modes, and the absence of invented balance
  data. The reviewer inspected committed blobs and excluded the in-progress remediation.

## 2026-08-04 — session/tenant remediation

- Added token-owned no-op release and a composed Service path: stale/rejected commands release
  without a fake revision; applied commands advance once; terminal commands retain an opaque
  certification for the payout transaction. Migration 00050 was already exercised locally, so
  subsequent constraint hardening landed forward-only as 00051 rather than rewriting it.
- Resolution is now a Service method over an externally unforgeable certification. It revalidates
  typed result shape and the tenant-owned result schema, recomputes exact bytes, locks the certified
  Founder-owned Company/run/hash, and predicates the session write on the same complete identity.
- Tenant descriptors now have executable command/snapshot/result validators invoked around every
  pure engine call. The canonical-wrong-schema snapshot and result cases fail, rather than relying
  on schema-reference labels or whitespace rejection.
- The database now has composite Founder→stream and hash→run constraints plus arm-specific trigger
  checks preventing claim-time state/result mutation. Canonical JSON rejects every non-safe-integer
  numeric spelling, including decimal/exponent aliases, before JSONB normalization.
- Real-Postgres coverage exercises release after stale/rejected commands, identity constraints,
  claim-arm mutation, forged result bytes, cross-Company resolution, and full fixture
  create→play→certify→resolve. Independent follow-up review remains required.

## 2026-08-04 — independent review of remediation (`47bee53^..47bee53`)

- **Review by:** Darwin
- **Recorded by:** Codex
- **Decision:** not approved; the original two HIGHs and four MEDIUMs are closed, but one new HIGH
  remains.
- A session froze `engine_version`, but `Play` selected tenant code using only `engine_ref`. After a
  deployment replaced v1 with v2 under the same reference, an old v1 session could execute v2 code
  while retaining a false v1 identity. The required remediation is exact-pair dispatch and
  regressions proving both nonterminal and terminal v1 commands defer unchanged under a v2-only
  registry.
- The reviewer independently passed the normal minigame test, vet, integration, and diff-check
  targets at exact commit `47bee53`.

## 2026-08-04 — exact engine-version remediation

- Tenant creation, play, and terminal-result validation now bind the frozen
  `(engine_ref, engine_version)` pair. An unavailable historical version returns the typed
  `ErrTenantVersion`; it is never replaced by whatever implementation currently owns the same
  reference.
- The real-Postgres service fixture starts under v1, swaps to a v2-only registry, and proves both a
  nonterminal and terminal command fail without changing revision, state, status, or claim token.
  The original v1 service then continues the same session, proving the rejection is a clean defer
  rather than a poison state.
- `make test-go GO_PACKAGES='./minigame'`, `make vet GO_PACKAGES='./minigame'`, and the normal root
  `make test-save-integration` pass. Independent follow-up review remains required.

## 2026-08-04 — independent approval of exact-version remediation (`a5c16c5^..a5c16c5`)

- **Review by:** Darwin
- **Recorded by:** Codex
- **Decision:** approved; no findings.
- Exact-pair dispatch rejects a v1 session against a v2-only registry before tenant execution.
  Both nonterminal and terminal regressions prove the row remains active at revision 1 with its
  original state and no claim token; continuing through the original v1 service proves the defer
  path does not poison the session.
- The reviewer independently passed the minigame test, vet, Postgres integration, and exact-range
  diff-check targets at commit `a5c16c5`.

## 2026-08-04 — remaining implementation-contract review

- The approved session/tenant slice reaches the last fully literal boundary. Applying C15–C18 to
  the shipped save, replay, and production APIs exposed five DESIGN-GAPs rather than balance-value
  gaps: no exact scaling formula grammar, no immutable command history for C11 replay, no internal
  production transition envelope, no single attended-window/counter definition, and incomplete
  fallback/offline-quality row shapes.
- Filed C19–C23 with proposed contracts in the RFC. No production artifact, payout side write,
  replay schema, or bot/offline mechanic will be improvised while those contracts are unresolved.
- The independently approved Postgres session lifecycle and pure tenant registry remain usable
  foundations; this bounce does not reopen them.

## 2026-08-05 — owner rulings C19-C23

- Owner accepted the closed scaling-source union, append-only session command log, internal
  resolve boundary, cross-run Founder-attended faucet cursor, and fallback/offline-quality row
  families. C22 explicitly corrects the earlier run-start origin: quota never resets on Exit.
- Reconciled that correction at C16's normative decision site. Implementation resumes with the
  command log and replay proof, which require no production balance literals.
