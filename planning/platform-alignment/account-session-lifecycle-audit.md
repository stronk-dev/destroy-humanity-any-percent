# Account & Session Bootstrap lifecycle audit

Coordinate: product tree `190a4fa`; audit checkpoint after `f6fdedc`; 2026-08-20.

This pass re-derived the active Account RFC from its full specification, plan/log, migrations,
repository and HTTP implementation, account/transport/gameserver integration tests, current Game
UI runtime, Leaderboards consumers, canonical docs, later bootstrap/GC/composition successors, and
tracked review history. It did not edit product code, owner-authored RFC/design text, canonical
product docs, or implementation-plan checkboxes.

## Bottom line

The backend account/security foundation is substantial and cold-green: anonymous account creation,
Argon2id recovery authentication and upgrades, exact JWTs, refresh-family serialization and theft
detection, database-backed access revocation, Founder archival/replacement, normalized local-save
import, anonymizing deletion, bounded request limiters, credential garbage collection, and a real
Postgres→Production→HTTP→WebSocket bootstrap path all exist.

The production player session is nevertheless incomplete:

- the Game UI stores the 15-minute access token and 30-day refresh token but never calls the
  refresh endpoint, replaces credentials, or reconnects the socket with a new token;
- expiry or revocation is collapsed into the generic `offline` state, leaving valid recovery and
  refresh material stranded in local storage with no recovery action;
- the one-time recovery code is silently persisted but never shown/downloaded, and no player UI
  invokes recovery login, import, New Founder, account deletion, or optional email attachment;
- the adopted design makes silent server-anonymous play the default and local-only play the outage
  fallback, while Account D4 still presents fully offline-anonymous play as a peer mode and the
  fallback runtime does not exist;
- AC2, AC3, AC5, AC6, and AC7 retain narrower witnesses than their literal requirements; and
- the active review record names obsolete, non-resolving hashes, predates explicit review
  provenance/range-union law, and does not reconcile later bootstrap, GC, transport, and UI work.

This is a proven backend foundation with an unshipped account/session lifecycle. It is not an
archival candidate until the accepted backend criteria and record are closed; the missing player
rights/recovery/fallback surfaces belong to an accepted successor rather than silent RFC widening.

## Current cold evidence

All commands ran from repository root:

- `make test-go GO_PACKAGES='./account ./transport ./gameserver' GO_TEST_FLAGS='-count=1'` — green;
- `make test-save-integration SAVE_TEST_PACKAGES='./account ./leaderboard ./gameserver'` — green on
  real PostgreSQL;
- `make test-game-ui-composed` — real Chromium bootstrap→live snapshot→Centrifuge presence
  handshake green;
- the current product coordinate's client run remains green: 39 files passed, two skipped; 6,655
  tests passed, 15 skipped.

The account integration fixture rejects extra response fields, runs a real Production intent,
forces refresh-family replay and concurrent rotation, performs New Founder/import/delete, and
exercises an unauthenticated limiter. The composed gameserver fixture proves production bootstrap,
socket subscription, New Founder integration, and session GC. These witnesses validate the
mechanical foundation; they do not supply the missing browser lifecycle or every literal AC arm.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Proven integration** | Real Postgres account creation and recovery session feed a strictly decoded HTTP response, authenticated real Production intent, and applied receipt. Request decoders reject unknown/trailing input and the test decoder rejects unknown/trailing response fields. The composed bootstrap successor additionally proves the current production binary/browser entry. | Preserve the strict negative controls and include current-history account/composition ranges in the final review union. |
| AC2 | **Proven integration; player refresh successor open** | Q-001 connects a real Account-backed socket, revokes its family, proves subsequent subscribe/alive failure, rejects revoked-before-connect, and retains an unrevoked control. The mutation bypassing revocation fails; Claude approved at `34d04a5`. | Preserve the witness. The player successor must still rotate before ordinary expiry and reconnect safely (RP-048). |
| AC3 | **Proven integration** | Q-001 performs second and third free Founder replacements, preserves/reads every archived stream, verifies exact catalog initials, and byte-compares unchanged cost/cooldown state. A one-use mutation fails; Claude approved at `34d04a5`. | Preserve the bounded repeated lifecycle in the final Account union. |
| AC4 | **Proven integration** | A signed token with exactly the five claims verifies; expiry and previous-key rotation work; a correctly signed sixth `role` claim is rejected. | Preserve this discriminating unit witness and current token implementation in the final review range. |
| AC5 | **Proven integration** | Q-001 starts at the actual import operation, plays through Exit/verification/projection, proves the imported Founder has no board row, and retains a server-created projection control. A producer/flag severing mutation fails; Claude approved at `34d04a5`. | Preserve the literal import-to-consumer chain in the final Account/Leaderboards union. |
| AC6 | **Proven integration** | Q-001 deletes an account with archived, active and imported Founder/Company streams, reloads every stream anonymized, and proves account-linked credentials/bootstrap secrets absent. The current-stream deletion mutation fails; Claude approved at `34d04a5`. | Preserve the all-stream/secret inventory; retention duration and player disclosure remain owner work. |
| AC7 | **Proven integration** | Q-001 independently exhausts account creation, recovery-session creation and refresh, requires exact typed rejections with no rejected-request mutation, then proves refill admits the next request. Limiter bypasses fail; Claude approved at `34d04a5`. | Preserve all three mounted-route controls. |

