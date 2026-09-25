import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import armFixture from "../../testdata/gameui/cosmetics-arm-v1.json";
import type { GameUICosmeticsArm, GameUISnapshot } from "../src/api/generated/types";
import type { ParsedGameUISnapshot } from "../src/game-ui/contracts";
import GameUIApp from "../src/game-ui/GameUIApp.svelte";
import type { IntentOutcome } from "../src/game-ui/intent-outcome";
import type { GameUIRuntime, GameUIRuntimeMessage } from "../src/game-ui/runtime";
import type { GameUISurfaceID } from "../src/game-ui/surface-catalog";
import { installNetworkTrap, sameOriginAllowed } from "./network-trap";

const browser = typeof document !== "undefined";
const NOW = 1_800_000_000_000;

const meterRow = (meter_id: string, value: number) => ({ band_id: value >= 70 ? "high" : "low", bands: [{ band_id: "low", floor_value: 0 }, { band_id: "high", floor_value: 70 }], max: 100, meter_id, min: 0, value });
const trust = ["employees", "investors", "press", "regulators", "users"].flatMap((c) => [meterRow(`trust.${c}.grievance`, 0), meterRow(`trust.${c}.standing`, 50)]);

const v4: GameUISnapshot = {
  constants_hash: `sha256:${"a".repeat(64)}`, evaluated_through_ms: NOW,
  facts: [{ fact_id: "bootstrap.needed", value: false }, { fact_id: "feature.achievements", value: true }, { fact_id: "feature.active_play", value: false },
    { fact_id: "feature.fiscal", value: true }, { fact_id: "feature.meters", value: true }, { fact_id: "feature.minigame.pitch", value: true }, { fact_id: "feature.pets", value: false }],
  features: {
    achievements: { rows: [
      { achievement_id: "achievement.first_gate", condition_scope: "run", copy_key: "achievement.first_gate", earned: "run", proof_kind: "provenance", score_grant: 2 },
      { achievement_id: "achievement.generators_purchased_1", condition_scope: "run", copy_key: "achievement.generators_purchased_1", earned: null, proof_kind: "provenance", score_grant: 1 },
      { achievement_id: "achievement.old_hand", condition_scope: "career", copy_key: "achievement.old_hand", earned: "lifetime", proof_kind: "possession", score_grant: 5 },
    ], score: { lifetime: 5, run: 2 } },
    active_play: null,
    fiscal: { credit: 4, credit_cap: { amount: 1000, reason_key: "cap.fiscal_credit" }, credit_per_period: 3,
      generator_levels: [{ generator_id: "generator.beige_tower", level: 0, level_cap: { amount: 100, reason_key: "cap.fiscal_level.beige_tower" }, next_level_cost: 1, ppm_per_level: 10_000 }],
      hoard: { cap_credits: 100, preview_ppm: 40_000, reason_note: "next_run" },
      period: { auto_ms: 300_000, early_ms: 100_000, early_success_ppm: 500_000, guaranteed_ms: 200_000, opened_wall_ms: NOW, seq: 0 },
      sweep_preview: { credit_after: 4, credited: 0, periods: 0, saturated: false },
      unlocks: [{ cost: 3, owned: false, unlock_id: "minigame.pitch" }, { cost: 5, owned: false, unlock_id: "unlock.arcade" }] },
    meters: { meters: [meterRow("doom.probability", 50), ...trust].sort((a, b) => a.meter_id < b.meter_id ? -1 : 1) },
    minigames: { rows: [{ active_session: false, human_content_locked: false, minigame_id: "pitch", unlocked: false }] },
    pets: null,
  },
  founder_revision: 7,
  generators: [{ generator_id: "generator.beige_tower", max_affordable: 2, next_cost: "1e1", next_cost_resource_id: "company.cash", owned: 1, provision_cap: { amount: 3, reason_key: "generator.beige_tower.provisioned_cap" }, provisioned: 3, rate_contribution: "1e0" }],
  manual_action: { action_id: "manual.click", bucket_cap_milli: 50_000, refill_milli_per_ms: 25, refilled_at_ms: NOW, tokens_milli: 50_000 },
  progress: [], resources: [{ amount: "1e2", cap: { amount: "1e1000", reason_key: "resource.company_cash.cap.phase0" }, rate_per_second: "1e0", resource_id: "company.cash" }],
  revision: 1,
  run: { category: "any_percent", exit_count: 0, founder_id: "01985555-1111-7111-8111-111111111111", run_seq: 1, run_started_at_ms: NOW - 1_000, tier: 0 },
  schema_version: 4, server_now_ms: NOW,
  transitions: { cross_gate: null, wind_down: { eligible: false } },
  upgrades: [{ cost_amount: "2e1", cost_resource_id: "company.cash", eligible: false, owned: true, upgrade_id: "upgrade.beige_tower_cache" }],
};

