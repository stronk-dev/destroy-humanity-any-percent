import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it, vi } from "vitest";

import type { GameUISnapshot } from "../src/api/generated/types";
import { t } from "../src/copy";
import type { ParsedGameUISnapshot } from "../src/game-ui/contracts";
import GameUIApp from "../src/game-ui/GameUIApp.svelte";
import type { IntentOutcome } from "../src/game-ui/intent-outcome";
import type { GameUIRuntime, GameUIRuntimeMessage } from "../src/game-ui/runtime";
import type { GameUISurfaceID } from "../src/game-ui/surface-catalog";

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

it.skipIf(!browser)("unlocks one nav button per live feature fact and renders the Meters board as text (GS3)", async () => {
  const { target, dispose } = await mounted();
  try {
    const nav = [...target.querySelectorAll("nav button")].map((node) => node.textContent);
    expect(nav).toEqual(expect.arrayContaining(["Trophy Case", "Earnings Calls", "Reputation Board"]));
    button(target, "Reputation Board").click(); await settle();
    const meters = target.querySelector(".meters")!;
    expect(meters.querySelectorAll("tbody tr")).toHaveLength(5);
    expect(meters.querySelectorAll("meter")).toHaveLength(11);
    expect(meters.textContent).toContain("50 of 100");
    expect(meters.textContent).toContain("p(doom)");
    for (const cell of meters.querySelectorAll("td")) expect(cell.textContent).toMatch(/of 100.*(Low|High)/su);
    await assertAxe(target, "meters");
  } finally { await dispose(); }
});

it.skipIf(!browser)("renders earned-run, earned-career and locked achievements as distinct text (GS2-A1)", async () => {
  const { target, dispose } = await mounted();
  try {
    button(target, "Trophy Case").click(); await settle();
    const rows = [...target.querySelectorAll(".achievements li")].map((row) => row.querySelector("strong")!.textContent);
    expect(rows).toEqual(["Earned this run", "Not earned yet", "Earned in your career"]);
    expect(target.querySelector(".achievements")!.textContent).toContain("2 of 3 earned");
    expect(target.textContent).not.toContain("Clout");
    await assertAxe(target, "achievements");
  } finally { await dispose(); }
});

it.skipIf(!browser)("derives Fiscal phases at the window edges and sends Founder-scoped intents (GS1-A2/A3/A4)", async () => {
  const { target, app, runtime, dispose } = await mounted();
  try {
    button(target, "Earnings Calls").click(); await settle();
    const harvest = () => target.querySelector<HTMLButtonElement>("[aria-describedby='fiscal-harvest-curtain']")!;
    // Window edges are unit-tested exactly (fiscal-phase); here each phase is
    // bound with a margin the host's real-time tick cannot cross.
    const bindOpened = async (elapsed: number) => {
      app.fixtureSnapshot({ ...v4, features: { ...v4.features, fiscal: { ...v4.features.fiscal!, period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - elapsed } } } });
      await settle();
    };
    await bindOpened(90_000);
    expect(target.querySelector(".fiscal")!.getAttribute("data-phase")).toBe("ripening");
    expect(target.querySelector(".phase")!.textContent).toContain("opens in 0:00:1");
    expect(harvest().disabled).toBe(true);
    await bindOpened(150_000);
    expect(target.querySelector(".phase")!.textContent).toContain("50 percent chance");
    expect(harvest().disabled).toBe(false);
    await bindOpened(250_000);
    expect(target.querySelector(".phase")!.textContent).toContain("Calling now always lands");
    // Unlock rows without presentation (unlock.arcade, F12) are withheld.
    expect(target.textContent).toContain("The Pitch");
    expect(target.querySelectorAll(".fiscal li")).toHaveLength(2);
    await assertAxe(target, "fiscal");

    // Intents: bind a quarter that opened long ago so the host's monotonic
    // tick cannot move the display phase back to ripening mid-test.
    app.fixtureSnapshot({ ...v4, revision: 2, features: { ...v4.features, fiscal: { ...v4.features.fiscal!, period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - 250_000 } } } });
    await settle();
    runtime.outcome = { outcome: "rejected", category: "not_eligible", detail: "period_not_ripe", currentRevision: 7, sessionExpired: false };
    harvest().click(); await settle();
    expect(runtime.requests.at(-1)).toMatchObject({ kind: "harvest_fiscal_period", expected_revision: 7 });
    expect(target.querySelector(".intent-notice")!.textContent).toBe("It is too soon for an earnings call.");
    runtime.outcome = { outcome: "applied", receipt: { harvest_outcome: "early_failed" } };
    harvest().click(); await settle();
    expect(target.querySelector(".intent-notice")!.textContent).toContain("nothing was credited");
    button(target, "Unlock for 3").click(); await settle();
    expect(runtime.requests.at(-1)).toMatchObject({ kind: "spend_fiscal_credit", expected_revision: 7, target: { kind: "unlock", unlock_id: "minigame.pitch" } });
    button(target, "Add a level for 1").click(); await settle();
    expect(runtime.requests.at(-1)).toMatchObject({ target: { kind: "generator_level", generator_id: "generator.beige_tower", levels: 1 } });
  } finally { await dispose(); }
});

