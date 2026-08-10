import type { CopyKey } from "../copy";

export interface AmountCap {
  readonly amount: string;
  readonly reason_key: CopyKey;
}
