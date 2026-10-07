# Cosmetic Shop v1 implementation plan

RFC: `rfc/cosmetic-shop-v1.md` (accepted 2026-09-25, all ODs at recommended defaults).

Current checkpoint (2026-10-07): RP-377's premature overlay check is locally corrected: the
driver awaits host settlement within the original action deadline, retaining exact receipt,
persisted wearer and rendered-overlay assertions. Focused Chromium/WebKit and the real-service
Cosmetic journey pass; designated review remains. The aggregate stops earlier at RP-378's Fiscal
refusal response timeout, so full composed/CI acceptance stays open. See the owning logs.

RP-365's care conflict has a reproduced backwards-attendance
route, both real Postgres/DOM and deterministic cutoff regression. Sources restored; replayable
probe and exact evidence in latest `log.md`. Shared-clock repair needs D-024/the draft successor's
explicit boundary; no pet clamp/retry is authorized. RP-364 Buy cause and prior reviews remain.

Implemented by: Claude. Every range awaits Codex's designated cross-party review. There is no
self-approval, no archival and no production mint (fixture-first).

Numbering is landing-order. The RFC's "Founder v22" (OD-16: next free) becomes **Founder v24**,
because Reputation took v22 and Pet Adoption took v23. The replay-inputs carry and the events
migration take the next free numbers at landing.

- [x] C1 Catalog grammar: `cosmetics` schema v1 loaders (Go `server/cosmetic`, TS
  `client/src/cosmetic`), shared reject fixtures, and the permanent-ID transition check (AC1, AC2).
  Codex's targeted review found RP-175 raw-number parity; the corrective range awaits Claude's
  designated review. This checkbox is implementation presence, not archival approval.
- [x] C2 Replay-bundle wiring: `cosmetics` joins the constants bundle (OD-10) and requires
  `pet_species` on the scalar Founder chain. Codex designated-approved this bounded C2 range
  after cold Go/TS checks and independent chain, identity and settlement severing; see `log.md`.
- [x] C3 Founder v24 `cosmetics` state: codec, validation, activation at the new-run boundary,
  Exit carry and the replay-inputs carry, in both runtimes (AC3, AC4, AC8). Codex's targeted
  C3/AC4 review found RP-176: the five mandated save-migration-corpus cases are absent and
  their specified home conflicts with new-run-only activation. This box is implementation
  presence, not C3 acceptance or archival approval; ruling-author reconciliation is required.
  RP-181's separate AC8 test-only supplement now covers nonempty owned/equipped carry through
  both named Exit paths in Go/TS and real Postgres; Claude designated review remains required.
- [x] C4 Intents `acquire_cosmetic` / `equip_cosmetic` / `unequip_cosmetic`, the three events and
  the migration, with Go/TS parity and Postgres integration (AC5, AC6, AC9). Codex's targeted
  C4/AC9 review found RP-177: only the Go decoder, not Postgres, rejected extra event payload
  keys. Append-only migration 00084 and retained DB negatives correct this locally, but Claude's
  designated review of the Codex correction remains required. Codex's further C4/AC5/§4.5
  review found RP-178: active Soul recovery blocks cosmetics as `exclusive_activity` at the
  top-level service preflight. The bounded exemption and a real-Postgres acquire/equip/unequip
  → recovery-resolution witness now pass, with ordinary gameplay still exclusive and fired
  exemption probes. Claude's designated review of this correction and Codex's review of the
  remaining C4 scope are still required. RP-179's shared two-pet equip transition is now
  supplied by a test-only corpus supplement, with independently fired Go/TS single-wearer
  mutations. Claude designated review of this evidence correction remains required.
  This checkbox records implementation presence, not full C4 approval.
- [x] C5 Snapshot v4 optional `features.cosmetics` arm and its decoder (AC10); Codex's targeted review found RP-173, whose corrective range awaits Claude's designated review, and RP-174 still requires ruling-author body reconciliation. This checkbox is implementation presence, not archival approval.
- [x] C6 Desk shelf, parody receipt, `CosmeticOverlay` and the curtain contract (AC11, AC12, AC15).
  Codex's bounded AC12 review found RP-183: Equip/Unequip lacked the visible curtains'
  `aria-describedby` links. A retained all-controls assertion fails first across all three
  engines; the link-only correction and independently firing Unequip probe pass locally.
  Claude designated review of this correction and the broader C6 review remain required.
  AC15 review also found RP-185's prop-only media evidence and RP-186's actual CSS specificity
  defect. The failing-first three-engine preference case and bounded CSS correction pass all
  shelf cases and both real-server journeys, with separate firing motion/caption probes.
  Claude must designated-review this Codex correction. The first full browser run was red on
  RP-187/RP-188; later Typer and Arcade corrections pass the cold full browser target. The
  historical kernel-history gate still blocks complete CI; this is not C6 or release acceptance.
- [x] C7 No real money: `verify-no-payment`, the copy currency lint, the Caddy
  `Permissions-Policy`/CSP, and the browser network trap (AC13).
  RP-180's syntax-aware package-gate correction is locally tested and awaits Claude's
  designated review; the old regex admitted a real dynamic economy import.
- [x] AC14 composed real-server witness (Buy → reload under the N5 trap, with a severed-producer
  failing case). A test-only epoch pins `cosmetics`; no production mint. The test and checkbox
  land in the same range; Claude's designated review remains an independent archival gate.
- [ ] C6/AC11 keyboard acceptance supplement (RP-167): real Enter and Space Buy activation is
  mutation-proven; the cold Linux three-engine lane passes, while Firefox cannot start on this
  macOS 27 host. The test-only Codex range still needs Claude's designated review. Do not cite
  C6's old `.click()` test as keyboard proof or claim the wider task/AT gate complete.
- [ ] C7/AC13 N5 network witness (RP-168): a browser-level observer now captures four HTTP
  transport/resource classes on a disallowed served path in the cold Linux three-engine lane,
  with RP-170's full-suite race corrected and a fired listener-severing control. Off-origin
  attempts blocked before a request remain unproved; the real-server AC14 flow is now locally
  witnessed but awaits Claude's cross-party review. Claude's
  review of both Codex test ranges is required.
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
  consumer with firing negatives. Claude's designated review is required; this box records
  implementation presence, not C8 approval. RP-180's package-gate correction also awaits review.

Box evidence: each flipped box's exercising tests landed in that batch's commit
(`afe6529b`, `8e315569`, `581886a4`, `500d944c`, `1a477d9e`, `1bba27ba`, `a4b22429`, and this
C8 commit). See `log.md`.

## Codex correction review index (2026-10-05)

All rows below await **Claude's designated cross-party review**, not just recording a
Codex first-filter result. Their predeclarations, executed failures and restored probes are
in `log.md`. Reviewing them does not automatically approve the original Claude C1–C8 span;
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
