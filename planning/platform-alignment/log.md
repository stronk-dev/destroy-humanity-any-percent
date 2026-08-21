# Platform alignment — append-only log

## 2026-08-20 — scope correction after owner challenge

- Owner correctly challenged the elapsed time and depth of the first checkpoint. `cb162a3` is now
  explicitly classified as control-plane setup plus a principal-risk scan, not the completed
  sibling-repository-style audit.
- Replaced the lightweight plan with an eight-wave program: exhaustive inventory, per-outcome
  traces, acceptance/test discrimination, runtime/release research, owner rulings, dependency
  graph, executable queue, and contradiction/designated-review closure.
- Product evidence remains pinned to `190a4fa`. No verdict is promoted merely because the planning
  structure exists.
- Began Wave 1 with counted populations: 14 top-level design documents, 24 active-directory RFC
  files, 46 archived RFCs, 23 live planning directories, 38 docs, 340 Go files (151 tests), 74
  migrations, 82 client source files, 56 client test artifacts, 91 balance files, ten copy files,
  72 Make targets, seven CI jobs, and five registered Game UI surfaces.
- Filed RP-016 after a code/data/RFC trace disproved Minigame Platform's “normative body
  reconciled” claim. Production registers only The Pitch; the duel RFC is draft and `balance/combat`
  is absent, while the platform header, MP1, MP5, rulings, and AC6 still require or assume duel as
  tenant #1. This is routed to the ruling author rather than repaired by implementation inference.
- Full design-body reading corrected the initial audit's D-004 framing: silent server-anonymous is
  already the adopted default and local-only play is the outage fallback. Filed RP-017–RP-020 for
  the Soul-recovery, cold-open-WR, minigame-socket, and production-hot-reload contradictions. None
  was rewritten because owner-authored/ruling-body reconciliation belongs to the author.
- Completed the first full read of all 14 tracked top-level design documents and recorded 121
  preliminary stable outcome IDs in `design-capability-ledger.tsv`. These are section-level
  inventory rows, not yet Wave-2 proof: aggregate rows remain explicitly marked for splitting into
  independently falsifiable user workflows.
- Completed R-001's first diagnostic wave. A full current local
  `make harness-check HARNESS_WORKERS=12` reached exit 0 below the 30-minute ceiling, while hosted
  run `32404232364` spent 29m04s silently inside the same `mode=check` before job cancellation.
  Five consecutive current-product hosted runs are cancelled. The last hosted-green harness was
  6m16s at `146deded`; the later `a6328df` capacity commit has no Actions run of its own.
- Code tracing shows the 12-worker dispatcher covers the 300 standard pacing simulations only;
  registry rows and their relevance experiments are serial. The production catalog expanded from
  two generator classes to the T0-T1 schema-v4 catalog without changing the visible 300-run count.
  Filed RP-021–RP-023 for unproven hosted headroom, partial worker coverage, and silent termination.
- Corrected two tempting but invalid focused-run inferences: `make t1-relevance` uses fixture data,
  and a scenario-only active-path run does not receive the registry loader's accepted epoch hash.
  Neither is the active authoritative row. The full registry-aware local check matched its golden;
  R-001 now requires an authority-preserving selector before per-row cost claims.
- Extracted every true active-RFC acceptance criterion into `active-acceptance-ledger.tsv`: 111
  unique `(RFC, AC)` rows across 21 product/process RFCs. A first parser reported 115 because it
  swallowed Combat Duel's nested D4 hardening list; the heading-boundary control caught and
  removed those four false ACs before the inventory was accepted.
- Completed the first 111-row acceptance classification without promoting test names to proof:
  39 draft rows, 56 mechanically backed rows awaiting execution/range review, nine unmet/partial
  rows, three current contradictions, and four withdrawn rows. `acceptance-audit.md` records the
  exact CI, Minigame Platform, Game UI, API, Minigame Surface, and Combat downgrades plus the next
  cold-execution batches.
- Ran Account/Leaderboard/Transport/Gameserver Batch A cold. Account and Leaderboard real-Postgres
  packages, Account/Transport/Gameserver unit packages, and Integration-named Gameserver tests all
  reached exit 0. Kept Account AC2 partial because refresh-family revocation and connected-socket
  reauthentication have no composed witness. Kept WebSocket AC4/AC5 partial because overflow→resync
  and drain→reconnect are only separately tested. Filed RP-024.
- The first focused Account selector never launched: `SAVE_TEST_FLAGS` is expanded unquoted and a
  parenthesized Go regex is parsed by `/bin/sh`. The safe package-wide rerun passed; the invalid
  attempt was not counted. Filed RP-025 rather than hiding the selector failure.
- Ran the complete browser suite cold (123 files; 20,007 passed; three visible skips), its isolated
  Game UI performance arm, the real Postgres/gameserver/Chromium composed bootstrap, and client
  type/boundary gates. The composed script explicitly ends at Desk/snapshot/presence, confirming it
  cannot prove AC1. Downgraded AC4 and filed RP-026 because drain/resync are co-injected with only
  an axe/mechanical-ID assertion, not the required story-beat or recovery behavior.

## 2026-08-20 — Codex truth audit and program foundation

- Audited local and remote `HEAD` at `190a4fa`; `origin/main` is identical. Preserved the user's
  sole pre-existing worktree edit in `AGENTS.md` and made no product, schema, balance, RFC, or
  player-copy change.
- Read RFC-0000, the active index, vision/architecture/voice rules, current state, backlog,
  research matrix, coverage map, every active RFC status, every live plan, representative logs,
  canonical docs, client/server boundaries, Make gates, and deployment artifacts.
- Remote verification found the repository public and every current-head Actions run cancelled.
  Push run `32009994004` proves server/client/schema/browser/composed-browser green but cancels the
  harness after 30 minutes inside `balance-harness -mode=check`; the following three scheduled runs
  repeat the cancelled conclusion.
- Reclassified outcomes at user-workflow granularity. The numeric/save/production and bounded
  T0–T1 server paths are genuinely proven. Accounts, UI, minigames, pets, MMO, leaderboards, and
  accessibility contain strong mechanical fragments without their complete user consumers. The
  nine-tier product, release packaging, backup/restore, export, and self-host/sunset deliverables
  are absent.
- Filed the release audit and alignment artifacts. The execution hold permits only R-001 diagnosis
  and lifecycle record reconciliation until owner choices and authored body repairs land.
- Verification note: a cold local `make verify GO_TEST_FLAGS='-count=1'` was launched. All Go
  packages and pre-harness generators passed before it entered the same long-running
  `balance-harness -mode=check` step. After continued silence it was manually interrupted; the
  aggregate exited 130 at `harness-check`. This is **not** a green or completed measurement and is
  not used to set a larger budget. It independently reproduces the hosted stall location only.
- Validation exposed RP-015: the mandated internal backlog, research matrix/dossiers, and coverage
  map are gitignored from the public repository. The internal ledgers were updated locally, while
  `backlog.md` and `release-platform-audit.md` carry the same program evidence in tracked form.
  Nothing was force-added or published; D-002 owns the durable public/private model.

## 2026-08-20 — Permits and First Content lifecycle reconciliation

- Read both active RFCs and their complete plan/log authority chains, inspected the exact candidate,
  mint, repair, baseline, archival, and cold-cache commits, and reconciled their cross-party verdict
  ranges. The Permits candidate set `{7d9cb37, 90633a6, d30ab9e}` and the FCE mint set
  `{3ff34bf, 84cf570, c41b388, 08c995e}` are genuinely designated-review-covered; later F1 and
  cold-cache repairs also have exact designated verdicts.
- Promoted only the criteria that survived fixture inspection and range review. Permits AC1/AC2
  are proven; FCE AC1/AC3 are proven; FCE AC2 is historically proven but its HEAD witness drifted;
  FCE AC5 is proven only under the accepted range-head ruling. Kept Permits AC3 and FCE AC4 partial.
- Filed RP-027–RP-033. Permits never landed the required exact two-resource Go/TypeScript replay
  row; the existing doctrine replay fixture constructs a cash-only T3→T4 gate. The Permits body
  still contradicts PT-C1/PT-C2/PT-C4. FCE's changelog does not resolve every consumed artifact to
  an exact reviewed range, and FCE5.5/AC5 still contradict the range-head ruling.
- Confirmed current canonical-doc drift: Economy/Purchasable docs call the active catalog pre-mint
  schema v3, Routes says production has no T3→T4 gate, and Pitch still says no production epoch pins
  its content. No canonical docs were silently rewritten because closeout also needs missing proof,
  authority reconciliation, and a new designated range.
- Current cold evidence passed: `./production ./replaycatalog ./harness -count=1` (4.758 s,
  5.759 s, 35.667 s); client (39 files passed, 2 skipped; 6,655 tests passed, 15 skipped);
  real-Postgres Integration tests for `./production ./gameserver`; `./economy ./routes ./doctrine
  ./epochseed -count=1`; schema, copy, and deployment-content-manifest checks.
- Wrote `permits-first-content-lifecycle-audit.md` with criterion-level verdicts and the smallest
  honest closeout order. Did not flip either implementation plan, edit owner-authored RFC text,
  archive, change product bytes, or touch the user's `AGENTS.md` edit.

## 2026-08-20 — Prestige & Exits lifecycle reconciliation

- Read the complete active Prestige RFC/plan/log, its post-rewrite implementation history, the
  current Exit/store/arithmetic/UI witnesses, the successor T0–T1 curriculum and first-hour
  authority, canonical docs, and all review references. The ignored local round-4 review file is
  not repository authority and lacks the explicit provenance required by today's archival law.
