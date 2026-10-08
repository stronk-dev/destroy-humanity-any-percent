# Server Garden

Server Garden (`rfc/minigame-server-garden.md`) is a persistent, Founder-scoped, wall-clock
crossbreeding garden. It is **not** a Minigame Platform session: it survives every Exit, never
occupies the one-active-session slot, and never blocks Exit. It reuses the platform only through
the faucet governor and payout kernel. Everything below is fixture-first: no epoch pins
`server_garden` yet (SG13).

## Artifact

- `server_garden` (schema v1) is an optional artifact on the scalar Founder chain. It requires
  `cosmetics` and a pinned Fiscal artifact, and it pins **Founder v25**.
- `server/garden` (loader) and `client/src/garden/catalog.ts` enforce every SG1 rule:
  - exact keys, nonnullable catalog fields and exact safe integer tokens (decimal/exponent
    spelling is refused even when it would parse to an integer);
  - ID grammar and raw-byte sort;
  - the domains;
  - reachability of every species from the starters;
  - geometric satisfiability;
  - a non-decreasing dimension ladder ending at the visible hardcap;
  - the evaluation budget;
  - the platform payout row;
  - Fiscal unlock and host rows;
  - copy keys;
  - append-only IDs across epochs.
- The fixture is `balance/testdata/server-garden/fixture-v1.json`. It pairs with
  `fiscal-fixture-v1.json`, the production Fiscal artifact plus
  `minigame.server_garden` (cost 3).
- The shared raw admission corpus is `testdata/garden/catalog-raw-fixtures-v1.json`:
  14 valid controls and 22 invalid inputs applied as identical literal replacements in Go/TS.
  The TS scanner rejects unfinished strings without walking past the input; the Node test
  contains regressions in a child process, while all three native browser engines execute
  the corrected malformed inputs directly. Catalog null rejection does not change the
  intentionally nullable SG2 save fields. This correction awaits designated review; it does
  not establish production minting, full Garden acceptance or release readiness.

## State and clock

The actual New-Founder initializer is tested with and without a Garden pin. Both real Service
Exit commands (`wind_down`, `accept_exit_offer`) retain literal nonempty Garden bytes on
Postgres, including dormant/mature/immature plots and all clock/lockout metadata, and advance
to the next run. Their retries are unchanged; both history axes verify against their own
event ownership. The deliberately due automatic Fiscal prefix is verified on Founder history,
not incorrectly aggregated into Company history (RP-207). TS also carries the same literal
state through Company and Founder replay, with missing/pre-v12 carry refusals.

**Unresolved epoch constraint (RP-222):** promoting an uncollected species to starter is
currently admitted by catalog/transition validation but invalidates unchanged carried state
under the next catalog. Ordinary numeric retunes pass. The author/owner must reconcile
starter-set evolution with SG2's starter-superset and byte-identical carry rules before such
an epoch can ship. No grant/migration exception or starter-set restriction is implied here.
These bounded G2/G3 witnesses await designated review; they are not public mint, full Garden
acceptance, a complete epoch-migration proof or release readiness.

- `save.State.ServerGarden` (Founder v25) holds the hidden salt, the tick anchor and sequence, the
  substrate, the lockout stamp, the plots and the seed collection. The stage is derived: mature
  means `matured_effect_ppm != null`. It decodes strictly, with exact keys and no zero-value
  defaults.
- Null is accepted only for the root salt/anchor/lockout stamp and the plot's maturation effect;
  tick sequence, coordinates and age must be actual integers. The Go codec checks nullability
  before typed decoding rather than defaulting null to zero. `state-admission-v1.json` has
  eight valid controls, 21 shared typed refusals and two Go-only raw duplicate-key refusals.
  The TS parser receives objects, so its duplicate rows explicitly demonstrate JSON.parse's
  loss of raw-key evidence; they are not credited as TS raw-ingress rejection. This bounded
  SG2 correction awaits designated review and changes neither Founder v25 nor replay versions.
- The garden activates at the new-run boundary or at New-Founder initialization, with every starter
  collected. Every later Exit carries it byte-identically. Replay-inputs v12 carries it in the
  Founder extensions.
- **The clock is the server wall timestamp of the Founder command.** The advance is lazy and
  integer-only:
  - draws are keyed by the absolute tick sequence (partition-invariant);
  - a fixed-point skip applies;
  - a visible 24-hour catch-up hardcap reports `catchup_forfeited_ms` with its reason key;
  - a clock regression is a no-op.
