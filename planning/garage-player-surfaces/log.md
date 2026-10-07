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
