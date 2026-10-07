// Same DOM observer in the composed journey and native-browser negatives.
// "offer" is an interruption, never proof that the locked notice rendered.
export async function observeLockedPitchView({ budgetMs = 30_000 } = {}) {
  if (!Number.isFinite(budgetMs) || budgetMs <= 0 || budgetMs > 30_000) throw new Error("invalid locked Pitch observation budget");
  const visible = (node) => {
    if (!node?.isConnected) return false;
    for (let ancestor = node; ancestor; ancestor = ancestor.parentElement) {
      const style = getComputedStyle(ancestor);
      if (style.visibility !== "visible" || style.display === "none" || Number(style.opacity) === 0) return false;
    }
    const box = node.getBoundingClientRect();
    return box.width > 0 && box.height > 0;
  };
  const deadline = performance.now() + budgetMs;
  while (performance.now() < deadline) {
    const main = document.querySelector("main");
    if (main?.getAttribute("data-surface") === "offer_sheet" && visible(main.querySelector("#offer-heading"))) return "offer";
    if (main?.getAttribute("data-surface") === "minigame_session") {
      const notices = [...main.querySelectorAll(".minigame-session [role=status]")]
        .filter((node) => node.textContent?.trim() === "Locked. Unlock it with Fiscal credit first." && visible(node));
      if (notices.length > 1) throw new Error("ambiguous locked Pitch rejection notice");
      if (notices.length === 1) return "notice";
    }
    await new Promise(requestAnimationFrame);
  }
  throw new Error("locked Pitch rejection notice or visible offer did not appear");
}
