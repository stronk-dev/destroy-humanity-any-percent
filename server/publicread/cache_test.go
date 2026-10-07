package publicread

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/publicapi"
)

// These tests exercise the real registry-mounted handlers and shared middleware,
// with controlled reader data and time. They do not prove Postgres or proxy setup.
type publicCacheFixture struct {
	router   http.Handler
	now      time.Time
	epochs   *fakeEpochs
	boards   *fakeBoards
	routes   *fakeRoutes
	evidence *fakeEvidence
}

func newPublicCacheFixture(t *testing.T) *publicCacheFixture {
	t.Helper()
	fixture := &publicCacheFixture{
		now:      time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC),
		epochs:   &fakeEpochs{rows: epochRows()},
		boards:   &fakeBoards{kinds: map[string]leaderboard.RankingKind{"any_percent": leaderboard.RankingTimeMS}, rows: boardRows()},
		routes:   &fakeRoutes{rows: routeRows()},
		evidence: exampleEvidence(),
	}
	var err error
	fixture.router, err = NewRouter(Dependencies{
		PolicyJSON: phase0PolicyJSON(t), CursorKeys: CursorKeys{CurrentID: "k1", Current: secret(1)},
		Epochs: fixture.epochs, Boards: fixture.boards, Routes: fixture.routes, Evidence: fixture.evidence,
		Clock: func() time.Time { return fixture.now }, Random: rand.Reader,
	})
	if err != nil {
		t.Fatal(err)
	}
	return fixture
}

type publicCacheCase struct {
	path         string
	cacheControl string
	change       func(*publicCacheFixture)
	contentType  string
}

func publicCacheCases(t *testing.T) map[string]publicCacheCase {
	t.Helper()
	query := url.Values{"epoch": {"8"}, "mandate": {"0"}, "variables": {variablesParam(t, publicapi.BoardVariables{})}}
	cases := map[string]publicCacheCase{
		ListBoardOperation: {
			path: "/api/public/v1/boards/any_percent?" + query.Encode(), cacheControl: "public,max-age=60", contentType: publicapi.ContentJSON,
			change: func(fixture *publicCacheFixture) { fixture.boards.rows[0].Key++ },
		},
		ListEpochsOperation: {
			path: "/api/public/v1/epochs", cacheControl: "public,max-age=3600", contentType: publicapi.ContentJSON,
			change: func(fixture *publicCacheFixture) { fixture.epochs.rows[0].Name = "Changed epoch" },
		},
		ListRoutesOperation: {
			path: "/api/public/v1/registry/routes", cacheControl: "public,max-age=300", contentType: publicapi.ContentJSON,
			change: func(fixture *publicCacheFixture) { fixture.routes.rows[0].AdoptionCount++ },
		},
	}
	for _, resource := range []struct{ id, suffix, contentType string }{
		{GetRunGenesisOperation, "genesis", publicapi.ContentJSON},
		{GetRunReplayLogOperation, "replay-log", publicapi.ContentGzip},
		{GetRunVerdictOperation, "verdict", publicapi.ContentJSON},
	} {
		cases[resource.id] = publicCacheCase{
			path:         "/api/public/v1/runs/" + evidenceStream + "/1/" + resource.suffix,
			cacheControl: "public,max-age=31536000,immutable", contentType: resource.contentType,
			// Reader fault only: stored production evidence is immutable. Verify that
			// conditional requests cannot mask changed bytes if a source is wrong.
			change: func(f *publicCacheFixture) {
				f.evidence.value.Genesis = []byte(`{"v":1,"changed":true}`)
				f.evidence.value.GenesisSHA256 = fmt.Sprintf("sha256:%x", sha256.Sum256(f.evidence.value.Genesis))
				f.evidence.value.ReplayLog = append(f.evidence.value.ReplayLog, 0xff)
				f.evidence.value.ReplayLogSHA256 = fmt.Sprintf("sha256:%x", sha256.Sum256(f.evidence.value.ReplayLog))
			},
		}
	}
	registry, err := Registry()
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != len(registry.Operations()) {
		t.Fatal("cache cases must cover the complete mounted public registry")
	}
	for _, operation := range registry.Operations() {
		if _, ok := cases[operation.ID]; !ok {
			t.Fatalf("public operation %s has no cache/limiter case", operation.ID)
		}
	}
	return cases
}

