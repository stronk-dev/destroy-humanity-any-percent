# Current repository state

Reconciled 2026-10-15 at implementation checkpoint `654523c7` (kernel0.3.216, epoch8).
GitHub main and Actions metadata were rechecked on that date: published main remains
`be66902b`, with completed failed CI38560135778. At this checkpoint local main is61
commits ahead; that is a publication coordinate, not a product-completion measure.
Cloud Clicker remains a **development snapshot**, not a released or
release-ready 1.0. The objective is the full nine-tier game through Transcendence plus recovery,
privacy, accessibility, rights, operations and preservation. A Phase-0 preview does not replace it.

<a id="whole-product-overview--2026-10-12"></a>
<a id="whole-product-overview--2026-10-14"></a>
<a id="whole-product-overview"></a>

## Whole-product overview — 2026-10-15

**Bottom line:** substantial server/client foundations and an executable early-game loop exist.
The current stage is **Garage integration and hardening**, not completion of the full game or
release readiness. Much of Tier 2 is fixture-first; Tiers 3–8 and the final Transcendence paths
still require construction. The current client era selector supports Tiers 0–2 and refuses later
tiers (`client/src/game-ui/contracts.ts`, `eraForSnapshot`). No percentage or release date is
justified by the number of commits, tests or archived RFCs.

This is a synthesis of the active RFC index, canonical docs, owning plans/logs, live decision
queue and current verification, not a claim to have rerun every acceptance criterion at HEAD.
The whole-product classification retains the 2026-10-12 baseline with the explicitly bounded
construction summarized below and in current verification; no additional tier or milestone is complete.
The61 commits after published `be66902b` primarily repair conformance, review/evidence gaps and test
reliability. They do not turn the incomplete Garage or substantial later-tier scope into a release.
“Fixture-first” means code exercised with controlled candidate data, not adopted release content.
“Locally verified” does not mean designated-approved, shipped or verified on a clean release host.

| Area | Actual progress | Still missing for the whole product |
|---|---|---|
| Engine and persistence | Shared Go/TS number handling, declarative economy, lazy production/offline accrual, versioned Postgres saves/migrations/replay, routes/gates and client worker foundations are implemented. | New feature arithmetic/replay must satisfy the same floor. Clout's partition/accrual defect remains open; core foundations do not make every new system correct. |
| Early-game player loop | Production-built browser client with real Postgres/WebSocket journey covers purchases, first gate, scripted and elective early endings, next Company and reconnect. Fiscal, active-play opportunity and Pitch are exercised through DOM controls. | Natural pacing/long-session recovery, remaining Garage acceptance and current-range designated reviews. The main journey verifies exact built client bytes, but not a Caddy/image/clean-host release artifact; accelerated fixture journeys are not complete nine-tier playthroughs. |
| Tier 2 / incorporation | Candidate economy/content, routes and era UI exist. Native candidate controls now work through the real service/Postgres: both gates, genuine first failure/next Company, three generators/three upgrades, incorporation and persisted/rendered reload, using explicit cash setups. This caught and repaired a real missing-presentation crash; the new range awaits designated review. Original C1–C4 and the C5 pacing subset have bounded Codex designated approval. Pacing/relevance reproduce but fail their floors. Headcount has a source-independent Go/TS allocation library, not live gameplay. | Natural pacing/relevance, RP-433's T0 policy-population conflict, adopted copy (including RP-437's misleading second-gate label), full player acceptance and content mint; epoch 9 is not minted. Headcount's hiring model (separate purchase versus generator-granted seats), direct-production versus multiplier-output model, and sibling contract reconciliation remain held; its live allocation system is not built. |
| Prestige / Reputation | Exit/reset foundations and earlier acceptance repairs are reviewed. Reputation nodes, plans, starters, persistence and UI have fixture-first implementations. Native purchase/cancel, unchanged nonzero current rate, next-run bonus, Exit plan and persisted third-Company starter/reload now have a production-browser/real-service journey with explicit budget/cash setup. | Career/relevance evidence remains red: the latest hosted maintenance run38507071157 fails career/relevance report comparison; earlier six starter timing ties remain recorded evidence, not a new diagnosis of this run. Run End's balance/carry display contract reconciliation, full feature review and adopted content also remain. The new fixture is not earned-budget pacing or full career verification. The two early endings are not the three designed final endings. |
| Minigames | Pitch has a real API-to-player-to-payout journey. In addition to graceful application reconstruction, the browser now survives an actual gameserver SIGKILL after a command commits but its reply is lost: identical automatic retry, preserved SQL snapshot/history, socket recovery and continued payout, without reload or another bootstrap. This is not a DB/host crash or clean release-host rehearsal. Solo and async platform fixtures both execute saved-state resume, certification, atomic payout and failure/retry cases against real Postgres. Soul, Typer, Arcade and Garden have substantial engines/components and bounded tests. Garden's native single/batched harvest through the built browser, service and SQL includes payout, retry and rendered reload; maturity is explicitly prepared test state. The `8a5dcebf` Garden read repair locally rejects unknown kinds and avoids first missing-data render poisoning; that is controlled-port evidence, not another service journey. | Typer/Arcade public API and host integration; shared wire-contract decisions; Typer's result-bound and now demonstrated deadline-arithmetic contract gaps (RP-162/RP-460); full platform fixture through public HTTP; Garden natural growth/long-session recovery and state-evolution clarification; authored content/mint, remaining whole-player journeys and accessibility. Pinball remains a draft. |
| Pet care / cosmetics | Controlled real-service Buy → adoption → equip/care → reload passes. Pet status announcements connect an actual stored care event through WebSocket to the mounted production client. Purchase/adoption completion respects newer keyboard focus; saved reads/remounts no longer repeat adoption speech. Native care now survives unrelated refresh work without moving the button or discarding consent (RP-440/RP-441), with duplicate/failure/unmount controls. | Pet/API schema reconciliation, shared attendance-clock repair (RP-365), adopted copy/content and consolidated review. These local repairs and passes do not establish general reliability or close the clock defect. |
| Clout / achievements | Achievement foundation, axis hooks and state/API exist. A built-client real-service fixture exercises native PR purchase, exact multiplier rendering, reload and the actual SQL head/event, with a failing producer-severing control. PR factors/progress now sit on the actual purchase cards. Native `?` help and once-per-transition re-attainment notices are locally verified through the real-service fixture; their text remains pending owner copy. | RP-307 numeric partition correctness, offline-episode meaning, projected-factor/hint definitions (RP-427), truthful role/harness bindings and measured/adopted content/balance. Current published-source hosted server CI still fails the 27 Clout partition cases. |
| Account / API / transport | Account security, backend deletion, revocation, exact-response and transport recovery witnesses were independently approved in earlier ranges. Generated API dispatch and newer consumers are partly implemented. Fifteen of epoch8's nineteen families now export complete closed descriptors; Economy has pricing/resource/generator/policy components but no full artifact export. Economy/Route admission repairs carry genuine versions through0.3.208. Credential responses now consistently refuse HTTP caching, locally verified against real Postgres; that new range awaits designated review. A recovery player-flow/storage/copy proposal is now drafted, not accepted or built. | Complete generated-client/route migration, startup failure and browser renewal contracts, visible recovery-code/recover-existing-account UX and long-session proof. |
| Privacy / player rights | Relation-complete 60-table inventory and bounded browser/event/backup/deletion traces expose actual retained data. Research publication classification/sanitization is complete. | Owner export/deletion/retention choices (D-008/D-009/D-015), implementations and player workflows. Public board survival after deletion and restoration from pre-deletion backups remain explicit issues. An instance backup is not player export. |
| Accessibility | Native keyboard, focus, pending feedback, reflow and announcement corrections exist across several surfaces; targeted browser checks execute. | Accepted cross-product task/assistive-technology matrix and actual task testing, including zoom/motion/longer workflows. Automated checks alone are not accessibility completion. |
| MMO and later game | Commons, guild, faction and doctrine primitives exist. Combat has shared Go/TS arithmetic, deterministic random streams, validated 36 temperament/runtime identities, strict ordered team/deck references and structural species-record decoding. These are standalone libraries with no live player consumer, not a full catalog or battle engines. | Combat catalog/effect definitions and Trust/Soul tables; complete world/feed/ticker/social surfaces, combat engines/integration and fair bot backfill, Tier 3–8 content, challenge/category set and all final endings. These are substantial unbuilt work, not merely closeout paperwork. |
| Deployment / operations | Accepted single-node Compose/Caddy floor; packaging, configuration, encrypted backup, release/rollback and operations tooling implemented with several bounded corrective approvals. Two complete independent local candidate builds at `102face6` match. Historical predecessor44a4a72 is also reconstructed twice with correct platform provenance, but fails current Compose/Caddy safety validation (RP-514). The Deployment plan/log and build records own evidence. | Remaining findings and exact-range reviews; establish a reviewed floor-compliant predecessor and full pair supply-chain proof, not retry the completed artifact reconstruction; R-006 clean Linux-host install/restore/rollback rehearsal. Local builds do not establish release readiness. |

### What the weeks of work have delivered

The working engine, persisted early-game loop and selected real-service player journeys are
substantial implementation progress. Tier-2 candidate construction, security/recovery witnesses,
backup restore checks and concrete UI/lifecycle corrections add useful capability and confidence.
Audit, research, review and record repairs are not additional playable content; their volume must
not be presented as equivalent progress toward the remaining tiers. This is still an early-game
development build, not a nearly finished 1.0 awaiting deployment.

Recent concrete gains include composed API request-ID/error-budget behavior, repeated real-DB
Guild fixtures, populated replay-archive preservation checks and live SQL Guild-input replay
through both runtimes with positive idle accrual. Rights probes now also show credential revival
from a pre-deletion encrypted backup under unchanged test keys and a valid-token clock. Those
findings still require policy/implementation, not a declaration that deletion is complete.

The next delivery boundaries are completing Garage's missing integrated capabilities; resolving
the specific numeric, clock, Headcount and public-wire contracts; adopting and validating content;
and building the world/combat/later-tier game. Rights, recovery, accessibility and real clean-host
operations remain required alongside that construction, not optional end-stage polish.

### Current verification and closeout

Original product-audit coordinate: `8a5dcebf`; at that coordinate clean main had75 commits after
local origin/main `4eaed0c9`. That remains a dated baseline, not today's publication count.
Fresh2026-10-15 read-only GitHub branch/run/job metadata queries confirm the same completed
failed CI38560135778 at `be66902b`. Its previously inspected failure logs retain their scope;
this metadata recheck is not another execution. Later local changes are not in it.
The agent has not pushed or treated publication as approval.
The whole-product table retains its stated baseline plus bounded additions, not a fresh
acceptance run for every feature. These are different evidence populations:

