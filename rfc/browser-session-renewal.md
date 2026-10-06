# RFC: Browser Session Renewal & Recovery

- **Status:** draft — NOT implementation authority
- **Author:** Codex (proposal; no owner adoption claimed)
- **Created:** 2026-10-06
- **Design refs:** `design/00` pillars 5–7 (wall-clock idle play, honest persistence, server authority); `design/06` backend/frontend, server clock and authenticated intents
- **Depends on:** Account & Session Bootstrap D2/D3; WebSocket Transport T1–T5; API Foundation A4/A5/C1/C7/C9
- **Parent / amends:** archived Game UI Screens runtime binding; Account/Transport consumer integration (no server security amendment)
- **Planning:** `planning/account-and-session-bootstrap/` contains the bounded pre-implementation diagnosis; implementation plan only after acceptance

## Summary

Bind the browser's stored, single-use refresh credential to every authenticated
HTTP consumer and the existing WebSocket recovery controller. A player can remain
in the same Founder/run beyond one fifteen-minute access lifetime without a reload,
replacement account or repeated gameplay action. Existing family serialization,
revocation and expiry stay server-authoritative. All specification below is a
proposal until the open boundaries are resolved and Marco accepts it.

## Measured motivation and scope

RP-048 already recorded this gap; RP-234 is its current Garden workflow evidence,
not a new independent feature. `make diagnose-browser-session` exercises the built
client, real gameserver, Postgres and WebSocket. Expiring its exact access row
produces Garden HTTP 401 and socket closure, with no client refresh. One explicitly
test-operated refresh rotates the pair; reload resumes the same Founder/Garden,
with unchanged gameplay heads. Keeping the test token valid makes the 401 oracle
fail. This is early database expiry, not natural JWT expiry or completed Garden
maturation. The earlier native fifteen-minute objective remains red and its exact
HTTP attribution is not retrospectively inferred.

Q-003 deliberately excluded Account rotation. Its implementation closes visibly
on 4001/4002, and the server deliberately rejects in-place Centrifuge refresh.
Neither an archived UI contract nor a desirable fix supplies new policy authority.

In scope: one shared browser renewal coordinator, registered refresh operation,
exact-identity HTTP continuation, credential persistence and ordinary transport
recovery. Out of scope: TTL changes, new server token claims, rotation/revocation
relaxation, replacement arbitration, cookie/identity-provider migration, new
gameplay/API routes, retention/export/deletion rulings, owner-authored copy,
public Garden activation, balance clocks, general SDK redesign and CI deadlines.

## Proposed specification

### S1 — One operation authority, unchanged server semantics

Register the existing `POST /api/v1/session/refresh`, not a second endpoint:
request exactly `{refresh_token:string}`, success HTTP 200 exactly
`{access_token:string,refresh_token:string}`. Its authentication is the presented
opaque refresh credential, not a required valid access token. Consume generated
descriptors/types through API Foundation's single operation/transport authority.
Pin the handler's complete existing error/status alternatives, including
`refresh_reused/session_family_revoked`, before acceptance; do not incorrectly
reuse the narrower Minigame error union. Existing compatibility pins remain intact.

No change to Account's fifteen-minute access / thirty-day refresh lifetimes,
family-row lock, single-use rotation or revocation-on-reuse. Every server response
is still validated against its real operation contract, not a client byte table.

### S2 — Demand-driven renewal, not a second clock or per-surface client

One coordinator is shared by snapshot, intents, Garden, Minigame and Soul ports,
and transport recovery. Renew on an exact authenticated HTTP
`401 {category:"unauthorized",detail:"access_token"}` or transport auth-expired
4001. Do not refresh on 4002 replacement, permission denial, application rejection,
ordinary network failure, rate limit, drain, unknown status/body or arbitrary error.
The existing terminal/offline state stays visible when renewal cannot succeed.

This proposal does not introduce a client-side TTL timer, alter game clocks or
send periodic gameplay writes. Success must work with suspended/background timers:
the first authenticated demand after resuming can trigger the same coordinator.

### S3 — Single-use rotation includes other tabs and ambiguous completion

Serialize renewal across all cooperating contexts sharing the credential document,
not merely a promise inside one runtime. After acquiring that ownership, reread
storage: if another context already replaced the failed access generation, use
that pair rather than replay its refresh. Do not overwrite another account's or
newer generation's credential document when an old response completes.

Persist a non-secret in-flight generation marker before dispatch. If dispatch may
have consumed the refresh but no validated replacement pair was durably stored
(network loss, malformed reply, storage failure, tab crash), do NOT retry that
refresh token, even after reload. Stop visibly and preserve recovery material;
the existing owner-ruled recover-existing-account workflow owns recovery, never
implicit bootstrap. A later newly authenticated generation clears that marker.
No token is duplicated into logs, events, URLs, lock names or marker fields.

