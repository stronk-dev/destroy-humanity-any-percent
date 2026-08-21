# D-002 Publication Sensitivity Audit

**Audit date:** 2026-08-21  
**Authority:** owner-ratified D-002 posture in `owner-ruling-packet.md`  
**Scope:** ignored files under `design/` and `planning/` that were presented as durable shared
memory, not ordinary build/cache output  
**Effect:** classification and routing only. Nothing in this audit publishes, force-adds, deletes,
or rewrites an ignored artifact.

## Verdict

The current ignore boundary is not an acceptable durable-memory model for a public repository, but
blindly tracking the ignored population would also be unsafe.

- There are **96 ignored artifacts** in scope.
- **89 are durable prose/data**, not generated output: 70 research files, `design/BACKLOG.md`, 17
  coverage-map files, and one consolidated historical fix queue.
- **Seven are generated diagnostic JSON files**. They are not canonical memory and should remain
  ignored.
- The high-confidence credential/path scan found **no private key, GitHub token, OpenAI-style key,
  or email-address match**. This is evidence against obvious secrets, not a guarantee that prose is
  publication-safe.
- One absolute workstation path is present in `design/research/cattery-reusables.md:3`.
- The sibling-repository research exposes more than that path: it names a live hostname, ports,
  service topology, filenames, endpoints, authentication gaps, and deployment mechanics. Treat it
  as private-source-derived material until the owner confirms the sibling repository itself is
  public or adopts a sanitized synthesis.
- The research corpus is rich in attributed quotations, allegations, court/regulatory events,
  trademark/copyright analysis, named people and companies, and sometimes explicitly
  model-derived or verification-pending claims. A secret scan cannot authorize its publication.

The industry-standard boundary for this repository is:

1. track durable project control-plane records;
2. keep reproducible generated diagnostics ignored;
3. publish research as source-linked synthesis with short necessary quotations and explicit
   verification status;
4. exclude private sibling implementation details, credentials, personal paths, and unneeded
   operational identifiers;
5. label historical snapshots as historical so they cannot become a second current authority.

## Reproducible population

The audited population is the sorted output of:

```sh
git ls-files --others --ignored --exclude-standard design planning
```

| Class | Count | Disposition |
|---|---:|---|
| A — durable control-plane candidates | 13 | Track after the stale disposition text is author-reconciled and each snapshot receives a current/historical authority label. |
| B — requires targeted sanitization or adoption | 9 | Do not track in current form. Apply the named correction or obtain the named owner ruling first. |
| C — research publication-rights pass | 67 | File-by-file source, quotation, allegation, privacy, and verification pass before tracking. |
| D — generated diagnostics | 7 | Keep ignored; retain only canonical summary evidence in tracked logs/reports. |
| **Total** | **96** | |

### Class A — durable control-plane candidates (13)

- `design/BACKLOG.md`
- `planning/coverage-map/README.md`
- `planning/coverage-map/deferred-and-dropped.md`
- `planning/coverage-map/fce-mint-punchlist.md`
- `planning/coverage-map/gap-backlog.md`
- `planning/coverage-map/map.md`
- `planning/coverage-map/research-integration.md`
- all six files under `planning/coverage-map/validated/`

No credential or machine-path hit was found in this class. These are useful repository memory, but
the coverage map is a 2026-08-05 reconstruction and contains old lifecycle/status claims. Before it
is tracked, its README and master outputs must say whether they are historical evidence or a
current authority; merely exposing stale rows would worsen repository truth.

**Progress:** all 13 received a complete control-plane review. `design/BACKLOG.md` and
`deferred-and-dropped.md` are maintained-ledger candidates; the other 11 form one frozen
2026-08-05–10 historical coverage-map archive. No sensitivity refusal fired. See
`publication-control-plane-review.md`.

### Class B — targeted sanitization or adoption (9)

**Progress:** all nine received a complete targeted review. Two require ruling-author
reconciliation, two require public-safe synthesis or bounded rewrite, two belong in a frozen
historical archive, and three are byte-identical duplicates of canonical production data that
must not become a second authority. See `publication-targeted-artifacts-review.md`.

