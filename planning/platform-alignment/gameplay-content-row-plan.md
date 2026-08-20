# Predeclared deploy-current gameplay-content row audit

Predeclared: 2026-08-21. Product coordinate: `190a4fa`. Frozen input population: the 19 files
classified `epoch_artifact` in `balance-file-inventory.tsv`. Platform configuration, schemas,
candidate/historical/measurement fixtures outside those artifacts, Copy catalogs, and authored
design prose are excluded because their separate populations are already bounded.

The existing family inventory proves which files load. It does not prove that every row has a live
producer, effect, mounted consumer, or discriminating witness. This pass measures the authored
units inside the deploy-current files without treating “valid JSON” or “epoch 8 includes it” as
gameplay credit.

## Question

For every independently addressable authored unit in the current epoch artifacts, what loads it,
what production transition or presentation it controls, whether it is reachable with current data,
which mounted player consumer exposes its result, what executable witness discriminates it, and
which authority route owns any dormant, empty, zero-output, uncomposed, contradicted, or unproved
state?

## Structural denominator

Walk every frozen JSON document deterministically in source order. Emit exactly one unit for:

1. every non-root object, including a top-level singleton policy object, an object element of an
   array, and a nested condition/proof/effect object;
2. every primitive array element, recorded as a relationship/value edge owned by its nearest
   object or document root;
3. every empty array, recorded as an explicit `empty_collection` unit so missing content cannot
   disappear from the denominator; and
4. every non-`schema_version` scalar directly on the document root, recorded as a
   `root_policy_field` because no enclosing content object otherwise owns it.

Do not emit the document root, array containers with elements, `schema_version`, or scalar fields
inside an emitted object; those scalar fields are the indivisible payload of that object unit.
JSON pointers are RFC-6901 escaped and, together with the repository-relative file, form the stable
unit identity. Array indices remain source-coordinate identities for this audit and are not claimed
as durable product IDs.

This grammar is intentionally generic rather than a hand-picked list of familiar `id` rows. It
must therefore retain zero-length relationship sets, primitive tier curves, nested predicate/proof
records, role/source edges, and otherwise easy-to-hide placeholders.

## Output schema and verdicts

`gameplay-content-row-ledger.tsv` contains:

```text
unit_id, artifact, family, json_pointer, unit_kind, authored_identity,
loader, runtime_producer, player_consumer, current_reachability,
executable_witness, failure_or_limit, verdict, authority_route
```

Closed verdict vocabulary:

- `proven_mounted_effect` — current row controls a production transition whose result is consumed
  in the default mounted workflow, with a row-discriminating witness;
- `proven_mounted_presentation` — current row is presentation/curriculum data consumed on a mounted
  path with a discriminating render/selection witness;
- `partial_mounted` — a real production row and mounted consumer exist, but the default workflow,
  exact row selection/effect, refusal path, or integrated executable witness is incomplete;
- `backend_active` — the row is loaded and reaches a production backend transition/projection, but
  its intended player surface is absent or incomplete;
- `backend_registered_dormant` — the row is accepted/registered, but no deploy-current trigger,
  source, eligible target, or reachable workflow activates it;
- `measurement_only` — the row configures developer/harness measurement and is not gameplay;
- `uncomposed` — a deploy-current file or row has no production loader/consumer path;
- `zero_or_empty_placeholder` — zero/no-op payload or an empty authored collection prevents the
  represented effect/content family from existing;
- `contradicted` — the deploy-current row conflicts with binding design/RFC/ruling or another
  current authority; and
- `blocked` — current reachability is established, but completing the factual verdict requires a
  named invalid measurement, author reconciliation, owner ruling, accepted RFC, or cross-party
  review. The factual limitation must still be explicit.

Loader presence alone cannot exceed `backend_registered_dormant`. A component, TypeScript parity
module, generated key/type, fixture, archived RFC, or schema-valid row cannot promote content to a
mounted verdict.

## Evidence lanes

For each unit inspect independently:

1. exact current JSON payload and enclosing semantic owner;
2. epoch-seed/replay-catalog inclusion and strict loader/validator path;
3. production state transition, projection, scheduler, or presentation selector that reads the
   field/row;
4. current trigger/source/target data, including empty arrays, zero values, disabled bands, and
   unreachable gates;
5. production client surface and default workflow, not an unmounted parity helper;
6. executable row-specific or bounded-family witness and a failure capable of discriminating the
   row; and
7. exact RP/research/decision/ruling-author/accepted-RFC/review route for every non-proven state.

Shared loader and validator evidence may be named on multiple child units only within its exact
family boundary. Transition, reachability, mounted-consumer, and witness claims do not inherit from
the file or family row.

## Controls and failure conditions

- The 19 input files must reconcile exactly with `balance-file-inventory.tsv`; a missing or extra
  epoch artifact fails.
- Re-extraction must preserve unit count, order, file, pointer, kind, and payload identity.
- Seeded removal, duplicate-unit, empty-collection omission, and root-policy omission cases must
  fail structural validation.
- A seeded backend-only row relabeled mounted must fail unless it names a production player
  consumer and row-discriminating executable witness.
- A seeded zero/empty row relabeled active must fail unless current trigger/effect evidence replaces
  the zero/empty condition.
- Empty collections are findings, not permission to shrink the denominator. Nonempty collections
  do not also receive a synthetic umbrella row.
- Invalid harness runs, warm-cache summaries, schema acceptance, and fixture-only consumers cannot
  satisfy executable gameplay evidence.

Manual second pass: every proven/partial mounted, dormant, zero/empty, contradicted, or blocked row; every row in
economy, curriculum, Prestige, routes, opportunities, Soul, minigames, meters, pets, and relevance;
and every family whose children receive more than one verdict. A contradiction pass must reconcile
the final result with `catalog-family-inventory.tsv`, `capability-outcome-ledger.tsv`,
`capability-map.md`, `reality-audit.md`, queues, RP ledgers, and the review draft.

## Authority limit

This wave may classify current content truth, add audit-only extractors/ledgers, file or refine RP
routes, and reconcile planning summaries. It may not edit balance/copy/content, design intent,
owner-authored text, active/archive RFC bodies, product code/tests, migrations, canonical product
docs, implementation-plan checkboxes, or acceptance bounds. Content repairs require their accepted
RFC and the mandatory cross-party review protocol.
