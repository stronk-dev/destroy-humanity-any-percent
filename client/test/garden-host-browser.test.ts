import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import views from "../../testdata/garden/view-fixtures-v1.json";
import type { GardenCurrentResponse, GameUISnapshot } from "../src/api/generated/types";
import { t } from "../src/copy";
import GameUIApp from "../src/game-ui/GameUIApp.svelte";
import { createBrowserGameUIRuntime, type RuntimeStorage } from "../src/game-ui/runtime";

const browser = typeof document !== "undefined";
type Active = Extract<GardenCurrentResponse, { kind: "active" }>;
const NOW = 1_800_000_000_000;
const founderID = "01985555-1111-7111-8111-111111111111";
const constantsHash = `sha256:${"a".repeat(64)}`;

function snapshot(revision: number): GameUISnapshot {
  return {
    constants_hash: constantsHash, evaluated_through_ms: NOW,
    facts: [{ fact_id: "bootstrap.needed", value: false }],
    founder_revision: revision, generators: [],
    manual_action: { action_id: "manual.click", bucket_cap_milli: 50_000, refill_milli_per_ms: 25, refilled_at_ms: NOW, tokens_milli: 50_000 },
    progress: [], resources: [], revision: 41,
    run: { category: "any_percent", exit_count: 0, founder_id: founderID, run_seq: 1, run_started_at_ms: NOW - 1_000, tier: 2 },
    features: { achievements: null, active_play: null, fiscal: null, meters: null, minigames: null, pets: null },
    schema_version: 4, server_now_ms: NOW,
    transitions: { cross_gate: null, wind_down: { eligible: false } }, upgrades: [],
  };
}

function active(revision = 7): Active {
  const value = structuredClone(views.active) as Active;
  value.founder_revision = revision;
  value.garden.next_tick_wall_ms = null; // Timer behavior has its separate witness.
  return value;
}

// Inject frames at the socket boundary, retaining the actual runtime handshake,
// envelope decoder, cursor/offset logic and host consumePublication path.
class Socket {
  readonly sent: unknown[] = [];
  readonly listeners = new Map<string, ((event: { data?: string }) => void)[]>();
  closed = false;
  addEventListener(kind: string, listener: (event: { data?: string }) => void) {
    this.listeners.set(kind, [...(this.listeners.get(kind) ?? []), listener]);
  }
  send(data: string) { this.sent.push(JSON.parse(data)); }
  close() { this.closed = true; }
  emit(kind: string, event: { data?: string } = {}) { for (const listener of this.listeners.get(kind) ?? []) listener(event); }
  reply(value: unknown) { this.emit("message", { data: JSON.stringify(value) }); }
  connect(recovered = false) {
    this.emit("open"); this.reply({ id: 1, connect: {} });
    for (const id of [2, 3]) this.reply({ id, subscribe: { recoverable: true, positioned: true, recovered, epoch: `epoch-${id}`, offset: 0, publications: [] } });
  }
  resync() {
    this.reply({ push: { channel: "world", pub: { offset: 1, data: {
      v: 2, ch: "world", kind: "system", rev: 0, constants_hash: constantsHash,
      ts: "2026-10-08T02:00:00Z", payload: { code: "resync_required" },
    } } } });
  }
  receipt(intentID?: string) {
    this.reply({ push: { channel: `player:${founderID}`, pub: { offset: 1, data: {
      v: 2, ch: `player:${founderID}`, kind: "receipt", rev: 8, constants_hash: constantsHash,
      ts: "2026-10-06T02:00:00Z", payload: { outcome: "applied", founder_revision: 8, ...(intentID === undefined ? {} : { intent_id: intentID }) },
    } } } });
  }
}

async function settle() {
  for (let index = 0; index < 5; index++) {
    await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync();
  }
}

function button(target: HTMLElement, text: string): HTMLButtonElement {
  const found = [...target.querySelectorAll<HTMLButtonElement>("button")].find((node) => node.textContent?.trim() === text);
  if (!found) throw new Error(`missing DOM control ${text}`);
  return found;
}

