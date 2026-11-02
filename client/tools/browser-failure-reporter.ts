import type { Reporter, TestCase } from "vitest/node";

// Retain completed failures even if a later test stalls until the outer guard.
// Use final verdicts: expected failures that pass are not failed test cases.
export default class BrowserFailureReporter implements Reporter {
  onTestCaseResult(test: TestCase): void {
    const result = test.result();
    if (result.state !== "failed") return;
    console.error("browser-test-failure", JSON.stringify({
      project: test.project.name, file: test.module.moduleId, test: test.fullName,
      state: result.state,
      errors: result.errors.map(error => ({ name: error.name, message: error.message, stack: error.stack })),
    }));
  }
}
