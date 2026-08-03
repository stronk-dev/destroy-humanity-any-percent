// Command gen-content-manifest binds the current simulation and copy identities
// into the checked-in deployment content manifest.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cloud-clicker/server/epochseed"
)

var hashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type contentManifest struct {
	SchemaVersion int    `json:"schema_version"`
	ConstantsHash string `json:"constants_hash"`
	CopyHash      string `json:"copy_hash"`
}

func manifestBytes(constantsHash, copyHash string) ([]byte, error) {
	if !hashPattern.MatchString(constantsHash) || !hashPattern.MatchString(copyHash) {
		return nil, errors.New("content manifest requires canonical SHA-256 identities")
	}
	data, err := json.MarshalIndent(contentManifest{SchemaVersion: 1, ConstantsHash: constantsHash, CopyHash: copyHash}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func run(root, output string, check bool) error {
	if root == "" || output == "" || filepath.IsAbs(output) || filepath.Clean(output) != output {
		return errors.New("invalid content-manifest paths")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	bundle, err := epochseed.Load(absoluteRoot)
	if err != nil {
		return err
	}
	copyHashBytes, err := os.ReadFile(filepath.Join(absoluteRoot, "client", "src", "copy", "generated", "copy-hash.txt"))
	if err != nil {
		return err
	}
	expected, err := manifestBytes(bundle.Hash, strings.TrimSpace(string(copyHashBytes)))
	if err != nil {
		return err
	}
	filename := filepath.Join(absoluteRoot, filepath.FromSlash(output))
	if !strings.HasPrefix(filename, absoluteRoot+string(filepath.Separator)) {
		return errors.New("content manifest escapes repository root")
	}
	if check {
		actual, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if !bytes.Equal(actual, expected) {
			return errors.New("deployment content manifest drift (run make copy-generate)")
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filename, expected, 0o644)
}

func main() {
	root := flag.String("root", ".", "repository root")
	output := flag.String("output", "deployment/content-manifest.v1.json", "repository-relative output")
	check := flag.Bool("check", false, "check output instead of writing it")
	flag.Parse()
	if err := run(*root, *output, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *check {
		fmt.Println("deployment content manifest ok")
	}
}
