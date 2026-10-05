# Demo Disc Arcade — implementation plan

RFC: `rfc/minigame-demo-disc-arcade.md` (accepted 2026-09-25, recommended defaults). Fixture-first;
no production mint (AR8). Implemented by Claude; every range awaits Codex designated review.

- [x] A1 — `arcade` artifact grammar (AR1.2/AR3.1/AR4.1): Go + TS loaders, one shared fixture,
  loader-bound rejections, candidate copy keys (AR6.6) in `copy/catalog/arcade-candidate.json`.
- [x] A2 — `mine_grid` 1.0.0 pure engine (AR3) in Go and TS; Go-generated corpus TS replays
  byte-for-byte (AR7); hidden-information invariant (AC4) with a failing case.
  RP-190's test/corpus budget correction counts all 60 attempts, rather than only 43 applied
  commands. Go/TS equalities fail first and their independent applied-only counter probes fire;
  the regenerated fixture differs only in budget metadata. Claude review remains required.
  RP-191 replaces the terminal-only TS metadata loop with 44 actual genesis/attempt outputs,
  five required phase/placement populations, clean decoder controls and forged-hidden-state
  refusal at both decoder and apply entry. Four independent severing controls fire; production
  probes are restored. This test-only supplement awaits Claude review, not full AC4 acceptance.
  RP-192's separate watched TS decoder correction rejects malformed numeric/list/row values,
  backed by 19 shared Go/TS negatives and clean controls. Kernel 0.3.144 records the actual
  acceptance-set change. It awaits Claude review and does not establish full raw JSON parity.
  RP-193 separately corrects raw duplicate/integer-token admission in TS and null/missing/
  unsafe numeric admission in Go. Twenty-five shared negatives and six valid spellings/bounds
  exercise four actual states; malformed nesting is refused at the declared shape depth.
  Kernel 0.3.145 records the watched decoder changes. Claude review remains required.
  RP-194 adds actual Go genesis/post-attempt witnesses and literal snapshot/result bytes to
  the test-only v2 corpus. TS compares all 76 states, including every rejected attempt;
  intermediate output mutations now fail. Original v1 gameplay metadata/population is preserved.
  This closes a bounded instrument gap locally; Claude review and full AR7 acceptance remain.
  RP-198 adds a real small-board rejected RNG draw, derived without searching or mocking.
  Actual Go registry and independent TS placement/apply compare four attempts / five literal
  states; modulo/threshold/forged-artifact probes fail. Existing flood/exclusion/chord controls
  already discriminate in both engines. The test-only supplement awaits Claude review, not
  full A2/AC3/AR7 or public-integration acceptance.
- [x] A3 — `snake` 1.0.0 pure engine (AR4) in Go and TS; corpus incl. mid-window terminal,
  `advance_past_terminal` without mutation, every rejection (AR7).
  RP-195 separately corrects both snapshot decoders' malformed numeric/raw/identity admission,
  using 45 shared negatives and seven legal controls over four actual states. Kernel 0.3.146
  records the acceptance-set change; independent guard probes fail. Claude review is required.
  RP-196 adds a separate actual 5×5 clearing fixture/witness: all 1,024 strategy seeds are
  recorded (one cleared / 1,023 excluded-food failures), selected seed 455 clears at tick 134.
  Go registry output and independent TS literal replay cover 26 attempts / 27 observations,
  including overlong terminal-window and phase refusals. Existing 6×5 evidence is preserved.
  Discrimination controls fail; Claude review is still required. No whole A3/AC5 or public
  integration approval is inferred here.
  RP-197 adds six actual rejected commands at genesis/moved states, including future turns
  after death and invalid turns after an earlier tick. Direct transition input objects remain
  unchanged; real registry and successful advance/quit controls execute. The real Postgres
  composition also compares full saved state/revision/command history and released claims.
  Future-turn, partial-input, rejected-persistence and stuck-claim probes fail independently.
  This test-only correction awaits Claude review, not full A3/A5/AC5/AC6 acceptance.
