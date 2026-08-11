# RFC: The First Content Epoch (epoch 6 — the owner-gated production mint)

- **Status:** implementing — the epoch-6 mint has landed and its designated cross-party review is
  **APPROVED**; the authorized dependent archival moves are in progress. This RFC exists so the
  completed owner sign-off approved a fully enumerated, precondition-checked change instead of a
  judgment call. Named successor of TP-C18 (The Pitch, ruled option a: fixture-first) and SR-C13
  (Soul Recovery, ruled: fixture-only now, minted together here).
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** none new — this RFC mints already-designed, already-ruled content. Balance law:
  `CLAUDE.md` (declarative data, hardcaps, formula shapes exact / constants are config).
- **Depends on:** every fixture-first content foundation — Meters + Achievements, Doctrine,
  Minigame Platform, Pet Care, Fiscal Quarters, Soul Foundation, Soul Recovery Activities, The
  Pitch, Minigame API & Surface — each at its own designated-review/archival gate (see FCE5).
- **Planning:** `planning/first-content-epoch/` (once implementing)

## Summary

Before this mint, the live epoch registry (`balance/epochs/phase0.json`, `current_epoch_id: 5`)
carried only the 7 base artifacts; every content system since — meters, achievements, doctrines, minigames, pets,
fiscal, soul (including recovery activities), pitch, minigame API — shipped **fixture-first**, with its reviewed
artifact bytes living in test fixtures and its production mint explicitly deferred to this RFC.
This RFC defines **epoch 6**: one dependency-complete mint that installs reviewed or owner-ratified
candidate bytes, registers the epoch, and activates the content under the already-shipped
activation laws. It authors **zero new mechanics**; first-production bytes use the explicit
SHA-pinned lane ruled in FCE-C2/C3, and every gate that made "reviewed" mean something applies.

## Motivation

Fixture-first was the honest scope for each individual system, but the game is not playable until
the content exists in a production epoch. The playability critical path (Minigame API & Surface →
Game-UI screens → first human-playable) terminates in content that must actually be minted. TP-C18
and SR-C13 both promised a dedicated, owner-gated RFC for that mint; this is it.

## Specification

### FCE1 — Scope: one dependency-complete mint, not a partial chain

The catalog loader (`server/replaycatalog/catalog.go`, `validArtifactNames`) enforces the artifact
dependency chain as a hard load failure:

```
achievements ⇔ meters · doctrines → meters · minigames → meters ·
pets → minigames · fiscal → pets · soul → fiscal · pitch → soul · minigame_api → pitch
```

An epoch carrying the public minigame API therefore **must** carry all 9 optional artifacts. Epoch
6 is exactly that: the 7 base artifact names + the 9 content artifacts, loaded as one bundle under
one new constants hash. Economy and Routes change under the Permits ruling; Categories changes
only to keep its exact `full_gate_set` synchronized with Routes (FCE-C10). TP-C18's "no partial
chain" ruling is thereby structural, not procedural.

### FCE2 — Byte provenance: reviewed promotion or owner-ratified first production

Each artifact's production bytes use one of two closed provenance lanes:

1. **byte-identical promotion** of an already-reviewed source artifact; or
2. **owner-ratified first-production bytes**, pinned by SHA-256, content-gated in both runtimes,
   and designated-reviewed before the mint consumes them.

The second lane owns the authored meters/achievements/pets documents and the literal composed
Categories/Economy/Fiscal/Soul candidates required by already-ruled cross-artifact contracts. The
ratified literal source → production manifest is
`planning/first-content-epoch/promotion-manifest.candidate.v1.json`. Its candidate review is filed
at `planning/first-content-epoch/log.md#2026-08-08-b0277a1`; no test helper may synthesize bytes
absent from that manifest.

**Any retune between review and mint is a new review.** If a provisional constant is changed from
the reviewed fixture bytes, the changed artifact re-runs its content gate and the change lands as
its own `BALANCE-CHANGE:` commit reviewed before the mint consumes it. The default is: mint the
provisional bytes as reviewed, retune later via epoch 7+ (retunes are cheap; the mint is what
unblocks playability).

### FCE3 — Registry mechanics

One mint commit (subject class `BALANCE-CHANGE:`) that:
1. adds the 9 optional production artifact files and installs the three ruled base-artifact
   replacements (FCE2);
2. appends the 9 optional rows to `artifacts` and the epoch-6 entry to `epochs` in
   `balance/epochs/phase0.json` — `{epoch_id: 6, name: "First Content", changelog_ref:
   "changelog/epoch-6.md", accepted_hashes: [<the new constants hash>]}` — and bumps
   `current_epoch_id` to 6;
3. writes `changelog/epoch-6.md` enumerating every artifact, its source fixture, its
   reviewed range, and the verdict entries consumed;
4. updates `deployment/content-manifest.v1.json` to the new `constants_hash` (and `copy_hash` if
   the copy pipeline output moves);
5. leaves every epoch-5-and-earlier accepted hash in place — historical replay must still resolve
   old bundles (AC3).

### FCE4 — Activation semantics: the shipped laws, no new ones

No save-schema change ships here — the Company v15–v18 and Founder v17–v20 migration chains landed
with their foundations. Activation follows each archived foundation's own law, restated here only
as the checklist the integration test covers:
- **Company-scoped** (meters, achievements, doctrines): new-mechanic activation is **NEW-RUN-BOUND**
  at the first epoch whose pinned catalog carries the artifact; in-flight runs stay frozen at
  genesis (`run_frozen_contributions` — no retroactive grants, no Company bump).
- **Founder-scoped** (minigames, pets, fiscal, soul, pitch, minigame API): activates on epoch
  adoption per each foundation's shipped rules. Pet v18 care state and pinned policies activate
  with empty maps; species/acquisition and first-record creation remain successor work. Fiscal
  accrual, soul meter, minigame/pitch start-eligibility, and Founder v21 session sequencing follow
  their archived activation clauses.
- **Leaderboards:** board binding across the epoch boundary follows the archived Leaderboards &
  Balance Epochs law unchanged; the mint itself makes no board decisions.

### FCE5 — The mint gates (ALL green before the mint commit exists)

