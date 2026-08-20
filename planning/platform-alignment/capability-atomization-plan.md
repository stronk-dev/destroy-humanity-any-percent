# Predeclared capability atomization and trace plan

Predeclared: 2026-08-21. Product coordinate: `190a4fa`. Authority population: all 121 unique parent
rows in `design-capability-ledger.tsv`, which cover the 14 binding design documents
`design/00-vision.md` through `design/13-world.md`.

The section-level ledger was useful for finding broad absences, but many rows combine independently
falsifiable promises. This plan freezes the split and evidence rules before child verdicts are
written. Parent preliminary states are leads only; children do not inherit them.

## Question

What is the smallest set of independently falsifiable user/operator outcomes that covers every
current design promise, and for each outcome what exact producer, consumer, deploy-current data,
default workflow, failure/refusal path, executable witness, and authority route exist at the
product coordinate?

## Atomization rule

One child row describes one actor-observable outcome that can receive one verdict. Split a parent
whenever any conjunct has a different:

- actor or workflow entry;
- producer/state transition;
- consumer or presentation surface;
- current content/data requirement;
- success witness;
- refusal/failure/offline/accessibility behavior; or
- implementation/RFC/owner/research blocker.

An interaction sequence may remain one row only when its acceptance witness must exercise the
whole sequence and no step can truthfully ship on its own. Shared infrastructure is not a user
outcome; record it as a producer for each relevant child instead of granting umbrella capability
credit.

Every parent receives at least one child ID `<parent>.01`; additional children increment in design
order. Parent IDs and exact design references remain stable. A child cannot use `and`, `/`, or a
comma-separated feature list to conceal separately shippable outcomes; grammatical conjunctions
inside one indivisible action/response are allowed and must be explained in `evidence_limit`.

## Mandatory split controls

The following known aggregates must not survive as one child:

- `V-003`: fallback is traced per multiplayer capability family, not as one global law;
- `T-002`: headcount allocation, Soul, faction choice, minigame access, and elective Exit are
  independent outcomes;
- `E-007`: Reputation, Network slots, Route Knowledge, Clout, Personal Wealth, and Age each receive
  their own producer/consumer trace;
- `M-001`: clock, economy hook, fallback, unlock staggering, and persistence are separate contract
  outcomes and then applied to each current minigame tenant where applicable;
- `S-003`: guild lifecycle, resource exchange, weekly ritual, and NPC network are separate;
- `A-003`: rate limiting, idempotency, invariant handling, minigame validation, board verification,
  and save protection are separate security outcomes;
- `UX-004`: the five ruled Run End beats are separate unless one composed witness proves the exact
  whole sequence;
- `CP-002`: each content family with a different loader/runtime consumer is separate.

As negative controls, atomic foundations such as canonical Decimal wire parity may remain one row
only when their existing cross-runtime witness and failure mode are genuinely shared. An entirely
absent feature still receives the exact intended consumer/data/acceptance shape; “future RFC” alone
is not a trace.

## Output schema and verdicts

`capability-outcome-ledger.tsv` contains exactly these columns:

```text
capability_id, parent_id, design_ref, user_outcome, actor, producer,
consumer, current_data, default_workflow, executable_witness,
failure_or_refusal, verdict, authority_route, evidence_limit
```

Closed verdict vocabulary:

- `proven_integration` — current producer, consumer, real data/default workflow, and a
  discriminating executable witness all exist;
- `proven_bounded_primitive` — the named outcome is itself a foundation contract and its complete
  boundary is executable, but no larger player feature is implied;
- `partial_integration` — at least one real producer-to-consumer path exists but a named part of the
  outcome/default/failure workflow is missing or contradicted;
- `backend_or_data_only` — current mechanics/data exist without the intended mounted consumer;
- `client_or_fixture_only` — a component/helper/fixture exists without the real producer/data path;
- `claimed_only` — a record claims the outcome but current implementation evidence does not;
- `absent` — no current implementation of the bounded outcome exists;
- `blocked` — the factual state is established but the next action requires named research,
  owner/ruling-author, accepted-RFC, or cross-party-review authority.

`blocked` is not used to avoid a capability verdict: the row records the factual implementation
state in `evidence_limit` and the exact authority blocker separately.

## Evidence lanes

For every child, inspect independently:

1. exact binding design lines and any later owner ruling that supersedes/reconciles them;
2. accepted/active/archive RFC criterion and lifecycle truth;
3. production backend producer, persistence/event boundary, and refusal behavior;
4. production client/operator consumer reachable from its real entry point;
5. current epoch/platform data, not a testdata or candidate fixture;
6. the default user/operator workflow, including provider-off/offline/recovery/accessibility where
   relevant;
7. a current executable witness whose fixture can trigger failure; and
8. the RP, research, decision, ruling-author, accepted-RFC, or review route for every gap.

Reuse the exact package, route, migration, client, workflow, catalog, copy, runtime/event, and test
inventories as indices, never as inherited proof. Inspect source/test bodies for every promoted or
partial row. Cold command evidence already recorded at the same product coordinate may be reused
only within its stated denominator and limitation.

## Reconciliation and failure conditions

- All 121 parents must map to one or more unique children; zero parents may disappear.
- Every child must map back to exactly one parent and one exact design reference.
- Parent coverage and child-ID sequence are machine-checked; duplicate, skipped, or malformed IDs
  fail the extraction.
- Every `proven_*` row must name a producer, consumer when applicable, current data/workflow, and
  executable witness. `none` in a required lane fails the verdict.
- Every backend/data-only row must name the missing intended consumer; every client/fixture-only row
  must name the missing producer/data path.
- Every non-proven row must have an existing authority route or file a new RP in both ledgers.
- A fixture, generated type, source file, archived RFC, or green aggregate cannot promote its child
  without the other lanes.
- Any unresolved dynamic content/consumer grammar is explicit and blocks completion; it may not
  shrink the denominator.

Manual second pass: every `proven_integration`, `proven_bounded_primitive`, and
`partial_integration` row; every row that cites mounted Game UI or current gameplay data; every
mandatory split-control family; and every new RP. A separate contradiction pass compares the final
child distribution and summary prose across `capability-map.md`, `reality-audit.md`, `inventory.md`,
the execution queue, and the review handoff.

## Authority limit

This wave may classify current truth, refine/file RP rows, bound research/decision/RFC dependencies,
and reconcile planning summaries. It may not change design intent, author owner rulings, accept or
implement an RFC, edit player copy/content, flip implementation plan boxes, archive work, loosen an
acceptance bound, or turn an absent future feature into release scope.
