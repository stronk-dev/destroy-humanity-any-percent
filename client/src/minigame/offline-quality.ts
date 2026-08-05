import { MAX_EXACT_INTEGER } from "../numeric";

export interface OfflineQualityGrade {
  readonly score_threshold: number;
  readonly grade_ppm: number;
}

export interface OfflineQualityPolicy {
  readonly score_fact: string;
  readonly grade_curve: readonly OfflineQualityGrade[];
  readonly decay_grid_ms: number;
  readonly decay_ppm_per_grid: number;
  readonly neutral_floor_ppm: number;
  readonly automation_destination: string;
}

export interface OfflineQualityState {
  readonly grade_ppm: number;
  readonly last_founder_attended_ms: number;
  readonly decay_remainder_ppm: number;
}

export interface OfflineQualityDeclarations {
  readonly score_fact_ids: ReadonlySet<string>;
  readonly automation_destinations: ReadonlySet<string>;
}

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export function parseOfflineQualityPolicy(source: unknown, declarations: OfflineQualityDeclarations): OfflineQualityPolicy {
  const root = exactObject(source, ["score_fact", "grade_curve", "decay_grid_ms", "decay_ppm_per_grid", "neutral_floor_ppm", "automation_destination"], "offline quality policy");
  if (typeof root.score_fact !== "string" || !mechanical.test(root.score_fact) || !declarations.score_fact_ids.has(root.score_fact) ||
    typeof root.automation_destination !== "string" || !mechanical.test(root.automation_destination) || !declarations.automation_destinations.has(root.automation_destination) ||
    !Array.isArray(root.grade_curve) || root.grade_curve.length === 0) {
    throw new SyntaxError("invalid offline quality declarations");
  }
  const neutralFloor = safeInteger(root.neutral_floor_ppm, 0, 1_000_000);
  let priorThreshold = -1;
  let priorGrade = neutralFloor;
  const curve = root.grade_curve.map((source, index) => {
    const row = exactObject(source, ["score_threshold", "grade_ppm"], "offline quality grade");
    const threshold = safeInteger(row.score_threshold, 0, MAX_EXACT_INTEGER);
    const grade = safeInteger(row.grade_ppm, neutralFloor, 1_000_000);
    if (threshold <= priorThreshold || grade < priorGrade || index === 0 && grade !== neutralFloor) throw new SyntaxError("noncanonical offline quality curve");
    priorThreshold = threshold;
    priorGrade = grade;
    return { score_threshold: threshold, grade_ppm: grade };
  });
  return {
    score_fact: root.score_fact,
    grade_curve: curve,
    decay_grid_ms: safeInteger(root.decay_grid_ms, 1, MAX_EXACT_INTEGER),
    decay_ppm_per_grid: safeInteger(root.decay_ppm_per_grid, 0, 1_000_000),
    neutral_floor_ppm: neutralFloor,
    automation_destination: root.automation_destination,
  };
}

export function offlineGradeForScore(policy: OfflineQualityPolicy, score: number): number {
  safeInteger(score, 0, MAX_EXACT_INTEGER);
  let grade = policy.neutral_floor_ppm;
  for (const row of policy.grade_curve) {
    if (score < row.score_threshold) break;
    grade = row.grade_ppm;
  }
  return safeInteger(grade, policy.neutral_floor_ppm, 1_000_000);
}

export function parseOfflineQualityState(source: unknown): OfflineQualityState {
  const row = exactObject(source, ["grade_ppm", "last_founder_attended_ms", "decay_remainder_ppm"], "offline quality state");
  return {
    grade_ppm: safeInteger(row.grade_ppm, 0, 1_000_000),
    last_founder_attended_ms: safeInteger(row.last_founder_attended_ms, 0, MAX_EXACT_INTEGER),
    decay_remainder_ppm: safeInteger(row.decay_remainder_ppm, 0, 999_999),
  };
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const value = source as Record<string, unknown>;
  const actual = Object.keys(value).sort(byteCompare);
  const expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
  return value;
}

function safeInteger(value: unknown, minimum: number, maximum: number): number {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < minimum || value > maximum) throw new SyntaxError("integer outside offline quality domain");
  return value;
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
