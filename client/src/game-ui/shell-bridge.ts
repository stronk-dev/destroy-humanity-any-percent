import policySource from "../../../balance/client-shell/phase0.json";

import type { IntentReceipt, SnapshotStream } from "../shell/contracts";
import { ShellController, type ShellView } from "../shell/controller";
import { parseClientShellPolicy } from "../shell/policy";
import { ShellRuntime } from "../shell/runtime";
import { toShellSnapshot, type ParsedGameUISnapshot } from "./contracts";

export class GameUISnapshotStream implements SnapshotStream {
  readonly #request: () => void;
  readonly #consumers = new Set<(snapshot: ReturnType<typeof toShellSnapshot>, receipt?: IntentReceipt) => void>();
  #disposed = false;

  constructor(request: () => void) { this.#request = request; }

  subscribe(consumer: (snapshot: ReturnType<typeof toShellSnapshot>, receipt?: IntentReceipt) => void): () => void {
    if (this.#disposed) throw new Error("Game UI snapshot stream is disposed");
    this.#consumers.add(consumer);
    return () => this.#consumers.delete(consumer);
  }

  publish(snapshot: ParsedGameUISnapshot, receipt?: IntentReceipt): void {
    if (this.#disposed) return;
    const shell = toShellSnapshot(snapshot);
    for (const consumer of this.#consumers) consumer(shell, receipt);
  }

  requestSnapshot(): void { if (!this.#disposed) this.#request(); }
  flush(): void { /* Intents are sent eagerly; there is no client-side authoritative queue. */ }
  dispose(): void { this.#disposed = true; this.#consumers.clear(); }
}

export class GameUIShell {
  readonly #controller;
  readonly #stream;
  readonly #runtime;
  #started = false;

  constructor(requestSnapshot: () => void) {
    const policy = parseClientShellPolicy(policySource);
    this.#controller = new ShellController(policy);
    this.#stream = new GameUISnapshotStream(requestSnapshot);
    this.#runtime = new ShellRuntime(this.#controller, policy, this.#stream);
  }

  start(): void {
    if (this.#started) return;
    this.#started = true;
    this.#runtime.start();
  }
  publish(snapshot: ParsedGameUISnapshot, receipt?: IntentReceipt): void { this.#stream.publish(snapshot, receipt); }
  subscribe(consumer: (view: ShellView) => void): () => void { return this.#controller.subscribe(consumer); }
  view(): ShellView { return this.#controller.view(); }
  dispose(): void { this.#runtime.dispose(); }
}
