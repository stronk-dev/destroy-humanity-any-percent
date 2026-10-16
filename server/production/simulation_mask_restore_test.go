package production

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

type maskRestoreProbe struct {
	t           *testing.T
	id          string
	wantPresent bool
	err         error
}

func (hook maskRestoreProbe) AfterAccrual(state *save.State, _ *economy.Catalog, _ save.Revision, _ accrualhook.Result, _ []multiplier.Contribution) ([]save.EventWrite, error) {
	value, present := state.GeneratorCounts[hook.id]
	if value != 0 || present != hook.wantPresent {
		hook.t.Errorf("mask changed count membership before the hook: value%d present%t want%t", value, present, hook.wantPresent)
	}
	// Restoration must undo masked-count writes, but retain real hook effects
	// and forward both returned events and errors without rewriting them.
	state.GeneratorCounts[hook.id] = 99
	state.LifetimeValue = decimal.New(2, 0)
	return []save.EventWrite{{Kind: save.EventGuildActivityEvaluated, SchemaVersion: 1, Payload: []byte(`{"probe":true}`)}}, hook.err
}

func TestSimulationHookRestoresMaskedCountMembership(t *testing.T) {
	const id = "generator.mask_probe"
	cause := errors.New("hook probe failure")
	for _, fail := range []bool{false, true} {
		for _, row := range []struct {
			name    string
			present bool
			value   int64
		}{{"absent", false, 0}, {"present_zero", true, 0}, {"present_nonzero", true, 7}} {
			t.Run(row.name+map[bool]string{false: "/success", true: "/error"}[fail], func(t *testing.T) {
				counts := map[string]int64{"generator.other": 3}
				if row.present {
					counts[id] = row.value
				}
				before := cloneInt64Counts(counts)
				state := &save.State{GeneratorCounts: counts, LifetimeValue: decimal.One}
				var wantErr error
				if fail {
					wantErr = cause
				}
				events, err := runSimulationHookWithMasks(maskRestoreProbe{t, id, row.present, wantErr}, state, nil, save.Revision{}, accrualhook.Result{}, nil, map[string]bool{id: true})
				if !errors.Is(err, wantErr) || !reflect.DeepEqual(state.GeneratorCounts, before) {
					t.Errorf("masked counts not restored exactly: got%v want%v error%v", state.GeneratorCounts, before, err)
				}
				if !state.LifetimeValue.Eq(decimal.New(2, 0)) || len(events) != 1 || events[0].Kind != save.EventGuildActivityEvaluated || string(events[0].Payload) != `{"probe":true}` {
					t.Fatal("count restoration erased a real hook effect or changed forwarded events")
				}
			})
		}
	}
}

// The catalog permits multiple resource scopes, and simulation masks accept a
// known catalog ID. A Company must not acquire an absent Founder count merely
// because its inert output was masked. No authored balance data is changed.
func crossScopeMaskCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	return foundationCatalogWithMutation(t, func(root map[string]any) {
		root["resources"] = append(root["resources"].([]any), map[string]any{
			"id": "founder.mask_probe", "scope": "founder", "numeric_kind": "decimal", "initial": "0", "minimum": "0",
			"hardcap": map[string]any{"amount": "1e1000", "reason_key": "resource.mask_probe.cap"},
		})
		root["generator_classes"] = append(root["generator_classes"].([]any), map[string]any{
			"id": "generator.founder_probe", "tier": 0, "category": "category.fixture",
			"price":      map[string]any{"resource_id": "founder.mask_probe", "base": "1e0", "curve": map[string]any{"kind": "geometric", "ratio": "1.13e0"}},
			"production": map[string]any{"resource_id": "founder.mask_probe", "base_rate": "1e0"},
			"ladder":     []any{}, "roles": []any{map[string]any{"kind": "stock_rate", "per_purchased_ppm": 1000}},
		})
	})
}

