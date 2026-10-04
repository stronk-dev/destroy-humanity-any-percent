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
