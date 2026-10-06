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
  constructor() { super(); queueMicrotask(() => { if (!this.closed) this.dispatchEvent(new Event("open")); }); }
  send(raw: string): void {
    const command = JSON.parse(raw) as Record<string, unknown>;
    this.commands.push(command);
    const reply = command.id === 1 ? { id: 1, connect: { client: "controlled" } } :
      { id: command.id, subscribe: { recoverable: true, positioned: true, epoch: "controlled", offset: 0, publications: [] } };
    queueMicrotask(() => { if (!this.closed) this.dispatchEvent(new MessageEvent("message", { data: JSON.stringify(reply) })); });
  }
  close(): void { this.closed = true; this.dispatchEvent(new CloseEvent("close", { code: 1000 })); }
}
function controlledBoundary(tier: number) {
  const initial = wireSnapshot(tier);
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
    const socket = new ControlledSocket(); sockets.push(socket); return socket as unknown as WebSocket;
  }, { protocol: "http:", host: "controlled.invalid" });
  return {
    runtime, storage, requests, sockets, headers, unexpected,
    get snapshotCalls() { return snapshotCalls; },
    deliverIntent(value: unknown, status = 200) { expect(heldIntents).toHaveLength(1); heldIntents.shift()!(Response.json(value, { status })); },
    deliverSnapshot(value: GameUISnapshot) { expect(heldSnapshots).toHaveLength(1); heldSnapshots.shift()!(Response.json(value)); },
    cleanup() {
      for (const finish of heldIntents.splice(0)) finish(Response.json({ outcome: "applied" }));
      for (const finish of heldSnapshots.splice(0)) finish(Response.json(initial));
    },
  };
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
function rejection(intentID: unknown, category: Category, detail = rawDetail) {
  return { current_revision: 7, intent_id: intentID, outcome: "rejected", rejection: { category, detail } };
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
            boundary.deliverIntent(status === 409 ? { category, detail } : rejection(boundary.requests[0]!.intent_id, category, detail), status);
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
            boundary.deliverIntent(rejection(boundary.requests[1]!.intent_id, "invalid"));
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
