# Account & Session Bootstrap — append-only log

## 2026-07-29 — start

- Owner accepted the new root RFC through the ordered batch manifest. HTTP owns intents; WebSocket
  remains streams-only. Imported founders are permanently unranked.
- The handed-off contract names Argon2id but not work factors. This security configuration is pinned
  to current OWASP minimum guidance: 19 MiB memory, 2 iterations, parallelism 1, 16-byte random salt,
  32-byte output. The encoded hash carries algorithm/version/parameters for later rehash upgrades.
- Account and session state is relational; gameplay state continues through the existing owner-aware
  Save Store. The implementation will not place PII or membership claims in JWTs or save payloads.

## 2026-07-29 — implementation

- Added migration 00010 with a deliberately three-column `accounts` table, optional email side
  table, one-active-Founder ownership history, refresh-token families, and revocable access-token
  records.
- Added 128-bit recovery codes with Argon2id storage, exact-five-claim HS256 JWTs with current and
  previous key acceptance, 15-minute access TTL, 30-day opaque refresh TTL, single-use rotation,
  and family-wide replay revocation.
- Added atomic account initialization, free New Founder archival/replacement, one-time local-save
  import through `save.RestoreState`, imported-Founder marking, and anonymizing account deletion.
- Added the closed chi `/api/v1` surface. It strictly decodes request keys, applies IP/account token
  buckets, resolves active ownership on every request, and mounts authoritative Production intent
  handling without accepting a client-supplied stream ID.
- Unit tests are green. The real-Postgres integration passed account creation, recovery login,
  exact response decoding, a real manual Production intent, refresh replay revocation, archived
  Founder readability, import marking, deletion with save retention, and typed rate limiting.
- Implementation review found that legacy imported saves need an authoritative migration baseline;
  imports now use canonical server import time instead of epoch zero. The current-version integration
  path remains unchanged.
- Full `make verify` passed with Postgres: Go vet and all server packages, generated-formula drift,
  both harness gates, TypeScript/Svelte checks, production client build, 6,412 client tests, schema
  and package-boundary verification, and 19,245 browser tests.
- The RFC remains `implementing` until the required independent diff review is recorded. Forward
  batch work may build on the committed surface without pretending self-review satisfies that gate.

## 2026-07-30 — independent review: account & session bootstrap (bdcc9a1) — the review I owed

Adversarial two-lens review; three HIGHs reproduced against live Postgres by the review harness and
re-verified against source by the reviewer. **Verdict: the credential/token core is genuinely
strong (see verified list in the review record: alg-pinned exact-claims JWT, two-key rotation,
crypto/rand + Argon2id at OWASP minimums, constant-time compares, sequential single-use rotation,
one-active-founder under lock, unforgeable server-set `imported`, zero-PII schema asserted against
information_schema, fail-closed origin/limiter). The session-family model and the import path are
NOT acceptable. Rulings below are now normative in the RFC.**

Findings (fix queue, ordered):

1. **HIGH — family revocation is not atomic; a concurrent refresh escapes it** (`store.go:492-498`).
   READ COMMITTED: the bulk family UPDATE's snapshot cannot see a session row inserted by a
   concurrently-committing rotation, and `RefreshSession` never validates family state — so replaying
   a consumed token at the moment the legit client rotates leaves a live, unrevoked refresh chain;
   theft detection and logout both defeated (reproduced: post-revocation token minted a fresh pair).
   **Ruling A-D2a:** add `session_families(family_id PK, account_id, revoked_at)`; issue creates the
   row; every refresh and every revocation takes `FOR UPDATE` on the family row first; refresh
   rejects when `revoked_at` is set. Serializes rotation against revocation completely.
2. **HIGH — import bricks the company stream** (reproduced twice): (a) a locally-prestiged save
   (`run_seq>1`) imports fine and then every intent 409s forever (no pin for that seq); (b) a
   client-supplied older `constants_hash` writes a revision that mismatches the run-1 pin
   (`run identity mismatch` forever). The direct `INSERT INTO save_revisions` bypasses the store
   path, contradicting D4's own text. **Ruling A-D4a:** the request's `constants_hash`/`version`
   select only the migration catalog; the server re-encodes the restored state under
   `repository.constantsHash`, resets `RunSeq` to 1 with a fresh `RunStartedAt` (imported founders
   are ranked-excluded, so run identity is server-owned), and writes through the standard store path
   so the existing pin holds by construction. AC: import-then-intent round-trip with a run_seq=3,
   old-hash fixture.
3. **MEDIUM — account deletion cascade destroys founder rows and the `imported` flag**
   (`00010_accounts.sql:15` ON DELETE CASCADE), which also turns any queued verified-run projection
   for that founder into a permanent hard error (`boards.go:73-77`). RFC D5 says "archives".
   **Ruling A-D5a:** `account_founders.account_id` becomes nullable with ON DELETE SET NULL; the
   delete transaction stamps `archived_at`; founder rows and `imported` survive anonymized.
4. **MEDIUM — Argon2 parameter bump would lock every account out**: `verifyRecoveryCode` hard-pins
   current constants (`credential.go:66-68`) and no rehash-on-verify exists, contradicting the
   planning log's own claim. **Ruling A-D1a:** verify with the STORED parameters subject to a floor
   (m ≥ 19 MiB, t ≥ 2, p ≥ 1); on successful verify with non-current parameters, rehash to current
   in the same transaction.
5. **MEDIUM — account-existence timing oracle on session creation** (measured 72×: 266 µs miss vs
   19.3 ms hit): dummy-verify a constant hash on the miss path.
6. **MEDIUM — per-IP limiter degrades to one global bucket behind any proxy** (`RemoteAddr` only;
   correct that XFF is untrusted, but deployment needs a trusted-proxy hop count in config) and its
   bucket map is unbounded (no eviction; IPv6 /64 growth). Fix: config `trusted_proxies`, LRU/TTL
   eviction, and apply an IP bucket to failed-auth requests on authenticated routes (currently
   unthrottled).
7. LOW — unknown paths/methods return chi plaintext, not the typed shape (set NotFound/
   MethodNotAllowed handlers). LOW — no session/access-token GC (~2.9k rows/user/month; add prune
   with the existing prune-job pattern). LOW — recovery-code input should normalize case/whitespace
   before validation, and the create response needs `Cache-Control: no-store`. LOW —
   `ErrAccountNotFound` from NewFounder maps to 500 instead of a typed rejection.
8. OBSERVATIONS — stale `fid` authorizes the archived founder's channel ≤15 min (RFC-sanctioned;
   transport may optionally re-resolve on subscribe); unauthenticated account creation is unbounded
   storage amplification (quota/reaper is a named follow-up); the surface is honestly unmounted
   (composition is the gameserver work already queued).