- The existing SG2 safe-integer counter domain also governs advance output. An unlocked,
  anchored forward advance computes pending ticks under the unchanged catch-up cap before
  salt, plot or clock mutation and refuses if the exact next counter exceeds MaxExactInteger.
  Zero-tick and last-safe advances remain valid; nothing is clamped or silently rounded.
  `garden-clock-boundaries.test.ts` / `clock_boundaries_test.go` use independent BigInt/math/big
  arithmetic for 49 cases across every substrate's tick/cap boundaries, maximum server stamp,
  locked/initializing state and counter frontier. These codec-valid synthetic frontiers are
  not evidence of reachable-player exploitation or full Garden acceptance. RP-219's scoped
  repair awaits designated review; save/replay schemas and clock/payout policies are unchanged.
- The hidden salt is drawn by the server when the first unlocked advance runs, frozen in that row's
  resolved inputs, and never projected.

## Independent tick witnesses

`testdata/garden/tick-witnesses-v1.json` is a separate, hand-derived sixteen-case SG4
population, not output of the Go engine/corpus generator. Go and TS enter through actual
catalog/state admission, execute the real advance and compare complete post-state/summary
bytes against literals. Cases cover growth before the diagonal census, dormant/immature
exclusion, parent minima and crossbreeding, recipe/plot order, effective-zero eligible draw
consumption, strict threshold and adjacent success, absolute tick keys, Chaos selection,
frozen effect, lowered maturation threshold and next-tick newborn growth. Literal founder
base and eight bounded draws at each of ticks 1/2 bind both production PRNG implementations.

Actual source mutations of diagonals, threshold, tick key and recipe ordering fail the new
population and are restored exactly. This is bounded pure-transition evidence, pending
designated review, not a statistical RNG, exhaustive rejection-sampling, full SG4/Garden,
coordinator, public-flow or mint claim. Cumulative saturation is exercised, but removal of
that upper clamp alone is observationally equivalent for nonnegative probabilities and
draws below one million; it is not claimed to be separately discriminated.

## Commands

- `garden_plant`, `garden_uproot`, `garden_set_substrate` and `garden_harvest` are Founder intents
  on `/api/v1/intents`. Their schemas are exact; a client time, tick, outcome or salt field is
  terminal `invalid`.
- **The advance pre-step** runs after the Fiscal sweep and before the body, for exactly the garden
  commands plus `spend_fiscal_credit` (a property test proves this set is sufficient). An applied
  trigger's receipt carries `garden_advance`, and `garden_advanced.v1` precedes the command event
  when visible.
- **Harvest** commits through the Founder→Company coordinator:
  - `save.Store.ApplyGardenHarvestTransaction` accepts no caller timestamp and samples the
    existing Founder Postgres clock after locking Founder then Company;
  - the applicability probe and applied transition use that same locked timestamp; a refusal
    rolls back the coordinator before the ordinary DB-stamped Founder-only rejection path;
    a later maturity/revision race returns the existing conflict rather than a partial credit;
  - the Founder intent record is the exactly-once authority;
  - the faucet window is keyed `(founder, "server_garden", attended_day)`;
  - the credit to `company.cash` saturates;
  - both logs bind the same `harvest_hash`;
  - a zero harvest touches no window;
  - past the daily sends, units forfeit with `cap.minigame_faucet` and seeds are still recorded;
  - a rejected harvest is Founder-only.
- The six SG8 event kinds are in migrations 00082 and 00083.

`command-witnesses-v1.json` supplies fifteen separately declared direct command outcomes,
including overlapping invalid predicates, missing/immature harvest precedence, dormant
nonstarter uproot without seed collection, and lockout expiry at ±1 ms with exact partial-tick
disclosure. Both engines compare full state bytes without clone-and-discard hiding mutations.
Eight gate combinations check unlock precedence and the inert `unrelated` Soul mode.

`TestGardenDueGrowthRefusalIntegration` adds six real-Service/Postgres arms with admitted
old-anchor genesis: four ordinary refusals after ≥2 due ticks, a null-salt/25-hour catch-up
refusal and matched applied planting. It checks full persisted state/revisions, rejection-only
Founder log/intent/receipt outbox, no events/Company-log/faucet writes, unchanged retries,
hash conflicts and continuation. Replaying the actual stored row separately checks in-memory
rollback; Store's persisted refusal boundary does not substitute for it. Disabling transition
restoration fails all five replay-refusal checks while persisted state remains protected and
the applied control still works. These bounded tests await designated review; they do not
establish full Garden/public activation or exhaustively prove all transaction fault boundaries.

