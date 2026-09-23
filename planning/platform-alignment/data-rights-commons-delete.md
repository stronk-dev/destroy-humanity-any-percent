# Commons membership after account deletion — bounded result

**Coordinate:** product `7e8aa70`; predeclared at `f050057` in
`data-rights-commons-delete-plan.md`; run 2026-09-23.
**Classification:** [V] for the executed SQL predicate/account-delete
outcome. No full World goroutine or player-surface claim.

## Population and execution

A temporary extension of `TestAccountSessionIntegration` on the declared
Postgres service seeded one Commons cohort, Founder assignment, Company
Compact membership (`member=true`) and member sample for the account's
**currently active** imported Founder and Company stream. The pre-delete
control required `account_founders.archived_at IS NULL`, account ownership,
and one match from each of two exact production predicates:

- `gameserver/world_source.go`'s `count(DISTINCT membership.founder_id)`
  joined to the server assignment with `membership.member=true`;
- `commonsprojection.Projector.refreshScope`'s server-scope member-sample
  join with `m.member=true`.

The first diagnostic invocation did not compile: the temporary test used
`imported.FounderID`, but the `Founder` response field is `ID`. This was a
test-driver error, not product evidence. After correcting the temporary
field reference, the same cold command completed without skip:

```text
make test-save-integration SAVE_TEST_PACKAGES='./account' SAVE_TEST_FLAGS='-run TestAccountSessionIntegration -v' SAVE_TEST_COUNT=1
=== RUN   TestAccountSessionIntegration
Commons deletion probe: active_before=1 world_before=1 samples_before=1 archived_stream=1 world_after=1 samples_after=1
--- PASS: TestAccountSessionIntegration (0.75s)
```

The real authenticated Account API deletion returned its existing 204 and
its original account/session/Founder assertions passed. The active stream
became archived, but the world-count and sample-selection predicates each
still returned one. In a rollback-only transaction, setting the same
membership's `member=false` made **both** counts zero; this negative control
shows the predicates respond to the membership flag. The temporary test edit
was removed by exact patch and `git diff --exit-code --
server/account/account_integration_test.go` returned 0.

## Boundary and next contract

RP-119 is now an executed live-count/sample-selection defect, not merely a
source hypothesis. The probe copied the exact production SQL predicates but
did not invoke `SampleWorld`, run `refreshScope`, render a player display or
test historic-contribution policy. It used a directly seeded Commons row,
not an actual signed-event projector chain. The current delete transaction
does not synthesize a leave or clear membership, and neither predicate checks
archived stream/account status. A later accepted Account/Commons contract must
decide active-vs-historical semantics, implement the chosen transition with
replay safety, and prove the whole projection→delete→world/health workflow.
D-009/D-015 and owner/legal disclosure remain open; no retention duration or
release claim is inferred.
