# Cosmetic Shop v1 implementation plan

RFC: `rfc/cosmetic-shop-v1.md` (accepted 2026-09-25, all ODs at recommended defaults).

## Current checkpoint — 2026-10-16

All original C1–C8 diff reviews are complete. C2 retains its bounded approval; the other
original batches retain their historical CHANGES REQUIRED verdicts. Corrective approval
does not accept those whole features or authorize archival.

Fresh independent Codex reviewers, under Marco's corrective-code delegation, approve the
complete corrections for RP-173, RP-175, RP-177–RP-183, RP-185/RP-186, RP-411, RP-424
and RP-453–RP-458. Exact ranges, old failures, corrected results and limits live in
`log.md`; the dated verdicts there are authoritative, not implementation checkboxes.
Latest approvals include the real-Postgres both-Exit carry, actual second-wearer transition,
full 200 × 288 isolation/receipt/bonus-consumer checks and actual-source package-guard negatives.

Keyboard RP-167, served-path observer RP-168, browser collection RP-169 and request-completion
RP-170 corrections also have independent approval over complete `8b4db5a9..c3311dbd`.
The later composed/negative-build construction and complete N5/AC11/AC13 gates are separate.

RP-377 `d82e76c2..9818e589` and RP-385 `d717e590..17c552db` driver repairs now have
independent corrective approval, including a real held-response old/fixed pair and exact
endpoint/current Postgres/browser/socket journeys. This does not establish the unique causes
of their original timeouts. RP-364 and Garage's RP-378 remain separate findings.

Next: finish remaining corrective/evidence reviews without repeating approved ranges.
Initial AC14/G10 and later fixture-build construction retain their existing review gates;
the corrective-only delegation does not approve new feature construction or acceptance.

Feature gates still open: ruling-author RP-174/RP-176 body reconciliation; D-024 shared-clock
authority for RP-365's reproduced backwards-attendance conflict; owner copy adoption and
production content; complete acceptance including manual AT and release integration.
No pet clamp/retry is authorized by the clock finding. The 27 Clout accounting failures and
historical RP-131 kernel guard remain red; selected passes are not full CI or a 1.0 claim.
Historical checkpoint details remain intact in the append-only log and Git history.

Original implementation by: Claude. Original-range verdicts and remaining corrective reviews
are recorded below and in the log; not every implementation checkbox is an approval. There is
no self-approval, no archival and no production mint (fixture-first).

Numbering is landing-order. The RFC's "Founder v22" (OD-16: next free) becomes **Founder v24**,
because Reputation took v22 and Pet Adoption took v23. The replay-inputs carry and the events
migration take the next free numbers at landing.

- [x] C1 Catalog grammar: `cosmetics` schema v1 loaders (Go `server/cosmetic`, TS
  `client/src/cosmetic`), shared reject fixtures, and the permanent-ID transition check (AC1, AC2).
  Codex's complete original review retains its historical RP-175 finding; the corrective range
  is now independently approved. This checkbox is implementation presence, not archival approval.
- [x] C2 Replay-bundle wiring: `cosmetics` joins the constants bundle (OD-10) and requires
  `pet_species` on the scalar Founder chain. Codex designated-approved this bounded C2 range
  after cold Go/TS checks and independent chain, identity and settlement severing; see `log.md`.
- [x] C3 Founder v24 `cosmetics` state: codec, validation, activation at the new-run boundary,
  Exit carry and the replay-inputs carry, in both runtimes (AC3, AC4, AC8). Codex's targeted
  C3/AC4 review found RP-176: the five mandated save-migration-corpus cases are absent and
  their specified home conflicts with new-run-only activation. This box is implementation
  presence, not C3 acceptance or archival approval; ruling-author reconciliation is required.
  RP-181's separate AC8 test-only supplement now covers nonempty owned/equipped carry through
  both named Exit paths in Go/TS and real Postgres; its complete corrective range is independently approved.
  The complete original C3 review additionally found RP-457's Go saved-state case aliases;
  its bounded correction is independently approved as recorded above, not full C3 acceptance.
