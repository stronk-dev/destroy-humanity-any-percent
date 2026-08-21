# Coverage slice: events-narrative

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** Reconstructed on 2026-08-05; retained as
> evidence, not current status or execution authority. See `planning/platform-alignment/`.

Independently reconstructed from actual files 2026-08-05. Domain: the event engine (3 layers),
the content that rides on it, conspiracy, media canonization, speedrun framing, the endings, the
9-tier content ladder, the Cookie Clicker / Bakery Inc. easter egg.

**Load-bearing distinction for this whole slice: ENGINE vs CONTENT.** The events *engine* (one
data-driven evaluator, three layers) is the reusable machine; the events *content* (packs of
authored events, meters, situations, arcs) rides on it as data. They travel the four stages
separately, and every content system's furthest possible stage is capped by its engine layer's
stage. The slice keeps them as separate rows.

## Table

| # | System | R (research) | D (design) | F (RFC + real status) | I (docs) | Furthest | Flags |
|---|---|---|---|---|---|---|---|
| 1 | Events engine **Layer 1** (personal Paradox evaluator) | `events-playstyles.md §1a,1e` — specifies (Clausewitz DSL, MTTH, on_actions, hidden nodes, category locks) | `09 §1–2` — specifies (data format + required-primitives list, normative) | `rfc/events-engine-layer1.md` **status: draft** (2026-08-03) | — | **F (draft)** | STATUS-OFF (index) |
| 2 | Events **Layer 2** (EU4 pressure-meter disasters) | `events-playstyles.md §1a "Disasters"`, §1e Layer-2 block — specifies | `09 §3` — specifies (7-meter launch set) | none (meter *substrate* only, see row 2b) | — | **D** | GAP |
| 2b | Pressure-meter *substrate* (Trust ×10, p(doom)) | `events-playstyles.md §1` | `02 §7`, `09 §3` | `rfc/meters-foundation.md` **implementing (C1–C13 ruled)** — *mechanics only, disasters scoped OUT* | `docs/meters.md` | **F (impl.)** | belongs to economy-progression slice; note here |
| 3 | Events **Layer 3** (Helldivers server arcs: Situations + Major Orders) | `events-playstyles.md §1c,1e` — specifies (Galactic Impact Modifier, MO types) | `09 §4` — specifies | none | — | **D** | GAP |
| 4 | Conspiracy system (Pressure meter → spawn/spread/monetize/make-true) | `conspiracy-media.md THEME 1` — specifies richly (QAnon mechanics, Denver, self-sealing rule) | `08 §4`, `09 §3` (Conspiracy Pressure row) | none (meter not even in Meters catalog) | — | **D** | GAP |
| 5 | Media canonization / Clout milestone events (Ignore/Embrace/Sue + Streisand math) | `conspiracy-media.md THEME 2` — specifies (3-branch table, John Oliver/Streisand numbers) | `08 §5` — specifies (ladder + Streisand table) | none | — | **D** | GAP |
| 6 | Speedrun framing — **boards/categories/verification/epochs** (infra) | `speedrun-governance.md` (primary source) — specifies | `08 §6`, `05 §6` | `rfc/leaderboards-and-epochs.md` **implementing (L1–L7 ruled)** | — | **F (impl.)** | — |
| 6b | Speedrun framing — **chrome** (title bar, splits panel, PB/gold-splits, run-end retrospective, TAS mode) | `run-narrative-ux.md`, `speedrun-governance.md §0.7` | `08 §6`, `01 Tier 7` (TAS) | none | — | **D** | GAP |
| 6c | Skips as discovered content / **Route Registry** (named alt-gate predicates) | `speedrun-governance.md §0.9,§3.4` — specifies | `08 §6`, `05 §6` | archived: `gate-predicates-and-routes.md`, `route-registry-event-order-convergence.md` **implemented** | `docs/routes.md` | **I** | naming *ledger/analytics* deferred |
| 7 | The endings (Ascension / Long Decay / Both Graphs; Ethical%/LAN/training-data variants) | `billionaires-decay.md §2F` — specifies (verified datasets) | `01 §Endings` — specifies | Exit/reset only (`prestige-and-exits`); world-first only (`leaderboards`) | — | **D** | GAP |
| 8 | The **9-tier content ladder** (tiers as paradigm shifts) | `tier-relevance.md`, `gaming-enshittification.md §4.1` — specifies | `01-tiers.md` — specifies (all 9 tiers) | `rfc/t0-t1-playable-content.md` **draft** (T0–T1 only) | — | **F (draft, T0–1 only)** | STATUS-OFF (index); tiers 2–8 GAP |
| 9 | Cookie Clicker / **Bakery Inc.** tenant easter egg | `cookie-clicker.md` (teardown; backs the `1.15^n` gag + game-within-game) | `01 Tier 3` + `Tier 7` completion | none | — | **D** | GAP |
| 10 | Event **content packs** (authored events, the satire beats) | flavor corpus (`societal-satire.md` etc.) | `09 §2` (T0–T1 events pack) | inside `events-engine-layer1` E3 + `t0-t1` (copy only) | — | **F (draft)** | capped by engine |
| 11 | Event reward primitive (event scrip / exchange shop / plot prize shop) | `liveservice-idle-tier.md §3`, `neopets-social-history.md §1` | `09 §5`, `09 (Neopets adoptions)` | none | — | **D** | GAP |
| 12 | Seasonal arcs / dispatches / evergreen conversion | `events-playstyles.md §1e`, `neopets-social-history.md` | `09 §6` | none | — | **D** | GAP |
| 13 | GM operations / war log / watchdogs | `events-playstyles.md §1c,1e` | `09 §7` | none | — | **D** | GAP |

