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

func TestComputeTermsClampsReputationToFounderHeadroom(t *testing.T) {
	data, err := os.ReadFile("../../balance/prestige/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := LoadPolicy(data)
	if err != nil {
		t.Fatal(err)
	}
	company := &save.State{LifetimeValue: decimal.New(1, 300)}
	founder := &save.State{ReputationLevel: decimal.MaxExactInteger - 7}
	terms, err := ComputeTerms(company, founder, policy, "ipo")
	if err != nil || terms.ReputationDelta != 7 {
		t.Fatalf("terms=%+v err=%v", terms, err)
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

func TestOfflineSpanCollapsePreservesExactDuration(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	state := &save.State{RunStartedAt: start, EvaluatedThrough: start.Add(600 * time.Hour), OfflineSpans: []save.OfflineSpan{}}
	for index := 0; index < 257; index++ {
		from := start.Add(time.Duration(index*2) * time.Hour)
		if err := RecordOfflineSpan(state, from, from.Add(time.Hour), 1); err != nil {
			t.Fatal(err)
		}
	}
	if len(state.OfflineSpans) != 256 || state.CollapsedOfflineMS != int64(time.Hour/time.Millisecond) {
		t.Fatalf("spans=%d collapsed=%d", len(state.OfflineSpans), state.CollapsedOfflineMS)
	}
	attended, err := AttendedMS(state, state.EvaluatedThrough)
	want := int64((600*time.Hour - 257*time.Hour) / time.Millisecond)
	if err != nil || attended != want {
		t.Fatalf("attended=%d want=%d err=%v", attended, want, err)
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
