import './prepare-first-hour-population-observation.mjs';
import {readFileSync, writeFileSync, lstatSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {execFileSync} from 'node:child_process';

const input=readFileSync('.cache/first-hour-population-runner.go','utf8');
const edits=[
 ['\tboundaries, err := firstHourBoundaries(policy, seed, spec.HorizonMS)', '\tcleanup, setupErr := manualCareerBeginResearch(&runtime)\n\tif setupErr != nil { return failFirstHour(result, setupErr), commands, career, tier2 }\n\tdefer cleanup()\n\tboundaries, err := firstHourBoundaries(policy, seed, spec.HorizonMS)'],
 ['\t\tif career == nil && tier2 == nil && runtime.milestones["milestone.first_elective_exit"] != nil', '\t\tif err := manualCareerObserveResearch(&runtime, boundary); err != nil { return failFirstHour(result, err), commands, career, tier2 }\n\t\tif career == nil && tier2 == nil && runtime.milestones["milestone.first_elective_exit"] != nil'],
];
function replace(source,before,after){if(source.split(before).length!==2)throw Error('career observer anchor drift');return source.replace(before,after);}
let generated=input;
for(const [before,after]of edits)generated=replace(generated,before,after);
let reversed=generated;
for(const [before,after]of [...edits].reverse())reversed=replace(reversed,after,before);
if(reversed!==input)throw Error('nonreversible career observation');
function write(path,data){try{if(!lstatSync(path).isFile()||lstatSync(path).isSymbolicLink())throw Error('unsafe career output');}catch(e){if(e.code!=='ENOENT')throw e;}writeFileSync(path,data);}
write('.cache/first-hour-career-runner.go',generated);
const overlay=JSON.parse(readFileSync('.cache/first-hour-population-overlay.json','utf8'));
overlay.Replace['harness/first_hour_runner.go']='../.cache/first-hour-career-runner.go';
const base='../planning/clout-v1-and-pr-interns/probes/';
overlay.Replace['harness/manual_first_hour_career.go']=base+'owned-first-hour-career.go.txt';
overlay.Replace['harness/manual_first_hour_career_test.go']=base+'owned-first-hour-career-test.go.txt';
write('.cache/first-hour-career-overlay.json',JSON.stringify(overlay,null,2)+'\n');
const files={};
for(const path of [...new Set([
 'server/harness/first_hour_runner.go','server/harness/reputation_career.go','server/harness/harness.go','server/harness/relevance_solver.go','server/harness/tier2_pacing.go',
 '.cache/first-hour-career-overlay.json',
 ...Object.values(overlay.Replace).map(path=>path.replace(/^\.\.\//,'')),
])].sort())files[path]=createHash('sha256').update(readFileSync(path)).digest('hex');
write('.cache/first-hour-career-sources.json',JSON.stringify({head:execFileSync('git',['rev-parse','HEAD'],{encoding:'utf8'}).trim(),files},null,2)+'\n');
console.log(JSON.stringify({entries:Object.keys(overlay.Replace).length,original:createHash('sha256').update(input).digest('hex'),generated:createHash('sha256').update(generated).digest('hex'),edits:edits.length}));
