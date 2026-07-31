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

## 2026-07-31 — review remediation and GD1–GD6 implementation round

- Banked the independent verdict/spec amendments in `9a1df1b`, then closed the authority TOCTOU
  and AB-BA window with one guild-first lock order and post-lock role validation. Leadership
  transfer, manual disband, and sweep now emit every membership/role mutation. Guild IDs are
  server-generated rather than aliasing idempotency IDs (`727676d`).
- GD1 is concrete: NFKC/lowercase/whitespace normalization, closed ASCII grammar, committed
  substring denylist plus additive deployment policy, and typed `name_policy` receipts.
- Minted Balance Epoch 4 (`3a00c24`) for the two owner literals and the declared
  `guild.stock_consumption` faction-slot source. No hardcap or nonzero multiplier changed.
- Save v11 owns `guild_tithe_carry_ppm`, `guild_boundary_seq`, and the consumed-window count.
  Production computes exact integer tier-progress deltas; the Guild hook carries division
  remainder and emits strict tithe events only when the server-resolved membership contribution is
  present. The idempotent projector saturates XP, records trailing-window activity, and exposes the
  population-invariant Guild Health term to Commons.
- Clearing boundaries persist one deterministic result per member and sequence; company-local
  application enforces a monotonic watermark and owns both debit and credit without cross-company
  locks. The stock-consumption provider publishes the current zero-rate factor through the declared
  economy source. RA will record these server-resolved slices rather than recomputing them.
- Account deletion now invokes a transactional Guild participant: membership history is closed and
  anonymized, leadership succession is deterministic, and the empty-guild case disbands. A real
  Postgres test covers tithe projection idempotency, clearing slices, deletion succession, and FK
  anonymization; the existing lifecycle/concurrency suite remains green.
- Focused Go, schema, client (6,452 tests), and full Postgres integration suites are green. The RFC
  remains `implementing` until this complete delta receives the required independent review. AC6's
  real socket/scheduler composition stays named under the gameserver round, not implied complete.

## 2026-07-31 — implementation adversarial pass

- Fixed two cross-component bugs before handoff: Guild Health now excludes founders who are no
  longer active members, and the clearing writer locks the guild then requires its authoritative
  active-account set to equal the supplied snapshot set. Old pinned economy catalogs also omit the
  new membership contribution cleanly instead of receiving an undeclared source.
- Found **GD5a (blocking composition, not the completed kernels):** `boundary_seq` is per-guild but
  save v11 records no guild identity. A company moving from boundary 10,000 in one guild to boundary
  5 in another would reject all future slices. Recorded in the RFC with the two legitimate owner
  choices; no improvised reset/global ordering was added. RA cannot resolve this ambiguity for us.
- Account deletion now explicitly fails closed when an active Guild membership exists but the
  transactional Guild participant was not attached; the database trigger independently rejects
  an active-history nulling attempt.

## 2026-07-31 — independent review: guild runtime round (727676d..f9d1b9a)

Mint lane (reviewer, direct): epoch-4 mint protocol-compliant, hash reproduced at HEAD, changelog
in the mint commit. GD5a ruled in the RFC (watermark = (guild_id, boundary_seq) pair, forward-only
reset). Runtime lane: adversarial with live Postgres, F1 re-verified first-hand in the trigger SQL.

**Verdict: NOT archivable yet — one HIGH blocks. Everything else is approved with rulings.**

All five prior-review fixes verified closed with evidence (TOCTOU re-reads post-lock in every
gated arm; self-demotion evented; per-member `member_left` on sweep AND manual disband; guild-first
order unified; server-generated ids with replay identity from the stored receipt). GD1/GD2/GD5/
GD6/save-v11 verified to contract, including: the shipped-evaluator tithe delta with carry
round-trip, event_id-keyed idempotent XP projection with no double-count path, GC-1 kernel
byte-matching the ruling, the set-equality snapshot hardening, the undeclared-source skip that
would otherwise have bricked pre-guild-epoch members, and honest fail-closed GD5a recording.

Findings and rulings:

1. **HIGH (blocks archival) — account deletion is bricked for any account with a CLOSED guild
   membership row** (verified first-hand: 00028's trigger raises on any update to a closed row —
   `OLD.left_at IS NOT NULL` is an unconditional arm — so the FK cascade's `account_id → NULL`
   aborts the deletion; proven live by the review). The shipped test only covers an active-row
   deletion. **Fix contract: the trigger admits exactly one closed-row transition —
   `account_id → NULL` with every other column unchanged**; regression fixture = join, leave,
   rejoin elsewhere, leave, delete account (two closed rows + one active).
2. **MEDIUM→ruling GD3-1 — the SPEC stands: "active" = ≥1 accrual EVALUATION in window,** not
   ≥1 whole XP produced (implemented). The implemented narrowing inflates H_guild by dropping
   parked/low-progress members from the denominator — player-favorable drift of a Commons health
   input, exactly what the Compact laws forbid. The projector writes activity on every member
   evaluation regardless of XP; fixture: 3 evaluating members, 1 producing → denominator 3.
3. **LOW→ruling GD1-1 — denylist matching also strips separators**: match against the normalized
   form AND the form with `[ _-]` removed (`a-dmin` falls). Denylist entries validate under a
   laxer rule (≥2 chars, charset `[a-z0-9]` only) so short terms are deniable.
4. LOW batch: deletion emits no presence to `guild:{id}` (add `insertPresence` in the deletion
   participant, same tx); residual cross-guild set_role deadlock (reject `targetGuild != guildID`
   from an unlocked read BEFORE locking the target row); clearing replay-after-membership-change
   errors instead of no-oping (order the seq≤last check first) and committed-seq retries with
   different snapshots should hard-fail, not silently no-op (compare snapshot hash); orphaned
   applications/invitations on disband/sweep (close them in the same tx); the F7 test debt now
   two rounds old — per-arm rejection suite, concurrent leader-uniqueness, lock-order regression —
   lands with the F1 fix round, no further deferral.
