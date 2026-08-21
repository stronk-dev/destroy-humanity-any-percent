# Row-level test oracle structural inventory

Coordinate: product tree `190a4fa`; 2026-08-21. This is the declaration/body-signal result of
`test-oracle-row-plan.md`; it does not assign semantic proof verdicts.

## Reconciled population

The extractor reconciles all 151 server test files and all 43 tracked client test/helper sources.
It emits **802 static oracle units**:

| Runtime / kind | Units |
|---|---:|
| Go top-level `Test*` | 591 |
| Go top-level `Fuzz*` | 1 |
| Client plain `it`/`test` | 174 |
| Client parameterized `.each` declarations | 19 |
| Client conditional `.skipIf` browser declarations | 15 |
| Client zero-declaration helper/type-contract units | 2 |

The 592 Go owners contain 88 static `t.Run` calls and one fuzz seed. Those are expansion signals,
not fabricated independent-case counts. Twenty-five files containing 47 Go functions are already
classified `integration` by the file ledger. The client browser declarations remain conditional:
ordinary Node execution skips them, while the declared browser lane supplies `document`.

Every unit records runtime, file, declaration/name, source line, exact body hash, kind, expansion,
dependency/guard, assertion, keyword-level negative signal, subject, and file class. The two helper
units are `browser-error-guard.ts` and the negative typecheck fixture
`game-ui-run-end-contract.ts`; neither is counted as a passing runtime test by existence.

## Signal limits

- Eighty-four bodies contain a direct environment/skip/architecture/deadline/guard signal. This is
  a lower bound, not dependency truth: shared test helpers can call `t.Skip` outside the top-level
  body, so the 40-site skip ledger and file execution lane remain authoritative.
- Seventy-four Go functions contain static subtest/fuzz expansion and 421 total units contain at
  least one broad negative keyword. Neither signal proves discrimination. “Invalid” in a variable
  name or a generic malformed-input rejection may be irrelevant to the claimed production defect.
- Assertion counts show that an oracle exists syntactically. They do not show its expected value can
  distinguish the subject, that a fixture can trigger the defect, or that a conditional test ran.
- Client `.each` and Go table loops can generate thousands of runner cases; static source cannot
  invent those dynamic denominators. Runner totals remain command evidence with their visible skips.

## Controls

A rerun reproduces the 802 rows and body hashes byte-for-byte. The extractor rejects a dropped or
duplicate declaration, omission of a zero-declaration helper, and a seeded body-hash change. It
also fails unless the prior 151/592 and 43-source populations reconcile exactly.

The next ledger must attach exact fixture/dependency/oracle/negative-control verdicts. No test name,
body hash, assertion count, negative keyword, subtest count, green package, or integration filename
receives semantic credit from this checkpoint.
