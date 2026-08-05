const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export interface MinigameCatalog {
  readonly minigameIds: readonly string[];
  readonly ratingSeasons: readonly string[];
}

export function parseMinigameCatalog(source: unknown): MinigameCatalog {
  const root = exactObject(source, ["schema_version", "minigame_ids", "rating_seasons"], "minigame catalog");
  if (root.schema_version !== 1) throw new SyntaxError("invalid minigame catalog version");
  return Object.freeze({ minigameIds: sortedMechanical(root.minigame_ids, "minigame ids"), ratingSeasons: sortedMechanical(root.rating_seasons, "rating seasons") });
}

function sortedMechanical(source: unknown, label: string): readonly string[] {
  if (!Array.isArray(source)) throw new SyntaxError(`${label} must be an array`);
  let prior = "";
  return Object.freeze(source.map((item) => {
    if (typeof item !== "string" || !mechanical.test(item) || byteCompare(prior, item) >= 0) throw new SyntaxError(`invalid ${label}`);
    prior = item; return item;
  }));
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const value = source as Record<string, unknown>;
  const actual = Object.keys(value).sort(byteCompare); const expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
  return value;
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
