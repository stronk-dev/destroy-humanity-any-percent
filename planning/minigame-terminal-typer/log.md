# Terminal Typer log

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude, per the owner's 2026-09-25 acceptance and implement direction.
**Review:** Codex designated cross-party review required before any archival. Nothing here is
self-approved.

Batches B1–B7 as in `plan.md`. Each batch lands with failing-first tests, a severing probe per
gate, and cold `-count=1` runs. Copy: mechanical keys ship with implementer-drafted candidate text
in `copy/catalog/typer-candidate.json`. That text is owner-adoption material, not ruled copy.
Prompt `text` rows in the fixture are PROVISIONAL placeholders (OD-9), not for production mint.

## 2026-09-25 — B1: content + pure engine (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated review.

Landed:
- `server/typer` and `client/src/typer`, registered in `kernel/affecting-paths.json`; kernel bumped
  0.3.102 → 0.3.103 in this commit.
- `ApplyInput.ServerTimeMs` (TT-PA1 item 2), which Pitch ignores.
- Fixture `balance/testdata/typer-v1.json` with 15 placeholder prompts.
- Candidate copy `copy/catalog/typer-candidate.json`, which is owner-adoption material.
- The shared corpus, `make typer-corpus`/`typer-corpus-check`, and `docs/minigame-terminal-typer.md`.

Evidence (cold, `-count=1`):
- `./typer ./minigame` pass. `client/test/typer-content-gate.test.ts` passes 6/6.
- Severing probes, each turned red (Go/TS):
  - no `max` on the clock (TestTyperClockNeverRewinds);
  - deadline `>=` (content gate);
  - the run substream used directly, as a compiling mutant; a first, non-compiling attempt is
    discarded (Go content gate; TS 2 failed);
  - clears counted as clean (TestTyperCleanLineAccounting and the gate);
  - unsorted nested `last_submission` fields (TestTyperSnapshotsAreRegistryCanonical);
  - TS case folding;
  - TS truncation instead of `line_too_long`;
  - **TS NFKC.** It first SURVIVED, because no vector used a character that NFKC folds and OD-7
    does not map. A fullwidth look-alike vector was added to Go, TS and the corpus. Rerun: TS
    2 failed.

Implementation decisions, all within ruled text:
- **Validation order.** `invalid_text` is validated at command decode, which is phase-independent
  like Pitch's `hand_too_large`. `illegal_phase` precedes `line_too_long`, because the length
  bound needs the content.
- **Terminal handling.** `end_run` and `timed_out` retain `last_submission`.
- **`ValidateResult`** is content-free and checks structural bounds: facts ≥ 0,
  `clean ≤ cleared`, `assisted ∈ {0,1}`. The engine enforces `run_length` and the hardcaps.

**DESIGN-GAP (minor, recorded):** the RFC does not say whether U+FFFD submitted legitimately is a
miss or `invalid_text`. Both runtimes treat it as `invalid_text`, because Go's decoder cannot tell
it apart from invalid UTF-8.

## 2026-09-25 — B2: TT-PA1 server-sampled command time (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated review.

- **What changed.** `Repository.claim` samples the DB clock in the claim transaction and carries it
  on the claimed session. `play` hands it to the tenant (`ServerTimeMs`). `completePlay` and
  `CompletePlayWithReceipt` insert it (both have a new `serverMS` parameter). The terminal
  resolution carries it in `resolutionIdentity.serverMS`, and `resolveTx` inserts it.
  `validateReplayTx` replays with each row's persisted `server_ts_ms` and with the resolution's
  sample for the terminal command. Kernel 0.3.103 → 0.3.104 (`server/minigame` is guarded).
- **AC3 witness.** `TestServerTimeSampleReachesTenantAndLogIntegration` (real Postgres) uses a
  tenant that records `ServerTimeMs` and stalls 25 ms inside `Apply`. The persisted stamp must
  equal the tenant's stamp for all 3 commands, terminal included, and certified replay must pass.
