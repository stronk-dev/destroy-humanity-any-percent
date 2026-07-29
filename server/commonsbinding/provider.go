package commonsbinding

import (
	"context"
	"database/sql"
	"errors"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

type SnapshotResolver interface {
	CompactSnapshot(context.Context, string) (commons.ContributionSnapshot, error)
}

type Provider struct {
	Catalogs  commons.CatalogSet
	Snapshots SnapshotResolver
}

func (provider Provider) Contributions(ctx context.Context, state *save.State, _ *economy.Catalog, revision save.Revision) ([]multiplier.Contribution, error) {
	if state == nil || !state.CompactMember {
		return nil, nil
	}
	if provider.Snapshots == nil {
		return nil, errors.New("commons health snapshot unavailable")
	}
	catalog, ok := provider.Catalogs.ResolveCommons(revision.ConstantsHash)
	if !ok {
		return nil, errors.New("commons catalog unavailable")
	}
	snapshot, err := provider.Snapshots.CompactSnapshot(ctx, revision.OwnerID)
	if errors.Is(err, sql.ErrNoRows) {
		snapshot.HealthPPM = catalog.NPCCompliancePPM
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return commons.Contribution(catalog, true, snapshot.HealthPPM, state.CompactSolidarityPPM)
}
