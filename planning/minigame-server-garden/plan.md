# Server Garden implementation plan

RFC: `rfc/minigame-server-garden.md` (accepted 2026-09-25, every OD at its recommended default).
Fixture-first: no production mint (SG13). Numbering is landing-order: the RFC's "Founder v22"
means next-free, and at landing that is **Founder v25** (v22 Reputation, v23 Pet, v24 Cosmetics).
Replay inputs take the next free wire version, and the event migration takes the next free number.

**Current execution, 2026-10-08:** pending/response-focus, connection-state and current-context
read/visibility and native help corrections are locally verified (RP-409/RP-410/RP-420/RP-421/RP-422); designated
review remains open. The latest [owning log](log.md) records exact ranges and checks. Continue
remaining accepted feature gates and consolidated review, not repeated help/loader
diagnostics or the renewal-blocked long-session journey. SG13, RP-222 and browser renewal remain
separate holds; G1–G7 below still mean implementation presence, not feature acceptance.

**Historical bounded instrument work, 2026-10-07:** RP-348 under `a9fadedd` adds
full child-result/elapsed diagnostics without changing any raw input, admission
assertion, loader program, 1000ms guard or 10000ms outer bound. Two real Node
controls exercise pre-entry exception and entered nontermination. Three restored
instrument faults fail 2/1/2 tests respectively; production/corpus remain exact.
Final client 9,816 passes / 637 explicit skips; types/build/boundaries/topology,
native raw-catalog 74 passes / six Node-only skips and isolated performance two
pass. This repairs hidden failure output, NOT the unexplained historical failure's
cause or reliability. No Linux/SQL/hosted/full-Garden claim. Entire new range from
4b44a563-exclusive requires Claude; every prior Garden/UI range remains owed.
Observe any recurrence on its first failure; do not raise/retry-hide the guard.

**Review checkpoint, 2026-10-06:** the checked G1–G7 rows below record implementation presence,
not designated acceptance. Codex's bounded SG1 review found RP-212–RP-214 in the original loader;
the separately predeclared catalog correction and 36-case raw-byte population are locally green
and await Claude's designated review. Kernel 0.3.149 covers this actual admission change, not a
repair of historical RP-131. No Garden archive, SG13 mint or whole-G1/Garden approval is claimed.
Next review lane: the remaining G1 state/clock/engine/commands/corpus, then G2–G7 under bounded
predeclarations. Retain SG2 nullable fields and all original engine/replay/activation evidence.

The separate SG2 codec review under `e8d3f2df` confirms RP-215 (null integer fields default to
zero). Its local correction uses scoped nullable-key lists and kernel 0.3.150, with original
save/replay versions unchanged. Actual codec provenance is G3 `415bea4d`, not G1: the shared
state shape was G1, but its Go strict decoder landed in G3. This is a bounded codec finding,
not completed G2/G3 or whole-G1 review. Eight valid / 21 typed refusal cases run in both
runtimes; two additional raw duplicate refusals are Go-only with visible TS parsing-loss
controls. Both positive/negative probes discriminate and restore. Claude review is required.

Separate test-only SG6 supplement under `3fbf39af`: real Go/TS harvests match 18 independently
derived boundary result/hash/post-state bytes, including the large-product floor and 36-plot
sum. Four restored TS mutations prove arithmetic, seed, hash and post-state checks can fail.
No production/balance/kernel change; kernel stays 0.3.150. This is not AC8's coordinator or
whole-G1/Garden acceptance, and its test range awaits Claude. Next: SG3/SG4 clock/tick review.

Separate RP-217 correction under `14ad4287`: actual Postgres diagnosis shows handler-time
harvest can reject on lag or persist future growth and block later Fiscal commands. SG-P2's
Garden-specific request now takes no timestamp; the DB sample and applicability probe run
under the canonical stream locks. Six skew arms plus retries/hash conflicts/history, two
restored probes and twenty cold repetitions pass. Kernel 0.3.151; cross-party review required,
not full-G5/SG3/SG4 or Garden acceptance. Continue remaining pure clock/tick and G2–G7 review.

