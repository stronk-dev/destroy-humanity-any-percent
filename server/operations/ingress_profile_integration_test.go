package operations

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func runIngressProfile(t *testing.T, ctx context.Context, client *http.Client, received <-chan []byte) {
	t.Helper()
	const alert = "CloudClickerPublicEndpointOrReadinessUnavailable"
	const endpoint = "http://caddy:8081/readyz"
	if !waitForHTTP(ctx, client, endpoint) || !waitForQueryValue(ctx, client, `up{job="caddy"}`, "1") ||
		!waitForQueryValue(ctx, client, "cloud_clicker_ready", "1") || !waitForAlertState(ctx, client, alert, "inactive") {
		t.Fatal("proxy route/scrape had no healthy baseline")
	}
	failedAt := time.Now().UTC()
	// The native owner interprets only these two fixed test requests. It must
	// validate the project/service labels before stopping/starting actual Caddy.
	t.Log("native_caddy_stop")
	for {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		response, err := client.Do(request)
		if err != nil {
			break
		}
		_ = response.Body.Close()
		select {
		case <-ctx.Done():
			t.Fatal("actual proxy route never became unavailable")
		case <-time.After(time.Second):
		}
	}
	if !waitForQueryValue(ctx, client, `up{job="caddy"}`, "0") ||
		!waitForQueryValue(ctx, client, `up{job="gameserver"}`, "1") || !waitForQueryValue(ctx, client, "cloud_clicker_ready", "1") {
		t.Fatal("actual Caddy outage did not isolate the proxy arm")
	}
	activeAt := waitForIngressPending(t, ctx, client, failedAt)
	quietUntil := activeAt.Add(4*time.Minute + 45*time.Second)
	if remaining := time.Until(quietUntil); remaining <= 0 {
		t.Fatal("lower-bound observation began too late")
	} else {
		requireNoOperationsAlert(t, ctx, client, `ALERTS{alertname="CloudClickerPublicEndpointOrReadinessUnavailable",job="caddy",alertstate="firing"}`, remaining)
	}
	if !waitForAlertState(ctx, client, alert, "firing") || time.Now().Before(activeAt.Add(5*time.Minute)) {
		t.Fatal("actual outage did not honor the five-minute firing bound")
	}
	firingStart := waitForOperationsNotification(t, ctx, received, alert, "caddy", "firing", failedAt, time.Time{})
	t.Log("actual stopped proxy reached the receiver after the unchanged five-minute rule")
	clearedAt := time.Now().UTC()
	t.Log("native_caddy_start")
	if !waitForHTTP(ctx, client, endpoint) || !waitForQueryValue(ctx, client, `up{job="caddy"}`, "1") ||
		!waitForAlertState(ctx, client, alert, "inactive") {
		t.Fatal("actual restarted Caddy did not recover route/scrape availability")
	}
	waitForOperationsNotification(t, ctx, received, alert, "caddy", "resolved", clearedAt, firingStart)
	t.Log("same proxy alert resolved after restarting the actual Caddy container")
}

func waitForIngressPending(t *testing.T, ctx context.Context, client *http.Client, after time.Time) time.Time {
	t.Helper()
	for ctx.Err() == nil {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://prometheus:9090/api/v1/alerts", nil)
		response, err := client.Do(request)
		if err != nil {
			t.Fatal("pending-state observation unavailable")
		}
		var payload struct {
			Status string `json:"status"`
			Data   struct {
				Alerts []struct {
					Labels   map[string]string `json:"labels"`
					State    string            `json:"state"`
					ActiveAt time.Time         `json:"activeAt"`
				} `json:"alerts"`
			} `json:"data"`
		}
		err = json.NewDecoder(response.Body).Decode(&payload)
		_ = response.Body.Close()
		if err != nil || response.StatusCode != http.StatusOK || payload.Status != "success" {
			t.Fatal("pending-state observation incomplete")
		}
		for _, alert := range payload.Data.Alerts {
			if alert.Labels["alertname"] != "CloudClickerPublicEndpointOrReadinessUnavailable" || alert.Labels["job"] != "caddy" {
				continue
			}
			if alert.State != "pending" || alert.ActiveAt.IsZero() || alert.ActiveAt.Before(after) {
				t.Fatal("outage entered firing before its declared pending population")
			}
			return alert.ActiveAt
		}
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
		}
	}
	t.Fatal("actual outage never reached its pending state")
	return time.Time{}
}
