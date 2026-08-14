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
oracle. The reference policy compares banking with every affordable one-unit generator or unowned
upgrade by exact projected time to the scenario's balance milestone. It obtains the current and
post-purchase resource rates from the production engine's masked rate assembly, then compares the
closed-form ratios without rounding: `(target - balance) * 1000 / rate`. Purchases win equal scores
and tie-break by raw-byte ID; banking advances to the milestone or next declared boundary. Before
the ablation matrix, one whole-item backward-elimination screen removes any purchase whose absence
strictly improves the projected-time reference trajectory. The same immutable screen and projected-
time metric govern reference, ablations, beam child selection, and the terminal completion used to
score each beam node. At run genesis it may bootstrap the cheapest milestone-resource
purchase through the pinned manual action: the requested count is derived from the real quote and
catalog yield, capped by the catalog's click bucket, and its window is derived from the catalog
refill rate. Relevance scenarios declare `beam_children >= 2`; each beam node orders purchase and
bank choices by the same projected-time metric, keeps at most that many, deduplicates by canonical
state plus virtual time, applies componentwise dominance, and completes survivors with the same
ranked policy before the width bound. Deterministic completion suffixes are memoized by state,
revision, virtual time, and remaining decision allowance; cached work is not counted as an executed
transition. Routine runs apply deterministic `deviation.v1` probes: each selected coordinate forces
one legal alternative and completes through the same ranked policy. Reports disclose the eligible,
selected, executed, and unprobed coordinate counts plus `maximum_forced_deviations: 1`; this is an
honest radius-1 falsifier, not an optimality proof. The expensive beam is available only through
`make relevance-beam`, is absent from repository, CI, and release gates, and treats equality with
the reference as healthy while a slower search is `greedy_oracle:beam_not_better`. Run cardinality
is checked before simulation. The static
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
matching or mechanically dependent item rows `instrument_affected`; their floor findings name the
reason and removed target (`instrument_affected:<reason>:<removed-id>:…`) so an upgrade whose sole
effect target was removed cannot masquerade as dead content. Any positive seed remains binding, and baseline
purchase counts reduce over positive seeds before taking the maximum across personas. Role
evidence comes only from unmasked baseline execution. Report schema v5 replaces the routine
`greedy_oracle` result with `deviation_oracle`; the legacy field remains null so a cheap probe cannot
be mistaken for a beam result. Report schema v6 independently plans the run count and checks it
against execution. It also partitions every forced-deviation attempt into reached, unreached, or
decision-starved counts; any starvation marks the oracle incomplete and fails the gate rather than
silently looking like a successful probe.

Internal segment boundaries are part of the measured trajectory. At the first decision boundary
where the pinned Routes requirements are reachable, every relevance arm advances to that point and
applies the real production `cross_gate` transition. The transition burns the real requirement,
advances tier, consumes counted work, and fails the run if production rejects it; the optimization
segment's terminal gate is never crossed automatically. The runner does not write `Tier` or
`GatesCrossed` directly.

Generic-persona role activations remain observational report evidence, not a content floor: stock,
manual, provision, and synergy roles require different honest execution contexts. The mandatory
`make t0-t1-role-check` gate instead enumerates every generator-role row in the pinned T0–T1 economy
candidate, exercises its minimum real production context, and pairs it with a generator-masked
control that must remove both the activation and its non-neutral effect.

Branch-specific purchasables have a separate candidate gate. `RunRelevanceBranchProofs` derives
upgrade rows the main reference does not buy and generator rows whose validated whole-path report
leaves below their floor without an instrument-affected label. It follows the shared ranked policy
from genesis through a legally reachable prefix and accepts a row only when the real engine selects
the purchasable over banking and an effect-masked completion loses at least the policy epsilon. For
a generator, the mask suppresses its production and declared roles while preserving the real
purchase and its cost in both cloned completions. The prefix removes and discloses every competing
upgrade as well as every generator beyond the target branch. The schema-v2 report distinguishes
generator and upgrade rows explicitly. A supplied whole-path report is accepted only when its
scenario, constants, and relevance-policy hashes match the loaded suite; omitting it performs the
measurement inline. `make t0-t1-branch-check` is therefore self-contained on a clean checkout;
`make t0-t1-branch-check-from-reports` is the explicit reuse path after the combined gate has just
written hash-matched diagnostics. This lane does not replace the main
unmasked reference, whole-path pacing, or the deviation oracle, and it cannot grant a trap exemption.

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
schema-v6 entry is deliberately a test fixture: it proves a dead upgrade, a deliberate trap
exemption, observational role evidence, group-supported substitution, and a registered forced-deviation
counterexample under its declared bound. Fixture findings are recorded in its golden report but do not block the production
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
- `make relevance-beam RELEVANCE_SCENARIO=…` runs the expensive beam explicitly. It is diagnostic
  only and is intentionally absent from every aggregate verification target.
- `make first-content-harness` validates the ratified first-content manifest and its complete
  16-artifact replay bundle, then writes the owner-facing candidate-versus-baseline pacing report.
  Pacing drift is reported rather than vetoed; deterministic invariant failures still fail the run.
- `make epoch7-content-harness` validates the ratified eighteen-artifact T0–T1 candidate as a live
  replay bundle and writes the owner-facing `content_dynamics.v1` candidate report. The strict
  scenario executes one natural Active-Play control pair, Fiscal sweeps after one and four periods,
  Pitch seeds 0–63, and Permit accrual at zero and maximum Fiscal credit: 69 runs total. It invokes
  production-owned simulation boundaries, rejects a changed natural opportunity draw, and derives
  its transition ceiling from the four policies rather than accepting an arbitrary budget.
- `make t0-t1-role-check` runs the candidate-specific role activation/control matrix.
- `make t0-t1-upgrade-check` runs both candidate branch-proof matrices and preserves a diagnostic
  report for each failing scenario.
- `make t0-t1-relevance-all` first runs the role matrix, then runs both phase-scoped relevance gates
  and both branch-proof matrices even when one reports content findings; the final target remains
  nonzero if any constituent failed. Either relevance target can still be run alone with
  `make t0-t1-relevance` or `make t1-relevance`.
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

Before that mint, the candidate lane loads the complete ratified promotion manifest directly and
does not fabricate an epoch coordinate, registry row, snapshot, or golden. Its schema-v1 report
records the manifest path, scenario and bundle hashes, declared/executed runs and transitions,
byte-sorted observations, and an empty invariant-failure list. Active-Play reconciliation and
partition invariance, Founder-v21 Fiscal event/state agreement, Pitch certified-result conversion,
Permit hardcap identity, exact artifact identity, and transition ceilings fail the run; the first
pacing values remain owner-facing observations without an invented pass envelope. After mint, the
separate baseline commit snapshots those accepted bytes and creates the first historical golden.

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
