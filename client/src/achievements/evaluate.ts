import type { AchievementCatalog, AchievementCondition, AchievementDefinition } from "./catalog";
import { MAX_EXACT_INTEGER } from "./catalog";

export interface AchievementObservation {
  readonly facts: ReadonlySet<string>;
  readonly counters: Readonly<Record<string, number>>;
  readonly exitCount: number;
  readonly generators: Readonly<Record<string, number>>;
}

export function achievementEligible(condition: AchievementCondition, observation: AchievementObservation): boolean {
  switch (condition.kind) {
    case "fact_present": return observation.facts.has(condition.factKind);
    case "counter_at_least": return (observation.counters[condition.counter] ?? 0) >= condition.minimum;
    case "exit_count_at_least": return observation.exitCount >= condition.count;
    case "owns_generator_at_least": return (observation.generators[condition.generatorId] ?? 0) >= condition.count;
    case "all_of": return condition.conditions.every((child) => achievementEligible(child, observation));
  }
}

export function newlyEarned(catalog: AchievementCatalog, runEarned: ReadonlySet<string>, lifetimeEarned: ReadonlySet<string>, run: AchievementObservation, career: AchievementObservation): readonly AchievementDefinition[] {
  return Object.freeze(catalog.definitions.filter((definition) => !runEarned.has(definition.id) && !lifetimeEarned.has(definition.id) && achievementEligible(definition.condition, definition.conditionScope === "run" ? run : career)));
}

export function achievementScore(catalog: AchievementCatalog, earned: ReadonlySet<string>): number {
  let result = 0;
  for (const id of earned) {
    const definition = catalog.byId.get(id);
    if (!definition || result > MAX_EXACT_INTEGER - definition.scoreGrant) throw new RangeError("invalid earned achievement set");
    result += definition.scoreGrant;
  }
  return result;
}
