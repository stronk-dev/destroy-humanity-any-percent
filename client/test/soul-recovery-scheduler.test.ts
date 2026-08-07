import { afterEach, describe, expect, it, vi } from "vitest";
import { RecoveryScheduler, type RecoveryProgressResponse } from "../src/soul/recovery-scheduler";

const session = "01986666-0950-7000-8000-000000000951";
const token = "01986666-0950-4000-8000-000000000952";
const rotated = "01986666-0950-4000-8000-000000000953";

afterEach(() => vi.useRealTimers());

describe("Soul recovery scheduler", () => {
  it("beats only while visible and never replays a hidden interval", async () => {
    vi.useFakeTimers();
    let now = 1000;
    let visibility: ((visible: boolean) => void) | undefined;
    const progress = vi.fn(async (): Promise<RecoveryProgressResponse> => ({ session_id: session,
      attended_progress_ms: 100, required_duration_attended_ms: 300, last_progress_server_ms: now, eligible: false }));
    const events: string[] = [];
    const scheduler = new RecoveryScheduler({ session_id: session, progress_token: token, beat_interval_ms: 100,
      transport: { progress }, now: () => now, visibility: { subscribe: (callback) => { visibility = callback; return () => { visibility = undefined; }; } } },
    { on_progress: () => events.push("progress"), on_pause: (reason) => events.push(`pause:${reason}`), on_resume: () => events.push("resume"),
      on_token_rotated: () => events.push("rotated"), on_terminal: (kind) => events.push(`terminal:${kind}`) });
    scheduler.start();
    now += 100; await vi.advanceTimersByTimeAsync(100);
    expect(progress).toHaveBeenCalledTimes(1);
    visibility?.(false);
    now += 10_000; await vi.advanceTimersByTimeAsync(10_000);
    expect(progress).toHaveBeenCalledTimes(1);
    visibility?.(true);
    now += 100; await vi.advanceTimersByTimeAsync(100);
    expect(progress).toHaveBeenCalledTimes(2);
    expect(events).toEqual(["progress", "pause:hidden", "resume", "progress"]);
    scheduler.stop("resolved");
    expect(events.at(-1)).toBe("terminal:resolved");
  });

  it("rotates the token and pauses rather than queueing failed beats", async () => {
    vi.useFakeTimers();
    let now = 1000;
    const progress = vi.fn().mockRejectedValueOnce(new Error("offline")).mockResolvedValue({ session_id: session,
      attended_progress_ms: 100, required_duration_attended_ms: 300, last_progress_server_ms: 1200, eligible: false });
    const events: string[] = [];
    const scheduler = new RecoveryScheduler({ session_id: session, progress_token: token, beat_interval_ms: 100,
      transport: { progress }, now: () => now, visibility: { subscribe: () => () => undefined } },
    { on_progress: () => events.push("progress"), on_pause: (reason) => events.push(`pause:${reason}`), on_resume: () => undefined,
      on_token_rotated: (value) => events.push(`token:${value}`), on_terminal: () => undefined });
    scheduler.start();
    now += 100; await vi.advanceTimersByTimeAsync(100);
    expect(events).toEqual(["pause:network"]);
    scheduler.reconnect(rotated);
    now += 100; await vi.advanceTimersByTimeAsync(100);
    expect(progress).toHaveBeenLastCalledWith(session, rotated);
    expect(events).toEqual(["pause:network", `token:${rotated}`, "progress"]);
    scheduler.stop();
  });
});