async function mounted(initial: GardenCurrentResponse = active(), connected = true, options: { holdInitialGarden?: boolean } = {}) {
  let currentSnapshot = snapshot(7), currentGarden = initial;
  const data = new Map([["cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "garden-host-token", refreshToken: "refresh", accountID: "account", recoveryCode: "recover" })]]);
  const storage: RuntimeStorage = { getItem: (key) => data.get(key) ?? null, setItem: (key, value) => { data.set(key, value); }, removeItem: (key) => { data.delete(key); } };
  const requests: { path: string; method: string; auth: string | null; body: unknown }[] = [];
  let release: ((value: unknown) => void) | undefined;
  let holdGardenRead = options.holdInitialGarden ?? false;
  let releaseGarden: ((value: GardenCurrentResponse) => void) | undefined;
  const sockets: Socket[] = [];
  const urls: string[] = [];
  const runtime = createBrowserGameUIRuntime(storage, async (input, init) => {
    const path = String(input), method = init?.method ?? "GET";
    const body = init?.body === undefined ? undefined : JSON.parse(String(init.body));
    requests.push({ path, method, auth: new Headers(init?.headers).get("Authorization"), body });
    let response: unknown;
    if (path === "/api/v1/founder/state" && method === "GET") response = currentSnapshot;
    else if (path === "/api/v1/garden/current" && method === "GET") {
      if (holdGardenRead) {
        holdGardenRead = false;
        response = await new Promise<GardenCurrentResponse>((resolve) => { releaseGarden = resolve; });
      } else response = currentGarden;
    }
    else if (path === "/api/v1/intents" && method === "POST") response = await new Promise((resolve) => { release = resolve; });
    else throw new Error(`unexpected HTTP ${method} ${path}`);
    return new Response(JSON.stringify(response), { status: 200, headers: { "Content-Type": "application/json" } });
  }, crypto, (url) => { urls.push(url); const socket = new Socket(); sockets.push(socket); return socket as unknown as WebSocket; }, { protocol: "https:", host: "garden.example.invalid" });
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } });
  await settle();
  expect(urls).toEqual(["wss://garden.example.invalid/connection/websocket"]);
  if (connected) {
    sockets[0]!.connect(); await settle();
    expect(sockets[0]!.sent).toEqual([{ id: 1, connect: { token: "garden-host-token" } }, { id: 2, subscribe: { channel: `player:${founderID}` } }, { id: 3, subscribe: { channel: "world" } }]);
  }
  return {
    target, requests, sockets,
    get socket() { return sockets.at(-1)!; },
    set(view: GardenCurrentResponse, revision: number) { currentGarden = view; currentSnapshot = snapshot(revision); },
    context(view: GardenCurrentResponse, value: GameUISnapshot) { currentGarden = view; currentSnapshot = value; },
    holdGarden() { if (releaseGarden) throw new Error("Garden read already held"); holdGardenRead = true; },
    resolveGarden(value: GardenCurrentResponse) { if (!releaseGarden) throw new Error("no held Garden read"); const resolve = releaseGarden; releaseGarden = undefined; resolve(value); },
    acknowledge(value: unknown) { if (!release) throw new Error("no pending intent"); const resolve = release; release = undefined; resolve(value); },
    async dispose() {
      releaseGarden?.({ kind: "inactive" }); releaseGarden = undefined;
      release?.({ outcome: "rejected", intent_id: "cleanup", current_revision: 8, rejection: { category: "not_eligible", detail: "plot_occupied" } });
      await settle(); await unmount(app); target.remove(); expect(sockets.at(-1)!.closed).toBe(true);
    },
    reads() { return requests.filter((row) => row.path === "/api/v1/garden/current").length; },
    intents() { return requests.filter((row) => row.path === "/api/v1/intents").map((row) => row.body as Record<string, unknown>); },
    assertHTTP() {
      for (const request of requests) {
        expect(["/api/v1/founder/state", "/api/v1/garden/current", "/api/v1/intents"]).toContain(request.path);
        expect(request.auth).toBe("Bearer garden-host-token");
        if (request.path !== "/api/v1/intents") { expect(request.method).toBe("GET"); expect(request.body).toBeUndefined(); }
        else expect(request.method).toBe("POST");
      }
    },
  };
}

