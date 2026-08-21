# Codex fix queue — 2026-07-30 (consolidated from the five package reviews)

> **FROZEN HISTORICAL QUEUE — NONCANONICAL.** Preserved through 2026-07-30 as review and
> remediation provenance. It assigns no present work. Current defects, decisions and authorized
> work live in `planning/platform-alignment/backlog.md`, `decision-queue.md` and
> `execution-queue.md`.

Full independent review of `bdcc9a1..2096916` is now complete and filed: verdicts + evidence live
in `planning/{account-and-session-bootstrap, prestige-and-exits, leaderboards-and-epochs,
websocket-transport-and-fanout, combat-shared-data}/log.md`. Every ruling referenced below is
normative in its RFC (changelogs 2026-07-30). Order is severity-then-dependency; each item names
its acceptance proof. HIGHs first — nothing new ships on top of a surface with an open HIGH.

## HIGH (fix before any new feature work)

1. **Session-family revocation race** — Account ruling **A-D2a**: `session_families` row, `FOR
   UPDATE` on every refresh/revocation, refresh rejects revoked family. Proof: the review's
   concurrent rotation-vs-revocation scenario as an integration test (replayed consumed token +
   simultaneous rotation → whole family dead).
2. **Import bricks the company stream** — Account ruling **A-D4a**: request hash/version select the
   migration catalog only; re-encode under current hash, `RunSeq := 1`, fresh `RunStartedAt`,
   standard store path. Proof: import a `run_seq=3` old-hash fixture, then round-trip an intent.
3. **Epoch mint strands in-flight runs at Exit** — Leaderboards ruling **L5b**: run N+1 starts
   under the server's CURRENT hash (D6 assembly from current catalog, revision + pin both). Proof:
   integration test minting epoch 2 with CHANGED bytes, then exiting an epoch-1 run.
4. **kernel/VERSION bump strands active runs** — Leaderboards ruling **L2b**: play-time pin check
   is hash-only; version drift → append-only `run_version_drift`, run playable, unrankable
   (`engine_mismatch` at verification, projection excludes). Proof: bump-version fixture keeps
   playing; board projection excludes it.
5. **Pre-v7 companies can never Exit** — Prestige ruling **P6a**: next save version backfills
   `RunStartedAt := EvaluatedThrough` where zero, flags run pre-timer. Proof: company v6→current
   corpus fixture (currently missing entirely) exits successfully.
6. **Transport queue disciplines are orphaned types** — wire `ConnectionQueue`/`History` semantics
   into the live publish path (drop-stale for gauge channels, 256-msg lossless bound for player
   channels) OR amend D2/docs to the byte-FIFO model deliberately; docs/transport.md currently
   claims wiring that does not exist. Proof: stalled-consumer test receives exactly one world
   snapshot on resume (AC3 live-path).
7. **Combat division gate is evadable** — recurse into subdirectories + fix strip order
   (strings before comments) + seeded self-tests for subdir, string-`//`, and template-literal
   cases. Proof: the review's three demonstrated escapes all fail the gate.

## MEDIUM

8. Constants-hash artifact-set authority (**L2a**): all composition sites derive from the epoch
   seed; parity test (seed set == harness == runtime). Currently three sites, three different sets.
9. Epoch seed sync at gameserver startup before readiness (**L5c**) — also unblocks composition.
10. Outbox: release the whole claimed remainder on any failure + per-founder ordering invariant +
    size guard at insert + bounded retries → dead-letter (transport findings 2–3).
11. `Drain`'s BroadcastDrain-error branch must still CloseForDrain/Shutdown (transport finding 4).
12. Argon2 rehash-on-verify with parameter floor + timing-oracle dummy verify (**A-D1a**).
13. Deletion preserves founder rows + `imported` via SET NULL + archive stamp (**A-D5a**).
14. Limiter: trusted-proxy config, bucket eviction, failed-auth IP bucket on authed routes.
15. Prestige: scripted-first via elective wind_down (**P5b**); span-collapse accumulator (**P6b**);
    TS mantissa/mantissa division + non-unit-mantissa vectors (**P2b**); ReputationDelta saturation
    (**P2c**); replay in recorded event order; gate tier = max(current, gate) never error;
    per-run decline scoping.
16. Mint livelock: validate changelog ref against the allocated id without burning retries
    (leaderboards finding 3); add `epochs`-table immutability trigger for closed epochs;
    `verified_runs` supersession contract (follow-up RFC line, not silent mutability).
