import { expect, it, vi } from "vitest";
import type { GameUISnapshot } from "../src/api/generated/types";
import { t } from "../src/copy";
import { parseGameUISnapshot } from "../src/game-ui/contracts";
import { decodeGameUIAnnouncement, decodeGameUIEvent } from "../src/game-ui/events";
import { FEATURES_PRESENTATION } from "../src/game-ui/features-presentation";
import { GAME_UI_PRESENTATION } from "../src/game-ui/presentation";
import { type GameUIRuntimeMessage } from "../src/game-ui/runtime";
import { formatAmount } from "../src/ui/amount-format";
import { decodeTransportEnvelope } from "../src/transport";
import { NOW, Runtime, meterRow, mounted, settle, v4 } from "./garage-surfaces-fixture";
import { browser, assertAxe, button, sharedStateVisibleText, semanticMeterRows, semanticMeterSnapshot, settleMeterLayout, assertSemanticMeters } from "./garage-surfaces-assertions";

// Same Garage cases, grouped by outcome so one serial file cannot dominate CI.


it.skipIf(!browser)("shows provisioned counts with their cap reason and owned upgrades as text on the Desk (GS6)", async () => {
  const { target, dispose } = await mounted();
  try {
    expect(target.textContent).toContain("Provisioned: 3");
    const card = [...target.querySelectorAll(".card")].find((node) => node.textContent?.includes("Provisioned: 3"))!;
    expect(card.textContent).toContain(t("generator.beige_tower.provisioned_cap", {}, "era_1995"));
    expect(target.textContent).toContain("Owned");
  } finally { await dispose(); }
});