Separate SG3 domain diagnosis `c9f449de` / `155486d0` confirms RP-219: codec-valid safe-counter
frontiers advance outside SG2's domain; Go/TS differ on the exact odd counter. Correction
predeclaration `711153cd` refuses before salt/growth mutation using the existing domain, not a
new game cap. Kernel 0.3.152; independent 49-case clock matrix and disabled/late-guard probes
in both runtimes. Every existing checklist/review limitation remains: this is not full SG4 tick,
original G1, G2–G7, public mint or release acceptance. Claude review is required for the new range.

Separate SG4 test-only supplement under `62bbc92a` / instrument correction `24512573`:
sixteen admitted literal transitions and selected PRNG draws agree in Go/TS. Actual-source
diagonal, threshold, absolute-key and recipe-order mutations fail the new witnesses and
restore exactly. RP-220 is evidence, not a production defect or full SG4/G1 acceptance.
The initial inadmissible zero-chance fixtures and empty-selector invocation are disclosed
in the log, never credited as evidence. No runtime/kernel/CI/mint change; Claude review required.

Separate SG5/AC7 test-only supplement under `0c125bec` / instrument correction `9405d75a`:
fifteen direct command outcomes, eight gates and six real-Service/Postgres due-growth arms.
Twenty cold repetitions pass; actual-source precedence, seed-only mutation and transition-
rollback probes fail and restore. Store's persistence guard independently holds when transition
restoration is severed. RP-221 is bounded evidence, not a runtime repair or full G1/G4/Garden
acceptance. Initial wrong Company credit-count expectation is disclosed. Claude review required.
Next: accepted G2/G3 bundle/activation/replay, then remaining G4–G7; no checkbox/lifecycle promotion.

Separate G2/G3 test-only supplement under `d6c2b6a2`, instrument correction `1aa23496` and
reader predeclaration `16b44c74`: eleven transitive bundle removals, admitted no-Garden/hash
controls, actual New-Founder initialization and both real-Service retained-state Exits.
Forty cold DB cases, exact permanent bytes/retries and both correctly scoped histories pass.
TS carries nonempty Garden on both replay axes; actual loader and Go/TS reset mutations fail
and restore exactly. The due Founder Fiscal prefix is required and independently verified,
not suppressed. RP-222's admitted starter-retune/unchanged-carry conflict remains an authored
contract question. Invalid initial fixtures/typecheck remain disclosed. No full G2/G3/Garden
verdict, checkbox promotion, runtime/kernel/CI/mint change; Claude review required.
Next: remaining G4–G7 event/coordinator/projection/player-surface review, without inventing
starter evolution or consuming any earlier unreviewed range.

Separate SG9 diagnostic `9928a54e` / correction authority `307b2539` confirms RP-223:
the real read consumed handler time rather than Postgres time. Original bounded G6 seam is
CHANGES REQUIRED. The new read-only Store operation binds actual same-Founder latest head
and DB timestamp; GardenView ignores handler time and retains discarded-clone/no-write
semantics. Kernel 0.3.153, unchanged API/save/replay shapes and other clock policies.
Twenty repeats of six DB arms, loaded-head/full persistence/error controls and both restored
clock/clone mutations pass, followed by cold Go/client/DB/three-engine/performance gates.
This Codex implementation is ready for Claude's exact-range review, not full G6/Garden or
archival acceptance. RP-222, RP-131, RP-218 and every earlier pending range remain separate.
Next: remaining accepted coordinator/event/read/player-surface review; no checkbox promotion.

