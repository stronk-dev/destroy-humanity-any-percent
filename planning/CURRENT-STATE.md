# CURRENT STATE — read this after CLAUDE.md, before anything else

**Last updated: 2026-08-14.** Written so a fresh agent (either side) can resume with no
conversational context. `CLAUDE.md` has the process laws; `rfc/README.md` has the RFC index;
this file has **where the project actually is right now and what is blocking what.**

---

## 1. Where the project is

- **The repository is PUBLISHED.** `origin/main` exists
  (`git@github.com:stronk-dev/destroy-humanity-any-percent`). **No published commit may ever be
  rewritten** — the two historical rewrites (`d4c2312`, `bef1a87`) predate publication and that
  escape hatch is permanently closed. Published packaging defects get a forward correction
  commit + a ledger entry (`kernel/baseline-history-corrections.json`), never a rewrite.
  (Unpushed local history keeps the narrow sanctioned rewrite window in CLAUDE.md; it was used
  once on 2026-08-14, disclosed, and verdict-checked.)
- **Twelve-plus foundations are archived** (numeric core, economy kernel, save layer, production
  engine, harness, routes, commons, client shell, factions, guilds, doctrine, relevance, fiscal,
  active-play, soul, soul recovery, pitch, minigame platform's slices, meters, achievements, pet
  care, founder attendance, UI Foundation — see `rfc/README.md`'s archive table for the exact
  list and canonical docs).
- **Epoch 6 "First Content" is LIVE** — 16 artifacts minted, reviewed, and archived. The game has
  real content: meters, achievements, doctrines, minigames, pets, fiscal, soul, pitch.
- **The Phase-A screens exist and work.** A real Chromium drives a real gameserver against real
  Postgres over a live WebSocket: bootstrap → Vision Slide → Desk → intents → visitor counter.
  The copy is owner-ratified and pinned by hash.
- **Every implementation commit on `origin/main` is covered by a designated cross-party verdict**
  (coverage audit 2026-08-12). One pre-publication local debt is named in §4.

## 2. The critical path — what stands between here and a person playing the game

```
[HERE] C26/C27 instrument repairs (Codex, in flight)  →  branch-aware T0–T1 price tuple
→  owner signs exact literals + re-ratifies 4 hashes  →  epoch-7 mint  →  AC1 live script
→  Game-UI archival  →  Deployment  →  THE PUSH
```

**The 2026-08-12→14 instrument saga is RESOLVED.** Full history in
`planning/t0-t1-content/log.md` (start at the 2026-08-12 "RESUMABLE HANDOFF" entry). Compressed:

1. T01-C20/C21 landed: candidate score is closed-form projected time-to-milestone with banking
   first-class; T1 milestone re-derived by measurement to `1e9` (reached at 4,208,672 ms).
2. The beam-oracle round: the projected-time rework silently destroyed the beam's search power
   (F-A); a terminal-completion repair restored it; then the OWNER RULED the whole apparatus back:
   **cheap deterministic forced-deviation probes run always (`deviation.v1`); the beam is a
   manual diagnostic (`make relevance-beam`) in NO gate.** Attestations/identity-hash/staleness
   machinery was rejected unbuilt. T1's 5,000,000 transition ceiling stands and is no longer
   approached (T1 ≈ 2.8M with real gate crossing).
3. T01-C22/C23 landed: the relevance runner crosses tier gates through the REAL `cross_gate`
   engine transition (fail-loud, terminal gate excluded), and role verification moved to an
   11-row pinned candidate role matrix with masked controls (`make t0-t1-role-check`,
   in `verify-harness`). All role_floor and T1-window findings from before these fixes were test
   artifacts and are gone; `rack_rail_standardization` was never dead.
4. The C24 near-free price tuple `{900,8,10,10}` was measured, reproduced, and **REJECTED by the
   owner**: pricing upgrades near-free to satisfy a floor treats the symptom. The holistic
   diagnosis then found the real causes and the owner accepted them (see §3).

**Currently ruled and awaiting implementation (Codex, in flight):**
- **T01-C26** — transitive dependency-aware `instrument_affected` propagation: an item whose only
  non-neutral path intersects a screen-removal's dependency set cannot be presented as a balance
  finding. Pinned regressions: Dot-Matrix→Continuous Feed; Beige Tower v2→Refurbished Sticker.
- **T01-C27** — candidate-owned near-greedy branch fixtures (real engine, legally reachable
  prefix, T01-C20 decision, masked control, policy epsilon; row count derived from upgrades the
  main reference does not buy). One-path-only certification is rejected; so are synthesized
  states and trap exemptions.

