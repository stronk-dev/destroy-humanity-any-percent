package production

import (
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/save"
)

func TestAxisSimulationStockRatePreservesCompanyVersion(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	wantRole := RoleActivation{GeneratorID: "generator.beige_tower", Kind: economy.RoleStockRate, TargetID: "faction.stock"}
	for _, tc := range []struct {
		name     string
		masked   bool
		withHook bool
		wantRole bool
		stock    int64
		progress int64
	}{
		{"real_stock_role", false, true, true, 1, 60},
		{"masked_role", true, true, false, 1, 0},
		{"no_hook", false, false, false, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := axisTimingState(t, bundle, start)
			state.FactionID, state.IncorporatedAt = "bootstrapper", start
			encoded := mustEncodeState(t, state)
			if _, err := save.RestoreState(encoded, 19, bundle.Economy, economy.ScopeCompany, time.Time{}); err != nil {
				t.Fatalf("population is not admitted Company19: %v", err)
			}
			dependencies := SimulationDependencies{Routes: bundle.Routes}
			if tc.withHook {
				dependencies.Hook = faction.AccrualHook{
					Catalogs: faction.CatalogSet{bundle.ConstantsHash: bundle.Faction},
					Policies: foundationCatchupPolicies{bundle.ConstantsHash: 120_000},
				}
			}
			mask := AblationMask{}
			if tc.masked {
				mask.GeneratorIDs = []string{"generator.beige_tower"}
			}
			result, err := SimulateAdvance(state, bundle.Economy, dependencies,
				save.Revision{Number: 1, ConstantsHash: bundle.ConstantsHash}, ModeOnline, start.Add(time.Minute), nil, mask)
			if err != nil {
				t.Fatalf("Company19 simulation refused: %v", err)
			}
			if hasRoleActivation(result.RoleActivations, wantRole) != tc.wantRole ||
				state.StockUnits != tc.stock || state.StockProgressMS != tc.progress {
				t.Fatalf("roles=%+v stock=%d progress=%d; want role=%t stock=%d progress=%d",
					result.RoleActivations, state.StockUnits, state.StockProgressMS, tc.wantRole, tc.stock, tc.progress)
			}
			if save.VersionForState(state) != 19 || state.AttainmentScoreRun != 6 || len(state.AchievementsAttainedRun) != 2 ||
				state.GeneratorCounts["generator.beige_tower"] != 1 || !state.EvaluatedThrough.Equal(start.Add(time.Minute)) {
				t.Fatal("simulation changed Company version/attainment/ownership or missed the requested cursor")
			}
			if _, err := save.RestoreState(mustEncodeState(t, state), 19, bundle.Economy, economy.ScopeCompany, time.Time{}); err != nil {
				t.Fatalf("simulation result does not restore as Company19: %v", err)
			}
		})
	}
}
