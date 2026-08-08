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

`make api-schema` regenerates canonical OpenAPI 3.1 at `docs/generated/api.json` and exact client
DTO/operation metadata at `client/src/api/generated/types.ts`. `make api-check` regenerates and
byte-compares both outputs as part of `make verify-server`. The committed
`docs/generated/api-compat-v1.json` baseline enforces the v1 additive-only law: existing
operations, request unions, statuses, required fields, and bounds cannot narrow or disappear;
responses may add optional fields or widen an enum/union. Updating the compatibility baseline is
an explicit `make api-pin` operation, never an incidental effect of ordinary generation.

Public pagination cursors contain canonical `{filter_sha256,key,op,v}` JSON followed by an
HMAC-SHA256 signature, encoded as unpadded base64url. The codec verifies the signature with the
current or previous deployment key before parsing JSON, binds the cursor to operation and complete
normalized query, and rejects noncanonical or oversized input. Board variables use canonical
exact-key JSON with integer booleans and explicit-null faction.

Public endpoint registration/readers and the generated client's thin HTTP transport remain
pending. The current generated document therefore covers the mounted authenticated Recovery and
minigame surface only; it does not claim the public-read acceptance criteria. Historical formulas
never fall back to current bytes.
