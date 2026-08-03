import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";

import { generatedOutputs, repositoryRoot } from "./copy-pipeline.mjs";

for (const [filename, bytes] of generatedOutputs()) {
  mkdirSync(path.dirname(filename), { recursive: true });
  writeFileSync(filename, bytes, "utf8");
  console.log(`generated ${path.relative(repositoryRoot, filename)}`);
}
