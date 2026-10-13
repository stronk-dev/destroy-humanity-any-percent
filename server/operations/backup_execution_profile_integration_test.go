package operations

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/internal/backupfixture"
	"cloud-clicker/server/save"
	"filippo.io/age"
)

type backupCLIMessage struct {
	Status  string                  `json:"status"`
	Reason  string                  `json:"reason"`
	Backup  string                  `json:"backup"`
	Header  deploymentbackup.Header `json:"header"`
	Foreign []string                `json:"foreign"`
}

func runBackupExecutionProfile(t *testing.T, ctx context.Context, client *http.Client, received <-chan []byte, source *sql.DB) {
	t.Helper()
	workspace := t.TempDir()
	write := func(name string, data []byte) string {
		t.Helper()
		path := filepath.Join(workspace, name)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	var migration int
	if err := source.QueryRowContext(ctx, `SELECT max(version_id) FILTER (WHERE is_applied) FROM goose_db_version`).Scan(&migration); err != nil {
		t.Fatal(err)
	}
	epoch := backupfixture.Epoch(t, 8)
	manifest := backupfixture.Manifest(t, epoch, migration)
	var release struct {
		ConstantsHash string `json:"constants_hash"`
	}
	if err := json.Unmarshal(manifest, &release); err != nil {
		t.Fatal(err)
	}
	const accountID = "01986666-c001-7000-8000-000000000001"
	for _, row := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO catalog_sets(constants_hash) VALUES($1)`, []any{release.ConstantsHash}},
		{`INSERT INTO epochs(epoch_id,name,started_at,changelog_ref) VALUES(8,'Operations backup fixture',now(),'changelog/epoch-8.md')`, nil},
		{`INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES(8,$1)`, []any{release.ConstantsHash}},
		{`INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'before-retry')`, []any{accountID}},
	} {
		if _, err := source.ExecContext(ctx, row.sql, row.args...); err != nil {
			t.Fatal(err)
		}
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(workspace, "backups")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	flags := []string{"--target=" + target, "--database-url-file=" + write("source-url", []byte(os.Getenv("TEST_DATABASE_URL"))),
		"--release-manifest=" + write("manifest.json", manifest), "--epoch=" + write("epoch.json", epoch),
		"--age-recipient=" + identity.Recipient().String(), "--server-id=01986666-b001-4000-8000-000000000001", "--metrics-dir=/textfile"}
	command := exec.CommandContext(ctx, "/test/deployment-backup", append([]string{"create"}, flags...)...)
	baseline, err := command.Output()
	var initial backupCLIMessage
	if err != nil || json.Unmarshal(baseline, &initial) != nil || initial.Status != "completed" || filepath.Dir(initial.Backup) != target {
		t.Fatalf("actual initial pg_dump/encrypted backup failed: %v", err)
	}
	original, err := os.ReadFile(initial.Backup)
	if err != nil || len(original) == 0 {
		t.Fatal("actual initial encrypted backup absent")
	}
	if !waitForQueryValue(ctx, client, "cloud_clicker_backup_last_result", "1") ||
		!waitForQueryValue(ctx, client, "cloud_clicker_backup_last_success_timestamp_seconds", fmt.Sprint(initial.Header.CompletedAt.Unix())) ||
		!waitForAlertState(ctx, client, "CloudClickerBackupMissingOrFailed", "inactive") {
		t.Fatal("actual completed backup did not establish a healthy scraped baseline")
	}
	foreign := filepath.Join(target, "owned-obstruction")
	if err := os.WriteFile(foreign, []byte("do not remove this obstruction automatically\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	workerCtx, stopWorker := context.WithCancel(ctx)
	worker := exec.CommandContext(workerCtx, "/test/deployment-backup", append([]string{"schedule"}, flags...)...)
	stdout, err := worker.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics bytes.Buffer
	worker.Stderr = &diagnostics
	failedAt := time.Now().UTC()
	if err := worker.Start(); err != nil {
		stopWorker()
		t.Fatal(err)
	}
	events := make(chan backupCLIMessage, 4)
	finished := make(chan error, 1)
	go func() {
		defer close(events)
		reader := bufio.NewScanner(stdout)
		reader.Buffer(make([]byte, 4096), 64<<10)
		var observationErr error
		for reader.Scan() {
			var message backupCLIMessage
			if err := json.Unmarshal(reader.Bytes(), &message); err != nil {
				observationErr = err
				stopWorker()
				break
			}
			select {
			case events <- message:
			default:
				observationErr = fmt.Errorf("scheduled backup observation overflow")
				stopWorker()
			}
			if observationErr != nil {
				break
			}
		}
		if err := reader.Err(); err != nil {
			observationErr = err
			stopWorker()
		}
		waitErr := worker.Wait()
		if observationErr != nil {
			finished <- observationErr
		} else {
			finished <- waitErr
		}
		close(finished)
	}()
	defer func() { stopWorker(); <-finished }()
	await := func(status string) backupCLIMessage {
		t.Helper()
		select {
		case message, open := <-events:
			if !open || message.Status != status {
				t.Fatalf("scheduled backup did not reach %s: observed=%s", status, message.Status)
			}
			return message
		case err := <-finished:
			t.Fatalf("scheduled worker ended before %s: %v", status, err)
		case <-ctx.Done():
			t.Fatalf("scheduled backup observation incomplete: %v", ctx.Err())
		}
		return backupCLIMessage{}
	}
	blocked := await("blocked")
	if blocked.Reason != "target_not_exclusive_or_invalid" || len(blocked.Foreign) != 1 || blocked.Foreign[0] != "owned-obstruction" {
		t.Fatal("actual scheduled round did not report its exact target obstruction")
	}
	if !waitForQueryValue(ctx, client, "cloud_clicker_backup_last_result", "0") ||
		!waitForAlertState(ctx, client, "CloudClickerBackupMissingOrFailed", "firing") {
		t.Fatal("actual blocked scheduled round did not fire through node-exporter")
	}
	firingStart := waitForBackupNotification(t, ctx, received, "firing", failedAt, time.Time{})
	kept, err := os.ReadFile(initial.Backup)
	if err != nil || !bytes.Equal(kept, original) {
		t.Fatal("failed round changed the completed encrypted baseline")
	}
	obstruction, err := os.ReadFile(foreign)
	if err != nil || string(obstruction) != "do not remove this obstruction automatically\n" {
		t.Fatal("blocked worker changed or deleted the foreign entry")
	}
	t.Log("actual scheduled target refusal reached the configured receiver without changing the completed backup")
	if _, err := source.ExecContext(ctx, `UPDATE accounts SET recovery_hash='after-retry' WHERE account_id=$1`, accountID); err != nil {
		t.Fatal(err)
	}
	clearedAt := time.Now().UTC()
	if err := os.Remove(foreign); err != nil {
		t.Fatal(err)
	}
	recovered := await("completed")
	if recovered.Header.StartedAt.Sub(failedAt) < 5*time.Minute || recovered.Backup == initial.Backup || filepath.Dir(recovered.Backup) != target ||
		recovered.Header.PayloadSHA256 == initial.Header.PayloadSHA256 {
		t.Fatal("recovery did not use the unchanged retry window and a new package")
	}
	if !waitForQueryValue(ctx, client, "cloud_clicker_backup_last_result", "1") ||
		!waitForQueryValue(ctx, client, "cloud_clicker_backup_last_success_timestamp_seconds", fmt.Sprint(recovered.Header.CompletedAt.Unix())) ||
		!waitForAlertState(ctx, client, "CloudClickerBackupMissingOrFailed", "inactive") {
		t.Fatal("actual successful scheduled retry did not resolve its rule")
	}
	waitForBackupNotification(t, ctx, received, "resolved", clearedAt, firingStart)
	// Restore the retry into a new disposable database. A repeated old backup
	// or a fake pg_dump signature cannot satisfy the changed account-row check.
	name := "ops_backup_restore_" + strings.ToLower(rand.Text())
	if _, err := source.ExecContext(ctx, "CREATE DATABASE "+name+" TEMPLATE template0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := source.ExecContext(cleanup, "DROP DATABASE "+name); err != nil {
			t.Error(err)
		}
	}()
	parsed, err := url.Parse(os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	parsed.Path = "/" + name
	hash := fmt.Sprintf("sha256:%x", sha256.Sum256(manifest))
	if _, err := deploymentbackup.RestorePostgresBackup(ctx, deploymentbackup.PostgresRestoreInput{
		BackupPath: recovered.Backup, ExpectedManifestSHA256: hash,
		IdentityFile: write("restore-identity", []byte(identity.String())), TargetDatabaseURLFile: write("restore-url", []byte(parsed.String())),
	}); err != nil {
		t.Fatal(err)
	}
	restored, err := save.OpenPostgres(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	var value string
	if err := restored.QueryRowContext(ctx, `SELECT recovery_hash FROM accounts WHERE account_id=$1`, accountID).Scan(&value); err != nil || value != "after-retry" {
		t.Fatal("actual scheduled retry did not restore the changed account row")
	}
	t.Log("unchanged five-minute scheduled retry produced a real encrypted dump, matching resolution and restored updated SQL data")
}
