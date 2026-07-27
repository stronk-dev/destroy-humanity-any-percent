.PHONY: setup install-browsers install-browsers-ci test test-go test-client test-browser typecheck vectors vet fuzz verify

setup:
	pnpm --dir client install --frozen-lockfile
	$(MAKE) install-browsers

install-browsers:
	pnpm --dir client run setup:browsers

install-browsers-ci:
	pnpm --dir client run setup:browsers:ci

test: test-go test-client test-browser

test-go:
	cd server && go test ./...

test-client:
	pnpm --dir client run test

test-browser:
	pnpm --dir client run test:browser

typecheck:
	pnpm --dir client run typecheck

vectors:
	node tools/gen-vectors.mjs

vet:
	cd server && go vet ./...

fuzz:
	cd server && go test ./decimal -run '^$$' -fuzz '^FuzzCanonicalRoundTrip$$'

verify: vet typecheck test
