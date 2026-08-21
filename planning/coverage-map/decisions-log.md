# Coverage-Map Decisions Log

Owner rulings on the open decision points the validation sweep surfaced. Append-only.

## 2026-08-05 — five owner rulings

1. **Trading / player market → NPC/aggregate market FIRST, architected for P2P later.**
   The RFC targets an order-book against an aggregate/NPC counterparty (research-backed, sidesteps
   the RMT surface). BUT the systems must be designed with **full flexibility so a player-to-player
   layer can be added without a rewrite** — the market's order/settlement model, the
   non-transferable-currency flag, and the anti-cheat boundary are to be built P2P-ready.
   **Follow-up: commission P2P-trading research** so the expansion is spec-backed when we reach it.
   (Backlog: trading = Wave-B NPC/aggregate; P2P = a later expansion gated on its research dossier.)

2. **Lobbying / Influence-spend → commission a research dossier FIRST.**
   It is load-bearing late-tier satire (regulatory capture, the broligarchy arc) and entirely
   unbacked. No RFC until a dossier specs the mechanic. (Backlog: lobbying stays Wave-C, now
   explicitly research-gated. Research dispatched 2026-08-05.)

3. **Soul → its own dedicated Soul RFC.**
   Soul is NOT a read-only carry and NOT a sub-field of the meters system. It gets a standalone
   foundation RFC with its full surface: drain verbs (crunch, Faustian VC term-sheets) and recover
   verbs (touch-grass activities that cost time and produce nothing), on the Founder scope via
   `ApplyFounderLogged`; it gates the human content (pet recognition greys out, hobby minigames
   lock) and the transcendence ending ("transcend at zero → you become training data"). This
   **un-orphans Soul** — the Meters-RFC read-only-carry demotion is superseded by this decision.
   (Backlog: Soul promoted from an orphaned-decision to a named foundation RFC to draft.)

4. **Review governance → option (c): BOTH gates.** Encoded in `AGENTS.md`
   ("Two review gates, both mandatory before archival"): (a) range-union binds delegated/self
   reviews too; (b) a designated independent adversarial pass is a mandatory archival gate
   regardless of any delegated approval. Resolves the 8-batch standing escalation.

5. **"No FOMO" refined from a hard rule to a consequence-balance target.**
   The binding constraint is **no net *loss* inflicted on a player for being offline** — not "no
   offline interaction." Asymmetric mechanics where an active player acts against an absent
   player's snapshot are permitted *if the absent side loses nothing* (loss-decoupled defense, the
   Clash "Clash Anytime" clone-base model — attacker earns from a clone, defender loses no
   resources/standing). Encoded in `design/00` anti-goals. **Consequence:** offline raiding is
   rejected only in its naive (real-defender-loss) form; a loss-decoupled form is on the deferred
   list (`deferred-and-dropped.md`), revivable — not banned. Marco flagged the intuition that a
   simplified base-build + active-defense/TD minigame may earn a slot later.

## 2026-08-05 — four Wave-A design rulings

6. **Fiscal Quarters scope → Founder-scoped, save v19.** Research-backed (sugar lumps persist across
   ascension and drive the months-long tail; CC's non-persisting Stock Market "feels wasted" — the
   failure this avoids). RFC updated; scope no longer open. Chain: minigames 17 → pets 18 → fiscal 19.

7. **Clicking → RETAINED as a timing skill (not dropped); server click-clamp confirmed.** Clicking is
   THE basic combo driver. Enrich it with rhythm/timing-game mechanics — owner named **Crypt of the
   NecroDancer** and **osu!**. Research dispatched (`rhythm-timing-games.md`) to (a) find a
   server-validatable timing model and (b) surface candidate rhythm/timing minigames. The active-play
   buff-window FOUNDATION (Lucky bank, golden opportunities, multiplicative combos, daemons, click
   clamp) is drafted now; the rhythm-timing click-input enhancement is a research-gated successor.

8. **Soul content-gating → GRADUATED** (soft greying/flavor early, hard mechanical locks only near-zero
   Soul). Telegraphed loss, no cliff; keeps satirical teeth at the bottom. Bakes into the Soul RFC.

9. **Soul meter-vs-currency → EXPLORE BOTH before deciding.** Owner wants both models worked out.
   Research dispatched (`soul-mechanic.md`) laying out the METER model and the CURRENCY model as two
   concrete designs with precedents (Darkest Dungeon/Amnesia/Sunless Sea meters vs Cultist Simulator/
   Disco Elysium/devil-deal currencies) against the sincerity law, with a recommendation. **The Soul
   RFC is HELD until this lands and the owner picks.**

## 2026-08-05 — Soul rulings + the three dossiers landed

