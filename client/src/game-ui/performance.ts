export const GAME_UI_PERFORMANCE_BUDGET = Object.freeze({
  cpuThrottle: 4,
  droppedFrameAllowancePPM: 50_000,
  durationMS: 60_000,
  formattedCommitsMaximum: 600,
  inputCount: 1_200,
  longTaskCeilingMS: 200,
  viewport: Object.freeze({ height: 720, width: 1280 }),
});

export function validatePerformanceObservation(observation: Readonly<{ formattedCommits: number; inputs: number; longestTaskMS: number }>): void {
  const failures: string[] = [];
  if (observation.inputs !== GAME_UI_PERFORMANCE_BUDGET.inputCount) {
    failures.push(`inputs=${observation.inputs}, expected=${GAME_UI_PERFORMANCE_BUDGET.inputCount}`);
  }
  if (observation.formattedCommits > GAME_UI_PERFORMANCE_BUDGET.formattedCommitsMaximum) {
    failures.push(`formatted_commits=${observation.formattedCommits}, maximum=${GAME_UI_PERFORMANCE_BUDGET.formattedCommitsMaximum}`);
  }
  if (observation.longestTaskMS > GAME_UI_PERFORMANCE_BUDGET.longTaskCeilingMS) {
    failures.push(`longest_task_ms=${observation.longestTaskMS}, maximum=${GAME_UI_PERFORMANCE_BUDGET.longTaskCeilingMS}`);
  }
  if (failures.length > 0) throw new RangeError(`Game UI performance budget exceeded: ${failures.join("; ")}`);
}
