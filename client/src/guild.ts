export interface GuildCatalog {
  readonly guildTithePpm: number;
  readonly clearingRatePpm: number;
  readonly npcExchangePpm: number;
  readonly stockIntakeCap: number;
  readonly consumptionBonusPpmPerUnit: number;
  readonly maxMembers: number;
  readonly minMembers: number;
  readonly graceDays: number;
  readonly guildXpTargetPerFounder: number;
  readonly clearingIntervalMs: number;
}

export function parseGuildCatalog(source: unknown): GuildCatalog {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError("guild catalog must be an object");
  const raw = source as Record<string, unknown>;
  const expected = ["schema_version", "guild_tithe_ppm", "clearing_rate_ppm", "npc_exchange_ppm", "stock_intake_cap", "consumption_bonus_ppm_per_unit", "max_members", "min_members", "grace_days", "guild_xp_target_per_founder", "clearing_interval_ms"].sort();
  if (Object.keys(raw).sort().join("\0") !== expected.join("\0") || raw.schema_version !== 1 || raw.guild_tithe_ppm !== 20_000 ||
      raw.clearing_rate_ppm !== 500_000 || raw.npc_exchange_ppm !== 250_000 || raw.stock_intake_cap !== 120 ||
      raw.consumption_bonus_ppm_per_unit !== 0 || raw.max_members !== 50 || raw.min_members !== 2 || raw.grace_days !== 7 ||
      raw.guild_xp_target_per_founder !== 250_000 || raw.clearing_interval_ms !== 300_000) throw new SyntaxError("invalid guild catalog");
  return Object.freeze({ guildTithePpm: 20_000, clearingRatePpm: 500_000, npcExchangePpm: 250_000, stockIntakeCap: 120,
    consumptionBonusPpmPerUnit: 0, maxMembers: 50, minMembers: 2, graceDays: 7, guildXpTargetPerFounder: 250_000, clearingIntervalMs: 300_000 });
}
