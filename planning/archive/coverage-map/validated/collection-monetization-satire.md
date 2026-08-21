# Coverage slice: collection-monetization-satire

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** Reconstructed on 2026-08-05; retained as
> evidence, not current status or execution authority. See `planning/platform-alignment/`.

Independently reconstructed from actual files on 2026-08-05. Every row cites file evidence per
stage (R = research dossier, D = design section, F = RFC contract, I = archived + docs). Stages
are claimed only where a file actually backs the mechanic.

## The load-bearing distinction (substrate vs. content)

The **purchasable-content FOUNDATION is implemented** — but it implements the *tier-relevance
doctrine* (`design/02 §11b`: upgrades, generator chains, synergy pools, milestone ladders, typed
generator roles, the ablation seam), NOT any monetization satire. It is a generic catalog/purchase
substrate. **Every piece of the actual collection-monetization satire content** (loot boxes, the
hat/cosmetic collection, breeding, the dark-pattern parody suite) **rides on that substrate and is
still at furthest-stage D — designed, fully researched, but with no content RFC.** The one
exception is `Horse Armor (Free)`, which appears as an explicit *stub shelf item* inside two draft
RFCs. The substrate is done; the satire that is the entire point of the domain is not started.

Note on dossiers: the primary backing dossier is **`gaming-enshittification.md`** (it *specifies*
the entire parody suite in §6, at ship-ready copy density). `billionaires-decay.md` was named as
the late-tier dossier, but it backs late-tier *era flavor and building/ticker content*, NOT the
monetization/cosmetic mechanics — those remain sourced from gaming-enshittification across all
tiers.

## Table

