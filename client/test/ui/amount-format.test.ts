import { describe, expect, it } from "vitest";

import { formatAmount } from "../../src/ui/amount-format";

const goldens = [
  ["0", "0"],
  ["9.99e2", "999"],
  ["1e3", "1.00 K"],
  ["1.2345e3", "1.23 K"],
  ["9.99995e5", "1.00 M"],
  ["1e15", "1.00 Qa"],
  ["-1.234e6", "−1.23 M"],
  ["1e303", "1.00 Ce"],
] as const;

describe("Amount Standard notation goldens", () => {
  it.each(goldens)("formats %s", (value, expected) => {
    expect(formatAmount(value)).toBe(expected);
  });

  it("rejects values outside the canonical numeric boundary", () => {
    expect(() => formatAmount("1000")).toThrow(/canonical/);
    expect(() => formatAmount("Infinity")).toThrow(/canonical/);
  });
});
