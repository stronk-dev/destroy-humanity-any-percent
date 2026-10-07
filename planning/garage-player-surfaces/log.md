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

### RP-312 test-first result

`make test-client` executed cold: **eight new failures**, 9,682 passes and
340 explicitly skipped browser cases. Four direct decoder assertions accept
the contradictory payload; four socket-path assertions receive an actual
announcement instead of `resync_required`. The four admitted boundary controls
pass, as do both retained original announcement tests. Revision 1 is primed
before revision 2, so a cursor gap cannot produce a vacuous recovery pass.
The regression also checks the real transport-storage key and Founder-state
endpoint, sourced before this executed run. The first `pnpm exec` launch emitted
no output and was stopped (exit130); it is not counted as test execution. Root
Make uses the installed executable and terminates normally. No product bytes
have changed in this test-first checkpoint.

### Reconnect witness refinement (predeclared before execution)

The retained host replay test invokes the listener again; it does not actually
close and recover a socket. Keep that host assertion and add a separate test
using the real runtime's existing fake-socket instrument: emit1006, advance the
unchanged one-second reconnect timer, assert the second socket's saved-position
subscription, recover the same achievement revision at a NEW offset, and
require exactly one announcement, two recovered notices, no resync/fetch and
the new stored position. Always unsubscribe. This satisfies a simulated
reconnect population, not real-WebSocket/server acquisition or GS2-A4. The
actual delivery-gate omission must fail this new replay oracle too.

### Designated verdict — original `301728c8^..301728c8`

**Review by:** Codex (the other party). **Recorded by:** Codex.
**Reviewed range:** `301728c8^..301728c8`, all14paths.
**Verdict: CHANGES REQUIRED — RP-312.** GS0.3 explicitly binds the event-domain
validators to the server. The original and retained meter decoder omit the
direction/value relation in `server/save/intent.go`: up must increase and down
must decrease. Four direct assertions and four actual runtime publications
reproduce acceptance/announcement of server-invalid payloads. This is not a
product/empirical decision and does not require changing the RFC's mechanics.

The original copy source, catalog entries, TS key/param additions, Go key list,
orphan additions and manifest were checked individually. Historical catalog
bytes hash to `sha256:3d6b80b85b5e41db00085d0cc7d4369fd68e6897523325336cebc3c0481159e2`,
matching original hash/TS/manifest mirrors; the four entries exactly match their
candidate source. The Go generator adds only those four keys, no removals or
preamble changes. Candidate copy is not owner adoption. Original plan/log scope
is explicit; the later remainder, GS2-A4 real-server acquisition and whole
Garage ranges are NOT approved or consumed here.

### RP-312 correction and supplemental evidence — implementer-side first filter

**Review by:** Codex (implementer-side, NOT designated). **Recorded by:** Codex.
**Scope:** new span after `7aab0e2e`, including68ddc0c8 predeclaration,
49576162 tests and every later repair/record edge. AcceptedGS0.3 authorizes only
the missing direction predicate; this source is outside the kernel watchlist.
No server, wire, schema, kernel, copy, balance or CI bytes changed. Original
tests/thresholds remain intact; no boxes/statuses/archives/mints promoted.

Executed evidence:

- Test-first full client run:8 new failures/9,682 passes/340 explicit skips.
  First repaired run reached4 later failures: the test's nested header expectation
  incorrectly rejected legitimate `Content-Type`. Corrected to a nested partial
  header assertion, retaining exactly-one fetch/URL/authorization, all prior
  recovery assertions and legal controls. This was an instrument error, not a
  reason to loosen product validation. The corrected positive run passed9,690.
- Compiling removal of ONLY the new predicate:8 failures recur, while four legal
  boundary controls and the retained suite survive. The source was restored
  exactly to SHA256 `851e4b28240a938f8a35ff7b365198a909d09a0b6f145ab930718fc729dd3f6d`.
- Bypass the real delivery condition (retain run-start duplicate guard): original
  generic recovery and achievement republish tests fail2. After adding the
  actual simulated1006→second-socket→saved-position→new-offset recovery witness,
  the same compiling bypass fails3 (24 controls pass). Restore runtime exactly
  to SHA256 `0a8c420eb9968aa76a3e9ece91c918a43e3d8c4530773c8818bce5104c0cb416`.
- Remove host cursor guard: the retained host test fails in Chromium and WebKit
  because replay overwrites the meter announcement with earned achievement copy.
  The selected negative run terminates normally:2 failures/32 selector skips;
  no attachment hang, fake timeout or skipped-failure success. Restore host
  exactly to SHA256 `a471dea6e8be0725ccd762c64f7829c2568fba8ce11a1cad2cacc50a06a8df27`.
- Final source-restored full `make test-client`: **9,691 pass/340 explicit skips**;
  focused runtime27/27. Native `make test-browser` selected Garage population:
  **34/34 across Chromium/WebKit**, plus its isolated performance **1 pass/22
  selector skips**. Default screenshot/failure settings unchanged. Final browser
  output and verdict fully observed; the earlier positive output was clipped in
  passive HTTP diagnostics only and not used as a substituted gate.
- `make typecheck build-client verify-client-boundary copy-check verify-ci-topology`
  exits0: zero TS/Svelte diagnostics,213-module build, unchanged generated copy
 658keys/hash and content manifest; topology13 negative controls discriminate.
  The full multi-commit copy-history check was allowed to finish unshortened;
  no ceiling or bypass was introduced.

The scoped first filter passes. **Claude still owes the complete new span after
7aab0e2e INCLUDING all record edges.** Current primitives/fixtures are not
Firefox, manual AT, real-WebSocket achievement acquisition, full Garage, whole
CI, mint, release or archival proof. ExistingRP-131/RP-307 and all owner/author/
content/platform/review holds remain. Next safe independent work is a separately
bounded designated review of `6594b646^..6594b646`, all5paths (Fiscal/buff/reflow).

### RP-312 exact first-filter coordinate

**Review by:** Codex (self/implementer-side first filter). **Recorded by:** Codex.
**Reviewed range:** `7aab0e2e..b1e903e3`, all three commits and all ten changed
paths: event predicate, runtime regressions, backlog, canonical Game UI docs,
current state, Garage plan/log, execution queue and 1.0 board/log. No edge
commit from that range is omitted. The full delta was inspected; both logs
preserve their entire baseline prefixes, all paths satisfy the predeclared
allowlist, and all probe-only runtime/host/server/kernel/balance/copy/deployment
bytes are unchanged. A fresh focused run from committed `b1e903e3` passes all
27 runtime cases. **First filter passes only; not designated approval.**
Claude's pending designated review must cover `7aab0e2e..b1e903e3` PLUS this
coordinate-record commit and any later touching edges. This entry does not
self-approve its own record edge or authorize archival. Original Claude
`301728c8^..301728c8` remains CHANGES REQUIRED; the correction's independent
verdict is a separate gate. Next: the separately scoped five-path remainder.

## 2026-10-07 — Fiscal/buff/reflow designated review, predeclared

**Review by:** Codex. **Recorded by:** Codex. **Original Claude range:**
`6594b646^..6594b646`, all five paths read completely (host, event decoders,
pet grid CSS and both test files). Baseline `900f409e`. Original GS0.3's RP-312
finding and the new Codex correction remain separate; no full Garage or later
implementation range is absorbed by this review.

Compare both Fiscal union arms and both accepted buff payload shapes to
`server/save/intent.go` and execute the retained decoder assertions cold.
Execute the current Garage native population in Chromium/WebKit, including
Fiscal badge clearing, Desk buff copy and all five mounted surfaces at 320px.
Firefox and manual 400%/AT are not inferred from those two engines or geometry.

Independent compiling counterexamples, one at a time with exact restoration:
omit the automatic credit-sum relation; leave the buff decoder unwired; omit
the Fiscal visit's badge clear; remove chrome nav wrapping. Each must fail its
named existing oracle, while unrelated controls may survive legitimately.
These are temporary review probes, not retained product changes. Do not raise
timeouts, change browser failure/attachment settings, alter copy or redesign
payloads. A record-only verdict is not acceptance of the entire Garage lane.

### Reflow probe correction, predeclared before the replacement run

Removing nav wrapping **survived**: both selected reflow tests and isolated
performance passed. That is not a demonstrated failing case. The retained
shrinkable/wrappable controls can still fit, so the probe did not establish
a violation of the named no-horizontal-overflow property. Nav source restored
exactly. Replace this target with an actual too-wide Desk card track: change
the existing bounded `minmax(min(15rem, 100%), 1fr)` to a fixed `40rem` minimum.
At the unchanged 320px viewport this must produce actual horizontal overflow
and fail the existing oracle. No measurement threshold, viewport, control,
timeout, screenshot setting or assertion changes; restore the track afterward.

### Designated verdict — `6594b646^..6594b646`

**Review by:** Codex (the other party). **Recorded by:** Codex.
**Reviewed range:** `6594b646^..6594b646`, all five changed paths.
**Verdict: APPROVED, bounded batch only.** Both Fiscal arms and the retained
buff v1/v2 shapes preserve the producer fields and server integer/ID/timing
domains. Announcements remain read-only: Fiscal badges while elsewhere,
visiting clears the badge, and a known buff announces on the Desk. Candidate
copy is not adopted by this verdict. The pet edit is a responsive grid-track
change only; it does not authorize the unresolved GS4×PA7 raw-care contract.

Executed independently at the unchanged current product baseline:

- Retained decoder/runtime population: 29/29 passes. Removing automatic
  credit-sum validation fails the contradiction case (one pass/one failure).
  Disconnecting the buff decoder fails both composite decoder cases. These
  are actual assertion failures, not compiler/timeout substitutions.
- Omitting Fiscal badge clearing fails the retained host case in Chromium
  and WebKit (two failures/32 selector skips). Source restored exactly.
- Nav wrapping omission survived, as disclosed above; it is NOT failing
  evidence. The replacement oversized-card probe actually causes Desk
  horizontal overflow and out-of-viewport amount/card nodes; the original
  reflow oracle fails in both browsers (two failures/32 selector skips).
  Neither viewport nor the one-pixel measurement comparison was loosened.
- All temporary source changes restored exactly. The event/host/pet hashes
  match the pre-probe observations, and `git diff --exit-code -- client server
  kernel balance copy deployment` is empty. No product byte remains changed.
- Final cold `make typecheck test-client verify-client-boundary` passes:
  zero TS/Svelte diagnostics; 9,691 tests pass/340 browser cases explicitly
  skipped in Node; boundaries pass. Final native Garage population passes
  34/34 across Chromium/WebKit, with the separate default performance lane
  passing one case/22 selector skips. No failure settings or timeouts changed.

This verdict covers ONLY the five-path original batch. It does not approve
other Garage implementation or record spans, RP-312's new Codex correction,
GS4×PA7, all-engine/400%/manual AT, real-server Fiscal/buff workflows, full
Garage acceptance, kernel history, hosted CI, content adoption or archival.
No boxes or lifecycle statuses flip. The next independently reviewable
producer batch is `cdb8fe61^..cdb8fe61` (GS5 opportunity projection, 12 paths);
its Desk consumer `f32f6175` remains a separate dependent review afterward.
RP-312's required Claude correction range is now bounded exactly to
`7aab0e2e..900f409e`, all four commits including its coordinate edge; this
separate record-only remainder review does not silently extend that product
span. Older Clout spans and all other review/release holds remain independent.

### Review boundary body reconciliation

The mutable board, queue, current state and Garage plan now name the closed
RP-312 span `7aab0e2e..900f409e` in their older active paragraphs too. The later
Fiscal/buff/reflow record-only span is `900f409e..bd028200` (two commits, six
planning paths), with the designated verdict covering the original five-path
`6594b646^..6594b646` only. This record correction changes no verdict, proof,
box, product byte or archival authority. This follow-up record edge remains
explicit; it is not silently relabelled as reviewed implementation. Next safe
work is still the original 12-path GS5 producer review.

## 2026-10-07 — GS5 producer designated review, predeclared

**Review by:** Codex (other party). **Recorded by:** Codex. Original Claude
range: `cdb8fe61^..cdb8fe61`, all twelve paths: Go feature projection/tests,
kernel wrapper, registered schema, generated OpenAPI/TS, client parser/tests,
and all three kernel-version mirrors. Product baseline `87fd23d4`, clean.
The separate Desk consumer `f32f6175`, later records and full Garage acceptance
are not in this verdict's scope.

RP-313 records a confirmed accepted-body conflict before execution. GS0.1
still requires the active-play arm under `active_play`; the implementation
and its log instead use optional `opportunity` because API C2 rejects the
null-to-union type change. API C2 and the v4 re-baseline description do not
silently adopt that sibling shape. Author reconciliation is required; no
accepted body, wire, pin, version, balance or player copy changes here.

Run the existing Go producer and TS parser populations cold plus registered
schema generation/compatibility tests. Supplement only test coverage for
read-only repeated projection, output-pointer isolation, ordered multiple
live buffs, missing artifact and negative clock refusal; these are primitive
proofs, not admitted full saves or an integrated player workflow. Counterexamples:
pending expiry changed to `>=`; actual saturation discarded; expired-buff
filter omitted; client expiry refusal omitted; projected target alias exposed.
Each probe must compile and fail its named assertion; surviving controls and
invalid instruments are recorded. One probe at a time, exact source hash
restoration before the next; no tolerance, timeout, browser, CI or copy changes.
Current cold evidence does not approve historical kernel protocol, later
dependent code, all-engine/manual accessibility, real-server acquisition,+content adoption, archival or release readiness. New Codex tests and records
require their own Claude review, never this designated verdict.

### Supplemental oracle counterexamples, declared before execution

The new primitive tests pass. Independently check their remaining oracles:
increment the source scheduler sequence during projection (read-only failure),
force the active-play fact false (activation failure), and omit the kernel's
negative attended-coordinate refusal (invalid-input failure). These are
temporary compiling probes, with the same exact restoration protocol. The
earlier predeclaration's run-on `,+content` is a formatting typo, not an extra
authorization; its content-adoption exclusion remains in force.

### GS5 producer evidence and designated verdict

**Review by:** Codex (other party). **Recorded by:** Codex.
**Reviewed range:** original Claude `cdb8fe61^..cdb8fe61`, all twelve paths
enumerated above. **Verdict: CHANGES REQUIRED — RP-313 accepted-body conflict.**
The implementation log's compatibility explanation is reasonable technical
evidence, not authority to replace GS0.1's required wire shape. Its author must
reconcile the active-play arm's location, absence/null rules and activation
fact with C2. GS0.1 also literally requires a discarded clone for derivations;
this path directly reads the supplied state through a pure existing helper.
The read-only tests pass, but do not amend that wording. Clarify that boundary
in the same author reconciliation rather than introduce a gratuitous clone or
silently claim one exists. No arithmetic defect was established in this scope.

Executed at the current unchanged product baseline, not represented as running
the entire historical checkout:

- Original three Go producer tests and two composite TS parser tests pass.
  Four new Go test functions add repeated-read/source-immutability, sorted
  multiple-live-buff, output-pointer isolation, four activation profiles and
  nil-state/catalog/negative-clock refusal coverage. They use isolated inputs;
  they do not claim full save admission or player workflow execution.
- Eight compiling temporary faults fail their named assertions: equality at
  pending expiry (one original failure), discarded saturation (one original
  failure), omitted buff expiry (one original failure), pointer alias exposure
  (one new failure), omitted client pending-expiry refusal (one composite
  failure/one valid control), source scheduler increment (two new failures),
  forced-false activation fact (two active profiles plus the original case;
  pre-v18/missing-artifact controls pass), omitted negative-clock refusal (one
  new failure). No compile error substitutes for an assertion failure.
- Every probe was restored exactly. SHA-256: features.go `e344ecee…1f0fa`,
  active_play.go `7816254d…634b`, contracts.ts `f8c126d1…3fe2`, all equal to
  their pre-probe values. Final source diff is empty outside the new test file.
- `make api-check` passes with all generated/pin bytes unchanged. Original
  batch's three mirrors advance together from kernel 0.3.135 to 0.3.136.
  Current kernel-source parity tests pass; this does not waive RP-131 history.
- Host cold selected Go execution disclosed its database skip. The same named
  population then ran through declared Postgres, `-count=1 -v`, and the stored
  schema-v4/Company-v18 rates test visibly PASSed without skipping. That test
  checks rates and transitions, not the opportunity arm; it is not a GS5-A4
  claim proof. Compose reported an existing orphan container; no unrelated
  container cleanup was performed.
- Final full cold gameui/publicapi/account/kernel packages and scoped vet pass.
  Final TS/Svelte check has zero diagnostics; full Node client population
  passes 9,691 with 340 explicit browser skips; UI boundaries pass. No browser,
  whole-CI, current-content acquisition, manual AT or release claim is made.

The new Codex test/predeclaration/record span after `87fd23d4` needs Claude's
separate exact-range review; this verdict does not designate-review our own
supplement. RP-312 and previous Clout spans remain independent. No acceptance
box, archival, content adoption, mint, publication or deployment authorization
changes. Next safe work: separately bounded inspection and execution of Desk
consumer `f32f6175^..f32f6175`, all fourteen paths, as diagnostic review while
RP-313 remains an approval hold. Its DOM and real-server claim proof cannot be
inferred from this producer's green primitive tests.

### GS5 supplement exact first-filter boundary

**Review by:** Codex (implementer-side self first filter).
**Recorded by:** Codex. **Reviewed range:** `87fd23d4..c41432d4`, all three
commits and all nine paths: one new Go test file, backlog, Garage plan/log,
decision/execution queues, current state and 1.0 board/log. All changed content
was inspected; existing log prefixes remain intact and no checkbox flips.
The only retained test/source change is the new isolated Go population; no
product, kernel, balance, generated API, copy, deployment or CI bytes changed.
Committed cold `-run Opportunity -count=1` passes. First filter passes only;
Claude must review this entire span PLUS this coordinate-record edge before
the supplement can be consumed as designated-approved. RP-313 remains open.
The original producer verdict covers only Claude's twelve-path original
commit; the new four-function supplement is not relabelled as independent
evidence review. Next is the separate fourteen-path consumer diagnostic scope.

## 2026-10-07 — GS5 Desk consumer review and corrections, predeclared

**Review by:** Codex (other party). **Recorded by:** Codex. Original Claude
range `f32f6175^..f32f6175`, all fourteen paths, including candidate/generated
copy, presentation data/parser, host/region/receipt adapter, native browser
tests, composed driver, orphan inventory, Go keys and deployment copy-hash.
Baseline `d90aded7`, clean. This is separate from the producer verdict RP-313,
which remains held; no adoption, full Garage approval or archival is inferred.

RP-314/315/316 record source findings immediately. First execute the retained
browser population and the existing composed lane at the unchanged baseline.
Use Chromium/WebKit; Firefox launch failure remains a separate open hold.
Prove RP-315 with a compiling delayed synthetic claim after 1 second: the
existing 400ms test may survive; do not count survival as discrimination.
Then replace it with a real 60-second timer observation. A per-test execution
budget must accommodate that required population; it does not extend an
opportunity's lifetime or change any acceptance bound/CI workflow.

Replace simulated clicking in the keyboard test with native Tab from the
manual button, Enter/Space activation and outgoing intent assertions. Retain
the axe/population checks. Demonstrate refusal using a seeded negative Tab
index/keyboard-prevention fault, one at a time and exact source restoration.

RP-314 test-first: applied buff receipt with `cap_reason_key:cap.active_combo`,
null Lucky delta and non-saturated/empty successor arm must display its reason;
live buffs must expose the projected cap number, not just its label. Repair
only region/receipt presentation under accepted GS5 using existing copy keys
and Amount; no new authored text, arithmetic, epoch, kernel, pin or wire.
The original and any new correction receive distinct verdicts/provenance.

RP-316 is not waived by a green random buff run. Inspect/execute the existing
driver, bind observed result to the branch actually exercised, and predeclare
any later snapshot-credit oracle construction separately. No forged test epoch,
server clock, alternate scheduler, direct gameplay intent or hidden payout fix.
Retained sources must be exact after every temporary fault; no concurrent
mutation while test handles are live. Other gaps become ledger rows, not
unannounced expansion. All new Codex tests/product/record edges need Claude.

### Desk test-first findings, before product correction

Retained baseline: 34 native Garage passes across Chromium/WebKit plus default
isolated performance pass. Existing composed target exits zero: actual GS5
claims `active.click` after 18 DOM manual clicks (zero expired attempts),
verifies its projected buff ID; Fiscal/Pitch/lifecycle/recovery pass, then the
existing cosmetic driver passes. This is NOT Lucky branch coverage or an
all-engine/release-artifact claim. Compose reports existing orphan containers;
no unrelated cleanup performed.

The original 400ms idle test **survives** a compiling 1-second auto-claim in
the actual region (two passes). The replacement native 60-second observation
checks retained intent history throughout and fails on that same fault in
about 1.1 seconds, both engines. Region restored to exact pre-probe SHA
`f9ea2af3…14503`. No product timer fault is retained. Its 70-second test-runner
budget accommodates the specified 60-second population; the acceptance horizon
is still exactly 60 seconds, with no fake timer or display-clock substitution.

New cap/keyboard tests against unchanged product source: six genuine assertion
failures, two admitted Chromium keyboard controls pass, 32 selector skips.
Both engines omit the capped-buff receipt reason and the live combo limit.
WebKit's native Tab skips Claim for Enter and Space; the old forced-focus click
masked this. RP-315 now records that actual product defect as well. Narrow
repair: explicit zero Tab index on Claim, existing keys/Amount for combo number
and past buff receipt reason. No positive index, forced focus, browser setting,
keyboard API, timing policy, new player prose, schema or kernel change.

### Corrected Desk baseline and counterexamples, predeclared

The corrected native population passes 40/40 across Chromium/WebKit, including
both actual 60-second observations (suite 63.93s), native Tab/Enter/Space,
cap receipt/number, existing axe/reflow, plus default performance. Full client
9,691/343 explicit browser skips, TS/Svelte zero, build/boundaries pass. Original
copy delta was mechanically inspected: 41 additions, zero existing-row changes,
the Go key set agrees with no removals/non-key structure change, and all 115
candidate rows match generated rows. Original type/file/manifest hash mirrors
agree at `3a890004…46860`; no candidate adoption is inferred. Current complete
copy/history/manifest gate passes at 658 keys/611 orphan warnings; topology's
13 negatives pass. The independent real-server rerun claims `active.production`
after 17 manual clicks, zero expiries, and both drivers exit zero. Neither run
exercises Lucky. A read-only process-inspection attempt was denied by sandbox;
the retained live command handle was polled normally to terminal completion.

RP-316 executable isolated probe evaluates the unchanged actual
`witnessOpportunityClaim` function body (no duplicate reimplementation), with
a Lucky receipt and a **null** next snapshot. It returns success after both
page observations. This confirms the oracle gap, not an actual server payout
defect or integration proof. Driver source remains exact at `c6c52404…6c84a`.

Before final correction closeout, independently sever each new presentation
path: suppress buff-receipt reason, omit only the cap Amount while retaining
its label, and make Claim's Tab index negative. Selected native tests must
fail, admitted controls may survive. One compiling probe at a time, corrected
region SHA `5ee8223a…e6516` restored after each. No screenshot/timeout/viewport/
browser setting or assertion changes. Final complete native run remains the
actual required minute, never the old 400ms shortcut.

### Desk consumer designated verdict and narrow correction checkpoint

**Review by:** Codex (other party, original Claude changes only).
**Recorded by:** Codex. **Reviewed range:** `f32f6175^..f32f6175`, every one
of its fourteen paths: generated catalog/hash/types, GameUIApp,
OpportunityRegion, presentation JSON/parser, opportunity receipt adapter,
Garage browser tests, composed driver, Garage candidate copy, orphan inventory,
deployment manifest and generated Go keys. **Verdict: CHANGES REQUIRED**,
RP-314/315/316. RP-313 producer-body hold remains separate. Source and all
generated deltas were inspected; original 41-key additions agree mechanically,
with zero existing catalog changes/Go removals and all 115 candidate rows
matching generated rows. This is not owner adoption of copy or pet placeholders.

Executed baseline and failing-first evidence appears above. The original idle
instrument survives a delayed command; the replacement catches it. Original
Lucky acceptance survives a null next snapshot in the isolated actual-function
probe. Two real composed runs instead exercise admitted buff arms, not Lucky.
The original keyboard instrument masks an actual WebKit focus defect. Cap
receipt evidence and numeric hardcap were missing in both engines.

After the narrow region correction, independent compiling faults produce:
receipt-reason suppression: 2 failures / 6 selected passes; cap Amount omission
with label retained: 2 failures / 6 passes; negative Claim Tab index: 4 failures
/ 4 passes. Each has 32 explicit selector skips and exits nonzero on the
intended assertions, not compilation/timeout. All faults removed, corrected
region SHA `5ee8223a531430fb1256e53f73a2a05e15020ccfef81dd8d7d4a70f2b18e6516`.
Unchanged composed driver SHA `c6c524046f4aa97649095f2f894e9da2a3ce26ca173cf545b08e6d2197a6c84a`.
Final full native run: 40/40 Chromium/WebKit, 63.75s, both actual-minute cases;
default isolated performance 1 pass/22 explicit skips. Earlier corrected-source
client/types/build/boundaries and full copy/manifest/topology gates pass as
recorded above. Nothing claims hosted CI, Firefox, AT or full Garage acceptance.

**Review by:** Codex (implementer, self first-filter ONLY).
**Recorded by:** Codex. New correction span starts after `d90aded7` and includes
`8a5d0def`, `d4798d77`, this region/docs/record change and every later tracking
edge through its exact final coordinate. Local first-filter approves the
bounded receipt/cap/keyboard/timer correction, NOT RP-316's unchanged oracle.
Claude must inspect the full range, including records; this original-code
verdict cannot serve as independent review of Codex's correction. No box,
acceptance/lifecycle/status, content/hash mint, archive, push or deployment.

RP-317 is additionally filed from source: the accepted unknown-opportunity
pair loses its invariant flag in the surface mapping and the host ignores
invariant notices. This is separate next test-first work, not an unannounced
host change in the current region repair. Safe continuation: predeclare
RP-316's snapshot proof and/or RP-317's diagnostic tests under accepted GS5.

### Exact Desk correction boundary for the other-party pass

**Review by:** Codex (self first-filter only). **Recorded by:** Codex.
First-filter range `d90aded7..97934966`: all four commits, all ten paths,
predeclaration, retained tests, region/docs repair, ledger and full tracking
reconciliation inspected. **Locally approved, bounded correction only**;
RP-316/RP-317 remain separately queued. Claude's designated pass must cover
`d90aded7..97934966` **plus this coordinate commit itself** (its only path is
this append-only log); approval of implementation alone would omit record
edges. The producer supplement remains exact `87fd23d4..d90aded7`, separately.
No acceptance box/status/archive/mint/publication or whole-CI claim.

### RP-317 accepted invariant reporting — predeclaration

Baseline `0f3a1a7a`, clean; separate from the exact Desk correction range
`d90aded7..0f3a1a7a`. Authority: accepted GS5 unknown-opportunity rejection
row plus GS0.2's unlisted-200/invalid-400 invariant rules. No owner decision
or authored-copy edit needed; RP-313 wire-body hold remains unmodified.

