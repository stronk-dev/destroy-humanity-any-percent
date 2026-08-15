# `first_hour_policy.v1` — draft for owner ratification (T01-C34)

Candidate bytes: `balance/testdata/t0-t1/first-hour-policy-v1.json` — **document version 3**,
`sha256:e5e5de7051beb0340e54f7013ce7d4a48c35bfcc3343220310290478445d10c3`. Versions 1
(`c56653c9…021c`) and 2 (`3c5c20bd…cd7c`) are **withdrawn**: Codex's cross-party review (T01-C37–C39) found it
under-specified and quantitatively unable to meet its own envelope. Do not ratify v1.
Drafted by Claude (design lane) per the 2026-08-15 ruling. **Nothing may consume it until ratified.**

## The anti-circularity rule this document exists to enforce

C34's real hazard is that whoever picks these numbers picks the pacing answer. So the causality is
fixed, and it runs one way only:

> **The personas are a model of players, chosen for plausibility. The CONTENT is what gets tuned
> to fit the envelopes.** If an envelope fails, the finding is against the economy, not against
> the persona. Cadence/session literals may be revised only on an owner ruling that says the model
> itself was unrealistic — never to make a gate pass.

Codex retracting its own drafted literals (`6201fa9`) was the correct instinct; this document is
the ratifiable replacement.

## The three players

**`chaos.t0_t1` — the first-session enthusiast.** One unbroken two-hour sitting, an action every
**2 s**, and at each boundary it picks **uniformly at random among whatever is currently legal**
(click / any affordable purchase / wait). No weighting, no cleverness — that is what makes it
chaos, and it means the 30-second first-generator envelope is a genuine test of click yield and
starting price rather than a scripted certainty. 64 seeds.

**`casual.t0_t1` — returns between other things.** Attended in **15-minute windows every 30
minutes** (≈60 min attended inside the 2 h horizon, which is what lets it reach the 45-minute
founder-attended Exit envelope at all), an action every **5 s**, and it always takes the
**cheapest affordable** purchase, else clicks, else waits. Deterministic by construction — so its
32 seeds are made meaningful by **session-start jitter up to 5 minutes**, modelling that people
come back at different moments. Without that jitter the 32 runs would be 32 identical copies.

**`reference.greedy` — the optimal-play upper bound.** Reuses the **shipped T01-C20
projected-time ranker** with banking as a first-class candidate. It authors no second economy
model, which is exactly what C30 required. 1 seed; fully deterministic.

## Determinism

Every draw derives from `SHA-256` over `policy_id`, `policy_version`, `seed`, `run_seq`, `decision_ordinal` joined by the unit separator `0x1f`,
first 8 bytes big-endian, modulo the count of legal commands enumerated in **raw-byte ascending
purchasable-ID order**. No wall clock, no global RNG, no map iteration. Keying on `run_seq` means
run 2 draws a fresh stream rather than replaying run 1's choices.

## Exit rule (the ruled C32 predicate, shared by all three)

Wind down at the first action boundary where: run ≥ 2, `gate.t0_to_t1` crossed, founder-attended
≥ 2,700,000 ms, **and** the previewed terms grant any persistent value. The last clause is what
stops the [45,90] envelope being satisfied by merely waiting — if progression or payout breaks,
the upper bound genuinely fails.

## What this document deliberately does not decide

Starting cash, click yield, generator prices, and the starter-package literals are all content and
live elsewhere. If chaos cannot buy a generator inside 30 s, that is a finding about the T0
economy — which is the whole point of running the personas against it.

## Owner decision requested

Ratify `first-hour-policy-v1.json` as drafted, or name the model changes you want first. On
ratification its hash joins scenario identity and Codex implements C34's runner against it.


## Revision 2 — what Codex's review changed, and the one thing it costs an owner ruling

