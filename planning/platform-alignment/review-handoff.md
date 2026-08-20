# Bounded cross-party adversarial review draft

Prepared 2026-08-20. Product coordinate: `190a4fa`. Implementer/audit author: Codex. Required
designated reviewer under `AGENTS.md`: Claude (the other party), not a Codex-spawned or self-named
reviewer.

State: **not the final repository-audit handoff.** The 2026-08-21 mechanical first filter found
that Wave 1 still lacked complete package/route/migration/client/catalog/executable/archive-risk
inventories and Wave 2 still retained coarse design rows. This document remains useful as the
review contract for the completed active-RFC/release/decision/dependency/queue milestone, but a
designated verdict over it must not be represented as approval of the unfinished exhaustive audit.
The server package/operation/migration boundary and the client source/workflow boundary have since
been reconciled; catalog/copy, executable/test-artifact, planning/docs, actor/worker/event depth,
archived-RFC risk sampling, and Wave-2 capability splitting remain before finalization.

## Range to review

The substantive audit program starts after product commit `190a4fa`. The original bounded core
through Wave 6 was:

```text
190a4fa..1e47752
```

That range contains 19 planning/documentation commits beginning with `cb162a3`; `c7eb024` is the
subsequent queue/handoff commit. Do not start the final designated pass from that historical range.
After Waves 1–2 close, the reviewer must resolve the then-current local tip and cite literal hashes;
`HEAD`, “latest,” or a branch name is not acceptable provenance.

The range changes README/current-state/RFC-index claims and adds the platform-alignment control
plane. It must contain no product code, schema, balance, copy, migration, active RFC normative-body,
or deployment behavior change. The user's uncommitted `AGENTS.md` edit is outside the range and
must remain unstaged.

## Review objectives

Review adversarially, not as a prose polish pass:

1. Re-derive the product coordinate and verify every commit in the range is planning/research or
   current-status/index reconciliation only.
2. Recount active RFCs and the 111 acceptance rows. Verify the ledger distribution (39 draft,
   five mechanical/review-pending, 20 proven/qualified, 33 partial/unmet, ten contradicted/failed,
   four withdrawn) from row values rather than trusting the summary.
3. Sample every lifecycle dossier against the actual RFC body, current code/data/docs/tests, plan,
   history, rewrite map, and cited review provenance. Prioritize every promoted/contradicted row.
4. Re-run or inspect the command-level negative controls. Confirm temporary mutations were restored,
   failing fixtures could reach the defect, cold `-count=1` was used where claimed, and invalid
   parallel/timeout runs were excluded rather than softened.
5. Challenge all release claims: provider-off, packaging/config, account rights, accessibility,
   operations/retention, and preservation. In particular verify RP-080, RP-083, RP-086, and RP-088
   from production consumers/call sites rather than prose.
6. Verify every RP row through the final audit tip exists in the tracked ledger and each has a research, owner,
   ruling-author, accepted-RFC, review, or explicit completed-process route.
7. Validate all 30 rows of `dependency-resource-ledger.tsv`: ten columns, unique IDs, one canonical
   owner/gate, named consumer, refusal owner, witness owner, state, and exact blocker. Look for
   duplicate API/client/account/telemetry/operations authority hidden by synonyms.
8. Check the owner packet does not silently adopt a decision. Challenge the recommended Phase-0
   preview floor and confirm every hosted-service obligation is either in the floor or explicitly
   owner-deferred.
9. Check `execution-queue.md` and `ready-batch-manifest.tsv`: only Q-001/Q-002/Q-003 may say READY,
   their authorities are accepted and bounded, conflicts force serial execution, and forbidden
   product/owner scope is explicit.
10. Confirm no active RFC is marked implemented/archived and no plan checkbox was flipped from
    audit inference.

## Required verdict format

Append the actual verdict to `planning/platform-alignment/log.md` with:

```text
Review by: Claude
Recorded by: [Claude, or exact transcriber]
Reviewed range: <literal start>..<literal end>
Method: [files/commands/mutations sampled]
Findings: [severity, file/row, evidence, required repair]
Verdict: APPROVED | CHANGES REQUIRED
Range-union claim: [exactly what this verdict covers; no broader claim]
```

An approval with findings still open is `CHANGES REQUIRED`. A delegated/self first filter may be
recorded separately but is not this designated pass. The reviewer does not archive active RFCs or
authorize product implementation merely by approving the audit control plane.

## Post-review routing

- Only after the remaining Wave-1/2 inventories close and the complete audit range is approved may
  this verdict be called the repository-audit review. The owner can then rule the READY decisions in
  `owner-ruling-packet.md`, and Q-001/Q-002/Q-003 remain the only implementation-ready batches.
- If findings change factual classifications, repair the ledgers/dossiers transactionally and
  obtain a new exact-range verdict over the repair edge.
- If the owner rules O-001 or O-002 while review is running, land that ruling separately; do not
  fold owner intent into the audit verdict.
