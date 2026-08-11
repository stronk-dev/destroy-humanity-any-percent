.PHONY: setup install-browsers install-browsers-ci test test-go test-go-ci test-save-integration validate-migrations test-client test-browser test-browser-ci test-game-ui-composed test-game-ui-performance typecheck build-client build-gameserver vectors vectors-check replay-fixture replay-fixture-check pitch-corpus pitch-corpus-check formulas formulas-check api-generate api-schema api-pin api-check harness harness-check content-harness first-content-harness t0-t1-relevance commons-harness-check harness-update epoch-hash game-ui-copy-candidate game-ui-copy-candidate-check copy-generate copy-check vet fuzz fuzz-ci verify-schema verify-routes-boundary verify-commons-boundary verify-client-boundary verify-kernel-version verify-combat-boundary verify-meters-boundary verify-achievements-boundary verify-server verify-server-ci verify-client verify

# Keep ordinary Go builds inside the writable repository sandbox. Override either
# variable when a developer deliberately wants another cache or a focused package set.
REPO_CACHE_DIR ?= $(CURDIR)/.cache
export GOCACHE ?= $(REPO_CACHE_DIR)/go-build
GO_PACKAGES ?= ./...
GO_TEST_FLAGS ?=
SAVE_TEST_PACKAGES ?= ./...
SAVE_TEST_FLAGS ?= -run Integration
SAVE_TEST_COUNT ?= 1
CI_TEST_PACKAGES ?= ./...
CI_TEST_FLAGS ?=
CI_TEST_COUNT ?= 1
CLIENT_BIN := $(CURDIR)/client/node_modules/.bin
BROWSER_TEST_FLAGS ?=

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

# Reproduce the Actions server job on its real architecture, with Postgres and
# cold test execution. This intentionally runs every package rather than the
# Integration-named subset used by the focused save-integration target.
test-go-ci:
	docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm test go test -p 1 $(CI_TEST_FLAGS) $(CI_TEST_PACKAGES) -count=$(CI_TEST_COUNT)

test-save-integration:
	docker compose -f compose.save-test.yml run --rm test go test -p 1 $(SAVE_TEST_FLAGS) $(SAVE_TEST_PACKAGES) -count=$(SAVE_TEST_COUNT)

# Validate the complete embedded migration chain on real Postgres while keeping
# the scope focused on the package that owns it. Migration-named unit probes and
# every save integration test both run cold.
validate-migrations:
	docker compose -f compose.save-test.yml run --rm test go test -p 1 -run 'Integration|Migration' ./save -count=1

test-client:
	cd client && $(CLIENT_BIN)/vitest run

test-browser:
	cd client && VITE_GAME_UI_PERFORMANCE=0 $(CLIENT_BIN)/vitest run --config vitest.browser.config.ts $(BROWSER_TEST_FLAGS)
	$(MAKE) test-game-ui-performance

test-browser-ci:
	docker compose -f compose.browser-test.yml run --rm browser

# Real browser -> Vite proxy -> composed gameserver/Postgres. The visible
# visitor counter is the assertion that runtime.ts completed its actual
# Centrifuge WebSocket handshake, not a mocked socket exchange.
test-game-ui-composed:
	docker compose -f compose.game-ui-test.yml up -d --wait game-ui-postgres
	node client/tools/test-game-ui-composed.mjs

test-game-ui-performance:
	cd client && VITE_GAME_UI_PERFORMANCE=1 $(CLIENT_BIN)/vitest run --config vitest.browser.config.ts test/game-ui-screens-browser.test.ts --testNamePattern 'observable 20 Hz / 10 Hz screen budget'

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

pitch-corpus:
	cd server && go test ./pitch -run '^TestPitchContentGate$$' -update-pitch-corpus

pitch-corpus-check:
	cd server && go test ./pitch -run '^TestPitchContentGate$$'

formulas:
	cd server && go run ./cmd/gen-formulas -output ../docs/generated/production-formulas.json

formulas-check: formulas
	git diff --exit-code -- docs/generated/production-formulas.json

api-generate:
	cd server && go run ./cmd/gen-api -root=..

api-schema: api-generate

api-pin:
	cd server && go run ./cmd/gen-api -root=.. -update-pin

api-check: api-generate
	git diff --exit-code -- docs/generated/api.json docs/generated/api-compat-v1.json client/src/api/generated/types.ts

harness:
	@test -n "$(HARNESS_OUTPUT)" || (echo "HARNESS_OUTPUT is required" >&2; exit 1)
	cd server && go run ./cmd/balance-harness -mode=run -root=.. -output="$(HARNESS_OUTPUT)"

harness-check: commons-harness-check
	cd server && go run ./cmd/balance-harness -mode=check -root=..

content-harness:
	cd server && go run ./cmd/balance-harness -mode=content -root=..

first-content-harness:
	cd server && go run ./cmd/balance-harness -mode=candidate -root=.. \
		-candidate-manifest=planning/first-content-epoch/promotion-manifest.candidate.v1.json \
		-output=../planning/first-content-epoch/composed-harness-report.v1.json

t0-t1-relevance:
	cd server && go run ./cmd/balance-harness -mode=relevance -root=.. \
		-scenario=balance/testdata/t0-t1/relevance-scenario-v2.json \
		-output=../planning/t0-t1-content/relevance-report.v2.json

commons-harness-check:
	cd server && go test ./harness -run '^TestCommonsPopulationInvariance$$' -count=1

harness-update:
	cd server && go run ./cmd/balance-harness -mode=update -root=..

epoch-hash:
	cd server && go run ./cmd/balance-harness -mode=epoch-hash -root=..

game-ui-copy-candidate:
	node client/tools/assemble-game-ui-copy.mjs

game-ui-copy-candidate-check:
	node client/tools/assemble-game-ui-copy.mjs --check

copy-generate:
	node client/tools/generate-copy.mjs
	cd server && go run ./cmd/gen-content-manifest -root=.. -output=deployment/content-manifest.v1.json

copy-check: game-ui-copy-candidate-check
	node client/tools/verify-copy.mjs
	cd server && go run ./cmd/gen-content-manifest -root=.. -output=deployment/content-manifest.v1.json -check

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

verify-meters-boundary:
	node client/tools/verify-meters-boundaries.mjs

verify-achievements-boundary:
	node client/tools/verify-achievements-boundaries.mjs

verify-server: vet test-go pitch-corpus-check formulas-check api-check harness-check verify-routes-boundary verify-commons-boundary

verify-server-ci:
	docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm test sh -c 'cd /workspace && make verify-server'

verify-client: typecheck build-client test-client verify-client-boundary verify-kernel-version verify-combat-boundary verify-meters-boundary verify-achievements-boundary copy-check

verify: verify-server verify-client verify-schema test-browser
