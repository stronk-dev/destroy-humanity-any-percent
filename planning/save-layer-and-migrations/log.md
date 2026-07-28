# Save Layer & Migrations — Running Log

## 2026-07-28 — Start

- Owner directed local work to continue without any git push.
- Reviewed the corrected save draft, implemented economy boundary, and current primary
  documentation for Goose, pgx, and GitHub Postgres service containers.
- Bound tooling to Goose 3.27.1 embedded sequential SQL migrations and pgx/v5 5.10 through
  `database/sql`. Goose runs SQL migrations transactionally by default; Postgres 16 is the only
  integration database.
- No ORM, SQL mock, SQLite substitute, standalone migration artifact, or runtime migration file
  path will be introduced.
- The unrelated CI RFC remains implementing solely because its hosted timing gate requires a
  future push; it does not block local save-layer work.

## 2026-07-28 — Persistence foundation

- Added pinned Goose and pgx modules plus an embedded transactional version-1 migration for
  owner-aware streams and revisions. Database constraints enforce owner/scope pairings, positive
  revisions/versions, object state, and canonical SHA-256 hash syntax.
- Added `OpenPostgres` and `Migrate`; migration files are compiled into the package.
- Added `economy.RestoreLedger`, the validation-only constructor required by RFC-0002. It requires
  exactly one canonical value for every in-scope catalog resource and rechecks minima/hardcaps.
- Added canonical version-1 state encoding/decoding, future-version refusal, exact-artifact SHA-256
  identity, strict unknown-field rejection, and poisoned-value tests.
- Focused economy/save tests pass. Repository stream operations and real-Postgres integration are
  the next implementation slice.
