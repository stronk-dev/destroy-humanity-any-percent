// Cosmetic Shop v1 §10 N5: a browser network witness. While installed, every
// fetch/WebSocket must target the same origin under an allowlisted path, and
// PaymentRequest, navigator.credentials.* and window.open are trapped. Any
// violation is recorded and fails the test that asserts `violations` empty.
const ALLOWED_PATHS = [/^\/api\//u, /^\/connection\/websocket$/u, /^\/@fs\//u, /^\/@vite\//u, /^\/node_modules\//u, /^\/src\//u, /^\/test\//u];

export interface NetworkTrap { readonly violations: string[]; restore(): void }

export function sameOriginAllowed(url: string, origin: string): boolean {
  let parsed: URL;
  try { parsed = new URL(url, origin); } catch { return false; }
  const expectedProtocols = [new URL(origin).protocol, new URL(origin).protocol === "https:" ? "wss:" : "ws:"];
  return parsed.host === new URL(origin).host && expectedProtocols.includes(parsed.protocol) && ALLOWED_PATHS.some((pattern) => pattern.test(parsed.pathname));
}

export function installNetworkTrap(): NetworkTrap {
  const violations: string[] = [];
  const origin = window.location.origin;
  const originalFetch = window.fetch, originalSocket = window.WebSocket, originalOpen = window.open;
  const originalPayment = (window as unknown as { PaymentRequest?: unknown }).PaymentRequest;
  const credentials = navigator.credentials as unknown as Record<string, unknown> | undefined;
  const originalCredentials = credentials ? { get: credentials.get, create: credentials.create, store: credentials.store } : undefined;
  window.fetch = (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === "string" ? input : input instanceof URL ? input.href : input.url;
    if (!sameOriginAllowed(url, origin)) { violations.push(`fetch ${url}`); return Promise.reject(new TypeError("network trap")); }
    return originalFetch.call(window, input, init);
  };
  window.WebSocket = new Proxy(originalSocket, { construct(target, args: [string | URL, (string | string[])?]) {
    const url = String(args[0]);
    if (!sameOriginAllowed(url, origin)) { violations.push(`websocket ${url}`); throw new TypeError("network trap"); }
    return new target(...args);
  } });
  window.open = ((url?: string | URL) => { violations.push(`window.open ${String(url)}`); return null; }) as typeof window.open;
  (window as unknown as { PaymentRequest: unknown }).PaymentRequest = function PaymentRequestTrap() { violations.push("PaymentRequest"); throw new TypeError("network trap"); };
  if (credentials) for (const name of ["get", "create", "store"]) credentials[name] = () => { violations.push(`navigator.credentials.${name}`); return Promise.reject(new TypeError("network trap")); };
  return { violations, restore() {
    window.fetch = originalFetch; window.WebSocket = originalSocket; window.open = originalOpen;
    (window as unknown as { PaymentRequest?: unknown }).PaymentRequest = originalPayment;
    if (credentials && originalCredentials) Object.assign(credentials, originalCredentials);
  } };
}