- Ran cold `./prestige ./production ./save`, all client tests, and real-Postgres Integration tests
  for `./production ./save ./gameserver`; all passed. Regenerated the governed first-hour
  experiment: 97/97 completed, no warnings/failures, Chaos elective p50 and Casual p95 both
  2,700,000 ms.
- Promoted AC8 only, with the explicit qualification that its lower bound is policy-constructed and
  proves availability under the governed persona, not unaided human choice timing. Kept AC1/AC7 at
  cold-witness green pending review and downgraded AC2–AC6 to partial against their literal proof
  requirements.
- Filed RP-034–RP-038 for stale normative/canonical contracts, the missing Advisor control, the
  deferred Quarter bridge, missing literal witnesses, and the absent tracked full-range designated
  verdict. Wrote `prestige-lifecycle-audit.md` with exact evidence and the safe closeout order.
- Made no product, balance, schema, RFC, canonical product-doc, or player-copy edits; flipped no
  plan boxes; did not archive or touch the user's `AGENTS.md` edit.

## 2026-08-20 — Leaderboards & Balance Epochs lifecycle reconciliation

- Read the complete active Leaderboards RFC/plan/log, design outcomes, current epoch/board/category
  implementation, canonical docs, the archived Run Genesis RFC/log, its live remediation thread,
  replay corpus, public API registry, gameserver composition, client tree, and all relevant review
  entries and commit history.
- Ran cold `./leaderboard ./production ./replaycatalog ./epochseed`, all client tests, real-Postgres
  Integration tests for `./leaderboard ./production ./gameserver`, focused epoch-guard tests, and
  the live repository history guard; all passed.
- Promoted AC2 and AC4 only to cold-witness green pending archival review. Kept AC1 and AC3 partial,
  classified AC6 as missing, and marked AC5 failed: Exit serializes current Compact membership,
  leaving clears it, and the projector trusts the terminal boolean without reading membership
  history. The existing fixture hard-codes true and cannot falsify this path.
- Filed RP-039–RP-043 for the Assisted defect, missing runtime/client consumers, missing literal
  witnesses, body/scope contradictions, and lifecycle/range-provenance defects. Confirmed the
  player capability has no production board/epoch HTTP binding, client reader/page, historical
  browse, archive-validator workflow, or combined Route Registry surface.
- Wrote `leaderboards-lifecycle-audit.md` with producer→consumer classification and safe closeout
  order. Made no product, balance, schema, RFC, canonical product-doc, or player-copy edits; flipped
  no plan boxes; did not archive or touch the user's `AGENTS.md` edit.

## 2026-08-20 — Minigame Platform Foundation lifecycle reconciliation

- Read the complete active platform RFC/plan/log, current minigame/production/gameserver and
  TypeScript implementation, canonical docs, production Pitch artifact/composition, archived Pitch
  authority, API successor tests/log, combat state, and relevant review histories.
- Ran cold `./minigame ./pitch ./replaycatalog ./production`, all client tests, TypeScript/Svelte
  checks, and real-Postgres Integration tests for `./minigame ./production ./gameserver`; all
  passed. These prove substantial solo/session/payout/API mechanics, not absent async/bot/decay
  workflows.
- Marked AC2 failed. The only decay call occurs during result resolution, whose fresh grade then
  overwrites the decayed value and zeroes its remainder; production start reads stored grades
  unchanged. No other transition advances the grade and no automated-output consumer uses the
  declared destination, so lapse-time decay is behaviorally dead.
- Kept AC1/AC3 partial: production hardcodes solo, while async-snapshot is only a fixture/enum; bot
  fallback has exact schema and reduction arithmetic but no bot tenant or match. Promoted AC5 from
  mechanical to proven because production Pitch completes through authenticated composition with
  a solo fallback and no peer. Kept AC4 cold-green pending final cross-RFC range reconciliation and
  AC6 contradicted because the duel tenant does not exist.
- Filed RP-044–RP-047 for offline-quality decay/application, async/bot acceptance gaps, RFC/plan/
  docs/successor contradictions, and missing final cross-RFC range union. Wrote
  `minigame-platform-lifecycle-audit.md` with the exact defect trace and safe closeout order.
- Made no product, balance, schema, RFC, canonical product-doc, or player-copy edits; flipped no
  plan boxes; did not archive or touch the user's `AGENTS.md` edit.

## 2026-08-20 — Account & Session Bootstrap lifecycle reconciliation

- Read the complete Account RFC/plan/log, account migrations/repository/API, current Game UI
  runtime, transport authentication, gameserver/bootstrap/GC successors, canonical docs,
  Leaderboards imported-founder consumers, tests, and review history.
- Ran cold `./account ./transport ./gameserver`, real-Postgres Integration tests for `./account
  ./leaderboard ./gameserver`, and the composed Chromium bootstrap/snapshot/Centrifuge handshake;
  all passed. The current product coordinate's 6,655 client tests were already green in this wave.
- Promoted AC1/AC4 to proven integration. Kept AC2 partial and downgraded AC3/AC5/AC6/AC7 to
  partial because their fixtures do not prove repeated unlimited/free Founder swaps, actual
  import→board exclusion, survival of the active imported stream, or limiter behavior on session
  and refresh routes.
- Recomputed the acceptance-summary distribution directly from all 111 TSV rows after noticing
  the prior family totals had drifted as lifecycle passes introduced qualified status labels. The
  normalized families are 39 draft, 23 mechanical/cold-pending, 11 proven/qualified, 29 partial or
  unmet, five contradicted/failed, and four withdrawn; they sum to the unchanged 111-row census.
- Confirmed RP-048: the Game UI persists refresh/recovery material but every HTTP/socket operation
  keeps using the original 15-minute access token. There is no refresh call, credential replacement,
  socket reauthentication, or expired-session recovery; failures become generic offline state.
- Filed RP-049–RP-051 for literal Account witnesses, obsolete review/lifecycle provenance, and the
  unowned anonymous-account storage-amplification/retention problem. Existing RP-004–RP-007 and
  RP-024 continue to own user data rights/recovery/fallback and revocation-to-socket integration.
- Wrote `account-session-lifecycle-audit.md` with the criterion trace and safe closeout order. Made
  no product, schema, RFC/design, canonical product-doc, or player-copy edits; flipped no plan
  boxes; did not archive or touch the user's `AGENTS.md` edit.

## 2026-08-20 — WebSocket Transport & Fan-out lifecycle reconciliation

- Read the complete Transport RFC/plan/log, embedded node, policy/wire/history/queue/outbox code,
  actual-socket/soak/drain/authz tests, production gameserver composition, current browser runtime,
  canonical docs, successor review records, rewrite map, and relevant implementation history.
- Ran cold `./transport ./gameserver` and all client tests; 5,000-socket fan-out remains green in
  the normal Transport package, and the client result is 6,655 passed/15 skipped. The preceding
  same-coordinate Account wave supplies the real-Postgres gameserver and composed browser
  handshake. A root `pnpm --filter` attempt never started because no root manifest exists and is
  explicitly excluded from evidence.
- Promoted AC1 only. The soak uses 5,000 real WebSockets over ten ticks, terminal monotonic
  subsequences, and wrong-kind/public-receipt negatives. Kept AC2–AC6 partial against integrated
  or literal requirements.
- Confirmed RP-052: the production Game UI stores no Centrifuge epoch/offset, never requests
  recovery, reconnects, resubscribes, interprets typed close codes, or honors `resume_after_ms`.
  Server recovery, overflow, and drain therefore stop at test protocol drivers.
- Confirmed RP-053: `PlayerRevisionCursor` is imported only by its unit test. Production event
  consumption has no per-scope gap detection, duplicate suppression, snapshot reset, or historical
  cursor behavior despite the RFC and canonical doc claiming that authority.
- Filed RP-054 for AC3's sub-second/two-frame witness and AC6's absent non-member Guild negative;
  filed RP-055 for stale T6/plan/log truth and the missing current full-span review union. The
  designated F2/F3 verdict exists in a successor log and maps through checked-in rewritten hashes,
  but later protocol-aware soak/composition/browser ranges remain outside it.
- Recomputed the 111-row acceptance distribution: 39 draft, 19 mechanical/cold-pending, 12
  proven/qualified, 32 partial or unmet, five contradicted/failed, and four withdrawn.
- Wrote `websocket-transport-lifecycle-audit.md` and exposed the accepted-scope recovery completion
  separately from Account token rotation. Made no product, schema, RFC/design, canonical
  product-doc, or player-copy edits; flipped no plan boxes; did not archive or touch the user's
  `AGENTS.md` edit.

## 2026-08-20 — CI Baseline lifecycle reconciliation

- Read the complete CI RFC/plan/log, current workflow and Make graph, all 20 schemas and their
  verifier bindings, canonical CI docs, R-001 diagnosis/queue, public Actions metadata, successor
  harness/Game UI records, primary setup-go cache documentation, and relevant review history.
- Reconfirmed public run `32404232364` at remote `cb162a3`: server, client, browser, schema, and
  composed-browser succeed in roughly 0.5–2 minutes; the harness step runs 30m02s and is cancelled,
  cancelling the workflow. This is the same product/workflow coordinate despite local planning-only
  audit commits.
- Executed and restored two predeclared failing probes. Changing shared Decimal `0^0` expectation
  from `1` to `2` made `make test-client` fail exactly one row (6,654 passed); adding a malformed
  production catalog made `make verify-schema` exit 2. After restoration, client returned to 6,655
  passed/15 skipped and the complete schema population passed.