No plan checkbox was changed. Plan item 6 is historically true for its 2026 slice; item 7 remains
open and cannot be closed with the old non-resolving review prose.

## Fifteen-minute player-session defect

The backend has a correct refresh mechanism, but the production consumer does not use it:

1. D2 issues a 15-minute access token and a 30-day refresh token.
2. `client/src/game-ui/runtime.ts` stores both tokens plus account/recovery data in one localStorage
   document after bootstrap.
3. Every snapshot and intent reads the same stored access token. No code updates that document or
   calls `/api/v1/session/refresh`; repository-wide client search finds `refreshToken` only in the
   interface, validator, and bootstrap write.
4. WebSocket connect captures the same access token. Centrifuge rejects protocol refresh, and the
   server's alive hook disconnects once database authentication fails or the JWT expires.
5. `GameUIApp.svelte` maps failed HTTP calls or socket close to a generic `offline=true`. It offers
   no refresh/login/recovery transition and does not distinguish expired credentials from server
   unreachability.
6. The composed browser test ends immediately after bootstrap, one snapshot, and presence; its
   oracle cannot cross the 15-minute boundary. Client unit tests use inert string tokens and never
   simulate 401→rotation→retry→socket reconnect.

Consequently the current browser session has a server-enforced fifteen-minute horizon even while a
valid refresh token remains stored. A valid successor witness must advance the authoritative clock,
observe proactive or 401-triggered single-use rotation, persist the new pair before use, reconnect
and resubscribe with the new access token, and prove the consumed token cannot create a second
descendant.

## Producer→consumer reality

| Layer | Reality at HEAD |
|---|---|
| Account/credential authority | Proven: minimal relational account, Argon2id recovery hash, exact JWT/access rows, serialized refresh families. |
| Founder lifecycle | Mechanically strong: create, archive/mint, import normalization, permanent imported flag, deletion anonymization. Literal repeated/import-to-board/all-stream witnesses remain incomplete. |
| HTTP/gameserver | Proven: mounted account/session/founder/state/intent/bootstrap routes and real Postgres composition. Optional email intentionally returns `not_configured`. |
| WebSocket auth | Proven mechanically at connect/alive and membership authorization. Revocation-to-live-socket acceptance remains uncomposed. |
| Production bootstrap consumer | Proven: Vision Slide silently creates a server account and persists credentials before play. |
| Long-lived client session | Broken/incomplete: no refresh, credential renewal, reauthentication, or expiry recovery. |
| Player account controls | Absent: no recovery-code disclosure/backup, recovery login, import, New Founder, delete, or email surface. |
| Local outage fallback | Absent: no local authoritative save/runtime or later import workflow in the production client. |

## Normative, canonical, lifecycle, and operational drift

1. Account D4 and `design/06` still describe fully offline-anonymous play broadly. The later owner
   ruling in `design/11 §1b` makes silent server-anonymous the default and local-only play a labeled
   outage fallback. The Account body was not reconciled and the fallback is unimplemented.
2. D1 says the recovery code is “shown once.” The API returns it once, but the Game UI silently
   stores it and exposes no display/download/confirmation workflow. Backend delivery is not user
   possession.
3. Canonical account docs accurately describe backend mechanics and later credential GC, but do
   not disclose that the production client cannot refresh a session or invoke account lifecycle
   operations. Game UI docs call credential persistence shipped without the corresponding expiry
   lifecycle.
4. The Account plan/log stop in July. The recorded reviews cite `bdcc9a1` and `2cdf1be`, neither of
   which resolves in current history, and omit `Review by:`/`Recorded by:` plus exact range union.
   Current implementation hashes include rewritten account work and later bootstrap, composition,
   GC, API, transport, Leaderboards, and Game UI successors; their separate reviews are not a final
   Account archival verdict.
5. Anonymous creation has per-IP in-memory limiting but no account quota/reaper. The original
   adversarial review explicitly left storage amplification open. Session/bootstrap secrets now
   have collectors; abandoned anonymous account/save trees do not. Release-safe abuse/retention
   ownership remains unspecified.

## Smallest honest closeout order

1. Ruling author reconciles D4 with the server-default/local-fallback owner ruling; keep the absent
   fallback and player rights/recovery surfaces routed to the existing account successor decision.
2. Add only AC2/AC3/AC5/AC6/AC7's missing literal backend witnesses under the accepted RFC, with
   demonstrated failing mutations.
3. Accept and implement the client session/recovery successor: refresh rotation, credential
   persistence ordering, socket reconnect/resubscribe, recovery-code possession, recovery login,
   and honest expired/revoked/offline states.
4. Keep export, account deletion UI, local fallback/import UI, and optional email recovery behind
   their explicit owner decisions; do not infer disclosure/copy or retention policy.
5. Decide anonymous-account abuse retention/reaping before calling deployment safe.
6. Reconcile canonical docs and the stale plan against later bootstrap/GC/composition work.
7. Construct exact current-history ranges for all Account work consumed and obtain the mandatory
   tracked cross-party verdict before archival.
