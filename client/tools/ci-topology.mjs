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

function requireExactMakeCommand(source, job, target) {
  const commands = jobSection(source, job)
    .filter((line) => /^      - run: make\s+/.test(line))
    .map((line) => line.trim());
  if (commands.length !== 1 || commands[0] !== `- run: make ${target}`) {
    throw new Error(`blocking ${job} job must run exactly make ${target}`);
  }
}

function requireExactMakeDependencies(source, target, expected) {
  const matching = lines(source).filter((line) => line.startsWith(`${target}:`));
  if (matching.length !== 1) throw new Error(`Makefile must define ${target} exactly once`);
  const dependencies = matching[0].slice(target.length + 1).trim().split(/\s+/).filter(Boolean);
  if (dependencies.length !== expected.length || dependencies.some((dependency, index) => dependency !== expected[index])) {
    throw new Error(`${target} must contain exactly the six blocking CI leaf commands in workflow order`);
  }
}

function requireColdGoRecipes(source) {
  const expected = [
    ["test-go", "\tcd server && go test -p 1 -count=1 $(GO_TEST_FLAGS) $(GO_PACKAGES)"],
    ["replay-fixture-check", "\tcd server && go test ./production -run '^TestApplyLoggedCrossRuntimeFixture$$' -count=1"],
    ["pitch-corpus-check", "\tcd server && go test ./pitch -run '^TestPitchContentGate$$' -count=1"],
    ["typer-corpus-check", "\tcd server && go test ./typer -run '^TestTyperContentGate$$' -count=1"],
  ];
  for (const [target, command] of expected) {
    if (lines(source).filter((line) => line.startsWith(`${target}:`)).length !== 1) {
      throw new Error(`Makefile must define ${target} exactly once`);
    }
    const recipe = section(source, new RegExp(`^${target}:`), /^[^\s#].*:/)
      .filter((line) => line.startsWith("\t") && !line.trim().startsWith("#"));
    if (recipe.length !== 1 || recipe[0] !== command) {
      throw new Error(`${target} must execute its complete Go population with a cold default`);
    }
  }
}

export function verifyCITopology(ciSource, maintenanceSource, makeSource, browserSource, browserEntrypoint) {
  const ciTriggers = triggerSection(ciSource);
  requireLine(ciTriggers, /^  push:\s*$/, "CI must run on push");
  requireLine(ciTriggers, /^  pull_request:\s*$/, "CI must run on pull requests");
  forbidLine(ciTriggers, /^  (schedule|workflow_dispatch):/, "blocking CI must not own maintenance triggers");
  requireLine(lines(ciSource), /^permissions:\s*$/, "CI permissions block is missing");
  requireLine(lines(ciSource), /^  contents: read\s*$/, "CI must use read-only contents permission");
  requireExactJobs(ciSource, ["server", "harness", "client", "browser", "game-ui-composed", "schema"], "blocking CI must contain exactly the six governed jobs");
  const blocking = [
    ["server", "verify-server-core"],
    ["harness", "verify-harness-fast"],
    ["client", "verify-client"],
    ["browser", "test-browser"],
    ["game-ui-composed", "test-game-ui-composed"],
    ["schema", "verify-schema"],
  ];
  for (const [job, target] of blocking) requireExactMakeCommand(ciSource, job, target);
  requireExactMakeDependencies(makeSource, "verify-push", blocking.map(([, target]) => target));
  requireColdGoRecipes(makeSource);

  // D2 retains every functional case and the independent performance lane.
  const browserTarget = section(makeSource, /^test-browser:/, /^[^\s#].*:/);
  requireLine(browserTarget, /^test-browser: test-browser-functional$/, "full browser gate must own both functional phases");
  requireLine(browserTarget, /^\t\$\(MAKE\) test-game-ui-performance$/, "full browser gate must retain isolated performance");
  const functional = section(makeSource, /^test-browser-functional:/, /^[^\s#].*:/);
  const browserCommand = "\tcd client && VITE_GAME_UI_PERFORMANCE=0 $(CLIENT_BIN)/vitest run --config vitest.browser.config.ts ";
  const commands = functional.filter((line) => line.startsWith(browserCommand));
  if (commands.length !== 2 ||
      !commands[0].startsWith(`${browserCommand}--exclude test/typer-table-browser.test.ts $(BROWSER_TEST_FLAGS) `) ||
      !commands[1].startsWith(`${browserCommand}test/typer-table-browser.test.ts $(BROWSER_TEST_FLAGS) `)) {
    throw new Error("functional browser gate must run broad collection then the complete native Typer file");
  }

  // The Linux reproduction must share Actions' Make authority. A copied Vitest
  // invocation can silently miss a new phase while the hosted job remains correct.
  const container = section(browserSource, /^  browser:\s*$/, /^  [A-Za-z0-9_-]+:\s*$/);
  const entrypoint = section(browserSource, /^    entrypoint:\s*$/, /^    [A-Za-z0-9_-]+:/)
    .filter((line) => !/^\s*#/.test(line)).join("\n");
  const expectedEntrypoint = ["    entrypoint:", "      - timeout", "      - --verbose", "      - --signal=KILL",
    "      - 10m", "      - bash", "      - /workspace/client/tools/test-browser-ci.sh"].join("\n");
  const setupCommands = lines(browserEntrypoint ?? "").filter((line) => line.trim() && !/^\s*#/.test(line)).join("\n");
  const expectedSetup = ["set -euo pipefail", 'case "${1:-}" in', "  test-browser|test-browser-focused) ;;",
    "  *) printf '%s\\n' 'expected test-browser or test-browser-focused' >&2; exit 2 ;;", "esac",
    "apt-get update", "apt-get install -y --no-install-recommends make", "corepack enable",
    "corepack prepare pnpm@11.15.1 --activate", "pnpm --dir client install --frozen-lockfile", 'exec make "$@"'].join("\n");
  const fullContainerRecipe = section(makeSource, /^test-browser-ci:/, /^[^\s#].*:/)
    .filter((line) => line.startsWith("\t") && !line.trim().startsWith("#"));
  if (entrypoint !== expectedEntrypoint || setupCommands !== expectedSetup ||
      container.filter((line) => line.startsWith("    command:")).join("\n") !== '    command: ["test-browser"]' ||
      fullContainerRecipe.length !== 1 || fullContainerRecipe[0] !== "\tdocker compose -f compose.browser-test.yml run --rm browser") {
    throw new Error("Linux browser container must retain containment, frozen setup and the unmodified default make test-browser target");
  }

  // This existing manual consumer supplies a shell command, not a test target.
  // A command override alone now reaches the restrictive test entrypoint.
  const coordination = section(makeSource, /^research-browser-coordination:/, /^[^\s#].*:/)
    .filter((line) => line.startsWith("\t") && !line.trim().startsWith("#"));
  if (coordination.length !== 1 || !coordination[0].startsWith("\tdocker compose -f compose.browser-test.yml run --rm --entrypoint bash browser -lc '") ||
      !coordination[0].includes("node client/tools/research-browser-coordination.mjs $(shell git rev-parse HEAD)")) {
    throw new Error("manual browser coordination research must explicitly select its shell entrypoint");
  }

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
  requireExactJobs(maintenanceSource, ["harness-evidence", "numeric-maintenance", "reputation-evidence"], "maintenance must contain exactly the three evidence jobs");

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

  const reputation = jobSection(maintenanceSource, "reputation-evidence");
  requireLine(reputation, /^    timeout-minutes: 60\s*$/, "exhaustive Reputation evidence must have its bounded 60-minute budget");
  requireLine(reputation, /^      - run: make reputation-harness-check\s*$/, "maintenance must run the exhaustive Reputation harness evidence");
  const reputationTarget = "reputation-harness-check";
  const reputationRecipe = section(makeSource, /^reputation-harness-check:/, /^[^\s#].*:/)
    .filter((line) => line.startsWith("\t") && !line.trim().startsWith("#"));
  const reputationCommand = "\tcd server && CLOUD_CLICKER_REPUTATION_EXHAUSTIVE=1 go test -p 1 -count=1 -timeout 60m ./harness -run '^(TestReputationTreeRelevance|TestReputationCareerStartersShortenRunThree)$$' -v";
  if (lines(makeSource).filter((line) => line.startsWith(`${reputationTarget}:`)).length !== 1 ||
      reputationRecipe.length !== 1 || reputationRecipe[0] !== reputationCommand) {
    throw new Error("Reputation maintenance target must opt in to both complete studies cold within its existing budget");
  }

  const setupGoCount = lines(maintenanceSource).filter((line) => /^      - uses: actions\/setup-go@v6\s*$/.test(line)).length;
  const cacheDisabledCount = lines(maintenanceSource).filter((line) => /^          cache: false\s*$/.test(line)).length;
  const moduleCacheCount = lines(maintenanceSource).filter((line) => /^          path: ~\/go\/pkg\/mod\s*$/.test(line)).length;
  if (setupGoCount !== 3 || cacheDisabledCount !== 3 || moduleCacheCount !== 3) {
    throw new Error("maintenance Go jobs must cache modules only, never build/test outputs");
  }
}
