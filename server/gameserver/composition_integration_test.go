package gameserver

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
	"cloud-clicker/server/transport"

	"github.com/coder/websocket"
)

type mutableClock struct {
	mu  sync.Mutex
	now time.Time
}

func (clock *mutableClock) Time() time.Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.now
}

func (clock *mutableClock) Set(now time.Time) {
	clock.mu.Lock()
	clock.now = now
	clock.mu.Unlock()
}

type wsReply struct {
	ID        uint32          `json:"id"`
	Error     *wsError        `json:"error"`
	Connect   json.RawMessage `json:"connect"`
	Subscribe json.RawMessage `json:"subscribe"`
	Push      *wsPush         `json:"push"`
}

type wsError struct {
	Code uint32 `json:"code"`
}

type wsPush struct {
	Channel     string         `json:"channel"`
	Publication *wsPublication `json:"pub"`
}

type wsPublication struct {
	Data json.RawMessage `json:"data"`
}

type compositionSocket struct {
	connection *websocket.Conn
	pending    []wsReply
}

type bootstrapResponseEnvelope struct {
	Account        account.CreatedAccount `json:"account"`
	Session        account.TokenPair      `json:"session"`
	GameUISnapshot json.RawMessage        `json:"game_ui_snapshot"`
}

func generatorIDs(rows []struct {
	ID string `json:"generator_id"`
}) []string {
	ids := make([]string, len(rows))
	for index := range rows {
		ids[index] = rows[index].ID
	}
	return ids
}

