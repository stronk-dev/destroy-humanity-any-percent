# D-008/D-009/D-015 — retained receipt-payload tranche

Status: predeclared research only. Product source `7e8aa70`; planning after
`14884fa`. No export/deletion policy or product edit is authorized here.

## Question and selected population

Trace the current receipt persistence and identity-bearing nested fields for
four families: core `intent_records.receipt` (including Company/Founder
Minigame and Soul transitions), Minigame create/command receipt `response`
rows, `guild_intent_records.receipt`, and the Soul recovery session receipt.
These are retained or potentially retained after account unlinking and can
have distinct response/public API shapes. Identify same-byte copies into
run/Founder logs, transport and archives. Do not collapse all JSON called a
"receipt" into one schema.

This tranche excludes bootstrap receipt ciphertext/tombstones (already mapped),
all unselected per-command receipt variants, and operator logs/alerts. It is
not a field-complete player export design.

## Method and controls

1. Read current Up migration types/constraints and the production insert,
   replay and deletion paths for each selected store. Name the exact receipt
   or response encoder/serializer and any dynamic envelope boundary.
2. Trace typed nested IDs, other-party data, result/score/quality and token
   fields. A `request_hash` or `certified_result_hash` is not a substitute
   for retained JSON; a public response shape is not automatically the stored
   idempotency receipt. Distinguish a Soul progress token from credentials.
3. Execute cold relevant non-DB and real-Postgres tests. Where a test leaves
   fixture rows, read only their JSON key sets or byte equality. State if
   deletion occurs in the *same* scenario, and whether the fixture can
   falsify missing nested fields or an incorrect retention claim.

## Exit and refusal

Produce a bounded receipt dossier, update parent inventory, decision evidence,
shared backlog for genuinely new gaps, 1.0 board and append-only logs. Record
residual receipt families explicitly. No product/test/migration, RFC-state,
owner-copy or release-state change; no duration, player disclosure or legal
conclusion is adopted by source inspection.
