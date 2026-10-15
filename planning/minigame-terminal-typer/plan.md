# Terminal Typer implementation plan

RFC: `rfc/minigame-terminal-typer.md` (accepted 2026-09-25; all ODs at recommended defaults).
Fixture-first (OD-11): no production epoch is minted by this plan.

2026-10-15 independent engine corrective review APPROVES RP-395 (`f423e704`),
RP-400 (`ba1365c9`), RP-401 (`ed92ede7`), RP-402 (`10d94e70`), RP-429 (`20a57019`)
and RP-461 (`1fe602e2`), each complete single-commit span. The latest owning log records
fresh Codex provenance, cold Go/TS vectors/admission/replay checks, full coordinates and
descendant limits. Earlier native/availability approvals remain separate. Other corrective
ranges, RP-162/RP-460, public HTTP/content and whole-feature acceptance remain open.

2026-10-14 public-wire preparation: `rfc/minigame-tenant-http-bindings.md` is the new draft
home for the TT-PA4/C2 reconciliation, not adoption. Its disjoint prefix avoids RP-562's
executed static-route shadowing; versioned metadata/directory avoids silently remapping
the v1 artifact's four operation IDs. The draft now specifies exact descriptors/guards,
metadata/version floors, transactions and explicit drain binding. Engineering proposal
is ready for normative review; parent-body reconciliation, adoption/designated review and
RP-162/RP-460 remain.
Do not treat this draft as permission to implement B5/B6, refresh a pin or expose Typer.

2026-10-14 bounded TT-PA2 correction: RP-563's availability projection now reads the same
pinned tier/Founder-Exit predicate as session creation. Complete candidate-bundle projection
regressions and affected cold package checks pass; no public binding or feature completion.
The owning log records independent corrective approval of `5cd7119f^..5cd7119f`
under Marco's 2026-10-15 Codex-review delegation; this is not whole-feature acceptance.

## Batches

- [ ] B1 — PARTIAL (full original `345dc0b9^..345dc0b9` designated Codex review is now CHANGES REQUIRED, recorded in the 2026-10-08 log; Selected corrective ranges are approved below; other exact ranges still need independent review. RP-162 result bounds and RP-460 timing arithmetic remain author-contract gaps) — Content + engine (TT3/TT4): `balance/testdata/typer-v1.json` fixture (placeholder
  prompts per OD-9), Go `server/typer` + TS `client/src/typer` loaders and pure engine,
  `ApplyInput.ServerTimeMs` (TT-PA1 item 2), shared content-gate corpus
  `testdata/typer/content-gate-v1.json`. ACs 1 (artifact half), 2, 4, 5, 6, 7, 10 (engine
  schema), 12, 15.
- [x] B2 (`3eb7e401`) — TT-PA1 time plumbing: coordinator samples the DB clock once per admitted command before
  tenant execution; command-append inserts that exact value; replay passes persisted stamps. AC3.
- [x] B3 (`885237a7` (composed AC9 witness `a3c87149`)) — TT-PA2 `tier_at_least` unlock arm (Go/TS loaders) + composed start resolver
  (`not_eligible/tier_required`, `not_eligible/curriculum_exit_required`). AC9.
- [x] B4 (`391beb76`) — TT-PA3 content resolver/bundle generalization + TT1 tenant row in a fixture minigames
  artifact + `minigame_api` loader chain. AC1 (definition/artifact/api chain half).
  Full original B3 (`885237a7^..885237a7`) and B4 (`391beb76^..391beb76`)
  designated Codex reviews now record CHANGES REQUIRED in the 2026-10-08 log:
  RP-152/RP-154/RP-155 and raw unlock admission RP-461. Separate Codex corrections
  still need independent review; checked rows record implementation presence, not feature acceptance.
- [ ] B5 — PARTIAL (`d7c1ce6e`: error details only; snapshot/command arms blocked by the TT-PA4 vs API C2 DESIGN-GAP) — TT-PA4 API arms (Typer snapshot/command arms) + generated client. AC10 (API half).
- [ ] B6 — PARTIAL (`ade1083b`: TyperTable + AC13 gates; registration waits on the pin and the C2 gap) — UI: `TyperTable` child under `client/src/game-ui/minigame/`, tenant-registry row,
  accessibility contract TT8. ACs 13, 14.
  Full original `ade1083b^..ade1083b` designated review is now CHANGES REQUIRED;
  the 2026-10-08 log records all three paths and the separate correction evidence.
