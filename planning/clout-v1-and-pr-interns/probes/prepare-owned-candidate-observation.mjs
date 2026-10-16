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
    'func (suite *RelevanceSuite) earliestAffordable(state *save.State, revision, nowMS, horizonMS int64, resourceID string, cost decimal.Decimal, mask production.AblationMask, counter *relevanceCounter) (int64, bool, error) {\n\tif scope := manualChoiceScopesResearch[suite]; scope != nil { return scope.earliest(state, revision, nowMS, horizonMS, resourceID, cost, counter) }\n\tif binding := manualCandidateBindingsResearch[suite]; binding != nil { return manualCandidateEarliestResearch(binding, nowMS, horizonMS, resourceID, cost, counter) }\n',
  ],
  ['\tif earliest > nowMS {\n', '\tif earliest > nowMS && manualCandidateBindingsResearch[suite] == nil && manualChoiceScopesResearch[suite] == nil {\n'],
  [
    '\tsimulation, err := production.SimulateTransition(request, candidateAtBuy, suite.Catalog,\n\t\tproduction.SimulationDependencies{Routes: suite.Routes, Hook: suite.hook}, save.Revision{Number: revision, ConstantsHash: suite.ConstantsHash},\n\t\tproduction.ModeOnline, relevanceNow(earliest), suite.external, nil, mask)\n',
    '\tsimulation, err := manualCandidateTransitionResearch(suite, request, candidateAtBuy, revision, earliest, mask)\n',
  ],
  [
    '\trate, err := production.SimulateResourceRateWithContributions(state, suite.Catalog, suite.Scenario.Milestone.ResourceID, suite.external, mask)\n',
    '\trate, err := manualCandidateRateResearch(suite, state, mask)\n',
  ],
  ['\tcacheKeys := []string{}\n', '\tcacheKeys := []manualCompletionIdentityResearch{}\n'],
  ['key, keyErr := relevanceCompletionCacheKey(state, revision, nowMS, maxDecisions-decision)', 'key, keyErr := manualCompletionKeyResearch(suite, state, revision, nowMS, maxDecisions-decision)'],
  ['cached, ok := cache[key]', 'cached, ok := manualCompletionGetResearch(suite, cache, key)'],
  ['cache[priorKey] = cached', 'manualCompletionSetResearch(suite, cache, priorKey, cached)'],
  ['cache[key] = entry', 'manualCompletionSetResearch(suite, cache, key, entry)'],
  ['\t\t\tif _, err := suite.advance(state, revision, advanceAt, mask, counter); err != nil {\n', '\t\t\tif err := manualAdvanceCompletionResearch(suite, state, revision, advanceAt, mask, counter); err != nil {\n', true],
  ['\t\tchosen := candidates[0]\n\t\tstate, revision, nowMS = chosen.State, chosen.Revision, chosen.AtMS\n', '\t\tchosen := candidates[0]\n\t\tif err := manualSelectedCompletionResearch(suite, state, chosen); err != nil { return relevanceRunResult{}, err }\n\t\tstate, revision, nowMS = chosen.State, chosen.Revision, chosen.AtMS\n', true],
  ['func (suite *RelevanceSuite) finishToMilestone(state *save.State, revision, nowMS int64, mask production.AblationMask, counter *relevanceCounter) (int64, *int64, []production.RoleActivation, error) {\n', 'func (suite *RelevanceSuite) finishToMilestone(state *save.State, revision, nowMS int64, mask production.AblationMask, counter *relevanceCounter) (int64, *int64, []production.RoleActivation, error) {\n\tif manualChoiceScopesResearch[suite] != nil { return manualFinishCompletionResearch(suite,state,revision,nowMS,mask,counter) }\n'],
];
let generated = original;
function replaceScoped(source, before, after, completionOnly) {
  const start = completionOnly ? source.indexOf('func (suite *RelevanceSuite) runRankedCompletionCached(') : 0;
  const end = completionOnly ? source.indexOf('\nfunc relevanceCompletionCacheKey(', start) : source.length;
  if (start < 0 || end < start) throw Error('completion region drift');
  const region = source.slice(start,end);
  if (region.split(before).length !== 2) throw Error(`solver anchor drift: ${before.slice(0,100)}`);
  return source.slice(0,start) + region.replace(before,after) + source.slice(end);
}
for (const [before, after, completionOnly] of changes) generated=replaceScoped(generated,before,after,completionOnly);
let reversed = generated;
for (const [before, after, completionOnly] of [...changes].reverse()) reversed=replaceScoped(reversed,after,before,completionOnly);
if (reversed !== original) throw Error('non-reversible solver transformation');
const clonePath = 'server/harness/harness.go';
const cloneOriginal = readFileSync(resolve(root, clonePath), 'utf8');
const cloneBefore = 'func cloneState(catalog *economy.Catalog, source *save.State) (*save.State, error) {\n\tencoded, err := save.EncodeState(source)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn save.RestoreState(encoded, save.VersionForState(source), catalog, economy.ScopeCompany, time.Time{})\n}';
const cloneAfter = cloneBefore.replace('\treturn save.RestoreState(encoded, save.VersionForState(source), catalog, economy.ScopeCompany, time.Time{})', '\trestored, err := save.RestoreState(encoded, save.VersionForState(source), catalog, economy.ScopeCompany, time.Time{})\n\tif err != nil { return nil, err }\n\tif err := manualCloneChoiceResearch(source, restored); err != nil { return nil, err }\n\treturn restored, nil');
if (cloneOriginal.split(cloneBefore).length !== 2) throw Error('clone anchor drift');
const cloneGenerated = cloneOriginal.replace(cloneBefore, cloneAfter);
if (cloneGenerated.replace(cloneAfter, cloneBefore) !== cloneOriginal) throw Error('non-reversible clone transformation');
const overlay = JSON.parse(readFileSync(resolve(root, '.cache/wait-logged-overlay.json'), 'utf8'));
// This caller comparison does not use stage observation callbacks. Unlike the
// production-only H3s run, harness tests also build production as a NORMAL
// dependency, where _test.go callback declarations are deliberately unavailable.
for (const target of ['production/engine.go', 'production/intents.go', 'production/active_play.go']) delete overlay.Replace[target];
overlay.Replace['harness/relevance_solver.go'] = '../.cache/owned-candidate-solver.go';
overlay.Replace['harness/harness.go'] = '../.cache/owned-candidate-harness.go';
overlay.Replace['harness/manual_owned_candidate_bridge.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-candidate-bridge.go.txt';
overlay.Replace['harness/manual_owned_choice_bridge.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-choice-bridge.go.txt';
overlay.Replace['harness/manual_owned_completion_bridge.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-completion-bridge.go.txt';
overlay.Replace['production/manual_owned_choice_export_test.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-choice-export.go.txt';
overlay.Replace['production/manual_owned_bank_export_test.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-bank-export.go.txt';
overlay.Replace['production/manual_owned_candidate_export_test.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-candidate-export.go.txt';
overlay.Replace['production/manual_owned_candidate_external_test.go'] = '../planning/clout-v1-and-pr-interns/probes/owned-candidate-external.go.txt';
for (const [destination, data] of [
  ['.cache/owned-candidate-solver.go', generated],
  ['.cache/owned-candidate-harness.go', cloneGenerated],
  ['.cache/owned-candidate-overlay.json', `${JSON.stringify(overlay, null, 2)}\n`],
]) {
  const target = resolve(root, destination);
  try { if (!lstatSync(target).isFile() || lstatSync(target).isSymbolicLink()) throw Error('unsafe generated file'); }
  catch (error) { if (error.code !== 'ENOENT') throw error; }
  writeFileSync(target, data);
}
console.log(JSON.stringify({ source: path, original: sha(original), generated: sha(generated), replacements: changes.length,
  clone_original: sha(cloneOriginal), clone_generated: sha(cloneGenerated), entries: Object.keys(overlay.Replace).length }));
