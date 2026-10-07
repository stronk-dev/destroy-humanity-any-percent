# RFC: Game UI First-read Failure & Recovery

- **Status:** draft — NOT implementation authority
- **Author:** Codex (proposal; no owner adoption claimed)
- **Created:** 2026-10-07
- **Design refs:** `design/00-vision.md` pillars6–7; `design/06-tech.md` frontend, server-authoritative reads; owner-ruled account recovery D-005
- **Depends on:** Garage Player Surfaces GS0.5/0.6; API Foundation; Account Bootstrap
- **Parent / amends:** archived Game UI Screens startup binding; Garage shared startup-state precedence
- **Planning:** measured motivation in `planning/garage-player-surfaces/log.md` and `acceptance-evidence.md`; implementation plan only after acceptance

## Summary

A returning player whose first authoritative read fails must not be left looking
at a completed request labelled loading. Provide an honest failure state and a
read-only retry without losing the stored account or creating a replacement.
Everything below is proposed until its open boundaries are resolved and accepted.

## Executed motivation

RP-342: the actual browser runtime, Response.json and production decoder feed a
mounted host in Chromium/WebKit at320/1280. Twenty-four failure executions cover
network rejection,401,503, malformed JSON, malformed v4 arm and valid legacy v3.
All complete the read with `aria-busy=false`, but visible loading remains;
offline/error text, alerts, controls and diagnostics are absent. Four healthy
controls render Desk. An explicitly dispatched visibilitychange requests a
second read; valid same-Founder replies recover, repeated401 remains unresolved.
No bootstrap/refresh/gameplay write or credential mutation occurs. These are
injected HTTP/native host observations, not real server auth or physical resume.

GS0.5 makes no-snapshot loading and last-snapshot reconnect explicit, but does
not resolve first-read failure precedence or name a read retry interface.
Bootstrap's idempotent write retry and the unaccepted browser-session-renewal
proposal are not authority for either. This RFC closes that boundary, not a
general client/network redesign.

## Proposed specification

### S1 — State precedence and ownership

The host owns `unread | reading | failed | bound` for the first authoritative
read. Loading is shown only while unread/reading; a settled failure takes
precedence over the no-snapshot loading branch. Do not bind malformed/legacy
data, fabricate resources, create a prediction worker/socket before a valid
snapshot, or expose gameplay controls from an absent snapshot. Credentials
continue selecting the returning-player path; absence still selects Vision.

First-read state does not redefine later snapshot refresh, mounted-arm errors,
Offer/Run-End lifecycle precedence, pending intents or terminal socket policy.
Later refresh failures continue retaining previously authoritative data.

### S2 — Exact refusal classification, no credential mutation

Preserve HTTP status and only validated public error information at the runtime
boundary. Do not log tokens, recovery codes, response bodies, raw mechanical
pairs or arbitrary thrown messages. Pin the existing founder-state operation's
complete response alternatives before acceptance; no test-side status registry.

Proposed rendering categories:

| Failure | Proposed visible state | Proposed action |
|---|---|---|
| Network or registered transient server refusal | Read unavailable, not still loading | Read-only Retry |
| Registered access-token401 | Session unavailable; existing account preserved | Explicit read-only Retry; later accepted renewal/recover-existing-account owns credential recovery |
| Malformed JSON, malformed/unsupported live schema | Data unavailable; no partial game surface | Read-only Retry; one bounded invariant per failure episode |
| Other registered refusal or unrecognized status/body | Fail closed, no raw details | Read-only Retry; exact classification/diagnostic policy must be pinned |

Retry never clears/replaces credentials, calls bootstrap/refresh, renews tokens,
changes Account TTL/rotation, sends a gameplay intent or fabricates a new account.
Repeated401 may remain refused; this RFC does not claim to solve session renewal.
An eventual accepted renewal coordinator may be consumed separately without
weakening these identity boundaries.

### S3 — One request at a time, no hidden loop

