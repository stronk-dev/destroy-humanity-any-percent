export const PET_STAT_IDS = ["hunger", "energy", "cleanliness", "affection"] as const;
export const PET_STATUS_BANDS = ["floor", "low", "normal", "high"] as const;
export const PET_MOODS = ["withdrawn", "restless", "neutral", "engaged"] as const;
export const PET_BEHAVIOR_STATES = ["idle", "care_response", "active", "resting"] as const;
export const PET_BEHAVIOR_EVENTS = ["grid_tick", "care_applied", "care_rejected"] as const;
export const PET_CARE_REJECTION_DETAILS = ["cooldown", "ineligible", "saturated", "unknown_pet", "unknown_action"] as const;

export type PetStatID = typeof PET_STAT_IDS[number];
export type PetStatusBand = typeof PET_STATUS_BANDS[number];
export type PetMood = typeof PET_MOODS[number];
export type PetBehaviorState = typeof PET_BEHAVIOR_STATES[number];
export type PetBehaviorEvent = typeof PET_BEHAVIOR_EVENTS[number];
export type PetCareRejectionDetail = typeof PET_CARE_REJECTION_DETAILS[number];

export const PET_BEHAVIOR_QUEUE_HARDCAP = 8 as const;
export const PET_BEHAVIOR_PRNG_LABEL = "pet.behavior.v1" as const;

export function validPetBehaviorQueueLength(length: number): boolean {
  return Number.isSafeInteger(length) && length >= 0 && length <= PET_BEHAVIOR_QUEUE_HARDCAP;
}
