import {readFileSync,writeFileSync,lstatSync} from 'node:fs';
const base=JSON.parse(readFileSync('.cache/first-hour-population-overlay.json','utf8'));
const path='planning/clout-v1-and-pr-interns/probes/owned-first-hour-population-test.go.txt';
const original=readFileSync(path,'utf8');
function replaceOne(source,before,after){if(source.split(before).length!==2)throw Error('census control anchor drift');return source.replace(before,after);}
function write(path,data){try{if(!lstatSync(path).isFile()||lstatSync(path).isSymbolicLink())throw Error('unsafe control output');}catch(e){if(e.code!=='ENOENT')throw e;}writeFileSync(path,data);}
let omitted=replaceOne(original,'row.Failure = "source/owner construction"','row.Failure = "source/owner construction"\n\tif row.Policy != "" { return }');
// The real observations do not run. A valid complete census must REFUSE this.
// Drop only the source-manifest byte check for this explicitly mutated test
// source; otherwise it would fail source identity before testing the census.
omitted=replaceOne(omitted,'if err != nil || populationDigestResearch(data) != want {','if err != nil || (path != "'+path+'" && populationDigestResearch(data) != want) {');
const corrected=omitted;
const before='if row.Complete {\n\t\t\t\tcomplete++\n\t\t\t\tactualRuns = append(actualRuns, row.Actual)\n\t\t\t\tlegacyRuns = append(legacyRuns, row.Legacy)\n\t\t\t} else if row.Failure == "" {\n\t\t\t\trow.Failure = "run did not complete; see executed test failure"\n\t\t\t}';
const after='if !row.Complete && row.Failure == "" {\n\t\t\t\trow.Failure = "run did not complete; see executed test failure"\n\t\t\t} else {\n\t\t\t\tcomplete++\n\t\t\t\tactualRuns = append(actualRuns, row.Actual)\n\t\t\t\tlegacyRuns = append(legacyRuns, row.Legacy)\n\t\t\t}';
const old=replaceOne(omitted,before,after);
for(const [label,source]of[['corrected',corrected],['old',old]]){
 const output='.cache/first-hour-population-census-'+label+'_test.go';write(output,source);
 const overlay=structuredClone(base);overlay.Replace['harness/manual_first_hour_population_test.go']='../'+output;
 write('.cache/first-hour-population-census-'+label+'.json',JSON.stringify(overlay,null,2)+'\n');
}
