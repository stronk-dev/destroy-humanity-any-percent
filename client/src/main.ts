import { mount } from "svelte";
import policySource from "../../balance/client-shell/phase0.json";
import App from "./shell/App.svelte";
import { ShellController } from "./shell/controller";
import { parseClientShellPolicy } from "./shell/policy";
import "./shell/styles.css";

const target = document.getElementById("app");
if (!target) throw new Error("missing app mount");
const controller = new ShellController(parseClientShellPolicy(policySource), undefined, matchMedia("(prefers-reduced-motion: reduce)").matches);
mount(App, { target, props: { controller } });
