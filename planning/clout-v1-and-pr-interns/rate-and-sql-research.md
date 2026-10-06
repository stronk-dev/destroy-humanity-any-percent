# R-012 — live rate precision and Postgres fidelity

2026-10-07. Baseline `9050fe4d`; predeclaration `9a35aa78`.
Test-only research, not a production repair, accepted save format or AC6 pass.

## Findings

The preceding anchor prototype cannot be transplanted literally. Canonicalizing
computed rates early changes valid payouts; canonical JSON bytes do not survive
Postgres jsonb. Both findings concern the research prototype, not new production
save defects. Existing Company saves preserve the producer context in the sampled
Go round-trip; they do not persist the prototype's canonical rate list.

| Declared population | Executed result |
|---|---|
| 64 admitted Go producer profiles: eight counts × two efficiencies × four times | All raw rate bits change on twelve-digit serialization; five final payouts change |
| Same 64 existing Company Encode/Restore contexts | Every raw rate and accrual delta preserved exactly; actual Evaluate cash equals raw primitive |
| Independent pinned TS arithmetic, same 64 diagnostics | All raw and canonical-rate outcomes agree with Go |
| 1,215 valid prototype snapshots through actual Postgres16.15 jsonb | Strict byte decoder refuses all; test-only logical reader preserves every field exactly |
| 16 strict-negative payloads through actual SQL casts | Twelve corrupt/domain payloads refused logically; one syntax refusal (typed22P02); three representation-only differences normalized |

CPU inputs use the admitted x6/owned-PR1 diagnostic state, consistent purchased
total, cash0 and a rehashed test-local cap1e100. Counts extend to MaxExactInteger;
times are2000/3114/2045/60000ms. This is an admitted synthetic population, not
naturally reached progression or served client producer evidence. Raw mantissa
round-trip text and IEEE bits are diagnostic strings, not a newly allowed wire.

Example: count123456789, online60000ms yields actual8.41341961825e15;
pre-rounding the rate yields8.41341961824e15. The other differences cover three
offline cells at that count and online60000ms at MaxExactInteger. Do not change
legitimate one-shot semantics merely to manufacture partition equality.

The SQL population consumes the frozen preceding artifact:594 valid cases ×
one/split, twelve boundaries ×before/after, three near-cap post-debit snapshots.
The new SQL reader decodes known typed fields/EOF, re-encodes them and invokes
the existing canonical-input/recomputed-wire validator. No numeric, source-count,
version or reconstruction check is dropped. Full field equality, not cash alone,
is required. Unknown fields and valid-but-wrong wire/rate remain refusals.

SQL discards duplicate-equal keys, order and excess trailing whitespace. Thus
those three original strict transport negatives become the same valid logical
object. This does NOT authorize accepting unvalidated transport or reconstructing
duplicate history from SQL. Trailing another JSON value fails its cast with
typed SQLSTATE22P02; unrelated SQL errors invalidate measurement.

Only a new transaction-local TEMP table is written, followed by explicit
rollback. Negative casts use savepoints. No live save table, migration, truncate
or cleanup operation runs. Actual environment:Postgres16.15,linux/arm64.

## Evidence and reproduction

Separate source-pinned corpora:

- [Producer observation](../../testdata/axis-stack/rate-source-research-v1.json):64 complete rows, fourteen selected sources.
- [SQL observation](../../testdata/axis-stack/anchor-sql-research-v1.json):1,215 complete normalized objects, sixteen negatives, ten selected sources and actual environment.

```sh
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxis.*Research -v -count=1'
make test-client typecheck
make test-save-integration SAVE_TEST_PACKAGES=./production SAVE_TEST_FLAGS='-run TestAxisAnchorSQLIntegrationResearch -v'
# Explicit measurement writers; not automatic CI lanes:
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisRateSerializationResearch -v -count=1' UPDATE_RATE_RESEARCH=1
make research-axis-anchor-sql
```

The ordinary host SQL test visibly skips as NOT EXECUTED. Enabling its writer
without the declared DB fails before writing. CPU and SQL writers are separate;
neither can claim a truncated population. A one-minute SQL operation guard can
only invalidate an incomplete run, never certify performance or relax a bound.

I substituted rounded accrual for raw accrual:Go fails context equality before
its writer, and TS fails five semantic payout checks plus its source check.
I made the SQL reader silently repair stored wire:the actual SQL population
fails at wrong-valid-rate admission before its writer. All three faulty source
files restore byte-exactly; both artifacts stay unchanged by the probes.

Before landing, local review caught environment metadata contaminating the SQL
golden comparison. The instrument now validates recorded/actual Postgres16 and
Linux ARM64/AMD64 identities, records actual writer provenance, and explicitly
excludes ONLY patch-version/architecture from replay equality. All source hashes,
normalized objects and statuses still compare exactly. Local comparator fixtures
distinguish environment-only changes, semantic changes and invalid environments.
They are not an executed AMD64 observation. The declared local AMD64 CI lane was
attempted, but fails before Go runs with exec-format error. No architecture
bypass, CI configuration edit, cleanup or hosted-green claim follows.

The AMD64 pull also changed the shared Go image tag, making the next ordinary
native invocation fail before execution. A declared native-service image pull
restored ARM64 without deleting images/data or editing Compose. Final cold
native replay executes all five research tests on real Postgres, with no skips
and unchanged observations. This repairs local test availability, not AMD64 CI.

Restored client9499passes/340visible skips; types/Svelte clean. Relevant vet and
CI topology pass. Cold production/economy/decimal remains RED only at the original
AC6 property; focused execution confirms128 cases,27fail/101pass. Both earlier
research corpora reproduce unchanged. This research does not make CI green.

## Next work and limits

The sampled existing producer context is a stronger candidate than a rounded
rate list; logical SQL validation is necessary instead of byte framing. Neither
is an adopted persistence contract. These are distinct experiments:the CPU
context round-trip does not traverse SQL, and the SQL prototype does not contain
the live raw producer context. Do not join them into unexecuted integration proof.

Next predeclare actual mode/offline cap/banking and boost/provision boundaries,
full resource/Company state, action settlement and replay/content-pin questions.
Then establish activation/migration, receipt/clock and actual SQL Service/Store
integration before a buildable repair contract. No new save field, changed K3,
balance value, raw-rate wire, tolerance or waived AC6 is authorized here.

Entire new span after9050fe4d, including predeclaration and later record edges,
needs Claude's designated review. Earlier spans remain independent. Full
nine-tier/platform1.0 remains active; no status/checkbox, archival/mint/push/
deployment or release promotion.
