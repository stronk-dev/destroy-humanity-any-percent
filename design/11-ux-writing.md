# UX & Writing

> How the game talks, teaches, and frames itself — and how we produce copy at the volume this design demands. Voice rules live in `08-satire-flavor.md §1`; this doc covers the surfaces, the run-narrative structure, and the production system. Research banked: `research/run-narrative-ux.md`. Structure below is the commitment; §6b's remaining adoption questions are ledgered in `BACKLOG.md` (walkthrough holes).

## 1. The contract screen (cold open)

The game states its conceit in the first ten seconds, before any clicking:

> **HUMANITY**
> *Category: Destroy Any%*
> Timer: 00:00:00.00 — PB: — — WR: 03:58:44 (Kardashev_Andy)
>
> [ BEGIN ATTEMPT ]

- One screen, era-1995 chrome, the timer already rendered. No lore dump — the category select IS the exposition (the Undertale/DDLC lesson: bold conceit stated up front, then play).
- Below the fold, small print in the honesty voice: *"Free forever. No purchases. No ads. The only thing this game harvests is the fictional planet."*
- Returning players land on the **splits screen** instead (your PB, your current attempt, what's new since last session — the welcome-back moment doubles as the offline-gains modal).

## 2. Teaching: diegetic, dripped, never a tutorial

- **No tutorial mode.** The Tier-0 era does the teaching by being small: one button, then one generator, then the shop appears (unfold-as-onboarding — the genre's native pattern).
- **The hint system is a character:** cattery's hint-overlay engine reused; hints are written as the era's UI talking (1995: a Clippy-shaped paperclip — which is also the Paperclips nod — offering era-appropriate "It looks like you're founding a company"; later eras: an onboarding-toast parody, then the AGI just doing it for you).
- **Progressive disclosure**: every new system arrives with exactly one contextual hint + one "why this matters" tooltip; everything else is discoverable. A `?` on every panel opens the relevant codex page (the in-game manual, written in-fiction as internal documentation).
- **First 10 minutes acceptance test:** a player who reads nothing must reach their first generator purchase in <60 s and their first "oh that's what this is" satire beat (the free Horse Armor) within ~10 min.

## 3. The run-end screen (death screens)

Prestige is the narrative spine (Hades model): **each Exit is a story beat, not a reset menu.**

Run-end sequence:
1. **The split card** — final splits vs PB, golds highlighted, category, timer stopped.
2. **The obituary** — a generated company obituary in deadpan-corporate voice: what the company was, its peak, its sins (pulled from this run's actual flags: "Pioneered 40 dark patterns. Trust at death: 12. The pet did not attend the funeral."), styled per exit type (acquihire = LinkedIn farewell post; collapse = TechCrunch post-mortem; IPO = an S-1 cover).
3. **The Founder card** — what persists: Reputation gained, Network additions, Route Knowledge discovered, Clout/Soul deltas, founder age.
4. **The teaser** — one line of forward narrative that advances *regardless of success* (the Hades rule: the story moves every run): the rival's arc, the regulator's memo, the pet waiting at home.
5. **[ NEW ROUTE ]** — into the next run with carry-over summary.

Exit obituaries are a high-variance template surface (see §6): hundreds of fragments, selected against run facts, so no two adjacent Exits read the same.

## 4. Run-count pacing (the roguelike commitment)

**The first ending is structurally multi-run.** Not a soft expectation — a design law:

- Route Knowledge gates the endgame: reaching Depletion requires skips/categories/permanent unlocks that cannot all be acquired in one run (the way Hades' first win realistically takes ~10–30 attempts). Target: **first ending after ~6–12 Exits** across ~2 months of mixed play (tunable; the pacing research will sharpen it).
- Each Exit must advance something visible (Reputation tree, Network slot, a story cycle beat) — failure is content, never mere loss.
- **Post-first-ending replayability:** the other endings are gated harder — Ending C needs both A and B or Ethical%; Ethical% realistically needs MMO co-op (the commons) or celebrated cheese; category runs (Legal%, Net Zero%, Low%…) are the Heat/Pact-style post-win ladder; 100% (all achievements/categories/collections) is the months-scale completionist tier.
- This gives the completion pyramid: **days** → first Exit rhythm established; **weeks** → most systems seen (~80% of content); **~2 months** → first ending; **months+** → alternate endings, categories, 100%.

## 5. Copy surfaces inventory

The corpus, by surface (targets are launch minimums; all grow seasonally):

| Surface | Form | Launch target |
|---|---|---|
| News ticker / dispatches | one-liners, templated slots | ~600 lines (research banks already hold ~200 seeds) |
| Event text (Layer 1/2/3) | title + 2-sentence body + 2–4 options | ~150 events × variants |
| Run-end obituaries | template + fragment pools | ~300 fragments |
| Achievements | name + flavor line | ~600 |
| Tooltips (generators/upgrades) | 1–2 lines, curtain rule applies | ~800 |
| Hints/codex | contextual + manual pages | ~120 hints, ~60 codex pages |
| Lore cards (incl. the silent ones) | one card, one fact | ~80 |
| Pet/creature barks | short, sincere (never satirical) | ~150 |
| Season/GM dispatch stock | pre-written + live-written | rolling |
| UI microcopy (buttons, empty states, errors) | the era voices | full pass per era |

Every line carries: a stable string key, era/voice tag, variant group, and (for factual claims) a research citation or `verify` flag.

## 6. The copy production system

- **Strings are data**, in the same hot-reloadable content-pack format as balance/events (`12-content-pipeline.md`): keyed, tagged, templated (`{company}`, `{number}`, `{pet_name}`), with **variant pools + selection weights + no-repeat windows** (the barks architecture: priority + cooldown beats pure random — the anti-"ninth-time" defense).
- **Batch authoring workflow:** copy is produced in themed batches (per tier, per event pack, per season) against the flavor bible; each batch gets a consistency pass (voice rules checklist + citation check + the dark-content calibration rule) before merge. Claude drafts at volume; Marco is editorial cut. The research files' one-liner banks are seed stock, not shipping copy, until they pass this.
- **A copy linter** (cheap CI): key uniqueness, template-slot validity, banned-pattern list (no winking phrases like "wink", no unflagged real-person claims), length limits per surface, citation-tag presence on factual lines.
- Localization-shaped from day one (keys + interpolation, no concatenation) even though English-first — this costs nothing now and keeps the door open.

## 7. General UX principles

- **The timer is always visible; the splits are one tap away.** The speedrun frame is chrome, not a mode.
- **One screen, tabs, no navigation depth >2.** Idle games live in a glance.
- **Era chrome changes; layout skeleton doesn't** — muscle memory survives redecoration.
- **Numbers legible at all magnitudes** (notation setting; default: short scale words → scientific at 1e21).
- **Every parody surface carries its curtain** (tooltip or small print).
- **Sincerity islands:** the pet panel and the lore cards are visually calmer — the satire chrome recedes there by design.
- **Accessibility baseline:** reduced-motion honored everywhere (creatures, particles, timer blink), colorblind-safe state signaling, keyboard operability for all core loops, screen-reader labels on the primary economy (full bar set by the rendering/compliance research).
