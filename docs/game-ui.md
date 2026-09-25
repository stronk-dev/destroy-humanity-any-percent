# Game UI

The Game UI is the Svelte Phase-A play surface mounted by the production client entrypoint. It
consumes the generated `game_ui_snapshot.v4` projection, decoded lifecycle events, bootstrap and
intent operations, and the ratified Copy/Presentation catalogs. Components do not import transport
or replay internals; `client/src/game-ui/runtime.ts` owns HTTP, WebSocket, and envelope decoding.

## Shipped surfaces

- Vision Slide: silently creates the anonymous account through the idempotent bootstrap
  coordinator and persists credentials before entering play.
- Desk: manual action, resources and visible cap explanations, generator purchases, upgrades,
  server-projected Gate/Wind Down controls, local splits, the free Horse Armor shelf, shareware
  registration/order form, and README.TXT.
- Offer Sheet: authoritative exit type, complete payout terms, server-clock-relative expiry,
  Company-only decline, and Founder-CAS-guarded acceptance.
- Run End: a payload-isolated component that accepts only the decoded `run_ended` event; its parent
  owns the exact-next-Company continuation control.
- Settings/System: save status, drain notice, and explicit resync action.
- **Trophy Case (`achievements`):** read-only, unlocked by `feature.achievements`. It shows the run
  and career score, an earned/total count, and each row's scope, text state (earned this run,
  earned in career, not earned yet), score grant, and the existing possession warning. The score is
  never labelled Clout.
- **Earnings Calls (`fiscal`):** unlocked by `feature.fiscal`.
  - Shows credit against its visible cap, the auto-sweep preview, the hoard preview (next run only),
    and a display-only phase (`fiscal-phase.ts`: ripening, early with its stated success chance, or
    guaranteed).
  - The Harvest button carries its curtain small print. There are +1 level buttons and unlock rows.
  - Unlock rows without a `features-presentation.json` row are withheld; `unlock.arcade` is
    withheld (F12).
  - Intents are Founder-scoped. Harvest outcomes and Fiscal rejections render in the status line.
- **Reputation Board (`meters`):** read-only, unlocked by `feature.meters`. It is a 5 × 2 table of
  constituency Standing/Grievance plus p(doom). Each cell has a native `<meter>`, numeric text and
  band text. Below 30rem it collapses to labelled rows. It carries the curtain and the "as of last
  update" note.
- **Pitch availability:** the Pitch tab is unlocked by `feature.minigame.pitch`. Before any create
  request it shows the Fiscal-lock or Soul-lock reason from the minigames arm.
- **Desk additions:** provisioned counts show `desk.provisioned_frame`, plus the cap reason text
  once provisioning reaches its visible cap. Owned upgrades show text, not only a disabled button.
  A resource cap whose reason key has no copy is withheld with a loud `console.error` invariant
  instead of crashing the Desk (F10).
- **Arm going null:** if a mounted surface's arm becomes null, the UI returns to the Desk.

Mechanical ID → copy mappings for these surfaces live in `client/src/game-ui/features-presentation.json`
(strict, byte-sorted, every key checked against the copy catalog), not in code.
- Cosmetic shop: when the optional arm `features.cosmetics` is active and an item is acquirable or
  owned, the Desk shows the $0.00 shelf in place of the static Horse Armor card. Below Founder v24
  the static card is unchanged. See [Cosmetics](cosmetics.md).
- Pet adoption: when the optional snapshot arm `features.pet_adoption` is present, the Desk shows
  the inline adoption card and then the welcome state. See [Pet adoption](pet-adoption.md).