func TestComposedGameserverPostgresSocketClearingAndGCIntegration(t *testing.T) {
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
	const cleanDatabase = `TRUNCATE bootstrap_receipts,accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`
	if _, err := db.ExecContext(ctx, cleanDatabase); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), cleanDatabase); err != nil {
			t.Errorf("clean composition database: %v", err)
		}
	})

	// Centrifuge validates Credentials.ExpireAt against the process wall clock.
	// Keep the injected application clock deterministic within the test while
	// anchoring it to the day the test runs, so this real-socket proof cannot
	// expire merely because its authored calendar date has passed.
	clock := &mutableClock{now: time.Now().UTC().Truncate(time.Second)}
	composition, err := Compose(ctx, CompositionConfig{
		DB:              db,
		RepositoryRoot:  filepathRoot(t),
		ServerID:        "018f0000-0000-4000-8000-000000000301",
		ActivityBracket: "activity.standard",
		SigningKeys: account.SigningKeys{
			CurrentID: "composition-integration",
			Current:   bytes.Repeat([]byte{0x42}, 32),
		},
		BootstrapKeys: account.BootstrapReceiptKeys{CurrentID: "bootstrap-integration", Current: bytes.Repeat([]byte{0x43}, 32)},
		Clock:         clock.Time,
	})
	if err != nil {
		t.Fatal(err)
	}
	if composition.Minigames == nil {
		t.Fatal("composed gameserver omitted the minigame platform")
	}
	serverContext, cancelServer := context.WithCancel(ctx)
	defer cancelServer()
	if err := composition.Server.Start(serverContext); err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(composition.Server.Handler())
	defer httpServer.Close()
	waitHTTPStatus(t, httpServer.Client(), httpServer.URL+"/readyz", http.StatusNoContent)

	invalidBootstrap := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/bootstrap", "",
		`{"idempotency_key":"guessable"}`)
	if invalidBootstrap.StatusCode != http.StatusBadRequest || responseBody(invalidBootstrap) != "{\"category\":\"invalid\",\"detail\":\"bootstrap\"}\n" {
		t.Fatal("invalid bootstrap key did not fail closed")
	}
	bootstrapKey := strings.Repeat("ab", 32)
	bootstrapResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/bootstrap", "",
		fmt.Sprintf(`{"idempotency_key":%q}`, bootstrapKey))
	if bootstrapResponse.StatusCode != http.StatusCreated || bootstrapResponse.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("bootstrap status=%d cache=%q body=%s", bootstrapResponse.StatusCode, bootstrapResponse.Header.Get("Cache-Control"), responseBody(bootstrapResponse))
	}
	var bootstrap bootstrapResponseEnvelope
	decodeCompositionResponse(t, bootstrapResponse, &bootstrap)
	created, tokens := bootstrap.Account, bootstrap.Session
	retryResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/bootstrap", "",
		fmt.Sprintf(`{"idempotency_key":%q}`, bootstrapKey))
	var retry bootstrapResponseEnvelope
	if retryResponse.StatusCode != http.StatusCreated || retryResponse.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("bootstrap retry status=%d cache=%q body=%s", retryResponse.StatusCode, retryResponse.Header.Get("Cache-Control"), responseBody(retryResponse))
	}
	decodeCompositionResponse(t, retryResponse, &retry)
	firstEncoded, _ := json.Marshal(bootstrap)
	retryEncoded, _ := json.Marshal(retry)
	if !bytes.Equal(firstEncoded, retryEncoded) {
		t.Fatalf("bootstrap retry changed receipt\nfirst=%s\nretry=%s", firstEncoded, retryEncoded)
	}
	uiResponse := compositionRequest(t, httpServer.Client(), http.MethodGet, httpServer.URL+"/api/v1/founder/state", tokens.AccessToken, "")
	if uiResponse.StatusCode != http.StatusOK {
		company, companyErr := composition.Accounts.ActiveCompanyState(ctx, created.AccountID)
		if companyErr == nil {
			_, projectionErr := composition.GameUI.GameUISnapshot(ctx, company.StreamID, clock.Time())
			t.Fatalf("Game UI snapshot status=%d body=%s projection_err=%v", uiResponse.StatusCode, responseBody(uiResponse), projectionErr)
		}
		t.Fatalf("Game UI snapshot status=%d body=%s", uiResponse.StatusCode, responseBody(uiResponse))
	}
	var uiSnapshot struct {
		ConstantsHash    string `json:"constants_hash"`
		EvaluatedThrough int64  `json:"evaluated_through_ms"`
		Generators       []struct {
			ID string `json:"generator_id"`
		} `json:"generators"`
		ManualAction struct {
			ID string `json:"action_id"`
		} `json:"manual_action"`
		Revision int64 `json:"revision"`
		Run      struct {
			FounderID string `json:"founder_id"`
			RunSeq    int64  `json:"run_seq"`
		} `json:"run"`
	}
	uiBytes := []byte(responseBody(uiResponse))
	if err := json.Unmarshal(uiBytes, &uiSnapshot); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bootstrap.GameUISnapshot, uiBytes) {
		t.Fatalf("transaction-local bootstrap snapshot differs from first committed read\nbootstrap=%s\ncommitted=%s", bootstrap.GameUISnapshot, uiBytes)
	}
	var bootstrapSnapshot struct {
		ConstantsHash string `json:"constants_hash"`
		Revision      int64  `json:"revision"`
	}
	if json.Unmarshal(bootstrap.GameUISnapshot, &bootstrapSnapshot) != nil || bootstrapSnapshot.ConstantsHash != composition.CurrentHash || bootstrapSnapshot.Revision != 1 {
		t.Fatalf("bootstrap Game UI snapshot=%+v", bootstrapSnapshot)
	}
	founder, err := composition.Accounts.ActiveFounder(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	if uiSnapshot.ConstantsHash != composition.CurrentHash || uiSnapshot.EvaluatedThrough != clock.Time().UnixMilli() ||
		uiSnapshot.Revision != 1 || uiSnapshot.Run.FounderID != founder.ID || uiSnapshot.Run.RunSeq != 1 ||
		uiSnapshot.ManualAction.ID != "manual.click" || !reflect.DeepEqual(generatorIDs(uiSnapshot.Generators), []string{
		"generator.answering_machine", "generator.beige_tower", "generator.beige_tower_v2", "generator.crt_wall",
		"generator.dot_matrix_queue", "generator.first_hire", "generator.garage_rack", "generator.legal_dept", "generator.nephew_intern",
	}) {
		t.Fatalf("Game UI snapshot=%+v", uiSnapshot)
	}
	if composition.CurrentHash != "sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6" {
		t.Fatalf("composed constants hash=%s", composition.CurrentHash)
	}
	currentBundle, ok := composition.Catalogs.bundle(composition.CurrentHash)
	if !ok {
		t.Fatal("current replay bundle unavailable")
	}
	if currentBundle.Curriculum == nil {
		t.Fatal("epoch-8 current replay bundle omitted curriculum")
	}
	activeCompany, err := composition.Accounts.ActiveCompanyState(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	companyState, err := save.RestoreState(activeCompany.State, activeCompany.Version, currentBundle.Economy, economy.ScopeCompany, clock.Time())
	if err != nil {
		t.Fatal(err)
	}
	var founderVersion int
	var founderHash string
	var founderBytes []byte
	if err := db.QueryRowContext(ctx, `
		SELECT r.version,r.constants_hash,r.state
		FROM save_streams s
		JOIN LATERAL (SELECT version,constants_hash,state FROM save_revisions WHERE stream_id=s.id ORDER BY revision DESC LIMIT 1) r ON true
		WHERE s.owner_kind='founder' AND s.owner_id=$1 AND s.scope='founder' AND s.archived_at IS NULL`, founder.ID).
		Scan(&founderVersion, &founderHash, &founderBytes); err != nil {
		t.Fatal(err)
	}
	founderState, err := save.RestoreState(founderBytes, founderVersion, currentBundle.Economy, economy.ScopeFounder, clock.Time())
	if err != nil {
		t.Fatal(err)
	}
	if activeCompany.Version != 18 || activeCompany.ConstantsHash != composition.CurrentHash || founderVersion != 21 || founderHash != composition.CurrentHash {
		t.Fatalf("fresh activation company=(v%d,%s) founder=(v%d,%s)", activeCompany.Version, activeCompany.ConstantsHash, founderVersion, founderHash)
	}
	initialOpportunity, err := currentBundle.Opportunities.Spawn(founder.ID, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(companyState.MeterValues) != 11 || companyState.AchievementsEarnedRun == nil || companyState.NextOpportunityAttendedMS != initialOpportunity.SpawnedAttendedMS ||
		companyState.PendingOpportunity != nil || companyState.ActiveBuffs == nil ||
		len(founderState.MinigameRatings) != 1 || founderState.MinigameOfflineQuality == nil || founderState.Pets == nil ||
		founderState.FiscalPeriodOpenedWallMS == 0 || founderState.Soul != currentBundle.Soul.Policy.Initial {
		t.Fatalf("fresh epoch-7 state company=%+v founder=%+v", companyState, founderState)
	}
	if err := currentBundle.ValidateFoundationState(companyState); err != nil {
		t.Fatalf("fresh Company state: %v", err)
	}
	if err := currentBundle.ValidateFoundationState(founderState); err != nil {
		t.Fatalf("fresh Founder state: %v", err)
	}

	websocketURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/connection/websocket"
	player := dialCompositionSocket(t, websocketURL, httpServer.Client(), tokens.AccessToken)
	defer player.connection.CloseNow()
	writeWS(t, player, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "player:" + founder.ID}})
	if reply := readWSID(t, player, 2); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("player subscribe=%+v", reply)
	}

	world := dialCompositionSocket(t, websocketURL, httpServer.Client(), tokens.AccessToken)
	defer world.connection.CloseNow()
	writeWS(t, world, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "world"}})
	if reply := readWSID(t, world, 2); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("world subscribe=%+v", reply)
	}
	worldEnvelope := readEnvelope(t, world, "world")
	if worldEnvelope.Kind != "snapshot" || worldEnvelope.Revision < 1 {
		t.Fatalf("world envelope=%+v", worldEnvelope)
	}

	firstIntent := `{"intent_id":"01985555-1111-7111-8111-111111111111","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`
	intentResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken, firstIntent)
	if intentResponse.StatusCode != http.StatusOK {
		t.Fatalf("intent status=%d body=%s", intentResponse.StatusCode, responseBody(intentResponse))
	}
	_ = responseBody(intentResponse)
	firstEnvelope := readEnvelope(t, player, "player:"+founder.ID)
	if firstEnvelope.Revision != 2 || firstEnvelope.Kind != "receipt" {
		t.Fatalf("first player envelope=%+v", firstEnvelope)
	}
	signCompact := `{"intent_id":"01985555-1112-7111-8111-111111111112","kind":"sign_compact","expected_revision":2,"tithe_ppm":100000}`
	signResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken, signCompact)
	if signResponse.StatusCode != http.StatusOK {
		t.Fatalf("compact status=%d body=%s", signResponse.StatusCode, responseBody(signResponse))
	}
	_ = responseBody(signResponse)
	signKinds := map[string]bool{}
	for len(signKinds) < 2 {
		envelope := readEnvelope(t, player, "player:"+founder.ID)
		if envelope.Revision != 3 {
			t.Fatalf("compact player revision=%d", envelope.Revision)
		}
		signKinds[envelope.Kind] = true
	}
	if !signKinds["event"] || !signKinds["receipt"] {
		t.Fatalf("compact player message kinds=%v", signKinds)
	}
	cohortID, err := composition.Commons.FounderCohort(ctx, founder.ID)
	if err != nil {
		t.Fatal(err)
	}
	writeWS(t, player, map[string]any{"id": 3, "subscribe": map[string]any{"channel": "cohort:" + cohortID}})
	if reply := readWSID(t, player, 3); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("Commons resolver did not authorize member: %+v", reply)
	}

	guildResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/guild/intents", tokens.AccessToken,
		`{"intent_id":"018f0000-0000-7000-8000-000000000401","kind":"create_guild","expected_revision":1,"name":"Small Systems","join_policy":"open"}`)
	if guildResponse.StatusCode != http.StatusOK {
		t.Fatalf("guild status=%d body=%s", guildResponse.StatusCode, responseBody(guildResponse))
	}
	var guildReceipt struct {
		GuildID string `json:"guild_id"`
	}
	decodeCompositionResponse(t, guildResponse, &guildReceipt)
	writeWS(t, player, map[string]any{"id": 4, "subscribe": map[string]any{"channel": "guild:" + guildReceipt.GuildID}})
	if reply := readWSID(t, player, 4); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("guild resolver did not authorize member: %+v", reply)
	}
	writeWS(t, player, map[string]any{"id": 5, "subscribe": map[string]any{"channel": "match:018f0000-0000-7000-8000-000000000499"}})
	if reply := readWSID(t, player, 5); reply.Error == nil || reply.Error.Code != 103 {
		t.Fatalf("unowned match resolver did not fail closed: %+v", reply)
	}
	secondCreatedResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/account", "", `{}`)
	if secondCreatedResponse.StatusCode != http.StatusCreated {
		t.Fatalf("second account status=%d body=%s", secondCreatedResponse.StatusCode, responseBody(secondCreatedResponse))
	}
	var secondCreated account.CreatedAccount
	decodeCompositionResponse(t, secondCreatedResponse, &secondCreated)
	secondSessionResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session", "",
		fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, secondCreated.AccountID, secondCreated.RecoveryCode))
	if secondSessionResponse.StatusCode != http.StatusOK {
		t.Fatalf("second session status=%d body=%s", secondSessionResponse.StatusCode, responseBody(secondSessionResponse))
	}
	var secondTokens account.TokenPair
	decodeCompositionResponse(t, secondSessionResponse, &secondTokens)
	secondGuildResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/guild/intents", secondTokens.AccessToken,
		`{"intent_id":"018f0000-0000-7000-8000-000000000402","kind":"create_guild","expected_revision":1,"name":"Small Systems Two","join_policy":"open"}`)
	if secondGuildResponse.StatusCode != http.StatusOK {
		t.Fatalf("second guild status=%d body=%s", secondGuildResponse.StatusCode, responseBody(secondGuildResponse))
	}
	var secondGuildReceipt struct {
		GuildID string `json:"guild_id"`
	}
	decodeCompositionResponse(t, secondGuildResponse, &secondGuildReceipt)
	writeWS(t, player, map[string]any{"id": 6, "subscribe": map[string]any{"channel": "guild:" + secondGuildReceipt.GuildID}})
	if reply := readWSID(t, player, 6); reply.Error == nil || reply.Error.Code != 103 {
		t.Fatalf("non-member guild subscription did not fail closed: %+v", reply)
	}
	legacyCompany, err := composition.Accounts.ActiveCompanyState(ctx, secondCreated.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	factionBytes, err := os.ReadFile(filepath.Join(filepathRoot(t), "balance/factions/phase0.json"))
	if err != nil {
		t.Fatal(err)
	}
	factionBytes = bytes.Replace(factionBytes, []byte(`"stock_cap": 100000`), []byte(`"stock_cap": 100001`), 1)
	commonsCatalog, ok := composition.Catalogs.ResolveCommons(composition.CurrentHash)
	if !ok {
		t.Fatal("Commons tithe band unavailable")
	}
	historicalFaction, err := faction.LoadCatalog(factionBytes, faction.CompactTitheBand{
		MinimumPPM: commonsCatalog.MinimumTithePPM, DefaultPPM: commonsCatalog.DefaultTithePPM, MaximumPPM: commonsCatalog.MaximumTithePPM,
	})
	if err != nil {
		t.Fatal(err)
	}
	historicalBundle := currentBundle
	historicalBundle.Artifacts = make(map[string][]byte, len(currentBundle.Artifacts))
	for name, artifact := range currentBundle.Artifacts {
		historicalBundle.Artifacts[name] = append([]byte(nil), artifact...)
	}
	historicalBundle.Artifacts["factions"] = append([]byte(nil), factionBytes...)
	historicalHash, err := save.ConstantsHashArtifacts(historicalBundle.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	historicalBundle.ConstantsHash = historicalHash
	historicalBundle.Faction = historicalFaction
	composition.Catalogs.replay[historicalHash] = historicalBundle
	if _, err := db.ExecContext(ctx, `UPDATE save_revisions SET version=1,state='{"balances":{"company.cash":"0","company.permits":"0"}}'::jsonb,constants_hash=$3 WHERE stream_id=$1 AND revision=$2`,
		legacyCompany.StreamID, legacyCompany.Revision, historicalHash); err != nil {
		t.Fatal(err)
	}
	if _, stockCap, err := composition.Clearing.members(ctx, secondGuildReceipt.GuildID); err != nil || stockCap != historicalFaction.StockCap {
		t.Fatalf("historical v1 clearing cap=%d want=%d err=%v", stockCap, historicalFaction.StockCap, err)
	}
	thirdCreatedResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/account", "", `{}`)
	var thirdCreated account.CreatedAccount
	decodeCompositionResponse(t, thirdCreatedResponse, &thirdCreated)
	thirdSessionResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session", "",
		fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, thirdCreated.AccountID, thirdCreated.RecoveryCode))
	var thirdTokens account.TokenPair
	decodeCompositionResponse(t, thirdSessionResponse, &thirdTokens)
	thirdJoinResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/guild/intents", thirdTokens.AccessToken,
		fmt.Sprintf(`{"intent_id":"018f0000-0000-7000-8000-000000000404","kind":"join_guild","expected_revision":1,"guild_id":%q}`, guildReceipt.GuildID))
	if thirdJoinResponse.StatusCode != http.StatusOK {
		t.Fatalf("third member join status=%d body=%s", thirdJoinResponse.StatusCode, responseBody(thirdJoinResponse))
	}
	_ = responseBody(thirdJoinResponse)
	thirdFounder, err := composition.Accounts.ActiveFounder(ctx, thirdCreated.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	companyBeforeClearing, err := composition.Accounts.ActiveCompanyState(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	testStore, err := save.NewStore(db, composition.Catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}
	loadedCompany, err := testStore.LoadLatest(ctx, companyBeforeClearing.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	loadedCompany.State.FactionID = "vc_funded"
	loadedCompany.State.IncorporatedAt = clock.Time()
	if _, err := testStore.Write(ctx, loadedCompany.Revision.StreamID, loadedCompany.Revision.Number, loadedCompany.Revision.ConstantsHash,
		loadedCompany.State, save.WriteContext{Cause: "gameserver.outstanding-reservation.integration"}); err != nil {
		t.Fatal(err)
	}
	thirdCompany, err := composition.Accounts.ActiveCompanyState(ctx, thirdCreated.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	loadedThird, err := testStore.LoadLatest(ctx, thirdCompany.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	loadedThird.State.FactionID = "bootstrapper"
	loadedThird.State.IncorporatedAt = clock.Time()
	loadedThird.State.StockUnits = 10
	if _, err := testStore.Write(ctx, loadedThird.Revision.StreamID, loadedThird.Revision.Number, loadedThird.Revision.ConstantsHash,
		loadedThird.State, save.WriteContext{Cause: "gameserver.outstanding-reservation.integration"}); err != nil {
		t.Fatal(err)
	}

	before, err := composition.Guilds.PendingSettlements(ctx, founder.ID, "", 0)
	if err != nil || before.GuildID != guildReceipt.GuildID || before.BaseSeq != 0 || len(before.Settlements) != 0 {
		t.Fatalf("pre-clearing batch=%+v err=%v", before, err)
	}
	composition.Clearing.limit = 1
	for boundary := 1; boundary <= 3; boundary++ {
		if count, err := composition.Clearing.Tick(ctx); err != nil || count != 2 {
			t.Fatalf("clearing boundary=%d count=%d with paginated v1 member err=%v", boundary, count, err)
		}
	}
	after, err := composition.Guilds.PendingSettlements(ctx, founder.ID, "", 0)
	if err != nil || after.GuildID != guildReceipt.GuildID || len(after.Settlements) != 3 {
		t.Fatalf("post-clearing consumer batch=%+v err=%v", after, err)
	}
	afterProducer, err := composition.Guilds.PendingSettlements(ctx, thirdFounder.ID, "", 0)
	reservedDebit := int64(0)
	for _, settlement := range afterProducer.Settlements {
		reservedDebit += settlement.DebitUnits
	}
	if err != nil || afterProducer.GuildID != guildReceipt.GuildID || len(afterProducer.Settlements) != 3 || reservedDebit <= 0 || reservedDebit > 10 {
		t.Fatalf("post-clearing producer batch=%+v err=%v", afterProducer, err)
	}
	clock.Set(clock.Time().Add(time.Second))
	secondIntent := `{"intent_id":"01985555-1113-7111-8111-111111111113","kind":"perform_manual_batch","expected_revision":4,"action_id":"manual.click","count":1,"window_ms":1}`
	secondResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken, secondIntent)
	if secondResponse.StatusCode != http.StatusOK {
		t.Fatalf("post-clearing intent status=%d body=%s", secondResponse.StatusCode, responseBody(secondResponse))
	}
	_ = responseBody(secondResponse)
	postClearingKinds := map[string]bool{}
	for !postClearingKinds["receipt"] {
		envelope := readEnvelope(t, player, "player:"+founder.ID)
		if envelope.Revision != 5 {
			t.Fatalf("post-clearing player revision=%d", envelope.Revision)
		}
		postClearingKinds[envelope.Kind] = true
	}
	if !postClearingKinds["event"] {
		t.Fatalf("first Compact accrual event missing: %v", postClearingKinds)
	}
	state, err := composition.Accounts.ActiveCompanyState(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	var persisted struct {
		GuildID  *string `json:"guild_boundary_guild_id"`
		Boundary int64   `json:"guild_boundary_seq"`
	}
	if err := json.Unmarshal(state.State, &persisted); err != nil || persisted.GuildID == nil || *persisted.GuildID != guildReceipt.GuildID || persisted.Boundary != 3 {
		t.Fatalf("settlement watermark=%+v state=%s err=%v", persisted, state.State, err)
	}
	latestThird, err := testStore.LoadLatest(ctx, thirdCompany.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	latestThird.State.StockUnits = 5
	if _, err := testStore.Write(ctx, latestThird.Revision.StreamID, latestThird.Revision.Number, latestThird.Revision.ConstantsHash,
		latestThird.State, save.WriteContext{Cause: "gameserver.rejoin-reservation.integration"}); err != nil {
		t.Fatal(err)
	}
	var abandonedMembershipID string
	if err := db.QueryRowContext(ctx, `SELECT membership_id FROM guild_members WHERE guild_id=$1 AND account_id=$2 AND left_at IS NULL`, guildReceipt.GuildID, thirdCreated.AccountID).Scan(&abandonedMembershipID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO guild_clearing_results(guild_id,boundary_seq,account_id,debit_units,credit_units,allocations,committed_at,snapshot_hash,founder_id,company_stream_id,run_seq,membership_id)
		VALUES($1,4,$2,1,0,'[]'::jsonb,$3,$4,$5,$6,1,$7)`, guildReceipt.GuildID, thirdCreated.AccountID, clock.Time(),
		"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", thirdFounder.ID, thirdCompany.StreamID, abandonedMembershipID); err != nil {
		t.Fatal(err)
	}
	// Commit, leave, and rejoin deliberately share one canonical millisecond.
	// Membership identity, not timestamp ordering, must release the abandoned
	// period's reservation.
	thirdLeaveResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/guild/intents", thirdTokens.AccessToken,
		`{"intent_id":"018f0000-0000-7000-8000-000000000405","kind":"leave_guild","expected_revision":2}`)
	if thirdLeaveResponse.StatusCode != http.StatusOK {
		t.Fatalf("third member leave status=%d body=%s", thirdLeaveResponse.StatusCode, responseBody(thirdLeaveResponse))
	}
	_ = responseBody(thirdLeaveResponse)
	thirdRejoinResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/guild/intents", thirdTokens.AccessToken,
		fmt.Sprintf(`{"intent_id":"018f0000-0000-7000-8000-000000000406","kind":"join_guild","expected_revision":3,"guild_id":%q}`, guildReceipt.GuildID))
	if thirdRejoinResponse.StatusCode != http.StatusOK {
		t.Fatalf("third member rejoin status=%d body=%s", thirdRejoinResponse.StatusCode, responseBody(thirdRejoinResponse))
	}
	_ = responseBody(thirdRejoinResponse)
	rejoinedMembers, _, err := composition.Clearing.members(ctx, guildReceipt.GuildID)
	if err != nil {
		t.Fatal(err)
	}
	released := false
	for _, member := range rejoinedMembers {
		if member.Stock.AccountID == thirdCreated.AccountID {
			released = member.Stock.AvailableUnits == 5
		}
	}
	if !released {
		t.Fatalf("rejoined producer retained pre-join reservations: %+v", rejoinedMembers)
	}
	if count, err := composition.Clearing.Tick(ctx); err != nil || count != 2 {
		t.Fatalf("post-rejoin clearing count=%d err=%v", count, err)
	}
	clock.Set(clock.Time().Add(time.Second))
	newFounderResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/founder", tokens.AccessToken, `{}`)
	if newFounderResponse.StatusCode != http.StatusCreated {
		t.Fatalf("New Founder status=%d body=%s", newFounderResponse.StatusCode, responseBody(newFounderResponse))
	}
	_ = responseBody(newFounderResponse)
	newSessionResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session", "",
		fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode))
	var newTokens account.TokenPair
	decodeCompositionResponse(t, newSessionResponse, &newTokens)
	newFounder, err := composition.Accounts.ActiveFounder(ctx, created.AccountID)
	if err != nil || newFounder.ID == founder.ID {
		t.Fatalf("new founder=%+v err=%v", newFounder, err)
	}
	newRunIntent := `{"intent_id":"01985555-1114-7111-8111-111111111114","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`
	newRunResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", newTokens.AccessToken, newRunIntent)
	if newRunResponse.StatusCode != http.StatusOK {
		t.Fatalf("new-Founder intent replayed prior debit: status=%d body=%s", newRunResponse.StatusCode, responseBody(newRunResponse))
	}
	_ = responseBody(newRunResponse)
	newState, err := composition.Accounts.ActiveCompanyState(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(newState.State, &persisted); err != nil || persisted.GuildID == nil || *persisted.GuildID != guildReceipt.GuildID || persisted.Boundary != 5 {
		t.Fatalf("new-Founder settlement baseline=%+v state=%s err=%v", persisted, newState.State, err)
	}

	clock.Set(clock.Time().Add(31 * 24 * time.Hour))
	collected, err := composition.Accounts.PruneExpiredSessions(ctx, clock.Time(), 1_000)
	if err != nil || collected.RefreshTokens != 4 || collected.AccessTokens != 4 || collected.Families != 4 {
		t.Fatalf("session GC=%+v err=%v", collected, err)
	}

	_ = player.connection.Close(websocket.StatusNormalClosure, "test complete")
	_ = world.connection.Close(websocket.StatusNormalClosure, "test complete")
	drainContext, cancelDrain := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelDrain()
	if err := composition.Server.Drain(drainContext, clock.Time()); err != nil {
		t.Fatal(err)
	}
	waitHTTPStatus(t, httpServer.Client(), httpServer.URL+"/readyz", http.StatusServiceUnavailable)
}

func TestComposedGameserverStartupPrimesAttachedClearingAndSessionGCIntegration(t *testing.T) {
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
			t.Errorf("clean startup-prime database: %v", err)
		}
	})

	now := time.Date(2026, 8, 2, 13, 0, 0, 0, time.UTC)
	clock := &mutableClock{now: now}
	composition, err := Compose(ctx, CompositionConfig{
		DB: db, RepositoryRoot: filepathRoot(t), ServerID: "018f0000-0000-4000-8000-000000000303",
		ActivityBracket: "activity.standard", Clock: clock.Time,
		SigningKeys:   account.SigningKeys{CurrentID: "composition-prime", Current: bytes.Repeat([]byte{0x63}, 32)},
		BootstrapKeys: account.BootstrapReceiptKeys{CurrentID: "bootstrap-prime", Current: bytes.Repeat([]byte{0x64}, 32)},
	})
	if err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(composition.Server.Handler())
	defer httpServer.Close()
	bootstrapResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/bootstrap", "",
		fmt.Sprintf(`{"idempotency_key":%q}`, strings.Repeat("cd", 32)))
	if bootstrapResponse.StatusCode != http.StatusCreated {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapResponse.StatusCode, responseBody(bootstrapResponse))
	}
	_ = responseBody(bootstrapResponse)
	var liveBootstrapReceipts int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM bootstrap_receipts WHERE tombstoned_at IS NULL`).Scan(&liveBootstrapReceipts); err != nil || liveBootstrapReceipts != 1 {
		t.Fatalf("live bootstrap receipts=%d err=%v", liveBootstrapReceipts, err)
	}
	createdResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/account", "", `{}`)
	var created account.CreatedAccount
	decodeCompositionResponse(t, createdResponse, &created)
	sessionResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session", "",
		fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode))
	var tokens account.TokenPair
	decodeCompositionResponse(t, sessionResponse, &tokens)
	guildResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/guild/intents", tokens.AccessToken,
		`{"intent_id":"018f0000-0000-7000-8000-000000000403","kind":"create_guild","expected_revision":1,"name":"Prime Systems","join_policy":"open"}`)
	if guildResponse.StatusCode != http.StatusOK {
		t.Fatalf("guild status=%d body=%s", guildResponse.StatusCode, responseBody(guildResponse))
	}
	_ = responseBody(guildResponse)
	var beforeClearings int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_clearing_results`).Scan(&beforeClearings); err != nil || beforeClearings != 0 {
		t.Fatalf("clearings before start=%d err=%v", beforeClearings, err)
	}

	clock.Set(now.Add(31 * 24 * time.Hour))
	serverContext, cancelServer := context.WithCancel(ctx)
	defer cancelServer()
	if err := composition.Server.Start(serverContext); err != nil {
		t.Fatal(err)
	}
	var clearings, sessions, accessTokens, families, tombstonedBootstrapReceipts int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_clearing_results`).Scan(&clearings); err != nil || clearings != 1 {
		t.Fatalf("attached clearing prime count=%d err=%v", clearings, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM sessions),(SELECT count(*) FROM access_tokens),(SELECT count(*) FROM session_families)`).Scan(&sessions, &accessTokens, &families); err != nil {
		t.Fatal(err)
	}
	if sessions != 0 || accessTokens != 0 || families != 0 {
		t.Fatalf("attached session GC left sessions=%d access=%d families=%d", sessions, accessTokens, families)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM bootstrap_receipts WHERE tombstoned_at IS NOT NULL AND key_id IS NULL AND nonce IS NULL AND ciphertext IS NULL`).Scan(&tombstonedBootstrapReceipts); err != nil || tombstonedBootstrapReceipts != 1 {
		t.Fatalf("attached bootstrap receipt GC tombstones=%d err=%v", tombstonedBootstrapReceipts, err)
	}
	drainContext, cancelDrain := context.WithTimeout(context.Background(), time.Second)
	defer cancelDrain()
	if err := composition.Server.Drain(drainContext, clock.Time()); err != nil {
		t.Fatal(err)
	}
}

func TestComposedAccountFamilyRevocationRevalidatesSocketsIntegration(t *testing.T) {
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
	const cleanDatabase = `TRUNCATE bootstrap_receipts,accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`
	if _, err := db.ExecContext(ctx, cleanDatabase); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), cleanDatabase); err != nil {
			t.Errorf("clean socket-revocation database: %v", err)
		}
	})

	composition, err := Compose(ctx, CompositionConfig{
		DB: db, RepositoryRoot: filepathRoot(t), ServerID: "018f0000-0000-4000-8000-000000000304",
		ActivityBracket: "activity.standard", Clock: time.Now,
		SigningKeys:   account.SigningKeys{CurrentID: "composition-revocation", Current: bytes.Repeat([]byte{0x66}, 32)},
		BootstrapKeys: account.BootstrapReceiptKeys{CurrentID: "bootstrap-revocation", Current: bytes.Repeat([]byte{0x67}, 32)},
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
		t.Fatalf("account status=%d body=%s", createdResponse.StatusCode, responseBody(createdResponse))
	}
	var created account.CreatedAccount
	decodeCompositionResponse(t, createdResponse, &created)
	createSession := func() account.TokenPair {
		response := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session", "",
			fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode))
		if response.StatusCode != http.StatusOK {
			t.Fatalf("session status=%d body=%s", response.StatusCode, responseBody(response))
		}
		var pair account.TokenPair
		decodeCompositionResponse(t, response, &pair)
		return pair
	}
	target := createSession()
	control := createSession()
	founder, err := composition.Accounts.ActiveFounder(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	websocketURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/connection/websocket"
	live := dialCompositionSocket(t, websocketURL, httpServer.Client(), target.AccessToken)
	defer live.connection.CloseNow()
	writeWS(t, live, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "player:" + founder.ID}})
	if reply := readWSID(t, live, 2); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("pre-revocation subscribe=%+v", reply)
	}

	rotateResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session/refresh", "",
		fmt.Sprintf(`{"refresh_token":%q}`, target.RefreshToken))
	if rotateResponse.StatusCode != http.StatusOK {
		t.Fatalf("rotate status=%d body=%s", rotateResponse.StatusCode, responseBody(rotateResponse))
	}
	var rotated account.TokenPair
	decodeCompositionResponse(t, rotateResponse, &rotated)
	reuseResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session/refresh", "",
		fmt.Sprintf(`{"refresh_token":%q}`, target.RefreshToken))
	if reuseResponse.StatusCode != http.StatusUnauthorized || !strings.Contains(responseBody(reuseResponse), `"category":"refresh_reused"`) {
		t.Fatalf("refresh-family reuse status=%d", reuseResponse.StatusCode)
	}

	revokedBeforeConnect, _, err := websocket.Dial(context.Background(), websocketURL, &websocket.DialOptions{
		HTTPClient: httpServer.Client(), HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeSocketCommand(t, revokedBeforeConnect, map[string]any{"id": 1, "connect": map[string]any{"token": rotated.AccessToken}})
	if status := waitSocketCloseStatus(t, revokedBeforeConnect, 3*time.Second); status != websocket.StatusCode(transport.CloseAuthExpired) {
		t.Fatalf("revoked-before-connect close=%d want=%d", status, transport.CloseAuthExpired)
	}

	// Centrifuge invokes Node.OnAlive on its 25-second presence cadence. That
	// callback re-authenticates the stored token against Postgres, so the
	// already-connected socket must be closed after family revocation.
	if status := waitSocketCloseStatus(t, live.connection, 30*time.Second); status != websocket.StatusCode(transport.CloseAuthExpired) {
		t.Fatalf("live revoked socket close=%d want=%d", status, transport.CloseAuthExpired)
	}
	data, _ := json.Marshal(map[string]any{"id": 3, "subscribe": map[string]any{"channel": "world"}})
	writeContext, cancelWrite := context.WithTimeout(context.Background(), time.Second)
	defer cancelWrite()
	if err := live.connection.Write(writeContext, websocket.MessageText, data); err == nil {
		t.Fatal("post-revocation subscription write survived the alive denial")
	}

	unrevoked := dialCompositionSocket(t, websocketURL, httpServer.Client(), control.AccessToken)
	defer unrevoked.connection.CloseNow()
	writeWS(t, unrevoked, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "player:" + founder.ID}})
	if reply := readWSID(t, unrevoked, 2); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("unrevoked family control subscribe=%+v", reply)
	}

	drainContext, cancelDrain := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelDrain()
	if err := composition.Server.Drain(drainContext, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func TestComposedGameserverExitVerificationAndBoardIntegration(t *testing.T) {
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
			t.Errorf("clean verification database: %v", err)
		}
	})

	now := time.Date(2026, 8, 2, 14, 0, 0, 0, time.UTC)
	composition, err := Compose(ctx, CompositionConfig{
		DB: db, RepositoryRoot: filepathRoot(t), ServerID: "018f0000-0000-4000-8000-000000000302",
		ActivityBracket: "activity.standard", Clock: func() time.Time { return now },
		SigningKeys:   account.SigningKeys{CurrentID: "composition-verification", Current: bytes.Repeat([]byte{0x24}, 32)},
		BootstrapKeys: account.BootstrapReceiptKeys{CurrentID: "bootstrap-verification", Current: bytes.Repeat([]byte{0x25}, 32)},
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

	founderID := "018f0000-0000-4000-8000-000000000502"
	initialFounder, err := composition.Accounts.ActiveFounder(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE account_founders SET archived_at=$2 WHERE account_id=$1 AND archived_at IS NULL`, created.AccountID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE save_streams SET archived_at=$2 WHERE owner_kind='founder' AND owner_id=$1 AND archived_at IS NULL`, initialFounder.ID, now); err != nil {
		t.Fatal(err)
	}
	store, err := save.NewStore(db, composition.Catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}
	founderState, companyState, frozen := progressedCompositionStates(t, composition.Catalogs, composition.CurrentHash, now)
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		composition.CurrentHash, founderState, save.WriteContext{Cause: "gameserver.composition.integration"})
	if err != nil {
		t.Fatal(err)
	}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		composition.CurrentHash, companyState, save.WriteContext{Cause: "gameserver.composition.integration"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id,created_at) VALUES($1,$2,$3)`, created.AccountID, founderID, now); err != nil {
		t.Fatal(err)
	}
	pinTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var genesis []byte
	if err := pinTx.QueryRowContext(ctx, `SELECT state::text FROM save_revisions WHERE stream_id=$1 AND revision=$2`, companyRevision.StreamID, companyRevision.Number).Scan(&genesis); err != nil {
		_ = pinTx.Rollback()
		t.Fatal(err)
	}
	if _, err := save.PinRunWithGenesisTx(ctx, pinTx, companyRevision.StreamID, founderID, 1, companyRevision.ConstantsHash, companyRevision.Version, genesis); err != nil {
		_ = pinTx.Rollback()
		t.Fatal(err)
	}
	if err := save.InsertRunFrozenContributionsTx(ctx, pinTx, companyRevision.StreamID, 1, frozen); err != nil {
		_ = pinTx.Rollback()
		t.Fatal(err)
	}
	if err := pinTx.Commit(); err != nil {
		t.Fatal(err)
	}

	manual := `{"intent_id":"01985555-5001-7000-8000-000000000501","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`
	manualResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken, manual)
	if manualResponse.StatusCode != http.StatusOK {
		t.Fatalf("manual status=%d body=%s", manualResponse.StatusCode, responseBody(manualResponse))
	}
	_ = responseBody(manualResponse)
	loadedCompany, err := store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	loadedFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	if companyVersion, founderVersion := save.VersionForState(loadedCompany.State), save.VersionForState(loadedFounder.State); companyVersion != 18 || founderVersion != 21 {
		t.Fatalf("pre-Exit activation versions company=%d founder=%d", companyVersion, founderVersion)
	}
	exit := `{"intent_id":"01985555-5002-7000-8000-000000000502","kind":"wind_down","expected_revision":2,"expected_founder_revision":1}`
	exitResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken, exit)
	if exitResponse.StatusCode != http.StatusOK {
		t.Fatalf("exit status=%d body=%s", exitResponse.StatusCode, responseBody(exitResponse))
	}
	_ = responseBody(exitResponse)
	var pinnedGenesis []byte
	var pinnedVersion int
	if err := db.QueryRowContext(ctx, `SELECT state,version FROM run_genesis WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&pinnedGenesis, &pinnedVersion); err != nil {
		t.Fatal(err)
	}
	pinnedCatalog, ok := composition.Catalogs.Resolve(composition.CurrentHash)
	if !ok {
		t.Fatal("current replay economy catalog unavailable")
	}
	if _, err := save.RestoreState(pinnedGenesis, pinnedVersion, pinnedCatalog, economy.ScopeCompany, time.Time{}); err != nil {
		t.Fatalf("pinned genesis version=%d does not restore: %v", pinnedVersion, err)
	}
	var pinnedArtifactCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM catalog_artifacts WHERE constants_hash=$1`, composition.CurrentHash).Scan(&pinnedArtifactCount); err != nil || pinnedArtifactCount != 19 {
		t.Fatalf("pinned artifact count=%d err=%v", pinnedArtifactCount, err)
	}
	var pinnedReplayBundle production.CatalogBundle
	if replaySet, err := replaycatalog.LoadDatabase(ctx, db); err != nil {
		t.Fatalf("pinned replay catalog load: %v", err)
	} else if pinnedReplayBundle, ok = replaySet.ResolveReplayCatalogs(composition.CurrentHash); !ok {
		t.Fatal("pinned replay catalog missing current hash")
	}
	if verdict := production.VerifyReplayRun(pinnedGenesis, pinnedVersion, pinnedReplayBundle, nil, composition.CurrentHash, false); verdict != production.ReplayLogGap {
		t.Fatalf("pinned replay preflight verdict=%s version=%d", verdict, pinnedVersion)
	}
	var runHash string
	if err := db.QueryRowContext(ctx, `SELECT constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&runHash); err != nil || runHash != composition.CurrentHash {
		t.Fatalf("run pin hash=%s want=%s err=%v", runHash, composition.CurrentHash, err)
	}
	if verdict, err := composition.Verification.VerifyStoredRun(ctx, companyRevision.StreamID, 1); err != nil || verdict != production.ReplayVerified {
		t.Fatalf("stored run verdict=%s err=%v", verdict, err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var status string
		var boardRows int
		statusErr := db.QueryRowContext(ctx, `SELECT status FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&status)
		boardErr := db.QueryRowContext(ctx, `SELECT count(*) FROM verified_runs WHERE run_id=$1 AND category_id='any_percent'`, companyRevision.StreamID+":1").Scan(&boardRows)
		if statusErr == nil && boardErr == nil && status == "verified" && boardRows == 1 {
			activeRun, err := composition.Accounts.ActiveCompanyState(ctx, created.AccountID)
			if err != nil {
				t.Fatal(err)
			}
			activeState, restoreErr := save.RestoreState(activeRun.State, activeRun.Version, pinnedReplayBundle.Economy, economy.ScopeCompany, now)
			if restoreErr != nil || activeRun.Version != 18 || activeRun.ConstantsHash != composition.CurrentHash || activeState.RunSeq != 2 ||
				activeState.OpportunitySpawnSeq != 0 || activeState.NextOpportunityAttendedMS <= 0 ||
				activeState.PendingOpportunity != nil || activeState.ActiveBuffs == nil {
				t.Fatalf("epoch-7 new-run activation=%+v state=%+v err=%v", activeRun, activeState, restoreErr)
			}
			var retainedExitWitness int
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_log log
				WHERE log.company_stream_id=$1 AND log.run_seq=1 AND EXISTS (
					SELECT 1 FROM founder_log founder
					WHERE founder.source_company_stream_id=log.company_stream_id
					  AND founder.source_run_seq=log.run_seq
					  AND founder.source_run_log_seq=log.seq
				)`, companyRevision.StreamID).Scan(&retainedExitWitness); err != nil || retainedExitWitness != 1 {
				t.Fatalf("retained Founder Exit witness=%d err=%v", retainedExitWitness, err)
			}
			proveFreshFounderReceivesOwnScriptedExit(t, ctx, composition, httpServer, store, created.AccountID,
				tokens.AccessToken, founderRevision.StreamID, founderID,
				companyRevision.StreamID, now)
			proveImportedFounderExitExcludedFromBoard(t, ctx, db, composition, httpServer, store, now)
			drainContext, cancelDrain := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancelDrain()
			if err := composition.Server.Drain(drainContext, now.Add(2*time.Second)); err != nil {
				t.Fatal(err)
			}
			loadedFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
			if err != nil || len(loadedFounder.State.ExitHistory) != 1 ||
				loadedFounder.State.MinigameSessionSeq != 0 || loadedFounder.State.FiscalCredit != pinnedReplayBundle.Fiscal.Credit.Hardcap ||
				loadedFounder.State.Soul != 80 {
				t.Fatalf("persisted founder exit history=%+v session_seq=%d fiscal_credit=%d soul=%d err=%v",
					loadedFounder.State.ExitHistory, loadedFounder.State.MinigameSessionSeq,
					loadedFounder.State.FiscalCredit, loadedFounder.State.Soul, err)
			}
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	var status, verdict string
	var boardRows int
	_ = db.QueryRowContext(ctx, `SELECT status,COALESCE(verdict,'') FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&status, &verdict)
	_ = db.QueryRowContext(ctx, `SELECT count(*) FROM verified_runs WHERE run_id=$1`, companyRevision.StreamID+":1").Scan(&boardRows)
	var backgroundErr error
	select {
	case backgroundErr = <-composition.Server.Failures():
	default:
	}
	t.Fatalf("terminal run did not reach board: status=%s verdict=%s rows=%d background=%v", status, verdict, boardRows, backgroundErr)
}

