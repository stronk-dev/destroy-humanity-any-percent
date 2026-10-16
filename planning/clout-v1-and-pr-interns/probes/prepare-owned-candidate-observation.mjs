import { createHash } from 'node:crypto';
import { lstatSync, readFileSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = process.cwd();
const sha = value => createHash('sha256').update(value).digest('hex');
const path = 'server/harness/relevance_solver.go';
const original = readFileSync(resolve(root, path), 'utf8');
const changes = [
  [
    'func (suite *RelevanceSuite) earliestAffordable(state *save.State, revision, nowMS, horizonMS int64, resourceID string, cost decimal.Decimal, mask production.AblationMask, counter *relevanceCounter) (int64, bool, error) {\n',
    'func (suite *RelevanceSuite) earliestAffordable(state *save.State, revision, nowMS, horizonMS int64, resourceID string, cost decimal.Decimal, mask production.AblationMask, counter *relevanceCounter) (int64, bool, error) {\n\tif binding := manualCandidateBindingsResearch[suite]; binding != nil { return manualCandidateEarliestResearch(binding, nowMS, horizonMS, resourceID, cost, counter) }\n',
  ],
  ['\tif earliest > nowMS {\n', '\tif earliest > nowMS && manualCandidateBindingsResearch[suite] == nil {\n'],
  [
    '\tsimulation, err := production.SimulateTransition(request, candidateAtBuy, suite.Catalog,\n\t\tproduction.SimulationDependencies{Routes: suite.Routes, Hook: suite.hook}, save.Revision{Number: revision, ConstantsHash: suite.ConstantsHash},\n\t\tproduction.ModeOnline, relevanceNow(earliest), suite.external, nil, mask)\n',
    '\tsimulation, err := manualCandidateTransitionResearch(suite, request, candidateAtBuy, revision, earliest, mask)\n',
  ],
  [
    '\trate, err := production.SimulateResourceRateWithContributions(state, suite.Catalog, suite.Scenario.Milestone.ResourceID, suite.external, mask)\n',
    '\trate, err := manualCandidateRateResearch(suite, state, mask)\n',
  ],
];
let generated = original;
for (const [before, after] of changes) {
  if (generated.split(before).length !== 2) throw Error('solver anchor drift');
  generated = generated.replace(before, after);
}
let reversed = generated;
for (const [before, after] of [...changes].reverse()) reversed = reversed.replace(after, before);
if (reversed !== original) throw Error('non-reversible solver transformation');
const overlay = JSON.parse(readFileSync(resolve(root, '.cache/wait-logged-overlay.json'), 'utf8'));
// This caller comparison does not use stage observation callbacks. Unlike the
// production-only H3s run, harness tests also build production as a NORMAL
// dependency, where _test.go callback declarations are deliberately unavailable.
for (const target of ['production/engine.go', 'production/intents.go', 'production/active_play.go']) delete overlay.Replace[target];
overlay.Replace['harness/relevance_solver.go'] = '../.cache/owned-candidate-solver.go';
overlay.Replace['harness/manual_owned_candidate_bridge.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-candidate-bridge.go.txt';
overlay.Replace['production/manual_owned_candidate_export_test.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-candidate-export.go.txt';
overlay.Replace['production/manual_owned_candidate_external_test.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-candidate-external.go.txt';
for (const [destination, data] of [
  ['.cache/owned-candidate-solver.go', generated],
  ['.cache/owned-candidate-overlay.json', `${JSON.stringify(overlay, null, 2)}\n`],
]) {
  const target = resolve(root, destination);
  try { if (!lstatSync(target).isFile() || lstatSync(target).isSymbolicLink()) throw Error('unsafe generated file'); }
  catch (error) { if (error.code !== 'ENOENT') throw error; }
  writeFileSync(target, data);
}
console.log(JSON.stringify({ source: path, original: sha(original), generated: sha(generated), replacements: changes.length, entries: Object.keys(overlay.Replace).length }));
