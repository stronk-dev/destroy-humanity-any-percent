package production

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/minigame"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type prestigeNoopProjector struct{}

func (prestigeNoopProjector) Project(context.Context, []save.EventRecord) error        { return nil }
func (prestigeNoopProjector) RepairFounder(context.Context, string, *save.State) error { return nil }

func TestPrestigeWindDownAndScriptedExitIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	replayBundle := epoch5TestBundle(t)
	economyBytes := replayBundle.Artifacts["economy"]
	routeBytes := replayBundle.Artifacts["routes"]
	prestigeBytes := replayBundle.Artifacts["prestige"]
	catalog, err := economy.LoadCatalog(economyBytes)
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(routeBytes)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := prestigecore.LoadPolicy(prestigeBytes)
	if err != nil {
		t.Fatal(err)
	}
	commonsCatalog, err := commons.LoadCatalog(replayBundle.Artifacts["commons"])
	if err != nil {
		t.Fatal(err)
	}
	factionCatalog, err := faction.LoadCatalog(replayBundle.Artifacts["factions"], faction.CompactTitheBand{
		MinimumPPM: commonsCatalog.MinimumTithePPM,
		DefaultPPM: commonsCatalog.DefaultTithePPM,
		MaximumPPM: commonsCatalog.MaximumTithePPM,
	})
	if err != nil {
		t.Fatal(err)
	}
	hash := replayBundle.ConstantsHash
	seedProductionEpoch(t, db, hash, replayBundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{hash: catalog}, routes: map[string]*routes.Catalog{hash: routeCatalog}, prestige: map[string]*prestigecore.Policy{hash: policy}, factions: map[string]*faction.Catalog{hash: factionCatalog}}
	commonsCatalogs := commons.CatalogSet{hash: commonsCatalog}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	projectionFailure := &failNextProjection{}
	service, err := NewService(store, resolver, nil, nil, nil, WithRouteCatalogs(resolver), WithRouteProjector(prestigeNoopProjector{}), WithCompactPolicies(commonsCatalogs), WithProgressionRuntime(resolver), WithCurrentConstantsHash(hash), WithCommonsWeightResolver(integrationWeight(1_000_000)), WithReplayCatalogs(ReplayCatalogSet{hash: replayBundle}), WithGuildSettlements(emptyGuildSettlements{}), WithEventProjector(projectionFailure))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 29, 16, 0, 0, 0, time.UTC)

	t.Run("elective wind down", func(t *testing.T) {
		owner := "01985555-1000-7000-8000-000000000001"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now.Add(-10*time.Minute), now, "0", decimal.New(8, 12), 1)
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, hash); err != nil {
			t.Fatal(err)
		}
		companyBefore, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		routePayload, _ := json.Marshal(map[string]any{"route_id": "route.nonprofit_wrapper_zip", "gate_id": "gate.t1_to_t2", "run_id": map[string]any{"company_stream_id": companyRevision.StreamID, "run_seq": 1}, "founder_id": owner})
		if _, err := db.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,1,1,'route_executed',$2,$3,$4,$5)`, companyRevision.StreamID, "01985555-1000-7000-8000-000000000099", hash, now.Add(-time.Minute), routePayload); err != nil {
			t.Fatal(err)
		}
		request := []byte(`{"intent_id":"01985555-1001-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`)
		projectionFailure.fail = true
		result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if !errors.Is(err, ErrInvalidEngineState) {
			t.Fatalf("post-commit Exit projection error=%v", err)
		}
		result, err = service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if err != nil || !result.Replay {
			t.Fatalf("result=%+v err=%v", result, err)
		}
		var receipt struct {
			Outcome         string `json:"outcome"`
			NewRevision     int64  `json:"new_revision"`
			FounderRevision int64  `json:"founder_revision"`
		}
		if err := json.Unmarshal(result.Receipt, &receipt); err != nil || receipt.Outcome != "applied" || receipt.NewRevision != 3 || receipt.FounderRevision != 2 {
			t.Fatalf("receipt=%s err=%v", result.Receipt, err)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || founder.State.ReputationLevel != 2 || founder.State.RouteKnowledgeBalance != 25 || len(founder.State.ExitHistory) != 1 || founder.State.ExitHistory[0].ExitType != "scripted_first" {
			t.Fatalf("founder=%+v err=%v", founder.State, err)
		}
		company, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.Revision.Number != 3 || company.State.RunSeq != 2 || company.State.Tier != 0 || !company.State.LifetimeValue.Eq(decimal.Zero) {
			t.Fatalf("company=%+v err=%v", company, err)
		}
		var endedPayload []byte
		if err := db.QueryRowContext(ctx, `SELECT payload FROM events WHERE stream_id=$1 AND kind='run_ended'`, companyRevision.StreamID).Scan(&endedPayload); err != nil {
			t.Fatal(err)
		}
		var ended struct {
			ExecutedRoutes []string `json:"executed_routes"`
			TerminalSeq    int64    `json:"terminal_seq"`
		}
		if err := json.Unmarshal(endedPayload, &ended); err != nil || len(ended.ExecutedRoutes) != 1 || ended.ExecutedRoutes[0] != "route.nonprofit_wrapper_zip" || ended.TerminalSeq != 1 {
			t.Fatalf("run_ended=%s err=%v", endedPayload, err)
		}
		var sequence int64
		var loggedPayload, replayInputs []byte
		if err := db.QueryRowContext(ctx, `SELECT seq,canonical_payload,replay_inputs FROM run_log WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&sequence, &loggedPayload, &replayInputs); err != nil || sequence != ended.TerminalSeq {
			t.Fatalf("run log sequence=%d terminal=%d payload=%s err=%v", sequence, ended.TerminalSeq, loggedPayload, err)
		}
		parsedInputs, err := parseReplayInputs(replayInputs)
		if err != nil {
			t.Fatal(err)
		}
		var terminal replayExitResolved
		if err := decodeReplayStrict(parsedInputs.Resolved, &terminal); err != nil || terminal.Kind != "exit" || terminal.SelectedExitType != "scripted_first" || terminal.NextConstantsHash != hash || len(terminal.ExecutedRouteIDs) != 1 {
			t.Fatalf("terminal replay inputs=%s err=%v", replayInputs, err)
		}
		var founderLogSeq, sourceRunSeq, sourceLogSeq int64
		var sourceStream, founderInputHash, founderKind string
		var founderAudit []byte
		if err := db.QueryRowContext(ctx, `SELECT seq,source_company_stream_id,source_run_seq,
			source_run_log_seq,constants_hash,replay_inputs->'resolved'->>'kind',receipt
			FROM founder_log WHERE founder_stream_id=$1 AND intent_id=$2`, founderRevision.StreamID,
			parsedInputs.Command.IntentID).Scan(&founderLogSeq, &sourceStream, &sourceRunSeq, &sourceLogSeq,
			&founderInputHash, &founderKind, &founderAudit); err != nil || founderLogSeq != 1 ||
			sourceStream != companyRevision.StreamID || sourceRunSeq != 1 || sourceLogSeq != sequence ||
			founderInputHash != hash || founderKind != founderExitResolvedKind {
			t.Fatalf("Founder Exit log seq=%d source=%s/%d/%d hash=%q kind=%q audit=%s err=%v",
				founderLogSeq, sourceStream, sourceRunSeq, sourceLogSeq, founderInputHash, founderKind,
				founderAudit, err)
		}
		var audit founderExitAuditReceipt
		if err := json.Unmarshal(founderAudit, &audit); err != nil || audit.Outcome != "applied" ||
			audit.FounderRevision == nil || *audit.FounderRevision != 2 ||
			audit.ResultConstantsHash != hash {
			t.Fatalf("Founder Exit audit=%+v raw=%s err=%v", audit, founderAudit, err)
		}
		founderGenesis, err := store.LoadFounderGenesis(ctx, founderRevision.StreamID)
		if err != nil || founderGenesis.Revision != 1 || founderGenesis.ConstantsHash != hash {
			t.Fatalf("Founder Exit genesis=%+v err=%v", founderGenesis, err)
		}
		founderHistory, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		if verdict := VerifyFounderHistory(founderHistory, ReplayCatalogSet{hash: replayBundle}); verdict != ReplayVerified {
			t.Fatalf("persisted Founder history verdict=%s history=%+v", verdict, founderHistory)
		}
		poisoned := founderHistory
		poisoned.Entries = append([]save.FounderHistoryEntry(nil), founderHistory.Entries...)
		poisoned.Entries[0].Events = []save.EventWrite{}
		if verdict := VerifyFounderHistory(poisoned, ReplayCatalogSet{hash: replayBundle}); verdict != ReplayStateDivergence {
			t.Fatalf("Founder event poison verdict=%s", verdict)
		}
		gap := founderHistory
		gap.Entries = append([]save.FounderHistoryEntry(nil), founderHistory.Entries...)
		gap.Entries[0].Sequence = 2
		if verdict := VerifyFounderHistory(gap, ReplayCatalogSet{hash: replayBundle}); verdict != ReplayLogGap {
			t.Fatalf("Founder sequence gap verdict=%s", verdict)
		}
		if verdict := VerifyFounderHistory(founderHistory, ReplayCatalogSet{}); verdict != ReplayConstantsMismatch {
			t.Fatalf("Founder missing-artifact verdict=%s", verdict)
		}
		terminalBundle := replayBundle
		terminalBundle.Next = &terminalBundle
		replayed, err := ApplyLoggedExit(companyBefore.State, loggedPayload, terminalBundle, replayInputs)
		if err != nil || string(replayed.Decision.Receipt) != string(result.Receipt) {
			t.Fatalf("terminal replay receipt=%s live=%s err=%v", replayed.Decision.Receipt, result.Receipt, err)
		}
		expectedEvents := append([]save.EventWrite(nil), replayed.Decision.FounderEvents...)
		expectedEvents = append(expectedEvents, replayed.Decision.CompanyEndedEvents...)
		expectedEvents = append(expectedEvents, replayed.Decision.CompanyStartedEvents...)
		rows, err := db.QueryContext(ctx, `SELECT kind,payload FROM events WHERE intent_id=$1 ORDER BY event_seq`, parsedInputs.Command.IntentID)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		index := 0
		for rows.Next() {
			var kind save.EventKind
			var payload []byte
			if err := rows.Scan(&kind, &payload); err != nil || index >= len(expectedEvents) || expectedEvents[index].Kind != kind || !jsonSemanticallyEqual(expectedEvents[index].Payload, payload) {
				t.Fatalf("terminal event[%d] replay=%v live=%s/%s err=%v", index, expectedEvents, kind, payload, err)
			}
			index++
		}
		if err := rows.Err(); err != nil || index != len(expectedEvents) {
			t.Fatalf("terminal replay events=%d live=%d err=%v", len(expectedEvents), index, err)
		}
		replay, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if err != nil || !replay.Replay || string(replay.Receipt) != string(result.Receipt) {
			t.Fatalf("replay=%+v err=%v", replay, err)
		}
	})

	t.Run("incorporated run exits and may incorporate again", func(t *testing.T) {
		owner := "01985555-1500-7000-8000-000000000001"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now.Add(-10*time.Minute), now, "0", decimal.New(8, 12), 2)
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, hash); err != nil {
			t.Fatal(err)
		}
		incorporate := []byte(`{"intent_id":"01985555-1501-7000-8000-000000000001","kind":"incorporate","expected_revision":1,"faction_id":"open_source"}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, incorporate); err != nil || result.Replay {
			t.Fatalf("incorporate=%+v err=%v", result, err)
		}
		exit := []byte(`{"intent_id":"01985555-1502-7000-8000-000000000001","kind":"wind_down","expected_revision":2,"expected_founder_revision":1}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(time.Minute), exit); err != nil || result.Replay {
			t.Fatalf("exit=%+v err=%v", result, err)
		}
		var endedFaction string
		if err := db.QueryRowContext(ctx, `SELECT payload->>'faction' FROM events WHERE stream_id=$1 AND kind='run_ended'`, companyRevision.StreamID).Scan(&endedFaction); err != nil || endedFaction != "open_source" {
			t.Fatalf("ended faction=%v err=%v", endedFaction, err)
		}
		company, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.Revision.Number != 4 || company.State.RunSeq != 2 || company.State.FactionID != "" {
			t.Fatalf("new run company=%+v err=%v", company, err)
		}
		advanced, err := store.ApplyIntent(ctx, companyRevision.StreamID, company.Revision.Number, "01985555-1503-7000-8000-000000000001", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", func(state *save.State, revision save.Revision) (save.IntentDecision, error) {
			state.Tier = 2
			return save.IntentDecision{Outcome: save.IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied","new_revision":5}`)}, nil
		})
		if err != nil || advanced.Outcome != save.IntentApplied {
			t.Fatalf("test progression=%+v err=%v", advanced, err)
		}
		reincorporate := []byte(`{"intent_id":"01985555-1504-7000-8000-000000000001","kind":"incorporate","expected_revision":5,"faction_id":"enterprise"}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(2*time.Minute), reincorporate); err != nil || result.Replay {
			t.Fatalf("reincorporate=%+v err=%v", result, err)
		}
		company, err = store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.State.FactionID != "enterprise" {
			t.Fatalf("reincorporated company=%+v err=%v", company, err)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || len(founder.State.ExitHistory) != 1 {
			t.Fatalf("founder=%+v err=%v", founder, err)
		}
	})

	t.Run("scripted first threshold", func(t *testing.T) {
		owner := "01985555-2000-7000-8000-000000000002"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now.Add(-16*time.Minute), now.Add(-time.Second), "1e10", decimal.New(1, 12), 2)
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, hash); err != nil {
			t.Fatal(err)
		}
		request := []byte(`{"intent_id":"01985555-2001-7000-8000-000000000002","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`)
		result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if err != nil || result.Replay {
			t.Fatalf("result=%+v err=%v receipt=%s", result, err, result.Receipt)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || len(founder.State.ExitHistory) != 1 || founder.State.ExitHistory[0].ExitType != "scripted_first" || founder.State.ReputationLevel != 1 {
			t.Fatalf("founder=%+v err=%v", founder.State, err)
		}
		var kinds []string
		rows, err := db.QueryContext(ctx, `SELECT kind FROM events WHERE stream_id=$1 ORDER BY revision,kind`, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var kind string
			_ = rows.Scan(&kind)
			kinds = append(kinds, kind)
		}
		rows.Close()
		if len(kinds) != 3 || kinds[0] != "gate_crossed" || kinds[1] != "run_ended" || kinds[2] != "run_started" {
			t.Fatalf("company event kinds=%v", kinds)
		}
	})

	t.Run("offer preview remains a promise", func(t *testing.T) {
		owner := "01985555-3000-7000-8000-000000000003"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now, now, "1e25", decimal.New(8, 12), 7)
		founderLoaded, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		founderLoaded.State.ExitHistory = append(founderLoaded.State.ExitHistory, save.ExitRecord{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour), ReputationDelta: 1})
		founderRevision, err = store.Write(ctx, founderRevision.StreamID, founderRevision.Number, hash, founderLoaded.State, save.WriteContext{Cause: "prestige.test.prior_exit"})
		if err != nil {
			t.Fatal(err)
		}
		companyLoaded, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		companyLoaded.State.RunSeq = 2
		companyRevision, err = store.Write(ctx, companyRevision.StreamID, companyRevision.Number, hash, companyLoaded.State, save.WriteContext{Cause: "prestige.test.second_run"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 2, hash); err != nil {
			t.Fatal(err)
		}
		cross := []byte(`{"intent_id":"01985555-3001-7000-8000-000000000003","kind":"cross_gate","expected_revision":2,"gate_id":"gate.t7_to_t8","route_id":null}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, cross); err != nil || result.Replay {
			t.Fatalf("cross=%+v err=%v", result, err)
		}
		offered, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || offered.State.OfferState == nil {
			t.Fatalf("offered=%+v err=%v", offered.State, err)
		}
		stored, err := prestigecore.DecodeStoredOfferTerms(offered.State.OfferState.TermsJSON)
		if err != nil {
			t.Fatal(err)
		}
		offerID := offered.State.OfferState.OfferID
		offered.State.LifetimeValue = decimal.New(27, 12)
		written, err := store.Write(ctx, companyRevision.StreamID, 3, hash, offered.State, save.WriteContext{Cause: "prestige.test.progress"})
		if err != nil || written.Number != 4 {
			t.Fatalf("write=%+v err=%v", written, err)
		}
		acceptBody, _ := json.Marshal(map[string]any{"intent_id": "01985555-3002-7000-8000-000000000003", "kind": "accept_exit_offer", "expected_revision": 4, "expected_founder_revision": founderRevision.Number, "offer_id": offerID})
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, acceptBody); err != nil || result.Replay {
			t.Fatalf("accept=%+v err=%v", result, err)
		}
		rows, err := db.QueryContext(ctx, `SELECT kind,payload::text FROM events WHERE stream_id=$1 AND intent_id=$2 ORDER BY event_seq`, companyRevision.StreamID, "01985555-3002-7000-8000-000000000003")
		if err != nil {
			t.Fatal(err)
		}
		var acceptedKinds []string
		var acceptedPayloads []string
		for rows.Next() {
			var kind, payload string
			if err := rows.Scan(&kind, &payload); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			acceptedKinds = append(acceptedKinds, kind)
			acceptedPayloads = append(acceptedPayloads, payload)
		}
		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}
		resolvedIndex, endedIndex := -1, -1
		for index, kind := range acceptedKinds {
			if kind == "exit_offer_resolved" {
				resolvedIndex = index
				var payload map[string]string
				if err := json.Unmarshal([]byte(acceptedPayloads[index]), &payload); err != nil || payload["offer_id"] != offerID || payload["resolution"] != "accepted" || len(payload) != 2 {
					t.Fatalf("accepted resolution payload=%s err=%v", acceptedPayloads[index], err)
				}
			}
			if kind == "run_ended" {
				endedIndex = index
			}
		}
		if resolvedIndex < 0 || endedIndex < 0 || resolvedIndex >= endedIndex {
			t.Fatalf("accepted event order=%v", acceptedKinds)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || founder.State.ReputationLevel < stored.PayoutPreview.ReputationDelta {
			t.Fatalf("founder reputation=%d preview=%d err=%v", founder.State.ReputationLevel, stored.PayoutPreview.ReputationDelta, err)
		}
	})

	t.Run("offer promise holds across the declared age population", func(t *testing.T) {
		ages := []struct {
			name    string
			ageMS   int64
			expired bool
		}{
			{name: "fresh", ageMS: 0},
			{name: "one_millisecond", ageMS: 1},
			{name: "half_life", ageMS: policy.OfferDurationMS / 2},
			{name: "last_live_millisecond", ageMS: policy.OfferDurationMS - 1},
			{name: "exact_expiry", ageMS: policy.OfferDurationMS, expired: true},
		}
		for index, row := range ages {
			t.Run(row.name, func(t *testing.T) {
				owner := fmt.Sprintf("01985555-31%02d-7000-8000-%012d", index, index+1)
				spawnAt := now.Add(-time.Duration(row.ageMS) * time.Millisecond)
				founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash,
					owner, spawnAt, spawnAt, spawnAt, "1e25", decimal.New(8, 12), 7)
				founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
				if err != nil {
					t.Fatal(err)
				}
				founder.State.ExitHistory = append(founder.State.ExitHistory, save.ExitRecord{
					RunID: 1, ExitType: "collapse", OccurredAt: spawnAt.Add(-time.Hour), ReputationDelta: 1,
				})
				founderRevision, err = store.Write(ctx, founderRevision.StreamID, founderRevision.Number, hash,
					founder.State, save.WriteContext{Cause: "prestige.ac2.prior_exit"})
				if err != nil {
					t.Fatal(err)
				}
				company, err := store.LoadLatest(ctx, companyRevision.StreamID)
				if err != nil {
					t.Fatal(err)
				}
				company.State.RunSeq = 2
				companyRevision, err = store.Write(ctx, companyRevision.StreamID, companyRevision.Number, hash,
					company.State, save.WriteContext{Cause: "prestige.ac2.second_run"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 2, hash); err != nil {
					t.Fatal(err)
				}
				crossID := fmt.Sprintf("01985555-32%02d-7000-8000-%012d", index, index+1)
				cross := []byte(fmt.Sprintf(`{"intent_id":%q,"kind":"cross_gate","expected_revision":2,"gate_id":"gate.t7_to_t8","route_id":null}`, crossID))
				if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, spawnAt, cross); err != nil || result.Replay {
					t.Fatalf("cross=%+v err=%v", result, err)
				}
				offered, err := store.LoadLatest(ctx, companyRevision.StreamID)
				if err != nil || offered.State.OfferState == nil {
					t.Fatalf("offered=%+v err=%v", offered.State, err)
				}
				stored, err := prestigecore.DecodeStoredOfferTerms(offered.State.OfferState.TermsJSON)
				if err != nil {
					t.Fatal(err)
				}
				offerID := offered.State.OfferState.OfferID
				offered.State.LifetimeValue = decimal.New(27, 12)
				progressed, err := store.Write(ctx, companyRevision.StreamID, offered.Revision.Number, hash,
					offered.State, save.WriteContext{Cause: "prestige.ac2.progress"})
				if err != nil || progressed.Number != 4 {
					t.Fatalf("progressed=%+v err=%v", progressed, err)
				}
				acceptID := fmt.Sprintf("01985555-33%02d-7000-8000-%012d", index, index+1)
				accept, _ := json.Marshal(map[string]any{"intent_id": acceptID, "kind": "accept_exit_offer",
					"expected_revision": 4, "expected_founder_revision": founderRevision.Number, "offer_id": offerID})
				result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, accept)
				if err != nil || result.Replay {
					t.Fatalf("accept=%+v err=%v", result, err)
				}
				loadedFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
				if err != nil {
					t.Fatal(err)
				}
				if row.expired {
					var receipt struct {
						Outcome   string `json:"outcome"`
						Rejection struct {
							Category string `json:"category"`
						} `json:"rejection"`
					}
					if err := json.Unmarshal(result.Receipt, &receipt); err != nil || receipt.Outcome != "rejected" ||
						receipt.Rejection.Category != "offer_expired" || loadedFounder.Revision.Number != founderRevision.Number {
						t.Fatalf("expired receipt=%s founder_revision=%d want=%d err=%v",
							result.Receipt, loadedFounder.Revision.Number, founderRevision.Number, err)
					}
					return
				}
				var payoutBytes []byte
				if err := db.QueryRowContext(ctx, `SELECT payload->'payout' FROM events
					WHERE stream_id=$1 AND intent_id=$2 AND kind='run_ended'`, companyRevision.StreamID, acceptID).Scan(&payoutBytes); err != nil {
					t.Fatal(err)
				}
				var payout prestigecore.Terms
				if err := json.Unmarshal(payoutBytes, &payout); err != nil {
					t.Fatal(err)
				}
				if !termsDominate(payout, stored.PayoutPreview) {
					t.Fatalf("age_ms=%d payout=%+v preview=%+v", row.ageMS, payout, stored.PayoutPreview)
				}
			})
		}
		preview := prestigecore.Terms{ReputationDelta: 5, RouteKnowledge: 3,
			NetworkSlotUnlocks: []save.NetworkSlot{{Slot: "network.alpha", CarriedRef: "upgrade.alpha"}}}
		forged := preview
		forged.ReputationDelta--
		if termsDominate(forged, preview) {
			t.Fatal("one-below-preview payout passed the AC2 oracle")
		}
	})

	t.Run("non-empty ledger facts survive different repeated exit kinds", func(t *testing.T) {
		owner := "01985555-3600-7000-8000-000000000003"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner,
			now, now, now, "1e25", decimal.New(8, 12), 7)
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		founder.State.ExitHistory = append(founder.State.ExitHistory, save.ExitRecord{
			RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour), ReputationDelta: 1,
		})
		founder.State.LedgerFactKinds["fact.career"] = true
		founderRevision, err = store.Write(ctx, founderRevision.StreamID, founderRevision.Number, hash,
			founder.State, save.WriteContext{Cause: "prestige.ac3.prior_facts"})
		if err != nil {
			t.Fatal(err)
		}
		company, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		company.State.RunSeq = 2
		company.State.LedgerFactKinds["fact.run_two"] = true
		companyRevision, err = store.Write(ctx, companyRevision.StreamID, companyRevision.Number, hash,
			company.State, save.WriteContext{Cause: "prestige.ac3.first_company_facts"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 2, hash); err != nil {
			t.Fatal(err)
		}
		cross := []byte(`{"intent_id":"01985555-3601-7000-8000-000000000003","kind":"cross_gate","expected_revision":2,"gate_id":"gate.t7_to_t8","route_id":null}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, cross); err != nil || result.Replay {
			t.Fatalf("cross=%+v err=%v", result, err)
		}
		offered, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || offered.State.OfferState == nil {
			t.Fatalf("offered=%+v err=%v", offered.State, err)
		}
		accept, _ := json.Marshal(map[string]any{"intent_id": "01985555-3602-7000-8000-000000000003",
			"kind": "accept_exit_offer", "expected_revision": offered.Revision.Number,
			"expected_founder_revision": founderRevision.Number, "offer_id": offered.State.OfferState.OfferID})
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, accept); err != nil || result.Replay {
			t.Fatalf("accept=%+v err=%v", result, err)
		}
		founder, err = store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || !exactFactSet(founder.State.LedgerFactKinds, "fact.career", "fact.run_two") {
			t.Fatalf("facts after offer Exit=%v err=%v", founder.State.LedgerFactKinds, err)
		}
		company, err = store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.State.RunSeq != 3 {
			t.Fatalf("company after offer Exit=%+v err=%v", company.State, err)
		}
		company.State.LedgerFactKinds["fact.run_three"] = true
		company.State.Tier = 1
		companyRevision, err = store.Write(ctx, companyRevision.StreamID, company.Revision.Number, hash,
			company.State, save.WriteContext{Cause: "prestige.ac3.second_company_facts"})
		if err != nil {
			t.Fatal(err)
		}
		windDown := []byte(fmt.Sprintf(`{"intent_id":"01985555-3603-7000-8000-000000000003","kind":"wind_down","expected_revision":%d,"expected_founder_revision":%d}`,
			companyRevision.Number, founder.Revision.Number))
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, windDown); err != nil || result.Replay {
			t.Fatalf("pending-event wind down=%+v err=%v", result, err)
		}
		founder, err = store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || !exactFactSet(founder.State.LedgerFactKinds, "fact.career", "fact.run_two", "fact.run_three") ||
			len(founder.State.ExitHistory) != 3 || founder.State.ExitHistory[1].ExitType == founder.State.ExitHistory[2].ExitType {
			t.Fatalf("facts/history after repeated Exits facts=%v history=%+v err=%v",
				founder.State.LedgerFactKinds, founder.State.ExitHistory, err)
		}
		forged := map[string]bool{"fact.career": true, "fact.run_two": true}
		if exactFactSet(forged, "fact.career", "fact.run_two", "fact.run_three") {
			t.Fatal("missing second-run fact passed the cumulative ledger oracle")
		}
	})

	t.Run("offers wait for scripted first exit", func(t *testing.T) {
		owner := "01985555-3500-7000-8000-000000000003"
		_, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now, now, "1e25", decimal.New(8, 12), 7)
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, hash); err != nil {
			t.Fatal(err)
		}
		cross := []byte(`{"intent_id":"01985555-3501-7000-8000-000000000003","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t7_to_t8","route_id":null}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, cross); err != nil || result.Replay {
			t.Fatalf("cross=%+v err=%v", result, err)
		}
		company, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.State.OfferState != nil {
			t.Fatalf("company=%+v err=%v", company.State, err)
		}
		var offers int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND kind='exit_offer_spawned'`, companyRevision.StreamID).Scan(&offers); err != nil || offers != 0 {
			t.Fatalf("offers=%d err=%v", offers, err)
		}
	})

	ownerAcrossEpoch := "01985555-4000-7000-8000-000000000004"
	founderAcrossEpoch, companyAcrossEpoch := createPrestigeStreams(t, ctx, store, catalog, hash, ownerAcrossEpoch, now, now.Add(-10*time.Minute), now, "0", decimal.New(8, 12), 1)
	accountAcrossEpoch := "01985555-4000-7000-8000-000000000099"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'leaderboard-ac3') ON CONFLICT DO NOTHING`, accountAcrossEpoch); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, accountAcrossEpoch, ownerAcrossEpoch); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PinRunToCurrentEpoch(ctx, companyAcrossEpoch.StreamID, ownerAcrossEpoch, 1, hash); err != nil {
		t.Fatal(err)
	}
	changedEconomyBytes := append(append([]byte(nil), economyBytes...), '\n')
	changedArtifacts := make(map[string][]byte, len(replayBundle.Artifacts))
	for name, data := range replayBundle.Artifacts {
		changedArtifacts[name] = append([]byte(nil), data...)
	}
	changedArtifacts["economy"] = changedEconomyBytes
	currentHash, err := save.ConstantsHashArtifacts(changedArtifacts)
	if err != nil || currentHash == hash {
		t.Fatalf("changed hash=%s old=%s err=%v", currentHash, hash, err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "changelog"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "changelog", "epoch-2.md"), []byte("# changed balance bytes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	epochRepository, err := leaderboard.NewRepository(db, root)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(changedArtifacts))
	for name := range changedArtifacts {
		names = append(names, name)
	}
	sort.Strings(names)
	artifacts := make([]leaderboard.Artifact, 0, len(changedArtifacts))
	for _, name := range names {
		artifacts = append(artifacts, leaderboard.Artifact{Name: name, Bytes: changedArtifacts[name]})
	}
	var firstEpochStarted time.Time
	if err := db.QueryRowContext(ctx, `SELECT started_at FROM epochs WHERE ended_at IS NULL`).Scan(&firstEpochStarted); err != nil {
		t.Fatal(err)
	}
	secondEpoch, err := epochRepository.MintEpoch(ctx, "Phase 0.1", firstEpochStarted.Add(time.Hour), "changelog/epoch-2.md", artifacts)
	if err != nil || secondEpoch.ID != 2 || secondEpoch.Hashes[0] != currentHash {
		t.Fatalf("second epoch=%+v err=%v", secondEpoch, err)
	}
	resolver.economy[currentHash], resolver.routes[currentHash], resolver.prestige[currentHash], resolver.factions[currentHash] = catalog, routeCatalog, policy, factionCatalog
	commonsCatalogs[currentHash] = commonsCatalog
	currentReplayBundle := loadReplayTestBundle(t, currentHash, changedArtifacts)
	currentService, err := NewService(store, resolver, nil, nil, nil, WithRouteCatalogs(resolver), WithRouteProjector(prestigeNoopProjector{}), WithCompactPolicies(commonsCatalogs), WithProgressionRuntime(resolver), WithCurrentConstantsHash(currentHash), WithCommonsWeightResolver(integrationWeight(1_000_000)), WithReplayCatalogs(ReplayCatalogSet{hash: replayBundle, currentHash: currentReplayBundle}), WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("epoch mint moves only the new run to changed bytes", func(t *testing.T) {
		request := []byte(`{"intent_id":"01985555-4001-7000-8000-000000000004","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`)
		if result, err := currentService.Handle(ctx, companyAcrossEpoch.StreamID, ModeOnline, now.Add(time.Hour), request); err != nil || result.Replay {
			t.Fatalf("epoch-crossing exit=%+v err=%v", result, err)
		}
		company, err := store.LoadLatest(ctx, companyAcrossEpoch.StreamID)
		if err != nil || company.Revision.ConstantsHash != currentHash || company.State.RunSeq != 2 {
			t.Fatalf("company revision=%+v state=%+v err=%v", company.Revision, company.State, err)
		}
		founder, err := store.LoadLatest(ctx, founderAcrossEpoch.StreamID)
		if err != nil || founder.Revision.ConstantsHash != currentHash {
			t.Fatalf("founder revision=%+v err=%v", founder.Revision, err)
		}
		var oldEpoch, newEpoch int64
		var oldHash, newHash string
		if err := db.QueryRowContext(ctx, `SELECT epoch_id,constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=1`, companyAcrossEpoch.StreamID).Scan(&oldEpoch, &oldHash); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRowContext(ctx, `SELECT epoch_id,constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=2`, companyAcrossEpoch.StreamID).Scan(&newEpoch, &newHash); err != nil || oldEpoch != 1 || oldHash != hash || newEpoch != 2 || newHash != currentHash {
			t.Fatalf("pins old=(%d,%s) new=(%d,%s) err=%v", oldEpoch, oldHash, newEpoch, newHash, err)
		}
		rows, err := db.QueryContext(ctx, `SELECT kind,constants_hash FROM events WHERE stream_id=$1 AND kind IN ('run_ended','run_started') ORDER BY revision`, companyAcrossEpoch.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var pairs [][2]string
		for rows.Next() {
			var pair [2]string
			if err := rows.Scan(&pair[0], &pair[1]); err != nil {
				t.Fatal(err)
			}
			pairs = append(pairs, pair)
		}
		if len(pairs) != 2 || pairs[0] != [2]string{"run_ended", hash} || pairs[1] != [2]string{"run_started", currentHash} {
			t.Fatalf("transition events=%v", pairs)
		}

		genesis, version, entries := persistedPrestigeReplay(t, db, companyAcrossEpoch.StreamID, founderAcrossEpoch.StreamID, 1, &currentReplayBundle)
		if verdict := VerifyReplayRun(genesis, version, replayBundle, entries, hash, false); verdict != ReplayVerified {
			t.Fatalf("epoch-crossing replay verdict=%s", verdict)
		}
		if verdict := VerifyReplayRun(genesis, version, replayBundle, entries, currentHash, false); verdict != ReplayConstantsMismatch {
			t.Fatalf("current-hash substitution verdict=%s", verdict)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		if err := leaderboard.NewQueueProjector().ProjectVerifiedRun(ctx, tx, companyAcrossEpoch.StreamID, 1); err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
		oldBoard, err := epochRepository.TimeBoard(ctx, "any_percent", leaderboard.Variables{}, 1, 0, 10, nil)
		if err != nil || len(oldBoard) != 1 || oldBoard[0].RunID != companyAcrossEpoch.StreamID+":1" {
			t.Fatalf("epoch-1 board=%+v err=%v", oldBoard, err)
		}
		newBoard, err := epochRepository.TimeBoard(ctx, "any_percent", leaderboard.Variables{}, 2, 0, 10, nil)
		if err != nil || len(newBoard) != 0 {
			t.Fatalf("epoch-crossing run leaked into epoch 2 board=%+v err=%v", newBoard, err)
		}
	})

	t.Run("company v6 migrates to an exitable pre-timer run", func(t *testing.T) {
		owner := "01985555-5000-7000-8000-000000000005"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, currentHash, owner, now, now, now, "1e10", decimal.Zero, 0)
		loaded, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := save.EncodeState(loaded.State)
		if err != nil {
			t.Fatal(err)
		}
		var legacy map[string]any
		if err := json.Unmarshal(encoded, &legacy); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"tier", "lifetime_value", "offer_state", "run_started_at_ms", "run_pre_timer", "offline_spans", "collapsed_offline_ms", "reputation_level", "reputation_unlock_ppm", "network_slots", "clout_lifetime", "soul", "age_ms", "notoriety", "advisor_mode", "exit_history", "faction_id", "incorporated_at_ms", "stock_units", "stock_progress_ms", "consumed_stock_units", "guild_tithe_carry_ppm", "guild_boundary_seq", "guild_consumed_window_units", "guild_boundary_guild_id", "generators_purchased_total", "upgrades_owned", "generators_provisioned", "provision_remainders_ppm", "stock_rate_remainder_ppm"} {
			delete(legacy, key)
		}
		legacyBytes, _ := json.Marshal(legacy)
		if _, err := db.ExecContext(ctx, `UPDATE save_revisions SET version=6,state=$2 WHERE stream_id=$1 AND revision=1`, companyRevision.StreamID, legacyBytes); err != nil {
			t.Fatal(err)
		}
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, currentHash); err != nil {
			t.Fatal(err)
		}
		request := []byte(`{"intent_id":"01985555-5001-7000-8000-000000000005","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`)
		result, err := currentService.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(16*time.Minute), request)
		if err != nil || result.Replay {
			t.Fatalf("v6 cross=%+v err=%v", result, err)
		}
		exit := []byte(`{"intent_id":"01985555-5002-7000-8000-000000000005","kind":"wind_down","expected_revision":2,"expected_founder_revision":1}`)
		result, err = currentService.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(16*time.Minute), exit)
		if err != nil || result.Replay {
			t.Fatalf("v6 exit=%+v err=%v", result, err)
		}
		company, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.State.RunSeq != 2 || company.State.RunPreTimer {
			t.Fatalf("new company=%+v err=%v", company.State, err)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || len(founder.State.ExitHistory) != 1 || founder.State.ExitHistory[0].ExitType != "scripted_first" {
			t.Fatalf("founder=%+v err=%v", founder.State, err)
		}
		var preTimer bool
		if err := db.QueryRowContext(ctx, `SELECT (payload->>'pre_timer')::boolean FROM events WHERE stream_id=$1 AND kind='run_ended'`, companyRevision.StreamID).Scan(&preTimer); err != nil || !preTimer {
			t.Fatalf("pre-timer event=%v err=%v", preTimer, err)
		}
	})

	t.Run("Founder attendance sample remains monotonic across both Exit orders", func(t *testing.T) {
		for index, validateBeforeExit := range []bool{true, false} {
			name := "exit_wins"
			if validateBeforeExit {
				name = "founder_command_wins"
			}
			t.Run(name, func(t *testing.T) {
				owner := []string{"01985555-7000-7000-8000-000000000001", "01985555-7000-7000-8000-000000000002"}[index]
				founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, currentHash, owner, now, now.Add(-10*time.Minute), now, "0", decimal.New(8, 12), 1)
				if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, currentHash); err != nil {
					t.Fatal(err)
				}
				sample, err := currentService.ResolveFounderAttendance(ctx, founderRevision.StreamID, companyRevision.StreamID, now)
				if err != nil || sample.CurrentRunPartialAttendedMS != 600_000 || sample.CompletedAttendedMS != 0 {
					t.Fatalf("initial sample=%+v err=%v", sample, err)
				}
				if validateBeforeExit {
					founder, loadErr := store.LoadLatest(ctx, founderRevision.StreamID)
					if loadErr != nil || ValidateFounderAttendanceSample(founder.State, founder.Revision.Number, founderRevision.Number, sample) != nil {
						t.Fatalf("command-first validation founder=%+v load=%v", founder, loadErr)
					}
				}
				intentID := []string{"01985555-7001-7000-8000-000000000001", "01985555-7001-7000-8000-000000000002"}[index]
				request := []byte(fmt.Sprintf(`{"intent_id":"%s","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`, intentID))
				if _, err := currentService.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request); err != nil {
					t.Fatal(err)
				}
				founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
				if err != nil || founder.State.AgeMS != sample.EffectiveFounderAttendedMS {
					t.Fatalf("post-Exit founder=%+v sample=%+v err=%v", founder.State, sample, err)
				}
				if !validateBeforeExit && !errors.Is(ValidateFounderAttendanceSample(founder.State, founder.Revision.Number, founderRevision.Number, sample), ErrFounderAttendanceStale) {
					t.Fatal("Exit-winning schedule accepted the stale pre-Exit sample")
				}
				next, err := currentService.ResolveFounderAttendance(ctx, founderRevision.StreamID, companyRevision.StreamID, now)
				if err != nil || next.CompletedAttendedMS != sample.EffectiveFounderAttendedMS || next.CurrentRunPartialAttendedMS != 0 ||
					next.EffectiveFounderAttendedMS != sample.EffectiveFounderAttendedMS || next.RunSeq != 2 {
					t.Fatalf("post-Exit sample=%+v err=%v", next, err)
				}
			})
		}
	})
}

