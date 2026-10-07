# Garage Player Surfaces implementation log

## 2026-09-25 — Predeclaration, Batch A (Claude)

**Implemented by:** Claude, on the owner's 2026-09-24 direction that Claude implements accepted
RFCs. Every batch awaits Codex's designated cross-party review; nothing here is self-approved.

Order: GS0.2 runtime outcome fix first (client-only), then the v4 projection with the live arms,
then the surfaces. The snapshot version is the next free Game UI snapshot version at landing
(v4 is free: `schema_version` 3 is the current maximum). Copy for new keys is implementer-drafted
candidate text in `copy/catalog/garage-surfaces-candidate.json`, marked for owner adoption; it is
not ruled copy. Each gate ships with a demonstrated severing probe.

## 2026-09-25 — GS0.2 intent outcomes (Claude)

`runtime.intent()` now resolves the typed outcome, and a non-2xx intent throws
`GameUIRequestError`. `act()` takes a `scope`: a Founder-scoped intent sends the Founder revision
and is skipped while that revision is unknown. Rejections render in one host `role="status"` line
through `noticeForOutcome`, which checks surface rows first, then the category fallbacks. Only
transport failures and 401/404/5xx set `offline`.

Server fact recorded: a stale revision is HTTP 200 `rejected revision_conflict/expected_revision`
(`server/production/intents.go`), not 409. It maps to one authoritative refresh plus the
`intent.conflict` line, with no auto-retry.

**Evidence (cold):**
- `game-ui-intent-outcome.test.ts`: 3/3.
- New browser case: a rejected buy shows "Not enough funds for that."; a revision conflict
  refreshes exactly once.
- Full lanes pass: `make typecheck test-client verify-client-boundary copy-check test-browser`
  (20136 browser tests).

**Severing probes:**
- **S1:** drop the notice assignment, which restores F2. The chromium test fails (1 failed).
- **S2:** make the rejection parser accept extra keys. The unit test fails (1 failed | 2 passed).

**Process note:** the generated copy outputs in this commit were built in a scratch worktree from
this lane's catalog only. The concurrent Typer lane's uncommitted `typer-candidate.json` keys are
therefore not committed here.

## 2026-09-25 — GS0.1 v4 projection (Claude)

Game UI snapshot **v4** (the next free version at landing) is live. It is served by
`get_game_ui_snapshot`.

**What v4 adds:** `features` with six keys, `generators[].provision_cap`, and six `feature.*`
facts. Arms projected: achievements, meters, fiscal and minigames.

**Kernel reuse:** every arm uses exported kernel functions on discarded clones:
- `meters.BandFor`;
- `fiscal.Catalog.Sweep` and `GeneratorLevelCost`;
- `soul.HumanContentLocked`;
- `minigameapi.SupportsTenant`.

No kernel-guarded file changed. The retained `GameUISnapshotV3` validates stored bootstrap
receipts. The client parser decodes every arm exactly and fails closed.

