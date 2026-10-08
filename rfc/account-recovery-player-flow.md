# RFC: Account Recovery Player Flow

- **Status:** draft — NOT implementation authority
- **Author:** Codex (proposal; copy and implementation contract not adopted)
- **Created:** 2026-10-08
- **Design refs:** `design/00-vision.md` pillars6–7; `design/06-tech.md` anonymous-first accounts and server authority; `design/08-satire-flavor.md` §1
- **Parent / amends:** Account & Session Bootstrap D1–D3; archived Game UI Screens bootstrap persistence and startup consumer
- **Depends on:** existing API Foundation generated `create_bootstrap`, `create_session`, `get_founder` and `get_game_ui_snapshot` operations
- **Tracking:** D-005 / RP-123; `planning/account-and-session-bootstrap/log.md`

## Outcome and authority

A player can save the account ID and recovery code during setup, then use those two values in a
different browser to resume the same account and its current Founder. This implements the
**already ruled** D-005 posture: one-time display, copy/download, recover-existing-account, no
email recovery. It does not reopen that choice.

The missing implementation contract and exact copy below remain proposals. The accepted Garage
manifest F01 explicitly calls for this successor; it is no longer undrafted, but not accepted.
Export, account deletion/retention, browser automatic renewal, offline-save import, switching
between different accounts, New Founder UI and the full 1.0 release floor remain separate.

## Proposed contract

### R1 — Save the credential before entering gameplay

Keep the existing idempotent bootstrap operation and its exact request, encrypted receipt,
expiry/tombstone semantics and persist-before-Desk ordering. Do not create an account through a
second route, regenerate a key after ambiguous failure, or retry by issuing a different identity.

After validating a bootstrap response, durably store the session pair, account ID and pending
recovery setup together. Show a setup panel containing the account ID and recovery code as
selectable, labelled text, the warning below, explicit Copy and Download buttons, and an explicit
saved-credential acknowledgement. Enter gameplay only after that acknowledgement is successfully
persisted. Reload before acknowledgement resumes this same panel, not another bootstrap; a
resumable setup panel is not a second completed account-creation event.

The acknowledgement does not require clipboard support or a download: selecting and saving the
displayed text manually is supported. Clipboard rejection stays visible and leaves the values
available. Download is a user-initiated local UTF-8 text file containing exactly the account ID
and recovery code, with non-secret field labels; its filename is `cloud-clicker-recovery.txt`.
No credential goes into a URL/filename. Do not claim that initiating a download proves it was saved.

### R2 — One credential document; explicit legacy handling

Use one versioned browser credential document at `cloud-clicker.credentials.v2`:

```text
{version:2, accountID, accessToken, refreshToken,
 recoverySetup:null | {accountID, recoveryCode}}
```

The nested account ID must match the outer one. Pending setup holds the recovery code only until
acknowledgement; acknowledgement replaces it with `null` in the same single-item write. Do not
persist a second permanent copy or erase the only pending copy before a successful write.
Storage failure blocks the transition visibly, preserves the in-memory credential, and offers
another explicit save attempt without another account/session request.

Read the existing v1 document without discarding it. On a validated authenticated startup, migrate
its session/account/code into pending v2 setup and show the previously hidden credential once.
Remove v1 only after the v2 write succeeds. If both keys survive interrupted cleanup, valid v2
takes precedence; finish removing v1 before claiming completed setup. A failed write/removal must
not falsely claim the secret is cleared. A completed v2 setup never displays the code again.

This specifies this application's local credential lifecycle, not deletion of browser backups,
clipboard contents or downloaded files. Do not describe localStorage as encrypted or secure
against same-origin script access. No server retention or deletion policy changes here.

### R3 — Recover, never silently replace

Offer Recover existing account at entry and in Settings, and keep that affordance reachable when
the first authenticated read fails. Opening it does not erase stored credentials, create an
account, reset progress or interpret a network failure as an absent account. This does not add
general startup Retry semantics or change D-023's unresolved failure-state precedence.

The form asks for account ID and recovery code. When local credentials identify an account,
recovery is restricted to that same account and the ID is read-only; different-account switching
is out of scope. With no local identity, the player supplies both values. Reuse existing server
credential admission/normalization, not a stricter invented credential grammar.

Submit once through generated `create_session`. Keep the form busy, prevent duplicate Enter/click
submissions, and retain focus. A successful session response is not yet recovered gameplay:
using the candidate access token, obtain the authenticated current Founder and valid live Game UI
snapshot, require their Founder identities to agree, then persist the new session document before
binding gameplay. Recovery must not call bootstrap or any gameplay intent.

If login succeeds but the read fails, retain the candidate pair only in this attempt's memory;
Retry load reuses it and performs reads only. It must not resubmit the recovery code or replay a
single-use refresh token. A failed persistence similarly retries the validated local write only.
Keep any prior durable credentials intact until replacement succeeds. Recovery code entered in
this form is never persisted; clear it on successful recovery, cancellation or unmount.
If the prior document has pending recovery setup, preserve that setup with the new session pair
and return to its acknowledgement panel; logging in cannot silently erase its only saved code.

