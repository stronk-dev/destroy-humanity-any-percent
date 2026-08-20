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
- **Exit:** owner rules D-001 with an explicit in/out/deferred list and canonical design home.
- **May authorize:** deployment/account/accessibility/sunset RFC drafting.
- **Cannot authorize:** implementation or a release label by itself.

## R-003 — nontechnical account recovery and rights study

- **Question:** Can a first-time player preserve, recover, export, and delete an anonymous account
  without developer tools or operator help?
- **Population:** at least five nontechnical participants on a clean browser profile; include one
  storage-loss case and one second-device case.
- **Tasks:** find/save recovery credential, recover after logout/storage loss, export data, delete
  account, understand retained anonymized history.
- **Threshold:** 5/5 complete each shipped task without help; zero unrecoverable accounts; every
  destructive action accurately previews retained/deleted data.
- **Controls:** a deliberately unavailable task must be reported unavailable, not inferred from
  Settings copy; network failure during each mutation must preserve a recoverable state.
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

- **Question:** Does the ruled release workflow work with keyboard, screen reader, 200%/400% zoom,
  reduced motion, coarse pointer, and common color-vision states?
- **Population:** every release-floor task, not just all rendered components.
- **Thresholds:** WCAG 2.2 AA automated checks plus successful manual workflows; no focus trap,
  inaccessible dynamic update, motion-only information, or pointer-only action.
- **Controls:** seeded focus-trap and unlabeled-status fixtures must be caught.
- **May authorize:** accessibility acceptance and release-floor closure.
- **Cannot authorize:** unsupported disability categories or later-tier surfaces not tested.

## R-006 — clean-host self-host, backup, and restore rehearsal

- **Precondition:** a reconciled Deployment RFC and owner decisions D-002/D-003/D-006.
- **Population:** clean supported host; empty database; populated database; corrupt/partial backup;
  old supported save schema; missing secret; gameserver restart mid-write.
- **Threshold:** documented install reaches readiness; backup restores byte/identity-equivalent
  player and epoch state; missing secrets and invalid backups fail closed; rollback preserves the
  migration invariant.
- **Controls:** restore into the wrong epoch/artifact set and a truncated backup must be rejected.
- **May authorize:** deployment/recovery archival and a self-host claim.
- **Cannot authorize:** a durability or disaster-recovery promise beyond the rehearsed topology.
