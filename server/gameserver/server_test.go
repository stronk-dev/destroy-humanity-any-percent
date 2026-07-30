package gameserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeDatabase struct{ err error }

func (database fakeDatabase) PingContext(context.Context) error { return database.err }

const testConstantsHash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type fakeEpochs struct {
	hash   string
	err    error
	events *[]string
}

func (epochs fakeEpochs) Sync(context.Context) (string, error) {
	if epochs.events != nil {
		*epochs.events = append(*epochs.events, "epochs")
	}
	return epochs.hash, epochs.err
}

func syncedEpochs() fakeEpochs { return fakeEpochs{hash: testConstantsHash} }

type fakeRealtime struct {
	mu           sync.Mutex
	events       []string
	broadcasted  chan struct{}
	broadcastErr error
	timeout      time.Duration
}

func (realtime *fakeRealtime) Run() error            { realtime.record("run"); return nil }
func (realtime *fakeRealtime) Handler() http.Handler { return http.NotFoundHandler() }
func (realtime *fakeRealtime) BroadcastDrain(_ string, _ time.Time) error {
	realtime.record("broadcast")
	select {
	case <-realtime.broadcasted:
	default:
		close(realtime.broadcasted)
	}
	return realtime.broadcastErr
}
func (realtime *fakeRealtime) CloseForDrain() { realtime.record("close") }
func (realtime *fakeRealtime) Shutdown(context.Context) error {
	realtime.record("shutdown")
	return nil
}
func (realtime *fakeRealtime) DrainTimeout() time.Duration { return realtime.timeout }
func (realtime *fakeRealtime) record(event string) {
	realtime.mu.Lock()
	realtime.events = append(realtime.events, event)
	realtime.mu.Unlock()
}
func (realtime *fakeRealtime) snapshot() []string {
	realtime.mu.Lock()
	defer realtime.mu.Unlock()
	return append([]string(nil), realtime.events...)
}

type fakeRelay struct {
	mu      sync.Mutex
	results []int
	events  *[]string
}

func (relay *fakeRelay) Flush(context.Context) (int, error) {
	relay.mu.Lock()
	defer relay.mu.Unlock()
	if relay.events != nil {
		*relay.events = append(*relay.events, "flush")
	}
	if len(relay.results) == 0 {
		return 0, nil
	}
	result := relay.results[0]
	relay.results = relay.results[1:]
	return result, nil
}

func TestDrainOrdersAdmissionCommitFlushAndSocketClose(t *testing.T) {
	intentEntered := make(chan struct{})
	releaseIntent := make(chan struct{})
	api := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/v1/intents" {
			close(intentEntered)
			<-releaseIntent
			response.WriteHeader(http.StatusOK)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	})
	realtime := &fakeRealtime{broadcasted: make(chan struct{}), timeout: time.Second}
	relay := &fakeRelay{results: []int{2, 0}}
	server, err := New(fakeDatabase{}, api, realtime, relay, syncedEpochs(), testConstantsHash)
	if err != nil {
		t.Fatal(err)
	}
	handler := server.Handler()
	intentDone := make(chan struct{})
	go func() {
		defer close(intentDone)
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/intents", nil))
	}()
	<-intentEntered
	drainDone := make(chan error, 1)
	go func() { drainDone <- server.Drain(context.Background(), time.Now().UTC()) }()
	<-realtime.broadcasted

	rejected := httptest.NewRecorder()
	handler.ServeHTTP(rejected, httptest.NewRequest(http.MethodPost, "/api/v1/intents", nil))
	if rejected.Code != http.StatusServiceUnavailable || rejected.Body.String() != "{\"category\":\"server_draining\",\"detail\":\"retry_same_intent_id\"}\n" {
		t.Fatalf("draining rejection code=%d body=%s", rejected.Code, rejected.Body.String())
	}
	if events := realtime.snapshot(); len(events) != 1 || events[0] != "broadcast" {
		t.Fatalf("connections closed before in-flight intent: %v", events)
	}

	close(releaseIntent)
	<-intentDone
	if err := <-drainDone; err != nil {
		t.Fatal(err)
	}
	if events := realtime.snapshot(); len(events) != 3 || events[0] != "broadcast" || events[1] != "close" || events[2] != "shutdown" {
		t.Fatalf("drain events=%v", events)
	}
}

