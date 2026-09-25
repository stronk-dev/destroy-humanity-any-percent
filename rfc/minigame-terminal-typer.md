# RFC: Terminal Typer (minigame content — Tier-1 tenant #2)

- **Status:** accepted — owner batch acceptance 2026-09-25; implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/03 §7` (Terminal Typer: type-the-command skill runs, garage-PC host,
  small click-power hook, solo/ghost fallback), `design/03` clock taxonomy (session skill,
  cooldown-gated), `design/03` intro + unlock stagger (Tier 1, tutorial-tier and FREE),
  `design/03 §12a` (dailies doctrine), `design/03 §12b` (scaling seam + Fairness Law),
  `design/02 §10` (streaks count up, never reset — supersedes the §7 streak), `design/02 §11`
  (pacing targets), `design/11 §3/§7` (Assisted precedent; accessibility baseline),
  `design/08 §1` (voice), `design/00` (pillars 6–7; "no unearned edge"),
  `design/research/rhythm-timing-games.md §4/§7` (the server-timing constraint and the assist
  precedent this RFC reuses)
- **Depends on:** Minigame Platform Foundation (accepted C1–C40, implementing — this registers a
  tenant), The Pitch (archived — the template this RFC mirrors, TP5), Minigame & Recovery API +
  Surface (accepted MA-C1–C15, implementing — the playability seam this consumes), Soul Foundation
  (archived — `soul_gate`), Prestige & Exits (implementing — the MA-C12 Exit rejection and the
  scripted first Exit), Accessibility of Player Workflows (draft — the evidence floor this surface
  must meet before release)
- **Parent / amends:** tenant on the Minigame Platform. **Names four platform/composition
  amendments** (TT-PA1–PA4); per TP5, a content RFC that needs anything outside the template
  skeleton says so explicitly.
- **Supersedes / superseded by:** —
- **Planning:** `planning/minigame-terminal-typer/` (once accepted)

## Summary

Terminal Typer is a solo, session-skill typing minigame hosted on the garage PC: the server deals
a seeded, fixed-length run of shell-command prompts from a pinned content corpus, the player types
each command and submits the line, and the server compares the submitted bytes to the target. The
certified result is a small set of integer facts (lines cleared, clean first-try lines, misses,
server-measured elapsed time); the platform — not the tenant — converts the payout fact through the
shipped faucet into `company.cash` and charges offline quality. A run is either **timed** (a
generous server-clocked budget) or **untimed** (the accessibility path), with payout recommended
identical across both, so typing speed is expression, never the only path to reward.

The one hard problem is time. A typing game's natural skill measure (speed) needs a clock, and the
only precise clock is the client's, which law 2 forbids trusting. This RFC adopts the coarse,
server-owned model the rhythm research already endorses when the window is far larger than network
jitter: the server samples its own clock once per admitted command, persists it (the command log
already has the column), and hands that exact value to the pure tenant. That requires one named
platform amendment (TT-PA1).

## Motivation

The live epochs cover T0–T1, but the only shipped tenant (The Pitch) is a Tier-5 Fiscal-unlocked
game, so live content has no minigame reachable in the tiers players actually inhabit. Terminal
Typer is the design's Tier-1 slot in the unlock stagger. It also exercises the platform surfaces
The Pitch did not: a non-Fiscal (free) unlock, a `tier`-sourced breadth scaling input, and a
server-time input.

**In scope (v1):** one tenant row; one pure engine (`server/typer` + `client/src/typer`) with Go/TS
parity; one pinned content artifact; timed and untimed solo runs; the four named amendments; the
UI surface child component and its accessibility contract; a content-gate corpus.

**Out of scope (declared successors, see DESIGN-GAPs):** ghost races / async snapshot mode;
keystroke recording; the daily-streak hook (owned by the unbuilt `02 §10` daily scaffold); the
click-power multiplier hook; achievements; personal-best persistence; a play cooldown; the Tier-3
"incantation-length k8s" corpus (content batch, lands with T3 content); any ranked or leaderboard
surface; all final prose.

## Specification

### TT1 — The tenant row (the platform's C37 grammar, filled in)

One row in the pinned `minigames` artifact. Structural IDs are literal; every integer is
**provisional balance data** (OD-10), present so the fixture loads, to be harness-tuned before a
production mint. Where a literal key drifts from the shipped schema-v3 grammar, the shipped
grammar governs 1:1 (the TP-C16 clause).

```json
{
  "minigame_id": "typer",
  "engine_ref": "typer",
  "engine_version": "1.0.0",
  "modes": ["solo"],
  "result_score_fact_ids": ["typer.assisted", "typer.clean_lines", "typer.elapsed_ms",
                            "typer.lines_cleared", "typer.misses"],
  "scaling": {"schema_version": 1, "scaling_inputs": [
    {"destination": "typer.era_tier", "destination_class": "breadth", "source_kind": "tier",
     "source_ref": "tier", "op": "identity", "operand": 0, "clamp_min": 1, "clamp_max": 9}
  ]},
  "payout": {"credited_resource_id": "company.cash", "sends_per_day": 5, "per_send_cap": 300,
             "conversion_ppm": 500000, "payout_score_fact_id": "typer.clean_lines",
             "cap_reason_key": "cap.minigame_faucet"},
  "fallback": {"kind": "solo"},
  "offline_quality": {"score_fact": "typer.clean_lines",
    "grade_curve": [{"score_threshold": 1, "grade_ppm": 200000},
                    {"score_threshold": 4, "grade_ppm": 500000},
                    {"score_threshold": 6, "grade_ppm": 800000},
                    {"score_threshold": 8, "grade_ppm": 1000000}],
    "decay_grid_ms": 3600000, "decay_ppm_per_grid": 10000, "neutral_floor_ppm": 200000,
    "automation_destination": "minigame.typer"},
  "rating_policy": {"starting_elo": 1000, "elo_floor": 0, "elo_ceiling": 3000,
                    "provisional_games": 10, "season_member": "s1"},
  "unlock_condition": {"kind": "tier_at_least", "tier": 1, "exit_history_at_least": 1},
  "soul_gate": "unrelated"
}
```

- **Fairness Law:** unranked solo; the only scaling input is `breadth` (which prompt eras are in
  the pool). `tier` is a legal source because no ranked destination exists; the loader's
  `ranked ∧ power → reject` rule is untouched.
- **Payout/quality fact:** `typer.clean_lines` (OD-2). `rating_delta` is always `null` (C40 — the
  neutral season row stays inert, exactly as Pitch).
- **`unlock_condition`:** the new `tier_at_least` arm (TT-PA2); the `exit_history_at_least` key
  exists only if OD-3 is ruled (a). **`soul_gate`:** `unrelated` per OD-5.
- **`automation_destination`:** registered by the tenant descriptor, bound 1:1 like Pitch's; no
  automated output consumes it in v1 (same status as Pitch).

Descriptor: engine `typer` `1.0.0`; command schema `typer.command.v1`; snapshot schema
`typer.snapshot.v1`; result schema `minigame.result.v1`; modes `[solo]`; destinations
`{"typer.era_tier": breadth}`; sorted error taxonomy per TT4.6.

### TT2 — Named platform/composition amendments

**TT-PA1 — Server-sampled command time reaches the tenant.** Today `minigame_session_commands`
persists `server_ts_ms`, but the value is computed by `clock_timestamp()` inside the INSERT,
*after* the tenant has already applied the command, and `ApplyInput` carries no time. Amendment:

1. The coordinator samples the database clock exactly once per admitted command,
   `floor(extract(epoch FROM clock_timestamp())*1000)::bigint`, in the same transaction that
   claims the session, before tenant execution.
2. `ApplyInput` gains `ServerTimeMs int64` (TS mirror identical). Every tenant receives it; tenants
   that do not use time (Pitch) ignore it, so their bytes are unchanged.
3. The command-append statements insert that exact sampled value as a parameter instead of
   re-sampling. Nonterminal and terminal (resolution) paths both.
4. Replay/verification passes each row's persisted `server_ts_ms` as `ServerTimeMs`. The value is
   a frozen, server-authored input — the same law as every other replay input; no tenant ever reads
   a clock. Rejected commands append nothing, so their sample is discarded.
5. The API never accepts a time field from the client (MA4 unchanged).

This is the only change that lets any tenant measure time honestly; it is reusable by later
session-skill tenants (arcade, rhythm) rather than Typer-specific.

**TT-PA2 — A tier unlock arm.** The shipped unlock union is `always | fact_equals |
fiscal_unlock`, and production start evaluates only `fiscal_unlock` (`fact_equals` is a UI shell
fact, not a server predicate; `always` would unlock Typer at Tier 0). Add exact arm
`{"kind":"tier_at_least","tier":<int 0..9>}` (plus the optional
`"exit_history_at_least":<int ≥0>` key if OD-3(a)) to both loaders. The composed start resolver
reads the pinned Company `tier` (and `len(founder.exit_history)`) server-side — never a client fact
— and rejects `not_eligible/tier_required` (or `not_eligible/curriculum_exit_required`) before
tenant creation. Go/TS loader parity; composed before/after tests.

**TT-PA3 — Pinned Typer content.** A separately hash-pinned artifact `balance/typer.json`,
`schema_version: 1`, joins `CatalogBundle` as `CatalogBundle.Typer` through the shipped
`TenantContentResolver` keyed `(constants_hash, "typer", "1.0.0")`. Content identity lives in
exact genesis/snapshot keys `{typer_content_hash, typer_schema_version}` (no DB column — the
TP-C14 precedent). The start coordinator's current hard requirement `bundle.Pitch != nil` is
generalized to "the resolver returns content for the requested tenant"; a bundle containing Pitch
but not Typer still starts Pitch. Loader chain: a `typer` definition row in `minigames` ⟺ the
`typer` artifact is pinned ⟺ the `minigame_api` tenants list contains
`{"engine_ref":"typer","engine_version":"1.0.0","minigame_id":"typer"}` — any one without the
others hard-fails load in both runtimes.

**TT-PA4 — API arms (additive, MA-C7).** The generated `minigame_api` schema gains the Typer
`1.0.0` discriminated snapshot arm and the Typer command union as request arms. No new endpoints,
no new error categories beyond the tenant taxonomy and TT-PA2's two details. The generated
tenant-surface registry gains one child keyed `(typer, 1.0.0)` (MA-C9).

No Founder or Company save-schema version changes: Typer adds no save field (ratings/quality maps
and the session sequence already exist at Founder v17/v21).

### TT3 — Content corpus (`balance/typer.json`)

Exact-key schema v1:

```json
{
  "schema_version": 1,
  "policy": {
    "run_length": 8,
    "timed_budget_ms": 120000,
    "max_line_bytes": 256,
    "elapsed_hardcap_ms": 86400000,
    "misses_hardcap": 1000000,
    "run_length_reason_key": "cap.typer_run_length",
    "line_bytes_reason_key": "cap.typer_line_bytes",
    "elapsed_reason_key": "cap.typer_elapsed",
    "misses_reason_key": "cap.typer_misses"
  },
  "eras": [
    {"era_id": "garage_2000", "min_tier": 1, "copy_key": "typer.era.garage_2000"}
  ],
  "prompts": [
    {"prompt_id": "<id>", "era_id": "garage_2000", "text": "<command>",
     "scene_copy_key": "typer.prompt.<id>.scene"}
  ]
}
```

Loader rules (Go + TS, one shared fixture):

- Exact keys everywhere; unknown/missing keys reject. All integers safe; `run_length ≥ 1`;
  `timed_budget_ms ≥ 1`; `max_line_bytes ∈ [1, 1024]`.
- `eras` sorted ascending by `min_tier`, unique `era_id` and `min_tier`; `min_tier ∈ [0, 9]`.
- `prompts` byte-sorted by `prompt_id`, unique; `prompt_id` matches the mechanical pattern;
  `era_id` references a declared era.
- `text` is 1..`max_line_bytes` bytes of printable ASCII (`0x20–0x7E`), no leading/trailing
  space, no two consecutive spaces. (ASCII-only keeps comparison byte-exact and keeps TT4.4's
  normalization a closed, one-way map.) `text` is **not** a copy key and is not localized: it is
  what must be typed (OD-9).
- **Reachability:** for every `era_tier` in the scaling row's `[clamp_min, clamp_max]`, the
  eligible pool (TT4.2) has at least `run_length` prompts; otherwise load fails. Every era is
  eligible at some reachable `era_tier`.
- Every `copy_key`, `scene_copy_key`, and reason key must resolve in the Copy catalog.

**Launch content (v1):** one era, `garage_2000`. The corpus rows are **owner-ratified content**
(OD-9). To make the fixture executable, implementation may check in a PROVISIONAL placeholder set
of at least 12 plain period-shell commands (e.g. `ls -la`, `cd /var/www/html`, `make install`);
placeholders carry no satire and must be replaced or ratified before any production mint. The
Tier-3 `cloud_2015` (k8s/YAML) era is a successor content batch; the schema needs no change to
add it.

### TT4 — Engine contract (deterministic, pure)

**TT4.1 Inputs.** `CreateInput`/`ApplyInput` as shipped plus TT-PA1's `ServerTimeMs`. The tenant
rejects any mismatch between the resolved content hash/schema version and the snapshot identity
(TP-C19 discipline). No ambient reads.

**TT4.2 Prompt order.** `era_tier = scaling_inputs["typer.era_tier"]`. The eligible pool is every
prompt whose era has `min_tier ≤ era_tier` (cumulative, OD-6), taken in byte-sorted `prompt_id`
order. `run_seed = Substream(seed, "typer.run.v1").Next()`; `rng = Substream(run_seed,
"typer.prompts.v1")`; a downward Fisher–Yates over the pool
(`for i = n−1 … 1: j = rng.Bound(i+1); swap(i, j)`, the same rejection-sampled `Bound` The Pitch's
deal uses); the run's prompts are positions `0 … run_length−1`. The order is recomputed from
`(seed, content, era_tier)` on every apply; the snapshot stores no PRNG cursor and **never
contains a future prompt**.

**TT4.3 Commands** (`typer.command.v1`, discriminator `kind`, exact keys):

| Command | Keys | Legal phase |
|---|---|---|
| `begin` | `{kind, assist_level}`; `assist_level ∈ timed \| untimed` | `ready` |
| `submit_line` | `{kind, text}`; `text` a JSON string | `typing` |
| `end_run` | `{kind}` | `ready`, `typing` |

The client never sends a score, a time, a prompt ID, or a result.

**TT4.4 Line validation and comparison.** In order, without mutation on rejection:
1. `text` must be valid UTF-8 with no C0 control characters or `U+007F` → else `invalid_text`.
2. Raw byte length ≤ `max_line_bytes` → else `line_too_long` (rejected, not truncated).
3. Normalize with the closed map of OD-7 (recommended: `U+2018`/`U+2019` → `'`,
   `U+201C`/`U+201D` → `"`, `U+00A0` → space, `U+2014` → `--`). Nothing else: no case folding, no
   Unicode normalization, no trimming.
