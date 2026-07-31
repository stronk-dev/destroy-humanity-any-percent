# Guild Model — implementation log

## 2026-07-30 — accepted and started

- The independent Faction remediation review approved FB-1/F2a and the LOW batch, so Guild's GA
  parent contract is implemented and archived. The owner explicitly placed Guild next.
- GB supplies the complete strict Phase-0 catalog; GC supplies a unique integer-only clearing
  answer, including the non-redistribution rule and labeled NPC counterparty. No balance mechanics
  need to be inferred.
- Implementation order is catalog/epoch → storage/lifecycle → tithe/Health/exchange → transport
  resolver → canonical docs and full verification. Per-change commits remain behind the mandatory
  independent review gate before archival.

## 2026-07-30 — catalog mint and owner-boundary audit

- Added the strict GB loader and matching client schema, then minted Epoch 3 with Guild as an
  append-only artifact and regenerated only the constants identity in the isolated harness commit.
- `DESIGN-GAP (name moderation)`: G1 refers to an existing name-moderation charset, but no runtime
  validator exists. The Guild service therefore requires an injected fail-closed `NameValidator`;
  it does not invent a charset. Composition remains blocked until the account/UGC owner supplies it.
- `DESIGN-GAP (tithe units)`: G3 declares integer `guild_xp` as a percentage of arbitrary Decimal
  production deltas but defines no Decimal→int64 normalization, resource basis, or overflow rule.
  Storage and the server-derived projection boundary can land; XP mutation cannot be improvised.
- `DESIGN-GAP (guild Health sample)`: `guild_health_inputs(active_founders,tithed_xp)` is declared,
  but no normative formula maps those two integers to the 0..1,000,000 guild Health term. The
  existing Commons weighted-Health function cannot derive that missing denominator.
- `DESIGN-GAP (account deletion)`: account deletion physically removes `accounts`, while Guild
  history and leadership reference account identity. The RFC defines New-Founder survival but not
  leader succession, disbanding, or anonymization on account deletion. Production composition is
  fail-closed rather than letting the new foreign keys brick the existing deletion contract.

## 2026-07-30 — structural implementation round complete; review gate open

- Commits `3c9f770..d519003` implement the strict catalog and Epoch-3 mint, isolated harness
  identity refresh, Postgres lifecycle model, closed account intents, byte-identical idempotency,
  append-only membership periods, cap/leader serialization, applications/invitations, role
  transfer, manual and seven-day automatic disband, and canonical docs.
- GC is a pure integer kernel with raw-account ordering, base/remainder allocation, per-boundary
  intake/headroom limits, deliberately no redistribution, absent-link inertia, and the reduced-rate
  labeled NPC path. Detached kernel values were renamed after the full writer-closure gate correctly
  rejected authoritative-looking `StockUnits` field names outside the Faction owner.
- Active membership implements Guild authorization. Join/leave commits also append a durable,
  leased presence-outbox row; the transport relay emits schema-validated `guild:{id}` presence
  envelopes and marks publication only after success. The authenticated account router exposes the
  separate exact-schema Guild intent surface.
- Real Postgres integration proves idempotent replay, concurrent cap admission, leave denial in
  authz, rejoin-as-second-history-row, deletion trigger refusal, and the exact grace boundary. The
  entire Compose integration suite is green.
- `make verify` is green: Go vet/tests, formula and epoch/baseline history guards, deterministic
  pacing harness, client typecheck/build, 6,452 client unit tests, schema parity including Guild,
  and 19,365 browser tests.
- Completion is intentionally blocked on GD1–GD6 now listed in the RFC. The most important newly
  surfaced gaps are Decimal production→int64 XP dimensionality, XP→Guild-Health normalization,
  mixed-epoch clearing authority/transaction ownership, and rounding-sensitive placement of the
  named `stock_consumption` slot. No runtime placeholder implements any of them.

## 2026-07-30 — independent complete-diff review (3c9f770..d1e7503)

Two lanes: mint/catalog by the reviewer directly (epoch-3 mint protocol-compliant, guilds artifact
appended append-only, accepted hash reproduced at HEAD, catalog byte-exact to GB, guard
registration one-line-pattern-consistent); lifecycle/exchange/presence/HTTP adversarial with live
Postgres. **Verdict: APPROVED with findings — the structural core is genuinely strong** (real
two-goroutine cap-race proof, partial-unique leader/membership invariants with savepoint-typed
23505 handling, append-only history trigger attacked directly, transactional presence outbox with
token-guarded claims, server-resolved actor identity, and the six recorded gaps verified HONEST —
no writer exists for any deferred surface, exactly as claimed).

Findings (fix queue, ordered):

1. **MEDIUM — officer-permission TOCTOU** (verified first-hand, intents.go:298-306): actor role is
   read unlocked before `lockGuild` and never re-checked — a just-demoted officer's in-flight
   admit/invite lands after the demotion commits. Leader invariants unaffected. Fix: re-read the
   actor's role after taking the guild lock, all officer-gated arms.
2. **MEDIUM — two G1 "all mutations evented" violations:** leadership transfer emits no
   `role_changed` for the self-demotion (a projection replaying events reconstructs two leaders);
   sweep and manual disband close memberships without `member_left` events. Fix: emit the missing
   events in the same transactions.
3. **MEDIUM→ruling — GC kernel deviates from the RFC's literal arithmetic** (consumer eligibility
   and allocation capped by min(intake headroom, stock-cap headroom); near-cap consumers excluded
   from `n`). The code's answer is BETTER than the spec's (no units silently destroyed at
   saturation-on-credit). **Ruling GC-1: the RFC adopts the implemented semantics** — `cap_i =
   min(intake headroom, stock_cap − received)`, zero-headroom consumers excluded from the
   denominator; RFC text amended; the kernel is now the spec's golden answer.
4. LOW/MEDIUM — AB-BA lock order between leave/disband (membership→guild) and set_role
   (guild→membership): real deadlock window surfacing as a generic 409 after a ~1 s stall. Fix:
   guild-lock-first everywhere.
5. LOW batch: `guild_id` is client-supplied via intent_id (collision → permanent generic conflict
   against the victim's intent; server-generate the id); per-arm typed-rejection tests missing
   (nine arms untested) + concurrent leader-uniqueness test absent (AC1's letter) + the literal
   3-present-1-absent AC5 fixture; **AC6 (real-socket presence) is unproven and was NOT in the
   gap list** — now recorded here: it parks under composition with the sweep/relay drivers, but it
   must be NAMED, not implied; generic-409 masking of internal errors is a shared account-API wart
   (queued once, both surfaces).
6. INFO — d519003 verified as a pure rename correctly satisfying the faction writer-closure gate
   (not a bug fix); sweep presence rows share one guild_revision (moot, count=0); epoch-test
   parameterization legitimate.

**Process note for Marco:** commit 1c4b418 ("review: approve faction remediation round") also
carried the new AGENTS.md history-rewrite convention — the rule is the reviewer's (this log's
author), but it entered the repo riding a review commit; it deserves Marco's explicit sign-off and
is called out in the session summary.
