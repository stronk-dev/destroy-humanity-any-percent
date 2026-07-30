# Combat Shared Kernel

The implemented combat foundation currently owns exact arithmetic and deterministic random streams;
it does not yet claim a content catalog or either battle engine.

The six Temperaments are ordered lazy, playful, curious, sassy, shy, chaotic. An attacker has
advantage over the next two entries, disadvantage against the previous two, and a neutral result
against itself and the opposite entry. The chart result is an enum, never a floating multiplier.

Damage is calculated identically in Go and TypeScript using integer stages:

1. base power multiplied by attacker ATK and floor-divided by 64;
2. advantage multiplied by 13/10 or disadvantage by 10/13, multiply before floor-divide;
3. a critical result multiplied by 3/2 and floor-divided;
4. positive base power receives minimum damage 1;
5. storage saturates to signed int32.

Every intermediate is int64/BigInt. HP and stamina use explicit integer clamp operations. Native
division is absent from `client/src/combat`; the shared `idiv` helper is outside that directory and a
fail-closed source gate proves it rejects a seeded direct-division expression on every verification
run.

SplitMix64 and unbiased rejection sampling have one Go authority in `server/determinism`; the
balance harness aliases it rather than maintaining a copy. A battle seed is the first SplitMix64
draw from the match seed. Independent consumers initialize from `battle_seed XOR fnv1a64(label)`,
so adding a random consumer cannot shift crit, obedience, or bot-policy draws. TypeScript implements
the same uint64 wrap semantics with BigInt. Shared vectors cover damage ordering, minimum damage,
saturation, battle seed, and all three registered Phase-0 substream labels.

The strict combat catalog and Trust/Soul input tables remain unimplemented because their active RFC
does not yet enumerate the promised closed effect union or literal piecewise table points. Those are
recorded as DESIGN-GAPs in the active planning log rather than represented here as shipped behavior.
