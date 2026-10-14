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
  it.each([401, 429].flatMap((status) => ["application/json", "Application/JSON; Charset=UTF-8"].map((contentType) => ({ status, contentType }))))(
    "preserves typed rejection $status with media $contentType and reads auth once without retry", async ({ status, contentType }) => {
      const category = status === 401 ? "unauthorized" : "rate_limited";
      const detail = status === 401 ? "access_token" : "account";
      const fetcher = vi.fn(async () => Response.json({ category, detail }, { status, headers: { "Content-Type": contentType } }));
      const token = vi.fn(() => "port-token");
      await expect(invoke(ports(fetcher, token))).rejects.toMatchObject({ status, category, detail });
      expect(fetcher).toHaveBeenCalledTimes(1);
      expect(token).toHaveBeenCalledTimes(1);
      const [path, init] = fetcher.mock.calls[0] as unknown as [string, RequestInit];
      expect(path).toBe(operations[id].path.replace("{minigame_id}", "pitch").replace("{session_id}", "s1"));
      expect(init.method).toBe(operations[id].method);
      expect(new Headers(init.headers).get("Authorization")).toBe("Bearer port-token");
    },
  );

  it.each(["network", "invalid JSON", "extra error field", "unknown category", "unknown detail", "undeclared status"])("classifies %s as transport failure once", async (fault) => {
    const fetcher = vi.fn(async () => {
      if (fault === "network") throw new TypeError("offline");
      if (fault === "invalid JSON") return new Response("fictional-private-body-marker", { status: 200, headers: { "Content-Type": "application/json" } });
      const body = fault === "unknown category" ? { category: "unknown_category", detail: "body" }
        : fault === "unknown detail" ? { category: "invalid", detail: "fictional-private-body-marker" }
        : fault === "extra error field" ? { category: "invalid", detail: "body", extra: "x" } : { category: "invalid", detail: "body" };
      return Response.json(body, { status: fault === "undeclared status" ? 418 : 400 });
    });
    let failure: unknown;
    try { await invoke(ports(fetcher)); } catch (error) { failure = error; }
    expect(failure).toBeInstanceOf(MinigameTransportError);
    if (fault === "unknown detail") {
      expect(failure).not.toBeInstanceOf(MinigameAPIError);
      expect((failure as Error).message).toBe("request failed (400)");
    }
    expect(fetcher).toHaveBeenCalledTimes(1);
  });
});

it("reads a rotated token on the next explicit port call and keeps JSON diagnostics body-free", async () => {
  let token = "old";
  const calls: RequestInit[] = [];
  const fetcher = vi.fn(async (_input: string, init?: RequestInit) => { calls.push(init ?? {}); return new Response("fictional-private-body-marker", { status: 200, headers: { "Content-Type": "application/json" } }); });
  const port = ports(fetcher, () => token);
  await expect(port.minigame.current()).rejects.toThrow("response was not JSON (200)");
  token = "new";
  await expect(port.garden.current()).rejects.toThrow("response was not JSON (200)");
  expect(fetcher).toHaveBeenCalledTimes(2);
  expect(calls.map((init) => new Headers(init.headers).get("Authorization"))).toEqual(["Bearer old", "Bearer new"]);
  expect(calls.every((init) => init.body === undefined && new Headers(init.headers).get("Content-Type") === null)).toBe(true);
});

it("preserves the API-error class for surface rejection mapping", async () => {
  const port = ports(async () => new Response('{"category":"conflict","detail":"recovery_session"}', { status: 409, headers: { "Content-Type": "application/json" } }));
  await expect(port.soul.progress("s1", "t1")).rejects.toBeInstanceOf(MinigameAPIError);
});

describe.each(cases.slice(0, 4))("minigame registry rejection binding $id", ({ id, invoke }) => {
  it.each([
    [401, "unauthorized", "minigame_session"],
    [409, "unauthorized", "access_token"],
    [409, "conflict", "recovery_session"],
    [409, "not_eligible", "fictional-private-body-marker"],
  ] as const)("refuses undeclared %s %s/%s without retry or reply-detail disclosure", async (status, category, detail) => {
    const fetcher = vi.fn(async () => Response.json({ category, detail }, { status }));
    let failure: unknown;
    try { await invoke(ports(fetcher)); } catch (error) { failure = error; }
    expect(failure).toBeInstanceOf(MinigameTransportError);
    expect(failure).not.toBeInstanceOf(MinigameAPIError);
    expect((failure as Error).message).toBe(`request failed (${status})`);
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("preserves a declared session conflict as a typed refusal", async () => {
    const fetcher = vi.fn(async () => Response.json({ category: "conflict", detail: "minigame_session" }, { status: 409 }));
    await expect(invoke(ports(fetcher))).rejects.toMatchObject({ status: 409, category: "conflict", detail: "minigame_session" });
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("binds idempotency details to this operation rather than the global error enum", async () => {
    const declared = id === "create_minigame_session" ? "minigame_session" : "minigame_command";
    const wrong = id === "create_minigame_session" ? "minigame_command" : "minigame_session";
    const fetcher = vi.fn(async () => Response.json({ category: "idempotency_conflict", detail: wrong }, { status: 409 }));
    await expect(invoke(ports(fetcher))).rejects.toBeInstanceOf(MinigameTransportError);
    expect(fetcher).toHaveBeenCalledTimes(1);
    const valid = vi.fn(async () => Response.json({ category: "idempotency_conflict", detail: declared }, { status: 409 }));
    await expect(invoke(ports(valid))).rejects.toMatchObject({ status: 409, category: "idempotency_conflict", detail: declared });
    expect(valid).toHaveBeenCalledTimes(1);
  });
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