function cell(target: HTMLElement, row: number, col: number): HTMLButtonElement {
  const found = [...target.querySelectorAll<HTMLButtonElement>(".garden button.cell")][row * 6 + col];
  if (!found) throw new Error("missing Garden cell");
  return found;
}

const cases = [
  { name: "plant", kind: "garden_plant", fields: { row: 1, col: 0, species_id: "strain_a" }, at: [1, 0], action: "Plant Strain A (PENDING OWNER NAME)" },
  { name: "uproot", kind: "garden_uproot", fields: { row: 0, col: 0 }, at: [0, 0], action: "Uproot" },
  { name: "harvest", kind: "garden_harvest", fields: { plots: [{ row: 0, col: 0 }] }, at: [0, 0], action: "Harvest" },
  { name: "harvest-all", kind: "garden_harvest", fields: { plots: [{ row: 0, col: 0 }, { row: 0, col: 1 }, { row: 1, col: 1 }] }, at: null, action: "Harvest all mature" },
  { name: "substrate", kind: "garden_set_substrate", fields: { substrate_id: "mainframe" }, at: null, action: "Mainframe" },
] as const;

for (const newerNavigation of [false, true]) {
  it.skipIf(!browser)(`Garden lifecycle inactive read ${newerNavigation ? "preserves newer navigation" : "returns to Desk"}`, async () => {
    const host = await mounted();
    try {
      button(host.target, "Server Garden").click(); await settle();
      button(host.target, "Harvest all mature").focus();
      host.holdGarden(); host.set({ kind: "inactive" }, 8);
      host.socket.receipt(); await settle();
      const settings = button(host.target, "Settings");
      if (newerNavigation) { settings.focus(); settings.click(); await settle(); }
      host.resolveGarden({ kind: "inactive" }); await settle();
      expect([...host.target.querySelectorAll("nav button")].some((node) => node.textContent?.trim() === "Server Garden")).toBe(false);
      expect(host.target.querySelector(".garden")).toBeNull();
      expect(document.activeElement).toBe(newerNavigation ? settings : host.target.querySelector("#desk-heading"));
      if (newerNavigation) expect(host.target.querySelector("#settings-heading")).not.toBeNull();
      expect(host.intents()).toEqual([]); host.assertHTTP();
    } finally { await host.dispose(); }
  });
}

for (const context of ["founder", "content"] as const) {
  it.skipIf(!browser)(`Garden lifecycle ${context} change replaces the old grid and ignores its late read`, async () => {
    const host = await mounted();
    try {
      host.holdGarden(); button(host.target, "Server Garden").click(); await settle();
      expect(host.target.querySelector(".garden .grid")).toBeNull();
      const next = snapshot(8);
      if (context === "founder") next.run.founder_id = "01985555-2222-7222-8222-222222222222";
      else next.constants_hash = `sha256:${"b".repeat(64)}`;
      const view = active(8); view.garden.substrate_id = "containerized";
      host.context(view, next); host.socket.resync(); await settle();
      await expect.poll(() => host.sockets.length).toBe(2);
      host.socket.connect(); await settle();
      host.resolveGarden(active()); await settle();
      expect(button(host.target, "Containerized").getAttribute("aria-pressed")).toBe("true");
      expect(button(host.target, "Bare metal").getAttribute("aria-pressed")).toBe("false");
      expect(host.intents()).toEqual([]);
      button(host.target, "Mainframe").focus();
      const { userEvent } = await import("vitest/browser");
      await userEvent.keyboard("{Enter}"); await settle();
      expect(host.intents()).toEqual([{ intent_id: expect.any(String), expected_revision: 8,
        kind: "garden_set_substrate", substrate_id: "mainframe" }]);
      host.assertHTTP();
    } finally { await host.dispose(); }
  });
}

it.skipIf(!browser)("Garden lifecycle late previous-Founder startup probe cannot revive the new Founder's inactive tab", async () => {
  const host = await mounted(active(), true, { holdInitialGarden: true });
  try {
    const next = snapshot(8); next.run.founder_id = "01985555-2222-7222-8222-222222222222";
    host.context({ kind: "inactive" }, next); host.socket.resync(); await settle();
    await expect.poll(() => host.sockets.length).toBe(2);
    host.socket.connect(); await settle();
    host.resolveGarden(active()); await settle();
    expect([...host.target.querySelectorAll("nav button")].some((node) => node.textContent?.trim() === "Server Garden")).toBe(false);
    expect(host.reads()).toBe(2);
    expect(host.intents()).toEqual([]); host.assertHTTP();
  } finally { await host.dispose(); }
});

