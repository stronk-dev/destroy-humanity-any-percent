import { createHash } from "node:crypto";
import { mkdtempSync, mkdirSync, readFileSync, realpathSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { afterEach, expect, test } from "vitest";
import { build } from "vite";
import { cssDependencyGraph, cssDependencyGraphFile } from "../tools/css-dependency-graph.mjs";

const roots = [];
afterEach(() => {
  for (const root of roots.splice(0)) rmSync(root, { recursive: true, force: true });
});

function write(path, content) {
  mkdirSync(dirname(path), { recursive: true });
  writeFileSync(path, content);
}

test.each([
  ["JavaScript style import", `import "style-only/theme.css";\nconsole.log("fixture");\n`, ""],
  ["nested stylesheet import", `import "./local.css";\nconsole.log("fixture");\n`, `@import "style-only/theme.css";\n`],
])("the production Vite graph records a CSS-only npm package via %s", async (_name, entry, localCSS) => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), "cloud-clicker-css-graph-")));
  roots.push(root);
  write(join(root, "index.html"), `<script type="module" src="/src/main.ts"></script>`);
  write(join(root, "src", "main.ts"), entry);
  if (localCSS) write(join(root, "src", "local.css"), localCSS);
  write(join(root, "node_modules", "style-only", "package.json"), `{"name":"style-only","version":"1.0.0","license":"MIT"}`);
  write(join(root, "node_modules", "style-only", "theme.css"), `body { color: red; }\n`);
  await build({ root, configFile: false, logLevel: "silent", plugins: [cssDependencyGraph(root)],
    build: { outDir: "dist", sourcemap: true } });

  const graph = JSON.parse(readFileSync(join(root, "dist", cssDependencyGraphFile), "utf8"));
  expect(graph.schema_version).toBe(2);
  expect(graph.package_css_modules).toEqual(["node_modules/style-only/theme.css"]);
  expect(graph.assets).toHaveLength(1);
  expect(graph.assets[0].path).toMatch(/\.css$/);
  expect(graph.assets[0].sha256).toBe(createHash("sha256")
    .update(readFileSync(join(root, "dist", graph.assets[0].path))).digest("hex"));
});

test("CSS URL package resources are attributed when their bytes ship", async () => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), "cloud-clicker-css-url-")));
  roots.push(root);
  write(join(root, "index.html"), `<script type="module" src="/src/main.ts"></script>`);
  write(join(root, "src", "main.ts"), `import "./local.css";\n`);
  write(join(root, "src", "local.css"), `body { background: url("../node_modules/style-only/logo.svg"); }\n`);
  write(join(root, "node_modules", "style-only", "package.json"), `{"name":"style-only","version":"1.0.0","license":"MIT"}`);
  write(join(root, "node_modules", "style-only", "logo.svg"), `<svg xmlns="http://www.w3.org/2000/svg"><text>fixture-logo</text></svg>`);
  await build({ root, configFile: false, logLevel: "silent", plugins: [cssDependencyGraph(root)],
    build: { outDir: "dist", sourcemap: true } });
  const graph = JSON.parse(readFileSync(join(root, "dist", cssDependencyGraphFile), "utf8"));
  expect(graph.assets).toHaveLength(1);
  const css = readFileSync(join(root, "dist", graph.assets[0].path), "utf8");
  const inline = css.match(/data:image\/svg\+xml;base64,([A-Za-z0-9+/=]+)/);
  expect(inline).not.toBeNull();
  expect(Buffer.from(inline[1], "base64").toString("utf8")).toContain("fixture-logo");
  expect(graph.package_css_assets).toContain("node_modules/style-only/logo.svg");
});

test("worker CSS package is recorded by the worker build graph when source maps omit it", async () => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), "cloud-clicker-worker-css-")));
  roots.push(root);
  write(join(root, "index.html"), `<script type="module" src="/src/main.ts"></script>`);
  write(join(root, "src", "main.ts"), `new Worker(new URL("./worker.ts", import.meta.url), { type: "module" });\n`);
  write(join(root, "src", "worker.ts"), `import "style-only/theme.css";\npostMessage("ready");\n`);
  write(join(root, "node_modules", "style-only", "package.json"), `{"name":"style-only","version":"1.0.0","license":"MIT"}`);
  write(join(root, "node_modules", "style-only", "theme.css"), `body::after { content: "worker-style-sentinel"; }\n`);
  const graphPlugin = cssDependencyGraph(root);
  await build({ root, configFile: false, logLevel: "silent", plugins: [graphPlugin],
    build: { outDir: "dist", sourcemap: true }, worker: { format: "es", plugins: () => [graphPlugin.worker()] } });
  const graph = JSON.parse(readFileSync(join(root, "dist", cssDependencyGraphFile), "utf8"));
  const assetFiles = readdirSync(join(root, "dist", "assets"));
  const shipped = assetFiles.some((name) => readFileSync(join(root, "dist", "assets", name)).includes("worker-style-sentinel"));
  expect(shipped).toBe(true);
  const mapSources = assetFiles.filter((name) => name.endsWith(".map"))
    .flatMap((name) => JSON.parse(readFileSync(join(root, "dist", "assets", name), "utf8")).sources);
  expect(mapSources.some((source) => source.includes("node_modules/style-only/theme.css"))).toBe(false);
  expect(graph.package_css_modules).toContain("node_modules/style-only/theme.css");
});
