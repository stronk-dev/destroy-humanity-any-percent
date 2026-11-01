import './prepare-owned-first-hour-observation.mjs';
import {readFileSync,writeFileSync,lstatSync} from 'node:fs';
import {createHash} from 'node:crypto';

const overlay=JSON.parse(readFileSync('.cache/owned-first-hour-overlay.json','utf8'));
const sources=[
  ['server/harness/first_hour_runner.go','.cache/owned-first-hour-runner.go','.cache/first-hour-lifecycle-runner.go',[
    ['func (runtime *firstHourRuntime) step(boundary firstHourBoundary) error {','func (runtime *firstHourRuntime) manualFirstHourStepBodyResearch(boundary firstHourBoundary) error {'],
    ['func (runtime *firstHourRuntime) applyEnding(now time.Time, wallMS, attended int64) error {','func (runtime *firstHourRuntime) manualFirstHourEndingBodyResearch(now time.Time, wallMS, attended int64) error {'],
    ['runtime.observeRunEnded(boundary.atMS, "collapse")\n\t\treturn nil','runtime.observeRunEnded(boundary.atMS, "collapse")\n\t\treturn manualFirstHourTerminalResearch(runtime)'],
  ]],
  ['server/harness/reputation_career.go','server/harness/reputation_career.go','.cache/first-hour-lifecycle-career.go',[
    ['func (runtime *firstHourRuntime) applyCareerExit(now time.Time, wallMS, attended int64) error {','func (runtime *firstHourRuntime) manualFirstHourCareerExitBodyResearch(now time.Time, wallMS, attended int64) error {'],
  ]],
  ['server/harness/tier2_pacing.go','server/harness/tier2_pacing.go','.cache/first-hour-lifecycle-tier2.go',[
    ['func (runtime *firstHourRuntime) applyTier2Exit(now time.Time, wallMS, attended int64) error {','func (runtime *firstHourRuntime) manualFirstHourTier2ExitBodyResearch(now time.Time, wallMS, attended int64) error {'],
  ]],
];
const sha=value=>createHash('sha256').update(value).digest('hex');
const records=[];
function safeWrite(path,data){try{if(!lstatSync(path).isFile()||lstatSync(path).isSymbolicLink())throw Error('unsafe generated path');}catch(error){if(error.code!=='ENOENT')throw error;}writeFileSync(path,data);}
for(const [destination,input,output,edits]of sources){
  const source=readFileSync(input,'utf8');let generated=source;
  for(const [before,after]of edits){if(generated.split(before).length!==2)throw Error(`lifecycle anchor drift: ${before}`);generated=generated.replace(before,after);}
  let reversed=generated;for(const [before,after]of [...edits].reverse()){if(reversed.split(after).length!==2)throw Error('lifecycle reverse drift');reversed=reversed.replace(after,before);}
  if(reversed!==source)throw Error('nonreversible lifecycle transformation');
  safeWrite(output,generated);overlay.Replace[destination.replace('server/','')]='../'+output;
  records.push({source:destination,original:sha(source),generated:sha(generated),replacements:edits.length});
}
overlay.Replace['harness/manual_first_hour_lifecycle.go']='../planning/clout-v1-and-pr-interns/probes/owned-first-hour-lifecycle.go.txt';
overlay.Replace['harness/manual_first_hour_lifecycle_test.go']='../planning/clout-v1-and-pr-interns/probes/owned-first-hour-lifecycle-test.go.txt';
safeWrite('.cache/first-hour-lifecycle-overlay.json',JSON.stringify(overlay,null,2)+'\n');
console.log(JSON.stringify({records,entries:Object.keys(overlay.Replace).length}));