4. **Cleared** iff normalized bytes equal the prompt `text` bytes exactly. Otherwise a **miss**
   with `first_mismatch_index` = the first byte offset where they differ, or the shorter length if
   one is a prefix of the other.

**TT4.5 Snapshot** (`typer.snapshot.v1`) — exactly eighteen keys:
`{typer_content_hash, typer_schema_version, phase, era_tier, assist_level, prompt_index,
prompts_total, current_prompt_id, current_prompt_text, current_prompt_misses, lines_cleared,
clean_lines, misses, started_server_ms, deadline_server_ms, last_server_ms, last_submission,
revision}`.
`phase ∈ ready | typing | terminal`; `assist_level`, `current_prompt_id`, `current_prompt_text`,
`started_server_ms`, `deadline_server_ms` (always null when untimed), `last_server_ms`, and
`last_submission` are nullable; `last_submission` is exact
`{outcome: cleared | miss, first_mismatch_index: int | null}` (null index on `cleared`).
Genesis: `phase: ready`, `prompt_index: 0`, `prompts_total: run_length`, counters 0, every
nullable null, `revision` per platform.

**TT4.6 Transitions.** Every applied command advances one revision. Let
`t = max(ServerTimeMs, last_server_ms ?? ServerTimeMs)` (a clock step backwards never rewinds the
run) and set `last_server_ms = t`.

