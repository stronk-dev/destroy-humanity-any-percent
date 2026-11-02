import assert from "node:assert/strict";
import { chromium, firefox, webkit } from "../../client/node_modules/playwright/index.mjs";

// CI diagnosis only: distinguish the published iframe-blur defect from lost
// native activation. Synthetic buttons, no product or credential integration.
const subjectHTML = `<button id="subject">Subject</button><script>
  window.observed = { blurs: 0, activations: 0, events: [] };
  addEventListener("blur", () => observed.blurs++);
  for (const type of ["keydown", "keypress", "keyup", "click"]) {
    document.addEventListener(type, event => observed.events.push({
      type, key: event.key ?? null, trusted: event.isTrusted,
      on_subject: event.target.id === "subject", repeat: event.repeat ?? false
    }));
  }
  document.querySelector("#subject").onclick = () => observed.activations++;
</script>`;

const observations = [];
for (const [engine, browserType] of Object.entries({ chromium, firefox, webkit })) {
  const browser = await browserType.launch({ headless: true });
  try {
    for (const placement of ["top-level", "iframe"]) {
      for (const arm of ["healthy", "other-page-pointer", "restore-frame-focus"]) {
        const contextA = await browser.newContext();
        const contextB = await browser.newContext();
        try {
          contextA.setDefaultTimeout(5_000);
          contextB.setDefaultTimeout(5_000);
          const pageA = await contextA.newPage();
          const pageB = await contextB.newPage();
          await pageB.setContent("<button>Other page</button>");
          let frame = pageA.mainFrame();
          if (placement === "iframe") {
            await pageA.setContent("<iframe></iframe>");
            await pageA.locator("iframe").evaluate((element, html) => { element.srcdoc = html; }, subjectHTML);
            frame = await (await pageA.locator("iframe").elementHandle()).contentFrame();
            assert(frame, "fixture iframe exists");
          } else {
            await pageA.setContent(subjectHTML);
          }
          await frame.locator("#subject").click();
          await frame.evaluate(() => {
            observed.blurs = 0;
            observed.activations = 0;
            observed.events = [];
          });
          const read = () => frame.evaluate(() => ({ ...observed,
            focused: document.hasFocus(), active_subject: document.activeElement?.id === "subject" }));
          const before = await read();
          assert(before.focused && before.active_subject, "native fixture initially owns focus");
          if (arm !== "healthy") await pageB.mouse.click(10, 10);
          // Same settling interval as the upstream repro, not a product budget.
          await pageA.waitForTimeout(200);
          const interrupted = await read();
          if (arm === "restore-frame-focus") await frame.evaluate(() => window.focus());
          // Preserve the installed Vitest provider's conditional iframe focus.
          await frame.evaluate(() => {
            if (!document.activeElement || document.activeElement === document.body ||
                document.activeElement.ownerDocument !== document) window.focus();
          });
          await pageA.keyboard.down("Enter");
          await pageA.keyboard.up("Enter");
          const after = await read();
          const row = { engine, version: browser.version(), placement, arm, before, interrupted, after };
          observations.push(row);
          console.log("iframe-focus-probe", JSON.stringify(row));
        } finally {
          await contextA.close();
          await contextB.close();
        }
      }
    }
  } finally { await browser.close(); }
}
assert.equal(observations.length, 18);
console.log("iframe-focus-probe-complete", JSON.stringify({ cases: observations.length,
  native_activation_misses: observations.filter(row => row.after.activations !== 1).length }));
