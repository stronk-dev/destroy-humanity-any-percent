package gameserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

const composedPetCatalogV2 = `{"schema_version":2,"stat_policy":{"grid_ms":60000,"stats":[{"stat_id":"hunger","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000},{"stat_id":"energy","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000},{"stat_id":"cleanliness","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000},{"stat_id":"affection","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000}],"diminishing_threshold_ppm":700000,"diminishing_factor_ppm":500000},"actions":[{"action_id":"care.action_a","stat_id":"hunger","delta_ppm":100000,"cooldown_attended_ms":60000,"min_eligible_ppm":0,"soul_gate":"essential"},{"action_id":"care.action_b","stat_id":"hunger","delta_ppm":100000,"cooldown_attended_ms":60000,"min_eligible_ppm":0,"soul_gate":"ordinary"},{"action_id":"care.action_c","stat_id":"hunger","delta_ppm":100000,"cooldown_attended_ms":60000,"min_eligible_ppm":0,"soul_gate":"recovery"}],"trust_policy":{"initial_ppm":500000,"neutral_ppm":500000,"floor_ppm":100000,"cap_ppm":1000000,"gain_ppm_per_effective_action":1000,"decay_ppm_per_grid":100},"mood_policy":[{"mood_member":"withdrawn","floor_ppm":0},{"mood_member":"restless","floor_ppm":250000},{"mood_member":"neutral","floor_ppm":500000},{"mood_member":"engaged","floor_ppm":750000}],"behavior_policy":[{"from_state":"idle","event":"care_applied","to_state":"care_response","duration_grid_ticks":1}]}`

