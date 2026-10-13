# Minigame Tenant HTTP Bindings — proposal plan

Authority today: drafting/evidence only; `rfc/minigame-tenant-http-bindings.md` is unaccepted.
Baseline `6a5a742a`. Full nine-tier1.0 remains the objective; this is one blocked integration seam.

Completed mechanical evidence: real registry construction, compatibility check, mount and
chi dispatch in six cases. Both static same-method variants shadow the old admitted URL;
disjoint prefix and different-method controls preserve it. Source is the manual
`../api-foundation/probes/tenant-route-probe.go.txt`, outside ordinary test collection.

Engineering specification now ready for normative review, not implementation:

- TB2 names every new request/success descriptor, all eighteen snapshot fields, exact error
  subsets and an explicit registered drain binding. Today's main-intent guard does not cover
  minigame routes; the new drain DTO does not widen old APIError/Pitch schemas.
- TB3 specifies retained operation metadata, exact role maps, both runtime models/accessors,
  raw admission, historical normalization, unchanged save floors and genuine loader identity.
- TB4 names before-write create/new/replay admission, one-transaction command replay/claim,
  account-scoped locks and read-only snapshot interfaces. Stored inner-resolution and full
  command response receipts are distinct. Atomicity still needs real Postgres proof.
- TB5 specifies a generated bounded descriptor closure and guarded schema-byte decoder,
  preserving raw UTF-8/BOM/literal identity. Existing SDK/pair checks are not full guards.

Next: obtain proposal adoption, author-body reconciliation and designated normative review.
Do not treat this draft's readiness for review as acceptance or mutate parents in another
author's name. RP-460/RP-162 remain separate author obligations. Review may require corrections;
none of the specified new endpoints, codecs, guards, drain bindings or client registration exists.

After those gates, the RFC's staged acceptance requires genuine catalog/integration identity,
real Postgres, native generated-client play/reload/failure paths, preserved Pitch dispatch and
complete review. No artifact mint, public route, UI registration or archival is authorized now.
