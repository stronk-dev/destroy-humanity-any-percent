import { parseIntegerCatalogJSON } from "../catalog-json";
import { identity, type Runtime } from "./identity";
import { type Temperament } from "./arithmetic";
import { wellFormedCombatString } from "./references";

export interface SpeciesStats {
  readonly hp: number;
  readonly atk: number;
  readonly def: number;
  readonly spd: number;
  readonly stamina_max: number;
}

export interface SpeciesRecord {
  readonly id: string;
  readonly temperament: Temperament;
  readonly runtime: Runtime;
  readonly stats: SpeciesStats;
  readonly moves: readonly [string, string, string, string];
  readonly tags: readonly string[];
}

/** C2's structural species record, not a complete catalog or battle policy.
 * Move declarations come from the caller's validated pinned catalog. Preserve
 * slots/tags; do not invent stat minima, effects or a synergy vocabulary. */
export function decodeSpeciesRecord(source: string, moves: ReadonlySet<string>): SpeciesRecord {
  try {
    const fields = object(parseIntegerCatalogJSON(source, "combat species"), ["id", "temperament", "runtime", "stats", "moves", "tags"]);
    const axes = identity(fields.temperament, fields.runtime);
    const stats = object(fields.stats, ["hp", "atk", "def", "spd", "stamina_max"]);
    const moveIDs = strings(fields.moves);
    if (moveIDs.length !== 4 || moveIDs.some(id => !moves.has(id))) throw new SyntaxError();
    return Object.freeze({
      id: text(fields.id), temperament: axes.temperament, runtime: axes.runtime,
      stats: Object.freeze({ hp: int32(stats.hp), atk: int32(stats.atk), def: int32(stats.def),
        spd: int32(stats.spd), stamina_max: int32(stats.stamina_max) }),
      moves: Object.freeze(moveIDs) as readonly [string, string, string, string],
      tags: Object.freeze(strings(fields.tags)),
    });
  } catch {
    throw new SyntaxError("invalid combat species record");
  }
}

function object(value: unknown, keys: readonly string[]): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError();
  const fields = value as Record<string, unknown>;
  if (Object.keys(fields).length !== keys.length || keys.some(key => !Object.hasOwn(fields, key))) throw new SyntaxError();
  return fields;
}

function text(value: unknown): string {
  if (typeof value !== "string" || !wellFormedCombatString(value)) throw new SyntaxError();
  return value;
}

function strings(value: unknown): string[] {
  if (!Array.isArray(value)) throw new SyntaxError();
  return value.map(text);
}

function int32(value: unknown): number {
  if (typeof value !== "number" || !Number.isInteger(value) || value < -2_147_483_648 || value > 2_147_483_647) throw new SyntaxError();
  // Canonical stored integer zero; the raw parser still validates token spelling.
  return value === 0 ? 0 : value;
}
