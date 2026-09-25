import { operations, type APIError, type MinigameCommandRequest, type MinigameCurrentResponse, type MinigameSessionResponse, type MinigameSessionResponseActive, type MinigameSessionResponseTerminal } from "../../api/generated/types";

// MinigameSessionPort is the only transport seam the minigame_session surface
// sees (MA-C9). Components never call fetch; every shape is the generated DTO.
export interface MinigameSessionPort {
  current(): Promise<MinigameCurrentResponse>;
  create(minigameID: string, idempotencyKey: string): Promise<MinigameSessionResponseActive>;
  command(sessionID: string, request: MinigameCommandRequest): Promise<MinigameSessionResponse>;
  resolve(sessionID: string): Promise<MinigameSessionResponseTerminal>;
}

// MinigameAPIError is a typed, registry-shaped rejection (non-2xx with an exact
// {category, detail} body). Anything else is a MinigameTransportError.
export class MinigameAPIError extends Error {
  constructor(readonly status: number, readonly category: APIError["category"], readonly detail: APIError["detail"]) {
    super(`${status} ${category}/${detail}`);
  }
}

export class MinigameTransportError extends Error {}

const categories = new Set<string>(["conflict", "idempotency_conflict", "internal_invariant", "invalid", "not_configured", "not_eligible", "rate_limited", "unauthorized", "unknown_id"]);

function apiError(status: number, body: unknown): MinigameAPIError | undefined {
  if (body === null || typeof body !== "object" || Array.isArray(body)) return undefined;
  const row = body as Record<string, unknown>;
  const keys = Object.keys(row).sort();
  if (keys.length !== 2 || keys[0] !== "category" || keys[1] !== "detail") return undefined;
  if (typeof row.category !== "string" || !categories.has(row.category) || typeof row.detail !== "string" || row.detail === "") return undefined;
  return new MinigameAPIError(status, row.category as APIError["category"], row.detail as APIError["detail"]);
}

export type Fetcher = (input: string, init?: RequestInit) => Promise<Response>;
export type OperationCall = <T>(method: string, path: string, body: unknown) => Promise<T>;

// createOperationCall is the shared authenticated JSON call used by the
// generated-operation ports: exact registry errors become MinigameAPIError,
// everything else MinigameTransportError.
export function createOperationCall(accessToken: () => string, fetcher: Fetcher = fetch): OperationCall {
  return async <T>(method: string, path: string, body: unknown): Promise<T> => {
    let response: Response;
    try {
      const headers: Record<string, string> = { Authorization: `Bearer ${accessToken()}` };
      if (body !== undefined) headers["Content-Type"] = "application/json";
      response = await fetcher(path, { method, headers, body: body === undefined ? undefined : JSON.stringify(body) });
    } catch (error) {
      throw new MinigameTransportError(error instanceof Error ? error.message : "request failed");
    }
    let value: unknown;
    try { value = await response.json(); }
    catch { throw new MinigameTransportError(`response was not JSON (${response.status})`); }
    if (!response.ok) throw apiError(response.status, value) ?? new MinigameTransportError(`request failed (${response.status})`);
    return value as T;
  };
}

export function operationPath(template: string, parameters: Readonly<Record<string, string>>): string {
  return template.replace(/\{([a-z_]+)\}/gu, (_, name: string) => {
    const value = parameters[name];
    if (value === undefined) throw new TypeError(`missing path parameter ${name}`);
    return encodeURIComponent(value);
  });
}

export function createBrowserMinigameSessionPort(accessToken: () => string, fetcher: Fetcher = fetch): MinigameSessionPort {
  const call = createOperationCall(accessToken, fetcher);
  const path = operationPath;
  return {
    current: () => call(operations.get_current_minigame_session.method, operations.get_current_minigame_session.path, undefined),
    create: (minigameID, idempotencyKey) => call(operations.create_minigame_session.method,
      path(operations.create_minigame_session.path, { minigame_id: minigameID }), { idempotency_key: idempotencyKey }),
    command: (sessionID, request) => call(operations.play_minigame_command.method,
      path(operations.play_minigame_command.path, { session_id: sessionID }), request),
    resolve: (sessionID) => call(operations.resolve_minigame_session.method,
      path(operations.resolve_minigame_session.path, { session_id: sessionID }), {}),
  };
}