The bounded clock repair (RP-217) preserves other minigame-resolution timestamp policies and
Founder attendance/faucet arithmetic. Six real-Postgres mature/immature cases cover matched,
ten-second-lagged and twenty-four-hour-ahead handler clocks, stored-receipt retries, hash conflicts
and ordinary-command continuation; those injected server-clock disagreements are not client
time fields or a measured normal-host skew population. The repair awaits designated review;
it neither migrates existing history nor establishes public activation or release readiness.

The pure harvest arithmetic uses bounded int64 in Go and BigInt for the product in TS, before
converting the safe quotient to a number. `testdata/garden/harvest-boundaries-v1.json` supplies
18 literal result/hash/post-state cases: a 4×4 operand matrix, all 36 maximum-valued plots,
and a mixed frozen/zero-seed/dormant harvest retaining an unharvested growing plot. Expected
quotients use independent exact arithmetic, not production harvest output. The matrix includes
one demonstrated floating-product off-by-one. This is pure-engine evidence; it does not replace
the Postgres transaction, faucet, public player workflow or release checks. Its new test range
awaits designated review, with production and balance bytes unchanged.

The RP-227 rollback supplement compares complete rows across ten affected persistence
populations at all ten exposed coordinator fault checkpoints, including Founder genesis,
both logs, receipt/event outbox and actual revision pruning. The retention case first uses
six ordinary Service harvests, so a saved Company revision genuinely would be deleted;
after rollback it remains, and a clean harvest advances the oldest retained revision 3→4.
Each checkpoint uses an independent fixture and a clean-success control. Temporarily
committing on fault exit makes all ten cases fail; restored code passes twenty repetitions
(200 fault cases). The mature plants are seeded only before fixture stream creation, not
public player progression. The Service fixture has no live dispatcher, so every outbox
column is compared. Designated review remains pending. These grouped checkpoints are not
separate hooks after every individual SQL statement and do not establish full AC8/G5.

RP-228 strengthens the real-Postgres Company replay helper: its complete replayed state
must equal the actual saved Company head after zero, one or ten recorded credits, not just
match receipts/events. Adding unrecorded replay cash after forming unchanged output bytes
fails only the new head check. Go/TS verdict tests consume all four exact committed credit
shapes; honest nonterminal histories report `log_gap`, and individually altered command or
resolved hashes report `state_divergence`. Actual persisted Founder harvest receipt/event
hashes are also poisoned in copied evidence, without changing DB rows or the verified
original history. Independent receipt/event checks survive different disabled hash guards
in Go and TS; both guard removals still fail the new population. Twenty repeated actual
harvest/history populations pass. Designated review remains pending; this is not complete
terminal cross-runtime history, mature HTTP/default-player progression or full AC9/G5.

SG8's separate "ordinary resource event" has no named existing kind in the closed registry
(RP-229). Current live/replay emits `garden_harvest_credited.v1`; specification-author
reconciliation is required before treating that unnamed additional event as implemented.

## Read

`GET /api/v1/garden/current` (`get_current_garden`) returns `inactive`, `locked` or `active`. It
runs the advance on a discarded clone, and it never returns the salt, a draw or a future tick.
The real Service read obtains the active same-Founder saved head and Postgres millisecond
clock in one read-only statement, restoring against that head's pinned catalog. Account's
handler clock is not projection authority. Query/load failures return errors, never a
handler-time fallback; reads write no saves, events, logs, intents, outbox rows or faucet state.
The pure projection retains an explicit timestamp for deterministic tests/replay.

RP-223's bounded correction has six real-Postgres salted/unsalted normal/±24-hour handler-clock
arms, full persisted non-mutation and loaded-head checks, and cancelled/missing-source refusals.
Restoring handler-time projection fails the skewed-clock cases; removing the discarded clone
fails the full-Founder non-mutation oracle. Kernel 0.3.153 records this runtime correction.
Designated review remains pending; this does not establish full G6/Garden or public activation.

RP-226 adds a fixture-only composed HTTP witness, rather than treating the inactive response
as proof of active integration. Two actual HTTP-created accounts use the real token/registry/
Production/Postgres path. Empty v25 starter Genesis has zero Fiscal credit; real automatic
Fiscal reporting funds the public unlock, then plant/read/retry/substrate/uproot and immature/
forged-field/foreign refusals exercise the live route. Reads bracket database time and leave
both accounts' eight persisted row populations unchanged (outbox delivery metadata is explicitly
excluded); refused commands preserve all four Founder/Company heads. Complete Founder
histories replay, and served views/receipts plus scoped event/outbox payloads exclude hidden
keys and the actual salt. A schema-valid planted salt leak fires the enumerator, not just the
schema firewall; detaching the composed producer also fails. Private Founder replay inputs
are intentionally excluded from public enumeration. Twenty final repetitions pass against
real Postgres; new test range requires Claude. No runtime/epoch/content change, mature HTTP
payout, default DOM flow, real idle wait, full G4/G6/Garden or public-release claim follows.