- `begin`: `started_server_ms = t`; `deadline_server_ms = t + timed_budget_ms` if `timed`, else
  null; `phase: typing`; current prompt = order[0].
- `submit_line` — validation (TT4.4 steps 1–2) runs first; a rejection mutates nothing. Then, if
  timed and `t > deadline_server_ms`: the line is **not scored**, and the run ends `timed_out`.
  Else compare:
  - *cleared:* `lines_cleared += 1`; `clean_lines += 1` iff `current_prompt_misses == 0`;
    `prompt_index += 1`; `current_prompt_misses = 0`; if `prompt_index == run_length` the run ends
    `completed`, else the next prompt becomes current.
  - *miss:* `misses` and `current_prompt_misses` each `+1`, saturating at `misses_hardcap`; the
    prompt stays current (the player retries).
- `end_run`: the run ends `ended_early` (legal before `begin`, yielding all-zero facts).
- **Terminal transition** (exactly one): `phase: terminal`, `current_prompt_id/text: null`,
  counters retained, result emitted.

Sorted rejection taxonomy: `illegal_phase | invalid_assist_level | invalid_text | line_too_long`.
An unknown `kind` is `illegal_phase`, mirroring Pitch.

**TT4.7 Certified result.** `outcome ∈ completed | ended_early | timed_out`; `rating_delta: null`;
score facts byte-sorted:

