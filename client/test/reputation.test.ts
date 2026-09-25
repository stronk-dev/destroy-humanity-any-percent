import economyJSON from "../../balance/catalogs/phase0.json";
import curriculumJSON from "../../balance/curriculum/t0-t1.json";
import fixtureTree from "../../balance/testdata/reputation-tree/fixture-v1.json";
import corpus from "../../testdata/reputation/tree-fixtures-v1.json";
import vectors from "../../testdata/reputation/bonus-vectors-v1.json";
import { describe, expect, it } from "vitest";

import { COPY_KEYS } from "../src/copy/generated/types";
import { parseCurriculumCatalog } from "../src/curriculum";
import { parseCatalog } from "../src/economy-kernel";
import { loadReputationTree, reputationAvailable, reputationBonusFactor, reputationUnlockPpm, type ReputationDeclarations } from "../src/reputation";

function fixtureEconomy(withDeclaration: boolean) {
  const source = structuredClone(economyJSON) as { multiplier_sources: unknown[] };
  if (withDeclaration) source.multiplier_sources.push({ id: "reputation.founder_bonus", slot: "prestige", target: "all", provider: "reputation_tree" });
  return parseCatalog(source);
}

const economy = fixtureEconomy(true);
const copyKeys = new Set<string>(COPY_KEYS);
const declarations: ReputationDeclarations = { economy, copyKeys, curriculum: parseCurriculumCatalog(curriculumJSON, economy, copyKeys, ["gate.t0_to_t1"]) };

describe("reputation tree loader", () => {
  it("loads the fixture and rejects every shared corpus fixture", () => {
    expect(corpus.schema_version).toBe(1);
    const tree = loadReputationTree(fixtureTree, declarations);
    expect(tree.nodes).toHaveLength(9);
    const rules = new Set<number>();
    for (const row of corpus.rejections) {
      expect(() => loadReputationTree(row.tree, declarations), row.name).toThrow(row.rule > 0 ? new RegExp(`rule ${row.rule}:`, "u") : /invalid reputation tree/u);
      rules.add(row.rule);
    }
    for (let rule = 1; rule <= 8; rule += 1) expect(rules.has(rule), `rule ${rule}`).toBe(true);
  });

  it("rejects a tree whose economy lacks the reputation.founder_bonus declaration", () => {
    expect(() => loadReputationTree(fixtureTree, { ...declarations, economy: fixtureEconomy(false) })).toThrow(/rule 1:/u);
  });
});

describe("reputation accounting and bonus", () => {
  it("derives available and unlock ppm", () => {
    const tree = loadReputationTree(fixtureTree, declarations);
    expect(reputationAvailable(10, 4)).toBe(6);
    for (const [level, spent] of [[3, 4], [-1, 0], [5, -1], [Number.MAX_SAFE_INTEGER + 1, 0]] as const) expect(() => reputationAvailable(level, spent)).toThrow();
    expect(reputationUnlockPpm(tree, ["reputation.starter.cash_small", "reputation.unlock.p05", "reputation.unlock.p25", "reputation.unlock.retired"])).toBe(250_000);
    expect(reputationUnlockPpm(tree, [])).toBe(0);
    expect(() => reputationUnlockPpm(tree, ["reputation.unlock.p25", "reputation.unlock.p05"])).toThrow();
  });

  it("byte-matches the Go-authored bonus vectors", () => {
    expect(vectors.schema_version).toBe(1);
    for (const vector of vectors.vectors) expect(reputationBonusFactor(vector.level, vector.spent, vector.per_level_ppm, vector.unlock_ppm), JSON.stringify(vector)).toBe(vector.factor);
    expect(() => reputationBonusFactor(1, 2, 1, 1)).toThrow();
  });
});