- [ ] B7 — PARTIAL (`a3c87149`, `9e82a7f0`, `ec2dfbf9`, `1260be38`, `11c7af4b`: AC9, AC11, AC8 service-level path including real Exit, fifth/sixth daily cap, offline quality and replay quota rejection; public MA path blocked) — Composed platform path (real Postgres): create → begin → submits → terminal → faucet →
  receipt; neutrality timed/untimed. ACs 8, 11.
  Full original `a3c87149^..a3c87149` and `9e82a7f0^..9e82a7f0` designated
  reviews are now CHANGES REQUIRED, not implicit approval of Codex's supplements.
- [ ] Canonical docs (`docs/minigame-terminal-typer.md`), designated Codex review, archival (Codex).

### B6 local display-clock correction (2026-10-05)

RP-187's cold CI update-depth error is reproduced by an increasing injected monotonic clock
in all three engines. The bounded display-only correction samples into one non-reactive local
before assigning both state fields. Ready/timed/new-server-revision states and all existing
native keyboard, composition, prompt, reflow/axe cases pass (24/24). Reinstating the reactive
read-back independently fails the retained clock case in all three. The full cold browser
CI lane passes 21,102 tests plus its separate performance case, with unit/type/build/boundaries
green. Claude's designated review of the exact correction span in `log.md` remains required.
This does not close public registration, owner content, manual AT, or B6 as a whole; RP-188's
earlier Snake failure and RP-131's historical kernel guard remain independent open findings.

Kernel protocol: every commit touching a `kernel/affecting-paths.json` prefix bumps
`kernel/VERSION` (+ Go/TS constants) in the same commit; new engine dirs are registered in the
commit that creates them.

### B1 raw JSON admission correction (2026-10-08)

RP-429: TT3/TT4 raw-input parity, not public-wire construction. Shared Go/TS vectors
cover duplicate keys at root/nested/escaped names and integer token spelling before
parsing can discard evidence. Engine creation/application and both replay-bundle
loaders exercise the boundary; valid escaped keys/text and command error taxonomy
remain intact. Kernel176->177 records the changed replay acceptance set. The log
owns executed results and review range. B1 remains partial: RP-162/RP-460, corrective review,
public C2 integration, adopted content and feature acceptance are not closed here.

### B6 native keyboard/lifecycle correction (2026-10-08)

RP-394: accepted TT8.3/TT8.6/TT9 child behavior, not a public-wire amendment.
Old-source native Chromium/WebKit checks fail24 times: pending control focus/
reachability, newer-focus theft, terminal focus loss and WebKit Tab entry.
Mode/Submit/End now retain native tab stops with guarded pending callbacks;
pre-render focus capture and post-render ownership/lifecycle checks preserve
newer choices and move removed terminal controls to Leave. Complete keyboard-
only untimed child run uses actual TS engine/pinned placeholder data, not server
or payout proof. The owning log records final affected checks and exact review
scope. B6 stays partial: public registration/C2, adopted content, full three-
engine/manual-AT acceptance and designated review remain separate obligations.

RP-475's interrupted-composition correction is also locally verified under TT8.2/TT9:
removing the text field clears its composition flag, while a same-input response retains it.
The old child suppresses a later run's native Enter in both Chromium/WebKit; the corrected
complete child population, types/build and shell boundaries pass. This is component lifecycle
proof, not public registration, OS IME certification or acceptance of B6.

2026-10-15 independent corrective review: RP-394 (`e019e670^..e019e670`),
RP-475 (`59e5cc49^..59e5cc49`) and RP-563 (`5cd7119f^..5cd7119f`) are APPROVED
by fresh Codex reviewer `/root/typer_ui_availability_corrective_review`, under Marco's
corrective-code delegation. The owning log records independently executed descendant
browser/Go/type/build checks and their limits. Other corrective ranges and original
feature acceptance remain separate; these three corrections no longer wait on Claude.
