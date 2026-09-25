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
