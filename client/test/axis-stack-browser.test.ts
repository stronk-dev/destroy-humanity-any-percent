import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import axisArm from "../../testdata/axis-stack/game-ui-arm-v1.json";
import type { GameUIAxisStackArm, GameUISnapshot } from "../src/api/generated/types";
import { parseGameUISnapshot } from "../src/game-ui/contracts";
import type { ParsedGameUISnapshot } from "../src/game-ui/contracts";
import GameUIApp from "../src/game-ui/GameUIApp.svelte";
import { decodeGameUIAnnouncement } from "../src/game-ui/events";
import type { IntentOutcome } from "../src/game-ui/intent-outcome";
import type { GameUIRuntime, GameUIRuntimeMessage } from "../src/game-ui/runtime";
import type { GameUISurfaceID } from "../src/game-ui/surface-catalog";
import type { TransportEnvelope } from "../src/transport";

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


// Clout v1 CV9/AC11 (producer → renderer): the Go projector's exact arm for a
// fixture Company v19 (TestAxisStackArmIsServerDerived golden) renders the
// axis readout and each PR row's progress/factor on the Desk; PR rows in the
// upgrade list resolve through the axis presentation rows.
function withAxis(arm: GameUIAxisStackArm | null): GameUISnapshot {
  const features = arm === null ? v4.features : { ...v4.features, axis_stack: arm };
  const upgrades = arm === null ? v4.upgrades : [...v4.upgrades,
    { cost_amount: "1e6", cost_resource_id: "company.cash", eligible: false, owned: true, upgrade_id: "upgrade.pr_intern_1" },
    { cost_amount: "5e7", cost_resource_id: "company.cash", eligible: false, owned: false, upgrade_id: "upgrade.pr_intern_2" }];
  const facts = [...v4.facts, { fact_id: "feature.axis_stack", value: arm !== null }].sort((a, b) => a.fact_id < b.fact_id ? -1 : 1);
  return parseGameUISnapshot({ ...v4, facts, features, upgrades, run: { ...v4.run, tier: 1 } }) as GameUISnapshot;
}

for (const width of [320, 1280]) {
  it.skipIf(!browser)(`coalesces re-attainment by Company transition without suppressing later notices (${width}px)`, async () => {
    const { page } = await import("vitest/browser");
    await page.viewport(width, 720);
    const runtime = new Runtime();
    runtime.current = withAxis(axisArm as GameUIAxisStackArm);
    const { target, app, dispose } = await mounted(runtime);
    const region = target.querySelector<HTMLElement>('.announcement[role="status"]')!;
    const updates: string[] = [];
    const observer = new MutationObserver(() => { updates.push(region.textContent ?? ""); });
    observer.observe(region, { childList: true, characterData: true, subtree: true });
    const send = (cursor: number, eventID: string, achievementID = "achievement.first_gate", runSeq = 1, streamID = "01985555-2222-7222-8222-222222222222") => {
      const envelope: TransportEnvelope = { v: 2, ch: `player:${v4.run.founder_id}`, kind: "event", rev: cursor,
        constants_hash: v4.constants_hash, ts: "2026-10-08T12:00:00Z", payload: { kind: "achievement_reattained.v1",
          event_id: eventID, scope: "company", rev: cursor, cursor_effect: "advance",
          payload: { achievement_id: achievementID, score_grant: 2, run_id: { company_stream_id: streamID, run_seq: runSeq } } } };
      const value = decodeGameUIAnnouncement(envelope);
      expect(value, "the existing server event must reach the announcement consumer").not.toBeUndefined();
      if (!value) throw new Error("re-attainment announcement not decoded");
      runtime.listener?.({ kind: "announcement", scope: "company", eventID, value });
    };
    try {
      const desk = target.querySelector<HTMLButtonElement>('nav button[aria-current="page"]')!;
      desk.focus();
      const before = target.querySelector("section.axis")!.textContent;
      const notice = "PENDING OWNER COPY: achievement progress restored this run";
      send(9, "first"); await settle();
      expect(region.textContent).toBe(notice);
      expect(updates).toEqual([notice]);
      send(9, "second", "achievement.generators_purchased_1"); await settle();
      send(9, "first"); await settle();
      expect(updates, "distinct events at one transition and replay produce one notice").toEqual([notice]);
      runtime.listener?.({ kind: "announcement", scope: "company", eventID: "earned", value: { cursor: 10, kind: "achievement_earned",
        payload: { achievement_id: "achievement.first_gate", condition_scope: "run", score_grant: 2,
          run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 } } } });
      await settle();
      const earned = region.textContent!;
      expect(earned).not.toBe(notice);
      send(9, "second", "achievement.generators_purchased_1"); await settle();
      expect(region.textContent, "an old transition cannot overwrite a newer announcement").toBe(earned);
      send(11, "later"); await settle();
      send(12, "next"); await settle();
      expect(updates, "identical text must still reach the live DOM for distinct transitions").toEqual([notice, earned, notice, notice]);
      app.fixtureSnapshot({ ...runtime.current, run: { ...runtime.current.run, run_seq: 2 } }); await settle();
      send(12, "old-company"); await settle();
      expect(updates).toHaveLength(4);
      send(9, "new-company", "achievement.first_gate", 2, "01985555-3333-7333-8333-333333333333"); await settle();
      expect(updates).toEqual([notice, earned, notice, notice, notice]);
      expect(target.querySelector("section.axis")!.textContent).toBe(before);
      expect(document.activeElement).toBe(desk);
      expect(runtime.requests).toEqual([]);
      expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(document.documentElement.clientWidth + 1);
    } finally { observer.disconnect(); await dispose(); await page.viewport(1280, 720); }
  });
}

