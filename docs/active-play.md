# Active-play opportunities and buff windows

Active play is a Company-run skill layer built on attended time. It activates only for a new run
pinned to an `opportunities` artifact and Company save version 18; older runs finish without the
mechanic. Pending opportunities, active buffs, the schedule cursor, and the spawn sequence are
replay-owned Company state and reset on Exit.

## Deterministic scheduler

The server derives each opportunity from the run identity and the declared SplitMix64 substream.
The pinned artifact owns the shaped interval sampler, minimum interval, scale, lifetime, and maximum
number of due transitions per command. Spawn evidence—including integer draws, selected effect and
generator, UUIDv7 identity, and attended-time coordinates—is frozen into replay inputs. TypeScript
replays the logged evidence rather than recomputing floating-point sampling.

Scheduling is lazy and bounded: a command may expire buffs, miss an opportunity, and spawn its
successor in one deterministic transition loop. Both legal compound orders (miss then spawn, and
spawn then immediate self-miss) are shared Go/TypeScript corpus cases. Offline time does not advance
the attended clock, so opportunities and buffs pause while the player is away.

## Effects, claims, and clicking

`claim_opportunity` is the only new client intent. The server resolves the live pending opportunity
and applies one pinned effect:

- production frenzy multiplies all generator production for a window;
- click frenzy multiplies the existing `perform_manual_batch` action during its window;
- building special derives its factor from purchased units of one logged generator;
- Lucky pays `min(bank_fraction × bank, rate_cap × rate) + epsilon`, saturated by the normal
  resource hardcap.

Clicks continue through the persisted manual-token clamp; clients never submit timestamps or
payouts. Claims schedule the next opportunity from the same deterministic stream.

## Combo hardcap

Active buff contributions enter the `event_buffs` multiplier slot in raw-byte instance order. The
effective product—including an all-generator frenzy combined with generator-specific specials—is
clamped to the pinned `combo_cap` before later multiplier slots. Saturation is visible through the
pinned `hardcap_reason_key` (the Phase-0 fixture uses `cap.active_combo`) in the claim receipt and
the affected event payloads. The shared saturation vector uses a raw product of 77 against a cap of
10, proving that both runtimes execute the clamp rather than merely accepting the schema.

## Event and database contracts

Schedule-only events remain schema v1:

- `opportunity_spawned.v1` records identity, attended spawn/expiry coordinates, effect, and optional
  selected generator;
- `opportunity_expired.v1` and `buff_expired.v1` record identity and the attended transition time.

Buff claims use schema v2 so cap visibility is exact:

- `opportunity_claimed.v1` adds `cap_reason_key` to the buff-result arm; Lucky claims retain their
  exact requested/credited delta, saturation flag, and resource-cap reason;
- `buff_started.v1` adds `hardcap_reason_key` alongside the buff identity, target, and attended
  activation/expiry coordinates.

Migration `00067_active_play_event_schema_v2.sql` extends the live Postgres constraint for exactly
those two v2 kinds while preserving `run_ended` v2 and every other kind's v1-only rule. The
integration suite writes a real buff claim through service, store, and Postgres and asserts both v2
rows, so the database contract is exercised rather than inferred from unit validation.

Live and replay transitions compare state, receipts, and ordered event bytes. The Company-v18 codec,
artifact activation, and Exit reset are enforced in both runtimes. Epoch 7 pins
`balance/opportunities/t0-t1.json`; new runs on that epoch activate the mechanic while older runs
continue against their immutable accepted bundle.
