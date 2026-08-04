# Achievements Foundation implementation log

## 2026-08-03 — accepted-contract reconciliation

The owner accepted C1–C10, narrowing the RFC from Clout + achievements to Achievements only. The
file still contained the rejected Clout mint/stack in its summary, normative sections, acceptance
criteria, open questions, README status, and last changelog entry. Reconciliation makes one rule:

- achievement IDs and exact achievement score are permanent and non-spendable;
- ordinary Company transitions latch run IDs; the existing multi-stream Exit settles Founder
  lifetime state idempotently and resets the next run;
- Clout has no writer or production factor in this foundation; future social activity owns its
  sole mint and PR-Intern content owns any run-local multiplier;
- Phase-A predicates use only implemented state/event boundaries; Meters predicates append later;
- production definitions are content and copy-key data, not invented foundation mechanics.

`make verify` passed at `61b6392` immediately before this implementation round.

## 2026-08-03 — catalog, proof, and evaluation foundation

- Added strict schema-v1 Go/TypeScript loaders. Achievement IDs are byte-sorted; score grants are
  positive exact integers; every copy, generator, event, resource, and counter binding resolves
  through an explicit composition registry.
- Added the bounded condition union and the three proof arms. Possession requires an ownership
  predicate plus justification copy, provenance refuses current-possession predicates, and burn
  requires a registered event/resource plus positive canonical Decimal minimum.
- Added pure evaluation against one observation snapshot, lifetime/run latch exclusion, and score
  derivation from the earned-ID set as the single authority.
- Added JSON Schema fixtures and a fail-closed package/source gate proving this foundation imports
  no economy spending, production/save owner, or lifetime-Clout symbol.
- Focused Go, client, TypeScript type-check, schema, and boundary gates pass. No production
  achievement or copy was invented; fixtures are explicitly pre-mint.

## 2026-08-04 — independent foundation review (`61b6392..d5d7685`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved; remediation required before the next Achievements landing.** The
  Clout-free scope and pure evaluation behavior hold, but the range starts and ends with a red
  mandatory gate and the proof/catalog boundary is not yet cross-runtime or structurally sound.

Findings, ordered:

1. **HIGH — the range re-claims a mandatory gate that is red.** `make typecheck` fails at the
   Meters corpus import of `node:fs`; that defect entered in the range's parent `61b6392`, but
   `d5d7685` nevertheless records that the Achievements TypeScript type-check passed. A batch may
   not inherit a red required gate and report it green. Close the shared Node-type defect and
   re-run the real root gate before this range is approved.
2. **HIGH — Go and TypeScript disagree on accepted achievement catalog bytes.** Go decodes every
   condition/proof arm into one struct and tests zero values, while TypeScript enforces exact keys.
   Thus Go accepts, for example, `fact_present` with an extra `counter:""`, `provenance` with an
   extra `event_kind:""`, or `burn` with `event_kinds:null`; TypeScript and schema reject all three.
   Go also trusts registry membership without mechanically validating registry-backed copy,
   generator, counter, event, resource, and possession-justification IDs, while TypeScript validates
   those identifiers. Add exact arm presence checks and a shared cross-runtime valid+invalid catalog
   corpus; schema validation outside the runtime loader is not a substitute for parity.
3. **HIGH — provenance is registered but not proven.** C6/CA1 require a provenance condition to be
   derived solely from its declared immutable event kinds. The registry exposes independent sets
   of counters/facts and event kinds, with no source mapping, and `parseProof` merely rejects an
   ownership predicate. A `fact_present`/counter condition can name one source while its proof lists
   any unrelated registered event kind and both loaders accept it. Provide typed condition-source
   metadata (or another executable derivation proof) before production definitions or runtime
   earning can land; otherwise AC2 cannot be satisfied from the run log alone.
4. **MEDIUM — the source boundary is not the fail-closed build-graph gate the log/docs claim.**
   `verify-achievements-boundaries.mjs` scans only immediate non-test Go files and direct path/symbol
   text. A nested/transitive package bypasses the import rule, and the check does not implement C8's
   future sole-Clout-owner/source-registry proof. Current code is clean, but the enforcement claim
   is stronger than the gate. Use a recursive import graph/AST fixture and keep the sole-mint
   registry obligation explicitly pending until its owner exists.

What held:

- Current production source imports only the Decimal boundary and contains no Clout writer,
  lifetime-Clout factor, ledger debit, persistence owner, or production contribution. Achievement
  score remains a distinct derived exact integer and unknown/false earned entries fail closed.
- On valid fixtures, Go and TypeScript agree on byte-ordered definitions, bounded predicate-tree
  evaluation, run/lifetime latch exclusion, possession/provenance basic shape checks, canonical
  positive burn amounts, and derived score. The schema and current direct-source boundary fixtures
  pass, as do the focused Go/client/build/vet gates listed in the Meters review above.
- No achievement artifact was added to the epoch seed. Literal production achievements/copy,
  save v16, live/replay evaluation and ordered events, Exit settlement, relevance observation,
  epoch identity, and archival remain explicitly pending and were not over-credited by this review.
