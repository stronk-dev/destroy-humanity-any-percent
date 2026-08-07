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

1. **Review-complete:** every contributing implementation has its designated cross-party verdict
   filed and its RFC archived — as of drafting, The Pitch (remediation in flight) and Soul Recovery
   (handoff pending) are the open items; the changelog cites every consumed verdict.
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

## Open questions (owner rulings wanted at acceptance)

1. **Epoch name** — recommend the mechanical `"First Content"` (matches "Faction Foundation" /
   "Canonical Boards"; flavor lives in data files, not registry names).
2. **Provisional bytes** — recommend minting the reviewed provisional constants as-is (FCE2
   default) rather than a pre-mint retune pass; epoch 7 is the retune lane once the composed
   harness report and real play data exist.
3. **Mint timing vs the playability seam** — the mint does not depend on Minigame API & Surface
   (internal services consume the bundle either way); recommend minting as soon as FCE5 is green
   rather than waiting for the API surface, so surface work tests against production content.

## Changelog

- 2026-08-07: created (draft) — the owner-gated dependency-complete mint promised by TP-C18 and
  SR-C13; scope pinned to the loader's enforced artifact chain; gates enumerated.