17. Rejected-intent run_log growth cap (catalog value, oldest-rejected pruning).
18. Combat vectors: disadvantage+crit ordering, atk=100 scaling, advantage rounding; `bound()`
    golden vectors (both suites, ≥1 forced rejection).

## LOW (batch at will)

Typed 404/405 + `NewFounder` typed unknown_id; session/access-token GC job; recovery-code input
normalization + `Cache-Control: no-store`; drain-broadcast skip-history (rev-0 system message
currently enters player history and replays on recovery) + 503-after-broadcast order + unexport
`Node.Drain`; post-prune rejected-intent retry returns typed idempotency result; board `run_id`
format CHECK; DeleteAccount ordered stream archival (deadlock window); `floorRatio`/`idiv`
domain alignment; combat `Clamp` vectors when the first engine calls it.

## Standing (unchanged)

- `cmd/gameserver` composition per the 2026-07-30 resolver ruling (deny-closed guild/match) — now
  additionally gated on items 9 and 11.
- Event relay under the T3 `scope` ruling (company|founder, per-scope revs, scope wire vectors).
- Replay verification stays blocked on the immutable initial-run-state DESIGN-GAP (run-genesis
  snapshot contract is mine to draft).
- Repo is still local-only; every guard is advisory until pushed — Marco's action.

## Implementation status — 2026-07-30 Codex round

Completed and committed, without pushing:

- HIGH 1–2: `adafd60` serializes refresh-family authority and normalizes imported companies onto
  current hash/run 1 through the ordinary save path.
- HIGH 3–5: `072a60e` preserves old runs across real epoch/kernel changes and migrates pre-v7
  companies through save v8; `d7b4754` refreshes only the resulting save-state hashes.
- HIGH 6: `c87be53` wires message-count and drop-stale policies into the live Centrifuge writer.
- HIGH 7: `112680b` makes the combat division boundary recursive and tokenizer-based, with all three
  review escapes as mandatory self-attacks.
- MEDIUM 8–9 and the mint-livelock part of 16: `ac022cd` creates the single manifest authority,
  requires DB reconciliation before readiness, and allocates epoch IDs before mutation;
  `8a291ae` is the separately guarded identity-only harness refresh; `0e6cbe5` strict-validates both
  in-process manifests and identity-only baseline schemas.
- MEDIUM 10–11 plus the three adjacent transport LOWs: `fc97436` enforces per-Founder outbox heads,
  whole-remainder release, bounded poison dead-lettering/invariant reporting, insert+DB size guards,
  broadcast-first drain with shutdown on failure, and live-only rev-0 courtesy frames.

Verification evidence for this round:

- Full real-Postgres `make test-save-integration`: green after the epoch batch and again after the
  outbox migration; the latter includes the A1/A2/B1 ordering and five-attempt dead-letter fixture.
- Full `TEST_DATABASE_URL=… make verify`: green — all Go packages/vet/formulas/history guards,
  manifest-backed balance harness, strict TypeScript/Svelte/build, 6,441 Node tests (3 skipped),
  schema/boundary gates, and 19,332 browser tests.
- Post-review `make harness-check`, focused epoch/harness/leaderboard tests, and `go vet ./...`:
  green after `0e6cbe5`.

Still open from the original list: MEDIUM 12–15, the immutability/supersession/rejected-log parts of
16–18, and the listed LOWs not explicitly closed above. The event relay, composed binary,
Leaderboards replay genesis contract, and child combat engines remain forward-feature work rather
than remediation claimed by this round.

## Round-2 review — 2026-07-30 (remediation round `adafd60..af3a87b` reviewed)

**All 7 original HIGHs verified fixed** (evidence in the package logs: account, prestige/
leaderboards, transport, combat entries dated 2026-07-30). L2a/L5c/mint-fix approved with
evidence; A-D2a/A-D4a to the ruling; L5b/L2b/P6a to the ruling with the changed-bytes mint test.

**New fix queue from round 2 (ordered):**

1. **HIGH — combat gate template-literal blind spot**: flat scanner treats a template's closing
   backtick as an opener; everything after any `${expr}` template is unscanned —
   `arithmetic.ts:47` blinds shipped code today. Replace token loop with `ts.createSourceFile` +
   AST walk (also closes the `.tsx/.mts` gap and regex false positives); add
   division-after-closed-template to the seeded self-tests.
