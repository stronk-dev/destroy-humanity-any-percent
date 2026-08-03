package faction

import (
	"encoding/json"
	"errors"
	"math/big"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

var ErrInvalidStockState = errors.New("invalid faction stock state")

type CatalogResolver interface {
	ResolveFaction(constantsHash string) (*Catalog, bool)
}

type CatchupPolicyResolver interface {
	ResolveCatchupCeilingMS(constantsHash string) (int64, bool)
}

type AccrualHook struct {
	Catalogs CatalogResolver
	Policies CatchupPolicyResolver
}

func (hook AccrualHook) AfterAccrual(state *save.State, economyCatalog *economy.Catalog, revision save.Revision, result accrualhook.Result, _ []multiplier.Contribution) ([]save.EventWrite, error) {
	if state == nil || hook.Catalogs == nil || hook.Policies == nil || result.ElapsedMS <= 0 {
		return nil, ErrInvalidStockState
	}
	if state.FactionID == "" {
		return nil, nil
	}
	catalog, ok := hook.Catalogs.ResolveFaction(revision.ConstantsHash)
	if !ok {
		return nil, ErrInvalidStockState
	}
	catchupCeilingMS, ok := hook.Policies.ResolveCatchupCeilingMS(revision.ConstantsHash)
	if !ok || catchupCeilingMS <= 0 {
		return nil, ErrInvalidStockState
	}
	faction, ok := catalog.Faction(state.FactionID)
	if !ok || state.StockUnits < 0 || state.StockUnits > catalog.StockCap || state.StockProgressMS < 0 || state.StockProgressMS >= catalog.StockIntervalMS ||
		state.ConsumedStockUnits < 0 || state.ConsumedStockUnits > decimal.MaxExactInteger {
		return nil, ErrInvalidStockState
	}
	// P6 defines attended spans by the catch-up ceiling. Offline catch-up still
	// advances the production cursor, but never enters the stock accumulator.
	if result.ElapsedMS > catchupCeilingMS {
		return nil, nil
	}
	factorPPM := big.NewInt(1_000_000)
	var generators []economy.GeneratorClassDefinition
	if economyCatalog != nil {
		generators = economyCatalog.GeneratorClassesForScope(economy.ScopeCompany)
	}
	for _, generator := range generators {
		count, exists := state.GeneratorCounts[generator.ID]
		if !exists || count < 0 || count > decimal.MaxExactInteger {
			return nil, ErrInvalidStockState
		}
		for _, role := range generator.Roles {
			if role.Kind == economy.RoleStockRate {
				factorPPM.Add(factorPPM, new(big.Int).Mul(big.NewInt(count), big.NewInt(role.PerPurchasedPPM)))
			}
		}
	}
	if state.StockRateRemainderPPM < 0 || state.StockRateRemainderPPM >= 1_000_000 {
		return nil, ErrInvalidStockState
	}
	numerator := new(big.Int).Mul(big.NewInt(result.ElapsedMS), factorPPM)
	numerator.Add(numerator, big.NewInt(state.StockRateRemainderPPM))
	scaledMS, remainder := new(big.Int), new(big.Int)
	scaledMS.QuoRem(numerator, big.NewInt(1_000_000), remainder)
	if !scaledMS.IsInt64() || scaledMS.Int64() > decimal.MaxExactInteger-state.StockProgressMS {
		return nil, ErrInvalidStockState
	}
	state.StockRateRemainderPPM = remainder.Int64()
	total := state.StockProgressMS + scaledMS.Int64()
	earned := total / catalog.StockIntervalMS
	state.StockProgressMS = total % catalog.StockIntervalMS
	if earned == 0 || state.StockUnits == catalog.StockCap {
		return nil, nil
	}
	before := state.StockUnits
	remaining := catalog.StockCap - state.StockUnits
	if earned > remaining {
		earned = remaining
	}
	state.StockUnits += earned
	if before != catalog.StockCap && state.StockUnits == catalog.StockCap {
		payload, _ := json.Marshal(map[string]any{
			"faction_id": faction.ID, "stock_resource": faction.Produces, "stock_cap": catalog.StockCap,
		})
		return []save.EventWrite{{Kind: save.EventFactionStockSaturated, SchemaVersion: 1, Payload: payload}}, nil
	}
	return nil, nil
}
