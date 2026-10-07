import { describe, expect, it, vi } from "vitest";

import { COPY_KEYS } from "../src/copy";
import { GameUIRequestError, noticeForError, noticeForOutcome, parseIntentErrorBody, parseIntentOutcome } from "../src/game-ui/intent-outcome";
import { createBrowserGameUIRuntime, type RuntimeStorage } from "../src/game-ui/runtime";
import { OPPORTUNITY_REJECTIONS } from "../src/game-ui/opportunity-claim";

class MemoryStorage implements RuntimeStorage {
  readonly values = new Map<string, string>();
  getItem(key: string): string | null { return this.values.get(key) ?? null; }
  setItem(key: string, value: string): void { this.values.set(key, value); }
  removeItem(key: string): void { this.values.delete(key); }
}

const intentBody = Object.freeze({ intent_id: "test-intent", kind: "harvest_fiscal_period", expected_revision: 7 });

function intentRuntime(reply: () => Promise<Response>, credentialed = true) {
  const storage = new MemoryStorage();
  if (credentialed) storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({
    accessToken: "test-access", refreshToken: "test-refresh", accountID: "test-account", recoveryCode: "test-recovery",
  }));
  const originalStorage = new Map(storage.values);
  const fetcher = vi.fn<typeof fetch>(async () => reply());
  const runtime = createBrowserGameUIRuntime(storage, fetcher);
  return {
    runtime, fetcher,
    assertRequests() {
      expect(fetcher.mock.calls).toEqual(credentialed ? [["/api/v1/intents", {
        method: "POST", headers: { Authorization: "Bearer test-access", "Content-Type": "application/json" }, body: JSON.stringify(intentBody),
      }]] : []);
      expect(storage.values).toEqual(originalStorage);
    },
  };
}

