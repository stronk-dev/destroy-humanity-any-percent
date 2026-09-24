package operations

import (
	"context"
	"encoding/json"
	"errors"
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
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	received := make(chan []byte, 16)
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
	metrics := http.NewServeMux()
	metrics.Handle("/metrics", registry.Handler())
	start(":8080", metrics)
	var rejecting atomic.Bool
	receiver := http.NewServeMux()
	receiver.HandleFunc("/healthz", func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) })
	receiver.HandleFunc("/alerts", func(response http.ResponseWriter, request *http.Request) {
		data, _ := io.ReadAll(io.LimitReader(request.Body, 64<<10))
		if rejecting.Load() {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		select {
		case received <- data:
		default:
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

func waitForQueryValue(ctx context.Context, client *http.Client, query, want string) bool {
	for ctx.Err() == nil {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://prometheus:9090/api/v1/query?query="+url.QueryEscape(query), nil)
		if response, err := client.Do(request); err == nil {
			var payload struct {
				Data struct {
					Result []struct {
						Value []any `json:"value"`
					} `json:"result"`
				} `json:"data"`
			}
			decodeErr := json.NewDecoder(response.Body).Decode(&payload)
			_ = response.Body.Close()
			if decodeErr == nil && len(payload.Data.Result) == 1 && len(payload.Data.Result[0].Value) == 2 && payload.Data.Result[0].Value[1] == want {
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
				Data struct {
					Alerts []struct {
						Labels map[string]string `json:"labels"`
						State  string            `json:"state"`
					} `json:"alerts"`
				} `json:"data"`
			}
			decodeErr := json.NewDecoder(response.Body).Decode(&payload)
			_ = response.Body.Close()
			if decodeErr == nil {
				for _, alert := range payload.Data.Alerts {
					if alert.Labels["alertname"] == name && alert.State == state {
						return true
					}
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
