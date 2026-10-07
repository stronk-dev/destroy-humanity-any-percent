import { expect, it } from "vitest";

import { formatAxisFactor } from "../src/game-ui/axis-factor-format";

it.each([
  ["1e0", "1"],
  ["1.2e0", "1.2"],
  ["1.16e0", "1.16"],
  ["1.24e0", "1.24"],
  ["2.1e0", "2.1"],
  ["1.88e0", "1.88"],
  ["1.000008e0", "1.000008"],
  ["1.11111111111e1", "11.1111111111"],
  ["1.005e1", "10.05"],
  ["1.12345678901e2", "112.345678901"],
  ["1e2", "1e2"],
  ["1e1000", "1e1000"],
  ["-1.005e1", "-10.05"],
  ["0", "0"],
])("displays canonical factor %s without rounding or float artifacts", (value, expected) => {
  expect(formatAxisFactor(value)).toBe(expected);
});

it.each(["1.20e0", "1.2", "NaN", "Infinity", "1e9000000000000000"])("rejects invalid factor %s", (value) => {
  expect(() => formatAxisFactor(value)).toThrow();
});
