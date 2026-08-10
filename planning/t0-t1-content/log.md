# T0–T1 Playable Content implementation log

## 2026-08-10 — Codex acceptance review

Foundation and Relevance are no longer blockers, but the content itself is not literal enough to
load: no exact schema-v4 rows, presentation binding, event-copy manifest, scripted-failure
transition, session-catchup log contract, or harness scenario exists. Two required roles name
mechanics the accepted Foundation explicitly did not ship. Epoch/relevance claims are also stale
after First Content.

T01-C1–T01-C9 capture the remaining owner contracts. Status remains draft and no balance bytes or
epoch entry were authored.

## 2026-08-10 — Candidate drafting and T01-C10 bounce

The owner-ratified T01-C1–C9 contracts were converted into draft candidate documents under
`balance/testdata/t0-t1/`. The economy schema-v4 and Routes candidates load through the real Go
loaders and are covered by focused candidate tests. Presentation, event-copy, curriculum, and
first-hour harness documents are exact candidate grammars for owner review; they do not touch
production artifact paths or authorize a mint.

The mandatory Relevance candidate cannot be authored honestly against schema v1. Its loader
requires a concrete `from_gate`, while T0 content is available from run genesis before the first
gate. Substituting `gate.t0_to_t1` would make the report skip T0. T01-C10 proposes the narrow
schema-v2 correction: nullable `from_gate` means run genesis in both policy windows and scenario
segments, with schema-v1 bytes preserved. Candidate assembly remains open pending that ruling.
