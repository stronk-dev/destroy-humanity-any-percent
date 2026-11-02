import { expect, test } from "vitest";
import { BaseSequencer, type TestProject, type TestSpecification, type Vitest } from "vitest/node";
import { BrowserSequencer } from "../tools/browser-sequencer";

const project = (name: string, groupOrder = 0) => ({ name,
  config: { sequence: { groupOrder }, isolate: true },
}) as unknown as TestProject;
const file = (owner: TestProject, name: string) => ({ project: owner,
  moduleId: `/workspace/client/test/${name}`,
}) as TestSpecification;
const context = (shard = { index: 1, count: 1 }) => ({
  config: { root: "/workspace/client", shard },
  cache: { getFileTestResults: () => undefined, getFileStats: () => undefined },
}) as unknown as Vitest;

test("front-loads the measured event family and real idle wait without changing project placement or other ordering", async () => {
  const a = project("chromium"), b = project("firefox"), c = project("webkit", 1);
  const specs = [a, b, c].flatMap(owner => [
    file(owner, "unknown-a.test.ts"), file(owner, "garage-idle-opportunity-browser.test.ts"),
    file(owner, "unknown-b.test.ts"), file(owner, "garage-surfaces-events-browser.test.ts"),
  ]).reverse();
  const baseline = await new BaseSequencer(context()).sort(specs);
  const actual = await new BrowserSequencer(context()).sort(specs);
  expect(actual).toHaveLength(baseline.length);
  expect(actual.map(spec => spec.project)).toEqual(baseline.map(spec => spec.project));
  expect(new Set(actual)).toEqual(new Set(specs));
  for (const owner of [a, b, c]) {
    const ordered = actual.filter(spec => spec.project === owner);
    expect(ordered.slice(0, 2).map(spec => spec.moduleId.split("/").pop()))
      .toEqual(["garage-surfaces-events-browser.test.ts", "garage-idle-opportunity-browser.test.ts"]);
    expect(ordered.slice(2)).toEqual(baseline.filter(spec => spec.project === owner && spec.moduleId.includes("unknown-")));
  }
  expect(specs.map(spec => spec.project)).not.toEqual(actual.map(spec => spec.project));
});

test("leaves a population without either known wait in Vitest's original order", async () => {
  const a = project("firefox");
  const specs = [file(a, "future-browser.test.ts"), file(a, "existing-browser.test.ts")];
  expect(await new BrowserSequencer(context()).sort(specs)).toEqual(await new BaseSequencer(context()).sort(specs));
});

test("preserves repeated specifications rather than deduplicating or mutating the input", async () => {
  const a = project("chromium"), slow = file(a, "garage-idle-opportunity-browser.test.ts");
  const specs = [file(a, "other.test.ts"), slow, slow];
  const before = [...specs];
  const actual = await new BrowserSequencer(context()).sort(specs);
  expect(actual).toEqual([slow, slow, before[0]]);
  expect(specs).toEqual(before);
});

test("inherits Vitest sharding and retains the complete disjoint shard union", async () => {
  const a = project("chromium");
  const specs = ["a", "b", "c", "d", "e"].map(name => file(a, `${name}.test.ts`));
  expect(BrowserSequencer.prototype.shard).toBe(BaseSequencer.prototype.shard);
  const left = await new BrowserSequencer(context({ index: 1, count: 2 })).shard(specs);
  const right = await new BrowserSequencer(context({ index: 2, count: 2 })).shard(specs);
  expect(left.filter(spec => right.includes(spec))).toEqual([]);
  expect([...left, ...right]).toHaveLength(specs.length);
  expect(new Set([...left, ...right])).toEqual(new Set(specs));
});
