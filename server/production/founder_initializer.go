package production

import (
	"time"

	"cloud-clicker/server/save"
)

type FounderInitializer struct{ Catalogs ReplayCatalogResolver }

func (initializer FounderInitializer) InitializeNewFounder(constantsHash, founderID string, now time.Time, founder, company *save.State) ([]save.FrozenContribution, error) {
	if initializer.Catalogs == nil || founderID == "" || founder == nil || company == nil {
		return nil, ErrInvalidEngineState
	}
	bundle, ok := initializer.Catalogs.ResolveReplayCatalogs(constantsHash)
	if !ok || !bundle.valid(constantsHash) {
		return nil, ErrInvalidEngineState
	}
	if bundle.foundationsActive() {
		if err := settleAndActivateFoundations(CatalogBundle{}, bundle, founder, company, company); err != nil {
			return nil, err
		}
	}
	if bundle.Opportunities != nil {
		if _, err := initializeActivePlayState(company, bundle.Opportunities, founderID); err != nil {
			return nil, err
		}
	}
	return FrozenFiscalContributions(bundle.Fiscal, founder)
}
