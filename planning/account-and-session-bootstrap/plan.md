# Account & Session Bootstrap — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/account-and-session-bootstrap.md`
- **Started:** 2026-07-29

1. [x] Add account/founder/session/access-token migrations and repository.
2. [x] Implement recovery credentials, exact-claim JWTs, refresh-family rotation, and revocation.
3. [x] Implement New Founder, import, deletion, and active-Founder state initialization.
4. [x] Add the closed chi API including authenticated Production intent submission and full state.
5. [x] Add rate limits, strict wire validation, real-Postgres integration, and security tests.
6. [x] Update canonical docs and run full verification.
7. [ ] Record independent review before archival.

## Q-001 — accepted-scope witness closeout

Authorized by `planning/platform-alignment/ready-batch-manifest.tsv` after the designated
platform-alignment review. This batch changes tests and evidence records only; it must not change
Account semantics, API schemas, production behavior, owner copy, retention policy, or archival
state.

8. [ ] AC2: compose real Account family revocation with a connected socket, a post-revocation
   subscribe/alive rejection, a revoked-before-connect rejection, and an unrevoked control.
9. [ ] AC3: repeat New Founder through second and third replacements; prove every old stream stays
   readable/archived and every replacement starts at exact catalog initials with no cost/cooldown.
10. [ ] AC5: begin with the actual import operation, carry the imported Founder through Exit and
    verification, and prove board projection refuses it while a server-created control projects.
11. [ ] AC6: delete an account containing archived, active, and imported Founder/Company streams;
    prove every stream survives anonymized while every account-linked credential/bootstrap row is
    absent.
12. [ ] AC7: exhaust account creation, recovery-session creation, and refresh independently; assert
    exact typed `rate_limited` responses, no rejected-request mutation, and an allowed refill
    control.
13. [ ] Run focused unit tests plus sequential Account, Gameserver, and Leaderboards cold Postgres
    populations; record one-seam failing mutations, diff review, and exact range for Claude.
