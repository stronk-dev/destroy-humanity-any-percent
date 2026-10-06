# Existing session-refresh contract census — 2026-10-06

Authority: accepted Account D2/D3/AC7, API Foundation C1/A4/A5;
predeclared `d3e7196f`. Test-only observation, not renewal acceptance.

## Stage one: executed parser/router population

`server/account/refresh_wire_census_test.go` uses actual NewAPI/chi routing,
refresh handler, production decoder/token-shape guard and limiter. `httptest`
uses an in-memory recorder, **not a socket**. Repository deliberately has no DB:
all supplied tokens are malformed and return before SQL. Constructor-only
unrelated intent fixture fails if called. No successful rotation proof.

Twenty-two parser cases plus limiter/method/path controls issue 27 requests:

| Population | Exact current outcome |
| --- | --- |
| Empty, whitespace, unfinished object, unknown member, second JSON value, trailing non-JSON, array/boolean/number/string roots, number/object/array token, configured 1024-byte limit exceeded | 400 `{"category":"invalid","detail":"body"}` |
| Missing/null/empty/malformed/short token, null root, duplicate malformed fields, uppercase REFRESH_TOKEN | 401 `{"category":"unauthorized","detail":"refresh_token"}` |
| Second malformed request after exhausting fixed-clock one-token IP bucket | 429 `{"category":"rate_limited","detail":"ip"}`; limiter precedes parser |
| Same malformed request after one-minute fixture refill | 400 exact invalid/body |
| GET refresh path | 405 `{"category":"invalid","detail":"method"}` |
| POST absent path | 404 `{"category":"unknown_id","detail":"route"}` |

Every oracle compares complete JSON bytes **including encoder newline**, exact
Content-Type application/json and status. Never echo unexpected response bodies;
a broken handler could expose tokens. Request strings are synthetic only.
Null/duplicate/case-insensitive fields are observed current parsing, not adopted
future policy. A new stricter request descriptor cannot silently change this
acceptance/error set or create a second API authority.

## Discrimination and exact restoration

- Cold initial 38743: all three top-level tests / 22 parser subcases pass.
- Invalid-body detail → census_probe: 33000 fails fourteen body arms and first
  limiter response, exact error-byte mismatch.
- Unauthorized-token detail → census_probe: cold command fails eight 401 arms;
  other cases green.
- Remove refresh limiter mount: cold command fails `status=400 want=429`;
  parser/method/path green.
- Before each next probe/final run, API restored exact SHA-256
  `36dc2876a0404b68de9e92e41cb950ea6d67e42e97014c201abacdccfa80cda5`;
  store unchanged `cff18305f908e5f59975fd8a14466b5c9b776536d53b26d2eccd071097c5a308`.
- Final 19985 root Account/publicapi tests use explicit -count=1 and pass;
  selected root vet/diff pass. Existing DB-conditional tests do not execute
  without TEST_DATABASE_URL: **not an integration gate**.

No edit while handles live; no production mutation survives. No plan checkbox,
schema/TTL/family/copy/kernel/CI policy change, archive or push.

## Still unexecuted

Predeclare real declared-Postgres canonical valid/unknown/expired/reused-token
populations, successful pair persistence/Founder identity, family-wide revocation,
limiter no-mutation, repository faults and committed-reply loss separately.
Existing tests/source are inputs, not new executed evidence. Generated consumer,
owner recovery policy and full browser renewal remain unaccepted boundaries.

RP-236 prevents further Docker populations pending safe capacity resolution.
Read-only inventory finds unused, explicitly project-labelled development caches:
cloud-clicker_go-cache 2.523 GB, cloud-clicker_browser-node-modules 123 MB,
cloud-clicker_browser-pnpm-store 0 B; zero container references. No deletion;
owner answer pending. Other projects' images/unattributed volumes excluded.

