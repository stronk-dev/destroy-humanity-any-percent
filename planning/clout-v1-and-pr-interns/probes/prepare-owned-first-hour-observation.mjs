import './prepare-owned-candidate-observation.mjs';
import {createHash} from 'node:crypto';
import {readFileSync,writeFileSync,lstatSync} from 'node:fs';
import {resolve} from 'node:path';

const root=process.cwd();
const path='server/harness/first_hour_runner.go';
const original=readFileSync(resolve(root,path),'utf8');
const edits=[
  ['founderAttended := runtime.founderAttendedMS + attended', 'founderAttended := runtime.founderAttendedMS + attended\n\tif err := manualFirstHourBoundaryResearch(runtime, attended, founderAttended); err != nil { return err }'],
  ['prestigecore.RecordOfflineSpan(runtime.company, runtime.company.EvaluatedThrough, now, runtime.suite.Bundle.Prestige.CatchupCeilingMS)', 'manualFirstHourGapResearch(runtime, now)'],
  ['production.SimulateAdvance(runtime.company, runtime.suite.Bundle.Economy,\n\t\tproduction.SimulationDependencies{Routes: runtime.suite.Bundle.Routes, Hook: runtime.lifetimeHook()}, runtime.companyRevision(), mode, now, runtime.external(), production.AblationMask{})', 'manualFirstHourAdvanceResearch(runtime, mode, now)'],
  ['production.SimulateAdvance(runtime.company, runtime.suite.Bundle.Economy,\n\t\t\tproduction.SimulationDependencies{Routes: runtime.suite.Bundle.Routes, Hook: runtime.lifetimeHook()}, runtime.companyRevision(), production.ModeOnline,\n\t\t\trelevanceNow(bankAtMS), runtime.external(), production.AblationMask{})', 'manualFirstHourAdvanceResearch(runtime, production.ModeOnline, relevanceNow(bankAtMS))'],
  ['runtime.company, runtime.revision = candidate.State, candidate.Revision', 'if err := manualFirstHourSelectResearch(runtime, ranker, candidate); err != nil { return err }'],
  ['return &RelevanceSuite{Scenario: RelevanceScenario{HorizonMS: runtime.spec.HorizonMS,', 'ranker := &RelevanceSuite{Scenario: RelevanceScenario{HorizonMS: runtime.spec.HorizonMS,'],
  ['external: append([]multiplier.Contribution(nil), runtime.external()...), hook: runtime.lifetimeHook()}, nil', 'external: append([]multiplier.Contribution(nil), runtime.external()...), hook: runtime.lifetimeHook()}\n\treturn manualFirstHourRankerResearch(runtime, ranker)'],
  ['production.SimulateTransition(request, runtime.company, runtime.suite.Bundle.Economy,\n\t\tproduction.SimulationDependencies{Routes: runtime.suite.Bundle.Routes, Hook: runtime.lifetimeHook()}, runtime.companyRevision(), mode, now, runtime.external(), nil, production.AblationMask{})', 'manualFirstHourTransitionResearch(runtime, request, now, mode)'],
];
let generated=original;
for(const [before,after] of edits){if(generated.split(before).length!==2)throw Error(`first-hour anchor drift: ${before}`);generated=generated.replace(before,after);}
let restored=generated;
for(const [before,after] of [...edits].reverse()){if(restored.split(after).length!==2)throw Error('first-hour reverse anchor drift');restored=restored.replace(after,before);}
if(restored!==original)throw Error('nonreversible first-hour transformation');
const overlay=JSON.parse(readFileSync(resolve(root,'.cache/owned-candidate-overlay.json'),'utf8'));
overlay.Replace['harness/first_hour_runner.go']='../.cache/owned-first-hour-runner.go';
overlay.Replace['harness/manual_owned_first_hour_bridge.go']='../planning/clout-v1-and-pr-interns/probes/owned-first-hour-bridge.go.txt';
overlay.Replace['harness/manual_owned_first_hour_test.go']='../planning/clout-v1-and-pr-interns/probes/owned-first-hour-test.go.txt';
for(const [destination,data] of [['.cache/owned-first-hour-runner.go',generated],['.cache/owned-first-hour-overlay.json',JSON.stringify(overlay,null,2)+'\n']]){
  const target=resolve(root,destination);
  try{if(!lstatSync(target).isFile()||lstatSync(target).isSymbolicLink())throw Error('unsafe generated path');}catch(error){if(error.code!=='ENOENT')throw error;}
  writeFileSync(target,data);
}
const sha=value=>createHash('sha256').update(value).digest('hex');
console.log(JSON.stringify({source:path,original:sha(original),generated:sha(generated),replacements:edits.length,entries:Object.keys(overlay.Replace).length}));
