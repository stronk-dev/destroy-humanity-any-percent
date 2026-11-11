// Public standalone runner: retain all backend checks for every existing mode.
import { composedMode } from "./observe-manual-budget.mjs";
import { assertComposedPortAvailable, runComposedPreflight } from "./test-game-ui-preflight.mjs";

const args = process.argv.slice(2);
if (!(args.length === 1 && ["--tier2-fixture", "--soul-recovery-full"].includes(args[0]))) composedMode(args);
await assertComposedPortAvailable();
runComposedPreflight();
await import("./game-ui-composed-journey.mjs");
