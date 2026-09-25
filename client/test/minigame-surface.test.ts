import minigameAPISchemaSource from "../../server/account/minigame_api_schema.go?raw";
import { describe, expect, it } from "vitest";

import type { MinigameSessionResponseActive, PitchSnapshot } from "../src/api/generated/types";
import { COPY_KEYS } from "../src/copy";
import { loadPitchContent, pitchContentFor, PitchContentMismatch } from "../src/minigame/pitch-content";
import { createBrowserMinigameSessionPort, MinigameAPIError, MinigameTransportError } from "../src/minigame/session-port";
import { MINIGAME_REJECTIONS, rejectionFor, stateFromCurrent, stateFromSession } from "../src/minigame/session-surface";
import { MINIGAME_TENANT_SURFACES, parseTenantSurfaceRegistry } from "../src/minigame/tenant-registry";

const descriptor = { constants_hash: `sha256:${"a".repeat(64)}`, engine_ref: "pitch", engine_version: "1.0.0", minigame_id: "pitch", mode: "solo" as const, revision: 2, session_id: "01986666-0000-7000-8000-000000000001" };
const snapshot: PitchSnapshot = { deck_count: 17, funding_target: "1e2", hand: ["api_call#1"], hands_remaining: 3, phase: "playing", pitch_content_hash: `sha256:${"b".repeat(64)}`, pitch_schema_version: 1, revision: 2, round: 1, round_best_valuation: "0", run_currency: 4, shop_offers: [], slotted_hacks: [] };

function json(status: number, body: unknown): Response { return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } }); }

describe("minigame session port", () => {
  it("binds the generated operation paths, methods, bodies, and bearer token", async () => {
    const calls: { input: string; init?: RequestInit }[] = [];
    const port = createBrowserMinigameSessionPort(() => "tok", async (input, init) => { calls.push({ input, init }); return json(200, { kind: "none" }); });
    await port.current();
    await port.create("pitch", "key-1");
    await port.command("s/1", { command_id: "c", expected_revision: 3, command: { kind: "end_shop" } });
    await port.resolve("s1");
    expect(calls.map((call) => `${call.init?.method} ${call.input} ${call.init?.body ?? ""}`)).toEqual([
      "GET /api/v1/minigames/sessions/current ",
      "POST /api/v1/minigames/pitch/sessions {\"idempotency_key\":\"key-1\"}",
      "POST /api/v1/minigames/sessions/s%2F1/commands {\"command_id\":\"c\",\"expected_revision\":3,\"command\":{\"kind\":\"end_shop\"}}",
      "POST /api/v1/minigames/sessions/s1/resolve {}",
    ]);
    expect(calls.every((call) => (call.init?.headers as Record<string, string>).Authorization === "Bearer tok")).toBe(true);
    expect((calls[0]!.init?.headers as Record<string, string>)["Content-Type"]).toBeUndefined();
  });

  it("types exact registry errors and refuses non-exact bodies as transport failures", async () => {
    const reply = (status: number, body: unknown) => createBrowserMinigameSessionPort(() => "tok", async () => json(status, body));
    await expect(reply(409, { category: "conflict", detail: "minigame_revision" }).current()).rejects.toMatchObject({ status: 409, category: "conflict", detail: "minigame_revision" });
    await expect(reply(409, { category: "conflict", detail: "minigame_revision", extra: 1 }).current()).rejects.toBeInstanceOf(MinigameTransportError);
    await expect(reply(409, { category: "bogus", detail: "x" }).current()).rejects.toBeInstanceOf(MinigameTransportError);
    await expect(createBrowserMinigameSessionPort(() => "tok", async () => { throw new TypeError("offline"); }).current()).rejects.toBeInstanceOf(MinigameTransportError);
  });
});

