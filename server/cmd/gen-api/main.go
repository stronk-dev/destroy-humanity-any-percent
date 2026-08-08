package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"cloud-clicker/server/account"
	"cloud-clicker/server/publicapi"
)

func main() {
	root := flag.String("root", "..", "repository root")
	updatePin := flag.Bool("update-pin", false, "replace the additive-v1 compatibility pin")
	flag.Parse()
	registry, err := account.PrivateAPIRegistry()
	if err != nil {
		fatal(err)
	}
	openapi, err := publicapi.GenerateOpenAPI(registry, "Cloud Clicker API v1")
	if err != nil {
		fatal(err)
	}
	types, err := publicapi.GenerateTypeScript(registry)
	if err != nil {
		fatal(err)
	}
	pins, err := publicapi.CanonicalOperationPins(registry)
	if err != nil {
		fatal(err)
	}
	pinPath := filepath.Join(*root, "docs/generated/api-compat-v1.json")
	if !*updatePin {
		prior, readErr := os.ReadFile(pinPath)
		if readErr != nil {
			fatal(fmt.Errorf("read compatibility pin (use -update-pin only for an intentional additive baseline): %w", readErr))
		}
		if err := publicapi.CheckCompatibilityPin(prior, registry); err != nil {
			fatal(err)
		}
	}
	outputs := map[string][]byte{
		filepath.Join(*root, "docs/generated/api.json"):           openapi,
		filepath.Join(*root, "client/src/api/generated/types.ts"): types,
	}
	if *updatePin {
		outputs[pinPath] = pins
	}
	for path, data := range outputs {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			fatal(err)
		}
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
