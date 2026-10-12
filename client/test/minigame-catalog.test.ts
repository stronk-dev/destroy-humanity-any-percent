import { describe, expect, it } from "vitest";
import fixture from "../../testdata/minigame/catalog-v2.json";
import pitchFixture from "../../testdata/minigame/pitch-v3.json";
import scalingSources from "../../testdata/minigame/scaling-source-refs-v1.json";
import { parseMinigameCatalog, parseMinigameCatalogJSON } from "../src/minigame/catalog";

describe("pinned minigame catalog", () => {
  const scaling = { schema_version: 1, scaling_inputs: [{ destination: "option.count", destination_class: "breadth",
    source_kind: "literal", source_ref: "3", op: "identity", operand: 0, clamp_min: 0, clamp_max: 8 }] };
  const withScaling = (value: unknown) => ({ ...fixture, minigames: [{ ...fixture.minigames[0]!, scaling: value }] });
  for (const row of scalingSources.cases) {
    it(`enforces scaling source/${row.name}`, () => {
      const policy = { schema_version: 1, scaling_inputs: [{ ...scaling.scaling_inputs[0]!, ...row.changes }] };
      if (row.valid) {
        expect(parseMinigameCatalog(withScaling(policy)).minigames[0]!.scaling).toEqual(policy);
        expect(parseMinigameCatalogJSON(JSON.stringify(withScaling(policy))).minigames[0]!.scaling).toEqual(policy);
      } else {
        expect(() => parseMinigameCatalog(withScaling(policy))).toThrow();
        expect(() => parseMinigameCatalogJSON(JSON.stringify(withScaling(policy)))).toThrow();
      }
    });
  }
  for (const [scope, keys] of [["root", ["schema_version", "scaling_inputs"]],
    ["row", ["destination", "destination_class", "source_kind", "source_ref", "op", "operand", "clamp_min", "clamp_max"]]] as const) {
    for (const key of keys) {
      for (const variant of ["missing", "null", "wrong_case", "case_collision"] as const) {
        it(`refuses scaling/${scope}/${key}/${variant}`, () => {
          const bad = structuredClone(scaling);
          const fields = (scope === "root" ? bad : bad.scaling_inputs[0]!) as unknown as Record<string, unknown>;
          if (variant === "missing") delete fields[key];
          else if (variant === "null") fields[key] = null;
          else {
            fields[key.toUpperCase()] = fields[key];
            if (variant === "wrong_case") delete fields[key];
          }
          expect(() => parseMinigameCatalog(withScaling(bad))).toThrow();
          expect(() => parseMinigameCatalogJSON(JSON.stringify(withScaling(bad)))).toThrow();
        });
      }
    }
  }
  it("retains explicit zero operands/minima and escaped canonical scaling keys", () => {
    const catalog = withScaling(scaling);
    const plain = parseMinigameCatalogJSON(JSON.stringify(catalog));
    const escaped = parseMinigameCatalogJSON(JSON.stringify(catalog).replace('"destination"', '"\\u0064estination"'));
    expect(plain.minigames[0]!.scaling).toEqual(scaling);
    expect(escaped).toEqual(plain);
  });

  it("loads the same complete policy row as Go", () => {
    const catalog = parseMinigameCatalog(fixture);
    expect(catalog.minigameIds).toEqual(["fixture.counter"]);
    expect(catalog.minigames[0]?.payout.payout_score_fact_id).toBe("score.total");
    expect(catalog.minigames[0]?.rating_policy.elo_ceiling).toBe(3000);
    expect(catalog.minigames[0]?.offline_quality.neutral_floor_ppm).toBe(500_000);
  });

  it("rejects missing policy bytes and noncanonical ordering", () => {
    expect(() => parseMinigameCatalog({ schema_version: 2, rating_seasons: [], minigames: [{ minigame_id: "partial" }] })).toThrow();
    expect(() => parseMinigameCatalog({ ...fixture, rating_seasons: ["ranked", "preseason"] })).toThrow();
  });

  it("loads the closed Fiscal unlock arm for The Pitch", () => {
    const definition = parseMinigameCatalog(pitchFixture).minigames[0]!;
    expect(definition.unlock_condition).toEqual({ kind: "fiscal_unlock", unlock_id: "minigame.pitch" });
    expect(definition.payout.credited_resource_id).toBe("company.cash");
  });

  it("loads the tier_at_least arm and rejects out-of-domain rows like Go (TT-PA2)", () => {
    const withUnlock = (unlock: unknown) => ({ ...pitchFixture, minigames: [{ ...pitchFixture.minigames[0]!, unlock_condition: unlock }] });
    for (const accepted of [{ kind: "tier_at_least", tier: 1 }, { exit_history_at_least: 1, kind: "tier_at_least", tier: 1 }, { kind: "tier_at_least", tier: 9 }]) {
      expect(parseMinigameCatalog(withUnlock(accepted)).minigames[0]!.unlock_condition).toEqual(accepted);
    }
    for (const rejected of [{ kind: "tier_at_least" }, { kind: "tier_at_least", tier: 10 }, { kind: "tier_at_least", tier: -1 }, { kind: "tier_at_least", tier: 1.5 },
      { extra: 1, kind: "tier_at_least", tier: 1 }, { exit_history_at_least: -1, kind: "tier_at_least", tier: 1 },
      { exit_history_at_least: null, kind: "tier_at_least", tier: 1 }, { kind: "tier_at_least", tier: "1" }]) {
      expect(() => parseMinigameCatalog(withUnlock(rejected)), JSON.stringify(rejected)).toThrow();
    }
  });
});
