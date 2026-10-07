import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it, vi } from "vitest";

import type { GameUISnapshot } from "../src/api/generated/types";
import { t, type CopyKey } from "../src/copy";
import { parseGameUISnapshot, type ParsedGameUISnapshot } from "../src/game-ui/contracts";
import { decodeGameUIAnnouncement, decodeGameUIEvent } from "../src/game-ui/events";
import { FEATURES_PRESENTATION } from "../src/game-ui/features-presentation";
import GameUIApp from "../src/game-ui/GameUIApp.svelte";
import FiscalSurface from "../src/game-ui/FiscalSurface.svelte";
import PetCareSurface from "../src/game-ui/pet/PetCareSurface.svelte";
import { GameUIRequestError, type IntentOutcome } from "../src/game-ui/intent-outcome";
import { createBrowserGameUIRuntime, type GameUIRuntime, type GameUIRuntimeMessage } from "../src/game-ui/runtime";
import type { GameUISurfaceID } from "../src/game-ui/surface-catalog";
import { formatAmount } from "../src/ui/amount-format";
import { decodeTransportEnvelope } from "../src/transport";

const browser = typeof document !== "undefined";
const NOW = 1_800_000_000_000;

declare module "vitest" {
  interface TaskMeta { firstReadObservations?: Record<string, unknown>[] }
}

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

// GS2-A1/A5: visible, read-only fixture populations, not a persisted earn,
// assistive-technology session, actual 400% zoom or all-engine acceptance.
for (const population of ["populated", "empty", "presentation-error"] as const) {
  for (const width of [320, 1280] as const) {
    for (const activation of ["{Enter}", " "] as const) {
      it.skipIf(!browser)(`GS2-A5 visible Trophy Case ${population}/${width}/${activation === " " ? "Space" : "Enter"}`, async ({ annotate }) => {
        const { page, userEvent } = await import("vitest/browser");
        await page.viewport(width, 720);
        const original = v4.features.achievements!;
        const arm = population === "empty" ? { rows: [], score: { run: 0, lifetime: 0 } }
          : population === "presentation-error" ? { ...original, rows: original.rows.map((row, index) => index === 0 ? { ...row, copy_key: "achievement.unregistered" } : row) }
          : original;
        const runtime = new Runtime();
        runtime.current = parseGameUISnapshot({ ...v4, features: { ...v4.features, achievements: arm } });
        const diagnostic = vi.spyOn(console, "error").mockImplementation(() => {});
        let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
        try {
          fixture = await mounted(runtime);
          const { target } = fixture;
          const nav = button(target, t("surface.achievements.title", {}, "era_1995"));
          const deskNav = button(target, t("surface.desk.title", {}, "era_1995"));
          const nextNav = button(target, t("surface.fiscal.title", {}, "era_1995"));
          nav.focus(); await userEvent.keyboard(activation); await settle();
          expect(document.activeElement).toBe(nav);
          expect(nav.getAttribute("aria-current")).toBe("page");
          const main = target.querySelector<HTMLElement>("main.game-ui")!;
          const panel = target.querySelector<HTMLElement>(".achievements")!;
          expect(main.dataset.surface).toBe("achievements"); expect(panel).not.toBeNull();
          const visible = (node: HTMLElement) => {
            expect(node).not.toBeNull();
            const rectangle = node.getBoundingClientRect();
            expect(rectangle.width).toBeGreaterThan(0); expect(rectangle.height).toBeGreaterThan(0);
            for (let ancestor: HTMLElement | null = node; ancestor; ancestor = ancestor.parentElement) {
              const style = getComputedStyle(ancestor);
              expect(style.display, `${node.tagName} ancestor display`).not.toBe("none");
              expect(style.visibility).toBe("visible"); expect(Number(style.opacity)).toBeGreaterThan(0);
              expect(["hidden", "clip"], "clipping cannot stand in for readable reflow").not.toContain(style.overflowX);
              if (ancestor === main) break;
            }
          };
          const exact = (parent: Element, selector: string, text: string) => {
            const nodes = [...parent.querySelectorAll<HTMLElement>(selector)].filter((node) => node.textContent === text);
            expect(nodes, `visible exact ${text}`).toHaveLength(1); visible(nodes[0]!);
            return nodes[0]!;
          };
          const heading = exact(panel, "h1", t("surface.achievements.title", {}, "era_1995"));
          expect(heading.id).toBe(panel.getAttribute("aria-labelledby"));
          expect(heading.getAttribute("tabindex")).toBe("-1");
          expect([...panel.querySelectorAll<HTMLElement>("button,input,select,textarea,a[href],[tabindex]")]
            .filter((node) => node.tabIndex >= 0)).toEqual([]);
          if (population === "presentation-error") {
            exact(panel, '[role="alert"]', t("common.surface_error", {}, "era_1995"));
            expect(panel.querySelectorAll("li,ul,strong,small")).toHaveLength(0);
            expect(panel.textContent).not.toContain("achievement.unregistered");
            expect(diagnostic.mock.calls).toEqual([["game UI invariant: achievements surface presentation unavailable"]]);
          } else {
            expect(panel.querySelector('[role="alert"]')).toBeNull();
            exact(panel, "p", t("achievements.score_frame", { run: arm.score.run, lifetime: arm.score.lifetime }, "era_1995"));
            const rows = [...panel.querySelectorAll<HTMLElement>("li")];
            expect(rows).toHaveLength(population === "empty" ? 0 : 3);
            if (population === "empty") {
              exact(panel, "p", t("achievements.empty", {}, "era_1995"));
              expect(panel.querySelectorAll("ul,strong,small")).toHaveLength(0);
            } else {
              exact(panel, "p", t("achievements.progress_frame", { earned: 2, total: 3 }, "era_1995"));
              const states: readonly CopyKey[] = ["achievements.state.earned_run", "achievements.state.locked", "achievements.state.earned_lifetime"];
              for (const [index, row] of rows.entries()) {
                const source = original.rows[index]!;
                exact(row, "h2", t(source.copy_key as CopyKey, {}, "era_1995"));
                exact(row, "strong", t(states[index]!, {}, "era_1995"));
                exact(row, "span", t(source.condition_scope === "run" ? "achievements.scope.run" : "achievements.scope.career", {}, "era_1995"));
                exact(row, "span", t("achievements.grant_frame", { score: source.score_grant }, "era_1995"));
                expect(row.querySelectorAll("small")).toHaveLength(index === 2 ? 1 : 0);
                if (index === 2) exact(row, "small", t("achievement.possession_warning", {}, "era_1995"));
                if (width === 320 && index > 0) expect(row.getBoundingClientRect().top).toBeGreaterThanOrEqual(rows[index - 1]!.getBoundingClientRect().bottom);
              }
            }
            expect(diagnostic).not.toHaveBeenCalled();
          }
          const root = document.documentElement;
          expect(root.clientWidth).toBe(width); expect(root.scrollWidth).toBeLessThanOrEqual(width + 1);
          for (const node of [main, panel, ...panel.querySelectorAll<HTMLElement>("*")]) {
            const rectangle = node.getBoundingClientRect();
            expect(rectangle.left).toBeGreaterThanOrEqual(-1); expect(rectangle.right).toBeLessThanOrEqual(width + 1);
            if (node.clientWidth > 0) expect(node.scrollWidth).toBeLessThanOrEqual(node.clientWidth + 1);
          }
          const observation = { population, width, activation, root_client: root.clientWidth,
            root_scroll: root.scrollWidth, panel_client: panel.clientWidth, panel_scroll: panel.scrollWidth };
          expect(document.activeElement).toBe(nav);
          await userEvent.keyboard("{Tab}"); expect(document.activeElement).toBe(nextNav);
          await userEvent.keyboard("{Shift>}{Tab}{/Shift}"); expect(document.activeElement).toBe(nav);
          await userEvent.keyboard("{Shift>}{Tab}{/Shift}"); expect(document.activeElement).toBe(deskNav);
          await userEvent.keyboard(activation); await settle();
          expect(main.dataset.surface).toBe("desk"); expect(document.activeElement).toBe(deskNav);
          expect(runtime.requests).toEqual([]);
          // Audit after the complete native navigation path so the scan is
          // not interleaved with gameplay keyboard actions.
          await userEvent.keyboard("{Tab}"); expect(document.activeElement).toBe(nav);
          await userEvent.keyboard(activation); await settle();
          expect(main.dataset.surface).toBe("achievements"); expect(document.activeElement).toBe(nav);
          await assertAxe(target, `GS2-A5 ${population}/${width}`);
          await annotate(JSON.stringify({ ...observation, read_only_intents: runtime.requests.length }), "GS2-A5-fixture-observation");
        } finally {
          try { if (fixture) await fixture.dispose(); }
          finally { diagnostic.mockRestore(); await page.viewport(1280, 720); }
        }
      });
    }
  }
}

// GS0.5/GS2: shared-state public fixtures, not a real network outage or earn.
function sharedStateVisibleText(parent: Element, selector: string, text: string): HTMLElement {
  const nodes = [...parent.querySelectorAll<HTMLElement>(selector)].filter((node) => node.textContent === text);
  expect(nodes, `shared-state visible exact ${text}`).toHaveLength(1);
  const node = nodes[0]!, rectangle = node.getBoundingClientRect(), style = getComputedStyle(node);
  expect(rectangle.width).toBeGreaterThan(0); expect(rectangle.height).toBeGreaterThan(0);
  expect(style.display).not.toBe("none"); expect(style.visibility).toBe("visible"); expect(Number(style.opacity)).toBeGreaterThan(0);
  return node;
}

for (const width of [320, 1280] as const) {
  it.skipIf(!browser)(`GS2 shared-state initial held read ${width}`, async () => {
    const { page } = await import("vitest/browser"); await page.viewport(width, 720);
    const runtime = new Runtime();
    let resolveRead!: (snapshot: ParsedGameUISnapshot) => void;
    const held = new Promise<ParsedGameUISnapshot>((resolve) => { resolveRead = resolve; });
    const read = vi.spyOn(runtime, "snapshot").mockReturnValue(held);
    let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
    try {
      fixture = await mounted(runtime); const { target } = fixture;
      expect(read).toHaveBeenCalledTimes(1);
      sharedStateVisibleText(target, "h1", t("surface.desk.title", {}, "era_1995"));
      sharedStateVisibleText(target, '[role="status"]', t("common.loading", {}, "era_1995"));
      expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
      expect(target.querySelectorAll("button,input,li,meter")).toHaveLength(0);
      expect(runtime.requests).toEqual([]);
      await assertAxe(target, "held first authoritative read");
      resolveRead(parseGameUISnapshot(structuredClone(v4))); await settle();
      expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
      expect(target.textContent).not.toContain(t("common.loading", {}, "era_1995"));
      expect(button(target, t("manual.click.title", {}, "era_1995"))).toBeDefined();
      expect(runtime.requests).toEqual([]);
    } finally {
      resolveRead(parseGameUISnapshot(structuredClone(v4))); await settle();
      try { if (fixture) await fixture.dispose(); }
      finally { read.mockRestore(); await page.viewport(1280, 720); }
    }
  });

  for (const [label, message] of [
    ["recovering", { kind: "transport_recovering" }],
    ["closed", { kind: "transport_closed" }],
    ["resync", { kind: "system", value: { kind: "resync_required" } }],
    ["restart", { kind: "system", value: { kind: "server_restarting", resume_after_ms: 1_000 } }],
  ] as const) {
    it.skipIf(!browser)(`GS2 shared-state ${label} keeps authoritative score ${width}`, async () => {
      const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
      const runtime = new Runtime(), { target, dispose } = await mounted(runtime);
      try {
        const nav = button(target, t("surface.achievements.title", {}, "era_1995"));
        nav.focus(); await userEvent.keyboard("{Enter}"); await settle();
        const assertScore = (run: number, states: readonly string[]) => {
          const panel = target.querySelector<HTMLElement>(".achievements")!; expect(panel).not.toBeNull();
          sharedStateVisibleText(panel, "p", t("achievements.score_frame", { run, lifetime: 5 }, "era_1995"));
          const rows = [...panel.querySelectorAll<HTMLElement>("strong")];
          expect(rows.map((row) => row.textContent)).toEqual(states);
          for (const text of states) sharedStateVisibleText(panel, "strong", text);
          expect(panel.querySelectorAll("button,input,select,textarea")).toHaveLength(0);
          expect(runtime.requests).toEqual([]);
        };
        const before = ["Earned this run", "Not earned yet", "Earned in your career"];
        assertScore(2, before);
        runtime.listener?.(message); await settle();
        sharedStateVisibleText(target, "p", t("common.stale_note", {}, "era_1995"));
        assertScore(2, before); expect(document.activeElement).toBe(nav);
        await new Promise((resolve) => setTimeout(resolve, 350)); await settle();
        assertScore(2, before); expect(document.activeElement).toBe(nav);
        runtime.listener?.({ kind: "transport_recovered" }); await settle();
        expect(target.textContent).not.toContain(t("common.stale_note", {}, "era_1995"));
        assertScore(2, before);
        const updated = parseGameUISnapshot({ ...v4, revision: 2, features: { ...v4.features,
          achievements: { ...v4.features.achievements!, score: { run: 3, lifetime: 5 },
            rows: v4.features.achievements!.rows.map((row, index) => index === 1 ? { ...row, earned: "run" } : row) } } });
        runtime.current = updated; runtime.listener?.({ kind: "snapshot", value: updated }); await settle();
        const panel = target.querySelector<HTMLElement>(".achievements")!;
        sharedStateVisibleText(panel, "p", t("achievements.score_frame", { run: 3, lifetime: 5 }, "era_1995"));
        expect([...panel.querySelectorAll("strong")].map((row) => row.textContent)).toEqual(["Earned this run", "Earned this run", "Earned in your career"]);
        expect(document.activeElement).toBe(nav); expect(runtime.requests).toEqual([]);
        expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(width + 1);
        await assertAxe(target, `GS2 shared-state ${label} recovery`);
      } finally { try { await dispose(); } finally { await page.viewport(1280, 720); } }
    });
  }

  it.skipIf(!browser)(`GS2 shared-state transport not ready alone ${width}`, async () => {
    const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
    const runtime = new Runtime(); runtime.current = parseGameUISnapshot(structuredClone(v4));
    const subscribe = vi.spyOn(runtime, "subscribe").mockImplementation((_founder, listener) => { runtime.listener = listener; return () => {}; });
    let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
    try {
      fixture = await mounted(runtime); const { target } = fixture;
      const nav = button(target, t("surface.achievements.title", {}, "era_1995")); nav.focus();
      await userEvent.keyboard("{Enter}"); await settle();
      sharedStateVisibleText(target, "p", t("common.stale_note", {}, "era_1995"));
      sharedStateVisibleText(target, ".achievements p", t("achievements.score_frame", { run: 2, lifetime: 5 }, "era_1995"));
      expect(document.activeElement).toBe(nav); expect(runtime.requests).toEqual([]);
      runtime.listener?.({ kind: "transport_recovered" }); await settle();
      expect(target.textContent).not.toContain(t("common.stale_note", {}, "era_1995"));
      expect(document.activeElement).toBe(nav); expect(runtime.requests).toEqual([]);
    } finally { try { if (fixture) await fixture.dispose(); } finally { subscribe.mockRestore(); await page.viewport(1280, 720); } }
  });

  for (const retainFact of [true, false]) {
    it.skipIf(!browser)(`GS2 shared-state null arm focuses Desk ${width}/${retainFact ? "retained-fact" : "removed-fact"}`, async () => {
      const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
      const { target, runtime, dispose } = await mounted();
      try {
        const nav = button(target, t("surface.achievements.title", {}, "era_1995")); nav.focus();
        await userEvent.keyboard(" "); await settle(); expect(document.activeElement).toBe(nav);
        const lost = parseGameUISnapshot({ ...v4, revision: 2, features: { ...v4.features, achievements: null },
          facts: v4.facts.map((fact) => fact.fact_id === "feature.achievements" ? { ...fact, value: retainFact } : fact) });
        runtime.current = lost; runtime.listener?.({ kind: "snapshot", value: lost }); await settle();
        expect(target.querySelector(".achievements")).toBeNull();
        expect(target.querySelector("main")?.dataset.surface).toBe("desk");
        const heading = sharedStateVisibleText(target, "h1", t("surface.desk.title", {}, "era_1995"));
        expect(document.activeElement).toBe(heading); expect(heading.getAttribute("tabindex")).toBe("-1");
        expect(button(target, t("manual.click.title", {}, "era_1995"))).toBeDefined();
        const settings = button(target, t("surface.settings.title", {}, "era_1995")); settings.focus();
        await userEvent.keyboard("{Enter}"); await settle(); expect(document.activeElement).toBe(settings);
        runtime.listener?.({ kind: "snapshot", value: { ...lost, revision: 3 } }); await settle();
        expect(document.activeElement).toBe(settings); expect(target.querySelector("main")?.dataset.surface).toBe("settings");
        expect(runtime.requests).toEqual([]); await assertAxe(target, "post-arm-loss user navigation");
      } finally { try { await dispose(); } finally { await page.viewport(1280, 720); } }
    });
  }

  for (const selection of ["settings", "desk"] as const) {
  it.skipIf(!browser)(`GS2 shared-state newer user selection cancels forced focus ${width}/${selection}`, async () => {
    const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
    const { target, runtime, dispose } = await mounted();
    try {
      const nav = button(target, t("surface.achievements.title", {}, "era_1995")); nav.focus();
      await userEvent.keyboard("{Enter}"); await settle();
      runtime.listener?.({ kind: "snapshot", value: parseGameUISnapshot({ ...v4, revision: 2, features: { ...v4.features, achievements: null } }) });
      flushSync();
      // Controlled synchronous DOM delivery before the queued focus task:
      // exercises the real nav handler, not physical human race timing.
      const choice = button(target, t(selection === "settings" ? "surface.settings.title" : "surface.desk.title", {}, "era_1995")); choice.focus(); choice.click();
      await settle(); expect(target.querySelector("main")?.dataset.surface).toBe(selection);
      expect(document.activeElement).toBe(choice); expect(runtime.requests).toEqual([]);
    } finally { try { await dispose(); } finally { await page.viewport(1280, 720); } }
  });
  }
}

