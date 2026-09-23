# RP-119 — Commons count after account deletion

**Status:** predeclared diagnostic, 2026-09-23; product source `7e8aa70`.
No Commons, account or retention behavior change is authorized by this probe.

## Question and population

Does the exact `gameserver/world_source.go` Commons count, and the sample
selection used by `commonsprojection.refreshScope`, continue including a
currently active Founder after its account is deleted and its stream is
archived? Prior source tracing says both SQL readers use `member=true` without
checking stream archival. A source inference is not an executed transaction.

Temporarily extend the existing real-Postgres Account API integration test.
After its final import and before deletion, seed one schema-valid cohort,
Founder assignment, Company Compact membership and member sample attached
to that test's **active** imported Founder and Company stream. Count the
member and selected sample with the exact production predicates before
deletion. Run the real authenticated `DELETE /api/v1/account` route. Assert
the account is gone, the Founder is unlinked/archived, the stream is archived,
and both production-predicate counts are observed afterward.

## Controls, limits and exit

Pre-delete counts must be one and the Founder active, or the run is invalid.
After deletion, in a rollback-only transaction, set `member=false` and demand
both query counts fall to zero; this proves the oracle is sensitive to the
membership flag. Run cold with `-count=1 -v` on the declared Postgres service.
Remove the temporary test edit exactly. Record the observed count and scope
in RP-119 and the decision/1.0 trackers; do not call a raw SQL predicate
probe a full World-goroutine/player-surface witness. Whether deletion should
exclude or preserve historic Commons contributions is D-009/D-015 owner work.
