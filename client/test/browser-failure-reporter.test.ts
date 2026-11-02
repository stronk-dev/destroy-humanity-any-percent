import { expect, it, vi } from "vitest";
import type { TestCase } from "vitest/node";
import BrowserFailureReporter from "../tools/browser-failure-reporter";

function testCase(state: "passed" | "failed" | "pending" | "skipped", message = "controlled assertion") {
  return { project: { name: "controlled-engine" }, module: { moduleId: "/controlled.test.ts" },
    fullName: "outer > inner", result: () => ({ state,
      errors: [{ name: "AssertionError", message, stack: "controlled-stack" }] }),
  } as unknown as TestCase;
}

it.each(["passed", "pending", "skipped"] as const)("does not label %s cases as failures", state => {
  const output = vi.spyOn(console, "error").mockImplementation(() => {});
  try {
    // Even a passed retry with retained errors must not be relabelled failed.
    new BrowserFailureReporter().onTestCaseResult(testCase(state));
    expect(output).not.toHaveBeenCalled();
  } finally { output.mockRestore(); }
});

it.each(["controlled assertion", "Expect test to fail"])("retains final failure details: %s", message => {
  const output = vi.spyOn(console, "error").mockImplementation(() => {});
  try {
    new BrowserFailureReporter().onTestCaseResult(testCase("failed", message));
    expect(output).toHaveBeenCalledTimes(1);
    expect(output.mock.calls[0]![0]).toBe("browser-test-failure");
    expect(JSON.parse(output.mock.calls[0]![1] as string)).toEqual({
      project: "controlled-engine", file: "/controlled.test.ts", test: "outer > inner", state: "failed",
      errors: [{ name: "AssertionError", message, stack: "controlled-stack" }],
    });
  } finally { output.mockRestore(); }
});
