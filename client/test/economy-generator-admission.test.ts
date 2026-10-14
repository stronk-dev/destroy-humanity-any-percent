import { describe, expect, it } from "vitest";
import fixture from "../../testdata/economy-foundation-v4.json";
import { parseCatalog } from "../src/economy-kernel";

type Root = typeof fixture;
const targets: { path: string; select(root: Root): object }[] = [];
for (const [index, row] of fixture.generator_classes.entries()) {
  targets.push({ path: row.id, select: (root) => root.generator_classes[index]! });
  for (const key of ["price", "production", "provisions", "provisioned_hardcap"] as const) {
    if (!row[key]) continue;
    targets.push({ path: `${row.id}/${key}`, select: (root) => root.generator_classes[index]![key]! });
  }
  for (const key of ["ladder", "roles"] as const) {
    row[key].forEach((child, childIndex) => targets.push({ path: `${row.id}/${key}/${childIndex}`, select: (root) => root.generator_classes[index]![key][childIndex]! }));
  }
}
for (const key of ["manual_policy", "offline_policy"] as const) targets.push({ path: key, select: (root) => root[key] });

describe("generator and operating-policy exact admission", () => {
  for (const target of targets) {
    for (const key of Object.keys(target.select(fixture))) it(`refuses alias/${target.path}/${key}`, () => {
      const source = structuredClone(fixture), row = target.select(source) as Record<string, unknown>;
      row[key.toUpperCase()] = row[key]; delete row[key];
      expect(() => parseCatalog(source)).toThrow();
    });
  }
  for (const [index, row] of fixture.generator_classes.entries()) {
    row.roles.forEach((role, roleIndex) => {
      for (const key of ["generator_id", "pool_id", "action_id", "per_purchased_ppm"]) {
        if (key in role) continue;
        it(`refuses ${role.kind}/null/${key}`, () => {
          const source = structuredClone(fixture);
          Object.assign(source.generator_classes[index]!.roles[roleIndex]!, { [key]: null });
          expect(() => parseCatalog(source)).toThrow();
        });
      }
    });
    for (const key of ["provisions", "provisioned_hardcap"] as const) {
      if (row[key]) continue;
      it(`refuses ${row.id}/null/${key}`, () => {
        const source = structuredClone(fixture); Object.assign(source.generator_classes[index]!, { [key]: null });
        expect(() => parseCatalog(source)).toThrow();
      });
    }
  }
  it("preserves all four role arms and absent optional objects", () => expect(() => parseCatalog(fixture)).not.toThrow());
});
