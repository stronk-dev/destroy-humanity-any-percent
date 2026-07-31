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
