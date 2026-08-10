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
  if (observation.inputs !== GAME_UI_PERFORMANCE_BUDGET.inputCount || observation.formattedCommits > GAME_UI_PERFORMANCE_BUDGET.formattedCommitsMaximum ||
      observation.longestTaskMS > GAME_UI_PERFORMANCE_BUDGET.longTaskCeilingMS) throw new RangeError("Game UI performance budget exceeded");
}
