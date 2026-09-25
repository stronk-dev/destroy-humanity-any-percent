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

const opportunityArm = (pending: boolean, buffs = true, saturated = false) => ({
  attended_now_ms: 2_000,
  buffs: buffs ? [{ buff_instance_id: "01986666-0000-7000-8000-00000000000b", effect_row_id: "active.production", expires_attended_ms: 6_000, selected_target: null }] : [],
  combo: { cap: "1e4", reason_key: "cap.active_combo", saturated },
  pending: pending ? { effect_row_id: "active.lucky", expires_attended_ms: 5_500, opportunity_id: "01986666-0000-7000-8000-000000000001", selected_generator_id: null } : null,
});
const withOpportunity = (arm: ReturnType<typeof opportunityArm> | null): GameUISnapshot => ({ ...v4,
  facts: v4.facts.map((fact) => fact.fact_id === "feature.active_play" ? { ...fact, value: arm !== null } : fact),
  features: { ...v4.features, ...(arm === null ? {} : { opportunity: arm }) } } as GameUISnapshot);

function regionIndex(target: HTMLElement): number {
  const desk = target.querySelector("section.desk")!;
  return [...desk.children].findIndex((child) => child.getAttribute("data-region") === "desk.region.opportunity");
}

it.skipIf(!browser)("keeps the opportunity region in a fixed Desk position and never moves focus on spawn (GS5-A1/A5)", async () => {
  const runtime = new Runtime(); runtime.current = withOpportunity(opportunityArm(false, false));
  const { target, app, dispose } = await mounted(runtime);
  try {
    const before = regionIndex(target);
    expect(before).toBeGreaterThan(0);
    const manual = target.querySelector<HTMLButtonElement>("section.manual button")!;
    manual.focus();
    app.fixtureSnapshot(withOpportunity(opportunityArm(true)));
    await settle();
    expect(regionIndex(target)).toBe(before);
    expect(document.activeElement).toBe(manual);
    const region = target.querySelector("[data-region='desk.region.opportunity']")!;
    expect(region.textContent).toContain("Lucky break");
    expect(region.textContent).toContain("4 s of attention left to claim it");
    expect(region.textContent).toContain("Production frenzy");
    expect(target.querySelector(".announcement")?.textContent).toContain("An opportunity appeared: Lucky break");
    await assertAxe(target, "opportunity region");
  } finally { await dispose(); }
});

it.skipIf(!browser)("sends no command while an opportunity waits on an idle Desk (GS5-A2)", async () => {
  const runtime = new Runtime(); runtime.current = withOpportunity(opportunityArm(true));
  const { app, dispose } = await mounted(runtime);
  try {
    app.fixtureMonotonicElapsed(60_000);
    await new Promise((resolve) => setTimeout(resolve, 400));
    await settle();
    expect(runtime.requests).toEqual([]);
  } finally { await dispose(); }
});

it.skipIf(!browser)("claims by keyboard and shows a saturated lucky payout's cap reason as text (GS5-A3/A5)", async () => {
  const runtime = new Runtime(); runtime.current = withOpportunity(opportunityArm(true, true, true));
  runtime.outcome = { outcome: "applied", receipt: { outcome: "applied", receipt: { opportunity: { opportunity_id: "01986666-0000-7000-8000-000000000001", effect_row_id: "active.lucky",
    selected_target: null, buff_instance_id: null, requested_delta: "5e3", actual_credited_delta: "1e3", saturated: true, cap_reason_key: "cap.cash", next_sampled_interval_ms: 1_000, next_opportunity_attended_ms: 9_000 } } } };
  const { target, dispose } = await mounted(runtime);
  try {
    const region = target.querySelector("[data-region='desk.region.opportunity']")!;
    expect(region.textContent).toContain("Your boosts are stacked up to the combo cap.");
    expect(region.textContent).toContain("Boost combo cap");
    const claim = button(target, "Claim");
    claim.focus(); claim.click(); await settle();
    expect(runtime.requests).toEqual([expect.objectContaining({ kind: "claim_opportunity", opportunity_id: "01986666-0000-7000-8000-000000000001", expected_revision: 1 })]);
    expect(region.textContent).toContain("The lucky break hit the cash cap. Cash cap");
    expect(region.textContent).toContain("Lucky break credited: 1e3");
    await assertAxe(target, "claimed");
  } finally { await dispose(); }
});

it.skipIf(!browser)("renders an expired claim's typed reason (GS5 error mapping)", async () => {
  const runtime = new Runtime(); runtime.current = withOpportunity(opportunityArm(true));
  runtime.outcome = { outcome: "rejected", category: "not_eligible", detail: "opportunity_expired", currentRevision: 1, sessionExpired: false };
  const { target, dispose } = await mounted(runtime);
  try {
    button(target, "Claim").click(); await settle();
    expect(target.querySelector(".intent-notice")?.textContent).toBe("Too late. That opportunity expired.");
  } finally { await dispose(); }
});

const petRow = { eligible_action_ids: ["care.feed", "care.pet"], name_key: "pet.name.server_room_cat.n04", palette_id: "pet_palette.fur_02",
  pet_id: "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa", species_id: "pet_species.server_room_cat", status_band: "normal" as const, temperament: "sassy" as const };
