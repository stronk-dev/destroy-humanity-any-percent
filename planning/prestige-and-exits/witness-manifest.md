# Prestige AC2–AC6 witness manifest

Predeclared 2026-08-21 for execution-queue row 2b. This batch may change tests, checked-in test
fixtures, canonical verification status, and planning records only. It may not change production
mechanics, balance artifacts, player copy, schemas, or CI workflows.

## Populations and oracles

| AC | Population | Required oracle | Negative/discrimination control |
|---|---|---|---|
| AC2 | Non-expired offer ages `0`, `1`, `duration/2`, and `duration-1` ms; each run progresses from its stored preview state before acceptance. One exact-`duration` row is the expiry boundary. | Every accepted payout field is at least its stored preview and Network slots are a superset; the boundary row rejects as expired without advancing Founder rewards. | The comparison helper must reject a forged payout one unit below preview; a production probe severs `PromiseTerms` to current-only. |
| AC3 formula | Every Notoriety integer `0..100`, plus deterministic samples spanning `101..MaxExactInteger` and invalid `-1` / `MaxExactInteger+1` inputs. | Exact equality to an independent integer reference for `clamp(90-floor(35*n/100),55,90)`; invalid inputs reject. | An off-by-one reference result must be rejected by the equality oracle. |
| AC3 ledger | Two successive real-Postgres Exits of different kinds with non-empty old Founder facts and distinct non-empty Company facts on each run. | Founder facts after each Exit equal the cumulative set; previously recorded fact keys and values are unchanged. | A production probe removes the Company→Founder fact union on the second Exit. |
| AC4 | Real account/API lifecycle: first Founder completes an Exit, `POST /api/v1/founder` creates and activates a distinct Founder, that Founder reaches the current curriculum boundary and its next Company command is replaced by `scripted_first`. | Old Founder remains archived with its history; new Founder begins with empty history, receives exactly one `scripted_first`, and advances to run 2. | A production probe carries old Exit history into `NewFounder`, preventing the fresh scripted transition. |
| AC5 | Eligible Tier-1 rows: plain, pending active-play opportunity, active buff, incorporated/faction stock, and live Exit offer. One Tier-0 row is ineligible. | Every eligible `wind_down` applies atomically and advances run sequence; the Tier-0 control returns typed `not_eligible` without either stream advancing. | A production probe rejects the pending-opportunity row solely because its event state is active. |
| AC6 | One declared Founder/prior-Company input with non-empty carry/reset-sensitive fields, evaluated twice against one checked-in complete encoded-state golden. | Both encodings are byte-identical to the checked-in golden. | In-memory removal of one carried Guild boundary and mutation of run sequence must each differ from the golden; a production probe drops the carry assignment. |

## Measurement limits

- The offer-age population proves non-expired boundary handling and payout monotonicity for the
  declared representative ages; it does not estimate real player click timing.
- The eligible-state matrix is the closed set of currently persisted event/lifecycle axes known to
  coexist with Wind Down. A later mechanic that adds a trapping state must extend the matrix.
- The New-Founder row proves server/account composition. It does not claim a player-facing
  New-Founder UI exists.
- These witnesses can close AC2–AC6's proof gaps only after cold execution and the mandatory
  cross-party exact-range review. They do not by themselves authorize archival.
