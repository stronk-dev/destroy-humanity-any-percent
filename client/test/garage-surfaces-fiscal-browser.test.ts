import { flushSync, mount, unmount } from "svelte";
import { expect, it, vi } from "vitest";
import { t, type CopyKey } from "../src/copy";
import { parseGameUISnapshot, type ParsedGameUISnapshot } from "../src/game-ui/contracts";
import FiscalSurface from "../src/game-ui/FiscalSurface.svelte";
import { type IntentOutcome } from "../src/game-ui/intent-outcome";
import { NOW, Runtime, mounted, settle, v4 } from "./garage-surfaces-fixture";
import { browser, assertAxe, button, sharedStateVisibleText, withRipeFiscal } from "./garage-surfaces-assertions";

// Same Garage cases, grouped by outcome so one serial file cannot dominate CI.


for (const width of [320, 1280] as const) {
  it.skipIf(!browser)(`GS1/OD-3 Fiscal ready navigation follows exact phase edges silently at ${width}px`, async () => {
    const { page, userEvent } = await import("vitest/browser");
    await page.viewport(width, 720);
    const runtime = new Runtime();
    const reads = vi.spyOn(runtime, "snapshot");
    const { target, app, dispose } = await mounted(runtime);
    try {
      const nav = [...target.querySelectorAll<HTMLButtonElement>("nav button")]
        .find((node) => node.textContent?.trim() === "Earnings Calls")!;
      const desk = button(target, "The Desk");
      desk.focus();
      const initialReads = reads.mock.calls.length;
      const announcements = [...target.querySelectorAll('[role="status"], [role="alert"], [aria-live]')]
        .map((node) => node.textContent);
      // Synchronous host-clock fixtures: the real 100ms tick cannot cross an
      // exact boundary between setting the estimate and checking the DOM.
      for (const [elapsed, ready] of [[99_999, false], [100_000, true], [199_999, true], [200_000, true]] as const) {
        app.fixtureMonotonicElapsed(elapsed); flushSync();
        const badge = nav.querySelector<HTMLElement>(".fiscal-ripe-badge");
        expect(badge !== null, `readiness at ${elapsed}ms`).toBe(ready);
        if (badge) {
          expect(badge.textContent).toBe("PENDING OWNER COPY: ready to harvest");
          expect(badge.closest('[role="status"], [role="alert"], [aria-live]')).toBeNull();
          expect(badge.getAttribute("aria-hidden")).not.toBe("true");
          expect(badge.getBoundingClientRect().width).toBeGreaterThan(0);
          expect(badge.getBoundingClientRect().height).toBeGreaterThan(0);
        }
        expect(document.activeElement).toBe(desk);
        expect(target.querySelector("main")?.getAttribute("data-surface")).toBe("desk");
        expect([...target.querySelectorAll('[role="status"], [role="alert"], [aria-live]')].map((node) => node.textContent))
          .toEqual(announcements);
        expect(runtime.requests).toHaveLength(0);
        expect(reads.mock.calls).toHaveLength(initialReads);
      }

      // An authoritative new period clears readiness; visiting the old ready
      // period does not claim a harvest or consume its indication.
      const ready = { ...v4, features: { ...v4.features, fiscal: { ...v4.features.fiscal!,
        period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - 150_000 } } } };
      app.fixtureSnapshot(ready); flushSync();
      expect(nav.textContent?.replace(/\s+/gu, " ").trim()).toBe("Earnings Calls PENDING OWNER COPY: ready to harvest");
      expect(nav.hasAttribute("aria-label")).toBe(false);
      expect(nav.hasAttribute("aria-labelledby")).toBe(false);
      runtime.listener?.({ kind: "announcement", scope: "founder", value: { cursor: 42, kind: "fiscal_period_harvested", payload: { source: "automatic", credit_after: 10 } } });
      await settle();
      expect(nav.querySelector(".nav-badge")?.textContent).toBe("(harvested)");
      nav.focus(); await userEvent.keyboard("{Enter}"); await settle();
      expect(target.querySelector(".fiscal")?.getAttribute("data-phase")).toBe("early");
      expect(nav.querySelector(".fiscal-ripe-badge")).not.toBeNull();
      expect(nav.querySelector(".nav-badge")).toBeNull();
      expect(document.activeElement).toBe(nav);
      expect(runtime.requests).toHaveLength(0);
      expect(target.scrollWidth).toBeLessThanOrEqual(width);
      await assertAxe(target, `Fiscal ready navigation ${width}`);
      app.fixtureSnapshot({ ...v4, founder_revision: 8, features: { ...v4.features,
        fiscal: { ...v4.features.fiscal!, period: { ...v4.features.fiscal!.period, seq: 1 } } } });
      flushSync();
      expect(nav.querySelector(".fiscal-ripe-badge")).toBeNull();
      app.fixtureSnapshot({ ...ready, features: { ...ready.features, fiscal: null } }); flushSync();
      expect(nav.querySelector(".fiscal-ripe-badge")).toBeNull();
      app.fixtureSnapshot({ ...ready, facts: ready.facts.map((fact) => fact.fact_id === "feature.fiscal" ? { ...fact, value: false } : fact) });
      flushSync();
      expect(nav.isConnected).toBe(false);
    } finally { reads.mockRestore(); await dispose(); }
  });
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
      const visiblePhase = () => sharedStateVisibleText(surface, ".phase", fixture.text);
      const phase = visiblePhase();
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
      const originalOpacity = surface.style.opacity;
      try {
        surface.style.opacity = "0";
        await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
        expect(getComputedStyle(phase).opacity).toBe("1");
        expect(() => visiblePhase()).toThrow("shared-state ancestor opacity");
        surface.style.opacity = "0.5"; visiblePhase();
      } finally { surface.style.opacity = originalOpacity; }
      visiblePhase();
    } finally { await unmount(app); target.remove(); }
  });
}

