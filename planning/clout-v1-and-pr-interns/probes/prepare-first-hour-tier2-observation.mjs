import './prepare-first-hour-career-observation.mjs';
import {readFileSync, writeFileSync, lstatSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {execFileSync} from 'node:child_process';

const base='planning/clout-v1-and-pr-interns/probes/';
const path=base+'owned-first-hour-lifecycle.go.txt';
const original=readFileSync(path,'utf8');
const before='if !declared {\n\t\treturn nil, fmt.Errorf("first-hour spec outside pinned scenario")';
const after='if !declared && !manualTier2DerivedSourceResearch(r, scenario) {\n\t\treturn nil, fmt.Errorf("first-hour spec outside pinned scenario")';
if(original.split(before).length!==2)throw Error('Tier2 source anchor drift');
const generated=original.replace(before,after);
if(generated.replace(after,before)!==original)throw Error('nonreversible Tier2 source binding');
function write(path,data){try{if(!lstatSync(path).isFile()||lstatSync(path).isSymbolicLink())throw Error('unsafe Tier2 research output');}catch(e){if(e.code!=='ENOENT')throw e;}writeFileSync(path,data);}
write('.cache/first-hour-tier2-lifecycle.go',generated);
const overlay=JSON.parse(readFileSync('.cache/first-hour-career-overlay.json','utf8'));
overlay.Replace['harness/manual_first_hour_lifecycle.go']='../.cache/first-hour-tier2-lifecycle.go';
overlay.Replace['harness/manual_first_hour_tier2.go']='../'+base+'owned-first-hour-tier2.go.txt';
overlay.Replace['harness/manual_first_hour_tier2_test.go']='../'+base+'owned-first-hour-tier2-test.go.txt';
write('.cache/first-hour-tier2-overlay.json',JSON.stringify(overlay,null,2)+'\n');
const files={};
for(const p of [...new Set([
 'server/harness/first_hour_runner.go','server/harness/reputation_career.go','server/harness/harness.go','server/harness/relevance_solver.go','server/harness/tier2_pacing.go',
 '.cache/first-hour-tier2-overlay.json',
 ...Object.values(overlay.Replace).map(p=>p.replace(/^\.\.\//,'')),
])].sort())files[p]=createHash('sha256').update(readFileSync(p)).digest('hex');
write('.cache/first-hour-tier2-sources.json',JSON.stringify({head:execFileSync('git',['rev-parse','HEAD'],{encoding:'utf8'}).trim(),files},null,2)+'\n');
console.log(JSON.stringify({entries:Object.keys(overlay.Replace).length,original:createHash('sha256').update(original).digest('hex'),generated:createHash('sha256').update(generated).digest('hex')}));
