# CURRENT STATE — read this after CLAUDE.md, before anything else

**Last updated: 2026-08-12.** Written so a fresh agent (either side) can resume with no
conversational context. `CLAUDE.md` has the process laws; `rfc/README.md` has the RFC index;
this file has **where the project actually is right now and what is blocking what.**

---

## 1. Where the project is

- **The repository is PUBLISHED.** `origin/main` exists
  (`git@github.com:stronk-dev/destroy-humanity-any-percent`). **No published commit may ever be
  rewritten** — the two historical rewrites (`d4c2312`, `bef1a87`) predate publication and that
  escape hatch is permanently closed. Published packaging defects get a forward correction
  commit + a ledger entry (`kernel/baseline-history-corrections.json`), never a rewrite.
- **Twelve foundations are archived** (numeric core, economy kernel, save layer, production
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
  (verified by a coverage audit, 2026-08-12). Keep it that way.

## 2. The critical path — what stands between here and a person playing the game

```
[BLOCKED HERE] the relevance instrument  →  epoch-7 mint  →  AC1 live script  →  Game-UI archival  →  THE PUSH
```

**The one live blocker is the relevance measurement instrument**, not the content. Full history is
in `planning/t0-t1-content/log.md` (read from the 2026-08-12 "RESUMABLE HANDOFF" entry to the
end). Short version of how we got here — four rounds, each one honest:

1. The mandatory relevance gate **could not run** (a static estimator 30,556× looser than reality
   was gating on a work budget). Fixed: runtime counter, estimator demoted.
2. It ran and reported **perfection — which was a lie**: the beam ranked by cheapest-cost while
   greedy ranked by payback, so the oracle could never falsify greedy. A witness trajectory
   133,523 ppm better than greedy existed and the gate reported 0 ppm / passed.
3. Repaired, the oracle **fired** (120,879 ppm) — but that turned out to measure a *truncated
   search*: `max_decisions` (a runaway guard) was the binding parameter, and
   `finishToMilestone` silently coasted 48% of the run. Fixed: guard exhaustion now fails loud.
4. Un-starved, the T1 arm **hits the 24-hour horizon without reaching its milestone** — because
   the milestone is a *cash balance* target and greedy buys whenever anything is affordable, so
   it converts cash to generators forever and can never accumulate.

**Currently ruled and awaiting implementation:**
- **T01-C20** — the candidate score becomes closed-form projected time-to-milestone, with
  **banking (not buying) as a first-class candidate**: bank `= (T−B)/R`, buy `= (T−(B−c))/(R+Δr)`,
  minimum wins; ties by raw-byte item ID, bank loses ties; the beam uses the SAME metric.
- **T01-C21** — the T1 milestone (`1e9`) is unreachable in-horizon and must be re-derived by
  measurement. (It was mis-specified by Claude, twice in the same class as the original `1e12`.)

**Then, in order:** re-derive `max_decisions` non-starved → branch-B budgets (measurement-only;
never raise a ceiling from an incomplete run) → re-derive all four candidate hashes → designated
review → **owner ratification** → content dispositions → epoch-7 mint runway (EH-C10 candidate
report → promotion manifest over the NARROWED artifact set → owner sign-off → mint).

## 3. Ratification state (owner-pinned by SHA — do not change these bytes)

**RATIFIED and pinned:** the Permits candidate trio; the epoch-6 content package (meters,
achievements, pets, categories, economy, fiscal, soul, copy, `minigame_api`); the T0–T1
nine-document candidate core; the Phase-A screen copy + presentation-v3 + event-copy-v2.
Exact hashes live in the RFCs' ratification sections.

**NOT ratified (and must not be):** the four relevance candidate hashes (T0/T1 scenario +
policy). They will move again with T01-C20/C21. **No content retune may be justified by any
current relevance finding** — the instrument is under repair and every delta is suspect.

## 4. Open debts (all named, none forgotten)

- **Review debt:** the designated cross-party review of `{94487f9, 764e789}` (fail-loud +
  instrument exclusions). A direct spot-check found all four F-items consistent; it still needs a
  real pass with probes, the T1 horizon reproduction, and gates.
- **Carried acceptance debt:** Game-UI AC1 (epoch-7-mint-gated); MA's surfaces (MA-C9); Soul
  Recovery AC4/AC5; Minigame Platform AC6 and Pet Care AC3 (both gated on the still-draft Combat
  Duel Engine, whose own DU-C2–C8 are deliberately deferred until its parent catalog exists).
- **Registered follow-ups:** the eight discipline-sweep dossiers' verification pass (search budget
  was exhausted when they were written — `[M]` density is above house standard); the dependency
  license audit before THE PUSH; the mobile/PWA provisions for `design/06` + UI Foundation; the
  Sunset Covenant RFC; the ARG surface trio; the AGI-tier meta-beat.
- **Owner rulings owed later:** the Offer Sheet's "nothing on this sheet is hidden" copy versus
  the withhold arm silently dropping an unbound row (before network-slot content ships).

## 5. Conventions established in-session that are NOT in CLAUDE.md's original text

These were ruled after real failures. They bind both agents.

1. **A check that cannot fail is not a check.** Every gate, oracle, floor, and assertion must have
   a demonstrated failing case. Four separate defects reached review because an assertion passed
   on broken code (`toContain` on a prefix; a golden vector that couldn't discriminate; a browser
   test that exited 0 while throwing 18 uncaught errors; an oracle structurally unable to falsify
   its subject).
2. **Run it, don't read it.** The most valuable reviews bypassed a check and executed the thing.
   That habit caught the arm64/amd64 FMA divergence, the cache-masked outage, and the vacuous
   oracle. Prefer `-count=1`; warm caches have masked a red baseline for days.
3. **The ruling author reconciles their own stale text.** When a review finds an owner ruling's
   normative text stale or self-contradicting, the implementer FILES the finding and waits; the
   author edits. (Claude broke this rule once and was correctly rejected for it.)
4. **A measurement that quietly degrades is worse than no measurement** — it looks like data.
   Fail loud, and make instrument artifacts (exclusions, guard exhaustion, truncation) visible
   first-class fields, never silent.
5. **Budgets are set from measurement, never chosen.** Never raise a ceiling from an incomplete
   run. Never loosen an acceptance bound to make a gate pass.
6. **Owner-authored content is owner-authored.** Implementers may not edit ruled copy text — not
   even a word — without an explicit disposition; detector-forced rewrites still come back for
   adoption.
7. **Decision rounds go to Marco as `AskUserQuestion` dialogs WITH the context embedded in the
   question/option text** — prose written before a tool call is invisible to him.

## 6. Working rhythm

Claude drafts RFCs and rules blockers; Codex implements and files blockers; **every batch gets a
designated cross-party review by the other side before archival** (Claude reviews Codex's
implementations; Codex reviews Claude's RFC/ruling commits). Marco rules design questions and
ratifies content hashes. Reviews are delegated to subagents with explicit probe instructions —
note the per-session subagent cap (`CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION`); it is cumulative,
not concurrent, and resuming an agent is free.

**THE PUSH is Marco's alone**, always.
