import { describe, expect, it } from "vitest";

import { COPY_KEYS } from "../src/copy";
import { GameUIRequestError, noticeForError, noticeForOutcome, parseIntentErrorBody, parseIntentOutcome } from "../src/game-ui/intent-outcome";
import { createBrowserGameUIRuntime, type RuntimeStorage } from "../src/game-ui/runtime";

class MemoryStorage implements RuntimeStorage {
  readonly values = new Map<string, string>();
  getItem(key: string): string | null { return this.values.get(key) ?? null; }
  setItem(key: string, value: string): void { this.values.set(key, value); }
  removeItem(key: string): void { this.values.delete(key); }
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
});
