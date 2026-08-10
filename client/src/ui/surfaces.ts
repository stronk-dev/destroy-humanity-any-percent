import type { DiscreteFact } from "../shell/contracts";

export type SurfaceUnlock =
  | { readonly kind: "always" }
  | { readonly kind: "fact_equals"; readonly fact_id: string; readonly value: DiscreteFact };

export interface SurfaceRow {
  readonly surface_id: string;
  readonly mount_id: string;
  readonly unlock: SurfaceUnlock;
}

export interface SurfaceInstance<T = unknown> {
  subscribe(listener: (value: T) => void): () => void;
  unmount(): void;
}

export type SurfaceFactory<T = unknown> = (mount: HTMLElement) => SurfaceInstance<T>;

const identifier = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

function exactObject(value: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError(`${label} must be an object`);
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
  return value as Record<string, unknown>;
}

function parseFact(value: unknown, label: string): DiscreteFact {
  if (typeof value === "boolean" || typeof value === "string") return value;
  if (typeof value === "number" && Number.isSafeInteger(value)) return value;
  throw new SyntaxError(`${label} must be a boolean, string, or safe integer`);
}

export function parseSurfaceRegistry(source: unknown, knownFactIDs: ReadonlySet<string>): readonly SurfaceRow[] {
  if (!Array.isArray(source) || source.length === 0) throw new SyntaxError("surface registry must be a non-empty array");
  const rows = source.map((raw, index): SurfaceRow => {
    const row = exactObject(raw, ["mount_id", "surface_id", "unlock"], `surfaces[${index}]`);
    if (typeof row.surface_id !== "string" || !identifier.test(row.surface_id)) throw new SyntaxError(`surfaces[${index}].surface_id is invalid`);
    if (typeof row.mount_id !== "string" || !identifier.test(row.mount_id)) throw new SyntaxError(`surfaces[${index}].mount_id is invalid`);
    if (row.unlock === null || typeof row.unlock !== "object" || Array.isArray(row.unlock)) throw new SyntaxError(`surfaces[${index}].unlock is invalid`);
    const kind = (row.unlock as Record<string, unknown>).kind;
    if (kind === "always") {
      exactObject(row.unlock, ["kind"], `surfaces[${index}].unlock`);
      return Object.freeze({ surface_id: row.surface_id, mount_id: row.mount_id, unlock: Object.freeze({ kind }) });
    }
    const unlock = exactObject(row.unlock, ["fact_id", "kind", "value"], `surfaces[${index}].unlock`);
    if (unlock.kind !== "fact_equals" || typeof unlock.fact_id !== "string" || !identifier.test(unlock.fact_id) || !knownFactIDs.has(unlock.fact_id)) {
      throw new SyntaxError(`surfaces[${index}].unlock references an unknown fact`);
    }
    return Object.freeze({
      surface_id: row.surface_id,
      mount_id: row.mount_id,
      unlock: Object.freeze({ kind: "fact_equals", fact_id: unlock.fact_id, value: parseFact(unlock.value, `surfaces[${index}].unlock.value`) }),
    });
  });
  for (let index = 1; index < rows.length; index += 1) {
    if (rows[index - 1].surface_id >= rows[index].surface_id) throw new SyntaxError("surface rows must be byte-sorted and unique");
  }
  if (new Set(rows.map((row) => row.mount_id)).size !== rows.length) throw new SyntaxError("surface mount IDs must be unique");
  return Object.freeze(rows);
}

export function surfaceUnlocked(row: SurfaceRow, facts: Readonly<Record<string, DiscreteFact>>): boolean {
  return row.unlock.kind === "always" || facts[row.unlock.fact_id] === row.unlock.value;
}

export class SurfaceHost<T = unknown> {
  readonly #rows: ReadonlyMap<string, SurfaceRow>;
  readonly #mounts: ReadonlyMap<string, HTMLElement>;
  readonly #factories: ReadonlyMap<string, SurfaceFactory<T>>;
  #active: { readonly instance: SurfaceInstance<T>; readonly unsubscribe: () => void } | undefined;

  constructor(rows: readonly SurfaceRow[], mounts: ReadonlyMap<string, HTMLElement>, factories: ReadonlyMap<string, SurfaceFactory<T>>) {
    this.#rows = new Map(rows.map((row) => [row.surface_id, row]));
    this.#mounts = mounts;
    this.#factories = factories;
    for (const row of rows) {
      if (!mounts.has(row.mount_id) || !factories.has(row.surface_id)) throw new RangeError(`surface ${row.surface_id} is not composed`);
    }
  }

  activate(surfaceID: string, facts: Readonly<Record<string, DiscreteFact>>, listener: (value: T) => void): void {
    const row = this.#rows.get(surfaceID);
    if (!row) throw new RangeError(`unknown surface ${surfaceID}`);
    if (!surfaceUnlocked(row, facts)) throw new RangeError(`surface ${surfaceID} is locked`);
    this.disposeActive();
    const instance = this.#factories.get(surfaceID)!(this.#mounts.get(row.mount_id)!);
    let unsubscribe: () => void;
    try { unsubscribe = instance.subscribe(listener); }
    catch (error) { instance.unmount(); throw error; }
    this.#active = { instance, unsubscribe };
  }

  disposeActive(): void {
    if (!this.#active) return;
    const active = this.#active;
    this.#active = undefined;
    try { active.unsubscribe(); }
    finally { active.instance.unmount(); }
  }

  dispose(): void { this.disposeActive(); }
}
