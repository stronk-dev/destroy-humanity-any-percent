package gameserver

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/replayverify"
	"cloud-clicker/server/transport"
)

type fakeOperationsSurface struct {
	requests int
}

func (surface *fakeOperationsSurface) Handler() http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		surface.requests++
		_, _ = response.Write([]byte("private-metrics"))
	})
}

func (surface *fakeOperationsSurface) HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-Operations-Observed", "true")
		next.ServeHTTP(response, request)
	})
}

func (*fakeOperationsSurface) SetReady(bool) {}

func TestOperationsSurfaceIsAttachedAtPrivateMetricsRoute(t *testing.T) {
	realtime := &fakeRealtime{broadcasted: make(chan struct{}), timeout: time.Second}
	server, err := New(fakeDatabase{}, http.NotFoundHandler(), realtime, &fakeRelay{}, syncedEpochs(), testConstantsHash)
	if err != nil {
		t.Fatal(err)
	}
	surface := &fakeOperationsSurface{}
	if err := server.AttachOperations(surface); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if response.Code != http.StatusOK || response.Body.String() != "private-metrics" || response.Header().Get("X-Operations-Observed") != "true" || surface.requests != 1 {
		t.Fatalf("operations route not composed: status=%d body=%q header=%q requests=%d", response.Code, response.Body.String(), response.Header().Get("X-Operations-Observed"), surface.requests)
	}
}

func TestInvariantLogsExcludeIdentifiersAndUnboundedDetail(t *testing.T) {
	var output bytes.Buffer
	sink := logInvariantSink{logger: slog.New(slog.NewJSONHandler(&output, nil))}
	sink.ReportRelayInvariant(transport.RelayInvariant{Kind: "player_message_dead_letter", FounderID: "0198-private-founder", Detail: "credential=private"})
	sink.ReportVerificationInvariant(replayverify.VerificationInvariant{Kind: "verification_poison_dead_letter", StreamID: "0198-private-stream", RunSeq: 42, Detail: "payload-private"})
	text := output.String()
	for _, want := range []string{"player_message_dead_letter", "verification_poison_dead_letter"} {
		if !strings.Contains(text, want) {
			t.Fatalf("bounded invariant kind missing: %s", text)
		}
	}
	for _, forbidden := range []string{"private-founder", "private-stream", "credential=private", "payload-private", "founder_id", "stream_id", "run_seq", "detail"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("private invariant field logged: %q in %s", forbidden, text)
		}
	}
}