func publicCacheRequest(router http.Handler, path, etag, ip string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.RemoteAddr = "10.0.0.2:4000"
	request.Header.Set("X-Forwarded-For", ip)
	request.Header.Set("X-Request-ID", "public-cache-check")
	if etag != "" {
		request.Header.Set("If-None-Match", etag)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertPublicCacheResponse(t *testing.T, response *httptest.ResponseRecorder, status int, cacheControl, etag string) {
	t.Helper()
	if response.Code != status || response.Header().Get("X-Request-ID") != "public-cache-check" ||
		response.Header().Get("Cache-Control") != cacheControl || response.Header().Get("ETag") != etag {
		t.Fatalf("status=%d headers=%v body=%q; want status=%d cache=%q etag=%q", response.Code, response.Header(), response.Body.String(), status, cacheControl, etag)
	}
	switch status {
	case http.StatusOK:
		want := fmt.Sprintf(`"%x"`, sha256.Sum256(response.Body.Bytes()))
		if etag != want || (response.Header().Get("Content-Type") != publicapi.ContentJSON && response.Header().Get("Content-Type") != publicapi.ContentGzip) ||
			response.Header().Get("Content-Length") != fmt.Sprint(response.Body.Len()) {
			t.Fatalf("200 must identify the exact served bytes: %v %q", response.Header(), response.Body.String())
		}
	case http.StatusNotModified:
		if response.Body.Len() != 0 || response.Header().Get("Content-Type") != "" || response.Header().Get("Content-Length") != "" {
			t.Fatalf("304 must be bodiless: %v %q", response.Header(), response.Body.String())
		}
	case http.StatusTooManyRequests:
		if response.Body.String() != "{\"category\":\"rate_limited\",\"detail\":\"ip\"}\n" || response.Header().Get("Content-Type") != publicapi.ContentJSON || response.Header().Get(EvidenceHashHeader) != "" {
			t.Fatalf("429 must be the typed, non-cacheable limiter response: %v %q", response.Header(), response.Body.String())
		}
	default:
		t.Fatalf("unsupported cache test status %d", status)
	}
}

func TestMountedPublicReadsCacheWithoutSpendingTokensAndInvalidateChangedBytes(t *testing.T) {
	policy, err := publicapi.LoadPolicy(phase0PolicyJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	for operation, test := range publicCacheCases(t) {
		t.Run(operation, func(t *testing.T) {
			fixture := newPublicCacheFixture(t)
			const ip = "192.0.2.7"
			first := publicCacheRequest(fixture.router, test.path, "", ip)
			etag := first.Header().Get("ETag")
			assertPublicCacheResponse(t, first, http.StatusOK, test.cacheControl, etag)
			if first.Header().Get("Content-Type") != test.contentType {
				t.Fatalf("operation %s served the wrong content type: %v", operation, first.Header())
			}
			// More cache hits than the real configured burst: none may charge a token.
			for index := 0; index <= policy.PublicLimiter.Burst; index++ {
				cached := publicCacheRequest(fixture.router, test.path, etag, ip)
				assertPublicCacheResponse(t, cached, http.StatusNotModified, test.cacheControl, etag)
				if operation == GetRunGenesisOperation || operation == GetRunReplayLogOperation {
					if cached.Header().Get(EvidenceHashHeader) != strings.Trim(etag, `"`) {
						t.Fatal("raw cached evidence lost its declared hash header")
					}
				}
			}
			// Exactly burst-1 remaining uncached requests succeed at this frozen time.
			for index := 1; index < policy.PublicLimiter.Burst; index++ {
				response := publicCacheRequest(fixture.router, test.path, "", ip)
				assertPublicCacheResponse(t, response, http.StatusOK, test.cacheControl, etag)
				if !bytes.Equal(response.Body.Bytes(), first.Body.Bytes()) {
					t.Fatal("unchanged reader data changed served bytes")
				}
			}
			assertPublicCacheResponse(t, publicCacheRequest(fixture.router, test.path, "", ip), http.StatusTooManyRequests, "", "")
			assertPublicCacheResponse(t, publicCacheRequest(fixture.router, test.path, etag, ip), http.StatusNotModified, test.cacheControl, etag)
			assertPublicCacheResponse(t, publicCacheRequest(fixture.router, test.path, `"wrong-tag"`, ip), http.StatusTooManyRequests, "", "")

			test.change(fixture)
			// An old tag cannot produce 304 for changed reader bytes, even at exhaustion.
			assertPublicCacheResponse(t, publicCacheRequest(fixture.router, test.path, etag, ip), http.StatusTooManyRequests, "", "")
			changed := publicCacheRequest(fixture.router, test.path, etag, "192.0.2.8")
			changedETag := changed.Header().Get("ETag")
			assertPublicCacheResponse(t, changed, http.StatusOK, test.cacheControl, changedETag)
			if changedETag == etag || bytes.Equal(changed.Body.Bytes(), first.Body.Bytes()) {
				t.Fatal("changed reader data must change both body and ETag")
			}
			assertPublicCacheResponse(t, publicCacheRequest(fixture.router, test.path, changedETag, ip), http.StatusNotModified, test.cacheControl, changedETag)

			// The actual configured refill supplies exactly one token, not a new burst.
			fixture.now = fixture.now.Add(time.Minute / time.Duration(policy.PublicLimiter.RefillPerMinute))
			refilled := publicCacheRequest(fixture.router, test.path, etag, ip)
			assertPublicCacheResponse(t, refilled, http.StatusOK, test.cacheControl, changedETag)
			if !bytes.Equal(refilled.Body.Bytes(), changed.Body.Bytes()) {
				t.Fatal("refilled request did not serve current reader bytes")
			}
			assertPublicCacheResponse(t, publicCacheRequest(fixture.router, test.path, etag, ip), http.StatusTooManyRequests, "", "")
		})
	}
}

func TestMountedPublicReadsShareOneIPBudgetAcrossOperations(t *testing.T) {
	policy, err := publicapi.LoadPolicy(phase0PolicyJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	registry, err := Registry()
	if err != nil {
		t.Fatal(err)
	}
	cases := publicCacheCases(t)
	operations := registry.Operations()
	fixture := newPublicCacheFixture(t)
	for index := 0; index < policy.PublicLimiter.Burst; index++ {
		test := cases[operations[index%len(operations)].ID]
		response := publicCacheRequest(fixture.router, test.path, "", "192.0.2.7")
		assertPublicCacheResponse(t, response, http.StatusOK, test.cacheControl, response.Header().Get("ETag"))
	}
	for _, operation := range operations {
		test := cases[operation.ID]
		assertPublicCacheResponse(t, publicCacheRequest(fixture.router, test.path, "", "192.0.2.7"), http.StatusTooManyRequests, "", "")
		other := publicCacheRequest(fixture.router, test.path, "", "192.0.2.8")
		assertPublicCacheResponse(t, other, http.StatusOK, test.cacheControl, other.Header().Get("ETag"))
	}
}
