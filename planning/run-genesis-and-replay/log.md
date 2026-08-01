# Run Genesis & Replay Verification — running log

## 2026-07-31 — acceptance pass: RA/RB bounced with executable blockers

Guild archived first as required (`06106c9`). The Run Genesis RFC and every current logged
transition call site were then read against RA/RB. The storage/genesis half is coherent, but the
declared closed `replay_inputs` object and four-argument `ApplyLogged` boundary cannot reproduce
the live engine yet. The gaps are source-demonstrated, not speculative:

1. `canonical_payload` excludes `intent_id`, while receipts/events require it; `State` also lacks
   stream/owner identity and the current revision/run-log sequence. `ApplyLogged` receives none of
   those values.
2. `EvaluationMode` changes accrual efficiency/caps but is absent from all four arguments and RA.
3. “catalog” is singular, while transitions consume immutable economy, Routes, Commons,
   Prestige, faction, and Guild artifacts under one constants hash.
4. The canonical hook chain reads a Commons participation weight from a projection, Guild
   membership through provider output, and future clearing slices; only a nullable Commons
   modifier is declared. The Commons weight alone changes persisted Solidarity and receipt events.
5. Cross-gate offer generation consumes the Founder snapshot and declined-offer count. Terminal
   Exit consumes Founder state, executed-route projection facts, current constants hash, and
   produces a next-run receipt whose snapshot depends on Founder carry. R1 explicitly declines a
   Founder genesis, so the current terminal receipt cannot be recomputed from the four arguments.
6. RA says `replay_inputs NOT NULL` and “no backfill / old rows verify log_gap” simultaneously.
   Postgres needs the exact legacy representation and new-write invariant before the migration is
   immutable.
7. RB returns only `(state,receipt)`, but the live owned result is `IntentDecision`/`ExitDecision`
   including events; the contract must say whether logged event parity is verified or deliberately
   outside the shared transition.
8. The hook-order test currently pins only Prestige→faction and does not name Guild→Commons or a
   closed extension registry. Replay determinism cannot depend on ServiceOption construction
   order.

The RFC now carries C1–C8 owner decisions with recommended executable shapes. No code schema was
improvised because changing any of these answers changes persisted replay meaning.

## 2026-07-31 — C1–C8 accepted; RA/RB implementation authorized

The owner accepted all eight proposed contracts as normative. Cross-run Founder mutation
verification remains a named non-goal; the terminal founder-carry input exists only to reproduce
the Exit receipt and embedded next-run snapshot. Event bytes and order are part of replay parity,
and the replayable Phase-0 accrual chain is closed in Prestige → faction → Guild → Commons order.
The replay-input schema starts directly at this accepted shape because no earlier producer exists.

The RFC is now implementing. The landing order remains: immutable replay inputs and the shared
`ApplyLogged` live/replay boundary first, then run genesis, then verification and projection.

## 2026-07-31 — RA replay-input persistence implemented

Migration 00030 adds nullable historical `run_log.replay_inputs` plus a fail-closed insert trigger:
old rows preserve the honest `log_gap` marker, while every new row must carry an object. The Store
now supplies the authoritative six-coordinate command envelope to logged mutations, validates it
against the returned replay object, normalizes the object, and inserts payload + inputs + receipt in
the gameplay transaction for applied and rejected commands, including Exits.

Production owns a strict version-1 resolved union. It freezes evaluation time/mode, canonically
ordered Decimal contribution strings, explicit Guild settlement batches, Commons weight, Route
context version, offer-count/Founder inputs, and terminal Founder carry/executed routes/selected
terms/next hash. Terminal carry is copied before Founder mutation. Postgres integration proves the
new-write trigger, ordinary and terminal capture, and rollback at the run-log fault boundary.

This is the RA landing only. RB remains open: the next batch makes the live engine consume this
same object through `ApplyLogged`, closes the hook registry, and removes live transition reads not
represented by its four arguments.

## 2026-07-31 — RB Go live/replay boundary implemented

The Go live path now computes the replay object first and invokes the exported `ApplyLogged`
boundary that future verification invokes. The boundary reconstructs the player intent from the
intent-less canonical payload plus the authoritative command envelope, rejects clock regression,
loads a closed six-artifact `CatalogBundle`, reconstructs canonical contribution values, and
returns state, receipt, and ordered event writes. Cross-gate offer resolution consumes only the
frozen Founder carry and decline count.

Terminal commands use the matching `ApplyLoggedExit` arm. It reconstructs the minimal Founder
view, computes the Exit receipt and next-run snapshot under the declared next hash, and returns the
Founder-derived output for the live transaction to merge into the full Founder stream. A persisted
elective-Exit integration fixture replays from the pre-command Company state and compares receipt
bytes plus every event kind/payload in committed order.

Replayable accrual construction is no longer service-option extensible. Its registry is exactly
Prestige → faction → Guild → Commons, with Commons consuming a frozen weight through
`commonsbinding.ResolvedHook`; adding a fifth hook now requires an explicit code/RFC change. The
catalog loader rejects any artifact map other than the six accepted kinds.

The Go half of RB is complete. The TypeScript validator port and cross-runtime full-run fixture are
the next plan item and remain unclaimed.

### Self-review correction before independent gate

