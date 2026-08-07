# Soul Foundation implementation log

## 2026-08-05 — Codex acceptance review: blocked (SB1-SB9)

Review by: Codex. Recorded by: Codex.

The review confirms the intended scalar order (Fiscal v19, then Soul v20), runtime independence of
Trust and Soul, and the rule that this foundation may use only fixture source/activity rows. The
existing Founder save already has a dormant signed `soul` field, so v20 must activate and constrain
that field rather than claim to introduce it.

Nine blockers (SB1-SB9) are filed: legacy-field activation, debit eligibility/atomic benefit, floor-price
saturation, the missing opportunity-costed recovery coordinator, exact artifact grammar,
pet/minigame/UI gate ownership, ending-fact/copy-pipeline contradictions, the deferred correlation
gate, and exact wire/upstream sequencing. No unverified research numbers were promoted into data or
player-facing copy.

## 2026-08-06 — owner rulings on acceptance blockers SB1-SB9 (all resolved)
- SB1: v20 ACTIVATES the existing dormant `soul` field (not "adds"); pre-v20 soul==0/reject-nonzero;
  init to soul_initial (<=max) at first activation, recorded in Founder Exit arm. (S1/S6 reconciled.)
- SB2 (owner): Soul-debit is a COMPONENT inside the owner transition (Event/longevity/contract) so
  benefit+debit commit atomically; NO standalone pay_soul_price in production; fixture-only test command.
- SB3 (owner): full-debit-or-reject (unaffordable/soul if it crosses floor); may_exhaust:true source
  consumes to floor once (single-use). No cheap-at-floor exploit.
- SB4 (owner): touch_grass is a REAL activity lifecycle (start: attended start+duration + Company
  production-suppression interval; resolve: Company+Founder txn grants Soul + ZERO output; mutually
  exclusive with production). NOT an instant command. (S3 reconciled.)
- SB5: soul artifact schema-v1 (policy/bands/source rows/activity rows); fixture never epoch-seeded;
  band-threshold partition of full domain; v20 biconditional, requires v19 Fiscal artifact.
- SB6: EXACT bounded meter+band is VISIBLE/published via the pet panel (transparency law — fixes my
  "never a number" error), never a currency/shop wallet; human_content_locked predicate; recovery/
  essential-care available at floor. (S1/S4 reconciled.)
- SB7 (owner): ending keys on a DATED `soul.depleted` Founder fact (emitted first time a debit hits
  floor) + current band, NOT a live threshold (design/02 §7). Copy semantic-safety is a Copy-Pipeline
  successor; export the enum, don't claim the pipeline flags self-harm. (S5/AC5 reconciled.)
- SB8: runtime Trust/Soul independence normative; the CI correlation gate is a balance-harness successor
  (routed, not dropped).
- SB9: after FQ v19 closes, enumerate v20 codec/migration + wire; soul_band_changed.v1 on real band
  change only; DEPENDENCY-BLOCKED on Fiscal v19 implementation.
Status -> accepted; SB1-SB9 ruled; impl dependency-blocked on Fiscal v19. Body/README reconciled.

## 2026-08-06 — Codex implementation acceptance recheck: blocked (SB10-SB16)

Review by: Codex. Recorded by: Codex.

Checked SB1-SB9 against the shipped Founder save/replay, minigame session, pet action, Company accrual,
and multi-stream coordinator surfaces. Fiscal remains an upstream block, but seven independent Soul
contracts are also open:

- SB10: no exact artifact keys, interval convention, or ending/gate grammar;
- SB11: once-only exhaust has no persisted usage authority;
- SB12: the owner-invoked debit component and dated-fact event order are not byte contracts;
- SB13: touch-grass suppression has no persistent Company/replay owner;
- SB14: start/resolve/cancel/reconnect/Exit/concurrency outcomes are undecided;
- SB15: hobby/essential/recovery classifications and composed resolvers do not exist;
- SB16: v20 activation/wire is explicitly deferred rather than enumerated.

The stale acceptance criteria were reconciled to the already-ruled component/activity/dated-fact
model. Executable proposals preserve one replay-owned Soul balance, keep the forbidden public debit
rail absent, and make the real production opportunity cost auditable. No Soul code or data landed.

