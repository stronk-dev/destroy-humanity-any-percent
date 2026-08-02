# Gameserver Composition — append-only log

## 2026-08-02 — implementation started

Review by: not applicable (owner acceptance). Recorded by: Codex.

The owner assigned the complete RFC after ratifying Wire v2 and providing GC1-GC3. The work is
accepted as one implementation round. Composition must use real owner-backed services; a missing
callable owner surface is added to that owner package with proof, never hidden behind a no-op
driver. No push is authorized.

## 2026-08-02 — GC1 implemented

Review by: Codex implementer self-review. Recorded by: Codex.

The production resolver now receives the exact Company stream, Founder, pinned constants hash,
and request context. The Commons projection returns the current run's committed sample; before
that sample it applies the published integer entry-weight formula to the projected membership's
declared tithe. `commons_member_samples.run_seq` prevents a prior run's sample from leaking into a
fresh membership. The production integration now uses this real projector, proves first accrual,
then replays that intent byte-identically. Resolver failures wrap `ErrInvalidEngineState` and the
HTTP boundary exposes the typed `internal_invariant` category. Focused Go and real-Postgres suites
passed. Kernel version 0.3.2 records the live resolved-input semantic change.

## 2026-08-02 — GC2 implemented

Review by: Codex implementer self-review. Recorded by: Codex.

Go and TypeScript now validate the exact world state recursively, including safe integer/ppm
bounds, revision binding, epoch identity, and the null-milestone/zero-progress biconditional. The
shared wire corpus carries the canonical snapshot plus unknown-field and revision negatives. The
gameserver world aggregator owns a serialized revision and advances it only after publish success;
a 1,000-snapshot soak proves strict monotonicity and the declared zero values for unshipped planet
and milestone systems. Transport, gameserver, all 6,494 active client tests, and TypeScript/Svelte
checks passed.

## 2026-08-02 — GC3 implementation landed

Review by: Codex implementer self-review. Recorded by: Codex.

`cmd/gameserver` now assembles the epoch synchronizer and immutable database catalog bundles before
readiness, then the save/Production/Account graph, Commons/Route/Guild projections, real Guild
settlement owner, Centrifuge node, player and Guild relays, replay verifier/board projector, world
aggregator, clearing/sweep/session-GC jobs, and bounded drain. Guild and Commons subscription
authorization read their owner tables; Match remains explicitly deny-closed.

The clearing driver exposed one pre-existing contradiction: the canonical Guild docs required
missing-faction members to remain inert, while `guild.Clear` rejected the all-zero neutral member
shape. The kernel now accepts only that complete neutral shape (partial faction state still fails),
with a regression test, generated-formula fingerprint refresh, and kernel version 0.3.3.

The composed integration test boots the actual graph against Postgres, creates an account/session,
submits play, consumes receipt and event envelopes over authenticated sockets, observes the world
stream, proves Guild allow/Match deny authorization, commits a Guild clearing boundary, proves the
next Production intent consumes the real pending batch, collects expired session credentials, and
drains to non-readiness. The complete containerized integration suite and focused unit/build gates
passed. A second composed test seeds only the not-yet-shipped T0→T2 content state, then proves the
actual Production play→Exit→background verification→archive→`any_percent` board path. Canonical
account, Guild, transport, README, and gameserver docs describe the shipped graph.

## 2026-08-02 — designated review of GC3 landing

Review by: Darwin (independent reviewer). Recorded by: Codex.

Exact reviewed range: `007f1f6..e01e3ec`. Verdict: not archivable. The real settlement wiring,
epoch/catalog startup order, Account/Guild authorization, Match denial, session-family GC locking,
world schema, kernel version, formula artifact, and diff hygiene held. Six blockers were found:
the clearing query permanently starved Guilds after its first 64; version-1 saves were restored
without their migration timestamp; historical saves were interpreted with the current Faction
catalog/cap; readiness preceded an initial worker/world/relay acknowledgement; Drain waited on jobs
while intent admission remained open; and normal membership churn killed the entire worker set.
The review also identified a missing `owner_kind='founder'` join constraint, missing lifecycle-error
ownership, and acceptance-proof debt around attached ticks, Commons authorization, and command
shutdown selection.

