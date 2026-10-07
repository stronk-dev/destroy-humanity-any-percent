import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

// Snapshot the just-built bytes before serving them; a dev server or a stale
// bundle must not count as a production-client journey.
export function productionClientFiles(dist) {
  const files = new Map([["/", readFileSync(path.join(dist, "index.html"))]]);
  for (const name of readdirSync(path.join(dist, "assets"))) {
    if (/\.(js|css)$/u.test(name)) files.set(`/assets/${name}`, readFileSync(path.join(dist, "assets", name)));
  }
  return files;
}

export function productionClientProof(files) {
  const html = files.get("/")?.toString("utf8") ?? "";
  const entry = /<script\b[^>]*\bsrc="([^"]+)"/u.exec(html)?.[1];
  const styles = [...html.matchAll(/<link\b[^>]*\bhref="([^"]+\.css)"/gu)].map((match) => match[1]);
  const workers = [...files.keys()].filter((name) => /^\/assets\/prediction\.worker-[^/]+\.js$/u.test(name));
  if (!entry?.startsWith("/assets/") || !files.has(entry) || styles.length === 0 || workers.length !== 1 ||
      html.includes("/@vite/client") || styles.some((name) => !files.has(name))) {
    throw new Error("production client build lacks its bundled entry, styles or prediction worker");
  }
  const required = new Set(["/", entry, ...styles, ...workers]);
  const observed = new Set();
  let workerStarted = false;
  return {
    response(route, status, body) {
      if (route.startsWith("/api/") || route === "/favicon.ico") return;
      // Reloads may revalidate a cached response. A bodyless 304 is usable
      // only after this browser context received and verified its 200 bytes.
      if (status === 304 && observed.has(route)) return;
      const expected = files.get(route);
      if (!expected || status !== 200 || !Buffer.from(body).equals(expected)) {
        throw new Error(`browser did not receive the built client bytes: ${route} HTTP${status}`);
      }
      observed.add(route);
    },
    worker(route) {
      if (route !== workers[0]) throw new Error(`browser started a non-bundled prediction worker: ${route}`);
      workerStarted = true;
    },
    finish() {
      const missing = [...required].filter((name) => !observed.has(name));
      if (missing.length > 0 || !workerStarted) {
        throw new Error(`production client browser proof incomplete: missing=${JSON.stringify(missing)}, worker_started=${workerStarted}`);
      }
      return [...observed].sort();
    },
  };
}
