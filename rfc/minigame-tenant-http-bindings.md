# RFC: Minigame Tenant HTTP Bindings

- **Status:** draft — proposed engineering route; not implementation authority
- **Author:** Codex
- **Created:** 2026-10-13
- **Parent / amends on acceptance:** API Foundation A5/C2/C9/C17; Minigame API MA-C7/MA-C9/MA-C15; Terminal Typer TT-PA4; Deployment DP5 admission binding
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

**Engineering proposal ready for normative review, not accepted:** the descriptor, guard,
metadata/version-floor and transaction specifications are below. Adoption, affected-author body
reconciliation and designated review remain required. RP-460 timing and RP-162 result bounds
remain separate author holds;
this proposal cannot decide them or make admitted unsafe snapshots playable.

## TB1 — Preserve existing dispatch and wire authority

All existing operation IDs, methods, paths, request schemas, response arms/statuses and
historical schema names remain unchanged. Do not grow `MinigameTenantCommand`, replace a
Pitch snapshot ref, refresh an old pin to hide a mismatch, or shadow a valid old URL with a
new handler. Compare generated artifacts and execute existing requests against the router;
metadata equality alone cannot prove dispatch preservation (RP-562).

## TB2 — New operations

All are private-v1, access-token authenticated, no-store, and use the existing account/IP
limits, request IDs and bounded body handling. Drain admission is explicitly specified in
TB2.3; today's `/api/v1/intents` guard does not cover these routes. No client-side URL guessing.

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
TB2.2 enumerates the operation-specific subsets; the existing broad error enum is not
permission to assign a wrong pair/status.

### TB2.1 — Literal named descriptors

Use the existing C17 DSL. All listed object fields are required and all objects closed;
nullable means `oneOf(null, named/type arm)`, not an optional/missing key. `S` is the existing
MAX_SAFE_INTEGER, not a new balance limit. `N(a,b)` means an integer descriptor with both
inclusive bounds. New names do not alias or mutate existing named schemas.

| Name | Exact definition |
|---|---|
| `TyperBeginCommand` | `{assist_level: timed\|untimed, kind: "begin"}` |
| `TyperSubmitLineCommand` | `{kind: "submit_line", text: string}` |
| `TyperEndRunCommand` | `{kind: "end_run"}` |
| `TyperTenantCommand` | oneOf the three command refs |
| `TyperCommandRequest` | `{command: TyperTenantCommand, command_id: opaque-id, expected_revision: N(1,S)}` |
| `TyperSubmissionCleared` | `{first_mismatch_index: null, outcome: "cleared"}` |
| `TyperSubmissionMiss` | `{first_mismatch_index: N(0,S), outcome: "miss"}` |
| `TyperLastSubmission` | oneOf the two feedback refs |
| `TyperSessionActive` | descriptor below with status literal active, plus `snapshot: TyperSnapshot` |
| `TyperSessionClaimed` | same fields, status literal claimed |
| `TyperSessionTerminal` | same fields, status literal resolved, plus `resolution_receipt: MinigameResolutionReceipt` |
| `TyperCommandResponse` | oneOf active/terminal refs |
| `TyperReadResponse` | oneOf active/claimed/terminal refs |

Every Typer session arm has exactly `{constants_hash: sha256-prefixed, engine_ref: "typer",
engine_version: "1.0.0", minigame_id: "typer", mode: "solo", revision: N(1,S),
session_id: uuid-v7, status}` before its named snapshot/receipt fields. Generate those fields
from one constructor parameterized by status; keep old Pitch constructors/refs unchanged.
Create reuses `MinigameCreateRequest`; resolve reuses `MinigameEmptyRequest`.

`TyperSnapshot` has exactly these eighteen fields:

