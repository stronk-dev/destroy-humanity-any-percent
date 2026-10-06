import { flushSync, mount, tick, unmount, type ComponentProps } from "svelte";
import { expect, it, vi } from "vitest";

import views from "../../testdata/garden/view-fixtures-v1.json";
import type { GardenCurrentResponse } from "../src/api/generated/types";
import GardenSurface from "../src/game-ui/garden/GardenSurface.svelte";
import { createBrowserGardenPort } from "../src/game-ui/garden/garden-port";
import GardenSurfaceHarness from "./GardenSurfaceHarness.svelte";

const browser = typeof document !== "undefined";
type Active = Extract<GardenCurrentResponse, { kind: "active" }>;
// Tests in this file run serially; each mounted host owns the recorded promises.
let reads: Promise<GardenCurrentResponse>[] = [];

// DTOs are supplied by an injected HTTP boundary, not simulated by the UI.
function growing(wait: number | null): Active {
  const view = structuredClone(views.fresh) as Active;
  view.garden.next_tick_wall_ms = wait === null ? null : view.server_ms + wait;
  return view;
}

function matured(): Active {
  const view = growing(null);
  view.server_ms += 300_000;
  view.garden.tick_seq += 1;
  view.garden.plots = view.garden.plots.map((plot) => plot.row === 1 && plot.col === 1
    ? { ...plot, stage: "mature", age_ticks: 4 } : plot);
  return view;
}

async function settle(): Promise<void> {
  await tick(); flushSync();
  // A recorded fetch is not completion: Response.json reads a native stream.
  // Join actual adapter promises without advancing the virtual deadline.
  await Promise.allSettled(reads);
  await tick(); flushSync();
}

function mounted(response: Active) {
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  let visibility: DocumentVisibilityState = "visible";
  const visibilitySpy = vi.spyOn(document, "visibilityState", "get").mockImplementation(() => visibility);
  const requests: { input: string; method: string | undefined; headers: Headers; body: BodyInit | null | undefined }[] = [];
  let next: GardenCurrentResponse | Error = response;
  const commands: string[] = [];
  const port = createBrowserGardenPort(() => "garden-refresh-token", async (input, init) => {
    requests.push({ input, method: init?.method, headers: new Headers(init?.headers), body: init?.body });
    if (next instanceof Error) throw next;
    return new Response(JSON.stringify(next), { status: 200, headers: { "Content-Type": "application/json" } });
  });
  reads = [];
  const current = port.current;
  port.current = () => {
    const request = current();
    reads.push(request);
    void request.catch(() => {});
    return request;
  };
  const target = document.createElement("main");
  document.body.append(target);
  const initial: ComponentProps<typeof GardenSurface> = {
    port, era: "era_1995", pending: false, refreshKey: 0, rejection: null,
    onPlant: () => commands.push("plant"), onUproot: () => commands.push("uproot"),
    onHarvest: () => commands.push("harvest"), onSetSubstrate: () => commands.push("substrate"),
    // Deliberately do not override visible(): exercise the actual Page Visibility reader.
  };
  const app = mount(GardenSurfaceHarness, { target, props: { initial } });
  let destroyed = false;
  return {
    target, app, requests, commands,
    respond(value: GardenCurrentResponse | Error) { next = value; },
    visibility(value: DocumentVisibilityState) {
      visibility = value;
      document.dispatchEvent(new Event("visibilitychange"));
    },
    async destroy() { await unmount(app); destroyed = true; },
    async cleanup() {
      if (!destroyed) await unmount(app);
      target.remove(); visibilitySpy.mockRestore(); vi.useRealTimers();
    },
    assertReads(count: number) {
      expect(requests).toHaveLength(count);
      for (const request of requests) {
        expect(request.input).toBe("/api/v1/garden/current");
        expect(request.method).toBe("GET");
        expect([...request.headers.entries()]).toEqual([["authorization", "Bearer garden-refresh-token"]]);
        expect(request.body).toBeUndefined();
      }
      expect(commands).toEqual([]);
    },
  };
}