The first RB commit still exposed an optional fifth `InvariantSink` parameter on `ApplyLogged`.
Although intended only for diagnostics, the concrete sink type controlled whether invariant events
were materialized, so it was an undeclared behavioral input and contradicted the four-argument
contract. The boundary now constructs its own collector and returns invariant reports alongside
state/receipt/events; the live service forwards those reports to metrics after the deterministic
transition. Both ordinary and terminal entry points now have exactly four data arguments.

## 2026-07-31 — independent review: RA/RB Go landing (21be3eb..ec872ec)

**Verdict: the reroute itself is APPROVED — RB's same-code-by-construction promise is real and
verified to an unusual depth** (every non-test transition entry accounted for; zero ambient reads
in the boundary — no time.Now anywhere in transition code, PRNG fully state/command-derived;
baselines byte-untouched; C1–C5/C7/C8 verified with evidence including terminal receipt+event
byte/order parity against live DB rows; hooks closed BY CONSTRUCTION — the extras registration
surface was deleted outright; ec872ec's self-caught fifth-argument removal honestly recorded).
**Two MEDIUMs must land before genesis storage; rulings below.**

1. **MEDIUM (verified first-hand) — run_log is mutable at the SQL layer**: 00030's trigger is
   INSERT-only and run_log has no `reject_immutable_change` at all — `UPDATE run_log SET
   replay_inputs=NULL` launders tampering into a benign legacy `log_gap` verdict. **Fix: one
   `BEFORE UPDATE OR DELETE` immutability trigger on run_log** (the same function every other
   forensic table uses). C6's letter was met; its premise wasn't.
2. **MEDIUM (verified first-hand) — the founder/catalog coherence guard was dropped in the
   reroute**: the old fail-closed `founder hash == company hash` check survives only in
   now-dead-code `afterPrestigeTransition` (zero callers); the live path builds the founder carry
   uncheckedand freezes it into replay_inputs — the mixed-catalog class RA exists to kill becomes
   UNAUDITABLE instead of impossible. **Ruling C5a: (a) Handle re-adds the fail-closed equality
   check before constructing the carry; (b) the C5 carry view gains `founder_constants_hash`,
   asserted equal to the bundle hash INSIDE ApplyLogged** — live fails closed AND replay audits
   the same invariant. Delete the dead function.
3. **Rulings on the LOWs:** **C4a** — the settlement-batch validator accepts well-formed non-empty
   batches now (the contract says representable; production writes empty until GD5 composition —
   a validator that contradicts the schema is re-versioning debt). **C4b** — the `route_hint`
   resolved arm is DELETED (founder-scope intents are non-replayable by design; a dead union arm
   would mislead the TS port). **RB-1** — prestige-less company-scope services are no longer
   legal: `NewService` requires the prestige runtime (the permissive construction is a trap that
   hard-errors on cross_gate). **C3a** — `CatalogBundle.valid` recomputes the constants hash over
   the six artifact bytes and compares to the label (the relabeling the repo's own test performs
   must be impossible in production); the catalog_artifacts-backed resolver lands with
   composition. **RB-2** — `FactionStockResource` is derived inside ApplyLogged from
   state.FactionID + bundle (restored-revision replays currently diverge on a receipt field).
   Fix the dead contribution-freeze assignment for invalid intents while there.
4. NOTES accepted as recorded: kernel-only union closure at the persistence layer (intentional
   layering; the save-layer writer is single); C6 reader clause lands with the verifier; the
   mid-round shape change needed no version bump (no producer existed).

## 2026-07-31 — RA/RB review remediation implemented

The two gating findings and all attached rulings landed in three reviewable commits. `bb04c72`
makes active `run_log` evidence immutable against both UPDATE and DELETE using the repository's
standard forensic trigger; its real-Postgres regression attacks replay-input laundering and row
deletion directly. `e9f28e0` restores Founder/Company catalog coherence before carry construction,
records `founder_constants_hash`, and reasserts it inside both logged transition arms. The dead
pre-reroute helper was deleted.

The same runtime landing closes C4a/C4b/RB-1/C3a/RB-2: ordered non-empty Guild settlement batches
are accepted with UUID and safe-integer validation; the founder-only route-hint union arm is gone;
production services require the Prestige/faction runtime; every six-artifact bundle recomputes its
hash from frozen exact bytes and rejects relabeling; and faction stock-resource identity is derived
inside `ApplyLogged`. Invalid intents no longer call or freeze contribution providers.

Verification is green: all Go packages, the focused replay/account suites, and the direct
Compose-owned Postgres integration command. The latter is intentionally recorded as
`docker compose -f compose.save-test.yml run --rm test`; an attempted environment-wrapped Make
alias incorrectly triggered sandbox escalation despite AGENTS.md already forbidding that wrapper.
No approval-dependent test path is part of this landing.

The TypeScript `ApplyLogged` port and cross-runtime fixture remain the next implementation item.

## 2026-08-01 — independent review: RA/RB remediation (bb04c72, e9f28e0) — APPROVED

Direct diff review; all seven rulings verified in source:

1. 00031 gives run_log the standard `reject_immutable_change` trigger (UPDATE OR DELETE) with a
   negative test — the tamper-laundering hole is closed.
2. **C5a closed elegantly:** the carry view carries `founder_constants_hash` (populated from the
   founder revision live-side) and ApplyLogged asserts it equals the bundle hash at BOTH the
   ordinary and terminal sites — because the live path routes through ApplyLogged, one in-boundary
   assertion both fail-closes live AND audits replay. Better than the two-mechanism shape I ruled.
3. `afterPrestigeTransition` dead code deleted. C4b: `route_hint` is an explicit typed rejection
   inside the boundary, not a dead union arm. RB-1: `NewService` now requires prestige runtime +
   faction catalogs + current hash. C3a: `replaycatalog.Load` recomputes `ConstantsHashArtifacts`
   over the six artifact bytes and compares to the label, with tests. RB-2:
   `deriveFactionStockResource` runs inside ApplyLogged before the transition; the incorporate
   path remains the in-boundary writer. C4a landed per the diff (batch validator accepts
   well-formed non-empty).

Next landing: TS ApplyLogged port + cross-runtime transition fixture, then genesis storage.

## 2026-08-01 — TypeScript ordinary ApplyLogged port

The TypeScript verification kernel now owns the ordinary (non-terminal) `ApplyLogged` boundary.
It loads the exact six hash-pinned artifacts, restores a strict Company save-v12 state, validates
the closed replay envelope and ordered resolved inputs, evaluates through the same Decimal rules,
and emits the authoritative receipt plus ordered events for every ordinary Phase-0 Company intent.
The closed accrual hook order is Prestige → faction → Guild → Commons. Pre-transition rejections
run before accrual where the Go service does; post-evaluation affordability/predicate rejections
retain the Go mutation boundary.

A Go-authored shared fixture currently covers 13 independent transitions: online/offline manual
work, buy, gate, deterministic offer spawn, Compact sign/leave, Open Source incorporation, offer
decline, the complete hook chain, invalid input, and two reject-before-accrual cases. Go regenerates
the committed artifact and fails on drift; TypeScript compares receipt, event bytes/order, and
post-state. Adding the offer branch exposed and fixed a real port bug: Go's `OfferID` consumes
separate SplitMix64 draws for its last two bytes, while the first TS draft reused one draw.

This is an intermediate RB landing, not AC2 completion. The RFC's ≥50 mixed-intent sequential run
still requires the terminal Exit arm and the genesis/verifier work; it is deliberately not being
papered over by counting independent transition cases. Root-only verification is green:
`make typecheck`, `make test-client` (6,467 passed), and `make replay-fixture-check`.

## 2026-08-01 — TypeScript terminal ApplyLogged arm

The TS boundary now reproduces terminal mutation as well: frozen Founder carry validation,
attended-time accounting, terms and stored-offer promises, final old-run evaluation, Founder-
derived output, new-run construction under the declared next bundle, terminal receipt, and the
Founder/ended-run/started-run event families in their persisted grouping and order.

Three Go-authored terminal fixtures cover wind-down scripted-first, accepted stored offer, and the
15-minute scripted cross-gate path. Each asserts the receipt, Founder carry output, final Company
state, next Company state, and all event bytes in TypeScript. The cross-gate fixture also records
the important attended-time premise: the run can be 20 minutes old while the evaluation cursor is
current; an unobserved 20-minute accrual gap is correctly classified as offline and cannot trigger
the scripted attended-time exit.

The shared artifact now contains 13 ordinary and 3 terminal transition cases. Root-only checks are
green: `make test-go GO_PACKAGES='./production'`, `make typecheck`, and `make test-client` (6,470
passed). AC2's sequential ≥50-intent full-run fixture remains part of the verifier landing.

## 2026-08-01 — immutable run genesis storage

Migration 00032 adds `run_genesis` keyed by Company stream/run sequence with immutable state bytes,
save version, and constants hash. A deferred constraint rejects any committed `run_epochs` insert
that lacks its matching genesis, so “every pin has exactly one genesis” is structural rather than a
production-call-site convention. The public pin helper captures the latest revision for existing
composition/test callers; account creation and Exit use the explicit pin+genesis transaction seam.

Genesis bytes are captured from PostgreSQL's `INSERT ... RETURNING state::text`. The first draft
stored the pre-insert Go encoder bytes; real Postgres correctly exposed that `jsonb` canonicalizes
its textual representation, making semantic equality weaker than R1's byte-identity requirement.
The database's committed representation is now the one authority at all three write sites.

Import cannot replace the initial run-1 genesis without violating immutability, and
`save_streams` intentionally has a permanent owner/scope uniqueness constraint. The resulting R1
implementation archives the untouched initial Founder and both streams, then creates a new active
imported Founder whose normalized Company state is revision 1/genesis under the current hash.
Account-authenticated operations already resolve the active Founder, so existing access tokens
continue to work; the imported exclusion marker remains relational and permanent.

The Postgres matrix proves account/import/Exit byte identity, genesis update rejection, and
transaction rollback when faults fire after the new-run pin or after the genesis insert. The
direct repository-owned integration command is green: `docker compose -f compose.save-test.yml
run --rm test`. Plan items 3 and 4 are complete; the six-verdict verifier is next.

## 2026-08-01 — independent review: TS port + genesis storage (0dc005f, f51f6a9, 6c9cdcd)

**Verdict: genesis storage is APPROVED — the strongest schema work yet** (deferred constraint
trigger makes a pin uncommittable without genesis; genesis bytes via `RETURNING state::text` so
byte-identity is by construction — the log honestly records the first draft failing against jsonb
canonicalization; three-step fault loop proves no partial state; all four write sites
single-transaction). **The TS port is NOT approved — two HIGHs produce false verdicts on HONEST
runs, plus a band of drift findings. The port's frame must change: Go is the reference semantics,
and ANY divergence in either direction is a bug** (a stricter TS is as wrong as a looser one —
the verifier's one job is reproducing Go's verdicts).

1. **HIGH (verified first-hand, replay.ts:153) — the offer-expiry sweep only runs on cross_gate
   in TS**; Go runs the after-prestige phase on EVERY applied intent. An offer expiring during a
   manual batch produces a Go log row with `exit_offer_expired` + cleared offer state; TS replays
   no event and a stale offer → false `state_divergence`. Fix: run the after-phase on every
   applied intent, exactly as Go; fixture: offer-expires-during-manual-batch.
2. **HIGH (verified first-hand, replay.ts crossGate) — TS enforces `state.tier === from` and
   assigns `tier = to`, contradicting the RULED Go semantics (tier = max(current, gate), never an
   error for catalog-legal input).** TS hard-fails histories Go accepts (skipping/out-of-order
   gates). Fix: port `setTierFromGate` exactly incl. the `to ≤ 9` cap; fixtures: skip-ahead gate
   and lower-gate-after-higher.
3. **MEDIUM — TS has no invariant collector**: Go's buy-generator fallback/clamp paths append
   `invariant_reported` into the committed event stream; any such row fails C7 parity in TS. Port
   the collector; fixture: `count.mode=max` exercising the fallback (currently zero max-mode
   coverage).
4. **MEDIUM — rejection-detail drift**: Go emits per-field invalid details with last-check-wins;
   TS collapses to `kind.fields` with one special case. Port the per-field details and check
   order exactly; fixtures per field.
5. **MEDIUM — clock-regression check placement**: Go checks before both arms (rejections
   included); TS only inside evaluate. Match Go. LOW batch (same rule — Go is reference, both
   directions): missing-key zero-fill vs exactKeys; UUID-pattern strictness TS-only;
   founder-revision equality check TS-only; null-carry cross-gate hard-error vs silent skip.
6. **MEDIUM (test methodology) — "byte identity" is canonical-JSON equality after a JS number
   round-trip**: key ORDER never asserted, >2^53 masking possible (currently saved only by
   MaxExactInteger == MAX_SAFE_INTEGER). Fix: compare raw fixture strings against
   TS-serialized-canonical output without a parse round-trip on the expected side.
7. **Genesis LOW:** `Store.PinRunToCurrentEpoch` derives genesis from the LATEST revision (a
   future-caller trap violating R1's first(run) definition) — restrict to revision 1 or assert;
   pre-00032 pins are permanently genesis-less (accepted: unpublished repo, rebuilt DBs —
   recorded).
8. **Import ruling A-D4b:** the redesign (archive-and-mint founder swap; genesis-preserving) is
   ACCEPTED — it is cleaner than the ruled A-D4a shape and R1 forced it. The once-only semantics
   are ruled REPLACED by **replace-while-pristine**: importing again before any intent archives
   the previous imported founder and mints another (each attempt archived, every minted founder
   `imported=true`, no ranking-integrity surface). docs/accounts-and-sessions.md's "once" claim
   must be amended to match; the per-account rate limit already bounds row litter.
9. Fixture-coverage debt recorded from the review's list (route-discounted gates, max-mode,
   compact-faction incorporation branches, terminal rejections, non-empty settlement batches) —
   lands with the verifier round alongside AC2's ≥50-intent sequential corpus (whose non-claiming
   in this round was honest and is appreciated).

