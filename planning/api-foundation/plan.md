# API Foundation implementation plan

RFC: `rfc/api-foundation.md`

- [x] Reconcile C1–C17 into the active A1–A8 contract.
- [x] Implement the closed schema-descriptor DSL and operation registry foundation.
- [x] Implement authenticated keyset cursors and normalized board variables.
- [x] Implement strict operational API config and shared request/cache/limiter middleware.
  - [x] Resolve C20's limiter/proxy/key-ID/request-ID literals.
- [ ] Register public DTOs/readers and raw verification evidence endpoints.
  - [ ] Resolve C18's artifact descriptor.
  - [x] Resolve C19's raw response arm.
- [x] Generate OpenAPI 3.1 and TypeScript from the registry; add compatibility pins/gate.
- [ ] Compose the public router, full-registry privacy/conformance integration, canonical docs,
  independent full-range review, and archive.

Carried dependency: historical formula serving stays unavailable until a protocol-compliant mint
stores formula bytes in the artifact set; no current-formula fallback is permitted.
