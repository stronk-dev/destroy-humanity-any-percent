# D-002 Disposition Execution — Public Policy and Source-Neutral Derivatives

**Execution date:** 2026-08-21
**Scope:** two ruling-author policy reconciliations, two Class-B public derivatives and the exact
ignore exceptions required to track them
**Effect:** research/planning publication boundary only; no raw dossier, product behavior, design
intent, RFC, player copy, push or deployment changed.

## Executed dispositions

| Input | Public result | Boundary |
|---|---|---|
| `design/research/README.md` | Normative body now implements D-002: public shared memory after file-specific review; ignored raw source is explicitly noncanonical. | Tracked as the research-process index. |
| `planning/coverage-map/decisions-log.md` | Appends the 2026-08-21 superseding ruling without rewriting the 2026-08-06/07 historical sequence. | Tracked as an append-only ruling record. |
| ignored `cattery-reusables.md` | `cattery-reusables-public.md` | Source-neutral patterns only; no sibling path, hostname, port, route, endpoint, credential, topology or filename. Raw input remains ignored and noncanonical. |
| ignored `cicd-deploy.md` | `cicd-deploy-public.md` | Dated public synthesis; sibling/machine identity and recipes removed; stale cost premise replaced by R-001's 974.510 s / 739.441 s / 212.927 s local observation. Raw input remains ignored and noncanonical. |

`.gitignore` exposes exactly those four reviewed public artifacts while retaining the raw corpus,
generated diagnostics and all other unexecuted dispositions behind the ignore boundary. Nothing
was force-added.

## Checks

- sensitive-identifier scan over both derivatives found no email, workstation path, sibling
  hostname or reviewed operational number;
- the two raw inputs remain ignored;
- the two derivatives, research README and decisions log are normally visible to Git;
- the research matrix points the affected rows at the public derivatives and does not claim that
  either synthesis adopts CI, deployment or pet behavior.

## Remaining D-002 execution

1. preserve the reviewed coverage-map and old Codex queue as frozen historical records;
2. reconcile and track the two maintained ledgers;
3. execute the five eligible and six revision-blocked Class-C dispositions; the 56 synthesis raw
   inputs remain ignored behind their tracked public batch records unless a later bounded public
   derivative is justified;
4. add and run the fresh-clone authority gate.