func proveFreshFounderReceivesOwnScriptedExit(t *testing.T, ctx context.Context, composition *Composition,
	httpServer *httptest.Server, store *save.Store, accountID, accessToken, oldFounderStreamID,
	oldFounderID, oldCompanyStreamID string, now time.Time) {
	t.Helper()
	newFounderResponse := compositionRequest(t, httpServer.Client(), http.MethodPost,
		httpServer.URL+"/api/v1/founder", accessToken, `{}`)
	if newFounderResponse.StatusCode != http.StatusCreated {
		t.Fatalf("fresh Founder status=%d body=%s", newFounderResponse.StatusCode, responseBody(newFounderResponse))
	}
	var fresh account.Founder
	decodeCompositionResponse(t, newFounderResponse, &fresh)
	if fresh.ID == oldFounderID || fresh.Imported {
		t.Fatalf("fresh Founder identity=%+v old=%s", fresh, oldFounderID)
	}
	oldFounder, err := store.LoadLatest(ctx, oldFounderStreamID)
	if err != nil || oldFounder.ArchivedAt == nil || len(oldFounder.State.ExitHistory) != 1 {
		t.Fatalf("old Founder after replacement=%+v err=%v", oldFounder, err)
	}
	oldCompany, err := store.LoadLatest(ctx, oldCompanyStreamID)
	if err != nil || oldCompany.ArchivedAt == nil || oldCompany.State.RunSeq != 2 {
		t.Fatalf("old Company after replacement=%+v err=%v", oldCompany, err)
	}
	activeFounder, err := composition.Accounts.ActiveFounder(ctx, accountID)
	if err != nil || activeFounder.ID != fresh.ID {
		t.Fatalf("active fresh Founder=%+v response=%+v err=%v", activeFounder, fresh, err)
	}
	activeCompany, err := composition.Accounts.ActiveCompanyState(ctx, accountID)
	if err != nil || activeCompany.Revision != 1 {
		t.Fatalf("fresh Company=%+v err=%v", activeCompany, err)
	}
	company, err := store.LoadLatest(ctx, activeCompany.StreamID)
	if err != nil || company.State.RunSeq != 1 {
		t.Fatalf("fresh Company state=%+v err=%v", company, err)
	}
	freshFounder, err := store.LoadSiblingLatest(ctx, activeCompany.StreamID, economy.ScopeFounder)
	if err != nil || len(freshFounder.State.ExitHistory) != 0 {
		t.Fatalf("fresh Founder inherited Exit history=%+v err=%v", freshFounder.State, err)
	}
	company.State.RunStartedAt = now.Add(-16 * time.Minute)
	company.State.EvaluatedThrough = now
	company.State.ManualTokenRefilledAt = now
	company.State.GatesCrossed["gate.t0_to_t1"] = true
	company.State.OpportunitySpawnSeq = 0
	company.State.NextOpportunityAttendedMS = int64((20 * time.Minute) / time.Millisecond)
	company.State.PendingOpportunity = nil
	company.State.ActiveBuffs = []save.ActiveBuff{}
	written, err := store.Write(ctx, company.Revision.StreamID, company.Revision.Number,
		company.Revision.ConstantsHash, company.State, save.WriteContext{Cause: "prestige.ac4.fresh_founder_due"})
	if err != nil || written.Number != 2 {
		t.Fatalf("fresh Founder due-state write=%+v err=%v", written, err)
	}
	command := []byte(`{"intent_id":"01985555-5005-7000-8000-000000000505","kind":"perform_manual_batch","expected_revision":2,"action_id":"manual.click","count":1,"window_ms":1}`)
	result, err := composition.Production.Handle(ctx, activeCompany.StreamID, production.ModeOnline, now, command)
	if err != nil || result.Replay {
		t.Fatalf("fresh Founder scripted command=%+v err=%v", result, err)
	}
	freshFounder, err = store.LoadSiblingLatest(ctx, activeCompany.StreamID, economy.ScopeFounder)
	if err != nil || len(freshFounder.State.ExitHistory) != 1 ||
		freshFounder.State.ExitHistory[0].ExitType != "scripted_first" {
		t.Fatalf("fresh Founder scripted history=%+v err=%v", freshFounder.State.ExitHistory, err)
	}
	company, err = store.LoadLatest(ctx, activeCompany.StreamID)
	if err != nil || company.State.RunSeq != 2 || company.Revision.Number != 4 {
		t.Fatalf("fresh Founder post-scripted Company=%+v err=%v", company, err)
	}
}

