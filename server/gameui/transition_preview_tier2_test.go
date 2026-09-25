package gameui

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// tier2PreviewBundle pins the Tier 2 content candidates (rfc/tier2-content.md
// §A/§B) plus the live faction catalog; fixture-first, nothing is minted.
func tier2PreviewBundle(t *testing.T) production.CatalogBundle {
	t.Helper()
	read := func(path string) []byte {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	catalog, err := economy.LoadCatalog(read("../../balance/testdata/t2/economy-candidate-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(read("../../balance/testdata/t2/routes-candidate-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	factions, err := faction.LoadCatalog(read("../../balance/factions/phase0.json"), faction.CompactTitheBand{MinimumPPM: 50_000, DefaultPPM: 100_000, MaximumPPM: 150_000})
	if err != nil {
		t.Fatal(err)
	}
	return production.CatalogBundle{ConstantsHash: "sha256:" + strings.Repeat("b", 64), Economy: catalog, Routes: routeCatalog, Faction: factions}
}

// tier2PreviewCompany adds the provisioning remainder the candidate's
// managed-services -> garage-rack edge requires to the shared fixture state.
func tier2PreviewCompany(t *testing.T, bundle production.CatalogBundle) *save.State {
	t.Helper()
	company := candidateState(t, bundle.Economy)
	for _, generator := range bundle.Economy.GeneratorClassesForScope(economy.ScopeCompany) {
		if generator.Provision != nil {
			company.ProvisionRemaindersPPM[generator.Provision.GeneratorID] = 0
		}
	}
	return company
}

func setPreviewCash(t *testing.T, state *save.State, amount string) {
	t.Helper()
	current, _ := state.Ledger.Balance("company.cash")
	if _, err := state.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString(amount).Sub(current)}}}); err != nil {
		t.Fatal(err)
	}
}

func TestTier2PreviewProjectsTheT1ToT2GateFromTheProductionTransition(t *testing.T) {
	bundle := tier2PreviewBundle(t)
	now := time.UnixMilli(1_800_000_000_000).UTC()
	company := tier2PreviewCompany(t, bundle)
	company.ManualTokenRefilledAt, company.Tier = now, 1
	company.GatesCrossed["gate.t0_to_t1"] = true
	founder := transitionPreviewFounder(t, bundle.Economy, now)
	revision := save.Revision{OwnerID: "01985555-1111-7111-8111-111111111111", Number: 7, ConstantsHash: bundle.ConstantsHash}
	for _, row := range []struct {
		cash     string
		eligible bool
	}{{"9.99e6", false}, {"1e7", true}} {
		setPreviewCash(t, company, row.cash)
		preview, err := previewPhaseATransitions(bundle, company, founder, revision, now, nil, false)
		if err != nil || preview.CrossGate == nil || preview.CrossGate.GateID != "gate.t1_to_t2" || preview.CrossGate.Eligible != row.eligible || preview.Incorporate != nil {
			t.Fatalf("cash %s preview=%+v err=%v", row.cash, preview, err)
		}
		clone, err := cloneCompanyState(company, bundle.Economy)
		if err != nil {
			t.Fatal(err)
		}
		decision, err := production.TransitionWithRoutes(production.IntentRequest{IntentID: "00000000-0000-7000-8000-000000000001",
			Kind: production.IntentCrossGate, ExpectedRevision: revision.Number, GateID: "gate.t1_to_t2"},
			clone, bundle.Economy, bundle.Routes, revision, production.ModeOnline, now, nil, nil)
		if err != nil || (decision.Outcome == save.IntentApplied) != row.eligible {
			t.Fatalf("cash %s production outcome=%s err=%v", row.cash, decision.Outcome, err)
		}
	}
	company.GatesCrossed["gate.t1_to_t2"] = true
	if crossed, err := previewPhaseATransitions(bundle, company, founder, revision, now, nil, false); err != nil || crossed.CrossGate != nil {
		t.Fatalf("crossed preview=%+v err=%v", crossed, err)
	}
}

func TestTier2PreviewOffersIncorporateOnlyWhenItCanApply(t *testing.T) {
	bundle := tier2PreviewBundle(t)
	now := time.UnixMilli(1_800_000_000_000).UTC()
	company := tier2PreviewCompany(t, bundle)
	company.ManualTokenRefilledAt = now
	founder := transitionPreviewFounder(t, bundle.Economy, now)
	revision := save.Revision{OwnerID: "01985555-1111-7111-8111-111111111111", Number: 7, ConstantsHash: bundle.ConstantsHash}
	want := []string{}
	for _, row := range bundle.Faction.Factions {
		want = append(want, row.ID)
	}
	if len(want) == 0 {
		t.Fatal("live faction catalog is empty")
	}
	for _, row := range []struct {
		tier    int64
		faction string
		offered bool
	}{{1, "", false}, {2, "", true}, {3, "", true}, {2, want[0], false}} {
		company.Tier, company.FactionID = row.tier, row.faction
		preview, err := previewPhaseATransitions(bundle, company, founder, revision, now, nil, false)
		if err != nil || (preview.Incorporate != nil) != row.offered {
			t.Fatalf("tier %d faction %q preview=%+v err=%v", row.tier, row.faction, preview, err)
		}
		if row.offered && len(preview.Incorporate) != len(want) {
			t.Fatalf("offered factions %v, want every pinned faction %v", preview.Incorporate, want)
		}
	}
}

func TestTier2IncorporateWireIsOmittedBelowTierTwo(t *testing.T) {
	absent, err := json.Marshal(transitionRows{WindDown: eligibilityTransition{Eligible: true}})
	if err != nil || strings.Contains(string(absent), "incorporate") {
		t.Fatalf("absent control wire=%s err=%v", absent, err)
	}
	present, err := json.Marshal(transitionRows{Incorporate: &incorporateTransition{Factions: []incorporateFaction{{CopyKey: "incorporate.open_source", FactionID: "open_source"}}}})
	var decoded map[string]any
	if err != nil || json.Unmarshal(present, &decoded) != nil || !reflect.DeepEqual(decoded["incorporate"], map[string]any{"factions": []any{map[string]any{"copy_key": "incorporate.open_source", "faction_id": "open_source"}}}) {
		t.Fatalf("present control wire=%s err=%v", present, err)
	}
}