const withPet = (): GameUISnapshot => ({ ...v4,
  facts: [...v4.facts.map((fact) => fact.fact_id === "feature.pets" ? { ...fact, value: true } : fact)],
  features: { ...v4.features,
    pet_adoption: { pet_adoption: { cap: 1, count: 1, name_keys: ["pet.name.server_room_cat.n04"], starter_species_id: "pet_species.server_room_cat" }, pets: [petRow] },
    cosmetics: { active: true, items: [{ acquirable: false, cosmetic_id: "horse_armor", lock: null, owned: true, worn_by: [petRow.pet_id] }], wearers: [{ pet_id: petRow.pet_id, worn: "horse_armor" }] } } } as GameUISnapshot);

it.skipIf(!browser)("never offers the pet surface without an adopted pet (GS4-A2)", async () => {
  const { target, dispose } = await mounted();
  try {
    expect([...target.querySelectorAll("nav button")].map((node) => node.textContent)).not.toContain("PENDING OWNER COPY: care panel title");
  } finally { await dispose(); }
});

it.skipIf(!browser)("states unavailable care actions in text, sends Founder-scoped care, and maps rejections (GS4-A1/A3)", async () => {
  const runtime = new Runtime(); runtime.current = withPet();
  const { target, dispose } = await mounted(runtime);
  try {
    button(target, "PENDING OWNER COPY: care panel title").click(); await settle();
    const surface = target.querySelector(".pet-care")!;
    const items = [...surface.querySelectorAll(".actions li")];
    expect(items.map((item) => item.querySelector("button")!.textContent)).toEqual(["feed", "groom", "pet", "play", "rest"].map((id) => `PENDING OWNER COPY: care.${id}`));
    for (const item of items) {
      const disabled = item.querySelector("button")!.disabled;
      expect(Boolean(item.querySelector("small")?.textContent?.includes("action unavailable")), item.textContent ?? "").toBe(disabled);
    }
    expect(items.filter((item) => !item.querySelector("button")!.disabled)).toHaveLength(2);
    expect(surface.textContent).toContain("PENDING OWNER COPY: band normal");
    expect(surface.querySelector(".overlay[data-render='horse_armor']")).not.toBeNull();
    await assertAxe(target, "pet care");
    button(target, "PENDING OWNER COPY: care.feed").click(); await settle();
    expect(runtime.requests.at(-1)).toEqual(expect.objectContaining({ kind: "care_action", pet_id: petRow.pet_id, action_id: "care.feed", expected_revision: 7 }));
    expect(target.querySelector(".intent-notice")?.textContent).toBe("PENDING OWNER COPY: care applied");
    runtime.outcome = { outcome: "rejected", category: "not_eligible", detail: "cooldown", currentRevision: 7, sessionExpired: false };
    button(target, "PENDING OWNER COPY: care.pet").click(); await settle();
    expect(target.querySelector(".intent-notice")?.textContent).toBe("PENDING OWNER COPY: rejection cooldown");
  } finally { await dispose(); }
});

it.skipIf(!browser)("badges the Fiscal nav on an off-surface harvest and announces a buff start on the Desk (GS0.3 remainder)", async () => {
  const { target, runtime, dispose } = await mounted();
  try {
    const fiscalNav = () => [...target.querySelectorAll<HTMLButtonElement>("nav button")].find((node) => node.textContent?.startsWith("Earnings Calls"))!;
    runtime.listener?.({ kind: "announcement", scope: "company", value: { cursor: 11, kind: "buff_started", payload: { buff_instance_id: "01986666-0000-7000-8000-00000000000b", effect_row_id: "active.production", expires_attended_ms: 6_000 } } });
    await settle();
    expect(target.querySelector(".announcement")?.textContent).toBe("Boost started: Production frenzy");
    runtime.listener?.({ kind: "announcement", scope: "founder", value: { cursor: 12, kind: "fiscal_period_harvested", payload: { source: "automatic", credit_after: 10 } } });
    await settle();
    expect(fiscalNav().querySelector(".nav-badge")?.textContent).toBe("(harvested)");
    fiscalNav().click(); await settle();
    expect(fiscalNav().querySelector(".nav-badge")).toBeNull();
  } finally { await dispose(); }
});

// GS0.4 / AC 320 px: no garage surface forces horizontal scrolling at a 320
// CSS px viewport (WCAG 1.4.10 reflow). Measured, not assumed.
it.skipIf(!browser)("reflows the Desk, Fiscal, Meters, Trophy Case and pet surfaces at 320 CSS px without horizontal overflow", async () => {
  const { page } = await import("vitest/browser");
  await page.viewport(320, 640);
  const runtime = new Runtime();
  const full = withPet();
  runtime.current = { ...full, features: { ...full.features, opportunity: opportunityArm(true, true, true) } } as GameUISnapshot;
  const { target, dispose } = await mounted(runtime);
  try {
    const overflow = (label: string) => {
      const root = document.documentElement;
      const offenders = [...target.querySelectorAll<HTMLElement>("*")].filter((node) => node.getBoundingClientRect().right > root.clientWidth + 1)
        .map((node) => `${node.tagName.toLowerCase()}.${node.className}`).slice(0, 5);
      expect({ label, scrollWidth: root.scrollWidth <= root.clientWidth + 1, offenders }).toEqual({ label, scrollWidth: true, offenders: [] });
    };
    overflow("desk");
    for (const [nav, label] of [["Earnings Calls", "fiscal"], ["Reputation Board", "meters"], ["Trophy Case", "achievements"], ["PENDING OWNER COPY: care panel title", "pet"]] as const) {
      const control = [...target.querySelectorAll("nav button")].find((node) => node.textContent?.startsWith(nav)) as HTMLButtonElement;
      control.click(); await settle();
      overflow(label);
    }
  } finally { await dispose(); await page.viewport(1280, 720); }
});
