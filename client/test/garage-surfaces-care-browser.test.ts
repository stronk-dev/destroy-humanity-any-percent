import { flushSync, mount, unmount } from "svelte";
import { expect, it, vi } from "vitest";
import { t } from "../src/copy";
import { parseGameUISnapshot, type ParsedGameUISnapshot } from "../src/game-ui/contracts";
import { decodeGameUIAnnouncement } from "../src/game-ui/events";
import PetCareSurface from "../src/game-ui/pet/PetCareSurface.svelte";
import { GameUIRequestError, type IntentOutcome } from "../src/game-ui/intent-outcome";
import { type GameUIRuntimeMessage } from "../src/game-ui/runtime";
import { decodeTransportEnvelope } from "../src/transport";
import { Runtime, mounted, settle, v4 } from "./garage-surfaces-fixture";
import { browser, assertAxe, button, sharedStateVisibleText, petRow, withPet, careText } from "./garage-surfaces-assertions";

// Same Garage cases, grouped by outcome so one serial file cannot dominate CI.


function petStatusAnnouncement(cursor: number, from: string, to: string, petID = petRow.pet_id): GameUIRuntimeMessage {
  const envelope = decodeTransportEnvelope({ v: 2, ch: `player:${v4.run.founder_id}`, kind: "event", rev: cursor,
    constants_hash: v4.constants_hash, ts: "2026-08-11T12:00:00Z",
    payload: { event_id: `pet-status-${cursor}`, kind: "pet_status_changed.v1", scope: "founder", rev: cursor,
      cursor_effect: "advance", payload: { pet_id: petID, from_status_band: from, to_status_band: to } } });
  if (!envelope) throw new Error("pet status transport fixture not admitted");
  const value = decodeGameUIAnnouncement(envelope);
  if (!value || value.kind !== "pet_status_changed") throw new Error("pet status announcement not admitted");
  return { kind: "announcement", scope: "founder", eventID: `pet-status-${cursor}`, value };
}

