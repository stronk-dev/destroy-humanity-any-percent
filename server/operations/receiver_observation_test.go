package operations

import (
	"context"
	"strings"
	"testing"
)

func TestReceiverNoncePreservesEarlierRealNotifications(t *testing.T) {
	observed := make(chan []byte, 4)
	first := `{"alerts":[{"labels":{"alertname":"CloudClickerGameserverRestartLoop"},"status":"firing"}]}`
	second := `{"alerts":[{"labels":{"alertname":"CloudClickerStoragePressure"},"status":"firing"}]}`
	observed <- []byte(first)
	observed <- []byte(second)
	observed <- []byte(`{"alerts":[{"labels":{"alertname":"CloudClickerReceiverTest"},"annotations":{"proof_nonce":"actual-nonce"}}]}`)
	if err := waitForReceiverNonce(context.Background(), observed, "actual-nonce"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{first, second} {
		select {
		case actual := <-observed:
			if string(actual) != want {
				t.Fatal("preflight changed/reordered an earlier notification")
			}
		default:
			t.Fatal("preflight consumed an earlier real notification")
		}
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
	observed := make(chan []byte, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		observed <- []byte(`{"alertname":"first"}`)
		observed <- []byte(`{"alertname":"second"}`)
		observed <- []byte(`{"alertname":"CloudClickerReceiverTest","proof_nonce":"nonce"}`)
	}()
	if err := waitForReceiverNonce(context.Background(), observed, "nonce"); err == nil || !strings.Contains(err.Error(), "overflow") {
		t.Fatal("incomplete overflowing observation accepted")
	}
	<-done
}
