package save

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cloud-clicker/server/decimal"
)

var ErrFounderGenesisUnavailable = errors.New("Founder genesis unavailable")

type FounderGenesis struct {
	FounderStreamID string
	Revision        int64
	State           []byte
	Version         int
	ConstantsHash   string
}

// InsertFounderGenesisTx pins the exact pre-command Founder revision from
// which the immutable Founder log begins. Conflicting retries fail closed.
func InsertFounderGenesisTx(ctx context.Context, tx *sql.Tx, genesis FounderGenesis) error {
	if tx == nil || !uuidPattern.MatchString(genesis.FounderStreamID) || genesis.Revision < 1 ||
		genesis.Revision > decimal.MaxExactInteger || genesis.Version < 1 || genesis.Version == 15 ||
		genesis.Version > LatestFounderVersion || len(genesis.State) == 0 ||
		!hashPattern.MatchString(genesis.ConstantsHash) {
		return ErrInvalidState
	}
	var scope string
	var revisionHash string
	if err := tx.QueryRowContext(ctx, `
		SELECT stream.scope,revision.constants_hash
		FROM save_streams stream
		JOIN save_revisions revision ON revision.stream_id=stream.id AND revision.revision=$2
		WHERE stream.id=$1 AND stream.owner_kind='founder' AND stream.archived_at IS NULL
		FOR SHARE OF stream,revision`, genesis.FounderStreamID, genesis.Revision).Scan(&scope, &revisionHash); err != nil {
		return err
	}
	if scope != "founder" || revisionHash != genesis.ConstantsHash {
		return fmt.Errorf("%w: Founder genesis identity differs from source revision", ErrInvalidState)
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO founder_genesis(founder_stream_id,revision,state,version,constants_hash)
		VALUES($1,$2,$3,$4,$5)
		ON CONFLICT DO NOTHING`, genesis.FounderStreamID, genesis.Revision, genesis.State,
		genesis.Version, genesis.ConstantsHash)
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if affected == 1 {
		return nil
	}
	var existing FounderGenesis
	if err := tx.QueryRowContext(ctx, `
		SELECT revision,state,version,constants_hash
		FROM founder_genesis WHERE founder_stream_id=$1 FOR SHARE`, genesis.FounderStreamID).
		Scan(&existing.Revision, &existing.State, &existing.Version, &existing.ConstantsHash); err != nil {
		return err
	}
	if existing.Revision != genesis.Revision || existing.Version != genesis.Version ||
		existing.ConstantsHash != genesis.ConstantsHash || !bytes.Equal(existing.State, genesis.State) {
		return fmt.Errorf("%w: conflicting Founder genesis", ErrInvalidState)
	}
	return nil
}

func (s *Store) LoadFounderGenesis(ctx context.Context, founderStreamID string) (FounderGenesis, error) {
	if s == nil || !uuidPattern.MatchString(founderStreamID) {
		return FounderGenesis{}, ErrInvalidStream
	}
	genesis := FounderGenesis{FounderStreamID: founderStreamID}
	if err := s.db.QueryRowContext(ctx, `
		SELECT revision,state,version,constants_hash
		FROM founder_genesis WHERE founder_stream_id=$1`, founderStreamID).
		Scan(&genesis.Revision, &genesis.State, &genesis.Version, &genesis.ConstantsHash); errors.Is(err, sql.ErrNoRows) {
		return FounderGenesis{}, ErrFounderGenesisUnavailable
	} else if err != nil {
		return FounderGenesis{}, err
	}
	return genesis, nil
}
