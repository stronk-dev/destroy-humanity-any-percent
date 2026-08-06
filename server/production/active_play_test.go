package production

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

func activePlayTestCatalog(t *testing.T) (*economy.Catalog, *activeplay.Catalog) {
	t.Helper()
	economyBytes, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var economyRoot map[string]any
	if json.Unmarshal(economyBytes, &economyRoot) != nil {
		t.Fatal("economy fixture")
	}
	sources := economyRoot["multiplier_sources"].([]any)
	sources = append(sources,
		map[string]any{"id": "active.building.generator.beige_tower", "slot": "event_buffs", "target": "generator.beige_tower", "provider": "active_play"},
		map[string]any{"id": "active.click", "slot": "event_buffs", "target": "manual.click", "provider": "active_play"},
		map[string]any{"id": "active.production", "slot": "event_buffs", "target": "all", "provider": "active_play"},
	)
	economyRoot["multiplier_sources"] = sources
	economyBytes, _ = json.Marshal(economyRoot)
	economyCatalog, err := economy.LoadCatalog(economyBytes)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("../../balance/testdata/active-play-foundation-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var root struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if json.Unmarshal(fixture, &root) != nil {
		t.Fatal("active fixture")
	}
	activeCatalog, err := activeplay.LoadCatalog(root.Baseline, economyCatalog)
	if err != nil {
		t.Fatal(err)
	}
	return economyCatalog, activeCatalog
}

func TestActivePlayScheduleUsesAttendedTimeAndReplayEvidence(t *testing.T) {
	_, catalog := activePlayTestCatalog(t)
	founder := "01985555-0000-7000-8000-000000000001"
	started := time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)
	state := &save.State{RunSeq: 1, RunStartedAt: started, EvaluatedThrough: started}
	initial, err := initializeActivePlayState(state, catalog, founder)
	if err != nil {
		t.Fatal(err)
	}
	policy := &prestigecore.Policy{CatchupCeilingMS: 5_000}
	// A one-hour wall gap is classified offline and advances no attended clock.
	offline, err := resolveActivePlaySchedule(state, catalog, policy, founder, started.Add(time.Hour))
	if err != nil || offline.AttendedNowMS != 0 || offline.Spawned != nil || offline.AfterNextOpportunityMS != initial.SpawnedAttendedMS {
		t.Fatalf("offline schedule=%+v err=%v", offline, err)
	}
	// A short, attended command at the exact coordinate materializes the row.
	now := started.Add(time.Duration(initial.SpawnedAttendedMS) * time.Millisecond)
	due, err := resolveActivePlaySchedule(state, catalog, policy, founder, now)
	if err != nil || due.Spawned == nil || due.Spawned.OpportunityID != initial.OpportunityID {
		t.Fatalf("due schedule=%+v err=%v", due, err)
	}
	replay := *state
	replay.ActiveBuffs = []save.ActiveBuff{}
	if _, err := applyActivePlaySchedule(&replay, catalog, policy, founder, now, due); err != nil {
		t.Fatal(err)
	}
	if replay.PendingOpportunity == nil || replay.PendingOpportunity.OpportunityID != initial.OpportunityID || replay.OpportunitySpawnSeq != 1 || replay.NextOpportunityAttendedMS != 0 {
		t.Fatalf("replayed state=%+v", replay)
	}
	tampered := due
	tampered.Spawned = spawnEvidence(activeplay.Spawn{Sequence: due.Spawned.Sequence, SampledIntervalMS: due.Spawned.SampledIntervalMS,
		EffectDraw: mustDifferentUint64(t, due.Spawned.EffectDraw), EffectRowID: due.Spawned.EffectRowID, OpportunityID: due.Spawned.OpportunityID,
		SpawnedAttendedMS: due.Spawned.SpawnedAttendedMS, ExpiresAttendedMS: due.Spawned.ExpiresAttendedMS})
	bad := *state
	bad.ActiveBuffs = []save.ActiveBuff{}
	if _, err := applyActivePlaySchedule(&bad, catalog, policy, founder, now, tampered); err == nil {
		t.Fatal("tampered integer draw accepted")
	}
}

func mustDifferentUint64(t *testing.T, value string) uint64 {
	t.Helper()
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	if parsed == ^uint64(0) {
		return parsed - 1
	}
	return parsed + 1
}