## Client surface

`client/src/game-ui/garden/GardenSurface.svelte` is the SG10 surface:
- The Garden tab appears only when the latest current-context read is `locked` or `active`.
  Startup and mounted reads share a presence guard scoped to Founder ID and constants hash.
  A confirmed `inactive` removes the tab and returns a mounted Garden to Desk with heading
  focus, without overriding newer navigation. Failed reads do not establish absence.
  A Founder/content change re-probes presence and remounts the panel, discarding the previous
  grid and its pending read. If that removes focused Garden content or navigation, focus
  recovers to the new heading unless the player moved elsewhere. An ordinary Company change
  under the same Founder and bundle preserves the grid/menu; Garden remains Founder-owned.
- The grid is a `role="grid"` of native buttons with a roving tabindex and arrow keys.
- A native `?` disclosure opens `garden.codex` with Enter, Space or pointer input, including
  locked/error states. Its 44px control describes the always-visible `garden.why`; substrate
  curtain text remains visible and associated with its buttons. Help does not dispatch intents
  or trigger reads, and stays open/focused across an ordinary refresh. All buttons have at least
  24px targets. The codex is still the terse candidate "Seed collection", not completed help
  content; the new accessible help label is explicitly pending owner copy.
- Menu actions, Refresh, Harvest all and substrates have explicit native tab stops in DOM
  order, including WebKit. Tab/Shift-Tab reaches the seed and mature menus without changing
  the grid's single roving stop. Enter/Space uses the existing callbacks; closing or acting
  in a menu returns focus to its plot unless a newer focus choice has already won.
- Pending command controls remain focusable with `aria-disabled`, point to the existing polite
  status region's visible `common.pending` message, and guard native/programmatic activation.
  An open menu cannot submit or close itself through a pending command; ordinary Close still
  works. Substrate lockout remains natively disabled, distinct from temporary pending state.
- Command availability also inherits the host's Founder/transport readiness. Before handshake,
  during reconnect and while resync awaits a recovered subscription, the last read remains
  visible with `common.stale_note`; grid, menu intents, Harvest all and substrates are natively
  disabled and describe that reason. Callback guards also refuse explicitly dispatched events.
  Recovery does not submit refused input: the player must activate a control again.
- After a read or readiness change, a surviving usable control keeps focus. Removed controls
  hand focus to the nearest surviving tab stop in the same region, otherwise the heading; a substrate made
  unusable by lockout falls back to the heading. Recovery never overrides a newer focus
  choice or targets an inactive grid cell outside the single roving tab stop. Pending-only
  prop updates leave the menu's existing plot-focus handoff in charge.
- Stage and dormancy are text plus a border shape, never colour.
- Growth announcements go to a polite live region once per refresh.
- It re-reads at `next_tick_wall_ms` while visible and after every receipt, with no client
  simulation.
  The delay is the positive difference from the response's `server_ms`, including a
  sub-second remainder; there is no one-second polling floor or client epoch conversion.
  Nonpositive advisory differences yield to a one-millisecond timer turn.
- Only the latest-started read may replace the view/status/announcement or schedule the
  next read. Older success and failure completions are ignored, including a response that
  arrives after a newer receipt refresh. Unmount still discards late completions.
- The grid fits 320 CSS px with at least 24 px targets.
- Commands go through the Game UI's Founder-scoped `act()`.
- Nonterminal streamed receipts invalidate the Garden read as well as refreshing the
  main snapshot. The existing terminal guard still preserves the run-end screen;
  receipt bytes never patch or simulate Garden state locally.

All garden copy is candidate text (`copy/catalog/garden-candidate.json`): species names read
`PENDING OWNER NAME`.

RP-224's bounded surface correction has a controlled deferred-port/receipt-prop harness
around the real mounted component: active/locked/error/late-error ordering, sequential and
unmount controls, plus native Enter/Space, all four arrow directions/edge clamps, Tab to
Harvest, exactly-once callback and focus return, and pending-control refusal. Removing
either response guard or arrow navigation breaks its corresponding witness. This is
component-level native-browser evidence, not real HTTP/Postgres/default-host command
integration, timer/Page Visibility proof or full accessibility acceptance. AC13 remains
blocked on its successor; designated review of the new correction is pending.
The native input witnesses separate independently observed navigation walks from Enter/Space
action paths, retaining every intermediate focus and roving-tabstop assertion. This replaces
the long duplicated test traversal after recorded deadline failures (RP-225), without changing
test limits or production keyboard behavior. The final local full browser/performance lane
passes; earlier failures remain evidence, not a claim of general hosted reliability.

