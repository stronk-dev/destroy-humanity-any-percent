import { expect, it } from "vitest";

const browser = typeof document !== "undefined";

// CI D2: the same test runs in three browser projects. Neither its screenshot
// source nor Vitest's path-hashed attachment may be shared between engines.
it.skipIf(!browser)("keeps native screenshot sources and attachments in their browser namespace", async ({ annotate }) => {
  const { page, server } = await import("vitest/browser");
  const marker = document.createElement("main");
  marker.textContent = `Artifact isolation: ${server.browser}`;
  document.body.append(marker);
  try {
    const screenshot = await page.screenshot({ base64: true });
    expect(screenshot.path.replaceAll("\\", "/")).toContain(`/__screenshots__/${server.browser}/`);
    expect(screenshot.base64.startsWith("iVBORw0KGgo")).toBe(true);
    // Exercise the actual runner's filesystem copy, not only path construction.
    await annotate(`Native ${server.browser} screenshot`, {
      path: screenshot.path, contentType: "image/png",
    });
  } finally { marker.remove(); }
});
