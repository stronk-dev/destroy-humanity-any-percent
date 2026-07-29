package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceFingerprintTracksExecutableAuthoritiesOnly(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	sources := make(map[string][]byte)
	for _, authority := range formulaAuthorities {
		if _, exists := sources[authority.path]; exists {
			continue
		}
		source, err := os.ReadFile(filepath.Join(root, authority.path))
		if err != nil {
			t.Fatal(err)
		}
		sources[authority.path] = source
	}
	read := func(path string) ([]byte, error) { return sources[path], nil }
	baseline, err := sourceFingerprintFrom(read)
	if err != nil {
		t.Fatal(err)
	}
	if len(baseline) != 64 {
		t.Fatalf("fingerprint = %q", baseline)
	}

	originalEngine := append([]byte(nil), sources["production/engine.go"]...)
	sources["production/engine.go"] = append([]byte("// formula provenance comment\n"), originalEngine...)
	commentOnly, err := sourceFingerprintFrom(read)
	if err != nil || commentOnly != baseline {
		t.Fatalf("comment changed fingerprint: got %s want %s err=%v", commentOnly, baseline, err)
	}
	sources["production/engine.go"] = bytes.Replace(originalEngine, []byte("func Rates("), []byte("func  Rates ("), 1)
	formatOnly, err := sourceFingerprintFrom(read)
	if err != nil || formatOnly != baseline {
		t.Fatalf("formatting changed fingerprint: got %s want %s err=%v", formatOnly, baseline, err)
	}

	sources["production/engine.go"] = bytes.Replace(originalEngine,
		[]byte("rate = rate.Mul(bySource[sourceID].Factor)"),
		[]byte("rate = rate.Add(bySource[sourceID].Factor)"), 1)
	if bytes.Equal(sources["production/engine.go"], originalEngine) {
		t.Fatal("executable mutation fixture did not modify Rates")
	}
	changed, err := sourceFingerprintFrom(read)
	if err != nil {
		t.Fatal(err)
	}
	if changed == baseline {
		t.Fatal("executable rate change did not alter fingerprint")
	}

	sources["production/engine.go"] = []byte(strings.ReplaceAll(string(originalEngine), "func Rates(", "func RenamedRates("))
	if _, err := sourceFingerprintFrom(read); err == nil {
		t.Fatal("missing authority did not fail closed")
	}
}

func TestGeneratedAuthoritiesMatchPublishedRuntimeTokens(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sourceFingerprint(root); err != nil {
		t.Fatal(err)
	}
}