## 2026-08-06 — owner rulings on the 3rd-round blockers SB10-SB16 (all accepted)
- SB10: exact soul artifact {policy,bands(partition,only near_zero locked),debit_sources(may_exhaust==
  single_use),recovery_activities,ending_policy}.
- SB11 (owner): soul_exhausted_source_ids set in Founder v20 (may_exhaust rows only); reject
  soul_source_consumed after use.
- SB12: pure ApplyDebit(...) component; owner txn persists eligibility+benefit atomically; event order
  benefit->soul_price_paid->[band_changed]->[soul_depleted+LedgerFact]; package-private test entry only.
- SB13 (owner): touch-grass = soul_recovery_sessions table + suppression segment frozen into replay_inputs
  => NO Company bump (the Fiscal F12 pattern); exclusive_activity reject; resolve advances evaluated_through;
  Founder-then-Company lock order (C38).
- SB14: one active session/founder; attended pauses offline; resolve = zero-output segment + Soul + logs
  atomic (Founder-then-Company, corrected from SB14's wording); cancel=0 Soul; Exit rejects while active.
- SB15: minigame artifact soul_gate {human_hobby|unrelated} + pet actions {essential|recovery|ordinary};
  HumanContentLocked() + band projection; composed resolver; near_zero locks human_hobby+ordinary.
- SB16: Founder v20 retains soul key, revalidates [floor,max], appends eligibility/activity state;
  biconditional floor 20 (needs fiscal+minigames+pets); dependency-blocked on Fiscal v19 impl.
Status -> accepted; SB1-SB16 ruled; impl dependency-blocked on Fiscal v19.

## 2026-08-06 — Codex post-Fiscal implementation recheck: blocked (SB17-SB23)

Review by: Codex. Recorded by: Codex.

Fiscal v19 is now implemented, reviewed, and archived, so the named dependency is closed. Re-read the
full Soul RFC against the v19 Founder codec, Exit resolved arms, minigame session coordinator,
`ApplyLogged` accrual path, event registry, and the current pet/minigame artifact grammars.

Seven executable contracts remain absent: literal artifact enum/registry rules; debit errors and event
bytes; recovery-session schema/commands; the zero-output Company replay boundary; exact v20 extension
and activation evidence; versioned pet/minigame gating bytes; and the cross-stream event/log order.
SB16 explicitly asks a future edit to enumerate several of these and the SB10-SB16 ruling section does
not do so. Implementing now would invent public API and transactional semantics. SB17-SB23 are filed
with concrete proposals; no Soul code or data was introduced.

## 2026-08-06 — owner rulings on SB17-SB23 (all accepted)
Artifact enums + copy-key registry (SB17, ending values are enum IDs never copy keys); ApplyDebit
typed errors + 3 exact event payloads (SB18); the 3 coordinator commands + UUIDv7 session row +
partial-unique active-session index (SB19); ApplySuppressedLogged — the shared zero-output boundary
that advances every watermark while asserting zero output (SB20); Founder v20 = v19 + exactly
soul_exhausted_source_ids, next_soul Exit arm recomputed-from-bytes (SB21); soul_gate via artifact
schema BUMPS, historical schemas valid only in Soul-less bundles (SB22); start/cancel/resolve event
ordering + Founder-then-Company single-transaction atomicity w/ fault injection (SB23).
Status -> SB1-SB23 complete; Soul implementable (Fiscal v19 archived).

## 2026-08-06 — implementation round: persistence/replay complete; recovery liveness DESIGN-GAP

Review by: pending cross-party designated review. Recorded by: Codex.

Implemented the strict Soul package and v20 activation, the pure debit component and exact events,
versioned pet/minigame gates, and the recovery persistence/replay skeleton. The recovery round adds:

- an immutable, claim-tokened `soul_recovery_sessions` lifecycle and one-active-session constraint;
- race-safe transaction guards for Company commands, Founder care, Exit, and minigame start;
- the shared Go/TypeScript `ApplySuppressedLogged` zero-output transition;
- Company-first then Founder-log resolution with one durable terminal receipt;
- exact event registry/database constraints, v20 migration/corpus, cross-runtime replay fixtures,
  full Founder-history verification, and real-Postgres fault injection at every persistence boundary.

The real database suite exposed and fixed three seam defects that unit tests did not cover: the
Founder-log multistream constraint initially excluded the new recovery arm; start and terminal events
share a session ID and therefore Founder-history event loading must scope by applied revision; and
retention attempted to delete immutable Founder genesis revision 1 after the sixth Founder command.
The shared retention helper now preserves the genesis revision in every write/coordinator path.

**DESIGN-GAP — recovery cannot accumulate attended time under the ruled exclusivity model.** SB14 says
ordinary Company commands reject while recovery is active, early resolve rejects without mutation,
and recovery eligibility advances by Founder attended time while pausing offline. The shipped
attendance resolver derives the current-run partial from Company state; that state advances only on a
successful Company transition, and any unresolved gap above the pinned catch-up ceiling is classified
offline. With every ordinary transition rejected, no legal writer advances `evaluated_through` or the
attended partial. A fixture can resolve at exactly the tolerance boundary, but one millisecond later
the gap becomes offline and the session can never become eligible. That exact-boundary fixture is not
accepted as proof of a live mechanic.

Required owner contract: name a server-authoritative way to advance recovery attendance while keeping
production at zero and preserving `rejected = no mutation`—for example, a claim-tokened recovery
presence/progress command or a composed socket-presence interval written through the suppressed
boundary. It must specify reconnect/offline classification, idempotency, replay bytes, and whether
progress commits before terminal resolution. No production `recovery_activities` row may mint until
that contract lands. Implementation remains active and is not ready for archival.

Verification completed from the repository root: replay fixture regeneration/check, focused Go and
TypeScript suites, `make test-save-integration` against the declared Postgres service, and the full
`make verify` gate. The full gate read through completion: Go vet/all packages, balance harness,
formula and schema drift, kernel history/parity at `0.3.73`, TypeScript/Svelte typecheck, production
client build, 6,582 unit assertions, and 19,755 browser assertions all passed. This is self-
verification only; it does not satisfy the cross-party designated review gate and does not authorize
archival.

## 2026-08-07 — SB24 ruled: the recovery progress heartbeat (closes the attendance DESIGN-GAP)
A 4th coordinator command `soul_recovery_progress {session_id, claim_token}` (never an intent):
server-stamped; session-row-only mutation; delta counts iff gap <= recovery_beat_ceiling_ms (catalog,
<= global ceiling), larger gaps add zero — absence PAUSES, never kills. Beats grant nothing and are
not replay bytes (nothing to farm); replay stays terminal-only — the resolve/cancel arm's
founder_attended start/end must equal attended_progress_ms (validated). Eligibility: progress >=
duration. Lazy watchdog: max_session_wall_ms auto-cancels at next touch (SB23 cancel path,
cancelled_by: watchdog, zero Soul) — no background job. rejected=no-mutation preserved (session row =
coordinator state). Production recovery rows may mint only after SB24 implements + reviews.

## 2026-08-07 — designated cross-party verdict: Soul v20 substrate — APPROVE (except the SB24 gap)

- **Review by:** the designated Claude reviewer (independent; make verify + test-save-integration
  re-run, real-Postgres fault injection at all nine boundaries verified). **Recorded by:** same.
- **Range:** the three Soul commits `a3f7f30` (artifact/debit foundation), `8d2e1a6` (v20
  activation), `203d40a` (recovery transaction substrate) — all previously unreviewed, reviewed in
  full here. Gap commits are docs-only EXCEPT `45f082d` (Active-Play RR-remediation code) which is
  explicitly NOT consumed — it belongs to the Active-Play round-3 thread. **Any Soul archival must
  cite these three commits and must not claim 45f082d.**

**Verified sound:** ApplySuppressedLogged runs the FULL Evaluate + hook chain then restores every
output authority (not the forbidden evaluated_through shortcut) with byte-identical Go/TS replay;
coordinator commands unreachable from any parser/transport; SB23 lock order + all-or-none proven at
every fault boundary; v20 codec exact (+1 key), next_soul recomputed-from-bytes both directions;
soul_gate schema bumps with no deploy-current read; ApplyDebit + events exact; genesis-safe retention
(pruning can no longer delete the Founder genesis revision career replay depends on); kernel
0.3.69→0.3.73 lockstep; heartbeat-gap containment verified (no soul artifact in any production
epoch, no endpoint).

**Findings routed to Codex (with the SB24 implementation):**
- LOW-1: the SB20 zero-output "assert" compares pointers it just restored — dead code presenting as
  an assertion. Make it a real check (snapshot outputs BEFORE restore; compare) so a future edit
  that drops a field from the restore list is caught.
- LOW-2: Go/TS band-loader divergence at max_inclusive == MaxExactInteger (Go special-cases, TS
  doesn't) — a pathological catalog validates on the server and is refused by the client. Align (drop
  the Go special-case) + a shared mutation vector.
- INFO-2 (no action): the claim-lease takeover branch is currently unreachable (claim+finish commit
  together) — fine as defensive machinery.

**Owner ruling on INFO-1 (meter decay during suppression):** ACCEPTED AS CANONICAL — meters PAUSE
during a recovery session (the suppression freezes meter time-decay along with production). This is
the correct reading of the recovery covenant: touch-grass costs production and time, and it must not
ALSO silently bleed Trust while you rest — a recovery that punishes recovering would be the Stardew
trap in meter form. The SB20 zero-set is formally extended to include meter time-decay; docs/soul.md
already discloses it.

**Verdict: the substrate is archival-ready EXCEPT SB24 (the heartbeat) — implement SB24 + the two
LOW fixes, then the closing designated review, then archival citing a3f7f30 + 8d2e1a6 + 203d40a +
the SB24 range.**

## 2026-08-07 — SB24 implementation recheck: blocked on SB25-SB27; adjacent fixes landed

Review by: Codex implementation recheck (self-review; not the designated gate). Recorded by: Codex.

SB24 correctly selects a heartbeat, but three executable contracts are still absent:

- **SB25 — no claim token can reach the heartbeat caller.** The exact request requires
  `{session_id,claim_token}`, but start/reconnect never issues or renews such a token. The only
  existing `claim_token` is generated inside the terminal transaction's `ClaimTx`, immediately
  consumed by `FinishTx`, and is therefore unavailable before progress. Reusing it, adding a stable
  progress capability, or changing start/reconnect are materially different security/lifecycle
  contracts. Pick one and define expiry/reclaim plus exact start/reconnect response bytes.
- **SB26 — the ordinary-command watchdog touch conflicts with the ruled lock order.** The existing
  race-safe eligibility guard runs after locking the Company stream. SB24 requires an expired
  session to execute the SB23 cancel path, which locks Founder then Company and writes both replay
  streams. Running that from the guard reverses the lock order. Rule a preflight coordinator shape
  (including the race/retry behavior when the ordinary command continues) or remove ordinary
  commands from watchdog touch ownership.
- **SB27 — the exact artifact and coordinator wire schemas are not reconciled.** SB10 fixes the Soul
  root and `policy` exact keys, while SB24 adds `recovery_beat_ceiling_ms` and
  `max_session_wall_ms` without locating them. The progress response names three carried values but
  does not enumerate the full exact receipt; watchdog adds `cancelled_by` without reconciling the
  terminal receipt schema. Locate the catalog fields and enumerate exact start/progress/watchdog
  response keys plus migration defaults/backfill rules.

Writing past these would invent a public capability, a cross-stream lock protocol, and immutable
catalog/wire bytes. The fixture-only containment therefore remains in force.

The implementable review findings are closed in this range: the suppression zero-output assertion
now snapshots every restored output authority before evaluation and compares after restoration; a
shared Go/TypeScript mutation rejects the MaxExactInteger nonterminal-band overlap; and an additional
self-review found that the live coordinator stored its rich terminal API receipt in `run_log` while
`ApplySuppressedLogged` reproduced the Company suppression receipt. The persistence coordinator now
stores a distinct replay-owned Company receipt, with a real-Postgres regression that applies the
stored payload/inputs and compares the persisted receipt. Kernel version advances to 0.3.74.

## 2026-08-07 — owner rulings SB25-SB27 (the heartbeat executable layer)
- SB25: distinct progress_token issued at start; reconnect = start-on-active returns the session with
  a ROTATED token (stale instances reject not_eligible/recovery_token); no independent expiry (the
  watchdog bounds the session); terminal claim token untouched.
- SB26: watchdog = coordinator-preflight ONLY (start/progress/resolve/cancel — outside Company locks,
  Founder-then-Company per SB23); ordinary commands stay reject-only + a read-only session_expired
  hint detail; race with a concurrent cancel resolves naturally (guard finds no session).
- SB27: beat-ceiling + max-wall join the soul artifact policy exact keys (schema_version stays 1 —
  legal only because no epoch ever pinned a soul artifact; reasoning recorded); exact start/reconnect
  + progress response keys; cancel receipts gain required cancelled_by (absent on resolve); NEW
  append-only migration, no backfill (no production rows).
Codex unblocked to implement SB24-SB27 + the closing designated review.

## 2026-08-07 — SB24-SB27 implementation: ready for designated review

Review by: pending cross-party designated review. Recorded by: Codex.

Implemented the distinct progress capability, reconnect rotation, session-row-only attended
heartbeat, and coordinator-owned lazy watchdog. The Soul artifact now owns the beat and wall-age
ceilings with Go/TypeScript exact-key and global-catch-up validation. Append-only migration 00070
adds the capability/progress authority without fabricating a backfill. Start and progress responses
use the ruled exact keys; cancel receipts distinguish player and watchdog ownership.

The real-Postgres integration drives a stale-token rejection, duplicate beat, long-gap pause,
resumption, eligibility, terminal-only replay, read-only expired-session hint, watchdog cancellation,
and the existing nine-boundary atomic fault matrix. Heartbeats add no event or replay row. Kernel
version advances to 0.3.75. This is an implementation handoff only: it does not satisfy the
cross-party designated gate and does not authorize archival.

## 2026-08-07 — designated cross-party CLOSING verdict: heartbeat batch (e6ca030 + 3ff2082) — APPROVE; Soul ARCHIVAL-ELIGIBLE

- **Review by:** the designated Claude reviewer (independent; make verify exit 0 at 0.3.75 with
  explicit typecheck, full Docker/Postgres integration suite exit 0, targeted verbose re-run of the
  soul-recovery integration test, and a LIVE MUTATION PROBE of the LOW-1 assertion — dropping a
  restore field made the integration test fail, probe reverted, re-passed). **Recorded by:** same.

**PART A — the Active-Play archival commit 398e62f: PASS.** Cites the exact four-range union + the
consumed verdict e9ebb0c; canonical docs/active-play.md added and accurate against code; RFC/README/
planning rotations correct; the commit is docs-only.

**PART B — SB24-SB27 + both routed LOWs: all VERIFIED CORRECT.**
- e6ca030: LOW-1 is now a REAL assertion (17-authority byte snapshot before Evaluate, compared after
  restore — mutation-probe-proven); LOW-2 band parity aligned with a shared rejection vector; the
  self-caught replay-receipt bug fixed with a real-Postgres regression comparing the stored receipt
  against a fresh ApplyLogged.
- Heartbeat/token (SB24/25): coordinator-only (no parser/transport surface), server-stamped,
  session-row-only, pause-never-kill proven (1000/1000/1000/5000 sequence), beat ceiling
  loader-checked <= the global ceiling in both runtimes, token rotation on reconnect with
  constant-time stale rejection.
- Watchdog (SB26): preflight of all four coordinator commands only, outside Company locks,
  Founder-then-Company via the shared coordinator; ordinary commands reject-only + read-only
  session_expired hint; no lock-order reversal on any path.
- Wire/catalog (SB27): exact policy keys with the never-epoch-pinned justification recorded in both
  loaders; exact 7-key start / 5-key progress receipts (key-count asserted); cancelled_by required on
  cancel, absent on resolve; migration 00070 new append-only with no-backfill comment.
- Terminal totals validated (end−start == attended_progress against the sole writer; replay checks
  totals never cadence — no beat is ever a replay byte); eligibility + early-resolve-no-mutation
  proven; kernel 0.3.73→74→75 lockstep; containment holds (no soul artifact in any epoch).
- Findings: 5 INFO only (transport-deferred rejection mapping, format-precheck before constant-time
  compare, a residual SB24-bullet text tension resolved by SB26, a defensive dead branch, 122-bit
  token entropy — all recorded, none blocking).

**Range-union: COMPLETE.** {a3f7f30, 8d2e1a6, 203d40a} (substrate verdict) ∪ {e6ca030, 3ff2082}
(this verdict) = every Soul implementation commit in existence; 398e62f is docs-only; no prior
verdict cites e6ca030/3ff2082.

**Verdict: APPROVE. Soul is ARCHIVAL-ELIGIBLE citing the five-commit union. Archival is the
implementer's move (this verdict is the gate); production recovery_activities may mint only after
archival + owner content per the RFC. This closes the LAST Wave-A foundation gate.**
