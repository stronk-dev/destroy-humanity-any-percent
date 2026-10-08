import { describe, expect, it } from "vitest";
import validFixture from "../../balance/testdata/server-garden/fixture-v1.json?raw";
import fiscalFixture from "../../balance/testdata/server-garden/fiscal-fixture-v1.json";
import catalogCorpus from "../../testdata/garden/catalog-fixtures-v1.json";
import rawCorpus from "../../testdata/garden/catalog-raw-fixtures-v1.json";
import { COPY_KEYS } from "../src/copy";
import { loadGardenCatalog, type GardenDeclarations } from "../src/garden/catalog";

interface RawCase {
  name: string; reject: boolean; contained?: boolean; raw?: string;
  from?: string; to?: string; occurrences?: number; prefix?: string; suffix?: string;
}

const declarations: GardenDeclarations = {
  copyKeys: new Set(COPY_KEYS), resourceIds: new Set(catalogCorpus.resource_ids),
  fiscalUnlockIds: new Set(fiscalFixture.unlock_rows.map((row) => row.unlock_id)),
  fiscalGeneratorIds: new Set(fiscalFixture.generator_level_rows.map((row) => row.generator_id)),
};

// The Go test uses the same first literal replacement and count assertion.
// JSON.parse/stringify would destroy the raw spelling this corpus exercises.
function input(test: RawCase): string {
  let bytes = test.raw ?? validFixture;
  if (test.from !== undefined) {
    expect(bytes.split(test.from).length - 1, test.name).toBe(test.occurrences);
    bytes = bytes.replace(test.from, test.to!);
  }
  return (test.prefix ?? "") + bytes + (test.suffix ?? "");
}

const cases: RawCase[] = rawCorpus.cases;
const browser = typeof document !== "undefined";

interface ChildResult {
  stdout: string; stderr: string; error?: Error & { code?: string };
  signal: string | null; status: number | null;
}

// Observation only: absence of the marker does not identify why entry was not
// observed. Retain the original child outputs and the whole-process deadline.
async function observeChild(program: string, args: string[] = []): Promise<{ result: ChildResult; elapsedMs: number }> {
  const moduleName = "node:child_process";
  const { spawnSync } = await import(/* @vite-ignore */ moduleName) as {
    spawnSync(command: string, args: string[], options: { cwd: URL; encoding: "utf8"; timeout: number }): ChildResult;
  };
  const node = (globalThis as unknown as { process: { execPath: string } }).process;
  const start = performance.now();
  const result = spawnSync(node.execPath, ["--input-type=module", "-e", program, ...args], {
    cwd: new URL("../", import.meta.url), encoding: "utf8", timeout: 1000,
  });
  return { result, elapsedMs: performance.now() - start };
}

function assertChildAdmission(result: ChildResult, elapsedMs: number, name: string, reject: boolean): void {
  const diagnostic = JSON.stringify({
    name, entry: result.stdout.includes("entered-loader\n") ? "observed" : "unobserved",
    elapsedMs, deadlineMs: 1000,
    error: result.error ? { message: result.error.message, code: result.error.code ?? null } : null,
    signal: result.signal, status: result.status, stdout: result.stdout, stderr: result.stderr,
  });
  // Every failure, including the first missing marker, carries the entire
  // observation. A guard termination is never accepted as a syntax rejection.
  expect(result.stdout, diagnostic).toContain("entered-loader\n");
  expect(result.error?.message, diagnostic).toBeUndefined();
  expect(result.signal, diagnostic).toBeNull();
  expect(result.status, diagnostic).toBe(0);
  expect(result.stderr, diagnostic).toBe("");
  expect(result.stdout, diagnostic).toBe(`entered-loader\n${reject ? "SyntaxError" : "accepted"}\n`);
}

