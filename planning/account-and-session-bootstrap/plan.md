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

8. [x] AC2: compose real Account family revocation with a connected socket, a post-revocation
   subscribe/alive rejection, a revoked-before-connect rejection, and an unrevoked control.
9. [x] AC3: repeat New Founder through second and third replacements; prove every old stream stays
   readable/archived and every replacement starts at exact catalog initials with no cost/cooldown.
10. [x] AC5: begin with the actual import operation, carry the imported Founder through Exit and
    verification, and prove board projection refuses it while a server-created control projects.
11. [x] AC6: delete an account containing archived, active, and imported Founder/Company streams;
    prove every stream survives anonymized while every account-linked credential/bootstrap row is
    absent.
12. [x] AC7: exhaust account creation, recovery-session creation, and refresh independently; assert
    exact typed `rate_limited` responses, no rejected-request mutation, and an allowed refill
    control.
13. [x] Run focused unit tests plus sequential Account, Gameserver, and Leaderboards cold Postgres
    populations; record one-seam failing mutations, diff review, and exact range for Claude.

Q-001 received the designated cross-party approval at `34d04a5`. This closes the bounded witness
batch only; item 7 and the broader Account player/rights/retention/archive work remain open.

## RP-048 / RP-234 — browser session-consumer diagnosis (2026-10-06)

Separate from Garden's pending `b58277cb..bff05b5e` range. Account D2/D3 and
Transport T1/T4/T5 authorize observing the existing credential/refresh/recovery
boundaries, not inventing browser renewal policy. Q-003 explicitly excluded it.
The next batch is test instrumentation, a draft follow-up and honest records only.
No production authentication, TTL, family locking, replay/revocation, kernel,
copy, catalog, CI topology, lifecycle or completion-checkbox change is authorized.

Use the real built client, gameserver, WebSocket and declared disposable Postgres
service. After actual bootstrap/Fiscal purchase, expire only that account's exact
access-token row; do not revoke or consume its refresh token. Observe actual
Garden HTTP 401 and socket closure, with no browser refresh request. Then perform
one explicitly test-operated HTTP refresh, store its pair and reload the real
client: the same Founder/Garden must recover with unchanged gameplay heads.
Predeclare healthy-before / expired / healthy-after populations, exact row counts,
no automatic-account replacement and no token/copy/hidden-state output. This is
controlled early database expiry, NOT natural JWT expiry or full timed Garden proof.

Retain a bounded manual diagnostic mode in the existing composed driver, without
changing its normal three-tick objective. Execute the counterfactual with expiry
severed, restore instrumentation byte-exact and rerun. Missing client policy goes
to a draft RFC linked to Account, Transport, API Foundation and archived Game UI;
the draft must not masquerade as accepted implementation authority.

## Existing refresh contract census — real-Postgres stage (2026-10-06)

Predeclared d33b5d56 after parser span ee8b2356..c2843089. Test-only preparation,
**Executed on native macOS against the composed lane's Postgres on 2026-10-07**. Seven populations:
live/access-expired rotation plus reuse, New Founder binding, unknown canonical
token, exact refresh expiry, actual closed-DB error, limiter non-mutation/refill.
Actual repository/API/Postgres fixture with HTTP over net.Pipe, not browser/OS
socket/production-engine proof. Ten named gameplay/account tables and three
credential tables snapshotted privately, tokens/data never logged.

The normal `make test-game-ui-composed` lane now requires all eight to run/pass using the
existing completeness observer, then resets the test database before other epochs/player flows.
The whole lane passed; the actual missing-DB control exits 1 despite Go/package exit 0.
The eighth case (2026-10-07) drops every response byte after actual committed rotation and proves
client EOF is not rollback: replay revokes all three generations without changing gameplay rows.
Account/publicapi cold non-DB tests and vet pass. This establishes these eight real-DB cases,
not the entire Account/publicapi real-DB population, browser renewal or Linux execution.

Pending: designated review of the prepared tests/observer and their new composed-lane binding;
generated consumer and browser policy remain separate unaccepted work. RP-236 still blocks the
image-based Linux route, not this proven native route. Under the current RFC-0000 procedure,
do not repeat unchanged controls or add source mutations merely to satisfy the superseded
per-edit routine; explicit Account acceptance criteria remain unchanged.
