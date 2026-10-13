package operations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func TestPrivateOperationsProfileIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_OPERATIONS_INTEGRATION") != "1" {
		t.Skip("requires isolated operations Compose profile")
	}
	// The preceding population measured 511.42s. The dead-letter arm adds
	// a quiet control, consecutive scrapes and unchanged notification intervals.
	// The twelve-minute completion guard is not an alert/latency bound.
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	received := make(chan []byte, 32)
	cleanupReceived := make(chan []byte, 32)
	var servers []*http.Server
	start := func(address string, handler http.Handler) {
		server := &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 2 * time.Second}
		servers = append(servers, server)
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				t.Errorf("fixture server %s: %v", address, err)
			}
		}()
	}
	// The gameserver target is the real production registry, not hand-written
	// exposition text, so label and series-shape defects reach real Prometheus.
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("isolated operations profile requires its actual Postgres service")
	}
	database, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := save.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(database)
	if err != nil {
		t.Fatal(err)
	}
	registry.SetReady(true)
	// These are actual production textfiles, collected by the real node-exporter
	// on the private network. The fixed clock intentionally makes success and
	// failure indistinguishable in the old whole-second timestamp comparison.
	backupAt := time.Now().UTC().Truncate(time.Second).Add(100 * time.Millisecond)
	if err := RecordOperation("/textfile", "backup", "success", backupAt); err != nil {
		t.Fatal(err)
	}
	if err := RecordCleanupJournal("/textfile", "", time.Now()); err != nil {
		t.Fatal(err)
	}
	// Model the declared one-minute host observation using the actual writer.
	// This is not a Docker journald fixture: that release-host gate stays open.
	observerCtx, stopObserver := context.WithCancel(ctx)
	observerDone := make(chan struct{})
	go func() {
		defer close(observerDone)
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-observerCtx.Done():
				return
			case at := <-ticker.C:
				if err := RecordCleanupJournal("/textfile", "", at); err != nil {
					t.Error(err)
					return
				}
			}
		}
	}()
	defer func() { stopObserver(); <-observerDone }()
	metrics := http.NewServeMux()
	metrics.Handle("/metrics", registry.Handler())
	start(":8080", metrics)
	var rejecting atomic.Bool
	receiver := http.NewServeMux()
	receiver.HandleFunc("/healthz", func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) })
	receiver.HandleFunc("/alerts", func(response http.ResponseWriter, request *http.Request) {
		data, err := io.ReadAll(http.MaxBytesReader(response, request.Body, 64<<10))
		if err != nil {
			t.Errorf("incomplete notification: %v", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		if rejecting.Load() {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		if strings.Contains(string(data), "CloudClickerCleanupJobFailed") {
			select {
			case cleanupReceived <- append([]byte(nil), data...):
			default:
				t.Error("cleanup notification observation overflow")
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		select {
		case received <- data:
		default:
			t.Error("notification observation overflow; evidence incomplete")
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	})
	start(":8090", receiver)
	defer func() {
		var wait sync.WaitGroup
		for _, server := range servers {
			wait.Add(1)
			go func(server *http.Server) { defer wait.Done(); _ = server.Shutdown(context.Background()) }(server)
		}
		wait.Wait()
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	for _, endpoint := range []string{"http://prometheus:9090/-/ready", "http://alertmanager:9093/-/ready", "http://node-exporter:9100/metrics"} {
		if !waitForHTTP(ctx, client, endpoint) {
			t.Fatalf("private operations target unavailable: %s", endpoint)
		}
	}
	if healthy, detail := waitForHealthyTargets(ctx, client); !healthy {
		t.Fatalf("Prometheus did not reach every private target: %s", detail)
	}
	if response, err := client.Get("http://caddy:2019/config/"); err == nil {
		_ = response.Body.Close()
		t.Fatalf("Caddy management API reachable from a peer container: status=%d", response.StatusCode)
	}
	nonce, err := VerifyAlertDelivery(ctx, client, "http://alertmanager:9093", "http://localhost:8090/healthz")
	if err != nil {
		t.Fatalf("receiver delivery proof failed: %v", err)
	}
	delivered := false
	for !delivered {
		select {
		case body := <-received:
			text := strings.ToLower(string(body))
			for _, forbidden := range []string{"account_id", "founder_id", "recovery_code", "client_ip"} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("private field delivered: %s", forbidden)
				}
			}
			delivered = strings.Contains(text, "cloudclickerreceivertest") && strings.Contains(text, nonce)
		case <-ctx.Done():
			t.Fatal("Alertmanager accepted but did not deliver the synthetic alert")
		}
	}

	// The ephemeral failed process is not the scrape target. Its actual bounded
	// logger event enters the journal decoder and persistent textfile; the fresh
	// registry's counter remains zero throughout. Real SQL production of this
	// event is separately tested by the gameserver integration population.
	failureQuery := `cloud_clicker_job_runs_total{job_name="credential_cleanup",result="failure"}`
	if !waitForQueryValue(ctx, client, failureQuery, "0") ||
		!waitForAlertState(ctx, client, "CloudClickerCleanupJobFailed", "inactive") {
		t.Fatal("Prometheus never scraped the pre-created cleanup failure series")
	}
	var jobLog bytes.Buffer
	failedRegistry, err := NewRegistryWithLogger(database, slog.New(slog.NewJSONHandler(&jobLog, nil)))
	if err != nil {
		t.Fatal(err)
	}
	if err := failedRegistry.ObserveJob("credential_cleanup", time.Now(), func() error { return errors.New("fixture cleanup failure") }); err == nil {
		t.Fatal("failed cleanup reported success")
	}
	row, _ := json.Marshal(map[string]string{"CONTAINER_TAG": "cloud-clicker-gameserver", "__CURSOR": "private-cleanup-fixture",
		"__REALTIME_TIMESTAMP": "1", "MESSAGE": strings.TrimSpace(jobLog.String())})
	cursor, err := CleanupFailureFromJournal(bytes.NewReader(row))
	if err != nil || cursor == "" {
		t.Fatal("failed process event did not reach the journal consumer")
	}
	cleanupDetectedAt := time.Now().UTC().Truncate(time.Second)
	if err := RecordCleanupJournal("/textfile", cursor, cleanupDetectedAt); err != nil {
		t.Fatal(err)
	}
	if !waitForAlertState(ctx, client, "CloudClickerCleanupJobFailed", "firing") {
		t.Fatal("persistent cleanup failure never fired CloudClickerCleanupJobFailed")
	}
	cleanupFiringStart := waitForOperationsNotification(t, ctx, cleanupReceived, "CloudClickerCleanupJobFailed", "node-exporter", "firing", cleanupDetectedAt, time.Time{})
	if !waitForQueryValue(ctx, client, failureQuery, "0") {
		t.Fatal("journal proof relied on the fresh process failure counter")
	}
	t.Log("retained failure reached the configured receiver while the fresh process counter stayed zero")

	// First prove a scraped healthy backup and no active backup alert. A prior
	// missing-metric notification from service startup cannot count as this
	// producer's failure: the received payload must start after failedAt.
	if !waitForQueryValue(ctx, client, "cloud_clicker_backup_last_result", "1") ||
		!waitForQueryValue(ctx, client, "cloud_clicker_backup_last_success_timestamp_seconds", fmt.Sprint(backupAt.Unix())) ||
		!waitForAlertState(ctx, client, "CloudClickerBackupMissingOrFailed", "inactive") {
		t.Fatal("real backup textfile did not establish a healthy baseline")
	}
	failedAt := time.Now().UTC()
	if err := RecordOperation("/textfile", "backup", "failure", backupAt.Add(time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	if !waitForQueryValue(ctx, client, "cloud_clicker_backup_last_result", "0") ||
		!waitForQueryValue(ctx, client, "cloud_clicker_backup_last_failure_timestamp_seconds", fmt.Sprint(backupAt.Unix())) ||
		!waitForAlertState(ctx, client, "CloudClickerBackupMissingOrFailed", "firing") {
		t.Fatal("same-second production backup failure did not fire through node-exporter")
	}
	firingStart := waitForBackupNotification(t, ctx, received, "firing", failedAt, time.Time{})
	t.Log("same-second backup failure reached the configured receiver through node-exporter and Prometheus")
	clearedAt := time.Now().UTC()
	if err := RecordOperation("/textfile", "backup", "success", backupAt.Add(2*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	if !waitForQueryValue(ctx, client, "cloud_clicker_backup_last_result", "1") ||
		!waitForAlertState(ctx, client, "CloudClickerBackupMissingOrFailed", "inactive") {
		t.Fatal("same-second successful retry did not resolve the backup rule")
	}
	waitForBackupNotification(t, ctx, received, "resolved", clearedAt, firingStart)
	t.Log("successful retry reached the receiver as resolution of that exact backup alert")

	// The server is reachable throughout. A real Postgres table lock makes the
	// collector's bounded count query time out, without replacing its metric
	// values or taking the scrape target offline. Its health signal must page.
	if !waitForQueryValue(ctx, client, "cloud_clicker_postgres_reachable", "1") ||
		!waitForQueryValue(ctx, client, "cloud_clicker_database_collection_success", "1") ||
		!waitForAlertState(ctx, client, "CloudClickerPostgresUnreachable", "inactive") {
		t.Fatal("actual Postgres collector did not establish a healthy baseline")
	}
	lock, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err := lock.ExecContext(ctx, "LOCK TABLE verification_dead_letters IN ACCESS EXCLUSIVE MODE"); err != nil {
		t.Fatal(err)
	}
	blockedAt := time.Now().UTC()
	if !waitForQueryValue(ctx, client, "cloud_clicker_postgres_reachable", "1") ||
		!waitForQueryValue(ctx, client, "cloud_clicker_database_collection_success", "0") ||
		!waitForQueryValue(ctx, client, `up{job="gameserver"}`, "1") ||
		!waitForAlertState(ctx, client, "CloudClickerPostgresUnreachable", "firing") {
		t.Fatal("real blocked database query did not fire its declared rule")
	}
	if elapsed := time.Since(blockedAt); elapsed < 2*time.Minute {
		t.Fatalf("database alert bypassed its two-minute lower bound: %s", elapsed)
	}
	databaseFiringStart := waitForOperationsNotification(t, ctx, received, "CloudClickerPostgresUnreachable", "gameserver", "firing", blockedAt, time.Time{})
	t.Log("real Postgres collection failure reached the receiver after the two-minute rule interval; ping and scrape stayed healthy")
	unblockedAt := time.Now().UTC()
	if err := lock.Rollback(); err != nil {
		t.Fatal(err)
	}
	if !waitForQueryValue(ctx, client, "cloud_clicker_database_collection_success", "1") ||
		!waitForAlertState(ctx, client, "CloudClickerPostgresUnreachable", "inactive") {
		t.Fatal("real Postgres collector did not recover after lock release")
	}
	waitForOperationsNotification(t, ctx, received, "CloudClickerPostgresUnreachable", "gameserver", "resolved", unblockedAt, databaseFiringStart)
	t.Log("lock release reached the configured receiver as resolution of that exact database alert")

	// Separately deny actual connections to only the declared disposable subject.
	// The control connection uses postgres, never the subject being disabled.
	adminURL := os.Getenv("TEST_ADMIN_DATABASE_URL")
	if adminURL == "" {
		t.Fatal("private connection-failure proof requires its separate control database")
	}
	admin, err := save.OpenPostgres(ctx, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	var subject string
	if err := database.QueryRowContext(ctx, "SELECT current_database()").Scan(&subject); err != nil || subject != "cloud_clicker_operations_test" {
		t.Fatal("connection-failure proof refuses a non-fixture database")
	}
	refusedAt := time.Now().UTC()
	restoreConnections := refuseOperationsDatabaseConnections(t, ctx, admin, subject)
	defer func() {
		if err := restoreConnections(); err != nil {
			t.Error(err)
		}
	}()
	if !waitForQueryValue(ctx, client, "cloud_clicker_postgres_reachable", "0") ||
		!waitForQueryValue(ctx, client, "cloud_clicker_database_collection_success", "0") ||
		!waitForQueryValue(ctx, client, `up{job="gameserver"}`, "1") ||
		!waitForAlertState(ctx, client, "CloudClickerPostgresUnreachable", "firing") {
		t.Fatal("actual refused database connections did not fire their rule")
	}
	if elapsed := time.Since(refusedAt); elapsed < 2*time.Minute {
		t.Fatalf("connection alert bypassed its two-minute lower bound: %s", elapsed)
	}
	connectionFiringStart := waitForOperationsNotification(t, ctx, received, "CloudClickerPostgresUnreachable", "gameserver", "firing", refusedAt, time.Time{})
	t.Log("actual refused Postgres connections reached the receiver after two minutes while metrics stayed reachable")
	restoredAt := time.Now().UTC()
	if err := restoreConnections(); err != nil {
		t.Fatal(err)
	}
	if !waitForQueryValue(ctx, client, "cloud_clicker_postgres_reachable", "1") ||
		!waitForQueryValue(ctx, client, "cloud_clicker_database_collection_success", "1") ||
		!waitForAlertState(ctx, client, "CloudClickerPostgresUnreachable", "inactive") {
		t.Fatal("real database pool did not reconnect after access restoration")
	}
	waitForOperationsNotification(t, ctx, received, "CloudClickerPostgresUnreachable", "gameserver", "resolved", restoredAt, connectionFiringStart)
	t.Log("automatic pool reconnection delivered resolution of that exact connection-failure alert")

	// Grow the actual transport dead-letter population through the current SQL
	// trigger, real Node encoder refusal and relay retry policy. An isolated
	// increase is insufficient; two consecutive scraped increases must page.
	deadLetters := newOperationsDeadLetterFixture(t, ctx, database, registry)
	deadLetters.publishHealthy(t, ctx)
	deadQuery := `cloud_clicker_dead_letters{queue="transport"}`
	if !waitForQueryValue(ctx, client, deadQuery, "0") ||
		!waitForAlertState(ctx, client, "CloudClickerDeadLetterGrowth", "inactive") {
		t.Fatal("actual transport dead letters did not establish a healthy baseline")
	}
	deadLetters.failOversized(t, ctx)
	if !waitForQueryValue(ctx, client, deadQuery, "1") {
		t.Fatal("first actual transport dead letter never reached Prometheus")
	}
	requireNoDeadLetterGrowthAlert(t, ctx, client, 45*time.Second)
	t.Log("one actual dead-letter increase stayed quiet over three evaluation intervals")
	growthAt := time.Now().UTC()
	for want := 2; want <= 4; want++ {
		deadLetters.failOversized(t, ctx)
		// Each new population must actually be scraped before producing the
		// next increase. A single batch of four rows cannot count as sustained
		// growth across the rule's consecutive collection intervals.
		if !waitForQueryValue(ctx, client, deadQuery, fmt.Sprint(want)) {
			t.Fatalf("transport dead-letter population %d never reached Prometheus", want)
		}
	}
	stoppedGrowingAt := time.Now().UTC()
	if !waitForAlertState(ctx, client, "CloudClickerDeadLetterGrowth", "firing") {
		t.Fatal("actual consecutive dead-letter growth never fired its rule")
	}
	deadLetterFiringStart := waitForOperationsNotification(t, ctx, received, "CloudClickerDeadLetterGrowth", "gameserver", "firing", growthAt, time.Time{}, map[string]string{"queue": "transport"})
	t.Log("actual relay-produced transport dead-letter growth reached the configured receiver")
	if !waitForAlertState(ctx, client, "CloudClickerDeadLetterGrowth", "inactive") {
		t.Fatal("stopping population growth did not resolve the dead-letter rule")
	}
	waitForOperationsNotification(t, ctx, received, "CloudClickerDeadLetterGrowth", "gameserver", "resolved", stoppedGrowingAt, deadLetterFiringStart, map[string]string{"queue": "transport"})
	deadLetters.publishHealthy(t, ctx)
	if !waitForQueryValue(ctx, client, deadQuery, "4") {
		t.Fatal("dead-letter resolution erased historical rows or healthy delivery added failures")
	}
	t.Log("growth stopped and the same alert resolved without deleting dead letters; valid delivery continued")
	if !waitForAlertState(ctx, client, "CloudClickerCleanupJobFailed", "inactive") {
		t.Fatal("re-reading retained failure kept the cleanup alert alive")
	}
	waitForOperationsNotification(t, ctx, cleanupReceived, "CloudClickerCleanupJobFailed", "node-exporter", "resolved", cleanupDetectedAt.Add(5*time.Minute), cleanupFiringStart)
	if !waitForQueryValue(ctx, client, failureQuery, "0") {
		t.Fatal("cleanup resolution depended on a process counter")
	}
	t.Log("that exact journal-driven cleanup alert resolved naturally; its persisted failure was not erased")

	// A receiver that rejects every notification must fail the delivery proof
	// on recorded failures, not merely time out.
	rejecting.Store(true)
	rejectCtx, rejectCancel := context.WithTimeout(ctx, 45*time.Second)
	defer rejectCancel()
	if _, err := VerifyAlertDelivery(rejectCtx, client, "http://alertmanager:9093", "http://localhost:8090/healthz"); err == nil ||
		!strings.Contains(err.Error(), "failed notification") {
		t.Fatalf("rejecting receiver was not detected as failed delivery: %v", err)
	}
}

func waitForBackupNotification(t *testing.T, ctx context.Context, received <-chan []byte, status string, after, firingStart time.Time) time.Time {
	t.Helper()
	return waitForOperationsNotification(t, ctx, received, "CloudClickerBackupMissingOrFailed", "node-exporter", status, after, firingStart)
}

func waitForOperationsNotification(t *testing.T, ctx context.Context, received <-chan []byte, name, job, status string, after, firingStart time.Time, wantLabels ...map[string]string) time.Time {
	t.Helper()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("%s %s notification missing: %v", name, status, ctx.Err())
		case body := <-received:
			var message struct {
				Alerts []struct {
					Status   string            `json:"status"`
					Labels   map[string]string `json:"labels"`
					StartsAt time.Time         `json:"startsAt"`
					EndsAt   time.Time         `json:"endsAt"`
				} `json:"alerts"`
			}
			if err := json.Unmarshal(body, &message); err != nil {
				t.Fatalf("invalid configured receiver notification: %v", err)
			}
			for _, alert := range message.Alerts {
				if alert.Labels["alertname"] != name || alert.Status != status {
					continue
				}
				if status == "firing" && alert.StartsAt.Before(after) {
					continue // stale startup-missing notification is not this failure
				}
				if status == "resolved" && (!alert.StartsAt.Equal(firingStart) || alert.EndsAt.Before(after)) {
					continue
				}
				if alert.StartsAt.IsZero() || alert.Labels["job"] != job ||
					alert.Labels["severity"] != "page" || alert.Labels["proof_nonce"] != "" {
					t.Fatalf("%s notification lost its actual producer identity: %+v", name, alert)
				}
				for _, labels := range wantLabels {
					for key, want := range labels {
						if alert.Labels[key] != want {
							t.Fatalf("%s notification label %s=%q, want %q", name, key, alert.Labels[key], want)
						}
					}
				}
				return alert.StartsAt
			}
		}
	}
}

func requireNoDeadLetterGrowthAlert(t *testing.T, ctx context.Context, client *http.Client, duration time.Duration) {
	t.Helper()
	end := time.Now().Add(duration)
	for time.Now().Before(end) {
		query := `ALERTS{alertname="CloudClickerDeadLetterGrowth",queue="transport"}`
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://prometheus:9090/api/v1/query?query="+url.QueryEscape(query), nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatalf("quiet-interval observation unavailable: %v", err)
		}
		var payload struct {
			Status string `json:"status"`
			Data   struct {
				Result []json.RawMessage `json:"result"`
			} `json:"data"`
		}
		decodeErr := json.NewDecoder(response.Body).Decode(&payload)
		_ = response.Body.Close()
		if decodeErr != nil || response.StatusCode != http.StatusOK || payload.Status != "success" || len(payload.Data.Result) != 0 {
			t.Fatalf("single-interval growth was not demonstrably quiet: HTTP=%d status=%q rows=%d error=%v", response.StatusCode, payload.Status, len(payload.Data.Result), decodeErr)
		}
		select {
		case <-ctx.Done():
			t.Fatalf("quiet-interval observation incomplete: %v", ctx.Err())
		case <-time.After(time.Second):
		}
	}
}

func waitForQueryValue(ctx context.Context, client *http.Client, query, want string) bool {
	for ctx.Err() == nil {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://prometheus:9090/api/v1/query?query="+url.QueryEscape(query), nil)
		if response, err := client.Do(request); err == nil {
			var payload struct {
				Status string `json:"status"`
				Data   struct {
					Result []struct {
						Value []any `json:"value"`
					} `json:"result"`
				} `json:"data"`
			}
			decodeErr := json.NewDecoder(response.Body).Decode(&payload)
			_ = response.Body.Close()
			if decodeErr == nil && response.StatusCode == http.StatusOK && payload.Status == "success" &&
				len(payload.Data.Result) == 1 && len(payload.Data.Result[0].Value) == 2 && payload.Data.Result[0].Value[1] == want {
				return true
			}
		}
		time.Sleep(time.Second)
	}
	return false
}

func waitForAlertState(ctx context.Context, client *http.Client, name, state string) bool {
	for ctx.Err() == nil {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://prometheus:9090/api/v1/alerts", nil)
		if response, err := client.Do(request); err == nil {
			var payload struct {
				Status string `json:"status"`
				Data   struct {
					Alerts []struct {
						Labels map[string]string `json:"labels"`
						State  string            `json:"state"`
					} `json:"alerts"`
				} `json:"data"`
			}
			decodeErr := json.NewDecoder(response.Body).Decode(&payload)
			_ = response.Body.Close()
			if decodeErr == nil && response.StatusCode == http.StatusOK && payload.Status == "success" {
				found := false
				for _, alert := range payload.Data.Alerts {
					if alert.Labels["alertname"] == name {
						found = true
						if alert.State == state {
							return true
						}
					}
				}
				if state == "inactive" && !found {
					return true
				}
			}
		}
		time.Sleep(time.Second)
	}
	return false
}

func waitForHTTP(ctx context.Context, client *http.Client, endpoint string) bool {
	for ctx.Err() == nil {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		response, err := client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode >= 200 && response.StatusCode < 300 {
				return true
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func waitForHealthyTargets(ctx context.Context, client *http.Client) (bool, string) {
	detail := "no response"
	for ctx.Err() == nil {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://prometheus:9090/api/v1/targets", nil)
		response, err := client.Do(request)
		if err == nil {
			var payload struct {
				Status string `json:"status"`
				Data   struct {
					Active []struct {
						Health string `json:"health"`
					} `json:"activeTargets"`
				} `json:"data"`
			}
			body, _ := io.ReadAll(response.Body)
			decodeErr := json.Unmarshal(body, &payload)
			_ = response.Body.Close()
			detail = string(body)
			if decodeErr == nil && payload.Status == "success" && len(payload.Data.Active) == 5 {
				healthy := true
				for _, target := range payload.Data.Active {
					healthy = healthy && target.Health == "up"
				}
				if healthy {
					return true, detail
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false, detail
}