**Compatibility-pin refresh:** `make api-pin`, authorized by accepted GS0.1 rule 4 ("live sync
requires v4"). This is the same class of authorization as GU-C26 for v3, and it is recorded here in
the same change, as `docs/api-foundation.md` requires.

**Blockers and deviations, recorded, not improvised:**
- **B-1, GS5 active play:** `ActivePlayArm.combo.saturated` needs the clamp result of
  `activePlayContributionsWithClamp`, which is unexported in the guarded `server/production/`.
  This lane may not edit guarded paths, and GS0.1 rule 2 forbids re-implementing the math.
  - **Consequence:** `features.active_play` is always `null`. This deviates from rule 1 because the
    opportunities artifact is live. The registered schema is null-only, so producing the arm later
    is an allowed response widening.
  - **Unblock:** export a read-only `production.ProjectActiveCombo`, or equivalent, in a
    kernel-owning lane.
- **B-2, GS4 pets:** out of this batch's scope (it waits on pet-adoption). The arm is null-only,
  with the same widening path.
- **Noted derivation:** `minigames` restates the inline create predicate from the guarded
  `production.StartMinigameAPISession` (a boolean rule, not arithmetic). The unit test pins it, and
  the composed lane asserts `pitch.unlocked == false` before the Fiscal unlock. The Pitch
  unlock-then-create witness already proves the server's side.
- **Noted derivation:** `hoard.preview_ppm` is the published formula
  `min(credit_after, cap_credits) × ppm_per_credit`. `fiscal.HoardFactor` returns only the
  decimal factor.

**Evidence (cold, `-count=1`):**
- `make test-go GO_PACKAGES='./gameui ./account ./gameserver'` passes. This includes 5 new
  `features_test.go` cases against the real pinned epoch bundle, loaded through
  `replaycatalog.Load`.
- Postgres integration via `compose.save-test.yml`: `./gameui` and `./account` pass, and
  `./gameserver` passes when run alone. A first combined run showed bootstrap 500s and SQL deadlocks
  while another lane's test container shared the database; the uncontended rerun is clean.
- Client: `make typecheck test-client test-browser` passes (20160 browser tests).
  `test-game-ui-composed` passes, and its live v4 assertion checks 11 meters, p(doom) = 50, a
  non-empty achievements arm, a Fiscal credit, Pitch locked, `feature.fiscal`, and null
  `active_play`/`pets`.

**Severing probes:**
- **S3:** sweep at the opened coordinate (unswept) makes GS1-A1 fail.
- **S4:** ignoring the Fiscal unlock makes the live-arms test fail.
- **S5:** dropping the Founder v19 gate makes the version test fail.
- **S6:** accepting an undeclared meter band makes the client parser test fail (1 failed | 15
  passed).

## 2026-09-25 — GS1/GS2/GS3/GS6 surfaces + GS7 availability (Claude)

- **Surfaces added:** Trophy Case (GS2), Earnings Calls (GS1) and Reputation Board (GS3) mount
  from registry rows gated by `feature.*` facts. The `minigame_session` row now unlocks on
  `feature.minigame.pitch`, as GS0.4 specifies.
- **Pitch availability:** the Pitch host shows the minigames-arm lock reason before any create.
- **Desk:** provisioned counts with the cap reason, owned-upgrade text, and F10 (a missing cap copy
  withholds the cap and logs an invariant instead of throwing).
- **Recorded deviation:** ID→copy rows live in a new strict `features-presentation.json` rather
  than a presentation-catalog v4 bump. The catalog bump belongs to the Game UI copy-candidate
  compiler lane. The rows are data, sorted, and every key is verified against the copy catalog at
  load.
- **Not done in this batch:** the 320 px reflow measurement (GS1-A5/GS3-A3) is not claimed. The
  Desk's measured 647 px defect is owned by the accessibility RFC.

**Evidence (cold):**
- `garage-surfaces.test.ts` 3/3. It covers the exact phase edges 99/100/199/200 and the
  presentation fail-closed checks.
- `garage-surfaces-browser.test.ts` 21/21 across three browsers:
  - nav from facts;
  - meters text and axe;
  - achievements text states;
  - Fiscal phases, Founder-scoped requests (`expected_revision: 7` against Company 1), the rejection
    and outcome text, and the spend targets;
  - GS6 cap reason;
  - F10;
  - arm-null → Desk;
  - Pitch availability.
- Full lanes pass: `make typecheck build-client test-client verify-client-boundary copy-check
  test-browser` (20193+) and `test-game-ui-composed`.
- **Composed-lane finding, fixed:** the first composed run failed with a Playwright strict-mode
  duplicate, because the availability hint reused the launcher's rejection text. It now uses its
  own `minigame.availability.*` keys.

**Severing probes:**
- **S7:** a Company-scoped harvest fails the Fiscal test.
- **S8:** a cap number with no reason fails GS6.
- **S9:** achievement state with no text fails GS2-A1.
- **S10:** `<=` at `early_ms` fails the phase unit test.
- **S11:** no availability hint fails GS7.

**Still open in this plan:**
- GS5 active play (blocker B-1, guarded export);
- GS0.3 event decoders and announcements;
- the pet slice, which is out of scope.

## 2026-09-25 — GS0.3 announcement decoders (Claude)

**What landed:** `decodeGameUIAnnouncement` adds exact, fail-closed decoders for
`achievement_earned.v1` and `meter_band_changed.v1`. The field domains mirror
`server/save/intent.go validateEventPayload`. The runtime forwards decoded announcements only
inside the existing delivery gate, and a malformed payload throws into the existing resync path.

**How the host announces:**
- One polite chrome `role="status"` region.
- Deduped by `scope\0cursor`.
- An achievement announces its catalog copy.
- A meter band change announces only while the Meters surface is mounted. Otherwise the Meters nav
  button gains the `(changed)` badge until visited.
- Unknown IDs are withheld with an invariant.

**Not in this batch:** the `fiscal_period_harvested.v1` decoder (its nav-badge consumer, OD-3) and
`buff_started.v1` (it waits on GS5/B-1). Both stay open in the plan.

**Evidence (cold):**
- The runtime announcement tests pass 14/14. They include a same-revision republish at a new offset
  and the malformed-payload rejections.
- The garage browser tests pass 24/24 across three browsers.
- Full client lanes and the composed lane pass.

**Severing probes:**
- **S12:** without the host cursor dedupe, the replay re-announces and the test fails.
- **S13, first attempt, survived:** removing a redundant inner `disposition === "deliver"` clause
  changed nothing, because the outer delivery gate already filters, and a consumed offset is
  dropped even earlier. That check could not fail, so the test now republishes the same revision
  at a new offset and the redundant clause is deleted.
- **S13′:** removing the real outer delivery gate fails the test.

## 2026-09-25 — Hand-off for designated cross-party review (Claude)

**Ready for Codex's designated review.** Nothing is self-approved or archived.

This lane's commits:
- `6458bc02`: predeclaration;
- `66ae01f9`: GS0.2;
- `5165abcf`: GS0.1 v4, including the `api-pin` re-baseline under GS0.1;
- `acf43132`: GS1/GS2/GS3/GS6/GS7 surfaces;
- `301728c8`: GS0.3;
- this record commit.

The Terminal Typer lane's commits are interleaved in the same span and are not part of this range.

**Open items:**
- GS5 active play: blocker B-1, a guarded `server/production` export.
- The GS0.3 remainder.
- The GS4 pet slice, which waits on pet-adoption.
- The 320 px reflow measurement.
- Adopting the candidate copy (`copy/catalog/garage-surfaces-candidate.json`).

## 2026-09-25 — Predeclaration: GS5, GS4 care panel, GS0.3 remainder, 320 px (Claude)

**Implemented by:** Claude. Awaiting Codex's designated review. This lane now owns kernel-guarded
paths: HEAD is at kernel `0.3.135`, and every guarded edit bumps the version in the same commit.

1. **GS5 (B-1 unblock):** `production.ProjectActiveCombo`, a read-only export that wraps the existing
   `activePlayContributionsWithClamp` and adds no arithmetic.
   - It feeds a new optional sibling arm, `features.opportunity`. The null-only `active_play` stays
     null, following the `pet_adoption` precedent, because the compatibility gate rejects widening a
     null-only property. The `feature.active_play` fact then follows `opportunity`.
   - `pending` is emitted only while `expires_attended_ms > attended_now_ms`, and expired buffs are
     omitted.
   - The Desk gets an always-present `desk.region.opportunity` with a claim button. It sends no
     synthetic commands and never moves focus.
   - Acceptance rows: GS5-A1/A2/A3/A5, plus A4 on the composed lane if the seconds-scale schedule
     allows it.
2. **GS4 care panel, over the PA7 arm (`features.pet_adoption.pets`):** status band and
   server-eligible actions, with one `care_action` button per eligible action in catalog order. The
   `CosmeticOverlay` is mounted on the live pet (cosmetic-shop G10).
   - **DESIGN-GAP GS4×PA7:** GS4's `PetsArm` wants decayed stats, mood, behaviour, per-action
     cooldowns and `soul_gate`. The later-accepted PA7 forbids projecting raw stats and cooldown
     cursors. Both RFCs are accepted, so no raw care field is added and the panel renders only what
     PA7 projects. The conflict goes to the RFC author.
3. **GS0.3 remainder:** decoders for `fiscal_period_harvested.v1` (the Fiscal nav badge, OD-3) and
   `buff_started.v1` (a Desk announcement).
4. **320 px reflow:** a browser measurement with no horizontal overflow, at the 320 CSS px viewport,
   on the Desk, Fiscal, Meters, Achievements and pet panel.
5. **Docs reconciliation:** the Reputation and Clout surfaces shipped in their own lanes, and
   `docs/game-ui.md` points at them rather than duplicating them.

## 2026-09-25 — GS5, GS4 care panel, GS0.3 remainder, 320 px: record (Claude)

**Implemented by:** Claude. Awaiting Codex's designated review; nothing is self-approved or
archived.

**Range:** `d035b8f8..` this commit, made up of:
- `d035b8f8`: predeclaration;
- `cdb8fe61`: GS5 arm plus kernel export, kernel `0.3.135 → 0.3.136` bumped in the same commit;
- `f32f6175`: Desk region plus composed witness;
- `7a61e4b6`: GS4;
- `6594b646`: decoders and reflow;
- this commit: docs and record.

**GS5:** the arm and the kernel export.
- `production.ProjectActiveCombo` wraps `activePlayContributionsWithClamp` read-only, with no new
  arithmetic.
- The projection lands as the optional `features.opportunity` arm. `active_play` stays null-only,
  because the compatibility gate rejects widening it, as with `pet_adoption`. `api-compat-v1.json`
  is unchanged, so there was no re-pin, and `feature.active_play` follows the new arm.
- **Tests:** Go projector tests (pending only while unexpired, expired buffs omitted, saturation
  from the kernel clamp, null below v18), a client parser that fails closed on expired rows, and
  Postgres for gameui/gameserver/account (compose exit 0; the gameui integration test was confirmed
  to run, not skip).
- **Severing:**
  - O1: pending projected with `>=` → red;
  - O2: saturation forced false → red (the first attempt didn't compile and was redone with
    `&& false`);
  - O3: no buff filter → red;
  - O4: the client parser without its expiry check → red.
- **Desk region:** always present; browser tests A1/A2/A3/A5 pass in three browsers.
  - B1: making the region conditional → red.
  - B2: an auto-claim `$effect` → red.
  - B3: discarding the receipt → red.
  - B4: focusing the claim button on spawn → red.
- **GS5-A4 composed:** real manual clicks until the region projects, then a DOM claim. The effect
  shows in the next snapshot, witnessed as 13, 18, 12 and 6 clicks across the runs.
  - Severing the server projection makes the witness fail with its own message. The first severed
    run instead hit the account limiter (429) at 150 ms per click, so the loop now runs at 4
    clicks/s, bounded at 60 clicks.

**GS4:** the `pet` surface, unlocked by `feature.pets`, which the server now sets true only when
`pet_adoption` holds at least one pet (GS4-A2, Go test). It shows one button per catalog care
action, and ineligible actions are disabled with visible text. Care is Founder-scoped with the
typed rejections, and the cosmetic overlay is mounted on the live pet.
- Severing: P1 greying without text → red; P2 Company revision → red; P3 fact on an empty arm → red
  (Go); P4 no overlay → red.
- **DESIGN-GAP GS4×PA7, for the RFC author:** GS4 specifies decayed stats, mood, behaviour,
  per-action cooldowns and `soul_gate`. The later-accepted PA7 forbids projecting raw care fields.
  Both RFCs are accepted, so nothing raw was added, and the cooldown/soul-lock states surface only
  as typed rejections after a click.

**GS0.3 remainder:** `fiscal_period_harvested.v1`, both arms, and `buff_started.v1`, v1 and v2,
mirror the producer and validator keys exactly.
- Unit severing: D1 dropping the credit-sum check → red; D3 leaving the buff decoder unwired → red.
- Host severing: D2 a badge that never clears → red.

**320 px:** the new reflow test failed first on real overflow: the chrome nav didn't wrap, titlebar
items couldn't shrink, the Desk `.cards` track minimum was `15rem`, and the adoption card inherited
that width. Those are fixed with the `minmax(min(…, 100%), 1fr)` pattern Achievements and Fiscal
already use, and it now passes in three browsers. That pre-fix failure is the check's demonstrated
failing case.

**Cold gates:**
- `make typecheck test-client`: 6905 passed.
- `verify-client-boundary`: 22 component files.
- `copy-check`: 657 keys.
- `test-browser`: 20,940 passed.
- `test-game-ui-composed`: GS5, Pitch and v4 lifecycle all pass.
- `make test-go` for gameui, production and account; Postgres for gameui, gameserver and account.
- `api-check` is clean.
- `verify-kernel-version` still stops at the pre-existing `50a3a514` history item.

**Candidate copy:** the new `desk.opportunity.*`, `desk.buff*`, `cap.active_combo`, `cap.cash` and
`fiscal.nav.harvest_badge` rows, plus `pet.care.*` (companion tone, `PENDING OWNER COPY`), are in
`copy/catalog/garage-surfaces-candidate.json` for Marco to adopt.

## 2026-10-04 — Codex GS0.2/GS1-A6 applied-refresh probe, predeclared

- **Observed composed failure:** replacing the Pitch test's direct Fiscal API
  setup with DOM-only Earnings Calls controls applied harvest, then the
  immediate Pitch unlock returned HTTP 200 `outcome:rejected` with
  `revision_conflict/expected_revision` (`current_revision:4`). The test
  did not retry; the exact player journey stopped before Pitch creation.
- **Question:** GS0.2 says an applied intent awaits an authoritative refresh
  before clearing pending. Does the host actually do so before permitting the
  next Founder action, or can it send the pre-harvest Founder revision?
- **Population:** real Game UI component in the browser with a runtime double
  that applies Fiscal harvest at Founder revision 7, defers the authoritative
  snapshot at revision 8, then observes the Pitch unlock control. Use the
  existing composed browser/Postgres path as the integrated confirmation.
- **Criterion:** the host starts one authoritative refresh, remains pending
  until it resolves, and the subsequent DOM unlock intent carries revision 8.
  The composed path must apply both DOM-issued Fiscal intents and reach the
  existing Pitch terminal receipt. No auto-retry of a rejected intent.
- **Negative controls:** on current code the deferred snapshot is never
  requested and the second intent carries revision 7. In the composed lane,
  sever either the harvest or unlock DOM click; each must fail at that action
  or locked Pitch creation. No direct API mutation may rescue the run.
- **Limit:** this targets GS0.2/GS1-A6/GS7-A7 sequencing; it cannot alone
  approve the full Garage Surfaces or MA implementation ranges.

## 2026-10-04 — Codex RP-147 Fiscal→Pitch player journey correction

- **Review by:** Codex. **Recorded by:** Codex. **Targeted source:** Claude's
  Garage Surfaces GS0.2 host and the MA composed Pitch witness. **Decision:**
  **CHANGES REQUIRED** on applied-intent sequencing and the GS1-A6/GS7-A7
  composed acceptance claim; this is not a range-union verdict for either RFC.
- The first browser-double draft deferred the initial bootstrap snapshot and
  failed at a missing nav button; that invalid fixture was corrected before
  counting it. With bootstrap normal and only the post-harvest snapshot held,
  the test failed on the actual GS0.2 property: zero authoritative refresh
  requests after a 200 applied harvest. The original composed path also failed
  at the following DOM unlock with `revision_conflict/expected_revision`.
- The host now awaits an authoritative refresh before clearing pending for
  applied intents outside its existing terminal-transition special path.
  The browser-double test holds that refresh, verifies the unlock remains
  disabled, then releases revision 8 and observes a DOM unlock intent with
  `expected_revision:8`. Rejected intents remain never auto-retried.
- The composed witness removed its direct `founderIntent` setup and uses the
  rendered Earnings Calls harvest and Pitch-specific unlock controls. It
  asserts both emitted applied intents, then plays Pitch by keyboard to the
  terminal receipt and checks the refreshed Company snapshot/current session.
  Against real Postgres/Vite/gameserver it passes unmodified. Temporarily
  disconnecting the harvest callback failed at `Fiscal harvest for Pitch
  emitted no intent request`; disconnecting the unlock callback failed at its
  `waitForRequest` (no intent). Both callbacks were restored.
- **Cold gates:** `make typecheck` (0 diagnostics), `make test-client` (6,905
  passed), `make verify-client-boundary`, `make copy-check`, `make build-client`,
  and full Chromium/WebKit `make test-browser` (6,984 passed, one skipped per
  browser, performance lane passed) all pass. `make test-game-ui-composed`
  passes after both restorations. The Firefox-inclusive browser gate remains
  unverified on this host because Playwright Firefox aborts before test import.
- **Cross-party gate:** this Codex-authored GS0.2/test correction needs Claude's
  exact-range designated review. Neither party has approved the complete
  Garage Surfaces or MA implementation spans; the historical kernel-history
  CI failure is separate and still red.

## 2026-10-04 — RP-172 intermittent composed Pitch request, bounded diagnostic

The existing DOM-only composed Pitch lane passed several cold runs but also timed out after
`Start a pitch` and after an apparently enabled Fiscal `Unlock for 3` click without seeing the
expected request. At the Fiscal timeout the button was still enabled and `main[aria-busy=false]`.
A reset of the dedicated test DB removed cross-epoch contamination but did not stop the failure.
Four temporary console-instrumented product runs passed, so they cannot establish whether the
failing click reached Svelte or was dropped in `act`; all product instrumentation was restored.

**Next diagnostic (test-only):** install a capture-phase browser event trace for only the Pitch
Start and Fiscal Unlock controls in the composed driver. On a no-request failure, report bounded
`pointerdown`/`pointerup`/`click` observations with button disabled state and `main[aria-busy]`.
The discriminator is whether a browser click actually reached an enabled control: absence or a
mid-click disable is a test/actionability timing boundary; an enabled click event with no request
is a host/handler boundary needing a deterministic browser double before product changes. The
trace must not turn a failed request into a pass, retry the action, or log credentials. Ownership
remains Garage GS0.2/GS1-A6 and Minigame GS7-A7; Cosmetic AC14 approval cannot repair it.

The first traced cold run reproduced the Fiscal timeout. It captured `pointerdown` and
`pointerup` on `Unlock for 3` while the button was **disabled** and `main[aria-busy=true]`, with
no `click` event. Thirty seconds later the page had settled and the button was enabled, but the
request waiter had been started for an action the browser never dispatched. The earlier locked
Pitch Start did have a real enabled `click` event and a response. This narrows RP-172 to a test
actionability race at the Fiscal boundary, not a demonstrated `act` drop. Next correction is
test-only: wait for idle host and the exact rendered control to be enabled, then invoke its DOM
button once in that same browser task; keep the prearmed request/response oracle. Apply the same
one-click discipline to both Pitch Start states. No retry after a click, no timeout increase, no
product code change. A severed callback must still fail with no request.

The first corrected cold run found a second failure before any Pitch Start click: the legitimate
Exit offer preempted `minigame_session`, and the page sat on `offer_sheet` with visible `Decline`.
That was not a lost click; GS0.4 explicitly gives lifecycle preemption precedence. The composed
driver must distinguish this surface transition from actionability. On an offer before Pitch
Start/Unlock, it may submit the visible Decline control, return through the Pitch/Fiscal nav and
then perform the still-unattempted action once. It may not synthesize a backend decline, count the
decline request as Fiscal unlock, or retry a Pitch action whose click was already dispatched.

The test-only correction now performs one DOM click in the same browser task that observes an
idle host and enabled exact Pitch/Fiscal control. The three corrected ordinary cold runs passed;
the prior failing run captured the disabled pointer sequence and no click. A temporary
disconnection of `FiscalSurface`'s Unlock handler then produced an enabled `click` event with no
request and failed the composed oracle at the intended assertion. The handler was restored
byte-exact; the full two-driver `make test-game-ui-composed` target passed. Offer preemption is
handled only before a Pitch action is dispatched via the visible Decline control, with its own
applied receipt required; the run output now counts such declines so that path cannot be claimed
executed merely because the branch exists. The final exact two-driver Make target naturally
reported **one visible offer decline** and still completed Pitch and Cosmetics in 16.6 seconds.
Hosted CI and a deterministic forced-offer negative remain separate. No production source was
retained in this range.

**RP-172 landing coordinate:** `d6295bc3^..d6295bc3` carries the test-only actionability and
offer-preemption correction, executed records, ledger, CI documentation and roadmap reconciliation.
**Review by:** Codex (implementer-side first filter). **Recorded by:** Codex. The disabled-pointer
failure, enabled-click/handler-severing failure, restored production bytes and repeated cold
positive runs are first-filter evidence only. Claude's designated cross-party review must cite
this exact range plus this coordinate record before the Garage/Minigame archival gates can consume
the correction. No production source or owner-authored text changed.

## 2026-10-07 — Original GS0.3 designated review: scope and counterexamples

**Review by:** Codex. **Recorded by:** Codex. Original Claude range:
`301728c8^..301728c8`, all fourteen changed paths inspected, including candidate
copy, generated consumers, orphan list, manifest and the plan/log edges. Current
baseline is `7aab0e2e`; later Fiscal/buff, lifecycle/recovery and Codex corrections
are separate ranges, not absorbed by this review. No full Garage verdict or
archival is authorized by this bounded review.

**Predeclared execution:** run the retained runtime announcement population and
native Chromium/WebKit Garage browser population cold, including consumed-offset
and same-revision/new-offset replay, host duplicate suppression and off-surface
badge clearing. Native Firefox remains an independently open execution boundary.
Compare the two original event decoders with the actual server payload validator.
The browser meter decoder visibly omits the server's direction/value relation
(RP-312); verify with four invalid tuples: up/equal, up/decreasing, down/equal,
down/increasing. Legal increasing/decreasing values, including 0/100 boundaries,
must survive. Exercise the malformed publications through the actual runtime
callback: no announcement; one authoritative-resync notice, socket closure,
cleared position and fetched snapshot, with cleanup of the subscription.

**Repair boundary:** tests first, then only the meter direction predicate under
accepted GS0.3 (same payload domains as `server/save/intent.go`). No wire/schema,
server, balance, copy, receipt, generic cursor or auth-policy change. Verify an
exact compiling omission restores the failures, then restore source byte-exact.
For the original replay witnesses, bypass the actual runtime delivery gate and
host cursor guard separately; record legitimate surviving controls and any
invalid instrument attempt. Do not infer real-server achievement acquisition,
all-engine accessibility, later decoder acceptance, full CI or release proof.
New Codex tests/repair/records require Claude's designated exact-range pass.
