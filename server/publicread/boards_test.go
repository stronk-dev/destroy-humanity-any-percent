package publicread

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/publicapi"
)

type fakeBoards struct {
	kinds   map[string]leaderboard.RankingKind // category -> kind for epoch 8
	rows    []leaderboard.PublicBoardRow
	kindErr error
	queries []leaderboard.PublicBoardQuery
}

func (fake *fakeBoards) PublicBoardRankingKind(_ context.Context, categoryID string, epochID int64) (leaderboard.RankingKind, error) {
	if fake.kindErr != nil {
		return "", fake.kindErr
	}
	if epochID != 8 {
		return "", leaderboard.ErrUnknownPublicEpoch
	}
	kind, ok := fake.kinds[categoryID]
	if !ok {
		return "", leaderboard.ErrUnknownPublicCategory
	}
	return kind, nil
}

// PublicBoardPage emulates keyset paging over rows already in board order.
func (fake *fakeBoards) PublicBoardPage(_ context.Context, kind leaderboard.RankingKind, query leaderboard.PublicBoardQuery) ([]leaderboard.PublicBoardRow, bool, error) {
	fake.queries = append(fake.queries, query)
	start := 0
	if query.AfterRunID != "" {
		for index, row := range fake.rows {
			if row.RunID == query.AfterRunID {
				start = index + 1
			}
		}
	}
	rows := fake.rows[start:]
	if len(rows) > query.Limit {
		return rows[:query.Limit], true, nil
	}
	return rows, false, nil
}

func boardRows() []leaderboard.PublicBoardRow {
	verified := time.Date(2026, 9, 1, 10, 0, 0, 5_000_000, time.UTC)
	return []leaderboard.PublicBoardRow{
		{RunID: "01986666-0000-4000-8000-000000000001:1", FounderID: "01986666-0000-4000-8000-0000000000f1", Rank: 1, Key: 1000, Magnitude: leaderboard.MagnitudeKey{Exponent: 12, Mantissa: 123_456_789_012}, VerifiedAt: verified, WorldFirst: true},
		{RunID: "01986666-0000-4000-8000-000000000002:1", FounderID: "01986666-0000-4000-8000-0000000000f2", Rank: 2, Key: 2000, Magnitude: leaderboard.MagnitudeKey{Exponent: 11, Mantissa: 100_000_000_000}, VerifiedAt: verified},
		{RunID: "01986666-0000-4000-8000-000000000003:1", FounderID: "01986666-0000-4000-8000-0000000000f3", Rank: 2, Key: 2000, Magnitude: leaderboard.MagnitudeKey{Exponent: 11, Mantissa: 100_000_000_000}, VerifiedAt: verified},
	}
}

func boardRouter(t *testing.T, boards *fakeBoards) http.Handler {
	t.Helper()
	router, err := NewRouter(Dependencies{PolicyJSON: phase0PolicyJSON(t), CursorKeys: CursorKeys{CurrentID: "k1", Current: secret(1)},
		Epochs: &fakeEpochs{rows: epochRows()}, Boards: boards,
		Clock: func() time.Time { return time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC) }, Random: rand.Reader})
	if err != nil {
		t.Fatal(err)
	}
	return router
}

func variablesParam(t *testing.T, value publicapi.BoardVariables) string {
	t.Helper()
	encoded, err := publicapi.EncodeBoardVariables(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func getBoard(router http.Handler, path string, values url.Values) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path+"?"+values.Encode(), nil))
	return response
}

