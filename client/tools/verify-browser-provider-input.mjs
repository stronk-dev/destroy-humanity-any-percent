import assert from "node:assert/strict";
import { chromium, firefox, webkit } from "../node_modules/playwright/index.mjs";
import { PlaywrightBrowserProvider } from "../node_modules/@vitest/browser-playwright/dist/index.js";

const html = `<button id="subject">Subject</button><script>
  window.observed = { activations: 0, events: [] };
  for (const type of ["keydown", "keypress", "keyup", "click"]) {
    document.addEventListener(type, event => observed.events.push({
      type, key: event.key ?? null, trusted: event.isTrusted,
      on_subject: event.target.id === "subject"
    }));
  }
  document.querySelector("#subject").onclick = () => observed.activations++;
</script>`;

const rows = [];
for (const [engine, browserType] of Object.entries({ chromium, firefox, webkit })) {
  const browser = await browserType.launch({ headless: true });
  const commands = new Map();
  const provider = new PlaywrightBrowserProvider({
    config: { browser: { name: engine } },
    browser: { registerCommand: (name, command) => commands.set(name, command) },
  }, {});
  const keyboard = commands.get("__vitest_keyboard"), click = commands.get("__vitest_click");
  assert.equal(typeof keyboard, "function"); assert.equal(typeof click, "function");
  try {
    for (const arm of ["enter", "space", "rejected-command"]) {
      const a = await browser.newContext(), b = await browser.newContext();
      let release = () => {}, first, second;
      try {
        const pageA = await a.newPage(), pageB = await b.newPage();
        provider.pages.set("input-a", pageA); provider.pages.set("input-b", pageB);
        await pageA.setContent("<iframe></iframe>");
        await pageA.locator("iframe").evaluate((node, source) => { node.srcdoc = source; }, html);
        const frameA = await (await pageA.locator("iframe").elementHandle()).contentFrame();
        assert(frameA, "real input iframe exists");
        await pageB.setContent(html);
        const frameB = pageB.mainFrame();
        await frameA.locator("#subject").click();
        await frameA.evaluate(() => { observed.activations = 0; observed.events = []; });
        let focused;
        const ready = new Promise(resolve => { focused = resolve; });
        const gate = new Promise(resolve => { release = resolve; });
        const originalError = new Error("controlled provider command rejection");
        const gatedFrame = { evaluate: async (...args) => {
          // Real provider focus executes before the deliberate interleaving gate.
          const value = await frameA.evaluate(...args); focused(); await gate;
          if (arm === "rejected-command") throw originalError;
          return value;
        } };
        first = keyboard({ provider, sessionId: "input-a", frame: async () => gatedFrame },
          arm === "space" ? " " : "{Enter}", { unreleased: [] });
        // Attach rejection observation before releasing the deliberate failure.
        const firstResult = first.then(() => undefined, error => error);
        await ready;
        let pointerEntered = false;
        const pointerFrame = { locator: (...args) => {
          pointerEntered = true;
          return frameB.locator(...args);
        } };
        second = click({ provider, sessionId: "input-b", iframe: pointerFrame, page: pageB }, "#subject");
        const enteredBeforeRelease = pointerEntered;
        // Old provider really delivers the competing pointer before the key.
        // Corrected Firefox queues it; never await queued work while holding A.
        if (enteredBeforeRelease) await second;
        release();
        const error = await firstResult;
        await second;
        const afterA = await frameA.evaluate(() => ({ ...observed,
          focused: document.hasFocus(), active: document.activeElement?.id === "subject" }));
        const afterB = await frameB.evaluate(() => observed);
        const row = { engine, arm, enteredBeforeRelease, afterA, afterB,
          originalRejection: error === originalError };
        rows.push(row); console.log("provider-native-input", JSON.stringify(row));
        assert.equal(enteredBeforeRelease, engine !== "firefox", "only Firefox serializes native input commands");
        assert.equal(afterB.activations, 1, "queued native pointer executes once, including after rejection");
        assert(afterB.events.some(event => event.type === "click" && event.trusted && event.on_subject));
        if (arm === "rejected-command") {
          assert.equal(error, originalError, "command rejection identity survives queue recovery");
          assert.equal(afterA.activations, 0);
        } else {
          assert.equal(error, undefined);
          assert.equal(afterA.activations, 1, "the actual native key activates exactly once");
          assert(afterA.events.some(event => event.type === "click" && event.trusted && event.on_subject));
          assert(afterA.events.some(event => event.type === "keydown" && event.trusted && event.on_subject));
          assert(afterA.events.some(event => event.type === "keyup" && event.trusted && event.on_subject));
        }
      } finally {
        release();
        await Promise.allSettled([first, second]);
        provider.pages.delete("input-a"); provider.pages.delete("input-b");
        await a.close(); await b.close();
      }
    }
  } finally {
    process.off("SIGTERM", provider.onSIGTERM);
    await browser.close();
  }
}
assert.equal(rows.length, 9);
console.log("provider-native-input-complete", JSON.stringify({ cases: rows.length }));
