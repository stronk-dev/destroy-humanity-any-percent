import { describe, expect, it, vi } from "vitest";
import { createAPIClient, operations, type APIError, type OperationInput } from "../src/api/generated/types";

const stream = "01986666-0000-4000-8000-000000000001";

async function hash(bytes: Uint8Array): Promise<string> {
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new Uint8Array(bytes)));
  return [...digest].map((byte) => byte.toString(16).padStart(2, "0")).join("");
}

describe("registry-generated HTTP client", () => {
  const schemaCases = [
    { id: "get_public_run_verdict", status: 200, body: { catalog_url: "/api/public/v1/catalogs/fixture", constants_hash: `sha256:${"a".repeat(64)}`,
      engine_version: "0.1.0", genesis_sha256: `sha256:${"b".repeat(64)}`, genesis_url: "/genesis", replay_log_sha256: `sha256:${"c".repeat(64)}`,
      replay_log_url: "/replay-log", run_id: `${stream}:1`, verdict: "verified" }, input: { path: { stream, seq: 1 }, request: null } },
    { id: "get_public_run_verdict", status: 404, body: { category: "unknown_id", detail: "run" }, input: { path: { stream, seq: 1 }, request: null } },
    { id: "refresh_session", status: 200, body: { access_token: "fixture-access", refresh_token: "fixture-refresh" }, input: { path: {}, request: { refresh_token: "fixture-parent" } } },
    { id: "refresh_session", status: 401, body: { category: "unauthorized", detail: "refresh_token" }, input: { path: {}, request: { refresh_token: "fixture-parent" } } },
  ] as const;

  it.each(schemaCases.flatMap((entry) => [null, "text/plain", "application/gzip"].map((contentType) => ({ ...entry, contentType }))))(
    "refuses single-schema $id/$status with media $contentType before reading the body", async ({ id, status, body, input, contentType }) => {
      const headers = new Headers();
      if (contentType !== null) headers.set("Content-Type", contentType);
      // Bytes avoid Response's automatic text/plain default in the absent case.
      const response = new Response(new TextEncoder().encode(JSON.stringify(body)), { status, headers });
      const read = vi.spyOn(response, "text");
      const fetcher = vi.fn(async () => response);
      await expect(createAPIClient(fetcher).call(id, input as OperationInput<typeof id>)).rejects.toThrow("invalid schema API content type");
      expect(read).not.toHaveBeenCalled();
      expect(fetcher).toHaveBeenCalledTimes(1);
    },
  );

  it.each(schemaCases.flatMap((entry) => ["application/json", "application/json; charset=utf-8"].map((contentType) => ({ ...entry, contentType }))))(
    "preserves single-schema $id/$status with declared JSON media $contentType", async ({ id, status, body, input, contentType }) => {
      const fetcher = vi.fn(async () => Response.json(body, { status, headers: { "Content-Type": contentType } }));
      const result = await createAPIClient(fetcher).call(id, input as OperationInput<typeof id>);
      expect(result.status).toBe(status);
      expect(result.body).toEqual(body);
      expect(fetcher).toHaveBeenCalledTimes(1);
    },
  );

  it("creates an account once without forwarding an access token", async () => {
    const account = { account_id: stream, created_at: "2026-10-08T00:00:00Z", recovery_code: "fixture-recovery" };
    const fetcher = vi.fn(async () => Response.json(account, { status: 201, headers: { "Cache-Control": "no-store" } }));
    const input = { path: {}, request: {}, accessToken: "must-not-leak" };
    const result = await createAPIClient(fetcher).call("create_account", input as unknown as OperationInput<"create_account">);
    expect(result.status).toBe(201);
    expect(result.body).toEqual(account);
    expect(result.headers.get("Cache-Control")).toBe("no-store");
    expect(fetcher).toHaveBeenCalledTimes(1);
    const [url, init] = fetcher.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe("/api/v1/account");
    expect(init.method).toBe("POST");
    expect(init.body).toBe("{}");
    expect(new Headers(init.headers).get("Authorization")).toBeNull();
  });

  it.each([
    ["create_founder", "POST", 201, { id: stream, created_at: "2026-10-08T00:00:00.12Z", imported: false }],
    ["get_founder", "GET", 200, { id: stream, created_at: "2026-10-08T00:00:00.12Z", display: {} }],
  ] as const)("dispatches %s with the current explicit credential and unchanged timestamp", async (id, method, status, founder) => {
    const fetcher = vi.fn(async () => Response.json(founder, { status }));
    const request = id === "create_founder" ? {} : null;
    const result = await createAPIClient(fetcher).call(id, { path: {}, request, accessToken: "fixture-access" } as OperationInput<typeof id>);
    expect(result.status).toBe(status);
    expect(result.body).toEqual(founder);
    expect(fetcher).toHaveBeenCalledTimes(1);
    const [url, init] = fetcher.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe("/api/v1/founder");
    expect(init.method).toBe(method);
    expect(init.body).toBe(method === "GET" ? undefined : "{}");
    expect(new Headers(init.headers).get("Authorization")).toBe("Bearer fixture-access");
  });

  it.each([
    ["create_account", "account_create"],
    ["create_founder", "founder_create"],
  ] as const)("preserves %s failure without repeating a non-idempotent creation", async (id, detail) => {
    const body: APIError = { category: "internal_invariant", detail };
    const fetcher = vi.fn(async () => Response.json(body, { status: 500 }));
    const input = id === "create_account" ? { path: {}, request: {} } : { path: {}, request: {}, accessToken: "fixture-access" };
    const result = await createAPIClient(fetcher).call(id, input as OperationInput<typeof id>);
    expect(result.status).toBe(500);
    expect(result.ok).toBe(false);
    expect(result.body).toEqual(body);
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it.each([
    ["create_session", "/api/v1/session", { account_id: stream, recovery_code: "fixture-recovery" }],
    ["refresh_session", "/api/v1/session/refresh", { refresh_token: "fixture-refresh" }],
  ] as const)("dispatches existing %s without an access token or automatic renewal", async (id, path, request) => {
    const pair = { access_token: "fixture-access", refresh_token: "fixture-descendant" };
    const fetcher = vi.fn(async () => Response.json(pair, { status: 200 }));
    // Invalid runtime callers cannot accidentally forward an unrelated access token
    // to a credential-issuing operation declared AuthNone.
    const input = { path: {}, request, accessToken: "must-not-leak" };
    const result = await createAPIClient(fetcher).call(id, input as unknown as OperationInput<typeof id>);
    expect(result.status).toBe(200);
    expect(result.body).toEqual(pair);
    expect(fetcher).toHaveBeenCalledTimes(1);
    const [url, init] = fetcher.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe(path);
    expect(init.method).toBe("POST");
    expect(init.body).toBe(JSON.stringify(request));
    expect(new Headers(init.headers).get("Content-Type")).toBe("application/json");
    expect(new Headers(init.headers).get("Authorization")).toBeNull();
  });

  it.each([
    [400, "invalid", "body"],
    [401, "unauthorized", "refresh_token"],
    [401, "refresh_reused", "session_family_revoked"],
    [429, "rate_limited", "ip"],
  ] as const)("preserves refresh %s %s/%s without retrying a consumed credential", async (status, category, detail) => {
    const body: APIError = { category, detail };
    const fetcher = vi.fn(async () => Response.json(body, { status }));
    const result = await createAPIClient(fetcher).call("refresh_session", { path: {}, request: { refresh_token: "fixture-refresh" } });
    expect(result.status).toBe(status);
    expect(result.ok).toBe(false);
    expect(result.body).toEqual(body);
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("uses registered paths, encodes path/query values, and omits undeclared credentials on public reads", async () => {
    const fetcher = vi.fn(async () => Response.json({ category: "invalid", detail: "variables" }, { status: 400 }));
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
    const fetcher = vi.fn(async () => Response.json({ category: "not_eligible", detail: "tier_required" }, { status: 409 }));
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
    ["public board", "."], ["public board", ".."],
    ["authenticated creation", "."], ["authenticated creation", ".."],
  ] as const)("refuses %s path navigation %s before HTTP", async (kind, value) => {
    // Request applies the real URL parser: encodeURIComponent alone leaves
    // these segments navigational. Never send a credential to the rewritten path.
    const fetcher = vi.fn(async (url: string, init?: RequestInit): Promise<Response> => {
      const actual = new Request(url, init);
      throw new Error(`unexpected HTTP to ${new URL(actual.url).pathname}`);
    });
    const client = createAPIClient(fetcher, "https://example.test");
    const call = kind === "public board"
      ? client.call("list_public_board", { path: { category: value }, request: null,
        query: { epoch: 8, mandate: 0, variables: "fixture" } })
      : client.call("create_minigame_session", { path: { minigame_id: value },
        request: { idempotency_key: "a".repeat(64) }, accessToken: "fixture-access" });
    await expect(call).rejects.toThrow("invalid API path segment");
    expect(fetcher).not.toHaveBeenCalled();
  });

  it.each([
    [".profile", ".profile"], ["version..next", "version..next"],
    ["%2e", "%252e"], ["%2e%2e", "%252e%252e"],
    [". /", ".%20%2F"], ["x/y?reserved#", "x%2Fy%3Freserved%23"],
  ])("preserves non-navigational path value %s and query dots", async (value, encoded) => {
    const body = { category: "invalid", detail: "variables" };
    const fetcher = vi.fn(async (url: string, init?: RequestInit) => {
      const actual = new Request(url, init);
      const parsed = new URL(actual.url);
      expect(parsed.pathname).toBe(`/api/public/v1/boards/${encoded}`);
      expect(parsed.searchParams.get("cursor")).toBe("..");
      expect(actual.headers.get("Authorization")).toBeNull();
      return Response.json(body, { status: 400 });
    });
    const response = await createAPIClient(fetcher, "https://example.test").call("list_public_board", {
      path: { category: value }, request: null, query: { epoch: 8, mandate: 0, variables: "fixture", cursor: ".." },
    });
    expect(response.status).toBe(400);
    expect(response.body).toEqual(body);
    expect(fetcher).toHaveBeenCalledTimes(1);
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

  it.each(["raw", "schema"] as const)("selects a same-status %s JSON representation by media type", async (kind) => {
    // Exercise the transport with the legal C19 same-status metadata population
    // checked by Go's TestGeneratedClientSameStatusRepresentations. No mounted
    // endpoint gains a representation; restore the generated row after the test.
    const operation = operations.get_public_run_replay_log;
    const original = operation.responses;
    const raw = original.find((descriptor) => descriptor.status === 200)!;
    const json = { status: 200, kind, contentType: "application/json", contentHashHeader: kind === "raw" ? "X-Content-SHA256" : "" };
    const text = '{ "category": "unknown_id", "detail": "run" }';
    const bytes = new TextEncoder().encode(text);
    const digest = await hash(bytes);
    Object.defineProperty(operation, "responses", { value: [raw, json, ...original.filter((descriptor) => descriptor.status !== 200)] });
    try {
      for (const contentType of ["application/gzip", "application/json; charset=utf-8"]) {
        const fetcher = vi.fn(async () => new Response(bytes, { status: 200, headers: {
          "Content-Type": contentType, "X-Content-SHA256": digest,
        } }));
        const response = await createAPIClient(fetcher).call("get_public_run_replay_log", { path: { stream, seq: 1 }, request: null });
        expect(response.body).toEqual(kind === "schema" && contentType.startsWith("application/json") ? JSON.parse(text) : bytes);
        expect(fetcher).toHaveBeenCalledTimes(1);
      }
      if (kind === "raw") {
        for (const contentHash of [null, "0".repeat(64)]) {
          const headers = new Headers({ "Content-Type": "application/json" });
          if (contentHash !== null) headers.set("X-Content-SHA256", contentHash);
          await expect(createAPIClient(async () => new Response(bytes, { status: 200, headers })).call(
            "get_public_run_replay_log", { path: { stream, seq: 1 }, request: null },
          )).rejects.toThrow("invalid raw API content hash");
        }
      }
      for (const contentType of ["text/plain", null]) {
        const headers = new Headers({ "X-Content-SHA256": digest });
        if (contentType !== null) headers.set("Content-Type", contentType);
        const response = new Response(bytes, { status: 200, headers });
        const read = vi.spyOn(response, "arrayBuffer");
        const readText = vi.spyOn(response, "text");
        await expect(createAPIClient(async () => response).call("get_public_run_replay_log", { path: { stream, seq: 1 }, request: null })).rejects.toThrow(/content type/u);
        expect(read).not.toHaveBeenCalled();
        expect(readText).not.toHaveBeenCalled();
      }
    } finally { Object.defineProperty(operation, "responses", { value: original }); }
  });

  it("decodes a raw operation's registered error as JSON, not bytes", async () => {
    const fetcher = vi.fn(async () => new Response('{"category":"unknown_id","detail":"run"}', { status: 404, headers: { "Content-Type": "application/json" } }));
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
    const fetcher = vi.fn(async () => new Response(body, { status, headers: { "Content-Type": "application/json" } }));
    await expect(createAPIClient(fetcher).call("list_public_epochs", { path: {}, request: null, query: {} })).rejects.toThrow();
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("does not echo a malformed response body in its JSON diagnostic", async () => {
    const marker = "fictional-private-body-marker";
    const fetcher = vi.fn(async () => new Response(marker, { status: 200, headers: { "Content-Type": "application/json" } }));
    await expect(createAPIClient(fetcher).call("list_public_epochs", { path: {}, request: null, query: {} })).rejects.toThrow("response was not JSON (200)");
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("preserves an injected body-read failure without retrying or relabelling it as invalid JSON", async () => {
    const failure = new DOMException("aborted", "AbortError");
    const response = new Response("", { status: 200, headers: { "Content-Type": "application/json" } });
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
  // RP-376: actual registered Soul error replies must be representable.
  const unavailable: APIError = { category: "not_configured", detail: "soul_recovery" };
  const missingCompany: APIError = { category: "unknown_id", detail: "company_stream" };
  const reused: APIError = { category: "refresh_reused", detail: "session_family_revoked" };
  void reused;
  void client.call("create_account", { path: {}, request: {} });
  void client.call("create_founder", { path: {}, request: {}, accessToken: "fixture-access" });
  void client.call("get_founder", { path: {}, request: null, accessToken: "fixture-access" });
  // @ts-expect-error account creation cannot introduce a private field
  void client.call("create_account", { path: {}, request: { email: "private" } });
  // @ts-expect-error New Founder has no caller-authored identity
  void client.call("create_founder", { path: {}, request: { id: stream }, accessToken: "fixture-access" });
  // @ts-expect-error New Founder remains authenticated
  void client.call("create_founder", { path: {}, request: {} });
  void client.call("create_session", { path: {}, request: { account_id: stream, recovery_code: "fixture-recovery" } });
  void client.call("refresh_session", { path: {}, request: { refresh_token: "fixture-refresh" } });
  // @ts-expect-error refresh's credential is not an access-token-authenticated call
  void client.call("refresh_session", { path: {}, request: { refresh_token: "fixture-refresh" }, accessToken: "private" });
  // @ts-expect-error recovery session issuance needs the recovery credential
  void client.call("create_session", { path: {}, request: { account_id: stream } });
  void unavailable; void missingCompany;
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
