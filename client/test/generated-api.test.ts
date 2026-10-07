import { describe, expect, it, vi } from "vitest";
import { createAPIClient, operations, type OperationInput } from "../src/api/generated/types";

const stream = "01986666-0000-4000-8000-000000000001";

async function hash(bytes: Uint8Array): Promise<string> {
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new Uint8Array(bytes)));
  return [...digest].map((byte) => byte.toString(16).padStart(2, "0")).join("");
}

describe("registry-generated HTTP client", () => {
  it("uses registered paths, encodes path/query values, and omits undeclared credentials on public reads", async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify({ category: "invalid", detail: "variables" }), { status: 400 }));
    const client = createAPIClient(fetcher, "https://example.test/");
    const input = { path: { category: "name/with?reserved" }, request: null,
      query: { epoch: 8, mandate: 0, variables: "a+b/=", limit: 50, cursor: undefined }, accessToken: "must-not-leak", requestID: "sdk-test" };
    const response = await client.call("list_public_board", input as unknown as OperationInput<"list_public_board">);
    expect(fetcher).toHaveBeenCalledTimes(1);
    const [url, init] = fetcher.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe("https://example.test/api/public/v1/boards/name%2Fwith%3Freserved?epoch=8&limit=50&mandate=0&variables=a%2Bb%2F%3D");
    expect(init.method).toBe("GET");
    expect(init.body).toBeUndefined();
    expect(new Headers(init.headers).get("Content-Type")).toBeNull();
    expect(new Headers(init.headers).get("Authorization")).toBeNull();
    expect(new Headers(init.headers).get("X-Request-ID")).toBe("sdk-test");
    expect(response.status).toBe(400);
    expect(response.body).toEqual({ category: "invalid", detail: "variables" });
  });

  it("serializes a registered POST once and reads the current access token per call", async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify({ category: "not_eligible", detail: "tier_required" }), { status: 409 }));
    const client = createAPIClient(fetcher);
    const request = { idempotency_key: "a".repeat(64) };
    for (const accessToken of ["old", "rotated"]) await client.call("create_minigame_session", { path: { minigame_id: "pitch" }, request, accessToken });
    expect(fetcher).toHaveBeenCalledTimes(2);
    for (let index = 0; index < 2; index++) {
      const [url, init] = fetcher.mock.calls[index] as unknown as [string, RequestInit];
      expect(url).toBe(operations.create_minigame_session.path.replace("{minigame_id}", "pitch"));
      expect(init.method).toBe("POST");
      expect(init.body).toBe(JSON.stringify(request));
      expect(new Headers(init.headers).get("Content-Type")).toBe("application/json");
      expect(new Headers(init.headers).get("Authorization")).toBe(`Bearer ${index === 0 ? "old" : "rotated"}`);
    }
  });

  it.each([
    ["get_public_run_genesis", "application/json", new TextEncoder().encode('{ "z": 1, "a": "<>&" }')],
    // Opaque byte fixture: this transport checks identity, not gzip validity/replay.
    ["get_public_run_replay_log", "application/gzip", new Uint8Array([31, 139, 8, 0, 1, 2, 3, 255])],
  ] as const)("preserves raw bytes for %s and checks its registered digest", async (id, contentType, bytes) => {
    const digest = await hash(bytes);
    const fetcher = vi.fn(async () => new Response(new Uint8Array(bytes), { status: 200, headers: { "Content-Type": contentType, "X-Content-SHA256": digest } }));
    const response = await createAPIClient(fetcher).call(id, { path: { stream, seq: 1 }, request: null });
    expect(response.status).toBe(200);
    if (response.status !== 200) throw new Error("raw success missing");
    expect(response.body).toEqual(bytes);
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it.each(["missing hash", "wrong hash", "changed bytes", "wrong content type"])("rejects raw evidence with %s", async (fault) => {
    const original = new TextEncoder().encode("original");
    const bytes = fault === "changed bytes" ? new TextEncoder().encode("changed") : original;
    const headers: Record<string, string> = { "Content-Type": fault === "wrong content type" ? "text/plain" : "application/gzip" };
    if (fault !== "missing hash") headers["X-Content-SHA256"] = fault === "wrong hash" ? "0".repeat(64) : await hash(original);
    const fetcher = vi.fn(async () => new Response(bytes, { status: 200, headers }));
    await expect(createAPIClient(fetcher).call("get_public_run_replay_log", { path: { stream, seq: 1 }, request: null })).rejects.toThrow(/invalid raw API/u);
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("decodes a raw operation's registered error as JSON, not bytes", async () => {
    const fetcher = vi.fn(async () => new Response('{"category":"unknown_id","detail":"run"}', { status: 404 }));
    const response = await createAPIClient(fetcher).call("get_public_run_genesis", { path: { stream, seq: 1 }, request: null });
    expect(response.status).toBe(404);
    if (response.status !== 404) throw new Error("error status missing");
    expect(response.body.category).toBe("unknown_id");
  });

  it("propagates network/abort failures once, preserves the signal, and never refreshes or retries", async () => {
    const controller = new AbortController();
    const failure = new DOMException("aborted", "AbortError");
    const fetcher = vi.fn(async () => { throw failure; });
    await expect(createAPIClient(fetcher).call("get_game_ui_snapshot", { path: {}, request: null, accessToken: "access", signal: controller.signal })).rejects.toBe(failure);
    expect(fetcher).toHaveBeenCalledTimes(1);
    expect((fetcher.mock.calls[0] as unknown as [string, RequestInit])[1].signal).toBe(controller.signal);
  });

  it.each([[418, "{}"], [200, "not JSON"]] as const)("rejects undeclared status or malformed JSON (%s)", async (status, body) => {
    const fetcher = vi.fn(async () => new Response(body, { status }));
    await expect(createAPIClient(fetcher).call("list_public_epochs", { path: {}, request: null, query: {} })).rejects.toThrow();
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("does not echo a malformed response body in its JSON diagnostic", async () => {
    const marker = "fictional-private-body-marker";
    const fetcher = vi.fn(async () => new Response(marker, { status: 200 }));
    await expect(createAPIClient(fetcher).call("list_public_epochs", { path: {}, request: null, query: {} })).rejects.toThrow("response was not JSON (200)");
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("preserves an injected body-read failure without retrying or relabelling it as invalid JSON", async () => {
    const failure = new DOMException("aborted", "AbortError");
    const response = new Response("", { status: 200 });
    // Isolate the SDK's propagation boundary: Chromium may normalize errors
    // from a genuinely errored Response stream before the SDK sees them.
    const read = vi.spyOn(response, "text").mockRejectedValueOnce(failure);
    const fetcher = vi.fn(async () => response);
    await expect(createAPIClient(fetcher).call("list_public_epochs", { path: {}, request: null, query: {} })).rejects.toBe(failure);
    expect(fetcher).toHaveBeenCalledTimes(1);
    expect(read).toHaveBeenCalledTimes(1);
  });

  it("rejects missing auth, missing/unknown path values, inexact integers and unknown query names before HTTP", async () => {
    const fetcher = vi.fn(async () => new Response("{}"));
    const client = createAPIClient(fetcher);
    // Deliberately bypass TS to exercise invalid runtime callers.
    await expect(client.call("get_game_ui_snapshot", { path: {}, request: null } as never)).rejects.toThrow("missing API access token");
    for (const path of [{ stream }, { stream, seq: 1, private: "x" }, { stream, seq: Number.MAX_SAFE_INTEGER + 1 }]) {
      await expect(client.call("get_public_run_genesis", { path, request: null } as never)).rejects.toThrow(/API.*parameter/u);
    }
    await expect(client.call("list_public_epochs", { path: {}, request: null, query: { secret: "x" } } as never)).rejects.toThrow("unknown API query parameter");
    expect(fetcher).not.toHaveBeenCalled();
  });
});

// Compiled by make typecheck; never executed. Status narrowing retains raw/JSON
// distinctions, and consumers cannot invent an operation, request, or auth mode.
function compileContract(client: ReturnType<typeof createAPIClient>): void {
  // @ts-expect-error unknown registry operation
  void client.call("invented_operation", { path: {}, request: null });
  // @ts-expect-error private operation requires an explicit access token
  void client.call("get_game_ui_snapshot", { path: {}, request: null });
  // @ts-expect-error public operations do not take private credentials
  void client.call("list_public_epochs", { path: {}, request: null, query: {}, accessToken: "secret" });
  // @ts-expect-error bootstrap has a generated exact request shape
  void client.call("create_bootstrap", { path: {}, request: { invented: "field" } });
  // @ts-expect-error normalized board query is required
  void client.call("list_public_board", { path: { category: "any_percent" }, request: null });
}
void compileContract;