**Then:** re-run the unchanged catalog on the corrected instrument → derive ONE branch-aware tuple
inside the accepted bounds (§3) → both branch rows green AND no whole-path pacing regression →
designated review → owner signs exact literals → re-ratify the four relevance candidate hashes →
content dispositions → epoch-7 mint runway (EH-C10 candidate report → promotion manifest over the
narrowed artifact set → owner sign-off → mint).

## 3. Ratification and ruled-balance state (owner-pinned — do not change these bytes)

**RATIFIED and pinned:** the Permits candidate trio; the epoch-6 content package (meters,
achievements, pets, categories, economy, fiscal, soul, copy, `minigame_api`); the T0–T1
nine-document candidate core; the Phase-A screen copy + presentation-v3 + event-copy-v2.
Exact hashes live in the RFCs' ratification sections.

**NOT ratified (and must not be):** the four relevance candidate hashes (T0/T1 scenario +
policy). They move again with the C26/C27 tuple round.

**Owner-accepted balance direction (2026-08-14, bounds not bytes):** both generators
(`answering_machine`, `nephew_intern`) are HEALTHY and unchanged — the earlier "weak generators"
reading inverted cause and effect. Price anchors: Reply-All 80, Hold Music 20k, Business Cards
20k, Continuous Feed 2k, CRT Degauss 50m, Handbook 75m, Refurbished Sticker 200m, Institutional
Memory 100m **retargeted to `generator.garage_rack`** (the sole accepted effect redesign). Final
literals are re-measured on the corrected instrument and return for explicit owner sign-off.
Local-delta-only evidence is insufficient: `{CRT=50m, Handbook=150m}` passes locally and regresses
the whole path — composition must be measured.

## 4. Open debts (all named, none forgotten)

- **Review debt:** the designated cross-party review of `{94487f9, 764e789}` (fail-loud +
  instrument exclusions, pre-publication). A spot-check found the F-items consistent; it still
  needs a real pass with probes. (In progress, Claude, 2026-08-14.)
- **Instrument follow-ups (LOW, recorded in verdicts):** H-A run-cardinality report field is
  tautological; H-B deviation probe outcomes (unreached/starved) undisclosed; shared
  gate-requirement predicate for `gateRequirementsMet`; cached completion returns hit-point
  `FinalState`; G-B beam-equality semantics fixed but `relevanceOracleOutcome`'s naming survives
  in the manual diagnostic only.
- **Carried acceptance debt:** Game-UI AC1 (epoch-7-mint-gated); MA's surfaces (MA-C9); Soul
  Recovery AC4/AC5; Minigame Platform AC6 and Pet Care AC3 (both gated on the still-draft Combat
  Duel Engine, whose own DU-C2–C8 are deliberately deferred until its parent catalog exists).
- **Registered follow-ups:** the eight discipline-sweep dossiers' verification pass; the
  dependency license audit before THE PUSH (in progress, Claude, 2026-08-14); the mobile/PWA
  provisions for `design/06` + UI Foundation; the Sunset Covenant RFC; the ARG surface trio; the
  AGI-tier meta-beat.
- **Owner rulings owed later:** the Offer Sheet's "nothing on this sheet is hidden" copy versus
  the withhold arm silently dropping an unbound row (before network-slot content ships).

## 5. Conventions established in-session that are NOT in CLAUDE.md's original text

CLAUDE.md's "Evidence discipline" section now carries rules 1–6; they are not repeated here. Two
additions that live only in this brief:

1. **Decision rounds go to Marco as `AskUserQuestion` dialogs WITH the context embedded in the
   question/option text** — prose written before a tool call is invisible to him. Plain language;
   no project jargon ("ceiling", "budget", "oracle") without a one-line translation.
2. **The instrument is a tool, not the project.** 2026-08-13's owner ruling exists because 48
   harness commits accumulated against 2 content commits. When verification cost starts compounding
   (tiers, attestations, identity hashes), stop and ask the owner what the check is FOR.

## 6. Working rhythm

Claude drafts RFCs and rules blockers; Codex implements and files blockers; **every batch gets a
designated cross-party review by the other side before archival** (Claude reviews Codex's
implementations; Codex reviews Claude's RFC/ruling commits). Marco rules design questions and
ratifies content hashes. Reviews are delegated to subagents with explicit probe instructions —
note the per-session subagent cap (`CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION`); it is cumulative,
not concurrent, and resuming an agent is free.

**THE PUSH is Marco's alone**, always.