| Artifact | Required treatment |
|---|---|
| `design/research/README.md` | Its “one private repo, never publish research” body contradicts the 2026-08-21 owner ruling. The ruling author must reconcile the normative body before the ignore policy changes. |
| `planning/coverage-map/decisions-log.md` | Preserve the historical 2026-08-06/07 sequence, but append the 2026-08-21 superseding public-shared-memory ruling and clearly mark the former final disposition superseded. This is ruling-author work. |
| `planning/codex-fixes-2026-07-30.md` | Labels the repo local-only and acts as a large historical second queue. Either archive it with a prominent frozen-through date and links to canonical successor records, or prove it redundant and remove it through an explicit owner-approved cleanup. |
| `design/research/cattery-reusables.md` | Remove the absolute local path and do not publish sibling topology, live hostname, endpoint/authentication details, or file-level extraction without an explicit sibling-publication ruling. Prefer a source-neutral reusable-pattern synthesis. |
| `design/research/cicd-deploy.md` | Remove sibling-repository file/topology recipe details and machine-specific benchmark identity; retain only Cloud Clicker conclusions and public-source evidence. Revalidate its now-stale CI cost claims against R-001. |
| three `planning/coverage-map/draft-artifacts/*.json` files | These are unadopted first-content drafts, not neutral planning metadata. Track only if the owning content authority adopts them or retain them as explicitly noncanonical examples without player-facing authority. |
| `planning/coverage-map/mint-content-rows-proposal.md` | Contains proposed/owner-authored content rows and mint history. Track only with an explicit historical/noncanonical label and confirmation that all adopted copy already has its canonical home. |

### Class C — research publication-rights pass (67)

This is every `design/research/*.md` file except the three Class-B research files above. The class is
not a claim that every file is legally risky. It records that the corpus deliberately mixes:

- direct and paraphrased source material;
- quotations of varying length;
- company and named-person conduct;
- litigation, criminal, safety and abuse allegations;
- trademark, patent, copyright and platform-policy analysis;
- `[P]` partial and `[M]` model-synthesis claims;
- “verify before shipping” queues that were designed for player copy, not public-source
  publication.

The pass must be file-by-file and record: source URL/primary-source preference, quote necessity and
length, allegation attribution, current verification state, personal/private operational detail,
and the chosen correction. It may approve a file unchanged, shorten/attribute it, replace it with a
synthesis, or keep only a public index plus a named private source store. “No secret regex hit” is
not an approval criterion.

**Inventory correction (Batch 27):** the 67-file denominator is reproducible, but Batch 04 counted
already-tracked `provenance-extracts.md` while ignored `compliance-2026-refresh.md` appeared in no
manifest. Batch 04–26 progress snapshots therefore overstate reviewed and revision-blocked counts
by one and understate unreviewed by one. Their per-file verdicts remain valid; Batch 27 reviews the
omitted member and restores the current count. See `publication-rights-batch-27.md`.

