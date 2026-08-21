# Coverage Map — the research → design → RFC → implementation tracker

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** Preserves the 2026-08-05 through
> 2026-08-10 coverage reconstruction. Do not use it to authorize current work. Read
> `planning/platform-alignment/CURRENT-STATE.md`, `decision-queue.md` and
> `execution-queue.md` for current status.

**Purpose.** A single queryable source of truth for *how far every distinct gameplay system has
travelled* through the four tiers, so the breadth-first foundation program can be driven to
completion and the RFC gaps (systems that are designed but have no contract yet) become an
ordered, dependency-aware backlog rather than a memory.

This exists because a 2026-08-05 status recap classified ~52 systems by a quick `grep`, not a real
validation. This directory replaces that with an **independently reconstructed and cross-checked**
map.

## The four stages a system travels

| Stage | Evidence required | Where it lives |
|---|---|---|
| **R — Researched** | A dossier that actually backs the mechanic (not just names it) | `design/research/*.md` |
| **D — Designed** | A design section that specifies the mechanic as intent | `design/*.md` |
| **F — Foundation** | An RFC contract (`draft`/`implementing`/`accepted`/`implemented`) | `rfc/*.md`, `rfc/archive/*.md` |
| **I — Implemented** | RFC archived + behavior in `docs/` | `rfc/archive/`, `docs/` |

A system's **furthest stage** is what the map records, plus the specific file evidence for each
stage it claims.

## Validation protocol (why this map is trustworthy)

Each domain slice is reconstructed by a subagent that reads the ACTUAL files and answers, per
system, with file+section citations — never from prior conversation. Every row must survive four
adversarial checks:

1. **R→D honesty:** does the design claim rest on research that actually says it? Flag design
   assertions with no research backing (`UNBACKED`), and research never carried into design
   (`ORPHAN`).
2. **D→F fidelity:** does the RFC actually specify what the design intends, or has it drifted?
   Flag divergence (`DRIFT`).
3. **Status truth:** does the recorded status match the RFC's own status line AND its planning
   log (accepted rulings, review verdicts)? Flag mislabels (`STATUS-OFF`).
4. **Gap identity:** every system at furthest-stage D with no RFC is a `GAP` — a required backlog
   entry, tagged with the foundation(s) it must build on.

Slice outputs land in `validated/<domain>.md`. The synthesized master table is `map.md`; the
ordered RFC backlog is `gap-backlog.md`. Both are regenerated from the slices, never hand-typed.

## Domains (validation slices)

1. **economy-progression** — numeric core, economy kernel, production/idle, save, prestige/exits,
   factions, meters, clout/achievements, leaderboards/epochs, compute credit, doctrine, soul,
   quarters/earnings, ideology toggles, relevance/tier-falloff.
2. **mmo-social** — commons, guilds, world layer, feed/dispatch/presence, trading/market,
   server story arcs (Helldivers layer-3).
3. **minigames-combat-pets** — minigame platform, combat (data/duel/lane/bots), pet care,
   pet battles, the specific minigames (chess/board/garden/stock-market/grimoire/pantheon),
   base building & raids.
4. **collection-monetization-satire** — loot boxes/free gacha, hats/cosmetics, purchasable
   content, the dark-pattern parody suite.
5. **events-narrative** — events engine (3 layers), conspiracy system, media canonization/Clout
   milestones, speedrun framing (splits/PB/categories/TAS), the ending(s), the 9-tier content
   ladder, the Cookie Clicker / Bakery Inc. easter egg.
6. **infra-ui-tooling** — transport/fan-out, account/session bootstrap, API, CI, gameserver
   composition, copy pipeline, client shell, UI foundation, game-UI screens, deployment,
   balance harness.

## Status of this map

- 2026-08-05: scaffold created; 6-domain validation sweep completed; `map.md` and `gap-backlog.md`
  synthesized from `validated/*.md`; `rfc/README.md` index reconciled (11 missing RFCs added, stale
  "not yet drafted" para fixed); two RFC headers reconciled to filed rulings (minigame C36, pet
  C17). **Outstanding:** two design-body reconciliations (design/02 §6 Clout one-mint; meters
  v15/v16) are routed with a verification gate, NOT yet applied — see `map.md` findings.