## 2026-08-01 — TypeScript parity remediation after independent review

The two false-verdict HIGHs are fixed against the Go reference: prestige expiry now runs after
every applied intent before gate-only spawning, and gate transitions use the same validated
adjacent target with `tier=max(current,target)`. The ordinary shared corpus now includes an offer
expiring during a manual batch, a skip-ahead gate, and a lower gate crossed after a higher tier.

The same landing ports Go's deterministic invariant collector and detailed affordability fallback,
including a separate Go-authored artifact bundle whose legal hardcap forces the fallback and the
`invariant_reported` event. Intent parsing now preserves Go's exact per-field detail and
last-check-wins order; clock regression is checked before preflight rejection; omitted struct
scalars take Go's zero values while unknown fields still fail closed; the TS-only Founder revision
and command UUID restrictions are gone.

The fixture no longer parses expected receipts, events, or states before comparing them. Go emits
canonical JSON strings into the fixture, and TypeScript compares its canonical serialization
directly to those raw strings. The corpus is 32 ordinary cases plus 3 terminal cases and the
fallback bundle. Root checks passed: `make replay-fixture-check`, `make test-client`, and
`make test-go GO_PACKAGES=./production`.

## 2026-08-01 — public pin helper and import-contract cleanup

The public `Store.PinRunToCurrentEpoch` helper no longer selects the stream head. It selects the
earliest persisted revision whose encoded `run_seq` equals the requested run, then inserts those
bytes as genesis. The epoch integration fixture now writes two revisions for run 2 before pinning
and proves genesis equals the first run-2 revision and differs from the latest revision. Production
account/import/Exit paths continue to use the stronger explicit transaction seam.

