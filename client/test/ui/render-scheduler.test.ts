import { afterEach, describe, expect, it, vi } from "vitest";

import { SharedRenderScheduler } from "../../src/ui/render-scheduler";

describe("shared UI render scheduler", () => {
  afterEach(() => vi.useRealTimers());

  it("limits a 20 Hz producer to ten flushes per second", () => {
    vi.useFakeTimers();
    const scheduler = new SharedRenderScheduler(100);
    const owner = {};
    let renders = 0;
    for (let index = 0; index < 20; index += 1) {
      scheduler.schedule(owner, () => { renders += 1; });
      vi.advanceTimersByTime(50);
    }
    expect(renders).toBe(10);
    expect(scheduler.pendingCount).toBe(0);
  });

  it("coalesces by owner and cancels unmounted work", () => {
    vi.useFakeTimers();
    const scheduler = new SharedRenderScheduler(100);
    const owner = {};
    let value = 0;
    scheduler.schedule(owner, () => { value = 1; });
    scheduler.schedule(owner, () => { value = 2; });
    scheduler.cancel(owner);
    vi.advanceTimersByTime(100);
    expect(value).toBe(0);
  });
});
