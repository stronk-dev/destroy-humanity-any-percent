# RFC: Minigame & Recovery API + Surface (the playability seam)

- **Status:** accepted — MA-C1–C9 ruled (2026-08-07); implementation-ready. The named successor
  that TP-C10 (The Pitch) and SR-C1/SR-C3 (Soul Recovery) both depend on for HUMAN playability;
  **every later minigame and coordinator activity inherits this boundary.**
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** `design/06 §transport` (authenticated command surface), `design/11` (first-session
  UX — the onboarding dossier's MUST/MUST-NOT list applies to these surfaces)
- **Depends on:** API Foundation (accepted, implementing — this AMENDS its authenticated `/api/v1/`
  surface, additive-only per its A1 policy), Minigame Platform (accepted, implementing — its RFC is
  NOT yet archived; the internal create/play/resolve services this exposes are implemented and
  reviewed through The Pitch's designated verdicts), Soul Foundation (archived — the SB19/SB24
  coordinator this exposes), UI Foundation (accepted — the component/token substrate the surfaces
  render on), Game-UI Screens (draft — consumes the surface contracts defined here).
- **Planning:** `planning/minigame-api-and-surface/` (once implementing)

## Summary

Two narrow, authenticated, additive `/api/v1/` surface families exposing machinery that is already
implemented, reviewed, and internal-only: **minigame sessions** (create → command → resolve for
platform tenants, The Pitch first) and the **soul-recovery coordinator** (start/reconnect →
progress heartbeat → cancel/resolve, per SB25–SB27). Plus one UI **surface contract** per family —
the component-boundary shape (inputs, callbacks, lifecycle) that Game-UI screens consume — so
"playable by a human" becomes a defined seam instead of an assumption. No new game mechanics, no
new persistence: this RFC is transport + surface only.

## Motivation

Wave-B's first two content RFCs both bounced on the same missing seam: the platform and coordinator
are internal Go services with no authenticated route, and Game-UI explicitly excludes minigame
surfaces. This RFC owns that seam once, as an API-Foundation amendment, so content RFCs never have
to. Out of scope: the public read API (`/api/public/v1/` — untouched), spectation, any WebSocket
streaming for minigames (bounded commands over HTTP per the established intents pattern; a
streaming variant is a declared successor if a real-time tenant ever needs it), and the Game-UI
screens themselves (that RFC consumes these contracts).

## Specification

### MA1 — The minigame session API (authenticated `/api/v1/minigames/`)

Additive endpoints, exact request/response schemas generated per the API Foundation's schema
authority (A2/A5):
- `POST /api/v1/minigames/{minigame_id}/sessions` — create. Founder identity from the
  authenticated session ONLY (no founder field in the body). Validates unlock (incl. the
  `fiscal_unlock` resolver arm, TP-C6), soul_gate, and the platform's session rules. Response: the
  session descriptor + the engine's initial snapshot.
- `POST /api/v1/minigames/sessions/{session_id}/commands` — one engine command per request (the
  tenant's closed command union, e.g. The Pitch's `play_hand | buy_hack | end_shop`, TP-C12),
  idempotent by the platform's existing intent-record discipline; response: the
  post-command snapshot + revision. Typed rejections reuse the closed taxonomy verbatim.
- `POST /api/v1/minigames/sessions/{session_id}/resolve` — triggers the certified resolve through
  the shipped composer; response: the terminal receipt (payout, quality, facts). Retry returns the
  stored receipt (the platform's idempotency, exposed not reinvented).
- `GET /api/v1/minigames/sessions/current` — reconnect: the founder's live session descriptor +
  snapshot, or 404-shaped empty (typed, not an error).
Rate limits per the API Foundation's operational middleware (C16/C20 literals); commands
additionally per-session budgeted (catalog data).

### MA2 — The recovery coordinator API (authenticated `/api/v1/soul-recovery/`)

The SB25–SB27 schemas, verbatim — this RFC adds NO fields:
- `POST /api/v1/soul-recovery/start {activity_id}` — start/reconnect (SB25: returns the existing session with a
  ROTATED progress token when one is active). Exact 7-key response.
- `POST /api/v1/soul-recovery/progress {session_id, progress_token}` — the heartbeat. Exact 5-key response;
  per-session rate limit derived from the catalog beat cadence (SR-C6: cadence = ceiling/3; the
  limiter allows modest jitter above it, rejects flooding).
- `POST /api/v1/soul-recovery/cancel {session_id}` and `/resolve {session_id}` — terminal; retry
  returns the stored receipt. (Route shapes per the SR-C10 ruling — flat command routes, the
  intents pattern.)
Typed rejections: `recovery_token`, `exclusive_activity`, `session_expired`,
`idempotency_conflict` — surfaced exactly as ruled, never exposing claim-lease internals.

### MA3 — The surface contracts (what Game-UI screens consume)

Two component-boundary contracts on the UI Foundation's token/component substrate:
- **`minigame_session` surface:** inputs = the session descriptor + snapshot + the tenant's
  command vocabulary; callbacks = command dispatch + resolve + exit-to-host; lifecycle = mount on
  create/reconnect, snapshot-driven re-render (no client simulation — the server snapshot is the
  only truth), terminal rendering from the receipt. Tenant-specific presentation (The Pitch's card
  table) plugs in as a child component keyed by `engine_ref`; the surface owns the chrome, errors,
  and lifecycle so tenants never touch transport.
- **`soul_recovery` surface (SR-C3, restated here as the owning contract):** inputs = session
  receipt + local toy seed; callbacks = the coordinator commands; one heartbeat scheduler
  (visible-only, cadence per SR-C6, missed ceiling = pause, reconnect-then-resume, never replay
  queued beats); cancel affordance; progress display; terminal return. Toy components receive
  presentation data/callbacks only.
Both surfaces meet the UI Foundation's axe-core/keyboard gates; both follow the onboarding
dossier's MUST/MUST-NOT screen rules where applicable.

### MA4 — Authority & privacy invariants

Founder identity exclusively from the authenticated session; no endpoint accepts a founder,
company, or clock coordinate from the client (the established law). The privacy assertion extends:
an integration test enumerates these endpoints' responses against the same no-hidden-info discipline
as the public API test. Nothing here widens replay surfaces — every mutation already flows through
the shipped, reviewed composers.

## Deviations from design

None — this exposes ruled machinery over the established API discipline.

## Acceptance criteria

1. A human-shaped client (integration test using only these endpoints) plays a full Pitch run:
   create → commands → resolve → receipt, with reconnect mid-run; and completes a full recovery
   session: start → heartbeats → resolve, with token rotation on reconnect and the watchdog path.
2. All schemas generated from the single authority; additive-only against `/api/v1/`; typed
   rejections byte-match the closed taxonomy.
3. Rate limiting proven: command flooding and heartbeat flooding rejected without state mutation.
4. The privacy enumeration test passes for every new endpoint.
5. Both surface contracts implemented as components passing the UI Foundation gates, with the
   tenant child-component seam proven by The Pitch's table mounting into `minigame_session`.

## Open questions

- ~~Whether `GET .../sessions/current` should cover minigame AND recovery in one call~~ — **RULED
  2026-08-07 (owner-side, pre-acceptance): no composite endpoint in v1.** Each family owns its own
  reconnect read: minigames via `GET .../sessions/current` as specified; recovery via SB25's
  idempotent `start` (which returns the active session with a rotated token) invoked only on user
  intent, with passive display derived from the founder state the client already receives. If
  acceptance review finds the reconnect screen needs a passive recovery read the founder snapshot
  doesn't carry, the remedy is a per-family `GET /api/v1/soul-recovery/current`, NOT a composite —
  a cross-family composite would be a second source of truth over two independently-owned session
  stores.
- The streaming successor's trigger condition (first real-time tenant) — declared, not designed.

## Codex acceptance review blockers (2026-08-07 — MA-C1–MA-C9)

The intended boundary is correct, but the draft describes several authorities that do not exist and
one recovery API that already does. Implementing past these points would either duplicate reviewed
routes or invent session semantics at the public boundary.

### MA-C1 — MA2 is already implemented, while its stated error surface contradicts the archive

Soul Recovery shipped the four authenticated routes in `server/account/api.go`, including the
session limiter and integration coverage. MA2 still says this RFC “adds” them and lists
`session_expired`; the designated Soul verdict canonically ruled that expiry is a watchdog terminal
receipt and never an API error. The actual handlers also expose exact pairs such as
`unknown_id/recovery_session`, `not_eligible/exclusive_activity`,
`conflict/recovery_session`, and `idempotency_conflict/recovery_session`.

**Proposed contract:** treat MA2 as an API-Foundation registration/generation pass over the existing
reviewed routes, not a second implementation. Export their exact request/response/error schemas from
the single operation registry, generate the client types, and retain the archived runtime behavior
byte-for-byte. Remove `session_expired` from the public taxonomy. Any runtime change discovered by
schema conformance returns to the Soul Recovery owner rather than being hidden in this RFC.

### MA-C2 — The composed gameserver has no minigame platform

`gameserver.Compose` constructs no `minigame.Repository`, tenant registry, content resolver, or
service; registers no Pitch tenant; and exposes no platform on `Composition`. The account API has no
minigame handler attachment. The internal services proven by tests are therefore unreachable from
the production binary.

**Proposed contract:** this RFC owns the previously parked Platform AC1 composition slice. Compose
the repository, a closed tenant registry containing Pitch `1.0.0`, a pinned
`TenantContentResolver`, and `minigame.Service`; attach one minigame coordinator adapter to the
authenticated API; expose the platform on `Composition`; and include it in bounded drain only if it
owns background work (otherwise explicitly prove none). A real-socket composed-binary integration
test—not a package-local handler fixture—drives the new endpoints.

### MA-C3 — Create cannot derive the server-owned `StartRequest`

The public request supplies only `minigame_id`, while `StartMinigameSession` currently requires a
fully populated server-internal request: session ID, Founder/Company/run identity, engine/version,
constants hash, mode, scaling inputs, and seed. No composed resolver constructs those values. The
platform contract says the seed comes from the save stream, but there is no persisted per-session
ordinal or exact derivation for repeated sessions in one run.

**Proposed contract:** add one server-only start resolver. It loads the authenticated active
Company/Founder and pinned definition, derives engine/version/hash/unlock/Soul/scaling without
client fields, creates the UUIDv7 session ID, and supplies an exact server-authored seed contract.
Owner must choose the missing coordinate: recommended is a persisted monotonic
`minigame_session_seq` per Founder/run and
`Substream(runidentity.Seed(founder_id, run_seq) xor uint64(seq), "minigame.session.v1").Next()`;
if no new state is desired, an opaque random uint64 frozen in the session is valid but explicitly
amends the platform's “save-seeded” wording. V1 mode is `solo` from the Pitch definition; the client
does not submit scaling, engine, hash, clock, or identity.

### MA-C4 — The separate resolve endpoint cannot reconstruct a terminal resolution

`minigame.Service.Play` returns a `CertifiedResolution` only in memory and deliberately leaves the
session claim-owned when a command is terminal. `ResolveMinigameSession` requires that object. If
the command handler returns to the client and expects a later `/resolve` call, there is no public or
persistent operation that can reconstruct the capability; a lost response can strand the claimed
session.

**Proposed contract (recommended):** the command endpoint detects a terminal `PlayDecision` and
immediately calls the production resolution composer in the same request, returning the stored
terminal receipt. `/resolve` becomes a retry/read of an already terminal session receipt (and
rejects an active session), never a client-triggered certification step. Alternatively persist a
reconstructible certified-resolution capability under a new reviewed platform contract; do not
smuggle the in-memory object across HTTP requests.

### MA-C5 — “Existing idempotency” is false for create and nonterminal commands

Session commands are sequenced by expected revision, but carry no client command ID or request hash.
A retry after a committed response loss receives `ErrSessionRevision`, not the stored prior
response. Create generates a fresh server session ID and likewise has no retry identity. Only
terminal resolution currently has a durable stored receipt.

**Proposed contract:** add API command IDs with the standard canonical-request SHA-256 discipline
and a unique `(session_id, command_id)` receipt row, claim-token-guarded in the same transition;
same ID+hash returns the stored snapshot/receipt, same ID+different hash rejects
`idempotency_conflict`. Create uses a client opaque idempotency key scoped to the authenticated
Founder and returns the same session descriptor on retry. If owner rejects new persistence, amend
MA1/AC1 honestly to at-most-once commands plus reconnect recovery; do not call revision conflicts
idempotency.

### MA-C6 — `GET .../current` has no repository query or exact empty shape

The repository can load only by `(founder_id, session_id)`. No “current by Founder” query exists,
and “404-shaped empty (typed, not an error)” is not a wire contract. Resolved-session inclusion and
the interaction with an active recovery session are unstated.

**Proposed contract:** add a Founder-scoped current-session query returning the sole
`active|claimed` minigame session under the existing exclusivity invariant. Response is the closed
union `{kind:"none"}` or `{kind:"active",session:<descriptor>,snapshot:<exact JSON>}` with HTTP 200;
resolved receipts are obtained only through the terminal command/resolve retry path. Do not expose
claim tokens or return recovery state from this endpoint.

### MA-C7 — MA1 has no exact wire or error mapping

“Session descriptor,” “post-command snapshot,” and “terminal receipt” leave required/optional keys,
snapshot representation, HTTP statuses, and tenant-rejection mapping undefined. The platform
snapshot is canonical raw JSON whose schema depends on `engine_ref`; API Foundation C18 forbids a
free-form JSON descriptor.

**Proposed contract:** enumerate exact create/current/command/terminal envelopes and a
discriminated `oneOf` snapshot arm per registered tenant/engine version, beginning with Pitch
`1.0.0`. Requests use the same generated tenant-command union. Map platform/store/tenant errors to
literal `{status,category,detail}` rows; transient SQL errors remain 5xx. Artifact/tenant growth
adds union arms through an additive API amendment, matching C18—never `map[string]any`.

### MA-C8 — The claimed per-session command budget does not exist in catalog data

The schema-v3 minigame definition has no command-rate or transition-budget field. The Pitch content
gate's 108-transition corpus budget is a CI expectation, not a player-session abuse policy. An
implementation would have to invent both the limit and refill semantics.

**Proposed contract:** v1 uses the API Foundation authenticated account limiter plus exactly one
in-flight claim per session; remove “catalog command budget” from MA1. If a measured abuse case later
needs a tenant budget, add a strict operational policy row (not simulation balance data) with
literal burst/refill/entry bounds and its own RFC.

### MA-C9 — MA3 is directional, not an executable component contract

Neither surface enumerates TypeScript props/events/states, the tenant component registry, or exact
error/reconnect rendering. `command vocabulary` cannot be inferred safely at runtime, and
`engine_ref -> child component` risks becoming a hardcoded second tenant registry. The recovery
toy seed also has no derivation. UI Foundation is accepted but not yet implemented, so its gates
cannot currently host these components.

**Proposed contract:** sequence implementation API/composition first, surface components after UI
Foundation lands. Define a generated tenant-surface registry keyed by the same
`(engine_ref,engine_version)` operation/schema arm, with exact props and callbacks; Pitch registers
one child. Define the recovery surface props from the generated MA2 DTOs plus a client-local,
non-authoritative toy seed generated on mount. Enumerate the closed UI states
`loading|active|paused_reconnect|required_terminal|error`, keyboard/cancel behavior, and copy keys.
AC5 runs through the UI Foundation browser/a11y fixture and remains blocked until that dependency is
implemented.

## Owner rulings on MA-C1–MA-C9 (2026-08-07)

- **MA-C1 — accepted as proposed.** MA2 is a REGISTRATION/GENERATION pass over the already-shipped,
  designated-reviewed recovery routes — never a second implementation. Exact schemas exported from
  the single operation registry; archived runtime behavior byte-for-byte; `session_expired` REMOVED
  from the public taxonomy (the Soul verdict's canonical reading: expiry is a watchdog terminal
  receipt, never an API error); the actual exact pairs (`unknown_id/recovery_session`,
  `not_eligible/exclusive_activity`, `conflict/recovery_session`,
  `idempotency_conflict/recovery_session`) are the documented surface. Conformance-discovered
  runtime changes route to the Soul Recovery owner.
- **MA-C2 — accepted as proposed.** This RFC owns the parked Platform AC1 composition slice:
  repository, closed tenant registry (Pitch `1.0.0`), pinned resolver, `minigame.Service`, one
  coordinator adapter on the authenticated API, platform exposed on `Composition`, bounded-drain
  participation proven either way; a real-socket composed-binary integration test drives the
  endpoints.
- **MA-C3 — RULED: the persisted sequence coordinate.** The seed contract is the recommended
  `minigame_session_seq` (persisted monotonic, Founder/run-scoped) with
  `Substream(runidentity.Seed(founder_id, run_seq) XOR uint64(seq), "minigame.session.v1").Next()`
  — the `fiscal_period_seq` precedent (F2), keeping seeds a pure function of immutable persisted
  inputs; the opaque-random alternative is REJECTED (it weakens the save-seeded law for one field's
  savings). The Summary's "no new persistence" claim is amended by this ruling: exactly this
  coordinate + MA-C5's receipt rows, nothing else. Placement and the (honest) save-schema bump
  follow the migration-chain law at implementation.
- **MA-C4 — accepted, recommended arm.** The command endpoint detects a terminal `PlayDecision`
  and composes the certified resolution IN THE SAME REQUEST, returning the stored terminal
  receipt; `/resolve` becomes retry/read of an already-terminal session (rejects an active one);
  the in-memory capability never crosses an HTTP boundary.
- **MA-C5 — accepted, persistence arm (the at-most-once amendment is REJECTED for the playability
  seam — a lost response stranding a claimed session is exactly what this RFC exists to prevent).**
  API command IDs with canonical-request SHA-256; unique `(session_id, command_id)` receipt rows
  claim-token-guarded in the same transition; same ID+hash → stored response, same ID+different
  hash → `idempotency_conflict`; create takes a client opaque idempotency key scoped to the
  authenticated Founder and returns the same descriptor on retry.
- **MA-C6 — accepted as proposed.** Founder-scoped current-session query (sole `active|claimed`
  under the exclusivity invariant); HTTP 200 closed union `{kind:"none"}` |
  `{kind:"active", session, snapshot}`; no claim tokens, no recovery state, resolved receipts only
  via the terminal/retry path.
- **MA-C7 — accepted as proposed.** Exact envelopes; discriminated `oneOf` snapshot arm per
  registered `(tenant, engine_version)` starting with Pitch `1.0.0` (the C18 pattern, never
  `map[string]any`); literal `{status, category, detail}` error rows; transient SQL stays 5xx;
  growth by additive amendment.
- **MA-C8 — accepted as proposed.** "Catalog command budget" is STRUCK from MA1; v1 = the API
  Foundation authenticated limiter + exactly one in-flight claim per session; any future tenant
  budget is a strict operational policy row with its own RFC, never simulation balance data.
- **MA-C9 — accepted as proposed.** Sequencing: API + composition first; surface components AFTER
  UI Foundation is implemented (AC5 explicitly blocked on it). Generated tenant-surface registry
  keyed by the same `(engine_ref, engine_version)` arm; Pitch registers one child; recovery
  surface props from generated MA2 DTOs + a client-local non-authoritative toy seed on mount;
  closed UI states `loading|active|paused_reconnect|required_terminal|error` with keyboard/cancel
  behavior and copy keys enumerated.

## Changelog

- 2026-08-07: created (draft) — the playability seam TP-C10/SR-C1/SR-C3 named.
- 2026-08-07: composite-reconnect open question ruled (no composite; per-family reads).
- 2026-08-07: Codex acceptance review filed MA-C1–MA-C9. Implementation remains blocked on honest
  recovery-route ownership, gameserver/platform composition, server-owned start inputs, terminal
  resolution/idempotency/current-session wire, exact schemas, and the executable UI boundary.
- 2026-08-07: MA-C1–C9 ALL RULED (accepted as proposed except where a decision was owed: C3 the
  persisted `minigame_session_seq` + substream contract; C5 the persistence arm). The Summary's
  "no new persistence" is amended to name exactly the C3 coordinate and C5 receipt rows.
  Implementation-ready.

## Codex implementation blockers (2026-08-07 — MA-C10–MA-C14)

The MA-C2 composition-only slice is implemented and tested. The public coordinator cannot be
implemented honestly from MA-C1–MA-C9 yet: the accepted proposals say that exact contracts will be
enumerated, but the normative body never enumerates them, and the chosen persisted seed coordinate
creates one replay/lifecycle decision that the acceptance round left to implementation.

### MA-C10 — The exact request/response envelopes promised by MA-C7 still do not exist

MA-C7 accepts a proposal to enumerate exact create/current/command/terminal envelopes and literal
error/status rows. No such table or schema appears in the RFC. In particular, it does not say
where create's idempotency key lives, whether command wraps the tenant command or flattens it, what
keys identify revision/engine/content, whether terminal command responses share the nonterminal
shape, or what body `/resolve` accepts. The API schema DSL cannot generate a closed `oneOf` from
directional nouns.

**Proposed contract:** append literal JSON shapes and status tables. Recommended requests are
create `{idempotency_key}`, command `{command_id,expected_revision,command}`, resolve `{}` and no
body for current. Responses are a closed union with a common exact descriptor
`{session_id,minigame_id,engine_ref,engine_version,constants_hash,mode,revision,status}` and a Pitch
`1.0.0` snapshot arm; terminal adds only the stored exact resolution receipt. Name the formats and
maximum bytes for both opaque IDs. Enumerate every deterministic mapping from repository,
platform, tenant, unlock/Soul, revision, and idempotency errors to `{status,category,detail}`;
unlisted/store errors are 500 `internal_invariant`, never guessed by handlers.

### MA-C11 — The sequence field's replay owner and atomic coordinator are unresolved

MA-C3 chooses a persisted `minigame_session_seq` and says placement/save bump follow the migration
law at implementation. Placement is semantic here. A side-table counter would be a second mutable
authority outside Founder replay. A Company-save field is run-local but cannot be mutated by the
existing session repository transaction without bypassing `ApplyLogged`. A Founder-save field can
use `ApplyFounderLogged`, but must reset at Exit and session creation must commit atomically with
the logged increment; the current `Repository.create` owns a separate transaction.

**Proposed contract:** Founder v21 adds exactly `minigame_session_seq`; Exit resets it to zero when
it advances the Company run. Add the server-only Founder command `start_minigame_session` and a
single coordinator transaction that locks Founder → Company → session, increments/logs the
sequence through the shared Founder transition boundary, derives the seed from the resulting
sequence and current Company `run_seq`, and inserts the genesis/session before commit. Retry by
create idempotency key returns the prior committed descriptor without incrementing. The command is
not client-parseable; Founder replay reproduces the sequence mutation from frozen resolved inputs.
If the owner instead selects Company state, specify the corresponding ApplyLogged coordinator and
lock order; a database-only counter is rejected by the no-second-authority law.

### MA-C12 — Exit with an active minigame strands a run-bound session

Sessions pin `(company_stream_id,run_seq,constants_hash)`. Nothing currently prevents `wind_down`
while a minigame is active. After Exit, commands can still advance the tenant snapshot, but terminal
resolution loads the Company's new run and fails the pinned-run invariant; the session cannot
resolve and remains the founder's sole active minigame forever. Resetting MA-C11's sequence at Exit
makes this lifecycle edge load-bearing.

**Proposed contract:** while one minigame session is `active|claimed`, Exit rejects
`not_eligible/minigame_session_active` before evaluation and mutates nothing. The production
service receives a read-only active-session resolver from composition; the same predicate runs in
live and replay from a frozen boolean resolved input. Alternative: Exit atomically cancels the
session and records a terminal receipt, which is a larger multi-stream contract. Do not let an
unlogged repository read decide replay behavior.

### MA-C13 — Command idempotency has no transaction boundary for terminal auto-resolution

MA-C5 requires a `(session_id,command_id)` request-hash/receipt row in the same transition. The
existing nonterminal `completePlay` and terminal `ResolveMinigameSession` use different
transactions; terminal resolution also writes two save streams, both logs, events, the faucet
window, the session, and its durable receipt. Writing an API receipt before or after those paths
would either replay an uncommitted result or permit duplicate tenant execution.

**Proposed contract:** extend the minigame repository with transaction-scoped command-receipt
helpers. Nonterminal completion appends the session command, updates its snapshot/revision, and
stores the canonical API response under the claim token in one transaction. Terminal completion
passes the command ID/hash/response into `ApplyMinigameResolutionTransaction`, which writes the
same receipt row in its existing all-or-nothing transaction. The command handler checks the row
before tenant execution; equal hash returns bytes, unequal hash returns
`idempotency_conflict/minigame_command`. Claim-lease expiry cannot let an old worker overwrite a
new receipt (`RowsAffected == 1` on the token guard).

### MA-C14 — API Foundation generation is still an upstream implementation dependency

The repository contains the schema DSL and operation registry foundation, but API Foundation's
OpenAPI/TypeScript generation, compatibility pins, middleware/router composition, and privacy gate
remain unchecked in `planning/api-foundation/plan.md`. MA-C1/MA-C7 require these routes to generate
from that single authority; manually mounting handlers now and promising later registration would
repeat the exact two-authority design the ruling rejected.

**Proposed contract:** implement MA runtime/coordinator code behind typed handlers after C10–C13,
but keep schema registration and endpoint mounting in the API Foundation continuation. The two
ranges then receive one combined real-socket/conformance review before this RFC can claim AC1–AC4.
Alternatively move the minimal generator/router slice into this RFC explicitly and consume it as
API Foundation work; never add a handwritten parallel router contract.

## Owner rulings on MA-C10–MA-C14 (2026-08-07)

- **MA-C10 — accepted, recommended shapes RULED as the normative wire.** Requests: create
  `{idempotency_key}` · command `{command_id, expected_revision, command}` (the tenant command as a
  nested object, never flattened) · resolve `{}` · current takes no body. Responses: a closed
  union over the common exact descriptor
  `{session_id, minigame_id, engine_ref, engine_version, constants_hash, mode, revision, status}`
  plus the discriminated Pitch `1.0.0` snapshot arm; terminal responses add ONLY the stored exact
  resolution receipt. Both opaque IDs (`idempotency_key`, `command_id`) use the C20 request-ID
  grammar for consistency: `^[A-Za-z0-9-]{1,64}$`, max 64 bytes, rejected-not-truncated. Every
  deterministic error mapping is enumerated as literal `{status, category, detail}` rows in the
  implementing schema table; anything unlisted (incl. store errors) is `500 internal_invariant` —
  handlers never guess.
- **MA-C11 — RULED: the Founder arm, exactly as proposed.** Founder v21 adds exactly
  `minigame_session_seq`; Exit resets it to zero when it advances the Company run. The server-only
  Founder command `start_minigame_session` runs in a single coordinator transaction locking
  Founder → Company → session, increments/logs the sequence through the shared Founder transition
  boundary, derives the seed from the resulting sequence + current `run_seq`, and inserts
  genesis/session before commit. Create-retry by idempotency key returns the prior committed
  descriptor WITHOUT incrementing. Not client-parseable; Founder replay reproduces the mutation
  from frozen resolved inputs. Company-state placement and the database-only counter are REJECTED
  (the no-second-authority law).
- **MA-C12 — RULED: reject-Exit, the primary arm.** While one minigame session is
  `active|claimed`, Exit rejects `not_eligible/minigame_session_active` before evaluation and
  mutates nothing; the production service receives a read-only active-session resolver from
  composition; the same predicate runs live and in replay from a frozen boolean resolved input.
  The atomic-cancel alternative is REJECTED for v1 (a larger multi-stream contract; a Pitch
  session always has a reachable terminal, so the player is never soft-locked — declared successor
  if a future tenant lacks that property). An unlogged repository read never decides replay
  behavior.
- **MA-C13 — accepted as proposed.** Transaction-scoped command-receipt helpers on the minigame
  repository: nonterminal completion appends command + snapshot/revision + canonical API response
  under the claim token in ONE transaction; terminal completion passes command ID/hash/response
  into `ApplyMinigameResolutionTransaction`, which writes the same receipt row inside its existing
  all-or-nothing transaction. Handler checks the row before tenant execution; equal hash → stored
  bytes; unequal → `idempotency_conflict/minigame_command`; the `RowsAffected == 1` token guard
  forbids an expired worker overwriting a newer receipt.
- **MA-C14 — accepted, primary arm.** MA runtime/coordinator code lands behind typed handlers
  after C10–C13; schema registration and endpoint mounting stay in the API Foundation continuation
  (C1–C20 are fully ruled — that work is unblocked); the two ranges receive ONE combined
  real-socket/conformance designated review before this RFC claims AC1–AC4. A handwritten parallel
  router is permanently rejected.

## Changelog (second rulings round)

- 2026-08-07: MA-C10–C14 ALL RULED — literal wire shapes (C20-grammar IDs); Founder v21
  `minigame_session_seq` with the atomic coordinator; Exit rejection over atomic cancel;
  transaction-scoped command receipts; combined-review sequencing with the API Foundation
  continuation.
