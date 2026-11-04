package operations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestReceiverNoncePreservesEarlierRealNotifications(t *testing.T) {
	observed := make(chan []byte, 32)
	cleanup := make(chan []byte, 32)
	preflight := make(chan []byte, 1)
	handler := operationsReceiver(&atomic.Bool{}, observed, cleanup, preflight, func(err error) { t.Errorf("unexpected receiver failure: %v", err) })
	first := `{"alerts":[{"labels":{"alertname":"CloudClickerGameserverRestartLoop"},"status":"firing"}]}`
	second := `{"alerts":[{"labels":{"alertname":"CloudClickerStoragePressure"},"status":"firing"}]}`
	later := `{"alerts":[{"labels":{"alertname":"CloudClickerGameserverRestartLoop"},"status":"resolved"}]}`
	for _, body := range []string{first, second, `{"alerts":[{"labels":{"alertname":"CloudClickerReceiverTest"},"annotations":{"proof_nonce":"actual-nonce"}}]}`, later} {
		response := httptest.NewRecorder()
		handler(response, httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(body)))
		if response.Code != http.StatusNoContent {
			t.Fatalf("receiver rejected valid notification: status=%d", response.Code)
		}
	}
	if len(observed) != 3 || len(preflight) != 1 || len(cleanup) != 0 {
		t.Fatal("receiver failed to separate its control from unchanged real-alert bytes")
	}
	if err := waitForReceiverNonce(context.Background(), preflight, "actual-nonce"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{first, second, later} {
		select {
		case actual := <-observed:
			if string(actual) != want {
				t.Fatal("preflight changed/reordered a real notification")
			}
		default:
			t.Fatal("preflight consumed an earlier real notification")
		}
	}
	if len(preflight) != 0 || len(observed) != 0 {
		t.Fatal("receiver duplicated a control or real notification")
	}
}

func TestReceiverNonceFailsClosedOnIncompleteOrUnsafeObservation(t *testing.T) {
	for _, field := range []string{"account_id", "founder_id", "recovery_code", "client_ip"} {
		observed := make(chan []byte, 1)
		observed <- []byte(field)
		if err := waitForReceiverNonce(context.Background(), observed, "nonce"); err == nil || !strings.Contains(err.Error(), field) {
			t.Fatal("preflight accepted a forbidden private field")
		}
	}
	closed := make(chan []byte, 1)
	close(closed)
	if err := waitForReceiverNonce(context.Background(), closed, "nonce"); err == nil {
		t.Fatal("closed notification stream passed preflight")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitForReceiverNonce(canceled, make(chan []byte, 1), "nonce"); err == nil {
		t.Fatal("incomplete notification observation passed preflight")
	}
	if err := waitForReceiverNonce(context.Background(), make(chan []byte, 1), ""); err == nil {
		t.Fatal("missing nonce passed preflight")
	}
	if err := waitForReceiverNonce(context.Background(), make(chan []byte), "nonce"); err == nil {
		t.Fatal("unbuffered observation channel accepted")
	}
}

func TestReceiverNonceRefusesBacklogOverflow(t *testing.T) {
	for _, stream := range []string{"real", "cleanup", "preflight"} {
		t.Run(stream, func(t *testing.T) {
			observed, cleanup, preflight := make(chan []byte, 32), make(chan []byte, 32), make(chan []byte, 1)
			full := observed
			body := `{"alerts":[{"labels":{"alertname":"CloudClickerStoragePressure"}}]}`
			switch stream {
			case "cleanup":
				full = cleanup
				body = `{"alerts":[{"labels":{"alertname":"CloudClickerCleanupJobFailed"}}]}`
			case "preflight":
				full = preflight
				body = `{"alerts":[{"labels":{"alertname":"CloudClickerReceiverTest"},"annotations":{"proof_nonce":"nonce"}}]}`
			}
			for len(full) < cap(full) {
				full <- []byte(body)
			}
			var failures []error
			handler := operationsReceiver(&atomic.Bool{}, observed, cleanup, preflight, func(err error) { failures = append(failures, err) })
			response := httptest.NewRecorder()
			handler(response, httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(body)))
			if response.Code != http.StatusInternalServerError || len(failures) != 1 || !strings.Contains(failures[0].Error(), "overflow") || len(full) != cap(full) {
				t.Fatal("overflow failed to preserve the full bounded observation and report incomplete evidence")
			}
			if stream != "real" && len(observed) != 0 {
				t.Fatal("refused control/cleanup overflow appended a real observation")
			}
		})
	}
}

func TestReceiverPreflightCannotConsumeRealAlertCapacity(t *testing.T) {
	observed, cleanup, preflight := make(chan []byte, 32), make(chan []byte, 32), make(chan []byte, 1)
	for len(observed) < cap(observed) {
		observed <- []byte(`{"alerts":[{"labels":{"alertname":"CloudClickerStoragePressure"}}]}`)
	}
	handler := operationsReceiver(&atomic.Bool{}, observed, cleanup, preflight, func(err error) { t.Errorf("unexpected receiver failure: %v", err) })
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(`{"alerts":[{"labels":{"alertname":"CloudClickerReceiverTest"},"annotations":{"proof_nonce":"nonce"}}]}`)))
	if response.Code != http.StatusNoContent || len(observed) != 32 {
		t.Fatal("delivery control displaced or consumed bounded real alerts")
	}
	if err := waitForReceiverNonce(context.Background(), preflight, "nonce"); err != nil || len(observed) != 32 {
		t.Fatal("preflight changed the full real-alert queue")
	}
}

