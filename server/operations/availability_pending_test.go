package operations

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type availabilityRoundTrip func(*http.Request) (*http.Response, error)

func (run availabilityRoundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return run(request)
}

func TestAvailabilityPendingBindsProducerAndCompleteObservation(t *testing.T) {
	after := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	for _, job := range []string{"caddy", "gameserver"} {
		t.Run(job, func(t *testing.T) {
			payload := func(state, producer string, active time.Time) string {
				data, err := json.Marshal(map[string]any{"status": "success", "data": map[string]any{"alerts": []any{
					map[string]any{"labels": map[string]string{"alertname": "CloudClickerPublicEndpointOrReadinessUnavailable", "job": producer}, "state": state, "activeAt": active},
				}}})
				if err != nil {
					t.Fatal(err)
				}
				return string(data)
			}
			valid := payload("pending", job, after)
			for _, subject := range []struct {
				name, body string
				status     int
				valid      bool
			}{
				{"healthy", valid, 200, true},
				{"premature firing", payload("firing", job, after), 200, false},
				{"zero start", payload("pending", job, time.Time{}), 200, false},
				{"stale start", payload("pending", job, after.Add(-time.Second)), 200, false},
				{"HTTP failure", valid, 503, false},
				{"query failure", strings.Replace(valid, `"success"`, `"error"`, 1), 200, false},
				{"truncated", valid[:len(valid)-1], 200, false},
				{"extra JSON", valid + `{}`, 200, false},
			} {
				t.Run(subject.name, func(t *testing.T) {
					client := &http.Client{Transport: availabilityRoundTrip(func(request *http.Request) (*http.Response, error) {
						if request.URL.String() != "http://prometheus:9090/api/v1/alerts" {
							t.Fatal("pending observer queried a different source")
						}
						return &http.Response{StatusCode: subject.status, Body: io.NopCloser(strings.NewReader(subject.body))}, nil
					})}
					at, err := waitForAvailabilityPending(context.Background(), client, job, after)
					if subject.valid {
						if err != nil || !at.Equal(after) {
							t.Fatalf("healthy pending population refused: at=%v error=%v", at, err)
						}
					} else if !errors.Is(err, ErrInvalid) || !at.IsZero() {
						t.Fatalf("incomplete pending population admitted: at=%v error=%v", at, err)
					}
				})
			}
			for _, producer := range []string{"other", ""} {
				ctx, cancel := context.WithCancel(context.Background())
				client := &http.Client{Transport: availabilityRoundTrip(func(*http.Request) (*http.Response, error) {
					cancel()
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload("pending", producer, after)))}, nil
				})}
				at, err := waitForAvailabilityPending(ctx, client, job, after)
				cancel()
				if !errors.Is(err, context.Canceled) || !at.IsZero() {
					t.Fatal("another producer substituted for the requested pending population")
				}
			}
			t.Run("canceled matching response", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				client := &http.Client{Transport: availabilityRoundTrip(func(*http.Request) (*http.Response, error) {
					cancel()
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(valid))}, nil
				})}
				at, err := waitForAvailabilityPending(ctx, client, job, after)
				if !errors.Is(err, context.Canceled) || !at.IsZero() {
					t.Fatalf("canceled matching observation returned success: at=%v error=%v", at, err)
				}
			})
		})
	}
}