| Fields | Descriptor |
|---|---|
| `typer_content_hash` | sha256-prefixed string |
| `typer_schema_version` | N(1,1) |
| `phase` | ready\|terminal\|typing string enum |
| `era_tier` | N(1,9) |
| `assist_level` | null or timed\|untimed string enum |
| `prompt_index`, `current_prompt_misses`, `lines_cleared`, `clean_lines`, `misses` | N(0,S) |
| `prompts_total`, `revision` | N(1,S) |
| `current_prompt_id` | null or mechanical-id string |
| `current_prompt_text` | null or string |
| `started_server_ms`, `deadline_server_ms`, `last_server_ms` | null or N(0,S) |
| `last_submission` | null or TyperLastSubmission ref |

No HTTP text-length or gameplay counter cap is invented here. Raw text admission and pinned
content-dependent relationships remain TT4/TT-PA1 engine contracts. Shape validation is not a
solution to RP-460: that defect must be reconciled before production activation, not hidden by
silently shrinking budgets. After structural decoding the host also checks session/snapshot
revision equality and terminal receipt session/minigame identity before attaching a child.

The directory has named closed `MinigameDirectoryNone`/`MinigameDirectoryActive` arms and
`MinigameDirectoryResponse = oneOf(their refs)`. Common fields are `constants_hash:
sha256-prefixed`, `kind: none|active` specialized per arm, `minigame_api_schema_version: N(1,2)`
and `tenants: array(MinigameDirectoryTenant)`. Only active adds `session:
MinigameLiveDescriptor`, whose eight existing descriptor fields use status active|claimed,
never resolved. `MinigameDirectoryTenant` has the three existing identity strings/formats
plus `operations: oneOf(null, MinigamePublicRoles)`. `MinigamePublicRoles` is a closed union
of the compiled Pitch/Typer four-role objects, each field a literal operation ID from TB3.
No new `operation-id` format, open map or custom DSL kind is needed. Association/order and
hash/tuple checks belong to the pinned binding consumer, not a fabricated schema keyword.

### TB2.2 — Status and error subsets

Success is200: directory → directory response; create → active; read → read response;
command → command response; resolve → terminal. Errors use the existing exact two-field
APIError, with the exception of the existing drain category's separate new descriptor below.
Literal bytes retain the repository's `{"category":...,"detail":...}\n` encoding. Construct
sorted ExactJSON rows from the following pairs, not a handwritten client table.

All five operations declare401 unauthorized/access_token,429 rate_limited/account or /ip,
500 internal_invariant/minigame_api, and503 not_configured/minigame_api.

| Operation | 400 invalid detail | 404 unknown_id details | 409 category/detail pairs |
|---|---|---|---|
| directory | body | founder | none |
| create | minigame_create | founder, minigame_tenant, minigame_session | conflict/minigame_session; idempotency_conflict/minigame_session; not_eligible/exclusive_activity, /tier_required, /curriculum_exit_required |
| read | body | founder, minigame_session, minigame_tenant | none |
| command | minigame_command | founder, minigame_session, minigame_tenant | conflict/minigame_session, /minigame_revision; idempotency_conflict/minigame_command; not_eligible/illegal_phase, /invalid_assist_level, /invalid_text, /line_too_long |
| resolve | body | founder, minigame_session, minigame_tenant | conflict/minigame_session |

Create additionally declares500 internal_invariant/session_id for server ID generation.
Typer does not acquire Pitch card/hack/Fiscal/Soul rejections. Structural admission failures
are400; a valid request encountering malformed internal bundle/state/response is500, not a
guessed invalid-request error. Store/unlisted errors likewise remain500. Read of a claimed
session is200, not409; resolve of a nonterminal session is409, never implicit resolution.

### TB2.3 — Registered drain admission

Current `gameserver.Server.intentAdmission` only gates POST `/api/v1/intents`. Add explicit
composition bindings for the new create/command operation IDs to that same gate: enter before
coordinator mutation, leave only when the admitted request finishes, and refuse new admission
after drain begins. Do not infer mutation from the HTTP method: directory/read and the read-only
resolve operation must not acquire an intent permit. Preserve the existing main-intent route
and old Pitch wire; do not claim this amendment audits every other mutation route.

