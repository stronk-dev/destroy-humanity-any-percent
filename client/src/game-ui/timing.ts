export interface RTASample { readonly serverNowMs: number; readonly runStartedAtMs: number; readonly sampledMonotonicMs: number }

export class RTATimer {
  #sample: RTASample;
  #terminalMS: number | undefined;
  constructor(sample: RTASample) { this.#sample = validateSample(sample); }
  resample(sample: RTASample): void { if (this.#terminalMS === undefined) this.#sample = validateSample(sample); }
  terminal(rtaMS: number): void {
    if (!Number.isSafeInteger(rtaMS) || rtaMS < 0) throw new RangeError("invalid terminal RTA");
    this.#terminalMS = rtaMS;
  }
  elapsed(monotonicMs: number): number {
    if (this.#terminalMS !== undefined) return this.#terminalMS;
    if (!Number.isFinite(monotonicMs) || monotonicMs < this.#sample.sampledMonotonicMs) throw new RangeError("invalid monotonic sample");
    return Math.max(0, this.#sample.serverNowMs - this.#sample.runStartedAtMs + Math.floor(monotonicMs - this.#sample.sampledMonotonicMs));
  }
}

function validateSample(sample: RTASample): RTASample {
  if (!Number.isSafeInteger(sample.serverNowMs) || !Number.isSafeInteger(sample.runStartedAtMs) || sample.runStartedAtMs < 1 || sample.serverNowMs < sample.runStartedAtMs ||
      !Number.isFinite(sample.sampledMonotonicMs) || sample.sampledMonotonicMs < 0) throw new RangeError("invalid RTA sample");
  return Object.freeze({ ...sample });
}

export interface LocalRunTiming {
  readonly category: string;
  readonly founder_id: string;
  readonly pb_rta_ms: number;
  readonly run_seq: number;
  readonly splits: readonly Readonly<{ gate_id: string; rta_ms: number }>[];
}
export interface LocalTimingDocument { readonly schema_version: 1; readonly records: readonly LocalRunTiming[] }

export function parseLocalTiming(source: string | null): LocalTimingDocument {
  if (source === null) return { schema_version: 1, records: [] };
  try {
    const value = JSON.parse(source) as LocalTimingDocument;
    if (value?.schema_version !== 1 || !Array.isArray(value.records)) throw new SyntaxError();
    const seen = new Set<string>();
    for (const record of value.records) {
      const key = `${record.founder_id}\0${record.run_seq}\0${record.category}`;
      if (seen.has(key) || !Number.isSafeInteger(record.run_seq) || record.run_seq < 1 || !Number.isSafeInteger(record.pb_rta_ms) || record.pb_rta_ms < 0 || !Array.isArray(record.splits)) throw new SyntaxError();
      seen.add(key);
      let prior = "";
      for (const split of record.splits) {
        if (typeof split.gate_id !== "string" || split.gate_id <= prior || !Number.isSafeInteger(split.rta_ms) || split.rta_ms < 0) throw new SyntaxError();
        prior = split.gate_id;
      }
    }
    return Object.freeze({ schema_version: 1, records: Object.freeze(value.records.map((record: LocalRunTiming) => Object.freeze({
      ...record,
      splits: Object.freeze(record.splits.map((split: Readonly<{ gate_id: string; rta_ms: number }>) => Object.freeze({ ...split }))),
    }))) });
  } catch {
    return { schema_version: 1, records: [] };
  }
}

export function timingStorageKey(founderID: string): string { return `cloud-clicker.timing.v1.${founderID}`; }
