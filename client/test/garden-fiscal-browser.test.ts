import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import type { GameUIFiscalArm } from "../src/api/generated/types";
import { t } from "../src/copy";
import FiscalSurface from "../src/game-ui/FiscalSurface.svelte";

const browser = typeof document !== "undefined";
const NOW = 1_800_000_000_000;
function fiscal(owned = false, credit = 4, offered = true): GameUIFiscalArm {
  return {
    credit, credit_cap: { amount: 1000, reason_key: "cap.fiscal_credit" }, credit_per_period: 3,
    generator_levels: [], hoard: { cap_credits: 100, preview_ppm: 0, reason_note: "next_run" },
    period: { auto_ms: 300_000, early_ms: 100_000, early_success_ppm: 500_000, guaranteed_ms: 200_000, opened_wall_ms: NOW, seq: 0 },
    sweep_preview: { credit_after: credit, credited: 0, periods: 0, saturated: false },
    unlocks: offered ? [{ cost: 3, owned, unlock_id: "minigame.server_garden" }] : [],
  };
}
async function mounted(arm: GameUIFiscalArm) {
  const target = document.createElement("div"); document.body.append(target);
  const purchases: string[] = [];
  const app = mount(FiscalSurface, { target, props: {
    arm, era: "era_1995", serverNowMs: NOW, pending: false, controlsEnabled: true,
    onHarvest() { throw new Error("unexpected Fiscal harvest"); },
    onSpendLevel() { throw new Error("unexpected level purchase"); },
    onSpendUnlock(id: string) { purchases.push(id); },
  } });
  await tick(); flushSync();
  return { target, purchases, dispose: async () => { await unmount(app); target.remove(); } };
}

it.skipIf(!browser)("Garden Fiscal purchase consumes the offered ID and existing copy (SG7/RP-232)", async () => {
  const host = await mounted(fiscal());
  try {
    const rows = host.target.querySelectorAll(".fiscal li"); expect(rows).toHaveLength(1);
    const row = rows[0]!;
    expect(row.querySelector("h3")?.textContent).toBe(t("garden.title", {}, "era_1995"));
    expect(row.querySelector("p")?.textContent).toBe(t("garden.why", {}, "era_1995"));
    const button = row.querySelector("button")!;
    expect(button.textContent).toBe(t("fiscal.unlock_buy", { cost: 3 }, "era_1995"));
    expect(button.disabled).toBe(false); button.click(); await tick(); flushSync();
    expect(host.purchases).toEqual(["minigame.server_garden"]);
    // The callback is not a receipt: no optimistic owned state appears.
    expect(row.querySelector("strong")).toBeNull(); expect(row.querySelector("button")).not.toBeNull();
  } finally { await host.dispose(); }
});

it.skipIf(!browser)("Garden Fiscal purchase preserves absent, owned and unaffordable states (SG7/RP-232)", async () => {
  for (const state of ["absent", "owned", "unaffordable"] as const) {
    const host = await mounted(fiscal(state === "owned", state === "unaffordable" ? 0 : 4, state !== "absent"));
    try {
      const rows = host.target.querySelectorAll(".fiscal li"); expect(rows).toHaveLength(state === "absent" ? 0 : 1);
      if (state === "owned") {
        expect(rows[0]!.querySelector("button")).toBeNull();
        expect(rows[0]!.querySelector("strong")?.textContent).toBe(t("fiscal.unlock_owned", {}, "era_1995"));
      } else if (state === "unaffordable") {
        const button = rows[0]!.querySelector("button")!; expect(button.disabled).toBe(true); button.click();
      }
      expect(host.purchases).toEqual([]);
    } finally { await host.dispose(); }
  }
});