For the two new mutators only,503 is `TyperUnavailableResponse = oneOf(APIError,
MinigameDrainError)`, with exact literals narrowed to the ordinary not-configured pair and
`{category:"server_draining",detail:"retry_same_intent_id"}`. The latter already exists in the
main-intent handler; define its two literal fields in a separate new named descriptor rather
than widening the old shared APIError enum or changing old operation replies. This amends
TT-PA4's endpoint proposal explicitly; it does not invent a new gameplay rejection.

The generated guard recognizes the drain arm separately from APIError. The host retains the
same create key/command ID for an explicit retry and never automatically resubmits it or
refreshes credentials. Reuse existing restart/error presentation, with no authored copy change.
DP5's bound/courtesy frame are unchanged. Acceptance must exercise a real admitted blocked
request, drain, refused second request and completion/flush of the first; verify no coordinator
call or row change for the refused request. A unit test of the old main-intent route alone does
not prove this binding.

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

Schema2 bindings are exactly:

| Tuple | create | read | command | resolve |
|---|---|---|---|---|
| pitch / pitch / 1.0.0 | create_minigame_session | get_current_minigame_session | play_minigame_command | resolve_minigame_session |
| typer / typer / 1.0.0 | create_typer_session | get_typer_session | play_typer_command | resolve_typer_session |

This table is the proposal's literal contract, implemented once as compiled Go binding rows;
directory schema role literals, operation association and generated client guards consume those
rows. A role must match the registered method/path/auth/request/success contracts, not merely an
existing ID. Cross-family/missing/extra/repeated role IDs refuse. Startup rejects schema2 tenants
without complete server/client bindings; source1 backend-only fixtures remain unpromoted.

Both loaders retain `schemaVersion: 1|2`, operation rows and tenant identities. Schema1's public
normalization returns Pitch's exact legacy roles and null for non-Pitch, without modifying its
artifact bytes or historical replay policy. Schema2 tenant roles are required, non-null, closed
four-field objects; operation IDs use `^[a-z][a-z0-9_]*$`, engine versions retain the catalog's
positive-major grammar, arrays are byte-sorted/unique and root operations are the exact role
union plus directory. Refuse unknown/aliased/missing/duplicate keys and noninteger version
tokens before ordinary parsing. Current TS's generic JSON.parse cannot establish raw
duplicate/numeric-token admission; Go already rejects those but its struct decoder does not
by itself prove exact key spelling. Close both sides' raw/shape conformance (RP-565), not a
re-encoded legacy artifact. Preserve canonical historical artifacts and test noncanonical refusals.

Go uses private owned catalog slices and value/copy-returning schema/operations/tenant/roles
accessors; the TS model deep-freezes the corresponding arrays/rows/role objects. Replay checks
structural membership without importing an HTTP registry, rendering a child or dispatching an
operation. Composition alone checks the compiled protocol associations. Both versions retain
the existing minigame_api presence → Founder21 floor, subject to higher active-feature floors;
no Company/Founder field, SQL migration or replay-envelope version is added solely for roles.

Preserve the existing Pitch/content dependency under MA-C15 and the complete Typer chain under
TT-PA3. This is not an implicit ruling on Arcade's Pitch-less D-020. Do not mint schema2 into
production until its accepted loader, bindings, adopted content and full gates are complete.
The actual watched loader/replay change must have a genuine kernel bump and mixed-history
tests in the same implementing batch; API-only scaffolding alone cannot claim a false bump.

## TB4 — Atomic namespace and stored-read binding

The binding is server configuration, not a request field: an immutable compiled entry keyed
`(minigame_id,engine_ref,engine_version)`, with its four operation IDs and admitted modes.
Composition checks it against the registry and pinned metadata; handlers select it by registered
operation ID. Public callers cannot supply a tuple, hash, Founder or binding callback. Historical
sessions resolve their own immutable hash, not the deployment-current bundle or current Company's
newer hash. Existing request hashes and idempotency scopes remain unchanged.

### TB4.1 — Create and create replay

Add a guarded variant of `save.Store.ApplyMinigameStartTransaction` with a required
`MinigameStartAdmission` pair:

