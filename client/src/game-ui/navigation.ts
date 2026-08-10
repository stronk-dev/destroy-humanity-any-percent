import type { GameUISurfaceID } from "./surface-catalog";

export type LifecycleNavigation = Readonly<{ cursor: number; surface: "offer_sheet" | "run_end" }>;

export class GameUINavigation {
  #active: GameUISurfaceID;
  #destructiveConfirmation = false;
  #lastLifecycleCursor = -1;
  #deferred: LifecycleNavigation | undefined;

  constructor(initial: GameUISurfaceID) { this.#active = initial; }
  get active(): GameUISurfaceID { return this.#active; }

  select(surface: GameUISurfaceID): void { this.#active = surface; }

  settingsConfirmation(open: boolean): void {
    this.#destructiveConfirmation = open;
    if (!open && this.#deferred) {
      this.#active = this.#deferred.surface;
      this.#deferred = undefined;
    }
  }

  lifecycle(event: LifecycleNavigation): void {
    if (!Number.isSafeInteger(event.cursor) || event.cursor < 0) throw new RangeError("invalid lifecycle cursor");
    if (event.cursor <= this.#lastLifecycleCursor) return;
    this.#lastLifecycleCursor = event.cursor;
    if (this.#active === "settings" && this.#destructiveConfirmation) { this.#deferred = event; return; }
    this.#active = event.surface;
  }
}