// RP-342 measurement, not acceptance of the observed failed-read display.
// Actual runtime/Response.json/decoder; injected HTTP, not real network/auth.
for (const firstRead of ["healthy", "network", "401", "503", "json", "arm", "legacy"] as const) {
  for (const width of [320, 1280] as const) {
    it.skipIf(!browser)(`RP-342 first-read observation ${firstRead}/${width}`, async ({ task }) => {
      const { page } = await import("vitest/browser"); await page.viewport(width, 720);
      const credentialDocument = JSON.stringify({ accessToken: "synthetic-access", refreshToken: "synthetic-refresh", accountID: "synthetic-account", recoveryCode: "synthetic-recovery" });
      const values = new Map([["cloud-clicker.credentials.v1", credentialDocument]]);
      const storage = { getItem: (key: string) => values.get(key) ?? null,
        setItem: (key: string, value: string) => { values.set(key, value); }, removeItem: (key: string) => { values.delete(key); } };
      let releaseFirst!: () => void; const held = new Promise<void>((resolve) => { releaseFirst = resolve; });
      let reads = 0, sockets = 0;
      const requests: { path: string; method: string }[] = [];
      const fetcher: typeof fetch = async (input, options) => {
        const path = String(input); requests.push({ path, method: options?.method ?? "GET" });
        if (path === "/api/v1/garden/current") return Response.json({ kind: "inactive" });
        if (path !== "/api/v1/founder/state") throw new Error("unexpected research request");
        reads += 1; if (reads === 1) await held;
        if (firstRead === "401") return Response.json({ category: "unauthorized", detail: "access_token" }, { status: 401 });
        if (reads > 1 || firstRead === "healthy") return Response.json(structuredClone(v4));
        if (firstRead === "network") throw new TypeError("synthetic network rejection");
        if (firstRead === "503") return Response.json({ category: "not_configured", detail: "server" }, { status: 503 });
        if (firstRead === "json") return new Response("{", { status: 200, headers: { "Content-Type": "application/json" } });
        if (firstRead === "arm") return Response.json({ ...v4, features: { ...v4.features, achievements: { ...v4.features.achievements, score: { run: -1, lifetime: 5 } } } });
        const { features: _features, ...legacy } = v4;
        const legacyWire = { ...legacy, schema_version: 3, generators: legacy.generators.map(({ provision_cap: _cap, ...row }) => row) };
        expect(parseGameUISnapshot(legacyWire).schema_version).toBe(3);
        return Response.json(legacyWire);
      };
      const runtime = createBrowserGameUIRuntime(storage, fetcher, crypto, () => {
        sockets += 1;
        return { addEventListener() {}, close() {}, send() {} } as unknown as WebSocket;
      }, { protocol: "http:", host: "research.invalid" });
      const read = vi.spyOn(runtime, "snapshot"); const diagnostic = vi.spyOn(console, "error").mockImplementation(() => {});
      const target = document.createElement("div"); document.body.append(target);
      const app = mount(GameUIApp, { target, props: { runtime } });
      const observations: Record<string, unknown>[] = []; task.meta.firstReadObservations = observations;
      const observe = (phase: string) => {
        const visible = (copy: CopyKey) => [...target.querySelectorAll<HTMLElement>("p")].some((node) => node.textContent === t(copy, {}, "era_1995") && node.getBoundingClientRect().height > 0 && getComputedStyle(node).visibility === "visible");
        const row = { firstRead, width, phase, reads, sockets, busy: target.querySelector("main")?.getAttribute("aria-busy"),
          loading: visible("common.loading"), offline: visible("settings.save_status.offline"), surfaceError: visible("common.surface_error"),
          alerts: target.querySelectorAll('[role="alert"]').length, controls: target.querySelectorAll("button,input").length,
          diagnostics: diagnostic.mock.calls.length, credentialsRetained: values.get("cloud-clicker.credentials.v1") === credentialDocument };
        observations.push(row); return row;
      };
      try {
        await settle(); expect(reads).toBe(1); expect(sockets).toBe(0);
        sharedStateVisibleText(target, '[role="status"]', t("common.loading", {}, "era_1995"));
        expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
        expect(target.querySelectorAll("button,input")).toHaveLength(0);
        releaseFirst();
        if (firstRead === "healthy") await expect(read.mock.results[0]!.value).resolves.toMatchObject({ schema_version: 4 });
        else await expect(read.mock.results[0]!.value).rejects.toThrow(firstRead === "legacy" ? /schema v4/u : undefined);
        await settle(); await new Promise((resolve) => setTimeout(resolve, 350)); await settle();
        const settled = observe("settled"); expect(settled.busy).toBe("false"); expect(reads).toBe(1);
        if (firstRead === "healthy") {
          expect(settled.loading).toBe(false); expect(settled.controls).toBeGreaterThan(0); expect(sockets).toBe(1);
        } else {
          expect(sockets).toBe(0);
          // Existing listener, explicit event dispatch: NOT physical hide/resume.
          document.dispatchEvent(new Event("visibilitychange")); await settle();
          expect(reads).toBe(2);
          if (firstRead === "401") await expect(read.mock.results[1]!.value).rejects.toThrow(/401/u);
          else await expect(read.mock.results[1]!.value).resolves.toMatchObject({ schema_version: 4, run: { founder_id: v4.run.founder_id } });
          await settle(); const recovered = observe("lifecycle");
          if (firstRead !== "401") { expect(recovered.loading).toBe(false); expect(recovered.controls).toBeGreaterThan(0); expect(sockets).toBe(1); }
        }
        expect(values.get("cloud-clicker.credentials.v1")).toBe(credentialDocument);
        expect(requests.every((request) => request.method === "GET" && ["/api/v1/founder/state", "/api/v1/garden/current"].includes(request.path))).toBe(true);
        await assertAxe(target, "first-read research population");
      } finally {
        releaseFirst(); await settle();
        try { await unmount(app); } finally { target.remove(); read.mockRestore(); diagnostic.mockRestore(); await page.viewport(1280, 720); }
      }
    });
  }
}

// GS0.5/0.6: the other existing shared-effect consumers. Public snapshot
// delivery, not lifecycle preemption, real service or manual AT evidence.
for (const context of [
  { surface: "fiscal", arm: "fiscal", fact: "feature.fiscal", title: "surface.fiscal.title", panel: ".fiscal" },
  { surface: "meters", arm: "meters", fact: "feature.meters", title: "surface.meters.title", panel: ".meters" },
  { surface: "reputation_tree", arm: "reputation", fact: "feature.reputation_tree", title: "reputation_tree.title", panel: ".reputation" },
] as const) {
  for (const width of [320, 1280] as const) {
    for (const retainFact of [true, false]) {
      it.skipIf(!browser)(`GS0 forced-context ${context.surface}/${width}/${retainFact ? "retained-fact" : "removed-fact"}`, async () => {
        const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
        const initial: GameUISnapshot = structuredClone(v4);
        initial.facts.push({ fact_id: "feature.reputation_tree", value: true });
        initial.features.reputation = {
          available: 4, bonus_factor_next_run: "1e0", bonus_factor_this_run: null,
          level: 4, per_level_ppm: 10_000, spent: 0, unlock_ppm: 0,
          nodes: [{ node_id: "reputation.unlock.p05", body_key: "reputation_tree.node.unlock_p05.body",
            title_key: "reputation_tree.node.unlock_p05.title", cost: 1, kind: "bonus_unlock", requires: [], state: "available" }],
        };
        const runtime = new Runtime(); runtime.current = parseGameUISnapshot(initial);
        let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
        try {
          fixture = await mounted(runtime); const { target } = fixture;
          const nav = button(target, t(context.title, {}, "era_1995")); nav.focus();
          await userEvent.keyboard("{Enter}"); await settle();
          expect(target.querySelector("main")?.dataset.surface).toBe(context.surface);
          expect(target.querySelector(context.panel)).not.toBeNull(); expect(document.activeElement).toBe(nav);
          const wire = structuredClone(initial); wire.revision = 2;
          wire.features[context.arm] = null;
          wire.facts = wire.facts.map((fact) => fact.fact_id === context.fact ? { ...fact, value: retainFact } : fact);
          const lost = parseGameUISnapshot(wire); runtime.current = lost;
          runtime.listener?.({ kind: "snapshot", value: lost }); await settle();
          expect(target.querySelector(context.panel)).toBeNull();
          expect(target.querySelector("main")?.dataset.surface).toBe("desk");
          const heading = sharedStateVisibleText(target, "h1", t("surface.desk.title", {}, "era_1995"));
          expect(document.activeElement).toBe(heading); expect(heading.getAttribute("tabindex")).toBe("-1");
          const settings = button(target, t("surface.settings.title", {}, "era_1995")); settings.focus();
          await userEvent.keyboard("{Enter}"); await settle();
          runtime.listener?.({ kind: "snapshot", value: { ...lost, revision: 3 } }); await settle();
          expect(target.querySelector("main")?.dataset.surface).toBe("settings");
          expect(document.activeElement).toBe(settings); expect(runtime.requests).toEqual([]);
          await assertAxe(target, `${context.surface} forced-context return`);
        } finally { try { if (fixture) await fixture.dispose(); } finally { await page.viewport(1280, 720); } }
      });
    }
  }
}

// RP-330: real semantic layout over decoder-admitted public fixtures, not
// production/persisted values or assistive-technology evidence.
const semanticMeterRows = v4.features.meters!.meters.map((row, index) => meterRow(row.meter_id, index * 7));
function semanticMeterSnapshot(rows = semanticMeterRows): ParsedGameUISnapshot {
  return parseGameUISnapshot({ ...v4, features: { ...v4.features, meters: { meters: rows } } });
}

async function settleMeterLayout(): Promise<void> {
  // Native media-query change events run at rendering time, not merely after
  // the driver's viewport RPC. Observe one real frame, then flush Svelte.
  await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
  await settle();
}

