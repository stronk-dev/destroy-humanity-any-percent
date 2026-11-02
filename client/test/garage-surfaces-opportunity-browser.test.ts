import { mount, unmount } from "svelte";
import { expect, it, vi } from "vitest";
import type { GameUIOpportunityArm } from "../src/api/generated/types";
import { t, type CopyKey } from "../src/copy";
import { isLiveSnapshot, parseGameUISnapshot, type ParsedGameUISnapshot } from "../src/game-ui/contracts";
import { decodeGameUIAnnouncement } from "../src/game-ui/events";
import { FEATURES_PRESENTATION } from "../src/game-ui/features-presentation";
import OpportunityRegion from "../src/game-ui/OpportunityRegion.svelte";
import { GameUIRequestError, type IntentOutcome } from "../src/game-ui/intent-outcome";
import { type GameUIRuntimeMessage } from "../src/game-ui/runtime";
import { formatAmount } from "../src/ui/amount-format";
import { decodeTransportEnvelope } from "../src/transport";
import { Runtime, mounted, opportunityArm, settle, v4, withOpportunity } from "./garage-surfaces-fixture";
import { browser, assertAxe, button, sharedStateVisibleText } from "./garage-surfaces-assertions";

// Same Garage cases, grouped by outcome so one serial file cannot dominate CI.



function regionIndex(target: HTMLElement): number {
  const desk = target.querySelector("section.desk")!;
  return [...desk.children].findIndex((child) => child.getAttribute("data-region") === "desk.region.opportunity");
}

// RP-359: all registered pending presentations, distinct live buff rows and
// an unknown pending row. Public fixtures, not actual issuance or payouts.
function effectPresentationSnapshot(effectID: string | null, revision = 1): ParsedGameUISnapshot {
  const arm: GameUIOpportunityArm = { attended_now_ms: 2_000,
    combo: { cap: "1e4", reason_key: "cap.active_combo", saturated: false },
    buffs: ["active.building", "active.click", "active.production"].map((effect_row_id, index) => ({
      buff_instance_id: `01986666-0000-7000-8000-00000000000${(index + 10).toString(16)}`,
      effect_row_id, expires_attended_ms: 6_000 + index * 1_000,
      selected_target: effect_row_id === "active.building" ? "generator.beige_tower" : null,
    })),
    pending: effectID === null ? null : { effect_row_id: effectID, expires_attended_ms: 5_500,
      opportunity_id: "01986666-0000-7000-8000-000000000001",
      selected_generator_id: effectID === "active.building" ? "generator.beige_tower" : null },
  };
  const base = withOpportunity(opportunityArm(false, false));
  return parseGameUISnapshot({ ...base, revision, features: { ...base.features, opportunity: arm } });
}

