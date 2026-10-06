import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import type { GameUISnapshot } from "../src/api/generated/types";
import { t, type CopyEra, type CopyKey } from "../src/copy";
import GameUIApp from "../src/game-ui/GameUIApp.svelte";
import { createBrowserGameUIRuntime, type RuntimeStorage } from "../src/game-ui/runtime";

// R9 host diagnostic: real runtime/parser/host/DOM, controlled network boundary.
// These are NOT live server, real socket, Postgres or minted-content witnesses.
const browser = typeof document !== "undefined";
const ids = ["reputation.unlock.p05", "reputation.starter.cash_small"] as const;
const rawDetail = "must_not_render_mechanical_details";
const categories = ["not_eligible", "unknown_id", "unaffordable", "invalid", "revision_conflict"] as const;
type Category = typeof categories[number];
function wireSnapshot(tier: number, founderRevision = 7): GameUISnapshot {
  return {
    constants_hash: `sha256:${"a".repeat(64)}`, evaluated_through_ms: 1_800_000_000_000,
    server_now_ms: 1_800_000_000_000, schema_version: 4, revision: 13,
    founder_revision: founderRevision,
    facts: [{ fact_id: "bootstrap.needed", value: false }, { fact_id: "feature.reputation_tree", value: true }],
    generators: [{ generator_id: "generator.beige_tower", max_affordable: 2, next_cost: "1e1", next_cost_resource_id: "company.cash", owned: 1, provision_cap: null, provisioned: 0, rate_contribution: "1e0" }],
    manual_action: { action_id: "manual.click", bucket_cap_milli: 50_000, refill_milli_per_ms: 25, refilled_at_ms: 1_800_000_000_000, tokens_milli: 50_000 },
    progress: [{ current: "5e-1", stage_id: "progress.tier", target: "1e0" }],
    resources: [{ amount: "1e2", cap: { amount: "1e1000", reason_key: "resource.company_cash.cap.phase0" }, rate_per_second: "1e0", resource_id: "company.cash" }],
    run: { category: "any_percent", exit_count: 0, founder_id: "01985555-1111-7111-8111-111111111111", run_seq: 1, run_started_at_ms: 1_799_999_000_000, tier },
    features: {
      achievements: null, active_play: null, fiscal: null, meters: null, minigames: null, pets: null,
      reputation: { available: 4, bonus_factor_next_run: "1e0", bonus_factor_this_run: null, level: 4, per_level_ppm: 10_000, spent: 0, unlock_ppm: 0,
        nodes: ids.map((node_id, index) => ({ node_id, body_key: `reputation_tree.node.${index ? "starter_cash_small" : "unlock_p05"}.body`, title_key: `reputation_tree.node.${index ? "starter_cash_small" : "unlock_p05"}.title`, cost: index + 1, kind: index ? "starter" : "bonus_unlock", requires: [], state: "available" })) },
    },
    transitions: { cross_gate: { eligible: true, gate_id: "gate.t0_to_t1", route_id: null }, wind_down: { eligible: false } },
    upgrades: [{ cost_amount: "2e1", cost_resource_id: "company.cash", eligible: true, owned: false, upgrade_id: "upgrade.beige_tower_cache" }],
  };
}

