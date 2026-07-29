import { validateSnapshot, type SnapshotStream } from "./contracts";
import type { ShellController } from "./controller";
import { bindLifecycle } from "./lifecycle";
import type { ClientShellPolicy } from "./policy";
import { PredictionWorkerClient } from "./worker-client";

export class ShellRuntime {
  readonly #controller: ShellController; readonly #policy: ClientShellPolicy; readonly #stream: SnapshotStream;
  #worker: PredictionWorkerClient | undefined; #unsubscribe: (() => void) | undefined; #unbindLifecycle: (() => void) | undefined;
  constructor(controller: ShellController, policy: ClientShellPolicy, stream: SnapshotStream) { this.#controller = controller; this.#policy = policy; this.#stream = stream; }
  start(documentTarget: Document = document, windowTarget: Window = window): void {
    if (this.#unsubscribe) return;
    this.#unsubscribe = this.#stream.subscribe((source, receipt) => {
      const snapshot = validateSnapshot(source); this.#controller.applyAuthoritative(snapshot, receipt);
      if (!this.#worker) this.#worker = new PredictionWorkerClient(this.#policy, snapshot, (output) => {
        if (output.kind === "predicted_snapshot") this.#controller.applyPrediction(output.snapshot);
        else if (output.kind === "offline_required") this.#stream.requestSnapshot();
      });
      else this.#worker.authoritative(snapshot);
    });
    this.#unbindLifecycle = bindLifecycle(documentTarget, windowTarget, this.#stream, () => this.#worker?.dispose());
    this.#stream.requestSnapshot();
  }
  dispose(): void { this.#unbindLifecycle?.(); this.#unsubscribe?.(); this.#worker?.dispose(); this.#stream.dispose(); this.#unbindLifecycle = undefined; this.#unsubscribe = undefined; this.#worker = undefined; }
}