- Promoted CI AC2 and AC4. Kept AC1/AC3 contradicted and marked AC5 contradicted: harness/numeric
  setup-go defaults cache Go build outputs despite dependency-only claims; schedule/manual runs all
  blocking jobs; and numeric-maintenance has no non-blocking failure behavior. Scheduled run
  `31562066740` proves a numeric failure fails the workflow.
- Filed RP-056 for the four-job/sub-five-minute/harness-excluded body and stale “first hosted run”
  plan; RP-057 for cache/nightly workflow contradiction; RP-058 for the fragmented full-range
  review union; and RP-059 because R-001 instrumentation has no active accepted RFC owner.
- Recomputed the 111-row distribution: 39 draft, 16 mechanical/cold-pending, 14 proven/qualified,
  32 partial or unmet, six contradicted/failed, and four withdrawn.
- Wrote `ci-baseline-lifecycle-audit.md`. Made no persistent workflow, Make, schema, product,
  RFC/design, canonical product-doc, or player-copy edit; flipped no plan boxes; did not archive,
  push, or touch the user's `AGENTS.md` edit.

## 2026-08-20 — API Foundation lifecycle reconciliation

- Read the complete API RFC/plan/log, schema/registry/cursor/policy/generator code, all generated
  artifacts, account/gameserver mounting, production Game UI callers, canonical API doc, tests,
  implementation history, and designated review records.
- Ran cold `./publicapi ./account ./gameserver`, `make api-check`, root `make typecheck`, and the
  client boundary gate green. A root `pnpm type-check` invocation did not launch due to the absent
  root manifest and is excluded from evidence.
- Executed and restored four discrimination probes: registered optional-field drift failed with
  exact generated diffs; nested-field removal rejected without naming the removed field; a direct
  production-runtime API fetch passed the boundary lint; and a live hand-mounted Founder response
  change passed `api-check`.
- Classified AC1/AC2 partial, AC3/AC5 unmet, and AC4 contradicted under C9. Recomputed the 111-row
  distribution to 39 draft, 14 mechanical/cold-pending, 14 proven/qualified, 33 partial/unmet,
  seven contradicted/failed, and four withdrawn.
- Filed RP-060–RP-065 for incomplete route authority, absent public composition/formula artifact,
  AC2 diagnostic/pin gaps, C9 body/client/lint contradiction, missing query/header/raw-success
  generator contracts, and stale lifecycle/review provenance.
- Wrote `api-foundation-lifecycle-audit.md`. Made no persistent product, schema, RFC/design,
  canonical product-doc, generated-artifact, or player-copy edit; flipped no plan boxes; did not
  archive, push, or touch the user's `AGENTS.md` edit.

## 2026-08-20 — Game UI Screens lifecycle reconciliation

- Read the complete Game UI RFC/plan/log, GU-C25–C28 owner rulings, Phase-A components and runtime,
  snapshot contracts, composed/browser/performance fixtures, canonical Game UI doc, implementation
  history, successor review records, and current acceptance ledgers.
- Ran the complete browser population, typecheck, and client-boundary gate. The restored baseline is
  123 browser files with 20,007 passed/3 skipped, the isolated performance selector 1 passed/10
  skipped, zero type diagnostics, and a green boundary gate.
- Executed and restored three failing probes. A direct transport import failed AC2's exact component
  boundary; widening Run End props to admit a snapshot failed AC3's compile-time negative; removing
  cap, drain, and resync output together left all 20,007 browser assertions green, disproving AC4's
  claimed behavior coverage.
- Promoted AC2 and AC3, contradicted AC4's acceptance oracle, and kept AC5 partial. The performance
  fixture consumes update-count, format-count, and long-task fields but never applies its 4x CPU
  throttle or dropped-frame budget, completes the nominal 60-second input population in under a
  second, and has no recorded manual reference-device result.
- Confirmed that current production stops at snapshot v2/bootstrap-to-Desk. The GU-C26/C27 snapshot
  v3 eligibility, Gate/Wind Down controls, and Run End-to-next-run continuation are absent. AC1,
  U2, the header, and canonical full-rendition prose also remain unreconciled with C7/C25–C28.
- Recomputed the 111-row distribution: 39 draft, 11 mechanical/cold-pending, 16 proven/qualified,
  33 partial/unmet, eight contradicted/failed, and four withdrawn. Filed RP-066/RP-067 and refined
  RP-008/RP-026 for the performance, body/canonical, implementation, oracle, and review-union gaps.
- Wrote `game-ui-lifecycle-audit.md`. Made no persistent product, RFC/design, canonical product-doc,
  or player-copy edit; flipped no plan boxes; did not archive, push, or touch the user's
  `AGENTS.md` edit.

## 2026-08-20 — Minigame API & Surface lifecycle reconciliation

- Read the complete MA RFC and fifteen rulings, plan/log, generated registry/OpenAPI/DTO authority,
  handlers, composed Minigame/Pitch and Soul Recovery paths, client tree, canonical docs,
  implementation history, successor verdicts, and current acceptance ledgers.
- Ran cold Account/Minigame/Minigame-API/gameserver units, Account and gameserver Postgres suites,
  API generation drift, and TypeScript/Svelte checks green. One parallel Compose attempt collided
  on the shared test database and is explicitly invalid/excluded; sequential reruns passed.
- Executed and restored four probes. Severing tenant content failed the composed lifecycle exactly;
  bypassing Recovery rate limiting failed its 429 check; accepting Recovery finish `founder_id`
  left Account unit tests green; and appending bytes to a supposedly byte-exact rejection also
  left them green.
- Promoted AC1 as composed backend proof. Kept AC2 partial because error-pair authority and exact
  bytes are not closed, kept AC3 partial because Recovery rejection has no authoritative-state
  comparison, contradicted AC4's every-endpoint enumeration, and kept AC5 unmet/body-blocked.
- Confirmed the absent consumer chain: no generated HTTP client, exact C9 component contract,
  tenant-surface registry, Pitch table mount, Recovery surface/toy/scheduler constructor, or
  browser/a11y human workflow. Canonical backend docs accurately preserve this limitation.
- Filed RP-068–RP-071 for error/byte authority, Recovery nonmutation, privacy enumeration, stale
  ruling body/status/plan, absent surface/client, and unassembled current review provenance.
- Recomputed the 111-row distribution: 39 draft, seven mechanical/cold-pending, 17
  proven/qualified, 35 partial/unmet, nine contradicted/failed, and four withdrawn.
- Wrote `minigame-api-surface-lifecycle-audit.md`. Made no persistent product, RFC/design,
  canonical product-doc, generated artifact, player-copy, or checkbox edit; did not archive, push,
  or touch the user's `AGENTS.md` edit.

## 2026-08-20 — Combat Shared Data & Arithmetic lifecycle reconciliation

- Read the complete Combat parent RFC/plan/log, Go/TypeScript arithmetic and shared RNG,
  cross-runtime vector corpus, recursive AST boundary, canonical doc, child status, implementation
  history, historical review entries, and current acceptance ledgers.
- Ran cold Combat/Determinism, the full 6,655-test client population, the combat boundary, and
  TypeScript/Svelte checks green after restoration.
- Executed and restored seven one-seam probes: ATK-stage removal, label collapse, and one chart-edge
  removal independently in both runtimes, plus a real nested `.mts` native-division file. Every
  claimed implemented behavior failed at the exact oracle; all probe bytes were removed.
- Promoted AC2–AC4. Kept AC1/AC5 specification-blocked because the RFC supplies neither exact
  effect/spell unions nor literal Trust/Obedience/Soul tables. Marked AC6 contradicted under its
  all-path wording: the client gate is proven, while server Combat's native division is unscanned.
- Confirmed canonical docs honestly describe only the kernel and gaps. The active plan/ledger are
  stale, all historical review hashes are non-resolving and lack current provenance/ranges, and no
  designated pass follows the final round-three remediation at current hashes.
- Filed RP-072–RP-074 for missing owner contracts, stale lifecycle/review truth, and the C3/AC6
  runtime-scope contradiction. Recomputed the 111-row distribution: 39 draft, five
  mechanical/cold-pending, 20 proven/qualified, 33 partial/unmet, ten contradicted/failed, and four
  withdrawn.
- Wrote `combat-shared-lifecycle-audit.md`. Made no persistent product, RFC/design, canonical
  product-doc, balance/content, or checkbox edit; did not archive, push, or touch the user's
  `AGENTS.md` edit.

## 2026-08-20 — Deployment/package/configuration release audit

- Read the complete draft Deployment RFC, active index, gameserver entry point/composition,
  transport and API policies, Make/Actions topology, all Compose/packaging files, license audit,
  canonical operations docs, Git remote state, and current alignment records.
- Built the real gameserver successfully as a 29 MB arm64 executable and invoked it without
  credentials; it failed closed with `invalid gameserver composition`. A tracked-file credential
  pattern scan found only disposable test database values and no private-key/GitHub-token pattern.
- Confirmed that every Compose file is test-only and no Dockerfile/Caddyfile/production Compose,
  client static server, backup/restore/rollback tool, release workflow, or third-party-license
  deliverable exists.
- Confirmed the binary is not the claimed embedded artifact: it reads moderation and transport
  policy from `CLOUD_CLICKER_REPOSITORY_ROOT`, embeds/serves no client, pins WebSocket origin to
  localhost, hardcodes Account trusted-proxy depth to zero, and cannot configure previous JWT,
  bootstrap, or cursor keys.
- Classified Deployment AC1 unmet, AC2 primitive-only, AC3 contradicted/incomplete, AC4 partial,
  and AC5 proven as an external-state fact. The draft's local-only/THE-PUSH center is stale even
  though the remaining release capability is absent.
