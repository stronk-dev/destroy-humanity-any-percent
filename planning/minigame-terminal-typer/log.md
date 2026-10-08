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

## 2026-10-04 — Codex targeted TT8/TT9 review predeclaration

**Reviewed by:** Codex, independent of Claude's `ade1083b` implementation. **Scope:** the
`TyperTable` child and its browser witness only; this is not a verdict on B1–B7 or an archival
approval. RP-149 records the observed gap before edits.

The accepted TT8.3 requires focus to remain in the input while one polite announcement names each
new prompt. The original markup had a static `<code>` and a polite status for miss/clear feedback,
but no live prompt announcement. The test named "begins by keyboard" focused a button then called
`.click()`. The predeclared negative was a same-mount prompt change requiring exactly one polite
prompt region with the new command while input focus remains, plus real Enter activation for ready
mode and submit. It had to fail before correction and after severing it. This did not authorize
Typer API arms, registry pinning, content mint, or a full AC13 release claim.

### Targeted finding and correction — Codex

**Review by:** Codex of Claude's `ade1083b` TT8 child/test only. **Recorded by:** Codex.
**Verdict:** CHANGES REQUIRED on the prompt-announcement and keyboard-evidence slice; no full
Typer B1–B7 verdict is inferred. The new same-mount prompt-change browser witness failed on the
original child: zero `.prompt-announcement[aria-live=polite]` elements. The claimed keyboard
begin test used a programmatic `.click()`; it now uses real browser Enter, as does line submission.

The correction keeps an initially empty polite live region mounted and updates its text when the
server-owned current prompt ID changes. It introduces no authored prose or gameplay rule. The
test-only `TyperTableHarness` advances the snapshot on the same mounted child and checks the live
text changes exactly once while input focus stays put. After correction, removing only the
announcement assignment failed the focused browser case at `expected '' to contain 'ls -la'`;
the line was restored and the suite passed again.

Cold evidence at the TT8 corrective `9b7137c2` worktree: Chromium and WebKit full browser
populations each passed 6,986 tests (one existing skip each) plus the performance lane;
`make test-client` passed 6,905 (82 skips); `make typecheck` reported zero errors/warnings;
`make verify-client-boundary`, `make copy-check` and `make build-client` passed. The default
Firefox-inclusive gate and manual screen-reader evidence are **not** claimed. This Codex-authored
corrective range requires Claude's designated cross-party review before the B6 slice can be
accepted; TT-PA4/C2 public wire, registration, provisional content and B7/AC13 remain open.

## 2026-10-04 — Codex TT4.4 raw-text differential predeclaration

**Reviewed by:** Codex, targeting Claude's B1 engine `345dc0b9`; **scope:** TT4.4 text validity
only, not a full B1 or Typer verdict. TT4.4 says valid UTF-8 excluding C0/DEL; U+FFFD encoded as
`EF BF BD` is valid UTF-8 and should be a scored miss against an ASCII prompt. The original Go and
TS validators explicitly rejected U+FFFD, recorded in B1 as a "DESIGN-GAP" even though the
accepted text rule names the admissible set. The critical discriminator was not merely allowing
U+FFFD after JSON parse: malformed raw UTF-8 and unpaired JSON surrogate escapes must still
reject `invalid_text` before Go's decoder can replace them with that same rune. The predeclared
three-arm decode test required the valid arm to fail on original HEAD while malformed arms reject.
A correction had to preserve Go/TS parity and account for the historical kernel-version gate.

### TT4.4 finding, correction and executed evidence — Codex

**Review by:** Codex of Claude's `345dc0b9` B1 text-validity slice. **Recorded by:** Codex.
**Verdict:** CHANGES REQUIRED on the U+FFFD rule; no full B1/B1–B7 approval is inferred.
The three-arm temporary Go probe failed on the valid raw U+FFFD arm with `invalid_text`, while
malformed raw UTF-8 and an unpaired `\ud800` rejected. Persistent Go and TS tests were then
added; both suites failed on the valid-U+FFFD arm before correction. The B1 log's prior
"DESIGN-GAP" label did not override TT4.4's explicit valid-UTF-8 rule.

The correction checks raw submit-line JSON before Go's decoder can substitute malformed bytes
or lone surrogate escapes, then permits a legitimately encoded U+FFFD in both runtimes. It
rejects malformed raw UTF-8, isolated high/low surrogate escapes and a high surrogate followed
by a non-low escape; a valid surrogate pair and an escaped literal backslash remain accepted.
The shared Typer corpus now contains the valid-U+FFFD miss, increasing its fixed transition budget
from 61 to 62 without changing content identity. Removing the new raw validator after correction
made the Go test fail because malformed input was accepted; the guard was restored.
`kernel/VERSION` and Go/TS constants move together from 0.3.136 to 0.3.137 in `c9040bbe`.

Cold evidence: Go `./typer ./minigame ./gameserver ./replaycatalog` and vet pass, Typer corpus
check passes, client suite passes 6,906 tests, typecheck has zero diagnostics, production client
build passes, and full Chromium/WebKit browser populations each pass 6,987 tests plus the
performance lane. The first Docker selector accidentally targeted `./gameserver ./minigame` and
printed `no tests to run`; it is **not** counted. The corrected real-Postgres `./production`
selector executes `TestTyperComposedIntegrationUnlockPlayPayoutAndNeutrality` PASS, crediting
3e0 in both timed and untimed modes. The default Firefox-inclusive gate remains open. This
Codex corrective range requires Claude's designated cross-party review before B1 acceptance.

### Post-commit kernel-history gate

