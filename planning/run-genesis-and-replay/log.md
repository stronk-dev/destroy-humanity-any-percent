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
