# Minigame Tenant HTTP Bindings — proposal plan

Authority today: drafting/evidence only; `rfc/minigame-tenant-http-bindings.md` is unaccepted.
Baseline `6a5a742a`. Full nine-tier1.0 remains the objective; this is one blocked integration seam.

Completed mechanical evidence: real registry construction, compatibility check, mount and
chi dispatch in six cases. Both static same-method variants shadow the old admitted URL;
disjoint prefix and different-method controls preserve it. Source is the manual
`../api-foundation/probes/tenant-route-probe.go.txt`, outside ordinary test collection.

Next engineering work before adoption/review (owned by Codex, not owner design questions):

1. Complete the literal Typer/directory descriptors and exact operation error subsets,
   using current snapshot/receipt ownership rather than another public schema table.
2. Specify the generated client guard mechanism from the same Go descriptors; bind trusted
   host time/claimed-session read behavior without adding mutation retries or polling.
3. TB4 now names the create new/replay admission callbacks, one-transaction command
   replay/claim, account-scoped locks and read-only snapshot interfaces. Implement and prove
   them against real Postgres only after adoption; source-derived placement is not atomicity
   evidence. Preserve the stored inner-resolution versus full-command-response distinction.
4. Close schema2's Go/TS/version-floor/replay representation and exact compiled role bindings;
   preserve all schema1 historical bytes. Reconcile TT-PA4/MA-C9/MA-C15 through their authors.
5. Obtain adoption and designated normative review. Do not infer either from dependency
   availability or from the routing probe. RP-460/RP-162 remain separate author obligations.

After those gates, the RFC's staged acceptance requires genuine catalog/integration identity,
real Postgres, native generated-client play/reload/failure paths, preserved Pitch dispatch and
complete review. No artifact mint, public route, UI registration or archival is authorized now.