At committed `c9040bbe`, `make verify-kernel-version` exits 2 at earlier
`50a3a514` against `0cf9f7a6`, naming six guarded `client/src/minigame/` paths
without a same-commit bump. The new Typer commit itself includes
`kernel/VERSION` 0.3.136→0.3.137 and matching Go/TS constants, but the
history walk halts before a whole-history green verdict. This is the
pre-existing CI defect, not a passed gate and not authority to create a
kernel-history exception without the pending owner ruling.

## 2026-10-04 — Codex B2 TT-PA1 designated-review predeclaration

**Review by:** Codex of Claude's exact B2 commit `3eb7e401^..3eb7e401`;
**recorded by:** Codex. This review targets only the server-sampled command-time
plumbing (claim → tenant → persisted command, including terminal resolution →
replay), not the full Typer RFC or later changes. Recheck the committed diff
against current code; execute the named real-Postgres AC3 population cold with
`-count=1` and verify the test actually runs. Then sever the command-log write
to resample at insert time and require that population to fail on stamp
divergence. Inspect the terminal and replay paths for any way a client time
field enters the tenant or a second sample replaces the claimed value. A
surviving mutant or a skipped test is CHANGES REQUIRED, not an approval.

### B2 review result and injected-clock correction predeclaration

**Review by:** Codex of `3eb7e401^..3eb7e401`; **recorded by:** Codex.
**Verdict:** CHANGES REQUIRED for AC3's exact injected-clock population (RP-151), not for the
sample plumbing itself. Cold real-Postgres `TestServerTimeSampleReachesTenantAndLogIntegration`
ran and passed. Independent mutants all failed as intended: nonterminal insert-time resampling
failed replay; terminal insert-time resampling failed the command-3 stamp equality; replay with
`ServerTimeMs: 0` failed replay verification. Each mutation was restored, and the witness passed
again. Source inspection found no client time field in `PlayRequest`; claim samples once in its
transaction and the persisted nonterminal/terminal writes take that sample.

The current witness nevertheless uses the real wall clock plus a 25 ms tenant stall, whereas
accepted AC3 names an *injected DB clock*. Predeclare a DB-local temporary sequence with exact
values 1000001, 1000002, 1000003, forced through a single test connection. First assert that
sequence on the current production sampler, which must fail. Then add a repository-local
test-injectable clock-query seam defaulting exactly to the current production SQL; the test sets
it to `nextval` of the temporary DB sequence. The tenant, persisted rows and replay must each
carry those exact values, including terminal. A mutant ignoring the injected query must fail the
new exact oracle. No client input or new game mechanic is authorized.

### B2 injected-clock attempt — retained negative result

The test-only DB sequence `1000001…1000003` plus an exact tenant/log oracle failed against
the original sampler at command 1 (`tenant/log 1791134241942`, expected `1000001`). An
unexported repository query seam defaulting to the production SQL let the test inject
`SELECT nextval('pg_temp.typer_clock_probe')::bigint`; it passed with all three exact stamps.
Severing that seam back to the production constant failed at command 1 again. This proves the
exact oracle can discriminate.

The seam was **withdrawn**, not committed: it changed `server/minigame/session.go`, a guarded
kernel path, solely for testability while keeping production behavior identical. That would
force a false kernel-version signal under the repository's guard. The exact-sequence test and
all temporary product mutations were removed; a cold real-Postgres rerun of the original B2
witness passed after restoration. `git diff` has no product or test changes from this review.
RP-151 remains CHANGES REQUIRED until a lawful test-only DB-clock injection is found or the
contract/versioning authority explicitly rules a different path. The three executed severing
probes support the existing mechanism, but do not erase AC3's named population requirement.

### Test-only alternative predeclaration

Before treating RP-151 as waiting on authority, test whether PostgreSQL can shadow the
unqualified `clock_timestamp()` in the existing sample SQL with a `pg_temp` function on one
pooled connection and an explicit `search_path`. The preflight must show the exact injected
sample value from the unchanged production query. If it does, add a separate composed
fixed-clock population while retaining the original real-clock/stall witness for its
insert-time resampling negative. No guarded production bytes or kernel version may change.

### B2 injected-clock witness — test-only correction

The first test-only `pg_temp.clock_timestamp()` attempt did **not** shadow PostgreSQL's
unqualified function lookup: production SQL still returned the real clock, and the preflight
failed. An explicitly first-listed regular test schema did shadow it; a preflight confirmed
the unchanged `sampleServerMSSQL` returned the test function's value. A fixed 1970 timestamp
then hit the database's monotonic session-time trigger; a future fixed timestamp passed but
would age badly. The final fixture stores one DB-derived time exactly one day ahead in an
isolated, per-test schema; its shadow function returns that fixed value. The test pins one
connection, verifies the injected SQL sample is at least 23 hours beyond the pre-injection
baseline, then checks all three tenant and persisted command stamps equal it, terminal
included, while verification replay succeeds. Cleanup drops the exact test-created schema
and fails loudly if that fails.

Cold real-Postgres runs of both `TestServerTimeSampleReachesTenantAndLogIntegration` and
`TestServerTimeInjectedDatabaseClockIntegration` pass with `-count=1`; `go vet ./minigame`
passes. Explicitly qualifying the production sample as `pg_catalog.clock_timestamp()`
bypassed the injected function and failed the new preflight (`sample` only 4 ms beyond
baseline), then was restored. The prior three independent nonterminal/terminal/replay
severing probes remain recorded above. Only `server/minigame/server_time_test.go` changes in
the supplemental code range: no guarded production byte, kernel version, API or game
behavior changed.

After the final test-only edit, the whole `./minigame` real-Postgres `Integration`
population passed cold (`make test-save-integration SAVE_TEST_PACKAGES='./minigame'`), as did
the non-Postgres `./minigame` package and vet. The exact injected-clock and real-clock tests
were also run together verbosely, both named and both passing; no silent skip is counted.