it.skipIf(!browser)("RP-378 disabled Fiscal harvest ignores focus and sends Enter to the prior navigation control", async () => {
  const { userEvent } = await import("vitest/browser");
  const { target, app, runtime, dispose } = await mounted();
  try {
    const nav = button(target, "Earnings Calls");
    nav.focus(); await userEvent.keyboard("{Enter}"); await settle();
    const harvest = target.querySelector<HTMLButtonElement>('[data-fiscal-action="harvest"]')!;
    expect(target.querySelector(".fiscal")?.getAttribute("data-phase")).toBe("ripening");
    expect(harvest.disabled).toBe(true);
    nav.focus();
    // The old composed sequence focuses without waiting for eligibility. The
    // browser ignores that focus, so Enter activates the still-focused nav.
    harvest.focus();
    expect(document.activeElement).toBe(nav);
    await userEvent.keyboard("{Enter}"); await settle();
    expect(runtime.requests).toHaveLength(0);
    expect(target.querySelector("main")?.getAttribute("data-surface")).toBe("fiscal");

    // A controlled eligible phase, not a longer wall-clock sleep or UI change.
    app.fixtureSnapshot({ ...v4, features: { ...v4.features, fiscal: { ...v4.features.fiscal!,
      period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - 250_000 },
    } } });
    await settle();
    expect(harvest.disabled).toBe(false);
    expect(runtime.requests).toHaveLength(0);
    runtime.outcome = { outcome: "rejected", category: "not_eligible", detail: "period_not_ripe", currentRevision: 7, sessionExpired: false };
    harvest.focus();
    expect(document.activeElement).toBe(harvest);
    await userEvent.keyboard("{Enter}"); await settle();
    expect(runtime.requests).toHaveLength(1);
    expect(runtime.requests[0]).toMatchObject({ kind: "harvest_fiscal_period", expected_revision: 7 });
    expect(target.querySelector(".fiscal .intent-notice")?.textContent).toBe("It is too soon for an earnings call.");
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

for (const [beforeCommit, nextCompany, failedFreshRead] of [
  [false, false, false], [true, false, false], [true, true, false], [true, false, true],
] as const) {
it.skipIf(!browser)(`${beforeCommit ? "refreshes Fiscal again when a shared read predates harvest" : "keeps Fiscal pending until the applied harvest refresh supplies the next Founder revision"}${nextCompany ? " across Companies" : failedFreshRead ? " with a failed fresh read" : ""} (GS0.2/GS1-A6)`, async () => {
  const { userEvent } = await import("vitest/browser");
  class DelayedFiscalRuntime extends Runtime {
    snapshotCalls = 0;
    releaseIntent: () => void = () => {};
    releaseReads: (() => void)[] = [];
    override async intent(body: Readonly<Record<string, unknown>>): Promise<IntentOutcome> {
      this.requests.push(body);
      if (body.kind === "harvest_fiscal_period") {
        if (beforeCommit) await new Promise<void>((resolve) => { this.releaseIntent = resolve; });
        this.current = { ...this.current, founder_revision: 8 };
        return { outcome: "applied", receipt: { harvest_outcome: "guaranteed", founder_revision: 8 } };
      }
      this.current = { ...this.current, founder_revision: 9 };
      return { outcome: "applied", receipt: { founder_revision: 9 } };
    }
    override async snapshot(): Promise<ParsedGameUISnapshot> {
      if (this.requests.length === 0) return this.current;
      this.snapshotCalls += 1;
      if (failedFreshRead && this.snapshotCalls === 2) throw new Error("injected fresh Founder read failure");
      const sampled = this.current;
      if (this.snapshotCalls <= (beforeCommit ? 2 : 1)) await new Promise<void>((resolve) => { this.releaseReads.push(resolve); });
      return sampled;
    }
  }
  const runtime = new DelayedFiscalRuntime();
  runtime.current = { ...v4, features: { ...v4.features, fiscal: { ...v4.features.fiscal!, period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - 250_000 } } } };
  const { target, app, dispose } = await mounted(runtime);
  try {
    button(target, "Earnings Calls").click();
    app.fixtureSnapshot({ ...v4, features: { ...v4.features, fiscal: { ...v4.features.fiscal!, period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - 250_000 } } } });
    await settle();
    button(target, "Hold the earnings call").focus(); await userEvent.keyboard("{Enter}");
    await settle();
    expect(runtime.requests[0]).toMatchObject({ kind: "harvest_fiscal_period", expected_revision: 7 });
    if (beforeCommit) {
      if (nextCompany) runtime.current = { ...runtime.current, revision: 1, run: { ...runtime.current.run, run_seq: 2 } };
      runtime.listener?.({ kind: "receipt" }); await settle();
      expect(runtime.snapshotCalls).toBe(1);
      runtime.releaseIntent(); await settle();
    }
    expect(runtime.snapshotCalls).toBe(1);
    expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
    expect(button(target, "Unlock for 3").disabled).toBe(false);
    expect(button(target, "Unlock for 3").getAttribute("aria-disabled")).toBe("true");
    runtime.releaseReads.splice(0).forEach((release) => release());
    await settle();
    if (beforeCommit) {
      expect(runtime.snapshotCalls, "a pre-harvest read cannot supply the applied Founder revision").toBe(2);
      expect(runtime.requests).toHaveLength(1);
      if (failedFreshRead) {
        expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
        expect(button(target, "Unlock for 3").disabled).toBe(true);
        await settle(); expect(runtime.requests).toHaveLength(1);
        return;
      }
      expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
      runtime.releaseReads.splice(0).forEach((release) => release()); await settle();
    }
    expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
    button(target, "Unlock for 3").focus(); await userEvent.keyboard("{Enter}");
    await settle();
    expect(runtime.requests[1]).toMatchObject({ kind: "spend_fiscal_credit", expected_revision: 8, target: { kind: "unlock", unlock_id: "minigame.pitch" } });
  } finally {
    runtime.releaseIntent(); runtime.releaseReads.splice(0).forEach((release) => release()); await settle();
    runtime.releaseReads.splice(0).forEach((release) => release()); await dispose();
  }
});
}

