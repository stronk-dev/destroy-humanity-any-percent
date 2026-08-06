# Copy pipeline

Player-facing text is authored as strict data, compiled into a deterministic client artifact, and
resolved by a framework-independent TypeScript package. Mechanical systems store copy keys rather
than English text.

## Authoring and generated artifacts

- `copy/catalog/*.json` contains schema-version-1 source catalogs. Entries have exactly `key`,
  `text`, `params`, `era_variants`, `provenance`, and `tone`.
- `copy/config.v1.json` owns the declared text bounds and the deliberately narrow statistic
  detector token/range configuration.
- `copy/references.v1.json` is the only authority for copy-bearing JSON paths in epoch artifacts.
  The verifier proves each registered path exists in that artifact's schema before walking the
  current artifact.
- `copy/code-reference-sites.v1.json` names explicit Go producer sites for code-owned reason keys.
  Generation emits `copy/generated/code-references.v1.json` and compiler-checked constants in
  `server/copykeys/generated.go`; the declared production sites use those constants, and their
  real Postgres integration tests assert the emitted event payloads. The source/function/field
  values are review locators, not a claim that a build script can prove arbitrary Go data flow.
  The gate proves registry grammar, generated drift, and copy completeness; it does not discover
  references by repository grep or pretend to replace runtime tests and diff review.
- `copy/provenance.v1.json` is the stable claim registry. A verified claim resolves to exactly one
  heading under `design/research/` and has at least one HTTPS source URL. **Since Amendment A1
  (2026-08-06) the tracked provenance source is `design/research/provenance-extracts.md`** — the
  publishable extracts file (the rest of the research corpus is unpublished/untracked); a verified
  claim may be retargeted ONLY under the unpublication-migration escape (old source is a
  `design/research/` path absent from the compared tree; id/status/urls immutable).
- `moderation/copy-denylist.txt` is the append-only known-red-name list. Each normalized term has
  an adjacent legal-section reference into the tracked extracts file (same A1 rule).
- `make copy-generate` writes the byte-sorted client catalog, generated key/param types, the
  independent `copy_hash`, the code-reference manifest, and the deterministic orphan report.

The `copy_hash` identifies English copy independently from the simulation `constants_hash`.
Changing punctuation does not segment a leaderboard. The generated
`deployment/content-manifest.v1.json` pins both current identities and `copy-check` recomputes both;
events, receipts, and saves continue storing mechanical keys.

## Runtime contract

`client/src/copy/` exports:

- `loadCopyCatalog(bytes)` for strict load-time validation;
- `resolveCopy(catalog, key, exactParams, era, options)`;
- generated `CopyKey` and `CopyParamsByKey` types;
- `applicationCopyCatalog` and the shorthand `t()`;
- `hashCopyCatalog` / `verifyCopyCatalogHash` for deployment or fetched-artifact identity checks.

Text is plain text. Svelte owns DOM escaping after resolution. Placeholders use `{name}`; `{{` and
`}}` produce literal braces. Parameter types are `string`, safe `integer`, and
`canonical_decimal`. Decimal text is validated through the shared numeric core and inserted
verbatim so this layer never invents numeric notation.

Missing/extra/wrongly typed params and unknown keys throw in development and tests. Production
requires an invariant reporter, returns the loud mechanical key, and emits exactly one
`copy_resolution_failed` invariant through it. Omitting the production sink is a configuration
error, never a silent fallback. Loaded fixture catalogs carry no application hash label; callers
verify bytes explicitly, while the generated singleton's identity is `COPY_HASH`. There is no
network fetch or ambient mutable catalog.

## Gates and their honest limits

`make copy-check` validates source grammar, generated drift, schema-backed completeness, explicit
code sites, provenance, known-red terms, protected history, deployment identity, detector fixtures,
and the orphan report. Research sources must be Git-tracked, so an untracked local dossier cannot
make a clean-checkout gate appear green. It fails closed on shallow Git history; the existing client
CI job checks out full history.

The statistic detector intentionally catches only literal numbers adjacent to `%`, configured
currency/unit tokens, and configured historical years. Runtime placeholders are ignored. A caught
literal, voluntary citation, or lore card requires verified provenance. This prevents known numeric
provenance omissions; it does not detect every factual assertion or re-perform research.

The known-name lint checks case-folded NFC tokens and separator/punctuation bypasses. It enforces
the maintained red list; it is not a general trademark, defamation, or editorial-safety proof.
Independent full-range review remains the human assurance record. No self-asserted
`legal_reviewed` boolean exists.

Denylist terms and their legal-section citations are history-protected together. The sole repair
exception is for a cited path that never existed anywhere in reachable Git history; deleting a
previously valid source cannot reopen that exception.

## Adding copy-bearing content

1. Add the content field to its strict content schema.
2. Append the exact JSON-pointer pattern to `copy/references.v1.json` in the same change.
3. Add the copy row and any verified claim row. Never derive display text from a key.
4. Run `make copy-generate`, then `make copy-check` from the repository root.
5. Review generated copy and the orphan report in the diff.