it.skipIf(!browser)("renders the server-derived axis readout and PR Intern progress on the Desk", async () => {
  const runtime = new Runtime();
  runtime.current = withAxis(axisArm as unknown as GameUIAxisStackArm);
  const { target, dispose } = await mounted(runtime);
  try {
    const panel = target.querySelector("section.axis");
    expect(panel, target.textContent ?? "").not.toBeNull();
    const text = panel!.textContent ?? "";
    expect(text).toContain("Achievement attainment this run: 8 of 44");
    expect(text).toContain("Unlocks at attainment 10 (now 8)");
    // Independent expected values from the Go projector golden, not the UI's
    // formatter: rounding these multipliers to 1 hides the production effect.
    expect([...panel!.querySelectorAll("output")].map((node) => node.textContent)).toEqual(["1.2", "1.2", "1.16"]);
    const rows = [...panel!.querySelectorAll("li")].map((row) => row.getAttribute("data-upgrade"));
    expect(rows).toEqual(["upgrade.pr_intern_1", "upgrade.pr_intern_2"]);
    expect(panel!.querySelectorAll("progress")).toHaveLength(1);
    const titles = [...target.querySelectorAll("article.card h3")].map((node) => node.textContent);
    expect(titles).toEqual(expect.arrayContaining(["PR Intern", "Senior PR Intern"]));
    for (const id of ["upgrade.pr_intern_1", "upgrade.pr_intern_2", "achievement.first_gate"]) expect(target.textContent).not.toContain(id);
    await assertAxe(target, "axis panel");
  } finally { await dispose(); }
});

it.skipIf(!browser)("preserves fractional PR multipliers when authoritative attainment changes", async () => {
  const runtime = new Runtime();
  runtime.current = withAxis(axisArm as unknown as GameUIAxisStackArm);
  const { target, app, dispose } = await mounted(runtime);
  try {
    for (const [input, product, secondFactor] of [[12, "1.3e0", "1.24e0"], [44, "2.1e0", "1.88e0"]] as const) {
      const updated = structuredClone(axisArm) as GameUIAxisStackArm;
      updated.input_value = input;
      updated.product = product;
      updated.contributions[0].factor = product;
      updated.interns[0].factor = product;
      updated.interns[1].factor = secondFactor;
      app.fixtureSnapshot(withAxis(updated));
      await settle();
      expect([...target.querySelectorAll("section.axis output")].map((node) => node.textContent)).toEqual(input === 12 ? ["1.3", "1.3", "1.24"] : ["2.1", "2.1", "1.88"]);
    }
    expect(runtime.requests).toEqual([]);
  } finally { await dispose(); }
});

it.skipIf(!browser)("renders no axis panel when the pinned economy has no axis stack", async () => {
  const runtime = new Runtime();
  runtime.current = withAxis(null);
  const { target, dispose } = await mounted(runtime);
  try { expect(target.querySelector("section.axis")).toBeNull(); } finally { await dispose(); }
});

