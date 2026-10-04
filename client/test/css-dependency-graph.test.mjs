import { createHash } from "node:crypto";
import { mkdtempSync, mkdirSync, readFileSync, realpathSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { afterEach, expect, test } from "vitest";
import { build } from "vite";
import { cssDependencyGraph, cssDependencyGraphFile } from "../tools/css-dependency-graph";

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
  expect(graph.schema_version).toBe(1);
  expect(graph.package_css_modules).toEqual(["node_modules/style-only/theme.css"]);
  expect(graph.assets).toHaveLength(1);
  expect(graph.assets[0].path).toMatch(/\.css$/);
  expect(graph.assets[0].sha256).toBe(createHash("sha256")
    .update(readFileSync(join(root, "dist", graph.assets[0].path))).digest("hex"));
});
