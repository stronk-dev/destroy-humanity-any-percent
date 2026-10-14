# D-008/D-009/D-015 — retained event-payload tranche

Status: predeclared research only. Product source `7e8aa70`; planning after
`eeb2a3f`. No export/deletion policy or product edit is authorized here.

## 2026-10-14 source refresh — declared at 089849ad

Re-derive the current ordered Up migration relation/kind/version population through 00084,
compare the actual Go AllEventKinds and version predicate, and trace the twelve added event
kinds plus run_started v2/Founder exit.v2/reputation frozen data. No policy or emitted payload
change. A temporary read-only Go test observer may log the native registry/version population;
remove it after execution without changing existing assertions. Keep a small source census and
exact command/result in the owning log, not a new tracking validator. Source SQL constraints
are not live installed-schema evidence: the validated Docker socket currently times out after
5.002526s with zero bytes, so no new Postgres/delete/restore run or integration claim follows.
Confirm that 00075–00084 add no table/column before retaining 60/60's narrow SQL-column claim;
nested new state/receipt/event bodies are separate coverage, never inferred from that count.

## Question and population

For the single core `events` table, enumerate the current accepted kind/version
denominator from the latest Up migration and the Go write validator. Trace the
identity-bearing payload families most relevant to account deletion:
run/Founder progression, Guild accrual, Minigame resolution/rating, and Soul
recovery. Identify whether payload validation, SQL retention, transport copying
and existing tests together prove removal or preserve linkage.

This tranche excludes separate `guild_events`, projection tables, free-form
receipts, historical save encoders and verification dead letters. It is a
schema/risk map, not a field-by-field export schema for all event kinds.

## Method and controls

1. Compare the latest `events_kind_check` and `events_schema_version_check`
   Up constraints against `save.AllEventKinds`, `validEventSchemaVersion` and
   `validateEventPayload`. Count accepted SQL kind/version pairs separately
   from Go-emittable pairs; retain the historical-version distinction.
2. Trace selected payload keys to producer(s), SQL writer, reader/transport
   and account deletion. Distinguish a same-player Founder/run reference from
   a Minigame session or another party's identity. Do not infer that UUID
   syntax, strict JSON decoding or archive immutability gives an export rule.
3. Execute cold relevant validator and real-Postgres tests if available. State
   their exact populated fields and whether account deletion occurs. Positive
   control: a persisted nested Founder/company/session identifier is visible.
   Negative control: a test that only counts events is not nested-payload
   preservation proof; a SQL-accepted historical schema version is not a
   currently emitted version.

## Exit and refusal

Record the denominator, selected identity schema, producer/consumer chain,
test limits and residual work in a dossier; update the shared backlog if a
new falsifiable gap appears, plus the parent inventory, decision evidence,
1.0 board and append-only logs. No owner copy, product/test/migration changes,
rights policy, RFC status, release promotion or legal conclusion.
