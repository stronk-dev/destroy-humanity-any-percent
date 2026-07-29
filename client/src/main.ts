import { mount } from "svelte";
import catalogSource from "../../balance/catalogs/phase0.json";
import policySource from "../../balance/client-shell/phase0.json";
import { parseCatalog } from "./economy-kernel";
import App from "./shell/App.svelte";
import type { AuthoritativeSnapshot, IntentReceipt, SnapshotStream } from "./shell/contracts";
import { ShellController } from "./shell/controller";
import { parseClientShellPolicy } from "./shell/policy";
import { ShellRuntime } from "./shell/runtime";
import "./shell/styles.css";

class PendingSnapshotStream implements SnapshotStream {
  subscribe(_consumer: (snapshot: AuthoritativeSnapshot, receipt?: IntentReceipt) => void): () => void { return () => {}; }
  requestSnapshot(): void {}
  flush(): void {}
  dispose(): void {}
}

const target = document.getElementById("app");
if (!target) throw new Error("missing app mount");
parseCatalog(catalogSource);
const policy = parseClientShellPolicy(policySource);
const controller = new ShellController(policy, undefined, matchMedia("(prefers-reduced-motion: reduce)").matches);
const runtime = new ShellRuntime(controller, policy, new PendingSnapshotStream());
mount(App, { target, props: { controller } });
runtime.start();
