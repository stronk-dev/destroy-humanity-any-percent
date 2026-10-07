package account

import (
	"io"
	"net/http"
	"testing"

	"cloud-clicker/server/internal/testhttp"
)

// Real Account authentication, registry routing and Postgres reads; the recovery
// engine is deliberately absent or a constructor-only fixture. This is error
// conformance, not an actual attended-recovery or browser gameplay proof.
func TestSoulRecoveryErrorRegistryIntegration(t *testing.T) {
	fixture := newRefreshPersistenceFixture(t, 60)
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"start_soul_recovery", "progress_soul_recovery", "resolve_soul_recovery", "cancel_soul_recovery"} {
		t.Run(operation+"/unavailable", func(t *testing.T) {
			row, ok := registry.Operation(operation)
			if !ok {
				t.Fatal("registered recovery operation missing")
			}
			response := requestJSON(t, fixture.server.Client, row.Method, fixture.server.URL+row.Path, fixture.initial.AccessToken, "{}")
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil || response.StatusCode != http.StatusServiceUnavailable || response.Header.Get("Content-Type") != "application/json" {
				t.Fatal("mounted unavailable handler lost its 503 JSON response")
			}
			assertSoulErrorContract(t, registry, operation, response.StatusCode, body, "not_configured", "soul_recovery")
		})
	}
	t.Run("start_soul_recovery/missing-company", func(t *testing.T) {
		api, err := NewAPI(fixture.repository, refreshCensusUnreachableIntents{t}, Phase0APIConfig(testBootstrapReceiptKeys()))
		if err != nil {
			t.Fatal(err)
		}
		recoveries := &integrationSoulRecoveries{}
		if err := api.AttachSoulRecoveries(recoveries); err != nil {
			t.Fatal(err)
		}
		// Archive only this disposable fixture's Company so real authentication
		// still succeeds but the actual active-Company lookup has no row.
		result, err := fixture.db.Exec(`UPDATE save_streams SET archived_at=$2 WHERE owner_kind='founder' AND owner_id=$1 AND scope='company' AND archived_at IS NULL`, fixture.account.FounderID, fixture.now)
		if err != nil {
			t.Fatal(err)
		}
		if rows, err := result.RowsAffected(); err != nil || rows != 1 {
			t.Fatal("fixture did not archive exactly its one Company")
		}
		server := testhttp.New(api.Router())
		defer server.Close()
		response := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/soul-recovery/start", fixture.initial.AccessToken, `{"activity_id":"defrag"}`)
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil || response.StatusCode != http.StatusNotFound || response.Header.Get("Content-Type") != "application/json" {
			t.Fatal("mounted missing-Company handler lost its 404 JSON response")
		}
		assertSoulErrorContract(t, registry, "start_soul_recovery", response.StatusCode, body, "unknown_id", "company_stream")
		if recoveries.start.FounderID != "" {
			t.Fatal("missing-Company refusal reached the recovery engine")
		}
	})
}
