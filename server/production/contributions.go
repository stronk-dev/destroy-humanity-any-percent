package production

import (
	"context"
	"errors"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

type CombinedContributionProviders []ContributionProvider

func (providers CombinedContributionProviders) Contributions(ctx context.Context, state *save.State, catalog *economy.Catalog, revision save.Revision) ([]multiplier.Contribution, error) {
	var combined []multiplier.Contribution
	seen := map[string]bool{}
	for _, provider := range providers {
		if provider == nil {
			return nil, errors.New("nil contribution provider")
		}
		values, err := provider.Contributions(ctx, state, catalog, revision)
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			if seen[value.SourceID] {
				return nil, errors.New("duplicate contribution source")
			}
			seen[value.SourceID] = true
			combined = append(combined, value)
		}
	}
	return combined, nil
}