- **Severing, each turned red:**
  - reinstating the insert-time `clock_timestamp()` read (anchor match confirmed): replay
    diverges;
  - replay that omits the persisted stamps: diverges.
- **Evidence (cold):**
  - `./minigame ./production ./typer ./pitch` unit tests pass.
  - Postgres integration for `./minigame`, `./production ./save` and `./gameserver ./account`
    passes on separate runs.
  - One earlier combined run failed across packages because another agent's concurrent
    `compose.save-test.yml` run restarted and truncated the shared Postgres service. Each package
    passed when re-run alone. This is noted as environment interference, not treated as a pass
    of the combined run.

## 2026-09-25 — B3: TT-PA2 `tier_at_least` unlock arm (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated review.

- **What changed.** The Go loader (`loadUnlockCondition`) and TS loader (`parseUnlock`) accept the
  exact arm with an optional `exit_history_at_least`. `UnlockCondition.TierUnlockFailure` is the
  single predicate. The start coordinator (`StartMinigameAPISession`) rejects with the new
  sentinel errors before Founder transition and tenant creation. Kernel 0.3.104 → 0.3.105.
- **Evidence (cold):**
  - `./minigame ./production` pass. `TestTierAtLeastUnlockArm` covers 4 accepted rows, 7 rejected
    rows and the predicate truth table. The TS `minigame-catalog.test.ts` passes 4/4.
  - Severing, each turned red:
    - Go without the tier bound;
    - Go without the exit clause;
    - TS without the tier integer check (1 failed).
- **Deferred, stated:**
  - The HTTP mapping of the two details and the schema detail enum move to B5 (TT-PA4). The
    concurrent Garage-surfaces agent holds uncommitted edits in `server/account/*_schema.go` and
    the generated API artifacts, and regenerating now would sweep them into this commit. Until
    then a gated start surfaces as an unmapped error; no production bundle pins a tier-gated row
    yet.
  - The composed AC9 start witness (Tier 0 rejects, Tier 1 with an exit starts) lands with B4,
    which makes a Typer row pinnable.
- **Note for other owners:** `server/gameui/features.go`, uncommitted and owned by the Garage agent,
  computes minigame availability with only the `fiscal_unlock` arm. It must call
  `TierUnlockFailure` for Typer's row, or the availability preview will disagree with the server.

## 2026-09-25 — B4: TT-PA3 content resolver, loader chain, composition (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated review.

- **What changed.**
  - `CatalogBundle.Typer` and `CatalogBundle.TenantContent`; `ResolveTenantContent` now delegates
    to it.
  - Bundle validation counts `typer` and requires `minigame_api`.
  - `replaycatalog.Load` loads `typer` and enforces the definition ⟺ artifact ⟺ API-tenant chain,
    including a definition row with no API.
  - The start coordinator checks the requested tenant's own content instead of
    `bundle.Pitch != nil`.
  - `gameserver.Compose` registers the Typer tenant.
  - TS `loadReplayCatalogBundle` mirrors all of it.
  - Fixtures: the TT1 row (`testdata/minigame/pitch-typer-v3.json`) and the two-tenant API
    fixture.
  - Kernel 0.3.105 → 0.3.106.
- **Evidence (cold):**
  - Go `./production ./replaycatalog ./gameserver ./minigame ./typer ./harness` pass.
    `TestLoadTyperChainIsAllOrNothing` covers the complete chain, 6 broken chains and the
    Pitch-only resolution. `TestTyperTenantRowLoads` passes.
  - TS: `replay.test.ts` passes 86/86 (the new chain test has 5 broken-chain cases).
    `test-client` passes 6691. Typecheck is clean.
- **Severing:**
  - Go chain checks disabled: red on "definition without api or artifact". The first attempt
    didn't compile (unused variables) and is not counted; it was redone as a compiling mutant.
  - TS chain check disabled: red.

## 2026-09-25 — B5 (partial): API error details; DESIGN-GAP TT-PA4 vs API Foundation C2 (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated review.

