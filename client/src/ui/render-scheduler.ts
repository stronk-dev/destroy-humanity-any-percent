export type RenderCallback = () => void;

export class SharedRenderScheduler {
  readonly intervalMs: number;
  #callbacks = new Map<object, RenderCallback>();
  #timer: ReturnType<typeof setTimeout> | undefined;

  constructor(intervalMs = 100) {
    if (!Number.isSafeInteger(intervalMs) || intervalMs <= 0) throw new RangeError("render interval must be a positive integer");
    this.intervalMs = intervalMs;
  }

  schedule(owner: object, callback: RenderCallback): void {
    this.#callbacks.set(owner, callback);
    if (this.#timer !== undefined) return;
    this.#timer = setTimeout(() => this.flush(), this.intervalMs);
  }

  cancel(owner: object): void {
    this.#callbacks.delete(owner);
    if (this.#callbacks.size === 0 && this.#timer !== undefined) {
      clearTimeout(this.#timer);
      this.#timer = undefined;
    }
  }

  flush(): void {
    if (this.#timer !== undefined) clearTimeout(this.#timer);
    this.#timer = undefined;
    const callbacks = [...this.#callbacks.values()];
    this.#callbacks.clear();
    for (const callback of callbacks) callback();
  }

  get pendingCount(): number { return this.#callbacks.size; }
}

export const amountRenderScheduler = new SharedRenderScheduler(100);
