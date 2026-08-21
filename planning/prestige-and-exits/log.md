# Prestige & Exits — append-only log

## 2026-07-29 — start

- Accepted through `planning/codex-batch-2026-07-29.md` after P1–P8 answered the implementation
  review. The first timing envelope is elective; the scripted first Exit remains a fixed curriculum
  segment.
- Implementation follows the accepted seam: save v7 first, pure cross-runtime arithmetic second,
  multi-stream storage third, then Production intents and deterministic evaluation hooks.
- `founder seed` in P3 is not a P1 persisted field. The implementation will derive the stable
  unsigned seed from the immutable Founder UUID and `run_seq`, and records that literal mechanical
  interpretation here rather than adding an undeclared save field.
- Phase-0 has no harvested-Quarter intent, IPO chain, carried Network content, or Reputation tree.
  Those accepted fixture seams remain empty/gated; threshold-crossing offers, Wind Down, and the
  full transaction are implemented without inventing content.
- `lifetimeValue` is named but its resource valuation is not defined in the RFC. The Phase-0 policy
  therefore declares `value_resource_id=company.cash`; authoritative positive accrual in that
  resource advances lifetime value. Future multi-resource valuation is balance data, not a hidden
  conversion invented in engine code.

## 2026-07-29 — save v7

- Added the exact P1 Company and Founder state fields plus P6's ordered offline spans. All times are
  integer milliseconds at rest and canonical UTC `time.Time` values in memory; prestige integers
  stay inside the exact-number domain.
- Validation rejects cross-scope leakage, malformed offers, duplicate/unsorted Network slots,
  non-append-only Exit histories, overlapping spans, unsafe timestamps, and non-finite or negative
  lifetime value. The list cap is 256 entries; the evaluation layer owns oldest-span collapse.
- Version 1–6 migrations receive explicit zero/null/empty defaults. The only-grows corpus now has a
  Founder-v6 fixture as well as the existing Company history. New account-created company runs set
  `run_started_at` from canonical server time.
- `go test -p 1 ./...` passes across every server package after the save bump.

## 2026-07-29 — numeric policy core

- Added a strict declarative Prestige policy and JSON Schema. The provisional P8 fixture supplies
  the `1e12` threshold, typed payout modifiers, collapse Knowledge, offer lifetime/drift, tier spawn
  gates, and Advisor constants; unknown fields fail in both the Go loader and schema gate.
- Added the P2 integer binary-search cube root in Go and TypeScript. Both consume one shared corpus
  covering zero, threshold, `n³` below/exact/above boundaries, modifier flooring, a large exact
  cube, and the numeric-domain cap. Neither runtime calls floating `cbrt`.
- Added exact moral reseeding, bounded Advisor multiplication, and the mandated SplitMix64 stream
  with fixed output vectors. Go tests, all 6,422 client tests, strict TypeScript/Svelte checks, and
  the expanded schema verifier pass.

## 2026-07-29 — atomic Exit store

- Added `ApplyExitTransaction`: it discovers the active sibling Founder stream, locks Founder then
  Company, checks both expected revisions, and writes Founder advance, final old-Company revision,
  new-run Company revision, all events, and the Company-keyed idempotency record in one transaction.
- Added the six Prestige event kinds to both the Go closed registry and migration 00011's database
  constraint. Strict payload validators cover offer transitions, complete run-end rendering facts,
  structural Assisted variables, new-run identity, and Founder payout facts.
- Real-Postgres tests inject rollback failures after each of eight write boundaries and prove both
  streams remain at revision 1 every time. The successful path advances Founder `1→2` and Company
  `1→3`; retry returns the same receipt and the identical three ordered event records without
  rerunning mutation.
- Adversarial replay review found that database UUID/timestamp order could differ from first-delivery
  group order. Replay now applies an explicit Founder-advance → run-ended → run-started order and
  the integration test compares event IDs position by position.

## 2026-07-29 — lifecycle runtime

- Implemented deterministic threshold-site offer generation, persisted preview terms and market
  modifier, typed expiry/decline, promise-preserving acceptance, elective Wind Down, and a gated
  IPO seam. Offer identity and PRNG draws derive only from immutable Founder/run inputs.
