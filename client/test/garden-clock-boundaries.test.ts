import { expect, it } from "vitest";
import fixture from "../../balance/testdata/server-garden/fixture-v1.json?raw";
import fiscal from "../../balance/testdata/server-garden/fiscal-fixture-v1.json";
import corpus from "../../testdata/garden/catalog-fixtures-v1.json";
import { COPY_KEYS } from "../src/copy";
import { loadGardenCatalog, gardenSubstrate } from "../src/garden/catalog";
import { advanceGarden, encodeGardenState, newGardenState, parseGardenState, type GardenState } from "../src/garden/engine";

const catalog = loadGardenCatalog(fixture, { copyKeys: new Set(COPY_KEYS), resourceIds: new Set(corpus.resource_ids),
  fiscalUnlockIds: new Set(fiscal.unlock_rows.map((row) => row.unlock_id)), fiscalGeneratorIds: new Set(fiscal.generator_level_rows.map((row) => row.generator_id)) });
const anchor = 1_000_000;
const maximum = Number.MAX_SAFE_INTEGER;
interface Boundary { name: string; substrate?: string; at: number; seq?: number; locked?: boolean; initial?: boolean; growing?: boolean; noSalt?: boolean }
const cases: Boundary[] = catalog.substrates.flatMap((substrate) => [
  ["regression", anchor - 1], ["equality", anchor],
  ["tick_minus_one", anchor + substrate.tick_ms - 1], ["tick", anchor + substrate.tick_ms], ["tick_plus_one", anchor + substrate.tick_ms + 1],
  ["cap_minus_one", anchor + catalog.catchupCapMs - 1], ["cap", anchor + catalog.catchupCapMs], ["cap_plus_one", anchor + catalog.catchupCapMs + 1],
  ["cap_plus_tick", anchor + catalog.catchupCapMs + substrate.tick_ms + 1], ["maximum_server_ms", maximum],
].map(([name, at]) => ({ name: `${substrate.substrate_id}/${name}`, substrate: substrate.substrate_id, at: at as number })));
const tick = gardenSubstrate(catalog, "bare_metal")!.tick_ms;
cases.push(
  { name: "locked", at: anchor, locked: true, initial: true }, { name: "initialize_at_zero", at: 0, initial: true },
  { name: "maximum_counter_zero_ticks", at: anchor, seq: maximum },
  { name: "last_safe_tick", at: anchor + tick, seq: maximum - 1 }, { name: "last_two_safe_ticks", at: anchor + 2 * tick, seq: maximum - 2 },
  { name: "overflow_one", at: anchor + tick, seq: maximum }, { name: "overflow_three", at: anchor + 3 * tick, seq: maximum - 1 },
  { name: "overflow_before_growth", at: anchor + tick, seq: maximum, growing: true }, { name: "overflow_before_salt", at: anchor + tick, seq: maximum, noSalt: true },
);

// Independent exact integer model; neither runtime's Advance/corpus output is authority.
function referenceClock(state: GardenState, at: number, unlocked: boolean, salt: string) {
  const post = structuredClone(state);
  const summary = { ticks_applied: 0, tick_seq_after: state.tick_seq, catchup_forfeited_ms: 0,
    catchup_reason_key: null as string | null, matured: [], spawned: [] };
  if (!unlocked) return { post, summary, overflow: null };
  if (state.tick_anchor_wall_ms === null) { post.salt_hex = salt; post.tick_anchor_wall_ms = at; return { post, summary, overflow: null }; }
  if (at <= state.tick_anchor_wall_ms) return { post, summary, overflow: null };
  let elapsed = BigInt(at) - BigInt(state.tick_anchor_wall_ms);
  const cap = BigInt(catalog.catchupCapMs);
  const loss = elapsed > cap ? elapsed - cap : 0n;
  if (elapsed > cap) elapsed = cap;
  const interval = BigInt(gardenSubstrate(catalog, state.substrate_id)!.tick_ms);
  const n = elapsed / interval;
  const next = BigInt(state.tick_seq) + n;
  if (next > BigInt(maximum)) return { post, summary, overflow: next.toString() };
  post.tick_seq = Number(next);
  post.tick_anchor_wall_ms = Number(BigInt(state.tick_anchor_wall_ms) + loss + n * interval);
  summary.ticks_applied = Number(n); summary.tick_seq_after = Number(next);
  summary.catchup_forfeited_ms = Number(loss); summary.catchup_reason_key = loss > 0n ? catalog.catchupReasonKey : null;
  return { post, summary, overflow: null };
}

it("pins the complete 49-case clock population", () => { expect(cases).toHaveLength(49); });
it.each(cases)("matches exact SG3 clock and SG2 domain: $name", (test) => {
  let state = newGardenState(catalog);
  state.tick_seq = test.seq ?? 0;
  state.substrate_id = test.substrate ?? state.substrate_id;
  if (!test.initial) { state.salt_hex = "0123456789abcdef"; state.tick_anchor_wall_ms = anchor; }
  if (test.noSalt) state.salt_hex = null;
  if (test.growing) state.plots.push({ row: 0, col: 0, species_id: "strain_a", age_ticks: 0, matured_effect_ppm: null });
  state = parseGardenState(state, catalog); // Bind the tested frontier to actual admission, not a fabricated invalid input.
  const before = JSON.stringify(encodeGardenState(state));
  const unlocked = !test.locked;
  const salt = unlocked && state.salt_hex === null ? "0123456789abcdef" : "";
  const expected = referenceClock(state, test.at, unlocked, salt);
  if (expected.overflow !== null) {
    let failure: unknown;
    try { advanceGarden(catalog, state, { serverMs: test.at, hostLevel: 0, unlocked, salt }); } catch (error) { failure = error; }
    expect(failure, `out-of-domain exact counter ${expected.overflow}, actual ${state.tick_seq}`).toBeInstanceOf(RangeError);
    expect(JSON.stringify(encodeGardenState(state))).toBe(before);
    return;
  }
  const actual = advanceGarden(catalog, state, { serverMs: test.at, hostLevel: 0, unlocked, salt });
  expect(JSON.stringify(actual)).toBe(JSON.stringify(expected.summary));
  expect(JSON.stringify(encodeGardenState(parseGardenState(state, catalog)))).toBe(JSON.stringify(encodeGardenState(expected.post)));
});
