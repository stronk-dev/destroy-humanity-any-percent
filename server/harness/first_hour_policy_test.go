package harness

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirstHourPolicyRegistryIsClosedAndIdentityBound(t *testing.T) {
	root := filepath.Join("..", "..")
	registry, data, err := LoadFirstHourPolicyRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"casual.t0_t1", "chaos.t0_t1", "reference.greedy"} {
		if _, ok := registry.Policy(id, 1); !ok {
			t.Fatalf("missing %s", id)
		}
	}
	if registry.Hash == "" {
		t.Fatal("policy bytes omitted from identity")
	}
	tampered := bytes.Replace(data, []byte(`"top_k": 3`), []byte(`"top_k": 1`), 1)
	if _, err := DecodeFirstHourPolicyRegistry(tampered); err == nil || !strings.Contains(err.Error(), "top_k") {
		t.Fatalf("seeded top-k tamper err=%v", err)
	}
}

func TestFirstHourPolicyRejectsOddSplitBonusAndUnknownFields(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(FirstHourPolicyPath)))
	if err != nil {
		t.Fatal(err)
	}
	odd := bytes.Replace(data, []byte(`"burnout_route_knowledge_bonus": 24`), []byte(`"burnout_route_knowledge_bonus": 25`), 1)
	if _, err := DecodeFirstHourPolicyRegistry(odd); err == nil {
		t.Fatal("odd split bonus accepted")
	}
	unknown := bytes.Replace(data, []byte(`"schema_version": 1`), []byte(`"schema_version": 1, "ambient": true`), 1)
	if _, err := DecodeFirstHourPolicyRegistry(unknown); err == nil {
		t.Fatal("unknown policy field accepted")
	}
}