func proveImportedFounderExitExcludedFromBoard(t *testing.T, ctx context.Context, db *sql.DB, composition *Composition,
	httpServer *httptest.Server, store *save.Store, now time.Time) {
	t.Helper()
	createdResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/account", "", `{}`)
	if createdResponse.StatusCode != http.StatusCreated {
		t.Fatalf("import-control account status=%d body=%s", createdResponse.StatusCode, responseBody(createdResponse))
	}
	var created account.CreatedAccount
	decodeCompositionResponse(t, createdResponse, &created)
	sessionResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/session", "",
		fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode))
	if sessionResponse.StatusCode != http.StatusOK {
		t.Fatalf("import-control session status=%d body=%s", sessionResponse.StatusCode, responseBody(sessionResponse))
	}
	var tokens account.TokenPair
	decodeCompositionResponse(t, sessionResponse, &tokens)

	currentCatalog, ok := composition.Catalogs.Resolve(composition.CurrentHash)
	if !ok {
		t.Fatal("current import economy catalog unavailable")
	}
	localLedger, err := economy.NewLedger(currentCatalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	localCompany := &save.State{Ledger: localLedger, GeneratorCounts: map[string]int64{}, GeneratorProvisioned: map[string]int64{},
		ProvisionRemaindersPPM: map[string]int64{}, UpgradesOwned: map[string]bool{}, EvaluatedThrough: now, ManualTokenRefilledAt: now,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{},
		RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.New(8, 12),
		OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}, RunSeq: 1, RunStartedAt: now, Tier: 2,
		ManualTokenMilli: currentCatalog.ManualPolicy().BucketCapMilli}
	for _, generator := range currentCatalog.GeneratorClassesForScope(economy.ScopeCompany) {
		localCompany.GeneratorCounts[generator.ID] = 0
		localCompany.GeneratorProvisioned[generator.ID] = 0
		if generator.Provision != nil {
			localCompany.ProvisionRemaindersPPM[generator.Provision.GeneratorID] = 0
		}
	}
	localVersion := save.CurrentVersion
	localBytes, err := save.EncodeStateVersion(localCompany, localVersion)
	if err != nil {
		t.Fatal(err)
	}
	restoredLocal, err := save.RestoreState(localBytes, localVersion, currentCatalog, economy.ScopeCompany, now)
	if err != nil {
		t.Fatalf("restore local import fixture: %v", err)
	}
	normalizedBytes, err := save.EncodeState(restoredLocal)
	if err != nil {
		t.Fatalf("encode normalized import fixture: %v", err)
	}
	if _, err := save.RestoreState(normalizedBytes, save.CurrentVersion, currentCatalog, economy.ScopeCompany, time.Time{}); err != nil {
		t.Fatalf("restore normalized import fixture: %v", err)
	}
	importBody, err := json.Marshal(struct {
		Version       int             `json:"version"`
		ConstantsHash string          `json:"constants_hash"`
		State         json.RawMessage `json:"state"`
	}{Version: localVersion, ConstantsHash: composition.CurrentHash, State: localBytes})
	if err != nil {
		t.Fatal(err)
	}
	importResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/founder/import", tokens.AccessToken, string(importBody))
	if importResponse.StatusCode != http.StatusOK {
		t.Fatalf("actual import status=%d body=%s", importResponse.StatusCode, responseBody(importResponse))
	}
	var imported account.Founder
	decodeCompositionResponse(t, importResponse, &imported)
	if !imported.Imported {
		t.Fatalf("actual import response=%+v", imported)
	}
	company, err := composition.Accounts.ActiveCompanyState(ctx, created.AccountID)
	if err != nil || company.Revision != 1 {
		t.Fatalf("imported active company=%+v err=%v", company, err)
	}

	manual := `{"intent_id":"01985555-5003-7000-8000-000000000503","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`
	manualResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken, manual)
	if manualResponse.StatusCode != http.StatusOK {
		t.Fatalf("imported manual status=%d body=%s", manualResponse.StatusCode, responseBody(manualResponse))
	}
	_ = responseBody(manualResponse)
	exit := `{"intent_id":"01985555-5004-7000-8000-000000000504","kind":"wind_down","expected_revision":2,"expected_founder_revision":1}`
	exitResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken, exit)
	if exitResponse.StatusCode != http.StatusOK {
		t.Fatalf("imported Exit status=%d body=%s", exitResponse.StatusCode, responseBody(exitResponse))
	}
	_ = responseBody(exitResponse)
	if verdict, err := composition.Verification.VerifyStoredRun(ctx, company.StreamID, 1); err != nil || verdict != production.ReplayVerified {
		t.Fatalf("imported stored run verdict=%s err=%v", verdict, err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var status string
		var boardRows, projectionClaims int
		statusErr := db.QueryRowContext(ctx, `SELECT status FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, company.StreamID).Scan(&status)
		boardErr := db.QueryRowContext(ctx, `SELECT count(*) FROM verified_runs WHERE run_id=$1`, company.StreamID+":1").Scan(&boardRows)
		claimErr := db.QueryRowContext(ctx, `SELECT count(*) FROM verification_projection_events projection
			JOIN events terminal ON terminal.event_id=projection.event_id
			WHERE terminal.stream_id=$1 AND terminal.kind='run_ended'`, company.StreamID).Scan(&projectionClaims)
		if statusErr == nil && boardErr == nil && claimErr == nil && status == "verified" && boardRows == 0 && projectionClaims == 1 {
			loaded, loadErr := store.LoadLatest(ctx, company.StreamID)
			if loadErr != nil || loaded.Revision.Number < 3 || loaded.State.RunSeq != 2 {
				t.Fatalf("imported post-Exit company=%+v err=%v", loaded, loadErr)
			}
			active, activeErr := composition.Accounts.ActiveFounder(ctx, created.AccountID)
			if activeErr != nil || active.ID != imported.ID || !active.Imported {
				t.Fatalf("imported identity after projection=%+v err=%v", active, activeErr)
			}
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	var status string
	var boardRows, projectionClaims int
	_ = db.QueryRowContext(ctx, `SELECT status FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, company.StreamID).Scan(&status)
	_ = db.QueryRowContext(ctx, `SELECT count(*) FROM verified_runs WHERE run_id=$1`, company.StreamID+":1").Scan(&boardRows)
	_ = db.QueryRowContext(ctx, `SELECT count(*) FROM verification_projection_events projection
		JOIN events terminal ON terminal.event_id=projection.event_id WHERE terminal.stream_id=$1`, company.StreamID).Scan(&projectionClaims)
	t.Fatalf("imported run projection status=%s rows=%d claims=%d", status, boardRows, projectionClaims)
}

func progressedCompositionStates(t *testing.T, catalogs *runtimeCatalogs, constantsHash string, now time.Time) (*save.State, *save.State, []save.FrozenContribution) {
	t.Helper()
	catalog, ok := catalogs.Resolve(constantsHash)
	if !ok {
		t.Fatal("progressed-state economy catalog unavailable")
	}
	founderLedger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	companyLedger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	base := func(ledger *economy.Ledger) *save.State {
		return &save.State{Ledger: ledger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: now, ManualTokenRefilledAt: now,
			GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{},
			MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{},
			LifetimeValue: decimal.Zero, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	}
	founder := base(founderLedger)
	company := base(companyLedger)
	company.GeneratorCounts = map[string]int64{}
	company.GeneratorProvisioned = map[string]int64{}
	company.ProvisionRemaindersPPM = map[string]int64{}
	company.UpgradesOwned = map[string]bool{}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		company.GeneratorCounts[generator.ID] = 0
		company.GeneratorProvisioned[generator.ID] = 0
		if generator.Provision != nil {
			company.ProvisionRemaindersPPM[generator.Provision.GeneratorID] = 0
		}
	}
	company.ManualTokenMilli = catalog.ManualPolicy().BucketCapMilli
	company.RunSeq = 1
	// Active-play attended time starts at the run boundary. This fixture drives
	// one immediate intent, so do not manufacture ten minutes of unprocessed
	// attended scheduling before its first command.
	company.RunStartedAt = now
	frozen, err := (production.FounderInitializer{Catalogs: catalogs}).InitializeNewFounder(constantsHash, "018f0000-0000-4000-8000-000000000502", now, founder, company)
	if err != nil {
		t.Fatal(err)
	}
	founder.MinigameSessionSeq = 9
	founder.FiscalCredit = 2
	founder.Soul = 80
	company.Tier = 2
	company.LifetimeValue = decimal.New(8, 12)
	return founder, company, frozen
}

func filepathRoot(t *testing.T) string {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(root, "..", ".."))
}

