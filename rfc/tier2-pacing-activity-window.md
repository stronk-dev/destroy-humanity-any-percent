# RFC: Tier 2 Pacing Activity Windows

- **Status:** draft — technical proposal, not implementation authority
- **Author:** Codex (proposal for Marco's adoption)
- **Created:** 2026-11-01
- **Design refs:** `design/02 §11` (Tier 2 pacing), `design/06 §idle-math`
- **Depends on:** Tier 2 Content P1/P2; existing first-hour boundary/seed contracts
- **Parent / amends:** [Tier 2 Content](tier2-content.md) P1 activity windows only
- **Planning:** `planning/tier2-content/`

## Summary

Resolve RP-698 without weakening the two-to-three-hour pacing target or changing
player mechanics. The NEW Tier 2 continuous test personas have explicit six-hour
sessions, matching their already-declared observation horizons. The old first-hour
policies and reports remain unchanged. This proposal does not implement or adopt
the held allocation policies, retune content, or repair production accounting.

## Evidence and scope

At `059e8706`, the current subset copies Chaos's sole two-hour session. All64
seeds'230400 boundaries end at7198000ms, below P2's7200000ms attended-time minimum.
A six-hour wall horizon does not extend a declared session. A finite median is
therefore below the floor; an unreached median cannot pass either. H3ai's actual
public native/Linux comparisons and independent review substantiate this bounded
constraint. The [owning log](../planning/clout-v1-and-pr-interns/log.md) preserves
the results, invalid initial assertion and exact reviewed ranges.

This amendment defines activity windows, not complete policy/artifact grammars.
Parent P1 still owns decisions, allocation, gates, exits, offers and objectives;
P2 still owns all populations, clocks, invariants and pacing criteria. Existing
Headcount/body prerequisites remain. No diagnostic subset becomes full P1/P2.

## Specification (proposed)

### W1 — Explicit new-persona activity windows

The planned `balance/testdata/t2/it-company-policy-v1.json` remains P1's single
policy authority; it is not created or served by this proposal. In its NEW rows:

Inheritance below names the current immutable `first_hour_policy.v1` registry
version3 (`balance/testdata/t0-t1/first-hour-policy-v1.json`, SHA256
`e5e5de7051beb0340e54f7013ce7d4a48c35bfcc3343220310290478445d10c3`),
not whichever future bytes happen to occupy that path.

| Policy identity | Session start | Duration | Repeat | Cadence | Wall horizon |
|---|---:|---:|---:|---:|---:|
| `chaos.t2` v1 | 0 | 21600000ms | 0 | 2000ms | 21600000ms |
| `reference.greedy` v2 | 0 | 21600000ms | 0 | 1000ms | 21600000ms |
| `casual.t2` v1 | inherited | inherited | inherited | 5000ms | 43200000ms |

Continuous personas retain zero start jitter. Casual preserves the complete
existing15-minute/30-minute session definition, independent per-session jitter,
half-open ranges and Founder-genesis wall clock. Its actual jittered durations
remain as declared, not forced to the nominal15 minutes.

These are headless test personas, not requirements that a player play for six
hours. No server tick, production rate, reward, offline cap/rate or attendance
rule changes. Reference's window is explicit too, so its new six-hour observation
does not silently inherit the same two-hour truncation; this is not evidence of
its unbuilt allocation policy's feasibility or performance.

### W2 — Session identity is not a scenario-horizon side effect

Session time is global Founder-genesis wall time and spans Company restarts;
scripted/elective endings do not restart the six-hour window. Session intervals
are half-open. With the declared horizon, Chaos has10800 potential boundaries
`0..21598000` and Reference21600 `0..21599000`, before early milestone stopping.
Both permit action boundaries at7200000 and10800000ms; neither guarantees that
the economy will reach the milestone there.

The scenario horizon clips activity; it never rewrites policy duration. A short
diagnostic horizon stops early; a horizon beyond six hours does not manufacture
more continuous activity. Extending a session again requires explicit versioned
policy authority, not a hidden runner fallback. Gate crossing remains the first
actual eligible boundary in ANY run. Unreached seeds remain visible, never dropped.

### W3 — Versioning and random streams

Use P1's NEW identities exactly: `chaos.t2` v1, `casual.t2` v1 and
`reference.greedy` v2. Copy the seed-derivation grammar as P1 requires, including
policy ID/version in the hash material. Do not secretly use the old IDs to keep
old draws. New names/versions legitimately change decision/jitter streams; this
is not a paired experiment differing only in session duration. That difference
must be disclosed when comparing old and new observations.

Do not mutate or alias `chaos.t0_t1` v1, `casual.t0_t1` v1 or
`reference.greedy` v1. The current strict schema-v1 loader remains closed; new
P1 rows require the parent's reviewed schema-v2 implementation, not relaxed v1
validation. Preserve old first-hour fixtures, hashes and historical reports.
New observations identify their actual policy/scenario/artifact bytes and reviewed
software source, at a new output path; never overwrite the old calibration as
though it measured these new personas. This amendment allocates no report schema.

### W4 — Unchanged release and measurement gates

Keep P2's `[7200000,10800000]` Founder-attended p50 envelope and all declared
seeds/horizons, must-reach/invariant/allocation/reference requirements. Keep parent
P1's gate set, exit-once rule, ignored offers and no-incorporation rule. Keep the
full nine-tier1.0 floor. This amendment does not select Headcount mechanics or an
accounting/search protocol, approve any gate literal, or waive RP-433/27AC6.

Before implementing this amendment, Marco must adopt it and authorize the parent's
contradictory literal-copy wording to be reconciled in the same normative change.
The implementing plan must separately identify remaining P1/Headcount prerequisites;
approval of this window cannot mark a partial registry or policy full/READY.
Do not advance from a window-enumeration check to a balance or release claim.
P2's complete-workload budget derivation (measured worst-case ×2, rounded and
ratified as specified there) remains required for its future full scenario;
this amendment assigns no new budget or presumed old-scenario allowance. Bounded
experiments retain their current guards/deadlines. An exhausted run is incomplete,
not a measurement from which to raise a ceiling or claim full-workload feasibility.

