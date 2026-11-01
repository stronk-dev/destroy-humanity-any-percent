import './prepare-first-hour-certified-observation.mjs';
import {readFileSync,writeFileSync,lstatSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {execFileSync} from 'node:child_process';

const overlay=JSON.parse(readFileSync('.cache/first-hour-certified-overlay.json','utf8'));
const input=readFileSync('.cache/first-hour-lifecycle-runner.go','utf8');
const start=input.indexOf('func (runtime *firstHourRuntime) commandApplies('), end=input.indexOf('func (runtime *firstHourRuntime) manualRequest(');
if(start<0||end<=start)throw Error('legal probe region drift');
const body=input.slice(start,end);
const edits=[
 ['now time.Time) bool {','now time.Time) bool {\n\tmanualFirstHourLegalStartResearch(runtime)'],
 ['\t\treturn false','\t\tmanualFirstHourLegalCloneFailureResearch(runtime)\n\t\treturn false'],
 ['return err == nil && result.Decision.Outcome == save.IntentApplied','return manualFirstHourLegalResultResearch(runtime, result, err)'],
];
let generated=body;
for(const [before,after]of edits){if(generated.split(before).length!==2)throw Error('legal observer anchor drift');generated=generated.replace(before,after);}
let reversed=generated;for(const[before,after]of [...edits].reverse()){if(reversed.split(after).length!==2)throw Error('legal observer reverse drift');reversed=reversed.replace(after,before);}
if(reversed!==body)throw Error('nonreversible legal observation');
function write(path,data){try{if(!lstatSync(path).isFile()||lstatSync(path).isSymbolicLink())throw Error('unsafe population output');}catch(e){if(e.code!=='ENOENT')throw e;}writeFileSync(path,data);}
const runner=input.slice(0,start)+generated+input.slice(end);
write('.cache/first-hour-population-runner.go',runner);
overlay.Replace['harness/first_hour_runner.go']='../.cache/first-hour-population-runner.go';
overlay.Replace['harness/manual_first_hour_population.go']='../planning/clout-v1-and-pr-interns/probes/owned-first-hour-population.go.txt';
overlay.Replace['harness/manual_first_hour_population_test.go']='../planning/clout-v1-and-pr-interns/probes/owned-first-hour-population-test.go.txt';
write('.cache/first-hour-population-overlay.json',JSON.stringify(overlay,null,2)+'\n');
const files={};
for(const path of [...new Set([
 'server/harness/first_hour_runner.go','server/harness/relevance_solver.go','server/harness/harness.go','server/harness/reputation_career.go','server/harness/tier2_pacing.go',
 '.cache/first-hour-population-overlay.json',
 ...Object.values(overlay.Replace).map(path=>path.replace(/^\.\.\//,'')),
])].sort())files[path]=createHash('sha256').update(readFileSync(path)).digest('hex');
write('.cache/first-hour-population-sources.json',JSON.stringify({head:execFileSync('git',['rev-parse','HEAD'],{encoding:'utf8'}).trim(),files},null,2)+'\n');
console.log(JSON.stringify({entries:Object.keys(overlay.Replace).length,input:createHash('sha256').update(input).digest('hex'),generated:createHash('sha256').update(runner).digest('hex'),replacements:edits.length}));