func compositionRequest(t *testing.T, client *http.Client, method, url, token, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func decodeCompositionResponse(t *testing.T, response *http.Response, target any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}

func responseBody(response *http.Response) string {
	defer response.Body.Close()
	data, _ := io.ReadAll(response.Body)
	return string(data)
}

func waitHTTPStatus(t *testing.T, client *http.Client, url string, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(url)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == want {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%s did not reach status %d", url, want)
}

func dialCompositionSocket(t *testing.T, endpoint string, client *http.Client, token string) *compositionSocket {
	t.Helper()
	connection, _, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPClient: client, HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}}})
	if err != nil {
		t.Fatal(err)
	}
	socket := &compositionSocket{connection: connection}
	writeWS(t, socket, map[string]any{"id": 1, "connect": map[string]any{"token": token}})
	if reply := readWSID(t, socket, 1); reply.Error != nil || reply.Connect == nil {
		_ = connection.CloseNow()
		t.Fatalf("connect reply=%+v", reply)
	}
	return socket
}

func writeWS(t *testing.T, socket *compositionSocket, value any) {
	t.Helper()
	writeSocketCommand(t, socket.connection, value)
}

func writeSocketCommand(t *testing.T, connection *websocket.Conn, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := connection.Write(ctx, websocket.MessageText, data); err != nil {
		t.Fatal(err)
	}
}

