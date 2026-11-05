# World committed-output producer check — predeclared 2026-11-05

Source:2cc892d0, draft World WL1/WL2 readiness. No product or specification edits.

Question: can existing durable intent receipts, generic events, or Guild progress
be consumed directly as committed passive production? Compare actual production
and transaction outcomes, rather than assuming balance growth is production.

Population: the existing immutable epoch5 fixture (1 cash/s, manual1, generator
price10×1.13^owned, offline90%), current production/save code, real declared
Postgres. Controlled initial cash100, one generator; separate cash1000 row reaches
the tier-progress target. Seven arms: online purchase, same-time manual, online
manual, offline purchase, progress-saturated purchase, unaffordable purchase and
SQL-injected revision-write failure. Retry the successful purchase with the same
intent and later observation time. No natural acquisition/current epoch population
or behavioral measurement is claimed.

Observe: isolated pure Evaluate receipt/progress on a clone; actual Service.Handle
receipt, persisted save/cursor/revision, complete per-stream event payloads,
intent/run-log/outbox counts and rejected/applied log coordinate. The pure clone
is a reference, not a committed-output observer. Inject failure only in the
owned disposable test DB after mutation, before revision insertion; cleanup the
trigger/function. Retain every row/error; abort rather than drop failures.

Predeclared controls: online passive1 vs purchase net-10.3; offline passive0.9 vs
purchase net-10.4; same-time manual passive0 vs positive1; online manual passive1
vs net2. At cash1000, passive1 can coexist with ProgressDeltaPPM0. Rejection may
have a durable log/receipt without a new save; SQL failure must have neither new
revision nor any intent/event/log/outbox write. Retry must retain exact receipt
and all database counts. An unselected study is an explicit skip, never credited
as execution; a selected study without Postgres must fail. Assertions/controls
failing invalidate the observation.

Exit: retain small source-bound results and the source trace, distinguish what
existing records prove from missing producer semantics. This may require a
transaction-owned typed output fact in the World proposal. It does not select
qualifying resources/manual/reward policy, raw vs capped output, time/population
windows, rounding, contribution formula, outbox schema or replay/retention policy;
no World implementation/adoption/CI or release acceptance follows.


Follow-up declared after the first valid33605 observation: Handle dispatches
committed events even on an idempotent retry. On the next run, observe its actual
EventProjector callback through the existing injection seam. Require the retry
of online_purchase to deliver the SAME single event ID/payload twice while all
five durable counts and the full saved state remain unchanged. This measures
delivery versus creation; no general exactly-once consumer is implemented.
