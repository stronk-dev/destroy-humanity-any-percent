package prestige

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

func TestPromiseTermsIsFieldwiseMonotonicUnion(t *testing.T) {
	preview := Terms{ReputationDelta: 8, RouteKnowledge: 3, CloutReachNote: "preview", NetworkSlotUnlocks: []save.NetworkSlot{{Slot: "network.alpha", CarriedRef: "upgrade.alpha"}}}
	current := Terms{ReputationDelta: 5, RouteKnowledge: 7, CloutReachNote: "current", NetworkSlotUnlocks: []save.NetworkSlot{{Slot: "network.beta", CarriedRef: "upgrade.beta"}}}
	got := PromiseTerms(preview, current)
	if got.ReputationDelta != 8 || got.RouteKnowledge != 7 || len(got.NetworkSlotUnlocks) != 2 || got.NetworkSlotUnlocks[0].Slot != "network.alpha" {
		t.Fatalf("terms=%+v", got)
	}
}

func TestLifetimeOfflineAndNewRunDeterminism(t *testing.T) {
	catalogBytes, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(catalogBytes)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 29, 18, 0, 0, 0, time.UTC)
	companyLedger, _ := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": "1e3"})
	founderLedger, _ := economy.NewLedger(catalog, economy.ScopeFounder)
	company := &save.State{Ledger: companyLedger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0}, EvaluatedThrough: now,
		ManualTokenRefilledAt: now, GatesCrossed: map[string]bool{}, RunSeq: 4, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.New(5, 2), RunStartedAt: now.Add(-time.Hour), OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	founder := &save.State{Ledger: founderLedger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: now, ManualTokenRefilledAt: now,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.Zero, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}, Notoriety: 20}
	receipt := economy.Receipt{Changes: []economy.Change{{ResourceID: "company.cash", Delta: "2e2"}}}
	if err := AccumulateLifetimeValue(company, receipt, "company.cash"); err != nil || company.LifetimeValue.String() != "7e2" {
		t.Fatalf("lifetime=%s err=%v", company.LifetimeValue.String(), err)
	}
	if err := RecordOfflineSpan(company, now.Add(-40*time.Minute), now.Add(-30*time.Minute), 5_000); err != nil {
		t.Fatal(err)
	}
	attended, err := AttendedMS(company, now)
	if err != nil || attended != int64(50*time.Minute/time.Millisecond) {
		t.Fatalf("attended=%d err=%v", attended, err)
	}
	first, err := NewRunState(catalog, company, founder, now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewRunState(catalog, company, founder, now)
	if err != nil {
		t.Fatal(err)
	}
	left, _ := save.EncodeState(first)
	right, _ := save.EncodeState(second)
	if string(left) != string(right) || first.RunSeq != 5 || first.MeterBands["trust.regulators.standing"] != 83 || first.MeterBands["trust.regulators.grievance"] != 17 {
		t.Fatalf("first=%s second=%s", left, right)
	}
}

func TestOfferIdentityAndTermsAreReplayable(t *testing.T) {
	at := time.Date(2026, 7, 29, 18, 0, 0, 0, time.UTC)
	first := OfferID("01985555-1000-7000-8000-000000000001", 3, 4, 2, at)
	second := OfferID("01985555-1000-7000-8000-000000000001", 3, 4, 2, at)
	if first != second || first[14] != '7' {
		t.Fatalf("offer IDs %q %q", first, second)
	}
	terms := Terms{ReputationDelta: 10, NetworkSlotUnlocks: []save.NetworkSlot{}, RouteKnowledge: 5, CloutReachNote: "clout.reach.preserved"}
	encoded, _ := json.Marshal(StoredOfferTerms{PayoutPreview: terms, MarketModifierPPM: 950_000})
	decoded, err := DecodeStoredOfferTerms(encoded)
	if err != nil || decoded.PayoutPreview.ReputationDelta != 10 || decoded.MarketModifierPPM != 950_000 {
		t.Fatalf("decoded=%+v err=%v", decoded, err)
	}
	if got := DriftTerms(Terms{ReputationDelta: decimal.MaxExactInteger, NetworkSlotUnlocks: []save.NetworkSlot{}, CloutReachNote: "x"}, 10, 50_000, true); got.ReputationDelta != decimal.MaxExactInteger {
		t.Fatalf("drift overflow=%d", got.ReputationDelta)
	}
}
