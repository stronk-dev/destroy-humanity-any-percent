// Cosmetic Shop v1 (rfc/cosmetic-shop-v1.md §2): the `cosmetics` artifact,
// byte-twin of server/cosmetic/catalog.go. There is no price field by
// construction (I1, N1): unknown keys at any level reject the artifact.

export const COSMETIC_SCHEMA_VERSION = 1 as const;
export const COSMETIC_SLOT_PET = "pet" as const;
export const COSMETIC_UNLOCK_TIER = "active_company_tier_at_least" as const;

export interface CosmeticItem {
  readonly cosmetic_id: string;
  readonly slot: typeof COSMETIC_SLOT_PET;
  readonly unlock: { readonly kind: typeof COSMETIC_UNLOCK_TIER; readonly tier: number };
}
export interface CosmeticCatalog { readonly items: readonly CosmeticItem[] }

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u;
const MAX_TIER = 8;
const MAX_ITEMS = 256;

export function parseCosmeticCatalog(source: unknown): CosmeticCatalog {
  const root = exactObject(source, ["schema_version", "items"], "cosmetics");
  if (root.schema_version !== COSMETIC_SCHEMA_VERSION) throw new SyntaxError("invalid cosmetics schema_version");
  if (!Array.isArray(root.items) || root.items.length === 0 || root.items.length > MAX_ITEMS) throw new SyntaxError("cosmetics items must be a non-empty bounded array");
  let prior: string | undefined;
  const items = root.items.map((raw): CosmeticItem => {
    const row = exactObject(raw, ["cosmetic_id", "slot", "unlock"], "cosmetic item");
    if (typeof row.cosmetic_id !== "string" || !mechanical.test(row.cosmetic_id)) throw new SyntaxError("cosmetic_id is not mechanical");
    if (prior !== undefined && byteCompare(prior, row.cosmetic_id) >= 0) throw new SyntaxError("cosmetic ids must be unique and byte-sorted");
    prior = row.cosmetic_id;
    if (row.slot !== COSMETIC_SLOT_PET) throw new SyntaxError("unknown cosmetic slot");
    const unlock = exactObject(row.unlock, ["kind", "tier"], "cosmetic unlock");
    if (unlock.kind !== COSMETIC_UNLOCK_TIER) throw new SyntaxError("unknown cosmetic unlock kind");
    if (typeof unlock.tier !== "number" || !Number.isSafeInteger(unlock.tier) || unlock.tier < 0 || unlock.tier > MAX_TIER) throw new SyntaxError("cosmetic unlock tier is outside [0,8]");
    return Object.freeze({ cosmetic_id: row.cosmetic_id, slot: COSMETIC_SLOT_PET, unlock: Object.freeze({ kind: COSMETIC_UNLOCK_TIER, tier: unlock.tier }) });
  });
  return Object.freeze({ items: Object.freeze(items) });
}

/** Parses raw artifact bytes, rejecting duplicate keys JSON.parse would collapse. */
export function loadCosmeticCatalog(bytes: string): CosmeticCatalog {
  if (!uniqueKeys(bytes)) throw new SyntaxError("cosmetics artifact has duplicate keys");
  return parseCosmeticCatalog(JSON.parse(bytes));
}

export function cosmeticItem(catalog: CosmeticCatalog, id: string): CosmeticItem | undefined {
  return catalog.items.find((item) => item.cosmetic_id === id);
}

/** §2 permanent-ID rule: every current id survives with its slot. */
export function validateCosmeticTransition(current: CosmeticCatalog | undefined, next: CosmeticCatalog | undefined): void {
  if (current === undefined) return;
  if (next === undefined) throw new RangeError("cosmetics cannot disappear between epochs");
  for (const item of current.items) {
    const successor = cosmeticItem(next, item.cosmetic_id);
    if (successor === undefined) throw new RangeError(`cosmetic ${item.cosmetic_id} was dropped`);
    if (successor.slot !== item.slot) throw new RangeError(`cosmetic ${item.cosmetic_id} changed slot`);
  }
}

function exactObject(value: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError(`${label} must be an object`);
  const actual = Object.keys(value).sort(byteCompare);
  const expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} keys are not exact`);
  return value as Record<string, unknown>;
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left), b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index += 1) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}

// Recursive-descent key-uniqueness check over JSON text (the Go loader
// rejects duplicate keys; JSON.parse would silently keep the last one).
// Structural errors return false; JSON.parse then decides validity.
function uniqueKeys(text: string): boolean {
  let index = 0;
  const skipSpace = () => { while (index < text.length && " \t\n\r".includes(text[index]!)) index += 1; };
  const readString = (): string | undefined => {
    if (text[index] !== "\"") return undefined;
    const start = index;
    index += 1;
    while (index < text.length && text[index] !== "\"") index += text[index] === "\\" ? 2 : 1;
    if (index >= text.length) return undefined;
    index += 1;
    try { return JSON.parse(text.slice(start, index)) as string; } catch { return undefined; }
  };
  const value = (): boolean => {
    skipSpace();
    const char = text[index];
    if (char === "{") {
      index += 1;
      const seen = new Set<string>();
      skipSpace();
      if (text[index] === "}") { index += 1; return true; }
      for (;;) {
        skipSpace();
        const key = readString();
        if (key === undefined || seen.has(key)) return false;
        seen.add(key);
        skipSpace();
        if (text[index] !== ":") return false;
        index += 1;
        if (!value()) return false;
        skipSpace();
        if (text[index] === ",") { index += 1; continue; }
        if (text[index] === "}") { index += 1; return true; }
        return false;
      }
    }
    if (char === "[") {
      index += 1;
      skipSpace();
      if (text[index] === "]") { index += 1; return true; }
      for (;;) {
        if (!value()) return false;
        skipSpace();
        if (text[index] === ",") { index += 1; continue; }
        if (text[index] === "]") { index += 1; return true; }
        return false;
      }
    }
    if (char === "\"") return readString() !== undefined;
    const match = /^(?:true|false|null|-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?)/u.exec(text.slice(index));
    if (!match) return false;
    index += match[0].length;
    return true;
  };
  if (!value()) return false;
  skipSpace();
  return index === text.length;
}
