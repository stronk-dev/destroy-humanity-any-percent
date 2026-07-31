package guild

import (
	"errors"
	"fmt"
	"testing"
)

const phase0Catalog = `{"schema_version":1,"guild_tithe_ppm":20000,"clearing_rate_ppm":500000,"npc_exchange_ppm":250000,"stock_intake_cap":120,"consumption_bonus_ppm_per_unit":0,"max_members":50,"min_members":2,"grace_days":7,"guild_xp_target_per_founder":250000,"clearing_interval_ms":300000}`

func TestLoadCatalogPhase0Literal(t *testing.T) {
	catalog, err := LoadCatalog([]byte(phase0Catalog))
	if err != nil {
		t.Fatal(err)
	}
	if catalog.GuildTithePPM != 20_000 || catalog.ClearingRatePPM != 500_000 ||
		catalog.NPCExchangePPM != 250_000 || catalog.StockIntakeCap != 120 ||
		catalog.ConsumptionBonusPPMPerUnit != 0 || catalog.MaxMembers != 50 ||
		catalog.MinMembers != 2 || catalog.GraceDays != 7 {
		t.Fatalf("catalog=%+v", catalog)
	}
}

func TestLoadCatalogRejectsEveryLiteralDriftAndUnknownShape(t *testing.T) {
	tests := []string{
		`{"schema_version":2,"guild_tithe_ppm":20000,"clearing_rate_ppm":500000,"npc_exchange_ppm":250000,"stock_intake_cap":120,"consumption_bonus_ppm_per_unit":0,"max_members":50,"min_members":2,"grace_days":7}`,
		`{"schema_version":1,"guild_tithe_ppm":20001,"clearing_rate_ppm":500000,"npc_exchange_ppm":250000,"stock_intake_cap":120,"consumption_bonus_ppm_per_unit":0,"max_members":50,"min_members":2,"grace_days":7}`,
		`{"schema_version":1,"guild_tithe_ppm":20000,"clearing_rate_ppm":500000,"npc_exchange_ppm":500000,"stock_intake_cap":120,"consumption_bonus_ppm_per_unit":0,"max_members":50,"min_members":2,"grace_days":7}`,
		`{"schema_version":1,"guild_tithe_ppm":20000,"clearing_rate_ppm":500000,"npc_exchange_ppm":250000,"stock_intake_cap":120,"consumption_bonus_ppm_per_unit":1,"max_members":50,"min_members":2,"grace_days":7}`,
		`{"schema_version":1,"guild_tithe_ppm":20000,"clearing_rate_ppm":500000,"npc_exchange_ppm":250000,"stock_intake_cap":120,"consumption_bonus_ppm_per_unit":0,"max_members":50,"min_members":2,"grace_days":7,"extra":true}`,
	}
	for index, data := range tests {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			if _, err := LoadCatalog([]byte(data)); !errors.Is(err, ErrInvalidCatalog) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	if _, err := LoadCatalog([]byte(phase0Catalog + `{}`)); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("trailing err=%v", err)
	}
}
