# Prestige & Exits lifecycle audit

Coordinate: product tree `190a4fa`; audit checkpoint after `79939b0`; 2026-08-20.

Post-coordinate reconciliation (2026-08-21): owner-delegated edits reconciled D2–D5/P3–P5,
`design/11`, the Prestige plan, and canonical docs to shipped behavior and D-012. RP-034–RP-036 are
closed. The original acceptance classifications below are the frozen audit basis. Post-coordinate
batch 2b has now added cold-green, mutation-proven AC2–AC6 witnesses; RP-037 is remediated pending
designated review and RP-038 remains open.

This pass re-derived the active Prestige RFC from its specification, plan, append-only log,
current implementation, successor T0–T1 curriculum, canonical docs, tests, and historical review
records. It did not edit product code, owner-authored RFC text, balance data, or plan checkboxes.

## Bottom line

The Exit transaction is substantial working software. Atomic Founder/Company persistence,
idempotent replay, exact Go/TypeScript arithmetic, current run-end rendering, and the current
first-hour path all execute successfully. The RFC is nevertheless not archival-ready:

- at the audit coordinate, five acceptance criteria retained narrower witnesses than their literal
  requirements (remediated after the coordinate; designated review pending);
- Advisor Mode's stored/mechanical seams remain, while D-012 now explicitly defers its player
  command/settings surface from the preview;
- the normative body and canonical Prestige doc have been reconciled after this audit coordinate;
- the Quarter-to-offer bridge remains explicitly deferred and P3 now names only the live site; and
- no tracked verdict supplies the explicit cross-party provenance and exact full-range union now
  required for archival.

The T0–T1 designated reviews prove the successor curriculum and first-hour implementation. They do
not silently become a designated review of the older Prestige foundation.

## Current cold evidence

All commands ran from the repository root with cold Go counts where applicable:

- `make test-go GO_PACKAGES='./prestige ./production ./save' GO_TEST_FLAGS='-count=1'` — green;
- `make test-client` — 39 files passed, two skipped; 6,655 tests passed, 15 skipped;
- `make test-save-integration SAVE_TEST_PACKAGES='./production ./save ./gameserver'` — green on
  real PostgreSQL, including the composed first-hour integration;
- `make first-hour-harness` with the owner-ratified payoff tuple
  `A=200, B=2e0, R=50, S=1e4, G=10` — 97/97 runs completed, zero warnings, zero failures.

The regenerated first-hour identity is scenario `sha256:18a6f16a...f85ba`, policy
`sha256:e5e5de70...d10c3`, constants `sha256:baa89050...28f6`. Chaos first elective Exit p50 and
Casual p95 are both 2,700,000 ms. All 64 Chaos observations equal 2,700,000 ms; the 32 Casual
observations range from 2,700,000 to 2,703,413 ms.

That is valid evidence for the governed policy's accepted interpretation: a persistent-value Exit
is available by the first action boundary at or after 45 attended minutes and remains inside the
90-minute ceiling. It is not evidence about unaided human choice timing. The lower bound is a
policy precondition; the upper bound and persistent-value predicate retain the content-sensitive
failure path. The generated report was an ignored local measurement artifact and was removed after
the values were recorded here.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Cold witness green; archival review missing** | The real-Postgres store test injects failures at eight core write boundaries plus five log/genesis/pin boundaries and verifies neither stream or related record commits. Success and idempotent recorded-order replay also execute. | Preserve the current negative cases; include this path in the full designated range review. |
| AC2 | **Post-coordinate witness green; designated review pending** | Four non-expired offer ages (`0`, `1`, half-life, last-live ms) progress then accept at the same boundary and dominate stored terms; exact expiry rejects without Founder advancement. Zeroing `PromiseTerms` fails all live rows. | Include the exact test/fixture/record span in the designated review. |
| AC3 | **Post-coordinate witness green; designated review pending** | The non-clamped Notoriety interval is exhaustive, 10,006 deterministic values span the clamped domain/endpoints, and invalid bounds reject against an independent big-integer reference. Two different real Exit kinds preserve three cumulative non-empty fact keys. Off-by-one reseed and removed fact-union mutations fail. | Include both pure and Postgres witnesses in the designated review. |
| AC4 | **Post-coordinate witness green; designated review pending** | The composed GameServer test completes an old-Founder Exit, uses the real New-Founder endpoint, confirms archival/distinct identity/empty history, reaches the current command boundary, and receives exactly one fresh `scripted_first`. Disabling zero-history trigger detection leaves history empty and fails. | Include Account endpoint, production trigger, and test span in the designated review. |
| AC5 | **Post-coordinate witness green; designated review pending** | A declared six-row real-Postgres matrix proves Wind Down for plain, pending opportunity, active buff, incorporated, and live-offer states; Tier 0 with prior history is the typed nonmutating control. An active-event trap fails the pending/buff rows. Existing pre-timer/cross-epoch witnesses remain green. | Include the matrix and existing boundary witnesses in the designated review. |
| AC6 | **Post-coordinate witness green; designated review pending** | Two complete encodings match a checked-in run-2 golden from one declared Founder/prior-Company input. In-test corruptions and a production Guild-boundary carry severing both differ from the golden. | Include the golden and producer in the designated review. |
| AC7 | **Cold witness green; archival review missing** | Strict run-ended validators, exact client decoding, a type-level no-snapshot boundary, and the Run End component test collectively cover the event-only consumer; current copy/schema checks are green from the preceding lifecycle batch. | Include producer, schema, copy slots, decoder, and surface in the full designated range review. |
| AC8 | **Proven under the owner-approved policy, with scope qualification** | Current 97-run measurement is green; the real-Postgres composed script reaches scripted run 1 and elective run 2 and has a recorded failing mutation. The lower edge is intentionally by construction, so the evidence proves content availability under the governed persona, not human behavioral timing. T0–T1's exact implementation ranges have designated approval. | Reconcile Prestige's plan/record to the successor evidence without changing the bound or overstating human behavior. |

