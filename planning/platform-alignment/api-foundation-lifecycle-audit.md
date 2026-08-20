# API Foundation lifecycle audit

Coordinate: product tree `190a4fa`; audit checkpoint after `35fa2ce`; 2026-08-20.

This pass re-derived the active API Foundation RFC from its complete specification and rulings,
live plan/log, schema/operation/cursor/policy packages, generated artifacts, account and gameserver
composition, production client callers, canonical API documentation, current tests, implementation
history, and tracked review provenance. Temporary discrimination mutations were restored. No
product code, owner-authored RFC/design text, canonical product documentation, or implementation
checkbox was persistently changed.

## Bottom line

The repository has a useful API *mechanics kit*: a closed schema DSL, immutable operation registry,
HMAC keyset cursors, strict policy loader, cache/rate/request-ID middleware, deterministic OpenAPI
and TypeScript generation, and an additive-compatibility checker. The cold package tests, root API
drift gate, and root TypeScript gate are green. A registered optional-response-field mutation makes
the generated-artifact gate fail with the exact diff.

That kit is not yet the accepted API platform:

- the generated registry contains 10 routes, while 11 additional live `/api/v1` account, session,
  founder, intent, and guild-intent routes are still hand-mounted and absent from OpenAPI;
- a temporary response-shape change to the unregistered live Founder handler leaves `make
  api-check` green, directly falsifying the one-authority coverage claim;
- `docs/generated/api.json` contains zero `/api/public/v1/` paths, and the production graph never
  constructs `publicapi.Runtime`, resolves cursor secrets, or mounts a public reader service;
- public catalog, epoch, board, Route Registry, genesis, replay-log, verdict DTOs/readers and their
  real-data conformance/privacy proof do not exist;
- the current 19-artifact epoch set still contains no historical formula artifact, so the accepted
  catalog transparency resource has no satisfiable current bundle;
- generated TypeScript is types plus operation metadata, not an HTTP client. The production Game
  UI directly calls bootstrap, founder-state, and intents through an injected raw fetcher;
- the boundary lint scans only Game UI `.svelte` files. A restored direct `/api/v1` `fetch()` added
  to `game-ui/runtime.ts` passed the gate;
- the compatibility checker rejects a removed nested field but reports only the root schema, not
  the removed field required by AC2; and
- generated OpenAPI omits the ruled request-ID, ETag, Cache-Control, conditional-request, and
  deprecation-header contracts.

AC1 and AC2 are therefore partial, AC3 and AC5 are unmet, and the corrected C9 form of AC4 is
contradicted. The stale literal AC4 in the acceptance section also contradicts C9's accepted ruling,
which independently blocks implementation/archival under the repository's body-reconciliation law.

## Current cold evidence

All valid commands ran from repository root:

- `make test-go GO_PACKAGES='./publicapi ./account ./gameserver' GO_TEST_FLAGS='-count=1'` — green;
  publicapi 0.279 s, account 0.548 s, gameserver 0.474 s;
- `make api-check` — green after regeneration and byte comparison;
- `make typecheck` — green, zero TypeScript/Svelte diagnostics;
- `make verify-client-boundary` — green before and after the restored blind-spot probe.

A direct root `pnpm type-check` attempt did not start because the repository has no root package
manifest. It is invalid evidence; the successful root `make typecheck` result replaces it.

The fast account/gameserver package pass does not claim real-Postgres behavior. That behavior was
already exercised in the immediately preceding same-product-coordinate Account and Minigame API
waves, but it cannot substitute for the absent public service.

## Restored discrimination probes

1. Added optional `probe_optional` to the registered `GameUIFact` response descriptor without
   updating generated files. `make api-check` exited 2 and showed the exact OpenAPI and TypeScript
   diffs. After restoring the descriptor and regenerating, the gate returned green. This proves
   registered additive drift is detected.