- Production accrual now advances the declared lifetime-value resource and records server-derived
  offline spans. The scripted-first path fires on the first eligible threshold crossing, commits
  through the same atomic Exit operation, and is prevented by append-only Founder history from
  repeating.
- New-run assembly is byte-deterministic and carries only declared Network/Reputation seams before
  reseeding moral bands. No Phase-0 fixture content was invented for those still-empty seams.
- `run_ended` now reads executed routes from canonical Company events through the request's expected
  revision. If another route commits concurrently, the Exit CAS fails; facts from the wrong
  revision cannot enter the obituary. The real-Postgres test asserts preservation.
- The event currently carries the terminal Company revision as `terminal_seq`. This is deliberately
  provisional under P7 and is replaced immediately by Leaderboards L1's transaction-local run-log
  sequence; Prestige cannot archive before that integration lands.
- **DESIGN-GAP:** D5 names Advisor Mode as a Founder-scoped toggle, but D1's closed intent union does
  not provide a toggle intent. The persisted field, multiplier semantics, and structural assisted
  fact exist; no undeclared player command was invented. The independent review/owner must either
  add a closed intent or confirm that another accepted RFC owns the control surface.
- AC8 remains blocked by this RFC's explicit T0–T1 playable-content dependency. The current harness
  has no eligible elective Exit policy and cannot honestly measure a 45–90 minute first elective
  Exit. The transaction/runtime hook is implemented; no synthetic balance milestone was added and
  presented as evidence.
- Focused verification is green: all Go packages; strict TS/Svelte checks; 6,422 client tests;
  client shell boundary lint; real-Postgres elective, scripted, preview-promise, route-fact, eight
  fault-boundary, and replay integration cases.

## 2026-07-29 — Leaderboards L1 integration

- The provisional terminal Company revision has been replaced by the Exit intent's transaction-local
  run-log sequence. `run_ended.terminal_seq` and the committed `run_log.seq` are asserted equal in
  the real-Postgres lifecycle test, closing P7 without a second write path.

## 2026-07-30 — independent review: prestige implementation (8c20820..db38c80, a67e93b) — the review I owed

Adversarial two-lens review; the four load-bearing findings re-verified against source by the
reviewer (state.go:355-363, prestige.go:129-131, runtime.go:173-176, exit.go:275-280 all read
first-hand). **Verdict: P2's core search, P4's atomicity (8-boundary fault injection with
byte-identical replay), and P5's once-ever are verified correct with evidence. One HIGH and a band
of MEDIUMs — several rooted in MY contracts — with rulings below, now normative in the RFC.**

Findings (fix queue, ordered):

1. **HIGH — migrated pre-v7 companies can never Exit**: v<7 migration leaves `RunStartedAt` zero,
   `AttendedMS` errors on zero, every exit fails un-typed forever; the scripted trigger silently
   never fires. **Ruling P6a-migration:** next save version backfills `RunStartedAt :=
   EvaluatedThrough` when zero at migration; such runs are flagged pre-timer (excluded from
   time-ranked boards — they predate the timer contract anyway). Corpus gains the company v6→v7→v8
   case (finding: the company-scope v6 branch currently has zero corpus coverage).
