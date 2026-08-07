# RFC: The First Content Epoch (epoch 6 — the owner-gated production mint)

- **Status:** draft — **owner-gated.** The mint commit requires Marco's explicit sign-off; this RFC
  exists so that sign-off approves a fully enumerated, precondition-checked change instead of a
  judgment call. Named successor of TP-C18 (The Pitch, ruled option a: fixture-first) and SR-C13
  (Soul Recovery, ruled: fixture-only now, minted together here).
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** none new — this RFC mints already-designed, already-ruled content. Balance law:
  `CLAUDE.md` (declarative data, hardcaps, formula shapes exact / constants are config).
- **Depends on:** every fixture-first content foundation — Meters + Achievements, Doctrine,
  Minigame Platform, Pet Care, Fiscal Quarters, Soul Foundation, Soul Recovery Activities, The
  Pitch — each at its own designated-review/archival gate (see FCE5).
- **Planning:** `planning/first-content-epoch/` (once implementing)

## Summary

The live epoch registry (`balance/epochs/phase0.json`, `current_epoch_id: 5`) carries only the 7
base artifacts; every content system since — meters, achievements, doctrines, minigames, pets,
fiscal, soul (including recovery activities), pitch — shipped **fixture-first**, with its reviewed
artifact bytes living in test fixtures and its production mint explicitly deferred to this RFC.
This RFC defines **epoch 6**: one dependency-complete mint that promotes the reviewed fixture bytes
into production balance artifacts, registers the epoch, and activates the content under the
already-shipped activation laws. It authors **zero new mechanics and zero new bytes** — it is a
promotion, its blast radius is the epoch registry, and every gate that made "reviewed" mean
something applies to it.

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
pets → minigames · fiscal → pets · soul → fiscal · pitch → soul
```

An epoch carrying Pitch therefore **must** carry all 8 optional artifacts. Epoch 6 is exactly that:
the 7 base artifacts (bytes unchanged) + the 8 content artifacts, loaded as one bundle under one
new constants hash. TP-C18's "no partial chain" ruling is thereby structural, not procedural.

### FCE2 — Byte provenance: promotion, never authorship

Each content artifact's production bytes are **byte-identical to the reviewed fixture bytes** that
passed that system's content gate and designated review (e.g. `testdata/minigame/pitch-v3.json`,
`testdata/soul/recovery-activities-v1.json`, `balance/testdata/fiscal-foundation-v1.json`, and the
meters/achievements/doctrines/minigames/pets equivalents — the implementing plan enumerates the
exact source-fixture → production-path table with hashes). Production paths follow the repo
convention: `balance/<kind>/first-content.json` (exact names implementer-chosen, one file per
artifact kind, registered in the epoch registry's `artifacts` list).

**Any retune between review and mint is a new review.** If a provisional constant is changed from
the reviewed fixture bytes, the changed artifact re-runs its content gate and the change lands as
its own `BALANCE-CHANGE:` commit reviewed before the mint consumes it. The default is: mint the
provisional bytes as reviewed, retune later via epoch 7+ (retunes are cheap; the mint is what
unblocks playability).

### FCE3 — Registry mechanics

One mint commit (subject class `BALANCE-CHANGE:`) that:
1. adds the 8 production artifact files (FCE2);
2. appends the 8 rows to `artifacts` and the epoch-6 entry to `epochs` in
   `balance/epochs/phase0.json` — `{epoch_id: 6, name: "First Content", changelog_ref:
   "changelog/epoch-6.md", accepted_hashes: [<the new constants hash>]}` — and bumps
   `current_epoch_id` to 6;
3. writes `balance/changelog/epoch-6.md` enumerating every artifact, its source fixture, its
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
- **Founder-scoped** (minigames, pets, fiscal, soul, pitch): activates on epoch adoption per each
  foundation's shipped rules (pet starter creation begins at the first pet-carrying epoch; fiscal
  accrual, soul meter, minigame/pitch start-eligibility follow their archived activation clauses).
- **Leaderboards:** board binding across the epoch boundary follows the archived Leaderboards &
  Balance Epochs law unchanged; the mint itself makes no board decisions.

### FCE5 — The mint gates (ALL green before the mint commit exists)

1. **Review-complete (reformulated 2026-08-07 after the foundation audit — the original "archived
   before the mint" wording was CIRCULAR for mint-blocked foundations and is superseded):** for
   every ARTIFACT-CONTRIBUTING foundation (the 8 artifact owners), (a) its implementation range is
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
4. **Copy coverage:** every copy key declared by the 8 artifacts ships through `verify-copy` green.
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
   all 15 artifacts; the promotion table (fixture hash = production hash) verified by test.
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
