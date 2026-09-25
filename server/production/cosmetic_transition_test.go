package production

import (
	"errors"
	"strings"
	"testing"

	"cloud-clicker/server/cosmetic"
	"cloud-clicker/server/save"
)

// AC2 at the epoch boundary: a next bundle that drops a pinned cosmetic id,
// or drops the artifact, cannot settle an Exit.
func TestSettleRejectsCosmeticIDDrop(t *testing.T) {
	load := func(data string) *cosmetic.Catalog {
		catalog, err := cosmetic.Load([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		return catalog
	}
	current := CatalogBundle{Cosmetics: load(`{"schema_version":1,"items":[{"cosmetic_id":"horse_armor","slot":"pet","unlock":{"kind":"active_company_tier_at_least","tier":1}}]}`)}
	dropped := CatalogBundle{Cosmetics: load(`{"schema_version":1,"items":[{"cosmetic_id":"zebra_armor","slot":"pet","unlock":{"kind":"active_company_tier_at_least","tier":1}}]}`)}
	for name, next := range map[string]CatalogBundle{"dropped id": dropped, "dropped artifact": {}} {
		err := settleAndActivateFoundations(current, next, &save.State{}, &save.State{}, &save.State{})
		if !errors.Is(err, ErrInvalidEngineState) || !strings.Contains(err.Error(), "cosmetic") {
			t.Fatalf("%s: settle accepted or wrong error: %v", name, err)
		}
	}
}
