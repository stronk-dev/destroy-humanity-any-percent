import { expect, it, vi } from "vitest";
import { t, type CopyKey } from "../src/copy";
import { parseGameUISnapshot, type ParsedGameUISnapshot } from "../src/game-ui/contracts";
import { Runtime, meterRow, mounted, settle, v4 } from "./garage-surfaces-fixture";
import { browser, assertAxe, button, sharedStateVisibleText, semanticMeterRows, semanticMeterSnapshot, settleMeterLayout, assertSemanticMeters } from "./garage-surfaces-assertions";

// Same Garage cases, grouped by outcome so one serial file cannot dominate CI.


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
