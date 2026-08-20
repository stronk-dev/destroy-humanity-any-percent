# Research and verification queue

No measurement may begin until its row is complete. Passing and failing outcomes both close the
research question; they authorize only the stated downstream action.

## R-001 — current-head harness non-termination / CI timeout

- **Question:** Why does `balance-harness -mode=check` fail to finish inside the 30-minute hosted
  budget at `190a4fa`, despite older records claiming a measured green lane?
- **Population:** every active relevance-registry entry and scenario dispatched by `mode=check` at
  current HEAD.
- **Completed first wave:** `r001-harness-diagnosis.md` proves five hosted cancellations, a
  current local green completion at 12 workers, workload expansion since the last hosted green,
  and that worker control excludes relevance. It also rejects scenario-only relevance runs as a
  substitute for active registry authority.
- **Authority gate:** RP-059. The active CI RFC explicitly excludes the balance harness and the
  relevant harness/content RFCs are archived; accept a CI amendment or harness-observability RFC
  before adding the measurement instrument.
- **Remaining arms:** instrumented native macOS and hosted Linux at 12 workers, split by standard
  pacing phase and authority-preserving registry row. Run 1/4-worker comparisons only for phases
  whose dispatcher actually accepts that control; the original whole-command 1/4/12 arm would be
  structurally misleading because relevance ignores the flag.
- **Measurements:** task cardinality, transitions, exclusions, guard exhaustion, wall time per row,
  CPU utilization, and final objective reached.
- **Controls:** a tiny known-completing fixture; an intentionally non-completing/guard-fired fixture
  that must be reported invalid; output byte equality across worker counts.
- **Exit:** every row identifies its cost and the full check completes with a measured margin under
  its governed budget, or the instrument reports a specific invalid termination loudly.
- **May authorize:** a CI/harness RFC amendment and implementation plan grounded in measured cost.
- **Cannot authorize:** raising a timeout from an incomplete run, skipping scenarios, weakening
  bounds, or relabeling a partial result green.

## R-002 — 1.0 release-floor audit

- **Question:** Which exact user, operator, rights, accessibility, packaging, and preservation
  outcomes are mandatory for Cloud Clicker 1.0 versus a development preview?
- **Population:** every release-relevant promise in design, active/archived RFCs, research findings,
  docs, current UI, deployment artifacts, and legal/licensing records.
- **Method:** map each obligation to producer, consumer, workflow, proof, failure behavior, and
  owner disposition. `planning/platform-alignment/release-platform-audit.md` is the tracked HEAD
  baseline; the internal research copy follows its local-only policy.
- **Controls:** include at least one intentionally absent capability and one backend-only capability
  so the instrument cannot equate a route with a user outcome.
- **Completed baseline:** the release audit and four bounded dossiers cover packaging/config,
  account rights, accessibility, operations/retention, provider-off behavior, and preservation at
  the product coordinate. `owner-ruling-packet.md` presents exact label/floor options.
- **Exit:** owner rules D-001 with an explicit in/out/deferred list and canonical design home.
- **May authorize:** deployment/account/accessibility/sunset RFC drafting.
- **Cannot authorize:** implementation or a release label by itself.

## R-003 — nontechnical account recovery and rights study

- **Precondition:** D-005/D-008/D-009 semantics adopted and an accepted prototype exposes every
  task. The study validates the posture; it cannot choose recovery/export/retention policy.
- **Question:** Can a first-time player preserve, recover, export, and delete an anonymous account
  without developer tools or operator help?
- **Population:** at least five nontechnical participants on a clean browser profile; include
  storage-loss, malformed/empty credential storage, expired access token, and second-device cases.
- **Tasks:** find/save recovery credential, recover after logout/storage loss, export data, delete
  account, understand retained anonymized history.
- **Threshold:** 5/5 complete each shipped task without help; zero unrecoverable accounts; every
  destructive action accurately previews retained/deleted data.
- **Controls:** a deliberately unavailable task must be reported unavailable, not inferred from
  Settings copy; network failure during each mutation must preserve a recoverable state; corrupt
  credentials must not silently mint a replacement account.
- **May authorize:** account UX/RFC acceptance evidence and copy questions for owner adoption.
- **Cannot authorize:** retention policy or user-facing wording.

## R-004 — full first-hour browser acceptance

- **Precondition:** Game UI AC1 body reconciled by the ruling author and a governed clock seam
  accepted.
- **Population:** all three live first-ending branches, clean/re-entry/offline cases, Chromium,
  Firefox, WebKit, and keyboard-only operation.
- **Threshold:** each path reaches the authoritative run-end and run-2 starter state; server and
  visible UI coordinates agree; no hidden fixture mutation; axe has no serious/critical findings.
- **Negative controls:** stale Founder revision, expired offer, dropped socket, corrupted stored
  bootstrap receipt, and a keyboard trap fixture must all fail visibly and recover honestly.
- **May authorize:** Game UI archival review.
- **Cannot authorize:** broader mobile/screen-reader accessibility claims.

## R-005 — accessibility release audit

- **Precondition:** D-001 names the exact release tasks and their surfaces exist under accepted
  owners. Already-observed focus/reflow/shell-motion failures remain valid implementation inputs.
