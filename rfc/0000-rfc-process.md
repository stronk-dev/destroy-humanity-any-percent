# RFC-0000: The RFC Process

- **Status:** accepted
- **Author:** Marco
- **Created:** 2026-07-27
- **Supersedes / superseded by:** —

## The four tiers of documentation

| Tier | Directory | What it is | Mutability |
|---|---|---|---|
| **Design** | `design/` | Intent and evidence: the game design docs. Where ideas come from. | Amended rarely; never the implementation spec |
| **RFC** | `rfc/` | Active implementation specs. Every system that gets built is specified by one or more RFCs before implementation. | Living until implemented, then frozen and moved to `rfc/archive/` |
| **Planning** | `planning/` | Per-RFC working documents: the implementation plan and a running log. The durable record of long-running jobs. | Living during implementation; archived on completion |
| **Docs** | `docs/` | Canonical description of what actually exists — architecture, data formats, runbooks, balancing values as-shipped. | Always current; updated as part of every implementing change |

**The flow:** design → active RFC (specification) → planning (how + log) → implementation →
canonical docs + RFC/planning archives. Once a feature ships, `docs/` is the only current
description of its behavior; archived RFCs explain why and how it arrived.

## RFC lifecycle

`draft` → `accepted` → `implementing` → `implemented` → (`superseded` | `withdrawn`)

- **draft**: under discussion; anything may change.
- **accepted**: scope and approach agreed by Marco; implementation may be planned.
- **implementing**: at least one planning doc exists and work is underway.
- **implemented**: shipped; canonical behavior has been distilled into `docs/`, and the frozen
  RFC has moved to `rfc/archive/`.
- **superseded**: a later RFC replaces it (must link both ways). **withdrawn**: abandoned, kept for the record.

## Rules

1. **Identity:** use a descriptive, stable filename such as `save-layer.md`. Existing numbered
   RFCs retain their historical identifiers, but new RFCs do not need a global sequence number.
   Follow-ups use a descriptive name and declare `Parent`/`Amends` metadata rather than
   pretending to be an unrelated top-level system.
2. **Scope discipline:** an RFC should be implementable in bounded work. If scope grows during drafting or implementation, **split**: the new scope becomes a new RFC referencing the parent. Note the split in both.
3. **Amendments:** small clarifications to a not-yet-implemented RFC are edited in place with a
   changelog line. Changing an implemented system's accepted contract requires a follow-up RFC
   linked to the archived parent. Fixing behavior to conform to that contract, or refactoring
   without changing it, does not require a new RFC; cite the existing contract and use the
   delivery procedure below. The archived parent remains immutable; `docs/` describes the
   corrected current behavior.
4. **Design linkage:** every RFC cites the `design/` sections it specifies. Deviations from design docs are called out explicitly in a "Deviations from design" section — the RFC wins once accepted, but the divergence must be visible.
5. **Agent rule:** coding agents implement **RFCs**, not design docs. If needed spec is missing, that's a `DESIGN-GAP` → propose a draft RFC, don't improvise.
6. **The index** (`rfc/README.md`) lists active work and the archive, including parent/follow-up
   relationships; keep it updated in the same commit as any status or location change.

## Delivery procedure — risk-based, not per-edit ceremony

Adopted by Marco's 2026-10-07 direction. This procedure governs routine work for both agents;
older program-wide instructions cannot require extra ceremonies for every small change.
It does not waive explicit feature acceptance criteria or lower the full nine-tier 1.0 floor.

1. **Define the outcome.** Identify the owning accepted contract, the behavior to change and
   the test that would expose a mistake. For a routine fix, a few sentences in the existing
   plan or work update suffice. No separate planning commit is required. New mechanics,
   architectural contracts and unresolved product choices still require an accepted RFC/ruling.
2. **Make a coherent change.** Keep the fix and its tests together. Repairs, refactors, test
   improvements and editorial corrections within an existing contract do not need their own
   RFCs. Do not mix unrelated systems or silently extend a test-only batch into production.
3. **Verify the risk.** Use the table below. Run the checks; inspect what they exercised and
   whether prerequisites actually ran. A test count, source inventory or consistent tracker
   is not evidence that the player can complete a workflow.
4. **Review and record.** Check the actual diff, record a concise result in the owning log,
   and get designated cross-party review for behavior/specification batches. Group related
   implementation, test and record commits into one reviewable range. Complete the feature's
   acceptance criteria before archival; don't repeat every acceptance experiment per edit.

The unit of delivery is a coherent behavior change, not each file edit or tool call. During
development, iterate with the focused test; run the affected package checks and necessary
integration checks on the finished batch. A further change invalidates evidence for what it
touches, not automatically every unrelated suite. Reuse existing test helpers and journeys;
do not build a new measurement framework for a routine regression.