- Minigame session (The Pitch): a nav tab present whenever the runtime supplies a minigame port.
  It hosts `client/src/game-ui/minigame/MinigameSessionSurface.svelte`; see
  [Minigame platform § Client surface](minigame-platform.md#client-surface). Leaving the tab keeps
  the server session. A terminal receipt triggers one authoritative snapshot refresh.
- Reputation tree: a nav tab shown when the `feature.reputation_tree` fact is true. It hosts
  `client/src/game-ui/ReputationTreeSurface.svelte` over the optional v4 `features.reputation` arm,
  and every node state comes from the server. Buy opens an inline Confirm/Cancel pair and moves
  focus to Confirm; Escape cancels and returns focus to Buy. Owned, locked and unaffordable nodes
  show their state as text and have no control. Purchases send `purchase_reputation_node` at the
  Founder revision; an applied receipt refreshes the snapshot.
  `ReputationPlanPanel.svelte` is an advisory plan shown on the Offer Sheet and beside Wind Down,
  empty by default. It orders selections in tree order and gates them on prerequisites and a
  projected budget. Deselecting a node also drops the selections that depended on it. A non-empty
  plan is sent as `reputation_plan`; the server re-validates the whole plan.
- Recovery: a nav tab hosting `client/src/game-ui/soul/SoulRecoverySurface.svelte` whenever the runtime
  supplies a Soul-recovery port; a terminal recovery refreshes the snapshot once.

The persistent chrome derives its era only from the authoritative tier (`0` is `era_1995`, `1` is
`era_2000`, `2` is `era_2010`; tier 3 and above still throw until a later tier ships its era).
The `era_2010` token values are candidate design data awaiting owner ratification with the Tier 2
content. RTA uses the snapshot's server-time sample plus monotonic elapsed time. Gate splits and
personal-best timing are local display records only and never feed an intent or leaderboard.
Presence is hidden until the real world-channel count arrives; the UI never invents a visitor.
Presentation schema v3 owns the literal `$0.00` and pre-naming `Founder` constants, so copy
placeholders never substitute a formatted zero or an unrelated company label. Missing constants
throw. Payout labels and any shipped network-slot titles also resolve only through that catalog;
unknown future slot IDs are withheld rather than rendered mechanically.

Live sync requires snapshot v4: v3's positive `founder_revision` and exact `transitions`, plus the
Garage Player Surfaces GS0.1 `features` object and a `generators[].provision_cap`
(`null | {amount, reason_key}` from the pinned economy `provisioned_hardcap`). Snapshots v1–v3 stay
decodable only for stored bootstrap receipts.

`features` has exactly six keys. Each arm is `null` when the pinned bundle lacks its artifact or the
save predates the version that activates it:

- **`achievements`:** Company v16. Every pinned row, its earned state (`run`, `lifetime` or `null`;
  an overlap fails the projection), and both scores.
- **`meters`:** Company v16. Committed values with `meters.BandFor` bands; decay is never
  extrapolated.
- **`fiscal`:** Founder v19. Persisted credit, period and levels. The `sweep_preview` runs
  `fiscal.Catalog.Sweep` on a discarded clone at canonical server time and equals what the next
  harvest's auto-sweep reports. `next_level_cost` comes from `GeneratorLevelCost`, and
  `hoard.preview_ppm` is the published `min(credit_after, cap_credits) × ppm_per_credit`.
- **`minigames`:** Founder v21 with `minigame_api`. It restates the create gates read-only: the
  `fiscal_unlock` rule, the `human_hobby` Soul gate, and the active-session predicate for every
  supported tenant.

`active_play` and `pets` are always `null` in this implementation; the planning log records the
blockers. Facts add the six `feature.*` booleans derived from the arms.

Intents return their typed outcome (GS0.2). A rejected intent is an HTTP 200 whose reason renders
in the chrome `role="status"` line. A stale revision (`revision_conflict`) triggers one
authoritative refresh and is never auto-retried. Founder-scoped intents send the Founder revision.
Only transport failures and 401/404/5xx mark the UI offline.

The
`transitions.wind_down.eligible` preview applies the same read-only MA-C12 active-minigame
predicate that Exit freezes into replay. While a session is `active|claimed`, the preview is false,
matching the server's `not_eligible/minigame_session_active` rejection. A bundle that pins
`minigame_api` cannot be projected without that resolver: the projector fails loud instead of
offering a control the server would refuse. The previewed `cross_gate` is the next adjacent
standard gate: the uncrossed `gate.t0_to_t1` at Tier 0 and, only when the pinned routes declare it
(Tier 2 content, `rfc/tier2-content.md` §E2), the uncrossed `gate.t1_to_t2` at Tier 1. At Tier 0 it
is never the curriculum's scripted Exit, which requires the first gate already crossed. At Tier 1 a
run-1 Company whose scripted first failure is due routes every command into that Exit; the preview
does not model this, and the terminal receipt/event stays authoritative.

`transitions.incorporate` is an optional control, present exactly when `incorporate` can apply
(Tier ≥ 2 and no faction). It lists every pinned faction by `faction_id` with its existing
`incorporation_copy_key` and is omitted otherwise, so snapshots below Tier 2 are byte-unchanged
and the API compatibility pin is untouched. The client rejects the control below Tier 2 or with an
unsorted list or mismatched copy key. The Desk renders one button per faction, and each submits
`incorporate {faction_id}`. At `era_2010` the Desk also shows the FarmVille-era energy bar: a
presentation-only chrome stub with its curtain in both the tooltip and small print. It holds no
state, always reads full, and its Refill button emits no intent.
The Game UI projector derives the first Gate by invoking the existing production transition on a
discarded decoded-state clone and applies the existing Tier-1 Wind Down rule. The production
kernel itself is unchanged. The first Gate is the only Phase-A gate exposed;
later gates/routes fail closed. Eligibility is advisory and the intent receipt remains authority.
Encrypted bootstrap receipts remain replayable under the schema version they were minted with:
stored v1/v2 snapshots legally omit transition controls, and v1 also lacks the Founder coordinate.
Offer acceptance stays disabled until a current live sync supplies that coordinate.
The runtime persists positioned player/world subscriptions, recovers missed publications after a
drop, and falls back to the same authenticated live snapshot operation on a revision gap, expired
history, queue overflow, or invalid frame. Recovery snapshots are delivered into the existing
`bindSnapshot` path; there is no parallel API client or snapshot schema. Drain delays reconnect by
the server-advertised bound. Auth-expired and replaced sockets surface offline instead of inventing
Account token-rotation or multi-tab arbitration behavior.
Company Gate events trigger one deduplicated authoritative refresh before the next transition;
action and refresh state are tracked independently, and a click that races that refresh waits for
its revision instead of disappearing. When Gate is committed before the WebSocket subscription is
ready, the HTTP command path also fetches the authoritative snapshot; the same-revision
`gate_crossed` event can still record its split when it arrives.
Declining an exit offer likewise holds the action boundary through an authoritative snapshot, so
the next player intent cannot reuse the revision consumed by the decline. Terminal actions still
rely on ordered event delivery so an eager snapshot cannot suppress `run_ended`: Wind Down and
offer acceptance remain disabled until the player subscription reports `transport_recovered`, and
disable again on close, drain, or resync. Once `run_ended` arrives, the trailing command receipt
does not refresh into the server-created next run; that snapshot is bound only when the player
chooses **Start the Next Company**. A concurrently generated offer delivered after `run_ended`
also cannot replace the terminal screen.

## Boundaries and verification

`make verify-client-boundary` scans the Game UI components alongside the archived UI primitives.
It rejects transport/replay imports, raw network calls, player-facing text literals, and governed
style literals. `make test-browser` applies the WCAG 2.2 AA axe gate to all five lifecycle surfaces and to every minigame surface state in
Chromium, Firefox, and WebKit and includes the sixty-second observable performance scenario. The
focused `make test-game-ui-performance` command runs that scenario alone.
`make test-game-ui-composed` additionally drives Chromium through the real Vite proxy, composed
gameserver, Postgres bootstrap transaction, authenticated live snapshot-v4 route (asserting the live `features` arms), and Centrifuge
world subscription; its schema/revision and visible visitor-counter assertions prove the production
HTTP synchronization and WebSocket handshake completed. The composed witness then closes the
production browser socket, commits an intent while disconnected, and requires the reconnect to send
the exact persisted player epoch/offset and advance it by replaying the missed receipt. Ordinary
server-side setup then satisfies the first-gate requirement; visible enabled controls alone submit
Gate and Wind Down through `runtime.ts`, render scripted and standard Run End, and continue only
after fetching the exact successor run. No browser intent bypass, fixture clock, or two-hour replay
is used. The harness preflights exclusive ownership of its gameserver port, builds and starts one
ignored repository-local binary, and waits for that exact process on teardown; another listener
fails the witness before bootstrap. On the third run the witness opens the Pitch tab. Start first
receives the server's real 409 `not_eligible/fiscal_unlock_required`, and the launcher notice is
checked. The player then buys `minigame.pitch` with real `harvest_fiscal_period` and
`spend_fiscal_credit` intents over the public intent API; no database write is involved. Cards are
selected and played by keyboard until the session reaches a terminal `applied` receipt. The Game UI
must then fetch a snapshot at or beyond the receipt's Company revision, and `current` must read
`none`. `make verify-game-ui` composes the existing client, browser,
and composed lanes.

The deterministic performance lane runs in an isolated Chromium process after the functional
Chromium/Firefox/WebKit matrix, then feeds 1,200 authoritative snapshot updates representing 60
seconds at 20 Hz through 600 shared formatter windows in a 1280×720 viewport. Isolation keeps
concurrent browser engines from becoming part of the measurement. The shared renderer may commit
a hot Amount at most 600 times and may not produce a long task over 200 ms.
Four-times CPU throttling and the five-percent dropped-frame allowance remain a manual release
profile; deterministic CI gates observable commits and long tasks as ruled. The 2026-08-22
reference run passed for 60,001.9 ms at 4× in pinned Chromium: 598 production prediction
publications, 355 visible Amount mutations, no Long Tasks, and zero estimated missed frames across
7,200 observed frame intervals. The complete environment and predeclared calculation are preserved
with the archived Game UI planning record.

## Live-content boundary

Epochs 7 and 8 provide the live T0–T1 economy and first-hour curriculum. The content contract's
pinned-seed proof drives that sequence through the real composed gameserver and Postgres. The Game
UI does not duplicate that two-hour/full-script proof in Chromium. Its remaining browser gate is
deliberately narrower: server-authorized Gate and Wind Down controls, both terminal surfaces, and
Run-End continuation into the already-created next run, with discriminating disconnect and
suppression failures.
