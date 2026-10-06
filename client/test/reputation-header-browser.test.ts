import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import economyJSON from "../../balance/catalogs/phase0.json";
import curriculumJSON from "../../balance/curriculum/t0-t1.json";
import declaration from "../../balance/testdata/reputation-tree/fixture-v1.json";
import vectors from "../../testdata/reputation/bonus-vectors-v1.json";
import type { GameUIReputationArm } from "../src/api/generated/types";
import { t, type CopyEra } from "../src/copy";
import { COPY_KEYS } from "../src/copy/generated/types";
import { parseCurriculumCatalog } from "../src/curriculum";
import { parseCatalog } from "../src/economy-kernel";
import { loadReputationTree, reputationBonusFactor } from "../src/reputation";
import { installTheme, UI_THEMES } from "../src/ui/themes";
import ReputationFocusHarness from "./fixtures/ReputationFocusHarness.svelte";

// Actual loader + controlled public component props, not live SQL/Exit/mint.
// Formula text below characterizes raw ppm pending copy; RP-293 remains RED.
const browser = typeof document !== "undefined";
const economySource = structuredClone(economyJSON) as { multiplier_sources: unknown[] };
economySource.multiplier_sources.push({ id: "reputation.founder_bonus", slot: "prestige", target: "all", provider: "reputation_tree" });
const economy = parseCatalog(economySource);
const copyKeys = new Set<string>(COPY_KEYS);
const declarations = { economy, copyKeys, curriculum: parseCurriculumCatalog(curriculumJSON, economy, copyKeys, ["gate.t0_to_t1"]) };
const ids = declaration.nodes.map((node) => node.node_id);
function vector(index: number): string {
  expect(vectors.schema_version).toBe(1); expect(vectors.vectors).toHaveLength(10);
  return vectors.vectors[index]!.factor;
}
const profiles = [
  { name: "no frozen row", level: 4, owned: [], perLevel: 10_000, firstUnlock: 50_000, frozen: null, next: "1e0" },
  { name: "unit frozen row", level: 4, owned: [], perLevel: 10_000, firstUnlock: 50_000, frozen: "1e0", next: "1e0" },
  { name: "first unlock", level: 4, owned: [ids[0]!], perLevel: 10_000, firstUnlock: 50_000, frozen: "1e0", next: vector(4) },
  { name: "spending keeps earned bonus", level: 4, owned: [ids[0]!, ids[1]!], perLevel: 10_000, firstUnlock: 50_000, frozen: "1e0", next: vector(4) },
  { name: "full tree", level: 552, owned: ids, perLevel: 10_000, firstUnlock: 50_000, frozen: "1.002e0", next: vector(5) },
  { name: "maximum earned level", level: Number.MAX_SAFE_INTEGER, owned: ids, perLevel: 1_000_000, firstUnlock: 50_000, frozen: "1.002e0", next: vector(8) },
  { name: "fractional-percent ppm", level: 3, owned: [ids[0]!], perLevel: 1, firstUnlock: 1, frozen: null, next: vector(7) },
] as const;
type Profile = typeof profiles[number];
function armFor(profile: Profile): GameUIReputationArm {
  const source = structuredClone(declaration);
  source.bonus.per_level_ppm = profile.perLevel;
  source.nodes[0]!.unlock_ppm = profile.firstUnlock;
  const tree = loadReputationTree(source, declarations);
  expect(tree.nodes).toHaveLength(9);
  const owned = new Set<string>(profile.owned);
  const spent = tree.nodes.filter((node) => owned.has(node.node_id)).reduce((sum, node) => sum + node.cost, 0);
  const unlock = tree.nodes.reduce((max, node) => node.kind === "bonus_unlock" && owned.has(node.node_id) ? Math.max(max, node.unlock_ppm) : max, 0);
  const available = profile.level - spent;
  expect(reputationBonusFactor(profile.level, spent, profile.perLevel, unlock)).toBe(profile.next);
  return {
    available, level: profile.level, spent, unlock_ppm: unlock, per_level_ppm: profile.perLevel,
    bonus_factor_this_run: profile.frozen, bonus_factor_next_run: profile.next,
    nodes: tree.nodes.map((node) => ({ node_id: node.node_id, kind: node.kind, cost: node.cost,
      requires: [...node.requires], title_key: node.title_key, body_key: node.body_key,
      state: owned.has(node.node_id) ? "owned" : node.requires.some((id) => !owned.has(id)) ? "locked" : node.cost > available ? "unaffordable" : "available" })),
  };
}
async function settle(): Promise<void> { await tick(); flushSync(); }
function header(target: HTMLElement, era: CopyEra, arm: GameUIReputationArm): void {
  const card = target.querySelector<HTMLElement>(".reputation > section.card")!;
  expect(card).toBeTruthy();
  expect([...card.querySelectorAll("p")].map((paragraph) => paragraph.textContent)).toEqual([
    t("reputation_tree.balance.available", { amount: arm.available }, era),
    t("reputation_tree.balance.level", { amount: arm.level }, era),
    t("reputation_tree.balance.spent", { amount: arm.spent }, era),
    arm.bonus_factor_this_run === null ? t("reputation_tree.bonus.this_run_none", {}, era) : t("reputation_tree.bonus.this_run", { factor: arm.bonus_factor_this_run }, era),
    t("reputation_tree.bonus.next_run", { factor: arm.bonus_factor_next_run }, era),
    // Existing raw ppm binding only; never proof of R9 percentage/formula copy.
    t("reputation_tree.bonus.formula", { perlevel: arm.per_level_ppm, unlock: arm.unlock_ppm }, era),
    t("reputation_tree.applies_next_run", {}, era),
  ]);
  expect(card.querySelector("button,input")).toBeNull();
  for (const id of ids) expect(target.textContent).not.toContain(id);
}