it.skipIf(!browser)("Garden lifecycle Company change retains its Founder-owned grid and open menu", async () => {
  const host = await mounted();
  try {
    button(host.target, "Server Garden").click(); await settle();
    const plot = cell(host.target, 1, 0);
    plot.focus(); plot.click(); await settle();
    const menu = host.target.querySelector(".garden .menu");
    const next = snapshot(8); next.run.run_seq = 2; next.revision = 42;
    host.context(active(8), next); host.socket.resync(); await settle();
    await expect.poll(() => host.sockets.length).toBe(2);
    host.socket.connect(); await settle();
    expect(cell(host.target, 1, 0)).toBe(plot);
    expect(host.target.querySelector(".garden .menu")).toBe(menu);
    expect(host.reads()).toBe(2); // Neither Founder nor catalog changed.
    const { userEvent } = await import("vitest/browser");
    button(host.target, "Plant Strain A (PENDING OWNER NAME)").focus();
    await userEvent.keyboard("{Enter}"); await settle();
    expect(host.intents()).toEqual([{ intent_id: expect.any(String), expected_revision: 8,
      kind: "garden_plant", row: 1, col: 0, species_id: "strain_a" }]);
    host.assertHTTP();
  } finally { await host.dispose(); }
});

it.skipIf(!browser)("Garden lifecycle new content discovers a previously inactive Garden", async () => {
  const host = await mounted({ kind: "inactive" });
  try {
    const next = snapshot(8); next.constants_hash = `sha256:${"b".repeat(64)}`;
    host.context(active(8), next); host.socket.resync(); await settle();
    await expect.poll(() => host.sockets.length).toBe(2);
    host.socket.connect(); await settle();
    button(host.target, "Server Garden").click(); await settle();
    expect(cell(host.target, 0, 0).dataset.stage).toBe("mature");
    expect(host.reads()).toBe(3);
    expect(host.intents()).toEqual([]); host.assertHTTP();
  } finally { await host.dispose(); }
});

it.skipIf(!browser)("Garden lifecycle older same-context probe cannot revive a newer inactive surface", async () => {
  const host = await mounted();
  try {
    button(host.target, "Server Garden").click(); await settle();
    const next = snapshot(8); next.constants_hash = `sha256:${"b".repeat(64)}`;
    host.holdGarden(); host.context(active(8), next); host.socket.resync(); await settle();
    await expect.poll(() => host.sockets.length).toBe(2);
    host.socket.connect(); await settle();
    expect(host.reads()).toBe(4); // Held context probe plus fresh mounted read.
    host.context({ kind: "inactive" }, next); host.socket.receipt(); await settle();
    expect(host.target.querySelector(".garden")).toBeNull();
    host.resolveGarden(active(8)); await settle();
    expect([...host.target.querySelectorAll("nav button")].some((node) => node.textContent?.trim() === "Server Garden")).toBe(false);
    expect(host.intents()).toEqual([]); host.assertHTTP();
  } finally { await host.dispose(); }
});

for (const focus of ["harvest", "navigation"] as const) it.skipIf(!browser)(`Garden lifecycle replaced ${focus} focus stays usable`, async () => {
  const host = await mounted();
  try {
    button(host.target, "Server Garden").click(); await settle();
    const origin = button(host.target, focus === "harvest" ? "Harvest all mature" : "Server Garden"); origin.focus();
    const next = snapshot(8); next.constants_hash = `sha256:${"b".repeat(64)}`;
    host.context(active(8), next); host.socket.resync(); await settle();
    if (focus === "harvest") expect(origin.isConnected).toBe(false);
    expect(document.activeElement).toBe(origin.isConnected ? origin : host.target.querySelector("#garden-heading"));
    expect(host.intents()).toEqual([]); host.assertHTTP();
  } finally { await host.dispose(); }
});