- Filed RP-075–RP-078, deepened RP-002/RP-003/RP-011, updated the release/capability/reality and
  execution/research records, and wrote `deployment-foundation-lifecycle-audit.md`. Made no product,
  RFC/design, canonical product-doc, deployment, balance/content, or secret edit; did not push or
  touch the user's `AGENTS.md` edit.

## 2026-08-20 — Account recovery and player-rights release audit

- Traced production bootstrap credentials through localStorage, every runtime HTTP/WebSocket use,
  Settings rendering, Account recovery/refresh/logout/email/Founder/import/delete endpoints, the
  export population, and the existing same-coordinate Postgres/browser evidence.
- Confirmed a strong atomic backend but no player rights/recovery consumer: the recovery and refresh
  secrets are silently stored, refresh is never called, Settings has no account action, email is
  hard-501, deletion is backend-only, and export is absent.
- Confirmed missing/malformed storage either starts replacement bootstrap or strands the Desk in a
  generic offline state. The unit-tested destructive Settings deferral has no production caller.
- Confirmed Settings' “progress is parked on this machine” copy is false at HEAD: no local gameplay
  save, durable intent queue, fallback owner, reconnect flush, or import consumer exists.
- Filed RP-079–RP-081, deepened RP-004–RP-007/RP-048, expanded R-003's failure population, and wrote
  `account-rights-release-audit.md`. Re-used already-recorded cold evidence rather than relabeling
  backend/component tests as a player workflow. Made no product/API/copy/RFC/design/canonical-doc/
  schema/player-data edit; did not push or touch the user's `AGENTS.md` edit.

## 2026-08-20 — Accessibility release audit

- Read UI Foundation/Game UI accessibility contracts and docs, theme and shell motion wiring, every
  current Svelte surface, browser configuration/tests, responsive CSS, and internal compliance/
  rendering/mobile research.
- Preserved the strong bounded baseline: five-surface axe WCAG 2.2 AA across three engines, native
  controls/text states, visible focus, Enter activation, Amount naming, and theme-duration reduction.
- Added two temporary desired-behavior probes, ran them through Chromium/Firefox/WebKit, and removed
  them plus generated screenshots. Lifecycle Offer preemption failed by dropping focus to `<body>`;
  complete Desk reflow failed at `647 > 320` CSS pixels in every engine.
- Confirmed production reduced motion samples the media query for CSS only: numeric shell counters
  retain `reducedMotion=false`, and no listener handles a preference change after mount.
- Restored `make test-browser BROWSER_TEST_FLAGS='test/game-ui-screens-browser.test.ts'` passed 30
  functional rows/3 skips across engines and 1 performance row/10 skips, proving probe cleanup and
  the declared suite's blindness to both failures.
- Filed RP-082–RP-084, expanded R-005 and the evidence/release/capability queues, and wrote
  `accessibility-release-audit.md`. Made no persistent product/test/RFC/design/copy/canonical-doc/
  theme change; did not push or touch the user's `AGENTS.md` edit.

## 2026-08-20 — Operations, retention, and preservation release audit

- Traced production health/readiness, background-job supervision, shutdown, logging/metrics,
  request identity, scheduled cleanup, all major retained-data families, and designed-sunset claims
  from primitives through actual gameserver composition.
- Ran cold `./gameserver ./cmd/gameserver ./account ./save -count=1`; all four packages passed.
  Preserved the proven supervision/drain and credential/bootstrap-GC strengths without promoting
  them to an operator workflow.
- Confirmed no production metrics exporter/route, request/access correlation, alert/SLO/dashboard/
  runbook, and that Production's optional invariant counter is composed as nil.
- Confirmed `Store.PruneIntentRecords` has no production caller despite canonical 30-day scheduler
  prose. Anonymous accounts, archives, append-only histories, dead letters, projections, and logs
  have no complete accepted retention schedule/disclosure.
- Confirmed the designed-sunset report's “deliverable already exists / not an engineering project”
  premise is false at the product coordinate: only test Compose exists and client/export/bots-
  default/final-artifact/mirror/runbook capabilities are absent.
- Filed RP-085–RP-088, added R-007/D-011 and blocked execution routes, completed Wave 4's bounded
  coordinate pass, and wrote `operations-retention-preservation-audit.md`. Made no product/test/
  schema/RFC/design/canonical-product-doc/player-copy/retention edit; did not push or touch the
  user's `AGENTS.md` edit.

## 2026-08-20 — Research/decision dependency and owner-ruling preparation

- Re-read the complete design capability ledger, all tracked RP routes, research/decision queues,
  roadmap/vision, research coverage matrix, Deployment draft, and active RFC index.
- Found and repaired RP-089's circular gate: D-005 required R-003, while R-003 required an absent
  workflow that only the recovery/export/deletion decisions could authorize. The sequence is now
  owner posture -> accepted prototype -> participant validation.
- Marked R-002's repository baseline complete while preserving D-001 as its owner-only exit. Added
  exact preconditions to R-003/R-005 and predeclared R-008 first-session comprehension and R-009
  aggregate telemetry/milestone calibration with failing controls.
- Expanded the owner queue through D-017 for Advisor Mode, async minigame scope, measured CI
  topology, complete data retention, gameplay telemetry, and UGC/social scope. These were choices
  previously hidden inside vague remediation routes.
- Wrote `owner-ruling-packet.md` with exact next-milestone options, a recommended bounded Phase-0
  preview route, required ruling fields, repository/account/operations/sunset option sets, and a
  readiness table. No option was adopted and no owner-authored design body was edited.
- Updated the dependency graph and executable queue so research follows decisions/prototypes rather
  than answering them. Made no product/test/schema/RFC/design/canonical-product-doc/player-copy edit;
  did not push or touch the user's `AGENTS.md` edit.

## 2026-08-20 — Dependency and shared-resource graph

- Re-read all active RFC headers against the completed lifecycle dossiers, implementation reality,
  release decisions, research prerequisites, and actual browser/runtime composition.
- Replaced the feature-list sketch with a release-authority, runtime/player-surface, and gameplay/
  content DAG. Classified every active RFC by its actual graph position and exact current gate.
- Added `dependency-resource-ledger.tsv`: 30 bounded resources with one canonical owner/gate,
  current producer, consumers, transformation/binding, refusal owner, integrated witness owner,
  state, and exact blocker.
- Confirmed RP-090: multiple headers treat optimistic whole-RFC status as a dependency and omit
  required edges. In particular, UI lacks API/Account authority; Commons lacks Account/API/Game UI;
  World lacks D-016/R-009 and public/player consumers; Combat Bots lacks Minigame integration; and
  Deployment omits client/license/Account/API/operations contracts.
- Preserved single-owner boundaries: API wire/client generation, Account credential/data rights,
  Transport delivery recovery, Game UI presentation, distinct operations metrics/gameplay
  telemetry, and one versioned release manifest.
- Marked Wave 6 complete at the coordinate. No RFC header/body was edited because the dependency
  defects require their ruling authors; no product/test/schema/design/canonical-doc/copy edit was
  made, and no push or `AGENTS.md` touch occurred.

## 2026-08-20 — Executable program and cross-party review handoff

- Reconciled the dependency ledger with the executable queue and the three genuinely accepted-
  scope READY batches. Detected RP-091: the old handoff still called authority-blocked R-001 and
  already-complete lifecycle reconciliation the smallest next batch.
- Added `ready-batch-manifest.tsv` for Q-001 Account witnesses, Q-002 Minigame API backend witnesses,
  and higher-risk Q-003 Transport production recovery. Each row records authority, exact scope,
  forbidden scope, failing controls, cold population, closeout/review protocol, and file/test
  conflict order. The batches are serial, not a parallel pile.
- Added `review-handoff.md` for the mandatory Claude-side designated adversarial pass. The core
  substantive range is `190a4fa..1e47752`; the reviewer must include these handoff commits and cite
  literal final hashes at review time. The checklist re-derives counts, promoted/contradicted rows,
  negative controls, release defects, all 91 RP routes, all 30 shared resources, owner-decision
  neutrality, READY authority, and the absence of product/RFC/archive changes.
- Marked Wave 7 complete and Wave 8 prepared/pending external designated review. No self-review was
  relabeled designated, no implementation or archival was authorized, no product/test/schema/RFC/
  design/canonical-doc/copy edit was made, and no push or `AGENTS.md` touch occurred.

## 2026-08-21 — Mechanical first filter and server boundary inventory

- Resumed after the workstation crash with branch `main` ahead of `origin/main` by 19 commits and
  only the user's tracked `AGENTS.md` edit dirty. Reconfirmed product evidence remains pinned to
  `190a4fa`; no product tree change was made.
- The first filter rejected the prior near-final sequencing claim: Wave 1 still explicitly left
  package, route, migration, client, catalog, executable, planning/docs, and archived-RFC
  populations incomplete, and Wave 2 still retained coarse design rows. Filed RP-092 and
  downgraded `review-handoff.md` to a bounded draft rather than relabeling a partial pass final.
- Recounted the server exactly. The prior 44-package/four-command summary was false: there are 45
  second-level package directories and five `main.go` entrypoints. Their 189 production files plus
  151 test files reconcile to all 340 Go files. Filed RP-093 and added
  `server-package-inventory.tsv` with production importers, runtime position, evidence class, and
  exact integration gap for every row.