class MemoryStorage implements RuntimeStorage {
  readonly values = new Map<string, string>();
  getItem(key: string): string | null { return this.values.get(key) ?? null; }
  setItem(key: string, value: string): void { this.values.set(key, value); }
  removeItem(key: string): void { this.values.delete(key); }
}
class ControlledSocket extends EventTarget {
  readonly commands: Record<string, unknown>[] = [];
  closed = false;
  constructor(readonly autoSubscribe = true) { super(); queueMicrotask(() => { if (!this.closed) this.dispatchEvent(new Event("open")); }); }
  send(raw: string): void {
    const command = JSON.parse(raw) as Record<string, unknown>;
    this.commands.push(command);
    if (command.id === 1) queueMicrotask(() => { if (!this.closed) this.reply({ id: 1, connect: { client: "controlled" } }); });
    else if (this.autoSubscribe) queueMicrotask(() => { if (!this.closed) this.acknowledge(command.id as 2 | 3); });
  }
  reply(value: unknown): void { this.dispatchEvent(new MessageEvent("message", { data: JSON.stringify(value) })); }
  acknowledge(id: 2 | 3): void {
    expect(this.closed).toBe(false);
    const command = this.commands.find((value) => value.id === id)!;
    expect(command).toBeTruthy();
    const request = command.subscribe as Record<string, unknown>;
    this.reply({ id, subscribe: { recoverable: true, positioned: true, recovered: request.recover === true, epoch: "controlled", offset: 0, publications: [] } });
  }
  networkClose(code: number): void { this.closed = true; this.dispatchEvent(new CloseEvent("close", { code })); }
  close(): void { this.closed = true; this.dispatchEvent(new CloseEvent("close", { code: 1000 })); }
}
function controlledBoundary(tier: number, windDownEligible = false, options: { holdSubscriptions?: boolean; inactive?: boolean } = {}) {
  const initial = wireSnapshot(tier);
  // Opt-in diagnostic projection, not a production-eligibility proof: the
  // server only offers Wind Down at tier >= 1. Tier 0 true exercises copy era.
  initial.transitions.wind_down.eligible = windDownEligible;
  if (options.inactive) {
    initial.features.reputation = null;
    initial.facts[1]!.value = false;
  }
  const storage = new MemoryStorage();
  storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "controlled-access", refreshToken: "controlled-refresh", accountID: "controlled-account", recoveryCode: "controlled-recovery" }));
  const requests: Record<string, unknown>[] = [];
  const headers: string[] = [];
  const unexpected: string[] = [];
  const sockets: ControlledSocket[] = [];
  const heldIntents: ((response: Response) => void)[] = [];
  const heldSnapshots: ((response: Response) => void)[] = [];
  let snapshotCalls = 0;
  const fetcher: typeof fetch = async (input, init) => {
    const url = typeof input === "string" ? input : input instanceof URL ? input.pathname : input.url;
    headers.push(new Headers(init?.headers).get("Authorization") ?? "");
    if (url === "/api/v1/founder/state") {
      snapshotCalls += 1;
      if (snapshotCalls === 1) return Response.json(initial);
      return new Promise<Response>((resolve) => { heldSnapshots.push(resolve); });
    }
    if (url === "/api/v1/garden/current") return Response.json({ kind: "inactive" });
    if (url === "/api/v1/intents" && init?.method === "POST" && typeof init.body === "string") {
      requests.push(JSON.parse(init.body) as Record<string, unknown>);
      return new Promise<Response>((resolve) => { heldIntents.push(resolve); });
    }
    unexpected.push(url); throw new Error(`unexpected controlled request: ${url}`);
  };
  const runtime = createBrowserGameUIRuntime(storage, fetcher, crypto, (url) => {
    expect(url).toBe("ws://controlled.invalid/connection/websocket");
    const socket = new ControlledSocket(!options.holdSubscriptions); sockets.push(socket); return socket as unknown as WebSocket;
  }, { protocol: "http:", host: "controlled.invalid" });
  return {
    runtime, storage, requests, sockets, headers, unexpected,
    get snapshotCalls() { return snapshotCalls; },
    deliverIntent(value: unknown, status = 200) { expect(heldIntents).toHaveLength(1); heldIntents.shift()!(Response.json(value, { status })); },
    deliverSnapshot(value: unknown, status = 200) { expect(heldSnapshots).toHaveLength(1); heldSnapshots.shift()!(Response.json(value, { status })); },
    publishDrain() {
      expect(sockets).toHaveLength(1);
      sockets[0]!.reply({ push: { channel: "world", pub: { offset: 0, data: {
        v: 2, ch: "world", kind: "system", rev: 0, constants_hash: initial.constants_hash,
        ts: new Date(initial.server_now_ms).toISOString(), payload: { code: "server_restarting", resume_after_ms: 0 },
      } } } });
    },
    publishOffer() {
      expect(sockets).toHaveLength(1);
      const channel = `player:${initial.run.founder_id}`;
      sockets[0]!.dispatchEvent(new MessageEvent("message", { data: JSON.stringify({ push: { channel, pub: { offset: 1, data: {
        v: 2, ch: channel, kind: "event", rev: 14, constants_hash: initial.constants_hash,
        ts: new Date(initial.server_now_ms).toISOString(),
        payload: { event_id: "controlled-offer-14", kind: "exit_offer_spawned", scope: "company", rev: 14, cursor_effect: "advance",
          payload: { exit_type: "collapse", expires_at_ms: initial.server_now_ms + 60_000,
            offer_id: "01985555-3333-7333-8333-333333333333",
            payout_preview: { clout_reach_note: "clout.reach.preserved", network_slot_unlocks: [], reputation_delta: 2, route_knowledge: 25 } } },
      } } } }) }));
      sockets[0]!.dispatchEvent(new MessageEvent("message", { data: JSON.stringify({ push: { channel, pub: { offset: 2, data: {
        v: 2, ch: channel, kind: "receipt", rev: 14, constants_hash: initial.constants_hash,
        ts: new Date(initial.server_now_ms).toISOString(),
        payload: { outcome: "applied", new_revision: 14, intent_id: "01985555-4444-7444-8444-444444444444" },
      } } } }) }));
    },
    cleanup() {
      for (const finish of heldIntents.splice(0)) finish(Response.json(rejection(requests.at(-1)?.intent_id, "invalid", rawDetail, Number(requests.at(-1)?.expected_revision ?? 7))));
      for (const finish of heldSnapshots.splice(0)) finish(Response.json(initial));
    },
  };
}

