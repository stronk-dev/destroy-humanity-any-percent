# T0–T1 Playable Content

Tiers 0 and 1 are live, measured content. Epoch 7 (`T0-T1 Playable Content`) pins the economy,
Routes, categories, relevance policy, opportunities, presentation, copy, and harness evidence at
`sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789`.
Epoch 8 (`First-Hour Payoff`) adds the curriculum artifact and pins the complete nineteen-artifact
bundle at `sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`.
Earlier epoch identities remain immutable.

## First-hour progression

The canonical scenario is `balance/testdata/t0-t1/harness-scenario-v1.json`; its policy authority is
`balance/testdata/t0-t1/first-hour-policy-v1.json`. It runs 64 Chaos seeds, 32 Casual seeds, and one
projected-time Reference seed over a two-hour horizon. All policies use the production transition,
cross `gate.t0_to_t1` through its real requirements, continue into run 2, and Exit through the
shipped prestige path.

The seven required milestones are first manual action, first generator, first upgrade, the run-1
Garage gate, the scripted first-company ending, the run-2 Garage gate, and the first later elective
Exit. The governed envelopes require the Chaos median to buy a generator within 30 seconds, buy an
upgrade within two minutes, cross into the Garage in six to eight minutes, and hit the scripted
ending in fourteen to sixteen minutes. The first elective Exit must land between 45 and 90 minutes
for Chaos p50 and Casual p95. At every Chaos seed, run 2 must reach the Garage gate strictly sooner
than run 1.

Casual sessions are Founder-genesis wall-clock windows whose gaps use the canonical offline-catchup
policy before online evaluation. This exercises the same session boundary that prevents a Founder
idle beyond the 24-hour accrual cap from becoming permanently unplayable.

## Scripted first-company ending

`balance/curriculum/t0-t1.json` owns the first-company curriculum. On Founder run 1, after the
Founder has crossed `gate.t0_to_t1` and reached 900,000 attended milliseconds, the first player
Company command after accrual is replaced atomically by a `scripted_first` terminal transition.
The Founder Exit count makes the transition once-per-Founder; a New Founder starts a fresh
lifecycle.

The ending and run-2 starter depend on how run 1 was played:

- `acquihire`: at least 200 purchased generators and one owned upgrade; run 2 receives `1e4`
  `company.cash`.
- `burnout`: after the acquihire check, cash is below twice the cheapest next generator price; run
  2 receives ten generated `generator.beige_tower` units and 50 Route Knowledge. Generated units
  are not purchased and do not earn purchased-count multipliers.
- `pivot`: the remaining branch, where cash is at least twice the cheapest next generator price;
  run 2 receives pre-owned `upgrade.reply_all_macro` and 25 Route Knowledge.

The terminal event is `run_ended` schema v3 and includes the selected `branch` and exact
`starter_package`. Go and TypeScript apply the same starter semantics during live execution and
replay. The old run retains its pinned epoch; the new run uses the current epoch without rewriting
history.

## Evidence and gates

- `make first-hour-harness HARNESS_WORKERS=12` runs the complete 97-run pacing suite and enforces
  the seven milestones, envelopes, same-seed run-2 relation, replay parity, resource bounds, role
  activation, and artifact identity.
- `make t0-t1-relevance-all` runs the two phase-scoped relevance reports, the generator-role
  activation/control matrix, and branch proofs. Every measured purchasable passes on the whole
  path or its identity-bound branch proof.
- `make epoch7-content-harness` verifies the minted T0–T1 artifact composition and content-dynamics
  evidence.
- `TestComposedGameserverReplaysRatifiedFirstHourAtPinnedSeed` replays the ratified Casual seed
  through the real gameserver and Postgres, checks all seven milestones and both terminal runs,
  verifies Founder history, and waits for durable replay verification and immutable archival.
- `make copy-check`, strict artifact loaders, epoch-history guards, and cross-runtime replay tests
  cover copy completeness, exact bytes, immutable epoch identity, and `run_ended` v3 parity.

The detailed design and review trail is frozen in
`rfc/archive/t0-t1-playable-content.md` and `planning/archive/t0-t1-content/`.