async function assertSemanticMeters(target: HTMLElement, rows: typeof semanticMeterRows, narrow: boolean): Promise<void> {
  const surface = target.querySelector(".meters")!;
  const value = (element: Element, row: (typeof rows)[number]) => {
    const meter = element.querySelector("meter")!;
    expect(meter).not.toBeNull();
    expect([meter.min, meter.max, meter.value]).toEqual([row.min, row.max, row.value]);
    expect([...element.querySelectorAll("span")].map((node) => node.textContent)).toEqual([
      t("meters.value_frame", { value: row.value, max: row.max }, "era_1995"),
      t(FEATURES_PRESENTATION.meterBands.get(row.band_id)!, {}, "era_1995"),
    ]);
  };
  expect(surface.querySelectorAll("meter")).toHaveLength(rows.length);
  if (narrow) {
    expect(surface.querySelectorAll("table")).toHaveLength(0);
    const pairs = [...surface.querySelectorAll<HTMLElement>("dl > div")];
    expect(pairs).toHaveLength(rows.length);
    expect(pairs.map((pair) => pair.dataset.meterId).sort()).toEqual(rows.map((row) => row.meter_id));
    for (const row of rows) {
      const pair = pairs.find((candidate) => candidate.dataset.meterId === row.meter_id)!;
      expect([...pair.children].map((child) => child.tagName)).toEqual(["DT", "DD"]);
      const term = pair.querySelector("dt")!, description = pair.querySelector("dd")!;
      const presentation = FEATURES_PRESENTATION.trustMeters.get(row.meter_id);
      const label = presentation
        ? `${t(presentation.constituency_key, {}, "era_1995")} ${t(presentation.axis_key, {}, "era_1995")}`
        : t(FEATURES_PRESENTATION.doomMeter.title_key, {}, "era_1995");
      expect(term.textContent?.replace(/\s+/gu, " ").trim()).toBe(label);
      expect(term.id).not.toBe("");
      expect([...target.querySelectorAll("[id]")].filter((node) => node.id === term.id)).toHaveLength(1);
      const meter = description.querySelector("meter")!;
      expect(meter).not.toBeNull();
      expect(document.getElementById(meter.getAttribute("aria-labelledby")!)).toBe(term);
      value(description, row);
    }
  } else {
    expect(surface.querySelectorAll("dl")).toHaveLength(0);
    expect(surface.querySelectorAll("table")).toHaveLength(1);
    expect(surface.querySelectorAll("tbody tr")).toHaveLength(5);
    expect(surface.querySelectorAll("td")).toHaveLength(10);
    expect(surface.querySelectorAll('thead th[scope="col"]')).toHaveLength(3);
    expect(surface.querySelectorAll('tbody th[scope="row"]')).toHaveLength(5);
    expect([...surface.querySelectorAll("thead th")].map((node) => node.textContent)).toEqual([
      t("meters.constituency_label", {}, "era_1995"), t("meters.axis.standing", {}, "era_1995"), t("meters.axis.grievance", {}, "era_1995"),
    ]);
    for (const row of rows) {
      const presentation = FEATURES_PRESENTATION.trustMeters.get(row.meter_id);
      if (presentation) {
        const tableRow = [...surface.querySelectorAll("tbody tr")].find((node) => node.querySelector("th")!.textContent === t(presentation.constituency_key, {}, "era_1995"))!;
        value(tableRow.querySelectorAll("td")[presentation.axis_key === "meters.axis.standing" ? 0 : 1]!, row);
      } else {
        expect(surface.querySelector(".doom h2")!.textContent).toBe(t(FEATURES_PRESENTATION.doomMeter.title_key, {}, "era_1995"));
        value(surface.querySelector(".doom")!, row);
      }
    }
  }
  const root = document.documentElement;
  expect(root.scrollWidth).toBeLessThanOrEqual(root.clientWidth + 1);
  expect([...target.querySelectorAll("*")].filter((node) => node.getBoundingClientRect().right > root.clientWidth + 1)).toEqual([]);
  await assertAxe(target, narrow ? "semantic narrow meters" : "wide meter table");
}

// GS3/GS0.5: exact last-authoritative values and stale disclosure, not a
// real-server reconnect or elapsed-time meter prediction implementation.
for (const width of [320, 1280] as const) {
  for (const [label, message] of [
    ["recovering", { kind: "transport_recovering" }],
    ["closed", { kind: "transport_closed" }],
    ["resync", { kind: "system", value: { kind: "resync_required" } }],
    ["restart", { kind: "system", value: { kind: "server_restarting", resume_after_ms: 1_000 } }],
    ["not-ready-alone", null],
  ] as const) {
    it.skipIf(!browser)(`GS3 reconnect ${label} retains committed meters ${width}`, async () => {
      const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
      const runtime = new Runtime(); runtime.current = semanticMeterSnapshot();
      const subscribe = message === null ? vi.spyOn(runtime, "subscribe").mockImplementation((_founder, listener) => { runtime.listener = listener; return () => {}; }) : undefined;
      let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
      try {
        fixture = await mounted(runtime); const { target } = fixture;
        const nav = button(target, t("surface.meters.title", {}, "era_1995")); nav.focus();
        await userEvent.keyboard("{Enter}"); await settleMeterLayout();
        const assertValues = async (rows: typeof semanticMeterRows) => {
          await assertSemanticMeters(target, rows, width < 480);
          expect(target.querySelector(".meters")!.querySelectorAll("button,input,select,textarea")).toHaveLength(0);
          expect(document.activeElement).toBe(nav); expect(runtime.requests).toEqual([]);
        };
        await assertValues(semanticMeterRows);
        if (message !== null) runtime.listener?.(message);
        await settle();
        sharedStateVisibleText(target, "p", t("common.stale_note", {}, "era_1995"));
        await assertValues(semanticMeterRows);
        await new Promise((resolve) => setTimeout(resolve, 350)); await settle();
        sharedStateVisibleText(target, "p", t("common.stale_note", {}, "era_1995"));
        await assertValues(semanticMeterRows);
        runtime.listener?.({ kind: "transport_recovered" }); await settle();
        expect(target.textContent).not.toContain(t("common.stale_note", {}, "era_1995"));
        await assertValues(semanticMeterRows);
        const changed = semanticMeterRows.map((row) => meterRow(row.meter_id, row.value + 8));
        const newer = parseGameUISnapshot({ ...semanticMeterSnapshot(changed), revision: 2 });
        runtime.current = newer; runtime.listener?.({ kind: "snapshot", value: newer }); await settle();
        expect(target.textContent).not.toContain(t("common.stale_note", {}, "era_1995"));
        await assertValues(changed);
      } finally { try { if (fixture) await fixture.dispose(); } finally { subscribe?.mockRestore(); await page.viewport(1280, 720); } }
    });
  }
}

for (const scenario of ["initial narrow mount", "live breakpoint changes", "narrow snapshot refresh"] as const) {
  it.skipIf(!browser)(`GS3-A3 semantic meters: ${scenario}`, async () => {
    const { page } = await import("vitest/browser");
    await page.viewport(scenario === "live breakpoint changes" ? 1280 : 320, 720);
    const runtime = new Runtime(); runtime.current = semanticMeterSnapshot();
    const { target, app, dispose } = await mounted(runtime);
    try {
      const nav = button(target, t("surface.meters.title", {}, "era_1995"));
      nav.focus(); nav.click(); await settle();
      expect(document.activeElement).toBe(nav);
      if (scenario === "live breakpoint changes") {
        for (const width of [1280, 320, 479, 480, 1280, 320]) {
          await page.viewport(width, 720); await settleMeterLayout();
          await assertSemanticMeters(target, semanticMeterRows, width < 480);
          expect(document.activeElement).toBe(nav);
        }
      } else {
        await assertSemanticMeters(target, semanticMeterRows, true);
        if (scenario === "narrow snapshot refresh") {
          const changed = semanticMeterRows.map((row) => meterRow(row.meter_id, row.value + 1));
          runtime.current = semanticMeterSnapshot(changed);
          app.fixtureSnapshot(runtime.current); await settle();
          await assertSemanticMeters(target, changed, true);
          await page.viewport(1280, 720); await settleMeterLayout();
          await assertSemanticMeters(target, changed, false);
        }
      }
      expect(document.activeElement).toBe(nav);
      expect(runtime.requests).toEqual([]);
    } finally { await dispose(); await page.viewport(1280, 720); }
  });
}

// GS0.5: these are valid wire snapshots, not malformed transport fixtures.
// Presentation errors must stay inside their read-only surface, not crash
// rendering or leak mechanical identifiers into player-visible content.
for (const surface of ["achievements", "meters"] as const) {
  for (const delivery of ["first-open", "mounted-refresh"] as const) {
    it.skipIf(!browser)(`GS0.5 contained presentation error ${surface} ${delivery}`, async () => {
      const titleKey: CopyKey = surface === "achievements" ? "surface.achievements.title" : "surface.meters.title";
      const unavailableID = surface === "achievements" ? "achievement.unregistered" : "critical";
      const brokenFeatures = surface === "achievements"
        ? { ...v4.features, achievements: { ...v4.features.achievements!, rows: v4.features.achievements!.rows.map((row, index) => index === 0 ? { ...row, copy_key: unavailableID } : row) } }
        : { ...v4.features, meters: { meters: v4.features.meters!.meters.map((row) => row.meter_id === "doom.probability"
          ? { ...row, value: 95, band_id: unavailableID, bands: [...row.bands, { band_id: unavailableID, floor_value: 90 }] } : row) } };
      const healthy = parseGameUISnapshot(structuredClone(v4));
      const broken = parseGameUISnapshot({ ...v4, revision: 2, features: brokenFeatures });
      if (!("features" in broken)) throw new Error("GS0.5 diagnostic must decode a v4 feature snapshot");
      expect(broken.features[surface]).not.toBeNull();
      const renderErrors: string[] = [];
      const observeError = (event: ErrorEvent) => { renderErrors.push(event.message); event.preventDefault(); };
      window.addEventListener("error", observeError);
      const diagnostic = vi.spyOn(console, "error").mockImplementation(() => {});
      const runtime = new Runtime(); runtime.current = healthy;
      const { target, app, dispose } = await mounted(runtime);
      const bindAndSettle = async (snapshot: ParsedGameUISnapshot) => {
        try { app.fixtureSnapshot(snapshot); await settle(); }
        catch (error) { renderErrors.push(String(error)); }
      };
      const open = async () => {
        try { button(target, t(titleKey, {}, "era_1995")).click(); await settle(); }
        catch (error) { renderErrors.push(String(error)); }
      };
      const assertHealthy = () => {
        const panel = target.querySelector(`.${surface}`)!;
        expect(panel).not.toBeNull();
        expect(panel.querySelectorAll(surface === "achievements" ? "li" : "meter")).toHaveLength(surface === "achievements" ? 3 : 11);
        expect(panel.querySelector("[role='alert']")).toBeNull();
      };
      const assertContained = (episodes: number) => {
        expect(renderErrors, "legal wire must not cause an uncontained render error").toEqual([]);
        const panel = target.querySelector(`.${surface}`)!;
        expect(panel).not.toBeNull();
        expect(panel.querySelectorAll("h1")).toHaveLength(1);
        expect(panel.querySelector("h1")?.textContent).toBe(t(titleKey, {}, "era_1995"));
        expect(panel.querySelector("h1")?.getAttribute("tabindex")).toBe("-1");
        expect(panel.querySelectorAll("[role='alert']")).toHaveLength(1);
        expect(panel.querySelector("[role='alert']")?.textContent).toBe(t("common.surface_error", {}, "era_1995"));
        expect(panel.querySelectorAll("button, input, li, table, meter")).toHaveLength(0);
        expect(target.textContent).not.toContain(unavailableID);
        expect(diagnostic).toHaveBeenCalledTimes(episodes);
        for (const call of diagnostic.mock.calls) expect(call).toEqual([`game UI invariant: ${surface} surface presentation unavailable`]);
        expect(runtime.requests).toHaveLength(0);
      };
      try {
        await open(); assertHealthy();
        expect(renderErrors).toEqual([]); expect(diagnostic).not.toHaveBeenCalled();
        if (delivery === "first-open") { button(target, t("surface.desk.title", {}, "era_1995")).click(); await settle(); }
        await bindAndSettle(broken);
        if (delivery === "first-open") await open();
        assertContained(1);
        await bindAndSettle(structuredClone(broken));
        app.fixtureMonotonicElapsed(3_000); await settle(); assertContained(1);
        await bindAndSettle({ ...healthy, revision: 3 }); assertHealthy();
        await bindAndSettle({ ...broken, revision: 4 }); assertContained(2);
        await assertAxe(target, `${surface} contained presentation error`);
        button(target, t("surface.desk.title", {}, "era_1995")).click(); await settle();
        expect(target.querySelector("main")?.dataset.surface).toBe("desk");
        expect(button(target, t("manual.click.title", {}, "era_1995"))).toBeDefined();
        button(target, t("surface.settings.title", {}, "era_1995")).click(); await settle();
        expect(target.querySelector("main")?.dataset.surface).toBe("settings");
        expect(target.querySelector("#settings-heading")?.textContent).toBe(t("surface.settings.title", {}, "era_1995"));
        expect(renderErrors).toEqual([]); expect(runtime.requests).toHaveLength(0);
      } finally {
        await dispose(); diagnostic.mockRestore(); window.removeEventListener("error", observeError);
      }
    });
  }
}

// GS1-A3 exact component-fixture time: no live host clock or server pacing
// claim. Expected display outputs do not call the phase implementation.
const fiscalEdgeFixtures = [
  { elapsed: 99_999, phase: "ripening", disabled: true, autoRemaining: "0:03:21",
    text: t("fiscal.period.ripening_frame", { remaining: "0:00:01" }, "era_1995") },
  { elapsed: 100_000, phase: "early", disabled: false, autoRemaining: "0:03:20",
    text: t("fiscal.period.early_frame", { success_percent: 50 }, "era_1995") },
  { elapsed: 199_999, phase: "early", disabled: false, autoRemaining: "0:01:41",
    text: t("fiscal.period.early_frame", { success_percent: 50 }, "era_1995") },
  { elapsed: 200_000, phase: "guaranteed", disabled: false, autoRemaining: "0:01:40",
    text: t("fiscal.period.guaranteed", {}, "era_1995") },
] as const;
for (const fixture of fiscalEdgeFixtures) {
  it.skipIf(!browser)(`GS1-A3 renders exact Fiscal edge ${fixture.elapsed} as visible ${fixture.phase} with native readiness`, async () => {
    const snapshot = parseGameUISnapshot({ ...v4, features: { ...v4.features,
      fiscal: { ...v4.features.fiscal!, period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - fixture.elapsed } },
    } });
    if (snapshot.schema_version !== 4 || !("features" in snapshot)) throw new Error("Fiscal edge fixture did not decode as v4");
    const arm = snapshot.features.fiscal!;
    expect(arm.period).toMatchObject({ early_ms: 100_000, guaranteed_ms: 200_000, auto_ms: 300_000, early_success_ppm: 500_000 });
    expect(snapshot.server_now_ms - arm.period.opened_wall_ms).toBe(fixture.elapsed);
    const target = document.createElement("div"); document.body.append(target);
    const onHarvest = vi.fn(), onSpendLevel = vi.fn(), onSpendUnlock = vi.fn();
    const app = mount(FiscalSurface, { target, props: {
      arm, era: "era_1995", serverNowMs: snapshot.server_now_ms, pending: false, controlsEnabled: true,
      onHarvest, onSpendLevel, onSpendUnlock,
    } });
    try {
      await settle();
      const surface = target.querySelector<HTMLElement>(".fiscal")!;
      expect(surface.dataset.phase).toBe(fixture.phase);
      expect(surface.querySelectorAll(".phase")).toHaveLength(1);
      const phase = surface.querySelector<HTMLElement>(".phase")!;
      expect(phase.textContent).toBe(fixture.text);
      const style = getComputedStyle(phase), rectangle = phase.getBoundingClientRect();
      expect(style.display).not.toBe("none"); expect(style.visibility).toBe("visible");
      expect(Number(style.opacity)).toBeGreaterThan(0);
      expect(rectangle.width).toBeGreaterThan(0); expect(rectangle.height).toBeGreaterThan(0);
      expect(phase.closest('[role="status"], [role="alert"], [aria-live]')).toBeNull();
      expect(phase.getAttribute("aria-hidden")).not.toBe("true");
      const region = surface.querySelector('[data-fiscal-region="harvest"]')!;
      const paragraphs = [...region.querySelectorAll("p")].map((node) => node.textContent);
      expect(paragraphs).toEqual([fixture.text, t("fiscal.period.auto_note", { remaining: fixture.autoRemaining }, "era_1995")]);
      const harvest = surface.querySelector<HTMLButtonElement>('[data-fiscal-action="harvest"]')!;
      expect(harvest.textContent).toBe(t("fiscal.harvest", {}, "era_1995"));
      expect(harvest.disabled).toBe(fixture.disabled);
      expect(harvest.getAttribute("aria-disabled")).toBeNull();
      expect(harvest.getAttribute("aria-describedby")).toBe("fiscal-harvest-curtain");
      expect(surface.querySelector("#fiscal-harvest-curtain")!.textContent).toBe(t("fiscal.harvest_tooltip", {}, "era_1995"));
      expect(onHarvest).not.toHaveBeenCalled();
      harvest.click(); await settle();
      expect(onHarvest).toHaveBeenCalledTimes(fixture.disabled ? 0 : 1);
      expect(onSpendLevel).not.toHaveBeenCalled(); expect(onSpendUnlock).not.toHaveBeenCalled();
    } finally { await unmount(app); target.remove(); }
  });
}

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