- [x] C4 Intents `acquire_cosmetic` / `equip_cosmetic` / `unequip_cosmetic`, the three events and
  the migration, with Go/TS parity and Postgres integration (AC5, AC6, AC9). Codex's targeted
  C4/AC9 review found RP-177: only the Go decoder, not Postgres, rejected extra event payload
  keys. Append-only migration 00084 and retained DB negatives now have independent corrective
  approval. Codex's further C4/AC5/§4.5
  review found RP-178: active Soul recovery blocks cosmetics as `exclusive_activity` at the
  top-level service preflight. The bounded exemption and a real-Postgres acquire/equip/unequip
  → recovery-resolution witness now pass, with ordinary gameplay still exclusive and fired
  exemption probes. The complete correction now has independent approval.
  Codex completed the full original C4 range review on 2026-10-08; its consolidated
  CHANGES REQUIRED verdict and the outstanding corrective reviews are in the owning log.
  RP-179's shared two-pet equip transition is now
  supplied by a test-only corpus supplement, with independently fired Go/TS single-wearer
  mutations. Its complete corrective range is independently approved.
  RP-411's later replay-admission review found Go defaulting missing/null Company context and
  accepting unsafe revisions or extraneous null context that TS rejects. The bounded strict
  correction and shared admission fixtures are independently approved; the latest owning log gives
  the exact range and executed results. No complete C4 or archival approval is inferred.
  The real-Postgres command population now also checks seven persisted refusal/retry cases,
  eight concurrent identical submissions, stale/corrected revisions and rollback when the last
  receipt-outbox insert fails. Focused race and selected Cosmetic/save tests pass; the owning
  log records scope and pending designated review. This is not HTTP authorization coverage.
  This checkbox records implementation presence, not full C4 approval.
  RP-453's event-key admission correction aligns Go with the existing exact-key database
  constraint while accepting equivalent JSON whitespace/escapes. The new semantic-key
  regressions fail on the old validator; owning log records executed SQL controls and the
  independent approval of this complete correction. No full C4 approval is implied.
- [x] C5 Snapshot v4 optional `features.cosmetics` arm and its decoder (AC10); complete original
  review retains CHANGES REQUIRED for RP-173/RP-174/RP-458. RP-458's complete correction is
  independently approved, as is RP-173's correction; RP-174 requires
  ruling-author body reconciliation. This is implementation presence, not archival approval.
- [x] C6 Desk shelf, parody receipt, `CosmeticOverlay` and the curtain contract (AC11, AC12, AC15).
  Codex's bounded AC12 review found RP-183: Equip/Unequip lacked the visible curtains'
  `aria-describedby` links. A retained all-controls assertion fails first across all three
  engines; the link-only correction and independently firing Unequip probe pass locally.
  Independent corrective review now approves the complete range; original C6 review is complete.
  AC15 review also found RP-185's prop-only media evidence and RP-186's actual CSS specificity
  defect. The failing-first three-engine preference case and bounded CSS correction pass all
  shelf cases and both real-server journeys, with separate firing motion/caption probes.
  Independent corrective review now approves their complete combined range. The first full
  browser run was red on RP-187/RP-188; later Typer and Arcade corrections pass the cold full browser target. The
  historical kernel-history gate still blocks complete CI; this is not C6 or release acceptance.
  RP-454/RP-455 add native applied/refused completion-focus regressions for wearer controls
  and Buy. The bounded guard correction is tracked in the owning log, separately from
  whole-product assistive-technology acceptance. RP-424/RP-454/RP-455/RP-456 now have
  independent corrective approval as recorded above; the complete original C6 diff retains
  its historical CHANGES REQUIRED disposition. Other acceptance obligations remain;
  no whole-C6 or feature acceptance is inferred.
- [x] C7 No real money: `verify-no-payment`, the copy currency lint, the Caddy
  `Permissions-Policy`/CSP, and the browser network trap (AC13).
  RP-180's syntax-aware package-gate correction is independently approved;
  the old regex admitted a real dynamic economy import.
  The complete original C7 diff now has a consolidated CHANGES REQUIRED verdict for
  RP-168/RP-180; its remaining original diff is reviewed, not retroactively approved.
- [x] AC14 composed real-server witness (Buy → reload under the N5 trap, with a severed-producer
  failing case). A test-only epoch pins `cosmetics`; no production mint. The test and checkbox
  land in the same range; Claude's designated review remains an independent archival gate.
