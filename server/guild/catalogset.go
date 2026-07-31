package guild

type CatalogSet map[string]*Catalog

func (set CatalogSet) ResolveGuild(constantsHash string) (*Catalog, bool) {
	catalog, ok := set[constantsHash]
	return catalog, ok && catalog != nil
}
