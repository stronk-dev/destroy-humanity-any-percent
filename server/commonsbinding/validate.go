// Package commonsbinding is the composition boundary between independently
// owned Commons and economy catalogs.
package commonsbinding

import (
	"fmt"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/economy"
)

func Validate(catalog *commons.Catalog, economyCatalog *economy.Catalog) error {
	if catalog == nil || economyCatalog == nil {
		return commons.ErrInvalidCatalog
	}
	source, ok := economyCatalog.MultiplierSource("commons.member")
	if !ok || source.Slot != economy.SlotCommons || source.Target != "all" || source.Provider != "commons" {
		return fmt.Errorf("%w: commons.member declaration", commons.ErrInvalidCatalog)
	}
	for _, weight := range catalog.SourceWeights {
		source, ok := economyCatalog.MultiplierSource(weight.SourceID)
		if !ok || source.Slot != weight.Slot || source.Slot == economy.SlotCommons {
			return fmt.Errorf("%w: source weight %q", commons.ErrInvalidCatalog, weight.SourceID)
		}
	}
	return nil
}
