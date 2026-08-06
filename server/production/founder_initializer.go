package production

import (
	"time"

	"cloud-clicker/server/save"
)

type FounderInitializer struct{ Catalogs ReplayCatalogResolver }

func (initializer FounderInitializer) InitializeNewFounder(constantsHash string, now time.Time, founder, company *save.State) ([]save.FrozenContribution, error) {
	if initializer.Catalogs == nil || founder == nil || company == nil {
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
	return FrozenFiscalContributions(bundle.Fiscal, founder)
}