10. **Soul model → METER-SHOWN, CURRENCY-TRIGGERED HYBRID** (owner ruling, from `soul-mechanic.md`).
    Soul is an always-visible condition read THROUGH THE PET (never a wallet UI), moved ONLY by
    discrete opt-in *clicked* debits at Faustian/crunch/longevity moments (curtain tooltip each) and
    deliberate recovery. Gets the meter's sincerity + the currency's ownership without commodifying
    the sincere pet-heart (the sincerity-law risk of a pure spend→power wallet). Soul RFC DRAFTED
    (`rfc/soul-foundation.md`).
11. **Soul recovery → DELIBERATE, OPPORTUNITY-COSTED; no passive idle refill.** Resolves the
    `design/02 §8` DESIGN-GAP the dossier caught (idle refill would make the drain toothless and the
    training-data ending unreachable). §8 reconciled. Touch-grass is a deliberate time-costed act;
    cheap for idle builds, a real tax for active — never a passive drift.
12. **Soul safety constraint:** the zero-Soul "training data" ending stays institutional metaphor,
    NEVER depicts self-harm/suicide (AGENTS.md §4); copy pipeline must flag violations.

### Research banked (three dossiers, all model-knowledge-led due to the WebFetch classifier outage —
numbers `[M]`-flagged on verify-before-shipping lists):
- `soul-mechanic.md` — the hybrid recommendation above.
- `rhythm-timing-games.md` — **server-validatable timing = the 20 Hz sim tick as the beat grid**
  (±1-tick server layer trusts no wall-clock; sub-tick Perfect stays cosmetic/solo). Feeds the
  active-play click successor (folded into `active-play-buff-windows.md` AB4). Candidate minigames:
  Build Cadence, Ship-on-the-Beat (⚠ patent check), Incident-Response Patapon, on-beat pet-leader
  abilities in The Lane.
- `roguelike-survivor-minigames.md` — **prototype "The Pitch" (Balatro-like) FIRST**: turn-based,
  deterministic-per-seed, PvE (AI-fallback nearly free), and its `score = chips × mult` IS our
  multiplicative buff-stack in the break_infinity regime — closer to skinning the shipped engine than
  a new minigame. Second: "The Tech Stack" (Slay-the-Spire-like). Real-time survivor-likes are the
  expensive trap; if any, Brotato's *wave boundaries* give server-certification checkpoints (not
  continuous VS, not 3D Megabonk). DESIGN-GAP flagged: an auto-battler would be our 3rd
  draft-team/fight-snapshot mode (duel + lane exist) — likely redundant; and law-10: any roster uses
  anonymous hires/services, never the sincere cats.

