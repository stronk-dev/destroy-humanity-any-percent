import source from "./surfaces.json";

import type { DiscreteFact } from "../shell/contracts";
import { parseSurfaceRegistry, type SurfaceRow } from "../ui/surfaces";

export const GAME_UI_FACT_IDS = new Set(["bootstrap.needed", "feature.achievements", "feature.active_play", "feature.fiscal", "feature.meters", "feature.minigame.pitch", "feature.pets", "feature.reputation_tree"]);
export const GAME_UI_SURFACES: readonly SurfaceRow[] = parseSurfaceRegistry(source, GAME_UI_FACT_IDS);
export type GameUISurfaceID = "achievements" | "desk" | "fiscal" | "meters" | "minigame_session" | "offer_sheet" | "reputation_tree" | "run_end" | "settings" | "soul_recovery" | "vision_slide";

export function defaultSurface(facts: Readonly<Record<string, DiscreteFact>>): GameUISurfaceID {
  return facts["bootstrap.needed"] === true ? "vision_slide" : "desk";
}
