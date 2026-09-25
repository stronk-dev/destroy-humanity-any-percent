import accountAPISource from "../../server/account/api.go?raw";
import { describe, expect, it } from "vitest";

import { COPY_KEYS } from "../src/copy";
import { MinigameAPIError, MinigameTransportError } from "../src/game-ui/minigame/session-port";
import { createBrowserSoulRecoveryPort, loadSoulRecoveryContent, RECOVERY_REJECTIONS, recoveryRejectionFor, toyCellOrder } from "../src/game-ui/soul/recovery-surface";

describe("soul recovery surface content", () => {
  it("derives the SR-C6 cadence as the pinned ceiling / 3 and exposes every pinned activity", () => {
    const content = loadSoulRecoveryContent();
    expect(content.beatIntervalMS).toBe(Math.floor(content.catalog.policy.recovery_beat_ceiling_ms / 3));
    expect(content.beatIntervalMS).toBeLessThan(content.catalog.policy.recovery_beat_ceiling_ms);
    expect(content.catalog.recovery_activities.map((row) => row.activity_id)).toEqual(["defrag", "repot", "server_room"]);
  });

  it("orders decorative toy cells as a seed-stable permutation", () => {
    const order = toyCellOrder(42, 48);
    expect([...order].sort((left, right) => left - right)).toEqual(Array.from({ length: 48 }, (_, index) => index));
    expect(toyCellOrder(42, 48)).toEqual(order);
    expect(toyCellOrder(43, 48)).not.toEqual(order);
  });
});

describe("soul recovery port", () => {
  it("binds the generated recovery operations with exact bodies", async () => {
    const calls: string[] = [];
    const port = createBrowserSoulRecoveryPort(() => "tok", async (input, init) => { calls.push(`${init?.method} ${input} ${init?.body}`); return new Response("{}", { status: 200 }); });
    await port.start("defrag");
    await port.progress("s1", "t1");
    await port.resolve("s1");
    await port.cancel("s1");
    expect(calls).toEqual([
      "POST /api/v1/soul-recovery/start {\"activity_id\":\"defrag\"}",
      "POST /api/v1/soul-recovery/progress {\"progress_token\":\"t1\",\"session_id\":\"s1\"}",
      "POST /api/v1/soul-recovery/resolve {\"session_id\":\"s1\"}",
      "POST /api/v1/soul-recovery/cancel {\"session_id\":\"s1\"}",
    ]);
  });
});

describe("soul recovery rejection table", () => {
  it("maps only pairs the server's recovery handlers actually write", () => {
    const written = new Set([...accountAPISource.matchAll(/writeError\(response, http\.Status([A-Za-z]+), "([a-z_]+)", "([a-z_]+)"\)/gu)].map((match) => `${match[1]} ${match[2]}/${match[3]}`));
    const status: Record<string, string> = { "400": "BadRequest", "404": "NotFound", "409": "Conflict", "429": "TooManyRequests", "503": "ServiceUnavailable" };
    for (const pair of RECOVERY_REJECTIONS.keys()) {
      const [code, rest] = pair.split(" ");
      expect(written.has(`${status[code!]} ${rest}`), pair).toBe(true);
    }
    for (const row of RECOVERY_REJECTIONS.values()) if ("notice" in row) expect(new Set<string>(COPY_KEYS).has(row.notice), row.notice).toBe(true);
  });

  it("reconnects on stale tokens and transport failures, ends on a gone session, and errors on the unknown", () => {
    expect(recoveryRejectionFor(new MinigameAPIError(400, "not_eligible", "recovery_token")).effect).toBe("reconnect");
    expect(recoveryRejectionFor(new MinigameTransportError("x")).effect).toBe("reconnect");
    expect(recoveryRejectionFor(new MinigameAPIError(404, "unknown_id", "recovery_session")).effect).toBe("gone");
    expect(recoveryRejectionFor(new MinigameAPIError(400, "not_eligible", "soul_recovery_not_ready"))).toEqual({ effect: "notice", notice: "soul.recovery_surface.not_ready" });
    expect(recoveryRejectionFor(new MinigameAPIError(500, "internal_invariant", "session_id")).effect).toBe("error");
    expect(recoveryRejectionFor(new MinigameAPIError(400, "invalid", "body")).effect).toBe("error");
  });
});
