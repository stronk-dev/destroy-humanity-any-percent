# API Foundation

The API foundation has three implemented mechanical authorities: an explicit Go schema/operation
registry, generated API contracts, and authenticated keyset cursors.

Schema descriptors cover exact objects, arrays, strings with closed formats/enums, bounded int64
integers, booleans, null, named references, and one-of unions. Registry construction rejects an
invalid descriptor, duplicate operation ID, duplicate method/path, public/auth mismatch, or an
unknown request/response schema. Runtime response fixtures validate against the same descriptor
values consumed by runtime mounting, OpenAPI generation, TypeScript generation, and the committed
v1 compatibility pin. Authenticated Soul Recovery and minigame routes mount exclusively from this
registry; missing, extra, unsorted, or nil runtime bindings fail during router construction.

Operations may declare exact scalar query parameters (string, bounded integer, or boolean). They
are byte-sorted by name and never shadow a path parameter. The reserved `cursor` parameter is a
non-required string, and it is present exactly when the operation declares a cursor key.
`Registry.ParseQuery` decodes only declared parameters. Each parameter may appear at most once,
integers are strict base-10 (no sign, padding, or leading zero), and every value is checked by the
same descriptor validator as response bodies. A failure returns `InvalidQueryError` naming the
parameter; undeclared parameters are ignored.

Query parameters are generated as OpenAPI `in: query` entries, and TypeScript gets
`queryParameters` plus an optional-aware `query` type. Query-free operations generate
byte-identical output. The compatibility pin records the query set only for operations that have
one, and rejects any change to an existing operation's query set. That is stricter than C2 (which
would allow a new optional request field), but never looser.

A schema response may also declare a sorted immutable set of exact JSON wire bytes. This is a
runtime-validation narrowing layered over the generated DTO, not a generated-client or OpenAPI
shape change. Minigame error responses use it to bind each operation/status to the shipped
category/detail pairs and encoder newline, so a schema-valid cross-product or appended byte is not
accepted as contract evidence.

`make api-schema` regenerates canonical OpenAPI 3.1 at `docs/generated/api.json` and exact client
DTO/operation metadata at `client/src/api/generated/types.ts`. `make api-check` regenerates and
byte-compares both outputs as part of `make verify-server`. The committed
`docs/generated/api-compat-v1.json` baseline enforces the v1 additive-only law: existing
operations, request unions, statuses, required fields, and bounds cannot narrow or disappear;
responses may add optional fields or widen an enum/union. Updating the compatibility baseline is
an explicit `make api-pin` operation, never an incidental effect of ordinary generation.
Schema names follow the current-plus-legacy convention: an unversioned name denotes the current
shape, while a retained historical shape uses an explicit `V<n>` suffix. A compatibility-pin
refresh must cite its authorizing ruling and be recorded in the owning planning log in the same
change; an otherwise valid widening is not permission for a silent re-baseline.

GU-C26 authorized the Game UI schema v3 compatibility-pin baseline. Accepted Garage Player
Surfaces GS0.1 (2026-09-25) authorizes the v4 re-baseline. The unversioned `GameUISnapshot` is the
current v4 shape. `GameUISnapshotV1`, `GameUISnapshotV2` and `GameUISnapshotV3` retain exact
stored-bootstrap receipt validation, and the live Founder-state operation returns only v4.

Public pagination cursors contain canonical `{filter_sha256,key,op,v}` JSON followed by an
HMAC-SHA256 signature, encoded as unpadded base64url. The codec verifies the signature with the
current or previous deployment key before parsing JSON, binds the cursor to operation and complete
normalized query, and rejects noncanonical or oversized input. Board variables use canonical
exact-key JSON with integer booleans and explicit-null faction.

## Public reads

`server/publicread` owns the unauthenticated `/api/public/v1/` registry. `account.APIErrorSchema()`
is the single `APIError` definition that both registries reference. `publicapi.MergeRegistries`
unions the private and public registries into one generation authority. A schema name may be
shared only if the definition is byte-identical, and operation IDs must not collide. As a result,
`docs/generated/api.json`, the TypeScript module and the compatibility pin cover both surfaces,
while each surface still mounts from its own registry.

