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