1. **Review-complete (reformulated 2026-08-07 after the foundation audit — the original "archived
   before the mint" wording was CIRCULAR for mint-blocked foundations and is superseded):** for
   every ARTIFACT-CONTRIBUTING foundation (the 9 optional artifact owners), (a) its implementation range is
   designated-cross-party-review-covered with a complete range union, and (b) every acceptance
   criterion RELEVANT TO ITS ARTIFACT's behavior is green. Foundations whose only remaining open
   items are the mint itself (Meters, Achievements, Pet Care — their final AC IS this RFC) archive
   WITH or immediately after the mint, citing it; foundations whose open items are explicitly
   non-artifact successors (Minigame Platform AC6's combat-duel adapter, Pet Care AC3's combat
   cross-verify — both gated on the still-draft Combat Duel Engine) do not block the mint on
   those, and their archival waits for those successors without holding the epoch hostage.
   Non-artifact substrate foundations (Founder Attendance, API Foundation) are OUTSIDE this gate
   entirely. The changelog cites every consumed verdict.
2. **Content gates green** for every artifact at its production bytes (the Pitch corpus drift
   check, soul lifecycle corpus, and each loader's validation — TP-C18's "minted only after the
   Pitch content gate passes").
3. **Composed balance-harness pass:** the harness runs the FULL epoch-6 bundle (the first composed
   pacing run — every prior run exercised systems in isolation or partial fixtures) and its report
   is recorded in the planning log. Pacing regressions against the Phase-0 observation targets are
   findings, not vetoes — but they are findings the owner sees before signing.
4. **Copy coverage:** every copy key declared by the 9 optional artifacts ships through `verify-copy` green.
5. **Guard discipline:** kernel version policy applied (the mint commit touches watched
   balance paths → honest bump or guard-green no-op per policy); `make verify` + the Postgres suite
   green at the mint commit.
6. **Owner sign-off:** Marco approves the enumerated mint (this RFC's acceptance + the mint
   commit's review) — the gate that makes it a decision, not a default.

### FCE6 — Out of scope

- **THE PUSH / deployment** — Deployment Foundation, Marco-only, standing. Minting epoch 6 in the
  repo does not deploy anything.
- **Any new mechanic, meter, row, or constant** not already reviewed in a fixture.
- **Production DB work** — no migrations; the save chains already shipped.
- **Epoch 7+ retunes** — declared successor lane for balance changes after playtesting.

## Acceptance criteria

1. Epoch 6 registered per FCE3; `LoadBundle` resolves the new constants hash to a bundle carrying
   all 16 artifacts; the promotion table's source hash = production hash invariant is verified by
   test.
2. Activation integration test: a founder present across the epoch 5→6 boundary gets new-run-bound
   Company activation (no retroactive grants) and Founder-scoped activation per the shipped laws;
   a fresh founder at epoch 6 gets the full content set from genesis.
3. Replay continuity: pre-epoch-6 run logs still replay byte-identically through their pinned
   epoch-5 (and earlier) bundles.
4. The composed harness report exists in `planning/first-content-epoch/log.md` and the epoch-6
   changelog cites every consumed review verdict and reviewed range.
5. `make verify` + the Postgres integration suite green at the mint commit.

## Open questions — ALL RULED by Marco, 2026-08-07

1. **Epoch name — RULED: `"First Content"`** (mechanical, matches the registry pattern; flavor
   lives in data files, not registry names).
2. **Provisional bytes — RULED: mint as-is.** Byte-identical promotion of the reviewed fixture
   bytes; epoch 7 is the retune lane once the composed harness report and real play data exist.
   FCE2's default is now normative.
3. **Mint timing — RULED: ASAP once FCE5 is green**, independent of the Minigame API & Surface
   work; surface work then tests against production content. Note recorded with the ruling: FCE5.1
   requires the OLDER contributing foundations (meters, achievements, minigame platform, pet care,
   founder attendance) to reach their designated-review/archival gates too — that closure work is
   on the mint's critical path (audit in flight at ruling time).

## Blockers found during content-row drafting (2026-08-07)

### FCE-B1 — The reviewed doctrines fixture does not compose with unchanged routes bytes

Loader-verified during the content-row proposal work: the reviewed doctrines fixture references
`gate.t3_to_t4`, which is absent from the live `balance/routes/phase0.json` —
`DoctrineCatalog.ValidateRoutes` hard-fails at bundle load, so the dependency-complete epoch-6
bundle AS CURRENTLY CONSTITUTED cannot load. **RULED (Marco, 2026-08-07): option (a) — the routes
artifact is EXTENDED with the missing `gate.t3_to_t4` row** (a BASE-artifact byte change — the one
place FCE1's "base bytes unchanged" gives way; re-accept + content gate + review as a
`BALANCE-CHANGE:` before the mint consumes it). The doctrine row is unchanged.

### FCE-B2 — Scope clarification: no pet species roster in this mint

The ruled pet artifact grammar carries care/decay/trust/FSM rows only — it has NO
species/temperament slot. The launch species roster (design/04) therefore CANNOT be minted here
and belongs to the pet-acquisition successor RFC (which owns the grammar extension). Epoch 6's pet
artifact is the care-numbers artifact, nothing more.

The full row proposal (loader-validated drafts for meters/achievements/pets, 17 DESIGN-GAPs, copy-key
list) is at `planning/coverage-map/mint-content-rows-proposal.md` (internal) pending owner rulings.

## Codex acceptance review blockers (2026-08-07 — FCE-C1–FCE-C6)

The mint machinery is implementable, but the exact epoch is not yet acceptance-ready. These are
content/identity decisions; implementation must not resolve them by choosing convenient bytes.

### FCE-C1 — The Routes-extension direction is ruled, but the gate row still has no requirement bytes

Owner ruling `9334bee` settles FCE-B1's branch: epoch 6 includes Doctrine and extends Routes with
`gate.t3_to_t4`; retargeting/omission are no longer options. The ruling does not yet make the row
executable. The archived route example uses the nonexistent `company.permits`, while substituting
`company.cash` would choose a resource and amount that the temporal-validity RFC deliberately
refused to invent. The economy catalog's Tier-3 `1e12` progress coordinate is an observation target,
not automatically a gate price. A gate ID without a requirement fails the strict Routes grammar.

**Proposed contract:** add a narrow T3-to-T4 gate-content ruling (inside this RFC or a named
pre-mint RFC) containing the literal exact gate row: requirement resource(s), canonical amount(s),
and route list/order. If the resource is `company.permits`, the same contract owns its economy row,
source/faucet, and replay semantics; if it is `company.cash`, record the deliberate deviation from
the permits/community-gate design. The `BALANCE-CHANGE:` landing re-runs both route loaders,
economy gate-reference validation, doctrine `ValidateRoutes`, chronology/depletion gates, composed
Go/TypeScript bundle parity, and designated review before epoch 6 consumes the bytes. “Extend
routes” chooses the direction; these exact bytes close it.

### FCE-C2 — FCE2's byte-identical-promotion claim is false for the three proposed row artifacts

The proposed meters artifact changes grievance initialization/decay and drops an inert input; the
pet artifact replaces several fixture literals with design-grounded values and adds production FSM
rows; the achievements artifact assembles twelve production rows rather than promoting one reviewed
fixture wholesale. Those may be good first-production bytes, but they are authored balance content,
not byte-identical promotion. Leaving FCE2 unchanged would make the changelog materially false.

**Proposed contract:** classify the three drafts as owner-ratified first-production artifacts. Pin
their exact pre-mint SHA-256 values in FCE2's source table; land each changed family through a
`BALANCE-CHANGE:` content commit, its real Go/TypeScript content gate, and a designated review
before the epoch commit consumes it. Keep byte-identical promotion as the rule only for the other
artifact families whose production bytes actually equal their reviewed source fixture.

### FCE-C3 — Owner approval must bind complete artifact bytes, not only three highlighted literals

The proposal correctly flags many provisional numbers beyond grievance and achievement score:
meter band/decay literals; pet grid, floors, cooldowns, Soul gates, uniform Trust gain, and FSM
durations. Selecting only the three headline decisions would leave the implementer silently owning
the rest of the balance artifact.

**Proposed contract:** after deciding the doctrine scope, grievance direction, and achievement
grant, the owner approves the complete meters/achievements/pets JSON documents by exact SHA-256.
The current drafts are `meters=320deca9ccbe70c1822f0d2664ea75dfd7627d7f098dfd1243ef432bea7bb485`,
`achievements=1a11d6c5a0c044ff8077574bb71f1c893bde93a050e20a91e0d776c7e79f8903`, and
`pets=5c1f27006871ddbd688cdb36e673a64ef5080c92950d22df486576dfae4aa1c1`.
Any ruling edit that changes a document records the replacement hash. Codex recommends grievance
start/target `0/0` (derived quantities do not manufacture themselves) and flat `+4`
`achievement_score` (the literal design law with only the currency name changed), but neither is
implemented without owner ratification.

### FCE-C4 — The achievements artifact is deliberately unloadable against the shipped copy registry

All thirteen required keys are named, but their player-facing text is absent. The proposal's own
loader probe only passed after substituting an expanded in-memory registry; production
`copykeys.All()` correctly rejects the artifact today. FCE5.4 cannot be checked against key names
alone.

**Proposed contract:** before the achievements balance commit, land all thirteen exact copy rows
through the Copy Pipeline, including the possession-warning disclosure, provenance registry entries,
generated Go keys, and manifest hash. `make copy-check` and an achievement load using the real
`copykeys.All()` are discriminating gates. The balance artifact never lands in an unloadable
intermediate commit.

### FCE-C5 — FCE4 promises pet starter creation that FCE-B2 says cannot exist

The activation section says pet starter creation begins at the first pet-carrying epoch. The ruled
pet artifact has no species/acquisition authority, and FCE-B2 correctly defers that system. Epoch 6
can activate the Founder v18 care-state schema and its pinned policies, but it cannot create a pet
record or make care reachable for a founder without one.

**Proposed contract:** reconcile FCE4 to say exactly that: v18 care state/policy activates, while
pet acquisition/species and first-record creation remain successor work. The activation fixture
asserts empty initialized maps and pinned policy bytes, not a fabricated starter pet.

### FCE-C6 — The exact promotion manifest and cross-runtime proof are incomplete

The proposal supplies exact draft files for only three optional families. FCE2 still delegates the
remaining source-fixture selection and hashes to an implementing plan, and the validation record
only claims the Go loaders for the new JSON documents (with AJV for two schemas). A constants hash
and cross-runtime replay bundle cannot be reviewed from family names.

**Proposed contract:** before status moves to accepted, append one literal table covering every
epoch-6 artifact: artifact name, production path, source path or “owner-authored,” SHA-256, schema
version, content-gate command, and consumed designated verdict. Run the exact meters,
achievements, and pets bytes through both Go and TypeScript loaders in shared fixtures; pets needs
that explicit parity fixture even though it has no JSON Schema file. The table—not an
implementer-chosen path—is the owner-approved mint manifest.

## Owner rulings on FCE-C1–FCE-C6 (2026-08-07)

- **FCE-C1 — RULED (Marco): PERMITS NOW.** The gate is NOT cash-only interpolation:
  `company.permits` is introduced as the game's second economy resource, and `gate.t3_to_t4`
  requires `[{company.cash, 1e12}, {company.permits, 12}]`. The complete narrow contract —
  resource row, the `generator.legal_dept` faucet, the gate row, copy, and gates — is the
  commissioned pre-mint RFC **`rfc/permits-and-t3-gate.md`**. Its bytes land as `BALANCE-CHANGE:`
  with designated review BEFORE this epoch consumes them; FCE1's "base bytes unchanged" now gives
  way for BOTH the routes AND economy artifacts, exactly as far as that RFC specifies and no
  further.
- **FCE-C2 — accepted as proposed.** The meters/achievements/pets documents are OWNER-RATIFIED
  FIRST-PRODUCTION ARTIFACTS, not promotions; FCE2 is reclassified dual-lane: byte-identical
  promotion for every family whose production bytes equal the reviewed fixture; ratified-authored
  (SHA-pinned, content-gated, designated-reviewed `BALANCE-CHANGE:` commits) for these three.
- **FCE-C3 — RESOLVED: Marco RATIFIED all three complete documents by SHA-256 (2026-08-07):**
  `meters = 320deca9ccbe70c1822f0d2664ea75dfd7627d7f098dfd1243ef432bea7bb485` ·
  `achievements = 1a11d6c5a0c044ff8077574bb71f1c893bde93a050e20a91e0d776c7e79f8903` ·
  `pets = 5c1f27006871ddbd688cdb36e673a64ef5080c92950d22df486576dfae4aa1c1`
  (drafts at `planning/coverage-map/draft-artifacts/`, provenance per value in the proposal doc).
  Grievance 0/0 confirmed (aligned with Codex's recommendation). **Achievement scoring: TIERED
  2/4/8 STANDS** — Codex's flat-+4 counter-recommendation was put to the owner explicitly and
  overruled; design/02 §6 owes the tiered amendment note. Any later edit to a ratified document
  records its replacement hash here.
- **FCE-C4 — accepted as proposed.** The thirteen achievement copy rows land through the Copy
  Pipeline BEFORE the achievements balance commit; `make copy-check` + a load against the real
  `copykeys.All()` are the discriminating gates; no unloadable intermediate commit.
- **FCE-C5 — accepted as proposed.** FCE4's pet clause is reconciled by this ruling: epoch 6
  activates the Founder v18 care-state schema and pinned policy bytes ONLY; no starter creation,
  no acquisition; the activation fixture asserts empty initialized maps + pinned policies. (The
  FCE4 bullet's "pet starter creation begins at the first pet-carrying epoch" is superseded by
  this ruling and FCE-B2.)
- **FCE-C6 — accepted as proposed.** Before status moves to accepted, this RFC gains the one
  literal promotion-manifest table (artifact, production path, source or "owner-authored",
  SHA-256, schema version, content-gate command, consumed verdict) covering every epoch-6
  artifact; the three authored artifacts get shared Go/TS parity fixtures (pets explicitly,
  despite having no JSON Schema file). The table is the owner-approved mint manifest.

## Changelog

- 2026-08-07: created (draft) — the owner-gated dependency-complete mint promised by TP-C18 and
  SR-C13; scope pinned to the loader's enforced artifact chain; gates enumerated.
- 2026-08-07: FCE-B1 (doctrines/routes composition failure — loader-verified) and FCE-B2 (no
  species slot in the pet grammar; roster deferred to the acquisition successor) recorded from
  the content-row drafting pass.
- 2026-08-07: all three open questions ruled by Marco (name "First Content"; mint provisional
  bytes as-is; mint ASAP once gates green). No open questions remain — acceptance-ready.
- 2026-08-07: Codex acceptance review filed FCE-C1–FCE-C6, then reconciled C1 with owner ruling
  `9334bee` (Routes extension selected). The draft remains blocked on the gate's literal
  requirement bytes, exact first-production byte ratification, copy rows, pet activation wording,
  and the complete cross-runtime promotion manifest.
- 2026-08-07: FCE-C1–C6 ALL RULED (owner round): permits now (`rfc/permits-and-t3-gate.md`
  commissioned); dual-lane FCE2; the three artifact documents RATIFIED by SHA (tiered scoring
  stands); copy-first sequencing; pet activation reconciled; promotion-manifest table owed at
  acceptance.

## Codex implementation blockers — copy activation (2026-08-07)

### FCE-C7 — The thirteen achievement copy rows have identifiers but no owner-ratified text

The mint proposal enumerates the thirteen required keys and the possession-warning disclosure,
but deliberately does not draft the English copy. The Copy Pipeline makes those strings shipped
content: an implementation cannot choose their text, parameters, era binding, tone, or provenance
without authoring product copy that this RFC has not approved.

**Proposed contract:** supply or ratify one literal thirteen-row `copy.v1` document before
implementation. Unless the owner chooses otherwise, the structural defaults are `params: []`,
`era: null`, `provenance: []`, and `tone: "achievement"`; the possession-warning row uses the
already-ruled exact disclosure. Ratification names the complete source document and its SHA-256,
not only the key list.

### FCE-C8 — Copy references cannot point at the achievements artifact before that artifact exists

FCE-C4 says copy lands before the achievements balance commit. The current copy gate resolves
every active reference against the seed artifact set, so adding the two achievements reference
pointers in that earlier commit would make the intermediate history unloadable. Conversely,
waiting to add all copy until the mint contradicts the ruled copy-first gate.

**Proposed contract:** use a three-stage fail-closed landing:

1. land the thirteen copy rows and generated outputs as intentional orphans; `make copy-check`
   remains green;
2. stage and content-gate the owner-ratified achievements candidate against those generated keys,
   without changing the active seed;
3. in the epoch-6 mint commit, add the achievements artifact and its two reference-pointer rows
   atomically, regenerate outputs, and require `make copy-check` green again.

No tracked commit may contain a reference to an absent artifact or an active artifact whose copy
keys do not resolve.

## Owner rulings on FCE-C7–FCE-C8 (2026-08-07)

- **FCE-C7 — RULED: the thirteen copy texts are supplied here verbatim** (structural defaults as
  proposed: `params: []`, `era: null`, `provenance: []`, `tone: "achievement"`; each row's text
  below is `Title — body`; if the copy grammar carries separate title/body fields, split at the
  em-dash). Codex assembles the literal thirteen-row `copy.v1` document from these texts and files
  its SHA-256 for ratification (the FCE-C3 pattern; the text is the owner-authored part):
  1. `achievement.career_attended_hour` — "Billable Hour — One full attended hour. HR has logged
     your enthusiasm."
  2. `achievement.career_attended_day` — "Day One (Cumulative) — Twenty-four attended hours across
     your career. The deck calls this dedication."
  3. `achievement.generators_purchased_1` — "CAPEX — Bought your first generator. It began
     depreciating before the receipt printed."
  4. `achievement.first_gate` — "Out of the Garage — Crossed your first tier gate. The
     commemorative plaque is already being engraved."
  5. `achievement.generators_purchased_25` — "Procurement Pipeline — Twenty-five generators
     purchased. The vendor sent a gift basket."
  6. `achievement.generators_owned_100` — "Server Farm — One hundred beige towers. The hum is
     audible from the parking lot."
  7. `achievement.old_hand` — "Exit Interview — Completed your first exit. You kept the hoodie."
  8. `achievement.gate_burn_t3` — "Burn Rate — Spent 1e9 cash crossing a single gate. The board
     calls this investing in growth."
  9. `achievement.generators_owned_300` — "Beige at Scale — Three hundred beige towers. Nobody
     remembers what the first one does."
  10. `achievement.generators_purchased_25_tier_3` — "Enterprise Refresh — Twenty-five generators
      in one Tier-3 run. The old ones were fine."
  11. `achievement.tier_5` — "Hyperscale — Reached Tier 5. Your infrastructure has
      infrastructure."
  12. `achievement.exit_count_5` — "Serial Founder — Five exits. At this point the exits are the
      product."
  13. `achievement.possession_warning` — the ALREADY-RULED exact possession disclosure, verbatim,
      unchanged (this ruling adds no text to it; it is listed for completeness of the
      thirteen-row document).
- **FCE-C8 — accepted as proposed.** The three-stage fail-closed landing is normative: (1) the
  thirteen copy rows + generated outputs land as INTENTIONAL ORPHANS with `make copy-check` green;
  (2) the ratified achievements candidate is staged and content-gated against the generated keys
  WITHOUT touching the active seed; (3) the epoch-6 mint commit adds the artifact + its two
  reference-pointer rows atomically and regenerates, `make copy-check` green again. No tracked
  commit may reference an absent artifact or carry unresolvable active copy keys.

## Changelog (C7–C8 round)

- 2026-08-07: FCE-C7 ruled (all thirteen copy texts authored in the ruling; document assembly +
  SHA ratification owed); FCE-C8 accepted (three-stage fail-closed landing normative).

## Codex implementation blocker — missing thirteenth copy byte (2026-08-07)

### FCE-C9 — `achievement.possession_warning` still has no literal text

FCE-C7 says all thirteen texts are supplied verbatim, but row 13 refers to an “ALREADY-RULED exact
possession disclosure.” No such literal exists in this RFC, the Achievements RFC/log, the design
documents, the mint proposal, the copy catalogs, or tracked history. The first twelve rows are
byte-determined; inventing the thirteenth would make the assembled SHA partly Codex-authored and
would contradict the requested owner ratification.

**Proposed contract:** supply the exact single-line text for
`achievement.possession_warning`. Its already-ruled structural fields remain `params: []`,
`era_variants: null`, `provenance: []`, and `tone: "achievement"`. Once supplied, Codex assembles
all thirteen byte-sorted rows, runs the intentional-orphan stage, and files the complete source
and generated-copy SHA-256 values together.

## Owner ruling on FCE-C9 (2026-08-07)

- **FCE-C9 — RULED, with the error owned:** the FCE-C7 row-13 reference to an "already-ruled exact
  possession disclosure" was WRONG — no such literal ever existed (a phantom citation in the
  Claude-side ruling; Codex was right to refuse to author it). The literal text is supplied here
  verbatim:
  **`achievement.possession_warning` = "Verified by possession: the game checked what you own
  right now. It did not ask how, and it will not ask again."**
  Structural fields as already ruled (`params: []`, `era_variants: null`, `provenance: []`,
  `tone: "achievement"`). All thirteen rows are now byte-determined; Codex assembles the
  byte-sorted document, runs the intentional-orphan stage, and files the source + generated-copy
  SHA-256 values together for ratification.

## Codex implementation blockers — literal bundle composition (2026-08-08)

### FCE-C10 — The Permits gate also changes the Categories artifact

`leaderboard.LoadCategoryCatalog` requires `full_gate_set` to exactly equal the sorted gate IDs
from Routes. Adding ruled `gate.t3_to_t4` while retaining `balance/categories/phase0.json` makes the
full epoch-6 bundle fail closed. Existing composition tests hid this by inserting the gate into a
decoded Categories object in memory; no candidate bytes or manifest row owned that change.

**Proposed contract:** ratify `balance/testdata/first-content/categories-v1.json` as the Categories
candidate, SHA-256
`8232b8932649aafdfef6a4502ee4d6003ab6665c37042926fa8eace2b619f8ef`. It differs from the
epoch-5 artifact only by the sorted insertion of `gate.t3_to_t4` in `full_gate_set`; category rows,
fact sets, predicates, and timers are byte-unchanged. Categories becomes the third ruled base-byte
replacement alongside Economy and Routes, and its designated candidate review must include the
pre-Permits-set rejection probe now in both runtimes.

### FCE-C11 — Three reviewed helpers synthesize production bytes that no artifact owns

The composed server tests currently append Fiscal multiplier declarations to Economy, append the
`minigame.pitch` unlock to Fiscal, and replace Soul's fixture-only debit source in memory. The
individual source fixtures therefore are not literal production artifacts. Directly promoting
them either fails load (Fiscal and epoch-seeded Soul) or omits a ruled seam (Pitch unlock).

**Proposed contract:** ratify the following literal candidates, each containing only already-ruled
rows and no new mechanic or number:

- Economy v3 (Permits rows + Fiscal multiplier declarations):
  `2d4807b7628e3e258536802625ba35806c32d0429c3e819a19a7a287e3c552a1`;
- Fiscal v1 (reviewed baseline + `minigame.pitch`, cost 3):
  `3847236f8001ed7e29ab41054fbeef38c5e5ea8b838e478d2c4057fdc417f2a9`;
- Soul v1 (reviewed recovery rows + empty production `debit_sources`, as SR-C13 ruled):
  `a57798f94892a86fd6ea727b76d5bfa663db27c4abd10180204c26ea83587de4`.

The candidate designated review consumes the prior component verdicts and re-proves the composed
bytes. Test helpers cease being artifact authors.

### FCE-C12 — Candidate manifest and activation SHA ratifications are now concrete

The machine-readable sixteen-artifact table is
`planning/first-content-epoch/promotion-manifest.candidate.v1.json`; its composed constants hash is
`sha256:1a4463bcf67440ce1ba01e6c6eb850c0614329cac63064ef07725d042c7cf21a` in both Go and
TypeScript. The two explicit activation inputs awaiting the requested owner ratification are:

- achievement copy source `0dd211486b3e988c0fffa5311ed95c216f5bc08b4f9fd6ef7068409c3a091cf3`
  (generated catalog point-in-time attestation
  `2f299e2a3babb8f04e02c8406a127f8d3fbab321689038fbb22614cc6dba166b`);
- `minigame_api` v1 `b16b5e0eb6f9426c8b1b94255e2d8e04f53f78b391fdbbb348ad7438d7bab31c`.

**Proposed contract:** ratify the C10/C11 candidates, the two activation inputs above, the literal
production paths below, and the composed candidate hash as the bundle to receive the designated
candidate review and composed harness run. Generated copy remains a point-in-time attestation and
is refreshed at the mint, matching the Permits-copy precedent. Ratification does not authorize the
mint; FCE5.3, candidate review, and the separate owner sign-off remain mandatory.

### FCE-C6 literal promotion table — ratified after C10–C12 candidate review

The JSON manifest is normative for full command strings and verdict references. Its seven
first-production rows consume the designated candidate-review entry `b0277a1`; the other rows
retain their named implementation verdicts.

| artifact | production path | source path | SHA-256 | schema | content gate | consumed verdict |
|---|---|---|---|---:|---|---|
| achievements | `balance/achievements/first-content.json` | `balance/testdata/first-content/achievements-v1.json` | `1a11d6c5a0c044ff8077574bb71f1c893bde93a050e20a91e0d776c7e79f8903` | 1 | achievements + replaycatalog + client | Candidate `b0277a1` |
| categories | `balance/categories/phase0.json` | `balance/testdata/first-content/categories-v1.json` | `8232b8932649aafdfef6a4502ee4d6003ab6665c37042926fa8eace2b619f8ef` | 1 | leaderboard + replaycatalog | FCE-C10 + Candidate `b0277a1` |
| commons | `balance/commons/phase0.json` | same | `33d4e73a32e12c973acf9633a1e829fd4da2de0753c6004821fb93ff14208c93` | 1 | schema | Commons Compact verdict |
| doctrines | `balance/doctrines/first-content.json` | `balance/testdata/first-content/doctrines-v1.json` | `a3bca5f7eb07fb3b5bf185ce6191771c044a033b47c6bba390582dd7e1745672` | 1 | doctrine + replaycatalog + client | Doctrine verdict + candidate review |
| economy | `balance/catalogs/phase0.json` | `balance/testdata/first-content/economy-v3.json` | `2d4807b7628e3e258536802625ba35806c32d0429c3e819a19a7a287e3c552a1` | 3 | schema + economy + fiscal + replaycatalog | Candidate `b0277a1` |
| factions | `balance/factions/phase0.json` | same | `e44f461eca6cc6c048edebc42356915e6d4be16f480b4795a1fcc458855005fe` | 1 | schema | Faction verdict |
| fiscal | `balance/fiscal/first-content.json` | `balance/testdata/first-content/fiscal-v1.json` | `3847236f8001ed7e29ab41054fbeef38c5e5ea8b838e478d2c4057fdc417f2a9` | 1 | fiscal + replaycatalog + client | Candidate `b0277a1` |
| guilds | `balance/guilds/phase0.json` | same | `e70e644fd62be3c37e0ae465ea55eb104dfc83f810f2d66f11806328d18366fa` | 1 | schema | Guild verdict |
| meters | `balance/meters/first-content.json` | `balance/testdata/first-content/meters-v1.json` | `320deca9ccbe70c1822f0d2664ea75dfd7627d7f098dfd1243ef432bea7bb485` | 1 | meters + replaycatalog + client | Candidate `b0277a1` |
| minigame_api | `balance/minigame-api/first-content.json` | `balance/testdata/minigame-api-candidate-v1.json` | `b16b5e0eb6f9426c8b1b94255e2d8e04f53f78b391fdbbb348ad7438d7bab31c` | 1 | minigameapi + replaycatalog + client | MA `ce69a4b` |
| minigames | `balance/minigames/first-content.json` | `testdata/minigame/pitch-v3.json` | `f08fd3ab1959da66f389ef918b936f81d8a2562762055e7b27f4f9e771ff0862` | 3 | minigame + replaycatalog + client | Minigame Platform verdict |
| pets | `balance/pets/first-content.json` | `balance/testdata/first-content/pets-v2.json` | `5c1f27006871ddbd688cdb36e673a64ef5080c92950d22df486576dfae4aa1c1` | 2 | pet + replaycatalog + client | Candidate `b0277a1` |
| pitch | `balance/pitch.json` | `balance/testdata/pitch-v1.json` | `bd4218199c5ef00eaa2851020f6d77fcf826a30eee1d399a371a711b9b0ee10f` | 1 | pitch + replaycatalog + client | Pitch `c76101a` |
| prestige | `balance/prestige/phase0.json` | same | `1873090781bed666c8f989169a9e59990547b1f713ac2f9a8215f51d3f0ea7ec` | 1 | schema | Prestige verdict |
| routes | `balance/routes/phase0.json` | `balance/testdata/permits-t3-gate-candidate-v1.json` | `6c7c4350bcd43840a141fb5c0525d9779f11ed0ed836a8783f21f22f6c880df2` | 1 | schema + routes + replaycatalog | Permits `88e2054` |
| soul | `balance/soul/first-content.json` | `balance/testdata/first-content/soul-v1.json` | `a57798f94892a86fd6ea727b76d5bfa663db27c4abd10180204c26ea83587de4` | 1 | soul + replaycatalog + client | Candidate `b0277a1` |

## Owner rulings on FCE-C10–C12 — RATIFIED (Marco, 2026-08-08)

Ruled after the candidate designated review (planning/first-content-epoch/log.md, 2026-08-08
APPROVED entry — all 16 hashes independently recomputed, every byte traced to a ruled or reviewed
source):

- **FCE-C10 — RATIFIED.** `balance/testdata/first-content/categories-v1.json`
  (`8232b8932649aafdfef6a4502ee4d6003ab6665c37042926fa8eace2b619f8ef`) is the Categories
  candidate — the THIRD ruled base-byte replacement alongside Economy and Routes, differing from
  epoch-5 by exactly the sorted insertion of `gate.t3_to_t4` in `full_gate_set`, with the
  both-runtimes rejection probe as its standing gate. FCE1's body statement of this change is
  hereby CONFIRMED (closing the review's OBS-1).
- **FCE-C11 — RATIFIED, all three composed candidates:** Economy v3
  `2d4807b7628e3e258536802625ba35806c32d0429c3e819a19a7a287e3c552a1` · Fiscal v1
  `3847236f8001ed7e29ab41054fbeef38c5e5ea8b838e478d2c4057fdc417f2a9` · Soul v1
  `a57798f94892a86fd6ea727b76d5bfa663db27c4abd10180204c26ea83587de4`. Test helpers cease being
  artifact authors; every row is already-ruled (Permits pins + the two reviewed fiscal multiplier
  declarations; the TP-C15 pitch unlock at cost 3; SR-C13's recovery rows with empty production
  debit_sources).
- **FCE-C12 — RATIFIED:** the achievement copy source
  `0dd211486b3e988c0fffa5311ed95c216f5bc08b4f9fd6ef7068409c3a091cf3` (generated catalog
  `2f299e2a…166b` as point-in-time attestation, refreshed at mint per the Permits precedent) ·
  `minigame_api` v1 `b16b5e0eb6f9426c8b1b94255e2d8e04f53f78b391fdbbb348ad7438d7bab31c` · the
  literal production paths in the FCE-C6 table · the composed candidate constants hash
  `sha256:1a4463bcf67440ce1ba01e6c6eb850c0614329cac63064ef07725d042c7cf21a` as THE bundle for the
  composed harness run. Per OBS-2, the manifest MAY gain a `provenance` field
  (source-fixture | owner-authored | base-unchanged) at the implementer's discretion — no re-hash
  of candidate documents is implied (the manifest is planning-tier).
  **Ratification does NOT authorize the mint: FCE5.3 (composed harness report, owner-read) and
  FCE5.6 (the mint sign-off) remain, in that order.**

## Changelog (C10–C12 round)

- 2026-08-08: FCE-C10–C12 RATIFIED in full after the APPROVED candidate designated review; the
  manifest's pending-verdict fields may now cite that review's entry. Remaining before the mint:
  the FCE5.3 composed harness report, then the FCE5.6 owner sign-off.

## FCE5.6 — OWNER SIGN-OFF: THE MINT IS AUTHORIZED (Marco, 2026-08-09)

Given with everything on the table: all 16 candidate documents owner-ratified and byte-pinned
(FCE-C3/C10/C11/C12); every implementation commit designated-review-covered; the FCE5.3 harness
report reproduced byte-identically by its designated review, the 7 Chaos findings proven a
sampler artifact, Casual zero-delta on raw data; and the review's Finding 1 (new-content dynamics
load-validated but not simulated) weighed against the provisional-bytes law, visible hardcaps,
and the epoch-7 retune lane. **Codex is authorized to execute the FCE3 mint commit** — the
`BALANCE-CHANGE:` copying every ratified document byte-identically, appending epoch 6 "First
Content", bumping `current_epoch_id`, and writing `changelog/epoch-6.md` citing every
consumed verdict — followed by the mint's own designated review, then the Meters/Achievements/
Pet Care archival moves citing it. THE PUSH remains separate and Marco-only. The new-content
harness scenarios (Finding 1's evidence gap) are registered epoch-7-lane work.

## Changelog (sign-off)

- 2026-08-09: FCE5.3 satisfied (verdict 270e97b); FCE5.6 GIVEN — mint authorized.

## Post-mint reconciliation (2026-08-10, Claude-side — closes the mint review's F2)

AC5's "green at the mint commit" is reconciled to **"green at the mint RANGE head"**: the
baseline-change guard's own two-commit protocol makes literal per-commit greenness impossible
(the Postgres suite requires the remediation commit; harness-check requires the separate baseline
commit). The mint verdict (2026-08-10) verified green at 08c995e. This is the acceptance text
matching the shipped guard protocol, not a weakening.

## Epoch-7 content-dynamics harness blockers (Codex, 2026-08-10 — EH-C1–EH-C7)

The sign-off correctly registered pacing coverage for Active-Play buff windows, Fiscal harvests,
Pitch payouts, and Permit accrual. That obligation cannot be discharged by adding rows to the
current scenario file: the shipped harness owns only Company `Transition` over the
economy/Routes/Commons slice. Fiscal harvest is a Founder transition, Active Play requires the
full pinned policy bundle and schedule resolution, and Pitch uses the tenant engine plus platform
payout kernel. Copying any of those mechanics into `server/harness` would create a second balance
authority. Implementation pauses for these executable contracts.

### EH-C1 — One production-owned simulation boundary is missing

**Proposed contract:** add a second closed harness lane, `content_dynamics.v1`. It loads the exact
active epoch through `epochseed` + `replaycatalog` and invokes production-owned pure boundaries:
Company evaluation/Active Play, Founder Fiscal commands, Pitch `Tenant.Create/Apply`, and the
platform payout selection/conversion kernel. Where the live resolver is currently private, expose
a simulation-only wrapper analogous to `SimulateTransition`; the source guard permits calls only
from `server/harness` and tests. The harness may assemble inputs, never reimplement arithmetic or
author replay-resolved records.

### EH-C2 — The scenario and report grammars are unspecified

**Proposed contract:** one strict registry binds the active epoch seed, a scenario document, and a
golden report. The scenario has `{schema_version:1,id,version,runs,transition_budget}`. `runs` is a
closed union of `active_play_window | fiscal_harvest | pitch_payout | permit_accrual`; each arm has
an ID, string seed range, horizon, and only the mechanical coordinates named below. The report
records the full constants hash, scenario hash, exact run count, ordered integer/Decimal
observations, and named invariant failures. Unknown keys, missing artifact owners, unsafe seeds,
duplicate IDs, or work above the declared budget fail before execution.

### EH-C3 — The four policies need exact, non-self-serving definitions

**Proposed contract:**

- `active_play_window`: initialize a new Company run through the production initializer, advance
  to the first naturally spawned opportunity, claim it, and compare its declared target's output
  during the complete buff window with an identical unclaimed control. Never synthesize a pending
  opportunity or select an effect in the report.
- `fiscal_harvest`: initialize Founder v21 at a period boundary and observe production-owned lazy
  harvests after exactly 1 and 4 complete Fiscal periods, including sequence, credited amount, and
  hardcap status.
- `pitch_payout`: seeds 0–63 under a versioned deterministic policy: play the first four
  byte-sorted hand IDs; in shops buy the cheapest affordable offer (price, then offer ID), otherwise
  end shop. Observe terminal round, certified payout score, and converted Company Cash.
- `permit_accrual`: one purchased `generator.legal_dept`, no other producers; compare frozen
  Fiscal credit 0 versus the hoard cap (100). Observe time to 12 Permits, time to the visible cap
  24, and the cap reason. This pins the intended `fiscal.hoard` ×2 all-target interaction.

### EH-C4 — Pacing observations versus blocking invariants are not separated

**Proposed contract:** exact state/receipt reconciliation, partition invariance, declared cap,
artifact identity, and transition-budget failures block immediately. First-lane pacing values are
owner-facing observations with no invented pass envelope: Active-Play bonus delta, Fiscal credits,
Pitch p50/p95 terminal round and payout, and Permit p50/p95 times. After owner review accepts the
initial golden, later drift uses the existing 10% warning / 25% failure rule.

### EH-C5 — Baseline governance must remain fail-closed

**Proposed contract:** the scenario/runner lands first. Its generated golden lands in a separate
`BALANCE-CHANGE:` commit under the existing full-history guard; extend the governed path set and
adversarial fixtures rather than creating a second exception. `make harness-check` runs both the
legacy Phase-0 lane and every registered content-dynamics golden.

### EH-C6 — Full-bundle identity and historical behavior need one rule

**Proposed contract:** registry entries name the epoch-seed path, not a hand-authored subset or
hash. The runner recomputes the complete artifact bundle and records its accepted constants hash.
An entry follows its pinned epoch bytes forever; advancing the active epoch requires a new entry,
not reinterpretation under deploy-current content.

### EH-C7 — Execution cardinality must be literal

**Proposed contract:** 1 Active-Play control pair, 2 Fiscal cases, 64 Pitch seeds, and 2 Permit
cases; the scenario declares the exact total transition budget derived from those arms. No
Postgres, HTTP, sleeps, background workers, or unbounded per-millisecond loops are permitted.
The command exposes `make content-harness` and the read-only gate remains `make harness-check`.

**Status:** owner rulings required before implementation. The proposals deliberately reuse the
shipped math and governance; Codex will not invent a parallel content simulator.

## Owner rulings on EH-C1–EH-C7 (2026-08-10)

All seven accepted AS PROPOSED — the contracts are exactly the right shape (production-owned
boundaries, no second balance authority, strict grammars, fail-closed governance, literal
cardinality). Binding highlights:

- **EH-C1:** the `content_dynamics.v1` lane loads the exact active epoch through
  `epochseed` + `replaycatalog` and invokes production-owned pure boundaries only; private
  resolvers get simulation-only wrappers with the source guard (`server/harness` + tests). The
  harness assembles inputs, NEVER reimplements arithmetic or authors replay-resolved records.
- **EH-C2:** the strict scenario/report registry as specified; unknown keys / unsafe seeds /
  budget overruns fail before execution.
- **EH-C3:** the four policy definitions as written — including the control-pair discipline
  (never synthesize an opportunity; never select an effect in the report), the 0-vs-hoard-cap
  Permit comparison pinning the intended `fiscal.hoard` ×2 interaction, and the versioned
  deterministic Pitch policy over seeds 0–63.
- **EH-C4:** blocking invariants vs owner-facing observations partitioned as specified; the
  initial golden has NO invented pass envelope — owner review accepts it, then the existing
  10%/25% drift rule governs.
- **EH-C5:** runner first, golden in a separate `BALANCE-CHANGE:` commit under the EXISTING
  full-history guard (extend the governed path set + adversarial fixtures; no second exception).
- **EH-C6:** registry entries name the epoch-seed path; entries follow their pinned epoch bytes
  forever; a new active epoch = a new entry.
- **EH-C7:** literal cardinality (1+2+64+2), declared total transition budget, no
  Postgres/HTTP/sleeps/workers; `make content-harness` + the read-only `make harness-check`.

**B6 is UNBLOCKED.** The initial golden's owner acceptance rides the normal report-read flow
(the FCE5.3 pattern: report + designated review, then Marco reads).

## Changelog (EH round)

- 2026-08-10: EH-C1–C7 all accepted as proposed; the content-dynamics lane is implementable.

## Epoch-7 content-dynamics implementation blockers (Codex, 2026-08-10 — EH-C8–EH-C9)

Implementation against the minted repository exposed two contradictions that the acceptance
review's abstract bundle model did not catch. Neither can be resolved by choosing convenient
fixture bytes inside the runner.

### EH-C8 — Epoch 6 cannot execute the ruled Active-Play arm

Epoch 6 pins sixteen artifacts but does not pin `opportunities`, and its economy artifact contains
no `active_play` multiplier-source declarations. Loading the exact active epoch through
`epochseed` + `replaycatalog` therefore correctly yields `CatalogBundle.Opportunities == nil`.
The only opportunity policy is `balance/testdata/active-play-foundation-v1.json`; injecting it
would violate EH-C1/EH-C6 and make a fixture look like active content. `docs/active-play.md`
explicitly confirms that no production epoch currently pins the artifact.

**Proposed contract:** land the strict runner and its production-boundary tests against a
fixture-only complete bundle, but keep the production content-dynamics registry empty and do not
generate the initial golden yet. The next owner-gated content epoch (the T0–T1 candidate unless
superseded) must include an exact owner-ratified `opportunities` artifact plus its required
`active_play` economy multiplier declarations. Only after that mint does the first registered
scenario run all four EH-C3 arms and establish the initial golden. A missing artifact owner is a
hard pre-execution error for a registered scenario; no skip or synthetic policy is allowed.

### EH-C9 — An epoch-seed path does not preserve historical artifact bytes

EH-C6 says each registry entry follows its pinned epoch bytes forever, but the seed has one global
artifact-path list and those production paths move forward at later mints. An entry containing only
`balance/epochs/phase0.json` will resolve deploy-current files after epoch 7 and cannot reconstruct
epoch 6. Accepted hashes prove identity but do not recover the bytes. Running only the newest entry
would contradict EH-C5's requirement to check every registered golden.

**Proposed contract:** every registry entry names `{epoch_seed_path, epoch_id,
bundle_snapshot_manifest}`. `make content-harness` generates the immutable snapshot from the exact
active `epochseed.Bundle`; it never accepts a hand-authored subset. Snapshot schema v1 records the
complete sorted artifact set as `{name, production_path, snapshot_path, sha256}`, the accepted
constants hash, and the source epoch coordinate. Artifact bytes live under
`testdata/harness/content-dynamics/bundles/<full-hash>/`. The read-only loader verifies set equality,
every per-file hash, the recomputed bundle hash, and acceptance by the named epoch before execution.
Later mints add new snapshots/entries and never reinterpret old ones. The baseline guard governs
the registry, scenario, snapshot manifest, snapshot bytes, and golden with adversarial
missing/extra/tampered-artifact fixtures.

## Owner rulings on EH-C8–EH-C9 (2026-08-10)

- **EH-C8 — accepted as proposed, with the consequence ROUTED:** the runner + production-boundary
  tests land against a fixture-only complete bundle; the production content-dynamics registry
  stays EMPTY and NO initial golden is generated until an epoch pins `opportunities` + its
  `active_play` economy declarations. **Consequence routed to T0–T1 (binding): the T0–T1 epoch
  candidate set GROWS by an owner-ratified `opportunities` artifact and the required economy
  multiplier declarations** — the current six-document proposal is incomplete by exactly that;
  Codex extends it in the same draft-ratify lane. A missing artifact owner is a hard
  pre-execution error; no skip, no synthetic policy.
- **EH-C9 — accepted as proposed.** Registry entries are `{epoch_seed_path, epoch_id,
  bundle_snapshot_manifest}`; `make content-harness` GENERATES the immutable snapshot from the
  exact active `epochseed.Bundle` (never a hand-authored subset); snapshot schema v1 with the
  complete sorted `{name, production_path, snapshot_path, sha256}` set + accepted constants hash
  + epoch coordinate; bytes under `testdata/harness/content-dynamics/bundles/<full-hash>/`;
  the read-only loader verifies set equality, every per-file hash, the recomputed bundle hash,
  and epoch acceptance before execution; later mints add entries, never reinterpret; the baseline
  guard governs all five surfaces with adversarial missing/extra/tampered fixtures.

## Changelog (EH-C8/C9 round)

- 2026-08-10: EH-C8/C9 accepted; the opportunities-artifact requirement routed into the T0–T1
  candidate set; the runner is implementable now (registry empty until the next mint).

## Epoch-7 mint-order blocker (Codex, 2026-08-11 — EH-C10)

EH-C8/C9 require the first production registry entry and immutable bundle snapshot only **after**
an epoch pins Opportunities; `GenerateRegisteredContentSnapshots` correctly refuses to generate a
snapshot for an epoch absent from the accepted seed. The current runway instead asks for that same
registered composed run before the owner sign-off and mint. Both orderings cannot hold.

**Proposed contract:** add a candidate-only content-dynamics report mode, analogous to FCE5.3,
which loads the complete ratified promotion manifest without registering a historical golden. The
owner reads that report before sign-off. The mint then makes epoch 7 authoritative; a separate
`BALANCE-CHANGE:` baseline commit snapshots the accepted epoch, adds the first registry entry and
golden, and receives its own designated review. Alternative: move the content-dynamics evidence
after mint and explicitly remove it from the pre-sign-off gates. Do not fabricate an epoch-7
registry coordinate before epoch 7 exists.
