# Design Backlog — the Topic Ledger

Every idea gets a row the moment it's uttered. Statuses:
`💡 candidate` (mentioned, not designed) → `📐 designed` (has a home in a design doc) → `🔬 researched` (matrix row ✅) → `📜 RFC` (spec'd/implementing) → `⛔ rejected` (with reason, kept for the record).
Rule: an idea missing from this ledger is a process bug. Companion files: `research/README.md` (research status), `rfc/README.md` (spec status).

## Repository alignment and release defects (audit 2026-08-20)

Evidence: `planning/platform-alignment/reality-audit.md` and its tracked backlog, decision and
execution queues. These rows originated at `190a4fa`; status corrections below are reconciled
through 2026-08-21. They do not choose a release floor or authorize implementation.

| ID | Observation | Status / owner |
|---|---|---|
| RP-001 | Current-head hosted CI never reaches a verdict: the harness exceeded its 30-minute job budget in repeated current-product runs. | 🔬 native R-001 complete (974.510 s); one hosted observation still blocked on code-transfer authority -> D-014 |
| RP-002 | Production packaging is absent: no Dockerfile, Caddyfile, full-stack Compose contract, or clean-host boot proof. | 📜 Deployment draft, body stale; D-001/D-006 |
| RP-003 | Backup, restore, rollback, and disaster rehearsal are absent. | 💡 release defect -> R-006 / Deployment |
| RP-004 | Player data export has no endpoint, schema, UI, or proof. | 💡 release defect -> D-008 |
| RP-005 | Account deletion is backend-only; no player can invoke or understand it in the game UI. | 💡 integration defect -> D-009/R-003 |
| RP-006 | The one-time recovery credential is silently stored in localStorage with no display/download/recovery workflow. | 💡 integration/risk question -> D-005/R-003 |
| RP-007 | Fully offline-anonymous play is promised but absent; production startup creates a server account. | ⚠ intent/implementation contradiction -> D-004 |
| RP-008 | Game UI lacks accepted v3 Gate/Wind Down/next-run controls; AC1 still names the C28-rejected full browser replay. | 📜 Game UI AC1; blocked on author reconciliation then R-004 |
| RP-009 | Accessibility proof covers axe/reduced-motion/focus primitives, not complete keyboard/screen-reader/zoom/touch workflows. | 💡 release proof gap -> R-005 |
| RP-010 | Only T0–T1 of nine tiers and one real minigame exist; the designed ending and most MMO/narrative surfaces are absent. | ⚠ milestone scope question -> D-001/D-007 |
| RP-011 | Client third-party license delivery remains absent. | 💡 packaging defect -> Deployment |
| RP-012 | Sunset/self-host/export promises are research intent only. | ✓ D-003 ruled: MIT source now, supported bundle only after Deployment/R-006, covenant deferred |
| RP-013 | Repository disposition was contradictory: research matrix said private while GitHub and CI governance said public. | ✓ D-002 ruled; policy bodies reconciled and first reviewed public artifacts tracked at `a7320fa` |
| RP-014 | Current-state, README, Deployment RFC, RFC handoff, coverage map, and multiple plans disagree with HEAD. | 🔬 record defect -> platform-alignment reconciliation |
| RP-015 | Mandated shared research/backlog artifacts were gitignored and absent from a fresh clone. | ✓ D-002 complete; exact public/private authority manifest, negative controls and fresh-clone proof recorded at `publication-disposition-execution-08.md` |
| RP-016 | Minigame Platform claims a combat-duel tenant that runtime/catalog evidence disproves. | ⚠ normative contradiction -> ruling author |
| RP-017 | Soul recovery intent contradicts itself over restorative hobbies/Go/Arcade. | ⚠ owner-authored design contradiction |
| RP-018 | The cold-open specimen fabricates a WR despite the later no-unvalidated-WR ruling. | ⚠ owner-authored design contradiction |
| RP-019 | Minigame MatchActor/tick design contradicts the accepted pure-tenant/DB-session architecture. | ⚠ design/RFC drift |
| RP-020 | Production hot-reload prose contradicts the later epoch-stamped-deploy ruling. | ⚠ owner-ruling/body contradiction |
| RP-021 | Hosted harness capacity was claimed without a hosted run at the capacity commit; later hosted runs exceed the ceiling. | 🔬 acceptance defect -> R-001 |
| RP-022 | Harness workers cover standard pacing only; relevance work stays serial. | 🔬 performance defect -> R-001 |
| RP-023 | Harness kills exposed no progress or invalid-termination artifact. | ✓ Harness Observability archived; designated-approved at `96a574d` |
| RP-024 | Account family revocation and live-socket reauthentication lack one composed negative witness. | ⚠ acceptance gap |
| RP-025 | Unquoted root test-selector flags reject valid parenthesized Go regexes before launch. | ⚠ tooling defect |
| RP-026 | Removing cap/drain/resync rendering together leaves all 20,007 Game UI browser assertions green. | ⚠ acceptance-oracle defect |
| RP-027 | Permits P2/P4/AC4 contradict accepted PT-C1/PT-C2/PT-C4 rulings. | ⚠ owner-ruling/body contradiction |
| RP-028 | Permits lacks the required exact two-resource Go/TS gate-crossing replay row. | ⚠ acceptance proof gap |
| RP-029 | Canonical Economy/Routes and related docs still describe the pre-mint world. | ⚠ canonical-doc defect |
| RP-030 | Epoch-6 changelog does not cite every consumed verdict and exact reviewed range required by FCE AC4. | ⚠ closeout/provenance defect |
| RP-031 | FCE5.5/AC5 still say mint-commit green despite the accepted range-head ruling. | ⚠ owner-ruling/body contradiction |
| RP-032 | FCE's exact epoch-6 activation witness drifted into a deploy-current epoch-8 test. | ⚠ historical acceptance regression |
| RP-033 | Epoch-7 harness scope was grafted onto the completed epoch-6 mint RFC and later shipped elsewhere. | ⚠ RFC scope/lifecycle defect |
| RP-034 | Prestige payload/archive/scripted-trigger body and canonical docs contradict current contracts. | ⚠ owner-ruling/body and docs contradiction |
| RP-035 | Advisor Mode has state/math/labels but no player toggle contract or surface. | ⚠ design/integration gap -> owner ruling |
| RP-036 | Prestige still calls Quarter harvest a live offer site while its missing bridge is only inline-deferred. | ⚠ owner-ruling/body contradiction |
| RP-037 | Prestige AC2–AC6 lack their literal property/lifecycle/state-matrix/golden witnesses. | ⚠ acceptance proof gap |
| RP-038 | Prestige lacks a tracked explicit cross-party full-range archival verdict. | ⚠ archival review/provenance gap |
| RP-039 | Leaderboards join→leave→Exit incorrectly returns an assisted run to Solo. | ⚠ implementation/acceptance defect |
| RP-040 | Leaderboard storage has no production reader or player surface. | ⚠ producer-without-consumer defect |
| RP-041 | Leaderboards AC1/AC3/AC6 lack their literal end-to-end witnesses. | ⚠ acceptance proof gap |
| RP-042 | Leaderboards body/scope disagrees with current verdicts, categories, retention, and deferred surfaces. | ⚠ owner-ruling/body contradiction |
| RP-043 | Leaderboards plan/review provenance and Run Genesis remediation lifecycle are stale. | ⚠ lifecycle/archival defect |
| RP-044 | Minigame offline-quality decay is computed only before a new result overwrites it; no lapse-time output consumer exists. | ⚠ implementation/acceptance defect |
| RP-045 | Minigame async and bot-match acceptance is represented only by enums/schema/arithmetic, not composed workflows. | ⚠ acceptance/integration gap |
| RP-046 | Minigame RFC, plan, docs, and archived Pitch dependency contradict current platform state. | ⚠ normative/canonical/lifecycle contradiction |
| RP-047 | Minigame Platform lacks a final exact cross-RFC review-range union for archival. | ⚠ archival review/provenance gap |
| RP-048 | Browser access tokens expire after 15 minutes because the Game UI never uses its stored refresh token or reauthenticates the socket. | ⚠ session lifecycle/integration defect |
| RP-049 | Account AC2/AC3/AC5/AC6/AC7 lacked literal revocation/repeated/cross-system/all-stream/all-route witnesses. | ✓ Q-001 designated-approved at `34d04a5` |
| RP-050 | Account plan/review provenance predates current rewritten history and later composition successors. | ⚠ lifecycle/archival defect |
| RP-051 | Anonymous account/save storage has no retention quota or reaper beyond resettable IP limiting. | ⚠ operational abuse/retention defect |
| RP-052 | The Game UI did not persist transport positions, recover, reconnect, handle typed close codes, or honor drain delay. | ✓ Q-003 designated-approved at `249719c` |
| RP-053 | The tested per-scope revision cursor was not used by the production runtime. | ✓ Q-003 designated-approved at `249719c` |
| RP-054 | Transport AC3's exact-one-frame requirement failed under a literal slow client; AC6 lacked a non-member-Guild witness. | ✓ owner-reconciled AC3 and AC6 designated-approved at `249719c` |
| RP-055 | Transport body, plan, log and review ranges are stale against later composition/client work. | ⚠ lifecycle/archival defect |
| RP-056 | CI's four-job/sub-five-minute/harness-excluded body and “first hosted run” plan contradict current topology and runs. | ⚠ owner-ruling/body/lifecycle contradiction |
| RP-057 | CI caches Go build outputs and scheduled failures block despite dependency-only/non-blocking claims. | ⚠ workflow/canonical contradiction |
| RP-058 | CI lacks one exact current-history full-span designated review union. | ⚠ archival review/provenance gap |
| RP-059 | R-001 harness instrumentation lacked active accepted implementation authority. | ✓ owner accepted Harness Observability; implemented, reviewed and archived |
| RP-060 | API generation covers only 10 of 21 live v1 routes; unregistered handler drift passes. | ⚠ contract/acceptance defect |
| RP-061 | No production public API runtime/readers/evidence or historical formula artifact exists. | ⚠ producer-without-consumer/sequencing defect |
| RP-062 | API compatibility does not name the removed field and omits 11 live routes from pins. | ⚠ acceptance-oracle/coverage defect |
| RP-063 | API AC4 contradicts C9; raw runtime fetches remain and their lint is blind. | ⚠ owner-ruling/body/implementation contradiction |
| RP-064 | API registry/generation cannot yet express ruled query, raw-success, or platform-header contracts. | ⚠ accepted-contract implementation gap |
| RP-065 | API body/plan/docs/review union are stale or incomplete against current history. | ⚠ lifecycle/archival defect |
| RP-066 | Game UI performance never consumes its CPU/drop-frame fields or records the manual device check. | ⚠ acceptance-measurement defect |
| RP-067 | Game UI header/U2/AC1/docs contradict C7/C25–C28 and lack a final current review union. | ⚠ owner-ruling/body/lifecycle defect |
| RP-068 | Minigame API error authority permitted invalid pairs and its byte-match test accepted extra bytes. | ✓ corrected Q-002/API lane designated-approved at `bfd9b65` |
| RP-069 | Recovery heartbeat flooding lacked an authoritative-state nonmutation witness. | ✓ corrected Q-002 designated-approved at `bfd9b65` |
| RP-070 | Minigame privacy enumeration omitted Recovery request families. | ✓ corrected Q-002 designated-approved at `bfd9b65` |
| RP-071 | Minigame API body/status/plan drift; exact surface contract, HTTP client, components, and final review union are absent. | ⚠ owner-ruling/body/integration/lifecycle defect |
| RP-072 | Combat effect/spell unions and Trust/Obedience/Soul tables lack exact implementable contracts. | ⚠ owner decision/design specification gap |
| RP-073 | Combat plan/ledger and non-resolving review provenance disagree with the implemented kernel. | ⚠ lifecycle/archival-review defect |
| RP-074 | Combat bans native division across combat paths but enforces only the client half while Go divides natively. | ⚠ normative/acceptance-scope contradiction |
| RP-075 | Production cannot consume the declared deployment origin/proxy policy: WebSocket origin is pinned to localhost and Account trusted-proxy depth is hardcoded to zero. | ⚠ deployment configuration defect -> D-006 / Deployment |
| RP-076 | Gameserver cannot configure previous JWT/bootstrap or cursor keys despite underlying rotation primitives and DP5. | ⚠ deployment/security integration defect -> D-006 / Deployment |
| RP-077 | The built gameserver depends on repository files, embeds no catalogs/client, and serves no static client despite DP1's single-artifact claim. | ⚠ packaging/normative contradiction -> Deployment |
| RP-078 | The public push already happened, but the Deployment draft/index still presented it as the pending center of the RFC. | ✓ owner-delegated Deployment rewrite removes the obsolete premise; designated review pending |
| RP-079 | Missing/malformed browser credentials silently create a replacement account or strand the Desk offline; there is no recover-existing-account branch. | ⚠ account recovery/failure-state defect -> D-005/R-003 |
| RP-080 | Settings claims offline progress is parked locally despite no local save/intent queue/import owner/reconnect flush. | ⚠ player-copy/implementation contradiction |
| RP-081 | The destructive-settings navigation guard is test-only; Settings has no destructive account control or confirmation consumer. | ⚠ orphaned safety primitive/integration defect |
| RP-082 | Lifecycle surface replacement drops focus to `<body>` and does not announce/focus the new Offer/Run-End context in any browser engine. | ⚠ accessibility implementation/oracle defect -> R-005 |
| RP-083 | The Desk overflows to 647 px at a 320 px container in Chromium/Firefox/WebKit while the green suite fixes 1280×720. | ⚠ accessibility reflow defect -> R-005 |
| RP-084 | Production reduced motion does not reach numeric shell interpolation/pulse and does not react to mid-session OS preference changes. | ⚠ accessibility integration defect -> R-005 |
| RP-085 | Production has readiness/drain but no metrics export, request/access correlation, alerts/SLOs/dashboard/runbook; invariant metrics are composed nil. | ⚠ observability/operations integration defect -> D-006 |
| RP-086 | The package-tested intent-record pruner is never called in production despite the documented 30-day scheduler policy. | ⚠ retention implementation/canonical contradiction |
| RP-087 | Anonymous accounts, archived identities, immutable histories, dead letters, projections, and logs lack a complete retention schedule/disclosure. | ⚠ privacy/storage-abuse owner gap -> D-009 |
| RP-088 | Sunset research called self-hosting an existing near-zero-cost deliverable, but no production bundle/client/export/bots-default/artifact/mirror/runbook exists. | ✓ research contradiction repaired in fresh-clone-proven public brief; implementation gaps remain -> D-003/R-006 |
| RP-089 | Recovery decision D-005 and participant study R-003 were circular: each required the other's absent output. | ✅ process dependency fixed: posture -> prototype -> participant validation |
| RP-090 | Active RFC headers omit real Account/API/UI/telemetry/operations resource edges and use stale whole-RFC statuses as if they satisfied dependencies. | ⚠ dependency/authority graph defect -> ruling authors |
| RP-091 | Execution handoff still called blocked R-001 and completed lifecycle audit the next batch. | ✅ queue handoff reconciled to review then Q-001/Q-002/Q-003 |
| RP-092 | Review handoff overstated final-audit readiness while exhaustive inventory and capability splitting remained open. | ⚠ sequencing defect -> handoff downgraded; finish Waves 1–2 first |
| RP-093 | Inventory understated server packages/commands as 44/4; exact counts are 45/5. | ✅ corrected with package/route/migration row ledgers |
| RP-094 | Half of client source files are outside the shipped entry graph; parity/catalog/test modules are not mounted player workflows. | ⚠ capability classification -> exact source/workflow ledgers |
| RP-095 | Bundled Shell telemetry is recorded in memory but has no production reader/exporter. | ⚠ orphaned primitive -> D-016/R-009 |
| RP-096 | Three copy catalogs named candidate are nevertheless merged into the shipped artifact. | ⚠ content boundary -> explicit shipped/held manifest or adoption |
| RP-097 | Copy orphan report marks directly used Game UI keys orphaned because it ignores client call sites. | ⚠ false oracle -> complete reference extraction + seeded negatives |
| RP-098 | Thirteen ignored local browser PNGs were counted as test artifacts despite not being asserted/tracked. | ⚠ evidence classification -> only tracked sources/asserted baselines count |
| RP-099 | Two root boundary targets bypass the mandated repo-local Go cache with task-named `/tmp` caches. | ⚠ tooling/process contradiction -> root cache protocol |
| RP-100 | Seventy-two phony Make targets were treated as an evidence population although many are mutators/setup/manual/invalid aggregates. | ✅ exact target classification ledger; per-gate proof still required |
| RP-101 | Ordinary host Go tests can silently skip 39 Postgres test sites across 27 files. | ⚠ evidence denominator -> Docker/hosted lane + visible executed/skipped population |
| RP-102 | Four complete/withdrawn/superseded planning threads remain in the live top-level namespace. | ⚠ lifecycle/shared-memory defect -> verified transactional closeout |
| RP-103 | Binding player/World/Match concurrency prose contradicts the deployed no-player-actor, two-stage world runtime and deferred match services. | ⚠ design/runtime architecture contradiction -> ruling author + accepted successors |
| RP-104 | The 72-hour Route Registry naming expiry has an orphaned method but no worker or test caller. | ⚠ lifecycle/canonical defect -> accepted bounded scheduler + expiry witness |
| RP-105 | Fiscal is archived with seven open plan boxes and stale commit-under-rejection normative text that contradicts its accepted rollback correction and implementation. | ⚠ archive/body reconciliation -> ruling author + authorized record closeout |
| RP-106 | Copy's A1 fix was cross-party-closed only in the Meters log; the Copy owner's archive still ends `not approved`. | ⚠ transactional provenance defect -> append-only owner-record coordinate |
| RP-107 | `docs/soul.md` denies the Soul artifact and recovery activities that the current epoch and recovery doc prove live. | ⚠ canonical successor-closeout defect -> Soul/First-Content docs reconciliation |
| RP-108 | The Copy registry omits 50 exact generated-key references across current Pitch, Soul, Curriculum, and Minigames artifacts, so its orphan report conflates shipped data and mounted Curriculum copy with unused text. | ⚠ closed-world oracle defect -> accepted all-family reference/consumption successor with seeded failures |
| RP-109 | Current Fiscal and Opportunities artifacts use four hardcap reason keys absent from the application copy catalog. | ⚠ content/completeness defect -> owner-authored rows + strict current-artifact refusal gate |
| RP-110 | `upgrade.reply_all_macro` multiplies manual clicks but binding Tier-1 intent calls it automation. | ⚠ content/design contradiction -> D-007/content-owner reconciliation + accepted T0-T1 successor |

Owner choices D-001–D-017 and their evidence prerequisites live in
`planning/platform-alignment/decision-queue.md`; agents may not answer them through code.

## Core systems (designed + researched, awaiting RFCs in roadmap order)

| Topic | Status | Home |
|---|---|---|
| Numeric core | 📜 implemented (archived, + kernel/save/accrual/fast-path) | `06`, `rfc/archive/` |
| Tier ladder + endings | 📐🔬 | `01` |
| Economy layers (Company/Founder/World/Guild, Clout, Soul, Quarters, time banking) | 📐🔬 | `02` |
| Minigames catalog (10 games incl. board suite, garden, market, spellbook) | 📐🔬 | `03` |
| Pets (care/battles/house/lootboxes) | 📐🔬 | `04` |
| MMO (milestones, feed, guilds, commons, PvP) | 📐🔬 | `05` |
| Event engine (3 layers) + seasonal arcs + GM ops | 📐🔬 | `09` |
| Factions + engagement builds + challenge runs + moral routes | 📐🔬 | `10` |
| Doctrines (Age-Up branching runs) | 📐 (research: covered by events-playstyles + run-narrative) | `10 §3b` |
| UX/writing: contract screen, FTUE, run-end, copy system | 📐🔬 (run-narrative-ux banked) | `11` |
| Content pipeline / packs / current-events drops | 📐 (architecture settled; no research needed) | `12` |
| Client engine/rendering (Pixi v8 etc.) | 🔬 → RFC-ready | `06` + `research/browser-rendering.md` |
| Audio/music/art + genAI policy | 🔬 → RFC-ready | `research/audio-art.md` |
| Save layer + CI baseline | 📜 implemented (archived) · production engine + combat data model: 📜 drafted | `rfc/` |

## Research status

This table records coverage, not current execution. Work may begin only from the active RFC and
platform-alignment queues.

| Topic | Status | Notes |
|---|---|---|
| Pacing & completion science | 🔬 covered | balance-harness and later observation program |
| EU compliance & moderation | 🔬 covered, current review still required at use | accounts/UGC RFCs and satire-risk rules |
| World map / attraction / Plague-Inc spread | 📐🔬 | `13-world.md` + `research/map-attraction.md` |

## Candidates (💡 = not yet designed; do NOT RFC from here)

| Idea | Take | Earliest home |
|---|---|---|
| **Poker as a separate MMO track** | Good — the easiest true-MMO minigame (async tables, trivial bot backfill, ranked seasons; rule-based bot already specced). Promote from "casino suite line-item" to its own track with leaderboard + guild tournaments | `03` upgrade |
| **Spore-style creature evolution** | Good fit — pet breeding already exists (`04`); Spore's lesson: the *early game* minigames and weird-evolution charm carry the memory. Candidate: pet evolution stages with player-steered weirdness; tier-0 arcade could host Spore-like cell minigame | `04` + `03` |
| **XCOM-style invade/defend tactical minigame** | Plausible post-1.0, late-tier (transcendence/space era) — turn-based small-board tactics; AI fallback feasible (alpha-beta/MCTS on small state); heavy build cost — season content, not launch | `03` post-1.0 |
| **FTL-style ship management** | Same family — a space-era minigame candidate (the Orbital/Offshore doctrine could gate it); note FTL's brutal completion stats (26%) — keep it optional | `03` post-1.0 |
| **Lane pusher ("push to prod")** — 1 lane, core at each end, towers, auto-walking units | **📜 Shared combat implementing; lane engine draft.** Temperaments question resolved: shared data model, pet = On-Call Leader. One lane, deck/loadout as the real decision. A lane pusher has no offline victim; bots are materially more tractable than asynchronous raid attackers. | `03` + `04`; `rfc/combat-data-model.md` + `rfc/combat-lane-engine.md` |
| **Clash-of-Clans-style async base building** | ⛔ **REJECTED 2026-07-28 (owner): too much system for a solo dev, and it's one of ten minigames.** Superseded by the lane pusher above. The async-fairness analysis is being preserved in `lane-pusher-design.md` as a "why we rejected this" section — the reasoning is worth keeping even though the system isn't. | — |
| **The punch-down multiplier** (200% payout for punching up, 5% four tiers down) | 📐 **HOUSED** (`03 §10` matchmaking payouts). The single best anti-bully mechanism found in any game researched: pure incentive gradient, **zero moderation cost**. Its original home (async raid matchmaking) was rejected on 2026-07-28, so it currently has no owner. Proposed home: lane-pusher matchmaking payouts. Applies equally to pet-battle ranked and any future PvP. | `03`/`05`, → whichever PvP RFC lands first |
| **Modding / open pack loading** | Deliberately deferred decision | `12 §8` |
| **Discord embedded guild view** | Post-launch channel; no ambient chrome | `05 §8` |
| **Between-runs hub (walkable, NPC queues)** | 💡 post-1.0 — deliberately rejected for launch in the §6b adoption decision (`11 §3`) | `11` |
| **The Guild Break Room** (Habbo/Club-Penguin-alike, minimal) | 📐 **designed 2026-07-28** (`05 §3`) (`research/social-spaces.md` — verdict: build one room, not a world). One seat-based room per guild docked on the guild page: founder avatars idle while online, pets + labeled NPC co-ops as ambient population, emote wheel + curated phrase packs (**no free text — the chat ladder stops at rung 2; same curate-offline posture as the ticker**), cap 12–20 so fullness is reachable. It is the venue the `05 §3` weekly co-presence ritual currently lacks (`DESIGN-GAP:` filed) → recommended v1.0-minimal, falls to post-1.0 only if the ritual does. Distinct from the rejected walkable hub (that rejection stands): presence surface, not narrative hub, nobody walks. Never-list: furni economy (satire target only — `04 §4` already owns it), typed chat, room instancing | `05 §3` upgrade + `04 §1` bonds |

## Synthesized systems (📐 designed 2026-07-28 — ledgered post-hoc; each was a process bug caught by audit)

| Topic | Status | Home |
|---|---|---|
| Daily/weekly clock (activity bar, count-up streaks, mirror calendar) | 📐 | `02 §10` |
| Exchange shop (the only event-reward interface) | 📐 | `09 §5` |
| The cohort object (~150, server-assigned, nested Health 0.3) | 📐 | `05 §5` |
| Route Registry (player-named routes, house-seeded) | 📐 | `05 §6`, `08 §6` |

## Walkthrough holes (💡 from the 2026-07-28 player-experience review — the next design queue, ranked)

| # | Hole | Class | Blocks |
|---|---|---|---|
| 1 | ~~Route Knowledge / skip mechanics~~ | **📐 designed 2026-07-28** — `08 §6`: routes are registered alternate gate-preconditions (never production multipliers), executed fresh each run; Route Knowledge spends only on hints; Depletion counts distinct routes executed across the career, mutually exclusive by Doctrine construction | — |
| 2 | ~~Commons discovery surface + join flow~~ | **📐 designed 2026-07-28** — `05 §5`: the Mutual Aid Compact is a line-item in every incorporation contract; join changes exactly one production line; one commons (Open Source = heavier participation, `10 §1` resolved); canon signatories allowed, graded, and the satire beat | — |
| 3 | ~~`run-narrative-ux §6b` adoption~~ | **📐 resolved 2026-07-28** (`11 §3`) — scripted first failure ADOPTED (~15 min, full payout, in every category's route so it advantages nobody); Advisor Mode ADOPTED as honest accessibility (+2%/run cap +50%, `Assisted` variable); hub NPC queues DELIBERATELY REJECTED for launch (Founder card + teaser carry the narrative; hub is post-1.0) | — |
| 4 | **Daily-bar contents + unlock timing** — the frame exists (`02 §10`), the objectives list doesn't; D1 depends on it | UNSPECIFIED | daily clock ship |
| 5 | ~~Clout carry-over math~~ | **📐 resolved 2026-07-28** (`02 §6`) — Clout never enters the production stack; production effects key on this-run Clout; carried Clout buys reach (famous, not fast) | — |
| 6 | ~~Pre-PB return screen~~ | **📐 resolved 2026-07-28** (`11 §1`) — empty comparison column until a PB exists; offline-gains docked, not modal | — |
| 7 | ~~Exit trigger UX~~ | **📐 resolved 2026-07-28** (`02 §3.3`) — Exits are offers with full terms and real declines; Wind Down always available; IPO is the doctrine-gated planned Exit | — |
| 8 | **Doctrine coverage** — `10 §3b` tables 5 of 8 transitions (T0→1, T6→7, T7→8 missing) | UNSPECIFIED | `10` |
| 9 | **Personal objective board** — named as the milestone-wall anti-sag (`09 §4`), zero objectives enumerated | UNSPECIFIED | events content |
| 10 | ~~Return-moment modal composition~~ | **📐 resolved 2026-07-28** (`11 §1`) — diorama fast-forward is the stage, one modal max (ripe Quarter only), everything else docks or badges | — |

**Protect list** (the walkthrough's three strongest moments — do not regress): the contract screen (`11 §1`) · the day-2 ripe-Quarter gamble · the run-end obituary sequence (`11 §3`).

## Nostalgia arcade sweep (💡 batch — home: `03` Demo Disc Arcade, which evolves per era: cover disc → Flash portal → app store)

Legal rule for all: mechanics are fair game, **expression and names are ours** (Tetris Holding v. Xio — no trade dress, no -tris names; compliance fictionalization rule applies).

| Idea | Take | Era |
|---|---|---|
| **Solitaire/FreeCell + the Boss Key** | Office time-waster, Soul-restoring; Boss Key (fake spreadsheet on demand) is era UX + satire in one | T0–1 |
| **"Defrag" (match-3 as literal disk defrag)** | The best theme-fusion in the batch — matching C:\ blocks | T0–1 |
| Falling-block stacker ("Cold Aisle" — server crates) | Distinct expression only; Tetris Co. is litigious | T0–1 |
| **Missile-Command mode for Incident Response** | DDoS defense visualization — upgrade to existing minigame | T3 |
| **Dope Wars → "GPU Wars"** street-price arbitrage | The Market's nostalgic cousin; strong satire fit | T2–4 |
| SkiFree yeti event | The unavoidable late-run joke | any |
| Screensaver idle-visuals (toaster/pipes analogs) | Free ambience for away-state | T0–2 |
| Winamp-like skinnable music player UI | Ships with the audio system; llama energy | T1–2 |
| **Desktop Tower Defense** | Genuine gap — the active mode for base-raid defense (`04`) — *note: the `04` base-raid home was rejected 2026-07-28; slot needs a new home.* Now verified: HandDrawnGames, Armor Games, 2007-03-03; if built, steal Bubble Tanks TD's merge-to-grow (`flash-era-arcade.md §2.3`) | T2–3 |
| Insaniquarium-like fish toy | Cross-feeds pet layer | T2 |
| Peggle→plinko/pachinko | Belongs in the free-casino parody tier | T5 |
| 2048-lineage daily puzzle + Wordle-shaped daily slot | Rotating daily challenge in the arcade | T5 |
| **Flappy-Bird-story event** (tenant deletes their hit at peak) | Lore card; fits the thesis perfectly | T3–5 |
| Cultural artifacts: dial-up audio, guestbook, away messages, download bars, top-8, "CloudTV" | Analogue-only (MTV rule); flavor bible material | per era |
| 💡 **Turn-based artillery ("Severance Negotiations")** — Territory War/Worms shape | **Top pick of the Flash sweep**: natively async PvP (a turn is a message), trivial *honest* AI (printed aim-variance dial), parabola physics. The arcade's PvP anchor (`flash-era-arcade.md §2.3`) | T2–3 |
| 💡 **Day-allocation defense ("The Night Shift")** — The Last Stand shape | Allocate the workday (repair / hunt exploits / recruit), night resolves deterministically — **structurally an idle turn with a watchable spectacle**; PvE, no fallback needed | T2–3 |
| 💡 **Launch loop ("Runway")** — Learn to Fly/Burrito Bison shape | Cheapest build in the sweep and the best theme fusion (runway is already the metaphor); **the launch loop is the idle genre's direct ancestor** — final upgrade turns the cabinet idle as the punchline | T2–3 |
| 💡 **`PEON` top-down arena shooter** — Pawn/Stick Arena shape (the owner's PawnGame, verified: pawngame.com, Westech Media 2006) | Bot-backfilled deathmatch with era-ghost bots; guildmates can join live. Plus the **weekly "Arcade Night" appointment** pattern stolen from Pawn's own community revival | T2–3 |
| 💡 **Foreman round game** — Transformice-shaman shape | One player/bot builds the path, the rest run it; sabotage is scored, not moderated. The era's one genuinely novel co-op mechanic | T2–3 |
| 💡 **`Corridor 30` ghost race** — Exit Path shape | 30-second gauntlet vs. recorded guild ghosts — async "multiplayer" with zero liveness dependency | T2–3 |
| 💡 **Extraction dig ("The Core Sample")** — Motherload shape | Dig→fuel-leash→upgrade→deeper while the resource exhausts beneath you — **our thesis as a minigame**; ending is a contract in the last vein. Fits the Depletion era better than T2 | T4+ |
| 💡 **Wave arena ("Inbox Zero")** — Boxhead shape | Cheapest action slot (grid arena, 4/8-dir aim, combo-unlocked weapon ladder); local 2-player is period-authentic and dodges netcode | T2–3 |
| 💡 **Duck-Life-style trainee loop** | Train stats *via skill minigames* then race with live input — the shipped 2009 instance of `creature-battler.md`'s "care gates options, play decides outcomes"; feeds `04` pet-training thinking | T2–3 |
| 💡 **Tower-tycoon miniature ("Corporation Inc." lineage)** | The portals satirised tech companies first — a software-company tower toy whose penny numbers run inside our billion-dollar dashboard; the scale joke is the point | T2–3 |
| 💡 **Portal chrome layer** (preloader-ad joke, sponsor bumpers for in-fiction companies, site-lock error, blam/rating widget, badges, quality toggle, the portal's death played in chrome) | Not a game — **the T2–3 arcade's identity**; near-free satire, 12 usages specced in `flash-era-arcade.md §3.3`. Portal name = owner decision (§3.2: `WastedLunch.com` recommended) | T2–3 |
| 💡 **Upgrade Complete / Achievement Unlocked homage** | Buy the arcade's own UI with arcade tokens; a cabinet whose achievement list *is* the game — the era roasting our thesis before we did; highest satire-per-hour | T2–3 |
| 💡 **Lore cards: the portal economy** (auction sale cards à la FlashGameLicense, the ad-network collapse ticker, Bloons-escapes-the-portal, Crush-the-Castle→mobile-billions, the multiplayer preservation hole) | Flavor bible material; every card shape verified or flagged in `flash-era-arcade.md §7` | T2–5 |

## Viral / absorption-game sweep (💡 batch — the "you grow by consuming" family, thematically our core loop)

**The insight:** the absorption genre *is* our thesis in miniature — you get bigger by eating what's smaller. Two branches:

**A. Absorption mechanics (growth by consumption)**

| Idea | Take | Home |
|---|---|---|
| **agar.io → "The M&A Arena"** | **Standout.** Shared-arena cell-eating maps perfectly onto acquisitions: players/AI companies absorb smaller ones, split to grow faster, and get eaten by whales. Real-time MMO minigame with trivial AI fallback (bot cells) — the genre's netcode is well-documented. **Acquisition as literal consumption.** | `03` new minigame, v1.0 |
| **Katamari-style rollup** | The absorption-scale joke made joyful: roll up the office → the campus → the city → the planet. Silhouette-scale progression we already use in the diorama. Late-tier ceremony or a Depletion-tier minigame | `03`/`13` post-1.0 |
| **Osmos / flOw** (the elegant ancestors; flOw literally preceded Spore's cell stage) | Propulsion costs mass — a beautiful "growth has a cost" mechanic. Cell-stage toy for the arcade | `03` |
| Hole.io / Donut County | Modern Katamari; same family, simpler | `03` |
| Snake/slither lineage | Already covered via arcade; slither = the multiplayer variant of the same family | `03` |

**B. Went-viral games (cultural artifacts worth referencing or lifting)**

| Idea | Take | Home |
|---|---|---|
| **Progress Quest** (the zero-player RPG parody, 2002) | **The ur-satire of our own genre.** Perfect as a second tenant easter egg beside Bakery Inc — a game that plays itself, hosted on our racks, generating revenue while mocking us | `01` T3 / `03` |
| **"Spend Bill Gates' Money"-style spender** (neal.fun lineage) | The billionaire tier's joke as a real minigame: spend your Personal Wealth on absurd itemized things and watch the number not move | `04`/`13` billionaire tier |
| **Among Us-style social deduction** | Genuinely good guild content: industrial espionage — a mole in your guild leaking to regulators. Human-human strong; AI fallback is weak (deduction bots are poor) → the one place "human-only is acceptable" | `05` post-1.0 |
| Subway-Surfers split-screen | Not a minigame — a **satirical UI event** at the brain-rot tier (Stimulation Clicker lineage, already in `08`) | `08` |
| Desert Bus | The anti-game joke: a "mandatory compliance training" minigame that is deliberately, honestly nothing — with an achievement for finishing | `03` gag |
| QWOP/Getting Over It | Deliberate-frustration lineage; a one-off gag control scheme, not a system | `03` gag |

## Legal notes on lifted mechanics (house rules)

- **Mechanics are not copyrightable; expression is** (*Tetris Holding v. Xio*, D.N.J. 2012 — rules unprotected, distinctive look/feel protected). So: lift mechanics freely, **never lift trade dress, names, or distinctive visual identity.**
- **2048: yes, safe — the safest in the batch.** Its own implementation (Cirulli, 2014) is **MIT-licensed**, so even the code is reusable with attribution; the sliding-merge mechanic descends from 1024 → Threes! and was never successfully litigated. We still ship our own name + visuals per house rule.
- **Tetris is the cautionary opposite** — mechanics fine, but Tetris Holding actively defends trade dress and `-tris` names. Our stacker is "Cold Aisle," visually distinct, always.
- **agar.io/slither.io**: mechanics are unprotected and widely cloned; the names are trademarked. Ours is "The M&A Arena."
- Every lifted mechanic gets a row in `assets/MANIFEST.json`-adjacent design notes stating what was taken (mechanic) and what was not (expression).

## Rejected (kept for the record)

| Idea | Reason |
|---|---|
| Real-money anything | Vision law #1 |
| genAI shipped assets (even diegetic) | `research/audio-art.md §3` — the policy is binary |
| SharedArrayBuffer / WASM sim / OffscreenCanvas day-1 | `research/browser-rendering.md` |
| Nakama as the starting platform | `06` |
| Softcaps | Paper Pilot / `02` |
