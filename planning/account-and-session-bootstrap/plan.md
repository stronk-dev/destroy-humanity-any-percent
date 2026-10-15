# Account & Session Bootstrap — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/account-and-session-bootstrap.md`
- **Started:** 2026-07-29

Current D2/AC4 correction (2026-10-13, RP-542): private JWT decoding validates original
object members against its DTO tags before Go can merge duplicates or fold key case.
Signed old-source regressions, legal rotation/encoding controls and mounted HTTP pre-DB
tripwires discriminate; cold affected packages/vet, race and API generation pass. The
owning log records independent corrective APPROVAL of `cf8c6f08^..cf8c6f08` under Marco's
2026-10-15 delegation. Issuance/signature/TTL/rotation,
authorization, public schema and kernel bytes are unchanged; no SQL/whole Account acceptance.

Current bounded witness correction (2026-10-08, RP-474): a real Linux/Postgres run exposed
the raw peer's missing application-level pong, closing3012 before the expected auth denial.
Both raw read helpers now acknowledge protocol pings; forced ping/batched-reply regressions
and the unchanged real revocation4001/healthy-drain4003 assertions pass. This is test-only,
not a production/auth/timing change or closure of the separate RP-450 drain failure.
Fresh independent Codex corrective review APPROVES completee4fd61fa..da21dae2;
the owning log holds exact provenance and actual Linux-amd64/Postgres re-execution.

Current bounded witness work (2026-10-08, RP-450): the hosted deadline occurs in drain after
the revocation assertions, not in their denial checks. The unrevoked healthy control now reads
through teardown and must observe code 4003 with the existing deadline. Old focused controls
also pass, so the original intermittent failure remains open rather than declared solved.
Fresh independent Codex corrective review APPROVES completedcd9475d..b0f26d47;
the owning log records exact provenance. RP-450 stays open; no authentication or timeout change.

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

## R-011 interrupted-dispatch primitive follow-up — 2026-10-12

At `a6fc7217`, observe native Web Locks/localStorage and real loopback HTTP, using synthetic
generations only. Question: does a marker written before dispatch survive actual owner-page
closure and cause a queued/reloaded successor to stop without redispatch? Four arms, five
repetitions per selected engine: healthy overlapping tabs; close after the server receives
the request but before its reply; the same closure without the marker (failure control);
native localStorage quota denial before dispatch. The default population is Chromium,
Firefox and WebKit (60 cases); an explicit native Chromium/WebKit subset is40 cases,
not full supported-browser completion. Use fresh test-owned contexts for each case.

Observe actual server request counts, native queued-lock state, complete callback results,
page interruption, marker/generation persistence and real quota exceptions. Healthy requires
one dispatch and generation2 in both tabs; interrupted requires one dispatch and stopped
queued/reloaded successors with generation1 intact; marker omission must produce two
dispatches; quota denial must produce zero. Missing prerequisites, incomplete cases, uncaught
errors or the existing30-second observation guard invalidate the experiment. Native quota
filling is bounded to16MiB; reaching that guard is invalid, not a measured capacity claim.

This is a test-only candidate with a synthetic HTTP receiver, not real Account rotation,
browser-process/OS crash, production storage durability, renewal code or an adopted marker
schema/lifetime. It may refine the draft's primitive evidence; it cannot authorize auth,
storage/recovery policy, timers, gameplay retries, full R-011 or the15-minute Garden claim.

Completed bounded native subset:40 Chromium/WebKit cases pass with the declared positive,
interruption, omission and quota outcomes. Firefox launch executes zero cases and is invalid;
full60-case coverage remains open. Evidence lives in the existing coordination dossier and
latest owning log. The draft now distinguishes this primitive result from the remaining real
credential/Account/process-lifecycle and owner-policy gates; no renewal implementation follows.

## R-011 real Account/browser follow-up — 2026-10-12

Baseline `e497cae2`: reuse the Garden driver's actual gameserver, pinned fixture and declared
disposable Postgres initialization, but not its gameplay or15-minute objective. Three fresh
browser contexts/accounts, one case each: healthy queued overlap, owner-page closure after
the real refresh response has committed but before downstream delivery, and identical closure
with the test marker omitted. Test-operated browser fetches call unchanged bootstrap/refresh
routes through a loopback proxy. No automatic coordinator is wired into the production client.

Require actual200 rotation and private Postgres session/access/family counts before releasing
or discarding its reply. Healthy sends once and both tabs use the same stored descendant;
interrupted sends once and queued/reloaded tabs stop while the family remains unrevoked;
omission sends twice, receives exact401 refresh_reused/session_family_revoked and revokes
both access generations. Capture all ten existing refresh-census account/game tables privately
before/after each case; require exact unchanged rows. No tokens, recovery codes, request/reply
bodies or private snapshots enter diagnostic output. The existing30-second observation guard
invalidates incomplete cases; no retries, new policy deadlines, clock/TTL edits or balance grants.

This can establish bounded Chromium→real Account→Postgres ambiguity/exclusion evidence, not
natural expiry, real production credential documents, automatic reconnect, supported-engine
completion, OS/process crash, marker adoption or full R-011. A focused manual Make target
runs it outside CI. Finish with the existing controlled-maturity Garden journey to check that
the new diagnostic branch leaves the normal driver path intact.

Completed: all three native Chromium/real Account/Postgres cases pass, including actual
committed rotation and exact reuse/family revocation without the marker. The ten watched
account/game tables remain unchanged. Existing controlled-maturity Garden journey also
passes through the normal branch. The coordination dossier and owning log retain the
initial invalid key attempt, exact source identity, commands and limits. Remaining work
is supported-profile/lifecycle and production credential persistence evidence plus draft
policy adoption, not repeating these three completed cases or claiming renewal implemented.