Codex reviewed v1 and was right on every count. Recorded plainly because this is the cross-party
gate working in the direction it less often runs (Codex reviewing Claude's design):

- **T01-C37 — RNG bytes were incomplete.** v2 pins integer encoding (`decimal_ascii_no_padding`),
  a fixed command-class enumeration (`perform_manual_batch` always index 0, purchasables by
  raw-byte ID within class, `wait` always last) so manual/wait ordering is no longer ambiguous,
  and separate **domain tags** (`decision` vs `jitter`) so the session-jitter draw cannot collide
  with a decision draw on the same material.
- **T01-C38 — casual's session clock was undefined.** v2 pins origin (founder genesis), units
  (wall ms), that sessions span runs and do **not** restart at run end, that inter-session gaps are
  offline spans, and that a session opens by applying offline catchup and then evaluating online —
  which is exactly the AC0 seam this project just built, so casual now exercises it every session.
- **T01-C39 — the ranker had one objective and then stopped.** v2 gives `reference.greedy` an
  ordered objective sequence: aim the shipped T01-C20 ranker at the next uncrossed gate's
  requirement, and after the final gate at the exit-readiness predicate. It is re-aimed at each
  phase boundary and never ranks toward an already-satisfied objective.

### The chaos plausibility correction — this one needs your affirmation

Codex computed that chaos v1 could reach its first generator inside 30 s only 4.4–7.9 % of the
time; I independently get 12.5 %. Either way it cannot produce a **p50** ≤ 30 s, so v1's own
envelope was unreachable by construction.

**Diagnosis: the model was wrong, not the economy.** v1 let chaos draw `wait` with equal weight
while holding zero cash and zero income — but waiting at zero income *cannot change affordability*,
so no real player would ever choose it. A person who has just opened an idle game with nothing to
buy clicks; they do not flip a coin every two seconds between clicking and doing nothing.

v2 therefore adds a general legality rule — **`wait` is legal only when production income is
positive** — which is independently justified rather than tuned to pass. Under it, chaos clicks
until it can afford something, and reaches its first generator by 30 s about **96.9 %** of the
time, median ≈ 22 s.

Per the anti-circularity rule written into this document, a persona may only be changed on an
owner ruling that **the model itself was unrealistic**. That is the claim here, and it is the one
thing in revision 2 that is yours to affirm rather than mine to assert.


## Revision 3 — the jitter in v2 was decorative, and I proved it

Codex's second-round review found three residuals in v2. All three were real; one was not a
detail.

- **Jitter diversity was unproven — and false.** v2 applied ONE offset to every session start,
  which is a pure phase shift: I measured it across offsets and got identical session count,
  identical inter-session gaps, and identical attended totals every time. All 32 casual seeds
  would have produced the same run, which would have made casual's **p95** Exit envelope a
  statistic over 32 copies of one number. v3 replaces it with **independent per-session draws**
  for both start offset and duration, keyed on `(policy_id, seed, session_index)`. Sampled over
  six seeds that now yields six distinct attended totals and varying session counts, so the p95
  measures something.
- **Non-decision draw material was ambiguous.** v3 pins `run_seq = 0` and the session index in
  the ordinal slot for jitter draws, so a jitter stream can never collide with a decision stream;
  and it pins **half-open** range semantics (lower inclusive, upper exclusive) everywhere.
- **First-action timing was unspecified.** v3: the first action of a session occurs **at** the
  session-start boundary, with `action_cadence_ms` from there.
- **Ranker scope was loose.** v3 pins the gate scope (only gates declared by the scenario's
  segments), multi-resource ordering (raw-byte ascending resource ID, largest remaining deficit
  first), the terminal objective (exit-readiness), and that the **C32 exit rule preempts the
  ranker** at any boundary where it is true.

Three rounds of review on one document is worth noting rather than hiding: each round moved from
structure → encoding → byte-level residuals, which is convergence, and one finding per round was
substantive rather than pedantic. A fourth round of *specification* findings would be a signal
that this document is the wrong shape; a fourth round of implementation findings would be normal.
