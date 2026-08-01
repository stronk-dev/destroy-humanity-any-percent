package gameserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
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
	const cleanDatabase = `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`
	if _, err := db.ExecContext(ctx, cleanDatabase); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), cleanDatabase); err != nil {
			t.Errorf("clean composition database: %v", err)
		}
	})

	clock := &mutableClock{now: time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)}
	composition, err := Compose(ctx, CompositionConfig{
		DB:              db,
		RepositoryRoot:  filepathRoot(t),
		ServerID:        "018f0000-0000-4000-8000-000000000301",
		ActivityBracket: "activity.standard",
		SigningKeys: account.SigningKeys{
			CurrentID: "composition-integration",
			Current:   bytes.Repeat([]byte{0x42}, 32),
		},
		Clock: clock.Time,
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
	founder, err := composition.Accounts.ActiveFounder(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
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
	companyAfterPlay, err := composition.Accounts.ActiveCompanyState(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	eventPayload, _ := json.Marshal(map[string]any{
		"founder_id":         founder.ID,
		"run_id":             map[string]any{"company_stream_id": companyAfterPlay.StreamID, "run_seq": 1},
		"progress_delta_ppm": 0,
		"xp_delta":           0,
	})
	if _, err := db.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,constants_hash,payload) VALUES($1,2,1,'guild_activity_evaluated',$2,$3)`,
		companyAfterPlay.StreamID, companyAfterPlay.ConstantsHash, eventPayload); err != nil {
		t.Fatal(err)
	}
	eventEnvelope := readEnvelope(t, player, "player:"+founder.ID)
	if eventEnvelope.Revision != 2 || eventEnvelope.Kind != "event" {
		t.Fatalf("player event envelope=%+v", eventEnvelope)
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
	writeWS(t, player, map[string]any{"id": 3, "subscribe": map[string]any{"channel": "guild:" + guildReceipt.GuildID}})
	if reply := readWSID(t, player, 3); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("guild resolver did not authorize member: %+v", reply)
	}
	writeWS(t, player, map[string]any{"id": 4, "subscribe": map[string]any{"channel": "match:018f0000-0000-7000-8000-000000000499"}})
	if reply := readWSID(t, player, 4); reply.Error == nil || reply.Error.Code != 103 {
		t.Fatalf("unowned match resolver did not fail closed: %+v", reply)
	}

	before, err := composition.Guilds.PendingSettlements(ctx, founder.ID, "", 0)
	if err != nil || before.GuildID != guildReceipt.GuildID || before.BaseSeq != 0 || len(before.Settlements) != 0 {
		t.Fatalf("pre-clearing batch=%+v err=%v", before, err)
	}
	if count, err := composition.Clearing.Tick(ctx); err != nil || count != 1 {
		t.Fatalf("clearing count=%d err=%v", count, err)
	}
	after, err := composition.Guilds.PendingSettlements(ctx, founder.ID, "", 0)
	if err != nil || after.GuildID != guildReceipt.GuildID || len(after.Settlements) != 1 || after.Settlements[0].BoundarySeq != 1 {
		t.Fatalf("post-clearing batch=%+v err=%v", after, err)
	}
	secondIntent := `{"intent_id":"01985555-1112-7111-8111-111111111112","kind":"perform_manual_batch","expected_revision":2,"action_id":"manual.click","count":1,"window_ms":1}`
	secondResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/intents", tokens.AccessToken, secondIntent)
	if secondResponse.StatusCode != http.StatusOK {
		t.Fatalf("post-clearing intent status=%d body=%s", secondResponse.StatusCode, responseBody(secondResponse))
	}
	_ = responseBody(secondResponse)
	secondEnvelope := readEnvelope(t, player, "player:"+founder.ID)
	if secondEnvelope.Revision != 3 || secondEnvelope.Kind != "receipt" {
		t.Fatalf("post-clearing player envelope=%+v", secondEnvelope)
	}
	state, err := composition.Accounts.ActiveCompanyState(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	var persisted struct {
		GuildID  *string `json:"guild_boundary_guild_id"`
		Boundary int64   `json:"guild_boundary_seq"`
	}
	if err := json.Unmarshal(state.State, &persisted); err != nil || persisted.GuildID == nil || *persisted.GuildID != guildReceipt.GuildID || persisted.Boundary != 1 {
		t.Fatalf("settlement watermark=%+v state=%s err=%v", persisted, state.State, err)
	}

	clock.Set(clock.Time().Add(31 * 24 * time.Hour))
	collected, err := composition.Accounts.PruneExpiredSessions(ctx, clock.Time(), 1_000)
	if err != nil || collected.RefreshTokens != 1 || collected.AccessTokens != 1 || collected.Families != 1 {
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
		SigningKeys: account.SigningKeys{CurrentID: "composition-verification", Current: bytes.Repeat([]byte{0x24}, 32)},
	})
	if err != nil {
		t.Fatal(err)
	}
	serverContext, cancelServer := context.WithCancel(ctx)
	defer cancelServer()
	if err := composition.Server.Start(serverContext); err != nil {
		t.Fatal(err)
	}

	accountID := "018f0000-0000-4000-8000-000000000501"
	founderID := "018f0000-0000-4000-8000-000000000502"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,'composition fixture',$2)`, accountID, now); err != nil {
		t.Fatal(err)
	}
	catalog, ok := composition.Catalogs.Resolve(composition.CurrentHash)
	if !ok {
		t.Fatal("current economy catalog unavailable")
	}
	store, err := save.NewStore(db, composition.Catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}
	founderState, companyState := progressedCompositionStates(t, catalog, now)
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
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id,created_at) VALUES($1,$2,$3)`, accountID, founderID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, founderID, 1, companyRevision.ConstantsHash); err != nil {
		t.Fatal(err)
	}

	manual := []byte(`{"intent_id":"01985555-5001-7000-8000-000000000501","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`)
	if result, err := composition.Production.Handle(ctx, companyRevision.StreamID, production.ModeOnline, now, manual); err != nil || result.Replay {
		t.Fatalf("manual result=%+v err=%v", result, err)
	}
	exit := []byte(`{"intent_id":"01985555-5002-7000-8000-000000000502","kind":"wind_down","expected_revision":2,"expected_founder_revision":1}`)
	if result, err := composition.Production.Handle(ctx, companyRevision.StreamID, production.ModeOnline, now.Add(time.Second), exit); err != nil || result.Replay {
		t.Fatalf("exit result=%+v err=%v", result, err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var status string
		var boardRows int
		statusErr := db.QueryRowContext(ctx, `SELECT status FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&status)
		boardErr := db.QueryRowContext(ctx, `SELECT count(*) FROM verified_runs WHERE run_id=$1 AND category_id='any_percent'`, companyRevision.StreamID+":1").Scan(&boardRows)
		if statusErr == nil && boardErr == nil && status == "verified" && boardRows == 1 {
			drainContext, cancelDrain := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancelDrain()
			if err := composition.Server.Drain(drainContext, now.Add(2*time.Second)); err != nil {
				t.Fatal(err)
			}
			loadedFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
			if err != nil || len(loadedFounder.State.ExitHistory) != 1 {
				t.Fatalf("founder exit history=%+v err=%v", loadedFounder.State.ExitHistory, err)
			}
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	var status, verdict string
	var boardRows int
	_ = db.QueryRowContext(ctx, `SELECT status,COALESCE(verdict,'') FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&status, &verdict)
	_ = db.QueryRowContext(ctx, `SELECT count(*) FROM verified_runs WHERE run_id=$1`, companyRevision.StreamID+":1").Scan(&boardRows)
	t.Fatalf("terminal run did not reach board: status=%s verdict=%s rows=%d", status, verdict, boardRows)
}

func progressedCompositionStates(t *testing.T, catalog *economy.Catalog, now time.Time) (*save.State, *save.State) {
	t.Helper()
	founderLedger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	companyLedger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": "0"})
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
	company.GeneratorCounts = map[string]int64{"generator.beige_tower": 0}
	company.ManualTokenMilli = catalog.ManualPolicy().BucketCapMilli
	company.RunSeq = 1
	company.Tier = 2
	company.LifetimeValue = decimal.New(8, 12)
	company.RunStartedAt = now.Add(-10 * time.Minute)
	return founder, company
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
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := socket.connection.Write(ctx, websocket.MessageText, data); err != nil {
		t.Fatal(err)
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