| Evidence | Result and limit |
|---|---|
| Latest published CI | [CI38560135778](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/38560135778) on `be66902b` completes failure. Browser/composed/schema/harness pass; finished failure logs identify client RP-131 at50a3a514 and server's27 RP-307 partition cases. Browser5m21s/composed6m07s/server8m11s exceed the five-minute floor. Later local commits are not in this run; no full-CI/latency/feature/release approval. CI38381138075 at4eaed0c9 remains prior evidence, not latest. |
| Previous completed published failures | [CI38222406320](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/38222406320) on `e0cb9187` failed historical kernel guard (RP-131), 27 Clout partition cases (RP-307), and Fiscal performance-census whitespace (RP-484), exceeding the blocking floor. The new run passes browser/performance but retains the guard/arithmetic failures. Earlier failures are not erased. |
| Recent repairs | Performance census, feedback/input, backup metadata/snapshot consistency and API admission/evidence corrections through `4eaed0c9` are now published, not independently approved by publication. Owning logs retain exact verification/review boundaries. The later hoard-preview guard is local and pending review/integration. Neither publication nor local passes establish whole-CI success. |
| Earlier hosted browser evidence | CI38381138075 at `4eaed0c9`, not the newer published run: all three functional projects complete426 files/34,704 passes/twelve expected deadlines/24 skips,305.18s. Five isolated Chromium performance cases pass with105 deselected and their declared exclusions. Browser6m38s/server6m40s exceed the five-minute floor. This dated census is retained evidence, not current-source acceptance or reliability proof; CI log owns scope. |
| Historical whole-job/environment audit | At `d618b0b4`, native client and Linux-amd64/Postgres server retain RP-131/RP-307; the later ordinary browser job has15 native-input failures in316.27s with all four suspect modules fetched. RP-525 is locally corrected; RP-526 remains open and the unverified serial config is withdrawn. Docker then stops responding at1.8GiB free. The subsequent environment and local-gate rows supersede availability, not that failure evidence; do not treat this old cleanup/diagnosis coordinate as today's blocker. [CI plan/log](scaffolding-and-ci/plan.md) retains exact results and limits. |
| Router configuration repair | RP-496 refuses either incomplete previous cursor-key pair. Old-source regression fails; corrected affected cold Go/vet, unchanged API generation and selected real Linux-amd64/Postgres composition checks pass. `36f81dae..2fbd246b` needs Claude review. |
| Subsequent player/API construction | Native queued Claim binds actual purchase/read delivery to refreshed revisions and consecutive SQL receipts. Public downloaded manifest/genesis/archive bytes now feed the actual TypeScript replay verifier while the real HTTP/SQL service is alive, with specific corruption refusals. Real Pitch processing claims block all three public Exit arms and release correctly; Go/TS Pitch/scaling admission repairs retain matching negatives. Owning logs hold executed scope and review ranges. These do not complete generated migration, held public minigame wires, natural pacing or a new gameplay milestone. |
| Original API reviews | Query `32560c4b`, epochs `e63796d5`, boards `4c316c4c`, routes `09d12987` and router `731bb3e7` now have complete original-range **CHANGES REQUIRED** dispositions. Do not restart those reviews or treat later repairs as approved. Later implementation ranges remain separate. |
| Kernel correction provenance | RP-131's non-verdict parser and RP-540's canonical citation grammar are locally repaired: actual CLI negatives refuse while completed findings and active/archive records remain admissible. A controlled old-tool exterior-file probe stays inside owned temporary storage, not host-private files. The existing DP-E correction cites Claude's completed verdict, with Codex only recorder. Real root verification still rejects `50a3a514`; no exception proposal is adopted or history rewritten. Corrective independent review remains; CI log owns evidence. |
| Last full local production/history checks | At the recorded `a713137e` coordinate, cold full production tests still fail27 `TestAxisAccrualPartitionProperty` cases (RP-307). Root `make verify-kernel-version` passes its checkout controls, then rejects historical `50a3a514` (RP-131). No assertion, bound, history row or version was altered. These are scoped local failures, not a new hosted/full-CI run. |
| Last whole native client verification | At `b8eca33c`, the full Node population passes but the full Chromium/WebKit functional run hits Typer's unchanged15s whole-game deadline (RP-526). Failure-only phase observation is now locally verified; one instrumented full population passes afterward without any input/timeout/concurrency change. The original cause and reliability remain open, not repaired by that pass. Other client guards pass while the actual history gate still fails50a3a514. [CI log](scaffolding-and-ci/log.md) owns both executions and scope; no Firefox/Linux/hosted/whole-CI or release promotion. |
| Original Clout UI review | P5 `2f2c5a04^..2f2c5a04` now has all30-path CHANGES REQUIRED for progress naming, fractional-factor display, purchase-card binding and contribution uniqueness. Its actual historical decoder misses two required refusals; current scoped producer/schema/Node/native checks pass with the SQL exclusion explicit. Later Codex fixes, real-service supplements and full CV9/AC11 remain separately pending review/acceptance. The Clout log owns the verdict. |
| Original Clout Gaia-law review | P4 `4c089f00^..4c089f00`, both paths, remains CHANGES REQUIRED (RP-537). Its local test-only repair now covers escaped write forms, checked codec copies and actual nonzero encode/restore controls. This is a conservative source guard, not a demonstrated current mint or full AC4/reward-union proof. Corrective independent review and remaining acceptance stay open; the Clout log owns executed evidence. |
| Original Clout loader review / source ownership | P1 `ace4e67e^..ace4e67e`, all17 paths, is CHANGES REQUIRED (RP-304/538). Original fixture/corpus reproduce byte-identically in memory. Both current loaders now locally refuse repeated contribution-source ownership; equal factors with distinct sources remain legal. Affected checks pass at kernel0.3.187, but broader production still fails27 partitions and the historical version guard is red. The new correction needs Claude review; full Clout acceptance remains. Clout log owns the ranges; no mint or release promotion. |
| Original Clout formula/save review | P2/P3 `f256b235^..f256b235`, all28 paths, is CHANGES REQUIRED, retaining arithmetic, replay/receipt and author-contract gaps. RP-539 additionally reproduces silent pre-v19 TypeScript serialization omissions; the local guard correction preserves healthy bytes with Go/browser/replay controls at kernel0.3.188. Original P1–P5 implementation dispositions are now complete, not approval of later corrections, P6 or full acceptance. SQL remains unverified from the interrupted Docker attempt; Clout log owns results and review boundaries. |
| Original Garage review / Fiscal safety | Original GS0.2 `66ae01f9`, GS0.1 `5165abcf` and surfaces `acf43132` now have full-path CHANGES REQUIRED dispositions, distinct from later Codex repairs. Surface review retains pending/readiness, exact refusal, semantic-layout and error-containment findings; catalog ownership correspondence remains RP-534. Current scoped Node103/native28 pass, not historical or full-feature approval. Actual loader-admitted hoard inputs expose int64 wrap to448,384; the new local check refuses before multiplication and preserves exact/capped/zero values. Go/client/native checks pass, with17 SQL skips explicit. Kernel-owned derivation/larger-value representation (RP-533) and actual DB/player/release proof remain open. The Garage log owns the ranges. |
| Local execution environment | Subsequent2026-10-14 Docker ping returns HTTP200/OK in0.003226s. Declared Linux-amd64/Postgres backup and Linux filesystem/second-process cleanup populations now execute and pass; the Deployment log owns exact scope. Test-owned containers/network/tmpfs are removed and the project is empty. No daemon restart or unrelated cleanup was performed. Other lanes require their own runs; earlier ping/capacity observations remain dated evidence. |
| Subsequent local release-lane verification | Declared Caddy/Postgres/WebSocket lane passes before and after RP-612's smoke-response completeness repair, with readiness/intent/admitted-work/courtesy/4003/restart and restored-account checks. The test composition uses a checkout mount and the same current code after restore; it is not a governed previous-image rollback or exact-artifact R-006. Engine28.4.0/Compose2.39.4-desktop.1 matches the recorded development pair, but the daemon is Linux/aarch64. Native-amd64 CR2/CR4 experiments and review remain open. Deployment log owns commands and the range after7644aed4. |
| Release artifact | No current-candidate clean-host R-006 install/restore/rollback proof. The executed local Caddy/Postgres release fixture is useful bounded evidence, not that rehearsal or approval of the complete Deployment range. |
| Current-source player/database integration | At `654523c7` (kernel0.3.216), full `make test-game-ui-composed` passes with main/endings/recovery, Cosmetic/Reputation, Clout, payment negatives, Garden and Tier-2 candidate journeys. Both persisted populations pass8/8 refresh leaves and ten required parents, with no captured skipped/failed test headers. The new real Guild/SQL capture additionally reproduces complete state/receipt/events in Go/TS, with positive idle/lifetime and zero-producer/omission/duplicate controls. [Clout log](clout-v1-and-pr-interns/log.md#2026-10-15--r-012-positive-idle-producer-in-the-live-guild-capture) owns this latest aggregate and its controlled setup; the earlier Garage result remains historical. Not natural progression, Clout-v19 conservation, Headcount, Arcade/Typer public integration, AT, hosted CI, independent approval or exact-release proof. |
| Local Linux browser gate / retained failures | The later `b45dcd1d` test batch completes broad/native/performance phases on Linux/ARM64 after repairing RP-645's contradictory timeout-fixture readiness assumption. [CI log](scaffolding-and-ci/log.md#2026-10-15--rp-645-deadline-populations-and-rp-637-cancellation-evidence) owns the failed/final runs and local reports. RP-637's Firefox Tier2 Codex failures remain unattributed and unfixed; passing later runs do not close them. No hosted AMD64, reliability, whole-CI or release approval follows. |
| Latest API behavior and archive tests | Request IDs/logging now reach both composed surfaces and drain refusals; identity-generation errors are typed, public errors spend the existing IP budget, and gzip500 alternatives survive binding. Those corrective ranges await Claude review; [API log](api-foundation/log.md) owns executed scopes. Guild fixtures pass repeated/shuffled Postgres runs (RP-121), and the existing archive test checks populated content, exact integers, digest, ordering, rollback/retry and11 corruption/omission controls (RP-124). These are conformance/proof repairs, not new endpoints, acceptance or another tier. |
| Latest rights observations | Actual Pitch/terminal Soul producers retain selected data after deletion while access is revoked; active Pitch adds a real nonterminal-command observation. Shared Guild histories retain nested deleted-consumer identifiers. HTTP deletion→encrypted backup restore additionally revives original access/refresh/recovery credentials from B_pre, while B_post preserves denial and a bystander stays healthy, under unchanged test keys/valid-token clock. [Rights inventory](platform-alignment/data-rights-inventory.md) and the [restore dossier](platform-alignment/data-rights-restore-resurrection.md) own limits. These expose policy/implementation gaps; they are not erasure acceptance, expiry/rotation, legal review or clean-host proof. |
| Current Linux API consumers | The declared Linux-amd64/Postgres publicapi/publicread/account/gameserver wave first fails the literal minigame lifecycle, then its focused branch and one same-context diagnostic pass unchanged. The latter retains complete output and one explicit separate-runtime skip. [Minigame API log](minigame-api-and-surface/log.md#2026-10-14--rp-614-linux-api-consumer-wave-fails-focused-branch-passes) owns RP-614, including the agent's initial assertion-capture failure. No source fix, cause/reliability closure, full server-job or hosted CI success is inferred. |
| Current full local server gate | At `3bde51b5`, the declared Linux-amd64/Postgres `verify-server-ci` exits2: vet passes,64 test packages pass, Production fails exactly27 RP-307 partition cases. Generation/boundary leaves after the failed test target do not execute. The [CI log](scaffolding-and-ci/log.md#2026-10-15--current-full-linux-server-gate-remains-red) owns scope and cleanup. This does not close RP-614 reliability, establish a skip census, or prove native-amd64 hosted timing/full-CI readiness. |
| Private operations review | Original Claude R5 `78ec39a7^..78ec39a7` now has bounded Codex approval after actual private scrapes, cleanup-rule firing, nonce delivery, receiver rejection and Caddy admin isolation. The fixture is mixed-architecture, not clean-host release evidence. Exact-index amd64 rules and affected Linux-amd64/Postgres startup/job checks pass; the new focused Make extraction after `ffc8bc25` separately needs Claude. Full DP-E/AC7 and aggregate/hosted CI remain open. |
| Subsequent operations construction | Actual locked SQL collection, connection refusal/reconnection, relay dead-letter growth and bounded filesystem allocation reach the private receiver with matching resolution. Same-second backup failure, startup-loss cleanup observation and ordinary drain cancellation have reproduced corrections. A later test supplement takes actual Docker pre-scrape restarts through the production observer and private services to receiver firing; actual replacement count0 resolves the same alert. This is a scripted process, not gameserver/release-host proof; the Deployment log owns that supplement's range. These Codex batches need consolidated designated review; actual host journald/retention, complete AC7 and exact-artifact R-006 remain open. |
| Latest local safety and proof work | Account JWT members, physical-package attribution, archive-secret scanning, live-backup cleanup ownership/identity, backup timestamp/endpoint binding, observer URL handling, real drain-method protocol coverage and rehearsal chronology/mutation/refusal checks are locally corrected or supplemented. The [Deployment log](deployment-foundation/log.md) and [Account log](account-and-session-bootstrap/log.md) own exact ranges and limitations. Full original Deployment R3/R6/R8/R9/R10/R12/R18/R20 reviews have CHANGES REQUIRED dispositions; they do not approve Codex's descendants. The current declared Linux/Postgres lane now passes RP-551's empty/populated database-query override restores and the other five integration functions. These changes add no tier, adopted content, clean-host proof or release approval. |
| Subsequent catalog/UI construction | Fifteen complete family exports are present and locally tested, not full C18/API acceptance. Economy's version/row/role/policy and shared Route admission corrections preserve current/stored artifacts and existing semantic checks. The latest Garden read repair passes its affected mounted native population and build/type guards. [API log](api-foundation/log.md) and [Garden log](minigame-server-garden/log.md) own exact commands/ranges/limits. Complete catalog serving, four family exports and designated review remain. |
| Latest maintenance evidence | [Run38507071157](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/38507071157), still on `4eaed0c9`, completes failure: numeric maintenance passes, career/relevance report comparisons fail, and the complete harness observation exits124. Its standard-pacing phase completes300/300 in2,711,416ms; relevance row0 stops at85/107, and row1 never starts. This is an incomplete measurement, not harness acceptance or authority to increase the50-minute budget. The CI log owns the artifact/hash and executed validator refusal. |

The [API log](api-foundation/log.md), [Garage log](garage-player-surfaces/log.md),
[Minigame API log](minigame-api-and-surface/log.md), [Account log](account-and-session-bootstrap/log.md)
and [Deployment log](deployment-foundation/log.md) own exact ranges, commands and exclusions.
This editorial overview does not rerun software acceptance, accept drafts, waive review debt,
authorize publication or promote a release.

### Decisions versus engineering work

Technical amendments need concrete proposals, normative review and adoption; they are not
requests for Marco to debug the implementation. The currently highlighted engineering drafts are:

- [Typed API dictionaries](../rfc/api-typed-dictionaries.md): preserve existing saved receipt
  bytes and describe their keyed values, versus design a separately versioned HTTP projection.
- [Bodiless API responses](../rfc/api-bodiless-responses.md): explicitly model existing204/304
  responses and conditional reads, versus separately ruled transport contracts.
- [Integer token encoding](../rfc/api-integer-token-encoding.md): describe Relevance's existing
  exact integral decimal/exponent tokens without widening all other integer fields.
- [Artifact-schema ownership](../rfc/api-artifact-schema-ownership.md): one Route-owned schema
  companion preserves the pure evaluator boundary; exact independently reviewed descriptor-only
  delivery avoids false gameplay-version signals. No companion or exemption is yet authorized.
- [Historical kernel guard](../rfc/kernel-history-guard-integrity.md): allow narrowly reviewed,
  append-only path exceptions for proved behavior-identical published changes. Individual
  historical candidates still need executed evidence and independent review.
- [Accrual conservation](../rfc/production-accrual-conservation.md): define absence episodes,
  settlement and origin lifetime across committed saves. Approval cannot repair the arithmetic.
- [Backup metadata integrity](../rfc/backup-protection-metadata.md): adopt the dedicated signing
  key/trust boundary and explicit authenticated adoption of legacy backups, rather than trust
  editable checksum-only metadata.

These seven need explicit adoption and normative review, not another generic permission to
work. The accrual draft also needs author reconciliation; its arithmetic, persistence,
compatibility and full-state measurements remain engineering obligations. Approval of a
direction cannot substitute for a viable implementation or make failing tests disappear.
The separate cursor policy/rotation-ledger contradiction (RP-495) belongs to the API and
Deployment authors; it is not resolved by the RP-496 configuration repair.

These holds apply to their affected work, not the whole repository. Continue accepted
construction, reproduced conformance repairs and remaining independent implementation reviews.
Recovery/renewal, rights, accessibility, content and later-game contracts have their own
unresolved boundaries; these proposals are not an exhaustive release decision packet.

Actual gameplay choices include Headcount's hiring/output pair and Clout's projected-factor/
hint lifetime (RP-414/RP-427), plus the held Manual input-delivery behavior. Actual release-policy
choices include export/deletion/retention (D-008/D-009/D-015), the task/assistive-technology matrix
(D-018), per-release restore proof (D-019), exact content/copy adoption and the final release call.
The [decision queue](platform-alignment/decision-queue.md) and its linked option sheets own
those choices. Typer wire/arithmetic, sibling-body contradictions and unimplemented integration
are engineering/author closure; do not relabel them as a need for generic owner approval.

### Remaining program, in practical groups

1. Complete Garage's player-facing integration and content: generated callers, minigame
   boundaries, recovery, natural growth/pacing, Headcount and shared clocks. Prepared cash or
   maturity is integration evidence, not a balanced release economy.
2. Repair the known numeric and CI failures honestly; consolidate corrective review ranges
   while continuing safe work under accepted contracts. Later API implementation reviews,
   including verification/download and generated consumers, remain work, not owner decisions.
3. Implement and exercise player export/deletion/recovery and task accessibility under
   adopted contracts. Instance backup and automated keyboard checks are not substitutes.
4. Finish Deployment findings/reviews; retain the new local candidate. The reconstructed
   historical predecessor fails the current safety floor (RP-514); establish a reviewed
   floor-compliant predecessor and full pair proof, then prove installation,
   restore, rollback and operations on an authorized supported clean host.
5. Build the substantial remaining world/social/event/combat/bot systems and Tier 3–8
   progression, challenges and all final endings. Follow the cumulative
   [milestone roadmap](roadmap-1.0.md#milestone-path); a preview never narrows 1.0.

New privacy/verification hold (2026-10-14): [RP-606 and the newer-content map](platform-alignment/data-rights-current-content.md)
trace hidden Garden salt into current replay carry and a controlled original-byte public
download. No deployed/SQL leak is proved. An authored resolution must preserve both hidden
information and independent verification before Garden publication/full acceptance.

## Historical checkpoints

Everything below preserves earlier evidence and then-current claims/next actions. In
particular, a historical entry saying “current” or “latest” is not a live instruction.
Use the overview above and owning plans/logs for current work. No historical result is deleted.

### Historical verification and closeout

- Current published source is `e0cb9187`; subsequent local repairs are unpushed. Hosted
  [CI38222406320](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/38222406320)
  (2026-10-12) passes schema/harness/composed, fails client/server/browser, and exceeds
  the five-minute blocking floor. Client fails the RP-131 historical version guard;
  server retains the 27 RP-307 accrual partition failures. All functional browser tests
  pass, but the isolated populated performance census fails on Fiscal layout whitespace.
  RP-484 repairs that census locally, without changing budgets or the population; the
  exact performance target, Fiscal edge checks and types pass. Its range after `e0cb9187`
  awaits designated review and has no hosted verdict. A subsequent test-only batch
  adds pre-adoption/adopted Desk performance populations, with 210 Chromium/WebKit
  functional checks passing and a missing-axis control caught. The latest catalog-width
  supplement includes all nine pinned generators, twelve upgrades and both resources:
  five isolated cases, the actual Go shared-projection check and types pass; final-input
  generator/resource losses are caught. This remains partial AC7 coverage, not a later-
  catalog/default-player/manual profile or a hosted verdict.
  The [Garage log](garage-player-surfaces/log.md)
  owns evidence. Earlier dated CI/local checkpoints below are historical, not current runs.
- Current finished source is `1b484c50` (2026-10-09). The latest Deployment corrections
  require healthy baselines before interrupted-backup, public-metrics and source/image-secret
  probes can count a failure as evidence (RP-481/RP-482). Old-behavior regressions fail;
  affected cold native Go/vet and declared Linux-amd64 packages pass with zero skips.
  The coherent range `1efea9cf..1b484c50` awaits designated review. These are release-check
  repairs, not new gameplay, runtime isolation or R-006 clean-host proof. The
  [Deployment log](deployment-foundation/log.md) owns commands, results and limitations.
- Read-only GitHub inspection at this closeout confirms latest published CI
  [37795560514](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/37795560514)
  at `cbd2069f` still fails server/client/browser; schema/harness/composed pass.
  There is no hosted verdict for the newer local source. The implementation checkpoint
  is committed and clean; this editorial reconciliation does not rerun software tests,
  adopt pending proposals, approve corrective ranges or authorize a push/release.
- Finished bounded checkpoint (2026-10-09): Recovery now preserves a known terminal result
  when a pending Finish fails, and clears superseded notices (RP-480). Old behavior fails
  eight native cases; corrected affected60 browser/84 unit checks, types/build/boundaries pass.
  The original Claude Recovery batch has a complete CHANGES REQUIRED disposition, not approval
  of later fixes. Ticker's retained TL4 counterexample (RP-479) also blocks an honest selector
  implementation alongside the existing Feed hold. These are correctness/review/research
  progress, not new playable tiers. The Minigame API and Ticker logs own evidence and review
  boundaries; the latest recorded hosted verdict below remains red, with no current-HEAD run.
- Preceding Pitch checkpoint `a4c09f31`: the bounded native Enter/Space/Tab correction is locally
  verified, pending designated review. The complete original Claude Pitch surface range now
  has a Codex CHANGES REQUIRED disposition, not retroactive approval of its later repairs.
  Fresh affected component/unit/copy checks pass; the historical kernel gate still fails.
  The [Minigame API log](minigame-api-and-surface/log.md) owns exact coverage and limits.
  This closes original-batch review work, not Garage acceptance, another tier or whole CI.
- Latest bounded API correction (2026-10-08, RP-476): the generated client now chooses
  the declared media representation when a status has alternatives. Both old-client
  regressions fail; affected Go/vet,160 Node and346 native Chromium/WebKit checks,
  types/build/boundaries and API regeneration pass. Main journey33328 also passes
  against the built production client, real server/Postgres/WebSocket, including both
  early endings, next Company and process-death recovery. This is not the full composed
  aggregate or hosted CI. The [API log](api-foundation/log.md) owns review handoff and
  related locally corrected OpenAPI header finding RP-477. Preceding Typer correction
  `59e5cc49` resets composition state when an authoritative terminal removes the input;
  old-behavior browser failures and affected browser/real-Postgres passes are in the
  [Typer log](minigame-terminal-typer/log.md). Both corrections await designated review;
  neither adds a playable tier or closes the parent feature.
- Latest finished source checkpoint `da21dae2` (2026-10-08): API policy admission now
  refuses missing/null members and case aliases instead of silently accepting an altered
  startup policy (`e4fd61fa`). Raw WebSocket test peers now answer protocol pings, preventing
  a no-pong disconnect from invalidating the account-revocation witness (`da21dae2`, test-only).
  Both regressions fail on the old behavior. Affected cold native tests/vet, unchanged API
  generation and the declared Linux-amd64/real-Postgres publicapi/publicread/gameserver
  population pass; the latter executes revocation and drain, first-hour replay, public evidence,
  proxy, Garden and Minigame/Soul cases with zero skips. The earlier Linux failure remains
  recorded, and RP-450's distinct hosted drain failure is not declared resolved. Exact ranges
  `53f3f663..e4fd61fa` and `e4fd61fa..da21dae2` await designated review; the
  [API log](api-foundation/log.md) and [Account log](account-and-session-bootstrap/log.md)
  own detailed evidence. Source is committed and the worktree is clean, with no live check
  handles or push. This is correctness/test-fixture work, not new gameplay, whole-CI green,
  acceptance, archival or release readiness. The latest recorded hosted verdict below remains red.
- Previous Combat source checkpoint `9f518477` (2026-10-08): C2's structural team/deck
  reference decoder is committed. Cold complete affected Go, vet, both division boundaries,
  88 Node checks, strict types and 176 native Chromium/WebKit checks pass; the
  [Combat log](combat-shared-data/log.md) owns the exact evidence and pending designated
  range `1e8a25bc..9f518477`. No catalog, battle engine, live integration, Linux/Firefox or
  whole-CI claim follows. Existing five-minute Recovery completion already has a recorded
  real-browser/service/Postgres pass; it is not new or unfinished work to recreate.
  Original Pet, Garden and Arcade reviews have complete dispositions; later corrective
  ranges remain independently reviewable. Fresh read-only GitHub inspection still finds
  run37795560514 at `cbd2069f` red (schema/harness/composed pass; server/client/browser fail).
  The source checkpoint is clean and 45 commits ahead of the locally recorded origin/main,
  with no current-head hosted verdict. This overview update is editorial closeout, not an
  accepted feature, content mint, release, push or another player milestone.
- Arcade's original A1–A7 batches now all have complete designated CHANGES REQUIRED
  dispositions, not implicit approval of later repairs. At source `281340d8`, selected
  Linux-amd64/real-Postgres engine/start/play/rollback/Exit populations and 363 client
  checks pass. Mounted Chromium/WebKit controls pass 38 checks; native Firefox fails
  to connect, so that aggregate is not green. The [Arcade log](minigame-demo-disc-arcade/log.md)
  owns exact ranges and corrective-review debt. Public Typer/Arcade wire and host,
  author/copy reconciliation and production content remain open; this closes original
  review work, not a playable tier or feature acceptance. Read-only GitHub inspection
  still confirms run 37795560514 at published `cbd2069f` is red: schema/harness/composed
  pass, server/client/browser fail. Later local checks are not a hosted verdict.
- Garden original G1–G7 now all have complete designated CHANGES REQUIRED dispositions;
  the remaining original reviews are finished, not implicitly approved. Current-source cold
  Garden/save/replaycatalog/kernel and minigame checks,155 client tests, types/copy/corpus/vet
  and selected real Linux-amd64/Postgres activation/Exit/rollback/history checks pass. Full
  production80654 remains red on the27 Clout partition cases plus stale policy provenance;
  `a31621ee` re-observes all44 policy pairs, changing only that digest and preserving the
  negative results. The underlying arithmetic defect remains. No whole-CI or new tier claim.
  The new SG8 event-admission repair after `c67f22e5` passes actual Postgres rollback and
  producer/integration checks, pending independent review. RP-471 additionally records an
  above-cap contradiction in AC6 for its author; no gameplay policy was changed. The
  [Garden log](minigame-server-garden/log.md) owns exact dispositions and corrective review
  handoffs; original-batch reviews must not restart. Hosted run37795560514 was rechecked
  read-only: schema/harness/composed pass, server/client/browser fail at published `cbd2069f`.
- Latest bounded source repair (2026-10-08, after `ebbb465c`): Garden no longer admits null
  row/column as plot zero. Old authenticated Linux/Postgres HTTP actually removed a plant
  for a null-row uproot. Final real HTTP now rejects all six command/field combinations
  without changing either saved stream; the72-case decoder matrix,75 affected client
  tests, types/vet/API/corpus and both shared numeric suites pass. The
  [Garden log](minigame-server-garden/log.md) owns the pending designated range and limits.
  This is a command-admission repair, not completed Garden acceptance, a new tier or whole-CI.
- Previous finished source checkpoint (`aca464ac`): Garden rereads at the earlier substrate-lock
  expiry or growth tick, so Mainframe's ten-minute lock does not remain visibly active until
  its fifteen-minute growth tick. The server projection, old-behavior browser failures,
  corrected mounted cases and built-client/real-service/Postgres harvest journey are recorded
  in the [Garden log](minigame-server-garden/log.md). The journey uses prepared maturity,
  not natural fifteen-minute growth. That production correction is `a257faaa..c54c6846`;
  designated review remains pending.
- The subsequent test-only range `c54c6846..aca464ac` waits for actual rendered startup
  readiness before connection assertions instead of assuming five event-loop turns suffice.
  A deliberately held initial read reproduces the old failure. Final affected native checks
  pass; the declared Linux Chromium/Firefox/WebKit four-file population passes 270 checks,
  zero skips. This does not close every intermittent-input report or establish whole CI.
  Both ranges are committed, with no live verification process or uncommitted source work.
  Original G5–G7 now have complete designated CHANGES REQUIRED dispositions, not approval of
  corrected descendants. Original G1–G4 coverage, corrective-range review, author/content
  holds and natural growth/renewal remain open. The latest owning log records the full G6/G7
  inspection and current cold real-Postgres/HTTP/schema/copy checks separately from the verdicts.
- Earlier finished implementation checkpoint (`e15e0ff5`): Recovery's built-client journey
  now resumes the same committed session after actual gameserver process death and a lost
  reply, with unchanged persisted progress before reconnect and no credit for the missed
  interval. The retained data-loss control fails at the new assertion. Fiscal's keyboard
  driver now waits for harvest eligibility before focusing and pressing Enter; the native
  reproduction and final affected real-service journey pass. These are committed, pending
  designated review, not proof of DB/host disaster recovery or whole-CI reliability.
  The [Minigame API log](minigame-api-and-surface/log.md) and
  [Garage log](garage-player-surfaces/log.md) own the exact ranges and evidence.
- Garden's complete original G5 range `f8d170a8^..f8d170a8` now has a designated Codex
  **CHANGES REQUIRED** disposition, preserving its clock/test/specification findings rather
  than implicitly approving later corrections. Current-source cold Linux-amd64/Postgres
  Garden and shared-coordinator consumers pass; all 36 client replay tests pass. Natural
  growth remains held by browser renewal; author reconciliation, corrective review and
  launch content remain open. The [Garden log](minigame-server-garden/log.md) owns this result.
- Previous backup safety checkpoint (2026-10-08, `9265049b`): concurrent encrypted
  backups can no longer replace an already completed backup with the same ID. The
  old-source regression reproduced actual overwritten envelopes; atomic no-replace
  commit now preserves the winning bytes and refuses the losing writer. Cold native
  and Linux-amd64 checks pass. Separately, `make test-deployment-backup` executes all
  six real Postgres16/encrypted dump-and-restore integration functions, zero skips.
  The [Deployment log](deployment-foundation/log.md) owns the pending designated
  range `d0fc603a..9265049b`. This is backup safety, not clean-host release recovery
  or closure of the pre-deletion-backup privacy issue.
- Recovery checkpoint `d0fc603a`: the existing watchdog terminal is now admitted
  by the generated API contract and consumed without a progress-rendering crash or
  duplicate host notification. Old HTTP/Postgres and mounted-browser regressions fail;
  affected cold Go, Linux-amd64/Postgres, Chromium/WebKit, full client, types and API
  checks pass, followed by the final built-client/real-service main journey. The
  [Minigame API log](minigame-api-and-surface/log.md) owns the pending designated
  range `d5b40996..d0fc603a` and its limits: injected server expiry/controlled browser
  replies, not a natural 24-hour browser observation or whole-feature acceptance.
- Published CI rechecked read-only at this checkpoint: latest run
  [37795560514](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/37795560514)
  at `cbd2069f` remains red. Schema, harness and composed player journey pass;
  client, server and browser jobs fail. Newer local repairs are not a hosted verdict.
  Source work is committed; no push, archival or release approval is implied.
- Preceding source checkpoint (2026-10-08, `2b33fb2c`): the mounted Recovery
  journey now drops a real heartbeat reply after the server commits it, verifies visible
  pause/no queued beats, then uses native Reconnect to rotate the same session's token.
  Missed time earns no progress; the next fresh interval advances the rendered state.
  The final affected built-client/real-service/Postgres main journey passes, including
  required backend populations, Pitch process-death recovery and ordinary Recovery cancel.
  Unaffected sibling journeys and five-minute completion were not rerun for this tool-only
  change. The [Minigame API log](minigame-api-and-surface/log.md) owns the pending review
  range `a1197552..2b33fb2c`; no whole-CI or whole-feature acceptance follows.
- Pet checkpoint finished at `a1197552`: all original P1–P7 batches now have full
  designated dispositions, retaining CHANGES REQUIRED for their specific defects rather
  than approving later corrections by implication. RP-463's saved identity case-alias
  repair at `ff5275fc` passes affected cold Go/client checks and actual Linux/Postgres
  adoption/Exit cases. Full client passes 10,321 with 977 explicit exclusions. The actual
  historical kernel-version gate remains red; PA7/C2, shared attendance, adopted content
  and corrective-range review remain open. Exact evidence lives in the
  [Pet log](pet-adoption-v1/log.md). These are review/correctness results, not another tier.
- Finished-work checkpoint (2026-10-08, source `852a2b7f`): the browser crash-recovery
  witness is committed at `75f5830b`. The whole composed lane passed before the final
  saved-snapshot/fault-wait tightening; the affected main journey then passed on the
  final bytes against the real service/Postgres. It proves same-command retry after
  a committed reply is lost during server process death, not DB/host disaster recovery.
  Its owning [Minigame API log](minigame-api-and-surface/log.md) records the pending
  designated range. The [typed-dictionary proposal](../rfc/api-typed-dictionaries.md)
  is also committed, but remains draft: RP-391 blocks honest generated-client migration
  of ordinary intent receipts. Owner choice and normative review remain pending;
  drafting the prerequisite is not implementing or completing that migration.
  Read-only GitHub inspection again confirms published CI `37795560514` at `cbd2069f`
  is red: schema/harness/composed pass; client/server/browser fail. Later local repairs
  and their pending reviews do not change that published verdict.
- Latest checkpoint (2026-10-08): original Typer B3/B4 reviews are complete as
  CHANGES REQUIRED; RP-461's raw unlock admission is locally corrected at `1fe602e2`.
  Affected Node/native Chromium-WebKit, cold Go/vet and type/build checks pass.
  Broader client verification exposed RP-462's stale producer source identity;
  actual64-profile re-observation preserves every result and changes only that
  source digest. Full client now passes10319 with977 explicit skips; selected
  Linux-amd64/Postgres producer/start tests also pass. Owning Typer/Clout logs
  retain exact evidence and pending corrective review. This is not whole CI,
  public Typer integration, a numeric-contract ruling or feature acceptance.
- Latest Cosmetic review checkpoint (2026-10-08): all original C1–C8 diffs now have
  recorded dispositions, not blanket approval of corrected descendants. C5's
  newly reproduced RP-458 is locally corrected: unowned unavailable items require
  a lock. Two old-client failures become green; affected209 client checks, types/
  build/boundaries, cold Go projector/account and byte-identical API generation
  pass. Exact ranges and limits are in the [Cosmetic log](cosmetic-shop-v1/log.md).
  Corrective review, author-body, shared-clock/content and full acceptance remain.
- Latest saved-state checkpoint (2026-10-08): complete original Cosmetic C1/C3
  reviews join C4/C6/C7/C8; their CHANGES REQUIRED verdicts do not approve Codex
  corrections. RP-457's ten Go case-alias acceptance failures are locally corrected,
  with valid encoding controls preserved. Affected cold Go/vet,120 client tests,
  types and declared Linux-amd64/real-Postgres Cosmetic/Exit/isolation checks pass.
  Canonical saved bytes/content are unchanged; kernel179→180 is an admission fix,
  not historical-guard repair. The [Cosmetic log](cosmetic-shop-v1/log.md) owns the
  exact review ranges. C5 disposition, author/body/content/shared-clock and whole
  feature acceptance remain open. This finishes a bounded review/repair checkpoint,
  not another tier, hosted CI success or progress toward final endings.
- Latest bounded closeout (2026-10-08): Cosmetic event-key admission RP-453 is locally
  corrected at `377279c4`, with cold Go/TS and actual Linux/Postgres checks. Original Claude
  C4 review is complete at `f097df7e`; the complete original C6/C7/C8 reviews are also finished
  in the latest Cosmetic log. Each retains CHANGES REQUIRED for its recorded defects;
  Codex corrections still require Claude. Completion-focus RP-454/RP-455 at `c802c2dc`
  passes affected native checks and the full real-service composed1397. The subsequent
  announcement-only RP-456 correction removes duplicate refusal delivery; six old-host
  native failures become green, including unchanged revision-conflict controls and refusal
  retention after shelf removal. Finished affected Chromium/WebKit280, types/build/boundaries
  pass; four performance-only exclusions are separate. That last correction does not change
  gameplay, persistence or callbacks, and the real-service journey was not rerun for it.
  The [Cosmetic log](cosmetic-shop-v1/log.md) owns exact review ranges and intermediate results.
  These are bounded repairs, not complete C6/feature acceptance, AT, hosted CI or a new tier.
  Latest inspected published CI remains red at `cbd2069f`; newer work is local, not pushed.
- Preceding source batch: `43afbbd1` (2026-10-08), Recovery RP-452. A terminal
  reply arriving after its surface unmounts no longer notifies the old host or
  moves focus. Native Chromium/WebKit regressions fail on the old behavior;
  corrected affected browser26/unit75 checks, types/build/boundaries and the
  existing full real-service composed target pass. The held-response regression
  uses a port double; the composed journey separately exercises ordinary real
  cancellation, not that delayed-response condition. The
  [Minigame API log](minigame-api-and-surface/log.md) owns the exact pending
  designated-review range `0e465815..43afbbd1`. This closes the local repair,
  not whole-feature acceptance. Source work is committed; no push or archival.
  Read-only CI recheck still finds the published red run37795560514 below.
- Batch wrap-up (2026-10-08): the account drain witness at `b0f26d47` now keeps
  its healthy control peer reading and requires the actual server-drain close code.
  The complete declared Linux/amd64/Postgres gameserver population passes without
  skips; the [Account log](account-and-session-bootstrap/log.md) owns the evidence
  and pending designated review. The original hosted timeout remains unexplained
  (RP-450), not closed by these local passes. A fresh focused native Firefox run
  fails to establish its browser session and executes no tests; RP-451 remains
  open. Latest published CI was rechecked read-only and remains the red run below.
  The source batches are committed, not accepted as whole features or released.
- Latest API correction (2026-10-08): RP-449 is locally corrected by separating strict
  registry HTTP admission from shared artifact descriptor validation. Unchanged Go/TypeScript
  formula/replay controls and actual Postgres request/catalog/formula populations pass;
  generated contracts/artifacts remain byte-identical. The [API log](api-foundation/log.md)
  owns the pending designated-review range. This does not repair the other CI failures or
  make the earlier published run green.
- Latest Garage checkpoint (2026-10-08): RP-448's analogous Founder handoff is locally
  corrected under GS0.2/GS1-A6. The shared dispatcher compares the actual read's Founder
  revision with the applied receipt, including across Company changes; Company revisions
  remain run-scoped. Native precommit tests fail the old host in both browsers. Current-read,
  across-Company, failed-fresh-read/no-retry and prior Company controls pass. Final affected
  Chromium/WebKit12631 passes1080 cases with four separate performance-only exclusions;
  types/build/boundaries38428 pass. Final real-service `make test-game-ui-composed`49119
  exits0 through main/Recovery, pets/Reputation, Clout, payment negatives, Garden and Tier2.
  The [Garage log](garage-player-surfaces/log.md) owns the pending designated-review range.
  At that checkpoint separate cold formulas/replaycatalog tests were RED (RP-449);
  the later API correction above repairs that local regression. Latest hosted CI below
  remains RED. This bounded correction is not whole-CI, full Garage or release acceptance.
- Latest Garage checkpoint (2026-10-08): RP-447's stale purchase handoff is locally corrected.
  Every Company action checks whether a coalesced read predates its applied revision and,
  if so, completes a post-response refresh before clearing pending. Founder/run identity
  keeps successor-Company genesis separate; no intent retry or driver assertion change.
  Old-host purchase regressions fail in Chromium/WebKit; the affected native1074 cases
  pass with four separate performance-only exclusions, and types/build/boundaries pass.
  Final whole `make test-game-ui-composed`93889 exits0, including native Recovery, pets,
  Reputation, Clout, payment negatives, Garden and Tier2 with exact SQL/history/reload checks.
  Earlier whole89786's failure below is preserved, not a claimed pass. The [Garage log](garage-player-surfaces/log.md)
  owns the exact pending designated-review range. This is local composed-CI evidence, not
  hosted/full-CI, Firefox, full feature acceptance, content mint or release readiness.
- Recovery checkpoint (2026-10-08): the production-built client now exercises native
  Begin, real attended progress, leave/stopped scheduler, same-session reconnect/token rotation
  and cancellation against the real service/Postgres. Native manual actions and two purchases
  first prove actual positive idle production; unchanged Recovery balances are therefore not
  an inert-Company check. `make test-soul-recovery-composed`59876 also passes the unchanged
  five-minute activity through native Finish, exact stored/event/rendered Soul40→45 and reload.
  Soul40 is explicit setup, not a natural debit/history-replay proof. The bounded lifecycle
  joins normal composed CI. Final whole89786 passes Recovery and the preceding player/DB
  populations, pets/Reputation, Clout, payment negatives and Garden, then is RED in Tier2:
  a native upgrade carries revision11 after the previous purchase committed12 (RP-447).
  This is not a whole-target pass. The [Minigame API log](minigame-api-and-surface/log.md)
  retains failed setup attempts, the earlier blocked-Begin negative control and pending
  designated-review range. No gameplay, clock, catalog, limiter, CI-workflow, acceptance or
  release change; longer activities, browser renewal and broader recovery/AT tasks remain open.
- Preceding local checkpoint (2026-10-08): session request admission at `13acd872` and eight
  bootstrap/Minigame/Soul write routes at `ef10238f` now enforce their existing exact schemas
  before decoding or mutation. Eighteen session and thirty activity/bootstrap refusal controls
  pass against real Postgres, including unchanged stored rows and usable valid controls.
  Affected cold native tests/vet and byte-identical API generation pass. The first whole
  composed46706 is RED at Garden's upstream-observation truncation (RP-445), not concealed.
  Its separate proxy correction resumes upstream collection after browser cancellation;
  twelve observer regressions pass, with mandatory complete-body checks unchanged. Final
  whole `make test-game-ui-composed`82013 exits0, executing persisted populations, main play,
  pets/Cosmetics/Reputation, Clout, no-payment negatives, Garden and Tier2. This is the existing
  CI target run locally, not a hosted/full-CI or release claim. The [API log](api-foundation/log.md),
  [Account log](account-and-session-bootstrap/log.md) and [Garden log](minigame-server-garden/log.md)
  retain exact pending-review ranges and intermediate failures. No schema widening, rate,
  payout, balance, kernel, CI-workflow, acceptance or release change follows.
- Earlier minigame checkpoint: `950be202` (2026-10-08). The existing public
  Pitch HTTP lifecycle now independently checks its persisted cash payout, certified result,
  save revisions, required events/receipts, single quota spend and Founder replay, including
  mid-session reconnect and unchanged-state retries. The complete `gameserver` package passes
  against declared real Postgres16 on Linux/amd64, with no skips; a deliberately corrupted
  quota fails the new check. Native package tests/vet also pass, but native DB skips are not
  counted as integration proof. This is test-only strengthening of the existing solo public
  API, not a new async endpoint or complete Minigame Platform acceptance. The
  [owning log](minigame-api-and-surface/log.md#2026-10-08--bind-the-public-pitch-lifecycle-to-its-stored-outcome)
  records the exact pending designated-review range. Hosted run 37778647275 was rechecked
  read-only for the earlier overview: terminal red, four jobs pass and client/server fail.
  At that inspection newer commits were local; origin/main has since advanced to `cbd2069f`
  by an external push during the Recovery work. Codex did not perform that push; the newer
  hosted result is reconciled below. No full-suite rerun or release promotion is
  inferred from this documentation reconciliation.
- Latest native-care checkpoint (2026-10-08), range after `c977ad36`: RP-440's
  global busy guard discarded a completed trusted click; RP-441's inserted busy message
  independently moved the button during a pointer sequence. Separate old-behavior controls
  fail. The correction retains shared feedback, reserves its layout space, and lets fresh
  care input wait in the existing revision-bound queue. Final affected Chromium/WebKit
  suite passes 1,060 cases with four separate performance-only exclusions; types, build and
  client boundaries pass. Whole `make test-game-ui-composed`76975 passes all its existing
  real-Postgres and production-browser populations, including both care variants,
  Reputation, Garden, Tier2 and forbidden-payment negatives. The [Garage log](garage-player-surfaces/log.md)
  retains intermediate red candidates, scoped evidence and pending designated review.
  No retry/timeout, clock-policy, copy, CI, acceptance or release change follows.
- Preceding Minigame Platform checkpoint (2026-10-08), range after `ad74fd7c`:
  both solo/async backend populations pass against actual Postgres, including saved-state
  resume, certification and payout/failure/retry cases. The absent-DB control now reports
  top-level SKIP, which cannot satisfy the composed driver's required PASS marker.
  Earlier whole99680 passes; **its final whole64767 is RED** at native Pet Care feed:
  trusted click during busy, zero care requests (RP-440). Refresh/persisted populations
  and main gameplay pass; later Reputation/Garden/Tier2/negative builds are not reached.
  The [Minigame Platform log](minigame-platform-foundation/log.md) owns scoped evidence
  and pending designated review; the [Cosmetic log](cosmetic-shop-v1/log.md) retains the
  original activation failure, corrected in the later Garage checkpoint above.
  That platform batch made no retry/timeout, production or CI change. This is not
  complete AC1, whole-feature acceptance or release proof.
- Preceding Garden observer checkpoint (2026-10-08), range after `8e7e4427`: RP-438 is
  locally corrected with real old-document Referer binding, not a blanket unfinished-read
  exemption. Ten observer cases include actual Chromium/HTTP cancellation and negative
  controls. Final `make test-game-ui-composed` (39780) **passes end to end**, including
  Reputation, Garden and Tier 2. Garden enumerates 67 complete upstream responses and
  separately counts four old-document cancellations. Original 66940 and the unsuccessful
  timestamp-only attempt remain recorded in the [Garden log](minigame-server-garden/log.md).
  No production gameplay, deadline, retry or CI-workflow change; designated review remains.
- Latest Reputation integration checkpoint (2026-10-08), range after `f3851e8b`: its new
  native production-browser/service/SQL journey passes, including Escape cancellation,
  next-run-only bonus and an Exit plan; a disconnected purchase callback fails. Budget/cash
  are controlled setup, not earning or pacing. Its original whole composed66940 is **RED**
  at Garden's incomplete proxy observation RP-438; Tier2 is not reached in that run. The
  subsequent observer repair and successful whole39780 above do not erase that failure.
  The [Reputation log](reputation-tree-v1/log.md) owns detailed results
  and pending designated review. No full-CI, mint or acceptance claim follows.
- Latest Tier-2 construction checkpoint (2026-10-08), range after `205e0c7d`: actual candidate
  presentation consumption fixes RP-436's built-client crash. Native real-service/SQL gate,
  purchase, incorporation and reload proof now joins the existing composed lane; controlled
  cash is not pacing or earned-cash evidence. The affected Linux three-engine suite passes
  1,539 cases with six separate performance exclusions; the native Firefox launch failure is
  retained, not relabelled as green. The [Tier-2 log](tier2-content/log.md) owns full commands,
  final aggregate outcome, instrumentation corrections and pending designated review. No
  full-CI, complete feature acceptance, content mint, archive or release milestone follows.
- Latest construction checkpoint (2026-10-08), range after `4b2c72a0`: the existing composed
  lane now includes native Garden single/batched harvest against the real API and Postgres.
  Final whole lane passes, including existing player/Cosmetic/Clout journeys and no-payment
  negatives. The Garden case independently binds fifteen cash, two sends, receipts/events/logs,
  unchanged-state retries and rendered reload; disconnecting its native harvest callback fails.
  Its explicit mature-state setup is not elapsed growth, replayable maturation history or
  production content activation. The shared driver's session-expiry diagnostic also passes:
  absent automatic renewal remains confirmed, and only a test-operated refresh recovers.
  All processes are terminal. The [Garden log](minigame-server-garden/log.md) owns exact results,
  failed instrumentation attempts and the pending designated review. No feature acceptance,
  archive, hosted-CI success or release milestone follows.
- Finished research checkpoint: `0d9a9f43` (2026-10-08). All Tier-2 measurement processes
  are terminal.
  Original C1–C4 and bounded C5 have designated approval. C6 remains CHANGES REQUIRED for
  the inherited-policy test hole; Codex's regression-tested correction at `e6a64aff` needs
  the other party's review. The seven-literal pacing report and combined relevance diagnostic
  independently reproduce, but their acceptance gates do not pass. Controlled branch tests
  prove effects for two Tier-2 upgrades, not whole-content relevance or ordinary player choices;
  nap pod's branch remains unselected. RP-433 requires a versioned comparison-population
  decision, not balance retuning or golden regeneration. The [Tier-2 log](tier2-content/log.md)
  owns commands, digests, results and exact review ranges. No new playable tier, minted content
  or release milestone follows from completing this research/review checkpoint.
- Latest published CI verified read-only: `cbd2069f` (2026-10-08),
  [run37795560514](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/37795560514).
  All six jobs are terminal: schema, fast harness and composed pass; client, server and
  browser fail. Client executes10294 passing tests/943 skips, then rejects historical
  kernel commit50a3a514 (RP-131). Server retains27 Clout partition failures (RP-307), and
  additionally reports formula/replaycatalog duplicate/nesting admission conflicts (RP-449)
  and an account-family socket-revocation deadline (RP-450). A cold local formulas/replaycatalog
  run reproduces the same admission failures; these are not hosted-only. Firefox's shell primitive
  accessibility/focus/naming test keeps era_1995 after Enter instead of era_2000 (RP-451);
  browser reports33683 passes/one failure/15 skips. Since that run, RP-449 is locally
  corrected and RP-450's drain witness is strengthened; both need designated review,
  and RP-450/RP-451 remain open. Newer local Garage/Recovery/API/Account work has not
  been pushed; local selected/composed passes do not establish hosted/full-CI green.
- Earlier published source inspected: `f3851e8b` (2026-10-08).
  [Hosted CI 37778647275](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/37778647275)
  is terminal **RED: four jobs pass, two fail**. Composed, browser, schema and fast harness
  pass. Client runs 10,258 passing tests/928 skips, then the historical kernel guard rejects
  `50a3a514` (RP-131). Server reproduces 27 Clout accrual-partition failures (RP-307).
  The previous Firefox Garden failure does not recur in this hosted run; that is not a
  causal diagnosis or general reliability proof. The newer Reputation and Garden-observer
  commits have local evidence only. Codex did not push during this reconciliation.
  The earlier large unpushed counts and the following run are historical:
  [Hosted CI 37763502364](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/37763502364)
  was **RED: three jobs pass, three fail**, with all six jobs terminal:
  schema, harness and composed pass; client, server and browser fail.
  Client executes 10,223 passing tests/928 skips before RP-131's historical kernel guard
  rejects `50a3a514`; server retains the same 27 Clout partition failures (RP-307).
  Browser executes 33,425 passes/one failure/15 skips: Firefox Garden uproot re-consent
  records no callback (RP-430). The separate browser performance command does not run after
  that failure. The composed job executes all three built-client/real-Postgres journeys and
  both no-payment negative builds successfully; it is not clean-host release proof.
  The unchanged focused Linux Firefox Garden reproduction passes, not a diagnosis or fix.
  Native macOS Firefox executes zero tests after browser-session connection failure.
  Garden's new diagnostics preserve native input, existing assertions and timeouts;
  all255 affected Linux three-engine checks and strict types pass. A deliberately dropped
  test callback produces the intended failure with input/focus evidence, then is restored.
  The [Garden log](minigame-server-garden/log.md) owns the bounded results and pending review range.
  No full-CI success, feature acceptance or release promotion follows.
- Arcade's raw JSON admission correction `e478b248` passes shared Go/TS rejection vectors,
  444 affected Linux Chromium/Firefox/WebKit cases without skips, all four selected actual
  Postgres Arcade integration functions, and affected type/build/package checks. The full
  client test command passes before the separately executed historical kernel guard remains
  red. This repairs client admission, not the missing public Arcade API/host journey or content
  mint. The [Arcade log](minigame-demo-disc-arcade/log.md) owns the exact range and pending review.
- At `e478b248`, the full declared Linux/amd64 Postgres16 encrypted backup target now executes
  all six integration functions without skips: empty/populated dump/restore identity, unsafe
  target refusal, recovery-identity negatives and pre/post-deletion state. This closes the
  earlier skipped full-backup execution, not Deployment acceptance. Pre-deletion restore still
  revives the deleted account (RP-125). Scheduling, measured RPO/RTO, private operations and
  exact-artifact clean-host R-006 remain open; the [Deployment log](deployment-foundation/log.md)
  owns the result. The isolated test project is removed; no deployment or push occurred.
- Previous source checkpoint: `c0a85602` (2026-10-08), clean working tree before this overview
  reconciliation; 95 commits ahead of the locally recorded `origin/main`, not pushed.
  Adoption welcome now binds the actual local receipt and matching authoritative pet, rather
  than celebrating an ordinary read. Old-source regressions fail in Chromium/WebKit;
  1,020 affected browser executions pass (four separate performance-only exclusions), as do
  types/build/boundary/copy checks and the full real-Postgres/production-client composed target.
  The latter verifies one welcome update and quiet reload/remount with preserved navigation
  focus. Native Firefox failed to launch; this is not full-browser, whole-CI or AT approval.
  The [Pet log](pet-adoption-v1/log.md) owns the exact range after `d039fe66`, still pending
  designated review. Preceding PR-card integration `d039fe66` also has affected native and
  real-service proof in the [Clout log](clout-v1-and-pr-interns/log.md); projected-factor/hint
  definitions remain owner questions, not silently invented behavior. Neither batch advances
  the completed product milestone or resolves the numeric/clock/privacy/release holds.
- Previous source checkpoint: `58a9aeb2` (2026-10-08), clean working tree before its
  status update; 92 commits ahead of the locally recorded `origin/main`, not pushed.
  Cosmetic focus correction `24cefcd5` passes 230 affected browser cases, its real-service
  journey and both no-payment negatives. Clout notice implementation `58a9aeb2` passes 69
  selected Node and 1,014 affected Chromium/WebKit checks (four separate performance skips),
  types/build/copy/boundaries and the whole real-Postgres/production-browser composed target.
  Four stored re-attainment events produce two live-region updates; prior-earned Founder
  history is controlled fixture setup, not a played prior run. Both exact ranges await
  designated review. Their owning logs record evidence; no full CI, content adoption,
  feature acceptance, archive or 1.0 milestone follows.
- Preceding source checkpoint: `ca862e50` (2026-10-08), clean working tree before its
  editorial reconciliation; 89 commits ahead of the locally recorded `origin/main`.
  Deployment restore now refuses schema-only/nonrelational target databases (RP-423).
  Twenty actual Postgres cases and affected native tests/vet pass; the broader Linux
  deployment-release population remains red at two preflight checks in the disk-full Docker
  environment. Five full-backup integration functions skip in the selected service population.
  This is a completed local safety correction, not full backup/restore, clean-host or hosted-CI
  proof. Its exact range after `ed48bea2` through `ca862e50` awaits designated review;
  the [Deployment log](deployment-foundation/log.md) owns the results and exclusions.
  No product milestone, release approval, archival or push follows from this checkpoint.
- Garage batch `267dbf55` connects actual pet-status events to the mounted
  Pet Care surface and fixes an announcement collision: distinct events committed at the same
  revision no longer suppress one another, while replays remain silent. Old-source regressions
  fail; 158 affected Node and 814 Chromium/WebKit checks pass, alongside types/build/copy checks.
  The whole real-service composed target passes; its final observer refinement also executes
  in the passing focused Clout journey. One new sentence remains pending owner copy. The exact
  range `c557a027` exclusive through `267dbf55` awaits designated review. This does not repair
  the shared attendance-clock defect or complete Pet/Garage acceptance. The
  [Garage log](garage-player-surfaces/log.md) owns the detailed evidence and limits.
- Preceding Garage batch `c557a027`: OD-15's previously missing upgrade explanation connects
  the server projector, generated API, strict decoder and Desk. Real-service production-client checks bind actual
  window/requirement reasons to visible text and disabled-button descriptions; the whole
  composed target passes. Three new explanatory strings remain visibly pending owner copy,
  and designated review remains open. The [Garage log](garage-player-surfaces/log.md) owns
  the complete batch results and exact range; this is not full Garage or release acceptance.
- Account batch `069532f5` prevents HTTP caching of credentials across account
  creation, bootstrap, session creation and refresh, including failures and rate refusals.
  Old-source regressions fail; the corrected full Linux/amd64/Postgres Account and Gameserver
  populations pass with no skips, alongside affected native tests, vet and unchanged API
  generation. This does not change token policy or implement recovery/renewal UX. The exact
  range `634b8360` exclusive through `069532f5` awaits designated cross-party review; the
  [Account log](account-and-session-bootstrap/log.md) owns the evidence. The working tree was
  clean at this closeout; no push, full-CI or release claim is made.
- Preceding product batch `4055b4fe` fixes terminal minigame failure recovery: rolled-back
  API commands release only their own resolution claim, allowing an immediate retry instead
  of waiting for the five-minute lease. Real Linux/amd64/Postgres tests exercise rollback,
  cancellation, replaced-worker ownership, committed-result protection and exactly-once retry;
  all ten minigame database integration functions execute and pass. Affected client checks,
  types, build and API generation pass. Earlier `b0eab61f` prevents detached Pitch surfaces from
  reacting to late responses. Both exact ranges need designated review; neither completes the
  Minigame/API features or resolves Typer/Arcade's public-wire choice. The
  [Minigame API log](minigame-api-and-surface/log.md) owns results and limits.
- `4b8861d9` implements the source-independent Headcount allocation library in Go/TS, with
  shared vectors, independent arbitrary-precision checks, native browser and Linux/amd64
  execution. Only tests consume it: **there is no live Headcount allocation gameplay yet**.
  Seat-source and sibling contract conflicts remain unruled (RP-413/RP-414). `dbd8e274` adds
  native PR help under the existing Clout contract, with affected browser and real-service
  checks; its text remains pending owner copy. Account Recovery's player flow is a draft,
  not an implemented capability. These advances do not change the milestone reached.
- Deployment's private operations composition remains unexecuted here. A fresh 2026-10-08
  check finds Docker's overlay at zero available / 100% and the declared Caddy, Alertmanager
  and node-exporter images absent. No image substitution, cleanup or successful Compose run
  is inferred; the owning Deployment log records this environment boundary. The previously
  executed native alert delivery tests are narrower evidence, not a substitute for this gate.
- Earlier composed test batch `c78501d1` adds actual no-payment negative builds to the existing
  composed CI target. Both built Buy callbacks (checkout fetch and PaymentRequest) must fail
  through the intended detector after native input and the authoritative purchase; unrelated
  failure or incomplete cleanup is rejected. The final local `make test-game-ui-composed`
  run84658 passes the production-byte early-game journey, normal/Clout cosmetic-care journeys,
  real Postgres populations and both expected negative failures. Preceding `ee7f1118` proves
  transaction rollback, seven persisted refusals and eight concurrent identical retries against
  real Postgres. These are completed local test/tool batches, not new later-game features,
  complete Cosmetic acceptance or hosted success. Their exact ranges and remaining review
  live in the [Cosmetic log](cosmetic-shop-v1/log.md).
- Cosmetic correction `2fff9748` fixes replay-input admission: Go now rejects missing,
  null, unsafe or extraneous recorded Company context consistently with TS. Shared valid cases
  preserve full receipt/event/state bytes; rejected cases cannot mutate state. Selected real
  Postgres command, retry/conflict, Exit, Soul recovery and history tests pass, as do affected
  client tests, types, vectors and vet. The broader production population still fails the same
  27 Clout partition cases, and the historical kernel guard remains red. The Cosmetic log owns
  the exact commands and pending designated review; this does not complete Cosmetic acceptance.
- Garden pending/focus, connection and read-context repairs (RP-409/RP-410/RP-420/RP-421) are
  locally verified. Inactive reads remove the tab; Founder/content changes discard old grids
  and late reads, while ordinary same-Founder Company changes preserve the Garden. Broader
  Chromium/WebKit1148 pass/four separate performance skips, types/boundaries and the whole
  real-service composed target pass. New lifecycle cases use injected HTTP/socket boundaries,
  not an implemented account-switching UX or real Garden maturation/payout proof. The
  [Garden log](minigame-server-garden/log.md) owns exact review ranges, still pending; renewal,
  full accessibility, hosted CI and Garden acceptance remain open.
  The subsequent SG10 batch (RP-422) adds native keyboard/pointer `?` help, an accessible
  association to the visible why-text, and minimum button target sizes. All170 affected
  Chromium/WebKit cases pass. This completes the local mechanism, not the placeholder codex
  content or the full accessibility floor; designated review remains pending.
- Pet economy-isolation evidence (RP-408) now compares actual adoption/no-adoption inputs through
  the pacing consumer: all97 paired seeds pass, as do both Company Exit paths in Go/TS with
  deliberate economic-input/Company-write controls. The complete native harness CI target passes;
  its ordinary test population covers every policy without repeating the full study per push.
  This test-only batch still needs designated review; it does not close Pet acceptance or prove
  hosted CI, a live adoption-to-Exit journey or release pacing.
- Detached gameplay-host lifecycle (RP-407) is locally corrected: late responses no longer
  recreate sockets, restart the shell, refresh or submit a queued command after unmount; pending
  focus listeners are removed immediately. Affected Chromium/WebKit200, Node76, types/build/
  boundaries and the whole real-service composed target59352 pass. The new race is tested with
  controlled runtime completion, not server fault injection; designated review remains open.
- Pet adoption/history and cosmetic Exit persistence now execute at `7b4bc001` against real
  Postgres16: both Exit arms and all nine adoption/history corruption controls pass (49678).
  Docker has recovered; the original stalled container is confirmed absent, not retrospectively
  green. This selected database proof does not complete Pet acceptance or its remaining contracts;
  the Pet Adoption log owns the command and limits.
- Typer keyboard/focus correction (RP-394) and malformed-snapshot refusals (RP-395/RP-400) are locally
  verified. Native Chromium/WebKit input checks, the shared Go/TS corpus and real-Postgres Typer
  service tests pass. The old Go snapshot could panic at the exhausted prompt index; it now
  refuses without preventing valid final-prompt completion. Go also now refuses null numeric
  counters and cleared feedback missing its mandatory index, matching TS; affected Go/vet,
  real-Postgres service tests,76 native Chromium/WebKit checks and types/build/corpus pass.
  Public Typer registration/wire,
  adopted content, full player journeys and designated review remain open.
- Client runner collision (RP-396) is locally corrected without dropping checks: the full
  Vitest population passes10017/831 existing skips, and the actual composed target executes
  all30 byte-identical Node controls plus all three real-service journeys. Types/build and
  CI topology pass. This does not clear RP-131's history guard, Clout's numeric defect or
  establish three-engine/hosted/full-CI success; the CI log owns the correction range.
  Earlier full native39477 was RED:21635 pass/43 fail/10 existing skips. The separate R9
  pending-oracle correction (RP-397) retains actual native duplicate refusal and fresh consent;
  Garden's tab-stop correction (RP-398) proves menu, seed, Harvest all, substrate and Refresh
  paths. Final affected298 and full Chromium/WebKit79091 now pass21706/10 existing skips.
  RP-399's earlier missing-iframe cause/reliability remains open: no Typer/provider repair,
  retries, larger budgets, Firefox/Linux/hosted/full-CI success or feature acceptance inferred.
- R-012's existing-context SQL seam is now locally proven for64 diagnostic
  Company profiles: real Postgres16 jsonb preserves full restored state, raw
  production rate bits and evaluation results. All64 changed-context controls
  discriminate; cold related research tests and vet pass. This is bounded Linux
  ARM64 research, not a new save representation or Clout acceptance. RP-307's
  original27 partition failures remain; the owning Clout log records review/limits.
- Reduced-motion integration (RP-084) is locally corrected: numeric counters now honor
  startup and live browser preference changes while preserving the native prediction Worker.
  Old-source Chromium/WebKit regressions fail; final native Worker/DOM/cadence/unmount checks,
  affected shell/host/cosmetic tests, types/build/boundaries and whole composed81031 pass.
  Garage log owns performance evidence and pending designated review; this is not full R-005,
  assistive-technology, hosted CI or release acceptance.
- Vision entry correction (RP-393) now covers native Tab/Enter/Space, pending focus,
  duplicate suppression, existing Retry and authoritative Desk focus without overriding
  newer choices. Nine old-source failures become ten focused passes; affected native suite
  passes960/four performance-only skips, types/build/boundaries pass. Whole composed42217
  exits0 through actual HTTP201 bootstrap, production-client gameplay and both Cosmetic/care
  populations. The initial new oracle's mistaken HTTP200 expectation is recorded, not a
  server defect. Designated review and full accessibility/first-read/renewal/CI gates remain;
  the Garage log owns this bounded range.
- Latest native Pet Adoption correction (RP-392): Tab/arrow selection, pending focus/visible
  feedback, duplicate suppression and Not now/reopen focus handoffs now execute in Chromium/
  WebKit. Old component fails four regressions; final14 checks, types and build pass. Actual
  composed6685 was **RED at care** (RP-388), retaining the threshold evidence. The separate
  oracle repair now binds actual SQL outcome, pinned policy and each read's attended sample;
  mounted band text binds the host's own refresh, not the receipt's earlier band. Helper25
  controls and actual whole69425 pass, including default/Clout care/reload and native Enter/
  Space adoption/welcome focus. This is a bounded journey pass, not shared-clock/PA7 resolution,
  reliability, full CI, feature or accessibility completion. Pet Adoption and Cosmetic logs
  own their distinct ranges and pending review.
- Latest bounded API work: the category catalog now has its explicit owner descriptor (one of
  nineteen current artifact families). The shared error schema no longer makes public reads
  import private Account; affected tests/vet and byte-identical generation pass. Remaining
  descriptor exports conflict with the replay guard's broad watched paths, and intent receipts
  need a typed-dictionary contract reconciliation (RP-391). Neither exception nor schema
  amendment is silently adopted. The API log owns evidence and pending review.
- Cosmetic test-only commit `1d301497` extends the actual composed service journey to native
  Enter and Space Buy activation, exact one-request/receipt/revision binding, owned focus and
  persisted reload. Root37641 passes all three journey populations; affected Chromium/WebKit
  checks pass68. Native Firefox fails to connect, so there is no all-engine/full-CI green claim.
  Explicit focus setup is not a keyboard-only navigation/AT study. No production content or
  feature acceptance changed; the Cosmetic log owns evidence and pending designated review.
- Gameplay repair `e55eb1e9`: transition controls retain native pending focus, removed controls
  recover focus without overriding newer choices, and queued Wind Down binds both refreshed
  revisions. Final host checks: 174 passes / four performance-only skips; types/build/source
  boundaries pass. Whole composed run 99336 passes real Postgres/WebSocket main and Cosmetic
  journeys. Designated review remains; no full accessibility/performance/Firefox/release claim.
- RP-381's two hosted WebKit constructor assertions are locally corrected, preserving the
  native parser failure and request/credential checks. Node 27 and Chromium/WebKit 54 pass;
  types pass. Production handling and CI configuration are unchanged.
- RP-382's activation/SQL/action-sequence regression checks now distinguish historical source
  provenance from changed outcomes. All retained observation bytes outside source digests
  still compare exactly; artifacts remain unchanged. Actual Postgres research tests and TS772
  pass. Full real-DB production remains red only on the original 27 Clout partition cases.
- API requiredness/int64 compatibility corrections and the historical catalog database source
  are locally verified. Catalog retrieval preserves the stored epoch identity and refuses
  corrupt/unaccepted evidence. Both replay runtimes now support optional stored formulas with
  generated client schema metadata and exact int64 validation; old/new stored sets coexist in
  real Postgres without a product mint. The catalog HTTP route, other owner descriptors and complete public
  third-party verification loop remain open; the API log owns the bounded evidence/review ranges.
- Earlier GitHub check on 2026-10-08 records [hosted CI run 37641907701](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/37641907701)
  at `0f9f4214` is **RED**: schema, harness and composed pass; server, client and browser fail.
  Browser's two failures and server research-provenance failures have the local corrections
  above. Clout partition failures and client's historical kernel guard at `50a3a514` remain
  unresolved. At that checkpoint there was no hosted result for those local repairs;
  the newer published-source run and its distinct Garden failure are recorded above.
- A fresh GitHub run/job/log read also finds [Maintenance evidence 37723923274](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/37723923274)
  at the same `0f9f4214` **RED** on 2026-10-08: numeric-maintenance and harness-evidence pass,
  reputation-evidence fails. Its completed `make reputation-harness-check` records six Casual
  run-three starter timing ties, career-report drift and relevance-report drift; all970 H5
  sources are retained/recomposed before failure. This corroborates existing local H4/H5
  failures in the Reputation log, not a timeout or a newly diagnosed root cause. No report
  regeneration, balance retune or criterion relaxation is authorized by this status check.
- Recent implementation/correction spans still need exact-range cross-party review. Older
  approvals do not cover later code, and a self-review is not the designated gate.
- Existing session creation/refresh now mount from the generated API authority with exact
  status/error alternatives and the same shared IP limiter. Real Postgres and composed
  browser/service checks pass locally; explicit calls do not implement automatic browser renewal.
  The API log owns the additive pin refresh and pending review. The main composed journey now
  builds and serves production assets, verifies exact browser-loaded HTML/JS/CSS/worker bytes,
  and passes its real Postgres/socket/gameplay population. Cosmetics already used a separate
  production build. Neither establishes release-artifact/clean-host proof.
- Existing account creation/current-Founder/New-Founder routes now also mount from that authority,
  preserving handler bytes, shared limiters and archival/session binding. Real-Postgres Account/
  Gameserver packages and affected SDK/compiler checks pass locally. Exact empty-object generated
  typing is corrected (RP-386); original operation pins remain intact. Recovery UX, deletion/
  retention policy, bodiless response contracts and full API/designated-review gates remain open;
  the API log owns this new range. No new hosted CI result is claimed.
- Latest whole composed command28799 **PASSES** both main and Cosmetic journeys. Its preceding
  red40231 equip timeout remains recorded (RP-385): a controlled mounted-host test reproduces
  refresh suppressing a native click after pointer-down, and the G10 driver now uses one-shot
  ready DOM activation for equip/unequip, preserving exact receipts/persistence/deadlines. Final46
  native browser checks and types pass. Original40231's unique cause remains untraced; future
  failures retain input/request/response metadata. No general reliability or hosted-green claim.
- Intermittent composed timeout RP-378 remains open despite subsequent passes. Manual input
  retention/rate choice RP-368 and attendance-clock D-024 remain unruled, not silently fixed.
- Focused Clout fixture and the expanded composed target's three populations now pass locally
  (56007), including purchase/SQL proof. Earlier70861 fails in main gameplay: correct
  locked-Pitch409, missing visible notice (RP-389); later variants do not run in that attempt.
  A controlled native host test demonstrates a valid preemption route: an Exit offer unmounts
  Pitch before or after its notice renders. It is not unique attribution of70861. Public failure
  diagnostics retain lifecycle context. The driver now handles that interruption without
  skipping its visible-notice gate; root23009 exits0 through all three journeys, with no
  declines in that actual run. Post-response re-consent remains controlled-boundary evidence.
  No unique original-run attribution, reliability,
  hosted-green or adopted PR content claim. One selected run passed PR but failed care equality
  (RP-388); the original run remains untraced. The latest recurrence and deterministic
  threshold diagnostic and locally corrected care oracle are recorded above; no reliability
  or unique original-run attribution is claimed.
  The Garage/Clout/Cosmetic logs own evidence and pending review.

### What the weeks of work accomplished

The program replaced unsupported completion claims with code/runtime traces and discriminating
witnesses; closed the three serial account/API/transport witness batches; implemented substantial
Garage features; and found/repaired real numeric, input, focus, clock, wire and lifecycle defects.
Deployment and research publication work also progressed. Much recent effort is integration,
review and hardening of that implementation, **not delivery of the remaining late-game tiers**.
Test counts show executed scope, not completion percentage.

### Remaining program, in practical groups

1. Stabilize the current baseline: Clout numeric correctness, truthful kernel-history guard,
   remaining reproduced defects and current-head hosted CI. Never suppress failures to get green.
2. Consolidate coherent review ranges and finish Garage contracts/journeys: account recovery,
   startup/renewal, manual input, attendance, pet/API and minigame wire boundaries; adopted
   content and measured pacing rather than fixture-only release claims.
3. Complete player rights and task accessibility under explicit decisions, with actual workflows.
4. Finish and review deployment corrections; rebuild exact artifacts and rehearse on an
   authorized supported clean host, including backup/restore/rollback and alerts.
5. Continue the full [milestone roadmap](roadmap-1.0.md#milestone-path): Tier 3–5 world/social/
   combat/events and then Tier 6–8 Transcendence, variants and challenges. These depend on shared
   contracts, content and real player proof; no preview narrows the goal.

The [active RFC index](../rfc/README.md), [owner/author decision queue](platform-alignment/decision-queue.md),
[Garage log](garage-player-surfaces/log.md), [Deployment plan](deployment-foundation/plan.md)
and [full roadmap](roadmap-1.0.md) own detailed authority, evidence and work. This overview
does not accept drafts, resolve decisions, archive implementations or authorize publication.

### Earlier Garage and platform checkpoints

The following entries preserve earlier evidence and then-current next actions. They are not the
latest executable queue; use the overview and owning plans above.

**Historical Garage work:** GS1-A3 exact rendered Fiscal fixture edges. Four public
decoded component cases cover early-1/early/guaranteed-1/guaranteed with exact
visible text/risk/countdown/native readiness/curtain/callbacks; eight actual
source faults fail and restore exactly. Final full native Garage262 and
isolated performance two passes; client9,814 passes /455 browser skips,
types/unchanged build/boundaries/static topology pass. No product/copy/clock/
CI change or live-time/service/pacing/Firefox/AT claim. RP-336 stale props/
body wording recorded for author action. Entire span after 57efdc39 through
final records needs Claude independently of Desk54c8c8da..57efdc39, source
guard d3ce0f76..54c8c8da and all earlier spans. Fresh Docker df still zero
available/100%; unused repo cache metadata inspected, no deletion. Await
human scoped cleanup authority, not broad pruning; capacity repair/recheck
then declared cold Linux/SQL tests remain priorities. Safe GS2-A3 source
binding work remains available. Full-nine-tier 1.0/author/body/GS4/full AC7/
default-player/privacy/platform/numeric/AT/review gates remain, no promotion.

**Preceding Garage work:** test-only GS6-A3 paired current-source Desk measurement
at 320 CSS px. Both native engines measure document/main 320/320, Desk/chrome
284/284, extents 0–320, nodes 90→93 as the three GS6 text additions become
visible. Six real faults fail/restored exactly; no historical-before-build/
647 px accessibility closure/actual zoom/AT/Firefox/real service claim.
Full native Garage 254 passes and isolated performance two passes; client
9,814 passes /451 browser skips, types/build/boundaries/static topology pass.
No product/CI change. Entire new span after 54c8c8da through final records
needs Claude separately from source guard d3ce0f76..54c8c8da and all earlier
ranges. Next accepted native work: exact rendered Fiscal phase fixtures
GS1-A3. Docker capacity/full Linux/persisted DOM/author/body/GS4/full AC7/
AT/default-player/privacy/platform/numeric/full-nine-tier 1.0 holds remain.
No checkbox/lifecycle/archive/push/release promotion.

**Preceding Garage work:** AC4 actual-source discrimination and RP-335 cosmetic
command source guard. Four valid component faults fail existing policies;
the old boundary's fake-button survivor is recorded, not present in healthy
product code. New verification-only Go/Svelte AST check derives kinds from
production and requires decoder membership; five real faults plus ten Go/
eleven Svelte negatives discriminate, sources restored. Final client 9,814
passes /450 browser skips, types/build/boundaries/isolation/payment/static
topology/isolated performance pass. Cold cosmetic parser/transition passes;
two SQL tests explicitly skip. No production/CI/Make byte moved; no whole-
program/runtime/Shop acceptance. Complete range d3ce0f76..54c8c8da
requires Claude independently of performance 127eb052..d3ce0f76, layout
bddfc58e..127eb052 and all earlier exact spans. Next accepted work: GS6-A3
pinned 320 px whole-Desk comparison. RP-331 awaits repaired/rechecked Docker
capacity (last 0 free/100%); no run/cleanup here. Full Linux browser remains
RED/held. Author/body/GS4/full AC7/actual zoom/AT/default-player/privacy/
platform/numeric/review/full-nine-tier 1.0 gates remain; no promotion.

**Latest repair:** `45bf2edb`, predeclared `a8f2a527`, failing-first
`7cac400d`: RP-326 Fiscal/care outcomes now belong to the submitting panel;
late responses do not display on another tab.24 baseline DOM failures;
five compiling placement/duplicate/origin faults fail and restore exactly.
Full native Garage Chromium/WebKit174/performance1, client9,737/410browser
skips/types/build/boundaries/copy/manifest/topology pass. No prior assertion
relaxed. Changed-tree real composed NOT rerun under RP-236; prior source pass
is not substituted. Native applied/ordinary refused runtime-double evidence,
not HTTP-failure/queued-origin/all-surface/AT/Firefox/fullGS0.6/release proof.
Entire new span aftere950216a through this tracking edge needs Claude.
Browser evidence separately exact4dcd9969..e950216a; Fiscal separately exact
b64a91af..4dcd9969, all earlier ranges independent. Next predeclare native
HTTP-error outcome-origin/refresh controls under GS0.2/GS0.6; no new Docker
browser/composed run before measured capacity repair/recheck. Full1.0 active.

**Latest verification:** unchanged full cold Linux ARM64 browser lane at
`36c28692` is RED:373/378 files pass,30,366 tests pass/four fail/six explicit
skips, two Firefox suites uncollected. Fiscal keyboard native-worker error,
Reputation keyboard timeouts and WebKit worker observation remain open
RP-327/328. Diagnostic3ade389d is incomplete/exit137: actual owned-container
overlay0available/100% plus Firefox `NS_ERROR_FILE_NO_DEVICE_SPACE` confirm
RP-236 environmental invalidity. Codex omitted required capacity preflight;
recorded, no broad cleanup or false root-cause/green inference. No third
Docker browser run until capacity resolved. Existing two-engine Fiscal proof
and all independent review ranges stand; evidence-only span after4dcd9969
through its tracking edge needs Claude. Next accepted native GS0.6 outcome
ownership/isolation diagnostic RP-326; full-nine-tier goal remains active.

**Latest Fiscal correction:** `891aee10`, predeclared `bd36f5bc`, failing-first
`d0651029`; RP-323/324 consumer defects locally repaired, not approved. Pending
intent/read, native keyboard, stale/restart, row reasons and focus boundaries
discriminate:33 baseline failures; eleven valid behavioral faults fail and
restore exactly. RP-325's two shortcuts survive old50 cases, then fail the
refined parser-accepted reason/unequal-distance populations. Two invalid
probe parse-error runs are disclosed, never counted. Final Garage Chromium/
WebKit150/performance1, client9,737/398 explicit browser skips/types/boundaries
pass. Same production hashes' build/copy/manifest/topology/actual Postgres/
WebSocket composed pass before final test-only refinements. Not a hosted-CI/
Firefox/AT/full-Garage/release gate. Entire new span after `b64a91af` THROUGH
this tracking edge needs Claude; earlier spans independent. RP-326 separately
records unscoped chrome outcomes, not yet a native failed case. Next: existing
full Linux ARM64 three-engine `make test-browser-ci`, cold isolated install (now RED above);
then separately predeclare GS0.6 notice-origin/isolation tests. No workflow/
policy change. All author/owner/privacy/numeric/content/platform/release holds
and full nine-tier scope remain, goal active/progress.

**Preceding shared refusal correction:** `efa557b8`, predeclared `3f231ba2` and
test-first `d362c95e`, locally repairs RP-322. Rate-limit/exclusive responses
use the existing refresh effect: held read prevents reactivation; only fresh
player consent sends Founder8/fresh ID. Baseline two mapper/four native
failures; separate arm faults fail one mapper/two native assertions each,
restored exactly. Final Garage Chromium/WebKit100/performance1, client9,737/
373 explicit browser skips/types/build/boundaries and actual Postgres/
WebSocket composed pass. Refusal population is runtime-double evidence, not
real-service429/Soul-session or minted release proof. Entire new span after
`3f867956..b64a91af` needs Claude. Earlier care range stays
exact `c7d8f815..3f867956`. Its then-next Fiscal tests/repair now execute above.
Firefox/AT/author/raw-
care/content/numeric/platform/release holds and full nine-tier scope remain.

**Preceding care correction:** `ea06183b`, test-first `0784f126` and predeclaration
`970d47a5`: care pending focus/text, native Tab and eligibility-change focus
handoff, stale explanation and restart readiness are locally repaired. Full
Garage Chromium/WebKit82/performance1, client9,735/364 browser skips, types/
build/boundaries/copy/manifest/topology pass. Eight independent native faults
and real DOM callback severing fail, all restored exactly. Actual built UI/
Postgres/WebSocket care after DOM adoption binds a positive receipt and public
state at Founder revision5; reload retains feed ineligibility. Test-only bundle,
not release-manifest/status-crossing/raw-care/AT/full-Garage acceptance.
Firefox zero executed before connection timeout; failed teardown stopped,
RP-256 remains. Entire new Codex span after `c7d8f815` THROUGH this tracking
edge needs Claude; now exact `c7d8f815..3f867956`. RP-319/320/321 locally
corrected, not closed. Its then-next shared refusal checks execute above.
All prior holds/full1.0 remain.

**Preceding Garage pet review:** original Claude `7a61e4b6^..7a61e4b6`, all eight
paths: designated **CHANGES REQUIRED for the existing RP-132/GS4×PA7 author
contract hold**, mechanical subset verified. Four native pet assertions and
cold Go fact pass; five compiling callback/revision/text/overlay/empty-map
faults fail and restore exactly. No permanent product change. Separate
RP-318 actual decoder diagnostic ignores12 valid pet-band transitions and
seven malformed payloads; required announcement copy is absent. Its then-next
care supplement and real-server check now execute above; raw internals/new
prose/GS4/CI/archival remain unproved. Full nine-tier goal stays active.

**Preceding Garage proof correction:** RP-316's shared GS5-A4 observer through
`0befc01c` now checks request/revision/run/cash/buff bindings and successful
successor reads. Test-first failures and six independent compiling omission
probes discriminate; source restored exactly. Native86/performance1, client
9,735/348 skips, types/build/boundaries and actual Postgres/WebSocket composed
lanes pass. Actual claim: click buff at revision26 after15 DOM clicks, NOT
Lucky integration. No product runtime/schema/kernel/copy/CI changes. Entire
Codex range `c0eb3dc5..876b331c` needs Claude; RP-316
is not closed. Its then-next original pet diagnostic review is recorded
above, preserving RP-132/GS4×PA7 author hold. RP-313 and
all prior holds/full nine-tier objective remain; no lifecycle promotion.

**Preceding Garage diagnostic correction:** RP-317's accepted GS5/GS0.2 mapper
flag and host reporting repair through `092ef3eb` passes types/client
9,692/348 skips/build/boundaries, native Chromium/WebKit Garage50/performance1,
and actual Postgres/WebSocket composed tests. Ordinary refusals are unchanged;
three compiling flag/outcome/error reporting faults fail independently and
are removed exactly. The earlier stalled combined baseline was invalidated,
not counted; single-engine failing baselines terminate normally. No Lucky,
Firefox/AT/full-Garage/whole-CI/release proof. Entire new Codex span after
`0f3a1a7a` including all tracking edges needs Claude. Desk correction is
separately exact `d90aded7..0f3a1a7a`, producer `87fd23d4..d90aded7`.
Its then-next RP-316 snapshot-credit oracle is now recorded above. All prior holds
and full nine-tier goal remain, without status/archive/mint/publication.

**Preceding Garage Desk review:** original Claude `f32f6175^..f32f6175`, all
fourteen paths: designated Codex **CHANGES REQUIRED**, RP-314/315/316. Narrow
receipt-reason/numeric-cap/native-focus correction through `d9e4e274` passes
40 Chromium/WebKit cases, including actual-minute idle and native Enter/Space.
Independent omission/negative-focus faults fail and are removed exactly.
Types, client 9,691/343 explicit browser skips, build/boundaries, complete
copy/manifest/topology pass; two actual composed runs claim click/production
buffs. The unchanged Lucky oracle accepts a null successor in an isolated
actual-function probe; no Lucky payout defect or integration coverage inferred.
Whole new Codex span after `d90aded7`, including all record edges, needs Claude.
Next accepted work: separately predeclare RP-316 oracle and RP-317 invariant
tests. RP-313 and all prior author/numeric/content/platform/review holds remain.
No full Garage/hosted CI/all-engine/AT/release claim or reduced 1.0 scope.

**Preceding Garage producer review:** original Claude `cdb8fe61^..cdb8fe61`, all
twelve paths, is **CHANGES REQUIRED — RP-313**: accepted GS0.1 still requires
`features.active_play`; the implementation uses optional `features.opportunity`.
Author reconciliation must address shape/fact/absence rules and the pure-read
versus discarded-clone wording without bypassing API C2. Eight compiling fault
probes fail; all temporary product bytes restored exactly. New read-only/alias/
activation/refusal tests pass, as do cold Go/vet, actual Postgres selected
population, API generation, client 9,691/340 skips, types and boundaries. The
existing Postgres witness checks rates/transitions, not a GS5 claim. New Codex
supplement `87fd23d4..d90aded7` needs Claude. Its then-next diagnostic review: Desk
consumer `f32f6175^..f32f6175`, all fourteen paths, is now recorded above;
no full Garage/CI/release gate.

**Preceding Garage review/correction:** designated Codex review of original Claude
Fiscal/buff/reflow `6594b646^..6594b646`, all five paths, is **APPROVED, bounded
batch only**. Credit-sum/buff-disconnection/badge-clear/oversized-card probes
fail their named oracles; nav-wrap omission survives honestly. Source restored
exactly; final 9,691 client tests, Chromium/WebKit Garage 34 and performance 1,
types/boundaries pass. This is not full Garage, all-engine/AT, real-server or
release proof. Its then-next GS5 producer review is recorded above, and its
separate Desk consumer remains pending. RP-312's Codex range
`7aab0e2e..900f409e` needs Claude.

The preceding original achievement/meter review
`301728c8^..301728c8` covers all 14 paths: **CHANGES REQUIRED**, RP-312 meter-direction
domain mismatch. Test-first `49576162` reproduces 8 failures; the accepted GS0.3
repair and 13 new cases now pass (9,691 client tests/340 explicit skips). Real
delivery-gate and host-dedupe omissions fail; exact source restored. Native
Chromium/WebKit Garage 34 and isolated performance 1 pass; types/build/boundary/
copy/manifest/topology gates pass. Firefox/real-server acquisition/full Garage/
whole CI/release are not claimed. Exact `7aab0e2e..900f409e`, including records, needs
Claude; prior Clout spans are independent. The subsequent remainder review is
recorded above. Full nine-tier goal and all prior
owner/numeric/platform/content/review holds remain.

**Latest Clout harness prerequisite:** all12accepted criteria mapped; new
Go simulation effect-mask tests pass16literal rates/32full-state advances,
8live companions/four refusal entrypoints. Compiling guard omission fails
8rates/16advances; source restored. Cold focused production+harness/vet pass;
full production still ONLYoriginal27AC6red. No runtime/balance/kernel change.
DG-D/D-021 architecture stays unselected; the provisional proposed-next note
is forward-corrected, not runtime authority. The then-next original Garage
GS0.3 review is recorded above. Newspan after64f836c2
including records needs Claude; all previous owner/platform/review holds remain.
[Criterion map and bounded evidence](clout-v1-and-pr-interns/acceptance-reconciliation.md).

**Latest Clout receipt checkpoint:** separate RP-31179285a14 at3868e3d2,
tests-first03f513ef. Real ordinary/refreshed/Exit receipt producers now emit
the exact CV5-derived object, propagate errors, preserve legacy bytes; kernel166.
Red baseline and compiling omission/ownership/factor controls discriminate;
four reports re-observed with ONLY source pins and new receipt object, not game
states/events/inputs/populations. Client9678/340skip/types/vet/build/topology/
serial native migrations+production/gameui pass; ONLYoriginal27AC6 stays red.
Next reconcile residual accepted Clout criteria and select a distinct unblocked
accepted lane. AC3/DG-B/representation/RP-308/owner/content/platform holds remain;
whole newspan after3868e3d2 INCLUDING records needs Claude, older spans independent.
No fullClout/wholeCI/AMD64/hosted/mint/release promotion.
Re-executed kernel-history gate still fails at historical50a3a514/RP-131;
the new semantic bump does not cure that independent CI blocker.
[Evidence/limits](clout-v1-and-pr-interns/receipt-projection.md).

**Previous Clout migration checkpoint:** test-only36924143 at7ca728ca, reader
corrections44c0f8a8. All five exact CV4 names now execute real Go/TS load/Exit/
replay and three invalid-state boundaries. Corpus10/baseline20 preserves all
11legacy/4Founder cases and the SHA-pinned source. Missing-row/false-SHA and
actual derivation omission controls fail; runtime source restored exactly.
Serial native migration and activation/persistence companions pass; initial
overlapped shared-DB observations are invalid and disclosed. Client9646/340skip/
types/vet/build/topology pass, original27AC6 remain red. No product bytes/fullCI/
fullClout/1.0 promotion. NEXT separately scope RP-311 accepted-CV5 missing derived
receipt object; AC3/DG-B/representation/RP-308 and all prior gates independent.
Entire newspan after7ca728ca INCLUDING records needs Claude.
[Evidence/limits](clout-v1-and-pr-interns/migration-corpus.md).

**Previous Clout persistence checkpoint:** test-onlyfe0e4196 atf038a467, with
predeclared adapter/outbox corrections. Sixteen real Service/Store/Postgres
paths pass:168 logged attempts(120applied/48refused),168 later retries,32
conflicts,48 exact six-stage DB faults and16 event-outbox negatives. Complete
rows/receipts/events, immutable resources, actual five-snapshot retention and
all stored transitions to the saved head agree. Seven identical initial rows
are diagnostic padding; runs remain OPEN. Claim/lookup/retention mutants fail
on actual properties and restore exactly. Broad production Integration,
focused12research/integration/noSkip, client9640/340skip/types/vet/build/
topology pass; cold Go still fails ONLYoriginal27AC6. No product/oldartifact
byte changed or wholeCI/hosted/AMD64/fullClout/1.0 claim. Next reconcile remaining
CV4 migration and CV5/AC3 receipt obligations; representation/RP-308 and all
prior owner/content/platform/review gates remain. Entire newspan afterf038a467
including records needs Claude. [Evidence/limits](clout-v1-and-pr-interns/sequence-persistence.md).

**Previous Clout integration checkpoint:** predeclared9974f1d3 at0d866359,
test-only CV3/CV4. All four actual opportunity effects appear in16 continuous
Go/TS sequences/136 command attempts plus48 unchanged-state claim refusals per
runtime. Claims, PR2, compute debit/burst, mode switches, expiry, permits and
provision ticks match full receipts/events/saves. Three mistaken clock/tick
instrument assumptions are disclosed; real census/event/contribution/burst
mutations fail and restore exactly. Client9640/340skip/types/vet/build/topology/
native11 populations pass; cold Go retains ONLY original27AC6 failures.
No product byte changed, wholeCI/mint/fullClout/1.0 promotion. Next separately
predeclare persisted claim/burst/retry boundaries; accumulation/RP-308 and all
prior holds/review spans remain independent. Entire newspan after0d866359
including all records needs Claude. [Evidence/limits](clout-v1-and-pr-interns/sequence-research.md).

**Latest public-projection checkpoint:** RP-302 test-only supplement after`c70861c1`
connects actual purchase/Exit to stored GameUISnapshot: current factor1/rate5
stays fixed after purchase; next1.003 updates, then actual run3 freezes1.003 and
projects5.015. Complete current rows/head/frozen rows,12-table reads/retries and
recorded Founder history pass; three source-fault families discriminate, exact
restoration. Cold relevant Go/vet/types8241client units/339skip/local topology
and actual production/gameui SQL38top-level/143subcases/no skips pass. Entire
new range including records needs Claude, earlier spans independent. Diagnostic
initial earned6/generated5/seeded history, not natural/default browser/fullAC7/
hostedCI/minted release/1.0. Next reconcile remaining accepted gates across
Reputation and other unblocked implementation lanes, not add duplicate evidence.
All full-goal/author/environment/release holds remain. [Evidence](reputation-tree-v1/log.md).

**Previous SQL checkpoint:** RP-301 test-only supplement after`16da5904` continues
three actual old-pin→tree activations through owned-purchase refusal, real frozen
production/Garage, second no-plan WindDown and run4 without later state seeding.
Original24/9applied/15refused/3fallback remains. Repeated non-cumulative starters,
immutable old rows, both completed Company verifiers, complete four-entry Founder
history/source links, exact twelve-table retries and ongoing run4 full replay pass.
Three compiling source faults each fail three new named checks; other21 arms pass,
sources restore exactly. Initial test-clock failure preserved/corrected, game gate
unchanged. Cold relevant Go/vet/types/client8241pass/339skip/local topology and
whole production SQL Integration14.852s pass. Full span including final records
needs Claude; earlier spans independent. Diagnostic initial earned6/stored offers/
seeded history, not natural/default browser/fullAC15/hostedCI/1.0 proof.
Its then-next AC7 census/public projection supplement now executes above.
Full goal and prior holds remain; no production/copy/epoch/CI workflow/cleanup/
status/archive/push. [Evidence](reputation-tree-v1/log.md).

**Previous replay checkpoint:** RP-300 test/generation-only supplement after
`a21da467` compares27 complete Company/Founder pairs in Go/TS,12 applied/15
refused per stream. All three exit types cover old-pin activation, five plan
first failures and absent-tree no-plan controls. Independent accounting/
starters/taxonomy, full state/receipt/events/result-pin and36 copied-evidence
controls pass; valid old-pin substitution is separate from malformed-hash
defense. Go faulted generation refuses overwrite; TS selector/detail/no-plan
faults fail3/12/3 new cases. All source and historical corpora restore exactly.
Complete relevant Go/vet/type/client8241 pass/339skip, local topology controls
and actual SQL36/143/no skips pass. Full new span including records needs Claude;
prior spans independent. Its then-next R8 verification census and cross-Exit
history continuation now execute above. Diagnostic
initial earned6/stored offers, not natural/default browser/SQL/fullAC9/hosted
CI/1.0 proof. Full goal and all prior holds remain; no product/copy/epoch/CI
workflow/cleanup/status/archive/push. [Evidence](reputation-tree-v1/log.md).

**Previous SQL checkpoint:** RP-299 test-only range after `e5a69731` executes24
WindDown/both-offer profiles: nine actual old-pin→new-tree activations and
fifteen first-failure whole-Exit refusals. Mid-run purchase cannot activate the
latest tree; full heads/pins/frozen/events/receipts/history and exact retries
verify. Three sequential plan-free fallbacks complete the refused inactive-tree
runs without reseeding. Four source-fault families discriminate; skipped plan
application hits the earlier parity guard, shared spent reset the new persisted
oracle. All production bytes restore exactly. Complete relevant Go/vet, strict
TS/Svelte and8186 client units/339 skips pass; actual production SQL36 top-level/
143 subcases/no skips passes. Full span needs Claude including final records;
earlier ranges independent. Its then-next separate R8 Go/TS cross-pin activation/
first-failure supplement now executes above. Diagnostic
initial earned6/stored offers, not natural/default browser/minted release/full
AC9/CI/1.0 proof. Full goal and prior holds remain; no production/copy/epoch/CI/
cleanup/status/archive/push. [Evidence](reputation-tree-v1/log.md).

**Previous replay checkpoint:** RP-298 test-only/generation supplement after
`c3ab42d0` now byte-compares ten offered-plan Company and ten Founder arms
in Go/TS (eight applied/two rejected each). Both offer kinds, payout-funded
plan6/refusal23/absent/empty/promise floor; complete state/receipt/events/pin,
independent payout18/20/starters/bonus/order and copied-input refusals.
Prospective-credit, TS event-cost and rejected-Founder delta faults fire;
sources restore exactly, faulted generation cannot overwrite the fixture.
Old corpus byte-unchanged. Cold complete relevant Go packages/vet, strict
TS/Svelte,8186 client units/339 skips and actual SQL35/119/no skips pass.
Not default player/browser/natural pacing/fullR8/AC9/CI/1.0. Full new span
needs Claude, earlier ranges independent. Its then-next persisted R6 activation/
refusal matrix now executes above; separate cross-pin Go/TS parity census is
still next. All full-goal/author/environment/
release holds remain. [Evidence](reputation-tree-v1/log.md).

**Previous SQL checkpoint:** test-only range after `8ee45af3` proves both live
gate-produced offer kinds with zero initial Reputation and Exit-funded plan6.
Unaffordable last entry preserves both heads/offer/game rows; applied Exit
verifies ordered events/receipt revisions/starters/frozen bonus/summary and
both Company/Founder histories. Copied cost/order, live producer/credit/valid
offer-ID/hash-conflict probes discriminate; exact source restored. Cold whole
production/save/reputation/gameui plus vet pass; real production Integration
35 top-level/119 subcases/no skips passes; client units8167 pass/339 skip.
No new TS offer corpus/browser/natural pacing/full AC9/CI/1.0 claim. Full span
needs Claude including records; earlier spans independent. Its then-next R8
offered-plan parity is locally supplemented above, preserving existing corpus/
authority. All author/environment/release holds and full goal remain.
[Evidence](reputation-tree-v1/log.md).

**Latest request checkpoint:** RP-297 test-only work after `6dc16a58` adds
paired R6 Wind Down/Accept Offer size/type/ID/order/identity boundaries.
Three tests/59 reported subcases pass; six compiling probes fail2/2/12/40/6/2
and source is restored exactly. Focused cold tests/vet and full production
SQL Integration pass (34 top-level/117 subcases/no skips). No production rule
needed correction. Full range needs Claude including records; earlier spans
remain independent. Its then-next offer-with-plan SQL population now executes
above. Parser proof is not that journey, a64-node live
tree, full AC9/CI or release acceptance. All prior author/environment/full1.0
holds remain. [Evidence](reputation-tree-v1/log.md).

**Previous SQL checkpoint:** test-only range after `82830dcb` now executes all14
R6/AC9 applied-plan write faults on declared Postgres. Exact injected errors,
complete persisted rows and both stream heads prove rollback; retention has
real old rows to prune. Normal Handle commits/replays/retries once. Committing
despite retention failure changes11 tables and fails the oracle; source is
restored exactly. Restored production Integration population passes cold
(34 top-level/117 subcases, no skips); focused Reputation/Exit/Bonus tests and
vet pass. This full range needs Claude including final records, independently
of earlier spans. Its request-validation census/population is now executed
above; offer-with-plan SQL remains next. No product/CI/epoch/copy
change, cleanup, archival, push or release promotion. All author/environment/
nine-tier/platform holds remain. [Evidence](reputation-tree-v1/log.md).

**Previous SQL checkpoint:** the cached declared Postgres lane executes despite
Docker's capacity hold. Test-only work after `ffd1b673` repairs three dormant
taxonomy setup defects (RP-294–296) and adds the missing composed two-Exit
Reputation career. All21 taxonomy profiles execute. Both completed Company
runs and the three-entry Founder history verify; run3 continuation matches
its full stored head. Corrupted factor/head evidence is refused, prerequisite
omission fails two profiles, and available-instead-of-earned bonus fails the
career. Sources are restored exactly. Whole production Integration population
passes cold; selected Reputation/Bonus tests and vet pass. This is diagnostic
SQL proof, not natural pacing, minted/browser/full CI or 1.0 approval.
Full span after `ffd1b673` needs Claude, including final records; header span
`2db69792..ffd1b673` and earlier ranges remain independently pending. No cleanup
or push. Its then-next R6/AC9 write-boundary population is now executed above.
Formula, invalidated-plan, preview/display, Firefox, harness and release holds
remain. Earlier checkpoint-specific next actions below are historical.

Latest tooling checkpoint: RP-281 is locally corrected underda7cfa56. Generator
and regenerated Go are paired; output equals gofmt(old output), with other
artifacts/copy658/hash/manifest byte-unchanged. Five goldens, collision refusal,
six corruption controls, three template severings and actual output drift
discriminate. Cold client checks/vet pass; core package union passes after three
sandbox loopback failures rerun narrowly (original core command stays failed).
New Go producer tree1c54c2c3d0139b50e97937bfe88c65f339cab6b5; old dated
career reports are not fresh evidence at this tree. Whole span after3df3ff32
needs Claude. No numeric/kernel/English/CI policy change or release promotion.
RP-284 reader/delivery grounding has now executed below; its display boundary
and RP-283 authoritative preview still need author contracts. Full1.0 holds remain.
[Evidence](reputation-tree-v1/log.md).

Previous recovery checkpoint: RP-290/RP-291 host purchase readiness locally
corrected afterb5ca3e7d. Test-first64 failures/16 controls; scoped readiness/
existing offline notice and internal nonterminal recovering status.80 new
native cases pass; compiling host faults fail32/8/64/16, exact restoration.
R7 exact sequences retain all event/dedup assertions and add immediate-order
checks (notification omission fails two). Final408 selected Chromium/WebKit
functional plus isolated performance,8167 units/313 skips/types/build/boundaries/
copy/manifest pass. Initial strict-message failures and three-engine attempt
remain recorded; latter is RED, Firefox0 execution with native launch failure.
No new auth/transport/wire/content policy or full AC12/SQL/mint/whole CI claim.
Fresh Docker root100%/39,784KiB, DB tmpfs1%; no cleanup/SQL proof. Full new span
needs Claude; RP-288 e312b4d7..b5ca3e7d and earlier ranges independently pending.
RP-289 tooling/RP-283/RP-284 author contracts and full nine-tier/platform holds
remain. Next ground authoritative plan input replacement without inventing
reselection policy. [Evidence](reputation-tree-v1/log.md).

Plan-update supplement after00335488 is test-only:20 new native cases and
three compiling faults (8/20/8 failures, exact restoration). Restored428
selected Chromium/WebKit plus isolated performance,8167 units/323 skips and
types/build/boundaries/copy/manifest pass. RP-292 reproduces hidden owned ids
remaining in a plan and negative over-budget projection; characterization
tests explicitly do not approve that UX. Reset/pruning/refusal contract awaits
the R9 author. Full new range needs Claude; recovery range b5ca3e7d..00335488
and earlier ranges remain independently pending. Next safe accepted work:
R9 header accounting, frozen/next bonus and published-formula rendering.
Firefox/SQL/capacity/author/owner/release holds and full1.0 scope are unchanged.

Header supplement after2db69792 is test/instrument only:32 new native cases,
five compiling faults32/20/20/8/32 and exact restoration. Six existing Go
projection/bonus tests pass cold. Final460 selected Chromium/WebKit, isolated
performance,8167 units/339 skips and root types/build/boundaries/copy/manifest
pass. Published-formula criterion RP-293 stays RED: five metadata controls
match but percentage names/types conflict with raw-ppm binding and admitted
fractional percentages. Author action is explicit, not solved by green DOM
characterization. New full range needs Claude, earlier spans independently
pending. Docker cache-only cleanup permission requested, NOT granted/executed;
root still100%. Next exact cached declared-Postgres career preflight/population.

Previous runtime checkpoint: RP-287 is locally corrected in the range after
61d6c8eb. Recovered/live older-start defects were reproduced before correction;
bounded Company start revision prevents them across reconnect/full-sync.
All61 reader cases pass, three compiling probes fire7/3/3 and restore exactly;
8167 units/255 skips, types/build/boundaries/copy/manifest and166 Chromium/WebKit
functional plus isolated performance pass. Four initially invalid fresh-state
fixtures are disclosed, not hidden or blamed on production. Controlled browser
runtime, not live server/SQL/mint/user Exit/display/full AC12. Whole new span
needs Claude, earlier spans independent. RP-283/RP-284 display/preview author
contracts remain; RP-131/Firefox and full1.0 holds remain. Fresh Postgres is
healthy, DB tmpfs1%, but Docker root still100%/39,784KiB free: no cleanup or fresh
SQL proof. Next safe accepted work: remaining R9 component/host states and
advisory-plan consumer grounding, without crossing those boundaries.
[Evidence](reputation-tree-v1/log.md).

Previous runtime checkpoint: RP-284's existing next-run event reader/delivery is
locally corrected in the range after0ae1fa11. Forty-five new cases discriminate;
four compiling omissions fail41/2/1/3. Restored8151 units/255 skips, types/build,
boundaries/copy/manifest pass; Chromium/WebKit134 functional plus one isolated
performance case pass. This is controlled-runtime plumbing, NOT Run End display,
real network/SQL/mint/full AC12. The post-Exit balance and payload-only GU-C3
boundary need author reconciliation, now named in the decision queue. RP-131
executes RED; Firefox remains unexecuted. Full new range needs Claude; earlier
ranges remain independent. Next accepted work: recovered-publication R7
identity/delivery checks, without crossing the unresolved display boundary.
[Evidence](reputation-tree-v1/log.md).

Previous runtime checkpoint: RP-285/RP-286 native Reputation Tab traversal is
locally corrected under2b2536a9. Unchanged-source baseline44 fails/10 controls
pass; corrected54 pass. Three compiling regressions fail36/52/12 as declared,
each exactly restored; final248 selected Chromium/WebKit cases pass. Header,
enabled rows, Confirm/Cancel and reverse traversal execute without browser
preference changes. Types/build8106 units/boundaries/copy/manifest pass;
255 Node skips remain visible. Historical RP-131 guard still RED, Firefox
unexecuted: not whole CI/AC12/manual AT/SQL/mint proof. Whole span after8a3bfb70
through final record edge needs Claude; all earlier ranges remain independent.
Its next-step note is historical: RP-281 is now corrected as above; RP-284
producer/consumer grounding remains. RP-283 remains an author contract gap.
No numeric/kernel/copy/CI policy change, owner adoption, release promotion or
archive. [Evidence](reputation-tree-v1/log.md). Older checkpoint next-step notes
below are historical, not the current queue.

Previous runtime checkpoint: RP-282 hidden Exit plans are locally corrected
underc12e985f. Completed-protocol baseline fails16 remount/Offer cases while24
empty/selected controls pass; initial incomplete Offer fixture is disclosed.
Two compiling mutations independently fail16/24 with controls green, exact
restoration passes194 selected Chromium/WebKit cases. Fresh panel selection now
synchronizes through its existing callback; current selected plans still send
in artifact order. Actual native host/runtime over controlled HTTP/socket input,
not live server/SQL/mint/full AC12. Types/build8106 units/boundaries/copy/manifest
pass;228 Node skips visible. RP-131 guard executes RED, Firefox unexecuted.
Tier0's opt-in eligibility is manufactured for the copy-era diagnostic, not a
reachable production Wind Down; the Tier1 subset proves the same host defect.
Whole span aftereb258a7a needs Claude. Next ground/predeclare RP-285's R9 Tab
header/row path; RP-283's missing authoritative preview bridge and RP-284's Run
End/new-route consumer remain separate. No copy/numeric/kernel/CI change,
owner adoption, mint or release promotion. [Evidence](reputation-tree-v1/log.md).

Previous runtime checkpoint: RP-279 persistent Amount costs are locally corrected
under285a247e. Native unchanged-source baseline fails40; seven compiling cost/
visibility/order/binding omissions discriminate, then exact restoration passes
154 Chromium/WebKit cases (40 cost/58 child/56 host). All four row states, both
eras, published notation and native confirmation/pending/replacement execute.
New cost evidence uses controlled component props, not host/SQL/mint/full AC12.
Types/build8106 units/boundaries/copy/manifest pass;208 Node skips visible.
Historical RP-131 guard is executed RED; Firefox remains unexecuted. Full span
after22e03946 needs Claude. Next ground bounded R9 plan-panel/host lifecycle
coverage: RP-282 stale advisory selection is a source finding, not yet a native
defect proof; RP-283 preview and RP-284 Run End producer gaps remain separate.
No copy/numeric/kernel/CI change, owner adoption, mint or release promotion.
[Evidence](reputation-tree-v1/log.md).

Previous runtime checkpoint: RP-278 host row rejection feedback is locally
corrected under441013a5: unchanged-production native baseline fails48 rejected
purchases/eight applied controls pass; restored host+child114 and shared GS0.2
two selected cases pass. Seven compiling omissions discriminate with exact
restoration. Actual client host/runtime over controlled HTTP/socket inputs,
not live server/SQL/mint/full AC12. Both conflicts refresh; attribution/clearance/
Founder revision execute. RP-280 declares only missing key with pending-owner
placeholder. Copy658/new generated-Go producer tree; prior career artifacts
remain historical. Types/build8106 units/boundaries/copy/manifest pass; RP-131
history stays RED, Firefox unexecuted. Full span afterb8ee639f needs Claude.
Next RP-279 persistent Amount costs. No owner-prose/epoch/kernel/CI-policy or
release promotion. [Evidence](reputation-tree-v1/log.md).
Go vet passes; RP-281 records inherited generated-copy formatting debt, not
formatter-green evidence. A paired generator/output repair needs its own range.

Previous runtime checkpoint: RP-277's Reputation purchasing-row busy state is
locally corrected under294a1b08. Native baseline and four compiling omissions
fail32 new cases each; restored Chromium/WebKit pass58. Both task/parent settling
orders, unrelated refresh and second-row attribution execute. Types/build8106
unit tests/boundaries/copy pass; RP-131 history remains RED, Firefox unexecuted.
Controlled component tasks/arms, not real host/SQL/mint or whole AC12. Full new
span after5bf6017f needs Claude. Next ground RP-278's host rejection/resync/row
path, then RP-279 costs. No copy/kernel/transport/CI policy or release promotion.
[Evidence](reputation-tree-v1/log.md).

Previous runtime checkpoint: RP-276's Reputation purchase focus is locally corrected
under31795a1e. Native baseline and compiling focus-call removal fail16 cases;
restored Chromium/WebKit pass26. Types/build8106 unit tests/boundaries/copy pass.
Firefox executes zero even outside sandbox; historical RP-131 remains RED, so
whole client/CI is not green. This is controlled component proof, not minted
host/SQL/full AC12 acceptance. Full span after020a25c6 needs Claude. Next
separately diagnose RP-277 pending row aria-busy; RP-278 row rejection and
RP-279 persistent costs remain. No balance/kernel/copy/CI change or release
promotion. [Evidence](reputation-tree-v1/log.md).

Previous evidence: RP-263's dated H4/H5 reports are recorded and fully reproduced
from committed6bd24e8d. Each execution covers97 paired careers and970 node-study
arms;676.102/469.043s, reports byte-identical. Complete declarations,51 synthetic
refusals/eight compiling omissions and real overwrite/dirty/selector refusals
discriminate. Retained-artifact and post-artifact cold fast harness pass; vet
passes on unchanged source. Six strict H4 failures and all statistics remain
unchanged: this is reliable negative evidence, not H4/H5 or release acceptance.
Whole new span after5c93ff6f needs Claude, independently of every prior span.
Next ground remaining R9/AC12 client-consumer coverage, preserving author/data/
mint holds. Docker remains100%/39784KiB free; no cleanup or fresh DB/browser/
deployment proof. Earlier H1/H2 keep their earlier producer, not freshness at
this changed tree. Full nine-tier/platform1.0 remains open.
[Evidence](reputation-tree-v1/log.md).

Previous preparation: RP-275 now retains raw H5 observations and shares report
composition with the ordinary checks. Ten malformed groups/eight retained
corruptions/one zero-effect control refuse; twelve compiling omissions fire.
Full97-pair/970-arm study atd96bfaf1 finishes641.626s, retaining/recomposing970
raw observations with unchanged medians/populations. Cold fast/vet pass; H4's
six ties and both historical comparisons stay RED. Whole span after5e083766
needs Claude. Next separately predeclare dated career provenance/full replay;
no new dated artifact, policy adoption, acceptance or release claim here.
[Evidence](reputation-tree-v1/log.md).

Previous observation: RP-272/273 expose the career reports' actual statistical
populations without changing the game or acceptance rules. Full97-pair/970-arm
run atb22ce51e finishes547.497s: Casual H4 has28 finite pairs/six ties,
one treated-only and three both-unreached; H5 cash/tower medians use28/32 and20/23 bought pairs. All27
node/persona groups visible. Cold fast/vet pass; H4 and historical comparisons
stay RED. Full new span aftera3e36f94 needs Claude. Next separately predeclare
dated H4/H5 observation provenance, not censoring or balance adoption. RP-274
and the existing author/capacity/review routes remain. No acceptance, mint,
archive or release claim. [Evidence](reputation-tree-v1/log.md).

Previous correction: RP-263 now distinguishes actual career inputs from catalog
RunKeys and binds/retains sources at both report consumers. Four real careers
preserve complete gameplay fingerprints;36 refusals/nine compiling omissions
discriminate. Full97-pair/970-arm study finishes636.768s:194/970 source admissions
pass; H4 remains93 comparisons/3 exclusions/six ties, both historical comparisons
RED. Cold local fast/core pass. Kernel161/live math/balance/CI/corpus/reports unchanged;
full span afterdfc2fb8c needs Claude. Next RP-272/273 denominator diagnosis;
RP-274 per-node attribution, H3 and RP-268/271 remain author routes. Docker
rechecked100%/39784KiB free; no cleanup or fresh DB/browser claim. Full1.0 remains
open, no box/mint/archive/push. [Evidence](reputation-tree-v1/log.md).

Previous evidence: RP-263 now has fresh dated H1/H2 reports and committed-producer
lineage. Full97-run/679-clock record and byte-identical replay pass; all30
thresholds are measured with exact pinned Prestige. Proposal set unchanged,
three Reference rows differ; none adopted. Twenty-four refusals/four compiling
omissions discriminate; cold local fast/core gates pass. Historical files,
kernel161/live math/balance/CI/corpus unchanged. Full span afterc352370a needs
Claude. Next H4/H5 report-envelope provenance audit, preserving fired criteria
and RP-268/RP-271 author boundaries. H3/H4 failures, H5 gaps, SQL/capacity,
all prior reviews and full1.0 remain. No box, mint, archive, push or release claim.
[Evidence](reputation-tree-v1/log.md).

Previous harness correction: RP-270's shared H3 oracle now rejects aggregate
source-coordinate mismatches. Valid-fixture baseline and compiling omission
discriminate all four fields; H2 retains independent defense. Cold31 H3/45 H2/
28 H4 diagnostic children plus legacy checks, fast/core/vet pass. Full five-arm
485-run/3395-clock study completes313.732s: neutrality passes, tiny changes0/679,
strong289; exact H3 criterion remains RED. Test-only, kernel161/live math/balance/
reports/corpus/CI policy unchanged; full span after0a486d79 needs Claude. Next
fresh H1/H2/report-lineage work. RP-268's missing artifact now also routes to
RP-271's each-Exit vs scripted-first-no-plan wording, requiring author body
reconciliation before policy changes. H3 author criterion reconciliation,
H4/H5, SQL/capacity, all prior reviews and full1.0 remain. No waiver, retune,
box, mint, archive, cleanup or push. [Evidence](reputation-tree-v1/log.md).

Previous harness correction: RP-269 admits exactly the ordered first-hour Exit
pair and binds the pinned H2 caller to declared full-study source/population
and aggregate coordinates. Forty-five diagnostic children, legacy threshold/
H1/H3 tests and six compiling omission probes pass/discriminate; exact source
restoration and unchanged retained30-point report verified. Cold fast/core/vet
pass. Generic legal subsets still calculate. Harness-only, kernel161/live math/
balance/reports/corpus/CI policy unchanged; full span after9f5b81a0 needs Claude.
Its original next queue is RP-270's separately observed H3 aggregate-source gap, then fresh H1/H2/
report lineage and RP-268 artifact-authority grounding. Historical recalculation
is not fresh production or calibration; H3/H4 failures, H5 gaps, SQL/capacity,
all prior reviews and full1.0 remain. No box, waiver, retune, mint, archive,
cleanup or push. [Evidence](reputation-tree-v1/log.md).

Previous harness correction: RP-264's H4 report gate now refuses false exclusions,
invalid clocks and inconsistent/missing savings; each census is fresh. All28
synthetic profiles and the legacy gate pass cold. Nine compiling omissions
discriminate; two expose caught panics rather than semantic refusal. Cold fast
harness/core/vet pass. Complete97-pair H4 and970-arm H5 finish in787.881s but
remain RED (six Casual ties and retained-report drift). A separate current
97-pair census confirms93 counted comparisons/3 valid exclusions, Casual4/14/21.
Harness-test-only; kernel161/live math/balance/reports/corpus/CI policy unchanged.
Whole new span afterdb8398a3 needs Claude. Its original next queue diagnoses RP-269's threshold
source/Exit admission and ground RP-268's named artifact authority; fresh
H1/H2/report lineage, H3 author reconciliation, H4/H5, SQL/capacity and all prior
reviews remain. No box, waiver, retune, mint, archival, push or 1.0 promotion.
[Evidence](reputation-tree-v1/log.md).

Previous harness study: RP-262 now executes the current first-hour producer.
All485 runs/3395 clocks complete; epoch8/no-row/unit neutrality passes, but
the tiny factor changes0/679 clocks and strong control changes289. The tiny
input changes lifetime production, not milestones; exact H3 criterion RED.
Four actual Reference controls and27 oracle/refusal children pass; five
compiling omissions discriminate. Cold core/vet/fast pass. Harness-only;
kernel161, live math/balance/reports/corpus unchanged. Full span afteree1064a0
needs Claude. H3 author reconciliation, report lineage/
fresh H1/H2, H4/H5, SQL/capacity and all prior reviews remain open. No waiver,
retune, mint or full1.0 promotion. [Evidence](reputation-tree-v1/log.md).

Previous harness correction: RP-267 now carries canonical lifetime accrual through
Reference candidates and waiting. Six reproduced actual-branch omissions are
corrected;18lifetime/16frozen-input observations and six compiling omission
probes discriminate. Actual Reference seed0 lifetime rises to1.4605083614e6/
3.54431965065e6 at its two Exits; paid2/0, purchases and all observed clocks
stay unchanged. Cold core/vet, fast harness and client8106/134/types/build plus
separate boundaries pass; composite remains RP-131 RED. Complete H4/H5 finish
in836.130s but remain report-drift RED with the same six Casual ties. Harness-
only; kernel161, live math/balance/reports/corpus unchanged. Full span after
973d983c needs Claude. Next RP-262 H3, then RP-264; H1/H2/report provenance and
fresh reproduction, SQL/capacity, all prior reviews and full1.0 remain open.
[Executed evidence](reputation-tree-v1/log.md).

Previous harness correction: RP-265 now threads frozen Founder inputs through the
Reference planner and actual waiting; RP-266 covers both simulation rate seams.
All16Reference observations,20producer profiles and six boundary controls pass;
nine compiling omissions discriminate and restore byte-exact. Cold core/vet,
fast harness and client8106/134/types/build plus separate boundaries pass.
Composite client remains RED at RP-131. Kernel0.3.161; no live formula or balance
change. Reference seed0 treated gate moves355000→357000ms, control stays357000ms.
Full H4/H5 reject unchanged v1 report drift in786.459s; H4 still fires the same
six Casual ties. Complete new span afterd18d4e09 needs Claude. Next: measure
RP-267's Reference lifetime-hook lead before RP-262 H3, then RP-264 exclusion
oracle; report provenance/refresh, actual SQL, all earlier
reviews and the full nine-tier/platform/release floor remain open.
[Executed evidence](reputation-tree-v1/log.md).

Previous harness correction: RP-263's helper/run-key portion now names the complete
paired tree/economy fixture rather than the base epoch. Four baseline failures
are corrected; seven controls pass and three omission probes fail as expected,
then restore byte-exact. Cold fast harness/core/vet/topology pass. Full H4/H5
reproduce unchanged in840.123s; H4 remains FAIL. Kernel160, balance, reports and
live runtime are unchanged. Complete new span after6963692b needs Claude.
Report-envelope provenance remains open. Next: execute RP-265's Reference
planner frozen-input lead before RP-262's H3 witness, then RP-264's exclusion
oracle. Actual SQL/capacity, independent reviews and the full 1.0 scope remain.
[Executed evidence](reputation-tree-v1/log.md).

Previous harness correction: RP-261 now refuses malformed career policies and
unknown node exclusions before simulation. Six refusals/four legal controls
pass after four baseline admissions; both guard omissions discriminate.
Cold fast harness/core/vet/topology pass. Exhaustive H4/H5 reproduce unchanged
in922.866s; H4 remains FAIL, not release evidence. Harness-only, kernel160 and
balance/runtime/reports/CI policy unchanged. Full new span after7da200f0 needs
Claude. Its original next queue is refined by the fixture/planner dependencies
above; SQL/capacity, independent reviews and 1.0 remain open.
[Executed evidence](reputation-tree-v1/log.md).

Latest Company replay proof: RP-260 under4d4846f2/f7822250/d23da963 retains
two complete source-bound Company runs: actual starter genesis/frozen1.003 and
a synthetic unit control. Both public verdicts and full terminal states pass;
18 corrupted inputs, four population controls and two false heads reject.
Six compiling omissions discriminate, with exact restoration; cold core/vet
pass. Test-only, runtime/kernel160/original corpus/schema/CI policy unchanged.
Full new span after721c0ee1 needs Claude. No SQL/AC15/browser/whole-CI claim;
Docker rechecked100%/39784KiB free. Next safe work: remaining Reputation harness
consumer audit, without changing measurement policy or loosening H4/H5.
[Executed evidence](reputation-tree-v1/log.md).

Latest portable proof: RP-259 under e3378e96 verifies 24 source-derived Founder
histories and rejects 144 corruptions plus four population controls. Compiling
comparison omissions discriminate; independent sequence defenses survive alone.
The c7c7a0fb first-filter refinement adds the previously missing paired-Exit
control: omitting its count check fails that control; omitting the whole
population validator fails all four. Historical three-control runs are retained.
Cold core/vet pass; runtime/kernel/corpus unchanged. New span after b6c7ebca
needs Claude. No SQL reader, two-Exit career or full R8 claim; the separate
Company-run supplement is RP-260 above. Docker remains100%/39784KiB free.
Next: remaining harness
consumers. [Exact evidence](reputation-tree-v1/log.md).

Latest reader correction: RP-257/RP-258 under336503b3 refuses inconsistent
pinned ownership/mirrors in server snapshots and noncanonical/below-one client
factors. Red-first populations and three compiling guard omissions discriminate;
cold core/vet and client8106/134/types/build/boundaries pass. Composite still
fails at RP-131; no fresh SQL or browser proof. Kernel0.3.160 unchanged. Complete
new span after4030c895 needs Claude; remaining pinned/history/harness consumers
are next. No archival or 1.0 promotion. [Evidence](reputation-tree-v1/log.md).

Latest test preparation: RP-253 under5425e684/373a6e1a/2e9cc4c4 retains the
complete twenty-case direct-purchase source plus an overdue-Fiscal rejection.
All21local profiles pass; three population corruptions reject and fail when
their validator is omitted. The SQL lane also prepares42unrecorded conflicts,
exact retries and stored-state/log/event/outbox verification, but explicitly
skips without Postgres. Cold core/vet pass. Docker remains100%/39784KiB free;
no new workload or cleanup. Actual SQL and persistence controls stay mandatory.
Runtime/kernel/corpus/CI policy unchanged; complete new span after86c96035
needs Claude. Next safe work is remaining pinned reader/writer audit, not
archival or release promotion. [Evidence](reputation-tree-v1/log.md).

Previous runtime correction: RP-255 under761462fc/4ca6dc7d enforces the existing
six frozen purchase-input fields before Go struct decoding. All570raw malformed
cases now refuse after376baseline admissions;19valid results remain unchanged.
TS already refuses456malformed objects and preserves114parsed-duplicate controls.
Go/TS gate omissions and source-hash/population/row corruptions fail independently,
with exact restoration. Kernel **0.3.160**; cold core/vet and client/types/build
plus separate boundaries/topology pass8047/134. Full verify-client remains RED
at RP-131/50a3a514. Native Chromium/WebKit pass590each; Firefox never executes
tests, including an approved narrow escalation (RP-256). No whole browser/CI pass.
Complete new range after4d690f29 needs Claude. RP-253 preparation is above;
remaining pinned readers are next; actual DB execution is still held at100%/
39784KiB free. No cleanup, restart, old-corpus/schema/migration edit, mint/push or full
B4/R8/RFC/player/1.0 promotion. [Exact evidence](reputation-tree-v1/log.md).

Previous verification supplement: RP-254 undereddddc61 adds explicit result-pin
comparisons for20direct purchases and3paired Founder Exits in both runtimes,
using unchanged Go-authored fixture pins. Four hash faults and four missing-
observation controls fail; the older direct state/receipt/event comparisons
alone survive their wrong-hash fault. The Go Exit fault is a test-boundary
oracle probe, not a runtime defect. Cold server-core/vet passes; client/types/
build and separate boundaries/topology pass7457/134. Full verify-client remains
RED at historical RP-131/50a3a514. Runtime, corpus and kernel0.3.159 unchanged;
complete new span afterc355fb7d needs Claude. Its RP-255 follow-up is recorded
above; remaining pinned readers and RP-253 actual persisted rejection taxonomy,
the capacity hold and full R8/B4/CI/player/1.0 obligations remain.
[Exact evidence](reputation-tree-v1/log.md).

Previous runtime correction: RP-252's Reputation purchase-event validator now
refuses all55malformed decision fixtures, including22previous admissions from
missing/null/ambiguous keys and signed overflow. Six legal controls and eleven
original producer events pass; three independent omissions fail20/7/2 and
restore exactly. Kernel **0.3.159**; cold server-core/vet and client/type/build
plus separate boundary/topology checks pass7456/134. Full verify-client remains
RED at RP-131/50a3a514. New complete range after822774df needs Claude.
Codex's full original B4 review (`541da96e^..541da96e`,17paths) is CHANGES
REQUIRED for RP-252/253/254, not approval of its own fixes: complete persisted
rejection taxonomy remains; the hash supplement above needs independent review. Docker
capacity rechecked100%/39784KiB free; no cleanup or fresh DB workload. No full
R1/R5/R8/B4/CI/player/RFC/release promotion. [Exact evidence](reputation-tree-v1/log.md).

Previous runtime correction: OD-7 append-only-ID admission now rejects removing
any of nine defined Reputation IDs at the linked/live/replay fixture boundaries
(28 negatives per runtime, 56 total). Five independent guard omissions fail;
both runtimes reject an incomplete shared population. Legal retunes, append,
initial activation and inactive history remain valid. The former retirement
positive is retained as a forbidden transition, with three legal boundary
expectation rows unchanged. Cold server-core/vet passes; client/type/build and
separate boundaries/topology pass7456/134, kernel **0.3.158**. Full verify-client
remains RED at historical RP-131/50a3a514. Complete new range after `cf458b2b`
needs Claude; no full R1/OD-7/B4, actual DB/player, CI or release promotion.
[Exact evidence and next work](reputation-tree-v1/log.md).

Previous runtime correction: RP-251's valid higher/lower next-epoch unlock effects
no longer fail on a stale mirror. Go live/Founder and TS Company/Founder rebind
only the mirror, preserving historical spend/ownership. Twelve cases pass;
four independent omissions fail their affected retunes and restore exactly.
Cold server-core/vet and client/type/build/boundaries/topology pass 7422/134;
kernel **0.3.157**. The new span after `8ab5a2c4` needs Claude. Full verify-client
was rerun and is RED at historical RP-131/50a3a514. At this earlier checkpoint
append-only-ID admission was still unexecuted; the correction above supplies
its bounded fixture evidence. Original full B4, actual DB/player, whole CI and release obligations remain;
retirement fixtures do not authorize removing nodes. [Exact evidence](reputation-tree-v1/log.md).

Previous command-admission correction: RP-250 now rejects pinned-invalid Reputation input
at the live purchase resolver and shared Founder command boundary, and rolls
back invalid successful output against the resulting epoch. Seven formerly
admitted corruptions reject; four valid/retired/inactive controls pass. Three
independent guard removals fail and restore exactly. Cold server-core/vet and
client/type/build/boundaries/topology pass (7416/134); kernel **0.3.156**.
The complete new span after501000ea needs Claude; full CI, actual DB admission,
complete R1/B4/AC11 and release acceptance remain unproved. RP-251's cross-epoch
append-only-ID observation was static at that checkpoint, not an owner ruling.
The subsequent corrections above implement existing R1/OD-7, not new refund
policy. [Exact evidence and next work](reputation-tree-v1/log.md).

Current strategic status lives in [the 1.0 board](roadmap-1.0.md) and its
[append-only checkpoints](roadmap-1.0-log.md). Current implementation authority lives in
the [active RFC index](../rfc/README.md), each RFC's plan/log and the
[executable queue](platform-alignment/execution-queue.md).

Latest R-011 primitive wave completes 120 intervals / 30 actual page closures
in the three declared Linux engines; exclusive-bypass and watchdog controls fail,
restored script passes. [Dossier](platform-alignment/browser-session-coordination.md)
does not claim real refresh/commit ambiguity, storage policy, native lifecycle,
automatic renewal or mature Garden. New span begins after `39f95329`, includes
`d023dc26`; Claude review pending independently of previous spans.
**Latest local browser repetition is INCOMPLETE, not green:** 3595 reports
299/300 file populations before stalling; Firefox reports no device space and
Docker's writable filesystem is 100% full (RP-236). Stopped at the existing
ten-minute CI ceiling, exit 137; no final denominator/performance. Prior 93957
passes 300/300 / 22482 tests / six existing skips plus performance, but earlier
Route import RP-235 and worker RP-218 recurrence remain open. Intentional fixture
and worker HTTP 404 controls discriminate; no production fix. See the
[retained observation dossier](platform-alignment/browser-module-research.md).
Root client/type/boundaries/topology passes and historical green runs are not
whole-CI/reliability closure. RP-131 remains separate. No unrelated Docker cleanup.

Accepted Account/API groundwork continues without Docker: the
[refresh parser/router census](platform-alignment/session-refresh-contract-census.md)
pins 27 actual in-memory requests and three restored failing controls. Cold
Account/publicapi tests/vet pass; deliberately no database or successful/reused/
expired-token rotation proof. Null/duplicate/case-insensitive parsing is observed,
not approved policy. New test-only span includes `d3e7196f`, pending Claude
independently; renewal draft, full R-011/Garden/1.0 obligations remain open.

Real refresh persistence stage under d33b5d56 is **prepared, not executed**:
four tests/seven DB populations compile, host explicitly shows dependency SKIPs.
No fresh Docker workload on full disk, no successful rotation/replay/expiry gate.
Actual populations/four severings/designated review remain pending; see census
dossier and Account plan/log. No contract or release promotion.

The separate manual completeness observer now rejects an actual Go exit-zero
run with four dependency skips and zero/seven cases (RP-238). Twenty synthetic
controls pass, not DB evidence. Its normal Postgres mode and the prepared
rotation/replay/expiry/severing populations remain unexecuted pending capacity.
Broader client verification caught/fixed the new Node/Vitest fixture filename
collision (RP-239), without config/test exclusions. Restored client population
passes 7366 tests/134 existing skips; type/build/boundaries and remaining gates
pass separately. `verify-client` remains RED at unchanged pushed RP-131, not
green CI. Actual repaired-source missing-DB control still rejects skipped pass.

Separate accepted API C1/C7/C9 compiler census now proves RP-240: fourteen
OpenAPI/TS rows match but exclude the existing refresh route; actual compiler
rejects its path/two errors. In-memory path/error counterfactuals discriminate,
twelve arms. Four scoped fetcher sites remain outside generated. api-check and
cold publicapi/client/type/topology pass for the subset, not complete contract,
dispatcher, DB/renewal or CI. [Dossier](platform-alignment/refresh-generated-contract.md),
new span after 29e1ff02, Claude pending. Capacity/policy/prior ranges remain open.

Codex's designated cross-party review now approves Claude's exact
`527246f1^..527246f1` Clout guard addition only, not full P6/AC9 or archival.
Independent first-hour guard removal produces unsupported neutral completion;
the witness fails. Phase-0 typed-refusal and scalar no-op probes also fail;
negative-guard-only removal survives the independent AxisInput rejection.
Restored cold default harness/selected vet pass (40.486 s). RP-241/D-021 routes
the existing DG-D evaluation-contract choice; no substitute attainment is adopted.
RP-242 corrects fixture-presence versus served-journey claims and distinguishes
artifact-pairing rejection from runtime's all-axis refusal. Existing server
projection test passes, not fresh browser or minted PR evidence. See the
[exact verdict and remaining work](clout-v1-and-pr-interns/log.md).
No production change, Docker deletion, balance/CI/Clout or release promotion.

Latest independent runtime correction is Reputation RP-243 under dc9e6fc6 /
8379f96d: Go accepted twenty cross-arm zero/empty/null starter fields that TS
rejects. Go now checks the exact raw arm before shared struct decoding; all
three kernel identities move to **0.3.154**. Original failure and restored Go/TS
severings are retained. Full root server-core/vet and client/type/build/boundaries
pass; root client composite remains RED at RP-131. No real DB/browser/full
Reputation or minted player-career claim. [Log](reputation-tree-v1/log.md) binds
the new Codex span after 6947ebe3, independently pending Claude. Original B1/B9
and prior ranges stay open; no copy, balance, migration, API, CI or owner choice
was changed. Docker remains 100%/39,784 KiB available; cleanup answer pending.

Reputation R1/R3 test-only wave under ce6a9d6b now executes sampled 147
parameter triples/462 spend cases/eight invalid tuples in both pure runtimes,
plus the ten canonical Go/TS vectors. Four independent earned-base/overspend
probes fail, also firing existing TS new-run replay and Founder restoration
assertions; exact source restore and final cold Go/vet/client/type/build/
boundary/topology pass (7369 client tests/134 existing skips). Kernel stays
0.3.154. Codex approves only the original a522fdf1 Available/BonusFactor
properties, not full B1/R1 codec/R3 frozen rows/Reputation. New test/doc/record
span after 5467f575 needs Claude separately; no real DB/browser/whole-CI or
mint proof. RP-244/[log](reputation-tree-v1/log.md) retains exact limits.

Reputation R4/AC8 now has an independent full four-path verdict on Claude's
original `7d130b89^..7d130b89` witness addition only. New shared Exit cases cover
all four starter nodes, cumulative cash/upgrade, retired IDs and tree-ordered
emitted IDs. Six assignment/upgrade/order mutations fail in Go/TS, sources restore
exact, and original-v9/current-v12 consumers pass. Cold full server-core/vet and
client/type/build/boundaries/topology pass; final client 7374/134. RP-245 and the
[log](reputation-tree-v1/log.md) retain remaining cap/idempotency/next-tree and
full producer/persisted-row/mint/default-player debt. Canonical staging/version
contradictions are reconciled, not new behavior. New Codex span after e365e0da
still needs Claude, distinct from the original witness verdict. No runtime/kernel,
owner/CI policy, archive/push or full Reputation/1.0 promotion.

Remaining R4 boundaries now execute under e5a8f7f3: four hash-derived fixture
next bundles cover retirement, already-owned upgrade and exact generated/permit
caps. Raw over-cap loaders and runtime guards reject; all eight source controls
fire and restore. TS defensive inputs are intentionally fault-injected parsed
copies, not admitted/persisted pins. Full cold server-core/vet and client/type/
build/boundaries/topology pass (7383/134); kernel/runtime/balance unchanged.
RP-247/[log](reputation-tree-v1/log.md) retains exact scope and first-filter-only
provenance; new Codex span after70f8c8c8 needs Claude separately. R1/R7 and
persisted rows/career/default-player/mint/full-RFC acceptance remain open.
RP-246's curriculum grammar observation is static/unexecuted and out of this
runtime scope. No full CI/1.0/archival or Docker cleanup authority inferred.

R1/R7 inspection reproduced RP-248: TS encoding admitted all 16 invalid
accounting/legacy-state inputs. The bounded correction rejects them, preserves
valid replay bytes, and catches both independent removals (13/3 failures).
Server-core/vet passes; client type/build/tests pass (7403/134), while full
verify-client remains RED at RP-131/50a3a514. Kernel is 0.3.155. New span after
6a8ccaf1 needs Claude. RP-249 reopens the required four-case R7 migration corpus
and baseline ratchet: existing unit tests are not that gate. Full 1.0 scope stays
unchanged; no archival, push, live DB or release-readiness promotion.

RP-249's actual four-case shared corpus and baseline15 now execute in Go/TS,
alongside eleven byte-unchanged legacy rows. The modern arm binds its referenced
source by SHA; both Exit paths preserve earned Reputation on activation and
the three corrupt inputs reject at their declared boundaries. Nine runtime
controls and three corpus corruptions (both lanes) fail and restore. Cold
server-core/vet and client/type/build/boundaries pass (7408/134); no runtime,
kernel, save schema or migration-body change. New span after39912364 needs
Claude. The subsequent full original B3 review records CHANGES REQUIRED for
RP-248/249 without approving Codex's later corrections. Seven older writable Go
sources (14,16–21) now activate via live new-run and public Founder Exit replay,
agree in full encoded Founder bytes and preserve earned Reputation. Four controls
fail; restored-source server-core/vet and client/type/build pass7408/134. The new
test-only range `80519365..29ed56e8` needs Claude. The subsequent Go-authored
seven-row/nine-bundle corpus now executes strict TS artifact loading/public
Founder replay with full state/receipt/event/hash equality. Three TS runtime
controls and three corpus corruptions in both lanes fail and restore exactly.
Cold server-core/vet and client/type/build/boundaries/topology pass7416/134;
supplement after29ed56e8 needs Claude. Complete pinned codec/AC11 and platform/
CI/owner holds remain; no runtime/kernel change or multi-event/DB claim.

Garden is fixture-first, with no public epoch pin. G1–G7 are implemented, not designated-accepted.
Latest SG7/SG10 correction (RP-232) supplies the missing Fiscal DOM purchase row,
using existing copy. Actual composed bootstrap/Fiscal harvest/purchase/plant/uproot
and timed reads through tick two execute against real server/Postgres/WebSocket.
The unchanged fifteen-minute journey times out before maturity: full harvest and
reload are NOT proven. RP-234 records absent browser session renewal; exact HTTP
status at that timeout was not emitted, so its causal link still needs measurement.
RP-233 retains an intermittent next-plot interaction. Fast Fiscal browser CI
regressions discriminate actual row removal in every engine. Full Linux browser/
performance and root client/type/build/topology pass; native Mac Firefox launch
fails with sandbox/framebuffer diagnostics. Historical kernel RP-131 remains red.
This Garden range needs Claude review and does not include any authentication fix.
Next safe work is separately scoped Account D2 / Transport consumer integration.
That separate diagnosis now executes: `make diagnose-browser-session` observes
actual runtime Garden HTTP 401/socket closure under exact early access-row expiry,
zero automatic refresh, then one test-operated rotation/reload restoring the same
Founder/Garden without gameplay mutation. A valid-token counterfactual fails the
401 oracle. This does not retroactively establish the original timed HTTP status
or implement renewal. `rfc/browser-session-renewal.md` is explicitly draft; R-011
predeclares browser coordination/ambiguity research before policy adoption/code.
Current local client/type/boundary/topology and full Linux browser/performance
pass (22482/six existing skips); historical CI and native-host caveats remain.
New range starts after `b68190f9`, independently pending Claude; all earlier review
ranges and the failed mature Garden objective remain unconsumed.
Current SG10 host correction (RP-231) invalidates mounted Garden on nonterminal streamed
receipts, not only the main snapshot; terminal-screen protection remains unchanged.
Nine actual host/browser-runtime cases per engine cover five DOM commands, Founder revision
and exact payload binding, pending/nonoptimistic state, refreshed next revision, refusal,
visibility and streamed reread. Actual callback and receipt-key severing fail and restore.
Full local browser/performance and client/type/build/boundary/copy/topology pass; historical
kernel guard still fails RP-131. HTTP JSON/socket frames are injected, not real Go/Postgres/
WebSocket/native input/mature progression/public activation proof. Claude review and all
earlier exact ranges remain pending; full G7/AC13/Garden/1.0 acceptance is not claimed.
Prior SG10 timing correction (RP-230) removes an arbitrary one-second minimum that delayed
near-tick views. Seven mounted-component/browser-adapter cases per engine and actual timer/
visibility mutations discriminate; full local browser/performance and client/type/build/
boundary/copy/topology pass. Historical kernel guard remains RP-131 red. This is controlled
timer/visibility/JSON evidence, not real-server/default-host receipt or full AC13/Garden proof;
Claude's exact-range review and every earlier range remain pending. Kernel stays 0.3.153.
Bounded catalog/codec/clock/tick/harvest/command reviews have supplied corrections and
discriminating witnesses; their exact Codex ranges still require Claude's cross-party verdicts.
The current G2/G3 review covers transitive bundle admission, the actual New-Founder initializer,
both real Service/Postgres Exits carrying nonempty permanent Garden, and Go/TS replay.
RP-222 records an unruled starter-retune/unchanged-carry conflict; no grant policy is invented.
The latest bounded SG9 correction (RP-223, kernel 0.3.153) uses a database-stamped actual
saved head for read-only growth projection, ignoring handler time. Its real-Postgres clock,
non-mutation and error populations and restored clock/clone probes pass, along with cold
Go/client/three-engine/performance gates. Claude's designated review remains pending;
this does not claim full G6, default player flow or Garden acceptance.
The latest SG10 surface correction (RP-224) ignores older success/failure completions after
a newer receipt refresh. Four ordering arms and actual guard mutations discriminate. Native
navigation/action cases retain all direction/edge/focus checks after a separately measured
test-method decomposition; final local three-engine/performance and client gates pass.
Two earlier full browser deadline failures remain recorded as RP-225, including a recurrence
of RP-218. No timeout/skip/CI change, general reliability closure or full G7/AC13 claim;
the new Codex correction also requires Claude. Kernel stays 0.3.153.
The latest G4/G6 test-only supplement (RP-226) now exercises the active Garden read/ordinary
commands through actual Compose, HTTP Account/token binding, registry and Postgres, with
both account isolation and real automatic Fiscal funding. Actual producer detachment and a
schema-valid secret leak fail; twenty final repetitions and related Go/Postgres gates pass.
This consumes the exact fixture-only bundle, not public launch content or a mature HTTP
payout/default DOM/real idle-time workflow. New range awaits Claude independently.
The subsequent SG6 rollback supplement (RP-227) checks all ten exposed transaction fault
checkpoints against complete affected rows, including Founder genesis/outbox and real
retention deletion. Actual early commit fails all ten; restored code passes 200 fault cases
and related DB/Go/vet checks. It is test-only, independently pending Claude, and not full
AC8/G5 or separate fault coverage after every individual SQL statement.
The latest SG6/SG8 replay supplement (RP-228) compares Company replay to its complete saved
head and checks exact Go/TS hash-verdict outcomes plus actual Founder-history corruption.
Unrecorded replay cash and both hash-guard removals fail; restored related DB/Go/client gates
pass. It is independently pending review, not full AC9/G5. RP-229 separately requires the
specification author to reconcile the unnamed additional "ordinary resource event".

Remaining work includes Garden's broader integration/public-flow/copy/activation gates, other
active RFC/review lanes, later tiers/endings/MMO consumers, account rights/retention, task-level
accessibility, packaging/content rights and exact-artifact clean-host Deployment rehearsal.
Historical kernel-guard RP-131 remains open. Local/native ARM Linux checks do not prove hosted
amd64 or whole-CI success; RP-218's original cause/hosted reliability remains separate.

Nothing here authorizes archival, push, publication, deployment or release. The initial
capability audit is a dated baseline, not a fresh measurement of all subsequent implementation.

## Historical August checkpoint — not current HEAD

The remainder preserves the 2026-08-21 navigation snapshot and verdict references at `440dbba`.
Its READY, missing-implementation and hosted-green statements describe that checkpoint only;
use the current links above. In particular, Game UI has since closed and Deployment has moved
past draft/no-implementation state; neither change establishes full 1.0 or clean-host acceptance.

This is a navigation brief, not a second execution queue. Current authorization lives in
[`platform-alignment/execution-queue.md`](platform-alignment/execution-queue.md), active RFC state
in [`../rfc/README.md`](../rfc/README.md), and implemented behavior in [`../docs/`](../docs/).

### Historical product status

Cloud Clicker is a development snapshot, not a 1.0 release. The owner-selected next milestone is a
bounded **Phase-0 Playable Preview** covering the audited T0–T1 vertical slice; its exact release
manifest is not yet accepted.

Substantial implemented foundations include:

- cross-runtime numeric, economy, save, production and offline-progress kernels;
- account/session, Postgres persistence, WebSocket transport and recovery;
- route, replay, epoch, Commons, guild, achievements, meters and pet-care foundations;
- one real minigame tenant (The Pitch), T0–T1 content and run-2 starter consequences; and
- a Svelte browser path that reaches the Desk through the real gameserver, Postgres and WebSocket.

This does not prove a complete first-hour browser workflow. Tiers 2–8, most minigames, combat
engines, player-facing social/world/event systems, production packaging and the designed endings
remain incomplete.

The evidence trace is in
[`platform-alignment/capability-map.md`](platform-alignment/capability-map.md) and
[`platform-alignment/capability-reality-audit.md`](platform-alignment/capability-reality-audit.md).

### Historical closed evidence work

- Q-001 Account witnesses are designated-approved at `34d04a5`.
- Q-002 Minigame API witnesses and the separate API registry tightening are designated-approved at
  `bfd9b65`.
- Q-003 Transport recovery is designated-approved at `249719c`.
- Harness Observability is implemented, designated-approved at `96a574d`, and archived at
  `edafd82`. Its complete local observation took 974.510 seconds, of which 739.441 seconds were
  standard pacing and 212.927 seconds active relevance.
- D-002 repository publication/shared-memory disposition is complete at `7ad5b54`. The
  `e44e1a6` verifier rejects forged publication states and passes from a fresh clone with 11 public
  and 56 private Class-C dossiers, three ignored duplicates and seven ignored diagnostics.
- Copy Amendment A1's owner record now reaches its exact cross-party closure (`3773510`), canonical
  Soul docs match the current epoch (`7d21484`), the root boundary gates honor the repository-local
  Go cache (`ab05f2d`), and four completed maintenance threads moved out of the live namespace
  (`b990b91`).
- Prestige D2–D5/P3–P5, D-012 and AC2–AC6's literal witnesses are designated-approved at
  `440dbba`; historical full-span archival provenance remains separate.
- Leaderboards D2/D4/D5/D6/L1/L4 and canonical docs now distinguish its strong backend from the
  absent player capability. The missing readers/browser/validator/Route surface has a draft
  successor, not implementation authority. Commons any-membership is now implemented and
  severing-proven; AC1 exact ranking, AC3 composed crossing replay/projection and AC6
  frozen-history witnesses also discriminate. The complete repair range is designated-approved at
  `440dbba`; the absent player surface and historical archival union remain separate.
- Fiscal's frozen RFC and archived plan agree with rejection rollback; its record-only repair is
  designated-approved at `440dbba`.
- Game UI's stale full-script browser requirement and fixed-era text are reconciled to the accepted
  narrowed control proof and authoritative T0/T1 era mapping; snapshot-v3/control implementation is
  now READY.

These closures do not make their broader parent RFCs archival-eligible; the exact remaining body,
consumer, lifecycle and range-union blockers are in the execution queue.

### Historical release blockers

The repository still lacks a coherent release floor:

- no supported production Docker/Caddy bundle, clean-host proof, rollback flow or backup/restore
  rehearsal;
- no player-facing account recovery, export or deletion workflow;
- incomplete accessibility evidence for keyboard, screen-reader, zoom/reflow and touch workflows;
- no accepted production observability, operator, incident or complete retention contract;
- the CI blocking lane is hosted-green, while its non-blocking maintenance/archival record remains open; and
- several active RFC bodies and canonical docs still contradict accepted rulings or current HEAD.

The complete defect and decision populations are
[`platform-alignment/backlog.md`](platform-alignment/backlog.md) and
[`platform-alignment/decision-queue.md`](platform-alignment/decision-queue.md).

### Historical execution posture

Prestige, Leaderboards and Fiscal audit-remediation ranges are designated-approved at `440dbba`.
That verdict does not manufacture broader historical archival coverage or accept deferred player
surfaces.

The amended blocking CI population passed hosted in run `32518522514` at `aa27705`, with the
slowest job completing in 2m57s. Its maintenance observation and exact review union remain CI
archival work, not a product-work blocker. Other product lanes require their named
owner/ruling-author reconciliation or decisions before implementation.

Do not infer release readiness from green unit tests, backend primitives or this summary. A shipped
capability still requires its producer, consumer, current data/workflow and discriminating
integrated proof.