**Review by:** Codex of Claude's bounded B2 mechanism; **recorded by:** Codex. The original
mechanism has strong executed evidence, but B2/AC3 is not closed on Codex's self-review of
the new supplemental test. Claude must cross-party review the exact Codex test range before
any plan/RFC promotion or archival. This does not approve B1–B7 as a whole.

## 2026-10-04 — Codex B3 TT-PA2 loader review and bounded parity correction

**Review by:** Codex, inspecting Claude's `885237a7` TT-PA2 loader/start batch. **Recorded by:** Codex. This is a targeted review, not a full B3/B4–B7 range-union verdict or Typer archival approval.

The optional `exit_history_at_least` field was decoded into `*int64`. Explicit JSON `null` therefore became `nil` and silently disabled the first-Exit gate in Go, while the TypeScript loader's `safeInteger` rejected the same row. A new Go negative for `null` failed first (`TestTierAtLeastUnlockArm`: accepted); the matching TypeScript negative was added. The Go loader now rejects the literal null before treating the field as optional. `docs/minigame-platform.md` records the exact behavior, and the kernel identity advances to 0.3.138 because `server/minigame/catalog.go` is watched.

Cold `make test-go GO_PACKAGES='./kernel ./minigame ./production' GO_TEST_FLAGS='-count=1'`, `make test-client`, `make typecheck`, and `make vet` pass. `make verify-kernel-version` still fails at historical commit `50a3a514` against `0cf9f7a6` (the same pre-existing, separately tracked history defect), before it can judge this new change; no green kernel-history claim is made. RP-152 records the parity defect. Claude must independently review this Codex correction before it can count as designated-approved.

Separate RP-153: current `server/gameui/features.go` projects any non-Fiscal minigame row as `unlocked:true`, even though TT-PA2's composed start rejects Typer at Tier 0 and before its first Exit. The Garage GS7 accepted surface is Pitch-only, the Typer route and production pin are absent, and no UI-availability implementation is inferred from B3's server-start authority. This remains an explicit successor contract/proof boundary before a Typer player-facing unlock claim.

### B3 production API-path witness (RP-154)

The previously recorded real-Postgres AC9 witness called `StartMinigameSession` (legacy direct path), while B3's claimed resolver is `StartMinigameAPISession`. The legacy population passed cold, but did not exercise the API function. A separate test now pins the same Typer fixture chain with Founder v21 and drives the production API resolver for Tier 0/one Exit, Tier 1/no Exit, and Tier 1/one Exit; the existing legacy population retains Founder v20. Rejections assert the precise sentinel, no receipt, no session, and no Founder sequence/revision advance; eligible start asserts a Typer session and receipt.

Both Typer integration populations pass cold on real Postgres. A temporary Tier override to 9 in the API resolver made the Tier-0 row create a session and fail the test; a separate temporary Exit-count override to 99 made the no-Exit row create a session and fail. Both mutations were restored, and `git diff` shows no residual production change. This is test evidence for the internal API resolver, not proof of a public Typer route or Game UI workflow. Claude's designated review of this new Codex test range remains required.

## 2026-10-04 — Codex B4 TT-PA3 loader-chain parity review

**Review by:** Codex on Claude's `391beb76` B4 loader chain. **Recorded by:** Codex. This is targeted and does not approve the full B4 span or Typer archival.

The Go replay loader checked that the `typer` minigame definition named engine `typer@1.0.0`. The TS replay loader checked only that an ID `typer` existed, so a hash-consistent bundle with `minigame_id:typer` but `engine_ref:pitch` was accepted by TS and refused by Go. The added TS wrong-engine fixture failed first with a resolved bundle; the bounded correction checks both engine ref and version. A separate temporary removal of the version condition made the wrong-version fixture resolve and fail its test, then was restored. Equivalent wrong-ref/version negatives are added to Go's chain test.

Cold `make test-client`, `make typecheck`, `make test-go GO_PACKAGES='./replaycatalog ./kernel' GO_TEST_FLAGS='-count=1'`, and `make vet` pass. The watched `client/src/replay.ts` change advances the shared kernel identity to 0.3.139, with current canonical Typer docs updated. The historical `50a3a514` kernel-history violation still blocks a green whole-history gate; this range makes no claim to repair it. RP-155 tracks the correction, which requires Claude's cross-party designated review. No public Typer route, production pin or full AC1 proof is claimed.

## 2026-10-04 — Codex B6 TT8.6 child keyboard witness

**Review by:** Codex on the bounded Claude B6 child; **recorded by:** Codex. This is a targeted child check, not a full B6/AC13 or Typer verdict. The existing browser suite drove Begin and Submit with real Enter but used `.click()` for End run and never asserted Exit to host. A test-only case now focuses the native End run button and activates it with Enter, then focuses Leave the table and activates it with Space, checking the exact command and callback. The first draft used “Leave table” rather than the candidate catalog's “Leave the table”; it failed as a test-label mistake, was corrected, and both Chromium and WebKit targeted runs passed 7/7 plus the separate performance lane.

Temporary disconnection of End run made the new case fail with no command; separate disconnection of Exit to host made it fail with zero exits. Both product mutations were restored and `git diff` confirms the child has no residual change. This proves the two child actions' native keyboard wiring, not registration under MA3, the public Typer wire, a full keyboard-only run, manual assistive-technology use or 400% browser zoom. Claude must cross-party review this Codex test range.

## 2026-10-04 — Codex B7 AC8 live Exit witness

**Review by:** Codex on Claude's B7 service-level composed path; **recorded by:** Codex. This is a bounded AC8 correction, not a full B7/AC8 or Typer verdict.

