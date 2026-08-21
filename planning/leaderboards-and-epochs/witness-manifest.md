# Leaderboards AC1/AC3/AC5/AC6 witness manifest

Predeclared 2026-08-21 against `8b8d13b`. This range is limited to the already accepted
Leaderboards backend. It does not implement or accept the draft reader/player RFC.

| Criterion | Population and method | Required success | Negative/discrimination probe | Limitation |
|---|---|---|---|---|
| AC1 exact ranking | Two distinct positive Decimal source values below the 12-significant-digit wire quantum, converted independently through canonical wire strings and `MagnitudeKey`, then inserted and queried on the same epoch/category/variables board. | Source values differ; canonical strings and keys converge; SQL returns a shared competition rank. | Alter one derived key after conversion; the shared-rank assertion must fail. | Proves the accepted server key/query boundary, not a future HTTP reader. |
| AC3 epoch crossing | The existing real-Postgres Prestige run starts pinned in epoch N, genuinely changed catalog bytes mint N+1, and the run exits afterward. The same fixture reconstructs the persisted immutable log, replays it against N (including the N+1 Exit successor catalog), projects it, and queries its row only in N. | Replay verdict is `verified`; projection uses the stored N pin; N contains the row and N+1 does not. | Replay the same log with N+1 as its run catalog or project it under N+1; the verdict/epoch assertion must fail. | Uses the backend repository boundary; public retrieval remains successor work. |
| AC5 structural assistance | Three real-Postgres projector runs: join→Exit, join→leave→Exit, and never-joined→Exit. Membership is derived from immutable run events, not terminal current-state arithmetic. | First two project `commons=true`; the control projects `commons=false`; each occupies the structurally distinct board tuple. | Remove `compact_signed` from the scanner; both positive cases must fail while the Solo control remains green. | Advisor semantics are unchanged and out of this repair. |
| AC6 frozen history | Project a board row in epoch N, mint distinct N+1 and N+2 catalog identities, then query N again through the repository reader. | The original row, rank, key and world-first bit remain byte-for-byte equivalent after both mints. | Query N+2 or rewrite the expected epoch; the original-row assertion must fail. | Public/browser historical navigation remains successor work. |

Every claimed gate runs cold (`-count=1`). Postgres witnesses run through the repository Compose
lane. A plan checkbox may move only in the later record commit that includes these tests in the
same designated-review range.