func TestPrestigeWindDownEligibleStateMatrixIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	bundle := activeContentBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{
		economy:  map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes:   map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes},
		prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction},
	}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	minigames, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	commonsCatalogs := commons.CatalogSet{bundle.ConstantsHash: bundle.Commons.(commonsbinding.ReplayPolicy).Catalog}
	service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil,
		WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}),
		WithGuildSettlements(emptyGuildSettlements{}), WithMinigameActivity(minigames),
		WithCompactPolicies(commonsCatalogs), WithCommonsWeightResolver(integrationWeight(1_000_000)))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 21, 17, 0, 0, 0, time.UTC)
	type matrixRow struct {
		name     string
		tier     int64
		eligible bool
		prepare  func(string, *save.State, *save.State)
	}
	rows := []matrixRow{
		{name: "plain", tier: 1, eligible: true},
		{name: "pending_opportunity", tier: 1, eligible: true, prepare: func(owner string, _ *save.State, company *save.State) {
			preparePendingOpportunityForExit(t, bundle, owner, company, now)
		}},
		{name: "active_buff", tier: 1, eligible: true, prepare: func(owner string, _ *save.State, company *save.State) {
			preparePendingOpportunityForExit(t, bundle, owner, company, now)
			attended := int64((20 * time.Minute) / time.Millisecond)
			company.ActiveBuffs = append(company.ActiveBuffs, save.ActiveBuff{
				BuffInstanceID: bundle.Opportunities.BuffID(owner, company.RunSeq, 0, attended),
				EffectRowID:    "active.production", ActivatedAttendedMS: attended - 1_000,
				ExpiresAttendedMS: attended + 5_000,
			})
		}},
		{name: "incorporated", tier: 1, eligible: true, prepare: func(_ string, _ *save.State, company *save.State) {
			company.FactionID = "enterprise"
			company.IncorporatedAt = now.Add(-time.Minute)
			company.StockUnits = 10
		}},
		{name: "live_offer", tier: 1, eligible: true, prepare: func(owner string, founder, company *save.State) {
			terms, termsErr := prestigecore.ComputeTerms(company, founder, bundle.Prestige, "acquisition")
			if termsErr != nil {
				t.Fatal(termsErr)
			}
			termsJSON, marshalErr := json.Marshal(prestigecore.StoredOfferTerms{
				PayoutPreview: terms, MarketModifierPPM: 1_000_000,
			})
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			company.OfferState = &save.ExitOfferState{OfferID: prestigecore.OfferID(owner, company.RunSeq, company.Tier, 0, now),
				ExitType: "acquisition", TermsJSON: termsJSON, SpawnedAt: now, ExpiresAt: now.Add(time.Minute)}
		}},
		{name: "tier_zero_control", tier: 0, eligible: false},
	}
	for index, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			owner := fmt.Sprintf("01985555-8%03d-7000-8000-%012d", index, index+1)
			founder := initialPrestigeWitnessState(t, bundle.Economy, economy.ScopeFounder, now)
			company := initialPrestigeWitnessState(t, bundle.Economy, economy.ScopeCompany, now)
			initializer := FounderInitializer{Catalogs: ReplayCatalogSet{bundle.ConstantsHash: bundle}}
			frozen, err := initializer.InitializeNewFounder(bundle.ConstantsHash, owner, now, founder, company)
			if err != nil {
				t.Fatal(err)
			}
			founder.ExitHistory = append(founder.ExitHistory, save.ExitRecord{
				RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour), ReputationDelta: 1,
			})
			company.RunSeq = 2
			company.Tier = row.tier
			company.RunStartedAt = now
			company.EvaluatedThrough = now
			company.ManualTokenRefilledAt = now
			if row.prepare != nil {
				row.prepare(owner, founder, company)
			}
			founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder,
				OwnerID: owner, Scope: economy.ScopeFounder}, bundle.ConstantsHash, founder,
				save.WriteContext{Cause: "prestige.ac5.matrix"})
			if err != nil {
				t.Fatal(err)
			}
			companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder,
				OwnerID: owner, Scope: economy.ScopeCompany}, bundle.ConstantsHash, company,
				save.WriteContext{Cause: "prestige.ac5.matrix"})
			if err != nil {
				t.Fatal(err)
			}
			pinTx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			genesis, err := save.EncodeState(company)
			if err == nil {
				_, err = save.PinRunWithGenesisTx(ctx, pinTx, companyRevision.StreamID, owner, 2,
					bundle.ConstantsHash, companyRevision.Version, genesis)
			}
			if err == nil {
				err = save.InsertRunFrozenContributionsTx(ctx, pinTx, companyRevision.StreamID, 2, frozen)
			}
			if err == nil {
				err = pinTx.Commit()
			} else {
				_ = pinTx.Rollback()
			}
			if err != nil {
				t.Fatal(err)
			}
			intentID := fmt.Sprintf("01985555-9%03d-7000-8000-%012d", index, index+1)
			request := []byte(fmt.Sprintf(`{"intent_id":%q,"kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`, intentID))
			result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
			if err != nil || result.Replay {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			loadedFounder, founderErr := store.LoadLatest(ctx, founderRevision.StreamID)
			loadedCompany, companyErr := store.LoadLatest(ctx, companyRevision.StreamID)
			if founderErr != nil || companyErr != nil {
				t.Fatalf("founder_err=%v company_err=%v", founderErr, companyErr)
			}
			var receipt struct {
				Outcome   string `json:"outcome"`
				Rejection struct {
					Category string `json:"category"`
					Detail   string `json:"detail"`
				} `json:"rejection"`
			}
			if err := json.Unmarshal(result.Receipt, &receipt); err != nil {
				t.Fatal(err)
			}
			if row.eligible {
				if receipt.Outcome != "applied" || loadedFounder.Revision.Number != 2 ||
					loadedCompany.Revision.Number != 3 || loadedCompany.State.RunSeq != 3 {
					t.Fatalf("eligible row receipt=%s founder=%+v company=%+v",
						result.Receipt, loadedFounder, loadedCompany)
				}
				return
			}
			if receipt.Outcome != "rejected" || receipt.Rejection.Category != "not_eligible" ||
				receipt.Rejection.Detail != "tier" || loadedFounder.Revision.Number != 1 ||
				loadedCompany.Revision.Number != 1 || len(loadedFounder.State.ExitHistory) != 1 {
				t.Fatalf("ineligible control receipt=%s founder=%+v company=%+v",
					result.Receipt, loadedFounder, loadedCompany)
			}
		})
	}
}

