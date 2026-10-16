import { createHash } from 'node:crypto';
import { lstatSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';

// Run from the repository root. Generated source is ignored, observation-only,
// and used solely by this manual Go overlay. No production file is edited.
const root = process.cwd();
readFileSync(resolve(root, 'server/go.mod'));
const cache = resolve(root, '.cache');
mkdirSync(cache, { recursive: true });
if (!lstatSync(cache).isDirectory() || lstatSync(cache).isSymbolicLink()) throw Error('unsafe cache directory');
const sha = data => createHash('sha256').update(data).digest('hex');
const overlay = JSON.parse(readFileSync(resolve(root, 'planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json'), 'utf8'));
const transforms = [
  {
    source: 'server/production/engine.go', output: 'wait-logged-engine.go', target: 'production/engine.go',
    pattern: /^(func evaluateWithSimulationPolicy\([^\n]+\)) \(EvaluationResult, error\) \{\n/m,
    returns: '(researchResult EvaluationResult, researchError error)',
    observation: '\tif observeLoggedEvaluationResearch != nil {\n\t\tfinish := observeLoggedEvaluationResearch(state, now, mode, contributions)\n\t\tdefer func() { finish(researchResult, researchError) }()\n\t}\n',
  },
  {
    source: 'server/production/intents.go', output: 'wait-logged-intents.go', target: 'production/intents.go',
    pattern: /^(func runAccrualHook\([^\n]+\)) \(\[\]save.EventWrite, error\) \{\n/m,
    returns: '(researchEvents []save.EventWrite, researchError error)',
    observation: '\tif observeLoggedHookResearch != nil {\n\t\tfinish := observeLoggedHookResearch(state, result, contributions)\n\t\tdefer func() { finish(researchEvents, researchError) }()\n\t}\n',
  },
  {
    source: 'server/production/active_play.go', output: 'wait-logged-active.go', target: 'production/active_play.go',
    pattern: /^(func applyActivePlaySchedule\([^\n]+\)) \(\[\]save.EventWrite, error\) \{\n/m,
    returns: '(researchEvents []save.EventWrite, researchError error)',
    observation: '\tif observeLoggedScheduleResearch != nil {\n\t\tfinish := observeLoggedScheduleResearch(state, now)\n\t\tdefer func() { finish(researchEvents, researchError) }()\n\t}\n',
  },
];
const manifest = {};
for (const row of transforms) {
  const original = readFileSync(resolve(root, row.source), 'utf8');
  const matches = [...original.matchAll(new RegExp(row.pattern.source, 'gm'))];
  if (matches.length !== 1) throw Error(`signature drift: ${row.source}`);
  const insertion = `${matches[0][1]} ${row.returns} {\n${row.observation}`;
  const generated = original.replace(row.pattern, insertion);
  // Removing precisely the observation and named return slots recovers every
  // original byte, including the entire arithmetic/control-flow body.
  if (generated.replace(insertion, matches[0][0]) !== original) throw Error('non-observational transformation');
  const path = resolve(cache, row.output);
  try { if (!lstatSync(path).isFile() || lstatSync(path).isSymbolicLink()) throw Error('unsafe generated source'); }
  catch (error) { if (error.code !== 'ENOENT') throw error; }
  writeFileSync(path, generated);
  overlay.Replace[row.target] = `../.cache/${row.output}`;
  manifest[row.source] = { original: sha(original), generated: sha(generated) };
}
overlay.Replace['production/manual_wait_logged_stages_test.go'] = '../planning/clout-v1-and-pr-interns/probes/wait-logged-stages.go.txt';
const output = `${JSON.stringify(overlay, null, 2)}\n`;
const path = resolve(cache, 'wait-logged-overlay.json');
try { if (!lstatSync(path).isFile() || lstatSync(path).isSymbolicLink()) throw Error('unsafe generated overlay'); }
catch (error) { if (error.code !== 'ENOENT') throw error; }
writeFileSync(path, output);
const config = `import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { mergeConfig } from "../client/node_modules/vite/dist/node/index.js";
import base from "../client/vite.config.ts";
const slot = fileURLToPath(new URL("../client/test/replay.test.ts", import.meta.url));
const source = fileURLToPath(new URL("../planning/clout-v1-and-pr-interns/probes/wait-logged-replay.test.ts.txt", import.meta.url));
export default mergeConfig(base, { plugins: [{ name: "manual-wait-logged-replay", enforce: "pre", load(id) {
  if (id.split("?")[0] === slot) return readFileSync(source, "utf8");
} }], test: { include: ["test/replay.test.ts"] } });
`;
const configPath = resolve(cache, 'wait-logged-replay.config.mjs');
try { if (!lstatSync(configPath).isFile() || lstatSync(configPath).isSymbolicLink()) throw Error('unsafe generated config'); }
catch (error) { if (error.code !== 'ENOENT') throw error; }
writeFileSync(configPath, config);
console.log(JSON.stringify({ manifest, overlay_sha256: sha(output), entries: Object.keys(overlay.Replace).length, client_config_sha256: sha(config) }, null, 2));
