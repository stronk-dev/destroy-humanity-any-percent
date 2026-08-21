# Absorption-Arena Design — "The M&A Arena"

> **Research proposals, not adopted product design.** This dated 2026-08-06 comparison does not
> place an arena in a release, approve a server-cost class, select an economy hook or multiplier,
> reserve names, or adopt any player-facing copy below. Every recommendation and §6/§6e value is a
> prototype hypothesis requiring current owner/design/RFC authority. `[M]` netcode, bot and cost
> claims require measurement; the legal section is issue-spotting only and is not freedom-to-
> operate advice. Current release and work status live in the platform-alignment queues.
>
> **Feeds:** `design/03-minigames.md` (historical roster context) · `design/BACKLOG.md`
> (research-routing context) · `design/13-world.md` (diorama silhouette-scale) · any future
> accepted M&A-Arena RFC. This dossier closes a research-coverage gap only; it does not unblock or
> schedule implementation.
>
> **Builds on, does not redo:** `research/lane-pusher-design.md` — **read its §4 (async + the AI fallback), §5 (minimum viable spec) and §7c (the punch-down multiplier) before this file.** That dossier established this project's real-time-PvP answer (live player, async snapshot opponent, seeded/replayable engine, behaviour-flag bots, the 200 %/5 % anti-bully multiplier "looking for a home"). This file asks whether an *eat-to-grow shared arena* — the genre's hardest netcode case — can be brought inside that same envelope, and finds that it can. · `research/creature-battler.md §7.4` (the anti-farm stack for snapshot PvP, reused wholesale) · `research/flash-era-arcade.md §4` (Bubble Tanks TD's "towers merge/grow by absorption" already flagged this thesis).
>
> **Provenance glyphs** (README house conventions): `[V]` verified against a fetched URL · `[P]` plausible/secondary · `[M]` model knowledge, unverified · `[sim]` my own reasoning/simulation, this dossier. **Legal-risk glyphs 🟢🟡🔴 appear only in the legal matrix and mean legal risk only.**
>
> **Research-budget note:** WebSearch was exhausted (200/200) before this pass; four Wikipedia pages were fetched directly (agar.io, Osmos, Katamari Damacy, slither.io — all `[V]`). **Netcode internals of the .io genre are not documented on those pages and are marked `[M]` throughout** — they are well-known engineering facts, but none is quoted against a fetched source. Anything `[M]` here that reaches shipped copy or an RFC acceptance criterion must be re-verified.

---

## 0. The one-paragraph answer

The comparison suggests a tractable prototype shape: one human, server-authoritative movement,
bots and read-only company snapshots. It may avoid the per-client broadcast cost of a many-human
arena, but equivalence to another minigame's runtime cost is unmeasured and must not be assumed.
Absorption offers a size-and-position axis distinct from stat combat. Virus-like hazards, decay and
relative-size rewards are candidate anti-snowball mechanisms; their formulas, theme and release
placement remain undecided.

---

## 1. agar.io — the canonical eat-to-grow arena, deconstructed

### 1.1 The verified mechanics `[V]`

Paraphrased from [Agar.io](https://en.wikipedia.org/wiki/Agar.io), fetched 2026-08-06:

| Mechanic | Rule | Provenance |
|---|---|---|
| **Mass → speed** | Larger cells move more slowly. | `[V]` |
| **Mass decay** | Cells gradually lose some mass. | `[V]` |
| **Splitting** | Split into two equal halves (spacebar); the new cell shoots toward the cursor — a ranged attack and a burst of reach; **max 16 cells** at once. | `[V]` |
| **Feeding / ejecting** | Release a small fraction of mass, including to cooperate with another cell. | `[V]` |
| **Green virus** | Splits a cell that eats it into many small pieces; players feed a virus to push it toward a target; **+100 mass** to consume one; small cells can hide under it. | `[V]` |
| **Red virus** | Cells over **200 mass** can swallow it for **+200 mass**; it consumes cells under 200. | `[V]` |

Three of those are doing structural work and should be lifted by shape, not by number:

1. **Mass→speed is the entire balancing force.** Growth is self-limiting: the bigger you get, the less you can catch or flee. It is a *softcap on aggression* expressed as physics — which we must be careful with, because **design law 5 is hardcaps, never softcaps.** The reconciliation (`§6b`): mass→speed is a *movement* curve, not an economic softcap; the *reward* ceiling is still a visible hardcap. We keep the slow-whale physics (it is the fun) and we do not let it stand in for a balance cap on payouts.
2. **Splitting is the risk verb.** You split to catch prey or cover ground, but each piece is smaller, slower to re-merge, and individually edible. It is the genre's only real decision, and it is a *tempo* decision — exactly where the lane-pusher dossier found real-time depth lives.
3. **Viruses + mass decay are the anti-snowball.** The genre already knows that unchecked growth is un-fun to play against, and its two answers are "the big get split" and "the big slowly bleed." **These are the anti-bully mechanics, native to the genre.** (`§6b` builds our antitrust skin on them.)

### 1.2 The netcode reality — how, and what it costs `[M]`

None of this is on the fetched pages; it is well-known engineering, marked `[M]` and **not for shipped copy without re-verification.**

- **Server-authoritative position.** The server owns every cell's position, mass and collision outcome. The **client sends only intent** — a cursor heading (a normalized direction vector) and discrete `split` / `eject` events. It never sends "I ate that cell." **This is a perfect fit for design law 2** (clients send intents, never results) and it is why absorption, despite being real-time, is not an anti-cheat nightmare: the authoritative sim is the only place an eat happens.
- **Fixed-tick sim.** The arena advances on a fixed timestep (~20–25 Hz `[M]`). Collision is resolved with a spatial index (uniform grid / quadtree) so each cell only tests neighbours — O(n) not O(n²).
- **Area-of-interest (interest management).** The server sends each client only the entities inside its viewport plus a margin, not the whole arena. **This is the single most expensive part of the many-human design** and it is *also* a hidden-information boundary we must respect for bots (law 6): you cannot see what is off-screen, and neither can a fair bot.
- **Client interpolation / extrapolation.** The client renders between authoritative snapshots (interpolation) and predicts its own cell's motion for responsiveness (soft prediction, reconciled against the server) — it never *authoritatively* re-sims. **Consequence for us:** unlike the pet duel and the lane (`creature-battler.md` / `lane-pusher-design.md`, which run the sim on *both* Go and TS for prediction parity and therefore carry a cross-runtime golden-vector burden), an absorption arena's client is a **thin interpolating renderer**. The arena sim runs in exactly one place — the Go server — which **removes the cross-language determinism burden from the arena entirely** (the string-wire `Decimal` law still governs the *economy payout*, not the physics).
- **The cost, stated plainly:** the expense scales with **humans-per-arena**, because each human needs an individually culled, individually interpolated stream at tick rate. Bots and snapshot cells cost only their physics (cheap). **Drop the human count to 1 and the interest-management/broadcast tier disappears** — you are left with a single collision sim and one output stream, i.e. another Match goroutine.

### 1.3 Bot behaviour `[M]`

A basic candidate policy seeks smaller cells, flees larger cells, avoids hazards when vulnerable
and drifts toward food when safe. Reaction radius, split aggression, edge avoidance and hazard
awareness could form visible skill dials, bounded to human-visible information. Adequacy, code size
and operating cost are unmeasured; an adversarial bot prototype must establish them.

---

## 2. Katamari Damacy — roll-up scale as the whole feel

Paraphrased from [Katamari Damacy](https://en.wikipedia.org/wiki/Katamari_Damacy) (fetched
2026-08-06, `[V]`):

- **The stick rule:** sufficiently small objects adhere on contact; size, weight, surface area and
  shape affect eligibility. `[V]`
- **The scale arc:** pickup classes expand from tiny objects to buildings and landscape-scale
  objects without changing the core verb. `[V]`
- **The objective:** Make a Star requires reaching a specified size within a time limit. `[V]`

**What Katamari owns that agar.io does not: the *reveal* of scale as a designed sequence.** In agar.io scale is emergent and flat (a big cell and a small cell look the same, only bigger). In Katamari the joy is the *threshold unlock* — the instant a class of object flips from "wall you bump into" to "thing you absorb." **This is the mechanic our tier ladder and our diorama already run** (`13-world.md`'s office→campus→city→planet silhouette progression). So Katamari's contribution to Cloud Clicker is **not a second arena** — it is the **ceremony/montage layer** and the **threshold-unlock feedback** that a bare agar.io arena lacks. Design implication in `§6d`: the M&A Arena should *announce* scale thresholds ("you can now acquire: Series-B startups"), Katamari-style, rather than let size be a silent number.

---

## 3. Osmos — the elegant single verb, and "growth has a cost"

Paraphrased from [Osmos](https://en.wikipedia.org/wiki/Osmos) (fetched 2026-08-06, `[V]`):

- **Movement *is* mass expenditure.** A mote expels mass to move and shrinks in the process. `[V]`
- **Absorption is a pure size hierarchy** — colour-coded: smaller motes are blue (edible), larger are red (lethal); touch a smaller one and grow, touch a larger one and it is game over. `[V]`
- **Level families:** *Ambient* (stationary motes, sometimes antimatter that shrinks on contact regardless of size), *Sentient* (active predator motes), *Force* (Attractor motes create orbital physics). `[V]`
- **The design thesis:** interaction centers on absorbing smaller entities, avoiding larger ones
  and spending mass to move. `[V]`

**Osmos's gift is the honest cost of growth.** Its propulsion makes movement a price as well as a
risk, which suggests a capital-expenditure metaphor. **Prototype hypothesis:** an Osmos-inspired
propulsion-cost toy may be clearer as a separate, non-ranked experiment rather than part of a
competitive arena. This is not a roster decision.

---

## 4. The .io genre broadly — server model, sessions, bot backfill

### 4.1 slither.io, verified `[V]`

From [Slither.io](https://en.wikipedia.org/wiki/Slither.io) (fetched 2026-08-06):

| Fact | Value | Provenance |
|---|---|---|
| **Growth** | eat pellets; a defeated snake's body "turns into bright, larger shining pellets" — **killing is the biggest food source** | `[V]` |
| **Session shape** | continuous; the avatar "remains in constant motion"; you die when your head hits another body or the boundary; **no session timer** | `[V]` |
| **Bots** | mobile AI mode: "all the other snakes are computer opponents," named `…(bot)`; player is faster in AI mode; pellet/skin behaviour differs | `[V]` |
| **Capacity** | launched supporting **500 players** per server; devs later targeted **600**, with stated stability/bandwidth difficulty | `[V]` |
| **Matchmaking** | none at launch; server-select added by **Dec 2025** | `[V]` |
| **Leaderboard** | top-ten by mass, server-wide | `[V]` |

Two things to lift. **(a) "Killing is the food source" is the loop's engine** — you don't just eat neutral pellets, you eat *players*, and their mass becomes food. In our skin, a failed rival company's assets hit the market. **(b) The single-player AI mode is the whole genre's honesty proof:** slither.io *ships a mode where every other snake is a bot* and it plays fine. That is not a fallback bolted on — it is evidence that a bot-populated arena **is the game**, which is the load-bearing premise of `§5`.

### 4.2 diep.io and the genre's session model `[M]`

- **diep.io** adds a *build* layer (your cell is a tank that levels up and picks upgrade branches) on the same server-authoritative shared-arena base `[M]`. Its lesson for us is a caution: the moment you add persistent build depth to a shared arena you re-introduce the balance-treadmill the lane-pusher dossier warned about. **Keep the M&A Arena's cell simple** — size and position, maybe one or two ability verbs (split, divest) — and put durable progression in the *economy hook*, not the cell.
- **Session length `[M]`:** .io sessions are short and self-terminating — you live until you're eaten, typically a few minutes; there is no round timer, the *arena* is persistent and *you* are ephemeral. This matches our "**session skill** clock" (`03`'s clock taxonomy: "short skill runs, cooldown-gated"). One run of the M&A Arena is one company's lifespan in the arena, minutes long, then a cooldown.
- **Genre origin `[M]`:** agar.io was built by Matheus Valadares (2015) in a very short time on a Node.js stack; slither.io by Steve Howse (2016). The genre's defining trait is *cheap to stand up, expensive to scale to many concurrent humans* — precisely the tension `§5` resolves by not scaling human count.

### 4.3 Matchmaking and the whale problem

.io games seed you in **small** and let the arena's existing size distribution be the difficulty. That is fine when the arena is bots + snapshots (you control the distribution) and dangerous when it is live whales (a new player is lunch). **We seed by tier band** (`§6c`): your arena is populated with companies near your size, so — like agar.io's fresh-spawn safety — you start among peers, not among giants.

---

## 5. The server-authority + AI-fallback problem — and the tractable form

This is the section the brief demanded, in the shape `lane-pusher-design.md §5` used.

### 5.1 The three candidate forms

| Form | What it is | Server cost | Bottable? | Replay/anti-cheat | Verdict |
|---|---|---|---|---|---|
| **A. Live many-human shared arena** (true agar.io) | 8–200 humans in one authoritative arena; interest-managed broadcast per client at 20+ Hz | **High** — the full .io netcode bill; the cost the brief warns about | Yes (bots fill empty slots) | Server-authoritative position ✓; but floating multi-human physics is hard to make bit-exact for replay `[M]` | **v-next / tournament stretch only** |
| **B. Single-human arena, bot+snapshot populated** | One human per Match goroutine; every other cell is a bot or a read-only company snapshot | Structurally avoids many-human fan-out; actual cost unmeasured | Candidate native fallback; quality unproven | One input stream + seed may permit Go-only replay; must be prototyped | **Prototype candidate** |
| **C. Turn-based / async "acquisition"** | Abstract M&A as bids/turns, no continuous space | Trivial | Yes | Trivial | **Rejected** — loses the genre's entire feel; absorption is spatial-continuous, and a turn-based bid game is a different (worse, already-covered-by-the-Market) minigame |

### 5.2 Why B is the answer, in one move

Form B removes per-human fan-out beyond one view, which is a useful structural reduction. The
remaining collision, serialization, snapshot and bot costs are not measured. Similarity to another
Match-goroutine shape does not establish an equal cost class; a bounded prototype and load profile
must do that before an RFC relies on affordability.

And it satisfies every law without strain:

- **Server-authoritative (law 2):** the client sends a heading vector + split/divest events; the server resolves all eats. Identical to agar.io's own model `[M]`.
- **AI fallback (law 6):** there is no fallback to bolt on because **the arena is bot-populated by construction** — slither.io ships exactly this and it plays fine `[V]`. Human presence is *additive* (see the small-pod middle, below), never *required*.
- **No offline victim:** the other companies are **read-only snapshots** — eating a snapshot writes nothing to its owner's save (the `creature-battler.md §7.4` / `lane-pusher §4.4` rule). Your own company can appear as a snapshot cell in others' arenas and lose nothing when eaten there. Async MMO by leaderboard + snapshot presence, exactly the lane-pusher pattern.
- **Deterministic/seeded:** `(seed, ordered inputs) → replay`, re-run on the Go sim in fixed-point. Because the client never re-sims authoritatively, the arena carries **no cross-language golden-vector burden** (a real simplification over our two combat engines).
- **Bots never cheat (law 6):** the bot's "sight" is bounded to a human's viewport radius — it cannot target a cell it could not see on screen. Interest management, which was a *cost* in Form A, is here a cheap *fairness constraint* on the bot.

### 5.3 The scalable middle (the honest MMO path)

Between B and A sits **small-pod live: 2–8 humans in one arena, bots backfilling empty seats.** Its
network and simulation cost is unmeasured. **Research trajectory to test:** prototype B first;
consider small-pod or many-human variants only after measurement and a separate product decision.
This does not assign either variant to a release.

### 5.4 Is real-time server sim unavoidable? — the honest flag

Form B would still be a real-time server simulation for each active run. Its proposed 20 Hz rate,
2–5 minute duration and per-player resource profile are hypotheses, not accepted budgets. Form A
adds multi-human fan-out and coordination concerns. An owner/RFC decision can compare them only
after a prototype records CPU, memory, bandwidth, concurrency and replay behavior.

---

## 6. Non-adopted design proposals for Cloud Clicker

### 6a. The tractable server-authoritative form (the pick, justified)

**Prototype candidate:** a single-human, real-time, server-authoritative absorption arena,
bot- and snapshot-populated. No release or implementation is authorized here.

| Decision | Value | Why |
|---|---|---|
| **Humans per arena** | Start prototype at **1**; later population is undecided | `§5.2` — isolates the simplest measurable shape |
| **Population** | bots (honest, viewport-bounded) + read-only company snapshots | law 6; no offline victim (`§5.2`) |
| **Authority** | server owns position/mass/eats; client sends heading + split/divest intents | law 2; agar.io's own model `[M]` |
| **Determinism** | `(seed, inputs) → replay`, fixed-point integers, Go-only re-sim | client is a thin interpolating renderer → **no cross-runtime golden vectors** for the arena (Decimal string-wire law still governs payouts) |
| **Clock** | **session-skill**, cooldown-gated (`03` clock taxonomy) — a run is one company's lifespan, minutes long | matches the .io ephemeral-session model `[V/M]` |
| **Cell complexity** | size + position + two verbs (**Split = spin-off**, **Divest = eject mass**); *no build tree* | diep.io caution (`§4.2`) — durable progression lives in the economy hook, not the cell |
| **Type chart** | **none** | absorption is a size/position axis, not type/stat — this is what keeps it from collapsing into the pet battler / lane (`§0`, addresses lane-pusher `§6`'s "same game thrice" gate) |
| **Host tier** | Tier 4+ `[non-adopted design proposal]` | thematic hypothesis only; current release placement is unresolved |

### 6b. Proposed economy hook, AI fallback, and relative-size reward experiment

- **Economy hook (declared clock = session-skill, cooldown-gated):** a run pays **Clout + tier-scaled cash**, with absorbed mass converting to a **one-shot cash harvest** at run-end (the Server Garden's harvest pattern, `03 §1`). Reaching the top of the arena's leaderboard pays a seasonal title. **Deck/verb unlocks by play, never purchase** (there is no purchase — law 1). The hook is bounded and published (law 9).
- **AI fallback:** native — the arena is bot-populated by construction (`§5.2`), bots are ~a few dozen lines of seek-smaller/flee-bigger with a **printed behaviour-flag skill dial** (`§1.3`, reusing lane-pusher `§4.3`'s ladder: `Intern → … → The Autoscaler`). Bots are viewport-bounded (never cheat).
- **Relative-size reward hypothesis.** Absorption has an intrinsic bully gradient, so the historical
  **200% up / 5% down** example could be tested as an acquisition-payout curve. Those constants are
  not adopted, and the moderation, abuse and balance costs are not zero or measured. Two related
  candidate anti-snowball mechanics are:
  1. **Antitrust = agar.io's virus.** The biggest cell in the arena attracts **regulators** (the spiky virus reskinned): touch one while large and you are **split apart** (agar.io's exact "big cell hits virus → shatters" rule `[V]`); small companies can shelter near a regulator that would shred a giant. *The biggest fish = antitrust attention*, mechanically true.
  2. **Diseconomies of scale = mass decay.** agar.io's "cells gradually lose mass over time" `[V]` becomes regulatory drag / bureaucratic bloat: the bigger you are, the more you bleed if you stop acquiring. A visible **hardcap on cell size** (law 5) sits on top — the softcap-looking slow-whale physics stays as *movement feel*, but the reward ceiling is an explicit number (`§1.1` reconciliation).

### 6c. Matchmaking, anti-farm, seeding

- **Tier-banded seeding:** your arena is populated with company snapshots near your size, so a new player starts among peers, not whales (`§4.3`). This is agar.io's fresh-spawn safety made explicit.
- **Anti-farm:** the snapshot read-only rule + no-repeat + rating-deviation stack from `creature-battler.md §7.4` transfers unchanged; bot runs pay reduced, non-ranked rewards (the `03 §5` anti-farm convention).

### 6d. The tech-satire skin

*(Naming is mechanical in code — `absorption_cell`, `split`, `eject` — flavour lives in data files, per CLAUDE.md.)*

| Mechanic | Skin | Source |
|---|---|---|
| **Eat a smaller cell** | **Acquire a competitor** — M&A as literal consumption (the thesis in miniature) | agar.io eat |
| **Split** | **Spin-off** — divide to move faster / cover more ground, but each spin-off is smaller and individually edible (over-splitting = the conglomerate that fragments and gets picked apart) | agar.io split `[V]` |
| **Feed / eject mass** | **Divest / strategic investment** — hand mass to another cell; the "teaming signal" becomes a satirical alliance/collusion beat | agar.io feed `[V]` |
| **Green/red virus** | **Antitrust regulator / the DOJ** — shatters the giant, shelters the minnow | agar.io virus `[V]` |
| **Mass decay** | **Diseconomies of scale / regulatory drag** — "the bigger you are, the more you bleed" | agar.io decay `[V]` |
| **Killing = food** | **A failed rival's assets hit the market** — you grow on the corpse | slither.io `[V]` |
| **Scale-threshold reveal** | **"You can now acquire: Series-B startups → unicorns → public companies → nation-states"** — Katamari's threshold-unlock announced, not silent | Katamari `[V]` |
| **Katamari rollup ceremony** | the **endgame montage** — roll up office → campus → city → planet — reuse the `13`-world diorama silhouette scale (post-1.0 ceremony, not the arena) | Katamari `[V]` |
| **Osmos propulsion-cost** | **capex: expansion is never free** — the arcade cell-stage toy / Soul-safe cozy verb, a *separate tiny build*, not fused into the arena | Osmos `[V]` |

**The satire writes itself:** the win state of eating everyone is *becoming the thing regulators exist to stop*, and the arena's own anti-snowball mechanic is antitrust. The biggest fish doesn't win — it gets broken up. That is both good game design (anti-snowball) and the exact point the game is making.

### 6e. In-game (8)

1. **`The M&A Arena`** — host: the finance floor / a war-room map table. A run is one acquisition spree; you enter small and grow by eating.
2. **`Spin-Off` (split)** — the reach/risk verb. Tooltip: *"Unlocks shareholder value. Also unlocks you, to being eaten."*
3. **`Divest` (eject)** — shed mass to move, to fund an ally, or to slip under a regulator. Osmos's honest cost, agar.io's feed.
4. **`The Regulator`** — the virus. A slow spiky entity that shatters any company over the size hardcap. Achievement `Too Big To Fail` for surviving a regulator brush at max size.
5. **`Diseconomies of Scale`** — the decay meter, printed. The bigger the cell, the faster it ticks down.
6. **Acquisition payout multiplier, printed on every eat:** `+200% (punching up)` in green, `+5% (punching down)` in grey — the God-Hand-rule visible number (law 10).
7. **`Whale Watch`** — the leaderboard; top-ten companies by market cap (mass), arena-wide, snapshot identities labelled *"not present."*
8. **Ticker** — *"BREAKING: MegaCorp acquired its 47th startup this quarter and has been referred to the Antitrust Regulator, which is on its way and is bigger than you are."*

---

## 7. Verify before ship

| # | Claim | Status |
|---|---|---|
| 1 | agar.io mechanics (mass→speed, decay, 16-cell split, virus +100/+200, 200-mass threshold) | `[V]` [Agar.io](https://en.wikipedia.org/wiki/Agar.io). Fandom/Wikipedia drift; re-check any number printed as a fact *about agar.io* in shipped copy. |
| 2 | **All .io netcode internals** (20–25 Hz tick, area-of-interest, interpolation, delta compression, Node origin, authors/dates) | ❌ `[M]` — **not on any fetched page.** Well-known but unsourced here. Re-verify before any RFC acceptance criterion or shipped copy leans on a specific number. |
| 3 | slither.io facts (500→600 players, no timer, AI mode, kill-as-food, Dec-2025 server select) | `[V]` [Slither.io](https://en.wikipedia.org/wiki/Slither.io). |
| 4 | Osmos facts (expel-to-move, size hierarchy, ambient/sentient/force levels, single-verb) | `[V]` [Osmos](https://en.wikipedia.org/wiki/Osmos). |
| 5 | Katamari facts (stick rule, size/weight/surface-area, thumbtacks→buildings, Make-a-Star timed) | `[V]` [Katamari Damacy](https://en.wikipedia.org/wiki/Katamari_Damacy). |
| 6 | **"Form B costs the same as the lane pusher"** | `[sim]` — engineering argument, not a benchmark. The claim is *directional and structural* (deleting the many-human broadcast tier removes the dominant cost), not a measured figure. **Prototype a Go collision sim and measure ticks/sec before committing the cost claim to an RFC.** |
| 7 | **diep.io build layer / genre session lengths / agar.io authorship** | `[M]` — model knowledge, no source fetched. Not load-bearing for the design decision; re-verify if quoted. |
| 8 | Cross-references to `lane-pusher-design.md` (§4 async, §5 spec, §7c multiplier) and `creature-battler.md §7.4` (anti-farm) | ✅ read and consistent as of this writing. If the punch-down multiplier or the anti-farm stack changes there, `§6b`/`§6c` change with them. |
| 9 | **Host tier = Tier 4+** and clock = session-skill | `[design proposal]` — mine, not adopted. `03`'s unlock stagger does not yet list the arena; the RFC or an owner ruling must place it. |
| 10 | Patent exposure on shared-arena absorption mechanics | ❌ **Not searched** (WebSearch budget exhausted). Exposure is probably nil (free, EU, non-commercial, and the mechanics are widely cloned), but run a real search before an RFC — same caveat as `lane-pusher-design.md §8` item 13. |

---

## 8. Legal issue-spotting — not clearance

The general mechanics/expression distinction in *Tetris Holding v. Xio* is a useful caution, not
project-wide clearance. No patent or trademark search was performed, and exact expression,
trade-dress, naming and jurisdiction-specific questions require current review before an RFC or
shipped surface relies on them. The matrix below records research risks only.

| 🟢 Lower concern hypothesis | 🟡 Fictionalize / review | 🔴 Exclude pending review |
|---|---|---|
| General absorption, movement, hazard, leaderboard and server-authority mechanic shapes may be lower concern than copied expression, subject to review. | Distinctive visual expression, presentation and trade dress require deliberate differentiation. | Third-party titles, characters and marks are excluded from shipped text absent explicit clearance. |
| Factual comparison may be supportable with accurate, source-adjacent reporting; this is not a shipped-copy recommendation. | Real vendor/regulator naming needs editorial and legal review; generic vocabulary is not a blanket clearance. | Do not depict a real company as the subject of fictional acquisition, failure or enforcement without specific review. |

**Non-adopted naming candidates:** **The M&A Arena** (the mode) · **Market Cap** (mass) ·
**Acquire** (eat) · **Spin-Off** (split) · **Divest** (eject/feed) · **The Regulator** (virus) ·
**Diseconomies of Scale** (decay) · **Whale Watch** (leaderboard) · **Capex** (propulsion cost).

---

## Sources

**Fetched Wikipedia pages (`[V]`, 2026-08-06):** [Agar.io](https://en.wikipedia.org/wiki/Agar.io) · [Osmos](https://en.wikipedia.org/wiki/Osmos) · [Katamari Damacy](https://en.wikipedia.org/wiki/Katamari_Damacy) · [Slither.io](https://en.wikipedia.org/wiki/Slither.io)

**Not fetched (WebSearch budget exhausted 200/200 this session):** .io netcode internals (area-of-interest, tick rates, interpolation), diep.io mechanics, agar.io authorship/origin, and any patent search — all marked `[M]`/`[not searched]` above and gated in `§7`.

**Cross-references (internal, read and consistent):** `research/lane-pusher-design.md §4, §5, §7c, §8` · `research/creature-battler.md §7.4` · `research/flash-era-arcade.md §4` (Bubble Tanks TD merge-to-grow) · `design/03-minigames.md` (clock taxonomy, anti-farm) · `design/BACKLOG.md §Viral/absorption sweep` · `design/13-world.md` (diorama silhouette scale).

**Legal:** *Tetris Holding, LLC v. Xio Interactive, Inc.*, D.N.J., 30 May 2012 (per `lane-pusher-design.md §9`).
