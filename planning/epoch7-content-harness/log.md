# Epoch-7 content-dynamics harness implementation log

Append-only. This is the successor evidence lane registered by the First Content Epoch sign-off.

## 2026-08-14 — candidate-runway acceptance bounce: EH-C11–EH-C14

- Resumed at the owner-ratified T0–T1 bytes. The existing lane owns only the empty registry and
  immutable snapshot generator/validator; no content-dynamics scenario parser, four-arm runner,
  candidate report mode, or report schema exists.
- The mint cannot yet be represented honestly: T01-C14 names a singular `relevance` artifact but
  the repo has two ratified policies and no cycle-free replay-bundle owner; the only promotion
  loader hardcodes the epoch-6 cardinality of sixteen; and the scenario/report bytes remain
  directional rather than exact.
- The single Active-Play pair is additionally underdetermined because the pinned catalog can draw
  a non-window Lucky effect and a fresh Company has no owned building target. A favorable seed or
  synthetic pending row would violate EH-C3.
- Filed EH-C11–EH-C14 in the RFC with concrete proposed contracts. No candidate artifact, epoch
  coordinate, production balance byte, or mint authority was changed. The relevance H-A/H-B LOW
  closures are independent and ready for their designated review.

## 2026-08-10 — implementation bounce: current harness cannot own the four mechanics

- Audited the active harness rather than treating “full bundle loaded” as “full bundle simulated.”
  `harness.Suite` passes only economy, Routes, and Commons into Company `Transition`; its two
  policies cannot execute a Founder Fiscal command, Active-Play schedule, Pitch tenant session, or
  platform payout.
- Adding four milestone rows would therefore be false coverage. Reimplementing the missing math in
  the harness would be worse: it would create the second authority the ApplyLogged and tenant
  boundaries were built to prevent.
- Filed EH-C1–EH-C7 in `rfc/first-content-epoch.md` with concrete proposed contracts: one
  full-bundle production-owned simulation lane, a closed four-arm grammar, literal policies and
  cardinality, an observation/invariant split, the existing separate-commit baseline discipline,
  and pinned-epoch identity.
- No balance bytes, baselines, or runtime semantics changed. Work is blocked only on owner rulings;
  the rest of the forward batch remains independently reviewable.

## 2026-08-10 — implementation audit: active content and historical bytes are absent

- Confirmed the epoch-6 seed has sixteen artifacts but no `opportunities` row, and the minted
  economy catalog has no `active_play` multiplier declarations. The exact replay-catalog load
  therefore has no Active-Play policy. Using the reviewed fixture would violate the ruled
  full-bundle identity instead of measuring live content.
- Confirmed an epoch-seed path alone cannot satisfy EH-C6 after a future mint: its artifact paths
  resolve mutable production files, while historical accepted hashes retain identity but not the
  bytes needed to rerun the old golden.
- Filed EH-C8 with a runner-first/fixture-only implementation lane and a hard requirement that the
  first four-arm golden wait for an epoch that actually pins opportunities. Filed EH-C9 with a
  generated, full-set, content-addressed bundle snapshot contract so every registered golden stays
  executable under its original bytes.
- No runner, golden, registry, production artifact, or balance baseline was authored past these
  unresolved identity contracts.

## 2026-08-10 — implementation handoff: EH-C9 snapshots + EH-C8 empty registry

- **Review by:** Codex self-review only. **Recorded by:** Codex. This is ready for designated
  review as a partial implementation slice; it does not claim the four-arm runner or a golden.
- Added the strict five-coordinate `content_dynamics.v1` registry with zero production entries,
  exactly as EH-C8 requires while epoch 6 has no Opportunities owner. `make content-harness` is
  therefore an honest no-op rather than a skip or synthetic-policy lane.
- Added generated, immutable full-bundle snapshots. Generation accepts only the complete active
  `epochseed.Bundle` at its accepted current epoch, writes content-addressed bytes once, and
  refuses a later rewrite. Loading verifies registry/manifest coordinates, accepted epoch hash,
  raw-byte-sorted declarations, per-file SHA-256, exact directory set, and recomputed bundle hash.
