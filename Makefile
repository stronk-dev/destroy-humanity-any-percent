.PHONY: setup install-browsers install-browsers-ci test test-go test-save-integration test-client test-browser typecheck vectors vectors-check formulas formulas-check harness harness-check harness-update vet fuzz fuzz-ci verify-schema verify-server verify-client verify

setup:
	pnpm --dir client install --frozen-lockfile
	$(MAKE) install-browsers

install-browsers:
	pnpm --dir client run setup:browsers

install-browsers-ci:
	pnpm --dir client run setup:browsers:ci

test: test-go test-client test-browser

test-go:
	cd server && go test -p 1 ./...

test-save-integration:
	@test -n "$$TEST_DATABASE_URL" || (echo "TEST_DATABASE_URL is required" >&2; exit 1)
	cd server && go test -p 1 ./... -run Integration -count=1

test-client:
	pnpm --dir client run test

test-browser:
	pnpm --dir client run test:browser

typecheck:
	pnpm --dir client run typecheck

vectors:
	node tools/gen-vectors.mjs

vectors-check: vectors
	git diff --exit-code -- testdata/decimal-vectors.json

formulas:
	cd server && go run ./cmd/gen-formulas -output ../docs/generated/production-formulas.json

formulas-check: formulas
	git diff --exit-code -- docs/generated/production-formulas.json

harness:
	@test -n "$(HARNESS_OUTPUT)" || (echo "HARNESS_OUTPUT is required" >&2; exit 1)
	cd server && go run ./cmd/balance-harness -mode=run -root=.. -output="$(HARNESS_OUTPUT)"

harness-check:
	cd server && go run ./cmd/balance-harness -mode=check -root=..

harness-update:
	cd server && go run ./cmd/balance-harness -mode=update -root=..

vet:
	cd server && go vet ./...

fuzz:
	cd server && go test ./decimal -run '^$$' -fuzz '^FuzzCanonicalRoundTrip$$'

fuzz-ci:
	cd server && go test ./decimal -run '^$$' -fuzz '^FuzzCanonicalRoundTrip$$' -fuzztime=30s

verify-schema:
	pnpm --dir client run verify:schema

verify-server: vet test-go formulas-check harness-check

verify-client: typecheck test-client

verify: verify-server verify-client verify-schema test-browser
