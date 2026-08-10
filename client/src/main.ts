import { mount } from "svelte";
import App from "./shell/App.svelte";
import "./shell/styles.css";

const target = document.getElementById("app");
if (!target) throw new Error("missing app mount");
mount(App, { target });