## 2026-08-06 — two DESIGN-GAP rulings from the research batch
13. **Board games vs Soul-recovery → cozy-only recovery.** The cozy/touch-grass category (zero reward)
    is the SOLE Soul-recovery source; board games PAY (Clout+cash) and LOCK at low Soul but do NOT
    restore it (a rewarding activity isn't restful — the Stardew trap). `design/03 §5` reconciled
    (struck the Soul-restore claim, kept the low-Soul lock). Cozy category to be added as §5c.
14. **Social deduction → keep as the designated HUMAN-ONLY feature.** Takes the plan's "human-only
    where AI is too much hassle" carve-out (an honest same-rules bot can't exist on hidden info).
    STRUCTURED communication only (preset claims/votes, never free-text). No bot backfill — empty
    lobby = unavailable. `design/03 §5b` reconciled (flag upgraded from design-risk to human-only,
    structured-comms, no-fallback). Ship only when structured deduction beats a coin flip (Break Room v2).

## 2026-08-06 — review-gate governance: the designated pass is CROSS-PARTY (Claude-side)
Ruled after Codex archived Relevance on a `Review by: Darwin, Recorded by: Codex` verdict (a
recorder-relabeled delegated review) without the cross-party gate. **The designated independent
archival gate is run by the OTHER agent — Claude reviews Codex's implementations.** Codex's own
reviewers (Darwin/etc.) are self/delegated first-filters that do NOT satisfy the gate and must not be
recorded as it. **The implementer never archives on its own review** — it hands off "ready for
designated review + archival" and waits for the cross-party verdict. Encoded in AGENTS.md rule (b)(c).
Relevance's code was independently re-verified APPROVE (retroactive), so its archival stands; the
breach was procedural.

## 2026-08-06 — repo disposition: private source of truth; public ship repo is separate & clean
Committing is permanent, so the private/public boundary is decided NOW, not deferred. **THIS repo is
the PRIVATE source of truth** — research dossiers, contracts, code all committed here permanently
(that's intended; it stays private). **The public/shippable repo is a SEPARATE, clean repo** created
at ship time from only the shippable layers (code, data, design docs, RFCs), NOT a history-filtered
fork of this one — so `design/research/` (which names games/trademarks/stats by design) never enters
public history. Research is kept private, NOT sanitized (sanitizing destroys its value). Marker added
to `design/research/README.md`.

## 2026-08-06 — CORRECTION to the repo-disposition entry: no second repo
The "separate clean public repo at ship time" idea is struck — nonstandard and unnecessary. The model
is: **one private repo; ship deployed artifacts (docker images), never the source.** Shipping a free
game does not require publishing source; players see only copy-pipeline-guarded shipped content. If
open-sourcing is ever genuinely wanted (unplanned), the mechanism is chosen then (private submodule or
history filter) — nothing is pre-built for it.

## 2026-08-07 — CORRECTION: repo disposition (final, executed) — supersedes both 2026-08-06 entries
The two entries above are stale. The executed, owner-ruled final state: **this repo IS the
publishable repo** — upstream `stronk-dev/destroy-humanity-any-percent` (GitHub, PUBLIC, empty,
never pushed). Research is **gitignored** (`design/research/*` with the single tracked exception
`provenance-extracts.md`) and was **history-filtered out** (570-commit filter, 2026-08-06, map at
`planning/history-rewrites/2026-08-06-unpublication-filter.map`; both history-walking guards
amended and green). RFCs are self-contained; design keeps lineage without recipe framing. No second
repo, no going private. THE PUSH remains Marco-only, standing.

## 2026-08-07 — world-layer dual-reward axis: CONFIRMED (was the last open scope question)
Resolved by Claude per the standing assignment. Elite-D Community Goals + GW2 percentile medals
back WL3 exactly as specced (`events-playstyles.md:219,338-339,542`); the ~half-pay-on-failure
clause is routed to the Layer-3 server-events successor, where failure states exist. gap-backlog
updated; no RFC change needed.

## 2026-08-07 — mint content rulings (Marco): B1 extend-routes; grievance 0/0 now, EU4-estates later; tiered achievement scores
FCE-B1 resolved by extending the routes artifact with gate.t3_to_t4 (reviewed base-byte change).
Grievance mints as initial 0 / decay toward 0; Marco's EU4-estates equilibrium model routed to the
design/09 Layer-2 pressure evaluator (dynamic decay targets need that grammar). Achievement scoring
ruled TIERED 2/4/8 (not flat 4); design/02 §6 owes a tiered amendment note. Details + applied draft:
mint-content-rows-proposal.md.

## 2026-08-08 — discipline-sweep owner rulings (two rounds)
1. **License: as-public-domain-as-possible.** Unlicense wherever dependencies permit; where a
   copyleft dependency (e.g. GPL chess engine) forces it, comply with the stricter license. NEVER
   closed source. → dependency-license audit added to the register (before THE PUSH).
2. **Law 10 stays as-is** (disclosure-only). The healthy-engagement "labeled AND bounded"
   amendment REJECTED; boundedness stays per-content judgment.
3. **Touch clicks: multitouch always counts; no clamp redesign.** The dynamic active-play
   mechanics (not spam-click-one-spot) are the real skill surface; we were overthinking.
4. **AI provenance: STANDARD PRACTICE ONLY** — CONTRIBUTING.md allowing genAI + about mention +
   credits. The dossier's "Every pixel human" premise was corrected (everything is AI-made under
   owner direction; art/audio public sources). AGI-tier self-learning-AI-breaks-rules content
   approved as a separate direction.
5. **Rested/return bonus: added as candidate** in design/02 (WoW rested-XP pattern; anti-streak).
6. **Beta: open world from wave one + diegetic EARLY ACCESS™ framing.** No invite gates; resets
   only as designed in-world events via the migration chain.
7. **ARG surfaces: all three approved as design directions** (corporate facade page on our domain,
   fictional parody-movement name, dev-commentary ending unlock) — each gets its own RFC; TIOAG
   ethics boundary binds.

## 2026-08-21 — CORRECTION: public repository with reviewed public shared memory

The 2026-08-07 disposition above is superseded in part. **This repository remains the one public
source of truth, and durable planning/research must no longer be hidden merely because it is
planning/research.** Public tracking is permitted only after file-specific sensitive-material and
publication-rights review plus any required sanitization, synthesis, specialist review or
historical/noncanonical labeling.

Ignored raw dossiers are local working source, not canonical shared memory, and tracked authorities
must not depend on their presence. Generated diagnostics remain ignored. No blanket unignore,
force-add, push, publication, product adoption or destructive cleanup is authorized. The exact
population, dispositions and fresh-clone closeout gate are owned by
`planning/platform-alignment/publication-sensitivity-audit.md` under owner decision D-002.