it.skipIf(!browser)("shows provisioned counts with their cap reason and owned upgrades as text on the Desk (GS6)", async () => {
  const { target, dispose } = await mounted();
  try {
    expect(target.textContent).toContain("Provisioned: 3");
    const card = [...target.querySelectorAll(".card")].find((node) => node.textContent?.includes("Provisioned: 3"))!;
    expect(card.textContent).toContain(t("generator.beige_tower.provisioned_cap", {}, "era_1995"));
    expect(target.textContent).toContain("Owned");
  } finally { await dispose(); }
});

it.skipIf(!browser)("keeps the Desk up and reports loudly when a cap reason has no copy (F10)", async () => {
  const runtime = new Runtime();
  runtime.current = { ...v4, resources: [{ ...v4.resources[0]!, cap: { amount: "1e1000", reason_key: "cap.not_in_catalog" } }] };
  const error = vi.spyOn(console, "error").mockImplementation(() => {});
  const { target, dispose } = await mounted(runtime);
  try {
    expect(target.querySelector("main")?.dataset.surface).toBe("desk");
    expect(error).toHaveBeenCalledWith(expect.stringContaining("cap.not_in_catalog"));
  } finally { error.mockRestore(); await dispose(); }
});

it.skipIf(!browser)("returns to the Desk when a mounted surface's arm goes null (GS0.5)", async () => {
  const { target, app, dispose } = await mounted();
  try {
    button(target, "Earnings Calls").click(); await settle();
    expect(target.querySelector("main")?.dataset.surface).toBe("fiscal");
    app.fixtureSnapshot({ ...v4, revision: 2, features: { ...v4.features, fiscal: null }, facts: v4.facts.map((fact) => fact.fact_id === "feature.fiscal" ? { ...fact, value: false } : fact) });
    await settle();
    expect(target.querySelector("main")?.dataset.surface).toBe("desk");
  } finally { await dispose(); }
});

it.skipIf(!browser)("shows the Pitch availability reason before any create request (GS7 MinigamesArm)", async () => {
  const runtime = new Runtime();
  const port = { current: vi.fn(async () => ({ kind: "none" as const })), create: vi.fn(), command: vi.fn(), resolve: vi.fn() };
  Object.assign(runtime, { minigame: port });
  const { target, dispose } = await mounted(runtime);
  try {
    button(target, "The Pitch").click(); await settle();
    expect(target.textContent).toContain("Not open yet. Unlock it with Investor Confidence");
    expect(port.create).not.toHaveBeenCalled();
    await assertAxe(target, "pitch availability");
  } finally { await dispose(); }
});

it.skipIf(!browser)("announces an earned achievement once per cursor and badges meter changes off-surface (GS0.3/GS2-A2/GS3)", async () => {
  const { target, runtime, dispose } = await mounted();
  try {
    const runID = { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 };
    const earned = { kind: "announcement" as const, scope: "company" as const, value: { cursor: 9, kind: "achievement_earned" as const,
      payload: { achievement_id: "achievement.generators_purchased_1", condition_scope: "run" as const, run_id: runID, score_grant: 1 } } };
    runtime.listener!(earned); await settle();
    const region = target.querySelector(".announcement")!;
    const first = region.textContent!;
    expect(first).toMatch(/^Achievement earned: /u);

    const band = { kind: "announcement" as const, scope: "company" as const, value: { cursor: 10, kind: "meter_band_changed" as const,
      payload: { direction: "up" as const, from_band: "low", meter_id: "doom.probability", run_id: runID, to_band: "high", value_after: 71, value_before: 69 } } };
    runtime.listener!(band); await settle();
    const metersNav = [...target.querySelectorAll("nav button")].find((node) => node.textContent?.includes("Reputation Board")) as HTMLButtonElement;
    expect(metersNav.textContent).toContain("(changed)");
    metersNav.click(); await settle();
    expect(metersNav.textContent).not.toContain("(changed)");
    runtime.listener!({ ...band, value: { ...band.value, cursor: 11 } }); await settle();
    expect(target.querySelector(".announcement")!.textContent).toBe("p(doom) is now High.");
    // Replaying cursor 9 after reconnect must announce nothing.
    runtime.listener!(earned); await settle();
    expect(target.querySelector(".announcement")!.textContent).toBe("p(doom) is now High.");
  } finally { await dispose(); }
});
