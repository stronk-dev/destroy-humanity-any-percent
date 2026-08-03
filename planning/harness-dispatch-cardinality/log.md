# Harness Dispatch Cardinality — maintenance log

## 2026-08-03 — independent review

- **Review by:** Darwin
- **Recorded by:** Codex
- **Reviewed range:** `8f6885c^..8f6885c`
- **Verdict:** approved, no findings.

The test-only hardening extracts the existing run-key construction and dispatch loop without
changing production semantics, then proves that four overlapping workers execute every one of 17
complete unique run keys exactly once. Returned reports are checked against their declared task
keys and fail closed on a mismatch. The focused race-enabled test and the full 300-run
`make harness-check` path pass; committed harness artifacts remain byte-identical.

