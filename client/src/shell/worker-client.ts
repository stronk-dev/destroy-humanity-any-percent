import type { AuthoritativeSnapshot } from "./contracts";
import type { ClientShellPolicy } from "./policy";
import type { WorkerCommand, WorkerOutput } from "./worker-protocol";

export class PredictionWorkerClient {
  readonly #worker: Worker;
  #disposed = false;
  constructor(policy: ClientShellPolicy, snapshot: AuthoritativeSnapshot, consume: (output: WorkerOutput) => void) {
    this.#worker = new Worker(new URL("./prediction.worker.ts", import.meta.url), { type: "module" });
    this.#worker.onmessage = (event: MessageEvent<WorkerOutput>) => consume(event.data);
    this.#worker.onerror = (event) => {
      const error = event.error instanceof Error ? event.error : new Error(event.message || "prediction worker failed");
      queueMicrotask(() => { throw error; });
    };
    this.#worker.postMessage({ kind: "initialize", policy, snapshot } satisfies WorkerCommand);
  }
  authoritative(snapshot: AuthoritativeSnapshot): void { if (!this.#disposed) this.#worker.postMessage({ kind: "authoritative_snapshot", snapshot } satisfies WorkerCommand); }
  dispose(): void { if (this.#disposed) return; this.#disposed = true; this.#worker.postMessage({ kind: "dispose" } satisfies WorkerCommand); this.#worker.terminate(); }
}
