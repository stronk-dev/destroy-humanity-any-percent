# Gameserver Composition — append-only log

## 2026-08-02 — implementation started

Review by: not applicable (owner acceptance). Recorded by: Codex.

The owner assigned the complete RFC after ratifying Wire v2 and providing GC1-GC3. The work is
accepted as one implementation round. Composition must use real owner-backed services; a missing
callable owner surface is added to that owner package with proof, never hidden behind a no-op
driver. No push is authorized.

## 2026-08-02 — GC1 implemented

Review by: Codex implementer self-review. Recorded by: Codex.

The production resolver now receives the exact Company stream, Founder, pinned constants hash,
and request context. The Commons projection returns the current run's committed sample; before
that sample it applies the published integer entry-weight formula to the projected membership's
declared tithe. `commons_member_samples.run_seq` prevents a prior run's sample from leaking into a
fresh membership. The production integration now uses this real projector, proves first accrual,
then replays that intent byte-identically. Resolver failures wrap `ErrInvalidEngineState` and the
HTTP boundary exposes the typed `internal_invariant` category. Focused Go and real-Postgres suites
passed. Kernel version 0.3.2 records the live resolved-input semantic change.

## 2026-08-02 — GC2 implemented

Review by: Codex implementer self-review. Recorded by: Codex.

Go and TypeScript now validate the exact world state recursively, including safe integer/ppm
bounds, revision binding, epoch identity, and the null-milestone/zero-progress biconditional. The
shared wire corpus carries the canonical snapshot plus unknown-field and revision negatives. The
gameserver world aggregator owns a serialized revision and advances it only after publish success;
a 1,000-snapshot soak proves strict monotonicity and the declared zero values for unshipped planet
and milestone systems. Transport, gameserver, all 6,494 active client tests, and TypeScript/Svelte
checks passed.
