// Cosmetic Shop v1 §3: Founder save v24 `cosmetics`, byte-twin of
// server/cosmetic/state.go. Empty collections are canonical [] / {}.
import { cosmeticItem, type CosmeticCatalog } from "./catalog";

export interface CosmeticState { owned: string[]; equipped: Record<string, string> }

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u;

export function emptyCosmeticState(): CosmeticState { return { owned: [], equipped: {} }; }

export function cloneCosmeticState(state: CosmeticState): CosmeticState {
  return { owned: [...state.owned], equipped: { ...state.equipped } };
}

/** Canonical wire form: owned as stored, equipped with byte-sorted keys. */
export function encodeCosmeticState(state: CosmeticState): CosmeticState {
  const equipped: Record<string, string> = {};
  for (const pet of Object.keys(state.equipped).sort(byteCompare)) equipped[pet] = state.equipped[pet]!;
  return { owned: [...state.owned], equipped };
}

export function cosmeticStatesEqual(left: CosmeticState, right: CosmeticState): boolean {
  return JSON.stringify(encodeCosmeticState(left)) === JSON.stringify(encodeCosmeticState(right));
}

/** §3 shape + catalog validation; pets is the Founder pets key set. */
export function parseCosmeticState(source: unknown, pets: ReadonlySet<string>, catalog: CosmeticCatalog): CosmeticState {
  if (source === null || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError("cosmetics must be an object");
  const raw = source as Record<string, unknown>;
  const keys = Object.keys(raw).sort();
  if (keys.length !== 2 || keys[0] !== "equipped" || keys[1] !== "owned") throw new SyntaxError("cosmetics keys are not exact");
  if (!Array.isArray(raw.owned)) throw new SyntaxError("cosmetics.owned must be an array");
  if (raw.equipped === null || typeof raw.equipped !== "object" || Array.isArray(raw.equipped)) throw new SyntaxError("cosmetics.equipped must be an object");
  const owned: string[] = [];
  for (const id of raw.owned) {
    if (typeof id !== "string" || !mechanical.test(id) || owned.length > 0 && byteCompare(owned[owned.length - 1]!, id) >= 0) throw new SyntaxError("cosmetics.owned must be mechanical, unique, and byte-sorted");
    if (cosmeticItem(catalog, id) === undefined) throw new SyntaxError(`owned cosmetic ${id} is not pinned`);
    owned.push(id);
  }
  if (owned.length > catalog.items.length) throw new SyntaxError("cosmetics.owned exceeds the catalog");
  const equipped: Record<string, string> = {};
  for (const [pet, id] of Object.entries(raw.equipped as Record<string, unknown>)) {
    if (!pets.has(pet)) throw new SyntaxError(`equipped key ${pet} is not a pet`);
    if (typeof id !== "string" || !owned.includes(id)) throw new SyntaxError(`pet ${pet} wears an unowned cosmetic`);
    if (cosmeticItem(catalog, id)?.slot !== "pet") throw new SyntaxError(`pet ${pet} wears a non-pet cosmetic`);
    equipped[pet] = id;
  }
  return { owned, equipped };
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left), b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index += 1) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
