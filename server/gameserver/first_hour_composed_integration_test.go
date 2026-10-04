package gameserver

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/harness"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/publicapi"
	"cloud-clicker/server/save"
)

func TestComposedGameserverReplaysRatifiedFirstHourAtPinnedSeed(t *testing.T) {
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
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), cleanDatabase) })

	root := filepathRoot(t)
	suite, err := harness.LoadFirstHourSuite(root,
		"balance/testdata/t0-t1/harness-scenario-v1.json",
		"balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec harness.RunSpec
	for _, candidate := range suite.Scenario.Runs {
		if candidate.PolicyID == "casual.t0_t1" {
			spec = candidate
			break
		}
	}
	if spec.PolicyID == "" {
		t.Fatal("ratified first-hour scenario omitted casual.t0_t1")
	}
	experiment := harness.FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0",
		RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}
	expected, commands := suite.RunExperimentScript(spec, 0, experiment)
	if expected.Outcome != "completed" || expected.Ending == nil || len(commands) == 0 {
		t.Fatalf("headless authority result=%+v commands=%d", expected, len(commands))
	}
	wantMilestones := map[string]int64{}
	for _, milestone := range expected.Milestones {
		if milestone.FirstMS == nil {
			t.Fatalf("headless authority did not reach %s", milestone.ID)
		}
		wantMilestones[milestone.ID] = *milestone.FirstMS
	}
	if len(wantMilestones) != 7 {
		t.Fatalf("headless milestone cardinality=%d want=7", len(wantMilestones))
	}

	base := time.Now().UTC().Truncate(time.Second)
	clock := &mutableClock{now: base}
	composition, err := Compose(ctx, CompositionConfig{
		PublicCursorKeys: testPublicCursorKeys(),
		DB:               db, RepositoryRoot: root, ServerID: "018f0000-0000-4000-8000-000000000331", ActivityBracket: "activity.standard",
		SigningKeys:   account.SigningKeys{CurrentID: "first-hour-composed", Current: bytes.Repeat([]byte{0x52}, 32)},
		BootstrapKeys: account.BootstrapReceiptKeys{CurrentID: "first-hour-composed", Current: bytes.Repeat([]byte{0x53}, 32)},
		Clock:         clock.Time,
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
	bootstrapResponse := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/bootstrap", "",
		fmt.Sprintf(`{"idempotency_key":%q}`, strings.Repeat("cd", 32)))
	if bootstrapResponse.StatusCode != http.StatusCreated {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapResponse.StatusCode, responseBody(bootstrapResponse))
	}
	var bootstrap bootstrapResponseEnvelope
	decodeCompositionResponse(t, bootstrapResponse, &bootstrap)
	founder, err := composition.Accounts.ActiveFounder(ctx, bootstrap.Account.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	store, err := save.NewStore(db, composition.Catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}

	actualMilestones := map[string]int64{}
	var runOneStream, runTwoStream string
	var runTwoGateState *save.State
	terminalIntentIDs := []string{}
	for index, command := range commands {
		clock.Set(base.Add(time.Duration(command.AtMS) * time.Millisecond))
		active, err := composition.Accounts.ActiveCompanyState(ctx, bootstrap.Account.AccountID)
		if err != nil {
			t.Fatalf("command %d active company: %v", index, err)
		}
		bundle, ok := composition.Catalogs.bundle(active.ConstantsHash)
		if !ok {
			t.Fatalf("command %d missing catalog %s", index, active.ConstantsHash)
		}
		before, err := save.RestoreState(active.State, active.Version, bundle.Economy, economy.ScopeCompany, clock.Time())
		if err != nil {
			t.Fatalf("command %d restore: %v", index, err)
		}
		if before.RunSeq == 1 && runOneStream == "" {
			runOneStream = active.StreamID
		}
		if before.RunSeq == 2 && runTwoStream == "" {
			runTwoStream = active.StreamID
		}
		request := command.Request
		request.IntentID = composedFirstHourIntentID(index)
		request.ExpectedRevision = active.Revision
		if request.Kind == production.IntentWindDown {
			request.ExpectedFounderRevision = latestFounderRevision(t, ctx, db, founder.ID)
		}
		beforeAttended := composedRunAttended(t, before, bundle, command.Mode, clock.Time())
		payload := composedFirstHourIntent(t, request)
		result, err := composition.Production.Handle(ctx, active.StreamID, command.Mode, clock.Time(), payload)
		if err != nil || !bytes.Contains(result.Receipt, []byte(`"outcome":"applied"`)) {
			attended, attendedErr := prestigecore.AttendedMS(before, clock.Time())
			cash, _ := before.Ledger.Balance("company.cash")
			cost := decimal.Zero
			if request.Kind == production.IntentBuyGenerator {
				cost, _ = bundle.Economy.BulkCost(request.GeneratorID, before.GeneratorCounts[request.GeneratorID], 1)
			}
			t.Fatalf("command %d at=%d kind=%s generator=%s upgrade=%s run=%d expected_ending=%+v owned=%v receipt=%s err=%v",
				index, command.AtMS, request.Kind, request.GeneratorID, request.UpgradeID, before.RunSeq, expected.Ending,
				map[string]any{"upgrades": before.UpgradesOwned, "run_started_at": before.RunStartedAt, "evaluated_through": before.EvaluatedThrough,
					"offline_spans": before.OfflineSpans, "attended_ms": attended, "attended_err": attendedErr, "cash": cash.String(), "cost": cost.String(),
					"generator_count": before.GeneratorCounts[request.GeneratorID]}, result.Receipt, err)
		}
		afterActive, err := composition.Accounts.ActiveCompanyState(ctx, bootstrap.Account.AccountID)
		if err != nil {
			t.Fatal(err)
		}
		afterBundle, ok := composition.Catalogs.bundle(afterActive.ConstantsHash)
		if !ok {
			t.Fatal("post-command catalog unavailable")
		}
		after, err := save.RestoreState(afterActive.State, afterActive.Version, afterBundle.Economy, economy.ScopeCompany, clock.Time())
		if err != nil {
			t.Fatal(err)
		}
		afterAttended := beforeAttended
		if before.RunSeq == after.RunSeq {
			afterAttended, err = prestigecore.AttendedMS(after, clock.Time())
			if err != nil {
				t.Fatal(err)
			}
		}
		recordFirstHourCommandMilestone(actualMilestones, request.Kind, before.RunSeq, afterAttended)
		if before.RunSeq == 1 && after.RunSeq == 2 {
			actualMilestones["milestone.scripted_failure"] = beforeAttended
			terminalIntentIDs = append(terminalIntentIDs, request.IntentID)
			var fiscalSweepEvents int
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events
				WHERE intent_id=$1 AND kind='fiscal_period_harvested.v1'`, request.IntentID).Scan(&fiscalSweepEvents); err != nil || fiscalSweepEvents != 1 {
				t.Fatalf("automatic Exit fiscal sweep events=%d err=%v", fiscalSweepEvents, err)
			}
			founderHistory, historyErr := store.LoadFounderHistory(ctx, activeFounderStreamID(t, ctx, db, founder.ID))
			if historyErr != nil || production.VerifyFounderHistory(founderHistory, composition.Catalogs.replay) != production.ReplayVerified {
				t.Fatalf("run-1 Founder replay verdict=%s err=%v",
					production.VerifyFounderHistory(founderHistory, composition.Catalogs.replay), historyErr)
			}
		}
		if request.Kind == production.IntentCrossGate && before.RunSeq == 2 {
			actualMilestones["milestone.run2_garage_gate"] = afterAttended
			runTwoGateState = after
		}
		if before.RunSeq == 2 && after.RunSeq == 3 {
			actualMilestones["milestone.first_elective_exit"] = latestFounderAgeMS(t, ctx, db, founder.ID)
			terminalIntentIDs = append(terminalIntentIDs, request.IntentID)
		}
	}
	if !equalFirstHourMilestones(wantMilestones, actualMilestones) {
		t.Fatalf("composed milestones=%v want=%v", actualMilestones, wantMilestones)
	}
	if runTwoGateState == nil || runTwoGateState.RunSeq != 2 || !runTwoGateState.GatesCrossed["gate.t0_to_t1"] {
		t.Fatalf("run-2 gate state=%+v", runTwoGateState)
	}
	switch expected.Ending.Branch {
	case "acquihire":
		cash, _ := runTwoGateState.Ledger.Balance("company.cash")
		if !cash.Gt(decimal.Zero) {
			t.Fatalf("acquihire starter missing at run-2 gate: %s", cash)
		}
	case "burnout":
		if runTwoGateState.GeneratorProvisioned["generator.beige_tower"] != 10 {
			t.Fatal("burnout generated starter missing")
		}
	case "pivot":
		if !runTwoGateState.UpgradesOwned["upgrade.reply_all_macro"] {
			t.Fatal("pivot starter missing")
		}
	}
	if len(terminalIntentIDs) != 2 {
		t.Fatalf("terminal intents=%v", terminalIntentIDs)
	}
	for _, intentID := range terminalIntentIDs {
		assertFirstHourTerminalOrder(t, ctx, db, intentID)
	}
	if runOneStream == "" || runTwoStream == "" {
		t.Fatalf("run streams run1=%q run2=%q", runOneStream, runTwoStream)
	}
	waitFirstHourRunVerifiedAndArchived(t, ctx, db, runOneStream, 1)
	waitFirstHourRunVerifiedAndArchived(t, ctx, db, runTwoStream, 2)
	founderHistory, err := store.LoadFounderHistory(ctx, activeFounderStreamID(t, ctx, db, founder.ID))
	if err != nil || production.VerifyFounderHistory(founderHistory, composition.Catalogs.replay) != production.ReplayVerified {
		t.Fatalf("Founder persisted replay verdict=%s err=%v", production.VerifyFounderHistory(founderHistory, composition.Catalogs.replay), err)
	}
	assertFirstHourPublicBoardAfterAccountDeletion(t, ctx, db, httpServer, bootstrap, founder.ID)
}

type firstHourRightsCounts struct {
	boards, markers, queue, archives, events int
}

type firstHourBoardEntry struct {
	FounderID string `json:"founder_id"`
	RunID     string `json:"run_id"`
}

func firstHourRightsCensus(t *testing.T, ctx context.Context, db *sql.DB, founderID string) firstHourRightsCounts {
	t.Helper()
	var counts firstHourRightsCounts
	err := db.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM verified_runs WHERE founder_id=$1),
		(SELECT count(*) FROM verification_projection_events marker JOIN verified_runs board ON board.event_id=marker.event_id WHERE board.founder_id=$1),
		(SELECT count(*) FROM verification_queue queue JOIN save_streams stream ON stream.id=queue.company_stream_id WHERE stream.owner_id=$1),
		(SELECT count(*) FROM run_log_archive archive JOIN save_streams stream ON stream.id=archive.company_stream_id WHERE stream.owner_id=$1),
		(SELECT count(*) FROM events event JOIN save_streams stream ON stream.id=event.stream_id WHERE stream.owner_id=$1)`, founderID).
		Scan(&counts.boards, &counts.markers, &counts.queue, &counts.archives, &counts.events)
	if err != nil {
		t.Fatal(err)
	}
	return counts
}

func firstHourPublicBoardRows(t *testing.T, server *httptest.Server, category string, epoch int64, mandate int, variables string) []firstHourBoardEntry {
	t.Helper()
	response := compositionRequest(t, server.Client(), http.MethodGet,
		fmt.Sprintf("%s/api/public/v1/boards/%s?epoch=%d&mandate=%d&variables=%s", server.URL, category, epoch, mandate, variables), "", "")
	body := responseBody(response)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("public board status=%d body=%s", response.StatusCode, body)
	}
	var page struct {
		Items []firstHourBoardEntry `json:"items"`
	}
	if err := json.Unmarshal([]byte(body), &page); err != nil {
		t.Fatalf("public board decode: %v body=%s", err, body)
	}
	return page.Items
}

