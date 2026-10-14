import { expect, it } from "vitest";

import { Runtime, mounted, opportunityArm, withOpportunity } from "./garage-surfaces-fixture";

const browser = typeof document !== "undefined";

// GS5-A2 stays in normal browser collection with its real minute and 70s bound.
// Its own file lets the runner schedule it beside, not after, interaction cases.
it.skipIf(!browser)("sends no command while an opportunity waits on an idle Desk (GS5-A2)", async () => {
  const runtime = new Runtime(); runtime.current = withOpportunity(opportunityArm(true));
  const { dispose } = await mounted(runtime);
  try {
    // Native timers actually run for the specified minute. The display-only
    // monotonic fixture cannot stand in for an elapsed setTimeout/interval.
    const started = performance.now();
    while (performance.now() - started < 60_000) {
      await new Promise((resolve) => setTimeout(resolve, 250));
      expect(runtime.requests).toEqual([]);
    }
    expect(performance.now() - started).toBeGreaterThanOrEqual(60_000);
  } finally { await dispose(); }
}, 70_000);
