import { BaseSequencer, type TestSpecification } from "vitest/node";

function priority(spec: TestSpecification): number {
  const file = spec.moduleId.split(/[\\/]/u).pop();
  // Measured long critical path, then the mandatory real 60-second wait.
  if (file === "garage-surfaces-events-browser.test.ts") return 0;
  if (file === "garage-idle-opportunity-browser.test.ts") return 1;
  return 2;
}

export class BrowserSequencer extends BaseSequencer {
  override async sort(specs: TestSpecification[]): Promise<TestSpecification[]> {
    const sorted = await super.sort(specs);
    const groups = new Map<TestSpecification["project"], { files: TestSpecification[]; next: number }>();
    for (const spec of sorted) {
      if (!groups.has(spec.project)) groups.set(spec.project, { files: [], next: 0 });
      groups.get(spec.project)!.files.push(spec);
    }
    for (const group of groups.values()) group.files.sort((a, b) => priority(a) - priority(b));
    // Retain every project's exact positions in Vitest's own sorted output.
    // Stable ties preserve its ordering for all other files; sharding is inherited.
    return sorted.map(spec => {
      const group = groups.get(spec.project)!;
      return group.files[group.next++]!;
    });
  }
}
