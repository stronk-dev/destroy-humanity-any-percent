# RFC: Arcade Pinball (a cover-disc pinball table for the Demo Disc Arcade)

- **Status:** draft — not implementation authority
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/03 §8` (Demo Disc Arcade: era-authentic toys, the cover-disc stage, the
  legal rule "mechanics yes, trade dress and names never"), `design/03` header (T0–1 minigames are
  tutorial-tier and free), `design/03 §Clock taxonomy` (session skill), `design/03 §12b` (scaling
  seam and Fairness Law), `design/06` (stack: "small canvas/Pixi layer only for particles +
  minigame boards"), `design/08 §1` (voice rules), `design/08 §2` (era 1995 presentation),
  `design/11 §2` and `§7` (one hint per system; accessibility baseline), `design/00` (pillars and
  anti-goals). **Pinball itself is not in any design document** — see Deviations D-P1.
- **Depends on:** Demo Disc Arcade (accepted 2026-09-25, implementing; AR1–AR8 and its owner
  decisions OD-1–OD-16 as ruled); Minigame Platform Foundation (accepted, implementing); Minigame &
  Recovery API + Surface (accepted, implementing; MA3 surface); The Pitch (archived; the parity
  template); UI Foundation and Game UI Screens (archived); Copy Pipeline (archived). It consumes the
  draft Accessibility of Player Workflows RFC as its floor if that RFC is accepted first.
- **Parent / amends:** successor/extension of `rfc/minigame-demo-disc-arcade.md`. It adds **one
  tenant** (`pinball` 1.0.0), **one pinned content artifact** (`balance/arcade-pinball.json`), and
  **six named amendments** (PB-A1–PB-A6) to the arcade RFC's container, composition, and surface
  rules. None changes a ruled platform grammar.
- **Supersedes / superseded by:** —
- **Planning:** `planning/minigame-arcade-pinball/` (once accepted)

## Summary

The owner asked for pinball after accepting the Demo Disc Arcade. It belongs on the arcade's
1995 `cover_disc` stage: the single bundled pinball table was the era's default desktop toy. This
RFC specifies that toy as a third arcade tenant, built exactly the way `snake` is built:

- a pure, deterministic engine over `(content, commands)`, with **integer fixed-point physics
  only** (no floating point anywhere in the rules path), so Go and TypeScript are byte-identical
  like The Pitch;
- the client runs the shared TypeScript engine locally and sends **batched `advance` commands**
  carrying per-tick flipper and plunger bits; the server replays them under a bounded per-command
  tick budget;
- table geometry is **data** in a pinned, strictly-typed artifact;
- the economy stance, Soul gate, unlock, exclusivity, and local-only best score are the arcade's,
  unchanged: zero payout, inert quality, neutral rating, `human_hobby`, `always`.

The table is our own. No bundled-OS table name, art, layout, or sound is used or imitated.

## Motivation

The arcade RFC ships a deduction grid and a Snake. Pinball adds the one real-time physics toy the
1995 desktop is remembered for, and it is the owner's explicit request. It also exercises the
platform on its heaviest deterministic workload so far: a continuous-looking simulation with a long
command log. The Snake pattern (local shared engine, batched input schedule, server replay) already
answers the transport and authority questions. This RFC adds the physics contract that makes that
pattern byte-safe.

**Out of scope:** tilt and nudge (OD-P3); multiball, upper flippers, ramps, spinners, drop targets,
kickbacks, ball save, extra balls, and modes; spin, friction, and rolling physics; audio; any
payout; boards, ghosts, daily tables; achievement hooks (arcade DG-2); later-era tables (arcade
DG-4); a second table.

## Specification

### PB1 — Placement in the arcade

**PB1.1 Tenant and definition.** Pinball is platform minigame `arcade.pinball`, engine `pinball`
version `1.0.0`, mode `solo`. It is listed on the `cover_disc` stage (PB-A1). Everything in arcade
AR1.4–AR1.10 applies unchanged:

- unlock `{"kind":"always"}`;
- `soul_gate: "human_hobby"` (lock only; no restoration — arcade D1/DG-5);
- zero-rate faucet, inert offline quality, neutral rating (`rating_delta: null`);
- no play cooldown;
- one active minigame per Founder, and the MA-C12 Exit rejection, mitigated by `quit` (PB4.4) and
  the arcade's `arcade.exit_blocked` Desk copy;
- `fallback: {"kind":"solo"}`;
- optional local-only best score (arcade OD-11).

**PB1.2 Definition row** (added to the pinned `minigames` artifact, schema v3; same shape as the
arcade rows, only these literals differ):

```json
{
  "minigame_id": "arcade.pinball",
  "engine_ref": "pinball",
  "engine_version": "1.0.0",
  "modes": ["solo"],
  "result_score_fact_ids": ["pinball.out_of_bounds_losses", "pinball.score", "pinball.ticks_played"],
  "scaling": {"schema_version": 1, "scaling_inputs": [
    {"destination": "minigame.arcade.pinball", "destination_class": "breadth", "source_kind": "literal",
     "source_ref": "1", "op": "identity", "operand": 0, "clamp_min": 1, "clamp_max": 1}]},
  "payout": {"credited_resource_id": "company.cash", "sends_per_day": 0, "per_send_cap": 0,
             "conversion_ppm": 0, "payout_score_fact_id": "pinball.score",
             "cap_reason_key": "cap.minigame_faucet"},
  "fallback": {"kind": "solo"},
  "offline_quality": {"score_fact": "pinball.score",
    "grade_curve": [{"score_threshold": 0, "grade_ppm": 200000}],
    "decay_grid_ms": 3600000, "decay_ppm_per_grid": 0, "neutral_floor_ppm": 200000,
    "automation_destination": "minigame.arcade.pinball"},
  "rating_policy": {"starting_elo": 1000, "elo_floor": 0, "elo_ceiling": 3000,
                    "provisional_games": 10, "season_member": "s1"},
  "unlock_condition": {"kind": "always"},
  "soul_gate": "human_hobby"
}
```

Rows sort byte-wise: `arcade.mine_grid` < `arcade.pinball` < `arcade.snake` < `pitch`. The only
scaling input is literal breadth 1, so the Fairness Law holds trivially.

**PB1.3 No randomness.** The engine draws no random numbers. The platform still freezes a session
seed, and the engine ignores it. Identical content and identical inputs produce identical games.
There is therefore no hidden information: the whole table state is public in every snapshot.

### PB2 — Numeric model (normative; the parity contract)

**PB2.1 Scale.** One fixed-point scale, `Q = 4096` (Q12), is used for every fractional quantity:

| Quantity | Representation |
|---|---|
| Position, length, radius | integer Q12 table units (`1 unit = 4096`) |
| Velocity | Q12 units per **substep** |
| Gravity, kick, launch velocity | Q12 units per substep (gravity: per substep, added each substep) |
| Unit normal component | integer in `[-4096, 4096]` |
| Restitution | integer in `[0, 4096]` (`4096` = 1.0) |

The y axis points **down** (screen convention). There is no other scale and no float.

**PB2.2 Integer kernel.** Exactly three helpers, identical in both runtimes:

- `tdiv(a, b)`: integer division **truncating toward zero** (Go's native `/` on `int64`). `b` is
  never zero at any call site (PB3 guarantees it; the engine fails closed as a tenant divergence
  if it would be).
- `isqrt(n)`: for `n ≥ 0`, the unique integer `r` with `r² ≤ n < (r+1)²`. Any algorithm is allowed
  because the result is mathematically defined; `Math.sqrt`/`math.Sqrt` on floats is **not**
  allowed.
- `clamp(v, lo, hi)`.

**PB2.3 Runtime integer types.** Go uses `int64`. The TypeScript engine uses **`BigInt` for every
arithmetic operation** in the rules path and converts to `Number` only when it writes a snapshot.
Every snapshot value is proven within `±(2^53 − 1)` by PB2.4, and the snapshot validator rejects
anything else. (Rationale: `Number` products in PB3.3 reach `2^61`, beyond exact double range;
BigInt removes the whole class of parity bugs for a cost of roughly 10⁵ operations per second of
play.)

**PB2.4 Magnitude table (normative; the loader bounds in PB5.3 guarantee it).** With the loader
limits (table ≤ 1024 × 2048 units, segment length ≤ 256 units, radii ≤ 64 units,
`max_speed ≤ ball_radius / 2`):

| Intermediate | Bound |
|---|---|
| Ball coordinate (including the one substep before out-of-bounds detection) | `< 2^24` |
| Segment direction component `d` | `≤ 2^20` |
| `dd = d·d` | `≤ 2^41` |
| `w` component after the AABB prefilter (PB3.3) | `< 2^21` |
| `wd = w·d` in the interior case | `< dd ≤ 2^41` |
| `d.x · wd` | `< 2^61` |
| Flipper surface-velocity product `Δtip · wd` | `< 2^57` |
| `dist2` | `< 2^43` |
| Velocity component (after clamp) | `≤ 2^16` |
| `vr · n` | `< 2^30` |
| Score | `≤ score_cap ≤ 999,999,999` |

Every intermediate stays below `2^62`, so Go `int64` cannot overflow. A test build asserts each
bound (AC-P2).

### PB3 — Physics

**PB3.1 Bodies.**

- **Ball.** One ball: centre `(x, y)`, velocity `(vx, vy)`, radius `ball_radius`. No spin.
- **Walls.** Capsules: a segment `A→B` with radius `r ≥ 0` (a thin wall has `r = 0`). Optional
  `kick` (a slingshot is a wall with `kick > 0`), `score`, and `group_id`.
- **Bumpers.** Circles: centre `C`, radius `r > 0`, with `kick`, `score`, `group_id`.
- **Flippers.** Exactly two capsules from a fixed `pivot` to `pivot + tip_path[step]`, radius `r`.
  `step` is an integer in `[0, len(tip_path) − 1]`; step 0 is rest, the last step is fully raised.
  Rotation is represented only by this pinned table of tip offsets — there is no trigonometry in
  the engine. Each flipper binds one input bit.
- **Sensors.** Axis-aligned inclusive boxes `x0 ≤ x ≤ x1, y0 ≤ y ≤ y1` over the ball **centre**,
  with `score` and `group_id`. Sensors never collide.
- **Zones.** `shooter` and `drain`, the same box shape, with rule roles (PB3.6, PB4.2).

**PB3.2 The substep.** Each tick runs `substeps_per_tick` substeps. One substep, in this exact
order:

1. **Flippers** (in byte order of `flipper_id`): `prev = step`. If the flipper's input bit is held,
   `step = min(step + 1, last)`; otherwise `step = max(step − 1, 0)`.
2. **Gravity:** `vy += gravity`. Then clamp both velocity components to `±max_speed`.
3. **Remember** `pre = (x, y)` for sensor edges.
4. **Integrate:** `x += vx; y += vy`.
5. **Collide**, one pass each, in this class order and byte order of id within a class: all walls,
   then both flippers, then all bumpers (PB3.3).
6. **Sensors** (byte order of `sensor_id`): a sensor **triggers** when the post-collision centre
   is inside and `pre` was outside. A trigger scores and lights (PB4.3).
7. **Loss check:** the ball is lost if its centre is inside `drain`, or if it is out of bounds
   (`x < 0`, `x > width`, `y < 0`, or `y > height`). An out-of-bounds loss also increments
   `out_of_bounds_losses` (PB4.5). Loss handling is PB4.2.

**PB3.3 Collision against a capsule or circle** (one collider `K`, radius `rk`; `R = ball_radius +
rk`):

1. **Prefilter (required; it enforces PB2.4).** For a segment, skip `K` unless
   `min(ax,bx) − R ≤ x ≤ max(ax,bx) + R` and the same for `y`. For a circle, skip unless
   `|x − cx| ≤ R` and `|y − cy| ≤ R`. The prefilter never changes a result — a ball outside the
   expanded box cannot be within `R` — it only bounds the arithmetic.
2. **Closest point `P`.** For a circle, `P = C`. For a segment: `d = B − A`, `w = ball − A`,
   `dd = d·d`, `wd = w·d`. If `wd ≤ 0`, `P = A`; else if `wd ≥ dd`, `P = B`; else
   `P = (ax + tdiv(dx·wd, dd), ay + tdiv(dy·wd, dd))`.
3. **Contact test.** `diff = ball − P`, `dist2 = diff·diff`. No contact if `dist2 ≥ R²`.
4. **Normal `n` (Q12).** If `dist2 = 0`, use the fallback: for a segment
   `n = (tdiv(−dy·4096, L), tdiv(dx·4096, L))` with `L = isqrt(dd)`; for a circle `n = (0, −4096)`.
   Otherwise `dist = isqrt(dist2)` (≥ 1) and `n = (tdiv(diff.x·4096, dist), tdiv(diff.y·4096,
   dist))`, `pen = R − dist`.
5. **Push-out (position first):** `x += tdiv(pen·n.x, 4096); y += tdiv(pen·n.y, 4096)`. With the
   fallback normal, `pen = R`.
6. **Surface velocity `vs`.** Zero for walls and bumpers. For a flipper whose `step ≠ prev` this
   substep: `Δtip = tip_path[step] − tip_path[prev]`, `wdc = clamp(wd, 0, dd)`, and
   `vs = (tdiv(Δtip.x·wdc, dd), tdiv(Δtip.y·wdc, dd))`. (Linear-in-radius approximation of the
   rotating surface.)
7. **Impulse.** `vr = v − vs`; `vn = tdiv(vr.x·n.x + vr.y·n.y, 4096)`. Only if `vn < 0`
   (approaching):
   - `j = tdiv(−vn · (4096 + restitution), 4096)`; `vx += tdiv(j·n.x, 4096)`;
     `vy += tdiv(j·n.y, 4096)`.
   - It is a **hit** if `−vn ≥ min_hit_speed`. On a hit: the collider scores and lights (PB4.3),
     and if `kick > 0`: `vx += tdiv(kick·n.x, 4096)`, `vy += tdiv(kick·n.y, 4096)`.
8. **Clamp** both velocity components to `±max_speed`.

There is no tangential friction and no spin. Resting contact (for example a cradled ball) is the
repeated no-hit case: push-out plus an impulse below the hit threshold.

**PB3.4 Tunnelling bound.** The loader requires `2·max_speed ≤ ball_radius` (so the ball's centre
moves at most `0.71·ball_radius` per substep) and that consecutive `tip_path` entries differ by at
most `ball_radius / 4` (checked as `16·|Δtip|² ≤ ball_radius²`). The relative motion of ball and
any collider surface in one substep is then below `R`, so the centre cannot cross a collider
without a detected overlap. Pushes from resolving one collider can still, in a tight corner, move
the ball through another. The out-of-bounds loss is the failsafe, and it is **counted and
reported, never silent** (PB4.5, AC-P6).

**PB3.5 The tick.** For tick `t = tick + 1`:

1. **Input.** `bits_t` = the bits of the input event at `t` if one exists, else `held`.
2. **Plunger** (bit 2, value 4), evaluated before `held` is updated:
   - Ball "in shooter" means its centre is inside the `shooter` zone.
   - If bit 2 is set in `bits_t` and the ball is in shooter: `plunger_charge =
     min(plunger_charge + 1, charge_max_ticks)`. If bit 2 is set and the ball is not in shooter:
     `plunger_charge = 0`.
   - If bit 2 was set in `held` and is clear in `bits_t` (a release): if the ball is in shooter
     and `plunger_charge > 0`, set `vx = tdiv(launch_vx · plunger_charge, charge_max_ticks)` and
     `vy = tdiv(launch_vy · plunger_charge, charge_max_ticks)`. In every release case,
     `plunger_charge = 0`.
3. `held = bits_t`.
4. Run the substeps (PB3.2). A terminal loss (PB4.2) stops the tick after the substep in which it
   happens.
5. If still playing and `t = max_game_ticks`, the game ends `time_limit` (PB4.4).
6. `tick = t`.

**PB3.6 Shooter lane.** The ball spawns at `(ball_start_x, ball_start_y)` with zero velocity,
which the loader requires to lie inside `shooter`. It rests on the shooter floor, an ordinary wall
(typically restitution 0). The plunger is not a body; it is the impulse in PB3.5. A ball that falls
back into the shooter lane can be relaunched.

### PB4 — Rules

**PB4.1 Balls.** `balls_per_game` balls (v1: 3), played one at a time. `ball_number` starts at 1.

**PB4.2 Ball loss.** On loss (PB3.2 step 7): if `ball_number < balls_per_game`, increment
`ball_number`, respawn the ball at the start with zero velocity, set `plunger_charge = 0`, and
continue the remaining substeps of the tick. Otherwise the game ends `game_over` at this tick.

**PB4.3 Scoring and groups.** A collider hit or a sensor trigger with `score > 0` adds its score.
All additions saturate: `score = min(score + points, score_cap)`. A saturated score is a visible
hardcap (`cap.arcade_pinball_score`, shown whenever `score = score_cap`).

If the element has a `group_id` and is not in `lit`, it is added to `lit`. When every member of
that group is in `lit`, the group's `bonus` is added (saturating) and every member of the group is
removed from `lit`. Member points are added before the bonus. Flippers never score or light.

**PB4.4 Terminal.** Outcomes, closed: `game_over | quit | time_limit`.

- `game_over`: the last ball is lost.
- `time_limit`: tick `max_game_ticks` completes (v1: 108,000 ticks = 30 minutes at the
  presentation rate). This is a visible hardcap, not a hidden timeout (OD-P6).
- `quit`: the `quit` command, from any non-terminal state.

The terminal transition writes `phase: "terminal"` and emits the result: facts, byte-sorted,
`pinball.out_of_bounds_losses`, `pinball.score`, `pinball.ticks_played` (= `tick`), and
`rating_delta: null`.

**PB4.5 The out-of-bounds counter is an instrument.** `out_of_bounds_losses` is a snapshot field
and a result fact so that a geometry defect is visible data rather than a quiet ball drain
(evidence discipline 3). The v1 table must produce zero across the AC-P6 fuzz corpus.

### PB5 — Content artifact `balance/arcade-pinball.json`

**PB5.1 Identity and chain (PB-A2).** A new pinned artifact, `schema_version: 1`, joins
`CatalogBundle` as `CatalogBundle.ArcadePinball` under the Pitch/arcade precedent:

- the canonical hash uses the established constants-hash construction;
- session content identity lives in exact genesis keys `{arcade_pinball_content_hash,
  arcade_pinball_schema_version}` (no database column);
- artifact chain: `arcade_pinball` requires `arcade` (which requires `minigame_api`, and so on). It
  is added to `validArtifactNames` with the same hard-fail discipline.

It is a separate artifact, not a key in `arcade.json`, so that retuning the table does not change
the content hash of `mine_grid`/`snake` sessions (OD-P2).

**PB5.2 Grammar (exact keys at every level; JSON numbers are exact safe integers only; `null` is
allowed only for `group_id`).**

```
{
  schema_version: 1,
  rules:   {balls_per_game, score_cap, max_game_ticks, max_ticks_per_advance,
            client_flush_ticks, presentation_tick_hz},
  physics: {substeps_per_tick, gravity_q12, max_speed_q12, min_hit_speed_q12, ball_radius_q12},
  table:   {width_q12, height_q12, ball_start_x, ball_start_y,
            shooter: {x0, y0, x1, y1}, drain: {x0, y0, x1, y1},
            launch_vx_q12, launch_vy_q12, charge_max_ticks,
            walls:    [{wall_id, ax, ay, bx, by, radius_q12, restitution_q12, kick_q12, score, group_id}],
            bumpers:  [{bumper_id, cx, cy, radius_q12, restitution_q12, kick_q12, score, group_id}],
            flippers: [{flipper_id, input_bit, pivot_x, pivot_y, radius_q12, restitution_q12,
                        tip_path: [[dx, dy], ...]}],
            sensors:  [{sensor_id, x0, y0, x1, y1, score, group_id}]},
  groups:  [{group_id, bonus, copy_key}]
}
```

All coordinates are Q12.

**PB5.3 Loader rules (both runtimes, hard-fail).**

- **Ids.** `wall_id`, `bumper_id`, `flipper_id`, and `sensor_id` match `^[a-z0-9_.]{1,48}$` and are
  unique **across all four lists**. Each list is byte-sorted by id. `groups` is byte-sorted by
  `group_id` (same pattern).
- **Table.** `1 ≤ width ≤ 1024·4096`, `1 ≤ height ≤ 2048·4096`.
- **Physics.** `1 ≤ substeps_per_tick ≤ 16`; `4096 ≤ ball_radius ≤ 32·4096`;
  `1 ≤ max_speed` and `2·max_speed ≤ ball_radius`; `0 ≤ gravity ≤ max_speed`;
  `1 ≤ min_hit_speed ≤ max_speed`.
- **Elements.** Every coordinate lies inside `[0, width] × [0, height]`. Wall and flipper radius
  `0 ≤ r ≤ 64·4096`; bumper radius `4096 ≤ r ≤ 64·4096`. `0 ≤ restitution ≤ 4096`;
  `0 ≤ kick ≤ max_speed`; `0 ≤ score ≤ 1,000,000`. A wall has `0 < |B − A|² ≤ (256·4096)²`.
- **Flippers.** Exactly two; `input_bit` values are exactly `{0, 1}`. `2 ≤ len(tip_path) ≤ 64`.
  Each tip satisfies `0 < |tip|² ≤ (256·4096)²` and `pivot + tip` lies inside the table.
  Consecutive tips satisfy `16·|Δtip|² ≤ ball_radius²`.
- **Zones and launch.** Each zone has `x0 < x1`, `y0 < y1`, and lies inside the table. The ball
  start lies inside `shooter`. `|launch_v*| ≤ max_speed`; `1 ≤ charge_max_ticks ≤ 600`.
- **Groups.** Every non-null `group_id` names a `groups` row. Every group has at least two members.
  `0 ≤ bonus ≤ 10,000,000`. `copy_key` must exist in the copy catalog (composition check).
- **Rules.** `1 ≤ balls_per_game ≤ 9`; `1 ≤ score_cap ≤ 999,999,999`;
  `600 ≤ max_game_ticks ≤ 216,000`; `1 ≤ max_ticks_per_advance ≤ 600`;
  `1 ≤ client_flush_ticks ≤ max_ticks_per_advance`; `30 ≤ presentation_tick_hz ≤ 120`.
- `presentation_tick_hz` and `client_flush_ticks` are **presentation-only**: the Go engine
  validates their ranges and never reads them.

**PB5.4 The v1 table (provisional content, OD-P4).** Coordinates below are in **units**; the
artifact stores `units × 4096`. The table is a plain single-level layout with no theme art (OD-P13
names the theme).

- **Rules:** 3 balls; `score_cap` 999,999,999; `max_game_ticks` 108,000;
  `max_ticks_per_advance` 240; `client_flush_ticks` 120; `presentation_tick_hz` 60.
- **Physics:** 8 substeps; gravity 28 (Q12/substep²); `max_speed` 16,384 (4 units/substep);
  `min_hit_speed` 1,024; ball radius 8 units.
- **Table:** 400 × 800 units. Ball start (378, 760). `shooter` (364, 600)–(392, 769);
  `drain` (20, 745)–(356, 800). Launch (0, −16,384); `charge_max_ticks` 45.
- **Walls** (radius 2, restitution 2,048, no kick, no score unless stated):
  - arch: (394,64)→(352,20)→(280,6)→(120,6)→(52,30)→(20,90) as five segments;
  - left wall: (20,90)→(20,340), (20,340)→(20,590);
  - aprons: (20,590)→(134,690) and (362,590)→(266,690);
  - shooter divider: (362,210)→(362,450), (362,450)→(362,590), (362,590)→(362,770);
  - shooter outer wall: (394,64)→(394,300), (394,300)→(394,540), (394,540)→(394,770);
  - shooter floor: (362,770)→(394,770), restitution 0;
  - slingshots: kick faces (64,560)→(112,630) and (336,560)→(288,630) with kick 6,144 and score
    10; backs (64,560)→(64,610)→(112,630) and (336,560)→(336,610)→(288,630);
  - standup targets (group `g.targets`, score 500 each): (26,250)→(26,275), (26,285)→(26,310),
    (26,320)→(26,345).
- **Bumpers** (radius 18, restitution 3,277, kick 8,192, score 100): (150,200), (240,180),
  (200,270).
- **Flippers** (radius 6, restitution 1,229, length 50): pivots (140,690) and (260,690). Each
  `tip_path` has 28 entries spaced evenly in angle from 30° below horizontal (rest, pointing
  toward the centre line) to 25° above. Entries are generated offline as
  `round(50·4096·cos θ), round(50·4096·sin θ)` (x negated for the right flipper) and **pinned as
  integers**; the generator is authoring convenience, and the engine reads only the integers.
  Step spacing is about 1.78 units, inside the 2-unit bound. The rest-position gap between the
  flipper capsules is about 21 units, wider than the 16-unit ball, so the ball can drain.
- **Sensors** (group `g.lanes`, score 1,000 each): top lanes (100,40)–(130,70),
  (180,40)–(210,70), (260,40)–(290,70).
- **Groups:** `g.lanes` bonus 5,000; `g.targets` bonus 10,000.

These literals are starting values, retuned by playtest (AC-P6 checks only geometry integrity).

### PB6 — Engine `pinball` 1.0.0: snapshot and commands

**PB6.1 Snapshot (exact sixteen keys).**

```
{phase: "playing"|"terminal", tick, ball_number, ball_x, ball_y, ball_vx, ball_vy, held,
 plunger_charge, flipper_steps: [{flipper_id, step}], lit: [id], score, out_of_bounds_losses,
 revision, arcade_pinball_content_hash, arcade_pinball_schema_version}
