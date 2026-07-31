# Leaderboards and Balance Epochs

The leaderboard foundation stores exact run commands, pins every run to immutable balance bytes,
and derives board rows from verified-run facts. No API accepts a Decimal ranking key: time boards
use integer milliseconds and count boards use exact integers.

## Kernel and replay identity

[`kernel/VERSION`](../kernel/VERSION) is the transition-semantics identity shared by Go and the
TypeScript client. Generated constants are embedded in both builds, and `make verify-client`
fails if either differs from the source file. Server builds also record their VCS revision as
forensic detail; unknown local revisions remain explicit rather than being fabricated.

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
and artifacts not executed by the harness (currently Prestige and factions) still participate in
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

Every logged Company command requires an existing pin whose hash matches the live save. A missing
pin or hash mismatch is an infrastructure error. A kernel-version mismatch does not strand the
run: the command commits and the transaction inserts one immutable `run_version_drift` row for the
run. Drifted runs verify as `engine_mismatch` once replay verification lands and are rejected by
board projection today. Thus playability survives a deployment while ranked integrity remains
fail-closed.

At Exit, the ended run retains its original pin. Run N+1 is assembled under the server's current
catalog hash and pinned to the epoch current in that same transaction. A real-Postgres fixture
mints changed artifact bytes between start and Exit and asserts both hashes, epochs, revisions,
events, and the next run's continued play.

## Exact boards

Verified rows carry four separate structural variables: Commons, Advisor, Glitched, and nullable
Faction. “Solo” is only the display name for Commons=false and Advisor=false. Imported Founders are rejected before a
projection claim commits, so they can never enter ranked storage.
Runs recorded in `run_version_drift` are rejected at the same projection boundary.

Time queries use SQL `rank()` ordered by integer milliseconds; equal keys therefore produce
standard competition ranks such as `1, 1, 3`. Count queries rank exact integers descending. Both
use `(key, run_id)` keyset cursors and remain ordinary queries when an epoch is closed, so historical
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
production kernel.

The Go live service now executes through the shared `ApplyLogged` transition and compares cleanly
when the persisted input is replayed from the pre-command state, including terminal receipt and
event order. The TypeScript port, genesis storage, full-run verifier, and archive compaction are not
yet claimed as implemented. Catalog initials still cannot reconstruct later runs because
Founder-carried effects alter their starting state; the active Run Genesis RFC owns that remaining
work in that order.