## Findings (evidence + adversarial checks)

### Furthest-stage tally
- **I (implemented):** 1 — Route Registry / skips backbone (row 6c).
- **F (RFC exists):** 4 — Layer-1 evaluator (draft), meters substrate (implementing), speedrun
  boards/epochs (implementing), tier ladder (T0–T1 draft only). Rows 6c aside, none of the
  *narrative* content is implemented.
- **D (design only, no RFC = GAP):** 9 — Layer 2, Layer 3, conspiracy, media canonization,
  speedrun chrome, the endings, tiers 2–8, the easter egg, event-reward shop, seasonal arcs, GM ops.
- **R-only / ORPHAN:** none material — every research dossier in the domain is carried into design.

The headline: **the entire narrative surface of the game is at stage D.** One evaluator is
drafted; everything the player would recognize as "the story" (events, meters-as-disasters,
conspiracy, canonization, the endings, tiers past the garage) has no contract.

### R→D honesty
Well-backed throughout; no `UNBACKED` design assertions found.
- Conspiracy (`08 §4`) and canonization (`08 §5`) map almost line-for-line onto
  `conspiracy-media.md` — QAnon drops/crumbs/bakers, self-sealing "suppression raises belief"
  rule (§1d), the Denver $8M gargoyle, Streisand's 6→420,000 downloads, Bob Murray's "Eat Shit,
  Bob" Emmy nod, the 2M:150 Area-51 ratio. The published-rule design law (law 9) is honored: the
  conspiracy "fighting it raises it" rule and Streisand math are stated as in-fiction published
  formulas, exactly as the research recommends.
- The endings (`01 §Endings`, `00 pillar 5`) rest on `billionaires-decay.md §2F`, whose figures
  were **re-anchored and verified 2026-07-28** (child mortality 22.3%→3.74%, extreme poverty
  49.1%→10.0%, trust 77%→17%) — the `DESIGN-GAP` that once propagated a mismatched-span error
  into `design/00` and `design/01` is recorded as CLOSED. This is the one place to watch: the
  correction note in `§2F` proves the design text was wrong before; a validator should confirm the
  current `01`/`00` text uses the ~1955 anchors (it does).
- Tier ladder eras cite `gaming-enshittification.md §4.1` and `tier-relevance.md`; the paradigm-
  shift-per-tier structure is a real research finding (Derivative/Kittens tier-relevance), not an
  assertion.
