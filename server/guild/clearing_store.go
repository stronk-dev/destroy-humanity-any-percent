package guild

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/save"
)

// ErrClearingSnapshotChanged is retryable: guild membership changed after a
// clearing worker read its candidate snapshot but before it acquired the guild
// lock. The worker must rebuild from committed membership instead of treating
// normal join/leave traffic as corruption.
var ErrClearingSnapshotChanged = errors.Join(ErrInvalidExchange, errors.New("guild clearing snapshot changed"))

type Settlement struct {
	GuildID     string
	BoundarySeq int64
	DebitUnits  int64
	CreditUnits int64
}

type SettlementBatch struct {
	GuildID     string
	BaseSeq     int64
	Settlements []Settlement
}

type ClearingMember struct {
	Stock           MemberStock
	FounderID       string
	CompanyStreamID string
	RunSeq          int64
}

func (service *Service) CommitClearingBoundary(ctx context.Context, guildID string, boundarySeq int64, committedAt time.Time, members []ClearingMember, stockCap int64) error {
	if service == nil || !uuidV7Pattern.MatchString(guildID) || boundarySeq <= 0 || boundarySeq > decimal.MaxExactInteger || committedAt.IsZero() {
		return ErrInvalidExchange
	}
	stocks := make([]MemberStock, len(members))
	identityByAccount := make(map[string]ClearingMember, len(members))
	for index, member := range members {
		if !uuidPattern.MatchString(member.Stock.AccountID) || !uuidPattern.MatchString(member.FounderID) || !uuidPattern.MatchString(member.CompanyStreamID) ||
			member.RunSeq <= 0 || member.RunSeq > decimal.MaxExactInteger || identityByAccount[member.Stock.AccountID].Stock.AccountID != "" {
			return ErrInvalidExchange
		}
		stocks[index] = member.Stock
		identityByAccount[member.Stock.AccountID] = member
	}
	states, clearings, err := Clear(service.catalog, stocks, stockCap)
	if err != nil {
		return err
	}
	snapshotHash, err := clearingSnapshotHash(members, stockCap)
	if err != nil {
		return err
	}
	byAccount := make(map[string]Settlement, len(states))
	allocations := make(map[string][]Allocation, len(states))
	before := make(map[string]MemberStock, len(stocks))
	for _, member := range stocks {
		before[member.AccountID] = member
	}
	for _, state := range states {
		prior := before[state.AccountID]
		byAccount[state.AccountID] = Settlement{GuildID: guildID, BoundarySeq: boundarySeq, DebitUnits: prior.AvailableUnits - state.AvailableUnits, CreditUnits: state.ReceivedUnits - prior.ReceivedUnits}
	}
	for _, clearing := range clearings {
		allocations[clearing.ProducerAccountID] = clearing.Allocations
	}
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var revision int64
	if err := tx.QueryRowContext(ctx, `SELECT revision FROM guilds WHERE guild_id=$1 AND disbanded_at IS NULL FOR UPDATE`, guildID).Scan(&revision); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrClearingSnapshotChanged
		}
		return err
	}
	var committedHash string
	err = tx.QueryRowContext(ctx, `SELECT snapshot_hash FROM guild_clearing_results WHERE guild_id=$1 AND boundary_seq=$2 LIMIT 1`, guildID, boundarySeq).Scan(&committedHash)
	if err == nil {
		if committedHash != snapshotHash {
			return ErrInvalidExchange
		}
		return tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT account_id FROM guild_members WHERE guild_id=$1 AND left_at IS NULL ORDER BY account_id FOR UPDATE`, guildID)
	if err != nil {
		return err
	}
	var activeAccounts []string
	for rows.Next() {
		var accountID string
		if err := rows.Scan(&accountID); err != nil {
			rows.Close()
			return err
		}
		activeAccounts = append(activeAccounts, accountID)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	memberAccounts := make([]string, len(stocks))
	for index := range stocks {
		memberAccounts[index] = stocks[index].AccountID
	}
	sort.Strings(memberAccounts)
	if !slices.Equal(activeAccounts, memberAccounts) {
		return ErrClearingSnapshotChanged
	}
	var last sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT max(boundary_seq) FROM guild_clearing_results WHERE guild_id=$1`, guildID).Scan(&last); err != nil {
		return err
	}
	if last.Valid && boundarySeq <= last.Int64 {
		return ErrInvalidExchange
	}
	if last.Valid && boundarySeq != last.Int64+1 || !last.Valid && boundarySeq != 1 {
		return ErrInvalidExchange
	}
	for _, state := range states {
		settlement := byAccount[state.AccountID]
		identity := identityByAccount[state.AccountID]
		encoded, _ := json.Marshal(allocations[state.AccountID])
		if allocations[state.AccountID] == nil {
			encoded = []byte("[]")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO guild_clearing_results(guild_id,boundary_seq,account_id,debit_units,credit_units,allocations,committed_at,snapshot_hash,founder_id,company_stream_id,run_seq) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			guildID, boundarySeq, state.AccountID, settlement.DebitUnits, settlement.CreditUnits, encoded, committedAt.UTC(), snapshotHash,
			identity.FounderID, identity.CompanyStreamID, identity.RunSeq); err != nil {
			return err
		}
	}
	for _, clearing := range clearings {
		revision++
		payload, _ := json.Marshal(clearing)
		if _, err := tx.ExecContext(ctx, `INSERT INTO guild_events(guild_id,revision,kind,actor_account,subject_account,payload) VALUES($1,$2,'exchange_cleared',$3,$3,$4)`, guildID, revision, clearing.ProducerAccountID, payload); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE guilds SET revision=$2 WHERE guild_id=$1`, guildID, revision); err != nil {
		return err
	}
	return tx.Commit()
}

func (service *Service) PendingSettlements(ctx context.Context, founderID, watermarkGuildID string, afterSeq int64) (SettlementBatch, error) {
	if service == nil || !uuidPattern.MatchString(founderID) || watermarkGuildID != "" && !uuidV7Pattern.MatchString(watermarkGuildID) || afterSeq < 0 {
		return SettlementBatch{}, ErrInvalidExchange
	}
	var guildID, accountID, companyStreamID string
	var runSeq int64
	var joinedAt, founderCreatedAt time.Time
	err := service.db.QueryRowContext(ctx, `SELECT member.guild_id,founder.account_id,member.joined_at,founder.created_at,stream.id,
		COALESCE(NULLIF(revision.state->>'run_seq','')::bigint,1)
		FROM account_founders founder
		JOIN guild_members member ON member.account_id=founder.account_id AND member.left_at IS NULL
		JOIN save_streams stream ON stream.owner_kind='founder' AND stream.owner_id=founder.founder_id AND stream.scope='company' AND stream.archived_at IS NULL
		JOIN LATERAL (SELECT state FROM save_revisions WHERE stream_id=stream.id ORDER BY revision DESC LIMIT 1) revision ON true
		WHERE founder.founder_id=$1 AND founder.archived_at IS NULL`, founderID).Scan(&guildID, &accountID, &joinedAt, &founderCreatedAt, &companyStreamID, &runSeq)
	if errors.Is(err, sql.ErrNoRows) {
		return SettlementBatch{}, nil
	}
	if err != nil {
		return SettlementBatch{}, err
	}
	eligibleFrom := joinedAt
	newFounderReset := founderCreatedAt.After(joinedAt)
	if newFounderReset {
		eligibleFrom = founderCreatedAt
	}
	comparison := "<"
	if newFounderReset {
		// Equal millisecond timestamps are excluded at a New-Founder boundary:
		// archived and fresh Founders never share settlement effects.
		comparison = "<="
	}
	var reset sql.NullInt64
	if err := service.db.QueryRowContext(ctx, `SELECT max(boundary_seq) FROM guild_clearing_results WHERE guild_id=$1 AND committed_at `+comparison+` $2`, guildID, eligibleFrom).Scan(&reset); err != nil {
		return SettlementBatch{}, err
	}
	resetSeq := int64(0)
	if reset.Valid {
		resetSeq = reset.Int64
	}
	if watermarkGuildID != guildID && (watermarkGuildID != "" || resetSeq != 0) {
		return SettlementBatch{GuildID: guildID, BaseSeq: resetSeq}, nil
	}
	baseSeq := afterSeq
	if watermarkGuildID == "" {
		baseSeq = resetSeq
	}
	eligibleComparison := ">="
	if newFounderReset {
		eligibleComparison = ">"
	}
	rows, err := service.db.QueryContext(ctx, `SELECT boundary_seq,debit_units,credit_units
		FROM guild_clearing_results
		WHERE guild_id=$1 AND account_id=$2 AND boundary_seq>$3 AND committed_at `+eligibleComparison+` $4
		  AND founder_id=$5 AND company_stream_id=$6 AND run_seq=$7
		ORDER BY boundary_seq`, guildID, accountID, baseSeq, eligibleFrom, founderID, companyStreamID, runSeq)
	if err != nil {
		return SettlementBatch{}, err
	}
	defer rows.Close()
	var result []Settlement
	for rows.Next() {
		var value Settlement
		if err := rows.Scan(&value.BoundarySeq, &value.DebitUnits, &value.CreditUnits); err != nil {
			return SettlementBatch{}, err
		}
		value.GuildID = guildID
		result = append(result, value)
	}
	if err := rows.Err(); err != nil {
		return SettlementBatch{}, err
	}
	return SettlementBatch{GuildID: guildID, BaseSeq: baseSeq, Settlements: result}, nil
}

func ApplySettlements(state *save.State, batch SettlementBatch, stockCap int64) error {
	if state == nil || stockCap <= 0 {
		return ErrInvalidExchange
	}
	if batch.GuildID == "" {
		if batch.BaseSeq != 0 || len(batch.Settlements) != 0 {
			return ErrInvalidExchange
		}
		return nil
	}
	if !uuidV7Pattern.MatchString(batch.GuildID) || batch.BaseSeq < 0 || batch.BaseSeq > decimal.MaxExactInteger {
		return ErrInvalidExchange
	}
	if state.GuildBoundaryGuildID != batch.GuildID {
		if state.GuildBoundaryGuildID == "" {
			if batch.BaseSeq < state.GuildBoundarySeq || batch.BaseSeq > state.GuildBoundarySeq && len(batch.Settlements) != 0 {
				return ErrInvalidExchange
			}
		} else if len(batch.Settlements) != 0 {
			return ErrInvalidExchange
		}
		state.GuildBoundaryGuildID = batch.GuildID
		state.GuildBoundarySeq = batch.BaseSeq
		state.GuildConsumedWindow = 0
	} else if batch.BaseSeq != state.GuildBoundarySeq {
		return ErrInvalidExchange
	}
	for _, settlement := range batch.Settlements {
		if settlement.GuildID != batch.GuildID || settlement.BoundarySeq <= state.GuildBoundarySeq || settlement.DebitUnits < 0 || settlement.CreditUnits < 0 || settlement.DebitUnits > state.StockUnits || settlement.CreditUnits > stockCap-state.ConsumedStockUnits {
			return ErrInvalidExchange
		}
		state.StockUnits -= settlement.DebitUnits
		state.ConsumedStockUnits += settlement.CreditUnits
		state.GuildConsumedWindow = settlement.CreditUnits
		state.GuildBoundarySeq = settlement.BoundarySeq
	}
	return nil
}

func clearingSnapshotHash(members []ClearingMember, stockCap int64) (string, error) {
	ordered := append([]ClearingMember(nil), members...)
	sort.Slice(ordered, func(left, right int) bool { return ordered[left].Stock.AccountID < ordered[right].Stock.AccountID })
	encoded, err := json.Marshal(struct {
		StockCap int64            `json:"stock_cap"`
		Members  []ClearingMember `json:"members"`
	}{StockCap: stockCap, Members: ordered})
	if err != nil {
		return "", ErrInvalidExchange
	}
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}