| Fact | Value | Role |
|---|---|---|
| `typer.assisted` | `1` if `untimed`, else `0` (0 if never begun) | display/analytics (the `design/11 §3` Assisted precedent) |
| `typer.clean_lines` | lines cleared with zero misses on that prompt | payout + quality (OD-2) |
| `typer.elapsed_ms` | `min(last_server_ms, deadline_server_ms ?? ∞) − started_server_ms`, 0 if never begun; hardcapped at `elapsed_hardcap_ms` (`cap.typer_elapsed`) | display only |
| `typer.lines_cleared` | total cleared | display |
| `typer.misses` | saturating miss count (`cap.typer_misses`) | display |

`ValidateResult` checks the closed outcome set, the fact set, and the bounds
(`clean_lines ≤ lines_cleared ≤ run_length`, `assisted ∈ {0,1}`).

### TT5 — Timing, authority, and anti-cheat

| Threat | Handling |
|---|---|
| Client claims a score or time | Impossible by construction: the command union has no such field; the API rejects extra keys; the platform selects the payout fact from the certified result. |
| Forged timestamps | The client sends none. Time is the server's single per-command sample (TT-PA1), persisted and replayed. |
| Network jitter vs. the clock | Accepted by design: the window is a whole-run budget (provisional 120 s) against ~10–100 ms jitter — the only regime where server arrival time is a legitimate judge (`rhythm-timing-games.md §4`, model 1). Round-trip time adds to elapsed on every line; elapsed is display-only, and payout is speed-neutral (OD-2), so latency never moves money. |
| Server clock stepping backwards | `t = max(sample, last_server_ms)`; boundary vector required. |
| Reading ahead | The snapshot exposes only the current prompt; the session seed is never sent. The corpus pool is public pinned data — memorizing it is skill, not an exploit. |
| Scripted/pasting client | Cannot be distinguished from a human at line granularity, and v1 does not try. The honest bound: a script reaches exactly the ceiling a careful human reaches (all clean lines), and the faucet caps both identically per attended day. No ranked or leaderboard surface exists in v1; any future ranked or ghost surface must bring its own audit. |
| Stalling Exit with an open session | `end_run` is legal in every non-terminal phase, so the MA-C12 Exit rejection always has a reachable exit. |
| Replay divergence | Resolution replays the full command log with persisted stamps and byte-compares (platform C20); a same-version engine change fails closed. |

