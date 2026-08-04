# Leaderboards and Balance Epochs

The leaderboard foundation stores exact run commands, pins every run to immutable balance bytes,
and derives board rows from verified-run facts. No API accepts a Decimal ranking key: time boards
use integer milliseconds and count boards use exact integers.

## Kernel and replay identity

[`kernel/VERSION`](../kernel/VERSION) is the transition-semantics identity shared by Go and the
TypeScript client. Generated constants are embedded in both builds, and `make verify-client`
fails if either differs from the source file. Server builds also record their VCS revision as
forensic detail; unknown local revisions remain explicit rather than being fabricated.
The version gate also walks the repository history from the introduction of
[`kernel/affecting-paths.json`](../kernel/affecting-paths.json): any registered non-test transition,
receipt, event, snapshot, or state-encoding path changing without the same commit bumping
`kernel/VERSION` fails. This makes version drift detectable by construction rather than review
memory.

Catalog bytes are immutable rows keyed by `constants_hash` and artifact name. Database triggers
reject update/delete attempts. An epoch is a deliberate interval with a monotonic id, start/end,
repository changelog reference, and append-only accepted hash set. Minting closes the current epoch
and creates the next epoch plus its artifacts in one transaction. A correctness-only hotfix can
append a hash to the current set but cannot replace historical bytes.

Epoch rows have a dedicated history guard. The sole legal update changes the current epoch's
`ended_at` from null to its closing instant while leaving every other column byte-identical;
subsequent updates and every delete fail. This allows minting the successor without leaving a
metadata rewrite path through the `epochs` table itself.

[`balance/epochs/phase0.json`](../balance/epochs/phase0.json) is the repository seed for that same
identity. It names the exact artifact bundle and the hashes accepted by each epoch. The harness
and runtime load that declaration through one strict `epochseed` authority; no composition site
owns a parallel list. Scenario-owned economy/routes/commons paths must equal their manifest entries,
and artifacts not executed by the harness (currently Prestige, factions, guilds, and categories) still participate in
its identity.
The history gate walks every commit after the seed: a correctness-only artifact change must append its
resulting hash to the current epoch in the same commit, while a `BALANCE-CHANGE:` must append one
new epoch and its numbered changelog. Hardcap reductions can never be hotfixes and require an
explicit `Cap migration:` policy in the new epoch changelog. `make epoch-hash` prints the exact
worktree hash to register; it does not modify the seed.

Before the realtime node starts or readiness can become true, the gameserver requires an epoch-seed
synchronizer to reconcile that exact bundle into Postgres. Reconciliation is idempotent, serializes
with mint/hotfix operations, and reconstructs every declared epoch and accepted-hash identity when
the database is empty; all historical epochs are closed and only the manifest current epoch stays
open. The current worktree bundle supplies artifact bytes for the current hash. Historical hashes
are restored as identities without fabricated bytes; replayable historical artifact bytes remain
the database backup's responsibility. An existing database must be an exact manifest prefix and may
advance only one deployed epoch. Both bootstrap and one-step advance install the complete declared
accepted-hash set, including hotfixes declared on the new epoch before its first deployment; only
the current worktree hash receives locally available artifact bytes. Epoch IDs are allocated
explicitly under the same transaction lock, so an invalid changelog reference cannot burn a
sequence number and poison every retry.

A `CONSTANTS-IDENTITY:` baseline commit is a narrow non-balance repair path: the guard proves that
only constants-hash fields changed, every new hash equals the manifest-computed value, and all pacing
and golden behavior is byte-semantically unchanged. Every manifest artifact must also be byte-for-byte
identical to its state at the previous pacing-baseline commit. The first historical identity repair
is the sole migration exception because its previous baseline predates the manifest authority. This
path can repair composition code; it cannot authorize code or tuning changes.

## Run pinning

