package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestObserveReleaseFloorAlertDeliveryRequiresFiringAndResolutionForAllSeven(t *testing.T) {
	var lock sync.Mutex
	notifications := 0
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/healthz":
			response.WriteHeader(http.StatusNoContent)
		case "/metrics":
			lock.Lock()
			count := notifications
			lock.Unlock()
			_, _ = fmt.Fprintf(response, "alertmanager_notifications_total %d\n", count)
		case "/api/v2/alerts":
			var alerts []struct {
				Labels map[string]string `json:"labels"`
				EndsAt string            `json:"endsAt"`
			}
			if json.NewDecoder(request.Body).Decode(&alerts) != nil || len(alerts) != len(ReleaseFloorAlertNames) {
				response.WriteHeader(http.StatusBadRequest)
				return
			}
			for index, alert := range alerts {
				if alert.Labels["alertname"] != ReleaseFloorAlertNames[index] || (posts == 0) != (alert.EndsAt == "") {
					response.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			lock.Lock()
			posts++
			notifications += len(alerts)
			lock.Unlock()
			response.WriteHeader(http.StatusAccepted)
		default:
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	clock := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	now := func() time.Time {
		clock = clock.Add(time.Second)
		return clock
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	manifest := "sha256:" + strings.Repeat("a", 64)
	result, err := ObserveReleaseFloorAlertDelivery(ctx, server.Client(), server.URL, server.URL+"/healthz", manifest, now)
	if err != nil || result.ManifestSHA256 != manifest || len(result.Alerts) != 7 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	encoded, _ := json.Marshal(result)
	if _, err := DecodeAlertDeliveryObservation(encoded); err != nil {
		t.Fatal(err)
	}
	result.Alerts[0].ResolvedDelivered = false
	if err := ValidateAlertDeliveryObservation(result); err == nil {
		t.Fatal("missing resolved delivery accepted")
	}
}

func TestDecodeAlertDeliveryRejectsUnknownSummary(t *testing.T) {
	start := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	alerts := make([]AlertObservation, len(ReleaseFloorAlertNames))
	for index, name := range ReleaseFloorAlertNames {
		alerts[index] = AlertObservation{Name: name, FiringDelivered: true, ResolvedDelivered: true}
	}
	result := AlertDeliveryObservation{SchemaVersion: 1, ManifestSHA256: "sha256:" + strings.Repeat("a", 64), StartedAt: start,
		CompletedAt: start.Add(time.Second), RuleFixturesPassed: true, Alerts: alerts, ObjectiveCompleted: true}
	data, _ := json.Marshal(result)
	data = append(data[:len(data)-1], []byte(`,"summary":"passed"}`)...)
	if _, err := DecodeAlertDeliveryObservation(data); err == nil {
		t.Fatal("unknown alert summary accepted")
	}
}

func TestObserveReleaseFloorAlertDeliveryRejectsSingleHealthNotification(t *testing.T) {
	notifications := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/healthz":
			response.WriteHeader(http.StatusNoContent)
		case "/metrics":
			_, _ = fmt.Fprintf(response, "alertmanager_notifications_total %d\n", notifications)
		case "/api/v2/alerts":
			notifications++
			response.WriteHeader(http.StatusAccepted)
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if _, err := ObserveReleaseFloorAlertDelivery(ctx, server.Client(), server.URL, server.URL+"/healthz",
		"sha256:"+strings.Repeat("a", 64), time.Now); err == nil {
		t.Fatal("one health notification accepted for seven alert families")
	}
}
