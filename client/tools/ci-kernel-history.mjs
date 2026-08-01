export function verifyCIKernelHistory(source) {
  const lines = source.split(/\r?\n/);
  const clientStart = lines.findIndex((line) => /^  client:\s*$/.test(line));
  if (clientStart < 0) throw new Error("CI has no client job");
  let clientEnd = lines.length;
  for (let index = clientStart + 1; index < lines.length; index++) {
    if (/^  [A-Za-z0-9_-]+:\s*$/.test(lines[index])) {
      clientEnd = index;
      break;
    }
  }
  const client = lines.slice(clientStart, clientEnd);
  if (!client.some((line) => /^      - run: make verify-client\s*$/.test(line))) {
    throw new Error("CI client job does not run make verify-client");
  }
  const checkout = client.findIndex((line) => /^      - uses: actions\/checkout@/.test(line));
  if (checkout < 0) throw new Error("CI client job has no checkout step");
  const checkoutStep = client.slice(checkout, client.findIndex((line, offset) => offset > checkout && /^      - /.test(line)));
  if (!checkoutStep.some((line) => /^          fetch-depth: 0\s*$/.test(line))) {
    throw new Error("CI client job must fetch complete history for verify-kernel-version");
  }
}
