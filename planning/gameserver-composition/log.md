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