The previous Typer test observed `repository.ActiveMinigame` before and after `end_run` but never called the production Exit path. Its test service lacked `WithMinigameActivity`, and its synthetic one-prior-Exit Founder was paired with Company `run_seq:1`. A newly added `wind_down` request failed first with `minigame activity resolver unavailable`, proving the old fixture could not witness the claimed gate. The fixture now binds the real repository, uses `run_seq = len(exit_history)+1` for the pinned run, and keeps a v21 Founder and non-future Fiscal clock for the actual Exit path. During diagnosis, a future resolution timestamp triggered Fiscal clock regression, and a v20 Founder failed the pinned minigame API save floor; neither was hidden as product evidence. A temporary diagnostic error-detail edit in `founder_exit_replay.go` was restored.

Cold real-Postgres `TestTyper` populations now pass: the open session's actual `wind_down` returns `not_eligible/minigame_session_active` without advancing Company or Founder; after `end_run` and resolution, a second `wind_down` applies and appends exactly one Founder Exit. Forcing the production activity result to false made the active Exit apply and failed the new assertion, then was restored; `git diff` has no production change. Cold non-Postgres `./production` and `go vet ./...` pass. This closes only the B7 Exit-before/after service-level evidence gap (RP-157). Faucet-cap forfeit, offline-quality charge and the MA-endpoint journey remain open under AC8; Claude cross-party review of this Codex test range is required.

## 2026-10-04 — Codex B7 AC8 faucet-cap and offline-quality witness

**Review by:** Codex on Claude's B7 service-level composed path; **recorded by:** Codex. Bounded first-filter correction only, not a designated cross-party review or full AC8 verdict.

The one-run composed test never reached the five-sends-per-day boundary and did not assert the offline-quality grade. The test now plays six same-day Typer sessions on one Founder against real Postgres; each run compares the Company cash delta with its terminal receipt, the first five credit, and the sixth credits zero with positive `configured_cap_forfeit_units` and exact `cap.minigame_faucet` reason. After the first clean-seven result, the Founder's Typer quality is 800,000 ppm. A compile-error in the first draft (treating `Ledger.Balance`'s boolean as an error) was fixed before the cold green run.

Discrimination: changing the live faucet comparison from `<` to `<=` let the sixth run credit `4e0` and failed the sixth-send assertion. Changing the quality transition from `grade` to `grade - 1` failed the 800,000-ppm assertion. Both mutations were restored with no production diff. An earlier mutation to the replay validator's comparison survived this test; it was not cited as a successful probe, and its narrower replay-validation implications remain for separate review. Cold real-Postgres `TestTyper` populations, non-Postgres `./production -count=1`, and `go vet ./...` pass. RP-158 records the bounded proof. B7 remains partial because AC8's public MA create/begin/submit/terminal/retry path needs the unresolved Typer public-wire decision; Claude's designated review of this Codex test range is required.

## 2026-10-04 — Codex B7 replay quota-forgery negative

**Review by:** Codex on the bounded replay behavior; **recorded by:** Codex. The earlier surviving `validateFaucetReplay` mutation broadened acceptance only for forged quota records, so valid-history integration was not the right discriminator. A separate fixture-derived unit test now admits the honest exhausted-quota forfeit and rejects a forged sixth credit with `quota_after` advanced. Repeating the `quota_before < sends_per_day` to `<=` mutation made this test fail on the honest exhausted record; it was restored. Cold `./production -count=1`, real-Postgres `TestTyper`, and vet pass with no residual production diff. RP-159 closes this narrow validation-test gap, but does not claim a forged persisted-log end-to-end proof or full AC8. Claude's cross-party designated review of the Codex test range remains required.

## 2026-10-04 — B5 partial API error details designated review

**Verdict: APPROVED, bounded to the error-detail increment. Review by:** Codex on Claude's exact `d7c1ce6e^..d7c1ce6e` range. **Recorded by:** Codex. This is the cross-party designated pass for that commit, not an approval of the unimplemented TT-PA4 request/snapshot arms, the full B5 batch or Typer archival.

The range changes the API error enum, exact response registry, handler mappings, generated contract/types, test rows and canonical Typer status note. The five new detail literals match TT-PA2/TT-PA4; the tier and curriculum sentinels precede the generic `ErrInvalidIntent` 400 arm. `make test-go GO_PACKAGES='./account ./publicapi' GO_TEST_FLAGS='-count=1'` and `make api-check` pass cold; regenerated API artifacts and the v1 compatibility pin remain byte-identical. Changing the tier mapping to the already-valid but wrong `fiscal_unlock_required` detail makes `TestMinigameDeterministicErrorTableIsClosed` fail on the exact expected 409 body; the mutation was restored and the tree is clean. The accepted API C2 versus TT-PA4 public-wire conflict is genuine and remains with the ruling authors. No public Typer route, snapshot arm, generated command consumer or AC8 endpoint journey is inferred from this verdict.

## 2026-10-04 — Codex B1 AC12 corpus-budget review finding

**Review by:** Codex on Claude's B1 `345dc0b9^..345dc0b9`; **recorded by:** Codex. This is a targeted CHANGES REQUIRED finding for AC12, not a verdict on the full 3,703-line B1 range.

The accepted AC12 defines the fixed transition budget as the sum of the corpus's command counts. The B1 generator incremented only on `expect == "applied"`, and the TS consumer likewise incremented only after applied commands. The checked-in corpus contained 67 commands, including five rejected attempts, but advertised 62. A new Go equality assertion failed first with `budget=62 must count all 67 commands`. The bounded test-only correction counts every scenario step in Go, regenerates the budget to 67 without changing any scenario, and increments the TS count before each attempt. Cold `make test-go GO_PACKAGES='./typer' GO_TEST_FLAGS='-count=1'` (including the content gate), full `make test-client` (6,906 passed), `make typecheck` and `make vet` pass; `make typer-corpus-check` also passes but reused Go's cache. Restoring the old TS applied-only counter made its content-gate test fail `expected 62 to be 67`; that mutation was restored. RP-160 tracks the finding. No engine, content-policy or kernel-affecting byte changed; the historical kernel-history red remains separate. Claude must cross-party review this Codex correction before B1/AC12 can close, and Codex still owes review of the rest of B1.