- No `ORPHAN`: even fine-grained research (the immortality lineage `conspiracy-media.md §1e`
  Gilgamesh Sleep-Test, Qin Shi Huang; the elite-society lanyard tiers §1c) is carried into design
  (`09 §2` Sleep Test idle check; `08 §4` lanyard midgame; `01 Tier 6` longevity ladder).

### D→F fidelity
- **Layer-1 evaluator (row 1)** — the one drafted content-engine RFC — declares a *principled,
  disclosed deviation* from `09 §1`, not a drift: `09` specifies Paradox wall-clock MTTH; the RFC
  determinizes it to a per-accrual-boundary hazard draw from the save-seeded stream
  (`events-engine-layer1.md` Summary + E1), because determinism law 2 forbids wall-clock RNG. It is
  called out in the Summary ("the one place we deliberately DIVERGE"). No hidden `DRIFT`, but note
  the RFC lacks the formal "Deviations from design" section RFC-0000 rule 4 prescribes — a minor
  process nit, not a fidelity failure. The RFC also correctly scopes Layers 2/3 as *successor* RFCs
  on the same evaluator (E3), matching `09`'s "one evaluator, three layers."
- **Meters foundation (row 2b)** deliberately implements only the pressure-meter *substrate* and
  scopes the *disasters* out ("the meter MECHANICS, not the events or endings that consume them",
  line 9; C6 defers the Events-L1 input arm "WHEN Events L1 lands"). So Layer 2 is genuinely
  unbuilt even though its numeric substrate is implementing — no drift, but a real dependency edge.
  It also corrected a design-fidelity bug in-flight (C1: the draft's 5-bar `public` Trust model was
  wrong vs `02 §7`'s ten Standing/Grievance axes — fixed to the binding shape).
- **Speedrun boards (row 6)** faithfully consume `05 §6`/`08 §6`: 4 canonical categories with
  code terminal predicates, player-authored predicate surface (threshold-promoted), epoch pinning,
  replay verification, world-first broadcast, `Assisted`/`Glitched` as structural variables. The
  RFC even absorbed the research's correction that a developer "cannot ship an unbeaten category"
  (`speedrun-governance.md §3.3`): Phase-0 Ethical% is TRUE-by-default until content emits a
  `darkpattern.*` fact (L7a note 1) — the `08 §6` "Pacifist unbeaten until world-first" copy is
  handled as a content-gated fact, not a decorated board.

### Status truth (`STATUS-OFF`)
- **`rfc/README.md` index is stale for two draft RFCs.** Line 69 lists **"Layer-1 events engine"**
  and (via the codex handoff) "T0–T1 content" under *"Remaining Phase-0 contracts (not yet
  drafted)"* — but `rfc/events-engine-layer1.md` and `rfc/t0-t1-playable-content.md` both exist as
  **draft** files dated 2026-08-03, and **neither appears in the Active index table** (which ends at
  Achievements Foundation). The index both under-claims (says undrafted) and omits them. `STATUS-OFF`
  on the index, not on the RFC files themselves (their own status lines read "draft" honestly).
- Meters (row 2b): RFC line "accepted (C1–C12 ruled; implementing)" matches README "implementing";
  C13 activation ruling recorded 2026-08-04. Consistent.
- Leaderboards/epochs, prestige: README "implementing" matches RFC lines. Consistent.
- Route Registry (row 6c): archived + `docs/routes.md` canonical — genuinely `I`. Caveat: the
  *public naming ledger / adoption curves* ("Registry Analytics") are explicitly deferred
  (`docs/routes.md` §Registry: "variants, time buckets, adoption curves, and its public read API
  belong to Registry Analytics"). The *mechanic* (first-executor grant, per-run execution, route
  discount/substitution) is implemented; the player-facing "first player names it, permanently"
  ledger is not.

### Gap identity — the GAP backlog with builds-on (dependency) notes
Each stage-D system below is a required backlog entry. Foundations named are the ones each must
build on before it can be drafted/implemented.

