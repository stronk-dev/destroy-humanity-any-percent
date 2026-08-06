package save

import (
	"bytes"
	"context"
	"database/sql"
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
)

type FrozenContribution struct {
	SourceID string
	Slot     multiplier.Slot
	Target   string
	Factor   string
}

func InsertRunFrozenContributionsTx(ctx context.Context, tx *sql.Tx, companyStreamID string, runSeq int64, values []FrozenContribution) error {
	if tx == nil || !uuidPattern.MatchString(companyStreamID) || runSeq < 1 || runSeq > decimal.MaxExactInteger {
		return ErrInvalidState
	}
	ordered := append([]FrozenContribution(nil), values...)
	sort.Slice(ordered, func(left, right int) bool { return ordered[left].SourceID < ordered[right].SourceID })
	for index, value := range ordered {
		factor, err := decimal.ParseCanonical(value.Factor)
		if !mechanicalIDPattern.MatchString(value.SourceID) || !multiplier.ValidSlot(value.Slot) ||
			value.Target != "all" && !mechanicalIDPattern.MatchString(value.Target) || err != nil ||
			!factor.IsStateValue() || !factor.Gt(decimal.Zero) || index > 0 && ordered[index-1].SourceID == value.SourceID {
			return ErrInvalidState
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO run_frozen_contributions(company_stream_id,run_seq,source_id,slot,target,factor)
			VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`, companyStreamID, runSeq, value.SourceID, value.Slot, value.Target, value.Factor); err != nil {
			return err
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT source_id,slot,target,factor FROM run_frozen_contributions
		WHERE company_stream_id=$1 AND run_seq=$2 ORDER BY source_id`, companyStreamID, runSeq)
	if err != nil {
		return err
	}
	defer rows.Close()
	existing := make([]FrozenContribution, 0, len(ordered))
	for rows.Next() {
		var value FrozenContribution
		if err := rows.Scan(&value.SourceID, &value.Slot, &value.Target, &value.Factor); err != nil {
			return err
		}
		existing = append(existing, value)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(existing) != len(ordered) {
		return ErrInvalidState
	}
	for index := range ordered {
		left, right := existing[index], ordered[index]
		if left.SourceID != right.SourceID || left.Slot != right.Slot || left.Target != right.Target || !bytes.Equal([]byte(left.Factor), []byte(right.Factor)) {
			return ErrInvalidState
		}
	}
	return nil
}

func LoadRunFrozenContributions(ctx context.Context, db *sql.DB, companyStreamID string, runSeq int64) ([]FrozenContribution, error) {
	if db == nil || !uuidPattern.MatchString(companyStreamID) || runSeq < 1 || runSeq > decimal.MaxExactInteger {
		return nil, ErrInvalidStream
	}
	rows, err := db.QueryContext(ctx, `SELECT source_id,slot,target,factor FROM run_frozen_contributions
		WHERE company_stream_id=$1 AND run_seq=$2 ORDER BY source_id`, companyStreamID, runSeq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []FrozenContribution
	for rows.Next() {
		var value FrozenContribution
		if err := rows.Scan(&value.SourceID, &value.Slot, &value.Target, &value.Factor); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if values == nil {
		values = []FrozenContribution{}
	}
	return values, nil
}