func assertFirstHourPublicBoardAfterAccountDeletion(t *testing.T, ctx context.Context, db *sql.DB, server *httptest.Server,
	bootstrap bootstrapResponseEnvelope, founderID string) {
	t.Helper()
	// RP-118: the script's real verifier and projector must produce the row;
	// no board or log row is seeded by this rights witness.
	deadline := time.Now().Add(5 * time.Second)
	var category, runID string
	var epoch int64
	var mandate int
	for time.Now().Before(deadline) {
		err := db.QueryRowContext(ctx, `SELECT category_id,run_id,epoch_id,mandate_level FROM verified_runs
			WHERE founder_id=$1 AND category_id='any_percent' ORDER BY run_id LIMIT 1`, founderID).
			Scan(&category, &runID, &epoch, &mandate)
		if err == nil {
			break
		}
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	if runID == "" {
		t.Fatal("first-hour projector produced no any_percent row")
	}
	variables, err := publicapi.EncodeBoardVariables(publicapi.BoardVariables{})
	if err != nil {
		t.Fatal(err)
	}
	contains := func(rows []firstHourBoardEntry) bool {
		for _, row := range rows {
			if row.FounderID == founderID && row.RunID == runID {
				return true
			}
		}
		return false
	}
	beforeRows := firstHourPublicBoardRows(t, server, category, epoch, mandate, variables)
	if !contains(beforeRows) {
		t.Fatalf("projected run %s absent from public board before deletion: %+v", runID, beforeRows)
	}
	for _, row := range beforeRows {
		if row.FounderID != founderID {
			t.Fatalf("unmatched Founder appeared on isolated pre-delete board: %+v", row)
		}
	}
	otherVariables, err := publicapi.EncodeBoardVariables(publicapi.BoardVariables{Glitched: 1})
	if err != nil {
		t.Fatal(err)
	}
	if rows := firstHourPublicBoardRows(t, server, category, epoch, mandate, otherVariables); contains(rows) {
		t.Fatalf("different board-variable partition returned projected run %s", runID)
	}
	before := firstHourRightsCensus(t, ctx, db, founderID)
	if before.boards < 2 || before.markers < 2 || before.queue < 2 || before.archives < 2 || before.events == 0 {
		t.Fatalf("joined pre-delete producers incomplete: %+v", before)
	}
	var linked int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM account_founders WHERE founder_id=$1 AND account_id=$2 AND archived_at IS NULL`,
		founderID, bootstrap.Account.AccountID).Scan(&linked); err != nil || linked != 1 {
		t.Fatalf("pre-delete active account-Founder link=%d err=%v", linked, err)
	}
	refresh := compositionRequest(t, server.Client(), http.MethodPost, server.URL+"/api/v1/session/refresh", "",
		fmt.Sprintf(`{"refresh_token":%q}`, bootstrap.Session.RefreshToken))
	if refresh.StatusCode != http.StatusOK {
		t.Fatalf("rights-witness refresh status=%d body=%s", refresh.StatusCode, responseBody(refresh))
	}
	var rotated account.TokenPair
	decodeCompositionResponse(t, refresh, &rotated)
	deletion := compositionRequest(t, server.Client(), http.MethodDelete, server.URL+"/api/v1/account", rotated.AccessToken, "")
	if deletion.StatusCode != http.StatusNoContent {
		t.Fatalf("rights-witness delete status=%d body=%s", deletion.StatusCode, responseBody(deletion))
	}
	_ = responseBody(deletion)
	var accountRows, founderRows, archivedStreams, streamRows int
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM accounts WHERE account_id=$1),
		(SELECT count(*) FROM account_founders WHERE founder_id=$2 AND account_id IS NULL AND archived_at IS NOT NULL),
		(SELECT count(*) FROM save_streams WHERE owner_kind='founder' AND owner_id=$2 AND archived_at IS NOT NULL),
		(SELECT count(*) FROM save_streams WHERE owner_kind='founder' AND owner_id=$2)`,
		bootstrap.Account.AccountID, founderID).Scan(&accountRows, &founderRows, &archivedStreams, &streamRows); err != nil ||
		accountRows != 0 || founderRows != 1 || streamRows < 2 || archivedStreams != streamRows {
		t.Fatalf("post-delete account=%d archived_founder=%d archived_streams=%d streams=%d err=%v",
			accountRows, founderRows, archivedStreams, streamRows, err)
	}
	after := firstHourRightsCensus(t, ctx, db, founderID)
	if after != before {
		t.Fatalf("joined rights families changed on deletion: before=%+v after=%+v", before, after)
	}
	afterRows := firstHourPublicBoardRows(t, server, category, epoch, mandate, variables)
	if !contains(afterRows) {
		t.Fatalf("deleted Founder's projected run %s is absent from public board: %+v", runID, afterRows)
	}
	for _, row := range afterRows {
		if row.FounderID != founderID {
			t.Fatalf("unmatched Founder appeared on isolated post-delete board: %+v", row)
		}
	}
	if rows := firstHourPublicBoardRows(t, server, category, epoch, mandate, otherVariables); contains(rows) {
		t.Fatalf("different board-variable partition returned deleted Founder's run %s", runID)
	}
	t.Logf("joined post-delete retained rows: boards=%d markers=%d queue=%d archives=%d events=%d; public board still contains the deleted Founder's run",
		after.boards, after.markers, after.queue, after.archives, after.events)
}

