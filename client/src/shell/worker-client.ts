import type { AuthoritativeSnapshot } from "./contracts";
import type { ClientShellPolicy } from "./policy";
import type { WorkerCommand, WorkerOutput } from "./worker-protocol";

export class PredictionWorkerClient {
  readonly #worker: Worker;
  readonly #clock: () => number;
  #disposed = false;
  constructor(policy: ClientShellPolicy, snapshot: AuthoritativeSnapshot, consume: (output: WorkerOutput) => void, clock: () => number = () => performance.now()) {
    this.#worker = new Worker(new URL("./prediction.worker.ts", import.meta.url), { type: "module" });
    this.#clock = clock;
    this.#worker.onmessage = (event: MessageEvent<WorkerOutput>) => consume(event.data);
    this.#worker.postMessage({ kind: "initialize", policy, snapshot, monotonicMs: clock() } satisfies WorkerCommand);
  }
  authoritative(snapshot: AuthoritativeSnapshot): void { if (!this.#disposed) this.#worker.postMessage({ kind: "authoritative_snapshot", snapshot, monotonicMs: this.#clock() } satisfies WorkerCommand); }
  dispose(): void { if (this.#disposed) return; this.#disposed = true; this.#worker.postMessage({ kind: "dispose" } satisfies WorkerCommand); this.#worker.terminate(); }
}
