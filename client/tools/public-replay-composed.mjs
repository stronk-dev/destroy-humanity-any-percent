// Existing Go/Postgres replay witnesses pass pinned inputs on stdin, never disk.
// The explicit Guild-transition mode has no HTTP/publication claim. The default
// mode downloads public evidence but not catalogs: the full public loop is open.
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { gunzipSync } from "node:zlib";
import { createServer } from "vite";

const clientRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
let input = "";
for await (const chunk of process.stdin) input += chunk;
const fixture = JSON.parse(input);
const vite = await createServer({
  configFile: path.join(clientRoot, "vite.config.ts"), root: clientRoot,
  server: { middlewareMode: true }, appType: "custom",
});
try {
  if (fixture.kind === "stored_guild_transition") {
    let stage = "load";
    try {
      const { loadReplayCatalogBundle, restoreReplayState, encodeReplayState, applyLogged, canonicalJSONString } = await vite.ssrLoadModule("/src/replay.ts");
      const artifacts = Object.fromEntries(Object.entries(fixture.artifacts)
        .map(([name, bytes]) => [name, Buffer.from(bytes, "base64").toString("utf8")]));
      const catalogs = await loadReplayCatalogBundle(fixture.constants_hash, artifacts);
      const restore = (raw) => restoreReplayState(raw, fixture.version, catalogs.economy, {
        meters: catalogs.meters, achievements: catalogs.achievements,
        doctrines: catalogs.doctrines, opportunities: catalogs.opportunities,
      });
      stage = "positive";
      const state = restore(fixture.Before);
      assert.equal(canonicalJSONString(encodeReplayState(state)), canonicalJSONString(fixture.Before));
      const result = await applyLogged(state, canonicalJSONString(fixture.Payload), catalogs, fixture.Inputs);
      assert.equal(result.outcome, "applied");
      for (const [actual, expected] of [[encodeReplayState(result.state), fixture.After], [result.receipt, fixture.Receipt], [result.events, fixture.Events]]) {
        assert.equal(canonicalJSONString(actual), canonicalJSONString(expected));
      }
      assert.equal(canonicalJSONString(encodeReplayState(restore(fixture.After))), canonicalJSONString(fixture.After));
      stage = "omission";
      const omitted = await applyLogged(restore(fixture.Before), canonicalJSONString(fixture.Payload), catalogs, fixture.Omitted);
      assert.equal(omitted.outcome, "applied");
      assert.equal(omitted.state.guildBoundarySeq, 0);
      assert.notEqual(omitted.state.consumedStockUnits, result.state.consumedStockUnits);
      assert.notEqual(canonicalJSONString(encodeReplayState(omitted.state)), canonicalJSONString(fixture.After));
      stage = "duplicate";
      const refused = restore(fixture.Before);
      const before = canonicalJSONString(encodeReplayState(refused));
      await assert.rejects(applyLogged(refused, canonicalJSONString(fixture.Payload), catalogs, fixture.Duplicate));
      assert.equal(canonicalJSONString(encodeReplayState(refused)), before);
      console.log("stored live Guild inputs → TypeScript replay: complete state/receipt/events + omission/duplicate controls: PASS");
    } catch {
      // Node assertion diagnostics include compared values. Never print private
      // saved heads/replay payloads, even when this observation fails.
      throw new Error(`stored Guild replay failed at ${stage}`);
    }
  } else if (fixture.kind === undefined) {
  const { createAPIClient } = await vite.ssrLoadModule("/src/api/generated/types.ts");
  const { loadReplayCatalogBundle, verifyReplayRun, canonicalJSONString } = await vite.ssrLoadModule("/src/replay.ts");
  const { KERNEL_VERSION } = await vite.ssrLoadModule("/src/kernel/version.ts");
  const calls = [];
  const api = createAPIClient(async (url, init) => {
    assert.equal(new Headers(init.headers).has("Authorization"), false, "public download sent credentials");
    calls.push(new URL(url).pathname);
    return fetch(url, init);
  }, fixture.base_url);
  const request = { path: { stream: fixture.stream, seq: 1 }, request: null, signal: AbortSignal.timeout(30_000) };
  const manifestResult = await api.call("get_public_run_verdict", request);
  assert.equal(manifestResult.status, 200, "verified run manifest was unavailable");
  const manifest = manifestResult.body;
  const base = `/api/public/v1/runs/${fixture.stream}/1`;
  assert.equal(manifest.run_id, `${fixture.stream}:1`);
  assert.equal(manifest.verdict, "verified");
  assert.equal(manifest.constants_hash, fixture.constants_hash);
  assert.equal(manifest.engine_version, KERNEL_VERSION);
  assert.equal(manifest.genesis_url, `${base}/genesis`);
  assert.equal(manifest.replay_log_url, `${base}/replay-log`);
  assert.equal(manifest.catalog_url, `/api/public/v1/catalogs/${fixture.constants_hash}`);
  const genesisResult = await api.call("get_public_run_genesis", request);
  const archiveResult = await api.call("get_public_run_replay_log", request);
  assert.deepEqual(calls, [`${base}/verdict`, manifest.genesis_url, manifest.replay_log_url]);
  for (const [result, expected, hash, contentType] of [
    [genesisResult, fixture.genesis, manifest.genesis_sha256, "application/json"],
    [archiveResult, fixture.archive, manifest.replay_log_sha256, "application/gzip"],
  ]) {
    assert.equal(result.status, 200);
    const bytes = Buffer.from(result.body);
    assert.deepEqual(bytes, Buffer.from(expected, "base64"), "SDK changed immutable evidence bytes");
    const digest = createHash("sha256").update(bytes).digest("hex");
    assert.equal(hash, `sha256:${digest}`);
    assert.equal(result.headers.get("etag"), `"${digest}"`);
    assert.equal(result.headers.get("x-content-sha256"), digest);
    assert.equal(result.headers.get("content-type"), contentType);
    assert.equal(result.headers.get("cache-control"), "public,max-age=31536000,immutable");
    assert.ok(result.headers.get("x-request-id"));
  }
  const genesis = JSON.parse(Buffer.from(genesisResult.body).toString("utf8"));
  const archive = JSON.parse(gunzipSync(archiveResult.body).toString("utf8"));
  assert.equal(archive.company_stream_id, fixture.stream);
  assert.equal(archive.run_seq, 1);
  assert.equal(archive.pin.engine_version, manifest.engine_version);
  assert.equal(archive.pin.constants_hash, manifest.constants_hash);
  assert.deepEqual(archive.genesis.state, genesis);
  assert.ok(archive.entries.length >= 2, "fixture lost its ordinary action or terminal Exit");
  const artifacts = Object.fromEntries(Object.entries(fixture.artifacts)
    .map(([name, bytes]) => [name, Buffer.from(bytes, "base64").toString("utf8")]));
  // The loader hashes these exact stored bytes. No current-catalog fallback.
  const catalogs = await loadReplayCatalogBundle(manifest.constants_hash, artifacts);
  // Same Company projection as VerifyStoredRun; preserve the archive's order.
  // The untouched downloaded gzip above still includes broader Founder events.
  const entries = archive.entries.map((entry) => ({
    seq: entry.seq, canonicalPayload: canonicalJSONString(entry.canonical_payload),
    replayInputs: entry.replay_inputs, receiptJSON: canonicalJSONString(entry.receipt),
    eventsJSON: canonicalJSONString(entry.events
      .filter((event) => event.stream_id === fixture.stream || event.kind === "founder_advanced")
      .map(({ kind, schema_version, intent_id, payload }) => ({ kind, schema_version, intent_id, payload }))),
    terminal: entry.replay_inputs.resolved.kind === "exit" && entry.receipt.outcome === "applied",
  }));
  const identity = { constantsHash: manifest.constants_hash, genesisVersion: archive.genesis.version };
  assert.equal(await verifyReplayRun(genesis, catalogs, entries, identity), "verified", "downloaded run did not verify in TypeScript");
  const corruptReceipt = entries.map((entry, index) => index === 0 ? { ...entry, receiptJSON: "{}" } : entry);
  assert.equal(await verifyReplayRun(genesis, catalogs, corruptReceipt, identity), "state_divergence");
  assert.equal(await verifyReplayRun(genesis, catalogs, entries.slice(1), identity), "log_gap");
  const changedEvents = entries.map((entry, index) => index === 0 ? { ...entry, eventsJSON: "[]" } : entry);
  assert.equal(await verifyReplayRun(genesis, catalogs, changedEvents, identity), "state_divergence");
  assert.equal(await verifyReplayRun(genesis, catalogs, entries, { ...identity, constantsHash: `sha256:${"0".repeat(64)}` }), "constants_mismatch");
  console.log("public SDK HTTP downloads → TypeScript replay: verified + receipt/event/gap/constants refusal; DB-pinned catalogs supplied: PASS");
  } else { throw new Error("unknown composed replay observation"); }
} finally {
  await vite.close();
}
