package guild

import (
	"context"
	"database/sql"
	"time"
)

type HealthReader struct {
	DB       *sql.DB
	Catalogs CatalogResolver
	Windows  HealthWindowResolver
	Clock    func() time.Time
}

func (reader HealthReader) FounderGuildHealthPPM(ctx context.Context, founderID, constantsHash string) (int64, bool, error) {
	if reader.DB == nil || reader.Catalogs == nil || reader.Windows == nil || !uuidPattern.MatchString(founderID) {
		return 0, false, ErrInvalidTithe
	}
	now := time.Now().UTC()
	if reader.Clock != nil {
		now = reader.Clock().UTC()
	}
	catalog, ok := reader.Catalogs.ResolveGuild(constantsHash)
	if !ok {
		return 0, false, ErrInvalidTithe
	}
	windowMS, ok := reader.Windows.GuildHealthWindowMS(constantsHash)
	if !ok || windowMS <= 0 {
		return 0, false, ErrInvalidTithe
	}
	var guildID string
	err := reader.DB.QueryRowContext(ctx, `SELECT m.guild_id FROM account_founders f JOIN guild_members m ON m.account_id=f.account_id WHERE f.founder_id=$1 AND m.left_at IS NULL`, founderID).Scan(&guildID)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	var active, xp int64
	err = reader.DB.QueryRowContext(ctx, `SELECT count(DISTINCT account_id),COALESCE(sum(xp),0) FROM guild_activity_windows WHERE guild_id=$1 AND window_start>$2 AND window_start<=$3`, guildID, now.Add(-time.Duration(windowMS)*time.Millisecond), now).Scan(&active, &xp)
	if err != nil {
		return 0, true, err
	}
	health, err := HealthPPM(xp, active, catalog.GuildXPTargetPerFounder)
	return health, true, err
}