for (const state of ["not-ready", "recovering", "resync"] as const) {
  it.skipIf(!browser)(`Garden connection ${state} visibly disables commands until the real handshake recovers`, async () => {
    const { userEvent } = await import("vitest/browser");
    const host = await mounted(active(), state !== "not-ready");
    try {
      button(host.target, "Server Garden").click(); await settle();
      if (state !== "not-ready") {
        cell(host.target, 0, 0).click(); await settle();
        expect(button(host.target, "Harvest").disabled).toBe(false);
        button(host.target, "Harvest").focus();
        if (state === "recovering") { host.socket.closed = true; host.socket.emit("close"); }
        else { host.set(active(8), 8); host.socket.resync(); }
        await settle();
        expect(document.activeElement).toBe(host.target.querySelector("#garden-heading"));
      }
      const garden = host.target.querySelector<HTMLElement>(".garden")!;
      const stages = [...garden.querySelectorAll<HTMLButtonElement>("button.cell")].map((node) => node.dataset.stage);
      expect.soft(garden.textContent).toContain(t("common.stale_note", {}, "era_1995"));
      const commands = [...garden.querySelectorAll<HTMLButtonElement>("button")].filter((node) => node.textContent?.trim() !== "Close");
      expect(commands.length).toBeGreaterThan(36);
      expect.soft(commands.map((control) => control.disabled)).toEqual(commands.map(() => true));
      button(host.target, "Harvest all mature").click();
      button(host.target, "Mainframe").click();
      cell(host.target, 0, 0).click();
      await settle();
      expect(host.intents()).toEqual([]);
      expect([...garden.querySelectorAll<HTMLButtonElement>("button.cell")].map((node) => node.dataset.stage)).toEqual(stages);
      if (state !== "not-ready") {
        await expect.poll(() => host.sockets.length).toBe(2);
        // The main HTTP resync alone is not a recovered transport handshake.
        expect(button(host.target, "Harvest all mature").disabled).toBe(true);
      }
      host.socket.connect(state === "recovering"); await settle();
      expect(garden.textContent).not.toContain(t("common.stale_note", {}, "era_1995"));
      expect(button(host.target, "Harvest all mature").disabled).toBe(false);
      expect(host.intents()).toEqual([]); // Recovery must not replay a refused click.
      const harvest = button(host.target, "Harvest all mature");
      harvest.focus(); await userEvent.keyboard("{Enter}"); await settle();
      expect(host.intents()).toEqual([{ intent_id: expect.any(String), expected_revision: state === "resync" ? 8 : 7,
        kind: "garden_harvest", plots: [{ row: 0, col: 0 }, { row: 0, col: 1 }, { row: 1, col: 1 }] }]);
      host.assertHTTP();
    } finally { await host.dispose(); }
  });
}

