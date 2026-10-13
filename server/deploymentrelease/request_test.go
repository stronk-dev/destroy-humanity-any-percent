package deploymentrelease

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/deploymentconfig"
)

func TestReadinessObservationRefusesMalformedEndpointWithoutPanic(t *testing.T) {
	for _, endpoint := range []string{"%", "http://[broken", "https://example.test:bad/readyz", "https://example.test/ready\nz"} {
		t.Run(endpoint, func(t *testing.T) {
			defer func() {
				if recover() != nil {
					t.Error("malformed readiness endpoint panicked instead of refusing evidence")
				}
			}()
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				t.Error("malformed endpoint reached HTTP transport")
				return &http.Response{StatusCode: http.StatusServiceUnavailable,
					Header: http.Header{deploymentconfig.DrainReadinessHeader: []string{deploymentconfig.DrainReadinessValue}},
					Body:   io.NopCloser(strings.NewReader(""))}, nil
			})}
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			if waitHTTPState(ctx, client, endpoint, false) {
				t.Error("malformed endpoint forged readiness-down evidence")
			}
		})
	}
}

func TestPreflightRefusesMalformedReceiverBeforeRuntimeCommands(t *testing.T) {
	for _, endpoint := range []string{"%", "http://[broken", "https://example.test:bad/health", "https://example.test/health\nz"} {
		t.Run(endpoint, func(t *testing.T) {
			defer func() {
				if recover() != nil {
					t.Error("malformed receiver endpoint panicked during preflight")
				}
			}()
			runner := &commandFixture{}
			runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
			runtime.ReceiverHealthURL = endpoint
			if err := runtime.Preflight(context.Background(), dockerFixtureBundle(t)); !errors.Is(err, ErrInvalid) {
				t.Errorf("malformed receiver not rejected as invalid configuration: %v", err)
			}
			if len(runner.calls) != 0 {
				t.Error("preflight ran runtime commands before rejecting a malformed receiver")
			}
		})
	}
}
