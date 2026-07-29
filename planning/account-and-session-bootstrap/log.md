# Account & Session Bootstrap — append-only log

## 2026-07-29 — start

- Owner accepted the new root RFC through the ordered batch manifest. HTTP owns intents; WebSocket
  remains streams-only. Imported founders are permanently unranked.
- The handed-off contract names Argon2id but not work factors. This security configuration is pinned
  to current OWASP minimum guidance: 19 MiB memory, 2 iterations, parallelism 1, 16-byte random salt,
  32-byte output. The encoded hash carries algorithm/version/parameters for later rehash upgrades.
- Account and session state is relational; gameplay state continues through the existing owner-aware
  Save Store. The implementation will not place PII or membership claims in JWTs or save payloads.
