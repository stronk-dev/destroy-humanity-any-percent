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

	var hardcapVector struct {
		Version          int    `json:"version"`
		ResourceID       string `json:"resource_id"`
		Start            string `json:"start"`
		ElapsedMS        int64  `json:"elapsed_ms"`
		Hardcap          string `json:"hardcap"`
		HardcapReasonKey string `json:"hardcap_reason_key"`
		Expect           string `json:"expect"`
	}
	vectorBytes, err := os.ReadFile("../../testdata/permits-hardcap-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(vectorBytes, &hardcapVector); err != nil {
		t.Fatal(err)
	}
	if hardcapVector.Version != 1 || hardcapVector.ResourceID != "company.permits" || hardcapVector.Hardcap != "2.4e1" ||
		hardcapVector.HardcapReasonKey != "resource.company_permits.cap.phase0" {
		t.Fatalf("hardcap vector = %+v", hardcapVector)
	}
	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{
		"company.cash": "0", hardcapVector.ResourceID: hardcapVector.Start,
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
	if _, err := Evaluate(state, catalog, engineCursor.Add(time.Duration(hardcapVector.ElapsedMS)*time.Millisecond), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if got := state.Ledger.Snapshot()[hardcapVector.ResourceID]; got != hardcapVector.Expect {
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

func TestPermitGateDebitsBothRequirementsAtomically(t *testing.T) {
	catalog := permitsCandidateCatalog(t)
	routeBytes, err := os.ReadFile("../../balance/testdata/permits-t3-gate-candidate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(routeBytes)
	if err != nil {
		t.Fatal(err)
	}
	newState := func(permits string) *save.State {
		ledger, restoreErr := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{
			"company.cash": "1e12", "company.permits": permits,
		})
		if restoreErr != nil {
			t.Fatal(restoreErr)
		}
		return &save.State{
			Ledger: ledger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0, "generator.legal_dept": 0},
			Tier: 3, RunSeq: 1, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
			LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{},
			EvaluatedThrough: engineCursor, ManualTokenMilli: 50_000, ManualTokenRefilledAt: engineCursor,
		}
	}
	revision := save.Revision{StreamID: "11111111-1111-4111-8111-111111111111", OwnerID: "22222222-2222-4222-8222-222222222222", Number: 1}
	request := IntentRequest{IntentID: "01989999-0001-7000-8000-000000000001", Kind: IntentCrossGate, GateID: "gate.t3_to_t4"}

	missing := newState("1.1999e1")
	decision, err := TransitionWithRoutes(request, missing, catalog, routeCatalog, revision, ModeOnline, engineCursor, nil, nil)
	if err != nil || decision.Outcome != save.IntentRejected || missing.Ledger.Snapshot()["company.cash"] != "1e12" ||
		missing.Ledger.Snapshot()["company.permits"] != "1.1999e1" || missing.GatesCrossed[request.GateID] {
		t.Fatalf("missing requirement decision=%s state=%+v err=%v", decision.Receipt, missing, err)
	}

	ready := newState("1.2e1")
	request.IntentID = "01989999-0002-7000-8000-000000000002"
	decision, err = TransitionWithRoutes(request, ready, catalog, routeCatalog, revision, ModeOnline, engineCursor, nil, nil)
	if err != nil || decision.Outcome != save.IntentApplied || ready.Ledger.Snapshot()["company.cash"] != "0" ||
		ready.Ledger.Snapshot()["company.permits"] != "0" || !ready.GatesCrossed[request.GateID] || ready.Tier != 4 {
		t.Fatalf("two-requirement decision=%s state=%+v err=%v", decision.Receipt, ready, err)
	}
}
