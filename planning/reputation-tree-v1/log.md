# Reputation Tree v1 — implementation log (append-only)

## 2026-09-25 — Predeclaration B1–B2 (Claude)

**Implemented by:** Claude. Awaiting Codex designated cross-party review for every range below;
nothing here is self-approved.

B1 predeclares: a strict `reputation_tree` loader in Go and TS enforcing R2 rules 1–8 (rule 8's
copy-key registration is enforced against the declared copy-key set, as Soul/Pitch do; the
`copy/references.v1.json` registration lands with the production artifact at mint), accounting
helpers (available = level − spent; derived unlock ppm), and the R3 factor
`1 + level×per_level_ppm×unlock_ppm / 1e12`. One Go-authored corpus drives both runtimes:
`testdata/reputation/tree-fixtures-v1.json` (one rejection fixture per rule) and
`testdata/reputation/bonus-vectors-v1.json` (AC5 vectors). Failing cases to demonstrate: deleting
each rule's check makes its fixture load; computing from `available` instead of `level` fails a
vector.

B2 predeclares the bundle wiring described in the plan. The fixture tree (the RFC's proposed R2
table) lives in `balance/testdata/reputation-tree/fixture-v1.json`; no production artifact or
epoch is created.

## 2026-09-25 — B1 landed (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

- **Delivered:**
  - `server/reputation/tree.go` and `client/src/reputation.ts`, both registered as kernel-guarded
    paths. The kernel version bump is in the same commit.
  - `curriculum.ValidateStarter`, exported as a pure wrapper so the tree reuses the exact starter
    grammar.
  - The fixture tree, the 24-case rejection corpus and the 10 Go-authored bonus vectors.
  - Placeholder node copy.
  - `docs/reputation-tree.md`.
- **Evidence, all runs cold:** `go test -count=1 ./reputation ./curriculum ./kernel` passes, and
  `vitest run test/reputation.test.ts` passes 4/4.
- **Severing probes:** each breaks one check, and each turned its test red; every file was then
  restored.
  - Go: rule 1 (whole declaration, provider only), rule 2, rule 3, rule 4, rule 5, rule 6
    (monotonic, final rung), rule 7 (whole headroom, resource branch, generator branch, duplicate
    upgrade), rule 8, and the AC5 mutant that computes from available instead of level.
  - TS: rules 1–8 and the AC5 mutant.
- **Finding during probing:** the TS rule-1 provider-only mutant first survived, because no fixture
  bound a declared row that had the wrong provider. I added `bonus_source_wrong_provider`, which
  uses `fiscal.hoard`; both the Go and TS provider mutants now fail.
- **DESIGN-GAP RT-DG-A:** R2 rule 8 names `copy/references.v1.json` registration. The copy
  reference registry is keyed to epoch artifact schemas, and this artifact has no epoch or schema
  yet. The loader enforces the declared-copy-key half now. Registration, and
  `balance/reputation-tree.schema.json`, land with the production artifact at the mint (B-mint),
  and the gap is recorded here rather than improvised.

## 2026-09-25 — B2 landed (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**What landed:**
- The Go and TS bundle loaders accept `reputation_tree`, which requires the `minigame_api` chain.
  The pairing rule applies in the loaders and in Go's `valid()`.
- The frozen-contribution resolver accepts provider `reputation_tree`.
- Tests: `server/replaycatalog/reputation_test.go` and
  `server/production/reputation_contributions_test.go` (both against the live epoch-8 artifact
  set), plus `client/test/reputation-bundle.test.ts`.
- Kernel version 0.3.108 → 0.3.109, bumped in the same commit.

**Evidence (cold):** `go test -count=1` passes for `./reputation`, `./replaycatalog`,
`./production`, `./curriculum` and `./economy`. `make typecheck test-client` passes (6702 tests).

**Severing probes.** Each check below was disabled and turned its test red, then was restored:
- loader pairing and loader chain, in Go and in TS;
- the `valid()` pairing check, in Go;
- the resolver's provider acceptance, in Go.

**Probe corrections:**
- The first resolver sever failed only because the build broke (an unused import). I redid it as a
  mutant that compiles and passes vet, and it still turned the test red.
- My first `valid()` pairing test was vacuous: the hash/count mismatch masked it. I replaced it with
  an isolated case. A live bundle stays valid until only its Economy is swapped for one that
  declares the source.

**Deliberately not done:** B2 raises no Founder version floor, because v22 lands in B3. No epoch
pins a tree, so this is not a live hole.

**Unrelated finding:** `server/replaycatalog/catalog_test.go` at HEAD (committed in `391beb76`) is
not gofmt-clean. Its owner or review should fix it; I left it untouched.
