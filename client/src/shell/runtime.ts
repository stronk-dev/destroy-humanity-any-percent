import { validateReceipt, validateSnapshot, type AuthoritativeSnapshot, type SnapshotStream } from "./contracts";
import type { ShellController } from "./controller";
import { bindLifecycle } from "./lifecycle";
import type { ClientShellPolicy } from "./policy";
import type { WorkerOutput } from "./worker-protocol";
import { PredictionWorkerClient } from "./worker-client";

export interface PredictionPort { authoritative(snapshot: AuthoritativeSnapshot): void; dispose(): void }
export type PredictionPortFactory = (policy: ClientShellPolicy, snapshot: AuthoritativeSnapshot, consume: (output: WorkerOutput) => void) => PredictionPort;

const browserWorkerFactory: PredictionPortFactory = (policy, snapshot, consume) => new PredictionWorkerClient(policy, snapshot, consume);

export class ShellRuntime {
  readonly #controller: ShellController;
  readonly #policy: ClientShellPolicy;
  readonly #stream: SnapshotStream;
  readonly #workerFactory: PredictionPortFactory;
  #worker: PredictionPort | undefined;
  #unsubscribe: (() => void) | undefined;
  #unbindLifecycle: (() => void) | undefined;

  constructor(controller: ShellController, policy: ClientShellPolicy, stream: SnapshotStream, workerFactory: PredictionPortFactory = browserWorkerFactory) {
    this.#controller = controller; this.#policy = policy; this.#stream = stream; this.#workerFactory = workerFactory;
  }

  start(documentTarget: Document | undefined = globalThis.document, windowTarget: Window | undefined = globalThis.window): void {
    if (this.#unsubscribe) return;
    this.#unsubscribe = this.#stream.subscribe((source, receiptSource) => {
      const snapshot = validateSnapshot(source);
      const receipt = receiptSource ? validateReceipt(receiptSource) : undefined;
      this.#controller.applyAuthoritative(snapshot, receipt);
      if (!this.#worker) this.#worker = this.#workerFactory(this.#policy, snapshot, (output) => this.#consumeWorker(output));
      else this.#worker.authoritative(snapshot);
    });
    if (documentTarget && windowTarget) this.#unbindLifecycle = bindLifecycle(documentTarget, windowTarget, this.#stream, () => this.#worker?.dispose(), (gapMs) => this.#controller.beginReturnStory(gapMs));
    this.#stream.requestSnapshot();
  }

  dispose(): void {
    this.#unbindLifecycle?.(); this.#unsubscribe?.(); this.#worker?.dispose(); this.#stream.dispose(); this.#controller.dispose();
    this.#unbindLifecycle = undefined; this.#unsubscribe = undefined; this.#worker = undefined;
  }

  #consumeWorker(output: WorkerOutput): void {
    if (output.kind === "predicted_snapshot") this.#controller.applyPrediction(output.snapshot);
    else if (output.kind === "offline_required") { this.#controller.beginReturnStory(output.gapMs); this.#stream.requestSnapshot(); }
    else this.#controller.recordWorker(output);
  }
}