func TestComposedMinigameAPILifecycleUsesPinnedTenantResolverIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	const cleanDatabase = `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`
	if _, err := db.ExecContext(ctx, cleanDatabase); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), cleanDatabase); err != nil {
			t.Errorf("clean minigame API database: %v", err)
		}
	})

	now := time.Now().UTC().Add(-time.Second).Truncate(100 * time.Millisecond)
	clock := &mutableClock{now: now}
	composition, err := Compose(ctx, CompositionConfig{
		DB: db, RepositoryRoot: composedMinigameRepositoryRoot(t, filepathRoot(t)),
		ServerID: "01986666-f100-4000-8000-000000000001", ActivityBracket: "activity.standard",
		Clock: clock.Time, SigningKeys: account.SigningKeys{CurrentID: "minigame-lifecycle", Current: bytes.Repeat([]byte{0x57}, 32)},
	})
	if err != nil {
		t.Fatal(err)
	}
	serverContext, cancelServer := context.WithCancel(ctx)
	defer cancelServer()
	if err := composition.Server.Start(serverContext); err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(composition.Server.Handler())
	defer httpServer.Close()
	waitHTTPStatus(t, httpServer.Client(), httpServer.URL+"/readyz", http.StatusNoContent)

	createdResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/account", "", `{}`)
	if createdResponse.StatusCode != http.StatusCreated {
		t.Fatalf("account create status=%d body=%s", createdResponse.StatusCode, responseBody(createdResponse))
	}
	var created account.CreatedAccount
	decodeCompositionResponse(t, createdResponse, &created)
	sessionResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session", "",
		fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode))
	if sessionResponse.StatusCode != http.StatusOK {
		t.Fatalf("session status=%d body=%s", sessionResponse.StatusCode, responseBody(sessionResponse))
	}
	var tokens account.TokenPair
	decodeCompositionResponse(t, sessionResponse, &tokens)

	// Buy the server-resolved fiscal unlock through the public intent surface;
	// no test-only Founder mutation may satisfy the Pitch authorization seam.
	clock.Set(now.Add(300 * time.Millisecond))
	harvest := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken,
		`{"intent_id":"01986666-f100-7000-8000-000000000002","kind":"harvest_fiscal_period","expected_revision":1}`)
	if harvest.StatusCode != http.StatusOK {
		t.Fatalf("fiscal harvest status=%d body=%s", harvest.StatusCode, responseBody(harvest))
	}
	_ = responseBody(harvest)
	spend := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken,
		`{"intent_id":"01986666-f100-7000-8000-000000000003","kind":"spend_fiscal_credit","expected_revision":2,"target":{"kind":"unlock","unlock_id":"minigame.pitch"}}`)
	if spend.StatusCode != http.StatusOK {
		t.Fatalf("fiscal unlock status=%d body=%s", spend.StatusCode, responseBody(spend))
	}
	_ = responseBody(spend)
	// Founder coordinators use the transaction's database timestamp. Rejoin the
	// injected application clock to that authority after the deliberately old
	// account-creation baseline used to make the Fiscal period immediately ripe.
	clock.Set(time.Now().UTC().Add(time.Second))

	registry, err := account.PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	createBody := `{"idempotency_key":"pitch-create-1"}`
	createResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/minigames/pitch/sessions", tokens.AccessToken, createBody)
	createdBytes := readCompositionBytes(t, createResponse)
	if createResponse.StatusCode != http.StatusOK || registry.ValidateResponse("create_minigame_session", http.StatusOK, createdBytes) != nil {
		t.Fatalf("create status=%d body=%s", createResponse.StatusCode, createdBytes)
	}
	var current minigameAPIEnvelope
	if json.Unmarshal(createdBytes, &current) != nil || current.Status != "active" || current.Revision != 1 || current.SessionID == "" {
		t.Fatalf("create envelope=%s", createdBytes)
	}

	// Reconnect is a read through the composed adapter, not a session-local
	// object retained by this test client.
	currentResponse := compositionRequest(t, httpServer.Client(), http.MethodGet, httpServer.URL+"/api/v1/minigames/sessions/current", tokens.AccessToken, "")
	currentBytes := readCompositionBytes(t, currentResponse)
	if currentResponse.StatusCode != http.StatusOK || registry.ValidateResponse("get_current_minigame_session", http.StatusOK, currentBytes) != nil {
		t.Fatalf("current status=%d body=%s", currentResponse.StatusCode, currentBytes)
	}
	var reconnect struct {
		Kind    string              `json:"kind"`
		Session minigameAPIEnvelope `json:"session"`
	}
	if json.Unmarshal(currentBytes, &reconnect) != nil || reconnect.Kind != "active" || reconnect.Session.SessionID != current.SessionID || reconnect.Session.Revision != 1 {
		t.Fatalf("reconnect=%s", currentBytes)
	}

	var terminalBytes, terminalCommand []byte
	for step := 0; step < 24 && current.Status == "active"; step++ {
		var command map[string]any
		switch current.Snapshot.Phase {
		case "playing":
			count := 4
			if len(current.Snapshot.Hand) < count {
				count = len(current.Snapshot.Hand)
			}
			if count == 0 {
				t.Fatal("active Pitch hand is empty")
			}
			command = map[string]any{"kind": "play_hand", "card_ids": current.Snapshot.Hand[:count]}
		case "shop":
			command = map[string]any{"kind": "end_shop"}
		default:
			t.Fatalf("active Pitch phase=%q", current.Snapshot.Phase)
		}
		requestValue := map[string]any{"command_id": fmt.Sprintf("pitch-command-%02d", step+1), "expected_revision": current.Revision, "command": command}
		terminalCommand, err = json.Marshal(requestValue)
		if err != nil {
			t.Fatal(err)
		}
		commandResponse := compositionRequest(t, httpServer.Client(), http.MethodPost,
			httpServer.URL+"/api/v1/minigames/sessions/"+current.SessionID+"/commands", tokens.AccessToken, string(terminalCommand))
		responseBytes := readCompositionBytes(t, commandResponse)
		validationErr := registry.ValidateResponse("play_minigame_command", http.StatusOK, responseBytes)
		if commandResponse.StatusCode != http.StatusOK || validationErr != nil {
			t.Fatalf("command %d status=%d validation=%v body=%s", step, commandResponse.StatusCode, validationErr, responseBytes)
		}
		if json.Unmarshal(responseBytes, &current) != nil {
			t.Fatalf("command %d response=%s", step, responseBytes)
		}
		if current.Status == "resolved" {
			terminalBytes = responseBytes
		}
	}
	if len(terminalBytes) == 0 || len(current.ResolutionReceipt) == 0 {
		t.Fatalf("Pitch did not auto-resolve: %+v", current)
	}

	// Command retry, explicit resolve, and create retry all return the exact
	// durable response bytes rather than re-executing tenant or payout logic.
	commandRetry := compositionRequest(t, httpServer.Client(), http.MethodPost,
		httpServer.URL+"/api/v1/minigames/sessions/"+current.SessionID+"/commands", tokens.AccessToken, string(terminalCommand))
	if retryBytes := readCompositionBytes(t, commandRetry); commandRetry.StatusCode != http.StatusOK || !bytes.Equal(retryBytes, terminalBytes) {
		t.Fatalf("terminal command retry status=%d equal=%v body=%s", commandRetry.StatusCode, bytes.Equal(retryBytes, terminalBytes), retryBytes)
	}
	resolveResponse := compositionRequest(t, httpServer.Client(), http.MethodPost,
		httpServer.URL+"/api/v1/minigames/sessions/"+current.SessionID+"/resolve", tokens.AccessToken, `{}`)
	resolveBytes := readCompositionBytes(t, resolveResponse)
	if resolveResponse.StatusCode != http.StatusOK || !bytes.Equal(resolveBytes, terminalBytes) || registry.ValidateResponse("resolve_minigame_session", http.StatusOK, resolveBytes) != nil {
		t.Fatalf("resolve status=%d equal=%v body=%s", resolveResponse.StatusCode, bytes.Equal(resolveBytes, terminalBytes), resolveBytes)
	}
	createRetry := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/minigames/pitch/sessions", tokens.AccessToken, createBody)
	if retryBytes := readCompositionBytes(t, createRetry); createRetry.StatusCode != http.StatusOK || !bytes.Equal(retryBytes, createdBytes) {
		t.Fatalf("create retry status=%d equal=%v body=%s", createRetry.StatusCode, bytes.Equal(retryBytes, createdBytes), retryBytes)
	}
	currentResponse = compositionRequest(t, httpServer.Client(), http.MethodGet, httpServer.URL+"/api/v1/minigames/sessions/current", tokens.AccessToken, "")
	currentBytes = readCompositionBytes(t, currentResponse)
	if currentResponse.StatusCode != http.StatusOK || string(currentBytes) != `{"kind":"none"}` {
		t.Fatalf("terminal current status=%d body=%s", currentResponse.StatusCode, currentBytes)
	}

	drainContext, cancelDrain := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelDrain()
	if err := composition.Server.Drain(drainContext, clock.Time()); err != nil {
		t.Fatal(err)
	}
}

