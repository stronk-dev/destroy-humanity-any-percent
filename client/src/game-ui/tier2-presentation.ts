import candidate from "../../../balance/testdata/t2/presentation-candidate-v3.json";

import { GAME_UI_PRESENTATION, parseGameUIPresentation, requirePresentation } from "./presentation";

// E4's existing candidate bindings, not catalog activation or owner copy adoption.
// Pinned bindings take precedence; unfamiliar IDs still fail loud. The actual
// snapshot, never this presentation map, determines which content is rendered.
const TIER2_PRESENTATION = parseGameUIPresentation(candidate);

export function generatorPresentation(id: string) {
  return GAME_UI_PRESENTATION.generators.get(id) ?? requirePresentation(TIER2_PRESENTATION.generators, id);
}

export function gatePresentation(id: string) {
  return GAME_UI_PRESENTATION.gates.get(id) ?? requirePresentation(TIER2_PRESENTATION.gates, id);
}

export function tier2UpgradePresentation(id: string) {
  return GAME_UI_PRESENTATION.upgrades.get(id) ?? requirePresentation(TIER2_PRESENTATION.upgrades, id);
}
