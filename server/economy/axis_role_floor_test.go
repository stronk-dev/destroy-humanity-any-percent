package economy_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"cloud-clicker/server/economy"
)

// CV1 item5 requires a role on axis upgrades, without changing the legacy
// static-upgrade grammar. Role identity is not proof of executed pool binding.
func TestAxisUpgradeRoleFloor(t *testing.T) {
	var corpus struct {
		Version    int    `json:"schema_version"`
		FixtureSHA string `json:"fixture_sha256"`
		Cases      []struct {
			Name   string   `json:"name"`
			IDs    []string `json:"ids"`
			Roles  []string `json:"roles"`
			Expect string   `json:"expect"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(repositoryFile(t, "testdata/axis-stack/role-floor-v1.json"), &corpus); err != nil {
		t.Fatal(err)
	}
	fixture := repositoryFile(t, "balance/testdata/axis-stack/economy-v5-fixture.json")
	digest := sha256.Sum256(fixture)
	if corpus.Version != 1 || len(corpus.Cases) != 6 || hex.EncodeToString(digest[:]) != corpus.FixtureSHA {
		t.Fatal("role-floor population or pinned fixture changed")
	}
	accepts, rejects := 0, 0
	for _, row := range corpus.Cases {
		switch row.Expect {
		case "accept":
			accepts++
		case "reject":
			rejects++
		default:
			t.Fatalf("unknown expectation %q", row.Expect)
		}
		t.Run(row.Name, func(t *testing.T) {
			var root map[string]json.RawMessage
			if err := json.Unmarshal(fixture, &root); err != nil {
				t.Fatal(err)
			}
			var upgrades []map[string]json.RawMessage
			if err := json.Unmarshal(root["upgrades"], &upgrades); err != nil {
				t.Fatal(err)
			}
			for _, target := range row.IDs {
				matched := 0
				for _, upgrade := range upgrades {
					var id string
					if err := json.Unmarshal(upgrade["id"], &id); err != nil {
						t.Fatal(err)
					}
					if id != target {
						continue
					}
					matched++
					encoded, err := json.Marshal(row.Roles)
					if err != nil {
						t.Fatal(err)
					}
					upgrade["roles"] = encoded
				}
				if matched != 1 {
					t.Fatalf("target %s matched %d rows", target, matched)
				}
			}
			encoded, err := json.Marshal(upgrades)
			if err != nil {
				t.Fatal(err)
			}
			root["upgrades"] = encoded
			data, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			_, err = economy.LoadCatalog(data)
			if row.Expect == "accept" {
				if err != nil {
					t.Fatalf("legal role control rejected: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "axis upgrade requires at least one role") {
				t.Fatalf("roleless axis upgrade must refuse at its role floor, got %v", err)
			}
		})
	}
	if accepts != 3 || rejects != 3 {
		t.Fatalf("population: %d accepts / %d rejects", accepts, rejects)
	}
}
