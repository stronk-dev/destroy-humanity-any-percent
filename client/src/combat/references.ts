import { parseIntegerCatalogJSON } from "../catalog-json";

/** C2's reference envelope only. Supply IDs from the validated pinned catalog;
 * ownership, instance selection and engine-specific legality are not inferred. */
export function decodeTeamReferences(source: string, declarations: ReadonlySet<string>): readonly [string, string, string] {
  return decodeReferences(source, declarations, 3) as readonly [string, string, string];
}

/** Preserve all eight slots: no sorting, deduplication or cycle-rule decisions. */
export function decodeDeckReferences(source: string, declarations: ReadonlySet<string>): readonly [string, string, string, string, string, string, string, string] {
  return decodeReferences(source, declarations, 8) as readonly [string, string, string, string, string, string, string, string];
}

function decodeReferences(source: string, declarations: ReadonlySet<string>, count: number): readonly string[] {
  const value = parseIntegerCatalogJSON(source, "combat loadout");
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError("invalid combat loadout references");
  const fields = value as Record<string, unknown>;
  if (Object.keys(fields).sort().join("\0") !== "ids\0v" || fields.v !== 1 || !Array.isArray(fields.ids) || fields.ids.length !== count) {
    throw new SyntaxError("invalid combat loadout references");
  }
  const ids: string[] = [];
  for (const id of fields.ids) {
    if (typeof id !== "string" || !wellFormedCombatString(id) || !declarations.has(id)) throw new SyntaxError("invalid combat loadout references");
    ids.push(id);
  }
  return Object.freeze(ids);
}

export function wellFormedCombatString(value: string): boolean {
  for (let index = 0; index < value.length; index++) {
    const point = value.codePointAt(index)!;
    if (point >= 0xd800 && point <= 0xdfff) return false;
    if (point > 0xffff) index++;
  }
  return true;
}