func createPrestigeStreams(t *testing.T, ctx context.Context, store *save.Store, catalog *economy.Catalog, hash, owner string, now, started, evaluated time.Time, cash string, lifetime decimal.Decimal, tier int64) (save.Revision, save.Revision) {
	t.Helper()
	founderLedger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	founder := &save.State{Ledger: founderLedger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: evaluated, ManualTokenRefilledAt: evaluated,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.Zero, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	companyLedger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": cash})
	if err != nil {
		t.Fatal(err)
	}
	company := &save.State{Ledger: companyLedger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0}, EvaluatedThrough: evaluated,
		ManualTokenMilli: catalog.ManualPolicy().BucketCapMilli, ManualTokenRefilledAt: evaluated,
		GatesCrossed: map[string]bool{}, RunSeq: 1, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []save.CompactSample{}, Tier: tier, LifetimeValue: lifetime, RunStartedAt: started, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: owner, Scope: economy.ScopeFounder}, hash, founder, save.WriteContext{Cause: "prestige.integration"})
	if err != nil {
		t.Fatal(err)
	}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: owner, Scope: economy.ScopeCompany}, hash, company, save.WriteContext{Cause: "prestige.integration"})
	if err != nil {
		t.Fatal(err)
	}
	return founderRevision, companyRevision
}

