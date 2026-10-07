package leaderboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"unicode/utf8"
)

var (
	ErrUnknownPublicCatalog    = errors.New("unknown public catalog")
	ErrInvalidPublicCatalog    = errors.New("invalid public catalog evidence")
	publicConstantsHashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

// PublicCatalogArtifact is an internal immutable-byte source, not a wire DTO.
// API Foundation C18 requires the owning artifact's schema before these bytes
// may be embedded in a public response. SHA256 identifies Bytes exactly.
type PublicCatalogArtifact struct {
	Name   string
	SHA256 string
	Bytes  []byte
}

type PublicCatalogBundle struct {
	ConstantsHash string
	Artifacts     []PublicCatalogArtifact
}

// PublicCatalog reads only constants hashes accepted by an epoch, including
// historical epochs. Authorization and artifacts share one statement snapshot;
// an orphan catalog set is not public. Missing or corrupt accepted bytes are an
// invariant failure, not an unknown hash or a partial success.
//
// The reader never loads current files or manufactures a formulas artifact.
// Formula availability and C18 wire validation belong to the HTTP consumer.
func (repository *Repository) PublicCatalog(ctx context.Context, constantsHash string) (PublicCatalogBundle, error) {
	if repository == nil || repository.db == nil {
		return PublicCatalogBundle{}, ErrInvalidPublicCatalog
	}
	if !publicConstantsHashPattern.MatchString(constantsHash) {
		return PublicCatalogBundle{}, ErrUnknownPublicCatalog
	}
	rows, err := repository.db.QueryContext(ctx, `
SELECT artifacts.artifact_name, artifacts.bytes
FROM (SELECT DISTINCT constants_hash FROM epoch_hashes WHERE constants_hash=$1) accepted
LEFT JOIN catalog_artifacts artifacts ON artifacts.constants_hash=accepted.constants_hash
ORDER BY artifacts.artifact_name COLLATE "C"`, constantsHash)
	if err != nil {
		return PublicCatalogBundle{}, err
	}
	defer rows.Close()
	artifacts := []Artifact{}
	accepted := false
	for rows.Next() {
		accepted = true
		var name sql.NullString
		var data []byte
		if err := rows.Scan(&name, &data); err != nil {
			return PublicCatalogBundle{}, err
		}
		if !name.Valid {
			return PublicCatalogBundle{}, ErrInvalidPublicCatalog
		}
		artifacts = append(artifacts, Artifact{Name: name.String, Bytes: data})
	}
	if err := rows.Err(); err != nil {
		return PublicCatalogBundle{}, err
	}
	if !accepted {
		return PublicCatalogBundle{}, ErrUnknownPublicCatalog
	}
	return publicCatalogBundle(constantsHash, artifacts)
}

func publicCatalogBundle(constantsHash string, artifacts []Artifact) (PublicCatalogBundle, error) {
	computed, normalized, err := validateArtifacts(artifacts)
	if err != nil || computed != constantsHash {
		return PublicCatalogBundle{}, ErrInvalidPublicCatalog
	}
	names := make([]string, 0, len(normalized))
	for name, data := range normalized {
		if !utf8.Valid(data) || !json.Valid(data) {
			return PublicCatalogBundle{}, ErrInvalidPublicCatalog
		}
		names = append(names, name)
	}
	sort.Strings(names)
	result := PublicCatalogBundle{ConstantsHash: computed, Artifacts: make([]PublicCatalogArtifact, 0, len(names))}
	for _, name := range names {
		data := normalized[name]
		result.Artifacts = append(result.Artifacts, PublicCatalogArtifact{Name: name, SHA256: publicEvidenceSHA256(data), Bytes: data})
	}
	return result, nil
}
