# RFC: Copy Pipeline Foundation

- **Status:** implemented
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/08` (the flavor bible — voice rules, era presentation, the tonal law), `design/11` (UX writing, first-session), the research-provenance conventions (`design/research/README.md`)
- **Research:** the whole flavor corpus; `gaming-enshittification.md §6` + `societal-satire.md` (voice), all `[V]/[M]` provenance discipline
- **Depends on:** implemented client/tooling foundations only (C1 ruling)
- **Unblocks:** UI Foundation (UF3 consumes this artifact), all content RFCs' copy
- **Owner ruling honored:** breadth-first — the copy SYSTEM (keys, artifact, lints), not the copy.
- **Planning:** `planning/archive/copy-pipeline-foundation/`

## Summary

Every player-facing string in the game flows through one pipeline: a `copy_catalog` artifact of
keyed strings, typed params, era variants, and — critically — **mechanical enforcement of the two
design laws that keep the satire safe**: the no-🔴-names rule and the no-unverified-statistics
rule. Content RFCs declare `copy_key` fields; this pipeline resolves, validates, and lints them.

## Specification

### CP1 — The copy catalog artifact

`copy/*.json` source files are strict `{schema_version:1,entries:[...]}` documents. Each entry is
exact-key `{key,text,params,era_variants,provenance,tone}`; `legal_reviewed` is deliberately absent
because an unauthenticated boolean proves nothing. Keys are namespaced
(`desk.buy.tooltip`, `event.garage_plaque.title`). Text uses typed placeholder syntax for params.
Era variants override the base per era (UF1's `era_1995`/`era_2000`); resolution is
`key@era → key` fallback. The catalog is a build artifact (like the formula document), diffable,
reviewable.

### CP2 — The completeness gate

`copy/references.v1.json` is the authority for copy-bearing catalog paths. The completeness gate
walks only those registered JSON-pointer patterns against the epoch-seed artifact map. Every
referenced key must resolve; a miss fails naming `artifact:path:key`. Unreferenced copy is emitted
to a deterministic warning report because copy may legitimately precede its content.

### CP3 — The legal lint (the tonal law, mechanical)

- **No-🔴-names:** the copy linter greps every string against the maintained denylist of
  trademarked/protected names from the research legal matrices (Neopets, Habbo, Blizzard marks,
  real living individuals in a mocking frame…). A 🔴 term in shipped copy fails CI. The denylist
  is `moderation/copy-denylist.txt`, sourced from the research files' legal-axis tables,
  extend-only.
- **Provenance enforcement:** the deliberately narrow detector covers literal numbers adjacent
  to `%`, configured currency/unit tokens, and historical four-digit years. Placeholders are
  ignored. A detected literal or explicit provenance declaration MUST resolve through
  `copy/provenance.v1.json` to a `verified` claim with a unique research anchor and HTTPS source;
  plausible/model/absent provenance fails shipping (the
  verify-before-shipping law made a gate). This is the single most important safety mechanism in
  the copy system — it makes "anything on a verify-before-shipping list is flagged, not shipped"
  impossible to violate by accident.

### CP4 — Voice (advisory, not gated)

The flavor-bible voice rules (`design/08`) stay human-reviewed — voice is not lintable. But the
pipeline provides the hooks: a `tone` tag per key (corporate/diegetic/lore-card/achievement),
and the lore-card carve-out (sourced population-level statistics allowed as no-joke cards, the
2026-07-28 owner decision) is a `tone: lore_card` flag the provenance lint requires `[V]` on but
the voice review knows carries no joke.

## Owner rulings on C1–C10 (2026-08-03)

- **C1 — accepted:** dependency direction corrected (Copy Pipeline → UI, not reverse); the real
  `client/src/copy/` boundary lands here, UI consumes it (no UI stub).
- **C2 — accepted:** strict artifact grammar as proposed — source files `{schema_version:1,
  entries:[exact-key rows]}`, globally-unique mechanical-ID keys, `tone` closed union
  `corporate|diegetic|lore_card|achievement`, era keys exactly `era_1995|era_2000` byte-sorted,
  NFC UTF-8, byte-deterministic generated `client/src/copy/generated/catalog.json` + TS types.
- **C3 — accepted:** params `string|integer|canonical_decimal`; plain text only (no HTML/Markdown);
  `{name}` placeholders with `{{`/`}}` literals; exact-param runtime contract (missing/extra/type
  throws in dev, returns key + invariant report in prod); `canonical_decimal` inserted verbatim
  (`<Amount>` owns notation); Svelte escapes after resolution.
- **C4 — accepted:** a checked-in strict `copy/references.v1.json` registry names every JSON path
  that is a copy reference (incl. `name_key`, `incorporation_copy_key`, `reason_key`) — the
  completeness gate walks the registry, never property-name matching.
- **C5 — accepted:** the existing catalogs' already-required copy fields (name_key etc.) are the
  first registry rows; their copy lands with this foundation so nothing ships key-less.
- **C6 — accepted, THE load-bearing ruling: a machine-readable `copy/provenance.v1.json`
  registry.** Rows `{claim_id, source_file, source_anchor, status, source_urls}`; `status ∈
  verified|plausible|model`; **verified requires ≥1 HTTPS URL and a resolving unique anchor under
  `design/research/`**; copy rows carry sorted-unique `provenance:[claim_id]`; **shipping accepts
  ONLY verified claims** (plausible/model representable for draft fixtures, fail the production
  gate); verified-row source changes are append/supersede, never in-place relabel. The gate proves
  registry CONSISTENCY, honestly not re-doing the research — that's the right, non-overclaiming
  boundary. (This replaces my `research-file#claim` grep, which Codex correctly showed can't prove
  identity.)
- **C7 — accepted:** the statistic detector is deliberately NARROW and honest — number adjacent to
  `%`/currency/closed-unit-token, or a four-digit year in the historical range; placeholders
  ignored (runtime mechanics); voluntary provenance allowed; **no production suppression flag** (a
  detected statistic without verified provenance fails, period). Under-detection is accepted and
  stated; the gate never pretends to catch every factual claim — voice review remains the human
  backstop.
- **C8 — accepted:** the denylist contract is the proposed case-folded NFC, separator-collapsed,
  Unicode-token matcher over an append-only, byte-sorted file. Every entry cites a research legal
  matrix; the full-history guard fails closed on shallow clones. It enforces known red-list terms,
  not general legal safety.
- **C9 — accepted:** `legal_reviewed` is removed. Independent commit-range review is the human
  assurance record; any future attestation must bind reviewer identity to a content hash.
- **C10 — accepted:** copy has a content-addressed `copy_hash` separate from `constants_hash`.
  Mechanical keys stay in receipts/events. The framework-independent client package owns
  `loadCopyCatalog`, `resolveCopy`, generated key/param types, and the singleton artifact. Root
  `copy-generate`/`copy-check` targets enforce generation, completeness, provenance, denylist
  history, and fixtures; CI runs the check with full Git history.

## Acceptance criteria

1. Catalog round-trip; era-fallback resolution; typed-param validation (wrong param type fails).
2. Completeness gate: a seeded content field referencing a missing key fails CI naming both;
   an orphan key flags without failing.
3. Legal lint: a seeded 🔴 name in copy fails CI; the denylist is extend-only (a removal fails).
4. Provenance gate: a seeded statistic without `[V]` provenance fails CI; a `[V]`-backed one
   passes; a `tone: lore_card` requires `[V]`.
5. UF3 integration: the UI's `t()` resolves against this artifact (the interface UF3 stubbed).

## Resolved acceptance review record (Codex, 2026-08-03)

The original draft was not executable. C1-C10 below record the gaps and proposals the owner
subsequently adopted above. They are retained as decision history, not open blockers.

### C1 — Status and dependency direction contradict the ruled queue

The RFC is `draft` and says it depends on UI Foundation. The ruled dependency is the reverse:
Copy Pipeline → UI Foundation → Game UI/T0–T1. UI's accepted C4 already names this RFC as its hard
dependency.

**Proposed contract:** Copy Pipeline depends only on the implemented client/tooling foundations;
UI Foundation is listed under `Unblocks`, not `Depends on`. Once C1–C10 are ruled, set status to
`accepted`; planning begins only then. No UI stub ships—the real `client/src/copy/` boundary lands
here and UI consumes it.

### C2 — The artifact grammar is not closed

`copy/*.json: {key,...}` does not say whether files contain one row or arrays, how rows compose,
what the checked-in build artifact is, how duplicate keys/order/encoding are handled, or where
`tone` lives (CP4 requires it but CP1 omits it). `era_variants` also permits arbitrary era names.

**Proposed contract:** source files each contain one strict root
`{schema_version:1, entries:[...]}`. Entries are exact-key
`{key,text,params,era_variants,provenance,tone,legal_reviewed}`; keys use the repository mechanical
ID grammar and are globally unique. `tone` is the closed Phase-A union
`corporate|diegetic|lore_card|achievement`; era keys are exactly `era_1995|era_2000` and are
byte-sorted. UTF-8 text is NFC-normalized, contains no control characters except newline, and has
declared byte/line bounds. The generator recursively reads source files in byte-path order,
rejects duplicate/trailing JSON, and writes canonical `client/src/copy/generated/catalog.json`
plus generated TS key/param types. The output root is `{schema_version:1,entries:[...]}` sorted by
key; generation is byte-deterministic.

### C3 — Placeholder and parameter semantics are undefined

The parameter `type` union, placeholder syntax, repeated/unused fields, missing/extra runtime
params, formatting, and markup/escaping policy are unspecified. A typed declaration alone cannot
make `t()` safe or deterministic.

**Proposed contract:** Phase A parameter kinds are exactly `string|integer|canonical_decimal`.
Text is plain text only—no HTML/Markdown and no raw-markup escape hatch. Placeholders use exact
`{name}` tokens with `{{`/`}}` as literal braces; every declared param appears at least once in
base text, no undeclared placeholder appears, variants use exactly the same param set, and duplicate
param declarations reject. Runtime `t(key, params, era)` requires exact params (missing/extra/type
mismatch throws in dev/tests and returns the key with one invariant report in production).
`integer` is a JS safe integer rendered base-10 without grouping; `canonical_decimal` is validated
but inserted verbatim—`<Amount>` owns notation. Svelte performs DOM escaping after resolution.

### C4 — Completeness has no catalog-field authority

“Every `copy_key` field across all catalogs” misses current `name_key`,
`incorporation_copy_key`, and player-visible `reason_key` fields. No registry identifies which
JSON paths are copy references, and recursively matching property names would silently include
unrelated future fields or miss renamed ones.

**Proposed contract:** add one checked-in strict `copy/references.v1.json` registry whose rows are
`{artifact_name,json_pointer_pattern,field_kind}` with `field_kind` in
`copy_key|name_key|reason_key`. Phase A enumerates, at minimum: economy upgrade `copy_key`, economy
resource/provisioned-hardcap `reason_key`, faction `incorporation_copy_key`, and leaderboard
category `name_key`; future RFCs append rows in the same change as new copy-bearing fields. The
gate resolves artifact paths through the epoch-seed authority, walks only registered pointers,
rejects a registered pointer matching no schema field, and reports `artifact:path:key` for every
miss. Code-owned reason keys are either moved to a declared catalog or registered in a separate
generated code-key manifest—never found by a repository grep.

### C5 — The current catalogs already require owner-authored copy

At HEAD, the production catalogs reference one cap reason, five category names, and four faction
incorporation keys. A strict completeness gate cannot go green without shipping their English
base text. Inventing those lines is content work and violates this RFC's “system, not copy” scope.

**Proposed contract:** the owner supplies a literal Phase-0 seed-copy fixture for every currently
registered production reference, reviewed as content, or explicitly splits production completeness
activation into the T0–T1 content mint. This RFC still gates all valid/invalid fixtures and exposes
the real resolver; it may not silently exempt current production files or generate keys as display
text. The selected deferral is named in AC2 and UI cannot claim production completeness before it
lands.

### C6 — Research Markdown cannot support the claimed provenance proof

`research-file#claim` is not a stable or verifiable identity. Most dossiers have prose-level
`[V]` markers, headings without claim IDs, and source URLs elsewhere in the file. A Markdown grep
cannot prove that a pointer names the intended verified claim, and line-number pointers drift.

**Proposed contract:** add a machine-readable, append-only `copy/provenance.v1.json` registry.
Each row is exact-key `{claim_id,source_file,source_anchor,status,source_urls}`; `status` is
`verified|plausible|model`, verified requires at least one HTTPS URL, paths must resolve under
`design/research/`, anchors must exist and be unique, and claim IDs are stable mechanical IDs.
Copy entries carry sorted unique `provenance:[claim_id]`. Shipping accepts only `verified` claims;
the other statuses remain representable for draft fixtures but fail the production gate. Changes
to a verified row's source identity are append-new/supersede, never in-place relabeling. The gate
proves registry consistency; it does not pretend to re-perform the research.

### C7 — “Detects a statistic” needs an exact, honest boundary

No regex can identify every factual claim, and “number + unit patterns” is undefined. Dates,
percentages, money, ratios, durations, version numbers, mechanical amounts, and fictional numbers
need different treatment. The current wording would both miss unnumbered factual claims and flag
mechanical UI copy unpredictably.

**Proposed contract:** the mechanical gate is deliberately narrower and stated honestly. It
requires provenance for text containing a decimal/integer adjacent to `%`, a recognized currency
symbol/code, or a closed unit token; and for four-digit years in the configured historical range.
Placeholders are ignored by the detector because their values are runtime mechanics. A copy row
may also declare provenance voluntarily. No suppression flag exists in production: a detected
literal must cite a verified claim or be rewritten without the literal. CI fixtures enumerate
every detector arm and near-miss. Docs say this prevents known numeric-provenance omissions; human
review remains responsible for unnumbered factual claims.

### C8 — The denylist has no matching or provenance contract

The research legal matrices are Markdown prose, not a complete term database. “Grep” leaves case,
Unicode, separators, substrings, inflections, quoted historical use, and multiword matching
undefined. The existing guild denylist guard does not automatically prove this new file is
extend-only, and the lint cannot prove an asset is legally safe merely because no known term
matched.

**Proposed contract:** `moderation/copy-denylist.txt` is an append-only normalized data file with
one case-folded NFC phrase per line, comments allowed, byte-sorted unique. Matching runs against
case-folded NFC text and a separator-collapsed form; single terms use Unicode word boundaries,
multiword phrases match normalized token sequences, and seeded punctuation/separator bypasses
must fail. A history guard checks every reachable commit and fails closed on shallow history,
using the existing governance-guard pattern. Each entry has a neighboring comment containing its
research legal-matrix reference. The lint is documented as enforcing the known 🔴 list, not proving
general trademark/defamation safety; human legal/editorial review remains authoritative.

### C9 — `legal_reviewed` is an unauthenticated boolean

Any author can set `legal_reviewed:true`; it proves nothing, and the RFC never says when it is
required or whether lints consult it. Keeping a decorative assurance flag would repeat the asset-
attestation overclaim the CI research explicitly warned against.

**Proposed contract:** remove `legal_reviewed` from copy entries unless an attestation owner and
review record are defined. Recommended Phase A contract: legal-risk lints are structural, while
human review is recorded through the existing commit-range review ledger; no JSON boolean implies
assurance. If retained later, it becomes a signed manifest with reviewer identity and reviewed
content hash, not a self-asserted field.

### C10 — Artifact identity, runtime ownership, and CI wiring are unresolved

The draft calls copy a “build artifact like formulas” but does not say whether copy bytes join the
simulation `constants_hash`. `design/12` says production content changes are epoch-stamped; putting
English punctuation in the balance identity would segment boards, while leaving it unversioned
would make historical UI/replay text drift. The exact `t()` API, missing-key production behavior,
orphan report destination, and root Make/CI gates are also absent.

**Proposed contract:** give copy its own content-addressed `copy_hash`, separate from simulation
`constants_hash`; deploy manifests pin both, and events/receipts continue storing mechanical keys
so replay semantics do not depend on English bytes. Copy changes follow review/deploy provenance
but do not mint a Balance Epoch unless they also change a simulation artifact. Export
`loadCopyCatalog(bytes)`, `resolveCopy(catalog,key,params,era)`, generated `CopyKey`, and a singleton
application catalog from `client/src/copy/`; no ambient network fetch. Missing keys are load-time
fatal for production catalogs, loud-key fallback only in dev fixtures. Orphans produce a
byte-stable checked-in report consumed as a warning, not a failing gate. Add root targets
`copy-generate` and `copy-check`; `copy-check` covers schema, generation drift, completeness,
provenance, denylist history, and fixtures, and is wired into the existing client/schema CI job.
The package has no Svelte dependency so UI consumes it without owning it.

## Changelog

- 2026-08-03: created (draft) — the copy system.
- 2026-08-03: C1–C10 ruled — dependency direction fixed; strict artifact/param grammar;
  reference/provenance registries; verified-only shipping; narrow statistic detector; normalized
  append-only denylist; unauthenticated review flag removed; separate copy identity and root CI
  gates. Accepted for implementation.
- 2026-08-03: Codex acceptance review recorded C1–C10; owner adopted all ten contracts.
