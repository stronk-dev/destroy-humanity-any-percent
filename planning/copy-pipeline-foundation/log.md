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

## 2026-08-03 — full local gate

- `make verify` passed at `665bd61`: Go vet; all server packages; formula drift; Commons and full
  balance harness; TypeScript/Svelte checks; production client build; 6,507 client tests; shell,
  kernel-history, combat, copy, and schema guards; and 19,530 browser assertions.
- No custom environment, cache wrapper, or package-directory command was used. This is the normal
  repository-root verification surface required by `AGENTS.md`.

## 2026-08-03 — independent review by Darwin (`261eafa..b0cedab`)

Decision: not approved for archival. Seven findings were demonstrated:

- HIGH: clean-checkout failure because two denylist citations existed only as untracked owner
  research files, contaminating the recorded local green gate.
- HIGH: arbitrary valid catalog bytes were mislabeled with the generated application `COPY_HASH`.
- HIGH: suffix-symbol currencies such as `12€` and `12$` bypassed statistic provenance.
- MEDIUM: production fallback could omit its invariant sink.
- MEDIUM: denylist terms were append-only but their legal citations could be retargeted.
- MEDIUM: code-site verification counted comment/dead string literals rather than Go syntax.
- MEDIUM: the plain-text grammar rejected links/tags but still admitted other Markdown constructs.

The RFC remains implementing. The complete authored verdict was delivered by Darwin against the
full implementation range; remediation follows in one independently re-reviewed batch.

## 2026-08-03 — review remediation implemented

- Retargeted the same append-only denylist terms to tracked legal matrices; all research/provenance
  dependencies must now be Git-tracked. The history guard protects the full term+citation pair and
  permits a citation correction only when the cited file was absent from the parent commit—the
  exact invalid-introduction case—then seals the corrected citation.
- Removed `copyHash` from arbitrary loaded catalogs. Byte identity is explicit through
  `hashCopyCatalog`/`verifyCopyCatalogHash`; a generated deployment manifest now pins the current
  recomputed `constants_hash` and `copy_hash` together.
- Fixed bidirectional currency adjacency and added suffix-symbol/near-miss fixtures.
- Made the invariant reporter mandatory in production at both the type and runtime boundary.
- Replaced raw-literal site counting with a Go AST verifier bound to source file, function, JSON
  field, and key; comment/wrong-field fixtures fail.
- Closed the source/runtime plain-text grammar over Markdown emphasis, code, headings, blockquotes,
  lists, rules, links, tags, and HTML comments, with runtime fixtures.
- Normal root checks after the batch: `make copy-generate`, `make copy-check`, `make typecheck`, and
  `make test-client` green (6,507 passed, 3 skipped; 1 test file skipped).

## 2026-08-03 — remediation full gate

- `make verify` passed at `44b4369`: all server tests including the new Go AST and deployment
  identity packages; formula and harness gates; production client build; 6,507 client tests;
  governance/schema checks; and 19,530 browser assertions.
- The remediation requires a new independent verdict. The original not-approved verdict remains
  authoritative for `261eafa..b0cedab` and is not relabeled or overwritten.

## 2026-08-03 — independent remediation review by Darwin (`44b4369^..2f7c9bd`)

Decision: not approved for archival. Five original findings were verified closed. Three remaining
contract gaps were demonstrated:

- the invalid-citation repair exception checked only immediate-parent absence, so deleting a once-
  valid source in one commit could reopen citation retargeting in the next;
- the AST gate rejected comments/wrong fields but still counted an unused matching map or a map in
  `if false`;
- valid CommonMark `1)` lists and four-space indented code still passed the claimed plain-text
  grammar.

The second authored rejection remains in the append-only record. A second remediation batch is
required before another archive verdict.

## 2026-08-03 — second remediation implemented

- Citation repair now requires that the invalid path never existed anywhere in reachable history,
  with live-history fixtures for both the valid tracked path and the originally uncommitted path.
- The Go AST gate now accepts only a matching field/key inside a reachable direct
  `json.Marshal(map literal)` call in the declared function; a compile-time-false serialized map
  plus a real wrong-field payload is the negative fixture.
- The shared build/runtime plain-text grammar now rejects `1)` lists, four-space code blocks,
  reference-link/table delimiters, and backslash escapes in addition to the earlier Markdown/HTML
  cases.
