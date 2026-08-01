package gameserver

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cloud-clicker/server/faction"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/save"
)

var ErrClearingDriver = errors.New("invalid guild clearing driver")

type ClearingDriver struct {
	db       *sql.DB
	service  *guild.Service
	catalogs save.CatalogResolver
	factions *faction.Catalog
	clock    func() time.Time
	limit    int
}

func NewClearingDriver(db *sql.DB, service *guild.Service, catalogs save.CatalogResolver, factions *faction.Catalog, clock func() time.Time) (*ClearingDriver, error) {
	if db == nil || service == nil || catalogs == nil || factions == nil {
		return nil, ErrClearingDriver
	}
	if clock == nil {
		clock = time.Now
	}
	return &ClearingDriver{db: db, service: service, catalogs: catalogs, factions: factions, clock: clock, limit: 64}, nil
}

func (driver *ClearingDriver) Tick(ctx context.Context) (int, error) {
	rows, err := driver.db.QueryContext(ctx, `SELECT guild_id FROM guilds WHERE disbanded_at IS NULL ORDER BY guild_id LIMIT $1`, driver.limit)
	if err != nil {
		return 0, err
	}
	var guildIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		guildIDs = append(guildIDs, id)
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	committed := 0
	for _, guildID := range guildIDs {
		members, err := driver.members(ctx, guildID)
		if err != nil {
			return committed, err
		}
		var previous sql.NullInt64
		if err := driver.db.QueryRowContext(ctx, `SELECT max(boundary_seq) FROM guild_clearing_results WHERE guild_id=$1`, guildID).Scan(&previous); err != nil {
			return committed, err
		}
		boundary := int64(1)
		if previous.Valid {
			boundary = previous.Int64 + 1
		}
		if err := driver.service.CommitClearingBoundary(ctx, guildID, boundary, save.CanonicalServerTime(driver.clock()), members, driver.factions.StockCap); err != nil {
			return committed, err
		}
		committed++
	}
	return committed, nil
}

func (driver *ClearingDriver) members(ctx context.Context, guildID string) ([]guild.MemberStock, error) {
	rows, err := driver.db.QueryContext(ctx, `
		SELECT member.account_id,revision.version,revision.constants_hash,revision.state::text
		FROM guild_members member
		JOIN account_founders founder ON founder.account_id=member.account_id AND founder.archived_at IS NULL
		JOIN save_streams stream ON stream.owner_id=founder.founder_id AND stream.scope='company' AND stream.archived_at IS NULL
		JOIN LATERAL (SELECT version,constants_hash,state FROM save_revisions WHERE stream_id=stream.id ORDER BY revision DESC LIMIT 1) revision ON true
		WHERE member.guild_id=$1 AND member.left_at IS NULL
		ORDER BY member.account_id`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []guild.MemberStock
	for rows.Next() {
		var accountID, constantsHash string
		var version int
		var encoded []byte
		if err := rows.Scan(&accountID, &version, &constantsHash, &encoded); err != nil {
			return nil, err
		}
		economyCatalog, ok := driver.catalogs.Resolve(constantsHash)
		if !ok {
			return nil, ErrClearingDriver
		}
		state, err := save.RestoreState(encoded, version, economyCatalog, "company", time.Time{})
		if err != nil {
			return nil, err
		}
		member := guild.MemberStock{AccountID: accountID}
		if state.FactionID != "" {
			definition, ok := driver.factions.Faction(state.FactionID)
			if !ok {
				return nil, ErrClearingDriver
			}
			member.Produces, member.Consumes = definition.Produces, definition.Consumes
			member.AvailableUnits, member.ReceivedUnits = state.StockUnits, state.ConsumedStockUnits
		}
		result = append(result, member)
	}
	return result, rows.Err()
}
