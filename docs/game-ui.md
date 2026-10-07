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
  Reputation R9's available-balance and next-route display is **not implemented**:
  its post-Exit balance source must be reconciled with GU-C3's payload-only boundary.
  The runtime now decodes the existing `run_started` v1/v2 payload, preserving
  absent/null/object tree arms, canonical factor and applied-starter array order.
  This reader is plumbing, not a rendered carry-over summary or AC12 acceptance.
- Settings/System: save status, drain notice, and explicit resync action.
- **Trophy Case (`achievements`):** read-only, unlocked by `feature.achievements`. It shows the run
  and career score, an earned/total count, and each row's scope, text state (earned this run,
  earned in career, not earned yet), score grant, and the existing possession warning. The score is
  never labelled Clout.
- **Achievement/meter announcements:** exact event decoders drive the single polite chrome
  announcement, not game-state arithmetic. Achievement copy is announced once per scoped cursor;
  meter changes announce on the Meters surface or badge its nav until visited. Unknown presentation
  IDs are withheld and reported. Meter values must be integers in 0–100, with strictly increasing
  values for `up` and strictly decreasing values for `down`, matching the server validator.
  Contradictory directions throw into authoritative recovery rather than announcing (RP-312).
  Fake-socket reconnect and native Chromium/WebKit host replay tests demonstrate suppression;
  this is not a real-server acquisition, Firefox or whole Garage acceptance claim.