Test first: the actual shared mapper with the actual opportunity mapping must
retain the not-pending copy and set the unknown-ID invariant, while expired
and not-pending refusals remain ordinary. Native mounted-host cases must emit
exactly one diagnostic for unknown opportunity, unlisted rejection and typed
400 invalid, with the existing appropriate player status and no duplicate
intent. Expired/not-pending controls emit zero diagnostics. No mechanical
pair appears in player status. Both native engines; no fixture clock required.

Minimal proposed seam: explicit optional surface invariant-pair set, keeping
existing copy maps and callers compatible; GS5 declares its single pair.
Host consumes the existing notice flag once in outcome/error branches, using
fixed diagnostic strings with no request, token, ID or player-data payload.
No kernel, schema, numeric, epoch, copy/hash, timing, retry or offline policy
change. Canonical docs accompany host behavior. Independently omit the GS5
flag and host outcome/error reporting in separate compiling probes; their
named tests must fail. Restore exact source after each, no live-handle edits.
Full client/types/build/boundaries plus focused native population and existing
composed target rerun. All new Codex predeclaration/tests/product/docs/records
after `0f3a1a7a` need Claude; no self-archive or acceptance box flip.

### RP-317 failing-first evidence

Unchanged production: full client population has 1 mapper failure, 9,691
passes/348 explicit browser skips. Native single-engine reruns each terminate
nonzero with exactly 3 diagnostic-count failures / 2 ordinary-refusal controls
passing / 20 selector skips. Missing unknown/unlisted/invalid diagnostics
reproduce in Chromium and WebKit; no compiler/timeout failure is counted.

The earlier combined-engine invocation stalled after reporting Chromium's
failures; its incomplete WebKit result is discarded. Ctrl-C/TERM did not
finish. Read-only lsof identified this run's exact listener PID 50228 on
51204; sandbox denied signal delivery, narrowly escalated TERM then KILL
stopped only that owned Vitest process. Single-engine reruns above need no
timeout/screenshot/failure-capture changes. A test-only era spelling was
corrected to the existing `era_1995` type before reruns; not a product defect.

Implementation seam refined without expanding scope: the optional surface
invariant-pair set lives as metadata on the existing readonly copy Map,
instead of adding a new act option/caller parameter. Ordinary copy maps stay
compatible; GS5 owns its one pair. The unchanged actual mapper and host are
still what the failing tests exercise. No product bytes changed yet.

### RP-317 corrected baseline and independent refusal proofs

The surface owns the single exact invariant pair as optional metadata on its
existing copy map. Shared maps/callers stay compatible. Host consumes the
existing flag once in outcome/error branches, using fixed console messages,
no request/token/ID payload. No server, kernel, wire, copy/hash, clock, retry
or offline behavior change. Canonical docs accompany the behavior.

Corrected types/Svelte: 0 errors/warnings. Full client: 9,692 passes/348
explicit browser skips. Build 213 modules and existing shell/UI boundaries
pass. Focused native baseline 10 passes/40 skips; full final Garage native
50/50 in Chromium/WebKit, 63.98s including actual-minute idle. Default
performance 1 pass/22 skips. Actual existing composed target exits zero:
GS5 click buff after 8 manual clicks, zero expiries; Fiscal/Pitch revision22,
both terminals/continuation/real WebSocket recovery, then cosmetic driver
8.197s/67 requests pass. Not Lucky, Firefox/AT, whole CI or release evidence.

Independent compiling fault populations: omit only GS5's invariant pair:
2 unknown-opportunity failures/8 controls pass; omit host outcome reporting:
4 unknown/unlisted failures/6 controls pass; omit host error reporting:
2 invalid failures/8 controls pass. Each has 40 selector skips, exits nonzero
on count assertions, and has unchanged tests/failure capture/timeouts. All
temporary faults removed. Restored SHA256: mapper `f4533d21…2bec6`, adapter
`ca91c5de…d2af`, host `cea53999…635c`; exact values observed before and after.
No mutation or commit was made under a relevant live check handle.

**Review by:** Codex (implementer, self first-filter only).
**Recorded by:** Codex. Local verdict **APPROVED, bounded correction**, spanning
predeclaration `090c1d61`, test-first `c990329d`, current product/docs/ledger/log
and all forthcoming tracking/coordinate edges after `0f3a1a7a`. Claude must
review that entire span separately; this is not the designated verdict.
Original Desk review remains CHANGES REQUIRED, RP-316 unchanged Lucky oracle
and RP-313 author-body hold remain open. Previous Desk correction is exact
`d90aded7..0f3a1a7a` and producer `87fd23d4..d90aded7`, independent obligations.
No acceptance checkbox, archive, content adoption/mint, push or deployment.

### Exact diagnostic correction boundary for the other-party pass

**Review by:** Codex (self first-filter only). **Recorded by:** Codex.
Locally approved `0f3a1a7a..092ef3eb`, all three commits: predeclaration,
test-first unit/native cases, three product seams, canonical docs and
ledger/log inspected. Claude must review **that range plus this tracking
commit itself**, covering its Garage log/plan, current state, execution queue,
roadmap board and append-only roadmap log. Together they are the complete
diagnostic correction span; this self first-filter is NOT designated review.
Original consumer remains CHANGES REQUIRED; RP-316 and RP-313 remain open.
Prior independent exact ranges are not absorbed or reset. No acceptance box,
status/archive/mint/publication or full CI/release claim.

### RP-316 authoritative claim-effect oracle — predeclaration

Baseline `c0eb3dc5`, clean. Previous turn was progress: accepted presentation
and diagnostic corrections plus executed counterexamples, not a wait or
impasse. Authority for this separate test/tooling lane: accepted GS5-A4's
DOM-only natural claim and authoritative next-snapshot effect. RP-313's
accepted wire-body conflict remains author-owned; no shape/pin rewrite here.

First mechanically lift the driver's existing receipt/buff/Lucky predicate
into one pure tooling oracle, imported by both the real driver and tests.
No duplicate test-side success predicate. Browser-safe .mjs tests can run in
the existing Node/native collectors, with no new exclusion or CI workflow.
Test-first reject Lucky with null/missing/stale/unrelated successor, missing
or mismatching cash, malformed delta; wrong request/receipt/revision/run
binding; unknown effect and a buff ID attached to the wrong effect. Existing
matching buff/Lucky and saturated zero-credit controls must remain admitted.
Fixtures prove the oracle only, not producer payout or server integration.

Correct only the acceptance observer: require the applied receipt's exact
intent/opportunity/revision/run coordinates; require the same post-command
Game UI revision and run (reads do not persist accrual); Lucky's canonical
nonnegative credited delta and receipt snapshot cash must match the unique
cash row in the actual successor; buff ID/effect must match a unique live row.
Use string identity, not a second payout formula or numeric tolerance. The
receipt's canonical cash includes lazy accrual, so comparing an earlier
pre-command bank plus Lucky alone would be false. Fail on unsuccessful read.
Expose actual branch/revision in the run diagnostic, never call a buff run
Lucky evidence. Single source oracle; canonical docs in the same change.

Preserve 60 DOM attempts, 250ms manual cadence, three-expiry fail-closed guard,
current epoch and real Postgres/WebSocket path. No extra synthetic gameplay
intent, server clock, seed/epoch surgery, balance/kernel/schema/copy/mint or
product change. One actual composed rerun must terminate and state its branch;
do not loop accounts until a desired branch appears. Demonstrate compiling
Lucky receipt-only, cash-mismatch and stale-revision acceptance faults in
separate probes, plus a buff-effect disconnect. Restore exact source after
each; no live-handle edit or weakened failure setting. Full client/types/
build/boundaries and targeted native oracle cases. Claude must review the
complete new span after `c0eb3dc5`, predeclaration through tracking edge;
no self-designated approval, boxes/status/archival/publication or full CI claim.

### RP-316 lifted baseline, before any predicate repair

The old receipt guard/Lucky-string/buff-ID predicate is mechanically lifted
into `client/tools/opportunity-claim-proof.mjs`; real composed caller imports
it. There is one predicate, not a competing test implementation. Natural
driver cadence, guards, input and reads are unchanged. New pure .mjs tests
remain included in the existing Node/browser collectors; no config exclusion.

Full client at the unchanged predicate: 21 intended assertion failures,
9,700 passes/348 explicit browser skips. The new 29-case population has
5 admitted effect/zero controls and 3 already-discriminating negative controls;
21 missing-state/cash/coordinate/effect negatives survive wrongly. This is
oracle discrimination, not a proven producer defect or integration evidence.
The saturated-zero fixture is explicit, not permission to require every
credit to be positive. No payout arithmetic will be copied into this checker.

Canonical-value construction refinement: inject the existing production
`parseCanonical` into the pure oracle. Vitest imports that same TS module;
the existing Vite server's SSR loader supplies it to the Node composed driver.
No copied regex/precision/exponent constants or alternative Decimal parser.
Only canonicality/nonnegativity and exact serialized cash identity are checked;
no accrual/payout formula, rounding or tolerance is implemented here.

### RP-316 first corrected baseline and a stronger credit refusal

First corrected population: all 29 oracle cases pass; full client 9,721/348
skips, types/Svelte zero, build/boundaries pass. Native oracle 58/58 plus
default performance pass. Actual composed rerun binds `active.production`
to revision31 after20 DOM manual clicks, zero expiries; other composed and
cosmetic lanes pass. This is buff integration, NOT Lucky acquisition proof.

Before closeout, test an unsupported positive credit (absent/mismatching cash
change, insufficient delta, no balance growth) and a falsely capped zero
(missing cap, below-cap bank, absent saturation). Current cash identity alone
must not count these as credited outcomes. This stays inside the predeclared
canonical credit/state proof, not a payout recomputation. Positive credit
requires the unique receipt cash-change after value to equal both snapshots,
with a canonical total delta at least the credited delta and actual bank
growth. Claim only credits cash; its lazy cash accrual is nonnegative in the
T0 population. Zero credit stays admitted only with the receipt saturation
and cap reason plus actual successor bank at its published cap. All values
use the same production Decimal parser/comparisons, no subtraction, formula,
precision constant, tolerance or new threshold. Add the negatives before
strengthening the oracle; disclose this initial incomplete instrument.

The seven added unsupported-credit/false-zero cases fail against the first
corrected checker (9,721 other passes/348 skips); all previous 29 cases still
pass. Strengthening follows those actual failures. A separate admitted control
with total cash delta30 and credited delta20 prevents demanding that total
delta equal Lucky alone, since lazy accrual is included. No producer/payout
code or acceptance bound changes.

The five independently compiling predicate probes terminate nonzero in both
engines: old Lucky shortcut 50 failures/24 passes; cash-identity omission
2/72; exact successor-revision omission 4/70; buff-effect disconnection 2/72;
capped-zero guard omission 6/68. Each restores exact checker SHA256
`0954dcc0e0f472ff49e37485b77e291d6324657c86dfc0d87bcf1fa0a01cd0e6`
before the next fault. No test/failure capture/timeout edits during probes.

The successor HTTP check also needs an executable counterexample. Move that
check into the same pure tooling module: page read returns actual status/body,
the real caller and tests consume the one status predicate. Registered 200
control; 0/201/401/500/503 negatives, even with snapshot-shaped data. No JSON
error body or unsuccessful read can count as a successful next snapshot.
No browser-config exclusion, public API/status change or gameplay bypass.

### RP-316 final cold verification and first-filter disposition — 2026-10-07

The sixth independent compiling fault disables only the successor status
predicate: 10 assertion failures/76 passes across Chromium/WebKit, including
all five unsuccessful statuses in each engine. Restored helper SHA256
`6b7b73f152bffee7ca2bd82cd69475a58540244187819108b1b5de21c1d0acd0`;
driver `392dba99bb9b023ebb2e162cdfc40d588ba0b44897d29ceb10fc17ff92b35749`.
No probe remains. Final native population 86/86, isolated performance1 with
22 explicit skips. Final types/Svelte zero errors/warnings, full client9,735
passes/348 browser skips, build213 modules, boundary14 shell/8 UI/22 Game UI.
No hosted/whole-CI claim from these selected local lanes.

Final `make test-game-ui-composed` terminates exit0 on the unmodified checker:
15 DOM manual clicks, `active.click`, zero expired attempts, exact successor
revision26. Fiscal Pitch5 DOM commands, credited1e0 at revision29; v4
transitions/both terminal states/next-run/real WebSocket recovery pass.
Cosmetic composed passes in9.036s with69 observed requests/no violation.
The earlier first-corrected run was production buff; this final one is click
buff. Neither is Lucky acquisition/integration evidence. No repeated-account
selection for Lucky or fabricated clock/epoch/gameplay command.

Review by: Codex (implementer first-filter only). Recorded by: Codex.
Scope: complete new change after `c0eb3dc5`, including predeclaration
`00fb7366`, mechanical test-first lift `cdd00ee5`, this oracle/docs/ledger
repair and forthcoming tracking edge. Local first filter passes narrowly:
one actual tooling predicate shared by the driver/tests; production numeric
parser, no copied payout arithmetic; strict request/revision/run/cash/buff
binding; guards and natural population preserved. No product runtime,
kernel/schema/balance/copy/workflow changes. RP-316 is locally corrected,
NOT designated-approved or closed. Claude must inspect the entire exact
new range including the final records before any acceptance/archival claim.
RP-313's author-body conflict and all other holds remain independent.

Next safe work: predeclare designated diagnostic review of original Claude
pet consumer `7a61e4b6^..7a61e4b6`, all paths. Existing RP-132/GS4×PA7
author-wire hold remains; inspecting/measuring does not resolve or bypass it.

## 2026-10-07 — Predeclare original Claude pet-care diagnostic review

Review by: Codex (designated other party). Recorded by: Codex.
Exact original implementation range: `7a61e4b6^..7a61e4b6`, all eight paths,
not subsequent CSS/diagnostic/planning edges. Review all line changes against
accepted GS4, PA7/PA8.5 and Cosmetic §7.4. The original slice explicitly
limits itself to the existing adoption arm, identity/status/eligible action
IDs and cosmetic overlay; it does not implement raw stats/cooldown/mood.
Existing RP-132 and GS4×PA7 author-body conflicts stay binding; no technical
pass is an author ruling, full GS4 acceptance or approval of the wire shape.

Execute current two original pet browser cases in Chromium/WebKit and the
original Go adopted-pet fact case cold. These are fixture/primitive checks,
not real-server care receipt or GS4-A5. Independently sever the care callback,
Founder-scoped revision, unavailable-action text, overlay and empty-map fact.
Each compiling probe runs serially with no live-handle edit and exact source
restoration before the next. A surviving probe is recorded, never hidden or
turned into success. No copied player content, permanent product change,
schema/kernel/mint/workflow/status/archive/push or acceptance-box flips.
Inspect known missing GS4 states/announcements separately from this narrow
slice; do not claim existing population covers those absent branches.

## 2026-10-07 — Designated original pet slice verdict

Review by: Codex. Recorded by: Codex. Exact range:
`7a61e4b6^..7a61e4b6`, ALL eight paths, full line-content diff inspected.
Verdict: **CHANGES REQUIRED for the existing accepted-contract hold**;
mechanical subset verified, not full GS4 acceptance. GS4 demands its raw
PetsArm/mood/cooldown/soul states; PA7 explicitly forbids those internals and
requires another wire shape (RP-132). The log's disclosed GS4×PA7 gap is
honest but not author reconciliation. Keep current privacy/compatibility
boundaries; the author must reconcile both bodies before contract approval.
No new payout/eligibility defect inferred from this conflict.

Executed current original pet browser cases:4 passes across Chromium/WebKit,
46 other cases explicitly unselected, plus isolated performance1/22 skips;
cold original Go adopted-pet fact passes (`-count=1`). The four separately
compiling care callback/Company-revision/unavailable-text/overlay faults each
fail one pet case per engine (2 fail/2 pass/46 unselected). Expected revision
is7, not1; disconnected callback emits no request. Empty-map fact fault fails
Go with `feature.pets must be false with no adopted pet`. All restored before
the next probe and final rerun; final browser4/performance1/Go pass.
Restored SHA256: PetCareSurface
`54fdf3350c11213e1c92f65227737fde3952b3d0d4f1f10bfce992973f37487a`,
GameUIApp `cea53999751ab5fc6dfca79ad5d6cadd6effb11988035135fb015d03db34635c`,
server/features `e344ecee553f1260e1a33802985d98c45970ff6035cd677ebf4548425f61f0fa`.
No source mutation remains. Current PetCare CSS differs from the original only
by the separately reviewed `6594b646` reflow fix; no absorption of its range.

The browser population proves catalog action order, text for ineligibility,
feed intent/Founder revision, applied notice and one cooldown rejection;
it does NOT prove the other three rejection rows, native keyboard, updated
server band, pending/reconnect behavior or real-server care/refresh. The Go
fixture has a pinned species bundle, not a fresh real deployed Founder.
GS4-A5 and owner-copy/mint/wire holds stay explicit. Later docs/records are
separate review ranges, no self-archive or full Garage/CI/release promotion.

### Predeclare separate remaining status-event diagnostic (RP-318)

Accepted GS0.3 explicitly includes `pet_status_changed.v1`; GS4 wants a polite
announcement only while mounted. Source inspection finds no decoder member or
host branch; the specified announcement key is absent from the current catalog.
This is NOT attributed to the original eight-path care slice, which did not
claim to implement this decoder. Execute the current decoder through the
existing Vite SSR loader, all12 distinct transitions of the server's four
valid bands and invalid UUID/same/unknown-band/extra/missing-field controls.
Report actual undefined/throw results; no fixture gets relabelled integration.
No file/product/test-collector/CI or owner-copy change. Author copy adoption
and raw-care conflict are not inferred from this diagnostic.

Actual unchanged decoder diagnostic terminates exit0: all12 valid distinct
band transitions return undefined; all seven invalid UUID/same-band/unknown
source-or-destination/extra/missing-source/missing-destination controls also
return undefined, never throw. There is no registered pet branch to trigger
the existing resync path. Node/Vite exits normally with the server closed;
no source/test/build/generated artifact is changed. RP-318 remains open.

Next bounded accepted work: test-only care consumer supplement under GS0.2,
GS0.5, GS4's exact refusal pairs/Founder scope and PA7's existing public arm:
actual native keyboard, all ordinary care refusals plus unknown invariants,
pending/reconnect disabling, and applied receipt followed by refreshed public
band/eligibility and next Founder revision. Preserve runtime fixture labels;
not SQL/real acquisition/GS4-A5 proof. No raw care field or owner copy added.
Predeclare and test these cases before any narrowly authorized code repair;
RP-132/GS4×PA7 and RP-318 copy hold remain, no lifecycle promotion.

## 2026-10-07 — Predeclare care-consumer supplement and narrow UI repairs

Baseline `c7d8f815`, clean. Accepted authority GS0.2/GS0.5/GS0.6/GS0.8,
GS4 refusal/Founder/keyboard/refresh requirements, PA7 public identity/band/
eligible actions and PA8.5 sprite. Existing author-wire/raw-care/copy holds
remain; no acceptance of a different end state. No player prose is authored.

First add native Chromium/WebKit cases for all four exact ordinary care
refusals and both unknown-ID invariants; Tab plus Enter/Space; three actual
recovering/resync/restart messages with disabled controls and existing reason
text; missing Founder revision; held intent AND separately held authoritative
refresh with one request, pending text, aria-disabled, native focus retained;
and successive public band/eligibility refreshes bound to Founder revisions.
These are labelled runtime-double checks, not real-server band acquisition.
Source findings RP-319 (pending removes focusability/no text) and RP-320
(restart bypasses care readiness) go into the ledger before measurement.

Repair only confirmed care presentation/binding defects under those accepted
rules: no server/kernel/math/schema/epoch/copy-content/protocol change.
Preserve native disabling for actual ineligibility and reconnect; pending
eligible controls stay focusable but their callback refuses action. Reuse
`common.pending`; demonstrate independent compiling omission faults against
the new native population, restore each exact source before another edit.

Also extend the existing built-client Cosmetic G10 composed driver with one
real care action after its real DOM adoption/equip path. Reuse its declared
test-only pinned bundle, named Postgres and built client/real WebSocket; do not
seed pet state, warp care time, synthesize gameplay fetches or change the
existing T1 cash-only setup. DOM care response must bind request/pet/action/
Founder revision and positive actual care; next persisted public arm and
visible panel must match receipt band and remove that action from eligibility.
Reload preserves it. Do not relabel this unminted setup as release-artifact
proof or require a status-band crossing that natural initial care may not
produce. RP-318 event/copy remains separate. The care state comes from the
actual service, never a fabricated receipt. Sever the actual UI care callback
and demonstrate real composed failure before restoring. Full client/types/
build/boundaries, targeted native, existing composed lanes. No silent skips,
new CI exclusion/workflow, all-Garage/hosted-CI/1.0 promotion or self-archive.
Claude must review the entire new span after `c7d8f815`, including all records.

Initial instrument run is invalid: new tests omitted the mandatory copy params
and guessed action keys without `.title`;28 harness errors prove no product
finding. Mechanical test-only correction resolves registered keys through
the actual copy resolver with `{}`/era1995; no text copied or rewritten.
Next terminal baseline:16 failing assertions/12 passes/50 unselected across
both engines. Actual pending native-disable/focus failures, WebKit Tab skip,
restart enabled and recovering missing explanation appear. Refusal mapping
and public-state/revision controls otherwise pass. The recovering assertion
initially used the Settings offline sentence; correct it to GS0.5's actual
`common.stale_note`, not a new requirement. RP-321 records the missing care
stale label. Same populated surface lacks that key, so no claim rests on the
wrong sentence. Before repair extend public-refresh case to focus the care
control and assert GS0.6 heading fallback after eligibility disables it.
No product edit yet. If that fails, repair the existing accepted fallback
with a focusable heading and guarded pre-update focus handoff only when the
currently focused care control becomes disabled; no spawn/reconnect focus
steal or global hotkey. Continue within the declared care-only UI boundary.

Final unmodified-product baseline after instrument corrections:18 actual
assertion failures/10 positive controls/50 unselected in the two-engine
population. Both public-refresh cases now fail with focus on body instead
of the care heading. Test-first cases land before the narrow product repair.

First narrow correction passes27/28 native assertions; only WebKit's disabled
focused-action fallback remains red. Its focus-loss timing can leave the
disabled action active through `tick()` before falling to body. Accept either
the captured action itself or body at handoff, never another newly selected
control; no wait/bound relaxation. Corrected14-case population passes28/28.
Add explicit no-focus-steal control and isolated pending component activation
case, because the host's same-kind single-flight defense could mask a missing
component pending guard. This strengthens the predeclared properties rather
than treating the host queue as evidence of an unexecuted callback guard.

First actual composed extension passes: same test-only bundle and existing
T1 cash setup; real DOM adoption/equip, then care.feed with positive actual
receipt, public normal band and feed ineligible at Founder revision5; reload
preserves it. No pet-state/clock/receipt surgery or status-crossing claim.
Existing Game UI/Fiscal/Pitch/recovery and Cosmetics lanes pass. RealSocket/
Postgres alone is not release-manifest or author-contract acceptance.

## 2026-10-07 — Care supplement: final local proof and first filter

Review by: Codex (implementer/self first-filter ONLY).
Recorded by: Codex. Scope: entire new span after `c7d8f815`, including
predeclaration `970d47a5`, failing-first `0784f126`, repair/tests/driver/docs/
ledger and subsequent tracking edge. Claude's designated pass is still owed.
No checkbox, RFC status, author contract or archival gate is advanced.

The final16-case care supplement passes32 Chromium/WebKit assertions. The
component-only pending test avoids the host single-flight masking its guard;
the no-focus-steal control requires the player's newly selected nav control
to remain focused. Eight independent compiling omission/fault probes against
that same population discriminate (fail/pass, with50 other tests unselected):

- omit care transport readiness:2/30;
- omit component pending callback guard:2/30;
- omit stale note:2/30;
- omit eligibility-change heading handoff:2/30;
- omit pending text:2/30;
- remove native Tab entry with tabindex=-1:4/28;
- skip only care's authoritative refresh:6/26;
- force unconditional heading focus:8/24, including no-focus-steal controls.

All probes restored exactly. Final SHA256: PetCareSurface
`4e5b731df9e459f8c3d0d118908ae781d0c247932c590e7843e5c9a00f381bd1`,
GameUIApp `d49738f1be6ac2a89230487181e641ecd4e530f48bdcaf456baa0dd26d7e8e7c`,
Cosmetic composed driver
`6816b3496e4b34bde76faeda0620404ee3bfdedb489ed79f8df045c0b55b29a5`.

Actual composed adversary: sever the DOM care callback only; the existing
main Game UI/Fiscal/Pitch/WebSocket populations still pass, then care fails
with `Garage care DOM callback emitted no care intent` (existing30s response
guard, root Make exit2). No server mutation or acceptance-bound relaxation.
Restore callback exactly, rebuild, then run the unmodified root composed
lane again: real DOM adoption → DOM feed → positive bound receipt/public
normal band/feed ineligible at Founder revision5 → reload retention passes.
It does NOT prove a band crossing, status announcement, raw-care contract,
minted release bundle or clean-host deployment. Final main lane also passes:
17 DOM clicks, active.click buff at revision28/zero expired attempts; Fiscal
five Pitch commands/credit1e0 at revision31; v4 transitions/both terminal
states/next-run/WebSocket recovery. Cosmetic6.457s/72 audited requests, no N5
violation. Natural opportunities are not a Lucky integration claim.

Final immutable-source gates: full Garage file82 assertions across Chromium/
WebKit, including actual-minute idle Desk; isolated performance1 pass/22
explicitly unselected. Types/Svelte zero errors/warnings; client9,735 passes/
364 explicit browser skips (10,099 total),105 files pass/22 skip; build213
modules; shell/UI/GameUI boundary counts14/8/22. Complete copy pipeline658
keys and content manifest pass; topology and13 seeded negatives pass. These
are existing root lanes, including the unchanged CI composed job; no workflow
or collector exclusion was added. Not an executed hosted-CI verdict.

Firefox retry: same existing root native lane,16 care cases selected; browser
session connection timeout at60.03s before import, zero tests/one unhandled
error. Teardown did not terminate, so this failed handle was stopped with
Ctrl-C (terminal exit130). No launch options/install/timeout/skip changed;
RP-256 remains open, not a green or skipped population. Copy-check completed
normally on its original handle; no duplicate run was used as replacement.

Self diff filter confirms only care component/host, two isolation tests,
existing composed tool and docs/ledger/log changed after the test-first commit.
No server/math/kernel/schema/epoch/copy-content/workflow/generated byte.
Existing RP-132/GS4×PA7 and RP-318 copy/event obligations remain open. Other
Garage controls' pending discipline is not accepted by this narrow repair.
Next separately predeclare shared HTTP/exclusive refusal consumer checks:
GS0.2 says429 and exclusive activity stay disabled until the next snapshot;
the current mapper returns effect=none. Establish actual native behavior
before inferring a production remedy or changing that shared authority.

## 2026-10-07 — Predeclare shared refusal/read-boundary supplement

Clean baseline `3f867956`; earlier care range is exactly
`c7d8f815..3f867956`, independently awaiting Claude. Authority: accepted
GS0.2 shared table and existing GS0.8 focus-preserving pending rules, GS4
Founder revision/no automatic retry. RP-322 records source evidence first.

