package faction

type CatalogSet map[string]*Catalog

func (set CatalogSet) ResolveFaction(constantsHash string) (*Catalog, bool) {
	catalog, ok := set[constantsHash]
	return catalog, ok
}
