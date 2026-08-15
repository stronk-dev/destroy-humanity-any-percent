# `first_hour_policy.v1` — draft for owner ratification (T01-C34)

Candidate bytes: `balance/testdata/t0-t1/first-hour-policy-v1.json`.
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

Every draw derives from `SHA-256` over `policy_id`, `policy_version`, `seed`, `run_seq`, `decision_ordinal` joined by the unit separator `0x1f`,,
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
