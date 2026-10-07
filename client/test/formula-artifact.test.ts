import { describe, expect, it } from "vitest";
import published from "../../docs/generated/production-formulas.json?raw";
import parity from "../../balance/testdata/formula-artifact-parity-v14.json";
import replayFixture from "../../testdata/replay/apply-logged-v1.json";
import { validateFormulaArtifact } from "../src/formulas/artifact";
import { loadReplayCatalogBundle, type ReplayArtifacts } from "../src/replay";

function controlledSource(control: typeof parity.cases[number]): string {
  expect(published.split(control.replace)).toHaveLength(2);
  return published.replace(control.replace, control.with);
}

describe("stored formula artifact", () => {
  it("accepts published bytes and keeps historical absence compatible", async () => {
    expect(() => validateFormulaArtifact(published)).not.toThrow();
    const legacy = await loadReplayCatalogBundle(replayFixture.constants_hash, replayFixture.artifacts);
    expect(legacy.artifacts.formulas).toBeUndefined();
    const artifacts = { ...replayFixture.artifacts, formulas: published };
    await expect(loadReplayCatalogBundle(replayFixture.constants_hash, artifacts)).rejects.toThrow(/label mismatch/);
    const bundle = await loadReplayCatalogBundle(await artifactHash(artifacts), artifacts);
    expect(bundle.artifacts.formulas).toBe(published);
  });
  for (const control of parity.cases) it(control.id, async () => {
    const source = controlledSource(control);
    const artifacts = { ...replayFixture.artifacts, formulas: source };
    const hash = await artifactHash(artifacts);
    if (control.valid) {
      expect(() => validateFormulaArtifact(source)).not.toThrow();
      expect((await loadReplayCatalogBundle(hash, artifacts)).artifacts.formulas).toBe(source);
    } else {
      expect(() => validateFormulaArtifact(source)).toThrow();
      await expect(loadReplayCatalogBundle(hash, artifacts)).rejects.toThrow();
    }
  });
  for (const control of [{depth: 9998, valid: true}, {depth: 9999, valid: false}]) {
    it(`matches Go JSON nesting at ${control.depth} overwritten containers`, () => {
      const source = published.replace('"pinned": null', '"pinned": '+"[".repeat(control.depth)+"null"+"]".repeat(control.depth)+', "pinned": null');
      if (control.valid) expect(() => validateFormulaArtifact(source)).not.toThrow();
      else expect(() => validateFormulaArtifact(source)).toThrow();
    });
  }
  for (const source of [published + "{}", published.replace('"production_rate"', '"production_rate"\u0000'), published + "\ud800", "[".repeat(1000)]) {
    it(`rejects malformed bytes ${JSON.stringify(source.slice(-20))}`, () => {
      expect(() => validateFormulaArtifact(source)).toThrow();
    });
  }
});

async function artifactHash(artifacts: ReplayArtifacts): Promise<string> {
  const encoder = new TextEncoder(); const chunks: Uint8Array[] = [];
  const frame = (length: number): Uint8Array => { const data = new Uint8Array(8); new DataView(data.buffer).setBigUint64(0, BigInt(length)); return data; };
  for (const name of Object.keys(artifacts).sort() as (keyof ReplayArtifacts)[]) {
    const key = encoder.encode(name); const value = encoder.encode(artifacts[name]);
    chunks.push(frame(key.length), key, frame(value.length), value);
  }
  const input = new Uint8Array(chunks.reduce((sum, chunk) => sum + chunk.length, 0)); let offset = 0;
  for (const chunk of chunks) { input.set(chunk, offset); offset += chunk.length; }
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", input));
  return `sha256:${[...digest].map((byte) => byte.toString(16).padStart(2, "0")).join("")}`;
}