Pending designated review prevents acceptance/archival, not all further work on already
accepted contracts. Continue safe work, keep its range explicit, and bring related changes
back as one review batch. Do not cite an unreviewed new contract as accepted authority.

### Verification by change risk

| Change | Meaningful verification |
|---|---|
| Editorial docs, comments, planning records | Check the diff, paths and factual claims. No software suite just because Markdown changed. Normative contract changes still need review; generated/consumed files need their consumer check. |
| Test/tool/diagnostic change | Execute it against its real subject; for a journey driver, run the journey. New assertions/validators need a relevant regression or invalid case. Diagnostics must expose the relevant signal. Don't rerun unrelated suites or create a tracking validator. |
| Local behavior fix or refactor | Focused regression/unit tests plus the affected package's type/build/lint checks. A bug regression must fail on the old behavior. Include the integration boundary if it changes. |
| Cross-system gameplay, persistence, security/privacy, money prohibition, numeric or deployment change | Affected unit/contract tests plus a real integration journey and meaningful failure paths. Use actual Postgres for DB behavior, mounted DOM/current server for player flows, both runtimes' vectors for shared math, and restore/clean-host checks for deployment where required. |

Before a feature closeout, run its entire acceptance population. Before publication or an
integrated milestone, run `make verify-push`, the aggregate of the push/PR job targets.
It shares targets with Actions, not necessarily its environment: a native Go run can skip DB
tests without Postgres, and native browser/CPU behavior can differ from Linux. Use the declared
service/environment checks below for affected boundaries; reconcile the hosted run when
available. Local green is not proof of hosted green. Exhaustive
balance/research and clean-host release checks stay at their declared feature/release gates,
not on every comment edit. If a lane is unavailable or red, report that limitation and keep the
affected acceptance/release claim open; don't weaken assertions, silently skip, or rerun until
the last green result conceals failures. Unrelated red lanes do not forbid scoped work.

### Choosing the actual checks

Run commands from the repository root. These are selectors for existing tools, not a new
checklist requiring every row for every change. Replace the example package/file with the
changed owner, include affected consumers, and name the requirement the test checks.

| Changed boundary | Existing check / required observation |
|---|---|
| Go logic, no DB boundary | `make test-go GO_PACKAGES='./publicapi ./publicread' GO_TEST_FLAGS='-count=1'` and `make vet GO_PACKAGES='./publicapi ./publicread'`. A narrower `-run` is useful while debugging; finish with the affected packages. |
| TypeScript logic | `client/node_modules/.bin/vitest run --root client test/game-ui-runtime.test.ts`, plus `make typecheck`; `make build-client` when shipped client/bundling changes. A runtime double is not a server integration. |
| DOM/input/focus/layout | `make test-browser BROWSER_TEST_FLAGS='test/game-ui-screens-browser.test.ts'` and `make typecheck`. Assert actual input, focus, state and rendered outcome, not only component existence. This target also runs the existing screen-performance check. |
| Persistence, migrations, DB-backed authorization | `make test-save-integration SAVE_TEST_PACKAGES='./save' SAVE_TEST_FLAGS='-run Integration -v'`; migration changes use `make validate-migrations`. Check that the named DB tests actually execute, including a rejected/rollback case relevant to the change. |
| Browser → server gameplay/receipts/recovery | `make test-game-ui-composed`. Verify the changed path is exercised, including its authoritative result. If not, extend that journey or add a focused real-boundary test; an unrelated green journey is insufficient. |
| Generated API, shared math, deployment, balance | Use the owning feature's existing contract/vector/restore/harness targets for the affected requirement. API changes include `make api-check`; shared numeric changes require both Go and client vectors. A schema drift check alone does not prove a handler or UI works. |
| Linux/CI-specific failure | `make test-go-ci CI_TEST_PACKAGES='./save' CI_TEST_FLAGS='-v'` for selected Linux-amd64/Postgres tests; `make verify-server-ci` for the server job; `make test-browser-ci` for Linux browsers. Containers approximate the hosted environment, not exact runner hardware. CI/workflow edits require checking the actual affected job, not YAML validity alone. |

For each test, answer: **what wrong behavior would make this fail?** For a fix, demonstrate
the regression on the old implementation before correcting it, or use an already recorded
reproduction. Assert meaningful outcomes and relevant failure behavior. A real boundary is
mandatory only when the change or claim crosses it; mocks remain useful for isolated logic.
If the intended failure escapes the test, improve the test before claiming the fix verified.

### What counts as a useful test

