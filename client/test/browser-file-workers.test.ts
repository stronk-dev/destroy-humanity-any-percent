import { expect, it } from "vitest";
import { browserFileWorkers } from "../tools/browser-file-workers";

it("retains the four-CPU model's per-engine width without oversizing smaller hosts", () => {
  const expected = [[1, 1], [2, 1], [3, 2], [4, 3], [5, 3], [14, 3], [64, 3], [Number.MAX_SAFE_INTEGER, 3]];
  for (const [cpus, workers] of expected) expect(browserFileWorkers(cpus!)).toBe(workers);
});

it("refuses invalid CPU observations instead of silently choosing a width", () => {
  for (const cpus of [0, -1, 1.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) {
    expect(() => browserFileWorkers(cpus)).toThrow("positive safe integer");
  }
});