for (const era of ["era_1995", "era_2000"] as const) {
  for (const profile of profiles) {
    it.skipIf(!browser)(`R9 header accounting/frozen-next bindings (formula-unit hold retained): ${era}, ${profile.name}`, async () => {
      const arm = armFor(profile);
      const target = document.createElement("main"); document.body.append(target); installTheme(target, UI_THEMES[era], false);
      const purchases: string[] = [];
      const app = mount(ReputationFocusHarness, { target, props: { initialArm: arm, era, onPurchase: (id: string) => { purchases.push(id); } } });
      try {
        await settle(); header(target, era, arm);
        expect(target.querySelectorAll(".reputation li")).toHaveLength(9);
        expect(purchases).toEqual([]);
        if (profile.name === "fractional-percent ppm") {
          expect(arm.per_level_ppm).toBe(1); expect(arm.unlock_ppm).toBe(1);
          // Actual admitted ppm cannot be represented as integer percentages.
          expect(Number.isSafeInteger(arm.per_level_ppm / 10_000)).toBe(false);
          expect(Number.isSafeInteger(arm.unlock_ppm / 10_000)).toBe(false);
          expect(() => t("reputation_tree.bonus.formula", { perlevel: 0.0001, unlock: 0.0001 }, era)).toThrow(/safe integer/);
        }
      } finally { await unmount(app); target.remove(); }
    });
  }

  it.skipIf(!browser)(`R9 header retains frozen factor through pending/spend and consumes a new frozen row only from props: ${era}`, async () => {
    const initial = armFor(profiles[2]); const afterSpend = armFor(profiles[3]);
    expect([initial.available, initial.level, initial.spent]).toEqual([3, 4, 1]);
    expect([afterSpend.available, afterSpend.level, afterSpend.spent]).toEqual([1, 4, 3]);
    expect(initial.bonus_factor_next_run).toBe(afterSpend.bonus_factor_next_run);
    const target = document.createElement("main"); document.body.append(target); installTheme(target, UI_THEMES[era], false);
    const purchases: string[] = [];
    const app = mount(ReputationFocusHarness, { target, props: { initialArm: initial, era, onPurchase: (id: string) => { purchases.push(id); } } });
    try {
      await settle(); header(target, era, initial);
      app.setPending(true); await settle(); header(target, era, initial);
      app.deliverArm(afterSpend); await settle(); header(target, era, afterSpend);
      const newRun = { ...afterSpend, bonus_factor_this_run: afterSpend.bonus_factor_next_run };
      app.deliverArm(newRun); await settle(); header(target, era, newRun);
      expect(purchases).toEqual([]); // Rendering props, not proof of a purchase/Exit.
    } finally { await unmount(app); target.remove(); }
  });
}
