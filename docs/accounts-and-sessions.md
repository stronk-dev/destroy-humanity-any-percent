# Accounts and Sessions

The server supports anonymous-first accounts without requiring email, a password, or any third
party identity provider. Creating an account returns a UUIDv7 account ID and a 128-bit lowercase
base32 recovery code exactly once. The account row stores only the ID, creation time, and an
Argon2id hash of that code. Optional email attachment is reserved by the API but returns the typed
`not_configured` response until deployment mail configuration is specified.

## Founder ownership

An account owns a history of Founder identities and has exactly one active Founder. Account
creation atomically creates the first Founder plus catalog-initialized company and Founder save
streams. `POST /api/v1/founder` is the free, unlimited New Founder operation: it archives the old
identity and all its active save streams, then creates fresh streams. Archives remain readable.
Existing access tokens remain valid until their normal expiry; authenticated operations resolve
the account's current active Founder instead of trusting the token's cached Founder claim.

Offline-anonymous saves can be submitted once to the initial company stream through
`POST /api/v1/founder/import`. The payload runs through the normal save migration and restoration
path. Its submitted version and constants hash select only that migration input: the server resets
the imported company to run 1 at the canonical import instant, validates and writes it through the
normal save-store path under the current constants hash, and permanently marks the Founder as
imported in the same transaction.
Leaderboard projections must exclude that relational flag because imported history was authored
by a client.

Deleting an account first archives every Founder identity and stream, then deletes the account row.
Foreign-key deletion nulls `account_founders.account_id` instead of cascading those rows: Founder
identity, archive time, and the permanent `imported` exclusion marker survive anonymized. Cascades
still remove optional email, sessions, session families, and access-token records. The retained
save/Founder history has UUID owners but no account linkage or PII.

## Credentials and sessions

Recovery codes are hashed with Argon2id using 19 MiB memory, two iterations, one lane, a random
16-byte salt, and a 32-byte result. Parameters and the algorithm version are encoded with the hash
so a later credential upgrade can distinguish old records. Login verifies the stored parameters
when they meet those security floors; a successful login with non-current parameters rehashes the
credential to current settings inside the same transaction. A missing account performs the same
Argon2id work against a constant dummy hash, preventing account-existence timing from bypassing the
KDF. Recovery input is case/outer-whitespace normalized before validation.

Successful recovery authentication issues:

- a 15-minute HS256 access token whose claims are exactly `sub`, `fid`, `exp`, `iat`, and `jti`;
- a random 256-bit opaque refresh token, stored only as a SHA-256 hash and expiring after 30 days.

The verifier accepts one current signing key and one previous key for operations-managed rotation.
Every issued access token is also represented in the database, so revocation is effective before
JWT expiry. Refresh tokens rotate once. Reusing a consumed token revokes every refresh and access
token in its family in the same transaction. A `session_families` row is the serialization point:
every rotation and revocation locks it before validating a token, so a concurrent rotation cannot
escape replay detection or logout. Session deletion applies the same family revocation.

## HTTP boundary

The chi router owns the versioned `/api/v1` surface:

- `POST /account`, `POST /session`, and `POST /session/refresh` are IP-rate-limited;
- authenticated account, session, Founder, import, state, and intent endpoints are
  account-rate-limited; failed authentication also consumes the caller's IP bucket;
- request decoders reject unknown keys, trailing JSON values, empty required bodies, and bodies
  over 64 KiB under the Phase-0 configuration;
- errors use `{category, detail}` and rate-limit failures use the `rate_limited` category;
- router-level 404/405 responses use that same typed shape, and the one-time recovery-code response
  carries `Cache-Control: no-store`;
- `POST /intents` resolves the active company stream and calls the authoritative Production
  service. The client never chooses a stream ID.

The in-memory Phase-0 token buckets resist clock regression by refusing to mint tokens when time
moves backwards. Their key map is a bounded LRU; entries idle for one full-refill interval are
evicted. `TrustedProxyHops` is explicit deployment configuration: zero ignores forwarded headers,
while a positive value selects the client address at that exact trusted depth from
`X-Forwarded-For`. Malformed/short chains fall back to the socket peer. Deployment may replace
storage without changing the HTTP contract.

## Verification

`go test ./account` covers credential encoding, exact JWT claims, expiry and signing-key rotation,
UUIDv7 shape, proxy extraction, failed-auth limiting, bounded limiter eviction, and limiter clock
regression. With `TEST_DATABASE_URL` set, it additionally exercises
the complete account → session → real Production intent path, refresh replay revocation, New
Founder archival, import, deletion/anonymization, and rate limiting against Postgres.
The integration suite also forces a legitimate rotation to race a replay and proves that no
descendant refresh or access token remains live. The account path also proves stored-parameter
credential upgrade and anonymized Founder/import-marker retention after deletion.
