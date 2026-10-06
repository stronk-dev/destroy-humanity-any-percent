/// <reference types="@vitest/browser/providers/playwright" />

import { configDefaults, defineConfig } from "vitest/config";
import { playwright } from "@vitest/browser-playwright";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import type { Plugin } from "vite";
import type { Page, Request, WebSocket as PlaywrightWebSocket } from "playwright";

const dependencyPath = (relative: string) => decodeURIComponent(new URL(relative, import.meta.url).pathname);
const performanceOnly = process.env.VITE_GAME_UI_PERFORMANCE === "1";
const requestAudits = new Map<string, { page: Page; urls: string[]; request: (value: Request) => void; socket: (value: PlaywrightWebSocket) => void }>();

// CI D2 / RP-235: passive module observation, not request interception or retry.
function observeModuleHTTP(): Plugin {
  return {
    name: "cloud-clicker-browser-module-observation",
    configureServer(server) {
      if (!server.httpServer) throw new Error("module observation requires the real Vite HTTP server");
      let started = 0, finished = 0, closedEarly = 0;
      const log = (value: Record<string, unknown>) => console.info("browser-module-http", JSON.stringify(value));
      server.middlewares.use((request, response, next) => {
        const url = new URL(request.url ?? "/", "http://test.invalid");
        if (!url.pathname.includes("/balance/routes-testdata/") && !url.pathname.includes("/src/shell/prediction.worker.ts")) { next(); return; }
        const id = ++started, began = performance.now();
        const identity = { id, path: url.pathname, method: request.method, user_agent: request.headers["user-agent"] ?? null,
          import: url.searchParams.has("import"), worker_file: url.searchParams.has("worker_file") };
        let terminal = false;
        log({ ...identity, phase: "start" });
        response.once("finish", () => {
          terminal = true; finished++;
          log({ ...identity, phase: "finish", status: response.statusCode,
            content_type: response.getHeader("content-type") ?? null, elapsed_ms: performance.now() - began });
        });
        response.once("close", () => {
          if (terminal) return;
          terminal = true; closedEarly++;
          log({ ...identity, phase: "closed_before_finish", status: response.statusCode, elapsed_ms: performance.now() - began });
        });
        next();
      });
      server.httpServer.once("close", () => {
        const pending = started - finished - closedEarly;
        log({ phase: "summary", started, finished, closed_before_finish: closedEarly, pending, valid: pending === 0 });
        if (pending !== 0) throw new Error("invalid module observation: requests have no terminal event");
      });
    },
  };
}

export default defineConfig({
  plugins: [svelte(), observeModuleHTTP()],
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
        async setReducedMotionPreference({ provider, sessionId }, preference: "reduce" | "no-preference") {
          if (preference !== "reduce" && preference !== "no-preference") throw new Error("invalid reduced-motion preference");
          const page = (provider as typeof provider & { getPage(id: string): Page }).getPage(sessionId);
          await page.emulateMedia({ reducedMotion: preference });
        },
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
