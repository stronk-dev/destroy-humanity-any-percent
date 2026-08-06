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
	"cloud-clicker/server/pet"
)

const stateCatalogJSON = `{
  "schema_version": 3,
  "resources": [{
    "id": "company.cash", "scope": "company", "numeric_kind": "decimal",
    "initial": "0", "minimum": "0",
    "hardcap": {"amount": "1e100", "reason_key": "resource.company_cash.cap.test"}
  }, {
    "id": "founder.reputation", "scope": "founder", "numeric_kind": "decimal",
    "initial": "0", "minimum": "0",
    "hardcap": {"amount": "1e100", "reason_key": "resource.founder_reputation.cap.test"}
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

func foundationStateCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../testdata/economy-foundation-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
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

func TestStateV10RoundTrip(t *testing.T) {
	state := testState(t)
	state.CompactMember = true
	state.CompactTithePPM = 100_000
	state.CompactSolidarityPPM = 875_000
	state.CompactSamples = []CompactSample{{HourStart: testCursor.Truncate(time.Hour), CompliancePPM: 875_000, CoveredMS: 3_600_000}}
	state.Tier = 3
	state.LifetimeValue = decimal.New(25, 11)
	state.RunStartedAt = testCursor.Add(-time.Hour)
	state.RunPreTimer = true
	state.OfflineSpans = []OfflineSpan{{From: testCursor.Add(-30 * time.Minute), To: testCursor.Add(-20 * time.Minute)}}
	state.CollapsedOfflineMS = 1_800_000
	state.FactionID = "open_source"
	state.IncorporatedAt = testCursor.Add(-45 * time.Minute)
	state.StockUnits = 42
	state.StockProgressMS = 12_345
	state.ConsumedStockUnits = 7
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
	if restored.Tier != 3 || restored.LifetimeValue.String() != "2.5e12" || !restored.RunStartedAt.Equal(testCursor.Add(-time.Hour)) || !restored.RunPreTimer ||
		len(restored.OfflineSpans) != 1 || !restored.OfflineSpans[0].To.Equal(testCursor.Add(-20*time.Minute)) || restored.CollapsedOfflineMS != 1_800_000 {
		t.Fatalf("restored prestige state = %+v", restored)
	}
	if restored.FactionID != "open_source" || !restored.IncorporatedAt.Equal(testCursor.Add(-45*time.Minute)) ||
		restored.StockUnits != 42 || restored.StockProgressMS != 12_345 || restored.ConsumedStockUnits != 7 {
		t.Fatalf("restored faction state = %+v", restored)
	}
}

func TestStateV14PurchasableContentRoundTripAndClosure(t *testing.T) {
	catalog := foundationStateCatalog(t)
	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": "1e4"})
	if err != nil {
		t.Fatal(err)
	}
	state := testState(t)
	state.Ledger = ledger
	state.GeneratorCounts = map[string]int64{"generator.high": 3, "generator.low": 7}
	state.GeneratorPurchasedTotal = 10
	state.UpgradesOwned = map[string]bool{"upgrade.click": true}
	state.GeneratorProvisioned = map[string]int64{"generator.high": 0, "generator.low": 11}
	state.ProvisionRemaindersPPM = map[string]int64{"generator.low": 987_654}
	state.StockRateRemainderPPM = 123_456
	state.FactionID = "open_source"
	state.IncorporatedAt = testCursor.Add(-time.Minute)
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, CurrentVersion, catalog, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if !restored.UpgradesOwned["upgrade.click"] || restored.GeneratorProvisioned["generator.low"] != 11 || restored.ProvisionRemaindersPPM["generator.low"] != 987_654 || restored.StockRateRemainderPPM != 123_456 {
		t.Fatalf("restored foundation state = %+v", restored)
	}
	var legacy map[string]any
	if err := json.Unmarshal(encoded, &legacy); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"upgrades_owned", "generators_provisioned", "provision_remainders_ppm", "stock_rate_remainder_ppm"} {
		delete(legacy, field)
	}
	v13, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	migrated, err := RestoreState(v13, 13, catalog, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(migrated.UpgradesOwned) != 0 || migrated.GeneratorProvisioned["generator.low"] != 0 || migrated.ProvisionRemaindersPPM["generator.low"] != 0 || migrated.StockRateRemainderPPM != 0 {
		t.Fatalf("v13 migration = %+v", migrated)
	}

	for _, field := range []string{"upgrades_owned", "generators_provisioned", "provision_remainders_ppm", "stock_rate_remainder_ppm"} {
		t.Run("missing-"+field, func(t *testing.T) {
			var object map[string]any
			if err := json.Unmarshal(encoded, &object); err != nil {
				t.Fatal(err)
			}
			delete(object, field)
			malformed, err := json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := RestoreState(malformed, CurrentVersion, catalog, economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
				t.Fatalf("missing %s error=%v", field, err)
			}
		})
	}

	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	object["upgrades_owned"] = []any{"upgrade.unknown"}
	object["generators_provisioned"].(map[string]any)["generator.low"] = float64(decimal.MaxExactInteger)
	malformed, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RestoreState(malformed, CurrentVersion, catalog, economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("unknown upgrade/cap overflow error=%v", err)
	}
}

func TestStateV16FoundationRoundTripAndVersionAuthority(t *testing.T) {
	state := testState(t)
	state.WireVersion = 16
	state.MeterValues = map[string]int{"doom.probability": 17, "trust.users.standing": 63}
	state.MeterDecayRemainders = map[string]int64{"doom.probability": 42, "trust.users.standing": 0}
	state.MeterInputRemainders = map[string]int64{"trust.users.standing:0": 99}
	state.AchievementsEarnedRun = map[string]bool{"achievement.first": true}
	state.AchievementScoreRun = 7

	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	if _, present := object["meter_bands"]; present {
		t.Fatal("v16 retained the superseded meter_bands field")
	}
	restored, err := RestoreState(encoded, 16, stateCatalog(t), economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if VersionForState(restored) != 16 || restored.MeterValues["doom.probability"] != 17 ||
		restored.MeterDecayRemainders["doom.probability"] != 42 || restored.MeterInputRemainders["trust.users.standing:0"] != 99 ||
		!restored.AchievementsEarnedRun["achievement.first"] || restored.AchievementScoreRun != 7 {
		t.Fatalf("restored foundation state = %+v", restored)
	}
	if _, err := EncodeStateVersion(restored, 14); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("v16 mechanics downgraded to v14: %v", err)
	}
}

func TestCompanyV17ComputeBurstRoundTripAndExactEnvelope(t *testing.T) {
	state := testState(t)
	state.WireVersion = 17
	state.MeterValues = map[string]int{"doom.probability": 17, "trust.users.standing": 63}
	state.MeterDecayRemainders = map[string]int64{"doom.probability": 42, "trust.users.standing": 0}
	state.MeterInputRemainders = map[string]int64{"trust.users.standing:0": 99}
	state.AchievementsEarnedRun = map[string]bool{"achievement.first": true}
	state.AchievementScoreRun = 7
	state.ComputeBurstRemainingMS = 12_345

	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	if _, present := object["compute_burst_remaining_ms"]; !present {
		t.Fatal("Company v17 omitted compute_burst_remaining_ms")
	}
	if _, present := object["minigame_ratings"]; present {
		t.Fatal("Company v17 leaked Founder minigame state")
	}
	restored, err := RestoreState(encoded, 17, stateCatalog(t), economy.ScopeCompany, time.Time{})
	if err != nil || VersionForState(restored) != 17 || restored.ComputeBurstRemainingMS != 12_345 {
		t.Fatalf("restored Company v17=%+v err=%v", restored, err)
	}
	delete(object, "compute_burst_remaining_ms")
	missing, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RestoreState(missing, 17, stateCatalog(t), economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("Company v17 accepted missing burst field: %v", err)
	}
}

func TestCompanyV18ActivePlayRoundTrip(t *testing.T) {
	state := testState(t)
	state.WireVersion = 18
	state.MeterValues = map[string]int{}
	state.MeterDecayRemainders = map[string]int64{}
	state.MeterInputRemainders = map[string]int64{}
	state.AchievementsEarnedRun = map[string]bool{}
	state.AchievementsEarnedLifetime = map[string]bool{}
	state.OpportunitySpawnSeq = 2
	state.NextOpportunityAttendedMS = 5000
	state.ActiveBuffs = []ActiveBuff{}
	selected := "generator.beige_tower"
	state.PendingOpportunity = &PendingOpportunity{OpportunityID: "01985555-0000-7000-8000-000000000001", SpawnedAttendedMS: 2000, ExpiresAttendedMS: 4000, EffectRowID: "active.building", SelectedGeneratorID: &selected}
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, 18, stateCatalog(t), economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if VersionForState(restored) != 18 || restored.OpportunitySpawnSeq != 2 || restored.NextOpportunityAttendedMS != 5000 || restored.PendingOpportunity == nil || *restored.PendingOpportunity.SelectedGeneratorID != selected || restored.ActiveBuffs == nil {
		t.Fatalf("v18=%+v", restored)
	}
	var object map[string]json.RawMessage
	_ = json.Unmarshal(encoded, &object)
	delete(object, "active_buffs")
	missing, _ := json.Marshal(object)
	if _, err := RestoreState(missing, 18, stateCatalog(t), economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("missing active_buffs accepted: %v", err)
	}
}

func TestFounderV17AndV18RoundTripWhileCompanyRejectsThem(t *testing.T) {
	catalog := stateCatalog(t)
	ledger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	state := &State{WireVersion: 17, Ledger: ledger, GeneratorCounts: map[string]int64{}, GeneratorProvisioned: map[string]int64{},
		ProvisionRemaindersPPM: map[string]int64{}, UpgradesOwned: map[string]bool{}, EvaluatedThrough: testCursor,
		ManualTokenRefilledAt: testCursor, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterValues: map[string]int{}, MeterDecayRemainders: map[string]int64{}, MeterInputRemainders: map[string]int64{},
		AchievementsEarnedRun: map[string]bool{}, AchievementsEarnedLifetime: map[string]bool{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []CompactSample{}, OfflineSpans: []OfflineSpan{}, NetworkSlots: []NetworkSlot{}, ExitHistory: []ExitRecord{},
		MinigameRatings: map[string]MinigameRatingState{}, MinigameOfflineQuality: map[string]MinigameOfflineQualityState{}}
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, 17, catalog, economy.ScopeFounder, time.Time{})
	if err != nil || restored.MinigameRatings == nil || restored.MinigameOfflineQuality == nil {
		t.Fatalf("v17 restore=%+v err=%v", restored, err)
	}
	if _, err := RestoreState(encoded, 17, catalog, economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("Company accepted v17: %v", err)
	}
	state.WireVersion = 18
	state.Pets = map[string]pet.CareState{"018f6b7c-9abc-7def-8abc-0123456789ac": {
		StatsPPM:                map[pet.StatID]int64{pet.StatHunger: 600_000, pet.StatEnergy: 700_000, pet.StatCleanliness: 800_000, pet.StatAffection: 900_000},
		StatDecayRemaindersPPM:  map[pet.StatID]int64{pet.StatHunger: 0, pet.StatEnergy: 0, pet.StatCleanliness: 0, pet.StatAffection: 0},
		CooldownUntilAttendedMS: map[string]int64{"care.feed": 0}, TrustPPM: 500_000,
		BehaviorState: pet.BehaviorIdle, BehaviorQueue: []pet.BehaviorQueueEntry{},
	}}
	state.ActiveBuffs = []ActiveBuff{}
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("Founder v18 accepted Company active-play state: %v", err)
	}
	state.ActiveBuffs = nil
	encoded, err = EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err = RestoreState(encoded, 18, catalog, economy.ScopeFounder, time.Time{})
	if err != nil || restored.Pets == nil || restored.Pets["018f6b7c-9abc-7def-8abc-0123456789ac"].BehaviorQueue == nil {
		t.Fatalf("v18 restore=%+v err=%v", restored, err)
	}
	state.MinigameRatings["combat.duel"] = MinigameRatingState{Elo: 1000, SeasonMember: "preseason"}
	state.MinigameOfflineQuality["combat.duel"] = MinigameOfflineQualityState{GradePPM: 500_000}
	state.WireVersion = 14
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("v14 silently discarded Founder feature state: %v", err)
	}
	state.WireVersion = 15
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("v15 silently discarded Founder feature state: %v", err)
	}
}

func TestFounderV19FiscalRoundTripAndExactEnvelope(t *testing.T) {
	catalog := stateCatalog(t)
	ledger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	state := &State{WireVersion: 19, Ledger: ledger, GeneratorCounts: map[string]int64{}, GeneratorProvisioned: map[string]int64{},
		ProvisionRemaindersPPM: map[string]int64{}, UpgradesOwned: map[string]bool{}, EvaluatedThrough: testCursor,
		ManualTokenRefilledAt: testCursor, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterValues: map[string]int{}, MeterDecayRemainders: map[string]int64{}, MeterInputRemainders: map[string]int64{},
		AchievementsEarnedRun: map[string]bool{}, AchievementsEarnedLifetime: map[string]bool{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []CompactSample{}, OfflineSpans: []OfflineSpan{}, NetworkSlots: []NetworkSlot{}, ExitHistory: []ExitRecord{},
		MinigameRatings: map[string]MinigameRatingState{}, MinigameOfflineQuality: map[string]MinigameOfflineQualityState{}, Pets: map[string]pet.CareState{},
		FiscalCredit: 17, FiscalPeriodOpenedWallMS: 1_786_000_000_000, FiscalPeriodSequence: 9,
		FiscalGeneratorLevels: map[string]int64{"generator.beige_tower": 3}, FiscalUnlocks: map[string]bool{"unlock.arcade": true}}
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, 19, catalog, economy.ScopeFounder, time.Time{})
	if err != nil || restored.FiscalCredit != 17 || restored.FiscalPeriodOpenedWallMS != 1_786_000_000_000 || restored.FiscalPeriodSequence != 9 ||
		restored.FiscalGeneratorLevels["generator.beige_tower"] != 3 || !restored.FiscalUnlocks["unlock.arcade"] {
		t.Fatalf("v19 restore=%+v err=%v", restored, err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"fiscal_credit", "fiscal_period_opened_wall_ms", "fiscal_period_seq", "fiscal_generator_levels", "fiscal_unlocks"} {
		candidate := make(map[string]json.RawMessage, len(object))
		for name, value := range object {
			candidate[name] = value
		}
		delete(candidate, key)
		missing, _ := json.Marshal(candidate)
		if _, err := RestoreState(missing, 19, catalog, economy.ScopeFounder, time.Time{}); !errors.Is(err, ErrInvalidState) {
			t.Fatalf("Founder v19 accepted missing %s: %v", key, err)
		}
	}
	if _, err := RestoreState(encoded, 19, catalog, economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("Company accepted Founder v19: %v", err)
	}
	state.WireVersion = 18
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("v18 silently discarded fiscal state: %v", err)
	}
}

func TestFounderV20SoulEligibilityRoundTripAndExactEnvelope(t *testing.T) {
	catalog := stateCatalog(t)
	ledger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	state := &State{WireVersion: 20, Ledger: ledger, GeneratorCounts: map[string]int64{}, GeneratorProvisioned: map[string]int64{},
		ProvisionRemaindersPPM: map[string]int64{}, UpgradesOwned: map[string]bool{}, EvaluatedThrough: testCursor,
		ManualTokenRefilledAt: testCursor, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterValues: map[string]int{}, MeterDecayRemainders: map[string]int64{}, MeterInputRemainders: map[string]int64{},
		AchievementsEarnedRun: map[string]bool{}, AchievementsEarnedLifetime: map[string]bool{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []CompactSample{}, OfflineSpans: []OfflineSpan{}, NetworkSlots: []NetworkSlot{}, ExitHistory: []ExitRecord{},
		MinigameRatings: map[string]MinigameRatingState{}, MinigameOfflineQuality: map[string]MinigameOfflineQualityState{}, Pets: map[string]pet.CareState{},
		FiscalPeriodOpenedWallMS: 1_786_000_000_000, FiscalGeneratorLevels: map[string]int64{}, FiscalUnlocks: map[string]bool{},
		Soul: 73, SoulExhaustedSourceIDs: []string{"soul.source.alpha", "soul.source.beta"}}
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, 20, catalog, economy.ScopeFounder, time.Time{})
	if err != nil || restored.Soul != 73 || strings.Join(restored.SoulExhaustedSourceIDs, ",") != "soul.source.alpha,soul.source.beta" {
		t.Fatalf("v20 restore=%+v err=%v", restored, err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	delete(object, "soul_exhausted_source_ids")
	missing, _ := json.Marshal(object)
	if _, err := RestoreState(missing, 20, catalog, economy.ScopeFounder, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("Founder v20 accepted missing exhausted-source set: %v", err)
	}
	state.WireVersion = 19
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("v19 silently discarded active Soul state: %v", err)
	}
	state.WireVersion = 20
	state.SoulExhaustedSourceIDs = []string{"soul.source.beta", "soul.source.alpha"}
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("v20 accepted unsorted Soul eligibility: %v", err)
	}
	if _, err := RestoreState(encoded, 20, catalog, economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("Company accepted Founder v20: %v", err)
	}
}

func TestStateV15AndV16CollectionsFailClosed(t *testing.T) {
	state := testState(t)
	state.WireVersion = 15
	state.MeterValues = map[string]int{}
	state.MeterDecayRemainders = map[string]int64{}
	state.MeterInputRemainders = map[string]int64{}
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	delete(object, "meter_values")
	missing, _ := json.Marshal(object)
	if _, err := RestoreState(missing, 15, stateCatalog(t), economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("v15 missing meter_values error = %v", err)
	}

	state.WireVersion = 16
	state.AchievementsEarnedRun = map[string]bool{}
	encoded, err = EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	delete(object, "achievement_score_run")
	missing, _ = json.Marshal(object)
	if _, err := RestoreState(missing, 16, stateCatalog(t), economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("v16 missing achievement_score_run error = %v", err)
	}

	state.WireVersion = 15
	state.AchievementsEarnedRun = map[string]bool{"achievement.too_early": true}
	state.AchievementScoreRun = 1
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("v15 silently discarded achievement state: %v", err)
	}
}

func TestV15CodecIsDecodeOnlyAtPersistenceBoundary(t *testing.T) {
	state := testState(t)
	state.WireVersion = 15
	state.MeterValues = map[string]int{}
	state.MeterDecayRemainders = map[string]int64{}
	state.MeterInputRemainders = map[string]int64{}
	if _, err := EncodeStateVersion(state, 15); err != nil {
		t.Fatalf("v15 migration codec unavailable: %v", err)
	}
	hash := ConstantsHash([]byte(stateCatalogJSON))
	store := &Store{catalogs: catalogMap{hash: stateCatalog(t)}}
	if _, err := store.validatedState(hash, economy.ScopeCompany, state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("persistence accepted standalone v15: %v", err)
	}
}

func TestRestoreV16RejectsSupersededCrossScopeAndUnsortedState(t *testing.T) {
	state := testState(t)
	state.WireVersion = 16
	state.MeterValues = map[string]int{}
	state.MeterDecayRemainders = map[string]int64{}
	state.MeterInputRemainders = map[string]int64{}
	state.AchievementsEarnedRun = map[string]bool{"achievement.a": true, "achievement.b": true}
	state.AchievementScoreRun = 2
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	var baseline map[string]any
	if json.Unmarshal(encoded, &baseline) != nil {
		t.Fatal("decode baseline")
	}
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "superseded meter bands", mutate: func(value map[string]any) { value["meter_bands"] = map[string]any{} }},
		{name: "case folded superseded meter bands", mutate: func(value map[string]any) { value["METER_BANDS"] = map[string]any{} }},
		{name: "founder ownership in company", mutate: func(value map[string]any) {
			value["achievements_earned_lifetime"] = []any{"achievement.a"}
			value["achievement_score_lifetime"] = float64(1)
		}},
		{name: "negative run score", mutate: func(value map[string]any) { value["achievement_score_run"] = float64(-1) }},
		{name: "unsorted run IDs", mutate: func(value map[string]any) {
			value["achievements_earned_run"] = []any{"achievement.b", "achievement.a"}
		}},
	}
	for _, item := range tests {
		t.Run(item.name, func(t *testing.T) {
			candidate := make(map[string]any, len(baseline))
			for key, value := range baseline {
				candidate[key] = value
			}
			item.mutate(candidate)
			data, _ := json.Marshal(candidate)
			if _, err := RestoreState(data, 16, stateCatalog(t), economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestLegacyRestoreTargetsCurrentWritableVersion(t *testing.T) {
	legacy := []byte(`{"balances":{"company.cash":"1e0"}}`)
	state, err := RestoreState(legacy, 1, stateCatalog(t), economy.ScopeCompany, testCursor)
	if err != nil {
		t.Fatal(err)
	}
	if VersionForState(state) != CurrentVersion {
		t.Fatalf("legacy writable version=%d want=%d", VersionForState(state), CurrentVersion)
	}
	if _, err := EncodeState(state); err != nil {
		t.Fatalf("legacy state could not migrate on write: %v", err)
	}
}

func TestStateV8MigratesCollapsedOfflineAccumulator(t *testing.T) {
	state := testState(t)
	state.Tier = 2
	state.RunStartedAt = testCursor.Add(-time.Hour)
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	var previous map[string]any
	if err := json.Unmarshal(encoded, &previous); err != nil {
		t.Fatal(err)
	}
	delete(previous, "collapsed_offline_ms")
	delete(previous, "faction_id")
	delete(previous, "incorporated_at_ms")
	delete(previous, "stock_units")
	delete(previous, "stock_progress_ms")
	delete(previous, "consumed_stock_units")
	delete(previous, "guild_tithe_carry_ppm")
	delete(previous, "guild_boundary_seq")
	delete(previous, "guild_consumed_window_units")
	delete(previous, "guild_boundary_guild_id")
	delete(previous, "generators_purchased_total")
	delete(previous, "upgrades_owned")
	delete(previous, "generators_provisioned")
	delete(previous, "provision_remainders_ppm")
	delete(previous, "stock_rate_remainder_ppm")
	v8, err := json.Marshal(previous)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(v8, 8, stateCatalog(t), economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if restored.CollapsedOfflineMS != 0 {
		t.Fatalf("collapsed offline = %d, want 0", restored.CollapsedOfflineMS)
	}
	current, err := EncodeState(restored)
	if err != nil {
		t.Fatal(err)
	}
	var migrated map[string]any
	if err := json.Unmarshal(current, &migrated); err != nil {
		t.Fatal(err)
	}
	if value, ok := migrated["collapsed_offline_ms"]; !ok || value != float64(0) {
		t.Fatalf("migrated collapsed_offline_ms=%v present=%v", value, ok)
	}
}

func TestGuildWatermarkV12PairAndLegacyMigration(t *testing.T) {
	state := testState(t)
	state.GuildBoundaryGuildID = "018f0000-0000-7000-8000-000000000012"
	state.GuildBoundarySeq = 10_000
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, CurrentVersion, stateCatalog(t), economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if restored.GuildBoundaryGuildID != state.GuildBoundaryGuildID || restored.GuildBoundarySeq != 10_000 {
		t.Fatalf("watermark=%q/%d", restored.GuildBoundaryGuildID, restored.GuildBoundarySeq)
	}

	var previous map[string]any
	if err := json.Unmarshal(encoded, &previous); err != nil {
		t.Fatal(err)
	}
	delete(previous, "guild_boundary_guild_id")
	delete(previous, "generators_purchased_total")
	delete(previous, "upgrades_owned")
	delete(previous, "generators_provisioned")
	delete(previous, "provision_remainders_ppm")
	delete(previous, "stock_rate_remainder_ppm")
	v11, err := json.Marshal(previous)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := RestoreState(v11, 11, stateCatalog(t), economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if legacy.GuildBoundaryGuildID != "" || legacy.GuildBoundarySeq != 10_000 || legacy.GeneratorPurchasedTotal != 42 {
		t.Fatalf("legacy watermark=%q/%d purchases=%d", legacy.GuildBoundaryGuildID, legacy.GuildBoundarySeq, legacy.GeneratorPurchasedTotal)
	}
	if _, err := EncodeState(legacy); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("unresolved legacy watermark encoded: %v", err)
	}

	state.GuildBoundaryGuildID = ""
	if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("unpaired current watermark encoded: %v", err)
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
	if fixture.CorpusVersion != 8 || baseline.SchemaVersion != 1 || baseline.MinimumCaseCount < 1 ||
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
			if vector.FromVersion < 7 {
				object := want.(map[string]any)
				object["tier"], object["lifetime_value"], object["offer_state"] = float64(0), "0", nil
				object["run_started_at_ms"], object["run_pre_timer"], object["offline_spans"] = float64(0), false, []any{}
				if vector.Scope == economy.ScopeCompany {
					evaluated, err := time.Parse(time.RFC3339Nano, object["evaluated_through"].(string))
					if err != nil {
						t.Fatal(err)
					}
					object["run_started_at_ms"], object["run_pre_timer"] = float64(evaluated.UnixMilli()), true
				}
				object["reputation_level"], object["reputation_unlock_ppm"] = float64(0), float64(0)
				object["network_slots"], object["clout_lifetime"] = []any{}, float64(0)
				object["soul"], object["age_ms"], object["notoriety"] = float64(0), float64(0), float64(0)
				object["advisor_mode"], object["exit_history"] = false, []any{}
			} else {
				want.(map[string]any)["run_pre_timer"] = false
			}
			want.(map[string]any)["collapsed_offline_ms"] = float64(0)
			want.(map[string]any)["faction_id"] = nil
			want.(map[string]any)["incorporated_at_ms"] = nil
			want.(map[string]any)["stock_units"] = float64(0)
			want.(map[string]any)["stock_progress_ms"] = float64(0)
			want.(map[string]any)["consumed_stock_units"] = float64(0)
			want.(map[string]any)["guild_tithe_carry_ppm"] = float64(0)
			want.(map[string]any)["guild_boundary_seq"] = float64(0)
			want.(map[string]any)["guild_consumed_window_units"] = float64(0)
			want.(map[string]any)["guild_boundary_guild_id"] = nil
			var purchased float64
			for _, value := range want.(map[string]any)["generators"].(map[string]any) {
				purchased += value.(float64)
			}
			want.(map[string]any)["generators_purchased_total"] = purchased
			want.(map[string]any)["upgrades_owned"] = []any{}
			provisioned := map[string]any{}
			for id := range want.(map[string]any)["generators"].(map[string]any) {
				provisioned[id] = float64(0)
			}
			want.(map[string]any)["generators_provisioned"] = provisioned
			want.(map[string]any)["provision_remainders_ppm"] = map[string]any{}
			want.(map[string]any)["stock_rate_remainder_ppm"] = float64(0)
			if !equalJSON(got, want) {
				t.Fatalf("migrated JSON = %s, want %s", encoded, vector.ExpectV5)
			}
			if _, err := RestoreState(encoded, CurrentVersion, stateCatalog(t), vector.Scope, time.Time{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFactionStateScopeAndPairing(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*State)
	}{
		{name: "stock before incorporation", mutate: func(state *State) { state.StockUnits = 1 }},
		{name: "id without time", mutate: func(state *State) { state.FactionID = "open_source" }},
		{name: "time without id", mutate: func(state *State) { state.IncorporatedAt = testCursor }},
		{name: "future incorporation", mutate: func(state *State) {
			state.FactionID, state.IncorporatedAt = "open_source", testCursor.Add(time.Millisecond)
		}},
		{name: "noncanonical incorporation", mutate: func(state *State) {
			state.FactionID, state.IncorporatedAt = "open_source", testCursor.Add(time.Nanosecond)
		}},
		{name: "negative stock", mutate: func(state *State) {
			state.FactionID, state.IncorporatedAt, state.StockUnits = "open_source", testCursor, -1
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := testState(t)
			test.mutate(state)
			if _, err := EncodeState(state); !errors.Is(err, ErrInvalidState) {
				t.Fatalf("error=%v", err)
			}
		})
	}

	ledger, err := economy.RestoreLedger(stateCatalog(t), economy.ScopeFounder, map[string]string{"founder.reputation": "0"})
	if err != nil {
		t.Fatal(err)
	}
	founder := &State{Ledger: ledger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: testCursor, ManualTokenRefilledAt: testCursor,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []CompactSample{}, LifetimeValue: decimal.Zero, OfflineSpans: []OfflineSpan{}, NetworkSlots: []NetworkSlot{}, ExitHistory: []ExitRecord{}, FactionID: "open_source", IncorporatedAt: testCursor}
	if _, err := EncodeState(founder); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("founder faction leak error=%v", err)
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
	_, err := RestoreState([]byte(`{}`), LatestSupportedVersion+1, stateCatalog(t), economy.ScopeCompany, time.Time{})
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("error = %v", err)
	}
}

func TestRestoreV13RequiresPurchaseAccumulatorPresence(t *testing.T) {
	encoded, err := EncodeState(testState(t))
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	delete(object, "generators_purchased_total")
	missing, _ := json.Marshal(object)
	if _, err := RestoreState(missing, CurrentVersion, stateCatalog(t), economy.ScopeCompany, time.Time{}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("missing v13 purchase accumulator err=%v", err)
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
		"routes":  []byte("{\"gate\":1}"),
		"commons": []byte("{\"weight\":1}"),
	})
	if err != nil {
		t.Fatal(err)
	}
	reordered, err := ConstantsHashArtifacts(map[string][]byte{
		"commons": []byte("{\"weight\":1}"),
		"economy": []byte("{}"),
		"routes":  []byte("{\"gate\":1}"),
	})
	if err != nil || reordered != first {
		t.Fatalf("reordered hash=%q want=%q err=%v", reordered, first, err)
	}
	commonsChanged, err := ConstantsHashArtifacts(map[string][]byte{
		"economy": []byte("{}"),
		"routes":  []byte("{\"gate\":1}"),
		"commons": []byte("{\"weight\":2}"),
	})
	if err != nil || commonsChanged == first {
		t.Fatalf("commons change hash=%q original=%q err=%v", commonsChanged, first, err)
	}
	economyChanged, err := ConstantsHashArtifacts(map[string][]byte{
		"economy": []byte("{}\n"),
		"routes":  []byte("{\"gate\":1}"),
		"commons": []byte("{\"weight\":1}"),
	})
	if err != nil || economyChanged == first || len(first) != len("sha256:")+64 {
		t.Fatalf("economy change hash=%q original=%q err=%v", economyChanged, first, err)
	}
	routesChanged, err := ConstantsHashArtifacts(map[string][]byte{
		"economy": []byte("{}"),
		"routes":  []byte("{\"gate\":2}"),
		"commons": []byte("{\"weight\":1}"),
	})
	if err != nil || routesChanged == first {
		t.Fatalf("Routes change hash=%q original=%q err=%v", routesChanged, first, err)
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