2. MEDIUM — CONSTANTS-IDENTITY must also require seed-artifact BYTES unchanged vs the previous
   baseline commit (else a semantic change to a hashed-but-unexecuted artifact ships as
   hotfix + identity refresh with no mint).
3. MEDIUM — ReconcileSeed bootstrap: empty DB must replay the FULL seed history (all epochs +
   accepted sets); today any post-mint or post-hotfix seed makes a fresh environment unbootable.
4. MEDIUM — 5k soak: assert monotonic world revisions (drop-stale-compatible), not exact
   sequence — currently fails deterministically under `-race`.
5. LOW batch: readiness-mid-drain flip; player-count decrement only on reserved frames;
   protobuf-client drop-stale fail-open; size-guard serialization alignment (jsonb::text);
   dead-letter transient/deterministic distinction + backoff; end-to-end store→relay→node test
   with A3-behind-dead-A2; golden-seed commits join the baseline subject walk.

**Still open from round 1:** MEDIUM 12–15 (Argon2 rehash+timing oracle, deletion SET NULL,
limiter proxy/eviction, prestige P5b/P6b/P2b/P2c/replay-order/gate-tier/decline-scope),
immutability/supersession/rejected-log parts of 16–17, combat vectors 18, and the LOW batch.

**New implementable RFCs (drafted 2026-07-30, in dependency order after the fix queue):**
`rfc/faction-incorporation.md` → `rfc/guild-model.md` (unblocks Commons Onboarding + the real
guild transport resolver) · `rfc/run-genesis-and-replay.md` (closes the replay DESIGN-GAP; enables
the verifier, board integrity, and `run_log_archive` compaction). Bounce blockers as usual.

## Round-3 review — 2026-07-30 (fix round `02cc096..e42248b` reviewed)

All verdicts filed in the package logs. Round-2 queue is closed except as noted. **Current queue,
ordered:**

