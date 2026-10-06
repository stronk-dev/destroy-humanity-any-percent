import economyJSON from "../../balance/catalogs/phase0.json";
import curriculumJSON from "../../balance/curriculum/t0-t1.json";
import fixtureTree from "../../balance/testdata/reputation-tree/fixture-v1.json";
import corpus from "../../testdata/reputation/tree-fixtures-v1.json";
import vectors from "../../testdata/reputation/bonus-vectors-v1.json";
import starterKeyRejections from "../../testdata/reputation/starter-key-rejections-v1.json";
import bonusDomain from "../../testdata/reputation/bonus-domain-v1.json";
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
  it("rejects every shared cross-arm starter key even when zero, empty or null", () => {
    expect(starterKeyRejections.schema_version).toBe(1);
    expect(starterKeyRejections.cases).toHaveLength(20);
    const legal = loadReputationTree(fixtureTree, declarations);
    expect(new Set(legal.nodes.flatMap((node) => node.kind === "starter" ? [node.starter.kind] : []))).toEqual(new Set(["resource_grant", "generated_generators", "preowned_upgrade"]));
    for (const row of starterKeyRejections.cases) {
      const source = structuredClone(fixtureTree);
      const node = source.nodes.find((entry) => entry.starter?.kind === row.kind);
      expect(node?.starter, row.kind).toBeDefined();
      const starter = node!.starter as Record<string, unknown>;
      expect(Object.hasOwn(starter, row.key), row.key).toBe(false);
      starter[row.key] = row.value;
      expect(() => loadReputationTree(source, declarations), JSON.stringify(row)).toThrow(/rule 7:/u);
    }
  });

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
  it("keeps the earned-level bonus unchanged across 462 legal spend cases", () => {
    expect(bonusDomain.schema_version).toBe(1);
    let triples = 0;
    let spends = 0;
    for (const level of bonusDomain.levels) {
      for (const perLevel of bonusDomain.per_level_ppm) {
        for (const unlock of bonusDomain.unlock_ppm) {
          const baseline = reputationBonusFactor(level, 0, perLevel, unlock);
          if (level === 0 || unlock === 0) expect(baseline).toBe("1e0");
          triples += 1;
          const candidates = new Set([0, 1, Math.floor(level / 2), level]);
          for (const spent of candidates) {
            if (spent > level) continue;
            expect(reputationAvailable(level, spent)).toBe(level - spent);
            expect(reputationBonusFactor(level, spent, perLevel, unlock), JSON.stringify({ level, spent, perLevel, unlock })).toBe(baseline);
            spends += 1;
          }
        }
      }
    }
    expect(triples).toBe(147);
    expect(spends).toBe(462);
  });

  it("rejects the eight shared invalid bonus-domain tuples", () => {
    expect(bonusDomain.schema_version).toBe(1);
    expect(bonusDomain.invalid).toHaveLength(8);
    for (const row of bonusDomain.invalid) {
      expect(() => reputationBonusFactor(row.level, row.spent, row.per_level_ppm, row.unlock_ppm), row.name).toThrow(RangeError);
      if (row.level < 0 || !Number.isSafeInteger(row.level) || row.spent < 0 || row.spent > row.level) {
        expect(() => reputationAvailable(row.level, row.spent), row.name).toThrow(RangeError);
      } else {
        expect(reputationAvailable(row.level, row.spent), row.name).toBe(row.level - row.spent);
      }
    }
  });

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
