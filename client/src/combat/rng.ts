const MASK = (1n << 64n) - 1n;
const INCREMENT = 0x9e3779b97f4a7c15n;

export class SplitMix64 {
  #state: bigint;
  constructor(seed: bigint) { this.#state = seed & MASK; }
  next(): bigint {
    this.#state = (this.#state + INCREMENT) & MASK;
    let value = this.#state;
    value = ((value ^ (value >> 30n)) * 0xbf58476d1ce4e5b9n) & MASK;
    value = ((value ^ (value >> 27n)) * 0x94d049bb133111ebn) & MASK;
    return (value ^ (value >> 31n)) & MASK;
  }
  bound(bound: bigint): bigint {
    if (bound <= 0n || bound > MASK) throw new RangeError("bound outside uint64");
    const threshold = ((1n << 64n) - bound) % bound;
    for (;;) { const draw = this.next(); if (draw >= threshold) return draw % bound; }
  }
}

export function battleSeed(matchSeed: bigint): bigint { return new SplitMix64(matchSeed).next(); }
export function substream(seed: bigint, label: string): SplitMix64 {
  if (label.length === 0) throw new RangeError("substream label is empty");
  return new SplitMix64(seed ^ fnv1a64(label));
}

export function fnv1a64(value: string): bigint {
  let hash = 0xcbf29ce484222325n;
  for (const byte of new TextEncoder().encode(value)) { hash ^= BigInt(byte); hash = (hash * 0x100000001b3n) & MASK; }
  return hash;
}