## 2026-10-04 — Codex B1 TT1 scaling-clamp review finding

**Review by:** Codex on Claude's B1 `345dc0b9^..345dc0b9`; **recorded by:** Codex. This is a targeted CHANGES REQUIRED finding and bounded correction, not an approval of the full B1 implementation.

TT1's scaling row clamps `typer.era_tier` to 1..9, but the Go and TS pure engines each accepted zero if the otherwise-valid corpus had a full tier-zero prompt pool; Go's standalone snapshot validator also allowed zero. Matching tier-zero content tests failed first: Go `Create` returned no error and TS `createTyper` resolved with `era_tier:0`. The correction uses the existing named minimum in both engines and Go snapshot validation; separately restoring the old Go snapshot comparison made the new snapshot negative fail, then was restored. Canonical docs state the boundary and correct the stale TT-PA1–PA3 status. This is defense of the pure pinned-input contract, not evidence that the current server route lets a client choose tier zero. Kernel identity advances from 0.3.139 to 0.3.140 in the same product change. Cold `./decimal ./typer ./kernel ./replaycatalog ./minigame`, full client tests (6,907 passing), typecheck, build, vet and decimal-vector check pass. `make verify-kernel-version` remains red at the pre-existing pushed `50a3a514` RP-131 violation before it can judge this worktree; no green whole-history claim is made. RP-161 tracks this correction; Claude's cross-party review of the Codex range is required before B1/TT1 can close.

## 2026-10-04 — Codex B1 TT4.7 result-bound DESIGN-GAP

**Review by:** Codex on Claude's B1 `345dc0b9^..345dc0b9`; **recorded by:** Codex. This is a targeted CHANGES REQUIRED finding, not a full B1 verdict.

TT4.7 states that `ValidateResult` checks `clean_lines ≤ lines_cleared ≤ run_length`. The current `minigame.Tenant.ValidateResult(*Result)` signature has no content, catalog or run-length argument; Typer's implementation checks `clean_lines ≤ lines_cleared` but not the upper bound. A temporary diagnostic supplied a structurally valid `completed` result with nine clean/cleared lines against the fixture's eight-line run; cold `./typer -run TestTyperDiagnosticResultValidatorRejectsOverlongResult -count=1` failed with `validator admitted 9 lines above the fixture's run_length=8`. The diagnostic test was removed and the tree returned clean. Engine transitions still stop at `run_length`, the platform calls this validator on server-produced results, and the faucet independently limits credited value; this finding does not assert a client score-injection path. It does mean the accepted validator claim is false and would not catch an overproducing tenant regression.

**DESIGN-GAP:** the Typer/Minigame Platform ruling authors must choose whether to pass pinned catalog context into result validation (with an appropriate shared interface/replay contract) or explicitly narrow TT4.7 to an engine/replay-owned bound and name the executable witness for that route. A content-free hardcoded `8` would violate the declarative balance law. No product code or normative ruling text was changed. RP-162 owns the gap; B1/AC7 remains partial.

## 2026-10-04 — Codex B1 AC2 prompt-order targeted evidence

**Review by:** Codex on the AC2 mechanism within Claude's B1 `345dc0b9^..345dc0b9`; **recorded by:** Codex. **Targeted evidence verified, no verdict on the full B1 commit/range.** This entry cannot be used as archival range-union approval.

The accepted order is a downward Fisher–Yates from `typer.prompts.v1` after the `typer.run.v1` seed substream. The checked-in Go-generated corpus and TS replay pass cold at current HEAD. Temporarily changing only Go's prompt substream to `typer.run.v1` made `TestTyperContentGate` fail as a stale corpus. The same isolated TS mutation made the replay case fail on the first completed scenario (zero clears instead of eight) and the explicit order case fail on differing dealt prompt IDs. Both mutations were restored, `git status` returned clean, and cold Go content-gate plus full client tests passed again (6,907 passed). This supports AC2's seeded-order mechanism; it does not resolve B1's AC7/RP-162 contract gap, supply a full B1 review, ratify provisional content, or authorize a production mint.

## 2026-10-04 — Codex B1 AC4/AC5 timing targeted evidence

**Review by:** Codex on the AC4/AC5 engine mechanisms within Claude's B1 `345dc0b9^..345dc0b9`; **recorded by:** Codex. Targeted evidence verified, not a full B1 range-union approval.

For AC4, changing Go's timed submit comparison from `t > deadline` to `t >= deadline` made the exact-deadline unit fail because the line ended unscored. The independent TS comparison mutation made its shared-corpus replay fail: the next command was `illegal_phase` instead of the expected applied submit. For AC5, reversing the Go `max(sample,last_server_ms)` comparison made `TestTyperClockNeverRewinds` fail at a backwards sample (10,000 instead of 50,000). Reversing the TS comparison made the shared-corpus terminal snapshot diverge (`last_server_ms` 1,000 instead of 41,000). Each mutation was isolated and restored; `git status` was clean, and cold `./typer -count=1` plus full client tests (6,907 passed) passed again. These results support the exact deadline and monotone-clock mechanisms only. AC3's database time authority, AC7's result-bound DESIGN-GAP, public endpoints and B1's remaining review are separate.

## 2026-10-04 — Codex B1 malformed-snapshot parity correction

