# Coverage validation slice: economy-progression

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** Reconstructed on 2026-08-05; retained as
> evidence, not current status or execution authority. See `planning/platform-alignment/`.

> Independently reconstructed from the actual files 2026-08-05. Each row's four-stage state (R
> Researched / D Designed / F Foundation / I Implemented) is rebuilt from file evidence, not from
> prior belief. Furthest stage recorded per system; adversarial flags below the table.

Legend: **I** = archived RFC + canonical `docs/` page · **F** = has an RFC contract
(draft/implementing/accepted, not archived) · **D** = designed, no RFC · **R** = research only.

## Systems table

| # | System | R (research backs mechanic?) | D (design intent) | F (RFC + actual status) | Furthest | Flags |
|---|---|---|---|---|---|---|
| 1 | **Numeric core** | ✅ `research/numeric-core.md` §§154–268 SPECIFIES the contract (12-digit quantization, wire format, closed-form bulk buy, test contract) | `02 §12`, `06 §big-numbers`, law 3 | RFC-0001 + hardening chain, all **implemented/archived** [a] | **I** | — |
| 2 | **Economy kernel** (cost curves / ceilings) | ✅ `research/economy-kernel.md` §§71–92 (data/runtime split, query/event boundaries) + `cookie-clicker.md` (r-curve, milestone mults) | `02 §2.1–2.2`, `§2c` (Neopets laws) | RFC-0002 + geometric-afford-fast-path + generator-production-state + resource-log-domain-parity + production-hardcap-saturation — **implemented/archived** [b] | **I** | — |
| 3 | **Production / idle math + offline** | ✅ `cookie-clicker.md` (offline, golden cookies, Lucky), `pacing-science.md`, `economy-kernel.md` | `02 §2.2` stack, `§2.4` offline (90%/24h), `06 §idle-math` | production-engine-and-intents + production-accrual-math + production-contract-integrity + millisecond-cursor-canonicalization — **implemented/archived** [c] | **I** | **partial** — the *active-play* sub-layer (§2.3) is unbuilt; see row 16 |
| 4 | **Save layer + migrations** | ✅ `tech-stack.md §1.7` (settled practice, research README L52) | `06-tech` (versioned schema, migration chain, NaN refusal — law 8) | save-layer-and-migrations + save-archive-cas — **implemented/archived** [d] | **I** | — |
| 5 | **Prestige / exits** (Founder layer, run reset) | ✅ `cookie-clicker.md` (ascension/heavenly-chips, cube root), `pacing-science.md` (first-Exit pacing), `run-narrative-ux.md`, `speedrun-governance.md` (New Route) | `02 §3` (cube-root formula, Founder sub-currencies, Exit fiction), `01` "Prestige inside the arc", `10 §5` | `rfc/prestige-and-exits.md` — **implementing** (not archived); core landed + reviewed, plan items 6 & 9 open | **F** | open **DESIGN-GAP** (Advisor Mode toggle intent, log L80–82); docs page present-tense while RFC still active |
| 6 | **Factions** (archetypes) | ✅ `events-playstyles.md` topic 2 / §2c (rule-variance grammar) | `10 §1` (4 factions + later factions), `02 §2.2` FactionModifiers | `rfc/archive/faction-incorporation.md` — **implemented/archived**, clean review | **I** | per-faction rule *packages* deliberately deferred as content (declared, not drift) |
| 7 | **Meters** (Trust / Externality / pressure) | ✅ `morality-systems.md` (New Vegas 5×2 architecture, meter legibility/persistence) | `02 §7` (Trust 5 constituencies × 2 bars; Externality→ledger; p(doom) removed from moral axis), `09 §3` (pressure meters) | `rfc/meters-foundation.md` — **accepted (C1–C12 ruled; implementing)**; ships 10 Trust axes + `doom.probability`; Externality = input-ledger only; Soul read-only. docs/meters.md present but "no production meter artifact minted yet" | **F** | **DRIFT** (owner-ruled): §7 calls Soul a meter + mandates Trust↔Soul CI gate; RFC defers Soul-as-meter (C3) & leaves gate unbuilt (C10). Body-reconciliation residual: M5 says v15, ruling/impl is v16 |
| 8 | **Clout + achievements** | ✅ `gaia-hyperinflation.md §4a` (one-mint law), `neopets-systems.md` (possession→credit markets), `cookie-clicker.md` (milk model) | `02 §6` (achievements +4 Clout, PR-Intern mult), `§6b` Gaia one-mint law, `§2c` law 3 | `rfc/clout-and-achievements-foundation.md` — **accepted (C1–C10 ruled; SCOPE NARROWED to Achievements; implementing)**; ships achievement IDs + integer score only, **Clout removed** (deferred to Feed/Social); neutral in production | **F** (achievements) / **D** (Clout) | **DRIFT** (owner-ruled): §6 L148 "every achievement grants +4 Clout" reversed by C1; §6 CloutStack removed from production by C2 — both follow §6b over §6's contradictory example. Clout *mint* now an unowned GAP (Feed/Social, mmo-social domain) |
| 9 | **Leaderboards / epochs** | ✅ `adaptive-balancing.md` (Balance Epochs, Board Mandates), `speedrun-governance.md` (categories, Route Registry, epoch_id board key), `pacing-science.md §5` | `05 §6`, `02 §11`, `10 §2` (per-build boards) | `rfc/leaderboards-and-epochs.md` — **implementing** (not archived) + substrate `rfc/archive/run-genesis-and-replay.md` **implemented/archived** | **F** (with archived substrate) | open **DESIGN-GAPs**: verified-run supersession + rejected-log cap unimplemented (log L277–301); plan items 3/6/7 open; L2b drift-handling ruling adds mechanism |
| 10 | **Compute credit** (banked offline time) | ✅ `events-playstyles.md §Time banking` (specifies Compute Credits + legibility law), `morality-systems.md §3.2e` | `02 §9` (accrual→banked cap, spend as acceleration burst), `10` Banker build | `rfc/doctrine-and-compute-credit.md` DC2 — **draft**; no planning dir, no docs, never started | **F** (draft) | **STATUS-OFF**: README L69/L72 lists "Compute Credit spend" as *not yet drafted* and the RFC is absent from the active table — but the draft exists (2026-08-03) |
| 11 | **Doctrine** (lobbying / capture intents) | ⚠️ thin — age-up itself is an AoM/EU4 borrow; `events-playstyles.md §2c` backs the rule-variance grammar; lobbying/capture = route content (`speedrun-governance.md §3.4`, `societal-satire.md §247`) | `10 §3b` (Doctrines, Age-Up 1-of-3, doctrine trees), `01` per-tier, `02 §6` (Clout-gated lobbying) | `rfc/doctrine-and-compute-credit.md` DC1 — **draft** (doctrine-pick intent only; doctrine-TREE content deferred). No planning dir, no docs | **F** (draft) | **STATUS-OFF** (same README staleness as row 10). The task-named **lobbying/capture "intents" are NOT specified** by any RFC — they are event/route/world content → **GAP** (see row 17) |
| 12 | **Soul** (personal ledger) | ✅ `morality-systems.md` (owns Soul axis: legibility, meters, persistence — research README L77) | `02 §8` (drains/gates/recovery/endgame), `10 §5`, `01` (longevity rungs cost Soul) | **No owning RFC.** `FounderState.Soul` exists as a read-only int64 carry (referenced by meters/prestige/pet-care/combat), but Soul-changing verbs are explicitly unowned (meters C3: "needs a successor owning a two-stream write") | **D** (mechanic) | **GAP** — drain/recover/gating mechanic + Trust↔Soul CI gate unbuilt; builds on Meters (punts it) + Save/Founder state |
| 13 | **Quarters / earnings calls** (real-time ripening) | ✅ `cookie-clicker.md §5.5` (sugar lumps — CpS-immune real-time meta-currency, the exact mechanic Quarters mirror), `adaptive-balancing.md §5` (Boss carve-out) | `02 §5` (Earnings Call ripens 24h, Investor Confidence, Golden Quarter, Sugar-Baking +1%/unspent cap 100), `02 §1` | **No RFC** (grep of `rfc/` = zero hits for earnings/quarter/investor-confidence) | **D** | **GAP** — builds on Production Engine (CpS-immune clock) + Save; not even named in README's undrafted list |
| 14 | **Ideology toggles** (enshittification mode) | ✅ `idle-landscape.md §431` item 6 (Grandmapocalypse "make it worse for a multiplier", "mechanically proven"), `morality-systems.md` (Grandmapocalypse pattern), `gaming-enshittification.md` | `10 §3` (Enshittification slider, Go Public, Autopilot Mode), `02 §2.3` (daemons gated behind slider), `01 T3` | **No RFC** (grep hit only copy-pipeline archive = copy text) | **D** | **GAP** — builds on Meters (Trust drain) + Production (daemons/revenue valve) + Prestige |
| 15 | **Relevance / tier-falloff enforcement** | ✅ `tier-relevance.md` + `balance-enforcement.md` SPECIFY (five families, restricted-play, Demaine bound, Riot ANY/ALL) | `02 §11b` (tier-relevance doctrine + enforcement report) | `rfc/relevance-harness.md` — **draft**; planning log's only verdict (Codex 2026-08-03): "not implementable yet; implementation did not start." | **F** (draft, blocked) | **STATUS-OFF** (absent from README active table though `02 §11b` names it as spec). Blocker nuance: log demanded a *Purchasable Content Foundation* first — that RFC is now **archived/implemented**, so the stated blocker is stale but the RFC is still unstarted |
| 16 | **Active-play buff windows** (golden opportunities / daemons / Lucky bank) — *discovered* | ✅ `cookie-clicker.md` (golden cookies distribution, Lucky formula verbatim, wrinklers) | `02 §2.3` (golden opportunities t⁵·exp, Lucky `min(0.15·bank,900·rate)`, daemons/wrinklers, macro bench) | **No RFC** — production foundation shipped accrual/offline only | **D** | **GAP** — builds on Production Engine + Meters (outrage events); the idle build's crown jewel (daemons) is gated behind the (also-unbuilt) enshittification slider, row 14 |
| 17 | **Lobbying / capture mechanic** (as distinct from doctrine-pick) — *discovered* | ⚠️ `speedrun-governance.md §3.4/§7` (Route Registry skips incl. Regulatory Capture Skip), `societal-satire.md §247`, `billionaires-decay.md` (Super PAC, Friedman doctrine) | `02 §6` (Clout gates regulatory outcomes), `09 §doctrine events` (Regulatory Heat → Clout-gated lobbying), `08 §6` skips | **No RFC** — doctrine RFC (row 11) explicitly excludes tree/lobbying content | **D** | **GAP** — builds on Clout mint (unbuilt) + Events engine (unbuilt, `09`) + Routes (implemented) |