`run_epochs` records Company stream/run sequence, epoch, constants hash, kernel version, build VCS
revision, and deterministic unsigned seed. Account creation pins run 1 in the same transaction as
the first save revisions. A Prestige Exit takes a shared lock on the current epoch and pins the new
run in the same transaction as `run_started`; an epoch mint cannot place the run in both epochs or
neither. Once written, a pin is immutable.

Every pin has exactly one `run_genesis` row containing the authoritative save version, constants
hash, and PostgreSQL-canonical bytes of the run's first revision. A deferred database trigger
rejects a transaction that commits a pin without its genesis; the genesis row rejects every update
and delete. Account creation, normalized import, and Exit-created run N+1 write the pin, first
revision, and genesis atomically. Fault injection attacks both sides of the pin/genesis boundary,
and the account/import/Exit integration matrix compares genesis with the actual first revision.

Every logged Company command requires an existing pin whose hash matches the live save. A missing
pin or hash mismatch is an infrastructure error. A kernel-version mismatch does not strand the
run: the command commits and the transaction inserts one immutable `run_version_drift` row for the
run. Drifted runs verify as `engine_mismatch` and are rejected by board projection. Thus
playability survives a deployment while ranked integrity remains fail-closed.

The shared replay verifier exists in Go and TypeScript and returns only the closed verdict union:
`verified`, `log_gap`, `state_divergence`, `constants_mismatch`, `clock_violation`, or
`engine_mismatch`. It restores immutable genesis, replays each sequence through the same
`ApplyLogged`/`ApplyLoggedExit` boundary used live, and compares canonical receipt bytes plus the
ordered event envelope bytes. A Go-authored 51-entry mixed run—manual accrual, a purchase,
Compact join/leave, incorporation, a gate-generated offer, decline, a long transition sequence,
and terminal Exit—must verify identically in both runtimes. The same corpus mutates one dimension
at a time to assert every failure verdict. Its successful path also returns the independently
replayed terminal Company state to the parity suites; both compare it with the shared
`final_state_json`, so terminal facts cannot agree while hidden state silently diverges.
The player validator consumes the archived genesis save version; its explicit v12→v13 migration
backfills the purchase accumulator exactly as Go does, while a mislabeled envelope fails closed.
Both runtimes reject pre-v12 genesis (none was ever produced), require the accumulator field on a
v13 envelope, and test non-zero plus exact-domain-saturated v12 backfills.

Ended runs enter `verification_queue` in the Exit transaction. Workers claim only the oldest
unfinished run per Company, using expiring UUID claim tokens; projection and the token-checked
terminal mark share one transaction. A kernel that does not match the pinned version defers the
row without spending an attempt. Database/scan failures retry and, after five attempts, enter a
separate immutable operational-poison ledger plus `InvariantSink` report—they never masquerade as
one of the six replay verdicts. Only deterministic stored evidence creates a verdict dead letter.
Event replay is scoped to the run's Company/Founder stream pair, and each Exit log row resolves its
own next-catalog bundle. Verified and dead queue records are immutable.

At Exit, the ended run retains its original pin. Run N+1 is assembled under the server's current
catalog hash and pinned to the epoch current in that same transaction. A real-Postgres fixture
mints changed artifact bytes between start and Exit and asserts both hashes, epochs, revisions,
events, and the next run's continued play.

## Exact boards

Verified rows carry four separate structural variables: Commons, Advisor, Glitched, and nullable
Faction. “Solo” is only the display name for Commons=false and Advisor=false. Imported Founders are rejected before a
projection claim commits, so they can never enter ranked storage.
Runs recorded in `run_version_drift` are rejected at the same projection boundary.

