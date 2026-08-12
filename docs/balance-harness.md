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
ablations. Reference runs additionally record action-removal diagnostics and one declared-width beam
oracle. The reference policy ranks one-unit generator buys and unowned upgrades by an exact
single-resource payback calculation, including lower-bound searches for first affordability and
first positive marginal output. Before the ablation matrix, one opportunity screen compares the
direct-payback trajectory with each purchase removed and monotonically removes a purchase when
that counterfactual reaches the milestone earlier. This is a whole-item backward-elimination
screen; it does not change the payback formula or comparator. The same immutable screen and
unchanged payback ordering govern the reference arm, every reference ablation, and every beam
rollout. At
run genesis it may bootstrap the cheapest milestone-resource
purchase through the pinned manual action: the requested count is derived from the real quote and
catalog yield, capped by the catalog's click bucket, and its window is derived from the catalog
refill rate. Relevance scenarios declare `beam_children >= 2`; each beam node orders candidates by
the same payback metric as the reference arm and then raw-byte ID, selects at most that many, and only then performs
the expensive R10 scoring and greedy completion. Beam nodes deduplicate by canonical state plus
virtual time before greedy rollout and use componentwise dominance before the width bound. A beam
path and its greedy completion share one `max_decisions` envelope: rollout receives only the
decisions the path has not consumed. Run cardinality is checked before simulation. The static
transition estimate is only a generous runaway-configuration guard, compared with the scenario's
declarative `preflight_ceiling`; it is never treated as measured work. The scenario's separate work
budget is enforced by one counter spanning every simulation call, including the reachability
preflight, and aborting without a partial report when the limit is reached. Before the factorial
ablation matrix begins, one reference-arm probe fails with `reference_decision_starved` when the
declared decision guard fires before the milestone, and with `milestone_unreachable:<id>` when the
target cannot be reached within the horizon for another reason. Guard exhaustion never silently
coasts to a milestone. A failing run removes the authoritative output and writes a
separate `*.diagnostic.json` envelope marked `authoritative: false`; diagnostics are never golden
inputs and a later passing run removes any stale diagnostic.

Schema v2 permits an availability/segment `from_gate` of null, meaning run genesis and sorting
before every declared gate; schema-v1 bytes retain their concrete-gate requirement. Each policy
item's availability window must contain at least one scenario segment, and that segment's
milestone is the only evidence eligible for the item's verdict. The report contains ordered item,
group, tier-contribution, role-activation, and failure rows with
only safe integers, booleans, and canonical hashes. Required baselines that do not reach their
milestone are named failures; unreachable ablations use the finite horizon-minus-baseline encoding.
An item passes through its own effect delta or one declared supporting group, while the trap test
always remains individual. A seed whose unmasked baseline buys none of the item contributes no
ablation signal. A persona is excluded only when every one of its seeds buys zero; report schema v3
records every such persona in the affected item or group row's `excluded_persona_ids`. Report
schema v4 additionally records the screen's complete sorted `instrument_excluded_ids` and marks
matching item rows `instrument_affected`; their floor findings use the same prefix so they cannot
be mistaken for independent content verdicts. Any positive seed remains binding, and baseline
purchase counts reduce over positive seeds before taking the maximum across personas. Role
evidence comes only from unmasked baseline execution.

Production candidates use phase-scoped relevance policies. The T0 scenario measures rows whose
windows close at `gate.t0_to_t1`; the cumulative T1 optimizer continues through `gate.t2_to_t3`,
but its report and ablation budget include only rows in the T1 window that closes at that target.
Earlier content remains available to the optimizer without earning evidence after its own window.
A scoped policy must exactly match the cumulative catalog rows needed to reach its milestone; a
row is judged only in the scenario whose target closes its declared window. Content that opens at
the terminal coordinate belongs to a later reachable scenario. `generator.legal_dept` currently
has no such later scenario and is named coverage debt rather than silently exempted.

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
- `make first-content-harness` validates the ratified first-content manifest and its complete
  16-artifact replay bundle, then writes the owner-facing candidate-versus-baseline pacing report.
  Pacing drift is reported rather than vetoed; deterministic invariant failures still fail the run.
- `make t0-t1-relevance-all` runs the phase-scoped T0 and cumulative T1 candidate gates; either
  target can be run alone with `make t0-t1-relevance` or `make t1-relevance`.
- `make harness-update` deliberately regenerates those tracked artifacts for review.

The successor `content_dynamics.v1` lane has an intentionally empty production registry until an
epoch pins an Opportunities artifact and its economy declarations. `make content-harness`
generates any newly registered current-epoch bundle snapshot; with the empty registry it is an
honest no-op. Snapshots are immutable and content-addressed under
`testdata/harness/content-dynamics/bundles/<full-hash>/`: their manifest records the complete
sorted artifact set, production and snapshot paths, per-file SHA-256, epoch coordinate, and full
bundle hash. The read-only `harness-check` rejects missing, extra, tampered, subsetted, rehashed,
or epoch-unaccepted snapshot bytes. Later epochs add snapshots rather than resolving old entries
through mutable production paths.

Drift above 10% warns and above 25% fails using integer cross-multiplication. After the initial
baseline, economy/Commons/relevance catalog or scenario inputs land first. A separate commit whose
subject begins `BALANCE-CHANGE:` then contains only the generated pacing, seed, and/or relevance
artifact. The repository guard scans every reachable baseline revision, not only HEAD, and fails
on shallow history, uncommitted artifacts, missing prior inputs, wrong subjects, or any unrelated
path in the artifact commit. CI fetches complete history so local and hosted enforcement are
identical. If a mixed artifact commit is discovered only after publication, it cannot be rewritten:
the sole forward remedy is an append-only entry in
`kernel/baseline-history-corrections.json`. Each entry names the exact immutable offending commit,
its correction class, rationale, and planning log. Commits changing that registry may change no
other path, entries may only be appended unchanged, and corrections forgive only the named mixed-
artifact packaging violation—not future commits or other guard failures. The guard also proves the
offending commit is an ancestor of the configured `refs/remotes/origin/main`; a missing remote-
tracking ref is a fail-closed CI configuration error, while a named unpublished commit is rejected
with the instruction to repackage it instead of spending an amnesty entry. The schema gate rejects
unknown scenario fields/kinds and non-string or unsafe seed
encodings. Go and TypeScript load the same mutation corpus, including semantic JSON integers written
with a decimal lexical form.
