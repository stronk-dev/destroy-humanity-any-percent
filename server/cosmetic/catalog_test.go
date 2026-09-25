package cosmetic

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type catalogFixture struct {
	SchemaVersion int `json:"schema_version"`
	Cases         []struct {
		Name     string `json:"name"`
		Valid    bool   `json:"valid"`
		Artifact string `json:"artifact"`
	} `json:"cases"`
}

func loadCatalogFixtures(t *testing.T) catalogFixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "cosmetic", "catalog-fixtures-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture catalogFixture
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.SchemaVersion != 1 || len(fixture.Cases) == 0 {
		t.Fatalf("fixture: %v", err)
	}
	return fixture
}

// AC1: every shared case, including each payment-shaped key, is decided the
// same way the TypeScript loader decides it.
func TestCatalogFixtures(t *testing.T) {
	for _, testCase := range loadCatalogFixtures(t).Cases {
		_, err := Load([]byte(testCase.Artifact))
		if testCase.Valid && err != nil {
			t.Errorf("%s: unexpected rejection: %v", testCase.Name, err)
		}
		if !testCase.Valid && (err == nil || !errors.Is(err, ErrInvalidCatalog)) {
			t.Errorf("%s: accepted or wrong error: %v", testCase.Name, err)
		}
	}
}

func TestPinnedFixtureArtifactLoads(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "cosmetics", "fixture-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	item, ok := catalog.Item("horse_armor")
	if len(catalog.Items) != 1 || !ok || item.Slot != SlotPet || item.Unlock != (Unlock{Kind: UnlockActiveCompanyTierAtLeast, Tier: 1}) {
		t.Fatalf("catalog=%+v", catalog.Items)
	}
}

// AC2: the permanent-ID rule.
func TestValidateTransitionKeepsIDsAndSlots(t *testing.T) {
	current := mustLoad(t, `{"schema_version":1,"items":[{"cosmetic_id":"horse_armor","slot":"pet","unlock":{"kind":"active_company_tier_at_least","tier":1}}]}`)
	retuned := mustLoad(t, `{"schema_version":1,"items":[{"cosmetic_id":"horse_armor","slot":"pet","unlock":{"kind":"active_company_tier_at_least","tier":2}}]}`)
	dropped := mustLoad(t, `{"schema_version":1,"items":[{"cosmetic_id":"zebra_armor","slot":"pet","unlock":{"kind":"active_company_tier_at_least","tier":1}}]}`)
	if err := ValidateTransition(current, retuned); err != nil {
		t.Fatalf("retuned unlock rejected: %v", err)
	}
	if err := ValidateTransition(current, dropped); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("dropped id accepted: %v", err)
	}
	if err := ValidateTransition(current, nil); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("disappearing artifact accepted: %v", err)
	}
	// slot is a closed one-member enum in v1, so a changed slot cannot be
	// loaded; exercise the rule on a constructed successor.
	changed := &Catalog{Items: []Item{{CosmeticID: "horse_armor", Slot: "hat", Unlock: current.Items[0].Unlock}}, byID: map[string]int{"horse_armor": 0}}
	if err := ValidateTransition(current, changed); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("slot change accepted: %v", err)
	}
	if err := ValidateTransition(nil, current); err != nil {
		t.Fatalf("activation rejected: %v", err)
	}
}

func mustLoad(t *testing.T, data string) *Catalog {
	t.Helper()
	catalog, err := Load([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
