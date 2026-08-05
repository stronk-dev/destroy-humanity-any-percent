# RFC: Minigame Platform Foundation

- **Status:** accepted — C1–C18 ruled; normative body reconciled; implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/03` (the minigame suite, clocks, economy hooks, AI-fallback law), `design/03 §12b` (the tier→minigame scaling seam + Fairness Law), `design/05 §4` (PvP table, punch-down multiplier)
- **Research:** `neopets-systems.md §4` (the portal faucet model: N-sends × per-game cap × conversion ratio, monthly re-tune), `kol-puzzle-pirates.md §B` (minigame-as-labor, offline-output quality grades), `creature-battler.md` / `lane-pusher-design.md` (the combat instances)
- **Depends on:** Production + Run Genesis + Transport + Gameserver Composition (implemented); Combat Shared Kernel (implemented — the duel engine is tenant #1). **NOT Clout** (removed per C1).
- **Owner ruling honored:** breadth-first — the PLATFORM every minigame is content on: session lifecycle, the scaling seam as code, the faucet governor, reward hooks, the registry. No individual minigame.
- **Planning:** `planning/minigame-platform-foundation/` (once implementing)

## Summary

Every future minigame — board games, the two combat engines, the Market, the arcade toys, pets'
battles — needs the same platform: a session/match lifecycle, the §12b scaling-seam contract as
executable code, the Neopets faucet governor (so a minigame can never become an economy exploit),
reward payout into the real economy, and a registry. Combat's C5 already prototyped the seam
("the seam is these two integers"); this generalizes it to a platform contract.

## Specification

### MP1 — Session/match lifecycle (closed state machine)

A minigame session is a **Postgres row** (C13's `minigame_sessions` — the verification-queue table
discipline: claim-token + RowsAffected, server-advanced only) — `{session_id, minigame_id,
founder, company_stream_id, run_seq, scaling_inputs (frozen), seed, mode: solo|async_snapshot,
state, result|null}`. **`live_pvp` is a reserved enum arm, NOT shipped (C8)** — Phase-A modes are
`solo` and `async_snapshot` only, the two combat proves replayable. Lifecycle: create (server
resolves scaling inputs, freezes them, seeds from the save stream — the combat C5 pattern) → play
(the tenant engine advances the session via server-validated play commands) → **resolve
(server-authoritative, replay-verified like combat — solo/snapshot both)** → payout (MP4, a
server-authored transition). No minigame reads live company state mid-session; the seam is always
the frozen integer inputs (replay-safe by the RA rule, exactly as combat proved).

### MP2 — The scaling-seam contract (§12b as code)

`scaling_inputs`: closed named integer functions over (company, founder) state, per the §12b
axes (`reward`, `resource_pool`, `challenge`, `breadth`, `era`, `offline_quality`), declared per
minigame in its registry entry, computed server-side at session create, frozen into the snapshot.
**The Fairness Law is loader-enforced:** a scaling input feeding a RANKED PvP power stat is a load
error; tier scaling reaches ranked play only as `breadth` (options inside a fixed ceiling — the
2026-07-28 hardcap decision generalized from pets to all ranked minigames). `offline_quality`
(the Puzzle Pirates grade) is the active/idle bridge: recent session performance charges the
quality tier of the associated automated output, decaying over ~31 days, never zeroed.

### MP3 — The faucet governor (the Neopets shape, mandatory)

Every minigame declares `payout: {sends_per_day, per_send_cap, conversion_ppm}` — N score-sends
per day, each capped, converted at a published ratio. This is the STRUCTURAL guarantee a minigame
cannot become an infinite faucet (the Cloud-Clicker economy laws' answer to "minigame prints
currency"): the daily cap is the governor, published per epoch, re-tunable on the epoch cadence
(the Neopets monthly-25th precedent as our BALANCE-CHANGE cadence). Rewards above the cap are
forfeit (accrual-only, reason-keyed). Bot/AI-fallback matches pay at a declared reduced,
non-ranked rate (anti-farm — the researched universal).

### MP4 — Reward payout & the AI-fallback law

Resolve → payout (rulings C1/C6/C17). **NO Clout** (Feed/Social owns Clout's mint). **The payout
is NOT a client intent** — a minigame result is a SERVER-CERTIFIED outcome: the server resolves
the session (replay-verified), then commits, in ONE Postgres transaction, the session row
(status=resolved, claim-token-guarded) AND a **server-authored Company-stream transition** —
`resolve_minigame_session` (server-only, never client-submitted) — that credits a tier-scaled
economy resource (governed by MP3) plus typed rating/season facts (real Elo, no rubber-band).
Evented and replay-logged like every transition; the certified result is the recorded fact
(replay never re-runs the minigame at verification). The client submits PLAY commands to the
session and NEVER a score. **Every minigame declares its AI fallback** (closed union, C18:
`solo|bot|npc_partner`) — a fallback-less minigame is a load error; the whole game is playable at
zero other players online.

### MP5 — The registry

`balance/minigames/*.json`: `{minigame_id, engine_ref, clock: turn|realtime|wallclock|async,
scaling_inputs[], payout, fallback, unlock_condition_ref}` (era_skins REMOVED per C9 — one token
theme per era, never per-feature). `unlock_condition_ref` uses the closed shell-fact grammar
(UI C5). The combat DUEL engine registers as tenant #1 (their existing catalogs become platform-conformant); the board
games, Market, and arcade toys are later content on this registry. `unlock_condition_ref` gates
by catalog fact (staggered across the tier arc — the CC hour-5-40-sag fix); the platform never
invents unlock logic.

## Owner rulings on C1–C11 (2026-08-04)

- **C1 — accepted, and this one is on me:** MP4 reintroduced the exact Clout-as-payout faucet I
  ruled OUT of Achievements an hour earlier. **Clout removed from payout and dependencies.**
  Phase-A payout arms are only owners that exist at implementation: an economy-resource credit
  (catalog resource ID) and typed rating/season facts (non-resource state). Feed/Social may add a
  social-activity Clout arm by successor RFC through the closed reward union; the generic platform
  never mints a free-form resource/effect.
- **C2 — accepted, the session is DB-authoritative:** sessions are Postgres rows
  (`minigame_sessions`), server-advanced only; concurrency and double-pay prevented by the SAME
  claim-token + RowsAffected discipline the verification queue uses; resolve/payout is one
  transaction keyed by session id (idempotent). No in-memory actor is authoritative for
  async/snapshot modes.
- **C3 — accepted:** `engine_ref` binds to a closed tenant interface — a typed
  `(command, snapshot, result)` descriptor + engine-version identity + deterministic error
  taxonomy; the combat duel adapter and a fixture tenant both implement it and are conformance-
  tested against the same boundary; an engine reading ambient state or returning an unvalidated
  payout is a load/contract error.
- **C4 — accepted:** scaling axes become typed rows (source field, integer range, composition
  order, ranked-power binding) so the Fairness Law is loader-PROVABLE, not asserted.
- **C5 — accepted:** faucet uses **attended-time day boundaries on the absolute grid** (not
  wall-clock — the provision-grid precedent kills the timezone/deploy ambiguity), integer-ppm
  conversion with declared rounding + overflow saturation, cap consumption persisted per session
  (a resolve retry can't double-spend — same idempotency), bot reduction a catalog literal.
- **C6 — accepted, the sharpest correction:** payout is NOT a client intent. A minigame result is
  a SERVER-CERTIFIED outcome: the server resolves the session (replay-verified for solo/snapshot,
  like combat), then the payout enters the economy through a **server-authored** transition inside
  the resolve transaction — evented and replay-logged, but never a client-submitted
  `claim_payout`. The client submits PLAY commands to the session; it never submits a score.
- **C7 — accepted:** fallback is a closed union with exact fields (solo: no peer; bot: catalog
  bot identity + reduced non-ranked rate literal + replay via the deterministic bot policy;
  npc_partner: the virtual-guild pattern). Bot results replay like any tenant result.
- **C8 — accepted, SCOPE RULING: `live_pvp` is DEFERRED out of this foundation.** Phase-A ships
  `solo` and `async_snapshot` ONLY (the modes combat already proves replayable). The `live_pvp`
  mode boundary is DECLARED (the enum arm reserved) but unimplemented; the PvP-service RFC owns
  disconnect/reconnect/authoritative-clock/finalization. AC1 drops to the two shipped modes. This
  keeps the platform honest — it ships what's replay-provable now.
- **C9 — accepted:** `unlock_condition_ref` reuses the closed shell-fact grammar (UI C5); **`era_skins`
  REMOVED** (UI Foundation: one token theme per era, never per-feature skins — my `era_skins` was
  a UI-law violation). AC6 corrected: the combat duel engine registers as tenant #1 via adapter;
  the lane engine follows when implemented (not claimed ready).
- **C10 — accepted, and the labor hazard ruled:** `offline_quality` is stored session state (a
  ppm grade + last-session timestamp), a deterministic decay transition on the attended grid; **it
  decays toward a NEUTRAL floor, never zero** (the ruling that keeps an optional minigame from
  becoming mandatory labor — a lapsed player's automated output degrades to baseline, never below).
- **C11 — accepted:** the registry + tenant schemas + fallback/faucet policy join epoch identity
  (a minigames artifact, mint like the others); session genesis + results join the verifier where
  they affect verification; the DB migration + cross-runtime corpus are named in the closure batch.

## Owner rulings on C12–C18 (2026-08-04)

- **C12 — accepted, body reconciled:** the normative text above is corrected to match the C1–C11
  rulings (live_pvp arm reserved-not-shipped, era_skins removed, duel engine only as tenant #1).
  My rulings appended without fixing the body — corrected now.
- **C13 — session persistence: reuse the verification-queue table discipline exactly.**
  `minigame_sessions(session_id PK, minigame_id, founder_id, company_stream_id, run_seq,
  scaling_inputs jsonb, seed, mode, status, claim_token, state jsonb, result jsonb|null,
  created_at, resolved_at)`; server-advanced only; concurrency + double-resolve prevented by
  claim-token + RowsAffected (the queue's proven pattern); lock order founder-then-session;
  expiry/retention as the queue's. No new invention — the pattern is shipped.
- **C14 — tenant descriptor:** a closed `(command_schema, snapshot_schema, result_schema,
  engine_version, error_taxonomy)` per engine_ref; the result union is `{outcome, score_facts,
  rating_delta|null}` (NO free-form payout — C1); the combat duel adapter and the fixture tenant
  both implement it, conformance-tested against one boundary. Exact byte envelopes are the tenant's
  (combat already has them); the PLATFORM contract is the descriptor shape.
- **C15 — scaling loader-provable:** each scaling row maps its destination to exactly one class
  `power|breadth|presentation`; **the Fairness Law is the loader rule `ranked ∧ power → reject`**;
  bounds/formulas per row are BALANCE DATA (harness-tuned, formula-artifact-exported) — the
  SCHEMA is ruled here, the numbers are not invented.
- **C16 — faucet clock (amended by C22):** attended-grid origin = the persistent Founder attended
  cursor (the quota spans runs and never resets on Exit), floor-division ppm conversion, **saturation
  distinguished from the configured cap** (the cap forfeits with a reason key; numeric overflow
  saturates — two different events); bot-reduction literal + reset order are catalog data.
- **C17 — server-authored payout transaction:** the resolve transaction commits, in ONE Postgres
  tx: the session row (status=resolved, claim-token-guarded) AND a server-authored Company-stream
  transition (the payout — a credit + rating facts, evented + replay-logged), same multi-write
  discipline as Exit; the internal command is `resolve_minigame_session` (server-only, never a
  client intent — C6); lock order company-then-session. Replay reads the certified result from the
  log, never re-runs the minigame at verification (the result IS the recorded fact).
- **C18 — fallback/offline_quality:** fallback union exact-keys `{kind: solo|bot|npc_partner,
  bot_ref?, rate_reduction_ppm?}`; `offline_quality` = `{grade_ppm, last_session_at}` on session
  state, decaying on the attended grid toward a **neutral floor (catalog literal, never zero)**
  via a fixed-grid formula (the provision-grid partition-invariance applies); the automation
  destination is the associated tenant's offline output. Grade SOURCE (score→grade curve),
  decay rate, and floor value are BALANCE DATA — the state/transition shape is ruled, the numbers
  deferred.

The through-line of C13–C18: **structure ruled precisely by reusing shipped patterns (queue table,
provision grid, Exit multi-write); balance NUMBERS deferred to harness-tuned catalog data** — the
platform is executable without inventing a single balance value.

## Owner rulings on C19-C23 (2026-08-05)

Same discipline as C12-C18: structure ruled by reusing shipped patterns, balance numbers deferred.

- **C19 - accepted:** the scaling-source union is closed with exact wire rows -
  `literal|tier|purchased_generator_count|founder_carry_counter|attended_quality_grade`, each a
  typed row with source field + integer bounds + the power|breadth|presentation destination class;
  the Fairness Law loader rule is `ranked AND power -> reject`. Per-row formulas/bounds are BALANCE
  DATA; the union and destination grammar are ruled here.
- **C20 - accepted, the replay requirement forces append-only command rows:** `minigame_sessions`
  (mutable head, C13) is joined by `minigame_session_commands` (append-only, `(session_id, seq)` PK,
  canonical command bytes + server-stamp) - the run_log pattern applied to sessions. Resolve
  replays the command log through the tenant engine and byte-compares snapshot/result, exactly as
  the run verifier replays the run log. Solo/async both.
- **C21 - accepted:** `resolve_minigame_session` canonical payload = `{session_id, result}`;
  replay inputs = the frozen scaling_inputs + seed + command log (all already persisted, nothing
  new in the Company replay_inputs union - the payout transition reads the CERTIFIED result);
  receipt = the credit + rating facts; event kinds `minigame_resolved.v1` (+ the credit is an
  ordinary resource event); rating/season facts land on Founder-scoped rating state via the
  Founder boundary (Pet Care C1's `ApplyFounderLogged` - ratings are Founder-persistent, not run
  state); the production method is server-only, never a client intent.
- **C22 - accepted:** the faucet quota counter is SESSION-scoped persisted on the founder's
  daily-window row keyed by the attended-grid day. **Origin is the Founder attended cursor**
  (Pet Care C4), since sends-per-day spans runs; quota does NOT reset on
  Exit, it resets on the attended-grid day boundary); the persisted cursor is the founder daily
  window. This corrects C16's run_started_at origin to the Founder attended cursor for the
  cross-run faucet.
- **C23 - accepted:** fallback rows exact - `{kind, bot_ref?(policy_id+version), npc_profile?,
  rate_reduction_ppm?}`; bot identity is the combat-bots contract's manifest (policy_id+version,
  deterministic replay); the score-to-grade curve + decay rate + neutral floor for offline_quality
  are BALANCE DATA; the state shape (grade_ppm + last_session_at, fixed-grid decay to neutral
  floor) is ruled.

## Implementation blockers C24-C27 (Codex, 2026-08-05)

C20's immutable session history is implemented and independently approved. The remaining code
reaches four contracts that the C19-C23 summary still names without an executable wire or atomic
owner boundary.

### C24 — C19 does not yet enumerate the transform grammar

The ruling names five source kinds and “integer bounds” but supplies no arm keys, operation order,
rounding, clamp, duplicate rule, or closed Founder counter paths. A loader still cannot distinguish
a valid scaling program from an invented one.

**Proposed contract:** every destination row is exact-key `{id,class,source}` with unique `id` and
`class=power|breadth|presentation`. `literal` source is `{kind,value}`. The other four arms add
their single closed field (`tier`; `generator_id`; `founder_counter`; or `minigame_id`) plus
`{offset,multiplier,denominator,min,max}`. Evaluation is checked exact-integer
`floor((source+offset)*multiplier/denominator)` followed by clamp; denominator is positive and
bounds/order are loader-validated. `founder_counter` is a closed catalog enum, not a JSON path.
Ranked destinations reject `class=power` when any source is `tier`; duplicate destination IDs
reject. The formula generator renders this exact program. Numeric operands and the allowed
Founder-counter rows remain owner/harness catalog data.

### C25 — Company payout plus Founder rating has no one-transaction boundary

C17 rules Company→session locking, C13 says Founder→session, the project-wide order is
Founder→Company, and C21 sends rating through `ApplyFounderLogged`. Calling that exported method
inside the resolve transaction would nest transactions; calling it afterward permits paid-without-
rating or rating-without-paid crashes.

**Proposed contract:** one repository operation `ApplyMinigameResolutionTransaction` locks
Founder, Company, then session. It runs two pure owned transitions under that transaction: the
server-authored Company credit/result transition and the Founder rating transition. It appends the
Company run log, Founder log, events, both revisions, faucet row, receipt, and token-owned session
terminal mark atomically, keyed idempotently by session ID. Neither exported Store method is
nested. Replay inputs on both sides reference the same immutable certified-result bytes/hash.
Fault injection follows every write, and retry returns the one committed receipt.

### C26 — The faucet clock depends on the unresolved Founder attendance authority

C22 correctly moves quota across runs, but `founder_attended_ms` is not implemented and a server
timestamp cannot distinguish attended from offline time. The payout counter cannot be replay-safe
until Pet C10 (or a successor shared attendance RFC) owns that clock.

**Proposed contract:** reuse the single persisted Founder attendance accumulator ruled at Pet C10.
The resolution transaction freezes its sampled total, derives `window_index=floor(total/day_ms)`,
and updates one `(founder_id,minigame_id,window_index)` row idempotently by session ID. Operation
order is bot reduction → ppm conversion with persisted remainder → per-session cap → remaining
window cap; configured-cap forfeiture and numeric saturation remain distinct typed outcomes.

### C27 — Fallback/offline-quality “exact rows” still contain unnamed objects

`npc_profile?`, the score-grade curve, remainder, and automation destination have no exact keys or
closed identity/version rule. A strict loader still has no valid production row to accept.

**Proposed contract:** close the arms as `{kind:"solo"}`;
`{kind:"bot",bot_ref:{policy_id,version},rate_reduction_ppm}`; and
`{kind:"npc_partner",npc_profile:{profile_id,version},rate_reduction_ppm}`. The offline policy is
exact-key `{score_fact,grade_curve,decay_grid_ms,decay_ppm_per_grid,neutral_floor_ppm,
automation_destination}`; curve thresholds are strictly increasing and grades nondecreasing,
decay carries an integer remainder, and the destination must be registered by the tenant. All
literals are balance data. Until those rows exist in a minted artifact, the loader enables no
production fallback/offline automation.

## Owner rulings on C24-C27 (2026-08-05)

- **C24 - accepted:** the scaling transform grammar is closed - per source-kind arm keys, a fixed
  operation order (source -> integer op -> clamp to declared bounds), floor rounding, duplicate-
  source rejection, and the closed Founder-counter paths (the founder_carry_counter arm names
  exactly the frozen career fields). Loader distinguishes valid from invented by the closed grammar;
  per-row numbers are balance data.
- **C25 - the transaction ruling: minigame resolve is a MULTI-STREAM transaction like Exit.** It
  does NOT call exported ApplyFounderLogged (which owns its own tx - nesting). Instead the internal
  Company-payout and Founder-rating transitions COMPOSE under ONE Postgres transaction, Founder-
  then-Company lock order (the project-wide order), exactly as the Exit transaction composes
  Company+Founder writes today. One tx: session row resolved + Company payout revision + Founder
  rating revision, or none. No paid-without-rated race.
- **C26 - BLOCKED ON the Founder Attendance Foundation** (same primitive as Pet C10). The cross-run
  faucet quota reads founder_attended_ms; it is not replay-safe until that clock lands. Honestly
  gated, not improvised.
- **C27 - accepted:** fallback rows fully closed - `npc_profile` is a catalog identity+version, the
  bot_ref is the combat-bots manifest (policy_id+version, deterministic), offline_quality state is
  `{grade_ppm, last_session_at}` with a fixed-grid decay-to-neutral-floor transition and the
  automation destination named (the tenant's offline output); the score->grade curve, remainder,
  and floor value are balance data. Every object has a closed identity/version rule.

## Owner rulings on C28-C30 (2026-08-05)

- **C28 - accepted: the transform grammar is one closed ROW (not a graph).** Each scaling row is
  exact-key `{destination, destination_class, source_kind, source_ref, op, operand, clamp_min,
  clamp_max}`; `op` is a closed union (identity|add|mul|floordiv) applied in the fixed order
  source->op(operand)->clamp; floor rounding, negative-safe; composition is ONE row per
  destination (no graph - a destination fed twice is a load error). The Fairness Law stays
  `ranked AND destination_class=power -> reject`. Loader distinguishes valid from invented by the
  closed grammar; operands/bounds are balance data.
- **C29 - accepted: Founder rating is a Founder-scope mechanic written through ApplyFounderLogged**
  (the boundary exists exactly for this). Persisted: a Founder-scope `minigame_ratings` map
  keyed by minigame_id (int Elo, initial 1000, floor 400 - the combat-bots contract's numbers as
  balance data), plus a season-fact union. The resolve multi-stream transaction (C25) writes the
  Company payout revision AND the Founder rating revision in one tx, Founder-then-Company; the
  rating write is a new closed ApplyFounderLogged resolved-input arm `minigame_resolved` with its
  own receipt + `minigame_rating_changed.v1` event. No invented Founder state - it's the Founder
  boundary doing what it's for.
- **C30 - accepted: the faucet payout row is exact-key** `{credited_resource_id, sends_per_day,
  per_send_cap, conversion_ppm}`; the carried conversion remainder is a persisted session-state
  field (`conversion_remainder_ppm`, floor-div with carry - the provision-grid pattern); session
  idempotency is the claim-token + `(session_id)` PK the verification queue uses. The credited
  resource is a declared catalog resource ID (never free-form). The locked transaction is now
  writable/replayable from named columns.

## Owner rulings on C31-C33 (2026-08-05)

- **C31 - accepted, wall timestamp corrected to attended watermark:** C18's `last_session_at`
  (wall time) contradicts the accepted attended-grid decay clock. Ruling: offline_quality state is
  `{grade_ppm, last_founder_attended_ms}` - the decay watermark is the FOUNDER ATTENDED cursor
  (Founder Attendance A1), never a wall timestamp, so replay derives decay deterministically. The
  grade curve is a closed row `{threshold_grade_ppm, ...}` with a fixed SHAPE (the enum/order is
  wire grammar); the threshold VALUES are balance data. Decay to the neutral floor, fixed-grid.
- **C32 - accepted, payout names its inputs:** a tenant result is a sorted typed score-fact array,
  so the payout row declares WHICH fact feeds it - add `payout_score_fact_id` (a declared fact ID
  from the tenant descriptor's result schema, C14) and `cap_reason_key` (a copy key for the
  forfeit path, distinct from numeric saturation). Neither is chosen in code - both are declared
  protocol/content the loader validates against the tenant's result union.
- **C33 - the carry-owner contradiction between my C22 and C30, resolved toward C22's cross-run
  row:** a session-local remainder cannot carry across sessions. Ruling: BOTH the conversion
  remainder AND the daily quota counter live on ONE cross-run row
  `minigame_faucet_window(founder_id, minigame_id, attended_day, quota_used, conversion_remainder_ppm)`
  keyed on the Founder-attended day; the resolve transaction updates that window row atomically
  (claim-token-guarded for idempotency) - carry persists cross-session because it's on the window,
  and retry is exactly-once because the session's contribution to the window is idempotent. C30's
  'session-state remainder' is withdrawn; the window row owns both.

## Implementation blockers C34-C35 (Codex, 2026-08-05)

The declared score/copy ownership and cross-run faucet window are implemented. The two remaining
composition gaps are exact wire/activation issues, not balance literals.

### C34 — C31 still does not enumerate its claimed fixed grade-curve row

C31 replaces wall time with the correct Founder-attended watermark, but its literal
`{threshold_grade_ppm, ...}` contains an ellipsis and does not say whether the threshold is a
score or a grade. That contradicts the "fixed shape" claim and cannot drive an exact loader.

**Proposed correction:** retain C31's outer state and attended clock, and adopt the earlier exact
curve row `{score_at_least,grade_ppm}`. Rows are nonempty, byte-order-stable in ascending
`score_at_least`, begin at zero, use strictly increasing score thresholds and nondecreasing grades,
and every grade is within `[neutral_floor_ppm,1_000_000]`. The persisted state is exact
`{grade_ppm,last_founder_attended_ms,decay_remainder_ppm}`. Numeric literals remain balance data.

### C35 — Founder rating shares Pet C16's unresolved version/artifact seam

C29 places `minigame_ratings` in the Founder save and requires the multi-stream resolution to write
it through the Founder replay boundary. The existing Founder wire has no ratings field, and Exit
still requires Founder/Company version equality. Adding the map ad hoc would either make replay
ignore it or activate it under a deploy-current schema.

**Proposed contract:** the first Founder version that carries Minigame state adds exact
`minigame_ratings` and `minigame_offline_quality` maps together and is activated only when the
pinned artifact bundle contains `minigames`. If Pet C16 owns Founder v17 first, this is its next
Founder-only version; otherwise it may share v17 only if one epoch atomically carries both exact
artifacts and both maps. The version transition, replay artifact biconditional, and mixed-version
Exit rules use the same reusable mechanism as Pet C16. Production resolution remains disabled
until one owner ruling fixes that ordering; no code assigns competing mechanics the same version.

## Owner rulings on C34-C35 (2026-08-05)

- **C34 - accepted:** the grade-curve row is exact-key `{score_threshold, grade_ppm}` (the
  threshold is a SCORE, the output is a grade - C31's ellipsis resolved), ascending by
  score_threshold, closed set; scores at/above a row's threshold take its grade, floor below;
  the neutral floor is the lowest grade. Threshold/grade NUMBERS are balance data; the row shape
  is wire grammar. Decay stays on the Founder-attended watermark (C31).
- **C35 - the SAME seam as Pet C16, ruled there: the Founder save version is an independent axis.**
  `minigame_ratings` activates on the FOUNDER version axis under the minigame artifact (pinned),
  written through ApplyFounderLogged in the resolve multi-stream transaction (C25/C29); Exit
  validates the Founder version against its pinned-epoch floor independently of the Company
  version - no ad-hoc field, no deploy-current schema. See Pet C16 for the full ruling.

## Implementation blocker C36 (Codex, 2026-08-05)

C35 now has its independent-axis Exit validator, but no exact Founder save object exists for the
two minigame maps. C29 says an integer Elo map plus a season-fact union without enumerating the
rating row, season identity, fact members, or fact object keys. C34 closes the offline-quality
state itself, not the enclosing map or its biconditional with the `minigames` artifact. Assigning
v17/v18 also collides with Pet C17's ordering question on the same scalar Founder axis.

**Proposed contract:** if the owner adopts the queue order, Founder v17 adds exact
`minigame_ratings` and `minigame_offline_quality` maps keyed by declared minigame ID, and the
`minigames` artifact is biconditional with v17+. Enumerate the exact rating row and closed season-
fact wire, then make replaycatalog Go/TypeScript accept base-nine plus this one named artifact and
derive Founder floor 17 from its presence. Pet state follows at v18. No production catalog row or
balance literal need ship with the schema implementation; activation remains New-Founder-forward
under a later protocol-compliant mint.

## Owner ruling on C36 (2026-08-05)

**The scalar-vs-feature-vector fork is decided: the Founder version stays a SCALAR monotonic
chain, with a fixed total activation order. Rejecting the feature-vector envelope for now.** The
reasoning is the same discipline the Company axis already runs on: a save at version N carries the
UNION of every field introduced at versions ≤ N, and a mechanic whose artifact is not pinned in the
epoch simply holds empty/default state. Activation is *always* artifact-gated (the activation-
boundary law), independent of the version number — so the version's only job is to describe the
save-schema shape. A scalar keeps replay/verification LINEAR (handle version N by knowing 1..N); a
feature vector makes it combinatorial (2^K feature subsets to reason about at Exit). The only cost
of the scalar is that a higher-versioned Founder mechanic drags in the empty maps of the lower ones
— which is exactly the near-free cost the Company save already pays for its own inactive mechanics,
not a new hazard.

- **Order: `minigames` = Founder v17, `pets` = Founder v18** (Codex's queue order — no reason to
  invert; minigames is the platform pet battles later consume). Company stays on its own axis
  (v14/v16) and rejects v17/v18; the two axes never compare for equality (C35/C16).
- **v17 Founder-save additions (exact, in the Founder save jsonb — NOT a side table; ratings live
  in the replay-owned save exactly as pet state does, C14 — a second mutable ratings authority
  beside the Founder save is forbidden, which resolves C29's open table-vs-save fork toward the
  save):**
  - `minigame_ratings`: map keyed by declared `minigame_id` → exact-key row
    `{elo, season_member, games_counted}`. `elo` is a safe integer; `season_member` is a member of
    the closed `rating_season` enum (the "season-fact union" — enum MEMBERS are wire grammar, the
    member catalog is deferred); `games_counted` is a safe integer for provisional logic. Starting
    elo, K-factor, and the provisional threshold are BALANCE data, not ruled here.
  - `minigame_offline_quality`: map keyed by declared `minigame_id` → the C34 state row
    `{grade_ppm, last_founder_attended_ms, decay_remainder_ppm}` (already ruled; watermark clock,
    never wall time).
  - The append-only resolution fact stays in the Founder log arm `minigame_resolution.v1` (C29),
    NOT the save: exact keys `{session_id, certified_result_hash, old_elo, new_elo, season_member}`,
    season facts sorted; the receipt/event are exact projections of that arm. The save holds current
    rating; the log holds the fact history — no duplication.
  - **Biconditional:** the `minigames` artifact is pinned in the epoch ⟺ Founder floor ≥ 17.
    replaycatalog (Go + TS) accepts base-nine plus this one named artifact and derives floor 17 from
    its presence.
- **v18 Founder-save additions:** the pet state map (C14, already save-owned). v18 REQUIRES the
  v17 `minigames` artifact still pinned (the linear chain) AND adds the complete `pets` artifact
  (enumerated in Pet C17's ruling). `pets` artifact pinned ⟺ Founder floor ≥ 18.
- **Escape hatch, not foreclosed:** if a future epoch ever needs a higher Founder mechanic WITHOUT
  a lower one, or needs to sunset a Founder mechanic, THAT is when the feature-vector envelope earns
  its complexity — in a named successor RFC. Until such a need is concrete, the scalar chain is the
  ruling and the feature vector is YAGNI.
- Structure ruled (keys, enum members, biconditional, chain order); numbers deferred (elo/K/
  provisional/season catalog). Activation remains New-Founder-forward under a protocol-compliant
  mint; no production row ships with the schema.

## Acceptance criteria

1. Session lifecycle: create→play→resolve→payout for a fixture tenant against the composed
   gameserver, in `solo` and `async_snapshot` modes ONLY (live_pvp deferred, C8); scaling inputs
   frozen at create, unread thereafter (replay-safe proof).
2. Fairness Law: loader rejects a scaling input wired to a ranked power stat; `breadth` scaling
   into ranked play passes; `offline_quality` charges/decays per the grade ladder.
3. Faucet governor: the daily cap forfeits excess (reason-keyed); a bot match pays the reduced
   rate; per-epoch re-tune follows BALANCE-CHANGE.
4. Server-authored payout: a resolved result enters the economy through the internal
   `resolve_minigame_session` transition in the claim-token-owned resolve transaction, evented and
   replay-logged; the public intent decoder cannot construct it and no side-channel write exists
   (grep-proven).
5. Fallback law: loader rejects a fallback-less minigame; the fixture tenant's solo path completes
   at zero connected peers.
6. Registry: the combat DUEL engine registers as a conformant tenant via adapter (the lane engine follows when implemented — not claimed ready, C9).

## Implementation blockers after Founder Attendance landed (Codex, 2026-08-05)

Founder Attendance `5c3f4c3` closes C26's clock dependency, but applying the remaining rulings to
the production/save boundaries exposes three narrower executable-contract gaps. None is a balance
literal; each changes the wire, persistence, or replay grammar.

### C28 — C24 calls the transform grammar closed without enumerating it

C24 accepts per-source exact-key arms and a fixed integer operation order, but neither C19 nor C24
lists those key sets, the operation union, negative rounding, or whether composition is one row or
a graph. A loader still cannot distinguish the intended grammar from an invented one.

**Proposed contract:** one row is
`{destination,destination_class,ranked,source,transform}`. Close `source` to
`{kind:"literal",value}`, `{kind:"tier"}`,
`{kind:"purchased_generator_count",generator_id}`,
`{kind:"founder_carry_counter",path}`, or
`{kind:"attended_quality_grade",minigame_id}`. Close `transform` to
`{offset,multiplier,denominator,clamp_min,clamp_max}` with positive denominator and mathematical
floor, applied exactly source → add offset → multiply → floor-divide → clamp. Destination IDs are
unique; Phase A has no transform graph. Tier taint is the source kind and rejects
`ranked && destination_class=="power"`. The owner may amend these names, but code needs one literal
grammar.

### C29 — Founder rating is named but has no state or replay schema

C21/C25 require a Founder rating transition in the atomic resolve transaction. No RFC names its
persisted fields, initial value/bounds, season-fact union, Founder save/table owner, resolved-input
arm, receipt, or event payload. Implementing it would invent Founder state and a new
`ApplyFounderLogged` arm.

**Proposed contract:** store ratings in a dedicated Founder-owned table keyed by
`(founder_id,minigame_id,season_id)`, with exact safe-integer rating and an append-only rating fact
row keyed by session. The Founder log receives a closed `minigame_resolution.v1` arm containing
the session ID, certified-result hash, old/new rating, and sorted season facts; its receipt/event
are exact projections of that arm. If rating should instead live in the Founder save, that requires
a Founder-axis save activation contract and must be stated explicitly.

### C30 — The faucet policy and idempotency row remain structurally incomplete

MP3 names sends/cap/conversion but not the credited resource; C5/C22 require a carried conversion
remainder and session idempotency without naming their columns. The locked transaction cannot be
written or replayed from the current prose.

**Proposed contract:** payout rows add exact keys
`{resource_id,attended_window_ms,sends_per_window,per_send_cap,conversion_ppm}`. Persist
`minigame_faucet_windows(founder_id,minigame_id,window_index,sends_used,credited_total,
conversion_remainder_ppm)` plus an immutable `(session_id)` application row carrying input score,
credited amount, forfeited amount/reason, and post-remainder. The session application row is the
idempotency authority; all integer bounds are `[0,MaxExactInteger]`. Bot/NPC reduction remains in
the already-ruled fallback row and runs before conversion.

## Post-ruling implementation blockers (Codex review, 2026-08-04)

C1–C11 settle the architecture and remove the Clout/UI/live-PvP contradictions. They do not yet
provide several exact schemas and arithmetic choices their own proposed contracts require. The
following are the narrow remaining contracts needed before code can have one correct shape.

### C12 — The normative body still contradicts three rulings

MP1 still declares `mode: solo|async_snapshot|live_pvp`; MP5 still requires
`unlock_condition_ref` and `era_skins`; AC6 still requires both combat engines. C8 reserves but
rejects live PvP, C9 replaces the unlock grammar/removes skins, and the lane engine is unimplemented.

**Proposed contract:** reconcile decision sites: Phase-A mode is exactly `solo|async_snapshot`
(the wire reserves no accepted `live_pvp` value until its owner RFC); registry unlock is exactly
the UI `always|fact_equals` union; remove `era_skins`; AC6 requires the fixture tenant and duel
adapter only, with lane registration moved to the lane RFC.

### C13 — Session persistence still has no executable row/command contract

C2 names Postgres, claim tokens, and one transaction but not the row keys/columns, transition
commands, revision/hash semantics, allowed concurrency, expiry, retention, or lock order.

**Proposed contract:** define `minigame_sessions` exactly with UUIDv7 session/command IDs,
account/founder ownership, minigame/mode, engine/constants identity, genesis/snapshot/result bytes,
status, revision, attended cursor, claim token/lease, paid revision, and created/updated/expiry
timestamps. Define exact create/play/abandon/resolve server commands and the closed rejection
taxonomy. Reuse canonical SHA256 idempotency over `(session_id,command_id,payload)`. Declare the
maximum active sessions per founder/mode, lease duration, expiry/retention literals, and lock order
relative to founder/company save streams.

### C14 — Tenant descriptors and transition results remain nouns

C3 accepts typed descriptors but supplies no exact descriptor rows, byte envelopes, result/event
union, engine registry interface, or error categories.

**Proposed contract:** enumerate the descriptor JSON and Go interface literally: schema versions,
maximum byte sizes, command/snapshot/result schema references, engine version, allowed modes and
event kinds; transition input/output envelopes with exact keys; and the deterministic rejection
union. State whether snapshot/result bytes are canonical JSON or opaque bytes plus content type.
The fixture tenant's literal schemas/vectors become the conformance oracle.

### C15 — Scaling is still not loader-provable

C4 says typed rows exist but neither lists their exact union nor maps destinations to
`power|breadth|presentation`. “Bounded exact integer” has no bounds/formulas.

**Proposed contract:** provide the exact source union and field names for the Phase-A arms accepted
in C4, per-arm integer range/formula, composition order, destination descriptor, duplicate rule,
and the literal fixture tenant destination registry. Define taint propagation: any tier-derived
source reaching a ranked `power` destination through any composition path rejects.

### C16 — The attended-day faucet clock and arithmetic are ambiguous

An “absolute grid” over attended time needs an origin and persisted cursor. C5 also says declared
rounding/overflow saturation without choosing the rounding or distinguishing a configured payout
cap from numeric overflow. Bot reduction and counter reset order remain open.

**Proposed contract:** identify one monotonically persisted Founder attended-time total and define
window index/origin exactly; if no such authority exists, add it with its save owner/migration.
Spell the checked integer formula and rounding (`floor` or carried remainder), numeric-overflow
behavior (reject/invariant—never silently saturate), order of bot multiplier/conversion/per-send/
window caps, and the exact counter/session idempotency row. Catalog values remain balance data, but
all required keys/ranges and reason-key flow are normative.

### C17 — Server-authored payout has no production transaction envelope

C6 correctly forbids client score claims but does not define the internal command, Company stream
event/receipt payload, replay-input row, or how a minigame DB transaction atomically commits a save
stream revision. “Inside resolve transaction” is not a lock/order/API contract.

**Proposed contract:** provide exact `apply_minigame_result` internal payload/resolved-input/
receipt/event schemas; declare session→founder→company lock order (or the existing canonical order
if different); identify the one repository method owning session resolution + run-log/save/events
+ payout-counter commit; and require fault injection after every write. Replay re-derives the
governed credit from immutable genesis/result/policy bytes and compares receipt/events bytewise.

### C18 — Fallback and offline-quality rows still lack literals/formulas

C7 calls the union exact without enumerating its keys. C10 names a ppm grade/timestamp and neutral
floor but no grade source, grid origin, decay formula/remainder, floor/cap, reset/scope, or
automation destination. These choices change state and replay.

**Proposed contract:** enumerate exact `solo|bot|npc_partner` rows, bot profile/version/policy,
ranked eligibility and payout fields, and the Phase-A fixture values. Either defer
`offline_quality` at loader level, or provide its complete catalog/state/transition contract and
one named neutral automation slot; prove no-play converges to neutral and never reduces canonical
offline production below 90%.

## Open questions

- Live-PvP match-actor lifecycle detail (matchmaking, bot backfill timing) — the PvP-service RFC
  owns it; this platform defines the `live_pvp` mode boundary.
- Whether the arcade (design/03 §9) toys need full sessions or a lightweight score-send-only path
  (recommend: lightweight — a Snake high score doesn't need match state).

## Post-foundation implementation blockers (Codex, 2026-08-04)

The independently approved session/tenant boundary implements the portions whose shapes are
literal. The remaining slices cannot yet be implemented without choosing mechanics that C15–C18
call exact but do not enumerate. These blockers supersede no ruling; they identify the missing
last-mile contracts revealed by applying those rulings to the shipped production/replay surfaces.

### C19 — The scaling-source union still has no wire shape or executable formula

C4 names five source concepts (`literal`, tier, purchased-generator count, Founder carry counter,
attended-quality grade), while C15 says the source union, field names, bounds, formulas,
composition, and taint propagation are exact. No normative section supplies those row keys or
formulas. A loader cannot distinguish a valid source row from an invented one, resolve it from the
shipped save owners, or export its formula artifact.

**Proposed contract:** enumerate each source arm as an exact-key JSON object, including its state
owner/path, integer transform (offset, multiplier/denominator, rounding, clamp), allowed range, and
whether tier-taint propagates through it. Enumerate the composition row and duplicate rule. Keep
all numeric literals catalog-owned, but make the formula grammar normative. The loader rejects any
ranked row whose transitive source graph is tier-tainted and ends at a `power` destination.

### C20 — Replay is required, but session commands are not persisted

C11 requires replay of every minigame command and byte comparison of snapshots/results. C13's
normative table stores only the mutable session head; it has no append-only command rows, command
IDs, committed sequence, or immutable result history. The current pure transition is replayable,
but there is no authoritative input corpus from which a verifier could replay it.

**Proposed contract:** add an immutable `minigame_session_log` keyed by `(session_id, revision)`
with canonical command bytes, exact engine identity, pre/post snapshot bytes or hashes, terminal
result bytes, and commit timestamp. The same transaction that advances/resolves the head appends
exactly one row; rejected commands do not append. Declare whether the genesis row is revision 0 or
stored only on the session, and make verifier verdicts/defer semantics reuse the run verifier's
typed taxonomy rather than an error string.

### C21 — The server-authored payout transition still has no production contract

C17 names `resolve_minigame_session`, but does not enumerate its canonical payload, replay inputs,
receipt, event kinds/payloads, rating/season fact destination, or the production method that owns
the Company-stream commit. The existing `ApplyLogged` boundary accepts only the closed public
intent union. Implementing this now would either add a hidden side write or invent a new replay
schema.

**Proposed contract:** specify exact-key internal command, replay-input, receipt, and event
envelopes; add a named server-only `ApplyLoggedMinigameResult` (or explicitly widen `ApplyLogged`)
that consumes only the certified result plus pinned policy bytes and returns the ordinary
state/receipt/events triple. Name the rating/season state owner. One repository method locks
Company then session and commits save revision, run log, events, faucet counter, and terminal
session mark with fault injection after every write.

### C22 — The attended-window authority and faucet counter are still ambiguous

C5 says attended-time absolute-grid day boundaries; C16 says origin `run_started_at` and a
persisted cursor per session. Neither identifies whether the quota resets on Exit, which persisted
Founder/Company value determines the current window, the counter row/key, or the exact order of
bot reduction, ppm conversion, per-send cap, window cap, remainder, and numeric saturation. These
choices change payout and replay.

**Proposed contract:** name one monotonic persisted attended-time authority and its reset scope;
define `window_index`, counter/session-idempotency keys, the complete checked-integer operation
order, remainder ownership, and both configured-cap and overflow receipts/reason keys. The policy
row supplies literals only after that structural formula is fixed.

### C23 — Fallback and offline-quality rows are called exact but remain partial

C18 gives optional `bot_ref` and reduction fields but omits the bot policy/version, NPC profile and
control-boundary keys required by C7. It also names offline-quality state without the score-to-grade
mapping, persisted owner, exact decay operation/remainder, or automation destination. A strict
loader cannot accept or reject an arm consistently from this text.

**Proposed contract:** enumerate exact-key `solo`, `bot`, and `npc_partner` arms and the exact-key
offline-quality policy row. Bot/NPC identity and deterministic policy version enter session
genesis. Offline quality names its persisted owner, score fact source, fixed-grid decay equation,
remainder, neutral floor/cap, and registered automation slot. Until these rows are owner-ruled, the
production loader accepts no fallback/offline-quality policy and therefore enables no production
minigame artifact.

## Acceptance blockers (Codex review, 2026-08-04)

The platform boundary is necessary, but the draft is not executable yet. In particular, its reward
owner predates the ruled Achievements correction, and its session/result nouns do not define a
server-authoritative persistence or replay contract. The following proposed closures keep the
platform generic without inventing any minigame.

### C1 — The payout contract reintroduces the forbidden Clout faucet

MP4 says minigames pay “Clout (achievement/rating).” Achievements C1 established that Feed/Social
owns Clout's single mint and no reward system may create a second source. Minigame Platform cannot
depend on an “Account/Clout foundation” that does not exist or mint Clout through a generic hook.

**Proposed contract:** remove Clout from MP4 and from this RFC's dependencies. Phase-A platform
payout arms are only owners that exist when implementation lands: an economy-resource credit with
a catalog-declared resource ID, and typed minigame rating/season facts that are non-resource state.
Feed/Social may add a social-activity Clout arm by successor RFC through the closed reward union;
the generic platform never gains a free-form resource/effect row.

### C2 — A session has no persisted authority or concurrency contract

MP1 lists fields and arrows but not where they live, who may advance them, whether multiple
sessions may coexist, how retries behave, or what prevents two resolve/payout workers from paying
one result twice. An in-memory match actor cannot be the authority for async snapshots or survive a
restart.

**Proposed contract:** define an append-only session command log plus one transactional session
head keyed by UUIDv7 `session_id`. The closed states are `created → active → resolved → paid` plus
terminal `abandoned`; transitions use `(session_id, expected_revision, command_id)` with the
existing canonical-payload/idempotency-hash discipline. Resolve records immutable result bytes;
payout claims that result in the same transaction as its authoritative credit/event writes and is
idempotent by session ID. Exact account/founder ownership, maximum concurrent sessions per mode,
abandon/expiry rules, retention, and lock order must be literal contracts, not handler choices.

### C3 — `engine_ref` does not define a tenant interface

There is no closed command/snapshot/result descriptor, engine-version identity, deterministic
error taxonomy, or proof that combat adapters and the fixture tenant implement the same boundary.
An arbitrary engine callback could read ambient state or return an unvalidated payout.

**Proposed contract:** register a typed tenant descriptor containing named command, snapshot, and
result schema versions plus an immutable engine version. The platform calls a pure deterministic
transition `(snapshot, canonical_command, frozen_inputs, seed/substream) →
(snapshot', result|null, events)`; the tenant cannot access DB, clock, network, or payout APIs.
The platform validates bytes before and after every call and owns the closed rejection taxonomy.
Tenant events are namespaced and registered like production events. Payout derives only from the
validated immutable result through MP3—never from tenant-supplied resource deltas.

### C4 — Scaling axes are labels, not functions

`reward`, `challenge`, and the other names have no row shapes, source fields, integer ranges,
composition order, or ranked-power binding grammar. The loader cannot prove the Fairness Law from
an undeclared relationship between a scaling value and an engine field.

**Proposed contract:** version a closed scaling-source union over implemented committed state
(literal, tier, purchased generator count, Founder carry counter, and attended-quality grade),
each producing a bounded exact integer under a published formula. A tenant declares each input's
axis and exact destination from its schema-owned destination registry. Ranked destinations are
classified `power|breadth|presentation`; ranked rows reject any non-neutral tier-derived path to
`power`. Session creation resolves the complete sorted input map once and stores those canonical
bytes in genesis; unknown sources/destinations and duplicate bindings fail catalog load.

### C5 — The faucet governor lacks time, arithmetic, and persistence semantics

“Per day,” score conversion, cap consumption, bot reduction, and forfeit are not defined precisely.
Wall-clock day boundaries create timezone/deploy ambiguity; a resolve retry can spend the cap twice;
and `score × conversion_ppm` may overflow or round differently across runtimes.

**Proposed contract:** the catalog row names the payout resource and exact integer score domain,
`sends_per_window`, `per_send_cap`, `conversion_ppm`, bot multiplier, and cap reason key. Windows
are UTC epochs `floor(server_ms / 86_400_000)`. Conversion uses checked integer arithmetic with a
persisted remainder only if the owner wants carry; otherwise it declares floor once. A
founder/minigame/window counter is updated in the payout transaction and keyed by session ID so
retry is a no-op. The receipt reports gross, applied, and forfeited amounts plus the visible reason
key. All literal values are balance data and joining the epoch artifact set is a mint.

### C6 — “Payouts are intents” permits a client to claim a result

The current public intent envelope accepts player commands, not server-certified minigame results.
Adding `claim_payout {score:...}` would violate the server-authoritative law; hiding an internal
side write behind the word intent would violate replay/event ownership.

**Proposed contract:** add a server-internal `apply_minigame_result` command that is never decoded
from the player HTTP intent route. It consumes `(session_id, immutable_result_hash)` after the
session repository locks and verifies `resolved`; the production/economy transition receives only
the already-governed applied credit and records its source session in replay inputs, receipt, and
events. The session `paid` mark and Company stream commit share one declared transaction/lock order
or an outbox projection with an exactly-once claim—choose one explicitly and fault-inject every
boundary.

### C7 — AI fallback and ranked identity are not closed data

“Solo-by-design, bot backfill, or NPC-partner” does not state exact rows, timing, bot identity,
rating eligibility, or how a bot result is replayed. “Declared reduced rate” has no literal owner.

**Proposed contract:** make fallback a closed union with exact fields. Solo has no peer; bot
backfill declares deterministic bot profile/version, backfill deadline, ranked eligibility, and
payout multiplier; NPC partner declares profile/version and control boundary. Bot decisions use
the tenant's seeded substream and the same legal-command validator. Every session genesis records
the selected fallback/profile version. Ranked eligibility and payout reduction derive from that
record, never from a client flag.

### C8 — Live PvP is simultaneously deferred and required

The open questions defer match actors/matchmaking, while MP1 and AC1 require `live_pvp` through the
composed server. Without disconnect, reconnect, authoritative-clock, and result-finalization rules,
the state machine cannot support that mode honestly.

**Proposed contract:** Phase A implements `solo` and `async_snapshot` only and reserves—but rejects—
`live_pvp` at registry load until a PvP Match Service RFC owns actor lifecycle, matchmaking,
disconnect/reconnect, clock, spectator/privacy, and bot-backfill timing. Combat adapters can prove
tenant conformance against deterministic solo/snapshot fixtures now. If live PvP must ship in this
RFC, those contracts become blockers here rather than an open question.

### C9 — Registry presentation and unlock fields repeat withdrawn UI paths

`unlock_condition_ref` has no grammar and `era_skins` contradicts UI Foundation: surfaces use one
token theme, never per-feature skins. The combat children are also not implemented tenants today,
so AC6 overclaims their readiness.

**Proposed contract:** replace the fields with the UI-ruled unlock union
`always|fact_equals` over the authoritative shell/snapshot fact manifest; remove `era_skins`.
Registry rows carry a copy key and optional presentation capability IDs only. AC6 uses a fixture
tenant first; duel/lane registration becomes each combat child's acceptance item when that engine
is implemented, not a platform-foundation prerequisite based on draft code.

### C10 — Offline quality has no observable or decay formula

“Recent performance,” “quality tier,” and “~31 days” define neither stored state nor a deterministic
transition. It also risks turning an optional minigame into mandatory labor if automated output
decays toward zero.

**Proposed contract:** either defer `offline_quality` entirely, or specify its owner catalog,
score-to-grade thresholds, exact fixed-grid decay/remainder, floor, cap, save scope/reset, reason
keys, and neutral no-play behavior. The floor must preserve the base offline-production promise;
quality may improve a named automation slot but never reduce the canonical 90% offline rate. Until
those literals are owner-ruled, the scaling-source union rejects `offline_quality`.

### C11 — Artifact/replay identity and migration scope are absent

The registry, tenant schemas, fallback profiles, faucet policy, and session genesis all affect
verification, but none joins epoch identity or the verifier. No DB migration or cross-runtime
corpus is named.

**Proposed contract:** one `minigames` artifact contains the sorted registry and referenced policy
rows; tenant engine versions/schemas are immutable code identity. Session genesis pins constants
hash, artifact bytes/hash, engine version, seed/substream, scaling inputs, fallback selection, and
initial snapshot. A shared verifier replays every command and compares snapshot/result/event bytes.
The implementation lands append-only migrations, crash/fault integration tests, ≥50-command
Go/TS (or single-runtime tenant) corpus, formula artifact rows for scaling/payout, and a protocol-
compliant balance mint before any production tenant or payout is enabled.

## Implementation blockers C31-C33 (Codex, 2026-08-05)

The exact scaling, fallback, and payout-conversion kernels are implemented without production
balance rows. Composition now reaches three wire/persistence contradictions that the accepted
rulings do not resolve.

### C31 — Offline quality still has no exact curve or attended-state wire

C27 names `grade_curve` but never enumerates a curve row. C18 stores `last_session_at` even though
the accepted decay clock is Founder attended time, not wall time. A loader cannot reject invented
threshold/grade keys, and replay cannot derive attended decay from a wall timestamp.

**Proposed contract:** the exact policy row is `{score_fact,grade_curve,decay_grid_ms,
decay_ppm_per_grid,neutral_floor_ppm,automation_destination}` where `grade_curve` is a nonempty
array of exact `{score_at_least,grade_ppm}` rows. Thresholds are strictly increasing from zero;
grades are nondecreasing and within `[neutral_floor_ppm,1_000_000]`; duplicate thresholds reject.
Persist `{grade_ppm,last_attended_ms,decay_remainder_ppm}` under Founder/minigame identity, replacing
the wall-clock `last_session_at`. Values remain balance data; these keys/order rules are wire.

### C32 — The exact payout row does not identify its certified score or cap reason

C30's four keys select a resource and conversion, but a tenant result contains a sorted array of
typed score facts. Nothing says which fact feeds payout. The configured-cap path also requires a
visible reason key while the exact row provides none. Choosing `score.total` or a copy key in code
would invent protocol/content ownership.

**Proposed contract:** extend the exact payout row with `score_fact` and `cap_reason_key`.
`score_fact` must be a declared result fact for the tenant descriptor and unique in a certified
result; missing/negative values deterministically reject payout before writes. `cap_reason_key`
must resolve in the copy manifest. The published formula names the selected fact; no convention or
first-array-element fallback exists.

### C33 — C22 and C30 assign conversion carry to incompatible owners

C22 makes quota cross-run on a Founder/minigame attended-window row. C30 calls the conversion
remainder a session-state field. A session-local remainder cannot carry into the next session, and
no exact window/application row exists for atomic retry. The transaction cannot preserve carry and
idempotency simultaneously from the current text.

**Proposed contract:** the authority is
`minigame_faucet_windows(founder_id,minigame_id,window_index,sends_used,credited_total,
conversion_remainder_ppm)` with a composite primary key. An immutable
`minigame_faucet_applications(session_id PK,founder_id,minigame_id,window_index,input_score,
reduced_score,converted_units,credited_units,forfeited_units,cap_reason_key,
remainder_before_ppm,remainder_after_ppm)` is the retry authority. The session may copy the final
remainder in its receipt/result but is not the carry owner. Lock order remains Founder → Company →
session → window; all rows and both stream revisions commit together.

## Implementation blockers C37-C40 (Codex, 2026-08-05)

The Founder v17/v18 activation chain and the C17 pet-catalog parity remediation are independently
approved. The next requested slice is the server-certified resolve composer. A source walk reaches
four remaining executable-contract gaps; none can be filled from a balance literal alone.

### C37 — The pinned minigames artifact is only an activation index

`minigame.LoadCatalog` currently accepts only `{schema_version,minigame_ids,rating_seasons}`.
C11 requires the registry plus payout, fallback, scaling, offline-quality, and rating policy to be
hash-pinned, but the replay bundle has no bytes for those policies. A composer supplied an
in-process `PayoutPolicy` would therefore execute deploy-current policy while replay resolves only
the ID/season artifact.

**Proposed contract:** before any resolve path is enabled, replace the pre-mint v1 artifact with
one exact `{schema_version,rating_seasons,minigames}` object. Each sorted minigame row is exact
`{minigame_id,engine_ref,engine_version,modes,result_score_fact_ids,scaling,payout,fallback,
offline_quality,rating_policy,unlock_condition}` and embeds the already-ruled closed sub-objects;
no policy reference resolves outside these bytes. `minigame_ids` is derived from the rows rather
than stored twice. Empty `minigames` remains the valid pre-content artifact. Go and TypeScript load
the same shared fixture, and `CatalogBundle.Minigames` is the sole live/replay policy resolver.

### C38 — Atomic resolution has no retry receipt or persistence-owned coordinator

C25 requires Founder → Company → session → faucet-window locking and says a retry returns the one
committed receipt. The session table stores only the certified tenant result, not the payout/rating
receipt, and the current `minigame.Service.ResolveTx` owns only Company/session validation. Neither
package can compose both save streams without duplicating private save-store writes or nesting an
exported transaction.

**Proposed contract:** add `save.Store.ApplyMinigameResolutionTransaction`, the narrow multi-stream
coordinator analogous to Exit. It owns Founder then Company stream locks, revisions/log sequences,
events/outboxes/retention, and one transaction callback that finalizes the already-claimed session
and faucet window. Session ID is the idempotency key and internal intent ID. A migration adds exact
`resolution_receipt jsonb` plus `resolution_company_revision` and `resolution_founder_revision` to
the terminal session row; all are null before resolution and immutable/non-null afterward. A retry
of a resolved session returns those bytes without re-running a tenant or consuming quota. Fault
injection after every write proves all-or-none across both revisions, both logs/events, window,
terminal session, and receipt.

### C39 — The two replay arms and visible event/receipt bytes are not closed

C21 names `resolve_minigame_session`, `minigame_resolution.v1`, and two event kinds, but does not
enumerate the Company resolved-input arm, Company replay input, terminal receipt, resource event,
or rating event payload. `ApplyLogged` consequently cannot reproduce the credit, and
`ApplyFounderLogged` has no minigame arm to reproduce v17 rating/offline-quality state.

**Proposed contract:** close one canonical internal payload
`{kind:"resolve_minigame_session",session_id,result}` with `intent_id=session_id`. Both logs bind
the same `certified_result_hash`. The Company resolved arm records the exact payout policy identity,
selected score, complete faucet before/after application, credited Decimal delta, and Founder-log
coordinate. The Founder `minigame_resolution.v1` arm uses C36's exact
`{session_id,certified_result_hash,old_elo,new_elo,season_member}` fact plus old/new offline-quality
state and the frozen Founder-attendance sample. The shared terminal receipt names both committed
revisions, credit, configured-cap forfeit/reason, rating change, and quality change. Register exact
`minigame_resolved.v1` and `minigame_rating_changed.v1` payloads in the closed event registry; both
runtimes compare receipt/event bytes and order in one sequential fixture.

### C40 — Rating and offline-quality mutation arithmetic is still ambiguous

The certified result carries `rating_delta`, while C36 still names a catalog K-factor; no rule says
which is authoritative, how floor/ceiling and provisional counts apply, or what `null` means.
Offline quality persists a decay remainder but C31/C34 never give the remainder equation or say
whether a new result decays-then-replaces, takes max, or composes with the old grade.

**Proposed contract:** make the server-certified tenant `rating_delta` authoritative for Phase A;
the platform checked-adds it to old Elo, saturates only at catalog floor/ceiling, and increments
`games_counted` once. `null` means unrated: Elo/count are unchanged but the Founder resolution and
quality update still commit. Remove platform K-factor from this RFC (an engine may use it inside
its versioned certified-result calculation). For offline quality, first integrate decay from the
stored attended watermark to the frozen sample with
`numerator=elapsed_ms*decay_ppm_per_grid+remainder`, `decay=floor(numerator/decay_grid_ms)`, and
`remainder=numerator mod decay_grid_ms`, saturating at the neutral floor; then replace the decayed
grade with the current certified score's catalog grade and zero the remainder at the sample
watermark. Shared boundary vectors cover `null`, negative delta, Elo floor/ceiling, sub-grid carry,
multi-grid decay, and same-sample retry.

## Changelog

- 2026-08-03: created (draft) — the platform.
- 2026-08-05: C19-C23 ruled - closed scaling-source union, append-only session command log (run_log pattern), resolve payload/events, Founder-attended-cursor faucet quota, exact fallback rows. All structure ruled, numbers deferred.
- 2026-08-04: C12–C18 ruled — body reconciled to the C1–C11 rulings; session table = verification-queue pattern, payout = Exit-style multi-write server transition, faucet = provision-grid clock, offline_quality decays to neutral floor; all STRUCTURE ruled by reusing shipped patterns, all NUMBERS deferred to balance data.
- 2026-08-04: C1–C11 ruled — Clout payout faucet removed (my error, caught), DB-authoritative sessions with queue-style idempotency, server-certified (not client-claimed) payout, live_pvp deferred (solo+async only ship), era_skins removed (UI law), offline_quality decays to neutral floor not zero, epoch/verifier identity named. Accepted, scope narrowed.
- 2026-08-04: Codex acceptance review found C1–C11. The draft is not implementable yet: Clout
  ownership contradicts the ruled single faucet, and session, tenant, scaling, payout, fallback,
  live-PvP, unlock, offline-quality, and replay contracts require owner rulings.
- 2026-08-04: post-ruling implementation pass found C12–C18: design ownership is settled, but the
  active body and exact persistence/wire/arithmetic contracts still need reconciliation/literals.
