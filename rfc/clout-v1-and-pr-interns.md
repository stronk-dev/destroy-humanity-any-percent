# RFC: Clout v1 — Achievements into PR Interns (the run-local attention stack)

- **Status:** accepted — owner batch acceptance 2026-09-25; implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/07` Phase 1 ("Clout v1 (achievements + PR Interns)"); `design/02 §6`
  (the carry rule, the two feeds, the tiered 2/4/8 grant amendment of 2026-08-07, the PR Intern
  conversion `(1 + Clout × factor)`), `design/02 §6b` (the Gaia law: exactly one Clout mint),
  `design/02 §2.2` (the production stack and the CloutStack position), `design/02 §2c.3` (the
  burn/provenance/possession law), `design/02 §11` and `§11b` (pacing targets and the relevance
  doctrine), `design/05 §2` (feed prominence scales with Clout, which is a reach surface),
  `design/10 §2` (every build has an anchor), `design/11 §2–3` (progressive disclosure, the
  Founder card), `design/08 §1` (voice rules; copy keys only here), and
  `design/research/cookie-clicker.md §1.4` (the milk/kitten source mechanic)
- **Depends on:** Achievements Foundation (archived; `docs/achievements.md`); Purchasable Content
  Foundation (archived; `docs/purchasable-content.md`); Production Engine (`docs/production-engine.md`);
  Save Layer (`docs/save-layer.md`); Relevance Harness (archived); Copy Pipeline (archived);
  API Foundation (accepted, implementing) for the generated snapshot client. For the
  third PR Intern row only: `rfc/tier2-content.md` (new-draft), which owns the T1→T2 gate ID.
- **Parent / amends:** Follow-up to the archived `rfc/archive/clout-and-achievements-foundation.md`.
  That RFC's C1/C2 rulings deferred "the PR-Intern run-local multiplier" to a successor RFC, and this
  is that successor. It extends the shipped achievement hook, the economy schema (v4 → v5), the
  multiplier slot registry and the Company save (v18 → v19). It does not reopen any archived ruling.
- **Supersedes / superseded by:** —
- **Release manifest row:** `rfc/v0.1-garage-release-manifest.md` G09, and OQ-5/VQ-5/E-3
- **Planning:** `planning/clout-v1-and-pr-interns/` (once accepted)
- **Evidence coordinate:** HEAD `c2d9bbc`; balance epoch 8; `LatestCompanyVersion = 18` and
  `LatestFounderVersion = 21` (`server/save/state.go:25–26`). This is a static trace. No product
  code was run for this draft.

## Summary

`design/07` puts "Clout v1 (achievements + PR Interns)" in v0.1. Four pieces are already shipped:

- the achievement engine;
- the non-spendable `achievement_score` with 2/4/8 grants and 12 pinned definitions;
- a Founder `clout_lifetime` field that nothing writes;
- a neutral seam for a future run-local multiplier.

The PR Interns do not exist. Before anything can be built, one design conflict has to be ruled.
`design/02 §6` still says achievements grant Clout. `§6b` says Clout has exactly one mint, which
is social activity. The archived Achievements C1 ruling followed §6b.

This RFC:

1. puts that conflict to the owner as **OD-1**, with three options and a recommendation;
2. specifies PR Interns as **one-time upgrades that multiply all production by
   `(1 + x × factor)`**, where `x` is a **run-local input axis** that the owner's OD-1/OD-2 ruling
   selects;
3. makes every part of the design work under any of those options, through one artifact-level
   parameter.

The recommended path needs no Clout mint in v0.1. Under it, PR Interns scale with **achievements
attained this run**, which gives exactly the Cookie Clicker milk loop. Clout stays social-only and
reach-only, as both binding §6 laws require.

## Motivation and scope

**Why now.** The release manifest lists G09 as UNOWNED (manifest §1a) and puts this RFC in its
Batch B, after VQ-5. Server Garden's "Clout-drop payload" edge (manifest §3) and the
Clout/achievement reader slice of `garage-player-surfaces` both wait on this contract.

**In scope:**

- the input-axis parameter;
- the run-local attainment set, if OD-2 selects it;
- the PR Intern catalog shape, the unlock condition and the multiplier slot;
- save v19, the events and receipts, and the snapshot surface contract;
- the harness and relevance impact;
- PROPOSED numbers, as retunable data.

**Out of scope:**

- **The social Clout mint** (posting, viral events, the podcast circuit). It is a DESIGN-GAP
  (DG-4). This RFC names its seam but does not specify it.
- **Every Clout reach consumer:** feed prominence (`design/05 §2`), lobbying checks, canonization
  events (`design/08 §5`), and the lanyard invitations (`design/08 §4`). All of them are v0.2 or
  later.
- **Any Clout arm in a reward union.** Server Garden drops, board-game payouts and pet-battle
  payouts are out; see Deviation D5.
- **Player-facing prose.** Copy keys are named here. Their text is owner-authored and pending
  (evidence discipline 6).
- **Rendering.** `garage-player-surfaces` owns the Svelte surface. This RFC owns only the
  producer/snapshot contract that the surface consumes.

## OD-1 — The Clout mint conflict (the first owner decision; blocks acceptance)

**The conflict, with evidence.**

- `design/02 §6`, feed 1: "achievements grant Clout on a **tiered scale — 2/4/8**". This line was
  amended on 2026-08-07 in the FCE-C3 round.
- `design/02 §6b`: "**Clout has exactly ONE mint (social activity through the declared decay
  curve).** Nothing — no product, no event, no reward table — emits Clout outside it."
- Archived Achievements **C1** (accepted 2026-08-03) followed §6b: "achievements mint a separate
  permanent non-spendable `achievement_score`". The Minigame Platform **C1** made the same ruling.
  `docs/achievements.md` line 3 records that the shipped code "does not mint Clout".
- The **FCE-C3** ruling (`rfc/first-content-epoch.md`, 2026-08-07) ratified tiered 2/4/8 as
  `achievement_score` grants. The flat alternative was described there as "the literal design law
  with only the currency name changed". The later amendment to the §6 text still says "Clout".
- `design/02 §6` "Conversion" and the carry rule both key PR Interns on "**Clout earned this
  run**".

Both texts cannot be true at once. PR Interns need a numeric input, and which one it is depends on
this ruling.

**One more fact the ruling must weigh (found while drafting).** The shipped `NewlyEarned` path
excludes any ID already in the Founder lifetime set (`server/production/foundation_transition.go:62`;
Achievements CA4). As a result, **`achievement_score_run` shrinks as a founder's career grows.**
Run 1 can score the full run-scoped catalog, but a veteran whose achievements are all
lifetime-owned scores 0. Any PR input built on first-earn grants therefore makes veterans slower,
because it reads lifetime state and turns it into a production penalty. That violates the spirit of
the carry rule in the opposite direction ("famous, not fast" also means "not slow"). This applies
to option B as well, because achievement-minted Clout is also first-earn only.

**Options.**

- **Option A — Achievements feed PR Interns through a run-local attainment axis. Clout stays
  social-only. (RECOMMENDED.)**
  - Achievements mint no Clout. The C1 ruling and the Gaia law stand unchanged.
  - The PR Intern input is `achievement_attainment_run`: the summed `score_grant` of every
    **run-scoped** catalog achievement whose condition, with its proof, has been satisfied during
    this run, whether or not the Founder already owns it (CV2).
  - Clout (`clout_lifetime`) remains the reach axis and is minted only by the future social RFC.
  - Why this option:
    - It keeps both binding laws exactly: one mint, and no lifetime state in production.
    - It reproduces the source mechanic faithfully. Cookie Clicker's kittens scale on
      achievements, not on an influence currency.
    - It makes "optional weirdness load-bearing" on every run, not only on run 1.
    - It needs no change to `design/02 §6b`.
  - Cost:
    - The §6 "Conversion" and carry-rule texts must be reconciled by the owner to say "achievement
      attainment this run" where they say "Clout earned this run" (Deviation D1; discipline 5).
    - "Clout v1" ships with a Clout number that stays at 0 until the social mint exists (OD-4).
- **Option B — Amend §6b so that achievements are a second, catalog-bounded Clout mint.**
  - `design/02 §6b` is amended to "exactly two declared mints: social activity, and first-earn
    achievement grants (finite, latch-once, catalog-bounded)".
  - Each first earn credits `clout_run += score_grant` through the sole Clout owner (CV6). PR
    Interns key on `clout_run`.
  - Why this option: it matches the literal §6 feed-1 text and the design/07 wording.
  - Cost:
    - It reverses a binding law that was ruled three times (Achievements C1, Minigame C1, and the
      §6b text itself). The C1 blocker requires that such a reversal "re-size the downstream
      sink/decay model before acceptance".
    - It inherits the veteran-shrink defect above. Making it re-attainable per run turns it into a
      per-run repeating faucet, which is exactly what §6b forbids.
    - Server Garden, board-game and pet payouts then lose the structural argument against their
      own Clout arms.
- **Option C — Strict social-only: PR Interns key on social Clout this run.**
  - Achievements affect neither Clout nor PR Interns. PR Interns key on `clout_run`, which only
    the social mint writes.
  - Why this option: it follows the literal carry-rule text and §6b with no reinterpretation.
  - Cost:
    - PR Interns are inert until the social mint ships, and that mint's decay curve is an open
      DESIGN-GAP (DG-4). v0.1 would then need that RFC too, or would have to ship dead content,
      which the relevance gate forbids.
    - It drops CC's "highest-leverage idea" (`design/02 §6`: making optional weirdness
      load-bearing) from the production stack.

**Recommendation: Option A with OD-2 = attainment.** The rest of this RFC is written so that all
three options can be implemented by setting `axis_stack.input` (CV1) and turning on or off the
optional Clout sections (CV6, CV7). Options B and C need **no rework** beyond those switches and
the owner's design amendment.

## Specification

### CV0 — Parameterization table

| Ruling | `axis_stack.input` | Attainment set (CV2) | Clout ledger (CV6) | Clout mint sources in this RFC | Design edit owed by owner |
|---|---|---|---|---|---|
| **A + OD-2 attainment (rec.)** | `achievement_attainment_run` | **yes** | no | none | §6 Conversion + carry-rule wording (D1) |
| A + OD-2 shipped score | `achievement_score_run` | no | no | none | as above, plus the veteran-shrink defect accepted in writing |
| B | `clout_run` | no | **yes** | `achievement_first_earn` | §6b amended; sink/decay re-sized |
| C | `clout_run` | no | **yes** | none (the social RFC adds `social_post`) | none; DG-4 must be resolved first |

The loader rejects a bundle whose `axis_stack.input` does not match the sections activated for it.
For example, `clout_run` without the CV6 ledger is rejected.

### CV1 — PR Interns: one-time axis-scaled upgrades (economy schema v5)

**The mechanic, as the design gives it.** `design/02 §6`: "PR Interns (kitten equivalents) each
multiply production by `(1 + Clout × factor)`, stacking multiplicatively". The design does not say
whether they are upgrades, generators or automation, and it gives no purchase currency or unlock
rule. That is DESIGN-GAP DG-1/DG-2, put to the owner as **OD-3**. This section specifies the
recommended reading, the kitten lineage (`research/cookie-clicker.md §1.4`), under which each PR
Intern is **a single purchasable upgrade** with its own factor, unlocked at an input threshold and
bought with cash through the existing `buy_upgrade` intent. Nothing in the design makes a PR Intern
an automation or clicking mechanic. The Tier-1 `Intern` autoclicker in `design/01` is a different
object (see OD-7).

**Schema v5 changes.** These are additive. Every v4 row stays valid unchanged.

1. A new top-level object, `axis_stack`, which is required if and only if any upgrade carries an
   axis-scaled effect:

   ```json
   "axis_stack": {
     "input": "achievement_attainment_run",
     "input_cap": 44,
     "cap_reason_key": "cap.axis_stack_input"
   }
   ```

   - `input` is the closed enum `achievement_attainment_run | achievement_score_run | clout_run`.
   - `input_cap` is a positive exact-safe integer and a **visible hardcap** (law 5).
   - `cap_reason_key` is a Copy Pipeline key.
2. A second arm of `upgrade_effect`, identified by `slot`:

   ```json
   { "source_id": "upgrade.pr_intern_1.axis", "slot": "axis_stack", "target": "all", "factor_ppm": 25000 }
   ```

   `factor_ppm` is a positive exact-safe integer of at most 1,000,000. The `target` is always
   `all`, as the design says: "multiply production". The static v4 arm keeps its
   `slot: "upgrades"` constant. An upgrade whose effects mix the two arms is rejected.
3. A new `condition` arm, `{ "kind": "axis_at_least", "minimum": <exact-safe int ≥ 1> }`. It is
   legal **only** inside an upgrade's `requires`. Route predicates, gate requirements and every
   other user of the condition union reject it, so the Route Registry's context version does not
   change. It reads the same clamped axis value as CV3.
4. Each axis-scaled upgrade declares a `multiplier_sources` row
   `{id:<source_id>, slot:"axis_stack", target:"all", provider:"axis_stack"}`.
5. Axis-scaled upgrades still carry ≥1 typed role (`design/02 §11b` law 4). The proposed role is
   `synergy_feed`, which is the role every shipped upgrade row carries. This trace has not checked
   whether an axis-scaled upgrade can carry that role *truthfully*: roles describe executed
   mechanics (`docs/purchasable-content.md`). If it cannot, it must not be mislabelled, and OD-3
   decides between (a) that role with a real pool binding and (b) a role-vocabulary extension by
   this RFC.

**Cross-artifact validation.** Both Go and TypeScript check the following.

- For the two achievement inputs, the bundle must pin the `achievements` artifact.
  `input_cap` must be ≥ the maximum reachable input:
  - for `achievement_attainment_run`, the sum of `score_grant` over `condition_scope: run` rows;
  - for `achievement_score_run`, the sum over all rows.

  The clamp in CV3 is then provably unreachable for achievement inputs. It still executes
  uniformly, as defence in depth, and a fixture drives it with a seeded out-of-bounds state.
- For `clout_run`, the CV6 ledger must be active, and `input_cap` must equal the ledger's declared
  `clout_run_cap`.

**Purchase.** Purchase uses `buy_upgrade` exactly as shipped:

1. accrue;
2. check the window, and check `requires` (including `axis_at_least`) against accrued state;
3. make the exact ledger debit;
4. record ownership;
5. emit one `upgrade_purchased`.

There is **no new client intent**. If rebought, the upgrade gets the existing `not_eligible/owned`
rejection. An `axis_at_least` shortfall is `requirement_not_met`.

**The axis never decrements within a run** (CV2). A PR Intern that has been bought can therefore
never become "unmet". Ownership is permanent for the run and resets at Exit with every other
upgrade.

### CV2 — The run-local attainment set (only when the input is `achievement_attainment_run`)

`achievements_attained_run` is a sorted-unique set of achievement IDs on **Company** state.
`attainment_score_run` is derived from it: the sum of `score_grant` over the set, under the run's
pinned achievements artifact. The stored value is validated against that derivation and a mismatch
is rejected, following the C3 pattern.

- **Evaluation site.** The existing achievement hook, which runs after Meters, gets a second pass.
  Both passes run against the **same** pre-achievement snapshot and the same observations.
  - The first pass is the shipped `NewlyEarned` pass, unchanged.
  - The second pass selects definitions where all three hold:
    1. `condition_scope == run`;
    2. the ID is not already in `achievements_attained_run`;
    3. the condition holds, and `achievementProofSatisfied` holds for the same action debits and
       event batch.

    It **does not read** `founder.AchievementsEarnedLifetime`. Staged IDs commit together, in
    byte order.
- **Relationship to earning.** Every ID newly earned this run with `condition_scope: run` is also
  attained in the same transition. The loader-level invariant is
  `achievements_earned_run ∩ run-scoped ⊆ achievements_attained_run`, and both runtimes reject any
  state that violates it.
- **Career-scoped definitions never attain.** Their conditions observe Founder lifetime counters.
  Letting them feed production would carry lifetime state into the stack, which the carry rule
  forbids.
- **Latching.** An ID attains at most once per run. If its condition oscillates, it grants nothing
  further. Rejected intents never evaluate. Evaluation also runs at the terminal boundary, for
  history completeness only, because the set is discarded at Exit.
- **Not a currency, not a mint.** The set does not settle to Founder, has no debit API, and is
  never imported by the Clout owner or by any reward union. `achievement_score` (run and lifetime)
  and the 2/4/8 grants are unchanged.

### CV3 — The multiplier: formula, slot, timing

For each owned axis-scaled upgrade `i`, with `x = min(input, input_cap)`:

```text
factor_i = (1,000,000 + x × factor_ppm_i) / 1,000,000        // exact integers, one Decimal quantization
axis_stack_product = Π_i factor_i                           // applied per contribution; see slot rule below
```

- **Slot.** A new slot, `axis_stack`, goes into `server/multiplier`'s ordered registry **after
  `milestones` and before `faction`**. This is the position closest to `design/02 §2.2`, where
  CloutStack comes directly after the generator-internal terms. Contributions inside the slot
  apply in raw-byte `source_id` order, as in every other slot. The slot is multi-provider and has
  exactly one provider ID, `axis_stack`. OD-6 confirms the position.
- **The input is Company-local.** The provider reads only Company state and the pinned bundle.
  Under the recommended input it performs **no Founder read at all**, which is the structural proof
  of the carry rule (AC3).
  - Under `achievement_score_run` it reads the shipped field, which is Company-local but depends on
    Founder carry through `NewlyEarned`. AC3 documents that dependency as the known defect.
  - Under `clout_run` it reads the CV6 Company field.
- **Timing.** The input changes only in the achievement hook (or the CV6 mint), and that hook runs
  after accrual. The rate is therefore piecewise constant between applied transitions. A new input
  value takes effect from the **next** accrual interval, exactly like a newly bought generator.
  Offline catch-up uses the constant input frozen at the cursor. Splitting one interval across
  evaluations cannot change the result (AC6).
- **Visibility and publication.** `docs/generated/production-formulas.json` is regenerated in a
  separate reviewed commit. It shows:
  - the slot and its position;
  - the `factor_i` formula;
  - the pinned input kind and cap;
  - the maximum product at the cap for the pinned catalog.

  `make formulas-check` fingerprints the new provider as a rate authority.
- **Read projection.** The shipped pure projection (`docs/production-engine.md` §Catalog and
  multiplier boundary) gains the axis-stack contributions through the same private functions, so
  UI rates and live rates cannot diverge.

### CV4 — Save v19 (Company) and activation

- **Company v19** adds `achievements_attained_run` (a sorted-unique array, empty at run start) and
  `attainment_score_run` (an exact-safe int, derived and validated). These fields are required at
  v19 and rejected before v19.
  - Under the `clout_run` input, v19 instead adds `clout_run` (an exact-safe int in
    `[0, clout_run_cap]`). The two field groups are mutually exclusive, as selected by the pinned
    economy artifact.
  - Under `achievement_score_run`, **no save version change** is needed.
- **The Founder axis does not change** under any option. `clout_lifetime` already exists at every
  Founder version as an exact int (`validatePrestigeState`).
- **Activation.** Activation is new-run-bound under the Achievements **C11 law**. A run pinned to
  a pre-v5 economy artifact finishes on its pinned semantics: no attainment, no axis slot, no PR
  rows. The first Exit into an epoch whose pinned economy artifact carries `axis_stack` assembles
  Company v19 in the new-run transaction:
  - an empty attainment set, or `clout_run = 0`;
  - `attainment_score_run = 0`.

  Ordinary writes then require the pinned v5 economy artifact and the paired achievements
  artifact. Nothing attains retroactively from pre-activation history.
- **Exit.** At Exit the attainment set and the score are **discarded**: the next run starts empty,
  and nothing settles to Founder. Terminal state records the final values for replay. Under
  B/C, CV6 settlement applies instead.
- **Migration corpus.** `testdata/save-migrations.json` gains the following cases, and the
  required case count ratchets through the reviewed baseline manifest:
  - `company-v18-to-v19-new-run`;
  - `company-v19-derived-mismatch-rejected`;
  - `company-v19-field-before-version-rejected`;
  - `company-v19-attained-not-superset-of-earned-rejected`;
  - `pre-activation-run-replays-through-exit`.
- **Replay inputs.** No new frozen input is needed under the recommended input, because
  attainment is a pure function of Company state, the action and the pinned bytes. The kernel
  version bumps because the hook's behaviour changes. `ApplyLogged` in Go and in TypeScript
  executes the second pass.

### CV5 — Events and receipts

- **`achievement_reattained.v1`** `{run_id, achievement_id, score_grant}`. It is emitted in byte
  order for every ID that attains this run **without** also being newly earned in the same batch;
  newly earned IDs keep their existing `achievement_earned.v1`. The two kinds never duplicate one
  ID in one transition.
  - It is registered in the closed event registry.
  - A Postgres migration extends the event-kind constraint for the new kind at v1 only (the
    `00067` pattern).
  - Applied migrations are append-only.
- **`upgrade_purchased`** is unchanged for PR Intern purchases.
- **The authoritative receipt snapshot** (applied receipts carry the complete snapshot) gains the
  new Company fields, plus a derived `axis_stack` object with this shape:
  `{input_kind, input_value, input_cap, cap_reason_key, saturated, contributions:[{source_id, upgrade_id, factor}], product}`.
  Decimals are canonical strings and integers are exact.
- **Exit / `run_ended`.** No payload change under the recommended input, because nothing
  persists. Under B/C, `run_ended` gains `clout_settled` (CV6), which the Founder card
  (`design/11 §3`) renders as the Clout delta.

### CV6 — The Clout ledger (only under OD-1 = B or C; inert under A)

- **Sole owner.** A new package, `server/clout` (TypeScript: `client/src/clout/`), exports the
  **only** transition that writes `clout_run`. The source registry is a closed enum.
  - Under B it holds `achievement_first_earn`, and the achievement hook calls
    `clout.Mint(state, achievement_first_earn, id, score_grant)` for each newly earned ID.
  - Under C this RFC registers no source. The social RFC adds `social_post`.
- **Hardcap.** `clout_run` is capped at `clout_run_cap`, a visible hardcap with its own
  `cap_reason_key`. It saturates atomically. The value is PROPOSED by the owner (OD-5), because it
  depends on the mint.
- **Event.** `clout_minted.v1` `{run_id, source, source_ref, amount_requested, amount_credited,
  clout_run_after, saturated}`.
- **Settlement.** Settlement runs inside `ApplyExitTransaction` under the existing founder→company
  lock order. It computes `clout_lifetime = min(clout_lifetime + clout_run, MaxExactInteger)`,
  records `clout_settled` in the Exit receipt, and starts the new run at `clout_run = 0` (the
  prestige D6 rule that this-run Clout is 0).
- **Gaia enforcement (all options).** Nothing mints Clout in v0.1 under Option A.
  - A recursive import-graph and source-registry test proves that `server/clout` is the only
    writer of `clout_run` and `CloutLifetime`, apart from the Exit settlement that the package
    itself exports.
  - Every closed reward/effect union (minigame payouts, fiscal, opportunities, the Commons)
    rejects a `clout` arm.
  - A seeded second writer must fail the test (AC4).

### CV7 — The social-mint seam (named, not specified)

`design/02 §6` feed 2 says "posting (a lightweight timed action), viral events, the podcast
circuit, thought-leadership upgrades". `§6b` says the mint runs "through the declared decay curve".
The research gives a Gaia first-post-of-day bonus decaying 50 → 3
(`research/gaia-hyperinflation.md §1`), but the design declares no curve, no timer and no intent
grammar. That is **DG-4**. This RFC fixes only how the future social RFC connects:

- it registers `social_post` in CV6's closed enum;
- it writes only through `clout.Mint`;
- it inherits `clout_run_cap`.

It does not specify the posting action. The recommendation (OD-4) is a separate RFC once the owner
declares the curve.

### CV8 — Catalog content: PROPOSED numbers (retunable data; the harness selects the shipped values)

The numbers below are **proposals**, not ratified literals. They follow the FCE-C3 lane: the owner
ratifies complete artifact bytes by SHA-256, and the harness can move them.

**Sizing basis.** On the pinned epoch-8 catalog, the run-scoped grants sum to **44**:

| Achievement | Grant |
|---|---|
| `first_gate` | 2 |
| `generators_purchased_1` | 2 |
| `generators_purchased_25` | 4 |
| `generators_owned_100` | 4 |
| `generators_owned_300` | 8 |
| `gate_burn_t3` | 8 |
| `generators_purchased_25_tier_3` | 8 |
| `tier_5` | 8 |

That puts `input_cap = 44`. In v0.1's T0–T2 reach, a run can realistically attain `first_gate`,
`purchased_1`, `purchased_25` and, with play, `owned_100` (beige towers), which gives x ≈ 8–12. Run
1's scripted collapse at about 15 min typically leaves x ≈ 8.

| Row | Window | Cost | `requires` | `factor_ppm` | Value at x=8 / x=12 / x=44 (cap) |
|---|---|---|---|---|---|
| `upgrade.pr_intern_1` | `gate.t0_to_t1` → null | `1e6` cash | `axis_at_least 6` | 25,000 | ×1.20 / ×1.30 / ×2.10 |
| `upgrade.pr_intern_2` | `gate.t0_to_t1` → null | `5e7` cash | `axis_at_least 10` | 20,000 | — / ×1.24 / ×1.88 |
| `upgrade.pr_intern_3` | `<tier2-content T1→T2 gate>` → null | owner/harness (after T2 rows exist) | `axis_at_least 12` | 15,000 | — / ×1.18 / ×1.66 |

- `pr_intern_3` is **blocked on `tier2-content`**, because its gate ID and T2 cash scale do not
  exist yet. It lands in the same mint as T2 content, or it is dropped (OD-5).
- The maximum stack at the cap is ≈ ×6.6 for the three rows. The regenerated formula artifact
  publishes that bound.
- Magnitude note: `design/02 §6` calls this "the single biggest multiplier family in the game".
  The v0.1 values are deliberately modest, because the catalog has 12 achievements and not the
  ~600 target. `factor_ppm` shrinks as the catalog grows. That is data, not code.
- Under B/C these rows are unchanged except for `input_cap` and the thresholds. Those are
  re-proposed once a mint rate exists.

**Copy keys** (text pending and owner-authored; every key registers through the Copy Pipeline
before the balance commit, following FCE-C4):

- `upgrade.pr_intern_1`, `upgrade.pr_intern_2`, `upgrade.pr_intern_3`;
- `cap.axis_stack_input`;
- `axis_stack.panel.title`, `axis_stack.formula.caption`, `axis_stack.hint.first` (the one
  contextual hint), `axis_stack.tooltip.why` (the one "why this matters" line);
- `codex.axis_stack` (the `?` page);
- `achievement.reattained.notice`;
- under B/C only: `clout.run.label`, `cap.clout_run`, `clout.settled.founder_card`.

### CV9 — UI surface contract (producer side; `garage-player-surfaces` renders)

- **Snapshot.** `game_ui_snapshot` gains the CV5 `axis_stack` object, plus
  `achievements_attained_run` with an `earned_this_run` flag per ID. It flows through the API
  Foundation's generated client (manifest edge 3g). There is no hand-written parallel client, and
  the surface never computes a factor locally.
- **Rendering obligations.** The rendering slice must satisfy all of the following:
  1. PR Intern rows appear in the existing Desk upgrade list with their `axis_at_least` progress
     (`input_value / minimum`), their current factor, and a projected factor.
  2. One axis-stack readout shows the product, the formula and the visible cap, following the
     `§11b` rule that every pool's composition renders.
  3. The contextual hint fires once, when the first PR row's window opens (progressive disclosure,
     `design/11 §2`).
  4. `?` opens `codex.axis_stack`.
  5. Re-attainment notices are coalesced per transition, so they do not flood AT live regions
     (manifest EC-6).
  6. All text comes through copy keys, and no mechanical ID is ever rendered.
- **Accessibility.** The accessibility floor F03 applies.
- **Clout display under A.** The Founder card's Clout delta renders only through the OD-9 ruling:
  either a truthful `0` or withheld. It is never a fabricated value.

### CV10 — Harness impact

- **Scenario bundle.** Axis evaluation needs the pinned `achievements` artifact.
  - If the scenario loader's bundle does not already carry it, adding it is part of this RFC. The
    constants hash then changes and the change lands as a `BALANCE-CHANGE:` commit.
  - A scenario that pins a v5 economy without achievements is rejected. It must not run with a
    silently neutral stack (discipline 3).
- **Policies.** The near-greedy reference persona must price the dynamic factor at the input value
  current at purchase time. Chaos and Casual need no change; they buy through the ordinary
  upgrade path.
- **Relevance.** PR rows are upgrade rows, so the shipped ANY/ALL gate covers them. Two things
  must hold:
  - the simulation-only upgrade effect mask must null axis-scaled contributions;
  - each PR row must (a) cost ≥ε time to some persona when deleted inside its window, and (b) be
    bought by the reference persona.
- **Pacing.**
  - The shipped envelopes must stay green on the minted content: first generator < 60 s,
    T0→T1 6–8 min Chaos median, first elective Exit in [45, 90] min, and run 2 strictly faster.
  - If the multiplier moves any baseline, the baseline ratchets in a reviewed commit and never
    widens a bound (discipline 4).
  - A new **observation**, not a target: time to the first PR Intern purchase, as p50/p95 per
    persona.
- **Invariant.** A new invariant, `axis_input_within_cap`, is added. The registry grows by
  reviewed ratchet.

## Deviations from design

- **D1 (Option A).** PR Interns key on `achievement_attainment_run`. `design/02 §6` Conversion and
  the carry rule say "Clout earned this run". The reasons are §6b and the veteran-shrink defect in
  OD-1. The owner reconciles the §6 wording (discipline 5). This RFC does not edit `design/`.
- **D2 (Option A).** Achievements grant no Clout, contrary to the literal §6 feed-1 text. This
  follows Achievements C1. The 2/4/8 grants are `achievement_score` per FCE-C3.
- **D3.** PR Interns are one-time cash upgrades unlocked at an input threshold. The design does not
  specify their form (DG-1/DG-2; OD-3). The kitten lineage is followed.
- **D4.** The `axis_stack` slot is placed after `milestones`. Design §2.2 lists CloutStack before
  FounderBonus, but the shipped registry already puts `prestige` last. Multiplication is
  commutative apart from quantization order, so the position is disclosed in the formula artifact
  (OD-6).
- **D5.** `design/03` lists Clout drops, board-game wins and pet-battle wins as Clout sources.
  Under §6b every one of these is a second faucet, so this RFC exposes no reward-union Clout arm,
  and CV6 enforces that structurally. The manifest's "Server Garden Clout payload" edge becomes
  "no Clout payload" unless OD-1 = B is extended to those tables by an explicit owner amendment.
- **D6.** v0.1 values are far below the "single biggest multiplier family" ambition (CV8
  magnitude note). The formula shape is exact; the constants are data.

## DESIGN-GAPs

- **DG-1:** Is a PR Intern an upgrade, a generator or automation? This goes to OD-3.
- **DG-2:** What is the purchase currency and the unlock rule? This goes to OD-3.
- **DG-3:** First-earn grants shrink across a career, so a production input built on them punishes
  veterans. This goes to OD-2.
- **DG-4:** The social-mint decay curve, timer and intent are undeclared. This goes to OD-4.
- **DG-5:** `clout_run_cap` has no design value. This goes to OD-5 under B/C.
- **DG-6:** There is a name collision between "PR Intern", the `design/01` Tier-1 `Intern`
  autoclicker, and the shipped `generator.nephew_intern`. This goes to OD-7.

## Acceptance criteria

Each criterion names the failing case that must be demonstrated (discipline 1). Gate claims use
`-count=1`.

1. **Loader (Go and TypeScript).** Economy v5 loads the CV8 rows. Each of the following seeded
   defects must reject in **both** runtimes:
   - an axis effect with `slot: upgrades`;
   - `target` other than `all`;
   - `factor_ppm` of 0 or greater than 1e6;
   - an unknown `input`;
   - `input_cap` below the reachable maximum;
   - `axis_at_least` inside a route predicate;
   - a mixed-arm upgrade;
   - `clout_run` without the CV6 ledger;
   - an axis upgrade without its `multiplier_sources` row.

   *Failing case:* a loader that ignores `slot` on effects accepts the first defect, and the test
   must fail on that loader.
2. **Formula parity.** Shared Go-authored vectors cover:
   - x=0, where the factor is exactly `1e0` and the stack is neutral;
   - x at the threshold;
   - x at the cap;
   - a seeded out-of-bounds x, which clamps and reports `saturated`;
   - two and three interns, applied multiplicatively in raw-byte order.

   TypeScript reproduces them byte for byte. *Failing case:* an additive-stacking mutant, and a
   float-division mutant, must each mismatch at least one vector.
3. **Carry rule, structural.** Take two Company states that are byte-identical but have different
   Founder lifetime sets and `clout_lifetime`. Under `achievement_attainment_run` they produce
   identical rates, receipts and axis snapshots. *Failing case:* the same test run with
   `input = achievement_score_run` against a veteran Founder must show a divergence. This records
   DG-3 as a discriminating fact, not an opinion.
4. **Gaia law.** A recursive import-graph and source-registry test lists every writer of
   `clout_run`/`CloutLifetime`. The allowed set is the CV6 package only, or the empty set under
   Option A. Every reward/effect union rejects a `clout` arm. *Failing case:* a seeded fixture
   package that writes `CloutLifetime` from the achievement hook must fail the test.
5. **Attainment semantics** (sequential Go/TypeScript corpus). The corpus covers:
   - run 2 re-attains `first_gate` → `achievement_reattained.v1`, `attainment_score_run` +2,
     `achievement_score_run` unchanged;
   - an oscillating condition latches once;
   - a burn-proof definition without the same-batch debit does not attain;
   - career definitions never attain;
   - newly earned IDs emit only `achievement_earned.v1`;
   - the set resets at Exit.

   *Failing case:* a mutant that includes career definitions, or reads the Founder lifetime set,
   changes at least one vector.
6. **Timing and partition invariance.** A property test over seeded intent policies shows that
   splitting any interval across evaluations yields identical state. It also shows that an input
   change affects only intervals after the transition that caused it. *Failing case:* a mutant
   that applies the new factor retroactively to the triggering interval fails.
7. **Save v19.** The CV4 corpus cases all pass. Old runs replay through Exit on pinned semantics.
   New-run assembly is atomic. *Failing case:* a v18 state carrying `achievements_attained_run`,
   and a v19 state with a tampered `attainment_score_run`, must each reject.
8. **Events and persistence.** A real-Postgres integration test writes a PR Intern purchase and a
   re-attainment through service → store → Postgres, and asserts the exact rows. *Failing case:*
   inserting an unregistered kind, or `achievement_reattained` at schema version 2, is rejected by
   the constraint.
9. **Harness.**
   - The relevance ANY/ALL gate is green for every PR row.
   - All shipped pacing envelopes are green on the minted bundle.
   - The new observation and invariant are reported.

   *Failing case:* a fixture PR row with `factor_ppm = 1` must be reported dead. A scenario with a
   v5 economy and no achievements artifact must fail loudly rather than run neutral.
10. **Formula publication.** `production-formulas.json` is regenerated in its own reviewed commit.
    *Failing case:* changing the slot position without regenerating fails `make formulas-check`.
11. **Surface contract.** The snapshot carries the CV9 fields through the generated client. The
    composed browser witness renders a PR row's progress and factor. *Failing case:* the witness
    fails when the producer's `axis_stack` is severed (manifest EC-2 dimension 3).
12. **Canon.** The following are updated in the final change: `docs/achievements.md`,
    `docs/purchasable-content.md`, `docs/production-engine.md`, `docs/save-layer.md`,
    `docs/balance-harness.md`, and a new `docs/axis-stack.md` (or a section in an existing doc).
    The RFC index row and the manifest's G09 row are updated. Designated cross-party review covers
    the full range (CLAUDE.md gates a–c).

## Owner decisions (numbered; recommended defaults first)

- **OD-1 — The Clout mint conflict.** The options are A, B and C, as set out above.
  **Recommended: A.** Achievements feed PR Interns through the run-local attainment axis, Clout
  stays social-only, and §6b is untouched.
- **OD-2 — The input under A.** **Recommended: `achievement_attainment_run`**, re-provable per run
  and run-scoped only. The alternative is the shipped `achievement_score_run`, whose shrink across
  a career AC3 proves.
- **OD-3 — The PR Intern form (DG-1/DG-2).** **Recommended:** one-time cash upgrades through
  `buy_upgrade`, with an `axis_at_least` unlock and the `synergy_feed` role. The alternatives are:
  - repeatable generator-like hires, where each count multiplies. This deviates from the kitten
    model and is unbounded without a new cap.
  - a separate `pr_interns` artifact with its own intent. That means more surface and a new
    relevance row kind.
- **OD-4 — The social mint in v0.1.** **Recommended:** defer it to a separate `clout-social-mint`
  RFC, drafted after the owner declares the decay curve, timer and intent (DG-4). v0.1 then ships
  with Clout at zero, reach-only. The alternative is to specify posting here, which requires the
  owner's curve.
- **OD-5 — Numbers.** Ratify the CV8 rows by SHA-256 of the complete v5 economy bytes (the FCE-C3
  pattern) after the harness run. `pr_intern_3` either lands with `tier2-content` or is dropped
  from v0.1. Under B/C, set `clout_run_cap`.
- **OD-6 — Slot position.** **Recommended:** after `milestones`, before `faction`, and disclosed
  in the formula artifact.
- **OD-7 — Copy and naming.** The owner authors all CV8 copy text. Recommended: keep the "PR
  Intern" name, which is the design's own, and have its copy explicitly distinguish it from the
  shipped `nephew_intern` generator and the `design/01` autoclicker `Intern`. Alternatively, rename
  the PR row's display name. Mechanical IDs are unaffected either way.
- **OD-8 — Design reconciliation.** Under A, the owner edits `design/02 §6` (Conversion, carry
  rule, feed 1) in the same edit that records OD-1, so that no normative text contradicts the
  ruling. Under B, the owner edits §6b and re-sizes the sink/decay model.
- **OD-9 — Clout on the Founder card in v0.1.** **Recommended:** withhold the Clout line until a
  mint exists. The honest alternative is to render `0`.

## Open questions

All blocking questions are listed above as OD-1 to OD-9. OD-1 and OD-2 must be ruled before
acceptance, and the body reconciled in the same edit, with any sections made inapplicable by the
ruling deleted. OD-5's literal bytes are ratified at mint review, not at acceptance.

## Changelog

- 2026-09-25: created (draft — not implementation authority) from a static trace at `c2d9bbc`,
  for the v0.1 manifest's G09 / OQ-5.

## Owner acceptance (2026-09-25)

Marco accepted this RFC in the 2026-09-25 batch (the answer to a direct batch question in-session,
recorded by Claude). Every owner decision this RFC lists is **ruled at its stated recommended
default**. Where the text offers alternatives, the recommended option is the normative one and the
alternatives are rejected. Provisional numbers stay provisional and are ratified by harness
measurement and SHA, as the RFC already requires. Owner-authored copy stays owner-authored:
implementation ships copy keys, with any placeholder or candidate text clearly marked.