- `New(ctx,tx,founder,founderRevision,company,companyRevision) error` receives the existing
  locked states/revisions. Production resolves the definition and role IDs from that exact
  Company hash and checks the selected compiled create binding. Run this before the first
  genesis/log/sequence/session write, not just inside the existing mutation callback: that
  callback currently runs after possible Founder-genesis insertion.
- `Replay(ctx,tx,storedSessionID) error` runs on an equal-hash create receipt before returning
  it. Select `session_id` along with the existing stored hash/response. The callback loads the
  session with the same transaction and authenticated Founder predicate, then checks its own
  pinned tuple/binding. It does not substitute the newly generated request session ID or require
  that an old completed session belong to the current run/hash. A mismatched create-key hash
  still returns the existing idempotency conflict; a key reused for a different minigame is
  not silently treated as a retry.

The public guarded entry also carries the authenticated account ID. Its existing Founder-row
lock must require `account_id` equality and `archived_at IS NULL`; the adapter's earlier
`ActiveCompanyState` read is only a locator. Preserve Founder → Founder stream → Company stream
→ session ordering. Both callbacks are read-only and mandatory for public bindings; an internal
backend start may retain the existing unguarded entry, but no mounted handler may call it.
Namespace refusal is before writes, not a write followed by output validation and rollback.

New-session response construction must check the owning operation's generated descriptor before
receipt insertion/commit. That validation is provided by composition at the receipt-builder
boundary, not a second schema in production. Preserve existing sequence/seed/unlock/Soul/genesis
construction, sampling and the Founder replay transition. No migration or replay field is added
merely to remember which HTTP route invoked the existing command.

### TB4.2 — Command admission, replay and claim

Add `minigame.Repository.ClaimOrReplayAPI` and a corresponding guarded service path, taking
the authenticated account/Founder, session ID, command ID/hash and a required read-only
`SessionAdmission` callback supplied by composition. One transaction must:

1. Lock the active Founder row with its account predicate, as above.
2. Load the requested session with `session_id AND founder_id`, locking that owned session.
   Run admission against its immutable tuple/hash and pinned binding before looking at its
   receipt hash, status, revision or acquiring a claim. Missing/wrong-family sessions both
   return `ErrSessionGone` and the same404 bytes. Never load globally and filter in Go.
3. Look up the existing command receipt under those locks. Equal hash returns the existing
   canonical response without tenant execution, revision checks, clock sampling or claim
   changes; unequal hash returns the existing command-idempotency conflict. This includes
   resolved sessions. The old unguarded resolved/status fast path must not bypass replay.
4. Only if no receipt exists, use the existing active/expired-claim predicate and lease to
   acquire the claim, and take its one persisted command-clock sample before committing.
   Admission failure must not update `claimed_at`, `updated_at`, token, state or revision.

The return is a closed internal result: stored receipt **or** newly claimed Session. It is not
an extra public JSON union. Preserve existing in-flight busy and stale-revision behavior;
do not wait, poll or invent a second receipt/lease system. Pure tenant execution remains outside
the claim transaction. Existing token-guarded nonterminal and terminal transactions remain the
only completion owners; their receipt builders validate the selected operation DTO before any
response receipt commits. Terminal payout and its command response still commit together.
The certification's frozen tuple cannot be replaced by a newer bundle during resolution.

### TB4.3 — Read and directory

Add transaction-taking `minigame.Repository.LoadTx` and `CurrentTx` variants using the same
Founder-scoped predicates as their existing reads. Composition opens a read-only repeatable-read
transaction, resolves the account's active Founder and Company in it, and reads the session and
pinned metadata in that same snapshot. No row claim, session-clock sample, evaluation, payout,
token lease extension or repair is performed. Missing Founder stays404; no session is `none`
only on the directory path. A session-ID read must check family admission before exposing its
status or snapshot. Claimed-session reads expose only the persisted snapshot and public status,
not the in-flight worker's candidate result or claim details.

The session stores an immutable **inner resolution receipt**, not the entire terminal command
HTTP response. Reconstruct the read envelope through `MinigameAPISessionReceipt` from the
persisted terminal Session plus that stored inner receipt, then validate its read descriptor.
Never rerun certification/payout or manufacture a fresh inner receipt. Command replay returns
its own stored full response instead of using this reconstruction. Compare a terminal read
against the first committed terminal command envelope in the real-Postgres acceptance population.