| # | System | R | D | F | I | Furthest |
|---|---|---|---|---|---|---|
| 1 | Purchasable-content foundation (substrate) | `tier-relevance.md`, `cookie-clicker.md` | `design/02 §11b`, §2 | `rfc/archive/purchasable-content-foundation.md` **implemented** | `docs/purchasable-content.md` | **I** |
| 2 | Loot boxes / free gacha (Surprise Mechanic Crates, pity=1, 50/50, kompu-gacha) | `gaming-enshittification.md §6.1` (specifies) | `design/04 §4`, `design/08 §3` | — none — | — | **D** `GAP` |
| 3 | Hats / cosmetics / collection book (Unusual, the Knife #387, Default Skin) | `gaming-enshittification.md §3, §6.3` (specifies) | `design/04 §4` | — none (game-ui-screens U3 explicitly excludes it) — | — | **D** `GAP` |
| 4 | Horse Armor (+ horse-armor-for-pets, Remastered) | `gaming-enshittification.md §6.3` (specifies) | `design/04 §4`, `design/01 §Tier 1` | `rfc/t0-t1-playable-content.md` (draft, **stub shelf only**), `rfc/game-ui-screens.md` (draft, static shelf) | — | **F (draft, stub)** |
| 5 | Breeding & rarity | `design/04 §5` refs `cattery-reusables.md`, `neopets-systems.md §3` | `design/04 §5` | — none (pet-care-foundation explicitly defers) — | — | **D** `GAP` |
| 6 | Dark-pattern parody suite (battle-pass, "Just Give Me The Thing", energy bar, shutdown/LAN prestige, layoffs, sub-stacking, NFT one-shot, kernel-anticheat, ads-that-pay, always-online toggle, time-gate) | `gaming-enshittification.md §6.4` (specifies) | `design/08 §3`, `design/01 §4.1 era map` | — none dedicated (layoffs/shutdown are event *content* on the draft events engine) — | — | **D** `GAP` |
| 7 | Era-mapped monetization-satire content (tier↔era beats) | `gaming-enshittification.md §4.1` (specifies) | `design/01` (per-tier era beats), `design/08 §2` | `rfc/t0-t1-playable-content.md` (draft — T0–T1 beats only) | — | **F (draft, T0–T1 only)** |

Related substrate outside this domain (cited, not owned here): the **exchange shop** — the
"only event-reward primitive," the anti-lootbox valve where event crates become posted-price shop
items (`liveservice-idle-tier.md §3` → `design/09 §5` → `rfc/events-engine-layer1.md` draft E1).
That belongs to the events-narrative slice; it is the surface event-crate content will ride on.

## Findings (evidence)

### 1. Purchasable-content foundation — genuinely IMPLEMENTED, but zero satire content on it yet
- Status is consistent across all three sources: RFC line 2 `Status: implemented`; `rfc/README.md`
  archive row (with three canonical docs); `docs/purchasable-content.md` exists. No `STATUS-OFF`.
- **Critical nuance, honest not a defect:** `docs/purchasable-content.md:6-9` — "The active
  Phase-0 catalog remains schema version 3 until the first T0–T1 content epoch is explicitly
  minted, so this document describes executable *capability* rather than claiming unshipped balance
  content." The engine can hold upgrades/chains/pools/ladders; **no upgrade, cosmetic, or crate
  content is loaded.** There are 3 placeholder generators (`rfc/t0-t1-playable-content.md:14`).
- **No `DRIFT`:** the RFC implements `design/02 §11b` (tier-relevance doctrine, `design/02:219`) —
  role law, purchased/generated split, synergy pools, staggered ladders — faithfully. But §11b is
  a *production-relevance* doctrine, entirely orthogonal to the monetization satire in `design/04
  §4` / `design/08 §3`. This is why the substrate can be "done" while the domain's actual content
  is untouched.

### 2. Loot boxes / free gacha — furthest-stage D, `GAP`
- R fully specifies it: `gaming-enshittification.md §6.1` gives ship-ready copy for Surprise
  Mechanic Crate, the hostile 100%-odds disclosure table ("Odds sum to 4,700.00%"), pity counter
  pinned at 1, the double-won 50/50, "Kompu Gacha (Legal Edition)," the apologizing near-miss.
- D carries it: `design/04 §4` ("The full Almanac §6 parody suite, implemented" — a design
  aspiration, not code) and `design/08 §3` (gaming-history parody suite index).
- **No RFC.** `rfc/t0-t1-playable-content.md` ships only a cosmetic shop with Horse Armor; event
  crates are explicitly routed through the exchange shop (`design/04 §4` line 58: "event rewards go
  through the exchange shop"), whose evaluator (`events-engine-layer1.md`) is itself draft. `GAP`
  on: purchasable-content foundation (substrate, done) + events engine + exchange shop + game-UI.

### 3. Hats / cosmetics / collection book — furthest-stage D, `GAP`
- R specifies the double-edge and the specific grails: `gaming-enshittification.md §3` (TF2 hats,
  CS:GO #387 knife, Fortnite OG skins, gacha banner art) and §6.3 (Unusual particle effects, the
  Knife #387, Default Skin (Legendary), AI-slop line). Confidence note flags #387 seven-figure
  valuations as folklore (§ confidence notes) — a verify-before-ship item.
- D: `design/04 §4` (Hats, Horse Armor, Unusual effects, the Knife, Default Skin, Surprise crates,
  collection book feeding Clout).
- **No RFC, and an explicit exclusion:** `rfc/game-ui-screens.md:53-56` U3 "What Phase A does NOT
  ship" lists "cosmetic shop beyond the static Horse Armor shelf … Each is a later [RFC]." No
  cosmetic-slot/equip system, no collection book exists. `GAP` on purchasable-content + game-UI.
- **Cross-dependency:** the `War-Themed Hat Simulator` achievement (50 hats, `design/04 §4`;
  `gaming-enshittification.md §6.2`) and collection-book completion feeding Clout both depend on
  this content; the Achievements foundation (`rfc/clout-and-achievements-foundation.md`,
  implementing) is generic and does not itself define these — they arrive with the cosmetics RFC.

### 4. Horse Armor — the ONE piece with a (draft) F, and only as a stub
- `rfc/t0-t1-playable-content.md:38-39` ships `Horse Armor (Free)` as the cosmetic shop's first
  item; open question :80-81 recommends it "ships as a stub shelf item … the cosmetics system is
  its own later RFC; the joke can't wait." `rfc/game-ui-screens.md:56` renders a "static Horse
  Armor shelf." So Horse Armor has draft-RFC presence, but as a hardcoded shelf entry, **not** the
  general cosmetic/equip system. Horse-armor-for-pets and the Remastered variant (`design/04 §4`)
  remain D. Furthest = F (draft, stub); the real cosmetic system is `GAP` (see #3).

### 5. Breeding & rarity — furthest-stage D, `GAP`
- D: `design/04 §5` (light breeding: bonded pets → egg with mixed personality weights + fur-palette
  blend; rare palettes with disclosed drop rates; cosmetic/personality only, no stat inheritance).
  R backing: `cattery-reusables.md` (10-swatch palette) + `neopets-systems.md §3`.
- **Explicit deferral, not silence:** `rfc/pet-care-foundation.md:82` — "Pet acquisition (how you
  get your first pet, breeding/rarity) — content, later RFC; this [RFC ships mechanics only]"; :9
  owner ruling "breadth-first — the care/trust/mood/FSM MECHANICS, not pets' content (species,
  cosmetics, the battle content)." `GAP` on: pet-care-foundation (implementing).

### 6. Dark-pattern parody suite — furthest-stage D, `GAP` (fragmented across future foundations)
- R specifies every element at copy density: `gaming-enshittification.md §6.4` (four-currency
  stack + "Just Give Me The Thing," Battle Pass Free/Free, self-refunding energy bar, 40px free
  Skip on a 24h time-gate, kernel-anticheat that compliments your tabs, always-online toggle that
  makes the game worse, the Shutdown Notice → offline/LAN prestige, the Layoffs event, ads-that-
  pay-you, subscription stacking with the crashing cancel page, the NFT Digits one-shot cutscene).
- D: `design/08 §3` (index) + `design/01 §4.1` era mapping places each at its tier.
- **No dedicated RFC.** The suite fragments across not-yet-drafted owners:
  - Layoffs event, Shutdown Notice → these are **Layer-1 event content** on `events-engine-layer1.md`
    (draft; E1 catalog is the evaluator, content ships in event packs — `rfc/events-engine-layer1.md:7`
    "first event pack ships there [T0–T1]").
  - Shutdown → LAN Party / offline-mode **prestige** touches `prestige-and-exits.md` (implementing),
    which has generic `wind_down`/collapse mechanics but no LAN/Stop-Killing-Games content.
  - Battle-pass, currency conversion, energy bar, subscription stacking, NFT one-shot, kernel
    anticheat → UI/purchasable content, no owner drafted.
  - `GAP`, tagged on: purchasable-content foundation + events engine + prestige-and-exits + game-UI.

### 7. Era-mapped monetization content — partial F (T0–T1 draft only)
- `gaming-enshittification.md §4.1` maps the enshittification stages to a tier ladder; `design/01`
  writes per-tier era beats (Tier 1 "cosmetic shop appears, first item Horse Armor (Free)"; Tier 2
  "crates + free auto-applied keys"; Tier 3 "loot-box panic, regulator NPCs"; Tier 4 GaaS/kernel
  anticheat). `design/08 §2` is the era style guide (UI decays with you, tiers 0–8).
- `rfc/t0-t1-playable-content.md` (draft) covers the T0–T1 era beats as catalog content only.
  Tiers 2–8 monetization-satire content = D. Furthest = F (draft) for the earliest slice, D beyond.

## Adversarial-check summary

- `UNBACKED`: none. Every design assertion in `design/04 §4` and `design/08 §3` traces to
  `gaming-enshittification.md §6` (with that dossier's own confidence notes flagging #387 pricing,
  Fortnite battle-pass date, Battlefront downvote count as verify-before-ship).
- `ORPHAN`: the gaming-enshittification §6 parody suite is a large, ship-ready content bank fully
  carried into design (D) but **stranded at the F boundary** — no RFC consumes it. Not an R→D
  orphan (design cites it thoroughly); it is an unclaimed content backlog at the F stage. Corollary
  weak-link: `billionaires-decay.md` was nominated as the late-tier monetization dossier but does
  not actually carry monetization/cosmetic mechanics — a naming mismatch, not a repo defect.
- `DRIFT`: none. Purchasable-content foundation faithfully implements `design/02 §11b`; it simply
  is not the satire layer, so there is nothing for it to drift from in this domain.
- `STATUS-OFF`: none. #1 implemented (verified 3 ways); #4 and #7 draft; all `GAP` rows have no RFC
  by design, matching `rfc/README.md`'s remaining-contracts list.
- `GAP` (furthest=D, no RFC): loot boxes/gacha (#2), hats/cosmetics/collection (#3), breeding &
  rarity (#5), dark-pattern parody suite (#6) — all building on the implemented purchasable-content
  foundation (plus events-engine, prestige-and-exits, pet-care-foundation, game-UI per row).