for (const action of ["harvest-all", "substrate", "menu-harvest"] as const) {
  for (const newerChoice of [false, true]) {
    it.skipIf(!browser)(`Garden response focus ${action} ${newerChoice ? "preserves newer navigation" : "recovers usable control"}`, async () => {
      const { userEvent } = await import("vitest/browser");
      const host = await mounted();
      try {
        button(host.target, "Server Garden").click();
        await settle();
        const origin = cell(host.target, 0, 0);
        origin.focus();
        if (action === "menu-harvest") await userEvent.keyboard("{Enter}{Tab}");
        else await userEvent.keyboard(action === "harvest-all" ? "{Tab}" : "{Tab}{Tab}{Tab}{Tab}{Tab}");
        const trigger = document.activeElement as HTMLButtonElement;
        expect(trigger.textContent?.trim()).toBe(action === "menu-harvest" ? "Harvest" : action === "harvest-all" ? "Harvest all mature" : "Mainframe");
        await userEvent.keyboard("{Enter}");
        await settle();
        expect(host.intents()).toHaveLength(1);
        const command = host.intents()[0]!;
        expect(command.kind).toBe(action === "substrate" ? "garden_set_substrate" : "garden_harvest");
        expect(document.activeElement).toBe(action === "menu-harvest" ? origin : trigger);
        const navigation = button(host.target, "Settings");
        if (newerChoice) navigation.focus();
        const next = active(8);
        if (action === "substrate") {
          next.garden.substrate_id = "mainframe";
          next.garden.substrate_lockout_until_ms = next.server_ms + 600_000;
        } else next.garden.plots = next.garden.plots.filter((plot) => action === "menu-harvest"
          ? plot.row !== 0 || plot.col !== 0 : plot.stage !== "mature");
        host.set(next, 8);
        host.acknowledge({ outcome: "applied", intent_id: command.intent_id, kind: command.kind, founder_revision: 8 });
        await settle();
        expect(host.reads()).toBe(3);
        expect(host.intents()).toHaveLength(1);
        if (newerChoice) expect(document.activeElement).toBe(navigation);
        else if (action === "substrate") {
          expect(trigger.matches(":disabled")).toBe(true);
          expect(document.activeElement).toBe(host.target.querySelector("#garden-heading"));
        } else if (action === "menu-harvest") {
          expect(origin.dataset.stage).toBe("empty");
          expect(document.activeElement).toBe(origin);
        } else {
          expect(trigger.isConnected).toBe(false);
          // These are the two adjacent surviving tab stops around Harvest all.
          expect([origin, button(host.target, "Bare metal")]).toContain(document.activeElement);
          expect((document.activeElement as HTMLElement).matches(":disabled")).toBe(false);
        }
        host.assertHTTP();
      } finally { await host.dispose(); }
    });
  }
}

