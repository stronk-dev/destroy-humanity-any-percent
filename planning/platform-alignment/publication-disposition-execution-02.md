# D-002 Disposition Execution — Frozen Historical Records

**Execution date:** 2026-08-21
**Scope:** reviewed historical coverage reconstruction, mint proposal and consolidated Codex fix
queue
**Effect:** preservation and authority labeling only; no historical claim was refreshed into
current truth and no product, design, RFC body, push or deployment changed.

## Preserved records

- The 11 reviewed coverage-map snapshot files moved together to
  `planning/archive/coverage-map/`, retaining their validated slices and reconstruction evidence.
- `mint-content-rows-proposal.md` moved into the same archive with an explicit noncanonical banner;
  its adopted payloads remain canonical only under `balance/`.
- `codex-fixes-2026-07-30.md` moved to `planning/archive/` as a frozen historical queue that assigns
  no current work.

Every preserved file has a top-of-file frozen/noncanonical warning and a pointer to current
platform-alignment authority. Live RFC/index references now point to the archive or canonical
production data. The Prestige audit explicitly records that tracking the old queue does not cure
its missing reviewer provenance.

The three `planning/coverage-map/draft-artifacts/*.json` duplicates remain ignored and untracked.
Their bytes were rechecked against `balance/{achievements,meters,pets}/first-content.json`; all
three comparisons are equal. No deletion or force-add occurred.

The changed local targets were checked directly and all exist. The previously advertised
`pnpm check:links` command is not a runnable repository gate: the root has no package manifest and
no Make/link-check target exists. Its invocation failed before inspecting links, so it is not cited
as evidence.

## Remaining D-002 execution

1. reconcile and track `design/BACKLOG.md` and `planning/coverage-map/deferred-and-dropped.md` as
   maintained ledgers;
2. execute the five eligible and six revision-blocked Class-C dispositions;
3. add and run the fresh-clone authority gate.
