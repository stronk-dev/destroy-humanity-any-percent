import { describe, expect, it } from "vitest";
import validCatalog from "../../balance/testdata/server-garden/fixture-v1.json?raw";
import fiscalFixture from "../../balance/testdata/server-garden/fiscal-fixture-v1.json";
import catalogCorpus from "../../testdata/garden/catalog-fixtures-v1.json";
import corpus from "../../testdata/garden/state-admission-v1.json";
import { COPY_KEYS } from "../src/copy";
import { loadGardenCatalog } from "../src/garden/catalog";
import { encodeGardenState, parseGardenState } from "../src/garden/engine";

const catalog = loadGardenCatalog(validCatalog, {
  copyKeys: new Set(COPY_KEYS), resourceIds: new Set(catalogCorpus.resource_ids),
  fiscalUnlockIds: new Set(fiscalFixture.unlock_rows.map((row) => row.unlock_id)),
  fiscalGeneratorIds: new Set(fiscalFixture.generator_level_rows.map((row) => row.generator_id)),
});

describe("Garden state admission (SG2)", () => {
  it("retains both nullable controls and nonnullable refusals", () => {
    expect(corpus.version).toBe(1);
    expect(corpus.cases).toHaveLength(31);
    expect(corpus.cases.filter((test) => test.reject)).toHaveLength(23);
    expect(corpus.cases.filter((test) => !test.reject)).toHaveLength(8);
    expect(corpus.cases.filter((test) => test.go_only_raw)).toHaveLength(2);
  });

  for (const test of corpus.cases) {
    it(test.name + (test.go_only_raw ? " (raw Go check; TS parsing-loss control)" : ""), () => {
      let input = corpus.valid;
      if (test.from !== undefined) {
        expect(input.split(test.from).length - 1).toBe(test.occurrences);
        input = input.replace(test.from, test.to!);
      }
      const parse = () => parseGardenState(JSON.parse(input), catalog);
      if (test.go_only_raw) {
        // This API receives an already parsed object, not raw bytes. Do not
        // credit it with duplicate rejection JSON.parse made unobservable.
        expect(JSON.parse(input)).toEqual(JSON.parse(corpus.valid));
        expect(JSON.stringify(encodeGardenState(parse()))).toBe(corpus.valid);
      } else if (test.reject) expect(parse).toThrow(SyntaxError);
      else expect(JSON.stringify(encodeGardenState(parse()))).toBe(input.replaceAll(":-0", ":0"));
    });
  }
});