for (const width of [320, 1280] as const) {
  it.skipIf(!browser)(`RP-419 distinct same-kind events at one revision retain individual replay identity/${width}`, async () => {
    const { page } = await import("vitest/browser");
    await page.viewport(width, 720);
    const { target, runtime, dispose } = await mounted();
    try {
      const message = (eventID: string, achievementID: string): GameUIRuntimeMessage => ({
        kind: "announcement", scope: "company", eventID,
        value: { cursor: 9, kind: "achievement_earned", payload: { achievement_id: achievementID, condition_scope: "run",
          run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 }, score_grant: 2 } },
      });
      const first = message("earned-first", "achievement.first_gate"), second = message("earned-second", "achievement.old_hand");
      runtime.listener?.(first); await settle();
      expect(target.querySelector(".announcement")?.textContent).toBe(t("achievements.earned_announcement", { achievement: t("achievement.first_gate", {}, "era_1995") }, "era_1995"));
      runtime.listener?.(second); await settle();
      const expected = t("achievements.earned_announcement", { achievement: t("achievement.old_hand", {}, "era_1995") }, "era_1995");
      expect(target.querySelector(".announcement")?.textContent).toBe(expected);
      runtime.listener?.(first); await settle();
      expect(target.querySelector(".announcement")?.textContent).toBe(expected);
    } finally { await dispose(); }
  });

  it.skipIf(!browser)(`RP-318 pet status mounted-only, replay and snapshot authority/${width}`, async () => {
    const { page, userEvent } = await import("vitest/browser");
    await page.viewport(width, 720);
    const runtime = new Runtime(); runtime.current = parseGameUISnapshot(withPet());
    const { target, app, dispose } = await mounted(runtime);
    try {
      const region = target.querySelector<HTMLElement>('.announcement[role="status"]')!;
      const normal = t("pet.care.band.normal", {}, "era_1995");
      const expected = (band: "high" | "low" | "floor") => `PENDING OWNER COPY: Your companion's status is now ${t(`pet.care.band.${band}`, {}, "era_1995")}.`;
      const offSurface = petStatusAnnouncement(8, "normal", "floor");
      runtime.listener?.(offSurface); await settle();
      expect(region.textContent).toBe("");
      const nav = button(target, t("pet.care.panel.title", {}, "era_1995"));
      nav.focus(); await userEvent.keyboard("{Enter}"); await settle();
      expect(document.activeElement).toBe(nav);
      expect(region.textContent).toBe("");
      runtime.listener?.(offSurface); await settle();
      expect(region.textContent).toBe("");
      const high = petStatusAnnouncement(9, "normal", "high");
      // Real Clout/care reproduction: automatic Fiscal harvest and pet status
      // are different events in the same committed Founder revision.
      runtime.listener?.({ kind: "announcement", scope: "founder", value: {
        cursor: 9, kind: "fiscal_period_harvested", payload: { source: "automatic", credit_after: 3 },
      } }); await settle();
      runtime.listener?.(high); await settle();
      sharedStateVisibleText(target, '.announcement[role="status"]', expected("high"));
      expect(target.querySelector(".pet-care .pet")?.textContent).toContain(normal);
      expect(document.activeElement).toBe(nav);
      runtime.listener?.(petStatusAnnouncement(10, "high", "low")); await settle();
      sharedStateVisibleText(target, '.announcement[role="status"]', expected("low"));
      // A distinct intervening sentence makes replay detection non-vacuous.
      runtime.listener?.(high); await settle();
      expect(region.textContent).toBe(expected("low"));
      const next = withPet();
      app.fixtureSnapshot(parseGameUISnapshot({ ...next, founder_revision: 8, features: { ...next.features,
        pet_adoption: { ...next.features.pet_adoption!, pets: [{ ...petRow, status_band: "high" }] } } }));
      await settle();
      expect(target.querySelector(".pet-care .pet")?.textContent).toContain(t("pet.care.band.high", {}, "era_1995"));
      expect(target.querySelector(".pet-care .pet")?.textContent).not.toContain(normal);
      expect(region.textContent).toBe(expected("low"));
      await userEvent.click(button(target, t("surface.desk.title", {}, "era_1995"))); await settle();
      expect(target.querySelector(".pet-care")).toBeNull();
      const later = petStatusAnnouncement(11, "low", "floor");
      runtime.listener?.(later); await settle();
      expect(region.textContent).toBe(expected("low"));
      await userEvent.click(nav); await settle();
      runtime.listener?.(later); await settle();
      expect(region.textContent).toBe(expected("low"));
      expect(runtime.requests).toEqual([]);
      expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(width);
      await assertAxe(target, "pet status feedback");
    } finally { await dispose(); }
  });

  it.skipIf(!browser)(`RP-318 pet status withholds wrong scope and absent pet/${width}`, async () => {
    const { page, userEvent } = await import("vitest/browser");
    await page.viewport(width, 720);
    const runtime = new Runtime(); runtime.current = parseGameUISnapshot(withPet());
    const diagnostic = vi.spyOn(console, "error").mockImplementation(() => {});
    const { target, dispose } = await mounted(runtime);
    try {
      await userEvent.click(button(target, t("pet.care.panel.title", {}, "era_1995"))); await settle();
      const correct = petStatusAnnouncement(9, "normal", "high");
      runtime.listener?.({ ...correct, scope: "company" } as GameUIRuntimeMessage); await settle();
      expect(target.querySelector(".announcement")?.textContent).toBe("");
      runtime.listener?.(petStatusAnnouncement(10, "normal", "high", "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb")); await settle();
      expect(target.querySelector(".announcement")?.textContent).toBe("");
      expect(diagnostic).toHaveBeenCalledExactlyOnceWith("game UI invariant: unannounceable pet status 01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb");
      runtime.listener?.(correct); await settle();
      expect(target.querySelector(".announcement")?.textContent).toBe(`PENDING OWNER COPY: Your companion's status is now ${t("pet.care.band.high", {}, "era_1995")}.`);
      expect(runtime.requests).toEqual([]);
    } finally { await dispose(); diagnostic.mockRestore(); }
  });
}

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
it.skipIf(!browser)("care native input queues behind a different intent without duplicating care", async () => {
  const { userEvent } = await import("vitest/browser");
  const runtime = new Runtime(); runtime.current = withPet();
  const { target, dispose } = await mounted(runtime);
  let finishPurchase!: (value: IntentOutcome) => void;
  let finishCare!: (value: IntentOutcome) => void;
  const purchase = new Promise<IntentOutcome>((resolve) => { finishPurchase = resolve; });
  const care = new Promise<IntentOutcome>((resolve) => { finishCare = resolve; });
  const intent = vi.spyOn(runtime, "intent").mockImplementation((body) => {
    runtime.requests.push(body);
    return body.kind === "buy_generator" ? purchase : care;
  });
  try {
    const buy = target.querySelector<HTMLButtonElement>("section[aria-labelledby='generators-heading'] button")!;
    await userEvent.click(buy); await settle();
    expect(runtime.requests).toHaveLength(1);
    button(target, careText("pet.care.panel.title")).click(); await settle();
    const feed = button(target, careText("pet.care.action.feed.title"));
    expect(feed.getAttribute("aria-disabled"), "another intent is not a pending care command").not.toBe("true");
    feed.focus(); await userEvent.keyboard("{Enter}"); await settle();
    expect(runtime.requests, "care must wait for the first intent and its read").toHaveLength(1);
    runtime.current = { ...withPet(), revision: 2, founder_revision: 8 };
    finishPurchase({ outcome: "applied", receipt: {} }); await settle();
    expect(runtime.requests).toHaveLength(2);
    expect(runtime.requests[1]).toMatchObject({ kind: "care_action", expected_revision: 8,
      pet_id: petRow.pet_id, action_id: "care.feed" });
    expect(feed.getAttribute("aria-disabled")).toBe("true");
    await userEvent.keyboard("{Enter}"); await settle();
    expect(runtime.requests, "same-kind pending care remains single flight").toHaveLength(2);
    runtime.current = { ...runtime.current, founder_revision: 9 };
    finishCare({ outcome: "applied", receipt: {} }); await settle();
    expect(feed.getAttribute("aria-disabled")).not.toBe("true");
    expect(target.querySelector(".pet-care .intent-notice")?.textContent).toBe(careText("pet.care.applied"));
  } finally {
    finishPurchase({ outcome: "applied", receipt: {} }); finishCare({ outcome: "applied", receipt: {} });
    await settle(); await dispose(); intent.mockRestore();
  }
});

