# Founder Attendance Foundation implementation log

Append-only. A fresh agent must be able to resume from this file and the accepted RFC alone.

## 2026-08-05 — implementation opened

- Owner accepted A1-A5 in `rfc/founder-attendance-foundation.md`.
- Existing `age_ms` remains the only persisted completed-run attendance authority.
- First batch is the Pet-C9 prerequisite: immutable Founder genesis plus Exit participation in
  `founder_log`, followed by career-long Founder replay. Attendance consumers do not land before
  that replay boundary is proven.
- The malformed kernel registry history was repaired pre-push as explicitly approved; the
  corrected commits are `70d2d4a` and `374c7c0`. `node client/tools/verify-kernel-version.mjs`
  passes at kernel `0.3.27`.

## 2026-08-05 — immutable Founder genesis implemented

- Migration 00056 adds one immutable Founder genesis per Founder stream. Existing Founder logs
  backfill only from the exact pre-command revision named by their first replay envelope; missing
  retained state aborts migration rather than fabricating history.
- A deferred constraint makes every Founder-log insert uncommittable without genesis and binds
  sequence 1 to the genesis revision.
- `ApplyFounderLogged` creates genesis from the exact loaded pre-command bytes in the same
  transaction as the first applied or rejected log row. Mutation errors roll the genesis back.
- The real-Postgres fixture proves byte identity, immutability, and rejection of a direct
  genesis-less log insert. Root save integration is green.
- Kernel registry grows with `server/save/foundergenesis.go`; semantic version is `0.3.28`.
