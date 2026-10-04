import { createHash } from "node:crypto";
import { readFileSync, realpathSync, writeFileSync } from "node:fs";
import { join, relative, resolve, sep } from "node:path";

export const cssDependencyGraphFile = "css-dependency-graph.json";

export function cssDependencyGraph(clientRoot) {
  const root = realpathSync(resolve(clientRoot));
  let cssAssets = [];
  let packageModules = [];
  return {
    name: "cloud-clicker-css-dependency-graph",
    async generateBundle(_options, bundle) {
      const assets = [];
      const modules = new Set();
      const visited = new Set();
      const plugin = this;
      async function visitStylesheet(id) {
        const source = id.split("?", 1)[0];
        if (visited.has(source)) return;
        visited.add(source);
        if (source.includes("node_modules" + sep)) {
          const path = relative(root, source).split(sep).join("/");
          if (path.startsWith("../") || path.startsWith("/") || !path.includes("node_modules/")) {
            plugin.error(`CSS package module outside client root: ${id}`);
          }
          modules.add(path);
        }
        const text = readFileSync(source, "utf8").replace(/\/\*[\s\S]*?\*\//g, "");
        const imports = [...text.matchAll(/@import\s+(?:url\(\s*)?(?:"([^"]+)"|'([^']+)')\s*\)?/g)];
        if ((text.match(/@import\b/g) ?? []).length !== imports.length) {
          plugin.error(`unsupported CSS import syntax: ${id}`);
        }
        for (const match of imports) {
          const specifier = match[1] ?? match[2];
          const resolved = await plugin.resolve(specifier, source);
          if (!resolved || !resolved.id.split("?", 1)[0].endsWith(".css")) {
            plugin.error(`unresolved CSS import: ${specifier} in ${id}`);
          }
          await visitStylesheet(resolved.id);
        }
      }
      for (const output of Object.values(bundle)) {
        if (output.type === "asset" && output.fileName.endsWith(".css")) {
          assets.push(output.fileName);
        }
        if (output.type !== "chunk") continue;
        for (const id of Object.keys(output.modules)) {
          const source = id.split("?", 1)[0];
          if (source.endsWith(".css") || id.includes("?svelte&type=style")) await visitStylesheet(id);
        }
      }
      cssAssets = assets.sort();
      packageModules = [...modules].sort();
    },
    writeBundle(options) {
      const dist = resolve(root, options.dir ?? "dist");
      const assets = cssAssets.map((path) => ({
        path,
        sha256: createHash("sha256").update(readFileSync(join(dist, path))).digest("hex"),
      }));
      writeFileSync(join(dist, cssDependencyGraphFile),
        JSON.stringify({ schema_version: 1, assets, package_css_modules: packageModules }) + "\n");
    },
  };
}