**Review by:** Codex on Claude's B1 TS replay path; **recorded by:** Codex. This is a bounded first-filter correction, not the cross-party designated review of the new Codex range or a full B1 verdict.

Go's Typer decoder refuses negative counters, non-enumerated assist levels and malformed nested submission feedback, and its catalog-aware Apply refuses counters above the pinned miss cap. The TS decoder checked fewer conditions. A matching Go test passed cold while the new TS test failed first: `ended_early` resolved with `typer.misses:-1`. The correction validates safe nonnegative counters, exact nullable/enum/submission shape and pinned catalog caps before applying a command. The negative fixtures derive the miss cap from content rather than freezing the provisional number. Temporarily removing only the TS cap check let an above-cap result resolve and failed the new test; separately removing the submission-shape guard let a `miss` with null mismatch index resolve and failed it. Both mutations were restored. Cold Go decimal/Typer/kernel/replay/minigame, full client tests (6,908 passing), typecheck, build, vet and decimal vectors pass. The watched TS engine change advances shared kernel identity to 0.3.141; canonical Typer docs were updated. RP-163 records the gap. The historical kernel-history failure at pushed `50a3a514` remains, and this local range is not whole-history CI green. Claude must cross-party review this Codex correction; B1/AC7 still waits on RP-162's ruling-author contract reconciliation.

## 2026-10-05 — Cold full-browser Typer update-depth failure (RP-187)

