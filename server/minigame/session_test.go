package minigame

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"

	"cloud-clicker/server/save"
)

const (
	testAccountID = "018f0000-0000-4000-8000-000000000101"
	testFounderID = "018f0000-0000-4000-8000-000000000102"
	testStreamID  = "018f0000-0000-4000-8000-000000000103"
	testSessionID = "018f0000-0000-7000-8000-000000000104"
	testHash      = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestCreateSessionValidation(t *testing.T) {
	valid := testCreateSession()
	if !validCreate(valid) {
		t.Fatal("valid session rejected")
	}
	tests := map[string]func(*CreateSession){
		"non-v7 session":     func(value *CreateSession) { value.SessionID = testFounderID },
		"live pvp":           func(value *CreateSession) { value.Mode = "live_pvp" },
		"leading-zero seed":  func(value *CreateSession) { value.Seed = "01" },
		"overflow seed":      func(value *CreateSession) { value.Seed = "18446744073709551616" },
		"array inputs":       func(value *CreateSession) { value.ScalingInputs = json.RawMessage("[]") },
		"array genesis":      func(value *CreateSession) { value.Genesis = json.RawMessage("[]") },
		"noncanonical input": func(value *CreateSession) { value.ScalingInputs = json.RawMessage(`{ "era":1,"trust_ppm":500000}`) },
		"duplicate input": func(value *CreateSession) {
			value.ScalingInputs = json.RawMessage(`{"era":1,"era":2,"trust_ppm":500000}`)
		},
		"decimal alias": func(value *CreateSession) {
			value.ScalingInputs = json.RawMessage(`{"era":1.0,"trust_ppm":500000}`)
		},
		"exponent alias": func(value *CreateSession) {
			value.ScalingInputs = json.RawMessage(`{"era":1e0,"trust_ppm":500000}`)
		},
		"flavor id": func(value *CreateSession) { value.MinigameID = "Shipping Wars" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := valid
			mutate(&value)
			if validCreate(value) {
				t.Fatal("invalid session accepted")
			}
		})
	}
}

