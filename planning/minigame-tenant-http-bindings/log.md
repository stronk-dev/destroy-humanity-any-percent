# Minigame Tenant HTTP Bindings — proposal log

## 2026-10-14 — Concrete public-wire route and routing evidence

Baseline `6a5a742a`; prior goal turn corrected Transport producer/client conformance. Continue
the accepted program by preparing the recorded TT-PA4/C2 integration hold, not implementing
around it. Inspect existing operation/response schemas, handlers/receipts, Typer's18 fields,
the pinned minigame_api loader and client registry. The existing API owns a Pitch ref/command
union, the artifact requires exactly four legacy IDs, and the client imports deploy-current
metadata. New routes alone cannot make a correctly bound public Typer path.

Before the mechanical check, declare same-method static-prefix versus disjoint-prefix and
different-method controls, both operation-ID orders. Actual `NewRegistry`, `CanonicalOperationPins`,
`CheckCompatibilityPin`, `Registry.Mount`, chi dispatch and response validators execute cold
in `TestTenantNamespaceRouteProbe`. Six cases pass0.190s: both shadow arms pass metadata checks
but change the old valid URL's reply to a body failing its original response descriptor;
all four controls preserve that reply and all added operations serve their own declared DTO.
The named mechanical-ID descriptor admits typer before routing. No registration/mount stub,
network socket, real Account authorization, Typer engine, Postgres or production handler run.

After the terminal run, move the owned probe to
`planning/api-foundation/probes/tenant-route-probe.go.txt`. It is deliberately outside CI:
normal verification must not assert that this known bad admission stays allowed. Reproduce by
temporarily moving it to `server/publicapi/tenant_route_probe_test.go` with apply_patch, running
root `make test-go GO_PACKAGES='./publicapi' GO_TEST_FLAGS='-count=1 -run TestTenantNamespaceRouteProbe -v'`,
waiting for termination, then moving it back. No source mutation or pin rewrite is needed.

RP-562 records this engineering/author contract finding, not a demonstrated live route fault.
New draft recommends disjoint Typer operations, truthful directory/claim discovery, metadata2
with explicit role IDs, atomic family admission before mutation, generated guards, pinned child
binding and whole player proof. Literal descriptor/guard and transaction/version-floor seams
remain named unfinished specification work; the draft is NOT acceptance-ready or authority.
Typer's arithmetic/result, authored copy/content, AT and broader release gates remain separate.

Accepted bodies, product/server/client/balance/schema/kernel/generated files are unchanged.
Index/owning plans and canonical compatibility limitation are reconciled; the probe/compiler
and document paths/diff are checked. No broad software suite is claimed for prose, no self-
approval, lifecycle/mint/push/deploy or full1.0 promotion. Review by: Codex (proposal/research
author first filter). Recorded by: Codex. Complete proposal/source/evidence/record range after
`6a5a742a` through this containing commit needs independent review before normative adoption.
Next: finish the four concrete engineering seams in the plan, then adoption/body reconciliation
and designated review; no request that the owner design JSON fields or transaction APIs.

Marco asks what "pending contract decisions" means. Distinguish genuine product/policy choices
from specification-author repairs and implementer engineering. This proposal's four unfinished
technical seams belong to Codex; approval is not being requested for an incomplete design.
Headcount's hiring/output choice and account export/deletion/retention semantics remain actual
owner choices in the existing decision queue. They do not halt unrelated accepted work.

## 2026-10-14 — Transaction seam specified from current owners

Resume at `e0f0ad5d`, clean tree. Previous turn made progress: preserved the executed routing
finding and separated technical work from owner choices; no product acceptance advanced.
Inspect actual save-start, minigame claim/receipt, gameserver adapter, Account ownership and
terminal resolution owners. Create replay returns before its mutation callback, which can
also run after a Founder-genesis insert; command replay and claim currently use separate
transactions. Therefore a handler-only family check or callback added solely around tenant
execution cannot enforce the proposed before-write guarantee on every path.

TB4 now names required new/replay admission callbacks before writes/return, an account-scoped
ClaimOrReplayAPI transaction, pinned historical binding and read-only snapshot interfaces.
Preserve create-key mismatch taxonomy, existing request hashes, lease/clock/receipt owners
and internal backend callers. Resolve a documentation ambiguity from actual production:
FinalizeResolutionTx stores the inner resolution receipt, while InsertCommandReceiptTx stores
the full terminal HTTP envelope. Reads reconstruct the former; command replay returns the
latter. No new namespace field, receipt store, migration, polling or payout algorithm proposed.

Verification: trace every proposed insertion point against the named current functions in
`server/save/minigame_start.go`, `server/minigame/session.go`/`service.go`,
`server/production/minigame_start.go`/`minigame_api_command.go`/`minigame_resolution.go`,
`server/gameserver/minigame_api.go` and `server/account/store.go`; inspect diff/paths.
No software suite for this draft-only change and no SQL/atomicity claim. Real concurrent
Postgres failure populations are specified, not executed. No production or accepted-body
change, adoption, independent approval, push/deploy or release advancement. Review by: Codex
(proposal author first filter). Recorded by: Codex. The complete draft range after `6a5a742a`
still needs independent normative review before adoption. Next: literal descriptors/guards
and metadata2 loader/version-floor details; not another owner implementation question.