- Enumerated all 24 exposed operations: three gameserver-hand infrastructure operations, 11
  account-hand `/api/v1` operations, and ten private-registry operations. Added
  `route-operation-inventory.tsv`. The production browser consumes only bootstrap, Game UI
  snapshot, intents, and WebSocket; 17 operations are backend-only, two are operator primitives,
  and `attach_email` remains unimplemented. This preserves backend-only routes as mechanical
  fragments rather than shipped player capability.
- Enumerated migrations 00001–00074 with domain ownership. Every file has exactly one Goose Up and
  Down marker. `make validate-migrations` passed cold against Postgres (`./save`, `-count=1`) and
  `make verify-schema` passed the current catalog/schema population. Added
  `migration-inventory.tsv`; explicitly retained isolated semantic rollback as open because marker
  presence and a current-chain aggregate cannot prove per-migration data/constraint restoration.
- Corrected stale `PROGRAM.md`, `acceptance-audit.md`, `plan.md`, `inventory.md`, and
  `inventory.tsv` claims. Active 111-row evidence reconciliation is complete with five exact
  review/provenance rows open; archived risk sampling is not complete. No self/delegated review was
  represented as the required Claude-side designated pass, and no push or `AGENTS.md` touch
  occurred.

## 2026-08-21 — Client entry graph and workflow inventory

- Traced `client/src/main.ts` through every ordinary static import and the Vite Worker URL, then
  reconciled that graph against the actual production sourcemaps. Added
  `client-source-inventory.tsv` with one row for every source file; its filenames exactly equal the
  82-file source population.
- The exact split is 41 files in the shipped entry/type graph and 41 outside it. The build emits 37
  authored JS/data/Svelte sources plus one CSS asset; three additional graph rows are type-erased
  compile contracts. Outside modules may remain legitimate parity/test/validator assets, but they
  are not mounted player consumers. Filed RP-094 to prevent client module breadth from being
  relabeled product breadth.
- Added `client-workflow-inventory.tsv` with 25 bounded default, failure, and accessibility paths:
  three proven-bounded, nine partial, one mechanical-only, eight backend-only, two absent, and two
  failed. It binds each workflow to a backend producer, production consumer, real/default data,
  witness, verdict, and exact route.
- Confirmed the shipped runtime makes exactly three HTTP calls—bootstrap, Game UI snapshot, and
  intents—plus the WebSocket connection. Refresh, Account rights/recovery, New Founder, Guild,
  Minigame, Soul, leaderboard/route readers, and later-domain surfaces have no production client.
  The client graph therefore agrees with the 24-operation inventory rather than merely with
  generated operation types.
- Confirmed `ShellController` records reconciliation and Worker metrics, but only a unit test calls
  `telemetry()` and no production exporter exists. Filed RP-095 under D-016/R-009 rather than
  claiming operational/gameplay instrumentation.
- The first root `pnpm --filter` build attempt was invalid because the repository has no root
  package manifest; it is excluded. The required root target `make build-client` then passed,
  transforming 152 dependency modules and emitting main JS/CSS plus the prediction Worker.
  `make typecheck verify-client-boundary` passed with zero diagnostics and the declared structural
  boundary population green.
- Updated the program/inventory/plan and both backlog ledgers. No product/client/test/RFC/design/
  canonical-product-doc/copy/catalog edit was made; no push or `AGENTS.md` touch occurred.

## 2026-08-21 — Balance, epoch, and copy-source boundary inventory

- Added `balance-file-inventory.tsv` and reconciled its filenames exactly against all 91 balance
  files: 19 deploy-current epoch artifacts, four platform configs, 15 schemas, and 53 test,
  candidate, historical, positive/negative, or measurement fixtures. No fixture was promoted to
  live content by harness use.
- Added `catalog-family-inventory.tsv` for the 23 live/platform families. The epoch seed owns 19
  artifacts; Client Shell, Transport, and epoch identity are additional composed platform inputs;
  API policy remains test-only/uncomposed. Each family names its schema/validator, loader, server
  consumer, client consumer, and exact capability gap.
- `make epoch-hash` reproduced epoch 8 hash
  `sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`,
  matching `deployment/content-manifest.v1.json` and the current epoch registry.
- Added `copy-file-inventory.tsv` and reconciled it exactly against all ten copy files. Read the
  generator rather than trusting names: it merges every `copy/catalog/*.json`, so the three
  `*-candidate.json` sources ship alongside `phase0.json`. The resulting 208 keys split 13 + 143 +
  1 + 51 across those files. Filed RP-096 for the absent machine-enforced held/shipped boundary.
- Ran `make copy-check` green. It verified the 208-key artifact, copy hash
  `sha256:9e816294f84e6c50e5050f59d33f49a1f09d42780abdb7f5b40bb5a5442c0e13`,
  Game UI presentation artifacts, and deployment manifest, while reporting 161 orphan warnings.
- Proved the orphan output is non-discriminating without mutating product code: it calls direct
  production keys in `GameUIApp.svelte` and `RunEndSurface.svelte` orphaned because the reference
  model covers selected epoch fields and two explicit Go sites, not client `t()` calls. Filed
  RP-097; the report cannot authorize cleanup or consumption claims until live/orphan negatives
  discriminate.
- Updated the program/inventory/plan/review draft and both backlog ledgers. No balance/copy/
  generated artifact/product/test/RFC/design/canonical-product-doc edit was made; no push or
  `AGENTS.md` touch occurred.

## 2026-08-21 — Make, CI, and client-test evidence inventory

- Added `make-target-inventory.tsv` and reconciled it exactly against all 72 names on the root
  `.PHONY` line. Only 26 currently have bounded-valid evidence. Ten are intentional mutators, nine
  are partial, five are release-invalid aggregates, and the rest are setup, aliases,
  manual/historical measurements, parameterized tools, or pending/authority-blocked lanes. Filed
  RP-100 so target count cannot stand in for executable proof.
- Added `ci-job-inventory.tsv` for all seven jobs with trigger scope, timeout, command, current
  hosted verdict, and exact capability limit. Preserved run `32404232364`: server, client, browser,
  game-ui-composed, and schema passed; harness cancelled at 30m02s; numeric was skipped. The
  workflow still has no current aggregate verdict and scheduled/manual numeric failure is blocking.
- Added `client-test-artifact-inventory.tsv` and reconciled it exactly against the current 56-file
  local tree. Only 43 TypeScript sources are tracked. Thirteen PNGs are generated under ignored
  `client/test/__screenshots__/`, no test calls a screenshot assertion, and a fresh clone cannot
  receive them. Filed RP-098 and split the inventory rather than treating local captures as
  acceptance artifacts.
- Classified every client test source by kind, subject, production relationship, and evidence
  limit. Two are browser suites; one is the browser error helper; one is a typecheck fixture; the
  other 39 are unit/parity/candidate/measurement sources. Module parity remains distinct from a
  mounted production workflow.
- Found two root recipes that contradict the current command/cache protocol:
  `verify-routes-boundary` and `verify-commons-boundary` create task-named `/tmp` Go caches instead
  of consuming the exported repository-local `GOCACHE`. Filed RP-099 without editing the Makefile;
  implementation requires accepted scope and a failing import fixture.
- Updated the program/inventory/plan/review draft and both backlog ledgers. No Make/workflow/
  product/test/RFC/design/canonical-product-doc edit was made; no push or `AGENTS.md` touch
  occurred. The remaining executable Wave-1 population is all 151 server test files and their
  row-level oracles/negative controls.

## 2026-08-21 — Server test-file and skip-denominator inventory

- Added `server-test-file-inventory.tsv` and reconciled it exactly against all 151 `*_test.go`
  files. The population contains 591 top-level `Test*` functions and one fuzz target: 25
  filename-declared integration files, 124 unit/package files, one corpus/fuzz file, and one
  fixture helper with no test function. Fifteen files directly show DB/HTTP/WebSocket composition
  signals.
- Kept this classification explicitly structural. Neither an `_integration_test.go` suffix nor a
  top-level function count establishes a discriminating oracle. Existing lifecycle dossiers and
  Batches A–N retain the semantic mutations already performed; remaining row-level oracle review
  is still open.
- Added `server-test-skip-inventory.tsv` for all 40 explicit skip sites in 28 files. Thirty-nine
  sites across 27 files skip on missing `TEST_DATABASE_URL`; one Commons projection property skips
  on an architecture-sized target. Filed RP-101 because ordinary host `make test-go` can exit green
  while silently excluding the Postgres population.
- Preserved lane truth: Postgres integration claims require the declared Docker/hosted service,
  whereas ordinary package runs prove only their executed non-skipped denominator. No environment
  convenience was used to promote a local green.
- Updated the program/inventory/plan/review draft and both backlog ledgers. No server/product/test/
  Make/workflow/RFC/design/canonical-product-doc edit was made; no push or `AGENTS.md` touch
  occurred.

## 2026-08-21 — Planning-thread and canonical-doc inventory

- Added `planning-thread-inventory.tsv` and reconciled it exactly against all 23 top-level
  planning directories. Each row names tracked/local file counts, authority, current state, and
  exact closeout gap. Twelve are active RFC planning threads; the remaining rows are archive,
  provenance, maintenance, historical measurement/review, a blocked draft child, or this audit.
- Identified four threads that explicitly describe their work as complete, withdrawn-history-only,
  or superseded while remaining live: `archived-four-review`, `harness-dispatch-cardinality`,
  `production-review-round2`, and `run-genesis-archival-remediation`. Filed RP-102. No directory was
  moved because exact provenance and authorized transactional closeout must precede archival.