- Adversarial tests cover missing, extra, tampered, manifest-rehashed, hand-subsetted, and rewrite
  attempts. The baseline guard now discovers future content-dynamics golden paths from registry
  history, and its accepted input surface includes the governed content-dynamics tree.
- Normal repository commands passed: focused cold `make test-go
  GO_PACKAGES='./harness ./cmd/balance-harness' GO_TEST_FLAGS='-count=1'`, `make content-harness`,
  and read-only `make harness-check`.
- Still open in this RFC: the production-owned four-arm simulation seams, strict scenario/report
  grammar, literal-cardinality runner, and the separate first golden after a real epoch pins
  Opportunities. No balance artifact or golden was created.

## 2026-08-14 — OWNER RULING: EH-C11–EH-C14 all ACCEPTED as proposed

Ruled by Marco (recorded by Claude from the owner's explicit per-contract selections, each made
with its alternatives presented):

1. **EH-C11 ACCEPTED.** One minted `relevance` artifact: the cumulative T1 schema-v2 policy,
   byte-for-byte, at `balance/relevance/t0-t1.json`. The T0 policy and both scenarios remain
   harness inputs. The strict decoder/validator moves to a cycle-free `server/relevancepolicy`
   package consumed by both harness and replaycatalog; the parsed policy joins `CatalogBundle`.
   Epoch 7 = exactly eighteen sorted artifacts. (Mint-both REJECTED as redundant weight;
   mint-neither REJECTED — it would reopen T01-C14.)
2. **EH-C12 ACCEPTED.** `planning/t0-t1-content/promotion-manifest.candidate.v1.json` in the
   shipped manifest-v1 grammar, status `ratified`, the exact eighteen rows; carryover rows keep
   production hashes and consumed verdicts; loader cardinality generalized; epoch-6 test still
   pins sixteen, a new epoch-7 test pins the exact set, hashes, and recomputed bundle hash.
   (Grammar v2 REJECTED — churn without need.)
3. **EH-C13 ACCEPTED.** Content-dynamics scenario/report schema v1 exactly as proposed: strict
   roots, byte-sorted run IDs and observation rows, canonical Decimal-string values, candidate at
   `testdata/harness/content-dynamics/scenarios/epoch-7-candidate.v1.json`; the derived budget and
   the scenario bytes + hash return for owner pinning before any golden is accepted. (Reusing the
   relevance schema REJECTED — couples unrelated contracts.)
4. **EH-C14 ACCEPTED.** One pinned deterministic founder/run coordinate whose first natural draw
   is duration-bearing, with the named minimum owned target; the engine performs the natural
   spawn/claim and fails loud on mismatch; the harness never manufactures an opportunity. The
   exact coordinate and target return with the scenario hash for owner acceptance. (Runtime seed
   search REJECTED — adapts to content changes instead of failing on them; vacuous-run tolerance
   REJECTED — fail-loud law.)

**The epoch-7 candidate runner is unblocked.** Deliverables back through designated review:
the runner + candidate mode, the eighteen-row manifest, the scenario bytes/hash and C14
coordinate for owner pinning, then the EH-C10 candidate report toward the mint.

## 2026-08-14 — implementation bounce: EH-C15, ratified bundle fails the live replay resolver

- Implemented the cycle-free Relevance owner, eighteen-row candidate manifest, strict four-arm
  scenario/report runner, production-owned simulation boundaries, and candidate Make/CLI lane far
  enough to execute all 69 declared runs. The provisional output was 786 real transitions and 27
  ordered observations; it was deleted rather than presented as evidence after the full-bundle
  validity failure below.
- The Founder-v21 Fiscal arm intentionally uses `ApplyFounderLogged`. It failed closed because the
  exact candidate bundle is not a valid replay bundle: Opportunities pins `1000 + 3000` ms while
  Prestige pins a 5000 ms catch-up ceiling, contradicting the archived strict `interval + lifetime
  > ceiling` law.
- Ran a discriminating in-memory probe through `ReplayCatalogSet.ResolveReplayCatalogs`: ratified
  `minimum_interval_ms: 1000` rejects; changing only that value to the archived, cross-runtime
  fixture's `2500` accepts. A second probe ran the complete 69-run candidate successfully after
  that one change and an 11-second C14 horizon. No source balance byte was edited by either probe.
