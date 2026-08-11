import { afterEach, beforeEach } from "vitest";

const asynchronousErrors: unknown[] = [];

function recordError(event: ErrorEvent): void {
  asynchronousErrors.push(event.error ?? new Error(event.message));
  event.preventDefault();
}

function recordRejection(event: PromiseRejectionEvent): void {
  asynchronousErrors.push(event.reason);
  event.preventDefault();
}

window.addEventListener("error", recordError);
window.addEventListener("unhandledrejection", recordRejection);

function assertNoAsynchronousErrors(): void {
  const errors = asynchronousErrors.splice(0);
  if (errors.length > 0) throw new AggregateError(errors, "browser test emitted an unhandled asynchronous error");
}

beforeEach(assertNoAsynchronousErrors);
afterEach(assertNoAsynchronousErrors);
