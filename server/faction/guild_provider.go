package faction

import (
	"context"
	"errors"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

type GuildMembershipResolver interface {
	FounderGuildMember(context.Context, string) (bool, error)
}

type StockConsumptionProvider struct {
	Catalogs guild.CatalogResolver
	Members  GuildMembershipResolver
}

func (provider StockConsumptionProvider) Contributions(ctx context.Context, state *save.State, _ *economy.Catalog, revision save.Revision) ([]multiplier.Contribution, error) {
	if state == nil || provider.Catalogs == nil || provider.Members == nil || state.GuildConsumedWindow < 0 {
		return nil, errors.New("guild stock-consumption provider unavailable")
	}
	member, err := provider.Members.FounderGuildMember(ctx, revision.OwnerID)
	if err != nil || !member {
		return nil, err
	}
	catalog, ok := provider.Catalogs.ResolveGuild(revision.ConstantsHash)
	if !ok || catalog == nil {
		return nil, errors.New("guild catalog unavailable")
	}
	consumed := decimal.FromFloat64(float64(state.GuildConsumedWindow))
	rate := decimal.FromFloat64(float64(catalog.ConsumptionBonusPPMPerUnit))
	factor := decimal.One.Add(consumed.Mul(rate).Div(decimal.FromFloat64(1_000_000))).Quantize(decimal.CanonicalSignificantDigits)
	if !factor.IsStateValue() || !factor.Gt(decimal.Zero) {
		return nil, errors.New("guild stock-consumption factor outside Decimal domain")
	}
	return []multiplier.Contribution{{Slot: multiplier.SlotFaction, SourceID: guild.StockConsumptionSourceID, Target: "all", Factor: factor}}, nil
}
