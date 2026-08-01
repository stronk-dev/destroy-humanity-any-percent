package gameserver

import (
	"context"
	"database/sql"
	"errors"

	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/transport"
)

type onlineCounter interface{ ConnectionCount() int }
type epochReader interface {
	Current(context.Context) (leaderboard.Epoch, error)
}

type databaseWorldSource struct {
	db       *sql.DB
	serverID string
	online   onlineCounter
	epochs   epochReader
}

func (source *databaseWorldSource) SampleWorld(ctx context.Context) (WorldSample, error) {
	if source == nil || source.db == nil || source.online == nil || source.epochs == nil {
		return WorldSample{}, ErrWorldAggregator
	}
	var serverHealth, activeFounders int64
	err := source.db.QueryRowContext(ctx, `SELECT health_ppm,real_members FROM commons_health_scopes WHERE scope_kind='server' AND scope_id=$1`, source.serverID).Scan(&serverHealth, &activeFounders)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return WorldSample{}, err
	}
	var compactMembers, foundersTotal int64
	if err := source.db.QueryRowContext(ctx, `SELECT count(DISTINCT membership.founder_id) FROM company_compact_memberships membership JOIN founder_commons_assignments assignment ON assignment.founder_id=membership.founder_id WHERE membership.member=true AND assignment.server_id=$1`, source.serverID).Scan(&compactMembers); err != nil {
		return WorldSample{}, err
	}
	if err := source.db.QueryRowContext(ctx, `SELECT count(*) FROM account_founders`).Scan(&foundersTotal); err != nil {
		return WorldSample{}, err
	}
	epoch, err := source.epochs.Current(ctx)
	if err != nil {
		return WorldSample{}, err
	}
	return WorldSample{Planet: transport.WorldPlanet{}, Commons: transport.WorldCommons{ServerHealthPPM: serverHealth, ActiveFounders: activeFounders, CompactMembers: compactMembers},
		Population: transport.WorldPopulation{Online: int64(source.online.ConnectionCount()), FoundersTotal: foundersTotal},
		Milestones: transport.WorldMilestones{}, Epoch: transport.WorldEpoch{EpochID: epoch.ID, Name: epoch.Name}}, nil
}
