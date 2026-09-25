package gameui

import (
	"sort"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

const phaseAStandardGateID = "gate.t0_to_t1"

// standardGateForTier is the next adjacent gate the Desk may preview:
// gate.t0_to_t1 at Tier 0 and, once the pinned routes declare it (Tier 2
// content, rfc/tier2-content.md §E2), gate.t1_to_t2 at Tier 1. Every other gate
// and route stays fail-closed.
var standardGateForTier = map[int64]string{0: phaseAStandardGateID, 1: "gate.t1_to_t2"}

type transitionPreview struct {
	CrossGate   *crossGatePreview
	WindDown    bool
	Incorporate []string
}

type crossGatePreview struct {
	Eligible bool
	GateID   string
}

// previewPhaseATransitions projects only the Phase-A presentation contract.
// It invokes the existing production transition on a decoded state copy; the
// real intent receipt remains authoritative and no kernel semantic changes.
//
// minigameActive is the same read-only MA-C12 predicate Exit freezes: while a
// session is active|claimed, wind_down rejects with
// not_eligible/minigame_session_active, so the preview must not offer it. The
// previewed cross_gate is the ordinary transition: at Tier 0 the curriculum's
// scripted Exit cannot apply (it requires gate.t0_to_t1 already crossed). At
// Tier 1 a run-1 Company whose scripted first failure is due routes every
// command into that Exit (AR-F3); the preview does not model that, and the
// command's terminal receipt/event remains authoritative, as for every control.
//
// Incorporate lists the pinned faction ids exactly when incorporate can apply
// (Tier >= 2, no faction yet); otherwise it is nil and the wire omits it.
func previewPhaseATransitions(bundle production.CatalogBundle, company, founder *save.State, revision save.Revision,
	now time.Time, contributions []multiplier.Contribution, minigameActive bool) (transitionPreview, error) {
	if bundle.Economy == nil || bundle.Routes == nil || company == nil || company.Ledger == nil ||
		company.Ledger.Scope() != economy.ScopeCompany || founder == nil || founder.Ledger == nil ||
		founder.Ledger.Scope() != economy.ScopeFounder || revision.OwnerID == "" || revision.Number < 1 ||
		revision.ConstantsHash != bundle.ConstantsHash || now.IsZero() {
		return transitionPreview{}, production.ErrInvalidEngineState
	}
	preview := transitionPreview{WindDown: company.Tier >= 1 && !minigameActive}
	if company.Tier >= 2 && company.FactionID == "" && bundle.Faction != nil {
		preview.Incorporate = make([]string, 0, len(bundle.Faction.Factions))
		for _, row := range bundle.Faction.Factions {
			preview.Incorporate = append(preview.Incorporate, row.ID)
		}
		sort.Strings(preview.Incorporate)
	}
	gateID, standard := standardGateForTier[company.Tier]
	if _, declared := bundle.Routes.Gate(gateID); !standard || !declared || company.GatesCrossed[gateID] {
		return preview, nil
	}
	preview.CrossGate = &crossGatePreview{GateID: gateID}
	clone, err := cloneCompanyState(company, bundle.Economy)
	if err != nil {
		return transitionPreview{}, err
	}
	request := production.IntentRequest{
		IntentID:         "00000000-0000-7000-8000-000000000000",
		Kind:             production.IntentCrossGate,
		ExpectedRevision: revision.Number,
		GateID:           gateID,
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
