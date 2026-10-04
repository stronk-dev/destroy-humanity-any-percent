import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";
import { cssDependencyGraph } from "./tools/css-dependency-graph.mjs";

const dependencyPath = (relative: string) => decodeURIComponent(new URL(relative, import.meta.url).pathname);
const clientRoot = decodeURIComponent(new URL(".", import.meta.url).pathname);
const licenseGraph = cssDependencyGraph(clientRoot);

// Notations 1.x publishes an ESM build but points its package entry at a legacy
// extensionless break_infinity subpath. Keep both build and tests on the
// project's pinned Decimal implementation through an explicit bundler alias.

export default defineConfig({
  plugins: [svelte(), licenseGraph],
  resolve: { alias: {
    "@antimatter-dimensions/notations": dependencyPath("./node_modules/@antimatter-dimensions/notations/dist/ad-notations.esm.js"),
    "break_infinity.js/break_infinity": dependencyPath("./node_modules/break_infinity.js/dist/break_infinity.esm.js"),
  } },
  ssr: { noExternal: ["@antimatter-dimensions/notations"] },
  build: { sourcemap: true },
  worker: { format: "es", plugins: () => [licenseGraph.worker()] },
});