Owner ruling A-D4b is reflected in canonical account documentation: import is replace-while-
pristine, not once-only; every replaced Founder is archived and every replacement remains excluded
from ranking.

The complete Postgres integration matrix passed through the prescribed
`docker compose -f compose.save-test.yml run --rm test` path on a fresh ephemeral schema. The first
attempt exposed a stale development tmpfs whose migration ledger said 32 was applied while its
deferred trigger was absent; resetting only that declared test container restored the committed
migration and the full matrix passed.

## 2026-08-01 — shared verifier and full-run corpus

The pure replay verifier now exists in both runtimes with the exact six-cause verdict union. It
restores genesis under the pinned artifact bundle, requires contiguous log sequence, applies
ordinary and terminal rows through the owned live transition functions, and compares canonical
receipt bytes plus ordered event-envelope bytes. Terminal completion is mandatory; a partial run
returns `log_gap`.

The shared Go-authored corpus is now a genuine full run: 50 ordinary mixed transitions followed by
terminal Exit (51 rows total), including online accrual, exact purchase, Compact join/leave,
incorporation, gate crossing, deterministic offer spawn and decline. Go and TypeScript both assert
`verified` and mutate the same corpus to cover log gap, catalog mismatch, engine drift, clock
regression, and receipt/event state divergence. `make replay-fixture-check`, `make typecheck`, and
`make test-client` pass at this landing.

