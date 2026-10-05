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
- The hidden salt is drawn by the server when the first unlocked advance runs, frozen in that row's
  resolved inputs, and never projected.

## Commands

- `garden_plant`, `garden_uproot`, `garden_set_substrate` and `garden_harvest` are Founder intents
  on `/api/v1/intents`. Their schemas are exact; a client time, tick, outcome or salt field is
  terminal `invalid`.
- **The advance pre-step** runs after the Fiscal sweep and before the body, for exactly the garden
  commands plus `spend_fiscal_credit` (a property test proves this set is sufficient). An applied
  trigger's receipt carries `garden_advance`, and `garden_advanced.v1` precedes the command event
  when visible.
- **Harvest** commits through the Founder→Company coordinator:
  - the Founder intent record is the exactly-once authority;
  - the faucet window is keyed `(founder, "server_garden", attended_day)`;
  - the credit to `company.cash` saturates;
  - both logs bind the same `harvest_hash`;
  - a zero harvest touches no window;
  - past the daily sends, units forfeit with `cap.minigame_faucet` and seeds are still recorded;
  - a rejected harvest is Founder-only.
- The six SG8 event kinds are in migrations 00082 and 00083.

## Read

`GET /api/v1/garden/current` (`get_current_garden`) returns `inactive`, `locked` or `active`. It
runs the advance on a discarded clone, and it never returns the salt, a draw or a future tick.

## Client surface

`client/src/game-ui/garden/GardenSurface.svelte` is the SG10 surface:
- The Garden tab appears only when the read is `locked` or `active`.
- The grid is a `role="grid"` of native buttons with a roving tabindex and arrow keys.
- Stage and dormancy are text plus a border shape, never colour.
- Growth announcements go to a polite live region once per refresh.
- It re-reads at `next_tick_wall_ms` while visible and after every receipt, with no client
  simulation.
- The grid fits 320 CSS px with at least 24 px targets.
- Commands go through the Game UI's Founder-scoped `act()`.

All garden copy is candidate text (`copy/catalog/garden-candidate.json`): species names read
`PENDING OWNER NAME`.

## Verification

- `make garden-corpus-check`
- `go test ./garden`
- `client/test/garden-engine.test.ts`, `garden-replay.test.ts` and `garden-founder-state.test.ts`,
  byte-matching the Go corpora
- `production.TestGarden*`
- The Postgres witnesses `TestGardenIntegrationPersistsReplayableFounderLog`,
  `TestGardenHarvestIntegration` and `TestGardenHarvestFaultsAreAllOrNothing`
- `client/test/garden-surface-browser.test.ts`, in three browsers