**Landed (C2-legal):**
- **Response-enum widening.** The `APIError` detail enum gains `curriculum_exit_required`,
  `invalid_assist_level`, `invalid_text`, `line_too_long` and `tier_required`.
- **Exact bytes and handler mapping.** The matching exact 409 bytes are in `minigameErrorJSON`.
  `writeMinigameResult` maps `ErrMinigameTierRequired` / `ErrMinigameCurriculumExitRequired`
  ahead of the `ErrInvalidIntent` → 400 fallthrough.
- **Generated artifacts.** Regenerated; `gen-api`'s compatibility check against the committed pin
  passes, and `api-compat-v1.json` is byte-unchanged.
- **Witness.** `TestMinigameDeterministicErrorTableIsClosed` gains 5 rows, each validated against
  the registry's exact bytes.
- **Severing, each turned red:**
  - removing the tier mapping;
  - removing the `line_too_long` exact pair.

**DESIGN-GAP (blocking the rest of TT-PA4, filed for the RFC author, not worked around):**
TT-PA4 says the v1 minigame API "gains the Typer 1.0.0 discriminated snapshot arm and the Typer
command union as request arms (additive, MA-C7)". Accepted API Foundation **C2** rules that
`/api/v1/` is additive-only and "request enums do NOT grow inside v1 … Exceptions are `/v2`, never a
bypass tag". Running the implementation proves the conflict:
- **Command arms.** Adding the Typer arms to `MinigameTenantCommand` grows a request `oneOf`.
  `CheckCompatibility` rejects it (`mode&compatibilityRequest … len(nextOne) != len(oldOne)`).
- **Snapshot arm.** Changing the response `snapshot` from `$ref PitchSnapshot` to a union ref is a
  ref change, which is rejected. `gen-api` failed with "schema MinigameSessionResponse: invalid API
  schema".

Refreshing the pin (`make api-pin`) would be exactly the bypass C2 forbids, so it was not done.
Options for the author:
- **(a) New v1 operations for Typer** (C2 allows new operations), for example
  `create/current/play/resolve` under `/api/v1/minigames/typer/...` with Typer-specific
  request/response schemas. This is additive and keeps the Pitch operations byte-stable.
- **(b) A `/v2` tenant-generic minigame namespace** with snapshot/command unions from day one.
- **(c)** An owner ruling that private-v1 minigame operations are pre-release and re-pinnable. This
  contradicts C2's "never a bypass" and is not recommended.

Recommended: **(a)**, because it keeps the stable Pitch wire unchanged. Until this is ruled, Typer
sessions are fully functional server-side (engine, TT-PA1 time, TT-PA2 unlock, TT-PA3 content and
composition), but there is no v1 HTTP route that can carry a Typer command or snapshot.
Consequences:
- **B6 (UI):** blocked on the transport shape.
- **B7:** the composed AC8 path must use the production service directly rather than the MA
  endpoints.

## 2026-09-25 — B7: composed platform path (AC8 partial, AC9, AC11) (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated review.

- **Legacy start path fix.** `StartMinigameSession`, the non-API start path that the composed
  tests drive, now also applies `TierUnlockFailure`, so neither start path bypasses TT-PA2.
  Kernel 0.3.106 → 0.3.107.
- **`TestTyperComposedIntegrationUnlockPlayPayoutAndNeutrality`** runs on real Postgres with a real
  store and the minigame platform, pinning the complete Typer chain. It checks:
  - Tier 0 rejects with `ErrMinigameTierRequired`, and Tier 1 with no exit rejects with
    `ErrMinigameCurriculumExitRequired` (AC9);
  - Tier 1 with one exit starts, then plays begin, one miss and 8 clears through
    `platform.Play`, using DB-clock stamps (TT-PA1);
  - resolution: certified facts clean 7 / cleared 8 / misses 1; faucet payout into `company.cash`
    equals the receipt's `credited_delta` (3e0); the retry returns identical bytes; Founder history
    is `ReplayVerified` (AC8, first half);
  - a timed and an untimed Founder with identical commands receive identical nonzero credit
    (AC11).
