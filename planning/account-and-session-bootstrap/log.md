# Account & Session Bootstrap — append-only log

## 2026-07-29 — start

- Owner accepted the new root RFC through the ordered batch manifest. HTTP owns intents; WebSocket
  remains streams-only. Imported founders are permanently unranked.
- The handed-off contract names Argon2id but not work factors. This security configuration is pinned
  to current OWASP minimum guidance: 19 MiB memory, 2 iterations, parallelism 1, 16-byte random salt,
  32-byte output. The encoded hash carries algorithm/version/parameters for later rehash upgrades.
- Account and session state is relational; gameplay state continues through the existing owner-aware
  Save Store. The implementation will not place PII or membership claims in JWTs or save payloads.

## 2026-07-29 — implementation

- Added migration 00010 with a deliberately three-column `accounts` table, optional email side
  table, one-active-Founder ownership history, refresh-token families, and revocable access-token
  records.
- Added 128-bit recovery codes with Argon2id storage, exact-five-claim HS256 JWTs with current and
  previous key acceptance, 15-minute access TTL, 30-day opaque refresh TTL, single-use rotation,
  and family-wide replay revocation.
- Added atomic account initialization, free New Founder archival/replacement, one-time local-save
  import through `save.RestoreState`, imported-Founder marking, and anonymizing account deletion.
- Added the closed chi `/api/v1` surface. It strictly decodes request keys, applies IP/account token
  buckets, resolves active ownership on every request, and mounts authoritative Production intent
  handling without accepting a client-supplied stream ID.
- Unit tests are green. The real-Postgres integration passed account creation, recovery login,
  exact response decoding, a real manual Production intent, refresh replay revocation, archived
  Founder readability, import marking, deletion with save retention, and typed rate limiting.
- Implementation review found that legacy imported saves need an authoritative migration baseline;
  imports now use canonical server import time instead of epoch zero. The current-version integration
  path remains unchanged.
- Full `make verify` passed with Postgres: Go vet and all server packages, generated-formula drift,
  both harness gates, TypeScript/Svelte checks, production client build, 6,412 client tests, schema
  and package-boundary verification, and 19,245 browser tests.
- The RFC remains `implementing` until the required independent diff review is recorded. Forward
  batch work may build on the committed surface without pretending self-review satisfies that gate.

## 2026-07-30 — independent review: account & session bootstrap (bdcc9a1) — the review I owed

Adversarial two-lens review; three HIGHs reproduced against live Postgres by the review harness and
re-verified against source by the reviewer. **Verdict: the credential/token core is genuinely
strong (see verified list in the review record: alg-pinned exact-claims JWT, two-key rotation,
crypto/rand + Argon2id at OWASP minimums, constant-time compares, sequential single-use rotation,
one-active-founder under lock, unforgeable server-set `imported`, zero-PII schema asserted against
information_schema, fail-closed origin/limiter). The session-family model and the import path are
NOT acceptable. Rulings below are now normative in the RFC.**

Findings (fix queue, ordered):

1. **HIGH — family revocation is not atomic; a concurrent refresh escapes it** (`store.go:492-498`).
   READ COMMITTED: the bulk family UPDATE's snapshot cannot see a session row inserted by a
   concurrently-committing rotation, and `RefreshSession` never validates family state — so replaying
   a consumed token at the moment the legit client rotates leaves a live, unrevoked refresh chain;
   theft detection and logout both defeated (reproduced: post-revocation token minted a fresh pair).
   **Ruling A-D2a:** add `session_families(family_id PK, account_id, revoked_at)`; issue creates the
   row; every refresh and every revocation takes `FOR UPDATE` on the family row first; refresh
   rejects when `revoked_at` is set. Serializes rotation against revocation completely.
2. **HIGH — import bricks the company stream** (reproduced twice): (a) a locally-prestiged save
   (`run_seq>1`) imports fine and then every intent 409s forever (no pin for that seq); (b) a
   client-supplied older `constants_hash` writes a revision that mismatches the run-1 pin
   (`run identity mismatch` forever). The direct `INSERT INTO save_revisions` bypasses the store
   path, contradicting D4's own text. **Ruling A-D4a:** the request's `constants_hash`/`version`
   select only the migration catalog; the server re-encodes the restored state under
   `repository.constantsHash`, resets `RunSeq` to 1 with a fresh `RunStartedAt` (imported founders
   are ranked-excluded, so run identity is server-owned), and writes through the standard store path
   so the existing pin holds by construction. AC: import-then-intent round-trip with a run_seq=3,
   old-hash fixture.
3. **MEDIUM — account deletion cascade destroys founder rows and the `imported` flag**
   (`00010_accounts.sql:15` ON DELETE CASCADE), which also turns any queued verified-run projection
   for that founder into a permanent hard error (`boards.go:73-77`). RFC D5 says "archives".
   **Ruling A-D5a:** `account_founders.account_id` becomes nullable with ON DELETE SET NULL; the
   delete transaction stamps `archived_at`; founder rows and `imported` survive anonymized.
4. **MEDIUM — Argon2 parameter bump would lock every account out**: `verifyRecoveryCode` hard-pins
   current constants (`credential.go:66-68`) and no rehash-on-verify exists, contradicting the
   planning log's own claim. **Ruling A-D1a:** verify with the STORED parameters subject to a floor
   (m ≥ 19 MiB, t ≥ 2, p ≥ 1); on successful verify with non-current parameters, rehash to current
   in the same transaction.
5. **MEDIUM — account-existence timing oracle on session creation** (measured 72×: 266 µs miss vs
   19.3 ms hit): dummy-verify a constant hash on the miss path.
6. **MEDIUM — per-IP limiter degrades to one global bucket behind any proxy** (`RemoteAddr` only;
   correct that XFF is untrusted, but deployment needs a trusted-proxy hop count in config) and its
   bucket map is unbounded (no eviction; IPv6 /64 growth). Fix: config `trusted_proxies`, LRU/TTL
   eviction, and apply an IP bucket to failed-auth requests on authenticated routes (currently
   unthrottled).
7. LOW — unknown paths/methods return chi plaintext, not the typed shape (set NotFound/
   MethodNotAllowed handlers). LOW — no session/access-token GC (~2.9k rows/user/month; add prune
   with the existing prune-job pattern). LOW — recovery-code input should normalize case/whitespace
   before validation, and the create response needs `Cache-Control: no-store`. LOW —
   `ErrAccountNotFound` from NewFounder maps to 500 instead of a typed rejection.
8. OBSERVATIONS — stale `fid` authorizes the archived founder's channel ≤15 min (RFC-sanctioned;
   transport may optionally re-resolve on subscribe); unauthenticated account creation is unbounded
   storage amplification (quota/reaper is a named follow-up); the surface is honestly unmounted
   (composition is the gameserver work already queued).

**Record correction:** the review harness flagged `GET /api/v1/founder/state` as a closed-surface
violation; it is NOT — the 2026-07-29 transport T4 ruling authorized it as the full-sync endpoint.
The account RFC's D3 list is amended to include it; the bookkeeping gap was mine.