func TestHealthAndReadinessAreDistinct(t *testing.T) {
	realtime := &fakeRealtime{broadcasted: make(chan struct{}), timeout: time.Second}
	server, _ := New(fakeDatabase{err: errors.New("database unavailable")}, http.NotFoundHandler(), realtime, &fakeRelay{}, syncedEpochs(), testConstantsHash)
	handler := server.Handler()
	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	ready := httptest.NewRecorder()
	handler.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if health.Code != http.StatusNoContent || ready.Code != http.StatusServiceUnavailable {
		t.Fatalf("health=%d ready=%d", health.Code, ready.Code)
	}
}

func TestDrainDeadlineStillClosesSockets(t *testing.T) {
	intentEntered := make(chan struct{})
	blocked := make(chan struct{})
	api := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/v1/intents" {
			close(intentEntered)
			<-blocked
		}
		response.WriteHeader(http.StatusOK)
	})
	realtime := &fakeRealtime{broadcasted: make(chan struct{}), timeout: 20 * time.Millisecond}
	server, _ := New(fakeDatabase{}, api, realtime, &fakeRelay{}, syncedEpochs(), testConstantsHash)
	go server.Handler().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/intents", nil))
	<-intentEntered
	err := server.Drain(context.Background(), time.Now().UTC())
	close(blocked)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline error=%v", err)
	}
	if events := realtime.snapshot(); len(events) != 3 || events[0] != "broadcast" || events[1] != "close" || events[2] != "shutdown" {
		t.Fatalf("deadline drain events=%v", events)
	}
}

func TestDrainBroadcastFailureStillClosesSockets(t *testing.T) {
	realtime := &fakeRealtime{broadcasted: make(chan struct{}), broadcastErr: errors.New("broadcast unavailable"), timeout: time.Second}
	server, _ := New(fakeDatabase{}, http.NotFoundHandler(), realtime, &fakeRelay{}, syncedEpochs(), testConstantsHash)
	err := server.Drain(context.Background(), time.Now().UTC())
	if err == nil || !strings.Contains(err.Error(), "broadcast unavailable") {
		t.Fatalf("drain err=%v", err)
	}
	if events := realtime.snapshot(); len(events) != 3 || events[0] != "broadcast" || events[1] != "close" || events[2] != "shutdown" {
		t.Fatalf("broadcast failure events=%v", events)
	}
}

func TestReadinessCannotRiseAfterDrainStarts(t *testing.T) {
	realtime := &fakeRealtime{broadcasted: make(chan struct{}), timeout: time.Second}
	server, _ := New(fakeDatabase{}, http.NotFoundHandler(), realtime, &fakeRelay{}, syncedEpochs(), testConstantsHash)
	server.ready.Store(true)
	server.draining.Store(true)
	server.ready.Store(false)
	server.markReadyIfRunning()
	ready := httptest.NewRecorder()
	server.Handler().ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if ready.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready rose during drain: %d", ready.Code)
	}
}

func TestStartRunsRealtimeAndReceiptRelay(t *testing.T) {
	realtime := &fakeRealtime{broadcasted: make(chan struct{}), timeout: time.Second}
	relay := &fakeRelay{}
	events := []string{}
	epochs := fakeEpochs{hash: testConstantsHash, events: &events}
	server, _ := New(fakeDatabase{}, http.NotFoundHandler(), realtime, relay, epochs, testConstantsHash)
	ctx, cancel := context.WithCancel(context.Background())
	if err := server.Start(ctx); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case <-server.relayDone:
	case <-time.After(time.Second):
		t.Fatal("relay did not stop")
	}
	if events := realtime.snapshot(); len(events) != 1 || events[0] != "run" {
		t.Fatalf("start events=%v", events)
	}
	if len(events) != 1 || events[0] != "epochs" {
		t.Fatalf("startup order=%v", events)
	}
}

func TestStartFailsClosedBeforeRealtimeOnEpochMismatch(t *testing.T) {
	realtime := &fakeRealtime{broadcasted: make(chan struct{}), timeout: time.Second}
	server, err := New(fakeDatabase{}, http.NotFoundHandler(), realtime, &fakeRelay{}, fakeEpochs{hash: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}, testConstantsHash)
	if err != nil {
		t.Fatal(err)
	}
	if err := server.Start(context.Background()); !errors.Is(err, ErrInvalidServer) {
		t.Fatalf("start err=%v", err)
	}
	if events := realtime.snapshot(); len(events) != 0 {
		t.Fatalf("realtime started before epoch sync: %v", events)
	}
	ready := httptest.NewRecorder()
	server.Handler().ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if ready.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready=%d", ready.Code)
	}
}
