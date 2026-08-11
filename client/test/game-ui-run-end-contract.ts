import type { RunEndedEvent, RunEndSurfaceProps } from "../src/game-ui/events";

declare const ended: RunEndedEvent;

({ ended } satisfies RunEndSurfaceProps);

// The run-end surface is deliberately incapable of reading a snapshot. If a
// snapshot prop is ever added, this expected type error disappears and tsc
// fails on the now-unused directive.
// @ts-expect-error RunEndSurface accepts decoded run_ended bytes only.
({ ended, snapshot: {} } satisfies RunEndSurfaceProps);
