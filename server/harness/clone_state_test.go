package harness

import (
	"bytes"
	"strconv"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// A clone must retain its source encoding even when a legacy scenario consumer
// receives a newer admitted Company. This does not enable axis-stack scenarios.
func TestCloneStatePreservesSaveVersionThroughSearch(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range []int{14, 17} {
		t.Run(strconv.Itoa(version), func(t *testing.T) {
			state, err := suite.newRelevanceState()
			if err != nil {
				t.Fatal(err)
			}
			state.WireVersion = version
			state.GeneratorCounts["generator.alpha"] = 1
			state.GeneratorPurchasedTotal = 1
			wantAt := int64(500)
			if version == 17 {
				state.MeterValues = map[string]int{}
				state.MeterDecayRemainders = map[string]int64{}
				state.MeterInputRemainders = map[string]int64{}
				state.AchievementsEarnedRun = map[string]bool{}
				state.ComputeBurstRemainingMS = 1000
				wantAt = 250
			}
			origin, err := save.EncodeState(state)
			if err != nil {
				t.Fatal("probe state not encodable", err)
			}
			state, err = save.RestoreState(origin, version, suite.Catalog, economy.ScopeCompany, time.Time{})
			if err != nil {
				t.Fatal("probe origin not admitted", err)
			}
			clone, cloneErr := cloneState(suite.Catalog, state)
			if cloneErr != nil {
				t.Errorf("clone refused admitted version%d: %v", version, cloneErr)
			} else {
				encoded, err := save.EncodeState(clone)
				if err != nil || save.VersionForState(clone) != version || !bytes.Equal(encoded, origin) {
					t.Errorf("clone changed version or bytes: version%d -> %d, error%v", version, save.VersionForState(clone), err)
				}
				clone.GeneratorCounts["generator.alpha"] = 2
				if state.GeneratorCounts["generator.alpha"] != 1 {
					t.Fatal("clone aliases source counts")
				}
			}
			cost := decimal.FromString("1.01e2")
			at, affordable, searchErr := suite.earliestAffordable(state, 1, 0, 1000, "company.cash", cost, production.AblationMask{}, &relevanceCounter{limit: 100})
			if searchErr != nil || !affordable || at != wantAt {
				t.Errorf("actual search version%d: at%d affordable%t error%v; want%d", version, at, affordable, searchErr, wantAt)
			}
			unchanged, err := save.EncodeState(state)
			if err != nil || !bytes.Equal(unchanged, origin) {
				t.Fatal("search consumed source origin", err)
			}
			t.Logf("admitted version%d: clone_error=%v search_error=%v at=%d expected=%d", version, cloneErr, searchErr, at, wantAt)
		})
	}
}
