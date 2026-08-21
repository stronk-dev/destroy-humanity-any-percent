function lines(source) {
  return source.split(/\r?\n/);
}

function section(source, startPattern, nextPattern) {
  const sourceLines = lines(source);
  const start = sourceLines.findIndex((line) => startPattern.test(line));
  if (start < 0) return [];
  let end = sourceLines.length;
  for (let index = start + 1; index < sourceLines.length; index += 1) {
    if (nextPattern.test(sourceLines[index])) {
      end = index;
      break;
    }
  }
  return sourceLines.slice(start, end);
}

function triggerSection(source) {
  return section(source, /^on:\s*$/, /^[A-Za-z0-9_-]+:\s*$/);
}

function jobSection(source, name) {
  return section(source, new RegExp(`^  ${name}:\\s*$`), /^  [A-Za-z0-9_-]+:\s*$/);
}

function requireLine(sourceLines, pattern, message) {
  if (!sourceLines.some((line) => pattern.test(line))) throw new Error(message);
}

function forbidLine(sourceLines, pattern, message) {
  if (sourceLines.some((line) => pattern.test(line))) throw new Error(message);
}

function requireExactJobs(source, expected, message) {
  const sourceLines = lines(source);
  const jobsStart = sourceLines.findIndex((line) => /^jobs:\s*$/.test(line));
  const jobs = sourceLines
    .slice(jobsStart + 1)
    .filter((line) => /^  [A-Za-z0-9_-]+:\s*$/.test(line))
    .map((line) => line.trim().slice(0, -1));
  if (jobs.length !== expected.length || jobs.some((job, index) => job !== expected[index])) {
    throw new Error(message);
  }
}

export function verifyCITopology(ciSource, maintenanceSource) {
  const ciTriggers = triggerSection(ciSource);
  requireLine(ciTriggers, /^  push:\s*$/, "CI must run on push");
  requireLine(ciTriggers, /^  pull_request:\s*$/, "CI must run on pull requests");
  forbidLine(ciTriggers, /^  (schedule|workflow_dispatch):/, "blocking CI must not own maintenance triggers");
  requireLine(lines(ciSource), /^permissions:\s*$/, "CI permissions block is missing");
  requireLine(lines(ciSource), /^  contents: read\s*$/, "CI must use read-only contents permission");
  requireExactJobs(ciSource, ["server", "harness", "client", "browser", "game-ui-composed", "schema"], "blocking CI must contain exactly the six governed jobs");

  const harness = jobSection(ciSource, "harness");
  requireLine(harness, /^    timeout-minutes: 5\s*$/, "blocking harness must retain the five-minute budget");
  requireLine(harness, /^      - run: make verify-harness-fast\s*$/, "blocking harness must run the fast gate");
  forbidLine(harness, /make (?:harness-check|harness-observe|verify-harness(?:\s|$))/, "blocking harness contains exhaustive work");
  forbidLine(lines(ciSource), /^  numeric-maintenance:\s*$/, "numeric maintenance must not appear as a skipped push job");

  const ciSetupGoCount = lines(ciSource).filter((line) => /^      - uses: actions\/setup-go@v6\s*$/.test(line)).length;
  const ciCacheDisabledCount = lines(ciSource).filter((line) => /^          cache: false\s*$/.test(line)).length;
  const ciModuleCacheCount = lines(ciSource).filter((line) => /^          path: ~\/go\/pkg\/mod\s*$/.test(line)).length;
  if (ciSetupGoCount !== 3 || ciCacheDisabledCount !== 3 || ciModuleCacheCount !== 2) {
    throw new Error("blocking Go jobs may cache modules only, never build/test outputs");
  }

  const maintenanceTriggers = triggerSection(maintenanceSource);
  requireLine(maintenanceTriggers, /^  schedule:\s*$/, "maintenance schedule is missing");
  requireLine(maintenanceTriggers, /^  workflow_dispatch:\s*$/, "maintenance manual trigger is missing");
  forbidLine(maintenanceTriggers, /^  (push|pull_request):/, "maintenance must not run on push or pull request");
  requireLine(lines(maintenanceSource), /^permissions:\s*$/, "maintenance permissions block is missing");
  requireLine(lines(maintenanceSource), /^  contents: read\s*$/, "maintenance must use read-only contents permission");
  requireExactJobs(maintenanceSource, ["harness-evidence", "numeric-maintenance"], "maintenance must contain exactly the two evidence jobs");

  const evidence = jobSection(maintenanceSource, "harness-evidence");
  requireLine(evidence, /^    timeout-minutes: 55\s*$/, "exhaustive harness job must have the bounded 55-minute budget");
  requireLine(evidence, /^      - run: make verify-harness-fast\s*$/, "maintenance harness must prove the fast gate first");
  requireLine(evidence, /timeout --signal=INT --kill-after=30s 50m make harness-observe .*HARNESS_OBSERVATION=harness-observation\.json/, "exhaustive harness must terminate before the job ceiling and emit an observation");
  requireLine(evidence, /make harness-observation-check HARNESS_OBSERVATION=harness-observation\.json/, "completed harness observation must be validated");
  requireLine(evidence, /^        if: always\(\)\s*$/, "harness observation must upload after success or failure");
  requireLine(evidence, /^        uses: actions\/upload-artifact@v7\s*$/, "harness observation must use the pinned artifact uploader");
  requireLine(evidence, /^          path: harness-observation\.json\s*$/, "the exact harness observation must be uploaded");
  requireLine(evidence, /^          if-no-files-found: error\s*$/, "missing harness observations must fail loud");

  const numeric = jobSection(maintenanceSource, "numeric-maintenance");
  requireLine(numeric, /^      - run: make fuzz-ci\s*$/, "numeric maintenance must run bounded fuzzing");
  requireLine(numeric, /^      - run: make vectors-check\s*$/, "numeric maintenance must check vector regeneration");

  const setupGoCount = lines(maintenanceSource).filter((line) => /^      - uses: actions\/setup-go@v6\s*$/.test(line)).length;
  const cacheDisabledCount = lines(maintenanceSource).filter((line) => /^          cache: false\s*$/.test(line)).length;
  const moduleCacheCount = lines(maintenanceSource).filter((line) => /^          path: ~\/go\/pkg\/mod\s*$/.test(line)).length;
  if (setupGoCount !== 2 || cacheDisabledCount !== 2 || moduleCacheCount !== 2) {
    throw new Error("maintenance Go jobs must cache modules only, never build/test outputs");
  }
}