for (const effectID of ["active.building", "active.click", "active.lucky", "active.production", "active.unregistered"] as const) {
  for (const width of [320, 1280] as const) {
    it.skipIf(!browser)(`GS5 effect presentation ${effectID}/${width}`, async () => {
      const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
      const runtime = new Runtime(); runtime.current = effectPresentationSnapshot(null);
      runtime.outcome = { outcome: "rejected", category: "not_eligible", detail: "opportunity_expired", currentRevision: 2, sessionExpired: false };
      const diagnostic = vi.spyOn(console, "error").mockImplementation(() => {});
      let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
      try {
        fixture = await mounted(runtime); const { target } = fixture;
        expect([...FEATURES_PRESENTATION.opportunityEffects.keys()]).toEqual(["active.building", "active.click", "active.lucky", "active.production"]);
        const region = target.querySelector<HTMLElement>(".opportunity")!, position = regionIndex(target);
        const manual = target.querySelector<HTMLButtonElement>("section.manual button")!; manual.focus();
        const buffs = () => [...region.querySelectorAll(".buffs li")].map((node) => [...node.querySelectorAll("span")].map((span) => span.textContent));
        const expectedBuffs = ["active.building", "active.click", "active.production"].map((id, index) => [
          t(FEATURES_PRESENTATION.opportunityEffects.get(id)!.title_key, {}, "era_1995"),
          t("desk.buff.remaining_frame", { seconds: 4 + index }, "era_1995"),
        ]);
        expect(buffs()).toEqual(expectedBuffs); expect(target.querySelector(".announcement")?.textContent).toBe("");
        const next = effectPresentationSnapshot(effectID, 2);
        runtime.current = next; runtime.listener?.({ kind: "snapshot", value: next }); await settle();
        expect(regionIndex(target)).toBe(position); expect(document.activeElement).toBe(manual);
        expect(buffs()).toEqual(expectedBuffs);
        sharedStateVisibleText(region, "output", formatAmount("1e4"));
        expect(region.querySelector("output")!.closest("p")?.textContent).toBe(`${t("cap.active_combo", {}, "era_1995")} ${formatAmount("1e4")} `);
        expect(runtime.requests).toEqual([]);
        const row = FEATURES_PRESENTATION.opportunityEffects.get(effectID);
        if (!row) {
          expect(region.querySelectorAll("button,input,select,textarea")).toHaveLength(0);
          expect([...region.querySelectorAll("h3")].map((node) => node.textContent)).toEqual([t("desk.buffs_label", {}, "era_1995")]);
          expect([...region.querySelectorAll("p")].map((node) => node.textContent)).toEqual([`${t("cap.active_combo", {}, "era_1995")} ${formatAmount("1e4")} `]);
          expect(target.querySelector(".announcement")?.textContent).toBe("");
          expect(diagnostic).toHaveBeenCalledExactlyOnceWith("game UI invariant: unknown opportunity effect active.unregistered");
        } else {
          sharedStateVisibleText(region, "h3", t(row.title_key, {}, "era_1995"));
          sharedStateVisibleText(region, "p", t(row.description_key, {}, "era_1995"));
          const remaining = region.querySelectorAll("p")[1]!;
          expect(remaining.childNodes[0]?.textContent).toBe(`${t("desk.opportunity.remaining_frame", { seconds: 4 }, "era_1995")} `);
          expect(remaining.querySelector("small")?.textContent).toBe(t("desk.opportunity.attended_note", {}, "era_1995"));
          expect(remaining.closest('[role="status"],[aria-live]')).toBeNull();
          const curtain = [...region.querySelectorAll("small")].filter((node) => node.textContent === t("desk.opportunity.lucky_tooltip", {}, "era_1995"));
          expect(curtain).toHaveLength(effectID === "active.lucky" ? 1 : 0);
          sharedStateVisibleText(target, ".announcement", t("desk.opportunity.spawned_announcement", { effect: t(row.title_key, {}, "era_1995") }, "era_1995"));
          const claim = button(region, t("desk.opportunity.claim", {}, "era_1995"));
          await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(claim);
          await userEvent.keyboard(width === 320 ? "{Enter}" : " "); await settle();
          expect(runtime.requests).toEqual([{ intent_id: expect.stringMatching(/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u),
            kind: "claim_opportunity", opportunity_id: "01986666-0000-7000-8000-000000000001", expected_revision: 2 }]);
          sharedStateVisibleText(target, ".intent-notice", t("desk.opportunity.rejection.expired", {}, "era_1995"));
          expect(document.activeElement).toBe(claim); expect(claim.disabled).toBe(false);
          expect(diagnostic).not.toHaveBeenCalled(); expect(buffs()).toEqual(expectedBuffs);
        }
        expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(document.documentElement.clientWidth);
        await assertAxe(target, `effect ${effectID}/${width}`);
      } finally { try { if (fixture) await fixture.dispose(); } finally { diagnostic.mockRestore(); await page.viewport(1280, 720); } }
    });
  }
}

// RP-358: decoded publication is announcement authority, never projection
// authority. A different admitted announcement separates replays: otherwise
// assigning the same string can conceal a missing dedupe guard in Svelte.
function opportunityAnnouncement(cursor: number, schema: 1 | 2 | "achievement"): GameUIRuntimeMessage {
  const payload = schema === "achievement"
    ? { achievement_id: "achievement.first_gate", condition_scope: "run", score_grant: 2,
      run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 } }
    : { activated_attended_ms: 2_000, buff_instance_id: "01986666-0000-7000-8000-00000000000b",
      effect_row_id: "active.production", expires_attended_ms: 6_000, selected_target: null,
      ...(schema === 2 ? { hardcap_reason_key: "cap.active_combo" } : {}) };
  const envelope = decodeTransportEnvelope({ v: 2, ch: `player:${v4.run.founder_id}`, kind: "event", rev: cursor,
    constants_hash: v4.constants_hash, ts: "2026-10-07T07:00:00Z",
    payload: { event_id: `opportunity-${cursor}`, kind: schema === "achievement" ? "achievement_earned.v1" : "buff_started.v1",
      scope: "company", rev: cursor, cursor_effect: "advance", payload } });
  if (!envelope) throw new Error("opportunity announcement envelope not admitted");
  const value = decodeGameUIAnnouncement(envelope);
  if (!value || value.kind !== (schema === "achievement" ? "achievement_earned" : "buff_started"))
    throw new Error("opportunity announcement payload not admitted");
  return { kind: "announcement", scope: "company", value };
}