---

## Findings

### STATUS-OFF

- **S1 — README index is stale on two drafted RFCs.** `rfc/README.md` L69/L71–72 lists "doctrine
  intents" and "Compute Credit spend" under *"Remaining Phase-0 contracts (not yet drafted)"*, and
  neither `rfc/doctrine-and-compute-credit.md` (draft, created 2026-08-03) nor
  `rfc/relevance-harness.md` (draft, 2026-08-01) appears in the README **Active** table (L10–26).
  All three are drafted; the index does not track them. `doctrine-and-compute-credit.md` even
  claims to resolve the precondition the README still lists as open ("must define doctrine-pick
  ordering"). Affects rows 10, 11, 15.

- **S2 — no *contradiction* between any RFC status line and the README status column** for the
  four active/implementing/implemented economy RFCs (prestige, leaderboards, meters, achievements,
  factions all consistent). The only status truth defect is S1 (the undrafted-list vs the existing
  drafts).

### DRIFT (all owner-ruled and recorded, none silent)

- **D1 — Soul demoted from meter to read-only carry** (row 7). `02 §7` L170 specifies "Trust and
  Soul **as meters**"; `meters-foundation.md` C3 removes Soul drain/recover, keeps
  `FounderState.Soul` as a read-only int64, and states Soul-changing verbs need an unwritten
  successor. Consequence: the design's Soul *system* has no owner (row 12 GAP).

- **D2 — Trust↔Soul CI correlation gate not built** (row 7). `02 §7` L170 mandates a CI gate
  keeping Trust and Soul decorrelated; `meters-foundation.md` C10 registers it as
  "content-blocked Relevance work, not falsely claimed green." Correctly honest, but the
  design-law gate does not exist yet.

- **D3 — Achievements mint no Clout** (row 8). `02 §6` L148 says "every achievement grants +4
  Clout" and makes CloutStack/PR-Interns "the single biggest multiplier family";
  `clout-and-achievements-foundation.md` C1 removes Clout entirely (achievements → score only) and
  C2 removes lifetime Clout from production, both following the `§6b` Gaia one-mint law. This is a
  resolution of an **internal design contradiction** (§6 example vs §6b law), not a divergence
  from settled intent — but `02 §6`'s body still reads as if achievements mint Clout and should be
  reconciled to §6b.

- **D4 — meters RFC body/ruling residual** (row 7). `meters-foundation.md` M5 still says "Save
  **v15**"; the accepted C13 ruling and shipped code are **v16 atomic** with Achievements
  (docs/meters.md L32 is correct). A body-not-reconciled-to-ruling residual under the project's
  "rulings reconcile the body" convention.

### GAP (furthest stage D — designed, no RFC; ordered by substrate readiness)

| System | Design | Builds on | Draftable now? |
|---|---|---|---|
| **Quarters / earnings calls** (row 13) | `02 §5` | Production Engine ✅ + Save ✅ (both implemented) | **Yes** — CpS-immune clock + Investor Confidence spend; substrate exists |
| **Ideology toggles / enshittification** (row 14) | `10 §3`, `02 §2.3`, `07 §sliders` | Production ✅ + Meters (implementing) + Prestige (implementing) | **Partly** — needs Meters Trust artifact minted; the daemon/revenue valve half is draftable |
| **Soul mechanic** (row 12) | `02 §8`, `10 §5` | Meters (implementing, punts Soul) + Save/Founder ✅ | **Partly** — needs the two-stream Soul-write successor meters C3 names |
| **Active-play buff windows** (row 16) | `02 §2.3` | Production ✅ + Meters (outrage events) | **Yes** — golden-opportunity spawn + Lucky bank formula are self-contained on the shipped engine |
| **Lobbying / capture mechanic** (row 17) | `02 §6`, `09 §doctrine events` | Clout mint (unbuilt) + Events engine (unbuilt `09`) + Routes ✅ | **No** — blocked on the unwritten Clout mint and Layer-1 events engine |
| **Clout mint** (row 8, split off) | `02 §6`, `§6b` | Feed/Social foundation (unwritten, mmo-social domain) | **No** — deferred to Feed/Social by achievements C1 |

### Draft RFCs that are effectively blocked (F but not moving)

- **Relevance Harness** (row 15) — draft, Codex verdict "not implementable yet." Its cited blocker
  (accept a Purchasable Content Foundation first) is now stale: `purchasable-content-foundation.md`
  is archived/implemented. Worth re-triaging whether the RFC is now unblocked.
- **Doctrine Intents & Compute Credit** (rows 10, 11) — draft, no planning dir, never started;
  substrate (Routes, Production, Run Genesis) all implemented, so it is draftable-to-implementing
  now.

### Note on furthest-stage classification

- **Prestige, Leaderboards, Meters, Achievements** are all "implementing" with **canonical
  `docs/` pages already written in present tense** while their RFCs remain un-archived and carry
  open plan items / unrecorded final independent-review-before-archival gates. They are counted as
  **F** (foundation), not **I**, because archival + the review gate have not landed — but they are
  materially closer to I than a bare draft.
- **Soul** is counted **D** because the *system* (§8 drains/gates/recovery) has no owner, even
  though a read-only `FounderState.Soul` value is persisted by the implemented save/founder layer.

---

## Footnote citations (RFC → docs, from `rfc/README.md` archive table)

- [a] Numeric core: `archive/0001-numeric-core.md`, `numeric-core-boundary-hardening.md`,
  `numeric-normalization-carry.md`, `numeric-boundary-parity.md`,
  `deterministic-decimal-aggregation.md`, `production-accrual-math.md` → `docs/numeric-core.md`.
- [b] Economy kernel: `archive/0002-economy-constants-and-ceilings.md`,
  `geometric-afford-fast-path.md`, `generator-production-state.md`,
  `resource-log-domain-parity.md`, `production-hardcap-saturation.md` → `docs/economy-kernel.md`.
- [c] Production: `archive/production-engine-and-intents.md`, `production-accrual-math.md`,
  `production-contract-integrity.md`, `millisecond-cursor-canonicalization.md`,
  `production-hardcap-saturation.md` → `docs/production-engine.md`.
- [d] Save: `archive/save-layer-and-migrations.md`, `save-archive-cas.md`,
  `millisecond-cursor-canonicalization.md` → `docs/save-layer.md`.
