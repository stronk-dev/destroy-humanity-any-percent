# Garage acceptance reconciliation

Source coordinate: `fc911784`, inspected again after predeclaration `124b086a` on
2026-10-07. This is an acceptance inventory, **not** an approval, checkbox flip,
RFC amendment or release certificate. The goal remains the complete nine-tier 1.0.

The accepted authority is [Garage Player Surfaces](../../rfc/garage-player-surfaces.md).
Later accepted producer RFCs do not silently amend its body. In particular,
RP-132 (pet contract) and RP-313 (opportunity sibling arm) still require their
ruling author's reconciliation. Existing implementation boxes do not prove the
acceptance gates below.

## Evidence coordinates and limits

- `G`: `server/gameui/features_test.go`, `projector_test.go`,
  `opportunity_test.go`, `opportunity_review_test.go`, `pet_care_fact_test.go`.
  Fresh root `make test-go GO_PACKAGES='./gameui ./fiscal ./achievements ./meters'
  GO_TEST_FLAGS='-count=1 -v'` exited 0 (handle 51228): 116 PASS lines including
  parent/subtest reports, **not** 116 independent top-level tests. The stored
  rate projector and Reputation public projection SQL tests explicitly skipped
  because `TEST_DATABASE_URL` was not set. No fresh persisted proof.
- `U`: `client/test/game-ui.test.ts`, `garage-surfaces.test.ts`,
  `game-ui-intent-outcome.test.ts`. Last full unit/types/build/boundary run at
  `2d3629ad` passed 9,758 client tests with 442 explicitly skipped browser
  declarations; those skips are not browser evidence. No fresh TS run in this audit.
- `B`: `client/test/garage-surfaces-browser.test.ts`, actual native Chromium
  and WebKit. Last full population at `a42406f0`: 238 functional executions
  and the separate performance observation. Runtime doubles and public snapshot
  fixtures, not real HTTP/server/DB/AT evidence. Compiler-only fault runs do not
  count as discriminating behavioral probes.
- `C`: `client/tools/test-game-ui-composed.mjs`, actual server/Postgres/socket/DOM
  driver. Latest successful execution predates production notice change
  `45bf2edb`; its producer/claim/Fiscal/Pitch observations remain historical,
  **not** a rerun of the changed host. The separate cosmetic composed driver
  recorded real adoption/feed/reload, not raw-care or all GS4 behavior.
- `F`: cold Linux three-engine browser lane at `36c28692` exited 2;
  Firefox imports/keyboard/worker failures and a WebKit worker observation remain
  RP-327/328. Subsequent passive diagnostic stopped with no final population
  verdict after measured Docker overlay 0 available / 100% full (RP-236).
  No third container population or unrelated cleanup is authorized by this file.
- `R`: exact review ranges and executed fault tables in [log.md](log.md) and
  [plan.md](plan.md). Codex's first filters cannot supply Claude's designated
  verdict on Codex work. Original Claude producer/consumer findings and all
  separately owed corrective/record ranges remain live.

## Named GS1–GS6 gates (29)

Each row names the actual producer/consumer and the strongest relevant evidence
found. “Partial” means the row's **whole accepted population is still open**;
it is not a percentage-complete claim. A seed listed by the RFC is not executed
evidence merely because its text exists.

