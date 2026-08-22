package operations

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPrivateOperationsProfileIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_OPERATIONS_INTEGRATION") != "1" {
		t.Skip("requires isolated operations Compose profile")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	metrics := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/metrics" {
			response.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(response, "# TYPE cloud_clicker_ready gauge\ncloud_clicker_ready 1\n# TYPE cloud_clicker_postgres_reachable gauge\ncloud_clicker_postgres_reachable 1\n# TYPE cloud_clicker_database_collection_success gauge\ncloud_clicker_database_collection_success 1\n")
	})
	start(":8080", metrics)
	receiver := http.NewServeMux()
	receiver.HandleFunc("/healthz", func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) })
	receiver.HandleFunc("/alerts", func(response http.ResponseWriter, request *http.Request) {
		data, _ := io.ReadAll(io.LimitReader(request.Body, 64<<10))
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
	nonce, err := VerifyAlertDelivery(ctx, client, "http://alertmanager:9093", "http://localhost:8090/healthz")
	if err != nil {
		t.Fatalf("receiver delivery proof failed: %v", err)
	}
	for {
		select {
		case body := <-received:
			text := strings.ToLower(string(body))
			for _, forbidden := range []string{"account_id", "founder_id", "recovery_code", "client_ip"} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("private field delivered: %s", forbidden)
				}
			}
			if strings.Contains(text, "cloudclickerreceivertest") && strings.Contains(text, nonce) {
				return
			}
		case <-ctx.Done():
			t.Fatal("Alertmanager accepted but did not deliver the synthetic alert")
		}
	}
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
