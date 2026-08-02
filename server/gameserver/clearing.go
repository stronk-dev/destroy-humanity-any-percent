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
	catalogs clearingCatalogs
	clock    func() time.Time
	limit    int
}

type clearingCatalogs interface {
	save.CatalogResolver
	ResolveFaction(string) (*faction.Catalog, bool)
}

func NewClearingDriver(db *sql.DB, service *guild.Service, catalogs clearingCatalogs, clock func() time.Time) (*ClearingDriver, error) {
	if db == nil || service == nil || catalogs == nil {
		return nil, ErrClearingDriver
	}
	if clock == nil {
		clock = time.Now
	}
	return &ClearingDriver{db: db, service: service, catalogs: catalogs, clock: clock, limit: 64}, nil
}

func (driver *ClearingDriver) Tick(ctx context.Context) (int, error) {
	committed := 0
	lastGuildID := "00000000-0000-0000-0000-000000000000"
	for {
		guildIDs, err := driver.guildPage(ctx, lastGuildID)
		if err != nil {
			return committed, err
		}
		for _, guildID := range guildIDs {
			changed, err := retryClearingSnapshot(ctx, func() error { return driver.commitGuild(ctx, guildID) })
			if err != nil {
				return committed, err
			}
			if changed {
				committed++
			}
		}
		if len(guildIDs) < driver.limit {
			return committed, nil
		}
		lastGuildID = guildIDs[len(guildIDs)-1]
	}
}

func (driver *ClearingDriver) commitGuild(ctx context.Context, guildID string) error {
	members, stockCap, err := driver.members(ctx, guildID)
	if err != nil {
		return err
	}
	var previous sql.NullInt64
	if err := driver.db.QueryRowContext(ctx, `SELECT max(boundary_seq) FROM guild_clearing_results WHERE guild_id=$1`, guildID).Scan(&previous); err != nil {
		return err
	}
	boundary := int64(1)
	if previous.Valid {
		boundary = previous.Int64 + 1
	}
	return driver.service.CommitClearingBoundary(ctx, guildID, boundary, save.CanonicalServerTime(driver.clock()), members, stockCap)
}

func retryClearingSnapshot(ctx context.Context, commit func() error) (bool, error) {
	for attempt := 0; attempt < 3; attempt++ {
		if err := commit(); err != nil {
			if errors.Is(err, guild.ErrClearingSnapshotChanged) {
				if err := ctx.Err(); err != nil {
					return false, err
				}
				continue
			}
			return false, err
		}
		return true, nil
	}
	// Ordinary membership churn is not a worker failure. The next scheduled
	// tick rebuilds the snapshot from the then-current membership.
	return false, nil
}

func (driver *ClearingDriver) guildPage(ctx context.Context, after string) ([]string, error) {
	rows, err := driver.db.QueryContext(ctx, `SELECT guild_id FROM guilds WHERE disbanded_at IS NULL AND guild_id>$1 ORDER BY guild_id LIMIT $2`, after, driver.limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var guildIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		guildIDs = append(guildIDs, id)
	}
	return guildIDs, rows.Err()
}

func (driver *ClearingDriver) members(ctx context.Context, guildID string) ([]guild.ClearingMember, int64, error) {
	rows, err := driver.db.QueryContext(ctx, `
		SELECT member.account_id,founder.founder_id,stream.id,revision.version,revision.constants_hash,revision.state::text,revision.created_at,
		       COALESCE(reserved.debit_units,0),COALESCE(reserved.credit_units,0)
		FROM guild_members member
		JOIN account_founders founder ON founder.account_id=member.account_id AND founder.archived_at IS NULL
		JOIN save_streams stream ON stream.owner_kind='founder' AND stream.owner_id=founder.founder_id AND stream.scope='company' AND stream.archived_at IS NULL
		JOIN LATERAL (SELECT version,constants_hash,state,created_at FROM save_revisions WHERE stream_id=stream.id ORDER BY revision DESC LIMIT 1) revision ON true
		LEFT JOIN LATERAL (
			SELECT sum(result.debit_units) debit_units,sum(result.credit_units) credit_units
			FROM guild_clearing_results result
			WHERE result.guild_id=member.guild_id AND result.account_id=member.account_id
			  AND result.founder_id=founder.founder_id AND result.company_stream_id=stream.id
			  AND result.run_seq=COALESCE(NULLIF(revision.state->>'run_seq','')::bigint,1)
			  AND result.boundary_seq>CASE
				WHEN COALESCE(revision.state->>'guild_boundary_guild_id','')=member.guild_id::text
				THEN COALESCE(NULLIF(revision.state->>'guild_boundary_seq','')::bigint,0)
				ELSE 0 END
		) reserved ON true
		WHERE member.guild_id=$1 AND member.left_at IS NULL
		ORDER BY member.account_id`, guildID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []guild.ClearingMember
	var stockCap int64
	for rows.Next() {
		var accountID, founderID, companyStreamID, constantsHash string
		var version int
		var encoded []byte
		var createdAt time.Time
		var reservedDebit, reservedCredit int64
		if err := rows.Scan(&accountID, &founderID, &companyStreamID, &version, &constantsHash, &encoded, &createdAt, &reservedDebit, &reservedCredit); err != nil {
			return nil, 0, err
		}
		economyCatalog, ok := driver.catalogs.Resolve(constantsHash)
		if !ok {
			return nil, 0, ErrClearingDriver
		}
		factionCatalog, ok := driver.catalogs.ResolveFaction(constantsHash)
		if !ok {
			return nil, 0, ErrClearingDriver
		}
		stockCap, err = mergePinnedStockCap(stockCap, factionCatalog.StockCap)
		if err != nil {
			return nil, 0, err
		}
		state, err := save.RestoreState(encoded, version, economyCatalog, "company", createdAt)
		if err != nil {
			return nil, 0, err
		}
		if reservedDebit < 0 || reservedDebit > state.StockUnits || reservedCredit < 0 || reservedCredit > factionCatalog.StockCap-state.ConsumedStockUnits {
			return nil, 0, ErrClearingDriver
		}
		member := guild.MemberStock{AccountID: accountID}
		if state.FactionID != "" {
			definition, ok := factionCatalog.Faction(state.FactionID)
			if !ok {
				return nil, 0, ErrClearingDriver
			}
			member.Produces, member.Consumes = definition.Produces, definition.Consumes
			member.AvailableUnits = state.StockUnits - reservedDebit
			member.ReceivedUnits = state.ConsumedStockUnits + reservedCredit
		}
		result = append(result, guild.ClearingMember{Stock: member, FounderID: founderID, CompanyStreamID: companyStreamID, RunSeq: state.RunSeq})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if len(result) == 0 || stockCap <= 0 {
		return nil, 0, guild.ErrClearingSnapshotChanged
	}
	return result, stockCap, nil
}

func mergePinnedStockCap(current, candidate int64) (int64, error) {
	if candidate <= 0 {
		return 0, ErrClearingDriver
	}
	if current == 0 || current == candidate {
		return candidate, nil
	}
	// DESIGN-GAP: cross-epoch clearing cannot choose one cap when active
	// members' pinned faction catalogs disagree. Fail closed until an RFC
	// defines the migration/exchange policy for a stock-cap change.
	return 0, ErrClearingDriver
}
