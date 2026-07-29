package save

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
)

const stateCatalogJSON = `{
  "schema_version": 3,
  "resources": [{
    "id": "company.cash", "scope": "company", "numeric_kind": "decimal",
    "initial": "0", "minimum": "0",
    "hardcap": {"amount": "1e100", "reason_key": "resource.company_cash.cap.test"}
  }],
  "generator_classes": [{
    "id": "generator.example",
    "price": {"resource_id":"company.cash","base":"1e0","curve":{"kind":"constant"}},
    "production": {"resource_id":"company.cash","base_rate":"1e0"}
  }],
  "manual_actions": [{
    "id":"manual.click","output":{"resource_id":"company.cash","amount_per_action":"1e0"}
  }],
  "multiplier_sources": [],
  "progress_coordinates": [
    {"tier":0,"kind":"resource_log","resource":"company.cash","target":"1e3"},
    {"tier":1,"kind":"count_fraction","count":"generators.total_owned","required":25},
    {"tier":2,"kind":"resource_log","resource":"company.cash","target":"1e9"},
    {"tier":3,"kind":"resource_log","resource":"company.cash","target":"1e12"}
  ],
  "manual_policy":{"refill_milli_per_ms":25,"bucket_cap_milli":50000},
  "offline_policy":{"efficiency":"9e-1","accrual_cap_ms":86400000,
    "bank_ratio_numerator":1,"bank_ratio_denominator":2,"bank_cap_ms":259200000,
    "burst_speed":"2e0","burst_max_duration_ms":14400000}
}`

var testCursor = time.Date(2026, 7, 28, 8, 0, 0, 123_000_000, time.UTC)

type migrationFixture struct {
	CorpusVersion int             `json:"corpus_version"`
	Cases         []migrationCase `json:"cases"`
}

type migrationCorpusBaseline struct {
	SchemaVersion     int      `json:"schema_version"`
	MinimumCaseCount  int      `json:"minimum_case_count"`
	RequiredCaseNames []string `json:"required_case_names"`
}

type migrationCase struct {
	Name              string          `json:"name"`
	FromVersion       int             `json:"from_version"`
	Scope             economy.Scope   `json:"scope"`
	MigrationBaseline string          `json:"migration_baseline"`
	Input             json.RawMessage `json:"input"`
	ExpectV5          json.RawMessage `json:"expect_v5"`
	ExpectError       bool            `json:"expect_error"`
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
		ComputeCreditMS: 1234, ManualTokenMilli: 12_345, ManualTokenRefilledAt: testCursor.Add(-time.Second),
		GatesCrossed: map[string]bool{"gate.example": true}, RunSeq: 3,
		DoctrinesByTransition: map[string]string{"transition.example": "doctrine.example"},
		StructureID:           "structure.example", LedgerFactKinds: map[string]bool{"exit.example": true},
		MeterBands: map[string]int{"trust.example.standing": 70}, RegionTraits: map[string]bool{"trait.example": true},
		HintsUnlocked: map[string]bool{},
	}
}

