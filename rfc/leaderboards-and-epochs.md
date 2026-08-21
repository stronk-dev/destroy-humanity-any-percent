# RFC: Leaderboards & Balance Epochs

- **Status:** implementing
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/05 §6` (derived boards, category model, Route Registry), `design/08 §6` (timer semantics, categories), `design/00` (world records as framing)
- **Depends on:** Production Engine (implemented), Save Layer (implemented — `constants_hash` per
  revision), Gate Predicates (implemented — `route_executed` events feed Glitchless/variables),
  account/session bootstrap, Prestige & Exits (run lifecycle)
- **Planning:** `planning/leaderboards-and-epochs/` (once implementing)

## Summary

Boards, epochs, and run verification — closing the **last entry in the deferred-decisions register** (ranking keys, from RFC-0002 draft D4). The architecture law: **boards are derived from replayed intent logs, never authored**; runs pin to the epoch live at their timer start; ranking keys are exact.

## Specification

### D1 — Ranking keys (the registered decision, resolved)

**Order keys are exact integers or exact times — never quantized `Decimal`s.** Time boards rank on integer milliseconds (RTA / Attended Time); count boards on exact integers. Where a magnitude must rank (a "largest bank" novelty board), the key is `(exponent, quantized_mantissa)` and **equal keys display as ties, sharing a rank** — never resolved arbitrarily.

### D2 — The run record

A run is `(founder_id, run_id, category, variables, epoch_id, constants_hash, seed, intent log)`. **Verification is replay**: the server re-simulates the intent log against the epoch's catalog and returns the closed six-verdict union: `verified`, or one of five machine failure causes (`log_gap`, `state_divergence`, `constants_mismatch`, `clock_violation`, `engine_mismatch`). The operational verification queue is machine-only; there is no video or human judgment queue. The TypeScript parity kernel is bundled, but no current API/archive-delivery/browser workflow lets a player invoke it; that delivery belongs to the draft Leaderboard Readers & Player Surface successor. Timer semantics per `design/08 §6`: RTA from `[BEGIN ATTEMPT]`; **Attended Time** = RTA minus offline spans (offline is published policy, not a rule dispute); IGT recorded, never ranked.

### D3 — Epochs

- An epoch is minted **deliberately**: `{epoch_id (monotonic), name, started_at, catalog constants_hash set, changelog_ref}`. Minting is a release act tied to a `BALANCE-CHANGE:` (Harness H3) and a public changelog entry — **a silent nerf is structurally impossible** because a balance change without an epoch fails the harness hook, and an epoch without a changelog fails this RFC's validation.
- **Hotfixes within an epoch** (correctness-only, `BALANCE-CHANGE:` absent) update the epoch's accepted-hash set; per-run `constants_hash` remains the forensic record of exactly what a run played under.
- **A run pins to the epoch live at its timer start, for its entire duration** (the ruled pin: at our cadence essentially all first-ending runs straddle a boundary; the most emotionally significant run in the game must land on a board).
- Boards key on `(category, variables, epoch_id, mandate_level)`. Old epochs' boards freeze and remain browsable forever ("patch X and earlier" is a first-class historical object, not an attic).

### D4 — Categories and variables

Phase 0 consumes five canonical epoch-owned category rows: Any%, 100%, Ethical%, Low%, and Valuation (L7a/L7b). Variables remain structural: `Glitched` (any `route_executed` this run), Commons membership at any point, Advisor, nullable Faction, and mandate level; Solo and Assisted are different board tuples, never a computed subtraction. Player-authored predicates, promotion, and Exhibition are deferred behind D-017 plus a later accepted contract. The combined Route Registry/records surface is designed intent but not part of this backend foundation; the draft Leaderboard Readers & Player Surface successor owns its API and player composition.

### D5 — Board Mandates

This foundation ships the exact bounded `mandate_level` board-key/storage plumbing only. No mandate catalog objects, opt-in intent, gameplay modifiers, or player consumer exist. Mandates 1–20 remain design/balance intent and require a later accepted gameplay/content RFC; their absence does not authorize constants or behavior in this foundation.

### D6 — World-first and broadcast hooks

The database atomically arbitrates and permanently dates the first verified completion per `(category, variables, epoch)`. No feed/dispatch event is currently emitted, and no distinct `machine` board class exists. Those consumers belong to the draft Leaderboard Readers & Player Surface successor; this foundation must not claim their delivery from the arbitration bit alone.

## Acceptance criteria

1. No board query path accepts a quantized `Decimal` as a sort key (type-enforced); a fixture with sub-quantum differences ranks as a shared-rank tie.
2. Replay verification: a tampered intent log fails with the correct machine cause; the bundled TypeScript parity kernel reproduces the Go authority's verdict on the same committed corpus. Player archive retrieval/invocation is successor acceptance.
3. Epoch pinning: a run started in epoch N and finished in N+1 ranks in N, replayed against N's catalog.
4. A balance change without an epoch mint fails the harness hook; an epoch without a changelog entry fails validation.
5. Solo/Assisted: joining the compact mid-run moves the run to Assisted **at verification, structurally** — no arithmetic adjustment path exists in code.
6. Frozen epoch boards remain queryable after two subsequent epochs.

## Open questions

- **G3:** threshold promotion (D4's 25-runs/10-founders)
  needs a flat, enumerable browse tail for unpromoted player categories —
  popularity-sorted-only discovery buries new content and calcifies the top
  list. Promotion thresholds gate RANKING, never VISIBILITY.

- Promotion thresholds (25/10) and mandate count: provisional, data.
- Board retention/pagination scale: implementation freedom.

## Executable contracts (answering the 2026-07-29 review)

### L1 — The run log (the missing table, owned here)

`run_log(run_id, seq bigint, intent_id, canonical_payload bytea, receipt jsonb, applied_revision, server_ts_ms)` — **written in the same transaction as intent commit** (the production engine gains one insert), `PRIMARY KEY (run_id, seq)`, seq strictly monotonic per run. `canonical_payload` is the exact canonical-JSON bytes the idempotency hash was computed over (so the log is self-verifying against `intent_records` hashes while those still exist). Verified runs compact into immutable `run_log_archive` objects transactionally. Abandoned-run cleanup and `run_ttl_days` are not implemented or owned by a current catalog; D-015's future accepted retention/operations contract must define them before deletion exists. Live run logs remain exempt from the 30-day `intent_records` prune; the policies are not silently conflated.

### L2 — Replay identity

A verified run pins `(constants_hash set, engine_version)`. **`engine_version` = the shared-kernel version constant** (semver, bumped by RFC whenever transition semantics change — a new source-of-truth `kernel/VERSION` read by both runtimes and embedded in builds) plus the build's VCS hash as forensic detail. Replay runs the **exact catalog bytes** fetched by hash from `catalog_artifacts(constants_hash, bytes, created_at)` — immutable, insert-only, populated at epoch mint/hotfix — under a kernel whose VERSION matches the run's. A version mismatch is verdict `engine_mismatch` (fifth machine cause, added to D2's four), never a silent replay-under-new-code.

**L2a — Artifact-set authority (review finding, 2026-07-30, MEDIUM):** the epoch seed
(`balance/epochs/phase0.json`) is the **single authority** for which artifacts compose a
constants hash. Today three sites hardcode divergent sets (seed: 4; `harness.go:210`: 3 without
prestige; prestige integration fixtures: 3 without commons) — self-consistent individually, but the
composed server would mint run hashes no epoch accepts. Contract: every composition site derives its
artifact list from the seed, and a parity test asserts seed set == every site. Acceptance: change
the seed's artifact list in a fixture and every composer follows without code edits.

### L3 — Timer lifecycle (binds to Prestige P6)

`[BEGIN ATTEMPT]` **is** the `run_started` event; run_id and seed come from Prestige P3 (founder seed ⊕ run_seq); RTA = `run_ended.server_ts − run_started.server_ts` (integer ms); Attended Time = RTA − Σ `offline_spans` (Prestige P6's server-derived spans; the client clock contributes nothing). Pause does not exist (an idle game has no pause; disconnection simply accrues an offline span). Terminalization = the Prestige Exit transaction; `run_ended` carries the terminal `run_log` seq (P7), and verification's completeness check is exactly `max(seq) == run_ended.terminal_seq`.

### L4 — Validator parity foundation; player delivery deferred

**The validator kernel is the existing TypeScript shared kernel** — no WASM and no second transition implementation. The committed corpus proves the Go authority and TypeScript module emit the same six-verdict union, including all five machine failures, and a divergence is an InvariantSink-severity parity bug. The module is included in the client build, but `verify(runLogArchive, catalogBytes)` is not currently bound to public archive/catalog readers or a player workflow. The draft Leaderboard Readers & Player Surface successor must own retrieval, invocation, refusal/error states, and browser acceptance before the designed transparency instrument is called shipped.

### L5 — Epoch storage and minting

`epochs(epoch_id bigserial, name, started_at, ended_at nullable, changelog_ref NOT NULL)` — **one current epoch enforced by partial unique index on `ended_at IS NULL`**; minting = one transaction: close current (set ended_at), insert next, insert its `catalog_artifacts` rows and `epoch_hashes(epoch_id, constants_hash)` set. Hotfix = insert one `epoch_hashes` row + its artifact (append-only; nothing is ever removed from an accepted set). **Run pinning:** the Prestige `run_started` write reads the current epoch **in the same transaction** (`FOR SHARE` on the current-epoch row) — a run started concurrently with a mint gets exactly one epoch, atomically. `changelog_ref` is a repo path (`changelog/epoch-<id>.md`); loader validation fails an epoch whose file is missing.

**L2b — Version-drift run policy (ruling, 2026-07-30):** play-time pin enforcement is hash-only.
On engine-version drift the command executes and the run is recorded once in append-only
`run_version_drift(company_stream_id, run_seq, observed_version, first_seen)`; drifted runs remain
playable, verify as `engine_mismatch`, and are excluded from board projection. A version bump may
never strand an active run.

**L5b — New-run hash transition (ruling, 2026-07-30):** the Exit transaction starts run N+1 under
the server's CURRENT constants hash (D6 assembly from the current catalog; new company revision and
pin both carry it). The ended run keeps its original pin. This is what makes D3's "pins to the epoch
live at its timer start" true across mints.

**L5c — Epoch seed sync (ruling, 2026-07-30):** `cmd/gameserver` reconciles DB epochs/hashes with
the repo seed idempotently at startup, before readiness; the process's constants hash is therefore
always in the current epoch's accepted set. An empty database replays every seed epoch and accepted
hash, with all predecessors closed; a non-empty database must be a valid prefix and may advance by
one epoch. Only the current worktree hash has artifact bytes available during an empty bootstrap;
historical identities are never assigned fabricated bytes.

### L6 — Board projection contract

`verified_runs(run_id PK, founder_id, category, variables jsonb, epoch_id, key_ms bigint | key_int bigint, verified_at, world_first bool)` — projected from verification events, **idempotent by event_id** (the commons/routes claim-then-validate pattern). Ordering: **standard competition ranking ("1224")** on the exact key; ties share a rank by construction because the key is the sort. Pagination: keyset cursor `(key, run_id)` — no offsets. Frozen boards are ordinary queries on ended epochs (freezing is a property of no-new-rows, not a mode). **World-first: partial unique index on `(category, variables, epoch_id) WHERE world_first`**; the verification transaction attempts `world_first = true` insert-on-conflict-do-nothing — atomic dedup, no race. `imported` founders (Account RFC D4) are excluded by the projection, never by query-time filtering.

### L7 — Categories and variables (ruling: two variables, not one bit)

Closed category schema: `{category_id, name, terminal_predicate, timer ∈ {rta, attended}}` where `terminal_predicate` is a **closed union over run-terminal facts** (`exit_type == X`, `tier_reached ≥ N`, `ledger_clean` — grows by RFC): the four canonical categories are four catalog rows with literal predicates. Variables are **structural and separate**: `commons: bool` (any compact membership this run), `advisor: bool` (Prestige D5), `glitched: bool` (any route_executed). **Boards key on the full variable tuple; "Solo" is the display name for `{commons: false, advisor: false}`** — nothing is ever a computed subtraction, and Commons-assisted vs Advisor-assisted remain separately queryable (resolving the routed-forward question).

**L7a — The canonical category catalog (owner answer, 2026-08-01, closing the projector
blocker):** four Phase-0 rows with LITERAL terminal predicates, expressed in a closed predicate
union evaluated over run-terminal facts (the `run_ended` payload — which gains two additive
fields, `gates_crossed` sorted list and `generators_purchased_total` int, schema_version bump per
the event registry rules; both already live in state):

- Predicate union (closed, grows by RFC): `any` · `all_gates` (gates_crossed == the catalog's
  full gate set) · `facts_superset(set_ref)` · `facts_disjoint(set_ref)` ·
  `count_at_most(field, literal)` · `all_of([...])`.
- `any_percent`: predicate `any`; timer `rta`.
- `hundred_percent`: `all_of([all_gates, facts_superset(completion_set)])`; timer `rta`.
  `completion_set` is a catalog-declared fact list (grows with content; starts with the Phase-0
  gate/doctrine facts).
- `ethical_percent`: `facts_disjoint(forbidden_set)`; timer `attended`. `forbidden_set` = the
  catalog-declared dark-pattern/externality fact kinds (the morality ledger's structural teeth —
  design/02 §7's Ethical% made executable).
- `low_percent`: `count_at_most(generators_purchased_total, low_max)`; timer `rta`; `low_max` a
  catalog literal (provisional 40).
- Future rows (`net_zero_percent`, `pacifist`) are content on this union, not new machinery;
  Exhibition and player-authored predicates stay behind the promotion thresholds (D4).

**L7b — the three L7a holes (owner answers, 2026-08-01):**

1. **Literal Phase-0 fact sets, honest about today's vocabulary.** `completion_set = []` — at
   Phase 0, 100% ≡ all_gates, DECLARED as such (side-completion facts arrive with content epochs;
   adding one is a mint). `facts_disjoint` gains prefix entries (a set member ending in `.`
   matches any fact with that prefix — one closed-union extension); `forbidden_set =
   ["darkpattern.", "externality."]` — structurally real, content-empty: every Phase-0 run
   trivially qualifies for Ethical% because no dark pattern is purchasable yet (TRUE), and the
   first content epoch emitting a `darkpattern.*` fact gives the category teeth with zero catalog
   surgery. The fact-namespace registry (`exit.*` today, `darkpattern.*`/`externality.*`
   declared) grows by RFC like every registry.
2. **Epoch ownership: the category catalog is a constants artifact** — `balance/categories/
   phase0.json` joins the epoch seed (append = mint, the factions/guilds pattern). Boards already
   key on epoch_id, so category evolution rides the existing machinery; a run is judged by its
   PINNED epoch's category catalog at verification, never the current one.
3. **The fifth canonical row — the count destination pre-timer runs need:** `valuation` — a
   count-class board on terminal `lifetime_value` under D1's exact `(exponent,
   quantized_mantissa)` key with shared-rank ties (the D1 magnitude rule finally has its
   shipping consumer). Predicate `any`; timer NONE (count boards have no timer field — the
   schema's timer becomes `rta | attended | none`). Pre-timer, and only pre-timer-compatible,
   boards accept pre-timer runs; imported and drifted runs remain excluded from ALL boards.

### L8 — CI hook (extends the hardened guard, weakens nothing)

The existing history guard already walks every reachable revision. Extension, same job: for any commit whose diff touches a `ConstantsHashArtifacts` path — **with `BALANCE-CHANGE:`** → the same commit must add an `epochs` seed row + changelog file (mint), else fail; **without** → the commit must add its resulting hash to the current epoch's accepted-set seed file (hotfix), else fail. Both are ordinary in-repo files, so the artifact-commit-touches-only-baseline rule and fetch-depth guarantees apply unchanged; reproducibility of a hotfix = its artifact row carries the exact bytes (L2). The cap-lowering migration rule routed here lands as: **a hotfix may not lower any hardcap** (guard compares the declared cap fields across the diff; lowering a cap is definitionally a balance change and requires a mint + the clamp-on-migration policy in its changelog).

`CONSTANTS-IDENTITY:` is composition repair only: every seed-declared artifact's bytes must equal
its bytes at the previous pacing-baseline commit. An artifact change between baselines therefore
cannot be relabeled identity-only even when the harness does not execute that artifact.

## Changelog

- 2026-07-28: created (draft). Closes the deferred-decisions register.
- 2026-07-29: updated implemented dependencies; Codex acceptance review recorded eight blockers.
- 2026-07-29: all eight answered with executable contracts L1–L8; Assisted ruled as two structural variables (commons, advisor); cap-lowering rule landed in L8.
- 2026-07-30: L8 guard implementation reviewed and approved; L2a added (seed as single artifact-set authority + parity test) from the review's MEDIUM finding; constants reverts documented as mint-only.
- 2026-07-30: core review found two architectural HIGHs rooted in this RFC's contracts; rulings L2b (version-drift runs stay playable, unrankable), L5b (run N+1 starts under the current hash), L5c (startup epoch seed sync) added.
- 2026-08-01: L7a — the four canonical category rows with literal predicates over a closed union; run_ended gains gates_crossed + generators_purchased_total (additive, schema bump).
- 2026-08-01: L7b — literal Phase-0 fact sets (empty completion set declared; prefix-matched forbidden namespaces); category catalog joins the epoch seed; fifth canonical row `valuation` (count-class, D1 magnitude key) as the pre-timer destination.
- 2026-07-30: L2a/L5c remediation centralizes artifact composition in `epochseed`, requires
  manifest reconciliation before gameserver readiness, and adds a hash-only baseline repair gate.
- 2026-07-30: round-2 review tightens the identity-only gate to pin every seed artifact's bytes at
  the preceding baseline, closing the hashed-but-unexecuted catalog escape.
- 2026-07-30: L5c empty-database reconciliation replays the full declared epoch/hash history instead
  of rejecting every deployment after the first mint or hotfix.
- 2026-07-30: governance-integrity remediation permits only the current→closed `ended_at`
  transition on epoch rows, rejects all historical rewrites/deletes, and constrains board run IDs
  to canonical Company-stream/run-sequence identity.
- 2026-07-29: accepted for implementation by `planning/codex-batch-2026-07-29.md`; implementation started immediately behind Prestige so L1 can replace its provisional terminal sequence.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
- 2026-08-21: reconciled D2/D4/D5/D6/L1/L4 to the six-verdict/five-category backend and routed
  absent readers, validator workflow, Route Registry composition, dispatch and machine boards to
  a draft successor; abandoned retention remains with D-015. No product behavior changed.