const careRefreshCases = (["pointerdown", "click"] as const).flatMap((refreshAt) =>
  (["refreshed", "read-failed", "unmounted"] as const).flatMap((completion) =>
    ([320, 1280] as const).map((width) => ({ refreshAt, completion, width }))));
for (const { refreshAt, completion, width } of careRefreshCases) {
  it.skipIf(!browser)(`care native pointer racing a background refresh/${refreshAt}/${completion}/${width}`, async () => {
    const { page, userEvent } = await import("vitest/browser");
    await page.viewport(width, 720);
    const runtime = new Runtime(); runtime.current = withPet();
    const { target, dispose } = await mounted(runtime);
    let finishRead!: (value: ParsedGameUISnapshot) => void;
    let failRead!: (error: Error) => void;
    const heldRead = new Promise<ParsedGameUISnapshot>((resolve, reject) => { finishRead = resolve; failRead = reject; });
    const read = vi.spyOn(runtime, "snapshot").mockImplementationOnce(() => heldRead)
      .mockImplementation(async () => runtime.current);
    const intent = vi.spyOn(runtime, "intent").mockImplementation(async (body) => {
      runtime.requests.push(body);
      runtime.current = { ...withPet(), founder_revision: 9 };
      return { outcome: "applied", receipt: { intent_id: body.intent_id, founder_revision: 9 } };
    });
    let disposed = false;
    const pointerEvents: { type: string; trusted: boolean }[] = [];
    try {
      button(target, careText("pet.care.panel.title")).click(); await settle();
      const feed = button(target, careText("pet.care.action.feed.title"));
      const initialY = feed.getBoundingClientRect().y;
      // Separate movement during the pointer sequence from a completed click
      // reaching its handler during refresh. Both must preserve native consent.
      for (const type of ["pointerdown", "click"] as const) feed.addEventListener(type, (event) => {
        pointerEvents.push({ type, trusted: event.isTrusted });
        if (type === refreshAt) flushSync(() => runtime.listener?.({ kind: "receipt" }));
      }, { capture: true, once: true });
      await userEvent.click(feed); await settle();
      expect(feed.getBoundingClientRect().y, "busy feedback must not move the care control").toBe(initialY);
      expect(pointerEvents, "refresh must not move the control out of the native pointer sequence").toEqual([
        { type: "pointerdown", trusted: true }, { type: "click", trusted: true },
      ]);
      expect(read).toHaveBeenCalledTimes(1);
      expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
      expect(runtime.requests, "do not submit against the old Founder revision").toHaveLength(0);
      if (completion === "unmounted") { await dispose(); disposed = true; }
      runtime.current = { ...withPet(), founder_revision: 8 };
      if (completion === "read-failed") failRead(new Error("authoritative read failed"));
      else finishRead(runtime.current);
      await settle();
      if (completion === "refreshed") {
        expect(runtime.requests, "the accepted click must not vanish during refresh").toHaveLength(1);
        expect(runtime.requests[0]).toEqual({ intent_id: expect.any(String), expected_revision: 8,
          kind: "care_action", pet_id: petRow.pet_id, action_id: "care.feed" });
        expect(read).toHaveBeenCalledTimes(2);
        expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
        expect(target.querySelector(".pet-care .intent-notice")?.textContent).toBe(careText("pet.care.applied"));
      } else {
        expect(runtime.requests, "failed reads and disposed hosts cannot dispatch queued care").toHaveLength(0);
        if (completion === "read-failed") expect(feed.disabled).toBe(true);
      }
    } finally {
      finishRead(runtime.current); await settle();
      if (!disposed) await dispose();
      intent.mockRestore(); read.mockRestore();
    }
  });
}

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
  const app = mount(PetCareSurface, { target, props: { pets: [petRow], cosmetics: null, era: "era_1995", pending: true, carePending: true, controlsEnabled: true, reducedMotion: true, onCare } });
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
