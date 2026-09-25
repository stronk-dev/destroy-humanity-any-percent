import { operations, type GardenCurrentResponse } from "../../api/generated/types";
import { createOperationCall, type Fetcher } from "../minigame/session-port";

// Server Garden SG9/SG10: the garden surface's only read seam, typed by the
// generated DTO. Commands go through the Founder intent route (runtime.intent).
export interface GardenPort {
  current(): Promise<GardenCurrentResponse>;
}

export function createBrowserGardenPort(accessToken: () => string, fetcher: Fetcher = fetch): GardenPort {
  const call = createOperationCall(accessToken, fetcher);
  return { current: () => call(operations.get_current_garden.method, operations.get_current_garden.path, undefined) };
}

export type GardenViewState =
  | Readonly<{ kind: "loading" }>
  | Readonly<{ kind: "locked" }>
  | Readonly<{ kind: "active"; view: Extract<GardenCurrentResponse, { kind: "active" }> }>
  | Readonly<{ kind: "stale"; view: Extract<GardenCurrentResponse, { kind: "active" }> }>
  | Readonly<{ kind: "error" }>;

/** Tick announcement counts between two consecutive active views (SG10 a11y 3). */
export function tickDelta(previous: GardenViewState, next: Extract<GardenCurrentResponse, { kind: "active" }>): { matured: number; spawned: number } {
  if (previous.kind !== "active" && previous.kind !== "stale") return { matured: 0, spawned: 0 };
  const before = new Map(previous.view.garden.plots.map((plot) => [`${plot.row},${plot.col}`, plot]));
  let matured = 0, spawned = 0;
  for (const plot of next.garden.plots) {
    const prior = before.get(`${plot.row},${plot.col}`);
    if (!prior) spawned += 1;
    else if (prior.stage === "growing" && plot.stage === "mature" && prior.species_id === plot.species_id) matured += 1;
  }
  return { matured, spawned };
}
