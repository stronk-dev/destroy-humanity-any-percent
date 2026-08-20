# Account recovery and rights release audit

Coordinate: product tree `190a4fa`; audit checkpoint after `7910fc2`; 2026-08-20.

This pass traced the actual player journey from bootstrap response through browser storage,
authenticated HTTP/WebSocket use, Settings, recovery/session endpoints, Founder import/switch,
account deletion, and export search. It re-used the same-coordinate cold Account/composed-browser
evidence recorded by the Account and Game UI lifecycle audits. No product, API, copy, RFC/design,
canonical-product-doc, schema, or player-data change was made.

## Bottom line

The backend account core is substantial. Bootstrap atomically creates an anonymous account,
Founder streams, session family, exact snapshot, and encrypted retry receipt. Recovery-code login,
single-use refresh rotation/reuse revocation, session revocation, Founder replacement/import, and
account anonymization have real Postgres implementations and tests.

None of the player recovery or rights journey is composed:

- bootstrap silently stores the one-time recovery code, refresh token, access token, and account ID
  together in `localStorage`; the player never sees or confirms possession of the recovery code;
- the runtime never calls recovery login or refresh. The 15-minute access token is reused for every
  HTTP and WebSocket operation until failure becomes the generic offline state;
- clearing/corrupting credentials routes the player to the cold open and creation of another
  account. There is no “recover existing account” branch, warning, or orphan-account handling;
- Settings contains only save-age/status and one account paragraph. It has no recovery/download,
  logout/revoke, email, import, New Founder, export, deletion, or retained-data disclosure;
- optional email always returns 501 `not_configured/email_provider`; and
- no export endpoint, schema, serializer, client action, portable import pairing, or proof exists.

The result is not “anonymous and complete” from a nontechnical player's perspective. It is a
server account whose only durable credential is hidden in one browser profile.

## Producer → consumer → workflow trace

| Outcome | Backend producer | Production consumer | Player workflow | Verdict |
|---|---|---|---|---|
| Create anonymous account | Atomic `/api/v1/bootstrap`, protected retry receipt, `no-store` | Game UI runtime | Begin Attempt | **Proven to silent creation** |
| Preserve recovery credential | One-time `recovery_code` in bootstrap response | Raw localStorage document only | None: never displayed/downloaded/confirmed | **Mechanical storage only** |
| Recover existing account | `POST /api/v1/session {account_id,recovery_code}` | No client call | None | **Backend-only** |
| Survive token expiry | `POST /api/v1/session/refresh`, family rotation/reuse revocation | Refresh token is stored but never read after parsing | Failure becomes offline; no retry/reconnect | **Backend-only; live session breaks at 15 min** |
| Log out/revoke | `DELETE /api/v1/session` | None | None | **Backend-only** |
| Attach recovery email | Mounted endpoint | None; handler always 501 | Settings says “if you ever add one” but offers no action | **Unconfigured/absent** |
| Start/import another Founder | `POST /api/v1/founder`, `/founder/import` | None | None | **Backend-only** |
| Delete account | Transactional archive/anonymize/delete endpoint | None | None; no confirmation or retention preview | **Backend proven; rights workflow absent** |
| Export player data | None | None | None | **Absent** |

## Storage and failure behavior

1. `credentials()` validates only that the four fields are strings. Empty strings therefore count
   as credentials, start the Desk path, and then fail generically. Invalid JSON/missing fields count
   as no credentials and start a new bootstrap path. Neither state explains recovery.
2. The bootstrap idempotency key is correctly persisted before the request and removed only after
   credentials are stored. This protects a retry during the server's receipt lifetime; it does not
   give the player a second-device or post-storage-loss recovery mechanism.
3. All four secrets/identifiers share one localStorage record. Losing that record loses the only UI
   route to the account; script access to that record exposes both immediate and recovery/session
   credentials. The appropriate posture is an owner/security decision, not an audit-authored fix.
4. `GameUINavigation.settingsConfirmation()` can defer lifecycle preemption behind a destructive
   Settings confirmation, but production never calls it and renders no such confirmation. The
   unit test proves an orphaned safety primitive, not a deletion experience.
5. Settings copy says offline progress is “parked on this machine until the server picks up again.”
   The runtime has no local authoritative save, persisted intent queue, local gameplay import
   owner, or reconnect/flush path. It merely sets `offline=true`; the sentence promises behavior
   that does not exist.

## Existing executable evidence and its limit

- Cold real-Postgres Account tests prove bootstrap/session/refresh/revocation/import/deletion
  mechanics and anonymized save survival.
- The real composed Chromium test proves first bootstrap credentials are persisted before the
  snapshot is exposed, then reaches Desk and a live world subscription.
- Browser component tests prove Settings renders, has no axe serious/critical finding, preserves
  focus through an era update, and reports snapshot age.

Those tests contain no player recovery, expiry refresh, logout, email, import, Founder switch,
export, or delete action. The Settings axe pass cannot promote controls that do not exist. R-003's
deliberately unavailable task control therefore fires: export/delete/recovery must be reported
unavailable rather than inferred from backend routes or prose.

## Contract and copy drift

1. Account D1 says the recovery code is “shown once.” Production never shows it; it stores it
   silently.
2. Account D2 promises ops-configured current+previous signing keys. Deployment audit RP-076 shows
   only the current key reaches production composition.
3. Account D4 still promises fully offline-anonymous play, while the newer owner ruling selects
   silent server-anonymous default plus labeled outage fallback. The fallback is absent.
4. Settings' email sentence gestures at an unavailable future endpoint without an action or
   `not_configured` disclosure.
5. Settings' offline sentence contradicts the runtime and can mislead a player into believing
   failed actions/progress are durable locally.
6. Canonical account docs accurately describe backend deletion, but there is no canonical player
   rights/recovery workflow because none exists.

## Smallest honest next order

1. Owner rules D-005 recovery posture, D-008 export contents/portability, D-009 retained-history
   disclosure, and the unresolved storage/abandoned-account posture behind RP-051/RP-079.
2. Ruling authors reconcile Account D1/D4 and the newer server-default/local-fallback intent.
3. Accept one successor contract owning refresh→retry→socket reauth, recovery-code presentation and
   existing-account recovery, logout, export, deletion, storage-loss/corruption behavior, exact
   Settings confirmations/disclosures, and offline fallback. Email may remain explicitly deferred.
4. Implement the generated/thin API client dependency rather than adding another raw HTTP path.
5. Run R-003 with clean, storage-loss, second-device, expired-token, corrupt-storage, and
   network-failure arms. Only then may the player rights/recovery capability be promoted.
