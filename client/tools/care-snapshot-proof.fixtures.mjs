import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { assertCareSnapshot } from "./care-snapshot-proof.mjs";

const catalog = JSON.parse(readFileSync(new URL("../../balance/pets/first-content.json", import.meta.url), "utf8"));
const hash = "sha256:test-only";
function fixture(attended = 1922) {
  const care = { stats_ppm: { hunger: 1_000_000, energy: 750000, cleanliness: 750000, affection: 750000 },
    stat_decay_remainders_ppm: { hunger: 356000, energy: 356000, cleanliness: 356000, affection: 356000 },
    evaluated_through_attended_ms: 1920, trust_ppm: 520000, cooldown_until_attended_ms: { "care.feed": 3601920 } };
  const receipt = { intent_id: "test-intent", founder_revision: 5, pet_id: "test-pet", action_id: "care.feed", stat_id: "hunger",
    after_ppm: 1_000_000, trust_after_ppm: 520000, next_eligible_attended_ms: 3601920, status_band: "high" };
  const persisted = { revision: 5, constants_hash: hash, age_ms: 0, care, command_attended_ms: 1920,
    receipt: structuredClone(receipt) };
  const view = { constants_hash: hash, founder_revision: 5, features: { opportunity: { attended_now_ms: attended },
    pet_adoption: { pets: [{ pet_id: "test-pet", name_key: "pet.name.server_room_cat.n00", species_id: "pet_species.server_room_cat",
      palette_id: "pet_palette.fur_02", temperament: "sassy", status_band: attended >= 1922 ? "normal" : "high",
      eligible_action_ids: ["care.groom", "care.pet", "care.play", "care.rest"] }] } } };
  return { view, receipt, persisted, catalog: structuredClone(catalog) };
}
function check({ view, receipt, persisted, catalog }) { return assertCareSnapshot(view, receipt, persisted, catalog, hash); }

test("binds high/high/normal to actual read samples without changing the frozen receipt or SQL state", () => {
  for (const attended of [1920, 1921, 1922]) {
    const input = fixture(attended);
    const before = structuredClone(input);
    assert.equal(check(input).status_band, attended < 1922 ? "high" : "normal");
    assert.deepEqual(input, before);
  }
});
test("retained mismatch fails the old equality while the exact later-read oracle passes", () => {
  const input = fixture();
  assert.throws(() => assert.equal(input.view.features.pet_adoption.pets[0].status_band, input.receipt.status_band),
    { code: "ERR_ASSERTION" });
  assert.equal(check(input).status_band, "normal");
});
test("completed Founder attendance and the current watermark clamp are both accounted for", () => {
  const priorClock = fixture(1800);
  assert.equal(check(priorClock).attended_ms, 1920);
  const completed = fixture(1920);
  completed.persisted.age_ms = 2;
  completed.view.features.pet_adoption.pets[0].status_band = "normal";
  assert.equal(check(completed).attended_ms, 1922);
});

for (const [label, corrupt, diagnostic] of [
  ["stale receipt band reused after threshold", (x) => { x.view.features.pet_adoption.pets[0].status_band = "high"; }, /read-time projection/u],
  ["arbitrary later band", (x) => { x.view.features.pet_adoption.pets[0].status_band = "low"; }, /read-time projection/u],
  ["wrong read sample", (x) => { x.view.features.opportunity.attended_now_ms = 1921; }, /read-time projection/u],
  ["missing clock", (x) => { delete x.view.features.opportunity; }, /run attendance/u],
  ["wrong Founder revision", (x) => { x.view.founder_revision = 4; }, /applied Founder revision/u],
  ["wrong SQL head", (x) => { x.persisted.revision = 4; }, /SQL head binds/u],
  ["wrong pinned snapshot", (x) => { x.view.constants_hash = "other"; }, /pinned fixture/u],
  ["wrong pinned SQL head", (x) => { x.persisted.constants_hash = "other"; }, /pinned fixture/u],
  ["unbound stored receipt", (x) => { x.persisted.receipt.intent_id = "other"; }, /actual stored receipt/u],
  ["wrong frozen band in both receipts", (x) => { x.receipt.status_band = x.persisted.receipt.status_band = "normal"; }, /frozen receipt band/u],
  ["wrong command attendance", (x) => { x.persisted.command_attended_ms = 1919; }, /frozen command attendance/u],
  ["lost applied stat", (x) => { x.persisted.care.stats_ppm.hunger = 750000; }, /SQL stat retains/u],
  ["lost Trust grant", (x) => { x.persisted.care.trust_ppm = 500000; }, /SQL Trust retains/u],
  ["lost cooldown", (x) => { delete x.persisted.care.cooldown_until_attended_ms["care.feed"]; }, /SQL cooldown retains/u],
  ["feed falsely eligible", (x) => { x.view.features.pet_adoption.pets[0].eligible_action_ids.push("care.feed"); }, /eligibility matches/u],
  ["eligible action missing", (x) => { x.view.features.pet_adoption.pets[0].eligible_action_ids.pop(); }, /eligibility matches/u],
  ["missing pet", (x) => { x.view.features.pet_adoption.pets = []; }, /contains the adopted pet/u],
  ["private field exposed", (x) => { x.view.features.pet_adoption.pets[0].stats_ppm = {}; }, /only permitted fields/u],
  ["invalid remainder", (x) => { x.persisted.care.stat_decay_remainders_ppm.energy = 360000; }, /stored stat remainder/u],
  ["incomplete catalog", (x) => { x.catalog.stat_policy.stats.pop(); }, /all four pinned stat rows/u],
  ["unsafe attendance", (x) => { x.view.features.opportunity.attended_now_ms = Number.MAX_SAFE_INTEGER + 1; }, /run attendance/u],
  ["overflowing combined attendance", (x) => { x.persisted.age_ms = Number.MAX_SAFE_INTEGER; }, /overflows/u],
]) test(`rejects ${label}`, () => {
  const input = fixture(); corrupt(input);
  assert.throws(() => check(input), diagnostic);
});
