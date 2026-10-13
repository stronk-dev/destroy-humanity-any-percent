# RFC: Minigame Tenant HTTP Bindings

- **Status:** draft — proposed engineering route; not implementation authority
- **Author:** Codex
- **Created:** 2026-10-13
- **Parent / amends on acceptance:** API Foundation A5/C2/C9/C17; Minigame API MA-C7/MA-C9/MA-C15; Terminal Typer TT-PA4
- **Design refs:** `design/03-minigames.md §7`, `design/06-tech.md`, `design/00-vision.md`
- **Source coordinate:** `6a5a742a`
- **Planning:** `planning/minigame-tenant-http-bindings/`

## Outcome

Let a player discover, resume and play Terminal Typer through the authenticated public
HTTP surface without changing the pinned Pitch request/response schemas. Recommend new
tenant-specific operations under a **disjoint** prefix, not re-pinning the old union or
mounting a static route inside its existing parameter domain. This is a proposal to close
the recorded TT-PA4/C2 hold, not adoption of it or completion of Typer.

The complete chain is pinned declaration → registered typed HTTP binding → authenticated
coordinator → persisted session/receipt → generated client → tenant child → player proof.
Adding routes alone does not complete that chain. No engine, scoring, timing, payout, rank,
assist-level, retry, content or save mechanic is changed by this contract.

## Evidence and readiness

The current registry pins generic create/command/resolve/current operations, with a Pitch-only
snapshot ref and command union. TT-PA4's requested additions violate C2. The current
`minigame_api` artifact separately requires exactly those four operation IDs, so new routes
must not silently reinterpret its v1 metadata. The current client registry also uses a
deploy-current imported artifact; it is not a general historical-session binding source.

The actual registry, compatibility checker and chi mounting were exercised in six mechanical
cases. A new static POST `/api/v1/minigames/typer/sessions` passes compatibility but replaces
the legacy generic operation's reply for the admitted `minigame_id=typer`, in both operation-ID
orders. Disjoint `/api/v1/typer/sessions` and different-method controls preserve that reply.
Manual source: `planning/api-foundation/probes/tenant-route-probe.go.txt`. This is a registered
stub-handler/routing experiment, not a real Typer HTTP, authentication or database run.

**Not acceptance-ready:** finish the literal C17 descriptor/guard specification and atomic
namespace-admission seam below, reconcile the affected authors' bodies, and obtain adoption
and designated review. RP-460 timing and RP-162 result bounds remain separate author holds;
this proposal cannot decide them or make admitted unsafe snapshots playable.

## TB1 — Preserve existing dispatch and wire authority

All existing operation IDs, methods, paths, request schemas, response arms/statuses and
historical schema names remain unchanged. Do not grow `MinigameTenantCommand`, replace a
Pitch snapshot ref, refresh an old pin to hide a mismatch, or shadow a valid old URL with a
new handler. Compare generated artifacts and execute existing requests against the router;
metadata equality alone cannot prove dispatch preservation (RP-562).

## TB2 — New operations

All are private-v1, access-token authenticated, no-store, and use the existing account/IP
limits, request IDs, drain refusal and bounded body handling. No client-side URL guessing.

| Operation ID | Method/path | Request | Success |
|---|---|---|---|
| `get_minigame_directory` | GET `/api/v1/minigames/active-directory` | no body | directory below |
| `create_typer_session` | POST `/api/v1/typer/sessions` | existing exact `{idempotency_key}` | Typer active session |
| `get_typer_session` | GET `/api/v1/typer/sessions/{session_id}` | no body | Typer live or terminal session |
| `play_typer_command` | POST `/api/v1/typer/sessions/{session_id}/commands` | `{command_id,expected_revision,command}` | Typer active or terminal session |
| `resolve_typer_session` | POST `/api/v1/typer/sessions/{session_id}/resolve` | exact `{}` | Typer terminal session |