describe("minigame surface rejection table", () => {
  // The exact pairs server/account/minigame_api_schema.go minigameErrorJSON
  // declares for create/command/resolve/current, minus 400/401/500 (generic).
  const SERVER_PAIRS = [
    "404 unknown_id/founder", "404 unknown_id/minigame_session", "404 unknown_id/minigame_tenant",
    "409 conflict/minigame_revision", "409 conflict/minigame_session", "409 idempotency_conflict/minigame_command",
    "409 idempotency_conflict/minigame_session", "409 not_eligible/duplicate_card", "409 not_eligible/exclusive_activity",
    "409 not_eligible/fiscal_unlock_required", "409 not_eligible/hack_slots_full", "409 not_eligible/hand_too_large",
    "409 not_eligible/human_content_locked", "409 not_eligible/illegal_phase", "409 not_eligible/insufficient_currency",
    "409 not_eligible/unknown_card", "409 not_eligible/unknown_offer", "429 rate_limited/account", "429 rate_limited/ip",
    "503 not_configured/minigame_api",
  ];

  it("maps exactly the server's pairs and nothing else", () => {
    expect([...MINIGAME_REJECTIONS.keys()].sort()).toEqual([...SERVER_PAIRS].sort());
    const source = minigameAPISchemaSource;
    const body = source.slice(source.indexOf("func minigameErrorJSON"), source.indexOf("func minigameAPIResponses"));
    const declared = new Set([...body.matchAll(/apiErrorPair\{"([a-z_]+)", "?([a-zA-Z_]+)"?\}/gu)].map((match) => `${match[1]}/${match[2] === "idempotencyDetail" ? "*" : match[2]}`));
    for (const pair of SERVER_PAIRS) {
      const suffix = pair.slice(4);
      expect(declared.has(suffix) || suffix.startsWith("idempotency_conflict/") && declared.has("idempotency_conflict/*"), pair).toBe(true);
    }
  });

  it("uses only declared copy keys", () => {
    const keys = new Set<string>(COPY_KEYS);
    for (const row of MINIGAME_REJECTIONS.values()) if (row.notice) expect(keys.has(row.notice), row.notice).toBe(true);
  });

  it("classifies transport, invariant, and unknown failures", () => {
    expect(rejectionFor(new MinigameTransportError("x")).effect).toBe("pause");
    expect(rejectionFor(new MinigameAPIError(500, "internal_invariant", "minigame_api"))).toMatchObject({ effect: "error", notice: "minigame.error.unavailable" });
    expect(rejectionFor(new MinigameAPIError(400, "invalid", "body"))).toMatchObject({ effect: "error", notice: "minigame.error.generic" });
    expect(rejectionFor(new Error("x")).effect).toBe("error");
    expect(rejectionFor(new MinigameAPIError(409, "not_eligible", "insufficient_currency"))).toMatchObject({ effect: "stay", clearSelection: false });
    expect(rejectionFor(new MinigameAPIError(409, "conflict", "minigame_revision"))).toMatchObject({ effect: "refetch", clearSelection: true });
  });
});

describe("minigame surface state", () => {
  it("derives launcher/active/paused from current and terminal from a resolved session", () => {
    expect(stateFromCurrent({ kind: "none" })).toEqual({ kind: "launcher", notice: null });
    expect(stateFromCurrent({ kind: "active", session: { ...descriptor, status: "active" }, snapshot }).kind).toBe("active");
    expect(stateFromCurrent({ kind: "active", session: { ...descriptor, status: "claimed" }, snapshot }).kind).toBe("paused_reconnect");
    const active: MinigameSessionResponseActive = { ...descriptor, status: "active", snapshot };
    expect(stateFromSession(active)).toEqual({ kind: "active", session: { ...descriptor, status: "active" }, snapshot, inFlight: null });
  });
});

describe("minigame tenant surface registry", () => {
  it("binds the pinned minigame_api tenants one-to-one", () => {
    expect([...MINIGAME_TENANT_SURFACES.entries()]).toEqual([["pitch@1.0.0", { minigame_id: "pitch", engine_ref: "pitch", engine_version: "1.0.0", child: "pitch_table" }]]);
  });

  it("refuses an artifact tenant with no child and a child with no artifact tenant", () => {
    expect(() => parseTenantSurfaceRegistry({ tenants: [{ engine_ref: "pitch", engine_version: "1.0.1", minigame_id: "pitch" }] })).toThrow(/no surface child/u);
    expect(() => parseTenantSurfaceRegistry({ tenants: [{ engine_ref: "pitch", engine_version: "1.0.0", minigame_id: "pitch" }] },
      new Map([["pitch@1.0.0", "pitch_table"], ["typer@1.0.0", "pitch_table"]]))).toThrow(/no pinned tenant/u);
  });
});

describe("pinned Pitch content", () => {
  it("accepts only the bundled bytes' identity", async () => {
    const content = await loadPitchContent();
    expect(content.hash).toMatch(/^sha256:[0-9a-f]{64}$/u);
    await expect(pitchContentFor(content.hash)).resolves.toBe(content);
    await expect(pitchContentFor(`sha256:${"0".repeat(64)}`)).rejects.toBeInstanceOf(PitchContentMismatch);
  });
});