func waitFirstHourRunVerifiedAndArchived(t *testing.T, ctx context.Context, db *sql.DB, streamID string, runSeq int64) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var status, verdict string
	var archiveHash string
	var terminalSeq int64
	for time.Now().Before(deadline) {
		err := db.QueryRowContext(ctx, `SELECT queue.status,COALESCE(queue.verdict,''),archive.sha256,archive.terminal_seq
			FROM verification_queue queue JOIN run_log_archive archive
			  ON archive.company_stream_id=queue.company_stream_id AND archive.run_seq=queue.run_seq
			WHERE queue.company_stream_id=$1 AND queue.run_seq=$2`, streamID, runSeq).
			Scan(&status, &verdict, &archiveHash, &terminalSeq)
		if err == nil && status == "verified" && verdict == "verified" && strings.HasPrefix(archiveHash, "sha256:") && terminalSeq > 0 {
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("run %d persisted verification status=%q verdict=%q archive=%q terminal_seq=%d", runSeq, status, verdict, archiveHash, terminalSeq)
}

func composedFirstHourIntentID(index int) string {
	value := index + 1
	return fmt.Sprintf("0198cccc-%04x-7%03x-8%03x-%012x", value&0xffff, value&0xfff, value&0xfff, value)
}

func composedFirstHourIntent(t *testing.T, request production.IntentRequest) []byte {
	t.Helper()
	payload := map[string]any{"intent_id": request.IntentID, "kind": request.Kind, "expected_revision": request.ExpectedRevision}
	switch request.Kind {
	case production.IntentPerformManualBatch:
		payload["action_id"], payload["count"], payload["window_ms"] = request.ActionID, request.Count, request.WindowMS
	case production.IntentBuyGenerator:
		payload["generator_id"] = request.GeneratorID
		payload["count"] = map[string]any{"mode": request.CountMode, "value": request.Count}
	case production.IntentBuyUpgrade:
		payload["upgrade_id"] = request.UpgradeID
	case production.IntentCrossGate:
		payload["gate_id"], payload["route_id"] = request.GateID, nil
	case production.IntentWindDown:
		payload["expected_founder_revision"] = request.ExpectedFounderRevision
	default:
		t.Fatalf("unsupported first-hour command %q", request.Kind)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func latestFounderRevision(t *testing.T, ctx context.Context, db *sql.DB, founderID string) int64 {
	t.Helper()
	var revision int64
	if err := db.QueryRowContext(ctx, `SELECT r.revision FROM save_streams s JOIN LATERAL (
		SELECT revision FROM save_revisions WHERE stream_id=s.id ORDER BY revision DESC LIMIT 1
	) r ON true WHERE s.owner_kind='founder' AND s.owner_id=$1 AND s.scope='founder' AND s.archived_at IS NULL`, founderID).Scan(&revision); err != nil {
		t.Fatal(err)
	}
	return revision
}

func latestFounderAgeMS(t *testing.T, ctx context.Context, db *sql.DB, founderID string) int64 {
	t.Helper()
	var state []byte
	if err := db.QueryRowContext(ctx, `SELECT r.state FROM save_streams s JOIN LATERAL (
		SELECT state FROM save_revisions WHERE stream_id=s.id ORDER BY revision DESC LIMIT 1
	) r ON true WHERE s.owner_kind='founder' AND s.owner_id=$1 AND s.scope='founder' AND s.archived_at IS NULL`, founderID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	var wire struct {
		AgeMS int64 `json:"age_ms"`
	}
	if err := json.Unmarshal(state, &wire); err != nil {
		t.Fatal(err)
	}
	return wire.AgeMS
}

func composedRunAttended(t *testing.T, state *save.State, bundle production.CatalogBundle, mode production.EvaluationMode, now time.Time) int64 {
	t.Helper()
	clone := *state
	clone.OfflineSpans = append([]save.OfflineSpan(nil), state.OfflineSpans...)
	if mode == production.ModeOffline && now.After(clone.EvaluatedThrough) {
		if err := prestigecore.RecordOfflineSpan(&clone, clone.EvaluatedThrough, now, bundle.Prestige.CatchupCeilingMS); err != nil {
			t.Fatal(err)
		}
	}
	attended, err := prestigecore.AttendedMS(&clone, now)
	if err != nil {
		t.Fatal(err)
	}
	return attended
}

func activeFounderStreamID(t *testing.T, ctx context.Context, db *sql.DB, founderID string) string {
	t.Helper()
	var streamID string
	if err := db.QueryRowContext(ctx, `SELECT id FROM save_streams
		WHERE owner_kind='founder' AND owner_id=$1 AND scope='founder' AND archived_at IS NULL`, founderID).Scan(&streamID); err != nil {
		t.Fatal(err)
	}
	return streamID
}

func recordFirstHourCommandMilestone(values map[string]int64, kind string, runSeq, attendedMS int64) {
	if _, ok := values["milestone.first_manual"]; !ok && kind == production.IntentPerformManualBatch {
		values["milestone.first_manual"] = attendedMS
	}
	if _, ok := values["milestone.first_generator"]; !ok && kind == production.IntentBuyGenerator {
		values["milestone.first_generator"] = attendedMS
	}
	if _, ok := values["milestone.first_upgrade"]; !ok && kind == production.IntentBuyUpgrade {
		values["milestone.first_upgrade"] = attendedMS
	}
	if kind == production.IntentCrossGate && runSeq == 1 {
		values["milestone.garage_gate"] = attendedMS
	}
}

func equalFirstHourMilestones(left, right map[string]int64) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func assertFirstHourTerminalOrder(t *testing.T, ctx context.Context, db *sql.DB, intentID string) {
	t.Helper()
	rows, err := db.QueryContext(ctx, `SELECT kind FROM events WHERE intent_id=$1 ORDER BY event_seq`, intentID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	kinds := []string{}
	for rows.Next() {
		var kind string
		if err := rows.Scan(&kind); err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, kind)
	}
	if err := rows.Err(); err != nil || len(kinds) < 3 || kinds[len(kinds)-2] != "run_ended" || kinds[len(kinds)-1] != "run_started" {
		t.Fatalf("terminal event order for %s=%v err=%v", intentID, kinds, err)
	}
}
