# API Foundation

The API foundation has two implemented mechanical authorities: an explicit Go schema/operation
registry and authenticated keyset cursors.

Schema descriptors cover exact objects, arrays, strings with closed formats/enums, bounded int64
integers, booleans, null, named references, and one-of unions. Registry construction rejects an
invalid descriptor, duplicate operation ID, duplicate method/path, public/auth mismatch, or an
unknown request/response schema. Runtime response fixtures validate against the same descriptor
values that future OpenAPI and TypeScript generation will consume.

Public pagination cursors contain canonical `{filter_sha256,key,op,v}` JSON followed by an
HMAC-SHA256 signature, encoded as unpadded base64url. The codec verifies the signature with the
current or previous deployment key before parsing JSON, binds the cursor to operation and complete
normalized query, and rejects noncanonical or oversized input. Board variables use canonical
exact-key JSON with integer booleans and explicit-null faction.

Public endpoint registration, generated OpenAPI/TypeScript, operational middleware, and readers
remain pending. Historical formulas never fall back to current bytes. Two endpoint-layer design
gaps remain explicit: heterogeneous catalog JSON descriptors and raw immutable verification
response descriptors.