- Normal root targets `make copy-check`, `make typecheck`, `make test-client`, and `make test-go`
  pass after the second remediation (6,507 client assertions; every server package green).

## 2026-08-03 — second-remediation full gate

- `make copy-check` passed from the repository root after commit `6d0b309`, including the live Git
  history fixtures, Go AST producer-site verification, and deployment manifest identity check.
- `make verify` passed from the repository root at `6d0b309`: Go vet and all server packages;
  generated-formula drift and balance harness; TypeScript/Svelte checks; the production client
  build; 6,507 client assertions; governance/schema guards; and 19,530 browser assertions.
- The second remediation now requires an independent verdict covering `6d0b309` and the preceding
  rejected ranges. Neither earlier rejection is relabeled or treated as approval.

## 2026-08-03 — independent second-remediation review

- **Review by:** Darwin
- **Recorded by:** Codex
- **New range:** `6d0b309^..637a9eb`
- **Combined implementation/remediation span:** `261eafa..637a9eb`
- **Decision:** not approved for archival.

The citation-history finding is closed: the correction exception now searches every commit
reachable from the parent, and Darwin's delete-then-retarget repository probe failed as required.
Two MEDIUM findings remain:

- the producer-site AST proof still accepts discarded marshal output and scans constant-dead or
  never-invoked nested code, so it does not yet prove the registered payload reaches the producer's
  authoritative database call;
- the shared text rule still accepts five-space CommonMark code, no-space blockquotes, empty ATX
  headings, HTML declarations, and XML processing instructions.

Darwin independently passed `make copy-check`, the focused normal root targets, and `make verify`
(6,507 client and 19,530 browser assertions). Green gates do not override the demonstrated false
assurances. A third remediation and independent verdict are required.

## 2026-08-03 — third remediation implemented

- Replaced general recursive AST scanning with one deliberately narrow authoritative producer
  form: a top-level marshal assignment to a non-discarded variable, followed without reassignment
  by exactly one database sink in a standard checked-error `if` initializer. Discarded output,
  nested function literals, constant-dead blocks, unchecked sinks, and shadowed payloads are
  discriminating negative fixtures.
- Closed the shared text grammar over arbitrary four-or-more-space indentation, no-space
  blockquotes, empty headings, HTML declarations, and XML processing instructions. The same new
  cases run through both the build validator and the TypeScript runtime loader.
- Focused normal root checks pass: `make copy-check`,
  `make test-go GO_PACKAGES='./cmd/verify-copy-sites ./cmd/gen-content-manifest'`, and
  `make test-client` (6,507 assertions).

## 2026-08-03 — third-remediation full gate

- `make verify` passed from the repository root at `6c1dfaf`: Go vet and every server package;
  formula and harness gates; TypeScript/Svelte checks; production client build; 6,507 client
  assertions; history/boundary/copy/schema guards; and 19,530 browser assertions.
- A new independent verdict must cover `6c1dfaf` plus this gate record and union with every earlier
  implementation/remediation range before archival.

## 2026-08-03 — independent third-remediation review

- **Review by:** Darwin
- **Recorded by:** Codex
- **New range:** `6c1dfaf^..b3efdce`
- **Full contiguous implementation/remediation union:** `261eafa..b3efdce`
- **Decision:** not approved for archival.

The prior named markup cases and the narrow AST fixtures were closed, but two MEDIUM findings
remained. A syntactically checked dummy database call could satisfy the AST proof while a later
authoritative event insert carried different bytes; indexed mutation, `copy`, and fake receivers
were further counterexamples. The shared text grammar also admitted CommonMark's two-trailing-space
hard line break. Darwin passed the focused root targets and `make verify`; the false assurance still
blocked archival.

## 2026-08-03 — fourth remediation: structural producer authority

- Withdrew the unsound general data-flow claim and deleted the bespoke AST command. C4 requires a
  declared generated code-key manifest, not a home-grown Go verifier. The checked registry now
  generates compiler-checked Go constants; the two production sites use those constants; and real
  Postgres integration tests assert the emitted payload field at each site.
- Canonical docs now state the honest boundary: build gates prove registry grammar, generated
  drift, and copy completeness; integration tests and range review prove runtime use. Source,
  function, and field entries remain precise review locators.
- The build and runtime text validators now reject two-or-more trailing spaces at line end, with a
  shared hard-break reproducer.