Tests first: native care activation with actual runtime-double typed409
conflict/intent,429 rate_limited/account and200 exclusive_activity; hold the
authoritative snapshot promise. Require one read, existing exact status key,
focusable aria-disabled pending control, no reactivation/request while held,
then Founder revision8 and fresh distinct intent ID only on the player's next
activation. This is consumer evidence, not an actual service rate-limit or
Soul-recovery session. Add400 invalid/care_action control with exactly one
fixed invariant and no mechanical bytes;401/404/503 plus transport errors
must take the existing offline path, disable care without retry/read, and
show the existing offline status in Settings. Parser/network trust is not
inferred from a runtime double. Retain all previous populations.

Only if current consumers fail the declared boundary: change the two shared
mapper arms to the EXISTING refresh effect, whose held read already keeps
pending focusable and guarded. No new timer/retry/session semantics, server/
wire/math/kernel/content/copy/CI change. Add exact mapper unit tests before
repair; independently sever429 and exclusive arms to demonstrate unit/native
failure, restore source exactly. Final types/client/build/boundaries, full
two-engine Garage native/performance and existing actual composed lanes.
Firefox's failed connection is not retried as a different environment or
excluded from acceptance. No box/status/archive/full-Garage/CI/1.0 promotion;
entire new range after `3f867956` including every record edge needs Claude.

Unmodified-product baseline: both new mapper cases fail (none vs refresh);
native18 selected assertions return4 failures/14 passes/82 unselected. Only
429/exclusive arms fail in both engines: snapshot read count is0 instead of1.
409 held-read/fresh consent and invalid/credential/server/network/unparsable
controls pass. The typed SyntaxError arm covers the table's existing
unparsable-body→offline path without claiming a real parser request.
Attempted CLIENT_TEST_FLAGS selector is not supported by the root target:
it ran the FULL client population,2 fail/9,735 pass/373 explicit browser skips,
not a focused6-case run. Both handles terminate normally; no product changed.
Tests land before the authorized two-arm mapper repair.

## 2026-10-07 — RP-322 shared refusal boundary repaired locally

Review by: Codex (implementer/self first-filter ONLY).
Recorded by: Codex. Range: entire new span after `3f867956`, including
predeclaration `3f231ba2`, failing-first `d362c95e`, two-arm correction,
docs/ledger/log and final tracking edge. Designated Claude pass remains owed.
Previous care range stays separately exact `c7d8f815..3f867956`.

Correction changes only429/exclusive mapper effect from none to the existing
refresh. No host queue, timer/retry/session, server/wire/math/kernel/copy/
epoch/workflow change. Native refusal population18/18 passes: a held read
keeps care focusable but aria-disabled/guarded; no extra request happens until
the read resolves AND the player activates anew, using Founder8/fresh ID.
Conflict, invalid and offline controls retain their intended distinctions.
They are runtime-double evidence, not service429 or real Soul-session tests.

Independent429→none fault: full client1 fail/9,736 pass/373 explicit browser
skips; native2 fail/16 pass/82 unselected. Restore429 before exclusive→none:
same counts, but only the exclusive mapper/native cases fail. Both are
compiling behavioral faults, not type failures. Exact restored mapper SHA256
`a7aa1eb05e014b9f7862a02834fc91492d5ff66f40c5b7be4b5fd20a6c751710`.
Final full immutable Garage file100/100 Chromium/WebKit, including minute
idle; performance1/22 explicit unselected; types/Svelte0 errors/warnings;
client9,737 pass/373 explicit browser skips (10,110 total),105 files pass/
22 skip;213-module built client; boundaries14/8/22 pass. No Firefox retry:
its zero-execution connection hold remains, no all-engine/AT/full-Garage gate.

Final existing actual built-client/Postgres/WebSocket composed lanes pass:
10 DOM clicks→active.click at revision21/zero expiries; Fiscal five Pitch
commands/one visible offer decline/credit1e0 at revision25; v4 transitions/
both terminal states/next-run/recovery. Actual DOM care feed receipt at
Founder5 and persisted public-state/reload pass; Cosmetics8.77s/75 audited
requests/no N5 violation. These actual workflows are not the runtime-double
429/exclusive population or minted release proof. Earlier copy/manifest/
topology observations remain dated, not claimed rerun in this boundary.

Self full-range diff filter: two production mapper lines only, two unit cases,
nine native consumer cases, canonical doc/ledger/log. No acceptance box/status/
archival or full1.0 promotion. Next accepted lane: independently predeclare
Fiscal GS1/GS0.5/GS0.8 native pending/read/keyboard/stale/restart/refusal checks.
Source Fiscal still natively disables pending buttons without aria/pending
text; its host readiness omits transportReady. The existing delayed Fiscal
test explicitly expects disabled=true: retain its revision proof but correct
that expectation only alongside new tests of the accepted focusable contract.
Do not silently copy care approval to Fiscal or modify unmeasured semantics.

## 2026-10-07 — Predeclare Fiscal native state/intent supplement

Clean baseline `b64a91af`. Previous turn progress: care/readiness proof and
RP-322 mapper correction changed authoritative state and executed evidence.
Authority: accepted Garage GS1-A2–A5, GS0.2/GS0.5/GS0.6/GS0.8, existing
GS1 copy keys and server projection. RP-323/324 source findings filed first.
No RFC-body, authored prose, backend/formula/clock/schema/content/epoch/CI edit.

Native Chromium/WebKit runtime-double population: Tab→Enter/Space harvest,
level and unlock paths in logical order; all three pending commands held
through intent AND authoritative read, focusable aria-disabled/control guard;
four applied harvest outcomes and all ordinary/row-cap/unknown refusals with
exact notices and invariant counts; recovering/resync/restart with stale
explanation/disabled controls/no auto intent; missing Founder revision;
unlock-owned and level-capped removal with heading fallback, plus no focus
steal when the player selects nav; isolated component pending callback guards
so the host queue cannot mask them. No real quarter/rate-limit/Soul/producer
proof inferred from these doubles. Keep the existing real composed workflow.

Correct only confirmed Fiscal consumer defects: existing pending/stale keys,
explicit native Tab entries and callback guards, host transport readiness,
captured-trigger focus fallback under GS0.2/GS0.6 (nearest surviving control
in the same region, else heading), snapshot-owned level-cap reason mapping.
Preserve host single-flight and disabled true for actual ineligibility.
The existing delayed Fiscal test keeps its held-read/revision assertions;
replace its contradictory pending native-disabled assertion only alongside
the new focusable aria-disabled proof in the same reviewed test range.

Tests land before product repair. Demonstrate independently compiling pending,
readiness/stale/Tab/focus and cap-mapping omission faults; remove each exact
mutation before another check. Final full Garage two-engine/performance,
types/client/build/boundaries/copy/manifest/topology and existing actual
Postgres/WebSocket composed. Firefox's zero-execution connection hold remains;
no repeated install/skip/timeout workaround or all-engine/AT acceptance.
Full new span after `b64a91af`, including predeclaration/tests/repair/docs/
ledger/log/tracking edge, requires Claude independently of prior ranges.
No checkbox/status/archive/mint/push/deploy/wholeCI/full1.0 promotion.

Initial native runtime-double baseline executes48 assertions:31 fail/17 pass/
98 unselected. Separately, the new test copy helper's broad Record params
does not satisfy the generated key-specific TS union; typecheck rejects that
instrument before product changes. Correct helper to parameterless keys plus
direct typed cost resolvers, no cast or safety relaxation. Re-run the unchanged
product baseline and types before recording it as the test-first checkpoint.

Corrected instrument types pass zero errors/warnings; unchanged-product native
baseline still31 failures/17 passes, including exact row-cap notice mismatch.
Before repair add the declared nearest-surviving same-region control case:
existing generator presentation/public fixture only, not a claim that this
second generator currently sells a Fiscal level. This must discriminate a
heading-only implementation, not leave that accepted clause unobserved.

Final test-first baseline:24 new Fiscal cases plus the preserved delayed-read
case execute50 assertions;33 fail/17 pass/98 unselected, types zero errors/
warnings. Both nearest-survivor cases fail on body; four harvest outcome
controls and missing-revision/no-focus-steal controls pass. The two browser
engines differ on transient native-disable focus loss; no engine's result
is substituted for the other. Native failures are assertions, not launch or
build errors. Tests are committed before any production change.

First repair passes50/50 selected native assertions and types zero errors/
warnings. Independent readiness omission fails2/48; stale-note omission
fails6/44,98 unselected in each. Probe orchestration then makes a restoration
error: replacing an empty line without structural context inserts stale text
inside the script. The pending-note probe and its attempted diagnostic rerun
execute ZERO tests with a Svelte parse/import failure. Neither is behavioral
evidence. Automation stops at the non-assertion failure; exact structural
restoration removes the misplaced text and restores the notes to the markup.
Expected component SHA256 is
`f68eefd0bb4c7db5c4828fec57c5200d1595804aef9292ecd814222d759f4479`;
host `f4ec76753a2ce837b2c2032067869214a6a5611838c370c3df28b469a53ea50a`.
Verify both before resuming. Future note probes replace the key with an
existing different key rather than an empty restoration target; no error
is attributed to product behavior or absorbed into passing/failing counts.

First full corrected population passes148 Garage assertions/performance1;
client9,737/397 explicit browser skips/types/build/boundaries/copy/manifest/
topology pass. Actual composed harvest/unlock/Pitch+WebSocket and care reload
pass. Before closeout, self first-filter finds two instrument blind spots,
RP-325: one default reason cannot distinguish a fixed key; one survivor cannot
distinguish nearest from first. Within the existing predeclared source-owned
reason/nearest-control properties, first execute those two shortcuts against
the current50 selected assertions, restore exact hashes, then refine only the
fixtures: different registered snapshot reason; two surviving buttons at
unequal distances. These legal public runtime fixtures are not new producer
catalog eligibility or minted-content claims. Prove the refined oracles fire
on those same shortcuts. Re-run final types/client/full native; production
bytes/build/composed/manifest stay unchanged if refinements are test-only.
Record every survivor as such, not a failed product check or discarded run.

Both shortcuts actually SURVIVE the old50 assertion population, performance1
also green: fixed `cap.fiscal_level.beige_tower` and stable first-survivor order.
Restore each source exactly. Source grammar inspection also finds that the
old one-survivor fixture orders levels beige→answering rather than byte-sorted;
its U2-double execution is NOT legal-wire evidence. Forward-correct the prior
"legal" description: refined multiple-survivor IDs are sorted and both their
initial/successor snapshots execute the actual parser. Cap variation likewise
uses a parsed snapshot and a different already-registered reason key. No new
current producer catalog or reason semantics are claimed. Refined population
has25 new Fiscal cases plus existing delayed-read case:52 selected assertions.

## 2026-10-07 — Fiscal consumer repair: final local proof / first filter

Review by: Codex (implementer/self first-filter ONLY).
Recorded by: Codex. Scope: full new span after `b64a91af`, including
predeclaration `bd36f5bc`, failing-first `d0651029`, oracle predeclaration
`86426db5`, production/test/docs/ledger/log correction and final tracking edge.
Designated Claude pass remains owed. Earlier shared refusal is separately
exact `3f867956..b64a91af`; care `c7d8f815..3f867956` and prior ranges remain.

Eleven compiling behavioral faults against the50-assertion population fire
independently (fail/pass,98 other assertions unselected): readiness2/48;
stale6/44; pending text6/44; harvest/level/unlock callback guards2/48 EACH;
Tab removal4/46; handoff omission6/44; heading-only instead of survivor2/48;
generic cap instead of row reason2/48; unconditional heading focus14/36,
including the selected-nav control. Two parse/import-error runs are invalid,
not included. Restored hashes match the recorded component/host pair.

RP-325's two survivors are retained: fixed default key and first-survivor
each pass all50. After parser-validated fixture refinement each fails2/50
in the52-assertion population, with98 unselected. No source repair was needed
for those additional cases: correct production bytes stay identical. Parsed
synthetic snapshots prove consumer grammar/selection, not current catalog
eligibility, real cap refusal or minted content. Every probe is removed.

Final immutable-source gates: full Garage Chromium/WebKit150/150, including
actual-minute Desk idle; performance1/22 explicit unselected; types/Svelte
zero errors/warnings; client9,737 pass/398 explicit browser skips (10,135
total),105 files pass/22 skip; boundaries14/8/22. Build213 modules, complete
copy658 keys/611 existing orphan warnings, content manifest and topology/
13 negatives pass before the final test-only refinements. Their production
source hashes remain identical; build/manifest/composed are not claimed
re-executed after the test-only fixture changes. No workflow or collector
exclusion was added; existing CI jobs collect the new cases.

Same final production tree's actual built-client/Postgres/WebSocket composed
passes:14 DOM clicks→active.production buff at revision25/zero expiries;
Fiscal five Pitch commands/no offer declines/credit1e0 at revision28; v4
transitions/both terminal states/next-run/recovery. Actual care positive
receipt/public band normal/feed ineligible at Founder5 persists across reload;
Cosmetics8.153s/74 audited requests/no N5 violation. Not Lucky, all refusal
arms, a status crossing or minted release-artifact proof. No new Firefox
native-Mac attempt; RP-256 zero-execution launch hold and manual AT remain.

Self full-range diff filter confirms two UI components only as permanent
product changes; test/refinement, canonical doc/ledger/log stay in-range.
No server/wire/math/kernel/balance/copy-content/epoch/workflow/generated byte.
RP-323/324/325 locally corrected, not designated-approved or closed. Separate
source finding RP-326: outcomes still use unscoped global chrome notice;
Fiscal/care own no outcome region, and tab selection does not scope it.
Do not infer GS0.6 acceptance from pending/focus proof. Next predeclare and
execute surface-local notice/async-origin/non-bleed checks before any repair.

The existing Linux browser image is locally available:
`mcr.microsoft.com/playwright:v1.62.0-noble`, image
`sha256:5e63ffc997394026acb443c255703c7278f97d43cdc3d49cf51bc8fdce7d5383`,
Linux ARM64 (not GitHub AMD64). After the committed tracking checkpoint,
predeclare and execute the existing full `make test-browser-ci` cold install/
three-engine population. That is an available safe local parity check; it
changes neither workflow nor Firefox acceptance policy and cannot be cited
as an executed hosted-CI run. Keep source immutable and poll its live handle.
Full nine-tier/product/platform/author/owner/numeric/content/review/release
holds remain. Goal active/progress; no box/status/archive/mint/publication.

## 2026-10-07 — RP-326 outcome ownership: predeclare native failing-first census

Baselinee950216a; full browser evidence separately4dcd9969..e950216a needs
Claude and is RED/incomplete (RP-327/328/236); no new Docker run. Fiscal
correction stays separatelyb64a91af..4dcd9969, previous care/shared refusals
independent. Read accepted GS0.6/GS1: panel owns its intent outcome status;
chrome owns cross-surface event announcements. Current globalintentNotice
violates the boundary in source; establish executable failures first.

Test-only supplement to existing Garage native file, existing runtime-double
fixtures and registered keys, Chromium/WebKit only.12 cases: Fiscal and care
each applied and ordinary refused outcomes; each in three populations:
exact own-panel polite result with no duplicate chrome result, completed
result then real nav transition to unrelated Meters with no result bleed,
held response then nav transition before completion with no wrong-tab result.
Require one correctly scoped request, actual nav focus preserved, unchanged
registered reason/success text. No new retention policy: late-away result
may be retained for origin or discarded; this census does not require either
or invent a result history. No manual AT claim. Runtime doubles are not real
care/quarter/refusal or parser-valid producer/minted-content evidence.

Success/failure thresholds are exact DOM placement/text/count, not duration
or a suffix match. Existing pending controls are distinct from outcome region;
count only exact outcome messages. Wrong-tab assertions inspect the entire
host, not merely a missing node in the new panel. Do not change prior tests
to fit new routing before recording the baseline. Cases fail at real DOM
assertions; import/compiler failures are invalid instruments, not proofs.
Production source unchanged until tests commit; narrowly repair only under
GS0.6 after the failed baseline, with own predeclaration and compiling faults
for placement/duplicate/origin. No copy text/kernel/server/schema/timer/queue/
retry/CI/owner decision change, no checkbox/status/archive/push. Keep source
immutable during live native handle; goal remains full1.0 active/progress.

## 2026-10-07 — RP-326 baseline executed; repair predeclared

Handle98729, rootmake native Chromium/WebKit selector `outcome-ownership`:
24/24 fail at DOM assertions,150 other cases explicitly unselected; two
files red,3.23s, valid HTTP24/24. Own-panel failures find exactly one result
in chrome; completed and held-away results leak into Meters in both engines.
Correct one-Founder7 request and nav focus controls execute before failures.
Typecheck65261 passes0errors/0warnings. No production byte changed. A
pnpm-help diagnostic emits no output and is stopped130; not a test result.

Now predeclare narrow GS0.6 repair: preserve the single existing mapped
notice/key and single-flight/refresh logic, bind its owner to the surface
at act invocation before any await, pass presentation-only mapped notice
to Fiscal/care, each renders one own polite outcome region. Render no
duplicate chrome outcome for either; other existing host results visible
only when their captured origin is current. Do not remap any reason/success
key or introduce receipt/history retention. Existing cross-surface stream
announcements remain in their separate deduped chrome region. Queued/new
commands retain existing revision, pending and consent semantics.

Demonstrate compiling faults: omit Fiscal notice, omit care notice, duplicate
chrome result, remove origin visibility condition, capture owner at late
completion rather than invocation (if that counterfactual survives, disclose
and improve a bounded distinguishing population, never count it as failed).
Restore exact hashes after each. Keep all prior tests; only adjust selectors
if needed to identify their now-owned outcome, preserving exact text/revision/
focus/no-retry assertions. Run full native Garage2engines+performance, strict
types/client/build/boundaries; Docker-backed real composed cannot run under
RP-236 and must not be claimed. No all-engine/AT/release acceptance. Tests
commit first; product/docs/records separately. Exact new boundary begins
e950216a exclusive through eventual record edge and needs Claude; earlier
browser/Fiscal/care/shared and all other spans independent. Goalactive.

## 2026-10-07 — RP-326 narrow repair and native discrimination

Review by: Codex (implementer first filter). Recorded by: Codex. Tests
predeclareda8f2a527, failing-first7cac400d; original24/24 real DOM failures
98729 are unchanged. Repair binds the one existing mapped notice to the
surface captured before act's queue/read awaits. Fiscal/care get optional
presentation-only CopyKey `notice` props, no duplicate mapping authority;
one own polite outcome region after the heading. Fiscal region spans the
existing grid. Generic chrome result is invisible outside its captured
origin; cross-surface stream announcements remain unchanged. No result
history, retry/timer/copy text/kernel/server/schema/balance/owner change.

Initial restored24/150unselected plus performance1/22unselected pass73952;
strict types67851 zeroerrors/warnings. Five independent compiling native
faults (full24selected,150explicitlyunselected each):

| Fault | Failed | Passed | Make exit |
| --- | ---: | ---: | ---: |
| Omit Fiscal notice | 8 | 16 | 2 |
| Omit care notice | 8 | 16 | 2 |
| Duplicate result in chrome | 8 | 16 | 2 |
| Remove chrome origin condition | 16 | 8 | 2 |
| Capture owner at response completion | 8 | 16 | 2 |

All fail at AssertionError, not syntax/import errors, and source is restored
before the next fault. Final exactSHA256:
GameUIApp9ea86275b688712047703e2e6b794fcf5d6b5b61706d874b51ee13462459223a,
FiscalSurface76cf10c5dece9652ea338c9768d38b8533fa89dfb708f464b7c32a20965df8f1,
PetCareSurface35e72f3987d5b58c913e8513bb5c0373e030dd7b361a9ae54a46107c3004af99.
No prior assertion, selector or test was deleted/relaxed to obtain the pass.

Full restored native Garage97700:174/174 in Chromium/WebKit,71.01s,
including existing real-minute idle control; original isolated performance
1/22explicitunselected,318ms/1.28s. HTTP170/170+performance1/1 valid, no
pending/earlycloses. Tool retrieval truncates15,665tokens of functional
trace; terminal count/summaries/performance directly retrieved, not a claim
of complete per-request traces. Root57092: types0/0, client9,737pass/
410declared browser skips of10,147;105filespass/22skip; Vite213modules,
index-DwwjDcZC.js/index-DhbUhBbR.css and unchanged workerMqspU_iu; boundary
14shell/8UI/22GameUI pass. Copy/topology91872 still live at this record;
source/HEAD held unchanged until terminal, no pass claim yet.

This is native consumer proof for Fiscal/care applied/ordinary refusals and
origin isolation, not all-surface/HTTP-failure/queued-origin/fullGS0.6/AT/
Firefox or hosted-CI acceptance. Real composed not rerun on this changed
production tree because RP-236 disk capacity remains unresolved; earlier
Fiscal real composed belongs to the earlier source, never substitute it.
RP-132/313/318 and author/copy/content/numeric/privacy/release/all prior
holds remain. Exact newspan starts e950216a exclusive through final records,
Claude gate remains; no boxes/status/archive/push/mint. Full1.0 progress.

Copy/topology91872 now terminal0 on the same source/HEAD:658 keys,
hasha5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611existingorphan warnings; generated content manifest unchanged/valid;
CI topology+13seeded negative controls pass. This is local copy/manifest/
topology evidence, never hosted execution. All verification/fault handles
terminal; source hashes still exact above. The isolated earlier failed
artifact/log-link patch is disclosed in platform log and applied no partial
log change; no behavioral implication. Commit product/docs/ledger/log now,
then reconcile tracking as a separate same-range record edge.

## 2026-10-07 — HTTP-error consumer census predeclared

Resume at clean6d700838. Accepted GS0.2/GS0.5/GS0.6 authorize this
test-only supplement; no new mechanics or production repair assumed.
Population: Fiscal harvest and care feed, each mounted from the existing
public runtime-double fixture, native Enter, immediate or held until a real
Meters navigation. Eight error arms: parsed400/409/429/401/404/503, transport
TypeError and malformed-response SyntaxError. Total32 declarations across
native Chromium/WebKit =64 selected assertions. Existing tests unchanged.

Predeclared outcomes:400 exact generic own-panel status, one fixed invariant,
no read/offline/retry;409/429 exact own-panel status and one held snapshot
read, no activation until Founder8 read completes, then new explicit consent
with distinct intent ID/revision8;401/404/503/transport/malformed go offline,
no snapshot/read/replay/mechanical message/invariant. After held-away errors,
no outcome text on Meters, navigation focus preserved; returning to origin
may show the last mapped notice but is not a retention/history promise.
Inspect whole host for leaks and exact status count/ownership. Initial
runtime-double errors do not prove fetch parsing, actual server refusal,
credential renewal, AT announcement delivery or queued-action origin.

Run selected population before any product change. A green baseline is
valid negative research: do not manufacture a repair. Then demonstrate
compiling faults independently: sever429 refresh, sever400 invariant flag,
capture notice owner on error completion. Each must fail real DOM/behavior
assertions (not compiler/import errors), restore exact source hashes before
next fault. If a fault survives, disclose/refine within this population.
Run restored full two-engine Garage including unchanged idle/performance,
strict types/client/build/boundaries. Docker-backed composed/fullLinux held
by RP-236; no new container run or cleanup, no prior-source substitution.
No copy/kernel/server/schema/CI/bounds change, no checkbox/status/archive/
mint/push. New review range starts6d700838 exclusive through final records;
Claude designated gate and all earlier independent ranges remain owed.

First instrument run82322 exits2:32pass/32fail/174unselected,6.20s;
immediate cases reject their unused deferred response during cleanup,
producing unhandled rejections. This is Codex test scaffolding error, NOT
a product baseline defect. Typecheck52568 zeroerrors/warnings. Correct
cleanup to reject only the late-away response actually consumed by intent;
no assertion removed/relaxed and no production source change.

Corrected baseline41186 passes64/64 executions (32 declarations, multiple
assertions per case),174 explicitly unselected,6.47s; isolated unchanged
performance1/22unselected,314ms/1.25s. ValidHTTP64/64+1/1. Tool retrieval
truncates10,341tokens; terminal summaries directly retrieved, not complete
HTTP traces. No product defect in this bounded population; keep source
unchanged and commit the test-only supplement before predeclared faults.

## 2026-10-07 — HTTP-error census completed, no production repair

Review by: Codex (implementer first filter). Recorded by: Codex.
Predeclarationca8c9b55/test8c364fbd. Green bounded baseline retained;
no production byte changed. Three separate compiling faults each run the
same64 selected executions/174 explicitly unselected, fail at actual
behavior/DOM assertions, Makeexit2:

| Fault | Handle | Failed | Passed | Diagnostic |
| --- | --- | ---: | ---: | --- |
| 429 refresh→none | 94261 | 8 | 56 | expected one snapshot, observed zero |
| 400 invariant→false | 73308 | 8 | 56 | expected fixed diagnostic, observed none |
| Error completion captures current surface | 80006 | 12 | 52 | exact mapped error leaks on Meters |

Restored each before next; final SHA256 matches baseline:
GameUIApp9ea86275b688712047703e2e6b794fcf5d6b5b61706d874b51ee13462459223a;
intent-outcomea7aa1eb05e014b9f7862a02834fc91492d5ff66f40c5b7be4b5fd20a6c751710.
Final94786 terminal0: full Garage Chromium/WebKit238/238,76.65s, original
real-minute idle control included; unchanged isolatedperformance1/22
explicitunselected,318ms/1.30s. HTTP234/234+1/1 valid/noearlyclose/pending.
Retrieval truncates intermediate HTTP output (10,794/15,406/10,420 tokens),
but final counts/performance/exit are directly retrieved; no full trace claim.
Root76930 terminal0: stricttypes0errors/0warnings; client9,737pass/442
explicitbrowser skips of10,179,105filespass/22skip; build213modules with
unchanged index-DwwjDcZC.js/index-DhbUhBbR.css/workerMqspU_iu; boundaries
14shell/8UI/22GameUI pass. No previous assertion deleted/relaxed.

Capacity recheck is read-only on the existing declared game's Postgres
container (`docker exec cloud-clicker-game-ui-postgres-1 df -h / /dev/shm`):
overlay125.7G/122.7Gused/0available/100%; sharedmemory63Mfree. RP-236
persists. No new Docker population, container/image/cache/volume deletion,
or inference that disk explains every earlier Linux failure. Real composed,
Firefox, AT, hostedCI and minted-release proof remain outstanding.

This closes only the predeclared consumer research population: existing
HTTP-error behavior satisfies the narrow contract, not all-surface/queued-
origin/wire-parse/account-renewal/fullGS0.6/fullGarage/release acceptance.
Next safe accepted work: separately predeclare real-runtime/fetch-double
error parsing and request-count preservation at the GS0.2 boundary; no
actual service/refusal/credential proof can be inferred from mocked fetch.
New exact review boundary begins6d700838 exclusive THROUGH this tracking
edge, including predeclaration/test/records, and requires Claude. Previous
notice exacte950216a..6d700838, browser4dcd9969..e950216a, Fiscalb64a91af..
4dcd9969 and all earlier ranges independent. No boxes/status/archive/copy/
mint/push; full-nine-tier1.0 goal remains active, meaningful progress.

## 2026-10-07 — GS0.2 real-runtime HTTP boundary predeclared