- Observe the actual requirement: a Buy test checks one submitted intent, an authoritative
  receipt and persisted ownership; a restore test reads restored data; an accessibility test
  exercises the supported input path. A component existing or a string prefix matching is not
  a substitute. Match the fixture to the failure, not just to the happy path.
- Prefer a regression on old broken behavior or a retained invalid-input/failure-path case.
  Deliberate source mutation is warranted for a new critical oracle whose discrimination those
  cases cannot establish, or when an explicit acceptance criterion requires it. No mandatory
  mutation of every assertion, and no repeated severing of unchanged checks in each batch.
- Surface uncaught errors, skipped dependencies and incomplete measurements. Diagnosing an
  intermittent failure does not authorize changing retries, timeouts or gameplay semantics.
- State the test boundary honestly. A mocked runtime can prove UI ordering or error handling,
  not persistence, server authorization or transport recovery. Claims about those boundaries
  need the corresponding real integration check; unrelated end-to-end journeys add no proof.
- Predeclare empirical research (question, population, method, threshold, limitations) when its
  results will decide product scope, balance or operating limits. Routine debugging, test runs,
  formatting and record corrections do not need a separate research protocol.

### One durable record, not mirrored narratives

The owning per-RFC plan/log is the detailed work record. `design/BACKLOG.md` records defects and
unresolved questions; other boards link to it rather than duplicate a second ledger. Update
`planning/roadmap-1.0.md` only for a material capability, blocker or next-action change; its log
gets milestone checkpoints, not a replay of every command. The alignment queue is optional
cross-system routing, not a second prerequisite for work authorized by the RFC and its plan.
Do not refresh historical audit inventories after every edit or invent progress percentages.

Keep ordinary Git for source, small necessary fixtures, concise results and reproducible
commands. Large/repetitive raw recordings belong in ignored local or explicitly authorized
artifact storage, with a small manifest/digest when needed. Ordinary CI, progress and review
must not require downloading raw research recordings. Do not delete existing evidence or
rewrite its history as part of adopting this procedure.

## Planning docs & the job log

For each RFC being implemented, create `planning/<rfc-slug>/`:

- `plan.md` — the implementation plan: task breakdown, sequencing, acceptance criteria (tests/gates), assignee (human/agent).
- `log.md` — an **append-only running log**, with one concise entry per coherent batch or
  useful handoff. Record outcome, changed paths/range, commands and results, unresolved limits,
  review status and next action. Do not paste transcripts or restate the whole release board.
  A fresh agent resumes from the current plan and latest checkpoint; older entries remain
  available for provenance, not mandatory cover-to-cover reading on every resumed fix.

On completion:
1. Distill outcomes into `docs/` (the canonical statement of what now exists).
2. Set the RFC status to `implemented`.
3. Move the RFC to `rfc/archive/` and its planning directory to `planning/archive/`. Never
   delete either — together they are the project's institutional memory.

## Docs conventions

- `docs/` is organized by system, not by history (`docs/architecture.md`, `docs/economy.md`, `docs/data-formats.md`, `docs/ops.md`, …).
- Every implementing PR/commit that changes behavior updates the relevant `docs/` page in the same change.
- When docs and code disagree, that is a bug in docs; fix it with the next change.
- Current code must be understandable from `docs/` without reconstructing a chain of RFCs.

## Changelog

- 2026-07-27: accepted as the initial numbered, lifecycle-tracked RFC process.
- 2026-07-27: amended by owner direction to rotate implemented RFCs into an archive, make
  `docs/` the sole canonical current description, and allow descriptive follow-up RFCs without
  consuming a global number.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
- 2026-10-07: Marco directs a proper, non-convoluted procedure. Adopt risk-based verification,
  coherent review batches, concise single-home records and separate raw research storage.
  Replace per-edit predeclaration/severing, duplicate-ledger and test-bytes-in-checkbox-commit
  routines; preserve explicit acceptance, independent archival review and full 1.0 obligations.
- 2026-10-07: remove the lifecycle rule's conflicting blanket follow-up requirement for routine
  contract-conformance fixes. Explicitly distinguish mocked checks from real integration claims.
  Verification: inspected the process diff against the risk table and existing CI targets;
  documentation-only, no software tests run. Normative amendment awaits cross-party review.
- 2026-10-07: make the procedure operational: one coherent batch, existing focused selectors,
  checks at the changed boundary, and explicit native-versus-hosted CI limits. Pending review
  does not halt safe work under accepted contracts. No new tracker, test framework or product
  change. Verification: command selectors checked against Makefile, Vitest configuration and
  Actions/Compose definitions; diff checked. No software suite for this documentation change;
  this normative clarification remains pending designated review, not self-approved.
