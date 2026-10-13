// Native RPC cannot cancel a key sequence already being delivered. Drain that
// sequence before teardown, and refuse every subsequent action after timeout.
// Shared by native-input journey fixtures; this is not a player-input policy.
const pendingNativeInput = new Set<Promise<unknown>>();

export function assertNativeInputDrained(): void {
  if (pendingNativeInput.size !== 0) throw new Error("previous test's native keyboard input has not drained");
}

export function lifetimeKeyboard(signal: AbortSignal, send: (keys: string) => PromiseLike<unknown>) {
  let closed = false;
  let flight: Promise<unknown> | undefined;
  return {
    async press(keys: string): Promise<void> {
      signal.throwIfAborted();
      if (closed) throw new Error("native keyboard fixture is closed");
      if (flight) throw new Error("overlapping native keyboard fixture actions");
      const work = Promise.resolve().then(() => {
        signal.throwIfAborted();
        if (closed) throw new Error("native keyboard fixture is closed");
        return send(keys);
      });
      flight = work;
      pendingNativeInput.add(work);
      try { await work; signal.throwIfAborted(); }
      finally { flight = undefined; pendingNativeInput.delete(work); }
    },
    async closeAndDrain(): Promise<void> { closed = true; await flight; },
  };
}
