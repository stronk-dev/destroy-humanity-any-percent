export type RecoveryPauseReason = "hidden" | "network";
export type RecoveryTerminalKind = "cancelled" | "resolved" | "watchdog";

export interface RecoveryProgressResponse {
  readonly session_id: string;
  readonly attended_progress_ms: number;
  readonly required_duration_attended_ms: number;
  readonly last_progress_server_ms: number;
  readonly eligible: boolean;
}

export interface RecoveryProgressTransport {
  progress(sessionId: string, progressToken: string): Promise<RecoveryProgressResponse>;
}

export interface RecoveryVisibility {
  subscribe(callback: (visible: boolean) => void): () => void;
}

export interface RecoverySchedulerInput {
  readonly session_id: string;
  readonly progress_token: string;
  readonly beat_interval_ms: number;
  readonly transport: RecoveryProgressTransport;
  readonly now: () => number;
  readonly visibility: RecoveryVisibility;
}

export interface RecoverySchedulerCallbacks {
  readonly on_progress: (response: RecoveryProgressResponse) => void;
  readonly on_pause: (reason: RecoveryPauseReason) => void;
  readonly on_resume: () => void;
  readonly on_token_rotated: (token: string) => void;
  readonly on_terminal: (kind: RecoveryTerminalKind) => void;
}

export class RecoveryScheduler {
  readonly #input: RecoverySchedulerInput;
  readonly #callbacks: RecoverySchedulerCallbacks;
  #progressToken: string;
  #timer: ReturnType<typeof setTimeout> | undefined;
  #unsubscribe: (() => void) | undefined;
  #visible = true;
  #running = false;
  #requestInFlight = false;
  #lastDispatchMS: number;

  constructor(input: RecoverySchedulerInput, callbacks: RecoverySchedulerCallbacks) {
    if (!mechanicalSession(input.session_id) || !opaqueToken(input.progress_token) ||
        !Number.isSafeInteger(input.beat_interval_ms) || input.beat_interval_ms < 1 ||
        typeof input.now !== "function" || typeof input.transport?.progress !== "function" ||
        typeof input.visibility?.subscribe !== "function") {
      throw new TypeError("invalid Soul recovery scheduler input");
    }
    this.#input = input;
    this.#callbacks = callbacks;
    this.#progressToken = input.progress_token;
    this.#lastDispatchMS = input.now();
  }

  start(): void {
    if (this.#running) return;
    this.#running = true;
    this.#lastDispatchMS = this.#input.now();
    this.#unsubscribe = this.#input.visibility.subscribe((visible) => this.#setVisible(visible));
    if (this.#visible) this.#schedule();
  }

  stop(kind?: RecoveryTerminalKind): void {
    if (!this.#running) return;
    this.#running = false;
    this.#clearTimer();
    this.#unsubscribe?.();
    this.#unsubscribe = undefined;
    if (kind !== undefined) this.#callbacks.on_terminal(kind);
  }

  reconnect(progressToken: string): void {
    if (!opaqueToken(progressToken)) throw new TypeError("invalid Soul recovery progress token");
    this.#progressToken = progressToken;
    this.#lastDispatchMS = this.#input.now();
    this.#callbacks.on_token_rotated(progressToken);
    if (this.#running && this.#visible && !this.#requestInFlight) {
      this.#clearTimer();
      this.#schedule();
    }
  }

  #setVisible(visible: boolean): void {
    if (visible === this.#visible) return;
    this.#visible = visible;
    if (!visible) {
      this.#clearTimer();
      this.#callbacks.on_pause("hidden");
      return;
    }
    this.#lastDispatchMS = this.#input.now();
    this.#callbacks.on_resume();
    if (this.#running && !this.#requestInFlight) this.#schedule();
  }

  #schedule(): void {
    if (!this.#running || !this.#visible || this.#timer !== undefined) return;
    this.#timer = setTimeout(() => {
      this.#timer = undefined;
      void this.#beat();
    }, this.#input.beat_interval_ms);
  }

  async #beat(): Promise<void> {
    if (!this.#running || !this.#visible || this.#requestInFlight) return;
    const now = this.#input.now();
    if (!Number.isSafeInteger(now) || now < this.#lastDispatchMS) {
      this.#callbacks.on_pause("network");
      return;
    }
    this.#lastDispatchMS = now;
    this.#requestInFlight = true;
    try {
      const response = await this.#input.transport.progress(this.#input.session_id, this.#progressToken);
      if (!this.#running) return;
      this.#callbacks.on_progress(response);
    } catch {
      if (this.#running) this.#callbacks.on_pause("network");
    } finally {
      this.#requestInFlight = false;
      if (this.#running && this.#visible) this.#schedule();
    }
  }

  #clearTimer(): void {
    if (this.#timer !== undefined) clearTimeout(this.#timer);
    this.#timer = undefined;
  }
}

function mechanicalSession(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u.test(value);
}

function opaqueToken(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u.test(value);
}