## 2026-08-02 — GC3 review remediation implemented

Review by: Codex implementer self-review. Recorded by: Codex.

Clearing now keyset-pages every active Guild, resolves each save's pinned Economy and Faction
catalog, supplies the revision `created_at` as the v1 migration baseline, and constrains the stream
identity completely. Membership changes have a typed retry path; three consecutive changes defer
only that Guild until the next scheduled pass. Different pinned stock caps within one Guild fail
closed under an explicit `DESIGN-GAP` because no cap-changing epoch migration policy exists.

Readiness now waits for one successful pass from every job plus an initial relay flush. Runtime job
errors and unexpected clean exits lower readiness and reach `cmd/gameserver`; parent cancellation
also cannot leave readiness stale. Drain broadcasts and closes admission before waiting on jobs.
The Postgres fixture now proves Commons cohort authorization, >1-page clearing, a v1 historical
member with a distinct pinned Faction cap, and the first real Compact settlement event. A second
fixture proves the attached clearing and session-GC jobs run through the startup barrier, while
unit tests cover scheduled callbacks and command stop selection. The Guild retry semantic is a
kernel-affecting change and is recorded by version `0.3.4`.

## 2026-08-02 — designated remediation re-review, first pass

Review by: Darwin (independent reviewer). Recorded by: Codex.

Exact reviewed range in progress: `e01e3ec..fed44f1`. Two remaining blockers were demonstrated.
A 500-run stress test made `/readyz` return 204 after a fatal worker exit because the failure path
and `markReadyIfRunning` did not serialize their health transition. The disband retry handled an
empty member read, but not a Guild disappearing after that read and before the commit-side lock.

## 2026-08-02 — remediation re-review blockers closed

Review by: Codex implementer self-review. Recorded by: Codex.

Worker failure, parent shutdown, drain, and readiness recovery now share the server state lock; the
reviewer's exact 500-run stress command is green. A Guild that disbands between snapshot and
commit returns the same typed membership-change sentinel and is deferred without killing sibling
workers, with a real-Postgres regression staged after a valid member snapshot. Because this changes
the guarded Guild boundary semantic, kernel version `0.3.5` records it. Awaiting the designated
reviewer's final verdict over the extended remediation range.

## 2026-08-02 — designated remediation re-review, settlement findings

Review by: Darwin (independent reviewer). Recorded by: Codex.

Exact reviewed range: `e01e3ec..450badc`. Verdict: not archivable. Every blocker from the original
GC3 review was verified closed, including readiness serialization, disband churn, startup/drain,
historical catalog resolution, pagination, and the composed acceptance proofs. Two settlement
identity seams remained: clearing results were account-keyed while faction stock and watermarks are
run-scoped, allowing an Exit or New Founder to reapply an earlier run's result; and later clearing
snapshots ignored committed-but-unapplied results, allowing repeated scheduler ticks to reserve the
same offline stock until application failed. Both findings are systemic and block archival.

## 2026-08-02 — settlement identity remediation implemented

Review by: Codex implementer self-review. Recorded by: Codex.

New clearing results bind the authoritative snapshot's Founder, Company stream, and run alongside
the account. Pending lookup requires exact identity; legacy unattributed results stay unclaimable.
The driver's single member query nets exact-run committed-but-unapplied debits and credits before
the pure clearing kernel runs, preventing repeated boundaries from over-reserving stock or
headroom. Exit carries the Guild watermark across the same stream; New Founder establishes a
forward no-effects baseline and cannot claim prior-identity rows. The Go/TypeScript replay boundary,
shared sequential fixture, direct unit tests, and composed real-Postgres fixture cover the lifecycle
rules. Kernel `0.3.6` records the changed deterministic transition. Awaiting designated review of
the remediation commit range.

## 2026-08-02 — designated settlement remediation review and archival

Review by: Darwin (independent reviewer). Recorded by: Codex.

