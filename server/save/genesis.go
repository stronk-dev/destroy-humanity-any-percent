package save

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cloud-clicker/server/decimal"
)

var ErrGenesisUnavailable = errors.New("run genesis unavailable")

type RunGenesis struct {
	CompanyStreamID string
	RunSeq          int64
	State           []byte
	Version         int
	ConstantsHash   string
}

// PinRunWithGenesisTx creates the run identity and its immutable starting
// state in one caller-owned transaction. The exact bytes must also be the
// first persisted revision for that run.
func PinRunWithGenesisTx(ctx context.Context, tx *sql.Tx, companyStreamID, founderID string, runSeq int64, constantsHash string, version int, state []byte) (int64, error) {
	if version < 1 || version == 15 || version > LatestCompanyVersion || len(state) == 0 {
		return 0, ErrInvalidState
	}
	epochID, err := PinRunToCurrentEpochTx(ctx, tx, companyStreamID, founderID, runSeq, constantsHash)
	if err != nil {
		return 0, err
	}
	if err := InsertRunGenesisTx(ctx, tx, RunGenesis{CompanyStreamID: companyStreamID, RunSeq: runSeq, State: state, Version: version, ConstantsHash: constantsHash}); err != nil {
		return 0, err
	}
	return epochID, nil
}

func InsertRunGenesisTx(ctx context.Context, tx *sql.Tx, genesis RunGenesis) error {
	if tx == nil || !uuidPattern.MatchString(genesis.CompanyStreamID) || genesis.RunSeq < 1 || genesis.RunSeq > decimal.MaxExactInteger || genesis.Version < 1 || genesis.Version == 15 || genesis.Version > LatestCompanyVersion || len(genesis.State) == 0 || !hashPattern.MatchString(genesis.ConstantsHash) {
		return ErrInvalidState
	}
	var pinnedHash string
	if err := tx.QueryRowContext(ctx, `SELECT constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=$2 FOR SHARE`, genesis.CompanyStreamID, genesis.RunSeq).Scan(&pinnedHash); errors.Is(err, sql.ErrNoRows) {
		return ErrEpochUnavailable
	} else if err != nil {
		return err
	}
	if pinnedHash != genesis.ConstantsHash {
		return fmt.Errorf("%w: genesis hash differs from run pin", ErrInvalidState)
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO run_genesis(company_stream_id,run_seq,state,version,constants_hash)
		VALUES($1,$2,$3,$4,$5)
		ON CONFLICT DO NOTHING`, genesis.CompanyStreamID, genesis.RunSeq, genesis.State, genesis.Version, genesis.ConstantsHash)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 0 {
		return nil
	}
	var existing RunGenesis
	if err := tx.QueryRowContext(ctx, `SELECT state,version,constants_hash FROM run_genesis WHERE company_stream_id=$1 AND run_seq=$2 FOR SHARE`, genesis.CompanyStreamID, genesis.RunSeq).Scan(&existing.State, &existing.Version, &existing.ConstantsHash); err != nil {
		return err
	}
	if existing.Version != genesis.Version || existing.ConstantsHash != genesis.ConstantsHash || !bytes.Equal(existing.State, genesis.State) {
		return fmt.Errorf("%w: conflicting run genesis", ErrInvalidState)
	}
	return nil
}

func (s *Store) LoadRunGenesis(ctx context.Context, companyStreamID string, runSeq int64) (RunGenesis, error) {
	if s == nil || !uuidPattern.MatchString(companyStreamID) || runSeq < 1 || runSeq > decimal.MaxExactInteger {
		return RunGenesis{}, ErrInvalidStream
	}
	genesis := RunGenesis{CompanyStreamID: companyStreamID, RunSeq: runSeq}
	if err := s.db.QueryRowContext(ctx, `SELECT state,version,constants_hash FROM run_genesis WHERE company_stream_id=$1 AND run_seq=$2`, companyStreamID, runSeq).Scan(&genesis.State, &genesis.Version, &genesis.ConstantsHash); errors.Is(err, sql.ErrNoRows) {
		return RunGenesis{}, ErrGenesisUnavailable
	} else if err != nil {
		return RunGenesis{}, err
	}
	return genesis, nil
}
