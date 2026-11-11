#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 0 ]; then
  printf '%s\n' 'expected no arguments: this lane runs the complete composed target' >&2
  exit 2
fi
go version
node --version
docker compose version
corepack enable
corepack prepare "$(node -p 'require("./client/package.json").packageManager')" --activate
pnpm --dir client install --frozen-lockfile
exec make test-game-ui-composed