for (const wait of [60_000, 137]) {
  it.skipIf(!browser)(`Garden refresh reads at the server-relative ${wait} ms boundary without client growth`, async () => {
    const host = mounted(growing(wait));
    try {
      await settle(); host.assertReads(1);
      expect([...host.target.querySelectorAll<HTMLButtonElement>("button.cell")][7]?.dataset.stage).toBe("growing");
      const original = host.target.innerHTML;
      host.respond(matured());
      await vi.advanceTimersByTimeAsync(wait - 1); await settle();
      host.assertReads(1);
      expect(host.target.innerHTML).toBe(original);
      await vi.advanceTimersByTimeAsync(1); await settle();
      host.assertReads(2);
      expect([...host.target.querySelectorAll<HTMLButtonElement>("button.cell")][7]?.dataset.stage).toBe("mature");
      expect(host.target.querySelector(".live")?.textContent).toBe("1 matured, 0 new seedlings");
      await vi.advanceTimersByTimeAsync(300_000); await settle();
      host.assertReads(2);
    } finally { await host.cleanup(); }
  });
}

it.skipIf(!browser)("Garden refresh defers a hidden deadline and reads once on visible return", async () => {
  const host = mounted(growing(60_000));
  try {
    await settle(); host.assertReads(1);
    const original = host.target.innerHTML;
    host.respond(matured()); host.visibility("hidden"); await settle();
    host.assertReads(1);
    await vi.advanceTimersByTimeAsync(360_000); await settle();
    host.assertReads(1); expect(host.target.innerHTML).toBe(original);
    host.visibility("visible"); await settle();
    host.assertReads(2);
    expect([...host.target.querySelectorAll<HTMLButtonElement>("button.cell")][7]?.dataset.stage).toBe("mature");
  } finally { await host.cleanup(); }
});

it.skipIf(!browser)("Garden receipt refresh replaces the earlier tick deadline", async () => {
  const host = mounted(growing(60_000));
  try {
    await settle(); host.assertReads(1);
    await vi.advanceTimersByTimeAsync(10_000);
    host.respond(growing(120_000)); host.app.update({ refreshKey: 1 }); await settle();
    host.assertReads(2);
    await vi.advanceTimersByTimeAsync(50_000); await settle();
    host.assertReads(2);
    host.respond(matured());
    await vi.advanceTimersByTimeAsync(69_999); await settle(); host.assertReads(2);
    await vi.advanceTimersByTimeAsync(1); await settle(); host.assertReads(3);
    expect([...host.target.querySelectorAll<HTMLButtonElement>("button.cell")][7]?.dataset.stage).toBe("mature");
  } finally { await host.cleanup(); }
});

it.skipIf(!browser)("Garden refresh with no deadline does not poll or simulate growth", async () => {
  const host = mounted(growing(null));
  try {
    await settle(); host.assertReads(1);
    const original = host.target.innerHTML;
    host.respond(matured());
    await vi.advanceTimersByTimeAsync(86_400_000); await settle();
    host.assertReads(1); expect(host.target.innerHTML).toBe(original);
  } finally { await host.cleanup(); }
});

it.skipIf(!browser)("Garden unmount removes scheduled reads and visibility listener", async () => {
  const host = mounted(growing(60_000));
  try {
    await settle(); host.assertReads(1);
    await host.destroy();
    host.visibility("visible");
    await vi.advanceTimersByTimeAsync(86_400_000); await settle();
    host.assertReads(1); expect(host.target.childElementCount).toBe(0);
  } finally { await host.cleanup(); }
});

it.skipIf(!browser)("Garden failed due read is stale and visible resume recovers without simulated growth", async () => {
  const host = mounted(growing(60_000));
  try {
    await settle(); host.assertReads(1);
    host.respond(new Error("transport offline"));
    await vi.advanceTimersByTimeAsync(60_000); await settle(); host.assertReads(2);
    expect(host.target.textContent).toContain("This view may be out of date. Refreshing.");
    expect([...host.target.querySelectorAll<HTMLButtonElement>("button.cell")][7]?.dataset.stage).toBe("growing");
    host.visibility("hidden"); await settle(); host.assertReads(2);
    host.respond(matured()); host.visibility("visible"); await settle(); host.assertReads(3);
    expect(host.target.textContent).not.toContain("This view may be out of date. Refreshing.");
    expect([...host.target.querySelectorAll<HTMLButtonElement>("button.cell")][7]?.dataset.stage).toBe("mature");
  } finally { await host.cleanup(); }
});
