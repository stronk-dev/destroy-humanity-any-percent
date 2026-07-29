import Decimal from "break_infinity.js";
import { parseCanonical } from "../numeric";
import type { AuthoritativeSnapshot, IntentReceipt, ProgressCoordinate } from "./contracts";
import { DisplayCounter, type CounterView } from "./display";
import { selectReturnStory, type ReturnStory } from "./lifecycle";
import type { ClientShellPolicy } from "./policy";
import type { PredictedSnapshot } from "./prediction";
import { reconcileDiscrete } from "./reconciliation";
import { ClientTelemetry, type ClientTelemetrySnapshot } from "./telemetry";
import type { WorkerOutput } from "./worker-protocol";

export type ShellTab = "company" | "world" | "pet" | "minigame";
export interface ProgressView extends ProgressCoordinate { readonly fillPpm: number }
export interface ShellView {
  readonly screen: "contract" | "main" | "run_end";
  readonly activeTab: ShellTab;
  readonly revision: number;
  readonly resources: Readonly<Record<string, CounterView>>;
  readonly discrete: Readonly<Record<string, boolean | number | string>>;
  readonly progress: readonly ProgressView[];
  readonly returnStory?: ReturnStory;
  readonly returnFastForwardComplete: boolean;
  readonly offlineGains: Readonly<Record<string, string>>;
  readonly attemptElapsedMs: number;
}

export class ShellController {
  readonly #policy: ClientShellPolicy;
  readonly #reduced: boolean;
  readonly #begin: () => void | Promise<void>;
  readonly #telemetry: ClientTelemetry;
  readonly #clock: () => number;
  readonly #listeners = new Set<(view: ShellView) => void>();
  #screen: "contract" | "main" | "run_end" = "contract";
  #tab: ShellTab = "company";
  #revision = 0;
  #hasSnapshot = false;
  #discrete: Readonly<Record<string, boolean | number | string>> = {};
  #progress: readonly ProgressCoordinate[] = [];
  #counters = new Map<string, DisplayCounter>();
  #lastAmounts = new Map<string, string>();
  #returnBase = new Map<string, string>();
  #returnStory: ReturnStory | undefined;
  #returnFastForwardComplete = false;
  #offlineGains: Readonly<Record<string, string>> = {};
  #returnTimer: ReturnType<typeof setTimeout> | undefined;
  #attemptStartedMs: number | undefined;
  #attemptEndedMs: number | undefined;

  constructor(policy: ClientShellPolicy, begin: () => void | Promise<void> = () => {}, reducedMotion = false, telemetry = new ClientTelemetry(), clock: () => number = () => performance.now()) {
    this.#policy = policy; this.#begin = begin; this.#reduced = reducedMotion; this.#telemetry = telemetry; this.#clock = clock;
  }

  subscribe(listener: (view: ShellView) => void): () => void {
    this.#listeners.add(listener); listener(this.view()); return () => this.#listeners.delete(listener);
  }

