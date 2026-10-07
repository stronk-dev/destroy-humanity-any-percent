package operations

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNativePrometheusCleanupIntegration(t *testing.T) {
	prometheusBinary := requireNativeOperationsBinary(t, "CLOUD_CLICKER_PROMETHEUS_TEST_BINARY", "prometheus, version 3.12.0 (", "test-operations-native")
	alertmanagerBinary := requireNativeOperationsBinary(t, "CLOUD_CLICKER_ALERTMANAGER_TEST_BINARY", "alertmanager, version 0.32.1 (", "test-operations-native")
	// The actual five-minute increase window and one-minute notification
	// grouping are unchanged. This is a completion guard, not a latency claim.
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	t.Cleanup(cancel)
	started := time.Now()
	deliveries := make(chan nativeAlertMessage, 32)
	receiver := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/alerts" || request.Method != http.MethodPost {
			response.WriteHeader(http.StatusNotFound)
			return
		}
		var message nativeAlertMessage
		if err := json.NewDecoder(http.MaxBytesReader(response, request.Body, 64<<10)).Decode(&message); err != nil {
			t.Errorf("real rule notification: %v", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		select {
		case deliveries <- message:
			response.WriteHeader(http.StatusNoContent)
		default:
			t.Error("notification observation overflow; evidence incomplete")
			response.WriteHeader(http.StatusInternalServerError)
		}
	}))
	t.Cleanup(receiver.Close)
	alertmanager := startNativeAlertmanager(t, ctx, alertmanagerBinary, receiver.URL)
	registry, err := NewRegistry(fakeDatabase{ping: errors.New("no database in native metric fixture")})
	if err != nil {
		t.Fatal(err)
	}
	registry.SetReady(true)
	metrics := httptest.NewServer(registry.Handler())
	t.Cleanup(metrics.Close)
	// These two services are deliberately unavailable, not simulated healthy
	// Caddy/node-exporter targets. Their network/privacy acceptance stays open.
	unavailable := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(unavailable.Close)
	prometheus := startNativePrometheus(t, ctx, prometheusBinary, metrics.URL, alertmanager, unavailable.URL)
	client := &http.Client{Timeout: 2 * time.Second}
	query := `cloud_clicker_job_runs_total{job_name="credential_cleanup",result="failure"}`
	for {
		value, found := nativePrometheusQuery(t, ctx, client, prometheus, query)
		if found {
			if value != "0" {
				t.Fatalf("first failure baseline is not zero: %q", value)
			}
			break
		}
		waitNativeOperationsPoll(t, ctx)
	}
	t.Logf("actual production zero series scraped after %s", time.Since(started).Round(time.Millisecond))
	failure := errors.New("native fixture cleanup failure")
	if err := registry.ObserveJob("credential_cleanup", time.Now(), func() error { return failure }); !errors.Is(err, failure) {
		t.Fatalf("actual failed job lost its failure: %v", err)
	}
	for {
		value, found := nativePrometheusQuery(t, ctx, client, prometheus, query)
		if found && value == "1" {
			break
		}
		waitNativeOperationsPoll(t, ctx)
	}
	t.Logf("first failed job scraped after %s", time.Since(started).Round(time.Millisecond))
	waitNativeCleanupNotification(t, ctx, deliveries, "firing")
	t.Logf("actual rule's firing notification received after %s", time.Since(started).Round(time.Millisecond))
	// No direct alert API injection, synthetic resolution, shortened rule or
	// reset counter. The real rule must age out the single failure itself.
	waitNativeCleanupNotification(t, ctx, deliveries, "resolved")
	value, found := nativePrometheusQuery(t, ctx, client, prometheus, query)
	if !found || value != "1" {
		t.Fatalf("resolution changed the monotonic process counter: %q found=%v", value, found)
	}
	t.Logf("actual rule's resolution received with failure counter still one after %s", time.Since(started).Round(time.Millisecond))
}

