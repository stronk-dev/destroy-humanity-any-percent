package publicread

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/publicapi"
)

// Actual registry/router/policy, controlled readers and frozen time. This proves
// response-budget enforcement, not SQL behavior or a pre-reader work bound.
func TestMountedPublicErrorsSpendTheSharedIPBudget(t *testing.T) {
	policy, err := publicapi.LoadPolicy(phase0PolicyJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	type errorCase struct {
		name, method, path string
		status             int
		body               []byte
		prepare            func(*publicCacheFixture)
	}
	boardPath := func(category, epoch, variables string) string {
		return "/api/public/v1/boards/" + category + "?" + url.Values{"epoch": {epoch}, "mandate": {"0"}, "variables": {variables}}.Encode()
	}
	variables := variablesParam(t, publicapi.BoardVariables{})
	cases := []errorCase{
		{"encoded epoch cursor", "GET", "/api/public/v1/epochs?cursor=%GG", 400, invalidCursor, nil},
		{"encoded route limit", "GET", "/api/public/v1/registry/routes?limit=1;2", 400, invalidLimit, nil},
		{"encoded board limit", "GET", boardPath("any_percent", "8", variables) + "&limit=%GG", 400, invalidLimit, nil},
		{"epoch limit", "GET", "/api/public/v1/epochs?limit=0", 400, invalidLimit, nil},
		{"epoch cursor", "GET", "/api/public/v1/epochs?cursor=bad", 400, invalidCursor, nil},
		{"epoch reader", "GET", "/api/public/v1/epochs", 500, internalPublic, func(f *publicCacheFixture) { f.epochs.err = errors.New("private database detail") }},
		{"route limit", "GET", "/api/public/v1/registry/routes?limit=0", 400, invalidLimit, nil},
		{"route cursor", "GET", "/api/public/v1/registry/routes?cursor=bad", 400, invalidCursor, nil},
		{"route reader", "GET", "/api/public/v1/registry/routes", 500, internalPublic, func(f *publicCacheFixture) { f.routes.err = errors.New("private database detail") }},
		{"board variables", "GET", boardPath("any_percent", "8", "bad"), 400, invalidVariables, nil},
		{"board reader", "GET", boardPath("any_percent", "8", variables), 500, internalPublic, func(f *publicCacheFixture) { f.boards.kindErr = errors.New("private database detail") }},
		{"unknown category", "GET", boardPath("missing", "8", variables), 404, unknownCategory, nil},
		{"unknown epoch", "GET", boardPath("any_percent", "7", variables), 404, unknownEpoch, nil},
		{"unknown path", "GET", "/api/public/v1/missing", 404, notFound, nil},
		{"wrong method", "POST", "/api/public/v1/epochs", 404, notFound, nil},
	}
	for _, suffix := range []string{"genesis", "replay-log", "verdict"} {
		path := "/api/public/v1/runs/" + evidenceStream + "/1/" + suffix
		cases = append(cases,
			errorCase{"private " + suffix, "GET", path, 404, unknownRun, func(f *publicCacheFixture) { f.evidence.err = leaderboard.ErrUnknownPublicRun }},
			errorCase{"failed " + suffix, "GET", path, 500, internalPublic, func(f *publicCacheFixture) { f.evidence.err = errors.New("private database detail") }},
			errorCase{"invalid seq " + suffix, "GET", strings.Replace(path, "/1/", "/01/", 1), 404, unknownRun, nil})
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newPublicCacheFixture(t)
			// Another IP obtains a valid tag. Even after this IP exhausts its
			// error budget, an authorized 304 must still be free.
			cachePath, cacheControl := "/api/public/v1/registry/routes", "public,max-age=300"
			if strings.HasPrefix(test.name, "route ") {
				cachePath, cacheControl = "/api/public/v1/epochs", "public,max-age=3600"
			}
			seed := publicCacheRequest(fixture.router, cachePath, "", "192.0.2.8")
			etag := seed.Header().Get("ETag")
			assertPublicCacheResponse(t, seed, http.StatusOK, cacheControl, etag)
			if test.prepare != nil {
				test.prepare(fixture)
			}
			request := func(ip string) *httptest.ResponseRecorder {
				r := httptest.NewRequest(test.method, test.path, nil)
				r.RemoteAddr = "10.0.0.2:4000"
				r.Header.Set("X-Forwarded-For", ip)
				r.Header.Set("X-Request-ID", "public-cache-check")
				// A wildcard cannot turn a private/missing/invalid representation
				// into a free 304 or hide a reader failure.
				r.Header.Set("If-None-Match", "*")
				response := httptest.NewRecorder()
				fixture.router.ServeHTTP(response, r)
				return response
			}
			assertError := func(response *httptest.ResponseRecorder) {
				t.Helper()
				if response.Code != test.status || response.Body.String() != string(test.body) ||
					response.Header().Get("X-Request-ID") != "public-cache-check" || response.Header().Get("Content-Type") != publicapi.ContentJSON ||
					response.Header().Get("Cache-Control") != "" || response.Header().Get("ETag") != "" || response.Header().Get(EvidenceHashHeader) != "" {
					t.Fatalf("error status=%d headers=%v body=%q; want %d %q", response.Code, response.Header(), response.Body.String(), test.status, test.body)
				}
			}
			for index := 0; index < policy.PublicLimiter.Burst; index++ {
				assertError(request("192.0.2.7"))
			}
			assertPublicCacheResponse(t, request("192.0.2.7"), http.StatusTooManyRequests, "", "")
			assertPublicCacheResponse(t, publicCacheRequest(fixture.router, cachePath, "", "192.0.2.7"), http.StatusTooManyRequests, "", "")
			assertPublicCacheResponse(t, publicCacheRequest(fixture.router, cachePath, etag, "192.0.2.7"), http.StatusNotModified, cacheControl, etag)
			assertError(request("192.0.2.9"))
			fixture.now = fixture.now.Add(time.Minute / time.Duration(policy.PublicLimiter.RefillPerMinute))
			assertError(request("192.0.2.7"))
			assertPublicCacheResponse(t, request("192.0.2.7"), http.StatusTooManyRequests, "", "")
		})
	}
}