- **Question:** Does the ruled release workflow work with keyboard, screen reader, 200%/400% zoom,
  reduced motion, coarse pointer, and common color-vision states?
- **Population:** every release-floor task, not just all rendered components.
- **Thresholds:** WCAG 2.2 AA automated checks plus successful manual workflows; no focus trap,
  inaccessible dynamic update, motion-only information, or pointer-only action.
- **Controls:** seeded focus-trap and unlabeled-status fixtures must be caught.
- **Observed negatives to retain:** lifecycle Offer/Run-End replacement must not drop focus to
  `<body>`; the complete Desk must not exceed a 320 CSS-pixel viewport; CSS-only reduction must not
  mask an unreduced numeric shell.
- **May authorize:** accessibility acceptance and release-floor closure.
- **Cannot authorize:** unsupported disability categories or later-tier surfaces not tested.

## R-006 — clean-host self-host, backup, and restore rehearsal

- **Precondition:** a reconciled Deployment RFC and owner decisions D-002/D-003/D-006.
- **Population:** clean supported host; empty database; populated database; corrupt/partial backup;
  old supported save schema; missing/current/previous secrets; wrong public origin/proxy depth;
  gameserver restart mid-write; package with repository source tree unavailable.
- **Threshold:** documented install reaches readiness; backup restores byte/identity-equivalent
  player and epoch state; missing secrets and invalid backups fail closed; rollback preserves the
  migration invariant.
- **Controls:** restore into the wrong epoch/artifact set and a truncated backup must be rejected.
- **May authorize:** deployment/recovery archival and a self-host claim.
- **Cannot authorize:** a durability or disaster-recovery promise beyond the rehearsed topology.

## R-007 — retention and observability rehearsal

- **Precondition:** D-001/D-006/D-009 and an accepted operations/retention owner.
- **Question:** Can the chosen operator detect, diagnose, retain, delete, restore, and disclose the
  ruled production data without silent loss, indefinite accidental retention, or identifier/IP
  leakage?
- **Population:** every persisted table/data family, application/security logs, metrics labels,
  request/job identities, dead letters, credential/intent/inactivity cleanup, alert path, backup,
  and deletion/export disclosure in the chosen topology.
- **Threshold:** every family has purpose, legal/product basis, duration/event trigger, cleanup
  owner, export/deletion behavior, and observable success/failure; fired alerts reach the named
  operator; cleanup is bounded/idempotent and preserves ruled immutable history.
- **Controls:** a stuck cleanup job, expired intent row, inactive anonymous account, dead-letter
  buildup, failed backup, high-cardinality/private metric label, raw-IP log, and missing request ID
  must each fail the corresponding gate loudly.
- **May authorize:** operations/retention acceptance and exact privacy/incident documentation.
- **Cannot authorize:** deleting replay authority, choosing retention periods, or selecting a
  monitoring vendor without the owner/legal rulings.

## R-008 — first-session comprehension and guidance study

- **Precondition:** D-001/D-007 choose the milestone and an accepted UI/content build exposes its
  complete first-session path. Owner-authored copy remains owner-controlled.
- **Question:** Can a first-time nontechnical player understand the contract, discover the core
  loop, recover from a refusal/outage, and explain the first Exit without outside instructions?
- **Population:** at least eight new-to-repository participants across desktop and supported mobile
  input; include keyboard-only and reduced-motion users without conflating this with R-005's AT
  conformance study.
- **Tasks:** start, identify the resource/rate/cap, buy and change quantity, understand an
  authoritative refusal, leave/return, handle a connection loss, reach the chosen session goal,
  and explain the consequence before confirming Exit.
- **Threshold:** 8/8 finish the required safety/recovery tasks; at least 7/8 finish the core loop
  without intervention; zero participants accept a destructive/terminal action under a materially
  false expectation.
- **Controls:** one intentionally unavailable action must be reported unavailable; a misleading
  hint and a hidden cap explanation must be detected by the observation rubric rather than scored
  as participant failure.
- **May authorize:** guidance/hint density, discoverability acceptance, and exact copy questions for
  owner adoption.
- **Cannot authorize:** changing mechanics, accessibility conformance, or invented player copy.

## R-009 — aggregate gameplay telemetry and milestone calibration

- **Precondition:** D-016 adopts a telemetry schema/privacy posture and an accepted instrument owner
  exists. No public milestone may launch first and use its failure as calibration.
- **Question:** Does the ruled aggregate-only population measure throughput and uncertainty well
  enough to set a community threshold without collecting unnecessary player histories?
- **Population:** synthetic discriminators plus a consent/lawful-basis-reviewed preview population
  representing idle, check-in, active, new, returning, bot-only, and low-population conditions.
- **Threshold:** published aggregation reproduces known synthetic totals/bounds, suppresses or
  widens low-population estimates as ruled, and a predeclared threshold remains inside its target
  completion interval under uncertainty margins.
- **Controls:** duplicate/replayed events, missing cohorts, bot-only traffic, late events, clock
  skew, private identifiers, and a 48-hour burn-through scenario must fail or be visibly excluded.
- **May authorize:** one exact milestone threshold and its published formula under an accepted
  content epoch.
- **Cannot authorize:** hidden DDA, per-player profiling, indefinite raw events, or a reusable
  threshold for later populations.
