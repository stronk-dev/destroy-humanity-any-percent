import { chart, type ChartResult, type Temperament } from "./arithmetic";

export type Runtime = "container" | "vm" | "bare_metal" | "serverless" | "edge" | "mainframe";
export interface CombatIdentity {
  readonly temperament: Temperament;
  readonly runtime: Runtime;
}

const runtimes: readonly string[] = Object.freeze(["container", "vm", "bare_metal", "serverless", "edge", "mainframe"]);

/** C1's identity only: no species, tags, stats, catalog or wire envelope. */
export function identity(temperament: unknown, runtime: unknown): CombatIdentity {
  if (typeof temperament !== "string" || typeof runtime !== "string" || !runtimes.includes(runtime)) {
    throw new RangeError("invalid combat identity");
  }
  // Validate through the existing temperament authority, without a second chart.
  chart(temperament as Temperament, temperament as Temperament);
  return Object.freeze({ temperament: temperament as Temperament, runtime: runtime as Runtime });
}

/** Runtime has no Phase-0 chart interaction. Revalidate caller-built objects. */
export function chartIdentities(attacker: CombatIdentity, defender: CombatIdentity): ChartResult {
  const left = identity(attacker.temperament, attacker.runtime);
  const right = identity(defender.temperament, defender.runtime);
  return chart(left.temperament, right.temperament);
}
