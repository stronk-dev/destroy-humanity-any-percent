# Game UI GU-C25–GU-C28 transition witness manifest

Predeclared 2026-08-22 for platform-alignment execution-queue row 4. This batch may change the
accepted Game UI RFC's production projection, generated API contract, governed Copy assembly,
client controls, tests, canonical docs, Make verification composition, and planning records. It
may not change balance values, gameplay formulas, CI workflows, or the archived first-hour
script.

## Populations and oracles

| Seam | Population | Required oracle | Negative/discrimination control |
|---|---|---|---|
| Production preview | A fresh Tier-0 Company below and at the standard `gate.t0_to_t1` requirement; a Tier-1 Company; and later-tier/route controls. | One pure preview leaves both input states byte-identical, exposes only the standard first gate with `route_id:null`, matches the real gate transition outcome, makes Wind Down eligible only under the production exit predicate, and fails closed for later gates/routes. | Sever the gate requirement comparison and the below-requirement row must change; sever the shared Wind Down predicate and the Tier-1 row must change. |
| Snapshot v3 | Initial transaction-local state and stored real-Postgres Company/Founder state, including v1/v2 compatibility fixtures. | Current mint is exact v3 with `transitions`; ordinary sync resolves current Company and Founder revisions; stored v1/v2 bootstrap receipts remain registry-valid and the client parses them without controls. | Remove `transitions`, add an undeclared field, or relabel a v2 fixture as v3 and the registry/client oracle must reject it. |
| Desk controls | Parsed v3 snapshots with cross-gate absent, ineligible, and eligible; Wind Down ineligible and eligible. | Controls render only for v3 server-projected rows; disabled state equals `eligible`; activation sends the exact existing intent through `runtime.intent` using observed revisions and renders the authoritative receipt result. | Disconnect each handler and its visible milestone assertion must fail; a v1/v2 snapshot must never render either control. |
| Terminal continuation | Both decoded `run_ended` variants with a next snapshot at `ended.run_seq+1`; request failure and same-run/multi-run mismatch controls. | The parent-owned control calls only `runtime.snapshot()`, preserves the byte-only `RunEndSurface` boundary, and clears terminal/offer state and selects Desk only on the exact successor. | Suppress either terminal rendering or disconnect continuation and the corresponding assertion must fail; failure/mismatch must remain on Run End with the existing offline state. |
| Existing browser stories | Cap explanation, drain notice, and resync story beat in the composed browser lane. | Three independent visible assertions identify the exact story outcome, not merely process exit. | Remove each outcome independently and its assertion must fail. |

## Measurement limits

- Eligibility is advisory display state. Intent receipts remain authoritative under concurrent
  changes and stale revisions.
- The browser population proves the UI-owned transition and terminal seams with ordinary
  server-side precondition setup. It does not replay the already-proven two-hour first-session
  policy and introduces no fixture clock or direct browser intent calls.
- The manual 4x throttle/frame observation remains manual evidence; existing CI-observable
  performance checks remain the automated bound.
- These witnesses authorize closeout only after cold execution, recorded severing failures, and
  the mandatory Claude cross-party exact-range review. The implementer does not archive.