func persistedPrestigeReplay(t *testing.T, db *sql.DB, companyStreamID, founderStreamID string, runSeq int64, nextCatalog *CatalogBundle) ([]byte, int, []ReplayLogEntry) {
	t.Helper()
	ctx := context.Background()
	var genesis []byte
	var version int
	if err := db.QueryRowContext(ctx, `SELECT state,version FROM run_genesis WHERE company_stream_id=$1 AND run_seq=$2`, companyStreamID, runSeq).Scan(&genesis, &version); err != nil {
		t.Fatal(err)
	}
	rows, err := db.QueryContext(ctx, `SELECT seq,canonical_payload,replay_inputs,receipt FROM run_log WHERE company_stream_id=$1 AND run_seq=$2 ORDER BY seq`, companyStreamID, runSeq)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	entries := []ReplayLogEntry{}
	for rows.Next() {
		var entry ReplayLogEntry
		if err := rows.Scan(&entry.Sequence, &entry.CanonicalPayload, &entry.ReplayInputs, &entry.ReceiptJSON); err != nil {
			t.Fatal(err)
		}
		wire, err := parseReplayInputs(entry.ReplayInputs)
		if err != nil {
			t.Fatal(err)
		}
		var discriminator struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(wire.Resolved, &discriminator); err != nil {
			t.Fatal(err)
		}
		entry.Terminal = discriminator.Kind == "exit"
		if entry.Terminal {
			entry.NextCatalog = nextCatalog
		}
		eventRows, err := db.QueryContext(ctx, `SELECT kind,schema_version,intent_id,payload FROM events
			WHERE intent_id=$3 AND stream_id IN ($1,$2)
			ORDER BY CASE WHEN stream_id=$1 THEN 1 ELSE 0 END,event_seq,event_id`, companyStreamID, founderStreamID, wire.Command.IntentID)
		if err != nil {
			t.Fatal(err)
		}
		events := []save.EventWrite{}
		for eventRows.Next() {
			var event save.EventWrite
			if err := eventRows.Scan(&event.Kind, &event.SchemaVersion, &event.IntentID, &event.Payload); err != nil {
				eventRows.Close()
				t.Fatal(err)
			}
			events = append(events, event)
		}
		if err := eventRows.Err(); err != nil {
			eventRows.Close()
			t.Fatal(err)
		}
		eventRows.Close()
		entry.EventsJSON = marshalReplayEvents(events)
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return genesis, version, entries
}

func termsDominate(actual, preview prestigecore.Terms) bool {
	if actual.ReputationDelta < preview.ReputationDelta || actual.RouteKnowledge < preview.RouteKnowledge {
		return false
	}
	actualSlots := make(map[string]string, len(actual.NetworkSlotUnlocks))
	for _, slot := range actual.NetworkSlotUnlocks {
		actualSlots[slot.Slot] = slot.CarriedRef
	}
	for _, slot := range preview.NetworkSlotUnlocks {
		if actualSlots[slot.Slot] != slot.CarriedRef {
			return false
		}
	}
	return true
}

func exactFactSet(actual map[string]bool, expected ...string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for _, fact := range expected {
		if !actual[fact] {
			return false
		}
	}
	return true
}

func preparePendingOpportunityForExit(t *testing.T, bundle CatalogBundle, founderID string, state *save.State, now time.Time) {
	t.Helper()
	if bundle.Opportunities == nil {
		t.Fatal("active-play catalog unavailable")
	}
	state.RunStartedAt = now.Add(-20 * time.Minute)
	state.EvaluatedThrough = now
	state.ManualTokenRefilledAt = now
	attended := int64((20 * time.Minute) / time.Millisecond)
	probe, err := bundle.Opportunities.Spawn(founderID, state.RunSeq, 0, 0)
	if err != nil || probe.SampledIntervalMS > attended {
		t.Fatalf("pending-event probe=%+v err=%v", probe, err)
	}
	pending, err := bundle.Opportunities.Spawn(founderID, state.RunSeq, 0, attended-probe.SampledIntervalMS)
	if err != nil || pending.SpawnedAttendedMS != attended {
		t.Fatalf("pending-event spawn=%+v err=%v", pending, err)
	}
	state.OpportunitySpawnSeq = 1
	state.NextOpportunityAttendedMS = 0
	state.PendingOpportunity = &save.PendingOpportunity{OpportunityID: pending.OpportunityID,
		SpawnedAttendedMS: pending.SpawnedAttendedMS, ExpiresAttendedMS: pending.ExpiresAttendedMS,
		EffectRowID: pending.EffectRowID, SelectedGeneratorID: activeNullableString(pending.SelectedGenerator)}
	if state.ActiveBuffs == nil {
		state.ActiveBuffs = []save.ActiveBuff{}
	}
}

func initialPrestigeWitnessState(t *testing.T, catalog *economy.Catalog, scope economy.Scope, now time.Time) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, scope)
	if err != nil {
		t.Fatal(err)
	}
	counts, provisioned, remainders := map[string]int64{}, map[string]int64{}, map[string]int64{}
	for _, generator := range catalog.GeneratorClassesForScope(scope) {
		counts[generator.ID] = 0
		provisioned[generator.ID] = 0
		if generator.Provision != nil {
			remainders[generator.Provision.GeneratorID] = 0
		}
	}
	state := &save.State{Ledger: ledger, GeneratorCounts: counts, GeneratorProvisioned: provisioned,
		ProvisionRemaindersPPM: remainders, UpgradesOwned: map[string]bool{}, EvaluatedThrough: now,
		ManualTokenRefilledAt: now, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{},
		HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.Zero,
		OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	if scope == economy.ScopeCompany {
		state.RunSeq = 1
		state.RunStartedAt = now
		state.ManualTokenMilli = catalog.ManualPolicy().BucketCapMilli
	}
	return state
}
