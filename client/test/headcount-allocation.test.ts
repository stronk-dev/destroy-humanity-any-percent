import { describe, expect, it } from "vitest";
import fixtureJSON from "../../testdata/headcount-allocation.json";
import { project, type Allocation, type Role } from "../src/headcount/allocation";

const fixture = fixtureJSON as {
  schema_version: number; hardcap: number; roles: Role[];
  effective_heads: { name: string; seats: number; k: number; assigned: Record<string, number>; expected: Allocation }[];
};
const rolesAt = (k: number): Role[] => fixture.roles.map((role) => role.kind === "convert" ? { ...role, heads_per_converter: k } : { ...role });

// Independent unbounded integer model. No floor-division capacity branch from
// the production function is reused in the oracle.
function exactModel(e: number, m: number, f: number, k: number, seats: number): Allocation {
  const product = BigInt(m) * BigInt(k);
  const diverted = product < BigInt(e) ? product : BigInt(e);
  const used = (diverted + BigInt(k) - 1n) / BigInt(k);
  return {
    producing_heads: { "role.producer": e - Number(diverted), "role.feeder": f },
    diverted: { "role.converter": Number(diverted) },
    idle_converters: { "role.converter": m - Number(used) },
    unassigned: seats - e - m - f,
  };
}

describe("headcount source-independent allocation", () => {
  it("loads the bounded shared corpus", () => {
    expect(fixture.schema_version).toBe(1);
    expect(fixture.hardcap).toBe(Number.MAX_SAFE_INTEGER);
    expect(fixture.effective_heads).toHaveLength(12);
  });
  it.each(fixture.effective_heads)("$name", (v) => {
    const roles = rolesAt(v.k);
    const before = JSON.stringify([roles, v.assigned]);
    const result = project(fixture.hardcap, v.seats, roles, v.assigned);
    expect(result).toEqual(v.expected);
    expect(JSON.stringify([roles, v.assigned])).toBe(before);
    expect(project(fixture.hardcap, v.seats, [...roles].reverse(), v.assigned)).toEqual(result);
  });
  it("agrees with exact integers throughout the small domain", () => {
    for (let e = 0; e <= 16; e++) for (let m = 0; m <= 16; m++) for (let f = 0; f <= 2; f++) for (let k = 1; k <= 9; k++) {
      const seats = e + m + f + 2;
      expect(project(100, seats, rolesAt(k), { "role.producer": e, "role.converter": m, "role.feeder": f }),
        `E=${e} M=${m} F=${f} k=${k}`).toEqual(exactModel(e, m, f, k, seats));
    }
  });
  it("agrees with exact integers near the maximum seat budget", () => {
    const max = Number.MAX_SAFE_INTEGER;
    for (const e of [0, 1, 2, 5_900_000_000_000_000, max - 100, max - 1, max]) {
      for (const m of new Set([0, Math.min(1, max - e), max - e])) for (const k of [1, 2, 3, max]) {
        expect(project(max, max, rolesAt(k), { "role.producer": e, "role.converter": m, "role.feeder": 0 }),
          `E=${e} M=${m} k=${k}`).toEqual(exactModel(e, m, 0, k, max));
      }
    }
  });
  type Input = { cap: number; seats: number; roles: Role[]; assigned: Record<string, number> };
  const corruptions: [string, (v: Input) => void][] = [
    ["zero cap", (v) => { v.cap = 0; }],
    ["unsafe cap", (v) => { v.cap = Number.MAX_SAFE_INTEGER + 1; }],
    ["negative seats", (v) => { v.seats = -1; }],
    ["over cap", (v) => { v.seats = 13; }],
    ["empty roles", (v) => { v.roles = []; }],
    ["missing assignment", (v) => { delete v.assigned["role.feeder"]; }],
    ["unknown assignment", (v) => { delete v.assigned["role.feeder"]; v.assigned["role.other"] = 0; }],
    ["negative assignment", (v) => { v.assigned["role.feeder"] = -1; }],
    ["unsafe assignment", (v) => { v.assigned["role.feeder"] = Number.MAX_SAFE_INTEGER + 1; }],
    ["over seat budget", (v) => { v.assigned["role.feeder"] = 2; }],
    ["duplicate roles", (v) => { v.roles[2] = v.roles[1]; }],
    ["invalid ID", (v) => { v.roles[1] = { id: "Bad", kind: "pool_feed" }; }],
    ["unknown kind", (v) => { v.roles[1] = { id: "role.feeder", kind: "unknown" } as unknown as Role; }],
    ["zero ratio", (v) => { v.roles[0] = { ...v.roles[0], heads_per_converter: 0 }; }],
    ["unsafe ratio", (v) => { v.roles[0] = { ...v.roles[0], heads_per_converter: Number.MAX_SAFE_INTEGER + 1 }; }],
    ["cross-pool source", (v) => { v.roles[0] = { ...v.roles[0], from_role_id: "role.other" }; }],
    ["self source", (v) => { v.roles[0] = { ...v.roles[0], from_role_id: "role.converter" }; }],
    ["feeder source", (v) => { v.roles[0] = { ...v.roles[0], from_role_id: "role.feeder" }; }],
    ["conversion chain", (v) => {
      v.roles[0] = { ...v.roles[0], from_role_id: "role.feeder" };
      v.roles[1] = { id: "role.feeder", kind: "convert", from_role_id: "role.producer", heads_per_converter: 1 };
    }],
    ["two converters", (v) => { v.roles[1] = { id: "role.feeder", kind: "convert", from_role_id: "role.producer", heads_per_converter: 1 }; }],
    ["nonconverter conversion metadata", (v) => { v.roles[1] = { ...v.roles[1], heads_per_converter: 1 }; }],
    ["RFC overflow example exceeds safe seat budget", (v) => {
      v.cap = v.seats = Number.MAX_SAFE_INTEGER; v.roles = rolesAt(3);
      v.assigned["role.converter"] = 3_100_000_000_000_000; v.assigned["role.producer"] = 9_000_000_000_000_000;
    }],
    ["fractional count", (v) => { v.assigned["role.feeder"] = .5; }],
    ["NaN count", (v) => { v.assigned["role.feeder"] = NaN; }],
    ["infinite ratio", (v) => { v.roles[0] = { ...v.roles[0], heads_per_converter: Infinity }; }],
  ];
  it.each(corruptions)("refuses %s without changing inputs", (_name, corrupt) => {
    const v: Input = { cap: 12, seats: 12, roles: rolesAt(1), assigned: { "role.converter": 3, "role.feeder": 0, "role.producer": 8 } };
    corrupt(v);
    const before = structuredClone(v);
    expect(() => project(v.cap, v.seats, v.roles, v.assigned)).toThrow("invalid headcount allocation");
    expect(v).toEqual(before);
  });
  it("does not share mutable result maps with later calls", () => {
    const v = fixture.effective_heads[0];
    const result = project(fixture.hardcap, v.seats, rolesAt(v.k), v.assigned);
    (result.producing_heads as Record<string, number>)["role.producer"] = 999;
    expect(project(fixture.hardcap, v.seats, rolesAt(v.k), v.assigned)).toEqual(v.expected);
  });
  it("keeps independent conversions and feeders separate", () => {
    const roles: Role[] = [
      { id: "role.convert_z", kind: "convert", from_role_id: "role.source_a", heads_per_converter: 2 },
      { id: "role.source_b", kind: "produce" },
      { id: "role.feed", kind: "pool_feed" },
      { id: "role.convert_a", kind: "convert", from_role_id: "role.source_b", heads_per_converter: 3 },
      { id: "role.source_a", kind: "produce" },
    ];
    expect(project(20, 20, roles, { "role.convert_z": 4, "role.source_b": 7, "role.feed": 2, "role.convert_a": 2, "role.source_a": 3 })).toEqual({
      producing_heads: { "role.source_a": 0, "role.source_b": 1, "role.feed": 2 },
      diverted: { "role.convert_a": 6, "role.convert_z": 3 },
      idle_converters: { "role.convert_a": 0, "role.convert_z": 2 }, unassigned: 2,
    });
  });
});
