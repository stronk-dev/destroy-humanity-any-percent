package production

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/doctrine"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func permitsCandidateCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/valid/permits-economy-candidate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestPermitsCandidatesComposeThroughProductionAndDoctrine(t *testing.T) {
	catalog := permitsCandidateCatalog(t)
	permits, ok := catalog.Resource("company.permits")
	if !ok || permits.Hardcap == nil || permits.Hardcap.Amount.String() != "2.4e1" || permits.Hardcap.ReasonKey != "resource.company_permits.cap.phase0" {
		t.Fatalf("permits resource = %+v exists=%v", permits, ok)
	}
	legal, ok := catalog.GeneratorClass("generator.legal_dept")
	if !ok || legal.Price.ResourceID != "company.cash" || legal.Production == nil || legal.Production.ResourceID != "company.permits" || legal.Production.BaseRate.String() != "1e-3" {
		t.Fatalf("legal department = %+v exists=%v", legal, ok)
	}

	routeBytes, err := os.ReadFile("../../balance/testdata/permits-t3-gate-candidate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(routeBytes)
	if err != nil {
		t.Fatal(err)
	}
	gate, ok := routeCatalog.Gate("gate.t3_to_t4")
	if !ok || len(gate.Requirement) != 2 || gate.Requirement[0].ResourceID != "company.cash" || gate.Requirement[1].ResourceID != "company.permits" {
		t.Fatalf("T3 gate = %+v exists=%v", gate, ok)
	}
	for _, requirement := range gate.Requirement {
		resource, exists := catalog.Resource(requirement.ResourceID)
		if !exists || resource.Scope != economy.ScopeCompany {
			t.Fatalf("gate resource %q is not a Company resource", requirement.ResourceID)
		}
	}

	doctrineBytes, err := os.ReadFile("../../balance/testdata/doctrines-catalog-parity-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if err := json.Unmarshal(doctrineBytes, &corpus); err != nil {
		t.Fatal(err)
	}
	doctrineCatalog, err := doctrine.LoadCatalog(corpus.Baseline)
	if err != nil {
		t.Fatal(err)
	}
	if err := doctrineCatalog.ValidateRoutes(routeCatalog); err != nil {
		t.Fatal(err)
	}
}

func TestPermitAccrualUsesMultiplierOfflinePolicyAndHardcap(t *testing.T) {
	catalog := permitsCandidateCatalog(t)
	rates, err := Rates(catalog, map[string]int64{"generator.beige_tower": 0, "generator.legal_dept": 2}, nil)
	if err != nil || len(rates["company.permits"]) != 1 || rates["company.permits"][0].String() != "2e-3" {
		t.Fatalf("neutral rates=%+v err=%v", rates, err)
	}
	rates, err = Rates(catalog, map[string]int64{"generator.beige_tower": 0, "generator.legal_dept": 2}, []multiplier.Contribution{{
		Slot: multiplier.SlotCommons, SourceID: "commons.member", Target: "all", Factor: decimal.New(2, 0),
	}})
	if err != nil || len(rates["company.permits"]) != 1 || rates["company.permits"][0].String() != "4e-3" {
		t.Fatalf("multiplied rates=%+v err=%v", rates, err)
	}

	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{
		"company.cash": "0", "company.permits": "2.3999e1",
	})
	if err != nil {
		t.Fatal(err)
	}
	state := &save.State{
		Ledger:                ledger,
		GeneratorCounts:       map[string]int64{"generator.beige_tower": 0, "generator.legal_dept": 1},
		EvaluatedThrough:      engineCursor,
		ManualTokenMilli:      50_000,
		ManualTokenRefilledAt: engineCursor,
	}
	if _, err := Evaluate(state, catalog, engineCursor.Add(time.Second), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if got := state.Ledger.Snapshot()["company.permits"]; got != "2.4e1" {
		t.Fatalf("near-cap permits = %s", got)
	}

	offlineLedger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{
		"company.cash": "0", "company.permits": "0",
	})
	if err != nil {
		t.Fatal(err)
	}
	offline := &save.State{
		Ledger:                offlineLedger,
		GeneratorCounts:       map[string]int64{"generator.beige_tower": 0, "generator.legal_dept": 1},
		EvaluatedThrough:      engineCursor,
		ManualTokenMilli:      50_000,
		ManualTokenRefilledAt: engineCursor,
	}
	result, err := Evaluate(offline, catalog, engineCursor.Add(24*time.Hour), ModeOffline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.ProductionMS != 86_400_000 || offline.Ledger.Snapshot()["company.permits"] != "2.4e1" {
		t.Fatalf("offline result=%+v permits=%s", result, offline.Ledger.Snapshot()["company.permits"])
	}
}