it.skipIf(!browser)("keeps Fiscal pending until the applied harvest refresh supplies the next Founder revision (GS0.2/GS1-A6)", async () => {
  class DelayedFiscalRuntime extends Runtime {
    snapshotCalls = 0;
    releaseRefresh: () => void = () => {};
    override async intent(body: Readonly<Record<string, unknown>>): Promise<IntentOutcome> {
      this.requests.push(body);
      if (body.kind === "harvest_fiscal_period") {
        this.current = { ...this.current, founder_revision: 8 };
        return { outcome: "applied", receipt: { harvest_outcome: "guaranteed", founder_revision: 8 } };
      }
      return { outcome: "applied", receipt: { founder_revision: 9 } };
    }
    override async snapshot(): Promise<ParsedGameUISnapshot> {
      if (this.requests.length === 0) return this.current;
      this.snapshotCalls += 1;
      if (this.snapshotCalls === 1) await new Promise<void>((resolve) => { this.releaseRefresh = resolve; });
      return this.current;
    }
  }
  const runtime = new DelayedFiscalRuntime();
  const { target, app, dispose } = await mounted(runtime);
  try {
    button(target, "Earnings Calls").click();
    app.fixtureSnapshot({ ...v4, features: { ...v4.features, fiscal: { ...v4.features.fiscal!, period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - 250_000 } } } });
    await settle();
    button(target, "Hold the earnings call").click();
    await settle();
    expect(runtime.requests[0]).toMatchObject({ kind: "harvest_fiscal_period", expected_revision: 7 });
    expect(runtime.snapshotCalls).toBe(1);
    expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
    expect(button(target, "Unlock for 3").disabled).toBe(false);
    expect(button(target, "Unlock for 3").getAttribute("aria-disabled")).toBe("true");
    runtime.releaseRefresh();
    await settle();
    expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
    button(target, "Unlock for 3").click();
    await settle();
    expect(runtime.requests[1]).toMatchObject({ kind: "spend_fiscal_credit", expected_revision: 8, target: { kind: "unlock", unlock_id: "minigame.pitch" } });
  } finally { runtime.releaseRefresh(); await dispose(); }
});

// Fiscal supplement uses public snapshot fixtures and runtime doubles. It is
// consumer evidence, never a real quarter clock or persisted Founder proof.
const fiscalText = (key: CopyKey): string => t(key, {}, "era_1995");
const withRipeFiscal = (): GameUISnapshot => ({ ...v4,
  features: { ...v4.features, fiscal: { ...v4.features.fiscal!, period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - 250_000 } } } });
const fiscalControls = (target: HTMLElement) => [
  button(target, fiscalText("fiscal.harvest")),
  button(target, t("fiscal.level_buy", { cost: 1 }, "era_1995")),
  button(target, t("fiscal.unlock_buy", { cost: 3 }, "era_1995")),
];

for (const activation of ["{Enter}", " "]) {
  it.skipIf(!browser)(`Fiscal supplement keyboard ${activation === " " ? "Space" : "Enter"} traverses and activates harvest, level and unlock`, async () => {
    const runtime = new Runtime(); runtime.current = withRipeFiscal();
    runtime.outcome = { outcome: "applied", receipt: { harvest_outcome: "guaranteed" } };
    const { target, dispose } = await mounted(runtime);
    try {
      button(target, fiscalText("surface.fiscal.title")).click(); await settle();
      const [harvest, level, unlock] = fiscalControls(target);
      const { userEvent } = await import("vitest/browser"); harvest.focus();
      await userEvent.keyboard(activation); await settle();
      expect(document.activeElement).toBe(harvest);
      await userEvent.keyboard("{Tab}"); expect(document.activeElement).toBe(level);
      await userEvent.keyboard(activation); await settle();
      expect(document.activeElement).toBe(level);
      await userEvent.keyboard("{Tab}"); expect(document.activeElement).toBe(unlock);
      await userEvent.keyboard(activation); await settle();
      expect(document.activeElement).toBe(unlock);
      expect(runtime.requests).toEqual([
        expect.objectContaining({ kind: "harvest_fiscal_period", expected_revision: 7 }),
        expect.objectContaining({ kind: "spend_fiscal_credit", expected_revision: 7, target: { kind: "generator_level", generator_id: "generator.beige_tower", levels: 1 } }),
        expect.objectContaining({ kind: "spend_fiscal_credit", expected_revision: 7, target: { kind: "unlock", unlock_id: "minigame.pitch" } }),
      ]);
    } finally { await dispose(); }
  });
}

for (const [label, index] of [["harvest", 0], ["level", 1], ["unlock", 2]] as const) {
  it.skipIf(!browser)(`Fiscal supplement ${label} stays focusable and guarded through held intent AND held read`, async () => {
    const runtime = new Runtime(); runtime.current = withRipeFiscal();
    const { target, dispose } = await mounted(runtime);
    let finish!: (value: IntentOutcome) => void, release!: (value: ParsedGameUISnapshot) => void;
    const heldIntent = new Promise<IntentOutcome>((resolve) => { finish = resolve; });
    const heldRead = new Promise<ParsedGameUISnapshot>((resolve) => { release = resolve; });
    const intent = vi.spyOn(runtime, "intent").mockImplementation((body) => { runtime.requests.push(body); return heldIntent; });
    const read = vi.spyOn(runtime, "snapshot").mockReturnValue(heldRead);
    const applied: IntentOutcome = { outcome: "applied", receipt: { harvest_outcome: "guaranteed" } };
    try {
      button(target, fiscalText("surface.fiscal.title")).click(); await settle();
      const control = fiscalControls(target)[index]; control.focus(); control.click(); await settle();
      const assertPending = () => {
        expect(control.disabled).toBe(false); expect(control.getAttribute("aria-disabled")).toBe("true");
        expect(document.activeElement).toBe(control);
        expect(target.querySelector(".fiscal")?.textContent).toContain(fiscalText("common.pending"));
        expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
      };
      assertPending(); control.click(); await settle(); expect(runtime.requests).toHaveLength(1);
      finish(applied); await settle(); expect(read).toHaveBeenCalledTimes(1); assertPending();
      control.click(); await settle(); expect(runtime.requests).toHaveLength(1);
      release({ ...withRipeFiscal(), founder_revision: 8 }); await settle();
      expect(control.getAttribute("aria-disabled")).not.toBe("true"); expect(document.activeElement).toBe(control);
      expect(target.querySelector(".fiscal")?.textContent).not.toContain(fiscalText("common.pending"));
      control.click(); await settle(); expect(runtime.requests).toHaveLength(2);
      expect(runtime.requests[1].expected_revision).toBe(8);
    } finally { finish(applied); release(runtime.current); await settle(); intent.mockRestore(); read.mockRestore(); await dispose(); }
  });
}

for (const [category, detail, key, index, invariant] of [
  ["not_eligible", "period_not_ripe", "fiscal.rejection.period_not_ripe", 0, false],
  ["unaffordable", "fiscal_credit", "fiscal.rejection.unaffordable", 1, false],
  ["not_eligible", "already_unlocked", "fiscal.rejection.already_unlocked", 2, false],
  ["cap_exceeded", "generator.beige_tower", "cap.fiscal_level.beige_tower", 1, false],
  ["unknown_id", "generator.not_in_catalog", "intent.rejection.unknown", 1, true],
  ["unknown_id", "unlock.not_in_catalog", "intent.rejection.unknown", 2, true],
] as const) {
  it.skipIf(!browser)(`Fiscal supplement refusal ${category}/${detail} renders its exact reason and invariant policy`, async () => {
    const runtime = new Runtime(); runtime.current = withRipeFiscal();
    runtime.outcome = { outcome: "rejected", category, detail, currentRevision: 7, sessionExpired: false };
    const { target, dispose } = await mounted(runtime);
    const diagnostics = vi.spyOn(console, "error").mockImplementation(() => {});
    try {
      button(target, fiscalText("surface.fiscal.title")).click(); await settle();
      const control = fiscalControls(target)[index]; control.focus(); control.click(); await settle();
      expect(runtime.requests).toHaveLength(1); expect(runtime.requests[0].expected_revision).toBe(7);
      expect(target.querySelector(".intent-notice")?.textContent).toBe(fiscalText(key));
      expect(target.querySelector(".intent-notice")?.textContent).not.toContain(`${category}/${detail}`);
      expect(diagnostics.mock.calls).toEqual(invariant ? [["game UI invariant: intent rejection"]] : []);
      expect(control.disabled).toBe(false); expect(document.activeElement).toBe(control);
      await assertAxe(target, `Fiscal refusal ${detail}`);
    } finally { diagnostics.mockRestore(); await dispose(); }
  });
}

it.skipIf(!browser)("Fiscal supplement cap notice follows the parsed snapshot reason instead of a fixed default key", async () => {
  const runtime = new Runtime(), value = withRipeFiscal(), fiscal = value.features.fiscal!;
  runtime.current = parseGameUISnapshot({ ...value, features: { ...value.features, fiscal: { ...fiscal,
    generator_levels: fiscal.generator_levels.map((row) => ({ ...row, level_cap: { ...row.level_cap, reason_key: "cap.fiscal_credit" } })),
  } } });
  runtime.outcome = { outcome: "rejected", category: "cap_exceeded", detail: "generator.beige_tower", currentRevision: 7, sessionExpired: false };
  const { target, dispose } = await mounted(runtime);
  try {
    button(target, fiscalText("surface.fiscal.title")).click(); await settle();
    fiscalControls(target)[1].click(); await settle();
    expect(runtime.requests).toHaveLength(1);
    expect(target.querySelector(".intent-notice")?.textContent).toBe(fiscalText("cap.fiscal_credit"));
    expect(target.querySelector(".intent-notice")?.textContent).not.toBe(fiscalText("cap.fiscal_level.beige_tower"));
  } finally { await dispose(); }
});

for (const outcome of ["early_succeeded", "early_failed", "guaranteed", "consumed_by_auto"] as const) {
  it.skipIf(!browser)(`Fiscal supplement renders the applied ${outcome} harvest outcome without predicting it`, async () => {
    const runtime = new Runtime(); runtime.current = withRipeFiscal();
    runtime.outcome = { outcome: "applied", receipt: { harvest_outcome: outcome } };
    const { target, dispose } = await mounted(runtime);
    try {
      button(target, fiscalText("surface.fiscal.title")).click(); await settle();
      fiscalControls(target)[0].click(); await settle();
      expect(runtime.requests).toHaveLength(1);
      expect(target.querySelector(".intent-notice")?.textContent).toBe(fiscalText(`fiscal.outcome.${outcome}`));
    } finally { await dispose(); }
  });
}

for (const [label, message] of [
  ["recovering", { kind: "transport_recovering" }],
  ["resync", { kind: "system", value: { kind: "resync_required" } }],
  ["restart", { kind: "system", value: { kind: "server_restarting", resume_after_ms: 1_000 } }],
] as const) {
  it.skipIf(!browser)(`Fiscal supplement marks ${label} values stale and refuses intents until recovery`, async () => {
    const runtime = new Runtime(); runtime.current = withRipeFiscal();
    const { target, dispose } = await mounted(runtime);
    try {
      button(target, fiscalText("surface.fiscal.title")).click(); await settle();
      runtime.listener?.(message); await settle();
      expect(fiscalControls(target).every((control) => control.disabled)).toBe(true);
      expect(target.querySelector(".fiscal")?.textContent).toContain(fiscalText("common.stale_note"));
      for (const control of fiscalControls(target)) control.click(); await settle(); expect(runtime.requests).toEqual([]);
      runtime.listener?.({ kind: "transport_recovered" }); await settle();
      expect(fiscalControls(target).every((control) => !control.disabled)).toBe(true); expect(runtime.requests).toEqual([]);
    } finally { await dispose(); }
  });
}

it.skipIf(!browser)("Fiscal supplement disables all Founder commands when its revision is unavailable", async () => {
  const runtime = new Runtime(); const value = withRipeFiscal();
  delete (value as unknown as Record<string, unknown>).founder_revision; runtime.current = value;
  const { target, dispose } = await mounted(runtime);
  try {
    button(target, fiscalText("surface.fiscal.title")).click(); await settle();
    expect(fiscalControls(target).every((control) => control.disabled)).toBe(true);
    for (const control of fiscalControls(target)) control.click(); await settle(); expect(runtime.requests).toEqual([]);
  } finally { await dispose(); }
});

