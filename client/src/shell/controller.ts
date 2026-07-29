import type { AuthoritativeSnapshot, IntentReceipt } from "./contracts";
import { DisplayCounter, type CounterView } from "./display";
import type { ClientShellPolicy } from "./policy";
import type { PredictedSnapshot } from "./prediction";

export type ShellTab = "company" | "world" | "pet" | "minigame";
export interface ShellView { readonly screen: "contract" | "main"; readonly activeTab: ShellTab; readonly revision: number; readonly resources: Readonly<Record<string, CounterView>>; readonly discrete: Readonly<Record<string, boolean | number | string>> }

export class ShellController {
  readonly #policy: ClientShellPolicy; readonly #reduced: boolean; readonly #begin: () => void | Promise<void>; readonly #listeners = new Set<(view: ShellView) => void>();
  #screen: "contract" | "main" = "contract"; #tab: ShellTab = "company"; #revision = 0; #discrete: Readonly<Record<string, boolean | number | string>> = {}; #counters = new Map<string, DisplayCounter>();
  constructor(policy: ClientShellPolicy, begin: () => void | Promise<void> = () => {}, reducedMotion = false) { this.#policy = policy; this.#begin = begin; this.#reduced = reducedMotion; }
  subscribe(listener: (view: ShellView) => void): () => void { this.#listeners.add(listener); listener(this.view(performance.now())); return () => this.#listeners.delete(listener); }
  async beginAttempt(): Promise<void> { await this.#begin(); this.#screen = "main"; this.#emit(); }
  selectTab(tab: ShellTab): void { this.#tab = tab; this.#emit(); }
  applyAuthoritative(snapshot: AuthoritativeSnapshot, receipt?: IntentReceipt, nowMs = performance.now()): void {
    if (snapshot.revision < this.#revision) return; this.#revision = snapshot.revision; this.#discrete = snapshot.discrete;
    for (const [id, resource] of Object.entries(snapshot.resources)) { let counter = this.#counters.get(id); if (!counter) { counter = new DisplayCounter(resource, this.#policy, this.#reduced); this.#counters.set(id, counter); } else counter.applyAuthoritative(resource, nowMs, receipt); }
    this.#emit(nowMs);
  }
  applyPrediction(snapshot: PredictedSnapshot, nowMs = performance.now()): void { if (snapshot.revision < this.#revision) return; for (const [id, value] of Object.entries(snapshot.resources)) this.#counters.get(id)?.applyPrediction(value); this.#emit(nowMs); }
  view(nowMs = performance.now()): ShellView { return { screen: this.#screen, activeTab: this.#tab, revision: this.#revision, resources: Object.fromEntries([...this.#counters].map(([id, counter]) => [id, counter.view(nowMs)])), discrete: this.#discrete }; }
  #emit(nowMs = performance.now()): void { const view = this.view(nowMs); for (const listener of this.#listeners) listener(view); }
}
