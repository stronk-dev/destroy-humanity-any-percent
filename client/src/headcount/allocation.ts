/** Headcount S2 arithmetic only. Not a catalog codec, seat source or intent. */
export type Role = Readonly<{
  id: string;
  kind: "produce" | "convert" | "pool_feed";
  from_role_id?: string;
  heads_per_converter?: number;
}>;

export type Allocation = Readonly<{
  producing_heads: Readonly<Record<string, number>>;
  diverted: Readonly<Record<string, number>>;
  idle_converters: Readonly<Record<string, number>>;
  unassigned: number;
}>;

const mechanicalID = /^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$/u;
const count = (value: number): boolean => Number.isSafeInteger(value) && value >= 0;

/** Complete-vector projection. Inputs remain unchanged; results are fresh. */
export function project(hardcap: number, seats: number, roles: readonly Role[], assigned: Readonly<Record<string, number>>): Allocation {
  const invalid = (): never => { throw new RangeError("invalid headcount allocation"); };
  if (!count(hardcap) || hardcap === 0 || !count(seats) || seats > hardcap || roles.length === 0 || Object.keys(assigned).length !== roles.length) invalid();
  const ordered = [...roles].sort((a, b) => a.id < b.id ? -1 : a.id > b.id ? 1 : 0);
  const byID = new Map<string, Role>();
  let remaining = seats;
  for (const role of ordered) {
    if (!mechanicalID.test(role.id) || byID.has(role.id) || !Object.hasOwn(assigned, role.id)) invalid();
    const heads = assigned[role.id];
    // Reject before adding/subtracting can leave the safe-integer domain.
    if (!count(heads) || heads > remaining) invalid();
    remaining -= heads;
    if (role.kind === "produce" || role.kind === "pool_feed") {
      if ((role.from_role_id ?? "") !== "" || (role.heads_per_converter ?? 0) !== 0) invalid();
    } else if (role.kind === "convert") {
      if (!count(role.heads_per_converter ?? -1) || role.heads_per_converter === 0) invalid();
    } else invalid();
    byID.set(role.id, role);
  }
  // Null-prototype records preserve mechanical IDs such as "constructor".
  const producing: Record<string, number> = Object.create(null);
  const diverted: Record<string, number> = Object.create(null);
  const idle: Record<string, number> = Object.create(null);
  for (const role of ordered) if (role.kind !== "convert") producing[role.id] = assigned[role.id];
  const converted = new Set<string>();
  for (const role of ordered) {
    if (role.kind !== "convert") continue;
    const source = byID.get(role.from_role_id ?? "");
    if (!source || source.kind !== "produce" || converted.has(source.id)) invalid();
    const sourceID = source!.id;
    converted.add(sourceID);
    const heads = assigned[sourceID], converters = assigned[role.id], k = role.heads_per_converter!;
    const redirected = converters > Math.floor(heads / k) ? heads : converters * k;
    producing[sourceID] = heads - redirected;
    diverted[role.id] = redirected;
    idle[role.id] = converters - Math.ceil(redirected / k);
  }
  return { producing_heads: producing, diverted, idle_converters: idle, unassigned: remaining };
}
