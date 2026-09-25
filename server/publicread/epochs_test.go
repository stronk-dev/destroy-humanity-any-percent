package publicread

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/httpapi"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/publicapi"
)

type fakeEpochs struct {
	rows  []leaderboard.PublicEpoch // newest-first
	err   error
	calls [][2]int64
}

func (fake *fakeEpochs) PublicEpochPage(_ context.Context, before int64, limit int) ([]leaderboard.PublicEpoch, bool, error) {
	fake.calls = append(fake.calls, [2]int64{before, int64(limit)})
	if fake.err != nil {
		return nil, false, fake.err
	}
	page := []leaderboard.PublicEpoch{}
	for _, row := range fake.rows {
		if before == 0 || row.ID < before {
			page = append(page, row)
		}
	}
	if len(page) > limit {
		return page[:limit], true, nil
	}
	return page, false, nil
}

func hashOf(fill string) string { return "sha256:" + strings.Repeat(fill, 64) }

func epochRows() []leaderboard.PublicEpoch {
	started := time.Date(2026, 7, 29, 12, 0, 0, 123_456_000, time.UTC)
	ended := started.Add(time.Hour)
	rows := []leaderboard.PublicEpoch{}
	for id := int64(3); id >= 1; id-- {
		row := leaderboard.PublicEpoch{ID: id, Name: "Epoch", StartedAt: started, ChangelogRef: "changelog/epoch-" + string(rune('0'+id)) + ".md",
			Changelog: []byte("# notes \"quoted\"\n"), Hashes: []string{}}
		if id != 3 {
			row.EndedAt = &ended
			row.Hashes = []string{hashOf("a"), hashOf("b")}
		}
		rows = append(rows, row)
	}
	return rows
}

func newHandler(t *testing.T, reader EpochReader, burst int) (http.Handler, *publicapi.CursorCodec) {
	t.Helper()
	registry, err := Registry()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../balance/api/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := publicapi.LoadPolicy(data)
	if err != nil {
		t.Fatal(err)
	}
	if burst > 0 {
		policy.PublicLimiter.Burst, policy.PublicLimiter.RefillPerMinute = burst, 1
	}
	ids, err := httpapi.NewRequestIDs(policy.RequestID.Pattern, policy.RequestID.MaxBytes, bytes.NewReader(bytes.Repeat([]byte{0xcc}, 4096)))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	runtime, err := publicapi.NewRuntime(policy, func() time.Time { return now }, ids)
	if err != nil {
		t.Fatal(err)
	}
	codec, err := publicapi.NewCursorCodec(bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 32), registry)
	if err != nil {
		t.Fatal(err)
	}
	return runtime.WithRequestID(EpochsHandler{Registry: registry, Cursors: codec, Runtime: runtime, Epochs: reader}), codec
}

func get(handler http.Handler, target string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.RemoteAddr = "192.0.2.7:4000"
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func decodePage(t *testing.T, body []byte) publicEpochPage {
	t.Helper()
	var page publicEpochPage
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&page); err != nil {
		t.Fatalf("page %s: %v", body, err)
	}
	return page
}

func TestEpochsPageIsExactNewestFirstAndCached(t *testing.T) {
	reader := &fakeEpochs{rows: epochRows()}
	handler, _ := newHandler(t, reader, 0)
	response := get(handler, "/api/public/v1/epochs", nil)
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "public,max-age=3600" || response.Header().Get("X-Request-ID") == "" ||
		response.Header().Get("Content-Type") != publicapi.ContentJSON || response.Header().Get("ETag") == "" {
		t.Fatalf("status=%d headers=%v body=%s", response.Code, response.Header(), response.Body.String())
	}
	if reader.calls[0] != [2]int64{0, DefaultPageLimit} {
		t.Fatalf("default page request = %v", reader.calls[0])
	}
	want := `{"items":[` +
		`{"accepted_hashes":[],"changelog_markdown":"# notes \"quoted\"\n","changelog_ref":"changelog/epoch-3.md","ended_at":null,"epoch_id":3,"name":"Epoch","started_at":"2026-07-29T12:00:00.123Z"},` +
		`{"accepted_hashes":["` + hashOf("a") + `","` + hashOf("b") + `"],"changelog_markdown":"# notes \"quoted\"\n","changelog_ref":"changelog/epoch-2.md","ended_at":"2026-07-29T13:00:00.123Z","epoch_id":2,"name":"Epoch","started_at":"2026-07-29T12:00:00.123Z"},` +
		`{"accepted_hashes":["` + hashOf("a") + `","` + hashOf("b") + `"],"changelog_markdown":"# notes \"quoted\"\n","changelog_ref":"changelog/epoch-1.md","ended_at":"2026-07-29T13:00:00.123Z","epoch_id":1,"name":"Epoch","started_at":"2026-07-29T12:00:00.123Z"}` +
		`],"next_cursor":null}` + "\n"
	if response.Body.String() != want {
		t.Fatalf("body mismatch:\n got %s\nwant %s", response.Body.String(), want)
	}
	cached := get(handler, "/api/public/v1/epochs", map[string]string{"If-None-Match": response.Header().Get("ETag")})
	if cached.Code != http.StatusNotModified || cached.Body.Len() != 0 {
		t.Fatalf("conditional request status=%d body=%q", cached.Code, cached.Body.String())
	}
}

