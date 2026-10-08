import assert from "node:assert/strict";

const statIDs = ["hunger", "energy", "cleanliness", "affection"];
const bands = ["floor", "low", "normal", "high"];
const publicKeys = ["eligible_action_ids", "name_key", "palette_id", "pet_id", "species_id", "status_band", "temperament"].sort();

function integer(value, label, maximum = Number.MAX_SAFE_INTEGER) {
  assert.ok(Number.isSafeInteger(value) && value >= 0 && value <= maximum, `care proof: invalid ${label}`);
  return value;
}

// Test oracle only. Use the SQL head and the actual response's public attendance
// sample, never the receipt's earlier band or a later wall-clock guess. The
// independent fixed-grid arithmetic checks the Go projection; no gameplay or
// persisted state is modified here. This is not an oracle for attendance policy.
export function assertCareSnapshot(view, receipt, persisted, catalog, constantsHash) {
  assert.equal(view.constants_hash, constantsHash, "care snapshot uses the pinned fixture");
  assert.equal(persisted.constants_hash, constantsHash, "care SQL head uses the pinned fixture");
  assert.equal(view.founder_revision, receipt.founder_revision, "care snapshot binds to applied Founder revision");
  assert.equal(persisted.revision, receipt.founder_revision, "care SQL head binds to applied Founder revision");
  assert.deepEqual(persisted.receipt, receipt, "care HTTP receipt equals the actual stored receipt");
  const care = persisted.care;
  assert.ok(care && typeof care === "object", "care SQL head contains the adopted pet");
  const watermark = integer(care.evaluated_through_attended_ms, "care watermark");
  assert.equal(watermark, persisted.command_attended_ms, "care watermark binds to frozen command attendance");
  assert.equal(care.stats_ppm?.[receipt.stat_id], receipt.after_ppm, "care SQL stat retains applied receipt outcome");
  assert.equal(care.trust_ppm, receipt.trust_after_ppm, "care SQL Trust retains applied receipt outcome");
  assert.equal(care.cooldown_until_attended_ms?.[receipt.action_id], receipt.next_eligible_attended_ms,
    "care SQL cooldown retains applied receipt outcome");
  const runAttended = integer(view.features?.opportunity?.attended_now_ms, "snapshot run attendance");
  const age = integer(persisted.age_ms, "completed Founder attendance");
  assert.ok(age <= Number.MAX_SAFE_INTEGER - runAttended, "care proof: Founder attendance overflows");
  // The current projector clamps the discarded read clone to its care watermark.
  // RP-365/D-024's shared-clock policy is deliberately not decided by this test.
  const attended = Math.max(age + runAttended, watermark);
  const grid = integer(catalog.stat_policy?.grid_ms, "pinned grid");
  assert.ok(grid > 0, "care proof: pinned grid is zero");
  assert.deepEqual([...catalog.stat_policy.stats.map((row) => row.stat_id)].sort(), [...statIDs].sort(),
    "care proof: all four pinned stat rows are required");
  assert.equal(catalog.mood_policy.length, bands.length, "care proof: all four pinned band thresholds are required");
  const stats = {};
  for (const row of catalog.stat_policy.stats) {
    const value = integer(care.stats_ppm?.[row.stat_id], "stored stat", 1_000_000);
    const floor = integer(row.floor_ppm, "pinned stat floor", 1_000_000);
    const rate = integer(row.decay_ppm_per_grid, "pinned stat decay", 1_000_000);
    const remainder = integer(care.stat_decay_remainders_ppm?.[row.stat_id], "stored stat remainder", grid - 1);
    assert.ok(value >= floor, "care proof: stored stat is below its floor");
    const whole = (BigInt(attended - watermark) * BigInt(rate) + BigInt(remainder)) / BigInt(grid);
    stats[row.stat_id] = whole >= BigInt(value - floor) ? floor : value - Number(whole);
  }
  const bandFor = (values) => {
    const scalar = Math.min(...statIDs.map((id) => values[id]));
    let selected = 0;
    let priorFloor = -1;
    catalog.mood_policy.forEach((row, index) => {
      const floor = integer(row.floor_ppm, "pinned band floor", 1_000_000);
      assert.ok(floor > priorFloor, "care proof: pinned band floors are not ordered");
      priorFloor = floor;
      if (floor <= scalar) selected = index;
    });
    return bands[selected];
  };
  assert.equal(receipt.status_band, bandFor(care.stats_ppm), "care frozen receipt band matches stored action state");
  const expectedBand = bandFor(stats);
  const expectedActions = catalog.actions.filter((action) => {
    assert.ok(statIDs.includes(action.stat_id), "care proof: pinned action has an unknown stat");
    const cooldown = integer(care.cooldown_until_attended_ms?.[action.action_id] ?? 0, "stored cooldown");
    const minimum = integer(action.min_eligible_ppm, "pinned eligibility floor", 1_000_000);
    return cooldown <= attended && stats[action.stat_id] >= minimum && stats[action.stat_id] < 1_000_000;
  }).map((action) => action.action_id).sort();
  const pet = view.features?.pet_adoption?.pets?.find((row) => row.pet_id === receipt.pet_id);
  assert.ok(pet, "care public snapshot contains the adopted pet");
  assert.deepEqual(Object.keys(pet).sort(), publicKeys, "care public snapshot exposes only permitted fields");
  assert.equal(pet.status_band, expectedBand, "care snapshot band matches the actual read-time projection");
  assert.deepEqual(pet.eligible_action_ids, expectedActions, "care snapshot eligibility matches the actual read-time projection");
  return { status_band: expectedBand, eligible_action_ids: expectedActions, attended_ms: attended };
}