for (const [label, index, navigate] of [["owned unlock", 2, false], ["capped level", 1, false], ["selected nav", 2, true]] as const) {
  it.skipIf(!browser)(`Fiscal supplement ${label} refresh preserves the accepted focus boundary`, async () => {
    const runtime = new Runtime(); runtime.current = withRipeFiscal();
    const { target, dispose } = await mounted(runtime);
    let finish!: (value: IntentOutcome) => void;
    const held = new Promise<IntentOutcome>((resolve) => { finish = resolve; });
    const intent = vi.spyOn(runtime, "intent").mockImplementation((body) => { runtime.requests.push(body); return held; });
    try {
      button(target, fiscalText("surface.fiscal.title")).click(); await settle();
      const control = fiscalControls(target)[index]; control.focus(); control.click(); await settle();
      const nav = button(target, fiscalText("surface.meters.title")); if (navigate) nav.focus();
      const next = withRipeFiscal(), fiscal = next.features.fiscal!;
      runtime.current = { ...next, founder_revision: 8, features: { ...next.features, fiscal: {
        ...fiscal,
        generator_levels: fiscal.generator_levels.map((row) => index === 1 ? { ...row, level: row.level_cap.amount, next_level_cost: null } : row),
        unlocks: fiscal.unlocks.map((row) => index === 2 && row.unlock_id === "minigame.pitch" ? { ...row, owned: true } : row),
      } } };
      finish({ outcome: "applied", receipt: {} }); await settle();
      expect(control.isConnected).toBe(false); expect(runtime.requests).toHaveLength(1);
      expect(document.activeElement).toBe(navigate ? nav : target.querySelector("#fiscal-heading"));
      expect(target.querySelector(".fiscal")?.textContent).toContain(fiscalText(index === 1 ? "cap.fiscal_level.beige_tower" : "fiscal.unlock_owned"));
    } finally { finish({ outcome: "applied", receipt: {} }); await settle(); intent.mockRestore(); await dispose(); }
  });
}

it.skipIf(!browser)("Fiscal supplement prefers the nearest of multiple surviving controls in the same region", async () => {
  const runtime = new Runtime(), value = withRipeFiscal(), fiscal = value.features.fiscal!;
  // Parsed public fixture with existing presentation rows, byte-sorted IDs;
  // not a claim that the production catalog sells these additional levels.
  const levels = [
    { ...fiscal.generator_levels[0], generator_id: "generator.answering_machine", next_level_cost: 2 },
    { ...fiscal.generator_levels[0], next_level_cost: 3 },
    { ...fiscal.generator_levels[0], generator_id: "generator.first_hire", next_level_cost: 1 },
  ];
  runtime.current = parseGameUISnapshot({ ...value, features: { ...value.features, fiscal: { ...fiscal, generator_levels: levels } } });
  const { target, dispose } = await mounted(runtime);
  const intent = vi.spyOn(runtime, "intent").mockImplementation(async (body) => {
    runtime.requests.push(body);
    runtime.current = parseGameUISnapshot({ ...value, founder_revision: 8, features: { ...value.features, fiscal: { ...fiscal,
      generator_levels: levels.map((row) => row.generator_id === "generator.first_hire" ? { ...row, level: row.level_cap.amount, next_level_cost: null } : row),
    } } });
    return { outcome: "applied", receipt: {} };
  });
  try {
    button(target, fiscalText("surface.fiscal.title")).click(); await settle();
    const level = button(target, t("fiscal.level_buy", { cost: 1 }, "era_1995"));
    const surviving = button(target, t("fiscal.level_buy", { cost: 3 }, "era_1995"));
    level.focus(); level.click(); await settle();
    expect(level.isConnected).toBe(false); expect(document.activeElement).toBe(surviving);
    expect(runtime.requests).toHaveLength(1);
    expect(runtime.requests[0]).toMatchObject({ target: { kind: "generator_level", generator_id: "generator.first_hire", levels: 1 } });
  } finally { intent.mockRestore(); await dispose(); }
});

it.skipIf(!browser)("Fiscal supplement component pending guards do not depend on the host single-flight queue", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const onHarvest = vi.fn(), onSpendLevel = vi.fn(), onSpendUnlock = vi.fn();
  const app = mount(FiscalSurface, { target, props: { arm: withRipeFiscal().features.fiscal!, era: "era_1995", serverNowMs: NOW, pending: true, controlsEnabled: true, onHarvest, onSpendLevel, onSpendUnlock } });
  try {
    await settle(); const { userEvent } = await import("vitest/browser");
    for (const control of fiscalControls(target)) {
      expect(control.disabled).toBe(false); expect(control.getAttribute("aria-disabled")).toBe("true");
      control.focus(); control.click(); await userEvent.keyboard("{Enter}"); await userEvent.keyboard(" "); await settle();
      expect(document.activeElement).toBe(control);
    }
    expect(onHarvest).not.toHaveBeenCalled(); expect(onSpendLevel).not.toHaveBeenCalled(); expect(onSpendUnlock).not.toHaveBeenCalled();
  } finally { await unmount(app); target.remove(); }
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

// GS3 event presentation is not value authority. Decode the real public
// envelope before runtime-double delivery; this is not a real socket test.
function meterBandMessage(cursor: number, meterID: string, before: number, after: number): GameUIRuntimeMessage {
  const envelope = decodeTransportEnvelope({ v: 2, ch: `player:${v4.run.founder_id}`, kind: "event", rev: cursor,
    constants_hash: v4.constants_hash, ts: "2026-10-07T07:00:00Z",
    payload: { event_id: `meter-${cursor}`, kind: "meter_band_changed.v1", scope: "company", rev: cursor, cursor_effect: "advance",
      payload: { direction: after > before ? "up" : "down", from_band: before >= 70 ? "high" : "low", meter_id: meterID,
        run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 },
        to_band: after >= 70 ? "high" : "low", value_after: after, value_before: before } } });
  if (!envelope) throw new Error("meter fixture envelope not admitted");
  const value = decodeGameUIAnnouncement(envelope);
  if (!value || value.kind !== "meter_band_changed") throw new Error("meter fixture announcement not admitted");
  return { kind: "announcement", scope: "company", value };
}

for (const meterID of ["doom.probability", "trust.users.standing"] as const) {
  for (const direction of ["up", "down"] as const) {
    for (const width of [320, 1280] as const) {
      it.skipIf(!browser)(`GS3 event authority ${meterID}/${direction}/${width}`, async () => {
        const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
        const before = direction === "up" ? 69 : 71, after = direction === "up" ? 71 : 69;
        const initial = semanticMeterRows.map((row) => row.meter_id === meterID ? meterRow(meterID, before) : row);
        const updated = initial.map((row) => row.meter_id === meterID ? meterRow(meterID, after) : row);
        const runtime = new Runtime(); runtime.current = semanticMeterSnapshot(initial);
        const diagnostic = vi.spyOn(console, "error").mockImplementation(() => {});
        let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
        try {
          fixture = await mounted(runtime); const { target } = fixture;
          const navTitle = t("surface.meters.title", {}, "era_1995"), badge = t("meters.nav_changed_badge", {}, "era_1995");
          const nav = button(target, navTitle), desk = button(target, t("surface.desk.title", {}, "era_1995"));
          const region = target.querySelector<HTMLElement>(".announcement")!;
          expect(region.getAttribute("role")).toBe("status"); expect(region.textContent).toBe("");
          const select = async (control: HTMLButtonElement) => { control.focus(); await userEvent.keyboard("{Enter}"); await settleMeterLayout(); expect(document.activeElement).toBe(control); };
          const values = async (rows: typeof initial) => {
            await assertSemanticMeters(target, rows, width < 480);
            expect(target.querySelector(".meters")!.querySelectorAll("button,input,select,textarea")).toHaveLength(0);
            expect(runtime.requests).toEqual([]); expect(document.activeElement).toBe(nav);
          };
          const meterPresentation = FEATURES_PRESENTATION.trustMeters.get(meterID);
          const label = meterPresentation
            ? t("meters.row_frame", { constituency: t(meterPresentation.constituency_key, {}, "era_1995"), axis: t(meterPresentation.axis_key, {}, "era_1995") }, "era_1995")
            : t(FEATURES_PRESENTATION.doomMeter.title_key, {}, "era_1995");
          const text = (value: number) => t("meters.band_changed_announcement", { meter: label,
            band: t(FEATURES_PRESENTATION.meterBands.get(value >= 70 ? "high" : "low")!, {}, "era_1995") }, "era_1995");
          desk.focus();
          runtime.listener?.(meterBandMessage(10, meterID, before, after)); await settle();
          expect(nav.textContent).toBe(`${navTitle}${badge}`); expect(region.textContent).toBe("");
          expect(document.activeElement).toBe(desk); expect(runtime.requests).toEqual([]);
          await select(nav); expect(nav.textContent).toBe(navTitle); expect(region.textContent).toBe("");
          await values(initial);
          runtime.listener?.(meterBandMessage(11, meterID, before, after)); await settle();
          sharedStateVisibleText(target, '.announcement[role="status"]', text(after));
          expect(nav.textContent).toBe(navTitle); await values(initial);
          const fresh = parseGameUISnapshot({ ...semanticMeterSnapshot(updated), revision: 2 });
          runtime.current = fresh; runtime.listener?.({ kind: "snapshot", value: fresh }); await settle();
          await values(updated); expect(region.textContent).toBe(text(after));
          await select(desk);
          runtime.listener?.(meterBandMessage(12, meterID, after, before)); await settle();
          expect(nav.textContent).toBe(`${navTitle}${badge}`); expect(region.textContent).toBe(text(after));
          expect(document.activeElement).toBe(desk); expect(runtime.requests).toEqual([]);
          await select(nav); expect(nav.textContent).toBe(navTitle); expect(region.textContent).toBe(text(after));
          await values(updated);
          runtime.listener?.(meterBandMessage(13, meterID, after, before)); await settle();
          sharedStateVisibleText(target, '.announcement[role="status"]', text(before)); await values(updated);
          for (const cursor of [10, 11]) {
            runtime.listener?.(meterBandMessage(cursor, meterID, before, after)); await settle();
            expect(region.textContent).toBe(text(before)); expect(nav.textContent).toBe(navTitle); await values(updated);
          }
          const newest = parseGameUISnapshot({ ...semanticMeterSnapshot(initial), revision: 3 });
          runtime.current = newest; runtime.listener?.({ kind: "snapshot", value: newest }); await settle();
          await values(initial); expect(region.textContent).toBe(text(before));
          expect(diagnostic).not.toHaveBeenCalled();
          runtime.listener?.(meterBandMessage(14, "trust.unregistered.standing", before, after)); await settle();
          expect(diagnostic).toHaveBeenCalledExactlyOnceWith("game UI invariant: unannounceable meter change trust.unregistered.standing");
          expect(region.textContent).toBe(text(before)); expect(nav.textContent).toBe(navTitle); await values(initial);
        } finally {
          try { if (fixture) await fixture.dispose(); }
          finally { diagnostic.mockRestore(); await page.viewport(1280, 720); }
        }
      });
    }
  }
}

// GS0.4/0.6: decoded lifecycle publication and actual forced native focus.
// Runtime-double delivery is not server issuance or a physical AT session.
function forcedLifecycleMessage(destination: "offer_sheet" | "run_end", cursor: number): Extract<GameUIRuntimeMessage, { kind: "event" }> {
  const kind = destination === "offer_sheet" ? "exit_offer_spawned" : "run_ended";
  const payout = { clout_reach_note: "clout.reach.preserved", network_slot_unlocks: [], reputation_delta: 2, route_knowledge: 25 };
  const payload = destination === "offer_sheet"
    ? { exit_type: "scripted_first", expires_at_ms: NOW + 60_000, offer_id: "01985555-3333-7333-8333-333333333333", payout_preview: payout }
    : { assisted: { advisor: false, commons: false }, attended_ms: 500, ended_at_ms: NOW + 1_000,
      executed_routes: [], exit_type: "scripted_first", faction: null, founder_id: v4.run.founder_id,
      gates_crossed: ["gate.t0_to_t1"], generators_purchased_total: 1, ledger_fact_kinds: [], lifetime_value: "1e3",
      payout, pre_timer: false, rta_ms: 2_000,
      run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 },
      started_at_ms: v4.run.run_started_at_ms, terminal_seq: 2, tier: 0 };
  const envelope = decodeTransportEnvelope({ v: 2, ch: `player:${v4.run.founder_id}`, kind: "event", rev: cursor,
    constants_hash: v4.constants_hash, ts: new Date(NOW + 1_000).toISOString(),
    payload: { event_id: `lifecycle-${cursor}`, kind, scope: "company", rev: cursor, cursor_effect: "advance", payload } });
  if (!envelope) throw new Error("lifecycle fixture envelope not admitted");
  const value = decodeGameUIEvent(envelope);
  if (!value || value.kind !== kind) throw new Error("lifecycle fixture event not admitted");
  return { kind: "event", revision: cursor, scope: "company", value };
}

async function assertForcedLifecycle(target: HTMLElement, runtime: Runtime, destination: "offer_sheet" | "run_end", width: number) {
  const id = destination === "offer_sheet" ? "offer-heading" : "run-end-heading";
  const key = destination === "offer_sheet" ? "screen.offer_sheet.heading" : "curriculum.scripted_first_failure.title";
  expect(target.querySelector("main")?.dataset.surface).toBe(destination);
  const heading = sharedStateVisibleText(target, "h1", t(key, {}, "era_1995"));
  expect(heading.id).toBe(id);
  expect({ tag: document.activeElement?.tagName, id: document.activeElement?.id }, "forced lifecycle exposes new context").toEqual({ tag: "H1", id });
  expect(document.activeElement).toBe(heading); expect(heading.getAttribute("tabindex")).toBe("-1");
  expect(runtime.requests).toEqual([]);
  expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(width + 1);
  await assertAxe(target, `forced ${destination}/${width}`);
}

for (const source of ["achievements", "fiscal", "meters"] as const) {
  for (const origin of ["nav", "removed-heading"] as const) {
    for (const destination of ["offer_sheet", "run_end"] as const) {
      for (const width of [320, 1280] as const) {
        it.skipIf(!browser)(`GS0 lifecycle focus ${source}/${origin}/${destination}/${width}`, async () => {
          const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
          const runtime = new Runtime(); runtime.current = parseGameUISnapshot(structuredClone(v4));
          const { target, dispose } = await mounted(runtime);
          try {
            const nav = button(target, t(`surface.${source}.title`, {}, "era_1995"));
            nav.focus(); await userEvent.keyboard("{Enter}"); await settle();
            expect(target.querySelector("main")?.dataset.surface).toBe(source); expect(document.activeElement).toBe(nav);
            const priorHeading = target.querySelector<HTMLElement>("h1")!;
            if (origin === "removed-heading") { priorHeading.focus(); expect(document.activeElement).toBe(priorHeading); }
            expect(runtime.requests).toEqual([]); await assertAxe(target, `lifecycle source ${source}/${width}`);
            runtime.listener?.(forcedLifecycleMessage(destination, 10)); await settle();
            expect(priorHeading.isConnected).toBe(false);
            await assertForcedLifecycle(target, runtime, destination, width);
          } finally { try { await dispose(); } finally { await page.viewport(1280, 720); } }
        });
      }
    }
  }
}

