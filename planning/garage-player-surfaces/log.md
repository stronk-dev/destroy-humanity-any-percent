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
