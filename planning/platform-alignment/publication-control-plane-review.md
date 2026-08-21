# D-002 Control-Plane Publication Review

**Reviewed:** 2026-08-21  
**Population:** all 13 Class-A artifacts from `publication-sensitivity-audit.md`  
**Result:** **2 are current-ledger candidates; 11 are historical-snapshot candidates; 0 contain a
publication-sensitive identifier found by this review.**

All files were read in full. They remain ignored and were not edited, moved, force-added, deleted,
or published.

## Verdict

The Class-A material is suitable for public repository memory after the global disposition blocker,
but most of it is not suitable as a live planning authority.

| Disposition | Count | Artifacts |
|---|---:|---|
| **Track as maintained current ledgers** | 2 | `design/BACKLOG.md`; `planning/coverage-map/deferred-and-dropped.md` |
| **Track as a frozen historical coverage-map archive** | 11 | `planning/coverage-map/README.md`, `fce-mint-punchlist.md`, `gap-backlog.md`, `map.md`, `research-integration.md`, and all six `validated/*.md` slices |
| **Refuse for sensitivity** | 0 | — |

The two current-ledger candidates still require ordinary currentness reconciliation when they are
added. The 11 historical artifacts should not be refreshed piecemeal into a second live control
plane; preserve their reconstruction evidence together under a dated archive path with a banner
that names the snapshot date and points to `planning/platform-alignment/` for current status.

## Evidence

### Current-ledger candidates

`design/BACKLOG.md` is the shared topic/defect ledger the repository process already names. Its
current release-defect section is connected to the platform-alignment queues and its older ideas
are explicitly candidates, designs, research, RFCs, or rejections rather than an implementation
queue. Keeping it ignored is a shared-memory defect. It contains product ideas, source names and
legal cautions, but no credential, personal contact, private host, workstation path, or private
repository recipe found by the review.

`deferred-and-dropped.md` records why removed mechanics are not gaps and the exact conditions under
which they may return. That is durable decision memory, not generated output. It is dated and
contains no sensitive identifier. Before tracking, reconcile any status that changed after
2026-08-05 or label the row's last-checked date.

### Historical-snapshot candidates

The remaining 11 artifacts are a coherent 2026-08-05 through 2026-08-10 reconstruction wave. They
contain valuable evidence, but their actionable language is stale:

- `map.md` already carries a 2026-08-16 staleness stamp and says a full revalidation is owed;
- the six domain slices describe old RFC/index states, pre-content capability counts, and review
  debt that later platform-alignment work superseded;
- `gap-backlog.md` calls several now-landed or superseded RFCs draftable/blocked;
- `fce-mint-punchlist.md` assigns now-historical work to Claude/Codex and predates completed mints;
- `research-integration.md` uses “NOW RUNNING,” “queued,” and “scheduled” for an August 6–10
  research wave that is no longer the live execution queue;
- the directory README presents its outputs as a source of truth without an archive/frozen
  qualifier.

Those are not reasons to discard the artifacts. They are reasons to preserve them transactionally
as one dated historical archive. Their evidentiary value is the reconstruction at that time; a
partial rewrite would destroy provenance while still failing to make the collection current.

## Publication and authority checks

- No credential, private key, email address, private host, absolute workstation path, or sibling
  deployment recipe was found in the 13-file population.
- External names appear as research/design references or legal cautions, not as private personal
  data.
- Proposed mechanics and copy remain clearly inside backlog/research-routing history and do not
  become adopted design merely by being tracked.
- Commit hashes, review-agent names and old queue assignments are project provenance. They are
  safe to preserve in a historical archive but must not be read as present authorization.
- This review does not certify every old status claim as still true; it certifies the correct
  authority disposition for the artifacts.

## Required tracking transaction

After the ruling author reconciles the stale private/never-publish disposition:

1. reconcile and normally add `design/BACKLOG.md` plus `deferred-and-dropped.md` as maintained
   ledgers;
2. move the other 11 artifacts together to a dated `planning/archive/coverage-map/` home, add a
   frozen-snapshot banner/read-order, and repair repository links in the same change;
3. remove only the corresponding durable-memory ignore rules, leaving diagnostics ignored;
4. run a fresh-clone authority check proving every process-named ledger exists and every link to
   the historical coverage map resolves;
5. do not mark old unchecked work READY merely because its historical record becomes visible.

No push, publication, archive move, ignore-policy change, product change, or owner-content adoption
is authorized by this review alone.

## Execution status — 2026-08-21

The 11-file historical coverage-map move is complete in
`publication-disposition-execution-02.md`. Each file is tracked under
`planning/archive/coverage-map/` with a frozen/noncanonical banner and current-authority pointers.
The two maintained ledgers remain pending their separate currentness reconciliation.