for (const width of [320, 1280] as const) {
  for (const path of ["spawn-desk", "spawn-settings", "buff-v1-desk", "buff-v1-settings", "buff-v2-desk", "buff-v2-settings"] as const) {
    it.skipIf(!browser)(`GS5 announcement authority ${path}/${width}`, async () => {
      const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
      const runtime = new Runtime(); runtime.current = parseGameUISnapshot(withOpportunity(opportunityArm(false, false)));
      let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
      try {
        fixture = await mounted(runtime); const { target } = fixture;
        const chrome = target.querySelector<HTMLElement>(".announcement")!;
        expect(chrome.getAttribute("role")).toBe("status"); expect(chrome.textContent).toBe("");
        const effect = (id: string) => t(FEATURES_PRESENTATION.opportunityEffects.get(id)!.title_key, {}, "era_1995");
        const spawn = t("desk.opportunity.spawned_announcement", { effect: effect("active.lucky") }, "era_1995");
        const buff = t("desk.buff.started_announcement", { effect: effect("active.production") }, "era_1995");
        const earned = t("achievements.earned_announcement", { achievement: t("achievement.first_gate", {}, "era_1995") }, "era_1995");
        expect(new Set([spawn, buff, earned]).size).toBe(3);
        let focused: HTMLElement = target.querySelector<HTMLButtonElement>("section.manual button")!; focused.focus();
        const selected = async (key: "surface.settings.title" | "surface.desk.title") => {
          focused = button(target, t(key, {}, "era_1995")); focused.focus();
          await userEvent.keyboard("{Enter}"); await settle(); expect(document.activeElement).toBe(focused);
        };
        const line = (expected: string) => {
          expect(chrome.textContent).toBe(expected); expect(document.activeElement).toBe(focused);
          expect(runtime.requests).toEqual([]);
          if (expected) { const box = chrome.getBoundingClientRect(); expect(box.width).toBeGreaterThan(0); expect(box.height).toBeGreaterThan(0); expect(getComputedStyle(chrome).visibility).toBe("visible"); }
        };
        const publish = async (cursor: number, schema: 1 | 2 | "achievement", expected: string) => {
          runtime.listener?.(opportunityAnnouncement(cursor, schema)); await settle(); line(expected);
        };
        const snapshot = async (arm: ReturnType<typeof opportunityArm>, revision: number) => {
          const value = parseGameUISnapshot({ ...withOpportunity(arm), revision });
          runtime.current = value; runtime.listener?.({ kind: "snapshot", value }); await settle();
          expect(document.activeElement).toBe(focused); expect(runtime.requests).toEqual([]);
        };
        if (path.endsWith("settings")) await selected("surface.settings.title");
        if (path.startsWith("spawn")) {
          const offer = opportunityArm(true, false);
          await snapshot(offer, 2); line(path.endsWith("settings") ? "" : spawn);
          if (path.endsWith("settings")) { await selected("surface.desk.title"); line(spawn); }
          await publish(11, 1, buff);
          runtime.listener?.({ kind: "transport_recovered" }); await settle();
          await snapshot(offer, 3); line(buff);
          await selected("surface.settings.title"); line(buff);
          await selected("surface.desk.title"); line(buff);
          const next = opportunityArm(true, false); next.pending!.opportunity_id = "01986666-0000-7000-8000-000000000002";
          await snapshot(next, 4); line(spawn);
          await publish(12, "achievement", earned);
          await snapshot(next, 5); line(earned);
          expect(target.querySelector(".opportunity .buffs")).toBeNull();
        } else {
          const schema = path.startsWith("buff-v1") ? 1 : 2;
          await publish(11, schema, path.endsWith("settings") ? "" : buff);
          await publish(12, "achievement", earned);
          if (path.endsWith("settings")) await selected("surface.desk.title");
          line(earned); expect(target.querySelector(".opportunity .buffs")).toBeNull();
          runtime.listener?.({ kind: "transport_recovered" }); await settle();
          await publish(11, schema, earned);
          expect(target.querySelector(".opportunity .buffs")).toBeNull();
          await publish(13, schema, buff);
          expect(target.querySelector(".opportunity .buffs")).toBeNull();
          await snapshot(opportunityArm(false, true), 2); line(buff);
          const remaining = () => [...target.querySelectorAll(".opportunity .buffs li span")].map((node) => node.textContent);
          expect(remaining()).toEqual([effect("active.production"), t("desk.buff.remaining_frame", { seconds: 4 }, "era_1995")]);
          await publish(14, "achievement", earned);
          const advanced = opportunityArm(false, true); advanced.attended_now_ms = 3_000;
          await snapshot(advanced, 3); line(earned);
          expect(remaining()).toEqual([effect("active.production"), t("desk.buff.remaining_frame", { seconds: 3 }, "era_1995")]);
          expect(target.querySelector(".opportunity .buffs")!.closest('[role="status"],[aria-live]')).toBeNull();
        }
        expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(document.documentElement.clientWidth);
        await assertAxe(target, `${path}/${width}`); line(earned);
      } finally { try { if (fixture) await fixture.dispose(); } finally { await page.viewport(1280, 720); } }
    });
  }
}