On successful commit, replace the old subscription and prediction state with the recovered
authoritative identity, reset identity-scoped revision cursors, and focus the Desk heading only
when setup is complete (otherwise focus the setup heading). Old
in-flight work must not write credentials, rebind UI or reopen a socket after cancellation,
unmount or replacement of its observed credential generation/account. Check the current document
before committing. This does not claim atomic cross-tab refresh coordination; that is owned by
the separate browser-renewal proposal.

### R4 — Honest failures, keyboard use and secret handling

Use exact existing operation/status alternatives: malformed input, denied credentials and rate
limit remain distinct from network/unexpected failures. A generic credential denial must not
reveal whether an account exists. Do not expose raw server bodies or submitted values in notices.
No automatic retries, account replacement, email promise or changed server TTL/security rule.

All fields/buttons are native keyboard-operable, with explicit labels and associated warnings.
Pending controls remain focusable with guarded activation; results use a status/alert region.
After a refusal, focus stays usable for correction; after cancellation it returns to the opener.
Do not trap Tab or override a newer focus/navigation choice. Support 320px reflow and reduced
motion. This does not replace the unaccepted cross-product task/assistive-technology matrix.

Secrets never enter analytics, console messages, captured test diagnostics, request IDs, route
parameters or public events. Test drivers must redact credentials while retaining status,
operation and identity-equality outcomes. Copy/download occur only on the player's gesture.

## Proposed copy — requires explicit owner adoption

Plain, sincere utility text; no satire around account loss. These are proposals, not owner copy.
Implementation places adopted text in the existing copy catalog, not component string literals.

| Key suffix (`account.recovery.`) | Proposed English |
|---|---|
| `setup_title` | Save your recovery details |
| `account_id` | Account ID |
| `code` | Recovery code |
| `warning` | Keep both values somewhere you can find them. Anyone with both can access your account. Without them, you may be unable to recover your account after losing browser access. Email recovery is not available. |
| `copy` / `copied` / `copy_failed` | Copy recovery details / Copied / Could not copy. You can select and save the text instead. |
| `download` | Download recovery details |
| `acknowledge` / `continue` | I have saved my recovery details / Continue |
| `open` / `title` / `submit` | Recover existing account / Recover your account / Recover account |
| `cancel` / `pending` | Cancel / Recovering account… |
| `invalid` / `denied` | Check the account ID and recovery code. / Those recovery details were not accepted. |
| `limited` / `failed` | Too many attempts. Please wait before trying again. / Could not recover the account. Your saved account details have not been replaced. |
| `load_failed` / `retry_load` | Signed in, but could not load your game. / Retry load |
| `storage_failed` / `retry_save` | Could not save account details in this browser. Keep your recovery details and try again. / Retry save |
| `changed` | Account details changed in another window. This recovery attempt has stopped. |

Copy and download use the same two labelled values; no tokens or game-state/export payload.

## Acceptance — proposed, none completed

1. Real built client/server/Postgres: bootstrap → save panel → native Copy/Download/manual
   acknowledgement → Desk. Match the actual issued ID/code privately; verify session persisted
   before gameplay, recovery secret removed afterwards, and reload does not redisplay it.
2. Close/reload before acknowledgement, deny clipboard, fail storage writes/removal, and load a
   real legacy v1 credential. Preserve the only recoverable values and the original account;
   no additional bootstrap or false saved/copied/cleared claim. Retain focused runtime/DOM tests.
3. In a clean browser context, enter the saved details through native controls and recover the
   same current Founder, Company, persisted progress and usable socket. No bootstrap/new account,
   Founder creation or replayed gameplay command. Wrong-code and unknown-account cases fail
   without revealing account existence or changing gameplay state.
4. Exercise rate limiting, failed post-login read, failed persistence, delayed replies, cancel,
   unmount and changed stored generation. Assert which exact operation executes once; read/save
   retries cannot create another session or replace newer credentials. Old UI/socket work stays
   detached. Never print the credential values while testing these boundaries.
5. Native keyboard and 320px checks cover setup, correction, cancellation and resumed Desk focus;
   affected client type/build/boundary, generated API and real Postgres populations pass. Extend
   the existing composed journey rather than introducing another generic test framework.
6. Canonical account/Game UI docs reflect actual behavior; exact implementation range receives
   designated cross-party review. R-003 evaluates the built player task before a release claim.
   No full-CI, assistive-technology, player-rights or complete 1.0 claim follows from this feature.

## Acceptance blockers and deviations

- Marco must adopt or edit the proposed credential-document migration, acknowledgement and
  recovery-attempt policy, and explicitly adopt the copy block. D-005's posture stays ruled.
- Designated specification review is required under the existing process. No implementation,
  owner-copy adoption, broader renewal/rights ruling or feature completion is implied by drafting.
- No intended gameplay deviation; this changes the browser consumer/storage contract and makes
  the existing server credential usable by the player. No server endpoint, migration or token
  lifetime change is proposed.