2. Removed nested registered field `GameUIResource.rate_per_second`. `make api-schema` exited 2,
   but its complete diagnostic was `invalid API operation: schema GameUISnapshot: invalid API
   schema`; neither `GameUIResource` nor `rate_per_second` was named. The mutation was restored and
   the gate returned green. This proves rejection but falsifies AC2's field-naming clause.
3. Added a direct `fetch("/api/v1/audit-probe")` to production `client/src/game-ui/runtime.ts`.
   `make verify-client-boundary` still exited 0 because the gate filters that directory down to
   `.svelte` files. The mutation was restored and the gate remained green.
4. Added `audit_probe` to the live hand-mounted `GET /api/v1/founder` response. `make api-check`
   exited 0 because that operation and schema are absent from the registry and generated contract.
   The mutation was restored, `gofmt` applied, and the generated diff returned clean.

The worktree after restoration differs only by the user's pre-existing `AGENTS.md` edit.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Partial** | Generation is deterministic, `api-check` is wired into `verify-server`, and a registered optional-field mutation fails with an exact artifact diff. But the source authority covers only 10 of 21 live `/api/v1` routes; a live unregistered handler response mutation passes. This is registry drift protection, not complete handler/router drift protection. | Migrate every v1 operation into one registry (or author-reconcile the claimed scope), bind runtime response validation, and retain both registered and formerly-handwritten negative probes. |
| AC2 | **Partial / literal diagnostic failed** | Current compatibility tests prove optional response growth passes and operation/field/constraint removal rejects. The restored nested-field removal also rejects. Its error names only `GameUISnapshot`, not the removed `GameUIResource.rate_per_second`; 11 live routes have no pins at all. | Make incompatibility errors report the exact operation/schema/property path and pin every live v1 operation. Preserve optional-addition and nested-removal command-level mutations. |
| AC3 | **Unmet** | There are zero generated or mounted public operations. `publicapi.Runtime` and cursor-secret resolution are test-only. No catalog/epoch/board/route/evidence reader, real verified-run round trip, or public/auth non-regression integration exists. | Implement all accepted readers and raw evidence against narrow repositories, compose the public runtime/router, add exact cache/limiter/304/error fixtures, and execute a real verified-run dereference/reverification workflow. |
| AC4 | **Contradicted under C9** | Generated TS types compile, but no generated HTTP client exists. `game-ui/runtime.ts` makes raw API calls for bootstrap, founder state, and intents, and the lint misses even a literal direct fetch in that file. The acceptance-section wording still demands deletion of a nonexistent prior client despite C9 replacing that criterion. | Ruling author first reconciles AC4's body. Then generate/implement the one HTTP transport boundary, migrate every caller, and enforce the rule across all client TS/Svelte sources with a failing runtime-file probe. |
| AC5 | **Unmet / currently vacuous** | The foundation unit fixture rejects a made-up `private_founder_id`, and Minigame successors prove private endpoint isolation. There is no real public registry to enumerate, no seeded-founder public responses, and no structural build-graph lint proving public readers cannot reach account/session/save repositories. | Enumerate every production public operation from the real registry against a private seeded founder, validate exact DTO allowlists and indistinguishable private/unknown responses, and retain a deliberately leaking DTO/reader negative. |

## Authority and generation gaps

### The registry is not the v1 authority

`account.API.Router` hand-mounts account create, session create/refresh/delete, email attach, account
delete, Founder create/read/import, Company intent, and Guild intent. It then mounts ten newer
Bootstrap/Game UI/Minigame/Soul operations through `PrivateAPIRegistry`. The generator consumes only
that latter registry. Therefore:

- OpenAPI clients cannot discover core account/session/founder/intent routes;
- compatibility pins cannot see their response-shape changes;
- the generated global error taxonomy is only the newer subset;
- startup duplicate/missing-binding checks do not govern the older routes; and
- A5's “router, OpenAPI, TypeScript, compatibility pins, and runtime validation consume those
  rows” is false for more than half of the live HTTP surface.

The mutation against `GET /api/v1/founder` makes this an executable coverage finding, not a route
count inference.

### The registry cannot yet express the accepted public contract