func TestSessionClaimIntegration(t *testing.T) {
	db := minigameIntegrationDB(t)
	seedMinigameRun(t, db)
	repository, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	created, err := repository.create(ctx, testCreateSession())
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != StatusActive || created.Revision != 1 || !jsonObjectEqual(created.State, []byte(`{"turn":0}`)) ||
		!jsonObjectEqual(created.ScalingInputs, []byte(`{"era":1,"trust_ppm":500000}`)) {
		t.Fatalf("created=%+v", created)
	}
	if _, err := db.ExecContext(ctx, "UPDATE minigame_sessions SET scaling_inputs='{}' WHERE session_id=$1", testSessionID); err == nil {
		t.Fatal("frozen scaling inputs were mutable")
	}
	otherAccount := "018f0000-0000-4000-8000-000000000111"
	otherFounder := "018f0000-0000-4000-8000-000000000112"
	if _, err := db.ExecContext(ctx, "INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')", otherAccount); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)", otherAccount, otherFounder); err != nil {
		t.Fatal(err)
	}
	cloneSQL := "INSERT INTO minigame_sessions(session_id,minigame_id,founder_id,company_stream_id,run_seq,engine_ref,engine_version,constants_hash,scaling_inputs,seed,mode,genesis,state) " +
		"SELECT $1,minigame_id,$2,company_stream_id,run_seq,engine_ref,engine_version,$3,scaling_inputs,seed,mode,genesis,state FROM minigame_sessions WHERE session_id=$4"
	if _, err := db.ExecContext(ctx, cloneSQL, "018f0000-0000-7000-8000-000000000113", otherFounder, testHash, testSessionID); err == nil {
		t.Fatal("database accepted a Founder who does not own the Company stream")
	}
	otherHash := "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if _, err := db.ExecContext(ctx, cloneSQL, "018f0000-0000-7000-8000-000000000114", testFounderID, otherHash, testSessionID); err == nil {
		t.Fatal("database accepted a constants hash not pinned to the run")
	}

	start := make(chan struct{})
	results := make(chan Session, 2)
	errorsFound := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			result, claimErr := repository.claim(ctx, testFounderID, testSessionID)
			results <- result
			errorsFound <- claimErr
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(errorsFound)
	var winner Session
	busy := 0
	for result := range results {
		if result.ClaimToken != "" {
			winner = result
		}
	}
	for claimErr := range errorsFound {
		if errors.Is(claimErr, ErrSessionBusy) {
			busy++
		} else if claimErr != nil {
			t.Fatal(claimErr)
		}
	}
	if winner.Status != StatusClaimed || winner.Revision != 1 || winner.ClaimedAt == nil || busy != 1 {
		t.Fatalf("winner=%+v busy=%d", winner, busy)
	}
	if _, err := db.ExecContext(ctx, "UPDATE minigame_sessions SET state='{\"turn\":99}' WHERE session_id=$1", testSessionID); err == nil {
		t.Fatal("claim transition mutated state without a revision")
	}
	if _, err := db.ExecContext(ctx, "UPDATE minigame_sessions SET claimed_at=clock_timestamp()-interval '6 minutes' WHERE session_id=$1 AND claim_token=$2", testSessionID, winner.ClaimToken); err != nil {
		t.Fatal(err)
	}
	reclaimed, err := repository.claim(ctx, testFounderID, testSessionID)
	if err != nil {
		t.Fatal(err)
	}
	if reclaimed.ClaimToken == winner.ClaimToken || reclaimed.Revision != winner.Revision {
		t.Fatalf("stale claim was not safely replaced: old=%+v new=%+v", winner, reclaimed)
	}
	if _, err := repository.completePlay(ctx, testFounderID, testSessionID, winner.ClaimToken,
		json.RawMessage(`{"advance":1}`), json.RawMessage(`{"turn":1}`)); !errors.Is(err, ErrClaimLost) {
		t.Fatalf("replaced-token error=%v", err)
	}
	winner = reclaimed
	if _, err := repository.completePlay(ctx, testFounderID, testSessionID,
		"018f0000-0000-4000-8000-000000000999", json.RawMessage(`{"advance":1}`),
		json.RawMessage(`{"turn":1}`)); !errors.Is(err, ErrClaimLost) {
		t.Fatalf("wrong-token error=%v", err)
	}
	played, err := repository.completePlay(ctx, testFounderID, testSessionID, winner.ClaimToken,
		json.RawMessage(`{"advance":1}`), json.RawMessage(`{"turn":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if played.Status != StatusActive || played.Revision != 2 || played.ClaimToken != "" || !jsonObjectEqual(played.State, []byte(`{"turn":1}`)) {
		t.Fatalf("played=%+v", played)
	}

	claimed, err := repository.claim(ctx, testFounderID, testSessionID)
	if err != nil {
		t.Fatal(err)
	}
	rollbackTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	identity := resolutionIdentity{sessionID: testSessionID, minigameID: "combat.duel", founderID: testFounderID, companyStreamID: testStreamID,
		runSeq: 1, engineRef: "combat.duel", engineVersion: "1.0.0", constantsHash: testHash, claimToken: claimed.ClaimToken}
	if _, err := resolveTx(ctx, rollbackTx, identity, json.RawMessage(`{"advance":1}`), json.RawMessage(`{"turn":2}`),
		json.RawMessage(`{"outcome":"complete","rating_delta":null,"score_facts":[]}`), json.RawMessage(`{"outcome":"applied"}`), 2, 2); err != nil {
		_ = rollbackTx.Rollback()
		t.Fatal(err)
	}
	if err := rollbackTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	afterRollback, err := repository.Load(ctx, testFounderID, testSessionID)
	if err != nil {
		t.Fatal(err)
	}
	if afterRollback.Status != StatusClaimed || afterRollback.Revision != 2 || afterRollback.ClaimToken != claimed.ClaimToken || afterRollback.Result != nil {
		t.Fatalf("resolve rollback leaked=%+v", afterRollback)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := resolveTx(ctx, tx, identity, json.RawMessage(`{"advance":1}`), json.RawMessage(`{"turn":2}`),
		json.RawMessage(`{"outcome":"complete","rating_delta":null,"score_facts":[]}`), json.RawMessage(`{"outcome":"applied"}`), 2, 2)
	if err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if resolved.Status != StatusResolved || resolved.Revision != 3 || resolved.ResolvedAt == nil {
		t.Fatalf("resolved=%+v", resolved)
	}
	if _, err := repository.claim(ctx, testFounderID, testSessionID); !errors.Is(err, ErrSessionGone) {
		t.Fatalf("resolved claim error=%v", err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE minigame_sessions SET state='{}' WHERE session_id=$1", testSessionID); err == nil {
		t.Fatal("resolved session was mutable")
	}
	if _, err := db.ExecContext(ctx, "UPDATE minigame_session_commands SET command='{}' WHERE session_id=$1", testSessionID); err == nil {
		t.Fatal("session command log was mutable")
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM minigame_session_commands WHERE session_id=$1", testSessionID); err == nil {
		t.Fatal("session command log was deletable")
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM minigame_sessions WHERE session_id=$1", testSessionID); err != nil {
		t.Fatalf("parent retention cascade failed: %v", err)
	}
	var retainedCommands int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM minigame_session_commands WHERE session_id=$1", testSessionID).Scan(&retainedCommands); err != nil || retainedCommands != 0 {
		t.Fatalf("parent cascade retained commands=%d err=%v", retainedCommands, err)
	}
}

func TestAPIReceiptAndCurrentSessionIntegration(t *testing.T) {
	db := minigameIntegrationDB(t)
	seedMinigameRun(t, db)
	repository, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	created, err := repository.create(ctx, testCreateSession())
	if err != nil {
		t.Fatal(err)
	}
	current, found, err := repository.Current(ctx, testFounderID)
	if err != nil || !found || current.SessionID != created.SessionID {
		t.Fatalf("current=%+v found=%v err=%v", current, found, err)
	}
	second := testCreateSession()
	second.SessionID = "018f0000-0000-7000-8000-000000000105"
	if _, err := repository.create(ctx, second); err == nil {
		t.Fatal("database admitted two active minigame sessions for one Founder")
	}

	createResponse := json.RawMessage(`{"kind":"pitch","session_id":"018f0000-0000-7000-8000-000000000104"}`)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := InsertCreateReceiptTx(ctx, tx, testFounderID, "create-1", testHash, testSessionID, createResponse); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	receipt, found, err := repository.CreateReceipt(ctx, testFounderID, "create-1", testHash)
	if err != nil || !found || receipt.SessionID != testSessionID || !bytes.Equal(receipt.Response, createResponse) {
		t.Fatalf("create receipt=%+v found=%v err=%v", receipt, found, err)
	}
	otherHash := "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if _, _, err := repository.CreateReceipt(ctx, testFounderID, "create-1", otherHash); !errors.Is(err, ErrAPIIdempotency) {
		t.Fatalf("create idempotency mismatch=%v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE minigame_create_receipts SET response='{}' WHERE founder_id=$1`, testFounderID); err == nil {
		t.Fatal("create receipt was mutable")
	}

	claimed, err := repository.claim(ctx, testFounderID, testSessionID)
	if err != nil {
		t.Fatal(err)
	}
	commandResponse := json.RawMessage(`{"revision":2,"session_id":"018f0000-0000-7000-8000-000000000104"}`)
	played, err := repository.CompletePlayWithReceipt(ctx, testFounderID, testSessionID, claimed.ClaimToken,
		json.RawMessage(`{"advance":1}`), json.RawMessage(`{"turn":1}`), "command-1", testHash, commandResponse)
	if err != nil || played.Revision != 2 {
		t.Fatalf("played=%+v err=%v", played, err)
	}
	commandReceipt, found, err := repository.CommandReceipt(ctx, testFounderID, testSessionID, "command-1", testHash)
	if err != nil || !found || !bytes.Equal(commandReceipt.Response, commandResponse) {
		t.Fatalf("command receipt=%+v found=%v err=%v", commandReceipt, found, err)
	}
	if _, _, err := repository.CommandReceipt(ctx, testFounderID, testSessionID, "command-1", otherHash); !errors.Is(err, ErrAPIIdempotency) {
		t.Fatalf("command idempotency mismatch=%v", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM minigame_command_receipts WHERE session_id=$1`, testSessionID); err == nil {
		t.Fatal("command receipt was deletable")
	}

	stale, err := repository.claim(ctx, testFounderID, testSessionID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE minigame_sessions SET claimed_at=clock_timestamp()-interval '6 minutes' WHERE session_id=$1", testSessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.claim(ctx, testFounderID, testSessionID); err != nil {
		t.Fatal(err)
	}
	staleTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := InsertCommandReceiptTx(ctx, staleTx, testFounderID, testSessionID, stale.ClaimToken,
		"command-stale", testHash, json.RawMessage(`{"revision":3}`)); !errors.Is(err, ErrClaimLost) {
		_ = staleTx.Rollback()
		t.Fatalf("stale claim published receipt: %v", err)
	}
	_ = staleTx.Rollback()
}

func TestServiceIntegration(t *testing.T) {
	db := minigameIntegrationDB(t)
	seedMinigameRun(t, db)
	repository, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	tenants, err := NewTenantRegistry(fixtureTenant{})
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(repository, tenants)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	started, err := service.Start(ctx, StartRequest{
		SessionID: testSessionID, MinigameID: "combat.duel", FounderID: testFounderID,
		CompanyStreamID: testStreamID, RunSeq: 1, EngineRef: "fixture.counter", EngineVersion: "1.0.0",
		ConstantsHash: testHash, ScalingInputs: map[string]int64{"era": 1, "trust_ppm": 500_000},
		Seed: "8", Mode: ModeSolo,
	})
	if err != nil || started.Revision != 1 || !jsonObjectEqual(started.State, []byte(`{"done":false,"total":1}`)) {
		t.Fatalf("start=%+v err=%v", started, err)
	}
	v2Tenants, err := NewTenantRegistry(fixtureTenant{version: "2.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	v2Service, err := NewService(repository, v2Tenants)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []json.RawMessage{
		json.RawMessage(`{"add":1,"finish":false}`),
		json.RawMessage(`{"add":1,"finish":true}`),
	} {
		if _, err := v2Service.Play(ctx, PlayRequest{FounderID: testFounderID, SessionID: testSessionID,
			ExpectedRevision: 1, Command: command}); !errors.Is(err, ErrTenantVersion) {
			t.Fatalf("version-drift command=%s error=%v", command, err)
		}
		afterVersionDrift, loadErr := repository.Load(ctx, testFounderID, testSessionID)
		if loadErr != nil || afterVersionDrift.Status != StatusActive || afterVersionDrift.Revision != 1 ||
			afterVersionDrift.ClaimToken != "" || !jsonObjectEqual(afterVersionDrift.State, started.State) {
			t.Fatalf("version-drift mutated session=%+v err=%v", afterVersionDrift, loadErr)
		}
	}
	if _, err := service.Play(ctx, PlayRequest{FounderID: testFounderID, SessionID: testSessionID,
		ExpectedRevision: 2, Command: json.RawMessage(`{"add":4,"finish":false}`)}); err != ErrSessionRevision {
		t.Fatalf("stale revision error=%v", err)
	}
	afterStale, err := repository.Load(ctx, testFounderID, testSessionID)
	if err != nil || afterStale.Status != StatusActive || afterStale.Revision != 1 {
		t.Fatalf("stale release=%+v err=%v", afterStale, err)
	}
	if _, err := service.Play(ctx, PlayRequest{FounderID: testFounderID, SessionID: testSessionID,
		ExpectedRevision: 1, Command: json.RawMessage(`{"add":4,"finish":false,"score":99}`)}); !errors.Is(err, ErrTenantRejected) {
		t.Fatalf("schema rejection error=%v", err)
	}
	played, err := service.Play(ctx, PlayRequest{FounderID: testFounderID, SessionID: testSessionID,
		ExpectedRevision: 1, Command: json.RawMessage(`{"add":4,"finish":false}`)})
	if err != nil || played.Resolution != nil || played.Session.Revision != 2 ||
		!jsonObjectEqual(played.Session.State, []byte(`{"done":false,"total":5}`)) {
		t.Fatalf("play=%+v err=%v", played, err)
	}
	terminal, err := service.Play(ctx, PlayRequest{FounderID: testFounderID, SessionID: testSessionID,
		ExpectedRevision: 2, Command: json.RawMessage(`{"add":2,"finish":true}`)})
	if err != nil || terminal.Resolution == nil || terminal.Resolution.Result() == nil || terminal.Session.Status != StatusClaimed ||
		!jsonObjectEqual(terminal.Session.State, []byte(`{"done":true,"total":7}`)) {
		t.Fatalf("terminal=%+v err=%v", terminal, err)
	}
	storedClaim, err := repository.Load(ctx, testFounderID, testSessionID)
	if err != nil || storedClaim.Status != StatusClaimed || !jsonObjectEqual(storedClaim.State, []byte(`{"done":false,"total":5}`)) {
		t.Fatalf("pre-resolve store=%+v err=%v", storedClaim, err)
	}
	driftTenants, err := NewTenantRegistry(fixtureTenant{bias: 1})
	if err != nil {
		t.Fatal(err)
	}
	driftService, err := NewService(repository, driftTenants)
	if err != nil {
		t.Fatal(err)
	}
	driftTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := driftService.validateCertifiedTx(ctx, driftTx, terminal.Resolution); !errors.Is(err, ErrTenantDivergence) {
		_ = driftTx.Rollback()
		t.Fatalf("same-version engine drift error=%v", err)
	}
	if err := driftTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	afterDrift, err := repository.Load(ctx, testFounderID, testSessionID)
	if err != nil || afterDrift.Status != StatusClaimed || afterDrift.Revision != 2 ||
		!jsonObjectEqual(afterDrift.State, storedClaim.State) {
		t.Fatalf("engine drift mutated session=%+v err=%v", afterDrift, err)
	}
	forged := *terminal.Resolution
	forged.bytes = json.RawMessage(`{"garbage":true}`)
	forgedTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.validateCertifiedTx(ctx, forgedTx, &forged); !errors.Is(err, ErrInvalidSession) {
		_ = forgedTx.Rollback()
		t.Fatalf("forged result error=%v", err)
	}
	if err := forgedTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	crossCompany := *terminal.Resolution
	crossCompany.identity.companyStreamID = "018f0000-0000-4000-8000-000000000119"
	crossTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.validateCertifiedTx(ctx, crossTx, &crossCompany); !errors.Is(err, ErrClaimLost) {
		_ = crossTx.Rollback()
		t.Fatalf("cross-company result error=%v", err)
	}
	if err := crossTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.validateCertifiedTx(ctx, tx, terminal.Resolution); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	resolved, err := resolveTx(ctx, tx, terminal.Resolution.identity, terminal.Resolution.command,
		terminal.Resolution.state, terminal.Resolution.bytes, json.RawMessage(`{"outcome":"applied"}`), 2, 2)
	if err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if resolved.Status != StatusResolved || resolved.Revision != 3 ||
		!jsonObjectEqual(resolved.State, []byte(`{"done":true,"total":7}`)) {
		t.Fatalf("resolved=%+v", resolved)
	}
	rows, err := db.QueryContext(ctx, "SELECT seq,command,applied_revision FROM minigame_session_commands WHERE session_id=$1 ORDER BY seq", testSessionID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	wantCommands := []string{`{"add":4,"finish":false}`, `{"add":2,"finish":true}`}
	for index, want := range wantCommands {
		if !rows.Next() {
			t.Fatalf("missing command row %d", index+1)
		}
		var sequence, appliedRevision int64
		var command []byte
		if err := rows.Scan(&sequence, &command, &appliedRevision); err != nil || sequence != int64(index+1) ||
			appliedRevision != int64(index+2) || !bytes.Equal(command, []byte(want)) {
			t.Fatalf("command row=%d/%s/%d err=%v", sequence, command, appliedRevision, err)
		}
	}
	if rows.Next() || rows.Err() != nil {
		t.Fatalf("unexpected command rows err=%v", rows.Err())
	}
}

func testCreateSession() CreateSession {
	return CreateSession{
		SessionID: testSessionID, MinigameID: "combat.duel", FounderID: testFounderID,
		CompanyStreamID: testStreamID, RunSeq: 1, EngineRef: "combat.duel", EngineVersion: "1.0.0",
		ConstantsHash: testHash, ScalingInputs: json.RawMessage(`{"era":1,"trust_ppm":500000}`),
		Seed: "18446744073709551615", Mode: ModeSolo, Genesis: json.RawMessage(`{"turn":0}`),
	}
}

func jsonObjectEqual(left, right []byte) bool {
	var leftValue, rightValue map[string]any
	return json.Unmarshal(left, &leftValue) == nil && json.Unmarshal(right, &rightValue) == nil &&
		len(leftValue) == len(rightValue) && string(mustJSON(leftValue)) == string(mustJSON(rightValue))
}

func mustJSON(value any) []byte {
	encoded, _ := json.Marshal(value)
	return encoded
}

func minigameIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := save.OpenPostgres(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := save.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedMinigameRun(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	statements := []struct {
		query string
		args  []any
	}{
		{"INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')", []any{testAccountID}},
		{"INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)", []any{testAccountID, testFounderID}},
		{"INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES($1,'founder',$2,'company')", []any{testStreamID, testFounderID}},
		{"INSERT INTO catalog_sets(constants_hash) VALUES($1)", []any{testHash}},
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	var epochID int64
	if err := tx.QueryRowContext(ctx, "INSERT INTO epochs(name,started_at,changelog_ref) VALUES('minigame test',clock_timestamp(),'changelog/epoch-1.md') RETURNING epoch_id").Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)", epochID, testHash); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO run_epochs(company_stream_id,run_seq,epoch_id,constants_hash,engine_version,build_vcs_hash,seed) VALUES($1,1,$2,$3,'0.3.24','test','1')", testStreamID, epochID, testHash); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO run_genesis(company_stream_id,run_seq,state,version,constants_hash) VALUES($1,1,'{}',16,$2)", testStreamID, testHash); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