- [x] A4 — AR-P1/P2/P3: `CatalogBundle.Arcade`, loader chain (`arcade` requires `minigame_api`),
  toy/definition/tenant cross-checks, resolver arms `(mine_grid|snake, 1.0.0) → arcade`, gameserver
  tenant registration, TS replay loader chain; fixture `minigames` + `minigame_api` candidates with
  the two AR2 rows and tenant rows.
- [x] A5 — composed platform witness (Postgres): `always` unlock, `human_hobby` lock at near-zero
  Soul, start → play → terminal → zero-credit applied resolution, `quit` releases the session.
- [x] A6 — client children `MineGridBoard` / `SnakeBoard` under `client/src/game-ui/minigame/`
  with the AR6.2/AR6.3/AR6.4 contract, test-only mounted (no pinned tenant row yet), axe in three
  browsers.
  RP-188's Codex test-only correction removes fixed-delay assumptions from batching/blur/resync
  and lead-limit witnesses. The real child and TS engine remain in the population; wall-only
  movement and each delivered callback are observed explicitly. Batch, blur, recovery and freeze
  mutations independently fail all three engines. Claude must review the corrective range;
  this checkbox records implementation presence, not A6 or full-RFC acceptance.
  RP-189's separate native keyboard supplement completes the existing corpus Mine Grid game
  in one mount and quits Snake through the real engines, with independently fired handler
  probes. It still needs Claude review and is not public task/assistive-technology acceptance.
- [x] A7 — docs (`docs/minigame-demo-disc-arcade.md`, platform pointers).

**Blocked (owner ruling pending, not built):** AR-P4 public API arms and AR-P5 tenant-registry
registration. Adding arms to the existing v1 minigame unions conflicts with API Foundation C2 (the
Typer lane's finding in `planning/minigame-terminal-typer/log.md`); the owner has not yet chosen
between new game-specific v1 operations (recommended), `/v2`, or a re-pin. Also not built here:
the arcade Game UI surface host (AR6.1/AR6.5), which needs the public wire; the OD-3 Fiscal
retirement and all production artifact bytes (AR8 mint).

## Codex correction review index (2026-10-05)

All listed ranges await **Claude's designated cross-party review**, not a recorded Codex first
filter. The original Claude A1–A7 work is not approved by these evidence supplements.

| Finding / bounded correction | Exact Codex span |
|---|---|
| RP-188: callback-driven scheduling evidence | `6652e467^..89434711` |
| RP-189: native-key Mine Grid completion and Snake quit | `9026bd40^..27622885` |
| RP-190: all attempted corpus commands in budget | `593ee762^..bcdee28d` |
| RP-191: actual nonterminal output and decoder/apply hidden-state refusal | `24b7b9d7^..fe49c408` |
| RP-192: watched TS snapshot value grammar, shared Go/TS negatives, kernel 0.3.144 | `8c01a8f5^..9288d1cc` |
| RP-193: raw Go/TS snapshot grammar, legal controls and kernel 0.3.145 | `d326fb2c^..45c35800` |
| RP-194: literal Go genesis/every-attempt snapshot/result byte witnesses, test-only v2 corpus | `c48ce7f2^..92ed5e68` |
| RP-195: both Snake raw/numeric/identity snapshot decoders, shared 45/7 controls, kernel 0.3.146 | `4f6173be^..45fcf3ae` |
| RP-196: test-only exact 5×5 clearing, complete strategy population and literal Go/TS replay | `329fc994^..25906f01` |
| RP-197: future/late-turn rejection, direct-input atomicity and real Postgres state/history/claim proof | `8ea62952^..5e82b3d3` |
| RP-198: real rejected shuffle draw, literal Mine Grid placement/entry replay and regeneration | `be0c0e00^..859f5216` |