Candidate coordination primitive: an exclusive Web Lock plus the durable marker.
The [Web Locks specification](https://w3c.github.io/web-locks/) describes exclusion
across cooperating contexts sharing a storage bucket; this does not prove our
supported-browser behavior or solve ambiguous server commit by itself. Availability,
storage-denial and crash populations remain empirical gates, not assumptions.
No unsafe localStorage "check then set" fallback is permitted. Exact marker shape,
lifetime, unsupported-environment behavior and bounded waits require the decisions
and measurements below before implementation.

### S4 — Same request identity, bounded continuation

On successful pair persistence, retry only the original exact authentication
failure, at most once. Preserve method/path/query, serialized body, intent ID,
expected revision and all existing idempotency keys. Never manufacture a new
gameplay action or treat a 200 application rejection as an auth failure.
If rotation reveals a different Founder binding, persist the valid new pair but
do not resubmit an old gameplay command into that Founder: perform authoritative
full sync and require a new player action. Local JWT inspection is only a routing
safety check, never permission or authoritative game state.
If a refreshed request is again refused, surface it normally without a loop.
An ambiguous request/commit is governed by existing receipt/recovery rules, not
by this one-time 401 continuation. Pure reads remain pure reads.

### S5 — Reconnect with ordinary history/recovery, no WS refresh bypass

When a pair rotates, close the old authenticated socket intentionally and
reconnect using the persisted new access token and existing per-Founder stream
positions. Do not emit Centrifuge refresh commands: production rejects them.
Recovered/live publications use the same decoder, offset/cursor dedupe and
authoritative full-sync fallback. Drain delay remains honored; replaced connections
remain terminal. Unsubscribe/disposal must suppress late renew/reconnect delivery.
Account/Founder changes must not deliver the old Founder's recovered publications
into another identity. No second history store or client-authored receipt exists.

## Acceptance criteria (proposed, none completed)

1. Exact registered refresh request/success/error alternatives, generated drift and
   compatibility controls; a wrong pair/schema/status must fail the oracle.
2. Actual runtime: snapshot, intent and all three operation ports resume after
   renewal; concurrent consumers perform one rotation. Severing any port binding
   must fail its own arm while valid-token controls remain green.
3. Two actual browser contexts share one family, receive overlapping expiry,
   rotate once, persist the same descendant and recover. Disable coordination to
   demonstrate consumed-token replay/family revocation is detected, not tolerated.
4. Delayed old replies, other-account replacement, storage denial, missing lock,
   tab closure before/after dispatch and lost refresh replies fail honestly without
   replay, lost recovery material, replacement bootstrap or silent continuation.
5. Exact original intent ID/body/key on one permitted retry; auth refusal commits
   nothing, successful continuation commits once, replay returns the same receipt;
   200 rejections, ambiguous commit, 429 and non-auth failures do not enter renewal.
6. Real Postgres/HTTP/WebSocket: expiry → rotated access → ordinary history recovery
   or authoritative full sync, preserving all committed receipts and latest state;
   revoke-family and replaced-socket negative populations still fail closed.
7. Unmodified real fifteen-minute Garden journey completes native tick growth,
   mature single/all harvest, cash/quota/log/hash binding and reload without any
   test-operated rotation, forced server clock, access-row edit or grant. Separately
   exercise hidden/background resume across expiry. Preserve earlier red attempts.
8. Root client/type/build/boundary, all declared browser engines, related cold
   Account/Transport/Gameserver Postgres lanes and CI-topology pass, with every live
   handle terminal. Historical RP-131/hosted reliability are not waived or claimed
   solved by this lane. Exact range gets mandatory cross-party review; docs and
   tracking reconcile, no self-archive or public activation.

## Open questions / acceptance blockers

- **Owner/author boundary S-OD1:** adopt demand-driven 401/4001 renewal, ordinary
  reconnect and conservative ambiguous-completion stop. This is not an existing
  owner ruling. Adopt exact marker lifetime/schema and unsupported-storage/lock
  posture without silently settling broader D-005/D-009/D-015 obligations.
- **Empirical prerequisite S-R1:** test exclusive Web Locks, shared storage,
  backgrounding and crash/commit ambiguity on supported browser/OS profiles.
  Predeclare populations and failure controls; choose bounded network/ownership
  observation budgets from evidence, never convenient arbitrary ceilings.
- **API author boundary S-A1:** pin the complete existing refresh error/status
  schema and exact generated dispatcher changes needed for C9. The current
  handwritten runtime calls are not permission for another raw API client or an
  unreviewed broad SDK migration. Split that foundational lane if necessary.
- **Recovery dependency S-A2:** ambiguous outcomes must remain honest even while
  D-005's actual recover-existing-account consumer/copy is unfinished. Specify the
  interim visible state and final release dependency; do not invent owner copy or
  claim a player can already recover from the current UI.

## Deviations from design

No intended mechanical deviation. The proposal closes a missing consumer for
existing session authority; the new browser concurrency/ambiguity policy is
explicitly unadopted. Full nine-tier 1.0 and the complete platform floor remain.

## Changelog

- 2026-10-06: proposed after bounded real-stack diagnosis; remains draft, no
  automatic authentication production code changed or owner acceptance recorded.
