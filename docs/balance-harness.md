# Balance Harness

The balance harness is the deterministic, headless consumer of the authoritative production
transition. It runs in memory without HTTP, WebSockets, Postgres, sleeping, or per-player tick
loops. Both the persisted service and harness call `production.Transition`; policies submit parsed
intents and never mutate state directly.

## Phase-0 scenario

`testdata/harness/scenarios/phase0-production.json` is scenario version 4 (the economy + Routes + Commons
balance bundle) and runs two versioned policies:

- `casual.phase0` v1: three eight-minute daily sessions at 09:00, 14:00, and 20:00 UTC, acting
  once per second; it buys the cheapest affordable generator or performs one manual action.
- `chaos.phase0` v1: continuously online, acting every five virtual minutes with SplitMix64-selected
  valid intent templates and counts.

Virtual time begins at `2000-01-01T09:00:00Z`. SplitMix64, rejection-sampled bounded draws, and a
separate deterministic UUIDv7 stream make every run a pure function of its complete run key. Runs
are parallelized into fixed report slots and reduced by complete key order. Reports contain only
integers and canonical Decimal strings—never binary JSON floats.

Each scenario names its economy, Routes, and Commons catalogs. `constants_hash` is computed from the
named, length-framed bundle of their exact bytes, so either balance input changes run identity while
artifact enumeration order does not. Runtime scenario loading rejects unknown milestone kinds;
schema validation is not the sole closed-registry gate.

The blocking suite contains 200 Chaos seeds over 30 virtual days and 100 Casual seeds over 21
days. Required milestones cover a manual action, first generator purchase, generator count one,
and complete T0 progress. Every declared policy/milestone pair contributes p50 and p95 observations
to the aggregate and drift baseline: two policies × four milestones × two statistics = 16 unique
values. The bounded Casual p50 first-purchase and p95 first-generator observations are both
10,000 ms, inside their 60-second and five-minute envelopes; unbounded observations detect drift
without pretending to be new pacing targets. Runtime and schema semantics reject unknown
references, duplicate tuples, and incomplete observation matrices.

## Reports and invariants

Every run records ordered milestone times, applied/rejected counts, source/sink totals, final
balances and state hash, maximum progression gap, and invariant failures. The v1 registry requires:
state encoding, numeric domain, resource bounds, receipt before/after-chain reconciliation,
monotone revisions, and every `must_reach` milestone.
Aggregate invariant diagnostics include the complete run key rather than only policy and seed, so
a failure remains attributable across scenario, policy, and constants revisions.

Receipt `before` and `after` values are the exact reconciliation authority. A signed 12-digit
`delta` remains useful for source/sink aggregates but cannot always re-add exactly after lossy
cancellation; some valid before/after pairs have no representable 12-digit delta. The harness
therefore checks the complete canonical state chain without weakening numeric precision rules.

## Relevance instrument

The relevance harness measures whether each generator or upgrade has a material effect inside its
declared progression window. It drives the same production simulation boundary as the service and
uses a guarded action-free `SimulateAdvance` seam for deterministic waiting; it never copies
production arithmetic. Its hash-pinned `relevance_policy` artifact owns windows, epsilon floors,
trap exemptions, and tier/category/declared groups without adding harness metadata to the economy
catalog.

For every declared run it records an unmasked baseline, per-item effect ablations, and group effect
ablations. Reference runs additionally record action-removal diagnostics and one width-eight beam
oracle. The reference policy ranks one-unit generator buys and unowned upgrades by an exact
single-resource payback calculation, including lower-bound searches for first affordability and
first positive marginal output. Beam nodes deduplicate by canonical state plus virtual time and use
componentwise dominance before the width bound. Both run cardinality and transition work are
checked before simulation, and every simulation call spends from the transition budget.

Each policy item's availability window must contain at least one scenario segment, and that segment's
milestone is the only evidence eligible for the item's verdict. The report contains ordered item,
group, tier-contribution, role-activation, and failure rows with
only safe integers, booleans, and canonical hashes. Required baselines that do not reach their
milestone are named failures; unreachable ablations use the finite horizon-minus-baseline encoding.
An item passes through its own effect delta or one declared supporting group, while the trap test
always remains individual. Role evidence comes only from unmasked baseline execution.

`testdata/harness/relevance/registry-v1.json` is the fail-closed scenario authority. Every registered
golden is discovered dynamically by the history and TypeScript schema gates. Each trap exemption's
mechanical justification key must appear in the entry's reviewable changelog. The current
schema-v4 entry is deliberately a test fixture: it proves a dead upgrade, a deliberate trap
exemption, a roleless generator, group-supported substitution, and a greedy/beam gap under its
declared bound. Fixture findings are recorded in its golden report but do not block the production
gate. When the active epoch first uses an economy schema v4 or later, its registry entry must bind
the exact epoch-owned economy, Routes, and relevance-policy artifacts, use the current epoch
changelog, and emit the accepted full epoch-bundle constants hash. `harness-check` executes the
entry and fails on any relevance finding. The current schema-v3 production catalog is not presented
as a relevance baseline.

## Commands and drift

- `make harness HARNESS_OUTPUT=/absolute/path/report.json` writes the complete canonical run and
  aggregate report.
- `make harness-check` is read-only and compares the seed-0 golden report, pacing baseline, and
  every registered relevance golden report.
- `make harness-update` deliberately regenerates those tracked artifacts for review.

Drift above 10% warns and above 25% fails using integer cross-multiplication. After the initial
baseline, economy/Commons/relevance catalog or scenario inputs land first. A separate commit whose
subject begins `BALANCE-CHANGE:` then contains only the generated pacing, seed, and/or relevance
artifact. The repository guard scans every reachable baseline revision, not only HEAD, and fails
on shallow history, uncommitted artifacts, missing prior inputs, wrong subjects, or any unrelated
path in the artifact commit. CI fetches complete history so local and hosted enforcement are
identical. The schema gate rejects unknown scenario fields/kinds and non-string or unsafe seed
encodings. Go and TypeScript load the same mutation corpus, including semantic JSON integers written
with a decimal lexical form.
