# Commons Onboarding & Governance — readiness log

This is draft readiness work, not an implementation plan or acceptance verdict.

## 2026-10-13 — Dependency availability is not an executable player/governance contract

Inspected HEAD `0073094e`, the complete draft/design05§5, current producers/consumers,
and original changes `c2d72b5c`/`fa5cef0d`. The original draft is Codex-authored;
`fa5cef0d` is a mixed bulk reconciliation co-authored by Claude, not proof that
every Commons clause was independently authored/reviewed. Review by: Codex
(author-side readiness/first filter). Recorded by: Codex. **No designated approval.**

The account/faction/guild/transport foundations exist. That does not close these
remaining implementation boundaries or make this draft accepted:

1. **Ballot authority and wire:** the paragraph gives directions/plurality/tie-hold,
   but not exact window-ID/request/receipt/result/tally/state schemas, revision and
   uniqueness ownership, durable opening epoch/constant identity, or atomic close/
   restart behavior. Calendar-derived identity alone does not persist the actual
   first-accrual opening time or pin its band. Define late-first-evaluation, no-vote,
   cohort-merge, concurrent vote/leave, restart and repeated-resolution cases.
2. **What the dial changes:** identify the authoritative field changed by
   `new_tithe_ppm` and its relation to per-Company `CompactTithePPM`, catalog default/
   bands and replay-owned accrual. Define the effective timestamp and splitting of
   a lazy interval across it. Do not infer that a group ballot may rewrite every
   member's private contribution or mutable constants. `ballot_step_ppm` is new
   grammar, not a field admitted by the existing strict Commons catalog; assign
   its version/mint and resource lane rather than inserting an untracked literal.
3. **Cohort producer → wire → panel:** `transport/wire.go` admits a generic
   `{scope,rev,state:<object>}` snapshot; only World gets domain-state validation.
   Channel authorization/history is not the closed neighbor/standing/Health/
   Capacity/NPC DTO or its publisher. Event envelopes currently admit only Company/
   Founder scope. Name the exact domain descriptors, producers, routes/channels,
   recovery and privacy/membership bindings before calling them "existing schemas".
   The current composition explicitly wires player, guild-presence and World
   producers; no closed cohort-panel producer is identified by this trace.
4. **Formula/UI dependencies:** the draft omits the accepted API ownership and
   current Game UI consumer contract. C11/C18 prohibit a current-document fallback
   for missing historically pinned formulas. Specify the artifact/identity the
   panel actually reads, refuse its absence honestly, and name the incorporation/
   cohort/leave consumers. A matching raw document is not a proven player workflow.
5. **Body consistency:** the Specification/Open Questions still call the ballot
   and panel boundaries DESIGN-GAPs while the later section closes all six blockers.
   The author/ruling author must reconcile the normative body, not merely retain
   two competing claims or promote status after acceptance of dependency RFCs.

Evidence: `server/production/intents.go` has Sign/Leave Compact but no ballot arm;
`server/commons/catalog.go`'s exact decoder has no ballot-step field;
`server/transport/wire.go`'s snapshot/event validation and
`server/gameserver/composition.go`'s publisher wiring show the boundary above.
Repository searches locate cohort authorization/history plus a subscribed-channel
integration fixture, not the claimed closed panel contract. Absence of ballot
implementation is expected for a draft; the defect is missing buildable specification,
not failure to implement unaccepted code. This is source/specification evidence,
not runtime or production-population execution. No software suite for this review.

Route RP-529 to draft author/lane owners: first close the technical DTO/ownership/
migration/dependency contracts, distinguish any genuinely unruled tithe/clock
semantics for the owner, then obtain normative cross-party review and explicit
adoption. Already-set direction-only governance/current standing/population-invariant
Health/NPC fallback design stays intact. No new mechanics or balance value chosen.

Only the index/readiness records change in this checkpoint. The draft body and
authored ruling/literal text are untouched; no product/schema, implementation queue,
accepted RFC status, review approval, archive or publication change. Next product
construction needs that specific closure, not another generic "unblocked" label.