- **Earnings Calls (`fiscal`):** unlocked by `feature.fiscal`.
  - Shows credit against its visible cap, the auto-sweep preview, the hoard preview (next run only),
    and a display-only phase (`fiscal-phase.ts`: ripening, early with its stated success chance, or
    guaranteed).
  - The Harvest button carries its curtain small print. There are +1 level buttons and unlock rows.
  - Unlock rows without a `features-presentation.json` row are withheld; `unlock.arcade` is
    withheld (F12).
  - Intents are Founder-scoped. Harvest outcomes and Fiscal rejections render in
    the Fiscal panel's own polite status line, never duplicated in chrome.
    A generator-level cap rejection uses that snapshot row's reason key, not
    the shared generic cap sentence. Unknown target refusals report one invariant.
  - Harvest, level and unlock controls keep native keyboard focus and visible
    pending text/`aria-disabled` while the intent and authoritative refresh are
    in flight; each component callback refuses pending activation. Native Tab,
    Enter and Space follow harvest → levels → unlocks. Actual ineligibility or
    unavailable transport still disables controls; stale values have the shared
    explanation. A removed focused control hands off to the nearest surviving
    enabled control in the same region, otherwise the heading, without taking
    focus from a different control the player selected. These native checks
    use runtime doubles; the separate composed lane exercises actual harvest
    and unlock against Postgres and a real WebSocket. Neither replaces Firefox,
    manual accessibility, all-Garage or release-manifest proof.
    The host maps the existing result/reason key once and passes a presentation-only
    `notice` to the panel. It captures the submitting surface before any queue/read
    await; late completion does not display on another tab. The host still stores
    only the last result, not an outcome history. Native runtime-double placement,
    completed-navigation and held-response controls cover Fiscal and care applied/
    ordinary refused outcomes. A separate native consumer census covers immediate
    and late-away HTTP 400/409/429/401/404/503, transport and malformed-response
    errors: panel-local messages, one held refresh for 409/429, fresh revision and
    intent ID only on new consent, or offline without replay. These errors are
    thrown by a runtime double, not parsed from real HTTP responses. This does
    not prove wire parsing, all-surface/queued origin, AT or real-service outcome
    delivery. Other Garage GS0.6 obligations remain independently open.
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
  focus to Confirm; Escape cancels and returns focus to Buy. Submission moves focus to the
  stable node row after the DOM updates, retaining it during pending and subsequent node-state
  replacement. Tab reaches the heading, then enabled Buy controls in artifact order;
  an open confirmation traverses Confirm then Cancel. These controls have explicit
  zero Tab indices for consistent native WebKit traversal; disabled buttons stay
  outside the sequence. Shift+Tab reverses the sequence. The heading's intentional
  Tab stop is the narrowly annotated R9 contract, not an interactive heading role.
  Rows accept programmatic focus but add no Tab stop. Owned, locked and unaffordable nodes
  show their state as text and have no control. Purchases send `purchase_reputation_node` at the
  Founder revision; an applied receipt refreshes the snapshot.
  Buy/Confirm stay disabled until both player and world subscriptions recover,
  including startup, drain and the interval between full sync and resubscription.
  The tree renders the existing offline status when transport is unrecovered or
  an authoritative read fails. An HTTP receipt does not optimistically update
  ownership or available Reputation; pending lasts through the authoritative
  refresh. Inactive arms hide the tab/surface; a mounted arm becoming null
  returns to Desk.
  The header binds available, earned level and spent separately; current-run
  bonus uses the supplied frozen factor (with a distinct no-row label), while
  next-run bonus uses the supplied current-Founder projection. Pending alone
  does not rewrite either factor. Known formula limitation (RP-293): pending
  copy receives raw ppm as `perlevel`/`unlock`, not R9's declared percentage
  parameters. This placeholder is not a published formula. R2 admits fractional
  percentages that cannot fit R9's integer types; author/copy reconciliation
  remains required, with no rounding or data-domain narrowing implemented.
  Every node displays its cost through the shared Amount component, after its body
  and before requirements/state. The cost remains visible for owned, locked and
  unaffordable nodes, and throughout confirmation, pending and authoritative
  replacement; it is not dependent on the Buy control. Integer costs use the
  existing canonical-number conversion and Standard notation, without changing
  purchase prices or Buy copy.
  Exit plan panels start with an empty selection and synchronize that selection
  to the host on mount, including after navigation or replacement by an Offer
  Sheet. An empty visible plan omits `reputation_plan` from the Exit intent;
  selections from a previous unmounted panel cannot cause invisible spending.
  The submitted row exposes `aria-busy="true"` while its existing host purchase task or the
  shared pending state remains outstanding; all Buy/Confirm controls stay disabled. Attribution
  clears after both settle, so a later unrelated refresh does not mark a completed row busy.
  Known purchase rejections appear in a polite status on the submitted row only;
  another submission clears previous row feedback. Revision-conflict receipts and
  typed HTTP409 errors use the declared Reputation key while retaining the shared
  authoritative refresh and its disabled-control boundary. The inline status replaces
  duplicate global rejection text on this surface; applied results still use the
  existing global status. The newly declared conflict text is explicitly pending owner
  copy, not adopted player prose. Unknown/network errors retain shared handling.
  For an active tree, the server rejects a null, malformed or unordered owned-node set and
  any unlock mirror that differs from the pinned tree's derivation. The client accepts only
  canonical bonus-factor strings of at least one; the current-run factor may remain null.
  `ReputationPlanPanel.svelte` is an advisory plan shown on the Offer Sheet and beside Wind Down,
  empty by default. It orders selections in tree order and gates them on prerequisites and a
  projected budget. Deselecting a node also drops the selections that depended on it. A non-empty
  plan is sent as `reputation_plan`; the server re-validates the whole plan.
  Native Tab reaches the disclosure summary, then enabled checkboxes in tree order
  and Clear when a selection exists; Shift+Tab reverses that sequence. Checkboxes
  and Clear have explicit zero Tab indices for WebKit traversal. Disabled controls
  remain outside the sequence. Space toggles a checkbox; Enter/Space opens the
  disclosure or activates Clear. This does not change the advisory budget,
  prerequisite rules, outgoing plan or authoritative preview source.
  Public arm/preview replacements update the displayed budget and unselected
  prerequisite/ownership controls. Known limitation (RP-292): an already selected
  node that becomes owned disappears but remains in the callback and cost sum;
  a reduced budget can leave a checked over-budget plan and negative projection.
  Clear resets both visible selection and outgoing plan. No automatic reset or
  pruning rule is implemented; R9 author reconciliation is pending. The server's
  atomic plan revalidation remains the authority, not this advisory display.
