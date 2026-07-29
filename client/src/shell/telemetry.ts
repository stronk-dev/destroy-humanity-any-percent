import type { IntentReceipt } from "./contracts";
import type { WorkerOutput } from "./worker-protocol";

export interface ClientTelemetrySnapshot {
  readonly epsilonLerpBreaches: number;
  readonly rejectionRebases: Readonly<Record<string, number>>;
  readonly discreteRebases: number;
  readonly workerTickOverruns: number;
  readonly workerTickOverrunMs: number;
}

export class ClientTelemetry {
  #epsilonLerpBreaches = 0;
  #discreteRebases = 0;
  #workerTickOverruns = 0;
  #workerTickOverrunMs = 0;
  readonly #rejectionRebases = new Map<string, number>();

  recordContinuousRebase(receipt?: IntentReceipt): void {
    if (!receipt || receipt.status !== "rejected") this.#epsilonLerpBreaches++;
  }

  recordDiscreteRebase(): void {
    this.#discreteRebases++;
  }

  recordReceipt(receipt?: IntentReceipt): void {
    if (receipt?.status !== "rejected") return;
    const category = receipt.rejectionCode ?? "rejected";
    this.#rejectionRebases.set(category, (this.#rejectionRebases.get(category) ?? 0) + 1);
  }

  recordWorker(output: WorkerOutput): void {
    if (output.kind !== "worker_metric" || output.name !== "tick_overrun") return;
    this.#workerTickOverruns++;
    this.#workerTickOverrunMs += output.value;
  }

  snapshot(): ClientTelemetrySnapshot {
    return {
      epsilonLerpBreaches: this.#epsilonLerpBreaches,
      rejectionRebases: Object.fromEntries([...this.#rejectionRebases].sort(([left], [right]) => left < right ? -1 : left > right ? 1 : 0)),
      discreteRebases: this.#discreteRebases,
      workerTickOverruns: this.#workerTickOverruns,
      workerTickOverrunMs: this.#workerTickOverrunMs,
    };
  }
}