// GS5 inherits GS0.5 inside the region. Independent property cases keep a
// missing reason from masking readiness; native click probes HTML disabled
// dispatch, not keyboard accessibility or real-service recovery.
for (const width of [320, 1280] as const) {
  for (const [label, message] of [
    ["recovering", { kind: "transport_recovering" }],
    ["closed", { kind: "transport_closed" }],
    ["resync", { kind: "system", value: { kind: "resync_required" } }],
    ["restart", { kind: "system", value: { kind: "server_restarting", resume_after_ms: 1_000 } }],
    ["not-ready-alone", null],
  ] as const) {
    for (const property of ["availability", "disclosure"] as const) {
      it.skipIf(!browser)(`GS5 connection ${label}/${property}/${width}`, async () => {
        const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
        const connectedSnapshot = (next = false) => {
          const source = claimSnapshot(next), arm = source.features.opportunity!;
          return parseGameUISnapshot({ ...source, features: { ...source.features, opportunity: { ...arm,
            attended_now_ms: next ? 3_000 : 2_000,
            pending: { ...arm.pending!, expires_attended_ms: 5_100 },
            buffs: arm.buffs.map((row) => ({ ...row, expires_attended_ms: 5_100 })),
          } } });
        };
        const runtime = new Runtime(); runtime.current = connectedSnapshot();
        runtime.outcome = { outcome: "rejected", category: "not_eligible", detail: "opportunity_expired", currentRevision: 2, sessionExpired: false };
        // Even a new subscription after transport_closed must stay unready
        // until the test explicitly supplies the real recovered message kind.
        const subscribe = vi.spyOn(runtime, "subscribe").mockImplementation((_founder, listener) => { runtime.listener = listener; return () => {}; });
        const read = vi.spyOn(runtime, "snapshot");
        let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
        try {
          fixture = await mounted(runtime); const { target } = fixture;
          const desk = button(target, t("surface.desk.title", {}, "era_1995")); desk.focus();
          const region = target.querySelector<HTMLElement>(".opportunity")!;
          const claim = button(region, t("desk.opportunity.claim", {}, "era_1995"));
          if (message !== null) {
            runtime.listener?.({ kind: "transport_recovered" }); await settle();
            expect(claim.disabled).toBe(false);
            runtime.listener?.(message); await settle();
          }
          const assertReconnect = (next = false) => {
            if (property === "availability") {
              claim.click(); expect(runtime.requests).toEqual([]);
              expect(claim.disabled).toBe(true);
            } else sharedStateVisibleText(region, "p", t("common.stale_note", {}, "era_1995"));
            expect(document.activeElement).toBe(desk);
            sharedStateVisibleText(region, "h3", t(FEATURES_PRESENTATION.opportunityEffects.get("active.lucky")!.title_key, {}, "era_1995"));
            sharedStateVisibleText(region, "p", `${t("desk.opportunity.remaining_frame", { seconds: next ? 3 : 4 }, "era_1995")} ${t("desk.opportunity.attended_note", {}, "era_1995")}`);
            sharedStateVisibleText(region, ".buffs li span", t(FEATURES_PRESENTATION.opportunityEffects.get("active.production")!.title_key, {}, "era_1995"));
            sharedStateVisibleText(region, ".buffs li span", t("desk.buff.remaining_frame", { seconds: next ? 3 : 4 }, "era_1995"));
          };
          assertReconnect(); expect(read).toHaveBeenCalledTimes(1);
          await new Promise((resolve) => setTimeout(resolve, 350)); await settle();
          assertReconnect(); expect(read).toHaveBeenCalledTimes(1);
          runtime.current = connectedSnapshot(true);
          runtime.listener?.({ kind: "snapshot", value: runtime.current }); await settle();
          expect(region.isConnected).toBe(true); expect(button(region, t("desk.opportunity.claim", {}, "era_1995"))).toBe(claim);
          assertReconnect(true); expect(read).toHaveBeenCalledTimes(1);
          runtime.listener?.({ kind: "transport_recovered" }); await settle();
          expect(claim.disabled).toBe(false); expect(document.activeElement).toBe(desk);
          expect(region.textContent).not.toContain(t("common.stale_note", {}, "era_1995"));
          target.querySelector<HTMLButtonElement>("section.manual button")!.focus();
          await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(claim);
          await userEvent.keyboard(width === 320 ? "{Enter}" : " "); await settle();
          expect(runtime.requests).toEqual([{ intent_id: expect.any(String), expected_revision: 2, kind: "claim_opportunity", opportunity_id: "01986666-0000-7000-8000-000000000002" }]);
          expect(read).toHaveBeenCalledTimes(1); expect(document.activeElement).toBe(claim);
          sharedStateVisibleText(target, '.intent-notice[role="status"]', t("desk.opportunity.rejection.expired", {}, "era_1995"));
          expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(width + 1); await assertAxe(target, `Claim connection ${label}/${property}/${width}`);
        } finally {
          try { if (fixture) await fixture.dispose(); }
          finally { subscribe.mockRestore(); read.mockRestore(); await page.viewport(1280, 720); }
        }
      });
    }
  }
}