- Reconciled the fresh-clone boundary. After this checkpoint there are 207 tracked planning files
  and 25 ignored/local-only files: 17 coverage-map records, six archived T0-T1 diagnostics, one
  platform-alignment diagnostic, and one historical Codex fix record. Deepened RP-015 rather than
  granting evidence credit to local files.
- Added `docs-file-inventory.tsv` and reconciled it exactly against all 38 canonical/generated docs.
  Every file now names its owning system, artifact kind, truth class, and evidence/repair route.
  Only the numeric foundation is currently unqualified; three generated artifacts are
  drift-checked and all partial/stale/contradicted claims retain their existing RP routes.
- Updated the program/inventory/plan/review draft and both backlog ledgers. No canonical product
  doc, product code, test, Make/workflow, RFC body, balance/copy/catalog, migration, or deployment
  file was edited; no archival move, push, or `AGENTS.md` touch occurred.

## 2026-08-21 — Runtime concurrency, background-job, and event-family inventory

- Added `runtime-concurrency-inventory.tsv` and traced the deployed gameserver from `main` through
  composition, startup, readiness, failure, drain, and shutdown. The repository-authored runtime
  launches exactly 11 long-lived goroutine instances: HTTP listener, root-context monitor, player
  relay, six attached jobs, job waiter, and transport world coalescer. Dependency-internal
  Centrifuge/HTTP goroutines are explicitly excluded from the count.
- Mapped all six composed jobs with their prime pass, cadence, producer/consumer boundary, failure
  behavior, shutdown, and witness: world aggregation, replay verification, guild presence relay,
  guild clearing, guild disband sweep, and credential GC. Retained RP-086 because the separate
  30-day intent-record pruner still has no production caller.
- Filed RP-103. Binding `design/06` promises a goroutine-per-player actor and one World goroutine,
  but no player actor/type/mailbox/launch exists and runtime world work is split between aggregation
  and transport coalescing. Match actor and Matchmaker are also absent and explicitly deferred by
  the accepted Minigame Platform text. This requires author/authority reconciliation, not an
  implementation guess.
- Added `event-family-inventory.tsv`. It bounds the 48 player EventKinds, two outbox message kinds,
  five transport envelope kinds, three system codes, five disconnect codes, six decoded Game UI
  lifecycle event kinds, seven guild-domain kinds, two presence kinds, six replay verdicts, five
  projection idempotency ledgers, and three runtime invariant kinds. The Game UI's six lifecycle
  decoders remain distinct from backend event-registry breadth.
- Ran `make test-go GO_PACKAGES='./gameserver ./transport ./routeprojection'
  GO_TEST_FLAGS='-count=1'`; all three packages passed cold. Separate structural checks proved the
  48-kind registry, six explicit source launch sites resolving to 11 instances, six attached jobs,
  and that `ExpireNames` occurs in exactly one server file (its own declaration). This does not
  promote skipped Postgres cases or source counts into semantic acceptance.
- Filed RP-104: Route Registry's documented 72-hour naming reservation expiry is represented by
  `Projector.ExpireNames`, but the method has no production caller, test caller, cadence, batch,
  failure owner, or metrics. No scheduler was invented without accepted authority.
- Updated the program/inventory/plan/review draft and both backlog ledgers. No product/server/client/
  test/Make/workflow/RFC/design/canonical-product-doc/balance/copy/catalog/migration/deployment edit
  was made; no archival move, push, or `AGENTS.md` touch occurred.
- Reconciled the self-changing planning denominator in the same checkpoint: after the two new
  ledgers land, there are 209 tracked planning files, 25 ignored/local-only records, and 55 tracked
  plus one ignored file inside the 56-file platform-alignment thread.

## 2026-08-21 — Archived-RFC replay predeclaration

- Added `archived-rfc-risk-plan.md` before classifying the archive population. It freezes the exact
  46-RFC/46-planning population, three historical slug mappings, additive risk weights, mandatory
  minimum-15 and ten-domain deep strata, three low-risk controls, deep trace, fired criteria, exit
  conditions, and authority limit.
- An exploratory read had already exposed a possible Fiscal Quarters plan/body contradiction. The
  plan explicitly withholds finding/sample credit until Fiscal is replayed under the same declared
  rules; the observation was not used to weaken thresholds or choose a convenient population.
- The score is triage, not a retroactive verdict: missing current-style `Reviewed range:` tokens
  raise provenance risk but do not automatically invalidate archives completed before the current
  protocol. Negative fixture-only/backend-only/superseded results remain valid research outputs.
- No archive/RFC body, product, test, canonical product doc, plan checkbox, or authored copy was
  changed. After this checkpoint there are 210 tracked planning files and 56 tracked plus one
  ignored file in the platform-alignment thread; no push or `AGENTS.md` touch occurred.

## 2026-08-21 — Archived-RFC structural and deep replay

- Added `archived-rfc-risk-inventory.tsv`; all 46 archived RFC files reconcile one-to-one with 46
  planning archives after exactly the three predeclared slug mappings. Every row records status,
  RFC lines, open plan boxes, review/range tokens, active direct references, canonical docs,
  existing RP ownership, formula-derived score, selection reason, and verdict. An independent AWK
  recomputation found zero score mismatches.
- Completed the mandatory 20-row deep sample in `archived-rfc-risk-audit.md`. It contains both
  open-plan archives, all nine rows scoring at least nine, all required current-CI/dependency risk
  owners, numeric/economy/save/production/client/content/runtime/multiplayer/replay/harness domains,
  and four low-risk controls. The other 26 rows remain explicitly structural-only rather than
  receiving inferred semantic approval.
- Filed RP-105: Fiscal is archived with all seven implementation/review boxes open and with F1/F11
  commit-under-rejection clauses contradicting the accepted rollback correction, code, tests, and
  canonical docs. The ruling/record authors own reconciliation; the audit edited no frozen body or
  checkbox.
- Filed RP-106: Copy Amendment A1 ends `not approved` in the Copy owner log; `ebb081f` and its
  cross-party A1-F1 closure exist only in the Meters archive. Retained RP-097 because current
  `make copy-check` passes at 208 keys while its 161-key orphan report calls live Game UI `t()`
  sites orphan.
- Filed RP-107: `docs/soul.md` says no Soul artifact/recovery activity is live while the current
  epoch and `docs/soul-recovery.md` prove `defrag`, `repot`, and `server_room` are minted. The Soul
  Foundation archive remains historically valid fixture-first; the successor closeout is stale.
- Downgraded T0–T1's player-facing “end to end” wording without weakening its genuine backend
  evidence: the archive's composed proof is explicitly Chromium-free, while current browser audit
  proves the production UI stops at Desk and lacks Gate/Wind Down/continue. RP-008/RP-033 already
  own the missing consumer and lifecycle split.
- Cold selected Go packages passed with `-count=1`; the harness package completed in 27.855 s.
  `make test-save-integration` passed through the declared Docker/Postgres service and exposed its
  real executed/no-test denominators. `make copy-check` passed with the non-discriminating warning
  qualification. A focused client Fiscal run emitted no denominator for more than 60 seconds and
  was terminated, so it is recorded invalid rather than green or used to change a budget.
- Reconciled program/inventory/plan/reality/execution-queue/review-draft, the affected row ledgers,
  and both backlog ledgers. After the two
  new tracked artifacts land, planning contains 212 tracked plus 25 ignored/local-only files; the
  platform-alignment thread contains 58 tracked plus one ignored file. No product, test, Make,
  workflow, RFC/archive body, canonical product doc, balance/copy/catalog, migration, deployment,
  plan-checkbox, push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Copy-key consumption predeclaration

- Added `copy-key-consumption-plan.md` before generating any row classification. It freezes the
  exact 208-key generated-catalog population; mounted literal/dynamic, registered current-artifact,
  explicit Go producer, other strict catalog, test/tool, and no-reference lanes; six closed
  verdicts; exact row fields; manual ambiguous-row review; failure rules; and authority limit.
- Predeclared controls force the known live keys that the current orphan report mislabels to resolve
  as mounted, keep deploy-current achievement/Pitch/Soul copy backend-only without a player surface,
  keep FixtureHost-only calls non-production, reject an injected absent fake key, and disclose
  unbounded dynamic selection as ambiguous. Catalog/type presence cannot earn player-capability
  credit.
- This checkpoint changes no copy source/generated report/tool, product code, tests, canonical
  product docs, RFC/archive body, balance/catalog, migration, deployment, plan checkbox, or authored
  prose. After it lands there are 213 tracked planning files and 59 tracked plus one ignored file
  in platform alignment; no push or `AGENTS.md` touch occurred.

## 2026-08-21 — Copy-key producer/consumer measurement

- Added the reproducible `copy-key-consumption-extractor.mjs`, its exact 208-row TSV, and the
  bounded audit. A rerun reproduced the checked-in inventory byte-for-byte. It emits the full
  denominator, 161 current warnings, 14 bounded computed resolver sites, and visible verdict
  counts; parser/denominator/dynamic drift fails loud.
- Classified all generated keys: 128 mounted player copy, 63 deploy-current backend/data-only, one
  shipped-but-currently-unselectable binding, eight fixture/tool-only event rows, and eight bounded
  no-reference candidates. The manual second pass covered every non-mounted row, every dynamic
  binding family, and all computed call sites; no real row remains ambiguous.
- Demonstrated the instrument's negative controls in the recorded run: it rejected a dropped row,
  a mounted key relabelled unused, and a backend-only key relabelled mounted. The absent fake is
  unreferenced, while an unbounded synthetic selector becomes `ambiguous_dynamic`.
