# Balance Baseline Change Guard Hardening — append-only log

## 2026-07-29 — start

- Independent review demonstrated that the guard inspects only HEAD, becomes a silent no-op with
  CI's two-commit checkout, and accepts unrelated source code inside a labeled baseline commit.
- This follow-up validates every reachable baseline commit, fails on shallow/dirty state, and makes
  later baseline commits generated-artifact-only. A3's Commons hash domain remains separate.

## 2026-07-29 — implementation

- The guard now rejects shallow repositories and dirty generated artifacts before inspecting the
  complete reachable history of pacing-baseline commits. Only the oldest bootstrap commit is
  exempt.
- Every later revision validates its own subject and artifact-only diff, plus the input interval
  from the previous baseline revision to the candidate's first parent. Cover commits are irrelevant.
- Temporary real-git repositories reproduce the covered bad commit and shallow checkout, reject
  source smuggling and missing prior inputs, and accept the intended input-then-artifact sequence.
- Server CI now requests full history. Commons inputs remain outside this guard until A3 changes the
  constants-hash domain deliberately.
- The tightened guard accepts the repository's complete real baseline history, and the read-only
  deterministic `make harness-check` passes against the committed artifacts.
- Full `make verify` is green: Go vet and every Go/Postgres package, formula and balance gates,
  TypeScript checks/build, 6,412 client tests, schema/boundary checks, and 19,245 browser tests.
  The implementation now waits at the mandatory independent review gate.