No Prestige plan box was changed. The checkbox law requires the exercising test and record flip to
sit in the same future designated-review range; audit inference is not such a range.

## Normative and canonical drift

1. D2 names `reputation_levels` and `expires_at`; the implemented strict terms/event contracts use
   `reputation_delta` and millisecond fields such as `expires_at_ms`.
2. D3 step 5 and P4 still call retained save revisions the old run's archive/source of record.
   P4b rules that `run_ended` plus the run log are the obituary authority and retention stays at
   five, but the ruling author never reconciled the earlier normative text.
3. D4/P5 and `docs/prestige-and-exits.md` still say the scripted ending occurs at a threshold
   crossing. The shipped, epoch-pinned curriculum replaces the first player Company command after
   accrual once run 1, the Garage gate, zero exits, and 900,000 attended ms are true.
4. P3 still declares Quarter harvest as a live offer evaluation site. An inline Fiscal note says
   the Founder-to-Company bridge is deferred, but the original normative sentence remains intact
   and the bridge does not exist.
5. D5 promises a Founder-scoped Advisor settings toggle and first-enable Assisted behavior. Search
   of server intents, client controls, and tests finds only persisted `advisor_mode`, multiplier
   math, replay fields, and run-event labeling. The Prestige log has carried this DESIGN-GAP since
   2026-07-29 without an owner ruling or successor.

## Review and range truth

The foundational implementation was `8e269a7`, `09e7363`, `ef6e6f7`, `c3c58b9`, and the
Leaderboards integration `ce0a5ec`; review remediation landed in `0a7f129` and `4589a64`.
Historical log entries describe adversarial reviews of the original lifecycle and `0a7f129`, but
they do not name `Review by:` or `Recorded by:` and cite pre-rewrite hashes. The only text that
describes a diff review covering the final P5c/P2d remediation is
`planning/archive/codex-fixes-2026-07-30.md`; that file is now a tracked, explicitly noncanonical
historical queue, but it still lacks explicit reviewer provenance and therefore does not satisfy
the current review gate. The tracked Prestige plan correctly leaves “Record independent review
before archival” open.

Under the current two-gate law, Prestige needs a new tracked designated cross-party adversarial
verdict whose cited post-rewrite ranges union every implementation and remediation commit it
claims to approve. A retrospective label may not be inferred from prose or Git authorship.

## Smallest honest closeout order

1. **Complete after this coordinate:** body/docs reconciliation and D-012 successor routing.
2. **Complete after this coordinate:** missing literal AC2–AC6 witnesses with demonstrated failures.
3. Reconcile the remaining AC8 plan evidence to the already-reviewed T0–T1 successor behavior.
4. Obtain the mandatory tracked cross-party full-range verdict.
5. Only then perform status/docs/plan/log/archive closeout transactionally in the reviewed range.