- Filed EH-C15 in the RFC with the recommended ratification path and the two broader alternatives.
  The existing Opportunities SHA, candidate constants hash, C14 scenario hash, and report cannot
  be honestly pinned until the owner rules this ratified-byte conflict. Implementation changes
  remain local and uncommitted; no golden, production artifact, epoch coordinate, or mint authority
  changed.

## 2026-08-14 — OWNER RULING on EH-C15: keep the 1000 ms spawn gap; raise lifetime

Ruled by Marco (recorded by Claude from the owner's explicit selection, alternatives presented —
the restore-2500 recommendation and the prestige-ceiling arm were both declined):

1. **`minimum_interval_ms` stays `1000`. `lifetime_ms` rises above 4000** so the composition law
   (`minimum + lifetime > catchup_ceiling` = 5000) holds. **Literal: `4500`** — chosen as the
   round value giving the same 500 ms safety margin the archived 2500+3000 corpus carries
   (1000+4500 = 5500 = 2500+3000). Recorded as the ruling's default literal; since the changed
   artifact returns for owner re-ratification anyway, the owner can adjust the literal at that
   gate without reopening this ruling's direction.
2. Consequences accepted: opportunities spawn as frequently as today but linger 1,500 ms longer —
   a player-visible claim-window change. Prestige's `catchup_ceiling_ms` and every other
   carryover artifact stay untouched.
3. **Process:** Codex edits the Opportunities candidate, re-runs its content gate, hands the
   change through designated review; the owner re-ratifies the Opportunities hash; then the
   eighteen-row manifest constants hash, the C14 coordinate (first spawn moves accordingly),
   scenario bytes/hash, and the EH-C10 report are re-derived in that order. No other ratified
   artifact moves.

## 2026-08-14 — Claude designated cross-party review of `{70ef082}` — APPROVED

- **Review by:** Claude. **Recorded by:** Claude. **Range:** `cde32e4..70ef082` = `{70ef082}`.
- The commit is exactly the ruled change and nothing else: `lifetime_ms` 3000 → 4500 in the
  Opportunities candidate, isolated `BALANCE-CHANGE:` subject. Independent SHA-256 matches the
  recorded `d2aab242…c7974`. Executed cold `-count=1`: `./activeplay ./replaycatalog
  ./production` GREEN; `make verify-schema` exit 0. The composition law now holds at the checked
  site (`production/replay.go:194`): 1000 + 4500 = 5500 > 5000. Codex's execution note (first
  spawn unchanged at 3,558 ms, expiry to 8,058 ms) is consistent with the gap staying untouched.
- **No findings. APPROVED.** Ready for owner re-ratification of the Opportunities pin; the
  full-bundle resolution proof lands with the resumed runner derivation, which is the next step
  after ratification.

## 2026-08-14 — OWNER RATIFICATION: Opportunities candidate re-ratified

