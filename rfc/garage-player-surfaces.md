# RFC: Garage Player Surfaces (Fiscal, Achievements, Meters, Pet, Active Play, Shop, The Pitch)

- **Status:** draft — not implementation authority
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/00-vision.md` (pillars, anti-goals, no real money), `design/02 §1/§5/§6/§7`
  (golden opportunities, Fiscal Quarters, Clout, the moral meters "all visible and published"),
  `design/04-pets.md §1` + tone guard (the pet layer is sincere, never a joke), `design/08 §1`
  (voice rules, curtain always pulled), `design/11-ux-writing.md §1/§2/§5/§6/§7` (return sequence,
  progressive disclosure, copy inventory and production, one screen/tabs/depth ≤2, accessibility
  baseline, sincerity islands)
- **Depends on:** archived Game UI Screens (`docs/game-ui.md` — the host, U2 contract, and
  `runtime.ts` ownership), archived UI Foundation (`docs/ui-foundation.md`), archived Copy Pipeline,
  the implemented backends below, Minigame & Recovery API + Surface (accepted — this RFC proposes
  its MA-C9 surface enumeration; see GS7), API Foundation (accepted; its generated client is queue
  item 3g), and draft Accessibility of Player Workflows (its A1–A5 floor is consumed here as a
  proposed acceptance floor; it is not yet accepted)
- **Parent / amends:** follow-up to archived Game UI Screens (whose U3 names these surfaces as
  later screen sets on the same U2 contract); does not edit the archive. Proposes body text for
  `rfc/minigame-api-and-surface.md` MA3/MA-C9 for that RFC's ruling author to reconcile.
- **Supersedes / superseded by:** —
- **Planning:** `planning/garage-player-surfaces/` after acceptance
- **Source coordinate:** repository HEAD `c2d9bbc` (2026-09-25). Every code path cited below was read
  at this coordinate.

## Summary

Seven implemented v0.1 backends — Fiscal Quarters, Achievements, Meters, Pet care, active-play
opportunities/buffs, purchasable content (provisioning caps) and The Pitch — have no player-facing
surface, and one (the cosmetic shop) has no backend at all beyond a static stub. This RFC specifies
one coordinated set of Game UI surfaces over the backends **as they exist**: one additive read
projection (`game_ui_snapshot` v4), one runtime correction (intent receipts are returned instead of
discarded), closed event decoders for announcements, and one section per surface enumerating its
surface ID, exact data source, props, states, intents, error mapping, keyboard map,
focus/announcement behavior, reduced motion, 320 px reflow and copy keys. It invents no mechanic:
where a surface would need behavior the backend does not have, the section stops at a numbered
`DESIGN-GAP` with options. GS7 is the Pitch/`minigame_session` enumeration that execution-queue
item 3h requires before `rfc/minigame-api-and-surface.md` AC5 can be implemented.

## Motivation

`planning/roadmap-1.0.md` names v0.1 *The Garage* as "Fiscal/Clout/shop, three named minigames, pet
care … era UI", requiring "a tested default player journey for every named feature". The server
and replay work for most of that is archived; the player cannot see any of it. Archived Game UI
Screens U3 deferred exactly these surfaces to "a later screen-set RFC on the same U2 contract", and
U2 binds that successor: *"A screen needing a field the wire lacks is a DESIGN-GAP escalation, never
a client-side derivation of authoritative data."* The current wire lacks every field these surfaces
need (finding F1), so this RFC must specify the projection as well as the components.

**Scope boundary.** In: the surfaces listed above, the v4 projection that feeds them, runtime
receipt/error handling, event decoding for announcements, and the Pitch surface enumeration.
Out: the Phase-0 Playable Preview manifest (D-001/D-007 own it; nothing here widens the preview);
the soul-recovery surface (MA-C9's second half — see OD-14); pet identity, species, naming,
acquisition, battles and the house (no backend — DG-4); cosmetic items, crates, hats and any
purchasable cosmetic (no backend — DG-5); the return-sequence diorama/modal (see OD-3); Clout
minting (the Achievements foundation deliberately does not mint Clout); boards/feed/Commons panels
(their own RFCs); any balance retune (DG-1/DG-2 route to a balance mint, not to this RFC).

## Current-state findings (evidence at `c2d9bbc`)

These are the load-bearing facts the specification is built on. Each is a read of current code or
data, not an inference.

| ID | Finding | Evidence |
|---|---|---|
| F1 | The live snapshot is `game_ui_snapshot` v3 with exactly 13 top-level keys; it carries **no** Fiscal, achievement, meter, pet, opportunity/buff, minigame-availability or provisioning-cap data. The client parser rejects any extra key. | `server/gameui/projector.go` (`snapshot` struct, `projectSnapshot`); `client/src/game-ui/contracts.ts` (`parseGameUISnapshot` exact field list); `server/account/game_ui_api_schema.go` |
| F2 | `runtime.intent()` returns `Promise<void>` and discards the receipt; the server answers a *rejected* intent with HTTP 200 and `{intent_id,outcome:"rejected",current_revision,rejection:{category,detail}}`. `GameUIApp.act()` maps every thrown error to `offline = true`. Rejection reasons are therefore invisible to every surface. | `client/src/game-ui/runtime.ts` (`intent`); `server/production/intents.go` (`marshalRejection`); `server/account/api.go` (`submitIntent`); `client/src/game-ui/GameUIApp.svelte` (`act`) |
| F3 | Player-channel events of every kind are published, but `decodeGameUIEvent` returns `undefined` for any kind other than gate/offer/run-end, so `achievement_earned.v1`, `meter_band_changed.v1`, `pet_care_applied.v1`, `pet_status_changed.v1`, `fiscal_period_harvested.v1`, `fiscal_credit_spent.v1`, `opportunity_*`, `buff_*` and `minigame_resolved.v1` are received and dropped. | `client/src/game-ui/events.ts`; `server/transport/wire.go` (event envelope); `server/save/intent.go` (`AllEventKinds`, payload validators) |
| F4 | No pet exists or can be created. Epoch 6 activates the pet policy with an **empty** pet map and "no fabricated starter pet"; no intent, route or content creates a pet ID. `care_action` on any ID rejects `unknown_id/unknown_pet`. | `changelog/epoch-6.md`; `docs/pet-care.md`; `server/production/founder_replay.go` (`unknown_pet`) |
| F5 | No cosmetic backend exists. The only cosmetic is presentation row `cosmetic.horse_armor_free` with `purchasable:false, stateful:false`, rendered statically on the Desk. | `client/src/game-ui/presentation.generated.json`; `GameUIApp.svelte` |
| F6 | Fiscal Quarters and The Pitch are **live** since epoch 6 (current artifact list includes `fiscal`, `minigames`, `minigame_api`, `pitch`), contradicting `docs/fiscal-quarters.md` ("no production epoch currently contains the optional `fiscal` artifact") and `docs/minigame-the-pitch.md` ("no production epoch pins the content"). Docs drift; not fixed here. | `balance/epochs/phase0.json`; `changelog/epoch-6.md` |
| F7 | The live Fiscal clock policy is `early_ms:100, guaranteed_ms:200, auto_ms:300` (milliseconds) with `credit_per_period:3`, `hardcap:1000` — a period ripens ~3×/second and credit saturates in under two minutes of wall time. The archived Fiscal RFC annotates these fields "design: ~20 h / ~24 h". | `balance/fiscal/first-content.json`; `rfc/archive/fiscal-quarters-foundation.md` |
| F8 | The live opportunity schedule is seconds-scale (`minimum_interval_ms:1000, scale_ms:500, lifetime_ms:4500`; buff durations 1–5 s) versus design/02 §1's 300–900 s spawns and ~77 s frenzies. It carries an owner ratification (`33418ba`), so it may be deliberate for the T0 curriculum. | `balance/opportunities/t0-t1.json`; `changelog/epoch-7.md` |
| F9 | The Game UI `wind_down.eligible` preview does not consult the active-minigame predicate, but Exit rejects `not_eligible/minigame_session_active` while a Pitch session is active (MA-C12). Combined with F2, the Wind Down button can be enabled and silently do nothing. | `server/gameui/projector.go` (`previewPhaseATransitions` call); `server/production/replay.go` (`minigame_session_active`) |
| F10 | Reason/copy keys these surfaces must render are absent from the generated client copy catalog: `cap.fiscal_credit`, `cap.fiscal_level.beige_tower`, `cap.active_combo`, `cap.cash`. `GameUIApp.capFor` throws on a missing cap key. Pitch hack copy is title-only (e.g. `pitch.hack.dark_pattern` = "Dark Pattern"): no effect text and no curtain. No copy exists for meters, pets or Fiscal. | `client/src/copy/generated/catalog.json`; `balance/fiscal/first-content.json`; `balance/opportunities/t0-t1.json`; `balance/pitch.json` |
| F11 | `act()` binds `expected_revision` to the **Company** revision. `care_action`, `harvest_fiscal_period` and `spend_fiscal_credit` are Founder commands whose `expected_revision` is the **Founder** revision (`ApplyFounderLogged`). | `GameUIApp.svelte` (`act`); `server/production/intents.go` (`handleFounderCare`, `handleFounderFiscal`) |
| F12 | Fiscal unlock row `unlock.arcade` (cost 5) has no consumer anywhere in server or client code; spending on it succeeds and changes nothing. | `balance/fiscal/first-content.json`; repository grep |
| F13 | While a Soul-recovery session is active, every intent on `/api/v1/intents` except `buy_route_hint`, `harvest_fiscal_period` and `spend_fiscal_credit` (so including `care_action` and `claim_opportunity`) is answered with the rejection `not_eligible/exclusive_activity` (optionally `session_expired:true`). | `server/production/intents.go` (`Handle`, `marshalExclusiveActivityRejection`) |

## Specification — proposed, pending owner acceptance

### GS0 — Shared contract (binding on every surface below)

#### GS0.1 Projection: `game_ui_snapshot` v4

The server projection gains schema version 4, served by the existing operation
`get_game_ui_snapshot` (`GET /api/v1/founder/state`) and the bootstrap receipt, registered in the
same private registry (`server/account/game_ui_api_schema.go`) and generated into
`client/src/api/generated/types.ts`. v4 = the exact v3 envelope plus **one** new key, `features`,
an object with exactly six keys, each either `null` or the exact arm defined in its surface section:

```text
features: {
  active_play:  null | ActivePlayArm     (GS5)
  achievements: null | AchievementsArm   (GS2)
  fiscal:       null | FiscalArm         (GS1)
  meters:       null | MetersArm         (GS3)
  minigames:    null | MinigamesArm      (GS7)
  pets:         null | PetsArm           (GS4)
}
```

and one extension to the existing `generators[]` row: `provision_cap: null | {amount, reason_key}`
(GS6).

Rules:

1. An arm is `null` exactly when the pinned bundle lacks the artifact **or** the save's feature
   schema is below the version that activates it (Meters/Achievements: Company v16; active play:
   Company v18; pets: Founder v18; Fiscal: Founder v19; minigames: Founder v21 with the
   `minigame_api` artifact). The projector never consults deploy-current catalogs
   (the existing `ResolveReplayCatalogs(constantsHash)` rule).
2. Every value is either a persisted field or a **read-only derivation computed by the existing
   kernel function on a discarded clone** — the same precedent as v3's Gate/Wind Down preview
   (`previewPhaseATransitions`). No new arithmetic path is written for the projection. If a needed
   derivation has no exported pure function (e.g. pet decay-to-now separated from a care command),
   the implementer files a blocker; it does not re-implement the math.
3. All arrays are byte-sorted by their mechanical ID; all integers are exact safe integers; all
   big numbers are canonical Decimal strings.
4. v1–v3 remain decodable for stored bootstrap receipts (existing rule). **Live sync requires v4**
   once this RFC is implemented, exactly as v3 is required today (`runtime.ts loadSnapshot`).
5. Facts: the projection's `facts` array adds exactly `feature.active_play`, `feature.achievements`,
   `feature.fiscal`, `feature.meters`, `feature.minigame.pitch` and `feature.pets` (boolean;
   `true` iff the corresponding arm is non-null, and for pets iff at least one pet exists, and for
   the Pitch iff the minigames arm contains a `pitch` row). These are the surfaces' unlock facts
   (GS0.4); `GAME_UI_FACT_IDS` in `client/src/game-ui/surface-catalog.ts` gains them.

The alternative of one new read operation per system is OD-1.

#### GS0.2 Runtime: receipts, revisions and error mapping

`client/src/game-ui/runtime.ts` remains the only network owner. Changes:

1. `intent(body)` returns `Promise<IntentOutcome>`:

   ```ts
   type IntentOutcome =
     | Readonly<{ outcome: "applied"; receipt: Readonly<Record<string, unknown>> }>
     | Readonly<{ outcome: "rejected"; category: string; detail: string; currentRevision: number;
                  sessionExpired: boolean }>;
   ```

   A 200 body whose `outcome` is neither value, or whose rejection is not exactly
   `{category,detail}` (+ optional `session_expired`), throws `SyntaxError` (fail closed).
2. Non-2xx responses throw `GameUIRequestError { status, category, detail }` parsed from the
   API error body; only a network failure or unparsable body throws a transport error.
3. `GameUIApp.act()` accepts a `scope: "company" | "founder"` and sends
   `expected_revision = snapshot.revision` or `founderRevision` respectively (fixes F11). Founder
   intents are disabled while `founderRevision` is undefined.
4. Surfaces receive outcomes through a shared mapper; no surface parses HTTP.

Shared error mapping (every surface uses this table, then its own rejection rows):

| Result | Surface behavior | Copy key |
|---|---|---|
| 200 `applied` | Clear pending; await authoritative refresh (existing receipt → `refresh()` path); announce the surface's success line if it has one. | per surface |
| 200 `rejected`, surface-specific pair | Clear pending; show the pair's text in the surface status region; keep focus on the triggering control if it is still enabled, else on the surface heading. | per surface |
| 200 `rejected` `not_eligible/exclusive_activity` | As above; controls stay disabled until the next snapshot. | `intent.rejection.exclusive_activity` |
| 200 `rejected`, unlisted pair | Show generic line; report one client invariant; never render the mechanical pair. | `intent.rejection.unknown` |
| 409 `conflict/intent` (stale revision) | Trigger one authoritative `refresh()`; show the "books changed, try again" line; do **not** auto-retry (a new intent needs a new `intent_id` and the player's fresh consent). | `intent.conflict` |
| 400 `invalid/*` | Client defect: report invariant, show generic line. | `intent.rejection.unknown` |
| 401/404 on the intent route | Existing offline/credential path. | existing `settings.save_status.offline` |
| 429 `rate_limited/*` | Show rate line; re-enable after the next snapshot. | `intent.rate_limited` |
| 5xx / transport failure | Existing `offline = true` path. | existing |

#### GS0.3 Events: closed announcement decoders

`client/src/game-ui/events.ts` gains exact, fail-closed decoders (same `exact()` discipline, same
field domains as `server/save/intent.go validateEventPayload`) for:
`achievement_earned.v1`, `meter_band_changed.v1`, `pet_status_changed.v1`,
`fiscal_period_harvested.v1` (both `automatic` and `manual` arms) and
`buff_started.v1` (schema v1 and v2). Decoded events drive **announcements only** (GS0.6); every
displayed value still comes from the next snapshot. A decoder failure takes the existing
`authoritativeResync()` path (as today for malformed lifecycle events). Other kinds remain ignored.

#### GS0.4 Surface registry and host

`client/src/game-ui/surfaces.json` gains these rows (byte-sorted with the existing five):

| surface_id | mount_id | unlock |
|---|---|---|
| `achievements` | `game.surface.achievements` | `fact_equals feature.achievements true` |
| `fiscal` | `game.surface.fiscal` | `fact_equals feature.fiscal true` |
| `meters` | `game.surface.meters` | `fact_equals feature.meters true` |
| `minigame_session` | `game.surface.minigame_session` | `fact_equals feature.minigame.pitch true` |
| `pet` | `game.surface.pet` | `fact_equals feature.pets true` |

Active play and purchasable content are **Desk regions**, not surfaces (GS5, GS6). `GameUISurfaceID`
grows by the five IDs. Each new surface is its own Svelte component with the props declared in its
section; `GameUIApp.svelte` only hosts, navigates and owns `act()`. The persistent nav renders one
native `<button>` per unlocked surface in registry order after Desk, before Settings, with
`aria-current="page"` on the active one. Lifecycle preemption (Offer/Run-End) keeps its existing
precedence over every new surface (`navigation.ts`); none of the new surfaces is a destructive
confirmation, so none defers preemption.

#### GS0.5 Shared states

Every surface implements this closed state set, derived from host state it does not own:

| State | Condition | Rendering |
|---|---|---|
| `loading` | No snapshot bound yet, or the surface's first authoritative read is in flight. | Heading + `common.loading` in `role="status"`; controls absent. |
| `empty` | The arm is non-null but has no rows to act on (defined per surface). | Heading + the surface's empty key. Never a fabricated placeholder row. |
| `active` | Arm non-null, transport ready or HTTP-synced. | Full surface. |
| `reconnect` | `offline || resyncing || !transportReady` in the host. | Last authoritative values stay visible and marked stale (`common.stale_note`); all intent controls disabled with the reason visible; no local prediction of Founder/Fiscal/pet values. |
| `error` | The arm fails to decode, or a required presentation/copy row is missing. | Surface-level `role="alert"` with `common.surface_error`; controls absent; one invariant reported. A missing arm never crashes the host (other surfaces keep working). |

The arm going `null` while mounted (e.g. new run on an older bundle) unmounts the surface and the
host returns to Desk with focus on the Desk heading (A1.1 rule for a context change the player did
not choose).

#### GS0.6 Shared accessibility contract

Proposed as the acceptance floor, consuming draft `rfc/accessibility-player-workflows.md` A1–A5:

- **Heading/focus:** one `<h1 tabindex="-1">` per surface; a player-selected tab change keeps focus
  on the nav button; a forced change focuses the new heading. Intent completion never moves focus
  except when the triggering control is removed (then: the nearest surviving control in the same
  region, else the heading).
- **Announcements:** each surface owns exactly one polite `role="status"` region for its own intent
  outcomes and one `role="alert"` for its `error` state. Cross-surface announcements (achievement
  earned) go to one polite chrome region owned by the host, deduplicated by event `cursor`; replay
  of the same cursor after reconnect announces nothing. Countdowns are never inside a live region.
- **Keyboard:** native controls only (`button`, `input[type=checkbox]`, `details/summary`); logical
  DOM order; no global hotkeys, no `accesskey`, no key interception, no Escape handler, no timed key
  sequences. Enter/Space per native semantics.
- **Non-color:** every state (eligible, cooldown, locked, owned, capped, earned, band) has text.
  `<meter>`/progress visuals always carry the numeric text beside them.
- **Reduced motion:** no new animation is introduced by any surface; transitions use theme
  duration tokens only, which are zero in `era_1995` and under reduced motion (UI Foundation). The
  live `MediaQueryList` listener required by A4 is inherited from the host once that RFC lands;
  surfaces may not add their own.
- **Reflow:** every grid uses `repeat(auto-fit, minmax(min(<n>rem, 100%), 1fr))`; no fixed widths;
  at 320 CSS px (and 400% zoom on 1280 px) the full mounted surface **including persistent chrome
  and the grown nav** has no horizontal page scroll and no clipped control. Tables (Meters) collapse
  to one labeled row per value below 30rem.
- **Target size:** every control ≥ 24×24 CSS px on coarse pointers.

#### GS0.7 Copy and presentation

- Every player-visible string is a Copy Pipeline key; zero literals (enforced by the existing
  `make verify-client-boundary` scan). Keys listed per surface are **key reservations only**; all
  prose is owner-authored or Claude-drafted for Marco's editorial cut (design/11 §6) and passes the
  copy linter, the statistic detector and provenance rules. Params use only `string`, safe
  `integer` and `canonical_decimal`.
- Tones: Fiscal/Meters/Active/Pitch use the existing `corporate`/`diegetic`; achievements use
  `achievement`; **pet copy uses `diegetic` and must be sincere** (design/04 tone guard; the
  Soul-recovery rows are the precedent). A dedicated sincere tone is OD-11.
- Mechanical-ID → copy-key mappings not owned by an artifact's own `copy_key` field live in the
  presentation catalog, which bumps to schema v4 with new row families declared per section.
  Unknown IDs are **withheld** and reported, never rendered mechanically (the existing
  network-slot/payout precedent in `docs/game-ui.md`).
- Every parodied dark pattern on these surfaces pulls its curtain in visible text or tooltip
  (design law 10): early Fiscal harvest risk, the `dark_pattern` Pitch hack, Lucky's bank formula.

#### GS0.8 Pending discipline

The existing `act()` single-flight queue governs all new intents: at most one in-flight intent in
the whole Game UI; a click of the same kind while pending is dropped; a different kind waits. Every
control whose intent is pending shows `common.pending` text and `aria-disabled` state while
remaining focusable (so focus is not lost).

---

### GS1 — Fiscal Quarters surface (`fiscal`)

**Backend:** `server/fiscal/transition.go`, `server/production/founder_replay.go`
(`applyFounderFiscalHarvestResolved`, `applyFounderFiscalSpendResolved`),
`server/production/intents.go` (`handleFounderFiscal`); client kernel `client/src/fiscal.ts`.
**Design:** design/02 §5 (Earnings Calls → Investor Confidence; the real-time clock nothing can buy).

**Surface:** `fiscal` / `game.surface.fiscal` / unlock `feature.fiscal`.

**Data source — `FiscalArm` (all from Founder state + pinned `fiscal` artifact):**

```text
{
  credit: int,                               // FiscalCredit (persisted)
  credit_cap: {amount: int, reason_key: id}, // credit_policy.hardcap / hardcap_reason_key
  credit_per_period: int,
  period: {opened_wall_ms: int, seq: int,    // FiscalPeriodOpenedWallMS / FiscalPeriodSequence
           early_ms: int, guaranteed_ms: int, auto_ms: int, early_success_ppm: int},
  sweep_preview: {periods: int, credited: int, credit_after: int, saturated: bool},
                                             // fiscal auto-sweep at server_now_ms on a clone
  hoard: {preview_ppm: int, cap_credits: int, reason_note: "next_run"},
                                             // hoard contribution the NEXT run would freeze
  generator_levels: [{generator_id, level: int, level_cap: {amount: int, reason_key},
                      ppm_per_level: int, next_level_cost: null | int}],
  unlocks: [{unlock_id, cost: int, owned: bool}]
}
```

`next_level_cost` is the server's `GeneratorLevelCost(id, level, 1)`; `null` at the level cap.
`hoard.preview_ppm` is the value the run-genesis frozen-contribution path would compute from the
post-sweep credit (read-only call of the same function). Unlock rows without a presentation row
are withheld (F12, DG-3).

**Props:**

```ts
type FiscalSurfaceProps = Readonly<{
  arm: FiscalArm; serverNowMs: () => number;           // host's estimatedServerNowMS
  era: CopyEra; pending: boolean; controlsEnabled: boolean;
  onHarvest(): void;
  onSpendLevel(generatorId: string): void;             // always levels: 1 (OD-4)
  onSpendUnlock(unlockId: string): void;
  lastOutcome: IntentOutcome | undefined;
}>;
```

**States:** GS0.5; `empty` never occurs (the credit panel always exists). Harvest phase is derived
display-only from `serverNowMs() - period.opened_wall_ms` against the three pinned windows:
`ripening` (disabled, shows remaining time), `early` (enabled, risk stated), `guaranteed`
(enabled). The receipt decides; a stale estimate that sends early produces an honest rejection.

**Intents** (Founder scope, `expected_revision = founderRevision`):
`{kind:"harvest_fiscal_period"}`,
`{kind:"spend_fiscal_credit", target:{kind:"generator_level", generator_id, levels:1}}`,
`{kind:"spend_fiscal_credit", target:{kind:"unlock", unlock_id}}`.

**Error mapping** (in addition to GS0.2):

| Rejection | Copy key |
|---|---|
| `not_eligible/period_not_ripe` | `fiscal.rejection.period_not_ripe` |
| `unaffordable/fiscal_credit` | `fiscal.rejection.unaffordable` |
| `not_eligible/already_unlocked` | `fiscal.rejection.already_unlocked` |
| `cap_exceeded/<generator_id>` | the row's `level_cap.reason_key` |
| `unknown_id/<target id>` | `intent.rejection.unknown` + invariant |

Applied harvest receipts carry `harvest_outcome ∈ {early_succeeded, early_failed, guaranteed,
consumed_by_auto}`; each maps to `fiscal.outcome.<outcome>` in the status region. `early_failed`
text states plainly that the period was consumed and nothing was credited.

**Keyboard map:** Tab order: harvest button → per-level rows (one "buy +1 level" button each) →
unlock rows (one button each). No other bindings.

**Focus/announcement:** outcome line in the surface status region. A ripening → ripe transition is
**not** announced (it would be a timer in a live region); the nav button text gains the
`fiscal.nav_ripe_badge` suffix instead (OD-3).

**Reduced motion:** countdown text updates at most 1 Hz; no progress animation.

**320 px:** credit panel, harvest panel, levels list and unlocks list stack vertically; each row is
a two-line card.

**Copy keys:** `surface.fiscal.title`, `fiscal.credit_label`, `fiscal.credit_frame {credit:integer,
cap:integer}`, `fiscal.period.ripening_frame {remaining:string}`, `fiscal.period.early_frame
{success_percent:integer}`, `fiscal.period.guaranteed`, `fiscal.period.auto_note {remaining:string}`,
`fiscal.harvest`, `fiscal.harvest_tooltip` (curtain: early-harvest risk and that wall time cannot be
bought), `fiscal.sweep_preview_frame {periods:integer, credited:integer}`, `fiscal.next_run_note`
(spends and hoard affect the **next** run only — docs/fiscal-quarters.md), `fiscal.hoard_frame
{percent:string, cap:integer}`, `fiscal.levels_label`, `fiscal.level_frame {level:integer,
cap:integer}`, `fiscal.level_buy {cost:integer}`, `fiscal.unlocks_label`, `fiscal.unlock_buy
{cost:integer}`, `fiscal.unlock_owned`, `fiscal.outcome.early_succeeded`, `fiscal.outcome.early_failed`,
`fiscal.outcome.guaranteed`, `fiscal.outcome.consumed_by_auto`, `fiscal.rejection.period_not_ripe`,
`fiscal.rejection.unaffordable`, `fiscal.rejection.already_unlocked`, `fiscal.nav_ripe_badge`;
catalog reason keys `cap.fiscal_credit`, `cap.fiscal_level.beige_tower` (F10). Presentation v4 row
family `fiscal_unlocks: [{id, title_key, description_key}]` (initially `minigame.pitch` only).

**Acceptance (each with a failing case):**

- **GS1-A1 projection:** Go and TS tests decode a v4 fixture with a non-null Fiscal arm and a
  real-Postgres projection test asserts `sweep_preview` equals the credit the next harvest receipt
  reports. *Fails if* the projector reads persisted credit without the sweep (seeded: pass the
  unswept state) or the TS decoder accepts an extra key.
- **GS1-A2 revision scope:** browser test with a runtime double asserts harvest sends the Founder
  revision. *Fails if* `scope` is dropped (the double returns 409 and the test asserts the applied
  path).
- **GS1-A3 phases:** fixture tests at `early_ms − 1`, `early_ms`, `guaranteed_ms` render disabled /
  risk-stated / guaranteed with the correct text. *Fails if* the phase is color-only (text assertion)
  or off-by-one.
- **GS1-A4 rejection visibility:** each rejection row renders its key in `role="status"`. *Fails if*
  the F2 behavior is restored (receipt discarded → no status text).
- **GS1-A5 axe/keyboard/reflow:** zero serious/critical axe findings in all states × three engines;
  keyboard-only harvest + spend; no horizontal scroll at 320 px. *Fails on* a seeded
  unlabeled button and a seeded `min-width: 40rem`.
- **GS1-A6 composed:** real gameserver + Postgres, DOM-only: harvest → spend `minigame.pitch` →
  `feature.minigame.pitch` becomes true and the Pitch nav button appears. *Blocked by* OD-5 if the
  pinned clock is retuned to hours (no fixture clock is permitted).

---

### GS2 — Achievements surface (`achievements`)

**Backend:** `server/achievements/evaluate.go`, `server/achievements/catalog.go`; save fields
`AchievementsEarnedRun`, `AchievementScoreRun`, `AchievementsEarnedLifetime`,
`AchievementScoreLifetime` (`server/save/state.go`); event `achievement_earned.v1`.
**Design:** design/02 §6, design/11 §5 (name + flavor line), design/08 §1.

**Surface:** `achievements` / `game.surface.achievements` / unlock `feature.achievements`.

**Data source — `AchievementsArm`:**

```text
{
  score: {run: int, lifetime: int},          // Company run score / Founder lifetime score
  rows: [{achievement_id, copy_key, condition_scope: "run"|"career",
          proof_kind: "provenance"|"burn"|"possession", score_grant: int,
          earned: "run"|"lifetime"|null}]
}
```

`copy_key` and `proof_kind` come from the pinned achievements artifact row; `earned` is `run` if in
the Company run set, `lifetime` if in the Founder lifetime set, else `null` (the kernels reject
overlap, so exactly one applies).

**Props:** `{ arm: AchievementsArm; era: CopyEra }` — read-only; no intents, no callbacks.

**States:** GS0.5; `empty` = zero rows in the pinned catalog (`achievements.empty`).

**Intents:** none. Achievements are earned by the server during ordinary commands only.

**Error mapping:** none beyond GS0.5 (`error` if a row's `copy_key` is missing from the catalog).

**Keyboard map:** the list is static content; only nav/tab stops. If OD-7 selects a
run/career filter, it is a native radio group.

**Focus/announcement:** host chrome announces `achievements.earned_announcement {achievement:string}`
once per `achievement_earned.v1` cursor, **on any surface** (design/11 §2: one contextual line; not a
modal). Rows with `proof_kind: "possession"` show the existing `achievement.possession_warning`.

**Reduced motion:** no unlock animation in any mode.

**320 px:** single-column list; score header wraps.

**Copy keys:** `surface.achievements.title`, `achievements.score_frame {run:integer,
lifetime:integer}`, `achievements.progress_frame {earned:integer, total:integer}`,
`achievements.scope.run`, `achievements.scope.career`, `achievements.state.earned_run`,
`achievements.state.earned_lifetime`, `achievements.state.locked`, `achievements.grant_frame
{score:integer}`, `achievements.earned_announcement {achievement:string}`, `achievements.empty`;
per-row text is the artifact's existing `copy_key` (12 rows present in the catalog today);
existing `achievement.possession_warning`.

**Acceptance:**

- **GS2-A1:** v4 fixture with run-earned, lifetime-earned and locked rows renders three distinct
  text states. *Fails if* state is conveyed only by color/opacity (text assertion on each row).
- **GS2-A2 announcement once:** replaying the same `achievement_earned.v1` envelope after a
  simulated reconnect produces exactly one announcement. *Fails with* a seeded decoder that ignores
  the cursor.
- **GS2-A3 score is not Clout:** a source test asserts no achievements copy key or component binds
  `CloutLifetime` or labels score as Clout until OD-6 rules otherwise. *Fails on* a seeded binding.
- **GS2-A4 composed:** DOM-only first generator purchase on the real server earns
  `achievement.generators_purchased_1`; the chrome announcement fires and the row reads earned after
  refresh. *Fails if* the event decoder is severed (no announcement) or the arm is severed (row stays
  locked).
- **GS2-A5:** axe/keyboard/320 px as GS1-A5.

---

### GS3 — Meters surface (`meters`)

**Backend:** `server/meters/transition.go`, `server/meters/catalog.go`; Company `MeterValues`,
`MeterBands`; event `meter_band_changed.v1`. **Design:** design/02 §7 — "All meters visible and
published … Cannot be bought"; p(doom) is a pressure meter, not a moral quantity.

**Surface:** `meters` / `game.surface.meters` / unlock `feature.meters`.

**Data source — `MetersArm`:**

```text
{
  meters: [{meter_id, value: int, min: int, max: int, band_id,
            bands: [{band_id, floor_value: int}]}]
}
```

Values are the committed Company values as of `evaluated_through_ms` (Meters advance only on
commands; offline time gives zero elapsed — docs/meters.md). The projection does **not**
extrapolate decay; the surface labels values "as of last update" (`meters.as_of_note`).

**Props:** `{ arm: MetersArm; era: CopyEra }` — read-only.

**Layout:** a 5 × 2 table (constituencies users/employees/regulators/press/investors ×
Standing/Grievance) plus a separate p(doom) row (OD-8 governs its visibility). Each cell: native
`<meter min max value>` + numeric text + band text.

**States:** GS0.5; `empty` impossible (the artifact requires all eleven IDs).

**Intents:** none — meters are not spendable (no intent may exist; the loader already enforces
disjointness from economy resources).

**Keyboard map:** none beyond tab stops; the table uses `<th scope>` headers.

**Focus/announcement:** while the Meters surface is mounted, `meter_band_changed.v1` announces
`meters.band_changed_announcement {meter:string, band:string}` politely; on other surfaces the nav
button gains `meters.nav_changed_badge` text until visited (no global announcement — design/11 §1
"everything else is a badge").

**Reduced motion:** no bar animation.

**320 px:** below 30rem the table becomes one labeled row per value (`<dl>`), preserving header
text for each value.

**Copy keys:** `surface.meters.title`, `meters.constituency.users|employees|regulators|press|investors`,
`meters.axis.standing`, `meters.axis.grievance`, `meters.doom.title`, `meters.doom.tooltip`
(pressure meter, not a moral score), `meters.band.low`, `meters.band.high` (presentation v4 row
family `meter_bands: [{band_id, title_key}]` so future bands fail closed), `meters.value_frame
{value:integer, max:integer}`, `meters.as_of_note`, `meters.curtain` (visible, published, never for
sale), `meters.band_changed_announcement {meter:string, band:string}`, `meters.nav_changed_badge`.
Presentation v4 row family `meters: [{id, title_key}]` for the eleven IDs.

**Acceptance:**

- **GS3-A1:** decoder rejects a meter value outside `[min,max]`, a missing ID or an unsorted array.
  *Fails with* a seeded lenient decoder.
- **GS3-A2 non-color:** every cell has numeric and band text. *Fails on* a seeded band-by-color
  variant.
- **GS3-A3 reflow:** 320 px collapses to the labeled list with every header preserved. *Fails on* a
  seeded fixed-width table.
- **GS3-A4 composed:** the real server's initial values (Standing 50, Grievance 0, p(doom) 50 from
  the pinned artifact) render as text after DOM bootstrap. *Fails if* the arm is severed.
- **GS3-A5:** axe/keyboard as GS1-A5.

---

### GS4 — Pet care surface (`pet`) — sincere surface

**Backend:** `server/pet/transition.go` (`ApplyCareTransition`, `CareStatus`), `server/pet/state.go`
(`CareState`), `server/production/intents.go` (`handleFounderCare`), `server/production/founder_replay.go`
(care + Soul gate); client kernel `client/src/pet/`. **Design:** design/04 §1 and tone guard.

**Mounting today:** because no pet can exist (F4), `feature.pets` is always `false` and the
surface never mounts in production. This section specifies the complete contract so it can be
implemented and tested against fixtures now and mounted the moment a pet producer ships (DG-4).
No placeholder pet, no "adopt" button and no empty-state pitch is shown to players.

**Surface:** `pet` / `game.surface.pet` / unlock `feature.pets`.

**Data source — `PetsArm`:**

```text
{
  attended_now_ms: int,                      // Founder attended time at server_now_ms (existing
                                             //   ResolveFounderAttendance sample, read-only)
  human_content_locked: bool,                // soul.Catalog.HumanContentLocked(Founder.Soul)
  pets: [{pet_id,
          stats: [{stat_id, value_ppm: int, floor_ppm: int}],   // decayed to attended_now_ms
          status_band, mood, behavior_state,                   // CareStatus on the decayed clone
          actions: [{action_id, stat_id, cooldown_until_attended_ms: int,
                     soul_gate: "essential"|"recovery"|"ordinary"}]}]
}
```

Decay-to-now reuses the kernel's decay step on a discarded clone (GS0.1 rule 2). Trust is **not**
projected until OD-10 rules.

**Props:**

```ts
type PetSurfaceProps = Readonly<{
  arm: PetsArm; era: CopyEra; pending: boolean; controlsEnabled: boolean;
  onCare(petId: string, actionId: string): void;
  lastOutcome: IntentOutcome | undefined;
}>;
```

**States:** GS0.5; `empty` cannot mount (unlock fact requires ≥1 pet).
Per action: `available`, `cooldown` (`cooldown_until_attended_ms > attended_now_ms`; remaining shown
as attended time with `pet.action.attended_note` — cooldowns pause while away), `soul_locked`
(`soul_gate == "ordinary" && human_content_locked`). Eligibility and saturation are decided by the
server; the UI does not pre-evaluate `min_eligible_ppm` or diminishing returns.

**Intents** (Founder scope): `{kind:"care_action", pet_id, action_id}`.

**Error mapping:**

| Rejection | Copy key |
|---|---|
| `not_eligible/cooldown` | `pet.rejection.cooldown` |
| `not_eligible/ineligible` | `pet.rejection.ineligible` |
| `not_eligible/saturated` | `pet.rejection.saturated` |
| `not_eligible/human_content_locked` | `pet.rejection.soul_locked` |
| `unknown_id/unknown_pet`, `unknown_id/unknown_action` | `intent.rejection.unknown` + invariant |

Applied receipt: `pet.care_applied {action:string}` in the status region; mood/band text changes
come from the refreshed snapshot.

**Keyboard map:** one button per action in catalog order; no other bindings.

**Focus/announcement:** outcome in the surface status region. `pet_status_changed.v1` announces
`pet.status_changed_announcement {band:string}` only while the surface is mounted.

**Reduced motion:** the design's visible-behavior cues (walk speed, ears, desaturation) are **not**
rendered in this RFC (no renderer exists; A4.3 forbids inventing one). `behavior_state` is text.

**320 px:** single column; stat rows are label + `<meter>` + percent text.

**Tone:** all pet copy is sincere (`diegetic`), never ironic, never a dark-pattern parody. The
Soul lock text states the fact gently; it does not shame. Visual treatment may only use existing
tokens (OD-11).

**Copy keys:** `surface.pet.title`, `pet.stat.hunger|energy|cleanliness|affection`,
`pet.stat_frame {percent:integer}`, `pet.band.floor|low|normal|high`,
`pet.mood.withdrawn|restless|neutral|engaged`, `pet.behavior.idle|care_response|active|resting`,
`pet.action.cooldown_frame {remaining:string}`, `pet.action.attended_note`, `pet.soul_locked_note`,
`pet.care_applied {action:string}`, `pet.status_changed_announcement {band:string}`,
`pet.rejection.cooldown|ineligible|saturated|soul_locked`. Presentation v4 row family
`pet_actions: [{id, title_key, description_key}]` for `care.feed`, `care.groom`, `care.pet`,
`care.play`, `care.rest`. Pet name/species keys are **not** reserved (DG-4).

**Acceptance:**

- **GS4-A1 fixture contract:** browser tests over v4 fixtures cover available / cooldown /
  soul-locked states and each rejection row. *Fails if* soul-locked is conveyed by greying only.
- **GS4-A2 never mounts without a pet:** with the real epoch-8 bundle and a fresh Founder, the pet
  nav button is absent. *Fails if* a seeded projector emits `feature.pets: true` for an empty map.
- **GS4-A3 Founder revision:** as GS1-A2.
- **GS4-A4 tone gate:** copy-lint row (new) rejects the `satire`/banned-winking list in `pet.*`
  keys. *Fails on* a seeded ironic row.
- **GS4-A5 composed:** **explicit blocker** until DG-4 ships a pet producer; recorded as blocked, not
  skipped or vacuously passed.

---

### GS5 — Active play (Desk region `desk.region.opportunity`)

**Backend:** `server/production/active_play.go` (schedule, `applyClaimOpportunity`,
`activePlayContributions`, combo clamp), Company `PendingOpportunity`, `ActiveBuffs`
(`server/save/state.go`); client `client/src/active-play.ts`. **Design:** design/02 §1 golden
opportunities; design/10 active playstyle.

**Placement:** a fixed region inside the Desk, immediately after the manual-action section. The
region element is **always present** (empty when nothing is pending) so DOM order never shifts.
Not a surface (OD-9): an opportunity on a separate tab would be structurally unclaimable.

**Data source — `ActivePlayArm`:**

```text
{
  attended_now_ms: int,                      // ResolveRateProjectionAttendedMS (already computed
                                             //   by the v3 projector for rates)
  pending: null | {opportunity_id, effect_row_id, selected_generator_id: null|id,
                   expires_attended_ms: int},
  buffs: [{buff_instance_id, effect_row_id, selected_target: null|id,
           expires_attended_ms: int}],
  combo: {cap: decimal, reason_key: id, saturated: bool}   // combo_policy + clamp result
}
```

`pending` is emitted only when `expires_attended_ms > attended_now_ms`; expired buffs are omitted.
Opportunities spawn **only inside Company commands** (lazy scheduler); the UI never sends a
synthetic command, polls, or otherwise tries to cause a spawn.

**Props:**

```ts
type OpportunityRegionProps = Readonly<{
  arm: ActivePlayArm; era: CopyEra; pending: boolean; controlsEnabled: boolean;
  onClaim(opportunityId: string): void;
  lastOutcome: IntentOutcome | undefined;
}>;
```

**States:** GS0.5 within the region; `empty` = no pending and no buffs (region renders nothing
visible but keeps its landmark label).

**Remaining time:** shown as attended seconds **as of the last snapshot**
(`expires_attended_ms − attended_now_ms`), not extrapolated, with `desk.opportunity.attended_note`
(the window pauses while you are away). Rationale: attended time advances only under command
cadence, so a wall-clock countdown would be false (OD-12).

**Intents** (Company scope): `{kind:"claim_opportunity", opportunity_id}`.

**Error mapping:**

| Rejection | Copy key |
|---|---|
| `not_eligible/opportunity_expired` | `desk.opportunity.rejection.expired` |
| `not_eligible/opportunity_not_pending` | `desk.opportunity.rejection.not_pending` |
| `unknown_id/opportunity_id` | `desk.opportunity.rejection.not_pending` + invariant |

Applied Lucky receipt with `saturated:true` renders `desk.opportunity.lucky_capped` plus the
receipt's `cap_reason_key` text (`cap.cash`); buff receipts with a non-null `cap_reason_key` render
`cap.active_combo` (hardcap visible — design law 5).

**Keyboard map:** one "claim" button in the region; Tab reaches it in DOM order after the manual
action. Focus is **never** moved to it on spawn (no unexpected context change).

**Focus/announcement:** a spawn is announced once politely (`desk.opportunity.spawned_announcement
{effect:string}`), keyed by `opportunity_id`, only while the Desk is mounted. Claim outcome goes to
the Desk status region. The time limit is a server rule accessibility may not extend (A2.2); the
announcement and a reachable control are the accommodation (OD-12).

**Reduced motion:** no pulse/float/particle; appearance is instant in every mode (the design's
golden-cookie animation has no renderer and is not invented).

**320 px:** region is one card; buff list stacks.

**Copy keys:** `desk.opportunity.region_label`, `desk.opportunity.claim`,
`desk.opportunity.remaining_frame {seconds:integer}`, `desk.opportunity.attended_note`,
`desk.opportunity.spawned_announcement {effect:string}`, `desk.opportunity.rejection.expired`,
`desk.opportunity.rejection.not_pending`, `desk.opportunity.lucky_frame {amount:canonical_decimal}`,
`desk.opportunity.lucky_capped`, `desk.opportunity.lucky_tooltip` (curtain: the published bank
formula `min(fraction × bank, cap × rate) + ε`), `desk.buffs_label`, `desk.buff.remaining_frame
{seconds:integer}`, `desk.buff.combo_capped`; catalog reason keys `cap.active_combo`, `cap.cash`
(F10). Presentation v4 row family `opportunity_effects: [{id, title_key, description_key}]` for
`active.building`, `active.click`, `active.lucky`, `active.production`; an unknown effect row
withholds the opportunity and reports an invariant.

**Acceptance:**

- **GS5-A1 stable DOM order:** inserting a pending opportunity does not change the index of any
  existing Desk control and does not move focus. *Fails on* a seeded conditional region.
- **GS5-A2 no synthetic commands:** a runtime double counts intents during 60 s of idle Desk with
  a pending opportunity: zero. *Fails on* a seeded poll/keepalive intent.
- **GS5-A3 cap visibility:** fixture Lucky receipt with `saturated:true` renders the cap reason
  text. *Fails if* receipts are discarded (F2 regression).
- **GS5-A4 composed:** DOM-only manual clicks on the real server until an opportunity is
  projected, then a DOM claim; the next snapshot shows a buff or credited Lucky delta. *Depends on*
  OD-13 (the current seconds-scale schedule makes this feasible; an hours-scale retune would need a
  test epoch, not a fixture clock).
- **GS5-A5:** axe on region states; 320 px; keyboard claim.

---

### GS6 — Purchasable content and the shop (Desk amendments; no shop surface)

**Backend reality:** purchasable content = generators, upgrades, ladders, pools, provisioning
(`docs/purchasable-content.md`) — already surfaced on the Desk except provisioning caps. Cosmetics
have **no backend** (F5). This section therefore specifies only what exists and records the shop
as DG-5.

**GS6.1 Provisioned counts and caps.** v4 `generators[]` rows add
`provision_cap: null | {amount: int, reason_key: id}` from the pinned economy catalog's per-target
provisioning cap (the receipt already exports it). The Desk generator card renders
`desk.provisioned_frame {count:integer}` when `provisioned > 0`, and when
`provisioned >= provision_cap.amount` renders the cap reason text (for `generator.beige_tower` the
existing `generator.beige_tower.provisioned_cap` key from presentation). A frozen counter is never
shown without its reason (design law 5).

**GS6.2 Upgrade state text.** Owned upgrades render `desk.upgrade.owned` text instead of relying on
a disabled button (non-color state, A3.2). The v3 row has no ineligibility reason; OD-15 decides
whether v4 adds one.

**GS6.3 Horse Armor shelf.** Unchanged: static, free, on the Desk, curtain intact
(`cosmetic.horse_armor_free.*`, owner-ruled GU-C14 copy). No `shop` surface, no nav entry, no
purchase control, no stateful ownership is introduced.

**States/intents/keyboard:** unchanged Desk behavior; `buy_generator`/`buy_upgrade` gain the GS0.2
rejection display (e.g. `not_eligible/owned`, `unaffordable/*` → `desk.rejection.*`).

**Copy keys:** `desk.provisioned_frame {count:integer}`, `desk.upgrade.owned`,
`desk.rejection.owned`, `desk.rejection.unaffordable`, `desk.rejection.not_in_window`.

**Acceptance:**

- **GS6-A1:** fixture with `provisioned == provision_cap.amount` renders the reason text. *Fails on*
  a seeded card that shows only the number.
- **GS6-A2:** no component submits any intent whose kind is not in the production intent list for a
  cosmetic ID (source test). *Fails on* a seeded "buy horse armor" button.
- **GS6-A3:** Desk 320 px regression (the measured 647 px defect in the accessibility draft) is
  owned by that RFC; this RFC's additions must not widen the Desk (fixture measurement before/after).

---

### GS7 — The Pitch: `minigame_session` surface (proposed MA-C9 enumeration)

**Status of this section:** a **proposal for the ruling author of
`rfc/minigame-api-and-surface.md`** to reconcile into MA3/MA-C9 (execution-queue item 3h). Until
that RFC's body adopts it, this section is not implementation authority for AC5. It binds only the
Pitch half of MA-C9; the recovery surface is OD-14.

**Backend (already mounted through the registry):** `server/account/minigame_api.go`,
`server/account/minigame_api_schema.go` (operations `create_minigame_session`,
`get_current_minigame_session`, `play_minigame_command`, `resolve_minigame_session`; exact
`APIError` pairs per status); generated DTOs in `client/src/api/generated/types.ts`
(`MinigameSessionResponse*`, `MinigameCurrentResponse`, `PitchSnapshot`, `MinigameTenantCommand`,
`MinigameResolutionReceipt`); engine `server/pitch/`, client `client/src/pitch/`.

**Surface:** `minigame_session` / `game.surface.minigame_session` / unlock `feature.minigame.pitch`.

**Availability — `MinigamesArm` (v4):**

```text
{ rows: [{minigame_id: "pitch", unlocked: bool,          // FiscalUnlocks["minigame.pitch"]
          human_content_locked: bool,                     // soul gate human_hobby at Founder Soul
          active_session: bool}] }                        // same read-only predicate as MA-C12
```

`active_session` also feeds the v3/v4 `transitions.wind_down.eligible` preview (fixes F9 — the
preview becomes `false` while a session is active, so the Desk shows the Wind Down reason instead of
a silent rejection).

**API-client binding (the 3h dependency):** components never call `fetch`. `runtime.ts` exposes

```ts
interface MinigameSessionPort {
  current(): Promise<MinigameCurrentResponse>;
  create(minigameId: string, idempotencyKey: string): Promise<MinigameSessionResponseActive>;
  command(sessionId: string, request: MinigameCommandRequest): Promise<MinigameSessionResponse>;
  resolve(sessionId: string): Promise<MinigameSessionResponseTerminal>;
}
// Non-2xx → throws MinigameAPIError { status, category, detail } validated against the exact
// registry error pairs; network failure → throws MinigameTransportError.
```

typed exclusively with the generated DTOs. When API Foundation's generated client (queue item 3g)
lands, the port's implementation delegates to it with no change to this interface (OD-2). A
second hand-written wire type is forbidden.

**Tenant registry:** `client/src/game-ui/minigame-tenants.generated.ts`, generated from the pinned
`balance/minigame-api/first-content.json` `tenants` rows, maps `(engine_ref, engine_version)` →
child component key. Pitch `("pitch","1.0.0")` → `PitchTable`. An unregistered tuple renders the
`error` state with `unknown_id/minigame_tenant` (fail closed); no hand-maintained second registry.

**Pitch content binding:** the snapshot carries only instance IDs and `pitch_content_hash`. The
client bundles the pinned `balance/pitch.json` bytes, verifies SHA-256 equals
`snapshot.pitch_content_hash`, and resolves card `base_metric`/`copy_key`, hack `price`/effect arm
and policy caps from it. Mismatch → `error` state (`minigame.error.content_mismatch`); no
deploy-current fallback (OD-16 offers the alternative).

**Surface props and state (TypeScript):**

```ts
type MinigameSessionSurfaceProps = Readonly<{
  port: MinigameSessionPort;
  availability: MinigamesArm["rows"][number];      // the pitch row
  era: CopyEra;
  transportReady: boolean;
  onExitToHost(): void;                             // leave table; session persists server-side
  onTerminal(receipt: MinigameResolutionReceipt): void; // host runs one authoritative refresh()
}>;

type MinigameSurfaceState =
  | { kind: "loading" }
  | { kind: "none" }                                // current() → {kind:"none"}: launcher
  | { kind: "active"; session: MinigameSessionDescriptor; snapshot: PitchSnapshot;
      inFlight: null | { commandId: string; request: MinigameCommandRequest } }
  | { kind: "paused_reconnect"; last: null | { session: MinigameSessionDescriptor;
      snapshot: PitchSnapshot }; retry: null | { commandId: string; request: MinigameCommandRequest } }
  | { kind: "required_terminal"; response: MinigameSessionResponseTerminal }
  | { kind: "error"; category: string; detail: string };

type PitchTableProps = Readonly<{
  snapshot: PitchSnapshot; content: PitchContent;   // hash-verified bundled artifact
  era: CopyEra; pending: boolean;
  onPlayHand(cardIds: readonly string[]): void;     // 1..policy.play_size instance IDs
  onBuyHack(offerId: string): void;
  onEndShop(): void;
}>;
```

MA-C9's ruled closed set is `loading|active|paused_reconnect|required_terminal|error`. This
proposal adds `none` as the launcher state because MA-C6 ruled `{kind:"none"}` as a first-class
current-session answer; if the ruling author prefers, `none` becomes a host-level launcher outside
the surface and the five-state set stands unchanged (OD-17).

**Lifecycle and transitions:**

| From | Trigger | To |
|---|---|---|
| mount | `current()` in flight | `loading` |
| `loading` | `{kind:"none"}` | `none` |
| `loading` | `{kind:"active", session.status:"active"}` | `active` |
| `loading` | `{kind:"active", session.status:"claimed"}` | `paused_reconnect` (another request holds the claim); retry `current()` after 1 s, at most 5 times, then `error conflict/minigame_session` |
| `none` | Start pressed | `create("pitch", key)` with a fresh opaque key held in memory until success → `active` |
| `active` | command sent | `active` with `inFlight` set; all table controls `aria-disabled` |
| `active` | response `status:"active"` | `active` with new snapshot, `inFlight` cleared |
| `active` | response `status:"resolved"` | `required_terminal`; `onTerminal(receipt)` |
| `active` | transport error with `inFlight` | `paused_reconnect` with `retry = inFlight` |
| `paused_reconnect` | host `transportReady` or Retry pressed | resend **the same** `command_id` and body (MA-C13 idempotent receipt); if none pending, `current()` |
| any | 409 `conflict/minigame_revision` | `current()` then `active` (selection cleared, status line) |
| `required_terminal` | Return pressed | `onExitToHost()` |
| any | unmapped error | `error` |

`resolve()` is used only as the MA-C4 retry/read: when a terminal command's response was lost and
the retried command reports the session terminal, or when the surface holds a `session_id` whose
last snapshot phase was `terminal` without a receipt. After a page reload a resolved session's
receipt is not re-shown (current returns `none`); its payout is already in the Game UI snapshot.
Retries never mint a new `command_id` for the same player action.

**Commands:** `play_hand {card_ids}` (selected instance IDs `<card_id>#<ordinal>`),
`buy_hack {offer_id}`, `end_shop`, each wrapped as
`{command_id: newIntentID(), expected_revision: session.revision, command}`.

**Error mapping** (exact registry pairs from `minigameErrorJSON`):

| Status | Pair | Surface result | Copy key |
|---|---|---|---|
| 409 | `not_eligible/fiscal_unlock_required` | `none` + status | `minigame.rejection.fiscal_unlock_required` |
| 409 | `not_eligible/human_content_locked` | `none` + status | `minigame.rejection.human_content_locked` |
| 409 | `not_eligible/exclusive_activity` | `none`/`active` + status | `intent.rejection.exclusive_activity` |
| 409 | `not_eligible/illegal_phase` | refetch via `current()` | `pitch.rejection.illegal_phase` |
| 409 | `not_eligible/hand_too_large` | stay `active` | `pitch.rejection.hand_too_large` |
| 409 | `not_eligible/duplicate_card`, `unknown_card` | stay `active`, clear selection, invariant | `pitch.rejection.bad_selection` |
| 409 | `not_eligible/insufficient_currency` | stay `active` | `pitch.rejection.insufficient_currency` |
| 409 | `not_eligible/hack_slots_full` | stay `active` | `pitch.rejection.hack_slots_full` (+ `cap.pitch_hack_slots`) |
| 409 | `not_eligible/unknown_offer` | refetch | `pitch.rejection.unknown_offer` |
| 409 | `conflict/minigame_revision` | refetch | `minigame.conflict.revision` |
| 409 | `conflict/minigame_session` | `paused_reconnect` | `minigame.state.paused_reconnect` |
| 409 | `idempotency_conflict/minigame_command` | `error` + invariant (client defect: same ID, different body) | `minigame.error.generic` |
| 409 | `idempotency_conflict/minigame_session` | `error` + invariant | `minigame.error.generic` |
| 404 | `unknown_id/minigame_session` | `current()` | — |
| 404 | `unknown_id/minigame_tenant`, `unknown_id/founder` | `error` | `minigame.error.generic` |
| 400 | `invalid/*` | `error` + invariant | `minigame.error.generic` |
| 401 | `unauthorized/access_token` | host credential/offline path | existing |
| 429 | `rate_limited/*` | stay; re-enable after 1 s | `intent.rate_limited` |
| 500/503 | `internal_invariant/*`, `not_configured/minigame_api` | `error` | `minigame.error.unavailable` |

**Keyboard map:** launcher: Start button. Table: hand as a `<fieldset>` of native checkboxes (Space
toggles; once `play_size` are checked, the remainder become `aria-disabled` with
`cap.pitch_play` text visible — never silently ignored), then "Pitch these cards" button; shop:
one button per offer, then "Close the shop" (`end_shop`); "Leave the table" button last
(`onExitToHost`, session persists). No hotkeys, no Escape binding, no drag.

**Focus/announcement:** on each new snapshot, focus stays on the pressed control if still present;
if the hand is replaced, focus moves to the hand fieldset legend (`tabindex=-1`). Round result and
phase change are announced once in the surface status region (`pitch.phase_announcement
{phase:string, round:integer}`); the terminal state focuses the terminal heading (A1.1 forced
context change).

**Terminal rendering:** from `MinigameSessionResponseTerminal` only: round reached (snapshot
`round`), `round_best_valuation` vs `funding_target` (both Amounts), `credited_delta` of
`credited_resource_id`, the `cap_reason_key` text when `configured_cap_forfeit_units > 0`, and the
quality grade change. A "funded / not funded" headline is **not** rendered: the terminal outcome
fact is not on the wire (DG-6).

**Reduced motion:** no card deal/flip animation in any mode.

**320 px:** hand as a wrapping grid `minmax(min(8rem,100%),1fr)`; shop stacks; status header wraps.

**Curtain and formulas:** each hack shows a mechanically generated effect line from its effect arm
(`pitch.effect.*` keys with the factor/amount as `canonical_decimal` params — design law 9); the
`dark_pattern` hack additionally shows its curtain (`pitch.hack.dark_pattern.disclosure`). Card
tiles show `base_metric` via `pitch.card.metric_frame`.

**Copy keys:** `surface.minigame_session.title`, `minigame.launcher.pitch.title`,
`minigame.launcher.pitch.description`, `minigame.launcher.start`, `minigame.launcher.locked_fiscal`,
`minigame.launcher.locked_soul`, `minigame.launcher.resume`, `minigame.state.loading`,
`minigame.state.paused_reconnect`, `minigame.leave_table`, `minigame.retry`,
`minigame.conflict.revision`, `minigame.error.generic`, `minigame.error.unavailable`,
`minigame.error.content_mismatch`, `minigame.rejection.fiscal_unlock_required`,
`minigame.rejection.human_content_locked`, `pitch.round_frame {round:integer, rounds:integer}`,
`pitch.target_frame {target:canonical_decimal}`, `pitch.best_frame {valuation:canonical_decimal}`,
`pitch.hands_remaining_frame {count:integer}`, `pitch.currency_frame {amount:integer}`,
`pitch.deck_frame {count:integer}`, `pitch.hand_label`, `pitch.play_selected {count:integer,
max:integer}`, `pitch.shop_label`, `pitch.buy_hack {price:integer}`, `pitch.end_shop`,
`pitch.slotted_label`, `pitch.card.metric_frame {metric:canonical_decimal}`,
`pitch.effect.card_factor {factor:canonical_decimal}`, `pitch.effect.flat_add
{amount:canonical_decimal}`, `pitch.effect.shape_factor.pair {factor:canonical_decimal}`,
`pitch.effect.shape_factor.full_hand {factor:canonical_decimal}`, `pitch.effect.chain_factor
{partner:string, factor:canonical_decimal}`, `pitch.hack.dark_pattern.disclosure`,
`pitch.phase_announcement {phase:string, round:integer}`, `pitch.phase.playing|shop|terminal`,
`pitch.terminal.heading`, `pitch.terminal.round_frame {round:integer}`,
`pitch.terminal.payout_frame {amount:canonical_decimal}`, `pitch.terminal.return`,
`pitch.rejection.illegal_phase|hand_too_large|bad_selection|insufficient_currency|hack_slots_full|unknown_offer`;
existing `pitch.card.*`, `pitch.hack.*`, `cap.pitch_*`, `cap.minigame_faucet`.

**Acceptance (these are proposed AC5 sub-gates for the Minigame RFC):**

- **GS7-A1 port boundary:** boundary scan proves no component under `client/src/game-ui/` imports
  `fetch` or declares a minigame wire type outside `api/generated`. *Fails on* a seeded raw fetch.
- **GS7-A2 idempotent retry:** runtime double drops the response to a committed `play_hand`; the
  surface's retry carries the identical `command_id` and body and renders the stored response.
  *Fails on* a seeded implementation that mints a new `command_id` (double records two IDs).
- **GS7-A3 registry fail-closed:** a snapshot with `(pitch, 9.9.9)` renders `error`. *Fails on* a
  seeded default-to-Pitch fallback.
- **GS7-A4 content hash:** a bundled artifact whose SHA differs from `pitch_content_hash` renders
  `error content_mismatch`. *Fails on* a seeded skip of the comparison.
- **GS7-A5 states:** each of `loading|none|active|paused_reconnect|required_terminal|error` and every
  error-table row has a browser test with its copy key. *Fails if* any row maps to the generic line
  when a specific key is declared.
- **GS7-A6 Wind Down honesty:** with an active session the Desk Wind Down button is disabled with
  its reason. *Fails on* the current projector (F9) — the test is the discriminating witness.
- **GS7-A7 composed:** DOM-only on the real server: Fiscal unlock (GS1-A6) → Start → play hands →
  `end_shop`/next round → terminal → Return; the Game UI snapshot's `company.cash` reflects
  `credited_delta`; a mid-run socket close exercises `paused_reconnect` and resumes. *Fails if* the
  command route is severed or the terminal refresh is skipped.
- **GS7-A8:** axe (all states × three engines), keyboard-only full run, 320 px.

---

## Deviations from design

1. **Return-sequence modal (design/11 §1).** Design allows one modal on return — the ripe Fiscal
   Quarter prompt. This RFC ships a nav badge instead, because the return sequence (diorama
   fast-forward, offline-gains line) it belongs to is unbuilt; a lone modal outside that sequence
   would be an unexplained interruption (OD-3).
2. **Pet visible-behavior cues (design/04 §1).** Walk speed, ears, desaturation and cursor-seeking
   are not rendered; `behavior_state`/mood are text. No renderer exists and A4.3 forbids inventing
   one here.
3. **Golden-opportunity presentation (design/02 §1).** No animated golden object; a plain region and
   button. Same reason.
4. **"Clout" naming (design/02 §6).** The achievements score is not labeled Clout, because the
   Achievements foundation does not mint Clout (docs/achievements.md) (OD-6).
5. **The shop (design/04 §4, roadmap v0.1 "shop").** Not delivered: no cosmetic backend (DG-5).

## Owner decisions (numbered; recommended default first)

- **OD-1 Projection shape.** *Recommended:* one additive `game_ui_snapshot` v4 with a `features`
  object (one revision coordinate, one bind path, one refresh). *Alternative:* one read operation per
  system (more requests, independent caching, more drift surface).
- **OD-2 API-client binding for the Pitch.** *Recommended:* runtime-owned `MinigameSessionPort` typed
  with generated DTOs now; swap implementation to the 3g generated client when it lands, with an
  exact-error conformance test. *Alternative:* block GS7 on 3g.
- **OD-3 Ripe-quarter modal.** *Recommended:* nav badge now; modal deferred to a Return-Sequence
  RFC. *Alternative:* ship the single modal now.
- **OD-4 Fiscal multi-level spend.** *Recommended:* UI offers `levels: 1` only (server-resolved
  cost shown). *Alternative:* a level-count input; requires a server-projected cost table (no client
  triangular arithmetic).
- **OD-5 Fiscal composed proof under a retuned clock.** *Recommended:* if DG-1 retunes to hours, add
  a dedicated test epoch with short windows used only by the composed lane. *Alternative:* accept a
  fixture-level proof for the harvest window and composed proof for spend only.
- **OD-6 Achievement score vs Clout.** *Recommended:* label it "score" until a Clout RFC mints Clout.
  *Alternative:* call it Clout now (a copy-level claim the backend does not honor).
- **OD-7 Achievements visibility.** *Recommended:* show all rows, locked ones with their text (the
  current copy states the condition; nothing is secret) and an earned/total count. *Alternative:*
  hide locked rows; or add a run/career filter.
- **OD-8 p(doom) at T0–T1.** *Recommended:* show it (design/02: all meters visible; the backend
  already moves it on `externality.emitted`), with the pressure-meter tooltip. *Alternative:* withhold
  it until its design tier (Tier 5+), accepting a hidden moving number.
- **OD-9 Active play placement.** *Recommended:* Desk region only. *Alternative:* an additional
  chrome badge visible from every surface.
- **OD-10 Pet Trust display.** *Recommended:* not shown numerically; mood/band text carry it
  (sincerity, not a score to optimize). *Alternative:* show Trust as a visible bar (hardcap-visibility
  reading).
- **OD-11 Sincere tone.** *Recommended:* use `diegetic` + a pet-scoped lint rule now; add a
  `sincere` tone to the copy schema only if the lint proves insufficient. *Alternative:* add the tone
  now (Copy Pipeline amendment).
- **OD-12 Opportunity timing display.** *Recommended:* attended seconds as of last snapshot, not
  extrapolated, with the "pauses while away" note; announce spawn once. *Alternative:* no remaining
  time at all; or a wall-clock extrapolation (rejected by this draft as potentially false).
- **OD-13 Opportunity schedule scale (F8).** *Recommended:* owner confirms the seconds-scale
  schedule is intended for T0–T1 before GS5 ships; otherwise route a retune to a balance mint.
- **OD-14 Soul-recovery surface.** *Recommended:* the Minigame RFC's ruling author enumerates it in
  the same MA-C9 reconciliation, using GS7's state/port pattern; this RFC stays Pitch-only.
  *Alternative:* extend this RFC with a GS8.
- **OD-15 Upgrade ineligibility reason.** *Recommended:* v4 adds `ineligible_reason:
  null|"owned"|"window"|"requirement"|"unaffordable"` derived from the existing projector checks.
  *Alternative:* keep boolean-only and show only "owned".
- **OD-16 Pitch presentation content.** *Recommended:* bundle the pinned `pitch.json` bytes and
  verify by hash. *Alternative:* presentation v4 carries Pitch rows (copy keys only; metrics and
  prices would then be missing from the table).
- **OD-17 `none` state.** *Recommended:* add `none` to the MA-C9 closed set. *Alternative:* keep the
  five ruled states and move the launcher into the host.

## DESIGN-GAPs (with options)

- **DG-1 Fiscal clock policy (F7).** The live pinned windows are 100/200/300 ms. A Fiscal surface
  over them shows a quarter ripening three times a second. Options: (a) owner confirms and the
  surface ships as-is; (b) a balance mint retunes to design/02 §5 values before GS1 ships
  (recommended); (c) ship GS1 behind the unlock fact but hold the nav button until (b).
- **DG-2 Docs drift (F6).** `docs/fiscal-quarters.md` and `docs/minigame-the-pitch.md` say the
  mechanics are inactive; epoch 6 activated them. Options: (a) correct the docs in a separate change
  now (recommended; doc-only); (b) correct in this RFC's closeout.
- **DG-3 Consumer-less unlock (F12).** `unlock.arcade` costs credit and does nothing. Options:
  (a) withhold rows without a presentation row (this draft's behavior); (b) remove the row in the next
  balance mint; (c) specify the arcade consumer in a design/03 §8 RFC.
- **DG-4 Pet identity.** No producer creates a pet; design/04 says every founder gets one at Tier 0.
  Options: (a) a Pet Identity RFC (species, name, acquisition, starter-pet activation at a new-run
  boundary) — recommended, and required before GS4 mounts; (b) keep GS4 unmounted indefinitely.
- **DG-5 Cosmetics/shop.** No cosmetic items, ownership, grant or intent exists. Options: (a) a
  Cosmetics Foundation RFC ($0.00, curtain-pulled, stateful ownership, collection book) before any
  shop surface — recommended; (b) keep the static Horse Armor shelf as v0.1's shop and amend the
  roadmap row.
- **DG-6 Pitch terminal outcome.** `funded | funding_failed` exists in the certified result but not
  in `MinigameResolutionReceipt`. Options: (a) additive receipt field via an API amendment;
  (b) render round/valuation/payout only (this draft).
- **DG-7 Missing reason copy (F10).** `cap.fiscal_credit`, `cap.fiscal_level.beige_tower`,
  `cap.active_combo`, `cap.cash` must be owner-authored and registered through `copy/references.v1.json`
  before their surfaces can render caps; `capFor` throws today.
- **DG-8 Opportunity spawns require commands.** A fully idle player never sees an opportunity
  (lazy scheduler). This is backend behavior, surfaced honestly, not fixed here. Options: (a) accept
  (active play rewards presence — recommended); (b) a design ruling on an idle-visible spawn, which
  would be a new scheduler mechanic and its own RFC.

## Acceptance criteria — every gate must discriminate

1. **Per-surface gates:** GS1-A1–A6, GS2-A1–A5, GS3-A1–A5, GS4-A1–A5 (A5 recorded as an explicit
   blocker), GS5-A1–A5, GS6-A1–A3, and — once adopted by the Minigame RFC — GS7-A1–A8, each with its
   seeded failing case executed and recorded (evidence discipline 1).
2. **Projection parity:** Go projector and TS decoder share v4 fixtures; the TS decoder rejects each
   of: extra key, missing arm key, unsorted rows, out-of-domain integer, non-canonical Decimal
   (five seeded fixtures, each must fail).
3. **Runtime receipts:** a unit test proves `intent()` returns both outcome arms and throws on an
   unknown `outcome`; restoring `Promise<void>` fails GS1-A4/GS5-A3.
4. **Boundary:** `make verify-client-boundary` passes with the new components (no transport/replay
   imports, no literals, governed styles only); a seeded literal fails it.
5. **Accessibility:** `make test-browser` axe (WCAG 2.2 AA, zero serious/critical) over every new
   surface/region state in Chromium, Firefox and WebKit; 320 px and 400 % zoom checks over the full
   mounted page including chrome and the grown nav; keyboard-only completion of every built task.
6. **Composed:** `make test-game-ui-composed` extended with the GS1-A6 → GS7-A7 Fiscal-to-Pitch
   journey and GS2-A4/GS3-A4/GS5-A4, DOM-only, no intent bypass or fixture clock. GS4-A5 is listed as
   blocked, never vacuously green.
7. **Performance:** the existing 60-second observable scenario passes unchanged with all regions
   mounted on the Desk (no new long task > 200 ms).
8. **Closeout:** `docs/game-ui.md` (and DG-2 docs if still stale) updated; per-change review
   recorded; cross-party designated review cites the exact range before archival.

## Open questions

- Whether the draft Accessibility RFC is accepted before this one; if not, GS0.6 is this RFC's own
  floor and the two are reconciled at acceptance.
- Whether the Phase-0 manifest (D-007) pulls any of these surfaces into the preview. This draft
  assumes none.

## Changelog

- 2026-09-25: created (draft) from a source read at `c2d9bbc`. Records findings F1–F13, proposes
  the v4 projection, runtime receipt handling, seven surface/region contracts, and the GS7 Pitch
  enumeration for Minigame RFC MA-C9 reconciliation. No implementation authorized.