// RP-303: count/axe alone do not prove native progressbar names. Use the
// browser's role/name lookup, rather than assuming a particular labeling attr.
it.skipIf(!browser)("associates native PR progress with the correct intern before and after refresh", async () => {
  const { page } = await import("vitest/browser");
  const unowned = structuredClone(axisArm) as GameUIAxisStackArm;
  unowned.interns.forEach((intern) => { intern.owned = false; });
  unowned.contributions = [];
  unowned.product = "1e0";
  const runtime = new Runtime();
  runtime.current = withAxis(unowned);
  const { target, app, dispose } = await mounted(runtime);
  try {
    const assertProgress = (name: string, id: string, value: number, maximum: number) => {
      const found = page.getByRole("progressbar", { name, exact: true }).query();
      expect(found, `native progressbar named ${name}`).not.toBeNull();
      const bar = found as HTMLProgressElement;
      expect(bar.closest("li")?.getAttribute("data-upgrade")).toBe(id);
      expect({ value: bar.value, maximum: bar.max }).toEqual({ value, maximum });
      const descriptionIDs = (bar.getAttribute("aria-describedby") ?? "").split(/\s+/u).filter(Boolean);
      const descriptions = descriptionIDs.map((ref) => document.getElementById(ref));
      expect(descriptions.every((node) => node !== null)).toBe(true);
      expect(descriptions.some((node) => node?.textContent === "Unlocks at attainment " + maximum + " (now 8)")).toBe(true);
      expect(descriptions.some((node) => node?.id === "axis-why")).toBe(true);
    };
    expect(target.querySelectorAll("section.axis progress")).toHaveLength(2);
    assertProgress("PR Intern", "upgrade.pr_intern_1", 6, 6);
    assertProgress("Senior PR Intern", "upgrade.pr_intern_2", 8, 10);
    app.fixtureSnapshot(withAxis(axisArm as GameUIAxisStackArm));
    await settle();
    expect(target.querySelectorAll("section.axis progress")).toHaveLength(1);
    expect(page.getByRole("progressbar", { name: "PR Intern", exact: true }).query()).toBeNull();
    assertProgress("Senior PR Intern", "upgrade.pr_intern_2", 8, 10);
    expect(runtime.requests).toEqual([]);
  } finally { await dispose(); }
});

// CV9.4: opening the Codex is an actual keyboard action, not merely a
// registered copy key. The current candidate text remains explicitly unadopted.
for (const tier of [0, 1, 2]) {
  for (const activation of ["{Enter}", " "]) {
    it.skipIf(!browser)(`opens and closes PR Codex with native keyboard: tier ${tier}, ${JSON.stringify(activation)}`, async () => {
      const { page, userEvent } = await import("vitest/browser");
      await page.viewport(320, 720);
      const runtime = new Runtime();
      const current = withAxis(axisArm as GameUIAxisStackArm);
      runtime.current = parseGameUISnapshot({ ...current, run: { ...current.run, tier } });
      let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
      try {
        fixture = await mounted(runtime);
        const { target, app } = fixture;
        const details = target.querySelector<HTMLDetailsElement>("section.axis details");
        expect(details, "CV9 requires an operable PR Codex disclosure").not.toBeNull();
        const summary = details!.querySelector<HTMLElement>("summary")!;
        const content = details!.querySelector<HTMLElement>(".codex")!;
        expect(summary.textContent?.trim()).toBe("?");
        expect(summary.getAttribute("aria-label")).toBe("PENDING OWNER COPY: PR Intern help");
        // Native summary roles differ between browsers; do not override the
        // native disclosure role just to make a button-role query succeed.
        expect(summary.getAttribute("role")).toBeNull();
        expect(content.textContent?.trim()).toBe("PENDING OWNER COPY: axis-stack Codex");
        expect(details!.open).toBe(false);
        const prior = [...target.querySelectorAll<HTMLButtonElement>('section[aria-labelledby="generators-heading"] button')].at(-1)!;
        prior.focus(); await userEvent.keyboard("{Tab}");
        expect(document.activeElement).toBe(summary);
        await userEvent.keyboard(activation); await settle();
        expect(details!.open).toBe(true);
        expect(content.getBoundingClientRect().height).toBeGreaterThan(0);
        expect(document.activeElement).toBe(summary);
        // A normal authoritative update must not close the player's help.
        const refreshed = structuredClone(runtime.current);
        app.fixtureSnapshot(parseGameUISnapshot({ ...refreshed, revision: refreshed.revision + 1 })); await settle();
        expect(details!.open).toBe(true);
        expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(document.documentElement.clientWidth + 1);
        expect(target.textContent).not.toContain("codex.axis_stack");
        await assertAxe(target, "open PR Codex");
        await userEvent.keyboard(activation); await settle();
        expect(details!.open).toBe(false);
        expect(document.activeElement).toBe(summary);
        await userEvent.keyboard("{Shift>}{Tab}{/Shift}");
        expect(document.activeElement).toBe(prior);
        expect(runtime.requests).toEqual([]);
      } finally {
        try { if (fixture) await fixture.dispose(); }
        finally { await page.viewport(1280, 720); }
      }
    });
  }
}
