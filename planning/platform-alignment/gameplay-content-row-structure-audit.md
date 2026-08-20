# Deploy-current gameplay-content structural inventory

Coordinate: product tree `190a4fa`; 2026-08-21. This is the structural result of the predeclared
grammar in `gameplay-content-row-plan.md`. It assigns no reachability or capability verdict.

## Reconciled population

All 19 `epoch_artifact` rows in `balance-file-inventory.tsv` are present exactly once as source
documents. The deterministic walker emits **579 authored units**:

| Structural kind | Units |
|---|---:|
| Object records in arrays | 297 |
| Nested condition/proof/effect/policy objects | 135 |
| Primitive relationship/value edges in arrays | 48 |
| Explicit empty collections | 39 |
| Root scalar policy fields | 43 |
| Top-level singleton policy objects | 17 |

| Family | Units | Family | Units |
|---|---:|---|---:|
| Achievements | 50 | Economy | 175 |
| Categories | 21 | Commons | 21 |
| Curriculum | 7 | Doctrines | 3 |
| Factions | 11 | Fiscal | 6 |
| Guilds | 10 | Meters | 56 |
| Minigame API | 5 | Minigames | 16 |
| Opportunities | 9 | Pets | 23 |
| Pitch | 37 | Prestige | 19 |
| Relevance | 55 | Routes | 45 |
| Soul | 10 |  |  |

Every unit has a repository file, RFC-6901 JSON pointer, structural kind, authored identity, and
SHA-256 of its exact JSON payload. A rerun reproduces the TSV byte-for-byte.

## Why the row denominator matters

The 19 file rows concealed several materially different subpopulations:

- the economy file alone contributes 175 objects/edges/policies, including generator ladder rows,
  upgrade requirements/effects, role/source relationships, manual/offline policy, and progress
  coordinates;
- achievement, route, and meter rows retain their nested condition/proof/predicate/band/input
  objects rather than receiving one family-level “loaded” result;
- primitive arrays such as roles, event kinds, tier curves, route predicates, and score facts
  remain exact edges rather than disappearing inside a parent object; and
- 39 empty collections remain visible. They include 19 Relevance grouping collections, nine meter
  input collections, four faction modifier-slot collections, three early-gate route collections,
  and one each in economy, categories, Commons, and Soul.

An empty collection is not automatically defective: the evidence pass must determine whether it is
an intentional empty relationship, a zero-output placeholder, dormant content, or a contradiction.
The important control is that it cannot silently vanish. The same applies to root scalar zeroes,
including the current guild consumption-bonus rate.

## Controls and authority limit

The extractor rejects removal or duplication of any unit and separately proves it rejects omission
of an empty collection or a root policy field. It also fails unless the frozen 19-file population
reconciles exactly with the prior file inventory.

This checkpoint is structural only. Loader, transition, current trigger, player consumer,
executable witness, reachability verdict, and authority route still belong in
`gameplay-content-row-ledger.tsv`. No schema validity, epoch inclusion, payload hash, parent family,
or file-level consumer may be inherited as row proof.