**Record correction:** the review harness flagged `GET /api/v1/founder/state` as a closed-surface
violation; it is NOT — the 2026-07-29 transport T4 ruling authorized it as the full-sync endpoint.
The account RFC's D3 list is amended to include it; the bookkeeping gap was mine.

## 2026-07-30 — remediation: family serialization and authoritative import

- Added `session_families` as the single lock and revocation authority. Refresh and revoke resolve
  the family, take its `FOR UPDATE` lock, then re-read and validate the presented token. A real
  Postgres test forces a legitimate rotation and consumed-token replay to queue behind the same
  family lock and proves the resulting family has no live refresh or access rows.
- Import now treats the submitted version/hash only as migration inputs, resets server-owned run
  identity to run 1 at the canonical import instant, and writes under the current hash through
  `save.Store.WriteInTransaction`. That store entry point shares Write's validation, stream lock,
  compare-and-swap, and retention policy while keeping the imported marker atomic with revision 2.
- The old-hash, run-3 reproducer now imports as current-hash run 1 and successfully commits a real
  Production intent at revision 3.
- `make test-go GO_PACKAGES='./account ./save'`, matching `go vet`, and the complete account suite
  against local Postgres are green.

## 2026-07-30 — MEDIUM security and anonymization remediation

- Recovery verification now parses the stored Argon2id parameters, enforces the 19-MiB/2-pass/1-lane
  floors, and flags any valid non-current hash for replacement. Login locks the account row, verifies,
  rehashes with fresh salt, and issues the session in one transaction. Missing valid-shaped account
  IDs pay an Argon2id verification against a constant dummy hash. Recovery input is case/outer-space
  normalized.
- Migration 00019 changes Founder ownership deletion from cascade to nullable `ON DELETE SET NULL`.
  `DeleteAccount` stamps every identity's archive time before deleting the account; the imported flag
  and Founder rows survive while email/session/token PII continues to cascade. Real-Postgres coverage
  asserts two retained identities, zero account links, zero active rows, and the one imported marker.
- Rate limiting now has a bounded LRU map and semantically free idle eviction after one full-refill
  interval. Explicit trusted-proxy depth controls X-Forwarded-For selection; zero ignores it and
  malformed/short chains fall back to the socket. Failed authenticated requests consume the same
  per-IP authority and become typed 429s after the burst.
- Batched adjacent LOWs: typed router 404/405, typed missing-account New Founder, normalized recovery
  input, and `Cache-Control: no-store` on the one-time credential response.

## 2026-07-30 — independent review: 2cdf1be (credentials, deletion, limiting)

**Approved.** Every ruling verified to the letter with live-Postgres proof: A-D1a stored-param
verify with floors + rehash-to-current inside the same row-locked transaction (concurrent logins
serialize on the account row); the timing oracle closed with a structurally valid dummy Argon2 on
the miss path (asserted never-authenticating); A-D5a via ON DELETE SET NULL + archive-stamp-before-
delete (the partial unique index needs no change — rows leave its predicate first), `imported`
survives, board projection works for deleted accounts; limiter gains bounded LRU + idle-TTL
(unbounded growth gone), trusted-proxy depth 0–8 with fail-to-socket on malformed chains, and
failed-auth requests now consume the IP bucket; the LOW batch (typed 404/405, no-store, input
normalization, typed NewFounder 404) all landed as claimed.

Residual findings: LOW — no upper bound on stored Argon2 params (a DB-write attacker could plant a
4 TiB-memory hash; add a ceiling alongside the floor); LOW — trusted-proxy docs must state the
origin-only-reachable-via-proxy precondition (XFF spoof bypasses bucketing when violated); LOW —
session/access-token GC remains open (correctly unclaimed). Observation: LRU capacity eviction can
be cycled to reset a bucket — equivalent to IP rotation, inherent to per-IP limiting, accepted.

## 2026-07-30 — round-3 LOW remediation

- Stored Argon2 parameters are rejected before allocation above 76 MiB, eight iterations, or four
  lanes while preserving the accepted upgrade window above current floors. Canonical deployment
  docs now require a positive trusted-proxy depth to be paired with an origin firewall.
- Session/access-token GC remains open: wiring a recurring job belongs with the not-yet-composed
  gameserver process. No orphan method is presented as a live cleanup path.

## 2026-08-21 — Q-001 accepted-scope witness closeout start

- The designated Claude review approved the platform-alignment range in `6d09379` and explicitly
  authorized Q-001/Q-002/Q-003 serially. Q-001 is first because Q-002 overlaps Account tests and
  Q-003 consumes Account/Transport composition.
- Froze Q-001 to AC2/AC3/AC5/AC6/AC7 witnesses only. No Account semantics, API schema, production
  behavior, copy, retention policy, client-session successor, archival state, or existing plan
  completion box may change without its separate authority.
- Predeclared the exact negative seams: live/revoked socket authorization, repeated Founder
  replacement, actual import-to-board refusal with server-created control, deletion survival of
  every stream family, and independent rate-limit enforcement/refill on all three unauthenticated
  operations.
- The batch will use root Make targets, cold `-count=1`, sequential Compose/Postgres packages, and
  restored temporary mutations. Its implementation range receives a Codex first-filter and then
  the mandatory exact-range Claude review; it will not archive the RFC.

## 2026-08-21 — Q-001 implementation and Codex first-filter

- Implementation commit: `e0f26d1`. The change is test-only: no Account/Transport/Leaderboard
  production source, API schema, authored copy, retention policy, or canonical behavior changed.
- AC2 now composes two real session families with the production Gameserver and Centrifuge socket.
  Refresh replay revokes only the target family; a rotated token is rejected before connect with
  close 4001, the already-connected socket is rejected by the real 25-second alive
  reauthentication path, its later subscribe write cannot survive, and the untouched family still
  connects and subscribes to the same Founder channel.
- AC3 performs first, second, and third New Founder replacements through HTTP. Every replacement
  is a distinct non-imported identity, every prior Company stream remains readable and archived,
  and all three new Company payloads are byte-identical catalog initials at revision 1. No cost,
  cooldown, or one-use transition occurs.
- AC5 constructs the supported local v14 save shape, calls the actual import endpoint, submits a
  real manual intent and Wind Down, obtains `ReplayVerified`, and waits for the production queue
  projector. The imported event has one projection claim and zero board rows; the existing
  server-created control in the same test has one `any_percent` row.
- AC6 enumerates all ten Founder/Company stream IDs across the original, three replacements, and
  active imported Founder before deletion. All ten remain readable and archived afterward; five
  Founder identities remain anonymized, including the imported marker, while account, email,
  session, access-token, family, and live bootstrap-secret counts are zero.
- AC7 independently exhausts create-account, recovery-login, refresh, and bootstrap buckets from
  distinct trusted-proxy IPs. Every rejected request returns the exact
  `{"category":"rate_limited","detail":"ip"}` body, leaves the seven relevant durable row
  counts unchanged, and succeeds after one measured refill interval. The rejected refresh uses a
  second valid family, so its later success proves the limiter did not consume it.