type minigameAPIEnvelope struct {
	SessionID         string           `json:"session_id"`
	Revision          int64            `json:"revision"`
	Status            string           `json:"status"`
	Snapshot          pitchAPISnapshot `json:"snapshot"`
	ResolutionReceipt json.RawMessage  `json:"resolution_receipt"`
}

type pitchAPISnapshot struct {
	Phase string   `json:"phase"`
	Hand  []string `json:"hand"`
}

func readCompositionBytes(t *testing.T, response *http.Response) []byte {
	t.Helper()
	defer response.Body.Close()
	var data bytes.Buffer
	if _, err := data.ReadFrom(response.Body); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func composedMinigameRepositoryRoot(t *testing.T, repositoryRoot string) string {
	t.Helper()
	base, err := epochseed.Load(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(base.Artifacts)+9)
	for name, data := range base.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["economy"] = composedEconomyCatalog(t, repositoryRoot)
	artifacts["routes"] = readCompositionFixture(t, repositoryRoot, "balance/testdata/permits-t3-gate-candidate-v1.json")
	artifacts["categories"] = composedCategoryCatalog(t, artifacts["categories"])
	artifacts["meters"] = compositionFixtureBaseline(t, repositoryRoot, "balance/testdata/meters-catalog-parity-v1.json")
	artifacts["achievements"] = []byte(`{"schema_version":1,"achievements":[]}`)
	artifacts["doctrines"] = compositionFixtureBaseline(t, repositoryRoot, "balance/testdata/doctrines-catalog-parity-v1.json")
	artifacts["minigames"] = readCompositionFixture(t, repositoryRoot, "testdata/minigame/pitch-v3.json")
	artifacts["pets"] = []byte(composedPetCatalogV2)
	artifacts["fiscal"] = composedFiscalCatalog(t, repositoryRoot)
	artifacts["soul"] = composedSoulCatalog(t, repositoryRoot)
	artifacts["pitch"] = readCompositionFixture(t, repositoryRoot, "balance/testdata/pitch-v1.json")
	artifacts["minigame_api"] = readCompositionFixture(t, repositoryRoot, "balance/testdata/minigame-api-candidate-v1.json")
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := replaycatalog.Load(hash, artifacts); err != nil {
		t.Fatalf("composed minigame candidate: %v", err)
	}

	root := t.TempDir()
	names := make([]string, 0, len(artifacts))
	for name := range artifacts {
		names = append(names, name)
	}
	sort.Strings(names)
	seed := epochseed.Seed{SchemaVersion: 1, CurrentEpochID: 1,
		Epochs: []epochseed.Epoch{{ID: 1, Name: "Composed minigame candidate", ChangelogRef: "changelog/epoch-1.md", AcceptedHashes: []string{hash}}}}
	for _, name := range names {
		path := "balance/composed/" + name + ".json"
		seed.Artifacts = append(seed.Artifacts, epochseed.Artifact{Name: name, Path: path})
		writeCompositionFixture(t, root, path, artifacts[name])
	}
	seedBytes, err := json.MarshalIndent(seed, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeCompositionFixture(t, root, epochseed.Path, append(seedBytes, '\n'))
	writeCompositionFixture(t, root, "changelog/epoch-1.md", []byte("# Composed minigame candidate\n"))
	writeCompositionFixture(t, root, "moderation/guild-names.txt", readCompositionFixture(t, repositoryRoot, "moderation/guild-names.txt"))
	writeCompositionFixture(t, root, "balance/transport/phase0.json", readCompositionFixture(t, repositoryRoot, "balance/transport/phase0.json"))
	return root
}

func composedEconomyCatalog(t *testing.T, repositoryRoot string) []byte {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(readCompositionFixture(t, repositoryRoot, "balance/testdata/valid/permits-economy-candidate-v1.json"), &value); err != nil {
		t.Fatal(err)
	}
	sources, ok := value["multiplier_sources"].([]any)
	if !ok {
		t.Fatal("economy candidate has no multiplier sources")
	}
	value["multiplier_sources"] = append(sources,
		map[string]any{"id": "fiscal.generator.beige_tower", "slot": "prestige", "target": "generator.beige_tower", "provider": "fiscal"},
		map[string]any{"id": "fiscal.hoard", "slot": "prestige", "target": "all", "provider": "fiscal"},
	)
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func composedCategoryCatalog(t *testing.T, data []byte) []byte {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	value["full_gate_set"] = []string{"gate.t2_to_t3", "gate.t3_to_t4", "gate.t4_to_t5", "gate.t7_to_t8"}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func composedFiscalCatalog(t *testing.T, repositoryRoot string) []byte {
	t.Helper()
	data := compositionFixtureBaseline(t, repositoryRoot, "balance/testdata/fiscal-foundation-v1.json")
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	rows, ok := value["unlock_rows"].([]any)
	if !ok {
		t.Fatal("Fiscal fixture has no unlock rows")
	}
	value["unlock_rows"] = append([]any{map[string]any{"unlock_id": "minigame.pitch", "cost": float64(3)}}, rows...)
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func composedSoulCatalog(t *testing.T, repositoryRoot string) []byte {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(readCompositionFixture(t, repositoryRoot, "testdata/soul/recovery-activities-v1.json"), &value); err != nil {
		t.Fatal(err)
	}
	value["debit_sources"] = []any{map[string]any{
		"source_id": "soul.event.fixture", "owner_kind": "event", "amount": float64(20),
		"may_exhaust": true, "single_use": true, "curtain_copy_key": "category.valuation",
	}}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func compositionFixtureBaseline(t *testing.T, repositoryRoot, path string) []byte {
	t.Helper()
	var wrapper struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if err := json.Unmarshal(readCompositionFixture(t, repositoryRoot, path), &wrapper); err != nil || len(wrapper.Baseline) == 0 {
		t.Fatalf("fixture %s has no baseline: %v", path, err)
	}
	return append([]byte(nil), wrapper.Baseline...)
}

func readCompositionFixture(t *testing.T, root, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeCompositionFixture(t *testing.T, root, path string, data []byte) {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