Exact reviewed range: `450badc..be0449d` (with the cumulative GC3 seam checked through
`e01e3ec..be0449d`). Verdict: approved; no archival blocker. Exact Founder/Company/run ownership
closes cross-lifecycle application, and the authoritative member snapshot subtracts every
committed-but-unapplied debit and adds every committed-but-unapplied credit for that identity.
The composed Postgres fixture proves three serial boundaries cannot reserve more than the
producer owns, Exit carries the watermark in both kernels, and New Founder advances through a
no-effects baseline without claiming old rows. Migration 00044, kernel 0.3.6, formula artifact,
and replay corpus are current.

The reviewer ran focused Go packages under the race detector, the real-Postgres integration
suite, client typecheck/build with 6,494 executed tests, replay/kernel/formula guards, and
`git diff --check`; all passed. Codex's complete post-commit `make verify` additionally passed
19,491 browser tests and the deterministic harness. One non-blocking debt remains: add an
outcome-sensitive near-cap reserved-credit fixture before enabling a nonzero stock-consumption
modifier or a broader clearing-worker topology. The current Phase-0 modifier is zero and the
single composed scheduler is serial.

Gameserver Composition is implemented. The active RFC and planning record rotate to their
archives; `docs/gameserver.md` and `docs/guilds.md` are canonical.

## 2026-08-02 — designated independent review of the FULL span (547e6ac..70b5747)

Review by: the project's designated Claude reviewer. Recorded by: same.

**Verdict: substantively sound; the archive KEEPS — with one procedural correction recorded and
four findings queued.** Independent verification: all 59 files diffed, every AC path traced,
full real-Postgres suite re-run green, guards executed at HEAD. GC1's live=entry-formula holds
BY CONSTRUCTION (the sampled event echoes the resolver); GC2 is byte-exact in both runtimes with
the monotonic soak and zero-value honesty enforced; the F4 seam is REAL at last
(WithGuildSettlements(guildService), no stub in non-test code, empty-vs-batch proven composed);
all six drivers attached; deny-closed match resolver socket-proven; drain order per T6; kernel
0.3.2→0.3.6 bumps all justified (0.3.5 the thinnest, over-bumping is the safe direction);
migrations appended, never edited; every checkbox flip carries its test; the archival entry cites
its verdict and range per the provenance rule.

**The procedural correction (F1):** the cited review ranges union to 007f1f6..be0449d — the GC1
(637b0d5) and GC2 (007f1f6) diffs fall OUTSIDE every recorded independent range and carried
self-review only. This review found two latent defects in exactly that uncovered slice —
coverage gaps find bugs, every time. The archival gate rule gains the corollary: **the cited
ranges must UNION to the full span being archived.**

Findings queued (none archival-blocking under the project severity bar):

1. **MEDIUM — migration 00043's backfill can stamp a PRIOR run's sample as current-run
   authoritative** (sample rows persist across Exit; backfill labels them with the current
   membership's run_seq — a re-signed member's first accrual would use the stale weight).
   Unpublished repo, no real databases — fix by follow-up migration (null out backfilled
   run_seq where the sample predates the membership's signed_at) BEFORE any real history
   migrates.
2. **LOW→ruling — member-with-absent-projection is a SERVER inconsistency**: reclassify from
   invalid_intent (400, client-fault) to ErrInvalidEngineState → internal_invariant, matching
   docs/commons.md's existing claim; the crash-window liveness gap (client must retry the SAME
   intent id to self-heal) gets one documented sentence.
3. **LOW→ruling — world_rev is per-process-lifetime**: documented as such in docs/transport.md +
   gameserver.md ("clients must not compare world rev across reconnects"); a persisted counter
   is deliberately NOT required while history is latest-only and single-node.
4. **LOW→ruling — same-guild rejoin strands unclaimable reservations**: ruled — the clearing
   tick RELEASES reservations whose committed_at precedes the member's current joined_at
   (deterministic, under-allocation-safe, bounded by Exit anyway); regression: leave/rejoin/tick
   fixture asserting released stock.
5. Recorded: the F8 mixed-cap DESIGN-GAP's whole-server blast radius (accepted fail-closed,
   carried); F6 cross-statement sampling skew (acceptable under the schema's no-cross-field-
   consistency promise); F7's composed-fixture accounting (honest in docs; cmd main() stays
   unit-covered — acceptable, noted).