// R6/R9 consistency diagnostic, not a new plan-persistence policy: the
// existing remounted panel starts empty, so its outgoing Exit must be empty.
for (const [tier, era] of [[0, "era_1995"], [1, "era_2000"]] as const) {
  for (const key of ["{Enter}", " "]) {
    for (const flow of ["desk-empty", "desk-selected", "desk-remount", "offer-replaced", "offer-selected"] as const) {
      it.skipIf(!browser)(`R9 host plan matches visible selection: ${era}, ${key === " " ? "Space" : "Enter"}, ${flow}`, async () => {
        const { userEvent } = await import("vitest/browser");
        const boundary = controlledBoundary(tier, true);
        const target = document.createElement("main"); document.body.append(target);
        const app = mount(GameUIApp, { target, props: { runtime: boundary.runtime, timingStorage: boundary.storage } });
        const namedButton = (selector: string, text: string): HTMLButtonElement => {
          const button = [...target.querySelectorAll<HTMLButtonElement>(selector)].find((row) => row.textContent === text);
          expect(button, `native control ${text} is mounted`).toBeTruthy(); return button!;
        };
        const boxes = () => [...target.querySelectorAll<HTMLInputElement>(".plan input[type=checkbox]")];
        const planState = (checked: boolean, projected: number) => {
          expect(boxes()).toHaveLength(2);
          expect(boxes().map((box) => box.checked)).toEqual([checked, checked]);
          expect(target.querySelector(".plan [role=status]")?.textContent).toBe(t("reputation_tree.plan.projected", { amount: projected }, era));
        };
        const openPlan = async () => {
          const details = target.querySelector<HTMLDetailsElement>("details.plan")!;
          expect(details).toBeTruthy(); expect(details.open).toBe(false);
          const summary = details.querySelector<HTMLElement>("summary")!;
          summary.focus(); await userEvent.keyboard(key); await settle(); expect(details.open).toBe(true);
        };
        const selectBoth = async () => {
          expect(boxes().map((box) => box.disabled)).toEqual([false, false]);
          // Select in reverse order; the emitted plan must still use tree order.
          boxes()[1]!.focus(); await userEvent.keyboard(" "); await settle();
          boxes()[0]!.focus(); await userEvent.keyboard(" "); await settle();
          expect(boundary.requests).toEqual([]);
        };
        try {
          await expect.poll(() => boundary.sockets[0]?.commands.length).toBe(3);
          await expect.poll(() => target.querySelector("details.plan")).toBeTruthy();
          await settle(); planState(false, 4);
          await openPlan();
          if (flow !== "desk-empty" && flow !== "offer-selected") { await selectBoth(); planState(true, 1); }
          if (flow === "desk-remount") {
            const settings = namedButton("nav button", t("surface.settings.title", {}, era));
            settings.focus(); await userEvent.keyboard(key); await settle();
            expect(target.querySelector("details.plan")).toBeNull();
            const desk = namedButton("nav button", t("surface.desk.title", {}, era));
            desk.focus(); await userEvent.keyboard(key); await settle();
            await openPlan(); planState(false, 4);
          }
          const offerFlow = flow.startsWith("offer-");
          if (offerFlow) {
            boundary.publishOffer();
            await expect.poll(() => boundary.snapshotCalls).toBe(2);
            const next = wireSnapshot(tier);
            next.revision = 14; next.transitions.wind_down.eligible = true;
            boundary.deliverSnapshot(next);
            await expect.poll(() => target.querySelector("#offer-heading")).toBeTruthy();
            await settle(); await openPlan(); planState(false, 6);
            if (flow === "offer-selected") { await selectBoth(); planState(true, 3); }
          }
          const selected = flow === "desk-selected" || flow === "offer-selected";
          expect(boundary.requests).toEqual([]);
          const exit = namedButton("button", t(offerFlow ? "screen.offer_sheet.accept" : "desk.wind_down", {}, era));
          expect(exit.disabled).toBe(false); exit.focus(); await userEvent.keyboard(key); await settle();
          expect(boundary.requests).toHaveLength(1);
          const request = boundary.requests[0]!;
          expect(request).toMatchObject({ kind: offerFlow ? "accept_exit_offer" : "wind_down", expected_revision: offerFlow ? 14 : 13, expected_founder_revision: 7 });
          expect(request.intent_id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
          if (offerFlow) expect(request.offer_id).toBe("01985555-3333-7333-8333-333333333333");
          if (selected) expect(request.reputation_plan).toEqual(ids);
          else expect(Object.hasOwn(request, "reputation_plan"), "empty visible plan must not submit hidden spending").toBe(false);
          expect(exit.disabled).toBe(true);
          boundary.deliverIntent(rejection(request.intent_id, "invalid", rawDetail, offerFlow ? 14 : 13));
          await expect.poll(() => exit.disabled).toBe(false);
          expect(boundary.snapshotCalls).toBe(offerFlow ? 2 : 1); expect(boundary.unexpected).toEqual([]);
          expect(boundary.headers.every((header) => header === "Bearer controlled-access")).toBe(true);
          expect(target.textContent).not.toContain(rawDetail);
          for (const id of ids) expect(target.textContent).not.toContain(id);
        } finally { boundary.cleanup(); await settle(); await unmount(app); target.remove(); }
      });
    }
  }
}
async function settle(): Promise<void> { for (let index = 0; index < 4; index += 1) { await tick(); flushSync(); } }
async function enterTree(target: HTMLElement, era: CopyEra, boundary: ReturnType<typeof controlledBoundary>): Promise<void> {
  await expect.poll(() => [...target.querySelectorAll<HTMLButtonElement>("nav button")].find((button) => button.textContent === t("reputation_tree.title", {}, era)), { timeout: 5000 }).toBeTruthy();
  await expect.poll(() => boundary.sockets[0]?.commands.length).toBe(3);
  const { userEvent } = await import("vitest/browser");
  const tab = [...target.querySelectorAll<HTMLButtonElement>("nav button")].find((button) => button.textContent === t("reputation_tree.title", {}, era))!;
  tab.focus(); await userEvent.keyboard("{Enter}"); await settle();
  expect(target.querySelectorAll(".reputation li")).toHaveLength(2);
  expect(boundary.snapshotCalls).toBe(1);
}
async function submit(row: HTMLLIElement, key: string): Promise<void> {
  const { userEvent } = await import("vitest/browser");
  row.querySelector<HTMLButtonElement>("button")!.focus();
  await userEvent.keyboard(key); await settle();
  expect(document.activeElement).toBe(row.querySelector("button"));
  await userEvent.keyboard(key); await settle();
}
function rejection(intentID: unknown, category: Category, detail = rawDetail, currentRevision = 7) {
  return { current_revision: currentRevision, intent_id: intentID, outcome: "rejected", rejection: { category, detail } };
}
function errorText(row: Element, category: Category, era: CopyEra): void {
  const status = row.querySelector("[role=status]");
  expect(status, "rejection must be inline on the submitted row").not.toBeNull();
  expect(status!.getAttribute("aria-live")).toBe("polite");
  expect(status!.textContent).toBe(t(`reputation_tree.error.${category}` as CopyKey, {}, era));
}
function silent(row: Element): void { expect(row.querySelector("[role=status]")?.textContent ?? "").toBe(""); }
function controls(target: HTMLElement, disabled: boolean): void {
  const buttons = [...target.querySelectorAll<HTMLButtonElement>(".reputation li button")];
  expect(buttons).toHaveLength(2);
  expect(buttons.every((button) => button.disabled === disabled)).toBe(true);
}

for (const [tier, era] of [[0, "era_1995"], [1, "era_2000"]] as const) {
  for (const key of ["{Enter}", " "]) {
    for (const category of categories) {
      for (const status of category === "revision_conflict" ? [200, 409] : [200]) {
        it.skipIf(!browser)(`R9 host row rejection ${category} HTTP${status}, ${era}, ${key === " " ? "Space" : "Enter"}`, async () => {
          const boundary = controlledBoundary(tier);
          const target = document.createElement("main"); document.body.append(target);
          const app = mount(GameUIApp, { target, props: { runtime: boundary.runtime, timingStorage: boundary.storage } });
          try {
            await enterTree(target, era, boundary);
            const rows = [...target.querySelectorAll<HTMLLIElement>(".reputation li")];
            await submit(rows[0]!, key);
            expect(boundary.requests).toHaveLength(1);
            expect(boundary.requests[0]).toMatchObject({ kind: "purchase_reputation_node", node_id: ids[0], expected_revision: 7 });
            expect(boundary.requests[0]!.intent_id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
            expect(rows.map((row) => row.dataset.state)).toEqual(["available", "available"]);
            expect(rows.filter((row) => row.getAttribute("aria-busy") === "true")).toEqual([rows[0]]);
            controls(target, true);
            const detail = category === "unaffordable" ? "reputation" : rawDetail;
            boundary.deliverIntent(status === 409 ? { category, detail } : rejection(boundary.requests[0]!.intent_id, category, detail, category === "revision_conflict" ? 8 : 7), status);
            if (category === "revision_conflict") {
              await expect.poll(() => boundary.snapshotCalls).toBe(2);
              await settle(); controls(target, true);
              boundary.deliverSnapshot(wireSnapshot(tier, 8));
            }
            await expect.poll(() => rows[0]!.hasAttribute("aria-busy")).toBe(false);
            await settle(); controls(target, false);
            expect(boundary.snapshotCalls).toBe(category === "revision_conflict" ? 2 : 1);
            expect(document.activeElement).toBe(rows[0]);
            errorText(rows[0]!, category, era); silent(rows[1]!);
            expect(target.querySelector(".intent-notice")!.textContent, "known inline rejection must not be spoken twice").toBe("");
            expect(target.textContent).not.toContain(rawDetail);
            for (const id of ids) expect(target.textContent).not.toContain(id);
            // A second native submission must clear the first row, use the new
            // authoritative Founder revision, and attribute its own rejection.
            await submit(rows[1]!, key);
            silent(rows[0]!); silent(rows[1]!);
            expect(boundary.requests).toHaveLength(2);
            expect(boundary.requests[1]).toMatchObject({ kind: "purchase_reputation_node", node_id: ids[1], expected_revision: category === "revision_conflict" ? 8 : 7 });
            expect(boundary.requests[1]!.intent_id).not.toBe(boundary.requests[0]!.intent_id);
            boundary.deliverIntent(rejection(boundary.requests[1]!.intent_id, "invalid", rawDetail, category === "revision_conflict" ? 8 : 7));
            await expect.poll(() => rows[1]!.hasAttribute("aria-busy")).toBe(false);
            await settle(); errorText(rows[1]!, "invalid", era); silent(rows[0]!);
            expect(document.activeElement).toBe(rows[1]);
            expect(boundary.headers.every((header) => header === "Bearer controlled-access")).toBe(true);
            expect(boundary.unexpected).toEqual([]);
          } finally { boundary.cleanup(); await settle(); await unmount(app); target.remove(); }
        });
      }
    }
    it.skipIf(!browser)(`R9 host applied refresh remains authoritative, ${era}, ${key === " " ? "Space" : "Enter"}`, async () => {
      const boundary = controlledBoundary(tier);
      const target = document.createElement("main"); document.body.append(target);
      const app = mount(GameUIApp, { target, props: { runtime: boundary.runtime, timingStorage: boundary.storage } });
      try {
        await enterTree(target, era, boundary);
        const row = target.querySelector<HTMLLIElement>(".reputation li")!;
        await submit(row, key);
        boundary.deliverIntent({ outcome: "applied", intent_id: boundary.requests[0]!.intent_id });
        await expect.poll(() => boundary.snapshotCalls).toBe(2);
        await settle(); controls(target, true); expect(row.dataset.state).toBe("available");
        silent(row); expect(target.querySelector(".intent-notice")!.textContent).toBe(t("reputation_tree.result.applied", {}, era));
        const next = wireSnapshot(tier, 8);
        next.features.reputation!.nodes[0]!.state = "owned";
        next.features.reputation!.spent = 1; next.features.reputation!.available = 3;
        next.features.reputation!.unlock_ppm = 50_000; next.features.reputation!.bonus_factor_next_run = "1.002e0";
        boundary.deliverSnapshot(next);
        await expect.poll(() => row.dataset.state).toBe("owned");
        await settle(); expect(row.hasAttribute("aria-busy")).toBe(false);
        expect(document.activeElement).toBe(row); silent(row);
        expect(boundary.snapshotCalls).toBe(2); expect(boundary.requests).toHaveLength(1);
        expect(boundary.unexpected).toEqual([]);
      } finally { boundary.cleanup(); await settle(); await unmount(app); target.remove(); }
    });
  }
}

// Hold actual runtime subscription replies, not a fabricated host state.
// Snapshot freshness alone must never imply recovered transport (R9).
async function enterHeldTree(target: HTMLElement, era: CopyEra, boundary: ReturnType<typeof controlledBoundary>): Promise<void> {
  await enterTree(target, era, boundary);
  expect(boundary.requests).toEqual([]);
}
function offlineNotice(target: HTMLElement, era: CopyEra, present: boolean): void {
  const notice = [...target.querySelectorAll(".reputation [role=alert]")].find((value) => value.textContent === t("settings.save_status.offline", {}, era));
  if (present) expect(notice, "unrecovered Reputation must explain its disabled controls").toBeTruthy();
  else expect(notice).toBeUndefined();
}
function readyRows(target: HTMLElement): HTMLLIElement[] {
  const rows = [...target.querySelectorAll<HTMLLIElement>(".reputation li")];
  expect(rows.map((row) => row.dataset.state)).toEqual(["available", "available"]);
  expect(rows.every((row) => !row.hasAttribute("aria-busy"))).toBe(true);
  return rows;
}
async function rejectNativePurchase(target: HTMLElement, era: CopyEra, key: string, boundary: ReturnType<typeof controlledBoundary>, revision = 7): Promise<void> {
  const rows = readyRows(target);
  controls(target, false); offlineNotice(target, era, false);
  await submit(rows[0]!, key);
  expect(boundary.requests).toHaveLength(1);
  expect(boundary.requests[0]).toMatchObject({ kind: "purchase_reputation_node", node_id: ids[0], expected_revision: revision });
  controls(target, true);
  boundary.deliverIntent(rejection(boundary.requests[0]!.intent_id, "invalid", rawDetail, revision));
  await expect.poll(() => rows[0]!.hasAttribute("aria-busy")).toBe(false);
  await settle(); controls(target, false); errorText(rows[0]!, "invalid", era);
  expect(document.activeElement).toBe(rows[0]);
  expect(target.textContent).not.toContain(rawDetail);
  for (const id of ids) expect(target.textContent).not.toContain(id);
  expect(boundary.unexpected).toEqual([]);
}

for (const [tier, era] of [[0, "era_1995"], [1, "era_2000"]] as const) {
  for (const key of ["{Enter}", " "]) {
    for (const phase of ["startup", "drain", "overflow", "invalid-frame", "network-drop"] as const) {
      it.skipIf(!browser)(`R9 host waits for both recovered channels: ${era}, ${key === " " ? "Space" : "Enter"}, ${phase}`, async () => {
        const boundary = controlledBoundary(tier, false, { holdSubscriptions: true });
        const target = document.createElement("main"); document.body.append(target);
        const app = mount(GameUIApp, { target, props: { runtime: boundary.runtime, timingStorage: boundary.storage } });
        try {
          await enterHeldTree(target, era, boundary);
          let socket = boundary.sockets[0]!;
          if (phase !== "startup") {
            socket.acknowledge(2); socket.acknowledge(3); await settle(); controls(target, false);
            if (phase === "drain") {
              boundary.publishDrain(); await settle();
              controls(target, true); readyRows(target);
              expect(target.querySelector(".notice")?.textContent).toContain(t("system.drain_notice.title", {}, era));
              socket.networkClose(4003);
            } else socket.networkClose(phase === "overflow" ? 4000 : phase === "invalid-frame" ? 4004 : 1006);
            await settle(); controls(target, true); expect(boundary.requests).toEqual([]);
            if (phase === "overflow" || phase === "invalid-frame") {
              await expect.poll(() => boundary.snapshotCalls).toBe(2);
              const fresh = wireSnapshot(tier, 8);
              boundary.deliverSnapshot(fresh);
              await expect.poll(() => boundary.sockets.length, { timeout: 5000 }).toBe(2);
              // Full sync is complete, but neither live channel has replied.
              await settle(); controls(target, true); readyRows(target);
            } else await expect.poll(() => boundary.sockets.length, { timeout: 5000 }).toBe(2);
            socket = boundary.sockets[1]!;
            await expect.poll(() => socket.commands.length).toBe(3);
            expect(boundary.sockets[0]!.closed).toBe(true);
          }
          await settle(); controls(target, true); readyRows(target); offlineNotice(target, era, true);
          expect(boundary.requests).toEqual([]);
          socket.acknowledge(2); await settle(); controls(target, true); offlineNotice(target, era, true);
          socket.acknowledge(3); await settle(); controls(target, false); offlineNotice(target, era, false);
          const synced = phase === "overflow" || phase === "invalid-frame";
          expect(boundary.snapshotCalls).toBe(synced ? 2 : 1);
          const player = socket.commands[1]!.subscribe as Record<string, unknown>;
          if (phase === "network-drop" || phase === "drain") expect(player).toMatchObject({ recover: true, epoch: "controlled", offset: 0 });
          else expect(Object.hasOwn(player, "recover")).toBe(false);
          await rejectNativePurchase(target, era, key, boundary, synced ? 8 : 7);
          expect(boundary.snapshotCalls).toBe(synced ? 2 : 1);
        } finally { boundary.cleanup(); await settle(); await unmount(app); target.remove(); }
      });
    }

    for (const code of [4001, 4002]) {
      it.skipIf(!browser)(`R9 host explains terminal socket closure without reauthentication: ${era}, ${key === " " ? "Space" : "Enter"}, ${code}`, async () => {
        const boundary = controlledBoundary(tier);
        const target = document.createElement("main"); document.body.append(target);
        const app = mount(GameUIApp, { target, props: { runtime: boundary.runtime, timingStorage: boundary.storage } });
        try {
          await enterTree(target, era, boundary);
          const row = readyRows(target)[0]!;
          row.querySelector<HTMLButtonElement>("button")!.focus();
          const { userEvent } = await import("vitest/browser");
          await userEvent.keyboard(key); await settle();
          const confirm = row.querySelector<HTMLButtonElement>("button")!;
          expect(confirm.textContent).toBe(t("reputation_tree.action.confirm", {}, era));
          boundary.sockets[0]!.networkClose(code); await settle();
          expect(confirm.disabled).toBe(true); expect(readyRows(target)).toContain(row);
          offlineNotice(target, era, true);
          // Native activation of the now disabled confirmation cannot spend.
          confirm.focus(); await userEvent.keyboard(key); await settle();
          expect(boundary.requests).toEqual([]); expect(boundary.snapshotCalls).toBe(1);
          expect(boundary.sockets).toHaveLength(1); expect(boundary.unexpected).toEqual([]);
          expect(boundary.storage.getItem("cloud-clicker.credentials.v1")).toContain("controlled-access");
        } finally { boundary.cleanup(); await settle(); await unmount(app); target.remove(); }
      });
    }

    for (const refresh of ["success", "failure"] as const) {
      it.skipIf(!browser)(`R9 host pending purchase and ${refresh} refresh remain authoritative: ${era}, ${key === " " ? "Space" : "Enter"}`, async () => {
        const boundary = controlledBoundary(tier);
        const target = document.createElement("main"); document.body.append(target);
        const app = mount(GameUIApp, { target, props: { runtime: boundary.runtime, timingStorage: boundary.storage } });
        try {
          await enterTree(target, era, boundary);
          const rows = readyRows(target);
          await submit(rows[0]!, key); controls(target, true);
          expect(rows.map((row) => row.getAttribute("aria-busy"))).toEqual(["true", null]);
          const { userEvent } = await import("vitest/browser");
          rows[1]!.querySelector<HTMLButtonElement>("button")!.focus(); await userEvent.keyboard(key); await settle();
          expect(boundary.requests).toHaveLength(1);
          boundary.deliverIntent({ outcome: "applied", intent_id: boundary.requests[0]!.intent_id });
          await expect.poll(() => boundary.snapshotCalls).toBe(2);
          await settle(); controls(target, true);
          expect(rows.map((row) => row.dataset.state)).toEqual(["available", "available"]);
          expect(target.textContent).toContain(t("reputation_tree.balance.available", { amount: 4 }, era));
          expect(rows[0]!.getAttribute("aria-busy")).toBe("true");
          rows[1]!.querySelector<HTMLButtonElement>("button")!.focus(); await userEvent.keyboard(key); await settle();
          expect(boundary.requests).toHaveLength(1);
          if (refresh === "failure") boundary.deliverSnapshot({ error: "controlled_read_failure" }, 500);
          else {
            const next = wireSnapshot(tier, 8);
            next.features.reputation!.nodes[0]!.state = "owned";
            next.features.reputation!.spent = 1; next.features.reputation!.available = 3;
            next.features.reputation!.unlock_ppm = 50_000; next.features.reputation!.bonus_factor_next_run = "1.002e0";
            boundary.deliverSnapshot(next);
          }
          await expect.poll(() => rows[0]!.hasAttribute("aria-busy")).toBe(false);
          await settle();
          if (refresh === "failure") {
            controls(target, true); readyRows(target); offlineNotice(target, era, true);
            expect(target.textContent).toContain(t("reputation_tree.balance.available", { amount: 4 }, era));
          } else {
            expect(rows[0]!.dataset.state).toBe("owned"); expect(rows[0]!.querySelector("button")).toBeNull();
            expect(rows[1]!.querySelector<HTMLButtonElement>("button")!.disabled).toBe(false);
            offlineNotice(target, era, false);
            expect(target.textContent).toContain(t("reputation_tree.balance.available", { amount: 3 }, era));
            await submit(rows[1]!, key);
            expect(boundary.requests).toHaveLength(2);
            expect(boundary.requests[1]).toMatchObject({ kind: "purchase_reputation_node", node_id: ids[1], expected_revision: 8 });
            boundary.deliverIntent(rejection(boundary.requests[1]!.intent_id, "invalid", rawDetail, 8));
            await expect.poll(() => rows[1]!.hasAttribute("aria-busy")).toBe(false);
            await settle(); errorText(rows[1]!, "invalid", era); expect(document.activeElement).toBe(rows[1]);
          }
          expect(boundary.snapshotCalls).toBe(2); expect(boundary.unexpected).toEqual([]);
          for (const id of ids) expect(target.textContent).not.toContain(id);
        } finally { boundary.cleanup(); await settle(); await unmount(app); target.remove(); }
      });
    }
  }

  for (const inactive of ["initial", "after-refresh"] as const) {
    it.skipIf(!browser)(`R9 host inactive tree is never mounted: ${era}, ${inactive}`, async () => {
      const boundary = controlledBoundary(tier, false, { inactive: inactive === "initial" });
      const target = document.createElement("main"); document.body.append(target);
      const app = mount(GameUIApp, { target, props: { runtime: boundary.runtime, timingStorage: boundary.storage } });
      try {
        await expect.poll(() => boundary.sockets[0]?.commands.length).toBe(3);
        if (inactive === "after-refresh") {
          await enterTree(target, era, boundary);
          await submit(readyRows(target)[0]!, "{Enter}");
          boundary.deliverIntent(rejection(boundary.requests[0]!.intent_id, "revision_conflict", rawDetail, 8));
          await expect.poll(() => boundary.snapshotCalls).toBe(2);
          const next = wireSnapshot(tier, 8); next.features.reputation = null; next.facts[1]!.value = false;
          boundary.deliverSnapshot(next);
        }
        await expect.poll(() => target.querySelector(".game-ui")?.getAttribute("data-surface")).toBe("desk");
        await settle(); expect(target.querySelector(".reputation")).toBeNull();
        expect([...target.querySelectorAll("nav button")].some((button) => button.textContent === t("reputation_tree.title", {}, era))).toBe(false);
        expect(boundary.requests).toHaveLength(inactive === "initial" ? 0 : 1);
        expect(boundary.unexpected).toEqual([]);
      } finally { boundary.cleanup(); await settle(); await unmount(app); target.remove(); }
    });
  }
}