func TestOperationsReceiverRefusesInvalidObservation(t *testing.T) {
	for _, body := range []string{
		`not JSON`, `null`, `{}`, `{"alerts":[]}`, `{"alerts":[{"labels":{}}]}`,
		`{"alerts":[{"labels":{"alertname":"CloudClickerReceiverTest"}}]}`,
		`{"alerts":[{"labels":{"alertname":"CloudClickerStoragePressure"}},{"labels":{"alertname":"CloudClickerReceiverTest"},"annotations":{"proof_nonce":"nonce"}}]}`,
		`{"alerts":[{"labels":{"alertname":"CloudClickerStoragePressure","founder_id":"fictional-private-marker"}}]}`,
	} {
		t.Run(body, func(t *testing.T) {
			observed, cleanup, preflight := make(chan []byte, 32), make(chan []byte, 32), make(chan []byte, 1)
			var failures []error
			handler := operationsReceiver(&atomic.Bool{}, observed, cleanup, preflight, func(err error) { failures = append(failures, err) })
			response := httptest.NewRecorder()
			handler(response, httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(body)))
			if response.Code != http.StatusInternalServerError || len(failures) != 1 || len(observed)+len(cleanup)+len(preflight) != 0 {
				t.Fatal("invalid/mixed/private notification entered an evidence stream")
			}
			if strings.Contains(failures[0].Error(), "fictional-private-marker") {
				t.Fatal("diagnostic disclosed notification bytes")
			}
		})
	}
}

func TestReceiverNonceRequiresExactStructuredIdentity(t *testing.T) {
	observed := make(chan []byte, 1)
	observed <- []byte(`{"alerts":[{"labels":{"alertname":"CloudClickerReceiverTest"},"annotations":{"proof_nonce":"nonce-suffix"}}]}`)
	close(observed)
	if err := waitForReceiverNonce(context.Background(), observed, "nonce"); err == nil {
		t.Fatal("nonce prefix masqueraded as the exact delivery identity")
	}
	foreign := make(chan []byte, 1)
	foreign <- []byte(`{"alerts":[{"labels":{"alertname":"CloudClickerStoragePressure"},"annotations":{"proof_nonce":"nonce"}}]}`)
	if err := waitForReceiverNonce(context.Background(), foreign, "nonce"); err == nil {
		t.Fatal("real alert masqueraded as a preflight control")
	}
}

func TestReceiverNonceRefusesAlreadyCanceledReadyControl(t *testing.T) {
	// Both select arms are ready. Cancellation must win without consuming the
	// control, rather than occasionally reporting a valid incomplete observation.
	for attempt := 0; attempt < 32; attempt++ {
		observed := make(chan []byte, 1)
		observed <- []byte(`{"alerts":[{"labels":{"alertname":"CloudClickerReceiverTest"},"annotations":{"proof_nonce":"nonce"}}]}`)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := waitForReceiverNonce(ctx, observed, "nonce"); err != context.Canceled || len(observed) != 1 {
			t.Fatalf("canceled observation accepted/consumed ready control: error=%v queued=%d", err, len(observed))
		}
	}
}

func TestOperationsReceiverPreservesCleanupMirror(t *testing.T) {
	observed, cleanup, preflight := make(chan []byte, 32), make(chan []byte, 32), make(chan []byte, 1)
	handler := operationsReceiver(&atomic.Bool{}, observed, cleanup, preflight, func(err error) { t.Errorf("unexpected receiver failure: %v", err) })
	body := `{"alerts":[{"labels":{"alertname":"CloudClickerCleanupJobFailed"},"annotations":{"summary":"CloudClickerReceiverTest nonce is only ordinary text here"}}]}`
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(body)))
	if response.Code != http.StatusNoContent || len(observed) != 1 || len(cleanup) != 1 || len(preflight) != 0 {
		t.Fatal("real cleanup alert or control-looking text was misrouted")
	}
	if string(<-observed) != body || string(<-cleanup) != body {
		t.Fatal("receiver changed the real alert or cleanup mirror bytes")
	}
}

func TestOperationsReceiverRetainsRejectionAndBodyLimit(t *testing.T) {
	observed, cleanup, preflight := make(chan []byte, 32), make(chan []byte, 32), make(chan []byte, 1)
	var rejecting atomic.Bool
	var failures []error
	handler := operationsReceiver(&rejecting, observed, cleanup, preflight, func(err error) { failures = append(failures, err) })
	rejecting.Store(true)
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(`{"alerts":[{"labels":{"alertname":"CloudClickerStoragePressure"}}]}`)))
	if response.Code != http.StatusInternalServerError || len(failures) != 0 || len(observed)+len(cleanup)+len(preflight) != 0 {
		t.Fatal("intentional receiver rejection produced accepted evidence")
	}
	rejecting.Store(false)
	response = httptest.NewRecorder()
	handler(response, httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(strings.Repeat("x", (64<<10)+1))))
	if response.Code != http.StatusBadRequest || len(failures) != 1 || len(observed)+len(cleanup)+len(preflight) != 0 {
		t.Fatal("oversized notification bypassed the existing body bound")
	}
}
