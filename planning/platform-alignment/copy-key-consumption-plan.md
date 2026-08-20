# Predeclared copy-key consumption audit

Predeclared: 2026-08-21. Product coordinate: `190a4fa`. Population: the 208 unique keys in the
checked-in generated application catalog `client/src/copy/generated/catalog.json`.

This plan is frozen before row classifications are generated. The prior archive replay already
demonstrated that the checked-in 161-key orphan report contains known production Game UI calls;
that observation is a required positive control, not permission to choose a convenient extraction
grammar around those keys.

## Question

For every shipped copy key, what current producer/reference lane can select it, does that lane reach
a mounted player workflow at the audited coordinate, and can the existing orphan report distinguish
live, backend-only, fixture-only, held-candidate, and genuinely unreferenced text?

## Exact authority lanes

Each key is checked against all lanes independently; the highest-capability lane never erases the
others:

1. **Mounted literal client call:** a literal `t()` or `resolveCopy()` call in the static production
   dependency graph rooted at `client/src/main.ts` (including the production Worker only when it
   can resolve copy).
2. **Mounted dynamic client binding:** a key selected through the Game UI presentation/event-copy
   artifacts, strict generated types, or a runtime DTO, with an exact mounted component/function
   consuming that binding. Merely appearing in a JSON file or TypeScript union is insufficient.
3. **Registered deploy-current artifact reference:** a value reached through
   `copy/references.v1.json` against the 19 artifacts in the current epoch. This proves shipped
   declarative input, not a mounted display.
4. **Explicit Go producer reference:** a generated constant or the two sites declared in
   `copy/code-reference-sites.v1.json`, with its runtime event/receipt producer and any mounted
   consumer recorded separately.
5. **Other strict catalog reference:** Pitch, Soul, Fiscal, opportunities, presentation/curriculum,
   or another exact loader-owned copy field not enumerated by the current registry. Classify its
   artifact as deploy-current, candidate/fixture, or platform input and trace the actual consumer.
6. **Test/fixture/tool reference:** tests, fixture host, candidate compilers, schemas, or generators
   only. Generated `CopyKey` membership alone is not a consumer.
7. **No discovered reference:** no lane above. This is a cleanup/research candidate, not proof that
   deletion is safe; dynamic keys may remain and must be called ambiguous unless their grammar is
   bounded.

The production reachability graph reuses `client-source-inventory.tsv`; it is not recomputed from
filename intuition. Candidate-labelled copy files remain shipped source inputs under RP-096, so
`source filename = candidate` is recorded independently from actual generated-catalog inclusion.

## Row schema and verdicts

The output contains exactly one TSV row per generated key with:

```text
key, source_catalog, source_label, params, tone, current_orphan_report,
mounted_literal_sites, mounted_dynamic_sites, deploy_artifact_refs,
go_producer_refs, other_catalog_refs, test_tool_refs, workflow,
verdict, evidence_limit
```

Closed verdicts:

- `mounted_player_copy` — an exact mounted production workflow selects the key;
- `shipped_backend_or_data_only` — deploy-current producer/data can select it, but no mounted player
  consumer exists;
- `shipped_unmounted_surface_copy` — intended surface binding is shipped but that surface is not
  mounted/implemented;
- `fixture_or_tool_only` — only test, fixture, or tooling lanes select it;
- `unreferenced_candidate` — no bounded lane selects it;
- `ambiguous_dynamic` — a dynamic selection grammar could include it and has not been bounded.

No row becomes `mounted_player_copy` merely because a component, catalog loader, generated type, or
backend producer exists.

## Controls and fired criteria

Required controls:

- `desk.buy_one`, `chrome.run_title.company_fallback`, and
  `screen.run_end.founder_note` must classify as mounted production copy even though the current
  orphan report includes them.
- At least one deploy-current achievements/Pitch/Soul key with no mounted player surface must remain
  backend/data-only; this prevents catalog presence from equaling user capability.
- The content-free `FixtureHost.svelte` calls must remain fixture/tool-only unless another mounted
  lane independently selects the same key.
- An injected in-memory fake key absent from every source/reference lane must classify unreferenced;
  if the extraction cannot reject it, the instrument is invalid.
- A dynamic nonliteral selection site with an unbounded source must produce `ambiguous_dynamic`, not
  silently mark all or none of the catalog live.

A criterion fires when the current orphan report labels a mounted key orphan, omits an actually
unreferenced key, relies on a registry that excludes a shipped copy-bearing family, or when two
canonical/current artifacts give incompatible source/activation state for the same key.

## Execution and limits

Extraction may use a read-only script or shell pipeline, but the checked-in result must retain exact
file:line evidence and reconcile to 208 unique catalog keys. Any parser truncation, syntax failure,
unbounded computed-key grammar, duplicate key, or missing catalog row fails loud; it cannot shrink
the denominator. A manual second pass inspects every `ambiguous_dynamic` row and every verdict family.

This audit may authorize only RP refinement, a Copy Pipeline successor/amendment, a bounded
reference-extraction witness, or transactional record correction. It cannot delete copy, rename or
adopt candidate content, edit owner-authored prose, change the generator, or implement a player
surface.
