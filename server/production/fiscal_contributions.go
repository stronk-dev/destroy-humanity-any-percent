package production

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

func FrozenFiscalContributions(catalog *fiscal.Catalog, founder *save.State) ([]save.FrozenContribution, error) {
	if catalog == nil {
		return []save.FrozenContribution{}, nil
	}
	if founder == nil || validateFounderFiscalState(catalog, founder) != nil {
		return nil, ErrInvalidEngineState
	}
	result := make([]save.FrozenContribution, 0, len(catalog.GeneratorLevelRows())+1)
	hoard, err := catalog.HoardFactor(founder.FiscalCredit)
	if err != nil {
		return nil, err
	}
	result = append(result, save.FrozenContribution{SourceID: catalog.Hoard.SourceID, Slot: catalog.Hoard.Slot,
		Target: catalog.Hoard.Target, Factor: hoard.String()})
	for _, row := range catalog.GeneratorLevelRows() {
		factor, err := catalog.GeneratorLevelFactor(row.GeneratorID, founder.FiscalGeneratorLevels[row.GeneratorID])
		if err != nil {
			return nil, err
		}
		result = append(result, save.FrozenContribution{SourceID: row.SourceID, Slot: row.Slot,
			Target: row.GeneratorID, Factor: factor.String()})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].SourceID < result[right].SourceID })
	return result, nil
}

type FrozenContributionProvider struct{ DB *sql.DB }

// ResolveFrozenContributions validates immutable run contributions against
// their pinned economy catalog. Persisted and transaction-local projections
// share this conversion so bootstrap cannot grow a second interpretation.
func ResolveFrozenContributions(catalog *economy.Catalog, values []save.FrozenContribution) ([]multiplier.Contribution, error) {
	if catalog == nil {
		return nil, ErrInvalidEngineState
	}
	result := make([]multiplier.Contribution, len(values))
	seen := map[string]bool{}
	for index, value := range values {
		declaration, ok := catalog.MultiplierSource(value.SourceID)
		if !ok || seen[value.SourceID] || declaration.Provider != "fiscal" || multiplier.Slot(declaration.Slot) != value.Slot || declaration.Target != value.Target {
			return nil, fmt.Errorf("%w: frozen Fiscal contribution declaration", ErrInvalidEngineState)
		}
		factor, err := decimal.ParseCanonical(value.Factor)
		if err != nil || !factor.IsStateValue() || !factor.Gt(decimal.Zero) {
			return nil, ErrInvalidEngineState
		}
		seen[value.SourceID] = true
		result[index] = multiplier.Contribution{SourceID: value.SourceID, Slot: value.Slot, Target: value.Target, Factor: factor}
	}
	return result, nil
}

func (provider FrozenContributionProvider) Contributions(ctx context.Context, state *save.State, catalog *economy.Catalog,
	revision save.Revision) ([]multiplier.Contribution, error) {
	if provider.DB == nil || state == nil || catalog == nil || state.RunSeq < 1 {
		return nil, ErrInvalidEngineState
	}
	values, err := save.LoadRunFrozenContributions(ctx, provider.DB, revision.StreamID, state.RunSeq)
	if err != nil {
		return nil, err
	}
	return ResolveFrozenContributions(catalog, values)
}
