.PHONY: test test-go test-client vectors vet fuzz

test: test-go test-client

test-go:
	cd server && go test ./...

test-client:
	pnpm --dir client run test

vectors:
	node tools/gen-vectors.mjs

vet:
	cd server && go vet ./...

fuzz:
	cd server && go test ./decimal -run '^$$' -fuzz '^FuzzJavaScriptParity$$'