Separate G7/SG10 diagnostic `37f76223` / corrected instrument `1bffaab2` confirms RP-224:
all four older-completion arms overwrite newer receipt-refresh state in every native engine.
Separately authorized `0b7a49f9` repair admits only the latest-started request on success and
failure; kernel remains 0.3.153 because the renderer is outside the kernel-watched registry.
Both response-guard mutations fail independently; native arrow severing fails all twelve
refined native cases. Initial type/inherited-disabled/text instrument errors and two full-run
deadline failures (RP-225, with RP-218 recurrence) remain disclosed. Observation and separately
predeclared decomposition preserve all directional/edge/focus and Enter/Space action checks,
without changing limits, skips, runtime navigation or CI. Final local client/type/build/
boundaries/topology and full Linux browsers/performance pass. Earlier cold Go/real Postgres/
corpus/vector/copy gates in this range pass; they are not rerun or relabelled as hosted proof.
New correction requires Claude's exact-range review; no full G7/AC13/default-host/timer/
visibility/Garden acceptance or checklist promotion. Continue accepted G4–G7 integration
review while RP-222 and every earlier designated-review range remain separate.

Separate G4/G6 test-only HTTP supplement under `3e518f21` closes RP-226's bounded missing
evidence route, not Garden acceptance. Actual Compose, Account/authentication, registry,
Production and Postgres consume the exact grown replay fixture. Two HTTP-created accounts
start empty v25/zero-credit; an actual automatic Fiscal event funds unlock. Public plant,
read, identical retry, substrate/uproot, immature/forged-field/foreign refusals, both accounts'
four persisted heads, read-only eight row populations, DB stamps and full Founder replay
pass. Private replay inputs and outbox delivery metadata are explicitly excluded from the
corresponding public/immutable comparisons. Producer detachment fails the test; a schema-
valid actual-salt leak fires its enumerator. Twenty final non-skipped repetitions and the
broader declared Garden/minigame Postgres plus cold Go/vet gates pass. Kernel 0.3.153 stays
unchanged, no product/schema/artifact/CI/mint change. Claude designated review remains required;
no successful mature HTTP payout/default DOM/real idle wait/full G4/G6/Garden claim or checkbox
promotion. Continue remaining accepted coordinator/event/replay/surface integration review.

Separate SG6/AC8 test-only rollback supplement under `d59960af` / `5ca2cf0a` addresses
RP-227: ten independent exposed checkpoints compare ten complete persisted row populations.
Founder genesis is included, and ordinary Service harvests make retention deletion real.
Actual early-commit mutation fails all ten cases; byte-exact restoration, twenty repetitions
/ 200 fault cases, broader related Postgres and selected cold Go/vet pass. Initial snapshot
SQL type error remains recorded. No product/kernel/CI change or checkbox promotion;
kernel 0.3.153. Claude review required. Grouped hooks are not every individual SQL write,
and the preseeded mature fixture does not establish public/HTTP progression or full AC8/G5.

Separate SG6/SG8/AC9 test-only supplement under `f980f4d6` addresses RP-228: Company
replayed final bytes equal actual saved heads for the zero/one/ten-entry helpers, all four
pinned Company shapes have exact Go/TS verifier verdict controls, and actual Founder hash
poisoning diverges in copied history. Actual unrecorded-cash/hash-guard probes fail and
restore, with surviving independent defenses disclosed. Related Postgres, twenty actual
harvest/history repetitions, selected Go/vet and full client/type/build/corpus pass.
Initial selector and canonical-reader failures remain recorded. No product/kernel/CI/mint
change (0.3.153), full AC9/G5 or checklist promotion; Claude review required. RP-229's
unnamed extra resource-event premise needs author reconciliation, not an invented kind.

Separate SG10 timing/visibility diagnosis under `dd4d96f3`, instrument correction
`891dee0d`, then repair authority `32304628` confirms RP-230's one-second minimum
delays a valid 137 ms deadline. Renderer now honors the server-relative remainder.
Seven mounted-component/actual-browser-port controls run in each native engine;
timer-dispatch and hidden-guard mutations discriminate and restore exactly. Virtual
timers, emulated visibility, injected JSON and receipt props are not default-host/real
server/actual idle-wait proof. No kernel/math/schema/copy/mint or checklist promotion;
Claude review required independently of all earlier ranges. Final gates in the log.