Cold green evidence:

- `make test-go GO_TEST_FLAGS='-count=1'` — complete Go tree green.
- `docker compose -f compose.save-test.yml run --rm test go test -p 1 ./account -run Integration -count=1`.
- `docker compose -f compose.save-test.yml run --rm test go test -p 1 ./gameserver -run Integration -count=1`.
- `docker compose -f compose.save-test.yml run --rm test go test -p 1 ./leaderboard -run Integration -count=1`.
- After restoring every mutation, the focused Account and Gameserver Q-001 populations passed
  again with `-count=1`; `git diff --check` and the focused Account/Gameserver/Leaderboard/Transport
  unit population were also green.

Demonstrated failing mutations, all restored before `e0f26d1`:

- skipped the final Founder in `DeleteAccount`'s stream archive loop: AC6 failed on the active
  imported Company stream with `ArchivedAt:<nil>`;
- rejected New Founder when any archived identity existed: AC3 failed on replacement 2 with HTTP
  500, proving the repeated-reset path is required;
- removed `limitUnauthenticated` from refresh: AC7 received HTTP 200 and a new token pair where the
  exact typed 429 was required;
- removed the imported branch from `QueueProjector`: AC5 observed `status=verified rows=4 claims=1`;
- bypassed `Node.OnAlive` database authentication: AC2 timed out with close status -1 instead of
  4001, proving the live-socket oracle depends on reauthentication rather than token expiry or the
  revoked-before-connect case.

Codex first-filter verdict: **APPROVED** for the test-only Q-001 scope.

- **Review by:** Codex.
- **Recorded by:** Codex.
- **Reviewed range:** `7ce6546..e0f26d1` (the test implementation commit after the predeclared
  plan boundary).

The diff was read in full;
the assertions bind real Postgres rows, real HTTP operations, a real WebSocket, replay verification,
and the actual board projector. No fixture silently skips in the declared Compose populations, and
no temporary production mutation remains. Designated cross-party review is still mandatory and
must cover the exact Q-001 range beginning after `f58a318`; this entry does not authorize archival
or Q-002.

## 2026-08-21 — Claude designated cross-party review of Q-001 `f58a318..8d8abf1` — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Range:** `f58a318..8d8abf1` = `{7ce6546, e0f26d1, 048f7ff, 8d8abf1}`, unioning the Codex
  first-filter's `7ce6546..e0f26d1` plus both record commits. The provenance correction in
  `8d8abf1` — labelling the Codex pass as a FIRST FILTER, never the designated review — is the
  cross-party rule observed correctly.
- **Scope verified:** every changed path is a `_test.go` file or planning record. No production
  source, schema, copy, or behavior byte moved.
- **Executed cold against real Postgres** (isolated port, orphans cleared): `./account` and
  `./gameserver` integration populations GREEN with `-count=1`. The AC2 socket witness runs for
  **23 s** because it rides the real 25-second alive-reauthentication window — verified `--- PASS`
  verbosely, not skipped.