### TT6 — Economy hooks (platform-only)

Payout = the certified `typer.clean_lines` through the shipped faucet (daily sends, per-send cap,
ppm conversion with carried remainder, forfeit with `cap.minigame_faucet`) into `company.cash`.
Offline quality charges and decays per C34/C40. No Clout. No run-local currency. No Company
resource enters the engine. The design's click-power multiplier is **not** implemented (DG-2 /
OD-4). Payout is identical for `timed` and `untimed` runs with the same facts (OD-2).

### TT7 — AI fallback and persistence

`fallback: {kind:"solo"}` — the prompt list is the opponent; the game is complete at zero players
online. Ghost races are a successor (DG-4). Sessions are run-bound (the platform pins
`(company_stream_id, run_seq)`); Exit is rejected while one is active (MA-C12). Founder-scope
offline quality and the faucet window persist across Exits (the `design/03` persistence rule).
Nothing else persists in v1.

### TT8 — Accessibility contract

`design/03 §7` is silent on accessibility; `design/11 §7` sets the baseline (keyboard operability,
reduced motion, non-color states, screen-reader labels). The draft Accessibility RFC's A2.2
forbids *requiring* a timed key sequence. This RFC therefore makes the timed run optional and
never the only path:

1. **Untimed path, equal reward.** `untimed` is always offered next to `timed` at `begin`; neither
   is default-selected by the surface; payout is identical (OD-2). The `typer.assisted` fact
   records the choice without stigma, matching Advisor Mode's labeled-not-judged rule. (WCAG 2.2.1
   *Timing Adjustable* has an "essential" exception that a speed test might claim `[M — verify
   before citing]`; this RFC deliberately does not rely on it.)
2. **Any input method.** The server sees only the submitted line, so the surface must not block
   paste, IME composition, dictation, on-screen keyboards, switch access, or word prediction.
   Enter submits; submission is suppressed while `isComposing` is true. The input is a native
   single-line text field with `autocomplete="off"`, `autocapitalize="off"`, `spellcheck="false"`,
   and `autocorrect="off"`; TT4.4's normalization catches the smart punctuation that mobile
   keyboards insert anyway.
3. **Screen reader.** The current prompt is a labeled static text element (monospace `<code>`)
   that can be read character by character; when a new prompt mounts, focus stays in the input and
   one polite announcement names it. A miss is announced as text using `first_mismatch_index`
   ("first difference at character N" — prose pending); correctness is never conveyed by hue
   alone.
4. **Time display.** In timed runs the remaining time is readable text derived from
   `deadline_server_ms − last_server_ms` minus local monotonic time since the response (the Game UI
   RTA pattern). The surface makes no periodic live announcements; cadence is the Accessibility
   RFC's open D-018 question and is not invented here. Reduced motion disables caret/typing
   animation.
5. **Reflow.** The longest permitted prompt (`max_line_bytes`) wraps inside the surface at 320 CSS
   px and at 400% zoom with no horizontal page scroll.
6. **Keyboard.** Every action (mode choice, submit, end run, exit to host) is reachable by
   Tab/Shift-Tab with Enter/Space; no custom shortcut intercepts browser keys; no focus trap.