Resume clean a42406f0. Accepted GS0.2 runtime contract, not a new API or
credential policy. Extend existing game-ui-intent-outcome unit suite with21
declarations exercising createBrowserGameUIRuntime and actual Response.json,
while substituting fetch only: six parsed non2xx statuses400/409/429/401/
404/503; seven malformed409 error bodies (extra/missing keys, null/array,
wrong detail type, nonmechanical category, empty detail); two valid200
outcome arms; two invalid200 receipts; nonJSON200/503; transport rejection;
missing credentials. Existing six tests unchanged. Synthetic mechanical
pairs validate parsing, not server registry membership or actual refusals.

Success exact: valid errors remain GameUIRequestError with unchanged status/
pair and full existing mapped effect/notice/invariant; malformed bodies
remain generic transport errors/offline, never actionable typed409; valid
outcomes retain receipt/revision/refusal fields; invalid200/nonJSON fail
closed; transport error identity retained. Every attempted credentialed
intent emits exactly one POST to the intent path, exact bearer/JSON headers
and body, leaves credential storage unchanged, and performs no implicit
read/renewal/retry. Missing credentials emits zero requests. Test dummy
credentials only. No real network/server/DB/browser/AT/renewal proof.

Run full root make test-client baseline before any product change. A green
baseline completes bounded research without inventing a repair. Commit
tests before four independent compiling faults: ignore non2xx branch;
accept extra error-body keys; strip expected_revision from submitted body;
duplicate intent POST. Each must cause assertion failures, not compiler/
import errors; restore exact runtime/parser hashes before the next probe.
Retain any surviving fault honestly; refine only a distinguishing population
within these declared properties. Final strict types/client/build/boundaries.
No copy/kernel/server/schema/CI/timeout/retry/credential policy changes,
no checkbox/status/archive/mint/push. Docker-backed runs remain RP-236 held.
New review range starts a42406f0 exclusive through final records, requires
Claude independently of HTTP-consumer6d700838..a42406f0 and earlier spans.

Baseline86969 terminal0: full client9,758pass/442explicitbrowser skips
of10,200,105filespass/22skip,5.83s; strict types0errors/0warnings. All21
new runtime/fetch-boundary declarations pass on unchanged production source.
Negative result retained: no repair authorized or needed in this population.
Commit tests now, then execute the four predeclared compiling faults.

## 2026-10-07 — Real-runtime HTTP boundary proof completed

Review by: Codex (implementer first filter). Recorded by: Codex.
Predeclared88ab6bdd/test2d3629ad. Baseline green, no production repair.
Four independent compiling faults, same full client population each time,
Makeexit2 and actual assertions (not import/compiler errors):

| Fault | Handle | Failed | Passed | Witness |
| --- | --- | ---: | ---: | --- |
| Ignore non2xx branch | 18847 | 7 | 9,751 | six new typed-error cases plus original runtime case |
| Admit extra error-body keys | 6789 | 1 | 9,757 | malformed409 must not become GameUIRequestError |
| Strip expected_revision | 60806 | 20 | 9,738 | exact single-request body comparison |
| Duplicate intent POST | 44237 | 21 | 9,737 | nineteen new call-count cases plus two original tests |

442declaredbrowser skips retained in each denominator. Duplicate-send
transport rejection stops on its first failed fetch, so that arm correctly
does not distinguish a second send that never occurs. Missing credentials
also correctly sends zero. No surviving applicable fault hidden. Source
restored before every next probe and final SHA256 matches baseline:
runtime0a8c420eb9968aa76a3e9ece91c918a43e3d8c4530773c8818bce5104c0cb416;
intent-outcomea7aa1eb05e014b9f7862a02834fc91492d5ff66f40c5b7be4b5fd20a6c751710.
Fault60806/44237 output retrieval truncates3,222/3,471tokens; exact failing
diagnostics and terminal totals retrieved, not complete trace claims.

Final90506 terminal0: stricttypes0errors/0warnings; client9,758pass/442
explicitbrowser skips of10,200,105filespass/22skip,5.60s; build213modules
with unchanged index-DwwjDcZC.js/index-DhbUhBbR.css/workerMqspU_iu; boundaries
14shell/8UI/22GameUI pass. Tests21new+six original declarations; all prior
assertions retained. Production/server/kernel/copy/CI/RFC/schema zero delta.
No fresh native/browser/composed/Go/hosted claim; previous native238 evidence
belongs to a42406f0, source-identical here, not a rerun. Docker RP-236 remains.

This proves actual runtime and Response.json over fetch-double replies,
not actual service/error populations, generated API conformance, credential
renewal, browser/AT delivery or release artifact. Next reconcile GS1–GS6
acceptance against current source and named evidence, separating author/
content/producer/default-workflow/engine/AT/review holds, then choose the
next unresolved accepted requirement. Do not generate more tiny tests just
to grow a count. New exact span beginsa42406f0 exclusive THROUGH final
records and needs Claude. Prior HTTP-consumer6d700838..a42406f0, notice
e950216a..6d700838 and all preceding spans remain independent. No boxes/
status/archive/mint/push; full nine-tier1.0 remains active/progress.

## 2026-10-07 — GS1–GS6 acceptance reconciliation predeclared

Resume cleanfc911784. Inventory all29 named GS1–GS6 gates and eight overall
RFC criteria against current producer/consumer/content/test/runtime records;
write acceptance-evidence.md with exact source coordinate, evidence scope,
named open gate and next route. No status/checkbox promotion, no assumption
that historical composed/native proofs execute on this source or cover all
fixtures/engines/default players. Review obligations remain independent.
Cold focused gameui/fiscal/achievements/meters Go population through root
Make, -count=1 and verbose explicit SQL skips; native browser populations
only after own bounded predeclaration. Docker RP-236 remains; no run/cleanup.

Source reconciliation identifies a candidate GS0.5 gap: Achievements.title
and Meters.band throw on unavailable presentation/copy, and no host boundary
renders common.surface_error while preserving other surfaces. Inspect actual
decoder admission before calling this a defect; malformed wire is distinct
from decoder-legal unresolved presentation. If legal, predeclare two native
error-state cases (unknown achievement copy and declared unknown meter band)
with healthy controls, exact existing alert text, no mechanical leak/intent,
one invariant and operable other navigation. Observe errors explicitly;
expected render crashes are baseline evidence, not passing acceptance. No
production edits in this audit range; a repair requires separate scope.
Missing specified copy remains an author hold, never invented prose.

New documentation/evidence range beginsfc911784 exclusive through final
tracking and needs Claude. Goal remains full nine-tier1.0; no archive/mint/
CI/copy/kernel/schema/owner decision/push, no test-count completion proxy.

## 2026-10-07 — Full Garage acceptance inventory, not a completion claim

Review by: Codex (source/evidence reconciliation, self first filter only).
Recorded by: Codex. Source fc911784; predeclaration124b086a. All29 named
GS1–GS6 gates plus eight overall criteria now have producer/consumer,
actual evidence scope and unresolved route in acceptance-evidence.md.
No product byte, checkbox, RFC body or lifecycle status changed.

Fresh root focused gameui/fiscal/achievements/meters Go run -count=1 -v,
handle51228 terminal0:116 PASS lines including parent/subtest reports.
Two SQL tests explicitly skip: stored schema-v4 rate projector and
Reputation current/next projection, TEST_DATABASE_URL unset. No fresh
database evidence. No fresh TS/native/three-engine/composed/copy claim.

RP-329 records source-admitted partial meter sets; RP-330 semantic narrow
layout mismatch; RP-331 names missing exact DOM acquisition/meter and
persisted Fiscal preview witnesses. These are not invented mechanics or
claims those kernels fail. Copy-linter inspection found the dedicated
companion safety rule and actual negative fixtures: GS4-A4 is not absent.
Its Garage diegetic/PA8.2 companion body reconciliation remains author-owned.
Unknown achievement-copy/declared unknown-band render containment remains
a candidate to reproduce, not a measured defect yet.

Next separately predeclare two native GS0.5 error populations with decoder-
admission/healthy controls, explicit render-error observations and exact
alert/invariant/navigation assertions. No product repair in this audit range.
Docker capacity/Firefox/AT/body/review/full1.0 holds retained. New complete
audit span beginsfc911784 exclusive through its final records, requires
Claude independently of a42406f0..fc911784 and every earlier range.

## 2026-10-07 — GS0.5 contained presentation error predeclared

Start cleancc62cea8, after the documentation-only acceptance audit range
fc911784..cc62cea8. Authority: accepted GS0.5/GS2/GS3, existing
common.surface_error copy only. New diagnostic range, not audit scope creep.