Opaque IDs reuse `[A-Za-z0-9-]{1,64}`, session IDs lowercase UUIDv7 and revisions1…MAX_SAFE_INTEGER.
Commands reuse TT4: `{kind:"begin",assist_level:"timed"|"untimed"}`, `{kind:"submit_line",text}`,
or `{kind:"end_run"}`. No score, result, clock, prompt ID, engine, mode, Founder or Company
authority enters the request. Shape-invalid requests refuse before tenant execution; pinned
text rules keep their existing `invalid_text`/`line_too_long` taxonomy.

Typer session replies use the existing eight descriptor fields plus `snapshot`, with literal
`minigame_id:"typer"`, `engine_ref:"typer"`, `engine_version:"1.0.0"`, `mode:"solo"`.
Writes return `status:"active"` or `status:"resolved"`; resolved adds only the existing exact
`resolution_receipt`. Read accepts `active`/`claimed` or `resolved`. A claimed read presents
busy/retry/Leave with existing copy, not enabled gameplay or a guessed claim lease. Reads never
claim, advance time, execute a command, resolve, pay out or repair a session.

Snapshot is TT4.5's exact eighteen fields, not a new envelope. Its nullable fields and exact
`last_submission` are retained. Numeric representation is the existing safe-integer wire;
era tier1…9 and schema version1. Content-dependent relationships and RP-460/RP-162 need the
engine's authored contracts, not arbitrary HTTP caps. No seed, command log, future prompt,
claim token, worker/recovery identity or private faucet state is exposed.

Keep the existing error categories/details and exact response-byte authority:400 malformed
create/command uses `invalid/minigame_create` or `invalid/minigame_command`; malformed read/
resolve uses `invalid/body`;401 `unauthorized/access_token`;404 owned-resource absence or
wrong family uses `unknown_id/minigame_session` (unknown configured tenant may use
`unknown_id/minigame_tenant`);409 uses the existing revision/session/idempotency/unlock/tenant
pairs;429 account/IP;500 existing invariant pairs;503 `not_configured/minigame_api`.
Enumerate the operation-specific allowed subsets in C17 descriptors before acceptance;
the existing broad error enum is not permission to assign a wrong pair/status.

## TB3 — Directory and pinned operation metadata

Directory exact common fields: `constants_hash`, `minigame_api_schema_version`, `tenants`,
`kind`. `kind:"none"` has no session; `kind:"active"` adds the existing descriptor for the
authenticated active Founder's active/claimed session, never a snapshot or resolved history.
No active Founder is404, not fabricated `none`. Pin data comes from that Company's immutable
bundle; session identity must agree with its pinned declaration. Do not read current files.

Each directory tenant has exact `{engine_ref,engine_version,minigame_id,operations}`;
operations is null for a declared but publicly unbound historical fixture, otherwise exact
`{create,read,command,resolve}` operation-ID strings. Never emit an arbitrary URL. IDs must
match the compiled registry/adapter; the client dispatches only generated known operations.

Propose `minigame_api` **schema2**, preserving schema1 unchanged:

- Exact root `{schema_version:2,operations,tenants}`.
- Root operations retain exact `{operation_id,version:1}` rows, sorted/unique. Their set is
  precisely the union of each tenant's four role IDs plus `get_minigame_directory`.
- Each tenant adds required exact `operations:{create,read,command,resolve}` to its existing
  three identity fields; tenant ordering/identity uniqueness stay unchanged.
- Role IDs must exist in the declared operation set and match the compiled binding for that
  tuple. Cross-tenant, wrong-method/schema/auth or missing adapter references refuse startup/
  activation. No user-supplied metadata determines dispatch or engine execution.
- Schema1 keeps its four fixed IDs. Pitch normalizes to those legacy bindings; schema1
  non-Pitch fixtures remain backend-only and are not presented as public binding support.
- Pure replay loaders validate shape/references without executing HTTP; composition checks
  the compiled HTTP contract. No applied SQL migration or historical artifact is rewritten.

