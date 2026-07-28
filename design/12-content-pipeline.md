# Content Pipeline & Extensibility

> The game is designed to grow for years: new events (historical AND ripped-from-the-news), minigames, tiers, factions, axes, seasons. This doc specifies the architecture that makes adding content a data problem, not a refactor — and the pipeline from "thing happened in the real world" to "it's in the game."

## 1. The principle: everything is a content pack

All game content lives in **content packs** — versioned directories of declarative data files (the formats already specified: balance data `02/06`, event schema `09`, strings `11`) plus optional code plugin points. The engine enumerates packs; packs never modify each other — they **append** (the CK3 on_actions lesson: append-don't-overwrite is what makes seasonal content additive and rollback-safe).

```
content/
  core/                  # base game: tiers, generators, upgrades, events, strings
  packs/
    season-01-going-concern/
    events-2026-current/   # ripped-from-the-news drops
    minigame-gomoku/
    faction-crypto/
  schema/                # JSON Schemas for every content type (CI-validated)
```

A pack manifest declares: id, version, dependencies (pack ids + engine version), content types included, and activation window (always / seasonal / GM-toggled). The GM dashboard **stages** packs; production activation ships as an **epoch-stamped deploy** (hot reload is dev-only — §7).

## 2. Content types and their extension points

| Type | Declarative? | Code needed? | Notes |
|---|---|---|---|
| Ticker lines / strings | ✅ fully | no | drop-in; the highest-velocity content |
| Events (all 3 layers) | ✅ fully | no | trigger DSL + effects vocabulary; new *effect kinds* need engine work (see §4) |
| Generators / upgrades / achievements | ✅ fully | no | balance data files |
| Pressure meters | ✅ fully | no | disaster schema |
| Situations / Major Orders | ✅ fully | no | GM-authorable |
| Lore cards / codex pages | ✅ fully | no | |
| Cosmetics (hats, palettes, effects) | ✅ mostly | rarely | sprite recolors + particle presets are data; novel render effects are engine work |
| Factions | ⚠️ partly | some | modifiers/inverted-upgrades are data; a genuinely new *verb* (the point of factions) usually needs an engine hook |
| Minigames | ⚠️ shell | yes | registered via a minigame interface (see §3); board/rules/AI are code + a data manifest (unlock tier, clock type, economy hook, rewards) |
| Tiers | ⚠️ partly | some | generator/upgrade/event/era-theme content is data; a new paradigm *grammar* (the point of tiers) is engine work behind a tier interface |
| Axes/currencies | ❌ | yes | new persistent axes (a hypothetical 5th scope) are engine changes → RFC required |

The honest rule: **flavor and balance are data; new verbs are code.** The architecture's job is to make the data column enormous and the code column well-socketed.

## 3. The minigame socket

Every minigame implements one interface (both sides):

- **Server:** `MatchActor` contract — init(config), handle(intent) → events, tick(), serialize/restore, and an `AIPolicy` slot (the bot). Registered under a minigame id.
- **Client:** a mount contract — a Svelte component receiving the match channel + a standard reward/exit surface; board rendering is the minigame's own business (DOM or the canvas layer).
- **Manifest (data):** unlock tier + cost, clock type, economy hooks (what it pays, what it consumes), matchmaking config (solo / async / live, bot backfill params), achievements.

Adding minigame #11 = one Go file + one Svelte file + one manifest. No engine surgery. (This socket is why the roadmap can promise "a minigame per season" with a straight face.)

## 4. The effects vocabulary (events' engine boundary)

The event DSL's power = its `effects` vocabulary (`modify`, `spend`, `trigger_event`, `grant`, `unlock`, `start_meter`, `dispatch`, …). This vocabulary is the **engine contract**: content packs may combine effects freely; new effect *kinds* are engine changes (small, but reviewed — each new verb is permanent API). Keep the vocabulary list in `docs/data-formats.md` once implemented; additions require at least a mini-RFC note.

## 5. The current-events pipeline (ripped-from-the-news)

The satire must be able to respond to the real world in days, safely:

1. **Intake:** a real-world event worth satirizing → a note in the season's planning log.
2. **Research check:** facts verified + sourced (added to the relevant research file — the citation rule holds for live content too).
3. **Author:** ticker lines / a Layer-1 event / a dispatch series, in a dated pack (`events-YYYY-current/`), passing the copy linter + voice checklist + **the legal guardrails** (compliance research → flavor-bible satire-risk rules; when in doubt, fictionalize harder).
4. **Ship:** pack push + hot reload — no client deploy for pure-data content.

   > **Resolved 2026-07-28: hot reload is dev-only.** Production content ships as **epoch-stamped deploys** through the full gate stack — every data change alters `constants_hash`, lands in the public changelog, and (when it touches balance) mints or amends a Balance Epoch. A silent nerf is structurally impossible; pure-data content still needs no *client* deploy, which was the original point. (`research/adaptive-balancing.md` Balance Epoch + `research/cicd-deploy.md`.)
5. **Age:** current-events packs get a review date; lines that aged badly are retired (packs make removal as easy as addition). The best ones graduate into `core/` as period pieces — the game accretes a history.

**Cadence guard:** current-events content is seasoning, not the meal — budgeted (a few drops/month max) so the game never becomes a news-cycle treadmill, and pack activation windows prevent stale "current" content from lingering.

## 6. Historical back-fill

The inverse pipeline: the research dossiers hold decades of material not yet shipped (the matrix's ~70% held-back banks). Historical packs are authored the same way, themed (e.g. `events-dotcom-bust/`, `events-2008/`), and slotted into the era they belong to — the tier ladder gives every era a natural home. This is the low-risk, high-volume content channel between seasons.

## 7. Versioning, saves, and compatibility

- Content packs are versioned; saves record the pack-set + versions they were written under.
- **Removal safety:** the engine must load a save referencing a retired pack (orphaned content degrades gracefully: owned items keep working with a `legacy` tag; missing event chains resolve/close cleanly). Rule: content can be retired, player property is never deleted.
- Balance changes to shipped data follow the balance-harness gate (pacing targets re-verified) + a changelog entry, per RFC-0000's docs rule.
- **Schema evolution:** content schemas are versioned like saves (migration chain); CI validates every pack against current schemas.

## 8. Modding (deliberate posture)

The architecture above is 90% of a modding API. Decision deferred (post-1.0): whether to open pack loading to players. Leaning yes-eventually (the nostalgia tier literally celebrates mod scenes; refusing would be off-thesis) — but MMO integrity requires mods stay client-cosmetic or private-instance only. Revisit as an RFC when the time comes; until then, nothing in the engine may *depend* on packs being trusted.

## 9. What this demands from the engine RFCs (checklist for drafting)

- Pack enumeration/manifest loading + dependency resolution + hot reload (scaffolding RFC).
- Append-only registration for events/on_actions/generators (production-engine + event-engine RFCs).
- The minigame socket (first minigame RFC establishes it).
- Effects-vocabulary registry with docs-page requirement (event-engine RFC).
- Save ↔ pack-version recording + graceful-orphan rules (save-layer RFC).
- Copy linter + schema validation in CI (scaffolding RFC).
