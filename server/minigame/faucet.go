package minigame

import (
	"context"
	"database/sql"
	"errors"
)

const attendedDayMS = int64(86_400_000)

type FaucetApplication struct {
	AttendedDay            int64
	QuotaBefore            int64
	QuotaAfter             int64
	RemainderBeforePPM     int64
	RemainderAfterPPM      int64
	ReducedScore           int64
	ConvertedUnits         int64
	CreditedUnits          int64
	ForfeitedUnits         int64
	ConfiguredCapReasonKey string
}

// applyFaucetWindowTx is deliberately unexported: the future resolve composer
// calls it only after the session's claim token has been validated in this
// same transaction. The claim-owned session transition is the idempotency
// authority; this function owns only the cross-run window arithmetic.
func applyFaucetWindowTx(ctx context.Context, tx *sql.Tx, founderID, minigameID string,
	effectiveFounderAttendedMS int64, policy PayoutPolicy, score, rateReductionPPM int64,
) (FaucetApplication, error) {
	if tx == nil || !uuidPattern.MatchString(founderID) || !mechanicalPattern.MatchString(minigameID) ||
		!validExactNonnegative(effectiveFounderAttendedMS) || !validExactNonnegative(score) {
		return FaucetApplication{}, ErrInvalidPayoutPolicy
	}
	if _, err := policy.MarshalJSON(); err != nil {
		return FaucetApplication{}, err
	}
	if err := lockFounder(ctx, tx, founderID); err != nil {
		return FaucetApplication{}, err
	}
	attendedDay := effectiveFounderAttendedMS / attendedDayMS
	if _, err := tx.ExecContext(ctx, insertFaucetWindowSQL, founderID, minigameID, attendedDay); err != nil {
		return FaucetApplication{}, err
	}
	var quotaUsed, remainderPPM int64
	if err := tx.QueryRowContext(ctx, lockFaucetWindowSQL, founderID, minigameID, attendedDay).Scan(&quotaUsed, &remainderPPM); err != nil {
		return FaucetApplication{}, err
	}
	conversion, err := ConvertPayout(score, rateReductionPPM, policy.ConversionPPM, remainderPPM)
	if err != nil {
		return FaucetApplication{}, err
	}
	credited := int64(0)
	quotaAfter := quotaUsed
	if quotaUsed < policy.SendsPerDay {
		quotaAfter++
		credited = conversion.ConvertedUnits
		if credited > policy.PerSendCap {
			credited = policy.PerSendCap
		}
	}
	forfeited := conversion.ConvertedUnits - credited
	reason := ""
	if forfeited > 0 {
		reason = policy.CapReasonKey
	}
	result, err := tx.ExecContext(ctx, updateFaucetWindowSQL, founderID, minigameID, attendedDay,
		quotaUsed, remainderPPM, quotaAfter, conversion.ConversionRemainderPPM)
	if err != nil {
		return FaucetApplication{}, err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return FaucetApplication{}, rowsErr
	} else if affected != 1 {
		return FaucetApplication{}, errors.New("minigame faucet window lock lost")
	}
	return FaucetApplication{
		AttendedDay: attendedDay, QuotaBefore: quotaUsed, QuotaAfter: quotaAfter,
		RemainderBeforePPM: remainderPPM, RemainderAfterPPM: conversion.ConversionRemainderPPM,
		ReducedScore: conversion.ReducedScore, ConvertedUnits: conversion.ConvertedUnits,
		CreditedUnits: credited, ForfeitedUnits: forfeited, ConfiguredCapReasonKey: reason,
	}, nil
}

const insertFaucetWindowSQL = `INSERT INTO minigame_faucet_window(
    founder_id,minigame_id,attended_day,quota_used,conversion_remainder_ppm
) VALUES($1,$2,$3,0,0) ON CONFLICT DO NOTHING`

const lockFaucetWindowSQL = `SELECT quota_used,conversion_remainder_ppm
FROM minigame_faucet_window
WHERE founder_id=$1 AND minigame_id=$2 AND attended_day=$3
FOR UPDATE`

const updateFaucetWindowSQL = `UPDATE minigame_faucet_window
SET quota_used=$6,conversion_remainder_ppm=$7,updated_at=clock_timestamp()
WHERE founder_id=$1 AND minigame_id=$2 AND attended_day=$3
  AND quota_used=$4 AND conversion_remainder_ppm=$5`