- The current orphan report labels 105 mounted keys orphan. Its 47 `referenced` rows are exactly 45
  keys reached by the seven registered artifact paths plus the two explicit Go sites, but 24 of
  those are still backend/data-only. It therefore proves registry membership only, not use or
  player reachability; RP-097 is now numerically bounded.
- Filed RP-108: 50 exact generated-key references in current Pitch (26), Soul (16), Curriculum
  (seven), and Minigames (one) artifacts sit outside the registry. This is why the report also calls
  mounted Curriculum and 39 backend/data rows orphan.
- Filed RP-109: current Fiscal and Opportunities artifacts carry four exact hardcap reason keys
  absent from the application catalog: `cap.fiscal_credit`, `cap.fiscal_level.beige_tower`,
  `cap.cash`, and `cap.active_combo`. Economy upgrade prefixes were explicitly expanded and are not
  part of this mismatch. The audit did not invent the missing owner-authored prose.
- Reconciled program/inventory/plan/reality/queue/review-draft and both backlog ledgers. After the
  three new tracked artifacts land, planning contains 216 tracked plus 25 ignored/local-only files;
  platform alignment contains 62 tracked plus one ignored file. No copy source/report/generator,
  balance artifact, product code, test, canonical product doc, RFC/archive body, migration,
  deployment, plan checkbox, push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Capability atomization predeclaration

- Added `capability-atomization-plan.md` before producing child rows. It freezes all 121 unique
  parent IDs across the 14 binding design documents, an actor/producer/consumer/data/workflow/
  witness/failure split grammar, exact 14-column output, closed verdicts, coverage checks, and the
  manual second-pass population.
- Mandatory controls prevent the known umbrella rows for multiplayer fallbacks, Tier 2, Founder
  currencies, minigame contracts, guilds, anti-cheat, Run End, and content-family extensibility from
  receiving one convenient verdict. Atomic foundation contracts may remain single only when their
  witness and failure boundary are actually shared.
- Parent preliminary states are explicitly non-authoritative leads. No child may inherit “proven”
  from a broad row, file/module, RFC archive, generated type, fixture, or aggregate green check.
- This checkpoint changes no design intent, product, test, canonical product doc, RFC/archive body,
  balance/copy/content, migration, deployment, plan checkbox, or authored player prose. After it
  lands there are 217 tracked planning files and 63 tracked plus one ignored file in platform
  alignment; no push or `AGENTS.md` touch occurred.

## 2026-08-21 — Structural capability atomization

- Added a reproducible structural extractor and inventory. All 121 parent IDs map to 432 unique,
  sequential child outcomes: V 32, T 55, E 84, M 36, P 17, S 28, A 22, roadmap 30, flavor 19,
  events 21, playstyles 21, UX 32, content pipeline 21, and world 14.
- One hundred twelve parents split; only nine single-purpose rows remain atomic (eight named/deferred
  minigame/social features and one missing Soul-recovery section). Mandatory controls split
  multiplayer fallback, Tier 2, six Founder currencies, minigame contracts, guilds, six anti-cheat
  boundaries, Run End beats, and content families.
- The recorded run rejected three seeded structural failures: removal of an atomic parent's only
  child, duplicate child identity, and collapsing the required anti-cheat split to one row. A rerun
  reproduced the checked-in TSV byte-for-byte.
- This checkpoint is intentionally structural: `parent_preliminary_state` and `parent_route` are
  carried only as review leads. No child is called proven/partial/absent until the 14-column outcome
  ledger attaches the predeclared producer, consumer, current-data, workflow, executable-witness,
  failure/refusal, verdict, route, and limitation lanes.
- After these two new tracked artifacts land, planning contains 219 tracked plus 25 ignored/local-
  only files; platform alignment contains 65 tracked plus one ignored file. No design intent,
  product, test, canonical product doc, RFC/archive body, balance/copy/content, migration,
  deployment, plan checkbox, push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Atomic capability evidence attachment

- Evidence attachment caught and corrected one defect in the committed structural denominator:
  `P-004.02` had combined exact cosmetic odds with dark-pattern disclosure even though the former
  is absent and the latter is mounted. The structural extractor now emits 433 children (Pets 18),
  not 432 (Pets 17). The prior checkpoint remains an append-only record of the discovered error.
- Added `capability-outcome-ledger.tsv`, `capability-outcome-validator.mjs`, and
  `capability-outcome-audit.md`. Every child has the predeclared actor, producer, consumer,
  deploy-current data, default workflow, executable witness, failure/refusal, verdict, authority
  route, and evidence limitation. No parent preliminary state grants proof.
- Final distribution: three `proven_integration`, 41 `proven_bounded_primitive`, 55
  `partial_integration`, 134 `backend_or_data_only`, seven `client_or_fixture_only`, three
  `claimed_only`, 188 `absent`, and two `blocked`. The three integrated rows are three binding
  design views of one server-anonymous bootstrap path, not three independent gameplay systems.
- The manual promoted/partial and mandatory-family pass split several false umbrella claims:
  server authority from absent public formulas; manual multiplication from absent automation;
  mounted disclosure from absent odds/acquisition/equipment; early Run End fields from absent
  continuation and unmounted route consequence; registry data from absent category choice; and
  semantic fixture structure from failed/unevidenced accessibility workflows. Historical pacing
  results cannot promote the current coordinate.
- Re-ran the structural extractor and compared its 433-row output byte-for-byte with the checked-in
  inventory. The outcome validator accepted all 433 ordered 14-field rows and rejected seeded
  dropped-outcome, duplicate-outcome, vacuous-promotion, and missing-route cases. Exact verdict
  recount reproduced the audit table.
- Reconciled the capability map, reality audit, program/inventory/plan, execution queue, tracked RP
  route, and review draft. Wave 2 is complete at `190a4fa`; row-level gameplay-content and semantic
  oracle populations remain before the final designated cross-party review.
- After the three new tracked artifacts land, planning contains 222 tracked plus 25 ignored/local-
  only files; platform alignment contains 68 tracked plus one ignored file. No design intent,
  product, test, canonical product doc, RFC/archive body, balance/copy/content, migration,
  deployment, implementation-plan checkbox, push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Deploy-current gameplay-content row predeclaration

- Froze the next audit population before row extraction: exactly the 19 files already classified
  `epoch_artifact` in `balance-file-inventory.tsv`. Schemas, platform configs, fixtures, Copy, and
  design prose retain their separate audited populations.
- `gameplay-content-row-plan.md` defines a generic JSON-unit grammar rather than hand-picking known
  IDs: every non-root object, primitive array edge, empty collection, and root policy scalar is
  retained. Empty Soul sources, empty route sets, zero-weight lists, and similar missing content
  therefore cannot disappear by construction.
- Predeclared 14 output lanes, nine closed verdicts, exact producer/consumer/current-trigger/witness
  requirements, seeded denominator/vacuity controls, the promoted/dormant/zero/manual review set,
  and planning-only authority. No content verdict or denominator count has been measured yet.
- After this predeclaration lands, planning contains 223 tracked plus 25 ignored/local-only files;
  platform alignment contains 69 tracked plus one ignored file. No product, test, design/RFC body,
  balance/copy/content, canonical product doc, migration, deployment, implementation-plan checkbox,
  push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Deploy-current gameplay-content structural population

- Added the deterministic JSON-unit extractor, 579-row structural ledger, and bounded structural
  audit. All 19 epoch artifacts reconcile exactly. Units split into 297 array-object rows, 135
  nested objects, 48 primitive edges, 39 empty collections, 43 root policies, and 17 singleton
  policies.
- Family counts are achievements 50, economy 175, categories 21, Commons 21, curriculum seven,
  doctrines three, factions 11, Fiscal six, guilds ten, meters 56, Minigame API five, minigames 16,
  opportunities nine, pets 23, Pitch 37, Prestige 19, Relevance 55, routes 45, and Soul ten.
- The 39 explicit empties include Relevance groups/edges, meter inputs, faction modifier slots,
  early-gate routes, the final generator ladder, category completion facts, Commons source weights,
  and Soul debit sources. They receive no automatic defect or acceptance verdict, but cannot vanish
  from the subsequent evidence denominator.
- A rerun reproduced the structure byte-for-byte. Seeded dropped-unit, duplicate-unit,
  empty-collection omission, and root-policy omission cases all failed. Payload hashes and valid
  epoch inclusion remain structural facts only.
- After these three new tracked artifacts land, planning contains 226 tracked plus 25 ignored/local-
  only files; platform alignment contains 72 tracked plus one ignored file. No product, test,
  design/RFC body, balance/copy/content, canonical product doc, migration, deployment,
  implementation-plan checkbox, push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Gameplay-content verdict vocabulary correction

- Before attaching any of the 579 row verdicts, source/witness inspection exposed a missing state in
  the predeclaration: several deploy-current rows have a real producer and mounted consumer, but
  only separate server/component fixtures rather than a row-discriminating default-workflow proof.
- Added `partial_mounted` rather than falsely forcing those rows into either `proven_mounted_*` or
  `backend_active`. Every use must name the missing selection/effect/refusal/integrated witness.
  The structural denominator and authority limit are unchanged; no content verdict exists yet.

## 2026-08-21 — Deploy-current gameplay-content evidence attachment

- Added the 579-row evidence ledger, validator, and audit. Final distribution: zero proven mounted
  effects/presentations, 173 partial-mounted, 180 backend-active, 141 registered-dormant, 55
  measurement-only, 21 zero/empty placeholders, and nine contradicted.
- Zero proven mounted rows is a witness result, not a claim that the UI has no data. The composed
  browser proof stops at Desk without a gameplay transition; screen tests inject snapshots/events;
  server content tests do not mount the production browser. Those lanes support partial rows only.