// Fiscal supplement uses public snapshot fixtures and runtime doubles. It is
// consumer evidence, never a real quarter clock or persisted Founder proof.
const fiscalText = (key: CopyKey): string => t(key, {}, "era_1995");
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

// GS0.2/0.6/0.8: actual native host entry and delayed completion. Public
// decoder fixtures deliberately retain eligibility; not a persisted spend,
// natural quarter, disabled/removal policy, auth or assistive-technology proof.
for (const [label, index, category, detail, key, body] of [
  ["harvest", 0, "not_eligible", "period_not_ripe", "fiscal.rejection.period_not_ripe", { kind: "harvest_fiscal_period" }],
  ["level", 1, "unaffordable", "fiscal_credit", "fiscal.rejection.unaffordable", { kind: "spend_fiscal_credit", target: { kind: "generator_level", generator_id: "generator.beige_tower", levels: 1 } }],
  ["unlock", 2, "not_eligible", "already_unlocked", "fiscal.rejection.already_unlocked", { kind: "spend_fiscal_credit", target: { kind: "unlock", unlock_id: "minigame.pitch" } }],
] as const) {
  for (const width of [320, 1280] as const) {
    for (const activation of ["{Enter}", " "] as const) {
      it.skipIf(!browser)(`GS1 native pending ${label}/${width}/${activation === " " ? "Space" : "Enter"}`, async () => {
        const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
        const runtime = new Runtime(); runtime.current = parseGameUISnapshot(structuredClone(withRipeFiscal()));
        const rejected: IntentOutcome = { outcome: "rejected", category, detail, currentRevision: 7, sessionExpired: false };
        const applied: IntentOutcome = { outcome: "applied", receipt: label === "harvest" ? { harvest_outcome: "guaranteed" } : {} };
        runtime.outcome = rejected;
        let finishRejected!: (value: IntentOutcome) => void, finishApplied!: (value: IntentOutcome) => void;
        let releaseRead!: (value: ParsedGameUISnapshot) => void;
        const refused = new Promise<IntentOutcome>((resolve) => { finishRejected = resolve; });
        const accepted = new Promise<IntentOutcome>((resolve) => { finishApplied = resolve; });
        const heldRead = new Promise<ParsedGameUISnapshot>((resolve) => { releaseRead = resolve; });
        let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
        let restore = () => {};
        try {
          fixture = await mounted(runtime); const { target } = fixture;
          const intent = vi.spyOn(runtime, "intent")
            .mockImplementationOnce((request) => { runtime.requests.push(request); return refused; })
            .mockImplementationOnce((request) => { runtime.requests.push(request); return accepted; });
          const read = vi.spyOn(runtime, "snapshot").mockReturnValueOnce(heldRead);
          restore = () => { intent.mockRestore(); read.mockRestore(); };
          const nav = button(target, fiscalText("surface.fiscal.title"));
          const settings = button(target, fiscalText("surface.settings.title"));
          const reach = async (destination: HTMLElement, stage: string) => {
            const controls = target.querySelectorAll("button,input,summary,[tabindex='0']").length;
            const trace: (string | undefined)[] = [];
            let tabs = 0;
            while (document.activeElement !== destination && tabs <= controls) {
              await userEvent.keyboard("{Tab}"); await settle(); tabs += 1;
              trace.push(document.activeElement?.tagName + ":" + document.activeElement?.textContent);
            }
            expect(document.activeElement, `${stage} native Tab ${tabs}/${controls + 1}: ${JSON.stringify(trace)}`).toBe(destination);
          };
          const request = (revision: number) => ({ intent_id: expect.any(String), expected_revision: revision, ...body });
          button(target, fiscalText("surface.desk.title")).focus();
          await reach(nav, "initial Fiscal nav"); await userEvent.keyboard(activation); await settle();
          expect(target.querySelector("main")?.dataset.surface).toBe("fiscal"); expect(document.activeElement).toBe(nav);
          const action = fiscalControls(target)[index];
          await reach(action, "initial Fiscal action");
          const pending = () => {
            expect(action.disabled).toBe(false); expect(action.getAttribute("aria-disabled")).toBe("true");
            expect(document.activeElement).toBe(action);
            sharedStateVisibleText(target, ".fiscal .fiscal-state", fiscalText("common.pending"));
            expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
          };
          await userEvent.keyboard(activation); await settle(); pending();
          await userEvent.keyboard("{Enter}"); await userEvent.keyboard(" "); await settle();
          expect(runtime.requests).toEqual([request(7)]); expect(read).not.toHaveBeenCalled();
          finishRejected(rejected); await settle();
          expect(action.hasAttribute("aria-disabled")).toBe(false); expect(action.disabled).toBe(false);
          expect(document.activeElement).toBe(action); expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
          sharedStateVisibleText(target, '.fiscal [role="status"] .intent-notice', fiscalText(key));
          expect(target.querySelector(".fiscal")?.textContent).not.toContain(fiscalText("common.pending"));
          expect(read).not.toHaveBeenCalled(); await assertAxe(target, `native Fiscal refusal ${label}/${width}`);

          await userEvent.keyboard(activation); await settle(); pending();
          expect(runtime.requests).toEqual([request(7), request(7)]);
          expect(target.querySelector(".fiscal .intent-notice")?.textContent).toBe("");
          finishApplied(applied); await settle(); pending(); expect(read).toHaveBeenCalledExactlyOnceWith();
          expect(target.querySelector('.fiscal [role="status"] .intent-notice')?.textContent)
            .toBe(label === "harvest" ? fiscalText("fiscal.outcome.guaranteed") : "");
          await userEvent.keyboard("{Enter}"); await userEvent.keyboard(" "); await settle();
          expect(runtime.requests).toEqual([request(7), request(7)]);
          await reach(settings, "newer Settings selection"); await userEvent.keyboard(activation); await settle();
          expect(target.querySelector("main")?.dataset.surface).toBe("settings"); expect(document.activeElement).toBe(settings);
          expect(target.querySelector('.intent-notice[role="status"]')?.textContent).toBe("");
          sharedStateVisibleText(target, '[aria-labelledby="settings-heading"] p', fiscalText("settings.save_status.saving"));
          const next = structuredClone(withRipeFiscal()); next.founder_revision = 8; next.revision = 2;
          runtime.current = parseGameUISnapshot(next); releaseRead(runtime.current); await settle();
          expect(target.querySelector("main")?.dataset.surface).toBe("settings"); expect(document.activeElement).toBe(settings);
          expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
          expect(target.querySelector('.intent-notice[role="status"]')?.textContent).toBe("");
          const status = target.querySelector<HTMLElement>('[aria-labelledby="settings-heading"] p')!;
          const [prefix, suffix] = t("settings.save_status.saved_frame", { ago: "__AGO__" }, "era_1995").split("__AGO__");
          const text = status.textContent!;
          expect(text.startsWith(prefix!)).toBe(true); expect(text.endsWith(suffix!)).toBe(true);
          const ago = text.slice(prefix!.length, suffix!.length ? -suffix!.length : undefined);
          expect(ago).toMatch(/^\d+:[0-5]\d:[0-5]\d$/u);
          sharedStateVisibleText(target, '[aria-labelledby="settings-heading"] p', t("settings.save_status.saved_frame", { ago }, "era_1995"));
          await assertAxe(target, `native Fiscal newer Settings ${label}/${width}`);
          await reach(nav, "return Fiscal nav"); await userEvent.keyboard(activation); await settle();
          const retry = fiscalControls(target)[index]; await reach(retry, "fresh Fiscal action");
          expect(retry.disabled).toBe(false); expect(retry.hasAttribute("aria-disabled")).toBe(false);
          runtime.outcome = { ...rejected, currentRevision: 8 };
          await userEvent.keyboard(activation); await settle();
          expect(runtime.requests).toEqual([request(7), request(7), request(8)]);
          expect(new Set(runtime.requests.map((row) => row.intent_id)).size).toBe(3);
          expect(read).toHaveBeenCalledExactlyOnceWith(); expect(document.activeElement).toBe(retry);
          sharedStateVisibleText(target, '.fiscal [role="status"] .intent-notice', fiscalText(key));
          expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(width + 1);
          await assertAxe(target, `native Fiscal retry ${label}/${width}`);
        } finally {
          finishRejected(rejected); finishApplied(applied); releaseRead(runtime.current); await settle();
          restore();
          try { if (fixture) await fixture.dispose(); }
          finally { await page.viewport(1280, 720); }
        }
      });
    }
  }
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