for (const width of [320, 1280]) {
  it.skipIf(!browser)(`OD-15 explains disabled upgrades from server reasons and removes stale explanations at ${width}px`, async () => {
    const { page, userEvent } = await import("vitest/browser");
    await page.viewport(width, 720);
    const states = [
      { id: "upgrade.beige_tower_cache", reason: "owned", key: "desk.upgrade.owned" },
      { id: "upgrade.continuous_feed_paper", reason: null, key: null },
      { id: "upgrade.hold_music_license", reason: "window", key: "desk.rejection.not_in_window" },
      { id: "upgrade.nephew_business_cards", reason: "unaffordable", key: "desk.rejection.unaffordable" },
      { id: "upgrade.reply_all_macro", reason: "requirement", key: "desk.rejection.requires" },
    ] as const;
    const runtime = new Runtime();
    runtime.current = parseGameUISnapshot({ ...v4, upgrades: states.map((row) => ({
      ...v4.upgrades[0], upgrade_id: row.id, eligible: row.reason === null,
      owned: row.reason === "owned", ineligible_reason: row.reason,
    })) });
    const { target, app, dispose } = await mounted(runtime);
    try {
      const section = target.querySelector<HTMLElement>('section[aria-labelledby="upgrades-heading"]')!;
      const cards = [...section.querySelectorAll<HTMLElement>("article")];
      expect(cards).toHaveLength(states.length);
      const visibleReason = (reason: HTMLElement, expected: string) => {
        expect(sharedStateVisibleText(section, `[id="${reason.id}"]`, expected)).toBe(reason);
      };
      for (const [index, row] of states.entries()) {
        const card = cards[index]!;
        const control = card.querySelector("button")!;
        const reasonID = `upgrade-reason-${row.id}`;
        expect(control.disabled).toBe(row.reason !== null);
        if (row.key) {
          const reason = document.getElementById(reasonID)!;
          expect(card.contains(reason)).toBe(true);
          visibleReason(reason, t(row.key, {}, "era_1995"));
          expect(control.getAttribute("aria-describedby")).toBe(reasonID);
          control.click();
        } else {
          expect(document.getElementById(reasonID)).toBeNull();
          expect(control.hasAttribute("aria-describedby")).toBe(false);
        }
      }
      const ownedReason = document.getElementById(`upgrade-reason-${states[0].id}`)!;
      for (const faded of [ownedReason, section]) {
        const originalOpacity = faded.style.opacity;
        try {
          faded.style.opacity = "0"; await settle();
          expect(getComputedStyle(ownedReason).visibility).toBe("visible");
          if (faded !== ownedReason) expect(getComputedStyle(ownedReason).opacity).toBe("1");
          expect(() => visibleReason(ownedReason, t("desk.upgrade.owned", {}, "era_1995")))
            .toThrow("shared-state ancestor opacity");
          faded.style.opacity = "0.5"; await settle();
          visibleReason(ownedReason, t("desk.upgrade.owned", {}, "era_1995"));
        } finally { faded.style.opacity = originalOpacity; await settle(); }
        visibleReason(ownedReason, t("desk.upgrade.owned", {}, "era_1995"));
      }
      expect(runtime.requests).toEqual([]);
      const main = target.querySelector<HTMLElement>("main.game-ui")!;
      expect(main.scrollWidth).toBeLessThanOrEqual(main.clientWidth + 1);
      await assertAxe(target, "OD-15 upgrade reasons");

      // A newer authoritative snapshot, not local balance arithmetic, enables
      // the formerly locked row and removes its description from the button.
      const current = runtime.current;
      runtime.current = parseGameUISnapshot({ ...current, revision: 2, upgrades: current.upgrades.map((row) => row.upgrade_id === "upgrade.reply_all_macro"
        ? { ...row, eligible: true, ineligible_reason: null } : row) });
      app.fixtureSnapshot(runtime.current); await settle();
      const card = [...section.querySelectorAll<HTMLElement>("article")].at(-1)!;
      const control = card.querySelector("button")!;
      expect(control.disabled).toBe(false);
      expect(document.getElementById("upgrade-reason-upgrade.reply_all_macro")).toBeNull();
      expect(control.hasAttribute("aria-describedby")).toBe(false);
      control.focus(); await userEvent.keyboard("{Enter}"); await settle();
      expect(runtime.requests).toHaveLength(1);
      expect(runtime.requests[0]).toMatchObject({ kind: "buy_upgrade", upgrade_id: "upgrade.reply_all_macro", expected_revision: 2 });
    } finally { await dispose(); }
  });
}

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
function forcedLifecycleMessage(destination: "offer_sheet" | "run_end", cursor: number,
  { source = v4, exitType = "scripted_first" }: Readonly<{ source?: GameUISnapshot; exitType?: string }> = {}): Extract<GameUIRuntimeMessage, { kind: "event" }> {
  const kind = destination === "offer_sheet" ? "exit_offer_spawned" : "run_ended";
  const payout = { clout_reach_note: "clout.reach.preserved", network_slot_unlocks: [], reputation_delta: 2, route_knowledge: 25 };
  const payload = destination === "offer_sheet"
    ? { exit_type: "scripted_first", expires_at_ms: NOW + 60_000, offer_id: "01985555-3333-7333-8333-333333333333", payout_preview: payout }
    : { assisted: { advisor: false, commons: false }, attended_ms: 500, ended_at_ms: NOW + 1_000,
      executed_routes: [], exit_type: exitType, faction: null, founder_id: source.run.founder_id,
      gates_crossed: ["gate.t0_to_t1"], generators_purchased_total: 1, ledger_fact_kinds: [], lifetime_value: "1e3",
      payout, pre_timer: false, rta_ms: 2_000,
      run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: source.run.run_seq },
      started_at_ms: source.run.run_started_at_ms, terminal_seq: 2, tier: source.run.tier };
  const envelope = decodeTransportEnvelope({ v: 2, ch: `player:${source.run.founder_id}`, kind: "event", rev: cursor,
    constants_hash: source.constants_hash, ts: new Date(NOW + 1_000).toISOString(),
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

// Accepted exact-next continuation, separate from transport/authentication proof.
// Runtime double reads hold the actual native activation's completion boundary.
for (const arm of ["healthy", "same-sequence", "skipped-sequence", "other-founder", "rejected-read", "heading-focus", "surviving-focus"] as const) {
  for (const width of [320, 1280] as const) {
    for (const activation of ["{Enter}", " "] as const) {
      it.skipIf(!browser)(`GS0 next-company ${arm}/${width}/${activation === " " ? "Space" : "Enter"}`, async () => {
        const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
        const runtime = new Runtime(); runtime.current = parseGameUISnapshot(structuredClone(v4));
        const subscribed = vi.spyOn(runtime, "subscribe");
        const read = vi.spyOn(runtime, "snapshot");
        let releaseRead!: () => void; const held = new Promise<void>((resolve) => { releaseRead = resolve; });
        let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
        try {
          fixture = await mounted(runtime); const { target } = fixture;
          expect(subscribed.mock.calls.map(([id]) => id)).toEqual([v4.run.founder_id]);
          runtime.listener?.(forcedLifecycleMessage("run_end", 10)); await settle();
          const terminal = target.querySelector<HTMLElement>('[aria-labelledby="run-end-heading"]')!;
          const terminalText = terminal.textContent;
          const continuation = button(target, t("screen.run_end.continue", {}, "era_1995"));
          const wire = structuredClone(v4); wire.revision = 2;
          wire.run.run_seq = arm === "same-sequence" ? 1 : arm === "skipped-sequence" ? 3 : 2;
          if (arm === "other-founder") wire.run.founder_id = "01985555-4444-7444-8444-444444444444";
          wire.run.run_started_at_ms = NOW; wire.run.exit_count = 1;
          const next = parseGameUISnapshot(wire);
          read.mockClear(); read.mockImplementationOnce(async () => {
            await held;
            if (arm === "rejected-read") throw new TypeError("synthetic next-run read rejection");
            return next;
          });
          continuation.focus(); await userEvent.keyboard(activation); await settle();
          expect(read).toHaveBeenCalledExactlyOnceWith(); expect(runtime.requests).toEqual([]);
          expect(continuation.disabled).toBe(false); expect(continuation.getAttribute("aria-disabled")).toBe("true");
          expect(document.activeElement).toBe(continuation); expect(terminal.textContent).toBe(terminalText);
          expect(target.querySelector("main")?.dataset.surface).toBe("run_end");
          await userEvent.keyboard(activation); await settle();
          expect(read).toHaveBeenCalledExactlyOnceWith();
          const settings = button(target, t("surface.settings.title", {}, "era_1995"));
          if (arm === "surviving-focus") {
            settings.focus(); await userEvent.keyboard(activation); await settle();
            expect(target.querySelector("main")?.dataset.surface).toBe("settings"); expect(document.activeElement).toBe(settings);
          }
          releaseRead(); await settle();
          if (arm === "healthy" || arm === "heading-focus" || arm === "surviving-focus") {
            expect(target.querySelector("main")?.dataset.surface).toBe("desk");
            expect(continuation.isConnected).toBe(false); expect(terminal.isConnected).toBe(false);
            const heading = sharedStateVisibleText(target, "h1", t("surface.desk.title", {}, "era_1995"));
            if (arm === "heading-focus") {
              expect({ tag: document.activeElement?.tagName, id: document.activeElement?.id }, "removed continuation exposes new context").toEqual({ tag: "H1", id: "desk-heading" });
              expect(document.activeElement).toBe(heading); expect(heading.getAttribute("tabindex")).toBe("-1");
            }
            if (arm === "surviving-focus") expect(document.activeElement, "completion preserves surviving newer focus").toBe(settings);
          } else {
            expect(target.querySelector("main")?.dataset.surface, "foreign or non-next reads must not bind").toBe("run_end");
            expect(terminal.isConnected).toBe(true); expect(terminal.textContent).toBe(terminalText);
            expect(continuation.isConnected).toBe(true); expect(continuation.disabled).toBe(false);
            expect(continuation.hasAttribute("aria-disabled")).toBe(false); expect(document.activeElement).toBe(continuation);
            sharedStateVisibleText(target, '[role="alert"]', t("settings.save_status.offline", {}, "era_1995"));
          }
          expect(subscribed.mock.calls.map(([id]) => id), "continuation cannot switch Founder subscriptions").toEqual([v4.run.founder_id]);
          expect(runtime.requests).toEqual([]); expect(read).toHaveBeenCalledExactlyOnceWith();
          expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(width + 1);
          await assertAxe(target, `continuation ${arm}/${width}`);
        } finally {
          releaseRead(); await settle();
          try { if (fixture) await fixture.dispose(); }
          finally { read.mockRestore(); subscribed.mockRestore(); await page.viewport(1280, 720); }
        }
      });
    }
  }
}

// Non-first standard terminal, native reachability after refusal, then recovery.
// Decoder fixtures prove rendering, not issuance or natural ending reachability.
for (const exitType of ["acquihire", "acquisition", "collapse", "ipo"] as const) {
  for (const { tier, era } of [{ tier: 0, era: "era_1995" }, { tier: 1, era: "era_2000" }, { tier: 2, era: "era_2010" }] as const) {
    for (const width of [320, 1280] as const) {
      for (const activation of ["{Enter}", " "] as const) {
        it.skipIf(!browser)(`GS0 standard-recovery ${exitType}/tier${tier}/${width}/${activation === " " ? "Space" : "Enter"}`, async () => {
          const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
          const source = structuredClone(v4); source.run = { ...source.run, run_seq: 2, exit_count: 1, tier };
          const runtime = new Runtime(); runtime.current = parseGameUISnapshot(source);
          const subscribed = vi.spyOn(runtime, "subscribe"); const read = vi.spyOn(runtime, "snapshot");
          let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
          let releaseRead = () => {};
          try {
            fixture = await mounted(runtime); const { target } = fixture;
            runtime.listener?.(forcedLifecycleMessage("run_end", 20, { source, exitType })); await settle();
            const heading = sharedStateVisibleText(target, "h1", t("screen.run_end.standard.title", {}, era));
            expect(document.activeElement).toBe(heading); expect(heading.getAttribute("tabindex")).toBe("-1");
            const terminal = target.querySelector<HTMLElement>('[aria-labelledby="run-end-heading"]')!;
            const label = GAME_UI_PRESENTATION.exitTypes.get(exitType); expect(label).toBeDefined();
            expect([...terminal.querySelectorAll("p")].map((p) => p.textContent)).toEqual([
              t("screen.run_end.exit_frame", { exit_type: t(label!.title_key, {}, era), tier }, era),
              t("screen.run_end.attended_frame", { attended: "0:00:00" }, era),
              t("screen.run_end.founder_note", {}, era), t("terms.clout_reach_note.text", {}, era),
              t("terms.reputation_delta.frame", { delta: formatAmount("2e0") }, era),
              t("terms.route_knowledge.frame", { delta: 25 }, era),
            ]);
            expect(terminal.querySelector("h2")?.textContent).toBe(t("screen.run_end.delta_heading", { run_seq: 3 }, era));
            expect(terminal.textContent).not.toContain(t("curriculum.scripted_first_failure.body", {}, era));
            const terminalText = terminal.textContent;
            const continuation = button(target, t("screen.run_end.continue", {}, era));
            const stages = ["same-sequence", "skipped-sequence", "other-founder", "rejected-read", "healthy"] as const;
            read.mockClear(); await assertAxe(target, `standard ${exitType}/${tier}`);
            for (const [index, stage] of stages.entries()) {
              const controls = target.querySelectorAll("button,input,summary,[tabindex='0']").length;
              let tabs = 0;
              const position = () => ({ tag: document.activeElement?.tagName, id: document.activeElement?.id,
                text: document.activeElement instanceof HTMLButtonElement ? document.activeElement.textContent : null,
                documentFocus: document.hasFocus() });
              const trace = [position()];
              while (document.activeElement !== continuation && tabs <= controls) {
                await userEvent.keyboard("{Tab}"); await settle(); tabs += 1; trace.push(position());
              }
              if (document.activeElement !== continuation) {
                throw new Error(`native traversal exhausted ${stage}/${tabs}/${controls + 1}: ${JSON.stringify({ trace })}`);
              }
              expect(document.activeElement, `${stage}: native traversal exhausted ${tabs}/${controls + 1}`).toBe(continuation);
              expect(continuation.disabled).toBe(false);
              const held = new Promise<void>((resolve) => { releaseRead = resolve; });
              const wire = structuredClone(source); wire.revision = 3;
              wire.run = { ...wire.run, tier: 0, exit_count: 2, run_started_at_ms: NOW + 1_000,
                run_seq: stage === "same-sequence" ? 2 : stage === "skipped-sequence" ? 4 : 3 };
              wire.server_now_ms = NOW + 2_000; wire.evaluated_through_ms = wire.server_now_ms;
              if (stage === "other-founder") wire.run.founder_id = "01985555-4444-7444-8444-444444444444";
              const successor = parseGameUISnapshot(wire);
              read.mockImplementationOnce(async () => {
                await held;
                if (stage === "rejected-read") throw new TypeError("synthetic standard continuation read rejection");
                return successor;
              });
              await userEvent.keyboard(activation); await settle();
              expect(read).toHaveBeenCalledTimes(index + 1); expect(continuation.disabled).toBe(false);
              expect(continuation.getAttribute("aria-disabled")).toBe("true"); expect(document.activeElement).toBe(continuation);
              expect(terminal.textContent).toBe(terminalText); expect(runtime.requests).toEqual([]);
              await userEvent.keyboard(activation); await settle(); expect(read).toHaveBeenCalledTimes(index + 1);
              releaseRead(); await settle();
              if (stage !== "healthy") {
                expect(target.querySelector("main")?.dataset.surface, stage).toBe("run_end");
                expect(terminal.isConnected).toBe(true); expect(terminal.textContent).toBe(terminalText);
                expect(continuation.disabled).toBe(false);
                expect(continuation.hasAttribute("aria-disabled")).toBe(false); expect(document.activeElement).toBe(continuation);
                sharedStateVisibleText(target, '[role="alert"]', t("settings.save_status.offline", {}, era));
                if (index === 0) await assertAxe(target, `standard refused ${exitType}/${tier}`);
              } else {
                expect(target.querySelector("main")?.dataset.surface).toBe("desk");
                expect(terminal.isConnected).toBe(false); expect(continuation.isConnected).toBe(false);
                const desk = sharedStateVisibleText(target, "h1", t("surface.desk.title", {}, "era_1995"));
                expect(document.activeElement).toBe(desk); expect(desk.id).toBe("desk-heading");
                expect(target.querySelector('[role="alert"]')).toBeNull();
                await assertAxe(target, `standard recovered ${exitType}/${tier}`);
                // Desk has no failure marker: sample the real save-state reader,
                // not an absent class selector that cannot detect stale offline.
                const settings = button(target, t("surface.settings.title", {}, "era_1995"));
                const remaining = target.querySelectorAll("button,input,summary,[tabindex='0']").length;
                let steps = 0;
                while (document.activeElement !== settings && steps <= remaining) {
                  await userEvent.keyboard("{Tab}"); await settle(); steps += 1;
                }
                expect(document.activeElement, `Settings native traversal ${steps}/${remaining + 1}`).toBe(settings);
                await userEvent.keyboard(activation); await settle();
                expect(target.querySelector("main")?.dataset.surface).toBe("settings");
                const status = target.querySelector<HTMLElement>('[aria-labelledby="settings-heading"] p')!;
                const [prefix, suffix] = t("settings.save_status.saved_frame", { ago: "__AGO__" }, "era_1995").split("__AGO__");
                const statusText = status.textContent!;
                expect(statusText.startsWith(prefix!)).toBe(true); expect(statusText.endsWith(suffix!)).toBe(true);
                const ago = statusText.slice(prefix!.length, suffix!.length ? -suffix!.length : undefined);
                expect(ago).toMatch(/^\d+:[0-5]\d:[0-5]\d$/u);
                sharedStateVisibleText(target, '[aria-labelledby="settings-heading"] p', t("settings.save_status.saved_frame", { ago }, "era_1995"));
                expect(statusText).not.toBe(t("settings.save_status.offline", {}, "era_1995"));
              }
              expect(subscribed.mock.calls.map(([id]) => id)).toEqual([source.run.founder_id]);
              expect(runtime.requests).toEqual([]); expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(width + 1);
            }
            expect(read).toHaveBeenCalledTimes(5);
          } finally {
            releaseRead(); await settle();
            try { if (fixture) await fixture.dispose(); }
            finally { read.mockRestore(); subscribed.mockRestore(); await page.viewport(1280, 720); }
          }
        });
      }
    }
  }
}
