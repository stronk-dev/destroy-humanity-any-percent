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

## Run pinning

`run_epochs` records Company stream/run sequence, epoch, constants hash, kernel version, build VCS
revision, and deterministic unsigned seed. Account creation pins run 1 in the same transaction as
the first save revisions. A Prestige Exit takes a shared lock on the current epoch and pins the new
run in the same transaction as `run_started`; an epoch mint cannot place the run in both epochs or
neither. Once written, a pin is immutable.

Every logged Company command requires an existing pin whose hash and kernel version match the live
save. This is fail-closed: a missing epoch is an infrastructure error, never an unranked command
that silently continues.

## Exact boards

Verified rows carry three separate structural variables: Commons, Advisor, and Glitched. “Solo” is
only the display name for Commons=false and Advisor=false. Imported Founders are rejected before a
projection claim commits, so they can never enter ranked storage.

Time queries use SQL `rank()` ordered by integer milliseconds; equal keys therefore produce
standard competition ranks such as `1, 1, 3`. Count queries rank exact integers descending. Both
use `(key, run_id)` keyset cursors and remain ordinary queries when an epoch is closed, so historical
boards freeze without a special mutable mode. A partial unique index permits exactly one
world-first per category/variable/epoch tuple; the projection first attempts the world-first insert
and falls back to a normal immutable row on conflict.

## Incomplete replay boundary

The live run log and identity storage exist, but replay verification and archive compaction are not
yet claimed as implemented. The accepted RFC's log records canonical commands and receipts but
does not preserve an immutable initial run state. That state cannot always be reconstructed from
catalog initials because later runs include Founder-carried Network and Reputation effects. The
planning record carries this as a DESIGN-GAP requiring an executable contract before the verifier
or player validator can honestly ship.
