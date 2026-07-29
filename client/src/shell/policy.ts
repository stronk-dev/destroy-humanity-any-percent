export const CLIENT_SHELL_SCHEMA_VERSION = 1;

export interface ClientShellPolicy {
  readonly tickMs: 50;
  readonly snapshotMs: number;
  readonly catchupCeilingMs: number;
  readonly epsilonLerpPpm: number;
  readonly lerpDurationMs: number;
  readonly reconnectStoryThresholdMs: number;
  readonly reducedMotionRenderMs: number;
}

const keys = ["schema_version", "tick_ms", "snapshot_ms", "catchup_ceiling_ms", "epsilon_lerp_ppm", "lerp_duration_ms", "reconnect_story_threshold_ms", "reduced_motion_render_ms"] as const;

export function parseClientShellPolicy(source: unknown): ClientShellPolicy {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError("client shell policy must be an object");
  const raw = source as Record<string, unknown>;
  const actual = Object.keys(raw).sort(); const expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError("client shell policy fields are not exact");
  const integer = (key: string, minimum: number, maximum: number): number => {
    const value = raw[key];
    if (typeof value !== "number" || !Number.isSafeInteger(value) || value < minimum || value > maximum) throw new SyntaxError(`${key} outside integer domain`);
    return value;
  };
  if (raw.schema_version !== CLIENT_SHELL_SCHEMA_VERSION || raw.tick_ms !== 50) throw new SyntaxError("unsupported client shell policy");
  const policy = {
    tickMs: 50 as const,
    snapshotMs: integer("snapshot_ms", 50, 1000),
    catchupCeilingMs: integer("catchup_ceiling_ms", 50, 5000),
    epsilonLerpPpm: integer("epsilon_lerp_ppm", 0, 1_000_000),
    lerpDurationMs: integer("lerp_duration_ms", 1, 400),
    reconnectStoryThresholdMs: integer("reconnect_story_threshold_ms", 5001, Number.MAX_SAFE_INTEGER),
    reducedMotionRenderMs: integer("reduced_motion_render_ms", 500, Number.MAX_SAFE_INTEGER),
  };
  if (policy.snapshotMs % policy.tickMs !== 0 || policy.catchupCeilingMs % policy.tickMs !== 0 || policy.reconnectStoryThresholdMs <= policy.catchupCeilingMs) throw new SyntaxError("invalid client shell policy relationship");
  return Object.freeze(policy);
}
