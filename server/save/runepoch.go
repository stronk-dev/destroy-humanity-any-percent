package save

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cloud-clicker/server/kernel"
	"cloud-clicker/server/runidentity"
)

var ErrEpochUnavailable = errors.New("balance epoch unavailable")

func PinRunToCurrentEpochTx(ctx context.Context, tx *sql.Tx, companyStreamID, founderID string, runSeq int64, constantsHash string) (int64, error) {
	if tx == nil || !uuidPattern.MatchString(companyStreamID) || !uuidPattern.MatchString(founderID) || runSeq < 1 || !hashPattern.MatchString(constantsHash) {
		return 0, ErrInvalidStream
	}
	var epochID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT e.epoch_id
		FROM epochs e
		JOIN epoch_hashes h ON h.epoch_id=e.epoch_id AND h.constants_hash=$1
		WHERE e.ended_at IS NULL
		FOR SHARE OF e`, constantsHash).Scan(&epochID); errors.Is(err, sql.ErrNoRows) {
		return 0, ErrEpochUnavailable
	} else if err != nil {
		return 0, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO run_epochs(company_stream_id,run_seq,epoch_id,constants_hash,engine_version,build_vcs_hash,seed)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT DO NOTHING`, companyStreamID, runSeq, epochID, constantsHash, kernel.Version, kernel.VCSRevision(), runidentity.SeedString(founderID, runSeq))
	if err != nil {
		return 0, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		var existingEpoch int64
		var existingHash, existingVersion, existingSeed string
		if err := tx.QueryRowContext(ctx, `SELECT epoch_id,constants_hash,engine_version,seed FROM run_epochs WHERE company_stream_id=$1 AND run_seq=$2`, companyStreamID, runSeq).Scan(&existingEpoch, &existingHash, &existingVersion, &existingSeed); err != nil {
			return 0, err
		}
		if existingHash != constantsHash || existingVersion != kernel.Version || existingSeed != runidentity.SeedString(founderID, runSeq) {
			return 0, fmt.Errorf("%w: conflicting run pin", ErrInvalidState)
		}
		return existingEpoch, nil
	}
	return epochID, nil
}

func (s *Store) PinRunToCurrentEpoch(ctx context.Context, companyStreamID, founderID string, runSeq int64, constantsHash string) (int64, error) {
	if s == nil {
		return 0, ErrInvalidStream
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var ownerID string
	var scope string
	if err := tx.QueryRowContext(ctx, `SELECT owner_id,scope FROM save_streams WHERE id=$1 FOR SHARE`, companyStreamID).Scan(&ownerID, &scope); err != nil {
		return 0, err
	}
	if ownerID != founderID || scope != "company" {
		return 0, ErrInvalidStream
	}
	epochID, err := PinRunToCurrentEpochTx(ctx, tx, companyStreamID, founderID, runSeq, constantsHash)
	if err != nil {
		return 0, err
	}
	var state []byte
	var version int
	var revisionHash string
	if err := tx.QueryRowContext(ctx, `SELECT state::text,version,constants_hash FROM save_revisions WHERE stream_id=$1 ORDER BY revision DESC LIMIT 1 FOR SHARE`, companyStreamID).Scan(&state, &version, &revisionHash); err != nil {
		return 0, err
	}
	if revisionHash != constantsHash {
		return 0, fmt.Errorf("%w: pin hash differs from latest revision", ErrInvalidState)
	}
	if err := InsertRunGenesisTx(ctx, tx, RunGenesis{CompanyStreamID: companyStreamID, RunSeq: runSeq, State: state, Version: version, ConstantsHash: constantsHash}); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return epochID, nil
}

func requireRunEpochTx(ctx context.Context, tx *sql.Tx, companyStreamID string, runSeq int64, constantsHash string) error {
	var storedHash, version string
	if err := tx.QueryRowContext(ctx, `SELECT constants_hash,engine_version FROM run_epochs WHERE company_stream_id=$1 AND run_seq=$2 FOR SHARE`, companyStreamID, runSeq).Scan(&storedHash, &version); errors.Is(err, sql.ErrNoRows) {
		return ErrEpochUnavailable
	} else if err != nil {
		return err
	}
	if storedHash != constantsHash {
		return fmt.Errorf("%w: run identity mismatch", ErrInvalidState)
	}
	if version != kernel.Version {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO run_version_drift(company_stream_id,run_seq,observed_version)
			VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, companyStreamID, runSeq, kernel.Version); err != nil {
			return err
		}
	}
	return nil
}