Ratified by Marco (recorded by Claude from the owner's explicit selection):
`balance/testdata/t0-t1/opportunities-v1.json` at
`sha256:d2aab242f2e5b9a73c11b32981fdb3971b4cbf6d96df3fbf0277abfc1ebc7974` (lifetime 4500,
spawn gap 1000), superseding the prior Opportunities pin. Consumed verdict: `d740318` over
`{70ef082}`. **The EH-C15 derivation chain is unpaused:** Codex re-derives the eighteen-row
manifest constants hash, the C14 coordinate, and the scenario bytes/hash, then produces the
EH-C10 candidate report; the scenario hash and C14 coordinate return for owner pinning, then the
mint sign-off.

## 2026-08-14 — EH-C11–EH-C15 runner + EH-C10 report — READY FOR DESIGNATED REVIEW

- **Review by:** Codex self-review only. **Recorded by:** Codex. This is an implementation
  handoff, not the designated cross-party verdict, owner scenario pin, golden acceptance, archival,
  or mint authorization. **Code/report commits:** `{2a0f7c9, 6c1d423}`; include this docs-tier
  record commit in the designated range.
- Added the cycle-free `relevancepolicy` owner and made the exact cumulative T1 policy a parsed
  member of the replay bundle. The candidate loader now proves the complete eighteen-artifact set
  is live-replay valid, not merely artifact-local loadable. The manifest pins every source hash,
  schema, production path, content gate, and the recomputed constants hash
  `sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789`.
- Added strict scenario/report schema v1 and `make epoch7-content-harness`. The runner executes the
  ruled literal cardinality: one natural Active-Play pair, two Founder-v21 Fiscal sweeps, Pitch
  seeds 0–63, and two Permit cases. Harness code assembles inputs only; production-owned wrappers
  call the real scheduler/claim/accrual, `ApplyFounderLogged`, Pitch tenant plus payout conversion,
  and engine evaluation with frozen Fiscal contributions. The source guard permits these wrappers
  only from harness/tests. The report validator rejects a missing or invented observation.
- **C14 coordinate returned for owner pinning:** founder
  `018f6b7c-9abc-7def-8abc-000000000001`, run sequence `1`, first natural effect
  `active.production`, target `generator.beige_tower`, spawn `3558` attended ms, complete buff
  duration `5000` ms. Scenario SHA-256 is
  `a6df65535eeafd7b96aa017a7bf6666d400058209903d359bb39f5ef7c2f6a54`. EH-C15's lifetime-only
  edit extends opportunity expiry to `8058` ms but does NOT move the spawn; the owner ruling's
  stale “first spawn moves accordingly” process sentence is returned to its author for
  reconciliation rather than silently edited by Codex.
- **EH-C10 owner-facing report:** SHA-256
  `cef5fa23a28e8b16bb8f630f9f63cf49a6df9241ab89a57b27594c4fac2a9e73`; 69/69 runs,
  786/4215 transitions, 27 byte-sorted observations, zero invariant failures. Active-Play output is
  `35.035` claimed versus `5.005` control (`30.03` bonus); Fiscal credits are `3` and `12`; Pitch
  p50/p95 final rounds are `3/4`, payout scores `3/4`, converted Cash `1/2`; zero-credit Permits
  reach 12/cap at `11,979,232/23,958,463` ms and hoard-credit reaches them at
  `5,989,616/11,979,232` ms. These are observations for owner review, not an invented pass band.
- Honest kernel lockstep bump `0.3.94 -> 0.3.95` covers the new replay artifact owner and watched
  production simulation boundary. No production balance path, epoch seed, registry, immutable
  snapshot, golden, deployment manifest, or current epoch changed.
- **Verification read through final exits:** focused cold packages
  `./activeplay ./harness ./production ./replaycatalog ./relevancepolicy ./cmd/balance-harness`
  GREEN; `make verify-kernel-version` GREEN at 0.3.95; `make verify-schema` GREEN;
  `make verify-harness HARNESS_TEST_COUNT=1` GREEN; cold native `make verify-server-core` GREEN;
  exact Linux/Docker `make verify-harness-ci` exit 0; exact Linux/Docker `make verify-server-ci`
  exit 0; aggregate `make verify` exit 0, including 6,650 client tests, 19,992 browser tests, and
  the Game-UI performance arm. The user's uncommitted `AGENTS.md` edit remains unstaged and
  untouched. Nothing was pushed.

## 2026-08-14 — Claude designated cross-party review of `{2a0f7c9, 6c1d423, d232312}` — APPROVED

- **Review by:** Claude. **Recorded by:** Claude. **Range:** `33418ba..d232312` (unions the span).
- **Executed:** `make epoch7-content-harness` — exit 0 and the committed EH-C10 report
  regenerates BYTE-IDENTICAL; full harness suite GREEN (25.6 s, uncached); `make
  verify-kernel-version` GREEN at 0.3.95; `make verify-server` exit 0. Independent SHA-256:
  manifest `5cb718f6…41bc`, scenario `a6df6553…6a54`, report `cef5fa23…9e73` — all match the
  handoff; the report's embedded scenario/constants hashes reconcile (sha256-prefixed).
- **Contracts honored as ruled:** EH-C11 (cycle-free `relevancepolicy` owner; cumulative T1
  policy in the parsed bundle; the candidate loader now proves LIVE-replay validity of the full
  eighteen-artifact set — the exact gap EH-C15 exposed, now structurally closed); EH-C12
  (eighteen-row manifest, per-row hash/schema/path/gate/verdict pins, recomputed constants hash);
  EH-C13 (strict schema v1, byte-sorted observations, canonical values); EH-C14 (pinned
  coordinate spawns NATURALLY at 3,558 attended ms with a 5,000-ms duration effect on a real
  owned target; 35.035 claimed vs 5.005 control = 30.03 bonus).
- **Plan checkboxes** flipped in the record commit with exercising tests in `2a0f7c9` of the same
  range — compliant under the 2026-08-05 two-commit refinement. The reworded box (registry entry
  deferred to post-mint) matches EH-C13's candidate-only grammar; honest scope, not a silent cut.
- **Note (non-blocking):** `declared_transitions: 4215` is the derived static ceiling beside
  `executed_transitions: 786`; disclosed side-by-side and nothing gates on its tightness. Fine as
  a runaway guard; must never be quoted as measured work (T01-C17 discipline).
- **Reconciliation owed by this reviewer, discharged here:** the EH-C15 ruling record I
  transcribed said the C14 "first spawn moves accordingly" — that inference was carried over from
  the superseded restore-2500 arm and is WRONG for the lifetime-only fix the owner actually
  chose: the spawn stays at 3,558 ms; only expiry moves to 8,058 ms. Codex correctly routed the
  stale sentence back instead of editing it. The ruling's operative content (literal 4500, gap
  unchanged, re-derivation order) is unaffected.
