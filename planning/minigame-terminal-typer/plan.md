# Terminal Typer implementation plan

RFC: `rfc/minigame-terminal-typer.md` (accepted 2026-09-25; all ODs at recommended defaults).
Fixture-first (OD-11): no production epoch is minted by this plan.

## Batches

- [ ] B1 — Content + engine (TT3/TT4): `balance/testdata/typer-v1.json` fixture (placeholder
  prompts per OD-9), Go `server/typer` + TS `client/src/typer` loaders and pure engine,
  `ApplyInput.ServerTimeMs` (TT-PA1 item 2), shared content-gate corpus
  `testdata/typer/content-gate-v1.json`. ACs 1 (artifact half), 2, 4, 5, 6, 7, 10 (engine
  schema), 12, 15.
- [ ] B2 — TT-PA1 time plumbing: coordinator samples the DB clock once per admitted command before
  tenant execution; command-append inserts that exact value; replay passes persisted stamps. AC3.
- [ ] B3 — TT-PA2 `tier_at_least` unlock arm (Go/TS loaders) + composed start resolver
  (`not_eligible/tier_required`, `not_eligible/curriculum_exit_required`). AC9.
- [ ] B4 — TT-PA3 content resolver/bundle generalization + TT1 tenant row in a fixture minigames
  artifact + `minigame_api` loader chain. AC1 (definition/artifact/api chain half).
- [ ] B5 — TT-PA4 API arms (Typer snapshot/command arms) + generated client. AC10 (API half).
- [ ] B6 — UI: `TyperTable` child under `client/src/game-ui/minigame/`, tenant-registry row,
  accessibility contract TT8. ACs 13, 14.
- [ ] B7 — Composed platform path (real Postgres): create → begin → submits → terminal → faucet →
  receipt; neutrality timed/untimed. ACs 8, 11.
- [ ] Canonical docs (`docs/minigame-terminal-typer.md`), designated Codex review, archival (Codex).

Kernel protocol: every commit touching a `kernel/affecting-paths.json` prefix bumps
`kernel/VERSION` (+ Go/TS constants) in the same commit; new engine dirs are registered in the
commit that creates them.
