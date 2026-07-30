# RFC: Leaderboards & Balance Epochs

- **Status:** implementing
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/05 §6` (derived boards, category model, Route Registry), `design/08 §6` (timer semantics, categories), `design/00` (world records as framing)
- **Research:** `design/research/speedrun-governance.md` (S1–S39 — the primary source), `design/research/adaptive-balancing.md §8` (Balance Epoch, Board Mandates; **as corrected by** speedrun-governance §5.3)
- **Depends on:** Production Engine (implemented), Save Layer (implemented — `constants_hash` per
  revision), Gate Predicates (implemented — `route_executed` events feed Glitchless/variables),
  account/session bootstrap, Prestige & Exits (run lifecycle)
- **Planning:** `planning/leaderboards-and-epochs/` (once implementing)

## Summary

Boards, epochs, and run verification — closing the **last entry in the deferred-decisions register** (ranking keys, from RFC-0002 draft D4). The architecture is the governance research's: **boards are derived from replayed intent logs, never authored**; runs pin to the epoch live at their timer start; ranking keys are exact.

## Specification

### D1 — Ranking keys (the registered decision, resolved)

**Order keys are exact integers or exact times — never quantized `Decimal`s.** Time boards rank on integer milliseconds (RTA / Attended Time); count boards on exact integers. Where a magnitude must rank (a "largest bank" novelty board), the key is `(exponent, quantized_mantissa)` and **equal keys display as ties, sharing a rank** — never resolved arbitrarily.

### D2 — The run record

A run is `(founder_id, run_id, category, variables, epoch_id, constants_hash, seed, intent log)`. **Verification is replay**: the server re-simulates the intent log against the epoch's catalog; a mismatch yields one of four machine causes (`log_gap`, `state_divergence`, `constants_mismatch`, `clock_violation`). No video, no queue, no human judgment in the loop; **the validator ships to players** (same shared kernel). Timer semantics per `design/08 §6`: RTA from `[BEGIN ATTEMPT]`; **Attended Time** = RTA minus offline spans (offline is published policy, not a rule dispute); IGT recorded, never ranked.

### D3 — Epochs

- An epoch is minted **deliberately**: `{epoch_id (monotonic), name, started_at, catalog constants_hash set, changelog_ref}`. Minting is a release act tied to a `BALANCE-CHANGE:` (Harness H3) and a public changelog entry — **a silent nerf is structurally impossible** because a balance change without an epoch fails the harness hook, and an epoch without a changelog fails this RFC's validation.
- **Hotfixes within an epoch** (correctness-only, `BALANCE-CHANGE:` absent) update the epoch's accepted-hash set; per-run `constants_hash` remains the forensic record of exactly what a run played under.
- **A run pins to the epoch live at its timer start, for its entire duration** (the governance correction — at our cadence essentially all first-ending runs straddle a boundary; the most emotionally significant run in the game must land on a board).
- Boards key on `(category, variables, epoch_id, mandate_level)`. Old epochs' boards freeze and remain browsable forever ("patch X and earlier" is a first-class historical object, not an attic).

### D4 — Categories and variables

Per `05 §6`, consumed here as data: 4 canonical categories (terminal conditions in code) + the player-authored predicate surface (threshold-promoted at ≥ 25 verified runs by ≥ 10 founders — provisional) + Exhibition. Variables: `Glitched` (any `route_executed` this run), **`Assisted`** (commons membership at any point — **structural disconnection**, per the governance spec: Solo and Assisted are different boards, never a computed subtraction), mandate level. The Route Registry's public ledger (Gate Predicates D3) renders alongside boards — routes and records are one surface.

### D5 — Board Mandates

The Ascension-style ladder from `adaptive-balancing §8`, consumed as balance data: Mandates 1–20 as additive rule modifiers, each a declared catalog object; `mandate_level` is a board key component. Mandate *content* is design/balance work, out of scope here — this RFC ships the key plumbing and validation.

### D6 — World-first and broadcast hooks

First verified completion per `(category, epoch)` emits a feed/dispatch event (the Ethical% world-first moment, `05 §5`); permanent, dated, tied to the verified run record. TAS/AGI boards (`08 §6`) are a distinct board class flagged `machine`, never merged with human boards.

## Acceptance criteria

1. No board query path accepts a quantized `Decimal` as a sort key (type-enforced); a fixture with sub-quantum differences ranks as a shared-rank tie.
2. Replay verification: a tampered intent log fails with the correct machine cause; the shipped validator reproduces the server verdict on the same log.
3. Epoch pinning: a run started in epoch N and finished in N+1 ranks in N, replayed against N's catalog.
4. A balance change without an epoch mint fails the harness hook; an epoch without a changelog entry fails validation.
5. Solo/Assisted: joining the compact mid-run moves the run to Assisted **at verification, structurally** — no arithmetic adjustment path exists in code.
6. Frozen epoch boards remain queryable after two subsequent epochs.

## Open questions

- Promotion thresholds (25/10) and mandate count: provisional, data.
- Board retention/pagination scale: implementation freedom.

## Executable contracts (answering the 2026-07-29 review)

### L1 — The run log (the missing table, owned here)

`run_log(run_id, seq bigint, intent_id, canonical_payload bytea, receipt jsonb, applied_revision, server_ts_ms)` — **written in the same transaction as intent commit** (the production engine gains one insert), `PRIMARY KEY (run_id, seq)`, seq strictly monotonic per run. `canonical_payload` is the exact canonical-JSON bytes the idempotency hash was computed over (so the log is self-verifying against `intent_records` hashes while those still exist). Retention: rows live until the run is **verified+archived** (log compressed into an immutable `run_log_archive` object, one blob per run) or **abandoned** (no terminal event within `run_ttl_days` catalog value, provisional 90 — then deleted, run unrankable). Exempt from the 30-day `intent_records` prune; the prune and this table never share a policy.

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

### L4 — Player validator delivery

**The validator is the existing TS shared kernel** — no WASM, no second implementation: the golden-vector regime already holds Go and TS byte-identical, and that regime *is* the parity proof. Delivery: the client bundle ships a `verify(runLogArchive, catalogBytes)` entry that replays and emits the same five-cause verdict. Fixtures: every verification cause gets one committed `(log, catalog, verdict)` fixture asserted by **both** suites. The server verdict is authoritative; the shipped validator is the transparency instrument (governance §5.3's requirement), and any Go/TS verdict divergence is by definition a kernel parity bug — InvariantSink severity, not a rules dispute.

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
always in the current epoch's accepted set.

### L6 — Board projection contract

`verified_runs(run_id PK, founder_id, category, variables jsonb, epoch_id, key_ms bigint | key_int bigint, verified_at, world_first bool)` — projected from verification events, **idempotent by event_id** (the commons/routes claim-then-validate pattern). Ordering: **standard competition ranking ("1224")** on the exact key; ties share a rank by construction because the key is the sort. Pagination: keyset cursor `(key, run_id)` — no offsets. Frozen boards are ordinary queries on ended epochs (freezing is a property of no-new-rows, not a mode). **World-first: partial unique index on `(category, variables, epoch_id) WHERE world_first`**; the verification transaction attempts `world_first = true` insert-on-conflict-do-nothing — atomic dedup, no race. `imported` founders (Account RFC D4) are excluded by the projection, never by query-time filtering.

### L7 — Categories and variables (ruling: two variables, not one bit)

Closed category schema: `{category_id, name, terminal_predicate, timer ∈ {rta, attended}}` where `terminal_predicate` is a **closed union over run-terminal facts** (`exit_type == X`, `tier_reached ≥ N`, `ledger_clean` — grows by RFC): the four canonical categories are four catalog rows with literal predicates. Variables are **structural and separate**: `commons: bool` (any compact membership this run), `advisor: bool` (Prestige D5), `glitched: bool` (any route_executed). **Boards key on the full variable tuple; "Solo" is the display name for `{commons: false, advisor: false}`** — nothing is ever a computed subtraction, and Commons-assisted vs Advisor-assisted remain separately queryable (resolving the routed-forward question).

### L8 — CI hook (extends the hardened guard, weakens nothing)

The existing history guard already walks every reachable revision. Extension, same job: for any commit whose diff touches a `ConstantsHashArtifacts` path — **with `BALANCE-CHANGE:`** → the same commit must add an `epochs` seed row + changelog file (mint), else fail; **without** → the commit must add its resulting hash to the current epoch's accepted-set seed file (hotfix), else fail. Both are ordinary in-repo files, so the artifact-commit-touches-only-baseline rule and fetch-depth guarantees apply unchanged; reproducibility of a hotfix = its artifact row carries the exact bytes (L2). The cap-lowering migration rule routed here lands as: **a hotfix may not lower any hardcap** (guard compares the declared cap fields across the diff; lowering a cap is definitionally a balance change and requires a mint + the clamp-on-migration policy in its changelog).

## Changelog

- 2026-07-28: created (draft). Closes the deferred-decisions register.
- 2026-07-29: updated implemented dependencies; Codex acceptance review recorded eight blockers.
- 2026-07-29: all eight answered with executable contracts L1–L8; Assisted ruled as two structural variables (commons, advisor); cap-lowering rule landed in L8.
- 2026-07-30: L8 guard implementation reviewed and approved; L2a added (seed as single artifact-set authority + parity test) from the review's MEDIUM finding; constants reverts documented as mint-only.
- 2026-07-30: core review found two architectural HIGHs rooted in this RFC's contracts; rulings L2b (version-drift runs stay playable, unrankable), L5b (run N+1 starts under the current hash), L5c (startup epoch seed sync) added.
- 2026-07-29: accepted for implementation by `planning/codex-batch-2026-07-29.md`; implementation started immediately behind Prestige so L1 can replace its provisional terminal sequence.