Separate SG10 actual-host diagnosis under `9c918361` / `9af68399`, repair authority
`86249481`, confirms RP-231: streamed receipts refresh main state but not mounted Garden.
The host now forwards its existing advisory refresh key inside the unchanged terminal
guard. Nine actual mounted-host/browser-runtime cases per engine cover five DOM commands,
Founder-vs-Company revision/UUID/fields, held receipt/pending/nonoptimistic state, next
revision, refusal, visibility and streamed reread. Actual callback and receipt-key
severing fail and restore exactly. Injected JSON/socket frames are controlled client
integration, not real Go/Postgres/WebSocket/native input/mature progression/public mint.
No schema/kernel/copy/CI change (0.3.153), checkbox or full G7/AC13/Garden promotion.
Exact new range begins `9c918361^` (`3c3af1de`), endpoint pinned in the log; Claude
required independently of earlier ranges. Continue fixture-only composed host/server
integration under SG13, retaining author/owner/CI and complete nine-tier 1.0 obligations.

- [x] G1 — `server/garden` + `client/src/garden`: SG1 loader (every rule, rejecting fixtures),
  SG2 state shape, SG3 advance, SG4 tick, SG5 pure commands, SG6 Founder-side harvest math and
  `harvest_hash`. Go-generated golden corpus replayed byte-for-byte by TS. AC1, AC3, AC4, AC5; pure
  half of AC7.
- [x] G2 — Pin `server_garden` in replay bundles (Go and TS), chain it on `cosmetics` (scalar Founder
  chain), and add the Fiscal cross-artifact rules (unlock row, host generator row). AC2 (bundle half).
- [x] G3 — Founder v25 `server_garden` save codec, activation at a new-run boundary and at
  New-Founder initialization, replay-inputs carry, Exit byte-identity. AC2, AC11.
- [x] G4 — Founder intents `garden_plant`, `garden_uproot`, `garden_set_substrate`; the SG-P3
  advance pre-step on garden commands and `spend_fiscal_credit`; the frozen server-drawn salt;
  events and migration; TS replay; the AC6 trigger-set property test; Postgres witness. AC6, AC7,
  AC9 (Founder half), AC12.
- [x] G5 — `garden_harvest` through `save.Store.ApplyGardenHarvestTransaction` (SG-P1/SG-P2):
  faucet window keyed `server_garden`, Company `credit_garden_harvest`, both logs binding
  `harvest_hash`, fault injection, idempotency. AC8, AC9.
- [x] G6 — `GET /api/v1/garden/current` new operation (discarded-clone projection, no salt).
  AC10, AC15.
- [x] G7 — Garden surface under `client/src/game-ui/` (grid, roving tabindex, non-colour stage,
  reduced motion, 320 px). AC13 stays blocked on the Accessibility RFC's acceptance; AC14 copy.
- [ ] Docs, then hand off for Codex's designated review. Never self-archive.

Separate SG7/SG10 purchase correction under `e02fbfc1` / diagnosis `d24b40f6`,
authority `6ba1252c`, confirms RP-232: actual producer offered Garden but Fiscal
DOM had no row. Existing title/explanation now map to the offered ID. Fast actual
FiscalSurface cases run in all three CI browsers and fail on row removal; no
optimistic ownership or absent/owned/unaffordable purchase. Callback/server-view
severing also breaks the real composed setup, then restores exactly.
The new manual built-client/real server/Postgres/WebSocket journey keeps the
existing real three five-minute ticks. It reaches ticks one/two but times out
at maturity, so harvest/cash/log/reload criteria remain unproven (RP-234).
The runtime never consumes its stored refresh token; route that separate
Account/Transport consumer before the complete journey. Retain RP-233's earlier
next-menu failure and native Mac Firefox launch failure; later setup/CI success
is not a reliability ruling. No checkbox/public pin/AC13/Garden promotion.
Exact new range begins `e02fbfc1^` (`b58277cb`), endpoint pinned in the log;
Claude required independently of all earlier ranges. Final gates are in the log.