- **No findings. APPROVED.** The scenario hash and C14 coordinate are ready for owner pinning;
  the EH-C10 report is ready for owner review toward the mint sign-off.

## 2026-08-14 — OWNER PINS + MINT SIGN-OFF

Ruled by Marco (recorded by Claude from the owner's explicit selections):

1. **PINNED:** the epoch-7 content-dynamics scenario at
   `sha256:a6df65535eeafd7b96aa017a7bf6666d400058209903d359bb39f5ef7c2f6a54` and the EH-C14
   coordinate (founder `018f6b7c-9abc-7def-8abc-000000000001`, run sequence `1`, first natural
   effect `active.production` on `generator.beige_tower`, spawn `3558` attended ms, duration
   `5000` ms). Any change to either reopens the candidate lane.
2. **MINT SIGNED.** The EH-C10 candidate report (`sha256:cef5fa23…9e73`, 69/69 runs, zero
   invariant failures) is accepted as the owner-facing mint evidence. **Codex is authorized to
   mint epoch 7** from the pinned eighteen-row manifest
   (`sha256:5cb718f6…41bc`): execute the epoch protocol, register the immutable snapshot,
   generate the first historical report and the accepted goldens under the balance-change guard,
   and hand the complete range back for designated cross-party review and the archival chain.
   The mint is local; **publishing/THE PUSH remains a separate owner decision.**

## 2026-08-14 — Codex mint execution: epoch landed; EH-C16 relevance-baseline seam

- **Mint commit:** `38eaffd` (`BALANCE-CHANGE: mint T0-T1 playable content epoch (T01 / FCE
  EH-C10-C15)`). The eighteen production artifacts hash to the pinned
  `sha256:6c7fab29…f789`; epoch 7 is appended without changing any earlier accepted hash; the
  deployment manifest, copy reference graph, Company-v18 activation tests, historical epoch-6
  load, and shared Go/TypeScript replay corpus are updated. Focused cold package tests,
  `make verify-schema`, `make copy-check`, and replay-fixture parity are green.
- **Historical content-dynamics output generated but NOT committed:** the immutable eighteen-file
  snapshot and first report reproduce the signed evidence exactly (69/69 runs, 786/4215
  transitions, scenario hash `sha256:a6df6553…a54`, first spawn 3558 ms). They remain pending the
  baseline commit and its full gates.
- **Testing-flow improvement:** `6035c18` adds explicit `HARNESS_WORKERS`/`-workers` control while
  preserving canonical run-key ordering; its 1-worker-versus-12-worker byte-equivalence test is
  green. This was necessary after the fixed four-worker full pacing updater exceeded the known
  practical runtime envelope. It changes concurrency only, never report bytes.

**EH-C16 — OWNER/DESIGN RULING REQUIRED (baseline gate; mint bytes unaffected).** Activating the
schema-v4 Economy exposes a pre-existing composition gap between two already-ruled evidence
contracts:

1. `LoadRelevanceRegistry` correctly requires the active schema-v4 Economy to have one
   epoch-owned relevance registry entry, but the epoch authority lookup still asks for the stale
   artifact name `relevance_policy`; EH-C11/T01-C14 minted the ruled singular artifact name
   `relevance`. That identifier must be reconciled to the minted name.
2. The active registry gate only accepts the whole-path `relevance.v6` report and rejects ANY
   finding. The ratified T1 whole-path report intentionally retains nine findings for the five
   branch-specific rows covered by C23/C29 (BTv2, CRT Degauss, Institutional Memory, Rack Rail,
   Refurbished Sticker). The reviewed schema-v2 branch report proves every one passes and has an
   empty failure set, but the production registry grammar has no field through which to bind or
   validate that evidence. Thus the ruled acceptance law (“whole-path + branch proofs green
   together”) cannot be represented in the historical registry.

**Codex recommendation:** extend an active relevance registry row with one exact `branch_report`
path. The baseline runner regenerates both reports from the same pinned production scenario;
loading/checking requires byte-exact scenario/constants/policy identity, a valid empty-failure
branch report, and exact set equality between whole-path failing purchasable IDs and passing branch
proof IDs. The baseline guard governs both report paths. This preserves every finding, adds no
exemption, and directly composes the already-ruled C23/C29 evidence. Alternatives rejected by
Codex: silently dropping the nine findings; exempting schema-v4 production from the registry;
pointing the active registry at candidate paths instead of epoch-owned production bytes; or
treating the branch report as unguarded planning evidence.

Until EH-C16 is ruled, the generated snapshot/report stay uncommitted, `harness-update`/`harness-check`
cannot honestly pass at epoch 7, and the mint range is **not ready for designated review or
archival**. No push or publication occurred; the owner's `AGENTS.md` edit remains untouched.

## 2026-08-14 — Codex cold-CI remediation and terminal verification

- **Server repair:** `be2b35b` updates stale post-mint fixtures to consume the live epoch-7
  catalogs and closes a production activation defect exposed by the cold lane: a fresh
  Company-v18 state did not initialize its deterministic Active-Play schedule. Founder identity
  now reaches `FounderInitializer`; new state initializes provision counts/remainders; the first
  natural spawn is pinned through the real Postgres/socket path. Kernel `0.3.95 -> 0.3.96` is an
  honest behavior bump.
- **Client/replay repair:** `7880705` updates stale economy/routes/Active-Play expectations and
  makes the TypeScript replay catalog parse the minted `relevance` artifact with the same
  gate-bounded completeness semantics as Go. Relevance now fail-closes without Opportunities;
  historical replay coverage remains explicit. Kernel `0.3.96 -> 0.3.97` is an honest wire/replay
  bump.
- **Terminal green checks (full output read):** exact cold Linux/Docker `make verify-server-ci`;
  real-Postgres `make test-save-integration`; `make verify-client` (typecheck, production build,
  6,650 tests, boundaries, copy/content guards); `make verify-schema`; `make test-browser-ci`
  (Chromium + Firefox + WebKit: 120 files, 19,992 passed, 3 skipped, plus the dedicated screen
  performance arm); and `make test-game-ui-composed` (real Chromium, embedded gameserver, real
  Postgres, bootstrap + snapshot + WebSocket presence handshake). Kernel guard is GREEN at
  `0.3.97`.
- **Terminal red check, intentionally preserved:** exact cold Linux/Docker `make
  verify-harness-ci` fails only the three EH-C16 baseline assertions: the now-stale
  honestly-empty registry test; the stale `relevance_policy` authority lookup; and the active
  schema-v4 catalog's unrepresentable whole-path-plus-branch evidence. The generated immutable
  snapshot/report remain uncommitted pending the owner ruling. No test was skipped, weakened, or
  reclassified to manufacture a green aggregate.
- **Protocol state:** the mint production bytes are landed locally and the ordinary server,
  client, schema, browser, and composed-stack CI lanes are repaired. The overall range remains
  **NOT ready for designated review or archival** until EH-C16 is ruled and the historical
  content-dynamics baseline can be committed and pass its exact Linux gate. Nothing was pushed;
  the owner's uncommitted `AGENTS.md` edit remains untouched and unstaged.

## 2026-08-14 — OWNER RULING: EH-C16 ACCEPTED as proposed

Ruled by Marco (recorded by Claude; the owner asked for and received the zoom-out — this is the
LAST instrument/bookkeeping item of the epoch-7 chain, and the next critical-path work is game
content: the first-session script, screens, minigames):

1. The active relevance registry row gains one exact `branch_report` binding, validated by
   identity (scenario/constants/policy hashes must match the bound whole-path report) and strict
   set equality: every whole-path finding not otherwise carried must correspond to a passing
   branch row; no exemptions, no dropped findings, any mismatch fails the gate.
2. The stale `relevance_policy` authority lookup is reconciled to the minted artifact name
   `relevance` (EH-C11/T01-C14).
3. The generated immutable snapshot/goldens may then be committed under the balance-change guard,
   `make verify-harness-ci` must go green, and the complete mint range comes back for the
   designated cross-party review and archival chain.

## 2026-08-15 — Codex EH-C16 + AC0 closeout — READY FOR DESIGNATED REVIEW

- **Declared implementation range:** `fd5c641^..133d13f`. This is an implementation/self-check
  handoff only; it is not a designated verdict, archival authorization, or publication action.
- **EH-C16:** the active epoch-7 relevance row binds the minted `relevance` artifact and one exact
  schema-v2 branch report. The gate requires byte-identical scenario/constants/policy identity,
  zero branch failures, and exact set equality between the five whole-path misses and the five
  passing branch proofs. The governed proof regenerated from production bytes at **1,858,591
  transitions**; its proof set is `generator.beige_tower_v2`, `upgrade.crt_degauss_button`,
  `upgrade.institutional_memory`, `upgrade.rack_rail_standardization`, and
  `upgrade.refurbished_sticker`.
- **Historical evidence:** isolated commit `a0e6788` (`BALANCE-CHANGE:`) adds the immutable
  eighteen-artifact snapshot at constants hash `sha256:6c7fab29…f789`, the first epoch-7
  content-dynamics golden (`sha256:18999011…1fc1`), the active whole-path report
  (`sha256:ab9c99ce…2394`), the branch report (`sha256:133312ba…e7b`), and the refreshed epoch-7
  pacing/golden baselines. No production content byte changed.
- **Testing-flow repairs found while executing the real gates:** update now validates sources
  before simulation and alone may materialize missing generated evidence; normal checks remain
  fail-closed. The history guard accepts the ruled input-then-isolated-baseline protocol while
  requiring every immutable byte to be added atomically and never changed later. A fast
  `make harness-guard-check` exposes packaging failures before long simulation. Routine
  `harness-check` recomputes the whole-path verdict but loads the committed identity-bound branch
  evidence instead of repeating its expensive derivation. The Docker lane now propagates
  `HARNESS_WORKERS`; canonical worker-count byte parity remains pinned.
- **AC0:** `aba9399` applies the frozen session-boundary offline interval through `ApplyLogged`
  before the online command (and before Exit), preserves reject-equals-no-mutation, and advances
  the replay-input grammar to v7 while retaining v2-v6 histories. A real-Postgres fixture idles an
  epoch-7 founder with provisioning state for 48 hours and proves the canonical 90%/24h catchup,
  persisted replay coordinates, byte-identical replay, the following online mutation, and stale
  Exit conflict behavior.
- **Terminal checks read through final exit:** `make harness-update HARNESS_WORKERS=12` GREEN;
  `make harness-guard-check` GREEN; native `make harness-check HARNESS_WORKERS=12` GREEN; exact
  cold Linux/Docker `make verify-harness-ci HARNESS_WORKERS=12` GREEN; exact cold Linux/Docker
  `make verify-server-ci` GREEN; full real-Postgres `make test-save-integration` GREEN; `make
  verify-client` GREEN (typecheck/build, **6,651 passed**, kernel **0.3.98**); `make verify-schema`
  GREEN; hosted-equivalent `make test-browser-ci` GREEN (**19,995 passed / 3 skipped** plus the
  performance arm). `git diff --check` is clean. The user's `AGENTS.md` remains the sole working
  tree change and was never staged or modified by Codex. Nothing was pushed or archived.