**Progress:** Batch 01 reviewed `numeric-core.md`, `economy-kernel.md`, `browser-rendering.md`, and
`balance-enforcement.md` completely and found them eligible as dated technical research without
file edits. Batch 02 reviewed `tech-stack.md`, `mobile-pwa.md`, `tier-relevance.md`, and
`adaptive-balancing.md`: three require bounded revision and the raw Adaptive Balancing dossier
requires a public synthesis or named private-source ruling. Batch 03 reviewed four minigame
mechanics dossiers: `absorption-arena.md` and `board-game-mechanics.md` require bounded revision;
`lane-pusher-design.md` and `rhythm-timing-games.md` require public syntheses plus a raw-source
disposition. Batch 04 reviewed release/provenance/compliance/assets: `release-platform-audit.md` is
eligible as a dated snapshot; `provenance-extracts.md` needs bounded revision; `compliance.md` and
`audio-art.md` need public syntheses plus current authority review. **Sixteen of 67 are reviewed:
five eligible, six revision-blocked, five synthesis/private-store-blocked, and 51 unreviewed.**
Batch 05 reviewed four generational-culture notebooks; all require public synthesis, with the
Millennial pass first requiring the rerun its own header demands and the Gen Alpha raw dossier
requiring a specifically restricted disposition. **Twenty of 67 are reviewed: five eligible, six
revision-blocked, nine synthesis/private-store-blocked, and 47 unreviewed.** See
`publication-rights-batch-01.md` through `publication-rights-batch-05.md`. Batch 06 completed the
cohort family with `culture-boomer.md` and `culture-genz.md`; both require public synthesis. **Twenty-
two of 67 are reviewed: five eligible, six revision-blocked, 11 synthesis/private-store-blocked,
and 45 unreviewed.** See `publication-rights-batch-06.md`. Batch 07 reviewed the historical
completeness sweep plus the pacing and run-narrative foundations: the sweep needs bounded staleness
revision, while both raw foundations require public synthesis. **Twenty-five of 67 are reviewed:
five eligible, seven revision-blocked, 13 synthesis/private-store-blocked, and 42 unreviewed.** See
`publication-rights-batch-07.md`. Batch 08 reviewed the endgame and Soul mechanic foundations; both
require public synthesis, and the Soul synthesis also requires owner/safety review. **Twenty-seven
of 67 are reviewed: five eligible, seven revision-blocked, 15 synthesis/private-store-blocked, and
40 unreviewed.** See `publication-rights-batch-08.md`. Batch 09 reviewed the Gaia hyperinflation
and regulatory-capture case studies; both require public synthesis plus current claim/legal/
editorial review. **Twenty-nine of 67 are reviewed: five eligible, seven revision-blocked, 17
synthesis/private-store-blocked, and 38 unreviewed.** See `publication-rights-batch-09.md`. Batch
10 reviewed the paired Neopets economy and social/corporate dossiers; both require public synthesis
plus current claim/legal/editorial review. **Thirty-one of 67 are reviewed: five eligible, seven
revision-blocked, 19 synthesis/private-store-blocked, and 36 unreviewed.** See
`publication-rights-batch-10.md`. Batch 11 reviewed cozy-recovery and ARG-mechanics research; both
require public synthesis, with the ARG synthesis additionally requiring safety/legal/editorial
review. **Thirty-three of 67 are reviewed: five eligible, seven revision-blocked, 21 synthesis/
private-store-blocked, and 34 unreviewed.** See `publication-rights-batch-11.md`. Batch 12 reviewed
the designed-sunset dossier; it requires public synthesis plus current legal/policy review.
**Thirty-four of 67 are reviewed: five eligible, seven revision-blocked, 22 synthesis/private-
store-blocked, and 33 unreviewed.** See `publication-rights-batch-12.md`. Batch 13 reviewed the
launch/distribution dossier; it requires public synthesis plus current channel/owner review.
**Thirty-five of 67 are reviewed: five eligible, seven revision-blocked, 23 synthesis/private-
store-blocked, and 32 unreviewed.** See `publication-rights-batch-13.md`. Batch 14 reviewed the
Flash-era arcade dossier; it requires public synthesis plus IP/editorial review. **Thirty-six of 67
are reviewed: five eligible, seven revision-blocked, 24 synthesis/private-store-blocked, and 31
unreviewed.** See `publication-rights-batch-14.md`. Batch 15 reviewed the media-formats, nostalgia
and preservation dossier; it requires public synthesis plus current legal/IP/editorial review.
**Thirty-seven of 67 are reviewed: five eligible, seven revision-blocked, 25 synthesis/private-
store-blocked, and 30 unreviewed.** See `publication-rights-batch-15.md`. Batch 16 reviewed the
internet-platform, creator-economy and digital-culture dossier; it requires public synthesis plus
current legal/safety/editorial review. **Thirty-eight of 67 are reviewed: five eligible, seven
revision-blocked, 26 synthesis/private-store-blocked, and 29 unreviewed.** See
`publication-rights-batch-16.md`. Batch 17 reviewed the AI-authorship/provenance dossier; it
requires public synthesis, ruling-author body reconciliation and current legal/editorial review.
**Thirty-nine of 67 are reviewed: five eligible, seven revision-blocked, 27 synthesis/private-
store-blocked, and 28 unreviewed.** See `publication-rights-batch-17.md`. Batch 18 reviewed the
extreme-wealth and postwar-decay dossier; it requires public synthesis plus current legal/
political/editorial review. **Forty of 67 are reviewed: five eligible, seven revision-blocked, 28
synthesis/private-store-blocked, and 27 unreviewed.** See `publication-rights-batch-18.md`. Batch
19 reviewed the map-attraction, visible-progress and persistent-world dossier; it requires public
synthesis plus IP/editorial review. **Forty-one of 67 are reviewed: five eligible, seven revision-
blocked, 29 synthesis/private-store-blocked, and 26 unreviewed.** See
`publication-rights-batch-19.md`. Batch 20 reviewed the roguelike, survivor-like and deckbuilder
minigame dossier; it requires public synthesis plus IP/editorial review. **Forty-two of 67 are
reviewed: five eligible, seven revision-blocked, 30 synthesis/private-store-blocked, and 25
unreviewed.** See `publication-rights-batch-20.md`. Batch 21 reviewed the social-spaces and
constrained-communication dossier; it requires public synthesis plus current child-safety/legal/
IP/editorial review. **Forty-three of 67 are reviewed: five eligible, seven revision-blocked, 31
synthesis/private-store-blocked, and 24 unreviewed.** See `publication-rights-batch-21.md`.
Batch 22 reviewed the conspiracy-culture and media-canonization dossier; it requires public
synthesis plus current safety/political/legal/editorial review. **Forty-four of 67 are reviewed:
five eligible, seven revision-blocked, 32 synthesis/private-store-blocked, and 23 unreviewed.** See
`publication-rights-batch-22.md`. Batch 23 reviewed the societal-challenges and externalities
dossier; it requires public synthesis plus current political/legal/environmental/editorial review.
**Forty-five of 67 are reviewed: five eligible, seven revision-blocked, 33 synthesis/private-
store-blocked, and 22 unreviewed.** See `publication-rights-batch-23.md`. Batch 24 reviewed the
game-monetization and enshittification dossier; it requires public synthesis plus current consumer-
protection/legal/IP/editorial review. **Forty-six of 67 are reviewed: five eligible, seven revision-
blocked, 34 synthesis/private-store-blocked, and 21 unreviewed.** See
`publication-rights-batch-24.md`. Batch 25 reviewed the Kingdom of Loathing and Puzzle Pirates
systems dossier; it requires public synthesis plus IP/editorial review. **Forty-seven of 67 are
reviewed: five eligible, seven revision-blocked, 35 synthesis/private-store-blocked, and 20
unreviewed.** See `publication-rights-batch-25.md`. Batch 26 reviewed the persistent digital-
community atlas; it requires public synthesis plus current child-safety/legal/IP/editorial review.
**Forty-eight of 67 are reviewed: five eligible, seven revision-blocked, 36 synthesis/private-
store-blocked, and 19 unreviewed.** See `publication-rights-batch-26.md`. Per the Batch-27 inventory
correction, that provisional snapshot was actually 47 reviewed: five eligible, six revision-
blocked, 36 synthesis/private-store-blocked, and 20 unreviewed. Batch 27 reviewed the omitted 2026
compliance-refresh dossier; it requires public synthesis plus current legal/policy/editorial
review. **Forty-eight of 67 are now reviewed: five eligible, six revision-blocked, 37 synthesis/
private-store-blocked, and 19 unreviewed.** See `publication-rights-batch-27.md`. Batch 28 reviewed
the Warcraft III custom-game ecosystem and creator-rights dossier; it requires public synthesis
plus current legal/IP/editorial review. **Forty-nine of 67 are reviewed: five eligible, six
revision-blocked, 38 synthesis/private-store-blocked, and 18 unreviewed.** See
`publication-rights-batch-28.md`. Batch 29 reviewed the licensed-IP live-service idle-craft
dossier; it requires public synthesis plus current consumer-protection/IP/editorial review.
**Fifty of 67 are reviewed: five eligible, six revision-blocked, 39 synthesis/private-store-
blocked, and 17 unreviewed.** See `publication-rights-batch-29.md`. Batch 30 reviewed gamification
across finance, health, labor and civic life; it requires public synthesis plus current financial/
consumer-protection/labor/IP/editorial review. **Fifty-one of 67 are reviewed: five eligible, six
revision-blocked, 40 synthesis/private-store-blocked, and 16 unreviewed.** See
`publication-rights-batch-30.md`. Batch 31 reviewed the 1995–2005 period-satire and first-session-
copy dossier; it requires public synthesis plus IP/editorial review. **Fifty-two of 67 are reviewed:
five eligible, six revision-blocked, 41 synthesis/private-store-blocked, and 15 unreviewed.** See
`publication-rights-batch-31.md`. Batch 32 reviewed the cryptocurrency, Web3 and proof-of-work
satire dossier; it requires public synthesis plus current financial/legal/political/environmental/
IP/editorial review. **Fifty-three of 67 are reviewed: five eligible, six revision-blocked, 42
synthesis/private-store-blocked, and 14 unreviewed.** See `publication-rights-batch-32.md`.
Batch 33 reviewed the player-trading and market-architecture dossier; it requires public synthesis
plus current economy/security/legal/IP/editorial review. **Fifty-four of 67 are reviewed: five
eligible, six revision-blocked, 43 synthesis/private-store-blocked, and 13 unreviewed.** See
`publication-rights-batch-33.md`. Batch 34 reviewed the idle/incremental landscape and design-
synthesis dossier; it requires public synthesis plus current product/platform/IP/editorial review.
**Fifty-five of 67 are reviewed: five eligible, six revision-blocked, 44 synthesis/private-store-
blocked, and 12 unreviewed.** See `publication-rights-batch-34.md`. Batch 35 reviewed the healthy-
engagement and design-for-stopping dossier; it requires public synthesis plus current health/legal/
product/IP/editorial review. **Fifty-six of 67 are reviewed: five eligible, six revision-blocked,
45 synthesis/private-store-blocked, and 11 unreviewed.** See `publication-rights-batch-35.md`.
Batch 36 reviewed the onboarding and first-session-retention dossier; it requires public synthesis
plus current product/privacy/IP/editorial review. **Fifty-seven of 67 are reviewed: five eligible,
six revision-blocked, 46 synthesis/private-store-blocked, and 10 unreviewed.** See
`publication-rights-batch-36.md`. Batch 37 reviewed the dynamic-events and differentiated-
playstyles dossier; it requires public synthesis plus current product/IP/editorial review.
**Fifty-eight of 67 are reviewed: five eligible, six revision-blocked, 47 synthesis/private-store-
blocked, and nine unreviewed.** See `publication-rights-batch-37.md`. Batch 38 reviewed the labor-
organizing and worker-side-satire dossier; it requires public synthesis plus current labor/legal/
political/IP/editorial review. **Fifty-nine of 67 are reviewed: five eligible, six revision-
blocked, 48 synthesis/private-store-blocked, and eight unreviewed.** See
`publication-rights-batch-38.md`. Batch 39 reviewed the Cookie Clicker design-teardown dossier;
it requires public synthesis plus current product/IP/editorial review. **Sixty of 67 are reviewed:
five eligible, six revision-blocked, 49 synthesis/private-store-blocked, and seven unreviewed.**
See `publication-rights-batch-39.md`. Batch 40 reviewed the spectator and race-formats dossier;
it requires public synthesis plus current product/platform/legal/IP/editorial review. **Sixty-one
of 67 are reviewed: five eligible, six revision-blocked, 50 synthesis/private-store-blocked, and
six unreviewed.** See `publication-rights-batch-40.md`. Batch 41 reviewed the believable-
artificial-pet-personality dossier; it requires public synthesis plus primary-source product/IP/
patent/editorial review. **Sixty-two of 67 are reviewed: five eligible, six revision-blocked, 51
synthesis/private-store-blocked, and five unreviewed.** See `publication-rights-batch-41.md`.
Batch 42 reviewed the commons and cooperative-game-theory dossier; it requires public synthesis
plus current economics/governance/security/product/IP/editorial review. **Sixty-three of 67 are
reviewed: five eligible, six revision-blocked, 52 synthesis/private-store-blocked, and four
unreviewed.** See `publication-rights-batch-42.md`. Batch 43 reviewed the creature-battler, AI and
async-PvP dossier; it requires public synthesis plus current product/security/gambling/IP/patent/
legal/editorial review. **Sixty-four of 67 are reviewed: five eligible, six revision-blocked, 53
synthesis/private-store-blocked, and three unreviewed.** See `publication-rights-batch-43.md`.
Batch 44 reviewed the morality-systems and Ethical%-architecture dossier; it requires public
synthesis plus current product/philosophy/psychology/legal/political/IP/editorial review.
**Sixty-five of 67 are reviewed: five eligible, six revision-blocked, 54 synthesis/private-store-
blocked, and two unreviewed.** See `publication-rights-batch-44.md`.
Batch 45 reviewed the tile-placement, spatial-puzzle and shared-world-map dossier; it requires
public synthesis plus current product/economy/governance/security/environmental/legal/IP/editorial
review. **Sixty-six of 67 are reviewed: five eligible, six revision-blocked, 55 synthesis/private-
store-blocked, and one unreviewed.** See `publication-rights-batch-45.md`.
Batch 46 reviewed the speedrun-governance, verification and leaderboard-integrity dossier; it
requires public synthesis plus current product/security/privacy/moderation/community-safety/legal/
IP/editorial review. **All 67 are reviewed: five eligible, six revision-blocked, and 56 synthesis/
private-store-blocked.** See `publication-rights-batch-46.md`.