RP-230's separate timing correction is exercised by seven mounted-component/browser-port
cases per engine: exact 60,000/137 ms server-relative boundaries, hidden due/visible resume,
receipt-prop deadline replacement, null deadline, unmount, and stale-read recovery. Calls
pass through the actual generated authenticated GET adapter with no body; growth appears
only after injected DTO JSON, and no command callback fires. Tests join actual adapter
completion, not just fetch invocation. Timer-dispatch and visibility-guard severing each
fail their own populations. Virtual timers/emulated visibility and an injected fetcher
are explicitly not native wall-clock/OS-hide, real HTTP/Postgres or default-host receipt
proof. No AC13/Garden/public activation claim; the correction awaits Claude review.

RP-231's host supplement mounts the actual GameUIApp with the actual browser runtime,
generated authenticated Garden adapter, JSON response parsing and socket frame decoder.
Five DOM command journeys cover plant/uproot/single harvest/harvest-all/substrate, exact
UUIDv7/Founder-not-Company revision/fields, held-response pending and nonoptimistic state,
receipt-triggered reads and the next command's refreshed revision. Refusal and inactive/
locked visibility controls also run. A streamed receipt now rereads the mounted Garden;
before the repair, that arm failed in all three native engines despite the main snapshot
refreshing. Actual plant-callback and receipt-key severing each break their populations.
Responses and socket frames remain injected boundary fixtures, not real Go/Postgres/
WebSocket, native input, mature gameplay progression or default public activation proof.
No fixture exports or replacement runtime are used. This separate correction still
requires Claude's designated review; no full G7/AC13/Garden acceptance follows.

## Verification

The Fiscal presentation map includes `minigame.server_garden`, using the existing
Garden title and explanation keys. A pinned catalog that offers this unlock now
has a real Fiscal purchase control; this does not add the Garden to production
epochs or override the SG13 activation requirements.
The normal browser CI population includes `garden-fiscal-browser.test.ts`:
purchase callback/ID, unchanged copy, no optimistic ownership, and absent/owned/
unaffordable controls. Removing the actual presentation row fails both cases in
all three engines. These component fixtures do not establish real maturation.

`make test-garden-composed` is a manual fixture-only built-client/real
server/Postgres/WebSocket journey, outside push CI. It uses the exact existing
grown replay bundle and its unchanged three five-minute ticks; every gameplay
intent comes from a DOM control. Execution evidence and limitations are recorded
in the Garden planning log; adding the driver is not by itself a passing claim.
The first native fifteen-minute run reached actual ticks one and two but missed
the maturity objective. Full harvest/reload proof remains open. The browser
runtime currently has no session-renewal path despite the account's fifteen-minute
access-token lifetime; RP-234 routes that separate Account/Transport consumer gap.
The run did not emit its collected boundary errors, so its precise failing HTTP
status is not established. The instrument now emits those errors on failure.
RP-233 separately retains an intermittent next-plot interaction failure; a later
setup success is not a reliability verdict.

- `make garden-corpus-check`
- `make test-go GO_PACKAGES=./garden GO_TEST_FLAGS='-count=1'`
- `client/test/garden-harvest-boundaries.test.ts` and the shared literal boundary corpus
- `client/test/garden-clock-boundaries.test.ts` / `garden.TestGardenAdvanceClockBoundaries`
- `client/test/garden-engine.test.ts`, `garden-replay.test.ts` and `garden-founder-state.test.ts`,
  byte-matching the Go corpora
- `production.TestGarden*`
- `production.TestGardenHarvestClockIntegration` (six real-Postgres clock-disagreement arms)
- `production.TestGardenReadDatabaseClockIntegration` and `TestGardenReadDatabaseFailureIntegration`
- `gameserver.TestComposedGardenAuthenticatedCommandsAndReadIntegration` (fixture-only live HTTP/Account/DB path)
- `client/test/garden-surface-witnesses-browser.test.ts` (native input and controlled read ordering)
- `client/test/garden-refresh-browser.test.ts` (controlled due/visibility and actual browser-port binding)
- `client/test/garden-host-browser.test.ts` (actual host/runtime with controlled HTTP/socket boundaries)
- The Postgres witnesses `TestGardenIntegrationPersistsReplayableFounderLog`,
  `TestGardenHarvestIntegration` and `TestGardenHarvestFaultsAreAllOrNothing`
- `client/test/garden-surface-browser.test.ts`, in three browsers
