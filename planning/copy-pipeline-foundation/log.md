# Copy Pipeline Foundation log

## 2026-08-03 — implementation opened

- Owner assigned the accepted Copy Pipeline Foundation as the first item in the ordered foundation queue.
- The working tree contains unrelated owner/agent edits; this job will stage only Copy Pipeline files and exact RFC/index hunks.
- Verification uses repository-root `make` targets only. No custom cache/environment wrapper or package-directory test command is part of this job.
- The owner accepted C1-C10. The RFC still contained stale C1-C7/blocking language at implementation start; reconciliation is the first change so code has one authoritative contract.

## 2026-08-03 — source, runtime, and safety gates implemented

- Reconciled C8-C10 and the stale blocker language before code changes.
- Added 12 production copy rows: every currently referenced category, faction-incorporation,
  catalog hardcap, and code-owned reason key resolves without deriving display text from a key.
- Added strict machine registries for catalog references, code-owned references, and research
  provenance. The sole literal statistic (`100%`) resolves to a verified claim with a unique
  research anchor and HTTPS source.
- Added a deterministic generator for the client catalog, generated key/param types, independent
  `copy_hash`, and byte-stable orphan report.
- Added the framework-independent client loader/resolver with exact param/type validation, literal
  braces, era fallback, canonical-Decimal validation, and production invariant fallback.
- Added completeness, narrow-statistic, known-name, provenance, append-only denylist/provenance,
  schema-field, and generated-drift checks. Fixtures discriminate missing keys, statistic
  provenance, known-name punctuation bypass, word-boundary near misses, protected-row mutation,
  and protected-term removal.
- Root-only verification: `make copy-generate` green; `make copy-check` green; `make typecheck`
  green; `make test-client` green (6,506 passed, 3 skipped; 1 test file skipped).

## 2026-08-03 — self-review hardening and canonical docs

- Replaced the initially checked-in code-key list with an explicit producer-site registry and a
  generated code-reference manifest. Generation requires each key to occur exactly once in its
  declared Go producer; the pipeline does not infer references by property-name grep.
- Moved text bounds and statistic detector tokens/year range into strict declared configuration;
  the generated runtime constants and build-time lint now share that authority.
- Added browser-standard SHA-256 verification for copy bytes and a mutation-discriminating test.
- Published `docs/copy-pipeline.md`, including the safety gates' deliberately limited claims.
- Re-ran normal root targets after hardening: `make copy-check`, `make typecheck`, and
  `make test-client` green (6,507 passed, 3 skipped; 1 test file skipped).
