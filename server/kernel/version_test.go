package kernel

import (
	"os"
	"strings"
	"testing"
)

func TestVersionSourceOfTruth(t *testing.T) {
	data, err := os.ReadFile("../../kernel/VERSION")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != Version {
		t.Fatalf("generated version %q differs from kernel/VERSION %q", Version, strings.TrimSpace(string(data)))
	}
}
