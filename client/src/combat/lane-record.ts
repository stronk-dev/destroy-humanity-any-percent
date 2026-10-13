import { parseIntegerCatalogJSON } from "../catalog-json";
import { chart, type Temperament } from "./arithmetic";
import { wellFormedCombatString } from "./references";

export type LaneTargets = "units" | "buildings" | "any";
export interface LaneRecord {
  readonly id: string;
  readonly temperament: Temperament;
  readonly cost: number;
  readonly hp: number;
  readonly atk: number;
  readonly spd: number;
  readonly range: number;
  readonly targets: LaneTargets;
  readonly tags: readonly string[];
  readonly bypass?: boolean;
}

/** C2's common structural lane row. No effect execution, stat minimum,
 * cheap-cycle threshold, complete catalog or battle legality is inferred. */
export function decodeLaneRecord(source: string): LaneRecord {
  try {
    const value = parseIntegerCatalogJSON(source, "combat lane record");
    if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError();
    const fields = value as Record<string, unknown>;
    const keys = ["id", "temperament", "cost", "hp", "atk", "spd", "range", "targets", "tags"];
    const hasBypass = Object.hasOwn(fields, "bypass");
    if (Object.keys(fields).length !== keys.length + Number(hasBypass) || keys.some(key => !Object.hasOwn(fields, key))) throw new SyntaxError();
    const temperament = text(fields.temperament) as Temperament;
    chart(temperament, temperament);
    if (fields.targets !== "units" && fields.targets !== "buildings" && fields.targets !== "any") throw new SyntaxError();
    if (!Array.isArray(fields.tags) || hasBypass && typeof fields.bypass !== "boolean") throw new SyntaxError();
    return Object.freeze({
      id: text(fields.id), temperament, cost: int32(fields.cost), hp: int32(fields.hp),
      atk: int32(fields.atk), spd: int32(fields.spd), range: int32(fields.range),
      targets: fields.targets, tags: Object.freeze(fields.tags.map(text)),
      ...(hasBypass ? { bypass: fields.bypass as boolean } : {}),
    });
  } catch {
    throw new SyntaxError("invalid combat lane record");
  }
}

function text(value: unknown): string {
  if (typeof value !== "string" || !wellFormedCombatString(value)) throw new SyntaxError();
  return value;
}

function int32(value: unknown): number {
  if (typeof value !== "number" || !Number.isInteger(value) || value < -2_147_483_648 || value > 2_147_483_647) throw new SyntaxError();
  return value === 0 ? 0 : value;
}