Retry performs the same existing authoritative GET, through the current runtime.
Reuse host refresh single-flight rather than a second queue. One native activation
while ready makes at most one read; activations while it is pending do not create
additional reads. Preserve focus on the Retry control while pending, with existing
pending text and aria-disabled semantics. On success, bind only a validated
snapshot, start existing subscriptions and remove the failure/retry region.

Existing visibility lifecycle can request a read through the same single-flight
path. An older completion cannot replace a newer account/snapshot or update a
disposed host. No new retry timers, backoff policy, deadlines or periodic polling.
Hung-request timeout/cancellation, if required, needs an explicit measured contract.

### S4 — Copy and accessibility

Proposed failure region: one Desk heading with negative tabindex and an own
role=alert; no misleading loading status after completion. Reserve
`startup.read.unavailable`, `startup.read.session_unavailable`,
`startup.read.invalid_data` and `startup.read.retry` for owner-authored/adopted
copy. These are reservations, not catalog edits or shipped strings. Resolve
whether existing generic keys can truthfully replace any reservation before
acceptance. No authored account-loss/recovery disclosure is changed.

Native Enter/Space/Tab/Shift-Tab, no key interception; retry pending stays
focusable, forced successful removal follows the ruled Desk-heading focus
contract unless a newer player choice wins. Failure/pending/success and repeated
refusal must remain readable at320 CSS px, with coarse-pointer target-size,
reduced-motion and manual assistive evidence under the adopted task matrix.

## Proposed acceptance criteria

1. Complete public operation/refusal classification; unknown/malformed bodies
   fail closed. Seeded status/body/legacy-admission changes fail the exact oracle.
2. Actual runtime + host: held/healthy/network/401/server/JSON/schema populations
   render truthful states. A loading-after-settlement fault must fail; no bug is
   blessed by the current observation instrument's green status.
3. Native Retry emits exactly one authoritative GET; held duplicate activation
   and concurrent lifecycle request do not multiply it. Seeded double dispatch
   and bootstrap/refresh/gameplay/credential-clear shortcuts each fail.
4. Valid same-account recovery binds current data, removes failure and preserves
   chosen focus; repeated401 does not bootstrap/renew. Old/disposed completions
   and account-generation changes fail their identity controls.
5. Error/pending/success all-state three-engine, reflow/coarse-pointer/axe/native
   keyboard plus adopted manual AT evidence; seeded missing label/hidden alert/
   focus loss must fail. Browser launch errors are not skipped acceptance.
6. Built client/real service recovery population after capacity repair: unavailable
   first read, actual route recovery, same Founder/run and no extra gameplay or
   account rows. Injected-response observations do not substitute for this gate.
7. Canonical docs, ledger, plans and exact designated cross-party range review
   reconcile before lifecycle/archival. No release claim follows from this RFC alone.

## Open boundaries before acceptance

- Owner adopts/edits the manual read-only Retry posture, especially401 wording:
  it does not promise restored authentication. D-005's unbuilt recover-existing
  workflow and draft browser-session-renewal remain separate dependencies.
- Pin every existing founder-state success/error status/body through API
  Foundation, including unrecognized responses and diagnostic classification.
- Owner authors/adopts exact startup copy or explicitly approves truthful reuse.
- Reconcile the GS0.5 no-snapshot/failed-read precedence in the author's body;
  do not merely append this proposal beside contradictory accepted text.
- Specify late completion/account-generation/disposal handling from actual
  host/runtime identity before implementation, without inventing renewal policy.
- Preserve all-engine/manual AT/real-service evidence and existing full1.0 gates.

## Deviations from design

None proposed for mechanics, timing, authority, account policy or copy ownership.
An explicit startup failure precedence and Retry control are additions to the
existing UI contract, not claims that current accepted text already allows them.

## Changelog

- 2026-10-07: proposed from predeclared RP-342 native/runtime observations;
  no product changes, owner ruling, independent approval or implementation claimed.
