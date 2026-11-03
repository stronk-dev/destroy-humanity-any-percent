import { flushSync, mount, unmount } from "svelte";
import { expect, it, vi } from "vitest";
import type { GameUISnapshot } from "../src/api/generated/types";
import { t, type CopyKey } from "../src/copy";
import { parseGameUISnapshot, type ParsedGameUISnapshot } from "../src/game-ui/contracts";
import GameUIApp from "../src/game-ui/GameUIApp.svelte";
import { createBrowserGameUIRuntime } from "../src/game-ui/runtime";
import { Runtime, mounted, settle, v4 } from "./garage-surfaces-fixture";
import { browser, assertAxe, button, sharedStateVisibleText } from "./garage-surfaces-assertions";

// Same Garage cases, grouped by outcome so one serial file cannot dominate CI.


declare module "vitest" {
  interface TaskMeta { firstReadObservations?: Record<string, unknown>[] }
}


it.skipIf(!browser)("unlocks one nav button per live feature fact and renders the Meters board as text (GS3)", async () => {
  const { target, dispose } = await mounted();
  try {
    const nav = [...target.querySelectorAll("nav button")].map((node) => node.textContent?.trim());
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
      // Required loading text must remain painted, not merely laid out inside
      // a transparent host. Exercise the exact shared assertion used above.
      const main = target.querySelector<HTMLElement>("main")!;
      const originalOpacity = main.style.opacity;
      try {
        main.style.opacity = "0";
        await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
        expect(getComputedStyle(main.querySelector("[role=status]")!).opacity).toBe("1");
        expect(() => sharedStateVisibleText(target, '[role="status"]', t("common.loading", {}, "era_1995")))
          .toThrow("shared-state ancestor opacity");
        main.style.opacity = "0.5";
        sharedStateVisibleText(target, '[role="status"]', t("common.loading", {}, "era_1995"));
      } finally { main.style.opacity = originalOpacity; }
      sharedStateVisibleText(target, '[role="status"]', t("common.loading", {}, "era_1995"));
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