Before release, the Typer task rows (open → begin in each mode → miss → clear → end run → receipt)
join the Accessibility RFC's release-task register (A5) and produce its evidence.

### TT9 — UI surface contract

Typer mounts as the `(typer, 1.0.0)` child of the MA3 `minigame_session` surface, which owns
transport, chrome, errors, and the closed UI states `loading | active | paused_reconnect |
required_terminal | error` (MA-C9). The child receives presentation props only:

- `snapshot` (the generated `typer.snapshot.v1` arm), `serverTimeSample` (from the response
  envelope), `pending` (a command is in flight), `reducedMotion`.
- callbacks: `dispatch(command: TyperCommand)` and `exitToHost()`.

The child renders the garage-PC terminal in the era token theme (DOM, monospace; no canvas).
Local keystroke echo is ordinary text-field editing; the child computes nothing authoritative. An
optional local diff highlight may accompany TT8.3's text feedback but never replaces it. On
reconnect in a timed run whose deadline has passed, the surface shows the expired state and offers
`end_run` (a late `submit_line` also terminates, unscored). Terminal rendering uses the stored
resolution receipt (MA-C4).

**Copy keys (mechanical; all prose owner-authored and pending, `design/08 §1`):**
`typer.title`, `typer.host`, `typer.mode.timed`, `typer.mode.untimed`, `typer.mode.untimed.note`
(the curtain line stating that payout is identical), `typer.begin`, `typer.submit`,
`typer.end_run`, `typer.feedback.cleared`, `typer.feedback.miss` (slot `{index}`),
`typer.time_remaining` (slot `{seconds}`), `typer.expired`,
`typer.outcome.{completed,ended_early,timed_out}`, `typer.era.garage_2000`,
`typer.prompt.<prompt_id>.scene` (the Oregon-Trail-flavored scene line per prompt — flavor only,
DG-10), and the four `cap.typer_*` reason keys. Every key carries its era/voice tag; the copy
linter gates them.

## DESIGN-GAPs

- **DG-1 — Timing authority.** `design/03 §7` says "skill runs" but not who measures time; the
  platform gives tenants no clock. Resolved *as a proposal* by TT-PA1 (OD-1); options are listed
  there.
- **DG-2 — The click-power multiplier hook has no owner.** The platform payout grammar credits one
  catalog resource; no arm grants a buff or multiplier. Options: (a) v1 pays `company.cash` via the
  faucet and defers the hook [recommended]; (b) an Active-Play amendment adding a
  "minigame-granted click-frenzy window" arm; (c) a production amendment feeding the Typer quality
  grade into the click multiplier slot. Both (b) and (c) are new economy mechanics and need their
  own RFC.
- **DG-3 — The daily streak.** `design/02 §10` supersedes §7's streak with count-up cumulative
  days, owned by a daily/weekly scaffold that does not exist. Deferred to that scaffold; Typer adds
  no streak state.
