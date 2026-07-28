# The Pet Layer

> Tamagotchi + pokemon battles + base building + hat collecting + free lootboxes — all the meme-y, usually-microtransaction-shaped content, deliberately free. Mechanically: **the cattery systems, lifted nearly wholesale** (`research/cattery-reusables.md`), given a tech-satire skin and wired into the Soul axis. The pet layer lives on the **founder** scope: it survives every Exit. The company dies; the cat does not.

## 1. The pet

**Fiction:** every founder gets a companion at Tier 0 — default: the **server-room cat** (industry-canon). Later adoptable variants: a robot vacuum with a personality, a rubber-duck (debugging canon), a Tamagotchi-style keychain blob, the Blucifer-red-eyed horse (conspiracy-tier unlock).

**Core sim (cattery numbers as starting values):**
- Four stats 0–100 (default 75): `Fed`, `Happy`, `Energy`, `Clean`. Elapsed-time decay (kitten 3/h, adult 2/h, elder 1/h) computed from `last_tick_at` — fully offline-correct.
- Care actions with variant overrides and diminishing returns (>90 → 0.5×): feed (+40 Fed…), pet (+20 Happy, −5 Energy), play (+25 Happy, −15 Energy…), groom (+25 Clean), passive rest. Item variants (fish/pellets/treat/yarn/feather/brush equivalents → server-room skins: thermal paste treat, cat6-cable yarn, anti-static brush).
- **Personality** (lazy/playful/curious/sassy/shy/chaotic): drives food-preference thresholds (a hungry sassy cat refuses pellets and tells you), play gating by energy, and the behavior state machine (activities → behavior chain queues — cattery's two-level AI, reused).
- **Visible stat feedback:** low Fed slows the walk animation; low Clean desaturates; low Happy droops the ears. Over-petting annoys; 3+ gentle pets triggers kneading + purr. Cats ask for things by walking to your cursor with an emoji.

**Trust & mood (the two-tier system, reused):**
- **Trust** (server-side, per-founder, long-term): +2 feed / +3 play / +1 pet / +1 groom, cap 100, decays 1/day of absence. Trust gates the good stuff: slow-blinks, kneading, lap-sitting, battle obedience.
- **Mood** (session-scoped): short-term reactions.
- **Trust is the Soul proxy** (`02-economy-balancing.md §8`): as founder Soul drains, the pet's recognition degrades *regardless of care* — UI greys, the cat stops coming when called, eventually it just watches you. Restoring Soul restores the bond. This is the game's most direct emotional instrument; play it straight, never as a joke.

**Bonds:** pets of guildmates and visitors build affinity (cattery bonds table); bonded pets visit each other's houses and are seen grooming each other in the feed.

## 2. Pet battles (pokemon-style)

- **Every pet reaches the same hardcapped stat ceiling** (design law 5, applied honestly — decided 2026-07-28 per `research/creature-battler.md §3.4`: in a 1v1 duel a 2.1% stat edge is a ~99.6% win rate, so any care→stat mapping wide enough to feel meaningful is a binary win condition). **Care buys options, consistency, and tempo, never raw stats:**
  - **Trust → Obedience:** slopes smoothly (~50%→30% disobedience across Trust 1.00→0.80) — a neglected pet ignores orders at the worst moment; a loved one is *reliable*, not stronger.
  - **Care quality → insurance & luck** (the Pokémon-Amie template): crit-rate doubling, survive-at-1-HP procs, faster stamina recovery.
  - **Soul drain is legible in combat**: as founder Soul empties, Obedience degrades *regardless of care* — the pet looks at you before ignoring the command. Play it straight.
  - A well-loved elder cat is still a monster — through options and reliability, which is the emotionally correct version anyway.
- **Moveset** by personality + learned tricks (taught via play minigames): `Pounce`, `Zoomies` (priority), `Loaf` (defense up), `Knead` (heal), `Hairball` (debuff), `Headbonk`. Type-triangle lite: Playful > Lazy > Sassy > Playful (a rotating weakness — the Clicker Heroes daily-rotation trick appears in tournament seasons).
- **Turn-based, ~2–3 minute matches.** Server-authoritative; same engine as board-game matches.
- **AI fallback:** battle bots at every rating tier (minimax over the small move space + personality-flavored policies: a "lazy" bot plays Loaf too much — human-feel by characterization). Named NPC trainers with fake ratings.
- **PvP:** async by default — you battle a snapshot of another player's pet with its own AI policy (their pet, their trained tricks, no wait for the owner). Live ranked matches during tournament events. Rewards: Clout, cosmetics, seasonal titles. Bot matches non-ranked.
- **Tone guard:** pets are never hurt — battles are play-fights; the loser gets the zoomies and a treat. (The satire quota is filled elsewhere; the pet layer is the sincere heart the Bogost rule requires.)

## 3. The house (cosmetic base — no raids)

> Re-scoped 2026-07-28: Clash-style layout/defense/raids **struck** (owner rejection; `research/lane-pusher-design.md §7` — layout is a solved puzzle players outsource: 82% of raid outcomes, one optimum, copying worth +41pp). Competitive spatial play lives in the Lane (`03 §10`) and the region draft (`13 §5`). Kept: everything below.

- **Fiction:** the founder's home/base, growing with tiers: garage corner → apartment → smart home → compound → bunker (the billionaire-layer rungs appear here: the yacht is technically a room).
- **Decor play:** place furniture, pet equipment (climbing towers, server-rack hammocks), and cosmetics; bonded pets visit and are seen lounging in the feed. Guildmates can tour your house.
- **Economy hook:** furniture from cash + Influence; pieces buff pet stats or house **Comfort** (a small Soul-regen multiplier — the home is a touch-grass amplifier).
- **The pet is the On-Call Leader** in the Lane (`03 §10`) — its house display shelf shows lane trophies.

## 4. Hats, cosmetics, and the free lootbox machine

The full Almanac §6 parody suite (`research/gaming-enshittification.md`), implemented:

- **Hats.** Hundreds. For the pet, the founder avatar, buildings, and the cursor. `War-Themed Hat Simulator` achievement at 50.
- **Horse Armor** — FREE (~~200 Microsoft Points~~), the pet is visibly annoyed by it; `Horse Armor (Remastered)` at a later tier, better textures, same zero price, captioned "we did it again."
- **Unusual effects** — free particle auras (Burning Flames, Sunbeams, Circling Peace Sign); market value readout: `$0.00 (priceless)`.
- **The Knife** — one Case Hardened Blue Gem #387; tooltip valuations inflate the longer you own it; does nothing; cannot be sold; the most beautiful object in the game.
- **Default Skin (Legendary)** — "Rarity is a social construct."
- **Surprise Mechanic Crates** — open instantly; keys free ×999,999,999; odds table all `100.00%` ("Odds sum to 4,700.00%. This is normal."); pity counter pinned at 1; you win both 50/50s; Kompu Gacha (Legal Edition) hands you the set; near-miss animations apologize.
- **Gacha banner art** — over-produced splash illustrations for 100% drop rates. The art budget goes here on purpose.
- **Vaulted-forever countdowns** that end with everyone getting the item and the timer restarting.
- **AI Slop line** — deliberately six-fingered, disclosure box included, finger count varies per viewing.
- **Battle Pass (Free Track / Free Track)** — two identical free tracks, pre-purchased, timer counts up.
- **Acquisition paths:** crates drop from play everywhere (achievements, quarters, battles, lane matches; **event rewards go through the exchange shop** — `09 §5` — where event crates are shop items with 100.00% disclosed odds); duplicates convert to `Compliance Points` (which convert 1:1 to everything, obviously). **Nothing is ever purchasable with real money. There is no real-money anything.**
- **Collection book** with completion achievements feeding Clout — collecting is load-bearing, joyful, and free. That's the whole point.

## 5. Breeding & rarity

- Light breeding: bonded pets (guild/visits) can produce an egg with mixed personality weights + a fur-palette blend (cattery's 10-swatch CSS palette generalizes — recolorable sprites make this free). Rare palettes exist; drop rates disclosed at parody-maximum honesty. No stat inheritance power-creep — cosmetic + personality only (care quality remains the only battle-power input).
- Elder pets retire to the house permanently (never die — this is not that game); retired elders give a small passive Soul regen ("the old cat sleeps on the warm server. p(doom) feels lower.").

## 6. Rendering

Cattery's CSS-only sprite system reused: nested-div creatures styled by custom properties, data-attribute poses, zero image assets, infinitely recolorable, `prefers-reduced-motion` respected. Physics/steering primitives (drag, cursor avoidance, nudgeToward) reused for the house/world ambience. Ambient event system (birds at the window, a bulldozer passing — escalating with tiers to drone deliveries and protest marches) reused for life.