for (const width of [320, 1280] as const) {
  for (const destination of ["offer_sheet", "run_end"] as const) {
    it.skipIf(!browser)(`GS0 lifecycle focus newer Settings cancels ${destination}/${width}`, async () => {
      const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
      const runtime = new Runtime(); runtime.current = parseGameUISnapshot(structuredClone(v4));
      const { target, dispose } = await mounted(runtime);
      try {
        const nav = button(target, t("surface.meters.title", {}, "era_1995")); nav.focus();
        await userEvent.keyboard("{Enter}"); await settle(); expect(document.activeElement).toBe(nav);
        const settings = button(target, t("surface.settings.title", {}, "era_1995"));
        runtime.listener?.(forcedLifecycleMessage(destination, 10));
        // Controlled same-callback newer choice, before rendering/queued focus.
        settings.focus(); settings.click(); await settle();
        expect(target.querySelector("main")?.dataset.surface).toBe("settings"); expect(document.activeElement).toBe(settings);
        runtime.listener?.(forcedLifecycleMessage(destination, 10)); await settle();
        expect(target.querySelector("main")?.dataset.surface).toBe("settings"); expect(document.activeElement).toBe(settings);
        expect(runtime.requests).toEqual([]); await assertAxe(target, `lifecycle cancellation ${destination}/${width}`);
      } finally { try { await dispose(); } finally { await page.viewport(1280, 720); } }
    });
  }
  it.skipIf(!browser)(`GS0 lifecycle focus newest Run-End supersedes Offer then ignores replay ${width}`, async () => {
    const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
    const runtime = new Runtime(); runtime.current = parseGameUISnapshot(structuredClone(v4));
    const { target, dispose } = await mounted(runtime);
    try {
      const nav = button(target, t("surface.meters.title", {}, "era_1995")); nav.focus();
      await userEvent.keyboard("{Enter}"); await settle();
      runtime.listener?.(forcedLifecycleMessage("offer_sheet", 10));
      runtime.listener?.(forcedLifecycleMessage("run_end", 11)); await settle();
      await assertForcedLifecycle(target, runtime, "run_end", width);
      const settings = button(target, t("surface.settings.title", {}, "era_1995")); settings.focus();
      await userEvent.keyboard("{Enter}"); await settle(); expect(document.activeElement).toBe(settings);
      runtime.listener?.(forcedLifecycleMessage("offer_sheet", 10));
      runtime.listener?.(forcedLifecycleMessage("run_end", 11)); await settle();
      expect(target.querySelector("main")?.dataset.surface).toBe("settings"); expect(document.activeElement).toBe(settings);
      expect(runtime.requests).toEqual([]); await assertAxe(target, `ordered lifecycle replay ${width}`);
    } finally { try { await dispose(); } finally { await page.viewport(1280, 720); } }
  });
}

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
  const { dispose } = await mounted(runtime);
  try {
    // Native timers actually run for the specified minute. The display-only
    // monotonic fixture cannot stand in for an elapsed setTimeout/interval.
    const started = performance.now();
    while (performance.now() - started < 60_000) {
      await new Promise((resolve) => setTimeout(resolve, 250));
      expect(runtime.requests).toEqual([]);
    }
    expect(performance.now() - started).toBeGreaterThanOrEqual(60_000);
  } finally { await dispose(); }
}, 70_000);

for (const activation of ["{Enter}", " "]) {
it.skipIf(!browser)(`claims by native Tab and ${activation === " " ? "Space" : "Enter"}, with saturated Lucky cap text (GS5-A3/A5)`, async () => {
  const runtime = new Runtime(); runtime.current = withOpportunity(opportunityArm(true, true, true));
  runtime.outcome = { outcome: "applied", receipt: { outcome: "applied", receipt: { opportunity: { opportunity_id: "01986666-0000-7000-8000-000000000001", effect_row_id: "active.lucky",
    selected_target: null, buff_instance_id: null, requested_delta: "5e3", actual_credited_delta: "1e3", saturated: true, cap_reason_key: "cap.cash", next_sampled_interval_ms: 1_000, next_opportunity_attended_ms: 9_000 } } } };
  const { target, dispose } = await mounted(runtime);
  try {
    const region = target.querySelector("[data-region='desk.region.opportunity']")!;
    expect(region.textContent).toContain("Your boosts are stacked up to the combo cap.");
    expect(region.textContent).toContain("Boost combo cap");
    const claim = button(target, "Claim");
    const { userEvent } = await import("vitest/browser");
    const manual = target.querySelector<HTMLButtonElement>("section.manual button")!;
    manual.focus();
    await userEvent.keyboard("{Tab}"); await settle();
    expect(document.activeElement).toBe(claim);
    await userEvent.keyboard(activation); await settle();
    expect(runtime.requests).toEqual([expect.objectContaining({ kind: "claim_opportunity", opportunity_id: "01986666-0000-7000-8000-000000000001", expected_revision: 1 })]);
    expect(region.textContent).toContain("The lucky break hit the cash cap. Cash cap");
    expect(region.textContent).toContain("Lucky break credited: 1e3");
    await assertAxe(target, "claimed");
  } finally { await dispose(); }
});
}

it.skipIf(!browser)("renders a capped buff receipt even after the live buff arm is empty (GS5)", async () => {
  const runtime = new Runtime();
  const arm = opportunityArm(true, false, false);
  arm.pending!.effect_row_id = "active.production";
  runtime.current = withOpportunity(arm);
  runtime.outcome = { outcome: "applied", receipt: { outcome: "applied", receipt: { opportunity: {
    opportunity_id: arm.pending!.opportunity_id, effect_row_id: "active.production",
    selected_target: null, buff_instance_id: "01986666-0000-7000-8000-00000000000b",
    requested_delta: null, actual_credited_delta: null, saturated: null,
    cap_reason_key: "cap.active_combo", next_sampled_interval_ms: 1_000,
    next_opportunity_attended_ms: 9_000,
  } } } };
  const { target, dispose } = await mounted(runtime);
  try {
    runtime.current = withOpportunity(opportunityArm(false, false, false));
    button(target, "Claim").click(); await settle();
    const region = target.querySelector("[data-region='desk.region.opportunity']")!;
    expect(region.textContent).toContain("Boost combo cap");
    expect(region.textContent).not.toContain("Lucky break credited");
    await assertAxe(target, "capped buff receipt");
  } finally { await dispose(); }
});

it.skipIf(!browser)("renders the projected combo cap number beside its label for live buffs (GS5)", async () => {
  const runtime = new Runtime();
  runtime.current = withOpportunity(opportunityArm(false, true, false));
  const { target, dispose } = await mounted(runtime);
  try {
    const region = target.querySelector("[data-region='desk.region.opportunity']")!;
    expect(region.textContent).toContain("Boost combo cap");
    expect([...region.querySelectorAll("output")].map((node) => node.textContent)).toContain(formatAmount("1e4"));
    await assertAxe(target, "unsaturated live buff cap");
  } finally { await dispose(); }
});

for (const refusal of [
  { name: "unknown opportunity", category: "unknown_id", detail: "opportunity_id", notice: "desk.opportunity.rejection.not_pending", reports: 1 },
  { name: "unlisted rejection", category: "not_eligible", detail: "unlisted_reason", notice: "intent.rejection.unknown", reports: 1 },
  { name: "invalid request", category: "invalid", detail: "body", notice: "intent.rejection.unknown", reports: 1, error: true },
  { name: "expired opportunity", category: "not_eligible", detail: "opportunity_expired", notice: "desk.opportunity.rejection.expired", reports: 0 },
  { name: "not-pending opportunity", category: "not_eligible", detail: "opportunity_not_pending", notice: "desk.opportunity.rejection.not_pending", reports: 0 },
] as const) {
  it.skipIf(!browser)(`reports exactly the required invariant for ${refusal.name} (GS5/GS0.2)`, async () => {
    const runtime = new Runtime(); runtime.current = withOpportunity(opportunityArm(true));
    runtime.outcome = { outcome: "rejected", category: refusal.category, detail: refusal.detail, currentRevision: 1, sessionExpired: false };
    if ("error" in refusal) vi.spyOn(runtime, "intent").mockImplementation(async (body) => {
      runtime.requests.push(body); throw new GameUIRequestError(400, refusal.category, refusal.detail);
    });
    const { target, dispose } = await mounted(runtime);
    const diagnostics = vi.spyOn(console, "error").mockImplementation(() => {});
    try {
      button(target, "Claim").click(); await settle();
      expect(diagnostics).toHaveBeenCalledTimes(refusal.reports);
      for (const call of diagnostics.mock.calls) {
        expect(call).toEqual([expect.stringMatching(/^game UI invariant: /u)]);
      }
      const status = target.querySelector(".intent-notice")?.textContent;
      expect(status).toBe(t(refusal.notice, {}, "era_1995"));
      expect(status).not.toContain(`${refusal.category}/${refusal.detail}`);
      expect(runtime.requests).toHaveLength(1);
      expect(runtime.requests[0]).toMatchObject({ kind: "claim_opportunity", expected_revision: 1 });
      expect(button(target, "Claim").disabled).toBe(false);
    } finally { diagnostics.mockRestore(); await dispose(); }
  });
}

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

// Care supplement: native host/runtime-double evidence, not a server care or
// raw-stat/cooldown projection. All values below are public PA7 fields.
const careText = (key: CopyKey): string => t(key, {}, "era_1995");
for (const activation of ["{Enter}", " "]) {
  it.skipIf(!browser)(`care supplement reaches the next eligible action with Tab and ${activation === " " ? "Space" : "Enter"}`, async () => {
    const runtime = new Runtime(); runtime.current = withPet();
    const { target, dispose } = await mounted(runtime);
    try {
      button(target, careText("pet.care.panel.title")).click(); await settle();
      const feed = button(target, careText("pet.care.action.feed.title"));
      const pet = button(target, careText("pet.care.action.pet.title"));
      const { userEvent } = await import("vitest/browser");
      feed.focus();
      await userEvent.keyboard("{Tab}"); await settle();
      expect(document.activeElement).toBe(pet);
      await userEvent.keyboard(activation); await settle();
      expect(runtime.requests).toEqual([expect.objectContaining({ kind: "care_action", pet_id: petRow.pet_id, action_id: "care.pet", expected_revision: 7 })]);
      expect(document.activeElement).toBe(pet);
      expect(target.querySelector(".intent-notice")?.textContent).toBe(careText("pet.care.applied"));
    } finally { await dispose(); }
  });
}

for (const [category, detail, key, reports] of [
  ["not_eligible", "cooldown", "pet.care.rejection.cooldown", 0],
  ["not_eligible", "ineligible", "pet.care.rejection.ineligible", 0],
  ["not_eligible", "saturated", "pet.care.rejection.saturated", 0],
  ["not_eligible", "human_content_locked", "pet.care.rejection.soul_locked", 0],
  ["unknown_id", "unknown_pet", "intent.rejection.unknown", 1],
  ["unknown_id", "unknown_action", "intent.rejection.unknown", 1],
] as const) {
  it.skipIf(!browser)(`care supplement renders ${category}/${detail} without a retry or mechanical disclosure`, async () => {
    const runtime = new Runtime(); runtime.current = withPet();
    runtime.outcome = { outcome: "rejected", category, detail, currentRevision: 7, sessionExpired: false };
    const { target, dispose } = await mounted(runtime);
    const diagnostics = vi.spyOn(console, "error").mockImplementation(() => {});
    try {
      button(target, careText("pet.care.panel.title")).click(); await settle();
      const feed = button(target, careText("pet.care.action.feed.title")); feed.focus(); feed.click(); await settle();
      expect(runtime.requests).toHaveLength(1);
      expect(runtime.requests[0]).toMatchObject({ kind: "care_action", pet_id: petRow.pet_id, action_id: "care.feed", expected_revision: 7 });
      expect(target.querySelector(".intent-notice")?.textContent).toBe(careText(key));
      expect(target.querySelector(".intent-notice")?.textContent).not.toContain(`${category}/${detail}`);
      expect(diagnostics).toHaveBeenCalledTimes(reports);
      for (const call of diagnostics.mock.calls) expect(call).toEqual([expect.stringMatching(/^game UI invariant: /u)]);
      expect(feed.disabled).toBe(false);
      expect(document.activeElement).toBe(feed);
      await assertAxe(target, `care ${detail}`);
    } finally { diagnostics.mockRestore(); await dispose(); }
  });
}

for (const [label, message, reasonKey] of [
  ["recovering", { kind: "transport_recovering" }, "common.stale_note"],
  ["resync", { kind: "system", value: { kind: "resync_required" } }, "system.resync.title"],
  ["restart", { kind: "system", value: { kind: "server_restarting", resume_after_ms: 1_000 } }, "system.drain_notice.title"],
] as const) {
  it.skipIf(!browser)(`care supplement prevents commands during ${label} and recovers without an automatic care intent`, async () => {
    const runtime = new Runtime(); runtime.current = withPet();
    const { target, dispose } = await mounted(runtime);
    try {
      button(target, careText("pet.care.panel.title")).click(); await settle();
      runtime.listener?.(message); await settle();
      const actions = [...target.querySelectorAll<HTMLButtonElement>(".pet-care .actions button")];
      expect(actions).toHaveLength(5);
      expect(actions.every((action) => action.disabled)).toBe(true);
      expect(target.textContent).toContain(careText(reasonKey));
      for (const action of actions) action.click(); await settle();
      expect(runtime.requests).toEqual([]);
      runtime.listener?.({ kind: "transport_recovered" }); await settle();
      expect(button(target, careText("pet.care.action.feed.title")).disabled).toBe(false);
      expect(runtime.requests).toEqual([]);
    } finally { await dispose(); }
  });
}

it.skipIf(!browser)("care supplement disables Founder controls when its revision is unavailable", async () => {
  const runtime = new Runtime(); const withoutRevision = { ...withPet() };
  delete (withoutRevision as unknown as Record<string, unknown>).founder_revision;
  runtime.current = withoutRevision;
  const { target, dispose } = await mounted(runtime);
  try {
    button(target, careText("pet.care.panel.title")).click(); await settle();
    for (const action of target.querySelectorAll<HTMLButtonElement>(".pet-care .actions button")) {
      expect(action.disabled).toBe(true); action.click();
    }
    await settle(); expect(runtime.requests).toEqual([]);
  } finally { await dispose(); }
});