1. **URGENT — `make verify-client` is RED at HEAD**: 02cc096's regex self-test string collapses
   `\/` to `/` and the gate fails its own self-test before scanning (verified first-hand; the
   commit's green claim was false — re-verify honestly and record it). One-character fix +
   `.js/.jsx` in the extension list.
2. MEDIUM — P5c: gate offer spawning on non-empty `exit_history` (offer path currently skips the
   scripted curriculum).
3. MEDIUM — P2d: TS `reputationDelta` exact BigInt product (reproduced ±1 vs Go); add the
   mismatch-point vector.
4. MEDIUM — epoch ADVANCE path must insert the declared accepted set (mint+hotfix before redeploy
   currently wedges the server unrecoverably).
5. LOW batch: 4000-mislabel for malformed frames + reservation validation-vs-overflow split;
   Argon2 stored-param ceiling; trusted-proxy deployment precondition in docs; session/token GC;
   epochs reopen-attack test; advantage floor-vs-round vector (bp=105); decline-count index.

Then the forward queue unchanged: faction-incorporation → guild-model → run-genesis-and-replay
drafts (bounce blockers as usual), event relay under T3 scope, `cmd/gameserver` composition.

## 2026-07-30 Codex round-3 implementation

- `e18cc6e` repairs the combat AST self-test's escaped regex and expands the recursive extension
  set to JS/JSX. The gate and focused client suite pass directly. A full `make verify-client` was
  attempted twice and reached the pinned-pnpm switch, then failed closed because the registry
  signature/download was unavailable; no green full-suite claim is made.
- `4638e23` implements P5c/P2d: first-run offers cannot spawn; subsequent-run offer acceptance is
  integration-tested; TS uses an exact BigInt product; the review's mismatch point is shared with
  Go. `0613925` installs the complete declared accepted set during epoch reconciliation/advance.
- `6e55f45` closes the implementable LOW batch: typed invalid-frame close, invalid reservation vs
  overflow, bounded stored Argon2 work, trusted-proxy origin precondition, explicit epoch-reopen
  attack, half-point combat floor vector, and the run-event lookup index. Session/token GC remains
  open because the composed long-running process owner does not exist yet; an orphan cleanup
  function is not falsely claimed as a job.
- `10de79d` closes a fresh T3 contract miss found during forward acceptance: event payloads now
  require `scope: company|founder` in both validators and the shared corpus.
- Forward RFCs were read in full and bounced with executable DESIGN-GAPs in their own files:
  Faction lacks ledger resources and accrual math; Guild inherits that and lacks exact clearing;
  Run Genesis lacks logged replay time/external inputs and an owned shared transition boundary.
- Verification: focused Go packages, `go vet`, schema validation, direct client Prestige/Combat/
  Transport suites, and full Compose/Postgres runs are green. No push performed.

## Round-4 review — 2026-07-30 (`e18cc6e..cb36be2` reviewed, diff-level)

**All six approved.** Gate re-run by the reviewer: GREEN at HEAD (`\\/` escape fixed, JS/JSX in the
extension list); P5c suppresses offer spawn at the source on empty exit_history (note: a
pre-existing save with a live offer + empty history could still accept — none exist, repo
unreleased, recorded only); P2d is exact BigInt in TS with matching truncation semantics and the
reproduced mismatch point as a vector; epoch advance inserts the DECLARED accepted set via the same
helper bootstrap uses (the mint+hotfix redeploy wedge is closed); T3 `scope` is exact-key enforced
in Go and TS with shared vectors; the LOW batch landed where the findings pointed (Argon2 ceiling,
reservation validation-vs-overflow split with its own close-code, reopen attack test, bp=105
rounding vector, decline-count index migration, proxy docs). `make verify-client` remains
externally blocked on the pnpm mirror — honestly recorded this time, and the gate itself was
verified green directly. Session/token GC correctly parked on the composed gameserver's job runner.

## Faction round reviewed — 2026-07-30 (9b5f4b3..91d089b): APPROVED, Guild unblocked

Verdict + evidence in planning/faction-incorporation/log.md. Queue before/alongside Guild:
1. **FB-1/P6c** — `catchup_ceiling_ms` into the prestige catalog (hash-pinned single source; one
   ServiceOption; construction-time equality check). Land BEFORE Guild's clearing (same boundary).
2. **F2a** — signatory incorporation continues membership (tithe max-raise, `compact_tithe_raised`).
3. LOW batch: `run_ended` gains `faction`; unknown-faction transition test + check-order pin;
   stock-writer closure assertion; composed-gameserver StatePolicyValidator assertion;
   next-run re-incorporation + incorporated-exit tests; pin accrual-hook chain order.
Then **Guild Model** (GA/GB/GC as amended), then Run Genesis (RA/RB first).

## Phase transition — 2026-08-03: the content program

The structural program is COMPLETE (every systems RFC implemented, reviewed, archived). The next
program, in order:

**Phase A — make it a game (critical path):**
1. `rfc/relevance-harness.md` (drafted) — implement FIRST so the first content epoch ships with
   relevance evidence.
2. `rfc/t0-t1-playable-content.md` (drafted today) — the first hour: T0–T1 catalog, first-session
   script, copy. Closes the oldest remaining contract.
3. Game-UI screens RFC (to draft next: the actual play surfaces on the implemented shell —
   tier HUD, buy panels, run-title speedrun bar, wind-down screen).

**Phase B — make it social:**
4. Commons Onboarding & Governance — UNBLOCKED (account/faction/guild/transport all exist);
   its six blockers get re-answered against the real contracts, then the client round.
5. Layer-1 events engine RFC (the Paradox-style evaluator design/09 §a specifies — the biggest
   unbuilt system remaining).

**Phase C — make it deep:**
6. Combat: C2 catalog content + duel engine + lane engine + bots (drafts exist with contracts).
7. First minigames beyond the arcade (Market with NEODAQ lessons, Velocity) + doctrine intents +
   Compute Credit spend.

**Ops, parallel, Marco-gated:** deployment RFC (cicd-deploy research banked) + THE PUSH — every
guard is advisory until the repo is public; recommended before Phase A content lands so content
epochs are born structural.

## Correction — 2026-08-03: the honest structural-debt list

"The structural program is complete" overclaimed. Accurate state, per the implementer's audit
(verified): CI's first hosted run is impossible until the push (Marco-gated); Combat Shared Data
carries three unfinished implementation items (C2 catalog, obedience/soul tables, boundary
gates); Prestige's [45,90] elective-Exit harness gate still runs on placeholder content (closes
with T0–T1); several archived systems have stale active planning state (housekeeping commit
welcome). The dispatch-integrity RFC is REJECTED — premise refuted against HEAD (see the RFC's
owner verdict); its D2 cardinality test is salvaged as test-only hardening. The Relevance
Harness draft gains V5 (six precisions) and the T0–T1 dependency; it IS indexed in rfc/README.md.
Archival-gate recommendation (any properly-labeled independent reviewer, full-span range union,
no permanent designated role) is RECORDED as the implementer's proposal — it matches practice
since the provenance rules landed, but the convention is Marco's to ratify, not ours.