func TestBoardReaderServesTheC12PageAndPagesByKeysetCursor(t *testing.T) {
	boards := &fakeBoards{kinds: map[string]leaderboard.RankingKind{"any_percent": leaderboard.RankingTimeMS}, rows: boardRows()}
	router := boardRouter(t, boards)
	faction := "faction.labs"
	values := url.Values{"epoch": {"8"}, "mandate": {"0"}, "limit": {"2"},
		"variables": {variablesParam(t, publicapi.BoardVariables{Advisor: 1, Commons: 0, Faction: &faction, Glitched: 0})}}
	first := getBoard(router, "/api/public/v1/boards/any_percent", values)
	if first.Code != http.StatusOK || first.Header().Get("Cache-Control") != "public,max-age=60" {
		t.Fatalf("first page %d %v %s", first.Code, first.Header(), first.Body.String())
	}
	want := `{"category_id":"any_percent","epoch_id":8,"items":[` +
		`{"founder_id":"01986666-0000-4000-8000-0000000000f1","key":{"kind":"time_ms","value":1000},"rank":1,"run_id":"01986666-0000-4000-8000-000000000001:1","verified_at":"2026-09-01T10:00:00.005Z","world_first":true},` +
		`{"founder_id":"01986666-0000-4000-8000-0000000000f2","key":{"kind":"time_ms","value":2000},"rank":2,"run_id":"01986666-0000-4000-8000-000000000002:1","verified_at":"2026-09-01T10:00:00.005Z","world_first":false}],` +
		`"mandate_level":0,"next_cursor":`
	if len(first.Body.String()) < len(want) || first.Body.String()[:len(want)] != want {
		t.Fatalf("page bytes:\n%s\nwant prefix:\n%s", first.Body.String(), want)
	}
	if boards.queries[0].Variables.Advisor != true || boards.queries[0].Variables.Faction == nil || *boards.queries[0].Variables.Faction != faction || boards.queries[0].Limit != 2 {
		t.Fatalf("normalized filter did not reach the reader: %+v", boards.queries[0])
	}
	var page publicBoardPage
	if err := json.Unmarshal(first.Body.Bytes(), &page); err != nil || page.NextCursor == nil || page.RankingKind != "time_ms" {
		t.Fatalf("page %s", first.Body.String())
	}
	values.Set("cursor", *page.NextCursor)
	second := getBoard(router, "/api/public/v1/boards/any_percent", values)
	var secondPage publicBoardPage
	if second.Code != http.StatusOK || json.Unmarshal(second.Body.Bytes(), &secondPage) != nil || len(secondPage.Items) != 1 ||
		secondPage.Items[0].Rank != 2 || secondPage.NextCursor != nil || boards.queries[1].AfterKey == nil || *boards.queries[1].AfterKey != 2000 {
		t.Fatalf("second page %d %s", second.Code, second.Body.String())
	}
	// The cursor MAC binds the whole normalized filter: any other filter rejects it.
	for name, mutate := range map[string]func(url.Values){
		"limit":     func(v url.Values) { v.Set("limit", "3") },
		"epoch":     func(v url.Values) { v.Set("epoch", "9") },
		"mandate":   func(v url.Values) { v.Set("mandate", "1") },
		"variables": func(v url.Values) { v.Set("variables", variablesParam(t, publicapi.BoardVariables{})) },
	} {
		mutated := url.Values{}
		for key, value := range values {
			mutated[key] = append([]string(nil), value...)
		}
		mutate(mutated)
		kinds := map[string]leaderboard.RankingKind{"any_percent": leaderboard.RankingTimeMS}
		response := getBoard(boardRouter(t, &fakeBoards{kinds: kinds, rows: boardRows()}), "/api/public/v1/boards/any_percent", mutated)
		if name == "epoch" {
			if response.Code != http.StatusNotFound {
				t.Fatalf("epoch 9 is unknown to the fake: %d", response.Code)
			}
			continue
		}
		if response.Code != http.StatusBadRequest || response.Body.String() != string(invalidCursor) {
			t.Fatalf("%s-mutated filter accepted the cursor: %d %s", name, response.Code, response.Body.String())
		}
	}
	// The same cursor on another category rejects too.
	other := &fakeBoards{kinds: map[string]leaderboard.RankingKind{"any_percent": leaderboard.RankingTimeMS, "low_percent": leaderboard.RankingTimeMS}, rows: boardRows()}
	if response := getBoard(boardRouter(t, other), "/api/public/v1/boards/low_percent", values); response.Code != http.StatusBadRequest {
		t.Fatalf("cross-category cursor accepted: %d", response.Code)
	}
}

