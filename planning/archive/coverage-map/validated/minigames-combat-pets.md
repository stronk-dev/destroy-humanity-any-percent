# Coverage Map — Validated: minigames-combat-pets

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** Reconstructed on 2026-08-05; retained as
> evidence, not current status or execution authority. See `planning/platform-alignment/`.

Reconstructed from actual files 2026-08-05. Stage legend: **R** research dossier · **D** design
section · **F** RFC (with actual status) · **I** implemented (archived + docs). Furthest = the
highest stage with real backing.

CRITICAL FRAME: the minigame **platform** having a foundation says nothing about the individual
minigames. Platform + each specific game are accounted separately. Every specific minigame's
furthest stage is **D** (design/03 only) with **no RFC** → all are **GAP**s building on the
platform foundation.

## Table

| # | System | R | D | F (status) | I | Furthest | Tags |
|---|---|---|---|---|---|---|---|
| 1 | Minigame platform (session/faucet/payout/ratings/registry) | `neopets-systems §4`, `kol-puzzle-pirates §B` (specify) | `03 §12b`, `05 §4` | `minigame-platform-foundation.md` — accepted, C1–C36 ruled, **implementing** | partial: `server/minigame/*` (session, tenant, scaling, faucet, payout, fallback, offline_quality), `docs/minigame-platform.md` | **F (impl. partial)** | STATUS-OFF (absent from README index); DRIFT |
| 2 | Combat data model (shared arithmetic/type chart/determinism) | `creature-battler §8`, `lane-pusher-design` (specify) | `03 §10`, `04 §2`, `05 §4` | `combat-data-model.md` — **implementing** | partial: `server/combat/arithmetic.go` + `client/src/combat`, `docs/combat.md`. **C2 catalog effect-union + C5 Trust/Soul tables NOT built (DESIGN-GAP)** | **F (impl. partial)** | GAP-in-RFC (catalog/tables unresolved) |
| 3 | Combat duel engine (pet battles engine) | `creature-battler §8` (specify) | `04 §2` | `combat-duel-engine.md` — **draft** | none | **F (draft)** | — |
| 4 | Combat lane engine ("Push to Prod") | `lane-pusher-design` (specify, sim-backed) | `03 §10` | `combat-lane-engine.md` — **draft** | none | **F (draft)** | — |
| 5 | Combat bots + verification + integration (AI fallback) | `adaptive-balancing §2`, `creature-battler` (specify) | `04 §2`, `05 §4` | `combat-bots-and-integration.md` — **draft** | none | **F (draft)** | — |
| 6 | Pet care (tamagotchi core: stats/decay/actions/trust/mood/FSM) | `cattery-reusables`, `neopets-systems §3`, `creature-battler §8.3` (specify) | `04 §1` | `pet-care-foundation.md` — accepted, C1–C17 ruled, **implementing** | partial: `server/pet/*` (grammar, state, catalog wire only), `docs/pet-care.md` | **F (impl. partial)** | STATUS-OFF (absent from README index); DRIFT |
| 7 | Pet battles (pokemon-style) | `creature-battler §3.4, §8` (specify) | `04 §2`, `03 §10b` | mechanics = duel engine (#3, draft); **content (rosters/seasons/movesets) has no RFC** | none | **F (draft, mechanics) / D (content)** | GAP (battle content) |
| 8 | Server Garden (rack/crossbreeding) | — (no dossier; broad mentions only) | `03 §1` | **none** | none | **D** | GAP (builds on platform #1) |
| 9 | Board-game suite (chess, checkers, connect-4, othello, gomoku, tic-tac-toe) | `adaptive-balancing §2` (AI/ratings), `06 §vs-AI` engines (mention) | `03 §5`, `05 §4` | **none** | none | **D** | GAP (builds on platform #1 + combat-bots contract) |
| 9b | Parlor taxonomy expansion (trick-taking/melding/racing/dice/trivia; duplicate scoring) | `neopets-systems §2` (mention) | `03 §5b` | **none** | none | **D** | GAP (builds on #9/#1) |
| 10 | The Market (stock, $=seconds) | `neopets-systems §2` (NEODAQ, specify shape) | `03 §4` | **none** | none | **D** | GAP (builds on platform #1) |
| 11 | Ship-It Spellbook (grimoire / push-your-luck / gamble) | `cookie-clicker` (grimoire mention) | `03 §3` | **none** | none | **D** | GAP (builds on platform #1) |
| 12 | Advisory Board (pantheon/loadout + swap budget) | `cookie-clicker` (pantheon mention) | `03 §2` | **none** | none | **D** | GAP (builds on platform #1) |
| 13 | Incident Response (pager reflex, regen pool) | — | `03 §6` | **none** | none | **D** | GAP (builds on platform #1) |
| 14 | Terminal Typer (session-skill typing) | — | `03 §7` | **none** | none | **D** | GAP (builds on platform #1) |
| 15 | Demo Disc Arcade (evolving nostalgia toys) | `flash-era-arcade` (mention) | `03 §8` | **none** (`t0-t1-playable-content.md` is generators/copy, NOT arcade) | none | **D** | GAP (builds on platform #1; lightweight-session path is an open Q) |
| 16 | Bakery, Inc. (game-within-game easter egg) | `cookie-clicker` (mention) | `03 §9`, `01 §T3` | **none** | none | **D** | GAP |
| 17 | Shipping Wars (solo era-lane, Age of War) | `tier-relevance §4` (mention) | `03 §12b` | **none** (named content successor in `combat-lane-engine.md`) | none | **D** | GAP (builds on lane engine #4 + platform #1) |
| 18 | The house (cosmetic base) | `lane-pusher-design §7` (raids-rejection) | `04 §3` | **none** | none | **D** | GAP |
| 19 | Hats / cosmetics / free lootbox machine | `gaming-enshittification` (specify) | `04 §4` | **none** | none | **D** | GAP |
| 20 | Breeding & rarity | `cattery-reusables`, `neopets-systems §3` (mention) | `04 §5` | **none** | none | **D** | GAP |
| ~~21~~ | ~~Base building & raids (Clash layout/defense)~~ | `lane-pusher-design §7` | `04 §3` **STRUCK** | — | — | **DROPPED** | Does not exist — struck 2026-07-28; replaced by Lane (#4) + house (#18). Do not reintroduce. |

## Findings

### Platform vs. individual-minigame gap (the headline)

**The platform is implemented; not one individual minigame is.** `minigame-platform-foundation.md`
(#1) has 36 rounds of rulings and real code in `server/minigame/` (session lifecycle, tenant
boundary, scaling grammar, faucet governor, payout/conversion kernel, fallback + offline_quality
loaders). But its only tenant today is a **test-only conformance fixture** — `docs/minigame-platform.md`:
"The current conformance tenant is test-only. The combat duel adapter will register when its engine
RFC supplies an implemented transition surface."

Every specific minigame (#8–#17) sits at furthest stage **D** with **no RFC**: Server Garden, the
board-game suite (chess + connect-4/othello/gomoku/tic-tac-toe/checkers), the parlor-taxonomy
families, the Market, Ship-It Spellbook, Advisory Board, Incident Response, Terminal Typer, Demo
Disc Arcade, Bakery Inc., Shipping Wars. **Individual-minigame content-contract GAP count: 10 named
minigames + parlor-taxonomy expansion (11 rows), all D→no-RFC, all building on platform #1** (and
Shipping Wars additionally on lane engine #4). `t0-t1-playable-content.md` is NOT a minigame RFC —
it mints T0–T1 generators/upgrades/gates/copy on existing schemas and only writes arcade *copy*, no
arcade session/score contract.

### STATUS-OFF / DRIFT — two implementing foundations are missing from the RFC index

`rfc/README.md`'s Active table does **not** list `minigame-platform-foundation.md` or
`pet-care-foundation.md`, even though both carry status "accepted … implementing", both have
extensive planning logs with owner rulings, both have shipped code (`server/minigame/*`,
`server/pet/*`) and canonical docs. The only trace is the Meters row's "unblocks … Pet Care" note.
Combat Shared Data and the three combat children ARE indexed. **Evidence:** README active rows end
at Achievements (line 25); `grep -i 'minigame\|pet.care' rfc/README.md` returns only the Meters
dependency mention. This is index drift — a fresh agent reading only the index would not know these
two foundations exist or are in flight.

### GAP-in-RFC — Combat Data Model is "implementing" but structurally incomplete

`combat-data-model.md` is indexed "implementing" and has real arithmetic/determinism code, but
`docs/combat.md` records: "The strict combat catalog and Trust/Soul input tables remain
unimplemented because their active RFC does not yet enumerate the promised closed effect union or
literal piecewise table points. Those are recorded as DESIGN-GAPs." So C2 (catalog `moves.effect`
closed union) and C5 (`(trust_ppm, soul)` piecewise tables) are unresolved inside an "implementing"
RFC. The three combat engine children (#3–#5, all draft) normatively depend on C1–C5 and cannot
proceed to golden vectors until C2/C5 land. Pet Care C5 is the intended real producer of the C5
`(trust_ppm, soul)` inputs — so the two partial foundations are mutually blocking on that seam.

### Combat engines / bots — coherent draft cluster, no code, correctly indexed

`combat-duel-engine.md`, `combat-lane-engine.md`, `combat-bots-and-integration.md` are all **draft**
(matching README), split from the parent per the 2026-07-29 Codex four-way review. They are the
mechanics for **pet battles** (#7) and the **Lane** (#4/#17) respectively. No planning dirs yet
(expected for drafts). No ORPHAN — each cites its parent and design ref. The duel engine is
explicitly the shared engine for pet battles AND board-game matches (`04 §2`, `03 §10b`), so #7's
mechanics are covered by #3; only pet-battle **content** (rosters/seasons) is a GAP, which
`pet-care-foundation.md` open questions and this map both flag.

### Pet layer — care foundation partial; battles/house/cosmetics/breeding all D

`pet-care-foundation.md` (#6) is the tamagotchi core (4 stats/decay/actions/trust/mood/FSM/bonds),
introduces the reusable `ApplyFounderLogged` boundary, C1–C17 ruled. Code so far is wire grammar
only (`server/pet/grammar.go`, `state.go`, `catalog.go`; `docs/pet-care.md`: "currently owns only
the cross-runtime wire grammar"). Founder save **v18** and the `pets` epoch artifact are specced
(Pet C16/C17 ↔ Minigame C35/C36: `minigames`=Founder v17, `pets`=Founder v18, scalar Founder
version axis independent of Company v14/v16) but not production-activated. The rest of `design/04`
(#7 battles content, #18 house, #19 cosmetics/lootbox, #20 breeding) has no RFC → all GAP.

### DROPPED — base building & raids does not exist

The domain brief lists "base building & raids (Clash-of-Clans layout)". Per `design/04 §3` this was
**struck 2026-07-28** (owner rejection, `lane-pusher-design §7`: layout is a solved puzzle, 82% of
raid outcomes from one copied optimum). Competitive spatial play moved to the Lane (#4) and the
region draft (`design/13 §5`); the pet house (#18) is the surviving cosmetic-only remnant. Recorded
as DROPPED, not a gap — and design law "no reintroduce" should be honored.

### No UNBACKED or true-ORPHAN systems found

Every RFC in the domain cites its design/research parents; every implementing foundation has a
docs page. The only drift is the README index omission (above), not code-without-spec.