Catalog schema2 does not imply a new gameplay/save field or automatic mint. Define its exact
Go/TS representation, loader/version-floor compatibility and activation contract before
acceptance. A genuine accepted loader/integration change registers/bump-checks the affected
kernel paths; do not mint a false version signal for API-only metadata scaffolding.

## TB4 — Atomic namespace and stored-read binding

Before creation or command mutation, under the same scope/transaction authority as the
coordinator, prove the selected session/definition tuple has this namespace's declared HTTP
binding. A preflight lookup followed by an unchecked transaction is insufficient across
Founder/Company/epoch changes. Legacy operations likewise cannot create an unsupported tenant
and only discover the wrong response shape after commit. Wrong-family requests produce the
same404 absence without claims, rows, sequences, receipts, payouts or gameplay mutation.

Reuse the existing create sequencing, request hashes, command receipts, sampled persisted DB
clock, certification, once-only terminal composition and claim ownership. Stored retry bytes
must not be reserialized with a different namespace or fresh sample. Lookup by session ID is
authenticated-Founder scoped; resolved reads use the exact stored terminal receipt. No global
lookup followed by caller-side ownership filtering. The concrete transaction seam/interface
and idempotency lookup ordering need to be named before acceptance, not improvised while coding.

## TB5 — Generated consumer and mounted child

Generate schemas/types/dispatch from the Go registry. Finish the exact Typer descriptors and
the client guard mechanism from that same authority before acceptance; no handwritten second
public schema or open-JSON escape. Bind both Pitch and Typer adapters by pinned tuple/role IDs.
The supported compiled child inventory is distinct from the registered children of one bundle:
every active declared tenant has exactly one matching child, while an unactivated compiled
adapter is not a fabricated catalog row. Reconcile MA-C9's current global-map wording.

Directory → selected role read must bind session ID/tuple/hash/status to the captured Founder/
Company context. Changed or failed reads never attach a mismatched snapshot, fabricate `none`,
retry a mutation, replace credentials or create a session. Preserve the host's explicit read/
error/re-consent and same-ID retry rules; no new polling or timing budget. A claimed session
keeps a reachable explicit read Retry and Leave. Unsupported bindings fail visibly, not as a
Pitch fallback. TyperTable receives the trusted host server-time estimate under TT8, never a
client-submitted clock. Current availability must reflect the actual TT-PA2 tier/Exit predicates
through their owning derivation, not merely existence of a compiled child.

No production promotion without adopted prompts/copy, parent content gates and a jointly
supported server/client binding for every active tenant. Preserve Pitch's existing player path.

## Acceptance and staged construction

1. Complete descriptors, transaction seam, source/schema2 compatibility and client-guard
   specification; reconcile TT-PA4/MA-C9 and obtain adoption/designated normative review.
2. Implement metadata/codecs and generated guards; exact shared success/refusal vectors,
   immutable old artifacts and a genuine kernel change where applicable. No public promotion.
3. Add disjoint registered routes and atomic binding. Real Postgres must prove before-mutation
   refusal, other-Founder concealment, read-only active/claimed/resolved reads, application
   reconstruction, same-ID retry and exactly-once certification/payout.
4. Generated client/native child: complete timed and untimed play by DOM controls, current/
   directory races, authoritative refresh, terminal/reload, reply-loss/reconstruction and
   keyboard/paste/composition/reflow behavior. Setup and natural-pacing evidence stay distinct.
5. Compare original v1 pins and execute old operation URLs/DTOs with the new router; preserve
   old schemas/dispatch, and demonstrate static-shadow and wrong-family negatives. Pin checks
   alone do not close this gate. No command/snapshot union widening on an old operation.
6. Resolve RP-460/RP-162 and all parent acceptance/content/AT gates, then complete the full
   API/Typer review ranges and docs/ledgers. No fixture-only or route-only archival/release.

This is not authority for Arcade, Pinball, live PvP or other unaccepted tenants. Their owners
may consume the pattern through exact subsequent contracts. It does not narrow nine-tier1.0,
change rights/retention or deployment policy, approve earlier work, publish or deploy anything.
