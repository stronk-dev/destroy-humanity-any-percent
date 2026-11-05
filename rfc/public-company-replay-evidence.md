# RFC: Public Company Replay Evidence

- **Status:** draft — original proposal independently technically approved 2026-11-05; input-v13 compatibility amendment awaits technical review; owner adoption/body authority pending; not implementation authority.
- **Author:** Codex
- **Created:** 2026-11-05
- **Design refs:** `design/06-tech.md` (hidden information, server authority, versioned replay); `design/05-mmo.md` (published community formulas).
- **Parent / amends:** API Foundation C2/C14/C19; Server Garden SG8/AC9/AC10/OD-12; Run Genesis & Replay.
- **Planning:** existing `planning/api-foundation/` and `planning/minigame-server-garden/` plans/logs.

## Problem and evidence

RP-606 now has an actual producer-to-public-download reproduction, independently
reviewed. A persistent Garden salt survives Exit and enters the original Company
archive served anonymously under C14. The adopted hidden-salt boundary and that
original-byte publication contract conflict. Hiding ordinary Garden responses
cannot repair the raw download. A new safe endpoint alone leaves the old leak open.

The independently approved `3a1b8737..ddda9f4c` feasibility study preserves complete
Company states, receipts, ordered events and next Company in both real replay
libraries when only the carried salt becomes a local null surrogate. Full Founder
output changes. Removing the field without an adapter fails the private codec.
See [evidence and limitations](../planning/platform-alignment/data-rights-current-content.md#2026-11-05--public-company-replay-projection-feasibility).
This supports the engineering direction below, not all-version acceptance or a
claim that the existing transport has been fixed.

The later independently approved Prestige correction `c195c0ae..a12218ab` adds
Company input13 and `run_ended`4 with frozen Founder age. The following explicit
draft amendment includes that defined representation; it does not infer coverage
from the prior input12 Garden feasibility study or expand its recorded verdict.
Input13 with an initialized Garden still needs the full acceptance population
below. Original technical approval and the pending owner question concerned the
proposal through input12; neither is silently recorded as approval of this delta.

## Specification

### P1 — Separate public evidence from private original history

Keep `run_log_archive`, run genesis, Founder history, verification inputs and saved
Founder state unchanged. Add a separately stored public Company representation.
No salt rotation, empty-Garden replacement, history rewrite, new simulation rule,
ranking suppression or weaker original verification is authorized.

Its promise is explicit: replay this Company's transitions from its published
Company genesis and server-supplied boundary facts, checking complete Company
outcomes/receipts/events. It does not verify private Founder growth or the truth
of every upstream boundary fact. That was not established by the existing Company
verifier either. An internal source hash binds a derived artifact to its original;
it is not a public cryptographic proof of omitted private facts.

This amendment excludes only the protected Garden salt from the current carried
Garden shape. It does not adopt the unresolved general export/deletion, retention,
ranking consent, player disclosure or nested-field publication policies. Those
remain independent publication gates; this RFC cannot certify full rights compliance.

### P2 — Explicit version and retirement of unsafe downloads

Mount only these new resources under `/api/public/v2/runs/{stream}/{seq}/`:
`verdict`, `genesis`, `replay-log`. Keep current run-ID/sequence admission, actual
`verified_runs` authorization and unknown/private/unranked404 semantics. Queue
status alone still cannot authorize a download. A known board run lacking a valid
public artifact returns the existing generic internal500 with no evidence bytes;
it must never fall back to the private archive. Operational errors are not404.

Use distinct operation IDs and a public-v2 stability surface in the same operation
registry. The closed C19 JSON/gzip raw-response alternatives remain sufficient.
The verdict is the exact existing nine-field DTO plus required
`evidence_profile:"company-public.v1"`; URLs for genesis/replay-log reference v2.
The catalog URL remains the existing pinned-hash v1 catalog resource. No new
catalog representation, partial union or handwritten generated SDK is authorized.

Genesis and replay-log are the exact separately stored public bytes, respectively
`application/json` and `application/gzip`. Manifest digests, strong ETags and
`X-Content-SHA256` identify those bytes, never the original private archive. Keep
C16's existing cache age, immutable flag, conditional handling, request-ID and
limiter rules for successful v2 resources. Private source hashes are not added to
public responses. Unknown profile/version must refuse, not be guessed from a path.

**Required owner choice:** at cutover, permanently retire the three old v1 run
resources for every run. They return the existing `unknown_id/run`404 body before
any cache/conditional response, with `Cache-Control: no-store`, no evidence body,
no ETag and no redirect. This uniform response discloses no run existence. No304,
query flag, alternate router mount or authenticated fallback may expose originals.
Other v1 operations and ranking stay available under their current contracts.

This is a deliberate security withdrawal of a published behavior, not an additive
v1 change. Owner adoption must explicitly authorize it and reconcile API C2/C14.
Keep the three old operation IDs, methods, paths, public-v1 surface, auth and
parameters registered, but replace each response set with only the existing exact
404 `APIError` response. Their handlers perform the uniform retirement response,
never read private evidence or invoke caching. Current generated OpenAPI and SDK
must therefore advertise only404 for those operations, with no raw/DTO200 success
arm. Preserve existing named schema definitions, even if now unreferenced; do not
hide an unrelated schema removal in this transition. New v2 operations have their
own descriptors and success types.

The generator's current unmodified compatibility check correctly rejects this
status removal. Establish a single explicit security baseline transition:

1. Preserve the exact pre-cutover `api-compat-v1.json` bytes as
   `docs/generated/history/api-compat-before-company-public-v1.json`; bind its
   SHA256 in the reviewed transition code/record. It is never regenerated.
2. A dedicated transition check requires all three exact original operation rows
   and changes ONLY their response lists to the already present404 alternative.
   All other operation/schema/generator bytes in that decoded baseline remain
   unchanged. Missing/extra target IDs, an altered original row/hash, a different
   replacement response or any other non-additive delta refuse.
3. Invoke the unchanged generic `CheckCompatibilityPin` on that derived baseline
   and the actual merged live registry. It continues to enforce every other old
   operation/schema and ordinary additive evolution. Separately require all three
   retired live rows to match their exact404-only definitions; the generic check
   must not permit accidentally reintroducing a success arm by response widening.
4. Only this checked transition may generate the successor active pin from the
   real merged registry. Afterward ordinary generation checks BOTH the frozen
   historical transition and the active pin. New v2 responses are pinned normally;
   future generation cannot regenerate the historical source or use an unverified
   `-update-pin` path to evade either check.

This narrow owner-authorized transition is not a generic exception tag, ignored
compatibility error or hand-edited successor pin. Keep the historical original
check's expected rejection as a test, while the validated transition passes.
Removal/status loss on any fourth operation must still fail. Do not describe
retired v1 as continuing to meet the old C14 promise.

Already downloaded/cached originals cannot be recalled by changing the origin.
Before publication/cutover, inventory actual deployments and caches, purge any
controlled unsafe cached responses, and record any irrecoverable exposure honestly.
Neither this RFC nor the fixture study asserts that deployed users were exposed.
If the inventory finds a real externally obtained salt that remains current,
withdrawal/purge/recording cannot restore its hidden status. Keep RP-606 and full
hidden-information acceptance OPEN for that population until a separately
owner-adopted remedy or explicit disposition and its required proof are complete.
This draft does not choose rotation, erase the exposure, or treat endpoint closure
as resolution of compromised persistent state. A clean future download is not
proof that an already downloaded secret became private again.

### P3 — Closed Company-only projection profile

The decompressed new archive has exactly
`{profile:"company-public.v1",archive:<Company archive object>}`. The inner archive
has the current private archive-v1 fields: `schema_version`, `company_stream_id`,
`run_seq`, `pin`, `genesis`, `entries`; their existing field/entry/event shapes and
ordering remain, except for the explicitly defined public replay-input difference.
`schema_version` stays1 inside this wrapper; it is not a claim of private-byte identity.
The profile discriminator is required independently of both archive and HTTP version.

Support only the currently defined Company replay-input versions2–13 and Founder
floors through25 with their existing matching pinned-artifact rules. First decode
and validate the original through the owning strict codecs. Unknown members,
versions, event kinds/schemas, unsupported pins or malformed inputs cannot be
copied through as opaque future fields. This is a closed profile, not recursive
"delete anything named salt" or an unconstrained JSON export.

For v12/v13 carry with the required `server_garden`, project that Garden's exact SG2
shape with the `salt_hex` member absent, preserving every other admitted member
and value. Omit it even when its private value is null. Earlier versions/floors
already lacking Garden remain absent. No other carried field, command, sequence,
receipt or event is removed or reordered. Unknown nested data must fail owning
schema admission rather than become a new public extension. Future input/Founder
versions need an explicit reviewed public profile revision before publication.
Input13's ending4 retains required `founder_age_ms` exactly as committed; the
profile neither removes that fact nor adds one to retained ending2/3. Existing
owning validation/source binding and versioned replay rules continue to apply.

Encode the newly derived JSON deterministically with the existing canonical replay
JSON rules, one final newline, then gzip best compression with zero modification
time, OS255 and no name/comment/extra headers. Persist the finished bytes once;
HTTP never regenerates them. Profile identity and stored byte hash distinguish
this representation from the original even when no protected field was present.
A future compressor/serialization change may not replace already stored bytes.

### P4 — Public replay adapter, never a private-state codec exception

Add dedicated Go and TypeScript public Company replay entry points. They exact-
decode P3's wrapper and public shapes, verify manifest/body digests and matching
run/pin/genesis identity, then reconstruct only a local replay input: insert
`salt_hex:null` into a public v12/v13 Garden carry before invoking the existing
Company transition library. A public carry containing a `salt_hex` member, even
null, refuses. The normal private codec keeps requiring and validating its field.

The adapter returns only the Company verdict/final state and Company transition
observations needed for verification. It never returns, persists, feeds back or
calls a private Founder-growth verifier with the surrogate Founder result.
Private replay and Founder history continue using the actual original salt.
Existing Company receipt/event/sequence/clock/terminal/constants checks all remain.
Public evidence must not be accepted as a private saved state, import or original
Founder log. Do not add a mode flag to ordinary persistence to make it fit.

For source-to-projection admission, compare complete Company states, receipts,
ordered events and terminal/next-Company outputs for every entry, not just final
verdict strings. Original and public-adapter replays must both verify. If a future
Company transition starts depending on the omitted salt, this comparison must
refuse publication; that needs a new contract, not a relaxed comparison.

### P5 — Immutable publication and source binding

Use one new append-only migration for `public_run_evidence`, keyed by
`(run_id,profile)` and linked to the original archive. Store its Company stream/run,
profile, original archive SHA256 and original genesis SHA256 (private provenance),
engine/constants identity, and finished public genesis/archive bytes plus each
public SHA256. Derive the original identities from the same source snapshot;
never accept caller-supplied provenance. Public genesis currently preserves the
admitted Company genesis bytes exactly, after profile admission; no Founder genesis
or private save is added.

An explicit idempotent publisher processes actually board-authorized runs from
stored originals. It validates original hashes/pins, performs P3/P4 admission and
full comparison, and atomically inserts the complete artifact. Source changes or
conflicting content for an existing key refuse without replacing anything.
Identical retries leave bytes/hashes unchanged. No HTTP request performs a write,
no migration fabricates artifacts or verified rows, and no read falls back to
current filesystem catalogs. Use existing context cancellation and projection
transaction conventions; a partial artifact must never be observable.

Backfill already compacted originals through the same publisher, without original
history mutation or a reset of board eligibility. A failure leaves ranking intact
and public evidence unavailable with the P2 invariant error; record the precise
unsupported/corrupt source internally without private data in public diagnostics.
Do not claim complete evidence availability until the whole actual eligible
population is successfully published or the owner explicitly dispositions a
separate historical incompatibility. A sampled two-entry run is not that census.

### P6 — Parent-body reconciliation and adoption

Acceptance must reconcile API A3/C4's privacy assertions, C14's body and accepted
ruling to distinguish private original archives from public Company projections;
C19 still requires exact stored bytes, now from the public artifact. C2 must name
only P2's three-resource security withdrawal, retaining ordinary compatibility for
all other operations and v2. Add no broad waiver or false claim of v1 continuity.

Garden SG8/AC9 must distinguish full private Founder replay from public Company
replay and retain both runtimes' exact outcome/hash obligations. AC10 must explicitly
include decompressed public archives, manifest/genesis, events/outbox/receipts and
alternate routes in hidden-information enumeration. OD-12's hidden persistent salt
and export exclusion stay intact. Original archive retention, backup protection and
rights-policy decisions remain governed by their existing owners.

These accepted-body/ruling edits require the existing author/owner reconciliation
authority. Technical review alone does not adopt this proposal or delegate those
edits. Implementation may begin only after explicit adoption and same-change body
reconciliation; owner copy and unrelated product choices are not delegated here.

## Acceptance criteria

1. Both public profile codecs accept all supported committed Company history
   versions/shapes and reject unknown/duplicate fields, wrong profile, illicit salt,
   unknown event/schema, wrong pin/version/floor and malformed gzip/JSON. Existing
   private salt/growth codecs and shared replay/vector populations remain green.
2. Original → stored public artifact → actual anonymous HTTP → generated consumer
   → Go and TypeScript Company verifiers passes with exact body/hash bindings.
   Exercise initialized Garden, ordinary commands, real mature harvest and Exit/
   next Company; use the real worker/compactor/board authorization on Postgres.
   Fixture-controlled growth is labeled; natural pacing is not inferred from it.
   Include actual input13 Garden-bearing ordinary and terminal histories with
   ending4 and frozen age, alongside retained input12 ending2/3. Both runtimes
   must match every unchanged Company outcome and the exact stored age; changing
   or dropping age and copying a salt must each be refused. Keep legacy age
   absent. Prior feasibility at12 or age vectors without Garden cannot substitute
   for this joined population.
   The complete pinned catalog download/loader loop must run, requiring its own
   adopted API prerequisites, not filesystem artifacts substituted for HTTP.
3. Compare every complete Company state/receipt/ordered-event/terminal/new-Company
   output. Four harvest credit/cap/forfeit/zero shapes retain both independent hash
   refusals. Corrupt receipt/events, sequence gap, constants/engine mismatch,
   mismatched source/genesis/run and nonterminal evidence must fail honestly.
4. Private originals, full Founder state/history and actual future Garden behavior
   remain unchanged and independently replayable before/after publication. A local
   surrogate must not escape into persistence/import/private replay. Repeated
   publication is byte-identical; conflict/cancellation/rollback leaves no partial
   artifact or original/history mutation. Test source corruption and concurrent retry.
5. Enumerate the generated salt, base and future draws across every public/raw/
   decompressed/conditional/alternate-route path. None may occur. Deliberately
   reconnecting the old raw source or copying salt into public projection must make
   this critical oracle fail. Retired v1 returns no evidence or304, including with
   the old ETag; board eligibility and other v1 resources continue normally.
6. Backfill supported real stored populations, including already compacted runs;
   cover mismatched/unsupported historical inputs explicitly. No private-source
   digest is mislabeled as the public digest and no server attestation is described
   as cryptographic proof of omitted Founder history. Record the deployment/cache
   exposure inventory and limits without deleting historical evidence.
7. Execute the actual registry→generator→OpenAPI/SDK/pins pipeline for the narrow
   retirement transition. The untouched old baseline must reject the removal;
   the checked historical transition and active pin must pass. Wrong source hash,
   changed target definition, missing/extra target, reintroduced200, altered schema,
   fourth-operation removal/status loss and bypassed transition must fail. Generated
   clients have no v1 evidence success arm; v2 retains its real success types.
8. Run cold affected Go/vet, both replay corpora, actual generated client checks,
   `make api-check`, migration checks and affected real-Postgres/public consumer
   populations. Run aggregate push verification before publication and reconcile
   hosted CI when available. Independent implementation review and existing full
   feature/release gates remain; a green profile unit test cannot close RP-606.

## Open decisions

Owner adoption must explicitly approve P2's v2 representation plus withdrawal of
all three old raw-evidence resources, and authorize the exact P6 body reconciliation.
The hidden-salt decision is already adopted and is not being reopened. Technical
review, implementation and acceptance proof remain distinct. No endpoint, schema,
migration, private history or CI behavior changes under this draft.

## Amendment record

- 2026-11-05: propose explicit input13/ending4 compatibility after the independently
  approved Prestige correction. Preserve the single Garden-salt projection,
  historical encodings, private originals, prior-compromise hold and all acceptance
  gates. The prior proposal's review/adoption request does not automatically cover
  this new delta; independent technical review and explicit adoption scope remain.
