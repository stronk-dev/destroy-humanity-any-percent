#!/usr/bin/env bash
set -euo pipefail

# Compose supplies the full lane by default. A focused invocation changes only
# the final Make target/selector, never image/setup, clean volumes or containment.
case "${1:-}" in
  test-browser|test-browser-focused) ;;
  *) printf '%s\n' 'expected test-browser or test-browser-focused' >&2; exit 2 ;;
esac

apt-get update
apt-get install -y --no-install-recommends make
corepack enable
corepack prepare pnpm@11.15.1 --activate
pnpm --dir client install --frozen-lockfile
exec make "$@"
