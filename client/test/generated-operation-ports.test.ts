import { describe, expect, it, vi } from "vitest";
import { operations } from "../src/api/generated/types";
import { createBrowserGardenPort } from "../src/game-ui/garden/garden-port";
import { createBrowserMinigameSessionPort, createOperationCall, MinigameAPIError, MinigameTransportError, type Fetcher } from "../src/game-ui/minigame/session-port";
import { createBrowserSoulRecoveryPort } from "../src/game-ui/soul/recovery-surface";

function ports(fetcher: Fetcher, token = () => "port-token") {
  return { minigame: createBrowserMinigameSessionPort(token, fetcher), garden: createBrowserGardenPort(token, fetcher), soul: createBrowserSoulRecoveryPort(token, fetcher) };
}

const cases = [
  { id: "get_current_minigame_session", invoke: (port: ReturnType<typeof ports>) => port.minigame.current() },
  { id: "create_minigame_session", invoke: (port: ReturnType<typeof ports>) => port.minigame.create("pitch", "key") },
  { id: "play_minigame_command", invoke: (port: ReturnType<typeof ports>) => port.minigame.command("s1", { command_id: "c1", expected_revision: 1, command: { kind: "end_shop" } }) },
  { id: "resolve_minigame_session", invoke: (port: ReturnType<typeof ports>) => port.minigame.resolve("s1") },
  { id: "get_current_garden", invoke: (port: ReturnType<typeof ports>) => port.garden.current() },
  { id: "start_soul_recovery", invoke: (port: ReturnType<typeof ports>) => port.soul.start("defrag") },
  { id: "progress_soul_recovery", invoke: (port: ReturnType<typeof ports>) => port.soul.progress("s1", "progress") },
  { id: "resolve_soul_recovery", invoke: (port: ReturnType<typeof ports>) => port.soul.resolve("s1") },
  { id: "cancel_soul_recovery", invoke: (port: ReturnType<typeof ports>) => port.soul.cancel("s1") },
] as const;

describe.each(cases)("generated operation port $id", ({ id, invoke }) => {
  it.each([401, 429])("preserves typed rejection %s and reads auth once without retry", async (status) => {
    const category = status === 401 ? "unauthorized" : "rate_limited";
    const detail = status === 401 ? "access_token" : "account";
    const fetcher = vi.fn(async () => new Response(JSON.stringify({ category, detail }), { status }));
    const token = vi.fn(() => "port-token");
    await expect(invoke(ports(fetcher, token))).rejects.toMatchObject({ status, category, detail });
    expect(fetcher).toHaveBeenCalledTimes(1);
    expect(token).toHaveBeenCalledTimes(1);
    const [path, init] = fetcher.mock.calls[0] as unknown as [string, RequestInit];
    expect(path).toBe(operations[id].path.replace("{minigame_id}", "pitch").replace("{session_id}", "s1"));
    expect(init.method).toBe(operations[id].method);
    expect(new Headers(init.headers).get("Authorization")).toBe("Bearer port-token");
  });

  it.each(["network", "invalid JSON", "extra error field", "unknown category", "undeclared status"])("classifies %s as transport failure once", async (fault) => {
    const fetcher = vi.fn(async () => {
      if (fault === "network") throw new TypeError("offline");
      if (fault === "invalid JSON") return new Response("fictional-private-body-marker", { status: 200 });
      const body = fault === "unknown category" ? { category: "unknown_category", detail: "body" }
        : fault === "extra error field" ? { category: "invalid", detail: "body", extra: "x" } : { category: "invalid", detail: "body" };
      return new Response(JSON.stringify(body), { status: fault === "undeclared status" ? 418 : 400 });
    });
    await expect(invoke(ports(fetcher))).rejects.toBeInstanceOf(MinigameTransportError);
    expect(fetcher).toHaveBeenCalledTimes(1);
  });
});

it("reads a rotated token on the next explicit port call and keeps JSON diagnostics body-free", async () => {
  let token = "old";
  const calls: RequestInit[] = [];
  const fetcher = vi.fn(async (_input: string, init?: RequestInit) => { calls.push(init ?? {}); return new Response("fictional-private-body-marker", { status: 200 }); });
  const port = ports(fetcher, () => token);
  await expect(port.minigame.current()).rejects.toThrow("response was not JSON (200)");
  token = "new";
  await expect(port.garden.current()).rejects.toThrow("response was not JSON (200)");
  expect(fetcher).toHaveBeenCalledTimes(2);
  expect(calls.map((init) => new Headers(init.headers).get("Authorization"))).toEqual(["Bearer old", "Bearer new"]);
  expect(calls.every((init) => init.body === undefined && new Headers(init.headers).get("Content-Type") === null)).toBe(true);
});

it("preserves the API-error class for surface rejection mapping", async () => {
  const port = ports(async () => new Response('{"category":"conflict","detail":"recovery_session"}', { status: 409 }));
  await expect(port.soul.progress("s1", "t1")).rejects.toBeInstanceOf(MinigameAPIError);
});

function compileContract(call: ReturnType<typeof createOperationCall>): void {
  // @ts-expect-error callers cannot supply a handwritten method/path instead of an operation ID
  void call("POST", "/api/v1/invented");
  // @ts-expect-error the authenticated port bridge cannot dispatch public operations
  void call("list_public_epochs", { path: {}, request: null, query: {} });
  // @ts-expect-error minigame creation uses the generated request shape
  void call("create_minigame_session", { path: { minigame_id: "pitch" }, request: { invented: "x" } });
}
void compileContract;