```

- Genesis: `phase: "playing"`, `tick: 0`, `ball_number: 1`, ball at the start with zero
  velocity, `held: 0`, `plunger_charge: 0`, both flipper steps 0, `lit: []`, `score: 0`,
  `out_of_bounds_losses: 0`.
- `flipper_steps` is byte-sorted by `flipper_id`; `lit` is byte-sorted.
- The snapshot fully determines the future given the inputs. Nothing else is stored. There is no
  hidden field (PB1.3).

**PB6.2 Input bits.** `bits` is an integer in `[0, 7]`: bit 0 (1) and bit 1 (2) are the flippers
whose `input_bit` is 0 and 1; bit 2 (4) is the plunger. An input event at tick `t` sets the held
bits **from tick `t` onward**, until the next event. The encoding is an edge list, so a command
carries at most one entry per tick and never repeats the held state.

**PB6.3 Commands (closed union, discriminator `kind`).**

- **`advance {from_tick, ticks, inputs: [{tick, bits}]}`**, legal only in `playing`. Validation,
  in this order (the first failure wins, and nothing mutates):
  1. `from_tick = tick`, else `stale_tick`;
  2. `1 ≤ ticks ≤ max_ticks_per_advance`, else `advance_window`;
  3. for each input, in list order: `0 ≤ bits ≤ 7`, else `invalid_bits`; `from_tick < input.tick
     ≤ from_tick + ticks`, else `advance_window`; strictly ascending ticks, else
     `inputs_not_ascending`; `bits` differs from the bits held just before that tick, else
     `inputs_redundant`;
  4. simulate. If the game becomes terminal at tick `d < from_tick + ticks`, reject
     `advance_past_terminal` **without mutation**. An honest client, simulating locally, submits
     exactly up to the terminal tick; the server never discards input silently (the Snake rule).
- **`quit {}`**, legal in `playing`: terminal `quit`.

Any command in `terminal` rejects `illegal_phase`. Every applied command advances `revision` by
exactly one.

**PB6.4 Rejection taxonomy (sorted).** `advance_past_terminal | advance_window | illegal_phase |
inputs_not_ascending | inputs_redundant | invalid_bits | stale_tick`.

**PB6.5 Cost bounds.**

- **Per command:** at most `max_ticks_per_advance × substeps_per_tick` substeps (v1: 1,920), each
  touching at most every collider once.
- **Wire:** at most `ticks` input entries; at v1's 240 that is under 8 KiB, well inside the 64 KiB
  API body limit.
- **Rate:** the client flushes every 120 ticks (2 s at 60 Hz), so about 30 commands per minute,
  inside the shared 300-per-minute account limiter.
- **Resolution replay:** the platform replays the whole command log at resolution. The worst case
  is a `time_limit` game: 108,000 ticks, 864,000 substeps, about 900 command rows. That cost is
  **measured, not assumed** (AC-P9); `max_game_ticks` is ratified from the measurement (OD-P6).

**PB6.6 Wall-time trust.** As arcade AR4.8: the server certifies an input schedule, not the real
time it took. A scripted client can play a perfect or slow-motion game, and because there is no
RNG a recorded input file replays exactly. With zero payout, no boards, and no ghosts this changes
nothing (OD-P10). Any future board, ghost, or payout on pinball needs a server-stamped pacing
contract as a platform successor.

### PB7 — Platform and composition amendments (named, explicit)

- **PB-A1 — arcade container.** The `cover_disc` stage's `toys` becomes
  `["arcade.mine_grid","arcade.pinball","arcade.snake"]`. The AR1.2 composition check allows
  `engine_ref ∈ {mine_grid, pinball, snake}`. Because `arcade.json` is still fixture-only (arcade
  AR8), this changes pre-mint bytes, not a minted artifact.
- **PB-A2 — `CatalogBundle.ArcadePinball`.** Artifact name `arcade_pinball`, path
  `balance/arcade-pinball.json`, the chain in PB5.1, and a Go and TypeScript loader that pass one
  shared fixture.
- **PB-A3 — content resolver.** The AR-P2 closed table gains `(pinball, 1.0.0) → arcade_pinball`.
  The AR-P3 start precondition then covers pinball without further change.
- **PB-A4 — `minigame_api`.** The tenant rows gain `{pinball, 1.0.0, arcade.pinball}`. The
  generated API adds one snapshot arm and one command arm. The closed rejection-detail enum adds
  the PB6.4 details under the same `{status, category, detail}` mapping. All additive under the v1
  compatibility baseline.
- **PB-A5 — canvas for the pinball board (amends arcade AR6.1 "DOM only, no canvas").**
  `design/06` allows canvas for minigame boards. The pinball child may render its **board only** on
  a single `<canvas>`. Everything else — controls, status, settings, live region — stays DOM. The
  client boundary verifier gains a positive rule: `<canvas>` and `getContext` are allowed only
  under the pinball child's directory, and colour/duration literals stay banned there. Canvas
  colours come from `--cc-*` tokens read with `getComputedStyle` on mount and on every era or
  theme change.
- **PB-A6 — copy denylist.** The trade names and table names of the era's bundled pinball product
  and its makers are added to `moderation/copy-denylist.txt`. Each term needs its legal-section
  reference in the tracked extracts file, which is owner-authored (DG-P3).

### PB8 — UI surface contract (the `pinball` child)

**PB8.1 Layout.** A focusable board region (`tabindex="0"`, labelled by
`arcade.pinball.board.label`, described by the controls help) containing the `<canvas>`
(`aria-hidden="true"`), and beside or below it a DOM status panel: score (with the cap text when
saturated), "ball N of 3", plunger charge as text and a native `<progress>`, and the lit members
per group as text. The board scales to its container at a fixed aspect ratio. At a 320 px viewport
the 400 × 800 table renders at 288 × 576 CSS px without horizontal page scroll (A3.1).

**PB8.2 Controls.**

- **Keyboard defaults:** left flipper `Z` or `ArrowLeft`; right flipper `/` or `ArrowRight`;
  plunger `Space` or `ArrowDown` (hold to charge, release to launch). Keys act **only while the
  board region has focus**, only without modifiers, and `preventDefault` applies only to bound
  keys. Browser shortcuts are never intercepted and Tab is never trapped (A2.2).
- **Remapping (OD-P8):** a local setting remaps the three actions. It rejects conflicts and
  modifier combinations and has a reset control. It is stored per viewer in `localStorage` inside
  `try/catch`, and the page works without it.
- **One-switch mode (OD-P8):** a local setting where one key does everything. On press, if the
  local predictor shows the ball in the shooter zone, the key holds the plunger bit; otherwise it
  holds both flipper bits. The choice is made at press time and held until release. It is purely a
  client input mapping; the server sees ordinary bits.
- **Pointer and touch:** three visible buttons (left, launch, right), at least 44 CSS px, held via
  `pointerdown` and released on `pointerup`, `pointercancel`, or lost pointer capture.
- **Pause:** `P`/`Esc` and a visible Pause button. The game also auto-pauses on
  `visibilitychange`→hidden, window blur, and focus leaving the board region. Pausing releases all
  inputs: on resume, if `held ≠ 0`, the client emits `bits: 0` at the first resumed tick. Pausing
  is fair by construction (the engine reads no clock).
- **Pace (OD-P8):** a local, labelled setting — 100% / 75% / 50% — that scales only the
  presentation tick rate. It is never tied to reduced motion (A4.2: motion preferences never alter
  sim rate).

**PB8.3 Local simulation and transport.** The client runs the shared TypeScript engine module
(the same one the corpus tests) at `presentation_tick_hz` in its own fixed-timestep loop,
separate from the 20 Hz idle sim. It draws the latest tick on `requestAnimationFrame` without
interpolation. It flushes `advance` every `client_flush_ticks`, immediately on terminal, and on
pause. It keeps at most one command in flight (MA-C8) and freezes the local simulation if its
unacknowledged lead reaches `max_ticks_per_advance`. A rejected or diverged flush triggers
`current` and a resync to the server snapshot, with the `arcade.resync.notice` status. The server
snapshot is the only truth.

**PB8.4 Screen readers and the timed-input exception.** Pinball is a real-time visual game. Like
Snake (arcade D7) it is an **optional toy with a declared timed-input exception**, mitigated by
pause, pace, one-switch mode, and remapping. The untimed `mine_grid` remains the accessible arcade
path. The polite live region announces only: new ball, ball lost, group bonus (at most one
announcement per 2 s, merged), and the outcome (always).

**PB8.5 Motion, CRT, and photosensitivity.** The ball and flippers move; that is the game's
content, like Snake's steps. Everything decorative is **off in every mode** in the v1 eras, because
`era_1995` has a zero-motion budget and `era_2000` shares the cover-disc stage:

- no ball trail or motion blur, no screen shake, no bumper pulse or scale animation, no lamp
  blinking, no score roll-up;
- lit elements are drawn statically, as a filled shape against an outlined unlit one, so state is
  never hue-only (A3.2);
- no content flashes more than 3 times in any 1-second period (WCAG 2.3.1), and there are no
  full-field red flashes;
- reduced motion (`prefers-reduced-motion`, via the single `MediaQueryList` path, A4.1) is
  therefore already satisfied by v1. The same path is wired so that any later-era decoration is
  gated by it. A CSS-only path fails AC-P12;
- CRT treatment is the arcade's token bezel only (arcade OD-8).

**PB8.6 Local best score.** Arcade OD-11 applies: a local-only, labelled display record that never
feeds an intent, receipt, or board.

**PB8.7 Copy keys (mechanical; all prose owner-authored and pending, OD-P13).**

- `arcade.toy.pinball.{title,description}`
- `arcade.pinball.board.label`, `arcade.pinball.controls.help`
- `arcade.pinball.control.{left,right,launch}`
- `arcade.pinball.status.{score,ball,charge,lit}` (params `score`, `ball`, `balls`, `percent`,
  `group`, `count`, `total`)
- `arcade.pinball.settings.{remap,remap_reset,remap_conflict,one_switch,pace}`
- `arcade.pinball.pace.{full,three_quarter,half}`
- `arcade.pinball.announce.{new_ball,ball_lost,group_bonus}`
- `arcade.pinball.group.{lanes,targets}` (the group `copy_key`s)
- `arcade.pinball.best_local`
- `arcade.outcome.{game_over,time_limit}` (`quit` exists)
- `cap.arcade_pinball_score`

T0 voice follows `design/11 §1b`: the earnest webmaster, warm and never mocking. The toy name and
all table text must avoid the era product's names and trade dress (`design/03 §8` legal rule).

### PB9 — Verification corpus (the TP-C25 / AR7 pattern)

A versioned pinball corpus is checked in against a **fixture artifact** with small tables that
make each edge reachable in few ticks. Each scenario is an exact `(content, commands, expected
snapshot or result)`. Go and TypeScript byte-compare every snapshot and result. The fixed
transition budget equals the sum of corpus command counts. Content changes regenerate the corpus in
a reviewable balance-change commit.

Scenarios:

- **Kernel:** `tdiv` across all four sign combinations and at `±2^61`; `isqrt` at 0, 1, 2, 3,
  every perfect square `k²` and `k² − 1` for `k` near `2^21.5`, and at `2^62`.
- **Collision:** a head-on wall bounce; a glancing hit below `min_hit_speed` (no score); a hit at
  exactly `min_hit_speed` (scores); endpoint (capsule cap) contact; the `dist2 = 0` fallback for a
  segment and for a circle; bumper kick; slingshot kick; a moving-flipper strike (nonzero `vs`); a
  cradled ball held for 600 ticks.
- **Rules:** plunger charge to max and release; release at partial charge; release with the ball
  outside the shooter; charge reset when the ball leaves the shooter while held; drain → ball 2;
  third drain → `game_over` mid-window; out-of-bounds loss (count incremented); sensor re-trigger
  only after exit; a group completing (bonus plus reset); `score_cap` saturation; `time_limit` on
  a fixture with `max_game_ticks` 600; `quit`.
- **Commands:** every PB6.4 code, including validation-order cases where two codes apply.

## Deviations from design

- **D-P1 — Pinball is not in `design/03`.** It is an owner-requested addition (2026-09-25), not
  derived from any design text. Until the owner amends design, this RFC is the only statement of
  intent. Proposed line for the owner to add under `design/03 §8`'s first bullet (owner-authored
  design text; this RFC does not edit it):

  > - **Cover-disc pinball (owner addition 2026-09-25):** one bundled-desktop-style pinball table
  >   on the 1995 stage — two flippers, a plunger, bumpers, lanes, three balls. Mechanics yes;
  >   the bundled table's name, art, layout, and sounds never.

- **D-P2 — Canvas board.** Consistent with `design/06`, but it amends the accepted arcade AR6.1
  "DOM only, no canvas" (PB-A5).
- **D-P3 — Timed-input exception.** It conflicts with draft Accessibility A2.2 ("no timed key
  sequence") and is declared an optional-toy exception with mitigations (PB8.4), as arcade D7 did
  for Snake.
- **D-P4 — Board loop rate.** The board runs its own 60 Hz presentation loop, not `design/06`'s
  20 Hz idle-sim loop. The engine itself is clockless; Snake's 125 ms tick is the precedent.
- **D-P5 — Inherited arcade deviations** D1 (no Soul restoration), D2 (no play cooldown), D3 (no
  achievement hooks), and D5 (token bezel only) apply unchanged.

## Design gaps (DESIGN-GAP — not improvised)

- **DG-P1 — Design intent for pinball.** Only D-P1's proposed line would give it one. Owner.
- **DG-P2 — Table theme and art direction.** Nothing in design says what our table depicts. v1
  ships a plain, untextured geometric table. A themed table (a garage-startup table would suit the
  satire) is owner-authored content in a successor.
- **DG-P3 — Denylist provenance.** PB-A6's denylist terms each need a legal-section reference in
  the tracked extracts file (`docs/copy-pipeline.md`). That extract does not exist and is
  owner/research-authored.
- **DG-P4 — Per-tenant replay cost.** The platform replays the whole command log inside the payout
  transaction and has no per-tenant cost contract. Pinball is the first tenant where that matters.
  This RFC measures it (AC-P9) and bounds it through `max_game_ticks`. A general contract (for
  example a replay budget declared on the tenant descriptor) is a platform successor.
- **DG-P5 — Accessibility ruling on optional timed toys.** Snake and pinball both rely on an
  "optional toy" exception that the draft Accessibility RFC does not define. That RFC, or its
  successor, should rule on it.
- **DG-P6 — Tilt and nudge.** Deferred (OD-P3). If adopted later: a nudge input bit, a warning
  counter, a tilt that disables flippers until the ball drains, all as data. Not specified here.

## Acceptance criteria (each has a demonstrated failing case)

1. **AC-P1 Artifact loader.** Go and TypeScript accept the fixture and the v1 artifact
   byte-identically. **Fails on**, in both runtimes: unsorted walls; an id duplicated across walls
   and sensors; a wall longer than 256 units; `2·max_speed > ball_radius`; a `tip_path` step
   greater than `ball_radius/4`; three flippers or duplicate `input_bit`; a one-member group; an
   unknown `group_id`; a zone outside the table; a ball start outside `shooter`; an extra key; and
   the literals `1.5` and `1e3`.
2. **AC-P2 Integer kernel and bounds.** The shared kernel vectors pass in both runtimes, and a
   test build asserts every PB2.4 bound over the whole corpus and the AC-P6 fuzz run. **Fails
   on:** a TypeScript `tdiv` written as `Math.trunc(Number(a)/Number(b))` (red at `2^55`); a floor
   division mutant (red on negative operands); an `isqrt` via `Math.sqrt` (red above `2^52`); a
   corpus mutant that removes the prefilter, which trips the bound assertion.
3. **AC-P3 Physics parity.** The PB9 corpus is byte-identical across Go and TypeScript, and the
   client predictor imports the same TypeScript module. **Fails on** each seeded mutant:
   integrate before gravity; walls and bumpers collided in swapped class order; velocity impulse
   applied before push-out; restitution applied to the tangential component; a missing velocity
   clamp; a flipper with zero surface velocity; kick applied below `min_hit_speed`.
4. **AC-P4 Rules.** Every PB9 rules scenario matches its expected snapshot. **Fails on:** a
   plunger that launches outside the shooter; a sensor that re-triggers while the ball stays
   inside; a group bonus that does not reset `lit`; a non-saturating score; a `game_over` after
   ball 2.
5. **AC-P5 Rejection without mutation.** Every PB6.4 code is reached; revision, command rows, and
   snapshot are unchanged; validation-order cases return the first code in PB6.3 order. **Fails
   on:** a mutant that applies a partial `advance` before rejecting `advance_past_terminal`, and a
   mutant that silently drops inputs after the terminal tick.
6. **AC-P6 Geometry integrity.** A seeded fuzz of at least 1,000 v1-table games with random input
   schedules (the seed list is checked in) reports `pinball.out_of_bounds_losses = 0` in every game
   and hits every sensor, bumper, and target at least once. **Fails on:** a v1-table mutant with a
   gap in the arch, which produces nonzero out-of-bounds losses and turns the gate red.
7. **AC-P7 Definition and economy.** The row loads under schema v3. The arcade policy test (every
   arcade row has `conversion_ppm = 0` and an inert quality curve) covers `arcade.pinball`.
   **Fails on:** a fixture with nonzero conversion.
8. **AC-P8 Amendments.** The resolver returns `arcade_pinball` bytes for exactly `(pinball,
   1.0.0)`; the other arcade pairs and Pitch are unchanged; the stage manifest lists
   `arcade.pinball`. **Fails on:** removing `arcade_pinball` from a bundle makes a pinball start
   reject, and an `arcade_pinball` artifact without `arcade` fails the chain.
9. **AC-P9 Composed path and replay cost.** Over a real socket: create, a scripted game of
   `advance` commands to `game_over`, auto-resolution returning credited `"0"`, no cap reason,
   unchanged rating and quality, and an identical retry. A second test resolves a
   `max_game_ticks` game and records the resolution replay time on the reference CI runner as a
   visible artifact field. OD-P6 is ratified from that number. **Fails on:** a replay that
   diverges from the stored terminal snapshot (a seeded one-bit input change in a stored command
   row) fails closed before payout.
10. **AC-P10 Exit and Soul.** With an active pinball session, Wind Down and `cross_gate` reject
    `minigame_session_active`, and after `quit` both succeed. Near-zero Soul rejects the start
    `human_content_locked`. **Fails on:** classifying the toy `unrelated` passes the Soul start.
11. **AC-P11 API.** The generated schema passes the additive-only baseline, and the details map
    exactly. **Fails on:** renaming an existing arm breaks `make api-check`.
12. **AC-P12 Surface and accessibility.** axe WCAG 2.2 AA shows zero serious/critical findings in
    Chromium, Firefox, and WebKit for pinball playing, paused, and terminal. Keyboard-only tests:
    launch the ball, operate both flippers, pause/resume, and quit; a one-switch test launches and
    flips with one key; a remap test rebinds the left flipper and rejects a conflict; blur and
    hidden-tab auto-pause release all bits. Reduced-motion tests assert live media-query switching
    through the single JS path. **Fails on:** a CSS-only reduced-motion mutant; a ball-trail
    mutant; a hue-only lit state; a key handler active without board focus; a handler that
    intercepts `Ctrl+W`.
13. **AC-P13 Canvas boundary and photosensitivity.** The client boundary verifier allows
    `<canvas>` only under the pinball child and rejects colour and duration literals there. No
    pinball code defines an animation in `era_1995`. **Fails on:** a seeded `<canvas>` in
    `mine_grid`, a hex colour literal in the pinball renderer, and a seeded lamp-blink loop.
14. **AC-P14 Copy.** Every PB8.7 key resolves with owner-adopted text, and the denylist rejects
    the era product's names. **Fails on:** a missing key, and a seeded denylisted table name.
15. **AC-P15 Fixture-first boundary.** The live epoch manifest does not pin `arcade_pinball`
    until the mint (OD-P14). **Fails on:** a test that pins it without the mint record.

## Owner decisions (numbered; recommended defaults)

1. **OD-P1 — Adopt pinball.** *Recommended:* add pinball to the `cover_disc` roster and adopt the
   D-P1 design line. Alternative: park it in `design/BACKLOG.md §Nostalgia arcade sweep`.
2. **OD-P2 — Artifact placement.** *Recommended:* a separate `balance/arcade-pinball.json`.
   Alternative: a `pinball` key in `arcade.json` (one hash for all three toys, so any table retune
   changes `mine_grid`/`snake` session identity).
3. **OD-P3 — Tilt.** *Recommended:* no nudge and no tilt in v1; a tilt without a nudge input has
   nothing to punish (DG-P6). Alternative: specify nudge plus tilt now (adds an input bit, a
   warning counter, and flipper-disable rules).
4. **OD-P4 — Physics and table literals.** *Recommended:* the PB5.4 values, all provisional and
   retuned by playtest. Alternative: the owner supplies a layout.
5. **OD-P5 — Ball rules.** *Recommended:* 3 balls; no ball save, extra balls, or multiball.
6. **OD-P6 — Game-length hardcap.** *Recommended:* 108,000 ticks (30 min), visible outcome
   `time_limit`, ratified from the AC-P9 replay measurement. Alternative: a lower cap if the
   measured replay cost is too high.
7. **OD-P7 — Score cap.** *Recommended:* 999,999,999 with visible `cap.arcade_pinball_score`.
8. **OD-P8 — Accessibility package.** *Recommended:* optional toy with the timed-input exception,
   plus pause (always), pace 100/75/50%, a one-switch mode, and local key remapping. Alternatives:
   drop one-switch or remapping; or defer pinball until the Accessibility RFC rules (DG-P5).
9. **OD-P9 — Canvas.** *Recommended:* adopt PB-A5 (canvas board only, token colours, scoped
   verifier allowance). Alternative: a DOM board with absolutely positioned elements (possible,
   but it contradicts `design/06`'s intent for minigame boards and costs layout work each tick).
10. **OD-P10 — Wall-time and TAS play.** *Recommended:* accept it while payout, boards, and ghosts
    are all absent (as arcade OD-6).
11. **OD-P11 — Local best score.** *Recommended:* yes, local-only and labelled (arcade OD-11).
12. **OD-P12 — Transport cadence.** *Recommended:* 60 Hz presentation, flush every 120 ticks, at
    most 240 ticks per `advance`. All provisional.
13. **OD-P13 — Name, theme, and copy.** The owner names the toy, chooses the table theme (DG-P2),
    and authors every PB8.7 string. No trade names or trade dress.
14. **OD-P14 — Delivery.** *Recommended:* fixture-first, like the arcade; the production mint of
    `arcade-pinball.json`, the definition row, and the `minigame_api` bytes rides in the arcade
    mint epoch if pinball is ready, otherwise a later owner-gated epoch.
15. **OD-P15 — Audio.** *Recommended:* none in v1, as the arcade.

Work slices, each reviewed separately: (1) kernel and loader (AC-P1, AC-P2); (2) engine and corpus
(AC-P3–AC-P6); (3) PB-A1–PB-A4 and the composed path (AC-P7–AC-P11); (4) the surface after MA3
(AC-P12–AC-P14).

## Open questions

- Should upper flippers, drop targets, or multiball ever come? They would be an engine `1.1.0` or
  `2.0.0` successor with new artifact grammar, not a table retune.
- Should the Flash-portal stage (T2–3) carry a second table? That is arcade DG-4's successor.
- Arcade AR-F3 (should the scripted first ending auto-quit an active session?) applies to pinball
  as well; it is the arcade's open question, not re-decided here.

## Changelog

- 2026-09-25: created (draft) at the owner's request ("where's pinball?") as a successor to the
  accepted Demo Disc Arcade RFC.