for (const row of cases) it.skipIf(!browser)(`Garden host ${row.name} binds Founder revision and refreshes after receipt`, async () => {
  const host = await mounted();
  try {
    button(host.target, "Server Garden").click(); await settle();
    expect(host.reads()).toBe(2); // Visibility probe, then mounted Garden read.
    if (row.at) { cell(host.target, row.at[0], row.at[1]).click(); await settle(); }
    const before = [...host.target.querySelectorAll<HTMLButtonElement>(".garden button.cell")].map((node) => node.dataset.stage);
    button(host.target, row.action).click(); await settle();
    expect(host.intents()).toHaveLength(1);
    const { intent_id, ...command } = host.intents()[0]!;
    expect(intent_id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
    expect(command).toEqual({ expected_revision: 7, kind: row.kind, ...row.fields });
    expect([...host.target.querySelectorAll<HTMLButtonElement>(".garden button.cell")].every((node) => !node.disabled && node.getAttribute("aria-disabled") === "true")).toBe(true);
    // ARIA alone cannot suppress events. Pending controls must not enqueue
    // another command, including a different kind, through the actual host.
    cell(host.target, 1, 0).click();
    expect(host.target.querySelector(".garden .menu")).toBeNull();
    for (const control of host.target.querySelectorAll<HTMLButtonElement>(".garden button[aria-disabled=true]:not(.cell)")) control.click();
    expect([...host.target.querySelectorAll<HTMLButtonElement>(".garden button.cell")].map((node) => node.dataset.stage)).toEqual(before);
    expect(host.reads()).toBe(2);
    const next = active(8);
    if (row.name === "plant") next.garden.plots.push({ row: 1, col: 0, species_id: "strain_a", stage: "growing", age_ticks: 0, maturation_ticks: 3, dormant: false });
    else if (row.name === "uproot" || row.name === "harvest") next.garden.plots = next.garden.plots.filter((plot) => plot.row !== 0 || plot.col !== 0);
    else if (row.name === "harvest-all") next.garden.plots = next.garden.plots.filter((plot) => plot.stage !== "mature");
    else { next.garden.substrate_id = "mainframe"; next.garden.substrate_lockout_until_ms = next.server_ms + 600_000; }
    next.garden.plots.sort((a, b) => a.row - b.row || a.col - b.col);
    host.set(next, 8);
    host.acknowledge({ outcome: "applied", intent_id, kind: row.kind, founder_revision: 8 });
    await settle();
    expect(host.reads()).toBe(3);
    expect(cell(host.target, 1, 0).disabled).toBe(false);
    if (row.name === "plant") expect(cell(host.target, 1, 0).dataset.stage).toBe("growing");
    else if (row.name === "substrate") expect(button(host.target, "Mainframe").getAttribute("aria-pressed")).toBe("true");
    else expect(cell(host.target, 0, 0).dataset.stage).toBe("empty");
    cell(host.target, 1, 0).click(); await settle();
    button(host.target, row.name === "plant" ? "Uproot" : "Plant Strain A (PENDING OWNER NAME)").click(); await settle();
    expect(host.intents()).toHaveLength(2);
    expect(host.intents()[1]?.expected_revision).toBe(8);
    host.assertHTTP();
  } finally { await host.dispose(); }
});

it.skipIf(!browser)("Garden host rejection displays Garden detail and refreshes without optimistic changes", async () => {
  const host = await mounted();
  try {
    button(host.target, "Server Garden").click(); await settle();
    cell(host.target, 0, 0).click(); await settle(); button(host.target, "Harvest").click(); await settle();
    const intent = host.intents()[0]!;
    host.acknowledge({ outcome: "rejected", intent_id: intent.intent_id, current_revision: 7, rejection: { category: "not_eligible", detail: "plant_not_mature" } });
    await settle();
    expect(host.target.querySelector(".garden [role=alert]")?.textContent).toBe("That plant is not mature yet.");
    expect(cell(host.target, 0, 0).dataset.stage).toBe("mature");
    expect(cell(host.target, 0, 0).disabled).toBe(false);
    expect(host.reads()).toBe(3); host.assertHTTP();
  } finally { await host.dispose(); }
});

it.skipIf(!browser)("Garden host keeps advisory invalidation when a late receipt's main read is already covered", async () => {
  const host = await mounted();
  try {
    button(host.target, "Server Garden").click(); await settle();
    button(host.target, "Mainframe").click(); await settle();
    const intent = host.intents()[0]!;
    const next = active(8); next.garden.substrate_id = "mainframe";
    next.garden.substrate_lockout_until_ms = next.server_ms + 600_000;
    host.set(next, 8);
    host.acknowledge({ outcome: "applied", intent_id: intent.intent_id, kind: "garden_set_substrate", founder_revision: 8 });
    await settle();
    const mainReads = host.requests.filter((request) => request.path === "/api/v1/founder/state").length;
    const gardenReads = host.reads();
    host.socket.receipt(intent.intent_id as string); await settle();
    expect(host.requests.filter((request) => request.path === "/api/v1/founder/state")).toHaveLength(mainReads);
    expect(host.reads()).toBe(gardenReads + 1);
    expect(button(host.target, "Mainframe").getAttribute("aria-pressed")).toBe("true");
    expect(host.intents()).toHaveLength(1); host.assertHTTP();
  } finally { await host.dispose(); }
});

it.skipIf(!browser)("Garden host streamed receipt re-reads mounted advisory state", async () => {
  const host = await mounted();
  try {
    button(host.target, "Server Garden").click(); await settle();
    expect(host.reads()).toBe(2);
    const next = active(8); next.garden.substrate_id = "mainframe";
    host.set(next, 8); host.socket.receipt(); await settle();
    expect(host.requests.filter((request) => request.path === "/api/v1/founder/state")).toHaveLength(2);
    expect(host.reads()).toBe(3);
    expect(button(host.target, "Mainframe").getAttribute("aria-pressed")).toBe("true");
    expect(host.intents()).toEqual([]); host.assertHTTP();
  } finally { await host.dispose(); }
});

for (const kind of ["inactive", "locked"] as const) it.skipIf(!browser)(`Garden host ${kind} visibility follows its actual read`, async () => {
  const host = await mounted(views[kind] as GardenCurrentResponse);
  try {
    const tabs = [...host.target.querySelectorAll<HTMLButtonElement>("nav button")].filter((node) => node.textContent?.trim() === "Server Garden");
    expect(tabs).toHaveLength(kind === "inactive" ? 0 : 1);
    if (kind === "locked") { tabs[0]!.click(); await settle(); expect(host.target.querySelector(".garden .grid")).toBeNull(); expect(host.target.querySelector(".garden")?.textContent).toContain("Locked"); }
    expect(host.intents()).toEqual([]); host.assertHTTP();
  } finally { await host.dispose(); }
});