1. **Events Layer 2 (pressure-meter disasters)** — builds on **Meters Foundation** (the substrate;
   the disaster is a hidden recurring event that reads meter bands) **+ Events Layer 1 evaluator**
   (`09 §3` disasters are events; `events-engine-layer1.md` E3 says "a meter is a hidden recurring
   event"). Cannot start until L1 lands.
2. **Events Layer 3 (Situations + Major Orders)** — builds on **Events Layer 1 evaluator +
   Commons/World Layer + Feed/Dispatch + Leaderboards/Epochs** (server-authoritative shared state,
   Galactic-Impact-Modifier contribution accounting, published formulas). The heaviest gap.
3. **Conspiracy system** — builds on **Meters Foundation (extended)** — note the Conspiracy
   Pressure meter is **NOT** in the current 11-row meter catalog (10 Trust + `doom.probability`
   only), so the registry must grow — **+ Events L1/L2** (spawn/spread events) **+ Clout**
   (monetize path converts Pressure→Clout). The Tier-7 conspiracy *inversion* additionally needs
   the tier ladder to Tier 7.
4. **Media canonization / Clout milestone events** — builds on **Clout** (milestone thresholds
   fire the arrivals) **+ Events L1 evaluator** (Ignore/Embrace/Sue is an options-with-effects
   event). BLOCKER: `docs/achievements.md` states the Achievements Foundation "does NOT mint Clout"
   yet — so the Clout *currency* these events key on is itself unbuilt. Depends transitively on a
   future Clout-minting RFC.
5. **Speedrun chrome (title bar / splits / PB / gold-splits / retrospective / TAS)** — builds on
   **Leaderboards & Epochs** (PB/WR data, splits = committed tier-transition events) **+ UI
   Foundation / game-UI screens + Prestige** (first-Exit reveals the PB/WR line per `08 §6` timer
   semantics). TAS mode additionally needs **Tier 7** content.
6. **The endings** — builds on **the tier ladder through Tier 8 (Depletion)** **+ Prestige/Exits**
   ("New Route" reset) **+ Leaderboards** (Ethical% world-first broadcast) **+ Ethical%/morality**
   ledger. Ending C requires having seen A and B (or Ethical%), so it is strictly last.
7. **Tiers 2–8 content ladder** — each tier is a **content epoch on Purchasable Content Foundation
   (implemented) + Doctrines + Meters + Minigames + Factions/Commons/World**; only **T0–T1** is
   drafted. Tier 2 introduces Soul/allocation, Tier 3 the market + easter egg, Tier 4 the region
   draft/world map, Tier 5 p(doom)/research tree, Tier 6 policies, Tier 7 Transcendence/TAS, Tier 8
   Depletion/endings. Sequential dependency; the single largest content program in the game.
8. **Cookie Clicker / Bakery Inc. easter egg** — builds on **Tier 3 content (Cloud/market/tenant
   sim) + Minigame Platform** (the clickable game-within-a-game) **+ Tier 7** (the completion beat,
   rack 42). Cross-cutting flavor, not a standalone system.
9. **Event reward primitive (scrip / exchange shop / plot prize shop)** — builds on **Events L1/L3**
   + the economy kernel (auto-convert at published rate). The *only* sanctioned event-reward valve
   (`09 §5`) — must exist before any Layer-3 arc pays out.
10. **Seasonal arcs / dispatches / evergreen conversion** — builds on **Events Layer 3 + Feed/
    Dispatch + GM ops**. The live-service cadence layer.
11. **GM operations (war log / watchdogs / dials)** — builds on **Events Layer 3**; a budgeted ops
    role + dashboard. `09 §7`'s watchdog (max-lifetime, forced-resolution, alarm) is flagged as the
    genre's #1 operational risk and should ship *with* Layer 3, not after.

### One design-law note surfaced
`09` (Black Sunday doctrine, Billing-Anomaly kit) and law 10 (curtain-pull) are honored in the
design text; no violations to flag. The conspiracy layer's dark-content calibration (`08 §1.6`,
and the amended `conspiracy-media.md §1b` Slender Man note: "model the folklore mechanism only, do
not build an event chain from the crime") is a live copy-review constraint any events-content RFC
must carry forward — worth a `[V]`/curtain gate in the eventual content RFCs.
