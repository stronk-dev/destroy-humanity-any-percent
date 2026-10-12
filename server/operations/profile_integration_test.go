package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPrivateOperationsProfileIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_OPERATIONS_INTEGRATION") != "1" {
		t.Skip("requires isolated operations Compose profile")
	}
	// The expanded population includes both backup delivery transitions at the
	// unchanged one-minute Alertmanager grouping interval, then its 45s refusal
	// window. This is a completion guard, not a relaxed alert/latency bound.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	received := make(chan []byte, 32)
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
	registry, err := NewRegistry(fakeDatabase{ping: errors.New("integration fixture has no database")})
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

	// A real composed cleanup failure must reach a firing rule through the
	// shipped scrape config: the zero series is scraped first, then one failure.
	failureQuery := `cloud_clicker_job_runs_total{job_name="credential_cleanup",result="failure"}`
	if !waitForQueryValue(ctx, client, failureQuery, "0") {
		t.Fatal("Prometheus never scraped the pre-created cleanup failure series")
	}
	if err := registry.ObserveJob("credential_cleanup", time.Now(), func() error { return errors.New("fixture cleanup failure") }); err == nil {
		t.Fatal("failed cleanup reported success")
	}
	if !waitForAlertState(ctx, client, "CloudClickerCleanupJobFailed", "firing") {
		t.Fatal("real cleanup failure never fired CloudClickerCleanupJobFailed")
	}

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
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("backup %s notification missing: %v", status, ctx.Err())
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
				if alert.Labels["alertname"] != "CloudClickerBackupMissingOrFailed" || alert.Status != status {
					continue
				}
				if status == "firing" && alert.StartsAt.Before(after) {
					continue // stale startup-missing notification is not this failure
				}
				if status == "resolved" && (!alert.StartsAt.Equal(firingStart) || alert.EndsAt.Before(after)) {
					continue
				}
				if alert.StartsAt.IsZero() || alert.Labels["job"] != "node-exporter" ||
					alert.Labels["severity"] != "page" || alert.Labels["proof_nonce"] != "" {
					t.Fatalf("backup notification lost its actual producer identity: %+v", alert)
				}
				return alert.StartsAt
			}
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
