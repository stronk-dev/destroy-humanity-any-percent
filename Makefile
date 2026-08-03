.PHONY: setup install-browsers install-browsers-ci test test-go test-save-integration test-client test-browser typecheck build-client build-gameserver vectors vectors-check replay-fixture replay-fixture-check formulas formulas-check harness harness-check commons-harness-check harness-update epoch-hash copy-generate copy-check vet fuzz fuzz-ci verify-schema verify-routes-boundary verify-commons-boundary verify-client-boundary verify-kernel-version verify-combat-boundary verify-server verify-client verify

# Keep ordinary Go builds inside the writable repository sandbox. Override either
# variable when a developer deliberately wants another cache or a focused package set.
REPO_CACHE_DIR ?= $(CURDIR)/.cache
export GOCACHE ?= $(REPO_CACHE_DIR)/go-build
GO_PACKAGES ?= ./...
GO_TEST_FLAGS ?=
CLIENT_BIN := $(CURDIR)/client/node_modules/.bin

setup:
	pnpm --dir client install --frozen-lockfile
	$(MAKE) install-browsers

install-browsers:
	pnpm --dir client run setup:browsers

install-browsers-ci:
	pnpm --dir client run setup:browsers:ci

test: test-go test-client test-browser

test-go:
	cd server && go test -p 1 $(GO_TEST_FLAGS) $(GO_PACKAGES)

test-save-integration:
	docker compose -f compose.save-test.yml run --rm test

test-client:
	cd client && $(CLIENT_BIN)/vitest run

test-browser:
	cd client && $(CLIENT_BIN)/vitest run --config vitest.browser.config.ts

typecheck:
	cd client && $(CLIENT_BIN)/tsc --noEmit && $(CLIENT_BIN)/svelte-check --tsconfig ./tsconfig.json

build-client:
	cd client && $(CLIENT_BIN)/vite build

build-gameserver:
	mkdir -p $(REPO_CACHE_DIR)/bin
	cd server && go build -o $(REPO_CACHE_DIR)/bin/gameserver ./cmd/gameserver

vectors:
	node tools/gen-vectors.mjs

vectors-check: vectors
	git diff --exit-code -- testdata/decimal-vectors.json

replay-fixture:
	cd server && go test ./production -run '^TestApplyLoggedCrossRuntimeFixture$$' -update-replay-fixture

replay-fixture-check:
	cd server && go test ./production -run '^TestApplyLoggedCrossRuntimeFixture$$'

formulas:
	cd server && go run ./cmd/gen-formulas -output ../docs/generated/production-formulas.json

formulas-check: formulas
	git diff --exit-code -- docs/generated/production-formulas.json

harness:
	@test -n "$(HARNESS_OUTPUT)" || (echo "HARNESS_OUTPUT is required" >&2; exit 1)
	cd server && go run ./cmd/balance-harness -mode=run -root=.. -output="$(HARNESS_OUTPUT)"

harness-check: commons-harness-check
	cd server && go run ./cmd/balance-harness -mode=check -root=..

commons-harness-check:
	cd server && go test ./harness -run '^TestCommonsPopulationInvariance$$' -count=1

harness-update:
	cd server && go run ./cmd/balance-harness -mode=update -root=..

epoch-hash:
	cd server && go run ./cmd/balance-harness -mode=epoch-hash -root=..

copy-generate:
	node client/tools/generate-copy.mjs

copy-check:
	node client/tools/verify-copy.mjs

vet:
	cd server && go vet $(GO_PACKAGES)

fuzz:
	cd server && go test ./decimal -run '^$$' -fuzz '^FuzzCanonicalRoundTrip$$'

fuzz-ci:
	cd server && go test ./decimal -run '^$$' -fuzz '^FuzzCanonicalRoundTrip$$' -fuzztime=30s

verify-schema:
	node client/tools/verify-schema.mjs

verify-routes-boundary:
	@imports=$$(cd server && GOCACHE=/tmp/cloud-clicker-routes-go-cache go list -f '{{range .Imports}}{{println .}}{{end}}' ./routes) || { echo 'routes import enumeration failed' >&2; exit 1; }; unexpected=$$(printf '%s\n' "$$imports" | grep '^cloud-clicker/server/' | grep -vx 'cloud-clicker/server/decimal' || true); if [ -n "$$unexpected" ]; then echo "routes package has disallowed internal imports:" >&2; echo "$$unexpected" >&2; exit 1; fi

verify-commons-boundary:
	@if cd server && GOCACHE=/tmp/cloud-clicker-commons-go-cache go list -deps ./commons | grep -qx 'cloud-clicker/server/production'; then echo 'commons package must not import production' >&2; exit 1; fi
	@if cd server && GOCACHE=/tmp/cloud-clicker-commons-go-cache go list -deps ./production | grep -qx 'cloud-clicker/server/commons'; then echo 'production package must not import commons' >&2; exit 1; fi

verify-client-boundary:
	node client/tools/verify-shell-boundaries.mjs

verify-kernel-version:
	node client/tools/verify-ci-kernel-history.mjs
	node client/tools/verify-ci-kernel-history-fixtures.mjs
	node client/tools/verify-kernel-version.mjs
	node client/tools/verify-kernel-version-fixtures.mjs

verify-combat-boundary:
	node client/tools/verify-combat-boundaries.mjs

verify-server: vet test-go formulas-check harness-check verify-routes-boundary verify-commons-boundary

verify-client: typecheck build-client test-client verify-client-boundary verify-kernel-version verify-combat-boundary copy-check

verify: verify-server verify-client verify-schema test-browser
