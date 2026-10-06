package gameserver

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

// SG5/SG8/SG9: actual authenticated HTTP over the complete fixture-only
// composition. No direct Founder state grant, clock acceleration or public mint.
// An immature harvest is a refusal witness, not successful payout evidence.
func TestComposedGardenAuthenticatedCommandsAndReadIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; composed Garden proof not executed")
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
	const clean = `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`
	if _, err := db.ExecContext(ctx, clean); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), clean); err != nil {
			t.Errorf("clean Garden HTTP database: %v", err)
		}
	})

	clock := &mutableClock{now: time.Now().UTC().Add(-time.Second)}
	composition, err := Compose(ctx, CompositionConfig{
		DB: db, RepositoryRoot: composedGardenRepositoryRoot(t, filepathRoot(t)),
		ServerID: "01986666-f110-4000-8000-000000000001", ActivityBracket: "activity.standard",
		PublicCursorKeys: testPublicCursorKeys(), Clock: clock.Time,
		SigningKeys:   account.SigningKeys{CurrentID: "garden-http", Current: bytes.Repeat([]byte{0x67}, 32)},
		BootstrapKeys: account.BootstrapReceiptKeys{CurrentID: "garden-bootstrap", Current: bytes.Repeat([]byte{0x68}, 32)},
	})
	if err != nil {
		t.Fatal(err)
	}
	serverContext, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := composition.Server.Start(serverContext); err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(composition.Server.Handler())
	defer httpServer.Close()
	waitHTTPStatus(t, httpServer.Client(), httpServer.URL+"/readyz", http.StatusNoContent)
	registry, err := account.PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}

	createAccount := func() (account.CreatedAccount, account.TokenPair, string) {
		t.Helper()
		clock.Set(time.Now().UTC().Add(-time.Second))
		response := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/account", "", `{}`)
		if response.StatusCode != http.StatusCreated {
			t.Fatalf("create account status=%d body=%s", response.StatusCode, responseBody(response))
		}
		var created account.CreatedAccount
		decodeCompositionResponse(t, response, &created)
		clock.Set(time.Now().UTC())
		response = compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session", "",
			fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode))
		if response.StatusCode != http.StatusOK {
			t.Fatalf("create session status=%d body=%s", response.StatusCode, responseBody(response))
		}
		var tokens account.TokenPair
		decodeCompositionResponse(t, response, &tokens)
		founder, err := composition.Accounts.ActiveFounder(ctx, created.AccountID)
		if err != nil {
			t.Fatal(err)
		}
		return created, tokens, founder.ID
	}
	first, firstTokens, firstFounder := createAccount()
	_, secondTokens, secondFounder := createAccount()
	owners := []any{firstFounder, secondFounder}
	store, err := save.NewStore(db, composition.Catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, initial := range gardenHTTPHeads(t, db, store, owners) {
		if initial.Key.Scope != economy.ScopeFounder {
			continue
		}
		state := initial.State
		if save.VersionForState(state) != 25 || state.ServerGarden == nil || state.FiscalCredit != 0 ||
			len(state.ServerGarden.Plots) != 0 || state.ServerGarden.SaltHex != nil || state.ServerGarden.TickAnchorWallMS != nil ||
			!reflect.DeepEqual(state.ServerGarden.SeedCollection, []string{"strain_a", "strain_b"}) {
			t.Fatal("HTTP account genesis is not an empty v25 starter Garden with zero Fiscal credit")
		}
	}
	publicPayloads := [][]byte{}
	salt := ""
	get := func(token, suffix, body string, status int) []byte {
		t.Helper()
		before := gardenHTTPPersistence(t, db, owners)
		low := gardenHTTPDatabaseMS(t, db)
		response := compositionRequest(t, httpServer.Client(), http.MethodGet, httpServer.URL+"/api/v1/garden/current"+suffix, token, body)
		encoded := readCompositionBytes(t, response)
		high := gardenHTTPDatabaseMS(t, db)
		if response.StatusCode != status {
			t.Fatalf("Garden GET status=%d want=%d body=%s", response.StatusCode, status, encoded)
		}
		if err := registry.ValidateResponse("get_current_garden", status, encoded); err != nil {
			t.Fatalf("Garden response violates registry: %s: %v", encoded, err)
		}
		if after := gardenHTTPPersistence(t, db, owners); !reflect.DeepEqual(before, after) {
			t.Fatalf("Garden GET changed persisted rows: before=%v after=%v", before, after)
		}
		var view gardenHTTPView
		if err := json.Unmarshal(encoded, &view); err != nil {
			t.Fatal(err)
		}
		if view.Kind == "active" && (view.ServerMS < low || view.ServerMS > high) {
			t.Fatalf("served non-DB stamp %d outside [%d,%d]", view.ServerMS, low, high)
		}
		assertGardenHTTPPublicJSON(t, encoded, salt)
		publicPayloads = append(publicPayloads, encoded)
		return encoded
	}
	const locked = `{"kind":"locked","unlock_id":"minigame.server_garden"}`
	for _, token := range []string{firstTokens.AccessToken, secondTokens.AccessToken} {
		if got := get(token, "", "", http.StatusOK); string(got) != locked {
			t.Fatalf("new Founder is not exactly locked: %s", got)
		}
	}
	if got := get("", "", "", http.StatusUnauthorized); string(got) != "{\"category\":\"unauthorized\",\"detail\":\"access_token\"}\n" {
		t.Fatalf("unauthenticated read=%s", got)
	}
	if got := get(firstTokens.AccessToken, "", `{"founder_id":"forged"}`, http.StatusBadRequest); string(got) != "{\"category\":\"invalid\",\"detail\":\"body\"}\n" {
		t.Fatalf("GET body refusal=%s", got)
	}

	requestNumber := 0
	post := func(token, kind string, revision int64, fields string) (gardenHTTPReceipt, []byte, string) {
		t.Helper()
		requestNumber++
		body := fmt.Sprintf(`{"intent_id":"01986666-f110-7000-8000-%012d","kind":%q,"expected_revision":%d,%s}`, requestNumber, kind, revision, fields)
		response := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", token, body)
		encoded := readCompositionBytes(t, response)
		if response.StatusCode != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", kind, response.StatusCode, encoded)
		}
		var receipt gardenHTTPReceipt
		if err := json.Unmarshal(encoded, &receipt); err != nil || receipt.Outcome == "" {
			t.Fatalf("invalid receipt %s err=%v", encoded, err)
		}
		assertGardenHTTPPublicJSON(t, encoded, salt)
		publicPayloads = append(publicPayloads, encoded)
		return receipt, encoded, body
	}
	applied := func(receipt gardenHTTPReceipt, encoded []byte) int64 {
		t.Helper()
		if receipt.Outcome != "applied" || receipt.FounderRevision <= 1 {
			t.Fatalf("not an applied Founder command: %s", encoded)
		}
		return receipt.FounderRevision
	}
	receipt, encoded, _ := post(firstTokens.AccessToken, "spend_fiscal_credit", 1, `"target":{"kind":"unlock","unlock_id":"minigame.server_garden"}`)
	revision := applied(receipt, encoded)
	var automatic int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE kind=$1 AND payload->>'source'='automatic'
		AND (payload->>'credited')::bigint >= 3 AND stream_id IN(SELECT id FROM save_streams WHERE owner_id=$2)`,
		save.EventFiscalPeriodHarvested, firstFounder).Scan(&automatic); err != nil || automatic == 0 {
		t.Fatalf("unlock lacks actual automatic Fiscal funding: rows=%d err=%v", automatic, err)
	}
	var view gardenHTTPView
	if err := json.Unmarshal(get(firstTokens.AccessToken, "", "", http.StatusOK), &view); err != nil || view.Kind != "active" || view.FounderRevision != revision || len(view.Garden.Plots) != 0 {
		t.Fatalf("unlock did not reach active empty view: %+v err=%v", view, err)
	}
	receipt, encoded, plantBody := post(firstTokens.AccessToken, "garden_plant", revision, `"row":0,"col":0,"species_id":"strain_a"`)
	revision = applied(receipt, encoded)
	if err := json.Unmarshal(get(firstTokens.AccessToken, "", "", http.StatusOK), &view); err != nil || view.FounderRevision != revision || len(view.Garden.Plots) != 1 || view.Garden.Plots[0].SpeciesID != "strain_a" || view.Garden.Plots[0].Stage != "growing" || view.Garden.Plots[0].Row != 0 || view.Garden.Plots[0].Col != 0 {
		t.Fatalf("plant read=%+v err=%v", view, err)
	}
	beforeRetry := gardenHTTPPersistence(t, db, owners)
	retry := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", firstTokens.AccessToken, plantBody)
	retryBytes := readCompositionBytes(t, retry)
	if retry.StatusCode != http.StatusOK || !bytes.Equal(retryBytes, encoded) || !reflect.DeepEqual(beforeRetry, gardenHTTPPersistence(t, db, owners)) {
		t.Fatalf("HTTP retry is not byte-idempotent: status=%d receipt=%s", retry.StatusCode, retryBytes)
	}
	publicPayloads = append(publicPayloads, retryBytes)

	var firstStream string
	if err := db.QueryRowContext(ctx, `SELECT id::text FROM save_streams WHERE owner_kind='founder' AND owner_id=$1 AND scope='founder' AND archived_at IS NULL`, firstFounder).Scan(&firstStream); err != nil {
		t.Fatal(err)
	}
	stateBefore, err := store.LoadLatest(ctx, firstStream)
	if err != nil || stateBefore.State.ServerGarden.SaltHex == nil {
		t.Fatalf("public plant did not persist a hidden salt: %v", err)
	}
	salt = *stateBefore.State.ServerGarden.SaltHex
	// Once the actual server-drawn value is known, also check every response
	// already served. Later reads/receipts check immediately, before UI oracles.
	for _, payload := range publicPayloads {
		assertGardenHTTPPublicJSON(t, payload, salt)
	}
	refuse := func(token, kind string, rev int64, fields, category, detail string) {
		t.Helper()
		before := gardenHTTPHeads(t, db, store, owners)
		receipt, data, _ := post(token, kind, rev, fields)
		if receipt.Outcome != "rejected" || receipt.Rejection.Category != category || receipt.Rejection.Detail != detail {
			t.Fatalf("wrong refusal %s/%s: %s", category, detail, data)
		}
		after := gardenHTTPHeads(t, db, store, owners)
		if !reflect.DeepEqual(before, after) {
			t.Fatal("refused command changed an account's Founder or Company persisted head")
		}
	}
	refuse(firstTokens.AccessToken, "garden_harvest", revision, `"plots":[{"row":0,"col":0}]`, "not_eligible", "plant_not_mature")
	for _, field := range []string{`"tick_seq":9`, `"server_ms":1`, `"salt_hex":"0123456789abcdef"`, fmt.Sprintf(`"founder_id":%q`, secondFounder)} {
		refuse(firstTokens.AccessToken, "garden_plant", revision, `"row":0,"col":1,"species_id":"strain_b",`+field, "invalid", "garden_plant.fields")
	}
	refuse(secondTokens.AccessToken, "garden_plant", 1, `"row":0,"col":0,"species_id":"strain_a"`, "not_eligible", "fiscal_unlock_required")
	if got := get(secondTokens.AccessToken, "?account_id="+first.AccountID+"&founder_id="+firstFounder, "", http.StatusOK); string(got) != locked {
		t.Fatalf("query hints overrode authenticated owner: %s", got)
	}

	receipt, encoded, _ = post(firstTokens.AccessToken, "garden_set_substrate", revision, `"substrate_id":"containerized"`)
	revision = applied(receipt, encoded)
	if err := json.Unmarshal(get(firstTokens.AccessToken, "", "", http.StatusOK), &view); err != nil || view.FounderRevision != revision || view.Garden.SubstrateID != "containerized" {
		t.Fatalf("substrate read=%+v err=%v", view, err)
	}
	receipt, encoded, _ = post(firstTokens.AccessToken, "garden_uproot", revision, `"row":0,"col":0`)
	revision = applied(receipt, encoded)
	if err := json.Unmarshal(get(firstTokens.AccessToken, "", "", http.StatusOK), &view); err != nil || view.FounderRevision != revision || len(view.Garden.Plots) != 0 {
		t.Fatalf("uproot read=%+v err=%v", view, err)
	}
	if got := get(secondTokens.AccessToken, "", "", http.StatusOK); string(got) != locked {
		t.Fatalf("foreign Garden changed: %s", got)
	}
	var windows, credits int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM minigame_faucet_window WHERE founder_id IN($1,$2)`, owners...).Scan(&windows); err != nil || windows != 0 {
		t.Fatalf("refusal consumed faucet: windows=%d err=%v", windows, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE kind='garden_harvest_credited.v1'`).Scan(&credits); err != nil || credits != 0 {
		t.Fatalf("refusal credited Company: credits=%d err=%v", credits, err)
	}

	for _, query := range []string{
		`SELECT payload::text FROM events WHERE stream_id IN(SELECT id FROM save_streams WHERE owner_id IN($1,$2))`,
		`SELECT payload::text FROM transport_player_outbox WHERE founder_id IN($1,$2)`,
	} {
		rows, err := db.QueryContext(ctx, query, owners...)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			publicPayloads = append(publicPayloads, payload)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		_ = rows.Close()
	}
	for _, payload := range publicPayloads {
		assertGardenHTTPPublicJSON(t, payload, salt)
	}
	for _, owner := range owners {
		var streamID string
		if err := db.QueryRowContext(ctx, `SELECT id::text FROM save_streams WHERE owner_id=$1 AND scope='founder' AND archived_at IS NULL`, owner).Scan(&streamID); err != nil {
			t.Fatal(err)
		}
		history, err := store.LoadFounderHistory(ctx, streamID)
		if err != nil {
			t.Fatal(err)
		}
		if verdict := production.VerifyFounderHistory(history, composition.Catalogs.replay); verdict != production.ReplayVerified {
			t.Fatalf("HTTP-created Founder history verdict=%v", verdict)
		}
	}
	t.Logf("two HTTP accounts; unlock/plant/substrate/uproot applied; immature/authority/foreign refusals; %d public payloads salt-free; private replay inputs excluded; both Founder histories verified", len(publicPayloads))
	drainContext, cancelDrain := context.WithTimeout(ctx, 2*time.Second)
	defer cancelDrain()
	if err := composition.Server.Drain(drainContext, clock.Time()); err != nil {
		t.Fatal(err)
	}
}

type gardenHTTPReceipt struct {
	Outcome         string `json:"outcome"`
	FounderRevision int64  `json:"founder_revision"`
	Rejection       struct {
		Category string `json:"category"`
		Detail   string `json:"detail"`
	} `json:"rejection"`
}

// These are partial assertion readers, not a second response schema; every
// complete response above must validate against the production API registry.
type gardenHTTPView struct {
	Kind            string `json:"kind"`
	FounderRevision int64  `json:"founder_revision"`
	ServerMS        int64  `json:"server_ms"`
	Garden          struct {
		SubstrateID string `json:"substrate_id"`
		Plots       []struct {
			Row       int64  `json:"row"`
			Col       int64  `json:"col"`
			SpeciesID string `json:"species_id"`
			Stage     string `json:"stage"`
		} `json:"plots"`
	} `json:"garden"`
}

func gardenHTTPDatabaseMS(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	var value int64
	if err := db.QueryRow(`SELECT (extract(epoch FROM clock_timestamp())*1000)::bigint`).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func gardenHTTPHeads(t *testing.T, db *sql.DB, store *save.Store, owners []any) map[string]save.Loaded {
	t.Helper()
	rows, err := db.Query(`SELECT id::text FROM save_streams WHERE owner_id IN($1,$2) ORDER BY id`, owners...)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	result := map[string]save.Loaded{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		loaded, err := store.LoadLatest(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		result[id] = loaded
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(result) != 4 {
		t.Fatalf("expected both account's Founder and Company heads, got %d", len(result))
	}
	return result
}

func gardenHTTPPersistence(t *testing.T, db *sql.DB, owners []any) map[string]string {
	t.Helper()
	result := map[string]string{}
	queries := map[string]string{
		"streams":     `SELECT * FROM save_streams WHERE owner_id IN($1,$2)`,
		"revisions":   `SELECT * FROM save_revisions WHERE stream_id IN(SELECT id FROM save_streams WHERE owner_id IN($1,$2))`,
		"founder_log": `SELECT * FROM founder_log WHERE founder_stream_id IN(SELECT id FROM save_streams WHERE owner_id IN($1,$2))`,
		"company_log": `SELECT * FROM run_log WHERE company_stream_id IN(SELECT id FROM save_streams WHERE owner_id IN($1,$2))`,
		"events":      `SELECT * FROM events WHERE stream_id IN(SELECT id FROM save_streams WHERE owner_id IN($1,$2))`,
		"intents":     `SELECT * FROM intent_records WHERE stream_id IN(SELECT id FROM save_streams WHERE owner_id IN($1,$2))`,
		"windows":     `SELECT * FROM minigame_faucet_window WHERE founder_id IN($1,$2)`,
		// Delivery leases/publication timestamps may move independently; compare
		// every immutable identity/message field and payload, not worker metadata.
		"outbox": `SELECT founder_id,stream_id,message_kind,source_id,scope,revision,constants_hash,payload FROM transport_player_outbox WHERE founder_id IN($1,$2)`,
	}
	for name, query := range queries {
		statement := `SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text),'[]'::jsonb)::text FROM (` + query + `) r`
		var value string
		if err := db.QueryRow(statement, owners...).Scan(&value); err != nil {
			t.Fatalf("snapshot %s: %v", name, err)
		}
		result[name] = value
	}
	return result
}

func assertGardenHTTPPublicJSON(t *testing.T, payload []byte, salt string) {
	t.Helper()
	if salt != "" && strings.Contains(string(payload), salt) {
		t.Fatalf("actual hidden salt leaked in public JSON: %s", payload)
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	var visit func(any)
	visit = func(value any) {
		switch node := value.(type) {
		case map[string]any:
			for key, child := range node {
				switch key {
				case "salt_hex", "garden_salt_hex", "garden_base", "rng_s", "draw", "draws":
					t.Fatalf("hidden Garden key %q in public JSON: %s", key, payload)
				}
				visit(child)
			}
		case []any:
			for _, child := range node {
				visit(child)
			}
		}
	}
	visit(value)
}

func composedGardenRepositoryRoot(t *testing.T, repositoryRoot string) string {
	t.Helper()
	var corpus struct {
		Bundles map[string]struct {
			ConstantsHash string            `json:"constants_hash"`
			Artifacts     map[string]string `json:"artifacts"`
		} `json:"bundles"`
	}
	if err := json.Unmarshal(readCompositionFixture(t, repositoryRoot, "testdata/replay/garden-v1.json"), &corpus); err != nil {
		t.Fatal(err)
	}
	grown, ok := corpus.Bundles["grown"]
	if !ok || len(grown.Artifacts) == 0 {
		t.Fatal("missing exact grown Garden fixture")
	}
	artifacts := map[string][]byte{}
	names := []string{}
	for name, data := range grown.Artifacts {
		artifacts[name] = []byte(data)
		names = append(names, name)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil || hash != grown.ConstantsHash {
		t.Fatalf("Garden fixture identity mismatch: %s err=%v", hash, err)
	}
	if _, err := replaycatalog.Load(hash, artifacts); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	sort.Strings(names)
	seed := epochseed.Seed{SchemaVersion: 1, CurrentEpochID: 1, Epochs: []epochseed.Epoch{{ID: 1, Name: "Composed Garden fixture only", ChangelogRef: "changelog/epoch-1.md", AcceptedHashes: []string{hash}}}}
	for _, name := range names {
		path := "balance/composed/" + name + ".json"
		seed.Artifacts = append(seed.Artifacts, epochseed.Artifact{Name: name, Path: path})
		writeCompositionFixture(t, root, path, artifacts[name])
	}
	encoded, err := json.Marshal(seed)
	if err != nil {
		t.Fatal(err)
	}
	writeCompositionFixture(t, root, epochseed.Path, encoded)
	writeCompositionFixture(t, root, "changelog/epoch-1.md", []byte("# Garden test fixture; not a production mint\n"))
	for _, path := range []string{"balance/api/phase0.json", "moderation/guild-names.txt", "balance/transport/phase0.json"} {
		writeCompositionFixture(t, root, path, readCompositionFixture(t, repositoryRoot, path))
	}
	return root
}
