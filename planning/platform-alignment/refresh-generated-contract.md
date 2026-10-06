# Refresh generated-contract census — 2026-10-06

Accepted API Foundation A2/A4/A5, C1/C7/C9; predeclared `6b279ef6`.
This is measured incompleteness, not a renewal implementation or adopted schema.

## Executed evidence

`make api-check` (60224) regenerates current OpenAPI/TS and compatibility-checks
without changing any generated artifact. Fourteen OpenAPI operations and fourteen
TS metadata rows agree exactly on operation ID/method/path. Neither includes
`POST /api/v1/session/refresh`; Account mounts that existing route separately.
Green generation/drift therefore does not prove all runtime routes are registered.

`make research-refresh-generated-contract` (19884) runs actual repository
TypeScript 5.9.2 against generated types and virtual callers, no emitted files:

| Actual generated contract caller | Compiler outcome |
| --- | --- |
| Registered bootstrap path | Compiles |
| Existing `invalid/body` APIError | Compiles |
| BootstrapSession's two-string pair shape | Compiles; no credential semantics proved |
| Existing refresh path in generated path union | TS2322 refusal |
| `unauthorized/refresh_token` in APIError | TS2322 detail refusal |
| `refresh_reused/session_family_revoked` in APIError | TS2322 category and detail refusals |

Every refusal is assignability on the virtual caller, not an import/module error.
Two synthetic counterfactuals are compiled **in memory only**: path-only addition
removes path refusal while both error refusals persist; error-only widening removes
both error refusals while path refusal persists. Twelve compiler arms total
(six baseline, three per counterfactual), 3815 ms. They discriminate representation,
not server rotation. No proposed operation ID, status, auth or schema is adopted.
The synthetic global error widening is deliberately not a proposed exact-pair
descriptor; API Foundation's production registry remains untouched.

AST census of the two named browser files finds three literal fetcher calls in
Game UI runtime (Founder state/bootstrap/intents) and one dynamic path call in the
minigame port. They are outside the generated directory. This verifies a C9 seam,
not a complete repository lint or an execution of those HTTP calls. Generated
metadata/type availability is not a generated dispatcher. WebSocket remains
separate; no HTTP retry/coordination/recovery policy was added.

[Retained report](refresh-generated-contract.v1.json) contains complete arms,
diagnostics, fourteen operations, four scoped call sites and ten before/after
source hashes plus actual HEAD. Instrument bytes were uncommitted during the run
and explicitly hashed; HEAD alone is not their identity. No private data/tokens.
Source stability covers only listed inputs/HEAD, not arbitrary files/environment.

## Verification and limits

- Cold publicapi tests/selected vet 75783 pass, no DB dependency.
- Root typecheck/client/topology 52260 pass: zero type/Svelte warnings/errors,
  7366 client tests/134 existing skips, 84 passed/17 skipped files, thirteen
  topology negative controls. New *.fixtures.mjs is not misclassified by Vitest.
- No full verify-client rerun or green claim: prior turn remains red at RP-131.
- No Docker workload/deletion. RP-236 capacity approval remains unresolved.
- No new HTTP/DB/browser proof. Account parser outcomes belong to the separate
  [executed census](session-refresh-contract-census.md); its prepared canonical
  rotation/replay/expiry populations and four severings remain unexecuted.

## Exact next boundary

RP-240 is the existing API Foundation C1/C7/C9 completion seam and draft renewal
S-A1 input. Finish real-DB outcome census before freezing error/status alternatives.
Pin an honest descriptor for existing parser behavior (including observed null/
missing/case-insensitive handling), without a silent acceptance-set tightening.
Then registry/generated dispatch migration must cover actual HTTP call sites,
auth mode and exact outcomes under its accepted authority/review range. Generated
BootstrapSession alone cannot authorize automatic renewal or recovery copy.
Owner coordination/ambiguity/storage/recovery decisions remain draft gates.

Review by: Codex (first-filter only). Recorded by: Codex. New independent range
begins `29e1ff02` exclusive, includes `6b279ef6`, instrument/Make/artifact/docs/
ledger/reconciliation and following pin. Claude required, not self-approved.
No production/generated/pin/auth/kernel/copy/CI membership/checkbox/archive/push
change, integrated-witness claim or reduction of proper full nine-tier 1.0.
