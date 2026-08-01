package commons

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
)

type vectors struct {
	Enclosure []struct {
		Name    string `json:"name"`
		Sources []struct {
			SourceID string `json:"source_id"`
			Factor   string `json:"factor"`
		} `json:"sources"`
		Expected string `json:"expected"`
	} `json:"enclosure"`
	Modifiers []struct {
		HealthPPM     int64  `json:"health_ppm"`
		SolidarityPPM int64  `json:"solidarity_ppm"`
		Expected      string `json:"expected"`
	} `json:"modifiers"`
}

func testCatalog(t *testing.T) *Catalog {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "balance", "commons", "phase0.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["source_weights"] = []map[string]any{
		{"source_id": "source.ethical", "slot": "doctrine", "weight_ppm": 1000000, "forsworn": false},
		{"source_id": "source.dark", "slot": "upgrades", "weight_ppm": 1000000, "forsworn": true},
	}
	data, _ = json.Marshal(raw)
	catalog, err := LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestFormulaVectors(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "commons", "formula-vectors.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures vectors
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	catalog := testCatalog(t)
	for _, fixture := range fixtures.Enclosure {
		t.Run("enclosure/"+fixture.Name, func(t *testing.T) {
			items := make([]multiplier.Contribution, 0, len(fixture.Sources))
			for _, source := range fixture.Sources {
				factor, err := decimal.ParseCanonical(source.Factor)
				if err != nil {
					t.Fatal(err)
				}
				weight, _ := catalog.SourceWeight(source.SourceID)
				items = append(items, multiplier.Contribution{SourceID: source.SourceID, Slot: weight.Slot, Target: "all", Factor: factor})
			}
			actual, err := EnclosureIndex(catalog, items)
			if err != nil {
				t.Fatal(err)
			}
			if actual.String() != fixture.Expected {
				t.Fatalf("got %s want %s", actual.String(), fixture.Expected)
			}
		})
	}
	for _, fixture := range fixtures.Modifiers {
		actual, err := Modifier(catalog, fixture.HealthPPM, fixture.SolidarityPPM)
		if err != nil {
			t.Fatal(err)
		}
		if actual.String() != fixture.Expected {
			t.Fatalf("H=%d s=%d got %s want %s", fixture.HealthPPM, fixture.SolidarityPPM, actual.String(), fixture.Expected)
		}
	}
}

func TestNonMemberContributionIsAbsent(t *testing.T) {
	items, err := Contribution(testCatalog(t), false, 1_000_000, 1_000_000)
	if err != nil || len(items) != 0 {
		t.Fatalf("items=%v err=%v", items, err)
	}
}

func TestCollectiveExponentComesFromCatalog(t *testing.T) {
	catalog := testCatalog(t)
	catalog.CollectiveExponentPPM = 2_000_000
	actual, err := Modifier(catalog, 675_000, 0)
	if err != nil || actual.String() != "1.75e0" {
		t.Fatalf("modifier=%s err=%v", actual.String(), err)
	}
}

func TestEntryParticipationWeightNormalizesAndClamps(t *testing.T) {
	catalog := testCatalog(t)
	for tithe, want := range map[int64]int64{50_000: 500_000, 100_000: 1_000_000, 150_000: 1_000_000} {
		got, err := EntryParticipationWeightPPM(catalog, tithe)
		if err != nil || got != want {
			t.Fatalf("tithe=%d weight=%d want=%d err=%v", tithe, got, want, err)
		}
	}
	if _, err := EntryParticipationWeightPPM(catalog, 49_999); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("below-band error=%v", err)
	}
}
