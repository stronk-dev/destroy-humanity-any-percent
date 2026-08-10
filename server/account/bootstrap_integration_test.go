package account

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

type bootstrapSnapshotFixture struct{}

func (bootstrapSnapshotFixture) InitialGameUISnapshot(_ context.Context, constantsHash, founderID string,
	company, founder *save.State, _ []save.FrozenContribution, now time.Time) (json.RawMessage, error) {
	if company == nil || founder == nil {
		return nil, errors.New("missing initial state")
	}
	return json.RawMessage(fmt.Sprintf(`{"constants_hash":%q,"evaluated_through_ms":%d,"facts":[],"generators":[],"manual_action":{"action_id":"manual.click","bucket_cap_milli":1,"refill_milli_per_ms":1,"refilled_at_ms":%d,"tokens_milli":0},"progress":[],"resources":[],"revision":1,"run":{"category":"any_percent","exit_count":0,"founder_id":%q,"run_seq":1,"run_started_at_ms":%d,"tier":0},"schema_version":1,"server_now_ms":%d,"upgrades":[]}`,
		constantsHash, now.UnixMilli(), now.UnixMilli(), founderID, now.UnixMilli(), now.UnixMilli())), nil
}

func bootstrapRepository(t *testing.T, now *time.Time) (*Repository, *sql.DB, BootstrapResponseValidator) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	truncateAccountIntegration(t, db)
	t.Cleanup(func() { truncateAccountIntegration(t, db) })
	bundle := epoch5AccountIntegrationBundle(t)
	seedAccountEpoch(t, db, bundle.Hash, bundle.Artifacts)
	catalog, err := economy.LoadCatalog(bundle.Artifacts["economy"])
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewRepository(db, integrationCatalogs{bundle.Hash: catalog}, bundle.Hash,
		SigningKeys{CurrentID: "bootstrap-session", Current: bytes.Repeat([]byte{0x51}, 32)}, func() time.Time { return *now }, nil)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := newPrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	return repository, db, func(data []byte) error {
		return registry.ValidateResponse("create_bootstrap", 201, data)
	}
}