// GS0.6/0.8/GS5: decoded public-arm fixtures, not live acquisition/payout,
// natural expiry, a new wire contract or assistive-technology evidence.
const claimOutcome: IntentOutcome = { outcome: "applied", receipt: { outcome: "applied", receipt: { opportunity: {
  opportunity_id: "01986666-0000-7000-8000-000000000001", effect_row_id: "active.lucky", selected_target: null,
  buff_instance_id: null, requested_delta: "5e3", actual_credited_delta: "1e2", saturated: true,
  cap_reason_key: "cap.cash", next_sampled_interval_ms: 1_000, next_opportunity_attended_ms: 9_000,
} } } };
function claimSnapshot(next = false, retained = true) {
  const snapshot = structuredClone(withOpportunity(opportunityArm(retained)));
  snapshot.resources[0]!.amount = next ? "1e3" : "9e2"; snapshot.resources[0]!.cap!.amount = "1e3";
  if (next) {
    snapshot.revision = 2; snapshot.founder_revision = 8;
    if (snapshot.features.opportunity?.pending) snapshot.features.opportunity.pending.opportunity_id = "01986666-0000-7000-8000-000000000002";
  }
  const decoded = parseGameUISnapshot(snapshot);
  if (!isLiveSnapshot(decoded)) throw new Error("Claim fixture must use the current public v4 decoder");
  return decoded;
}
for (const { blocker, completion } of (["purchase", "background-read"] as const).flatMap((blocker) =>
  (["refreshed", "read-failed", "unmounted"] as const).map((completion) => ({ blocker, completion })))) {
  it.skipIf(!browser)(`GS5 Claim queues behind ${blocker}/${completion} without stale or duplicate submission`, async () => {
    const { page, userEvent } = await import("vitest/browser"); await page.viewport(320, 720);
    const runtime = new Runtime(); runtime.current = claimSnapshot();
    let finishPurchase!: (value: IntentOutcome) => void, finishClaim!: (value: IntentOutcome) => void;
    let finishRead!: (value: ParsedGameUISnapshot) => void, failRead!: (error: Error) => void;
    const purchase = new Promise<IntentOutcome>((resolve) => { finishPurchase = resolve; });
    const claimReply = new Promise<IntentOutcome>((resolve) => { finishClaim = resolve; });
    const readReply = new Promise<ParsedGameUISnapshot>((resolve, reject) => { finishRead = resolve; failRead = reject; });
    const { target, dispose } = await mounted(runtime);
    const intent = vi.spyOn(runtime, "intent").mockImplementation((request) => {
      runtime.requests.push(request); return request.kind === "buy_generator" ? purchase : claimReply;
    });
    const read = vi.spyOn(runtime, "snapshot").mockReturnValueOnce(readReply);
    let disposed = false;
    const refusal: IntentOutcome = { outcome: "rejected", category: "not_eligible", detail: "opportunity_expired", currentRevision: 2, sessionExpired: false };
    const applied: IntentOutcome = { outcome: "applied", receipt: { new_revision: 2 } };
    try {
      const buy = target.querySelector<HTMLButtonElement>('section[aria-labelledby="generators-heading"] button')!;
      if (blocker === "purchase") {
        await userEvent.click(buy); await settle();
        expect(runtime.requests).toEqual([expect.objectContaining({ kind: "buy_generator", expected_revision: 1 })]);
        await userEvent.keyboard("{Enter}"); await settle();
        expect(runtime.requests, "same-kind purchase remains single flight").toHaveLength(1);
      } else {
        buy.focus(); runtime.listener?.({ kind: "receipt" }); await settle();
        expect(read).toHaveBeenCalledExactlyOnceWith();
      }
      const initialRequests = blocker === "purchase" ? 1 : 0;
      const claim = button(target, t("desk.opportunity.claim", {}, "era_1995"));
      const bound = target.querySelectorAll("button,input,summary,[tabindex='0']").length + 1;
      for (let step = 0; document.activeElement !== claim && step < bound; step++) {
        await userEvent.keyboard("{Tab}"); await settle();
      }
      expect(document.activeElement).toBe(claim);
      await userEvent.keyboard("{Enter}"); await settle();
      expect(runtime.requests, "Claim waits rather than sending against revision1").toHaveLength(initialRequests);
      if (blocker === "purchase") { finishPurchase(applied); await settle(); }
      expect(read).toHaveBeenCalledExactlyOnceWith();
      expect(runtime.requests, "Claim waits for the authoritative read").toHaveLength(initialRequests);
      if (completion === "unmounted") { await dispose(); disposed = true; }
      runtime.current = parseGameUISnapshot({ ...claimSnapshot(), revision: 2 });
      if (completion === "read-failed") failRead(new Error("fixture snapshot failure"));
      else finishRead(runtime.current);
      await settle();
      if (completion === "refreshed") {
        expect(runtime.requests).toEqual([
          ...(blocker === "purchase" ? [expect.objectContaining({ kind: "buy_generator", expected_revision: 1 })] : []),
          { intent_id: expect.any(String), kind: "claim_opportunity", expected_revision: 2,
            opportunity_id: "01986666-0000-7000-8000-000000000001" },
        ]);
        expect(document.activeElement).toBe(claim);
        expect(claim.disabled).toBe(false); expect(claim.getAttribute("aria-disabled")).toBe("true");
        await userEvent.keyboard("{Enter}"); await userEvent.keyboard(" "); await settle();
        expect(runtime.requests, "same-kind Claim repeats stay dropped").toHaveLength(initialRequests + 1);
        finishClaim(refusal); await settle();
        expect(claim.hasAttribute("aria-disabled")).toBe(false);
        sharedStateVisibleText(target, ".intent-notice", t("desk.opportunity.rejection.expired", {}, "era_1995"));
      } else {
        expect(runtime.requests, "failed read or unmount cancels queued Claim").toHaveLength(initialRequests);
        if (completion === "read-failed") expect(claim.disabled).toBe(true);
      }
    } finally {
      finishPurchase(applied); finishRead(runtime.current); finishClaim(refusal); await settle();
      try { if (!disposed) await dispose(); }
      finally { intent.mockRestore(); read.mockRestore(); await page.viewport(1280, 720); }
    }
  });
}
for (const path of ["retained", "removed", "newer-selection"] as const) {
  for (const width of [320, 1280] as const) {
    for (const activation of ["{Enter}", " "] as const) {
      it.skipIf(!browser)(`GS5 held native Claim ${path}/${width}/${activation === " " ? "Space" : "Enter"}`, async () => {
        const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
        const runtime = new Runtime(); runtime.current = claimSnapshot();
        const refusal: IntentOutcome = { outcome: "rejected", category: "not_eligible", detail: "opportunity_expired", currentRevision: 1, sessionExpired: false };
        runtime.outcome = refusal;
        let reject!: (value: IntentOutcome) => void, apply!: (value: IntentOutcome) => void, release!: (value: ParsedGameUISnapshot) => void;
        const heldRefusal = new Promise<IntentOutcome>((resolve) => { reject = resolve; });
        const heldApplied = new Promise<IntentOutcome>((resolve) => { apply = resolve; });
        const heldRead = new Promise<ParsedGameUISnapshot>((resolve) => { release = resolve; });
        let fixture: Awaited<ReturnType<typeof mounted>> | undefined;
        let restore = () => {};
        try {
          fixture = await mounted(runtime); const { target } = fixture;
          const intent = vi.spyOn(runtime, "intent")
            .mockImplementationOnce((request) => { runtime.requests.push(request); return heldRefusal; })
            .mockImplementationOnce((request) => { runtime.requests.push(request); return heldApplied; });
          const read = vi.spyOn(runtime, "snapshot").mockReturnValueOnce(heldRead);
          restore = () => { intent.mockRestore(); read.mockRestore(); };
          const reach = async (destination: HTMLElement, stage: string) => {
            const bound = target.querySelectorAll("button,input,summary,[tabindex='0']").length + 1;
            const trace: string[] = [];
            for (let step = 0; document.activeElement !== destination && step < bound; step++) {
              await userEvent.keyboard("{Tab}"); await settle();
              trace.push(`${document.activeElement?.tagName}:${document.activeElement?.textContent}`);
            }
            expect(document.activeElement, `${stage}: ${JSON.stringify(trace)}`).toBe(destination);
          };
          const request = (revision: number, id: string) => ({ intent_id: expect.any(String), expected_revision: revision, kind: "claim_opportunity", opportunity_id: id });
          const oldID = "01986666-0000-7000-8000-000000000001", newID = "01986666-0000-7000-8000-000000000002";
          const claim = button(target, t("desk.opportunity.claim", {}, "era_1995"));
          const feedback = target.querySelector<HTMLElement>('.intent-notice[role="status"]')!;
          const assertFeedback = (...keys: CopyKey[]) => {
            const ownRegions = [...target.querySelectorAll('[role="status"], [aria-live="polite"]')]
              .filter((node) => !node.classList.contains("announcement"));
            expect(ownRegions, "Desk pending and outcome have one stable owner, separate from chrome").toEqual([feedback]);
            expect(feedback.textContent?.trim()).toBe(keys.map((key) => t(key, {}, "era_1995")).join(" "));
            for (const row of target.querySelectorAll(".opportunity p, .opportunity .buffs")) {
              expect(row.closest('[role="status"], [aria-live]'), "values and attended clocks are not announcements").toBeNull();
            }
          };
          assertFeedback();
          target.querySelector<HTMLButtonElement>("section.manual button")!.focus();
          await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(claim);
          const pending = () => {
            expect(claim.disabled).toBe(false); expect(claim.getAttribute("aria-disabled")).toBe("true");
            expect(document.activeElement).toBe(claim);
            assertFeedback("common.pending");
            const description = sharedStateVisibleText(target, '#desk-pending', t("common.pending", {}, "era_1995"));
            expect(feedback.contains(description)).toBe(true);
            expect(claim.getAttribute("aria-describedby")).toBe(description.id);
            expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
          };
          await userEvent.keyboard(activation); await settle(); pending();
          await userEvent.keyboard("{Enter}"); await userEvent.keyboard(" "); await settle();
          expect(runtime.requests).toEqual([request(1, oldID)]); expect(read).not.toHaveBeenCalled();
          reject(refusal); await settle();
          expect(claim.disabled).toBe(false); expect(claim.hasAttribute("aria-disabled")).toBe(false);
          expect(document.activeElement).toBe(claim); expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
          sharedStateVisibleText(target, '.intent-notice[role="status"]', t("desk.opportunity.rejection.expired", {}, "era_1995"));
          assertFeedback("desk.opportunity.rejection.expired");
          expect(claim.hasAttribute("aria-describedby")).toBe(false);
          expect(target.querySelector(".opportunity")?.textContent).not.toContain(t("common.pending", {}, "era_1995"));
          expect(read).not.toHaveBeenCalled(); await assertAxe(target, `Claim refusal ${path}/${width}`);
          await userEvent.keyboard(activation); await settle(); pending();
          expect(runtime.requests).toEqual([request(1, oldID), request(1, oldID)]);
          assertFeedback("common.pending");
          apply(claimOutcome); await settle();
          expect(claim.disabled).toBe(false); expect(claim.getAttribute("aria-disabled")).toBe("true");
          expect(document.activeElement).toBe(claim); expect(read).toHaveBeenCalledExactlyOnceWith();
          expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
          assertFeedback("common.pending", "desk.opportunity.lucky_capped");
          expect(claim.getAttribute("aria-describedby")).toBe("desk-pending");
          sharedStateVisibleText(target, '#desk-pending', t("common.pending", {}, "era_1995"));
          // Amount's existing inline fragment contributes a trailing text space.
          const creditText = `${t("desk.opportunity.lucky_frame", { amount: "1e2" }, "era_1995")} ${formatAmount("1e2")} `;
          const actualReceiptText = [...target.querySelectorAll(".opportunity p")].map((row) => row.textContent);
          expect(actualReceiptText, `actual complete receipt text: ${JSON.stringify(actualReceiptText)}`).toContain(creditText);
          sharedStateVisibleText(target, ".opportunity p", creditText);
          sharedStateVisibleText(target, ".opportunity p", `${t("desk.opportunity.lucky_capped", {}, "era_1995")} ${t("cap.cash", {}, "era_1995")}`);
          await userEvent.keyboard("{Enter}"); await userEvent.keyboard(" "); await settle();
          expect(runtime.requests).toEqual([request(1, oldID), request(1, oldID)]);
          const settings = button(target, t("surface.settings.title", {}, "era_1995"));
          if (path === "newer-selection") {
            await reach(settings, "newer Settings"); await userEvent.keyboard(activation); await settle();
            expect(document.activeElement).toBe(settings); expect(target.querySelector("main")?.dataset.surface).toBe("settings");
            expect(target.querySelector(".intent-notice")?.textContent).toBe("");
          }
          runtime.current = claimSnapshot(true, path === "retained"); release(runtime.current); await settle();
          expect(read).toHaveBeenCalledExactlyOnceWith(); expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
          if (path === "newer-selection") {
            expect(document.activeElement).toBe(settings); expect(target.querySelector("main")?.dataset.surface).toBe("settings");
            expect(target.querySelector(".intent-notice")?.textContent).toBe("");
            const desk = button(target, t("surface.desk.title", {}, "era_1995"));
            await reach(desk, "return Desk"); await userEvent.keyboard(activation); await settle(); expect(document.activeElement).toBe(desk);
          } else if (path === "removed") {
            expect(claim.isConnected).toBe(false); expect(document.activeElement).toBe(target.querySelector("#desk-heading"));
          } else {
            expect(document.activeElement).toBe(claim); expect(claim.disabled).toBe(false); expect(claim.hasAttribute("aria-disabled")).toBe(false);
            expect(claim.hasAttribute("aria-describedby")).toBe(false);
            runtime.outcome = { ...refusal, currentRevision: 2 };
            await userEvent.keyboard(activation); await settle();
            expect(runtime.requests).toEqual([request(1, oldID), request(1, oldID), request(2, newID)]);
            expect(new Set(runtime.requests.map((row) => row.intent_id)).size).toBe(3); expect(document.activeElement).toBe(claim);
            sharedStateVisibleText(target, '.intent-notice[role="status"]', t("desk.opportunity.rejection.expired", {}, "era_1995"));
          }
          assertFeedback(path === "retained" ? "desk.opportunity.rejection.expired" : "desk.opportunity.lucky_capped");
          expect(target.querySelector(".opportunity button") === null).toBe(path !== "retained");
          expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(width + 1);
          await assertAxe(target, `Claim completion ${path}/${width}`);
        } finally {
          reject(refusal); apply(claimOutcome); release(runtime.current); await settle(); restore();
          try { if (fixture) await fixture.dispose(); } finally { await page.viewport(1280, 720); }
        }
      });
    }
  }
}
for (const width of [320, 1280] as const) {
  for (const activation of ["{Enter}", " "] as const) {
    it.skipIf(!browser)(`GS5 immediate removed Claim focus ${width}/${activation === " " ? "Space" : "Enter"}`, async () => {
      const { page, userEvent } = await import("vitest/browser"); await page.viewport(width, 720);
      const runtime = new Runtime(); runtime.current = claimSnapshot(); runtime.outcome = claimOutcome;
      const { target, dispose } = await mounted(runtime);
      try {
        runtime.current = claimSnapshot(true, false);
        const claim = button(target, t("desk.opportunity.claim", {}, "era_1995"));
        target.querySelector<HTMLButtonElement>("section.manual button")!.focus();
        await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(claim);
        await userEvent.keyboard(activation); await settle();
        expect(runtime.requests).toEqual([{ intent_id: expect.any(String), expected_revision: 1, kind: "claim_opportunity", opportunity_id: "01986666-0000-7000-8000-000000000001" }]);
        expect(claim.isConnected).toBe(false); expect(document.activeElement).toBe(target.querySelector("#desk-heading"));
        expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(width + 1); await assertAxe(target, "immediate Claim removal");
      } finally { try { await dispose(); } finally { await page.viewport(1280, 720); } }
    });
  }
}
for (const state of ["pending", "unavailable"] as const) {
  for (const activation of ["{Enter}", " "] as const) {
    it.skipIf(!browser)(`GS5 isolated Claim callback ${state}/${activation === " " ? "Space" : "Enter"}`, async () => {
      const { userEvent } = await import("vitest/browser"); const onClaim = vi.fn();
      const target = document.createElement("div"); document.body.append(target);
      const app = mount(OpportunityRegion, { target, props: { arm: claimSnapshot().features.opportunity!, era: "era_1995",
        pending: state === "pending", controlsEnabled: state !== "unavailable", lastClaim: undefined, onClaim } });
      try {
        await settle(); const claim = button(target, t("desk.opportunity.claim", {}, "era_1995"));
        expect(claim.disabled).toBe(state === "unavailable");
        if (state === "pending") {
          claim.focus(); expect(document.activeElement).toBe(claim); expect(claim.getAttribute("aria-disabled")).toBe("true");
          // This callback-only fixture has no host; the mounted host population
          // above proves the visible description and the single feedback region.
          expect(claim.getAttribute("aria-describedby")).toBe("desk-pending");
          expect(target.querySelectorAll('[role="status"], [aria-live]')).toHaveLength(0);
        }
        await userEvent.keyboard(activation); await settle(); expect(onClaim).not.toHaveBeenCalled();
      } finally { await unmount(app); target.remove(); }
    });
  }
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
