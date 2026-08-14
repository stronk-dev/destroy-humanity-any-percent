package account

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

var (
	ErrBootstrapExpired   = errors.New("bootstrap receipt expired")
	ErrBootstrapInvariant = errors.New("bootstrap invariant failed")
	bootstrapKeyPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// BootstrapReceiptKeys is deployment-owned authenticated-encryption material.
// Previous keys remain readable until every receipt written under them expires.
type BootstrapReceiptKeys struct {
	CurrentID string
	Current   []byte
	Previous  map[string][]byte
}

func (keys BootstrapReceiptKeys) validate() error {
	if keys.CurrentID == "" || len(keys.Current) != 32 {
		return ErrInvalidRequest
	}
	for id, key := range keys.Previous {
		if id == "" || id == keys.CurrentID || len(key) != 32 {
			return ErrInvalidRequest
		}
	}
	return nil
}

func (keys BootstrapReceiptKeys) resolve(id string) ([]byte, bool) {
	if id == keys.CurrentID {
		return keys.Current, true
	}
	key, ok := keys.Previous[id]
	return key, ok
}

type BootstrapSnapshotBuilder interface {
	InitialGameUISnapshot(context.Context, string, string, *save.State, *save.State, []save.FrozenContribution, time.Time) (json.RawMessage, error)
}

type BootstrapFaultInjector func(string) error
type BootstrapResponseValidator func([]byte) error

type bootstrapAccount struct {
	AccountID    string `json:"account_id"`
	CreatedAt    string `json:"created_at"`
	RecoveryCode string `json:"recovery_code"`
}

type bootstrapSession struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type bootstrapResponse struct {
	Account        bootstrapAccount `json:"account"`
	Session        bootstrapSession `json:"session"`
	GameUISnapshot json.RawMessage  `json:"game_ui_snapshot"`
}

var bootstrapFaultPoints = []string{
	"after_request_lock", "after_account_insert", "after_founder_link", "after_streams", "after_session_family",
	"after_tokens", "after_snapshot", "after_response_validation", "after_receipt_encryption", "after_receipt_insert", "before_commit",
}

func injectBootstrapFault(injector BootstrapFaultInjector, point string) error {
	if injector != nil {
		return injector(point)
	}
	return nil
}

func bootstrapDigest(key string) ([32]byte, bool) {
	if !bootstrapKeyPattern.MatchString(key) {
		return [32]byte{}, false
	}
	decoded, err := hex.DecodeString(key)
	if err != nil || len(decoded) != 32 {
		return [32]byte{}, false
	}
	return sha256.Sum256([]byte(key)), true
}

func bootstrapAdditionalData(digest [32]byte, keyID string) []byte {
	result := append([]byte("cloud-clicker.bootstrap-receipt.v1\x00"), digest[:]...)
	result = append(result, 0)
	return append(result, keyID...)
}

func encryptBootstrapReceipt(random io.Reader, keys BootstrapReceiptKeys, digest [32]byte, plaintext []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(keys.Current)
	if err != nil {
		return nil, nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(random, nonce); err != nil {
		return nil, nil, err
	}
	return nonce, aead.Seal(nil, nonce, plaintext, bootstrapAdditionalData(digest, keys.CurrentID)), nil
}

func decryptBootstrapReceipt(keys BootstrapReceiptKeys, digest [32]byte, keyID string, nonce, ciphertext []byte) ([]byte, error) {
	key, ok := keys.resolve(keyID)
	if !ok {
		return nil, ErrInvalidRequest
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil || len(nonce) != aead.NonceSize() {
		return nil, ErrInvalidRequest
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, bootstrapAdditionalData(digest, keyID))
	if err != nil {
		return nil, ErrInvalidRequest
	}
	return plaintext, nil
}

// CreateBootstrap atomically creates the account, both Founder streams, a
// session family, the exact initial UI snapshot, and its protected retry
// receipt. The advisory lock serializes equal keys before any durable row is
// created; all later failures roll the whole transaction back.
func (repository *Repository) CreateBootstrap(ctx context.Context, key string, builder BootstrapSnapshotBuilder,
	keys BootstrapReceiptKeys, validate BootstrapResponseValidator, injector BootstrapFaultInjector) (json.RawMessage, error) {
	digest, ok := bootstrapDigest(key)
	if !ok {
		return nil, ErrInvalidRequest
	}
	if repository == nil || builder == nil || validate == nil || keys.validate() != nil {
		return nil, ErrBootstrapInvariant
	}
	now := save.CanonicalServerTime(repository.clock())
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	lockID := int64(binary.BigEndian.Uint64(digest[:8]))
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, lockID); err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "after_request_lock"); err != nil {
		return nil, err
	}
	var keyID sql.NullString
	var nonce, ciphertext []byte
	var refreshExpires, tombstonedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT key_id,nonce,ciphertext,refresh_expires_at,tombstoned_at
		FROM bootstrap_receipts WHERE request_digest=$1 FOR UPDATE`, digest[:]).Scan(&keyID, &nonce, &ciphertext, &refreshExpires, &tombstonedAt)
	if err == nil {
		if tombstonedAt.Valid || !refreshExpires.Valid || !refreshExpires.Time.After(now) {
			if !tombstonedAt.Valid {
				if _, updateErr := tx.ExecContext(ctx, `UPDATE bootstrap_receipts SET key_id=NULL,nonce=NULL,ciphertext=NULL,tombstoned_at=$2
					WHERE request_digest=$1 AND tombstoned_at IS NULL`, digest[:], now); updateErr != nil {
					return nil, updateErr
				}
			}
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			return nil, ErrBootstrapExpired
		}
		if !keyID.Valid || len(nonce) == 0 || len(ciphertext) == 0 {
			return nil, ErrBootstrapInvariant
		}
		plaintext, err := decryptBootstrapReceipt(keys, digest, keyID.String, nonce, ciphertext)
		if err != nil || validate(plaintext) != nil {
			return nil, ErrBootstrapInvariant
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return json.RawMessage(plaintext), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	accountID, err := newUUIDv7(now, repository.random)
	if err != nil {
		return nil, err
	}
	founderID, err := newUUIDv7(now, repository.random)
	if err != nil {
		return nil, err
	}
	recoveryCode, err := newRecoveryCode(repository.random)
	if err != nil {
		return nil, err
	}
	recoveryHash, err := hashRecoveryCode(recoveryCode, repository.random)
	if err != nil {
		return nil, err
	}
	states, err := repository.initialStates(now, founderID, nil)
	if err != nil {
		return nil, err
	}
	var createdAt time.Time
	if err := tx.QueryRowContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,$2,$3) RETURNING created_at`, accountID, recoveryHash, now).Scan(&createdAt); err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "after_account_insert"); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id,created_at) VALUES($1,$2,$3)`, accountID, founderID, now); err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "after_founder_link"); err != nil {
		return nil, err
	}
	if err := insertFounderStreams(ctx, tx, founderID, repository.constantsHash, states); err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "after_streams"); err != nil {
		return nil, err
	}
	familyID, err := newUUIDv7(now, repository.random)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO session_families(family_id,account_id,created_at) VALUES($1,$2,$3)`, familyID, accountID, now); err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "after_session_family"); err != nil {
		return nil, err
	}
	pair, err := repository.issueTokenPairAt(ctx, tx, accountID, founderID, familyID, now)
	if err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "after_tokens"); err != nil {
		return nil, err
	}
	snapshot, err := builder.InitialGameUISnapshot(ctx, repository.constantsHash, founderID,
		states.State[economy.ScopeCompany], states.State[economy.ScopeFounder], states.Frozen, now)
	if err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "after_snapshot"); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(bootstrapResponse{
		Account: bootstrapAccount{AccountID: accountID, CreatedAt: createdAt.UTC().Format("2006-01-02T15:04:05.000Z"), RecoveryCode: recoveryCode},
		Session: bootstrapSession{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken}, GameUISnapshot: snapshot,
	})
	if err != nil || validate(encoded) != nil {
		return nil, ErrBootstrapInvariant
	}
	if err := injectBootstrapFault(injector, "after_response_validation"); err != nil {
		return nil, err
	}
	nonce, ciphertext, err = encryptBootstrapReceipt(repository.random, keys, digest, encoded)
	if err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "after_receipt_encryption"); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO bootstrap_receipts
		(request_digest,account_id,key_id,nonce,ciphertext,created_at,refresh_expires_at)
		VALUES($1,$2,$3,$4,$5,$6,$7)`, digest[:], accountID, keys.CurrentID, nonce, ciphertext, now, now.Add(refreshTTL)); err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "after_receipt_insert"); err != nil {
		return nil, err
	}
	if err := injectBootstrapFault(injector, "before_commit"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return json.RawMessage(encoded), nil
}

// PruneExpiredBootstrapReceipts destroys credential ciphertext in bounded
// batches while retaining the request digest tombstone permanently.
func (repository *Repository) PruneExpiredBootstrapReceipts(ctx context.Context, before time.Time, limit int) (int64, error) {
	if repository == nil || before.IsZero() || limit < 1 || limit > 10_000 {
		return 0, ErrInvalidRequest
	}
	result, err := repository.db.ExecContext(ctx, `WITH expired AS MATERIALIZED (
		SELECT request_digest FROM bootstrap_receipts
		WHERE tombstoned_at IS NULL AND refresh_expires_at<=$1
		ORDER BY refresh_expires_at,request_digest LIMIT $2 FOR UPDATE SKIP LOCKED
	) UPDATE bootstrap_receipts receipt SET key_id=NULL,nonce=NULL,ciphertext=NULL,tombstoned_at=$1
	WHERE receipt.request_digest IN (SELECT request_digest FROM expired)`, save.CanonicalServerTime(before), limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