- **DG-4 — Ghost races.** Racing against "other players' keystroke recordings" needs per-keystroke
  timing (client clock: cosmetic-only under the rhythm research's fine-layer rule), a privacy and
  anonymization contract for other players' recordings, and the unbuilt `async_snapshot` tenant
  mode. A successor RFC owns all three; bot ghosts need a published pace manifest.
- **DG-5 — Accessibility.** The design is silent on motor and screen-reader play for a typing
  game. TT8 proposes the untimed path and input-method neutrality; the equal-payout choice is
  OD-2.
- **DG-6 — "Cooldown-gated."** The clock taxonomy names a cooldown; the platform has none. OD-8
  recommends letting the faucet govern (unlimited play, capped pay).
- **DG-7 — Run-1 collision.** T1 is reached around the scripted first Exit (attended time 900 s
  after the Garage gate). An active Typer session makes Exit reject with `minigame_session_active`,
  so the scripted replacement of the next Company command would reject until the run ends. This
  interaction is unspecified anywhere (OD-3).
- **DG-8 — Achievements.** The Achievements predicate union has no minigame-fact condition. Typer
  achievements need an Achievements-grammar successor; out of scope.
- **DG-9 — Personal best.** "Skill expression" invites a PB display, but no Founder state owns it.
  Deferred; v1 shows only the current run's facts.
- **DG-10 — "Oregon-trail-flavored."** Only the tone is specified. v1 treats it as flavor (one
  scene copy key per prompt); no Oregon-Trail mechanics (random events, resource loss, death
  screens) are invented.

## Deviations from design

1. **Economy hook:** cash through the faucet in place of the `design/03 §7` click-power multiplier
   (DG-2, OD-4).
2. **No streak, no ghost races, no cooldown in v1** (DG-3/4/6) — `§7`'s streak is already
   superseded by `02 §10`.
3. **Untimed accessibility mode** — an addition not in `design/03`; justified by `design/11 §7`,
   the Accessibility RFC's A2.2, and the rhythm research's assist precedent (OD-2).
4. **Unlock requires one prior Exit** (only if OD-3(a)): the game is still Tier 1 and free, but it is
   hidden in run 1 to protect the scripted curriculum. It also removes a conflict with `design/02
   §11` (first minigame at about 2–3 h).
5. **Pacing tension flagged, not resolved:** `design/03` puts Typer at T1 while `design/02 §11`'s
   pacing table says "First minigame ~2–3 hours (Tier 2)". OD-12.

## Acceptance criteria

Each gate names the failing case that proves it can fail.

1. **Tenant row + artifact load (Go/TS).** Both loaders accept the TT1 row and `balance/typer.json`
   from one shared fixture. *Fails:* an unknown key, a prompt with a non-ASCII byte or a double
   space, an eligible pool smaller than `run_length` at some reachable `era_tier`, a Typer row
   without the pinned `typer` artifact, and a `minigame_api` tenant row without the definition each
   hard-fail load in both runtimes.
2. **Deterministic order.** `(seed, content, era_tier)` produces byte-identical prompt orders in
   Go and TS across the corpus seeds. *Fails:* a mutant with an unsorted pool, or one using
   `typer.run.v1` directly without the `typer.prompts.v1` substream, breaks the shared vectors.
3. **TT-PA1 time plumbing.** With an injected DB clock, the value the tenant receives equals the
   persisted `server_ts_ms` for every applied command, terminal included, and verification replay
   reproduces the snapshot/result byte-for-byte from persisted stamps. *Fails:* the current
   sample-at-insert behavior, reinstated as a mutant, diverges under a clock that advances between
   claim and insert; a tenant-side `time.Now()` is caught by the purity grep/test.
4. **Deadline boundary.** A submit stamped exactly at `deadline_server_ms` is scored; one stamped
   `deadline + 1` ends `timed_out` unscored. *Fails:* a `≥` mutant fails the exact-deadline vector.
5. **Clock monotonicity.** A backwards stamp sequence keeps `last_server_ms` monotone and `elapsed`
   non-negative. *Fails:* dropping the `max` makes the vector's elapsed negative, which
   `ValidateResult` rejects.
6. **Comparison and normalization.** Shared vectors: exact clear; case mismatch (`LS -la`) is a
   miss at index 0; prefix/extension misses; each OD-7 mapping clears; an unmapped look-alike
   stays a miss; `line_too_long` and `invalid_text` reject without mutation or revision advance.
   *Fails:* case-folding or NFC mutants fail the case vectors; a truncating mutant fails the
   over-length vector.
7. **Clean-line accounting.** Miss-then-clear increments `lines_cleared` but not `clean_lines`.
   *Fails:* a mutant counting clears as clean.
8. **Composed platform path.** Real Postgres + composed binary through the MA endpoints:
   create → begin → submits → terminal → faucet payout into `company.cash` → receipt; retry
   returns identical bytes; the faucet cap forfeits with its reason; offline quality charges; Exit
   rejects while active and succeeds after `end_run`. *Fails:* a seeded second resolution with the
   same session ID credits nothing new.
9. **Unlock (TT-PA2).** Start at Tier 0 rejects `not_eligible/tier_required`; at Tier 1 (with the
   OD-3 exit-history condition, if ruled) it succeeds; the predicate reads only pinned server
   state. *Fails:* removing the resolver check lets the Tier-0 start pass, and the test goes red.
10. **Authority.** A command containing `score`, `server_ms`, `elapsed_ms`, or `prompt_id` is
    rejected by schema. *Fails:* a fixture carrying `server_ms` must be rejected; accepting it
    fails the test.
11. **Payout neutrality (OD-2 as ruled).** An identical command sequence in `timed` and `untimed`
    yields an identical credited amount (or, under OD-2(b), exactly the declared reduction).
    *Fails:* a mutant keying payout on `typer.assisted`.
12. **Content gate.** A versioned corpus names at least one `(seed, commands, stamps)` scenario per
    prompt showing it reachable, and covers all three outcomes, both modes, every rejection code,
    and the boundary vectors of AC4–AC6, compared byte-for-byte in Go and TS. The fixed transition
    budget equals the sum of corpus command counts. *Fails:* adding a prompt no corpus seed deals,
    or an era no reachable `era_tier` admits.
13. **Accessibility.** Through the UI Foundation browser gate (Chromium, Firefox, WebKit): a
    keyboard-only untimed run completes; axe shows zero serious/critical issues on every Typer
    state; the longest prompt reflows at 320 px; paste and composition events submit correctly.
    *Fails:* a seeded `preventDefault` on paste, a hue-only miss indicator, or a submit during
    `isComposing`. Manual screen-reader records per Accessibility RFC A5.
14. **Copy.** Every TT9 key resolves; the copy linter passes. *Fails:* one removed key fails
    resolution.
15. **Isolation.** No Company resource enters the engine, and the engine emits no economy delta
    (grep + test, Pitch AC5).

## Owner decisions (recommended defaults marked)

1. **OD-1 — Timing model.** (a) **[recommended]** TT-PA1: a server-sampled per-command clock
   handed to the pure tenant, with a coarse whole-run budget; (b) untimed-only v1 with no platform
   amendment (no speed skill at all); (c) client-reported keystroke timeline — rejected for any
   payout (forgeable), acceptable later as cosmetic only.
2. **OD-2 — Payout fact and neutrality.** (a) **[recommended]** payout/quality on
   `typer.clean_lines`, identical for timed and untimed runs (accuracy is the rewarded skill; speed
   is expression; accessible by construction); (b) untimed pays a declared, curtain-labeled reduced
   rate (the rhythm-research "visibly lower cap" precedent); (c) a speed-weighted score (excludes
   motor-impaired players from the top of the reward).
3. **OD-3 — Unlock and the run-1 collision.** (a) **[recommended]**
   `tier_at_least {tier:1, exit_history_at_least:1}` — free, Tier 1, from run 2 onward; (b) tier
   only, accepting that a Company command during an active run-1 session gets a rejected scripted
   Exit; (c) tier only, plus a Prestige amendment deferring the scripted Exit while a minigame is
   active.
4. **OD-4 — Click-power hook.** (a) **[recommended]** defer, cash-only v1; (b) Active-Play buff arm
   successor; (c) production multiplier from the quality grade successor.
5. **OD-5 — Soul gate.** (a) **[recommended]** `unrelated` — typing commands is the founder's work,
   not a hobby, and work that survives the loss of joy is on-thesis; (b) `human_hobby` (locks at
   near-zero Soul like board games).
6. **OD-6 — Pool semantics across eras.** (a) **[recommended]** cumulative (every era with
   `min_tier ≤ era_tier`; breadth); (b) latest era only (a `challenge`-style difficulty jump).
7. **OD-7 — Input normalization.** (a) **[recommended]** the closed four-entry smart-punctuation
   map in TT4.4 `[M — verify iOS/Android smart-punctuation behavior before shipping]`; (b) strict
   bytes, with no map.
8. **OD-8 — Cooldown.** (a) **[recommended]** none; the faucet's daily sends are the payout clock
   and play is unlimited; (b) a platform-level session cooldown in a successor RFC.
9. **OD-9 — Corpus authorship.** (a) **[recommended]** prompt `text` rows are owner-ratified
   content, not localized, with placeholders allowed only in fixtures; (b) Claude drafts the
   corpus for editorial cut (the `design/11 §6` batch workflow).
10. **OD-10 — Provisional literals.** Ratify as placeholders: `run_length 8`,
    `timed_budget_ms 120000`, `max_line_bytes 256`, and the Pitch-mirrored faucet, quality, and
    rating rows. T1 cash magnitude must be harness-set before minting.
11. **OD-11 — Mint.** (a) **[recommended]** fixture-first; the production mint (the `typer`
    artifact plus the `minigames`/`minigame_api` edits) goes in a later content epoch after the
    content gate passes (the TP-C18 precedent).
12. **OD-12 — Pacing.** Confirm that `design/03`'s Tier-1 placement stands against `design/02
    §11`'s "first minigame ~2–3 h (Tier 2)"; OD-3(a) largely reconciles the two.

## Open questions

- Does the shipped faucet consume a daily send on a zero-score resolution? If so, an immediate
  `ended_early` spends a send and pays nothing. Verify at acceptance review; if it does, choose
  between accepting it (the cap is published) and a platform amendment that skips zero-score sends.
- Successors named here: Typer Ghost Races (DG-4), the Tier-3 corpus batch, Typer achievements
  (DG-8), the personal-best owner (DG-9), and whichever click-power arm OD-4 selects.

## Changelog

- 2026-09-25: created (draft) — Tier-1 tenant on the Minigame Platform, mirroring The Pitch
  template; four named platform/composition amendments; twelve owner decisions.

## Owner acceptance (2026-09-25)

Marco accepted this RFC in the 2026-09-25 batch (the answer to a direct batch question in-session,
recorded by Claude). Every owner decision this RFC lists is **ruled at its stated recommended
default**. Where the text offers alternatives, the recommended option is the normative one and the
alternatives are rejected. Provisional numbers stay provisional and are ratified by harness
measurement and SHA, as the RFC already requires. Owner-authored copy stays owner-authored:
implementation ships copy keys, with any placeholder or candidate text clearly marked.