| Gate | Producer → player consumer | Evidence present | Remaining acceptance / route |
|---|---|---|---|
| GS1-A1 projection | `features.go:projectFiscal` → `contracts.ts:parseFeatures` → `FiscalSurface.svelte` | G `TestFiscalArmSweepPreviewMatchesTheNextHarvest`; U v4 Fiscal decode | **Partial:** preview/next harvest comparison is pure kernel/clone, not the specified real Postgres projection/receipt witness. Shared v4 fixture and relevant decoder fault population also need exact proof. Add persisted witness after RP-236 capacity repair. |
| GS1-A2 revision | `GameUIApp.svelte:act` → Fiscal callbacks → runtime | B phase/revision test and held-refresh supplement; actual C Fiscal commands before host change | **Partial:** Founder revision distinguishes Company; notice/HTTP supplements have compiling faults. No fresh changed-host composed run or designated corrective approval. |
| GS1-A3 phases | pinned period → `fiscal-phase.ts` → Fiscal phase text | U exact 99/100/199/200 edges; B safely separated phase samples | B samples are not exact boundary fixtures; verify every accepted edge's rendered text without a fixture clock masquerading as real time. Fiscal clock mint/owner gates remain distinct. |
| GS1-A4 refusals | receipt + mapper → Fiscal own polite notice | B ordinary/cap/unknown/exclusive/HTTP populations and RP-324/326 corrections | **Partial:** native two-engine fixture proof; exact review ranges and all-engine/all-state proof open. Real HTTP service failures are not supplied by thrown runtime-double errors. |
| GS1-A5 accessibility | Fiscal DOM + host chrome/nav | B axe, native Enter/Space harvest/level/unlock, pending/stale/focus/320 px supplements | **Partial:** three-engine all-state census, full-page 400% zoom, seeded label/width discrimination and manual assistive workflows not all established. F is RED, not waived. |
| GS1-A6 composed | real Fiscal harvest/spend → projected Pitch fact/nav | C actual DOM harvest/unlock/Pitch sequence; recorded receipt sequencing repair | Historical source proof only; changed host must rerun after capacity repair. No synthetic clock allowed; short live clock is not approval of final pacing. |
| GS2-A1 state text | `projectAchievements` → `AchievementsSurface.svelte` | B earned-run/lifetime/locked text and count; G live arms/overlap refusal | Bounded fixture proof, not server acquisition. Unknown required copy has no contained error state (candidate below). Relevant seeded non-color fault/review still need exact gate coverage. |
| GS2-A2 once | `events.ts` cursor → host chrome announcement | B simulated reconnect/dedupe; RP-312 repair and executed omission probes | Bounded announcement proof; not actual server acquisition, full reconnect/engine population or designated repair approval. |
| GS2-A3 not Clout | achievement score projection → score copy/component | B current DOM lacks “Clout”; source component labels score | **Partial:** RFC explicitly demands a source binding guard with a seeded forbidden binding. Current DOM assertion is not that complete guard; Clout's separate accepted consumer must not be banned by a broad scan. |
| GS2-A4 composed | first real generator purchase → earned event + refreshed row | C buys generators; separately reads achievement rows exist | **Missing named witness:** driver does not assert that purchase earns this ID, emits its announcement and changes the displayed row. Add DOM/event/refresh binding and independently sever decoder/arm after RP-236. |
| GS2-A5 accessibility | Trophy Case DOM + chrome/nav | B static state axe; historical combined reflow measurement | **Partial:** all-state/all-engine/400%/error/empty/focus populations and seeded failures not established as a complete gate. |
| GS3-A1 decoder | v4 meter rows → `contracts.ts:parseFeatures` | U rejects out-of-domain value and undeclared band; generic sorted-row validation | **Known source gap RP-329:** parser accepts partial sets (current U positive fixture has only doom). No eleven-ID completeness oracle. Predeclare failed controls before decoder repair; no new wire/balance/schema. |
| GS3-A2 non-color | committed values → `MetersSurface.svelte` numeric/band text | B all ten trust cells match numeric + Low/High; eleven native meters; doom text | Bounded fixture proof; seed omission and missing/presentation error populations still need explicit coverage/review. |
| GS3-A3 reflow | Meters table/CSS → narrow-page labels | B historical 320 px combined measurement; source `td::before` labels | **Source mismatch RP-330:** below 30rem remains block-styled table, not RFC's semantic `<dl>`. Do not call width alone header/AT proof. Implement accepted semantics or obtain author amendment; retain full-page chrome measurements. |
| GS3-A4 composed | real initial pinned meter values → rendered DOM | C authentic public read asserts 11 rows and doom=50 | **Missing named witness:** not DOM text for each Standing=50/Grievance=0/doom=50; add that and arm-severing case after RP-236. |
| GS3-A5 accessibility | meter DOM/header semantics → reading/navigation | B static axe | **Partial:** three engines, semantic narrow reading/header preservation, all states and actual assistive workflow remain. Do not turn a CSS pseudo-label into screen-reader evidence. |
| GS4-A1 care states | PA7 public identity/band/eligibility → `pet/PetCareSurface.svelte` | B availability/refusal/pending/keyboard/captured notices; actual decoder-accepted PA7 fixtures | **Author-blocked RP-132:** accepted GS4 asks raw stats/mood/behavior/cooldowns while PA7 exposes another arm. Current tests must not invent raw private fields or claim that contract covered. |
| GS4-A2 absent pet | `pet_care_fact_test.go` adoption count → nav unlock | G `TestPetCareFactRequiresAnAdoptedPet`; B petless nav absent | Bounded Go/fixture proof; accepted historical “epoch-8 fresh Founder” is not a current release pin. Preserve source/epoch identity and separate current real-server absence assertion. |
| GS4-A3 revision | Founder `care_action` → runtime/refresh | B Founder7/8 command and held refresh; separate cosmetic composed adoption/feed/reload | Bounded public-arm proof, not full raw-care contract. Corrective ranges still need Claude; changed host real service rerun remains. |
| GS4-A4 tone | owner pet copy → Copy Pipeline/lint | `client/tools/copy-pipeline.mjs:334–345` companion rule bans price/urgency/curtain/satire tokens; `verify-copy.mjs:74–81` executed failing fixtures belong to the existing copy gate | Dedicated gate **exists**, not merely general lint. Last recorded copy pass predates this audit; no fresh run claimed. PA8.2 requires `companion`, whereas this Garage body still says `diegetic`: author-owned reconciliation remains, not implementer copy edits. |
| GS4-A5 composed | adoption producer → real care/receipt/state/reload | Historical cosmetic composed positive feed witness | DG-4 producer now mechanically exists, but RP-132 body reconciliation and RP-318 missing pet-status decoder/copy remain. **Not silently cleared:** full GS4 population is blocked, never a skipped success. |
| GS5-A1 stable order | optional opportunity arm → persistent Desk region | B stable region/index/focus native test | Bounded proof; original wire mismatch RP-313 remains. Scope review does not follow from local tests. |
| GS5-A2 no synthetic commands | host cadence → no idle intent | B actual 60 s native wait; RP-315 timer correction/probe | Bounded two-engine actual-timer evidence, not a shortened virtual run. Full F/hosted population still RED/held. |
| GS5-A3 cap | Lucky/buff receipt → Desk notice + pinned reason | B saturated Lucky/capped-buff/projected cap; RP-314 compiling omissions | Bounded receipt fixture proof, not real Lucky payout. Correction awaiting designated review; RP-313 authority hold independent. |
| GS5-A4 composed | real DOM clicks/claim → receipt-bound successor | C actual click buff at revision26; shared driver's oracle unit/native tests and six faults | Actual buff branch only, **not Lucky integration**. Changed-host rerun held; old vacuous shortcut survivor and corrected fixture disclosed under RP-316. |
| GS5-A5 accessibility | opportunity region → native claim/text | B Tab + Enter/Space, axe and 320 px | Partial all-engine/state/400%/AT proof; announcement timing and off-surface withholding require their own population, not keyboard substitution. |
| GS6-A1 cap reason | generator `provision_cap` → Desk card | B provisioned=cap fixture checks exact reason; G generator projection | Bounded reason-text proof; executed omission seed and exact review coverage must be pinned for this gate. No whole Desk acceptance inferred. |
| GS6-A2 no fake purchase | intent allowlist → unchanged free shelf | Current code has no fake cosmetic purchase in this Garage slice | **Not established as prescribed source guard/seed.** Later accepted Cosmetic Shop has legitimate intents: narrow this Garage rule to forbidden legacy shelf actions, never declare all cosmetic intents illegal. Author body reconciliation may be needed. |
| GS6-A3 no wider Desk | Desk additions + persistent chrome → 320 px page | Historical combined reflow measurement | **Partial:** lacks a named pinned before/after full-page comparison in this audit. Accessibility RFC owns old 647 px defect; do not transfer or erase it. |