### Class D — generated diagnostics (7)

- six `planning/archive/t0-t1-content/*.diagnostic.json` files;
- `planning/platform-alignment/r001-registered.diagnostic.json`.

These are failed/intermediate measurement products. Their durable findings already belong in
tracked logs, verdicts and canonical evidence artifacts. Tracking every diagnostic would create
noise and accidental second authorities; the existing `planning/**/*.diagnostic.json` ignore rule
is correct.

## Exact blockers and next queue

1. **Policy reconciliation — complete:** the stale bodies in `design/research/README.md` and
   `planning/coverage-map/decisions-log.md` now implement the 2026-08-21 ruling while preserving the
   earlier historical sequence.
2. **Targeted treatment — complete:** the two public-safe derivatives are tracked by
   `publication-disposition-execution-01.md`; the two frozen historical records are preserved by
   `publication-disposition-execution-02.md`; the three redundant canonical duplicates correctly
   remain ignored.
3. **Control-plane tracking — complete:** the two maintained ledgers are reconciled and tracked by
   `publication-disposition-execution-03.md`; the 11 historical coverage-map records are tracked in
   their frozen archive.
4. **Research dispositions:** the five eligible dossiers are tracked unchanged by
   `publication-disposition-execution-04.md`; three of six bounded revisions are tracked by
   `publication-disposition-execution-05.md`; `publication-disposition-execution-06.md` tracks the
   revised absorption and board-game dossiers; `publication-disposition-execution-07.md` tracks
   the final frozen completeness sweep. All six bounded revisions are complete. The 56 synthesis/
   private-store raw inputs remain ignored/noncanonical behind their tracked per-file public
   records unless a later bounded derivative is justified.
5. **Fresh-clone gate:** fail if a tracked authority links to an absent ignored artifact without a
   named durable private-store contract, or if generated diagnostics are treated as canonical
   evidence.

Batch 17's four live ledger/colophon contradictions were reconciled on disk under the owner's direct
delegation. The raw dossier remains ignored, noncanonical and synthesis/private-store-blocked; the
correction does not approve automatic tracking.

No step above authorizes a push, publication, deployment, content adoption, or product change.
