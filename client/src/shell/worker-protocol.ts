import type { AuthoritativeSnapshot } from "./contracts";
import type { ClientShellPolicy } from "./policy";
import type { PredictionOutput } from "./prediction";

export type WorkerCommand =
  | { readonly kind: "initialize"; readonly policy: ClientShellPolicy; readonly snapshot: AuthoritativeSnapshot }
  | { readonly kind: "authoritative_snapshot"; readonly snapshot: AuthoritativeSnapshot }
  | { readonly kind: "clock_pulse"; readonly monotonicMs: number }
  | { readonly kind: "dispose" };
export type WorkerOutput = PredictionOutput;