- The dormant pass traces pets with no acquisition, meters with no current source, later-tier
  routes/economy/offers/achievements, and lower Soul bands/recovery made unreachable by empty debit
  sources. All 39 structural empties remain classified rather than removed from the denominator.
- Filed RP-110 in both ledgers: `upgrade.reply_all_macro` multiplies a manual action but binding
  Tier-1 intent calls it automation. Exact current contradictions also retain RP-109 for four
  missing hardcap reason keys and RP-017 for the three Soul recovery rows.
- Reproduced structural payload identity and ran the verdict validator. Seeded dropped/duplicate,
  backend-as-mounted, empty-as-active, and missing-route cases all failed.
- After the three new tracked artifacts land, planning contains 229 tracked plus 25 ignored/local-
  only files; platform alignment contains 75 tracked plus one ignored file. No product, test,
  design/RFC body, balance/copy/content, canonical product doc, migration, deployment,
  implementation-plan checkbox, push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Row-level test-oracle predeclaration

- Froze the final semantic evidence population before declaration extraction: all 592 top-level Go
  Test/Fuzz functions in 151 reconciled files plus every static `it`/`test` declaration in the 43
  tracked client test/helper sources. Zero-declaration helpers receive explicit non-oracle units.
- Predeclared exact body identity, subject/lane/fixture/dependency/assertion/negative-control lanes,
  nine verdicts, mandatory skip/mock/positive-only limits, seeded failures, and the manual promoted/
  browser/composed/Postgres/current-data/acceptance population.
- Known fired failures are controls, not optional anecdotes: Game UI outcome-removal, API registry
  and client-lint blindness, invalid hosted harness/cache state, focus/320 px failures, and ordinary
  host Postgres skips must remain non-green in the row result.
- After this predeclaration lands, planning contains 230 tracked plus 25 ignored/local-only files;
  platform alignment contains 76 tracked plus one ignored file. No product/test, design/RFC body,
  balance/copy/content, canonical product doc, migration, deployment, implementation-plan checkbox,
  push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Row-level test-oracle structural population

- Added the deterministic declaration/body extractor, 802-row structural ledger, and bounded
  structural audit. The frozen population reconciles 591 Go tests, one Go fuzz owner, 174 plain
  client tests, 19 parameterized client declarations, 15 conditional browser declarations, and two
  zero-declaration helper/type-contract units.
- The Go owners contain 88 static subtest calls and one fuzz seed. Twenty-five integration files own
  47 functions. Eighty-four bodies have direct dependency/guard signals and 421 units have broad
  negative keywords; both are inspection leads, not semantic verdicts.
- A rerun reproduced all declaration identities and body hashes byte-for-byte. Seeded dropped,
  duplicate, helper-omission, and body-hash-drift cases failed. The extractor also refuses drift
  from the prior 151/592 server and 43-source client populations.
- After these three new tracked artifacts land, planning contains 233 tracked plus 25 ignored/local-
  only files; platform alignment contains 79 tracked plus one ignored file. No product/test,
  design/RFC body, balance/copy/content, canonical product doc, migration, deployment,
  implementation-plan checkbox, push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Row-level test-oracle evidence attachment

- Added the 802-row semantic ledger, validator, and audit. Final distribution: zero unconditional
  integrated, 171 bounded-discriminating, 533 positive-only, 43 fixture/mock-only, 51 dependency-
  conditional, one non-discriminating, one invalid/guarded, and two helper units.
- Rejected an initial blanket Decimal/Combat promotion before publication. Nineteen rows without an
  explicit rejection or an exact previously fired mutation remain positive-only; only eight exact
  arithmetic/vector mutation rows receive bounded credit.
- Preserved every mandatory failed control: the Game UI axe row stays non-discriminating, API
  generation is bounded to registered operations, focus/320 px probes remain absent from the green
  browser population, command-level client/harness failures are not rehabilitated by unit rows, and
  all 51 real-Postgres functions remain conditional outside their dedicated lane.
- The validator reproduced all identities/body hashes and rejected seeded dropped, duplicate, body-
  drift, conditional-as-integrated, helper-as-positive, blind-oracle-promotion, and missing-route
  cases.
- After these three new tracked artifacts land, planning contains 236 tracked plus 25 ignored/local-
  only files; platform alignment contains 82 tracked plus one ignored file. No product/test,
  design/RFC body, balance/copy/content, canonical product doc, migration, deployment,
  implementation-plan checkbox, push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Final Codex-side contradiction pass and review handoff

- Reproduced the 433-child, 208-key, 579-unit, and 802-oracle structural outputs byte-for-byte and
  reran all three semantic validators. Every seeded corruption fired.
- Independently recomputed 111 active acceptance rows, 30 dependency/resource rows, three READY
  batches, and the contiguous RP-001–RP-110 ledger. Recounted 23 planning threads, 38 docs, 70 RFC
  Markdown files including 46 archives, 340 Go files, and 82 client sources.
- Audited the 41-commit/86-path range through `871c86a`: only root/current-status/index and platform-
  alignment planning paths changed. The final closeout adds two audit artifacts; the designated
  reviewer must resolve and cite the literal post-commit tip.
- Reconciled two stale audit-status claims: row semantics are no longer pending in `inventory.md`,
  and archived-RFC risk sampling is no longer pending in `acceptance-audit.md`. No append-only
  historical checkpoint was rewritten and no owner-authored normative contradiction was touched.
- Final shared-memory count is 238 tracked planning files, 84 in platform alignment, plus 25
  ignored/local-only planning files (one in platform alignment). The next action is the mandatory
  designated Claude review; Codex has not approved its own work.

## 2026-08-21 — Designated-review follow-up and README normalization

- Applied designated review Finding A without changing any row verdict: every current summary now
  says the zero-integrated oracle result is declaration-isolated and excludes four externally
  recorded T0–T1 severing probes. Those probes remain valid integrated backend evidence while the
  dependency skips and incomplete default browser workflow remain explicit.
- Replaced the root README's audit-summary shape with a conventional project entry point: concise
  pitch/status, stack, requirements, setup/build/verification, honest Postgres-lane warning,
  gameserver runtime boundary, repository layout, development process, and license.
- Removed the long design-document catalog and internal research-policy essay from the README;
  canonical design, current-state, RFC, planning, and docs links remain discoverable through their
  owning indexes.
- No product code/test, design/RFC body, canonical product doc, balance/copy/content, migration,
  deployment behavior, implementation-plan checkbox, push, or `AGENTS.md` edit occurred.

## 2026-08-21 — Q-003 Transport consumer execution and owner blocker

- After Q-001 approval `34d04a5` and corrected Q-002/API approval `bfd9b65`, predeclared Q-003 at
  `a94110e` and implemented the accepted D4/T4 browser recovery controller plus AC6 Guild negative
  at `c63e7e6`.
- Cold unit, cross-browser, boundary, build, composed-browser/Postgres, and sequential
  Transport/Account/Gameserver populations pass. Recovery-command, recovered-publication, and
  forced-open Guild mutations all fail their intended witnesses.
- Narrowed RP-052/RP-053 to implemented-pending-review. Narrowed RP-054 to the remaining measured
  contradiction: the literal 11.29-second connected stall closes on Centrifuge byte-queue pressure
  before the post-admission stale hook can resume exactly one frame. AC6 is no longer bundled into
  that open finding.
- Execution row 3d is now blocked on ruling-author action. No later serial batch, designated Q-003
  approval request, plan checkbox, RFC promotion/archive, or push is authorized until AC3 receives
  either a pre-queue implementation authority or body reconciliation.
- The old final-contradiction validator now fails its deliberate planning-only-range assertion on
  the subsequently authorized Q-001–Q-003 product paths. That expected coordinate guard is not a
  Q-003 verification failure and is recorded rather than bypassed or relabeled green.

## 2026-08-21 — Q-003 AC3 supported-seam check

- Exhausted the pinned Centrifuge v0.38.0 public/experimental seams after the literal stall failed.
  Its latest-publication batching coalesces only within a fixed channel-wide delay before queue
  admission; ten seconds breaks healthy `world` cadence, while a short delay still fills a stalled
  client's queue. The writer exposes no per-client queued-item replacement or consumption ACK, and
  `OnTransportWrite` remains post-admission/pre-socket-write.
- Kept 3d blocked rather than introducing a custom writer, changing the dependency, weakening the
  4 Hz contract, or pretending typed-overflow recovery satisfies exact-one connected delivery.
  RP-054 now has exhausted implementation evidence for its two lawful owner choices.

## 2026-08-21 — Transport AC3 owner ruling

- Marco ruled the exact packet count to be inappropriate for this idle-game outcome. Reconciled the
  RFC body to require a bounded ten-second stalled-client path that converges to newest authoritative
  state, either live or after typed 4000 reconnect/full sync, without losing committed progress.
- Promoted queue row 3d from owner-blocked to READY for only that literal witness and mutation. Q-003
  still needs restored cold gates and the mandatory minimal exact-range cross-party verdict; no
  later serial product batch, archival, deployment, or push is authorized yet.

## 2026-08-21 — Transport AC3 first ruled-witness correction

- The first ten-second run failed at its over-strict close-code assertion: the fully blocked socket
  surfaced EOF because it could not deliver its own 4000 close frame. Corrected the owner ruling's
  transcription from “typed 4000 required” to “recover after typed or abnormal disconnect”; bounded
  convergence and no-loss obligations are unchanged. The failure is retained as evidence rather
  than hidden or made green by weakening the player outcome.
