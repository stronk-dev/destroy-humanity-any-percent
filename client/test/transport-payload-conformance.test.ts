import { describe, expect, it } from "vitest";
import fixture from "../../testdata/transport/payload-conformance-vectors.json";
import { decodeTransportEnvelope } from "../src/transport";

describe("Transport T3 owned payload conformance", () => {
  it("executes the complete shared population", () => {
    expect(fixture.version).toBe(1);
    expect(fixture.cases).toHaveLength(174);
    expect(fixture.cases.filter(row => row.valid)).toHaveLength(12);
  });

  it.each(fixture.cases)("$name", vector => {
    const before = JSON.stringify(vector.envelope);
    if (vector.valid) {
      const decoded = decodeTransportEnvelope(vector.envelope);
      expect(decoded).toBe(vector.envelope);
      expect(decoded?.payload).toBe(vector.envelope.payload);
    } else {
      expect(() => decodeTransportEnvelope(vector.envelope)).toThrow(SyntaxError);
    }
    expect(JSON.stringify(vector.envelope)).toBe(before);
  });
});
