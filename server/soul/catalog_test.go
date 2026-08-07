package soul

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func fixtureCatalogBytes(t *testing.T) []byte {
	t.Helper()
	value := map[string]any{
		"schema_version": 1,
		"policy": map[string]any{"soul_floor": 0, "soul_initial": 100, "soul_max": 100,
			"recovery_beat_ceiling_ms": 5000, "max_session_wall_ms": 86400000},
		"bands": []any{
			map[string]any{"band_member": "near_zero", "min_inclusive": 0, "max_inclusive": 9, "human_content_locked": true, "reason_key": "category.low_percent"},
			map[string]any{"band_member": "hollow", "min_inclusive": 10, "max_inclusive": 39, "human_content_locked": false, "reason_key": "category.ethical_percent"},
			map[string]any{"band_member": "dimming", "min_inclusive": 40, "max_inclusive": 74, "human_content_locked": false, "reason_key": "category.hundred_percent"},
			map[string]any{"band_member": "whole", "min_inclusive": 75, "max_inclusive": 100, "human_content_locked": false, "reason_key": "category.any_percent"},
		},
		"debit_sources":       []any{map[string]any{"source_id": "soul.fixture", "owner_kind": "fixture", "amount": 20, "may_exhaust": true, "single_use": true, "curtain_copy_key": "category.valuation"}},
		"recovery_activities": []any{map[string]any{"activity_id": "touch_grass.fixture", "duration_attended_ms": 5000, "recovery_amount": 15, "reason_key": "category.any_percent"}},
		"ending_policy":       map[string]any{"whole_variant": "earnest_ascension", "depleted_variant": "training_data"},
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func fixtureDeclarations(epoch bool) Declarations {
	return Declarations{EpochSeeded: epoch, CatchupCeilingMS: 5000, CopyKeys: map[string]struct{}{
		"category.low_percent": {}, "category.ethical_percent": {}, "category.hundred_percent": {}, "category.any_percent": {}, "category.valuation": {},
	}}
}

func TestLoadCatalogExactGrammarAndBands(t *testing.T) {
	catalog, err := LoadCatalog(fixtureCatalogBytes(t), fixtureDeclarations(false))
	if err != nil {
		t.Fatal(err)
	}
	for value, want := range map[int64]BandMember{0: BandNearZero, 9: BandNearZero, 10: BandHollow, 39: BandHollow, 40: BandDimming, 74: BandDimming, 75: BandWhole, 100: BandWhole} {
		band, ok := catalog.BandFor(value)
		if !ok || band.Member != want {
			t.Fatalf("BandFor(%d) = %#v,%v; want %s", value, band, ok, want)
		}
	}
	locked, err := catalog.HumanContentLocked(0)
	if err != nil || !locked {
		t.Fatalf("near-zero gate = %v,%v", locked, err)
	}
}

func TestLoadCatalogRejectsMissingNestedKeysAndFixtureEpochRows(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal(fixtureCatalogBytes(t), &root); err != nil {
		t.Fatal(err)
	}
	delete(root["policy"].(map[string]any), "soul_initial")
	data, _ := json.Marshal(root)
	if _, err := LoadCatalog(data, fixtureDeclarations(false)); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatal("missing nested key accepted")
	}
	if _, err := LoadCatalog(fixtureCatalogBytes(t), fixtureDeclarations(true)); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatal("fixture source accepted in epoch-seeded artifact")
	}
}

func TestLoadCatalogRejectsHeartbeatCeilingAboveGlobalCatchup(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal(fixtureCatalogBytes(t), &root); err != nil {
		t.Fatal(err)
	}
	root["policy"].(map[string]any)["recovery_beat_ceiling_ms"] = float64(5001)
	data, _ := json.Marshal(root)
	if _, err := LoadCatalog(data, fixtureDeclarations(false)); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatal("heartbeat ceiling above global catch-up accepted")
	}
}

func TestLoadCatalogRejectsMaxExactNonterminalBandParityVector(t *testing.T) {
	data, err := os.ReadFile("../../testdata/soul/catalog-invalid-max-exact-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(data, fixtureDeclarations(false)); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatal("MaxExactInteger nonterminal band accepted")
	}
}

func TestApplyDebitFullPriceOnceAndDatedDepletion(t *testing.T) {
	catalog, err := LoadCatalog(fixtureCatalogBytes(t), fixtureDeclarations(false))
	if err != nil {
		t.Fatal(err)
	}
	state := DebitState{Soul: 12, ExhaustedSourceIDs: []string{}}
	result, err := ApplyDebit(&state, catalog, DebitCommand{SourceID: "soul.fixture", OwnerKind: OwnerFixture, EligibilityRef: "offer.fixture"})
	if err != nil || result.Debit != 12 || result.SoulAfter != 0 || !result.DepletedFirstTime || !state.DepletedFact {
		t.Fatalf("exhaust debit = %#v state=%#v err=%v", result, state, err)
	}
	if _, err := ApplyDebit(&state, catalog, DebitCommand{SourceID: "soul.fixture", OwnerKind: OwnerFixture, EligibilityRef: "offer.fixture.2"}); !errors.Is(err, ErrSourceConsumed) {
		t.Fatalf("repeat exhaust error = %v", err)
	}
	full := DebitState{Soul: 12, ExhaustedSourceIDs: []string{}}
	catalog.DebitSources[0].MayExhaust, catalog.DebitSources[0].SingleUse = false, false
	if _, err := ApplyDebit(&full, catalog, DebitCommand{SourceID: "soul.fixture", OwnerKind: OwnerFixture, EligibilityRef: "offer.fixture"}); !errors.Is(err, ErrUnaffordable) {
		t.Fatalf("partial ordinary debit error = %v", err)
	}
}