Population: two decoder-legal unavailable mappings (one achievement row's
unknown copy key; doom's declared critical band absent from presentation),
each first-open and already-mounted refresh, native Chromium/WebKit. Four
declarations/eight executions. Healthy controls first; actual decoder must
admit the complete snapshot. Unknown mechanical keys must never enter DOM.
Observe synchronous and window render errors explicitly; no console-error
suppression without a spy/asserted diagnostic. Intended gate: contained
surface heading/one exact role=alert, no controls/content values, exactly one
invariant per error episode (unchanged refresh/tick does not duplicate),
healthy snapshot restores content; renewed missing mapping reports a new
episode; Desk/Settings navigation remains operable; zero intents.

Run root native selector and types. First failed baseline is expected; no
production edit before recording it. If legal input reproduces render failure,
ledger it and predeclare a separate bounded implementation step before repair.
Do not repair malformed wire/module import failures, full shared state/focus,
missing-ID decoder, semantic reflow or any producer/copy/kernel/schema here.
No Docker population; capacity still held. Full engines/AT/composed/review/
body/release remain independent. Complete new span startscc62cea8 exclusive
through its final record edge, requires Claude, never self-approval/archive.

## 2026-10-07 — GS0.5 legal-wire render failure reproduced

Predeclared4c9e06b2. Four new diagnostic declarations, two native engines.
Healthy controls and actual parseGameUISnapshot admission reach the error
in each case; first-open/live-refresh unknown achievement copy and declared
unknown doom band throw uncontained RangeErrors. Fresh native29456 terminal2:
eight failing executions,238 unselected declarations; also reported render
exceptions, not sixteen independent failing tests. Performance lane not
reached. RP-332 ledgered; no production change yet.

Instrumentation disclosure: initial typecheck34418 failed my missing union
narrowing in the diagnostic. Native93058 independently showed the same eight
runtime failures, but is not counted as a typed-clean baseline. Corrected
the test narrowing only; typecheck40369 then exited0 (zero errors/warnings).
Typed-clean cold native29456 reproduced all eight; full output retrieved
without tool truncation, selected tail displayed. Window errors are observed
and asserted absent, not converted to successful acceptance by preventDefault.
Vitest still reports the thrown errors alongside the failing assertions.

Test-first checkpoint carries only diagnostic/tests/ledger/records. Separate
implementation predeclaration follows before any production bytes. Existing
tests unchanged; audit exactfc911784..cc62cea8 remains independent. New
error range startscc62cea8 exclusive through final records, requires Claude.

## 2026-10-07 — RP-332 implementation step predeclared

Failed-first532855fe; accepted GS0.5/GS2/GS3. Only AchievementsSurface and
MetersSurface production bytes plus canonical docs/ledger/records: derive
required-mapping availability before rendering their values; show existing
common.surface_error alert beneath the unchanged heading on failure; one
fixed invariant on entering error, not on unchanged snapshot/tick; healthy
authoritative refresh restores content. Existing closed mapping guards stay
as defense-in-depth; no fallback values/strings, timer/retry/network or new
copy. Component-local containment, not a catch-all that hides other defects.

Verify typed-clean targeted native then full existing Garage Chromium/WebKit
population/performance, full client/types/build/boundaries. Independent
compiling fault probes: suppress error branch in each panel; omit alert
semantics; suppress invariant; fail to reset episode diagnostic on recovery.
Each must fail an intended behavioral assertion; restore exact production
hashes before final gates, no source mutation during a live check. Original
failed render baseline is also a demonstrated severing case. Record survivors
or instrumentation mistakes, never loosen assertions/budgets. No Docker run,
schema/kernel/balance/owner-copy/host/semantic reflow/missing-ID repair.
All-engine/AT/composed/default-player/body/review/full1.0 remain open. Whole
new span aftercc62cea8 through final records requires Claude, not self-approval.

## 2026-10-07 — RP-332 containment locally corrected and discriminating

Review by: Codex (self first filter, NOT designated approval).
Recorded by: Codex. Predeclarations4c9e06b2/81b4f51d; failed-first532855fe.
Only two read-only components changed: validate required mapping availability
before content rendering, existing common.surface_error alert, one fixed
invariant per mounted error episode; unchanged snapshots/ticks don't duplicate,
healthy refresh restores content and resets reporting for the next episode.
Host/network, guarded kernel, wire, schema, balance, owner copy and CI unchanged.
Original tests retained; no catch-all that swallows unrelated errors.

Typed targeted28557 exits0 (zero errors/warnings); native80980 exits0:
8/8 new executions,238 unselected; unchanged isolated performance1 passes.
Five independent compiling faults (each terminal root Make exit2):

| Fault | Handle | Failed / passed / unselected | Actual discrimination |
|---|---|---|---|
| Bypass Trophy Case error branch | 19611 | 4 / 4 / 238 | observed RangeError and absent containment |
| Bypass Meters error branch | 48196 | 4 / 4 / 238 | observed RangeError and absent containment |
| Change alert to note | 45938 | 4 / 4 / 238 | zero alerts instead of one |
| Omit Meters invariant | 36045 | 4 / 4 / 238 | zero diagnostics instead of one |
| Never reset Trophy Case episode | 45455 | 4 / 4 / 238 | second episode reports one total instead of two |

All faults removed before final gates, exact production SHA256 restored:
Achievements877a39910e1b859e962a0929d674973c3010052ccf20bcd1bac2226fae2901a9;
Meters1d22055bf9ed92f260958487f6c519ef622147b6cee319b932421833c8b1df32.
No source edits while a corresponding check was live.

Final plain root native67775 terminal0:246/246 Garage executions across
Chromium/WebKit,77.15s including the unchanged actual60s idle check. Module
summary242/242 valid. Separate unchanged isolated performance1/22 unselected,
317ms observation/1.32s process. Intermediate native output was tool-truncated
(12,031→3,000 and25,833→24,000 tokens); terminal populations/diagnostics are
retained, not a full-trace claim. Initial compound read/check launch hit EPERM
before collection; no test result claimed, reissued the ordinary Make command.

Final10237 terminal0: types0/0, client9,758pass/446 explicit browser skips
(105 files pass/22 skip),5.15s; build213 modules/index-Crfq0_YP.js,
unchanged CSS and worker; boundaries14shell/8UI/22GameUI pass.
Copy/topology19498 terminal0:658 keys, unchangeda5df8920…1adeb0e5e hash,
611 orphan warnings; content manifest/topology and13 negative controls pass.
Fresh read-only declared Postgres container df still125.7G/122.7Gused/
0available/100% overlay,63M available shm. No Docker population/cleanup.

Performance source inspection found RP-333: the passing isolated screen
scenario has null feature arms and sixty simulated seconds, so it is not
the populated Garage AC7 witness. The gate stays open; no threshold change
or performance regression inferred. RP-329/330/331 and all author/body/
Firefox/AT/composed/review/full1.0 obligations remain. Next accepted bounded
work: predeclare complete/missing meter-ID controls before GS3-A1 repair.
Whole error span aftercc62cea8 INCLUDING following record edge needs Claude;
audit exactfc911784..cc62cea8 and earlier ranges independently owed.

## 2026-10-07 — RP-332 tracking closeout, not designated approval

Production/docs/ledger/evidence checkpoint8ea7cb94 follows failed-first
532855fe and both predeclarations. Living plan/current-state/roadmap/queue
now point to the bounded correction and next accepted GS3-A1 RP-329 work.
No box/status/archive crossed. Review by: Codex (record consistency first
filter only). Recorded by: Codex. Whole range startscc62cea8 exclusive
THROUGH this record edge, including all tests/production/docs/record commits;
Claude must supply its designated verdict. Audit exactfc911784..cc62cea8,
runtimea42406f0..fc911784 and all earlier ranges remain separately owed.
All RP-329/330/331/333/Docker/engine/AT/author/body/default-player/platform/
privacy/numeric/full-nine-tier1.0 holds preserved; goal active/progress.

## 2026-10-07 — GS3-A1 complete meter-ID decoder repair predeclared

Resume clean5fbf4cff. Previous goal turn is progress (whole acceptance audit,
actual render baseline and RP-332 two-panel correction), not a waiting turn.
Accepted authority: GS3-A1 rejects missing IDs/out-of-range values/unsorted
rows, GS0.1 sorted exact arms; RP-329 identifies current admission gap.

One existing production authority already exports REQUIRED_METER_IDS from
client/src/meters/catalog.ts; the Go catalog enforces the same eleven IDs.
Import that immutable ID contract into game-ui/contracts.ts, never edit its
guarded file or add a second list/count/balance version. Tighten only non-null
meters arms; null/legacy snapshots/valid band declarations remain compatible.
No API registry/wire producer/schema/kernel/balance/copy/UI/CI change.

Failed-first TS population: derive complete public rows mechanically from
the existing first-content meter artifact (test input, not deploy-current
production fallback). 56 declarations: complete/null/immutable controls;
each11 missing IDs; each11 count-preserving unknown-ID replacements;
one extra row; empty/doom-only/trust-only subsets; three ordering cuts;
22 lower/upper value violations; extra/missing row fields. Reconcile the
existing positive v4 fixture to a complete set and retain its original
domain/undeclared-band negative checks with full rows, not incidental count
failures. Add a cold Go producer control over actual projectMeters/pinned
catalog: exact sorted IDs and refusal of each missing saved value. No DB
claim. The full shared-v4 producer/decoder fixture criterion remains open.

Record a typed-clean failed baseline before production. Bounded repair uses
sortedRows plus exact count/ID equality against the exported contract. Seed
independent compiling faults: omit completeness; retain count only; retain
positional IDs only without length; omit ordering; omit value bounds. Record
actual failing assertions/populations and any survivor or instrumentation
error; restore exact hashes, no edit during a live matching check. Final
root types/client/build/boundaries and focused cold Go; native Garage remains
mandatory on the changed decoder, no Docker population until RP-236 repair.
Complete new span starts5fbf4cff exclusive through final records and requires
Claude separately from RP-332cc62cea8..5fbf4cff, auditfc911784..cc62cea8 and
all earlier ranges. No checkbox/status/archive/mint/owner-copy/push. Whole
nine-tier1.0/engine/AT/author/body/default-player/platform gates stay intact.

## 2026-10-07 — RP-329 typed-clean admission baseline reproduced

Predeclaredbff97b20, no production edits. 56 new TS declarations derived
from balance/meters/first-content.json, no copied ID list. Existing positive
v4 fixture now has the complete set; its original out-of-range/undeclared-band
negative assertions retain their expected domain error and all unaffected
rows. No assertion relaxed or existing check removed.

Root typecheck23579 terminal0, zero errors/warnings. Root full client94222
terminal2:27 failing admission assertions,9,787 passing tests/446 explicit
browser skips (one failing file/104 passing/22 skipped). All11 missing IDs,
all11 count-preserving unknown replacements, extra row, three subsets and
non-mutating incomplete-input refusal fail because current decoder admits
them. The other29 new controls pass, including complete/null data, each22
value bounds, three unsorted cuts and exact-field failures. Failure text is
actual “expected function to throw”; not malformed fixtures/import/compiler
or unrelated assertion failures. Population6.07s.

Cold root Go17490 -count=1 -v scoped TestMetersProjectionUsesCompleteCatalogIDs
terminal0: actual pinned projectMeters produces exactly catalog IDs and rejects
each11 missing saved values,12 parent/subtest PASS reports,0.237s. No SQL
selected or DB proof claimed. RP-329 ledger updated. Test-first checkpoint
only tests/records; production completeness repair follows. Complete new
range after5fbf4cff through following records needs Claude independently of
RP-332cc62cea8..5fbf4cff/auditfc911784..cc62cea8/all prior spans.

## 2026-10-07 — GS3-A1 first green and producer-control faults predeclared

Client repair imports the existing immutable REQUIRED_METER_IDS only; exact
count/positional ID check follows existing sortedRows validation. First root
types/full client53254 terminal0:zero type errors/warnings,9,814 passing/
446 explicit browser skips (105 files pass/22 skip),5.69s. This is local
admission proof, not all-engine/shared-v4/real DB/release acceptance.

Before executing negative probes, add two scoped cold Go producer-control
faults to the five already predeclared decoder faults: omit one emitted row,
and sever the !present saved-value refusal (leave all other checks intact).
Actual TestMetersProjectionUsesCompleteCatalogIDs must fail its complete-ID
or each missing-value assertion respectively. These are transient probes,
not authority to land producer/schema/kernel/balance changes; source restored
exactly before final cold Go/vet and native/client gates. Record surviving
probes/errors without changing the fixture. Baseline restored hashes:
contracts43bfd6140dc7479eb800871fbc9225b914fe2a36a4a6364a813cfa4a3eabcb2b;
server features e344ecee553f1260e1a33802985d98c45970ff6035cd677ebf4548425f61f0fa.

## 2026-10-07 — GS3-A1 negative probes and restored native baseline

Authority/predeclarations: bff97b20 and c5d5220c; failed-first tests
43ef0a21. Five independent, executable client faults ran through the root
`make test-client`. These are actual failed admission assertions, not compiler,
fixture or import errors. Every population retained 446 explicit browser
skips and 105 collected non-browser files (one failed, 104 passed).

| Transient fault | Run | Failed / passed | Discriminating failure |
|---|---|---|---|
| Completeness check disabled | 66328 | 27 / 9,787 | Missing IDs, count-preserving replacements, extra/subset rows and incomplete-input refusal |
| Count only, ID equality removed | 37878 | 11 / 9,803 | Every count-preserving unknown-ID replacement was admitted |
| Positional IDs only, length removed | 99706 | 4 / 9,810 | Missing final ID, empty/doom-only prefixes and immutable-input partial refusal |
| Incoming row copy silently sorted before validation | 82545 | 3 / 9,811 | All three unsorted full-set populations were admitted |
| Meter value bounds removed | 31646 | 23 / 9,791 | All 22 new lower/upper bounds plus the retained original malformed-arm value check |

Each root command exited 2. The ordering fault is deliberate normalization
of a copy, not merely bypassing the sorted assertion and relying on another
ID assertion to reject the same data. None of these faults was committed.

Two separately predeclared producer-control probes ran cold through the root
Go target with `-count=1 -v -run TestMetersProjectionUsesCompleteCatalogIDs`.
Omitting the final emitted row (17280) exited 2 at the exact projected-ID
comparison. Severing the missing-saved-value guard (89370) exited 2 with all
eleven missing-value subtests failing: each returned an arm with fabricated
zero instead of nil/ErrInvalidProjection. The map lookup was adjusted to avoid
an unused-variable compile error; this was an actual behavioral failure.
No test assertion was relaxed and no producer change was retained.

Restoration: contracts.ts SHA-256
43bfd6140dc7479eb800871fbc9225b914fe2a36a4a6364a813cfa4a3eabcb2b;
server/gameui/features.go SHA-256
e344ecee553f1260e1a33802985d98c45970ff6035cd677ebf4548425f61f0fa.
Server diff is empty. No source edit occurred during a matching live check.

Clean root types/client/build/boundaries (96059) exited 0: zero type errors
or warnings; 9,814 passed / 446 explicit browser skips, 105 passing files /
22 skipped files; client population 14.14 s. Build: 214 modules,
index-DOWKMmbZ.js; unchanged index-DhbUhBbR.css and prediction.worker-MqspU_iu.js.
Boundary scan covers 14 shell, 8 UI and 22 Game UI components. Cold root Go
`./gameui ./meters -count=1 -v` (13144) exited 0; both explicitly reported
SQL integration skips because TEST_DATABASE_URL is unset. No persisted-proof
claim. Root vet for those two packages exited 0. Native full Garage and
copy/topology results follow after their terminal outcomes; no premature pass.

Review by: Codex (self first filter only). Recorded by: Codex. The complete
new span begins after 5fbf4cff and includes predeclarations, failed-first
tests, production/docs and final records. Claude's separate designated pass
is still required. No acceptance checkbox, RFC status, archive, mint, push
or release claim changed; the full nine-tier 1.0 objective remains active.

## 2026-10-07 — GS3-A1 final native and tracking gates

Root full Garage Chromium/WebKit invocation (44628) exited 0: 246 / 246
executions, two passing files, 80.28 s including the actual 60-second idle
population. Its chained, unchanged isolated performance lane passed one
test / 22 unselected declarations (386 ms test time). This is RP-333's
null-feature screen guard, not the populated Garage overall AC7. No browser
assertion or timeout was changed to accommodate the stricter decoder.

One intermediate browser log chunk was tool-truncated (12,053 original
tokens, 9,000 returned); the later terminal chunk retained all result lines.
No claim of complete retained module-HTTP trace or all-engine evidence.
Root copy/content-manifest/topology (88254) exited 0: 658 copy keys,
unchanged SHA-256 a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611 existing orphan warnings; content manifest valid; 13 topology negative
controls rejected. No workflow, generated contract, content or balance edit.

The production repair is exactly the existing ID-contract import plus
count/positional-ID admission after sortedRows; source hashes remain the
restored values above. Documentation and RP-329 state now distinguish this
local repair from designated approval and integrated release proof.
Next safe accepted work: RP-330 / GS3-A3 semantic narrow Meters layout,
with failed-first header/value association and breakpoint controls; RP-333
populated performance and the remaining exact guards follow separately.
RP-331 persisted/default-player witnesses still require Docker capacity
restoration/recheck. No Docker run or resource cleanup was attempted here.

Review by: Codex (self first filter, not the designated pass). Recorded by:
Codex. No catalog/kernel/producer/schema/copy/CI mutation survives the probes;
the landed proof uses existing contracts, retains previous assertions and
all null/legacy admission paths. The complete new range begins after
5fbf4cff, including bff97b20/43ef0a21/c5d5220c and production/final records;
it requires Claude before any acceptance or archival claim. RP-332 exact
cc62cea8..5fbf4cff, audit fc911784..cc62cea8, runtime a42406f0..fc911784
and all preceding ranges remain separately owed. No boxes/status/archive/
mint/push/release changes; all full-nine-tier 1.0 gates remain in force.

## 2026-10-07 — GS3-A3 semantic narrow-layout proof predeclared

Resume clean bddfc58e. RP-329 exact review range 5fbf4cff..bddfc58e remains
pending Claude, not merged into this distinct layout range. Accepted GS3
requires a real definition list below 30rem, with every constituency/axis
header retained beside its value; RP-330 records the current CSS-only table.

Three native browser declarations, each Chromium/WebKit: narrow initial
mount at 320 px; live 1280→320→479→480→1280→320 px changes without remount;
narrow authoritative snapshot refresh followed by widening. Public fixtures
use all existing IDs with distinct safe values/bands to prevent a wrong
row/value association from passing on repeated 0/50. Decode each fixture
before mount/update. Derive expected header text from the existing canonical
presentation/copy data, not mechanical ID splitting or authored new text.

Narrow oracle: exactly eleven unique dt/dd pairs in actual dl elements,
each with complete term text (constituency + axis; p(doom) separately),
exact numeric/band spans, and native meter min/max/value whose aria-labelledby
resolves to its own dt. No hidden duplicate table/list or CSS-only pseudo-label
may satisfy the assertion. Wide control: original one table / five rows / ten
cells, eleven total native meters and no duplicate content. Keep nav focus,
zero emitted intents, full mounted-page horizontal bounds and serious/
critical axe floor at each layout. Retain all existing browser assertions.
320 CSS px is a viewport population, not a claim of actual 400% zoom or AT.

Before production, run typed-clean selected native cases and record the
current semantic failures. If established, separately record the bounded
implementation step: MetersSurface only, responsive native semantic markup
using existing data/copy/tokens, live media change and cleanup. No host,
runtime, wire, catalog, producer, kernel, balance, copy or CI change.
Predeclare later exact label, numeric/band association, breakpoint and resize
listener faults before executing them; actual assertions must fail, source
restored exactly. Final full native Garage, types/client/build/boundaries
and copy/topology remain required. No Docker run until capacity repair.
New whole range starts after bddfc58e and ends after its final records;
Claude designated review remains mandatory. No boxes/status/archive/mint/
release promotion or substitution for RP-331/333/all-engine/AT/full 1.0.

## 2026-10-07 — GS3-A3 semantic baseline established before production

Predeclared d9a8ca27. Three native declarations / six Chromium-WebKit
executions use distinct decoder-admitted public values and the existing
canonical presentation/copy. Narrow cases require real dt/dd associations,
native meter values/names, exact numeric/band spans, one rendered population;
wide checks retain the table/headers and exact per-cell values. Live resizing
and refresh cases also retain navigation focus, zero intents, page bounds
and the axe floor. Existing declarations/assertions are unchanged.

First types 41227 and refined wide-value types 48582 pass with zero errors/
warnings. Initial native 28005 exited 2: four narrow semantic failures and
two *instrumentation* failures at the healthy wide control. The oracle read
the raw HTML value attribute; Svelte legitimately updates the native meter
property, and a zero-valued meter need not have that attribute. Corrected
the oracle to exact meter.min/max/value properties (stronger actual native
outcome), not a permissive null/default fallback. This initial run is not the
claimed six-case semantic baseline, and no production change used it.

Corrected root types 73951 passes, zero errors/warnings. Native selected
root population 50408 exits 2: six actual semantic failures / 246 unselected
executions, two failing engine files, 1.79 s. The wide healthy control now
passes; all three scenarios in both engines fail because a table remains
at narrow width instead of the accepted definition list. No collection,
module, fixture, parser or native-value error substitutes for that finding.
The later snapshot/breakpoint assertions will execute once the baseline
defect is repaired; their discrimination must still be demonstrated.

This checkpoint is test/planning only; MetersSurface is byte-unchanged.
RP-330 ledger/plan/queue updated. Next separate bounded production step and
predeclared label/value/breakpoint/listener probes, final full native/client/
copy/topology gates. Complete new layout span begins after bddfc58e and
includes this test-first record through its eventual production/final edge.
Claude review is separately owed; RP-329 exact 5fbf4cff..bddfc58e and all
earlier ranges remain pending. No checkbox/status/archive/mint/AT/all-engine/
persisted/release claim; no Docker run or cleanup, full 1.0 goal active.

## 2026-10-07 — GS3-A3 bounded repair and severing step

After test-first 09655d1e / corrected six-failure baseline 50408, implement
only MetersSurface.svelte: a real trust definition list and separate p(doom)
definition row below 30rem, the existing table at/above it, exactly one
rendered population. Reuse current presentation/copy/committed values and
shared meter/value/band markup. Every narrow dt names its native meter via
aria-labelledby; no new copy or invented data. A mount-owned matchMedia
`(width < 30rem)` listener handles resizing and is removed on unmount. This
is layout observation, not another reduced-motion listener or gameplay input.
Retain RP-332 error containment, existing full-page heading/nav behavior.

Predeclare independent executable faults before probes: remove axis text
from the narrow term; omit its aria-labelledby binding; substitute a zero
native meter value; change the breakpoint to 20rem; omit the live change
listener while retaining its initial read; suppress band text. Actual named
association/value/breakpoint/live-resize assertions must fail, not parsing
or type errors. Restore the exact healthy component SHA after each probe.
Do not loosen existing tests, timeouts, copy or budgets. Final root types/
client/build/boundaries, full native Garage and copy/topology gates remain
mandatory; no green claim before terminal results. Whole range after
bddfc58e through final records requires Claude, no archival or acceptance
promotion. RP-329 5fbf4cff..bddfc58e stays separate.

## 2026-10-07 — GS3-A3 bounded layout and independent probes executed

Production is confined to MetersSurface.svelte under 639add6e. Its one
shared value/band snippet serves either the original wide table or actual
narrow definition rows. Narrow terms retain constituency/axis (and a separate
p(doom) row) and name the native meter via aria-labelledby. No derived meter
values, player copy, host/runtime/wire/catalog/kernel/balance/CI change.
Live strict-below-30rem layout observation is mount-owned and cleans up on
unmount; RP-332 error guard and diagnostic behavior remain intact.

Types 96517 passes with zero errors/warnings. First healthy native attempt
70487 exited 2: four passed, two Chromium resize cases observed the old
layout before its native media-query event. WebKit passed. The test driver's
viewport RPC and four timer flushes are not a rendering-time event barrier.
Added one actual requestAnimationFrame observation plus the existing flush
after viewport changes; no threshold, assertion, timeout or budget loosened,
no production timing workaround. Corrected selected run 33397 passes six /
246 unselected executions and the unchanged isolated performance test.
The listener severing below still fails after the real-frame barrier, proving
the barrier does not disguise a frozen layout. The first attempt is disclosed,
not counted as green evidence or six semantic baseline failures.

All six predeclared independent faults compile/execute in the root selected
Chromium/WebKit population; each command exits 2, with 246 unselected cases.

| Transient component fault | Run | Failed / passed | Actual failed oracle |
|---|---|---|---|
| Omit axis text from narrow term | 61336 | 6 / 0 | Exact “Employees Grievance” receives “Employees” |
| Remove aria-labelledby | 89505 | 6 / 0 | Name reference resolves to null instead of its own term |
| Force native value to zero | 82356 | 6 / 0 | Exact native value 7 receives 0; text remaining correct does not mask it |
| Use 20rem instead of 30rem | 90890 | 6 / 0 | Narrow population still has a table |
| Omit change listener, retain initial read | 53381 | 4 / 2 | Live narrow/wide transitions freeze; initial narrow mount still passes |
| Suppress band text | 89952 | 6 / 0 | Exact value/band spans receive empty band instead of Low |

Healthy component SHA-256 verified after *each* restoration:
5ebc07105fb33e745935dfe5a8f53b3c378cdd081b603ebce0eba103b26f4a95.
No mutation was committed or left in source, and no source edit occurred
during a matching live check. Existing assertions remain unchanged. Final
full root client/native/build/boundary/copy/topology outcomes follow only
after their terminal results. No Go/SQL changes or fresh persisted claim;
RP-329's earlier cold Go evidence is not re-labelled as a new layout run.

Review by: Codex (self first filter only). Recorded by: Codex. Complete
layout range begins after bddfc58e through production/final records and
requires Claude separately from RP-329 5fbf4cff..bddfc58e and earlier spans.
No acceptance boxes, RFC/lifecycle/archive/mint/push/release status changed.
RP-331/333/Docker/Firefox/AT/full shared fixture/full nine-tier 1.0 remain open.

## 2026-10-07 — GS3-A3 final current-source gates

Root types/client/build/boundaries 34720 exits 0: zero type errors/warnings;
9,814 passed / 449 explicit browser skips, 105 passing files / 22 skipped;
client population 6.35 s. Build: 214 modules, index-eU1nW9jC.js,
index-DaRqgLww.css, unchanged prediction.worker-MqspU_iu.js. Boundary scan:
14 shell, 8 UI and 22 Game UI components.

Full current native Garage Chromium/WebKit 83278 exits 0: 252 / 252
executions, two passing files, 77.96 s (including the real 60-second idle
scenario). Existing RP-332 error cases and all previous assertions still
pass. Chained unchanged isolated performance: one pass / 22 unselected,
323 ms test time; RP-333 null-feature limitation remains. Terminal result
lines retained; individual worker HTTP lines omitted from displayed summary,
not substituted for a claim of complete reviewed network behavior.

Root copy/content-manifest/topology 38297 exits 0: 658 keys with unchanged
SHA-256 a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611 existing orphan warnings; content manifest valid; all 13 topology
negative controls rejected. This static topology pass is not full Linux CI.
Canonical docs also qualify the former undated “all three browsers” reflow
claim: changed-source evidence is Chromium/WebKit viewport/DOM only, not
actual Firefox/400% zoom/AT acceptance.

Healthy production component SHA remains
5ebc07105fb33e745935dfe5a8f53b3c378cdd081b603ebce0eba103b26f4a95.
No transient fault or unrelated source change remains. Next accepted work:
RP-333 populated Garage Desk performance population under unchanged budgets,
then the remaining exact source guards. RP-331 persisted/default-player
proof remains held until Docker capacity is repaired/rechecked. No new
Docker population, resource cleanup, DB/Go change or fresh SQL claim.

Review by: Codex (self first filter only). Recorded by: Codex. The complete
layout span after bddfc58e includes d9a8ca27/09655d1e/639add6e, later native
driver refinement, production/docs and final tracking; Claude must review
that entire range before acceptance or archival. RP-329 exact 5fbf4cff..
bddfc58e and all earlier review ranges remain separately owed. No box,
RFC/lifecycle/archive/mint/push/release promotion; full 1.0 goal active.

## 2026-10-07 — GS3-A3 record edge and next authorized work

Production `a4443495` and this final tracking edge synchronize RP-330,
the all-gate inventory, current plan, CURRENT-STATE, full 1.0 board and
execution queue. No checkbox/lifecycle state changed. Complete designated
review range starts bddfc58e exclusive THROUGH this record commit, including
predeclarations, tests, native-frame driver refinement, production/docs and
all record edits. Review by: Codex (self first filter only). Recorded by:
Codex. Claude supplies the mandatory separate verdict; no self-archival.
Prior exact RP-329 5fbf4cff..bddfc58e, RP-332 cc62cea8..5fbf4cff, audit
fc911784..cc62cea8 and earlier ranges remain owed independently.

Next accepted agent-side work is RP-333 populated Garage Desk performance
under the existing observable budgets, not optimizations or CI changes.
Preserve the existing null-feature guard; predeclare the added population
and its failures. Real SQL/composed RP-331 requires repaired/rechecked Docker
capacity, not another run into a full filesystem. All other author/body/
engine/actual zoom/AT/default-player/platform/privacy/numeric/review/full-
nine-tier 1.0 gates remain live. Goal active/progress; no release shortcut.

## 2026-10-07 — RP-333 population and RP-334 instrument audit predeclared

Resume clean 127eb052. Previous turn is progress: RP-329 exact complete-ID
range 5fbf4cff..bddfc58e and RP-330 semantic-layout range bddfc58e..127eb052
are committed with negative probes and current tracking; neither is
designated-approved. All earlier independent verdict obligations remain.

Accepted Garage overall AC7 consumes the existing Game UI GU-C8 observable
scenario/budget unchanged: 1280×720, 1,200 inputs representing 60 simulated
seconds at 20 Hz, at most 600 hot Amount commits and no long task >200 ms.
The separate manual 4× CPU / dropped-frame / real-duration profile is not
replaced. GS5/GS6 Desk population here means the always-mounted opportunity
region with pending opportunity, all four effect buffs/combo, provisioned
count and reason text, owned-upgrade text, existing free shelf and grown
Garage feature nav. Fiscal/Achievements/Meters are separate tabs, not fake
Desk mounts. Later Adoption/Cosmetics/AxisStack/Reputation/T2 populations
remain separately scoped; this is not the complete current release profile.

Before adding a green populated measurement, audit two instrument risks
recorded as candidate RP-334. Run the unchanged root isolated performance
lane. Then, one transient test-file fault at a time: omit its fixtureSnapshot
loop updates (leave declared input label and budget intact); add an actual
synchronous final main-thread task lasting twice the ruled long-task ceiling
before its final tick/disconnect. Independently record actual busy duration
and the driver's collected counters. Record a survivor honestly: it rejects
the original instrument as proof for that property, not the product.
No cache/timeout/budget/observer/mock-clock workaround. Restore original
test source exactly between probes; no edit while a matching check is live.
Original test SHA-256 a585474a5e441ed4030776b08ba686b2ee26100b1d30cc0739376b77cf5f6466;
budget source 8a7fc92f255152f784ef421da886ae3a6d860fd2e8afebbf94ca4d1eeb11217e.

If a risk fires, predeclare a test-only instrument correction before acting:
actual completed-input census, nonzero native visible activity and exact
terminal Amount, mandatory observer support, rendering-time record drain
and cleanup, then actual behavioral seeds. Preserve the old scenario byte-
identically as a regression; do not infer it covered the new population.
Add the populated scenario under the existing isolated selector, no workflow,
Make target/topology, budget, numeric/formatter/scheduler/host/region/product
change. Decoder-admit the public fixture; assert the named Garage additions
remain present through every input, with zero gameplay intents. Report measured
counts/long tasks, exclusions/support/completion visibly. Run root types/full
client/build/boundaries and the isolated performance lane; retain broader
browser/SQL/release evidence boundaries and pending full Linux CI.

New evidence/test/planning span starts after 127eb052 through its final
records and requires Claude. No optimization, owner copy, catalog/kernel/
balance/schema/CI change, checkbox/status/archive/mint/push/release claim.
Docker capacity, RP-331/Firefox/AT/actual zoom/manual profile/author/body/
numeric/privacy/platform/default-player/full-nine-tier 1.0 gates remain.

## 2026-10-07 — RP-334 native findings and bounded test-only correction

Fresh local checkout remains clean 04b10f97 before the probes; no newer Claude
commit is present locally. Existing root isolated baseline 38175 passes one /
22 unselected. Omitted measured fixture updates 71274 also passes one; final
actual busy task at twice the unchanged 200 ms ceiling 70295 also passes one.
The original validator cannot reject either fault. Browser console output on
passing tests was not displayed, so separate observation-only reruns deliberately
throw AFTER the validator returns and cleanup completes: 31335 reports zero
visible commits, claimed inputs 1,200 and terminal 100; 98149 reports 400 ms
actual synchronous duration, 217 commits, 1,200 claimed inputs and longest task
0. Their exits 2 are reporting instrumentation, not discriminating gate exits.
Both faults/restoration use apply_patch; no source edit during a live check.
Restored original test and budget SHA match the preceding predeclaration.

Correction to my population wording: Lucky is an immediate payout, NOT a buff
(docs/active-play and the pinned opportunities artifact). Do not invent a Lucky
buff merely because the decoder admits its ID. The intended four effect rows
are pending Lucky plus building/click/production buff rows, valid selected
building target, sorted UUIDv7 instances and combo. This supersedes my earlier
"all four effect buffs" sketch, not the RFC or any owner-authored text.

Predeclare the implementation now, before source changes: add one independent
populated scenario under the EXISTING isolated selector, preserving the complete
old scenario byte-identically. Construct a decoder-admitted public snapshot
using current presentation/catalog IDs, pending Lucky/three buffs/combo, capped
provisioning, owned upgrade, free shelf and supported achievements/Fiscal/Meters
nav facts with corresponding arms. These are separate tabs, not mounts. Pet,
Pitch without its runtime port, Adoption/Cosmetics/AxisStack/Reputation/T2 and
later current-release populations remain explicit exclusions, not all-release
performance claims. Select the CASH resource Amount, not the earlier static
combo Amount. Keep runtime.current coherent with the measured samples so a
background recovery sample cannot erase the population.

The new test-only driver must count inputs only after real delivery, require
nonzero native visible commits, exact terminal formatted cash, connected hot
output and named-region/copy census after EVERY input, and zero intents. Native
MutationObserver and Long Tasks support are mandatory. Flush native queued
records after real rendering/task turns before disconnect; teardown in finally.
Retain every original numeric ceiling and 60-simulated-second interpretation.
Use the existing runner annotation attachment API for an inspectable JSON
observation; no new reporter/workflow/Make target/config/production plumbing.

Predeclared faults, one at a time against the completed new driver: omit input
delivery; hold hot Amount rendering fixed while inputs continue; inject a final
actual 400 ms task; remove Long Tasks availability; remove an opportunity buff
from the measured population after setup. Each must execute and fail the named
activity/terminal/budget/support/census oracle, not fail compilation. Restore
healthy test source exactly between probes. The old guard may survive and that
survival is retained, not covered up by changing its assertion or population.
Root isolated lane, types/full client/build/boundaries then final records.
Full Linux browser/SQL/manual profile/AT/review holds remain unchanged.

Review by: Codex (self first filter only). Recorded by: Codex. Complete new
range begins 127eb052 exclusive through final records; Claude verdict owed.
No production/numeric/catalog/kernel/balance/copy/CI/schema/owner-text change,
acceptance checkbox, lifecycle, archival, push or release promotion authorized.

## 2026-10-07 — RP-334 probe results and terminal-oracle supplement predeclared

Test-only implementation preserves the original scenario block byte-identically
to 127eb052 (root Node comparison executed) and the unchanged budget SHA.
Initial type run 29968 caught my wrong era literal; initial native 55225 caught
my unsorted fixture facts. Both are fixture/instrument mistakes, not product
defects or valid discrimination. Corrected fixture uses eraForSnapshot and
byte-sorted facts; types 97386 pass zero errors/warnings; native 47270 passes
both original and new scenario / 22 unselected.

Healthy test SHA c61f3952e3e009cec9a5ec39777572f59702fc66c6271714c05702f05d7e5aff.
Five actual root Chromium fault runs exit 2, each with original guard passing:
33596 no delivery -> inputs 0 / visible commits 0; 32177 fixed cash with 1,200
delivered updates -> visible commits 0; 2249 final real busy task -> longest
task 400 vs maximum 200; 49004 unavailable-support sensor -> invalid native
observer observation before mount; 80927 remove building buff from subsequent
sample -> exact population census fails on first delivery (inputs 1, completed
false). The latter is decoder-admitted; schema acceptance alone cannot satisfy
the render population. Healthy SHA checked after each restoration, no live-source
edits or committed transient fault. These are instrument/population probes,
not evidence of measured gameplay performance regressions.

Runner annotation retains the full JSON in its MESSAGE as well as requesting
an attachment. The local runner did not persist a JSON attachment file, and its
default success reporter did not display the message; do not cite a nonexistent
disk artifact. Healthy reporting-only rerun 92955 deliberately throws after
all validation/cleanup, exposing completed=true, 1,200 actual inputs / 1,200
region checks, 217 commits, longest task 0, terminal "1.30 K", support true,
zero intents and exclusions. That exit 2 is a reporting diagnostic, not a red
performance gate. Diagnostic restored exactly. Final native 85140 exits 0,
two passes /22 unselected; root static 34127 exits 0: types 0/0, client
9,814 pass/450 explicit browser skips, unchanged 214-module production build
and 14/8/22 boundaries. No production byte moved.

Before final closeout, separately predeclare one additional compiling fixture
fault: corrupt ONLY the final delivered cash to a different canonical amount
while retaining real prior visible rendering and all 1,200 inputs/regions.
The terminal-cash oracle must fail, not just the positive-activity floor.
Restore the same healthy SHA and rerun the isolated lane. No new assertion,
budget/threshold, config/Make/CI or production change authorized.

## 2026-10-07 — RP-333/334 final bounded evidence and current queue

Additional terminal-cash fault 40172 exits 2 at the exact terminal oracle:
10.0 K instead of 1.30 K despite 1,200 actual inputs/region checks and 218
visible commits. Positive activity cannot mask wrong terminal data. Restored
healthy test SHA matches c61f3952e3e009cec9a5ec39777572f59702fc66c6271714c05702f05d7e5aff;
budget SHA remains 8a7fc92f255152f784ef421da886ae3a6d860fd2e8afebbf94ca4d1eeb11217e.
Final healthy isolated 51810 exits 0: two passing scenarios /22 unselected,
1.83 s total, 837 ms tests. Original test block byte-identical comparison
already executed. No transient/reporting fault or production diff remains.

Final source-static 34127 (zero types/warnings, 9,814 unit passes /450 browser
declarations skipped, unchanged build/boundaries) and copy/content/topology
13646 exit 0: 658 copy keys / unchanged catalog hash /611 existing orphan
warnings, content manifest valid, all 13 topology negative controls rejected.
The topology result is structural CI evidence, NOT a rerun of red/held Linux
three-engine CI. No DB, changed-host composed, Go behavior, Firefox/WebKit
performance, actual 60-second/4× manual profile, AT or release claim.

Canonical docs, backlog, acceptance inventory, plan, CURRENT-STATE, full 1.0
board/checkpoint log and execution queue synchronized in this implementation
range. New driver locally corrects RP-334; RP-333/full AC7 stays partial:
GS4/pet/later Desk/full current release population, actual player/manual
profile and designated review remain. The bounded population and explicitly
reported exclusions cannot be cited as proof for those other populations.

Review by: Codex (self first filter only). Recorded by: Codex. Entire new
range begins 127eb052 exclusive THROUGH this final implementation/record
commit, including 04b10f97 and fb302c78, all tests/docs/tracking. Claude must
supply its designated verdict independently of layout bddfc58e..127eb052,
complete-ID 5fbf4cff..bddfc58e, RP-332 cc62cea8..5fbf4cff, audit
fc911784..cc62cea8, runtime a42406f0..fc911784 and every earlier span.
No review relay sent, checkbox/RFC/lifecycle/archive/mint/push/release action.

Next accepted native/source work: source/boundary guards with predeclared
real negative cases; keep legitimate later Cosmetic Shop intents distinct
from the forbidden legacy shelf. Existing boundary scanner contains seeded
copy/style/network fixtures, but overall AC4 still needs its current-source
fault bound to exact executed evidence. RP-331/SQL/default-player/Docker,
author/body/engine/actual zoom/AT/privacy/platform/numeric/full-nine-tier 1.0
holds remain. Active goal progress, not completion or a shortened release path.

## 2026-10-07 — overall AC4 current-source boundary probes predeclared

Fresh clean d3ce0f76. Previous goal turn is progress: RP-333/334 test-only
measurement and six actual failure cases committed with current tracking.
New span starts d3ce0f76 exclusive; prior exact 127eb052..d3ce0f76 is
independently owed to Claude, not re-reviewed or archived here.

Accepted Garage overall AC4 requires current mounted components to pass the
existing boundary and a literal seed to fail. Inspecting the existing scanner
shows internal seeded fixtures already present; this work adds executed
current-source evidence, not another inferred missing validator. No production
repair is preauthorized by a diagnostic survivor; record it first if found.

Run healthy root make verify-client-boundary. Then one transient source fault
at a time in OpportunityRegion.svelte: (1) the existing scanner's unregistered
paragraph text, (2) one raw fetch in a syntactically valid unused function,
(3) a type-only import from the real shell/runtime module, (4) literal color
instead of governed token. Each seed must parse as valid Svelte and root
boundary must fail its named policy, not a syntax error. No live browser or
other matching source check during edits. Restore exact source SHA between
every run and finish with healthy root boundary/typecheck. Do not commit any
seed, alter owner copy, budgets, scanner, production behavior, CI or Make.

Original OpportunityRegion SHA-256:
5ee8223a531430fb1256e53f73a2a05e15020ccfef81dd8d7d4a70f2b18e6516.
Original scanner SHA-256:
ed06b3583ea48846c9b0134130f14fe3ab7804fdf9804ba799f15b658803ed32.
This only binds AC4's source policy to executed evidence. It does not prove
runtime behavior, all acceptance criteria, full browser/SQL/AT/default-player
populations or release readiness. GS6-A2 cosmetic source authority is separate:
the later accepted Shop's legitimate commands must not be incorrectly banned.
No checkbox/lifecycle/archive/push/release change; no review relay authorized.

## 2026-10-07 — AC4 current-source results; GS6-A2 baseline predeclared

Healthy boundary 2a5deeee baseline passes 14 shell/8 UI/22 Game UI files.
Four independent valid-Svelte seeds fail root make verify-client-boundary,
each exit 2: 767166 text -> Copy pipeline policy; d42dda raw fetch -> raw
network policy; d76e00 real shell/runtime type import -> authoritative/runtime
import policy; cbe6c1 literal color -> governed token policy. Every seed
first passed the actual Svelte parser. Source SHA verified after each
restoration; scanner never changed. Final root 35939 boundary/typecheck exits
0 with zero errors/warnings. These are source-policy witnesses, not runtime
or cross-party acceptance. No production diff or transient seed remains.

Next accepted GS6-A2 baseline, before any guard repair: inspect current
mounted component source against the actual production cosmetic intent
constants/isCosmeticIntent dispatch. Current legitimate host calls are
acquire_cosmetic/equip_cosmetic/unequip_cosmetic; the public registry's generic
intent envelope is not itself an authority for their kind strings. Existing
boundary scanner has no intent-kind check. Insert one syntactically valid
registered-copy button in the legacy free shelf invoking existing act with
kind buy_horse_armor and cosmetic_id horse_armor; no gameplay/browser run,
no real command submitted. Existing boundary should not reject it merely on
copy/style/import/network grounds. Record an actual survivor as missing
source-guard coverage, not a real unregistered button present at healthy HEAD.
Restore GameUIApp exactly and record its initial SHA before the seed.

If confirmed, predeclare a bounded verification-tool/source-test correction
under GS6-A2 before implementing. It must derive kinds from the production
authority, not a new manually maintained list, parse actual component AST,
retain registered free-shop commands, reject unknown/dynamic cosmetic kind
evidence loudly and demonstrate its own seeded failures. No product, schema,
balance/copy/kernel or CI topology/Make change follows from this probe.
All designated review, author/body, Docker, full performance/accessibility/
default-player/full-nine-tier release holds remain; no status promotion.

## 2026-10-07 — RP-335 source-guard finding and repair predeclared

Valid registered-copy fake shelf command parses and root boundary passes
(run 9f47eb). This is an executed missing GS6-A2 source-kind guard, not a fake
command shipping at HEAD. Host restored to SHA
9ea86275b688712047703e2e6b794fcf5d6b5b61706d874b51ee13462459223a.
Existing verify-cosmetic-boundary enforces package dependencies/imports and
verify-no-payment enforces dependency/SDK rules; neither inspects command
kinds. Do not substitute either positive check for the absent source witness.

Predeclare bounded verification-only repair under accepted GS6-A2. Existing
root verify-client-boundary will invoke a standard-library-only Go AST
extractor in client/tools (not a service, generator contract or production
package). Derive string constants selected by actual isCosmeticIntent and
require their presence in ParseIntent's request.Kind switch. Unsupported
predicate/constant/decoder shapes, missing/duplicate authority or empty sets
fail loudly. No manually maintained accepted-kind table, production edit or
new network/dependency requirement. Existing client verification already uses
the installed Go toolchain for cosmetic/achievement dependency boundaries;
the extractor runs from root with the existing Make-exported cache.

Use existing Svelte AST traversal to verify explicit cosmetic command objects,
including local declaration objects, in script and template expressions.
Unregistered or nonliteral kinds with cosmetic_id fail; ambiguous properties/
spreads cannot masquerade as a proved command. Record the supported syntax and
limits rather than claim arbitrary whole-program analysis. Existing legitimate
acquire/equip/unequip callbacks must continue to pass. Helper and checker carry
positive/negative source fixtures, including comment/string lookalikes,
unknown/duplicate constants, decoder omission, dynamic kinds and opaque spread
envelopes. No payment or registered cosmetic ID allowlist invented here.

Actual source probes after implementation: repeat fake buy_horse_armor button;
replace a legitimate cosmetic kind with an unregistered one; replace that
kind with a dynamic expression; sever one required production decoder case
in a transient source-only probe. Each must parse/execute and fail the named
source authority check, not compilation; restore exact source hashes between
probes. A checker bypass must be caught by its built-in negative fixtures.
Final root boundary/types/client/build, verification-helper behavior, cold
existing cosmetic parser tests and static CI topology; no full Linux/SQL claim.

New verification/evidence range remains d3ce0f76 exclusive through final
records, including AC4 predeclaration/results; mandatory Claude verdict
separate from 127eb052..d3ce0f76 and all prior spans. No product behavior,
owner text, schema/catalog/kernel/balance/copy/Make/workflow/CI topology,
checkbox/lifecycle/archive/mint/push/release change authorized. Proper full
nine-tier 1.0, author/body/AT/engine/default-player/platform/privacy/numeric
and Docker capacity gates remain. A source-test improvement is not release
acceptance or authority to reconcile somebody else's stale RFC body.

## 2026-10-07 — RP-335 executed source guard and local first filter

Review by: Codex (implementer self/first filter, NOT designated).
Recorded by: Codex. Complete evidence/tool/docs/tracking span begins after
d3ce0f76, including 2a5deeee/d2b4b5a4/da2bc019; final commit is this record's
containing commit. Mandatory Claude review must cover that whole range,
separately from performance 127eb052..d3ce0f76 and all previous exact spans.
No designated verdict, acceptance checkbox or archival eligibility claimed.

Verification-only implementation adds client/tools/cosmetic-intent-kinds.go
and extends the existing Svelte boundary scanner. Go AST reads actual string
constants selected by isCosmeticIntent and checks one occurrence of each in
ParseIntent's request.Kind cases. Unsupported/missing/duplicate authority
fails; stdlib only, no service import, network call, dependency download or
new accepted-kind table. Two positives (including non-code lookalikes) and
ten negatives run inside the helper. Svelte checks explicit cosmetic objects
and actual host dispatch envelopes, with verified current Exit-plan/Garden
source shapes and eleven negative fixtures. Ordinary local callback helpers
are not incorrectly treated as the host dispatcher. Source syntax only:
not arbitrary alias/control-flow/whole-program analysis, proof of forwarding
semantics, cosmetic eligibility, persisted receipts or full Shop acceptance.

Instrument mistakes disclosed: 66b8f5 rejected the actual host Exit/Garden
wrappers; 9a6a79 then confused an unrelated GardenSurface act(callback) with
the host. Both exited 2 and were checker-binding errors, not product defects.
Explicit host/wrapper bindings corrected them; no gameplay source edit or
weakened unknown/dynamic-cosmetic-kind rule. Corrected healthy dbb022 exits 0.

Actual source probes, one at a time and no gameplay/runtime execution:

| Run | Fault | Executed result |
|---|---|---|
| b9ca66 | Repeat valid registered-copy fake legacy-shelf button | Svelte parse passes; root boundary exits 2: buy_horse_armor outside production authority. |
| 0e5ca7 | Real onAcquire kind replaced by buy_horse_armor | Svelte parse passes; boundary exits 2 on the same authority check. |
| c1f627 | Real onAcquire kind replaced by dynamic id | Svelte parse passes; boundary exits 2: dynamic kind cannot prove authority. |
| 33e23d | Remove Acquire from actual ParseIntent compound case | Go AST parsing executes; boundary exits 2: required decoder occurrence is zero, NOT a compile/type failure. |
| 33b5e1 | Bypass scanner's unregistered-kind rejection | Boundary exits 2: built-in fake-command negative survives and invalidates the checker. |

Each source restored before the next probe. Final SHA-256 identities:
GameUIApp 9ea86275b688712047703e2e6b794fcf5d6b5b61706d874b51ee13462459223a;
production intents.go 5a284dcb009b804feed72c2af472dd58e4a0314d7f792255ef02f6f9af462975;
scanner 1cded97a8ac6c23abc3b960ed5480a86c5d0d0e1ee3a031d7b46ba09ffaf8313;
helper 7d5702a93d54862c96fcfe60047c905d7533d4e1da2b335607d9d98372d971d4.
No fake button or production decoder fault is committed. AC4's separate four
actual-source policy seeds/results remain in the preceding entry.

Final healthy verification, terminal results, not cached/inferred summaries:

- 40143: root boundary/types/client/build/cosmetic-boundary/no-payment exits 0.
  14 shell/8 UI/22 Game UI files; actual three command kinds, ten Go/eleven
  Svelte negatives rejected. Types zero errors/warnings; client 9,814 passes /
  450 explicit browser skips, 105 files pass/22 skip. Unchanged 214-module
  build. Cosmetic isolation 76 stdlib dependencies/two client files/22
  negatives; payment allowlist/SDK gate six negatives/two near misses.
- 5365: root make test-go GO_PACKAGES=./production with cold -count=1 -v
  -run TestCosmetic exits 0. Eight top-level passes plus two child reports;
  TestCosmeticCorpus drives actual ParseIntent/canonical payload/Founder
  transition and pinned-byte comparisons. Two SQL cases explicitly SKIP
  because TEST_DATABASE_URL is unset; NOT integration green. No fixture update.
- 4c146e: verify-ci-topology exits 0 with thirteen negatives rejected;
  this static check is NOT a full Linux browser/hosted CI pass.
- 45940: isolated native Chromium performance exits 0, two tests pass /
  twenty-two unselected, including the existing populated supplement.
- 880d8f: root make vet with GO_PACKAGES='../client/tools/cosmetic-intent-kinds.go'
  exits 0; the new stdlib helper itself, not only the service, is vetted.

Docs/ledger/acceptance inventory/current board/plan/queue/checkpoint log are
reconciled in this implementation range; diff check passes. No product/API/
copy/catalog/kernel/balance/Make/workflow/CI topology byte changed. Full Linux
browser remains RED/held by last measured Docker overlay 0 free/100%; no
SQL/container rerun or unowned cleanup. RP-132/313/318 author/body, RP-331
persisted DOM, RP-333 complete performance, Firefox/actual zoom/AT/default
player/privacy/platform/numeric/full-nine-tier 1.0 and independent review
remain. Next accepted safe work: predeclare GS6-A3 whole-Desk 320 px pinned
before/after comparison, not a shortcut release claim. No push/archive/mint.

## 2026-10-07 — GS6-A3 paired whole-Desk measurement predeclared

Accepted GS6-A3 requires fixture measurement before/after the Desk additions
without claiming ownership of accessibility's historical 647 px regression.
Existing combined native reflow test measures the after fixture only. Source
history acf43132 introduced provisioned/cap and owned-upgrade text; its parent
lacked those lines and also differs in many unrelated current components.
Do not substitute an old source build or inferred historical dimensions here.

Add a test-only paired current-component fixture measurement: same decoder-
admitted v4 state, same 1995 theme, chrome/nav/resources/manual/opportunity/
free shelf/curtain/generator/upgrade content; before has provisioned=0 and
upgrade owned=false, after has provisioned=cap=3 and upgrade owned=true.
The cap itself remains identical in both public fixtures; no bogus private
field or producer-generated/shared artifact is claimed. Revision alone also
advances to exercise the actual fixture refresh. Before/after mean absence/
presence of the GS6 text in current source, NOT the historical executable.
Named census must prove those states actually rendered; omitting all added
text cannot silently turn both arms into the same passing measurement.

At native 320 CSS px, record document/main/Desk/chrome widths, all descendant
bounding rectangles and own scroll/client widths; exclude only genuinely
zero-layout nodes. Both populations must fit the viewport (existing <=1 px
rounding allowance), after must not widen measured extents beyond before,
and no overflow:hidden/clip masking of content may fake containment. Check
exact provisioning/reason/owned text, retained visible shelf/curtain/manual/
opportunity/chrome/nav, retained manual focus and zero runtime intents.
Restore wide viewport and unmount in finally. Run Chromium and WebKit through
the root native lane, no Docker; Firefox/full Linux CI/actual 400% zoom/AT/
real service/later Desk/release/whole accessibility acceptance remain distinct.

Demonstrate actual failures one at a time: force added provision text to a
600 px minimum width; hide it with display:none; omit its cap reason; set
Desk fixed inline size 647 px; mask a wide added child with overflow:hidden;
and bypass the paired input update. Each should fail the appropriate actual
source/render/census/layout oracle, not compile errors. Restore exact hashes
between probes, never edit source during a matching live process. Record
instrument mistakes/invalid probes separately. Product changes only if a
fresh real healthy-source failure is established and separately predeclared.

No product/copy/balance/CI/Make/authority change, checkbox or lifecycle move.
Whole new span after 54c8c8da through subsequent records requires Claude
independently of source guard d3ce0f76..54c8c8da, performance
127eb052..d3ce0f76 and every previous exact span. Full-nine-tier goal active.

## 2026-10-07 — GS6-A3 paired Desk results and local first filter

Review by: Codex (implementer self/first filter, NOT designated).
Recorded by: Codex. Complete new range starts after 54c8c8da, includes
757f1ca2 and the test/docs/tracking commit containing this entry. Claude
must independently review the whole span; source guard d3ce0f76..54c8c8da,
performance 127eb052..d3ce0f76 and earlier spans remain separately owed.

One test-only native scenario mounts the complete current 1995 Desk with
decoder-admitted paired fixtures. Only provisioned=0→3, owned=false→true
and refresh revision change; cap and every other public field remain the
same. Required census proves provision/count-cap/owned text absent then
present/visible, shelf and curtain visible without a purchase button,
manual/opportunity/chrome/nav retained, manual focus unchanged, zero intents.
Measures document/main/Desk/chrome widths plus all visible descendant
rectangles and own scroll widths; negative left extents, overflow and clipped
ancestors fail. No selector-only claim that both states rendered.

Initial native 011b3f exits 2 in both engines because my unqualified cards
census counted an opportunity card as a generator/upgrade. This was an
instrument selector mistake, not a product failure; exact existing section
bindings corrected it without relaxing layout/text/cap checks. Healthy
0379e8 then passes both native cases and both isolated performance cases.

Six independently executed actual faults, two failures each, terminal exit 2:

| Run | Seed | Discriminating result |
|---|---|---|
| 702a71 | Actual provision text min-inline-size 600 px | Baseline fits; after document scroll width 652 and descendant extents fail. |
| 82bcb9 | Actual provision text display:none | Actual visibility oracle fails, not merely text presence. |
| c3a36d | Actual provision-cap reason withheld | After exact reason census expects one, gets zero. |
| 467258 | Actual Desk inline-size 647 px | Whole-page baseline fails (page 699 including padding/borders), NOT a replay of the old historical 647 px measurement. |
| 5d2793 | Wide actual provision text hidden by its card overflow | Page masking cannot hide descendant/own-scroll failure; card scroll 632, child right extent 652. |
| ba901e | Driver omits after fixture delivery | After exact populated census fails; cannot report the absent arm twice. |

All five product source seeds passed the actual Svelte parser first. No
compile error counted as discrimination; no source edit during matching
live runs. Original host restored after every seed to SHA-256
9ea86275b688712047703e2e6b794fcf5d6b5b61706d874b51ee13462459223a.
Driver restored after delivery probe and reporting diagnostic to
21eb8ab5c57931edcdc544885c3dd20e1c21ebef6e022cd4a86eb6efef7826f5.
No source fault remains or is committed.

Reporting-only diagnostic 094d5b deliberately throws AFTER all healthy
measurement/axe checks, exposing both completed JSON observations through
the runner annotation. Its exit 2 is NOT a product failure or passing gate.
Chromium and WebKit agree: document/main client=scroll=320, Desk/chrome
client=scroll=284; extents left=0/right=320 in both arms; 90 then 93 measured
nodes, provisioned 0→3, owned false→true, intents zero. Annotation alone is
not evidence an on-disk artifact was persisted by the local success reporter.

Final healthy root results:

- c0f0ab (full browser session 37384): both complete Garage engine populations
  pass, 254/254 tests in 77.91 s; chained isolated performance two passes /
  twenty-two explicitly unselected. No shortened real sixty-second idle test.
- b3f623: types zero errors/warnings; client 9,814 passes /451 explicit browser
  skips, 105 files pass/22 skip; unchanged 214-module build with the same
  worker/JS/CSS outputs; boundary 14/8/22 with ten Go/eleven Svelte negatives;
  static CI topology thirteen negatives rejected. All terminal exit 0.
- Final source hashes/diff check confirm no product change. No Go/SQL/server
  execution invented for this client-test-only span.

Docs/inventory/board/queue/plan/checkpoint log reconciled in this range.
Bounded current-source fixture comparison only: not a historical executable,
actual 400% zoom, Firefox, AT, real service, later Desk/full release population
or closure of accessibility's 647 px record. Full Linux browser remains
RED/held; RP-236 last Docker overlay 0 free/100%, no run/cleanup here. All
author/body/GS4/RP-331 persisted DOM/RP-333 complete performance/default-player/
privacy/platform/numeric/review/full-nine-tier 1.0 gates remain. Next accepted
native work: predeclare exact rendered Fiscal phase-edge fixtures GS1-A3,
fixed fixture time explicitly not live time. No lifecycle/archive/push/mint.

## 2026-10-07 — GS1-A3 exact rendered phase fixtures predeclared

Fresh clean HEAD 57efdc39, no newer local Claude commit. Previous turn is
progress (source guard and paired narrow Desk committed), not a verified
wait or no-progress restatement. Accepted Garage GS1-A3 requires exact
rendered early_ms-1 / early_ms / guaranteed_ms fixtures and firing text/
off-by-one cases. Existing unit helper covers 99/100/199/200 exactly; the
host browser test intentionally samples 90k/150k/250k to avoid crossing a
real timer boundary. Keep it intact; do not pretend it already supplies the
exact rendered fixture population.

RP-336 source/body finding: normative GS1 props still say serverNowMs callback
and lastOutcome; current component takes numeric reactive time and mapped
CopyKey notice, as the host actually supplies. The RP-326 local notice repair
is already recorded, not authority to rewrite someone else's RFC body. File
for author correspondence/reconciliation. No timing bug inferred, no new
interface/contract or body edit inside this range.

Add four native component-fixture cases using the actual FiscalSurface and
the actual public v4 decoder with existing period thresholds 100k/200k/300k:
elapsed 99,999 /100,000 /199,999 /200,000 ms. Fixed numeric serverNowMs=NOW
and period.opened_wall_ms=NOW-elapsed. No fake timers, shortened server clock,
host fixture clock or live time claim. The direct component must not start a
host worker/network and the fixed prop must stay fixed during DOM settlement.
Expected phase/text/disabled outputs declared independently of fiscalPhase:
99,999 ripening with remaining 0:00:01 and disabled; 100,000 and199,999 early
with explicit 50% risk/consumed-on-miss copy and enabled; 200,000 guaranteed
with its exact copy and enabled. Auto-note remaining strings respectively
0:03:21 /0:03:20 /0:01:41 /0:01:40. Use exact registered Copy output, visible
phase text and curtain linkage; no prefix-only, data-phase-only or color-only
oracle. No live-region timer, no callback before click; native disabled click
does nothing, each enabled click delivers one spy callback (not a real intent
or receipt). Unmount/remove in finally, retain every existing host assertion.

Run native Chromium/WebKit via root focused and full Garage lanes. Demonstrate
independent actual compiling faults: early helper <→<=; guaranteed helper
<→<=; suppress the actual phase paragraph text; hide phase paragraphs; change
risk percent division; omit ripening from native disabled condition. Each
must fail named rendered/visibility/readiness oracles, not syntax/type errors;
restore exact hashes between faults, no matching live process during edits.
Final types/client/build/boundary/static topology/full two-engine Garage and
chained existing performance. No fixture corpus, copy, product, balance,
schema/API/CI/Make change authorized absent separately recorded failed-first
real bug. Initial setup mistakes/diagnostic-only runs remain explicit.

Source identities before faults: FiscalSurface SHA-256
76cf10c5dece9652ea338c9768d38b8533fa89dfb708f464b7c32a20965df8f1;
fiscal-phase.ts e426fe70ebad1c7587fb7586b5768400ea1c27edc4003d767686f22d84e201f4.
Entire new span after 57efdc39 through final records requires Claude
independently of narrow Desk 54c8c8da..57efdc39, source guard
d3ce0f76..54c8c8da and every previous range. Full Linux/SQL/RP-236/331/
RP-336/author/body/AT/GS4/full AC7/default-player/privacy/platform/numeric/
full-nine-tier 1.0 holds remain; no checkbox/lifecycle/archive/mint/push.

## 2026-10-07 — exact Fiscal baseline/fault results; final probes predeclared

Actual new native fixtures pass eight executions in dd7a79 and bb5088, with
two isolated performance cases each; no current-source timing defect found.
Initial typechecks ea3473/ce9392 exited 2 because the fixture did not narrow
the versioned ParsedGameUISnapshot union. Schema_version alone cannot narrow
its generated numeric version type; explicit schema+features guard fixes it.
a0def3 typecheck then exits 0 with zero errors/warnings. Instrument mistakes,
not a product finding and not green verification until corrected.

Six actual independent faults execute under the focused native gate:
fc6ba4 early off-by-one two fail/six pass; c6bda4 guaranteed off-by-one
two fail/six pass; c0f2d5 phase text removed eight fail; f797d5 phase text
hidden eight fail; 34cb2b risk division wrong four fail/four pass; 911292
ripening no longer disables harvest two fail/six pass. Every terminal exit 2
is the named rendered phase/text/visibility/readiness assertion, not syntax.
Four Svelte mutations passed the actual compiler parser first. Both product
sources restored to the preceding pinned hashes before each new fault; all
six are now restored. Existing host assertions remain unchanged.

Predeclare two additional bounded source probes before touching source:
remaining() ceil→floor must fail exact one-millisecond countdown/auto-note
outputs; replace the actual enabled onHarvest() call with a no-op expression
must fail the spy delivery counts while disabled-before-edge remains a clean
control. No product repair, fake time, new population, copy or scope change.
Each source must parse and execute, then restore to the original FiscalSurface
SHA; no matching live process during edits. Final full two-engine Garage,
chained performance and root types/client/build/boundary/topology afterwards.
New driver SHA before additional probes:
8af13159c9438ee23b22f024ad1ea0414d1bdc5669868e9b1ebb8020c92992f2.
All preceding authority/review/full-nine-tier exclusions remain unchanged.

## 2026-10-07 — exact rendered Fiscal edges final evidence / first filter

Review by: Codex (implementer self/first filter, NOT designated).
Recorded by: Codex. Entire new span starts after 57efdc39, includes
2ff59fbe/f1914ace and the containing test/docs/tracking commit. Claude's
mandatory independent verdict must cover that full span separately from
narrow Desk 54c8c8da..57efdc39, source guard d3ce0f76..54c8c8da, performance
127eb052..d3ce0f76 and every earlier range. No self archival/approval claim.

Actual additional faults: 224f7e remaining ceil→floor fails four/two-engine
cases (ripening text says zero instead of one second; last early auto-note
also differs), with four controls passing. 43db27 enabled callback removed
fails six cases because click delivery is zero, while both disabled cases
still pass. Both terminal exit 2 are actual rendered/callback assertions;
both seeds pass the actual Svelte parser and execute. Together with the
preceding six faults, the eight fault populations fail 2/2/8/8/4/2/4/6
executions respectively, with unaffected controls retained. Source restored
exactly between probes and finally to:
FiscalSurface 76cf10c5dece9652ea338c9768d38b8533fa89dfb708f464b7c32a20965df8f1;
fiscal-phase e426fe70ebad1c7587fb7586b5768400ea1c27edc4003d767686f22d84e201f4;
test driver 8af13159c9438ee23b22f024ad1ea0414d1bdc5669868e9b1ebb8020c92992f2.
No product source fault or production diff is committed.

Final healthy native b6db84 (session83281) exits 0: full current Garage
Chromium/WebKit 262/262 passes, 78.48 s; chained isolated performance two
passes /22 explicitly unselected. The existing real sixty-second idle
population and all prior host assertions remain unchanged. Four new cases
per engine directly mount actual FiscalSurface over decoded public fixtures
and fixed server-time props; no host worker, injected fake timer, service
clock/pacing adjustment, real intent or receipt. They verify exact visible
text, independent expected rounding/risk, native disabled state and callback
delivery. Not a full-Garage/real-service/Firefox/AT/release proof.

8451c7 root types/client/build/boundary/topology exits 0: types zero errors/
warnings; client 9,814 passes /455 explicit browser skips, 105 files pass /
22 skip; unchanged 214-module worker/JS/CSS build; boundary14/8/22 and ten
Go/eleven Svelte source negatives; static topology thirteen negatives. This
static CI check is not the still-red full Linux/hosted CI gate. No Go/SQL
run invented for this test-only span. Earlier type setup failures remain
disclosed, not product defects or purported passing gates.

Fresh read-only Docker preflight during final verification: 32c79f confirms
both declared Postgres services healthy; 0235dc actual owned game-ui Postgres
container df still overlay125.7G/122.7G used/zero available/100%, shm63M free.
83801f global Docker accounting reports images80.06GB, volumes40.44GB and
build cache3.185GB; its reclaimable figures do NOT authorize broad pruning.
Three cloud-clicker-labelled old cache volumes are inspected read-only
(2bad3f), with no attached containers (42767c). Unlabelled artifacts and all
databases/release artifacts/other projects remain unowned for cleanup. Asked
the human asynchronously for tightly scoped, ownership-verified cleanup;
no approval assumed, nothing deleted and no full-disk container test begun.
Capacity/CI/RP-331 holds are freshly verified, not copied from old prose.

Canonical docs, ledger RP-336, acceptance inventory, current board, queue,
plan and checkpoint log reconciled in this range. RP-336 is author-owned
props/body correspondence, NOT an executed phase bug; this rendering proof
does not clear it. All other author/body/GS4/full AC7/actual zoom/AT/default
player/privacy/platform/numeric/review/full-nine-tier 1.0 gates remain.
Next priority: finish read-only capacity ownership/size inventory; if the
human authorizes exact scoped cleanup, recover capacity and re-run declared
full cold Linux/SQL populations. Otherwise safe accepted GS2-A3 score-vs-
Clout source binding guard remains available with predeclared faults. No
checkbox/lifecycle/archive/mint/push/deploy or preview substitution.

## 2026-10-07 — GS2-A3 source guard baseline predeclaration

At 3609e776, clean checkout; no newer local Claude commit or verdict is
inferred. Accepted GS2-A3 requires a source guard against achievement score
binding lifetime Clout or being labelled Clout. The existing achievements
boundary scans only Go ownership/dependencies; GS2-A1's rendered assertion is
not the required source gate. Clout OD-1=A keeps achievement score distinct
from the separate reach/PR axis; legitimate Clout consumers are outside this
bounded check.

Before correction, run the healthy existing achievements gate, then seed a
valid Svelte local `CloutLifetime = arm.score.lifetime` actually used by the
score frame. Separately relabel the canonical source score frame as Clout,
keeping schema/parameters unchanged. Run the existing gate on each. Green
survivors establish missing coverage, not a healthy-HEAD player bug. Restore
subjects byte-exactly after each; no regenerated/committed faulty product or
copy. No subject changes while a check runs. Baseline hashes are recorded.

Bounded correction if probes survive: extend the existing root achievements
boundary already in verify-client, without a new CI lane. Parse the actual
Svelte component, ignoring comments but rejecting Clout identifiers/literal
member names and rendered Clout labels. Inspect canonical copy through the
existing Copy Pipeline including era variants, not stale generated artifacts
or a second copy table. Scope to Trophy Case/achievement copy families, never
the separate Clout consumer. Include positive comment/score/unrelated-Clout
controls, negative fixtures, actual component/copy faults and checker-bypass
mutation. Parser/schema failures do not count as semantic discrimination.

No mechanics, player copy, protocol, balance, numeric, kernel, Make/CI or RFC
body changes. Final: healthy boundary, types/client/build/UI/isolation/topology
and focused native Trophy Case rendering. Full Linux/SQL stays capacity-held;
no cleanup approval/deletion/rerun. Self review is first filter; complete new
span after 3609e776 needs Claude separately from all preceding spans. No boxes,
lifecycle/archive/mint/push/release promotion.

## 2026-10-07 — GS2-A3 actual baseline survivors / RP-337

305b8e8b predeclared. Healthy existing gate 9cdd67 exit0. Actual component
binding c5d151 passes; initial standalone parser instrumentation 014fb7 fails
because a filesystem-path CommonJS compiler import has no named export.
That import mistake is not a syntax/product defect. Repeated with the actual
compiler default export: valid Svelte source parses, existing gate still
passes. Canonical source label fault 3c4bdb also passes unchanged Go-only gate.
RP-337 records absent coverage, not broken healthy UI. Subjects restored to
SHA256 component877a39910e1b859e962a0929d674973c3010052ccf20bcd1bac2226fae2901a9,
catalogf5346bbca960416f0cd89fba421c689201e16ff6f18c36745f0d1f9486c07b3a.
Proceed only with the predeclared existing-tool extension. Required faults:
actual component binding, computed-member binding, source base score label,
source era score label and disabling the component check. All must produce
semantic assertion failures, with exact restoration and positive controls.
No product/owner copy/generated artifact/Make/CI/RFC-body edit committed.

## 2026-10-07 — RP-337 source guard final evidence / first filter

Review by: Codex (implementer self/first filter, NOT designated).
Recorded by: Codex. Entire new range starts after3609e776, includes305b8e8b,
a32b8358 and this containing tool/docs/tracking commit. Claude must review
the entire range independently of Fiscal57efdc39..3609e776, narrow Desk,
cosmetic/source/performance and all earlier ranges. No self archival.

Existing achievements boundary now parses the actual Trophy Case Svelte AST
and uses the existing Copy Pipeline to build canonical source copy, including
era variants. No generated output/independent copy table is used as source
authority. Seven parser-valid component fixtures reject identifier/member/
computed-member/template/text/copy-key faults; four copy fixtures reject base/
era/key faults and missing required presentation. Comments/current score and
separate unrelated Clout copy are positive controls. Original Go package/
transitive dependency guards remain intact. Conservative lexical source
firewall, not arbitrary computed-string/dataflow, producer eligibility,
persisted-player or overall Trophy Case acceptance proof.

Executed real faults, all terminal make exit2 on GS2-A3 semantic errors:
6e94ed actual score binding renamed CloutLifetime; bd30bc computed-key totals
binding; 4e7b7e source score frame relabelled Clout; 357189 source grant frame
era_2000 relabelled Clout. Copy faults pass the canonical schema/parameter
loader before the new semantic check rejects them. c4bb52 disabling the
component checker makes its negative fixture fail with unexpectedly-passed,
not a parser/type failure. All component/copy subjects restored exactly to
the baseline hashes above (35cfc0/e41a6c). No transient seed is committed.

Healthy285f5b/8063b4 root source/UI/cosmetic/static topology gates exit0:
GS2 seven/four negatives; original UI14shell/8UI/22GameUI with ten Go/eleven
Svelte cosmetic negatives; package isolation22 negatives; topology13. Root
type/client/build1c137d→0aab3b terminal0: zero type errors/warnings,
9,814 passes /455 explicit browser skips,105 filespass22skip; unchanged214
module assets prediction.worker-MqspU_iu.js/index-eU1nW9jC.js/index-DaRqgLww.css.
Focused native GS2-A1 53fe51→e3a4dd exits0: two actual Chromium/WebKit
executions,260 explicitly unselected, chained isolated performance two
passes/22 unselected. Not262 native passes or Firefox/full-state proof.
No-payment5b6673 passes with six negatives/two controls. Copy8f64bb→3358d1
terminal0:658 keys, unchanged hash a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611 existing orphan warnings, generated-content manifest passes. No HEAD or
source mutation during history scan. Instrument import error disclosed above.

No Go behavior/test claim invented for this source-tool-only change. No
production/copy/generated/API/balance/clock/Make/CI/RFC byte change. No fresh
Docker capacity/recovery/hosted claim; scoped cleanup permission remains
unanswered, nothing deleted and no full-disk container population begun.
All prior exact review ranges and RP-331/full Linux/author/body/GS4/full AC7/
AT/default-player/privacy/platform/numeric/full-nine-tier 1.0 holds remain.
Next accepted safe work: predeclare GS2-A1 actual non-color state-text faults
and GS2-A5 bounded fixture/error/empty narrow readability. Capacity recovery
can take priority only with exact scoped authorization and sufficient measured
headroom. No checkbox/lifecycle/archive/mint/push/deploy/release promotion.

## 2026-10-07 — GS2-A1/A5 visible Trophy Case predeclaration

At85fdcfdf clean tree; previous turn progressed RP-337, not full1.0. No new
Claude verdict inferred. Test-only accepted GS2-A1/A5/GS0.5 supplement over
decoder-admitted public fixtures and actual GameUIApp/Trophy Case, not a
service/catalog acquisition proof. Keep every existing assertion unchanged.
Cross three populations (run/career/locked, zero rows, unavailable row copy)
with320/1280 CSS px and native Enter/Space navigation. Exact visible title,
score/state/scope/grant/possession/empty/error text; no invented rows or
read-only controls; expected diagnostic once only for missing presentation.
Measure whole-page/panel/descendant bounds and own scroll, reject clipping or
hidden text as a false reflow success, require single-column rows at320.
Native focus/Tab/Shift-Tab must stay on nav stops and return to Desk without
intents. Axe serious/critical census for these states only. No fake clocks,
400% zoom/AT/Firefox/all-state/full GS2-A5 claims.

Before healthy/fault runs, add these fixtures; then separately seed actual
parser-valid sources: state text empty (original GS2-A1 must fail too), state
text hidden, score header swapped run/career values, possession warning
removed, min-width40rem, overflow-hidden around the widened panel, empty
message removed, nav accessible name erased. Missing-copy branch alert
removed is a ninth fault. Every fault must fail its relevant semantic/native
assertion with other controls retained; parser/type failures are not evidence.
Restore source subjects byte-exactly before the next probe/final gates.
No product/copy/clock/protocol/balance/kernel/CI/Make/RFC-body changes here.

Final types/client/build/source/copy checks and full current native Garage
Chromium/WebKit population plus existing isolated performance. Full Linux/
SQL held on capacity: no cleanup permission/deletion/full-disk run. Exact
complete range after85fdcfdf needs Claude, independently of prior ranges;
self is first filter only, no checkbox/lifecycle/archive/push/release change.

## 2026-10-07 — GS2 baseline driver correction and native nav failure RP-338

6c578197 predeclared twelve test cases/twenty-four native executions. First
0aa30a exit2: Chromium twelve cases reach the reporting line then fail wrong
annotation API; WebKit twelve fail native Tab before reporting. Typecheck
688797 independently catches the wrong overload; corrected to annotation
type string. Observation widths now captured while Trophy Case is mounted,
not incorrectly sampled after returning Desk. These are driver defects.

Initial commentary suspected axe could perturb focus; that hypothesis is
DISCONFIRMED, not a product explanation. Moving axe after the complete
native keyboard route gives f8e07e terminal2: Chromium12 pass, WebKit12 fail
the same native Tab (body instead of adjacent Earnings Calls), with explicit
pre-Tab focused Trophy Case assertion passing. No axe has run at that point.
Typecheckf5dd68 passes zero errors/warnings. Visibility/layout/state text
checks preceding Tab pass in both engines. RP-338 records actual native
failure at healthy product source, distinct from the annotation instrument
error. Existing tests focused individual nav buttons but did not prove this
Tab route; Fiscal action controls already have explicit tabindex0.

This checkpoint remains TEST-ONLY: source hashes unchanged, no product
correction here. Commit the red discriminating population and record before
a separately scoped implementation. No acceptance box or broad keyboard
proof; GS2-A5/full Linux/real AT/review remains unfinished. Cross-party must
cover the failed-test checkpoint as well as subsequent repair. No archive.

## 2026-10-07 — RP-338 accepted native nav correction predeclaration

Separate from test-only6c578197..ffabda04: accepted GS0.4 keyboard rule
and GS2 keyboard map/GS2-A5 authorize making existing nav buttons native Tab
stops in the supported engine population. No new navigation/filter/mechanics
or browser preference waiver. Narrow correction: explicit tabindex0 on the
existing GameUIApp nav buttons in unchanged DOM order, matching existing
Fiscal buttons. Keep negative tabindex heading and every original/new test
assertion unchanged. No script key interception, forced focus per Tab,
native-browser configuration or test-provider changes allowed.

First rerun the exact red population; WebKit must traverse normally with no
synthetic helper, Chromium controls remain green. Then independently revert
just the nav attributes to prove actual loss of traversal fails the same
gate. Also execute the nine prior predeclared Trophy Case visibility/text/
empty/error/width/unlabelled-nav faults. Every temporary product fault is
restored exactly; only scoped native attributes are permanent. Final full
Garage Chromium/WebKit/types/client/build/copy/source/static CI gates.
No renderer/clock/copy/protocol/balance/kernel/CI/Make/RFC-body scope added.
Docs and records in repair range; full new span after85fdcfdf needs Claude
with explicit test-only and product-repair boundaries. Full three-engine/
all-state/AT/persisted-player/1.0/archival approval remains unproved.

## 2026-10-07 — RP-338 native nav correction / final first filter

Review by: Codex (implementer self/first filter, NOT designated).
Recorded by: Codex. Full new range starts after85fdcfdf and includes test
predeclaration6c578197, failed test-onlycheckpointffabda04, separate accepted
product predeclaratione30b5e00 and this containing repair/docs/tracking commit.
Claude must cover the full union, independently of source guard3609e776..
85fdcfdf, Fiscal57efdc39..3609e776 and all preceding ranges. No self archival.

Citation precision: e30b5e00's phrase "GS0.4 keyboard rule" refers to that
section's native-button host contract; the explicit keyboard-map/acceptance
requirement is GS2/GS2-A5 (GS1-A5 by reference). No invented GS0.4 clause,
new mechanic or change to anyone else's RFC/ruling text. The initial test-only
scope ended atffabda04; production work began only after separate predeclaration.

Permanent source diff is ten existing nav buttons' explicit tabindex0,
unchanged DOM order/copy/conditions/handlers, no script keyboard interception,
forced focus between Tabs, provider preference/configuration or test weakening.
80ae98→89078d native healthy population exits0:24/24 new executions,262
explicitly unselected; chained isolated performance two /22 unselected pass.
This intervention removes the prior twelve native WebKit failures. Actual
Fiscal nav tabindex removal6cc49d then fails twelve WebKit cases (Tab skips
to Reputation Board), while twelve Chromium controls pass. Same keyboard
route, not a fixture manually focusing the supposed next stop.

Ten source faults executed through native browser collection, not parser/type
errors. Each command terminal exit2; untouched population controls retained:

| Fault | Terminal output | Failed / passed selected executions | Actual witness |
|---|---|---|---|
| state text removed | 02ccaf | 10 /16 | original GS2-A1 plus exact new visible state |
| state display:none | 889096 | 8 /18 | new visibility fails; original raw-text assertions survive |
| run/career score swapped | a1eaa8 | 8 /16 | exact run2/career5 header |
| possession warning omitted | b2b198 | 8 /16 | expected warning census |
| min-width40rem | a9c38f | 12 /12 | native single-column/scroll: actual narrow overflow656>321 |
| wide panel masked by host overflow:hidden | 978954 | 24 /0 | ancestor clipping refusal |
| empty copy removed | 16fbef | 8 /16 | visible exact empty message |
| missing-copy alert removed | 11dfd6 | 8 /16 | exact role/text alert |
| Settings nav name erased | e7f679 | 24 /0 | actual critical axe button-name violation, not button lookup |
| Fiscal nav Tab attribute removed | 6cc49d | 12 /12 | native WebKit next-stop failure; Chromium control survives |

First two selectors include unchanged GS2-A1 as well as new GS2-A5; remaining
selectors execute only twelve new cases×two engines. No failed population
relabeled passing. One patch with reverse-ordered hunks failed atomically
while restoring hidden state/applying score swap; verified actual diff,
restored hidden seed, then applied the independent score probe correctly.
No unintended source edit/seed committed. Final hashes8089de:
AchievementsSurface877a39910e1b859e962a0929d674973c3010052ccf20bcd1bac2226fae2901a9
(original unchanged); GameUIApp5fa422dfdf96a8484477f26c687c23c3ee6f887c8d3caffaaebd4fcc23c6272d
(only intended explicit nav attributes). All seed CSS/copy changes restored.
Earlier annotation-driver mistakes and disconfirmed axe hypothesis retained
above, not silently treated as product causes or passing measurements.

Finalb59ba1→e86a58 terminal0: full current native Garage Chromium/WebKit
286/286,78.74s, real minute idle population retained; chained isolated
performance two /22 unselected pass. c7d6de→4e9f1f terminal0: types zero
errors/warnings; client9,814 passes /467 explicit browser skips,105 files
pass22skip;214-module build. UI JS index-Dt0SMABe.js changes as expected,
worker prediction.worker-MqspU_iu.js and CSS index-DaRqgLww.css unchanged.
Source guard seven component/four copy negatives, UI14shell/8UI/22GameUI
with ten Go/eleven Svelte cosmetic negatives and topology13 negatives pass.
Copy3d903e→8e6919 terminal0:658 keys, unchanged hash
a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611 existing orphan warnings, content manifest passes. No HEAD/source changes
during the history scan. No Go/SQL/persisted/hosted gate invented.

Twelve fixtures=populated(run/career/locked)/empty/unavailable presentation
×320/1280×Enter/Space. Exact visible state/title/scope/grant/possession/score
or empty/error; read-only/no intents; readable single-column/own-scroll/bounds
and no ancestor masking; native navigation and serious/critical axe checks.
Annotations capture Trophy geometry before Desk navigation (test annotations,
not committed measurement artifacts). Not loading/reconnect/null-arm/every
conditional nav/Firefox/actual400%zoom/manual AT/current service acquisition
or complete GS2-A5/platform/1.0 acceptance. Changed host composed needs real
SQL after capacity repair; prior-source composed output is not substituted.

Docs/ledger RP-338/inventory/plan/board/queue/checkpoint log reconciled.
Human Docker cleanup permission still unanswered; no deletion/full-disk run,
fresh capacity recovery or Linux-CI claim. All RP-331/author/body/GS4/full AC7/
AT/default-player/privacy/platform/numeric/full-nine-tier1.0/review holds
remain. Next accepted work: predeclare GS0.5 Trophy Case loading/reconnect/
null-arm context-change focus fixtures; actual GS2-A4 stays capacity-held.
No boxes/lifecycle/archive/mint/push/deploy/release/preview promotion.

## 2026-10-07 — shared Trophy Case states failed-first predeclaration

At0fe2f57c clean checkout; last turn progressed actual RP-338 repair, not
full release. Fresh802666/9b21f8/554a42: declared Postgres healthy but actual
overlay still zero available/100%, no deletion/cleanup approval/full-disk run.
Current source lacks a no-snapshot loading branch, a Trophy stale marker and
forced arm-loss Desk heading focus. These are candidates, not executed defects.

TEST-ONLY baseline range: decoder-admitted public fixtures, actual host and
native Chromium/WebKit at320/1280. Hold the initial runtime snapshot promise:
credentialled returning-player Desk heading plus common.loading role=status,
no fabricated rows/controls, then resolve and require real existing Desk.
On Trophy Case, cross recovering/closed/resync/restart with both widths:
keep exact last run2/career5 and rows, visibly mark common.stale_note, no
local score change/intents while native clock runs. Recover with unchanged
snapshot, remove stale marker; deliver an explicit valid updated snapshot
(new generator achievement earned, run3/career5) and require only that value.
Null-arm ×retained/removed feature fact ×two widths: actual snapshot delivery
forces Desk, unmounts Trophy Case, focuses a negative-tabindex Desk heading;
repeated null delivery must not move later user nav focus. Cover a newer
player-selected surface cancelling a pending forced-focus callback. No
fixture-only focus implementation or browser preference change.

Record executed failing populations before any product correction. Later
repair needs separate accepted GS0.5/GS0.6 predeclaration, not a silent scope
expansion of this test-only boundary. Existing assertions/copy/clock behavior
remain unchanged. Further discrimination faults will be predeclared after
the baseline identifies actual defects. Not real service/persistence/AT/
400%zoom/Firefox/all-state/full GS2 acceptance. No RFC/copy/balance/kernel/
CI/Make changes or lifecycle/push/owner-body promotion. Entire new span after
0fe2f57c needs Claude separately from85fdcfdf..0fe2f57c and earlier spans.

## 2026-10-07 — shared-state actual red baseline RP-339/340/341

923a51ab predeclared. e772cb→8536f7 types pass zero errors/warnings.
ac352d→fb52bf native32-case population terminal2:28 fail/four controls pass.
Four held-first-read cases fail actual absent Desk heading; sixteen connection
populations fail actual absent visible common.stale_note; eight null-arm
cases return Desk but fail missing heading negative tabindex. Cancellation
four controls pass trivially with no current focus callback, not proof a
future helper discriminates. No framework/type/collector errors caused red.

For independent native focus evidence, reorder the new null-arm assertions
(no requirement removed) to inspect activeElement before tabindex and rerun
the eight selected cases. All eight fail actual native heading focus, in both
retained/removed-fact populations. RP-339/340/341 enter the shared ledger
before product repair. Current scores/rows are authoritative fixtures, not
live persistence or a proved score-prediction defect. Controlled cancellation
uses real DOM nav handler synchronously, not physical human race timing.

This boundary is TEST-ONLY; GameUIApp unchanged from0fe2f57c, no product/copy/
API/clock/balance/kernel/CI/Make/RFC body changed. Preserve failed-first
checkpoint and original assertions. Next independently predeclare GS0.5/0.6
host loading/stale/forced-focus repair and actual severing population.
Full Linux/SQL still capacity-held; no deletion/retry. Complete new range
after0fe2f57c needs Claude including this checkpoint, separate from earlier
spans. No box/lifecycle/archive/push/release/owner-authored text promotion.

## 2026-10-07 — accepted shared host repair / further fault predeclaration

Separate from test-only923a51ab..6d36a9d3. GS0.5/GS0.6 already specify
loading heading/status, stale last-authoritative disclosure and forced Desk
heading focus; no owner choice/mechanics/wording is invented. Narrow host
repair: credentialled initial Desk loading uses existing common.loading,
no fabricated controls; Trophy-only stale note uses existing common.stale_note
on offline/resyncing/not-ready; original authoritative arm binding unchanged.
Desk h1 gets negative tabindex. The existing null-arm effect schedules heading
focus after DOM update, cancelled if a newer player selection/snapshot wins;
normal navigation never gains autofocus. Existing lifecycle precedence,
other surfaces' stale regions, protocols, theme/balance/copy/clock/provider/
CI/Make/RFC body and all old assertions untouched.

Before repair add independent not-ready-only population (subscriber has not
reported recovery; HTTP snapshot exists, offline/resyncing false), and a
newer explicit Desk nav selection cancellation control alongside Settings.
Both widths; forty total native shared-state executions, not thirty-two.
Record new baseline: missing stale marker must fire not-ready controls too;
cancellation controls may pass trivially now and require failing guard probe.

After repair independently seed actual valid source faults: remove loading
status role; remove loading heading; omit stale copy; hide stale text; omit
not-ready branch; keep stale marker after recovery; no-op forced focus;
remove Desk heading tabindex; bypass pending-focus cancellation guards (the
newer Desk nav choice must then lose focus). Each must fire real semantic/
native assertions; controls retained, no parse/type-error credit. Restore
exact source between probes and finally. Final full native Garage two-engine,
types/client/build/source/topology/copy. No SQL/full-Linux/hosted/Firefox/AT/
all-context/first-read-rejection/retry proof invented; capacity remains held,
no cleanup authority/deletion. Whole new span after0fe2f57c needs Claude
including failed-first tests and this separate product scope. No archival.

## 2026-10-07 — RP-339–341 local repair / first filter

Review by: Codex (implementer first filter, not designated review).
Recorded by: Codex. Proposed complete reviewed span:0fe2f57c-exclusive
through this containing repair/record commit, including923a51ab,
6d36a9d3 and88153822. Claude verdict still required independently of
85fdcfdf..0fe2f57c and every earlier exact span. No archival gate consumed.

Supplemented baseline645459 terminal2:32 failures/eight controls pass,
286 unselected, forty native executions. Not-ready-alone also fails missing
stale marker; Settings/Desk cancellation still pass trivially before callback.
Initial repair imported tick beside the existing timer variable: f97900
collects zero tests/two import errors;4fcaa3 types three errors. No behavioral
or mutation credit. Alias afterDOMUpdate resolves collision without changing
timer;38d5b3 types clean, b3bce6 forty selected native passes/performance two.

Actual accepted host correction: existing loading heading/status while Desk
first read is held; Trophy-only stale note for offline/resync/not-ready;
last scores/rows stay authoritative; Desk h1 negative tabindex; existing
null-arm effect schedules heading focus after DOM update and cancels if
newer selection/snapshot wins. Normal nav never autofocuses. Other shared
effect contexts are changed mechanically but not proven by Trophy fixtures.
No copy/protocol/clock/balance/kernel/CI/Make/RFC-body edit. The forty fixture
population uses public Runtime messages/decoder-admitted v4 at320/1280:
held read, recovering/closed/resync/restart/not-ready, null arm with retained/
removed fact, repeated null after user navigation, Settings/Desk cancellation.
The synchronous race uses real DOM handlers, not physical human timing.

Nine compiling source faults run independently on the same forty cases:

| Fault | Terminal output | Fail/pass | Actual discrimination |
|---|---|---|---|
| Loading role omitted |399def|4/36|exact status absent|
| Loading heading empty |e2f1f8|4/36|exact visible heading absent|
| Stale note omitted |ee1ef0|20/20|all five transport populations lack marker|
| Stale note hidden |7d1d6b|20/20|native rectangle zero|
| Not-ready condition omitted |50b0d8|8/32|not-ready-alone AND restart fail|
| Stale note retained after recovery |45638d|20/20|marker wrongly persists|
| Forced focus no-op |34bccf|8/32|activeElement remains nav/body|
| Desk tabindex omitted |1d9a3a|8/32|native heading focus fails|
| Cancellation guards bypassed |5c4d09|4/36|newer Desk nav loses focus; Settings controls survive|

All terminate2 on real semantic/native assertions; no parse/type-error credit.
Exact restoration71c5c5: host
3ccce9406c83d57be75f4d89a9cea7cd32ea3665eeb5df947c313ec910f24217;
driver3ed18590ba4a52f050f3a32316b100c358f447294301d12b07d0adb63f70a06c.

Final immutable-source root checks f5f331→24fd6d terminal0: full native
Garage326/326,83.21s, real minute idle retained; chained isolated performance
two passes /22 unselected. c74ec0→b3c2a0 terminal0: types zero errors/warnings,
client9,814 passes /487 explicit browser skips (105 files pass22skip),214-
module production build. UI JS index-CaLZW7xF.js changes; worker
prediction.worker-MqspU_iu.js and CSS index-DaRqgLww.css unchanged. Source
boundary seven component/four copy negatives, shell/UI cosmetic ten Go/
eleven Svelte negatives, CI topology13 negatives pass. dc33f6→6e906e
terminal0:658 copy keys, unchanged
a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611 existing orphan warnings, deployment content manifest passes. No source/
HEAD edit during live checks/history scan. Diff whitespace clean.

Docs/ledger/inventory/plan/board/queue/checkpoint log reconciled in this
repair commit. Bounded local first filter only, not real service/persistence,
all shared states/all contexts/first-read rejection/retry/Firefox/manual AT/
actual400%zoom/full GS2/platform/1.0. Next accepted work: separately predeclare
remaining Fiscal/Meters/Reputation forced-context focus populations; first-read
failure/retry needs truth audit before new UI. Actual GS2-A4/Linux/SQL remain
capacity-held; fresh554a42 overlay0 available/100%. Human scoped cleanup
permission unanswered; no deletion or full-disk run. All prior author/body/
GS4/AC7/default-player/privacy/platform/numeric/full-nine-tier1.0/review holds
remain. Goal active/progress; no boxes/lifecycle/archive/mint/push/deploy/
release/shortened-preview promotion.

## 2026-10-07 — remaining forced-context focus proof predeclaration

At82614848 clean source, RP-339–341 local correction is committed, not
designated-approved. New TEST-ONLY range under accepted GS0.5/0.6, with no
product/copy/RFC/CI change. Fiscal/Meters/Reputation each uses an exact-decoder
admitted mounted public arm, native Enter nav activation, then snapshot arm
null with feature fact retained/removed at320/1280 in Chromium/WebKit:
twelve declarations/twenty-four native executions. Require actual initial
surface, selected nav focus, null unmount/Desk selection/visible negative-
tabindex heading focus, no intents, then a later native Settings choice
retains focus through repeated null delivery. Native serious/critical axe.
Reputation uses the existing public wire grammar, not a new fixture adapter
or mechanics. No claim of all contexts/lifecycle/manual AT/Firefox/live SQL.

Run healthy population first; separately remove each original armless
predicate (Fiscal/Meters/Reputation) in valid host source and require its
eight executions to fail; other sixteen controls retained. Independently
no-op forced focus and remove Desk tabindex: each must fail twenty-four
native focus assertions. Restore exact host after each probe. No framework/
compile errors credited. Final full Garage/types/client/build/boundaries/
copy/topology and records. Complete82614848-exclusive span needs Claude
independently of0fe2f57c..82614848 and all earlier exact spans; no box/archive/
push/cleanup permission inferred. Real Linux/SQL capacity hold stays live.

## 2026-10-07 — remaining shared-effect consumers / test-only first filter

Review by: Codex (implementer first filter, not designated reviewer).
Recorded by: Codex. Entire82614848-exclusive through this containing proof/
record commit needs Claude, includingfb2a9488 predeclaration. Separate earlier
0fe2f57c..82614848 and all earlier exact spans remain owed; no archival gate.

Healthy987eab→1e8ca2 terminal0:24 selected native cases/326 unselected;
isolated performance two/22 unselected pass. df4cdf→645bd5 types zero errors/
warnings. Fiscal/Meters/Reputation each uses a populated decoder-admitted
public arm (Reputation existing R9 node grammar), native Enter selection,
arm-null at320/1280 with fact retained/removed, actual Desk selection and
heading focus/negative tabindex. Later native Settings choice survives repeated
null delivery, no intents/serious-critical axe. Twelve declarations, two
engines; not live service/physical race/all-context/lifecycle/manual AT.

Five independent compiling actual host faults, all terminate2:

| Fault | Output | Fail/pass selected | Assertion |
|---|---|---|---|
| Fiscal arm-loss predicate omitted |728260|8/16|surface remains Fiscal, not Desk|
| Meters predicate omitted |5ce14c|8/16|surface remains Meters|
| Reputation predicate omitted |401927|8/16|surface remains Reputation|
| Heading focus no-op |c688fb|24/0|actual nav/body, not Desk heading|
| Desk tabindex removed |f7f5ec|24/0|native focus fails|

Exact restoration66e771/bc50be: no host diff against82614848;
SHA2563ccce9406c83d57be75f4d89a9cea7cd32ea3665eeb5df947c313ec910f24217.
Driver4372677b22026fdd853b4a784d0e9dd9845e65d4f6dffece4b38b767d2f5705d.
No product/API/copy/clock/balance/kernel/CI/Make/RFC-body change in range.

Final immutable-source checks4b09cf→6809b1 terminal0: full native Garage
350/350,83.94s, real60s idle retained; chained isolated performance two pass/
22 unselected.3bf328→a6d53c terminal0: types zero errors/warnings, client
9,814 passes/499 explicit browser skips (105 files pass22skip),214-module
build unchanged UI/worker/CSS. Source boundary7component/4copy negatives,
shell/UI10Go/11Svelte cosmetic negatives and topology13 negatives pass.
412eb0→1795f8 terminal0:658 copy keys unchanged
a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611 existing orphan warnings, content manifest pass. No source/HEAD edits
during live checks/history scan. Docs/ledger/inventory/plan/board/queue/
checkpoint log reconciled; no checkbox/lifecycle/archive promotion.

Next source candidate RP-342 recorded immediately, not a confirmed runtime
defect: refresh catches first failure as offline while no-snapshot Desk
renders loading; ShellRuntime start requests once and visibility lifecycle
can request another read, not periodic startup polling. Initial path occurs
before transport subscription/worker exists. Existing archived bootstrap
retry contract is NOT authority for a new snapshot retry interface. Next
TEST-ONLY truth audit must predeclare actual rejected/malformed/healthy reads,
visible status/control/request census and existing lifecycle recovery before
correction or new policy. Wrong guessed archived path game-ui.md search
returned absent; corrected actual game-ui-screens.md inspected, no authority
invented. No new empirical/owner decision resolved by source inspection.

Full Linux/SQL/GS2-A4 remain capacity-held, cleanup question unanswered/no
deletion/full-disk retry. All Firefox/actual zoom/manual AT/body/GS4/AC7/
default-player/privacy/platform/numeric/full-nine-tier1.0/review holds remain.
Goal active/progress, not complete/blocked; no push/deploy/mint/release/
shortened-preview substitution.

## 2026-10-07 — GS3 reconnect values/disclosure failed-first predeclaration

At4b2fd904 clean checkout. Previous turn progressed RP-342 measured diagnosis/
draft; no startup Retry ruling or cleanup answer received. New TEST-ONLY range
under accepted GS3/GS0.5, not the startup draft. Native Chromium/WebKit320/1280,
public decoder-admitted distinct eleven-value Meters arm. Native Enter nav
selection, then recovering/closed/resync/restart and independent not-ready-
alone (HTTP snapshot exists, subscriber has not reported recovery). Ten
declarations/twenty native executions. Require visible common.stale_note,
exact per-ID native meter/value/band associations in wide and narrow layouts,
last values unchanged through350ms real native elapsed, no intents/focus theft.
Recovery clears marker without changing values; only delivered valid newer
snapshot changes all values/bands. Native axe/reflow/read-only checks retained.

No missing-note defect claimed from source alone. Run actual red baseline
before correction; file its ledger observation. Later accepted host note repair
requires separate product predeclaration, not expansion of test-only scope.
No source/copy/clock/transport/balance/kernel/CI/Make/RFC body/threshold change,
all old tests intact. Not real service/SQL/default-player/all-state/all-engine/
manual AT/actual zoom/full GS3/1.0 proof. Complete new span after4b2fd904 needs
Claude independently ofa551d3c2..4b2fd904 and every earlier span. Linux/SQL
capacity/Firefox/privacy/platform/body/full1.0/review holds stay live; no
cleanup/delete/push/archive or startup authority inferred.

## 2026-10-07 — RP-343 actual failed-first Meters disclosure

f3b2f421 predeclared, TEST-ONLY. ba066e→6ee0f0 types zero errors/warnings.
9752ac→4641ad terminal2:20/20 selected native cases fail exact missing visible
common.stale_note,378 explicitly unselected. Both engines five connection
states ×320/1280. Actual initial eleven-ID native values/layout/focus/read-only/
axe controls pass BEFORE note failure. No framework/compile/timeout cause or
inference that values themselves predict incorrectly. Existing generic
meters.as_of_note remains; it does not supply required connection disclosure.
RP-343 recorded immediately. No product/copy/clock/CI/Make/RFC byte changed.
Later scoped host correction needs separate GS0.5 product predeclaration;
no startup draft/D-023/cleanup decision inferred. Complete new span after
4b2fd904 requires Claude including this red checkpoint; no promotion.

## 2026-10-07 — accepted RP-343 repair / fault predeclaration

Separate PRODUCT scope after test-onlyf3b2f421..0a191ebd. Accepted GS3 consumes
GS0.5's exact reconnect condition and common.stale_note; extend the EXISTING
host marker's surface membership from Trophy-only to Trophy OR Meters.
No new copy/component props, value math, event/recovery/auth/clock/provider/
wire/balance/CI/Make/RFC-body behavior; marker remains a paragraph, not a new
live region. Preserve every original assertion and the failed-first checkpoint.
Actual Meters values/bands still come directly from bound authoritative arm.

After repair require20 healthy selected executions, then six independent
compiling faults on the same population: omit Meters marker membership;
hide marker; omit not-ready condition (restart and not-ready-alone must fail);
keep offline true on recovery only for Meters (marker must wrongly persist);
replace Meters values with +1 whenever stale (native value/text exact oracle
must fail); ignore delivered snapshot message (new authoritative values must
not appear). Initial not-ready controls may fire the value fault before its
later stale phase; report that honestly. Restore exact host between each
probe/finally, no type/import error credit. First three are disclosure faults,
last two independently prove value/update discrimination, not a demonstrated
production prediction bug. Final full Garage/types/client/build/copy/boundary/
topology and transactional docs/ledger/inventory/plan/queue/log closeout.

Whole4b2fd904-exclusive through repair/records needs Claude independently of
a551d3c2..4b2fd904 and all earlier spans, including red test and this scope.
RP-342/D-023 remain draft/owner-held, no Retry/auth/cleanup answer inferred.
Real-service/SQL/all-state/all-engine/manual AT/zoom/full GS3/1.0 gates remain;
no box/lifecycle/archive/push/mint/deploy/release/preview promotion.

## 2026-10-07 — RP-342 first-read truth audit predeclaration

At a551d3c2 clean checkout. Previous turn progressed committed host repair
and cross-consumer proof; no current review/lifecycle authority inferred.
OBSERVATION/TEST-ONLY scope: real createBrowserGameUIRuntime, injected fetch
with actual Response.json/production decoder, mounted host and native
Chromium/WebKit at320/1280. Synthetic stored credentials, no real secrets.
Hold first founder-state response and prove actual loading/aria-busy/no
controls before release; then network rejection, HTTP401, HTTP503, malformed
JSON, malformed v4 arm and legacy-v3 response. Observe settled read/pending,
visible loading/offline/error copy, control count, diagnostic count, same
credential retention and exact request census. Wait350ms real native time
to distinguish immediate hidden auto-retry from an idle terminal display.
This cannot prove indefinite failure or real server auth/network behavior.

Capture structured observations; do NOT encode the bad loading behavior as
a desired assertion. Actual read rejection and healthy controls validate
measurement provenance. Exercise existing visibilitychange listener by
explicit event dispatch (not physical hide/resume): non401 cohort receives
a valid same-Founder v4 on second read;401 remains401. Assert no bootstrap,
refresh or gameplay write/credential mutation, and no socket before valid
snapshot. Separate healthy-first-response control at both widths.
Fourteen declarations/twenty-eight native executions. No timers/clock policy
change or runtime-double-only error attribution.

Predeclared discrimination: replace failure replies with valid200 -> each
failure oracle must fail; prevent actual founder-state read -> held/read
counts must fail; no-op lifecycle request -> successful recovery population
must fail; suppress credential presence -> startup population must fail.
Use only valid source/instrument faults, not parser/type errors; restore
exact source/driver between probes. Final local types/client/build/native/
copy/boundaries/topology as proportionate. Production unchanged permanently.

GS0.5 says loading when no snapshot, reconnect with last values, error for
decode/presentation failures; it does not name a completed-first-read failure
precedence, credential-expiry flow or a snapshot retry interface. Archived
bootstrap idempotency and draft browser-session-renewal are not authority for
those. Confirm observed behavior first, then file precise DESIGN-GAP and
propose bounded draft contract if required; no owner-copy/accepted-body edit.
Entire new span after a551d3c2 needs Claude independently of prior spans.
Capacity/Firefox/AT/full-nine-tier1.0/privacy/platform/review holds remain;
no deletion/push/archival/release promotion.

## 2026-10-07 — RP-342 executed first-read diagnosis / draft route

Review by: Codex (implementer/researcher first filter, not designated review).
Recorded by: Codex. Entire a551d3c2-exclusive through this containing record
commit requires Claude independently of82614848..a551d3c2,
0fe2f57c..82614848 and all earlier exact spans. No archival gate consumed.

ced3b5c7 predeclared. Actual createBrowserGameUIRuntime + injected fetch,
real Response.json/decoder + mounted native host. Synthetic stored credentials;
read-only path census excludes bootstrap/refresh/gameplay writes and checks
unchanged credential document. Actual held initial read, no snapshot/socket/
controls; first result then350ms native elapsed. Six failures/network401503/
JSON/arm/valid-v3 ×320/1280 ×two selected engines=24, plus four healthy controls.
Existing visibilitychange listener exercised by dispatch, not physical hide.
Second injected reply valid same Founder except401 remains401. Fake socket
only starts after accepted v4; no real server auth/network/SQL inference.

Initial151e2d→48dd8f types clean, e78086→75e341 terminal2: four legacy cases
failed WRONG fixture, provision_cap was still a v4-only generator field;24
other cases pass. Correct fixture removes features AND provision_cap; explicitly
prove v3 decoder admission before runtime's live-v4 refusal. No assertion loosened.
58e785→d2b0c8 healthy28/performance two passes. Console.info did not appear in
native runner output, so first measurement output was not externally readable;
replace with serializable TaskMeta observation rows, not pretend unseen logs
were inspected. First types7e2de6 fails missing TaskMeta field; explicit module
augmentation fixes it (887370→c754d0 zero errors/warnings). Initial JSON output
budget truncates an unselected-test middle: whole parse fails; bounded metadata
extraction locates28 but is NOT cited as a complete parsed report. No evidence
promotion from that partial output. Final larger-budget report is parsed whole.

Four valid independent faults all terminal2 on actual assertions:

| Fault | Output | Fail/pass selected | Actual discrimination |
|---|---|---|---|
| All failure replies become valid200 |4626f4|24/4|failure promises resolve, rejection oracle fires|
| Startup stream read severed |3037f0|28/0|actual read count0 instead of1|
| Existing visibility request severed |d3f094|24/4|second read count1 instead of2|
| Runtime credential presence suppressed |efdf1e|28/0|wrong startup path/read count0|

All source/instrument faults restored. b947a6 no runtime/shell/lifecycle diff;
9cfc52 hashes: host3ccce9406c83d57be75f4d89a9cea7cd32ea3665eeb5df947c313ec910f24217;
runtime0a8c420eb9968aa76a3e9ece91c918a43e3d8c4530773c8818bce5104c0cb416;
shell3cac14fe775d6e7d58c668630909e84018814ddad0d791f75fb8972ba8beda6e;
lifecycle2756204599b10b8db160f09f35d8e1d84dd981835267b10bfc3db555a0543c28;
driverd36e1996072d30604a3dfc400162ec470400d63de184a335af5230d3409e1798.
No permanent production/copy/balance/kernel/clock/CI/Make change.

Final JSON6b745b→690b15 terminal0, complete parse success=true,28 passed,
zero failed,350 explicitly unselected,two modules; all28 metadata populations
present,52 settled/lifecycle rows. Preserve exact bounded metadata projection
in first-read-observation.v1.json, with subject hashes/command/limitations.
It is not a fabricated full runner artifact or acceptance verifier; rows do
not independently name engines, selected native invocation provides population.
All six failures aftersettlement:busy=false/loading=true/controls0/socket0/
visible offline=false/error=false/alerts0/diagnostics0/credentials retained.
Healthy four:loading=false/controls13/socket1. Valid lifecycle replies recover
the five non401 cohorts at reads2;401 remains loading/controls0/socket0 with
credentials preserved. Observation controls DO NOT assert bad display desired.
Absence of automatic retry is observed350ms only, not indefinite behavior.

Final full nativee179d3→fb5b1a terminal0:378/378,93.18s, existing real60s idle
retained; chained isolated performance two/22 unselected pass.0ebed6→49ec29
terminal0:types zero errors/warnings,client9,814 passes/513 explicit browser
skips,105 files pass22skip;214-module build byte-unchanged UI/worker/CSS.
Boundary7component/4copy negatives,10Go/11Svelte cosmetic negatives and CI13
topology negatives pass.41b233→79fe02 terminal0:658 copy keys unchanged hash
a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611 existing orphan warnings/content manifest pass. No source/HEAD mutation
during live checks/history scan. Default green means instrument/regr, NOT
startup defect correction/full CI/full release.

DESIGN-GAP: GS0.5 loading includes no snapshot, reconnect assumes last values,
arm error does not name completed-first-read failure precedence/retry UX.
Archived bootstrap retry and draft browser-session-renewal do not supply
authority. Draft game-ui-first-read-recovery.md proposes honest failure,
single-flight read-only Retry, exact status/identity/copy/AT/live-service
gates, no credential clearing/implicit account replacement/token renewal.
D-023 owner posture/wording and author-body reconciliation pending; asynchronous
question sent, not a ruling. No accepted RFC body/status altered, no authored
player copy touched. RP-342 remains OPEN on construction/acceptance.

039731 retained-metadata census:28 executions/52 rows/14 names each twice,
current host/driver hashes match. Read-only in-memory missing-population and
wrong-driver corruptions independently reject; no retained artifact altered.
This census is not promoted to a complete artifact-schema/acceptance validator.

Ledger/docs/index/inventory/decision/plan/board/queue/checkpoint log agree.
Next accepted work: predeclare GS3 Meters connection-state/stale-value proof
under GS0.5, separate failed-first correction if it fires. Current populated
AC7 supplement exists already (read at554–685), not falsely queued as absent;
full performance/excluded/manual-profile obligations remain. No Docker cleanup
answer/deletion/full-disk run; full Linux/SQL/GS2-A4 held. Firefox/AT/body/GS4/
AC7/privacy/platform/numeric/full-nine-tier1.0/all prior review gates remain.
Goal active/progress, no checkbox/lifecycle/archive/push/mint/deploy/release/
shortened-preview substitution.

## 2026-10-07 — RP-343 Meters disclosure repaired / bounded acceptance evidence

Review by: Codex (implementer first filter, NOT designated review).
Recorded by: Codex. Whole 4b2fd904-exclusive through this containing repair/
record commit requires Claude, including f3b2f421 test predeclaration,
0a191ebd red checkpoint and bfd7b4ff separate product predeclaration. Earlier
a551d3c2..4b2fd904, 82614848..a551d3c2, 0fe2f57c..82614848 and all earlier
exact ranges remain independently owed. No archival gate consumed.

Ten new declarations / twenty native executions: recovering, closed, resync,
server-restart and independent not-ready-only, at 320/1280 in Chromium/WebKit.
Real decoder-admitted public fixture; runtime double, not real service.
Native Enter selection/focus, exact eleven rows with distinct committed
values/bands and native labelled associations in both layouts, full-page
reflow/axe/read-only controls. Connection state retains exact values through
350 ms real elapsed; recovery clears note WITHOUT changing values. Only an
explicit newer snapshot changes all values (+8, including a band crossing).

Baseline 9752ac→4641ad terminal 2: all twenty cases fail the exact missing
stale-note assertion after initial value/focus/layout/axe controls pass;
378 other cases explicitly unselected. Types ba066e→6ee0f0 clean. Existing
as-of footer does not state connection failure. RP-343 records disclosure,
not a demonstrated production extrapolation defect. Test-only checkpoint
0a191ebd precedes separately predeclared repair bfd7b4ff. Permanent product
diff is ONE surface-membership condition in the existing stale paragraph;
no new player text/live region/math/props/events/auth/clock/wire/CI behavior.

Healthy focused da6355→72d18e terminal 0: twenty pass / 378 unselected,
chained performance two pass / 22 unselected. No assertion weakened.
Six independent compiling actual source faults, each restored before next:

| Fault | Output | Fail/pass selected | Actual assertion |
|---|---|---|---|
| Omit Meters note membership | 2ac613 | 20/0 | Exact stale note absent |
| Hide stale paragraph | 14f8e4 | 20/0 | Native visible rectangle is zero |
| Omit not-ready predicate | 4e76d7 | 8/12 | Restart and not-ready-only lack note; other controls pass |
| Keep Meters offline after recovery | afa808 | 20/0 | Stale note incorrectly remains |
| Add one to stale Meters values | faedc6 | 20/0 | Native committed value differs; some fire at initial not-ready control |
| Ignore delivered snapshot binding | 47a062 | 20/0 | New committed value 8 absent; old 0 remains |

All fault commands terminal 2 for semantic/native assertions, not type-error
credit. Last two prove value/update oracle discrimination, not an existing
production value bug. Restoration 0c7640: host intended SHA256
8c81a722e581659d30db984854e9b9598e6575afe71b2b6eead8d1b9c59fa298;
driver 77a2e80ea0f562aee0ffccd7fc2cd477e85cd742601c8a0a5953d2579e88d31c;
unchanged MetersSurface 5ebc07105fb33e745935dfe5a8f53b3c378cdd081b603ebce0eba103b26f4a95.
cf6da4 diff check clean. RP-342 retained observation pins its earlier source,
not this changed host/driver; no historical subject identity silently rewritten.

Final full native 087a3b→5bc84d terminal 0: 398/398, 98.10 s, original real
60 s idle retained. Chained isolated performance two / 22 unselected pass.
211548→cc9fb2 terminal 0: types zero errors/warnings; client 9,814 passes /
523 explicit browser skips, 105 files pass / 22 skip. Build 214 modules,
UI JS index-BzPaPJEo.js changes as expected; worker prediction.worker-MqspU_iu.js
and CSS index-DaRqgLww.css unchanged. Achievement boundary seven component/
four copy negatives; shell/UI boundaries and ten Go/eleven Svelte cosmetic
negatives; static CI topology thirteen negatives pass. b61466→0367fd copy
terminal 0: 658 keys, unchanged SHA256 a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611 existing orphan warnings, content manifest pass. No source/HEAD edits
during live final checks/history scan. Not full Linux/hosted CI/Postgres.

Process finding against my own record: the three RP-343 entries were inserted
ahead of the prior RP-342 section, rather than at EOF (f3b2f421, 0a191ebd,
bfd7b4ff). Commit chronology establishes the actual order; no prior entry
was edited/deleted. This still violates append-only placement. Disclosed here
at EOF; do not reorder reviewed/history-cited records to conceal the slip.
All future entries append after the exact current EOF.

Docs/ledger/inventory/plan/board/queue/checkpoint log reconciled. Next accepted
work: inspect current GS3 on-surface announcement/off-surface badge evidence,
then predeclare missing discrimination if any. No startup Retry authority
inferred from draft RP-342/D-023. Cleanup question unanswered; no deletion/
full-disk run. Real service/SQL/default-player/all-state/all-engine/Firefox/
manual AT/actual zoom/body/GS4/full AC7/privacy/platform/numeric/full-nine-tier
1.0 and all independent review remain. No boxes/lifecycle/archive/push/mint/
deploy/release/shortened-preview substitution. Goal active/progress.

## 2026-10-07 — GS3 event/badge/value separation predeclaration

At 9f485993. TEST-ONLY accepted GS3 focus/announcement + GS0.3/0.5 scope:
existing combined achievement/meter fixture asserts a badge and on-surface
text, but does not explicitly assert unchanged global text off-surface or
event-versus-snapshot value authority. First run that existing population
with a compiling off-surface global-announcement fault and restore; record
its executed survivor/failure, not infer vacuity from reading.

Add a bounded native population: p(doom) and one Standing constituency ×up/
down ×320/1280 ×Chromium/WebKit (eight declarations, sixteen executions).
Production transport-envelope and announcement decoders admit each event;
runtime-double delivery is NOT a real socket/server acquisition. Exact
eleven-ID public snapshot, existing presentation/copy only. Native Enter nav,
off-surface badge with unchanged global status, visit clears badge without
invented announcement/value changes, on-surface exact polite announcement,
only explicit newer snapshot updates native values/bands. Opposite band
event supplies distinguishable text before old consumed cursor replay;
old replay must not overwrite it. Unknown presentation ID withheld with one
diagnostic, no new text/badge/value/focus/intents. Axe/reflow/read-only controls.

After baseline, independently execute compiling source faults: on-surface
announcement omitted; off-surface announcement introduced; nav badge omitted;
visit badge clearing omitted; cursor deduplication bypassed; polite status
role removed; snapshot binding ignored. Restore actual production source
between faults and at end. Every claim cites native semantic failure, not
type/import failures. All old assertions/time budgets preserved. No permanent
product/copy/clock/transport/schema/balance/CI/Make/RFC body change under this
scope; any actual production defect gets a ledger and separate product scope.
Final full Garage/types/client/build/copy/boundaries/static topology, then
transactional inventory/plan/board/queue/log records. Entire new span after
9f485993 requires Claude, independently of earlier exact repair/research
spans. Not AT/Firefox/all-state/reconnect/actual zoom/real-service/SQL/full1.0
proof. Cleanup and startup draft remain unanswered, no inferred deletion/
Retry/archival/push/release authority. Goal active/progress.

## 2026-10-07 — RP-344 executed old oracle survivor

96439cfe predeclared. Actual host off-surface branch now additionally writes
the exact meter announcement while retaining its badge. Existing combined
"announces an earned achievement once" population a6d71d→69fffe terminal 0:
both selected native engines pass / 396 unselected; chained performance two
also pass. This disproves that old test's no-global-announcement discrimination,
NOT that production currently announces globally. Actual source restored before
adding tests; new bounded population separately checks exact unchanged region,
badge/visit/event/snapshot/old-cursor/unknown-presentation/focus/read-only paths.
RP-344 filed immediately. No permanent production change under test-only scope.

## 2026-10-07 — GS3 additional value/refusal discriminator predeclaration

Same 96439cfe test-only range, assertions unchanged. Also seed an event-driven
meter-value binding (instead of waiting for a snapshot), and a fallback from
unknown meter ID to the p(doom) presentation. These directly challenge the
event/value separation and withholding assertions; restore source afterward.
The final snapshot-binding fault has not yet run: its attempted patch used
an outdated short line, so apply_patch refused it without changing source.
Six earlier probes already restored; no failed command is counted as evidence.

## 2026-10-07 — RP-344 discriminating event/badge/value population

Review by: Codex (implementer first filter, NOT designated review).
Recorded by: Codex. Complete 9f485993-exclusive through this containing proof/
record commit requires Claude, including 96439cfe. Earlier exact scopes,
including 4b2fd904..9f485993, remain independently owed; no archival gate used.

Production envelope/announcement decoders admit each public Company event,
then runtime-double delivery drives native host. Eight declarations / sixteen
executions: p(doom) and users Standing, up/down through 69↔71, 320/1280 in
Chromium/WebKit. Initial eleven-value public snapshot has distinct per-ID
values and explicit band crossing for the selected meter. Native Enter nav,
focus retained, exact catalog text/status role/badge text, visit clearing,
event-value separation before fresh snapshot, opposite announcement before
replaying two old consumed cursors. Unknown decoder-legal presentation ID
withheld with one diagnostic; no text/value/focus/command fabrication.
Exact eleven native values, non-color bands, layout/axe/read-only checks.
Repeated synthetic crossings are a controlled presentation population, not a
claim this is a producer-issued real event sequence or reconnect session.

Initial types 17aff9→f14f28 terminal 0, zero errors/warnings. New baseline
bf834a→d91ce6 terminal 2, sixteen failures from my invented separator space
in nav.textContent: Svelte emits title+badge without that source whitespace.
No production/copy defect inferred. Correct expectation of existing exact
adjacent text, retaining badge/clearing/global-text/value assertions. Healthy
9d18ea→9e5479 terminal 0: sixteen passes / 398 unselected, chained performance
two / 22 unselected. Not a passing production-change claim; product unchanged.

Nine independent actual compiling faults, all terminal 2, each sixteen failed
/ 398 explicitly unselected; restored between probes and at end:

| Fault | Terminal output | Actual discrimination |
|---|---|---|
| Omit on-surface announcement | f63d34 | Exact visible announcement absent |
| Introduce off-surface global announcement | 4aa6f5 | Global text changes instead of remaining unchanged; identical fault survived old test at69fffe |
| Omit nav badge | c91c69 | Exact changed badge missing |
| Omit clear-on-visit | 97a030 | Badge incorrectly remains after visit |
| Bypass consumed-cursor guard | dbd6fa | Old High/Low text overwrites the distinguishable latest opposite text |
| Remove polite status role | 13cb65 | Native region role null, not status |
| Ignore snapshot binding | 650f1a | Native value retains69/71 instead of fresh71/69 |
| Substitute p(doom) for unknown ID | e31c09 | Required unknown-ID diagnostic missing |
| Bind values directly from event | e7e26f | Native value changes before authoritative snapshot |

No type-error/timeout failure counted. A refused stale-line patch for the
snapshot fault changed no source and received no evidence credit; correct
actual line then mutated/executed. Restoration/diff check 1f821e: host no
residual diff; SHA256 8c81a722e581659d30db984854e9b9598e6575afe71b2b6eead8d1b9c59fa298
matches 9f485993; driver da680c12e0f9ef38e9a4b8bea900a218db382b91e82e6e12e41990663dd87473.
No permanent product/catalog/provider/clock/balance/wire/CI/Make/RFC changes.

Final native 13803f→90dfc8 terminal 0: 414/414, 103.02 s, original real60s
idle unchanged; chained performance two / 22 unselected pass.428fd7→181fac
terminal 0: types zero errors/warnings, client 9,814 passes / 531 explicit
browser skips, 105 files pass / 22 skip.214-module build unchanged JS/CSS/
worker identities index-BzPaPJEo.js, index-DaRqgLww.css,
prediction.worker-MqspU_iu.js. Achievement boundary seven component/four copy
negatives; shell/UI and ten Go/eleven Svelte cosmetic negatives; thirteen CI
topology negatives pass.0f2960→0e964d copy terminal 0: 658 keys, unchanged
SHA256 a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
611 existing orphan warnings, content manifest pass. No source/HEAD edits
during live final checks/history scan. Actual native assertion faults and
bounded passing population, NOT all eleven event IDs/eras/states, real
socket/producer/reconnect/SQL/Firefox/manual AT/zoom/full GS3/full CI proof.

Read-only capacity recheck d23ee1 terminal 0: declared Postgres overlay still
125.7G/122.7G, Available0, Use100%. No cleanup answer, deletion or full-disk
Linux/SQL rerun. RP-342/D-023 draft unruled; no startup Retry constructed.
Docs/ledger/inventory/plan/board/queue/checkpoint log reconciled. Next accepted
work: predeclare GS0.4/0.6 forced lifecycle focus truth audit from mounted
Garage surfaces. Current lifecycle handler switches surface; current Desk
unit/browser assertions do not assert native forced focus. SOURCE CANDIDATE
only, not an executed defect; determine authority/populations before fix.
Body/GS4/full AC7/default-player/AT/Firefox/privacy/platform/numeric/full-nine-
tier1.0 and every designated review remain. No checkbox/lifecycle/archive/
push/mint/deploy/release/shortened-preview substitution. Goal active/progress.

## 2026-10-07 — RP-082 current Garage lifecycle focus predeclaration

At fbea2158 clean checkout; preceding turn progressed RP-343 repair and
RP-344 discriminating tests. TEST-ONLY under accepted GS0.4/0.6 (the active
follow-up's own floor, explicitly independent of accepting the broader draft
Accessibility RFC at its final sequencing note). RP-082 already owns the
older measured Offer/Run-End focus defect; do not create a duplicate ledger ID.

Three mounted surfaces (Fiscal, Trophy Case, Meters), persistent nav focus
versus focused soon-removed surface heading, Offer versus Run-End destination,
320/1280, Chromium/WebKit: twenty-four declarations / forty-eight executions.
Production envelope/lifecycle decoders admit public fixtures before runtime-
double delivery. Native Enter selects source surface; explicit heading focus
is a controlled removed-target population, not a physical screen-reader walk.
Require actual new destination, exact visible known heading, focus on that
heading after render, non-tab-stop tabindex=-1, no gameplay intents, native
axe and reflow. Do not credit screen selection alone as forced focus.

Separate cancellation controls from Meters: after delivering Offer or Run-End,
select/focus Settings synchronously before Svelte renders; newer choice must
cancel stale pending focus. Four declarations / eight native executions.
Two ordered controls (320/1280, four native executions): Offer followed by
newer Run-End in one callback, then a Settings choice and replay of old Offer/
Run-End cursors. Newest actual destination owns focus; consumed replay must
not steal focus or change surface. These are controlled same-callback races,
not claims about physical OS input timing, all lifecycle contexts or real auth.

Run unchanged source first, record actual red or absence honestly. Any repair
requires separate product predeclaration, existing copy/payload/navigation
precedence only. No permanent product/clock/auth/timer/snapshot/value/CI/Make/
balance/RFC body/copy change under this range. No budgets/assertions loosened.
Keep all earlier tests. Whole fbea2158-exclusive through eventual records
needs Claude separately from 9f485993..fbea2158 and earlier spans. Capacity/
Firefox/AT/body/privacy/platform/numeric/full-nine-tier1.0/review holds remain;
cleanup/startup questions unanswered, no deletion/Retry/archive/push/mint/
release authority inferred. Goal active/progress.

## 2026-10-07 — RP-082 current native red checkpoint

3451fdb0 predeclared.735986→59011c types terminal 0, zero errors/warnings.
66150f→06b9e3 native terminal 2: 52 failures / 8 passes / 414 explicitly
unselected. All 48 source-surface populations fail actual focus AFTER exact
decoded event destination/visible heading/source-heading removal controls
pass. Persistent nav remains BUTTON; removed focused heading leaves BODY.
Four ordered Offer→Run-End controls also fail with BUTTON. Eight controlled
newer-Settings selection/replay controls pass. No timeout/import/type error.
Public envelopes and lifecycle payloads decode before delivery; synthetic
runtime is not real service/AT/physical OS event timing.

RP-082 updated immediately, not duplicated. No product/catalog/copy/clock/
auth/payload/precedence/CI/Make/RFC body changed. This test-only checkpoint
records a genuine red population, not a green gate or new accepted draft.
Separate accepted-product scope follows before any repair. Complete range
after fbea2158 needs Claude, all earlier spans independently owed.