- **Desk opportunity region (GS5):** always present, directly after the manual action. It sits in a
  fixed DOM position and never moves focus, so an opportunity spawning never shifts the page. It
  renders the optional `features.opportunity` arm:
  - the pending opportunity, with its effect presentation row and the attended seconds remaining as
    of the last snapshot (never a wall-clock countdown);
  - live buffs;
  - the combo hardcap label and projected number while live buffs are present,
    plus the saturation explanation when the kernel clamp saturates.
  
  Claim sends a Company-scoped `claim_opportunity`. An applied receipt's `receipt.opportunity`
  evidence shows the lucky credit, and the `cap.cash` reason when saturated. Expired and gone
  claims map to typed notices. An unknown opportunity ID retains the not-pending
  notice and reports one client invariant; ordinary expired/not-pending refusals
  do not. The shared host likewise reports unlisted rejections and typed invalid
  requests once, without logging request bodies or identifiers. A buff receipt's
  non-null cap reason remains visible
  even if the successor snapshot has no live capped buffs; this is receipt evidence,
  not a claim that the current buff product is still saturated. Claim has an explicit
  zero Tab index: native Tab from the manual button reaches it in Chromium and
  WebKit, then Enter or Space activates the normal Company intent. Focus is not
  moved on spawn. The idle witness observes an actual minute of native timers,
  not a display-clock jump. The UI never sends a synthetic command to cause a spawn. A spawn is
  announced politely once per `opportunity_id` while the Desk is mounted.
  The composed claim checker binds the response to its DOM request and exact
  next Company revision/run. Lucky checks the canonical credited delta and
  exact receipt-snapshot cash against the successor's unique cash row; the
  receipt cash already includes lazy accrual, so no payout formula is duplicated.
  A positive Lucky credit also needs a matching cash change with bank growth and
  a total delta containing that credit. Zero is admitted only with the receipt's
  saturation evidence and an actual successor bank at its published hardcap.
  Buff claims check a unique matching live ID and effect. The pure checker is
  shared by the real driver and counterexample tests, uses the existing numeric
  parser, and reports which branch actually ran; its fixtures are not server
  payout or Lucky-acquisition evidence.
- **Pet (GS4):** a nav tab unlocked by `feature.pets`, which the server sets true only when an adopted
  pet exists. It renders the PA7 projection only: name, status band text, and one button per
  catalog care action in catalog order. An action the server does not list as eligible is disabled
  and says so in text (never greying alone). Care sends a Founder-scoped `care_action`, and the four
  `not_eligible` pairs map to typed notices. A worn cosmetic mounts `CosmeticOverlay` on the live
  pet's sprite (cosmetic-shop G10). GS4's raw stats, mood, behaviour and cooldowns are not projected:
  PA7 forbids them, and the conflict is recorded as a DESIGN-GAP in
  `planning/garage-player-surfaces/log.md`.
  Care actions remain focusable with `aria-disabled` and visible `common.pending`
  until both the intent and authoritative refresh settle; pending activation is
  ignored. Actual ineligibility or unavailable transport still disables them,
  with the stale-state note visible. Explicit zero Tab indices support native
  Tab/Enter/Space in Chromium and WebKit. If the currently focused action becomes
  disabled after a refresh, focus returns to the care heading without taking it
  from another control the player has selected.
  The controlled real-server Cosmetics driver adopts a pet, equips through the shelf, checks
  the live annoyed/no-text overlay and reload, exercises the browser's reduced-motion preference,
  and verifies Unequip removes the overlay without losing ownership. It also feeds
  that actually adopted pet through its DOM care control, binds a positive care
  receipt to the request and Founder revision, checks the next persisted public
  band/eligibility and rendered disabled Feed, and reloads that state. This does
  not require or prove a status-band crossing or the missing status announcement.
  It uses a test-only epoch,
  not a production content pin; the GS4×PA7 contract and G10 release gates are not closed by it.
