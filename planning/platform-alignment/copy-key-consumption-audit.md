# Copy-key consumption audit at `190a4fa`

Measured 2026-08-21 under the frozen rules in `copy-key-consumption-plan.md`. The exact row ledger is
`copy-key-consumption-inventory.tsv`; `copy-key-consumption-extractor.mjs` reproduces it from the
checked-in source catalogs, generated catalog/report, current epoch, production client reachability
ledger, explicit Go registry, fixtures, and bounded dynamic resolver sites.

## Result

| Verdict | Keys | What is actually established |
|---|---:|---|
| `mounted_player_copy` | 128 | A bounded production client path can select the key in the current Game UI. |
| `shipped_backend_or_data_only` | 63 | A current artifact or Go producer owns the key, but no mounted player reader exists. |
| `shipped_unmounted_surface_copy` | 1 | The literal call ships, but current data makes its guarded branch unselectable. |
| `fixture_or_tool_only` | 8 | Only the two event-copy testdata revisions select these keys. |
| `unreferenced_candidate` | 8 | No exact bounded consumer/reference outside source/generated membership was found. |
| `ambiguous_dynamic` | 0 | All 14 computed resolver call sites were enumerated and bounded at this coordinate. |
| **Total** | **208** | Unique generated application-catalog denominator reconciles exactly. |

The checked-in orphan report has the following cross-tab, which is the decisive result:

| Current report | Mounted | Backend/data only | Unmounted binding | Fixture/tool only | Unreferenced | Total |
|---|---:|---:|---:|---:|---:|---:|
| `orphan` | 105 | 39 | 1 | 8 | 8 | 161 |
| `referenced` | 23 | 24 | 0 | 0 | 0 | 47 |

The report is valid only as “not selected by the seven registered JSON paths or two registered Go
sites.” It is not an unused-copy or player-consumption oracle. Of its 161 warnings, 105 are mounted
player copy. The three predeclared live controls—`desk.buy_one`,
`chrome.run_title.company_fallback`, and `screen.run_end.founder_note`—all resolve mounted while
remaining labelled orphan by the current report.

## Producer and consumer boundary

- The current registry recognizes 45 keys through its seven JSON paths and two more through the Go
  code-site registry. Those 47 keys are exactly the report's `referenced` set, but 24 are still
  backend/data-only; registry membership is not a player workflow.
- Fifty generated keys have exact references in deploy-current artifacts outside the registry:
  Pitch 26, Soul 16, Curriculum seven, and Minigames one. Pitch and Soul have no mounted player
  surface. Curriculum's branch keys are nevertheless mounted through the Run End DTO/component,
  and the minigame cap key remains backend-only. RP-108 records this closed-world registry defect.
- The main client graph contains 71 keys at literal resolver call rows and 62 at bounded dynamic
  binding rows; overlapping lanes explain why those figures are not summed as a denominator. The
  extractor enumerates all 14 nonliteral resolver call sites and fails if that set changes.
- `terms.network_slot_unlock.frame` is the sole shipped-but-unselectable row: the literal Offer/Run
  End term call exists, but the generated presentation map has zero network-slot bindings and
  current `ComputeTerms` emits an empty slot list. It is not promoted to mounted behavior.
- Current category data defines five categories and the client has a five-key map, but the Game UI
  projector hardcodes `any_percent`. Only that category is mounted from the current producer;
  Ethical, 100%, Low%, and Valuation remain data-only. FixtureHost references do not promote them.
- The eight event keys occur only in `balance/testdata/t0-t1/event-copy-v1.json` and `v2.json`.
  They remain fixture/tool-only even though their mechanical event kinds exist in production.

The eight no-reference candidates are:

- `chrome.run_title.pb_label`
- `chrome.run_title.rta_label`
- `screen.run_end.new_route`
- `screen.vision_slide.skip`
- `surface.offer_sheet.title`
- `surface.run_end.title`
- `system.offline_progress.frame`
- `system.offline_progress.tooltip`

They are cleanup/adoption research candidates only. This audit does not authorize deleting or
rewriting owner-authored copy, and their unconditional inclusion in the shipped candidate catalog
remains part of RP-096.

## Out-of-population current-artifact failure

A recursive exact-field pass over all 19 current epoch artifacts found four hardcap reason keys that
are not among the 208 generated application keys:

| Missing key | Current field |
|---|---|
| `cap.fiscal_credit` | `balance/fiscal/first-content.json /credit_policy/hardcap_reason_key` |
| `cap.fiscal_level.beige_tower` | `balance/fiscal/first-content.json /generator_level_rows/0/hardcap_reason_key` |
| `cap.cash` | `balance/opportunities/t0-t1.json /effects/2/hardcap_reason_key` |
| `cap.active_combo` | `balance/opportunities/t0-t1.json /combo_policy/hardcap_reason_key` |

Economy upgrade `copy_key` values are intentionally prefixes and were excluded only after their
`.title` and `.description` expansions were proven present. The four rows above are exact reason
keys, not prefixes. Fiscal and opportunity mechanics can therefore emit/store identifiers that the
application resolver cannot render. RP-109 owns the completeness/refusal gap; supplying the missing
owner-authored text is not an audit-author decision.

## Controls and reproducibility

Run from the repository root:

```text
node planning/platform-alignment/copy-key-consumption-extractor.mjs \
  > /tmp/copy-key-consumption-rerun.tsv
cmp planning/platform-alignment/copy-key-consumption-inventory.tsv \
  /tmp/copy-key-consumption-rerun.tsv
```

The recorded run emitted 208 unique rows, 161 current warnings, 14 bounded computed call sites, and
the verdict counts above; `cmp` was clean. In the same run the instrument demonstrated rejection of
three seeded failures: a dropped denominator row, a live key relabelled unused, and a backend-only
key relabelled mounted. The in-memory absent key resolves `unreferenced_candidate`; the synthetic
unbounded dynamic selector resolves `ambiguous_dynamic`. The current data-only controls
`achievement.first_gate`, `pitch.card.api_call`, and `soul.recovery.defrag.title` did not receive
player-capability credit.

The manual second pass inspected every non-mounted row, the complete eight-key unreferenced and
eight-key fixture cohorts, the sole unmounted row, each dynamic-binding family, and all 14 computed
call sites. No real ambiguous row remains. This is static reachability and bounded current-data
evidence, not proof that every mounted branch has a browser acceptance witness or that every player
will encounter every key.

## Fired criteria and routing

Three criteria fired:

1. the current report labels mounted production copy orphan (RP-097, now exactly 105 rows);
2. the registry excludes four shipped copy-bearing artifact families and 50 present catalog keys
   (RP-108);
3. current artifacts contain four exact reason keys absent from the application catalog (RP-109).

The report did not omit any of the eight bounded no-reference keys. No duplicate, truncation,
parser failure, or unbounded real dynamic selector occurred. The audit author changed no copy,
balance, product, test, generator, RFC, canonical product doc, or owner-authored content. An
accepted Copy successor must own discovery across mounted client calls and all strict current
artifact fields, demonstrate the same seeded failures, and distinguish player-mounted,
backend-only, fixture-only, held, and unreferenced states before any cleanup claim is allowed.