## 2026-08-01 — independent review: parity fixes + verifier (c746863, f716aa7, ac2dc6f)

Two-lane review (agent adversarial pass + reviewer's direct read of verifier.go). **Verdicts:
c746863 closes all nine prior findings at the letter (each fix verified against Go line-by-line
with its named fixture) but introduces ONE new HIGH at the intersection of two fixes; f716aa7
APPROVED (genesis = earliest revision of the requested run, regression-tested, F11 caveat
recorded); ac2dc6f's verifier is the right flow in both kernels but is NOT approved — one HIGH
soundness hole plus reviewer-found fail-open behavior.**

Fix queue, ordered:

1. **HIGH (new regression, c746863) — invariant-vs-offer-sweep event ordering**: TS pushes
   `invariant_reported` AFTER the sweep; Go appends invariants inside appliedDecision BEFORE the
   after-phase — an applied buy tripping `residual_clamp` while an offer expires produces
   opposite event orders → false `state_divergence` on an honest run. No fixture combines the
   two. Fix: match Go's append site; fixture: invariant+expiry in one intent.
2. **HIGH (ac2dc6f) — rejected exit attempts are unrepresentable AND terminal handling is
   unsound**: live provably logs rejected exit commands with exit-kind inputs; Terminal:false
   rows fail the resolved-union decode (false `state_divergence`), Terminal:true rows set
   `terminal=true` even on a REJECTED outcome — subsequent rows become `log_gap` and a run ENDING
   in a rejected exit verifies as complete (**false `verified`**). Fix: terminal only on APPLIED
   outcome; rejected exit rows replay through the exit arm and continue; fixtures: wind_down-at-
   tier-0 mid-run and as final row. (This was named debt in the prior review's finding 9 —
   "lands with the verifier round" — it did not land and was not re-recorded: tracking slip,
   second occurrence of the class; debt items now carry forward EXPLICITLY in plan.md until
   closed.)
3. **MEDIUM (reviewer, first-hand) — `canonicalJSONEqual` fails OPEN on double-invalid bytes**
   (`canonical()` → nil on decode failure; `bytes.Equal(nil,nil)` is true; same for
   `marshalReplayEvents` nil) — undecodable receipt/event bytes on both sides VERIFY. Fix: any
   decode failure → verdict `state_divergence` (fail closed), negative test with corrupt bytes.
4. **MEDIUM (reviewer, first-hand) — verdict classification by error-string matching**
   (`strings.Contains(err.Error(), "clock regression")`, mirrored in TS): a reworded message
   silently reclassifies clock violations. Fix: sentinel error + errors.Is in Go; TS matches a
   shared constant.
5. **MEDIUM batch (agent, all verified with line cites):** exit-arm clock check still missing
   pre-branch in TS (half of prior finding 5; tampered terminal rejection can verify where Go
   says clock_violation); `count: null` detail divergence (`count.mode` vs `count`); TS
   input-side JS parse round-trip breaks the payload echo on number literals Go tolerates
   (`3.0`, >2^53) → false divergence on honest rejections; C6's NULL→log_gap reader clause
   STILL unimplemented while docs/production-engine.md promises it (third deferral — it lands in
   the queue round or the docs claim is retracted, one or the other); null-carry cross_gate
   silent-skip vs Go hard-error.
6. **LOW batch:** terminal_seq never cross-checked against entry sequence (R2.3's letter —
   honestly unchecked in plan); engine_mismatch is caller-supplied with no run_version_drift
   wiring while ac2dc6f's docs edit dropped the qualifier (retract or wire); corpus is 44/51
   identical manual clicks with no max-mode/invariant/expiry/rejected rows in-sequence and the
   finding-9 fixture debt (route-discounted gate, non-empty settlement batch,
   compact_tithe_raised) still absent; event-only tamper unasserted at verifier level (C7's
   silent channel!); final_state_json generated but unasserted; Go unbounded exit_history_count
   is a tampered-row memory bomb (`make([]save.ExitRecord, huge)`); f716aa7's earliest-SURVIVING-
   revision caveat recorded against future pruning.

The structural core stands: same-boundary replay in both kernels, deterministic verdict priority,
recomputed-hash bundles, genuinely-sequential corpus verifying `verified` from shared bytes with
a green drift gate. The gap between "the flow is right" and "the verdicts are sound" is exactly
findings 1–5.

## 2026-08-01 — verifier soundness remediation

The two HIGHs are closed at their seams. Invariant events are appended before the prestige
after-phase exactly like Go, and the fallback artifact case now combines `afford_fallback` with an
expired offer so the committed order is asserted directly. Exit-arm selection is derived from the
immutable resolved-input discriminator rather than a caller flag; a rejected Exit never marks the
run terminal, subsequent rows still replay, and a final rejected Exit returns `log_gap`.

Both verifiers now use typed clock violations, enforce command `run_log_seq` against the log row,
fail closed on either malformed JSON side, and assert event-only tampering. The shared corpus adds
a tier-0 rejected `wind_down` followed by another command, plus the final-rejection and terminal-
clock mutations. TS parser parity closes `count:null`, number-spelling, optional scalar, command-ID,
Founder-revision, and null-carry differences; Go bounds `exit_history_count` before allocation.

Every remaining corpus and database-reader obligation is now carried explicitly in `plan.md` per
the independent-review ruling; none is implied closed by this landing.

The persistence adapter now resolves immutable genesis, pinned engine identity, drift rows, exact
artifact bytes, run-log rows, and intent events itself. Legacy NULL replay inputs return `log_gap`;
`run_version_drift` or a pinned-version mismatch returns `engine_mismatch` without caller input.
Exit commits enqueue the ended run transactionally. The queue uses SKIP LOCKED claims with a crash
lease, requires a projector before a run can be marked verified, and dead-letters every non-
verified verdict. Category projection remains an explicit unchecked plan item because the closed
L7 category catalog does not yet exist; the queue seam does not invent one.

## 2026-08-01 — independent review: 42bc8a3 (verifier soundness round) — APPROVED

Direct diff review by the reviewer; all six verdict findings verified closed in source:

1. F2 both halves: exit rows are routed by the resolved-union DISCRIMINATOR (`kind=="exit"`),
   not the caller's Terminal flag, and `terminal = outcome == applied` — rejected exits replay
   through the exit arm and the run continues; a run ending in a rejected exit is `log_gap`
   (incomplete), never `verified`.
2. F1: the combined fixture (afford_fallback invariant + offer expiry in ONE intent) exists with
   Go's exact event order — `generator_purchased, invariant_reported, exit_offer_expired` —
   asserted cross-runtime.
3. Reviewer findings closed exactly: `ErrReplayClockViolation` sentinel + `errors.Is` in Go, a
   typed `ReplayClockViolation` class in TS (no string matching anywhere); `canonicalJSONEqual`
   returns (bytes, ok) with a trailing-value check — undecodable bytes now fail closed.
4. R2.3's letter: `command.run_log_seq` cross-checked against entry sequence → `log_gap`.
5. The tampered-row memory bomb bounded (`ExitHistoryCount ≤ MaxExactInteger`).

a3854f3 (verification queue) under separate adversarial review; verdict follows.

## 2026-08-01 — independent review: a3854f3 (verification queue) — NOT APPROVED

Adversarial review held against the receipt relay's remediated discipline; the four worst findings
re-verified first-hand by the reviewer (repository.go read directly; the package contains ONE file
and ZERO tests). The transactional skeleton is right (same-tx enqueue with the exit commit,
SKIP LOCKED claims with crash leases, NULL→log_gap lane, DB-derived drift closing F9, projection
and mark in one transaction, immutable dead letters). **The failure semantics are not — the queue
inverts the one discipline the relay rounds taught us, and it must not be composed until this
closes:**

1. **CRITICAL — transient DB failures become immutable wrong verdicts**: any artifacts/scan error
   maps to `(ReplayConstantsMismatch, nil)` — a connection blip during the catalog read
   permanently dead-letters a valid run (dead rows unclaimable, re-enqueue a PK no-op). The relay's
   deterministic-vs-transient split (failBatch's boolean) is the required shape: transient errors
   RETURN AN ERROR (release the claim, retry under lease); only deterministic evidence yields a
   verdict.
2. **CRITICAL — events() silently truncates on iteration error** (no rows.Err() after the loop):
   a mid-scan failure yields a shorter event array → state_divergence → immutable dead letter.
3. **HIGH — the events query is unscoped by stream** (`WHERE intent_id=$1`): intent ids are
   client-supplied and only per-stream unique — an adversary reusing a victim's intent_id poisons
   the victim's verification into a permanent dead letter. The codebase's own exit-replay query
   scopes by stream pair (exit.go:381); the reader must scope to the run's founder+company streams.
4. **HIGH — verdicts depend on deploy timing**: `engine != kernel.Version` dead-letters every
   queued run after any version bump. Ruling: a binary that cannot replay a run's pinned engine
   DEFERS the run (available_at pushed, attempts not spent) — `engine_mismatch` is ONLY the
   run_version_drift verdict (L2b's actual meaning). Verification is a projection of immutable
   evidence; wall-clock deploy state is not evidence.
5. **HIGH — no claim token / RowsAffected check**: a lease-expired worker double-projects (its tx
   still commits ProjectVerifiedRun after matching 0 rows on the mark), and can produce
   verified+dead-lettered simultaneously. The relay's token-guarded mark/release is the required
   pattern; the Projector contract must also state idempotency-by-run explicitly.
6. **MEDIUM batch:** attempts counted but unbounded with no failure cap, no terminal lane for
   deterministic PROJECTION poison, and no InvariantSink despite R3 requiring it verbatim; single
   `bundle.Next` slot is last-write-wins across multiple exit rows (a rejected exit's
   next-hash from a pre-mint epoch poisons replay of its own row — keep per-entry Next);
   `verification_queue`'s verified-verdict rows are mutable (house style: immutability trigger on
   the authoritative record); C6's legacy lane is production-dead code (no backfill) — fine, but
   then the lane needs its test via fixture, not via unreachable production paths.
7. LOW: global claim order allows same-company runs to project out of order (per-stream head
   claims, the relay shape, before a projector exists).

**Process finding (named, third instance of the class): plan.md flipped C6 and F9 to [x] in a
commit that ships ZERO tests for the flipped items or the package containing them.** Findings 1–5
would all have surfaced under the integration-test discipline every adjacent package follows.
New rule, effective now: **a plan checkbox may flip only in a commit that carries the test
exercising the claimed behavior** — same footing as the review-before-archive convention.

**42bc8a3 loosenings adjudicated (the scope check's three flags):** all three are the RULED
Go-parity alignments from the previous verdict — (a) the dropped re-canonicalization echo check is
finding F5's fix (JS round-trip broke honest Go-tolerated literals); (b) `count: null` coercion is
F4's fix (Go's nil-map → `count.mode` detail); (c) invariant reordering is F1's fix (Go's append
site). Approved as intended; commit messages must name their findings next time.

## 2026-08-01 — verification-queue remediation implementation

- Rebuilt the worker around token ownership: UUID claim tokens are checked under `FOR UPDATE`
  before any projector runs, and the projector plus terminal mark commit together. Claims are
  per-Company heads, so later runs cannot project around an unfinished earlier run.
- Split immutable deterministic verdict dead letters from operational poison. Transient database,
  cursor, and projector failures retry; the fifth failure (including a fifth-attempt worker crash)
  enters a separate immutable poison ledger and emits `verification_poison_dead_letter` through a
  mandatory `InvariantSink`. No operational failure is labeled with a gameplay verdict.
- Pinned-kernel mismatch now defers for five minutes and refunds the claim attempt. Only immutable
  `run_version_drift` evidence produces `engine_mismatch`.
- Scoped event reads to the run's Company/Founder stream pair, checked both `rows.Err()` paths, and
  made next-catalog selection per log entry rather than a run-wide last-write-wins slot.
- Added migration 00034 for claim-token pairing, poison evidence, and terminal queue immutability.
  The real-Postgres suite covers stale-token rejection, terminal immutability, transient retry and
  capped poison, fifth-attempt crash recovery, version deferral, deterministic catalog evidence,
  legacy NULL→`log_gap`, DB-derived drift, the cross-stream intent-ID collision attack, and forced
  cursor failure. The production suite contains the two-rejected-Exit/two-next-catalog regression.
- Green gates: focused Go packages and the root real-Postgres integration target.

## 2026-08-01 — independent review: queue remediation (873ec9c, 2599a11) — APPROVED

Direct review, all seven findings verified closed in source with a NAMED integration test each
(nine tests, real Postgres): token-owned claims with RowsAffected assertions and terminal
immutability; the transient/deterministic split via sentinel errors (transient → release + retry
under lease, capped attempts → poison dead-letter with InvariantSink per R3's letter); version
skew DEFERS without spending an attempt (`engine_mismatch` now means exactly a run_version_drift
row); events scoped to the run's stream pair with `rows.Err()` fail-closed at all three read
sites; per-entry Next catalogs; per-company in-order claiming; legacy NULL and drift as
database-verdict tests. The checkbox convention was honored — every flipped box ships its test.
2599a11 adds the projection-poison and ordering proofs. The L7 projector remains
contract-blocked; the missing predicates are the OWNER's gap, closed below in the Leaderboards RFC
(L7a).

## 2026-08-01 — L7 projector, terminal facts, and archive-at-mark implementation

- `20b115b` makes `generators_purchased_total` a save-v13 exact run accumulator rather than
  inferring Low% from currently owned units. V1–v12 migration backfills from owned counts; accepted
  buys increment it in Go and TypeScript; `run_ended` schema v2 adds it plus sorted crossed gates.
  The shared verifier now returns terminal state to its parity tests and asserts the generated
  `final_state_json` in both runtimes. Repeated TS verification also caught and fixed mutation of
  recorded Founder carry inputs.
- `ff0a1d5` implements the closed L7a predicate decoder/evaluator and transaction-owned queue
  projector. One verified run can enter multiple immutable categories; Faction/Glitched come from
  stream-scoped events, Commons/Advisor from `run_ended.assisted`; imported, drifted, and pre-timer
  runs claim without time-board rows. Real Postgres proves four-category projection, exact keys and
  variable tuple, retry idempotence, pre-timer exclusion, and drift exclusion.
- Archive-at-mark writes deterministic `gzip+json.v1` bytes over pin, genesis, command inputs,
  receipts, and totally ordered Company/Founder events; hashes the compressed bytes; deletes live
  log rows only behind the archive row; and token-marks the queue in the same transaction. The
  integration proof compares rollback/retry bytes, checks live-row survival after rollback,
  compaction after commit, archive immutability, and idempotent retry.
- Green before commits: focused Go packages, 6,492 client tests, typecheck/Svelte diagnostics,
  schema verification, replay-fixture drift check, and the root real-Postgres integration suite.

### DESIGN-GAP — production L7 fact membership and epoch ownership

L7a closes the predicate SHAPES but does not enumerate the mechanical members of
`completion_set` or `forbidden_set`; the current state has no Phase-0 completion/doctrine fact
catalog and only one declared exit fact. It also does not assign category-catalog bytes to an
epoch/hash authority, so a singleton production file would reclassify queued historical runs after
a retune. Per AGENTS.md, this round commits the strict schema plus a clearly synthetic test fixture,
not invented morality/completion mechanics or an unversioned production catalog. The evaluator,
projector, terminal event, storage shape, and archive path are implemented; composition and the L7
plan checkbox remain open until the owner specifies literal set members and version ownership.

The Run Genesis acceptance item “pre-timer runs enter count boards” also remains open: all four L7a
rows explicitly select `rta` or `attended`, so there is no count-keyed canonical category to enter.
The implemented projector structurally excludes pre-timer runs from all four time boards without
fabricating a fifth category.

## 2026-08-01 — self-review follow-up: versioned TypeScript genesis migration

Save v13 activated a pre-existing validator seam: Go receives `run_genesis.version`, while the
TypeScript verifier previously assumed only the current envelope. The player verifier now accepts
the archived genesis version, explicitly migrates v12 by summing then-owned counts with the same
exact-domain saturation as Go, rejects a mislabeled v12 envelope, and retains the verdict-only API.
The shared 51-entry run proves both current-v13 and migrated-v12 genesis reach the same verified
terminal state.

## 2026-08-01 — L7b owner gaps closed

- The production category catalog is now an epoch-owned constants artifact and the projector loads
  it by each run's pinned hash. The synthetic current-catalog fallback is gone.
- Valuation supplies the exact magnitude-key destination for pre-timer runs. Real-Postgres coverage
  proves those runs enter only Valuation and no RTA/Attended board, closing both carried L7 debts in
  the same commit as their tests.

## 2026-08-01 — independent review: category projector round (20b115b..104256c)

**Verdict: the projector/archive machinery is APPROVED with evidence** (transaction-owned
multi-category projection idempotent by run_ended event_id with a claim table; variables from the
terminal payload cross-checked against a deterministic event scan; imported/drifted/pre-timer
enforced inside the projector; archive-at-mark byte-deterministic across rollback/retry with the
run_log trigger narrowed to archive-backed DELETE only; the fixture-only category rows honestly
NOT invented into balance/). **Two HIGHs block the next landing; both verified first-hand.**

1. **HIGH — cross-runtime rejection-ORDER divergence on the new purchase-total cap**: Go checks
   it before cost/affordability (intents.go:882), TS after (replay.ts:381) — a saturated-total
   run attempting an unaffordable buy produces different rejection receipts → the player
   validator calls an honest server-verified run divergent (L4's definitional parity bug, and the
   committed fixture itself reaches saturation). Fix: move the TS check to mirror Go's site;
   fixture: saturated-total + unaffordable buy.
2. **HIGH — transition semantics changed without a `kernel/VERSION` bump** (VERSION still 0.1.0,
   last touched at abcd110, while this round changed EVERY applied receipt's bytes via
   wireSnapshot + run_ended v2): the L2b deferral ruling keys on engine_version ≠ kernel.Version —
   with no bump, pre-deploy runs verified post-deploy replay under new semantics into immutable
   deterministic dead letters, the exact wrong-verdict class the queue remediation closed.
   Test-only blast radius today (unpublished), but the mechanism is broken by construction.
   **Ruling KV-1: any commit changing receipt/event/snapshot/state-encoding bytes bumps
   kernel/VERSION, enforced by a CI guard — a diff touching the kernel-affecting surface
   (production transition code, save encoding, event schemas) without a VERSION diff fails the
   build.** The guard's path list is declared in-repo and grows by RFC, like every guard.
3. MEDIUM — the TS v12→v13 genesis backfill is tested only at the trivial zero point (a one-line
   summation error ships green): add non-zero and saturated parity cases. LOW batch: v13
   missing-key asymmetry (Go zero-fills, TS errors — pointer-presence pattern per run_ended v2);
   TS genesis floor at v12 vs Go's v1 (ruling: Go-side floor at 12 — no pre-v12 genesis rows
   exist and the archive stores version explicitly; declare, don't migrate blind); imported-lane
   projector test rides the legacy suite (add the branch test); run_ended v1 rows dead-end as
   poison (accepted — nothing pre-deploy can project, moot pre-launch, recorded).

## 2026-08-01 — category-review remediation and KV-1 implementation

- Moved the TypeScript purchase-total cap check to Go's pre-cost/pre-affordability position. The
  shared Go-authored fixture now combines a saturated career total with an unaffordable purchase
  and asserts the typed `cap_exceeded/generators_purchased_total` receipt in both runtimes.
- Bumped the shared kernel identity to 0.2.0 and added the fail-closed history/worktree guard over
  the declared kernel-affecting path registry. Receipt/event/snapshot/state semantic changes can no
  longer land without the version source changing in the same commit.
- Closed the migration band: non-zero and saturated v12 backfills are asserted in TypeScript; both
  verifiers reject pre-v12 genesis; Go v13 decoding now requires explicit accumulator presence just
  like TypeScript. The imported-projector branch has its own real-Postgres fixture.