`GET /api/public/v1/epochs` (`list_public_epochs`) returns the C12 `PublicEpochPage`:

- `items` are newest-first. Each has `accepted_hashes` (byte-sorted in Go, an empty list when
  there are none), the exact UTF-8 `changelog_markdown` read from the staged `changelog_ref`, and
  UTC millisecond `started_at`/`ended_at` (`ended_at` is explicit `null` while open).
- `next_cursor` is a C15 keyset cursor on `epoch_id`, bound to the normalized filter
  `{"limit":N}`. `limit` defaults to 50 and is bounded 1..100.
- **Rejections:** a malformed limit returns exactly `400 {"category":"invalid","detail":"limit"}`.
  A malformed, tampered or filter-mismatched cursor returns `400 invalid/cursor`.
- **Failures:** repository or changelog failures, including a missing or non-UTF-8 staged
  changelog, return `500 internal_invariant/public_api`.
- **Caching and validation:** successful bytes are validated against the registry row. They are
  then served through the shared cache/limiter runtime with the `catalogs_epochs` class
  (`public,max-age=3600`, strong ETag, and a 304 that spends no rate token).
- **Page source:** `leaderboard.Repository.PublicEpochPage`, which reads the page from a single SQL
  statement.

`GET /api/public/v1/boards/{category}` (`list_public_board`) returns the C12 `BoardPage`
`{category_id, epoch_id, items, mandate_level, next_cursor, ranking_kind, variables}`.

- **Query:** C13's normalized query. `variables` is the base64url canonical JSON
  `{advisor,commons,faction,glitched}` (sorted keys, 0/1 flags, explicit-null faction). `epoch`,
  `mandate` (0..20) and `variables` are required; `limit` (1..100, default 50) and `cursor` are
  optional.
- **Ranking kind:** resolved from the pinned categories artifact of every stored constants hash
  accepted into the epoch. Timed categories (`rta`/`attended`) are `time_ms` and the untimed
  valuation category is `magnitude`. `count` is declared for the union, but no category produces
  it yet. Catalogs that disagree, or a stored catalog that cannot load, are internal invariants.
- **Items:** `{founder_id, key, rank, run_id, verified_at, world_first}`, where `key` is the closed
  union `{kind:"time_ms"|"count",value}` / `{kind:"magnitude",exponent,quantized_mantissa}`.
  `rank` is the competition rank over the whole board.
- **Cursor:** the keyset cursor (`{key|exponent+quantized_mantissa, run_id}`) is MAC-bound to the
  complete normalized filter (category, variables, epoch, mandate, limit). A cursor whose arm
  doesn't match the ranking kind is rejected.
- **Paging:** the reader fetches `limit+1` rows, so a page that is exactly full carries no cursor.
- **Rejections:**
  - `400 invalid/{cursor,epoch,limit,mandate,variables}`;
  - `404 unknown_id/{category,epoch}`;
  - `429 rate_limited/ip`;
  - `500 internal_invariant/public_api`.
- **Caching:** the `boards` class (`public,max-age=60`).

`publicread.NewRouter` composes the surface. It loads the strict policy, resolves the named cursor
secrets (`CursorSecretResolver`, see `docs/gameserver.md`), builds the request-ID runtime and
limiter, and mounts every public registry operation (and only those) through `Registry.Mount`.
Unknown public paths and non-GET methods return `404 unknown_id/route` and never fall through to
the account router. The gameserver mounts it at `/api/public/v1/`, and composition fails closed
without a valid policy or cursor pair. The composed-server Postgres witness checks all of this:
the served epoch page, request-ID echo, cache headers, a 304 on a matching ETag, the unknown-route
404, and fail-closed composition. The composed Game UI lane also fetches the page through the Vite
proxy.

The catalogs, verification and registry readers, the thin generated-client transport, and the full public privacy enumeration all
remain open. The C18 catalog union waits for every artifact owner's exact descriptor, and
historical formulas never fall back to current bytes.