func TestEpochsCursorRoundTripsAndBindsTheLimit(t *testing.T) {
	reader := &fakeEpochs{rows: epochRows()}
	handler, _ := newHandler(t, reader, 0)
	first := decodePage(t, get(handler, "/api/public/v1/epochs?limit=2", nil).Body.Bytes())
	if len(first.Items) != 2 || first.Items[0].EpochID != 3 || first.Items[1].EpochID != 2 || first.NextCursor == nil {
		t.Fatalf("first page = %+v", first)
	}
	second := get(handler, "/api/public/v1/epochs?limit=2&cursor="+*first.NextCursor, nil)
	page := decodePage(t, second.Body.Bytes())
	if second.Code != http.StatusOK || len(page.Items) != 1 || page.Items[0].EpochID != 1 || page.NextCursor != nil {
		t.Fatalf("second page status=%d %+v", second.Code, page)
	}
	if reader.calls[len(reader.calls)-1] != [2]int64{2, 2} {
		t.Fatalf("keyset continuation request = %v", reader.calls[len(reader.calls)-1])
	}
	// The cursor is bound to the normalized filter: reusing it under another limit is invalid.
	for _, target := range []string{
		"/api/public/v1/epochs?limit=3&cursor=" + *first.NextCursor,
		"/api/public/v1/epochs?cursor=" + *first.NextCursor,
		"/api/public/v1/epochs?limit=2&cursor=" + *first.NextCursor + "A",
		"/api/public/v1/epochs?limit=2&cursor=",
	} {
		response := get(handler, target, nil)
		if response.Code != http.StatusBadRequest || response.Body.String() != string(invalidCursor) {
			t.Fatalf("%s: status=%d body=%q", target, response.Code, response.Body.String())
		}
	}
}

func TestEpochsRejectsInvalidLimitsWithExactBytes(t *testing.T) {
	reader := &fakeEpochs{rows: epochRows()}
	handler, _ := newHandler(t, reader, 0)
	for _, target := range []string{"?limit=0", "?limit=101", "?limit=07", "?limit=-1", "?limit=x", "?limit=1&limit=2"} {
		response := get(handler, "/api/public/v1/epochs"+target, nil)
		if response.Code != http.StatusBadRequest || response.Body.String() != string(invalidLimit) || response.Header().Get("X-Request-ID") == "" {
			t.Fatalf("%s: status=%d body=%q", target, response.Code, response.Body.String())
		}
	}
	if len(reader.calls) != 0 {
		t.Fatalf("invalid requests reached the repository: %v", reader.calls)
	}
	for _, limit := range []string{"1", "100"} {
		if response := get(handler, "/api/public/v1/epochs?limit="+limit, nil); response.Code != http.StatusOK {
			t.Fatalf("limit %s rejected: %d", limit, response.Code)
		}
	}
}

func TestEpochsFailuresAndLimiterUseDeclaredExactPairs(t *testing.T) {
	registry, err := Registry()
	if err != nil {
		t.Fatal(err)
	}
	failing, _ := newHandler(t, &fakeEpochs{err: errors.New("database down")}, 0)
	response := get(failing, "/api/public/v1/epochs", nil)
	if response.Code != http.StatusInternalServerError || registry.ValidateResponse(ListEpochsOperation, response.Code, response.Body.Bytes()) != nil {
		t.Fatalf("failure status=%d body=%q", response.Code, response.Body.String())
	}
	limited, _ := newHandler(t, &fakeEpochs{rows: epochRows()}, 1)
	if first := get(limited, "/api/public/v1/epochs", nil); first.Code != http.StatusOK {
		t.Fatalf("first limited request status=%d", first.Code)
	}
	second := get(limited, "/api/public/v1/epochs?limit=3", nil)
	if second.Code != http.StatusTooManyRequests || registry.ValidateResponse(ListEpochsOperation, second.Code, second.Body.Bytes()) != nil {
		t.Fatalf("limited status=%d body=%q", second.Code, second.Body.String())
	}
	// Every 400 body the handler emits is a declared exact pair; a wrong pair is refused.
	for _, body := range [][]byte{invalidCursor, invalidLimit} {
		if registry.ValidateResponse(ListEpochsOperation, http.StatusBadRequest, body) != nil {
			t.Fatalf("declared pair rejected: %s", body)
		}
	}
	if registry.ValidateResponse(ListEpochsOperation, http.StatusBadRequest, []byte(`{"category":"invalid","detail":"body"}`+"\n")) == nil {
		t.Fatal("undeclared 400 pair accepted")
	}
}

func TestEpochsRegistryMergesWithThePrivateSurface(t *testing.T) {
	public, err := Registry()
	if err != nil {
		t.Fatal(err)
	}
	operation, ok := public.Operation(ListEpochsOperation)
	if !ok || !operation.Public || operation.Auth != publicapi.AuthNone || operation.Surface != publicapi.SurfacePublicV1 {
		t.Fatalf("operation = %+v", operation)
	}
}