- **Severing probe (mine, not the record's):** in a scratch worktree I made production
  `revokeFamily` (`server/account/store.go:722`) a no-op and re-ran the AC2 witness: **FAIL in
  3.6 s.** The witness discriminates on the exact production behavior it claims — under the
  platform-alignment taxonomy this is an integrated witness with a demonstrated severing failure,
  now recorded here as admissible evidence per Finding A of the audit verdict.
- **Plan checkboxes** (AC2/AC3/AC5/AC6/AC7 + the cold-run box) flip with their exercising tests in
  `e0f26d1` of the same range — compliant under the two-commit refinement.

**Verdict: APPROVED.** Q-001 closes. Q-002 (Minigame API witnesses) may begin, serially, per the
accepted queue. No archival, promotion, or push is authorized by this verdict.

## 2026-10-04 — AC6 joined public-board deletion supplement, first filter only

**Review by:** Codex (implementer first filter). **Recorded by:** Codex. A bounded test-only
supplement to AC6 extends the existing composed first-hour Postgres population, without changing
Account or Leaderboards production code. Its actual verifier/projector output is read through
the unauthenticated public board before and after a real refreshed-session account deletion. The
account is removed and the Founder/streams archived, yet the exact projected Founder/run remains
public. The negative board-variable partition excludes it; a temporary archived-Founder reader
exclusion made the post-delete assertion fail and was restored byte-exact. Cold focused Postgres
gameserver/leaderboard/account controls pass `-count=1`; non-Postgres Go and vet pass. The fuller
RP-118 dead-letter/poison/backup chain, D-009/D-015 policy and designated Claude review remain
open. Exact method, counts and limits: `planning/platform-alignment/data-rights-public-board-delete.md`.

## 2026-10-04 — AC6 × Deployment DP6 backup-restore boundary (RP-125)

**Review by:** Codex (implementer first filter). **Recorded by:** Codex. A retained
`deploymentbackup` Postgres integration test now brackets real repository deletion with encrypted
pre/post backups. The pre-deletion restore recreates the account and active Founder/streams;
the post-deletion restore preserves their deleted/archived state. The deletion-severing probe
fails at the source census; the full local ARM64 backup integration lane passes after byte-exact
restoration. This observes a cross-RFC rights risk, not an Account AC6 policy change or a public
workflow. `planning/platform-alignment/data-rights-restore-resurrection.md` states the limits;
D-009/D-015 and Claude's designated review remain open.

## 2026-10-06 — RP-048 / RP-234 session-consumer diagnostic predeclaration

Resumed at `b68190f9`, clean main eight commits ahead of the observed remote.
Account D2/D3 remain accepted; Q-001 is closed, not account/player completion.
Transport Q-003 explicitly excludes rotation authority. Source confirms that
browser runtime stores refreshToken without consuming it, and production
`Node.OnRefresh` deliberately rejects with auth_expired. Thus writing an
in-place WS refresh or silently retrying a single-use refresh would be invention.

This new range is test-only diagnosis / draft / records, distinct from Garden
`b58277cb..bff05b5e`. Predeclare the real built-client / actual server / disposable
Postgres test: healthy bootstrap and Garden purchase, then expiry of only the
exact issued access row (identified by validated JWT jti, not broad account SQL),
actual Garden HTTP 401 and closed subscribed socket within its existing alive
window, zero runtime refresh calls; an explicit test-operated POST session/refresh
must rotate exactly once and real reload must resume the same Founder/Garden.
Gameplay saved heads must remain unchanged across expiry/control/reload; original
refresh is unconsumed before the control, consumed afterward, and family unrevoked.
No raw credential/token/hidden Garden data may be printed. Bound socket observation
at 35 seconds, unchanged server security timing. Healthy/expired/rotated arms and
a severed test-expiry counterfactual must discriminate before the final claim.

This early database-expiry fault does NOT establish the precise original 67349
HTTP status, natural fifteen-minute JWT expiry, automatic renewal or completed
Garden maturation/harvest/reload. Its normal composed-driver objective stays intact.
No production auth/TTL/schema/kernel/copy/CI/epoch/checkbox/archive/push changes.
Missing browser concurrency, ambiguity, recovery and generated-operation authority
must be proposed in an explicit draft follow-up before implementation. Codex is
not its accepting owner or designated reviewer.

## 2026-10-06 — executed session boundary, controlled expiry and counterfactual

Authority predeclaration: `25fb5e67`; this is a new diagnosis/draft range, not
Garden's pending repair or a reopened Q-001/Q-003 approval. The retained manual
`make diagnose-browser-session` runs the existing built-client/Go/real-WebSocket
fixture through actual DOM bootstrap/Fiscal harvest/purchase. It then expires
only the validated token jti row in the declared disposable DB, not the JWT/server
clock. Runtime Garden GET returns exact 401 unauthorized/access_token; its actual
subscribed socket closes inside the existing alive window. No browser refresh
request occurs; the original refresh remains unconsumed and its family unrevoked.
One explicit test-operated POST session/refresh returns a distinct pair: two
session/access rows, exactly one consumed refresh and zero revoked families.
Actual reload restores the same Founder/Garden, with full gameplay heads unchanged,
no added gameplay intent, one original bootstrap and exactly one control refresh.

Terminal runs (all through the root Make target):

- 31533 fails before expiry: the Fiscal nav label includes its harvest badge, so
  the exact bare-title selector is no longer valid. Use the unchanged Desk control.
- 58141 expires the correct row but fails instrumentation: psql UPDATE RETURNING
  also emits `UPDATE 1`. A CTE SELECT count gives the exact affected-row oracle.
- 69863 completes the declared observation in 17.759 s, including real socket loss.
- 73000 keeps that exact diagnostic token valid (+15 minutes) instead of expiring
  it; the affected-row count still passes, but actual runtime HTTP 200 fails the
  required 401 assertion. This proves discrimination, not an assertion-only failure.
- Restored driver SHA256
  `e194c17d46737b4a5489c09fe1ffeb2c55153864edb8479e67baf3a6df6e3739`
  byte-exactly; 49226 completes the observation in 17.583 s. No production source
  was mutated. All driver-owned children/listeners terminate in finally.

The normal three-tick Garden objective remains intact in the other driver arm.
Its full maturity/harvest remains unproven. This controlled DB-expiry result does
not retroactively manufacture the omitted original 67349 HTTP status or measure
natural JWT expiry. It proves a missing default-client consumer and functioning
server refresh/reload control. CLI rejects unknown arguments; no target joins CI.

Draft `rfc/browser-session-renewal.md` separates the missing authority: exact
registered refresh DTO/errors, shared/cross-tab single-use coordination, ambiguous
completion, original request identity, normal history reconnect and honest
recovery dependencies. Web Locks source checked at its primary specification,
not inferred supported-browser proof. All choices/questions remain explicitly
unaccepted. No automatic auth behavior, server TTL/locks/security, copy, kernel,
public content, checkbox, review or archival status changed.

## 2026-10-06 — final diagnostic/CI evidence and first-filter scope

After terminal restoration/control runs, instrument-only reporting now retains
HTTP path/status pairs in failures, without request headers or credential bodies.
Final actual native diagnostic 85508 completes the same controlled population in
17.829 seconds. The original three-tick objective has no changed clock, bound or
acceptance assertion; it was NOT rerun or promoted in this range.
Final driver SHA256 is
`ca4958a07ae7a092839562b857da694498752590b1296b104f158761be5a6676`;
the difference from the restored counterfactual digest is failure-only status
reporting, not expiry/control/gameplay behavior.

Root client/type/boundary/topology 23925: 7366 pass / 134 browser-only skips,
zero TS/Svelte errors/warnings, 14/8/22 boundary counts and 13 rejected topology
negative controls. Complete declared Linux browser lane 28282: 300 populations,
22482 pass / six existing skips, 48.61 seconds, no uncaught worker/runtime errors;
the separate unchanged performance selector passes at 507 ms / 2.23 seconds,
with its 22 unrelated selector exclusions. Three real client builds execute 213
modules; final native build 374 ms. Node syntax and diff checks pass. These are
named local populations, not hosted/full verify-push or historical kernel green;
unchanged RP-131 and native Firefox/hosted reliability caveats remain open.

**Review by:** Codex (implementer first filter). **Recorded by:** Codex.
**Verdict:** APPROVED for the bounded diagnosis/draft scope only, not designated
review, owner adoption or production renewal.
Scope first-filter only: manual driver/Make alias, draft successor and truthful
canonical/planning records. No authentication/gameplay production source, token
TTL, family locking, migration, schemas, copy, epoch, kernel identity, CI
workflow/verification dependencies or completion boxes change. Initial invalid
instruments and valid-token counterfactual are disclosed; final objective is
diagnosis, not automatic renewal. Draft Founder-change guard prevents proposing
blind reissue of an old action into a different identity; API/policy/research
questions remain unaccepted. The complete new review span begins `b68190f9`
exclusive and includes `25fb5e67` plus this diagnostic/draft/record commit;
its literal endpoint is pinned after commit. Claude's designated cross-party
verdict is required; this is neither it nor acceptance/archival authority.

R-011's first wave is predeclared in the shared research queue. Next safe work
is actual supported-browser primitive observation and an exact refresh contract
census; pending policy does not authorize production renewal. RP-233's accepted
SG10 native interaction diagnosis remains independently available. All earlier
review ranges, owner rights/retention/accessibility/deployment obligations and
proper nine-tier 1.0 remain active. No push/publication/deployment/archive occurs.

Implementation/draft tip: `98ab59d8`. The new substantive span is
`b68190f9..98ab59d8`, two commits / sixteen paths, including its `25fb5e67`
predeclaration. This pin is a subsequent record edge: the requested designated
review must include this record commit too; the closing relay names that full
literal endpoint. No independent verdict is transcribed or implied, and no
earlier Garden/Account/Transport range is consumed. Final tree clean after the
record; all checks/diagnostics terminal, production controls unchanged.

## 2026-10-06 — R-011 primitive wave start, no renewal implementation

Revalidated `39f95329`, clean main eleven commits ahead of the observed remote;
no new Claude verdict or owner adoption appears. Previous turn is progress: real
controlled-expiry evidence plus the explicit draft. Full prior designated-review
request is `b68190f9..39f95329`, not just its substantive tip; it stays independent.

R-011's shared predeclaration is now refined before any measurement: use the
unchanged declared Playwright/Linux Compose runtime, three actual browser engines,
four arms × ten repetitions per engine (same-name exclusive lock; no-lock control;
different-name control; same-name isolated-browser-context control). Each arm
records actual central-observer entry/exit order and maximum concurrent holders.
Exclusive must peak at one; each deliberately independent/unlocked control must
peak at two while both holders remain blocked on explicit test release. Query
native pending/held locks to prove real contention, not absence inferred from a
short sleep. Ten additional owner-page-close repetitions per engine require a
queued successor, observed close start, then actual successor acquisition and
completion. This is page termination, NOT full browser/OS process crash.

Per-engine synthetic localStorage write/read must be shared between same-context
pages, absent in an isolated context, then visible after replacement-owner reload.
No production credentials, API, runtime coordinator, marker policy or player copy
exists in this primitive instrument. Record engine/version/UA, OS/arch, secure
context/API availability, exact 120 arm + 30 termination completions, all traces,
guard/exclusion/error fields and instrument SHA. A watchdog firing invalidates
the observation; its ceiling is not a proposed production timing budget.

Retain a root manual research target outside CI/verify and a generated observation
artifact. The script must fail if exclusive ownership is deliberately bypassed,
if any overlap control cannot overlap, or if a holder/successor/population never
completes. Restore its bytes before the final full research population. Retain
initial failures honestly; only completed populations authorize primitive selection
research, not session safety, unsupported/mobile/storage-denial behavior, natural
Garden maturity or implementation/owner adoption. Do not edit while handles live.
No accepted-body/production/auth/TTL/revocation/kernel/catalog/schema/CI/checkbox/
archival/push change; mandatory cross-party review remains pending.

## 2026-10-06 — R-011 primitive wave observed; local browser CI is red

Predeclaration `d023dc26`, separate new range after `39f95329`. Manual root
research target/script and generated artifact only; bounded dossier lives at
`planning/platform-alignment/browser-session-coordination.md`. Existing three
Linux engines complete 120 exclusion/control intervals and 30 actual page-close
successions, exact storage sharing/isolation/replacement visibility. This is no
real token rotation, process crash, native suspension/storage-denial, marker or
automatic-renewal proof. Source HEAD and actual dirty-instrument SHA are explicit.

Handles terminal in sequence: 48229 initial full population; 6764 exclusive
bypass red; original SHA `23b9fe248f42c46945a92401a3b5e65f72bc8521f14b54f0f0ac1bb997b499f8`
restored. Failure time/source/hash and poll-stop refinement; 35833 full population;
20140 bypass red (`exclusive lock admitted overlapping holders`); 1306 forced
one-ms guard red (`guard_exhausted:true`, zero completions). Two failed patch
context attempts changed no file. Permission-denied Docker invocation was a
terminal environment failure, not a measurement; declared narrow escalation used.
Final restored SHA `9dbad4855302ab7cefee97b79c51a0753b5fb4515987eb1094dbff363034a583`;
23130 all 150 cases pass. Independent Node read recomputes every holder peak,
repetition count and closure ordering and matches report hash. No edits while live.
Normal observation guard is not a production budget; failures retain explicit
invalidity, never count a partial population or a stale file as a new pass.

9136 client/type/boundary/topology: 7366 pass / 134 Node DOM skips, zero TS/Svelte
diagnostics, boundary 14/8/22, 13 rejected topology controls. Node syntax green.
Complete unchanged browser CI 54700 **FAILS** Firefox Route JSON import; 299/300
populations / 22466 tests / six existing skips, 51.98 s. Sixteen Route tests never
import; performance not reached. RP-235 filed immediately, precise failed request
status/cause not yet observed. No retry-to-green, browser removal or raised bound.
Historical RP-131, earlier worker/native-host and hosted caveats remain separate.

**Review by:** Codex (implementer first filter). **Recorded by:** Codex.
**Verdict:** bounded research scope inspected; evidence controls discriminate,
but complete local browser CI is red. NOT designated approval, owner adoption,
production renewal, completed R-011 or archive authority. New full span starts
`39f95329` exclusive, including `d023dc26`, this instrument/report/draft-evidence
and records; literal tip pinned after commit. Claude required independently of
`b68190f9..39f95329`, Garden and all earlier spans. No accepted body, production,
credential/TTL/revocation/kernel/catalog/schema/copy/CI workflow/checkbox/archive/
push change. Next separately predeclare RP-235 request/import diagnosis, then
real refresh contract/ambiguity; proper nine-tier 1.0 and full platform floor active.

R-011 substantive range pinned: `39f95329..ff2e9f19`, two commits / thirteen
paths including `d023dc26`. Designated review must include this following pin
record edge as well; its literal endpoint is supplied in the handoff. This is a
coordinate, not a verdict, R-011 completion, policy adoption or a green CI claim.

## 2026-10-06 — predeclare existing refresh parser/router census

Baseline `ee8b2356`, no live verification/probe handles. CI observation pin closes
separately; Docker disk is full and no new Postgres/browser run may be represented
as complete. Ask owner for precisely scoped unused project-cache cleanup or disk
expansion; no broad prune or user-data deletion. Continue work not requiring Docker.

Authority: accepted Account D2/D3/AC7 and API Foundation C1/A4/A5 observation of
existing response bytes. New batch is **test-only parser/router census**, not
registry/handler semantics, TTL/revocation, coordinator or draft acceptance.
Use actual NewAPI/chi routing and refresh handler, an intentionally DB-less
Repository, controlled clock and constructor-only unrelated intent stub that
fails if called. Every supplied token is malformed and must return before DB.
This explicitly does NOT exercise valid/unknown opaque tokens, successful refresh,
single-use/replay/expiry/transaction faults, real HTTP socket or real Postgres.
Those are a separately retained second census stage, not silently skipped here.

Population: empty/whitespace/malformed JSON, unknown/trailing members or values,
array/boolean/number/string roots, wrong-type refresh fields, over-limit body;
missing/null/empty/malformed/short token, duplicate and case-insensitive field
spelling; method/path refusals; actual unauthenticated limiter before parser,
one fixed-clock exhausted refusal and subsequent clock-refill control. Pin exact
current status/category/detail and Content-Type with full byte equality, never
prefix/substring; do not canonize permissive parsing as approved future policy.
No credential/body secrets in failure output (all request bodies synthetic).

Predeclare three independent temporary severings: refresh invalid-body error byte,
refresh unauthorized-token error byte, and refresh's limiter mounting. Each must
fail the new cold root Account population, then restore API SHA byte-exact before
the next probe/final pass. No edits while handles live. Root make test-go with
GO_PACKAGES=./account and explicit -count=1, followed by root vet selector and
diff/scope checks. Full Docker/CI remains invalid under RP-236 and RP-131;
these focused passes cannot waive it. Retain errors/probes and literal range for
Claude independently; no plan checkbox, docs production behavior promotion,
archival, push or owner policy adoption. Next real-Postgres stage waits for a
safe capacity resolution, not a replacement mock success.

## 2026-10-06 — parser census green; real rotation still unmeasured

Predeclared `d3e7196f`: test-only actual NewAPI/chi/refresh decoder/token-shape/
limiter, 22 parser subcases plus two auxiliary tests, 27 exact requests. Fourteen
body arms 400; eight null/missing/duplicate/uppercase/malformed-token arms 401;
limiter precedes parsing/refills; method/path 405/404. Exact Content-Type/status/
full JSON bytes, no unexpected credential-bearing response output. Intentionally
absent DB/unrelated fail-if-called fixture disclosed; no valid/reused-token proof.

38743 cold selected population green. Invalid-body byte probe 33000 fails
fourteen arms and first limiter response; unauthorized-token byte probe fails
eight arms; limiter removal fails status=400 want=429. All terminal before edits.
Each restores API SHA 36dc2876a0404b68de9e92e41cb950ea6d67e42e97014c201abacdccfa80cda5;
store unchanged cff18305f908e5f59975fd8a14466b5c9b776536d53b26d2eccd071097c5a308.
Final 19985 cold Account/publicapi passes, selected root vet/diff passes. DB
conditionals not counted as integration. Dossier at
platform-alignment/session-refresh-contract-census.md retains exact limits.

Read-only Docker inventory identifies unused Cloud Clicker labelled caches,
zero references: go-cache 2.523 GB, browser-node-modules 123 MB, pnpm-store zero.
Owner cleanup/capacity answer pending; nothing deleted. Initial verbose df preview
truncated before volumes; complete larger retrieval supplies sizes. A failed
multi-file record patch changed nothing (wrong append context); corrected below.

Review by: Codex (implementer first-filter). Recorded by: Codex. New span begins
ee8b2356 exclusive, including d3e7196f; implementation/records/pin all belong in
designated review. CI observation span 6ed42d45..d3e7196f stays independent.
No production source, accepted schema/body/policy, plan checkbox/archive/push
change; proper 1.0 unchanged. Continue real-DB stage after safe capacity resolution.

Exact substantive census span `ee8b2356..9c638cd0`, two commits / ten paths:
shared d3e7196f CI pin/Account predeclaration edge plus 9c638cd0 test/records.
The CI-only observation handoff is `6ed42d45..d3e7196f`, five commits / twelve
paths, including that disclosed shared record edge. This census pin belongs to
the Account range too; final relay supplies its literal endpoint. No reviewer
verdict consumed either span. Review by: Codex (first-filter only). Recorded by:
Codex. Source tree clean after substantive commit, twenty ahead of observed
origin/main; no fetch/push/publication, deletion or acceptance performed.

## 2026-10-06 — predeclare real refresh response/persistence census

Previous goal turn made progress (CI evidence/instrument and executed parser
census); no claim it completed the full session question. Baseline c2843089,
clean main twenty-one ahead of observed remote, no live handles/new Claude
verdict. Docker capacity choice still pending; no deletion/prune or new Docker
measurement authorized by this record. This turn may prepare/compile real-DB
witnesses but must label them UNEXECUTED until the declared DB actually runs.

Authority: accepted Account D1/D2/D3/AC2/AC7, API Foundation C1/A4/A5 existing
handler/status census. Test-only scope: no production/schema/router binding,
generated client, auth TTL/revocation/locking, owner copy/retention, browser
policy, CI topology, acceptance checkbox/archive/push change.

Populations, serial in declared Postgres 16 through root test-save-integration:
(1) live-access and exactly access-expired refresh rotations; exact two string
success fields, valid token shapes, persisted old consumption/new descendant,
account/Founder identity and TTLs; exact consumed-token reuse error and subsequent
revoked-family refusal, all descendants revoked. (2) real NewFounder setup then
refresh binds the new active Founder without revoking the family. (3) unknown
canonical opaque token with no writes, exact 30-day expired refresh with family
revocation/no new credential, and actual closed database handle mapping through
current generic handler error (observed current 401, not endorsed outage policy).
(4) successful rotation, exact rate-limited valid descendant refusal with unchanged
full credential/gameplay rows, refill allowing that same still-unconsumed token.

Use existing bootstrapRepository fixture, real NewAPI/chi/Repository and net/http
over testhttp net.Pipe (not OS socket/native browser). Constructor-only unrelated
intent fixture fails if called; no new game engine stub claimed as composition.
Read exact status/content-type/full error bytes; success exactly access_token and
refresh_token, both strings, no duplicate/extra/null/trailing fields. Keep tokens,
unexpected bodies, recovery hashes and full SQL snapshots out of failure output.
Snapshot the named account/Founder/save/history/outbox rows, and separately all
session-family/refresh/access rows; prove real token consumption/revocation/expiry
by SQL and authentication, not response-only assertions. No full-table or release
coverage claim beyond the actual named population.

After capacity is safely resolved, execute two cold complete new-test populations
then existing Account/publicapi population through the declared service. Four
independent temporary severings scheduled: consumed_at update no-op, family
revocation no-op, reuse error literal change, expired-refresh guard bypass. Each
must fail its new scoped witness, restore actual API/store SHA before next/final
execution. Until then they are NOT demonstrated failures and no gate may promote.
Record exact flags/denominators/source hash/limitations and pending probes in a
dossier; local compile/test discovery/vet are not DB acceptance. Any unexpected
real result is a finding, never an edited expectation to turn red green.
Only new test/record checkpoint may commit while DB evidence remains unavailable;
plan boxes, contract/RFC states and all prior independent reviews stay unchanged.

## 2026-10-06 — real refresh witnesses prepared, explicitly unexecuted

Predeclared d33b5d56. Four top-level tests/seven real-DB populations prepared:
real Repository/chi/net.Pipe HTTP, existing epoch-5 Postgres fixture, exact
success/error bytes and literal fifteen-minute/thirty-day boundaries. Old-token
consumption, descendant revocation, Founder binding and full named ten-table
gameplay/three-table credential oracles. Closed fault uses actually closed
secondary sql.DB; primary fixture/cleanup stays open. No private snapshot/token
output. API/store SHA unchanged. Test SHA after refinement:
fefd7326b0792ecdd9e30d14e615f9e2b9dc028fd2b5cbdeb5318f7a3468ebf3.

32238 host compile: all seven DB populations SKIP without TEST_DATABASE_URL;
two grouped parents misleadingly display PASS with every child SKIP. Parent-level
dependency skips added; 86768 explicitly shows four top-level SKIPs. Existing
22 parser cases/two controls execute green, selected vet/diff pass. No DB gate
or new severing executed. SQL alias changed to unambiguous r on static inspection,
not after an actual DB red. Records patch twice missed exact append context;
failed atomically, no partial file changes; actual lines re-read before repair.

Read-only existing declared Postgres capacity: overlay 100%, 39,784 KiB available,
DB tmpfs >8 GB free. Two healthy existing project DB services are not verification
handles, not stopped/restarted. Owner cleanup answer pending; no new Docker
workload or deletion. One absent guessed save/postgres.go lookup corrected with
rg --files/read of database.go. All verification handles terminal before edits.

Review by: Codex (preparation first-filter only). Recorded by: Codex. New span
starts c2843089 exclusive, includes d33b5d56/test/records/pin. Claude required
independently; no real-DB evidence, completed API contract, browser policy,
checkbox/acceptance/archive/push claim. Next separately predeclare a test-output
observer that rejects dependency/empty transcripts before calling them proof.

## 2026-10-06 — pin prepared span; predeclare completeness observer separately

Prepared substantive span c2843089..3121a376 (two commits/eight paths), not
executed/approved; following record edge belongs in its designated range too.
New observation scope begins after 3121a376 and includes this shared pin/
predeclaration. Authority Account AC2/AC7 evidence instrumentation only, not
production renewal, schema generation or CI workflow authority.

Add a manual root test-refresh-census observer outside CI/verify. It invokes the
existing root test-save-integration leaf with fixed ./account selector, anchored
four-test regex, -json/-count=1. Parse streaming Go events, require account package
start/pass, exact run→pass for all four parents/seven leaves, zero skip/fail,
no duplicated/missing/unexpected execution. Child exit zero is necessary but
insufficient. Retain complete count, skip/missing/error/capture fields, actual
source/test/instrument hashes and before/after source identity. No raw token/
private-row/test Output values in the report. Do not infer DB success from a
package pass, fixture transcript, source read or host compilation.

Predeclare fixture controls: honest full event stream is accepted only as a
synthetic validator control; empty stream, package-only pass, parent-only pass,
one missing leaf, skip disguised by parent/package pass, fail, duplicate/unknown
case, malformed JSON/event, wrong package/order, nonzero child exit all reject.
For an actual negative, dedicated clearly labelled host missing-DB mode invokes
root test-go on the same anchored names without TEST_DATABASE_URL; actual command
may exit zero, observer MUST return invalid/incomplete/nonzero. This never claims
to exercise rotation. Its report mode/command cannot masquerade as real-DB mode.
Normal mode launches no workload until Docker capacity safely resolved. Script
must not retry, rewrite expectations, skip a required case or change bounds.

Root Node fixture/syntax target and actual host negative may execute now; only
they and compilation count as current evidence. Live handles drain to terminal
before any source/record edit. New instrument/Make/docs/records in its own commit;
no CI membership/dependency/kernel/copy/production/checkbox/archival/push change.
Real normal-mode populations/four prepared DB source probes stay explicitly
pending. Claude required independently for both spans; no self-designated verdict.

## 2026-10-06 — completeness observer rejects actual skipped green command

Manual test-refresh-census/observer implemented under 08a11814 predeclaration,
outside CI/verify. Four parents/seven leaves, exact package/order/terminal census,
no private Output retention; listed input hashes and HEAD checked before/after.
Twenty synthetic controls pass; they are validator fixtures, not DB success.
Root verify-ci-topology and its thirteen negative controls remain green.

Two initial host-control invocations fail before Go executes: Make consumes the
regex dollar/closing quote; the attempted JS replacement also preserves a single
dollar. Both are instrument errors, not demonstrated dependency controls. Callback
replacement preserves doubled dollar through Make. Final actual 79284 completes:
child Go/Make exit 0, package pass, four dependency SKIPs, zero/seven leaves,
24 JSON events/one Make line; observer exit 1, invalid/incomplete, source stable,
no truncation/capture error, 2061 ms. Retained refresh-population-negative.v1.json
pins mode, command, source and limits. RP-238 records the evidence hazard. No
test body/token/private snapshot emitted into the retained report.

Normal declared-Postgres mode remains unexecuted pending RP-236 capacity. This
does not prove real rotation or any of four prepared source severings. No Docker
workload/deletion, policy/CI/dependency/kernel/copy change or checkbox promotion.
Source stability means the fourteen listed inputs plus HEAD, not environment or
all ignored files. All execution handles terminal before records/diff inspection.

Review by: Codex (instrument first-filter only). Recorded by: Codex. Observer
range starts 3121a376 exclusive, includes shared 08a11814 predeclaration, script,
fixtures, Make target, artifact/docs/ledger/records and following pin. Claude
required independently; prepared range c2843089..08a11814 remains preparation,
not executed acceptance. No self-designated verdict, archive, push or release.

## 2026-10-06 — broader client gate fires; predeclare fixture-name repair

Cold root Account/publicapi tests and selected vet pass (2693), DB populations
not executed. Root verify-client (4122) passes typecheck (zero errors/warnings)
and build, then fails: Vitest discovers tools/observe-refresh-population.test.mjs,
which imports node:test and registers no Vitest suite. 84 client files/7366 tests
pass, 17 files/134 existing tests skip, one extra file fails; remainder of root
verify-client does not execute. Standalone twenty Node controls passing does not
make this composite green. New tooling caused this failure, not a product bug.

Repair within observer scope: rename standalone Node fixture to established
*.fixtures.mjs convention and update its explicit Make invocation. Do not change
Vitest config/include/exclude, CI membership, old tests or assertions. Retain red
and RP-239; rerun standalone controls, complete verify-client and actual missing-DB
control on changed Make source before recording any green. Previous negative
report remains historical evidence with its actual old Make hash; retain a second
actual report for repaired source rather than pretending hashes are unchanged.
All handles terminal before this predeclaration; no Docker/new DB/production
policy, checkbox, acceptance, archive or push change. Claude remains designated.

## 2026-10-06 — fixture-name repair executes; composite stays honestly red

Under dbcce391 predeclaration, moved Node controls to *.fixtures.mjs, updated
explicit Make invocation. No changed Vitest/workflow config, old test assertions
or exclusions. 68948: twenty standalone controls pass; root typecheck zero
errors/warnings, build, unchanged client population 84 files/7366 tests passes,
17 files/134 existing skips; shell/UI boundaries and kernel checkout/adversarial
fixtures pass. verify-client then fails unchanged RP-131 at pushed 50a3a514:
historical guarded paths lack the same-commit signal. Composite exit 2, NOT green.
No cosmetic bump/waiver/rewrite or draft kernel-history RFC implementation.

Remaining root gates execute separately (40302), exit 0: topology/thirteen
controls, combat/meters/achievements, cosmetic boundaries/22 negatives,
no-payment/six negatives plus two near-misses, copy/content manifest. Copy keeps
610 orphan warnings, not silently clean. Earlier cold Account/publicapi/vet
2693 pass without DB populations. No hosted Actions or browser run claimed.

Actual repaired-source host control 62186: child 0/package pass/four skips,
zero/seven cases, observer exit 1 invalid/incomplete; 24 events/one Make line,
2374 ms, unchanged fourteen listed inputs+HEAD, no capture error/truncation.
Second retained refresh-population-negative-repaired.v1.json pins changed Make
SHA e3e02c01b35ae8a197dfe02aaa6eaa1feebc97ac715498bcad48e145046786c3.
Original negative artifact remains historical, not overwritten/relabelled.
No live verification handles before edits. Current read-only Docker overlay
still 100%, 39,784 KiB available; three candidate cache volumes remain unused,
no deletion/new workload. Normal Postgres/two populations/four severings pending.

Review by: Codex (first-filter only). Recorded by: Codex. RP-239 locally repaired,
designated Claude review pending. Observer range starts after 3121a376 and must
include 08a11814, 035e5549, dbcce391, repair/artifact/records and following pin.
No acceptance/checkbox/archive/push or proper full nine-tier 1.0 scope reduction.

## 2026-10-06 — pin observer including its caught and repaired defect

Substantive observer range 3121a376..474e9e9d, four commits/thirteen paths,
includes predeclaration, initial tooling, fired composite red and naming repair.
This following pin-record edge is part of the complete designated range too;
closing relay names the final literal tip. No production/auth/workflow/test
exclusion bytes changed. Actual skip negative runs and historical artifacts are
distinct; normal Postgres/two populations/four severings remain unexecuted.
Standalone fixtures/current client population and remaining separately executed
gates pass; verify-client composite is still RP-131 RED. All handles terminal,
no cleanup/deletion or push. Review by: Codex (first-filter only). Recorded by:
Codex. Claude mandatory independently, no archival eligibility inferred.

## 2026-10-07 — execute refresh cases on the existing composed Postgres lane

Account D2/D3; routine test integration under the current RFC-0000 procedure. Added the four
existing refresh test parents to `test-game-ui-composed.mjs`, using the existing seven-case
observer, actual API/repository and declared ephemeral Postgres. Private Go output is consumed
in memory, not printed or retained. Reset the fixture between Account, historical-content and
browser populations; no runtime/token/policy/CI-topology change or new observer/artifact system.

`make test-game-ui-composed` (82596) PASS: 7/7 refresh cases, persisted Fiscal/historical Exit
checks, real DOM/server/WebSocket transitions and both endings/continuation/recovery, Fiscal
refusals/fresh consent, achievement acquisition, opportunity effect, Pitch, and cosmetic
Buy/adopt/care/equip/reload. Refresh cases include correct stored family/expiry/Founder binding,
reuse revocation, rejected-request non-mutation and limiter refill. HTTP uses net.Pipe; the
closed-DB case observes current 401 mapping, not adopted outage policy or browser renewal.

`make test-refresh-observer` PASS (20 retained fixture controls); actual
`node client/tools/observe-refresh-population.mjs --missing-db-control` (59722) correctly exits 1:
Go child/package pass, four skips, 0/7 completed cases, stable listed sources, no truncation.
Cold `make test-go GO_PACKAGES='./account ./publicapi' GO_TEST_FLAGS=-count=1` (46945) and
selected `make vet` PASS; these latter ordinary runs skip DB cases and are not DB evidence.
Node syntax/diff checks pass. All handles terminal before records; declared DB other sessions
zero and ports 18081/18082/5173 absent. No cleanup, push, archive or full CI/release claim.

Review by: Codex (first-filter only). Recorded by: Codex. New binding/docs/records range starts
after `0f9f4214`; designated Claude review remains, alongside earlier prepared-test/observer
ranges. The native route removes the execution gap for these seven cases, not RP-236's full
Linux image/storage blocker. Next: generated refresh contract and browser policy need accepted
authority before implementation; wider Account rights/retention/acceptance remain open.

## 2026-10-07 — committed refresh with a lost reply

Account D2 and renewal's unresolved ambiguity input. Added one real API/Postgres failure-path
test: healthy rotation first, then run the unchanged handler while discarding response writes
and closing the real net.Pipe HTTP connection after the handler returns. No response credentials
are captured. The client receives EOF, yet storage contains the consumed token and unseen new
pair. Retrying the consumed credential returns exact refresh_reused/session_family_revoked,
revokes all three refresh/access generations, and rejects the previously received access token.
Further retry and all gameplay/account rows remain unchanged. This proves committed reply loss,
not an actual browser/tab crash, automatic renewal or permission to change replay detection.

The existing composed selector/observer now requires eight cases/five parents. Whole
`make test-game-ui-composed` (46895) PASS, including main and cosmetic player journeys.
Observer fixtures 20 PASS; actual missing-DB control (43644) rejects child/package 0 with five
skips/0-of-8 cases, exit 1, stable listed sources/no capture error. Cold Account/publicapi
tests (97139) and selected vet PASS; ordinary dependency skips are not Postgres proof.
All handles terminal before records; fixture DB other sessions zero, target listeners absent.

Review by: Codex (first-filter only). Recorded by: Codex. New tests/observer/docs/records range
starts after `96e5ba82`; designated review remains. Earlier ranges are not approved by this
entry. No production auth/TTL/parser, generated contract, kernel, copy, CI topology, archive or
push changes. Asked Marco whether to accept/delegate the draft browser-renewal proposal; no
answer or acceptance inferred. API S-A1 and recovery S-A2 still need resolution before their
consumer implementation. No full CI/Account/1.0 claim.

## 2026-10-08 — missing recovery player-flow contract drafted

At `1f95a62a`, rechecked Account D1/D2, archived Game UI's idempotent bootstrap/security contract,
accepted Garage manifest F01, D-005 and the actual runtime. Session creation is already generated
and mounted; the browser still silently stores the code and has no recovery UI. F01 explicitly
requires a successor contract, so no new UI/storage policy is inferred from the posture alone.

Drafted `rfc/account-recovery-player-flow.md`: one resumable save-credential setup, native
copy/download/manual acknowledgement, legacy migration without discarding the only code,
generated same-account login/read/persist/bind sequence and bounded failure/lifecycle handling.
Proposed exact sincere utility copy is explicitly unadopted. Existing renewal/first-read policy,
export/deletion/retention and release floors remain separate; no server auth or TTL amendment.
The active index, D-005 route and RP-123 now link this draft instead of losing it in chat.

Verification: inspected referenced server/SDK operations, current credential/startup source and
the actual spec/record diff. No software suite for a draft-only change and no passing acceptance
claim. Review by: Codex (drafter first filter); Recorded by: Codex. New normative proposal after
`1f95a62a` needs owner contract/copy adoption and designated cross-party specification review.
Next: obtain that bounded adoption, then implement the existing real player task with its tests;
do not reopen D-005 or treat unrelated rights choices as a blocker for this scoped successor.