- **Fiscal badge and buff announcements (GS0.3 remainder):**
  - A `fiscal_period_harvested.v1` event badges the Fiscal nav `(harvested)` while the player is
    elsewhere, and visiting Fiscal clears it (OD-3: no modal).
  - A `buff_started.v1` event announces politely while the Desk is mounted.
  - Both decoders mirror the exact server validators and fail closed.
- **320 px reflow:** the Desk, Fiscal, Meters, Trophy Case and pet surfaces are measured in all three
  browsers at a 320 CSS px viewport, with no element past the viewport edge and no horizontal
  scroll. The chrome nav wraps, and card grids use `minmax(min(…, 100%), 1fr)`.
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

`active_play` and `pets` stay null-only on the wire, because the compatibility gate rejects widening
a registered null-only property. Their projections land as the optional sibling arms
`opportunity` and `pet_adoption`, alongside `reputation`, `cosmetics` and `axis_stack`. Each is
documented in its own section, and the Reputation and Clout surfaces are covered in
[Reputation tree](reputation-tree.md) and [Axis stack](axis-stack.md).

- **`opportunity` (GS5):** Company v18 with the opportunities artifact. It carries the pending
  opportunity only while `expires_attended_ms > attended_now_ms`, and the live buffs only. The
  combo `saturated` flag comes from the read-only kernel export `production.ProjectActiveCombo`,
  the same clamp rates use.

`feature.active_play` follows `opportunity`. `feature.pets` is true only when `pet_adoption` holds
at least one pet.

Intents return their typed outcome (GS0.2). A rejected intent is an HTTP 200 whose
reason renders in `role="status"`: Fiscal and pet care own their outcome region;
other existing host results use the chrome line only while their captured
submitting surface is current. Care receives the host's single mapped presentation
key just like Fiscal. Cross-surface event announcements keep their separate,
cursor-deduplicated chrome region. A stale revision (`revision_conflict`) triggers one
authoritative refresh and is never auto-retried. Founder-scoped intents send the Founder revision.
An applied intent keeps its controls pending through an authoritative snapshot refresh, so the
next action uses the updated Founder or Company revision even if the stream receipt arrives late.
HTTP429 rate limits and `not_eligible/exclusive_activity` rejections also trigger
that authoritative refresh. Pending controls cannot reactivate while it is held;
the response never retries the rejected command. A subsequent player action
uses the refreshed revision and a fresh intent ID. Native care controls retain
focus and `aria-disabled` pending semantics during that read.
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
An ordinary unexpected close emits the internal `transport_recovering` notice,
which marks the host offline and transport not ready without disposing its
subscription. The existing reconnect still owns its delay, saved positions and
recovery/full-sync path; both subscribe acknowledgements emit `transport_recovered`.
Terminal `transport_closed` retains its existing disposal behavior. These are
internal host notifications, not new wire messages or authentication policy.
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
does not itself refresh into the server-created next run. The HTTP applied-action
path and transport recovery can nevertheless sample that next run while Run End
stays selected. Only **Start the Next Company** clears the terminal state, after
a fresh exact-successor snapshot check. A concurrently generated offer delivered
after `run_ended` also cannot replace the terminal screen.

If an HTTP sample has already advanced beyond `run_started`, the runtime still
delivers its immutable summary only for the sampled Founder, exact run sequence
and start time, on the Company scope. Old/future-run or other-Founder duplicates
remain suppressed; channel-offset dedup and revision-gap recovery stay unchanged.
One last-delivered start-event ID prevents that exception from reviving the same
summary at a new outbox offset. A delivered Company-start revision high-water
also suppresses a superseded start after a newer start has arrived, even when
the last HTTP snapshot still describes the older run. Both values survive
socket reconnect/full-sync within the subscription; this is bounded memory,
not an unbounded history set. It does not alter the generic revision cursor,
offset persistence or recovery policy. Delivery does not navigate or derive a
balance from the event's factor/starter IDs.

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
checked. The player then visits Earnings Calls and uses its rendered harvest and Pitch-specific
unlock buttons; both issue real applied intents over the public API, without a database setup write.
Cards are
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