func bootstrapCounts(t *testing.T, db *sql.DB) [7]int {
	t.Helper()
	var result [7]int
	err := db.QueryRow(`SELECT (SELECT count(*) FROM accounts),(SELECT count(*) FROM account_founders),
		(SELECT count(*) FROM save_streams),(SELECT count(*) FROM session_families),(SELECT count(*) FROM sessions),
		(SELECT count(*) FROM access_tokens),(SELECT count(*) FROM bootstrap_receipts)`).Scan(
		&result[0], &result[1], &result[2], &result[3], &result[4], &result[5], &result[6])
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCreateBootstrapAtomicRetryRotationExpiryAndFaultsIntegration(t *testing.T) {
	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	repository, db, validate := bootstrapRepository(t, &now)
	ctx := context.Background()
	keysV1 := BootstrapReceiptKeys{CurrentID: "receipt-v1", Current: bytes.Repeat([]byte{0x61}, 32)}
	key := fmt.Sprintf("%064x", 1)
	first, err := repository.CreateBootstrap(ctx, key, bootstrapSnapshotFixture{}, keysV1, validate, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := bootstrapCounts(t, db); got != [7]int{1, 1, 2, 1, 1, 1, 1} {
		t.Fatalf("first bootstrap rows=%v", got)
	}
	second, err := repository.CreateBootstrap(ctx, key, bootstrapSnapshotFixture{}, BootstrapReceiptKeys{
		CurrentID: "receipt-v2", Current: bytes.Repeat([]byte{0x62}, 32), Previous: map[string][]byte{"receipt-v1": keysV1.Current},
	}, validate, nil)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("rotated retry equality=%v err=%v", bytes.Equal(first, second), err)
	}
	var decoded bootstrapResponse
	if err := json.Unmarshal(first, &decoded); err != nil {
		t.Fatal(err)
	}
	var ciphertext []byte
	if err := db.QueryRow(`SELECT ciphertext FROM bootstrap_receipts`).Scan(&ciphertext); err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{decoded.Account.RecoveryCode, decoded.Session.AccessToken, decoded.Session.RefreshToken} {
		if bytes.Contains(ciphertext, []byte(secret)) {
			t.Fatalf("receipt ciphertext exposed secret %q", secret)
		}
	}

	baseline := bootstrapCounts(t, db)
	for index, point := range bootstrapFaultPoints {
		faultKey := fmt.Sprintf("%064x", index+2)
		_, err := repository.CreateBootstrap(ctx, faultKey, bootstrapSnapshotFixture{}, keysV1, validate, func(actual string) error {
			if actual == point {
				return errors.New("injected bootstrap fault")
			}
			return nil
		})
		if err == nil {
			t.Fatalf("fault point %s committed", point)
		}
		if got := bootstrapCounts(t, db); got != baseline {
			t.Fatalf("fault point %s left rows=%v want=%v", point, got, baseline)
		}
	}
	_, err = repository.CreateBootstrap(ctx, fmt.Sprintf("%064x", 88), bootstrapSnapshotFixture{}, keysV1,
		func([]byte) error { return errors.New("registry rejected response") }, nil)
	if !errors.Is(err, ErrBootstrapInvariant) || bootstrapCounts(t, db) != baseline {
		t.Fatalf("response validation did not fail atomically: err=%v rows=%v", err, bootstrapCounts(t, db))
	}

	concurrentKey := fmt.Sprintf("%064x", 99)
	var concurrent [2]json.RawMessage
	var concurrentErr [2]error
	var wait sync.WaitGroup
	wait.Add(2)
	for index := range concurrent {
		go func(index int) {
			defer wait.Done()
			concurrent[index], concurrentErr[index] = repository.CreateBootstrap(ctx, concurrentKey, bootstrapSnapshotFixture{}, keysV1, validate, nil)
		}(index)
	}
	wait.Wait()
	if concurrentErr[0] != nil || concurrentErr[1] != nil || !bytes.Equal(concurrent[0], concurrent[1]) {
		t.Fatalf("concurrent retry equal=%v errors=%v", bytes.Equal(concurrent[0], concurrent[1]), concurrentErr)
	}
	if got := bootstrapCounts(t, db); got != [7]int{2, 2, 4, 2, 2, 2, 2} {
		t.Fatalf("concurrent bootstrap rows=%v", got)
	}

	now = now.Add(refreshTTL)
	if _, err := repository.CreateBootstrap(ctx, key, bootstrapSnapshotFixture{}, keysV1, validate, nil); !errors.Is(err, ErrBootstrapExpired) {
		t.Fatalf("expired retry err=%v", err)
	}
	if _, err := repository.CreateBootstrap(ctx, key, bootstrapSnapshotFixture{}, keysV1, validate, nil); !errors.Is(err, ErrBootstrapExpired) {
		t.Fatalf("tombstone retry err=%v", err)
	}
	pruned, err := repository.PruneExpiredBootstrapReceipts(ctx, now, 1)
	if err != nil || pruned != 1 {
		t.Fatalf("bounded receipt GC pruned=%d err=%v", pruned, err)
	}
	if _, err := repository.CreateBootstrap(ctx, concurrentKey, bootstrapSnapshotFixture{}, keysV1, validate, nil); !errors.Is(err, ErrBootstrapExpired) {
		t.Fatalf("GC tombstone retry err=%v", err)
	}
	var tombstoned bool
	var liveSecrets int
	digest := sha256.Sum256([]byte(key))
	if err := db.QueryRow(`SELECT tombstoned_at IS NOT NULL,
		((key_id IS NOT NULL)::int+(nonce IS NOT NULL)::int+(ciphertext IS NOT NULL)::int)
		FROM bootstrap_receipts WHERE request_digest=$1`, digest[:]).Scan(&tombstoned, &liveSecrets); err != nil || !tombstoned || liveSecrets != 0 {
		t.Fatalf("tombstone=%v live secrets=%d err=%v", tombstoned, liveSecrets, err)
	}
	if got := bootstrapCounts(t, db); got != [7]int{2, 2, 4, 2, 2, 2, 2} {
		t.Fatalf("expired retry created rows=%v", got)
	}
	if _, err := db.Exec(`DELETE FROM bootstrap_receipts`); err == nil {
		t.Fatal("permanent bootstrap tombstones were deletable")
	}
	if _, err := db.Exec(`UPDATE bootstrap_receipts SET tombstoned_at=NULL WHERE request_digest=$1`, digest[:]); err == nil {
		t.Fatal("bootstrap tombstone transition was reversible")
	}
}
