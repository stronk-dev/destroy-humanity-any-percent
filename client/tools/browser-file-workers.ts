// CI D1's four-CPU model uses three files per engine in the installed runner.
// Keep that same width on larger development VMs instead of multiplying their
// CPU count by every engine. This is not a measured universal safe capacity.
const CI_MODEL_CPUS = 4;

export function browserFileWorkers(availableCPUs: number): number {
  if (!Number.isSafeInteger(availableCPUs) || availableCPUs < 1) {
    throw new Error("browser CPU availability must be a positive safe integer");
  }
  return Math.max(1, Math.min(CI_MODEL_CPUS, availableCPUs) - 1);
}
