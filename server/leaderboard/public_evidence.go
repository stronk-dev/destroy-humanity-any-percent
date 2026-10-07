package leaderboard

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"cloud-clicker/server/decimal"
)

var (
	ErrUnknownPublicRun         = errors.New("unknown public run")
	ErrInvalidPublicRunEvidence = errors.New("invalid public run evidence")
)

// PublicRunEvidence is C14's immutable evidence source, not a wire DTO. It
// contains no account/session data. Genesis and ReplayLog are the exact stored
// bytes, never JSON re-encoding or gzip recompression. Hashes use sha256:hex.
type PublicRunEvidence struct {
	RunID           string
	EngineVersion   string
	ConstantsHash   string
	GenesisVersion  int
	Genesis         []byte
	GenesisSHA256   string
	ReplayLog       []byte
	ReplayLogSHA256 string
}

// PublicRunEvidence authorizes only actual board records. A verified queue row,
// archived bytes or a pinned run alone cannot publish private replay evidence.
// One statement reads authorization and immutable evidence at the same snapshot.
// Missing/corrupt evidence for a public run is an invariant, not an unknown run.
func (repository *Repository) PublicRunEvidence(ctx context.Context, streamID string, runSeq int64) (PublicRunEvidence, error) {
	if repository == nil || repository.db == nil {
		return PublicRunEvidence{}, ErrInvalidPublicRunEvidence
	}
	if !uuidPattern.MatchString(streamID) || runSeq < 1 || runSeq > decimal.MaxExactInteger {
		return PublicRunEvidence{}, ErrUnknownPublicRun
	}
	runID := streamID + ":" + strconv.FormatInt(runSeq, 10)
	var engine, constantsHash, genesisHash, encoding, archiveHash sql.NullString
	var version sql.NullInt64
	var genesis, archive []byte
	err := repository.db.QueryRowContext(ctx, `
SELECT p.engine_version,p.constants_hash,g.constants_hash,g.version,g.state,a.encoding,a.bytes,a.sha256
FROM (SELECT 1 FROM verified_runs WHERE run_id=$1 LIMIT 1) authorized
LEFT JOIN run_epochs p ON p.company_stream_id=$2 AND p.run_seq=$3
LEFT JOIN run_genesis g ON g.company_stream_id=$2 AND g.run_seq=$3
LEFT JOIN run_log_archive a ON a.company_stream_id=$2 AND a.run_seq=$3
`, runID, streamID, runSeq).Scan(&engine, &constantsHash, &genesisHash, &version, &genesis, &encoding, &archive, &archiveHash)
	if errors.Is(err, sql.ErrNoRows) {
		return PublicRunEvidence{}, ErrUnknownPublicRun
	}
	if err != nil {
		return PublicRunEvidence{}, err
	}
	if !engine.Valid || !constantsHash.Valid || !genesisHash.Valid || constantsHash.String != genesisHash.String ||
		!version.Valid || version.Int64 < 1 || !json.Valid(genesis) ||
		!encoding.Valid || encoding.String != "gzip+json.v1" || len(archive) == 0 ||
		!archiveHash.Valid || archiveHash.String != publicEvidenceSHA256(archive) {
		return PublicRunEvidence{}, ErrInvalidPublicRunEvidence
	}
	return PublicRunEvidence{
		RunID: runID, EngineVersion: engine.String, ConstantsHash: constantsHash.String,
		GenesisVersion: int(version.Int64), Genesis: genesis, GenesisSHA256: publicEvidenceSHA256(genesis),
		ReplayLog: archive, ReplayLogSHA256: archiveHash.String,
	}, nil
}

func publicEvidenceSHA256(data []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data))
}
