interface IntentEnvelope { readonly intentId: string; readonly expectedRevision: number }
export type AuthoritativeIntent =
  | (IntentEnvelope & { readonly kind: "buy_generator"; readonly generatorId: string; readonly count: { readonly mode: "exact"; readonly value: number } | { readonly mode: "max" } })
  | (IntentEnvelope & { readonly kind: "perform_manual_batch"; readonly actionId: string; readonly count: number; readonly windowMs: number })
  | (IntentEnvelope & { readonly kind: "cross_gate"; readonly gateId: string; readonly routeId: string | null })
  | (IntentEnvelope & { readonly kind: "buy_route_hint"; readonly routeId: string })
  | (IntentEnvelope & { readonly kind: "sign_compact"; readonly tithePpm: number })
  | (IntentEnvelope & { readonly kind: "leave_compact" });
export interface IntentRequestAdapter { request(intent: AuthoritativeIntent): Promise<void> }

const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const intentPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

export class IntentDispatcher {
  readonly #adapter: IntentRequestAdapter;
  constructor(adapter: IntentRequestAdapter) { this.#adapter = adapter; }
  send(intent: AuthoritativeIntent): Promise<void> { validateIntent(intent); return this.#adapter.request(intent); }
}

function validateIntent(intent: AuthoritativeIntent): void {
  if (!intentPattern.test(intent.intentId) || !Number.isSafeInteger(intent.expectedRevision) || intent.expectedRevision <= 0) throw new SyntaxError("invalid intent envelope");
  if (intent.kind === "buy_generator") {
    requireExactKeys(intent, ["intentId", "kind", "expectedRevision", "generatorId", "count"]); requireId(intent.generatorId);
    requireExactKeys(intent.count, intent.count.mode === "exact" ? ["mode", "value"] : ["mode"]);
    if (intent.count.mode === "exact" && (!Number.isSafeInteger(intent.count.value) || intent.count.value <= 0)) throw new SyntaxError("invalid exact generator count");
    if (intent.count.mode !== "exact" && intent.count.mode !== "max") throw new SyntaxError("invalid generator count mode");
    return;
  }
  if (intent.kind === "perform_manual_batch") {
    requireExactKeys(intent, ["intentId", "kind", "expectedRevision", "actionId", "count", "windowMs"]); requireId(intent.actionId); requirePositiveInteger(intent.count); requirePositiveInteger(intent.windowMs); return;
  }
  if (intent.kind === "cross_gate") {
    requireExactKeys(intent, ["intentId", "kind", "expectedRevision", "gateId", "routeId"]); requireId(intent.gateId); if (intent.routeId !== null) requireId(intent.routeId); return;
  }
  if (intent.kind === "buy_route_hint") {
    requireExactKeys(intent, ["intentId", "kind", "expectedRevision", "routeId"]); requireId(intent.routeId); return;
  }
  if (intent.kind === "sign_compact") {
    requireExactKeys(intent, ["intentId", "kind", "expectedRevision", "tithePpm"]); requirePositiveInteger(intent.tithePpm); if (intent.tithePpm > 1_000_000) throw new SyntaxError("invalid tithe ppm"); return;
  }
  requireExactKeys(intent, ["intentId", "kind", "expectedRevision"]);
}

function requireExactKeys(source: object, keys: readonly string[]): void {
  const actual = Object.keys(source).sort(); const expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError("intent fields are not exact");
}
function requireId(value: string): void { if (!idPattern.test(value)) throw new SyntaxError("invalid mechanical id"); }
function requirePositiveInteger(value: number): void { if (!Number.isSafeInteger(value) || value <= 0) throw new SyntaxError("invalid positive integer"); }
