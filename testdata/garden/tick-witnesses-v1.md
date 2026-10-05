# Independent Garden tick witnesses

This is test evidence for accepted RFC SG4, not balance, production content or an alternate
implementation specification. The JSON's sixteen expected transitions were derived from the
RFC, not copied from `Advance`, `advanceGarden` or the Go-generated engine corpus.

Plot tuples are `[row, col, species_id, age_ticks, matured_effect_ppm]`; maturation tuples
are `[row, col, species_id]`; spawn tuples add `recipe_id`. Cases replace the fixture's five
recipe chances in canonical order and, when specified, A's maturation threshold or bare-metal's
chance factor. Every scenario must pass the actual raw catalog loader and SG2 state codec.
Readers reconstruct full expected state from unchanged initial metadata, the literal post-plots,
and the declared one/two complete ticks. They never compute expected growth or recipe selection.

For tick 1, eligible plots consume `21447, 266542, 710139, 814499, 415932, …` in row-major order.
For tick 2, they consume `400727, 763831, 9588, …`.

- A at `(0,2)` in a 3×3 grid makes only `(0,1)`, `(1,1)`, `(1,2)` eligible. At chance 300,000,
  only the first two spawn. Earlier ineligible cells must consume no draw.
- B at `(0,0)` and A at `(2,2)` make five cells eligible: `(0,1)`, `(1,0)`, `(1,1)`, `(1,2)`,
  `(2,1)`. Factor 500,000 and B chance 1 produce effective zero, but the first two cells still
  consume draws. A chance 1M has effective 500,000; only the fifth cell's 415,932 selects A.
- Two mature parents in opposite corners of a 2×2 grid share the two empty cells. At tick 1,
  A's first canonical band 300,000 captures both draws, even with B chance 1M. At tick 2 under
  Chaos, A's effective band is 300,000 and B's is 900,000: both draws select B. Their cumulative
  total saturates. Removing just the upper cumulative clamp cannot change selection for draws
  <1M and nonnegative bands, so that internal mutation is not claimed as distinguishable.
- Chance 21,447 rejects the equal first draw; chance 21,448 accepts it. Tick 2 with A chance
  300,000 selects only the third eligible cell, unlike tick 1.
- All spawned plots are age zero in that tick; after a second tick they are age one. The parent
  that matured on the first tick retains its frozen effect and age. Neither spreading nor
  mutation collects seeds; the full-state assertion retains only the two starters.

Reproduce the arithmetic draw literals without importing production code:

```js
const mask = (1n << 64n) - 1n;
function hash(label) {
  let h = 14695981039346656037n;
  for (const b of new TextEncoder().encode(label)) h = ((h ^ BigInt(b)) * 1099511628211n) & mask;
  return h;
}
function next(state) {
  state = (state + 11400714819323198485n) & mask;
  let z = ((state ^ (state >> 30n)) * 13787848793156543929n) & mask;
  z = ((z ^ (z >> 27n)) * 10723151780598845931n) & mask;
  return [state, (z ^ (z >> 31n)) & mask];
}
if (next(0n)[1] !== 0xe220a8397b1dcdafn) throw Error("seed-zero control");
const base = next(0x0123456789abcdefn ^ hash("garden.founder.v1"))[1];
console.log(base.toString()); // 7474842082278028174
for (const tick of [1n, 2n]) {
  let state = base ^ tick ^ hash("garden.tick.v1");
  const draws = [];
  while (draws.length < 8) {
    const pair = next(state); state = pair[0];
    if (pair[1] >= (1n << 64n) % 1000000n) draws.push(Number(pair[1] % 1000000n));
  }
  console.log(Number(tick), draws);
}
```

These selected literal values are not statistical evidence or an exhaustive rejection-sampling
test. See `planning/minigame-server-garden/log.md` for the initial invalid instrument, corrective
predeclaration, actual-source mutations, restored verification and pending review provenance.
