export interface AuthoritativeIntent { readonly intentId: string; readonly kind: string; readonly expectedRevision: number; readonly [field: string]: unknown }
export interface IntentRequestAdapter { request(intent: AuthoritativeIntent): Promise<void> }

export class IntentDispatcher {
  readonly #adapter: IntentRequestAdapter;
  constructor(adapter: IntentRequestAdapter) { this.#adapter = adapter; }
  send(intent: AuthoritativeIntent): Promise<void> { return this.#adapter.request(intent); }
}
