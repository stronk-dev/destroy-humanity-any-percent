package publicread

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/publicapi"
)

func phase0PolicyJSON(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "balance", "api", "phase0.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func secret(fill byte) []byte { return bytes.Repeat([]byte{fill}, 32) }

func routerFor(t *testing.T, policy []byte, keys CursorKeys, epochs EpochReader) (http.Handler, error) {
	t.Helper()
	return NewRouter(Dependencies{PolicyJSON: policy, CursorKeys: keys, Epochs: epochs,
		Clock: func() time.Time { return time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC) }, Random: rand.Reader})
}

func TestCursorSecretResolverBindsPolicyNamesToTheDeploymentPair(t *testing.T) {
	policy, err := publicapi.LoadPolicy(phase0PolicyJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	first := CursorSecretResolver(policy, CursorKeys{CurrentID: "k1", Current: secret(1)})
	if value, ok := first("k1"); !ok || !bytes.Equal(value, secret(1)) {
		t.Fatal("current name must resolve to the current secret")
	}
	if value, ok := first("k0"); !ok || !bytes.Equal(value, secret(1)) {
		t.Fatal("first deployment: the previous name shares the current secret (C20)")
	}
	rotated := CursorSecretResolver(policy, CursorKeys{CurrentID: "k1", Current: secret(1), PreviousID: "k0", Previous: secret(2)})
	if value, ok := rotated("k0"); !ok || !bytes.Equal(value, secret(2)) {
		t.Fatal("a deployment previous pair must own the previous name")
	}
	for _, keys := range []CursorKeys{
		{CurrentID: "cursor-current", Current: secret(1)},                            // names do not match the policy
		{CurrentID: "k1", Current: secret(1), PreviousID: "k9", Previous: secret(2)}, // previous pair under a foreign name
		{CurrentID: "k1"}, // no material
	} {
		resolve := CursorSecretResolver(policy, keys)
		_, currentOK := resolve("k1")
		_, previousOK := resolve("k0")
		if currentOK && previousOK {
			t.Fatalf("keys %+v resolved both policy names", keys.CurrentID+"/"+keys.PreviousID)
		}
	}
}

func TestNewRouterFailsClosedWithoutPolicyOrCursorSecrets(t *testing.T) {
	epochs := &fakeEpochs{rows: epochRows()}
	if _, err := routerFor(t, phase0PolicyJSON(t), CursorKeys{}, epochs); err == nil {
		t.Fatal("missing cursor secrets must fail composition")
	}
	if _, err := routerFor(t, phase0PolicyJSON(t), CursorKeys{CurrentID: "k1", Current: []byte("short")}, epochs); err == nil {
		t.Fatal("a sub-32-byte cursor secret must fail composition")
	}
	if _, err := routerFor(t, []byte(`{}`), CursorKeys{CurrentID: "k1", Current: secret(1)}, epochs); err == nil {
		t.Fatal("an invalid policy must fail composition")
	}
	if _, err := routerFor(t, phase0PolicyJSON(t), CursorKeys{CurrentID: "k1", Current: secret(1)}, nil); err == nil {
		t.Fatal("a missing reader must fail composition")
	}
}

func TestNewRouterMountsTheRegistryWithRequestIDsAndCaching(t *testing.T) {
	router, err := routerFor(t, phase0PolicyJSON(t), CursorKeys{CurrentID: "k1", Current: secret(1)}, &fakeEpochs{rows: epochRows()})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/public/v1/epochs?limit=2", nil)
	request.Header.Set("X-Request-ID", "support-123")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("X-Request-ID") != "support-123" ||
		response.Header().Get("ETag") == "" || response.Header().Get("Cache-Control") != "public,max-age=3600" {
		t.Fatalf("mounted epochs: %d %v %s", response.Code, response.Header(), response.Body.String())
	}
	var page publicEpochPage
	if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil || len(page.Items) != 2 || page.NextCursor == nil {
		t.Fatalf("page %s", response.Body.String())
	}
	for _, path := range []string{"/api/public/v1/unknown", "/api/public/v1/epochs/1"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound || response.Body.String() != string(notFound) || response.Header().Get("X-Request-ID") == "" {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body.String())
		}
	}
	post := httptest.NewRecorder()
	router.ServeHTTP(post, httptest.NewRequest(http.MethodPost, "/api/public/v1/epochs", strings.NewReader("{}")))
	if post.Code != http.StatusNotFound {
		t.Fatalf("public surface is read-only: %d", post.Code)
	}
}

func TestCursorSurvivesTheGovernedRotation(t *testing.T) {
	epochs := &fakeEpochs{rows: epochRows()}
	before, err := routerFor(t, phase0PolicyJSON(t), CursorKeys{CurrentID: "k1", Current: secret(1)}, epochs)
	if err != nil {
		t.Fatal(err)
	}
	first := httptest.NewRecorder()
	before.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/public/v1/epochs?limit=1", nil))
	var page publicEpochPage
	if err := json.Unmarshal(first.Body.Bytes(), &page); err != nil || page.NextCursor == nil {
		t.Fatalf("first page %s", first.Body.String())
	}
	// Rotation (C20: names are fixed, only deployment secrets move): a new
	// secret becomes k1 and the old one becomes k0.
	after, err := routerFor(t, phase0PolicyJSON(t), CursorKeys{CurrentID: "k1", Current: secret(3), PreviousID: "k0", Previous: secret(1)}, epochs)
	if err != nil {
		t.Fatal(err)
	}
	next := httptest.NewRecorder()
	after.ServeHTTP(next, httptest.NewRequest(http.MethodGet, "/api/public/v1/epochs?limit=1&cursor="+*page.NextCursor, nil))
	if next.Code != http.StatusOK {
		t.Fatalf("a pre-rotation cursor must still decode: %d %s", next.Code, next.Body.String())
	}
	// After the previous key is removed, the old cursor must reject.
	retired, err := routerFor(t, phase0PolicyJSON(t), CursorKeys{CurrentID: "k1", Current: secret(4), PreviousID: "k0", Previous: secret(3)}, epochs)
	if err != nil {
		t.Fatal(err)
	}
	stale := httptest.NewRecorder()
	retired.ServeHTTP(stale, httptest.NewRequest(http.MethodGet, "/api/public/v1/epochs?limit=1&cursor="+*page.NextCursor, nil))
	if stale.Code != http.StatusBadRequest || stale.Body.String() != string(invalidCursor) {
		t.Fatalf("a retired-key cursor must reject: %d %s", stale.Code, stale.Body.String())
	}
}