**Observed / recorded by:** Codex. The restored Cosmetic motion correction's full Linux
`make test-browser-ci` failed Chromium's `offers timed and untimed without a default and
begins by keyboard` with Svelte `effect_update_depth_exceeded`. WebKit also failed Snake's
batching assertion (RP-188). Final result: 21,097 passed, 2 failed, 3 skipped; Make exited 2
and did not reach the separate performance population. All Cosmetic cases passed in all
three engines and the complete real-server composed target passed both drivers, but neither
substitutes for a full browser-green claim.

The unchanged Typer child effect assigns `sampledAt = monotonicNow(); now = sampledAt`.
Reading a reactive field it just wrote is a candidate feedback loop; it is not yet a
deterministically isolated cause. Next diagnose through the existing injected monotonic clock
and actual browser execution. Do not increase timeouts, weaken keyboard/error assertions or
hide the full-suite failure behind retries. No Typer code changed here.

## 2026-10-05 — RP-187 deterministic child-clock predeclaration

**Review by / recorded by:** Codex; bounded original child range `ade1083b^..ade1083b`.
Accepted TT8.4/TT9 authorize the local monotonic display sample, not a different engine time
or payout rule. Predeclare a test-only increasing clock through the child's existing
`monotonicNow` prop: each read advances it, so equality from browser clock quantization cannot
accidentally stop a feedback effect. Mount ready, then advance to a timed snapshot and a new
server sample/revision; require no update-depth/asynchronous error, unchanged gameplay command
list and correct remaining-time text at each authoritative sample. Keep the native-keyboard,
composition, prompt and error-guard assertions; no deadline or test timeout is relaxed.

If this unchanged-child population reproduces the loop, record that as the verified cause.
Then only the sampling effect may change: compute one local sample and assign both reactive
fields from that non-reactive local, never reading an effect-written state field. No engine,
content, copy, API or kernel behavior change. Reinstating the feedback read must fail the same
deterministic test in all three engines. Restore it and run the entire Typer child population
and full CI-equivalent browser lane, keeping RP-188 and RP-131 independently open as needed.
Claude must designated-review any resulting Codex correction; this does not close B6's public
registration/content/assistive-technology gates.

## 2026-10-05 — B6 deterministic sampling verdict (RP-187)

**Review by / recorded by:** Codex. **Original range reviewed:** `ade1083b^..ade1083b`,
bounded to TT8.4/TT9's local display sampling. **Verdict:** CHANGES REQUIRED.
The retained injected increasing-clock test fails on unchanged production in all three
engines (3 failed, 21 unrelated cases filtered; exit 1). Each reports
`effect_update_depth_exceeded`; the timed state also fails to render after the broken effect
prevents its update. These are assertion and browser error-guard failures, not a passing
console warning. No higher timeout, retry or exception suppression was used.

The original effect registers a dependency on its own written `sampledAt` when it evaluates
`now = sampledAt`. Ordinary browser clock quantization can return the same value on two reads
and mask this cycle; the increasing input prevents that accidental equality. The declared
local-sample correction follows next, with the same test and a reinstated-read negative.

## 2026-10-05 — B6 local sampling correction and restored full browser evidence

**Implemented / recorded by:** Codex. **Review needed by:** Claude, designated cross-party.
The effect now obtains one non-reactive local sample and assigns both clock fields from it;
it still depends on the authoritative sample/revision and retains the existing interval.
No engine, payout, content, copy or public wire changes. The test harness only exposes the
child's existing injected clock for ready/timed/new-response populations.

Executed evidence:

- Unchanged production fails the increasing-clock population in all three engines (the
  failing-first test is committed separately, before the correction).
- Corrected complete Typer child suite: 24/24, including native Begin/Submit/End/Leave keys,
  composition/paste, prompt announcements, misses, remaining time and reflow/axe.
- Reintroduce only `now = sampledAt` while preserving the local sample: the same retained
  test fails in all three engines with update-depth errors and absent timed rendering (exit 1).
  Restore the exact corrected assignment before other verification.
- `make typecheck test-client build-client verify-client-boundary verify-cosmetic-boundary
  verify-no-payment`: PASS, zero diagnostics, 6,953 unit passes/86 browser skips, production
  build and all named gates. The extra unit skip is the new browser-only clock case.
- One cold full `make test-browser-ci` after the correction: PASS, 252 file populations,
  21,102 tests/3 performance-case skips (43.69 seconds), then the separate Chromium performance
  case (1 pass/20 filtered skips, 2.03 seconds). The setup/ordinary assertions were not retried
  or loosened. These counts include shared numeric/unit vectors, not that many player journeys.

The previous red full run remains recorded. Snake's unchanged case passed this later run,
which does not explain or close RP-188; its fixed-delay assumption needs separate diagnosis.
The fresh kernel-history failure recorded in the Cosmetic log still stands (RP-131); its
guard/source was not altered. This is local browser-green, not complete or hosted CI green,
full Typer acceptance or archival. Claude must independently review the Codex correction.

## 2026-10-05 — RP-187 exact designated-review handoff

**Review needed by:** Claude. **Recorded by:** Codex. Exact Codex range:
`a5c3d389^..58bc34ad`, covering the predeclaration, failing-first test/harness, bounded
display-effect fix, canonical docs, reinstated-feedback failure and truthful verification/
planning checkpoint. This is not a designated verdict or full B6 approval. RP-188's unchanged
Snake case and RP-131's historic version guard remain separately open.

## 2026-10-08 — B6 native keyboard and response-focus correction (RP-394)

Outcome: accepted TT8.3/TT8.6/TT9 child keyboard operations remain reachable
while pending; responses retain owned input focus without overriding newer
choices, and completion leaves a reachable host exit. Targeted original B6
review (`ade1083b^..ade1083b`, child keyboard/focus only): CHANGES REQUIRED,
not a full B6 or Typer verdict. Baseline for the new correction is `76b58dc2`.

Regression: unchanged component20251 fails24 checks, but includes two mistaken
new test labels (Submit versus existing Enter copy and a missing feedback prefix).
Correct those test errors before product edits; rerun89425 then fails24 real
checks/two newer-focus controls pass. Pending native disable loses focus/removes
reachability, prompt updates steal focus, terminal removes the input without
handoff, and WebKit native Tab skips mode choices. Chromium's complete untimed
run reaches the real engine's terminal result then fails the Leave-focus check.

Correction: explicit native button tab stops, aria-disabled/busy and handler
guards; pre-render focus capture plus post-render connected/current-state and
newer-focus checks. No engine/time/payout/API/kernel/content/copy changes.
Existing harness gains only pending and Leave controls. Native full-run test
dispatches actual DOM commands into the TS engine with pinned placeholder data
and synthetic times, checking miss/correction, every prompt, exact terminal facts
and native Leave. No direct engine command substitutes for a gameplay control.

Executed: final `make test-browser-focused BROWSER_TEST_FLAGS='test/typer-table-browser.test.ts test/typer-content-gate.test.ts test/minigame-surface-browser.test.ts --project=chromium --project=webkit'`65145
passes72 checks; child46/shared-corpus18/host8. Forward/backward Tab leaves the
child without a trap, pending Enter/Space/direct-click duplicates are refused,
and newer-focus controls pass. Root Node shared-corpus9 passes. Final
`make typecheck build-client verify-client-boundary`35699 passes, zero errors/
warnings; `git diff --check` passes. Earlier added traversal test used ES2023
toReversed; typecheck44075 correctly rejected it. Use ES2022 array-copy/reverse,
then rerun affected browser/types/build rather than changing project settings.

Coverage limits: no public Typer API, actual server command clock, payout,
production content mint, Firefox/AT, whole CI or feature acceptance. No unrelated
composed journey is substituted for the unavailable Typer public path. Existing
TT-PA4/C2, RP-162, owner content and historical version-guard holds remain.
Next: continue the original B1/engine review or other accepted implementation;
public registration still requires its named author/owner contract reconciliation.

Review by: Codex (original targeted cross-party finding; new correction's
implementer first filter). Recorded by: Codex. The entire new range after
`76b58dc2` through this containing component/test/docs/record commit requires
Claude's designated review, independently of the earlier Typer correction ranges.
No B6 checkbox/acceptance/archive/push or 1.0 promotion.

## 2026-10-08 — B1 malformed typing-index panic correction (RP-395)

Accepted TT4.5/TT4.6 and the platform's fail-closed tenant contract. Baseline
`e019e670`. Review by: Codex on the original B1 engine boundary (`345dc0b9^..345dc0b9`);
targeted CHANGES REQUIRED, not a full B1 verdict. Recorded by: Codex.

New cold regression38938 reproduces standalone admission and an index[8]/length8
panic through BOTH real Go tenant and registry Apply. The corrupt snapshot keeps
phase typing/current prompt but sets prompt_index and lines_cleared to prompts_total.
The test reports recovered panics as failures, never as valid refusals. TS already
refuses the same state. Both tests construct a real final-prompt state first and
require genuine completion afterward, deriving run length from pinned content.

One structural guard now refuses exhausted typing before lookup, with existing
ErrInvalidTenant/ErrTenantDivergence and no output/input mutation. No new engine
taxonomy, runtime recovery wrapper, command/snapshot field, content or payout
policy. Shared kernel168->169 in all three mirrors because this is a genuine
replay acceptance-set repair; engine1.0.0/content corpus stay unchanged.

Cold `make test-go GO_PACKAGES='./decimal ./typer ./minigame ./replaycatalog ./kernel' GO_TEST_FLAGS='-count=1'`88068
and affected vet42066 pass; the retained corpus re-observation matches without
regeneration. Real Postgres Typer selector89096 executes replay-quota refusal,
server-owned start eligibility, unlock/play/payout/Exit/neutrality (timed3e0,
untimed3e0), plus the tenant row. These are service-level witnesses, not public
Typer command/receipt HTTP or browser integration. Node corpus10 passes; final
affected Chromium/WebKit99263 passes74. Types/build/boundaries35839 pass with
zero errors/warnings; diff check passes.

Broader client88284 is RED:10017 tests pass/831 skips but Vitest collects two
Node-test tool files and rejects them as "No test suite found" (RP-396). Their
Node assertions actually pass25+5; no exclusion/runner fix is hidden in this
engine commit. Its chained later checks did not run, so those were executed
separately above. Kernel99372 remains RED at historical pushed50a3a514 after
its checkout/topology fixtures pass; it does not approve the new bump or full CI.

New correction first filter by: Codex (implementer). Recorded by: Codex.
Exact new range after `e019e670` through this containing engine/test/version/docs/
record commit needs Claude, independently of B6 and earlier B1 repairs. RP-162,
public C2/content/AT/whole feature/hosted gates remain; no checkbox/archive/push.
Next: fix RP-396 in the CI owner lane without dropping the existing Node tests,
then continue the original engine review and unresolved author contract work.

## 2026-10-08 — B1 strict snapshot value admission (RP-400)

TT4.5/TT4.6 correction: nonnullable numbers cannot be null; nonnull feedback
requires both keys, including the cleared line's explicit null index. Old-source
regression76783 fails13 cases across actual engine-produced ready/typing/terminal
snapshots. Go silently coerced null counters to zero or admitted missing feedback;
TS already refused. Standalone validation and real tenant/registry Apply are
exercised, with valid phase controls and no-output/no-input-mutation refusals.
The bounded decoder correction and matched TS regression preserve existing
taxonomy, content, commands, time and payout policy. Kernel169->170 in all mirrors
is a genuine replay acceptance-set change, not a fix for historical RP-131.

Final cold five-package Go82200 and vet23861 pass. Real Postgres75173 executes
quota refusal, pinned-tier/start/Exit eligibility, unlock/play/payout/neutrality
(timed3e0/untimed3e0), and tenant-row tests: four top-level tests, not public Typer
HTTP/browser integration. Native Chromium/WebKit47183 passes76 checks; Node
shared-corpus11 passes. Types/build/boundaries/corpus67712 pass with zero errors/
warnings. Initial native launch was sandbox EPERM, before any test; the authorized
local-listener run above executes the population. Diff check passes.

Review by: Codex (original B1 targeted finding; correction implementer first filter).
Recorded by: Codex. Exact new range `e2b391e7` exclusive through this containing
commit needs Claude's designated review. Original B1 remains partial; RP-162,
public C2/content mint, full AT/feature/hosted CI and other review spans stay open.
No checkbox, acceptance, archival or push. Next: finish the remaining accepted
engine review and consolidate related correction ranges for cross-party review.

## 2026-10-08 — B1 runtime solo-mode parity (RP-401)

TT1/TT4.1 correction: TS application must enforce the same solo-only runtime
identity as creation and Go. Old TS mode-by-phase regression fails all nine
cases (chunk56cdb2): async_snapshot, unknown ranked and empty mode across actual
ready/typing/terminal states. Ready/typing incorrectly emit terminal results;
terminal reaches illegal_phase rather than identity refusal. Go tenant/registry
companions16359 pass unchanged. An intermediate test edit had an extra brace;
its transform failure is a test-author error, not product evidence.

One TS mode guard now refuses before snapshot/command application. Valid solo
states remain admitted, including terminal's existing illegal_phase behavior.
Kernel170->171 is the genuine TS replay acceptance-set correction; no server
semantics, content, commands, time/payout rules or corpus bytes changed.

Finished batch: Node54644 passes21, native Chromium/WebKit49787 passes96,
cold four-package Go14423/vet25132 pass, types/build/boundaries/corpus7479 pass
with zero errors/warnings. Diff check passes. No persistence boundary changed;
the previous Postgres service evidence is not relabeled as a new public journey.

Review by: Codex (original B1 targeted finding; new correction implementer first
filter). Recorded by: Codex. Exact new range `ba1365c9` exclusive through this
containing commit requires Claude's designated review. B1 remains partial with
RP-162 and the remaining review; public C2/content/AT/whole feature/CI remain
open. No checkbox, acceptance, archive or push. Continue the accepted engine
review; held public API contracts cannot be invented to complete registration.

## 2026-10-08 — B1 exact catalog tier admission (RP-402)

TT3 correction: era min_tier must be a number, not null coerced to zero. The
corrected old-source regression22999 fails only null: loader accepts tier0 and
both tenant/registry create a ready snapshot. Genuine numeric0/1 and seven other
invalid type/range controls pass; TS already refuses null in loader and creation.
Initial65503 also wrongly expected the tenant's error class at the registry
boundary; corrected to its existing non-taxonomy-error divergence, then reran
before any product edit. Those seven test-author failures are not product evidence.

The bounded raw-era admission check preserves numeric zero and the existing
range/reachability rules. Kernel171->172 is a genuine loader acceptance-set bump;
no content bytes, policy, command, API, time or payout behavior changed.
Cold four-package Go5709/vet57816 pass. Node76756 passes22 and native Chromium/
WebKit58832 passes98. Types/build/boundaries/corpus19389 pass with zero errors/
warnings. Real Postgres58289 executes all four Typer service/tenant-row tests,
including start refusal, unlock/play/Exit, replay quota and timed/untimed payout
neutrality. These service checks are not the still-unavailable public Typer
HTTP/browser journey. Diff check passes.

Review by: Codex (original B1 targeted finding; correction implementer first
filter). Recorded by: Codex. Exact new range `ed92ede7` exclusive through this
containing commit needs Claude's designated review. B1, RP-162, public C2/content
mint, AT, full CI and older review spans remain open. No checkbox/acceptance/
archive/push. Continue the remaining accepted review; the separate first-read
Retry posture question is delivered, not an owner ruling or implementation grant.
