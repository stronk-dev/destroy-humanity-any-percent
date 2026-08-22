package gameui

import (
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

const phaseAStandardGateID = "gate.t0_to_t1"

type transitionPreview struct {
	CrossGate *crossGatePreview
	WindDown  bool
}

type crossGatePreview struct {
	Eligible bool
	GateID   string
}

// previewPhaseATransitions projects only the Phase-A presentation contract.
// It invokes the existing production transition on a decoded state copy; the
// real intent receipt remains authoritative and no kernel semantic changes.
func previewPhaseATransitions(bundle production.CatalogBundle, company, founder *save.State, revision save.Revision,
	now time.Time, contributions []multiplier.Contribution) (transitionPreview, error) {
	if bundle.Economy == nil || bundle.Routes == nil || company == nil || company.Ledger == nil ||
		company.Ledger.Scope() != economy.ScopeCompany || founder == nil || founder.Ledger == nil ||
		founder.Ledger.Scope() != economy.ScopeFounder || revision.OwnerID == "" || revision.Number < 1 ||
		revision.ConstantsHash != bundle.ConstantsHash || now.IsZero() {
		return transitionPreview{}, production.ErrInvalidEngineState
	}
	preview := transitionPreview{WindDown: company.Tier >= 1}
	if company.Tier != 0 || company.GatesCrossed[phaseAStandardGateID] {
		return preview, nil
	}
	preview.CrossGate = &crossGatePreview{GateID: phaseAStandardGateID}
	clone, err := cloneCompanyState(company, bundle.Economy)
	if err != nil {
		return transitionPreview{}, err
	}
	request := production.IntentRequest{
		IntentID:         "00000000-0000-7000-8000-000000000000",
		Kind:             production.IntentCrossGate,
		ExpectedRevision: revision.Number,
		GateID:           phaseAStandardGateID,
	}
	decision, err := production.TransitionWithRoutes(request, clone, bundle.Economy, bundle.Routes,
		revision, production.ModeOnline, now, contributions, nil)
	if err != nil {
		return transitionPreview{}, err
	}
	preview.CrossGate.Eligible = decision.Outcome == save.IntentApplied
	return preview, nil
}

func cloneCompanyState(state *save.State, catalog *economy.Catalog) (*save.State, error) {
	encoded, err := save.EncodeState(state)
	if err != nil {
		return nil, err
	}
	cloned, err := save.RestoreState(encoded, save.VersionForState(state), catalog, economy.ScopeCompany, time.Time{})
	if err != nil {
		return nil, err
	}
	cloned.FactionStockResource = state.FactionStockResource
	return cloned, nil
}
