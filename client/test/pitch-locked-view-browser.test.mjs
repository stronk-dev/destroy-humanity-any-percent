import { expect, it } from "vitest";
import { observeLockedPitchView } from "../tools/pitch-locked-view.mjs";

const notice = "Locked. Unlock it with Fiscal credit first.";
function fixture(surface, html) {
  const main = document.createElement("main"); main.dataset.surface = surface;
  main.innerHTML = html; document.body.append(main); return main;
}

for (const arm of ["notice", "offer"]) it.skipIf(typeof document === "undefined")(`locked Pitch observer distinguishes visible ${arm}`, async () => {
  const main = fixture(arm === "notice" ? "minigame_session" : "offer_sheet", arm === "notice"
    ? `<section class="minigame-session"><p role="status">${notice}</p></section>` : '<h1 id="offer-heading">Offer</h1>');
  try { expect(await observeLockedPitchView()).toBe(arm); } finally { main.remove(); }
});

for (const arm of ["missing", "prefix", "hidden", "transparent-parent", "wrong-surface", "wrong-region", "missing-offer-heading", "hidden-offer"]) {
  it.skipIf(typeof document === "undefined")(`locked Pitch observer rejects ${arm} instead of accepting HTTP-only proof`, async () => {
    const offer = arm.includes("offer");
    const main = fixture(offer ? "offer_sheet" : arm === "wrong-surface" ? "desk" : "minigame_session",
      offer ? arm === "hidden-offer" ? '<h1 id="offer-heading" hidden>Offer</h1>' : ""
        : `<section class="${arm === "wrong-region" ? "other" : "minigame-session"}" ${arm === "transparent-parent" ? 'style="opacity:0"' : ""}><p role="status" ${arm === "hidden" ? "hidden" : ""}>${arm === "missing" ? "" : notice}${arm === "prefix" ? " extra" : ""}</p></section>`);
    try { await expect(observeLockedPitchView({ budgetMs: 80 })).rejects.toThrow("did not appear"); }
    finally { main.remove(); }
  });
}

it.skipIf(typeof document === "undefined")("locked Pitch observer rejects ambiguous notices and invalid expanded budgets", async () => {
  const main = fixture("minigame_session", `<section class="minigame-session"><p role="status">${notice}</p><p role="status">${notice}</p></section>`);
  try {
    await expect(observeLockedPitchView()).rejects.toThrow("ambiguous");
    for (const budgetMs of [0, -1, 30_001, Infinity, NaN]) await expect(observeLockedPitchView({ budgetMs })).rejects.toThrow("invalid");
  } finally { main.remove(); }
});
