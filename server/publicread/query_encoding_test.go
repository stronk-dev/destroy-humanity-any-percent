package publicread

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"cloud-clicker/server/publicapi"
)

// Exercise the actual mounted handlers before the controlled page readers. A
// malformed optional field must not disappear and become a default first page.
func TestMountedPublicQueriesRejectMalformedEncoding(t *testing.T) {
	boardQuery := url.Values{"epoch": {"8"}, "mandate": {"0"}, "variables": {variablesParam(t, publicapi.BoardVariables{})}}
	type queryCase struct {
		query string
		body  []byte
	}
	for _, operation := range []struct{ id, path string }{
		{ListEpochsOperation, "/api/public/v1/epochs?"},
		{ListRoutesOperation, "/api/public/v1/registry/routes?"},
		{ListBoardOperation, "/api/public/v1/boards/any_percent?" + boardQuery.Encode() + "&"},
	} {
		cases := []queryCase{
			{"cursor=%", invalidCursor},
			{"cursor=%GG", invalidCursor},
			{"%63ursor=%GG", invalidCursor},
			{"cursor=ignored;other=1", invalidCursor},
			{"limit=%", invalidLimit},
			{"limit=%GG", invalidLimit},
			{"%6cimit=%GG", invalidLimit},
			{"limit=1;other=2", invalidLimit},
			{"limit=%GG&limit=1", invalidLimit},
		}
		if operation.id == ListBoardOperation {
			cases = append(cases, queryCase{"epoch=%GG", invalidEpoch},
				queryCase{"mandate=%GG", invalidMandate}, queryCase{"variables=%GG", invalidVariables})
		}
		for _, test := range cases {
			for _, etag := range []string{"", "*"} {
				t.Run(operation.id+"/"+test.query+"/if-none-match="+etag, func(t *testing.T) {
					fixture := newPublicCacheFixture(t)
					response := publicCacheRequest(fixture.router, operation.path+test.query, etag, "192.0.2.7")
					if response.Code != http.StatusBadRequest || response.Body.String() != string(test.body) ||
						response.Header().Get("Content-Type") != publicapi.ContentJSON || response.Header().Get("X-Request-ID") != "public-cache-check" ||
						response.Header().Get("ETag") != "" || response.Header().Get("Cache-Control") != "" {
						t.Fatalf("malformed query status=%d headers=%v body=%q; want400 %q", response.Code, response.Header(), response.Body.String(), test.body)
					}
					if len(fixture.epochs.calls)+len(fixture.boards.queries)+len(fixture.routes.calls) != 0 {
						t.Fatal("malformed query reached a page reader")
					}
				})
			}
		}
	}
}

func TestMountedPublicQueryEncodingPreservesValidPagingAndIgnoredFields(t *testing.T) {
	fixture := newPublicCacheFixture(t)
	first := publicCacheRequest(fixture.router, "/api/public/v1/epochs?%6cimit=1&unknown=%GG&bad%name=ignored", "", "192.0.2.7")
	page := decodePage(t, first.Body.Bytes())
	if first.Code != http.StatusOK || len(page.Items) != 1 || page.Items[0].EpochID != 3 || page.NextCursor == nil {
		t.Fatalf("valid encoded limit/ignored fields: %d %s", first.Code, first.Body.String())
	}
	second := publicCacheRequest(fixture.router, "/api/public/v1/epochs?limit=1&%63ursor="+url.QueryEscape(*page.NextCursor), "", "192.0.2.7")
	continued := decodePage(t, second.Body.Bytes())
	if second.Code != http.StatusOK || len(continued.Items) != 1 || continued.Items[0].EpochID != 2 ||
		len(fixture.epochs.calls) != 2 || fixture.epochs.calls[1] != [2]int64{3, 1} {
		t.Fatalf("encoded cursor lost continuation: %d %s calls=%v", second.Code, second.Body.String(), fixture.epochs.calls)
	}
	// A malformed duplicate cannot disappear beside an otherwise valid token.
	response := publicCacheRequest(fixture.router, "/api/public/v1/epochs?limit=1&cursor="+*page.NextCursor+"&cursor=%GG", "", "192.0.2.7")
	if response.Code != http.StatusBadRequest || response.Body.String() != string(invalidCursor) || len(fixture.epochs.calls) != 2 {
		t.Fatalf("malformed duplicate replaced with valid cursor: %d %s calls=%v", response.Code, response.Body.String(), fixture.epochs.calls)
	}
	if strings.Contains(response.Body.String(), *page.NextCursor) {
		t.Fatal("error leaked the cursor")
	}
}
