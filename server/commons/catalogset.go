package commons

// CatalogSet binds Commons policy to the same constants hash as the economy
// catalog without importing any consumer package.
type CatalogSet map[string]*Catalog

func (set CatalogSet) ResolveCommons(constantsHash string) (*Catalog, bool) {
	catalog, ok := set[constantsHash]
	return catalog, ok && catalog != nil
}
func (set CatalogSet) CompactTitheBand(constantsHash string) (int64, int64, bool) {
	catalog, ok := set.ResolveCommons(constantsHash)
	if !ok {
		return 0, 0, false
	}
	return catalog.MinimumTithePPM, catalog.MaximumTithePPM, true
}
func (set CatalogSet) CompactCohortTarget(constantsHash string) (int, bool) {
	catalog, ok := set.ResolveCommons(constantsHash)
	if !ok {
		return 0, false
	}
	return catalog.CohortTargetSize, true
}

func (set CatalogSet) GuildHealthWindowMS(constantsHash string) (int64, bool) {
	catalog, ok := set.ResolveCommons(constantsHash)
	if !ok {
		return 0, false
	}
	return catalog.SolidarityWindowMS, true
}