## Proposed parent reconciliation

On explicit adoption, retain the parent's normative body but qualify the Chaos
row's “copies the literals” with the W1 session exception, and define Reference's
six-hour session explicitly. Add W2/W3's global window, clipping and new-ID seed
rules beside the table. The author/ruling-author must authorize that edit; this
draft does not silently rewrite their accepted body. All other P1/P2 text and
Headcount holds remain; do not remove pending mechanics to land an easier subset.

## Acceptance criteria (future implementation, not achieved here)

1. The complete adopted P1 registry loads under its reviewed v2 grammar; malformed
   windows/identities refuse. Old v1 loaders, fixture bytes and results remain valid
   under their historical source; new rows cannot masquerade as old identities.
2. Actual boundary enumeration matches an independent exact reference for every
   declared seed:64 Chaos,32 Casual and1 Reference. Include time0, both pacing
   edges, session-end exclusion and shorter/longer horizon clipping controls.
3. Actual new-ID decision/jitter material is verified. An old-ID fallback fails;
   old cohorts are not relabelled as unchanged randomized pairs.
4. Real public runner journeys prove the global window spans both endings without
   restarting it, and preserve authoritative returns, full state/history and
   first eligible crossing in any run. A reset-on-ending or forced-final-horizon
   action must fail the check. A pure enumerator is not this journey.
5. Run the FULL parent P2 population and required calibration/other gates after
   their prerequisites resolve. Retain must-reach/invariant/guard failures; window
   feasibility alone does not pass pacing. No retuning or bound change here.
6. Canonical docs, exact independent review and parent/plan reconciliation precede
   feature closeout. Owner content/mint/release approval remains separate.

## Deviations from existing contracts

Proposes one explicit exception to P1's literal inheritance: new continuous
persona sessions last six hours instead of the old two. It preserves the intended
P2 horizon and target but changes measured activity and random cohorts, so owner
adoption is required. No deviation from the design pacing target is proposed.

## Open authority and engineering

Owner adoption/parent reconciliation are pending. Headcount/output/seat contracts,
complete schema-v2 policies/loader/runner, full pacing/relevance/role evidence,
report provenance implementation and balance/content adoption retain their parent
gates. Technical review is neither adoption nor proof those features exist.

## Changelog

- 2026-11-01: propose explicit new-persona activity windows for RP-698; preserve
  old inputs, new-ID randomness, pacing floor and full parent acceptance scope.