### TB4.4 — Required failure evidence

Real Postgres must show no writes for wrong-family create/command/read/resolve, including an
existing opposite-family command receipt; missing and other-Founder IDs return the same absence.
An equal create retry after Company run/hash change must bind the stored session, not mint a new
one. An account/Founder change paused between adapter lookup and transaction admission must not
execute under the old identity. Concurrent same-ID commands must return the recorded response
after completion without another tenant application. Claimed/terminal reads must leave session,
receipt, stream/log/event and faucet rows unchanged. Inspect timestamps/tokens as well as revision
counts: a claim followed by release is not a no-write refusal. Run in the declared service;
source inspection or a mocked transaction is not evidence that these guarantees hold.

## TB5 — Generated consumer and mounted child

Generate schemas/types/dispatch from the Go registry; no handwritten second public schema or
open-JSON escape. Bind both Pitch and Typer adapters by pinned tuple/role IDs.
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

### TB5.1 — Generated guards, not TS casts

Add `GenerateTypeScriptGuards(registry, operationIDs)` for the four old Minigame operations and
five new operations. Emit `client/src/api/generated/minigame-guards.ts` from their complete
transitive named-descriptor closure and ExactJSON response literals. Generation refuses an
unsupported kind/format or an integer descriptor without both bounds within ±S; it must not
pretend to validate every arbitrary int64 API DTO. The existing C17 kind/ref/depth/cycle rules
remain unchanged. No parallel handwritten field list, inferred struct reflection or schema file.

The generated module exports a closed GuardedMinigameOperationID and
`decodeMinigameResponseBytes(operationID,status,contentType,bytes)`. Consume schema bytes once;
decode UTF-8 fatally while preserving a leading BOM (`ignoreBOM:true`), rather than Response.text
silently replacing bytes or discarding that prefix. Guarded Go response composition checks
UTF8 validity before the existing registry validator. Validate raw text before JSON.parse can
erase duplicate keys or numeric spelling, reusing parseIntegerCatalogJSON for this integer-only
closure. Strings carrying canonical Decimals remain strings and use the
existing parseCanonical validator; emit existing format patterns from their Go authority and
require full-string matches, not JS `$`'s before-final-newline shortcut. Exact objects require
own fields, no extras, precise required/null distinctions and no coercion. Integer bounds are
inclusive, oneOf means exactly one matching arm, refs preserve their declared depth. Error
ExactJSON checks compare original bytes to UTF-8 encoded registered literals, not reserialized
pairs or a BOM-stripped text value.
Diagnostics contain operation/status and fixed schema paths, never reply values or unknown keys.

Extend the generated transport with an optional schema-byte decoder, invoked after existing
status/media checks and before a typed result returns. `createGuardedMinigameAPIClient` supplies
that generated decoder and restricts dispatch to its generated ID set. Existing two-argument
createAPIClient behavior and raw evidence bytes/hash checking are unchanged. Consume a response
body only once, propagate abort/read failures, and add no token storage/refresh/retry/timeouts.
RP-564's semantic error-pair filter is not this raw/structural guard and must not be cited as it.

Ports use generated input/output types and distinguish the exact drain arm from APIError. The
host then verifies the pinned tuple/hash, role map, session/receipt/revision and captured owner
context before child attachment. Shape checks do not replace those associations or engine
content semantics. Shared Go→TS raw vectors must include every snapshot field's absence/null/
type/range, feedback arms, ambiguous unions, duplicate/escaped keys, number-token aliases,
wrong operation/status/error literals, malformed UTF-8/BOMs, and valid ready/typing/claimed/
terminal specimens. Run
Node and supported real browser engines; zero-execution launch failures are not coverage.

## Acceptance and staged construction

1. Review this descriptor/source/schema2/guard/transaction proposal; reconcile affected
   TT-PA4/MA-C9/MA-C15/DP5 bodies and obtain adoption/designated normative review.
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
