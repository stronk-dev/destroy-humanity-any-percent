import source from "./surfaces.json";

import type { DiscreteFact } from "../shell/contracts";
import { parseSurfaceRegistry, type SurfaceRow } from "../ui/surfaces";

export const GAME_UI_FACT_IDS = new Set(["bootstrap.needed"]);
export const GAME_UI_SURFACES: readonly SurfaceRow[] = parseSurfaceRegistry(source, GAME_UI_FACT_IDS);
export type GameUISurfaceID = "desk" | "offer_sheet" | "run_end" | "settings" | "vision_slide";

export function defaultSurface(facts: Readonly<Record<string, DiscreteFact>>): GameUISurfaceID {
  return facts["bootstrap.needed"] === true ? "vision_slide" : "desk";
}
