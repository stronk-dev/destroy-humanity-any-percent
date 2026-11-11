package deploymentrelease

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/save"
)

const (
	crashChildFlag = "CLOUD_CLICKER_TEST_CRASH_CHILD"
	crashIntentID  = "01985555-d143-7000-8000-000000000001"
)

func TestCaddyIntegrationCrashRollsBackAdmittedPurchaseAndRetriesExactlyOnce(t *testing.T) {
	databaseURL, origin := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_CADDY_ORIGIN")
	if databaseURL == "" || origin == "" {
		t.Skip("deployment release integration environment is required")
	}
	requireCrashFixture(t, databaseURL, origin)
	ctx, cancel := context.WithTimeout(t.Context(), 40*time.Second)
	defer cancel()
	database, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.ExecContext(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public`); err != nil {
		t.Fatal(err)
	}
	if err := save.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}

	// Caddy's integration-only CA is outside the test container's trust store.
	client := &http.Client{Timeout: 12 * time.Second, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}}
	defer client.CloseIdleConnections()
	process := startCrashChild(t, ctx)
	waitCaddy(t, client, origin+"/readyz", http.StatusNoContent)
	bundle, err := epochseed.Load("/workspace")
	if err != nil {
		t.Fatal(err)
	}
	session, err := OpenAuthenticatedSmoke(ctx, client, origin, bundle.Hash)
	if err != nil {
		t.Fatal(err)
	}
	defer session.connection.CloseNow()
	var streamID string
	if err := database.QueryRowContext(ctx, `SELECT id FROM save_streams
		WHERE owner_kind='founder' AND owner_id=$1 AND scope='company'`, session.founderID).Scan(&streamID); err != nil {
		t.Fatal(err)
	}

	const seed = `{"intent_id":"01985555-d143-7000-8000-000000000002","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":12,"window_ms":500}`
	requireCrashReceipt(t, submitCrashIntent(ctx, session, seed), 12)
	before := readCrashPurchaseState(t, ctx, database, streamID)
	if before.revision != 2 || before.revisions != 2 || before.generators != 0 ||
		before.intents != 0 || before.events != 0 || before.commands != 0 || before.receipts != 0 {
		t.Fatalf("unexpected state before the purchase: %+v", before.crashPurchaseCounts)
	}

	blocker, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback()
	var blockerPID int
	if err := blocker.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
		t.Fatal(err)
	}
	// applyIntent has already inserted the save, events and run-log row when
	// this lock blocks its receipt insert. None of those writes can commit yet.
	if _, err := blocker.ExecContext(ctx, `LOCK TABLE intent_records IN SHARE MODE`); err != nil {
		t.Fatal(err)
	}
	const purchase = `{"intent_id":"` + crashIntentID + `","kind":"buy_generator","expected_revision":2,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`
	requestDone := make(chan crashIntentResponse, 1)
	go func() { requestDone <- submitCrashIntent(ctx, session, purchase) }()
	var waitingPID int
	waitForCrashCondition(t, ctx, "purchase reaching the blocked receipt insert", func(ctx context.Context) (bool, error) {
		err := database.QueryRowContext(ctx, `SELECT pid FROM pg_stat_activity
			WHERE $1 = ANY(pg_blocking_pids(pid)) AND query LIKE '%INSERT INTO intent_records%'`, blockerPID).Scan(&waitingPID)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return err == nil, err
	})
	if visible := readCrashPurchaseState(t, ctx, database, streamID); visible != before {
		t.Fatalf("blocked purchase already visible: %+v", visible.crashPurchaseCounts)
	}
	select {
	case early := <-requestDone:
		t.Fatalf("purchase finished before the crash: status=%d err=%v", early.status, early.err)
	default:
	}
	if err := process.command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	<-process.done
	var exit *exec.ExitError
	if !errors.As(process.err, &exit) {
		t.Fatalf("expected abrupt gameserver termination, got %v", process.err)
	}
	status, ok := exit.ProcessState.Sys().(syscall.WaitStatus)
	if !ok || status.Signal() != syscall.SIGKILL {
		t.Fatalf("gameserver was not killed: %v", process.err)
	}
	interrupted := <-requestDone
	if interrupted.status != http.StatusBadGateway || interrupted.err != nil {
		t.Fatalf("expected Caddy to report the interrupted purchase, status=%d err=%v", interrupted.status, interrupted.err)
	}
	// Postgres can keep the dead client's backend waiting for the lock until
	// the query resumes. Release the fixture lock only after the child is dead.
	if err := blocker.Rollback(); err != nil {
		t.Fatal(err)
	}
	waitForCrashCondition(t, ctx, "old Postgres backend exiting", func(ctx context.Context) (bool, error) {
		var gone bool
		err := database.QueryRowContext(ctx, `SELECT NOT EXISTS (SELECT 1 FROM pg_stat_activity WHERE pid=$1)`, waitingPID).Scan(&gone)
		return gone, err
	})
	if rolledBack := readCrashPurchaseState(t, ctx, database, streamID); rolledBack != before {
		t.Fatalf("interrupted purchase survived the crash: %+v", rolledBack.crashPurchaseCounts)
	}

	startCrashChild(t, ctx)
	waitCaddy(t, client, origin+"/readyz", http.StatusNoContent)
	if restarted := readCrashPurchaseState(t, ctx, database, streamID); restarted != before {
		t.Fatalf("restart changed the rolled-back state: %+v", restarted.crashPurchaseCounts)
	}
	first := submitCrashIntent(ctx, session, purchase)
	requireCrashReceipt(t, first, 1)
	after := readCrashPurchaseState(t, ctx, database, streamID)
	if after.revision != 3 || after.revisions != 3 || after.generators != 1 || after.intents != 1 ||
		after.events < 1 || after.purchases != 1 || after.commands != 1 || after.receipts != 1 || after.state == before.state {
		t.Fatalf("exact retry did not persist one purchase: %+v", after.crashPurchaseCounts)
	}
	duplicate := submitCrashIntent(ctx, session, purchase)
	requireCrashReceipt(t, duplicate, 1)
	if !bytes.Equal(first.body, duplicate.body) || readCrashPurchaseState(t, ctx, database, streamID) != after {
		t.Fatal("duplicate intent changed the receipt or the saved purchase")
	}
}

func requireCrashFixture(t *testing.T, databaseURL, origin string) {
	t.Helper()
	database, err := url.Parse(databaseURL)
	if err != nil || database.Scheme != "postgres" || database.Host != "release-postgres:5432" ||
		database.Path != "/cloud_clicker_release_test" || origin != "https://game.example:8443" {
		t.Fatal("crash witness requires the disposable deployment release Compose fixture")
	}
}

func TestReleaseCrashGameserverChild(t *testing.T) {
	if os.Getenv(crashChildFlag) != "1" {
		t.Skip("started by the Caddy crash integration test")
	}
	databaseURL, origin := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_CADDY_ORIGIN")
	requireCrashFixture(t, databaseURL, origin)
	ctx := t.Context()
	database, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	keys := account.SigningKeys{CurrentID: "release-crash-integration", Current: bytes.Repeat([]byte{0x74}, 32)}
	bootstrap := account.BootstrapReceiptKeys{CurrentID: "release-crash-bootstrap", Current: bytes.Repeat([]byte{0x75}, 32)}
	startReleaseServer(t, ctx, database, "/workspace", origin, keys, bootstrap)
	<-ctx.Done()
}

type crashChildProcess struct {
	command *exec.Cmd
	done    chan struct{}
	err     error
}

func startCrashChild(t *testing.T, ctx context.Context) *crashChildProcess {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(ctx, binary, "-test.run=^TestReleaseCrashGameserverChild$")
	command.Env = append(os.Environ(), crashChildFlag+"=1")
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	child := &crashChildProcess{command: command, done: make(chan struct{})}
	go func() {
		child.err = command.Wait()
		close(child.done)
	}()
	t.Cleanup(func() {
		_ = command.Process.Kill()
		<-child.done
	})
	return child
}

func waitForCrashCondition(t *testing.T, ctx context.Context, name string, condition func(context.Context) (bool, error)) {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if ok, err := condition(ctx); err != nil {
			t.Fatalf("observing %s: %v", name, err)
		} else if ok {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("waiting for %s: %v", name, ctx.Err())
		case <-ticker.C:
		}
	}
}

type crashIntentResponse struct {
	status int
	body   []byte
	err    error
}

func submitCrashIntent(ctx context.Context, session *SmokeSession, payload string) crashIntentResponse {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, session.origin+"/api/v1/intents", strings.NewReader(payload))
	if err != nil {
		return crashIntentResponse{err: err}
	}
	request.Header.Set("Authorization", "Bearer "+session.accessToken)
	request.Header.Set("Content-Type", "application/json")
	response, err := session.client.Do(request)
	if err != nil {
		return crashIntentResponse{err: err}
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	return crashIntentResponse{status: response.StatusCode, body: body, err: err}
}

func requireCrashReceipt(t *testing.T, response crashIntentResponse, count int) {
	t.Helper()
	var receipt struct {
		Outcome string `json:"outcome"`
		Count   int    `json:"applied_count"`
	}
	if response.err != nil || response.status != http.StatusOK || json.Unmarshal(response.body, &receipt) != nil ||
		receipt.Outcome != "applied" || receipt.Count != count {
		t.Fatalf("expected %d applied actions, status=%d outcome=%s count=%d err=%v",
			count, response.status, receipt.Outcome, receipt.Count, response.err)
	}
}

type crashPurchaseCounts struct {
	revision, revisions, generators, intents, events, purchases, commands, receipts int
}

type crashPurchaseState struct {
	crashPurchaseCounts
	state string
}

func readCrashPurchaseState(t *testing.T, ctx context.Context, database *sql.DB, streamID string) crashPurchaseState {
	t.Helper()
	var result crashPurchaseState
	err := database.QueryRowContext(ctx, `SELECT revision, state::text,
		(SELECT count(*) FROM save_revisions WHERE stream_id=$1),
		COALESCE((state->'generators'->>'generator.beige_tower')::int, 0),
		(SELECT count(*) FROM intent_records WHERE stream_id=$1 AND intent_id=$2),
		(SELECT count(*) FROM events WHERE stream_id=$1 AND intent_id=$2),
		(SELECT count(*) FROM events WHERE stream_id=$1 AND intent_id=$2 AND kind='generator_purchased'),
		(SELECT count(*) FROM run_log WHERE company_stream_id=$1 AND intent_id=$2),
		(SELECT count(*) FROM transport_player_outbox WHERE stream_id=$1 AND source_id=$2 AND message_kind='receipt')
		FROM save_revisions WHERE stream_id=$1 ORDER BY revision DESC LIMIT 1`, streamID, crashIntentID).Scan(
		&result.revision, &result.state, &result.revisions, &result.generators, &result.intents, &result.events,
		&result.purchases, &result.commands, &result.receipts)
	if err != nil || result.state == "" {
		t.Fatalf("reading purchase state: %v", err)
	}
	return result
}