- [x] C6/AC11 keyboard acceptance supplement (RP-167): real Enter and Space Buy activation is
  mutation-proven; the cold Linux three-engine lane passes, while Firefox cannot start on this
  macOS 27 host. The complete corrective range is independently approved. Do not cite
  C6's old `.click()` test as keyboard proof or claim the wider task/AT gate complete.
- [ ] C7/AC13 N5 network witness (RP-168): a browser-level observer now captures four HTTP
  transport/resource classes on a disallowed served path in the cold Linux three-engine lane,
  with RP-170's full-suite race corrected and a fired listener-severing control. The
  observer correction now has independent approval, not acceptance of the later construction.
  The composed lane now includes two actual fixture builds whose native Buy callback attempts off-origin
  checkout fetch or PaymentRequest; exact primary-trap rejections are locally witnessed before
  egress. Green/unrelated failures, missing injection, secondary route aborts and incomplete
  cleanup cannot satisfy the negative verifier. Final execution and limits live in the latest
  log; the real-server AC14 flow and later fixture-build supplement await Claude's cross-party review of the
  relevant correction/evidence ranges. No whole-feature acceptance is inferred.
- [x] Controlled real-server G10 pet-overlay witness: the existing Garage mount now has actual
  DOM adoption/equip, pet-panel/reload, browser reduced-motion preference and unequip proof.
  The test-only driver supplement and restored consumer/motion probes land in this range.
  Claude designated review is required; this is not production content or release acceptance.
- [ ] G10 release acceptance and owner adoption of the candidate copy; the live mount itself
  already exists in Garage `7a61e4b6`, rather than being an unimplemented consumer.
- [ ] Designated cross-party review (Codex) and archival. Not self-approved.
- [x] C8 Mechanical-isolation property test, package gates, docs (AC7, AC16).
  Codex's targeted AC7 review found RP-182: receipts were discarded and the Company arms used
  nil Founder contributions. The test-only correction retains 200 × 288 steps, compares every
  receipt/complete state, consumes correctly frozen bonuses and checks a separate next-run
  consumer with firing negatives. Its complete corrective range is independently approved; this box records
  implementation presence, not C8 approval. The complete original C8 diff now has a consolidated
  historical CHANGES REQUIRED verdict for RP-182. RP-180's package-gate correction is also approved.

Box evidence: each flipped box's exercising tests landed in that batch's commit
(`afe6529b`, `8e315569`, `581886a4`, `500d944c`, `1a477d9e`, `1bba27ba`, `a4b22429`, and this
C8 commit). See `log.md`.

## Codex correction review index (2026-10-05)

This historical range index does not itself supply approval. RP-173, RP-175, RP-177–RP-183
and RP-185/RP-186 now have fresh independent Codex corrective approval under Marco's delegation; remaining rows
still need designated review. Exact verdicts, executed checks and limits are in `log.md`.
Reviewing corrections does not automatically approve the original Claude C1–C8 span;
its bounded outstanding findings and ruling-author blockers still apply.

| Finding / bounded correction | Exact Codex span |
|---|---|
| RP-173: snapshot reader contradictions | `ebb88bab^..547ed7f0` |
| RP-175: raw JSON integer parity | `7644808d^..76a9fe04` |
| RP-177: Postgres event payload key sets | `fca686c5^..ce7688a7` |
| RP-178: cosmetics during Soul recovery | `380d854b^..646daa08` |
| RP-179: second actual equip in shared replay | `4b319313^..144f5e8e` |
| RP-180: syntax-aware package gate | `c09e7e9a^..ad789790` |
| RP-181: both nonempty persisted Exit paths | `e7bc1c8b^..66cfa3db` |
| RP-182: production receipts and frozen bonus consumer | `52bd6963^..7c16d890` |
| RP-183: wearer-control disclosure links | `e5c63de5^..e02f6560` |
| RP-184: persisted live pet-overlay workflow and truth reconciliation | `b717b0db^..e4e6e567` |
| RP-185/RP-186: actual three-engine motion evidence and CSS precedence | `e47f9dea^..3be0e5dd` |

Reuse the declared root Make/Compose targets from the logs; selected green checks are not
green complete CI. RP-174 and RP-176 remain ruling-author body-reconciliation blockers.
