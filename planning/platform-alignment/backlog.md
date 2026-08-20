# Tracked alignment backlog

This is the tracked interim ledger for repository/release findings while `design/BACKLOG.md` and
`design/research/*` remain intentionally gitignored from the public repository. It does not replace
the internal design backlog. D-002 must decide the durable publication/storage model; until then,
every alignment finding is recorded in both places so a fresh clone does not lose the program.

| ID | Observation | Class | Route |
|---|---|---|---|
| RP-001 | Current-head hosted CI never reaches a verdict; harness exceeds 30 minutes in four consecutive runs. | defect | R-001 / CI Baseline |
| RP-002 | Production Docker/Caddy/Compose packaging is absent. | release defect | D-001/D-006 / Deployment |
| RP-003 | Backup, restore, rollback, and disaster rehearsal are absent. | release defect | R-006 |
| RP-004 | Player data export has no endpoint, schema, UI, or proof. | release defect | D-008 |
| RP-005 | Account deletion is backend-only and has no player workflow/disclosure. | integration defect | D-009/R-003 |
| RP-006 | One-time recovery credentials are silently stored with no display/download/recovery workflow. | risk/integration defect | D-005/R-003 |
| RP-007 | Fully offline-anonymous play is promised but absent from the production client. | intent/implementation contradiction | D-004 |
| RP-008 | Game UI proves bootstrap-to-Desk, not the complete first hour. | acceptance defect | Game UI AC1 / R-004 |
| RP-009 | Accessibility proof does not cover complete assistive user workflows. | release proof gap | R-005 |
| RP-010 | Only T0–T1 and one minigame exist; most of the designed product and ending are absent. | scope question | D-001/D-007 |
| RP-011 | Client third-party license delivery is absent. | packaging defect | Deployment |
| RP-012 | Sunset/self-host/export promises are unimplemented intent. | owner question | D-003/R-006 |
| RP-013 | Repository disposition contradicts itself: public GitHub; internal records say private/local-only. | owner question | D-002 |
| RP-014 | Current state, README, Deployment RFC, RFC handoff, coverage map, and live plans disagree with HEAD. | record defect | lifecycle reconciliation |
| RP-015 | The mandated shared `design/BACKLOG.md`, research matrix, dossiers, and coverage map are gitignored, so a fresh clone cannot receive the supposed shared memory. | process defect | D-002/D-010; tracked interim ledger |
