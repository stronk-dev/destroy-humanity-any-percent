package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

var (
	ErrInvalidRequest    = errors.New("invalid account request")
	ErrAuthentication    = errors.New("authentication failed")
	ErrRefreshReuse      = errors.New("refresh token reuse")
	ErrAccountNotFound   = errors.New("account not found")
	ErrFounderNotFound   = errors.New("founder not found")
	ErrImportUnavailable = errors.New("founder import unavailable")
)

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 30 * 24 * time.Hour
)

type Clock func() time.Time

type Repository struct {
	db            *sql.DB
	saves         *save.Store
	catalogs      save.CatalogResolver
	constantsHash string
	keys          SigningKeys
	clock         Clock
	random        io.Reader
}

type CreatedAccount struct {
	AccountID    string    `json:"account_id"`
	FounderID    string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	RecoveryCode string    `json:"recovery_code"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Founder struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Imported  bool      `json:"imported"`
}

type FounderState struct {
	FounderID     string          `json:"founder_id"`
	StreamID      string          `json:"-"`
	Revision      int64           `json:"revision"`
	Version       int             `json:"version"`
	ConstantsHash string          `json:"constants_hash"`
	State         json.RawMessage `json:"state"`
}

func NewRepository(db *sql.DB, catalogs save.CatalogResolver, constantsHash string, keys SigningKeys, clock Clock, random io.Reader) (*Repository, error) {
	if db == nil || catalogs == nil || constantsHash == "" || keys.validate() != nil {
		return nil, ErrInvalidRequest
	}
	if _, ok := catalogs.Resolve(constantsHash); !ok {
		return nil, ErrInvalidRequest
	}
	if clock == nil {
		clock = time.Now
	}
	if random == nil {
		random = rand.Reader
	}
	saves, err := save.NewStore(db, catalogs, nil)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	return &Repository{db: db, saves: saves, catalogs: catalogs, constantsHash: constantsHash, keys: keys, clock: clock, random: random}, nil
}

func (repository *Repository) CreateAccount(ctx context.Context) (CreatedAccount, error) {
	now := save.CanonicalServerTime(repository.clock())
	accountID, err := newUUIDv7(now, repository.random)
	if err != nil {
		return CreatedAccount{}, err
	}
	founderID, err := newUUIDv7(now, repository.random)
	if err != nil {
		return CreatedAccount{}, err
	}
	recoveryCode, err := newRecoveryCode(repository.random)
	if err != nil {
		return CreatedAccount{}, err
	}
	recoveryHash, err := hashRecoveryCode(recoveryCode, repository.random)
	if err != nil {
		return CreatedAccount{}, err
	}
	states, err := repository.initialStates(now)
	if err != nil {
		return CreatedAccount{}, err
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return CreatedAccount{}, err
	}
	defer tx.Rollback()
	var createdAt time.Time
	if err := tx.QueryRowContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,$2,$3) RETURNING created_at`, accountID, recoveryHash, now).Scan(&createdAt); err != nil {
		return CreatedAccount{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id,created_at) VALUES($1,$2,$3)`, accountID, founderID, now); err != nil {
		return CreatedAccount{}, err
	}
	if err := insertFounderStreams(ctx, tx, founderID, repository.constantsHash, states); err != nil {
		return CreatedAccount{}, err
	}
	if err := tx.Commit(); err != nil {
		return CreatedAccount{}, err
	}
	return CreatedAccount{AccountID: accountID, FounderID: founderID, CreatedAt: createdAt.UTC(), RecoveryCode: recoveryCode}, nil
}

func (repository *Repository) CreateSession(ctx context.Context, accountID, recoveryCode string) (TokenPair, error) {
	recoveryCode = normalizeRecoveryCode(recoveryCode)
	if accountID == "" || !validRecoveryCode(recoveryCode) {
		return TokenPair{}, ErrAuthentication
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return TokenPair{}, err
	}
	defer tx.Rollback()
	var encoded string
	if err := tx.QueryRowContext(ctx, `SELECT recovery_hash FROM accounts WHERE account_id=$1 FOR UPDATE`, accountID).Scan(&encoded); errors.Is(err, sql.ErrNoRows) {
		verifyRecoveryCode(dummyRecoveryHash, recoveryCode)
		return TokenPair{}, ErrAuthentication
	} else if err != nil {
		return TokenPair{}, err
	}
	valid, needsRehash := verifyRecoveryCodeForUpgrade(encoded, recoveryCode)
	if !valid {
		return TokenPair{}, ErrAuthentication
	}
	if needsRehash {
		upgraded, err := hashRecoveryCode(recoveryCode, repository.random)
		if err != nil {
			return TokenPair{}, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE accounts SET recovery_hash=$2 WHERE account_id=$1`, accountID, upgraded); err != nil {
			return TokenPair{}, err
		}
	}
	founderID, err := activeFounderForUpdate(ctx, tx, accountID)
	if err != nil {
		return TokenPair{}, ErrAuthentication
	}
	now := save.CanonicalServerTime(repository.clock())
	familyID, err := newUUIDv7(now, repository.random)
	if err != nil {
		return TokenPair{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO session_families(family_id,account_id,created_at) VALUES($1,$2,$3)`, familyID, accountID, now); err != nil {
		return TokenPair{}, err
	}
	pair, err := repository.issueTokenPair(ctx, tx, accountID, founderID, familyID)
	if err != nil {
		return TokenPair{}, err
	}
	if err := tx.Commit(); err != nil {
		return TokenPair{}, err
	}
	return pair, nil
}

func (repository *Repository) RefreshSession(ctx context.Context, refreshToken string) (TokenPair, error) {
	hash, ok := opaqueTokenHash(refreshToken)
	if !ok {
		return TokenPair{}, ErrAuthentication
	}
	now := save.CanonicalServerTime(repository.clock())
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return TokenPair{}, err
	}
	defer tx.Rollback()
	var familyID string
	if err := tx.QueryRowContext(ctx, `SELECT family_id FROM sessions WHERE token_hash=$1`, hash[:]).Scan(&familyID); errors.Is(err, sql.ErrNoRows) {
		return TokenPair{}, ErrAuthentication
	} else if err != nil {
		return TokenPair{}, err
	}
	accountID, familyRevoked, err := lockSessionFamily(ctx, tx, familyID)
	if err != nil {
		return TokenPair{}, err
	}
	if familyRevoked {
		return TokenPair{}, ErrAuthentication
	}
	var expiresAt time.Time
	var consumedAt, revokedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT expires_at,consumed_at,revoked_at FROM sessions WHERE token_hash=$1 AND family_id=$2 AND account_id=$3 FOR UPDATE`, hash[:], familyID, accountID).Scan(&expiresAt, &consumedAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return TokenPair{}, ErrAuthentication
	}
	if err != nil {
		return TokenPair{}, err
	}
	if consumedAt.Valid || revokedAt.Valid {
		if err := revokeFamily(ctx, tx, familyID, now); err != nil {
			return TokenPair{}, err
		}
		if err := tx.Commit(); err != nil {
			return TokenPair{}, err
		}
		return TokenPair{}, ErrRefreshReuse
	}
	if !expiresAt.After(now) {
		if err := revokeFamily(ctx, tx, familyID, now); err != nil {
			return TokenPair{}, err
		}
		if err := tx.Commit(); err != nil {
			return TokenPair{}, err
		}
		return TokenPair{}, ErrAuthentication
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET consumed_at=$2 WHERE token_hash=$1`, hash[:], now); err != nil {
		return TokenPair{}, err
	}
	founderID, err := activeFounderForUpdate(ctx, tx, accountID)
	if err != nil {
		return TokenPair{}, ErrAuthentication
	}
	pair, err := repository.issueTokenPair(ctx, tx, accountID, founderID, familyID)
	if err != nil {
		return TokenPair{}, err
	}
	if err := tx.Commit(); err != nil {
		return TokenPair{}, err
	}
	return pair, nil
}

func (repository *Repository) Authenticate(ctx context.Context, accessToken string) (Claims, error) {
	claims, err := verifyAccessToken(repository.keys, accessToken, repository.clock())
	if err != nil {
		return Claims{}, ErrAuthentication
	}
	var accountID, founderID string
	var expiresAt time.Time
	var revokedAt sql.NullTime
	err = repository.db.QueryRowContext(ctx, `SELECT account_id,founder_id,expires_at,revoked_at FROM access_tokens WHERE jti=$1`, claims.TokenID).Scan(&accountID, &founderID, &expiresAt, &revokedAt)
	if err != nil || accountID != claims.Subject || founderID != claims.FounderID || revokedAt.Valid || !expiresAt.After(repository.clock()) {
		return Claims{}, ErrAuthentication
	}
	return claims, nil
}

func (repository *Repository) RevokeSession(ctx context.Context, claims Claims) error {
	now := save.CanonicalServerTime(repository.clock())
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var familyID string
	if err := tx.QueryRowContext(ctx, `SELECT family_id FROM access_tokens WHERE jti=$1 AND account_id=$2`, claims.TokenID, claims.Subject).Scan(&familyID); err != nil {
		return ErrAuthentication
	}
	accountID, _, err := lockSessionFamily(ctx, tx, familyID)
	if err != nil || accountID != claims.Subject {
		return ErrAuthentication
	}
	if err := tx.QueryRowContext(ctx, `SELECT family_id FROM access_tokens WHERE jti=$1 AND account_id=$2 AND family_id=$3 FOR UPDATE`, claims.TokenID, claims.Subject, familyID).Scan(&familyID); err != nil {
		return ErrAuthentication
	}
	if err := revokeFamily(ctx, tx, familyID, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (repository *Repository) NewFounder(ctx context.Context, accountID string) (Founder, error) {
	now := save.CanonicalServerTime(repository.clock())
	founderID, err := newUUIDv7(now, repository.random)
	if err != nil {
		return Founder{}, err
	}
	states, err := repository.initialStates(now)
	if err != nil {
		return Founder{}, err
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return Founder{}, err
	}
	defer tx.Rollback()
	oldFounder, err := activeFounderForUpdate(ctx, tx, accountID)
	if err != nil {
		return Founder{}, ErrAccountNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE account_founders SET archived_at=$2 WHERE account_id=$1 AND archived_at IS NULL`, accountID, now); err != nil {
		return Founder{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE save_streams SET archived_at=$2 WHERE owner_kind='founder' AND owner_id=$1 AND archived_at IS NULL`, oldFounder, now); err != nil {
		return Founder{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id,created_at) VALUES($1,$2,$3)`, accountID, founderID, now); err != nil {
		return Founder{}, err
	}
	if err := insertFounderStreams(ctx, tx, founderID, repository.constantsHash, states); err != nil {
		return Founder{}, err
	}
	if err := tx.Commit(); err != nil {
		return Founder{}, err
	}
	return Founder{ID: founderID, CreatedAt: now, Imported: false}, nil
}

func (repository *Repository) ActiveFounder(ctx context.Context, accountID string) (Founder, error) {
	var founder Founder
	err := repository.db.QueryRowContext(ctx, `SELECT founder_id,created_at,imported FROM account_founders WHERE account_id=$1 AND archived_at IS NULL`, accountID).Scan(&founder.ID, &founder.CreatedAt, &founder.Imported)
	if errors.Is(err, sql.ErrNoRows) {
		return Founder{}, ErrFounderNotFound
	}
	if err != nil {
		return Founder{}, err
	}
	founder.CreatedAt = founder.CreatedAt.UTC()
	return founder, nil
}

func (repository *Repository) ActiveCompanyState(ctx context.Context, accountID string) (FounderState, error) {
	var state FounderState
	err := repository.db.QueryRowContext(ctx, `
		SELECT af.founder_id,s.id,r.revision,r.version,r.constants_hash,r.state
		FROM account_founders af
		JOIN save_streams s ON s.owner_kind='founder' AND s.owner_id=af.founder_id AND s.scope='company'
		JOIN LATERAL (SELECT revision,version,constants_hash,state FROM save_revisions WHERE stream_id=s.id ORDER BY revision DESC LIMIT 1) r ON true
		WHERE af.account_id=$1 AND af.archived_at IS NULL AND s.archived_at IS NULL`, accountID).Scan(&state.FounderID, &state.StreamID, &state.Revision, &state.Version, &state.ConstantsHash, &state.State)
	if errors.Is(err, sql.ErrNoRows) {
		return FounderState{}, ErrFounderNotFound
	}
	return state, err
}

func (repository *Repository) ImportFounder(ctx context.Context, accountID, constantsHash string, version int, rawState []byte) (Founder, error) {
	migrationCatalog, ok := repository.catalogs.Resolve(constantsHash)
	if !ok || version < 1 {
		return Founder{}, ErrInvalidRequest
	}
	now := save.CanonicalServerTime(repository.clock())
	// Pre-v4 saves do not carry canonical millisecond cursors. Import time is the
	// migration baseline, matching the standard restore path's requirement that
	// the caller supply an authoritative instant rather than inventing epoch zero.
	state, err := save.RestoreState(rawState, version, migrationCatalog, economy.ScopeCompany, now)
	if err != nil {
		return Founder{}, ErrInvalidRequest
	}
	// Run identity is server-owned. Imported history is intentionally unranked;
	// the migrated company begins a fresh run under the current authoritative
	// catalog and the run-1 pin created with its account.
	state.RunSeq = 1
	state.RunStartedAt = now
	state.OfflineSpans = []save.OfflineSpan{}
	state.OfferState = nil
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return Founder{}, err
	}
	defer tx.Rollback()
	founderID, err := activeFounderForUpdate(ctx, tx, accountID)
	if err != nil {
		return Founder{}, ErrAccountNotFound
	}
	var streamID string
	if err := tx.QueryRowContext(ctx, `SELECT s.id FROM save_streams s WHERE owner_kind='founder' AND owner_id=$1 AND scope='company' AND archived_at IS NULL`, founderID).Scan(&streamID); err != nil {
		return Founder{}, err
	}
	if _, err := repository.saves.WriteInTransaction(ctx, tx, streamID, 1, repository.constantsHash, state, save.WriteContext{Cause: "founder_import"}); errors.Is(err, save.ErrConflict) {
		return Founder{}, ErrImportUnavailable
	} else if err != nil {
		return Founder{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE account_founders SET imported=true WHERE founder_id=$1`, founderID); err != nil {
		return Founder{}, err
	}
	if err := tx.Commit(); err != nil {
		return Founder{}, err
	}
	return repository.ActiveFounder(ctx, accountID)
}

func (repository *Repository) DeleteAccount(ctx context.Context, accountID string) error {
	now := save.CanonicalServerTime(repository.clock())
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT founder_id FROM account_founders WHERE account_id=$1 FOR UPDATE`, accountID)
	if err != nil {
		return err
	}
	var founders []string
	for rows.Next() {
		var founderID string
		if err := rows.Scan(&founderID); err != nil {
			rows.Close()
			return err
		}
		founders = append(founders, founderID)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(founders) == 0 {
		return ErrAccountNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE account_founders SET archived_at=COALESCE(archived_at,$2) WHERE account_id=$1`, accountID, now); err != nil {
		return err
	}
	for _, founderID := range founders {
		if _, err := tx.ExecContext(ctx, `UPDATE save_streams SET archived_at=COALESCE(archived_at,$2) WHERE owner_kind='founder' AND owner_id=$1`, founderID, now); err != nil {
			return err
		}
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM accounts WHERE account_id=$1`, accountID)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return ErrAccountNotFound
	}
	return tx.Commit()
}

func (repository *Repository) issueTokenPair(ctx context.Context, tx *sql.Tx, accountID, founderID, familyID string) (TokenPair, error) {
	now := save.CanonicalServerTime(repository.clock())
	refreshToken, refreshHash, err := newOpaqueToken(repository.random)
	if err != nil {
		return TokenPair{}, err
	}
	jti, err := newUUIDv7(now, repository.random)
	if err != nil {
		return TokenPair{}, err
	}
	refreshExpires := now.Add(refreshTTL)
	accessExpires := now.Add(accessTTL)
	if _, err := tx.ExecContext(ctx, `INSERT INTO sessions(token_hash,family_id,account_id,created_at,expires_at) VALUES($1,$2,$3,$4,$5)`, refreshHash[:], familyID, accountID, now, refreshExpires); err != nil {
		return TokenPair{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO access_tokens(jti,family_id,account_id,founder_id,created_at,expires_at) VALUES($1,$2,$3,$4,$5,$6)`, jti, familyID, accountID, founderID, now, accessExpires); err != nil {
		return TokenPair{}, err
	}
	accessToken, err := signAccessToken(repository.keys, Claims{Subject: accountID, FounderID: founderID, ExpiresAt: accessExpires.Unix(), IssuedAt: now.Unix(), TokenID: jti})
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (repository *Repository) initialStates(now time.Time) (map[economy.Scope][]byte, error) {
	catalog, ok := repository.catalogs.Resolve(repository.constantsHash)
	if !ok {
		return nil, ErrInvalidRequest
	}
	result := make(map[economy.Scope][]byte, 2)
	for _, scope := range []economy.Scope{economy.ScopeCompany, economy.ScopeFounder} {
		ledger, err := economy.NewLedger(catalog, scope)
		if err != nil {
			return nil, err
		}
		counts := make(map[string]int64)
		for _, generator := range catalog.GeneratorClassesForScope(scope) {
			counts[generator.ID] = 0
		}
		state := &save.State{Ledger: ledger, GeneratorCounts: counts, EvaluatedThrough: now,
			ManualTokenRefilledAt: now, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
			LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{},
			HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{},
			LifetimeValue: decimal.Zero, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{},
			ExitHistory: []save.ExitRecord{}}
		if scope == economy.ScopeCompany {
			state.RunSeq = 1
			state.RunStartedAt = now
			state.ManualTokenMilli = catalog.ManualPolicy().BucketCapMilli
		}
		encoded, err := save.EncodeState(state)
		if err != nil {
			return nil, err
		}
		if _, err := save.RestoreState(encoded, save.CurrentVersion, catalog, scope, time.Time{}); err != nil {
			return nil, err
		}
		result[scope] = encoded
	}
	return result, nil
}

func insertFounderStreams(ctx context.Context, tx *sql.Tx, founderID, constantsHash string, states map[economy.Scope][]byte) error {
	for _, scope := range []economy.Scope{economy.ScopeCompany, economy.ScopeFounder} {
		var streamID string
		if err := tx.QueryRowContext(ctx, `INSERT INTO save_streams(owner_kind,owner_id,scope) VALUES('founder',$1,$2) RETURNING id`, founderID, scope).Scan(&streamID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES($1,1,$2,$3,$4)`, streamID, save.CurrentVersion, states[scope], constantsHash); err != nil {
			return err
		}
		if scope == economy.ScopeCompany {
			if _, err := save.PinRunToCurrentEpochTx(ctx, tx, streamID, founderID, 1, constantsHash); err != nil {
				return err
			}
		}
	}
	return nil
}

func activeFounderForUpdate(ctx context.Context, tx *sql.Tx, accountID string) (string, error) {
	var founderID string
	err := tx.QueryRowContext(ctx, `SELECT founder_id FROM account_founders WHERE account_id=$1 AND archived_at IS NULL FOR UPDATE`, accountID).Scan(&founderID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrFounderNotFound
	}
	return founderID, err
}

func revokeFamily(ctx context.Context, tx *sql.Tx, familyID string, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE session_families SET revoked_at=COALESCE(revoked_at,$2) WHERE family_id=$1`, familyID, now)
	if err != nil {
		return err
	}
	if changed, err := result.RowsAffected(); err != nil || changed != 1 {
		if err != nil {
			return err
		}
		return ErrAuthentication
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at=COALESCE(revoked_at,$2) WHERE family_id=$1`, familyID, now); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE access_tokens SET revoked_at=COALESCE(revoked_at,$2) WHERE family_id=$1`, familyID, now)
	return err
}

func lockSessionFamily(ctx context.Context, tx *sql.Tx, familyID string) (string, bool, error) {
	var accountID string
	var revokedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT account_id,revoked_at FROM session_families WHERE family_id=$1 FOR UPDATE`, familyID).Scan(&accountID, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, ErrAuthentication
	}
	return accountID, revokedAt.Valid, err
}

func newOpaqueToken(random io.Reader) (string, [32]byte, error) {
	var data [32]byte
	if _, err := io.ReadFull(random, data[:]); err != nil {
		return "", [32]byte{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(data[:])
	return token, sha256.Sum256([]byte(token)), nil
}

func opaqueTokenHash(token string) ([32]byte, bool) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decoded) != 32 || base64.RawURLEncoding.EncodeToString(decoded) != token {
		return [32]byte{}, false
	}
	return sha256.Sum256([]byte(token)), true
}
