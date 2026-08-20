# Predeclared archived-RFC risk replay

Predeclared: 2026-08-21. Product coordinate: `190a4fa`. Population: all 46 Markdown files directly
under `rfc/archive/` and their 46 corresponding directories under `planning/archive/`.

This plan is frozen before the acceptance replay is classified. An exploratory read exposed a
possible Fiscal Quarters plan/body contradiction; that observation receives no final finding or
sample credit until Fiscal is replayed under the same rules below.

## Question

Which archived RFCs still have trustworthy implemented scope at the current product coordinate,
and which archive records overstate behavior, acceptance evidence, canonical docs, or review
provenance?

The audit tests archived scope only. It does not assume that an archived backend primitive creates
a current player capability, and it does not reopen later successor scope merely because the
original system has changed.

## Structural population pass (all 46; mandatory)

For every archived RFC, record:

1. exact RFC filename and mapped planning directory;
2. status text and RFC line count;
3. open plan-checkbox count;
4. explicit `Review by:` and `Reviewed range:` token counts in the archived log;
5. direct links from active RFCs to the archived filename;
6. canonical docs named by `rfc/README.md`;
7. existing RP rows that already route a current defect through the archived system;
8. risk score, selected stratum, and final replay verdict.

Filename equality must reconcile the 46 RFCs and 46 planning directories after applying only these
three historical slug mappings:

- `0002-economy-constants-and-ceilings` -> `0002-economy-kernel`;
- `founder-attendance-foundation` -> `founder-attendance`;
- `t0-t1-playable-content` -> `t0-t1-content`.

No missing/extra pair may be hidden by a sample.

## Risk score (predeclared)

Add the following independent weights; ties sort by raw-byte RFC filename:

- **+5**: an existing tracked RP row names a current behavior, docs, lifecycle, oracle, or
  provenance defect owned by the archive;
- **+4**: the archived plan contains any unchecked checkbox;
- **+4**: the archived log lacks either an explicit `Review by:` token or an explicit
  `Reviewed range:` token (provenance-format risk, not automatic historical invalidation);
- **+3**: at least one active RFC directly links the archived RFC filename;
- **+2**: the archived RFC exceeds 500 lines (scope/reconciliation risk).

The score ranks audit attention. It is not a severity verdict and cannot retroactively apply a
newer review format to work completed before that format existed.

## Mandatory deep-replay strata

The deep sample must contain at least 15 archives and all of the following, even if that exceeds 15:

1. every archive with an unchecked plan box;
2. every archive scoring 9 or higher;
3. every archive owning an existing RP defect that affects release, privacy/retention,
   accessibility, current CI/evidence, or an active RFC dependency;
4. at least one remaining archive from each domain: numeric, economy, save, production, client/UI,
   content/copy, runtime/transport, multiplayer, replay/leaderboard, and harness/CI;
5. three low-risk controls with no open checkbox and no existing RP defect, preferring records with
   explicit current-style review provenance.

The final artifact must name selection reason per row. No archive may be silently excluded merely
because its test suite is expensive or its plan/log is awkward.

## Deep replay procedure

For each selected archive, trace:

```text
archived intent and literal AC
  -> current producer/schema/data
  -> current composed consumer or explicit fixture-only boundary
  -> canonical docs
  -> executable witness and demonstrated discriminator
  -> archived plan/log/review range
  -> current active-successor relationship
```

Read the full archived RFC, plan, and log. Re-run the narrowest current cold root gate that can
exercise the claimed behavior. Existing lifecycle mutations may be reused only when their exact
failure is recorded; source/test presence alone is structural evidence.

## Fired criteria

Classify a row as a defect/contradiction when any of these fires:

- archive status says implemented while the plan still presents implementation/review work as
  unfinished without explicitly marking it carried out of scope;
- a normative body clause contradicts an accepted ruling/correction instead of being reconciled;
- canonical docs promise a producer, scheduler, consumer, retention, or player workflow absent at
  the product coordinate;
- an acceptance witness cannot trigger the claimed failure, silently skips its required
  population, or lacks the stated fixture/data;
- an archive claims a designated review/range that does not resolve or cover the claimed span;
- current active RFCs consume a dependency state that the archive itself does not actually prove.

Valid negative results include fixture-only implementation, backend-only implementation, superseded
current behavior with a linked successor, and historical provenance that cannot be reconstructed.
These must be reported honestly; they are not softened to preserve `implemented` wording.

## Exit criteria and authority limit

The replay completes only when:

- all 46 structural rows reconcile;
- the mandatory deep sample and ten-domain floor are met;
- every fired criterion has a tracked RP route and exact evidence;
- all commands, skipped denominators, and invalid runs are recorded;
- current summaries and the review handoff agree with the result.

This research may authorize only record correction, owner/ruling-author work, accepted successor
RFC drafting, or a bounded witness-remediation queue. It cannot edit frozen archives, repair
owner-authored rulings, implement product behavior, retroactively approve review ranges, or archive
active work.