2. **MEDIUM — the scripted first failure is skippable** (elective `wind_down` accepted for a
   zero-exit founder, contradicting D4's "cannot be skipped"). **Ruling P5b:** an elective
   `wind_down` from a founder with empty `exit_history` IS the first failure — the engine types it
   `scripted_first` with full first-exit payout regardless of attended time. D1's always-open door
   and D4's unskippable curriculum are both preserved; AC4 gains this case.
3. **MEDIUM — offline-span collapse absorbs online time into offline** (runtime.go:173-176 rewrites
   span[1].From), shrinking attended time and falsifying the doc's "without changing the total
   offline duration". **Ruling P6b:** overflow drops the oldest span into a new
   `collapsed_offline_ms` accumulator (save field, next version); attended = RTA − (Σ spans +
   accumulator); total-offline invariance becomes a property test.
4. **MEDIUM — Go/TS division parity is not "by construction"**: Go `Div` is mantissa/mantissa (one
   rounding), break_infinity is reciprocal-multiply (two) — a non-unit-mantissa threshold (schema-
   legal balance data) can flip a cube boundary by 1 ulp between runtimes; no vector covers it.
   **Ruling P2b:** the TS kernel wraps division as mantissa/mantissa to match Go; golden vectors
   gain non-unit-mantissa thresholds (e.g. 2.5e12) at cube boundaries.
5. **MEDIUM — `ReputationDelta` overflow: Go errors (trapping exits at extreme lifetime values ≈
   ≥5e41), TS computes through floats.** **Ruling P2c:** both runtimes SATURATE the modified delta
   at MaxExactInteger; never an error on the exit path; parity vector at the saturation boundary.
6. **MEDIUM — idempotent replay reorders events when accrual-hook events prefix the exit trio**
   (exitEventOrder maps unknown kinds last). Fix: return recorded order (order by committed event
   seq); the existing byte-identical-replay test extends to a prefixed fixture.
7. **MEDIUM — out-of-order gate crossing hard-errors forever** (`setTierFromGate`: a legal
   lower-tier gate after a higher one → `ErrInvalidEngineState` un-typed, every retry). **Ruling:**
   tier is `max(current, gate tier)`; the gate still records `gate_crossed`; never an internal
   error for catalog-legal input.
8. **MEDIUM — the "archived-final" company revision is pruned by the same op's 5-revision
   retention.** **Ruling P4b (spec correction, mine):** the `run_ended` event + run log are the
   obituary's source of record (AC7 already guarantees the screen renders from the event alone);
   P4's "archives are revision-history" language is corrected — revision retention stays at 5.
9. LOW — DeleteAccount archives both streams in one unordered UPDATE (deadlock window vs the
   exit's founder-then-company order; make it two ordered updates). LOW — decline-drift count is
   whole-stream, so declines carry across runs undocumented (scope `CountEvents` by run or document
   the carryover in P3 — ruling: scope by run; drift resets each run). LOW — planning-log claim
   about offer identity deriving only from immutable inputs is false (OfferID embeds now-ms and
   mutable tier/declineCount; harmless for uniqueness, log corrected by this entry).
10. OBSERVATIONS — Advisor Mode is flag-only with the toggle intent correctly logged as a gap;
    the runtime is wired only in tests and `catchup_ceiling_ms` is not yet bound to its catalog
    value (composition work); P6's attended-time-is-intent-cadence semantics noted as pacing-
    relevant (spec-faithful); a67e93b's rebaseline is protocol-compliant with the noted wrinkle
    that six intervening commits carried stale golden hashes (pre-guard; the epoch guard now
    prevents recurrence).

## 2026-07-30 — HIGH remediation: save v8 pre-timer migration

- Save v8 adds the persisted `run_pre_timer` fact. Restoring a pre-v7 Company backfills
  `RunStartedAt` from its already-authoritative `EvaluatedThrough` cursor and marks the run; new
  runs remain ordinary timer runs. `run_ended` carries the fact for verification and board policy.
- The migration corpus now contains the previously absent Company-v6 branch and ratchets its
  append-only baseline. A real Postgres fixture rewrites an actual Company revision as v6, loads it
  through the normal store, crosses a threshold after 16 attended minutes, and completes the
  scripted Exit. The next run is v8 and not pre-timer.
- Focused Go tests/vet and the full Postgres integration target are green.

## 2026-07-30 — prestige review remediation batch

- Implemented P2b/P2c with a direct mantissa/exponent ratio in TypeScript, `big.Int` ppm arithmetic
  in Go, and exact-domain saturation in both runtimes. Shared vectors now cover non-unit threshold
  cube boundaries and the saturation path.
- Save v9 adds `collapsed_offline_ms`. Evicting the oldest of 256 offline spans transfers only its
  exact duration into the accumulator; a 257-span alternating online/offline property test proves
  Attended Time stays invariant. V8 restoration defaults the accumulator to zero and current
  encoding makes the field mandatory.
- Exit replay now follows a database-authored `event_seq`, removing the kind-based reorder. The
  real-Postgres fixture includes an offer event between the Founder and terminal Company events and
  requires initial delivery and replay to be position-identical. Upgrade backfill preserves the
  former closed-kind replay order for already-committed intents; new rows record insertion order.
- An empty-history Wind Down is `scripted_first`; gate tier advancement is monotonic; decline drift
  counts only current-run events; extreme Reputation payout is clamped while terms are computed to
  remaining Founder headroom, preserving both the hardcap and preview-is-a-promise.
- Canonical docs now describe save v9 and the implemented replay, timer, arithmetic, and lifecycle
  behavior. Independent review remains required before archival.

## 2026-07-30 — independent review: c3af220 (lifecycle findings round)

**Approved with two MEDIUM findings, one requiring a new ruling.** Verified to the letter: P5b's
elective path (wind_down + empty history → scripted_first, full modifier, no attended gate;
threshold path and once-ever intact); P6b save v9 accumulator with exact-duration eviction, Σ-bound
invariants, corpus + leak-check coverage; P2b TS mantissa/mantissa ratio structurally identical to
Go with non-unit-mantissa cube-boundary vectors in the shared corpus; P2c big.Int saturation both
runtimes + founder-headroom clamp preserving preview-promise; replay order by committed event_seq
with a faithful backfill and a prefixed-event fixture; gate tier = max(current, gate) with
gate_crossed intact; decline drift scoped per run with validator-enforced run_seq.

1. **MEDIUM — P5b is incomplete on the OFFER path** (verified first-hand at prestige.go:132-150):
   only wind_down retypes; `accept_exit_offer` takes the stored offer's exit type with no
   empty-history check, and offer spawning is not history-gated — a run-1 founder who crosses a
   tier-1+ gate can accept a spawned acquisition and skip the curriculum forever. **Ruling P5c:
   offers do not spawn while `exit_history` is empty** — the market does not notice you before
   your first collapse (fiction-consistent, and it keeps offer previews honest instead of retyping
   an accepted acquisition into a scripted collapse). AC: spawn-gated fixture + the existing
   once-ever suite.
2. **MEDIUM — P2c residual: TS `reputationDelta` still computes delta×ppm/1e6 through
   break_infinity floats** — reproduced ±1 divergence vs Go's exact big.Int at
   delta=9007199254740988 × the actual collapse modifier; the corpus samples only agreeing points,
   so both suites are green while runtimes disagree. **Ruling P2d: the delta-modifier product is
   exact integer arithmetic in BOTH runtimes** (TS BigInt — the values are integers by
   construction); the reproduced mismatch point joins the shared corpus as a vector.
3. LOW — decline-drift discontinuity at deploy (historical declined events lack run_seq; mid-run
   drift silently resets once — accepted as a one-time upgrade artifact, recorded here) and the
   count is an unindexed JSON scan per cross_gate (index or counter-cache when it shows up in
   profiles).

## 2026-07-30 — P5c/P2d remediation

- Offer spawning returns without drawing while Founder Exit history is empty. A real-Postgres
  fixture proves a T8 crossing emits no offer before the scripted exit, while a seeded second run
  still spawns and accepts its promised offer.
- TypeScript modifier application is exact BigInt multiplication and integer division. The shared
  mismatch vector agrees with Go at 6,755,399,441,055,741.
- Added the expression index used by run-scoped decline counts.

## 2026-08-21 — authority and canonical-body reconciliation

- Owner direction delegated the blocked ruling-body reconciliation after D-012 had already
  deferred Advisor Mode from the Phase-0 preview. No implementation authority was inferred for
  the missing toggle, settings control, activation timing, or authored player copy.
- Reconciled D2 to the strict `expires_at_ms` / `reputation_delta` event contract; D3/P4 to
  `run_ended` plus the durable run log as ended-run authority; and P3 to the only live
  threshold-crossing offer site. The deferred Quarter bridge remains successor work.
- Reconciled D4/P5 and canonical docs to the epoch-pinned T0–T1 trigger: run 1, zero Founder Exits,
  Garage gate already crossed, 900,000 attended ms, then replacement of the next otherwise-valid
  player Company command. Early Wind Down remains `scripted_first` and offers remain history-gated.
- Reconciled D5 and `design/11` around D-012 while preserving the already-authored quoted label
  byte-for-byte. The persisted field/math/event label are explicitly mechanical seams, not a
  shipped player capability.
- No product code, balance data, player-facing copy, plan checkbox, RFC status, push, publication,
  or archival changed. AC2–AC6 literal witness remediation and the cross-party archival review
  remain open.
