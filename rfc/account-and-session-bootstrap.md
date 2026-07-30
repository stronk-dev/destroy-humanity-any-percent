# RFC: Account & Session Bootstrap

- **Status:** implementing
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/06 §backend` (JWT accounts; anonymous play with local-only saves + upgrade path), `design/research/compliance-2026-refresh.md` (data minimisation; no third parties), `design/research/save-layer` archive (owner-aware streams; `New Founder` follow-up)
- **Depends on:** Save Layer (implemented — this RFC supplies the missing owner: who a `founder` UUID belongs to)
- **Unblocks:** Transport #1, Leaderboards (identity), Prestige (session), Commons-onboarding #1 — the shared root of the 2026-07-29 blocker lists
- **Planning:** `planning/account-and-session-bootstrap/` (once implementing)

## Summary

The keystone the five drafts all named as missing: accounts, sessions, tokens, and the Founder lifecycle (including the long-deferred free `New Founder`). Deliberately minimal: **email-less anonymous-first accounts**, one token scheme, one HTTP surface, no third-party anything.

## Specification

### D1 — Accounts are anonymous-first

- `POST /api/v1/account` creates an account: `{account_id (UUIDv7), created_at, recovery_code}`. **No email, no password, no PII at creation** — the recovery code (128-bit, base32, shown once) *is* the credential. This is the compliance dossier's data-minimisation posture made structural: an account row contains an ID, a timestamp, and an Argon2id hash of the recovery code. Nothing else.
- Optional later attach: `POST /api/v1/account/email` binds an email for recovery (verification mail is the only outbound mail; provider config is deployment data). Attaching is never required for any feature.
- **Account ⟷ Founder:** an account owns founders (`account_founders(account_id, founder_id, created_at, archived_at)`). Exactly one **active** founder per account (partial unique index on `archived_at IS NULL`). `POST /api/v1/founder` = the **`New Founder`** action: archives the active founder (never deletes — Save Layer streams are already archive-only) and creates a fresh one. **Free, unlimited, no cooldown, one are-you-sure client-side** (`design/10 §5`; the affordance server-authority removed, restored).

### D2 — Sessions and tokens

- `POST /api/v1/session {account_id, recovery_code}` → `{access_token, refresh_token}`.
- **Access token: JWT, HS256, 15-minute TTL**, claims exactly `{sub: account_id, fid: active_founder_id, exp, iat, jti}` — nothing else in the token; membership/authorization is always a server-side lookup at use (Transport D1's rule, honored here). Signing key from deployment secret; **key rotation = two accepted keys (current+previous), rotated by ops, no in-band negotiation.**
- **Refresh token: opaque 256-bit, stored hashed** in `sessions(token_hash, account_id, created_at, expires_at (30 d), revoked_at)`; single-use rotation (refresh consumes and reissues; reuse of a consumed token revokes the session family — standard theft detection).
- `DELETE /api/v1/session` revokes. Founder switch (`New Founder`) revokes nothing but changes `fid` at next refresh.
- **A-D2a (review ruling, 2026-07-30):** family state is a row: `session_families(family_id PK, account_id, revoked_at)`, created at issue. Every refresh and every revocation locks the family row `FOR UPDATE` before validating; refresh rejects a revoked family. Per-row revocation flags alone are NOT sufficient (a concurrent rotation's insert escapes a bulk UPDATE's snapshot — demonstrated).
- **A-D1a (review ruling, 2026-07-30):** recovery-code verification uses the STORED Argon2 parameters subject to a floor (m ≥ 19 MiB, t ≥ 2, p ≥ 1) and rehashes to current parameters on successful verify; a parameter bump must never lock accounts out. Session creation dummy-verifies a constant hash when the account is missing (timing-oracle fix).
- Centrifuge connects with the access token; channel subscriptions re-check membership server-side (already normative in Transport D1).

### D3 — The HTTP surface (chi, versioned, closed)

`/api/v1/`: `POST account` · `POST account/email` · `POST session` · `POST session/refresh` · `DELETE session` · `POST founder` · `GET founder` (active founder's public profile: id, created_at, display fields) · `GET founder/state` (authoritative full sync — added by the Transport T4 ruling, 2026-07-29) · `POST founder/import` (D4). **Closed set; additions by RFC.** Router-edge 404/405 use the typed rejection shape; failed-auth requests on authenticated routes consume an IP bucket; limiter state is evicted (TTL/LRU) and deployment config names trusted proxy hops. All errors are the typed-rejection shape from Production C1 (`{category, detail}`); rate limits per-IP on the unauthenticated three (token bucket, config), per-account elsewhere.
**Intent submission (`POST /api/v1/intents`) mounts here as well** — the "request path" Transport D2 references: body = one Production-C1 intent envelope, response = the receipt; auth = access token; the transport RFC owes only the *streaming* side, not the request side. This closes Transport blocker #2's "choose HTTP or WS-RPC": **HTTP for intents, WebSocket for streams.**

### D4 — Anonymous local play and the upgrade path

The client may run **fully offline-anonymous** (local saves, no account) per `design/06`. "Upgrade" = create account + `POST /api/v1/founder/import` with the local save; the server validates it through the Save Layer's restore path exactly as any save, then **marks the founder `imported: true` — imported founders are permanently excluded from ranked boards** (their history is client-authored and unverifiable; Leaderboards consumes this flag). Everything else works.

**A-D4a (review ruling, 2026-07-30):** the request's `constants_hash`/`version` select only the
migration catalog. The server re-encodes the restored state under its CURRENT constants hash,
resets `RunSeq` to 1 with a fresh `RunStartedAt`, and writes through the standard store path so the
existing run-1 pin holds by construction (run identity is server-owned; the founder is ranked-
excluded regardless). Acceptance fixture: a locally-prestiged (`run_seq=3`), old-hash save imports
and then plays — the review demonstrated both axes brick the stream without this ruling.

### D5 — Compliance hooks (from the 2026 refresh, made structural)

Account deletion (`DELETE /api/v1/account`): archives all founders — **A-D5a (review ruling, 2026-07-30): `account_founders.account_id` is nullable with ON DELETE SET NULL; the delete transaction stamps `archived_at`, so founder rows and the `imported` flag survive anonymized** (a cascade delete breaks board projection and destroys the exclusion marker) — deletes the email binding and sessions, retains the anonymized save streams (they contain no PII by construction — an account row is the *only* PII surface, and only if email was attached). The accountable-person line, MAU counting (DSA Art. 24(3)), and the age-neutral posture are ops/docs items riding this RFC's data model; the children's-risk-assessment inputs cite D1's no-PII-at-creation design.

## Acceptance criteria

1. Account create → session → intent submit → receipt round-trips against the real stack (Postgres + production engine) with no fields beyond the schemas above (exact-key validated both directions).
2. Refresh rotation: reuse of a consumed refresh token revokes the family; a revoked session's access token fails at channel subscribe within TTL expiry.
3. `New Founder`: archives-not-deletes, unlimited, free; the archived founder's streams remain readable; the new founder starts from catalog initials.
4. JWT claims are exactly the five named; a token with extra claims is rejected.
5. Import path: a valid local save imports through the standard restore path; `imported` founders are excluded from ranked-board queries (fixture with the Leaderboards projection once it exists; until then, the flag round-trips).
6. Account deletion leaves no PII row while save streams survive anonymized.
7. Rate limits enforced with typed `rate_limited` rejections on the unauthenticated endpoints.

## Open questions

- Email provider config shape: deployment data, not blocking (the endpoint can 501 until configured).
- Display-name policy plugs into the existing name-moderation rules (`compliance.md`) when profiles gain names; not blocking.

## Changelog

- 2026-07-29: created (draft) — the shared root of the five drafts' blocker lists.
- 2026-07-30: independent review (see planning log) — rulings A-D1a, A-D2a, A-D4a, A-D5a; D3 list gains `founder/state` (transport T4) and router-edge/limiter contracts.
- 2026-07-29: accepted and implementation assigned through
  `planning/codex-batch-2026-07-29.md`.
- 2026-07-30: review remediation implements A-D1a/A-D5a and the D3 limiter contract: transactional
  stored-parameter Argon2 upgrade plus dummy verification, anonymized Founder retention, trusted
  proxy depth, bounded idle/LRU buckets, and failed-auth IP limiting.
