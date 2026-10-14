/** C4's ordering projection, not a placement command/wire or battle legality. */
export interface LaneOrderKey {
  readonly tick: number;
  readonly player_slot: number;
  readonly placement_seq: number;
}

/** Equal keys compare equal; duplicate placement policy belongs to the engine. */
export function compareLaneOrder(left: LaneOrderKey, right: LaneOrderKey): -1 | 0 | 1 {
  validate(left);
  validate(right);
  for (const [a, b] of [[left.tick, right.tick], [left.player_slot, right.player_slot], [left.placement_seq, right.placement_seq]]) {
    if (a! < b!) return -1;
    if (a! > b!) return 1;
  }
  return 0;
}

function validate(key: LaneOrderKey): void {
  if (key === null || typeof key !== "object" || Array.isArray(key)) throw new RangeError("invalid combat lane order key");
  for (const value of [key.tick, key.player_slot, key.placement_seq]) {
    if (typeof value !== "number" || !Number.isInteger(value) || value < -2_147_483_648 || value > 2_147_483_647) {
      throw new RangeError("invalid combat lane order key");
    }
  }
}