- **Severing, each turned red:**
  - The legacy-path tier check disabled: Tier-0 start accepted, and the test fails.
  - Payout keyed on `typer.assisted` (AC11's named mutant): both modes credited 0, and the test
    fails. Honest note: the mutant was caught by the nonzero clause, not by the timed/untimed
    equality alone. At conversion 0.5 an assisted fact of 1 floors to 0 in the first send.
- **Not yet covered by AC8:** the faucet-cap forfeit, "Exit rejects while active and succeeds
  after `end_run`", and "through the MA endpoints". The last one is blocked by the TT-PA4/C2
  DESIGN-GAP (no v1 Typer route).

### B7 addendum — `end_run` releases the Exit block (Claude)

The composed test also opens a Typer session and checks `ActiveMinigame` is true (the MA-C12 Exit
block). It then plays `end_run` before `begin` (outcome `ended_early`), resolves, and checks
`ActiveMinigame` is false. Severing: making `end_run` legal only while typing turns it red with
`illegal_phase`. Test-only commit (`_test.go` is outside the kernel guard).

## 2026-09-25 — B6 (partial): `TyperTable` child and TT8 accessibility contract (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated review.

`client/src/game-ui/minigame/TyperTable.svelte` follows the TT9 prop contract: `snapshot`,
`serverTimeSample`, `pending`, `era`, `dispatch`, `exitToHost`. It is presentation only.

- **Modes.** Timed and untimed are offered as two buttons, neither preselected. The untimed button
  describes the equal-payout note (`aria-describedby`).
- **Prompt and input.** The current prompt is a labelled `<code>`. The input is a native text field
  with `autocomplete=off`, `autocapitalize=off`, `spellcheck=false` and `autocorrect=off`.
- **Submitting.** Enter submits the form; Enter during IME composition never submits, and paste is
  never blocked. The field clears only when the prompt advances, and focus stays in it.
- **Feedback and time.** A miss is announced as text with a 1-based position (polite status). In
  timed runs the remaining time is plain text (no periodic announcements); once expired, the
  expired state and `end_run` are shown.
- **Reflow.** The prompt wraps (`overflow-wrap:anywhere`), so a 256-byte token reflows at 320 px.

**Evidence:**
- `test/typer-table-browser.test.ts` passes 15/15 across chromium, firefox and webkit: axe on the
  ready, typing, miss and expired states; keyboard begin; composition and paste; text miss
  feedback; time; 320 px reflow.
- `verify-client-boundary` scans 10 component files, all clean.
- AC13's named mutants, each red on chromium (1 failed | 4 passed):
  - seeded paste `preventDefault`;
  - submit during composition;
  - hue-only miss, with the text removed;
  - no prompt wrapping.

**Not yet done:**
- **Registration.** The child isn't registered in `tenant-registry.ts`. The pinned
  `balance/minigame-api/first-content.json` lists only Pitch, and the registry fails closed on a
  child with no pinned tenant. There is also no v1 route carrying Typer commands or snapshots (the
  TT-PA4/C2 gap). A keyboard-only run through the host surface therefore waits on both.
- **Scene lines.** The per-prompt `typer.prompt.<id>.scene` line isn't shown, because the client
  doesn't bundle Typer content. It would follow the Pitch hash-verified pattern once a mint pins
  it.

### 2026-09-25 — Browser-lane fix for the TT-PA3 replay test (Claude)

`make test-browser` failed 2 cases in `client/test/replay.test.ts` in the browser configuration
(first reproduced at `ec47af68`): a dynamic `await import("…/typer-v1.json?raw")` cannot be fetched
by the Vite browser runner. The fix is test-only: a static top-level `?raw` import of the same bytes.
Cold results: browser `replay.test.ts` passes 258/258 across 3 browsers, and node passes 86/86. No
product code changed.