describe("Server Garden raw catalog admission (SG1)", () => {
  it("retains the complete literal-byte population", () => {
    expect(rawCorpus.version).toBe(1);
    expect(rawCorpus.valid).toBe("balance/testdata/server-garden/fixture-v1.json");
    expect(cases).toHaveLength(36);
    expect(cases.filter((test) => test.reject)).toHaveLength(22);
    expect(cases.filter((test) => !test.reject)).toHaveLength(14);
  });

  for (const test of cases.filter((candidate) => !candidate.contained)) {
    it(test.name, () => {
      const load = () => loadGardenCatalog(input(test), declarations);
      if (test.reject) expect(load).toThrow(SyntaxError);
      else expect(load).not.toThrow();
    });
  }

  // A scanner regression must fail locally, not hang the parent test runner.
  // Node APIs are imported only inside the Node-only body, never at collection
  // time: the unmodified browser configuration collects this test too.
  it.skipIf(browser)("contains malformed-string regressions in a child process", async () => {
    const serialized = JSON.stringify(Object.fromEntries(Object.entries(declarations).map(([key, values]) => [key, [...values]])));
    const program = `
      import { loadGardenCatalog } from "./src/garden/catalog.ts";
      const declarations = Object.fromEntries(Object.entries(JSON.parse(process.argv[2])).map(([key, values]) => [key, new Set(values)]));
      console.log("entered-loader");
      try {
        loadGardenCatalog(process.argv[1], declarations);
        console.log("accepted"); process.exit(process.argv[3] === "accept" ? 0 : 1);
      } catch (error) {
        console.log(error.constructor.name);
        process.exit(process.argv[3] === "reject" && error instanceof SyntaxError ? 0 : 2);
      }
    `;
    for (const test of [cases[0]!, ...cases.filter((candidate) => candidate.contained)]) {
      const { result, elapsedMs } = await observeChild(program, [input(test), serialized, test.reject ? "reject" : "accept"]);
      assertChildAdmission(result, elapsedMs, test.name, test.reject);
    }
  }, 10_000);

  for (const control of [
    { name: "startup-exception", program: 'throw new Error("SG1 child startup control");', entered: false },
    { name: "entered-nontermination", program: 'console.log("entered-loader"); while (true) {}', entered: true },
  ]) {
    it.skipIf(browser)(`reports and refuses real ${control.name} control`, async () => {
      const { result, elapsedMs } = await observeChild(control.program);
      expect(Number.isFinite(elapsedMs)).toBe(true);
      expect(elapsedMs).toBeGreaterThanOrEqual(0);
      let failure: unknown;
      try { assertChildAdmission(result, elapsedMs, control.name, true); }
      catch (error) { failure = error; }
      expect(failure).toBeInstanceOf(Error);
      const message = (failure as Error).message;
      expect(message).toContain(`"name":"${control.name}"`);
      expect(message).toContain(`"entry":"${control.entered ? "observed" : "unobserved"}"`);
      expect(message).toMatch(/"elapsedMs":\d+(?:\.\d+)?/);
      expect(message).toContain('"deadlineMs":1000');
      if (control.entered) {
        expect(result.error?.code).toBe("ETIMEDOUT");
        expect(result.signal).toBe("SIGTERM");
        expect(result.status).toBeNull();
        expect(message).toContain('"code":"ETIMEDOUT"');
        expect(message).toContain('"signal":"SIGTERM"');
        expect(message).toContain('"status":null');
        expect(message).toContain('"stdout":"entered-loader\\n"');
        expect(message).toContain('"stderr":""');
      } else {
        expect(result.error).toBeUndefined();
        expect(result.signal).toBeNull();
        expect(result.status).toBe(1);
        expect(message).toContain('"error":null');
        expect(message).toContain('"signal":null');
        expect(message).toContain('"status":1');
        expect(message).toContain('"stdout":""');
        expect(message).toContain('"stderr":');
        expect(message).toContain("SG1 child startup control");
      }
    }, 10_000);
  }

  for (const test of cases.filter((candidate) => candidate.contained)) {
    // Actual native browser loader calls; subprocess-only in Node. The full
    // browser lane must not be run against the known nonterminating baseline.
    it.skipIf(!browser)(`browser rejects ${test.name}`, () => {
      expect(() => loadGardenCatalog(input(test), declarations)).toThrow(SyntaxError);
    });
  }
});
