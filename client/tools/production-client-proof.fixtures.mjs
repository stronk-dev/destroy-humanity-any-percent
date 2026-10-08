import assert from "node:assert/strict";
import test from "node:test";
import { productionClientProof } from "./production-client-proof.mjs";

const files = () => new Map([
  ["/", Buffer.from('<script type="module" src="/assets/index-hash.js"></script><link rel="stylesheet" href="/assets/index-hash.css">')],
  ["/assets/index-hash.js", Buffer.from("entry bytes")],
  ["/assets/index-hash.css", Buffer.from("style bytes")],
  ["/assets/prediction.worker-hash.js", Buffer.from("worker bytes")],
]);
function loaded() {
  const build = files();
  const proof = productionClientProof(build);
  for (const [route, bytes] of build) proof.response(route, 200, bytes);
  return proof;
}
test("requires exact browser-loaded entry, CSS, worker and HTML bytes plus a started worker", () => {
  const proof = loaded();
  proof.worker("/assets/prediction.worker-hash.js");
  proof.response("/api/v1/founder/state", 200, Buffer.from("API is not a static asset"));
  assert.equal(proof.finish().length, 4);
});
test("rejects the old development entry instead of calling it a production build", () => {
  const build = files();
  build.set("/", Buffer.from('<script type="module" src="/@vite/client"></script><script type="module" src="/src/main.ts"></script>'));
  assert.throws(() => productionClientProof(build), /lacks its bundled entry/u);
});
test("rejects source requests, stale or changed asset bytes and failed responses", () => {
  const proof = productionClientProof(files());
  assert.throws(() => proof.response("/src/main.ts", 200, "source"), /built client bytes/u);
  assert.throws(() => proof.response("/@vite/client", 200, "HMR"), /built client bytes/u);
  assert.throws(() => proof.response("/assets/index-hash.js", 200, "stale bytes"), /built client bytes/u);
  assert.throws(() => proof.response("/assets/index-hash.js", 404, "entry bytes"), /built client bytes/u);
});
test("accepts cache revalidation only after matching 200 bytes in the same observation", () => {
  const proof = productionClientProof(files());
  assert.throws(() => proof.response("/assets/index-hash.js", 304), /built client bytes/u);
  assert.throws(() => proof.response("/assets/index-hash.js", 302, "entry bytes"), /built client bytes/u);
  proof.response("/assets/index-hash.js", 200, "entry bytes");
  proof.response("/assets/index-hash.js", 304);
  // Revalidation alone cannot satisfy the rest of the journey's population.
  assert.throws(() => proof.finish(), /incomplete/u);
});
test("rejects vacuous observations, missing assets and a worker never started", () => {
  assert.throws(() => productionClientProof(files()).finish(), /incomplete/u);
  assert.throws(() => loaded().finish(), /worker_started=false/u);
  const proof = productionClientProof(files());
  proof.worker("/assets/prediction.worker-hash.js");
  proof.response("/", 200, files().get("/"));
  assert.throws(() => proof.finish(), /missing=.*index-hash/u);
  assert.throws(() => proof.worker("/src/shell/prediction.worker.ts"), /non-bundled/u);
});