func startNativePrometheus(t *testing.T, ctx context.Context, binary, metricsURL, alertmanagerURL, unavailableURL string) string {
	t.Helper()
	configuration, err := os.ReadFile(filepath.Join("..", "..", "deployment", "operations", "prometheus.yml"))
	if err != nil {
		t.Fatal(err)
	}
	rules, err := filepath.Abs(filepath.Join("..", "..", "deployment", "operations", "cloud-clicker-alerts.yml"))
	if err != nil {
		t.Fatal(err)
	}
	address := nativeOperationsAddress(t)
	replacements := map[string]struct {
		count int
		value string
	}{
		"gameserver:8080":    {1, strings.TrimPrefix(metricsURL, "http://")},
		"alertmanager:9093":  {2, strings.TrimPrefix(alertmanagerURL, "http://")},
		"caddy:2020":         {1, strings.TrimPrefix(unavailableURL, "http://")},
		"node-exporter:9100": {1, strings.TrimPrefix(unavailableURL, "http://")},
		"localhost:9090":     {1, address},
		"/etc/prometheus/cloud-clicker-alerts.yml": {1, rules},
	}
	text := string(configuration)
	for from, replacement := range replacements {
		if strings.Count(text, from) != replacement.count {
			t.Fatalf("declared operations scrape configuration changed at %q", from)
		}
		text = strings.ReplaceAll(text, from, replacement.value)
	}
	root := t.TempDir()
	configPath := filepath.Join(root, "prometheus.yml")
	if err := os.WriteFile(configPath, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return startNativeOperationsProcess(t, ctx, "Prometheus", binary, root, address,
		"--config.file="+configPath, "--storage.tsdb.path="+filepath.Join(root, "data"), "--web.listen-address="+address)
}

func nativePrometheusQuery(t *testing.T, ctx context.Context, client *http.Client, endpoint, query string) (string, bool) {
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/api/v1/query?query="+url.QueryEscape(query), nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  []json.RawMessage `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil || response.StatusCode != http.StatusOK || payload.Status != "success" || len(payload.Data.Result) > 1 {
		t.Fatalf("real Prometheus query failed: HTTP=%d status=%q rows=%d error=%v", response.StatusCode, payload.Status, len(payload.Data.Result), err)
	}
	if len(payload.Data.Result) == 0 {
		return "", false
	}
	row := payload.Data.Result[0]
	if row.Metric["job"] != "gameserver" || row.Metric["job_name"] != "credential_cleanup" || row.Metric["result"] != "failure" || row.Metric["exported_job"] != "" || len(row.Value) != 2 {
		t.Fatalf("scraped production metric lost its bounded job labels: %+v", row)
	}
	var value string
	if json.Unmarshal(row.Value[1], &value) != nil {
		t.Fatal("invalid Prometheus sample value")
	}
	return value, true
}

func waitNativeCleanupNotification(t *testing.T, ctx context.Context, deliveries <-chan nativeAlertMessage, status string) {
	t.Helper()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("actual cleanup %s notification missing: %v", status, ctx.Err())
		case message := <-deliveries:
			for _, alert := range message.Alerts {
				if alert.Labels["alertname"] != "CloudClickerCleanupJobFailed" {
					continue
				}
				if alert.Labels["job_name"] != "credential_cleanup" || alert.Labels["result"] != "failure" || alert.Labels["job"] != "gameserver" || alert.Labels["proof_nonce"] != "" {
					t.Fatalf("cleanup payload is not from the actual metric/rule path: %+v", alert)
				}
				if alert.Status == status {
					return
				}
			}
		}
	}
}

func waitNativeOperationsPoll(t *testing.T, ctx context.Context) {
	t.Helper()
	select {
	case <-ctx.Done():
		t.Fatalf("real operations observation incomplete: %v", ctx.Err())
	case <-time.After(time.Second):
	}
}
