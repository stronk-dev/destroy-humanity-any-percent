# API Foundation implementation log

## 2026-08-03 — accepted-contract reconciliation

C1–C17 are owner-ruled. A1–A8 now name the operation/schema single authority, literal public DTO
families, normalized board query, validate-before-parse HMAC cursor, and operational middleware
semantics. The two historical acceptance-blocker sections remain evidence, while the status and
changelog now truthfully mark implementation unblocked.

Formula fallback remains prohibited: hashes without stored formula artifact bytes return honest
unavailability until the future artifact-growth mint.

## 2026-08-03 — operation schema and cursor foundation

- Added an explicit Go schema DSL over exact object, array, string, bounded integer, boolean, null,
  ref, and oneOf. Named definitions validate at registry construction and the same values validate
  exact runtime response bytes.
- Added the sorted operation registry with method/path/surface/auth/public/schema validation,
  duplicate-route rejection, and status-specific response validation.
- Added validate-before-parse HMAC-SHA256 keyset cursors. Tokens bind operation + normalized filter,
  enforce canonical JSON, accept exactly current/previous 32+ byte deployment keys, reject padding,
  tampering, and cross-query reuse, and parse the key only after MAC verification.
- Added canonical board-variable encoding with explicit-null faction and integer booleans.
- Focused package and mandatory TypeScript gates pass. The `node:fs` corpus import was removed
  after independent review proved the prior typecheck claim false.
- C18/C19 retain exactness instead of weakening the DSL: heterogeneous artifact JSON needs
  per-owner discriminated schemas, and immutable gzip/raw evidence needs a raw response descriptor.