it.skipIf(!browser)("care supplement retains keyboard focus and pending text through the intent AND authoritative refresh", async () => {
  const runtime = new Runtime(); runtime.current = withPet();
  const { target, dispose } = await mounted(runtime);
  let finishIntent!: (value: IntentOutcome) => void;
  let finishRead!: (value: ParsedGameUISnapshot) => void;
  const heldIntent = new Promise<IntentOutcome>((resolve) => { finishIntent = resolve; });
  const heldRead = new Promise<ParsedGameUISnapshot>((resolve) => { finishRead = resolve; });
  const intent = vi.spyOn(runtime, "intent").mockImplementation((body) => { runtime.requests.push(body); return heldIntent; });
  const read = vi.spyOn(runtime, "snapshot").mockReturnValue(heldRead);
  try {
    button(target, careText("pet.care.panel.title")).click(); await settle();
    const feed = button(target, careText("pet.care.action.feed.title")); feed.focus();
    const { userEvent } = await import("vitest/browser");
    await userEvent.keyboard("{Enter}"); await settle();
    const assertPending = () => {
      expect(feed.disabled).toBe(false);
      expect(feed.getAttribute("aria-disabled")).toBe("true");
      expect(document.activeElement).toBe(feed);
      expect(target.querySelector(".pet-care")?.textContent).toContain(careText("common.pending"));
      expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
    };
    assertPending();
    feed.click(); button(target, careText("pet.care.action.pet.title")).click(); await settle();
    expect(runtime.requests).toHaveLength(1);
    finishIntent({ outcome: "applied", receipt: {} }); await settle();
    expect(read).toHaveBeenCalledTimes(1); assertPending();
    feed.click(); await settle(); expect(runtime.requests).toHaveLength(1);
    finishRead({ ...withPet(), founder_revision: 8 }); await settle();
    expect(feed.disabled).toBe(false);
    expect(feed.getAttribute("aria-disabled")).not.toBe("true");
    expect(document.activeElement).toBe(feed);
    expect(target.querySelector(".pet-care")?.textContent).not.toContain(careText("common.pending"));
  } finally {
    finishIntent({ outcome: "applied", receipt: {} }); finishRead(runtime.current); await settle();
    intent.mockRestore(); read.mockRestore(); await dispose();
  }
});

it.skipIf(!browser)("care supplement refreshes public band and eligibility before binding the next Founder intent", async () => {
  const runtime = new Runtime(); runtime.current = withPet();
  const { target, dispose } = await mounted(runtime);
  const read = vi.spyOn(runtime, "snapshot");
  const intent = vi.spyOn(runtime, "intent").mockImplementation(async (body) => {
    runtime.requests.push(body);
    const next = withPet();
    runtime.current = { ...next, founder_revision: 8,
      features: { ...next.features, pet_adoption: { ...next.features.pet_adoption!, pets: [{ ...petRow, status_band: "high", eligible_action_ids: ["care.groom"] }] } } };
    return { outcome: "applied", receipt: {} };
  });
  try {
    button(target, careText("pet.care.panel.title")).click(); await settle();
    const feed = button(target, careText("pet.care.action.feed.title")); feed.focus(); feed.click(); await settle();
    expect(read).toHaveBeenCalledTimes(1);
    const surface = target.querySelector(".pet-care")!;
    expect(surface.textContent).toContain(careText("pet.care.band.high"));
    expect(surface.textContent).not.toContain(careText("pet.care.band.normal"));
    expect(button(target, careText("pet.care.action.feed.title")).disabled).toBe(true);
    expect(button(target, careText("pet.care.action.groom.title")).disabled).toBe(false);
    expect(document.activeElement).toBe(surface.querySelector("h1"));
    button(target, careText("pet.care.action.groom.title")).click(); await settle();
    expect(runtime.requests).toEqual([
      expect.objectContaining({ kind: "care_action", action_id: "care.feed", expected_revision: 7 }),
      expect.objectContaining({ kind: "care_action", action_id: "care.groom", expected_revision: 8 }),
    ]);
  } finally { intent.mockRestore(); read.mockRestore(); await dispose(); }
});

for (const arm of ["conflict", "rate limit", "exclusive activity"] as const) {
  it.skipIf(!browser)(`care shared refusal ${arm} waits for fresh state and fresh consent without an automatic retry`, async () => {
    const runtime = new Runtime(); runtime.current = withPet();
    const { target, dispose } = await mounted(runtime);
    let release!: (value: ParsedGameUISnapshot) => void;
    const held = new Promise<ParsedGameUISnapshot>((resolve) => { release = resolve; });
    const read = vi.spyOn(runtime, "snapshot").mockReturnValue(held);
    const diagnostics = vi.spyOn(console, "error").mockImplementation(() => {});
    const intent = vi.spyOn(runtime, "intent").mockImplementation(async (body) => {
      runtime.requests.push(body);
      if (runtime.requests.length !== 1) return { outcome: "applied", receipt: {} };
      if (arm === "exclusive activity") return { outcome: "rejected", category: "not_eligible", detail: "exclusive_activity", currentRevision: 7, sessionExpired: false };
      throw arm === "conflict" ? new GameUIRequestError(409, "conflict", "intent") : new GameUIRequestError(429, "rate_limited", "account");
    });
    try {
      button(target, careText("pet.care.panel.title")).click(); await settle();
      const feed = button(target, careText("pet.care.action.feed.title")); feed.focus(); feed.click(); await settle();
      expect(runtime.requests).toHaveLength(1);
      expect(runtime.requests[0]).toMatchObject({ kind: "care_action", action_id: "care.feed", expected_revision: 7 });
      expect(target.querySelector(".intent-notice")?.textContent).toBe(careText(arm === "conflict" ? "intent.conflict" : arm === "rate limit" ? "intent.rate_limited" : "intent.rejection.exclusive_activity"));
      expect(read).toHaveBeenCalledTimes(1);
      expect(feed.disabled).toBe(false);
      expect(feed.getAttribute("aria-disabled")).toBe("true");
      expect(document.activeElement).toBe(feed);
      feed.click();
      const { userEvent } = await import("vitest/browser"); await userEvent.keyboard("{Enter}"); await settle();
      expect(runtime.requests).toHaveLength(1);
      release({ ...withPet(), founder_revision: 8 }); await settle();
      expect(runtime.requests).toHaveLength(1);
      expect(feed.getAttribute("aria-disabled")).not.toBe("true");
      expect(document.activeElement).toBe(feed);
      feed.click(); await settle();
      expect(runtime.requests).toHaveLength(2);
      expect(runtime.requests[1]).toMatchObject({ kind: "care_action", action_id: "care.feed", expected_revision: 8 });
      expect(typeof runtime.requests[0].intent_id).toBe("string");
      expect(runtime.requests[1].intent_id).not.toBe(runtime.requests[0].intent_id);
      expect(diagnostics).not.toHaveBeenCalled();
    } finally {
      release(runtime.current); await settle(); read.mockRestore(); intent.mockRestore(); diagnostics.mockRestore(); await dispose();
    }
  });
}

it.skipIf(!browser)("care shared refusal invalid request reports one invariant, stays online and never retries", async () => {
  const runtime = new Runtime(); runtime.current = withPet();
  const { target, dispose } = await mounted(runtime);
  const read = vi.spyOn(runtime, "snapshot");
  const intent = vi.spyOn(runtime, "intent").mockImplementation(async (body) => {
    runtime.requests.push(body); throw new GameUIRequestError(400, "invalid", "care_action");
  });
  const diagnostics = vi.spyOn(console, "error").mockImplementation(() => {});
  try {
    button(target, careText("pet.care.panel.title")).click(); await settle();
    const feed = button(target, careText("pet.care.action.feed.title")); feed.focus(); feed.click(); await settle();
    expect(runtime.requests).toHaveLength(1); expect(read).not.toHaveBeenCalled();
    expect(target.querySelector(".intent-notice")?.textContent).toBe(careText("intent.rejection.unknown"));
    expect(target.querySelector(".intent-notice")?.textContent).not.toContain("invalid/care_action");
    expect(diagnostics.mock.calls).toEqual([["game UI invariant: invalid intent response"]]);
    expect(feed.disabled).toBe(false); expect(document.activeElement).toBe(feed);
    button(target, careText("surface.settings.title")).click(); await settle();
    expect(target.textContent).not.toContain(careText("settings.save_status.offline"));
  } finally { read.mockRestore(); intent.mockRestore(); diagnostics.mockRestore(); await dispose(); }
});

for (const [label, error] of [
  ["401", new GameUIRequestError(401, "unauthenticated", "account")],
  ["404", new GameUIRequestError(404, "unknown_id", "account")],
  ["503", new GameUIRequestError(503, "unavailable", "server")],
  ["transport", new TypeError("test network failure")],
  ["unparsable", new SyntaxError("test malformed response")],
] as const) {
  it.skipIf(!browser)(`care shared refusal ${label} takes the offline path without replaying or leaking mechanical text`, async () => {
    const runtime = new Runtime(); runtime.current = withPet();
    const { target, dispose } = await mounted(runtime);
    const read = vi.spyOn(runtime, "snapshot");
    const intent = vi.spyOn(runtime, "intent").mockImplementation(async (body) => { runtime.requests.push(body); throw error; });
    const diagnostics = vi.spyOn(console, "error").mockImplementation(() => {});
    try {
      button(target, careText("pet.care.panel.title")).click(); await settle();
      button(target, careText("pet.care.action.feed.title")).click(); await settle();
      const actions = [...target.querySelectorAll<HTMLButtonElement>(".pet-care .actions button")];
      expect(actions).toHaveLength(5); expect(actions.every((action) => action.disabled)).toBe(true);
      for (const action of actions) action.click(); await settle();
      expect(runtime.requests).toHaveLength(1); expect(read).not.toHaveBeenCalled();
      expect(target.querySelector(".intent-notice")?.textContent).toBe("");
      expect(target.querySelector(".pet-care")?.textContent).toContain(careText("common.stale_note"));
      expect(target.textContent).not.toContain(error.message); expect(diagnostics).not.toHaveBeenCalled();
      button(target, careText("surface.settings.title")).click(); await settle();
      expect(target.textContent).toContain(careText("settings.save_status.offline"));
    } finally { read.mockRestore(); intent.mockRestore(); diagnostics.mockRestore(); await dispose(); }
  });
}

it.skipIf(!browser)("care supplement refuses pending activation inside the component, independent of the host queue", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const onCare = vi.fn();
  const app = mount(PetCareSurface, { target, props: { pets: [petRow], cosmetics: null, era: "era_1995", pending: true, controlsEnabled: true, reducedMotion: true, onCare } });
  try {
    await settle();
    const feed = button(target, careText("pet.care.action.feed.title")); feed.focus();
    expect(feed.disabled).toBe(false);
    expect(feed.getAttribute("aria-disabled")).toBe("true");
    feed.click(); await settle(); expect(onCare).not.toHaveBeenCalled();
    const { userEvent } = await import("vitest/browser");
    await userEvent.keyboard("{Enter}"); await userEvent.keyboard(" "); await settle();
    expect(onCare).not.toHaveBeenCalled(); expect(document.activeElement).toBe(feed);
  } finally { await unmount(app); target.remove(); }
});