## Overall criteria (8)

| Criterion | Current evidence / remaining gate |
|---|---|
| 1. All per-surface gates discriminate | The 29-row inventory above is not a pass. Fault tables prove particular repairs, not every listed population. GS7's adopted API lane is separate and is not re-certified here. |
| 2. Shared projection parity | Go/TS positive/negative tests exist, but inspected Garage fixtures are independently constructed, not one identified shared v4 artifact. Five required seeded classes and producer/decoder same-fixture binding need explicit inventory/proof. |
| 3. Runtime receipts | `a42406f0..fc911784` adds 21 actual-runtime/Response.json cases (injected fetch, not live service); existing outcome tests retained. Typed outcomes/errors/unknown outcome/exact rejection, one unchanged request and four behavioral faults are recorded. No generated-client or renewal conformance inferred. |
| 4. Boundary | Last boundary pass at `2d3629ad`; literal/transport/style rules remain mandatory. A fresh audit-only doc does not re-execute them. Prescribed seeded literal must be tied to exact evidence, not merely validator source. |
| 5. Accessibility | Native Chromium/WebKit subsets pass. Linux three-engine lane is RED and diagnostic environment invalid; manual AT, every state, full-page 400% zoom/coarse-pointer target measurements remain open. |
| 6. Composed | Historical Fiscal→Pitch and buff claim paths are real; GS2-A4 and GS3-A4 DOM assertions missing, GS4 accepted-body hold persists, changed host not rerun. RP-236 must be repaired/rechecked without unowned cleanup before Docker evidence. |
| 7. Performance | Separate native performance observation passes unchanged; it does not by itself identify the complete Desk-region mounted population or replace the original 60 s scenario, full browser population or manual reference profile. Audit its driver/population before promoting this criterion. |
| 8. Closeout | Canonical docs describe implementation/limits, but producer body holds and separately owed Claude ranges remain. No RFC status change, checkbox flip, archive or release promotion authorized. |

## Shared error-state investigation and next bounded work

GS0.5 requires a contained `common.surface_error` alert, no controls, one
invariant and other surfaces still operable. Source inspection found
`AchievementsSurface.title()` and `MetersSurface.band()` throw for missing
copy/presentation; `GameUIApp` has no render boundary. Their decoder admits
an unknown mechanical achievement copy key and a meter band declared in the
wire but absent from presentation. **Candidate defect, not yet a fresh executed
browser failure.** This is distinct from malformed wire, module-import failure
or changing authoritative data.

Next: separately predeclare decoder-admission/healthy controls and native
mount/navigation cases for those two populations; record actual crashes and
contract assertions. No production repair inside this audit range. If reproduced,
use a separate failed-first GS0.5 repair range with existing copy, error isolation,
one diagnostic per failing surface and retained navigation. Do not broaden it
into an invented renderer, wire contract or owner-copy rewrite.

Following safe accepted work: GS3-A1 complete-ID decoder; GS3-A3 semantic narrow
layout; exact source guards; missing real DOM acquisition/meter
witnesses after capacity restoration. RP-132/313/318 author obligations remain
separate. No additional empirical or owner decisions are answered by this file.