func TestStateV6RoundTrip(t *testing.T) {
	state := testState(t)
	state.CompactMember = true
	state.CompactTithePPM = 100_000
	state.CompactSolidarityPPM = 875_000
	state.CompactSamples = []CompactSample{{HourStart: testCursor.Truncate(time.Hour), CompliancePPM: 875_000, CoveredMS: 3_600_000}}
	encoded, err := EncodeState(state)
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
	if restored.ComputeCreditMS != 1234 || restored.ManualTokenMilli != 12_345 ||
		!restored.ManualTokenRefilledAt.Equal(testCursor.Add(-time.Second)) {
		t.Fatalf("restored production state = %+v", restored)
	}
	if restored.RunSeq != 3 || !restored.GatesCrossed["gate.example"] ||
		restored.DoctrinesByTransition["transition.example"] != "doctrine.example" ||
		restored.StructureID != "structure.example" || !restored.LedgerFactKinds["exit.example"] ||
		restored.MeterBands["trust.example.standing"] != 70 || !restored.RegionTraits["trait.example"] {
		t.Fatalf("restored route state = %+v", restored)
	}
	if !restored.CompactMember || restored.CompactTithePPM != 100_000 || restored.CompactSolidarityPPM != 875_000 || len(restored.CompactSamples) != 1 || restored.CompactSamples[0].CoveredMS != 3_600_000 {
		t.Fatalf("restored compact state = %+v", restored)
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
	baselineData, err := os.ReadFile("../../testdata/save-migrations-baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	var baseline migrationCorpusBaseline
	if err := json.Unmarshal(baselineData, &baseline); err != nil {
		t.Fatal(err)
	}
	if fixture.CorpusVersion != 5 || baseline.SchemaVersion != 1 || baseline.MinimumCaseCount < 1 ||
		len(fixture.Cases) != baseline.MinimumCaseCount {
		t.Fatalf("migration corpus version=%d cases=%d baseline=%+v", fixture.CorpusVersion, len(fixture.Cases), baseline)
	}
	caseNames := make(map[string]bool, len(fixture.Cases))
	for _, vector := range fixture.Cases {
		if vector.Name == "" || caseNames[vector.Name] {
			t.Fatalf("missing or duplicate migration case name %q", vector.Name)
		}
		caseNames[vector.Name] = true
	}
	for _, required := range baseline.RequiredCaseNames {
		if !caseNames[required] {
			t.Fatalf("required migration case %q was removed", required)
		}
	}
	for _, vector := range fixture.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			baseline, err := time.Parse(time.RFC3339Nano, vector.MigrationBaseline)
			if err != nil {
				t.Fatal(err)
			}
			restored, err := RestoreState(vector.Input, vector.FromVersion, stateCatalog(t), vector.Scope, baseline)
			if vector.ExpectError {
				if !errors.Is(err, ErrInvalidState) {
					t.Fatalf("migration error = %v, want ErrInvalidState", err)
				}
				return
			}
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
			if err := json.Unmarshal(vector.ExpectV5, &want); err != nil {
				t.Fatal(err)
			}
			if !equalJSON(got, want) {
				t.Fatalf("migrated JSON = %s, want %s", encoded, vector.ExpectV5)
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
		withProductionState := strings.TrimSuffix(data, "}") +
			`,"compute_credit_ms":0,"manual_token_milli":50000,"manual_token_refilled_at":"2026-07-28T08:00:00Z"}`
		if _, err := RestoreState([]byte(withProductionState), CurrentVersion, stateCatalog(t), economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
			t.Fatalf("RestoreState(%s) error = %v", data, err)
		}
	}
}

func TestRestoreRejectsProductionPolicyViolations(t *testing.T) {
	tests := []string{
		`{"balances":{"company.cash":"0"},"generators":{"generator.example":0},"evaluated_through":"2026-07-28T08:00:00Z","compute_credit_ms":259200001,"manual_token_milli":50000,"manual_token_refilled_at":"2026-07-28T08:00:00Z"}`,
		`{"balances":{"company.cash":"0"},"generators":{"generator.example":0},"evaluated_through":"2026-07-28T08:00:00Z","compute_credit_ms":0,"manual_token_milli":50001,"manual_token_refilled_at":"2026-07-28T08:00:00Z"}`,
		`{"balances":{"company.cash":"0"},"generators":{"generator.example":0},"evaluated_through":"2026-07-28T08:00:00Z","compute_credit_ms":0,"manual_token_milli":50000,"manual_token_refilled_at":"2026-07-28T08:00:01Z"}`,
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

func TestEncodeRejectsNonCanonicalProductionCursors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*State)
	}{
		{name: "evaluated sub-millisecond", mutate: func(state *State) {
			state.EvaluatedThrough = state.EvaluatedThrough.Add(time.Nanosecond)
		}},
		{name: "manual sub-millisecond", mutate: func(state *State) {
			state.ManualTokenRefilledAt = state.ManualTokenRefilledAt.Add(time.Nanosecond)
		}},
		{name: "non-UTC location", mutate: func(state *State) {
			state.EvaluatedThrough = state.EvaluatedThrough.In(time.FixedZone("test", 3600))
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := testState(t)
			test.mutate(state)
			if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
				t.Fatalf("error = %v, want ErrInvalidState", err)
			}
		})
	}
}

func TestConstantsHashUsesExactArtifactBytes(t *testing.T) {
	first := ConstantsHash([]byte("{}"))
	second := ConstantsHash([]byte("{}\n"))
	if first == second || len(first) != len("sha256:")+64 {
		t.Fatalf("unexpected hashes %q %q", first, second)
	}
}

func TestConstantsHashArtifactsIsNamedOrderedAndExact(t *testing.T) {
	first, err := ConstantsHashArtifacts(map[string][]byte{
		"economy": []byte("{}"),
		"commons": []byte("{\"weight\":1}"),
	})
	if err != nil {
		t.Fatal(err)
	}
	reordered, err := ConstantsHashArtifacts(map[string][]byte{
		"commons": []byte("{\"weight\":1}"),
		"economy": []byte("{}"),
	})
	if err != nil || reordered != first {
		t.Fatalf("reordered hash=%q want=%q err=%v", reordered, first, err)
	}
	commonsChanged, err := ConstantsHashArtifacts(map[string][]byte{
		"economy": []byte("{}"),
		"commons": []byte("{\"weight\":2}"),
	})
	if err != nil || commonsChanged == first {
		t.Fatalf("commons change hash=%q original=%q err=%v", commonsChanged, first, err)
	}
	economyChanged, err := ConstantsHashArtifacts(map[string][]byte{
		"economy": []byte("{}\n"),
		"commons": []byte("{\"weight\":1}"),
	})
	if err != nil || economyChanged == first || len(first) != len("sha256:")+64 {
		t.Fatalf("economy change hash=%q original=%q err=%v", economyChanged, first, err)
	}
	if _, err := ConstantsHashArtifacts(nil); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("empty bundle err=%v", err)
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

func TestNewCompanyStreamRequiresAlignedCursors(t *testing.T) {
	state := testState(t)
	if err := validateInitialCursors(economy.ScopeCompany, state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("mismatched cursor error = %v, want ErrInvalidState", err)
	}
	state.ManualTokenRefilledAt = state.EvaluatedThrough
	if err := validateInitialCursors(economy.ScopeCompany, state); err != nil {
		t.Fatal(err)
	}
}
