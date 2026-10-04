/// <reference types="@vitest/browser/providers/playwright" />

import { configDefaults, defineConfig } from "vitest/config";
import { playwright } from "@vitest/browser-playwright";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import type { Page, Request, WebSocket as PlaywrightWebSocket } from "playwright";

const dependencyPath = (relative: string) => decodeURIComponent(new URL(relative, import.meta.url).pathname);
const performanceOnly = process.env.VITE_GAME_UI_PERFORMANCE === "1";
const requestAudits = new Map<string, { page: Page; urls: string[]; request: (value: Request) => void; socket: (value: PlaywrightWebSocket) => void }>();

export default defineConfig({
  plugins: [svelte()],
  resolve: { alias: {
    "@antimatter-dimensions/notations": dependencyPath("./node_modules/@antimatter-dimensions/notations/dist/ad-notations.esm.js"),
    "break_infinity.js/break_infinity": dependencyPath("./node_modules/break_infinity.js/dist/break_infinity.esm.js"),
  } },
  optimizeDeps: { include: ["svelte", "@antimatter-dimensions/notations"] },
  ssr: { noExternal: ["@antimatter-dimensions/notations"] },
  test: {
    // This CSS provenance fixture runs nested Vite builds and uses node:fs.
    // Keep other future test formats in browser collection by default.
    exclude: [...configDefaults.exclude, "test/css-dependency-graph.test.mjs"],
    setupFiles: ["./test/browser-error-guard.ts"],
    testNamePattern: performanceOnly
      ? /observable 20 Hz \/ 10 Hz screen budget/u
      : /^(?!.*observable 20 Hz \/ 10 Hz screen budget).*$/u,
    api: { host: "127.0.0.1" },
    browser: {
      enabled: true,
      headless: true,
      provider: playwright(),
      commands: {
        startRequestAudit({ provider, sessionId }) {
          if (requestAudits.has(sessionId)) throw new Error("request audit already active");
          const page = (provider as typeof provider & { getPage(id: string): Page }).getPage(sessionId);
          const urls: string[] = [];
          const request = (value: Request) => { urls.push(value.url()); };
          const socket = (value: PlaywrightWebSocket) => { urls.push(value.url()); };
          page.on("request", request);
          page.on("websocket", socket);
          requestAudits.set(sessionId, { page, urls, request, socket });
        },
        async waitForRequestAudit({ sessionId }, expected: string[]) {
          const audit = requestAudits.get(sessionId);
          if (!audit) throw new Error("request audit not active");
          const deadline = Date.now() + 5_000;
          while (Date.now() < deadline) {
            if (expected.every((url) => audit.urls.includes(url))) return;
            await new Promise((resolve) => setTimeout(resolve, 25));
          }
          throw new Error(`request audit timed out: expected=${JSON.stringify(expected)} observed=${JSON.stringify(audit.urls)}`);
        },
        stopRequestAudit({ sessionId }) {
          const audit = requestAudits.get(sessionId);
          if (!audit) throw new Error("request audit not active");
          audit.page.off("request", audit.request);
          audit.page.off("websocket", audit.socket);
          requestAudits.delete(sessionId);
          return audit.urls;
        },
      },
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