## Phase-A order corrected — 2026-08-03

The implementer's audit was right and is accepted: Relevance/T0-T1 were drafted against
nonexistent mechanics. Corrected order: **1. Purchasable Content Foundation (drafted, awaiting
acceptance review) → 2. Relevance Harness → 3. T0–T1 content → 4. Game UI.** The T0–T1 draft's
false claims are corrected in place. Full foundation chosen over generators-only Relevance,
per the recommendation — the role law and §11b doctrine demand the fields exist before the
instrument measures them. Dispatch-hardening round (8f6885c..7f28677) acknowledged: cardinality
tests + honest withdrawal + Darwin approval, the salvage exactly as ruled.

## Visibility track — 2026-08-03 (owner priority: SEE progress)

Parallel to the Phase-A pipeline, effective immediately and FIRST in the queue:

1. **`make dev`** — one command: compose Postgres + migrate + `cmd/gameserver` + Vite dev
   server, seeded with a dev account (skip the recovery-code ceremony in dev mode via env
   flag, dev-only, loader-asserted absent in prod config). Everything exists; wire the command.
2. **Dev Shell** — Game-UI RFC U1 subset ONLY (desk: manual button, generator list with buy 1,
   balance header, run-title bar, minimal wind-down screen), explicitly ugly, against the
   placeholder catalog. U2's wire-only contract applies from day one (no throwaway
   architecture); U1's full polish stays with the Game-UI RFC proper.
3. From then on every round demos in the dev shell; the owner watches progress instead of
   reading about it.

Then the pipeline continues: Foundation catalog/save-v14 batch (unblocked today) → Relevance →
T0–T1 → full Game UI.

## SUPERSEDED + the Breadth-First Foundation Program — 2026-08-03 (owner ruling)

The visibility track (dev shell now) is WITHDRAWN by owner ruling: no UI on foundations that
will shift. The program is breadth-first: every foundational layer across the whole design,
THEN iterative content (games, tiers, minigames) on stable ground.

**Foundation inventory — done:** numeric core · economy kernel · save · production ·
gates/routes · commons · prestige · leaderboards+replay+verification · accounts · factions ·
guilds · transport (wire v2) · gameserver composition · balance harness + epochs + KV-1 ·
combat shared kernel (arithmetic/RNG).

**Foundation inventory — missing, the drafting/implementation queue:**
1. Purchasable Content Foundation — ACCEPTED, implementing (upgrades/chains/synergies/roles/
   ablation seam).
2. Relevance Harness — drafted (the measurement foundation).
3. Events Engine L1 — drafted (evaluator = the foundation for ALL narrative content; L2 meters
   + L3 server-events schemas join it as foundation scope, not content).
4. **UI Foundation** — drafted today: primitives, era-theming system, notation/formatting,
   copy-key resolution, wire-only component law + lint, navigation shell, a11y baseline —
   NO screens.
5. **API Foundation** — drafted today: versioning policy, generated schemas, the documented
   public read API (the fan-tooling stance made real), SDK boundary.
6. Minigame Platform Foundation — to draft: session/match lifecycle, the §12b scaling-seam as
   code, N-sends×cap faucet governor (the Neopets shape), reward payout hooks, registry —
   every minigame becomes content on this platform.
7. Achievements Foundation — to draft: the achievements-as-currency economy (design/02 §6) —
   closed achievement schema, burn/provenance/possession checks (the Neopets law), Clout feeds.
8. Clout mint/decay — to draft (design/02 §6 + the Gaia one-mint law).
9. Meters Foundation (Trust/Externality/Soul mechanics, design/02 §7–8; MeterBands exist as
   state only) — to draft; Events L2 consumes it.
10. Pet Care Foundation — to draft (cattery port: care stats, trust/mood, behavior FSM;
    the combat seam already awaits its two integers).
11. Combat engines (duel/lane/bots) — drafted, implementable.
12. World Layer Foundation — to draft: community milestones, depletion ratchet, the
    contribution/impact modifier (design/05 §a).
13. Feed/dispatch curation foundation — to draft (design/05 §2, amplification-gated).
14. Copy/content pipeline — to draft (copy-key system, era voice enforcement, provenance-tag
    lint for statistics).
