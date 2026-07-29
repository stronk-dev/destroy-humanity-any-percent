# Balance Harness

The balance harness is the deterministic, headless consumer of the authoritative production
transition. It runs in memory without HTTP, WebSockets, Postgres, sleeping, or per-player tick
loops. Both the persisted service and harness call `production.Transition`; policies submit parsed
intents and never mutate state directly.

## Phase-0 scenario

`testdata/harness/scenarios/phase0-production.json` runs two versioned policies:

- `casual.phase0` v1: three eight-minute daily sessions at 09:00, 14:00, and 20:00 UTC, acting
  once per second; it buys the cheapest affordable generator or performs one manual action.
- `chaos.phase0` v1: continuously online, acting every five virtual minutes with SplitMix64-selected
  valid intent templates and counts.

Virtual time begins at `2000-01-01T09:00:00Z`. SplitMix64, rejection-sampled bounded draws, and a
separate deterministic UUIDv7 stream make every run a pure function of its complete run key. Runs
are parallelized into fixed report slots and reduced by complete key order. Reports contain only
integers and canonical Decimal strings—never binary JSON floats.

The blocking suite contains 200 Chaos seeds over 30 virtual days and 100 Casual seeds over 21
days. Required milestones cover a manual action, first generator purchase, generator count one,
and complete T0 progress. Casual p50 first purchase and p95 first generator are both 10,000 ms,
inside their 60-second and five-minute envelopes.

## Reports and invariants

Every run records ordered milestone times, applied/rejected counts, source/sink totals, final
balances and state hash, maximum progression gap, and invariant failures. The v1 registry requires:
state encoding, numeric domain, resource bounds, receipt before/after-chain reconciliation,
monotone revisions, and every `must_reach` milestone.

Receipt `before` and `after` values are the exact reconciliation authority. A signed 12-digit
`delta` remains useful for source/sink aggregates but cannot always re-add exactly after lossy
cancellation; some valid before/after pairs have no representable 12-digit delta. The harness
therefore checks the complete canonical state chain without weakening numeric precision rules.

## Commands and drift

- `make harness HARNESS_OUTPUT=/absolute/path/report.json` writes the complete canonical run and
  aggregate report.
- `make harness-check` is read-only and compares the seed-0 golden report plus pacing baseline.
- `make harness-update` deliberately regenerates both tracked artifacts for review.

Drift above 10% warns and above 25% fails using integer cross-multiplication. After the initial
baseline, catalog/scenario inputs land first. A separate commit whose subject begins
`BALANCE-CHANGE:` then contains only the generated pacing baseline and optional golden-seed
artifact. The repository guard scans every reachable baseline revision, not only HEAD, and fails
on shallow history, uncommitted artifacts, missing prior inputs, wrong subjects, or any unrelated
path in the artifact commit. CI fetches complete history so local and hosted enforcement are
identical. The schema gate rejects unknown scenario fields/kinds and non-string or unsafe seed
encodings.
