// Passed directly to page.evaluate; also exercised in native-browser tests.
// Only DOM globals: no runtime/API access and exactly one enabled activation.
export async function activateCosmeticBuy(input) {
  const expectedLabel = typeof input === "string" ? input : input.label;
  const expectedState = typeof input === "string" ? "unowned" : input.state;
  const budgetMs = typeof input === "string" ? 30_000 : input.budgetMs;
  if (typeof expectedLabel !== "string" || expectedLabel.length === 0 ||
      !["unowned", "owned"].includes(expectedState) || !Number.isFinite(budgetMs) || budgetMs <= 0 || budgetMs > 30_000) {
    throw new Error("invalid cosmetic DOM activation boundary");
  }
  const deadline = performance.now() + budgetMs;
  while (performance.now() < deadline) {
    const item = document.querySelector('[data-testid="cosmetic-shelf"] [data-cosmetic="horse_armor"]');
    const main = item?.closest("main");
    const controls = [...item?.querySelectorAll("button") ?? []].filter((control) => control.textContent?.trim() === expectedLabel);
    if (controls.length > 1) throw new Error("ambiguous cosmetic DOM activation label");
    const control = controls[0];
    if (main?.getAttribute("data-surface") === "desk" && main.getAttribute("aria-busy") === "false" &&
        item?.getAttribute("data-state") === expectedState && control?.isConnected && !control.disabled &&
        control.textContent?.trim() === expectedLabel) {
      control.scrollIntoView({ block: "center", inline: "nearest" });
      const bounds = control.getBoundingClientRect(), style = getComputedStyle(control);
      const hit = document.elementFromPoint(bounds.x + bounds.width / 2, bounds.y + bounds.height / 2);
      if (bounds.width > 0 && bounds.height > 0 && style.display !== "none" && style.visibility === "visible" &&
          (hit === control || control.contains(hit))) {
        control.click();
        return;
      }
    }
    await new Promise(requestAnimationFrame);
  }
  throw new Error("cosmetic control never became visibly actionable before the single DOM activation");
}

// The real-service driver calls this on passive browser observations. An
// untrusted DOM click, missing keyboard input or truncated trace is not native
// activation proof, even if the server purchase happened to succeed.
export function assertNativeCosmeticBuyTrace({ events, dropped_events }, key, label) {
  const down = events.filter((event) => event.type === "keydown");
  const up = events.filter((event) => event.type === "keyup");
  const clicks = events.filter((event) => event.type === "click");
  const ready = (event) => event.trusted === true && event.focused === true &&
    event.button?.trim() === label && event.cosmetic === "horse_armor" && event.state === "unowned" &&
    event.connected === true && event.disabled === false && event.surface === "desk" && event.busy === "false";
  if (!["Enter", " "].includes(key) || dropped_events !== 0 || events.length !== 3 ||
      down.length !== 1 || up.length !== 1 || clicks.length !== 1 ||
      down[0].key !== key || up[0].key !== key || up[0].trusted !== true ||
      !ready(down[0]) || !ready(clicks[0]) || events[0] !== down[0] ||
      events[key === "Enter" ? 1 : 2] !== clicks[0]) {
    throw new Error(`cosmetic Buy lacks one native ${JSON.stringify(key)} activation: ${JSON.stringify({ events, dropped_events })}`);
  }
}
