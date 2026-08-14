package production

import (
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/save"
)

func TestT0T1CandidateRoleActivations(t *testing.T) {
	catalog := t0T1CandidateCatalog(t)
	rows := 0
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		for _, role := range generator.Roles {
			rows++
			targetID := candidateRoleTarget(role)
			t.Run(fmt.Sprintf("%s/%s/%s", generator.ID, role.Kind, targetID), func(t *testing.T) {
				controlState, control := runT0T1CandidateRole(t, catalog, generator, role, AblationMask{})
				maskedState, masked := runT0T1CandidateRole(t, catalog, generator, role,
					AblationMask{GeneratorIDs: []string{generator.ID}})
				want := RoleActivation{GeneratorID: generator.ID, Kind: role.Kind, TargetID: targetID}
				if !hasRoleActivation(control.RoleActivations, want) {
					t.Fatalf("candidate row did not activate: want=%+v got=%+v", want, control.RoleActivations)
				}
				if hasRoleActivation(masked.RoleActivations, want) {
					t.Fatalf("masked control retained activation: want=%+v got=%+v", want, masked.RoleActivations)
				}
				assertCandidateRoleEffectRemoved(t, role, controlState, maskedState)
			})
		}
	}
	if rows != 11 {
		t.Fatalf("candidate role row count=%d want=11", rows)
	}
}

func t0T1CandidateCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/t0-t1/economy-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func runT0T1CandidateRole(t *testing.T, catalog *economy.Catalog, generator economy.GeneratorClassDefinition,
	role economy.GeneratorRole, mask AblationMask) (*save.State, SimulationResult) {
	t.Helper()
	started := time.Date(2026, 8, 14, 10, 0, 0, 0, time.UTC)
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	counts, provisioned, remainders := map[string]int64{}, map[string]int64{}, map[string]int64{}
	for _, definition := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[definition.ID], provisioned[definition.ID] = 0, 0
		if definition.Provision != nil {
			remainders[definition.Provision.GeneratorID] = 0
		}
	}
	count := int64(1)
	switch role.Kind {
	case economy.RoleProvision:
		count = 10
	case economy.RoleStockRate:
		count = 1_000
	}
	counts[generator.ID] = count
	// Legal Department produces permits, so a Beige Tower supplies the cash-rate
	// target on which its institutional-knowledge pool can be exercised.
	if role.Kind == economy.RoleSynergyFeed && generator.Production != nil && generator.Production.ResourceID != "company.cash" {
		counts["generator.beige_tower"] = 1
	}
	state := &save.State{
		Ledger: ledger, GeneratorCounts: counts, UpgradesOwned: map[string]bool{}, GeneratorProvisioned: provisioned,
		ProvisionRemaindersPPM: remainders, EvaluatedThrough: started, RunStartedAt: started, RunSeq: 1,
		ManualTokenMilli: catalog.ManualPolicy().BucketCapMilli, ManualTokenRefilledAt: started,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{},
		MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
	}
	dependencies := SimulationDependencies{Routes: foundationRoutes(t)}
	if role.Kind == economy.RoleStockRate {
		factionCatalog := foundationFactionCatalog(t)
		dependencies.Hook = faction.AccrualHook{
			Catalogs: faction.CatalogSet{foundationConstantsHash: factionCatalog},
			Policies: foundationCatchupPolicies{foundationConstantsHash: 120_000},
		}
		state.FactionID = "bootstrapper"
		state.IncorporatedAt = started
	}
	request := IntentRequest{IntentID: "01990000-0814-7000-8000-000000000001", Kind: IntentPerformManualBatch,
		ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}
	result, err := SimulateTransition(request, state, catalog, dependencies,
		save.Revision{Number: 1, ConstantsHash: foundationConstantsHash}, ModeOnline, started.Add(time.Minute), nil, nil, mask)
	if err != nil || result.Decision.Outcome != save.IntentApplied {
		t.Fatalf("candidate role transition decision=%+v err=%v", result.Decision, err)
	}
	return state, result
}

func candidateRoleTarget(role economy.GeneratorRole) string {
	switch role.Kind {
	case economy.RoleProvision:
		return role.GeneratorID
	case economy.RoleSynergyFeed:
		return role.PoolID
	case economy.RoleManualOutput:
		return role.ActionID
	case economy.RoleStockRate:
		return "faction.stock"
	default:
		return ""
	}
}

func assertCandidateRoleEffectRemoved(t *testing.T, role economy.GeneratorRole, control, masked *save.State) {
	t.Helper()
	switch role.Kind {
	case economy.RoleProvision:
		if control.GeneratorProvisioned[role.GeneratorID] <= masked.GeneratorProvisioned[role.GeneratorID] {
			t.Fatalf("provision effect control=%d masked=%d", control.GeneratorProvisioned[role.GeneratorID], masked.GeneratorProvisioned[role.GeneratorID])
		}
	case economy.RoleStockRate:
		if control.StockProgressMS <= masked.StockProgressMS && control.StockUnits <= masked.StockUnits {
			t.Fatalf("stock effect control=%d/%d masked=%d/%d", control.StockUnits, control.StockProgressMS, masked.StockUnits, masked.StockProgressMS)
		}
	default:
		controlCash, _ := control.Ledger.Balance("company.cash")
		maskedCash, _ := masked.Ledger.Balance("company.cash")
		if !controlCash.Gt(maskedCash) {
			t.Fatalf("cash effect control=%s masked=%s", controlCash, maskedCash)
		}
	}
}