The L7 evaluator is a strict, closed data union: `any`, `all_gates`, fact-set superset/disjoint,
an exact count ceiling, and bounded `all_of`. It rejects unknown predicate kinds, open fields,
route-gate drift, and any Phase-0 category shape other than Any%, 100%, Ethical%, Low%, and
Valuation. The Phase-0 completion set is empty, so 100% is explicitly all gates. Ethical% rejects
exact or prefix-matched facts in the registered `darkpattern.*` and `externality.*` namespaces;
neither namespace has a Phase-0 producer yet, so every otherwise eligible Phase-0 run qualifies.
The
queue-owned projector reads the sole schema-v2 `run_ended`, checks its terminal sequence against
the immutable log, derives Faction and Glitched from run-scoped events, takes Commons and Advisor
from the terminal assisted record, and projects every matching category in the queue's mark
transaction. It loads category and route bytes by the run's pinned constants hash, never from the
current worktree, so later epochs cannot reclassify historical runs. Imported and version-drift
runs claim projection with no board rows. Pre-timer runs enter only Valuation and remain
structurally excluded from every time-keyed category.

`run_ended` schema v2 carries sorted crossed gates plus an exact
`generators_purchased_total`. The latter is a save-v13 run accumulator incremented only by accepted
generator purchases; v1–v12 saves backfill it from then-owned counts. This avoids redefining Low%
when later systems provision or consume generator units.

Time queries use SQL `rank()` ordered by integer milliseconds; equal keys therefore produce
standard competition ranks such as `1, 1, 3`. Count queries rank exact integers descending. Both
use `(key, run_id)` keyset cursors. Valuation parses the canonical terminal Decimal into an exact
`(exponent, 12-digit padded quantized_mantissa)` pair and ranks both columns descending; equal pairs
share a rank, including values beyond native integer or floating-point range. Magnitude pagination
uses `(zero, exponent, mantissa, run_id)`: canonical `(0,0)` is the unique last value, below every
positive sub-unit magnitude. All queries remain ordinary when an epoch is closed, so historical
boards freeze without a special mutable mode. A partial unique index permits exactly one
world-first per category/variable/epoch tuple; the projection first attempts the world-first insert
and falls back to a normal immutable row on conflict.
The database independently constrains every `run_id` to canonical `company-uuid:positive-run-seq`
form, so direct projection/storage callers cannot insert an identity that the verifier cannot
resolve.

## Replay-input boundary

The live run log now separates the three replay facts: canonical payload is what the player said,
`replay_inputs` is what the server resolved, and receipt/events are what happened. New Company log
rows require a versioned replay-input object in the gameplay transaction. Historical rows remain
nullable and are explicitly unrankable rather than receiving invented backfill values. The command
envelope is persistence-authoritative; the per-intent resolved union is exact-key validated by the
production kernel. Active run-log rows are immutable at SQL (`UPDATE` and `DELETE` both fail), so
tampering cannot be laundered into a historical `log_gap`. On a verified verdict, the queue writes
one deterministic `gzip+json.v1` archive containing pin, genesis, every canonical command/input/
receipt, and totally ordered Company/Founder event evidence; stores its SHA-256; deletes the live
rows; and token-marks the queue in the same transaction. The delete trigger permits that one
archive-backed compaction and no update. Rollback/retry tests prove byte-identical archives and no
partial deletion.

The Go live service now executes through the shared `ApplyLogged` transition and compares cleanly
when the persisted input is replayed from the pre-command state. The TypeScript verification kernel
loads the same seven-artifact bundle and reproduces ordinary transitions plus wind-down, stored-offer,
and scripted cross-gate terminal transitions. Replay-inputs v3 carries Founder lifetime-achievement
state needed for deterministic v16 settlement; v2 remains accepted only for historical
pre-foundation evidence. The shared Go-authored corpus compares receipts,
event bytes/order, final Company state, Founder-derived output, and next-run snapshots. Immutable
genesis storage, queue failure handling, the L7 evaluator/projector, terminal-state comparison, and
archive compaction are live. The production category file is an epoch-owned constants artifact;
the category loader, projector, Go replay loader, TypeScript artifact identity, and schema gate all
consume the same pinned bytes.