Review by: Codex (first-filter only). Recorded by: Codex. New range begins
`ee8b2356` exclusive and includes `d3e7196f`, test, dossier, reconciliation and
following pin. Claude required independently of previous spans. R-011, natural
fifteen-minute Garden and proper nine-tier 1.0 remain open; renewal RFC draft.

## Real-Postgres preparation — NOT executed evidence

Predeclared d33b5d56; refresh_wire_integration_test.go declares seven populations
in four tests: live/access-expired rotation/reuse, New Founder binding, unknown
canonical token, exact refresh expiry, closed-DB mapping, limiter non-mutation/
refill. Existing epoch-5 repository/DB fixture, HTTP over net.Pipe (no browser/
OS socket). Exact bytes/literal 15-minute/30-day bounds and persisted outcomes
are asserted, **not yet verified against actual Postgres**.

Private snapshots cover accounts, account_emails, account_founders, save_streams,
save_revisions, events, intent_records, run_genesis, founder_genesis,
transport_player_outbox, and separately session_families/sessions/access_tokens.
Host 32238/86768 compiles; all DB cases skip. Grouped parent PASS with all child
SKIPs corrected to explicit parent SKIP, not a failed/passing rotation. Existing
parser/vet/diff green; no new failing severing demonstrated. Source SHA
fefd7326b0792ecdd9e30d14e615f9e2b9dc028fd2b5cbdeb5318f7a3468ebf3.
Docker still 100%, 39,784 KiB free; no fresh workload/deletion. Two complete DB
populations and four predeclared severings remain pending capacity/review;
prepared tests are not integrated witnesses or readiness promotion.

## Completeness observer — actual negative, not rotation evidence

Under separate 08a11814 predeclaration, `make test-refresh-census` wraps the existing
declared Postgres root leaf, fixed four-test selector, -json/-count=1. Manual only,
outside CI/verify. Requires package start/pass and all four parents/seven leaves
to run/pass; rejects missing/skipped/failed/duplicate/order/source/capture failures.
`make test-refresh-observer` passes twenty synthetic controls, not DB populations.
Existing topology/thirteen negative controls pass without CI membership changes.

Actual host missing-DB control 79284: child exit 0/package pass, four top-level
SKIPs, zero/seven leaves; observer rejects with exit 1, invalid/incomplete. Capture
24 events/one Make line, source stable, no truncation/error, 2061 ms. Two earlier
quote-expansion instrument errors execute no Go tests and are not this control.
[Retained actual report](refresh-population-negative.v1.json) excludes private
test Output and states mode explicitly. Source check covers fourteen listed
inputs plus HEAD, not arbitrary ignored files/environment. RP-238 records why
exit-zero alone cannot certify integration.

Normal mode is unexecuted, not a successful Postgres gate. Two complete real
populations/four severings await capacity; draft renewal/whole 1.0 remain open.
Review by: Codex (first-filter). Recorded by: Codex. Separate range begins
3121a376 exclusive, includes 08a11814 and all instrument/record/pin edges;
designated Claude review pending, no archival/status/push authorization.

### Composite collision caught and repaired, not whole-CI green

Root verify-client 4122 fails on new Node-only *.test.mjs fixture, zero Vitest
tests. Repair predeclared dbcce391 renames only to *.fixtures.mjs, updates explicit
Make invocation; no config/test exclusions. 68948 passes twenty controls,
type/build, 7366 client tests/134 existing skips and shell/UI boundaries, then
fails unchanged pushed RP-131 history guard. Remaining gates separately pass
40302; copy retains 610 orphan warnings. Composite remains RED.
Repaired actual missing-DB control 62186 again rejects child exit 0/four skips/
zero-seven cases, observer exit 1, 24 events, 2374 ms, source stable and capture
complete. [Second report](refresh-population-negative-repaired.v1.json) pins
the changed Make source; original report remains historical. RP-239 is locally
corrected, Claude review pending. No new Docker/rotation/release proof.