class Runtime implements GameUIRuntime {
  readonly requests: Readonly<Record<string, unknown>>[] = [];
  outcome: IntentOutcome = { outcome: "applied", receipt: {} };
  current: ParsedGameUISnapshot = v4;
  hasCredentials(): boolean { return true; }
  async bootstrap(): Promise<ParsedGameUISnapshot> { return this.current; }
  async snapshot(): Promise<ParsedGameUISnapshot> { return this.current; }
  async intent(body: Readonly<Record<string, unknown>>): Promise<IntentOutcome> { this.requests.push(body); return this.outcome; }
  listener: ((message: GameUIRuntimeMessage) => void) | undefined;
  subscribe(_founderID: string, listener: (message: GameUIRuntimeMessage) => void): () => void { this.listener = listener; queueMicrotask(() => listener({ kind: "transport_recovered" })); return () => {}; }
}

interface AppExports { fixtureSnapshot(value: ParsedGameUISnapshot): void; fixtureSurface(value: GameUISurfaceID): void; fixtureMonotonicElapsed(value: number): void }

async function settle(): Promise<void> { for (let index = 0; index < 4; index += 1) { await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync(); } }

async function assertAxe(target: HTMLElement, label: string): Promise<void> {
  const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
  expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), label).toEqual([]);
}

function button(target: HTMLElement, text: string): HTMLButtonElement {
  const found = [...target.querySelectorAll("button")].find((candidate) => candidate.textContent?.trim() === text);
  if (!found) throw new Error(`missing button ${text}`);
  return found;
}

async function mounted(runtime = new Runtime()) {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  await settle();
  return { target, app, runtime, dispose: async () => { await unmount(app as never); target.remove(); } };
}

const arms = armFixture.cases as unknown as Record<string, GameUICosmeticsArm>;
function withShop(arm: GameUICosmeticsArm | undefined, tier: number): GameUISnapshot {
  const features = arm === undefined ? v4.features : { ...v4.features, cosmetics: arm };
  return { ...v4, features, run: { ...v4.run, tier } } as GameUISnapshot;
}
const shelf = (target: HTMLElement) => target.querySelector("[data-testid=cosmetic-shelf]");
const staticCard = (target: HTMLElement) => [...target.querySelectorAll("section.card h2")].some((node) => node.textContent?.includes("Horse Armor"));

// Cosmetic Shop v1 AC11 (host half): visibility per §7.3/OD-4, the §6
// pre-activation static card, and one acquire through runtime.intent whose
// ownership renders only from the next authoritative snapshot.
it.skipIf(!browser)("hides the shelf at T0, shows it at T1, keeps it once owned, and keeps the static card before activation", async () => {
  const runtime = new Runtime();
  runtime.current = withShop(undefined, 0);
  const trap = installNetworkTrap();
  const { target, dispose } = await mounted(runtime);
  try {
    expect(shelf(target)).toBeNull();
    expect(staticCard(target)).toBe(true);
    runtime.current = withShop(arms["locked-at-tier-0"], 0);
    runtime.listener?.({ kind: "snapshot", value: runtime.current });
    await settle();
    expect(shelf(target)).toBeNull();
    expect(staticCard(target)).toBe(false);
    runtime.current = withShop(arms["acquirable-at-tier-1"], 1);
    runtime.listener?.({ kind: "snapshot", value: runtime.current });
    await settle();
    expect(shelf(target)).not.toBeNull();
    runtime.outcome = { outcome: "applied", receipt: { intent_id: "x", outcome: "applied", founder_revision: 8, kind: "acquire_cosmetic",
      cosmetics: { owned: ["horse_armor"], equipped: {} }, event: { kind: "cosmetic_acquired.v1", payload: { cosmetic_id: "horse_armor", order_number: 1 } } } };
    const buy = shelf(target)!.querySelector<HTMLButtonElement>("button")!;
    buy.click();
    await settle();
    expect(runtime.requests.at(-1)).toMatchObject({ kind: "acquire_cosmetic", cosmetic_id: "horse_armor", expected_revision: 7 });
    expect(Object.keys(runtime.requests.at(-1)!).sort()).toEqual(["cosmetic_id", "expected_revision", "intent_id", "kind"]);
    runtime.current = withShop(arms["owned-no-wearer"], 0);
    runtime.listener?.({ kind: "snapshot", value: runtime.current });
    await settle();
    expect(shelf(target)).not.toBeNull();
    expect(shelf(target)!.textContent).toContain("Owned.");
    const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
    expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical")).toEqual([]);
    // N5: the whole shop flow issued no off-origin request and touched no payment API.
    expect(trap.violations).toEqual([]);
  } finally { trap.restore(); await dispose(); }
});

// N5 failing case: the trap itself records off-origin checkout and payment calls.
it.skipIf(!browser)("the network trap rejects off-origin checkout and PaymentRequest", async () => {
  const trap = installNetworkTrap();
  try {
    await window.fetch("https://example.invalid/checkout").catch(() => undefined);
    try { new (window as unknown as { PaymentRequest: new () => unknown }).PaymentRequest(); } catch { /* trapped */ }
    window.open("https://pay.example.invalid");
    expect(trap.violations).toEqual(["fetch https://example.invalid/checkout", "PaymentRequest", "window.open https://pay.example.invalid"]);
    expect(sameOriginAllowed("/api/v1/intents", window.location.origin)).toBe(true);
    expect(sameOriginAllowed("/checkout", window.location.origin)).toBe(false);
  } finally { trap.restore(); }
});
