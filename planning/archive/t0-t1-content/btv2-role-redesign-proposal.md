# T01-C28 proposal: Beige Tower v2 role redesign

- **Status:** RULED 2026-08-14 — owner signed the CLEAN REMOVAL variant, not the recommendation:
  provision role REMOVED, `synergy_feed → pool.institutional_knowledge` becomes the sole
  non-production role; copy enters owner adoption; consequences recorded in `log.md`. Exact
  literals still return after measurement.
- **Author:** Claude (design lane, per the 2026-08-14 C28 ruling). Measurements: Codex.
- **Ruling constraints honored:** engine unchanged (fractional-provision arm REJECTED); redesign
  within the shipped role vocabulary (`provision`, `synergy_feed`, `manual_output`, `stock_rate`)
  or a justified composition; owner signs before any byte changes.

## The defect, in one paragraph

`generator.beige_tower_v2` (T1, price `3.5831808e8`, production `4.90222789063e5` cash/s) has one
non-production role: provision `generator.beige_tower` at `100000` ppm — 0.1 tower per source per
60 s grid, each provisioned tower producing `1e0` cash/s. The sidekick output is ~10 orders of
magnitude below the hero's own production; Codex's sweeps (price to 1e8, production to 12x,
synergy grafts, provisioning to 100,000x) all leave the worst-case relevance delta at zero. No
literal fixes this role at this tier scale. First validly measured verdict on the row
(2026-08-14, post-C26/C27 instrument).

## Recommended redesign: ADD `synergy_feed`, KEEP `provision`

Add one role row:

```json
{ "kind": "synergy_feed", "pool_id": "pool.institutional_knowledge", "per_count_ppm": <MEASURED> }
```

and change nothing else about the generator: provision row, production, price, and copy all stay.

**Why this shape:**

1. **It is the tier's own pattern.** Every other T1 generator feeds `pool.institutional_knowledge`
   (garage_rack 2000 ppm, crt_wall 3000 ppm, legal_dept 4000 ppm). BTv2 — the capstone that is
   literally a tower of accumulated corporate history — is the only T1 generator missing from the
   tier's knowledge pool. Thematically it belongs there more than any of them.
2. **The ratified copy survives untouched.** BTv2's owner-ratified description ("Procurement now
   orders towers by itself… New Beige Towers just arrive") is written around provisioning.
   Deleting the provision role would force an owner copy adoption round; keeping it as flavor-true
   (if numerically small) mechanics avoids that entirely.
3. **Doctrine and coverage stay whole.** design/02 §11b family 1 (kinetic chain,
   purchased/generated split) keeps its only shipped edge; the C23 role matrix keeps its only
   `provision` fixture row (coverage would otherwise become debt, legal_dept-precedent); AC0's
   provisioning-active premise for the mint stays true. The matrix grows 11 → 12 rows (BTv2 gains
   a second row); the pinned count in `TestT0T1CandidateRoleActivations` updates in the same
   change with the fixture that proves the new row activates and masks off.
4. **Smallest honest diff.** One added role row plus one added pool-source row; the ratified
   candidate lane reopens for exactly the BTv2-related bytes and nothing else.

**Recorded design consequence (not solved here):** unit-per-grid provisioning of a fixed-rate
lower-tier target cannot carry relevance across a ~5×10⁵ intra-tier scale gap, because provisioned
counts earn no per-count multipliers (the purchased/generated split working as designed). The
chain family will need target-side scaling (upgrades/multipliers applying to generated counts'
production, or higher-value provision targets) when T2+ content is drafted. This belongs to the
future T2 content RFC, not to C28.

## Measurement plan (Codex, under the ratified instrument)

1. Sweep `per_count_ppm` for the new row over a coarse grid (suggest 2,000–50,000 ppm; the tier's
   existing feeds are 2,000–4,000, but BTv2's max owned count is far lower — the capstone is
   bought ~1–3 times, so its per-count weight should plausibly be highest in the pool).
2. **Primary acceptance:** worst-persona relevance delta ≥ the 1,000-ms floor with trap floor
   passing, under both the T1 main path and (if the main path cannot exercise it) the C27 branch
   proof — same evidence discipline as the signed tuple.
3. **If no ppm in the swept range passes:** the fallback lever is BTv2's price (earlier arrival =
   more owned time inside the window), swept within `[1e8, 3.5831808e8]` as ONE atomic revision
   of the BTv2 row (price + requirement + ppm together). No other row moves. If THAT cannot pass,
   C28 comes back as a finding with both sweeps recorded — not a stretched literal.
4. Whole-path pacing must not regress (T0 419,462 ms; screened T1 2,724,115 ms baselines), and
   the refreshed branch reports must stay empty-failure.
5. Exact literals, reports, and the re-derived economy hash return for owner sign-off through the
   standard designated review.

## Alternatives considered and rejected

- **Drop `provision`, synergy only:** invalidates ratified copy (owner adoption round), orphans
  the C23 provision fixture, removes the doctrine's only shipped chain edge, and weakens AC0's
  premise. Strictly more cost than keeping the row.
- **`stock_rate` (faction stock contributor):** a differentiated job, but structurally invisible
  to the main relevance path (no faction context there — the exact C23 lesson), so its relevance
  evidence would rest permanently on fixtures plus C26 disclosure. A capstone generator whose
  measured job cannot appear in the measured trajectory is the wrong shape for the tier's most
  expensive purchase.
- **`manual_output`:** collides with `first_hire`'s niche (20,000 ppm, same tier) and reads
  thematically wrong (nobody clicks a beige tower).
- **Fractional-provision engine arm:** REJECTED by owner ruling 2026-08-14; not revisited.