15. Doctrine intents + Compute Credit spend — named contracts, to draft.
16. Deployment foundation — to draft from cicd-deploy research (Marco-gated push included).

Order within the program: continue current pipeline (1→2) while the drafting batch fills
(4, 5 today; 6–10 next; 11–16 behind). Content iteration begins only when the breadth is
foundationed — per owner ruling.

## Foundation batch drafted — 2026-08-03

Four foundation RFCs drafted (breadth-first, mechanics not content):
- **Meters Foundation** — the moral axis (Trust×5, Externality, Soul, p(doom)) with the
  not-spendable invariant enforced in code, band transitions as facts, attended-time decay.
  MeterBands state exists; this gives it laws.
- **Clout & Achievements Foundation** — achievements-as-currency (the design's stated
  highest-leverage mechanic, currently ZERO implementation) + Clout, with the Gaia one-mint law
  and the Neopets burn/provenance/possession check law structural from line one.
- **Minigame Platform Foundation** — session lifecycle + the §12b scaling seam as code + the
  Neopets faucet governor + payout-as-intent + the AI-fallback law; the two combat engines
  register as its first tenants.

Now on the board (drafted, awaiting Codex acceptance review, then implementation): UI Foundation,
API Foundation, Meters, Clout&Achievements, Minigame Platform, Events L1, Relevance Harness,
Commons Onboarding. Remaining to draft in the foundation program: Pet Care, World Layer, Feed
curation, Copy pipeline, doctrine intents, Compute Credit spend, deployment. Plus the F1
mint-precondition fix routed into T0–T1.

## Foundation batch 2 — 2026-08-03

UI Foundation C1–C6 and API Foundation C1–C10 RULED (both accepted; era table corrected to
1995/2000 per design/08, operation-registry single source, dereferenceable-because-public
verification endpoints, formula artifacts become epoch artifacts, the exact additive-only compat
algorithm). Three more foundations drafted:
- **Copy Pipeline Foundation** — the copy catalog + the two satire-safety design laws made
  MECHANICAL CI gates (no-🔴-names denylist lint; statistics require [V] provenance or fail).
  UF3's named dependency.
- **Pet Care Foundation** — the cattery port, deterministic + server-authoritative; CLOSES
  combat C5's fixture-only (trust_ppm, soul) boundary with real output.
- **World Layer Foundation** — milestones + the published Helldivers contribution formula +
  the Jevons depletion ratchet; gives GC2's zero-valued world-snapshot fields their sources;
  the layer that makes it an MMO.

Full board now: UI (accepted), API (accepted), Meters, Clout&Achievements, Minigame Platform,
Copy Pipeline, Pet Care, World Layer, Events L1, Relevance Harness, Commons Onboarding, T0–T1
(F1-gated) — all drafted. Remaining to draft: Feed curation, doctrine intents, Compute Credit
spend, deployment. The foundation breadth is nearly complete on paper.

## Foundation program COMPLETE ON PAPER — 2026-08-03

UI C7–C8 ruled (Vite-fixture+Vitest a11y, no Storybook; App.svelte→fixture-host migration owned
here; dep order Copy Pipeline → UI → {Game UI, T0–T1}; exported boundary client/src/ui/).
Final draft batch:
- **Feed & Dispatch** — the feed's source (rate-shaped dispatches) + the amplification-gated
  curation law (broadcast passes an editorial gate, vote-based amplification FORBIDDEN — the
  Neopets Beauty-Contest lesson).
- **Doctrine & Compute Credit** — the two deferred intent contracts: doctrine-pick (closes the
  routes RFC's ordering item) + Compute Credit spend (banked time gets its mechanic, F1-aware).
- **Deployment Foundation** — the deploy pipeline + THE PUSH, the one action only Marco can take
  that turns four epochs of guards from advisory into structural.

**The entire foundation breadth is now drafted.** Full board: UI (accepted C1–C8), API (accepted
C1–C10), Meters, Clout&Achievements, Minigame Platform, Copy Pipeline, Pet Care, World Layer,
Feed&Dispatch, Doctrine&ComputeCredit, Deployment, Events L1, Relevance Harness, Purchasable
Content Foundation (IMPLEMENTED), Commons Onboarding, T0–T1 (F1-gated), Game UI (deferred behind
UI Foundation). Nothing foundational remains undrafted. Content (tiers, minigames, pets, combat
content, endings) begins only when the breadth is implemented — per owner ruling.
