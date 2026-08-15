import { isStateValue, MAX_EXACT_INTEGER, parseCanonical } from "./numeric";
import type { EconomyCatalog } from "./economy-kernel";

export type CurriculumBranchName = "acquihire" | "burnout" | "pivot";
export type CurriculumStarter =
  | { readonly kind: "resource_grant"; readonly resource_id: string; readonly amount: string }
  | { readonly kind: "generated_generators"; readonly generator_id: string; readonly count: number }
  | { readonly kind: "preowned_upgrade"; readonly upgrade_id: string };
export interface CurriculumBranch {
  readonly branch: CurriculumBranchName;
  readonly minimum_purchased_generators?: number;
  readonly minimum_owned_upgrades?: number;
  readonly cheapest_price_factor?: string;
  readonly route_knowledge_bonus: number;
  readonly starter_package: CurriculumStarter;
  readonly title_key: string;
  readonly body_key: string;
}
export interface CurriculumCatalog {
  readonly schema_version: 1;
  readonly first_failure: {
    readonly run_seq: 1; readonly founder_exit_count: 0; readonly attended_ms: 900000;
    readonly gate_id: "gate.t0_to_t1"; readonly exit_type: "scripted_first";
    readonly evaluation: "first_player_company_command_after_accrual";
    readonly requested_command_effect: "replaced_by_terminal_transition";
    readonly next_run_key: string; readonly branches: readonly CurriculumBranch[];
  };
}

const mechanicalID = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export function parseCurriculumCatalog(source: unknown, economy: EconomyCatalog, copyKeys: ReadonlySet<string>, gateIds: readonly string[]): CurriculumCatalog {
  const root = exact(source, ["schema_version", "first_failure"]);
  const failure = exact(root.first_failure, ["run_seq", "founder_exit_count", "attended_ms", "gate_id", "exit_type", "evaluation", "requested_command_effect", "next_run_key", "branches"]);
  if (root.schema_version !== 1 || failure.run_seq !== 1 || failure.founder_exit_count !== 0 || failure.attended_ms !== 900000 || failure.gate_id !== "gate.t0_to_t1" ||
      failure.exit_type !== "scripted_first" || failure.evaluation !== "first_player_company_command_after_accrual" || failure.requested_command_effect !== "replaced_by_terminal_transition" ||
      typeof failure.next_run_key !== "string" || !copyKeys.has(failure.next_run_key) || !gateIds.includes(failure.gate_id) || !Array.isArray(failure.branches) || failure.branches.length !== 3) throw new SyntaxError("invalid curriculum catalog");
  const expected: CurriculumBranchName[] = ["acquihire", "burnout", "pivot"];
  const branches = failure.branches.map((value, index): CurriculumBranch => {
    if (!value || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError("invalid curriculum branch");
    const raw = value as Record<string, unknown>; const branch = expected[index]!;
    const common = ["branch", "route_knowledge_bonus", "starter_package", "title_key", "body_key"];
    const extras = branch === "acquihire" ? ["minimum_purchased_generators", "minimum_owned_upgrades"] : branch === "burnout" ? ["cheapest_price_factor"] : [];
    exact(raw, [...common, ...extras]);
    if (raw.branch !== branch || typeof raw.title_key !== "string" || typeof raw.body_key !== "string" || !copyKeys.has(raw.title_key) || !copyKeys.has(raw.body_key) ||
        !Number.isSafeInteger(raw.route_knowledge_bonus) || (raw.route_knowledge_bonus as number) < 0) throw new SyntaxError("invalid curriculum branch binding");
    let starter: CurriculumStarter;
    if (branch === "acquihire") {
      if (!positiveSafe(raw.minimum_purchased_generators) || !positiveSafe(raw.minimum_owned_upgrades)) throw new SyntaxError("invalid acquihire branch");
      const item = exact(raw.starter_package, ["kind", "resource_id", "amount"]);
      if (item.kind !== "resource_grant" || typeof item.resource_id !== "string" || !mechanicalID.test(item.resource_id) || typeof item.amount !== "string") throw new SyntaxError("invalid acquihire starter");
      const resource = economy.resource(item.resource_id); const amount = parseCanonical(item.amount);
      if (resource?.scope !== "company" || !amount.gt(0) || !isStateValue(amount)) throw new SyntaxError("invalid acquihire starter");
      starter = { kind: "resource_grant", resource_id: item.resource_id, amount: item.amount };
    } else if (branch === "burnout") {
      if (typeof raw.cheapest_price_factor !== "string") throw new SyntaxError("invalid burnout branch");
      const factor = parseCanonical(raw.cheapest_price_factor);
      if (!factor.gt(0) || !isStateValue(factor)) throw new SyntaxError("invalid burnout branch");
      const item = exact(raw.starter_package, ["kind", "generator_id", "count"]);
      if (item.kind !== "generated_generators" || typeof item.generator_id !== "string" || !mechanicalID.test(item.generator_id) || !positiveSafe(item.count)) throw new SyntaxError("invalid burnout starter");
      const generator = economy.generatorClass(item.generator_id);
      if (!generator || economy.resource(generator.price.resourceId)?.scope !== "company") throw new SyntaxError("invalid burnout starter");
      starter = { kind: "generated_generators", generator_id: item.generator_id, count: item.count as number };
    } else {
      const item = exact(raw.starter_package, ["kind", "upgrade_id"]);
      const upgrade = typeof item.upgrade_id === "string" ? economy.upgrade(item.upgrade_id) : undefined;
      if (item.kind !== "preowned_upgrade" || typeof item.upgrade_id !== "string" || !mechanicalID.test(item.upgrade_id) || !upgrade || economy.resource(upgrade.cost.resourceId)?.scope !== "company") throw new SyntaxError("invalid pivot starter");
      starter = { kind: "preowned_upgrade", upgrade_id: item.upgrade_id };
    }
    return Object.freeze({ branch, ...(branch === "acquihire" ? { minimum_purchased_generators: raw.minimum_purchased_generators as number, minimum_owned_upgrades: raw.minimum_owned_upgrades as number } : branch === "burnout" ? { cheapest_price_factor: raw.cheapest_price_factor as string } : {}), route_knowledge_bonus: raw.route_knowledge_bonus as number, starter_package: starter, title_key: raw.title_key, body_key: raw.body_key });
  });
  return Object.freeze({ schema_version: 1, first_failure: Object.freeze({ run_seq: 1, founder_exit_count: 0, attended_ms: 900000, gate_id: "gate.t0_to_t1", exit_type: "scripted_first", evaluation: "first_player_company_command_after_accrual", requested_command_effect: "replaced_by_terminal_transition", next_run_key: failure.next_run_key, branches: Object.freeze(branches) }) });
}

function positiveSafe(value: unknown): boolean {
  return Number.isSafeInteger(value) && (value as number) >= 1 && (value as number) <= MAX_EXACT_INTEGER;
}

function exact(source: unknown, keys: readonly string[]): Record<string, unknown> {
  if (!source || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError("invalid curriculum object");
  const record = source as Record<string, unknown>;
  if (Object.keys(record).sort().join("\0") !== [...keys].sort().join("\0")) throw new SyntaxError("invalid curriculum keys");
  return record;
}