describe("GS0.2 intent outcomes", () => {
  it("parses applied receipts and exact rejections, and fails closed otherwise", () => {
    expect(parseIntentOutcome({ intent_id: "i", outcome: "applied", new_revision: 2 })).toMatchObject({ outcome: "applied" });
    expect(parseIntentOutcome({ current_revision: 3, intent_id: "i", outcome: "rejected", rejection: { category: "not_eligible", detail: "exclusive_activity", session_expired: true } }))
      .toEqual({ outcome: "rejected", category: "not_eligible", detail: "exclusive_activity", currentRevision: 3, sessionExpired: true });
    expect(() => parseIntentOutcome({ current_revision: 3, intent_id: "i", outcome: "rejected", rejection: { category: "x", detail: "y", extra: 1 } })).toThrow(SyntaxError);
    expect(() => parseIntentOutcome({ current_revision: 3, intent_id: "i", outcome: "rejected", rejection: { category: "x", detail: "y" }, extra: 1 })).toThrow(SyntaxError);
    expect(() => parseIntentOutcome({ outcome: "maybe" })).toThrow(SyntaxError);
    expect(parseIntentErrorBody(409, { category: "conflict", detail: "intent" })).toBeInstanceOf(GameUIRequestError);
    expect(parseIntentErrorBody(409, { category: "conflict" })).toBeUndefined();
  });

  it("maps rejections to declared copy, surface rows first, never to offline", () => {
    const rejected = (category: string, detail: string) => ({ outcome: "rejected" as const, category, detail, currentRevision: 1, sessionExpired: false });
    expect(noticeForOutcome(rejected("revision_conflict", "expected_revision"))).toEqual({ effect: "refresh", notice: "intent.conflict", invariant: false });
    expect(noticeForOutcome(rejected("unaffordable", "company.cash")).notice).toBe("intent.rejection.unaffordable");
    expect(noticeForOutcome(rejected("not_eligible", "minigame_session_active")).notice).toBe("intent.rejection.minigame_session_active");
    expect(noticeForOutcome(rejected("not_eligible", "whatever"))).toEqual({ effect: "none", notice: "intent.rejection.unknown", invariant: true });
    expect(noticeForOutcome(rejected("unaffordable", "fiscal_credit"), new Map([["unaffordable/fiscal_credit", "intent.rejection.cap_exceeded"]])).notice).toBe("intent.rejection.cap_exceeded");
    expect(noticeForError(new GameUIRequestError(409, "conflict", "intent")).effect).toBe("refresh");
    expect(noticeForError(new GameUIRequestError(429, "rate_limited", "account")).notice).toBe("intent.rate_limited");
    expect(noticeForError(new TypeError("network")).effect).toBe("offline");
    const keys = new Set<string>(COPY_KEYS);
    for (const key of ["intent.conflict", "intent.rate_limited", "intent.rejection.unknown", "intent.rejection.unaffordable", "intent.rejection.cap_exceeded", "intent.rejection.exclusive_activity", "intent.rejection.minigame_session_active"]) expect(keys.has(key), key).toBe(true);
  });

  it("keeps the GS5 unknown-ID invariant separate from ordinary opportunity refusals", () => {
    const rejected = (category: string, detail: string) => ({ outcome: "rejected" as const, category, detail, currentRevision: 1, sessionExpired: false });
    expect(noticeForOutcome(rejected("unknown_id", "opportunity_id"), OPPORTUNITY_REJECTIONS))
      .toEqual({ effect: "none", notice: "desk.opportunity.rejection.not_pending", invariant: true });
    expect(noticeForOutcome(rejected("not_eligible", "opportunity_not_pending"), OPPORTUNITY_REJECTIONS))
      .toEqual({ effect: "none", notice: "desk.opportunity.rejection.not_pending", invariant: false });
    expect(noticeForOutcome(rejected("not_eligible", "opportunity_expired"), OPPORTUNITY_REJECTIONS))
      .toEqual({ effect: "none", notice: "desk.opportunity.rejection.expired", invariant: false });
  });

  it("holds rate-limited commands behind the existing authoritative refresh effect (GS0.2)", () => {
    expect(noticeForError(new GameUIRequestError(429, "rate_limited", "account")))
      .toEqual({ effect: "refresh", notice: "intent.rate_limited", invariant: false });
  });

  it("holds exclusive-activity commands behind the existing authoritative refresh effect (GS0.2)", () => {
    expect(noticeForOutcome({ outcome: "rejected", category: "not_eligible", detail: "exclusive_activity", currentRevision: 7, sessionExpired: false }))
      .toEqual({ effect: "refresh", notice: "intent.rejection.exclusive_activity", invariant: false });
  });

  it("returns the receipt outcome from the browser runtime and types non-2xx errors", async () => {
    const storage = new MemoryStorage();
    storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "a", refreshToken: "r", accountID: "c", recoveryCode: "d" }));
    const replies = [
      new Response(JSON.stringify({ current_revision: 4, intent_id: "i", outcome: "rejected", rejection: { category: "unaffordable", detail: "company.cash" } }), { status: 200 }),
      new Response(JSON.stringify({ category: "conflict", detail: "intent" }), { status: 409 }),
    ];
    const runtime = createBrowserGameUIRuntime(storage, (async () => replies.shift()!) as typeof fetch);
    await expect(runtime.intent({ kind: "buy_generator" })).resolves.toMatchObject({ outcome: "rejected", category: "unaffordable", currentRevision: 4 });
    await expect(runtime.intent({ kind: "buy_generator" })).rejects.toMatchObject({ status: 409, category: "conflict", detail: "intent" });
  });

  // Real runtime and Response parsing; fetch-double inputs are not service proof.
  for (const arm of [
    { status: 400, category: "invalid", detail: "intent", effect: "none", notice: "intent.rejection.unknown", invariant: true },
    { status: 409, category: "conflict", detail: "intent", effect: "refresh", notice: "intent.conflict", invariant: false },
    { status: 429, category: "rate_limited", detail: "account", effect: "refresh", notice: "intent.rate_limited", invariant: false },
    { status: 401, category: "unauthorized", detail: "access_token", effect: "offline", notice: null, invariant: false },
    { status: 404, category: "unknown_id", detail: "founder", effect: "offline", notice: null, invariant: false },
    { status: 503, category: "unavailable", detail: "server", effect: "offline", notice: null, invariant: false },
  ] as const) {
    it(`runtime HTTP boundary preserves typed ${arm.status}, mapped effects and exactly one unchanged request`, async () => {
      const { runtime, assertRequests } = intentRuntime(async () => new Response(JSON.stringify({ category: arm.category, detail: arm.detail }), { status: arm.status }));
      const result: unknown = await runtime.intent(intentBody).catch((error: unknown) => error);
      expect(result).toBeInstanceOf(GameUIRequestError);
      expect(result).toMatchObject({ status: arm.status, category: arm.category, detail: arm.detail });
      expect(noticeForError(result)).toEqual({ effect: arm.effect, notice: arm.notice, invariant: arm.invariant });
      assertRequests();
    });
  }

  for (const [label, body] of [
    ["extra key", { category: "conflict", detail: "intent", extra: true }],
    ["missing detail", { category: "conflict" }],
    ["null", null],
    ["array", [{ category: "conflict", detail: "intent" }]],
    ["nonstring detail", { category: "conflict", detail: 3 }],
    ["nonmechanical category", { category: "CONFLICT", detail: "intent" }],
    ["empty detail", { category: "conflict", detail: "" }],
  ] as const) {
    it(`runtime HTTP boundary refuses malformed 409 ${label} without turning it into actionable conflict`, async () => {
      const { runtime, assertRequests } = intentRuntime(async () => new Response(JSON.stringify(body), { status: 409 }));
      const result: unknown = await runtime.intent(intentBody).catch((error: unknown) => error);
      expect(result).toBeInstanceOf(Error);
      expect(result).not.toBeInstanceOf(GameUIRequestError);
      expect(noticeForError(result)).toEqual({ effect: "offline", notice: null, invariant: false });
      assertRequests();
    });
  }

  it("runtime HTTP boundary returns an applied receipt intact without implicit reads or renewal", async () => {
    const body = { intent_id: "test-intent", outcome: "applied", new_revision: 8, harvest_outcome: "guaranteed" };
    const { runtime, assertRequests } = intentRuntime(async () => new Response(JSON.stringify(body), { status: 200 }));
    await expect(runtime.intent(intentBody)).resolves.toEqual({ outcome: "applied", receipt: body });
    assertRequests();
  });

  it("runtime HTTP boundary returns a rejected receipt with its revision and session marker intact", async () => {
    const body = { current_revision: 8, intent_id: "test-intent", outcome: "rejected", rejection: { category: "not_eligible", detail: "exclusive_activity", session_expired: true } };
    const { runtime, assertRequests } = intentRuntime(async () => new Response(JSON.stringify(body), { status: 200 }));
    await expect(runtime.intent(intentBody)).resolves.toEqual({ outcome: "rejected", category: "not_eligible", detail: "exclusive_activity", currentRevision: 8, sessionExpired: true });
    assertRequests();
  });

  for (const [label, body] of [
    ["unknown outcome", { outcome: "maybe" }],
    ["extra rejection key", { current_revision: 8, intent_id: "test-intent", outcome: "rejected", rejection: { category: "not_eligible", detail: "exclusive_activity", extra: true } }],
  ] as const) {
    it(`runtime HTTP boundary fails closed on 200 ${label}`, async () => {
      const { runtime, assertRequests } = intentRuntime(async () => new Response(JSON.stringify(body), { status: 200 }));
      await expect(runtime.intent(intentBody)).rejects.toBeInstanceOf(SyntaxError);
      assertRequests();
    });
  }

  for (const status of [200, 503]) {
    it(`runtime HTTP boundary rejects non-JSON ${status} as offline without replay`, async () => {
      const { runtime, assertRequests } = intentRuntime(async () => new Response("not JSON", { status }));
      const result: unknown = await runtime.intent(intentBody).catch((error: unknown) => error);
      expect(result).toBeInstanceOf(SyntaxError);
      expect(noticeForError(result)).toEqual({ effect: "offline", notice: null, invariant: false });
      assertRequests();
    });
  }

  it("runtime HTTP boundary preserves a transport rejection without retrying or rewriting credentials", async () => {
    const failure = new TypeError("test transport failure");
    const { runtime, assertRequests } = intentRuntime(async () => { throw failure; });
    await expect(runtime.intent(intentBody)).rejects.toBe(failure);
    assertRequests();
  });

  it("runtime HTTP boundary refuses missing credentials before any request", async () => {
    const { runtime, assertRequests } = intentRuntime(async () => new Response("unexpected request", { status: 503 }), false);
    await expect(runtime.intent(intentBody)).rejects.toThrow("missing game UI credentials");
    assertRequests();
  });
});