it.skipIf(!browser)("care supplement does not steal focus from a nav control selected while the intent is pending", async () => {
  const runtime = new Runtime(); runtime.current = withPet();
  const { target, dispose } = await mounted(runtime);
  let finish!: (value: IntentOutcome) => void;
  const held = new Promise<IntentOutcome>((resolve) => { finish = resolve; });
  const intent = vi.spyOn(runtime, "intent").mockImplementation((body) => { runtime.requests.push(body); return held; });
  try {
    button(target, careText("pet.care.panel.title")).click(); await settle();
    const feed = button(target, careText("pet.care.action.feed.title")); feed.focus(); feed.click(); await settle();
    const nav = button(target, careText("surface.meters.title")); nav.focus();
    const next = withPet();
    runtime.current = { ...next, founder_revision: 8,
      features: { ...next.features, pet_adoption: { ...next.features.pet_adoption!, pets: [{ ...petRow, eligible_action_ids: [] }] } } };
    finish({ outcome: "applied", receipt: {} }); await settle();
    expect(feed.disabled).toBe(true);
    expect(document.activeElement).toBe(nav);
    expect(runtime.requests).toHaveLength(1);
  } finally { finish({ outcome: "applied", receipt: {} }); await settle(); intent.mockRestore(); await dispose(); }
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
it.skipIf(!browser)("GS6-A3 compares the whole 320 px Desk before and after its provision and owned text becomes visible", async ({ annotate }) => {
  const { page } = await import("vitest/browser");
  await page.viewport(320, 720);
  // Paired current-component fixtures, not a historical executable or a
  // producer/persisted fixture. The unchanged legacy shelf is in both arms.
  const after = parseGameUISnapshot(withOpportunity(opportunityArm(true, true, true)));
  const before = parseGameUISnapshot({ ...after,
    generators: after.generators.map((row) => ({ ...row, provisioned: 0 })),
    upgrades: after.upgrades.map((row) => ({ ...row, owned: false })),
  });
  const runtime = new Runtime(); runtime.current = before;
  const { target, app, dispose } = await mounted(runtime);
  const observations: Readonly<Record<string, unknown>>[] = [];
  try {
    const manual = target.querySelector<HTMLButtonElement>("section.manual button")!;
    manual.focus();
    const nav = [...target.querySelectorAll("nav button")].map((node) => node.textContent);
    const exactNode = (parent: Element, selector: string, text: string) => [...parent.querySelectorAll<HTMLElement>(selector)]
      .filter((node) => node.textContent === text);
    const isVisible = (node: HTMLElement) => {
      const rectangle = node.getBoundingClientRect(), style = getComputedStyle(node);
      expect(style.display).not.toBe("none"); expect(style.visibility).toBe("visible");
      expect(Number(style.opacity)).toBeGreaterThan(0);
      expect(rectangle.width).toBeGreaterThan(0); expect(rectangle.height).toBeGreaterThan(0);
    };
    const measure = (populated: boolean) => {
      const main = target.querySelector<HTMLElement>("main.game-ui")!, desk = main.querySelector<HTMLElement>("section.desk")!;
      const chrome = main.querySelector<HTMLElement>("header.chrome")!;
      expect(main.dataset.surface).toBe("desk");
      expect(document.documentElement.clientWidth).toBe(320);
      expect([...chrome.querySelectorAll("nav button")].map((node) => node.textContent)).toEqual(nav);
      expect(document.activeElement).toBe(manual);
      expect(runtime.requests).toEqual([]);
      isVisible(manual); isVisible(chrome);
      const generator = desk.querySelector<HTMLElement>('section[aria-labelledby="generators-heading"] .card')!;
      const upgrade = desk.querySelector<HTMLElement>('section[aria-labelledby="upgrades-heading"] .card')!;
      const provisions = exactNode(generator, "span", t("desk.provisioned_frame", { count: 3 }, "era_1995"));
      const reasons = exactNode(generator, "span", t("generator.beige_tower.provisioned_cap", {}, "era_1995"));
      const owned = exactNode(upgrade, "strong", t("desk.upgrade.owned", {}, "era_1995"));
      for (const nodes of [provisions, reasons, owned]) {
        expect(nodes).toHaveLength(populated ? 1 : 0);
        for (const node of nodes) isVisible(node);
      }
      const shelfTitle = exactNode(desk, "h2", t("cosmetic.horse_armor_free.title", {}, "era_1995"));
      expect(shelfTitle).toHaveLength(1); isVisible(shelfTitle[0]!);
      const shelf = shelfTitle[0]!.parentElement!;
      const curtain = exactNode(shelf, "small", t("cosmetic.horse_armor_free.disclosure", {}, "era_1995"));
      expect(curtain).toHaveLength(1); isVisible(curtain[0]!);
      expect(shelf.querySelectorAll("button")).toHaveLength(0);
      const opportunity = desk.querySelector<HTMLElement>('[data-region="desk.region.opportunity"]')!;
      expect(opportunity).not.toBeNull(); isVisible(opportunity);
      expect(desk.querySelectorAll('section[aria-labelledby="generators-heading"] .cards .card')).toHaveLength(before.generators.length);
      expect(desk.querySelectorAll('section[aria-labelledby="upgrades-heading"] .cards .card')).toHaveLength(before.upgrades.length);
      const nodes = [document.documentElement, document.body, main, ...main.querySelectorAll<HTMLElement>("*")];
      const visible = nodes.filter((node) => { const rect = node.getBoundingClientRect(); return rect.width > 0 && rect.height > 0; });
      const offenders = visible.filter((node) => {
        const rect = node.getBoundingClientRect();
        return rect.left < -1 || rect.right > 321 || node.clientWidth > 0 && node.scrollWidth > node.clientWidth + 1;
      }).map((node) => ({ tag: node.tagName, class: node.className, left: node.getBoundingClientRect().left,
        right: node.getBoundingClientRect().right, client: node.clientWidth, scroll: node.scrollWidth }));
      expect(offenders, "whole-page descendants must fit, not only the Desk box").toEqual([]);
      for (const added of [main, desk, chrome, ...provisions, ...reasons, ...owned]) {
        for (let node: HTMLElement | null = added; node !== null; node = node.parentElement) {
          expect(["hidden", "clip"], "overflow masking is not containment").not.toContain(getComputedStyle(node).overflowX);
        }
      }
      const widths = [document.documentElement, main, desk, chrome].map((node) => ({ client: node.clientWidth, scroll: node.scrollWidth }));
      const extents = { left: Math.min(...visible.map((node) => node.getBoundingClientRect().left)),
        right: Math.max(...visible.map((node) => node.getBoundingClientRect().right)) };
      observations.push({ population: populated ? "GS6-text-present" : "GS6-text-absent", widths, extents,
        measured_nodes: visible.length, provisioned: populated ? 3 : 0, owned: populated, intents: runtime.requests.length });
      return { widths, extents };
    };
    await settleMeterLayout(); const baseline = measure(false);
    app.fixtureSnapshot({ ...after, revision: before.revision + 1 });
    await settleMeterLayout(); const populated = measure(true);
    for (const [index, width] of populated.widths.entries()) {
      expect(width.scroll).toBeLessThanOrEqual(baseline.widths[index]!.scroll + 1);
      expect(width.client).toBe(baseline.widths[index]!.client);
    }
    expect(populated.extents.right).toBeLessThanOrEqual(baseline.extents.right + 1);
    expect(populated.extents.left).toBeGreaterThanOrEqual(baseline.extents.left - 1);
    console.info("GS6-A3 paired current-source Desk measurement", JSON.stringify(observations));
    await assertAxe(target, "GS6 current-source paired Desk after");
  } finally {
    try { await dispose(); } finally { await page.viewport(1280, 720); }
    await annotate(`GS6-A3 fixture comparison: ${JSON.stringify(observations)}`);
  }
});

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

// RP-326: consumer placement and origin boundaries only. These existing
// runtime-double public fixtures are not producer/catalog or real-AT proof.
for (const owner of ["fiscal", "care"] as const) {
  for (const result of ["applied", "refused"] as const) {
    const outcome: IntentOutcome = result === "applied"
      ? { outcome: "applied", receipt: { harvest_outcome: "guaranteed" } }
      : { outcome: "rejected", category: "not_eligible", detail: owner === "fiscal" ? "period_not_ripe" : "cooldown", currentRevision: 7, sessionExpired: false };
    const noticeKey: CopyKey = owner === "fiscal"
      ? result === "applied" ? "fiscal.outcome.guaranteed" : "fiscal.rejection.period_not_ripe"
      : result === "applied" ? "pet.care.applied" : "pet.care.rejection.cooldown";
    const titleKey: CopyKey = owner === "fiscal" ? "surface.fiscal.title" : "pet.care.panel.title";
    const actionKey: CopyKey = owner === "fiscal" ? "fiscal.harvest" : "pet.care.action.feed.title";
    const selector = owner === "fiscal" ? ".fiscal" : ".pet-care";
    const runtimeForOwner = () => {
      const runtime = new Runtime(), value = withPet();
      runtime.current = { ...value, features: { ...value.features, fiscal: withRipeFiscal().features.fiscal } };
      runtime.outcome = outcome;
      return runtime;
    };
    const assertRequest = (runtime: Runtime) => {
      expect(runtime.requests).toHaveLength(1);
      expect(runtime.requests[0]).toMatchObject(owner === "fiscal"
        ? { kind: "harvest_fiscal_period", expected_revision: 7 }
        : { kind: "care_action", pet_id: petRow.pet_id, action_id: "care.feed", expected_revision: 7 });
    };
    const activate = async (target: HTMLElement) => {
      button(target, t(titleKey, {}, "era_1995")).click(); await settle();
      const action = button(target, t(actionKey, {}, "era_1995"));
      action.focus();
      const { userEvent } = await import("vitest/browser");
      await userEvent.keyboard("{Enter}"); await settle();
    };

    it.skipIf(!browser)(`outcome-ownership ${owner}/${result} has one exact own-panel polite result, not chrome`, async () => {
      const runtime = runtimeForOwner(), { target, dispose } = await mounted(runtime);
      try {
        await activate(target); assertRequest(runtime);
        const message = t(noticeKey, {}, "era_1995");
        const regions = [...target.querySelectorAll<HTMLElement>("[role=status]")].filter((node) => node.textContent?.trim() === message);
        expect(regions).toHaveLength(1);
        expect(target.querySelector(selector)?.contains(regions[0]!), "outcome belongs to its panel").toBe(true);
        expect(target.querySelector(".chrome")?.textContent).not.toContain(message);
      } finally { await dispose(); }
    });

    it.skipIf(!browser)(`outcome-ownership ${owner}/${result} completed result does not bleed into a different tab`, async () => {
      const runtime = runtimeForOwner(), { target, dispose } = await mounted(runtime);
      try {
        await activate(target); assertRequest(runtime);
        const message = t(noticeKey, {}, "era_1995");
        expect(target.textContent).toContain(message);
        const nav = button(target, t("surface.meters.title", {}, "era_1995")); nav.focus(); nav.click(); await settle();
        expect(target.querySelector(".meters")).not.toBeNull();
        expect(document.activeElement).toBe(nav);
        expect(target.textContent).not.toContain(message);
      } finally { await dispose(); }
    });

    it.skipIf(!browser)(`outcome-ownership ${owner}/${result} late completion does not announce on a different tab`, async () => {
      const runtime = runtimeForOwner(), { target, dispose } = await mounted(runtime);
      let finish!: (value: IntentOutcome) => void;
      const held = new Promise<IntentOutcome>((resolve) => { finish = resolve; });
      const intent = vi.spyOn(runtime, "intent").mockImplementation((body) => { runtime.requests.push(body); return held; });
      try {
        await activate(target); assertRequest(runtime);
        const nav = button(target, t("surface.meters.title", {}, "era_1995")); nav.focus(); nav.click(); await settle();
        expect(target.querySelector(".meters")).not.toBeNull();
        finish(outcome); await settle();
        expect(document.activeElement).toBe(nav);
        expect(target.textContent).not.toContain(t(noticeKey, {}, "era_1995"));
        assertRequest(runtime);
      } finally { finish(outcome); await settle(); intent.mockRestore(); await dispose(); }
    });
  }
}

// GS0.2/GS0.6: thrown runtime errors, not actual fetch/server refusal proof.
const httpErrorArms = [
  { label: "400", error: new GameUIRequestError(400, "invalid", "intent"), notice: "intent.rejection.unknown", effect: "none" },
  { label: "409", error: new GameUIRequestError(409, "conflict", "intent"), notice: "intent.conflict", effect: "refresh" },
  { label: "429", error: new GameUIRequestError(429, "rate_limited", "account"), notice: "intent.rate_limited", effect: "refresh" },
  { label: "401", error: new GameUIRequestError(401, "unauthenticated", "account"), notice: null, effect: "offline" },
  { label: "404", error: new GameUIRequestError(404, "unknown_id", "account"), notice: null, effect: "offline" },
  { label: "503", error: new GameUIRequestError(503, "unavailable", "server"), notice: null, effect: "offline" },
  { label: "transport", error: new TypeError("test network failure"), notice: null, effect: "offline" },
  { label: "malformed", error: new SyntaxError("test malformed response"), notice: null, effect: "offline" },
] as const;

for (const owner of ["fiscal", "care"] as const) {
  for (const timing of ["immediate", "late-away"] as const) {
    for (const arm of httpErrorArms) {
      it.skipIf(!browser)(`HTTP-error ownership ${owner}/${arm.label}/${timing} preserves reason, recovery and explicit consent`, async () => {
        const runtime = new Runtime(), value = withPet();
        runtime.current = { ...value, features: { ...value.features, fiscal: withRipeFiscal().features.fiscal } };
        const { target, dispose } = await mounted(runtime);
        let rejectResponse!: (reason: unknown) => void;
        const response = new Promise<IntentOutcome>((_resolve, reject) => { rejectResponse = reject; });
        let releaseRead!: (value: ParsedGameUISnapshot) => void;
        const refresh = new Promise<ParsedGameUISnapshot>((resolve) => { releaseRead = resolve; });
        const read = vi.spyOn(runtime, "snapshot").mockReturnValue(refresh);
        const diagnostics = vi.spyOn(console, "error").mockImplementation(() => {});
        const intent = vi.spyOn(runtime, "intent").mockImplementation(async (body) => {
          runtime.requests.push(body);
          if (runtime.requests.length > 1) return { outcome: "applied", receipt: { harvest_outcome: "guaranteed" } };
          if (timing === "late-away") return response;
          throw arm.error;
        });
        const title = careText(owner === "fiscal" ? "surface.fiscal.title" : "pet.care.panel.title");
        const actionText = careText(owner === "fiscal" ? "fiscal.harvest" : "pet.care.action.feed.title");
        const selector = owner === "fiscal" ? ".fiscal" : ".pet-care";
        try {
          button(target, title).click(); await settle();
          const action = button(target, actionText); action.focus();
          const { userEvent } = await import("vitest/browser");
          await userEvent.keyboard("{Enter}"); await settle();
          expect(runtime.requests).toHaveLength(1);
          expect(runtime.requests[0]).toMatchObject(owner === "fiscal"
            ? { kind: "harvest_fiscal_period", expected_revision: 7 }
            : { kind: "care_action", action_id: "care.feed", pet_id: petRow.pet_id, expected_revision: 7 });
          if (timing === "late-away") {
            expect(read).not.toHaveBeenCalled(); expect(diagnostics).not.toHaveBeenCalled();
            const nav = button(target, careText("surface.meters.title")); nav.focus(); nav.click(); await settle();
            rejectResponse(arm.error); await settle();
            expect(target.querySelector(".meters")).not.toBeNull();
            expect(document.activeElement).toBe(nav);
            if (arm.notice) expect(target.textContent).not.toContain(careText(arm.notice));
            button(target, title).click(); await settle();
          }
          expect(runtime.requests).toHaveLength(1);
          expect(target.textContent).not.toContain(arm.error.message);
          expect(diagnostics.mock.calls).toEqual(arm.label === "400" ? [["game UI invariant: invalid intent response"]] : []);
          if (arm.notice) {
            const message = careText(arm.notice);
            const regions = [...target.querySelectorAll<HTMLElement>("[role=status]")].filter((node) => node.textContent?.trim() === message);
            expect(regions).toHaveLength(1);
            expect(target.querySelector(selector)?.contains(regions[0]!)).toBe(true);
            expect(target.querySelector(".chrome")?.textContent).not.toContain(message);
          } else {
            expect(target.querySelector(`${selector} .intent-notice`)?.textContent).toBe("");
          }
          const currentAction = button(target, actionText);
          if (arm.effect === "refresh") {
            expect(read).toHaveBeenCalledTimes(1);
            expect(currentAction.disabled).toBe(false);
            expect(currentAction.getAttribute("aria-disabled")).toBe("true");
            currentAction.focus(); currentAction.click(); await userEvent.keyboard("{Enter}"); await settle();
            expect(runtime.requests).toHaveLength(1);
            runtime.current = { ...runtime.current, founder_revision: 8 };
            releaseRead(runtime.current); await settle();
            expect(runtime.requests).toHaveLength(1);
            expect(currentAction.getAttribute("aria-disabled")).not.toBe("true");
            expect(document.activeElement).toBe(currentAction);
            await userEvent.keyboard("{Enter}"); await settle();
            expect(runtime.requests).toHaveLength(2);
            expect(runtime.requests[1]).toMatchObject({ expected_revision: 8 });
            expect(typeof runtime.requests[0].intent_id).toBe("string");
            expect(runtime.requests[1].intent_id).not.toBe(runtime.requests[0].intent_id);
          } else {
            expect(read).not.toHaveBeenCalled();
            expect(currentAction.disabled).toBe(arm.effect === "offline");
            if (arm.effect === "offline") { currentAction.click(); await settle(); expect(runtime.requests).toHaveLength(1); }
            button(target, careText("surface.settings.title")).click(); await settle();
            if (arm.effect === "offline") expect(target.textContent).toContain(careText("settings.save_status.offline"));
            else expect(target.textContent).not.toContain(careText("settings.save_status.offline"));
          }
        } finally {
          if (timing === "late-away") rejectResponse(arm.error);
          releaseRead(runtime.current); await settle();
          read.mockRestore(); diagnostics.mockRestore(); intent.mockRestore(); await dispose();
        }
      });
    }
  }
}