func TestBoardReaderRendersTheMagnitudeArmAndRejectsForeignCursorArms(t *testing.T) {
	boards := &fakeBoards{kinds: map[string]leaderboard.RankingKind{"valuation": leaderboard.RankingMagnitude}, rows: boardRows()}
	router := boardRouter(t, boards)
	values := url.Values{"epoch": {"8"}, "mandate": {"0"}, "limit": {"1"}, "variables": {variablesParam(t, publicapi.BoardVariables{})}}
	response := getBoard(router, "/api/public/v1/boards/valuation", values)
	var page publicBoardPage
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &page) != nil || page.RankingKind != "magnitude" ||
		page.Items[0].Key["exponent"] != float64(12) || page.Items[0].Key["quantized_mantissa"] != float64(123_456_789_012) || page.Items[0].Key["value"] != nil {
		t.Fatalf("magnitude page %d %s", response.Code, response.Body.String())
	}
	values.Set("cursor", *page.NextCursor)
	if next := getBoard(router, "/api/public/v1/boards/valuation", values); next.Code != http.StatusOK || boards.queries[1].AfterMag == nil || boards.queries[1].AfterMag.Exponent != 12 {
		t.Fatalf("magnitude cursor %d %s", next.Code, next.Body.String())
	}
	var key boardCursorKey
	query := leaderboard.PublicBoardQuery{}
	value := int64(5)
	key = boardCursorKey{Key: &value, RunID: "r"}
	if applyBoardCursor(leaderboard.RankingMagnitude, key, &query) {
		t.Fatal("a time-arm cursor must not page a magnitude board")
	}
	exponent, mantissa := int64(1), int64(100_000_000_000)
	if applyBoardCursor(leaderboard.RankingTimeMS, boardCursorKey{Exponent: &exponent, QuantizedMantissa: &mantissa, RunID: "r"}, &query) {
		t.Fatal("a magnitude-arm cursor must not page a time board")
	}
	if applyBoardCursor(leaderboard.RankingMagnitude, boardCursorKey{Key: &value, Exponent: &exponent, QuantizedMantissa: &mantissa, RunID: "r"}, &query) ||
		applyBoardCursor(leaderboard.RankingTimeMS, boardCursorKey{Key: &value, Exponent: &exponent, QuantizedMantissa: &mantissa, RunID: "r"}, &query) {
		t.Fatal("a cursor carrying both arms must reject for every ranking kind")
	}
	if applyBoardCursor(leaderboard.RankingTimeMS, boardCursorKey{Key: &value}, &query) {
		t.Fatal("a cursor without its run tie-breaker must reject")
	}
}

func TestBoardReaderMapsEveryDeclaredRejection(t *testing.T) {
	base := func() url.Values {
		return url.Values{"epoch": {"8"}, "mandate": {"0"}, "variables": {variablesParam(t, publicapi.BoardVariables{})}}
	}
	kinds := map[string]leaderboard.RankingKind{"any_percent": leaderboard.RankingTimeMS}
	cases := []struct {
		name    string
		path    string
		mutate  func(url.Values)
		status  int
		body    []byte
		kindErr error
	}{
		{"missing epoch", "any_percent", func(v url.Values) { v.Del("epoch") }, 400, invalidEpoch, nil},
		{"zero epoch", "any_percent", func(v url.Values) { v.Set("epoch", "0") }, 400, invalidEpoch, nil},
		{"missing mandate", "any_percent", func(v url.Values) { v.Del("mandate") }, 400, invalidMandate, nil},
		{"mandate above bound", "any_percent", func(v url.Values) { v.Set("mandate", "21") }, 400, invalidMandate, nil},
		{"missing variables", "any_percent", func(v url.Values) { v.Del("variables") }, 400, invalidVariables, nil},
		{"noncanonical variables", "any_percent", func(v url.Values) { v.Set("variables", "eyJhZHZpc29yIjowfQ") }, 400, invalidVariables, nil},
		{"limit above bound", "any_percent", func(v url.Values) { v.Set("limit", "101") }, 400, invalidLimit, nil},
		{"tampered cursor", "any_percent", func(v url.Values) { v.Set("cursor", "AAAA") }, 400, invalidCursor, nil},
		{"unknown category", "low_percent", func(url.Values) {}, 404, unknownCategory, nil},
		{"unknown epoch", "any_percent", func(v url.Values) { v.Set("epoch", "7") }, 404, unknownEpoch, nil},
		{"reader failure", "any_percent", func(url.Values) {}, 500, internalPublic, errors.New("database down")},
	}
	for _, test := range cases {
		values := base()
		test.mutate(values)
		response := getBoard(boardRouter(t, &fakeBoards{kinds: kinds, rows: boardRows(), kindErr: test.kindErr}), "/api/public/v1/boards/"+test.path, values)
		if response.Code != test.status || response.Body.String() != string(test.body) {
			t.Fatalf("%s: %d %s", test.name, response.Code, response.Body.String())
		}
	}
}
