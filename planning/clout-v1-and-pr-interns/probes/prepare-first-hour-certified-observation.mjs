import './prepare-first-hour-lifecycle-observation.mjs';
import {readFileSync, writeFileSync, lstatSync} from 'node:fs';
import {createHash} from 'node:crypto';

// Promote the EXISTING proof declarations into this manual dependency build.
// Keep their test tails in _test.go, not duplicate proof/arithmetic copies.
const base='planning/clout-v1-and-pr-interns/probes/';
const overlay=JSON.parse(readFileSync('.cache/first-hour-lifecycle-overlay.json','utf8'));
const sha=x=>createHash('sha256').update(x).digest('hex');
function write(path,data){try{if(!lstatSync(path).isFile()||lstatSync(path).isSymbolicLink())throw Error('unsafe generated path');}catch(e){if(e.code!=='ENOENT')throw e;}writeFileSync(path,data);}
function importsFor(header,body){
  return header.replace(/\t(?:(\w+) )?"([^"]+)"\n/g,(line,alias,path)=>{
    const name=alias||path.split('/').at(-1);
    return new RegExp('\\b'+name+'\\.').test(body)?line:'';
  });
}
function replaceOne(source,before,after){if(source.split(before).length!==2)throw Error('certificate anchor drift: '+before);return source.replace(before,after);}
const records=[];
for(const [file,testKey,cut]of [
  ['production-cell-enclosure.go.txt','production/manual_production_cell_enclosure_test.go','func TestProductionCellEnclosureResearch('],
  ['production-wait-cells.go.txt','production/manual_production_wait_cells_test.go','func TestProductionWaitCellConstructorResearch('],
  ['wait-state-time.go.txt','production/manual_wait_state_time_test.go','func TestWaitStateTimeCertificateResearch('],
  ['wait-first-hour-lifetime.go.txt','production/manual_wait_first_hour_lifetime_test.go','func TestWaitFirstHourLifetimeBindingResearch('],
]){
  const source=readFileSync(base+file,'utf8');
  const headerEnd=source.indexOf('\n)\n')+3, split=source.indexOf(cut);
  if(headerEnd<3||split<headerEnd||source.split(cut).length!==2)throw Error('proof split drift');
  const header=source.slice(0,headerEnd), prefix=source.slice(headerEnd,split), tail=source.slice(split);
  if(header+prefix+tail!==source)throw Error('nonreversible split');
  let helper=prefix;
  if(file==='production-wait-cells.go.txt')helper=replaceOne(helper,'fromMS          int64 // Query floor only; NEVER a new accrual origin.','fromMS          int64 // Query floor only; NEVER a new accrual origin.\n\tmatches func(int64, decimal.Decimal) (bool, error)');
  const stem=file.replace('.go.txt',''), normal='.cache/certified-'+stem+'.go', tests='.cache/certified-'+stem+'_test.go';
  const normalSource=importsFor(header,helper)+helper, testSource=importsFor(header,tail)+'\n'+tail;
  write(normal,normalSource);write(tests,testSource);
  overlay.Replace[testKey]='../'+tests;
  overlay.Replace['production/manual_certified_'+stem.replaceAll('-','_')+'.go']='../'+normal;
  records.push({file,original:sha(source),helper:sha(normalSource),tests:sha(testSource)});
}
const queryOriginal=readFileSync(base+'wait-query-order.go.txt','utf8');
let query=replaceOne(queryOriginal,'if context.fromMS == 0 && before.Gte(quote) {','if context.matches == nil && context.fromMS == 0 && before.Gte(quote) {');
query=replaceOne(query,'if value.Gte(quote) {\n\t\t\treturn at, nil','met := value.Gte(quote)\n\t\tif met && context.matches != nil {\n\t\t\tvar err error\n\t\t\tmet, err = context.matches(at, value)\n\t\t\tif err != nil { return -1, err }\n\t\t}\n\t\tif met {\n\t\t\treturn at, nil');
write('.cache/certified-query-order.go',query);
delete overlay.Replace['production/manual_wait_query_order_helper_test.go'];
overlay.Replace['production/manual_certified_query_order.go']='../.cache/certified-query-order.go';
const originalBridge=readFileSync(base+'owned-first-hour-bridge.go.txt','utf8');
let bridge=replaceOne(originalBridge,'guard                          func() error','guard                          func() error\n\tcertified bool\n\tcertificateWork []production.ManualFirstHourCertificateObservationResearch\n\tselections []manualFirstHourTraceResearch');
bridge=replaceOne(bridge,'scope.queries++\n\t\tpoints := 0','scope.queries++\n\t\tif scope.certified { return scope.queryCertified(origin, nowMS, wait, requirements, budget) }\n\t\tpoints := 0');
bridge=replaceOne(bridge,'scope.binding = binding\n\t}\n\truntime.company, runtime.revision', 'scope.binding = binding\n\t\tif err := scope.observeSelectedResearch(binding.Head); err != nil { return err }\n\t}\n\truntime.company, runtime.revision');
bridge=replaceOne(bridge,'scope.binding = next\n\t}\n\treturn result, err','scope.binding = next\n\t\tif err := scope.observeSelectedResearch(next.Head); err != nil { return result, err }\n\t}\n\treturn result, err');
write('.cache/certified-first-hour-bridge.go',bridge);
overlay.Replace['harness/manual_owned_first_hour_bridge.go']='../.cache/certified-first-hour-bridge.go';
overlay.Replace['production/manual_first_hour_certificate.go']='../'+base+'first-hour-certificate.go.txt';
overlay.Replace['harness/manual_first_hour_certificate.go']='../'+base+'owned-first-hour-certificate.go.txt';
overlay.Replace['harness/manual_first_hour_certificate_test.go']='../'+base+'owned-first-hour-certificate-test.go.txt';
write('.cache/first-hour-certified-overlay.json',JSON.stringify(overlay,null,2)+'\n');
console.log(JSON.stringify({records,query_original:sha(queryOriginal),query:sha(query),bridge_original:sha(originalBridge),bridge:sha(bridge),entries:Object.keys(overlay.Replace).length}));
