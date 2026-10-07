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
  - [x] C14 repository source: board-record authorization, exact immutable bytes/pinned metadata,
    hash consistency and real-Postgres refusal tests.
  - [x] C14 HTTP handlers/manifest descriptors and a real verified-run download round trip
    (TestComposedGameserverExitVerificationAndBoardIntegration; pinned catalogs supplied from DB).
  - [ ] Complete the third-party loop using public catalogs and the TypeScript verifier;
    same-epoch Go download proof does not close this requirement.
- [x] Generate OpenAPI 3.1 and TypeScript from the registry; add compatibility pins/gate.
  - [x] Generate typed HTTP dispatch and raw/status response associations; bind real Game UI
    bootstrap/state reads (generated-api/runtime tests and real composed journey).
  - [ ] Migrate remaining HTTP callers; register missing routes/contracts and enforce C9's
    raw-fetch lint. Typed dispatch over the current subset is not AC4 completion.
- [ ] Compose the public router, full-registry privacy/conformance integration, canonical docs,
  independent full-range review, and archive.
  - [x] Test caching and shared IP limiting through every currently mounted public operation,
    including changed-content and exhausted-budget cases (A8/C16/C20).
  - [ ] Resolve the undeclared bodiless 304 response arm before claiming full schema/status
    conformance. Runtime caching tests do not resolve this contract gap.

Carried dependency: historical formula serving stays unavailable until a protocol-compliant mint
stores formula bytes in the artifact set; no current-formula fallback is permitted.

## 2026-08-21 corrective lane — exact schema-response bytes

Authorized by accepted A4/A5 and split from rejected Q-002 range `34d04a5..dfaeafe` after
designated finding B (`ba8ca65`).

- [x] Add one validation-only exact-JSON arm to schema response descriptors; construction rejects
  empty, unsorted, duplicate, or schema-invalid literals and registry snapshots deep-clone bytes.
- [x] Bind the four Minigame operations' existing HTTP error mappings to operation/status-specific
  exact bytes without changing any handler output, status, category, detail, generated schema, or
  generated client artifact.
- [x] Demonstrate rejection of a schema-valid wrong category/detail pair, an appended response byte,
  invalid literal registration, and post-construction caller/snapshot mutation.
- [x] Update canonical API Foundation docs; run cold focused/root Go, vet, `api-check`, typecheck,
  and sequential Account/Gameserver Postgres populations.
- [x] Record an exact-range Codex first-filter and hand this API Foundation range plus Q-002's
  test-only remainder to Claude. Do not claim public endpoints, surface completion, or archival.

## 2026-10-07 — existing public board read failure

RP-367 locally repaired within A3/A6/C13; row-iteration regression and affected package tests/vet
pass. The normal composed lane now requires the existing real-Postgres board ranking/page case
to execute, not skip. It passes, but the whole lane is RED at the separate Pitch account-limiter
failure RP-368. Next: diagnose actual request cadence without changing limits/retry semantics;
obtain designated review of this repair. Catalog descriptors, verification readers, generated
client, privacy/acceptance and archival remain incomplete.
