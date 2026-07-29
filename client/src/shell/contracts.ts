import { parseCanonical } from "../numeric";

export interface ResourceCap { readonly amount: string; readonly reasonKey: string }
export interface ResourceState { readonly amount: string; readonly ratePerSecond: string; readonly cap?: ResourceCap }
export type DiscreteFact = boolean | number | string;
export interface ProgressCoordinate { readonly stageId: string; readonly current: string; readonly target: string }
export interface AuthoritativeSnapshot {
  readonly revision: number; readonly evaluatedThroughMs: number; readonly constantsHash: string;
  readonly resources: Readonly<Record<string, ResourceState>>;
  readonly discrete: Readonly<Record<string, DiscreteFact>>;
  readonly progress: readonly ProgressCoordinate[];
}
export interface IntentReceipt { readonly revision: number; readonly intentId: string; readonly status: "applied" | "rejected"; readonly rejectionCode?: string }
export interface SnapshotStream { subscribe(consumer: (snapshot: AuthoritativeSnapshot, receipt?: IntentReceipt) => void): () => void; requestSnapshot(): void; flush(): void; dispose(): void }

const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const hashPattern = /^sha256:[0-9a-f]{64}$/;
const intentPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

function requireExactKeys(source: object, keys: readonly string[], name: string): void {
  const actual = Object.keys(source).sort();
  const expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) {
    throw new SyntaxError(`${name} fields are not exact`);
  }
}

export function validateSnapshot(snapshot: AuthoritativeSnapshot): AuthoritativeSnapshot {
  requireExactKeys(snapshot, ["revision", "evaluatedThroughMs", "constantsHash", "resources", "discrete", "progress"], "snapshot");
  if (!Number.isSafeInteger(snapshot.revision) || snapshot.revision < 0 || !Number.isSafeInteger(snapshot.evaluatedThroughMs) || snapshot.evaluatedThroughMs < 0 || !hashPattern.test(snapshot.constantsHash)) throw new SyntaxError("invalid snapshot envelope");
  for (const [id, resource] of Object.entries(snapshot.resources)) {
    if (!idPattern.test(id)) throw new SyntaxError("invalid resource id");
    requireExactKeys(resource, resource.cap ? ["amount", "ratePerSecond", "cap"] : ["amount", "ratePerSecond"], "resource");
    const amount = parseCanonical(resource.amount); const rate = parseCanonical(resource.ratePerSecond);
    if (amount.lt(0) || rate.lt(0)) throw new SyntaxError("negative resource state");
    if (resource.cap) {
      requireExactKeys(resource.cap, ["amount", "reasonKey"], "resource cap");
      const cap = parseCanonical(resource.cap.amount);
      if (cap.lt(0) || amount.gt(cap) || !idPattern.test(resource.cap.reasonKey)) throw new SyntaxError("invalid resource cap");
    }
  }
  for (const [id, fact] of Object.entries(snapshot.discrete)) if (!idPattern.test(id) || !["boolean", "number", "string"].includes(typeof fact) || typeof fact === "number" && !Number.isSafeInteger(fact)) throw new SyntaxError("invalid discrete fact");
  const stages = new Set<string>();
  for (const item of snapshot.progress) {
    requireExactKeys(item, ["stageId", "current", "target"], "progress coordinate");
    if (!idPattern.test(item.stageId) || stages.has(item.stageId) || parseCanonical(item.current).lt(0) || parseCanonical(item.target).lte(0)) throw new SyntaxError("invalid progress coordinate");
    stages.add(item.stageId);
  }
  return snapshot;
}

export function validateReceipt(receipt: IntentReceipt): IntentReceipt {
  requireExactKeys(receipt, receipt.status === "rejected" ? ["revision", "intentId", "status", "rejectionCode"] : ["revision", "intentId", "status"], "intent receipt");
  if (!Number.isSafeInteger(receipt.revision) || receipt.revision < 0 || !intentPattern.test(receipt.intentId)) throw new SyntaxError("invalid intent receipt envelope");
  if (receipt.status === "rejected") {
    if (!receipt.rejectionCode || !idPattern.test(receipt.rejectionCode)) throw new SyntaxError("invalid rejection code");
  } else if (receipt.status !== "applied") throw new SyntaxError("invalid intent receipt status");
  return receipt;
}
