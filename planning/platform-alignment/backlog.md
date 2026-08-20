# Tracked alignment backlog

This is the tracked interim ledger for repository/release findings while `design/BACKLOG.md` and
`design/research/*` remain intentionally gitignored from the public repository. It does not replace
the internal design backlog. D-002 must decide the durable publication/storage model; until then,
every alignment finding is recorded in both places so a fresh clone does not lose the program.

| ID | Observation | Class | Route |
|---|---|---|---|
| RP-001 | Current-head hosted CI never reaches a verdict; the harness is killed at 30 minutes in five consecutive current-product runs. | defect | R-001 / CI Baseline |
| RP-002 | Production Docker/Caddy/Compose packaging is absent. | release defect | D-001/D-006 / Deployment |
| RP-003 | Backup, restore, rollback, and disaster rehearsal are absent. | release defect | R-006 |
| RP-004 | Player data export has no endpoint, schema, UI, or proof. | release defect | D-008 |
| RP-005 | Account deletion is backend-only and has no player workflow/disclosure. | integration defect | D-009/R-003 |
| RP-006 | One-time recovery credentials are silently stored with no display/download/recovery workflow. | risk/integration defect | D-005/R-003 |
| RP-007 | The ruled local-only outage fallback and import path are absent from the production client; older tech/Account text also overstates offline mode relative to the adopted server-anonymous default. | intent/body/implementation contradiction | ruling-author reconciliation plus account successor |
| RP-008 | Game UI proves bootstrap-to-Desk, not the complete first hour. | acceptance defect | Game UI AC1 / R-004 |
| RP-009 | Accessibility proof does not cover complete assistive user workflows. | release proof gap | R-005 |
| RP-010 | Only T0–T1 and one minigame exist; most of the designed product and ending are absent. | scope question | D-001/D-007 |
| RP-011 | Client third-party license delivery is absent. | packaging defect | Deployment |
| RP-012 | Sunset/self-host/export promises are unimplemented intent. | owner question | D-003/R-006 |
| RP-013 | Repository disposition contradicts itself: public GitHub; internal records say private/local-only. | owner question | D-002 |
| RP-014 | Current state, README, Deployment RFC, RFC handoff, coverage map, and live plans disagree with HEAD. | record defect | lifecycle reconciliation |
| RP-015 | The mandated shared `design/BACKLOG.md`, research matrix, dossiers, and coverage map are gitignored, so a fresh clone cannot receive the supposed shared memory. | process defect | D-002/D-010; tracked interim ledger |
| RP-016 | Minigame Platform says its normative body is reconciled and requires a combat-duel tenant, while its later ruling, combat RFCs, catalog tree, and production registry prove no such tenant exists. | normative contradiction | ruling author reconciliation before Minigame Platform acceptance/archival |
| RP-017 | `design/03` rules that paying hobbies never restore Soul and names missing §5c as the sole recovery source, but later calls Go restorative and explicitly grants Arcade Soul restoration. | owner-authored design contradiction | design author reconciliation; no RFC inference |
| RP-018 | `design/11`'s cold-open specimen displays a fabricated WR while its later owner ruling forbids any WR before validated boards exist. | owner-authored design contradiction | design author reconciliation before shipping cold-open copy |
| RP-019 | `design/12` specifies a `MatchActor`/tick minigame socket, while the accepted platform contract and runtime use pure tenants plus DB-authoritative sessions. | design/RFC architecture drift | design deviation reconciliation; Minigame Platform body repair |
| RP-020 | `design/12` still says production ships by “pack push + hot reload” immediately before a ruling that production hot reload is forbidden and epoch-stamped deploys are mandatory. | owner-ruling/body contradiction | ruling author reconciliation |
| RP-021 | The T0-T1 capacity change claimed a measured hosted harness lane, but no Actions run exists for that commit; archival consumed native/Docker checks, and every later hosted run exceeds the adopted ceiling. | acceptance/closeout defect | R-001 / CI Baseline successor |
| RP-022 | `HARNESS_WORKERS` parallelizes only standard pacing; relevance registry rows and their internal experiment arms remain serial. | performance/architecture defect | R-001 measurement, then CI Baseline amendment |
| RP-023 | `mode=check` emits no phase/row progress or termination artifact, so a hosted kill discloses neither completed population nor invalidation state. | evidence/instrumentation defect | R-001 / fail-loud diagnostic contract |
| RP-024 | Account AC2 has real family-revocation tests and transport reauthentication code, but no fixture composes revocation with an already-connected socket and proves the required subscribe/disconnect boundary. | acceptance proof gap | Account/Transport integrated negative witness |
| RP-025 | Root Make test selectors expand `SAVE_TEST_FLAGS` unquoted; a valid parenthesized Go `-run` regex becomes shell syntax and the declared focused lane never starts. | tooling defect | CI Baseline selector hardening with a seeded regex fixture |
| RP-026 | Game UI AC4's browser fixture injects drain and resync together but asserts only accessibility/mechanical-ID absence; it proves neither story beat's copy, timing, recovery action, nor refresh transition. | acceptance-oracle defect | Game UI AC4 successor fixture under accepted body |
