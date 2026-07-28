package save

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
)

const stateCatalogJSON = `{
  "schema_version": 2,
  "resources": [{
    "id": "company.cash", "scope": "company", "numeric_kind": "decimal",
    "initial": "0", "minimum": "0",
    "hardcap": {"amount": "1e100", "reason_key": "resource.company_cash.cap.test"}
  }],
  "generator_classes": [{
    "id": "generator.example",
    "price": {"resource_id":"company.cash","base":"1e0","curve":{"kind":"constant"}},
    "production": {"resource_id":"company.cash","base_rate":"1e0"}
  }]
}`

var testCursor = time.Date(2026, 7, 28, 8, 0, 0, 123456789, time.UTC)

type migrationFixture struct {
	Version int             `json:"version"`
	Cases   []migrationCase `json:"cases"`
}

type migrationCase struct {
	Name              string          `json:"name"`
	FromVersion       int             `json:"from_version"`
	Scope             economy.Scope   `json:"scope"`
	MigrationBaseline string          `json:"migration_baseline"`
	Input             json.RawMessage `json:"input"`
	ExpectV2          json.RawMessage `json:"expect_v2"`
}

func stateCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	catalog, err := economy.LoadCatalog([]byte(stateCatalogJSON))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func testState(t *testing.T) *State {
	t.Helper()
	ledger, err := economy.RestoreLedger(stateCatalog(t), economy.ScopeCompany, map[string]string{
		"company.cash": "1.23456789012e42",
	})
	if err != nil {
		t.Fatal(err)
	}
	return &State{
		Ledger: ledger, GeneratorCounts: map[string]int64{"generator.example": 42}, EvaluatedThrough: testCursor,
	}
}

func TestStateV2RoundTrip(t *testing.T) {
	encoded, err := EncodeState(testState(t))
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, CurrentVersion, stateCatalog(t), economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if got := restored.Ledger.Snapshot()["company.cash"]; got != "1.23456789012e42" {
		t.Fatalf("balance = %s", got)
	}
	if restored.GeneratorCounts["generator.example"] != 42 || !restored.EvaluatedThrough.Equal(testCursor) {
		t.Fatalf("restored state = %+v", restored)
	}
}

func TestSaveMigrationCorpus(t *testing.T) {
	data, err := os.ReadFile("../../testdata/save-migrations.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture migrationFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Cases) == 0 {
		t.Fatalf("migration fixture version=%d cases=%d", fixture.Version, len(fixture.Cases))
	}
	for _, vector := range fixture.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			baseline, err := time.Parse(time.RFC3339Nano, vector.MigrationBaseline)
			if err != nil {
				t.Fatal(err)
			}
			restored, err := RestoreState(vector.Input, vector.FromVersion, stateCatalog(t), vector.Scope, baseline)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := EncodeState(restored)
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(vector.ExpectV2, &want); err != nil {
				t.Fatal(err)
			}
			if !equalJSON(got, want) {
				t.Fatalf("migrated JSON = %s, want %s", encoded, vector.ExpectV2)
			}
			if _, err := RestoreState(encoded, CurrentVersion, stateCatalog(t), vector.Scope, time.Time{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func equalJSON(left, right any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return string(leftJSON) == string(rightJSON)
}

func TestRestoreRejectsPoisonedAndMalformedState(t *testing.T) {
	validCursor := `"2026-07-28T08:00:00Z"`
	tests := []string{
		`{"balances":{"company.cash":"NaN"},"generators":{"generator.example":0},"evaluated_through":` + validCursor + `}`,
		`{"balances":{"company.cash":1},"generators":{"generator.example":0},"evaluated_through":` + validCursor + `}`,
		`{"balances":{},"generators":{"generator.example":0},"evaluated_through":` + validCursor + `}`,
		`{"balances":{"company.cash":"0"},"generators":{},"evaluated_through":` + validCursor + `}`,
		`{"balances":{"company.cash":"0"},"generators":{"generator.example":-1},"evaluated_through":` + validCursor + `}`,
		`{"balances":{"company.cash":"0"},"generators":{"generator.example":9007199254740992},"evaluated_through":` + validCursor + `}`,
		`{"balances":{"company.cash":"0"},"generators":{"generator.example":0,"generator.extra":0},"evaluated_through":` + validCursor + `}`,
		`{"balances":{"company.cash":"0"},"generators":{"generator.example":0},"evaluated_through":"2026-07-28T10:00:00+02:00"}`,
		`{"balances":{"company.cash":"0"},"generators":{"generator.example":0},"evaluated_through":"2026-07-28T08:00:00.000Z"}`,
		`{"balances":{"company.cash":"0"},"generators":{"generator.example":0},"evaluated_through":` + validCursor + `,"unknown":true}`,
	}
	for _, data := range tests {
		if _, err := RestoreState([]byte(data), CurrentVersion, stateCatalog(t), economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
			t.Fatalf("RestoreState(%s) error = %v", data, err)
		}
	}
}

func TestRestoreRejectsFutureVersion(t *testing.T) {
	_, err := RestoreState([]byte(`{}`), CurrentVersion+1, stateCatalog(t), economy.ScopeCompany, time.Time{})
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("error = %v", err)
	}
}

func TestEncodeRejectsUnsafeGeneratorCountAndMissingCursor(t *testing.T) {
	state := testState(t)
	state.GeneratorCounts["generator.example"] = decimal.MaxExactInteger + 1
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("unsafe count error = %v", err)
	}
	state.GeneratorCounts["generator.example"] = 0
	state.EvaluatedThrough = time.Time{}
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("missing cursor error = %v", err)
	}
}

func TestConstantsHashUsesExactArtifactBytes(t *testing.T) {
	first := ConstantsHash([]byte("{}"))
	second := ConstantsHash([]byte("{}\n"))
	if first == second || len(first) != len("sha256:")+64 {
		t.Fatalf("unexpected hashes %q %q", first, second)
	}
}

func TestStreamOwnerScopeValidation(t *testing.T) {
	state := testState(t)
	validHash := ConstantsHash([]byte(stateCatalogJSON))
	context := WriteContext{Cause: "test"}
	valid := StreamKey{OwnerKind: OwnerFounder, OwnerID: "11111111-1111-4111-8111-111111111111", Scope: economy.ScopeCompany}
	if err := validateWrite(valid, validHash, state, context); err != nil {
		t.Fatal(err)
	}
	invalid := []StreamKey{
		{OwnerKind: OwnerGuild, OwnerID: valid.OwnerID, Scope: economy.ScopeCompany},
		{OwnerKind: OwnerWorld, OwnerID: valid.OwnerID, Scope: economy.ScopeFounder},
		{OwnerKind: OwnerFounder, OwnerID: "not-a-uuid", Scope: economy.ScopeCompany},
	}
	for _, key := range invalid {
		if err := validateWrite(key, validHash, state, context); !errors.Is(err, ErrInvalidStream) {
			t.Fatalf("validateWrite(%+v) error = %v", key, err)
		}
	}
}
