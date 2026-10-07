import { apiErrorCategories, createAPIClient, operations, type APIFetcher, type APIError, type OperationID, type OperationInput, type OperationTypes, type MinigameCommandRequest, type MinigameCurrentResponse, type MinigameSessionResponse, type MinigameSessionResponseActive, type MinigameSessionResponseTerminal } from "../../api/generated/types";

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

const categories = new Set<string>(apiErrorCategories);

function apiError(status: number, body: unknown): MinigameAPIError | undefined {
  if (body === null || typeof body !== "object" || Array.isArray(body)) return undefined;
  const row = body as Record<string, unknown>;
  const keys = Object.keys(row).sort();
  if (keys.length !== 2 || keys[0] !== "category" || keys[1] !== "detail") return undefined;
  if (typeof row.category !== "string" || !categories.has(row.category) || typeof row.detail !== "string" || row.detail === "") return undefined;
  return new MinigameAPIError(status, row.category as APIError["category"], row.detail as APIError["detail"]);
}

export type Fetcher = APIFetcher;
type AuthenticatedOperationID = { [K in OperationID]: typeof operations[K]["auth"] extends "access_token" ? K : never }[OperationID];
type PortInput<K extends AuthenticatedOperationID> = Omit<OperationInput<K>, "accessToken">;
type PortSuccess<K extends AuthenticatedOperationID> = Exclude<OperationTypes[K]["response"], APIError>;
export type OperationCall = <K extends AuthenticatedOperationID>(id: K, input: PortInput<K>) => Promise<PortSuccess<K>>;

// createOperationCall is the shared authenticated JSON call used by the
// generated-operation ports: exact registry errors become MinigameAPIError,
// everything else MinigameTransportError.
export function createOperationCall(accessToken: () => string, fetcher: Fetcher = fetch): OperationCall {
  const api = createAPIClient(fetcher);
  return async <K extends AuthenticatedOperationID>(id: K, input: PortInput<K>): Promise<PortSuccess<K>> => {
    let response: { status: number; ok: boolean; body: unknown };
    try {
      // Erase only the generic body union here; generated call inputs and each
      // port's return DTO remain typed. The SDK has checked the status already.
      response = await api.call(id, { ...input, accessToken: accessToken() } as OperationInput<K>) as unknown as typeof response;
    } catch (error) {
      throw new MinigameTransportError(error instanceof Error ? error.message : "request failed");
    }
    if (!response.ok) throw apiError(response.status, response.body) ?? new MinigameTransportError(`request failed (${response.status})`);
    return response.body as PortSuccess<K>;
  };
}

export function createBrowserMinigameSessionPort(accessToken: () => string, fetcher: Fetcher = fetch): MinigameSessionPort {
  const call = createOperationCall(accessToken, fetcher);
  return {
    current: () => call("get_current_minigame_session", { path: {}, request: null }),
    create: (minigameID, idempotencyKey) => call("create_minigame_session", { path: { minigame_id: minigameID }, request: { idempotency_key: idempotencyKey } }),
    command: (sessionID, request) => call("play_minigame_command", { path: { session_id: sessionID }, request }),
    resolve: (sessionID) => call("resolve_minigame_session", { path: { session_id: sessionID }, request: {} }),
  };
}
