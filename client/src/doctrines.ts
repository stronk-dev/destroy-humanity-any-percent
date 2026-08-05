import type { RoutesCatalog } from "./routes";

export const DOCTRINES_SCHEMA_VERSION = 1;

export interface DoctrineTransition {
  readonly transitionId: string;
  readonly sourceTier: number;
  readonly gateId: string;
  readonly doctrineIds: readonly string[];
}

export class DoctrineCatalog {
  readonly transitions: readonly DoctrineTransition[];
  readonly #byId: ReadonlyMap<string, DoctrineTransition>;

  constructor(transitions: readonly DoctrineTransition[]) {
    this.transitions = Object.freeze([...transitions]);
    this.#byId = new Map(transitions.map((value) => [value.transitionId, value]));
  }

  transition(id: string): DoctrineTransition | undefined { return this.#byId.get(id); }
  allows(transitionId: string, doctrineId: string): boolean { return this.#byId.get(transitionId)?.doctrineIds.includes(doctrineId) ?? false; }
}

const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const transitionPattern = /^transition\.t([0-9]+)_to_t([0-9]+)$/;
const gatePattern = /^gate\.t([0-9]+)_to_t([0-9]+)$/;

export function loadDoctrineCatalog(source: string | unknown): DoctrineCatalog {
  const parsed = typeof source === "string" ? parseJSON(source) : source;
  const root = exactObject(parsed, ["schema_version", "transitions"], "doctrine catalog");
  if (root.schema_version !== DOCTRINES_SCHEMA_VERSION || !Array.isArray(root.transitions) || root.transitions.length === 0) throw new SyntaxError("invalid doctrine catalog root");
  let previous = "";
  const transitions = root.transitions.map((source, index) => {
    const raw = exactObject(source, ["transition_id", "source_tier", "gate_id", "doctrine_ids"], `transitions[${index}]`);
    const transitionId = mechanical(raw.transition_id, "transition_id");
    const gateId = mechanical(raw.gate_id, "gate_id");
    const sourceTier = safeTier(raw.source_tier);
    if (previous !== "" && byteCompare(previous, transitionId) >= 0) throw new SyntaxError("transitions must be byte-sorted and unique");
    previous = transitionId;
    if (boundary(transitionPattern, transitionId) !== sourceTier || boundary(gatePattern, gateId) !== sourceTier || !Array.isArray(raw.doctrine_ids) || raw.doctrine_ids.length < 2) throw new SyntaxError("invalid doctrine transition boundary");
    let priorDoctrine = "";
    const doctrineIds = raw.doctrine_ids.map((value) => {
      const id = mechanical(value, "doctrine_id");
      if (priorDoctrine !== "" && byteCompare(priorDoctrine, id) >= 0) throw new SyntaxError("doctrine_ids must be byte-sorted and unique");
      priorDoctrine = id;
      return id;
    });
    return Object.freeze({ transitionId, sourceTier, gateId, doctrineIds: Object.freeze(doctrineIds) });
  });
  return new DoctrineCatalog(transitions);
}

export function validateDoctrineRoutes(catalog: DoctrineCatalog, routes: RoutesCatalog): void {
  for (const transition of catalog.transitions) if (!routes.gate(transition.gateId)) throw new SyntaxError(`doctrine gate ${transition.gateId} is absent from routes`);
  for (const gate of routes.gates) for (const route of gate.routes) for (const condition of route.predicate) {
    if ((condition.kind === "doctrine_is" || condition.kind === "doctrine_is_not") && !catalog.allows(condition.transition, condition.doctrineId)) throw new SyntaxError(`route ${route.routeId} references an undeclared doctrine`);
  }
}

function boundary(pattern: RegExp, value: string): number | undefined {
  const match = pattern.exec(value);
  if (!match) return undefined;
  const from = Number(match[1]); const to = Number(match[2]);
  return Number.isSafeInteger(from) && from >= 0 && from <= 8 && to === from + 1 ? from : undefined;
}
function parseJSON(source: string): unknown { const parsed = JSON.parse(source); return parsed; }
function isRecord(source: unknown): source is Record<string, unknown> { return typeof source === "object" && source !== null && !Array.isArray(source); }
function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> { if (!isRecord(source)) throw new SyntaxError(`${label} must be an object`); const actual = Object.keys(source).sort(byteCompare); const expected = [...keys].sort(byteCompare); if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`); return source; }
function mechanical(value: unknown, label: string): string { if (typeof value !== "string" || !idPattern.test(value)) throw new SyntaxError(`${label} must be a mechanical id`); return value; }
function safeTier(value: unknown): number { if (typeof value !== "number" || !Number.isSafeInteger(value) || value < 0 || value > 8) throw new SyntaxError("source_tier must be an exact tier"); return value; }
function byteCompare(left: string, right: string): number { const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right); for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!; return a.length - b.length; }