func waitSocketCloseStatus(t *testing.T, connection *websocket.Conn, timeout time.Duration) websocket.StatusCode {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		if _, _, err := connection.Read(ctx); err != nil {
			return websocket.CloseStatus(err)
		}
	}
}

func readWS(t *testing.T, socket *compositionSocket) wsReply {
	t.Helper()
	if len(socket.pending) > 0 {
		reply := socket.pending[0]
		socket.pending = socket.pending[1:]
		return reply
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, data, err := socket.connection.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var reply wsReply
		if err := decoder.Decode(&reply); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("decode websocket reply %s: %v", data, err)
		}
		socket.pending = append(socket.pending, reply)
	}
	if len(socket.pending) == 0 {
		t.Fatalf("empty websocket frame: %s", data)
	}
	reply := socket.pending[0]
	socket.pending = socket.pending[1:]
	return reply
}

func readWSID(t *testing.T, socket *compositionSocket, id uint32) wsReply {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		reply := readWS(t, socket)
		if reply.ID == id {
			return reply
		}
	}
	t.Fatalf("no websocket reply for command %d", id)
	return wsReply{}
}

func readEnvelope(t *testing.T, socket *compositionSocket, channel string) transport.Envelope {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		reply := readWS(t, socket)
		if reply.Push == nil || reply.Push.Publication == nil || reply.Push.Channel != channel {
			continue
		}
		var envelope transport.Envelope
		if err := json.Unmarshal(reply.Push.Publication.Data, &envelope); err != nil {
			t.Fatalf("decode envelope %s: %v", reply.Push.Publication.Data, err)
		}
		return envelope
	}
	t.Fatalf("no publication for %s", channel)
	return transport.Envelope{}
}