  async beginAttempt(): Promise<void> { await this.#begin(); this.#attemptStartedMs = this.#clock(); this.#attemptEndedMs = undefined; this.#screen = "main"; this.#emit(); }
  showRunEnd(): void { if (this.#attemptStartedMs === undefined) return; this.#attemptEndedMs = Math.max(0, this.#clock() - this.#attemptStartedMs); this.#screen = "run_end"; this.#emit(); }
  returnToContract(): void { this.#screen = "contract"; this.#attemptStartedMs = undefined; this.#attemptEndedMs = undefined; this.#emit(); }
  selectTab(tab: ShellTab): void { this.#tab = tab; this.#emit(); }

  beginReturnStory(gapMs: number, optionalModalReady = false): void {
    if (this.#returnStory) return;
    const story = selectReturnStory(gapMs, optionalModalReady, this.#policy);
    if (!story) return;
    this.#returnStory = story; this.#returnFastForwardComplete = false; this.#returnBase = new Map(this.#lastAmounts); this.#offlineGains = {};
    if (this.#returnTimer) clearTimeout(this.#returnTimer);
    this.#returnTimer = setTimeout(() => this.completeReturnFastForward(), story.fastForwardMs);
    this.#emit();
  }

  completeReturnFastForward(): void {
    if (!this.#returnStory) return;
    if (this.#returnTimer) clearTimeout(this.#returnTimer);
    this.#returnTimer = undefined; this.#returnFastForwardComplete = true; this.#emit();
  }

  dismissReturnStory(): void {
    if (this.#returnTimer) clearTimeout(this.#returnTimer);
    this.#returnTimer = undefined; this.#returnStory = undefined; this.#returnFastForwardComplete = false; this.#returnBase.clear(); this.#emit();
  }

  applyAuthoritative(snapshot: AuthoritativeSnapshot, receipt?: IntentReceipt, nowMs = this.#clock()): void {
    if (snapshot.revision < this.#revision) return;
    if (receipt && receipt.revision !== snapshot.revision) throw new RangeError("receipt revision does not match snapshot");
    const returning = this.#returnStory !== undefined;
    this.#telemetry.recordReceipt(receipt);
    for (const [id, resource] of Object.entries(snapshot.resources)) {
      let counter = this.#counters.get(id);
      if (!counter) {
        counter = new DisplayCounter(resource, this.#policy, this.#reduced); this.#counters.set(id, counter);
      } else if (returning) {
        counter.replaceAuthoritative(resource);
      } else {
        const decision = counter.applyAuthoritative(resource, nowMs, receipt);
        if (decision.mode === "rebase") this.#telemetry.recordContinuousRebase(receipt);
      }
    }
    for (const id of this.#counters.keys()) if (!(id in snapshot.resources)) this.#counters.delete(id);
    if (this.#hasSnapshot) {
      const discreteKeys = new Set([...Object.keys(this.#discrete), ...Object.keys(snapshot.discrete)]);
      for (const key of discreteKeys) if (reconcileDiscrete(this.#discrete[key], snapshot.discrete[key] as boolean | number | string).mode === "rebase") this.#telemetry.recordDiscreteRebase();
    }
    this.#revision = snapshot.revision; this.#discrete = snapshot.discrete; this.#progress = snapshot.progress; this.#hasSnapshot = true;
    if (returning) this.#offlineGains = calculateGains(this.#returnBase, snapshot);
    this.#lastAmounts = new Map(Object.entries(snapshot.resources).map(([id, resource]) => [id, resource.amount]));
    this.#emit(nowMs);
  }

  applyPrediction(snapshot: PredictedSnapshot, nowMs = this.#clock()): void {
    if (snapshot.revision < this.#revision || this.#returnStory) return;
    for (const [id, value] of Object.entries(snapshot.resources)) this.#counters.get(id)?.applyPrediction(value);
    this.#emit(nowMs);
  }

  recordWorker(output: WorkerOutput): void { this.#telemetry.recordWorker(output); }
  telemetry(): ClientTelemetrySnapshot { return this.#telemetry.snapshot(); }

  view(nowMs = this.#clock()): ShellView {
    return {
      screen: this.#screen,
      activeTab: this.#tab,
      revision: this.#revision,
      resources: Object.fromEntries([...this.#counters].map(([id, counter]) => [id, counter.view(nowMs)])),
      discrete: this.#discrete,
      progress: this.#progress.map(toProgressView),
      ...(this.#returnStory ? { returnStory: this.#returnStory } : {}),
      returnFastForwardComplete: this.#returnFastForwardComplete,
      offlineGains: this.#offlineGains,
      attemptElapsedMs: this.#attemptStartedMs === undefined ? 0 : this.#attemptEndedMs ?? Math.max(0, nowMs - this.#attemptStartedMs),
    };
  }

  dispose(): void { if (this.#returnTimer) clearTimeout(this.#returnTimer); this.#returnTimer = undefined; this.#listeners.clear(); }
  #emit(nowMs = this.#clock()): void { const view = this.view(nowMs); for (const listener of this.#listeners) listener(view); }
}

function toProgressView(progress: ProgressCoordinate): ProgressView {
  const fill = Decimal.min(parseCanonical(progress.current).div(parseCanonical(progress.target)), 1).mul(1_000_000).floor().toNumber();
  return { ...progress, fillPpm: Math.max(0, fill) };
}

function calculateGains(before: ReadonlyMap<string, string>, snapshot: AuthoritativeSnapshot): Readonly<Record<string, string>> {
  const gains: Record<string, string> = {};
  for (const [id, resource] of Object.entries(snapshot.resources)) {
    const prior = before.get(id);
    if (!prior) continue;
    const gain = parseCanonical(resource.amount).sub(parseCanonical(prior));
    if (gain.gt(0)) gains[id] = gain.toString();
  }
  return gains;
}
