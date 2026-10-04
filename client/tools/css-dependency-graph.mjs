import { createHash } from "node:crypto";
import { existsSync, readFileSync, realpathSync, writeFileSync } from "node:fs";
import { dirname, join, relative, resolve, sep } from "node:path";

export const cssDependencyGraphFile = "css-dependency-graph.json";

export function cssDependencyGraph(clientRoot) {
  const root = realpathSync(resolve(clientRoot));
  const cssAssets = new Set();
  const packageModules = new Set();
  const packageAssets = new Set();
  async function scanBundle(plugin, bundle) {
    const visited = new Set();
    function packagePath(source) {
      if (!source.includes("node_modules" + sep)) return null;
      const path = relative(root, source).split(sep).join("/");
      if (path.startsWith("../") || path.startsWith("/") || !path.includes("node_modules/")) {
        plugin.error(`CSS package resource outside client root: ${source}`);
      }
      return path;
    }
    async function visitStylesheet(id) {
      const source = id.split("?", 1)[0];
      if (visited.has(source)) return;
      visited.add(source);
      const modulePath = packagePath(source);
      if (modulePath) packageModules.add(modulePath);
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
      const urls = [...text.matchAll(/url\(\s*(?:"([^"]+)"|'([^']+)'|([^\s)]+))\s*\)/g)];
      if ((text.match(/url\s*\(/g) ?? []).length !== urls.length) {
        plugin.error(`unsupported CSS URL syntax: ${id}`);
      }
      for (const match of urls) {
        const specifier = match[1] ?? match[2] ?? match[3];
        if (specifier.startsWith("data:") || specifier.startsWith("#")) continue;
        if (/^[a-z][a-z0-9+.-]*:/i.test(specifier)) {
          plugin.error(`external CSS URL is not a bundled resource: ${specifier}`);
        }
        const resolved = await plugin.resolve(specifier, source);
        const file = resolved?.id.split("?", 1)[0] ?? resolve(dirname(source), specifier);
        if (!existsSync(file)) plugin.error(`unresolved CSS URL: ${specifier} in ${id}`);
        const resourcePath = packagePath(file);
        if (resourcePath) packageAssets.add(resourcePath);
      }
    }
    for (const output of Object.values(bundle)) {
      if (output.type === "asset" && output.fileName.endsWith(".css")) {
        cssAssets.add(output.fileName);
      }
      if (output.type !== "chunk") continue;
      for (const id of Object.keys(output.modules)) {
        const source = id.split("?", 1)[0];
        if (source.endsWith(".css") || id.includes("?svelte&type=style")) await visitStylesheet(id);
      }
    }
  }
  const main = {
    name: "cloud-clicker-css-dependency-graph",
    async generateBundle(_options, bundle) {
      await scanBundle(this, bundle);
    },
    writeBundle(options) {
      const dist = resolve(root, options.dir ?? "dist");
      const assets = [...cssAssets].sort().map((path) => ({
        path,
        sha256: createHash("sha256").update(readFileSync(join(dist, path))).digest("hex"),
      }));
      writeFileSync(join(dist, cssDependencyGraphFile),
        JSON.stringify({ schema_version: 2, assets, package_css_modules: [...packageModules].sort(),
          package_css_assets: [...packageAssets].sort() }) + "\n");
    },
  };
  main.worker = () => ({
    name: "cloud-clicker-worker-css-dependency-graph",
    async generateBundle(_options, bundle) {
      await scanBundle(this, bundle);
    },
  });
  return main;
}
