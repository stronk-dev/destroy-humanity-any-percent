import pinnedMinigameAPI from "../../../../balance/minigame-api/first-content.json";

// The client tenant-surface registry is keyed by the same (engine_ref,
// engine_version) arm as the pinned minigame_api artifact (MA-C9). It is not a
// second tenant registry: every artifact row must have exactly one child
// presentation, and a child with no artifact row is refused at load.
export type TenantChild = "pitch_table";

export interface TenantSurfaceRow {
  readonly minigame_id: string;
  readonly engine_ref: string;
  readonly engine_version: string;
  readonly child: TenantChild;
}

const CHILDREN: ReadonlyMap<string, TenantChild> = new Map([["pitch@1.0.0", "pitch_table"]]);

export function tenantArm(engineRef: string, engineVersion: string): string { return `${engineRef}@${engineVersion}`; }

export function parseTenantSurfaceRegistry(source: unknown, children: ReadonlyMap<string, TenantChild> = CHILDREN): ReadonlyMap<string, TenantSurfaceRow> {
  if (source === null || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError("minigame_api artifact must be an object");
  const tenants = (source as Record<string, unknown>).tenants;
  if (!Array.isArray(tenants) || tenants.length === 0) throw new SyntaxError("minigame_api artifact has no tenants");
  const rows = new Map<string, TenantSurfaceRow>();
  for (const raw of tenants) {
    if (raw === null || typeof raw !== "object" || Array.isArray(raw)) throw new SyntaxError("minigame_api tenant must be an object");
    const { engine_ref: ref, engine_version: version, minigame_id: id } = raw as Record<string, unknown>;
    if (typeof ref !== "string" || typeof version !== "string" || typeof id !== "string") throw new SyntaxError("minigame_api tenant fields are invalid");
    const arm = tenantArm(ref, version);
    const child = children.get(arm);
    if (!child) throw new RangeError(`minigame tenant ${arm} has no surface child`);
    if (rows.has(arm)) throw new SyntaxError(`duplicate minigame tenant ${arm}`);
    rows.set(arm, Object.freeze({ minigame_id: id, engine_ref: ref, engine_version: version, child }));
  }
  for (const arm of children.keys()) if (!rows.has(arm)) throw new RangeError(`surface child ${arm} has no pinned tenant`);
  return rows;
}

export const MINIGAME_TENANT_SURFACES = parseTenantSurfaceRegistry(pinnedMinigameAPI);
