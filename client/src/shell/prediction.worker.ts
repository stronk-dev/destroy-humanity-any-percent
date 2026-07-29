/// <reference lib="webworker" />
import { PredictionMachine } from "./prediction";
import type { WorkerCommand, WorkerOutput } from "./worker-protocol";

let machine: PredictionMachine | undefined;
let timer: number | undefined;
function pulse(monotonicMs: number): void {
  if (!machine) return;
  for (const output of machine.pulse(monotonicMs)) self.postMessage(output satisfies WorkerOutput);
}
self.onmessage = (event: MessageEvent<WorkerCommand>) => {
  const command = event.data;
  if (command.kind === "initialize") { if (timer !== undefined) self.clearInterval(timer); machine = new PredictionMachine(command.policy); machine.initialize(command.snapshot, command.monotonicMs); timer = self.setInterval(() => pulse(performance.now()), command.policy.tickMs); return; }
  if (command.kind === "dispose") { if (timer !== undefined) self.clearInterval(timer); machine = undefined; self.close(); return; }
  if (!machine) throw new Error("prediction worker is not initialized");
  if (command.kind === "authoritative_snapshot") { machine.applyAuthoritative(command.snapshot, command.monotonicMs); return; }
  pulse(command.monotonicMs);
};
