# RFC: Leaderboards & Balance Epochs

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/05 §6` (derived boards, category model, Route Registry), `design/08 §6` (timer semantics, categories), `design/00` (world records as framing)
- **Research:** `design/research/speedrun-governance.md` (S1–S39 — the primary source), `design/research/adaptive-balancing.md §8` (Balance Epoch, Board Mandates; **as corrected by** speedrun-governance §5.3)
- **Depends on:** Production Engine (implemented — the intent log is the run record), Save Layer (implemented — `constants_hash` per revision), Gate Predicates (draft — `route_executed` events feed Glitchless/variables)
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

## Changelog

- 2026-07-28: created (draft). Closes the deferred-decisions register.
