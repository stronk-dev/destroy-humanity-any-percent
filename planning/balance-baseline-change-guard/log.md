# Balance Baseline Change Guard Hardening — append-only log

## 2026-07-29 — start

- Independent review demonstrated that the guard inspects only HEAD, becomes a silent no-op with
  CI's two-commit checkout, and accepts unrelated source code inside a labeled baseline commit.
- This follow-up validates every reachable baseline commit, fails on shallow/dirty state, and makes
  later baseline commits generated-artifact-only. A3's Commons hash domain remains separate.
