/// <reference types="@vitest/browser/providers/playwright" />

import { defineConfig } from "vitest/config";
import { playwright } from "@vitest/browser-playwright";
import { svelte } from "@sveltejs/vite-plugin-svelte";

const dependencyPath = (relative: string) => decodeURIComponent(new URL(relative, import.meta.url).pathname);
const performanceOnly = process.env.VITE_GAME_UI_PERFORMANCE === "1";

export default defineConfig({
  plugins: [svelte()],
  resolve: { alias: {
    "@antimatter-dimensions/notations": dependencyPath("./node_modules/@antimatter-dimensions/notations/dist/ad-notations.esm.js"),
    "break_infinity.js/break_infinity": dependencyPath("./node_modules/break_infinity.js/dist/break_infinity.esm.js"),
  } },
  optimizeDeps: { include: ["svelte", "@antimatter-dimensions/notations"] },
  ssr: { noExternal: ["@antimatter-dimensions/notations"] },
  test: {
    setupFiles: ["./test/browser-error-guard.ts"],
    testNamePattern: performanceOnly
      ? /observable 20 Hz \/ 10 Hz screen budget/u
      : /^(?!.*observable 20 Hz \/ 10 Hz screen budget).*$/u,
    api: { host: "127.0.0.1" },
    browser: {
      enabled: true,
      headless: true,
      provider: playwright(),
      viewport: { width: 1280, height: 720 },
      instances: performanceOnly
        ? [{ browser: "chromium" }]
        : [
            { browser: "chromium" },
            { browser: "firefox" },
            { browser: "webkit" },
          ],
    },
  },
});
