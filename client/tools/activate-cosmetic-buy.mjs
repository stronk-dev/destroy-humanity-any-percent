// Passed directly to page.evaluate; also exercised in native-browser tests.
// Only DOM globals: no runtime/API access and exactly one enabled activation.
export async function activateCosmeticBuy(expectedLabel) {
  const deadline = performance.now() + 30_000;
  while (performance.now() < deadline) {
    const item = document.querySelector('[data-testid="cosmetic-shelf"] [data-cosmetic="horse_armor"]');
    const main = item?.closest("main"), control = item?.querySelector("button");
    if (main?.getAttribute("data-surface") === "desk" && main.getAttribute("aria-busy") === "false" &&
        item?.getAttribute("data-state") === "unowned" && control?.isConnected && !control.disabled &&
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
  throw new Error("cosmetic AC14 Buy never became visibly actionable before the single DOM activation");
}
