import { expect, it, vi } from "vitest";
import type { GameUISnapshot } from "../src/api/generated/types";
import { t, type CopyKey } from "../src/copy";
import { parseGameUISnapshot, type ParsedGameUISnapshot } from "../src/game-ui/contracts";
import { GameUIRequestError, type IntentOutcome } from "../src/game-ui/intent-outcome";
import { Runtime, mounted, opportunityArm, settle, withOpportunity } from "./garage-surfaces-fixture";
import { browser, assertAxe, button, settleMeterLayout, withRipeFiscal, petRow, withPet, careText } from "./garage-surfaces-assertions";

// Same Garage cases, grouped by outcome so one serial file cannot dominate CI.


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
      for (let ancestor: HTMLElement | null = node; ancestor; ancestor = ancestor.parentElement) {
        expect(Number(getComputedStyle(ancestor).opacity), "required content ancestor opacity").toBeGreaterThan(0);
      }
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
    // A laid-out child keeps opacity:1 when its card is transparent. Exercise
    // the same visibility oracle on the actual required provision text so
    // rectangles/text presence cannot stand in for painted content.
    const generator = target.querySelector<HTMLElement>('section[aria-labelledby="generators-heading"] .card')!;
    const provision = exactNode(generator, "span", t("desk.provisioned_frame", { count: 3 }, "era_1995"))[0]!;
    const originalOpacity = generator.style.opacity;
    try {
      generator.style.opacity = "0"; await settleMeterLayout();
      expect(getComputedStyle(provision).opacity).toBe("1");
      expect(provision.getBoundingClientRect().width).toBeGreaterThan(0);
      expect(() => isVisible(provision), "transparent ancestors must invalidate required text").toThrow("required content ancestor opacity");
      generator.style.opacity = "0.5"; await settleMeterLayout();
      isVisible(provision);
    } finally {
      generator.style.opacity = originalOpacity; await settleMeterLayout();
    }
    isVisible(provision);
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
  for (const firstState of ["pending", "stale"] as const) {
    it.skipIf(!browser)(`single surface feedback ${owner}/${firstState} retains one live region across recovery and completion`, async () => {
      const runtime = new Runtime(), value = withPet();
      runtime.current = { ...value, features: { ...value.features, fiscal: withRipeFiscal().features.fiscal } };
      const { target, dispose } = await mounted(runtime);
      const selector = owner === "fiscal" ? ".fiscal" : ".pet-care";
      const title = careText(owner === "fiscal" ? "surface.fiscal.title" : "pet.care.panel.title");
      const actionText = careText(owner === "fiscal" ? "fiscal.harvest" : "pet.care.action.feed.title");
      const noticeKey: CopyKey = owner === "fiscal" ? "fiscal.outcome.guaranteed" : "pet.care.applied";
      let finish!: (value: IntentOutcome) => void;
      const response = new Promise<IntentOutcome>((resolve) => { finish = resolve; });
      let releaseRead!: (value: ParsedGameUISnapshot) => void;
      const refresh = new Promise<ParsedGameUISnapshot>((resolve) => { releaseRead = resolve; });
      const intent = vi.spyOn(runtime, "intent").mockImplementation((body) => { runtime.requests.push(body); return response; });
      const read = vi.spyOn(runtime, "snapshot").mockReturnValue(refresh);
      try {
        const { userEvent } = await import("vitest/browser");
        await userEvent.click(button(target, title)); await settle();
        const panel = target.querySelector<HTMLElement>(selector)!;
        const region = panel.querySelector<HTMLElement>('[role="status"]')!;
        const assertFeedback = (...keys: CopyKey[]) => {
          expect(panel.querySelectorAll('[role="status"], [aria-live="polite"]')).toHaveLength(1);
          expect(panel.querySelector('[role="status"]')).toBe(region);
          for (const key of keys) {
            const message = careText(key);
            expect(region.textContent).toContain(message);
            expect(region.textContent!.split(message)).toHaveLength(2);
          }
          for (const timer of panel.querySelectorAll('[data-fiscal-region="harvest"]')) expect(region.contains(timer)).toBe(false);
        };
        assertFeedback();
        if (firstState === "stale") {
          runtime.listener?.({ kind: "transport_recovering" }); await settle();
          assertFeedback("common.stale_note");
          expect(runtime.requests).toHaveLength(0);
          runtime.listener?.({ kind: "transport_recovered" }); await settle();
          assertFeedback(); expect(region.textContent).not.toContain(careText("common.stale_note"));
        }
        const action = button(target, actionText); action.focus();
        await userEvent.keyboard("{Enter}"); await settle();
        assertFeedback("common.pending");
        expect(document.activeElement).toBe(action);
        expect(action.getAttribute("aria-disabled")).toBe("true");
        await userEvent.keyboard("{Enter}"); await settle();
        expect(runtime.requests).toHaveLength(1);
        expect(runtime.requests[0]).toMatchObject(owner === "fiscal"
          ? { kind: "harvest_fiscal_period", expected_revision: 7 }
          : { kind: "care_action", pet_id: petRow.pet_id, action_id: "care.feed", expected_revision: 7 });
        finish({ outcome: "applied", receipt: { harvest_outcome: "guaranteed" } }); await settle();
        expect(read).toHaveBeenCalledExactlyOnceWith();
        assertFeedback("common.pending", noticeKey);
        releaseRead({ ...runtime.current, founder_revision: 8 }); await settle();
        assertFeedback(noticeKey);
        expect(region.textContent).not.toContain(careText("common.pending"));
        expect(document.activeElement).toBe(action);
        expect(action.getAttribute("aria-disabled")).not.toBe("true");
        expect(target.querySelector(".announcement")?.textContent).not.toContain(careText(noticeKey));
        await assertAxe(target, `single surface feedback ${owner}/${firstState}`);
      } finally {
        finish({ outcome: "applied", receipt: { harvest_outcome: "guaranteed" } });
        releaseRead(runtime.current); await settle();
        intent.mockRestore(); read.mockRestore(); await dispose();
      }
    });
  }
}

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
            const regions = [...target.querySelectorAll<HTMLElement>("[role=status]")].filter((node) =>
              (node.querySelector(".intent-notice") ?? node).textContent?.trim() === message);
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