The only parameter descriptor is validated one-for-one against `{path}` segments and always emits
OpenAPI `in: path`. There is no query-parameter representation for boards' required `variables`,
`epoch`, `mandate`, `limit`, and `cursor`. Raw responses generate an OpenAPI binary body, but
`GenerateTypeScript` excludes raw success arms from `OperationTypes`; a verification operation
would expose only its JSON error DTOs to the purported client. Response headers have no generic
descriptor, so request IDs, ETags, cache policy, and future deprecation headers cannot be generated
from the single authority.

These are implementation gaps inside accepted A4–A8/C13/C19 scope, not reasons to weaken the
public contract. They must land before public rows can be registered truthfully.

### Public middleware is mechanically present but not live

The strict policy and middleware tests are good isolated proofs: unknown/duplicate config keys and
missing cursor secrets fail; HMAC is checked before parse; 304 precedes the limiter; ordinary
requests are bounded; request IDs echo; exact content bytes determine ETag. Repository-wide call
search finds `LoadPolicy`, `ResolveCursorCodec`, and `NewRuntime` only in `server/publicapi` tests.
No production composition loads `balance/api/phase0.json`, supplies stable deployment secrets, or
wraps a public route. Treating those tests as endpoint acceptance would confuse a producer with a
consumer.

## Normative, canonical, plan, and review drift

1. The acceptance section retains pre-ruling AC4 (“diff shows deletion”), while C9 explicitly
   replaces it with generated-only HTTP calling plus a raw-fetch lint. The body was never
   reconciled. Under the repository law, the ruling author must repair this before implementation
   continues.
2. A5 says one registry owns the API, but the 11 legacy routes remain outside it. The plan/log
   describe the generation slice as authenticated-only even though later Game UI and Bootstrap
   operations expanded it; canonical `docs/api-foundation.md` repeats that stale scope.
3. C18's sequencing note required artifact union growth with/immediately after the First Content
   mint. Epochs 6–8 have since landed and the current epoch declares 19 artifacts, but no public
   catalog schema arms or formula artifact exist. The plan still correctly leaves C18/public
   readers open, while the sequencing promise is stale.
4. A4/A8 require documented platform headers and conventions; generated OpenAPI contains no
   response headers at all and the canonical doc describes mechanics, not the ruled public header
   contract.
5. The initial schema/cursor work was reviewed by Darwin on the implementer side. The later
   Minigame designated Claude verdict explicitly consumed `f0fa2ae` and `4a8bdba`, which is valid
   scoped coverage for operational policy and authenticated generation, but it does not consume
   the earlier full foundation span or any not-yet-built public/client work. No API Foundation log
   entry declares an exact current-history range union and cross-party closeout verdict.

## Smallest honest closeout order

1. Ruling author reconciles AC4 with C9 and updates stale C18 sequencing/body statements without
   pretending the missing formula mint/public surface already happened.
2. Extend the registry authority for exact query parameters, response headers, and raw-success
   client typing; migrate all 11 legacy routes so every live v1 operation is generated, pinned,
   mounted, and runtime-conformance checked.
3. Improve compatibility diagnostics to the exact operation/schema/property path and add
   command-level additive/removal mutations over both a legacy-migrated and public operation.
4. Implement the generated/thin client transport and migrate all browser API callers; lint every
   TS/Svelte source outside the generated boundary, including alias/injected-fetch forms.
5. Perform the accepted formula artifact growth mint and register exact discriminated catalog
   arms for the complete historical artifact population; never serve current formula bytes under
   an old hash.
6. Implement and compose catalog, epoch, board, Route Registry, genesis, replay-log, and verdict
   readers through narrow read-only interfaces plus the real policy/cursor runtime.
7. Add full-registry privacy/build-graph conformance, exact cache/limiter/304/error tests, real
   verified-run third-party dereference/reverification, and authenticated-route non-regression.
8. Reconcile plan/log/canonical docs, construct the exact current-history implementation union,
   obtain the designated cross-party verdict, and only then archive transactionally.
