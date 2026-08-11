import { t, type CopyEra } from "../copy";
import { canonicalString } from "../numeric";
import { formatAmount } from "../ui/amount-format";
import type { PrestigeTerms } from "./events";
import { GAME_UI_PRESENTATION } from "./presentation";

// Copy-key order is the one stable presentation order. Unknown future slot IDs
// are withheld until their owner-authored presentation rows ship; raw IDs are
// never rendered as titles.
export function renderPrestigeTermRows(terms: PrestigeTerms, era: CopyEra): readonly string[] {
  const rows: string[] = [];
  const note = GAME_UI_PRESENTATION.cloutReachNotes.get(terms.clout_reach_note);
  if (note) rows.push(t(note.text_key, {}, era));
  for (const slot of terms.network_slot_unlocks) {
    const binding = GAME_UI_PRESENTATION.networkSlots.get(slot.slot);
    if (binding) rows.push(t("terms.network_slot_unlock.frame", { title: t(binding.title_key, {}, era) }, era));
  }
  rows.push(t("terms.reputation_delta.frame", { delta: formatAmount(canonicalString(terms.reputation_delta)) }, era));
  rows.push(t("terms.route_knowledge.frame", { delta: terms.route_knowledge }, era));
  return rows;
}
