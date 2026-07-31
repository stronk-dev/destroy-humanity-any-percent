package guild

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/save"
)

type HealthWindowResolver interface {
	GuildHealthWindowMS(constantsHash string) (int64, bool)
}

type Projector struct {
	db      *sql.DB
	windows HealthWindowResolver
}

func NewProjector(db *sql.DB, windows HealthWindowResolver) (*Projector, error) {
	if db == nil || windows == nil {
		return nil, ErrInvalidTithe
	}
	return &Projector{db: db, windows: windows}, nil
}

type tithePayload struct {
	FounderID string `json:"founder_id"`
	RunID     struct {
		CompanyStreamID string `json:"company_stream_id"`
		RunSeq          int64  `json:"run_seq"`
	} `json:"run_id"`
	ProgressDeltaPPM int64 `json:"progress_delta_ppm"`
	XPDelta          int64 `json:"xp_delta"`
}

func (projector *Projector) Project(ctx context.Context, events []save.EventRecord) error {
	for _, event := range events {
		if event.Kind == save.EventGuildTitheAccrued {
			if err := projector.projectTithe(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

func (projector *Projector) projectTithe(ctx context.Context, event save.EventRecord) error {
	var payload tithePayload
	decoder := json.NewDecoder(bytes.NewReader(event.Payload))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&payload) != nil || decoder.Decode(&struct{}{}) != io.EOF || payload.FounderID != event.OwnerID ||
		payload.RunID.CompanyStreamID != event.StreamID || payload.RunID.RunSeq < 1 || payload.ProgressDeltaPPM <= 0 ||
		payload.ProgressDeltaPPM > 1_000_000 || payload.XPDelta <= 0 || payload.XPDelta > decimal.MaxExactInteger {
		return ErrInvalidTithe
	}
	windowMS, ok := projector.windows.GuildHealthWindowMS(event.ConstantsHash)
	if !ok || windowMS <= 0 {
		return ErrInvalidTithe
	}
	tx, err := projector.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO guild_projection_events(event_id,event_kind) VALUES($1,$2) ON CONFLICT DO NOTHING`, event.EventID, event.Kind)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return tx.Commit()
	}
	var accountID, guildID string
	err = tx.QueryRowContext(ctx, `SELECT f.account_id,m.guild_id FROM account_founders f JOIN guild_members m ON m.account_id=f.account_id
		WHERE f.founder_id=$1 AND m.joined_at<=$2 AND (m.left_at IS NULL OR m.left_at>$2)
		ORDER BY m.joined_at DESC LIMIT 1`, payload.FounderID, event.OccurredAt).Scan(&accountID, &guildID)
	if errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	}
	if err != nil {
		return err
	}
	var revision, guildXP int64
	if err := tx.QueryRowContext(ctx, `SELECT revision,guild_xp FROM guilds WHERE guild_id=$1 FOR UPDATE`, guildID).Scan(&revision, &guildXP); err != nil {
		return err
	}
	appliedXP := payload.XPDelta
	if appliedXP > decimal.MaxExactInteger-guildXP {
		appliedXP = decimal.MaxExactInteger - guildXP
	}
	windowStart := event.OccurredAt.UTC()
	windowFloor := event.OccurredAt.Add(-time.Duration(windowMS) * time.Millisecond)
	if _, err := tx.ExecContext(ctx, `UPDATE guilds SET guild_xp=guild_xp+$2,revision=revision+1 WHERE guild_id=$1`, guildID, appliedXP); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO guild_activity_windows(guild_id,window_start,account_id,xp) VALUES($1,$2,$3,$4)
		ON CONFLICT(guild_id,window_start,account_id) DO UPDATE SET xp=LEAST(9007199254740991,guild_activity_windows.xp+EXCLUDED.xp)`, guildID, windowStart, accountID, appliedXP); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO guild_health_inputs(guild_id,window_start,active_founders,tithed_xp)
		SELECT $1,$2,count(DISTINCT account_id),COALESCE(sum(xp),0) FROM guild_activity_windows WHERE guild_id=$1 AND window_start>$3 AND window_start<=$2
		ON CONFLICT(guild_id,window_start) DO UPDATE SET active_founders=EXCLUDED.active_founders,tithed_xp=EXCLUDED.tithed_xp`, guildID, windowStart, windowFloor); err != nil {
		return err
	}
	payloadJSON, _ := json.Marshal(map[string]any{"xp_delta": appliedXP, "source_event_id": event.EventID})
	if _, err := tx.ExecContext(ctx, `INSERT INTO guild_events(guild_id,revision,kind,actor_account,subject_account,intent_id,payload)
		VALUES($1,$2,'guild_xp_accrued',$3,$3,$4,$5)`, guildID, revision+1, accountID, nullableUUID(event.IntentID), payloadJSON); err != nil {
		return err
	}
	return tx.Commit()
}

func HealthPPM(windowXP, activeFounders, targetPerFounder int64) (int64, error) {
	if windowXP < 0 || activeFounders < 0 || targetPerFounder <= 0 || windowXP > decimal.MaxExactInteger || activeFounders > decimal.MaxExactInteger {
		return 0, ErrInvalidTithe
	}
	if activeFounders == 0 {
		return 0, nil
	}
	denominator := activeFounders * targetPerFounder
	if denominator <= 0 || denominator > decimal.MaxExactInteger {
		return 0, ErrInvalidTithe
	}
	if windowXP >= denominator {
		return 1_000_000, nil
	}
	numerator := new(big.Int).Mul(big.NewInt(windowXP), big.NewInt(1_000_000))
	numerator.Quo(numerator, big.NewInt(denominator))
	return numerator.Int64(), nil
}
