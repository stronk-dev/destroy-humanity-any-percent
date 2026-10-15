package production

import (
	"bytes"
	"reflect"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

// Candidate modern-endpoint regimes only. Legacy fixed-external waits and the
// content diagnostic have different policies; this is not a production factory.
func calculationEndpointStartsResearch(t *testing.T, bundle CatalogBundle, origin *save.State, mode EvaluationMode, horizonMS int64) []int64 {
	t.Helper()
	attended, err := prestigecore.AttendedMS(origin, origin.EvaluatedThrough)
	if err != nil {
		t.Fatal(err)
	}
	starts := []int64{0}
	add := func(value int64) {
		if value > 0 && value <= horizonMS {
			starts = append(starts, value)
		}
	}
	ceiling := bundle.Prestige.CatchupCeilingMS
	if ceiling < horizonMS {
		add(ceiling + 1)
	}
	capMS := bundle.Economy.OfflinePolicy().AccrualCapMS
	if mode == ModeOnline && capMS < horizonMS {
		add(capMS + 1)
	}
	for _, buff := range origin.ActiveBuffs {
		// Before the absence threshold attendance advances with wall time;
		// beyond it the original attended coordinate is frozen, not continued.
		offset := buff.ExpiresAttendedMS - attended
		if offset <= ceiling {
			add(offset)
		}
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i] < starts[j] })
	unique := starts[:0]
	for _, value := range starts {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}

func calculationRegimeSearchResearch(qualifies func(int64) bool, starts []int64, horizonMS int64) (int64, bool) {
	for index, from := range starts {
		to := horizonMS
		if index+1 < len(starts) {
			to = starts[index+1] - 1
		}
		if at, found := calculationMonotoneSearchResearch(qualifies, from, to); found {
			return at, true
		}
	}
	return 0, false
}

// A reference on observed endpoint values, not the production harness search.
// This is valid only on an interval whose predicate is already known monotone.
func calculationMonotoneSearchResearch(affordable func(int64) bool, from, to int64) (int64, bool) {
	if affordable(from) {
		return from, true
	}
	if from == to || !affordable(to) {
		return 0, false
	}
	low, high := from+1, to
	for low < high {
		middle := low + (high-low)/2
		if affordable(middle) {
			high = middle
		} else {
			low = middle + 1
		}
	}
	return low, true
}

func TestAxisCalculationTemporalSearchResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	const horizon = int64(1501)
	const founderID = "01986666-2000-7000-8000-000000000001"
	cells, evaluations, missed, plainControls, unreachable := 0, 0, 0, 0, 0
	for _, profile := range []struct {
		name                 string
		production, building bool
		increment            string
		cuts                 []int64
		wantFound            bool
	}{
		{"production", true, false, "4e0", []int64{1000}, true},
		{"building", false, true, "1.8e0", []int64{1500}, true},
		{"combined", true, true, "4e0", []int64{1000, 1500}, true},
		{"plain_reachable", false, false, "1e0", nil, true},
		{"plain_already_affordable", false, false, "0", nil, true},
		{"plain_unreachable", false, false, "4e0", nil, false},
	} {
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			t.Run(profile.name+"/"+string(mode), func(t *testing.T) {
				initial := axisTimingState(t, bundle, start)
				initial.RunStartedAt = start.Add(-4000 * time.Millisecond)
				if profile.production {
					effect, ok := bundle.Opportunities.Effect("active.production")
					if !ok || effect.DurationMS != 5000 || effect.Factor.String() != "7e0" {
						t.Fatal("pinned production effect changed")
					}
					initial.ActiveBuffs = append(initial.ActiveBuffs, save.ActiveBuff{
						BuffInstanceID: "01985555-0000-7000-8000-000000000001", EffectRowID: effect.ID,
						ActivatedAttendedMS: 0, ExpiresAttendedMS: 5000})
				}
				if profile.building {
					effect, ok := bundle.Opportunities.Effect("active.building")
					if !ok || effect.DurationMS != 2000 || effect.PerOwnedPPM != 100000 {
						t.Fatal("pinned building effect changed")
					}
					target := "generator.beige_tower"
					initial.ActiveBuffs = append(initial.ActiveBuffs, save.ActiveBuff{
						BuffInstanceID: "01985555-0000-7000-8000-000000000002", EffectRowID: effect.ID, SelectedTarget: &target,
						ActivatedAttendedMS: 3500, ExpiresAttendedMS: 5500})
				}
				if err := bundle.ValidateFoundationState(initial); err != nil {
					t.Fatal(err)
				}
				origin := mustEncodeState(t, initial)
				efficiency := decimal.One
				if mode == ModeOffline {
					efficiency = mustDecimal(t, "9e-1")
				}
				goal := mustDecimal(t, "1e4").Add(mustDecimal(t, profile.increment).Mul(efficiency))
				cash := make([]decimal.Decimal, horizon+1)
				var earliest int64
				found := false
				for atMS := int64(0); atMS <= horizon; atMS++ {
					state := activationRestore(t, bundle, origin, 19, start)
					at := start.Add(time.Duration(atMS) * time.Millisecond)
					schedule, err := resolveActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, founderID, at)
					if err != nil || schedule.AttendedNowMS != 4000+atMS {
						t.Fatalf("endpoint%d schedule: %v", atMS, err)
					}
					if _, err := applyActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, founderID, at, schedule); err != nil {
						t.Fatal(err)
					}
					active, err := activePlayContributions(state, bundle.Opportunities, schedule.AttendedNowMS)
					if err != nil {
						t.Fatal(err)
					}
					contributions, err := assembleContributions(state, bundle.Economy, active)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := Evaluate(state, bundle.Economy, at, mode, contributions); err != nil {
						t.Fatal(err)
					}
					evaluations++
					factor := decimal.One
					if profile.production && atMS < 1000 {
						factor = factor.Mul(mustDecimal(t, "7e0"))
					}
					if profile.building && atMS < 1500 {
						factor = factor.Mul(mustDecimal(t, "1.1e0"))
					}
					want := mustDecimal(t, "1e4").Add(mustDecimal(t, "1.15115e0").Mul(factor).
						Mul(decimal.FromFloat64(float64(atMS))).Div(mustDecimal(t, "1e3")).Mul(efficiency)).Quantize(decimal.CanonicalSignificantDigits)
					cash[atMS] = mustDecimal(t, state.Ledger.Snapshot()["company.cash"])
					if !cash[atMS].Eq(want) {
						t.Fatalf("endpoint%d cash=%s want=%s", atMS, cash[atMS], want)
					}
					if !found && cash[atMS].Gte(goal) {
						earliest, found = atMS, true
					}
				}
				if found != profile.wantFound {
					t.Fatalf("population reachable=%t want=%t", found, profile.wantFound)
				}
				predicate := func(atMS int64) bool { return cash[atMS].Gte(goal) }
				unsplitAt, unsplitFound := calculationMonotoneSearchResearch(predicate, 0, horizon)
				if profile.production || profile.building {
					if unsplitFound {
						t.Fatalf("unsplit negative control unexpectedly found%d", unsplitAt)
					}
					missed++
				} else {
					if unsplitFound != found || found && unsplitAt != earliest {
						t.Fatal("plain monotone/already-affordable control failed")
					}
					if found {
						plainControls++
					} else {
						unreachable++
					}
				}
				candidateAt, candidateFound := int64(0), false
				starts := calculationEndpointStartsResearch(t, bundle, initial, mode, horizon)
				if !reflect.DeepEqual(starts, append([]int64{0}, profile.cuts...)) {
					t.Fatalf("derived starts%v differ from independent fixture boundaries%v", starts, profile.cuts)
				}
				for index, from := range starts {
					end := horizon
					if index+1 < len(starts) {
						end = starts[index+1] - 1
					}
					for atMS := from + 1; atMS <= end; atMS++ {
						if cash[atMS].Lt(cash[atMS-1]) {
							t.Fatalf("regime [%d,%d] is not monotone at%d", from, end, atMS)
						}
					}
				}
				candidateAt, candidateFound = calculationRegimeSearchResearch(predicate, starts, horizon)
				if candidateFound != found || found && candidateAt != earliest || !bytes.Equal(origin, mustEncodeState(t, initial)) {
					t.Fatal("boundary-aware result differs from exhaustive earliest endpoint or mutates origin")
				}
				cells++
				t.Logf("reachable=%t earliest%dms; unsplit found=%t; regime search agrees with1502 actual endpoints", found, earliest, unsplitFound)
			})
		}
	}
	if cells != 12 || evaluations != 18024 || missed != 6 || plainControls != 4 || unreachable != 2 {
		t.Fatalf("incomplete search census: cells%d evaluations%d misses%d controls%d unreachable%d", cells, evaluations, missed, plainControls, unreachable)
	}
}
