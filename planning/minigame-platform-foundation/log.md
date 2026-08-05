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

## 2026-08-05 — immutable session command history and terminal replay

- Appended migrations 00052-00053. Every applied nonterminal or terminal command is recorded under
  `(session_id, seq)` in the same transaction as its resulting session revision, with canonical
  command bytes and a server timestamp. Direct update/delete is rejected; parent-session retention
  may cascade, so account deletion is not bricked by an immutable child trigger.
- Terminal resolution now locks the exact claimed session, replays all committed commands from
  genesis through the frozen engine/version and scaling inputs, compares the pre-terminal state,
  then re-executes and byte-compares the pending terminal snapshot/result before any write.
- Real-Postgres coverage proves command order/revision identity, rollback, direct immutability,
  parent cascade, and a same-version code-drift attack that leaves the claimed session unchanged.
  The normal minigame test/vet targets and root `make test-save-integration` pass.
- This closes C20's persistence/replay slice. Payout/faucet composition remains a separate checked
  item and no production balance data was introduced.

## 2026-08-05 — independent approval of C20 (`2517a51^..2517a51`)

- **Review by:** Darwin
- **Recorded by:** Codex
- **Decision:** approved; no findings.
- The reviewer verified atomic append/head updates, exact contiguous sequence/revision identity,
  replay from stored genesis through frozen engine/scaling identity, byte comparisons before the
  terminal write, same-version drift rollback, and forward-only migration/retention behavior.
- Exact-range diff, minigame test/vet, and real-Postgres minigame integration passed. A shared-DB
  all-package attempt encountered unrelated stale epoch/save state; focused production and repeated
  minigame Postgres reruns passed, and no failure touched the reviewed range.

## 2026-08-05 — remaining Minigame implementation contracts

- C19 still names source concepts without an exact transform grammar; C21 splits one resolution
  across Company payout, Founder rating, and session state under incompatible lock/transaction
  descriptions; C22 depends on the absent Founder attended-time authority; and C23's exact rows
  still contain unnamed nested objects.
- Filed C24-C27 with proposed contracts. The independently approved C20 command-history slice
  remains complete; no production scaling, payout, faucet, or fallback semantics will be invented.

## 2026-08-05 — Founder clock unblocked; remaining wire gaps narrowed

- Founder Attendance landed in `5c3f4c3` with pinned-catalog offline classification, shared Go/TS
  bounds vectors, and a real-Postgres two-order Exit proof. C26 no longer blocks on an absent clock.
- Reapplying C24-C27 to the now-real clock exposed three remaining non-balance gaps: C24 never
  enumerated its claimed exact transform keys/operation union; Founder rating has no persisted or
  replay schema; and the faucet policy/counter omits its resource ID, remainder columns, and
  session-idempotency record.
- Filed C28-C30 with executable proposals. The independently approved session history remains
  intact. Production payout stays blocked rather than inventing a wire or a second attendance
  source.

## 2026-08-05 — C28 scaling grammar and exact resolver

- Implemented the owner-ruled exact-key scaling row and closed source/operation unions. The loader
  rejects unknown keys, duplicate destinations, unknown Founder carry paths, noncanonical integer
  literals, invalid bounds or operands, and every ranked `power` destination.
- Resolution uses exact integer intermediates before clamp and mathematical floor division for
  negative values. Focused tests cover all five source arms, every operation class, negative floor,
  a product beyond `int64` before clamp, and the loader's fail-closed cases.
- The published formula artifact now fingerprints the scaling loader/resolver and states the exact
  grammar, operation order, rounding rule, and Fairness Law. No production catalog row or balance
  literal was introduced; fallback/offline-quality rows remain open in the parent checklist.

## 2026-08-05 — C27 fallback grammar; offline curve gap isolated

- Implemented all three exact fallback arms: solo carries no peer fields; bot freezes a mechanical
  policy identity/version and reduction; NPC partner freezes the corresponding profile
  identity/version and reduction. The strict loader rejects mixed arms, unknown/extra keys,
  malformed identities/versions, and reductions outside the ppm domain.
- Added focused tests for every accepted arm and the discriminating fail-closed cases. The formula
  artifact now fingerprints the loader and publishes the closed arm order and validation rule.
- `DESIGN-GAP:` the ruled offline-quality outer row still does not enumerate the nested
  `grade_curve` row keys, ordering, or duplicate rule. No wire shape or production value was
  improvised; that sub-slice remains explicitly open while the independently useful fallback
  grammar lands.

## 2026-08-05 — C30 payout policy and exact conversion kernel

- Implemented the exact four-key payout row with declared-resource validation and exact-domain
  bounds. Unknown resources, old/extra keys, noninteger values, and invalid ppm fail load.
- Implemented fallback reduction followed by carried-ppm conversion using exact integer
  intermediates. Tests prove operation order, modulo carry across sequential sends, invalid-state
  rejection, and the maximum exact score whose intermediate product exceeds `int64`.
- Published and source-fingerprinted the payout grammar and formula. No cap/window write or
  production faucet is enabled.
- `DESIGN-GAP:` payout composition still needs the certified score-fact selector, configured-cap
  reason key, and exact attended-window persistence row. C30 does not provide those fields, so the
  transaction cannot be made byte-replayable without a narrow owner ruling.

## 2026-08-05 — remaining composition bounce C31-C33

- Filed executable contracts for the offline grade-curve/attended-state wire, explicit certified
  payout score plus cap reason, and one cross-run faucet-window authority with an immutable
  per-session application record.
- The proposals preserve the shipped exact kernels and Founder→Company→session lock order. No
  balance value or production artifact is requested; implementation continues when the owner
  rules the wire ownership.

## 2026-08-05 — exact-key adversarial hardening

- Self-review found that `encoding/json` accepts duplicate object keys even with unknown-field
  rejection. Added one recursive duplicate-key guard shared by scaling, fallback, and payout
  loaders, including nested identity objects; every ambiguity now fails before typed decode.
- Added regression rows for duplicate scaling operations, fallback discriminators/identity keys,
  and payout caps. The formula fingerprint covers the new authority.
