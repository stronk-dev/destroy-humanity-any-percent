package faction

import "cloud-clicker/server/save"

type CatalogSet map[string]*Catalog

func (set CatalogSet) ResolveFaction(constantsHash string) (*Catalog, bool) {
	catalog, ok := set[constantsHash]
	return catalog, ok
}

func (set CatalogSet) ValidateState(constantsHash string, state *save.State) error {
	catalog, ok := set.ResolveFaction(constantsHash)
	if !ok {
		return ErrInvalidStockState
	}
	return catalog.ValidateState(state)
}