## 2026-08-02 — full-span review findings implemented

Review by: Codex implementer self-review. Recorded by: Codex.

The four queued findings from the designated full-span review are fixed forward without changing
the archived RFC. Migration 00045 makes `commons_member_samples.run_seq` nullable and clears only
labels whose sample predates the current membership's `updated_at`; a transaction-isolated real
Postgres test applies the migration body to stale and current samples and distinguishes both.
The live and Exit Commons paths now share one required-weight resolver, and an absent projection
for a save-declared member is `ErrInvalidEngineState`/`internal_invariant`, never client fault.
Canonical Commons docs name the same-intent retry that replays the committed event after a
projection crash window.

Clearing reservations now require `committed_at >= current joined_at`. The composed integration
fixture creates outstanding producer debits, leaves and rejoins the same Guild, proves the full
stock is available, then commits the next clearing tick. World revision is documented in both
transport and gameserver truth as monotonic only within one process lifetime. `AGENTS.md` now
requires archival review ranges to union to the full implementation span. Kernel `0.3.7` records
the server-owned replay-input/error semantic change.

Focused Go packages and the complete real-Postgres integration suite passed. Awaiting the full
repository gates and designated independent review of the exact follow-up commit range.

## 2026-08-02 — designated review of full-span finding remediation

Review by: Darwin (independent reviewer). Recorded by: Codex.

Exact reviewed range: `70b5747..65a0371`. Verdict: not approved; three MEDIUM findings block
closure and one LOW weakens the local proof. Ordinary and Exit post-commit projector failures were
returned raw instead of the public `internal_invariant` classification. Migration 00045's Down
path recreated the stale sample attribution that Up invalidated. Reservation ownership used
inclusive millisecond timestamps, which cannot distinguish an abandoned membership from a rejoin
at the same canonical instant. The migration proof's test name also omitted `Integration`, so the
standard local real-Postgres target skipped it. World-revision documentation, the full-span-union
rule, migration Up behavior, and source-of-truth kernel parity were approved.

## 2026-08-02 — designated-review blockers remediated

Review by: Codex implementer self-review. Recorded by: Codex.

Both post-commit projection paths now use one wrapper that classifies projector failure as
`ErrInvalidEngineState`; ordinary and Exit integration fixtures prove that the first response is
server-owned and a same-intent retry replays successfully. Migration 00045 now fails closed on
rollback when invalidated rows exist, its real-Postgres proof is selected by the standard
`Integration` target, and the fixture asserts both Up and the refusing Down path.

Timestamp ownership is removed rather than patched. Migration 00046 gives every new clearing
result the immutable Guild `membership_id` captured in its snapshot. Commit-side validation locks
and compares account-plus-membership identities, reservation and settlement reads require the
exact membership/Founder/Company/run tuple, and legacy unattributed rows remain unclaimable. One
real-Postgres fixture makes boundary commit, leave, and rejoin share the exact same millisecond and
proves both reservation release and settlement exclusion. Kernel `0.3.8` records these server and
replay-input semantics. Awaiting the full standard gates and designated review of the extended
range.

The focused Go packages, complete real-Postgres `Integration` target, formula regeneration/drift
check, deterministic balance harness, 6,494 client tests, 19,491 browser tests, schema guards,
typecheck, and build all passed through the standard root Make targets at kernel `0.3.8`.

During designated re-review Darwin identified that direct SQL execution did not prove Goose could
parse migration 00045's procedural Down guard, and that 00046's ordinary Down would discard known
membership-period ownership after the first attributed boundary. Both Down paths now use explicit
Goose statement boundaries and fail closed when rollback would destroy identity. Two isolated
real-Postgres fixtures drive the actual Goose provider in both directions: Up must apply, and Down
must refuse after the respective identity-bearing row exists.
The same-millisecond Guild fixture also submits the abandoned M1 snapshot after M2 is active and
asserts the commit-side `(account_id,membership_id)` lock comparison returns the retryable
membership-churn sentinel; the validation is now both implemented and directly proven.