func TestSimulationAdvancePreservesCompanyShapeWithCrossScopeMask(t *testing.T) {
	catalog := crossScopeMaskCatalog(t)
	factions := foundationFactionCatalog(t)
	start := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	mask := AblationMask{GeneratorIDs: []string{"generator.founder_probe"}}
	if _, err := simulationPolicyFor(catalog, mask); err != nil {
		t.Fatal("known catalog mask was not admitted", err)
	}
	for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
		for _, fail := range []bool{false, true} {
			t.Run(string(mode)+map[bool]string{false: "/success", true: "/error"}[fail], func(t *testing.T) {
				initial := foundationState(t, catalog, start)
				initial.GeneratorCounts["generator.low"], initial.GeneratorPurchasedTotal = 1, 1
				initial.FactionID, initial.IncorporatedAt = "bootstrapper", start
				setCash(t, initial, "1e4")
				origin := mustEncodeState(t, initial)
				restore := func() *save.State {
					t.Helper()
					state, err := save.RestoreState(origin, save.VersionForState(initial), catalog, economy.ScopeCompany, time.Time{})
					if err != nil {
						t.Fatal("Company codec admission failed", err)
					}
					if _, exists := state.GeneratorCounts["generator.founder_probe"]; exists {
						t.Fatal("scoped origin already contains the inert mask ID")
					}
					return state
				}
				policies := foundationCatchupPolicies{foundationConstantsHash: 120_000}
				if fail {
					policies = foundationCatchupPolicies{}
				}
				hook := faction.AccrualHook{Catalogs: faction.CatalogSet{foundationConstantsHash: factions}, Policies: policies}
				revision := save.Revision{Number: 1, ConstantsHash: foundationConstantsHash}
				control, actual := restore(), restore()
				want, controlErr := SimulateAdvance(control, catalog, SimulationDependencies{Hook: hook}, revision, mode, start.Add(time.Second), nil, AblationMask{})
				got, err := SimulateAdvance(actual, catalog, SimulationDependencies{Hook: hook}, revision, mode, start.Add(time.Second), nil, mask)
				if fail {
					if !errors.Is(controlErr, faction.ErrInvalidStockState) || !errors.Is(err, faction.ErrInvalidStockState) {
						t.Fatal("real hook failure changed", controlErr, err)
					}
				} else if controlErr != nil || err != nil || actual.StockProgressMS <= 0 {
					t.Fatal("real stock hook did not execute successfully", controlErr, err, actual.StockProgressMS)
				}
				if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(actual.GeneratorCounts, control.GeneratorCounts) {
					t.Errorf("inert cross-scope mask changed wait result or Company count shape: got%v want%v", actual.GeneratorCounts, control.GeneratorCounts)
				}
				encoded, encodeErr := save.EncodeState(actual)
				if encodeErr != nil {
					t.Errorf("real masked wait left an unencodable Company: %v", encodeErr)
				} else {
					if !bytes.Equal(encoded, mustEncodeState(t, control)) {
						t.Error("masked wait lost complete real hook/failed-scratch state")
					}
					if _, err := save.RestoreState(encoded, save.VersionForState(actual), catalog, economy.ScopeCompany, time.Time{}); err != nil {
						t.Error("masked wait cannot restore its actual Company version", err)
					}
				}
				unknown := restore()
				if _, err := SimulateAdvance(unknown, catalog, SimulationDependencies{Hook: hook}, revision, mode, start.Add(time.Second), nil, AblationMask{GeneratorIDs: []string{"generator.unknown_probe"}}); !errors.Is(err, ErrInvalidEngineState) || !bytes.Equal(origin, mustEncodeState(t, unknown)) {
					t.Fatal("unknown-mask preflight stopped refusing unchanged", err)
				}
				if !bytes.Equal(origin, mustEncodeState(t, initial)) {
					t.Fatal("test consumed the admitted origin")
				}
			})
		}
	}
}