func TestActivePlayDynamicContributionUsesStaticDeclarationAndComboCap(t *testing.T) {
	economyCatalog, catalog := activePlayTestCatalog(t)
	state := &save.State{GeneratorCounts: map[string]int64{"generator.beige_tower": 2}, ActiveBuffs: []save.ActiveBuff{
		{BuffInstanceID: "01985555-0000-7000-8000-000000000001", EffectRowID: "active.production", ActivatedAttendedMS: 1, ExpiresAttendedMS: 100},
		{BuffInstanceID: "01985555-0000-7000-8000-000000000002", EffectRowID: "active.production", ActivatedAttendedMS: 2, ExpiresAttendedMS: 100},
	}}
	values, err := activePlayContributions(state, catalog, 50)
	if err != nil || len(values) != 2 {
		t.Fatalf("values=%+v err=%v", values, err)
	}
	if _, err := validateContributions(economyCatalog, values); err != nil {
		t.Fatal(err)
	}
	product := values[0].Factor.Mul(values[1].Factor).Quantize(12)
	if product.Gt(catalog.Combo.Cap) {
		t.Fatalf("uncapped product %s > %s", product, catalog.Combo.Cap)
	}
}

func TestActivePlayComboCapCoversAllAndGeneratorSpecificTargets(t *testing.T) {
	economyCatalog, catalog := activePlayTestCatalog(t)
	catalog.Combo.Cap = decimal.FromFloat64(10)
	target := "generator.beige_tower"
	state := &save.State{GeneratorCounts: map[string]int64{target: 100}, ActiveBuffs: []save.ActiveBuff{
		{BuffInstanceID: "01985555-0000-7000-8000-000000000001", EffectRowID: "active.production", ExpiresAttendedMS: 100},
		{BuffInstanceID: "01985555-0000-7000-8000-000000000002", EffectRowID: "active.building", SelectedTarget: &target, ExpiresAttendedMS: 100},
	}}
	values, saturated, err := activePlayContributionsWithClamp(state, catalog, 50)
	if err != nil || !saturated {
		t.Fatalf("values=%+v saturated=%v err=%v", values, saturated, err)
	}
	if _, err := validateContributions(economyCatalog, values); err != nil {
		t.Fatal(err)
	}
	product := decimal.One
	for _, value := range values {
		if value.Target == "all" || value.Target == target {
			product = product.Mul(value.Factor)
		}
	}
	product = product.Quantize(decimal.CanonicalSignificantDigits)
	if product.Gt(catalog.Combo.Cap) {
		t.Fatalf("cross-target slot product %s > %s", product, catalog.Combo.Cap)
	}
}

func TestActivePlayBundleCapsSchedulerToRepresentableOnlineHorizon(t *testing.T) {
	bundle := activePlayReplayBundle(t)
	policy := *bundle.Prestige
	policy.CatchupCeilingMS = bundle.Opportunities.Schedule.MinimumIntervalMS + bundle.Opportunities.Schedule.LifetimeMS
	bundle.Prestige = &policy
	if bundle.valid(bundle.ConstantsHash) {
		t.Fatal("bundle accepted an online horizon that can cross two pending-opportunity expiries")
	}
}

func TestActivePlayLuckyBoundaryVectors(t *testing.T) {
	fixture, err := os.ReadFile("../../balance/testdata/active-play-foundation-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var root struct {
		Vectors []struct {
			Name      string `json:"name"`
			Bank      string `json:"bank"`
			Rate      string `json:"rate"`
			Fraction  string `json:"fraction"`
			RateCap   string `json:"rate_cap"`
			Epsilon   string `json:"epsilon"`
			Requested string `json:"requested"`
		} `json:"lucky_vectors"`
	}
	if err := json.Unmarshal(fixture, &root); err != nil || len(root.Vectors) != 5 {
		t.Fatalf("vectors=%d err=%v", len(root.Vectors), err)
	}
	for _, vector := range root.Vectors {
		t.Run(vector.Name, func(t *testing.T) {
			values := make([]decimal.Decimal, 0, 5)
			for _, encoded := range []string{vector.Bank, vector.Rate, vector.Fraction, vector.RateCap, vector.Epsilon} {
				value, parseErr := decimal.ParseCanonical(encoded)
				if parseErr != nil {
					t.Fatal(parseErr)
				}
				values = append(values, value)
			}
			got, evaluateErr := luckyPayoutRequested(values[0], values[1], values[2], values[3], values[4])
			if evaluateErr != nil || got.String() != vector.Requested {
				t.Fatalf("got=%s want=%s err=%v", got, vector.Requested, evaluateErr)
			}
		})
	}
}
