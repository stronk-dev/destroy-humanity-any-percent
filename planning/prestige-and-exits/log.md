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
